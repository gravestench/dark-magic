package movement

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestRulesModuleVitalityAllocationEvidence(t *testing.T) {
	catalog := Catalog{byClass: map[string]ClassRates{
		"amazon":      ownedClassRates(20, 84, 20, 4, 4),
		"sorceress":   ownedClassRates(10, 74, 20, 4, 4),
		"necromancer": ownedClassRates(15, 79, 20, 4, 4),
		"paladin":     ownedClassRates(25, 89, 20, 4, 4),
		"barbarian":   ownedClassRates(25, 92, 20, 4, 4),
		"druid":       ownedClassRates(25, 84, 20, 4, 4),
		"assassin":    ownedClassRates(20, 95, 15, 5, 5),
	}}
	checkVitalityAllocationEvidence(t, catalog)
}

func TestOwnedRulesModuleVitalityAllocationEvidence(t *testing.T) {
	_, catalog := loadOwnedMovementCatalog(t)
	checkVitalityAllocationEvidence(t, catalog)
}

// checkVitalityAllocationEvidence composes existing capabilities without introducing a live allocation API.
func checkVitalityAllocationEvidence(t *testing.T, catalog Catalog) {
	t.Helper()
	state := lua.NewState()
	defer state.Close()
	module := RulesModule(catalog)
	state.PreloadModule(module.Name, module.Loader)
	if err := state.DoFile("testdata/stamina_test.lua"); err != nil {
		t.Fatal(err)
	}
}
