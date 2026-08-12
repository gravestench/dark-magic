-- Resolve the currently verified fixed-point mitigation stages.

local M = {}

local function percent(value, resistance)
    return math.floor(value * (100 - resistance) / 100)
end

function M.apply(raw, channel, defense)
    assert(raw >= 0, "damage must be non-negative")
    if not defense then return raw end
    if channel == "physical" then
        local resistance = math.max(-100, math.min(100, defense:get("physical_resist")))
        return math.max(0, percent(raw, resistance) - defense:get("physical_reduction_raw"))
    end
    if channel == "fire" then
        local maximum = math.max(0, math.min(95, defense:get("max_fire_resist")))
        local resistance = math.max(-100, math.min(maximum, defense:get("fire_resist")))
        return math.max(0, percent(raw, resistance))
    end
    return raw
end

return M
