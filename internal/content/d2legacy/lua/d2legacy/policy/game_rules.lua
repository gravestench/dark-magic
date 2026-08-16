-- Immutable expansion 1.14d rules shared by every authoritative consumer.

local initial = require("engine.initial_data/v1")
local state = require("engine.authority_state/v1")

local M = {}
local rules

local function copy(value)
    if type(value) ~= "table" then
        return value
    end
    local result = {}
    for key, item in pairs(value) do
        result[key] = copy(item)
    end
    return result
end

local function integer(value, name, minimum, maximum)
    assert(type(value) == "number" and value == math.floor(value), name .. " must be an integer")
    assert(value >= minimum and value <= maximum, name .. " is outside its supported range")
    return value
end

local function build()
    local configured = initial.get("d2legacy.game_rules") or {}
    local generation = initial.get("engine.game_data_generation_id") or "synthetic"
    local result = {
        schema = "d2legacy.game_rules/v1",
        target = configured.target or "lod-1.14d",
        expansion = configured.expansion ~= false,
        difficulty = configured.difficulty or 0,
        hardcore = configured.hardcore == true,
        ladder = configured.ladder == true,
        player_count = configured.player_count or 1,
        maximum_players = configured.maximum_players or 8,
        game_data_generation_id = generation,
    }
    assert(result.target == "lod-1.14d", "only expansion 1.14d rules are supported")
    assert(result.expansion, "Classic rules are unsupported")
    integer(result.difficulty, "difficulty", 0, 2)
    integer(result.player_count, "player count", 1, 8)
    integer(result.maximum_players, "maximum players", 1, 8)
    assert(result.player_count <= result.maximum_players, "player count exceeds game capacity")
    assert(
        type(result.game_data_generation_id) == "string" and result.game_data_generation_id ~= "",
        "game-data generation is required"
    )
    return result
end

function M.initialize()
    assert(rules == nil, "game rules are already initialized")
    rules = build()
    state.register("d2legacy.game_rules", rules.schema, rules)
end

function M.get()
    assert(rules, "game rules are not initialized")
    return copy(rules)
end

function M.difficulty()
    assert(rules, "game rules are not initialized")
    return rules.difficulty
end

function M.effective_player_count(game_player_count)
    assert(rules, "game rules are not initialized")
    integer(game_player_count, "game player count", 1, rules.maximum_players)
    return math.max(game_player_count, rules.player_count)
end

return M
