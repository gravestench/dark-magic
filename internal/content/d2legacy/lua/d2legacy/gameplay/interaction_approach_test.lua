local test = require("d2legacy.tests/v1")
local approach = require("d2legacy.gameplay.interaction_approach")

local selected = { id = "warp:test", x = 10, y = 10, radius = 3.5 }

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("prediction_cannot_actuate_before_authoritative_route_completion", function(t)
            t:run(function()
                local ready, finished = approach.resolve(selected, 10, 10, true, function()
                    return true
                end)
                test.expect(ready):is_false()
                test.expect(finished):is_false()
            end)
        end),
        test.case("completed_route_uses_selected_radius_and_line_of_sight", function(t)
            t:run(function()
                local ready, finished = approach.resolve(selected, 13.5, 10, false, function()
                    return true
                end)
                test.expect(ready):is_true()
                test.expect(finished):is_true()

                ready, finished = approach.resolve(selected, 13.6, 10, false, function()
                    return true
                end)
                test.expect(ready):is_false()
                test.expect(finished):is_true()

                ready = approach.resolve(selected, 10, 10, false, function()
                    return false
                end)
                test.expect(ready):is_false()
            end)
        end),
    },
})
