-- Return life and mana to a player from the physical damage of a landed hit.
--
-- High-confidence recovered from pinned 1.10f D2MOO (UNIT/SUnitDmg.cpp);
-- awaiting 1.14d runtime confirmation. Only physical damage leeches, capped
-- at the defender's life before the hit. The attacker's leech percentages are
-- scaled by 64, divided by the session's steal divisors, applied to the
-- physical damage, scaled by the defender's drain percentage, and returned to
-- raw 8.8 units. A monster with no authored drain cannot be leeched from.
--
-- Deferred: monster and hireling leech, skill-granted leech, defender mana
-- loss, and the leech overlays.

local stat_sources = require("d2legacy.gameplay.stat_sources")
local difficulty_levels = require("d2legacy.data.difficulty_levels")

local M = {}

local SCALE = 64

local function percentage(value, percent)
    return math.floor(value * percent / 100)
end

local function leeched_raw(physical, leech_percent, divisor, drain)
    if leech_percent <= 0 then
        return 0
    end
    local amount = percentage(physical, math.floor(leech_percent * SCALE / divisor))
    if drain ~= 100 then
        amount = percentage(amount, drain)
    end
    return math.floor(amount / SCALE)
end

-- Returns the raw life and mana actually restored.
function M.apply(attacker, target, result, entities, ecs)
    local vitals = ecs.get(attacker, "d2legacy.player.vitals")
    if not result or not vitals or vitals:get("health") <= 0 or not ecs.get(attacker, "d2legacy.player.identity") then
        return 0, 0
    end
    local life_percent = stat_sources.resolve(entities, attacker, "lifedrainmindam", 0)
    local mana_percent = stat_sources.resolve(entities, attacker, "manadrainmindam", 0)
    if life_percent <= 0 and mana_percent <= 0 then
        return 0, 0
    end
    local monster = ecs.get(target, "d2legacy.monster.stats")
    local drain = monster and monster:get("drain_percent") or 100
    if drain <= 0 then
        return 0, 0
    end
    local health_before = result.remaining_health_raw + result.damage_raw
    local physical = math.min(result.mitigated.physical, health_before)
    if physical <= 0 then
        return 0, 0
    end

    local life = leeched_raw(physical, life_percent, difficulty_levels.life_steal_divisor(), drain)
    local mana = leeched_raw(physical, mana_percent, difficulty_levels.mana_steal_divisor(), drain)

    -- Player life is stored whole; leeched fractions below one point are lost
    -- until player life moves to raw storage.
    local health = vitals:get("health")
    local restored_life = math.min(math.floor(life / 256), vitals:get("max_health") - health)
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

return M
