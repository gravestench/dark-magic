-- Small geometry helpers shared by projectile policies.
-- These functions know nothing about Diablo entities, damage, or targeting.

local M = {}

function M.normalized_direction(from_x, from_y, to_x, to_y)
    local dx, dy = to_x - from_x, to_y - from_y
    local length = math.sqrt(dx * dx + dy * dy)
    assert(length > 0, "target must differ from projectile origin")
    return dx / length, dy / length
end

-- Return distance from a point to the closest point on a finite segment. The
-- second result is the [0,1] position along the segment for stable hit ordering.
function M.segment_distance(ax, ay, bx, by, px, py)
    local dx, dy = bx - ax, by - ay
    local length_squared = dx * dx + dy * dy
    local along = 0
    if length_squared > 0 then
        along = ((px - ax) * dx + (py - ay) * dy) / length_squared
        along = math.max(0, math.min(1, along))
    end
    local closest_x, closest_y = ax + along * dx, ay + along * dy
    local x, y = px - closest_x, py - closest_y
    return math.sqrt(x * x + y * y), along
end

return M
