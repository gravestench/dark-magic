// Package monster joins legacy monster records into authoritative definitions
// and materializes their session-owned ECS entities.
package monster

import (
	"fmt"
	"math"
	"strings"

	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	models "github.com/gravestench/dark-magic/internal/game/data/model"
)

// Difficulty selects one of the three authored stat columns.
type Difficulty uint8

const (
	Normal Difficulty = iota
	Nightmare
	Hell
)

// Definition is the cohesive gameplay view of the historically split MonStats
// and MonStats2 rows plus one effective MonLvl row. Source IDs remain visible so
// a diagnostic can explain exactly which legacy facts produced the archetype.
type Definition struct {
	ID             string            `json:"id"`
	BaseID         string            `json:"base_id"`
	GraphicsID     string            `json:"graphics_id"`
	NameKey        string            `json:"name_key"`
	MonsterType    string            `json:"monster_type"`
	AI             string            `json:"ai"`
	Token          string            `json:"token"`
	WeaponClass    string            `json:"weapon_class"`
	Level          int64             `json:"level"`
	LifeMin        gamecombat.Amount `json:"life_min"`
	LifeMax        gamecombat.Amount `json:"life_max"`
	Defense        int64             `json:"defense"`
	AttackRating   int64             `json:"attack_rating"`
	PhysicalMin    gamecombat.Amount `json:"physical_min"`
	PhysicalMax    gamecombat.Amount `json:"physical_max"`
	Experience     int64             `json:"experience"`
	ColliderRadius float64           `json:"collider_radius"`
	SelectRadius   float64           `json:"select_radius"`
	Velocity       int64             `json:"velocity"`
	Killable       bool              `json:"killable"`
	MonLvlLevel    int               `json:"mon_level_row"`
}

// FromCatalog resolves the raw legacy halves and their level baseline before
// joining them. Callers never need to know that old Excel forced monster facts
// across MonStats.txt, MonStats2.txt, and MonLvl.txt.
func FromCatalog(snapshot gamedata.Snapshot, id string, difficulty Difficulty) (Definition, error) {
	stats, found := snapshot.MonstersByID[strings.TrimSpace(id)]
	if !found {
		return Definition{}, fmt.Errorf("monster: unknown definition %q", id)
	}
	graphicsID := strings.TrimSpace(stats.MonStatsEx)
	if graphicsID == "" {
		graphicsID = strings.TrimSpace(stats.Id)
	}
	graphics, found := snapshot.MonsterGfxByID[graphicsID]
	if !found {
		return Definition{}, fmt.Errorf("monster: %q lacks MonStats2 row %q", stats.Id, graphicsID)
	}
	var level *models.MonsterLevelStats
	if !stats.NoRatio {
		row, found := snapshot.MonsterLevelByLevel[stats.Level]
		if !found {
			return Definition{}, fmt.Errorf("monster: %q lacks MonLvl row %d", stats.Id, stats.Level)
		}
		level = &row
	}
	return JoinDefinition(stats, graphics, level, difficulty)
}

type difficultyValues struct {
	minLife, maxLife int
	defense, attack  int
	minDamage        int
	maxDamage        int
	experience       int
	levelDefense     int
	levelAttack      int
	levelLife        int
	levelDamage      int
	levelExperience  int
}

// JoinDefinition removes the legacy spreadsheet split from gameplay code. It
// applies only the independently corroborated integer relationship:
//
//	effective = MonLvl baseline * MonStats difficulty ratio / 100
//
// noRatio rows use their MonStats values directly. More advanced player-count,
// quality, boss, and patch rules remain later, explicitly verified layers.
func JoinDefinition(stats models.MonsterStats, graphics models.MonsterStats2, level *models.MonsterLevelStats, difficulty Difficulty) (Definition, error) {
	stats.Id = strings.TrimSpace(stats.Id)
	graphics.Id = strings.TrimSpace(graphics.Id)
	graphicsID := strings.TrimSpace(stats.MonStatsEx)
	if graphicsID == "" {
		graphicsID = stats.Id
	}
	if stats.Id == "" || graphics.Id != graphicsID {
		return Definition{}, fmt.Errorf("monster: MonStats %q expects MonStats2 %q, got %q", stats.Id, graphicsID, graphics.Id)
	}
	if !stats.Enabled || stats.Npc || !stats.IsSpawn {
		return Definition{}, fmt.Errorf("monster: %q is not an enabled ordinary spawn", stats.Id)
	}
	if len(strings.TrimSpace(stats.Code)) != 2 || graphics.SizeX <= 0 || graphics.SizeY <= 0 {
		return Definition{}, fmt.Errorf("monster: %q has invalid presentation or collider facts", stats.Id)
	}
	values, err := selectDifficulty(stats, level, difficulty)
	if err != nil {
		return Definition{}, err
	}
	if !stats.NoRatio {
		values.minLife, err = scale(values.levelLife, values.minLife)
		if err == nil {
			values.maxLife, err = scale(values.levelLife, values.maxLife)
		}
		if err == nil {
			values.defense, err = scale(values.levelDefense, values.defense)
		}
		if err == nil {
			values.attack, err = scale(values.levelAttack, values.attack)
		}
		if err == nil {
			values.minDamage, err = scale(values.levelDamage, values.minDamage)
		}
		if err == nil {
			values.maxDamage, err = scale(values.levelDamage, values.maxDamage)
		}
		if err == nil {
			values.experience, err = scale(values.levelExperience, values.experience)
		}
		if err != nil {
			return Definition{}, fmt.Errorf("monster: scale %q: %w", stats.Id, err)
		}
	}
	if values.minLife <= 0 || values.maxLife < values.minLife || values.defense < 0 || values.attack < 0 || values.minDamage < 0 || values.maxDamage < values.minDamage || values.experience < 0 {
		return Definition{}, fmt.Errorf("monster: %q has invalid effective stats", stats.Id)
	}
	lifeMin, err := gamecombat.FromWhole(int64(values.minLife))
	if err != nil {
		return Definition{}, err
	}
	lifeMax, err := gamecombat.FromWhole(int64(values.maxLife))
	if err != nil {
		return Definition{}, err
	}
	damageMin, err := gamecombat.FromWhole(int64(values.minDamage))
	if err != nil {
		return Definition{}, err
	}
	damageMax, err := gamecombat.FromWhole(int64(values.maxDamage))
	if err != nil {
		return Definition{}, err
	}
	diameter := math.Max(float64(graphics.SizeX), float64(graphics.SizeY))
	effectiveLevel := stats.Level
	if level != nil {
		effectiveLevel = level.Level
	}
	definition := Definition{
		ID: stats.Id, BaseID: strings.TrimSpace(stats.BaseId), GraphicsID: graphics.Id,
		NameKey: strings.TrimSpace(stats.NameStr), MonsterType: strings.TrimSpace(stats.MonType), AI: strings.TrimSpace(stats.AI),
		Token: strings.ToUpper(strings.TrimSpace(stats.Code)), WeaponClass: strings.ToUpper(strings.TrimSpace(graphics.BaseWeapon)),
		Level: int64(effectiveLevel), LifeMin: lifeMin, LifeMax: lifeMax, Defense: int64(values.defense), AttackRating: int64(values.attack),
		PhysicalMin: damageMin, PhysicalMax: damageMax, Experience: int64(values.experience),
		ColliderRadius: diameter / 2, SelectRadius: diameter / 2, Velocity: int64(stats.Velocity), Killable: stats.Killable,
	}
	if level != nil {
		definition.MonLvlLevel = level.Level
	}
	return definition, nil
}

func scale(base, ratio int) (int, error) {
	value, err := gamecombat.MultiplyDivide(int64(base), int64(ratio), 100, gamecombat.RoundTowardZero)
	if err != nil || value > math.MaxInt || value < math.MinInt {
		return 0, fmt.Errorf("integer ratio overflows")
	}
	return int(value), nil
}

func selectDifficulty(stats models.MonsterStats, level *models.MonsterLevelStats, difficulty Difficulty) (difficultyValues, error) {
	if difficulty > Hell {
		return difficultyValues{}, fmt.Errorf("monster: invalid difficulty %d", difficulty)
	}
	if !stats.NoRatio && level == nil {
		return difficultyValues{}, fmt.Errorf("monster: %q requires a MonLvl row", stats.Id)
	}
	values := difficultyValues{}
	switch difficulty {
	case Normal:
		values.minLife, values.maxLife, values.defense, values.attack = stats.MinHP, stats.MaxHP, stats.AC, stats.A1TH
		values.minDamage, values.maxDamage, values.experience = stats.A1MinD, stats.A1MaxD, stats.Exp
		if level != nil {
			values.levelDefense, values.levelAttack, values.levelLife = level.Defense, level.AttackRating, level.Life
			values.levelDamage, values.levelExperience = level.Damage, level.Experience
		}
	case Nightmare:
		values.minLife, values.maxLife, values.defense, values.attack = stats.MinHPN, stats.MaxHPN, stats.ACN, stats.A1THN
		values.minDamage, values.maxDamage, values.experience = stats.A1MinDN, stats.A1MaxDN, stats.ExpN
		if level != nil {
			values.levelDefense, values.levelAttack, values.levelLife = level.DefenseN, level.AttackRatingN, level.LifeN
			values.levelDamage, values.levelExperience = level.DamageN, level.ExperienceN
		}
	case Hell:
		values.minLife, values.maxLife, values.defense, values.attack = stats.MinHPH, stats.MaxHPH, stats.ACH, stats.A1THH
		values.minDamage, values.maxDamage, values.experience = stats.A1MinDH, stats.A1MaxDH, stats.ExpH
		if level != nil {
			values.levelDefense, values.levelAttack, values.levelLife = level.DefenseH, level.AttackRatingH, level.LifeH
			values.levelDamage, values.levelExperience = level.DamageH, level.ExperienceH
		}
	}
	return values, nil
}
