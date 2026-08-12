-- Follow an explicit ownership relationship for combat credit.
--
-- The immediate source remains useful for effects such as "your skeleton made
-- this kill".  The ultimate owner receives player-facing credit and rewards.

local ecs = require("engine.ecs/v1")
local M = {}

local function selectable_entity(entities, wanted)
    for _, entity in ipairs(entities) do
        local selectable = ecs.get(entity, "d2legacy.world.selectable")
        if selectable and selectable:get("id") == wanted then return entity end
    end
    return nil
end

function M.resolve(entities, source_id)
    local source = selectable_entity(entities, source_id)
    local relation = source and ecs.get(source, "d2legacy.owned_unit")
    if relation and relation:get("active") then
        return {
            source_id = source_id,
            immediate_owner_id = relation:get("owner_id"),
            ultimate_owner_id = relation:get("ultimate_owner_id"),
        }
    end
    return {
        source_id = source_id,
        immediate_owner_id = source_id,
        ultimate_owner_id = source_id,
    }
end

return M
