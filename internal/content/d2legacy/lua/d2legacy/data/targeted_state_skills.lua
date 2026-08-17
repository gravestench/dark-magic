-- Decode explicitly admitted friendly-target timed multi-stat skills.

local records = require("engine.records/v1")
local skill_modifiers = require("d2legacy.data.skill_modifiers")
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

local function shifted(row, column, label)
    return required_integer(row, column, label) * (2 ^ required_integer(row, "HitShift", label))
end

local function shifted_gains(row, prefix, label)
    local result = {}
    for tier = 1, 5 do
        result[tier] = shifted(row, prefix .. tier, label)
    end
    return result
end

local function decode(row, skills_by_name, states_by_name)
    local id = required_integer(row, "Id", "targeted state skill")
    local label = row.skill or ("skill " .. id)
    assert(row.srvstfunc == "" and row.srvdofunc == "25", label .. " has unsupported server functions")
    assert(
        row.TargetPet == "1" and row.TargetAlly == "1" and row.range == "none",
        label .. " has unsupported target policy"
    )
    assert(row.leftskill == "" and row.general == "" and row.InGame == "1", label .. " has unsupported flags")
    assert(row.auralencalc == "ln12", label .. " has unsupported duration formula")
    assert(
        row.aurastat1 == "firemindam"
            and row.aurastatcalc1 == "enma"
            and row.aurastat2 == "firemaxdam"
            and row.aurastatcalc2 == "exma"
            and row.aurastat3 == "item_tohit_percent"
            and row.aurastatcalc3 == "toht",
        label .. " has unsupported stat bundle"
    )
    assert(row.EType == "fire" and row.ToHitCalc == "", label .. " has unsupported damage or attack formula")
    local state = assert(states_by_name[row.aurastate], label .. " has an unknown state")
    local synergy_ids, synergy_percent =
        skill_modifiers.hard_level_sum_percent(row, "EDmgSymPerCalc", "Param8", skills_by_name, label)
    return {
        behavior = "state.targeted-timed-stats",
        skill_id = id,
        mana_cost_raw = required_integer(row, "mana", label) * (2 ^ required_integer(row, "manashift", label)),
        mana_cost_per_level_raw = required_integer(row, "lvlmana", label)
            * (2 ^ required_integer(row, "manashift", label)),
        minimum_mana_cost_raw = required_integer(row, "minmana", label)
            * (2 ^ required_integer(row, "manashift", label)),
        effect_delay = 1,
        complete_delay = 2,
        state_id = state.state,
        target_allies = true,
        target_owned_units = true,
        fallback_to_self = true,
        duration_base = required_integer(row, "Param1", label),
        duration_per_level = required_integer(row, "Param2", label),
        attack_rating_percent_base = required_integer(row, "ToHit", label),
        attack_rating_percent_per_level = required_integer(row, "LevToHit", label),
        minimum_damage_raw = shifted(row, "EMin", label),
        maximum_damage_raw = shifted(row, "EMax", label),
        minimum_damage_per_level_raw = shifted_gains(row, "EMinLev", label),
        maximum_damage_per_level_raw = shifted_gains(row, "EMaxLev", label),
        damage_synergy_skill_ids = synergy_ids,
        damage_synergy_percent_per_level = synergy_percent,
    }
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported targeted-state skill IDs are required")
    local rows = assert(records.load("data/global/excel/skills.txt"))
    local skills_by_name = skill_modifiers.by_name(rows)
    local skills_by_id = index(rows, "Id")
    local states_by_name = index(assert(records.load("data/global/excel/states.txt")), "state")
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills_by_id[tostring(skill_id)]
        if row then
            local definition = decode(row, skills_by_name, states_by_name)
            assert(not result[definition.skill_id], "duplicate targeted-state skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
