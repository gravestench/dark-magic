package missile

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	gamemodel "github.com/gravestench/dark-magic/internal/game/data/model"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestFireBoltFromCatalogNormalizesReviewedLegacyRows(t *testing.T) {
	snapshot := fireBoltSnapshot()
	cast, projectile, err := FireBoltFromCatalog(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if cast.ManaCostRaw != 640 || cast.SkillID != 36 || !cast.Interruptible {
		t.Fatalf("cast = %#v", cast)
	}
	if projectile.SpeedPerTick != 0.8 || projectile.LifetimeTicks != 50 || projectile.MaxRange != 40 || projectile.CollisionRadius != 0.5 {
		t.Fatalf("motion = %#v", projectile)
	}
	if projectile.DamageChannel != gamecombat.Fire || projectile.MinimumDamage.Raw() != 768 || projectile.MaximumDamage.Raw() != 1536 {
		t.Fatalf("damage = %s %d..%d", projectile.DamageChannel, projectile.MinimumDamage.Raw(), projectile.MaximumDamage.Raw())
	}
	if projectile.Presentation.DCC != "data/global/missiles/Firebolt.dcc" || projectile.Presentation.Directions != 16 {
		t.Fatalf("presentation = %#v", projectile.Presentation)
	}
}

func TestRealFireBoltRowsMatchReviewedNormalizer(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the Diablo II MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := gamedata.New(recordstore.New(assets)).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := FireBoltFromCatalog(snapshot); err != nil {
		t.Fatal(err)
	}
}

func TestFireBoltFromCatalogRejectsUnreviewedBehavior(t *testing.T) {
	snapshot := fireBoltSnapshot()
	record := snapshot.SkillsByID["36"]
	record.SrvDoFunc = "99"
	snapshot.SkillsByID["36"] = record
	if _, _, err := FireBoltFromCatalog(snapshot); err == nil {
		t.Fatal("unreviewed server behavior was admitted")
	}
}

func fireBoltSnapshot() gamedata.Snapshot {
	return gamedata.Snapshot{
		SkillsByID: map[string]gamemodel.SkillData{"36": {
			ID: "36", SkillName: "Fire Bolt", SrvMissile: "firebolt", Interrupt: "1",
			Mana: "5", ManaShift: "7", HitShift: "7", EType: "fire", EMin: "6", EMax: "12",
		}},
		MissilesByName: map[string]gamemodel.Missile{"firebolt": {
			Missile: "firebolt", PSrvDoFunc: "1", Vel: 20, Range: 50, Size: 1,
			CollideType: 3, CollideKill: true, Skill: "Fire Bolt", CelFile: "Firebolt",
			AnimSpeed: 16, NumDirections: 16, LoopAnim: 1,
		}},
	}
}
