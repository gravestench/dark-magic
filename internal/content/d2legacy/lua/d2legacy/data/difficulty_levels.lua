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
    }
    return selected
end

-- The expansion resistance penalty is added to a player's elemental
-- resistances before the maximum-resistance cap. It belongs to the game, not
-- to the character's durable base resistance.
function M.resist_penalty()
    return current().resist_penalty
end

return M
