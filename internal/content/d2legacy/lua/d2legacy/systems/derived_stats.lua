-- Rebuild effective combat defenses from durable base facts and named sources.

local ecs = require("engine.ecs/v1")
local M = {}

local function source_totals(entities, target)
    local totals = {physical_resist=0, fire_resist=0, max_fire_resist=0,
        physical_reduction_raw=0, attack_rating=0, defense=0}
    for _, entity in ipairs(entities) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        if source and source:get("target"):id() == target:id() then
            local stat = source:get("stat")
            if totals[stat] ~= nil then totals[stat] = totals[stat] + source:get("value") end
        end
    end
    return totals
end

function M.register()
    ecs.system({id="d2legacy.combat.derived_stats", phase="pre_simulation",
        query={any={"d2legacy.combat.defense", "d2legacy.player.combat_stats", "d2legacy.stat.source"}},
        read={"d2legacy.combat.defense", "d2legacy.player.combat_stats", "d2legacy.stat.source"},
        write={"d2legacy.combat.defense", "d2legacy.player.combat_stats"},
        update=function(_, entities)
            for _, target in ipairs(entities) do
                local defense = ecs.get(target, "d2legacy.combat.defense")
                local combat = ecs.get(target, "d2legacy.player.combat_stats")
                if defense or combat then
                    local totals=source_totals(entities,target)
                    if combat then
                        combat:set("attack_rating",combat:get("base_attack_rating")+totals.attack_rating)
                        combat:set("defense",combat:get("base_defense")+totals.defense)
                    end
                    if defense then
                        defense:set("physical_resist", defense:get("base_physical_resist") + totals.physical_resist)
                        defense:set("fire_resist", defense:get("base_fire_resist") + totals.fire_resist)
                        defense:set("max_fire_resist", 75 + totals.max_fire_resist)
                        defense:set("physical_reduction_raw", totals.physical_reduction_raw)
                    end
                end
            end
        end})
end

return M
