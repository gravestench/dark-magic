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
        test.case("applies_the_owned_prayer_direct_effect_bands", function(t)
            t:run(function()
                local progression = require("d2legacy.policy.skill_progression")
                local want = { 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 19, 21, 23, 25 }
                for level, value in ipairs(want) do
                    test.expect(progression.banded(2, { 1, 1, 2, 2, 3 }, level)):equals(value)
                end
            end)
        end),
        test.case("aligns_periodic_effects_to_the_recovered_global_phase", function(t)
            t:run(function()
                local progression = require("d2legacy.policy.skill_progression")
                test.expect(progression.next_periodic_tick(0, 50)):equals(1)
                test.expect(progression.next_periodic_tick(1, 50)):equals(51)
                test.expect(progression.next_periodic_tick(49, 50)):equals(51)
                test.expect(progression.next_periodic_tick(50, 50)):equals(51)
                test.expect(progression.next_periodic_tick(51, 50)):equals(101)
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
                local movement = { 13, 18, 22, 25, 28, 30, 32, 33, 35, 36, 37, 38, 39, 40, 40, 41, 41, 42, 42, 43 }
                for level, value in ipairs(movement) do
                    test.expect(progression.diminishing(7, 50, level)):equals(value)
                    test.expect(progression.linear(50, 25, level)):equals(25 + level * 25)
                end
                local cleansing_reduction =
                    { 39, 46, 51, 56, 60, 63, 65, 67, 69, 70, 72, 73, 75, 76, 76, 78, 78, 79, 79, 80 }
                for level, value in ipairs(cleansing_reduction) do
                    test.expect(progression.diminishing(30, 90, level)):equals(value)
                    test.expect(100 - progression.diminishing(30, 90, level)):equals(100 - value)
                end
                local thorns = {
                    250,
                    290,
                    330,
                    370,
                    410,
                    450,
                    490,
                    530,
                    570,
                    610,
                    650,
                    690,
                    730,
                    770,
                    810,
                    850,
                    890,
                    930,
                    970,
                    1010,
                }
                for level, value in ipairs(thorns) do
                    test.expect(progression.linear(250, 40, level)):equals(value)
                end
                local meditation = {
                    300,
                    325,
                    350,
                    375,
                    400,
                    425,
                    450,
                    475,
                    500,
                    525,
                    550,
                    575,
                    600,
                    625,
                    650,
                    675,
                    700,
                    725,
                    750,
                    775,
                }
                for level, value in ipairs(meditation) do
                    test.expect(progression.linear(300, 25, level)):equals(value)
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
