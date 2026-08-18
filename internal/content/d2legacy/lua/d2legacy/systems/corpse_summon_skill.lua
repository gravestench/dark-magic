-- Consume one revalidated corpse and materialize an ordinary friendly monster
-- carrying generic ownership, AI, combat, and resistance components.

local ecs = require("engine.ecs/v1")
local corpse_target = require("d2legacy.gameplay.corpse_target")
local limits = require("d2legacy.owned_units.limits")
local monster_data = require("d2legacy.data.monster")
local progression = require("d2legacy.policy.skill_progression")
local summon_policy = require("d2legacy.policy.summon")
local spawn_monster = require("d2legacy.commands.spawn_monster")
local M = {}

local function learned_levels(entities, owner)
    local result = {}
    for _, entity in ipairs(entities) do
        local learned = ecs.get(entity, "d2legacy.player.learned_skill")
        if learned and learned:get("owner"):id() == owner:id() then
            result[learned:get("skill_id")] = learned:get("level")
        end
    end
    return result
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

local function materialized_monster(caster, corpse, definition, level)
    local progress = assert(ecs.get(caster, "d2legacy.player.progress"))
    if definition.materialization == "fixed_pet" then
        return monster_data.summon(definition.summon_monster_id, summon_policy.pet_level(level, progress:get("level")))
    end
    if definition.materialization == "revived_corpse" then
        local identity = assert(ecs.get(corpse, "d2legacy.monster.identity"))
        local stats = assert(ecs.get(corpse, "d2legacy.monster.stats"))
        local monster =
            monster_data.revive(identity:get("definition_id"), math.min(stats:get("level"), progress:get("level")))
        monster.evil = false
        monster.velocity = math.floor(monster.velocity * (100 + definition.velocity_percent) / 100)
        return monster
    end
    error("unsupported corpse-summon materialization " .. tostring(definition.materialization))
end

local function summon_components(context, caster, corpse, definition, level, levels)
    local identity = assert(ecs.get(caster, "d2legacy.player.identity"))
    local corpse_position = assert(ecs.get(corpse, "d2legacy.world.position"))
    local location = assert(ecs.get(caster, "d2legacy.world.location"))
    local monster = materialized_monster(caster, corpse, definition, level)
    local mastery = levels[definition.mastery_skill_id] or 0
    local combined = level + mastery
    local after_three = math.max(level - 3, 0)
    local life_percent = definition.life_percent_base
        + after_three * definition.life_percent_per_level_after_three
        + mastery * definition.mastery_life_percent_per_level
    local damage_percent = after_three * definition.damage_percent_per_level_after_three
        + mastery * definition.mastery_damage_percent_per_level
    local mastery_life_raw = mastery * definition.mastery_life_flat_per_level * 256
    local mastery_damage_raw = mastery * definition.mastery_damage_flat_per_level * 256
    monster.life_min = summon_policy.whole_life_raw(monster.life_min, life_percent, mastery_life_raw)
    monster.life_max = summon_policy.whole_life_raw(monster.life_max, life_percent, mastery_life_raw)
    local minimum = summon_policy.damage_raw(monster.physical_min, mastery_damage_raw, damage_percent)
    local maximum = summon_policy.damage_raw(monster.physical_max, mastery_damage_raw, damage_percent)
    monster.physical_min, monster.physical_max = minimum, maximum
    monster.attack_rating = monster.attack_rating + combined * definition.attack_rating_per_combined_level
    monster.defense = monster.defense + combined * definition.defense_per_combined_level
    local summon_id = "summon:" .. identity:get("player") .. ":" .. definition.skill_id .. ":" .. context.tick
    local components = spawn_monster.components({
        tick = context.tick,
        payload = {
            spawn_id = summon_id,
            x = corpse_position:get("x"),
            y = corpse_position:get("y"),
            act = location:get("act"),
            level_id = location:get("level_id"),
            seed = summon_id,
            definition = monster,
        },
    }, 1)
    local resist_level = levels[definition.summon_resist.skill_id] or 0
    local resist = 0
    if resist_level > 0 then
        resist =
            progression.diminishing(definition.summon_resist.minimum, definition.summon_resist.maximum, resist_level)
    end
    components["d2legacy.combat.defense"] = {
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
    }
    local limit = summon_policy.pet_limit(definition, level)
    components["d2legacy.owned_unit"] = {
        owner = caster,
        owner_id = "player:" .. identity:get("player"),
        ultimate_owner_id = "player:" .. identity:get("player"),
        category = definition.category,
        group = definition.category_group,
        limit = limit,
        replacement = "replace_oldest",
        created_tick = context.tick,
        expires_tick = context.tick + summon_policy.duration_ticks(definition, level),
        durable_id = "",
        durable = false,
        unsummon = definition.unsummon,
        warp_with_owner = definition.warp_with_owner,
        range_limited = definition.range_limited,
        active = true,
        survives_owner_death = false,
    }
    if definition.duration_base_ticks == 0 and definition.duration_per_level_ticks == 0 then
        components["d2legacy.owned_unit"].expires_tick = 0
    end
    if definition.granted_skill_name ~= "" then
        components["d2legacy.monster.granted_skill"] = {
            skill = definition.granted_skill_name,
            level = summon_policy.granted_skill_level(definition, level, mastery),
        }
    end
    components["d2legacy.world.selectable"].owner = identity:get("player")
    return components, summon_id, limit
end

local function emit(structural, context, caster, cast, definition, outcome, reason, summon_id, corpse_id, limit)
    local identity = assert(ecs.get(caster, "d2legacy.player.identity"))
    structural:create({
        ["d2legacy.skill.summon_event"] = {
            kind = "corpse_summon",
            outcome = outcome,
            reason = reason,
            tick = context.tick,
            player = identity:get("player"),
            skill_id = cast:get("skill_id"),
            skill_level = cast:get("skill_level"),
            target_id = corpse_id,
            corpse_id = corpse_id,
            summon_id = summon_id,
            category = definition.category,
            limit = limit,
        },
    })
end

function M.register(definitions)
    ecs.system({
        id = "d2legacy.skill.corpse_summon",
        phase = "pre_simulation",
        after = { "d2legacy.skill.cast_lifecycle" },
        query = {
            any = {
                "d2legacy.skill.cast",
                "d2legacy.player.learned_skill",
                "d2legacy.world.selectable",
                "d2legacy.owned_unit",
            },
        },
        read = {
            "d2legacy.skill.cast",
            "d2legacy.player.learned_skill",
            "d2legacy.player.identity",
            "d2legacy.player.progress",
            "d2legacy.world.selectable",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.inactive",
            "d2legacy.monster.death",
            "d2legacy.monster.revivable",
            "d2legacy.owned_unit",
        },
        write = {
            "d2legacy.skill.cast",
            "d2legacy.monster.death",
            "d2legacy.owned_unit",
            "d2legacy.skill.summon_event",
            "d2legacy.monster.identity",
            "d2legacy.monster.stats",
            "d2legacy.combat.melee_profile",
            "d2legacy.combat.knockback_target",
            "d2legacy.monster.appearance",
            "d2legacy.monster.ai",
            "d2legacy.combat.defense",
            "d2legacy.monster.corpse_selectable",
            "d2legacy.monster.revivable",
            "d2legacy.monster.granted_skill",
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
                    local corpse, reason = corpse_target.named(caster, cast:get("target_id"), entities, definition)
                    if not corpse then
                        emit(
                            structural,
                            context,
                            caster,
                            cast,
                            definition,
                            "invalidated",
                            reason,
                            "",
                            cast:get("target_id"),
                            0
                        )
                    else
                        local levels = learned_levels(entities, caster)
                        local components, summon_id, limit =
                            summon_components(context, caster, corpse, definition, cast:get("skill_level"), levels)
                        local victims = limits.victims(candidates(entities, caster), {
                            id = definition.category,
                            group = definition.category_group,
                            base_max = limit,
                            replacement = "replace_oldest",
                        })
                        for _, victim in ipairs(victims) do
                            structural:destroy(victim.entity)
                        end
                        structural:destroy(corpse)
                        structural:create(components)
                        emit(
                            structural,
                            context,
                            caster,
                            cast,
                            definition,
                            "summoned",
                            "",
                            summon_id,
                            cast:get("target_id"),
                            limit
                        )
                    end
                    cast:set("effect_emitted", true)
                end
            end
        end,
    })
end

return M
