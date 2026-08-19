// Package assetinspect provides a headless way to verify and describe assets
// from either a directory or an archive-backed fs.FS.
package assetinspect

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

// Report is a JSON-ready structural observation. Details contain decoder facts,
// never proprietary source bytes or rendered pixels, so callers may safely
// serialize the result without copying owned asset data.
type Report struct {
	Path    string `json:"path"`
	Type    string `json:"type"`
	Bytes   int64  `json:"bytes"`
	SHA256  string `json:"sha256"`
	Details any    `json:"details,omitempty"`
}

// Inspect reads one virtual content path and delegates to the matching headless
// decoder. The caller retains ownership of the layered filesystem, while this
// function confines ownership of the opened file to the duration of the call.
func Inspect(source fs.FS, path string) (Report, error) {
	file, err := source.Open(path)
	if err != nil {
		return Report{}, fmt.Errorf("opening asset %q: %w", path, err)
	}
	defer closeFileWithoutReporting(file)

	data, err := io.ReadAll(file)
	if err != nil {
		return Report{}, fmt.Errorf("reading asset %q: %w", path, err)
	}

	return InspectData(path, data)
}

// InspectData describes already-loaded asset bytes without retaining the input
// slice. Services with composite loaders can therefore keep ownership and reuse
// their buffers after the inspection completes.
func InspectData(path string, data []byte) (Report, error) {
	digest := sha256.Sum256(data)

	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if extension == "" {
		extension = "unknown"
	}

	report := Report{
		Path:   path,
		Type:   extension,
		Bytes:  int64(len(data)),
		SHA256: hex.EncodeToString(digest[:]),
	}

	details, err := decodeDetails(extension, data)
	if err != nil {
		return Report{}, fmt.Errorf("decoding %s asset %q: %w", extension, path, err)
	}

	report.Details = details

	return report, nil
}

// closeFileWithoutReporting preserves preview and inspection APIs that complete
// before deferred close errors are known; callers still retain filesystem ownership.
func closeFileWithoutReporting(file fs.File) {
	_ = file.Close()
}
