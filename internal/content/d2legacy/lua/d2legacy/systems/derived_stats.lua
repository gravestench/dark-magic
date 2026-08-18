-- Rebuild effective combat defenses from durable base facts and named sources.

local ecs = require("engine.ecs/v1")
local stat_sources = require("d2legacy.gameplay.stat_sources")
local M = {}

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
                "d2legacy.player.resource_stats",
                "d2legacy.stat.source",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.defense",
            "d2legacy.player.combat_stats",
            "d2legacy.combat.action_rate",
            "d2legacy.player.movement_stats",
            "d2legacy.player.resource_stats",
            "d2legacy.stat.source",
        },
        write = {
            "d2legacy.combat.defense",
            "d2legacy.player.combat_stats",
            "d2legacy.combat.action_rate",
            "d2legacy.player.movement_stats",
            "d2legacy.player.resource_stats",
        },
        update = function(_, entities)
            for _, target in ipairs(entities) do
                local defense = ecs.get(target, "d2legacy.combat.defense")
                local combat = ecs.get(target, "d2legacy.player.combat_stats")
                local action_rate = ecs.get(target, "d2legacy.combat.action_rate")
                local movement = ecs.get(target, "d2legacy.player.movement_stats")
                local resources = ecs.get(target, "d2legacy.player.resource_stats")
                if defense or combat or action_rate or movement or resources then
                    if combat then
                        local attack_rating =
                            stat_sources.resolve(entities, target, "attack_rating", combat:get("base_attack_rating"))
                        combat:set(
                            "attack_rating",
                            stat_sources.resolve(entities, target, "item_tohit_percent", attack_rating)
                        )
                        combat:set(
                            "defense",
                            stat_sources.resolve(entities, target, "defense", combat:get("base_defense"))
                        )
                    end
                    if defense then
                        defense:set(
                            "physical_resist",
                            stat_sources.resolve(
                                entities,
                                target,
                                "physical_resist",
                                defense:get("base_physical_resist")
                            )
                        )
                        defense:set(
                            "fire_resist",
                            stat_sources.resolve(entities, target, "fire_resist", defense:get("base_fire_resist"))
                        )
                        defense:set(
                            "cold_resist",
                            stat_sources.resolve(entities, target, "cold_resist", defense:get("base_cold_resist"))
                        )
                        defense:set(
                            "lightning_resist",
                            stat_sources.resolve(
                                entities,
                                target,
                                "lightning_resist",
                                defense:get("base_lightning_resist")
                            )
                        )
                        defense:set("max_fire_resist", stat_sources.resolve(entities, target, "max_fire_resist", 75))
                        defense:set("max_cold_resist", stat_sources.resolve(entities, target, "max_cold_resist", 75))
                        defense:set(
                            "max_lightning_resist",
                            stat_sources.resolve(entities, target, "max_lightning_resist", 75)
                        )
                        defense:set(
                            "physical_reduction_raw",
                            stat_sources.resolve(entities, target, "physical_reduction_raw", 0)
                        )
                    end
                    if action_rate then
                        action_rate:set(
                            "attack_rate",
                            stat_sources.resolve(entities, target, "attackrate", action_rate:get("base_attack_rate"))
                        )
                        action_rate:set(
                            "item_fasterattackrate",
                            stat_sources.resolve(entities, target, "item_fasterattackrate", 0)
                        )
                    end
                    if movement then
                        movement:set("velocitypercent", stat_sources.resolve(entities, target, "velocitypercent", 0))
                        movement:set(
                            "item_fastermovevelocity",
                            stat_sources.resolve(entities, target, "item_fastermovevelocity", 0)
                        )
                        movement:set(
                            "staminarecoverybonus",
                            stat_sources.resolve(entities, target, "staminarecoverybonus", 0)
                        )
                        movement:set(
                            "item_staminadrainpct",
                            stat_sources.resolve(entities, target, "item_staminadrainpct", 0)
                        )
                        movement:set("armor_run_drain", stat_sources.resolve(entities, target, "armor_run_drain", 0))
                    end
                    if resources then
                        resources:set(
                            "manarecoverybonus",
                            stat_sources.resolve(entities, target, "manarecoverybonus", 0)
                        )
                        resources:set("manarecovery", stat_sources.resolve(entities, target, "manarecovery", 0))
                    end
                end
            end
        end,
    })
end

return M
