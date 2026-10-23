local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

return test.suite({
    name = "Golem spawn writer",
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/skill" },
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "75",
                skill = "Clay Golem",
                skilldesc = "clay golem",
                charclass = "nec",
                srvdofunc = "56",
                pettype = "golem",
                petmax = "1",
                summon = "fixturegolem",
                summode = "S1",
                range = "none",
                leftskill = "",
                InGame = "1",
                InTown = "1",
                mana = "15",
                lvlmana = "3",
                minmana = "1",
                manashift = "8",
                passivestat1 = "velocitypercent",
                passivecalc1 = "skill('Golem Mastery'.dm34)",
                calc1 = "skill('Golem Mastery'.ln12)",
                aurastat1 = "item_slow",
                aurastatcalc1 = "dm34",
                auraevent1 = "damagedinmelee",
                auraeventfunc1 = "27",
                Param1 = "35",
                Param2 = "35",
                Param3 = "0",
                Param4 = "75",
                Param8 = "20",
            },
            {
                Id = "79",
                skill = "Golem Mastery",
                Param1 = "20",
                Param2 = "20",
                Param3 = "0",
                Param4 = "40",
                Param5 = "25",
                Param6 = "25",
            },
            { Id = "985", skill = "BloodGolem", Param8 = "5" },
            {
                Id = "89",
                skill = "Summon Resist",
                passivestat1 = "passive_summon_resist",
                passivecalc1 = "dm12",
                Param1 = "20",
                Param2 = "75",
            },
            { Id = "990", skill = "IronGolem", Param8 = "35" },
            { Id = "994", skill = "FireGolem", Param8 = "6" },
        },
        ["data/global/excel/PetType.txt"] = {
            { ["pet type"] = "golem", group = "0", basemax = "0", unsummon = "1", warp = "1" },
        },
        ["data/global/excel/SkillDesc.txt"] = {
            {
                skilldesc = "clay golem",
                dsc3texta1 = "Sksyn",
                dsc3texta2 = "mastery",
                dsc3texta3 = "resist",
                dsc3texta4 = "blood",
                dsc3texta5 = "iron",
                dsc3texta6 = "fire",
            },
        },
        ["data/global/excel/monstats.txt"] = {
            {
                Id = "fixturegolem",
                BaseId = "fixturegolem",
                MonStatsEx = "fixturegolem",
                NameStr = "FixtureGolem",
                AI = "NecroPet",
                Align = "1",
                Code = "FG",
                enabled = "1",
                noRatio = "1",
                boss = "1",
                primeevil = "1",
                minHP = "25",
                maxHP = "25",
                AC = "7",
                A1TH = "0",
                A1MinD = "0",
                A1MaxD = "0",
                Velocity = "12",
                aidel = "15",
                aidist = "35",
            },
        },
        ["data/global/excel/monstats2.txt"] = {
            { Id = "fixturegolem", BaseW = "hth", SizeX = "2", SizeY = "2", MeleeRng = "1" },
        },
        ["data/global/excel/monlvl.txt"] = {
            { Level = "1", ["L-AC"] = "10", ["L-TH"] = "20" },
        },
    },
    cases = {
        test.case("admitted_golem_retains_monster_boss_markers", {
            test.submit_system(fixtures.command(
                "system.player.enter",
                fixtures.player_entry({
                    level = 1,
                    mana = 100,
                    max_mana = 100,
                }),
                { tick = 1, sequence = 1 }
            )),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                ecs.set(player, "d2legacy.skill.cast", {
                    skill_id = 75,
                    skill_level = 1,
                    target_x = 3,
                    target_y = 0,
                    target_id = "",
                    effect_tick = 2,
                    complete_tick = 3,
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local golems = ecs.query({ all = { "d2legacy.owned_unit", "d2legacy.monster.identity" } })
                test.expect(#golems):equals(1)
                test.expect(ecs.get(golems[1], "d2legacy.monster.boss") ~= nil):equals(true)
                test.expect(ecs.get(golems[1], "d2legacy.monster.freeze_immune") ~= nil):equals(true)
                test.expect(ecs.get(golems[1], "d2legacy.monster.prime_evil") ~= nil):equals(true)
            end),
            test.expect_checkpoint_parity(1),
        }),
    },
})
