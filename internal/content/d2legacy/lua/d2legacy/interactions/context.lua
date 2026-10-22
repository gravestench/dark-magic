-- Shared interaction relationship and distance reads for authority and presentation.
local ecs = require("engine.ecs/v1")
local M = {}

function M.for_owner(owner)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.interaction.context" } })) do
        local value = ecs.get(entity, "d2legacy.interaction.context")
        if value:get("owner") == owner then return value end
    end
    return nil
end

function M.in_range(owner, entity, target, players)
    for _, player in ipairs(players or ecs.query({ all = { "d2legacy.player.identity", "d2legacy.world.position" } })) do
        local identity = ecs.get(player, "d2legacy.player.identity")
        if identity and identity:get("player") == owner then
            local position = ecs.get(player, "d2legacy.world.position")
            if not position then return false end
            local player_location = ecs.get(player, "d2legacy.world.location")
            local target_location = ecs.get(entity, "d2legacy.world.location")
            if player_location and target_location and player_location:get("level_id") ~= target_location:get("level_id") then
                return false
            end
            local dx, dy = position:get("x") - target:get("x"), position:get("y") - target:get("y")
            return dx * dx + dy * dy <= target:get("radius") * target:get("radius")
        end
    end
    -- Existing administrator/bootstrap interactions can precede player admission.
    return true
end

function M.target(context, players)
    if not context then return nil end
    local entity = context:get("target")
    local ok, target = pcall(ecs.get, entity, "d2legacy.interaction.target")
    if not ok or not target or ecs.get(entity, "d2legacy.world.inactive") or not M.in_range(context:get("owner"), entity, target, players) then
        return nil
    end
    return target
end

function M.active(owner)
    local context = M.for_owner(owner)
    local target = M.target(context)
    if not target then return nil end
    return context, target
end

return M
