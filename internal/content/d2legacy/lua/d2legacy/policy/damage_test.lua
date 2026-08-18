local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local enter = fixtures.player_entry({
    fire_resistance = 20,
    cold_resistance = 20,
    health = 10,
    max_health = 10,
})

return test.suite({
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("derived_mitigation_and_lethal_damage_vectors", {
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
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = player,
                        source_id = "shield:fire",
                        stat = "fire_resist",
                        value = 30,
                    },
                })
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = player,
                        source_id = "shield:cold",
                        stat = "cold_resist",
                        value = 30,
                    },
                })
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = player,
                        source_id = "armor:physical",
                        stat = "physical_resist",
                        value = 25,
                    },
                })
                ecs.create({
                    ["d2legacy.stat.source"] = {
                        target = player,
                        source_id = "armor:flat",
                        stat = "physical_reduction_raw",
                        value = 20,
                    },
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local mitigation = require("d2legacy.policy.mitigation")
                local damage = require("d2legacy.policy.damage")
                local bundle = require("d2legacy.policy.damage_bundle")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local defense = ecs.get(player, "d2legacy.combat.defense")
                test.assert(defense:get("fire_resist") == 50, [=[defense:get("fire_resist") == 50]=])
                test.assert(defense:get("cold_resist") == 50, [=[defense:get("cold_resist") == 50]=])
                test.assert(defense:get("physical_resist") == 25, [=[defense:get("physical_resist") == 25]=])
                test.assert(
                    mitigation.apply(1000, "fire", defense) == 500,
                    [=[mitigation.apply(1000, "fire", defense) == 500]=]
                )
                test.assert(
                    mitigation.apply(1000, "cold", defense) == 500,
                    [=[mitigation.apply(1000, "cold", defense) == 500]=]
                )
                test.assert(
                    mitigation.apply(1000, "physical", defense) == 730,
                    [=[mitigation.apply(1000, "physical", defense) == 730]=]
                )
                local mixed_target = ecs.create({
                    ["d2legacy.monster.stats"] = {
                        level = 1,
                        health = 5000,
                        max_health = 5000,
                    },
                    ["d2legacy.combat.defense"] = {
                        physical_resist = 25,
                        fire_resist = 50,
                        cold_resist = 0,
                        max_fire_resist = 75,
                        max_cold_resist = 75,
                        physical_reduction_raw = 20,
                    },
                })
                local mixed = damage.resolve(mixed_target, { physical = 1000, fire = 1000 }, ecs)
                test.assert(
                    mixed.channel == "mixed"
                        and mixed.rolled.physical == 1000
                        and mixed.rolled.fire == 1000
                        and mixed.mitigated.physical == 730
                        and mixed.mitigated.fire == 500
                        and mixed.damage_raw == 1230
                        and mixed.remaining_health_raw == 3770,
                    [=[typed channels remain independent through mitigation and join only at health commit]=]
                )
                local poison_target = ecs.create({
                    ["d2legacy.monster.stats"] = {
                        level = 1,
                        health = 5000,
                        max_health = 5000,
                    },
                    ["d2legacy.combat.defense"] = {
                        fire_resist = 50,
                        cold_resist = 0,
                        max_fire_resist = 75,
                        max_cold_resist = 75,
                    },
                })
                local poison = damage.resolve(poison_target, { fire = 1000, poison = 1000 }, ecs)
                test.assert(
                    poison.mitigated.fire == 500
                        and poison.mitigated.poison == 1000
                        and poison.damage_raw == 500
                        and poison.remaining_health_raw == 4500,
                    [=[poison stays typed but cannot become immediate damage before duration policy exists]=]
                )
                local result = damage.resolve(player, bundle.single("fire", 255), ecs)
                test.assert(
                    result.rolled_damage_raw == 255
                        and result.damage_raw == 0
                        and result.remaining_health_raw == 2560
                        and not result.lethal,
                    [=[whole-health storage reports only damage actually committed]=]
                )
                result = damage.resolve(player, bundle.single("fire", 4096), ecs)
                test.assert(
                    result.channel == "fire"
                        and result.rolled_damage_raw == 4096
                        and result.damage_raw == 2048
                        and result.remaining_health_raw == 512
                        and not result.lethal,
                    [=[shared result records rolled, mitigated, and remaining damage in order]=]
                )
                result = damage.resolve(player, bundle.single("fire", 1024), ecs)
                test.assert(
                    result.rolled_damage_raw == 1024
                        and result.damage_raw == 512
                        and result.remaining_health_raw == 0
                        and result.lethal,
                    [=[shared result marks lethal damage after application]=]
                )
            end),
        }),
    },
})
