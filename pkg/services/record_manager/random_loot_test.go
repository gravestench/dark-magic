package record_manager

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/examples/services/config_file"
	"github.com/rs/zerolog"
	lua "github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/loaders/mpqLoader"
	tsv_loader2 "github.com/gravestench/dark-magic/pkg/services/loaders/tsvLoader"
	luaservice "github.com/gravestench/dark-magic/pkg/services/lua"
)

const luaRandomLootPreInit = `
function randomIntn(n)
    -- Generate a random number between 0 and (n-1)
    local randomNumber = math.random(n)
    return randomNumber
end

function round(num)
    local intPart = num / 1
    local fracPart = num - intPart
    if fracPart >= 0.5 then
        return intPart + 1
    end
    return intPart
end

function filterPredicate(inputTable, predicateFunc)
    local filteredTable = {}

    for _, obj in ipairs(inputTable) do
        if predicateFunc(obj) then
            table.insert(filteredTable, obj)
        end
    end

    return filteredTable
end

function filterEmptyStrings(array)
    local filteredArray = {}
    for i = 1, #array do
        local value = array[i]
        if value ~= "" then
            table.insert(filteredArray, i, value)
        end
    end
    return filteredArray
end

function monsterIdEquals(monster, id)
   return monster.BaseId == id
end

function randomPick(array)
    return array[randomIntn(#array)]
end

-- Function to pick an object based on probability from separate arrays
function pickObjectByProbability(drops, probs)
    assert(#drops == #probs, "The number of drops and probabilities must be the same.")

    local sumOfProbabilities = 0

    -- Calculate the sum of probabilities
    for i = 1, #probs do
        sumOfProbabilities = sumOfProbabilities + probs[i]
    end

    -- Generate a random number between 0 and the sum total
    local randomValue = math.random() * sumOfProbabilities

    -- Pick the object based on the rolled value
    local pickedObject = nil
    local currentSum = 0
    for i = 1, #probs do
        currentSum = currentSum + probs[i]
        if randomValue <= currentSum then
            pickedObject = drops[i]
            break
        end
    end

    return randomValue, pickedObject
end

function randomLoot(monsterID, lvlMon, lvlChar, lvlArea)

    -- get the monster record
    local isMonster = function(monster)
        return monsterIdEquals(monster, monsterID)
    end

    -- filter all records with the same base ID
    monsterRecords = filterPredicate(records.MonsterStats, isMonster)
    if (#monsterRecords < 1) then
        return
    end


    -- pick one of the records
    chosenMonsterRecord = randomPick(monsterRecords)

    -- get the treasure record
    local matchTreasure = function(t)
        local r = chosenMonsterRecord
        return t.TreasureClass == r.TreasureClass1
    end

    treasureRecords = filterPredicate(records.Treasures, matchTreasure)
    if #treasureRecords < 1 then
        return
    end


    chosenTreasure = randomPick(treasureRecords)

    local t = chosenTreasure
    drops = {t.Item1, t.Item2, t.Item3, t.Item4, t.Item5, t.Item6, t.Item7, t.Item8}
    probs = {t.Prob1, t.Prob2, t.Prob3, t.Prob4, t.Prob5, t.Prob6, t.Prob7, t.Prob8}

    sum = 0
    for i, j in ipairs(drops) do
        sum = sum + probs[i]
    end


    -- print the drops and their probabilities
    for i, j in ipairs(drops) do
        if (probs[i]+0) < 1 then goto continue end
        chance = round(probs[i]/sum*1000)
        chance = chance / 10
        ::continue::
    end

    picks = {}
    for _ = 1, chosenTreasure.Picks do
        roll, picked = pickObjectByProbability(drops, probs)
        table.insert(picks, picked)
    end

    return picks
end
`

const luaRandomLootInit = `
for _ = 1, 100 do
	randomMonsterRecord = records.MonsterStats[randomIntn(#records.MonsterStats)]
    randomLoot(randomMonsterRecord.BaseId, 50, 50, 50)
end
`

func Benchmark_RandomLoot(b *testing.B) {
	rt := runtime.New("benchmark")
	go rt.Run()

	records := &Service{}
	luaService := &luaservice.Service{}

	rt.Add(&config_file.Service{})
	rt.Add(&tsv_loader2.Service{})
	rt.Add(&mpqLoader.Service{})
	rt.Add(luaService)
	rt.Add(records)

	time.Sleep(time.Second * 6)

	b.Run("native", func(b *testing.B) {
		for i := 0; i < 100; i++ {
			randomMonsterRecord := records.MonsterStats[rand.Intn(len(records.MonsterStats))]
			records.randomLoot(randomMonsterRecord.Id, 50, 50, 50)
		}
	})

	s := luaService.LuaState()
	_ = s.DoString(luaRandomLootPreInit)

	b.Run("lua", func(b *testing.B) {
		_ = s.DoString(luaRandomLootInit)
	})
}

func (s *Service) randomLoot(monsterID string, lvlMon, lvlChar, lvlArea int) []string {
	var rMonsters []models.MonsterStats

	// filter all records with the same base ID
	for _, monster := range s.MonsterStats {
		if monster.Id == monsterID {
			rMonsters = append(rMonsters, monster)
		}
	}

	if len(rMonsters) < 1 {
		return nil
	}

	// pick one of the records
	chosenMonsterRecord := rMonsters[rand.Intn(len(rMonsters))]

	// get the treasure record
	var treasureRecords []models.TreasureClass
	for _, t := range s.Treasures {
		if t.TreasureClass == chosenMonsterRecord.TreasureClass1 {
			treasureRecords = append(treasureRecords, t)
		}
	}

	if len(treasureRecords) < 1 {
		fmt.Println("could not find treasure record")
		return nil
	}

	chosenTreasure := treasureRecords[rand.Intn(len(treasureRecords))]

	drops := []string{chosenTreasure.Item1, chosenTreasure.Item2, chosenTreasure.Item3, chosenTreasure.Item4, chosenTreasure.Item5, chosenTreasure.Item6, chosenTreasure.Item7, chosenTreasure.Item8}
	probs := []float64{chosenTreasure.Prob1, chosenTreasure.Prob2, chosenTreasure.Prob3, chosenTreasure.Prob4, chosenTreasure.Prob5, chosenTreasure.Prob6, chosenTreasure.Prob7, chosenTreasure.Prob8}

	var sum float64
	for i := range drops {
		sum += probs[i]
	}

	// print the drops and their probabilities
	for i := range drops {
		if (probs[i] + 0) < 1 {
			continue
		}
		chance := float64(probs[i]) / float64(sum) * 1000
		chance /= 10
	}

	var picks []string
	for i := 0; i < chosenTreasure.Picks; i++ {
		_, picked, err := pickObjectByProbability(drops, probs)
		if err != nil {
			return nil
		}
		picks = append(picks, picked)
	}

	return picks
}

// Function to pick an object based on probability from separate arrays
func pickObjectByProbability(drops []string, probs []float64) (float64, string, error) {
	if len(drops) != len(probs) {
		return 0, "", errors.New("the number of drops and probabilities must be the same")
	}

	var sumOfProbabilities float64

	// Calculate the sum of probabilities
	for i := 0; i < len(probs); i++ {
		sumOfProbabilities += probs[i]
	}

	// Generate a random number between 0 and the sum total
	rand.Seed(time.Now().UnixNano())
	randomValue := rand.Float64() * float64(sumOfProbabilities)

	// Pick the object based on the rolled value
	var pickedObject string
	var currentSum float64
	for i := 0; i < len(probs); i++ {
		currentSum += probs[i]
		if randomValue <= float64(currentSum) {
			pickedObject = drops[i]
			break
		}
	}

	return randomValue, pickedObject, nil
}

func TestService_BindLogger(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	type args struct {
		logger *zerolog.Logger
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			s.BindLogger(tt.args.logger)
		})
	}
}

func TestService_DependenciesResolved(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			if got := s.DependenciesResolved(); got != tt.want {
				t.Errorf("DependenciesResolved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_ExportToLua(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	type args struct {
		state *lua.LState
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			s.ExportToLua(tt.args.state)
		})
	}
}

func TestService_Init(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	type args struct {
		rt runtime.R
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			s.Init(tt.args.rt)
		})
	}
}

func TestService_InitRoutes(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	type args struct {
		group *gin.RouterGroup
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			s.InitRoutes(tt.args.group)
		})
	}
}

func TestService_IsLoaded(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			if got := s.IsLoaded(); got != tt.want {
				t.Errorf("IsLoaded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_LoadRecords(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			if err := s.LoadRecords(); (err != nil) != tt.wantErr {
				t.Errorf("LoadRecords() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Logger(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *zerolog.Logger
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			if got := s.Logger(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Logger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Name(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_ResolveDependencies(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	type args struct {
		runtime runtime.R
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			s.ResolveDependencies(tt.args.runtime)
		})
	}
}

func TestService_Slug(t *testing.T) {
	type fields struct {
		logger                 *zerolog.Logger
		cfg                    config_file.Dependency
		tsv                    tsv_loader2.Dependency
		Belts                  []models.BeltData
		CharStartingAttributes []models.CharStats
		Inventory              []models.InventoryData
		Overlays               []models.Overlay
		PetTypes               []models.PetType
		AutoMapEntries         []models.AutoMapEntry
		States                 []models.State
		Hirelings              []models.Hireling
		HirelingDescriptions   []models.HirelingDescription
		Missiles               []models.Missile
		DifficultyLevels       []models.Difficultylevel
		Shrines                []models.Shrine
		GambleRecords          []models.GambleRecord
		NpcTradeRecords        []models.NPCTrade
		ExperienceBreakpoints  []models.ExperienceData
		ItemArmor              []models.ItemArmor
		ItemWeapon             []models.ItemWeapon
		ItemWeaponClass        []models.WeaponClass
		ItemMisc               []models.MiscItem
		ItemTypes              []models.ItemType
		ItemAutoMagic          []models.AutoMagicData
		ItemStatCost           []models.ItemStatCost
		ItemRatio              []models.ItemRatio
		ItemUnique             []models.ItemUnique
		ItemHiQualityMods      []models.ItemHighQualityModifiers
		ItemProperties         []models.ItemProperty
		CubeRecipes            []models.CubeRecipe
		Books                  []models.Book
		Gems                   []models.GemData
		Runes                  []models.RuneWordData
		SetItems               []models.SetItemData
		SetBonuses             []models.SetBonusData
		Skills                 []models.SkillData
		SkillDesc              []models.SkillDescData
		Treasures              []models.TreasureClass
		TreasuresExpansion     []models.TreasureClassEx
		MagicPrefixes          []models.MagicPrefix
		MagicSuffixes          []models.MagicSuffix
		RarePrefixes           []models.RarePrefix
		RareSuffixes           []models.RareSuffix
		UniquePrefixes         []models.UniquePrefix
		UniqueSuffixes         []models.UniqueSuffix
		Objects                []models.Object
		ObjectTypes            []models.ObjectType
		ObjectGroups           []models.ObjectGroup
		ObjectModes            []models.ObjectMode
		Sounds                 []models.SoundEntry
		SoundEnvironments      []models.SoundEnvironment
		LevelPresets           []models.LevelPreset
		LevelType              []models.LevelType
		LevelWarp              []models.LevelWarp
		LevelDetails           []models.LevelData
		LevelMaze              []models.LevelMazeData
		LevelSubstitutions     []models.LevelSubstitutionData
		MonsterUniqueModifiers []models.MonsterUniqueModifier
		MonsterEquipment       []models.MonsterEquipment
		MonsterLevelStats      []models.MonsterLevelStats
		MonsterPresets         []models.MonsterPreset
		MonsterProperties      []models.MonsterProp
		MonsterSequences       []models.MonsterSequence
		MonsterStats           []models.MonsterStats
		MonsterStats2          []models.MonsterStats2
		MonsterSounds          []models.MonsterSounds
		MonsterUniqueNames     []models.MonsterUniqueName
		loaded                 bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger:                 tt.fields.logger,
				cfg:                    tt.fields.cfg,
				tsv:                    tt.fields.tsv,
				Belts:                  tt.fields.Belts,
				CharStartingAttributes: tt.fields.CharStartingAttributes,
				Inventory:              tt.fields.Inventory,
				Overlays:               tt.fields.Overlays,
				PetTypes:               tt.fields.PetTypes,
				AutoMapEntries:         tt.fields.AutoMapEntries,
				States:                 tt.fields.States,
				Hirelings:              tt.fields.Hirelings,
				HirelingDescriptions:   tt.fields.HirelingDescriptions,
				Missiles:               tt.fields.Missiles,
				DifficultyLevels:       tt.fields.DifficultyLevels,
				Shrines:                tt.fields.Shrines,
				GambleRecords:          tt.fields.GambleRecords,
				NpcTradeRecords:        tt.fields.NpcTradeRecords,
				ExperienceBreakpoints:  tt.fields.ExperienceBreakpoints,
				ItemArmor:              tt.fields.ItemArmor,
				ItemWeapon:             tt.fields.ItemWeapon,
				ItemWeaponClass:        tt.fields.ItemWeaponClass,
				ItemMisc:               tt.fields.ItemMisc,
				ItemTypes:              tt.fields.ItemTypes,
				ItemAutoMagic:          tt.fields.ItemAutoMagic,
				ItemStatCost:           tt.fields.ItemStatCost,
				ItemRatio:              tt.fields.ItemRatio,
				ItemUnique:             tt.fields.ItemUnique,
				ItemHiQualityMods:      tt.fields.ItemHiQualityMods,
				ItemProperties:         tt.fields.ItemProperties,
				CubeRecipes:            tt.fields.CubeRecipes,
				Books:                  tt.fields.Books,
				Gems:                   tt.fields.Gems,
				Runes:                  tt.fields.Runes,
				SetItems:               tt.fields.SetItems,
				SetBonuses:             tt.fields.SetBonuses,
				Skills:                 tt.fields.Skills,
				SkillDesc:              tt.fields.SkillDesc,
				Treasures:              tt.fields.Treasures,
				TreasuresExpansion:     tt.fields.TreasuresExpansion,
				MagicPrefixes:          tt.fields.MagicPrefixes,
				MagicSuffixes:          tt.fields.MagicSuffixes,
				RarePrefixes:           tt.fields.RarePrefixes,
				RareSuffixes:           tt.fields.RareSuffixes,
				UniquePrefixes:         tt.fields.UniquePrefixes,
				UniqueSuffixes:         tt.fields.UniqueSuffixes,
				Objects:                tt.fields.Objects,
				ObjectTypes:            tt.fields.ObjectTypes,
				ObjectGroups:           tt.fields.ObjectGroups,
				ObjectModes:            tt.fields.ObjectModes,
				Sounds:                 tt.fields.Sounds,
				SoundEnvironments:      tt.fields.SoundEnvironments,
				LevelPresets:           tt.fields.LevelPresets,
				LevelType:              tt.fields.LevelType,
				LevelWarp:              tt.fields.LevelWarp,
				LevelDetails:           tt.fields.LevelDetails,
				LevelMaze:              tt.fields.LevelMaze,
				LevelSubstitutions:     tt.fields.LevelSubstitutions,
				MonsterUniqueModifiers: tt.fields.MonsterUniqueModifiers,
				MonsterEquipment:       tt.fields.MonsterEquipment,
				MonsterLevelStats:      tt.fields.MonsterLevelStats,
				MonsterPresets:         tt.fields.MonsterPresets,
				MonsterProperties:      tt.fields.MonsterProperties,
				MonsterSequences:       tt.fields.MonsterSequences,
				MonsterStats:           tt.fields.MonsterStats,
				MonsterStats2:          tt.fields.MonsterStats2,
				MonsterSounds:          tt.fields.MonsterSounds,
				MonsterUniqueNames:     tt.fields.MonsterUniqueNames,
				loaded:                 tt.fields.loaded,
			}
			if got := s.Slug(); got != tt.want {
				t.Errorf("Slug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genericExportArrayToLua(t *testing.T) {
	type args struct {
		arr   interface{}
		state *lua.LState
	}
	tests := []struct {
		name string
		args args
		want *lua.LTable
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genericExportArrayToLua(tt.args.arr, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genericExportArrayToLua() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genericExportToLua(t *testing.T) {
	type args struct {
		obj   interface{}
		state *lua.LState
	}
	tests := []struct {
		name string
		args args
		want *lua.LTable
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genericExportToLua(tt.args.obj, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genericExportToLua() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serveJSONData(t *testing.T) {
	type args struct {
		c    *gin.Context
		data interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serveJSONData(tt.args.c, tt.args.data)
		})
	}
}

func Test_toLuaValue(t *testing.T) {
	type args struct {
		value interface{}
		state *lua.LState
	}
	tests := []struct {
		name string
		args args
		want lua.LValue
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toLuaValue(tt.args.value, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toLuaValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
