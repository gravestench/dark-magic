-- Decode explicitly supported targetless, timed self-state skills.

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

local function shifted(row, value_column, shift_column, label)
    return required_integer(row, value_column, label) * (2 ^ required_integer(row, shift_column, label))
end

local function synergy_ids(expression, skills_by_name, marker, label)
    local result = {}
    for name in string.gmatch(expression or "", "skill%('([^']+)'%.blvl%)") do
        local skill = assert(skills_by_name[string.lower(name)], label .. " has an unknown synergy " .. name)
        result[#result + 1] = required_integer(skill, "Id", label)
    end
    assert(#result > 0 and string.find(expression, marker, 1, true), label .. " has an unsupported formula")
    return result
end

local function same_ids(left, right)
    if #left ~= #right then
        return false
    end
    for index, value in ipairs(left) do
        if value ~= right[index] then
            return false
        end
    end
    return true
end

local function decode(row, skills_by_name, states_by_name)
    local id = required_integer(row, "Id", "self-state skill")
    local label = row.skill or ("skill " .. id)
    assert(row.srvstfunc == "" and row.srvdofunc == "18", label .. " has unsupported server functions")
    assert(
        row.aurastate and row.aurastate ~= "" and row.aurastat1 == "skill_armor_percent",
        label .. " has an unsupported state or stat"
    )
    assert(row.aurastatcalc1 == "ln12", label .. " has an unsupported stat formula")
    assert(string.find(row.auralencalc or "", "*par7", 1, true), label .. " has an unsupported synergy formula")
    assert(
        row.auraevent1 == "damagedinmelee" and row.auraeventfunc1 == "2",
        label .. " has an unsupported reactive event"
    )
    assert(string.find(row.calc1 or "", "*par8", 1, true), label .. " has an unsupported reaction formula")
    local duration_synergies = synergy_ids(row.auralencalc, skills_by_name, "ln34", label)
    local reaction_synergies = synergy_ids(row.calc1, skills_by_name, "ln56", label)
    assert(same_ids(duration_synergies, reaction_synergies), label .. " has mismatched reactive synergies")
    local armor_state = assert(states_by_name[row.aurastate], label .. " has an unknown armor state")
    local state_group = required_integer(armor_state, "group", label)
    assert(state_group > 0, label .. " armor state has no exclusive group")
    local reaction_state = assert(states_by_name.freeze, label .. " requires the freeze state")
    return {
        behavior = "state.self-timed-stat",
        skill_id = id,
        mana_cost_raw = shifted(row, "mana", "manashift", label),
        effect_delay = 1,
        complete_delay = 2,
        state_id = row.aurastate,
        stat = "defense",
        stat_operation = "percent",
        stat_base = required_integer(row, "Param1", label),
        stat_per_level = required_integer(row, "Param2", label),
        duration_base = required_integer(row, "Param3", label),
        duration_per_level = required_integer(row, "Param4", label),
        duration_synergy_per_level = required_integer(row, "Param7", label),
        duration_synergy_skill_ids = duration_synergies,
        exclusive_group = "state-group:" .. state_group,
        on_melee_hit_state_id = reaction_state.state,
        on_melee_hit_duration_base = required_integer(row, "Param5", label),
        on_melee_hit_duration_per_level = required_integer(row, "Param6", label),
        on_melee_hit_synergy_percent = required_integer(row, "Param8", label),
        on_melee_hit_synergy_skill_ids = reaction_synergies,
        on_melee_hit_disables_action = true,
    }
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported self-state skill IDs are required")
    local rows = assert(records.load("data/global/excel/skills.txt"))
    local states_by_name = index(assert(records.load("data/global/excel/states.txt")), "state")
    local skills = index(rows, "Id")
    local skills_by_name = {}
    for _, row in ipairs(rows) do
        if row.skill and row.skill ~= "" then
            skills_by_name[string.lower(row.skill)] = row
        end
    end
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        -- Narrow module fixtures do not have to reproduce every declared
        -- retail row. The mounted-data coverage report is the strict target
        -- completeness gate; any row present here is still decoded by ID.
        if row then
            local definition = decode(row, skills_by_name, states_by_name)
            assert(not result[definition.skill_id], "duplicate self-state skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
