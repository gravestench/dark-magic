package main

import "sort"

// normalize derives deterministic case metrics only after validation has established every required phase.
func normalize(observed probeCase) caseReport {
	result := caseReport{
		ID:                   observed.ID,
		Scenario:             observed.Scenario,
		Difficulty:           observed.Difficulty,
		StashedGoldUnchanged: true,
	}
	byDeath := make(map[int]map[string]observation)

	for _, current := range observed.Observations {
		if byDeath[current.DeathIndex] == nil {
			byDeath[current.DeathIndex] = make(map[string]observation)
		}

		byDeath[current.DeathIndex][current.Phase] = current

		if current.Phase == "rejoined" {
			recordRejoinedState(&result, current)
		}
	}

	for _, deathIndex := range sortedDeathIndexes(byDeath) {
		result.Deaths = append(result.Deaths, normalizeDeath(deathIndex, byDeath[deathIndex]))
	}

	return result
}

// recordRejoinedState captures the final ownership snapshot without retaining mutable observation slices.
func recordRejoinedState(result *caseReport, rejoined observation) {
	result.SaveRejoinObserved = true
	result.RejoinedCorpseCount = rejoined.CorpseCount
	result.RejoinedEquipmentCount = len(rejoined.Equipment)
	result.RejoinedInventoryCount = len(rejoined.Inventory)
}

// sortedDeathIndexes prevents Go map iteration order from leaking into the serialized report.
func sortedDeathIndexes(byDeath map[int]map[string]observation) []int {
	indexes := make([]int, 0, len(byDeath))
	for deathIndex := range byDeath {
		indexes = append(indexes, deathIndex)
	}

	sort.Ints(indexes)

	return indexes
}

// normalizeDeath measures one timeline while leaving absent optional phases represented by their zero values.
func normalizeDeath(deathIndex int, phases map[string]observation) deathReport {
	before := phases["before_death"]
	started := phases["death_started"]
	complete := phases["death_animation_complete"]
	result := deathReport{
		DeathIndex:           deathIndex,
		DeathAnimationFrames: complete.Frame - started.Frame,
	}

	recordRespawnMetrics(&result, before, phases)
	recordRecoveryMetrics(&result, before, phases)

	return result
}

// recordRespawnMetrics compares the baseline with restored town control only when both respawn boundaries exist.
func recordRespawnMetrics(result *deathReport, before observation, phases map[string]observation) {
	input, hasInput := phases["respawn_input"]

	town, hasTown := phases["town_control"]
	if !hasInput || !hasTown {
		return
	}

	result.RespawnObserved = true
	result.RespawnInputToControlFrames = town.Frame - input.Frame
	result.ExperienceLoss = max64(0, before.Experience-town.Experience)
	result.CarriedGoldLoss = max64(0, before.CarriedGold-town.CarriedGold)
	result.GroundGoldAfterRespawn = town.GroundGold
	result.CorpseCountAfterRespawn = town.CorpseCount
	result.EquipmentRemovedAtRespawn = missing(before.Equipment, town.Equipment)
	result.InventoryRemovedAtRespawn = missing(before.Inventory, town.Inventory)
}

// recordRecoveryMetrics measures only recovery that follows town control, preserving a missing phase as no recovery.
func recordRecoveryMetrics(result *deathReport, before observation, phases map[string]observation) {
	town, hasTown := phases["town_control"]

	recovered, hasRecovered := phases["corpse_recovered"]
	if !hasTown || !hasRecovered {
		return
	}

	result.RecoveryObserved = true
	result.ExperienceRecovered = max64(0, recovered.Experience-town.Experience)
	result.EquipmentRestored = restored(before.Equipment, town.Equipment, recovered.Equipment)
	result.InventoryRestored = restored(before.Inventory, town.Inventory, recovered.Inventory)
}

// missing counts baseline item identities absent after respawn, independent of their slot placement.
func missing(before, after []slotItem) int {
	afterIDs := itemIDs(after)
	count := 0

	for id := range itemIDs(before) {
		if !afterIDs[id] {
			count++
		}
	}

	return count
}

// restored counts baseline items that were absent after respawn and present again after corpse recovery.
func restored(before, after, recovered []slotItem) int {
	afterIDs := itemIDs(after)
	recoveredIDs := itemIDs(recovered)
	count := 0

	for id := range itemIDs(before) {
		if !afterIDs[id] && recoveredIDs[id] {
			count++
		}
	}

	return count
}

// itemIDs builds a membership set because item identity, rather than slot movement, defines ownership restoration.
func itemIDs(items []slotItem) map[string]bool {
	result := make(map[string]bool, len(items))
	for _, item := range items {
		result[item.ID] = true
	}

	return result
}

// max64 clamps loss and recovery metrics at zero so inconsistent direction cannot produce negative consequences.
func max64(left, right int64) int64 {
	if left > right {
		return left
	}

	return right
}
