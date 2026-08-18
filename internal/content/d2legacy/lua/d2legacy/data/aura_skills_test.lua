local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("decodes_multiple_selected_party_aura_definitions_without_identity_branches", function(t)
            t:run(function()
                local skills = {
                    {
                        Id = "98",
                        skill = "Fixture Might",
                        srvstfunc = "",
                        srvdofunc = "65",
                        aura = "1",
                        immediate = "1",
                        leftskill = "",
                        range = "none",
                        InGame = "1",
                        aurafilter = "73731",
                        aurarangecalc = "ln12",
                        aurastate = "might",
                        auratargetstate = "might",
                        aurastat1 = "damagepercent",
                        aurastatcalc1 = "ln34",
                        mana = "0",
                        lvlmana = "0",
                        Param1 = "16",
                        Param2 = "2",
                        Param3 = "40",
                        Param4 = "10",
                        perdelay = "50",
                    },
                    {
                        Id = "998",
                        skill = "Fixture Defiance",
                        srvstfunc = "",
                        srvdofunc = "65",
                        aura = "1",
                        immediate = "1",
                        leftskill = "",
                        range = "none",
                        InGame = "1",
                        aurafilter = "73731",
                        aurarangecalc = "ln12",
                        aurastate = "defiance",
                        auratargetstate = "defiance",
                        aurastat1 = "skill_armor_percent",
                        aurastatcalc1 = "ln34",
                        mana = "0",
                        lvlmana = "0",
                        Param1 = "8",
                        Param2 = "1",
                        Param3 = "20",
                        Param4 = "5",
                        perdelay = "25",
                    },
                    {
                        Id = "100",
                        skill = "Fixture Resist Fire",
                        srvstfunc = "",
                        srvdofunc = "65",
                        aura = "1",
                        immediate = "1",
                        leftskill = "",
                        range = "none",
                        InGame = "1",
                        aurafilter = "73731",
                        aurarangecalc = "ln12",
                        aurastate = "resistfire",
                        auratargetstate = "resistfire",
                        aurastat1 = "fireresist",
                        aurastatcalc1 = "dm34",
                        aurastat2 = "maxfireresist",
                        aurastatcalc2 = "skill('Fixture Resist Fire'.blvl)",
                        passivestate = "passive_resistfire",
                        passivestat1 = "maxfireresist",
                        passivecalc1 = "skill('Fixture Resist Fire'.blvl)/2",
                        mana = "0",
                        lvlmana = "0",
                        Param1 = "16",
                        Param2 = "2",
                        Param3 = "35",
                        Param4 = "150",
                        perdelay = "50",
                    },
                    {
                        Id = "999",
                        skill = "Fixture Aim",
                        srvstfunc = "",
                        srvdofunc = "65",
                        aura = "1",
                        immediate = "1",
                        leftskill = "",
                        range = "none",
                        InGame = "1",
                        aurafilter = "73731",
                        aurarangecalc = "ln12",
                        aurastate = "blessedaim",
                        auratargetstate = "blessedaim",
                        aurastat1 = "item_tohit_percent",
                        aurastatcalc1 = "ln34",
                        passivestate = "penetrate",
                        passivestat1 = "item_tohit_percent",
                        passivecalc1 = "skill('Fixture Aim'.blvl) * par8",
                        mana = "0",
                        lvlmana = "0",
                        Param1 = "16",
                        Param2 = "2",
                        Param3 = "75",
                        Param4 = "15",
                        Param8 = "5",
                        perdelay = "50",
                    },
                }
                test.mock_module("engine.records/v1", {
                    load = function(path)
                        if path == "data/global/excel/skills.txt" then
                            return skills
                        end
                        return {
                            { state = "might", aura = "1", stat = "damagepercent" },
                            { state = "defiance", aura = "1", stat = "skill_armor_percent" },
                            { state = "resistfire", aura = "1", stat = "fireresist" },
                            { state = "passive_resistfire" },
                            { state = "blessedaim", aura = "1", stat = "item_tohit_percent" },
                            { state = "penetrate" },
                        }
                    end,
                }, { "load" })
                test.unload_module("d2legacy.data.aura_skills")
                local definitions = require("d2legacy.data.aura_skills").load({ 98, 100, 998, 999 })
                test.expect(definitions[98].activation):equals("selected_right")
                test.expect(definitions[98].radius_base):equals(16)
                test.expect(definitions[98].stat_value_base):equals(40)
                test.expect(definitions[98].record_refresh_delay):equals(50)
                test.expect(definitions[998].radius_per_level):equals(1)
                test.expect(definitions[998].stat_value_per_level):equals(5)
                test.expect(definitions[998].stat):equals("defense")
                test.expect(definitions[998].stat_operation):equals("percent")
                test.expect(#definitions[100].stats):equals(2)
                test.expect(definitions[100].stats[1].stat):equals("fire_resist")
                test.expect(definitions[100].stats[1].progression):equals("dm34")
                test.expect(definitions[100].stats[2].stat):equals("max_fire_resist")
                test.expect(definitions[100].stats[2].progression):equals("self_hard_level")
                test.expect(definitions[100].learned_passive.value_numerator):equals(1)
                test.expect(definitions[100].learned_passive.value_divisor):equals(2)
                test.expect(definitions[999].stat):equals("item_tohit_percent")
                test.expect(definitions[999].learned_passive.state_id):equals("penetrate")
                test.expect(definitions[999].learned_passive.stat):equals("item_tohit_percent")
                test.expect(definitions[999].learned_passive.operation):equals("percent")
                test.expect(definitions[999].learned_passive.value_numerator):equals(5)
                test.expect(definitions[999].learned_passive.value_divisor):equals(1)
            end)
        end),
    },
})
