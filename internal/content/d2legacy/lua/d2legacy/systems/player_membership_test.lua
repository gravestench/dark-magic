local test = require("d2legacy.tests/v1")
local source_id
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
                local sources = ecs.query({ all = { "d2legacy.items.world_source" } })
                test.expect(#sources):equals(1)
                source_id = sources[1]:id()
                test.assert(
                    #ecs.query({ all = { "d2legacy.player.identity" } }) == 1,
                    [=[#ecs.query({ all = { "d2legacy.player.identity" } }) == 1]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.items.layout" }, none = { "d2legacy.items.world_source" } }) == 1,
                    "one admitted player layout"
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
                local sources = ecs.query({ all = { "d2legacy.items.world_source" } })
                test.expect(#sources):equals(1)
                test.expect(sources[1]:id()):equals(source_id)
                test.assert(
                    #ecs.query({ all = { "d2legacy.player.identity" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.player.identity" } }) == 0]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.items.layout" }, none = { "d2legacy.items.world_source" } }) == 0,
                    "departure removes every player layout"
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
