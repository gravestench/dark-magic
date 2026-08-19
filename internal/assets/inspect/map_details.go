package assetinspect

import (
	"sort"

	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
)

// decodeDS1Details exposes structural dimensions and object count without
// requiring the external DT1 libraries needed to render the stamp.
func decodeDS1Details(data []byte) (any, error) {
	asset, err := ds1.FromBytes(data)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"version": asset.Version,
		"width":   asset.Width,
		"height":  asset.Height,
		"act":     asset.Act,
		"objects": len(asset.Objects),
	}, nil
}

// decodeDT1Details deduplicates type and style identifiers so callers can see
// the tileset's structural coverage without inspecting every decoded tile.
func decodeDT1Details(data []byte) (any, error) {
	asset, err := dt1.FromBytes(data)
	if err != nil {
		return nil, err
	}

	types := make(map[int32]bool)
	styles := make(map[int32]bool)

	for _, tile := range asset.Tiles {
		types[tile.Type] = true
		styles[tile.Style] = true
	}

	return map[string]any{
		"tiles":  len(asset.Tiles),
		"types":  sortedInt32Keys(types),
		"styles": sortedInt32Keys(styles),
	}, nil
}

// sortedInt32Keys makes set-derived report fields deterministic, preventing map
// iteration order from leaking into serialized inspection output.
func sortedInt32Keys(values map[int32]bool) []int32 {
	result := make([]int32, 0, len(values))
	for value := range values {
		result = append(result, value)
	}

	sort.Slice(result, func(left, right int) bool {
		return result[left] < result[right]
	})

	return result
}
