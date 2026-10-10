-- Interpret the cumulative class thresholds in Experience.txt.

local records = require("engine.records/v1")
local M = {}

local CLASSES = {"Amazon", "Sorceress", "Necromancer", "Paladin", "Barbarian", "Druid", "Assassin"}

-- Kill-experience scaling from the same table. The MaxLvl row authors each
-- class's level cap and, in ExpRatio, the binary shift applied to every
-- level's ExpRatio. Narrow record fixtures can omit the row; they then have
-- no cap and no scaling.
function M.experience_rules()
    local rules = {ratio_shift = 0, ratios = {}, max_levels = {}}
    for _, row in ipairs(records.load("data/global/excel/experience.txt")) do
        local ratio = math.floor(tonumber(row.ExpRatio) or 0)
        if row.Level == "MaxLvl" then
            rules.ratio_shift = ratio
            for _, class in ipairs(CLASSES) do
                rules.max_levels[class] = math.floor(tonumber(row[class]) or 0)
            end
        elseif tonumber(row.Level) then
            rules.ratios[math.floor(tonumber(row.Level))] = ratio
        end
    end
    return rules
end

function M.load()
    local rows = records.load("data/global/excel/experience.txt")
    local thresholds = {}
    for _, class in ipairs(CLASSES) do
        thresholds[class] = {}
        for _, row in ipairs(rows) do
            -- The authored MaxLvl row stores the level cap, not a threshold.
            local value = row.Level ~= "MaxLvl" and tonumber(row[class]) or nil
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
