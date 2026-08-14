local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/player" },
    cases = {
        test.case("departure_removes_player_and_admission_owned_state", {
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.player.enter",
                payload = fixtures.amazon_entry,
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    #ecs.query({ all = { "d2legacy.player.identity" } }) == 1,
                    [=[#ecs.query({ all = { "d2legacy.player.identity" } }) == 1]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.items.layout" } }) == 1,
                    [=[#ecs.query({ all = { "d2legacy.items.layout" } }) == 1]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.interaction.context" } }) == 1,
                    [=[#ecs.query({ all = { "d2legacy.interaction.context" } }) == 1]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.interaction.null_target" } }) == 1,
                    [=[#ecs.query({ all = { "d2legacy.interaction.null_target" } }) == 1]=]
                )
            end),
            test.submit_system({
                tick = 2,
                sequence = 2,
                kind = "system.player.leave",
                payload = { player = "alice" },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    #ecs.query({ all = { "d2legacy.player.identity" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.player.identity" } }) == 0]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.items.layout" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.items.layout" } }) == 0]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.interaction.context" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.interaction.context" } }) == 0]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.interaction.null_target" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.interaction.null_target" } }) == 0]=]
                )
            end),
        }),
    },
})
