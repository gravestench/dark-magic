local test = require("d2legacy.tests/v1")

return test.suite({
    name = "test harness contract",
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("a_case_state_is_private", function(t)
            t:run(function()
                _G.lua_test_private_value = "first case"
                test.expect({ answer = 42 }, "deep comparison"):deep_equals({ answer = 42 })
            end)
        end),
        test.case("b_case_gets_a_fresh_vm", function(t)
            t:run(function()
                test.expect(_G.lua_test_private_value, "global leaked from another case"):is_nil()
            end)
        end),
        test.property("generated_cases_are_reproducible", {
            seeds = { 1, 42 },
            generator = test.generators.integer(-100, 100),
        }, function(t, value)
            t:run(function()
                test.expect(value * 2, "seeded example"):equals(value + value)
            end)
        end),
    },
})
