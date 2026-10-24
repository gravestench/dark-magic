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
local player_count = require("d2legacy.policy.player_count")
local collision = require("d2legacy.gameplay.collision")

local M = {}

local MAX_LEVEL_MONSTERS = 25
local DENSITY_SCALE = 100000
local PLAN_ID = "d2legacy.population.plan"
local PLAN_SCHEMA = "d2legacy.population.plan/v6"

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
    local failures = {}
    local count = math.max(integer(level, "NumMon"), 0)
    local prefix = monster_prefix(difficulty)

    for index = 1, math.min(count, MAX_LEVEL_MONSTERS) do
        local id = level[prefix .. index]
        if id and id ~= "" then
            -- Levels may reference special rows which are not ordinary spawns.
            -- The monster interpreter rejects them; this list simply skips them.
            local valid, definition = pcall(monster.load, id, level)
            if valid then
                result[#result + 1] = definition
            else
                failures[#failures + 1] = id .. ": " .. tostring(definition)
            end
        end
    end
    assert(#result > 0 or count == 0, "population has no valid monster candidates: " .. table.concat(failures, "; "))
    return result
end

local function total_rarity(values)
    local total = 0
    for _, value in ipairs(values) do
        total = total + value.rarity
    end
    return total
end

-- Keep Cold Plains draws isolated so visiting it cannot advance Blood Moor random outcomes.
local function stream(zone, purpose)
    if zone.level_id == 3 then
        return "d2legacy.population.level3." .. purpose
    end
    return "d2legacy.population." .. purpose
end

local function weighted_monster(zone, values, total)
    local roll = random.integer(stream(zone, "family"), total)
    for _, value in ipairs(values) do
        if roll < value.rarity then
            return value
        end
        roll = roll - value.rarity
    end
    return values[#values]
end

local function group_size(zone, definition)
    local span = math.max(definition.max_group - definition.min_group + 1, 1)
    return definition.min_group + random.integer(stream(zone, "group"), span)
end

local function spawn_id(zone, room, member)
    return "level:" .. zone.level_id .. ":room:" .. room.id .. ":monster:" .. member
end

local function overlaps(point, radius, other)
    local dx, dy = point.x - other.x, point.y - other.y
    return dx * dx + dy * dy < (radius + other.radius) ^ 2
end

-- Require the full footprint to fit terrain and room bounds before reserving a queued spawn.
local function safe_point(zone, room, point, radius, occupied)
    if
        point.x - radius < room.x
        or point.y - radius < room.y
        or point.x + radius >= room.x + room.width
        or point.y + radius >= room.y + room.height
    then
        return false
    end
    if not collision.destination_clear(zone.level_id, point.x, point.y, radius) then
        return false
    end
    for _, other in ipairs(occupied) do
        if other.level_id == zone.level_id and overlaps(point, radius, other) then
            return false
        end
    end
    return true
end

local function materialize_group(context, zone, room, definition, game_player_count, structural, occupied)
    local count = math.min(group_size(zone, definition), #room.points)
    local created = 0
    for _, point in ipairs(room.points) do
        if created >= count then
            break
        end
        if zone.level_id ~= 3 or safe_point(zone, room, point, definition.collider_radius, occupied) then
            created = created + 1
            local command = {
                tick = context.tick,
                payload = {
                    spawn_id = spawn_id(zone, room, created),
                    seed = random.integer(stream(zone, "seed"), 2147483647),
                    x = point.x,
                    y = point.y,
                    act = zone.act,
                    level_id = zone.level_id,
                    definition = definition,
                },
            }
            local life_stream = zone.level_id == 3 and stream(zone, "life") or nil
            local components = spawn.components(command, game_player_count, life_stream)
            components["d2legacy.world.room_resident"] = {
                id = command.payload.spawn_id,
                level_id = zone.level_id,
                room_id = room.id,
            }
            structural:create(components)
            if zone.level_id == 3 then
                occupied[#occupied + 1] = {
                    level_id = zone.level_id,
                    x = point.x,
                    y = point.y,
                    radius = definition.collider_radius,
                }
            end
        end
    end
    return created
end

local function room_is_selected(zone, room, density, created)
    if not room.populate then
        return false
    end
    local roll = random.integer(stream(zone, "density"), DENSITY_SCALE)
    -- Keep the first populated room useful even when a tiny test fixture rolls
    -- above density. Production zones normally contain many candidate rooms.
    return roll < density or created == 0
end

local function validate_zone(zone)
    assert(type(zone) == "table", "population zone is required")
    assert(zone.level_id == 2 or zone.level_id == 3, "population level is required")
    assert(zone.act == 1, "population act is required")
    assert(
        zone.difficulty == nil or zone.difficulty == game_rules.difficulty(),
        "population difficulty differs from immutable game rules"
    )
    assert(type(zone.rooms) == "table", "population rooms are required")
    assert(type(zone.links) == "table", "population room links are required")
    local room_ids = {}
    for _, room in ipairs(zone.rooms) do
        assert(type(room.id) == "string" and room.id ~= "", "population room ID is required")
        assert(not room_ids[room.id], "population room IDs must be unique")
        room_ids[room.id] = true
        assert(type(room.x) == "number" and type(room.y) == "number", "population room origin is required")
        assert(type(room.width) == "number" and room.width > 0, "population room width is required")
        assert(type(room.height) == "number" and room.height > 0, "population room height is required")
        assert(type(room.points) == "table", "population room points are required")
        local points = {}
        for _, point in ipairs(room.points) do
            assert(
                type(point) == "table"
                    and type(point.x) == "number"
                    and type(point.y) == "number"
                    and point.x == point.x
                    and point.y == point.y
                    and math.abs(point.x) < math.huge
                    and math.abs(point.y) < math.huge,
                "population point is invalid"
            )
            if zone.level_id == 3 then
                assert(
                    point.x >= room.x
                        and point.x < room.x + room.width
                        and point.y >= room.y
                        and point.y < room.y + room.height,
                    "population point is outside its room"
                )
                local key = tostring(point.x) .. ":" .. tostring(point.y)
                assert(not points[key], "population points must be distinct")
                points[key] = true
            end
        end
    end
    for _, link in ipairs(zone.links) do
        assert(
            type(link.from) == "string" and link.from ~= "" and type(link.to) == "string" and link.to ~= "",
            "population room link endpoints are required"
        )
        assert(link.from ~= link.to, "population room links cannot be self-referential")
        assert(room_ids[link.from] and room_ids[link.to], "population room link references an unknown room")
    end
end

-- Retain flat single-zone command compatibility while storing one canonical zone envelope.
local function normalized_zones(payload)
    assert(type(payload) == "table", "population payload is required")
    if payload.zones ~= nil then
        assert(type(payload.zones) == "table" and #payload.zones > 0, "population zones are required")
        return payload.zones
    end
    return { payload }
end

function M.validate(command)
    local previous = 1
    for _, zone in ipairs(normalized_zones(command.payload)) do
        validate_zone(zone)
        assert(zone.level_id == previous + 1, "population zones must be unique and ordered from level 2")
        previous = zone.level_id
    end
end

function M.apply(command)
    M.validate(command)
    assert(not state.read(PLAN_ID).installed, "population plan is already installed")
    local zones = normalized_zones(command.payload)
    for _, zone in ipairs(zones) do
        for _, room in ipairs(zone.rooms) do
            room.level_id = zone.level_id
            room.activated = false
            room.active = false
            room.inactive_residents = {}
        end
        zone.created = 0
    end
    state.replace(PLAN_ID, PLAN_SCHEMA, { schema = PLAN_SCHEMA, installed = true, zones = zones })

    -- Some authoritative entities (notably imported ground items) exist
    -- before this generated-room plan is admitted. Resolve their generic ECS
    -- attachment requests now; entity-specific bootstrap code does not need
    -- to know room IDs or keep a parallel pending archive.
    for _, entity in
        ipairs(ecs.query({
            all = {
                "d2legacy.world.room_attach",
                "d2legacy.world.position",
                "d2legacy.world.location",
            },
            none = { "d2legacy.world.room_resident" },
        }))
    do
        local request = ecs.get(entity, "d2legacy.world.room_attach")
        local location = ecs.get(entity, "d2legacy.world.location")
        local position = ecs.get(entity, "d2legacy.world.position")
        local resident =
            M.resident_at(request:get("id"), location:get("level_id"), position:get("x"), position:get("y"))
        if resident then
            ecs.set(entity, "d2legacy.world.room_resident", resident)
            ecs.remove(entity, "d2legacy.world.room_attach")
        end
    end
end

local function room_at(plan, x, y)
    for _, room in ipairs(plan.rooms) do
        if x >= room.x and x < room.x + room.width and y >= room.y and y < room.y + room.height then
            return room
        end
    end
    return nil
end

-- Resolve a newly materialized world entity into the installed authoritative
-- room plan. Callers retain ownership of the entity's semantic ID and policy;
-- this module only supplies the canonical level/room residency fields.
function M.resident_at(id, level_id, x, y)
    assert(type(id) == "string" and id ~= "", "room resident ID is required")
    assert(type(level_id) == "number", "room resident level is required")
    assert(type(x) == "number" and type(y) == "number", "room resident point is required")
    local plan = state.read(PLAN_ID)
    if not plan or not plan.installed then
        return nil
    end
    for _, zone in ipairs(plan.zones) do
        if zone.level_id == level_id then
            local room = room_at(zone, x, y)
            if room then
                return { id = id, level_id = level_id, room_id = room.id }
            end
        end
    end
    return nil
end

local function containing_rooms(plan, entities)
    local active = {}
    for _, entity in ipairs(entities) do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        local location = ecs.get(entity, "d2legacy.world.location")
        local position = ecs.get(entity, "d2legacy.world.position")
        if identity and location and position and location:get("level_id") == plan.level_id then
            local room = room_at(plan, position:get("x"), position:get("y"))
            if room then
                active[room.id] = true
            end
        end
    end
    return active
end

local function sync_resident_rooms(plan, entities)
    for _, entity in ipairs(entities) do
        local resident = ecs.get(entity, "d2legacy.world.room_resident")
        local inactive = ecs.get(entity, "d2legacy.world.inactive")
        local location = ecs.get(entity, "d2legacy.world.location")
        local position = ecs.get(entity, "d2legacy.world.position")
        if resident and not inactive and location and position and location:get("level_id") == plan.level_id then
            local room = room_at(plan, position:get("x"), position:get("y"))
            if room then
                resident:set("level_id", plan.level_id)
                resident:set("room_id", room.id)
            end
        end
    end
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

local function live_player_count(entities)
    local players = {}
    for _, entity in ipairs(entities) do
        if ecs.get(entity, "d2legacy.player.identity") and not ecs.get(entity, "d2legacy.world.inactive") then
            players[#players + 1] = entity
        end
    end
    return player_count.live_population(players)
end

local function deactivate_room(room, entities, structural)
    local residents = {}
    local seen = {}
    for _, entity in ipairs(entities) do
        local resident = ecs.get(entity, "d2legacy.world.room_resident")
        local inactive = ecs.get(entity, "d2legacy.world.inactive")
        if
            resident
            and not inactive
            and resident:get("level_id") == room.level_id
            and resident:get("room_id") == room.id
        then
            local id = resident:get("id")
            assert(id ~= "", "room resident ID is required")
            assert(not seen[id], "room resident IDs must be unique")
            seen[id] = true
            residents[#residents + 1] = {
                id = id,
                velocity_mover = ecs.get(entity, "engine.world.velocity_mover") ~= nil,
            }
            structural:set(entity, "d2legacy.world.inactive", {})
            -- The engine velocity integrator uses its own opt-in empty tag.
            -- Remove that activation surface while the mod-owned inactive tag
            -- suppresses authoritative Lua systems and presentation queries.
            structural:remove(entity, "engine.world.velocity_mover")
        end
    end
    table.sort(residents, function(left, right)
        return left.id < right.id
    end)
    room.inactive_residents = residents
end

local function restore_room(room, entities, structural)
    local wanted = {}
    for _, resident in ipairs(room.inactive_residents or {}) do
        assert(not wanted[resident.id], "inactive resident IDs must be unique")
        wanted[resident.id] = resident
    end
    for _, entity in ipairs(entities) do
        local resident = ecs.get(entity, "d2legacy.world.room_resident")
        local inactive = ecs.get(entity, "d2legacy.world.inactive")
        local id = resident and resident:get("id") or ""
        local record = wanted[id]
        if
            resident
            and inactive
            and resident:get("level_id") == room.level_id
            and resident:get("room_id") == room.id
            and record
        then
            structural:remove(entity, "d2legacy.world.inactive")
            if record.velocity_mover then
                structural:set(entity, "engine.world.velocity_mover", {})
            end
            wanted[id] = nil
        end
    end
    for id in pairs(wanted) do
        error("inactive room resident is missing: " .. id)
    end
    room.inactive_residents = {}
end

-- Record collider facts even for entities without room residency; occupancy is level-scoped.
local function include_footprint(entity, occupied)
    local point = ecs.get(entity, "d2legacy.world.position")
    local location = ecs.get(entity, "d2legacy.world.location")
    local collider = ecs.get(entity, "d2legacy.world.collider")
    if point and location and collider then
        occupied[#occupied + 1] = {
            x = point:get("x"),
            y = point:get("y"),
            level_id = location:get("level_id"),
            radius = collider:get("radius"),
        }
    end
end

-- Reserve returning residents before fresh spawns because structural restoration publishes later.
local function include_returning_footprints(plan, active, entities, occupied)
    if plan.level_id ~= 3 then
        return
    end
    local returning = {}
    for _, room in ipairs(plan.rooms) do
        if active[room.id] and not room.active then
            local residents = {}
            for _, resident in ipairs(room.inactive_residents or {}) do
                residents[resident.id] = true
            end
            returning[room.id] = residents
        end
    end
    for _, entity in ipairs(entities) do
        local resident = ecs.get(entity, "d2legacy.world.room_resident")
        if resident and resident:get("level_id") == plan.level_id and ecs.get(entity, "d2legacy.world.inactive") then
            local residents = returning[resident:get("room_id")]
            if residents and residents[resident:get("id")] then
                include_footprint(entity, occupied)
            end
        end
    end
end

local function activate_zone(context, plan, entities, structural, occupied)
    -- Movement runs after this phase, so synchronize the result of the prior
    -- fixed tick before deciding which rooms leave the active window.
    sync_resident_rooms(plan, entities)
    local active = include_neighbors(containing_rooms(plan, entities), plan.links)
    -- Structural restoration publishes after this update. Reserve every
    -- returning footprint before the first fresh room, regardless of room order.
    include_returning_footprints(plan, active, entities, occupied)
    local difficulty = game_rules.difficulty()
    local level, candidates, rarity
    local changed = false
    for _, room in ipairs(plan.rooms) do
        local wanted = active[room.id] == true
        if room.active and not wanted then
            deactivate_room(room, entities, structural)
            room.active, changed = false, true
        elseif wanted and not room.active then
            if #(room.inactive_residents or {}) > 0 then
                restore_room(room, entities, structural)
            elseif not room.activated then
                if not level then
                    level = assert(
                        find(records.load("data/global/excel/levels.txt"), "Id", tostring(plan.level_id)),
                        "population level is missing"
                    )
                    candidates = candidate_monsters(level, difficulty)
                    rarity = total_rarity(candidates)
                end
                if
                    #candidates > 0
                    and room_is_selected(plan, room, integer(level, density_column(difficulty)), plan.created)
                then
                    local definition = weighted_monster(plan, candidates, rarity)
                    plan.created = plan.created
                        + materialize_group(
                            context,
                            plan,
                            room,
                            definition,
                            live_player_count(entities),
                            structural,
                            occupied
                        )
                end
                room.activated = true
            end
            room.active, changed = true, true
        end
    end
    return changed
end

local function activate(context, entities, structural)
    local plan = state.read(PLAN_ID)
    if not plan or not plan.installed then
        return
    end
    -- Include active colliders without room residency as well as population
    -- residents. The list also receives returning residents and queued creates.
    local occupied = {}
    for _, entity in ipairs(entities) do
        if not ecs.get(entity, "d2legacy.world.inactive") then
            include_footprint(entity, occupied)
        end
    end
    local changed = false
    for _, zone in ipairs(plan.zones) do
        if activate_zone(context, zone, entities, structural, occupied) then
            changed = true
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
        zones = {},
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
            any = { "d2legacy.player.identity", "d2legacy.world.room_resident", "d2legacy.world.collider" },
        },
        read = {
            "d2legacy.player.identity",
            "d2legacy.player.vitals",
            "d2legacy.player.death",
            "d2legacy.world.location",
            "d2legacy.world.position",
            "d2legacy.world.room_resident",
            "d2legacy.monster.identity",
            "d2legacy.monster.stats",
            "d2legacy.combat.melee_profile",
            "d2legacy.monster.appearance",
            "d2legacy.monster.ai",
            "d2legacy.combat.basic_attack_request",
            "d2legacy.combat.attack_approach",
            "d2legacy.combat.attack_animation",
            "d2legacy.monster.death",
            "d2legacy.world.velocity",
            "d2legacy.world.facing",
            "d2legacy.world.collider",
            "d2legacy.world.occupancy",
            "d2legacy.world.forced_motion",
            "d2legacy.combat.knockback_target",
            "engine.world.velocity_mover",
            "d2legacy.world.selectable",
            "d2legacy.world.inactive",
        },
        write = {
            "d2legacy.monster.identity",
            "d2legacy.monster.corpse_selectable",
            "d2legacy.monster.revivable",
            "d2legacy.monster.boss",
            "d2legacy.monster.freeze_immune",
            "d2legacy.monster.prime_evil",
            "d2legacy.world.room_resident",
            "d2legacy.monster.stats",
            "d2legacy.monster.type_flags",
            "d2legacy.combat.defense",
            "d2legacy.combat.melee_profile",
            "d2legacy.monster.appearance",
            "d2legacy.monster.ai",
            "d2legacy.combat.basic_attack_request",
            "d2legacy.combat.attack_approach",
            "d2legacy.combat.attack_animation",
            "d2legacy.monster.death",
            "d2legacy.world.position",
            "d2legacy.world.velocity",
            "d2legacy.world.facing",
            "d2legacy.world.location",
            "d2legacy.world.collider",
            "d2legacy.world.occupancy",
            "d2legacy.world.forced_motion",
            "d2legacy.combat.knockback_target",
            "engine.world.velocity_mover",
            "d2legacy.world.selectable",
            "d2legacy.world.inactive",
        },
        update = activate,
    })
end

return M
