-- Remove one authenticated session member and state owned only by that member.
--
-- Transport credentials and reconnect policy belong to the host. This command
-- defines the deterministic d2legacy consequence after membership ends, so a
-- replay never retains a moving ghost or dangling player-owned inventory.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local M = {}

local function player_entity(player_id)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
        if ecs.get(entity, "d2legacy.player.identity"):get("player") == player_id then
            return entity
        end
    end
    return nil
end

local function destroy_matching(component, predicate)
    for _, entity in ipairs(ecs.query({ all = { component } })) do
        local value = ecs.get(entity, component)
        if predicate(value, entity) then
            ecs.destroy(entity)
        end
    end
end

function M.validate(command)
    assert(
        type(command.payload) == "table"
            and type(command.payload.player) == "string"
            and command.payload.player:match("%S"),
        "player departure requires a player ID"
    )
end

function M.apply(command)
    local player_id = command.payload.player
    local player = player_entity(player_id)
    if not player then
        return
    end -- disconnect cleanup is intentionally idempotent

    destroy_matching("d2legacy.player.learned_skill", function(value)
        return value:get("owner"):id() == player:id()
    end)
    destroy_matching("d2legacy.stat.source", function(value)
        return value:get("target"):id() == player:id()
    end)
    destroy_matching("d2legacy.owned_unit", function(value)
        return value:get("owner"):id() == player:id()
    end)

    local layouts = {}
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.items.layout" } })) do
        if ecs.get(entity, "d2legacy.items.layout"):get("owner") == player_id then
            layouts[entity:id()] = true
            ecs.destroy(entity)
        end
    end
    local items = {}
    destroy_matching("d2legacy.item.identity", function(value, entity)
        if layouts[value:get("owner"):id()] then
            items[entity:id()] = true
            return true
        end
        return false
    end)
    destroy_matching("d2legacy.item.stat_modifier", function(value)
        return items[value:get("item"):id()] == true
    end)
    destroy_matching("d2legacy.interaction.context", function(value)
        if value:get("owner") ~= player_id then
            return false
        end
        local target = value:get("target")
        local ok, marker = pcall(ecs.get, target, "d2legacy.interaction.null_target")
        if ok and marker then
            ecs.destroy(target)
        end
        return true
    end)

    ecs.destroy(player)
end

function M.register()
    commands.register({
        kind = "system.player.leave",
        authorities = { "system", "administrator" },
        validate = M.validate,
        apply = M.apply,
    })
end

return M
