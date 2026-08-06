-- Expansion character-creation scene.
--
-- This is a reference for combining manifest-defined character animations,
-- focusable hit regions, localized options, scoped narration, modal text entry,
-- and engine-owned save creation without embedding asset facts in Lua code.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local saves = require("dm.save/v1")
local data = require("dm.data/v1")
local audio = require("dm.audio/v1")
local locale = require("dm.locale/v1")
local dc6 = require("darkmagic.ui.dc6")
local controls = require("darkmagic.ui.controls")
local cursor = require("darkmagic.ui.cursor")
local dialog = require("darkmagic.ui.dialog")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.character_create

return {
    create = function(self)
        self.root = render.create("hud")
        self.background = dc6.frontend_background(
            self.root,
            "hud",
            screen.background,
            manifest.palettes[screen.palette],
            manifest.layouts.frontend_tiles
        )
        self.controls = controls.new()
        self.expansion = true
        self.hardcore = false
        self.classes = {}
        -- Each class definition owns its presentation paths, anchor, timing,
        -- narration, and hit rectangle. Adding a class should be a data change.
        for _, definition in ipairs(screen.classes) do
            local class = { definition = definition }
            if render.assets_available() then
                class.node = render.create("hud", self.root)
                local function show(state)
                    local path = definition[state]
                    dc6.anchored_animation(
                        class.node,
                        path,
                        manifest.palettes[definition.palette],
                        definition.anchor.x,
                        definition.anchor.y,
                        definition.frames_per_second or 15,
                        "loop"
                    )
                end
                class.show = show
                show("unselected")
            end
            class.control = self.controls:add({
                id = string.lower(definition.class),
                label = definition.class,
                x = definition.hit.x,
                y = definition.hit.y,
                width = definition.hit.width,
                height = definition.hit.height,
                on_state = function(_, state)
                    if class.show then
                        class.show(
                            (state == "hover" or state == "focused") and "hover" or "unselected"
                        )
                    end
                end,
                on_activate = function()
                    if self.selected and self.selected ~= class then
                        local deselect = self.selected.definition.deselect_sound
                        if deselect and audio.exists(deselect) then
                            audio.play(deselect, { bus = "ui" })
                        end
                    end
                    self.selected = class
                    if definition.select_sound and audio.exists(definition.select_sound) then
                        audio.play(definition.select_sound, { bus = "ui" })
                    end
                    if class.show then
                        class.show("selected")
                    end

                    -- The native save capability validates and canonicalizes
                    -- identities. Returning false keeps the dialog open.
                    self.dialog = dialog.text_entry(
                        self.root,
                        screen.dialog,
                        manifest.fonts.exocet10,
                        manifest.palettes[screen.dialog.palette],
                        manifest.palettes[manifest.fonts.exocet10.palette],
                        "Character name",
                        "",
                        function(name)
                            local id, err = saves.create_named(
                                name,
                                definition.class,
                                self.expansion,
                                self.hardcore
                            )
                            if not id then
                                self.error = err
                                return false
                            end
                            assert(saves.select(id))
                            scenes.replace("game_loading")
                            return true
                        end
                    )
                end,
            })
            self.classes[#self.classes + 1] = class
        end
        -- Expansion and hardcore are ordinary checkboxes whose values are
        -- passed into the immutable save-creation request.
        for _, id in ipairs({ "expansion", "hardcore" }) do
            local definition = screen.options[id]
            local visual, label
            if render.assets_available() then
                visual = render.create("hud", self.root)
                label = render.create("hud", self.root)
                label:set_text(
                    manifest.fonts.exocet10.table,
                    manifest.fonts.exocet10.sheet,
                    manifest.palettes[manifest.fonts.exocet10.palette],
                    assert(locale.text(definition.label)),
                    {
                        red = 210,
                        green = 180,
                        blue = 110,
                        max_width = definition.width - 28,
                        align = "left",
                    }
                )
                label:set_position(definition.x + 28, definition.y + definition.height / 2)
            end
            local option = self.controls:add_checkbox({
                id = id,
                label = assert(locale.text(definition.label)),
                checked = self[id],
                x = definition.x,
                y = definition.y,
                width = definition.width,
                height = definition.height,
                on_change = function(control, checked)
                    self[id] = checked
                    if visual then
                        visual:set_dc6(
                            definition.sheet,
                            manifest.palettes[definition.palette],
                            0,
                            checked and 1 or 0
                        )
                    end
                end,
            })
            if visual then
                visual:set_dc6(
                    definition.sheet,
                    manifest.palettes[definition.palette],
                    0,
                    option.checked and 1 or 0
                )
                visual:set_position(definition.x + 10, definition.y + definition.height / 2)
            end
        end
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,
    update = function(self)
        self.cursor:update()
        if self.dialog and self.dialog.open then
            self.dialog:update()
            if input.pressed("cancel") then
                self.dialog:close()
            end
            return
        end
        self.controls:update()
        if input.pressed("cancel") then
            scenes.replace("main_menu")
        end
    end,
}
