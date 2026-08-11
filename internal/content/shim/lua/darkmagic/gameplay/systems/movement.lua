-- Integrate authoritative velocity into bounded, collision-aware position.
--
-- Input does not run here. A fixed-tick command handler has already translated
-- player intent into velocity. This system only applies that simulation fact to
-- every entity with Position + Velocity + Bounds.

local ecs = require("dm.ecs/v1")

local M = {}

local function clamp(value, minimum, maximum)
    return math.max(minimum, math.min(maximum, value))
end

local function collision_cell(value)
    -- Positions describe subtile centers. Adding one half before flooring maps
    -- a centered footprint back onto the cell containing that center.
    return math.floor(value + 0.5)
end

local function footprint_blocked(collision, x, y, radius)
    if not collision then return false end
    for _, offset_x in ipairs({-radius, radius}) do
        for _, offset_y in ipairs({-radius, radius}) do
            if collision:blocked(collision_cell(x + offset_x), collision_cell(y + offset_y)) then
                return true
            end
        end
    end
    return false
end

local function move_x(position, velocity, bounds, collider, collision, elapsed)
    local x, y = position:get("x"), position:get("y")
    local radius = collider:get("radius")
    local candidate = clamp(x + velocity:get("x") * elapsed, radius, bounds:get("width") - radius)

    -- Collision cells are integers; continuous movement is rounded only at the
    -- query boundary so slow movement does not lose fractional progress.
    if not footprint_blocked(collision, candidate, y, radius) then
        position:set("x", candidate)
    end
end

local function move_y(position, velocity, bounds, collider, collision, elapsed)
    local x, y = position:get("x"), position:get("y")
    local radius = collider:get("radius")
    local candidate = clamp(y + velocity:get("y") * elapsed, radius, bounds:get("height") - radius)

    -- X is read again after the X step. Resolving the axes separately gives a
    -- simple sliding response when only one side of a diagonal is blocked.
    if not footprint_blocked(collision, x, candidate, radius) then
        position:set("y", candidate)
    end
end

local function update_entity(entity, collision, elapsed)
    local position = ecs.get(entity, "dm.world.position")
    local velocity = ecs.get(entity, "dm.world.velocity")
    local bounds = ecs.get(entity, "dm.world.bounds")
    local collider = ecs.get(entity, "dm.world.collider")
    move_x(position, velocity, bounds, collider, collision, elapsed)
    move_y(position, velocity, bounds, collider, collision, elapsed)
end

function M.register(collision)
    ecs.system({
        id = "darkmagic.world.integrate",
        phase = "movement",
        query = { all = { "dm.world.position", "dm.world.velocity", "dm.world.bounds", "dm.world.collider" } },
        read = { "dm.world.velocity", "dm.world.bounds", "dm.world.collider" },
        write = { "dm.world.position" },
        update = function(context, entities)
            for _, entity in ipairs(entities) do
                update_entity(entity, collision, context.delta_seconds)
            end
        end,
    })
end

return M
