-- Decode explicitly admitted right-selected auras whose periodic work targets
-- authored world corpses. The definition describes target policy, chance, and
-- ordered operations; runtime code never branches on a skill name or ID.

local records = require("engine.records/v1")
local M = {}

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

local function assert_empty_aura_stats(row, label)
    for index_number = 1, 6 do
        assert((row["aurastat" .. index_number] or "") == "", label .. " has unsupported maintained aura stat")
        assert((row["aurastatcalc" .. index_number] or "") == "", label .. " has unsupported aura stat formula")
    end
end

local function decode(row, states_by_name)
    local id = required_nonnegative_integer(row, "Id", "corpse-periodic aura skill")
    local label = row.skill or ("skill " .. id)
    assert(row.srvstfunc == "" and row.srvdofunc == "82", label .. " has unsupported server functions")
    assert(
        row.aura == "1"
            and (row.immediate or "") == ""
            and row.leftskill == ""
            and row.range == "none"
            and row.InGame == "1"
            and row.InTown == "1",
        label .. " has unsupported activation flags"
    )
    assert(row.aurafilter == "4354", label .. " has unsupported corpse filter")
    assert(row.aurarangecalc == "ln12", label .. " has unsupported radius formula")
    assert((row.auratargetstate or "") == "", label .. " has unsupported target state")
    assert(row.calc1 == "dm34", label .. " has unsupported chance formula")
    assert(row.calc2 == "ln56" and row.calc3 == "ln56", label .. " has unsupported recovery formula")
    assert_empty_aura_stats(row, label)
    local state = assert(states_by_name[row.aurastate], label .. " has an unknown owner state")
    assert(state.aura == "1", label .. " state is not an aura")
    local mana_shift = required_nonnegative_integer(row, "manashift", label)
    assert(mana_shift <= 8, label .. " has unsupported mana precision")
    return {
        behavior = "aura.selected-corpse-periodic",
        activation = "selected_right",
        skill_id = id,
        mana_cost_raw = required_nonnegative_integer(row, "mana", label) * (2 ^ mana_shift),
        mana_cost_per_level_raw = required_nonnegative_integer(row, "lvlmana", label) * (2 ^ mana_shift),
        minimum_mana_cost_raw = required_nonnegative_integer(row, "minmana", label) * 256,
        effect_delay = 0,
        complete_delay = 0,
        state_id = state.state,
        -- The aura owner still needs an explicit state relationship for sound,
        -- overlay, and connected projection even though no party target state
        -- is authored.
        target_state_id = state.state,
        radius_base = required_nonnegative_integer(row, "Param1", label),
        radius_per_level = required_nonnegative_integer(row, "Param2", label),
        target_policy = "owner",
        stats = {},
        pulse = {
            target_policy = "eligible_corpses",
            chance_progression = "dm34",
            chance_minimum = required_nonnegative_integer(row, "Param3", label),
            chance_maximum = required_nonnegative_integer(row, "Param4", label),
            effects = {
                {
                    order = 1,
                    kind = "restore_owner_life",
                    progression = "linear_raw",
                    value_base = required_nonnegative_integer(row, "Param5", label) * 256,
                    value_per_level = required_nonnegative_integer(row, "Param6", label) * 256,
                },
                {
                    order = 2,
                    kind = "restore_owner_mana",
                    progression = "linear_raw",
                    value_base = required_nonnegative_integer(row, "Param5", label) * 256,
                    value_per_level = required_nonnegative_integer(row, "Param6", label) * 256,
                },
                { order = 3, kind = "consume_corpse", progression = "constant", value_base = 0 },
            },
            period_ticks = required_nonnegative_integer(row, "perdelay", label),
        },
        record_refresh_delay = required_nonnegative_integer(row, "perdelay", label),
        learned_passive = nil,
    }
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported corpse-periodic aura skill IDs are required")
    local skills = index(assert(records.load("data/global/excel/skills.txt")), "Id")
    local states = index(assert(records.load("data/global/excel/states.txt")), "state")
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        if row then
            local definition = decode(row, states)
            assert(not result[definition.skill_id], "duplicate corpse-periodic aura skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
