-- Keep the authoritative collision map registry shared by every gameplay
-- consumer. Maps are immutable native handles; only which level owns which
-- handle changes as entry worlds and transitions materialize.

local M = {}
local levels = {}
local fallback = nil

function M.set(level_id, collision)
    if collision == nil then
        fallback = level_id
        return
    end
    assert(type(level_id) == "number" and level_id > 0, "collision level ID is required")
    levels[level_id] = collision
end

function M.for_level(level_id)
    if level_id then
        return levels[level_id] or fallback
    end
    return fallback
end

-- Diablo's melee-range trace uses the flying/missile-barrier collision bit,
-- represented by DT1 BlockJump, rather than the visual BlockLOS bit. Tests may
-- provide a policy-compatible table; production receives engine.world/v1 maps.
function M.melee_clear(level_id, from_x, from_y, to_x, to_y)
    local collision = M.for_level(level_id)
    if not collision then
        return true
    end
    return collision:barrier_clear(from_x, from_y, to_x, to_y)
end

return M
