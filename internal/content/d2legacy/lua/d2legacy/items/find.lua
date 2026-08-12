-- Deterministic item/layout lookup helpers shared by small policy modules.

local ecs=require("engine.ecs/v1")
local M={}
function M.layout(owner,entities)
    entities=entities or ecs.query({all={"d2legacy.items.layout"}})
    for _,entity in ipairs(entities) do
        local layout=ecs.get(entity,"d2legacy.items.layout")
        if layout and layout:get("owner")==owner then return entity,layout end
    end
end
function M.item(owner_entity,id,entities)
    entities=entities or ecs.query({all={"d2legacy.item.identity"}})
    for _,entity in ipairs(entities) do
        local item=ecs.get(entity,"d2legacy.item.identity")
        if item and item:get("owner"):id()==owner_entity:id() and item:get("id")==id then return entity,item end
    end
end
return M
