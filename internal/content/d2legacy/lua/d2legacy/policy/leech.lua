-- Return life and mana to an attacker from the physical damage of a landed hit.
--
-- High-confidence recovered from pinned 1.10f D2MOO (UNIT/SUnitDmg.cpp and
-- MISSILES/MissMode.cpp); awaiting 1.14d runtime confirmation.
--
-- Only physical damage leeches, capped at the defender's life before the hit.
-- A monster defender scales leech by its MonStats Drain for the difficulty,
-- and a blank column cannot be leeched from; any other defender counts as 100.
--
-- Players and hirelings leech a percentage. The percent is scaled by 64,
-- divided by the session's steal divisor (players only), applied to the
-- physical damage, scaled by drain, and returned to raw 8.8 units.
--
-- Other monsters leech a flat amount rolled from their drain-damage stats.
-- Basic attacks use the default skill damage share of 128/128, which adds the
-- roll to itself. Life is capped by the physical damage and mana by the
-- defender's mana; both heal the monster's life after drain, the defender
-- loses the mana, and healing stops at maximum life.
--
-- Deferred: monster missile leech, skill-granted leech, stamina leech, and the
-- recovered player-path quirk that also drains defender mana.

local random = require("engine.authority_random/v1")
local stat_sources = require("d2legacy.gameplay.stat_sources")
local difficulty_levels = require("d2legacy.data.difficulty_levels")

local M = {}

local SCALE = 64
local SKILL_DAMAGE_SHARE = 128

local function percentage(value, percent)
    return math.floor(value * percent / 100)
end

local function attacker_kind(attacker, ecs)
    if ecs.get(attacker, "d2legacy.player.identity") then
        return "player"
    end
    if not ecs.get(attacker, "d2legacy.monster.stats") then
        return nil
    end
    local relation = ecs.get(attacker, "d2legacy.owned_unit")
    if relation and relation:get("active") and relation:get("category") == "hireable" then
        return "hireling"
    end
    return "monster"
end

local function cue(structural, tick, attacker, overlay_id)
    structural:create({
        ["d2legacy.presentation.effect_cue"] = {
            kind = "leech",
            tick = tick,
            target = attacker,
            overlay_id = overlay_id,
            sound = "",
        },
    })
end

local function percent_raw(physical, leech_percent, divisor, drain)
    if leech_percent <= 0 then
        return 0
    end
    local amount = percentage(physical, math.floor(leech_percent * SCALE / divisor))
    if drain ~= 100 then
        amount = percentage(amount, drain)
    end
    return math.floor(amount / SCALE)
end

local function heal_monster(stats, amount)
    local health = stats:get("health")
    local restored = math.max(0, math.min(amount, stats:get("max_health") - health))
    if restored > 0 then
        stats:set("health", health + restored)
    end
    return restored
end

-- Player life is stored whole; leeched fractions below one point are lost
-- until player life moves to raw storage.
local function heal_player(vitals, life, mana)
    local health = vitals:get("health")
    local restored_life = math.max(0, math.min(math.floor(life / 256), vitals:get("max_health") - health))
    if restored_life > 0 then
        vitals:set("health", health + restored_life)
    end
    local mana_raw = vitals:get("mana_raw")
    local restored_mana = math.max(0, math.min(mana, vitals:get("max_mana_raw") - mana_raw))
    if restored_mana > 0 then
        vitals:set("mana_raw", mana_raw + restored_mana)
        vitals:set("mana", math.floor((mana_raw + restored_mana) / 256))
    end
    return restored_life * 256, restored_mana
end

local function percent_leech(kind, attacker, physical, drain, entities, ecs, structural, tick)
    local life_percent = stat_sources.resolve(entities, attacker, "lifedrainmindam", 0)
    local mana_percent = stat_sources.resolve(entities, attacker, "manadrainmindam", 0)
    if life_percent <= 0 and mana_percent <= 0 then
        return 0, 0
    end
    local life_divisor, mana_divisor = 1, 1
    if kind == "player" then
        life_divisor, mana_divisor = difficulty_levels.life_steal_divisor(), difficulty_levels.mana_steal_divisor()
    end
    local life = percent_raw(physical, life_percent, life_divisor, drain)
    local mana = percent_raw(physical, mana_percent, mana_divisor, drain)

    -- The overlay follows the leech stats, even when nothing was restored.
    if life_percent > 0 and mana_percent > 0 then
        local pick = random.integer("d2legacy.combat.leech.overlay", 2)
        cue(structural, tick, attacker, pick == 0 and "steallife" or "stealmana")
    elseif life_percent > 0 then
        cue(structural, tick, attacker, "steallife")
    else
        cue(structural, tick, attacker, "stealmana")
    end

    if kind == "player" then
        return heal_player(ecs.get(attacker, "d2legacy.player.vitals"), life, mana)
    end
    -- Monsters have no mana pool, so a hireling's mana leech restores nothing.
    return heal_monster(ecs.get(attacker, "d2legacy.monster.stats"), life), 0
end

local function flat_roll(entities, attacker, minimum_stat, maximum_stat)
    local maximum = stat_sources.resolve(entities, attacker, maximum_stat, 0) * 256
    local amount = 0
    if maximum >= 8 then
        local minimum = stat_sources.resolve(entities, attacker, minimum_stat, 0) * 256
        if maximum > 0 then
            amount = minimum
            if maximum > minimum then
                amount = amount + random.integer("d2legacy.combat.leech.roll", maximum - minimum)
            end
        end
    end
    amount = math.max(amount, 0)
    return amount + math.floor(amount * SKILL_DAMAGE_SHARE / 128)
end

local function flat_leech(attacker, target, physical, drain, entities, ecs, structural, tick)
    local life_roll = flat_roll(entities, attacker, "lifedrainmindam", "lifedrainmaxdam")
    local mana_roll = flat_roll(entities, attacker, "manadrainmindam", "manadrainmaxdam")
    if life_roll <= 0 and mana_roll <= 0 then
        return 0, 0
    end
    local defender = ecs.get(target, "d2legacy.player.vitals")
    local defender_mana = defender and defender:get("mana_raw") or 0
    local total = math.min(life_roll, physical) + math.min(mana_roll, defender_mana)
    if drain ~= 100 then
        total = percentage(total, drain)
    end
    local restored = heal_monster(ecs.get(attacker, "d2legacy.monster.stats"), total)
    if restored > 0 then
        cue(structural, tick, attacker, "steallife")
    end
    if defender and mana_roll > 0 then
        local remaining = defender_mana - mana_roll
        if remaining < 256 then
            remaining = 0
        end
        defender:set("mana_raw", remaining)
        defender:set("mana", math.floor(remaining / 256))
    end
    return restored, 0
end

-- Returns the raw life and mana restored to the attacker. A missile hit
-- leeches with its owner's current stats.
function M.apply(attacker, target, result, entities, ecs, structural, tick, source)
    local kind = attacker and attacker_kind(attacker, ecs)
    if not result or not kind or (source == "missile" and kind == "monster") then
        return 0, 0
    end
    if kind == "player" then
        if ecs.get(attacker, "d2legacy.player.vitals"):get("health") <= 0 then
            return 0, 0
        end
    elseif ecs.get(attacker, "d2legacy.monster.stats"):get("health") <= 0 then
        return 0, 0
    end
    local monster = ecs.get(target, "d2legacy.monster.stats")
    local drain = monster and monster:get("drain_percent") or 100
    if drain <= 0 then
        return 0, 0
    end
    local physical = math.max(0, math.min(result.mitigated.physical, result.remaining_health_raw + result.damage_raw))
    if kind == "monster" then
        return flat_leech(attacker, target, physical, drain, entities, ecs, structural, tick)
    end
    if physical <= 0 then
        return 0, 0
    end
    return percent_leech(kind, attacker, physical, drain, entities, ecs, structural, tick)
end

return M
