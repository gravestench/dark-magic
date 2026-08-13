local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("clamps_resistance_and_preserves_integer_order", {
            test.run(function()
                local mitigation = require("d2legacy.policy.mitigation")
                local defense = {
                    values = {
                        physical_resist = 25,
                        physical_reduction_raw = 20,
                        fire_resist = 120,
                        max_fire_resist = 95,
                    },
                    get = function(self, name)
                        return self.values[name]
                    end,
                }
                test.assert(
                    mitigation.apply(1000, "physical", defense) == 730,
                    [=[mitigation.apply(1000, "physical", defense) == 730]=]
                )
                test.assert(
                    mitigation.apply(1000, "fire", defense) == 50,
                    [=[mitigation.apply(1000, "fire", defense) == 50]=]
                )
                test.assert(
                    mitigation.apply(1000, "cold", defense) == 1000,
                    [=[mitigation.apply(1000, "cold", defense) == 1000]=]
                )
                test.assert(
                    mitigation.apply(1000, "fire", nil) == 1000,
                    [=[mitigation.apply(1000, "fire", nil) == 1000]=]
                )
            end),
        }),
        test.case("rejects_negative_damage", {
            test.run(function()
                local mitigation = require("d2legacy.policy.mitigation")
                local ok, message = pcall(mitigation.apply, -1, "fire", nil)
                test.assert(not ok, [=[not ok]=])
                test.assert(
                    string.find(message, "damage must be non%-negative"),
                    [=[string.find(message, "damage must be non%-negative")]=]
                )
            end),
        }),
    },
})
