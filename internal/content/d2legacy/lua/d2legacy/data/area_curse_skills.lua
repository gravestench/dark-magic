-- Decode explicitly admitted point-centered timed curse skills.

local records = require("engine.records/v1")
local M = {}

local function index(rows, column)
    local result = {}
    for _, row in ipairs(rows) do
        local key = row[column]
        if key and key ~= "" then
            result[key] = row
        end
    end
    return result
end

local function required_integer(row, column, label)
    local value = tonumber(row[column])
    assert(value and value >= 0 and value == math.floor(value), label .. " has invalid " .. column)
    return value
end

local function decode(row, states_by_name)
    local id = required_integer(row, "Id", "area curse skill")
    local label = row.skill or ("skill " .. id)
    assert(row.srvstfunc == "" and row.srvdofunc == "30", label .. " has unsupported server functions")
    assert(
        row.aurafilter == "3" and row.range == "none" and row.LineOfSight == "4",
        label .. " has unsupported target policy"
    )
    assert(row.aurarangecalc == "ln12" and row.auralencalc == "ln34", label .. " has unsupported area formula")
    assert(row.aurastat1 == "damageresist" and row.aurastatcalc1 == "-par5", label .. " has unsupported stat formula")
    assert(row.interrupt == "1" and row.InGame == "1", label .. " has unsupported action flags")
    local state = assert(states_by_name[row.auratargetstate], label .. " has an unknown target state")
    assert(state.curse == "1", label .. " target state is not a curse")
    return {
        behavior = "state.point-area-curse",
        skill_id = id,
        mana_cost_raw = required_integer(row, "mana", label) * (2 ^ required_integer(row, "manashift", label)),
        minimum_mana_cost_raw = required_integer(row, "minmana", label)
            * (2 ^ required_integer(row, "manashift", label)),
        effect_delay = 1,
        complete_delay = 2,
        requires_point_target = true,
        state_id = state.state,
        exclusive_group = "state-group:curse",
        reject_lower_priority = true,
        radius_base = required_integer(row, "Param1", label),
        radius_per_level = required_integer(row, "Param2", label),
        duration_base = required_integer(row, "Param3", label),
        duration_per_level = required_integer(row, "Param4", label),
        stat = "physical_resist",
        stat_operation = "add",
        stat_value = -required_integer(row, "Param5", label),
        immune_divisor = 5,
    }
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported area-curse skill IDs are required")
    local skills = index(assert(records.load("data/global/excel/skills.txt")), "Id")
    local states = index(assert(records.load("data/global/excel/states.txt")), "state")
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        if row then
            local definition = decode(row, states)
            assert(not result[definition.skill_id], "duplicate area-curse skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
