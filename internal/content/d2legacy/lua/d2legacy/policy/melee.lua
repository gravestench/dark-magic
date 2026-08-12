-- Diablo II 1.10f melee hit policy, preserving the recovered integer order.

local random = require("engine.authority_random/v1")
local M = {}

-- Melee range is clearance between the physical footprints of two actors, not
-- center-to-center distance. Selectable radii describe pointer affordances and
-- may differ from collision footprints, so callers pass collider radii here.
function M.reach(attack_range, attacker_radius, target_radius)
    assert(attack_range >= 0, "melee range must be non-negative")
    assert(attacker_radius >= 0 and target_radius >= 0,
        "melee collider radii must be non-negative")
    return attack_range + attacker_radius + target_radius
end

function M.hit_chance(attacker_level, defender_level, attack_rating, defense)
    assert(attacker_level > 0 and defender_level > 0, "unit levels must be positive")
    if defense < 0 then attack_rating, defense = attack_rating-defense, 0 end
    if attack_rating < 0 then defense, attack_rating = defense-attack_rating, 0 end
    if defense < 0 then defense = 0 end
    local rating_factor = 100
    if attack_rating + defense ~= 0 then
        rating_factor = math.floor(100 * attack_rating / (attack_rating + defense))
    end
    local chance = math.floor(2 * attacker_level * rating_factor / (attacker_level + defender_level))
    return math.max(5, math.min(95, chance))
end

function M.hits(attacker_level, defender_level, attack_rating, defense)
    local chance = M.hit_chance(attacker_level, defender_level, attack_rating, defense)
    return random.integer("d2legacy.combat.basic_melee.hit", 100)
        < chance
end

function M.damage(minimum, maximum)
    assert(minimum >= 0 and maximum >= minimum, "invalid melee damage range")
    return minimum + random.integer(
        "d2legacy.combat.basic_melee.damage", maximum - minimum + 1)
end

return M
