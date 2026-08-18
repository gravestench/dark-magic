-- Decode explicitly admitted right-selected party auras which own a periodic
-- direct-effect schedule. Authored columns may also contribute maintained stat
-- sources; Meditation deliberately exercises both paths on one definition.

local records = require("engine.records/v1")
local M = {}

local maintained_stat_recipes = {
    manarecoverybonus = { stat = "manarecoverybonus", operation = "add", progression = "ln34" },
}

local function required_nonnegative_integer(row, column, label)
    local value = tonumber(row[column])
    assert(value and value >= 0 and value == math.floor(value), label .. " has invalid " .. column)
    return value
end

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

local function banded_effect(row, label, source_skill_id, level_source)
    assert(required_nonnegative_integer(row, "HitShift", label) == 8, label .. " has unsupported healing precision")
    return {
        kind = "heal_life",
        progression = "banded",
        source_skill_id = source_skill_id,
        level_source = level_source,
        value_base = required_nonnegative_integer(row, "EMin", label),
        value_per_level = {
            required_nonnegative_integer(row, "EMinLev1", label),
            required_nonnegative_integer(row, "EMinLev2", label),
            required_nonnegative_integer(row, "EMinLev3", label),
            required_nonnegative_integer(row, "EMinLev4", label),
            required_nonnegative_integer(row, "EMinLev5", label),
        },
    }
end

local function decode_effect(row, stat, formula, skills_by_name, order, label)
    local maintained = maintained_stat_recipes[stat]
    if maintained then
        assert(formula == maintained.progression, label .. " has unsupported maintained-stat formula")
        return {
            mode = "maintained",
            stat = maintained.stat,
            operation = maintained.operation,
            progression = maintained.progression,
            value_base = required_nonnegative_integer(row, "Param3", label),
            value_per_level = required_nonnegative_integer(row, "Param4", label),
            order = order,
        }
    end
    if stat == "item_poisonlengthresist" then
        assert(formula == "100-dm34", label .. " has unsupported duration-reduction formula")
        return {
            kind = "scale_remaining_timed_state",
            progression = "one_minus_diminishing",
            value_minimum = required_nonnegative_integer(row, "Param3", label),
            value_maximum = required_nonnegative_integer(row, "Param4", label),
            order = order,
        }
    end
    assert(stat == "hitpoints", label .. " has unsupported periodic direct-effect stat " .. stat)
    if formula == "edns" then
        local result = banded_effect(row, label, required_nonnegative_integer(row, "Id", label), "aura")
        result.order = order
        return result
    end
    local referenced_name = formula:match("^skill%('([^']+)'%.edns%)$")
    assert(referenced_name, label .. " has unsupported direct-effect formula " .. formula)
    local referenced = assert(skills_by_name[referenced_name], label .. " references unknown skill " .. referenced_name)
    local referenced_label = referenced.skill or ("skill " .. tostring(referenced.Id))
    local result = banded_effect(
        referenced,
        referenced_label,
        required_nonnegative_integer(referenced, "Id", referenced_label),
        "learned_skill"
    )
    result.order = order
    return result
end

local function decode(row, states_by_name, skills_by_name)
    local id = required_nonnegative_integer(row, "Id", "periodic party aura skill")
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
    assert(row.aurafilter == "73731" or row.aurafilter == "73729", label .. " has unsupported target filter")
    assert(row.aurarangecalc == "ln12", label .. " has unsupported radius formula")
    local stats, effects = {}, {}
    for index_number = 1, 6 do
        local stat = row["aurastat" .. index_number] or ""
        local formula = row["aurastatcalc" .. index_number] or ""
        if stat ~= "" then
            local decoded = decode_effect(row, stat, formula, skills_by_name, index_number, label)
            if decoded.mode == "maintained" then
                decoded.mode = nil
                stats[#stats + 1] = decoded
            else
                effects[#effects + 1] = decoded
            end
        else
            assert(formula == "", label .. " has a formula without an aura stat")
        end
    end
    assert(#effects > 0, label .. " has no periodic direct effects")
    local state = assert(states_by_name[row.aurastate], label .. " has an unknown owner state")
    local target_state = assert(states_by_name[row.auratargetstate], label .. " has an unknown target state")
    assert(state.aura == "1" and target_state.aura == "1", label .. " state is not an aura")
    local mana_shift = required_nonnegative_integer(row, "manashift", label)
    assert(mana_shift <= 8, label .. " has unsupported mana precision")
    return {
        behavior = "aura.selected-party-periodic",
        activation = "selected_right",
        skill_id = id,
        mana_cost_raw = required_nonnegative_integer(row, "mana", label) * (2 ^ mana_shift),
        mana_cost_per_level_raw = required_nonnegative_integer(row, "lvlmana", label) * (2 ^ mana_shift),
        minimum_mana_cost_raw = required_nonnegative_integer(row, "minmana", label) * 256,
        effect_delay = 0,
        complete_delay = 0,
        state_id = state.state,
        target_state_id = target_state.state,
        radius_base = required_nonnegative_integer(row, "Param1", label),
        radius_per_level = required_nonnegative_integer(row, "Param2", label),
        target_policy = row.aurafilter == "73729" and "aligned_players" or "aligned_players_and_monsters",
        stats = stats,
        pulse = {
            effects = effects,
            period_ticks = required_nonnegative_integer(row, "perdelay", label),
        },
        record_refresh_delay = required_nonnegative_integer(row, "perdelay", label),
        learned_passive = nil,
    }
end

local function state_policy(states)
    local policy = { poison_state_id = "poison", duration_reduced_states = {} }
    for state_id, row in pairs(states) do
        local authored_shrine = row.curse == "1" and state_id:match("^shrine_") ~= nil
        if row.curse == "1" and (row.curable == "1" or authored_shrine) then
            policy.duration_reduced_states[state_id] = true
        end
    end
    return policy
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported periodic party-aura skill IDs are required")
    local skill_rows = assert(records.load("data/global/excel/skills.txt"))
    local skills = index(skill_rows, "Id")
    local skills_by_name = index(skill_rows, "skill")
    local states = index(assert(records.load("data/global/excel/states.txt")), "state")
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        if row then
            local definition = decode(row, states, skills_by_name)
            assert(not result[definition.skill_id], "duplicate periodic party-aura skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result, state_policy(states)
end

return M
