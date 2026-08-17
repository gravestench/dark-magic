-- Decode explicitly supported ordinary melee actions from immutable Skills.txt
-- records. Admission is still exact-ID and target-locked in the coverage
-- manifest; satisfying this shape never admits another row implicitly.

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

local function required(row, column, expected, label)
    assert(row[column] == expected, label .. " has unsupported " .. column)
end

local function integer_or(row, column, fallback)
    local value = tonumber(row[column])
    if value == nil then
        return fallback
    end
    assert(value >= 0 and value == math.floor(value), "melee-action skill has invalid " .. column)
    return value
end

local function decode(row)
    local id = required_integer(row, "Id", "melee-action skill")
    local label = row.skill or ("skill " .. id)
    required(row, "srvstfunc", "1", label)
    required(row, "srvdofunc", "1", label)
    required(row, "cltstfunc", "1", label)
    required(row, "cltdofunc", "1", label)
    required(row, "range", "both", label)
    required(row, "itypea1", "weap", label)
    required(row, "anim", "A1", label)
    required(row, "UseAttackRate", "1", label)
    required(row, "TargetableOnly", "1", label)
    required(row, "SearchEnemyXY", "1", label)
    required(row, "AttackNoMana", "1", label)
    required(row, "leftskill", "1", label)
    required(row, "interrupt", "1", label)
    required(row, "general", "1", label)
    required(row, "InGame", "1", label)
    assert(required_integer(row, "mana", label) == 0, label .. " must have zero mana cost")
    assert(required_integer(row, "minmana", label) == 0, label .. " must have zero minimum mana")
    assert(required_integer(row, "SrcDam", label) == 128, label .. " has unsupported source damage")
    return {
        behavior = "action.melee",
        skill_id = id,
        mana_cost_raw = 0,
        effect_delay = 0,
        complete_delay = 1,
        animation_mode = row.anim,
        -- The target Attack row leaves selector 0 implicit with an empty cell.
        weapon_selection = integer_or(row, "weapsel", 0),
    }
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported melee-action skill IDs are required")
    local skills = index(assert(records.load("data/global/excel/skills.txt")), "Id")
    local definitions = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        -- Narrow module fixtures need not reproduce every declared retail row.
        -- Mounted-data coverage and the owned-archive boot are the strict target
        -- completeness gates; any row present here is still decoded by ID.
        if row then
            local definition = decode(row)
            assert(not definitions[definition.skill_id], "duplicate melee-action skill ID")
            definitions[definition.skill_id] = definition
        end
    end
    return definitions
end

return M
