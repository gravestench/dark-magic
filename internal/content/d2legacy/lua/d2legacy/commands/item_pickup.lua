-- Ground items become owned inventory only after range and space checks.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local owner = require("d2legacy.items.command_owner")
local find = require("d2legacy.items.find")
local placement = require("d2legacy.items.placement")
local world = require("d2legacy.items.world")
local M = {}

function M.validate(command)
    assert(type(command.payload) == "table", "pickup payload is required")
    owner.resolve(command)
    local id = command.payload.target_id
    assert(type(id) == "string" and id ~= "" and #id <= 128, "pickup target ID is required")
end

function M.apply(command)
    local player_id = owner.resolve(command)
    local player = world.player(player_id)
    local layout = find.layout(player_id)
    if not player or not layout then
        return
    end
    local entities = ecs.query({ all = { "d2legacy.item.identity", "d2legacy.item.placement" } })
    for _, entity in ipairs(entities) do
        local selected = ecs.get(entity, "d2legacy.world.selectable")
        local current = ecs.get(entity, "d2legacy.item.placement")
        if
            selected
            and selected:get("id") == command.payload.target_id
            and current:get("container") == "world"
            and world.can_pick_up(player, entity)
        then
            local item = ecs.get(entity, "d2legacy.item.identity")
            local destination = placement.first_free(layout, entity, item, entities)
            if not destination then
                return
            end
            item:set("owner", layout)
            placement.set(current, destination)
            world.clear(entity)
            return
        end
    end
end

function M.register()
    commands.register({ kind = "item.pickup", authorities = { "player" }, validate = M.validate, apply = M.apply })
end

return M
