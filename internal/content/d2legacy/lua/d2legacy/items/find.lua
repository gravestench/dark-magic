-- Deterministic item and owner lookup helpers.

local ecs = require("engine.ecs/v1")
local M = {}

function M.layout(owner, entities)
    entities = entities or ecs.query({
        all = { "d2legacy.items.layout" },
    })
    for _, entity in ipairs(entities) do
        local layout = ecs.get(entity, "d2legacy.items.layout")
        if layout and layout:get("owner") == owner then
            return entity, layout
        end
    end
    return nil, nil
end

function M.item(owner_entity, item_id, entities)
    entities = entities or ecs.query({
        all = { "d2legacy.item.identity" },
    })
    for _, entity in ipairs(entities) do
        local item = ecs.get(entity, "d2legacy.item.identity")
        local belongs_to_owner = item
            and item:get("owner"):id() == owner_entity:id()
        if belongs_to_owner and item:get("id") == item_id then
            return entity, item
        end
    end
    return nil, nil
end

return M
