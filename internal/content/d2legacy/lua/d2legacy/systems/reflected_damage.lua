-- Consume factual successful melee damage and apply record-authored reflection.
--
-- The melee resolver owns hit/block/miss and the initial damage transaction.
-- This system observes that immutable result, so ranged attacks and unsuccessful
-- melee attempts cannot trigger reflection by resemblance. An empty ECS marker
-- makes the consumer idempotent without coupling it to other result consumers.

local ecs = require("engine.ecs/v1")
local damage_policy = require("d2legacy.policy.damage")
local damage_bundle = require("d2legacy.policy.damage_bundle")
local stat_sources = require("d2legacy.gameplay.stat_sources")
local M = {}

local function selectable_index(entities)
    local result = {}
    for _, entity in ipairs(entities) do
        local selectable = ecs.get(entity, "d2legacy.world.selectable")
        if selectable then
            result[selectable:get("id")] = entity
        end
    end
    return result
end

local function reflected_physical_raw(melee, result, stages, percentage)
    if not melee:get("hit") or melee:get("outcome") ~= "hit" or result:get("damage_raw") <= 0 then
        return 0
    end
    local committed_physical = math.min(stages:get("physical_mitigated_raw"), result:get("damage_raw"))
    return math.floor(committed_physical * percentage / 100)
end

function M.register()
    ecs.system({
        id = "d2legacy.combat.reflect_melee",
        phase = "effects",
        query = {
            any = {
                "d2legacy.combat.melee_event",
                "d2legacy.world.selectable",
                "d2legacy.stat.source",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.melee_event",
            "d2legacy.combat.event",
            "d2legacy.combat.damage_bundle",
            "d2legacy.combat.reflection_observed",
            "d2legacy.world.selectable",
            "d2legacy.player.identity",
            "d2legacy.player.vitals",
            "d2legacy.monster.stats",
            "d2legacy.combat.defense",
            "d2legacy.stat.source",
        },
        write = {
            "d2legacy.combat.reflection_observed",
            "d2legacy.combat.event",
            "d2legacy.combat.damage_bundle",
            "d2legacy.player.vitals",
            "d2legacy.monster.stats",
        },
        update = function(context, entities, structural)
            local by_id = selectable_index(entities)
            for _, event_entity in ipairs(entities) do
                local melee = ecs.get(event_entity, "d2legacy.combat.melee_event")
                if melee and not ecs.get(event_entity, "d2legacy.combat.reflection_observed") then
                    structural:set(event_entity, "d2legacy.combat.reflection_observed", {})
                    local stages = ecs.get(event_entity, "d2legacy.combat.damage_bundle")
                    local result = ecs.get(event_entity, "d2legacy.combat.event")
                    local attacker = by_id[melee:get("attacker_id")]
                    local defender = by_id[melee:get("target_id")]
                    -- PvP has a distinct one-eighth rule and remains outside the
                    -- current hostility foundation. Do not apply PvM behavior to
                    -- a player attacker while that contract is unimplemented.
                    if
                        attacker
                        and defender
                        and not ecs.get(attacker, "d2legacy.player.identity")
                        and stages
                        and result
                    then
                        local percentage = stat_sources.resolve(entities, defender, "thorns_percent", 0)
                        local reflected_raw = reflected_physical_raw(melee, result, stages, percentage)
                        if reflected_raw > 0 then
                            local reflected =
                                damage_policy.resolve(attacker, damage_bundle.single("physical", reflected_raw), ecs)
                            structural:create({
                                ["d2legacy.combat.event"] = {
                                    kind = reflected.lethal and "unit_died" or "damage_applied",
                                    tick = context.tick,
                                    attacker_id = melee:get("target_id"),
                                    target_id = melee:get("attacker_id"),
                                    source_kind = "melee_reflection",
                                    damage_channel = reflected.channel,
                                    rolled_damage_raw = reflected.rolled_damage_raw,
                                    damage_raw = reflected.damage_raw,
                                    remaining_health_raw = reflected.remaining_health_raw,
                                },
                                ["d2legacy.combat.damage_bundle"] = damage_bundle.stage_component(
                                    reflected.rolled,
                                    reflected.mitigated
                                ),
                            })
                        end
                    end
                end
            end
        end,
    })
end

return M
