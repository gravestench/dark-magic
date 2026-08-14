-- Keeps presentation scenes independent from the native transport lifecycle.
-- The capability is optional so the same scenes continue to work in UI-only tests.
local available, network = pcall(require, "engine.network/v1")

local M = {}

local function pending_operation()
    if not available then
        return false
    end
    local status = network.status()
    return status.phase == "selecting" and (status.mode == "host" or status.mode == "join")
end

function M.start_selected()
    if not available then
        return true
    end
    return network.start_selected()
end

function M.cancel_destination(default_scene)
    if not pending_operation() then
        return default_scene
    end
    network.cancel()
    return "tcpip"
end

return M
