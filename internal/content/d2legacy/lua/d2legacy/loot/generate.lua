-- Assemble one complete authoritative loot result.
--
-- Each stage lives in its own module: treasure selection, base-item lookup,
-- quality, affixes, and property interpretation. This file only tells their
-- story in order and serializes the final durable facts for the death event.

local treasure = require("d2legacy.loot.treasure_class")
local quality = require("d2legacy.loot.quality")
local items = require("d2legacy.loot.items")
local affixes = require("d2legacy.loot.affixes")
local properties = require("d2legacy.loot.properties")

local M = {}

local function quoted(value)
    return string.format("%q", value)
end

local function sorted_keys(value)
    local keys = {}
    for key in pairs(value) do
        keys[#keys + 1] = key
    end
    table.sort(keys, function(left, right)
        return tostring(left) < tostring(right)
    end)
    return keys
end

-- Lua's standard library has no JSON encoder. This deliberately small encoder
-- accepts only the scalar/table trees produced by this module. Sorted object
-- keys keep the serialized death event stable across replay and checkpoints.
local function encode(value, force_array)
    local kind = type(value)
    if kind == "string" then
        return quoted(value)
    end
    if kind == "number" or kind == "boolean" then
        return tostring(value)
    end
    if kind ~= "table" then
        return "null"
    end

    if force_array or #value > 0 then
        local entries = {}
        for _, entry in ipairs(value) do
            entries[#entries + 1] = encode(entry)
        end
        return "[" .. table.concat(entries, ",") .. "]"
    end

    local fields = {}
    for _, key in ipairs(sorted_keys(value)) do
        fields[#fields + 1] = quoted(tostring(key)) .. ":" .. encode(value[key])
    end
    return "{" .. table.concat(fields, ",") .. "}"
end

local function unresolved_drop(drop, context)
    return {
        code = drop.code,
        quality = "unresolved",
        level = context.monster_level,
        path = table.concat(drop.path, " > "),
    }
end

local function generated_drop(drop, base, context)
    local rolled_quality = quality.roll(base, context, drop.quality)
    local prefixes, suffixes = affixes.roll(base, rolled_quality, context.monster_level, context.version)
    local stats, effects, unsupported = properties.apply(prefixes, suffixes)

    return {
        code = drop.code,
        quality = rolled_quality,
        level = context.monster_level,
        path = table.concat(drop.path, " > "),
        inventory_file = base.inventory_file,
        world_file = base.world_file,
        width = base.width,
        height = base.height,
        base_cost = base.base_cost,
        prefixes = prefixes,
        suffixes = suffixes,
        stats = stats,
        effects = effects,
        unsupported = unsupported,
    }
end

function M.roll(treasure_class, context)
    local result = {}
    assert(type(context.player_count) == "table", "loot NoDrop player context is required")
    for _, drop in ipairs(treasure.roll(treasure_class, context.player_count)) do
        local base = items.base(drop.code)
        if base then
            result[#result + 1] = generated_drop(drop, base, context)
        else
            result[#result + 1] = unresolved_drop(drop, context)
        end
    end
    return result
end

-- Loot is always a list, including when NoDrop produces no entries. Passing an
-- explicit array hint avoids serializing an empty result as the JSON object {}.
function M.encode(drops)
    return encode(drops, true)
end

return M
