-- Decode explicitly admitted right-selected party stat auras.

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
    local id = required_integer(row, "Id", "party aura skill")
    local label = row.skill or ("skill " .. id)
    assert(row.srvstfunc == "" and row.srvdofunc == "65", label .. " has unsupported server functions")
    assert(
        row.aura == "1" and row.immediate == "1" and row.leftskill == "" and row.range == "none" and row.InGame == "1",
        label .. " has unsupported activation flags"
    )
    assert(row.aurafilter == "73731", label .. " has unsupported target filter")
    assert(row.aurarangecalc == "ln12", label .. " has unsupported radius formula")
    assert(row.aurastat1 == "damagepercent" and row.aurastatcalc1 == "ln34", label .. " has unsupported stat formula")
    assert(
        required_integer(row, "mana", label) == 0 and required_integer(row, "lvlmana", label) == 0,
        label .. " unexpectedly consumes mana"
    )
    local state = assert(states_by_name[row.aurastate], label .. " has an unknown owner state")
    local target_state = assert(states_by_name[row.auratargetstate], label .. " has an unknown target state")
    assert(state.aura == "1" and state.stat == "damagepercent", label .. " owner state is not a damage aura")
    assert(target_state.aura == "1", label .. " target state is not an aura")
    return {
        behavior = "aura.selected-party-stat",
        activation = "selected_right",
        skill_id = id,
        mana_cost_raw = 0,
        minimum_mana_cost_raw = 0,
        effect_delay = 0,
        complete_delay = 0,
        state_id = state.state,
        target_state_id = target_state.state,
        radius_base = required_integer(row, "Param1", label),
        radius_per_level = required_integer(row, "Param2", label),
        stat = "damagepercent",
        stat_operation = "percent",
        stat_value_base = required_integer(row, "Param3", label),
        stat_value_per_level = required_integer(row, "Param4", label),
        record_refresh_delay = required_integer(row, "perdelay", label),
    }
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported party-aura skill IDs are required")
    local skills = index(assert(records.load("data/global/excel/skills.txt")), "Id")
    local states = index(assert(records.load("data/global/excel/states.txt")), "state")
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        if row then
            local definition = decode(row, states)
            assert(not result[definition.skill_id], "duplicate party-aura skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
