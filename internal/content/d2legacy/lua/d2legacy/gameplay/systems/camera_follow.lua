-- Copy a followed entity's world position into a presentation camera entity.
--
-- The camera is disposable presentation state. The target is authoritative
-- world state. This one-way copy keeps camera concerns out of the player entity.

local ecs = require("engine.ecs/v1")

local M = {}

local function follow(entity)
    local relationship = ecs.get(entity, "d2legacy.world.camera_follow")
    local target_position = ecs.get(relationship:get("target"), "d2legacy.world.position")
    local camera_position = ecs.get(entity, "d2legacy.world.position")
    camera_position:set("x", target_position:get("x"))
    camera_position:set("y", target_position:get("y"))
end

function M.register()
    ecs.system({
        id = "d2legacy.world.camera_follow",
        phase = "presentation",
        query = { all = { "d2legacy.world.position", "d2legacy.world.camera_follow" } },
        read = { "d2legacy.world.camera_follow" },
        write = { "d2legacy.world.position" },
        update = function(_, entities)
            for _, entity in ipairs(entities) do follow(entity) end
        end,
    })
end

return M
