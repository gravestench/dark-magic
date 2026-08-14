// Package modcache installs immutable mod packages and resolves user-selected
// profiles into an exact, dependency-ordered session set.
package modcache

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"unicode"
)

const (
	ManifestSchema   = "dark-magic.mod/v1"
	ProfileSchema    = "dark-magic.mod-profile/v1"
	IndexSchema      = "dark-magic.mod-index/v1"
	LockSchema       = "dark-magic.mod-lock/v1"
	EngineAPI        = "v1"
	manifestPath     = "mod.json"
	maxManifestBytes = 64 << 10
)

var ErrInvalidManifest = errors.New("modcache: invalid manifest")

type Manifest struct {
	Schema          string       `json:"schema"`
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Version         string       `json:"version"`
	Kind            string       `json:"kind"`
	EngineAPI       string       `json:"engine_api"`
	Redistributable bool         `json:"redistributable"`
	ContentRoots    []string     `json:"content_roots"`
	Entrypoints     Entrypoints  `json:"entrypoints"`
	Dependencies    []Dependency `json:"dependencies"`
}

type Entrypoints struct {
	ClientComponents    []string `json:"client_components"`
	AuthorityComponents []string `json:"authority_components"`
}

type Dependency struct {
	ID      string `json:"id"`
	Version string `json:"version,omitempty"`
}

func ReadManifest(source fs.FS) (Manifest, error) {
	file, err := source.Open(manifestPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("%w: read %s: %v", ErrInvalidManifest, manifestPath, err)
	}
	data, readErr := io.ReadAll(io.LimitReader(file, maxManifestBytes+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil {
		return Manifest{}, fmt.Errorf("%w: read %s: %v", ErrInvalidManifest, manifestPath, errors.Join(readErr, closeErr))
	}
	if len(data) > maxManifestBytes {
		return Manifest{}, fmt.Errorf("%w: %s exceeds %d bytes", ErrInvalidManifest, manifestPath, maxManifestBytes)
	}
	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("%w: decode %s: %v", ErrInvalidManifest, manifestPath, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Manifest{}, fmt.Errorf("%w: trailing %s data", ErrInvalidManifest, manifestPath)
	}
	if err := ValidateManifest(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func ValidateManifest(manifest Manifest) error {
	if manifest.Schema != ManifestSchema || !validID(manifest.ID) || strings.TrimSpace(manifest.Name) == "" ||
		strings.TrimSpace(manifest.Version) == "" || manifest.EngineAPI != EngineAPI ||
		(manifest.Kind != "game" && manifest.Kind != "extension") {
		return ErrInvalidManifest
	}
	seenDependencies := make(map[string]struct{}, len(manifest.Dependencies))
	for _, dependency := range manifest.Dependencies {
		if !validID(dependency.ID) || dependency.ID == manifest.ID {
			return fmt.Errorf("%w: invalid dependency %q", ErrInvalidManifest, dependency.ID)
		}
		if _, duplicate := seenDependencies[dependency.ID]; duplicate {
			return fmt.Errorf("%w: duplicate dependency %q", ErrInvalidManifest, dependency.ID)
		}
		seenDependencies[dependency.ID] = struct{}{}
	}
	seenEntrypoints := make(map[string]struct{})
	for _, component := range append(append([]string(nil), manifest.Entrypoints.ClientComponents...), manifest.Entrypoints.AuthorityComponents...) {
		if !validID(component) || !strings.HasPrefix(component, manifest.ID+".") {
			return fmt.Errorf("%w: invalid component %q", ErrInvalidManifest, component)
		}
		if _, duplicate := seenEntrypoints[component]; duplicate {
			return fmt.Errorf("%w: duplicate component %q", ErrInvalidManifest, component)
		}
		seenEntrypoints[component] = struct{}{}
	}
	seenRoots := make(map[string]struct{}, len(manifest.ContentRoots))
	for _, root := range manifest.ContentRoots {
		if !validID(root) || strings.ContainsRune(root, '.') || reservedContentRoot(root) {
			return fmt.Errorf("%w: invalid content root %q", ErrInvalidManifest, root)
		}
		if _, duplicate := seenRoots[root]; duplicate {
			return fmt.Errorf("%w: duplicate content root %q", ErrInvalidManifest, root)
		}
		seenRoots[root] = struct{}{}
	}
	return nil
}

func reservedContentRoot(root string) bool {
	return root == "components" || root == "lua" || root == "mods"
}

func validID(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for index, character := range value {
		if unicode.IsLower(character) || unicode.IsDigit(character) || (index > 0 && strings.ContainsRune("._-", character)) {
			continue
		}
		return false
	}
	return true
}
