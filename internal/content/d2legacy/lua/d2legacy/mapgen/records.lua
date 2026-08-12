-- Small record lookup and parsing helpers for Diablo II map generation.
-- The engine exposes immutable rows exactly as authored; d2legacy decides what
-- their columns mean and reports malformed legacy data near the policy using it.

local records = require("engine.records/v1")
local M = {}

local paths = {
    levels = "data/global/excel/Levels.txt",
    presets = "data/global/excel/LvlPrest.txt",
    types = "data/global/excel/LvlTypes.txt",
    mazes = "data/global/excel/LvlMaze.txt",
}

function M.integer(value, fallback)
    local parsed = tonumber(value)
    if parsed == nil then return fallback end
    return math.floor(parsed)
end

local function find(rows, column, wanted)
    local text = tostring(wanted)
    for index, row in ipairs(rows) do
        if row[column] == text then return row, index end
    end
end

function M.level(level_id)
    return find(assert(records.load(paths.levels)), "Id", level_id)
end

function M.preset_for_level(level_id)
    return find(assert(records.load(paths.presets)), "LevelId", level_id)
end

function M.preset_by_definition(definition)
    return find(assert(records.load(paths.presets)), "Def", definition)
end

function M.maze_for_level(level_id)
    return find(assert(records.load(paths.mazes)), "Level", level_id)
end

-- LvlTypes has no explicit ID. Diablo uses its zero-based row position, while
-- Lua arrays begin at one; the authored null row therefore remains element one.
function M.level_type(level_type_id)
    return assert(records.load(paths.types))[level_type_id + 1]
end

return M
