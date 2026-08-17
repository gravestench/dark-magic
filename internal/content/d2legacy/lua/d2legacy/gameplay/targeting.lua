-- Pick the live world entity underneath a pointer.
--
-- This belongs to d2legacy because the component names, selectable kinds, and
-- priority field are Diablo mod vocabulary.  The engine only supplies ordered
-- ECS queries; this module decides what those facts mean to pointer gameplay.
--
-- Keep this file intentionally boring.  A pointer query copies component data
-- into ordinary Lua tables, scores each circular footprint, and forgets the ECS
-- handles before returning.  Presentation therefore cannot retain a writable
-- reference to authoritative state between frames.

local ecs = require("engine.ecs/v1")

local M = {}

local function finite(value)
    return type(value) == "number" and value == value and value ~= math.huge and value ~= -math.huge
end

local function copy_candidate(entity, pointer_x, pointer_y, level_id)
    local selectable = ecs.get(entity, "d2legacy.world.selectable")
    local position = ecs.get(entity, "d2legacy.world.position")
    if not selectable or not position then
        return nil
    end

    local radius = selectable:get("radius")
    local location = ecs.get(entity, "d2legacy.world.location")
    if level_id and location and location:get("level_id") ~= level_id then
        return nil
    end
    local x, y = position:get("x"), position:get("y")
    if not finite(radius) or radius <= 0 or not finite(x) or not finite(y) then
        return nil
    end

    local dx, dy = pointer_x - x, pointer_y - y
    local distance = (dx * dx + dy * dy) / (radius * radius)
    if distance > 1 then
        return nil
    end

    return {
        id = selectable:get("id"),
        kind = selectable:get("kind"),
        label = selectable:get("label"),
        owner = selectable:get("owner"),
        radius = radius,
        priority = selectable:get("priority"),
        x = x,
        y = y,
        distance = distance,
    }
end

local function comes_before(left, right)
    if left.priority ~= right.priority then
        return left.priority > right.priority
    end
    if left.distance ~= right.distance then
        return left.distance < right.distance
    end
    return left.id < right.id
end

function M.selectable_at(pointer_x, pointer_y, level_id)
    if not finite(pointer_x) or not finite(pointer_y) then
        return nil
    end

    local best
    for _, entity in
        ipairs(ecs.query({
            all = { "d2legacy.world.selectable", "d2legacy.world.position" },
        }))
    do
        local candidate = copy_candidate(entity, pointer_x, pointer_y, level_id)
        if candidate and (not best or comes_before(candidate, best)) then
            best = candidate
        end
    end

    if best then
        best.distance = nil
    end
    return best
end

return M
