-- Player skill selection and the learned-skill collection.
--
-- A player has one assignment component for the two mouse-button slots. Each
-- learned skill is its own entity pointing back to the player. That avoids an
-- arbitrary mutable array hidden inside one component value.

local ecs = require("engine.ecs/v1")

local M = {}

function M.register()
    ecs.component({
        name = "d2legacy.player.skill_assignment",
        version = 1,
        fields = {
            { name = "left", type = "i64" },
            { name = "right", type = "i64" },
        },
    })

    ecs.component({
        name = "d2legacy.player.learned_skill",
        version = 1,
        fields = {
            { name = "owner", type = "entity" },
            { name = "skill_id", type = "i64" },
            { name = "level", type = "i64" },
            { name = "list_row", type = "i64" },
            { name = "left_allowed", type = "bool" },
            { name = "right_allowed", type = "bool" },
        },
    })

    ecs.component({
        name = "d2legacy.player.skill_hard_level",
        version = 1,
        fields = { { name = "level", type = "i64" } },
    })
end

return M
