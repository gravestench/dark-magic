-- Materialize one death's already-rolled loot without rolling it again.

local ecs = require("engine.ecs/v1")
local bases = require("d2legacy.loot.items")
local generation = require("d2legacy.loot.generate")
local item_entity = require("d2legacy.items.entity")
local M = {}

local function source_layout(entities)
    local sources = {}
    for _, entity in ipairs(entities) do
        if ecs.get(entity, "d2legacy.items.world_source") then
            sources[#sources + 1] = entity
        end
    end
    assert(#sources == 1, "world item source layout unavailable")
    local source = sources[1]
    assert(ecs.get(source, "d2legacy.items.layout"):get("owner") == "", "world source has a player owner")
    return source
end

function M.spawn(context, entities, structural, monster, drops)
    local layout = source_layout(entities)
    local position = ecs.get(monster, "d2legacy.world.position")
    local location = ecs.get(monster, "d2legacy.world.location")
    local identity = ecs.get(monster, "d2legacy.monster.identity")
    if not layout or not position or not location then
        return
    end

    for index, drop in ipairs(drops) do
        local base = bases.base(drop.code)
        -- shortcut: currency amounts and specialized quality effects remain in
        -- the rolled facts; extend their policies before claiming full loot fidelity.
        if base and drop.code ~= "gld" and base.width > 0 and base.height > 0 then
            local item = {}
            for key, value in pairs(base) do
                item[key] = value
            end
            item.id = "loot:" .. context.tick .. ":" .. monster:id() .. ":" .. index
            item.origin = generation.encode({ "monster", identity:get("spawn_id"), context.tick, index })
            item.container = "world"
            item.x, item.y = math.floor(position:get("x")), math.floor(position:get("y"))
            item.act, item.level_id = location:get("act"), location:get("level_id")
            item.identified = drop.quality == "normal" or drop.quality == "superior" or drop.quality == "low"
            item.generation = {
                quality = drop.quality,
                level = drop.level,
                facts = generation.encode({ drop }),
            }
            structural:create(item_entity.components(layout, item))
        end
    end
end

return M
