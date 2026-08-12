-- Admit a player's request to use a skill.
--
-- Go has already authenticated the player, tick, and sequence when these
-- functions run. Validation below is Diablo policy: the payload must describe
-- one finite target and a positive learned-skill level.

local commands = require("dm.authority_command/v1")
local ecs = require("dm.ecs/v1")

local M = {}

local function finite(value)
    return type(value) == "number" and value == value
        and value ~= math.huge and value ~= -math.huge
end

local function find_player(player_id)
    for _, entity in ipairs(ecs.query({ all = { "dm.player.identity" } })) do
        local identity = ecs.get(entity, "dm.player.identity")
        if identity:get("player") == player_id then return entity end
    end
    return nil
end

function M.validate(command)
    local payload = command.payload
    assert(type(payload) == "table", "cast payload must be a table")
    assert(payload.skill_id == 36, "only migrated skill 36 is available")
    assert(type(payload.skill_level) == "number" and payload.skill_level >= 1,
        "skill level must be positive")
    assert(finite(payload.target_x) and finite(payload.target_y),
        "cast target must be finite")
    assert(payload.target_id == nil or type(payload.target_id) == "string",
        "target ID must be a string")
end

function M.apply(command)
    local player = assert(find_player(command.player), "cast player does not exist")
    local payload = command.payload
    ecs.set(player, "d2legacy.skill.cast_request", {
        player = command.player,
        skill_id = payload.skill_id,
        skill_level = payload.skill_level,
        target_x = payload.target_x,
        target_y = payload.target_y,
        target_id = payload.target_id or "",
        request_tick = command.tick,
    })
end

function M.register()
    commands.register({
        kind = "d2legacy.skill.cast",
        authorities = { "player" },
        validate = M.validate,
        apply = M.apply,
    })
end

return M
