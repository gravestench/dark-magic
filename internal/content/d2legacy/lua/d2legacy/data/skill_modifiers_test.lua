local test = require("d2legacy.tests/v1")

return test.suite({
    name = "cross-skill modifier definitions",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("resolves_a_hard_level_sum_without_named_behavior", function(t)
            t:run(function()
                local modifiers = require("d2legacy.data.skill_modifiers")
                local skills = {
                    { Id = "36", skill = "Fire Bolt" },
                    { Id = "47", skill = "Fire Ball" },
                    { Id = "56", skill = "Meteor" },
                }
                local row = {
                    EDmgSymPerCalc = "(skill('Fire Ball'.blvl)+skill('Meteor'.blvl))*par8",
                    Param8 = "16",
                }
                local ids, percentage = modifiers.hard_level_sum_percent(
                    row,
                    "EDmgSymPerCalc",
                    "Param8",
                    modifiers.by_name(skills),
                    "fixture"
                )
                test.expect(ids[1]):equals(47)
                test.expect(ids[2]):equals(56)
                test.expect(percentage):equals(16)
            end)
        end),
        test.case("rejects_a_different_expression_shape", function(t)
            t:run(function()
                local modifiers = require("d2legacy.data.skill_modifiers")
                local row = {
                    EDmgSymPerCalc = "skill('Fire Ball'.lvl)*par8",
                    Param8 = "16",
                }
                local ok = pcall(
                    modifiers.hard_level_sum_percent,
                    row,
                    "EDmgSymPerCalc",
                    "Param8",
                    modifiers.by_name({ { Id = "47", skill = "Fire Ball" } }),
                    "fixture"
                )
                test.assert(not ok, [=[unreviewed modifier shape was accepted]=])
            end)
        end),
    },
})
