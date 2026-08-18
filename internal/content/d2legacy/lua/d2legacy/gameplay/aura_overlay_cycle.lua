-- Select the one aura graphic shown for each affected unit.
--
-- Gameplay keeps every distinct aura relationship active at once. Expansion
-- presentation alternates their state graphics instead of drawing every
-- ground overlay on top of the unit. Each aura's owned Skills row supplies its
-- periodic delay; converting that 25 Hz delay to seconds keeps the visual
-- handoff aligned with the record-driven aura pulse without changing gameplay.

local M = {}
local TICKS_PER_SECOND = 25

local function sorted_keys(values)
    local result = {}
    for key in pairs(values) do
        result[#result + 1] = key
    end
    table.sort(result)
    return result
end

local function compare(left, right)
    if left.state_id ~= right.state_id then
        return left.state_id < right.state_id
    end
    return left.entity_id < right.entity_id
end

local function signature(entries)
    local values = {}
    for _, entry in ipairs(entries) do
        values[#values + 1] = entry.state_id
    end
    return table.concat(values, "\0")
end

local function duration(entry)
    local ticks = tonumber(entry.aura_period_ticks)
    assert(ticks and ticks >= 1, "aura snapshot has no positive record period")
    return ticks / TICKS_PER_SECOND
end

local function preserved_index(entries, previous)
    if previous then
        for index, entry in ipairs(entries) do
            if entry.state_id == previous.state_id then
                return index
            end
        end
    end
    return 1
end

function M.select(snapshots, previous, elapsed)
    assert(type(snapshots) == "table", "state snapshots are required")
    previous = previous or {}
    elapsed = math.max(tonumber(elapsed) or 0, 0)
    local selected, groups, next_cycles = {}, {}, {}
    for _, snapshot in ipairs(snapshots) do
        if snapshot.aura then
            local key = tostring(snapshot.target_entity_id)
            groups[key] = groups[key] or {}
            groups[key][#groups[key] + 1] = snapshot
        else
            selected[#selected + 1] = snapshot
        end
    end

    for _, key in ipairs(sorted_keys(groups)) do
        local entries = groups[key]
        table.sort(entries, compare)
        local current_signature = signature(entries)
        local old = previous[key]
        local index = preserved_index(entries, old)
        local remaining = old and old.signature == current_signature and old.remaining or duration(entries[index])
        remaining = remaining - elapsed
        while remaining <= 0 do
            index = index % #entries + 1
            remaining = remaining + duration(entries[index])
        end
        local chosen = entries[index]
        next_cycles[key] = {
            signature = current_signature,
            state_id = chosen.state_id,
            remaining = remaining,
        }
        selected[#selected + 1] = chosen
    end

    table.sort(selected, function(left, right)
        return left.entity_id < right.entity_id
    end)
    return selected, next_cycles
end

return M
