-- Selectable/paged retained list widget.
--
-- A list looks like a special UI object, but this implementation is deliberately
-- built from ORDINARY controls. Each visible row is just another control managed
-- by controls.lua. That automatically gives rows the same mouse-up activation,
-- keyboard focus, accessibility state, and hover rules as every other widget.
--
-- The list object in this file only adds list-specific ideas:
--   * which items belong on the current page;
--   * which item is selected;
--   * how a row is drawn;
--   * the difference between selecting and activating.

local render = require("engine.render/v1")
local text = require("d2.ui.text")

local M = {}

function M.create(root, manager, id, definition, items, options)
    definition = definition or {}
    options = options or {}

    local x = assert(definition.x, "list x is required")
    local y = assert(definition.y, "list y is required")
    local width = assert(definition.width, "list width is required")
    local row_height = assert(definition.row_height, "list row height is required")

    -- Let callers say either `page_size` or `rows`; otherwise show five rows.
    local page_size = definition.page_size or definition.rows or 5
    assert(page_size > 0, "list page size must be positive")

    -- `result` is the public Lua object returned to the caller.
    local result = {
        id = id,
        items = items or {},
        rows = {},
        page = 1,
        page_size = page_size,
        selected_index = definition.selected_index,
    }

    local function item_label(item, index)
        -- A screen may provide a custom labeling function for its own item type.
        if options.item_label then return options.item_label(item, index) end

        -- For table items, try common human-readable fields in friendly order.
        if type(item) == "table" then return tostring(item.label or item.name or item.id or index) end

        -- Strings/numbers/etc. can simply become text directly.
        return tostring(item)
    end

    local function row_style(row)
        if not row.item then return options.empty_style or "font_lab_caption" end

        -- Selection beats transient hover/focus styling.
        if result.selected_index == row.global_index then
            return options.selected_style or "character_create_option"
        end

        if row.control.state == "hover" or row.control.state == "focused" or row.control.state == "pressed" then
            return options.hover_style or "label_button_hover"
        end

        return options.normal_style or "font_lab_caption"
    end

    local function draw_row(row)
        -- Headless list rows still exist as controls but have no render nodes.
        if not row.label_node then return end

        local visible = row.item ~= nil
        row.label_node:set_visible(visible)
        if row.background then row.background:set_visible(visible) end
        if not visible then return end

        if row.background then
            local selected = result.selected_index == row.global_index
            local hovered = row.control.state == "hover" or row.control.state == "focused" or row.control.state == "pressed"

            if selected or hovered then
                -- Selection background wins. Otherwise use hover background.
                local color = selected and (options.selected_background or {52, 42, 27, 210})
                    or (options.hover_background or {34, 29, 23, 180})
                row.background:fill_rect(width, row_height, color[1], color[2], color[3], color[4] or 255)
                row.background:set_visible(true)
            else
                row.background:set_visible(false)
            end
        end

        local label = item_label(row.item, row.global_index)
        local _, height = text.set(row.label_node, row_style(row), label, width - 12, options.align or "left")

        -- This expression looks fancy but the -height/2 + height/2 cancel out.
        -- It preserves the row-center positioning shape used when this widget was
        -- recovered/refined, while making the intended row center explicit.
        row.label_node:set_position(x + width / 2, row.y + row_height / 2 - height / 2 + height / 2)
    end

    local function draw_all()
        for _, row in ipairs(result.rows) do draw_row(row) end
    end

    local function page_count()
        -- Even an empty list conceptually has page 1; that keeps paging callers
        -- from needing a special "zero pages" branch.
        return math.max(1, math.ceil(#result.items / result.page_size))
    end

    function result:refresh()
        -- Clamp page in case the item collection shrank.
        self.page = math.max(1, math.min(page_count(), self.page))

        -- Lua arrays start at 1. Page 1 starts at 1, page 2 at page_size+1, etc.
        local first = (self.page - 1) * self.page_size + 1

        for row_index, row in ipairs(self.rows) do
            local index = first + row_index - 1
            row.global_index = index
            row.item = self.items[index]

            -- Empty physical row slots are kept around but disabled/hidden. This
            -- avoids constantly destroying/recreating controls while paging.
            row.control.visible = row.item ~= nil
            row.control.enabled = row.item ~= nil
            row.control.label = row.item and item_label(row.item, index) or ""
            draw_row(row)
        end

        if options.on_page then options.on_page(self.page, page_count()) end
    end

    function result:set_items(next_items)
        self.items = next_items or {}

        -- If the selected item disappeared, selection becomes empty rather than
        -- pointing beyond the new list.
        if self.selected_index and self.selected_index > #self.items then self.selected_index = nil end
        self:refresh()
    end

    function result:select(index, activate)
        if index == nil or self.items[index] == nil then return false end

        -- A repeated activation of the already-selected row is treated as a
        -- stronger "open/confirm this item" gesture. Character lists use this.
        local repeated = self.selected_index == index
        self.selected_index = index
        draw_all()

        if not repeated and options.on_select then options.on_select(self.items[index], index) end
        if (activate == true or repeated) and options.on_activate then options.on_activate(self.items[index], index) end
        return true
    end

    function result:set_page(page)
        local next_page = math.max(1, math.min(page_count(), page))
        if next_page == self.page then return false end
        self.page = next_page
        self:refresh()
        return true
    end

    -- Tiny convenience methods make call sites read like English.
    function result:next_page() return self:set_page(self.page + 1) end
    function result:previous_page() return self:set_page(self.page - 1) end
    function result:page_count() return page_count() end

    -- Create a STABLE pool of physical row controls once. Paging changes the
    -- data assigned to each row; it does not rebuild the input structure.
    for row_index = 1, page_size do
        local row_y = y + (row_index - 1) * row_height
        local row = { y = row_y }

        if render.assets_available() then
            row.background = render.create(options.layer or "hud", root)
            row.background:set_position(x + width / 2, row_y + row_height / 2)
            row.background:set_visible(false)
            row.label_node = render.create(options.layer or "hud", root)
        end

        row.control = manager:add({
            id = id .. "_row_" .. tostring(row_index),
            role = "listitem",
            label = "",
            x = x,
            y = row_y,
            width = width,
            height = row_height,
            scope = options.scope or definition.scope,

            on_activate = function()
                -- `row` is this iteration's local table, so the closure remembers
                -- whichever item/page index refresh assigned to this physical row.
                if row.item then result:select(row.global_index, false) end
            end,

            on_state = function()
                draw_row(row)
            end,
        })

        result.rows[#result.rows + 1] = row
    end

    result:refresh()
    return result
end

return M
