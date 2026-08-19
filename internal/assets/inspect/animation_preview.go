package assetinspect

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/gravestench/dark-magic/internal/assets/decode"
	dc6 "github.com/gravestench/dc6/pkg"
	"github.com/gravestench/dcc"
)

// DCCPreview renders one decoded DCC animation frame as PNG. Bounds are checked
// before frame access so malformed user selections return stable range errors.
func DCCPreview(source fs.FS, path string, direction, frame int) ([]byte, error) {
	file, err := source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening DCC asset %q: %w", path, err)
	}
	defer closeFileWithoutReporting(file)

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading DCC asset %q: %w", path, err)
	}

	asset, err := dcc.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("decoding DCC asset %q: %w", path, err)
	}

	directions := asset.Directions()
	if direction < 0 || direction >= len(directions) {
		return nil, fmt.Errorf("direction %d out of range [0,%d)", direction, len(directions))
	}

	frames := directions[direction].Frames()
	if frame < 0 || frame >= len(frames) {
		return nil, fmt.Errorf("frame %d out of range [0,%d)", frame, len(frames))
	}

	return encodePreviewPNG(frames[frame], "encoding DCC preview")
}

// DC6Preview renders one DC6 frame with the decoder's fallback palette. This
// keeps headless diagnostics available before a game palette is initialized.
func DC6Preview(source fs.FS, path string, direction, frame int) ([]byte, error) {
	file, err := source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening DC6 asset %q: %w", path, err)
	}
	defer closeFileWithoutReporting(file)

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading DC6 asset %q: %w", path, err)
	}

	asset, err := dc6.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("decoding DC6 asset %q: %w", path, err)
	}

	if direction < 0 || direction >= len(asset.Directions) {
		return nil, fmt.Errorf("direction %d out of range [0,%d)", direction, len(asset.Directions))
	}

	frames := asset.Directions[direction].Frames
	if frame < 0 || frame >= len(frames) {
		return nil, fmt.Errorf("frame %d out of range [0,%d)", frame, len(frames))
	}

	decoded, err := assetdecode.FrameImage(asset, frames[frame])
	if err != nil {
		return nil, err
	}

	return encodePreviewPNG(decoded, "encoding preview")
}
