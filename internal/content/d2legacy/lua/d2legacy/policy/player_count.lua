-- Purpose-specific multiplayer counts. These values are deliberately not one
-- generic multiplier: each Diablo consumer selects only the context it needs.

local game_rules = require("d2legacy.policy.game_rules")

local M = {}

local function integer(value, name, minimum, maximum)
    assert(type(value) == "number" and value == math.floor(value), name .. " must be an integer")
    assert(value >= minimum and value <= maximum, name .. " is outside its supported range")
    return value
end

function M.monster_spawn(game_player_count, evil)
    local rules = game_rules.get()
    integer(game_player_count, "game player count", 1, rules.maximum_players)
    assert(type(evil) == "boolean", "monster alignment is required")

    local effective = evil and game_rules.effective_player_count(game_player_count) or 1
    local bonus = 50 * (effective - 1)
    return {
        game_player_count = game_player_count,
        effective_player_count = effective,
        life_bonus_percent = bonus,
        experience_bonus_percent = bonus,
    }
end

-- nearby_party_member_count excludes the credited player. The current recovered
-- consumer supplies living party members in the same level; narrower spatial
-- eligibility remains target-version probe work for consumers that require it.
-- monster_player_count is pinned when that monster is spawned and caps later
-- NoDrop benefits even if additional players enter the game before its death.
function M.no_drop(game_player_count, nearby_party_member_count, monster_player_count)
    local rules = game_rules.get()
    integer(game_player_count, "game player count", 1, rules.maximum_players)
    integer(nearby_party_member_count, "nearby party member count", 0, rules.maximum_players - 1)
    integer(monster_player_count, "monster player count", 1, rules.maximum_players)

    local effective = game_rules.effective_player_count(game_player_count)
    local party_members = nearby_party_member_count + 1
    assert(party_members <= effective, "nearby party count exceeds effective player count")
    local eligible = math.floor((effective - party_members) / 2) + party_members
    eligible = math.min(eligible, monster_player_count)

    return {
        game_player_count = game_player_count,
        effective_player_count = effective,
        nearby_party_member_count = nearby_party_member_count,
        monster_player_count = monster_player_count,
        no_drop_player_count = eligible,
    }
end

return M
