-- Reusable Skills.txt level progressions used by behavior definitions.

local M = {}

function M.linear(base, per_level, level)
    return base + math.max(level - 1, 0) * per_level
end

-- Damage columns use five authored per-level bands: levels 2-8, 9-16,
-- 17-22, 23-28, and 29+. Values are already shifted into 8.8 raw units by
-- the definition decoder before they reach this function.
function M.damage(base, gains, level)
    local remaining = math.max(level - 1, 0)
    local result = base
    for index, width in ipairs({ 7, 8, 6, 6 }) do
        local count = math.min(remaining, width)
        result = result + count * gains[index]
        remaining = remaining - count
    end
    return result + remaining * gains[5]
end

function M.mana_cost(definition, level)
    local cost = M.linear(definition.mana_cost_raw, definition.mana_cost_per_level_raw or 0, level)
    return math.max(cost, definition.minimum_mana_cost_raw or 0)
end

function M.damage_range(definition, level)
    local minimum =
        M.damage(definition.minimum_damage_raw, definition.minimum_damage_per_level_raw or { 0, 0, 0, 0, 0 }, level)
    local maximum =
        M.damage(definition.maximum_damage_raw, definition.maximum_damage_per_level_raw or { 0, 0, 0, 0, 0 }, level)
    return minimum, maximum
end

return M
