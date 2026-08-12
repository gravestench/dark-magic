package mapgen

import (
	"fmt"
	"path"
	"strings"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	model "github.com/gravestench/dark-magic/internal/game/data/model"
)

// PresetGenerator is the first typed-table implementation of Generator. It
// proves the table-to-zone recipe boundary; decoding DS1/DT1 bytes remains a
// later materialization step owned by the world layer.
type PresetGenerator struct{ data gamedata.Snapshot }

func NewPresetGenerator(data gamedata.Snapshot) *PresetGenerator {
	return &PresetGenerator{data: data}
}

func (generator *PresetGenerator) Generate(request Request) (*Zone, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	level, found := generator.data.LevelsByID[request.LevelID]
	if !found {
		return nil, fmt.Errorf("%w: level %d is absent from Levels", ErrRequest, request.LevelID)
	}
	if level.Act+1 != int(request.Act) {
		return nil, fmt.Errorf("%w: level %d belongs to act %d, not %d", ErrRequest, request.LevelID, level.Act+1, request.Act)
	}
	if level.DrlgType != 2 {
		return nil, fmt.Errorf("%w: level %d uses DRLG type %d, not preset type 2", ErrRequest, request.LevelID, level.DrlgType)
	}
	preset, found := presetForLevel(generator.data.LevelPresets, request.LevelID)
	if !found {
		return nil, fmt.Errorf("%w: level %d has no LvlPrest definition", ErrRequest, request.LevelID)
	}
	variants := presetFiles(preset)
	if len(variants) == 0 {
		return nil, fmt.Errorf("%w: preset %d has no DS1 variants", ErrZone, preset.Def)
	}
	variant := int(NewStreams(request.Seed).For("preset-variant").Uint64n(uint64(len(variants))))
	width, height := preset.SizeX, preset.SizeY
	if width <= 0 || height <= 0 {
		width, height = levelSize(level, request.Difficulty)
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("%w: preset %d has no positive dimensions", ErrZone, preset.Def)
	}
	tiles, err := maskedTilePaths(generator.data.LevelTypes, level.LevelType, preset.Dt1Mask)
	if err != nil {
		return nil, err
	}
	stamp := Stamp{
		ID: 1, PresetDef: preset.Def, Width: width, Height: height,
		DS1Path: assetPath(variants[variant]), TilePaths: tiles, Variant: variant,
		Populate: preset.Populate != 0, LogicalWalls: preset.Logicals != 0,
	}
	if request.Act == 1 && request.LevelID == 1 {
		stamp.Role = actOneTownRole(stamp.DS1Path)
	}
	return NewZone(Definition{
		Request: request, Kind: Preset, Bounds: Bounds{Width: width, Height: height},
		Stamps: []Stamp{stamp}, Rooms: []Room{{ID: 1, Width: width, Height: height, StampID: 1}},
		Trace: []string{
			fmt.Sprintf("Levels[%d] selected preset DRLG type 2", level.Id),
			fmt.Sprintf("LvlPrest[%d] selected variant %d of %d", preset.Def, variant+1, len(variants)),
			fmt.Sprintf("LvlTypes[%d] mask %#x selected %d DT1 files", level.LevelType, uint32(preset.Dt1Mask), len(tiles)),
		},
	})
}

func presetForLevel(records []model.LevelPreset, levelID int) (model.LevelPreset, bool) {
	for _, record := range records {
		if record.LevelId == levelID {
			return record, true
		}
	}
	return model.LevelPreset{}, false
}

func presetFiles(record model.LevelPreset) []string {
	files := []string{record.File1, record.File2, record.File3, record.File4, record.File5, record.File6}
	// Files=0 does not mean File1-only. Static whole-level presets can still
	// author several alternatives; Rogue Encampment is the important example.
	// In that form every non-zero File field is an eligible variant.
	limit := len(files)
	if record.Files > 0 {
		limit = min(record.Files, len(files))
	}
	result := make([]string, 0, limit)
	for _, value := range files[:limit] {
		value = strings.TrimSpace(value)
		if value != "" && value != "0" {
			result = append(result, value)
		}
	}
	return result
}

func actOneTownRole(ds1Path string) string {
	name := strings.ToLower(path.Base(ds1Path))
	for marker, direction := range map[string]string{
		"townn": "north", "towne": "east", "towns": "south", "townw": "west",
	} {
		if strings.HasPrefix(name, marker) {
			return "act1-town:exit-" + direction
		}
	}
	return "act1-town"
}

func levelSize(level model.LevelData, difficulty Difficulty) (int, int) {
	switch difficulty {
	case Nightmare:
		return level.SizeXN, level.SizeYN
	case Hell:
		return level.SizeXH, level.SizeYH
	default:
		return level.SizeX, level.SizeY
	}
}

func maskedTilePaths(types []model.LevelType, levelType, mask int) ([]string, error) {
	// LvlTypes IDs are their zero-based row positions; row zero is the authored
	// null entry. Levels.txt stores that ID directly.
	if levelType < 0 || levelType >= len(types) {
		return nil, fmt.Errorf("%w: LvlTypes ID %d is outside %d records", ErrZone, levelType, len(types))
	}
	record := types[levelType]
	files := levelTypeFiles(record)
	result := make([]string, 0, len(files))
	for index, value := range files {
		if uint32(mask)&(uint32(1)<<index) == 0 {
			continue
		}
		value = strings.TrimSpace(value)
		if value != "" && value != "0" {
			result = append(result, assetPath(value))
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: LvlTypes ID %d mask %#x selected no DT1 files", ErrZone, levelType, uint32(mask))
	}
	return result, nil
}

func levelTypeFiles(record model.LevelType) []string {
	return []string{
		record.File1, record.File2, record.File3, record.File4, record.File5, record.File6, record.File7, record.File8,
		record.File9, record.File10, record.File11, record.File12, record.File13, record.File14, record.File15, record.File16,
		record.File17, record.File18, record.File19, record.File20, record.File21, record.File22, record.File23, record.File24,
		record.File25, record.File26, record.File27, record.File28, record.File29, record.File30, record.File31, record.File32,
	}
}

func assetPath(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	value = strings.TrimPrefix(value, "/")
	if !strings.HasPrefix(strings.ToLower(value), "data/global/tiles/") {
		value = "data/global/tiles/" + value
	}
	return path.Clean(value)
}
