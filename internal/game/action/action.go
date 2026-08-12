// Package action owns arbitration between mutually exclusive player actions.
//
// Input adapters should not know which gameplay package implements an action.
// They only ask this package to cancel the old exclusive action when a newer,
// explicit player intent supersedes it.
package action

import "github.com/gravestench/akara"

const (
	AttackApproachComponent  = "dm.combat.attack_approach"
	AttackAnimationComponent = "dm.combat.attack_animation"
)

var exclusiveComponents = []string{AttackApproachComponent, AttackAnimationComponent}

// ActiveExclusive reports whether a gameplay action currently owns the
// player's velocity/animation. An idle input snapshot must not overwrite that
// state; only a new explicit intent may cancel it.
func ActiveExclusive(world *akara.World, entity akara.Entity) bool {
	if world == nil {
		return false
	}
	for _, name := range exclusiveComponents {
		if store, found := akara.GetDynamicStore(world, name); found && store.Has(entity) {
			return true
		}
	}
	return false
}

// MatchesExclusive distinguishes a repeated command from a replacement. Both
// the selected skill and semantic target must match; changing either means the
// player changed their mind and the old action may be canceled.
func MatchesExclusive(world *akara.World, entity akara.Entity, skillID int64, targetID string) bool {
	if world == nil {
		return false
	}
	for _, name := range exclusiveComponents {
		store, found := akara.GetDynamicStore(world, name)
		if !found {
			continue
		}
		component, present := store.Get(entity)
		if !present {
			continue
		}
		activeSkill, skillErr := component.Get("skill_id")
		activeTarget, targetErr := component.Get("target_id")
		return skillErr == nil && targetErr == nil && activeSkill == skillID && activeTarget == targetID
	}
	return false
}

// CancelExclusive removes pending actions which cannot coexist with a new
// movement or interaction intent. The list intentionally lives in one place so
// future actions do not make session input depend on combat implementation.
func CancelExclusive(world *akara.World, entity akara.Entity) {
	if world == nil {
		return
	}
	for _, name := range exclusiveComponents {
		if store, found := akara.GetDynamicStore(world, name); found {
			store.Remove(entity)
		}
	}
}
