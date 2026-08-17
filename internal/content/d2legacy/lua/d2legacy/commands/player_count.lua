-- Typed authority mutations backing the host-authorized `/players X` command.
-- The admission cap remains immutable GameRules data and is never changed here.

local commands = require("engine.authority_command/v1")
local player_count = require("d2legacy.policy.player_count")

local M = {}

function M.validate_override(command)
    local payload = command.payload
    local value = payload and payload.count
    assert(type(value) == "number" and value == math.floor(value), "player-count override must be an integer")
    assert(value >= 1 and value <= 8, "player-count override must be from 1 through 8")
end

function M.validate_follow(command)
    assert(type(command.payload) == "table", "player-count follow command payload is required")
end

function M.register()
    commands.register({
        kind = "game.player_count.override",
        authorities = { "system", "administrator" },
        validate = M.validate_override,
        apply = function(command)
            player_count.set_override(command.payload.count)
        end,
    })
    commands.register({
        kind = "game.player_count.follow_population",
        authorities = { "system", "administrator" },
        validate = M.validate_follow,
        apply = function()
            player_count.clear_override()
        end,
    })
end

return M
