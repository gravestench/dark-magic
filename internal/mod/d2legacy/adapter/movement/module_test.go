package movement

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestRulesModuleUsesActualElapsedStep(t *testing.T) {
	state := lua.NewState()
	defer state.Close()
	module := RulesModule(Catalog{byClass: map[string]ClassRates{"amazon": {Walk: 6, Run: 9}}})
	state.PreloadModule(module.Name, module.Loader)
	if err := state.DoFile("testdata/module_test.lua"); err != nil {
		t.Fatal(err)
	}
}
