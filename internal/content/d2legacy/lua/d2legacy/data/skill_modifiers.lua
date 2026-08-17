-- Decode reviewed Skills.txt cross-skill modifier shapes into exact IDs.
--
-- This is intentionally not a general legacy expression evaluator. Each
-- admitted shape is small, fail-closed, and reusable by behavior families.

local M = {}

local function required_integer(row, column, label)
    local value = tonumber(row[column])
    assert(value and value >= 0 and value == math.floor(value), label .. " has invalid " .. column)
    return value
end

function M.by_name(rows)
    local result = {}
    for _, row in ipairs(rows) do
        if row.skill and row.skill ~= "" then
            result[string.lower(row.skill)] = row
        end
    end
    return result
end

function M.hard_level_sum_percent(row, expression_column, parameter_column, skills_by_name, label)
    local expression = row[expression_column] or ""
    if expression == "" then
        return {}, 0
    end
    local names = {}
    for name in string.gmatch(expression, "skill%('([^']+)'%.blvl%)") do
        names[#names + 1] = name
    end
    assert(#names > 0, label .. " has an unsupported " .. expression_column)
    local terms = {}
    local ids = {}
    for _, name in ipairs(names) do
        terms[#terms + 1] = "skill('" .. name .. "'.blvl)"
        local synergy = assert(skills_by_name[string.lower(name)], label .. " has an unknown synergy " .. name)
        ids[#ids + 1] = required_integer(synergy, "Id", label)
    end
    local parameter_number = assert(string.match(parameter_column, "^Param(%d+)$"), "unsupported parameter column")
    local supported = "(" .. table.concat(terms, "+") .. ")*par" .. parameter_number
    assert(
        expression == supported,
        label .. " has an unsupported " .. expression_column .. " shape " .. expression .. " != " .. supported
    )
    return ids, required_integer(row, parameter_column, label)
end

function M.single_hard_level_percent_multiplier(
    row,
    expression_column,
    base_token,
    parameter_column,
    skills_by_name,
    label
)
    local expression = row[expression_column] or ""
    local name = string.match(expression, "skill%('([^']+)'%.blvl%)")
    assert(name, label .. " has an unsupported " .. expression_column)
    local parameter_number = assert(string.match(parameter_column, "^Param(%d+)$"), "unsupported parameter column")
    local supported = base_token .. " * (100 + skill('" .. name .. "'.blvl) * par" .. parameter_number .. ") / 100"
    assert(
        expression == supported,
        label .. " has an unsupported " .. expression_column .. " shape " .. expression .. " != " .. supported
    )
    local synergy = assert(skills_by_name[string.lower(name)], label .. " has an unknown synergy " .. name)
    return { required_integer(synergy, "Id", label) }, required_integer(row, parameter_column, label)
end

return M
