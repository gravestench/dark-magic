-- Integrate authoritative velocity into bounded, collision-aware position.
--
-- Input does not run here. A fixed-tick command handler has already translated
-- player intent into velocity. This system only applies that simulation fact to
-- every entity with Position + Velocity + Bounds.

local ecs = require("engine.ecs/v1")

local M = {}
local collisions = {}
local default_collision = nil

local function clamp(value, minimum, maximum)
    return math.max(minimum, math.min(maximum, value))
end

local function footprint_blocked(collision, x, y, radius)
    if not collision then
        return false
    end
    local reach = math.ceil(radius)
    for offset_y = -reach, reach do
        for offset_x = -reach, reach do
            -- Match the authoritative A* clearance sampler. The extra half
            -- subtile converts a continuous circle into collision-cell centers;
            -- a medium (radius 1) player therefore reserves the surrounding
            -- 3x3 neighborhood used by Riiablo's size-2 path clearance.
            local distance = math.sqrt(offset_x * offset_x + offset_y * offset_y)
            if distance <= radius + 0.5 and collision:blocked_position(x + offset_x, y + offset_y) then
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

local function collision_for(entity)
    local location = ecs.get(entity, "d2legacy.world.location")
    if location then
        return collisions[location:get("level_id")] or default_collision
    end
    return default_collision
end

local function update_entity(entity, elapsed)
    local position = ecs.get(entity, "d2legacy.world.position")
    local velocity = ecs.get(entity, "d2legacy.world.velocity")
    local bounds = ecs.get(entity, "d2legacy.world.bounds")
    local collider = ecs.get(entity, "d2legacy.world.collider")
    local collision = collision_for(entity)

    if collision then
        local x, y = collision:integrate_velocity(
            position:get("x"),
            position:get("y"),
            velocity:get("x"),
            velocity:get("y"),
            collider:get("radius"),
            bounds:get("width"),
            bounds:get("height"),
            elapsed
        )
        position:set("x", x)
        position:set("y", y)
        return
    end

    move_x(position, velocity, bounds, collider, collision, elapsed)
    move_y(position, velocity, bounds, collider, collision, elapsed)
end

function M.register(collision)
    default_collision = collision
    ecs.system({
        id = "d2legacy.world.integrate",
        phase = "movement",
        query = {
            all = {
                "d2legacy.world.position",
                "d2legacy.world.velocity",
                "d2legacy.world.bounds",
                "d2legacy.world.collider",
            },
        },
        read = {
            "d2legacy.world.velocity",
            "d2legacy.world.bounds",
            "d2legacy.world.collider",
            "d2legacy.world.location",
        },
        write = { "d2legacy.world.position" },
        update = function(context, entities)
            for _, entity in ipairs(entities) do
                update_entity(entity, context.delta_seconds)
            end
        end,
    })
end

function M.set_collision(level_id, collision)
    if collision == nil then
        default_collision = level_id
        return
    end
    assert(type(level_id) == "number" and level_id > 0, "collision level ID is required")
    collisions[level_id] = collision
end

return M
