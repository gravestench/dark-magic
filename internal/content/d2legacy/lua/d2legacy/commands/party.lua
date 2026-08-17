-- Player-bound semantic commands for authoritative party relationships.

local commands = require("engine.authority_command/v1")
local party = require("d2legacy.policy.party")

local M = {}

local function player_field(command, field)
    assert(type(command.payload) == "table", "party command payload is required")
    local value = command.payload[field]
    assert(type(value) == "string" and value:match("%S"), "party command " .. field .. " is required")
end

function M.register()
    commands.register({
        kind = "party.invite",
        authorities = { "player" },
        validate = function(command)
            player_field(command, "target")
        end,
        apply = function(command)
            party.invite(command.player, command.payload.target, command.tick)
        end,
    })
    commands.register({
        kind = "party.cancel_invite",
        authorities = { "player" },
        validate = function(command)
            player_field(command, "target")
        end,
        apply = function(command)
            party.cancel(command.player, command.payload.target)
        end,
    })
    commands.register({
        kind = "party.accept",
        authorities = { "player" },
        validate = function(command)
            player_field(command, "inviter")
        end,
        apply = function(command)
            party.accept(command.player, command.payload.inviter)
        end,
    })
    commands.register({
        kind = "party.leave",
        authorities = { "player" },
        validate = function(command)
            assert(type(command.payload) == "table", "party leave payload is required")
        end,
        apply = function(command)
            party.leave(command.player)
        end,
    })
end

return M
