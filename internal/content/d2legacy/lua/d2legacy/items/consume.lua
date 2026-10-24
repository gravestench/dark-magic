local ecs = require("engine.ecs/v1")
local M = {}

function M.destroy(item_entity)
    -- Item modifiers reference their item identity. Retire them first so
    -- equipment and checkpoints never encounter dangling item references.
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.item.stat_modifier" } })) do
        local modifier = ecs.get(entity, "d2legacy.item.stat_modifier")
        if modifier:get("item"):id() == item_entity:id() then
            ecs.destroy(entity)
        end
    end
    ecs.destroy(item_entity)
end

return M
