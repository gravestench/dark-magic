-- Decode explicitly admitted right-selected party stat auras.

local records = require("engine.records/v1")
local M = {}

local stat_recipes = {
    damagepercent = { stat = "damagepercent", operation = "percent", formula = "ln34" },
    coldresist = { stat = "cold_resist", operation = "add", formula = "dm34" },
    fireresist = { stat = "fire_resist", operation = "add", formula = "dm34" },
    item_tohit_percent = { stat = "item_tohit_percent", operation = "percent", formula = "ln34" },
    lightresist = { stat = "lightning_resist", operation = "add", formula = "dm34" },
    maxfireresist = { stat = "max_fire_resist", operation = "add", formula = "self_hard_level" },
    maxcoldresist = { stat = "max_cold_resist", operation = "add", formula = "self_hard_level" },
    maxlightresist = { stat = "max_lightning_resist", operation = "add", formula = "self_hard_level" },
    skill_armor_percent = { stat = "defense", operation = "percent", formula = "ln34" },
}

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

local function decode_learned_passive(row, states_by_name, label)
    local state_name = row.passivestate or ""
    if state_name == "" then
        assert(
            (row.passivestat1 or "") == "" and (row.passivecalc1 or "") == "",
            label .. " has a passive stat without a passive state"
        )
        return nil
    end

    local recipe = assert(stat_recipes[row.passivestat1], label .. " has an unsupported learned passive stat")
    local expression = row.passivecalc1 or ""
    local skill_name = string.match(expression, "^skill%('([^']+)'%.blvl%) %* par8$")
    local numerator, divisor
    if skill_name then
        numerator = required_integer(row, "Param8", label)
        divisor = 1
    else
        skill_name = string.match(expression, "^skill%('([^']+)'%.blvl%)/2$")
        assert(skill_name, label .. " has an unsupported learned passive formula")
        numerator = 1
        divisor = 2
    end
    assert(string.lower(skill_name) == string.lower(row.skill), label .. " passive does not use its own hard level")
    local state = assert(states_by_name[state_name], label .. " has an unknown learned passive state")
    return {
        state_id = state.state,
        stat = recipe.stat,
        operation = recipe.operation,
        value_numerator = numerator,
        value_divisor = divisor,
    }
end

local function decode_aura_stat(row, index_number, label)
    local authored_stat = row["aurastat" .. index_number] or ""
    if authored_stat == "" then
        return nil
    end
    local recipe = assert(stat_recipes[authored_stat], label .. " has an unsupported aura stat")
    local expression = row["aurastatcalc" .. index_number] or ""
    local result = {
        stat = recipe.stat,
        operation = recipe.operation,
        progression = recipe.formula,
    }
    if recipe.formula == "ln34" then
        assert(expression == "ln34", label .. " has an unsupported linear aura stat formula")
        result.value_base = required_integer(row, "Param3", label)
        result.value_per_level = required_integer(row, "Param4", label)
    elseif recipe.formula == "dm34" then
        assert(expression == "dm34", label .. " has an unsupported diminishing aura stat formula")
        result.value_minimum = required_integer(row, "Param3", label)
        result.value_maximum = required_integer(row, "Param4", label)
    else
        local skill_name = assert(
            string.match(expression, "^skill%('([^']+)'%.blvl%)$"),
            label .. " has an unsupported hard-level aura stat formula"
        )
        assert(
            string.lower(skill_name) == string.lower(row.skill),
            label .. " aura stat does not use its own hard level"
        )
    end
    return result
end

local function decode_aura_stats(row, label)
    local result = {}
    for index_number = 1, 6 do
        local stat = decode_aura_stat(row, index_number, label)
        if stat then
            result[#result + 1] = stat
        end
    end
    assert(#result > 0, label .. " has no supported aura stats")
    return result
end

local function state_stat_is_authored(row, state_stat)
    if not state_stat or state_stat == "" then
        return false
    end
    for index_number = 1, 6 do
        if row["aurastat" .. index_number] == state_stat then
            return true
        end
    end
    return false
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
    local aura_stats = decode_aura_stats(row, label)
    assert(
        required_integer(row, "mana", label) == 0 and required_integer(row, "lvlmana", label) == 0,
        label .. " unexpectedly consumes mana"
    )
    local state = assert(states_by_name[row.aurastate], label .. " has an unknown owner state")
    local target_state = assert(states_by_name[row.auratargetstate], label .. " has an unknown target state")
    assert(
        state.aura == "1" and state_stat_is_authored(row, state.stat),
        label .. " owner state does not match any authored aura stat"
    )
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
        stats = aura_stats,
        stat = aura_stats[1].stat,
        stat_operation = aura_stats[1].operation,
        stat_value_base = aura_stats[1].value_base,
        stat_value_per_level = aura_stats[1].value_per_level,
        record_refresh_delay = required_integer(row, "perdelay", label),
        learned_passive = decode_learned_passive(row, states_by_name, label),
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
