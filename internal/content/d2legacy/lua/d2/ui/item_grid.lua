-- Reusable presentation adapter for authoritative SPATIAL item containers.
--
-- Read this file closely if you want to make inventory-like UI. It demonstrates
-- one of Dark Magic's most important rules:
--
--     Lua presentation reads a SNAPSHOT and submits INTENT.
--     It does not directly mutate authoritative gameplay state.
--
-- This module knows where cells are drawn and clicked. It never decides whether
-- an item fits, overlaps legally, can be swapped, etc. `engine.items/v1` queues a
-- move request; fixed-tick Go authority validates and applies that request.

local render = require("engine.render/v1")
local input = require("engine.input/v1")

local M = {}

-- Find the item whose rectangular grid footprint covers one logical cell.
local function item_at(snapshot, container, column, row)
    for _, item in ipairs(snapshot.items) do
        -- An item can be larger than one cell. These comparisons are half-open:
        -- x <= column < x+width, y <= row < y+height.
        if item.container == container
            and column >= item.x and column < item.x + item.width
            and row >= item.y and row < item.y + item.height then
            return item
        end
    end
end

-- The "held" container is authoritative cursor-item state. It is not just a
-- temporary picture attached to the local mouse.
local function held_item(snapshot)
    for _, item in ipairs(snapshot.items) do
        if item.container == "held" then return item end
    end
end

-- Convert one cell activation into a gameplay INTENT.
local function activate(grid, column, row)
    local held = held_item(grid.snapshot)

    if held ~= nil then
        -- Player is carrying something: ask authority to place it at this cell.
        -- The final `true` means an occupied destination may perform the allowed
        -- atomic swap semantics. Lua does not implement that swap itself.
        grid.items.move(held.id, {
            container = grid.container,
            x = column,
            y = row,
        }, true)
        return
    end

    -- Empty hand: clicking an occupied cell asks to move that item into `held`.
    local item = item_at(grid.snapshot, grid.container, column, row)
    if item ~= nil then grid.items.move(item.id, { container = "held" }) end
end

function M.create(root, controls, definition)
    -- Build one small Lua object containing everything the presenter needs.
    local grid = {
        root = root,
        controls = controls,
        -- Capability is required lazily here so presentation-only test harnesses
        -- that never create an item grid do not need gameplay item authority.
        items = require("engine.items/v1"),
        container = assert(definition.container),
        columns = assert(definition.columns),
        rows = assert(definition.rows),
        left = assert(definition.left),
        top = assert(definition.top),
        cell_width = assert(definition.cell_width),
        cell_height = assert(definition.cell_height),
        palette = assert(definition.palette),
        -- Cache render handles by stable item ID so we do not recreate art every frame.
        nodes = {},
    }

    -- Register every logical grid cell as an input control. Grid coordinates are
    -- zero-based because the authoritative item model uses zero-based cells.
    for row = 0, grid.rows - 1 do
        for column = 0, grid.columns - 1 do
            -- IMPORTANT Lua closure detail: save per-iteration values in locals.
            -- The callback below should remember THIS cell, not whatever the loop
            -- variables contain on a later iteration.
            local cell_column, cell_row = column, row

            controls:add({
                -- Stable IDs are useful for tests/accessibility/debugging.
                id = string.format("%s_%d_%d", grid.container, column, row),
                -- Human-facing labels use +1 because people usually count from 1.
                label = string.format("%s %d, %d", grid.container, column + 1, row + 1),
                -- Convert logical cell coordinates into screen pixels.
                x = grid.left + column * grid.cell_width,
                y = grid.top + row * grid.cell_height,
                width = grid.cell_width,
                height = grid.cell_height,
                on_activate = function() activate(grid, cell_column, cell_row) end,
            })
        end
    end

    -- Populate snapshot/render state immediately so the panel is correct before
    -- waiting for its first scene update.
    M.update(grid)
    return grid
end

function M.update(grid)
    -- Read a fresh VALUE snapshot. This is copied descriptive state, not a table
    -- Lua is allowed to mutate as gameplay authority.
    local snapshot = assert(grid.items.snapshot())
    local cursor_x, cursor_y = input.cursor()

    -- Store the exact snapshot that click handlers should reason about this frame.
    grid.snapshot = snapshot
    grid.held = held_item(snapshot) ~= nil

    for _, item in ipairs(snapshot.items) do
        local drawing = grid.nodes[item.id]

        -- Lazily create art only when an item becomes relevant to this presenter.
        -- An item somewhere else (stash, equipment, world, etc.) needs no node here.
        if drawing == nil and item.inventory_dc6 ~= ""
            and (item.container == grid.container or item.container == "held") then
            local node = render.create("modal", grid.root)
            local width, height = node:set_dc6(item.inventory_dc6, grid.palette, 0, 0)
            drawing = { node = node, width = width, height = height }
            grid.nodes[item.id] = drawing
        end

        if drawing ~= nil then
            -- Keep cached nodes but show them only while this presenter owns the
            -- item's current visual role.
            local visible = item.container == grid.container or item.container == "held"
            drawing.node:set_visible(visible)

            if item.container == grid.container then
                -- Item x/y are logical grid cells. Its DC6 image is positioned
                -- from the top-left cell plus half decoded image dimensions.
                drawing.node:set_position(
                    grid.left + item.x * grid.cell_width + drawing.width / 2,
                    grid.top + item.y * grid.cell_height + drawing.height / 2
                )
            elseif item.container == "held" then
                -- Authority says the item is held; presentation simply follows
                -- the local pointer with its image.
                drawing.node:set_position(cursor_x, cursor_y)
            end
        end
    end
end

return M
