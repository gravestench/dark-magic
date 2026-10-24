local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local suite = {
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
                    source_direction = "east",
                    arrival_x = 6,
                    arrival_y = 40,
                    world_width = 400,
                    world_height = 400,
                    trigger_radius = 2,
                    trigger_lane_min = 10,
                    trigger_lane_max = 14,
                },
                {
                    source_level = 1,
                    destination_level = 2,
                    source_x = 275,
                    source_y = 84,
                    source_direction = "east",
                    arrival_x = 6,
                    arrival_y = 40,
                    world_width = 280,
                    world_height = 200,
                    trigger_radius = 2,
                    trigger_lane_min = 82,
                    trigger_lane_max = 86,
                },
                {
                    source_level = 2,
                    destination_level = 1,
                    source_x = 20,
                    source_y = 20,
                    source_direction = "west",
                    arrival_x = 30,
                    arrival_y = 30,
                    world_width = 400,
                    world_height = 400,
                    trigger_radius = 0.5,
                },
                {
                    source_level = 1,
                    destination_level = 2,
                    source_x = 5,
                    source_y = 100,
                    source_direction = "west",
                    arrival_x = 6,
                    arrival_y = 40,
                    world_width = 400,
                    world_height = 400,
                    trigger_radius = 2,
                    trigger_lane_min = 98,
                    trigger_lane_max = 102,
                },
                {
                    source_level = 1,
                    destination_level = 2,
                    source_x = 200,
                    source_y = 5,
                    source_direction = "north",
                    arrival_x = 6,
                    arrival_y = 40,
                    world_width = 400,
                    world_height = 400,
                    trigger_radius = 2,
                    trigger_lane_min = 198,
                    trigger_lane_max = 202,
                },
                {
                    source_level = 1,
                    destination_level = 2,
                    source_x = 300,
                    source_y = 395,
                    source_direction = "south",
                    arrival_x = 6,
                    arrival_y = 40,
                    world_width = 400,
                    world_height = 400,
                    trigger_radius = 2,
                    trigger_lane_min = 298,
                    trigger_lane_max = 302,
                },
            },
        },
    },
    cases = {
        test.case("town_strip_and_reverse_circle_use_their_own_geometry", {
            test.run(function()
                local world_transition = require("d2legacy.systems.world_transition")
                test.assert(world_transition.at_trigger(1, 11.5, 12), "first seam lost its radius")
                test.assert(world_transition.at_trigger(1, 10, 14), "town strip lost its outer lane")
                test.assert(not world_transition.at_trigger(1, 10, 15), "town strip entered adjacent fenced lane")
                test.assert(not world_transition.at_trigger(1, 13, 12), "town strip extended beyond normal depth")
                test.assert(not world_transition.at_trigger(2, 21, 20), "second seam used the first seam's radius")
                test.assert(world_transition.at_trigger(2, 20.4, 20), "second seam rejected its own trigger")
            end),
        }),
        test.case("outside_pointer_through_town_seam_targets_reachable_trigger", {
            test.run(function()
                local geometry = require("d2legacy.gameplay.seam_geometry")
                local x, y = geometry.pointer_target(1, 6, 12, 401, 12, 400, 400)
                test.expect(x):equals(10)
                test.expect(y):equals(12)

                x, y = geometry.pointer_target(1, 6, 30, 401, 30, 400, 400)
                test.expect(x):equals(401)
                test.expect(y):equals(30)

                -- This ray grazes the trigger near the hero but exits far from the east opening.
                x, y = geometry.pointer_target(1, 6, 12, 401, 100, 400, 400)
                test.expect(x):equals(401)
                test.expect(y):equals(100)

                -- A near-trigger ray toward the south boundary is not an east exit click.
                x, y = geometry.pointer_target(1, 8, 11, 20, 401, 400, 400)
                test.expect(x):equals(20)
                test.expect(y):equals(401)

                x, y = geometry.pointer_target(1, 6, 12, 11, 12, 400, 400)
                test.expect(x):equals(11)
                test.expect(y):equals(12)

                x, y = geometry.pointer_target(2, 6, 12, 401, 12, 400, 400)
                test.expect(x):equals(401)
                test.expect(y):equals(12)

                for _, facing in ipairs({
                    { hero = { 9, 100 }, outside = { -1, 100 }, seam = { 5, 100 } },
                    { hero = { 200, 9 }, outside = { 200, -1 }, seam = { 200, 5 } },
                    { hero = { 300, 391 }, outside = { 300, 401 }, seam = { 300, 395 } },
                }) do
                    x, y = geometry.pointer_target(
                        1,
                        facing.hero[1],
                        facing.hero[2],
                        facing.outside[1],
                        facing.outside[2],
                        400,
                        400
                    )
                    test.expect(x):equals(facing.seam[1])
                    test.expect(y):equals(facing.seam[2])
                end
            end),
        }),
        test.case("town_bridge_shoulder_pointer_targets_safe_strip", {
            test.run(function()
                local geometry = require("d2legacy.gameplay.seam_geometry")
                for _, case in ipairs({
                    { hero_y = 79, target_y = 79, want_x = 281, want_y = 79 },
                    { hero_y = 80, target_y = 80, want_x = 275, want_y = 82 },
                    { hero_y = 81, target_y = 81, want_x = 275, want_y = 82 },
                    { hero_y = 84, target_y = 84, want_x = 275, want_y = 84 },
                    { hero_y = 88, target_y = 88, want_x = 275, want_y = 86 },
                    { hero_y = 89, target_y = 89, want_x = 281, want_y = 89 },
                    -- The ray reaches the east edge beside the strip, but misses at the trigger plane.
                    { hero_y = 70, target_y = 85, want_x = 281, want_y = 85 },
                    { hero_y = 80, target_y = 100, want_x = 281, want_y = 100 },
                }) do
                    local x, y = geometry.pointer_target(1, 269, case.hero_y, 281, case.target_y, 280, 200)
                    test.expect(x):equals(case.want_x)
                    test.expect(y):equals(case.want_y)
                end

                local x, y = geometry.pointer_target(1, 269, 80, 279, 80, 280, 200)
                test.expect(x):equals(279)
                test.expect(y):equals(80)
                x, y = geometry.pointer_target(2, 269, 80, 281, 80, 280, 200)
                test.expect(x):equals(281)
                test.expect(y):equals(80)
                x, y = geometry.pointer_target(1, 269, 80, 269, -1, 280, 200)
                test.expect(x):equals(269)
                test.expect(y):equals(-1)
            end),
        }),
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
}

for lane = 9, 15 do
    local current_lane = lane
    local crosses = lane >= 10 and lane <= 14
    suite.cases[#suite.cases + 1] = test.case("town_strip_lane_" .. lane, {
        test.submit_system({
            tick = 1,
            sequence = 1,
            kind = "system.player.enter",
            payload = fixtures.player_entry({ x = 10, y = current_lane }),
        }),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
            local expected = crosses and 2 or 1
            test.expect(ecs.get(player, "d2legacy.world.location"):get("level_id")):equals(expected)
        end),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
            local expected = crosses and 2 or 1
            test.expect(ecs.get(player, "d2legacy.world.location"):get("level_id")):equals(expected)
        end),
    })
end

for _, facing in ipairs({
    { name = "west", x = 5, y = 100 },
    { name = "north", x = 200, y = 5 },
    { name = "south", x = 300, y = 395 },
}) do
    suite.cases[#suite.cases + 1] = test.case("town_strip_cardinal_" .. facing.name, {
        test.submit_system({
            tick = 1,
            sequence = 1,
            kind = "system.player.enter",
            payload = fixtures.player_entry({ x = facing.x, y = facing.y, world_width = 400, world_height = 400 }),
        }),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
            test.expect(ecs.get(player, "d2legacy.world.location"):get("level_id")):equals(2)
        end),
    })
end

return test.suite(suite)
