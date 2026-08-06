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
		MonsterStatsTable:    &fstest.MapFile{Data: []byte("Id\tCode\nzombie\tZM\n")},
		MonsterStats2Table:   &fstest.MapFile{Data: []byte("Id\tHeight\nzombie\t64\n")},
		MonsterLevelsTable:   &fstest.MapFile{Data: []byte("Level\tAC\n1\t100\n")},
		MonsterPropsTable:    &fstest.MapFile{Data: []byte("Id\tprop1\nzombie\tres-all\n")},
		MonsterSoundsTable:   &fstest.MapFile{Data: []byte("Id\tAttack1\nzombie\tzombie_attack\n")},
		MonsterEquipTable:    &fstest.MapFile{Data: []byte("monster\titem1\nzombie\thax\n")},
		MissilesTable:        &fstest.MapFile{Data: []byte("Missile\tVel\narrow\t24\n")},
		StatesTable:          &fstest.MapFile{Data: []byte("state\tgroup\npoison\t1\n")},
		OverlaysTable:        &fstest.MapFile{Data: []byte("overlay\tFilename\npoison\tpoisonoverlay\n")},
		PetTypesTable:        &fstest.MapFile{Data: []byte("pet type\tgroup\nvalkyrie\t1\n")},
		ExperienceTable:      &fstest.MapFile{Data: []byte("Amazon\tExpRatio\n500\t100\n")},
		InventoryTable:       &fstest.MapFile{Data: []byte("class\tgridX\nAmazon\t10\n")},
		BeltsTable:           &fstest.MapFile{Data: []byte("name\tnumboxes\nBelt\t8\n")},
		HirelingTable:        &fstest.MapFile{Data: []byte("Hireling\tId\nRogue Scout\t1\n")},
		DifficultyTable:      &fstest.MapFile{Data: []byte("Name\tResistPenalty\nNormal\t0\n")},
		SkillDescTable:       &fstest.MapFile{Data: []byte("skilldesc\tSkillPage\tstr name\nattack\t1\tstrAttack\n")},
		SoundEnvironTable:    &fstest.MapFile{Data: []byte("Handle\tSong\nAct1\tlevel_music\n")},
		AutoMapTable:         &fstest.MapFile{Data: []byte("LevelName\tTileName\tStyle\tCel1\nAct 1 Town\tfl\t1\t10\n")},
		NPCTable:             &fstest.MapFile{Data: []byte("npc\tbuy mult\nAkara\t1024\n")},
		ShrinesTable:         &fstest.MapFile{Data: []byte("Shrine Type\tShrine name\tCode\nMana\tMana Shrine\t1\n")},
		MonsterPresetsTable:  &fstest.MapFile{Data: []byte("Act\tPlace\n1\tzombie\n")},
		GambleTable:          &fstest.MapFile{Data: []byte("name\tcode\nCap\tcap\n")},
		ObjectTypesTable:     &fstest.MapFile{Data: []byte("Name\tToken\nShrine\tSH\n")},
		ObjectGroupsTable:    &fstest.MapFile{Data: []byte("GroupName\tID0\tDENSITY0\tPROB0\nCaves\t1\t10\t100\n")},
		ObjectModesTable:     &fstest.MapFile{Data: []byte("Name\tToken\nNeutral\tNU\n")},
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
	if first.MonstersByID["zombie"].Code != "ZM" || first.MonsterGfxByID["zombie"].Id == "" || len(first.MonsterLevels) != 1 || first.MonsterPropsByID["zombie"].Prop1 != "res-all" || first.MonsterSoundByID["zombie"].Attack1 != "zombie_attack" || len(first.MonsterEquipment) != 1 {
		t.Fatalf("typed monster tables = %#v", first)
	}
	if len(first.MissilesByName) != 1 || len(first.StatesByName) != 1 || len(first.OverlaysByName) != 1 || len(first.PetTypes) != 1 {
		t.Fatalf("typed combat presentation tables = %#v", first)
	}
	if first.SkillDescByName["attack"].StrName != "strAttack" || first.SoundEnvByHandle["Act1"].Song != "level_music" || len(first.AutoMapEntries) != 1 {
		t.Fatalf("typed presentation support tables = %#v", first)
	}
	if first.NPCTradesByID["Akara"].BuyMult != 1024 || first.ShrinesByType["Mana"].Code != 1 || len(first.MonsterPresets) != 1 || first.GambleItemsByCode["cap"].Name != "Cap" {
		t.Fatalf("typed world-interaction tables = %#v", first)
	}
	if first.ObjectTypesByName["Shrine"].Token != "SH" || first.ObjectGroupsByName["Caves"].ObjectID0 != 1 || first.ObjectModesByName["Neutral"].Token != "NU" {
		t.Fatalf("typed object metadata tables = %#v", first)
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
		MonsterStatsTable:    &fstest.MapFile{Data: []byte("Id\tCode\nzombie\tZM\n")},
		MonsterStats2Table:   &fstest.MapFile{Data: []byte("Id\tHeight\nzombie\t64\n")},
		MonsterLevelsTable:   &fstest.MapFile{Data: []byte("Level\tAC\n1\t100\n")},
		MonsterPropsTable:    &fstest.MapFile{Data: []byte("Id\tprop1\nzombie\tres-all\n")},
		MonsterSoundsTable:   &fstest.MapFile{Data: []byte("Id\tAttack1\nzombie\tzombie_attack\n")},
		MonsterEquipTable:    &fstest.MapFile{Data: []byte("monster\titem1\nzombie\thax\n")},
		MissilesTable:        &fstest.MapFile{Data: []byte("Missile\tVel\narrow\t24\n")},
		StatesTable:          &fstest.MapFile{Data: []byte("state\tgroup\npoison\t1\n")},
		OverlaysTable:        &fstest.MapFile{Data: []byte("overlay\tFilename\npoison\tpoisonoverlay\n")},
		PetTypesTable:        &fstest.MapFile{Data: []byte("pet type\tgroup\nvalkyrie\t1\n")},
		ExperienceTable:      &fstest.MapFile{Data: []byte("Amazon\tExpRatio\n500\t100\n")},
		InventoryTable:       &fstest.MapFile{Data: []byte("class\tgridX\nAmazon\t10\n")},
		BeltsTable:           &fstest.MapFile{Data: []byte("name\tnumboxes\nBelt\t8\n")},
		HirelingTable:        &fstest.MapFile{Data: []byte("Hireling\tId\nRogue Scout\t1\n")},
		DifficultyTable:      &fstest.MapFile{Data: []byte("Name\tResistPenalty\nNormal\t0\n")},
		SkillDescTable:       &fstest.MapFile{Data: []byte("skilldesc\tSkillPage\nattack\t1\n")},
		SoundEnvironTable:    &fstest.MapFile{Data: []byte("Handle\tSong\nAct1\tlevel_music\n")},
		AutoMapTable:         &fstest.MapFile{Data: []byte("LevelName\tTileName\tStyle\nAct 1 Town\tfl\t1\n")},
		NPCTable:             &fstest.MapFile{Data: []byte("npc\tbuy mult\nAkara\t1024\n")},
		ShrinesTable:         &fstest.MapFile{Data: []byte("Shrine Type\tShrine name\tCode\nMana\tMana Shrine\t1\n")},
		MonsterPresetsTable:  &fstest.MapFile{Data: []byte("Act\tPlace\n1\tzombie\n")},
		GambleTable:          &fstest.MapFile{Data: []byte("name\tcode\nCap\tcap\n")},
		ObjectTypesTable:     &fstest.MapFile{Data: []byte("Name\tToken\nShrine\tSH\n")},
		ObjectGroupsTable:    &fstest.MapFile{Data: []byte("GroupName\tID0\tDENSITY0\tPROB0\nCaves\t1\t10\t100\n")},
		ObjectModesTable:     &fstest.MapFile{Data: []byte("Name\tToken\nNeutral\tNU\n")},
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
