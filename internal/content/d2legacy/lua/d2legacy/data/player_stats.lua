-- Interpret class-dependent combat constants from CharStats.txt.

local records = require("engine.records/v1")
local M = {}

local function class_row(class)
    local wanted = string.lower(assert(class, "class is required"))
    for _, row in ipairs(records.load("data/global/excel/charstats.txt")) do
        if string.lower(row.class or "") == wanted then
            return row
        end
    end
    error("CharStats row is missing for " .. class)
end

function M.base_attack_rating(class, dexterity, item_attack_rating)
    local factor = math.floor(tonumber(class_row(class).ToHitFactor) or 0)
    return (item_attack_rating or 0) + 5 * ((dexterity or 0) - 7) + factor
end

function M.base_defense(dexterity, armor_class)
    return math.floor((dexterity or 0) / 4) + (armor_class or 0)
end

return M
