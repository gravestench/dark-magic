-- Presentation-facing interaction API.
--
-- Pointer choices become replayable command intents. Snapshot reads come from
-- authoritative ECS state, so a screen never maintains a second source of
-- truth for the currently open NPC or world object.

local intents = require("engine.command_intent/v1")
local ecs = require("engine.ecs/v1")

local M = { api = 1 }

local function split(value)
    local result = {}
    for token in string.gmatch(value or "", "[^,]+") do
        result[#result + 1] = token
    end
    return result
end

function M.open(target_id)
    intents.submit("interaction.open", { target = target_id })
end

function M.open_at(x, y)
    intents.submit("interaction.open", { at = true, x = x, y = y })
end

function M.close()
    intents.submit("interaction.close", {})
end

function M.snapshot()
    for _, entity in ipairs(ecs.query({
        all = { "d2legacy.interaction.context" },
    })) do
        local context = ecs.get(entity, "d2legacy.interaction.context")
        if context:get("owner") == "local-player" then
            local target = context:get("target")
            local value = target
                and ecs.get(target, "d2legacy.interaction.target")
                or nil
            if not value then
                return { active = false, categories = {}, services = {} }
            end
            return {
                active = true,
                target_id = value:get("id"),
                npc = value:get("npc"),
                vendor = value:get("vendor"),
                categories = split(value:get("categories")),
                services = split(value:get("services")),
            }
        end
    end
    return { active = false, categories = {}, services = {} }
end

return M
