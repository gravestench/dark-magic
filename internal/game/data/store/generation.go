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

var ErrNoAuthoritativeTables = errors.New("recordstore: no authoritative tables")

const (
	GenerationSchema  = "dark-magic.game-data-generation/v3"
	ParserSchema      = "dark-magic.authoritative-records/v2"
	AuthoritativeRoot = "data/global/excel"
	AnimationDataPath = "data/global/AnimData.d2"
)

// Some retail MPQs can open known hash-table members which their internal
// (listfile) does not enumerate. Keep this list restricted to authoritative
// paths consumed during application/server startup; discovered members still
// enter the same immutable generation and provenance boundary.
var requiredUnlistedPaths = []string{
	"data/global/excel/MonPreset.txt",
	"data/global/excel/MonStats2.txt",
	"data/global/excel/SkillDesc.txt",
}

type generationSource interface {
	fs.FS
	List(root, suffix string) ([]string, error)
	ResolveSource(name string) (layer, resolvedPath string, err error)
}

type GenerationFile struct {
	Path       string `json:"path"`
	Source     string `json:"source"`
	SourcePath string `json:"source_path"`
	Size       int    `json:"size"`
	Digest     string `json:"digest"`
}

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
	paths, err := source.List(AuthoritativeRoot, ".txt")
	if err != nil {
		return nil, Generation{}, fmt.Errorf("recordstore: list authoritative tables: %w", err)
	}
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
			return nil, Generation{}, fmt.Errorf("recordstore: inspect required authoritative table %q: %w", name, statErr)
		}
	}
	if len(paths) == 0 {
		return nil, Generation{}, fmt.Errorf("%w below %q", ErrNoAuthoritativeTables, AuthoritativeRoot)
	}
	if _, statErr := fs.Stat(source, AnimationDataPath); statErr == nil {
		paths = append(paths, AnimationDataPath)
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return nil, Generation{}, fmt.Errorf("recordstore: inspect authoritative animation data: %w", statErr)
	}
	sort.Strings(paths)
	files := make(fstest.MapFS, len(paths))
	generation := Generation{Schema: GenerationSchema, Parser: ParserSchema, Files: make([]GenerationFile, 0, len(paths))}
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
		generation.Files = append(generation.Files, GenerationFile{Path: name, Source: layer,
			SourcePath: resolvedPath, Size: len(data), Digest: "sha256:" + hex.EncodeToString(digest[:])})
		files[name] = &fstest.MapFile{Data: append([]byte(nil), data...)}
	}
	encoded, err := json.Marshal(generation)
	if err != nil {
		return nil, Generation{}, err
	}
	digest := sha256.Sum256(encoded)
	generation.ID = "sha256:" + hex.EncodeToString(digest[:])
	pinned := New(files)
	pinned.generationID = generation.ID
	pinned.provenance = make(map[string]Provenance, len(generation.Files))
	for _, file := range generation.Files {
		folded := strings.ToLower(file.Path)
		if existing := pinned.canonical[folded]; existing != "" && existing != file.Path {
			return nil, Generation{}, fmt.Errorf("recordstore: authoritative paths differ only by case: %q and %q", existing, file.Path)
		}
		pinned.canonical[folded] = file.Path
		pinned.provenance[file.Path] = Provenance{Layer: file.Source, Path: file.SourcePath}
	}
	return pinned, generation, nil
}

func (generation Generation) Validate() error {
	if generation.Schema != GenerationSchema || generation.Parser != ParserSchema ||
		!strings.HasPrefix(generation.ID, "sha256:") || len(generation.ID) != len("sha256:")+64 || len(generation.Files) == 0 {
		return fmt.Errorf("recordstore: invalid game-data generation")
	}
	return nil
}
