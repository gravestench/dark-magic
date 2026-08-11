package models

// LevelLink is one paired Levels.txt Vis#/Warp# relationship. The slot number
// is also the main-index authored by ordinary DS1 special tiles.
type LevelLink struct {
	Slot             int
	DestinationLevel int
	WarpID           int
}

// Links returns the eight authored visibility/warp pairs in stable slot order.
// Empty visibility slots are omitted. Keeping the pair together prevents code
// from accidentally resolving Vis3 with Warp2.
func (level LevelData) Links() []LevelLink {
	destinations := [...]int{level.Vis0, level.Vis1, level.Vis2, level.Vis3, level.Vis4, level.Vis5, level.Vis6, level.Vis7}
	warps := [...]int{level.Warp0, level.Warp1, level.Warp2, level.Warp3, level.Warp4, level.Warp5, level.Warp6, level.Warp7}
	result := make([]LevelLink, 0, len(destinations))
	for slot, destination := range destinations {
		if destination <= 0 {
			continue
		}
		result = append(result, LevelLink{Slot: slot, DestinationLevel: destination, WarpID: warps[slot]})
	}
	return result
}
