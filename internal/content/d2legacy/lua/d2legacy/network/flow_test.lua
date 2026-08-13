local test = require("d2legacy.tests/v1")

local function load_flow(status)
    local calls = { started = 0, cancelled = 0 }
    test.mock_module("engine.network/v1", {
        status = function()
            return status
        end,
        start_selected = function()
            calls.started = calls.started + 1
            return true
        end,
        cancel = function()
            calls.cancelled = calls.cancelled + 1
        end,
    }, { "status", "start_selected", "cancel" })
    return require("d2legacy.network.flow"), calls
end

return test.suite({
    name = "deferred network character flow",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("pending_host_starts_after_character_selection", function(t)
            t:run(function()
                local flow, calls = load_flow({ phase = "selecting", mode = "host" })
                test.expect(flow.start_selected()):is_true()
                test.expect(calls.started):equals(1)
            end)
        end),
        test.case("offline_selection_does_not_touch_transport", function(t)
            t:run(function()
                local flow, calls = load_flow({ phase = "idle", mode = "" })
                test.expect(flow.start_selected()):is_true()
                test.expect(calls.started):equals(0)
            end)
        end),
        test.case("pending_join_starts_after_character_selection", function(t)
            t:run(function()
                local flow, calls = load_flow({ phase = "selecting", mode = "join" })
                test.expect(flow.start_selected()):is_true()
                test.expect(calls.started):equals(1)
            end)
        end),
        test.case("cancelling_pending_host_returns_to_tcpip", function(t)
            t:run(function()
                local flow, calls = load_flow({ phase = "selecting", mode = "host" })
                test.expect(flow.cancel_destination("main_menu")):equals("tcpip")
                test.expect(calls.cancelled):equals(1)
            end)
        end),
    },
})
