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

-- engine.records caches decoded Go rows, but each load still copies all rows
-- into fresh Lua tables. Map generation performs many indexed lookups, so keep
-- one immutable Lua table per source for this short-lived policy runtime.
local loaded = {}
local indexes = {}

local function rows(path)
    if not loaded[path] then
        loaded[path] = assert(records.load(path))
    end
    return loaded[path]
end

local function index(path, column)
    local key = path .. "\0" .. column
    if indexes[key] then
        return indexes[key]
    end
    local result = {}
    for row_index, row in ipairs(rows(path)) do
        local value = row[column]
        if value ~= nil and result[value] == nil then
            result[value] = { row = row, index = row_index }
        end
    end
    indexes[key] = result
    return result
end

function M.integer(value, fallback)
    local parsed = tonumber(value)
    if parsed == nil then
        return fallback
    end
    return math.floor(parsed)
end

local function find(path, column, wanted)
    local match = index(path, column)[tostring(wanted)]
    if match then
        return match.row, match.index
    end
end

function M.level(level_id)
    return find(paths.levels, "Id", level_id)
end

function M.preset_for_level(level_id)
    return find(paths.presets, "LevelId", level_id)
end

function M.preset_by_definition(definition)
    return find(paths.presets, "Def", definition)
end

function M.maze_for_level(level_id)
    return find(paths.mazes, "Level", level_id)
end

-- LvlTypes has no explicit ID. Diablo uses its zero-based row position, while
-- Lua arrays begin at one; the authored null row therefore remains element one.
function M.level_type(level_type_id)
    return rows(paths.types)[level_type_id + 1]
end

return M
