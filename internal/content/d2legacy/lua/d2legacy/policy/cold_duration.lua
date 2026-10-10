-- Expansion 1.14d difficulty scaling for monster cold/freeze lengths.

local difficulty_levels = require("d2legacy.data.difficulty_levels")

local M = {}

-- Freeze and chill read separate DifficultyLevels divisors for the session's
-- fixed difficulty. Every other cold state uses the chill divisor.
local function monster_divisor(state_id)
    if state_id == "freeze" then
        return difficulty_levels.monster_freeze_divisor()
    end
    return difficulty_levels.monster_cold_divisor()
end

function M.monster_frames(frames, state_id)
    assert(type(frames) == "number" and frames > 0 and frames == math.floor(frames), "cold frames must be positive")
    return math.max(math.floor(frames / monster_divisor(state_id)), 1)
end

-- Cold resistance changes effect length independently from the damage cap.
-- A raw resistance of 100 or more is cold immunity even when ordinary damage
-- mitigation is capped by maximum resistance. Players do not receive the PvM
-- Nightmare/Hell length divisor.
function M.target_frames(frames, target, state_id, ecs)
    assert(type(frames) == "number" and frames > 0 and frames == math.floor(frames), "cold frames must be positive")
    local defense = ecs.get(target, "d2legacy.combat.defense")
    local resistance = defense and math.max(-100, math.min(100, defense:get("cold_resist"))) or 0
    if resistance >= 100 then
        return 0
    end
    local adjusted = math.floor(frames * (100 - resistance) / 100)
    if ecs.get(target, "d2legacy.monster.stats") then
        adjusted = math.floor(adjusted / monster_divisor(state_id))
    end
    return math.max(adjusted, 1)
end

return M
