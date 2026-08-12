return {
    tests = {
        applied_state_expires_at_the_authored_tick = {
            {run = function()
                local ecs = require("engine.ecs/v1")
                local target = ecs.create()
                ecs.create({["d2legacy.state.request"]={
                    operation="apply", target=target, state_id="poison",
                    source_id="monster:fallen", duration=2,
                    policy="refresh_same_source",
                }})
            end},
            {step = 1},
            {run = function()
                local ecs = require("engine.ecs/v1")
                assert(#ecs.query({all={"d2legacy.state.instance"}}) == 1)
            end},
            {checkpoint_restore = true},
            {step = 2},
            {run = function()
                local ecs = require("engine.ecs/v1")
                assert(#ecs.query({all={"d2legacy.state.instance"}}) == 0)
                assert(#ecs.query({all={"d2legacy.state.event"}}) == 2)
            end},
        },
    },
}
