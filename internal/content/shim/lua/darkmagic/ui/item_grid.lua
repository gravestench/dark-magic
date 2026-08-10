-- Reusable presentation adapter for authoritative spatial item containers.
--
-- This module knows where cells are drawn and clicked. It never decides
-- whether an item fits: dm.items queues intent and Go authority owns that rule.
local render = require("dm.render/v1")
local input = require("dm.input/v1")

local M = {}

local function item_at(snapshot, container, column, row)
    for _, item in ipairs(snapshot.items) do
        if item.container == container
            and column >= item.x and column < item.x + item.width
            and row >= item.y and row < item.y + item.height then
            return item
        end
    end
end

local function held_item(snapshot)
    for _, item in ipairs(snapshot.items) do
        if item.container == "held" then return item end
    end
end

local function activate(grid, column, row)
    local held = held_item(grid.snapshot)
    if held ~= nil then
        grid.items.move(held.id, {
            container = grid.container,
            x = column,
            y = row,
        }, true)
        return
    end
    local item = item_at(grid.snapshot, grid.container, column, row)
    if item ~= nil then grid.items.move(item.id, { container = "held" }) end
end

function M.create(root, controls, definition)
    local grid = {
        root = root,
        controls = controls,
        items = require("dm.items/v1"),
        container = assert(definition.container),
        columns = assert(definition.columns),
        rows = assert(definition.rows),
        left = assert(definition.left),
        top = assert(definition.top),
        cell_width = assert(definition.cell_width),
        cell_height = assert(definition.cell_height),
        palette = assert(definition.palette),
        nodes = {},
    }
    for row = 0, grid.rows - 1 do
        for column = 0, grid.columns - 1 do
            local cell_column, cell_row = column, row
            controls:add({
                id = string.format("%s_%d_%d", grid.container, column, row),
                label = string.format("%s %d, %d", grid.container, column + 1, row + 1),
                x = grid.left + column * grid.cell_width,
                y = grid.top + row * grid.cell_height,
                width = grid.cell_width,
                height = grid.cell_height,
                on_activate = function() activate(grid, cell_column, cell_row) end,
            })
        end
    end
    M.update(grid)
    return grid
end

function M.update(grid)
    local snapshot = assert(grid.items.snapshot())
    local cursor_x, cursor_y = input.cursor()
    grid.snapshot = snapshot
    for _, item in ipairs(snapshot.items) do
        local drawing = grid.nodes[item.id]
        if drawing == nil and item.inventory_dc6 ~= ""
            and (item.container == grid.container or item.container == "held") then
            local node = render.create("modal", grid.root)
            local width, height = node:set_dc6(item.inventory_dc6, grid.palette, 0, 0)
            drawing = { node = node, width = width, height = height }
            grid.nodes[item.id] = drawing
        end
        if drawing ~= nil then
            local visible = item.container == grid.container or item.container == "held"
            drawing.node:set_visible(visible)
            if item.container == grid.container then
                drawing.node:set_position(
                    grid.left + item.x * grid.cell_width + drawing.width / 2,
                    grid.top + item.y * grid.cell_height + drawing.height / 2
                )
            elseif item.container == "held" then
                drawing.node:set_position(cursor_x + drawing.width / 2, cursor_y + drawing.height / 2)
            end
        end
    end
end

return M
