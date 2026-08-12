-- Derive stable actor facing from authoritative velocity.
--
-- A stopped actor retains its last direction. Consumers choose their authored
-- direction count on the component, so the same system serves 16-way players
-- and 8-way monsters without knowing either actor type.

local direction = require("d2legacy.policy.direction")
local ecs = require("engine.ecs/v1")
local M = {}

function M.register()
    ecs.system({id="d2legacy.world.velocity_facing",phase="presentation",
        query={all={"d2legacy.world.velocity","d2legacy.world.facing"}},
        read={"d2legacy.world.velocity"},write={"d2legacy.world.facing"},
        update=function(_, entities)
            for _, entity in ipairs(entities) do
                local velocity = ecs.get(entity, "d2legacy.world.velocity")
                local facing = ecs.get(entity, "d2legacy.world.facing")
                local x, y = velocity:get("x"), velocity:get("y")
                if x ~= 0 or y ~= 0 then
                    facing:set("direction", direction.quantize(
                        x, y, facing:get("directions")))
                end
            end
        end})
end

return M
