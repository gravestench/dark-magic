-- Decode the first targetless, non-damage skill behavior from Skills.txt.

local records = require("engine.records/v1")
local M = {}

function M.load()
    local result = {}
    for _, row in ipairs(records.load("data/global/excel/skills.txt")) do
        if string.lower(row.skill or "") == "frozen armor" then
            local id = assert(tonumber(row.Id), "Frozen Armor ID is required")
            result[math.floor(id)] = {
                state_id = "frozen_armor",
                -- Param1 is the base duration for this focused behavior. The
                -- fallback keeps older tables executable when the field is blank.
                duration = math.max(1, math.floor(tonumber(row.Param1) or 600)),
            }
        end
    end
    return result
end

return M
