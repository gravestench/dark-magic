package movement

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestRulesModuleVitalityAllocationEvidence(t *testing.T) {
	checkVitalityAllocationEvidence(t, ownedClassCatalog())
}

func TestOwnedRulesModuleVitalityAllocationEvidence(t *testing.T) {
	_, catalog := loadOwnedMovementCatalog(t)
	checkVitalityAllocationEvidence(t, catalog)
}

// checkVitalityAllocationEvidence exercises the production allocation rule without a live spending transaction.
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
