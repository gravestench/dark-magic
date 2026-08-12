-- Authoritative item movement, held-item swapping, and weapon-set selection.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local owner = require("d2legacy.items.command_owner")
local find = require("d2legacy.items.find")
local placement = require("d2legacy.items.placement")
local M = {}

local function item_entities()
    return ecs.query({
        any = {
            "d2legacy.items.layout",
            "d2legacy.item.identity",
        },
    })
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
    owner.resolve(command)
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

    local displaced = placement.validate(
        layout,
        item_entity,
        item,
        command.payload.destination,
        entities,
        command.payload.place_held
    )
    placement.set(current, command.payload.destination)
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
