-- Authoritative quest and vendor item-service transactions.
--
-- A presentation submits only a service identity. The rule component chooses
-- the escrow target slot, material slots, and price. Every requirement is
-- checked before the target changes or any material entity is destroyed.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local owner = require("d2legacy.items.command_owner")
local find = require("d2legacy.items.find")
local M = {}

local function split(value)
    local result = {}
    for token in string.gmatch(value or "", "[^,]+") do
        result[#result + 1] = token
    end
    return result
end

local function rule_by_id(service_id)
    local wanted = string.lower(service_id)
    for _, entity in ipairs(ecs.query({
        all = { "d2legacy.item.service_rule" },
    })) do
        local rule = ecs.get(entity, "d2legacy.item.service_rule")
        if rule:get("id") == wanted then return rule end
    end
    error("unknown item service")
end

local function item_in_service_slot(layout_entity, slot)
    for _, entity in ipairs(ecs.query({
        all = {
            "d2legacy.item.identity",
            "d2legacy.item.placement",
        },
    })) do
        local item = ecs.get(entity, "d2legacy.item.identity")
        local placed = ecs.get(entity, "d2legacy.item.placement")
        local matches = item:get("owner"):id() == layout_entity:id()
            and placed:get("container") == "quest_service"
            and placed:get("slot") == slot
        if matches then return entity, item end
    end
    return nil, nil
end

local function append_service(item, service_id)
    local applied = split(item:get("applied_services"))
    applied[#applied + 1] = service_id
    item:set("applied_services", table.concat(applied, ","))
end

local function destroy_item(item_entity)
    -- Item modifiers are immutable children of the item identity. Remove them
    -- first so a consumed service material cannot leave a dangling entity
    -- reference that later equipment/checkpoint code might inspect.
    for _, entity in ipairs(ecs.query({
        all = { "d2legacy.item.stat_modifier" },
    })) do
        local modifier = ecs.get(entity, "d2legacy.item.stat_modifier")
        if modifier:get("item"):id() == item_entity:id() then
            ecs.destroy(entity)
        end
    end
    ecs.destroy(item_entity)
end

function M.validate(command)
    local service = command.payload and command.payload.service
    assert(type(service) == "string" and service ~= "",
        "service identity is required")
    owner.resolve(command)
    rule_by_id(service)
end

function M.apply(command)
    local layout_entity, layout = assert(find.layout(owner.resolve(command)))
    local rule = rule_by_id(command.payload.service)
    local target_entity, target = item_in_service_slot(
        layout_entity,
        rule:get("target_slot")
    )
    assert(target_entity, "service target slot is empty")

    local materials = {}
    for _, slot in ipairs(split(rule:get("consume_slots"))) do
        local material = item_in_service_slot(layout_entity, slot)
        assert(material, "service material slot is empty")
        materials[#materials + 1] = material
    end
    assert(layout:get("carried_gold") >= rule:get("gold_cost"),
        "insufficient carried gold")

    append_service(target, rule:get("id"))
    for _, material in ipairs(materials) do
        destroy_item(material)
    end
    layout:set(
        "carried_gold",
        layout:get("carried_gold") - rule:get("gold_cost")
    )
end

function M.register()
    commands.register({
        kind = "item.service_complete",
        authorities = { "player", "administrator" },
        validate = M.validate,
        apply = M.apply,
    })
end

return M
