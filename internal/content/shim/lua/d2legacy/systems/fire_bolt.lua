-- Create one Fire Bolt projectile when a cast reaches its effect tick.

local ecs = require("dm.ecs/v1")
local geometry = require("d2legacy.policy.geometry")

local M = {}

local function selectable_id(entity)
    local selectable = ecs.get(entity, "dm.world.selectable")
    if selectable then return selectable:get("id") end
    local identity = ecs.get(entity, "dm.player.identity")
    if identity then return "player:" .. identity:get("player") end
    return "entity:" .. entity:id()
end

function M.register(definition)
    ecs.system({
        id = "d2legacy.missile.spawn_fire_bolt",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = { all = {
            "d2legacy.skill.cast", "dm.world.position", "dm.world.location",
        } },
        read = {
            "d2legacy.skill.cast", "dm.world.position", "dm.world.location",
            "dm.world.selectable", "dm.player.identity",
        },
        write = {
            "d2legacy.skill.cast", "d2legacy.missile.projectile",
            "dm.world.position", "dm.world.location",
        },
        update = function(context, casters, structural)
            for _, caster in ipairs(casters) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                if not cast:get("effect_emitted")
                    and context.tick >= cast:get("effect_tick") then
                    local position = ecs.get(caster, "dm.world.position")
                    local location = ecs.get(caster, "dm.world.location")
                    local dx, dy = geometry.normalized_direction(
                        position:get("x"), position:get("y"),
                        cast:get("target_x"), cast:get("target_y"))
                    structural:create({
                        ["dm.world.position"] = {
                            x = position:get("x"), y = position:get("y"),
                        },
                        ["dm.world.location"] = location:snapshot(),
                        ["d2legacy.missile.projectile"] = {
                            owner_id = selectable_id(caster),
                            target_x = cast:get("target_x"), target_y = cast:get("target_y"),
                            velocity_x = dx * definition.speed_per_tick,
                            velocity_y = dy * definition.speed_per_tick,
                            previous_x = position:get("x"),
                            previous_y = position:get("y"),
                            remaining_ticks = definition.lifetime_ticks,
                            collision_radius = definition.collision_radius,
                            minimum_damage_raw = definition.minimum_damage_raw,
                            maximum_damage_raw = definition.maximum_damage_raw,
                            damage_channel = definition.damage_channel,
                        },
                    })
                    cast:set("effect_emitted", true)
                end
            end
        end,
    })
end

return M
