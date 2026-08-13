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
            end),
        }),
    },
})
