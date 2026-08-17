-- Rebuild effective combat defenses from durable base facts and named sources.

local ecs = require("engine.ecs/v1")
local resolution = require("d2legacy.policy.stat_resolution")
local M = {}

local function sources_by_stat(entities, target)
    local result = {}
    for _, entity in ipairs(entities) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        if source and source:get("target"):id() == target:id() then
            local stat = source:get("stat")
            result[stat] = result[stat] or {}
            result[stat][#result[stat] + 1] = {
                id = source:get("source_id"),
                operation = source:get("operation") ~= "" and source:get("operation") or "add",
                value = source:get("value"),
                order = source:get("order"),
            }
        end
    end
    return result
end

local function resolve(base, sources, stat)
    return resolution.resolve(base, sources[stat] or {})
end

function M.register()
    ecs.system({
        id = "d2legacy.combat.derived_stats",
        phase = "pre_simulation",
        query = {
            any = {
                "d2legacy.combat.defense",
                "d2legacy.player.combat_stats",
                "d2legacy.combat.action_rate",
                "d2legacy.player.movement_stats",
                "d2legacy.stat.source",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.defense",
            "d2legacy.player.combat_stats",
            "d2legacy.combat.action_rate",
            "d2legacy.player.movement_stats",
            "d2legacy.stat.source",
        },
        write = {
            "d2legacy.combat.defense",
            "d2legacy.player.combat_stats",
            "d2legacy.combat.action_rate",
            "d2legacy.player.movement_stats",
        },
        update = function(_, entities)
            for _, target in ipairs(entities) do
                local defense = ecs.get(target, "d2legacy.combat.defense")
                local combat = ecs.get(target, "d2legacy.player.combat_stats")
                local action_rate = ecs.get(target, "d2legacy.combat.action_rate")
                local movement = ecs.get(target, "d2legacy.player.movement_stats")
                if defense or combat or action_rate or movement then
                    local sources = sources_by_stat(entities, target)
                    if combat then
                        combat:set("attack_rating", resolve(combat:get("base_attack_rating"), sources, "attack_rating"))
                        combat:set("defense", resolve(combat:get("base_defense"), sources, "defense"))
                    end
                    if defense then
                        defense:set(
                            "physical_resist",
                            resolve(defense:get("base_physical_resist"), sources, "physical_resist")
                        )
                        defense:set("fire_resist", resolve(defense:get("base_fire_resist"), sources, "fire_resist"))
                        defense:set("max_fire_resist", resolve(75, sources, "max_fire_resist"))
                        defense:set("physical_reduction_raw", resolve(0, sources, "physical_reduction_raw"))
                    end
                    if action_rate then
                        action_rate:set(
                            "attack_rate",
                            resolve(action_rate:get("base_attack_rate"), sources, "attackrate")
                        )
                        action_rate:set("item_fasterattackrate", resolve(0, sources, "item_fasterattackrate"))
                    end
                    if movement then
                        movement:set("velocitypercent", resolve(0, sources, "velocitypercent"))
                        movement:set("item_fastermovevelocity", resolve(0, sources, "item_fastermovevelocity"))
                        movement:set("staminarecoverybonus", resolve(0, sources, "staminarecoverybonus"))
                        movement:set("item_staminadrainpct", resolve(0, sources, "item_staminadrainpct"))
                        movement:set("armor_run_drain", resolve(0, sources, "armor_run_drain"))
                    end
                end
            end
        end,
    })
end

return M
