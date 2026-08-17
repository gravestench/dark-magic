-- Purpose-specific multiplayer counts. These values are deliberately not one
-- generic multiplier: each Diablo consumer selects only the context it needs.

local game_rules = require("d2legacy.policy.game_rules")
local state = require("engine.authority_state/v1")

local M = {}
local STATE_ID = "d2legacy.player_count"
local STATE_SCHEMA = "d2legacy.player_count/v1"
local MAXIMUM_GAMEPLAY_COUNT = 8

local function integer(value, name, minimum, maximum)
    assert(type(value) == "number" and value == math.floor(value), name .. " must be an integer")
    assert(value >= minimum and value <= maximum, name .. " is outside its supported range")
    return value
end

local function current()
    return assert(state.read(STATE_ID), "player-count state is not initialized")
end

local function commit(value)
    value.revision = value.revision + 1
    state.replace(STATE_ID, STATE_SCHEMA, value)
end

function M.initialize()
    state.register(STATE_ID, STATE_SCHEMA, {
        schema = STATE_SCHEMA,
        revision = 0,
    })
end

function M.snapshot()
    return current()
end

function M.set_override(value)
    integer(value, "player-count override", 1, MAXIMUM_GAMEPLAY_COUNT)
    local count = current()
    count.override = value
    commit(count)
end

function M.clear_override()
    local count = current()
    if count.override ~= nil then
        count.override = nil
        commit(count)
    end
end

function M.effective(game_player_count)
    local rules = game_rules.get()
    integer(game_player_count, "game player count", 1, rules.maximum_players)
    local overridden = current().override
    if overridden ~= nil then
        return integer(overridden, "player-count override", 1, MAXIMUM_GAMEPLAY_COUNT)
    end
    return game_player_count
end

function M.monster_spawn(game_player_count, evil)
    local rules = game_rules.get()
    integer(game_player_count, "game player count", 1, rules.maximum_players)
    assert(type(evil) == "boolean", "monster alignment is required")

    local effective = evil and M.effective(game_player_count) or 1
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
    integer(monster_player_count, "monster player count", 1, MAXIMUM_GAMEPLAY_COUNT)

    local effective = M.effective(game_player_count)
    local party_members = nearby_party_member_count + 1
    local eligible_party_members = math.min(party_members, effective)
    local eligible = math.floor((effective - eligible_party_members) / 2) + eligible_party_members
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
