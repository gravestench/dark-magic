-- Reusable presentation adapter for authoritative NAMED item slots.
--
-- `item_grid.lua` handles things placed at (column,row). This file handles
-- things placed into named wells such as `head`, `torso`, or `rarm`.
--
-- The same authority rule applies: Lua draws a copied snapshot and submits move
-- intent. It never decides whether a sword belongs in a helmet slot; the engine
-- validates body-location compatibility and occupancy on the fixed tick.

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
        -- Ask authority to place the held item into this named destination. The
        -- `true` preserves atomic occupied-slot swap behavior when it is legal.
        slots.items.move(held.id, { container = slots.container, slot = slot }, true)
        return
    end

    -- Empty hand: lift the current occupant into the authoritative held container.
    local item = slotted_item(slots.snapshot, slots.container, slot)
    if item ~= nil then slots.items.move(item.id, { container = "held" }) end
end

function M.create(root, controls, definition)
    local slots = {
        root = root,
        items = require("dm.items/v1"),
        container = assert(definition.container),
        palette = assert(definition.palette),
        -- Geometry is keyed by semantic/body-location name rather than array index.
        geometry = assert(definition.slots),
        nodes = {},
    }

    -- `pairs` is appropriate because the semantic slot name matters, not table order.
    for slot, geometry in pairs(slots.geometry) do
        -- Capture this loop's key in a dedicated local for the callback below.
        local body_loc = slot

        if geometry.placeholder then
            -- Empty equipment wells can have authored DC6 placeholder art. This
            -- image is presentation only; it is not an item or authoritative occupant.
            local placeholder = geometry.placeholder
            local node = render.create("modal", root)
            local width, height = node:set_dc6(placeholder.sheet, slots.palette, 0, placeholder.frame or 0)

            -- Geometry x/y describes the well's top-left. Center the placeholder
            -- inside the well, then apply any small authored correction offsets.
            node:set_position(
                geometry.x + geometry.width / 2 + (placeholder.offset_x or 0),
                geometry.y + geometry.height / 2 + (placeholder.offset_y or 0)
            )
            geometry.placeholder_node = node
        end

        -- The whole equipment well is clickable, not just whatever item art is inside.
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

        -- Lazily allocate presentation only for an item currently in this named
        -- container or in the shared held container.
        if drawing == nil and item.inventory_dc6 ~= ""
            and (item.container == slots.container or item.container == "held") then
            local node = render.create("modal", slots.root)
            local width, height = node:set_dc6(item.inventory_dc6, slots.palette, 0, 0)
            drawing = { node = node, width = width, height = height }
            slots.nodes[item.id] = drawing
        end

        if drawing ~= nil then
            -- Compact conditional: if the item is in this slot container, look
            -- up that slot's geometry; otherwise geometry is nil.
            local geometry = item.container == slots.container and slots.geometry[item.slot] or nil

            drawing.node:set_visible(geometry ~= nil or item.container == "held")

            if geometry ~= nil then
                -- Equipment art is centered in the authored well; it is not
                -- stretched to fill the well.
                drawing.node:set_position(geometry.x + geometry.width / 2, geometry.y + geometry.height / 2)
            elseif item.container == "held" then
                drawing.node:set_position(cursor_x, cursor_y)
            end
        end
    end
end

return M
