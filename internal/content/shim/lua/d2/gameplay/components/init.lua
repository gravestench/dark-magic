-- Register every component schema needed by the playable world slice.
--
-- This is intentionally boring. It is the table of contents for world data:
-- readers can see the domains at a glance and open only the file they need.

local modules = {
    require("d2.gameplay.components.player_identity"),
    require("d2.gameplay.components.player_animation"),
    require("d2.gameplay.components.player_skills"),
    require("d2.gameplay.components.player_belt"),
    require("d2.gameplay.components.world_spatial"),
}

local M = {}

function M.register()
    for _, module in ipairs(modules) do module.register() end
end

return M
