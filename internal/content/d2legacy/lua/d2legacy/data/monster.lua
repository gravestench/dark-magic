-- Join MonStats, MonStats2, and MonLvl into one monster definition.
--
-- The original spreadsheets were split because the legacy table tooling had a
-- column limit. They describe one conceptual monster. Go safely decodes rows;
-- this mod owns their Diablo-specific relationship and difficulty formulas.

local records = require("engine.records/v1")
local game_rules = require("d2legacy.policy.game_rules")

local M = {}

local RAW_SCALE = 256
local component_names = {
    "HD",
    "TR",
    "LG",
    "RA",
    "LA",
    "RH",
    "LH",
    "SH",
    "S1",
    "S2",
    "S3",
    "S4",
    "S5",
    "S6",
    "S7",
    "S8",
}

-- Record capabilities return defensive copies. Build the three immutable
-- lookup tables once per authoritative VM instead of copying and scanning the
-- complete legacy tables for every monster considered during population.
local catalog

local function index_by(rows, key)
    local result = {}
    for _, row in ipairs(rows) do
        local value = row[key]
        if value and value ~= "" and result[value] == nil then
            result[value] = row
        end
    end
    return result
end

local function monster_catalog()
    if catalog then
        return catalog
    end
    catalog = {
        stats = index_by(records.load("data/global/excel/monstats.txt"), "Id"),
        graphics = index_by(records.load("data/global/excel/monstats2.txt"), "Id"),
        levels = index_by(records.load("data/global/excel/monlvl.txt"), "Level"),
    }
    return catalog
end

local function integer(row, key, fallback)
    local value = tonumber(row[key])
    if value == nil then
        return fallback or 0
    end
    return math.floor(value)
end

local function truth(row, key)
    local value = string.lower(row[key] or "")
    return value == "1" or value == "true"
end

local function find(rows, key, wanted)
    for _, row in ipairs(rows) do
        if row[key] == wanted then
            return row
        end
    end
    return nil
end

local function difficulty_value(row, columns, difficulty)
    if difficulty == 2 then
        return integer(row, columns.hell)
    end
    if difficulty == 1 then
        return integer(row, columns.nightmare)
    end
    return integer(row, columns.normal)
end

local function columns(normal)
    return {
        normal = normal,
        nightmare = normal .. "(N)",
        hell = normal .. "(H)",
    }
end

local function ratio(base, percentage)
    return math.floor(base * percentage / 100)
end

local function graphics_components(graphics)
    local result = {}
    for _, name in ipairs(component_names) do
        local variant = graphics[name .. "v"]
        if truth(graphics, name) and variant and variant ~= "" then
            result[name] = string.upper(variant)
        end
    end
    return result
end

local function raw_stats(stats, difficulty)
    return {
        life_min = difficulty_value(stats, columns("minHP"), difficulty),
        life_max = difficulty_value(stats, columns("maxHP"), difficulty),
        defense = difficulty_value(stats, columns("AC"), difficulty),
        attack = difficulty_value(stats, columns("A1TH"), difficulty),
        damage_min = difficulty_value(stats, columns("A1MinD"), difficulty),
        damage_max = difficulty_value(stats, columns("A1MaxD"), difficulty),
        experience = difficulty_value(stats, columns("Exp"), difficulty),
    }
end

local function scaled_stats(stats, level, difficulty)
    if truth(stats, "noRatio") or not level then
        return raw_stats(stats, difficulty)
    end

    local percentages = raw_stats(stats, difficulty)
    return {
        life_min = ratio(difficulty_value(level, columns("HP"), difficulty), percentages.life_min),
        life_max = ratio(difficulty_value(level, columns("HP"), difficulty), percentages.life_max),
        defense = ratio(difficulty_value(level, columns("AC"), difficulty), percentages.defense),
        attack = ratio(difficulty_value(level, columns("TH"), difficulty), percentages.attack),
        damage_min = ratio(difficulty_value(level, columns("DM"), difficulty), percentages.damage_min),
        damage_max = ratio(difficulty_value(level, columns("DM"), difficulty), percentages.damage_max),
        experience = ratio(difficulty_value(level, columns("XP"), difficulty), percentages.experience),
    }
end

local function treasure_class(stats, difficulty)
    if difficulty == 2 then
        return stats["TreasureClass1(H)"] or ""
    end
    if difficulty == 1 then
        return stats["TreasureClass1(N)"] or ""
    end
    return stats.TreasureClass1 or ""
end

local function aggro_radius(stats)
    local radius = integer(stats, "aidist", 35)
    if radius > 0 then
        return radius
    end
    return 35
end

local function graphics_id(stats)
    if stats.MonStatsEx and stats.MonStatsEx ~= "" then
        return stats.MonStatsEx
    end
    return stats.Id
end

local function runtime_definition(stats, graphics, values, difficulty)
    local size = math.max(integer(graphics, "SizeX", 2), integer(graphics, "SizeY", 2)) / 2
    return {
        id = stats.Id,
        base_id = stats.BaseId or "",
        graphics_id = graphics_id(stats),
        name_key = stats.NameStr or stats.Id,
        ai = stats.AI or "",
        token = string.upper(stats.Code or ""),
        weapon_class = string.upper(graphics.BaseW or "HTH"),
        components = graphics_components(graphics),
        life_min = values.life_min * RAW_SCALE,
        life_max = values.life_max * RAW_SCALE,
        level = integer(stats, "Level", 1),
        defense = values.defense,
        attack_rating = values.attack,
        physical_min = values.damage_min * RAW_SCALE,
        physical_max = values.damage_max * RAW_SCALE,
        experience = values.experience,
        treasure_class = treasure_class(stats, difficulty),
        collider_radius = size,
        select_radius = size,
        velocity = integer(stats, "Velocity", 5),
        think_interval = math.max(integer(stats, "aidel", 1), 1),
        aggro_radius = aggro_radius(stats),
        attack_range = math.max(integer(graphics, "MeleeRng", 1), 1),
        min_group = math.max(integer(stats, "MinGrp", 1), 1),
        max_group = math.max(integer(stats, "MaxGrp", 1), 1),
        rarity = math.max(integer(stats, "Rarity", 1), 1),
    }
end

function M.load(id, ...)
    assert(select("#", ...) == 0, "monster difficulty comes from immutable game rules")
    local difficulty = game_rules.difficulty()
    local indexed = monster_catalog()
    local stats = assert(indexed.stats[id], "missing MonStats row")
    assert(
        truth(stats, "enabled") and truth(stats, "isSpawn") and not truth(stats, "npc"),
        "monster is not an ordinary spawn"
    )

    local graphics = assert(indexed.graphics[graphics_id(stats)], "missing MonStats2 row")
    local level = indexed.levels[stats.Level]
    return runtime_definition(stats, graphics, scaled_stats(stats, level, difficulty), difficulty)
end

local function encoded_components(values)
    local result = {}
    for key, component in pairs(values) do
        result[#result + 1] = key .. "=" .. component
    end
    table.sort(result)
    return table.concat(result, ",")
end

function M.all(...)
    assert(select("#", ...) == 0, "monster difficulty comes from immutable game rules")
    local result = {}
    for _, row in ipairs(records.load("data/global/excel/monstats.txt")) do
        local valid, value = pcall(M.load, row.Id)
        if valid and next(value.components) then
            value.components = encoded_components(value.components)
            result[#result + 1] = value
        end
    end
    table.sort(result, function(left, right)
        return left.id < right.id
    end)
    return result
end

return M
