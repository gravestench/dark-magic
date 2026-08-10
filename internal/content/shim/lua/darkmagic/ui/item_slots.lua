-- Reusable presentation adapter for authoritative named item slots.
-- Slot geometry comes from Inventory.txt; compatibility and occupancy remain
-- fixed-tick Go authority decisions.
local render = require("dm.render/v1")
local input = require("dm.input/v1")

local M = {}

local function held_item(snapshot)
    for _, item in ipairs(snapshot.items) do
        if item.container == "held" then return item end
    end
end

local function slotted_item(snapshot, container, slot)
    for _, item in ipairs(snapshot.items) do
        if item.container == container and item.slot == slot then return item end
    end
end

local function activate(slots, slot)
    local held = held_item(slots.snapshot)
    if held ~= nil then
        slots.items.move(held.id, { container = slots.container, slot = slot }, true)
        return
    end
    local item = slotted_item(slots.snapshot, slots.container, slot)
    if item ~= nil then slots.items.move(item.id, { container = "held" }) end
end

function M.create(root, controls, definition)
    local slots = {
        root = root,
        items = require("dm.items/v1"),
        container = assert(definition.container),
        palette = assert(definition.palette),
        geometry = assert(definition.slots),
        nodes = {},
    }
    for slot, geometry in pairs(slots.geometry) do
        local body_loc = slot
        if geometry.placeholder then
            local placeholder = geometry.placeholder
            local node = render.create("modal", root)
            local width, height = node:set_dc6(placeholder.sheet, slots.palette, 0, placeholder.frame or 0)
            node:set_position(
                geometry.x + geometry.width / 2 + (placeholder.offset_x or 0),
                geometry.y + geometry.height / 2 + (placeholder.offset_y or 0)
            )
            geometry.placeholder_node = node
        end
        controls:add({
            id = slots.container .. "_" .. body_loc,
            label = slots.container .. " " .. body_loc,
            x = geometry.x,
            y = geometry.y,
            width = geometry.width,
            height = geometry.height,
            on_activate = function() activate(slots, body_loc) end,
        })
    end
    M.update(slots)
    return slots
end

function M.update(slots)
    local snapshot = assert(slots.items.snapshot())
    local cursor_x, cursor_y = input.cursor()
    slots.snapshot = snapshot
    slots.held = held_item(snapshot) ~= nil
    for _, item in ipairs(snapshot.items) do
        local drawing = slots.nodes[item.id]
        if drawing == nil and item.inventory_dc6 ~= ""
            and (item.container == slots.container or item.container == "held") then
            local node = render.create("modal", slots.root)
            local width, height = node:set_dc6(item.inventory_dc6, slots.palette, 0, 0)
            drawing = { node = node, width = width, height = height }
            slots.nodes[item.id] = drawing
        end
        if drawing ~= nil then
            local geometry = item.container == slots.container and slots.geometry[item.slot] or nil
            drawing.node:set_visible(geometry ~= nil or item.container == "held")
            if geometry ~= nil then
                drawing.node:set_position(geometry.x + geometry.width / 2, geometry.y + geometry.height / 2)
            elseif item.container == "held" then
                drawing.node:set_position(cursor_x, cursor_y)
            end
        end
    end
end

return M
