-- Owned inventory scroll use. Pair creation and consumption share one admitted tick.
local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local initial = require("engine.initial_data/v1")
local owner = require("d2legacy.items.command_owner")
local find = require("d2legacy.items.find")
local consume = require("d2legacy.items.consume")
local collision = require("d2legacy.gameplay.collision")
local population = require("d2legacy.bootstrap.population")
local warp = require("d2legacy.systems.world_warp")
local seams = require("d2legacy.systems.world_transition")
local M = {}

local function player(player_id)
    for _, entity in
        ipairs(ecs.query({
            all = {
                "d2legacy.world.player_control",
                "d2legacy.world.location",
                "d2legacy.world.position",
                "d2legacy.world.bounds",
                "d2legacy.player.vitals",
            },
            none = { "d2legacy.world.inactive", "d2legacy.player.death" },
        }))
    do
        if ecs.get(entity, "d2legacy.world.player_control"):get("player") == player_id then
            return entity
        end
    end
end

local function nearby(level, x, y, width, height)
    -- Bounded placement is not equivalent to native portal/arrival searches.
    for ring = 0, 4 do
        for dy = -ring, ring do
            for dx = -ring, ring do
                if math.max(math.abs(dx), math.abs(dy)) == ring then
                    local px, py = x + dx, y + dy
                    if
                        px >= 1
                        and py >= 1
                        and px < width - 1
                        and py < height - 1
                        and not seams.at_trigger(level, px, py)
                        and collision.destination_clear(level, px, py, 1)
                    then
                        return px, py
                    end
                end
            end
        end
    end
end

local function endpoint(id, pair, player_id, level, x, y, destination, returning)
    local value = {
        id = id,
        pair_id = pair,
        owner = player_id,
        return_endpoint = returning,
        token = "TP",
        label = "Town Portal",
        level_id = level,
        x = x,
        y = y,
        radius = 3.5,
        destination_level = destination.level_id,
        destination_x = destination.x,
        destination_y = destination.y,
        destination_width = destination.width,
        destination_height = destination.height,
    }
    local resident = population.resident_at(id, level, x, y)
    if resident then
        value.resident_id, value.room_id = id, resident.room_id
    end
    return value
end

function M.apply(command)
    local player_id = owner.resolve(command)
    local layout = find.layout(player_id)
    if not layout then
        return
    end
    local item_entity, item = find.item(layout, command.payload.item_id)
    if not item_entity or item:get("code") ~= "tsc" or ecs.get(item_entity, "d2legacy.world.inactive") then
        return
    end
    local placement = ecs.get(item_entity, "d2legacy.item.placement")
    if not placement or placement:get("container") ~= "inventory" then
        return
    end
    for _, candidate in ipairs(ecs.query({ all = { "d2legacy.item.identity", "d2legacy.item.placement" } })) do
        local identity = ecs.get(candidate, "d2legacy.item.identity")
        if
            identity:get("owner"):id() == layout:id()
            and ecs.get(candidate, "d2legacy.item.placement"):get("container") == "held"
        then
            return
        end
    end
    local entity = player(player_id)
    if not entity or ecs.get(entity, "d2legacy.player.vitals"):get("health") <= 0 then
        return
    end
    local location = ecs.get(entity, "d2legacy.world.location")
    -- Only the generated Act I entry pair is supported by this slice.
    if location:get("act") ~= 1 or location:get("level_id") ~= 2 then
        return
    end
    local spawns = initial.get("d2legacy.world_spawns") or {}
    local town = spawns["1"]
    if not town then
        return
    end
    local tx, ty = nearby(1, town.x, town.y, town.width, town.height)
    if not tx then
        return
    end
    local position, bounds = ecs.get(entity, "d2legacy.world.position"), ecs.get(entity, "d2legacy.world.bounds")
    local fx, fy = nearby(2, position:get("x"), position:get("y"), bounds:get("width"), bounds:get("height"))
    if not fx then
        return
    end
    -- Canonical entity identity is session-unique and checkpointed. Public IDs
    -- must not grow with otherwise valid owner/item labels or truncate them.
    local id = "town-portal:item:" .. item_entity:id()
    local field_id, town_id = id .. ":field", id .. ":town"
    local field = { level_id = 2, x = fx, y = fy, width = bounds:get("width"), height = bounds:get("height") }
    local destination = { level_id = 1, x = tx, y = ty, width = town.width, height = town.height }
    -- Validate replacement positions before expiring the old pair. This preserves
    -- refused-cast resources, unlike recovered 1.10f failed-recast ordering.
    for _, existing in ipairs(ecs.query({ all = { "d2legacy.world.town_portal" } })) do
        local portal = ecs.get(existing, "d2legacy.world.town_portal")
        if portal:get("owner") == player_id then
            portal:set("expired", true)
        end
    end
    warp.create(endpoint(field_id, town_id, player_id, 2, fx, fy, destination, false))
    warp.create(endpoint(town_id, field_id, player_id, 1, tx, ty, field, true))
    consume.destroy(item_entity)
end

function M.register()
    commands.register({
        kind = "item.use",
        authorities = { "player" },
        validate = function(command)
            assert(type(command.payload.item_id) == "string" and command.payload.item_id ~= "", "item ID is required")
            owner.resolve(command)
        end,
        apply = M.apply,
    })
    -- Room activation restores inactive residents first. Retiring afterward
    -- preserves registered references until restoration clears them.
    ecs.system({
        id = "d2legacy.world.portal_expiry",
        phase = "collision",
        query = { all = { "d2legacy.world.town_portal" }, none = { "d2legacy.world.inactive" } },
        read = { "d2legacy.world.town_portal" },
        write = {},
        update = function(_, entities, structural)
            for _, entity in ipairs(entities) do
                if ecs.get(entity, "d2legacy.world.town_portal"):get("expired") then
                    structural:destroy(entity)
                end
            end
        end,
    })
end

return M
