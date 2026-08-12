-- Legacy vendor catalog backed by fixed-tick interaction and item authority.
--
-- This file is deliberately a presentation adapter. It chooses which tab to
-- draw and translates a click into "buy this item". It never invents stock,
-- prices, placement coordinates, or permission to trade.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")
local locale = require("engine.locale/v1")
local controls = require("d2.ui.controls")
local button = require("d2.ui.button")
local text = require("d2.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2.presentation/v1"))
local screen = assert(manifest.screens.vendor)

local function number(record, key)
    return assert(tonumber(record[key]), "Inventory.txt field is not numeric: " .. key)
end

local function monster_inventory()
    local records = require("engine.records/v1")
    for _, row in ipairs(assert(records.load(screen.item_grid.records))) do
        if row.class == screen.item_grid.record_class then return row end
    end
    error("Inventory.txt has no Monster record")
end

local function item_at(snapshot, category, page, column, row)
    for _, item in ipairs(snapshot.items) do
        if item.container == "vendor" and item.category == category and item.page == page
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

local function category_allowed(context, category)
    for _, candidate in ipairs(context.categories) do
        if candidate == category then return true end
    end
    return false
end

local function select_tab(self, index)
    self.category = screen.tabs.categories[index]
    self.page = index == 3 and 1 or 0
    for candidate, tab in ipairs(self.tabs) do
        tab:set_dc6(screen.tabs.sheet, self.palette, 0, candidate == index and candidate + 3 or candidate - 1)
    end
end

local function tab_label(root, value, x, y, width)
    local node = render.create("modal", root)
    local _, height = text.set_font(node, "font16", value, width, "center", {text_color="white"})
    node:set_position(x + width / 2, y + height / 2)
end

local function refresh(self)
    self.snapshot = assert(self.items.snapshot())
    self.context = assert(self.interaction.snapshot())
    self.__darkmagic_item_held = held_item(self.snapshot) ~= nil

    local cursor_x, cursor_y = input.cursor()
    for _, item in ipairs(self.snapshot.items) do
        local drawing = self.item_nodes[item.id]
        local visible = item.container == "held" or (item.container == "vendor"
            and item.category == self.category and item.page == self.page)

        if visible and drawing == nil and item.inventory_dc6 ~= "" then
            local node = render.create("modal", self.root)
            local width, height = node:set_dc6(item.inventory_dc6, manifest.palettes.units, 0, 0)
            drawing = {node=node, width=width, height=height}
            self.item_nodes[item.id] = drawing
        end

        if drawing ~= nil then
            drawing.node:set_visible(visible)
            if item.container == "held" then
                drawing.node:set_position(cursor_x, cursor_y)
            elseif visible then
                drawing.node:set_position(
                    self.grid_left + item.x * self.cell_width + drawing.width / 2,
                    self.grid_top + item.y * self.cell_height + drawing.height / 2
                )
            end
        end
    end

    text.set(self.gold_text, "panel_heading", tostring(self.snapshot.stashed_gold), 180, "right")
end

local function activate_cell(self, column, row)
    -- Buying while another item is held is invalid. Authority would reject it,
    -- but avoiding a doomed command also gives the UI predictable behavior.
    if held_item(self.snapshot) ~= nil then return end
    local item = item_at(self.snapshot, self.category, self.page, column, row)
    if item ~= nil then self.items.buy_to_held(item.id, self.context.vendor) end
end

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("modal")
        self.controls = controls.new()
        self.item_nodes = {}
        self.category, self.page = "misc", 0
        if not render.assets_available() then return end

        self.items = require("engine.items/v1")
        self.interaction = require("engine.interaction/v1")
        self.context = assert(self.interaction.snapshot())
        assert(self.context.active and self.context.vendor ~= "", "vendor overlay requires an active vendor interaction")

        local palette = manifest.palettes[screen.palette]
		self.palette = palette
        local panel = render.create("modal", self.root)
        local panel_width, panel_height = panel:set_dc6_combined(screen.sheet, palette, 0, 0)
        panel:set_position(screen.x + panel_width / 2, screen.y + panel_height / 2)

        local record = monster_inventory()
        self.grid_left = screen.x + number(record, "gridLeft") - number(record, "invLeft")
        self.grid_top = screen.y + number(record, "gridTop")
        self.cell_width, self.cell_height = number(record, "gridBoxWidth"), number(record, "gridBoxHeight")
        self.columns, self.rows = number(record, "gridX"), number(record, "gridY")

        -- Riiablo corroborates the original tab layout: four adjacent frames at
        -- the panel top; active art is exactly four frames after inactive art.
        local tab_x = screen.x
		self.tabs = {}
        for index, category in ipairs(screen.tabs.categories) do
            local node = render.create("modal", self.root)
			self.tabs[index] = node
            local width, height = node:set_dc6(screen.tabs.sheet, palette, 0, index - 1)
            node:set_position(tab_x + width / 2, screen.y + height / 2)
			tab_label(self.root, locale.text(screen.tabs.labels[index]), tab_x + 2, screen.y + 7, width - 4)
            local this_index, this_category, this_x = index, category, tab_x
            self.controls:add({
                id="vendor_tab_" .. index, label=locale.text(screen.tabs.labels[index]),
                x=this_x, y=screen.y, width=width, height=height,
                enabled=category_allowed(self.context, category),
                on_activate=function() select_tab(self, this_index) end,
            })
            tab_x = tab_x + width + 1
        end
		select_tab(self, 4)

        for row = 0, self.rows - 1 do
            for column = 0, self.columns - 1 do
                local cell_column, cell_row = column, row
                self.controls:add({
                    id=string.format("vendor_%d_%d", column, row), label="Vendor item",
                    x=self.grid_left + column * self.cell_width,
                    y=self.grid_top + row * self.cell_height,
                    width=self.cell_width, height=self.cell_height,
                    on_activate=function() activate_cell(self, cell_column, cell_row) end,
                })
            end
        end

		text.create(self.root, "panel_heading", locale.text("stash"), screen.x + 110, screen.y + panel_height - 73, 180, "left")
        self.gold_text = render.create("modal", self.root)
        self.gold_text:set_position(screen.x + 110, screen.y + panel_height - 65)

        local close = {sheet="data/global/ui/PANEL/buysellbtn.DC6", palette="sky", up_frame=10, down_frame=11,
            x=screen.x + screen.close.x, y=screen.y + panel_height - screen.close.bottom_inset - 32, width=32, height=32}
        button.create(self.root, self.controls, "close", close, locale.text(screen.close.label), {
            layer="modal", show_label=false, tooltip=locale.text(screen.close.label),
            on_activate=function() self.interaction.close(); scenes.toggle_overlay("vendor", "left") end,
        })
        refresh(self)
    end,

    update = function(self)
        if self.items then refresh(self) end
        self.controls:update()
        if input.pressed("cancel") then
            if self.interaction then self.interaction.close() end
            scenes.pop()
        end
    end,
}
