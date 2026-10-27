local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local endpoints = {
    {
        id = "warp:test-town",
        pair_id = "warp:test-field",
        token = "TP",
        label = "BLUE TEST WARP",
        level_id = 1,
        resident_id = "warp:test-town",
        room_id = "town-room",
        x = 12,
        y = 12,
        radius = 4,
        destination_level = 2,
        destination_x = 20,
        destination_y = 20,
        destination_width = 200,
        destination_height = 160,
    },
    {
        id = "warp:test-field",
        pair_id = "warp:test-town",
        token = "PP",
        label = "RED TEST WARP",
        level_id = 2,
        resident_id = "warp:test-field",
        room_id = "field-room",
        x = 22,
        y = 20,
        radius = 4,
        destination_level = 1,
        destination_x = 10,
        destination_y = 12,
        destination_width = 100,
        destination_height = 80,
    },
}

local function active_motion()
    return test.run(function()
        local ecs = require("engine.ecs/v1")
        local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
        ecs.set(player, "d2legacy.world.forced_motion", {
            kind = "knockback",
            target_x = 16,
            target_y = 12,
            speed = 25,
            start_x = 8,
            start_y = 12,
            requested_distance = 8,
            applied_distance = 2,
            start_tick = 1,
        })
    end)
end

local function assert_interrupted_at(tick, x, applied_distance)
    return test.run(function()
        local ecs = require("engine.ecs/v1")
        local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
        test.expect(ecs.get(player, "d2legacy.world.location"):get("level_id")):equals(2)
        test.expect(ecs.get(player, "d2legacy.world.position"):get("x")):equals(20)
        test.expect(ecs.get(player, "d2legacy.world.position"):get("y")):equals(20)
        test.assert(not ecs.get(player, "d2legacy.world.forced_motion"), "warp retained forced motion")
        local events = ecs.query({ all = { "d2legacy.world.forced_motion_event" } })
        test.expect(#events):equals(1)
        local event = ecs.get(events[1], "d2legacy.world.forced_motion_event")
        test.expect(event:get("kind")):equals("knockback")
        test.expect(event:get("outcome")):equals("interrupted")
        test.expect(event:get("tick")):equals(tick)
        test.expect(event:get("target_id")):equals("player:alice")
        test.expect(event:get("start_x")):equals(8)
        test.expect(event:get("start_y")):equals(12)
        test.expect(event:get("end_x")):equals(x)
        test.expect(event:get("end_y")):equals(12)
        test.expect(event:get("requested_distance")):equals(8)
        test.expect(event:get("applied_distance")):equals(applied_distance)
    end)
end

local function assert_no_second_outcome()
    return test.run(function()
        local ecs = require("engine.ecs/v1")
        local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
        test.expect(ecs.get(player, "d2legacy.world.position"):get("x")):equals(20)
        test.assert(not ecs.get(player, "d2legacy.world.forced_motion"), "motion restarted after warp")
        test.expect(#ecs.query({ all = { "d2legacy.world.forced_motion_event" } })):equals(1)
    end)
end

local function warp_after_checkpoint(restore)
    local actions = {
        test.submit_system({
            tick = 1,
            sequence = 1,
            kind = "system.player.enter",
            payload = fixtures.amazon_entry,
        }),
        test.step(1),
        active_motion(),
        test.step(1),
        test.run(function()
            local ecs = require("engine.ecs/v1")
            local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
            test.expect(ecs.get(player, "d2legacy.world.position"):get("x")):equals(11)
            test.expect(ecs.get(player, "d2legacy.world.forced_motion"):get("applied_distance")):equals(3)
        end),
    }
    if restore then
        actions[#actions + 1] = test.restore_checkpoint()
    end
    actions[#actions + 1] = test.submit({
        tick = 3,
        sequence = 1,
        player = "alice",
        kind = "interaction.open",
        payload = { target = "warp:test-town" },
    })
    actions[#actions + 1] = test.step(1)
    actions[#actions + 1] = assert_interrupted_at(3, 11, 3)
    actions[#actions + 1] = test.step(1)
    actions[#actions + 1] = assert_no_second_outcome()
    return actions
end

return test.suite({
    profile = "authority",
    tier = "integration",
    covers = { "internal/game/transition" },
    initial_data = {
        ["d2legacy.world_warps"] = { endpoints = endpoints },
    },
    cases = {
        test.case("paired_warps_use_interaction_admission_and_shared_relocation", {
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
                kind = "interaction.open",
                payload = { target = "warp:test-town" },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                local location = ecs.get(player, "d2legacy.world.location")
                local position = ecs.get(player, "d2legacy.world.position")
                local bounds = ecs.get(player, "d2legacy.world.bounds")
                local velocity = ecs.get(player, "d2legacy.world.velocity")
                test.expect(#ecs.query({ all = { "d2legacy.world.warp" } })):equals(2)
                local residents = ecs.query({ all = { "d2legacy.world.warp", "d2legacy.world.room_resident" } })
                test.expect(#residents):equals(2)
                test.expect(location:get("level_id")):equals(2)
                test.expect(location:get("transition_kind")):equals("warp")
                test.expect(location:get("transition_source")):equals(1)
                test.expect(location:get("transition_sequence")):equals(1)
                test.expect(position:get("x")):equals(20)
                test.expect(position:get("y")):equals(20)
                test.expect(bounds:get("width")):equals(200)
                test.expect(bounds:get("height")):equals(160)
                test.expect(velocity:get("x")):equals(0)
                test.expect(velocity:get("y")):equals(0)
                local motion = ecs.get(player, "d2legacy.player.motion")
                test.expect(motion:get("owner")):equals("none")
                test.assert(not motion:get("active"), "warp retained motion ownership")
            end),
            test.restore_checkpoint(),
            test.submit({
                tick = 3,
                sequence = 1,
                player = "alice",
                kind = "interaction.open",
                payload = { target = "warp:test-field" },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                test.expect(ecs.get(player, "d2legacy.world.location"):get("level_id")):equals(1)
                local relocation = ecs.get(player, "d2legacy.world.location")
                test.expect(relocation:get("transition_sequence")):equals(2)
                test.expect(relocation:get("transition_source")):equals(2)
                test.expect(relocation:get("transition_kind")):equals("warp")
                test.expect(ecs.get(player, "d2legacy.world.position"):get("x")):equals(10)
                test.expect(ecs.get(player, "d2legacy.world.position"):get("y")):equals(12)
            end),
        }),
        test.case("warp_interrupts_active_forced_motion_before_relocation", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            active_motion(),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "interaction.open",
                payload = { target = "warp:test-town" },
            }),
            test.step(1),
            assert_interrupted_at(2, 10, 2),
            test.step(1),
            assert_no_second_outcome(),
        }),
        test.case("warp_interruption_after_active_checkpoint", warp_after_checkpoint(false)),
        test.case("restored_warp_interruption_matches_active_checkpoint", warp_after_checkpoint(true)),
        test.case("stale_out_of_range_warp_request_is_a_harmless_no_op", {
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
                local position = ecs.get(player, "d2legacy.world.position")
                position:set("x", 80)
                position:set("y", 80)
            end),
            active_motion(),
            test.submit({
                tick = 2,
                sequence = 1,
                player = "alice",
                kind = "interaction.open",
                payload = { at = true, x = 12, y = 12 },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({ all = { "d2legacy.world.player_control" } })[1]
                test.expect(ecs.get(player, "d2legacy.world.location"):get("level_id")):equals(1)
                local position = ecs.get(player, "d2legacy.world.position")
                test.assert(position:get("x") > 70 and position:get("y") > 70, "stale request relocated player")
                test.assert(ecs.get(player, "d2legacy.world.forced_motion"), "stale warp removed forced motion")
                test.expect(#ecs.query({ all = { "d2legacy.world.forced_motion_event" } })):equals(0)
            end),
        }),
    },
})
