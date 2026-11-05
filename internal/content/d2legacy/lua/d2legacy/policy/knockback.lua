-- Supplied-state missile chance policy, statically verified for LoD 1.13c.
-- The caller owns the result and seed pair; this module retains no combat state.
local integer = require("engine.integer/v1")

local M = {}

function M.missile(context, ops)
    if context.state54_result ~= 0 then
        return
    end
    if context.chance == 0 then
        return
    end

    local roll = ops.rand(context.missile_seed, 100)
    if roll < context.chance and integer.band_u32(context.damage.result, 8) == 0 then
        context.damage.result = context.damage.result + 8
    end
end

return M
