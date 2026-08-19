package content

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	"github.com/gravestench/mpq"
)

// OpenSource opens a directory, MPQ, or ZIP path using the normalized content
// filesystem contract.
func OpenSource(fileName string) (fs.FS, error) {
	expanded, err := darkpaths.ExpandHost(fileName)
	if err != nil {
		return nil, fmt.Errorf("content: expand source %q: %w", fileName, err)
	}

	fileName = expanded

	info, err := os.Stat(fileName)
	if err != nil {
		return nil, fmt.Errorf("content: inspect source %q: %w", fileName, err)
	}

	if info.IsDir() {
		return Directory(fileName), nil
	}

	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".mpq":
		return MPQ(fileName)
	case ".zip":
		return ZIP(fileName)
	default:
		return nil, fmt.Errorf("content: unsupported source %q", fileName)
	}
}

// Directory opens root through the normalized content-path contract while leaving resource ownership with the OS.
func Directory(root string) fs.FS {
	return normalizedFS{FS: os.DirFS(root)}
}

// ZIP opens a zip archive through the normalized content-path contract; callers must close the returned source.
func ZIP(fileName string) (fs.FS, error) {
	expanded, err := darkpaths.ExpandHost(fileName)
	if err != nil {
		return nil, fmt.Errorf("content: expand ZIP path %q: %w", fileName, err)
	}

	fileName = expanded

	reader, err := zip.OpenReader(fileName)
	if err != nil {
		return nil, fmt.Errorf("content: open zip %q: %w", fileName, err)
	}

	return &closeableFS{FS: normalizedFS{FS: reader}, close: reader.Close}, nil
}

// MPQ opens a Diablo MPQ archive with normalized lookup and flat member enumeration.
// Callers must close the returned source because the adapter owns the archive handle.
func MPQ(fileName string) (fs.FS, error) {
	expanded, err := darkpaths.ExpandHost(fileName)
	if err != nil {
		return nil, fmt.Errorf("content: expand MPQ path %q: %w", fileName, err)
	}

	fileName = expanded

	archive, err := mpq.New(fileName)
	if err != nil {
		return nil, fmt.Errorf("content: open MPQ %q: %w", fileName, err)
	}
	// MPQ archives predate io/fs and expose members through (listfile), not ReadDir. A missing listfile still permits
	// direct access to known paths, so listfile failure deliberately does not reject an otherwise usable archive.
	listed, _ := archive.Listfile()

	return &listedCloseableFS{
		FS:    normalizedFS{FS: archive, backslash: true},
		close: archive.Close,
		paths: normalizedArchivePaths(listed),
	}, nil
}

// Close releases a filesystem source when its concrete adapter owns resources and is otherwise a no-op.
func Close(source fs.FS) error {
	if closer, ok := source.(interface{ Close() error }); ok {
		return closer.Close()
	}

	return nil
}

// normalizedArchivePaths converts an archive index once so every later enumeration shares the content-path contract.
// Invalid and root-like listfile entries are ignored because they cannot identify mountable files.
func normalizedArchivePaths(listed []string) []string {
	paths := make([]string, 0, len(listed))
	for _, name := range listed {
		clean, err := Normalize(name)
		if err == nil && clean != "." {
			paths = append(paths, clean)
		}
	}

	return paths
}

// normalizedFS adapts host and archive path conventions to the slash-separated io/fs contract.
type normalizedFS struct {
	fs.FS
	backslash bool
}

// Open normalizes the requested path and translates the legacy MPQ missing-file sentinel into fs.ErrNotExist.
// That translation lets layered lookup continue to lower-priority sources without masking other archive failures.
func (n normalizedFS) Open(name string) (fs.File, error) {
	clean, err := Normalize(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	if n.backslash {
		clean = strings.ReplaceAll(clean, "/", `\`)
	}

	file, err := n.FS.Open(clean)
	if n.backslash && isMissingMPQFile(err) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	return file, err
}

// isMissingMPQFile recognizes both direct and context-prefixed errors emitted by the pre-io/fs MPQ dependency.
func isMissingMPQFile(err error) bool {
	return err != nil && (err.Error() == "file not found" || strings.HasSuffix(err.Error(), ": file not found"))
}

// closeableFS adds explicit ownership cleanup to an otherwise ordinary filesystem adapter.
type closeableFS struct {
	fs.FS
	close func() error
}

// Close releases the resource owned by this adapter and returns the underlying close result unchanged.
func (f *closeableFS) Close() error { return f.close() }

// listedCloseableFS adapts archive formats that can name their members but do
// not implement io/fs directory walking. Paths returns a copy because mounted
// content is shared by loaders running on several goroutines.
type listedCloseableFS struct {
	fs.FS
	close func() error
	paths []string
}

// Close releases the indexed archive resource and returns the underlying close result unchanged.
func (f *listedCloseableFS) Close() error { return f.close() }

// Paths returns a defensive copy so concurrent loaders cannot mutate the shared archive index.
func (f *listedCloseableFS) Paths() []string {
	return append([]string(nil), f.paths...)
}
