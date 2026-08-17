-- Typed damage-stage vocabulary. Channels stay independent until the health
-- commit boundary so later conversion, resistance, duration, and proc stages
-- do not have to reconstruct facts from one scalar.

local M = {}

M.channels = { "physical", "fire", "lightning", "cold", "magic", "poison" }
M.immediate_channels = { "physical", "fire", "lightning", "cold", "magic" }

local known = {}
for _, channel in ipairs(M.channels) do
    known[channel] = true
end

function M.single(channel, raw)
    assert(known[channel], "unsupported damage channel: " .. tostring(channel))
    assert(raw >= 0, "damage must be non-negative")
    return { [channel] = raw }
end

function M.normalize(values)
    assert(type(values) == "table", "damage bundle must be a table")
    local normalized = {}
    for channel, raw in pairs(values) do
        assert(known[channel], "unsupported damage channel: " .. tostring(channel))
        assert(type(raw) == "number" and raw >= 0, "damage must be non-negative")
    end
    for _, channel in ipairs(M.channels) do
        normalized[channel] = values[channel] or 0
    end
    return normalized
end

function M.immediate_total(values)
    local total = 0
    for _, channel in ipairs(M.immediate_channels) do
        total = total + values[channel]
    end
    return total
end

function M.label(values)
    local label = "none"
    for _, channel in ipairs(M.channels) do
        if values[channel] > 0 then
            if label ~= "none" then
                return "mixed"
            end
            label = channel
        end
    end
    return label
end

function M.stage_component(rolled, mitigated)
    local component = {}
    for _, channel in ipairs(M.channels) do
        component[channel .. "_rolled_raw"] = rolled[channel]
        component[channel .. "_mitigated_raw"] = mitigated[channel]
    end
    return component
end

return M
