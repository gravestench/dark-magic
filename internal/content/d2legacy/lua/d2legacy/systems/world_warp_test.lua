local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local endpoints = {
    {
        id = "warp:test-town",
        pair_id = "warp:test-field",
        token = "TP",
        label = "BLUE TEST WARP",
        level_id = 1,
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
                test.expect(location:get("level_id")):equals(2)
                test.expect(position:get("x")):equals(20)
                test.expect(position:get("y")):equals(20)
                test.expect(bounds:get("width")):equals(200)
                test.expect(bounds:get("height")):equals(160)
                test.expect(velocity:get("x")):equals(0)
                test.expect(velocity:get("y")):equals(0)
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
                test.expect(ecs.get(player, "d2legacy.world.position"):get("x")):equals(10)
                test.expect(ecs.get(player, "d2legacy.world.position"):get("y")):equals(12)
            end),
        }),
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
            end),
        }),
    },
})
