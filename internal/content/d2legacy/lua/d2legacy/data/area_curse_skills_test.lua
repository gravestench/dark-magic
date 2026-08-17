local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "66",
                skill = "Amplify Damage",
                srvstfunc = "",
                srvdofunc = "30",
                aurafilter = "3",
                auratargetstate = "amplifydamage",
                auralencalc = "ln34",
                aurarangecalc = "ln12",
                aurastat1 = "damageresist",
                aurastatcalc1 = "-par5",
                range = "none",
                LineOfSight = "4",
                mana = "4",
                lvlmana = "0",
                minmana = "1",
                manashift = "8",
                interrupt = "1",
                Param1 = "3",
                Param2 = "1",
                Param3 = "200",
                Param4 = "75",
                Param5 = "100",
                InGame = "1",
            },
            {
                Id = "72",
                skill = "Weaken",
                srvstfunc = "",
                srvdofunc = "30",
                aurafilter = "3",
                auratargetstate = "weaken",
                auralencalc = "ln34",
                aurarangecalc = "ln12",
                aurastat1 = "damagepercent",
                aurastatcalc1 = "-par5",
                range = "none",
                LineOfSight = "4",
                mana = "4",
                lvlmana = "0",
                minmana = "1",
                manashift = "8",
                interrupt = "1",
                Param1 = "9",
                Param2 = "1",
                Param3 = "350",
                Param4 = "60",
                Param5 = "33",
                InGame = "1",
            },
        },
        ["data/global/excel/states.txt"] = {
            { state = "amplifydamage", id = "9", curse = "1" },
            { state = "weaken", id = "19", curse = "1" },
        },
    },
    cases = {
        test.case("decodes_point_area_duration_stat_and_replacement_policy", function(t)
            t:run(function()
                local definitions = require("d2legacy.data.area_curse_skills").load({ 66, 72 })
                local definition = definitions[66]
                test.expect(definition.behavior):equals("state.point-area-curse")
                test.expect(definition.mana_cost_raw):equals(4 * 256)
                test.expect(definition.radius_base):equals(3)
                test.expect(definition.radius_per_level):equals(1)
                test.expect(definition.duration_base):equals(200)
                test.expect(definition.duration_per_level):equals(75)
                test.expect(definition.stat):equals("physical_resist")
                test.expect(definition.stat_value):equals(-100)
                test.expect(definition.immune_divisor):equals(5)
                test.assert(definition.requires_point_target and definition.reject_lower_priority)
                local weaken = definitions[72]
                test.expect(weaken.stat):equals("damagepercent")
                test.expect(weaken.stat_operation):equals("percent")
                test.expect(weaken.stat_value):equals(-33)
                test.expect(weaken.immune_divisor):equals(0)
                test.expect(weaken.radius_base):equals(9)
                test.expect(weaken.duration_base):equals(350)
            end)
        end),
    },
})
