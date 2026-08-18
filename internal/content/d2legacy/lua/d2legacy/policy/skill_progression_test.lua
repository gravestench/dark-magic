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
        test.case("uses_staged_integer_arithmetic_for_dm_parameter_pairs", function(t)
            t:run(function()
                local progression = require("d2legacy.policy.skill_progression")
                local expected =
                    { 52, 66, 76, 85, 92, 98, 102, 106, 110, 113, 116, 118, 121, 123, 124, 127, 128, 129, 130, 131 }
                for level, value in ipairs(expected) do
                    test.expect(progression.diminishing(35, 150, level)):equals(value)
                end
                local resist_all =
                    { 60, 68, 75, 80, 85, 88, 91, 93, 96, 97, 99, 101, 102, 103, 104, 106, 106, 107, 108, 108 }
                for level, value in ipairs(resist_all) do
                    test.expect(progression.diminishing(50, 120, level)):equals(value)
                end
            end)
        end),
        test.case("applies_a_snapshotted_cross_skill_damage_percentage", function(t)
            t:run(function()
                local progression = require("d2legacy.policy.skill_progression")
                local definition = {
                    minimum_damage_raw = 768,
                    maximum_damage_raw = 1536,
                    minimum_damage_per_level_raw = { 0, 0, 0, 0, 0 },
                    maximum_damage_per_level_raw = { 0, 0, 0, 0, 0 },
                }
                local minimum, maximum = progression.damage_range(definition, 1, 32)
                test.expect(minimum):equals(1013)
                test.expect(maximum):equals(2027)
            end)
        end),
    },
})
