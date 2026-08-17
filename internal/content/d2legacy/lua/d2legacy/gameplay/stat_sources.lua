-- Resolve named stat-source facts for one ECS target.
--
-- Source production and consumption stay independent: equipment, states,
-- auras, and future effects publish the same component vocabulary while each
-- gameplay boundary asks only for the stat it owns.

local ecs = require("engine.ecs/v1")
local resolution = require("d2legacy.policy.stat_resolution")
local M = {}

function M.collect(entities, target, stat)
    local result = {}
    for _, entity in ipairs(entities) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        if source and source:get("target"):id() == target:id() and source:get("stat") == stat then
            result[#result + 1] = {
                id = source:get("source_id"),
                operation = source:get("operation") ~= "" and source:get("operation") or "add",
                value = source:get("value"),
                order = source:get("order"),
            }
        end
    end
    return result
end

function M.resolve(entities, target, stat, base)
    return resolution.resolve(base or 0, M.collect(entities, target, stat))
end

return M
