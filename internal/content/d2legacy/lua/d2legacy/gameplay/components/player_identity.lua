-- Core facts that say who a player character is and how they are progressing.
--
-- Components are data shapes only. They contain no update logic. Keeping these
-- related schemas together makes it easy to answer “what data describes a
-- player?” without also reading movement, rendering, or UI code.

local ecs = require("engine.ecs/v1")

local M = {}

function M.register()
    ecs.component({
        name = "d2legacy.player.identity",
        version = 1,
        fields = {
            { name = "character_id", type = "string" }, -- Durable save identity.
            { name = "player", type = "string" }, -- Session/connection owner.
            { name = "name", type = "string" }, -- Player-facing character name.
            { name = "class", type = "string" }, -- Amazon, Sorceress, and so on.
        },
    })

    ecs.component({
        name = "d2legacy.player.progress",
        version = 1,
        fields = {
            { name = "level", type = "i64" },
            { name = "experience", type = "i64" },
            { name = "unspent_skill_points", type = "i64" },
        },
    })

    ecs.component({
        name = "d2legacy.player.vitals",
        version = 1,
        fields = {
            { name = "health", type = "i64" },
            { name = "max_health", type = "i64" },
            { name = "mana", type = "i64" },
            { name = "max_mana", type = "i64" },
            -- Simulation keeps mana in Diablo's 8.8 fixed-point unit so costs
            -- such as fractional skill costs are never rounded away. The whole
            -- fields above remain convenient display values for Lua and UI.
            { name = "mana_raw", type = "i64" },
            { name = "max_mana_raw", type = "i64" },
        },
    })
end

return M
