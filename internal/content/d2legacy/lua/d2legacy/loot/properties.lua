-- Interpret rolled property codes into portable stat and effect facts.
--
-- Properties.txt is a tiny instruction table. Some function numbers mean a
-- direct stat operation; others describe skills, damage, ethereal state, or a
-- specialized effect. Unknown functions remain explicit for later work rather
-- than being silently converted into a plausible but incorrect stat.

local records = require("engine.records/v1")

local M = {}

local catalog
local direct_functions = {
    [1] = true,
    [2] = true,
    [8] = true,
    [13] = true,
    [14] = true,
    [15] = true,
    [16] = true,
    [17] = true,
    [22] = true,
}

local function integer(row, key, fallback)
    return math.floor(tonumber(row[key]) or fallback or 0)
end

local function read_steps(row)
    local result = {}
    for index = 1, 7 do
        local fn = integer(row, "func" .. index)
        local stat = row["stat" .. index] or ""
        if fn ~= 0 or stat ~= "" then
            result[#result + 1] = {
                fn = fn,
                stat = stat,
                set = integer(row, "set" .. index) ~= 0,
                value = integer(row, "val" .. index),
            }
        end
    end
    return result
end

local function definitions()
    if catalog then return catalog end
    catalog = {}

    for _, row in ipairs(records.load("data/global/excel/properties.txt")) do
        if row.code and row.code ~= "" then
            catalog[row.code] = read_steps(row)
        end
    end
    return catalog
end

local function effective_function(step, previous)
    if step.fn == 3 or step.fn == 9 then return previous, previous end
    if step.fn ~= 0 then return step.fn, step.fn end
    return step.fn, previous
end

local function direct_stat(modifier, step, fn)
    return {
        code = step.stat,
        parameter = modifier.parameter,
        value = modifier.value,
        set = step.set,
        fn = fn,
    }
end

local function specialized_effect(modifier, fn)
    if fn == 5 then
        return { kind = "minimum_damage", value = modifier.value }
    elseif fn == 6 then
        return { kind = "maximum_damage", value = modifier.value }
    elseif fn == 7 then
        return { kind = "damage_percent", value = modifier.value }
    elseif fn == 10 then
        return {
            kind = "skill_tab",
            value = modifier.value,
            class = math.floor(modifier.parameter / 3),
            tab = modifier.parameter % 3,
        }
    elseif fn == 11 then
        return {
            kind = "proc",
            skill_id = modifier.parameter,
            chance = modifier.minimum,
            level = modifier.maximum,
        }
    elseif fn == 19 then
        return {
            kind = "charged_skill",
            skill_id = modifier.parameter,
            charges = modifier.minimum,
            level = modifier.maximum,
        }
    elseif fn == 20 then
        return { kind = "indestructible", value = modifier.value }
    elseif fn == 23 then
        return { kind = "ethereal", value = modifier.value }
    end
    return nil
end

local function unsupported_property(modifier, step, fn)
    return {
        property = modifier.code,
        fn = fn,
        stat = step.stat,
        parameter = modifier.parameter,
        value = modifier.value,
    }
end

local function interpret_modifier(modifier, stats, effects, unsupported)
    local steps = assert(definitions()[modifier.code],
        "unknown property " .. modifier.code)
    local previous = 0

    for _, step in ipairs(steps) do
        local fn
        fn, previous = effective_function(step, previous)
        if direct_functions[fn] then
            stats[#stats + 1] = direct_stat(modifier, step, fn)
        else
            local effect = specialized_effect(modifier, fn)
            if effect then
                effects[#effects + 1] = effect
            elseif fn ~= 0 then
                unsupported[#unsupported + 1] = unsupported_property(
                    modifier,
                    step,
                    fn
                )
            end
        end
    end
end

local function interpret_affixes(affixes, stats, effects, unsupported)
    for _, affix in ipairs(affixes) do
        for _, modifier in ipairs(affix.modifiers) do
            interpret_modifier(modifier, stats, effects, unsupported)
        end
    end
end

function M.apply(prefixes, suffixes)
    local stats = {}
    local effects = {}
    local unsupported = {}

    interpret_affixes(prefixes, stats, effects, unsupported)
    interpret_affixes(suffixes, stats, effects, unsupported)

    table.sort(stats, function(left, right)
        if left.code ~= right.code then return left.code < right.code end
        return left.parameter < right.parameter
    end)
    return stats, effects, unsupported
end

return M
