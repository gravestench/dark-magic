local test = require("d2legacy.tests/v1")

return test.suite({
    name = "corpse-periodic aura definitions",
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "124",
                skill = "Fixture Corpse Aura",
                srvstfunc = "",
                srvdofunc = "82",
                aura = "1",
                immediate = "",
                leftskill = "",
                range = "none",
                InGame = "1",
                InTown = "1",
                aurafilter = "4354",
                aurarangecalc = "ln12",
                aurastate = "corpse_aura",
                auratargetstate = "",
                calc1 = "dm34",
                calc2 = "ln56",
                calc3 = "ln56",
                mana = "0",
                lvlmana = "0",
                minmana = "0",
                manashift = "8",
                Param1 = "16",
                Param2 = "0",
                Param3 = "10",
                Param4 = "100",
                Param5 = "25",
                Param6 = "5",
                perdelay = "50",
            },
        },
        ["data/global/excel/states.txt"] = {
            { state = "corpse_aura", id = "50", aura = "1" },
        },
    },
    cases = {
        test.case("decodes_owner_state_and_ordered_corpse_operations", function(t)
            t:run(function()
                local definition = require("d2legacy.data.corpse_aura_skills").load({ 124 })[124]
                test.expect(definition.behavior):equals("aura.selected-corpse-periodic")
                test.expect(definition.activation):equals("selected_right")
                test.expect(definition.state_id):equals("corpse_aura")
                test.expect(definition.target_state_id):equals("corpse_aura")
                test.expect(definition.target_policy):equals("owner")
                test.expect(definition.radius_base):equals(16)
                test.expect(definition.radius_per_level):equals(0)
                test.expect(definition.mana_cost_raw):equals(0)
                test.expect(definition.pulse.target_policy):equals("eligible_corpses")
                test.expect(definition.pulse.chance_progression):equals("dm34")
                test.expect(definition.pulse.chance_minimum):equals(10)
                test.expect(definition.pulse.chance_maximum):equals(100)
                test.expect(definition.pulse.period_ticks):equals(50)
                test.expect(#definition.pulse.effects):equals(3)
                test.expect(definition.pulse.effects[1].kind):equals("restore_owner_life")
                test.expect(definition.pulse.effects[1].value_base):equals(25 * 256)
                test.expect(definition.pulse.effects[1].value_per_level):equals(5 * 256)
                test.expect(definition.pulse.effects[2].kind):equals("restore_owner_mana")
                test.expect(definition.pulse.effects[3].kind):equals("consume_corpse")
            end)
        end),
    },
})
