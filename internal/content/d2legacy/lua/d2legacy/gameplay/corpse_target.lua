-- Resolve a named, currently consumable corpse for a skill transaction.

local ecs = require("engine.ecs/v1")
local movement_rules = require("d2legacy.movement_rules/v1")
local M = {}

function M.named(caster, wanted, entities, policy)
    policy = policy or {}
    if wanted == "" then
        return nil, "corpse_target_required"
    end
    local caster_location = ecs.get(caster, "d2legacy.world.location")
    if not caster_location or movement_rules.is_town(caster_location:get("level_id")) then
        return nil, "unavailable_in_town"
    end
    for _, entity in ipairs(entities) do
        local selectable = ecs.get(entity, "d2legacy.world.selectable")
        if selectable and selectable:get("id") == wanted then
            local death = ecs.get(entity, "d2legacy.monster.death")
            local location = ecs.get(entity, "d2legacy.world.location")
            if
                selectable:get("kind") ~= "corpse"
                or not death
                or not death:get("corpse_usable")
                or ecs.get(entity, "d2legacy.world.inactive")
                or not location
                or location:get("act") ~= caster_location:get("act")
                or location:get("level_id") ~= caster_location:get("level_id")
            then
                return nil, "corpse_target_unavailable"
            end
            if policy.requires_revivable_corpse and not ecs.get(entity, "d2legacy.monster.revivable") then
                return nil, "corpse_target_not_revivable"
            end
            return entity, ""
        end
    end
    return nil, "corpse_target_unavailable"
end

return M
