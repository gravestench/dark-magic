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

local function asset(name)
    return name and name ~= "" and "data/global/items/" .. name .. ".dc6" or ""
end

local function body_slots(row, kind)
    if kind == "weapon" then return "rarm,larm" end
    for _, item_type in ipairs(records.load("data/global/excel/ItemTypes.txt")) do
        if (item_type.Code or item_type.code) == row.type then
            local result = {}
            for _, field in ipairs({"BodyLoc1", "BodyLoc2"}) do
                local slot = item_type[field]
                if slot and slot ~= "" then result[#result + 1] = slot end
            end
            return table.concat(result, ",")
        end
    end
    return ""
end

local function composite(row)
    local tokens = {"HD","TR","LG","RA","LA","RH","LH","SH","S1","S2","S3","S4","S5","S6","S7","S8"}
    local token = tokens[integer(row, "component", -1) + 1]
    local appearance = string.upper(row.alternategfx or "")
    return token and appearance ~= "" and token .. "=" .. appearance or ""
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
        material_flags = integer(row, "bitfield1"),
        body_slots = body_slots(row, kind),
        belt_eligible = integer(row, "belt") ~= 0,
        inventory_dc6 = asset(row.invfile),
        world_dc6 = asset(row.flippyfile),
        world_animated = true,
        composite = composite(row),
        weapon_class = string.upper(row.wclass or ""),
        melee_weapon_class = string.upper(row.wclass or ""),
        melee_range = kind == "weapon" and 1 + integer(row, "rangeadder") or 0,
        physical_min = integer(row, "mindam") * 256,
        physical_max = integer(row, "maxdam") * 256,
        attack_rate = kind == "weapon" and -integer(row, "speed") or 0,
        defense = integer(row, "minac"),
        base_defense_max = integer(row, "maxac"),
        speed_penalty = kind == "armor" and integer(row, "speed") or 0,
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
