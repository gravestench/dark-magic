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

// WriteD2LegacyArchive writes the embedded, redistributable d2legacy mod as a deterministic, mountable ZIP.
// WalkDir ordering plus fixed metadata ensures equal embedded trees produce byte-identical distribution artifacts.
func WriteD2LegacyArchive(destination io.Writer) error {
	writer := zip.NewWriter(destination)
	source := D2Legacy()

	err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		return writeD2LegacyArchiveFile(writer, source, name)
	})
	if err != nil {
		// Close is still required to release the compressor, but the walk failure remains the public error.
		_ = writer.Close()
		return fmt.Errorf("content: archive d2legacy: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("content: close d2legacy archive: %w", err)
	}

	return nil
}

// writeD2LegacyArchiveFile copies one embedded file with fixed metadata so host timestamps and modes cannot leak in.
func writeD2LegacyArchiveFile(writer *zip.Writer, source fs.FS, name string) error {
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return err
	}

	header := &zip.FileHeader{Name: path.Clean(name), Method: zip.Deflate}
	// SetModTime preserves the existing DOS and extended timestamp header bytes used by distributed archives.
	header.SetModTime(archiveTimestamp) //nolint:staticcheck // Replacing it could change the serialized ZIP format.
	header.SetMode(0o644)

	file, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}
