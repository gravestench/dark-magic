-- Roll item quality using ItemRatio.txt.
--
-- Diablo checks qualities from rarest to most ordinary: unique, set, rare,
-- magic, superior, normal, then low. The table stores denominators in 128ths;
-- smaller denominators mean better odds. All arithmetic stays integral.

local records = require("engine.records/v1")
local random = require("engine.authority_random/v1")

local M = {}

local SCALE = 128

local checks = {
    { quality = "unique", column = "Unique", diminishing = 250, magic_find = true },
    { quality = "set", column = "Set", diminishing = 500, magic_find = true },
    { quality = "rare", column = "Rare", diminishing = 600, magic_find = true },
    { quality = "magic", column = "Magic", diminishing = 0, magic_find = true },
    { quality = "superior", column = "HiQuality", diminishing = 0, magic_find = false },
    { quality = "normal", column = "Normal", diminishing = 0, magic_find = false },
}

local function integer(row, key, fallback)
    return math.floor(tonumber(row[key]) or fallback or 0)
end

local function matching_ratio(version, uber, class_specific)
    local rows = records.load("data/global/excel/itemratio.txt")
    for _, row in ipairs(rows) do
        local row_uber = integer(row, "Uber") ~= 0
        local row_class = integer(row, "Class Specific") ~= 0
        if integer(row, "Version") == version
            and row_uber == uber
            and row_class == class_specific then
            return row
        end
    end
    return nil
end

local function effective_magic_find(magic_find, diminishing)
    if diminishing <= 0 then return magic_find end
    return math.floor(magic_find * diminishing / (magic_find + diminishing))
end

local function level_adjusted(row, column, monster_level, item_level)
    local base = integer(row, column)
    assert(base > 0, "invalid ItemRatio " .. column)

    local divisor = integer(row, column .. "Divisor")
    local adjustment = 0
    if divisor > 0 then
        adjustment = math.floor((monster_level - item_level) / divisor)
    end
    return math.max((base - adjustment) * SCALE, SCALE)
end

local function apply_magic_find(value, magic_find, diminishing)
    if magic_find <= 0 then return value end
    local effective = effective_magic_find(magic_find, diminishing)
    return math.floor(value * 100 / (100 + effective))
end

local function apply_minimum(row, column, value)
    local minimum = integer(row, column .. "Min")
    if minimum > 0 then return math.max(value, minimum) end
    return value
end

local function apply_treasure_modifier(value, modifier)
    return value - math.floor(value * modifier / 1024)
end

local function denominator(row, check, item, context, modifier)
    local value = level_adjusted(
        row,
        check.column,
        context.monster_level,
        item.level
    )
    if check.magic_find then
        value = apply_magic_find(
            value,
            context.magic_find or 0,
            check.diminishing
        )
    end
    value = apply_minimum(row, check.column, value)
    value = apply_treasure_modifier(value, modifier or 0)
    return math.max(value, SCALE)
end

function M.roll(item, context, modifiers)
    modifiers = modifiers or {
        unique = 0,
        set = 0,
        rare = 0,
        magic = 0,
    }

    local row = assert(matching_ratio(
        context.version or 100,
        item.uber or false,
        item.class_specific or false
    ), "missing ItemRatio row")

    for _, check in ipairs(checks) do
        local chance = denominator(
            row,
            check,
            item,
            context,
            modifiers[check.quality]
        )
        if random.integer("d2legacy.loot.quality", chance) < SCALE then
            return check.quality
        end
    end
    return "low"
end

return M
