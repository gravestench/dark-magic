package missile

import (
	"fmt"
	"path"
	"strings"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
)

// PresentationFromCatalog joins the client-facing portion of one Missiles.txt
// row. It deliberately does not claim that the row's server function, formulas,
// or collision flags are implemented; those remain explicit simulation work.
func PresentationFromCatalog(snapshot gamedata.Snapshot, missileID string) (Presentation, error) {
	missileID = strings.TrimSpace(missileID)
	record, found := snapshot.MissilesByName[missileID]
	if !found {
		return Presentation{}, fmt.Errorf("missile: unknown presentation record %q", missileID)
	}
	cel := strings.TrimSpace(record.CelFile)
	if cel == "" {
		return Presentation{}, fmt.Errorf("missile: %q has no CelFile", missileID)
	}
	fps := int64(25)
	if record.AnimSpeed > 0 {
		// Missiles.txt stores animation speed in sixteenths of a game frame.
		fps = max(int64(record.AnimSpeed)*25/16, 1)
	}
	return Presentation{
		MissileID: missileID, DCC: path.Join("data/global/missiles", cel+".dcc"),
		Palette: "data/global/palette/units/pal.dat", TravelSound: strings.TrimSpace(record.TravelSound), HitSound: strings.TrimSpace(record.HitSound),
		Directions: int64(max(record.NumDirections, 1)), FramesPerSecond: fps, Loop: record.LoopAnim != 0,
		OffsetX: float64(record.XOffset), OffsetY: float64(record.YOffset), OffsetZ: float64(record.ZOffset),
	}, nil
}
