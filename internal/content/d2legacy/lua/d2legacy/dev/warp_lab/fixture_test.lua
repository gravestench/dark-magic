local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "ecs",
    tier = "fast",
    covers = { "internal/game/transition" },
    cases = {
        test.case("intent_walks_to_portal_and_arrives_at_pair", {
            test.run(function()
                local fixture = require("d2legacy.dev.warp_lab.fixture")
                warp_fixture_module = fixture
                warp_fixture = fixture.create({ x = 10.5, y = 0.5 }, { x = 100, y = 50 }, { x = 0.5, y = 0.5 })
                fixture.intent(warp_fixture, "warp-lab:a", "0,0;10,0")
            end),
            test.update(1000),
            test.update(1000),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local position = ecs.get(warp_fixture.player, "d2legacy.world.position")
                local status = ecs.get(warp_fixture.player, "d2legacy.lab.warp.state")
                local actor = ecs.get(warp_fixture.player, "d2legacy.lab.warp.actor")
                test.assert(
                    position:get("x") == 102 and position:get("y") == 52,
                    [=[position:get("x") == 102 and position:get("y") == 52]=]
                )
                test.assert(
                    status:get("warp_count") == 1 and actor:get("direction") == 3,
                    [=[status:get("warp_count") == 1 and actor:get("direction") == 3]=]
                )
                test.assert(
                    ecs.get(warp_fixture.player, "d2legacy.lab.warp.intent") == nil,
                    [=[ecs.get(warp_fixture.player, "d2legacy.lab.warp.intent") == nil]=]
                )
                warp_fixture_module.move(warp_fixture, 110, 52, "102,55;110,55;110,52")
            end),
            test.update(1000),
            test.update(1000),
            test.update(1000),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local position = ecs.get(warp_fixture.player, "d2legacy.world.position")
                local actor = ecs.get(warp_fixture.player, "d2legacy.lab.warp.actor")
                test.assert(
                    position:get("x") == 110 and position:get("y") == 52,
                    [=[position:get("x") == 110 and position:get("y") == 52]=]
                )
                test.assert(actor:get("direction") == 2, [=[actor:get("direction") == 2]=])
                test.assert(
                    ecs.get(warp_fixture.player, "d2legacy.lab.warp.move_intent") == nil,
                    [=[ecs.get(warp_fixture.player, "d2legacy.lab.warp.move_intent") == nil]=]
                )
            end),
        }),
    },
})
