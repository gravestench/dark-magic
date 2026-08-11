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
