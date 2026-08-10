-- Reusable mutually-exclusive tab/selection group.
--
-- Tabs are another example of building a "bigger" widget from ordinary controls.
-- Each tab is just one control. This wrapper adds the rule that exactly one tab
-- ID is considered selected and asks all tab visuals to refresh when it changes.

local render = require("dm.render/v1")
local text = require("darkmagic.ui.text")

local M = {}

function M.create(root, manager, id, definition, tabs, options)
    definition = definition or {}
    options = options or {}
    tabs = tabs or {}

    local x = assert(definition.x, "tabs x is required")
    local y = assert(definition.y, "tabs y is required")
    local tab_width = assert(definition.tab_width, "tab width is required")
    local height = assert(definition.height, "tab height is required")
    local gap = definition.gap or 0
    local orientation = definition.orientation or "horizontal"

    local result = {
        id = id,
        -- If no explicit initial ID exists, use the first tab's ID when present.
        selected = definition.selected or (tabs[1] and tabs[1].id),
        tabs = {},
    }

    local function draw(entry)
        if not entry.label_node then return end

        local selected = result.selected == entry.id
        local state = entry.control.state
        local style

        -- Persistent SELECTED state wins over transient pointer/focus state.
        if selected then
            style = options.selected_style or "character_create_option"
        elseif state == "pressed" then
            style = options.pressed_style or options.selected_style or "character_create_option"
        elseif state == "hover" or state == "focused" then
            style = options.hover_style or "label_button_hover"
        else
            style = options.normal_style or "label_button_normal"
        end

        text.set(entry.label_node, style, entry.label, tab_width, "center")

        if entry.background then
            if selected then
                local color = options.selected_background or {52, 42, 27, 210}
                entry.background:fill_rect(tab_width, height, color[1], color[2], color[3], color[4] or 255)
                entry.background:set_visible(true)
            else
                entry.background:set_visible(false)
            end
        end
    end

    local function draw_all()
        for _, entry in ipairs(result.tabs) do draw(entry) end
    end

    function result:select(tab_id)
        -- Selecting what is already selected is not a change.
        if tab_id == self.selected then return false end

        -- Validate that the requested ID actually belongs to this group.
        local found
        for _, entry in ipairs(self.tabs) do
            if entry.id == tab_id then found = entry break end
        end
        if not found then return false end

        self.selected = tab_id
        draw_all()

        -- Return both the semantic ID and original tab definition to the caller.
        if options.on_change then options.on_change(tab_id, found.definition) end
        return true
    end

    for index, tab in ipairs(tabs) do
        -- Only one axis advances. The compact `condition and value or 0` form
        -- chooses the correct spacing term for horizontal/vertical layouts.
        local tx = x + (orientation == "horizontal" and (index - 1) * (tab_width + gap) or 0)
        local ty = y + (orientation == "vertical" and (index - 1) * (height + gap) or 0)

        local entry = {
            id = assert(tab.id, "tab id is required"),
            label = assert(tab.label, "tab label is required"),
            definition = tab,
        }

        if render.assets_available() then
            entry.background = render.create(options.layer or "hud", root)
            entry.background:set_position(tx + tab_width / 2, ty + height / 2)
            entry.background:set_visible(false)

            entry.label_node = render.create(options.layer or "hud", root)
            entry.label_node:set_position(tx + tab_width / 2, ty + height / 2)
        end

        -- The tab itself remains an ordinary control with ordinary activation/focus.
        entry.control = manager:add({
            id = id .. "_" .. entry.id,
            role = "tab",
            label = entry.label,
            x = tx,
            y = ty,
            width = tab_width,
            height = height,
            scope = options.scope or definition.scope,
            on_activate = function() result:select(entry.id) end,
            on_state = function() draw(entry) end,
        })

        result.tabs[#result.tabs + 1] = entry
    end

    -- Draw the initial selected/unselected state before first input update.
    draw_all()
    return result
end

return M
