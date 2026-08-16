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
    local effective = context.effective_player_count
    local nearby = context.nearby_party_member_count
    assert(
        type(effective) == "number" and effective >= 1 and effective <= 8 and effective == math.floor(effective),
        "effective player count must be an integer from 1 through 8"
    )
    assert(
        type(nearby) == "number" and nearby >= 0 and nearby <= 7 and nearby == math.floor(nearby),
        "nearby party member count must be an integer from 0 through 7"
    )
    -- LoD 1.14d counts each additional effective game player as one half and
    -- each additional nearby party member as another half, flooring the sum.
    return math.floor(1 + (effective - 1) / 2 + nearby / 2)
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
    local no_drop = adjusted_no_drop(class, context)
    local total = weighted_total(class) - class.no_drop + no_drop
    assert(total > 0, "treasure class has no weighted outcomes")

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
    local all = classes()
    local class = assert(all[name], "unknown treasure class " .. tostring(name))
    assert(#path < MAX_DEPTH, "maximum treasure-class depth exceeded")
    assert(not active[name], "treasure-class cycle at " .. name)

    active[name] = true
    path = copied_list(path)
    path[#path + 1] = name
    local effective_quality = stronger_quality(inherited_quality, class.quality)

    for _, entry in ipairs(chosen_entries(class, context)) do
        if all[entry.code] then
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
