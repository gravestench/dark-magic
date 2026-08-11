package monster

import (
	"testing"

	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	models "github.com/gravestench/dark-magic/internal/game/data/model"
)

func ordinaryFixture() (models.MonsterStats, models.MonsterStats2, models.MonsterLevelStats) {
	stats := models.MonsterStats{
		Id: "fallen", BaseId: "fallen", MonStatsEx: "fallen2", NameStr: "FallenName", MonType: "fallen", AI: "Fallen", Code: "FA", MonSound: "fallen_sound",
		Enabled: true, IsSpawn: true, Killable: true, Level: 2, Velocity: 6,
		Aidel: 3, Aidist: 12,
		MinHP: 101, MaxHP: 186, AC: 100, A1TH: 100, A1MinD: 100, A1MaxD: 200, Exp: 100,
		MinHPN: 200, MaxHPN: 300, ACN: 150, A1THN: 160, A1MinDN: 170, A1MaxDN: 180, ExpN: 190,
	}
	graphics := models.MonsterStats2{Id: "fallen2", SizeX: 2, SizeY: 3, MeleeRng: 2, BaseWeapon: "hth", HD: true, HDv: "lit", TR: true, TRv: "med", SH: false, SHv: "lit"}
	level := models.MonsterLevelStats{Level: 2, Life: 9, Defense: 8, AttackRating: 7, Damage: 6, Experience: 5, LifeN: 10, DefenseN: 20, AttackRatingN: 30, DamageN: 40, ExperienceN: 50}
	return stats, graphics, level
}

func TestJoinDefinitionCombinesLegacyMonsterTables(t *testing.T) {
	stats, graphics, level := ordinaryFixture()
	definition, err := JoinDefinition(stats, graphics, &level, Normal)
	if err != nil {
		t.Fatal(err)
	}
	if definition.ID != "fallen" || definition.GraphicsID != "fallen2" || definition.MonLvlLevel != 2 {
		t.Fatalf("source identity = %#v", definition)
	}
	if definition.LifeMin != gamecombat.MustWhole(9) || definition.LifeMax != gamecombat.MustWhole(16) {
		t.Fatalf("life = %v..%v, want 9..16", definition.LifeMin, definition.LifeMax)
	}
	if definition.Defense != 8 || definition.AttackRating != 7 || definition.PhysicalMin != gamecombat.MustWhole(6) || definition.PhysicalMax != gamecombat.MustWhole(12) || definition.Experience != 5 {
		t.Fatalf("effective stats = %#v", definition)
	}
	if definition.ColliderRadius != 1.5 || definition.Token != "FA" || definition.WeaponClass != "HTH" {
		t.Fatalf("presentation = %#v", definition)
	}
	if definition.Components["HD"] != "LIT" || definition.Components["TR"] != "MED" || definition.Components["SH"] != "" || len(definition.Components) != 2 {
		t.Fatalf("joined components = %#v", definition.Components)
	}
	if definition.ThinkInterval != 3 || definition.AggroRadius != 12 || definition.AttackRange != 2 {
		t.Fatalf("AI definition = %#v", definition)
	}
}

func TestJoinDefinitionSelectsDifficultyColumns(t *testing.T) {
	stats, graphics, level := ordinaryFixture()
	definition, err := JoinDefinition(stats, graphics, &level, Nightmare)
	if err != nil {
		t.Fatal(err)
	}
	if definition.LifeMin != gamecombat.MustWhole(20) || definition.LifeMax != gamecombat.MustWhole(30) || definition.Defense != 30 || definition.AttackRating != 48 || definition.Experience != 95 {
		t.Fatalf("nightmare definition = %#v", definition)
	}
}

func TestFromCatalogHidesLegacySplit(t *testing.T) {
	stats, graphics, level := ordinaryFixture()
	snapshot := gamedata.Snapshot{
		MonstersByID:        map[string]models.MonsterStats{"fallen": stats},
		MonsterGfxByID:      map[string]models.MonsterStats2{"fallen2": graphics},
		MonsterLevelByLevel: map[int]models.MonsterLevelStats{2: level},
		MonsterSoundByID:    map[string]models.MonsterSounds{"fallen_sound": {ID: "fallen_sound", DeathSound: "fallen_death"}},
	}
	definition, err := FromCatalog(snapshot, "fallen", Normal)
	if err != nil || definition.ID != "fallen" || definition.LifeMax != gamecombat.MustWhole(16) || definition.DeathSound != "fallen_death" {
		t.Fatalf("definition = %#v, err = %v", definition, err)
	}
}

func TestJoinDefinitionNoRatioUsesDirectValues(t *testing.T) {
	stats, graphics, _ := ordinaryFixture()
	stats.NoRatio = true
	definition, err := JoinDefinition(stats, graphics, nil, Normal)
	if err != nil {
		t.Fatal(err)
	}
	if definition.LifeMin != gamecombat.MustWhole(101) || definition.Defense != 100 {
		t.Fatalf("direct definition = %#v", definition)
	}
}
