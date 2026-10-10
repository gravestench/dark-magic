-- Interpret the session difficulty's row from DifficultyLevels.txt.
--
-- The immutable game rules choose the difficulty; this table supplies what
-- that choice means. Consumers read the selected row here instead of keeping
-- their own per-difficulty constants.

local records = require("engine.records/v1")
local game_rules = require("d2legacy.policy.game_rules")

local M = {}

local NAMES = { [0] = "Normal", [1] = "Nightmare", [2] = "Hell" }

-- Record capabilities return defensive copies, so the selected row is decoded
-- once per authoritative VM rather than for every consumer query.
local selected

local function integer(row, key)
    local value = tonumber(row[key])
    assert(value and value == math.floor(value), "DifficultyLevels " .. row.Name .. " " .. key .. " must be an integer")
    return value
end

local function row_for(difficulty)
    local name = assert(NAMES[difficulty], "unsupported difficulty")
    local rows, err = records.load("data/global/excel/difficultylevels.txt")
    assert(rows, "load DifficultyLevels: " .. tostring(err))
    for _, row in ipairs(rows) do
        if row.Name == name then
            return row
        end
    end
    error("DifficultyLevels row is missing for " .. name)
end

local function current()
    if selected then
        return selected
    end
    local row = row_for(game_rules.difficulty())
    selected = {
        resist_penalty = integer(row, "ResistPenalty"),
        monster_freeze_divisor = integer(row, "MonsterFreezeDivisor"),
        monster_cold_divisor = integer(row, "MonsterColdDivisor"),
        death_experience_penalty = integer(row, "DeathExpPenalty"),
    }
    assert(selected.monster_freeze_divisor > 0, "DifficultyLevels MonsterFreezeDivisor must be positive")
    assert(selected.monster_cold_divisor > 0, "DifficultyLevels MonsterColdDivisor must be positive")
    return selected
end

-- The expansion resistance penalty is added to a player's elemental
-- resistances before the maximum-resistance cap. It belongs to the game, not
-- to the character's durable base resistance.
function M.resist_penalty()
    return current().resist_penalty
end

-- Monster freeze and chill lengths are divided by separate authored columns.
function M.monster_freeze_divisor()
    return current().monster_freeze_divisor
end

function M.monster_cold_divisor()
    return current().monster_cold_divisor
end

-- Percentage of the current level's experience span lost on death.
function M.death_experience_penalty()
    return current().death_experience_penalty
end

return M
