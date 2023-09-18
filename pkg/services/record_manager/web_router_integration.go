package record_manager

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed diabloiidatafileguide.html.gz
var guide []byte

func (s *Service) Slug() string {
	return "records"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", s.extractGzip(guide))
	})

	for subgroup, records := range map[string]any{
		"Belts":                  s.belts,
		"CharStartingAttributes": s.charStartingAttributes,
		"Inventory":              s.inventory,
		"Overlays":               s.overlays,
		"PetTypes":               s.petTypes,
		"AutoMapEntries":         s.autoMapEntries,
		"States":                 s.states,
		"Hirelings":              s.hirelings,
		"HirelingDescriptions":   s.hirelingDescriptions,
		"Missiles":               s.missiles,
		"DifficultyLevels":       s.difficultyLevels,
		"Shrines":                s.shrines,
		"GambleRecords":          s.gambleRecords,
		"NpcTradeRecords":        s.npcTradeRecords,
		"ExperienceBreakpoints":  s.experienceBreakpoints,
		"ItemArmor":              s.itemArmor,
		"ItemWeapon":             s.itemWeapon,
		"ItemWeaponClass":        s.itemWeaponClass,
		"ItemMisc":               s.itemMisc,
		"ItemTypes":              s.itemTypes,
		"ItemAutoMagic":          s.itemAutoMagic,
		"ItemStatCost":           s.itemStatCost,
		"ItemRatio":              s.itemRatio,
		"ItemUnique":             s.itemUnique,
		"ItemHiQualityMods":      s.itemHiQualityMods,
		"ItemProperties":         s.itemProperties,
		"CubeRecipes":            s.cubeRecipes,
		"Books":                  s.books,
		"Gems":                   s.gems,
		"Runes":                  s.runes,
		"SetItems":               s.setItems,
		"SetBonuses":             s.setBonuses,
		"Skills":                 s.skills,
		"SkillDesc":              s.skillDesc,
		"Treasures":              s.treasures,
		"TreasuresEx":            s.treasuresExpansion,
		"MagicPrefixes":          s.magicPrefixes,
		"MagicSuffixes":          s.magicSuffixes,
		"RarePrefixes":           s.rarePrefixes,
		"RareSuffixes":           s.rareSuffixes,
		"UniquePrefixes":         s.uniquePrefixes,
		"UniqueSuffixes":         s.uniqueSuffixes,
		"Objects":                s.objects,
		"ObjectTypes":            s.objectTypes,
		"ObjectGroups":           s.objectGroups,
		"ObjectModes":            s.objectModes,
		"Sounds":                 s.sounds,
		"SoundEnvironments":      s.soundEnvironments,
		"LevelPresets":           s.levelPresets,
		"LevelType":              s.levelType,
		"LevelWarp":              s.levelWarp,
		"LevelDetails":           s.levelDetails,
		"LevelMaze":              s.levelMaze,
		"LevelSubstitutions":     s.levelSubstitutions,
		"MonsterUniqueModifiers": s.monsterUniqueModifiers,
		"MonsterEquipment":       s.monsterEquipment,
		"MonsterLevelStats":      s.monsterLevelStats,
		"MonsterPresets":         s.monsterPresets,
		"MonsterProperties":      s.monsterProperties,
		"MonsterSequences":       s.monsterSequences,
		"MonsterStats":           s.monsterStats,
		"MonsterStats2":          s.monsterStats2,
		"MonsterSounds":          s.monsterSounds,
		"MonsterUniqueNames":     s.monsterUniqueNames,
	} {
		group.GET(subgroup, func(c *gin.Context) {
			serveJSONData(c, records)
		})
	}

}

func serveJSONData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func (s *Service) extractGzip(data []byte) []byte {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		s.logger.Fatal().Msg("ExtractTarGz: NewReader failed")
	}

	out := bytes.NewBuffer([]byte{})

	// Copy the decompressed content to the output file
	_, err = io.Copy(out, r)
	if err != nil {
		s.logger.Fatal().Msgf("extracting file: %v", err)
	}

	return out.Bytes()
}
