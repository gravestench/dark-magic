package assetinspect

import (
	dc6 "github.com/gravestench/dc6/pkg"
	"github.com/gravestench/dcc"
)

// decodeDC6Details summarizes every direction so the reported frame total
// remains useful even when directions contain different numbers of frames.
func decodeDC6Details(data []byte) (any, error) {
	asset, err := dc6.FromBytes(data)
	if err != nil {
		return nil, err
	}

	frames := 0
	for _, direction := range asset.Directions {
		frames += len(direction.Frames)
	}

	return map[string]any{
		"version":    asset.Version,
		"directions": len(asset.Directions),
		"frames":     frames,
	}, nil
}

// decodeDCCDetails reports header-level animation facts without retaining the
// decoded asset, keeping inspection independent of rendering ownership.
func decodeDCCDetails(data []byte) (any, error) {
	asset, err := dcc.FromBytes(data)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"version":     asset.Version,
		"directions":  len(asset.Directions()),
		"coded_bytes": asset.TotalSizeCoded,
	}, nil
}
