-- Recovered Diablo II melee arithmetic, preserving the observed integer order.
-- Expansion 1.14d is the only product target; owned-runtime vectors still gate
-- an exact-version claim for this formula.

local random = require("engine.authority_random/v1")
local M = {}
local U32 = 4294967296
local I32_SIGN = 2147483648

local function signed32(value)
    local wrapped = value % U32
    if wrapped >= I32_SIGN then
        return wrapped - U32
    end
    return wrapped
end

local function truncate(value)
    return value < 0 and math.ceil(value) or math.floor(value)
end

-- Melee range is clearance between the physical footprints of two actors, not
-- center-to-center distance. Selectable radii describe pointer affordances and
-- may differ from collision footprints, so callers pass collider radii here.
function M.reach(attack_range, attacker_radius, target_radius)
    assert(attack_range >= 0, "melee range must be non-negative")
    assert(attacker_radius >= 0 and target_radius >= 0,
        "melee collider radii must be non-negative")
    return attack_range + attacker_radius + target_radius
end

function M.effective_defense(defense, ignore_target_ac, attacker_is_player, defender_is_monster, boss, hireling)
    if attacker_is_player and defender_is_monster and ignore_target_ac ~= 0 and not boss and not hireling then
        return 0
    end
    return defense
end

function M.fractional_defense(defense, value, half)
    if value <= 0 then
        return defense
    end
    if half then
        value = truncate(value / 2)
    end
    local percent = math.min(value, 100)
    local cut
    if defense > 1048576 then
        cut = signed32(truncate(defense / 100) * percent)
    else
        cut = truncate(signed32(defense * percent) / 100)
    end
    return signed32(defense - cut)
end

function M.target_type_attack_rating(
    attack_rating,
    demon_bonus,
    undead_bonus,
    player_attacker,
    monster_defender,
    demon,
    low_undead,
    high_undead
)
    if not player_attacker or not monster_defender then
        return attack_rating
    end
    if demon and demon_bonus ~= 0 then
        attack_rating = signed32(attack_rating + demon_bonus)
    end
    if (low_undead or high_undead) and undead_bonus ~= 0 then
        attack_rating = signed32(attack_rating + undead_bonus)
    end
    return attack_rating
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
