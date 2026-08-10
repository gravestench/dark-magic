package item

import (
	"fmt"
	"strings"
)

// ServiceRule is server-owned quest/vendor service data. Lua submits only ID;
// authority resolves which authored sockets supply target and materials.
type ServiceRule struct {
	ID           string
	TargetSlot   string
	ConsumeSlots []string
	GoldCost     int64
}

type ServiceCatalog map[string]ServiceRule

func (catalog ServiceCatalog) Rule(id string) (ServiceRule, error) {
	normalizedID := strings.ToLower(strings.TrimSpace(id))
	rule, found := catalog[normalizedID]
	if !found {
		return ServiceRule{}, fmt.Errorf("item: unknown service %q", id)
	}
	rule.ID = strings.TrimSpace(rule.ID)
	if rule.ID == "" {
		rule.ID = normalizedID
	}
	rule.TargetSlot = strings.TrimSpace(rule.TargetSlot)
	if rule.TargetSlot == "" || rule.GoldCost < 0 {
		return ServiceRule{}, fmt.Errorf("item: service %q is invalid", id)
	}
	seen := map[string]struct{}{rule.TargetSlot: {}}
	for index, slot := range rule.ConsumeSlots {
		slot = strings.TrimSpace(slot)
		if slot == "" {
			return ServiceRule{}, fmt.Errorf("item: service %q has an empty material slot", id)
		}
		rule.ConsumeSlots[index] = slot
		if _, duplicate := seen[slot]; duplicate {
			return ServiceRule{}, fmt.Errorf("item: service %q repeats slot %q", id, slot)
		}
		seen[slot] = struct{}{}
	}
	return rule, nil
}

func (state *State) completeService(rule ServiceRule) (string, error) {
	targetID := state.questSlotItem(rule.TargetSlot)
	if targetID == "" {
		return "", fmt.Errorf("item: service target slot %q is empty", rule.TargetSlot)
	}
	consumed := make([]string, 0, len(rule.ConsumeSlots))
	for _, slot := range rule.ConsumeSlots {
		id := state.questSlotItem(slot)
		if id == "" {
			return "", fmt.Errorf("item: service material slot %q is empty", slot)
		}
		consumed = append(consumed, id)
	}
	if state.layout.Gold.Carried < rule.GoldCost {
		return "", fmt.Errorf("item: insufficient carried gold")
	}
	// Validation above observes the old state. Only after every requirement is
	// satisfied do we mutate target, consume materials, and debit gold together.
	target := state.items[targetID]
	target.AppliedServices = append(append([]string(nil), target.AppliedServices...), rule.ID)
	state.items[targetID] = target
	for _, id := range consumed {
		delete(state.items, id)
		delete(state.placements, id)
	}
	state.layout.Gold.Carried -= rule.GoldCost
	return targetID, nil
}

func (state *State) questSlotItem(slot string) string {
	for id, placement := range state.placements {
		if placement.Container == ContainerQuest && placement.Slot == slot {
			return id
		}
	}
	return ""
}
