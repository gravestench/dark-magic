-- Decode exact point-target relocation skills from immutable Skills/Levels
-- records. Admission remains exact-ID; a matching server function never opts
-- another skill into this family by resemblance.

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

local function integer(row, column, label, signed)
    local value = tonumber(row[column])
    assert(value and value == math.floor(value) and (signed or value >= 0), label .. " has invalid " .. column)
    return value
end

local function required(row, column, expected, label)
    assert(row[column] == expected, label .. " has unsupported " .. column)
end

local function decode(row, levels)
    local id = integer(row, "Id", "point-movement skill", false)
    local label = row.skill or ("skill " .. id)
    required(row, "srvstfunc", "", label)
    required(row, "srvdofunc", "27", label)
    required(row, "cltstfunc", "", label)
    required(row, "cltdofunc", "", label)
    required(row, "anim", "SC", label)
    required(row, "range", "none", label)
    required(row, "warp", "1", label)
    required(row, "interrupt", "1", label)
    required(row, "leftskill", "", label)
    required(row, "general", "", label)
    required(row, "InGame", "1", label)

    local shift = integer(row, "manashift", label, false)
    local level_policy = {}
    for _, level in ipairs(levels) do
        if level.Id and level.Id ~= "" then
            local level_id = integer(level, "Id", label .. " Levels", false)
            local policy = integer(level, "Teleport", label .. " level " .. level_id, false)
            assert(policy >= 0 and policy <= 2, label .. " level has unsupported Teleport policy")
            level_policy[level_id] = policy
        end
    end
    return {
        behavior = "movement.point-relocate",
        skill_id = id,
        mana_cost_raw = integer(row, "mana", label, false) * (2 ^ shift),
        mana_cost_per_level_raw = integer(row, "lvlmana", label, true) * (2 ^ shift),
        minimum_mana_cost_raw = integer(row, "minmana", label, false) * (2 ^ shift),
        effect_delay = 1,
        complete_delay = 2,
        requires_point_target = true,
        level_policy = level_policy,
    }
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported point-movement skill IDs are required")
    local skills = index(assert(records.load("data/global/excel/skills.txt")), "Id")
    local levels = assert(records.load("data/global/excel/Levels.txt"))
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        if row then
            local definition = decode(row, levels)
            assert(not result[definition.skill_id], "duplicate point-movement skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
