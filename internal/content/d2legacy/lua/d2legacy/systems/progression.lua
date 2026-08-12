-- Promote players from the cumulative, class-specific Experience.txt table.

local ecs = require("engine.ecs/v1")
local policy = require("d2legacy.data.progression")
local M = {}

function M.register(thresholds)
    ecs.system({
        id="d2legacy.player.progression", phase="effects",
        query={all={"d2legacy.player.identity", "d2legacy.player.progress"}},
        read={"d2legacy.player.identity", "d2legacy.player.progress"},
        write={"d2legacy.player.progress"},
        update=function(_, entities)
            for _, entity in ipairs(entities) do
                local identity = ecs.get(entity, "d2legacy.player.identity")
                local progress = ecs.get(entity, "d2legacy.player.progress")
                local level = policy.level_for(thresholds, identity:get("class"),
                    progress:get("experience"), progress:get("level"))
                if level ~= progress:get("level") then progress:set("level", level) end
            end
        end,
    })
end

return M
