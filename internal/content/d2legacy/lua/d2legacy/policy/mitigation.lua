-- Resolve the currently verified fixed-point mitigation stages.

local M = {}
local elemental_resistance = {
    fire = { resistance = "fire_resist", maximum = "max_fire_resist" },
    cold = { resistance = "cold_resist", maximum = "max_cold_resist" },
    lightning = { resistance = "lightning_resist", maximum = "max_lightning_resist" },
}

local function percent(value, resistance)
    return math.floor(value * (100 - resistance) / 100)
end

function M.apply(raw, channel, defense)
    assert(raw >= 0, "damage must be non-negative")
    if not defense then
        return raw
    end
    if channel == "physical" then
        local resistance = math.max(-100, math.min(100, defense:get("physical_resist")))
        return math.max(0, percent(raw, resistance) - defense:get("physical_reduction_raw"))
    end
    local element = elemental_resistance[channel]
    if element then
        local maximum = math.max(0, math.min(95, defense:get(element.maximum)))
        local resistance = math.max(-100, math.min(maximum, defense:get(element.resistance)))
        return math.max(0, percent(raw, resistance))
    end
    return raw
end

return M
