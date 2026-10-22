-- Ground-item attachment to generic authoritative world/ECS state.
--
-- Item placement decides whether the item is in the world. This module owns
-- spatial transitions and checkpointed ground-target allocation. Room policy
-- remains in the population owner and no item-specific inactive archive exists.

local ecs = require("engine.ecs/v1")
local population = require("d2legacy.bootstrap.population")
local collision = require("d2legacy.gameplay.collision")
local state = require("engine.authority_state/v1")
local deterministic = require("engine.deterministic/v1")

local M = {}

local TARGET_STATE = "d2legacy.items.ground_targets"
local TARGET_SCHEMA = "d2legacy.items.ground_targets/v1"

function M.initialize()
    state.register(TARGET_STATE, TARGET_SCHEMA, { last = 0 })
end

local function framed(values)
    local result = {}
    for _, value in ipairs(values) do
        value = tostring(value)
        result[#result + 1] = #value .. ":" .. value
    end
    return table.concat(result)
end

local function selectable(owner, item, incarnation)
    local sequence = assert(state.read(TARGET_STATE), "ground target state is not initialized")
    -- Retired selectors must not alias a redrop or an ECS ID reused after restore.
    assert(sequence.last >= 0 and sequence.last < 9007199254740991
        and sequence.last == math.floor(sequence.last), "ground target sequence exhausted")
    sequence.last = sequence.last + 1
    state.replace(TARGET_STATE, TARGET_SCHEMA, sequence)
    local label = item.name_key and item.name_key ~= "" and item.name_key or item.code
    -- Replay may assign this sequence to another item; bind its namespace, provenance and incarnation too.
    local binding = deterministic.digest(framed({ owner, item.id, item.code, item.origin or "",
        item.generation and item.generation.facts or "", incarnation }))
    return { id = "item:" .. string.format("%.0f", sequence.last) .. ":" .. binding, kind = "item", label = label,
        owner = "", priority = 15, radius = 3.5 }
end

local function active_player(player)
    if not player or ecs.get(player, "d2legacy.world.inactive") or ecs.get(player, "d2legacy.player.death") then
        return false
    end
    local vitals = ecs.get(player, "d2legacy.player.vitals")
    return vitals ~= nil and vitals:get("health") > 0
end

function M.player(player_id)
    for _, entity in ipairs(ecs.query({ all = {
        "d2legacy.player.identity", "d2legacy.world.position", "d2legacy.world.location",
    } })) do
        if ecs.get(entity, "d2legacy.player.identity"):get("player") == player_id
            and active_player(entity) then return entity end
    end
end

function M.within_reach(player, location, x, y)
    if not active_player(player) or not location then return false end
    local from_location = ecs.get(player, "d2legacy.world.location")
    if not from_location then return false end
    if from_location:get("act") ~= location:get("act")
        or from_location:get("level_id") ~= location:get("level_id") then return false end
    local from = ecs.get(player, "d2legacy.world.position")
    if not from then return false end
    local dx, dy = from:get("x") - x, from:get("y") - y
    return dx * dx + dy * dy <= 3.5 * 3.5
        and collision.line_clear(from_location:get("level_id"), from:get("x"), from:get("y"), x, y)
end

function M.can_pick_up(player, item)
    if not item or not active_player(player) or ecs.get(item, "d2legacy.world.inactive") then return false end
    local position = ecs.get(item, "d2legacy.world.position")
    return position and M.within_reach(player, ecs.get(item, "d2legacy.world.location"),
        position:get("x"), position:get("y")) or false
end

function M.initial_components(owner, item)
    if item.container ~= "world" or type(item.level_id) ~= "number" or item.level_id <= 0 then
        return {}
    end
    assert(type(item.act) == "number" and item.act > 0, "world item act is required")
    assert(type(item.x) == "number" and type(item.y) == "number", "world item position is required")
    local selected = selectable(owner, item, "creation")
    return {
        ["d2legacy.world.selectable"] = selected,
        ["d2legacy.world.position"] = { x = item.x, y = item.y },
        ["d2legacy.world.location"] = { act = item.act, level_id = item.level_id },
        ["d2legacy.world.room_attach"] = { id = selected.id },
    }
end

function M.place(entity, location, x, y, command)
    assert(location, "world item owner location is required")
    local level_id = location:get("level_id")
    local identity = assert(ecs.get(entity, "d2legacy.item.identity"))
    for _, value in ipairs({ command.tick, command.sequence }) do
        assert(value >= 0 and value <= 9007199254740991 and value == math.floor(value),
            "ground command identity exceeds exact integer range")
    end
    local owner = assert(ecs.get(identity:get("owner"), "d2legacy.items.layout")):get("owner")
    local origin = ecs.get(entity, "d2legacy.item.origin")
    local generation = ecs.get(entity, "d2legacy.item.generation")
    local selected = selectable(owner, {
        id = identity:get("id"), code = identity:get("code"), origin = origin and origin:get("key") or "",
        generation = generation and { facts = generation:get("facts") } or nil,
    }, framed({ command.authority, command.player, string.format("%.0f", command.tick),
        string.format("%.0f", command.sequence) }))
    local id = selected.id
    ecs.set(entity, "d2legacy.world.selectable", selected)
    ecs.set(entity, "d2legacy.world.position", { x = x, y = y })
    ecs.set(entity, "d2legacy.world.location", location:snapshot())
    ecs.remove(entity, "d2legacy.world.inactive")
    local resident = population.resident_at(id, level_id, x, y)
    if resident then
        ecs.set(entity, "d2legacy.world.room_resident", resident)
        ecs.remove(entity, "d2legacy.world.room_attach")
    else
        ecs.remove(entity, "d2legacy.world.room_resident")
        ecs.set(entity, "d2legacy.world.room_attach", { id = id })
    end
end

function M.clear(entity)
    for _, component in ipairs({
        "d2legacy.world.selectable",
        "d2legacy.world.inactive",
        "d2legacy.world.room_resident",
        "d2legacy.world.room_attach",
        "d2legacy.world.location",
        "d2legacy.world.position",
    }) do
        ecs.remove(entity, component)
    end
end

return M
