package recordstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"testing/fstest"
)

// ErrNoAuthoritativeTables distinguishes a missing game-data installation from individual table read failures.
var ErrNoAuthoritativeTables = errors.New("recordstore: no authoritative tables")

const (
	// GenerationSchema versions the serialized generation manifest consumed across session boundaries.
	GenerationSchema = "dark-magic.game-data-generation/v3"
	// ParserSchema versions the table interpretation rules independently from the generation manifest.
	ParserSchema = "dark-magic.authoritative-records/v2"
	// AuthoritativeRoot limits generation identity to simulation-affecting tabular data.
	AuthoritativeRoot = "data/global/excel"
	// AnimationDataPath adds the simulation timing data stored outside AuthoritativeRoot to the generation.
	AnimationDataPath = "data/global/AnimData.d2"
)

// Some retail MPQs can open known hash-table members which their internal
// (listfile) does not enumerate. Keep this list restricted to authoritative
// paths consumed during application/server startup; discovered members still
// enter the same immutable generation and provenance boundary.
var requiredUnlistedPaths = []string{
	"data/global/excel/MonLvl.txt",
	"data/global/excel/MonPreset.txt",
	"data/global/excel/MonStats2.txt",
	"data/global/excel/PetType.txt",
	"data/global/excel/SkillDesc.txt",
}

// generationSource captures the layered-content operations needed to freeze both effective bytes and provenance.
type generationSource interface {
	fs.FS
	List(root, suffix string) ([]string, error)
	ResolveSource(name string) (layer, resolvedPath string, err error)
}

// GenerationFile records the immutable bytes and winning source that contribute to a generation identity.
type GenerationFile struct {
	Path       string `json:"path"`
	Source     string `json:"source"`
	SourcePath string `json:"source_path"`
	Size       int    `json:"size"`
	Digest     string `json:"digest"`
}

// Generation is the deterministic manifest shared by peers to identify one authoritative game-data view.
type Generation struct {
	Schema string           `json:"schema"`
	Parser string           `json:"parser"`
	Files  []GenerationFile `json:"files"`
	ID     string           `json:"-"`
}

// Pin copies the effective authoritative table and timing bytes into an immutable store.
// Later mounts, invalidations, or file edits cannot change this session view.
func Pin(source generationSource) (*Store, Generation, error) {
	if source == nil {
		return nil, Generation{}, fmt.Errorf("recordstore: generation source is required")
	}

	paths, err := authoritativePaths(source)
	if err != nil {
		return nil, Generation{}, err
	}

	files, generation, err := snapshotGeneration(source, paths)
	if err != nil {
		return nil, Generation{}, err
	}

	generation.ID, err = generationDigest(generation)
	if err != nil {
		return nil, Generation{}, err
	}

	pinned, err := newPinnedStore(files, generation)
	if err != nil {
		return nil, Generation{}, err
	}

	return pinned, generation, nil
}

// authoritativePaths discovers every simulation-affecting file before sorting it into deterministic manifest order.
// A table remains mandatory even when animation timing exists, preventing an incomplete install from looking valid.
func authoritativePaths(source generationSource) ([]string, error) {
	paths, err := source.List(AuthoritativeRoot, ".txt")
	if err != nil {
		return nil, fmt.Errorf("recordstore: list authoritative tables: %w", err)
	}

	paths, err = includeRequiredUnlistedPaths(source, paths)
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("%w below %q", ErrNoAuthoritativeTables, AuthoritativeRoot)
	}

	paths, err = includeAnimationData(source, paths)
	if err != nil {
		return nil, err
	}

	sort.Strings(paths)

	return paths, nil
}

// includeRequiredUnlistedPaths probes known retail MPQ members omitted from some internal listfiles. The folded index
// prevents duplicate entries when a source enumerates the same authored path with different casing.
func includeRequiredUnlistedPaths(source generationSource, paths []string) ([]string, error) {
	listed := make(map[string]struct{}, len(paths))
	for _, name := range paths {
		listed[strings.ToLower(name)] = struct{}{}
	}

	for _, name := range requiredUnlistedPaths {
		if _, found := listed[strings.ToLower(name)]; found {
			continue
		}
		if _, statErr := fs.Stat(source, name); statErr == nil {
			paths = append(paths, name)
			listed[strings.ToLower(name)] = struct{}{}
		} else if !errors.Is(statErr, fs.ErrNotExist) {
			return nil, fmt.Errorf("recordstore: inspect required authoritative table %q: %w", name, statErr)
		}
	}

	return paths, nil
}

// includeAnimationData conditionally adds authoritative timing bytes without treating their absence as an error.
func includeAnimationData(source generationSource, paths []string) ([]string, error) {
	if _, statErr := fs.Stat(source, AnimationDataPath); statErr == nil {
		paths = append(paths, AnimationDataPath)
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return nil, fmt.Errorf("recordstore: inspect authoritative animation data: %w", statErr)
	}

	return paths, nil
}

// snapshotGeneration copies effective bytes and provenance into one manifest. Copying severs ownership from mutable
// mounts so later edits cannot alter an active session's data.
func snapshotGeneration(
	source generationSource,
	paths []string,
) (fstest.MapFS, Generation, error) {
	files := make(fstest.MapFS, len(paths))
	generation := Generation{
		Schema: GenerationSchema,
		Parser: ParserSchema,
		Files:  make([]GenerationFile, 0, len(paths)),
	}

	for _, name := range paths {
		data, readErr := fs.ReadFile(source, name)
		if readErr != nil {
			return nil, Generation{}, fmt.Errorf("recordstore: pin %q: %w", name, readErr)
		}
		layer, resolvedPath, resolveErr := source.ResolveSource(name)
		if resolveErr != nil {
			return nil, Generation{}, fmt.Errorf("recordstore: resolve %q: %w", name, resolveErr)
		}

		digest := sha256.Sum256(data)
		generation.Files = append(generation.Files, GenerationFile{
			Path:       name,
			Source:     layer,
			SourcePath: resolvedPath,
			Size:       len(data),
			Digest:     "sha256:" + hex.EncodeToString(digest[:]),
		})
		files[name] = &fstest.MapFile{Data: append([]byte(nil), data...)}
	}

	return files, generation, nil
}

// generationDigest hashes the JSON manifest without its derived ID, keeping identity deterministic and avoiding a
// recursive hash definition. File provenance contributes alongside bytes so different winning layers remain distinct.
func generationDigest(generation Generation) (string, error) {
	encoded, err := json.Marshal(generation)
	if err != nil {
		return "", err
	}

	digest := sha256.Sum256(encoded)

	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

// newPinnedStore builds the case-insensitive index only after the generation is fully copied. Rejecting case-folded
// collisions prevents lookups from selecting a different authored file according to insertion order.
func newPinnedStore(files fstest.MapFS, generation Generation) (*Store, error) {
	pinned := New(files)
	pinned.generationID = generation.ID
	pinned.provenance = make(map[string]Provenance, len(generation.Files))

	for _, file := range generation.Files {
		folded := strings.ToLower(file.Path)
		if existing := pinned.canonical[folded]; existing != "" && existing != file.Path {
			return nil, fmt.Errorf(
				"recordstore: authoritative paths differ only by case: %q and %q",
				existing,
				file.Path,
			)
		}
		pinned.canonical[folded] = file.Path
		pinned.provenance[file.Path] = Provenance{Layer: file.Source, Path: file.SourcePath}
	}

	return pinned, nil
}

// Validate rejects manifests that peers cannot safely use as a non-empty authoritative generation. It deliberately
// validates shape rather than recomputing the digest because callers may receive the manifest without source bytes.
func (generation Generation) Validate() error {
	if generation.Schema != GenerationSchema || generation.Parser != ParserSchema ||
		!strings.HasPrefix(generation.ID, "sha256:") ||
		len(generation.ID) != len("sha256:")+64 ||
		len(generation.Files) == 0 {
		return fmt.Errorf("recordstore: invalid game-data generation")
	}

	return nil
}
