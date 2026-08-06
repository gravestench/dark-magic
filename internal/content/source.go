package content

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravestench/mpq"
)

// OpenSource opens a directory, MPQ, or ZIP path using the normalized content
// filesystem contract.
func OpenSource(fileName string) (fs.FS, error) {
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

// Directory opens root as a content filesystem.
func Directory(root string) fs.FS {
	return normalizedFS{FS: os.DirFS(root)}
}

// ZIP opens a zip archive as a content filesystem.
func ZIP(fileName string) (fs.FS, error) {
	reader, err := zip.OpenReader(fileName)
	if err != nil {
		return nil, fmt.Errorf("content: open zip %q: %w", fileName, err)
	}
	return &closeableFS{FS: normalizedFS{FS: reader}, close: reader.Close}, nil
}

// MPQ opens a Diablo MPQ archive as a content filesystem.
func MPQ(fileName string) (fs.FS, error) {
	archive, err := mpq.New(fileName)
	if err != nil {
		return nil, fmt.Errorf("content: open MPQ %q: %w", fileName, err)
	}
	return normalizedFS{FS: archive, backslash: true}, nil
}

// Close closes a filesystem source if it owns resources.
func Close(source fs.FS) error {
	if closer, ok := source.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

type normalizedFS struct {
	fs.FS
	backslash bool
}

func (n normalizedFS) Open(name string) (fs.File, error) {
	clean, err := Normalize(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	if n.backslash {
		clean = strings.ReplaceAll(clean, "/", `\`)
	}
	file, err := n.FS.Open(clean)
	// gravestench/mpq predates io/fs and returns an unwrapped sentinel text.
	// Translate it at this adapter boundary so layered lookup can continue to
	// lower-priority archives.
	if n.backslash && err != nil && (err.Error() == "file not found" || strings.HasSuffix(err.Error(), ": file not found")) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return file, err
}

type closeableFS struct {
	fs.FS
	close func() error
}

func (f *closeableFS) Close() error { return f.close() }
