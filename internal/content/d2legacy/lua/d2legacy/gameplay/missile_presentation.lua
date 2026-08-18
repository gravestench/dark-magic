-- Resolve copied missile facts into one DCC presentation recipe.
-- Asset spelling and legacy direction interleave stop here; authoritative
-- movement remains plain world-space velocity and position.

local M = {}

-- Missiles.txt does not use Overlay.txt's Trans numbering. In the owned target
-- records missile mode 1 is the luminous/additive path; mode 0 is ordinary
-- indexed alpha. Keep that table-specific interpretation at this adapter.
local blend_modes = {
    [0] = "alpha",
    [1] = "screen",
}

-- Missiles.txt's NumDirections names the independently encoded directions in
-- the DCC. Those directions are not stored in angular order. These tables are
-- the target DCC order sampled from the 64-way direction space for every
-- direction count present in the owned Expansion 1.14d Missiles.txt.
local dcc_directions = {
    [1] = { 0 },
    [4] = { 0, 1, 2, 3 },
    [8] = { 4, 0, 5, 1, 6, 2, 7, 3 },
    [16] = { 4, 8, 0, 9, 5, 10, 1, 11, 6, 12, 2, 13, 7, 14, 3, 15 },
    [32] = {
        4,
        16,
        8,
        17,
        0,
        18,
        9,
        19,
        5,
        20,
        10,
        21,
        1,
        22,
        11,
        23,
        6,
        24,
        12,
        25,
        2,
        26,
        13,
        27,
        7,
        28,
        14,
        29,
        3,
        30,
        15,
        31,
    },
}

function M.direction(velocity_x, velocity_y, directions)
    if directions <= 1 then
        return 0
    end
    if velocity_x == 0 and velocity_y == 0 then
        return 0
    end
    local encoded = assert(dcc_directions[directions], "unsupported missile DCC direction count")
    -- The embedded Lua 5.1 runtime exposes atan2 separately; math.atan accepts
    -- one argument and silently ignored X here, making opposite horizontal
    -- casts select the same art.
    local angle = math.atan2(velocity_y, velocity_x)
    -- Actual direction zero lies on the +X,+Y world diagonal. Quantize in
    -- world space before applying the DCC interleave; projection into the
    -- isometric pixel diamond happens later and must not collapse 16/32-way
    -- art back to the eight actor facings.
    local turns = (angle - math.pi / 4) / (2 * math.pi)
    local bucket = math.floor(turns * directions + 0.5) % directions
    return encoded[bucket + 1]
end

function M.logical_direction(logical, directions)
    if directions <= 1 then
        return 0
    end
    assert(dcc_directions[directions], "unsupported missile DCC direction count")
    -- Semantic actor directions and the lab's explicit selector already use
    -- the target DCC's readable legacy IDs. Only continuous velocity needs
    -- angular quantization and interleave conversion.
    return logical % directions
end

function M.resolve(snapshot)
    if not snapshot.dcc or snapshot.dcc == "" then
        return nil
    end
    local transparency_mode = snapshot.transparency_mode or 0
    return {
        path = snapshot.dcc,
        palette = snapshot.palette ~= "" and snapshot.palette or "data/global/palette/units/pal.dat",
        direction = snapshot.logical_direction ~= nil
                and M.logical_direction(snapshot.logical_direction, snapshot.directions)
            or M.direction(snapshot.velocity_x, snapshot.velocity_y, snapshot.directions),
        frames_per_second = snapshot.frames_per_second > 0 and snapshot.frames_per_second or 25,
        loop = snapshot.loop and "loop" or "once",
        blend = assert(blend_modes[transparency_mode], "unsupported missile transparency mode"),
    }
end

return M
