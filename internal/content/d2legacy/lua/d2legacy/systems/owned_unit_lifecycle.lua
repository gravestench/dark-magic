-- Expire temporary summons at their checkpointed absolute lifetime boundary.

local ecs = require("engine.ecs/v1")
local M = {}

function M.register()
    ecs.system({
        id="d2legacy.owned_unit.lifecycle", phase="effects",
        -- Room inactivity retains the authoritative entity and excludes its
        -- simulation. The absolute expiration remains unchanged, so an
        -- already-expired unit is removed on its first active tick. Exact
        -- Expansion 1.14d inactive timer aging remains probe-gated.
        query={
            all={"d2legacy.owned_unit"},
            none={"d2legacy.world.inactive"},
        },
        read={"d2legacy.owned_unit"}, write={"d2legacy.owned_unit"},
        update=function(context, entities, structural)
            for _, entity in ipairs(entities) do
                local relation = ecs.get(entity, "d2legacy.owned_unit")
                local expires = relation:get("expires_tick")
                if relation:get("active") and expires > 0 and context.tick >= expires then
                    structural:destroy(entity)
                end
            end
        end,
    })
end

return M
