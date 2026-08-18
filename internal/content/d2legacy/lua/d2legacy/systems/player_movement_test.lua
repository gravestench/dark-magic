local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/player" },
    cases = {
        test.case("players_admitted_at_one_realm_anchor_receive_independent_movement_space", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ x = 10, y = 12 }),
            }),
            test.submit_system({
                tick = 2,
                sequence = 2,
                kind = "system.player.enter",
                payload = fixtures.player_entry({
                    character_id = "second-hero",
                    player = "bob",
                    name = "Second Hero",
                    x = 10,
                    y = 12,
                }),
            }),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local positions = {}
                local players = ecs.query({
                    all = { "d2legacy.player.identity", "d2legacy.world.position" },
                })
                for _, player in ipairs(players) do
                    local identity = ecs.get(player, "d2legacy.player.identity")
                    local position = ecs.get(player, "d2legacy.world.position")
                    positions[identity:get("player")] = { x = position:get("x"), y = position:get("y") }
                end
                test.assert(positions.alice ~= nil and positions.bob ~= nil, "both realm players were not admitted")
                test.assert(
                    math.sqrt((positions.alice.x - positions.bob.x) ^ 2 + (positions.alice.y - positions.bob.y) ^ 2)
                        >= 2,
                    "realm players retained overlapping collision footprints"
                )
            end),
            test.submit({
                tick = 3,
                sequence = 1,
                player = "alice",
                kind = "player.move",
                payload = { x = 1, y = 0, running = false },
            }),
            test.submit({
                tick = 3,
                sequence = 1,
                player = "bob",
                kind = "player.move",
                payload = { x = 1, y = 0, running = false },
            }),
            test.step(12),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local players = ecs.query({
                    all = { "d2legacy.player.identity", "d2legacy.world.position" },
                })
                for _, player in ipairs(players) do
                    local identity = ecs.get(player, "d2legacy.player.identity")
                    local position = ecs.get(player, "d2legacy.world.position")
                    test.assert(
                        position:get("x") > 12,
                        identity:get("player") .. " did not sustain independent authoritative movement"
                    )
                end
            end),
            test.expect_checkpoint_parity(1),
        }),
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
                local motion = ecs.get(players[1], "d2legacy.player.motion")
                local expected = 9 * 0.7071067811865476
                test.assert(
                    math.abs(velocity:get("x") - expected) < 0.000000001,
                    [=[math.abs(velocity:get("x") - expected) < 0.000000001]=]
                )
                test.assert(
                    math.abs(velocity:get("y") - expected) < 0.000000001,
                    [=[math.abs(velocity:get("y") - expected) < 0.000000001]=]
                )
                test.assert(animation:get("mode") == "RN", [=[animation:get("mode") == "RN"]=])
                test.assert(animation:get("start_tick") == 2, [=[animation:get("start_tick") == 2]=])
                test.assert(facing:get("direction") == 4, [=[facing:get("direction") == 4]=])
                test.assert(mode:get("running") == true, [=[mode:get("running") == true]=])
                test.assert(
                    motion:get("owner") == "locomotion" and motion:get("kind") == "direction" and motion:get("active"),
                    "directional command did not own explicit player motion"
                )
                test.assert(
                    position:get("x") > fixtures.amazon_entry.x,
                    [=[position:get("x") > fixtures.amazon_entry.x]=]
                )
                test.assert(
                    position:get("y") > fixtures.amazon_entry.y,
                    [=[position:get("y") > fixtures.amazon_entry.y]=]
                )
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
                local motion = ecs.get(player, "d2legacy.player.motion")
                test.assert(animation:get("mode") == "WL", [=[animation:get("mode") == "WL"]=])
                test.assert(animation:get("start_tick") == 3, [=[animation:get("start_tick") == 3]=])
                test.assert(facing:get("direction") == 15, [=[facing:get("direction") == 15]=])
                test.assert(
                    motion:get("owner") == "locomotion"
                        and motion:get("kind") == "waypoint"
                        and motion:get("has_target"),
                    "waypoint command did not become explicit player motion"
                )
            end),
        }),
    },
})
