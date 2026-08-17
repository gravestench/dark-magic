local test = require("d2legacy.tests/v1")

return test.suite({
    name = "skill record progressions",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("uses_all_five_authored_damage_bands", function(t)
            t:run(function()
                local progression = require("d2legacy.policy.skill_progression")
                local gains = { 1, 10, 100, 1000, 10000 }
                local cases = {
                    { 1, 5 },
                    { 8, 12 },
                    { 9, 22 },
                    { 16, 92 },
                    { 17, 192 },
                    { 22, 692 },
                    { 23, 1692 },
                    { 28, 6692 },
                    { 29, 16692 },
                    { 30, 26692 },
                }
                for _, item in ipairs(cases) do
                    test.expect(progression.damage(5, gains, item[1])):equals(item[2])
                end
            end)
        end),
        test.case("applies_linear_mana_with_an_authored_floor", function(t)
            t:run(function()
                local progression = require("d2legacy.policy.skill_progression")
                local decreasing = {
                    mana_cost_raw = 15 * 256,
                    mana_cost_per_level_raw = -256,
                    minimum_mana_cost_raw = 5 * 256,
                }
                test.expect(progression.mana_cost(decreasing, 1)):equals(15 * 256)
                test.expect(progression.mana_cost(decreasing, 6)):equals(10 * 256)
                test.expect(progression.mana_cost(decreasing, 30)):equals(5 * 256)
            end)
        end),
    },
})
