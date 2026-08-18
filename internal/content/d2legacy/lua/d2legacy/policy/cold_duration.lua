-- Expansion 1.14d difficulty scaling for monster cold/freeze lengths.

local M = {}

local divisors = {
    [0] = 1,
    [1] = 2,
    [2] = 4,
}

function M.monster_frames(frames, difficulty)
    assert(type(frames) == "number" and frames > 0 and frames == math.floor(frames), "cold frames must be positive")
    local divisor = assert(divisors[difficulty], "unsupported difficulty")
    return math.max(math.floor(frames / divisor), 1)
end

-- Cold resistance changes effect length independently from the damage cap.
-- A raw resistance of 100 or more is cold immunity even when ordinary damage
-- mitigation is capped by maximum resistance. Players do not receive the PvM
-- Nightmare/Hell length divisor.
function M.target_frames(frames, target, difficulty, ecs)
    assert(type(frames) == "number" and frames > 0 and frames == math.floor(frames), "cold frames must be positive")
    local defense = ecs.get(target, "d2legacy.combat.defense")
    local resistance = defense and math.max(-100, math.min(100, defense:get("cold_resist"))) or 0
    if resistance >= 100 then
        return 0
    end
    local adjusted = math.floor(frames * (100 - resistance) / 100)
    if ecs.get(target, "d2legacy.monster.stats") then
        adjusted = math.floor(adjusted / assert(divisors[difficulty], "unsupported difficulty"))
    end
    return math.max(adjusted, 1)
end

return M
