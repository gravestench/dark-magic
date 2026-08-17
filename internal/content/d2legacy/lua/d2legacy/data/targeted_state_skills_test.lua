local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            { Id = "37", skill = "Warmth" },
            {
                Id = "52",
                skill = "Enchant",
                srvstfunc = "",
                srvdofunc = "25",
                range = "none",
                TargetPet = "1",
                TargetAlly = "1",
                leftskill = "",
                general = "",
                InGame = "1",
                mana = "25",
                lvlmana = "1",
                manashift = "8",
                minmana = "1",
                aurastate = "enchant",
                auralencalc = "ln12",
                aurastat1 = "firemindam",
                aurastatcalc1 = "enma",
                aurastat2 = "firemaxdam",
                aurastatcalc2 = "exma",
                aurastat3 = "item_tohit_percent",
                aurastatcalc3 = "toht",
                EType = "fire",
                HitShift = "7",
                EMin = "16",
                EMax = "20",
                EMinLev1 = "3",
                EMinLev2 = "7",
                EMinLev3 = "11",
                EMinLev4 = "15",
                EMinLev5 = "19",
                EMaxLev1 = "5",
                EMaxLev2 = "9",
                EMaxLev3 = "13",
                EMaxLev4 = "17",
                EMaxLev5 = "21",
                EDmgSymPerCalc = "(skill('Warmth'.blvl))*par8",
                Param1 = "3600",
                Param2 = "600",
                Param8 = "9",
                ToHit = "20",
                LevToHit = "9",
                ToHitCalc = "",
            },
        },
        ["data/global/excel/states.txt"] = { { state = "enchant", id = "16", group = "" } },
    },
    cases = {
        test.case("decodes_timed_multi_stat_and_hard_level_modifier", function(t)
            t:run(function()
                local definition = require("d2legacy.data.targeted_state_skills").load({ 52 })[52]
                test.expect(definition.behavior):equals("state.targeted-timed-stats")
                test.expect(definition.mana_cost_raw):equals(25 * 256)
                test.expect(definition.mana_cost_per_level_raw):equals(256)
                test.expect(definition.duration_base):equals(3600)
                test.expect(definition.duration_per_level):equals(600)
                test.expect(definition.attack_rating_percent_base):equals(20)
                test.expect(definition.attack_rating_percent_per_level):equals(9)
                test.expect(definition.minimum_damage_raw):equals(16 * 128)
                test.expect(definition.maximum_damage_per_level_raw[5]):equals(21 * 128)
                test.expect(definition.damage_synergy_skill_ids[1]):equals(37)
                test.expect(definition.damage_synergy_percent_per_level):equals(9)
            end)
        end),
    },
})
