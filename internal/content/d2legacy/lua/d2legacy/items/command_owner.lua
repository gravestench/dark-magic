-- Resolve which durable item layout a command is allowed to change.

local M = {}

function M.resolve(command)
    local requested = command.payload.owner
    if command.authority == "player" then
        local owns_request = not requested
            or requested == ""
            or requested == command.player
        assert(owns_request, "cannot change another owner's items")
    end
    if requested and requested ~= "" then return requested end
    return command.player
end

return M
