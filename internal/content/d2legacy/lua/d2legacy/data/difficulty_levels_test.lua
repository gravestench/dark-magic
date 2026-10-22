local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local enter = fixtures.player_entry({
    difficulty = 2,
    fire_resistance = 20,
    cold_resistance = 10,
    lightning_resistance = 0,
    poison_resistance = 30,
})

return test.suite({
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.game_rules"] = {
            target = "lod-1.14d",
            expansion = true,
            difficulty = 2,
        },
    },
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
                MonsterFreezeDivisor = "4",
                MonsterColdDivisor = "4",
                DeathExpPenalty = "10",
                LifeStealDivisor = "3",
                ManaStealDivisor = "3",
            },
        },
    },
    cases = {
        test.case("summon_resist_keeps_monstats_base_and_skips_fire_when_absorbing", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.create({
                    ["d2legacy.monster.stats"] = { level = 1, health = 100, max_health = 100 },
                    ["d2legacy.combat.defense"] = {
                        base_physical_resist = 12,
                        base_magic_resist = 18,
                        magic_resist = 18,
                        base_fire_resist = 20,
                        base_cold_resist = 30,
                        base_lightning_resist = 40,
                        base_poison_resist = 50,
                        summon_resist_bonus = 34,
                        fire_absorb_percent = 50,
                    },
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local mitigation = require("d2legacy.policy.mitigation")
                local monster = ecs.query({ all = { "d2legacy.monster.stats" } })[1]
                local defense = ecs.get(monster, "d2legacy.combat.defense")
                test.expect(defense:get("physical_resist")):equals(12)
                test.expect(defense:get("magic_resist")):equals(18)
                test.expect(defense:get("base_fire_resist")):equals(20)
                test.expect(defense:get("fire_resist")):equals(20)
                test.expect(defense:get("cold_resist")):equals(64)
                test.expect(defense:get("lightning_resist")):equals(74)
                test.expect(defense:get("poison_resist")):equals(84)
                test.expect(mitigation.apply(100, "fire", defense, true)):equals(80)
                test.expect(mitigation.apply(100, "poison", defense, true)):equals(16)
            end),
        }),
        test.case("selected_row_follows_immutable_game_difficulty", {
            test.run(function()
                local levels = require("d2legacy.data.difficulty_levels")
                test.expect(levels.resist_penalty()):equals(-100)
            end),
        }),
        test.case("players_take_the_session_resist_penalty_before_the_cap", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = enter,
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local defense = ecs.get(player, "d2legacy.combat.defense")
                test.expect(defense:get("base_fire_resist")):equals(20)
                test.expect(defense:get("fire_resist")):equals(-80)
                test.expect(defense:get("cold_resist")):equals(-90)
                test.expect(defense:get("lightning_resist")):equals(-100)
                test.expect(defense:get("base_poison_resist")):equals(30)
                test.expect(defense:get("poison_resist")):equals(-70)
                test.expect(defense:get("max_poison_resist")):equals(75)
                test.expect(defense:get("physical_resist")):equals(0)
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = player,
                        source_id = "shield:fire",
                        stat = "fire_resist",
                        value = 170,
                    },
                })
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = player,
                        source_id = "charm:poison",
                        stat = "poison_resist",
                        value = 180,
                    },
                })
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = player,
                        source_id = "charm:max-poison",
                        stat = "max_poison_resist",
                        value = 25,
                    },
                })
                ecs.create({
                    ["d2legacy.monster.stats"] = {
                        level = 1,
                        health = 100,
                        max_health = 100,
                    },
                    ["d2legacy.combat.defense"] = {
                        base_fire_resist = 20,
                        fire_resist = 20,
                        max_fire_resist = 75,
                        max_cold_resist = 75,
                        max_lightning_resist = 75,
                    },
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local damage = require("d2legacy.policy.damage")
                local mitigation = require("d2legacy.policy.mitigation")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local defense = ecs.get(player, "d2legacy.combat.defense")
                -- 20 base - 100 Hell + 170 item = 90, then capped at 75.
                test.expect(defense:get("fire_resist")):equals(90)
                test.expect(mitigation.apply(1000, "fire", defense)):equals(250)
                test.expect(mitigation.apply(1000, "cold", defense)):equals(1900)
                test.expect(defense:get("poison_resist")):equals(110)
                test.expect(defense:get("max_poison_resist")):equals(100)
                test.expect(mitigation.apply(1000, "poison", defense)):equals(50)
                local vitals = ecs.get(player, "d2legacy.player.vitals")
                local health = vitals:get("health")
                local poison = damage.resolve(player, { poison = 1000 }, ecs)
                test.expect(poison.mitigated.poison):equals(50)
                test.expect(poison.damage_raw):equals(0)
                test.expect(vitals:get("health")):equals(health)
                local monster = ecs.query({ all = { "d2legacy.monster.stats" } })[1]
                test.expect(ecs.get(monster, "d2legacy.combat.defense"):get("fire_resist")):equals(20)
            end),
        }),
    },
})
