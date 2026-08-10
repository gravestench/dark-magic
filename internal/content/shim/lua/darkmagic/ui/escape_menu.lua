-- Diablo II in-game Escape menu composed with Dark Magic retained controls.
--
-- This is a great example of a SMALL STATE MACHINE implemented with plain Lua.
-- The menu has several layouts/pages (`main`, `options`, `sound`, etc.). Every
-- row is still an ordinary control. The extra machinery in this file only:
--
--   * builds the different row TYPES (image, enum/value, range slider);
--   * switches which focus scope/page is active;
--   * positions the two animated pentagram focus markers;
--   * reports semantic actions back to the owning scene through callbacks.
--
-- `ui.compat` owns recovered Diablo II facts. This file owns Dark Magic's own
-- implementation and resource/control lifecycle.

local render = require("dm.render/v1")
local input = require("dm.input/v1")
local audio = require("dm.audio/v1")
local settings = require("dm.settings/v1")
local data = require("dm.data/v1")
local controls = require("darkmagic.ui.controls")
local text = require("darkmagic.ui.text")
local slider = require("darkmagic.ui.slider")
local compat = require("darkmagic.ui.compat")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local definition = assert(compat.ingame.escape_menu)
local palette = assert(manifest.palettes[definition.palette])

local M = {}

-- Lightweight Lua object pattern, same idea used by controls.Manager.
local EscapeMenu = {}
EscapeMenu.__index = EscapeMenu

-- Safe helper for optional retained nodes.
local function set_visible(node, visible)
    if node then node:set_visible(visible) end
end

local function play_select()
    if audio.exists(definition.select_sound) then
        audio.play(definition.select_sound, { bus = "ui" })
    end
end

-- Render exact-font text, but gracefully hide the node if optional font assets
-- are unavailable. `pcall` converts a Lua error into `(false,error)` values.
local function set_text_node(node, font, value, width, alignment)
    local ok, rendered_width, rendered_height = pcall(
        text.set_font,
        node,
        font,
        value,
        width,
        alignment or "center",
        { text_color = "white" }
    )

    if not ok then
        node:set_visible(false)
        return 0, 0
    end

    return rendered_width, rendered_height
end

-- Build one animated pentagram pointer. The two pointers share this structure.
local function create_pent(root)
    if not render.assets_available() then return nil end

    local node = render.create("modal", root)
    node:set_z(50)

    -- Explicit method form is used inside pcall: function, self, then arguments.
    local ok, frames, width, height = pcall(
        node.set_dc6_animation,
        node,
        definition.pentagram.sheet,
        palette,
        0,
        definition.pentagram.frames_per_second,
        "loop",
        "offsets"
    )

    if not ok then
        node:set_visible(false)
        return nil
    end

    -- Pause managed playback because this menu drives both pents from one clock.
    node:animation_pause()
    node:set_visible(false)

    return {
        node = node,
        frames = frames,
        width = width,
        height = height,
        -- One full loop duration in seconds.
        duration = frames / definition.pentagram.frames_per_second,
    }
end

-- Every row type stores different optional visuals. This helper gives all of
-- them one consistent show/hide operation.
local function show_item(item, visible)
    -- Control visibility matters just as much as render visibility: invisible
    -- menu rows must not remain focusable/clickable.
    item.control.visible = visible
    set_visible(item.art, visible)
    set_visible(item.label_node, visible)
    set_visible(item.value_node, visible)
    set_visible(item.track_node, visible)
    set_visible(item.active_node, visible)
    set_visible(item.thumb_node, visible)
end

function EscapeMenu:hide_pents()
    if self.left_pent then self.left_pent.node:set_visible(false) end
    if self.right_pent then self.right_pent.node:set_visible(false) end
end

function EscapeMenu:place_pents(item)
    -- Ignore stale state callback from a row belonging to another hidden layout.
    if self.current_layout ~= item.layout then return end
    if not self.left_pent or not self.right_pent then return end

    local half_width = item.focus_width / 2
    local gap = definition.label_side_gap
    local center_y = item.y + item.height / 2

    -- Place one pent just outside each side of the row's actual focus width.
    self.left_pent.node:set_position(
        self.center.x - half_width - gap - self.left_pent.width / 2,
        center_y
    )
    self.right_pent.node:set_position(
        self.center.x + half_width + gap + self.right_pent.width / 2,
        center_y
    )

    self.left_pent.node:set_visible(true)
    self.right_pent.node:set_visible(true)
end

function EscapeMenu:update_pent_animation(elapsed)
    if not self.left_pent or not self.right_pent then return end

    self.pent_elapsed = self.pent_elapsed + elapsed

    local right_duration = self.right_pent.duration
    if right_duration > 0 then
        -- `% duration` wraps elapsed time into the current loop position.
        self.right_pent.node:animation_seek(self.pent_elapsed % right_duration)
    end

    local left_duration = self.left_pent.duration
    if left_duration > 0 then
        local position = self.pent_elapsed % left_duration

        if definition.pentagram.left_reversed then
            -- Reverse one side while still staying in the same loop interval.
            position = (left_duration - position) % left_duration
        end

        self.left_pent.node:animation_seek(position)
    end
end

function EscapeMenu:update_value_visual(item)
    if not item.value_node or not item.values then return end

    local value = item.values[item.value_index]

    -- Original menu lays label and current value using their INTRINSIC text widths.
    -- A max-width wrap here would cause long values to overlap neighboring rows.
    local width, height = set_text_node(item.value_node, "font30", value, 0, "right")
    local right = self.center.x + definition.menu_width / 2
    item.value_node:set_position(right - width / 2, item.y + item.height / 2)
    item.value_width = width
    item.value_height = height
end

function EscapeMenu:activate(item)
    play_select()

    if item.target then
        -- Standalone options overlay begins AT `options`, so PREVIOUS MENU there
        -- means close the overlay rather than navigating to an invisible `main` page.
        if self.start_layout == "options" and self.current_layout == "options"
            and item.id == "previous_menu" then
            self.on_close()
            return
        end

        self:set_layout(item.target)
        return
    end

    if item.action == "close" then
        self.on_close()
        return
    end

    if item.action == "save_exit" then
        self.on_save_exit()
        return
    end

    if item.values and #item.values > 0 then
        -- Cycle 1..N. `% #values` wraps N back to zero, then +1 returns 1.
        item.value_index = (item.value_index % #item.values) + 1
        self:update_value_visual(item)

        if self.on_option_change then
            self.on_option_change(self.current_layout, item.id, item.values[item.value_index])
        end
    end
end

-- IMAGE ROW ---------------------------------------------------------------
-- Large main Escape labels are themselves localized DC6 art. Use combined frame
-- tiles when available, otherwise fall back to bitmap text.
function EscapeMenu:create_image_item(layout_id, layout, row, y)
    local item = {
        id = row.id,
        layout = layout_id,
        target = row.target,
        action = row.action,
        values = row.values,
        value_index = 1,
        y = y,
        height = definition.row_height,
        focus_width = definition.menu_width,
    }

    if render.assets_available() and row.sheet then
        local node = render.create("modal", self.root)
        node:set_z(30)

        -- Large localized labels are split into SPATIAL DC6 frames, commonly a
        -- 256px left piece plus a remainder. They are not animation frames.
        local ok, width, height = pcall(node.set_dc6_combined, node, row.sheet, palette, 0, 0)

        if ok then
            item.art = node
            item.focus_width = math.max(width, 1)
            item.height = math.max(height, definition.pentagram.height)
            node:set_position(self.center.x, y + item.height / 2)
        else
            node:set_visible(false)
        end
    end

    if not item.art and render.assets_available() then
        -- Fallback for missing row art: create an equivalent text label.
        local node = render.create("modal", self.root)
        node:set_z(30)
        local width, height = set_text_node(node, layout.font or "font42", row.label, definition.menu_width, "center")

        if width > 0 and height > 0 then
            item.label_node = node
            item.focus_width = math.max(width, 200)
            item.height = math.max(height, definition.pentagram.height)
            node:set_position(self.center.x, y + item.height / 2)
        end
    end

    item.control = self.manager:add({
        -- Prefix IDs with layout so identical row names on different pages cannot collide.
        id = layout_id .. ":" .. row.id,
        label = row.label,
        role = "menuitem",
        scope = layout_id,
        visible = false,
        x = self.center.x - item.focus_width / 2,
        y = y,
        width = item.focus_width,
        height = item.height,
        on_activate = function() self:activate(item) end,
        on_state = function(_, state)
            if state == "hover" or state == "focused" or state == "pressed" then
                self:place_pents(item)
            end
        end,
    })

    return item
end

-- ENUM / TEXT ROW ---------------------------------------------------------
-- Used for smaller options pages with left label and optional right-side value.
function EscapeMenu:create_enum_item(layout_id, row, y)
    local item = {
        id = row.id,
        layout = layout_id,
        target = row.target,
        action = row.action,
        values = row.values,
        value_index = 1,
        y = y,
        height = 40,
        focus_width = definition.menu_width,
    }

    if render.assets_available() then
        item.label_node = render.create("modal", self.root)
        item.label_node:set_z(30)

        -- width=0 requests intrinsic width, matching old menu spacing behavior.
        local width = set_text_node(item.label_node, "font30", row.label, 0, "left")
        local left = self.center.x - definition.menu_width / 2
        item.label_node:set_position(left + width / 2, y + item.height / 2)

        if item.values then
            item.value_node = render.create("modal", self.root)
            item.value_node:set_z(30)
            self:update_value_visual(item)
        end
    end

    item.control = self.manager:add({
        id = layout_id .. ":" .. row.id,
        label = row.label,
        role = item.values and "option" or "menuitem",
        scope = layout_id,
        visible = false,
        x = self.center.x - definition.menu_width / 2,
        y = y,
        width = definition.menu_width,
        height = item.height,
        -- Unimplemented options are PRESENT but intentionally non-interactive.
        enabled = not row.unavailable,
        on_activate = function() self:activate(item) end,
        on_state = function(_, state)
            if state == "hover" or state == "focused" or state == "pressed" then
                self:place_pents(item)
            end
        end,
    })

    return item
end

-- RANGE ROW ---------------------------------------------------------------
-- Sound/music volumes use the generic slider widget but real settings capability.
function EscapeMenu:create_range_item(layout_id, row, y)
    local range = assert(row.range)
    local item = {
        id = row.id,
        layout = layout_id,
        y = y,
        height = 40,
        focus_width = definition.menu_width,
    }

    if render.assets_available() then
        item.label_node = render.create("modal", self.root)
        item.label_node:set_z(30)
        local width, height = set_text_node(item.label_node, "font30", row.label, 190, "left")
        item.label_node:set_position(self.center.x - 155, y + item.height / 2)
        item.label_width, item.label_height = width, height
    end

    item.control = slider.create(self.root, self.manager, layout_id .. ":" .. row.id, {
        x = self.center.x - 10,
        y = y + 1,
        width = definition.option_assets.range_width,
        height = definition.option_assets.range_height,
        thumb_size = definition.option_assets.range_thumb_size,
        min = range.min,
        max = range.max,
        step = range.step,
        -- Initial presentation reads current setting through capability.
        value = settings.get(range.setting),
        scope = layout_id,
    }, row.label, {
        layer = "modal",
        palette = palette,
        track_sheet = definition.option_assets.range_track,
        thumb_sheet = definition.option_assets.range_thumb,
        show_label = false,
        show_value = false,
        on_change = function(_, value)
            -- Slider interaction asks settings capability to change the value.
            settings.set(range.setting, value)
            if self.on_option_change then self.on_option_change(layout_id, row.id, value) end
        end,
    })

    -- Mirror slider child handles into the generic item shape so show_item can
    -- hide/show every row type with one function.
    item.track_node = item.control.track_node
    item.active_node = item.control.active_node
    item.thumb_node = item.control.thumb_node
    return item
end

function EscapeMenu:create_layout(layout_id, layout)
    -- Main font42 rows are taller than smaller option rows.
    local row_height = layout.font == "font42" and definition.row_height or 40
    local title_height = layout.title and 50 or 0
    local total_height = title_height + (#layout.rows * row_height)

    -- Center the entire page vertically around configured menu center.
    local y = self.center.y - total_height / 2
    local items = {}

    if layout.title and render.assets_available() then
        local title_node = render.create("modal", self.root)
        title_node:set_z(30)
        local _, height = set_text_node(title_node, "font42", layout.title, definition.menu_width, "center")
        title_node:set_position(self.center.x, y + height / 2)
        title_node:set_visible(false)
        self.titles[layout_id] = title_node
        y = y + title_height
    end

    for _, row in ipairs(layout.rows) do
        local item

        -- Choose row constructor from row data rather than hard-coding each page.
        if row.range then
            item = self:create_range_item(layout_id, row, y)
        elseif layout.font == "font42" and not row.values then
            item = self:create_image_item(layout_id, layout, row, y)
        else
            item = self:create_enum_item(layout_id, row, y)
        end

        items[#items + 1] = item
        self.items_by_id[layout_id .. ":" .. row.id] = item
        y = y + row_height
    end

    self.items[layout_id] = items
end

function EscapeMenu:set_layout(layout_id)
    local layout = assert(definition.layouts[layout_id], "unknown escape-menu layout: " .. tostring(layout_id))
    self.current_layout = layout_id
    self:hide_pents()

    -- FIRST hide all pages. This also disables their controls by visible=false.
    -- Then change focus scope. This order prevents a brief accidental focus jump
    -- into a newly selected page before its authored default is applied.
    for id, items in pairs(self.items) do
        for _, item in ipairs(items) do show_item(item, false) end
        set_visible(self.titles[id], false)
    end

    self.manager:set_scope(layout_id)

    -- Reveal just the requested page.
    local items = assert(self.items[layout_id], "missing escape-menu layout controls: " .. layout_id)
    for _, item in ipairs(items) do show_item(item, true) end
    set_visible(self.titles[layout_id], true)

    -- Prefer recovered/authored default focus. Otherwise move to first eligible row.
    local default = layout.default_focus and self.items_by_id[layout_id .. ":" .. layout.default_focus]
    if default then
        assert(self.manager:set_focus(default.control), "escape-menu default focus is not eligible: " .. default.control.id)
    else
        self.manager:move_focus(1)
    end

    local focused = self.manager.focus and self.items_by_id[self.manager.focus.id]
    if focused then self:place_pents(focused) end
end

function EscapeMenu:update(elapsed)
    self:update_pent_animation(elapsed)
    self.manager:update()

    -- Re-place pents after input in case focus moved this frame.
    local focused = self.manager.focus and self.items_by_id[self.manager.focus.id]
    if focused then self:place_pents(focused) end
end

function EscapeMenu:accessibility()
    -- Reuse same value-only accessibility snapshot as every other control manager.
    return self.manager:accessibility()
end

function M.new(root, options)
    options = options or {}

    local self = setmetatable({
        root = root,
        center = options.center or definition.center,
        -- Each menu page is a focus scope. wrap_focus=false reproduces the
        -- desired stop-at-ends navigation behavior here.
        manager = controls.new({ active_scope = options.start_layout or "main", wrap_focus = false }),
        start_layout = options.start_layout or "main",
        current_layout = nil,
        items = {},
        items_by_id = {},
        titles = {},
        pent_elapsed = 0,
        -- Owning scene decides what these semantic actions actually do.
        on_close = assert(options.on_close, "escape-menu on_close callback is required"),
        on_save_exit = options.on_save_exit or options.on_close,
        on_option_change = options.on_option_change,
    }, EscapeMenu)

    self.left_pent = create_pent(root)
    self.right_pent = create_pent(root)

    -- Build every page once as retained objects; switching layouts only toggles
    -- visibility/focus scope instead of reconstructing the menu repeatedly.
    for layout_id, layout in pairs(definition.layouts) do
        self:create_layout(layout_id, layout)
    end

    self:set_layout(self.start_layout)
    return self
end

return M