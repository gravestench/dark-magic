-- Authoritative item movement, held-item swapping, and weapon-set selection.
--
-- Validation is intentionally separate from mutation. The generic host admits
-- the command, this file checks Diablo container rules, and placement.lua
-- performs the small deterministic move or one-item swap.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local owner = require("d2legacy.items.command_owner")
local find = require("d2legacy.items.find")
local placement = require("d2legacy.items.placement")
local world = require("d2legacy.items.world")
local M = {}

local function item_entities()
    return ecs.query({
        any = {
            "d2legacy.items.layout",
            "d2legacy.item.identity",
        },
    })
end

local function player_location(player_id)
    for _, entity in ipairs(ecs.query({
        all = { "d2legacy.player.identity", "d2legacy.world.location" },
    })) do
        local identity = ecs.get(entity, "d2legacy.player.identity")
        if identity:get("player") == player_id then
            return ecs.get(entity, "d2legacy.world.location")
        end
    end
    return nil
end

function M.validate_move(command)
    local payload = command.payload
    assert(type(payload) == "table", "item move payload is required")
    assert(type(payload.item_id) == "string" and payload.item_id ~= "",
        "item ID is required")
    assert(type(payload.destination) == "table",
        "item destination is required")
    assert(type(payload.destination.container) == "string",
        "item destination container is required")
    assert(payload.destination.container ~= "vendor",
        "vendor stock requires a transaction")
    if payload.place_held then
        assert(placement.is_held_destination(payload.destination.container),
            "invalid held-item destination")
    end
    if payload.destination.container == "world" then
        assert(type(payload.destination.x) == "number", "world item x is required")
        assert(type(payload.destination.y) == "number", "world item y is required")
        assert(player_location(owner.resolve(command)), "world item owner location is required")
    end
    local entities = item_entities()
    local layout_entity = assert(find.layout(owner.resolve(command), entities))
    local item_entity = assert(find.item(layout_entity, payload.item_id, entities))
    assert(not ecs.get(item_entity, "d2legacy.world.inactive"), "item is unavailable")
end

function M.apply_move(command)
    local entities = item_entities()
    local layout_entity, layout = assert(find.layout(
        owner.resolve(command),
        entities
    ))
    local item_entity, item = assert(find.item(
        layout_entity,
        command.payload.item_id,
        entities
    ))
    local current = assert(ecs.get(item_entity, "d2legacy.item.placement"))
    if command.payload.place_held then
        assert(current:get("container") == "held", "item is not held")
    end

    -- A held item may replace exactly one occupied item. That displaced item
    -- becomes the new held item, matching legacy inventory reordering.
    local displaced = placement.validate(
        layout,
        item_entity,
        item,
        command.payload.destination,
        entities,
        command.payload.place_held
    )
    local destination = command.payload.destination
    placement.set(current, destination)
    if destination.container == "world" then
        world.place(
            item_entity,
            layout:get("owner"),
            item:get("id"),
            assert(player_location(owner.resolve(command))),
            destination.x,
            destination.y
        )
    else
        world.clear(item_entity)
    end
    if displaced then
        placement.set(ecs.get(displaced, "d2legacy.item.placement"), {
            container = "held",
        })
    end
end

function M.validate_weapon_set(command)
    local set = command.payload and command.payload.set
    assert(set == 0 or set == 1, "weapon set must be 0 or 1")
    owner.resolve(command)
end

function M.apply_weapon_set(command)
    local _, layout = assert(find.layout(owner.resolve(command)))
    layout:set("active_weapon_set", command.payload.set)
end

function M.register()
    commands.register({
        kind = "item.move",
        authorities = { "player", "administrator" },
        validate = M.validate_move,
        apply = M.apply_move,
    })
    commands.register({
        kind = "item.weapon_set",
        authorities = { "player", "administrator" },
        validate = M.validate_weapon_set,
        apply = M.apply_weapon_set,
    })
end

return M
