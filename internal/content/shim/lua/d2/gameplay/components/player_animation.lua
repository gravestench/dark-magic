-- Authoritative player movement choice and legacy composite selection.
--
-- Appearance answers “which body/equipment recipe should be drawn?” Animation
-- answers “what is that body doing right now?” They are deliberately separate:
-- changing a helmet should not rewrite facing, and turning should not rewrite
-- equipment.

local ecs = require("engine.ecs/v1")

local M = {}

function M.register()
    ecs.component({
        name = "d2.player.movement_mode",
        version = 1,
        fields = {
            -- This is the player's walk/run choice. Actual motion still depends
            -- on authoritative velocity being non-zero.
            { name = "running", type = "bool" },
        },
    })

    ecs.component({
        name = "d2.player.appearance",
        version = 1,
        fields = {
            { name = "cof", type = "string" },
            { name = "token", type = "string" },
            { name = "palette", type = "string" },
            { name = "weapon_class", type = "string" },
        },
    })

    ecs.component({
        name = "d2.player.animation",
        version = 1,
        fields = {
            { name = "direction", type = "i64" }, -- Logical eight-way direction; presentation maps legacy storage order.
            { name = "mode", type = "string" },   -- NU, WL, RN, and future modes.
        },
    })
end

return M
