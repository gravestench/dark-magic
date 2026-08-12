-- Presentation-side constructors for d2legacy player commands.
--
-- These functions do not validate gameplay or mutate ECS state. They only turn
-- pointer/HUD choices into serializable requests. Authoritative command modules
-- authenticate and validate them on the next fixed simulation tick.

local intents = require("engine.command_intent/v1")
local movement = require("engine.player/v1")
local M = {}

M.request_running = movement.request_running
M.request_move = movement.request_move
M.movement_pending = movement.movement_pending

function M.assign_skill(side, skill_id)
    assert(side == "left" or side == "right", "skill side must be left or right")
    assert(type(skill_id) == "number" and skill_id >= 0, "skill ID must be non-negative")
    intents.submit("player.assign_skills", {[side] = skill_id})
end

function M.request_skill(side, x, y, target_id)
    intents.submit("player.use_skill", {
        side = side,
        target_x = x,
        target_y = y,
        target_id = target_id or "",
    })
end

return M
