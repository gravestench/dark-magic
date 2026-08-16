package recordstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing/fstest"
)

var ErrNoAuthoritativeTables = errors.New("recordstore: no authoritative tables")

const (
	GenerationSchema  = "dark-magic.game-data-generation/v2"
	ParserSchema      = "dark-magic.records-tsv/v1"
	AuthoritativeRoot = "data/global/excel"
)

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

// Pin copies the effective authoritative table bytes into an immutable store.
// Later mounts, invalidations, or file edits cannot change this session view.
func Pin(source generationSource) (*Store, Generation, error) {
	if source == nil {
		return nil, Generation{}, fmt.Errorf("recordstore: generation source is required")
	}
	paths, err := source.List(AuthoritativeRoot, ".txt")
	if err != nil {
		return nil, Generation{}, fmt.Errorf("recordstore: list authoritative tables: %w", err)
	}
	if len(paths) == 0 {
		return nil, Generation{}, fmt.Errorf("%w below %q", ErrNoAuthoritativeTables, AuthoritativeRoot)
	}
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
