-- Interactive Diablo II UI/widget laboratory.
--
-- THIS SCENE IS A PLAYGROUND FOR MOD AUTHORS.
--
-- Every widget below is the SAME ordinary Lua module used by shipping screens.
-- There is no special editor-only Button, Slider, List, or Dialog. That makes
-- UI Lab both a visual test and living documentation: copy one small example,
-- change its data/callback, and you already have the basic pattern for a mod.
--
-- Reading tip: do not try to memorize the whole create() function. Each labeled
-- block is intentionally independent. Pick the widget you care about, read that
-- block, then open its implementation under `d2/ui/` for the next layer.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")
local controls = require("d2legacy.ui.controls")
local button = require("d2legacy.ui.button")
local label_button = require("d2legacy.ui.label_button")
local checkbox = require("d2legacy.ui.checkbox")
local text_field = require("d2legacy.ui.text_field")
local slider = require("d2legacy.ui.slider")
local scrollbar = require("d2legacy.ui.scrollbar")
local list = require("d2legacy.ui.list")
local tabs = require("d2legacy.ui.tabs")
local panel = require("d2legacy.ui.panel")
local progress_bar = require("d2legacy.ui.progress_bar")
local tooltip = require("d2legacy.ui.tooltip")
local dialog = require("d2legacy.ui.dialog")
local text = require("d2legacy.ui.text")
local cursor = require("d2legacy.ui.cursor")
local compat = require("d2legacy.ui.compat")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local ui_lab = {}

-- These are semantic text styles from the manifest. The lab uses them as a tiny
-- visual language for normal/hover/active diagnostic controls.
local lab_control_normal = "ui_lab_control_idle"
local lab_control_hover = "ui_lab_control_hover"
local lab_control_active = "ui_lab_control_active"

-- Make a shallow copy of a control definition, then move only its x/y position.
-- This is how the lab can reuse REAL recovered main-menu art while displaying it
-- somewhere convenient for testing without mutating shared compat/manifest data.
local function copy_at(source, x, y)
    local result = {}
    for key, value in pairs(source) do result[key] = value end
    result.x = x
    result.y = y
    return result
end

-- Small convenience for one retained text label positioned from top-left box facts.
local function put(root, style, value, left, top, width, alignment)
    local node = render.create("hud", root)
    local _, height = text.set(node, style, value, width, alignment or "left")
    node:set_position(left + width / 2, top + height / 2)
    return node
end

-- Diagnostic rectangle helper. The UI lab deliberately uses plain fills for a
-- few backing surfaces so it does not require additional proprietary art.
local function box(root, x, y, width, height, r, g, b, a)
    local node = render.create("hud", root)
    node:fill_rect(width, height, r, g, b, a or 255)
    node:set_position(x + width / 2, y + height / 2)
    return node
end

local function set_line(node, value, width)
    text.set(node, "font_lab_caption", value, width, "left")
end

function ui_lab.create(self)
    self.root = render.create("hud")
    self.background = box(self.root, 0, 0, manifest.resolution.width, manifest.resolution.height, 18, 15, 13, 255)
    put(self.root, "font_lab_heading", "UI LAB", 40, 14, 720, "center")
    put(self.root, "font_lab_caption", "Reusable Lua widgets + recovered Diablo II interaction behavior", 40, 55, 720, "center")
    put(self.root, "font_lab_caption", "Mouse-up activates   drag ranges   arrows navigate/adjust   Enter activates   Esc returns", 40, 565, 720, "center")

    -- ONE shared manager owns focus/hit testing for all ordinary controls below.
    self.controls = controls.new()

    -- Status line makes callback results visible without needing logs/debugger.
    self.status = put(self.root, "font_lab_caption", "status: ready", 410, 535, 330, "left")
    local function status(value) set_line(self.status, "status: " .. value, 330) end

    -- AUTHORED BUTTON --------------------------------------------------------
    -- Use the same recovered 3-wide main-menu facts and draw-mode-4 Exocet
    -- treatment as the shipping frontend, just copied to a lab position.
    local authored_source = compat.screen_control("main_menu", "single_player", manifest.screens.main_menu.controls.single_player)
    local authored = copy_at(authored_source, 60, 105)
    self.primary = button.create(self.root, self.controls, "authored_button", authored, "AUTHORED BUTTON", {
        tooltip = "Mouse-down depresses; mouse-up inside activates; label uses D2 draw mode 4",
        -- Callback expresses MEANING. button.lua itself only knows presentation/activation.
        on_activate = function() status("authored button activated") end,
    })

    -- DISABLED BUTTON --------------------------------------------------------
    -- Same widget, but enabled=false proves the manager excludes it from focus
    -- and button.lua can display authored disabled frames.
    local disabled_source = compat.screen_control("main_menu", "credits", manifest.screens.main_menu.controls.credits)
    local disabled = copy_at(disabled_source, 60, 150)
    self.disabled = button.create(self.root, self.controls, "disabled_button", disabled, "DISABLED", { enabled = false })

    -- LABEL BUTTON -----------------------------------------------------------
    -- A text-only action still uses the same controls.Manager as bitmap buttons.
    put(self.root, "font_lab_caption", "TEXT-ONLY / LABEL BUTTON", 60, 198, 280, "left")
    self.label_button = label_button.create(self.root, self.controls, {
        id = "label_button", x = 60, y = 218, width = 272, height = 24,
    }, "TEXT-ONLY ACTION", {
        normal_style = lab_control_normal,
        hover_style = lab_control_hover,
        pressed_style = lab_control_active,
        on_activate = function() status("label button activated") end,
    })

    -- CHECKBOX ---------------------------------------------------------------
    -- controls.lua owns boolean toggle. Widget callback receives the new value.
    put(self.root, "font_lab_caption", "CHECKBOX", 60, 255, 280, "left")
    self.checkbox = checkbox.create(self.root, self.controls, "checkbox", {
        x = 60, y = 278, checked = true, palette = "fechar",
    }, "Expansion character", {
        -- `_` is unused control parameter; `checked` is value we care about.
        on_change = function(_, checked) status("checkbox = " .. tostring(checked)) end,
    })

    -- TEXT FIELD -------------------------------------------------------------
    -- Generic manager owns text editing/cursor; visual helper owns D2 text-box art.
    put(self.root, "font_lab_caption", "TEXT FIELD", 60, 316, 280, "left")
    self.text_field = text_field.create(self.root, self.controls, "text_field", {
        x = 60, y = 342, kind = "name", value = "DarkMagic", max_length = 15, palette = "fechar",
    }, "Character Name", {
        on_change = function(_, value) status("text = " .. value) end,
    })

    -- PROGRESS BAR -----------------------------------------------------------
    -- Passive display; no controls.Manager registration. Slider below drives it.
    self.progress = progress_bar.create(self.root, {
        x = 60, y = 492, width = 272, height = 12, min = 0, max = 100, value = 40,
    }, "PROGRESS BAR", { show_value = true })

    -- SLIDER -----------------------------------------------------------------
    -- This example uses recovered authored OptBarC/OptSkull art rather than the
    -- slider helper's diagnostic rectangle fallback.
    local settings_slider = compat.widgets.option_slider
    put(self.root, "font_lab_caption", "SETTINGS SLIDER", 60, 397, 290, "left")
    self.slider = slider.create(self.root, self.controls, "slider", {
        x = 60, y = 424,
        width = settings_slider.width,
        height = settings_slider.height,
        thumb_size = settings_slider.thumb_size,
        min = 0, max = 1, step = 0.05, value = 0.4,
    }, "Volume", {
        palette = assert(manifest.palettes[settings_slider.palette]),
        track_sheet = settings_slider.track_sheet,
        thumb_sheet = settings_slider.thumb_sheet,
        show_label = false,
        show_value = false,
        on_change = function(_, value)
            -- One widget callback can update another ordinary Lua widget.
            self.progress:set_value(value * 100)
            status("slider = " .. tostring(value))
        end,
    })

    -- PAGED LIST -------------------------------------------------------------
    -- Physical row controls stay stable while page data changes. A first click
    -- selects; activating selected row again dispatches on_activate.
    put(self.root, "font_lab_caption", "SELECTABLE / PAGED LIST", 405, 98, 275, "left")
    self.list = list.create(self.root, self.controls, "sample_list", {
        x = 405, y = 122, width = 260, row_height = 27, page_size = 4,
    }, {
        -- Ordinary Lua tables are the sample model; list.lua only needs IDs/labels.
        {id="amazon", label="Amazon"},
        {id="sorceress", label="Sorceress"},
        {id="necromancer", label="Necromancer"},
        {id="paladin", label="Paladin"},
        {id="barbarian", label="Barbarian"},
        {id="assassin", label="Assassin"},
        {id="druid", label="Druid"},
    }, {
        on_select = function(item) status("selected list item: " .. item.label) end,
        on_activate = function(item) status("activated list item: " .. item.label) end,
        -- Remember friendly page label for separate Prev/Next buttons.
        on_page = function(page, pages) self.list_page = string.format("%d/%d", page, pages) end,
    })

    self.previous_page = label_button.create(self.root, self.controls, {
        id = "list_previous", x = 405, y = 236, width = 80, height = 24,
    }, "< PREV", {
        normal_style = lab_control_normal, hover_style = lab_control_hover, pressed_style = lab_control_active,
        -- Semicolon here separates two Lua statements on one line. The project
        -- style usually avoids semicolons; this compact lab callback is existing code.
        on_activate = function() self.list:previous_page(); status("list page " .. (self.list_page or "")) end,
    })

    self.next_page = label_button.create(self.root, self.controls, {
        id = "list_next", x = 585, y = 236, width = 80, height = 24,
    }, "NEXT >", {
        normal_style = lab_control_normal, hover_style = lab_control_hover, pressed_style = lab_control_active,
        on_activate = function() self.list:next_page(); status("list page " .. (self.list_page or "")) end,
    })

    -- SCROLLBAR --------------------------------------------------------------
    -- Authored TextSlid arrows/gutter/thumb become one composed Lua widget.
    put(self.root, "font_lab_caption", "TEXT SCROLLBAR", 680, 98, 105, "left")
    self.scrollbar = scrollbar.create(self.root, self.controls, "text_scrollbar", {
        x = 738, y = 122, height = 138, min = 0, max = 100, step = 10, value = 30,
    }, nil, {
        show_value = false,
        on_change = function(_, value) status("scrollbar = " .. tostring(value)) end,
    })

    -- TABS -------------------------------------------------------------------
    -- Mutually-exclusive selection group built from ordinary controls.
    put(self.root, "font_lab_caption", "TAB / SELECTION GROUP", 405, 280, 330, "left")
    self.tabs = tabs.create(self.root, self.controls, "tabs", {
        x = 405, y = 303, tab_width = 100, height = 25, gap = 5, selected = "one",
    }, {
        {id="one", label="TAB ONE"},
        {id="two", label="TAB TWO"},
        {id="three", label="TAB THREE"},
    }, {
        normal_style = lab_control_normal,
        hover_style = lab_control_hover,
        pressed_style = lab_control_active,
        selected_style = lab_control_active,
        on_change = function(id) status("tab = " .. id) end,
    })

    -- PANEL / CONTAINER ------------------------------------------------------
    -- A panel is retained-node grouping/clip/visibility convenience, not a scene.
    self.panel = panel.create(self.root, {
        x = 405, y = 346, width = 330, height = 100,
    }, {
        background_color = {28, 23, 18, 230},
    })

    -- panel:node creates + tracks child in one call.
    self.panel_title = self.panel:node("hud")
    text.set(self.panel_title, "character_create_option", "PANEL / CONTAINER", 300, "center")
    self.panel_title:set_position(570, 365)

    self.panel_copy = self.panel:node("hud")
    text.set(self.panel_copy, "font_lab_caption", "Groups retained nodes and provides one visibility/clip boundary for composite widgets.", 300, "center")
    self.panel_copy:set_position(570, 402)

    -- TOOLTIP ---------------------------------------------------------------
    -- Standalone tooltip demonstrates viewport clamping and anchor origin rules.
    self.tip = tooltip.create(self.root, "Standalone tooltip: viewport clamping + retained text/backing", 650, 487, {
        origin_x = "center", origin_y = "bottom", max_width = 220,
    })

    -- This target uses controls.Manager DIRECTLY to show that widgets are optional
    -- convenience. A mod can register a plain control table itself.
    self.tooltip_target = self.controls:add({
        id = "tooltip_target", role = "button", label = "Tooltip target",
        x = 550, y = 460, width = 200, height = 25,
        on_state = function(_, state)
            local active = state == "hover" or state == "focused" or state == "pressed"
            self.tip:set_visible(active)

            -- Nested conditional expression selects active/hover/normal text style.
            text.set(
                self.tooltip_label,
                state == "pressed" and lab_control_active or (active and lab_control_hover or lab_control_normal),
                "HOVER FOR TOOLTIP",
                200,
                "center"
            )
        end,
    })
    self.tooltip_label = put(self.root, lab_control_normal, "HOVER FOR TOOLTIP", 550, 460, 200, "center")

    -- MODAL ------------------------------------------------------------------
    self.modal_button = label_button.create(self.root, self.controls, {
        id = "modal_button", x = 405, y = 495, width = 130, height = 25,
    }, "OPEN MODAL", {
        normal_style = lab_control_normal,
        hover_style = lab_control_hover,
        pressed_style = lab_control_active,
        on_activate = function()
            -- Do not open a second copy while current dialog is already alive.
            if self.modal and self.modal.open then return end

            -- Reuse real character-select dialog geometry to exercise production path.
            local definition = manifest.screens.character_select.delete_dialog
            self.modal = dialog.confirm(
                self.root,
                definition,
                manifest.fonts.exocet10,
                manifest.palettes[definition.palette],
                manifest.palettes.units,
                "Modal confirmation widget",
                "YES",
                "NO",
                function(decision) status("modal result = " .. tostring(decision)) end
            )
        end,
    })

    -- Explicit scene-owned cursor is reused by shell-wide cursor decorator; it
    -- does not create a duplicate automatic pointer.
    self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
end

function ui_lab.update(self)
    if self.modal and self.modal.open then
        -- While modal is open, do NOT update controls behind it.
        self.modal:update()
        if input.pressed("cancel") then self.modal:close(false) end
    else
        self.controls:update()
        if input.pressed("cancel") then scenes.replace("main_menu") end
    end

    if self.cursor then self.cursor:update() end
end

return ui_lab
