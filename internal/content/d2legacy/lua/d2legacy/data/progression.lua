-- Interpret the cumulative class thresholds in Experience.txt.

local records = require("engine.records/v1")
local M = {}

function M.load()
    local rows = records.load("data/global/excel/experience.txt")
    local thresholds = {}
    for _, class in ipairs({"Amazon", "Sorceress", "Necromancer", "Paladin", "Barbarian", "Druid", "Assassin"}) do
        thresholds[class] = {}
        for _, row in ipairs(rows) do
            local value = tonumber(row[class])
            if value and value >= 0 then
                thresholds[class][#thresholds[class] + 1] = math.floor(value)
            end
        end
    end
    return thresholds
end

function M.level_for(thresholds, class, experience, current)
    local class_thresholds = thresholds[class]
    if not class_thresholds or #class_thresholds == 0 then return current end
    local level = current
    for candidate = current + 1, #class_thresholds do
        if experience < class_thresholds[candidate] then break end
        level = candidate
    end
    return level
end

return M
