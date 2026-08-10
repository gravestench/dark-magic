-- Selectable/paged retained list widget.
--
-- Rows register as ordinary darkmagic.ui.controls entries so mouse-up
-- activation, focus, accessibility, and keyboard navigation remain shared with
-- every other UI control. A second activation of the selected row is exposed
-- separately for character-list style UX.
local render = require("dm.render/v1")
local text = require("darkmagic.ui.text")

local M = {}

function M.create(root, manager, id, definition, items, options)
    definition = definition or {}
    options = options or {}
    local x = assert(definition.x, "list x is required")
    local y = assert(definition.y, "list y is required")
    local width = assert(definition.width, "list width is required")
    local row_height = assert(definition.row_height, "list row height is required")
    local page_size = definition.page_size or definition.rows or 5
    assert(page_size > 0, "list page size must be positive")

    local result = {
        id = id,
        items = items or {},
        rows = {},
        page = 1,
        page_size = page_size,
        selected_index = definition.selected_index,
    }

    local function item_label(item, index)
        if options.item_label then return options.item_label(item, index) end
        if type(item) == "table" then return tostring(item.label or item.name or item.id or index) end
        return tostring(item)
    end

    local function row_style(row)
        if not row.item then return options.empty_style or "font_lab_caption" end
        if result.selected_index == row.global_index then return options.selected_style or "character_create_option" end
        if row.control.state == "hover" or row.control.state == "focused" or row.control.state == "pressed" then
            return options.hover_style or "label_button_hover"
        end
        return options.normal_style or "font_lab_caption"
    end

    local function draw_row(row)
        if not row.label_node then return end
        local visible = row.item ~= nil
        row.label_node:set_visible(visible)
        if row.background then row.background:set_visible(visible) end
        if not visible then return end

        if row.background then
            local selected = result.selected_index == row.global_index
            local hovered = row.control.state == "hover" or row.control.state == "focused" or row.control.state == "pressed"
            if selected or hovered then
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
        row.label_node:set_position(x + width / 2, row.y + row_height / 2 - height / 2 + height / 2)
    end

    local function draw_all()
        for _, row in ipairs(result.rows) do draw_row(row) end
    end

    local function page_count()
        return math.max(1, math.ceil(#result.items / result.page_size))
    end

    function result:refresh()
        self.page = math.max(1, math.min(page_count(), self.page))
        local first = (self.page - 1) * self.page_size + 1
        for row_index, row in ipairs(self.rows) do
            local index = first + row_index - 1
            row.global_index = index
            row.item = self.items[index]
            row.control.visible = row.item ~= nil
            row.control.enabled = row.item ~= nil
            row.control.label = row.item and item_label(row.item, index) or ""
            draw_row(row)
        end
        if options.on_page then options.on_page(self.page, page_count()) end
    end

    function result:set_items(next_items)
        self.items = next_items or {}
        if self.selected_index and self.selected_index > #self.items then self.selected_index = nil end
        self:refresh()
    end

    function result:select(index, activate)
        if index == nil or self.items[index] == nil then return false end
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

    function result:next_page() return self:set_page(self.page + 1) end
    function result:previous_page() return self:set_page(self.page - 1) end
    function result:page_count() return page_count() end

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
