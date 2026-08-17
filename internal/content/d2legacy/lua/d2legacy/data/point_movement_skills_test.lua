local test = require("d2legacy.tests/v1")

return test.suite({
    name = "point movement skill definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "54",
                skill = "Teleport",
                srvstfunc = "",
                srvdofunc = "27",
                cltstfunc = "",
                cltdofunc = "",
                anim = "SC",
                range = "none",
                warp = "1",
                interrupt = "1",
                leftskill = "",
                general = "",
                InGame = "1",
                mana = "24",
                lvlmana = "-1",
                manashift = "8",
                minmana = "1",
            },
        },
        ["data/global/excel/Levels.txt"] = {
            { Id = "0", Teleport = "0" },
            { Id = "2", Teleport = "1" },
            { Id = "73", Teleport = "2" },
        },
    },
    cases = {
        test.case("decodes_signed_mana_and_authored_level_policy", function(t)
            t:run(function()
                local definition = require("d2legacy.data.point_movement_skills").load({ 54 })[54]
                test.expect(definition.behavior):equals("movement.point-relocate")
                test.expect(definition.mana_cost_raw):equals(24 * 256)
                test.expect(definition.mana_cost_per_level_raw):equals(-256)
                test.expect(definition.minimum_mana_cost_raw):equals(256)
                test.assert(definition.requires_point_target)
                test.expect(definition.level_policy[0]):equals(0)
                test.expect(definition.level_policy[2]):equals(1)
                test.expect(definition.level_policy[73]):equals(2)
            end)
        end),
    },
})
