local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
return test.suite({
    profile = "authority",
    tier = "integration",
    initial_data = {
        ["d2legacy.world_transitions"] = {
            seams = {
                {
                    source_level = 1,
                    destination_level = 2,
                    source_x = 10,
                    source_y = 12,
                    arrival_x = 6,
                    arrival_y = 40,
                    world_width = 400,
                    world_height = 400,
                },
            },
        },
    },
    cases = {
        test.case("entry_crosses_authored_seam", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                test.assert(
                    ecs.get(player, "d2legacy.world.location"):get("level_id") == 2,
                    [=[ecs.get(player, "d2legacy.world.location"):get("level_id") == 2]=]
                )
                local position = ecs.get(player, "d2legacy.world.position")
                test.assert(
                    position:get("x") == 6 and position:get("y") == 40,
                    [=[position:get("x") == 6 and position:get("y") == 40]=]
                )
                local bounds = ecs.get(player, "d2legacy.world.bounds")
                test.assert(
                    bounds:get("width") == 400 and bounds:get("height") == 400,
                    [=[bounds:get("width") == 400 and bounds:get("height") == 400]=]
                )
                local motion = ecs.get(player, "d2legacy.player.motion")
                test.expect(motion:get("owner")):equals("none")
                test.assert(not motion:get("active"), "level transition retained motion ownership")
            end),
        }),
        test.case("seam_interrupts_checkpointed_forced_motion_before_relocation", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ x = 6 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                ecs.set(player, "d2legacy.world.forced_motion", {
                    kind = "knockback",
                    target_x = 14,
                    target_y = 12,
                    speed = 25,
                    start_x = 5,
                    start_y = 12,
                    requested_distance = 9,
                    applied_distance = 1,
                    start_tick = 1,
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                test.expect(ecs.get(player, "d2legacy.world.position"):get("x")):equals(7)
                test.expect(ecs.get(player, "d2legacy.world.forced_motion"):get("applied_distance")):equals(2)
            end),
            test.expect_checkpoint_parity(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                local position = ecs.get(player, "d2legacy.world.position")
                test.expect(ecs.get(player, "d2legacy.world.location"):get("level_id")):equals(2)
                test.expect(position:get("x")):equals(6)
                test.expect(position:get("y")):equals(40)
                test.assert(not ecs.get(player, "d2legacy.world.forced_motion"), "seam retained forced motion")
                local events = ecs.query({ all = { "d2legacy.world.forced_motion_event" } })
                test.expect(#events):equals(1)
                local event = ecs.get(events[1], "d2legacy.world.forced_motion_event")
                test.expect(event:get("kind")):equals("knockback")
                test.expect(event:get("outcome")):equals("interrupted")
                test.expect(event:get("tick")):equals(3)
                test.expect(event:get("target_id")):equals("player:alice")
                test.expect(event:get("start_x")):equals(5)
                test.expect(event:get("start_y")):equals(12)
                test.expect(event:get("end_x")):equals(8)
                test.expect(event:get("end_y")):equals(12)
                test.expect(event:get("requested_distance")):equals(9)
                test.expect(event:get("applied_distance")):equals(3)
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                test.expect(ecs.get(player, "d2legacy.world.position"):get("x")):equals(6)
                test.assert(not ecs.get(player, "d2legacy.world.forced_motion"), "motion restarted after seam")
                test.expect(#ecs.query({ all = { "d2legacy.world.forced_motion_event" } })):equals(1)
            end),
        }),
        test.case("invalid_destination_does_not_interrupt_forced_motion", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.player_entry({ x = 6 }),
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local transition = require("d2legacy.gameplay.transition")
                local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                ecs.set(player, "d2legacy.world.forced_motion", {
                    kind = "knockback",
                    target_x = 14,
                    target_y = 12,
                    speed = 25,
                    start_x = 5,
                    start_y = 12,
                    requested_distance = 9,
                    applied_distance = 1,
                    start_tick = 1,
                })
                local writer = {
                    create = function(_, components)
                        return ecs.create(components)
                    end,
                    remove = function(_, entity, component)
                        return ecs.remove(entity, component)
                    end,
                }
                local ok = pcall(transition.apply, { tick = 1 }, player, {
                    level_id = 2,
                    x = -1,
                    y = 40,
                    width = 400,
                    height = 400,
                }, writer)
                test.assert(not ok, "invalid transition unexpectedly succeeded")
                test.assert(ecs.get(player, "d2legacy.world.forced_motion"), "invalid transition removed motion")
                test.expect(#ecs.query({ all = { "d2legacy.world.forced_motion_event" } })):equals(0)
                test.expect(ecs.get(player, "d2legacy.world.location"):get("level_id")):equals(1)
            end),
        }),
    },
})
