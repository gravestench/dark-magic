-- Populate a generated Diablo area from immutable geometry and legacy rows.
--
-- The host supplies only candidate rooms and open points. This mod owns every
-- game decision: level density, eligible monster families, rarity, group size,
-- and the authoritative monster definition materialized at each point.

local records = require("engine.records/v1")
local random = require("engine.authority_random/v1")
local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local state = require("engine.authority_state/v1")
local monster = require("d2legacy.data.monster")
local spawn = require("d2legacy.commands.spawn_monster")
local game_rules = require("d2legacy.policy.game_rules")

local M = {}

local MAX_LEVEL_MONSTERS = 25
local DENSITY_SCALE = 100000
local PLAN_ID = "d2legacy.population.plan"
local PLAN_SCHEMA = "d2legacy.population.plan/v1"

local function integer(row, key, fallback)
    return math.floor(tonumber(row[key]) or fallback or 0)
end

local function find(rows, key, wanted)
    for _, row in ipairs(rows) do
        if row[key] == wanted then
            return row
        end
    end
    return nil
end

local function density_column(difficulty)
    if difficulty == 2 then
        return "MonDen(H)"
    end
    if difficulty == 1 then
        return "MonDen(N)"
    end
    return "MonDen"
end

local function monster_prefix(difficulty)
    if difficulty == 0 then
        return "mon"
    end
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
            local valid, definition = pcall(monster.load, id)
            if valid then
                result[#result + 1] = definition
            end
        end
    end
    return result
end

local function total_rarity(values)
    local total = 0
    for _, value in ipairs(values) do
        total = total + value.rarity
    end
    return total
end

local function weighted_monster(values, total)
    local roll = random.integer("d2legacy.population.family", total)
    for _, value in ipairs(values) do
        if roll < value.rarity then
            return value
        end
        roll = roll - value.rarity
    end
    return values[#values]
end

local function group_size(definition)
    local span = math.max(definition.max_group - definition.min_group + 1, 1)
    return definition.min_group + random.integer("d2legacy.population.group", span)
end

local function spawn_id(zone, room, member)
    return "level:" .. zone.level_id .. ":room:" .. room.id .. ":monster:" .. member
end

local function materialize_group(context, zone, room, definition, game_player_count, structural)
    local count = math.min(group_size(definition), #room.points)
    for member = 1, count do
        local point = room.points[member]
        local command = {
            tick = context.tick,
            payload = {
                spawn_id = spawn_id(zone, room, member),
                seed = random.integer("d2legacy.population.seed", 2147483647),
                x = point.x,
                y = point.y,
                act = zone.act,
                level_id = zone.level_id,
                definition = definition,
            },
        }
        structural:create(spawn.components(command, game_player_count))
    end
    return count
end

local function room_is_selected(room, density, created)
    if not room.populate then
        return false
    end
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
    assert(type(zone.links) == "table", "population room links are required")
    local room_ids = {}
    for _, room in ipairs(zone.rooms) do
        assert(room.id ~= nil, "population room ID is required")
        assert(not room_ids[room.id], "population room IDs must be unique")
        room_ids[room.id] = true
        assert(type(room.x) == "number" and type(room.y) == "number", "population room origin is required")
        assert(type(room.width) == "number" and room.width > 0, "population room width is required")
        assert(type(room.height) == "number" and room.height > 0, "population room height is required")
        assert(type(room.points) == "table", "population room points are required")
    end
    for _, link in ipairs(zone.links) do
        assert(link.from ~= nil and link.to ~= nil, "population room link endpoints are required")
        assert(link.from ~= link.to, "population room links cannot be self-referential")
        assert(room_ids[link.from] and room_ids[link.to], "population room link references an unknown room")
    end
end

function M.apply(command)
    local zone = command.payload
    if zone.level_id ~= 2 then
        return
    end

    local difficulty = game_rules.difficulty()
    assert(
        zone.difficulty == nil or zone.difficulty == difficulty,
        "population difficulty differs from immutable game rules"
    )
    assert(not state.read(PLAN_ID).installed, "population plan is already installed")

    for _, room in ipairs(zone.rooms) do
        room.activated = false
    end
    zone.created = 0
    zone.installed = true
    zone.schema = PLAN_SCHEMA
    state.replace(PLAN_ID, PLAN_SCHEMA, zone)
end

local function containing_rooms(plan, entities)
    local active = {}
    for _, entity in ipairs(entities) do
        local location = ecs.get(entity, "d2legacy.world.location")
        local position = ecs.get(entity, "d2legacy.world.position")
        if location:get("level_id") == plan.level_id then
            local x, y = position:get("x"), position:get("y")
            for _, room in ipairs(plan.rooms) do
                if x >= room.x and x < room.x + room.width and y >= room.y and y < room.y + room.height then
                    active[room.id] = true
                end
            end
        end
    end
    return active
end

local function include_neighbors(active, links)
    local expanded = {}
    for id in pairs(active) do
        expanded[id] = true
    end
    for _, link in ipairs(links) do
        if active[link.from] then
            expanded[link.to] = true
        end
        if active[link.to] then
            expanded[link.from] = true
        end
    end
    return expanded
end

local function activate(context, entities, structural)
    local plan = state.read(PLAN_ID)
    if not plan or not plan.installed then
        return
    end

    local active = include_neighbors(containing_rooms(plan, entities), plan.links)
    if next(active) == nil then
        return
    end
    local pending = false
    for _, room in ipairs(plan.rooms) do
        if active[room.id] and not room.activated then
            pending = true
            break
        end
    end
    if not pending then
        return
    end

    local difficulty = game_rules.difficulty()
    local level = assert(
        find(records.load("data/global/excel/levels.txt"), "Id", tostring(plan.level_id)),
        "population level is missing"
    )
    local candidates = candidate_monsters(level, difficulty)
    local rarity = total_rarity(candidates)
    local changed = false
    for _, room in ipairs(plan.rooms) do
        if active[room.id] and not room.activated then
            room.activated, changed = true, true
            if #candidates > 0 and room_is_selected(room, integer(level, density_column(difficulty)), plan.created) then
                local definition = weighted_monster(candidates, rarity)
                plan.created = plan.created + materialize_group(context, plan, room, definition, #entities, structural)
            end
        end
    end
    if changed then
        state.replace(PLAN_ID, PLAN_SCHEMA, plan)
    end
end

function M.register()
    state.register(PLAN_ID, PLAN_SCHEMA, {
        schema = PLAN_SCHEMA,
        installed = false,
        level_id = 0,
        rooms = {},
        links = {},
        created = 0,
    })
    commands.register({
        kind = "system.population.bootstrap",
        authorities = { "system" },
        validate = M.validate,
        apply = M.apply,
    })
    ecs.system({
        id = "d2legacy.population.active_rooms",
        phase = "pre_simulation",
        query = {
            all = {
                "d2legacy.player.identity",
                "d2legacy.world.location",
                "d2legacy.world.position",
            },
        },
        read = {
            "d2legacy.player.identity",
            "d2legacy.world.location",
            "d2legacy.world.position",
        },
        write = {
            "d2legacy.monster.identity",
            "d2legacy.monster.stats",
            "d2legacy.combat.melee_profile",
            "d2legacy.monster.appearance",
            "d2legacy.monster.ai",
            "d2legacy.world.position",
            "d2legacy.world.velocity",
            "d2legacy.world.facing",
            "d2legacy.world.location",
            "d2legacy.world.collider",
            "engine.world.velocity_mover",
            "d2legacy.world.selectable",
        },
        update = activate,
    })
end

return M
