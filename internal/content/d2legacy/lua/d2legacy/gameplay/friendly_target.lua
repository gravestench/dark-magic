-- Resolve a requested friendly unit against current authoritative ECS facts.

local ecs = require("engine.ecs/v1")
local M = {}

local function alive(entity, kind)
    if kind == "player" then
        local vitals = ecs.get(entity, "d2legacy.player.vitals")
        return vitals and vitals:get("health") > 0 and not ecs.get(entity, "d2legacy.player.death")
    end
    if kind == "friendly" then
        local stats = ecs.get(entity, "d2legacy.monster.stats")
        return stats and stats:get("health") > 0
    end
    return false
end

local function same_location(left, right)
    local left_location = ecs.get(left, "d2legacy.world.location")
    local right_location = ecs.get(right, "d2legacy.world.location")
    return left_location
        and right_location
        and left_location:get("act") == right_location:get("act")
        and left_location:get("level_id") == right_location:get("level_id")
end

function M.named(caster, wanted, candidates, policy)
    if wanted == "" then
        return policy.fallback_to_self and caster or nil
    end
    for _, candidate in ipairs(candidates) do
        local selectable = ecs.get(candidate, "d2legacy.world.selectable")
        if selectable and selectable:get("id") == wanted then
            local kind = selectable:get("kind")
            local allowed = candidate:id() == caster:id()
                or kind == "player" and policy.target_allies
                or kind == "friendly" and policy.target_owned_units
            if
                allowed
                and alive(candidate, kind)
                and same_location(caster, candidate)
                and not ecs.get(candidate, "d2legacy.world.inactive")
            then
                return candidate
            end
            break
        end
    end
    return policy.fallback_to_self and caster or nil
end

return M
