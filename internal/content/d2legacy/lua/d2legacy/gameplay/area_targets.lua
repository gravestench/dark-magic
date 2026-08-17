-- Stable point-centered unit selection for authoritative area effects.

local ecs = require("engine.ecs/v1")
local M = {}

function M.hostile_monsters(entities, caster, center_x, center_y, radius)
    local caster_location = assert(ecs.get(caster, "d2legacy.world.location"))
    local result = {}
    for _, entity in ipairs(entities) do
        local selectable = ecs.get(entity, "d2legacy.world.selectable")
        local stats = ecs.get(entity, "d2legacy.monster.stats")
        local position = ecs.get(entity, "d2legacy.world.position")
        local location = ecs.get(entity, "d2legacy.world.location")
        local collider = ecs.get(entity, "d2legacy.world.collider")
        if
            selectable
            and selectable:get("kind") == "hostile"
            and stats
            and stats:get("health") > 0
            and position
            and location
            and collider
            and location:get("act") == caster_location:get("act")
            and location:get("level_id") == caster_location:get("level_id")
        then
            local dx = position:get("x") - center_x
            local dy = position:get("y") - center_y
            local reach = radius + collider:get("radius")
            if dx * dx + dy * dy <= reach * reach then
                result[#result + 1] = { entity = entity, id = selectable:get("id") }
            end
        end
    end
    table.sort(result, function(left, right)
        return left.id < right.id
    end)
    return result
end

return M
