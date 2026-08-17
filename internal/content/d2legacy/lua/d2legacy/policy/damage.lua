-- Shared direct-damage policy for the first migrated missile family.
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

function M.apply(target, damage_raw, ecs, channel)
    local applied = mitigation.apply(damage_raw, channel or "fire", ecs.get(target, "d2legacy.combat.defense"))
    local monster = ecs.get(target, "d2legacy.monster.stats")
    if monster then
        local before = monster:get("health")
        local remaining = math.max(0, before - applied)
        monster:set("health", remaining)
        return remaining, before > 0 and remaining == 0, applied
    end

    local player = assert(ecs.get(target, "d2legacy.player.vitals"), "missile target has no health component")
    local before_raw = player:get("health") * 256
    local remaining_raw = math.max(0, before_raw - applied)
    player:set("health", math.floor(remaining_raw / 256))
    return remaining_raw, before_raw > 0 and remaining_raw == 0, applied
end

return M
