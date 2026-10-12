-- Resolve TreasureClassEx recursively into terminal item codes.
--
-- Positive Picks performs weighted draws that include NoDrop. Negative Picks
-- means the row lists guaranteed copies. Recursion is deliberately bounded and
-- cycle checked so malformed mod data fails a tick instead of hanging a server.

local records = require("engine.records/v1")
local random = require("engine.authority_random/v1")

local M = {}

local MAX_DEPTH = 64
local catalog
local item_types
local base_rows

local function integer(row, key, fallback)
    return math.floor(tonumber(row[key]) or fallback or 0)
end

local function copied_list(values)
    local result = {}
    for index, value in ipairs(values) do
        result[index] = value
    end
    return result
end

local function copied_quality(value)
    return {
        unique = value.unique,
        set = value.set,
        rare = value.rare,
        magic = value.magic,
    }
end

local function stronger_quality(parent, child)
    local result = copied_quality(parent)
    for _, name in ipairs({ "unique", "set", "rare", "magic" }) do
        local candidate = child[name] or 0
        if candidate > (result[name] or 0) then
            result[name] = candidate
        end
    end
    return result
end

local function read_entries(row)
    local result = {}
    for index = 1, 10 do
        local code = row["Item" .. index]
        if code and code ~= "" then
            result[#result + 1] = {
                code = code,
                weight = integer(row, "Prob" .. index),
            }
        end
    end
    return result
end

local function read_class(row)
    return {
        name = row["Treasure Class"],
        picks = integer(row, "Picks"),
        no_drop = integer(row, "NoDrop"),
        entries = read_entries(row),
        quality = {
            unique = integer(row, "Unique"),
            set = integer(row, "Set"),
            rare = integer(row, "Rare"),
            magic = integer(row, "Magic"),
        },
    }
end

local function classes()
    if catalog then
        return catalog
    end
    catalog = {}

    local rows = records.load("data/global/excel/treasureclassex.txt")
    for _, row in ipairs(rows) do
        local name = row["Treasure Class"]
        if name and name ~= "" then
            assert(not catalog[name], "duplicate treasure class " .. name)
            catalog[name] = read_class(row)
        end
    end
    return catalog
end

local function implicit_class(name)
    local kind, tier = string.match(name, "^(%a+)(%d+)$")
    tier = tonumber(tier)
    if not tier or tier < 3 or tier > 96 or tier % 3 ~= 0 then return nil end
    if not item_types then
        local rows = assert(records.load("data/global/excel/ItemTypes.txt"))
        item_types = {}
        for _, row in ipairs(rows) do
            local code = row.Code or row.code
            if code and code ~= "" then item_types[code] = row end
        end
    end
    local family = item_types[kind]
    if not family or integer(family, "TreasureClass") == 0 then return nil end

    local function matches(code, wanted, active)
        if not code or code == "" then return false end
        if code == wanted then return true end
        assert(not active[code], "cyclic ItemTypes inheritance at " .. code)
        local row = item_types[code]
        if not row then return false end
        active[code] = true
        local found = matches(row.Equiv1, wanted, active) or matches(row.Equiv2, wanted, active)
        active[code] = nil
        return found
    end
    if not base_rows then
        local rows = {}
        for _, file in ipairs({ "weapons.txt", "armor.txt", "misc.txt" }) do
            for _, row in ipairs(assert(records.load("data/global/excel/" .. file))) do
                rows[#rows + 1] = row
            end
        end
        base_rows = rows
    end

    -- Recovered 1.10f DATATBLS_CreateItemTypeTreasureClasses groups three qlevels
    -- and weights by the primary item's type rarity. Target rules require LoD.
    local class = { name = name, picks = 1, no_drop = 0, entries = {}, quality = {} }
    for _, row in ipairs(base_rows) do
        local level = integer(row, "level")
        local member = matches(row.type, kind, {}) or matches(row.type2, kind, {})
        local missile_potion = matches(row.type, "mpot", {}) or matches(row.type2, "mpot", {})
        if integer(row, "spawnable") ~= 0 and integer(row, "quest") == 0
            and row.code and row.code ~= ""
            and level > tier - 3 and level <= tier and member
            and (kind == "mpot" or not missile_potion) then
            class.entries[#class.entries + 1] = {
                code = row.code,
                weight = math.max(1, integer(item_types[row.type] or {}, "Rarity")),
            }
        end
    end
    return class
end

local function class_for(name)
    local all = classes()
    if not all[name] then all[name] = implicit_class(name) end
    return all[name]
end

local function guaranteed_entries(class)
    local result = {}
    local remaining = -class.picks

    for _, entry in ipairs(class.entries) do
        for _ = 1, entry.weight do
            if remaining <= 0 then
                return result
            end
            result[#result + 1] = entry
            remaining = remaining - 1
        end
    end
    return result
end

local function weighted_total(class)
    local total = class.no_drop
    for _, entry in ipairs(class.entries) do
        assert(entry.weight >= 0, "negative treasure weight")
        total = total + entry.weight
    end
    return total
end

local function no_drop_exponent(context)
    assert(type(context) == "table", "NoDrop player context is required")
    local count = context.no_drop_player_count
    assert(
        type(count) == "number" and count >= 1 and count <= 8 and count == math.floor(count),
        "NoDrop player count must be an integer from 1 through 8"
    )
    return count
end

local function adjusted_no_drop(class, context)
    if class.no_drop <= 0 then
        return 0
    end
    local item_total = weighted_total(class) - class.no_drop
    if item_total <= 0 then
        return class.no_drop
    end

    local chance = class.no_drop / (class.no_drop + item_total)
    local adjusted_chance = chance ^ no_drop_exponent(context)
    return math.floor(item_total * adjusted_chance / (1 - adjusted_chance))
end

local function weighted_entry(class, total, no_drop)
    local roll = random.integer("d2legacy.loot.treasure_class", total)
    if roll < no_drop then
        return nil
    end
    roll = roll - no_drop

    for _, entry in ipairs(class.entries) do
        if roll < entry.weight then
            return entry
        end
        roll = roll - entry.weight
    end
    return nil
end

local function random_entries(class, context)
    -- Native picks stop when no item has weight, even with NoDrop set, so an
    -- empty generated family such as 1.14d bow15 drops nothing.
    local item_total = weighted_total(class) - class.no_drop
    if item_total <= 0 then
        return {}
    end
    local no_drop = adjusted_no_drop(class, context)
    local total = item_total + no_drop

    local result = {}
    for _ = 1, class.picks do
        local entry = weighted_entry(class, total, no_drop)
        if entry then
            result[#result + 1] = entry
        end
    end
    return result
end

local function chosen_entries(class, context)
    assert(class.picks ~= 0, "treasure class has zero Picks: " .. class.name)
    if class.picks < 0 then
        return guaranteed_entries(class)
    end
    return random_entries(class, context)
end

local function append_drop(drops, entry, path, quality)
    drops[#drops + 1] = {
        code = entry.code,
        path = copied_list(path),
        quality = copied_quality(quality),
    }
end

local function expand(name, path, inherited_quality, active, drops, context)
    local class = assert(class_for(name), "unknown treasure class " .. tostring(name))
    assert(#path < MAX_DEPTH, "maximum treasure-class depth exceeded")
    assert(not active[name], "treasure-class cycle at " .. name)

    active[name] = true
    path = copied_list(path)
    path[#path + 1] = name
    local effective_quality = stronger_quality(inherited_quality, class.quality)

    for _, entry in ipairs(chosen_entries(class, context)) do
        if class_for(entry.code) then
            expand(entry.code, path, effective_quality, active, drops, context)
        else
            append_drop(drops, entry, path, effective_quality)
        end
    end
    active[name] = nil
end

function M.adjusted_no_drop(no_drop, item_total, context)
    return adjusted_no_drop({ no_drop = no_drop, entries = { { weight = item_total } } }, context)
end

function M.roll(name, context)
    if not name or name == "" then
        return {}
    end
    local drops = {}
    local no_quality = { unique = 0, set = 0, rare = 0, magic = 0 }
    expand(name, {}, no_quality, {}, drops, context)
    return drops
end

return M
