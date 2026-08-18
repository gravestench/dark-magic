-- Materialize every admitted golem through one record-derived transaction.
-- Point and item targets are revalidated at the effect tick; replacement and
-- item consumption commit only after the new ordinary monster is complete.

local ecs = require("engine.ecs/v1")
local target = require("d2legacy.gameplay.golem_target")
local limits = require("d2legacy.owned_units.limits")
local monster_data = require("d2legacy.data.monster")
local progression = require("d2legacy.policy.skill_progression")
local summon_policy = require("d2legacy.policy.summon")
local spawn_monster = require("d2legacy.commands.spawn_monster")
local M = {}

local function learned_levels(entities, owner)
    local effective, hard = {}, {}
    for _, entity in ipairs(entities) do
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned and learned:get("owner"):id() == owner:id() then
            local id, level = learned:get("skill_id"), learned:get("level")
            local hard_component = ecs.get(entity, "d2legacy.player.skill_hard_level")
            local purchased = hard_component and hard_component:get("level") or level
            effective[id] = level
            hard[id] = purchased
        end
    end
    return effective, hard
end

local function candidates(entities, owner)
    local result = {}
    for _, entity in ipairs(entities) do
        local relation = ecs.get(entity, "d2legacy.owned_unit")
        if relation and relation:get("owner"):id() == owner:id() then
            result[#result + 1] = {
                entity = entity,
                category = relation:get("category"),
                group = relation:get("group"),
                active = relation:get("active"),
                created_tick = relation:get("created_tick"),
            }
        end
    end
    return result
end

local function linear(base, per_level, level)
    return progression.linear(base, per_level, math.max(level, 1))
end

local function golem_stats(definition, level, effective, hard, monster)
    local mastery = effective[definition.mastery_skill_id] or 0
    local mastery_life = mastery > 0
            and linear(definition.mastery_life_base, definition.mastery_life_per_level, mastery)
        or 0
    local life_percent = mastery_life
        + math.max(level - 1, 0) * definition.life_percent_per_level
        + (hard[definition.blood_skill_id] or 0) * definition.blood_life_per_hard_level
    monster.life_min = summon_policy.whole_life_raw(monster.life_min, life_percent, 0)
    monster.life_max = summon_policy.whole_life_raw(monster.life_max, life_percent, 0)

    local damage_percent = math.max(level - 1, 0) * definition.damage_percent_per_level
        + (hard[definition.fire_skill_id] or 0) * definition.fire_damage_per_hard_level
    monster.physical_min = summon_policy.damage_raw(monster.physical_min, 0, damage_percent)
    monster.physical_max = summon_policy.damage_raw(monster.physical_max, 0, damage_percent)
    local attack = mastery > 0 and linear(definition.mastery_attack_base, definition.mastery_attack_per_level, mastery)
        or 0
    monster.attack_rating = monster.attack_rating
        + attack
        + (hard[definition.clay_skill_id] or 0) * definition.clay_attack_per_hard_level
    local defense_levels = definition.defense_level_source == "effective_self" and level
        or hard[definition.iron_skill_id]
        or 0
    monster.defense = monster.defense + defense_levels * definition.iron_defense_per_hard_level
    if mastery > 0 then
        monster.velocity = math.floor(
            monster.velocity
                * (100 + progression.diminishing(
                    definition.mastery_velocity_minimum,
                    definition.mastery_velocity_maximum,
                    mastery
                ))
                / 100
        )
    end
    monster.evil = false
end

local function defense(definition, level, effective)
    local resist_level = effective[definition.summon_resist_skill_id] or 0
    local resist = resist_level > 0
            and progression.diminishing(
                definition.summon_resist_minimum,
                definition.summon_resist_maximum,
                resist_level
            )
        or 0
    local absorb = definition.fire_absorb_maximum > 0
            and progression.diminishing(definition.fire_absorb_minimum, definition.fire_absorb_maximum, level)
        or 0
    return {
        base_physical_resist = 0,
        base_fire_resist = resist,
        base_cold_resist = resist,
        base_lightning_resist = resist,
        physical_resist = 0,
        fire_resist = resist,
        cold_resist = resist,
        lightning_resist = resist,
        max_fire_resist = 75,
        max_cold_resist = 75,
        max_lightning_resist = 75,
        physical_reduction_raw = 0,
        fire_absorb_percent = absorb,
    }
end

local function intrinsic(definition, level)
    local thorns = definition.thorns_base > 0 and linear(definition.thorns_base, definition.thorns_per_level, level)
        or 0
    local fire_minimum, fire_maximum = 0, 0
    if definition.fire_maximum > 0 then
        fire_minimum = progression.banded(definition.fire_minimum, definition.fire_minimum_bands, level)
        fire_maximum = progression.banded(definition.fire_maximum, definition.fire_maximum_bands, level)
    end
    return thorns, fire_minimum, fire_maximum
end

local function components(context, caster, cast, definition, resolved, effective, hard)
    local identity = assert(ecs.get(caster, "d2legacy.player.identity"))
    local location = assert(ecs.get(caster, "d2legacy.world.location"))
    local progress = assert(ecs.get(caster, "d2legacy.player.progress"))
    local level = cast:get("skill_level")
    local monster =
        monster_data.summon(definition.summon_monster_id, summon_policy.pet_level(level, progress:get("level")))
    golem_stats(definition, level, effective, hard, monster)
    -- Iron Golem's consumed base item participates before its generated
    -- property sources. A weapon contributes its listed damage and armor its
    -- listed defense; both remain ordinary monster combat values afterward.
    monster.physical_min = monster.physical_min + (resolved.weapon_minimum_raw or 0)
    monster.physical_max = monster.physical_max + (resolved.weapon_maximum_raw or 0)
    monster.defense = monster.defense + (resolved.item_defense or 0)
    local summon_id = "summon:" .. identity:get("player") .. ":" .. definition.skill_id .. ":" .. context.tick
    local result = spawn_monster.components({
        tick = context.tick,
        payload = {
            spawn_id = summon_id,
            x = resolved.x,
            y = resolved.y,
            act = location:get("act"),
            level_id = location:get("level_id"),
            seed = summon_id,
            definition = monster,
        },
    }, 1)
    result["d2legacy.combat.defense"] = defense(definition, level, effective)
    result["d2legacy.owned_unit"] = {
        owner = caster,
        owner_id = "player:" .. identity:get("player"),
        ultimate_owner_id = "player:" .. identity:get("player"),
        category = definition.category,
        group = definition.category_group,
        limit = 1,
        replacement = "replace_oldest",
        created_tick = context.tick,
        expires_tick = 0,
        durable_id = "",
        durable = false,
        unsummon = definition.unsummon,
        warp_with_owner = definition.warp_with_owner,
        range_limited = definition.range_limited,
        active = true,
        survives_owner_death = false,
    }
    result["d2legacy.world.selectable"].owner = identity:get("player")
    local thorns, fire_minimum, fire_maximum = intrinsic(definition, level)
    result["d2legacy.summon.intrinsic_stats"] = {
        source_id = summon_id,
        thorns_percent = thorns,
        fire_minimum_raw = fire_minimum,
        fire_maximum_raw = fire_maximum,
        sources_materialized = false,
    }
    local slow = definition.slow_maximum > 0
            and progression.diminishing(definition.slow_minimum, definition.slow_maximum, level)
        or 0
    local steal = definition.life_steal_maximum > 0
            and progression.diminishing(definition.life_steal_minimum, definition.life_steal_maximum, level)
        or 0
    result["d2legacy.combat.reactive_effect"] = {
        owner = caster,
        source_id = summon_id,
        slow_percent_on_melee_hit = slow,
        slow_duration_ticks = definition.slow_duration_ticks,
        life_steal_percent = steal,
        stolen_life_owner_percent = definition.owner_share_percent,
        owner_healing_share_percent = definition.owner_healing_share_percent,
        owner_health_observed = ecs.get(caster, "d2legacy.player.vitals"):get("health"),
    }
    if definition.granted_skill_name ~= "" then
        local granted_level = math.min(
            linear(definition.granted_skill_base, definition.granted_skill_per_level, level),
            definition.granted_skill_cap
        )
        result["d2legacy.monster.granted_skill"] = {
            skill = definition.granted_skill_name,
            level = granted_level,
        }
        result["d2legacy.combat.periodic_damage"] = {
            source_id = summon_id .. ":" .. definition.granted_skill_name,
            channel = "fire",
            target_policy = "hostile_monsters",
            radius = linear(
                definition.granted_aura_radius_base,
                definition.granted_aura_radius_per_level,
                granted_level
            ),
            minimum_raw = progression.banded(
                definition.granted_aura_minimum,
                definition.granted_aura_minimum_bands,
                granted_level
            ),
            maximum_raw = progression.banded(
                definition.granted_aura_maximum,
                definition.granted_aura_maximum_bands,
                granted_level
            ),
            period_ticks = definition.granted_aura_period_ticks,
            next_tick = context.tick + definition.granted_aura_period_ticks,
        }
    end
    if resolved.entity then
        result["d2legacy.summon.item_provenance"] = {
            source_item_entity_id = resolved.entity:id(),
            item_id = resolved.item_id,
            item_code = resolved.item_code,
            item_types = resolved.item_types,
            identified = resolved.identified,
            modifiers_transferred = false,
            resolved_weapon_minimum_raw = resolved.weapon_minimum_raw,
            resolved_weapon_maximum_raw = resolved.weapon_maximum_raw,
            resolved_defense = resolved.item_defense,
        }
    end
    return result, summon_id
end

local function emit(structural, context, caster, cast, definition, outcome, reason, summon_id, target_id)
    local identity = assert(ecs.get(caster, "d2legacy.player.identity"))
    structural:create({
        ["d2legacy.skill.summon_event"] = {
            kind = "golem_summon",
            outcome = outcome,
            reason = reason,
            tick = context.tick,
            player = identity:get("player"),
            skill_id = cast:get("skill_id"),
            skill_level = cast:get("skill_level"),
            target_id = target_id or "",
            corpse_id = "",
            summon_id = summon_id or "",
            category = definition.category,
            limit = 1,
        },
    })
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.skill.golem_summon",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = {
            any = {
                "d2legacy.skill.cast",
                "d2legacy.player.learned_skill",
                "d2legacy.player.skill_hard_level",
                "d2legacy.item.identity",
                "d2legacy.item.placement",
                "d2legacy.item.stat_modifier",
                "d2legacy.owned_unit",
            },
        },
        read = {
            "d2legacy.skill.cast",
            "d2legacy.player.learned_skill",
            "d2legacy.player.skill_hard_level",
            "d2legacy.player.identity",
            "d2legacy.player.progress",
            "d2legacy.player.vitals",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.inactive",
            "d2legacy.item.identity",
            "d2legacy.item.placement",
            "d2legacy.item.melee",
            "d2legacy.item.armor",
            "d2legacy.item.stat_modifier",
            "d2legacy.owned_unit",
        },
        write = {
            "d2legacy.skill.cast",
            "d2legacy.owned_unit",
            "d2legacy.skill.summon_event",
            "d2legacy.monster.identity",
            "d2legacy.monster.stats",
            "d2legacy.combat.melee_profile",
            "d2legacy.combat.knockback_target",
            "d2legacy.monster.appearance",
            "d2legacy.monster.ai",
            "d2legacy.combat.defense",
            "d2legacy.combat.reactive_effect",
            "d2legacy.summon.intrinsic_stats",
            "d2legacy.summon.item_provenance",
            "d2legacy.monster.granted_skill",
            "d2legacy.combat.periodic_damage",
            "d2legacy.world.position",
            "d2legacy.world.velocity",
            "d2legacy.world.facing",
            "d2legacy.world.location",
            "d2legacy.world.collider",
            "d2legacy.world.occupancy",
            "d2legacy.world.selectable",
            "engine.world.velocity_mover",
        },
        update = function(context, entities, structural)
            for _, caster in ipairs(entities) do
                local cast = ecs.get(caster, "d2legacy.skill.cast")
                local definition = cast and definitions[cast:get("skill_id")] or nil
                if definition and not cast:get("effect_emitted") and context.tick >= cast:get("effect_tick") then
                    local resolved, reason = target.resolve(
                        caster,
                        cast:get("target_x"),
                        cast:get("target_y"),
                        cast:get("target_id"),
                        definition,
                        entities
                    )
                    if not resolved then
                        emit(
                            structural,
                            context,
                            caster,
                            cast,
                            definition,
                            "invalidated",
                            reason,
                            "",
                            cast:get("target_id")
                        )
                    else
                        local effective, hard = learned_levels(entities, caster)
                        local values, summon_id =
                            components(context, caster, cast, definition, resolved, effective, hard)
                        local victims = limits.victims(candidates(entities, caster), {
                            id = definition.category,
                            group = definition.category_group,
                            base_max = 1,
                            replacement = "replace_oldest",
                        })
                        for _, victim in ipairs(victims) do
                            structural:destroy(victim.entity)
                        end
                        structural:create(values)
                        if resolved.entity then
                            structural:destroy(resolved.entity)
                        end
                        emit(
                            structural,
                            context,
                            caster,
                            cast,
                            definition,
                            "summoned",
                            "",
                            summon_id,
                            resolved.item_id or ""
                        )
                    end
                    cast:set("effect_emitted", true)
                end
            end
        end,
    })
end

return M
