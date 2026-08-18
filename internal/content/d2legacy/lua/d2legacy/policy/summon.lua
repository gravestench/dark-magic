-- Integer progressions shared by record-driven summon families.

local M = {}

function M.pet_limit(definition, level)
    assert(type(definition) == "table" and level >= 1, "invalid summon pet-limit inputs")
    local base_max = definition.category_base_max or 0
    assert(base_max >= 0, "invalid summon category base maximum")
    local authored
    if definition.limit_policy == "tiered" then
        authored = level < 4 and level or 2 + math.floor(level / 3)
    elseif definition.limit_policy == "skill_level" then
        authored = level
    else
        error("unsupported summon pet-limit policy " .. tostring(definition.limit_policy))
    end
    return math.max(authored, base_max)
end

function M.pet_level(skill_level, owner_level)
    assert(skill_level >= 1 and owner_level >= 1, "invalid summon pet-level inputs")
    return math.min(math.max(skill_level + math.floor(3 * owner_level / 4), 1), owner_level)
end

function M.whole_life_raw(base_raw, percent, flat_raw)
    assert(base_raw >= 0 and base_raw % 256 == 0 and percent >= 0 and flat_raw >= 0, "invalid summon life inputs")
    return math.floor(base_raw / 256 * (100 + percent) / 100) * 256 + flat_raw
end

function M.damage_raw(base_raw, flat_raw, percent)
    assert(base_raw >= 0 and flat_raw >= 0 and percent >= 0, "invalid summon damage inputs")
    return math.floor((base_raw + flat_raw) * (100 + percent) / 100)
end

function M.granted_skill_level(definition, level, mastery_level)
    assert(type(definition) == "table" and level >= 1 and mastery_level >= 0, "invalid granted-skill inputs")
    if definition.granted_skill_level_policy == "none" then
        return 0
    end
    if definition.granted_skill_level_policy == "mastery_plus_half_after_two" then
        return mastery_level + (level < 4 and 0 or math.floor((level - 2) / 2))
    end
    error("unsupported granted-skill level policy " .. tostring(definition.granted_skill_level_policy))
end

function M.duration_ticks(definition, level)
    assert(type(definition) == "table" and level >= 1, "invalid summon duration inputs")
    return (definition.duration_base_ticks or 0) + (level - 1) * (definition.duration_per_level_ticks or 0)
end

return M
