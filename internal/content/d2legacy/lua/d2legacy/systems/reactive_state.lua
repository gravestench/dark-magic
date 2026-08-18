-- Turn factual combat results into reactions snapshotted by active states.
--
-- The state instance identifies the authored trigger and carries all dynamic
-- cast values. Static missile facts remain in the decoded definition. Empty
-- observation markers keep this consumer independent from death/reflection.

local ecs = require("engine.ecs/v1")
local cold_duration = require("d2legacy.policy.cold_duration")
local damage = require("d2legacy.policy.damage")
local damage_bundle = require("d2legacy.policy.damage_bundle")
local game_rules = require("d2legacy.policy.game_rules")
local geometry = require("d2legacy.policy.geometry")
local projectile_spawn = require("d2legacy.gameplay.projectile_spawn")
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

local function reactions_for(entities, defender, trigger)
    local result = {}
    for _, entity in ipairs(entities) do
        local instance = ecs.get(entity, "d2legacy.state.instance")
        if instance and instance:get("target"):id() == defender:id() and instance:get("reaction") == trigger then
            result[#result + 1] = instance
        end
    end
    return result
end

local function living(entity)
    local monster = ecs.get(entity, "d2legacy.monster.stats")
    if monster then
        return monster:get("health") > 0
    end
    local player = ecs.get(entity, "d2legacy.player.vitals")
    return player and player:get("health") > 0
end

local function apply_reaction_state(reaction, target, source_suffix, structural)
    if not living(target) then
        return
    end
    local player = ecs.get(target, "d2legacy.player.identity")
    local cannot_freeze = player or ecs.get(target, "d2legacy.monster.freeze_immune")
    local state_id = cannot_freeze and reaction:get("reaction_chill_state_id") or reaction:get("reaction_state_id")
    local stat = cannot_freeze and reaction:get("reaction_chill_stat") or reaction:get("reaction_stat")
    local stat_value = cannot_freeze and reaction:get("reaction_chill_stat_value")
        or reaction:get("reaction_stat_value")
    local duration =
        cold_duration.target_frames(reaction:get("reaction_duration"), target, game_rules.difficulty(), ecs)
    if state_id == "" or duration <= 0 then
        return
    end
    structural:create({
        ["d2legacy.state.request"] = {
            operation = "apply",
            target = target,
            state_id = state_id,
            source_id = reaction:get("source_id") .. source_suffix,
            duration = duration,
            policy = "refresh_same_source",
            stat = stat,
            stat_operation = stat ~= "" and "add" or "",
            stat_value = stat_value,
            stat_order = 300,
            action_disabled = not cannot_freeze and reaction:get("reaction_disables_action"),
        },
    })
end

local function emit_direct_cold(context, reaction, attacker, defender_id, attacker_id, structural)
    local amount = damage.roll(reaction:get("reaction_minimum_damage_raw"), reaction:get("reaction_maximum_damage_raw"))
    local result = damage.resolve(attacker, damage_bundle.single("cold", amount), ecs)
    structural:create({
        ["d2legacy.combat.event"] = {
            kind = result.lethal and "unit_died" or "damage_applied",
            tick = context.tick,
            attacker_id = defender_id,
            target_id = attacker_id,
            source_kind = "state_reaction",
            damage_channel = result.channel,
            rolled_damage_raw = result.rolled_damage_raw,
            damage_raw = result.damage_raw,
            remaining_health_raw = result.remaining_health_raw,
        },
        ["d2legacy.combat.damage_bundle"] = damage_bundle.stage_component(result.rolled, result.mitigated),
    })
    if not result.lethal then
        apply_reaction_state(reaction, attacker, ":melee-response", structural)
    end
end

local function emit_reaction_presentation(context, reaction, defender, definitions, structural)
    local definition =
        assert(definitions[reaction:get("reaction_skill_id")], "active reaction references an unknown skill definition")
    if definition.reaction_overlay == "" then
        return
    end
    structural:create({
        ["d2legacy.presentation.effect_cue"] = {
            kind = "state_reaction",
            tick = context.tick,
            target = defender,
            overlay_id = definition.reaction_overlay,
            sound = "",
        },
    })
end

local function spawn_return_missile(context, event_entity, reaction, defender, attacker, definition, structural)
    local missile = assert(definition.reaction_missile, "ranged state reaction has no missile definition")
    local source = assert(ecs.get(defender, "d2legacy.world.position"))
    local target = assert(ecs.get(attacker, "d2legacy.world.position"))
    local dx, dy = geometry.normalized_direction(source:get("x"), source:get("y"), target:get("x"), target:get("y"))
    local owner_id = projectile_spawn.selectable_id(defender)
    local target_id = projectile_spawn.selectable_id(attacker)
    local cast_id = reaction:get("source_id") .. ":return:tick:" .. context.tick .. ":event:" .. event_entity:id()
    structural:create(projectile_spawn.components(defender, missile, {
        owner_id = owner_id,
        cast_id = cast_id,
        velocity_x = dx * missile.speed_per_tick,
        velocity_y = dy * missile.speed_per_tick,
        target_x = target:get("x"),
        target_y = target:get("y"),
        minimum_damage_raw = reaction:get("reaction_minimum_damage_raw"),
        maximum_damage_raw = reaction:get("reaction_maximum_damage_raw"),
        on_hit_state_id = reaction:get("reaction_state_id"),
        on_hit_state_source_id = reaction:get("source_id") .. ":missile-response:" .. target_id,
        on_hit_state_duration = reaction:get("reaction_duration"),
        on_hit_state_duration_policy = "cold",
    }))
    local sound = projectile_spawn.sound_cue(defender, context.tick, "missile_spawn", missile.travel_sound)
    if sound then
        structural:create(sound)
    end
end

function M.register(definitions)
    definitions = definitions or {}
    ecs.system({
        id = "d2legacy.state.react_to_combat",
        phase = "effects",
        after = { "d2legacy.monster.death" },
        query = {
            any = {
                "d2legacy.combat.melee_event",
                "d2legacy.combat.event",
                "d2legacy.state.instance",
                "d2legacy.world.selectable",
            },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.melee_event",
            "d2legacy.combat.event",
            "d2legacy.state.reaction_observed",
            "d2legacy.state.instance",
            "d2legacy.world.selectable",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.player.identity",
            "d2legacy.player.vitals",
            "d2legacy.monster.stats",
            "d2legacy.monster.freeze_immune",
            "d2legacy.combat.defense",
        },
        write = {
            "d2legacy.state.reaction_observed",
            "d2legacy.state.request",
            "d2legacy.combat.event",
            "d2legacy.combat.damage_bundle",
            "d2legacy.player.vitals",
            "d2legacy.monster.stats",
            "d2legacy.missile.projectile",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.room_resident",
            "d2legacy.presentation.effect_cue",
        },
        update = function(context, entities, structural)
            local by_id = selectable_index(entities)
            for _, event_entity in ipairs(entities) do
                if not ecs.get(event_entity, "d2legacy.state.reaction_observed") then
                    local melee = ecs.get(event_entity, "d2legacy.combat.melee_event")
                    local result = ecs.get(event_entity, "d2legacy.combat.event")
                    if melee then
                        structural:set(event_entity, "d2legacy.state.reaction_observed", {})
                        local defender = by_id[melee:get("target_id")]
                        local attacker = by_id[melee:get("attacker_id")]
                        if defender and attacker and living(attacker) then
                            if melee:get("hit") and melee:get("damage_raw") > 0 then
                                for _, reaction in ipairs(reactions_for(entities, defender, "melee_damage_freeze")) do
                                    apply_reaction_state(reaction, attacker, ":melee-response", structural)
                                    emit_reaction_presentation(context, reaction, defender, definitions, structural)
                                end
                            end
                            for _, reaction in ipairs(reactions_for(entities, defender, "melee_attack_cold")) do
                                emit_reaction_presentation(context, reaction, defender, definitions, structural)
                                emit_direct_cold(
                                    context,
                                    reaction,
                                    attacker,
                                    melee:get("target_id"),
                                    melee:get("attacker_id"),
                                    structural
                                )
                            end
                        end
                    elseif result and result:get("source_kind") == "missile" then
                        structural:set(event_entity, "d2legacy.state.reaction_observed", {})
                        local defender = by_id[result:get("target_id")]
                        local attacker = by_id[result:get("attacker_id")]
                        if defender and attacker and living(defender) and living(attacker) then
                            for _, reaction in ipairs(reactions_for(entities, defender, "missile_hit_return")) do
                                local definition = assert(
                                    definitions[reaction:get("reaction_skill_id")],
                                    "active reaction references an unknown skill definition"
                                )
                                emit_reaction_presentation(context, reaction, defender, definitions, structural)
                                spawn_return_missile(
                                    context,
                                    event_entity,
                                    reaction,
                                    defender,
                                    attacker,
                                    definition,
                                    structural
                                )
                            end
                        end
                    end
                end
            end
        end,
    })
end

return M
