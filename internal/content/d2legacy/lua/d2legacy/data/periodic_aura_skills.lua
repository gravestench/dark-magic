-- Decode explicitly admitted right-selected party auras whose authored stat is
-- a periodic direct effect rather than a maintained modifier.

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
    local id = required_integer(row, "Id", "periodic party aura skill")
    local label = row.skill or ("skill " .. id)
    assert(row.srvstfunc == "" and row.srvdofunc == "65", label .. " has unsupported server functions")
    assert(
        row.aura == "1"
            and (row.immediate or "") == ""
            and row.leftskill == ""
            and row.range == "none"
            and row.InGame == "1"
            and row.InTown == "1",
        label .. " has unsupported activation flags"
    )
    assert(row.aurafilter == "73731", label .. " has unsupported target filter")
    assert(row.aurarangecalc == "ln12", label .. " has unsupported radius formula")
    assert(
        row.aurastat1 == "hitpoints" and row.aurastatcalc1 == "edns",
        label .. " has an unsupported periodic direct effect"
    )
    for index_number = 2, 6 do
        assert((row["aurastat" .. index_number] or "") == "", label .. " has unsupported extra aura stats")
    end
    assert(required_integer(row, "HitShift", label) == 8, label .. " has unsupported healing precision")
    local state = assert(states_by_name[row.aurastate], label .. " has an unknown owner state")
    local target_state = assert(states_by_name[row.auratargetstate], label .. " has an unknown target state")
    assert(state.aura == "1" and target_state.aura == "1", label .. " state is not an aura")
    local mana_shift = required_integer(row, "manashift", label)
    assert(mana_shift <= 8, label .. " has unsupported mana precision")
    return {
        behavior = "aura.selected-party-periodic",
        activation = "selected_right",
        skill_id = id,
        mana_cost_raw = required_integer(row, "mana", label) * (2 ^ mana_shift),
        mana_cost_per_level_raw = required_integer(row, "lvlmana", label) * (2 ^ mana_shift),
        minimum_mana_cost_raw = required_integer(row, "minmana", label) * 256,
        effect_delay = 0,
        complete_delay = 0,
        state_id = state.state,
        target_state_id = target_state.state,
        radius_base = required_integer(row, "Param1", label),
        radius_per_level = required_integer(row, "Param2", label),
        stats = {},
        pulse = {
            kind = "heal_life",
            value_base = required_integer(row, "EMin", label),
            value_per_level = {
                required_integer(row, "EMinLev1", label),
                required_integer(row, "EMinLev2", label),
                required_integer(row, "EMinLev3", label),
                required_integer(row, "EMinLev4", label),
                required_integer(row, "EMinLev5", label),
            },
            period_ticks = required_integer(row, "perdelay", label),
        },
        record_refresh_delay = required_integer(row, "perdelay", label),
        learned_passive = nil,
    }
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported periodic party-aura skill IDs are required")
    local skills = index(assert(records.load("data/global/excel/skills.txt")), "Id")
    local states = index(assert(records.load("data/global/excel/states.txt")), "state")
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        if row then
            local definition = decode(row, states)
            assert(not result[definition.skill_id], "duplicate periodic party-aura skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
