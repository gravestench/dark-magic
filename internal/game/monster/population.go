package monster

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	models "github.com/gravestench/dark-magic/internal/game/data/model"
	"github.com/gravestench/dark-magic/internal/game/mapgen"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const PopulationPolicy = "synthetic.room_density_v1"
const populationStream = "mapgen/blood-moor-population"
const subtilesPerTile = 5

// Placement resolves a requested subtile to deterministic walkable world
// space. A materialized world.Map satisfies this interface without exposing
// rendering or texture residency to population planning.
type Placement interface {
	OpenPointNearSubtile(x, y float64) (float64, float64, bool)
}

// PopulationPlan is the inspectable, canonical result between zone generation
// and authoritative monster materialization.
type PopulationPlan struct {
	Policy       string   `json:"policy"`
	ZoneChecksum string   `json:"zone_checksum"`
	Seed         uint64   `json:"seed"`
	Stream       string   `json:"stream"`
	LevelID      int      `json:"level_id"`
	Spawns       []Spawn  `json:"spawns"`
	Trace        []string `json:"trace"`
}

func (plan PopulationPlan) Checksum() (string, error) {
	encoded, err := json.Marshal(plan)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

// BuildBloodMoorPopulation derives ordinary packs from authored Levels and
// MonStats facts. Exact retail density/eligibility remains unverified, so the
// first policy is named and versioned instead of masquerading as compatibility.
func BuildBloodMoorPopulation(zone *mapgen.Zone, placement Placement, snapshot gamedata.Snapshot) (PopulationPlan, error) {
	if zone == nil || placement == nil {
		return PopulationPlan{}, fmt.Errorf("monster: population requires a zone and placement")
	}
	request := zone.Request()
	if request.Act != 1 || request.LevelID != 2 || zone.Kind() != mapgen.Outdoor {
		return PopulationPlan{}, fmt.Errorf("monster: first population policy only supports Blood Moor")
	}
	level, found := snapshot.LevelsByID[request.LevelID]
	if !found {
		return PopulationPlan{}, fmt.Errorf("monster: Levels row %d is unavailable", request.LevelID)
	}
	difficulty := Difficulty(request.Difficulty)
	candidates, err := populationCandidates(snapshot, level, difficulty)
	if err != nil {
		return PopulationPlan{}, err
	}
	density := populationDensity(level, difficulty)
	zoneChecksum, err := zone.Checksum()
	if err != nil {
		return PopulationPlan{}, err
	}
	plan := PopulationPlan{Policy: PopulationPolicy, ZoneChecksum: zoneChecksum, Seed: request.Seed, Stream: populationStream, LevelID: request.LevelID}
	plan.Trace = append(plan.Trace, fmt.Sprintf("Levels[%d] density=%d candidates=%d", request.LevelID, density, len(candidates)))
	if density <= 0 || len(candidates) == 0 {
		return plan, nil
	}

	populated := populatedStamps(zone.Stamps())
	fallbackRooms := make([]mapgen.Room, 0)
	for _, room := range zone.Rooms() {
		if !populated[room.StampID] {
			continue
		}
		fallbackRooms = append(fallbackRooms, room)
		random := mapgen.NewStreams(request.Seed).For(fmt.Sprintf("blood-moor-population/room-%d", room.ID))
		roll := int(random.Uint64n(100000))
		if roll >= density {
			plan.Trace = append(plan.Trace, fmt.Sprintf("room=%d density-roll=%d suppressed", room.ID, roll))
			continue
		}
		if _, err := appendPopulationPack(&plan, room, roll, false, random, candidates, placement, zone.Warps(), level.WarpDist); err != nil {
			return PopulationPlan{}, err
		}
	}
	if len(plan.Spawns) == 0 {
		for _, room := range fallbackRooms {
			random := mapgen.NewStreams(request.Seed).For(fmt.Sprintf("blood-moor-population/room-%d", room.ID))
			roll := int(random.Uint64n(100000))
			placed, err := appendPopulationPack(&plan, room, roll, true, random, candidates, placement, zone.Warps(), level.WarpDist)
			if err != nil {
				return PopulationPlan{}, err
			}
			if placed > 0 {
				break
			}
		}
	}
	if len(plan.Spawns) == 0 {
		plan.Trace = append(plan.Trace, fmt.Sprintf("minimum-pack fallback relaxed WarpDist=%d to 2 tiles", level.WarpDist))
		for _, room := range fallbackRooms {
			random := mapgen.NewStreams(request.Seed).For(fmt.Sprintf("blood-moor-population/room-%d", room.ID))
			roll := int(random.Uint64n(100000))
			placed, err := appendPopulationPack(&plan, room, roll, true, random, candidates, placement, zone.Warps(), 2)
			if err != nil {
				return PopulationPlan{}, err
			}
			if placed > 0 {
				break
			}
		}
	}
	return plan, nil
}

func appendPopulationPack(plan *PopulationPlan, room mapgen.Room, roll int, forced bool, random mapgen.Random, candidates []Definition, placement Placement, warps []mapgen.Warp, warpDistance int) (int, error) {
	definition := weightedDefinition(candidates, random.Uint64n(totalRarity(candidates)))
	groupSize := definition.MinGroup
	if span := definition.MaxGroup - definition.MinGroup + 1; span > 1 {
		groupSize += int(random.Uint64n(uint64(span)))
	}
	placed := 0
	for member := 0; member < groupSize; member++ {
		x, y := memberAnchor(room, member)
		x, y, ok := placement.OpenPointNearSubtile(x, y)
		if !ok || nearWarp(x, y, warps, warpDistance) {
			continue
		}
		id := fmt.Sprintf("blood-moor:r%d:p0:m%d", room.ID, member)
		spawn, err := NewSpawn(id, definition, random.Uint64(), x, y, 1, 2)
		if err != nil {
			return 0, err
		}
		plan.Spawns = append(plan.Spawns, spawn)
		placed++
	}
	policy := "density"
	if forced {
		policy = "minimum-pack-fallback"
	}
	plan.Trace = append(plan.Trace, fmt.Sprintf("room=%d density-roll=%d policy=%s family=%s group=%d placed=%d", room.ID, roll, policy, definition.ID, groupSize, placed))
	return placed, nil
}

// SubmitPopulation queues the plan through the existing privileged command
// authority. Renderer culling never participates in this decision.
func SubmitPopulation(session *gamesession.Session, plan PopulationPlan, actor string, tick uint64) error {
	if session == nil || plan.Policy != PopulationPolicy || tick == 0 {
		return fmt.Errorf("monster: valid population plan, session, and tick are required")
	}
	actor = strings.TrimSpace(actor)
	if actor == "" {
		actor = "population"
	}
	for index, spawn := range plan.Spawns {
		command, err := Command(spawn, actor, uint64(index+1), tick, simulation.AuthoritySystem)
		if err != nil {
			return err
		}
		if err := session.Submit(command); err != nil {
			return err
		}
	}
	return nil
}

func populationCandidates(snapshot gamedata.Snapshot, level models.LevelData, difficulty Difficulty) ([]Definition, error) {
	ids := normalMonsterIDs(level)
	if difficulty != Normal {
		ids = nightmareMonsterIDs(level)
	}
	limit := min(max(level.NumMon, 0), len(ids))
	ids = ids[:limit]
	result := make([]Definition, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		definition, err := FromCatalog(snapshot, id, difficulty)
		if err != nil {
			return nil, fmt.Errorf("monster: population definition %q: %w", id, err)
		}
		result = append(result, definition)
	}
	return result, nil
}

func normalMonsterIDs(level models.LevelData) []string {
	return []string{level.Mon1, level.Mon2, level.Mon3, level.Mon4, level.Mon5, level.Mon6, level.Mon7, level.Mon8, level.Mon9, level.Mon10, level.Mon11, level.Mon12, level.Mon13, level.Mon14, level.Mon15, level.Mon16, level.Mon17, level.Mon18, level.Mon19, level.Mon20, level.Mon21, level.Mon22, level.Mon23, level.Mon24, level.Mon25}
}
func nightmareMonsterIDs(level models.LevelData) []string {
	return []string{level.Nmon1, level.Nmon2, level.Nmon3, level.Nmon4, level.Nmon5, level.Nmon6, level.Nmon7, level.Nmon8, level.Nmon9, level.Nmon10, level.Nmon11, level.Nmon12, level.Nmon13, level.Nmon14, level.Nmon15, level.Nmon16, level.Nmon17, level.Nmon18, level.Nmon19, level.Nmon20, level.Nmon21, level.Nmon22, level.Nmon23, level.Nmon24, level.Nmon25}
}
func populationDensity(level models.LevelData, difficulty Difficulty) int {
	if difficulty == Nightmare {
		return level.MonDenN
	}
	if difficulty == Hell {
		return level.MonDenH
	}
	return level.MonDen
}
func populatedStamps(stamps []mapgen.Stamp) map[uint32]bool {
	result := make(map[uint32]bool, len(stamps))
	for _, stamp := range stamps {
		result[stamp.ID] = stamp.Populate
	}
	return result
}
func totalRarity(definitions []Definition) uint64 {
	total := uint64(0)
	for _, definition := range definitions {
		total += uint64(definition.Rarity)
	}
	return total
}
func weightedDefinition(definitions []Definition, roll uint64) Definition {
	for _, definition := range definitions {
		if roll < uint64(definition.Rarity) {
			return definition
		}
		roll -= uint64(definition.Rarity)
	}
	return definitions[len(definitions)-1]
}
func memberAnchor(room mapgen.Room, member int) (float64, float64) {
	angle := float64(member) * math.Pi / 2
	centerX := float64((room.X+room.Width/2)*subtilesPerTile) + 2
	centerY := float64((room.Y+room.Height/2)*subtilesPerTile) + 2
	return centerX + math.Cos(angle)*float64(member/4), centerY + math.Sin(angle)*float64(member/4)
}
func nearWarp(x, y float64, warps []mapgen.Warp, distance int) bool {
	if distance <= 0 {
		return false
	}
	for _, warp := range warps {
		wx, wy := float64(warp.X*subtilesPerTile)+2.5, float64(warp.Y*subtilesPerTile)+2.5
		if math.Hypot(x-wx, y-wy) < float64(distance*subtilesPerTile) {
			return true
		}
	}
	return false
}
