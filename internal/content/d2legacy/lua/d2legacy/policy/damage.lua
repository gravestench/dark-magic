-- Diablo damage and death policy for the first Fire Bolt slice.
--
-- Damage remains in the same 8.8 fixed-point unit as the decoded legacy data.
-- Player health currently has a whole-number compatibility field, while monster
-- health is already raw. This module is the one visible place that bridges them.

local random = require("engine.authority_random/v1")

local M = {}

function M.roll_fire(minimum_raw, maximum_raw)
    assert(maximum_raw >= minimum_raw, "invalid Fire damage range")
    return minimum_raw + random.integer(
        "d2legacy.combat.fire_bolt.damage",
        maximum_raw - minimum_raw + 1)
end

function M.apply(target, damage_raw, ecs)
    local monster = ecs.get(target, "d2.monster.stats")
    if monster then
        local before = monster:get("health")
        local remaining = math.max(0, before - damage_raw)
        monster:set("health", remaining)
        return remaining, before > 0 and remaining == 0
    end

    local player = assert(ecs.get(target, "d2.player.vitals"),
        "Fire Bolt target has no health component")
    local before_raw = player:get("health") * 256
    local remaining_raw = math.max(0, before_raw - damage_raw)
    player:set("health", math.floor(remaining_raw / 256))
    return remaining_raw, before_raw > 0 and remaining_raw == 0
end

return M
