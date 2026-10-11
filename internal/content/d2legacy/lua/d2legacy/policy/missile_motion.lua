-- Owned 1.14d axial steps: Vel 30 gives 92160/65536; Vel 12 gives 36864/65536.
local M = {}

function M.speed_per_tick(velocity)
    return velocity * 3 / 64
end

return M
