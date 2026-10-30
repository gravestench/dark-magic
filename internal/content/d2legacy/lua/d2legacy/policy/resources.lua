-- Shared fixed-point resource transactions.

local M = {}

function M.mana_raw(vitals)
    local available = vitals:get("mana_raw")
    if available == 0 and vitals:get("mana") > 0 then
        return vitals:get("mana") * 256
    end
    return available
end

function M.spend_mana(vitals, cost_raw)
    assert(cost_raw >= 0, "mana cost must not be negative")
    local available = M.mana_raw(vitals)
    if available < cost_raw then
        return false
    end
    local remaining = available - cost_raw
    vitals:set("mana_raw", remaining)
    vitals:set("mana", math.floor(remaining / 256))
    return true
end

function M.restore_mana(vitals, amount_raw)
    assert(amount_raw >= 0, "mana restoration must not be negative")
    local current = M.mana_raw(vitals)
    local maximum = vitals:get("max_mana_raw")
    if maximum == 0 and vitals:get("max_mana") > 0 then
        maximum = vitals:get("max_mana") * 256
    end
    local restored = math.min(current + amount_raw, maximum)
    vitals:set("mana_raw", restored)
    vitals:set("mana", math.floor(restored / 256))
    return restored - current
end

-- Poison already stores the fractional deficit below the ceiling projection.
-- Healing uses that same remainder instead of rounding away each frame's gain.
function M.restore_health(vitals, amount_raw)
    assert(amount_raw >= 0 and amount_raw == math.floor(amount_raw), "health restoration must be nonnegative 8.8 units")
    local current = vitals:get("health") * 256 - vitals:get("poison_remainder_raw")
    local restored = math.min(current + amount_raw, vitals:get("max_health") * 256)
    local projected = math.floor((restored + 255) / 256)
    vitals:set("health", projected)
    vitals:set("poison_remainder_raw", projected * 256 - restored)
    return restored - current
end

return M
