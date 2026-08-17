-- Integrate authoritative velocity into bounded, collision-aware position.
--
-- Input does not run here. A fixed-tick command handler has already translated
-- player intent into velocity. This system only applies that simulation fact to
-- every entity with Position + Velocity + Bounds.

local ecs = require("engine.ecs/v1")
local entity_identity = require("d2legacy.policy.entity_identity")
local collision_registry = require("d2legacy.gameplay.collision")

local M = {}
local EPSILON = 0.000000001

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

local function dynamically_blocked(entity, units, x, y, radius)
    local location = ecs.get(entity, "d2legacy.world.location")
    local current = ecs.get(entity, "d2legacy.world.position")
    for _, other in ipairs(units) do
        if other.entity:id() ~= entity:id() and other.occupancy:get("blocks_movement") then
            local other_location = other.location
            if
                (not location and not other_location)
                or (
                    location
                    and other_location
                    and location:get("act") == other_location:get("act")
                    and location:get("level_id") == other_location:get("level_id")
                )
            then
                local dx = x - other.position:get("x")
                local dy = y - other.position:get("y")
                local combined = radius + other.collider:get("radius")
                local candidate_distance = dx * dx + dy * dy
                if candidate_distance < combined * combined then
                    local current_dx = current:get("x") - other.position:get("x")
                    local current_dy = current:get("y") - other.position:get("y")
                    local current_distance = current_dx * current_dx + current_dy * current_dy
                    -- Admission/teleport anchors can temporarily overlap. Let
                    -- a unit separate from an existing overlap, never deepen
                    -- it or enter a previously disjoint footprint.
                    if candidate_distance <= current_distance then
                        return true
                    end
                end
            end
        end
    end
    return false
end

local function move_x(entity, units, position, velocity_x, bounds, collider, collision, elapsed)
    if velocity_x == 0 then
        return
    end
    local x, y = position:get("x"), position:get("y")
    local radius = collider:get("radius")
    local candidate = clamp(x + velocity_x * elapsed, radius, bounds:get("width") - radius)

    -- Collision cells are integers; continuous movement is rounded only at the
    -- query boundary so slow movement does not lose fractional progress.
    if
        not footprint_blocked(collision, candidate, y, radius)
        and not dynamically_blocked(entity, units, candidate, y, radius)
    then
        position:set("x", candidate)
    end
end

local function move_y(entity, units, position, velocity_y, bounds, collider, collision, elapsed)
    if velocity_y == 0 then
        return
    end
    local x, y = position:get("x"), position:get("y")
    local radius = collider:get("radius")
    local candidate = clamp(y + velocity_y * elapsed, radius, bounds:get("height") - radius)

    -- X is read again after the X step. Resolving the axes separately gives a
    -- simple sliding response when only one side of a diagonal is blocked.
    if
        not footprint_blocked(collision, x, candidate, radius)
        and not dynamically_blocked(entity, units, x, candidate, radius)
    then
        position:set("y", candidate)
    end
end

local function collision_for(entity)
    local location = ecs.get(entity, "d2legacy.world.location")
    if location then
        return collision_registry.for_level(location:get("level_id"))
    end
    return collision_registry.for_level(nil)
end

local function forced_velocity(position, forced, elapsed)
    local dx = forced:get("target_x") - position:get("x")
    local dy = forced:get("target_y") - position:get("y")
    local distance = math.sqrt(dx * dx + dy * dy)
    if distance <= EPSILON or elapsed <= 0 then
        return 0, 0
    end
    local step = math.min(forced:get("speed") * elapsed, distance)
    return dx / distance * step / elapsed, dy / distance * step / elapsed
end

local function finish_forced(context, entity, position, velocity, forced, outcome, applied, structural)
    velocity:set("x", 0)
    velocity:set("y", 0)
    structural:create({
        ["d2legacy.world.forced_motion_event"] = {
            kind = forced:get("kind"),
            outcome = outcome,
            tick = context.tick,
            target_id = entity_identity.semantic_id(entity),
            start_x = forced:get("start_x"),
            start_y = forced:get("start_y"),
            end_x = position:get("x"),
            end_y = position:get("y"),
            requested_distance = forced:get("requested_distance"),
            applied_distance = applied,
        },
    })
    structural:remove(entity, "d2legacy.world.forced_motion")
end

local function update_entity(context, entity, units, structural)
    local position = ecs.get(entity, "d2legacy.world.position")
    local velocity = ecs.get(entity, "d2legacy.world.velocity")
    local bounds = ecs.get(entity, "d2legacy.world.bounds")
    local collider = ecs.get(entity, "d2legacy.world.collider")
    local collision = collision_for(entity)
    local forced = ecs.get(entity, "d2legacy.world.forced_motion")
    local velocity_x, velocity_y = velocity:get("x"), velocity:get("y")
    if forced then
        velocity_x, velocity_y = forced_velocity(position, forced, context.delta_seconds)
        velocity:set("x", velocity_x)
        velocity:set("y", velocity_y)
        if velocity_x == 0 and velocity_y == 0 then
            finish_forced(
                context,
                entity,
                position,
                velocity,
                forced,
                "completed",
                forced:get("applied_distance"),
                structural
            )
            return
        end
    end
    if velocity_x == 0 and velocity_y == 0 then
        return
    end

    local before_x, before_y = position:get("x"), position:get("y")

    if collision then
        local x, y = collision:integrate_velocity(
            position:get("x"),
            position:get("y"),
            velocity_x,
            velocity_y,
            collider:get("radius"),
            bounds:get("width"),
            bounds:get("height"),
            context.delta_seconds
        )
        local radius = collider:get("radius")
        if velocity_x ~= 0 and not dynamically_blocked(entity, units, x, position:get("y"), radius) then
            position:set("x", x)
        end
        if velocity_y ~= 0 and not dynamically_blocked(entity, units, position:get("x"), y, radius) then
            position:set("y", y)
        end
    else
        move_x(entity, units, position, velocity_x, bounds, collider, collision, context.delta_seconds)
        move_y(entity, units, position, velocity_y, bounds, collider, collision, context.delta_seconds)
    end

    if forced then
        local moved = math.max(math.abs(position:get("x") - before_x), math.abs(position:get("y") - before_y))
        local applied = forced:get("applied_distance") + moved
        forced:set("applied_distance", applied)
        local remaining_x = forced:get("target_x") - position:get("x")
        local remaining_y = forced:get("target_y") - position:get("y")
        if math.sqrt(remaining_x * remaining_x + remaining_y * remaining_y) <= EPSILON then
            finish_forced(context, entity, position, velocity, forced, "completed", applied, structural)
        elseif moved <= EPSILON then
            finish_forced(
                context,
                entity,
                position,
                velocity,
                forced,
                applied > EPSILON and "partial" or "blocked",
                applied,
                structural
            )
        end
    end
end

function M.register(collision)
    collision_registry.set(collision)
    ecs.system({
        id = "d2legacy.world.integrate",
        phase = "movement",
        query = {
            all = {
                "d2legacy.world.position",
                "d2legacy.world.velocity",
                "d2legacy.world.bounds",
                "d2legacy.world.collider",
                "d2legacy.world.occupancy",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.world.velocity",
            "d2legacy.world.bounds",
            "d2legacy.world.collider",
            "d2legacy.world.location",
            "d2legacy.world.occupancy",
            "d2legacy.world.forced_motion",
            "d2legacy.world.selectable",
        },
        write = {
            "d2legacy.world.position",
            "d2legacy.world.velocity",
            "d2legacy.world.forced_motion",
            "d2legacy.world.forced_motion_event",
        },
        update = function(context, entities, structural)
            local units = {}
            for _, entity in ipairs(entities) do
                units[#units + 1] = {
                    entity = entity,
                    position = ecs.get(entity, "d2legacy.world.position"),
                    collider = ecs.get(entity, "d2legacy.world.collider"),
                    occupancy = ecs.get(entity, "d2legacy.world.occupancy"),
                    location = ecs.get(entity, "d2legacy.world.location"),
                }
            end
            for _, entity in ipairs(entities) do
                update_entity(context, entity, units, structural)
            end
        end,
    })
end

function M.set_collision(level_id, collision)
    collision_registry.set(level_id, collision)
end

return M
