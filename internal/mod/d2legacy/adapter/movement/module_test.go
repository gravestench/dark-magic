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
	if err := state.DoString(`
local rules = require("d2legacy.movement_rules/v1")
local walk, run = rules.rates("Amazon", 23, 0)
assert(walk / 25 * 65536 == 30224 and run / 25 * 65536 == 42512)
for _, elapsed in ipairs({0.04, 0.08}) do
    local x, y, moving = rules.velocity(10, 20, "Amazon", {running = true, target = {x = 10.3, y = 20}}, 0, 0, elapsed)
    assert(math.abs(x * elapsed - 0.3) < 1e-12 and y == 0 and moving)
end
`); err != nil {
		t.Fatal(err)
	}
}
