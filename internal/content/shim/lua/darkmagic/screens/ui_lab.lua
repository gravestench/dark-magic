-- Interactive Diablo II UI/widget laboratory.
--
-- ui_lab exercises the same retained render nodes and Lua control hooks used by
-- shipping screens. It is intentionally not a second GUI implementation.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local label_button = require("darkmagic.ui.label_button")
local checkbox = require("darkmagic.ui.checkbox")
local text_field = require("darkmagic.ui.text_field")
local slider = require("darkmagic.ui.slider")
local scrollbar = require("darkmagic.ui.scrollbar")
local list = require("darkmagic.ui.list")
local tabs = require("darkmagic.ui.tabs")
local panel = require("darkmagic.ui.panel")
local progress_bar = require("darkmagic.ui.progress_bar")
local tooltip = require("darkmagic.ui.tooltip")
local dialog = require("darkmagic.ui.dialog")
local text = require("darkmagic.ui.text")
local cursor = require("darkmagic.ui.cursor")
local compat = require("darkmagic.ui.compat")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local ui_lab = {}

local function copy_at(source, x, y)
    local result = {}
    for key, value in pairs(source) do result[key] = value end
    result.x = x
    result.y = y
    return result
end

local function put(root, style, value, left, top, width, alignment)
    local node = render.create("hud", root)
    local _, height = text.set(node, style, value, width, alignment or "left")
    node:set_position(left + width / 2, top + height / 2)
    return node
end

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

    self.controls = controls.new()
    self.status = put(self.root, "font_lab_caption", "status: ready", 410, 535, 330, "left")
    local function status(value) set_line(self.status, "status: " .. value, 330) end

    -- AUTHORED BUTTON: uses the same recovered 3-wide main-menu facts and
    -- draw-mode-4 Exocet treatment as the shipping frontend.
    local authored_source = compat.screen_control("main_menu", "single_player", manifest.screens.main_menu.controls.single_player)
    local authored = copy_at(authored_source, 60, 105)
    self.primary = button.create(self.root, self.controls, "authored_button", authored, "AUTHORED BUTTON", {
        tooltip = "Mouse-down depresses; mouse-up inside activates; label uses D2 draw mode 4",
        on_activate = function() status("authored button activated") end,
    })

    -- DISABLED BUTTON: verifies focus exclusion and authored disabled frames.
    local disabled_source = compat.screen_control("main_menu", "credits", manifest.screens.main_menu.controls.credits)
    local disabled = copy_at(disabled_source, 60, 150)
    self.disabled = button.create(self.root, self.controls, "disabled_button", disabled, "DISABLED", { enabled = false })

    put(self.root, "font_lab_caption", "TEXT-ONLY / LABEL BUTTON", 60, 198, 280, "left")
    self.label_button = label_button.create(self.root, self.controls, {
        id = "label_button", x = 60, y = 218, width = 272, height = 24,
    }, "TEXT-ONLY ACTION", {
        on_activate = function() status("label button activated") end,
    })

    put(self.root, "font_lab_caption", "CHECKBOX", 60, 255, 280, "left")
    self.checkbox = checkbox.create(self.root, self.controls, "checkbox", {
        x = 60, y = 278, checked = true, palette = "fechar",
    }, "Expansion character", {
        on_change = function(_, checked) status("checkbox = " .. tostring(checked)) end,
    })

    put(self.root, "font_lab_caption", "TEXT FIELD", 60, 316, 280, "left")
    self.text_field = text_field.create(self.root, self.controls, "text_field", {
        x = 60, y = 342, kind = "name", value = "DarkMagic", max_length = 15, palette = "fechar",
    }, "Character Name", {
        on_change = function(_, value) status("text = " .. value) end,
    })

    -- PROGRESS is driven by the slider below to prove non-interactive UI values
    -- can share the same retained update path.
    self.progress = progress_bar.create(self.root, {
        x = 60, y = 492, width = 272, height = 12, min = 0, max = 100, value = 40,
    }, "PROGRESS BAR", { show_value = true })

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
            self.progress:set_value(value * 100)
            status("slider = " .. tostring(value))
        end,
    })

    -- PAGED LIST: single activation selects; activating the selected row again
    -- dispatches on_activate, matching the character-list style distinction.
    put(self.root, "font_lab_caption", "SELECTABLE / PAGED LIST", 405, 98, 275, "left")
    self.list = list.create(self.root, self.controls, "sample_list", {
        x = 405, y = 122, width = 260, row_height = 27, page_size = 4,
    }, {
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
        on_page = function(page, pages) self.list_page = string.format("%d/%d", page, pages) end,
    })

    self.previous_page = label_button.create(self.root, self.controls, {
        id = "list_previous", x = 405, y = 236, width = 80, height = 24,
    }, "< PREV", { on_activate = function() self.list:previous_page(); status("list page " .. (self.list_page or "")) end })
    self.next_page = label_button.create(self.root, self.controls, {
        id = "list_next", x = 585, y = 236, width = 80, height = 24,
    }, "NEXT >", { on_activate = function() self.list:next_page(); status("list page " .. (self.list_page or "")) end })

    -- Authored TextSlid scrollbar recovered by OpenDiablo2. The arrows and
    -- gutter/thumb are actual DC6 frames when the game asset is mounted.
    put(self.root, "font_lab_caption", "TEXT SCROLLBAR", 680, 98, 105, "left")
    self.scrollbar = scrollbar.create(self.root, self.controls, "text_scrollbar", {
        x = 738, y = 122, height = 138, min = 0, max = 100, step = 10, value = 30,
    }, nil, {
        show_value = false,
        on_change = function(_, value) status("scrollbar = " .. tostring(value)) end,
    })

    -- TABS / mutually-exclusive selection.
    put(self.root, "font_lab_caption", "TAB / SELECTION GROUP", 405, 280, 330, "left")
    self.tabs = tabs.create(self.root, self.controls, "tabs", {
        x = 405, y = 303, tab_width = 100, height = 25, gap = 5, selected = "one",
    }, {
        {id="one", label="TAB ONE"},
        {id="two", label="TAB TWO"},
        {id="three", label="TAB THREE"},
    }, {
        on_change = function(id) status("tab = " .. id) end,
    })

    -- PANEL / CONTAINER: demonstrates grouped backing and tracked child nodes.
    self.panel = panel.create(self.root, {
        x = 405, y = 346, width = 330, height = 100,
    }, {
        background_color = {28, 23, 18, 230},
    })
    self.panel_title = self.panel:node("hud")
    text.set(self.panel_title, "character_create_option", "PANEL / CONTAINER", 300, "center")
    self.panel_title:set_position(570, 365)
    self.panel_copy = self.panel:node("hud")
    text.set(self.panel_copy, "font_lab_caption", "Groups retained nodes and provides one visibility/clip boundary for composite widgets.", 300, "center")
    self.panel_copy:set_position(570, 402)

    -- STANDALONE TOOLTIP + MODAL DIALOG.
    self.tip = tooltip.create(self.root, "Standalone tooltip: viewport clamping + retained text/backing", 650, 487, {
        origin_x = "center", origin_y = "bottom", max_width = 220,
    })
    self.tooltip_target = self.controls:add({
        id = "tooltip_target", role = "button", label = "Tooltip target",
        x = 550, y = 460, width = 200, height = 25,
        on_state = function(_, state) self.tip:set_visible(state == "hover" or state == "focused") end,
    })
    self.tooltip_label = put(self.root, "label_button_normal", "HOVER FOR TOOLTIP", 550, 460, 200, "center")

    self.modal_button = label_button.create(self.root, self.controls, {
        id = "modal_button", x = 405, y = 495, width = 130, height = 25,
    }, "OPEN MODAL", {
        on_activate = function()
            if self.modal and self.modal.open then return end
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

    -- Existing explicit cursor is intentionally retained; the shell-wide cursor
    -- policy recognizes and reuses scene-owned cursors rather than duplicating it.
    self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
end

function ui_lab.update(self)
    if self.modal and self.modal.open then
        self.modal:update()
        if input.pressed("cancel") then self.modal:close(false) end
    else
        self.controls:update()
        if input.pressed("cancel") then scenes.replace("main_menu") end
    end
    if self.cursor then self.cursor:update() end
end

return ui_lab
