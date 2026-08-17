-- Resolve the presentation-side end of a click-to-operate approach.
--
-- The hero position may be locally predicted, while movement_pending is backed
-- by the authoritative ECS route. Never actuate from prediction alone.

local M = {}

function M.resolve(selected, hero_x, hero_y, movement_pending, line_clear)
    assert(type(selected) == "table", "selected interaction is required")
    assert(type(line_clear) == "function", "line-clear query is required")
    if movement_pending then
        return false, false
    end
    local radius = selected.radius or 0
    local dx, dy = hero_x - selected.x, hero_y - selected.y
    local ready = radius > 0
        and dx * dx + dy * dy <= radius * radius
        and line_clear(hero_x, hero_y, selected.x, selected.y)
    return ready, true
end

return M
