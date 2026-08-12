-- The player's authoritative belt container shape.
--
-- Diablo II belts have a bounded maximum, so explicit typed slots are clearer
-- at the Lua/Go boundary than an untyped nested table. Capacity says how many
-- of these sixteen possible slots the currently equipped belt exposes.

local ecs = require("engine.ecs/v1")

local M = {}

local function fields()
    local result = {{ name = "capacity", type = "i64" }}
    for slot = 1, 16 do
        result[#result + 1] = { name = "slot_" .. slot, type = "string" }
    end
    return result
end

function M.register()
    ecs.component({
        name = "d2legacy.player.belt",
        version = 1,
        fields = fields(),
    })
end

return M
