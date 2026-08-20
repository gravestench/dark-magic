package modcache

import (
	"errors"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"
)

// PackageFS exposes all private package files beneath mods/<id>/ and only the
// manifest-declared content roots at the shared VFS root. Code and metadata
// therefore cannot collide merely because two packages both contain boot.lua.
type PackageFS struct {
	id      string
	source  fs.FS
	exports map[string]struct{}
}

// NewPackageFS validates every declared shared root before exposing the package,
// so a mount cannot advertise missing or non-directory content.
func NewPackageFS(manifest Manifest, source fs.FS) (*PackageFS, error) {
	if err := ValidateManifest(manifest); err != nil {
		return nil, err
	}

	if source == nil {
		return nil, errors.New("modcache: package filesystem is required")
	}

	exports := make(map[string]struct{}, len(manifest.ContentRoots))
	for _, root := range manifest.ContentRoots {
		info, err := fs.Stat(source, root)
		if err != nil || !info.IsDir() {
			return nil, &fs.PathError{Op: "mount content root", Path: root, Err: errors.Join(err, fs.ErrInvalid)}
		}

		exports[root] = struct{}{}
	}

	return &PackageFS{id: manifest.ID, source: source, exports: exports}, nil
}

// Open maps private and exported virtual paths while synthesizing only the two
// directory levels that do not physically exist in the package source.
func (packageFS *PackageFS) Open(name string) (fs.File, error) {
	mapped, virtual, err := packageFS.mapPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	if virtual != "" {
		return &virtualDirectory{name: virtual, entries: packageFS.virtualEntries(virtual)}, nil
	}

	return packageFS.source.Open(mapped)
}

// ReadDir mirrors Open's namespace policy so directory discovery cannot reveal
// files that direct opens would keep private.
func (packageFS *PackageFS) ReadDir(name string) ([]fs.DirEntry, error) {
	mapped, virtual, err := packageFS.mapPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}

	if virtual != "" {
		return packageFS.virtualEntries(virtual), nil
	}

	return fs.ReadDir(packageFS.source, mapped)
}

// mapPath distinguishes physical package paths from synthesized directories and
// rejects every undeclared shared root.
func (packageFS *PackageFS) mapPath(name string) (mapped string, virtual string, err error) {
	if !fs.ValidPath(name) {
		return "", "", fs.ErrInvalid
	}

	if name == "." {
		return "", ".", nil
	}

	if name == "mods" {
		return "", "mods", nil
	}

	prefix := "mods/" + packageFS.id
	if name == prefix {
		return ".", "", nil
	}

	if strings.HasPrefix(name, prefix+"/") {
		return strings.TrimPrefix(name, prefix+"/"), "", nil
	}

	root := strings.SplitN(name, "/", 2)[0]
	if _, exported := packageFS.exports[root]; exported {
		return name, "", nil
	}

	return "", "", fs.ErrNotExist
}

// virtualEntries exposes the package ID beneath mods and sorted shared exports
// at the root, keeping directory listings deterministic.
func (packageFS *PackageFS) virtualEntries(name string) []fs.DirEntry {
	if name == "mods" {
		return []fs.DirEntry{virtualEntry{name: packageFS.id}}
	}

	entries := []fs.DirEntry{virtualEntry{name: "mods"}}
	for _, root := range packageFS.sortedExports() {
		entries = append(entries, virtualEntry{name: root})
	}

	return entries
}

// sortedExports copies map keys into a stable order so filesystem traversal and
// overlays do not depend on Go map iteration.
func (packageFS *PackageFS) sortedExports() []string {
	result := make([]string, 0, len(packageFS.exports))
	for root := range packageFS.exports {
		result = append(result, root)
	}

	sort.Strings(result)

	return result
}

type virtualEntry struct{ name string }

// Name returns the synthesized directory's single path segment.
func (entry virtualEntry) Name() string { return entry.name }

// IsDir identifies every synthesized entry as a directory.
func (entry virtualEntry) IsDir() bool { return true }

// Type exposes directory mode without requiring an additional stat operation.
func (entry virtualEntry) Type() fs.FileMode { return fs.ModeDir }

// Info returns immutable metadata for the synthesized directory.
func (entry virtualEntry) Info() (fs.FileInfo, error) { return virtualInfo(entry), nil }

type virtualInfo virtualEntry

// Name returns the synthesized directory's single path segment.
func (info virtualInfo) Name() string { return info.name }

// Size is zero because synthesized directories contain no stored bytes.
func (info virtualInfo) Size() int64 { return 0 }

// Mode makes synthesized namespaces read-only to match immutable packages.
func (info virtualInfo) Mode() fs.FileMode { return fs.ModeDir | 0o555 }

// ModTime is stable and empty because virtual directories have no source mtime.
func (info virtualInfo) ModTime() time.Time { return time.Time{} }

// IsDir identifies the synthesized metadata as a directory.
func (info virtualInfo) IsDir() bool { return true }

// Sys has no platform backing value because the directory is virtual.
func (info virtualInfo) Sys() any { return nil }

type virtualDirectory struct {
	name    string
	entries []fs.DirEntry
	offset  int
}

// Stat returns stable metadata for the directory's current virtual path.
func (directory *virtualDirectory) Stat() (fs.FileInfo, error) {
	return virtualInfo(virtualEntry{name: path.Base(directory.name)}), nil
}

// Read rejects byte reads to satisfy the same contract as physical directories.
func (directory *virtualDirectory) Read([]byte) (int, error) {
	return 0, errors.New("is a directory")
}

// Close is a no-op because synthesized directories own no operating-system handle.
func (directory *virtualDirectory) Close() error { return nil }

// ReadDir maintains an offset for positive counts and returns all remaining
// entries for non-positive counts, matching fs.ReadDirFile semantics.
func (directory *virtualDirectory) ReadDir(count int) ([]fs.DirEntry, error) {
	if directory.offset >= len(directory.entries) && count > 0 {
		return nil, io.EOF
	}

	end := len(directory.entries)
	if count > 0 && directory.offset+count < end {
		end = directory.offset + count
	}

	result := append([]fs.DirEntry(nil), directory.entries[directory.offset:end]...)
	directory.offset = end

	return result, nil
}
