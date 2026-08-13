local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/state" },
    cases = {
        test.case("applied_state_expires_at_the_authored_tick", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local target = ecs.create()
                ecs.create({
                    ["d2legacy.state.request"] = {
                        operation = "apply",
                        target = target,
                        state_id = "poison",
                        source_id = "monster:fallen",
                        duration = 2,
                        policy = "refresh_same_source",
                    },
                })
            end),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    #ecs.query({ all = { "d2legacy.state.instance" } }) == 1,
                    [=[#ecs.query({ all = { "d2legacy.state.instance" } }) == 1]=]
                )
            end),
            test.restore_checkpoint(),
            test.step(2),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    #ecs.query({ all = { "d2legacy.state.instance" } }) == 0,
                    [=[#ecs.query({ all = { "d2legacy.state.instance" } }) == 0]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.state.event" } }) == 2,
                    [=[#ecs.query({ all = { "d2legacy.state.event" } }) == 2]=]
                )
            end),
        }),
    },
})
