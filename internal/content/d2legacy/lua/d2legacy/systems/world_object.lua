-- Data-selected authoritative world-object operation families.
--
-- Only a synthetic one-shot family is admitted today. It proves that object
-- state is ordinary ECS authority and interaction dispatch is reusable; owned
-- Expansion 1.14d Objects/TBL evidence must admit each retail family later.

local ecs = require("engine.ecs/v1")
local operations = require("d2legacy.interactions.operations")

local M = {}

function M.register()
    operations.register("d2legacy.object.once_operation", function(_, entity, operation)
        local state = assert(ecs.get(entity, "d2legacy.object.state"), "object state is required")
        if state:get("disabled") or state:get("locked") or state:get("used") then
            return
        end
        state:set("mode", operation:get("result_mode"))
        state:set("used", true)
        state:set("revision", state:get("revision") + 1)
    end)
end

return M
