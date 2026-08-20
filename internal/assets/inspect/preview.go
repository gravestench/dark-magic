package assetinspect

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"path/filepath"
	"strings"
)

// Preview routes supported assets to their format-specific PNG renderer. The
// explicit extension boundary prevents one decoder from guessing another format.
func Preview(source fs.FS, path string, direction, frame int) ([]byte, error) {
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	switch extension {
	case "dc6":
		return DC6Preview(source, path, direction, frame)
	case "dcc":
		return DCCPreview(source, path, direction, frame)
	case "ds1":
		return DS1Preview(source, path)
	default:
		return nil, fmt.Errorf("PNG preview is not supported for %q assets", extension)
	}
}

// encodePreviewPNG centralizes serialization while accepting the public error
// text each preview API has historically exposed to callers.
func encodePreviewPNG(preview image.Image, errorMessage string) ([]byte, error) {
	var output bytes.Buffer
	if err := png.Encode(&output, preview); err != nil {
		return nil, fmt.Errorf("%s: %w", errorMessage, err)
	}

	return output.Bytes(), nil
}
