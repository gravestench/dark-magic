-- Compose the first playable d2legacy world without leaking level identities
-- into the generic host. Rogue Encampment and Blood Moor are mod choices; the
-- Go adapter transports admitted recipes and resolves their authored exits.

local preset = require("d2legacy.mapgen.preset")
local outdoor = require("d2legacy.mapgen.outdoor")
local maze = require("d2legacy.mapgen.maze")
local M = {}

function M.dungeon(seed, difficulty)
    return maze.generate(8, seed, difficulty)
end

function M.dungeon_connection(wilderness, dungeon)
    local function stamp_id(zone, role)
        for _, stamp in ipairs(zone.stamps) do
            if stamp.role == role then return stamp.id end
        end
        error("d2legacy entry world has no " .. role)
    end
    return {
        first_level = wilderness.request.level_id, second_level = dungeon.request.level_id,
        first_stamp = stamp_id(wilderness, "den-of-evil-entry"),
        second_stamp = stamp_id(dungeon, "previous-level"),
    }
end

function M.town(seed, difficulty)
    return preset.generate(1, seed, difficulty)
end

function M.wilderness(seed, difficulty, town)
    local stamp = assert(town.stamps and town.stamps[1], "d2legacy entry town has no stamp")
    local direction = assert(stamp.role:match("^act1%-town:exit%-(%a+)$"),
        "d2legacy entry town has no cardinal exit")
    return outdoor.generate(2, seed, direction, difficulty, true)
end

-- Describe how the two recipes meet. The values here are policy: which levels
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

    local opposite = {north = "south", east = "west", south = "north", west = "east"}
    assert(opposite[town_direction] == entry.direction,
        "d2legacy entry-world edges do not face one another")

    return {
        first_level = town.request.level_id,
        first_direction = town_direction,
        second_level = wilderness.request.level_id,
        second_direction = entry.direction,
        second_tile_x = entry.x,
        second_tile_y = entry.y,
    }
end

return M
