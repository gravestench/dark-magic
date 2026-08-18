local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("one_click_submits_one_cast", {
            test.run(function()
                local gate = require("d2legacy.gameplay.held_skill_input")
                local state
                local ready
                ready, state = gate.update(state, true, true, false)
                test.assert(ready, [=[the press edge is admitted]=])
                gate.submitted(state)
                ready, state = gate.update(state, false, true, false)
                test.assert(not ready, [=[later rendered frames in the same click do not queue another cast]=])
                ready, state = gate.update(state, false, false, false)
                test.assert(not ready, [=[release only rearms the next physical click]=])
            end),
        }),
        test.case("held_pointer_repeats_after_authoritative_completion", {
            test.run(function()
                local gate = require("d2legacy.gameplay.held_skill_input")
                local ready, state = gate.update(nil, true, true, false)
                test.assert(ready)
                gate.submitted(state)
                ready, state = gate.update(state, false, true, true)
                test.assert(not ready, [=[an active cast cannot accumulate a queued cast]=])
                ready, state = gate.update(state, false, true, false)
                test.assert(ready, [=[continued hold repeats exactly when authority becomes idle]=])
            end),
        }),
    },
})
