-- Select and materialize MagicPrefix and MagicSuffix records.
--
-- Eligibility, frequency, group exclusion, and property ranges are Diablo item
-- policy. Three separate named RNG streams make slot choice, weighted record
-- choice, and authored value rolls independently replayable.

local records = require("engine.records/v1")
local random = require("engine.authority_random/v1")

local M = {}

local function integer(row, key, fallback)
    return math.floor(tonumber(row[key]) or fallback or 0)
end

local function truth(row, key)
    local value = string.lower(row[key] or "")
    return value == "1" or value == "true"
end

local function contains_name(values, wanted)
    for _, value in ipairs(values) do
        if value.name == wanted then return true end
    end
    return false
end

local function type_values(row, prefix)
    local result = {}
    for index = 1, 7 do
        local value = row[prefix .. index]
        if value and value ~= "" then result[#result + 1] = value end
    end
    return result
end

local function item_matches(item, wanted)
    return item.type == wanted or item.type2 == wanted
end

local function excluded(row, item)
    for _, value in ipairs(type_values(row, "etype")) do
        if item_matches(item, value) then return true end
    end
    return false
end

local function included(row, item)
    for _, value in ipairs(type_values(row, "itype")) do
        if item_matches(item, value) then return true end
    end
    return false
end

local function eligible(row, item, options, used_groups)
    if not truth(row, "spawnable") then return false end
    if integer(row, "frequency") <= 0 then return false end
    if integer(row, "level") > options.level then return false end

    local maximum = integer(row, "maxlevel")
    if maximum > 0 and options.level > maximum then return false end
    if options.version < 100 and integer(row, "version") >= 100 then return false end
    if options.quality == "rare" and not truth(row, "rare") then return false end

    local group = integer(row, "group")
    if group ~= 0 and used_groups[group] then return false end
    if excluded(row, item) then return false end
    return included(row, item)
end

local function rolled_modifier(row, index)
    local code = row["mod" .. index .. "code"]
    if not code or code == "" then return nil end

    local minimum = integer(row, "mod" .. index .. "min")
    local maximum = integer(row, "mod" .. index .. "max")
    if minimum > maximum then minimum, maximum = maximum, minimum end

    return {
        code = code,
        parameter = integer(row, "mod" .. index .. "param"),
        minimum = minimum,
        maximum = maximum,
        value = minimum + random.integer(
            "d2legacy.loot.affix_value",
            maximum - minimum + 1
        ),
    }
end

local function materialized(row, kind)
    local result = {
        name = row.Name,
        kind = kind,
        group = integer(row, "group"),
        level_requirement = integer(row, "levelreq"),
        modifiers = {},
    }
    for index = 1, 3 do
        local modifier = rolled_modifier(row, index)
        if modifier then result.modifiers[#result.modifiers + 1] = modifier end
    end
    return result
end

local function candidate_rows(table_name, item, options, used_groups, selected)
    local result = {}
    local total = 0
    local rows = records.load("data/global/excel/" .. table_name)

    for _, row in ipairs(rows) do
        if eligible(row, item, options, used_groups)
            and not contains_name(selected, row.Name) then
            result[#result + 1] = row
            total = total + integer(row, "frequency")
        end
    end
    table.sort(result, function(left, right) return left.Name < right.Name end)
    return result, total
end

local function weighted_candidate(candidates, total)
    local roll = random.integer("d2legacy.loot.affix_choice", total)
    for _, candidate in ipairs(candidates) do
        local weight = integer(candidate, "frequency")
        if roll < weight then return candidate end
        roll = roll - weight
    end
    return nil
end

local function choose_kind(arguments)
    for _ = 1, arguments.limit do
        if arguments.total[1] >= arguments.options.max_total then return end

        -- Each available prefix/suffix slot first has the legacy 50% chance to
        -- remain empty. This draw happens before building the weighted pool.
        if random.integer("d2legacy.loot.affix_slot", 2) == 1 then
            local candidates, total = candidate_rows(
                arguments.table_name,
                arguments.item,
                arguments.options,
                arguments.used_groups,
                arguments.selected
            )
            if total > 0 then
                local row = weighted_candidate(candidates, total)
                local value = materialized(row, arguments.kind)
                arguments.selected[#arguments.selected + 1] = value
                arguments.total[1] = arguments.total[1] + 1
                if value.group ~= 0 then
                    arguments.used_groups[value.group] = true
                end
            end
        end
    end
end

function M.roll(item, quality, level, version)
    if quality ~= "magic" and quality ~= "rare" then return {}, {} end

    local rare = quality == "rare"
    local options = {
        quality = quality,
        level = level,
        version = version or 100,
        max_total = rare and 6 or 2,
    }
    local prefixes = {}
    local suffixes = {}
    local used_groups = {}
    local total = { 0 }

    choose_kind({
        table_name = "magicprefix.txt",
        kind = "prefix",
        item = item,
        options = options,
        limit = rare and 3 or 1,
        used_groups = used_groups,
        selected = prefixes,
        total = total,
    })
    choose_kind({
        table_name = "magicsuffix.txt",
        kind = "suffix",
        item = item,
        options = options,
        limit = rare and 3 or 1,
        used_groups = used_groups,
        selected = suffixes,
        total = total,
    })
    return prefixes, suffixes
end

return M
