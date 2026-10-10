local test = require("d2legacy.tests/v1")

return test.suite({
    name = "cold duration policy",
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.game_rules"] = {
            target = "lod-1.14d",
            expansion = true,
            difficulty = 2,
        },
    },
    -- Synthetic divisors: the owned 1.14d Hell row authors 4 for both columns,
    -- so distinct values here prove each effect reads its own column.
    records = {
        ["data/global/excel/difficultylevels.txt"] = {
            {
                Name = "Normal",
                ResistPenalty = "0",
                MonsterFreezeDivisor = "1",
                MonsterColdDivisor = "1",
                DeathExpPenalty = "0",
                LifeStealDivisor = "1",
                ManaStealDivisor = "1",
            },
            {
                Name = "Nightmare",
                ResistPenalty = "-40",
                MonsterFreezeDivisor = "2",
                MonsterColdDivisor = "2",
                DeathExpPenalty = "5",
                LifeStealDivisor = "2",
                ManaStealDivisor = "2",
            },
            {
                Name = "Hell",
                ResistPenalty = "-100",
                MonsterFreezeDivisor = "5",
                MonsterColdDivisor = "3",
                DeathExpPenalty = "10",
                LifeStealDivisor = "3",
                ManaStealDivisor = "3",
            },
        },
    },
    cases = {
        test.case("freeze_and_chill_read_their_own_session_divisors", {
            test.run(function()
                local policy = require("d2legacy.policy.cold_duration")
                test.expect(policy.monster_frames(30, "freeze")):equals(6)
                test.expect(policy.monster_frames(30, "cold")):equals(10)
                test.expect(policy.monster_frames(2, "freeze")):equals(1)
            end),
        }),
        test.case("only_monsters_take_the_divisor_after_cold_resistance", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local policy = require("d2legacy.policy.cold_duration")
                local defense = {
                    fire_resist = 0,
                    cold_resist = 50,
                    lightning_resist = 0,
                    max_fire_resist = 75,
                    max_cold_resist = 75,
                    max_lightning_resist = 75,
                }
                local monster = ecs.create({
                    ["d2legacy.monster.stats"] = { level = 1, health = 10, max_health = 10 },
                    ["d2legacy.combat.defense"] = defense,
                })
                local player = ecs.create({ ["d2legacy.combat.defense"] = defense })
                test.expect(policy.target_frames(60, monster, "freeze", ecs)):equals(6)
                test.expect(policy.target_frames(60, monster, "cold", ecs)):equals(10)
                test.expect(policy.target_frames(60, player, "freeze", ecs)):equals(30)
            end),
        }),
    },
})
