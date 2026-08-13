local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/player" },
    cases = {
        test.case("player_entry_and_movement_use_authoritative_systems", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "player.move",
                payload = { x = 1, y = 1, running = true },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local players = ecs.query({
                    all = {
                        "d2legacy.player.identity",
                        "d2legacy.world.velocity",
                        "d2legacy.player.animation",
                    },
                })
                test.assert(#players == 1, [=[#players == 1]=])
                local velocity = ecs.get(players[1], "d2legacy.world.velocity")
                local position = ecs.get(players[1], "d2legacy.world.position")
                local animation = ecs.get(players[1], "d2legacy.player.animation")
                local facing = ecs.get(players[1], "d2legacy.world.facing")
                local mode = ecs.get(players[1], "d2legacy.player.movement_mode")
                local expected = 15 * 0.7071067811865476
                test.assert(
                    math.abs(velocity:get("x") - expected) < 0.000000001,
                    [=[math.abs(velocity:get("x") - expected) < 0.000000001]=]
                )
                test.assert(
                    math.abs(velocity:get("y") - expected) < 0.000000001,
                    [=[math.abs(velocity:get("y") - expected) < 0.000000001]=]
                )
                test.assert(animation:get("mode") == "RN", [=[animation:get("mode") == "RN"]=])
                test.assert(facing:get("direction") == 4, [=[facing:get("direction") == 4]=])
                test.assert(mode:get("running") == true, [=[mode:get("running") == true]=])
                test.assert(position:get("x") > fixtures.amazon_entry.x,
                    [=[position:get("x") > fixtures.amazon_entry.x]=])
                test.assert(position:get("y") > fixtures.amazon_entry.y,
                    [=[position:get("y") > fixtures.amazon_entry.y]=])
            end),
            test.submit({
                tick = 3,
                sequence = 2,
                player = "alice",
                kind = "player.move",
                payload = { x = 0, y = 0, running = false, target = { x = 20, y = 16 } },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.player.identity" } })[1]
                local animation = ecs.get(player, "d2legacy.player.animation")
                local facing = ecs.get(player, "d2legacy.world.facing")
                test.assert(animation:get("mode") == "WL", [=[animation:get("mode") == "WL"]=])
                test.assert(facing:get("direction") == 15, [=[facing:get("direction") == 15]=])
            end),
        }),
    },
})
