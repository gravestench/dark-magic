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

func (packageFS *PackageFS) sortedExports() []string {
	result := make([]string, 0, len(packageFS.exports))
	for root := range packageFS.exports {
		result = append(result, root)
	}
	sort.Strings(result)
	return result
}

type virtualEntry struct{ name string }

func (entry virtualEntry) Name() string               { return entry.name }
func (entry virtualEntry) IsDir() bool                { return true }
func (entry virtualEntry) Type() fs.FileMode          { return fs.ModeDir }
func (entry virtualEntry) Info() (fs.FileInfo, error) { return virtualInfo(entry), nil }

type virtualInfo virtualEntry

func (info virtualInfo) Name() string       { return info.name }
func (info virtualInfo) Size() int64        { return 0 }
func (info virtualInfo) Mode() fs.FileMode  { return fs.ModeDir | 0o555 }
func (info virtualInfo) ModTime() time.Time { return time.Time{} }
func (info virtualInfo) IsDir() bool        { return true }
func (info virtualInfo) Sys() any           { return nil }

type virtualDirectory struct {
	name    string
	entries []fs.DirEntry
	offset  int
}

func (directory *virtualDirectory) Stat() (fs.FileInfo, error) {
	return virtualInfo(virtualEntry{name: path.Base(directory.name)}), nil
}
func (directory *virtualDirectory) Read([]byte) (int, error) { return 0, errors.New("is a directory") }
func (directory *virtualDirectory) Close() error             { return nil }
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
