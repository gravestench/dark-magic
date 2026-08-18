-- Interpret class-dependent combat and resource constants from CharStats.txt.

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

function M.mana_regen_frames(class)
    local authored = class_row(class).ManaRegen
    if authored == nil or authored == "" then
        -- Narrow record fixtures which do not exercise resources can omit the
        -- column. Mounted target rows always author it.
        return 0
    end
    local frames = math.floor(assert(tonumber(authored), "CharStats ManaRegen is invalid"))
    -- The recovered runtime substitutes 7500 ticks when 25 * ManaRegen is
    -- zero. Keeping the equivalent 300-frame record value here lets the
    -- per-tick consumer retain one arithmetic path.
    if frames == 0 then
        return 300
    end
    assert(frames > 0, "CharStats ManaRegen must be nonnegative")
    return frames
end

return M
