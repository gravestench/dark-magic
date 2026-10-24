-- Move players through prepared level seams during the authoritative collision tick.
-- Geometry is shared with presentation, but only this system relocates entities.

local ecs = require("engine.ecs/v1")
local geometry = require("d2legacy.gameplay.seam_geometry")
local transition = require("d2legacy.gameplay.transition")

local M = {}

-- Keep the public trigger query on the same geometry used by collision ticks.
function M.at_trigger(level, x, y)
    return geometry.at_trigger(level, x, y)
end

-- Register crossing at the collision barrier after movement updates player position.
function M.register()
    local seams = geometry.configured_seams()
    ecs.system({
        id = "d2legacy.world.transition",
        -- Collision follows movement in the engine's stable phase order. A
        -- seam is geometry too, so crossing it belongs at this barrier.
        phase = "collision",
        query = {
            all = {
                "d2legacy.world.player_control",
                "d2legacy.player.motion",
                "d2legacy.world.location",
                "d2legacy.world.position",
                "d2legacy.world.bounds",
                "d2legacy.world.velocity",
            },
        },
        read = {
            "d2legacy.world.player_control",
            "d2legacy.world.forced_motion",
            "d2legacy.world.selectable",
        },
        write = {
            "d2legacy.world.location",
            "d2legacy.world.position",
            "d2legacy.world.bounds",
            "d2legacy.world.velocity",
            "d2legacy.player.motion",
            "d2legacy.world.forced_motion",
            "d2legacy.world.forced_motion_event",
        },
        update = function(context, entities, structural)
            for _, entity in ipairs(entities) do
                local location = ecs.get(entity, "d2legacy.world.location")
                local position = ecs.get(entity, "d2legacy.world.position")
                local seam =
                    geometry.matching_seam(seams, location:get("level_id"), position:get("x"), position:get("y"))
                if seam then
                    transition.apply(context, entity, {
                        level_id = seam.destination_level,
                        x = seam.arrival_x,
                        y = seam.arrival_y,
                        width = seam.world_width,
                        height = seam.world_height,
                    }, structural)
                end
            end
        end,
    })
end

return M
