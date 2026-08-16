-- Populate a generated Diablo area from immutable geometry and legacy rows.
--
-- The host supplies only candidate rooms and open points. This mod owns every
-- game decision: level density, eligible monster families, rarity, group size,
-- and the authoritative monster definition materialized at each point.

local records = require("engine.records/v1")
local random = require("engine.authority_random/v1")
local commands = require("engine.authority_command/v1")
local monster = require("d2legacy.data.monster")
local spawn = require("d2legacy.commands.spawn_monster")
local game_rules = require("d2legacy.policy.game_rules")

local M = {}

local MAX_LEVEL_MONSTERS = 25
local DENSITY_SCALE = 100000

local function integer(row, key, fallback)
    return math.floor(tonumber(row[key]) or fallback or 0)
end

local function find(rows, key, wanted)
    for _, row in ipairs(rows) do
        if row[key] == wanted then return row end
    end
    return nil
end

local function density_column(difficulty)
    if difficulty == 2 then return "MonDen(H)" end
    if difficulty == 1 then return "MonDen(N)" end
    return "MonDen"
end

local function monster_prefix(difficulty)
    if difficulty == 0 then return "mon" end
    return "nmon"
end

local function candidate_monsters(level, difficulty)
    local result = {}
    local count = math.max(integer(level, "NumMon"), 0)
    local prefix = monster_prefix(difficulty)

    for index = 1, math.min(count, MAX_LEVEL_MONSTERS) do
        local id = level[prefix .. index]
        if id and id ~= "" then
            -- Levels may reference special rows which are not ordinary spawns.
            -- The monster interpreter rejects them; this list simply skips them.
            local valid, definition = pcall(monster.load, id, difficulty)
            if valid then result[#result + 1] = definition end
        end
    end
    return result
end

local function total_rarity(values)
    local total = 0
    for _, value in ipairs(values) do total = total + value.rarity end
    return total
end

local function weighted_monster(values, total)
    local roll = random.integer("d2legacy.population.family", total)
    for _, value in ipairs(values) do
        if roll < value.rarity then return value end
        roll = roll - value.rarity
    end
    return values[#values]
end

local function group_size(definition)
    local span = math.max(
        definition.max_group - definition.min_group + 1,
        1
    )
    return definition.min_group
        + random.integer("d2legacy.population.group", span)
end

local function spawn_id(zone, room, member)
    return "level:" .. zone.level_id
        .. ":room:" .. room.id
        .. ":monster:" .. member
end

local function materialize_group(command, zone, room, definition)
    local count = math.min(group_size(definition), #room.points)
    for member = 1, count do
        local point = room.points[member]
        spawn.materialize({
            tick = command.tick,
            payload = {
                spawn_id = spawn_id(zone, room, member),
                seed = random.integer("d2legacy.population.seed", 2147483647),
                x = point.x,
                y = point.y,
                act = zone.act,
                level_id = zone.level_id,
                definition = definition,
            },
        })
    end
    return count
end

local function room_is_selected(room, density, created)
    if not room.populate then return false end
    local roll = random.integer("d2legacy.population.density", DENSITY_SCALE)
    -- Keep the first populated room useful even when a tiny test fixture rolls
    -- above density. Production zones normally contain many candidate rooms.
    return roll < density or created == 0
end

function M.validate(command)
    local zone = command.payload
    assert(type(zone) == "table", "population zone is required")
    assert(type(zone.level_id) == "number", "population level is required")
    assert(type(zone.rooms) == "table", "population rooms are required")
end

function M.apply(command)
    local zone = command.payload
    if zone.level_id ~= 2 then return end

    local difficulty = game_rules.difficulty()
    assert(zone.difficulty == nil or zone.difficulty == difficulty,
        "population difficulty differs from immutable game rules")

    local level = assert(find(
        records.load("data/global/excel/levels.txt"),
        "Id",
        tostring(zone.level_id)
    ), "population level is missing")
    local density = integer(level, density_column(difficulty))
    local candidates = candidate_monsters(level, difficulty)
    if #candidates == 0 then return end

    local rarity = total_rarity(candidates)
    local created = 0
    for _, room in ipairs(zone.rooms) do
        if room_is_selected(room, density, created) then
            local definition = weighted_monster(candidates, rarity)
            created = created + materialize_group(
                command,
                zone,
                room,
                definition
            )
        end
    end
end

function M.register()
    commands.register({
        kind = "system.population.bootstrap",
        authorities = { "system" },
        validate = M.validate,
        apply = M.apply,
    })
end

return M
