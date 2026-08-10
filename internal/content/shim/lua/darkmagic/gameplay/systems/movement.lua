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

local function move_x(position, velocity, bounds, collision, elapsed)
    local x, y = position:get("x"), position:get("y")
    local candidate = clamp(x + velocity:get("x") * elapsed, 0, bounds:get("width"))

    -- Collision cells are integers; continuous movement is rounded only at the
    -- query boundary so slow movement does not lose fractional progress.
    if not collision or not collision:blocked(math.floor(candidate), math.floor(y)) then
        position:set("x", candidate)
    end
end

local function move_y(position, velocity, bounds, collision, elapsed)
    local x, y = position:get("x"), position:get("y")
    local candidate = clamp(y + velocity:get("y") * elapsed, 0, bounds:get("height"))

    -- X is read again after the X step. Resolving the axes separately gives a
    -- simple sliding response when only one side of a diagonal is blocked.
    if not collision or not collision:blocked(math.floor(x), math.floor(candidate)) then
        position:set("y", candidate)
    end
end

local function update_entity(entity, collision, elapsed)
    local position = ecs.get(entity, "dm.world.position")
    local velocity = ecs.get(entity, "dm.world.velocity")
    local bounds = ecs.get(entity, "dm.world.bounds")
    move_x(position, velocity, bounds, collision, elapsed)
    move_y(position, velocity, bounds, collision, elapsed)
end

function M.register(collision)
    ecs.system({
        id = "darkmagic.world.integrate",
        phase = "movement",
        query = { all = { "dm.world.position", "dm.world.velocity", "dm.world.bounds" } },
        read = { "dm.world.velocity", "dm.world.bounds" },
        write = { "dm.world.position" },
        update = function(context, entities)
            for _, entity in ipairs(entities) do
                update_entity(entity, collision, context.delta_seconds)
            end
        end,
    })
end

return M
