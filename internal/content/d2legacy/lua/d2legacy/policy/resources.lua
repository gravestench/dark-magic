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

-- Health remains projected as whole points in the current player-vitals
-- contract. Accepting raw 8.8 amounts preserves the Diablo operation boundary
-- and makes the eventual fully fixed-point health migration mechanical.
function M.restore_health(vitals, amount_raw)
    assert(amount_raw >= 0 and amount_raw % 256 == 0, "health restoration must be whole 8.8 points")
    local current = vitals:get("health")
    local restored = math.min(current + amount_raw / 256, vitals:get("max_health"))
    vitals:set("health", restored)
    return (restored - current) * 256
end

return M
