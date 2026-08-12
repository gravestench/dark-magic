-- Join a terminal treasure code to its immutable base-item record.
--
-- Go decodes rows and protects file access. This mod decides that weapons,
-- armor, and misc records form the Diablo base-item namespace and which fields
-- become generation and presentation facts.

local records = require("engine.records/v1")

local M = {}

local sources = {
    { path = "weapons.txt", kind = "weapon" },
    { path = "armor.txt", kind = "armor" },
    { path = "misc.txt", kind = "misc" },
}

local function integer(row, key, fallback)
    return math.floor(tonumber(row[key]) or fallback or 0)
end

local function converted(row, kind)
    return {
        code = row.code,
        kind = kind,
        name_key = row.namestr or "",
        type = row.type or "",
        type2 = row.type2 or "",
        level = integer(row, "level"),
        level_requirement = integer(row, "levelreq"),
        magic_level = integer(row, "magic lvl"),
        inventory_file = row.invfile or "",
        world_file = row.flippyfile or "",
        width = integer(row, "invwidth", 1),
        height = integer(row, "invheight", 1),
        base_cost = integer(row, "cost"),
        uber = (row.ubercode or "") ~= "",
        class_specific = (row.type or "") == "class",
    }
end

local function find_in(source, code)
    local rows = records.load("data/global/excel/" .. source.path)
    for _, row in ipairs(rows) do
        if row.code == code then return converted(row, source.kind) end
    end
    return nil
end

function M.base(code)
    for _, source in ipairs(sources) do
        local item = find_in(source, code)
        if item then return item end
    end
    return nil
end

return M
