-- Materialize the owner-scoped party view consumed by local presentation and
-- the authenticated network projector. This cache is never gameplay authority.

local ecs = require("engine.ecs/v1")
local party = require("d2legacy.policy.party")
local M = {}

function M.register()
    ecs.system({
        id = "d2legacy.party.presentation_projection",
        phase = "presentation",
        query = { all = { "d2legacy.player.identity", "d2legacy.player.progress" } },
        read = { "d2legacy.player.identity", "d2legacy.player.progress" },
        write = { "d2legacy.player.party_view" },
        update = function(_, entities, structural)
            for _, entity in ipairs(entities) do
                local player_id = ecs.get(entity, "d2legacy.player.identity"):get("player")
                structural:set(entity, "d2legacy.player.party_view", party.project(player_id, entities))
            end
        end,
    })
end

return M
