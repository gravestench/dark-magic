local test = require("d2legacy.tests/v1")

return test.suite({
    name = "periodic party aura definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "99",
                skill = "Fixture Prayer",
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
        },
        ["data/global/excel/states.txt"] = {
            { state = "prayer", id = "34", aura = "1", stat = "" },
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
                test.expect(definition.pulse.kind):equals("heal_life")
                test.expect(definition.pulse.value_base):equals(2)
                test.expect(definition.pulse.value_per_level[5]):equals(3)
                test.expect(definition.pulse.period_ticks):equals(50)
            end)
        end),
    },
})
