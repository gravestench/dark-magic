-- Resolve a stable gameplay identity for semantic events.

local ecs = require("engine.ecs/v1")

local M = {}

function M.semantic_id(entity)
    local selectable = ecs.get(entity, "d2legacy.world.selectable")
    if selectable then
        return selectable:get("id")
    end
    return "entity:" .. entity:id()
end

return M
