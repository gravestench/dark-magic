-- Shared ordered direct-damage transaction for melee and missile families.
--
-- Damage remains in the same 8.8 fixed-point unit as the decoded legacy data.
-- Player health currently has a whole-number compatibility field, while monster
-- health is already raw. This module is the one visible place that bridges them.

local random = require("engine.authority_random/v1")
local mitigation = require("d2legacy.policy.mitigation")

local M = {}

function M.roll(minimum_raw, maximum_raw)
    assert(maximum_raw >= minimum_raw, "invalid missile damage range")
    return minimum_raw + random.integer("d2legacy.combat.missile.damage", maximum_raw - minimum_raw + 1)
end

function M.resolve(target, damage_raw, ecs, channel)
    channel = channel or "fire"
    local applied = mitigation.apply(damage_raw, channel, ecs.get(target, "d2legacy.combat.defense"))
    local monster = ecs.get(target, "d2legacy.monster.stats")
    if monster then
        local before = monster:get("health")
        local remaining = math.max(0, before - applied)
        monster:set("health", remaining)
        return {
            channel = channel,
            rolled_damage_raw = damage_raw,
            damage_raw = applied,
            remaining_health_raw = remaining,
            lethal = before > 0 and remaining == 0,
        }
    end

    local player = assert(ecs.get(target, "d2legacy.player.vitals"), "damage target has no health component")
    -- The current player-vitals component stores whole health. Quantize the
    -- applied amount once at this boundary so the result record agrees with
    -- the state that was actually committed. Exact Expansion 1.14d player-life
    -- storage and fractional-damage ordering remain an owned-runtime probe.
    local before = player:get("health")
    local applied_raw = math.floor(applied / 256) * 256
    local remaining = math.max(0, before - math.floor(applied_raw / 256))
    player:set("health", remaining)
    return {
        channel = channel,
        rolled_damage_raw = damage_raw,
        damage_raw = applied_raw,
        remaining_health_raw = remaining * 256,
        lethal = before > 0 and remaining == 0,
    }
end

return M
