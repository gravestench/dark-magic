package clientapp

import (
	"strings"

	cof "github.com/gravestench/cof"
	cofpkg "github.com/gravestench/cof/pkg"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
)

// combatTimingAdapter is the narrow normalization seam between legacy asset
// metadata and gameplay. Combat receives plain integers and never imports a
// codec or knows where AnimData.d2 was mounted.
type combatTimingAdapter struct {
	records map[string]gamecombat.AttackTiming
}

func newCombatTimingAdapter(data *cof.AnimationData) combatTimingAdapter {
	adapter := combatTimingAdapter{records: make(map[string]gamecombat.AttackTiming)}
	if data == nil {
		return adapter
	}
	for _, name := range data.GetRecordNames() {
		record := data.GetRecord(name)
		if record == nil {
			continue
		}
		impact := firstAttackFrame(record)
		timing := gamecombat.AttackTiming{
			Frames: int64(record.FramesPerDirection()), Speed: int64(record.Speed()), ImpactFrame: impact,
		}
		if timing.Frames > 0 && timing.Speed > 0 && impact >= 0 && impact < timing.Frames {
			adapter.records[strings.ToUpper(name)] = timing
		}
	}
	return adapter
}

func (adapter combatTimingAdapter) AttackTiming(token, weaponClass string) (gamecombat.AttackTiming, bool) {
	timing, found := adapter.records[strings.ToUpper(token+"A1"+weaponClass)]
	return timing, found
}

func firstAttackFrame(record *cof.AnimationDataRecord) int64 {
	impact := int64(-1)
	for frame, event := range record.Events() {
		if event == cofpkg.EventAttack && (impact < 0 || int64(frame) < impact) {
			impact = int64(frame)
		}
	}
	return impact
}
