-- Purpose-specific multiplayer counts. These values are deliberately not one
-- generic multiplier: each Diablo consumer selects only the context it needs.

local game_rules = require("d2legacy.policy.game_rules")

local M = {}

local function integer(value, name, minimum, maximum)
    assert(type(value) == "number" and value == math.floor(value), name .. " must be an integer")
    assert(value >= minimum and value <= maximum, name .. " is outside its supported range")
    return value
end

-- nearby_party_member_count excludes the credited player. Party authority will
-- populate it once proximity and same-level membership are implemented; zero
-- is the only valid pre-party value and must remain explicit at call sites.
function M.no_drop(game_player_count, nearby_party_member_count)
    local rules = game_rules.get()
    integer(game_player_count, "game player count", 1, rules.maximum_players)
    integer(nearby_party_member_count, "nearby party member count", 0, rules.maximum_players - 1)

    return {
        game_player_count = game_player_count,
        effective_player_count = game_rules.effective_player_count(game_player_count),
        nearby_party_member_count = nearby_party_member_count,
    }
end

return M
