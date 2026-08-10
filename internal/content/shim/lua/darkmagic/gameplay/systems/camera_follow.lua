-- Copy a followed entity's world position into a presentation camera entity.
--
-- The camera is disposable presentation state. The target is authoritative
-- world state. This one-way copy keeps camera concerns out of the player entity.

local ecs = require("dm.ecs/v1")

local M = {}

local function follow(entity)
    local relationship = ecs.get(entity, "dm.world.camera_follow")
    local target_position = ecs.get(relationship:get("target"), "dm.world.position")
    local camera_position = ecs.get(entity, "dm.world.position")
    camera_position:set("x", target_position:get("x"))
    camera_position:set("y", target_position:get("y"))
end

function M.register()
    ecs.system({
        id = "darkmagic.world.camera_follow",
        phase = "presentation",
        query = { all = { "dm.world.position", "dm.world.camera_follow" } },
        read = { "dm.world.camera_follow" },
        write = { "dm.world.position" },
        update = function(_, entities)
            for _, entity in ipairs(entities) do follow(entity) end
        end,
    })
end

return M
