-- Resolve impact requests emitted by reusable approach/animation mechanisms.
--
-- Movement and animation merely say that an impact moment happened. This
-- authoritative system chooses an in-range target, rolls Diablo's hit and
-- damage formulas, mutates health, and emits a factual result event. Keeping
-- that distinction lets presentation animate without becoming game authority.

local ecs = require("engine.ecs/v1")
local combat_target = require("d2legacy.gameplay.combat_target")
local policy = require("d2legacy.policy.melee")
local damage_policy = require("d2legacy.policy.damage")
local damage_bundle = require("d2legacy.policy.damage_bundle")
local stat_sources = require("d2legacy.gameplay.stat_sources")
local M = {}

local function identity(entity)
    local selectable = ecs.get(entity, "d2legacy.world.selectable")
    return selectable and selectable:get("id") or "entity:" .. entity:id()
end

local function target_for(attacker, wanted, hand, candidates)
    local profile = ecs.get(attacker, "d2legacy.combat.melee_profile")
    local attack_range = hand == "larm" and profile:get("dual_wield") and profile:get("secondary_range")
        or profile:get("range")
    -- Named attacks and targetless Shift-Attacks both revalidate current
    -- alignment, life, location, range, and barrier state at impact.
    return combat_target.select_melee(attacker, wanted, attack_range, candidates)
end

local function combat_values(entity, entities)
    local monster = ecs.get(entity, "d2legacy.monster.stats")
    if monster then
        return monster:get("level"),
            stat_sources.resolve(entities, entity, "item_tohit_percent", monster:get("attack_rating")),
            monster:get("defense")
    end
    local progress = assert(ecs.get(entity, "d2legacy.player.progress"), "player has no progress")
    local stats = assert(ecs.get(entity, "d2legacy.player.combat_stats"), "player has no combat stats")
    return progress:get("level"), stats:get("attack_rating"), stats:get("defense")
end

local function event(structural, values, damage)
    local components = {
        ["d2legacy.combat.melee_event"] = values,
        ["d2legacy.combat.attack_result"] = {
            tick = values.tick,
            attacker_id = values.attacker_id,
            target_id = values.target_id,
            source_kind = "melee",
            outcome = values.outcome,
            attack_rating = values.attack_rating,
            defense = values.defense,
            hit_chance = values.hit_chance,
        },
    }
    if damage then
        components["d2legacy.combat.event"] = {
            kind = damage.lethal and "unit_died" or "damage_applied",
            tick = values.tick,
            attacker_id = values.attacker_id,
            target_id = values.target_id,
            source_kind = "melee",
            damage_channel = damage.channel,
            rolled_damage_raw = damage.rolled_damage_raw,
            damage_raw = damage.damage_raw,
            remaining_health_raw = damage.remaining_health_raw,
        }
        components["d2legacy.combat.damage_bundle"] = damage_bundle.stage_component(damage.rolled, damage.mitigated)
    end
    structural:create(components)
end

local function damage_range(profile, hand)
    if hand == "larm" and profile:get("dual_wield") then
        return profile:get("secondary_physical_min"), profile:get("secondary_physical_max")
    end
    return profile:get("physical_min"), profile:get("physical_max")
end

local function weapon_damage(attacker, profile, hand, entities)
    local minimum, maximum = damage_range(profile, hand)
    local result = { physical = policy.damage(minimum, maximum) }
    local fire_minimum = stat_sources.resolve(entities, attacker, "firemindam", 0)
    local fire_maximum = stat_sources.resolve(entities, attacker, "firemaxdam", 0)
    assert(fire_minimum >= 0 and fire_maximum >= fire_minimum, "invalid weapon fire-damage range")
    if fire_maximum > 0 then
        result.fire = policy.damage(fire_minimum, fire_maximum)
    end
    return damage_bundle.normalize(result)
end

local function hand_attack_rating(profile, total, hand)
    if hand == "larm" and profile:get("dual_wield") then
        return total - profile:get("primary_attack_rating") + profile:get("secondary_attack_rating")
    end
    return total
end

function M.register()
    ecs.system({
        id = "d2legacy.combat.basic_melee",
        phase = "combat",
        -- One broad deterministic query supplies both attackers and candidate
        -- targets. Systems may not perform a hidden second ECS query while a
        -- tick is running.
        query = {
            any = { "d2legacy.combat.basic_attack_request", "d2legacy.world.selectable", "d2legacy.stat.source" },
            none = { "d2legacy.world.inactive" },
        },
        read = {
            "d2legacy.combat.basic_attack_request",
            "d2legacy.world.selectable",
            "d2legacy.world.position",
            "d2legacy.world.location",
            "d2legacy.world.collider",
            "d2legacy.combat.melee_profile",
            "d2legacy.monster.stats",
            "d2legacy.player.vitals",
            "d2legacy.player.progress",
            "d2legacy.player.combat_stats",
            "d2legacy.combat.defense",
            "d2legacy.stat.source",
        },
        write = {
            "d2legacy.combat.basic_attack_request",
            "d2legacy.monster.stats",
            "d2legacy.player.vitals",
            "d2legacy.combat.melee_event",
            "d2legacy.combat.attack_result",
            "d2legacy.combat.event",
            "d2legacy.combat.damage_bundle",
        },
        update = function(context, entities, structural)
            for _, attacker in ipairs(entities) do
                local request = ecs.get(attacker, "d2legacy.combat.basic_attack_request")
                local profile = ecs.get(attacker, "d2legacy.combat.melee_profile")
                if request and profile then
                    assert(request:get("request_tick") <= context.tick, "future melee request")
                    local target, target_id =
                        target_for(attacker, request:get("target_id"), request:get("hand"), entities)
                    structural:remove(attacker, "d2legacy.combat.basic_attack_request")
                    local base = {
                        kind = "hit_resolved",
                        tick = context.tick,
                        attacker_id = identity(attacker),
                        target_id = target_id,
                        hit = false,
                        damage_raw = 0,
                        remaining_health_raw = 0,
                        hand = request:get("hand"),
                        attack_rating = 0,
                        defense = 0,
                        hit_chance = 0,
                        outcome = "invalidated",
                    }
                    local result
                    if target then
                        local attacker_level, attack_rating = combat_values(attacker, entities)
                        local defender_level, _, defense = combat_values(target, entities)
                        attack_rating = hand_attack_rating(profile, attack_rating, request:get("hand"))
                        base.attack_rating = attack_rating
                        base.defense = defense
                        base.hit_chance = policy.hit_chance(attacker_level, defender_level, attack_rating, defense)
                        base.outcome = "miss"
                        if policy.hits(attacker_level, defender_level, attack_rating, defense) then
                            result = damage_policy.resolve(
                                target,
                                weapon_damage(attacker, profile, request:get("hand"), entities),
                                ecs
                            )
                            base.hit = true
                            base.outcome = "hit"
                            base.remaining_health_raw = result.remaining_health_raw
                            base.damage_raw = result.damage_raw
                        end
                    end
                    event(structural, base, result)
                end
            end
        end,
    })
end

return M
