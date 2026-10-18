-- Invalidated interactions must not reopen merely because an actor returns to
-- range. A new admitted interaction.open is required to start another menu.
local ecs = require("engine.ecs/v1")
local contexts = require("d2legacy.interactions.context")
local M = {}

function M.register()
    ecs.system({
        id = "d2legacy.npc_dialogue.lifetime",
        phase = "presentation",
        query = { any = { "d2legacy.interaction.context", "d2legacy.player.identity" } },
        read = {
            "d2legacy.interaction.target", "d2legacy.player.identity",
            "d2legacy.world.position", "d2legacy.world.location", "d2legacy.world.inactive",
        },
        write = { "d2legacy.interaction.context" },
        update = function(_, entities)
            for _, entity in ipairs(entities) do
                local context = ecs.get(entity, "d2legacy.interaction.context")
                if context and context:get("dialogue_page") ~= "" and not contexts.target(context, entities) then
                    context:set("dialogue_page", "")
                    context:set("dialogue_line", "")
                    context:set("dialogue_revision", context:get("dialogue_revision") + 1)
                end
            end
        end,
    })
end

return M
