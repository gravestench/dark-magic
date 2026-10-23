-- Experience a player gains from killing a monster.
--
-- Level scaling was recovered from pinned 1.10f D2MOO
-- (UNIT/SUnitDmg.cpp SUNITDMG_ComputeExperienceGain and
-- SUNITDMG_AddExperienceForPlayer). The nonpositive-share branch follows
-- PXL-R05a's 1.13c static evidence. Both await 1.14d runtime confirmation.
--
-- The monster's experience is reduced when the levels are far apart, scaled by
-- the killer level's Experience.txt ExpRatio, raised by item_addexperience,
-- and positive shares gain nothing at the class level cap. The recovered
-- nonpositive-share return precedes that cap check. Hireling experience is a
-- separate consumer.

local M = {}

local INT_MAX = 2147483647

-- The VM is Lua 5.1 with float numbers: shift by dividing by a power of two.
local function shr(value, bits)
    return math.floor(value / 2 ^ bits)
end

-- Killer above the monster, indexed by the level difference (capped at 10).
local ABOVE = { 256, 256, 256, 256, 256, 256, 207, 159, 110, 61, 13 }
-- Monster above a killer below level 25, indexed the same way.
local BELOW = { 256, 256, 256, 256, 256, 256, 225, 174, 92, 38, 5 }

-- The recovered helper computes value * numerator / denominator while
-- avoiding 32-bit overflow; its precision-losing branches are kept.
function M.percentage(value, numerator, denominator)
    if denominator == 0 then
        return 0
    end
    if value <= 1048576 then
        if numerator <= 65536 then
            return math.floor(numerator * value / denominator)
        end
        if denominator <= shr(numerator, 4) then
            return value * math.floor(numerator / denominator)
        end
    elseif denominator <= shr(value, 4) then
        return numerator * math.floor(value / denominator)
    end
    return math.floor(numerator * value / denominator)
end

local function level_factor(table, difference)
    return table[math.min(difference, #table - 1) + 1]
end

-- rules comes from d2legacy.data.progression.experience_rules().
function M.gain(rules, class, killer_level, monster_level, experience, add_experience_percent)
    if experience <= 0 then
        return 1
    end
    experience = math.min(experience, shr(INT_MAX, 8))
    local max_level = rules.max_levels[class]
    if max_level and max_level > 0 and killer_level >= max_level then
        return 0
    end

    local result = experience
    if monster_level <= killer_level then
        local factor = level_factor(ABOVE, killer_level - monster_level)
        if factor ~= 256 then
            result = M.percentage(experience, factor, 256)
        end
    elseif killer_level < 25 or monster_level <= 0 then
        local factor = level_factor(BELOW, monster_level - killer_level)
        if factor ~= 256 then
            result = M.percentage(experience, factor, 256)
        end
    else
        result = M.percentage(experience, killer_level, monster_level)
    end

    local shift = rules.ratio_shift
    if result > 0 and shift > 0 and shift < 32 then
        local ratio = rules.ratios[killer_level] or 0
        if result <= shr(INT_MAX, shift + shr(ratio, shift)) then
            result = shr(result * ratio, shift)
        else
            result = ratio * shr(result, shift)
        end
    end

    if add_experience_percent ~= 0 then
        result = result + M.percentage(result, add_experience_percent, 100)
    end
    return result
end

-- Total experience never exceeds the threshold of the class level cap.
function M.capped_total(rules, thresholds, class, experience, gained)
    local max_level = rules.max_levels[class]
    local cap = max_level and thresholds[class] and thresholds[class][max_level]
    if not cap then
        return experience + gained
    end
    return math.min(experience + gained, cap)
end

return M
