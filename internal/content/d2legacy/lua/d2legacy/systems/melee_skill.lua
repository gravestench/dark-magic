-- Convert an admitted generic cast into the reusable melee-action adapter.

local ecs = require("engine.ecs/v1")
local M = {}

function M.register(definitions)
    ecs.system({
        id = "d2legacy.skill.emit_melee_action",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = {
            all = { "d2legacy.skill.cast", "d2legacy.player.identity" },
            none = { "d2legacy.player.death" },
        },
        read = { "d2legacy.skill.cast", "d2legacy.player.identity" },
        write = { "d2legacy.skill.cast", "d2legacy.skill.cast_event" },
        update = function(context, casters, structural)
            for _, caster in ipairs(casters) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                local definition = definitions[cast:get("skill_id")]
                if
                    definition
                    and definition.behavior == "action.melee"
                    and not cast:get("effect_emitted")
                    and context.tick >= cast:get("effect_tick")
                then
                    local identity = ecs.get(caster, "d2legacy.player.identity")
                    structural:create({
                        ["d2legacy.skill.cast_event"] = {
                            kind = "skill_effect",
                            tick = context.tick,
                            player = identity:get("player"),
                            skill_id = cast:get("skill_id"),
                            skill_level = cast:get("skill_level"),
                            behavior = definition.behavior,
                            animation_mode = definition.animation_mode,
                            weapon_selection = definition.weapon_selection,
                            target_x = cast:get("target_x"),
                            target_y = cast:get("target_y"),
                            target_id = cast:get("target_id"),
                            reason = "",
                        },
                    })
                    cast:set("effect_emitted", true)
                end
            end
        end,
    })
end

return M
