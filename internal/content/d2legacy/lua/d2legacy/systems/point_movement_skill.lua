-- Resolve point-target relocation families as atomic same-level movement.

local ecs = require("engine.ecs/v1")
local collision = require("d2legacy.gameplay.collision")
local entity_identity = require("d2legacy.policy.entity_identity")
local player_motion = require("d2legacy.gameplay.player_motion")

local M = {}

local function same_level(left, right)
    return left and right and left:get("act") == right:get("act") and left:get("level_id") == right:get("level_id")
end

local function dynamically_clear(entity, units, x, y, radius)
    local location = ecs.get(entity, "d2legacy.world.location")
    for _, other in ipairs(units) do
        if
            other.entity:id() ~= entity:id()
            and other.occupancy:get("blocks_movement")
            and same_level(location, other.location)
        then
            local dx, dy = x - other.position:get("x"), y - other.position:get("y")
            local combined = radius + other.collider:get("radius")
            if dx * dx + dy * dy < combined * combined then
                return false
            end
        end
    end
    return true
end

local function outcome(entity, cast, definition, units)
    local position = ecs.get(entity, "d2legacy.world.position")
    local location = ecs.get(entity, "d2legacy.world.location")
    local bounds = ecs.get(entity, "d2legacy.world.bounds")
    local collider = ecs.get(entity, "d2legacy.world.collider")
    local level_id = location:get("level_id")
    local policy = definition.level_policy[level_id]
    if policy == nil or policy == 0 then
        return "level-disabled"
    end
    local x, y, radius = cast:get("target_x"), cast:get("target_y"), collider:get("radius")
    if x < radius or y < radius or x > bounds:get("width") - radius or y > bounds:get("height") - radius then
        return "outside-world"
    end
    if
        not collision.destination_clear(level_id, x, y, radius) or not dynamically_clear(entity, units, x, y, radius)
    then
        return "blocked"
    end
    if policy == 2 and not collision.line_clear(level_id, position:get("x"), position:get("y"), x, y) then
        return "line-blocked"
    end
    return "completed"
end

local function relocate(context, entity, cast, definition, units, structural)
    local position = ecs.get(entity, "d2legacy.world.position")
    local start_x, start_y = position:get("x"), position:get("y")
    local result = outcome(entity, cast, definition, units)
    if result == "completed" then
        position:set("x", cast:get("target_x"))
        position:set("y", cast:get("target_y"))
        local velocity = ecs.get(entity, "d2legacy.world.velocity")
        velocity:set("x", 0)
        velocity:set("y", 0)
        player_motion.stop(entity)
        structural:remove(entity, "d2legacy.world.forced_motion")
    end
    structural:create({
        ["d2legacy.world.relocation_event"] = {
            kind = "skill",
            outcome = result,
            tick = context.tick,
            source_id = "skill:" .. entity_identity.semantic_id(entity) .. ":" .. definition.skill_id,
            target_id = entity_identity.semantic_id(entity),
            start_x = start_x,
            start_y = start_y,
            end_x = position:get("x"),
            end_y = position:get("y"),
        },
    })
    cast:set("effect_emitted", true)
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.skill.point_movement",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = {
            any = { "d2legacy.skill.cast", "d2legacy.world.occupancy" },
            none = { "d2legacy.player.death", "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.skill.cast",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.bounds",
            "d2legacy.world.collider",
            "d2legacy.world.occupancy",
            "d2legacy.world.selectable",
            "d2legacy.player.death",
            "d2legacy.world.inactive",
        },
        write = {
            "d2legacy.skill.cast",
            "d2legacy.world.position",
            "d2legacy.world.velocity",
            "d2legacy.player.motion",
            "d2legacy.world.forced_motion",
            "d2legacy.world.relocation_event",
        },
        update = function(context, entities, structural)
            local units = {}
            for _, entity in ipairs(entities) do
                local occupancy = ecs.get(entity, "d2legacy.world.occupancy")
                local position = ecs.get(entity, "d2legacy.world.position")
                local collider = ecs.get(entity, "d2legacy.world.collider")
                if occupancy and position and collider then
                    units[#units + 1] = {
                        entity = entity,
                        occupancy = occupancy,
                        position = position,
                        collider = collider,
                        location = ecs.get(entity, "d2legacy.world.location"),
                    }
                end
            end
            for _, entity in ipairs(entities) do
                local cast = ecs.get(entity, "d2legacy.skill.cast")
                local definition = cast and definitions[cast:get("skill_id")]
                if
                    definition
                    and definition.behavior == "movement.point-relocate"
                    and not cast:get("effect_emitted")
                    and context.tick >= cast:get("effect_tick")
                then
                    relocate(context, entity, cast, definition, units, structural)
                end
            end
        end,
    })
end

return M
