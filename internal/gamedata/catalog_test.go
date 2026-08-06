package gamedata

import (
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/recordstore"
)

func TestCatalogBuildsClonedTypedSnapshotAndIndexes(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		CharStatsTable:       &fstest.MapFile{Data: []byte("class\tstr\nAmazon\t20\n")},
		LevelsTable:          &fstest.MapFile{Data: []byte("Name\tId\nRogue Encampment\t1\n")},
		ObjectsTable:         &fstest.MapFile{Data: []byte("Class\tName\n1\tChest\n")},
		SkillsTable:          &fstest.MapFile{Data: []byte("Id\tskill\n0\tAttack\n")},
		SoundsTable:          &fstest.MapFile{Data: []byte("Sound\tFileName\tLoop\nmenu_music\tmusic.wav\t1\n")},
		TreasureClassExTable: &fstest.MapFile{Data: []byte("Treasure Class\tPicks\nAct 1 Good\t1\n")},
		ArmorTable:           &fstest.MapFile{Data: []byte("name\tcode\nCap\tcap\n")},
		WeaponsTable:         &fstest.MapFile{Data: []byte("name\tcode\nHand Axe\thax\n")},
		MiscTable:            &fstest.MapFile{Data: []byte("name\tcode\nHealing Potion\thp1\n")},
		ItemTypesTable:       &fstest.MapFile{Data: []byte("ItemType\tCode\nArmor\tarmo\n")},
		ItemRatioTable:       &fstest.MapFile{Data: []byte("Function\tVersion\nUnique\t100\n")},
		ItemStatCostTable:    &fstest.MapFile{Data: []byte("Stat\tSave Bits\nstrength\t10\n")},
		PropertiesTable:      &fstest.MapFile{Data: []byte("code\tfunc1\tstat1\nstr\t1\tstrength\n")},
		UniqueItemsTable:     &fstest.MapFile{Data: []byte("index\tcode\nThe Gnasher\thax\n")},
		SetItemsTable:        &fstest.MapFile{Data: []byte("index\tset\titem\nCiverb's Ward\tCiverb's Vestments\tsml\n")},
		MagicPrefixTable:     &fstest.MapFile{Data: []byte("Name\tversion\nStrong\t100\n")},
		MagicSuffixTable:     &fstest.MapFile{Data: []byte("Name\tversion\nof Strength\t100\n")},
		AutoMagicTable:       &fstest.MapFile{Data: []byte("Name\tversion\nAmazon Bow\t100\n")},
		RarePrefixTable:      &fstest.MapFile{Data: []byte("name\tversion\nBitter\t100\n")},
		RareSuffixTable:      &fstest.MapFile{Data: []byte("name\tversion\nGrasp\t100\n")},
		GemsTable:            &fstest.MapFile{Data: []byte("name\tcode\nChipped Amethyst\tgcw\n")},
		RunesTable:           &fstest.MapFile{Data: []byte("Name\tcomplete\nAncient's Pledge\t1\n")},
		CubeMainTable:        &fstest.MapFile{Data: []byte("description\tenabled\nPotion Upgrade\t1\n")},
		SetsTable:            &fstest.MapFile{Data: []byte("index\tname\nCiverb's Vestments\tCiverb's Vestments\n")},
		LevelTypesTable:      &fstest.MapFile{Data: []byte("Name\tAct\nAct 1 Town\t1\n")},
		LevelPresetsTable:    &fstest.MapFile{Data: []byte("Name\tDef\tLevelId\nRogue Encampment\t1\t1\n")},
		LevelMazeTable:       &fstest.MapFile{Data: []byte("Name\tLevel\tRooms\nDen of Evil\t8\t4\n")},
		LevelWarpTable:       &fstest.MapFile{Data: []byte("Name\tId\nCave Entrance\t1\n")},
		LevelSubTable:        &fstest.MapFile{Data: []byte("Name\tItemSuperType\nWilderness\t1\n")},
	}
	catalog := New(recordstore.New(source))
	first, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if first.CharStatsByClass["Amazon"].Strength != 20 || first.SoundsByName["menu_music"].Loop != 1 {
		t.Fatalf("typed snapshot = %#v", first)
	}
	if first.LevelsByID[1].Name != "Rogue Encampment" || first.ObjectsByClass[1].Name != "Chest" || first.SkillsByID["0"].SkillName != "Attack" || first.TreasureByName["Act 1 Good"].Picks != 1 {
		t.Fatalf("typed core indexes = %#v", first)
	}
	if first.ArmorByCode["cap"].Name != "Cap" || first.WeaponsByCode["hax"].Name != "Hand Axe" || first.MiscByCode["hp1"].Name != "Healing Potion" || first.ItemTypesByCode["armo"].ItemType != "Armor" {
		t.Fatalf("typed item indexes = %#v", first)
	}
	if len(first.ItemRatios) != 1 || first.ItemStatsByName["strength"].SaveBits != 10 || first.PropertiesByCode["str"].Stat1 != "strength" || first.UniqueByIndex["The Gnasher"].Code != "hax" || first.SetItemsByIndex["Civerb's Ward"].Item != "sml" {
		t.Fatalf("typed item-rule indexes = %#v", first)
	}
	if len(first.MagicPrefixes) != 1 || len(first.MagicSuffixes) != 1 || len(first.AutoMagic) != 1 || len(first.RarePrefixes) != 1 || len(first.RareSuffixes) != 1 {
		t.Fatalf("typed affix tables = %#v", first)
	}
	if first.GemsByCode["gcw"].Name != "Chipped Amethyst" || len(first.RuneWords) != 1 || len(first.CubeRecipes) != 1 || first.SetsByIndex["Civerb's Vestments"].Name != "Civerb's Vestments" {
		t.Fatalf("typed socketing and crafting tables = %#v", first)
	}
	if len(first.LevelTypes) != 1 || first.LevelPresetByDef[1].LevelId != 1 || first.LevelMazeByLevel[8].Rooms != 4 || len(first.LevelWarps) != 1 || len(first.LevelSubs) != 1 {
		t.Fatalf("typed world-generation tables = %#v", first)
	}
	delete(first.CharStatsByClass, "Amazon")
	first.CharStats[0].Strength = 1
	second, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if second.CharStatsByClass["Amazon"].Strength != 20 || second.CharStats[0].Strength != 20 {
		t.Fatal("catalog exposed mutable snapshot ownership")
	}
}

func TestCatalogInvalidationAtomicallyRebuildsTypedData(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		CharStatsTable:       &fstest.MapFile{Data: []byte("class\tstr\nAmazon\t20\n")},
		LevelsTable:          &fstest.MapFile{Data: []byte("Name\tId\nRogue Encampment\t1\n")},
		ObjectsTable:         &fstest.MapFile{Data: []byte("Class\tName\n1\tChest\n")},
		SkillsTable:          &fstest.MapFile{Data: []byte("Id\tskill\n0\tAttack\n")},
		SoundsTable:          &fstest.MapFile{Data: []byte("Sound\tFileName\nmenu_music\tmusic.wav\n")},
		TreasureClassExTable: &fstest.MapFile{Data: []byte("Treasure Class\tPicks\nAct 1 Good\t1\n")},
		ArmorTable:           &fstest.MapFile{Data: []byte("name\tcode\nCap\tcap\n")},
		WeaponsTable:         &fstest.MapFile{Data: []byte("name\tcode\nHand Axe\thax\n")},
		MiscTable:            &fstest.MapFile{Data: []byte("name\tcode\nHealing Potion\thp1\n")},
		ItemTypesTable:       &fstest.MapFile{Data: []byte("ItemType\tCode\nArmor\tarmo\n")},
		ItemRatioTable:       &fstest.MapFile{Data: []byte("Function\tVersion\nUnique\t100\n")},
		ItemStatCostTable:    &fstest.MapFile{Data: []byte("Stat\tSave Bits\nstrength\t10\n")},
		PropertiesTable:      &fstest.MapFile{Data: []byte("code\tfunc1\tstat1\nstr\t1\tstrength\n")},
		UniqueItemsTable:     &fstest.MapFile{Data: []byte("index\tcode\nThe Gnasher\thax\n")},
		SetItemsTable:        &fstest.MapFile{Data: []byte("index\tset\titem\nCiverb's Ward\tCiverb's Vestments\tsml\n")},
		MagicPrefixTable:     &fstest.MapFile{Data: []byte("Name\tversion\nStrong\t100\n")},
		MagicSuffixTable:     &fstest.MapFile{Data: []byte("Name\tversion\nof Strength\t100\n")},
		AutoMagicTable:       &fstest.MapFile{Data: []byte("Name\tversion\nAmazon Bow\t100\n")},
		RarePrefixTable:      &fstest.MapFile{Data: []byte("name\tversion\nBitter\t100\n")},
		RareSuffixTable:      &fstest.MapFile{Data: []byte("name\tversion\nGrasp\t100\n")},
		GemsTable:            &fstest.MapFile{Data: []byte("name\tcode\nChipped Amethyst\tgcw\n")},
		RunesTable:           &fstest.MapFile{Data: []byte("Name\tcomplete\nAncient's Pledge\t1\n")},
		CubeMainTable:        &fstest.MapFile{Data: []byte("description\tenabled\nPotion Upgrade\t1\n")},
		SetsTable:            &fstest.MapFile{Data: []byte("index\tname\nCiverb's Vestments\tCiverb's Vestments\n")},
		LevelTypesTable:      &fstest.MapFile{Data: []byte("Name\tAct\nAct 1 Town\t1\n")},
		LevelPresetsTable:    &fstest.MapFile{Data: []byte("Name\tDef\tLevelId\nRogue Encampment\t1\t1\n")},
		LevelMazeTable:       &fstest.MapFile{Data: []byte("Name\tLevel\tRooms\nDen of Evil\t8\t4\n")},
		LevelWarpTable:       &fstest.MapFile{Data: []byte("Name\tId\nCave Entrance\t1\n")},
		LevelSubTable:        &fstest.MapFile{Data: []byte("Name\tItemSuperType\nWilderness\t1\n")},
	}
	catalog := New(recordstore.New(source))
	if _, err := catalog.Snapshot(); err != nil {
		t.Fatal(err)
	}
	source[CharStatsTable].Data = []byte("class\tstr\nAmazon\t25\n")
	catalog.Invalidate(CharStatsTable)
	updated, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if updated.CharStatsByClass["Amazon"].Strength != 25 {
		t.Fatalf("reloaded strength = %d, want 25", updated.CharStatsByClass["Amazon"].Strength)
	}
}
