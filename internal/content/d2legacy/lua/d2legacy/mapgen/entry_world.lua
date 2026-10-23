-- Compose the first three playable d2legacy zones without leaking level
-- identities into the generic host. Lua selects recipes and seam roles; Go
-- transports and validates the admitted topology.

local preset = require("d2legacy.mapgen.preset")
local outdoor = require("d2legacy.mapgen.outdoor")
local cold_plains = require("d2legacy.mapgen.cold_plains")
local M = {}
local opposite = {north = "south", east = "west", south = "north", west = "east"}

-- A seam must be anchored to exactly one authored warp with the intended destination.
local function named_warp(zone, role, target)
    local selected
    for _, warp in ipairs(zone.warps or {}) do
        if warp.role == role then
            assert(selected == nil, "d2legacy entry world has duplicate " .. role)
            selected = warp
        end
    end
    assert(selected, "d2legacy entry world has no " .. role)
    assert(selected.destination_level == target,
        "d2legacy entry world " .. role .. " targets the wrong level")
    assert(opposite[selected.direction], "d2legacy entry world has non-cardinal " .. role)
    return selected
end

-- These geometry values are provisional Dark Magic policy until source behavior is established.
local function seam_geometry()
    return {trigger_radius = 2, max_arrival_distance = 10, arrival_inset = 6}
end

function M.town(seed, difficulty)
    return preset.generate(1, seed, difficulty)
end

function M.wilderness(seed, difficulty, town)
    local stamp = assert(town.stamps and town.stamps[1], "d2legacy entry town has no stamp")
    local direction = assert(stamp.role:match("^act1%-town:exit%-(%a+)$"),
        "d2legacy entry town has no cardinal exit")
    return outdoor.generate(2, seed, direction, difficulty)
end

-- The admitted Blood Moor exit selects the Cold Plains approach direction.
function M.cold_plains(seed, difficulty, blood_moor)
    assert(blood_moor.request and blood_moor.request.level_id == 2,
        "d2legacy entry world requires an admitted Blood Moor")
    local exit = named_warp(blood_moor, "next-level-exit", 3)
    return cold_plains.generate(3, seed, exit.direction, difficulty)
end

-- Describe how the town and wilderness recipes meet. The values here are policy: which levels
-- connect, which authored roles identify the connection, and which edges face
-- one another. Go receives only this finished specification and resolves safe
-- collision-space points in the materialized maps.
function M.seam(town, wilderness)
    local town_stamp = assert(town.stamps and town.stamps[1], "d2legacy entry town has no stamp")
    local town_direction = assert(town_stamp.role:match("^act1%-town:exit%-(%a+)$"),
        "d2legacy entry town has no cardinal exit")

    local entry
    for _, warp in ipairs(wilderness.warps or {}) do
        if warp.role == "town-entry" then
            assert(entry == nil, "d2legacy entry wilderness has duplicate town entries")
            entry = warp
        end
    end
    assert(entry, "d2legacy entry wilderness has no town entry")
    assert(entry.destination_level == town.request.level_id,
        "d2legacy entry wilderness targets the wrong level")

    assert(opposite[town_direction] == entry.direction,
        "d2legacy entry-world edges do not face one another")

    local seam = seam_geometry()
    seam.first_level = town.request.level_id
    seam.first_direction = town_direction
    seam.second_level = wilderness.request.level_id
    seam.second_direction = entry.direction
    seam.second_tile_x = entry.x
    seam.second_tile_y = entry.y
    seam.second_warp_id = entry.id
    seam.second_warp_role = entry.role
    return seam
end

-- Bind both generated warp identities so restored topology cannot silently drift from the admitted recipes.
function M.generated_seam(blood_moor, plains)
    assert(blood_moor.request and blood_moor.request.level_id == 2,
        "d2legacy entry world requires level 2 first")
    assert(plains.request and plains.request.level_id == 3,
        "d2legacy entry world requires level 3 second")
    local exit = named_warp(blood_moor, "next-level-exit", 3)
    local entry = named_warp(plains, "blood-moor-entry", 2)
    assert(opposite[exit.direction] == entry.direction,
        "d2legacy entry world generated edges do not face one another")
    local seam = seam_geometry()
    seam.first_generated = true
    seam.first_level = 2
    seam.first_direction = exit.direction
    seam.first_tile_x = exit.x
    seam.first_tile_y = exit.y
    seam.first_warp_id = exit.id
    seam.first_warp_role = exit.role
    seam.second_level = 3
    seam.second_direction = entry.direction
    seam.second_tile_x = entry.x
    seam.second_tile_y = entry.y
    seam.second_warp_id = entry.id
    seam.second_warp_role = entry.role
    return seam
end

return M
