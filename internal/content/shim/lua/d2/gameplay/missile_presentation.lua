-- Resolve copied missile facts into one DCC presentation recipe.
-- Asset spelling and legacy direction interleave stop here; authoritative
-- movement remains plain world-space velocity and position.

local M = {}

local logical_to_cof = {1, 3, 5, 7, 0, 2, 4, 6}
local cof_to_dcc = {4, 0, 5, 1, 6, 2, 7, 3}

function M.direction(velocity_x, velocity_y, directions)
    if directions <= 1 then return 0 end
    local angle = math.atan(velocity_y, velocity_x)
    local bucket = math.floor((angle + math.pi / 8) / (math.pi / 4)) % 8
    local logical = ({3, 4, 0, 5, 1, 6, 2, 7})[bucket + 1]
    if directions == 8 then
        local cof = logical_to_cof[logical + 1]
        return cof_to_dcc[cof + 1]
    end
    return logical % directions
end

function M.logical_direction(logical, directions)
    if directions <= 1 then return 0 end
    if directions == 8 then
        local cof = logical_to_cof[(logical % 8) + 1]
        return cof_to_dcc[cof + 1]
    end
    return logical % directions
end

function M.resolve(snapshot)
    if not snapshot.dcc or snapshot.dcc == "" then return nil end
    return {
        path = snapshot.dcc,
        palette = snapshot.palette ~= "" and snapshot.palette or "data/global/palette/units/pal.dat",
        direction = snapshot.logical_direction ~= nil
            and M.logical_direction(snapshot.logical_direction, snapshot.directions)
            or M.direction(snapshot.velocity_x, snapshot.velocity_y, snapshot.directions),
        frames_per_second = snapshot.frames_per_second > 0 and snapshot.frames_per_second or 25,
        loop = snapshot.loop and "loop" or "once",
    }
end

return M
