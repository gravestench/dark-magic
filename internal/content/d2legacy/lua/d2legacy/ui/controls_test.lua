local test = require("d2legacy.tests/v1")

local function load_controls()
    test.mock_module("engine.input/v1", {}, {})
    return require("d2legacy.ui.controls")
end

return test.suite({
    name = "shared UI control state",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("disabling_updates_authored_visual_state_immediately", function(t)
            t:run(function()
                local controls = load_controls()
                local manager = controls.new()
                local observed = {}
                manager:add({
                    id = "action",
                    x = 0,
                    y = 0,
                    width = 100,
                    height = 30,
                    on_state = function(_, state)
                        observed[#observed + 1] = state
                    end,
                })

                manager:set_enabled("action", false)

                test.expect(manager:get("action").state):equals("disabled")
                test.expect(observed):deep_equals({ "disabled" })
            end)
        end),
        test.case("reenabling_restores_an_opaque_normal_state", function(t)
            t:run(function()
                local controls = load_controls()
                local manager = controls.new()
                local observed = {}
                manager:add({
                    id = "action",
                    x = 0,
                    y = 0,
                    width = 100,
                    height = 30,
                    enabled = false,
                    on_state = function(_, state)
                        observed[#observed + 1] = state
                    end,
                })

                manager:set_enabled("action", true)

                test.expect(manager:get("action").state):equals("normal")
                test.expect(observed):deep_equals({ "normal" })
            end)
        end),
    },
})
