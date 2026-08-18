local test = require("d2legacy.tests/v1")

return test.suite({
    name = "periodic party aura definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "99",
                skill = "Prayer",
                srvstfunc = "",
                srvdofunc = "65",
                aura = "1",
                immediate = "",
                leftskill = "",
                range = "none",
                InGame = "1",
                InTown = "1",
                aurafilter = "73731",
                aurarangecalc = "ln12",
                aurastate = "prayer",
                auratargetstate = "prayer",
                aurastat1 = "hitpoints",
                aurastatcalc1 = "edns",
                mana = "16",
                lvlmana = "3",
                minmana = "1",
                manashift = "4",
                Param1 = "16",
                Param2 = "2",
                perdelay = "50",
                HitShift = "8",
                EMin = "2",
                EMinLev1 = "1",
                EMinLev2 = "1",
                EMinLev3 = "2",
                EMinLev4 = "2",
                EMinLev5 = "3",
            },
            {
                Id = "109",
                skill = "Cleansing",
                srvstfunc = "",
                srvdofunc = "65",
                aura = "1",
                immediate = "",
                leftskill = "",
                range = "none",
                InGame = "1",
                InTown = "1",
                aurafilter = "73731",
                aurarangecalc = "ln12",
                aurastate = "cleansing",
                auratargetstate = "cleansing",
                aurastat1 = "item_poisonlengthresist",
                aurastatcalc1 = "100-dm34",
                aurastat2 = "hitpoints",
                aurastatcalc2 = "skill('Prayer'.edns)",
                mana = "0",
                lvlmana = "0",
                minmana = "0",
                manashift = "8",
                Param1 = "16",
                Param2 = "2",
                Param3 = "30",
                Param4 = "90",
                perdelay = "50",
                HitShift = "8",
            },
        },
        ["data/global/excel/states.txt"] = {
            { state = "prayer", id = "34", aura = "1", stat = "" },
            { state = "cleansing", id = "45", aura = "1", stat = "" },
            { state = "poison", id = "2" },
            { state = "curable", id = "9", curse = "1", curable = "1" },
            { state = "incurable", id = "10", curse = "1", curable = "" },
            { state = "shrine_armor", id = "128", curse = "1", curable = "" },
        },
    },
    cases = {
        test.case("decodes_direct_healing_and_fixed_point_pulse_cost", function(t)
            t:run(function()
                local definition = require("d2legacy.data.periodic_aura_skills").load({ 99 })[99]
                test.expect(definition.behavior):equals("aura.selected-party-periodic")
                test.expect(definition.activation):equals("selected_right")
                test.expect(definition.state_id):equals("prayer")
                test.expect(definition.radius_base):equals(16)
                test.expect(definition.radius_per_level):equals(2)
                test.expect(definition.mana_cost_raw):equals(256)
                test.expect(definition.mana_cost_per_level_raw):equals(48)
                test.expect(definition.minimum_mana_cost_raw):equals(256)
                test.expect(#definition.pulse.effects):equals(1)
                test.expect(definition.pulse.effects[1].kind):equals("heal_life")
                test.expect(definition.pulse.effects[1].value_base):equals(2)
                test.expect(definition.pulse.effects[1].value_per_level[5]):equals(3)
                test.expect(definition.pulse.period_ticks):equals(50)
            end)
        end),
        test.case("decodes_ordered_duration_and_linked_healing_effects", function(t)
            t:run(function()
                local definitions, policy = require("d2legacy.data.periodic_aura_skills").load({ 109 })
                local definition = definitions[109]
                test.expect(definition.mana_cost_raw):equals(0)
                test.expect(#definition.pulse.effects):equals(2)
                test.expect(definition.pulse.effects[1].kind):equals("scale_remaining_timed_state")
                test.expect(definition.pulse.effects[1].progression):equals("one_minus_diminishing")
                test.expect(definition.pulse.effects[1].value_minimum):equals(30)
                test.expect(definition.pulse.effects[1].value_maximum):equals(90)
                test.expect(definition.pulse.effects[2].kind):equals("heal_life")
                test.expect(definition.pulse.effects[2].source_skill_id):equals(99)
                test.expect(definition.pulse.effects[2].level_source):equals("learned_skill")
                test.expect(policy.poison_state_id):equals("poison")
                test.expect(policy.duration_reduced_states.curable):equals(true)
                test.expect(policy.duration_reduced_states.shrine_armor):equals(true)
                test.expect(policy.duration_reduced_states.incurable):equals(nil)
            end)
        end),
    },
})
