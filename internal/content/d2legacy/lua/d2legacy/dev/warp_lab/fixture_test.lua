local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "policy",
    tier = "fast",
    tests = {
        intent_walks_to_portal_and_arrives_at_pair = {
            {
                run = function()
                    local fixture = require("d2legacy.dev.warp_lab.fixture")
                    warp_fixture_module = fixture
                    warp_fixture = fixture.create({ x = 10.5, y = 0.5 }, { x = 100, y = 50 }, { x = 0.5, y = 0.5 })
                    fixture.intent(warp_fixture, "warp-lab:a", "0,0;10,0")
                end,
            },
            { engine_update_ms = 1000 },
            { engine_update_ms = 1000 },
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    local position = ecs.get(warp_fixture.player, "d2legacy.world.position")
                    local status = ecs.get(warp_fixture.player, "d2legacy.lab.warp.state")
                    local actor = ecs.get(warp_fixture.player, "d2legacy.lab.warp.actor")
                    assert(position:get("x") == 102 and position:get("y") == 52)
                    assert(status:get("warp_count") == 1 and actor:get("direction") == 3)
                    assert(ecs.get(warp_fixture.player, "d2legacy.lab.warp.intent") == nil)
                    warp_fixture_module.move(warp_fixture, 110, 52, "102,55;110,55;110,52")
                end,
            },
            { engine_update_ms = 1000 },
            { engine_update_ms = 1000 },
            { engine_update_ms = 1000 },
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    local position = ecs.get(warp_fixture.player, "d2legacy.world.position")
                    local actor = ecs.get(warp_fixture.player, "d2legacy.lab.warp.actor")
                    assert(position:get("x") == 110 and position:get("y") == 52)
                    assert(actor:get("direction") == 2)
                    assert(ecs.get(warp_fixture.player, "d2legacy.lab.warp.move_intent") == nil)
                end,
            },
        },
    },
})
