-- Saved-character selection scene.
--
-- Save enumeration, selection, and deletion use opaque engine-owned IDs. Lua
-- owns paging, focus, presentation, confirmation, and activation policy.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local saves = require("dm.save/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local ui_button = require("darkmagic.ui.button")
local cursor = require("darkmagic.ui.cursor")
local dc6 = require("darkmagic.ui.dc6")
local dialog = require("darkmagic.ui.dialog")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.character_select
local font = manifest.fonts.exocet10

local function page_count(character_count, page_size)
    return math.max(1, math.ceil(character_count / page_size))
end

return {
    create = function(self)
        self.characters = saves.characters()
        if #self.characters == 0 then
            scenes.replace("character_create")
            return
        end

        self.root = render.create("hud")
        self.background = dc6.frontend_background(
            self.root,
            "hud",
            screen.background,
            manifest.palettes[screen.palette],
            manifest.layouts.frontend_tiles
        )
        self.controls = controls.new()
        self.now = 0
        self.page = 1
        self.page_size = screen.grid.columns * screen.grid.rows
        self.selected_id = self.characters[1].id
        self.slots = {}
        self.class_presentations = {}
        for _, definition in ipairs(manifest.screens.character_create.classes) do
            self.class_presentations[definition.class] = definition
        end
        if render.assets_available() then
            self.title = render.create("hud", self.root)
            self.title:set_z(100)
        end

        local function update_title(character)
            if not self.title or not character then
                return
            end
            local definition = screen.title
            local title_font = manifest.fonts[definition.font]
            self.title:set_text(
                title_font.table,
                title_font.sheet,
                manifest.palettes[title_font.palette],
                character.name,
                { red = 210, green = 180, blue = 110, max_width = definition.width, align = "center" }
            )
            self.title:set_position(definition.x, definition.y)
        end

        local function launch_selected()
            if not self.selected_id then
                return
            end
            assert(saves.select(self.selected_id))
            scenes.replace("game_loading")
        end

        local function select_character(character)
            self.selected_id = character.id
            update_title(character)
            for _, slot in ipairs(self.slots) do
                for _, piece in ipairs(slot.selection or {}) do
                    piece:set_visible(slot.character ~= nil and slot.character.id == self.selected_id)
                end
            end
        end

        local function activate_character(character)
            local repeated_pointer = input.pressed("pointer_primary")
                and self.last_pointer_id == character.id
                and self.now - self.last_pointer_time <= screen.double_activation_seconds
            select_character(character)
            if repeated_pointer then
                launch_selected()
            end
            if input.pressed("pointer_primary") then
                self.last_pointer_id = character.id
                self.last_pointer_time = self.now
            end
        end

        -- Slot controls are stable; paging only swaps the character records and
        -- retained visuals they reference.
        for slot_index = 1, self.page_size do
            local column = (slot_index - 1) % screen.grid.columns
            local row = math.floor((slot_index - 1) / screen.grid.columns)
            local x = screen.grid.x + column * screen.grid.column_step
            local y = screen.grid.y + row * screen.grid.row_step
            local slot = {}
            slot.control = self.controls:add({
                id = "character_" .. slot_index,
                label = "Character slot " .. slot_index,
                x = x,
                y = y,
                width = screen.grid.cell_width,
                height = screen.grid.cell_height,
                on_activate = function()
                    if slot.character then
                        activate_character(slot.character)
                    end
                end,
            })
            if render.assets_available() then
                slot.selection = {}
                local selection_x = x
                for frame = 0, 1 do
                    local piece = render.create("hud", self.root)
                    local width, height = piece:set_dc6(
                        screen.selection,
                        manifest.palettes[screen.selection_palette],
                        0,
                        frame
                    )
                    piece:set_position(selection_x + width / 2, y + height / 2)
                    selection_x = selection_x + width
                    slot.selection[#slot.selection + 1] = piece
                end
                slot.label = render.create("hud", self.root)
                slot.preview = render.create("hud", self.root)
                slot.preview_overlay = render.create("hud", self.root)
                slot.preview:set_clip(x, y, screen.grid.cell_width, screen.grid.cell_height)
                slot.preview_overlay:set_clip(x, y, screen.grid.cell_width, screen.grid.cell_height)
            end
            self.slots[#self.slots + 1] = slot
        end

        local function refresh_page()
            local pages = page_count(#self.characters, self.page_size)
            self.page = math.max(1, math.min(pages, self.page))
            local first = (self.page - 1) * self.page_size + 1
            for slot_index, slot in ipairs(self.slots) do
                local character = self.characters[first + slot_index - 1]
                slot.character = character
                slot.control.visible = character ~= nil
                slot.control.enabled = character ~= nil
                if character then
                    slot.control.label = character.name
                end
                if slot.selection then
                    for _, piece in ipairs(slot.selection) do
                        piece:set_visible(character ~= nil and character.id == self.selected_id)
                    end
                    slot.label:set_visible(character ~= nil)
                    slot.preview:set_visible(character ~= nil)
                    slot.preview_overlay:set_visible(false)
                    if character then
                        local flags = {}
                        if character.expansion then
                            flags[#flags + 1] = "Expansion"
                        end
                        if character.hardcore then
                            flags[#flags + 1] = "Hardcore"
                        end
                        local metadata_font = manifest.fonts[screen.metadata_font]
                        local label_width, label_height = slot.label:set_text(
                            metadata_font.table,
                            metadata_font.sheet,
                            manifest.palettes[metadata_font.palette],
                            string.format(
                                "%s\nLevel %d %s\n%s",
                                character.name,
                                character.level,
                                character.class,
                                table.concat(flags, " ")
                            ),
                            { red = 210, green = 180, blue = 110, max_width = 185, align = "left" }
                        )
                        local column = (slot_index - 1) % screen.grid.columns
                        local row = math.floor((slot_index - 1) / screen.grid.columns)
                        slot.label:set_position(
                            screen.grid.x + column * screen.grid.column_step + screen.grid.text_offset.x + label_width / 2,
                            screen.grid.y + row * screen.grid.row_step + screen.grid.text_offset.y + label_height / 2
                        )
                        local presentation = assert(self.class_presentations[character.class])
                        if character.appearance then
                            -- Save importers resolve equipment codes into an
                            -- immutable COF/DCC snapshot. The shim supplies it
                            -- directly to the generic composite renderer.
                            slot.preview:set_cof_animation(
                                character.appearance.cof,
                                character.appearance.palette,
                                character.appearance.direction,
                                character.appearance.components,
                                "loop"
                            )
                        else
                            -- Class-only/legacy saves do not claim equipment
                            -- state and retain the verified front-end fallback.
                            slot.preview:set_dc6_animation(
                                presentation.selected,
                                manifest.palettes[screen.preview_palette],
                                0,
                                presentation.frames_per_second or 15,
                                "loop",
                                "offsets"
                            )
                        end
                        slot.preview:set_scale(screen.grid.preview_scale, screen.grid.preview_scale)
                        slot.preview:set_position(
                            screen.grid.x + column * screen.grid.column_step + screen.grid.preview_offset.x,
                            screen.grid.y + row * screen.grid.row_step + screen.grid.preview_offset.y
                        )
                        if not character.appearance and presentation.selected_overlay then
                            slot.preview_overlay:set_dc6_animation(
                                presentation.selected_overlay,
                                manifest.palettes[screen.preview_palette],
                                0,
                                presentation.frames_per_second or 15,
                                "loop",
                                "offsets"
                            )
                            slot.preview_overlay:set_scale(screen.grid.preview_scale, screen.grid.preview_scale)
                            slot.preview_overlay:set_position(
                                screen.grid.x + column * screen.grid.column_step + screen.grid.preview_offset.x,
                                screen.grid.y + row * screen.grid.row_step + screen.grid.preview_offset.y
                            )
                            slot.preview_overlay:set_visible(true)
                        end
                    end
                end
            end
        end
        self.refresh_page = refresh_page

        self.scrollbar = self.controls:add_scrollbar({
            id = "pages",
            label = "Character pages",
            x = screen.scrollbar.x,
            y = screen.scrollbar.y,
            width = screen.scrollbar.width,
            height = screen.scrollbar.height,
            orientation = "vertical",
            min = 1,
            max = page_count(#self.characters, self.page_size),
            value = 1,
            step = 1,
            on_change = function(_, value)
                self.page = math.floor(value + 0.5)
                refresh_page()
                if self.update_scrollbar_visual then
                    self:update_scrollbar_visual()
                end
            end,
        })
        if render.assets_available() and self.scrollbar.max > 1 then
            local definition = screen.scrollbar
            local palette = manifest.palettes[definition.palette]
            self.scrollbar_visual = {
                up = render.create("hud", self.root),
                down = render.create("hud", self.root),
                thumb = render.create("hud", self.root),
            }
            local up_width, up_height = self.scrollbar_visual.up:set_dc6(
                definition.sheet, palette, 0, definition.up_frame
            )
            local down_width, down_height = self.scrollbar_visual.down:set_dc6(
                definition.sheet, palette, 0, definition.down_frame
            )
            local thumb_width, thumb_height = self.scrollbar_visual.thumb:set_dc6(
                definition.sheet, palette, 0, definition.thumb_frame
            )
            self.scrollbar_visual.up:set_position(definition.x + up_width / 2, definition.y + up_height / 2)
            self.scrollbar_visual.down:set_position(
                definition.x + down_width / 2,
                definition.y + definition.height - down_height / 2
            )
            self.update_scrollbar_visual = function()
                local span = definition.height - up_height - down_height - thumb_height
                local ratio = (self.scrollbar.value - self.scrollbar.min)
                    / math.max(1, self.scrollbar.max - self.scrollbar.min)
                self.scrollbar_visual.thumb:set_position(
                    definition.x + thumb_width / 2,
                    definition.y + up_height + thumb_height / 2 + span * ratio
                )
            end
            self:update_scrollbar_visual()
        end

        local button_actions = {
            new = function()
                scenes.replace("character_create")
            end,
            exit = function()
                scenes.replace("main_menu")
            end,
            ok = launch_selected,
            delete = function()
                local selected
                for _, character in ipairs(self.characters) do
                    if character.id == self.selected_id then
                        selected = character
                        break
                    end
                end
                if not selected then
                    return
                end
                self.modal = dialog.confirm(
                    self.root,
                    screen.delete_dialog,
                    font,
                    manifest.palettes[screen.delete_dialog.palette],
                    manifest.palettes[font.palette],
                    string.format(assert(locale.text("darkmagic.character_select.delete_prompt")), selected.name),
                    assert(locale.text("darkmagic.dialog.ok")),
                    assert(locale.text("darkmagic.dialog.cancel")),
                    function(confirmed)
                        self.modal = nil
                        if not confirmed then
                            return
                        end
                        assert(saves.delete(selected.id))
                        self.characters = saves.characters()
                        if #self.characters == 0 then
                            scenes.replace("character_create")
                            return
                        end
                        self.selected_id = self.characters[1].id
                        self.scrollbar.max = page_count(#self.characters, self.page_size)
                        refresh_page()
                    end
                )
            end,
        }
        for _, id in ipairs({ "new", "delete", "exit", "ok" }) do
            local definition = screen.controls[id]
            ui_button.create(self.root, self.controls, id, definition, assert(locale.text(definition.label)), {
                layer = "hud",
                on_activate = button_actions[id],
            })
        end

        refresh_page()
        update_title(self.characters[1])
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,

    update = function(self, elapsed)
        if not self.cursor then
            return
        end
        self.now = self.now + elapsed
        self.cursor:update()
        if self.modal then
            local modal = self.modal
            modal:update()
            if modal.open and input.pressed("cancel") then
                modal:close(false)
            end
            return
        end
        self.controls:update()
        if input.pressed("cancel") then
            scenes.replace("main_menu")
        end
    end,
}
