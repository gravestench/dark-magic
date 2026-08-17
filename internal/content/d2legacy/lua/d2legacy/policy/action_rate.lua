-- Convert resolved attack-rate facts into one deterministic animation schedule.
--
-- This is intentionally skill-agnostic. Any Skills.txt behavior with
-- UseAttackRate may pass its pinned AnimData record through the same policy.

local M = {}

local function trunc_div(numerator, denominator)
    assert(denominator > 0, "action-rate divisor must be positive")
    if numerator < 0 then
        return math.ceil(numerator / denominator)
    end
    return math.floor(numerator / denominator)
end

local function clamp(value, minimum, maximum)
    return math.max(minimum, math.min(maximum, value))
end

function M.effective_item_rate(item_fasterattackrate)
    item_fasterattackrate = item_fasterattackrate or 0
    if item_fasterattackrate == 0 then
        return 0
    end
    assert(item_fasterattackrate > -120, "item_fasterattackrate must be greater than -120")
    return trunc_div(120 * item_fasterattackrate, item_fasterattackrate + 120)
end

function M.rate_percent(facts)
    facts = facts or {}
    local primary = facts.primary_weapon_attack_rate or 0
    local rate = (facts.attack_rate or 100) + M.effective_item_rate(facts.item_fasterattackrate)
    if facts.dual_wield then
        local secondary = facts.secondary_weapon_attack_rate or 0
        rate = rate + trunc_div(primary + secondary, 2) - primary
    end
    if facts.sequence then
        rate = rate - 30
    end
    return clamp(rate, 15, 175)
end

function M.animation_speed(base_speed, rate_percent)
    assert(type(base_speed) == "number" and base_speed > 0, "base animation speed must be positive")
    return clamp(trunc_div(base_speed * rate_percent, 100), 1, 32767)
end

local function tick_for_cursor(cursor, speed)
    if cursor <= 0 then
        return 1
    end
    return math.ceil(cursor / speed)
end

function M.schedule(timing, facts)
    assert(type(timing) == "table", "AnimData timing is required")
    local rate = M.rate_percent(facts)
    local speed = M.animation_speed(timing.speed, rate)
    local attack_delay
    for _, event in ipairs(timing.events or {}) do
        if event.kind == "attack" then
            local delay = tick_for_cursor(event.frame * timing.frame_scale, speed)
            attack_delay = attack_delay and math.min(attack_delay, delay) or delay
        end
    end
    assert(attack_delay, "AnimData record " .. tostring(timing.name) .. " has no attack event")
    local complete_delay = tick_for_cursor(timing.frames * timing.frame_scale, speed)
    assert(attack_delay < complete_delay, "AnimData attack event must precede completion for " .. timing.name)
    return {
        attack_delay = attack_delay,
        complete_delay = complete_delay,
        animation_speed = speed,
        rate_percent = rate,
    }
end

return M
