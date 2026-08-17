-- Extensible authoritative operation dispatch for interacted world entities.
--
-- Interaction admission owns target identity and range. Behavior families then
-- opt in by component, so doors, shrines, warps, and later object families do
-- not turn the generic interaction command into one giant kind switch.

local ecs = require("engine.ecs/v1")
local M = {}
local handlers = {}

function M.register(component, handler)
    assert(type(component) == "string" and component ~= "", "operation component is required")
    assert(type(handler) == "function", "operation handler is required")
    assert(not handlers[component], "operation handler already registered for " .. component)
    handlers[component] = handler
end

function M.apply(owner, entity)
    local names = {}
    for component in pairs(handlers) do
        names[#names + 1] = component
    end
    table.sort(names)
    for _, component in ipairs(names) do
        local value = ecs.get(entity, component)
        if value then
            handlers[component](owner, entity, value)
            return true
        end
    end
    return false
end

return M
