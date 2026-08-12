return {
    tests = {
        clamps_resistance_and_preserves_integer_order = {
            {run = function()
                local mitigation = require("d2legacy.policy.mitigation")
                local defense = {
                    values = {
                        physical_resist = 25,
                        physical_reduction_raw = 20,
                        fire_resist = 120,
                        max_fire_resist = 95,
                    },
                    get = function(self, name) return self.values[name] end,
                }
                assert(mitigation.apply(1000, "physical", defense) == 730)
                assert(mitigation.apply(1000, "fire", defense) == 50)
                assert(mitigation.apply(1000, "cold", defense) == 1000)
                assert(mitigation.apply(1000, "fire", nil) == 1000)
            end},
        },
        rejects_negative_damage = {
            {run = function()
                local mitigation = require("d2legacy.policy.mitigation")
                local ok, message = pcall(mitigation.apply, -1, "fire", nil)
                assert(not ok)
                assert(string.find(message, "damage must be non%-negative"))
            end},
        },
    },
}
