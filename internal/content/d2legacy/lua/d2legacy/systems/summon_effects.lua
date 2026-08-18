-- Bridge durable summon effect facts into the ordinary stat, state, healing,
-- and damage vocabularies. Consumers dispatch on data shape, never skill IDs
-- or pet names, so later summoned units can reuse the same transactions.

local ecs = require("engine.ecs/v1")
local area_targets = require("d2legacy.gameplay.area_targets")
local damage_policy = require("d2legacy.policy.damage")
local damage_bundle = require("d2legacy.policy.damage_bundle")
local progression = require("d2legacy.policy.skill_progression")
local M = {}

local function selectable_index(entities)
    local result = {}
    for _, entity in ipairs(entities) do
        local selected = ecs.get(entity, "d2legacy.world.selectable")
        if selected then
            result[selected:get("id")] = entity
        end
    end
    return result
end

local function source(structural, target, source_id, stat, operation, value, order)
    if not target or value == 0 then
        return
    end
    structural:create({
        ["d2legacy.stat.source"] = {
            target = target,
            source_id = source_id,
            owner_source_id = source_id,
            stat = stat,
            operation = operation,
            value = value,
            order = order,
        },
    })
end

local function heal_monster(entity, raw)
    local stats = entity and ecs.get(entity, "d2legacy.monster.stats")
    if not stats or raw <= 0 then
        return 0, math.max(raw, 0)
    end
    local before = stats:get("health")
    local after = math.min(stats:get("max_health"), before + raw)
    stats:set("health", after)
    return after - before, raw - (after - before)
end

local function heal_owner(entity, raw)
    local vitals = entity and ecs.get(entity, "d2legacy.player.vitals")
    if not vitals or raw <= 0 then
        return 0
    end
    local points = math.floor(raw / 256)
    local before = vitals:get("health")
    local after = math.min(vitals:get("max_health"), before + points)
    vitals:set("health", after)
    return (after - before) * 256
end

local function observe_owner_healing(summon, effect)
    if effect:get("owner_healing_share_percent") <= 0 then
        return
    end
    local owner = effect:get("owner")
    local vitals = ecs.get(owner, "d2legacy.player.vitals")
    if not vitals then
        return
    end
    local health = vitals:get("health")
    local previous = effect:get("owner_health_observed")
    if health > previous then
        heal_monster(summon, math.floor((health - previous) * 256 * effect:get("owner_healing_share_percent") / 100))
    end
    effect:set("owner_health_observed", health)
end

local function apply_slow(structural, target, effect)
    local percentage = effect:get("slow_percent_on_melee_hit")
    if percentage <= 0 then
        return
    end
    -- Expansion caps Slows Target at 50% for players and boss-class units,
    -- and at 90% for ordinary monsters.
    local exceptional = ecs.get(target, "d2legacy.player.identity")
        or ecs.get(target, "d2legacy.monster.boss")
        or ecs.get(target, "d2legacy.monster.prime_evil")
    local cap = exceptional and 50 or 90
    percentage = math.min(percentage, cap)
    local source_id = effect:get("source_id") .. ":slow"
    structural:create({
        ["d2legacy.state.request"] = {
            operation = "apply",
            target = target,
            state_id = "slowed",
            source_id = source_id,
            duration = effect:get("slow_duration_ticks"),
            policy = "refresh_same_source",
            stat = "velocitypercent",
            stat_operation = "percent",
            stat_value = -percentage,
            stat_order = 10,
            exclusive_group = "slows-target",
            replacement_priority = percentage,
            reject_lower_priority = true,
        },
    })
    structural:create({
        ["d2legacy.state.stat_request"] = {
            target = target,
            owner_source_id = source_id,
            source_id = source_id .. ":attack",
            stat = "attackrate",
            operation = "percent",
            value = -percentage,
            order = 11,
        },
    })
end

local function blood_attack(attacker, event, effect)
    local total = math.floor(event:get("damage_raw") * effect:get("life_steal_percent") / 100)
    local owner_part = math.floor(total * effect:get("stolen_life_owner_percent") / 100)
    local summon_part = total - owner_part
    local _, overflow = heal_monster(attacker, summon_part)
    local owner = effect:get("owner")
    heal_owner(owner, owner_part + overflow)
    local vitals = ecs.get(owner, "d2legacy.player.vitals")
    if vitals then
        effect:set("owner_health_observed", vitals:get("health"))
    end
end

local function emit_periodic_damage(structural, context, emitter, target, target_id, periodic)
    local rolled = damage_policy.roll(periodic:get("minimum_raw"), periodic:get("maximum_raw"))
    local result = damage_policy.resolve(target, damage_bundle.single(periodic:get("channel"), rolled), ecs)
    local selected = assert(ecs.get(emitter, "d2legacy.world.selectable"))
    structural:create({
        ["d2legacy.combat.event"] = {
            kind = result.lethal and "unit_died" or "damage_applied",
            tick = context.tick,
            attacker_id = selected:get("id"),
            target_id = target_id,
            source_kind = "periodic_damage",
            damage_channel = result.channel,
            rolled_damage_raw = result.rolled_damage_raw,
            damage_raw = result.damage_raw,
            remaining_health_raw = result.remaining_health_raw,
        },
        ["d2legacy.combat.damage_bundle"] = damage_bundle.stage_component(result.rolled, result.mitigated),
    })
end

function M.register()
    ecs.system({
        id = "d2legacy.summon.materialize_intrinsic_stats",
        phase = "effects",
        query = {
            any = {
                "d2legacy.summon.intrinsic_stats",
                "d2legacy.summon.item_provenance",
                "d2legacy.item.stat_modifier",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.summon.intrinsic_stats",
            "d2legacy.summon.item_provenance",
            "d2legacy.item.stat_modifier",
            "d2legacy.stat.source",
        },
        write = { "d2legacy.summon.intrinsic_stats", "d2legacy.summon.item_provenance", "d2legacy.stat.source" },
        update = function(_, entities, structural)
            for _, summon in ipairs(entities) do
                local stats = ecs.get(summon, "d2legacy.summon.intrinsic_stats")
                if stats and not stats:get("sources_materialized") then
                    local id = "summon:" .. stats:get("source_id")
                    source(structural, summon, id .. ":thorns", "thorns_percent", "add", stats:get("thorns_percent"), 1)
                    source(structural, summon, id .. ":fire-min", "firemindam", "add", stats:get("fire_minimum_raw"), 2)
                    source(structural, summon, id .. ":fire-max", "firemaxdam", "add", stats:get("fire_maximum_raw"), 3)
                    stats:set("sources_materialized", true)
                end
                local provenance = ecs.get(summon, "d2legacy.summon.item_provenance")
                if provenance and not provenance:get("modifiers_transferred") then
                    local source_item = provenance:get("source_item_entity_id")
                    for _, modifier_entity in ipairs(entities) do
                        local modifier = ecs.get(modifier_entity, "d2legacy.item.stat_modifier")
                        if modifier and modifier:get("item"):id() == source_item then
                            -- Local percentages were folded into the consumed
                            -- base item's damage or defense during target
                            -- validation. All other properties remain ordinary
                            -- stat sources on the resulting monster.
                            if modifier:get("operation") ~= "local_percent" then
                                source(
                                    structural,
                                    summon,
                                    "iron-item:" .. provenance:get("item_id") .. ":" .. modifier:get("source_id"),
                                    modifier:get("stat"),
                                    modifier:get("operation"),
                                    modifier:get("value"),
                                    100 + modifier:get("order")
                                )
                            end
                            structural:destroy(modifier_entity)
                        end
                    end
                    provenance:set("modifiers_transferred", true)
                end
            end
        end,
    })

    ecs.system({
        id = "d2legacy.summon.react_to_combat",
        phase = "effects",
        query = {
            any = {
                "d2legacy.combat.melee_event",
                "d2legacy.combat.reactive_effect",
                "d2legacy.world.selectable",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.melee_event",
            "d2legacy.combat.reactive_effect",
            "d2legacy.combat.summon_reaction_observed",
            "d2legacy.world.selectable",
            "d2legacy.monster.stats",
            "d2legacy.monster.boss",
            "d2legacy.monster.prime_evil",
            "d2legacy.player.identity",
            "d2legacy.player.vitals",
        },
        write = {
            "d2legacy.combat.summon_reaction_observed",
            "d2legacy.combat.reactive_effect",
            "d2legacy.monster.stats",
            "d2legacy.player.vitals",
            "d2legacy.state.request",
            "d2legacy.state.stat_request",
        },
        update = function(_, entities, structural)
            local by_id = selectable_index(entities)
            for _, summon in ipairs(entities) do
                local effect = ecs.get(summon, "d2legacy.combat.reactive_effect")
                if effect then
                    observe_owner_healing(summon, effect)
                end
            end
            for _, event_entity in ipairs(entities) do
                local event = ecs.get(event_entity, "d2legacy.combat.melee_event")
                if
                    event
                    and event:get("hit")
                    and event:get("outcome") == "hit"
                    and not ecs.get(event_entity, "d2legacy.combat.summon_reaction_observed")
                then
                    structural:set(event_entity, "d2legacy.combat.summon_reaction_observed", {})
                    local attacker, defender = by_id[event:get("attacker_id")], by_id[event:get("target_id")]
                    local attacking_effect = attacker and ecs.get(attacker, "d2legacy.combat.reactive_effect")
                    local defending_effect = defender and ecs.get(defender, "d2legacy.combat.reactive_effect")
                    if attacking_effect then
                        apply_slow(structural, defender, attacking_effect)
                        if attacking_effect:get("life_steal_percent") > 0 then
                            blood_attack(attacker, event, attacking_effect)
                        end
                    end
                    if defending_effect then
                        apply_slow(structural, attacker, defending_effect)
                    end
                end
            end
        end,
    })

    ecs.system({
        id = "d2legacy.combat.periodic_area_damage",
        phase = "effects",
        query = {
            any = { "d2legacy.combat.periodic_damage", "d2legacy.world.selectable" },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.periodic_damage",
            "d2legacy.world.selectable",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.collider",
            "d2legacy.monster.stats",
            "d2legacy.player.vitals",
            "d2legacy.combat.defense",
        },
        write = {
            "d2legacy.combat.periodic_damage",
            "d2legacy.monster.stats",
            "d2legacy.player.vitals",
            "d2legacy.combat.event",
            "d2legacy.combat.damage_bundle",
        },
        update = function(context, entities, structural)
            for _, emitter in ipairs(entities) do
                local periodic = ecs.get(emitter, "d2legacy.combat.periodic_damage")
                if periodic and context.tick >= periodic:get("next_tick") then
                    periodic:set(
                        "next_tick",
                        progression.next_periodic_tick(context.tick, periodic:get("period_ticks"))
                    )
                    assert(
                        periodic:get("target_policy") == "hostile_monsters",
                        "unsupported periodic damage target policy"
                    )
                    local position = assert(ecs.get(emitter, "d2legacy.world.position"))
                    for _, target in
                        ipairs(
                            area_targets.hostile_monsters(
                                entities,
                                emitter,
                                position:get("x"),
                                position:get("y"),
                                periodic:get("radius")
                            )
                        )
                    do
                        emit_periodic_damage(structural, context, emitter, target.entity, target.id, periodic)
                    end
                end
            end
        end,
    })
end

return M
