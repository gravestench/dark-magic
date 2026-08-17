local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local enter = fixtures.player_entry({
    fire_resistance = 20,
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
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local defense = ecs.get(player, "d2legacy.combat.defense")
                test.assert(defense:get("fire_resist") == 50, [=[defense:get("fire_resist") == 50]=])
                test.assert(defense:get("physical_resist") == 25, [=[defense:get("physical_resist") == 25]=])
                test.assert(
                    mitigation.apply(1000, "fire", defense) == 500,
                    [=[mitigation.apply(1000, "fire", defense) == 500]=]
                )
                test.assert(
                    mitigation.apply(1000, "physical", defense) == 730,
                    [=[mitigation.apply(1000, "physical", defense) == 730]=]
                )
                local result = damage.resolve(player, 255, ecs, "fire")
                test.assert(
                    result.rolled_damage_raw == 255
                        and result.damage_raw == 0
                        and result.remaining_health_raw == 2560
                        and not result.lethal,
                    [=[whole-health storage reports only damage actually committed]=]
                )
                result = damage.resolve(player, 4096, ecs, "fire")
                test.assert(
                    result.channel == "fire"
                        and result.rolled_damage_raw == 4096
                        and result.damage_raw == 2048
                        and result.remaining_health_raw == 512
                        and not result.lethal,
                    [=[shared result records rolled, mitigated, and remaining damage in order]=]
                )
                result = damage.resolve(player, 1024, ecs, "fire")
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
