package recordManager

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

	group.GET("Belts", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.belts)
	})

	group.GET("CharStartingAttributes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.charStartingAttributes)
	})

	group.GET("Inventory", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.inventory)
	})

	group.GET("Overlays", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.overlays)
	})

	group.GET("PetTypes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.petTypes)
	})

	group.GET("AutoMapEntries", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.autoMapEntries)
	})

	group.GET("States", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.states)
	})

	group.GET("Hirelings", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.hirelings)
	})

	group.GET("HirelingDescriptions", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.hirelingDescriptions)
	})

	group.GET("Missiles", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.missiles)
	})

	group.GET("DifficultyLevels", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.difficultyLevels)
	})

	group.GET("Shrines", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.shrines)
	})

	group.GET("GambleRecords", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.gambleRecords)
	})

	group.GET("NpcTradeRecords", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.npcTradeRecords)
	})

	group.GET("ExperienceBreakpoints", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.experienceBreakpoints)
	})

	group.GET("ItemArmor", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemArmor)
	})

	group.GET("ItemWeapon", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemWeapon)
	})

	group.GET("ItemWeaponClass", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemWeaponClass)
	})

	group.GET("ItemMisc", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemMisc)
	})

	group.GET("ItemTypes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemTypes)
	})

	group.GET("ItemAutoMagic", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemAutoMagic)
	})

	group.GET("ItemStatCost", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemStatCost)
	})

	group.GET("ItemRatio", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemRatio)
	})

	group.GET("ItemUnique", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemUnique)
	})

	group.GET("ItemHiQualityMods", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemHiQualityMods)
	})

	group.GET("ItemProperties", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.itemProperties)
	})

	group.GET("CubeRecipes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.cubeRecipes)
	})

	group.GET("Books", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.books)
	})

	group.GET("Gems", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.gems)
	})

	group.GET("Runes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.runes)
	})

	group.GET("SetItems", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.setItems)
	})

	group.GET("SetBonuses", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.setBonuses)
	})

	group.GET("Skills", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.skills)
	})

	group.GET("SkillDesc", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.skillDesc)
	})

	group.GET("Treasures", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.treasures)
	})

	group.GET("TreasuresEx", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.treasuresExpansion)
	})

	group.GET("MagicPrefixes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.magicPrefixes)
	})

	group.GET("MagicSuffixes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.magicSuffixes)
	})

	group.GET("RarePrefixes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.rarePrefixes)
	})

	group.GET("RareSuffixes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.rareSuffixes)
	})

	group.GET("UniquePrefixes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.uniquePrefixes)
	})

	group.GET("UniqueSuffixes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.uniqueSuffixes)
	})

	group.GET("Objects", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.objects)
	})

	group.GET("ObjectTypes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.objectTypes)
	})

	group.GET("ObjectGroups", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.objectGroups)
	})

	group.GET("ObjectModes", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.objectModes)
	})

	group.GET("Sounds", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.sounds)
	})

	group.GET("SoundEnvironments", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.soundEnvironments)
	})

	group.GET("LevelPresets", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.levelPresets)
	})

	group.GET("LevelType", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.levelType)
	})

	group.GET("LevelWarp", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.levelWarp)
	})

	group.GET("LevelDetails", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.levelDetails)
	})

	group.GET("LevelMaze", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.levelMaze)
	})

	group.GET("LevelSubstitutions", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.levelSubstitutions)
	})

	group.GET("MonsterUniqueModifiers", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.monsterUniqueModifiers)
	})

	group.GET("MonsterEquipment", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.monsterEquipment)
	})

	group.GET("MonsterLevelStats", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.monsterLevelStats)
	})

	group.GET("MonsterPresets", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.monsterPresets)
	})

	group.GET("MonsterProperties", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.monsterProperties)
	})

	group.GET("MonsterSequences", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.monsterSequences)
	})

	group.GET("MonsterStats", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.monsterStats)
	})

	group.GET("MonsterStats2", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.monsterStats2)
	})

	group.GET("MonsterSounds", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.monsterSounds)
	})

	group.GET("MonsterUniqueNames", func(c *gin.Context) {
		c.JSON(http.StatusOK, &s.monsterUniqueNames)
	})
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
