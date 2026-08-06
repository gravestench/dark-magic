package content

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"path"
	"time"
)

var archiveTimestamp = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// WriteShimArchive writes the embedded, redistributable Dark Magic content as
// a deterministic ZIP suitable for mounting or distribution.
func WriteShimArchive(destination io.Writer) error {
	writer := zip.NewWriter(destination)
	source := Shim()
	err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}
		header := &zip.FileHeader{Name: path.Clean(name), Method: zip.Deflate}
		header.SetModTime(archiveTimestamp)
		header.SetMode(0o644)
		file, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := file.Write(data); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = writer.Close()
		return fmt.Errorf("content: archive shim: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("content: close shim archive: %w", err)
	}
	return nil
}
