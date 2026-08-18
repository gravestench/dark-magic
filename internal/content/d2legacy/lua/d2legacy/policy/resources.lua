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

return M
