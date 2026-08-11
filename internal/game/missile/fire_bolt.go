package missile

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	gamemodel "github.com/gravestench/dark-magic/internal/game/data/model"
	gameskill "github.com/gravestench/dark-magic/internal/game/skill"
)

const (
	FireBoltSkillID = int64(36)
	legacyFrameRate = 25
)

// FireBoltFromCatalog promotes the first production missile skill from the
// legacy tables into the small, trusted simulation vocabulary. This is
// intentionally narrow: an unfamiliar server behavior must get its own
// reviewed normalizer instead of being guessed from superficially similar rows.
func FireBoltFromCatalog(snapshot gamedata.Snapshot) (gameskill.Definition, Definition, error) {
	skill, found := snapshot.SkillsByID[strconv.FormatInt(FireBoltSkillID, 10)]
	if !found {
		return gameskill.Definition{}, Definition{}, fmt.Errorf("missile: Fire Bolt skill %d is unavailable", FireBoltSkillID)
	}
	if err := validateFireBoltSkill(skill); err != nil {
		return gameskill.Definition{}, Definition{}, err
	}
	missileID := strings.TrimSpace(skill.SrvMissile)
	missile, found := snapshot.MissilesByName[missileID]
	if !found {
		return gameskill.Definition{}, Definition{}, fmt.Errorf("missile: Fire Bolt references unavailable missile %q", missileID)
	}
	if err := validateFireBoltMissile(missile); err != nil {
		return gameskill.Definition{}, Definition{}, err
	}

	mana, err := shiftedRaw(skill.Mana, skill.ManaShift, "mana")
	if err != nil {
		return gameskill.Definition{}, Definition{}, err
	}
	minimumDamage, err := shiftedRaw(skill.EMin, skill.HitShift, "minimum fire damage")
	if err != nil {
		return gameskill.Definition{}, Definition{}, err
	}
	maximumDamage, err := shiftedRaw(skill.EMax, skill.HitShift, "maximum fire damage")
	if err != nil {
		return gameskill.Definition{}, Definition{}, err
	}
	presentation, err := PresentationFromCatalog(snapshot, missileID)
	if err != nil {
		return gameskill.Definition{}, Definition{}, err
	}

	// Skills.txt gives no separate cooldown for Fire Bolt. The generic cast
	// lifecycle releases its effect on the following authoritative tick and
	// completes one tick later; animation remains a presentation concern.
	cast := gameskill.Definition{
		SkillID: FireBoltSkillID, Behavior: gameskill.BehaviorStraightMissile,
		TargetPolicy: gameskill.TargetPoint, ManaCostRaw: mana,
		EffectDelay: 1, CompleteDelay: 2, Interruptible: true,
	}
	projectile := Definition{
		SkillID: FireBoltSkillID,
		// Missiles.txt velocity is world-coordinate distance per second. The
		// simulation advances at the legacy rate of 25 fixed ticks per second.
		SpeedPerTick:    float64(missile.Vel) / legacyFrameRate,
		LifetimeTicks:   uint64(missile.Range),
		MaxRange:        float64(missile.Vel) / legacyFrameRate * float64(missile.Range),
		CollisionRadius: float64(missile.Size) / 2,
		DamageChannel:   gamecombat.Fire,
		MinimumDamage:   gamecombat.FromRaw(minimumDamage),
		MaximumDamage:   gamecombat.FromRaw(maximumDamage),
		Presentation:    presentation,
	}
	return cast, projectile, nil
}

func validateFireBoltSkill(skill gamemodel.SkillData) error {
	if !strings.EqualFold(strings.TrimSpace(skill.SkillName), "Fire Bolt") ||
		strings.TrimSpace(skill.SrvMissile) != "firebolt" ||
		strings.TrimSpace(skill.EType) != "fire" ||
		strings.TrimSpace(skill.Interrupt) != "1" ||
		strings.TrimSpace(skill.SrvStFunc) != "" ||
		strings.TrimSpace(skill.SrvDoFunc) != "" {
		return fmt.Errorf("missile: skill %d no longer matches the reviewed Fire Bolt behavior", FireBoltSkillID)
	}
	return nil
}

func validateFireBoltMissile(missile gamemodel.Missile) error {
	if strings.TrimSpace(missile.Missile) != "firebolt" ||
		!strings.EqualFold(strings.TrimSpace(missile.Skill), "Fire Bolt") ||
		missile.PSrvDoFunc != "1" || missile.CollideType != 3 || !missile.CollideKill ||
		missile.Vel <= 0 || missile.Range <= 0 || missile.Size <= 0 {
		return fmt.Errorf("missile: record %q no longer matches the reviewed Fire Bolt projectile", missile.Missile)
	}
	return nil
}

func shiftedRaw(value, shift, label string) (int64, error) {
	parsedValue, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsedValue < 0 {
		return 0, fmt.Errorf("missile: invalid Fire Bolt %s %q", label, value)
	}
	parsedShift, err := strconv.ParseUint(strings.TrimSpace(shift), 10, 8)
	if err != nil || parsedShift > 8 || parsedValue > math.MaxInt64>>parsedShift {
		return 0, fmt.Errorf("missile: invalid Fire Bolt %s shift %q", label, shift)
	}
	return parsedValue << parsedShift, nil
}
