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
local ui_button = require("darkmagic.ui.button")
local cursor = require("darkmagic.ui.cursor")
local dialog = require("darkmagic.ui.dialog")
local text = require("darkmagic.ui.text")

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
        if render.assets_available() then
            -- The campfire is independent foreground animation, not part of
            -- the tiled background. It is drawn above idle heroes just as the
            -- original frontend composition authored it.
            self.campfire = render.create("hud", self.root)
            dc6.anchored_animation(
                self.campfire,
                screen.campfire.sheet,
                manifest.palettes[screen.campfire.palette],
                screen.campfire.anchor.x,
                screen.campfire.anchor.y,
                screen.campfire.frames_per_second,
                "loop"
            )
            self.campfire:set_z(10)

            -- Heading and focused-class copy are renderer-owned bitmap text.
            -- Locale keys are derived from stable lowercase class IDs so mods
            -- can replace prose without changing scene behavior.
            self.heading = render.create("hud", self.root)
            self.class_name = render.create("hud", self.root)
            self.class_description = render.create("hud", self.root)
            self.heading:set_z(100)
            self.class_name:set_z(100)
            self.class_description:set_z(100)
            local heading = screen.labels.heading
            text.set(self.heading, heading.style, assert(locale.text(heading.key)), heading.width, "center")
            self.heading:set_position(heading.x, heading.y)
        end
        self.controls = controls.new()
        self.expansion = true
        self.hardcore = false
        self.creation_options = {}
        self.classes = {}
        self.transitions = {}
        -- Each class definition owns its presentation paths, anchor, timing,
        -- narration, and hit rectangle. Adding a class should be a data change.
        for _, definition in ipairs(screen.classes) do
            local placement = assert(screen.stage[definition.class])
            local class = { definition = definition, placement = placement }
            if render.assets_available() then
                class.node = render.create("hud", self.root)
                class.overlay = render.create("hud", self.root)
            end

            -- A state can have a base DC6 plus an optional synchronized overlay.
            -- Verified frame counts keep transition timing deterministic even in
            -- headless tests where renderer assets are intentionally unavailable.
            class.show = function(state)
                class.state = state
                local frames_per_second = definition.frames_per_second or 15
                local frames = definition[state .. "_frames"] or 0
                local loop = (state == "forward" or state == "back") and "once" or "loop"
                if class.node then
                    frames = dc6.anchored_animation(
                        class.node,
                        definition[state],
                        manifest.palettes[screen.class_palette],
                        placement.anchor.x,
                        placement.anchor.y,
                        frames_per_second,
                        loop
                    )
                    class.node:set_visible(true)
                    local overlay_path = definition[state .. "_overlay"]
                    class.overlay:set_visible(overlay_path ~= nil)
                    if overlay_path then
                        dc6.anchored_animation(
                            class.overlay,
                            overlay_path,
                            manifest.palettes[screen.class_palette],
                            placement.anchor.x,
                            placement.anchor.y,
                            frames_per_second,
                            loop
                        )
                    end
                end
                return frames / frames_per_second
            end
            class.open_name_dialog = function()
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
            end
            class.finish_selection = function()
                class.show("selected")
                self.selection_transition = nil
                for _, presentation in ipairs(self.creation_options) do
                    presentation.control.visible = true
                    if presentation.visual then
                        presentation.visual:set_visible(true)
                        presentation.label:set_visible(true)
                    end
                end
                class.open_name_dialog()
            end
            class.show("unselected")
            class.control = self.controls:add({
                id = string.lower(definition.class),
                label = definition.class,
                x = placement.hit.x,
                y = placement.hit.y,
                width = placement.hit.width,
                height = placement.hit.height,
                on_state = function(_, state)
                    if self.class_name and (state == "hover" or state == "focused") then
                        local class_id = string.lower(definition.class)
                        local name_label = screen.labels.class
                        text.set(
                            self.class_name,
                            name_label.style,
                            assert(locale.text("darkmagic.character_class." .. class_id .. ".name")),
                            name_label.width,
                            "center"
                        )
                        self.class_name:set_position(name_label.x, name_label.y)

                        local description = screen.labels.description
                        text.set(
                            self.class_description,
                            description.style,
                            assert(locale.text("darkmagic.character_class." .. class_id .. ".description")),
                            description.width,
                            "center"
                        )
                        self.class_description:set_position(description.x, description.y)
                    end
                    if self.selected ~= class and class.state ~= "back" then
                        class.show(
                            (state == "hover" or state == "focused") and "hover" or "unselected"
                        )
                    end
                end,
                on_activate = function()
                    -- Do not let repeated input replace renderer resources in
                    -- the middle of an authored walk transition.
                    if self.selection_transition or (self.dialog and self.dialog.open) then
                        return
                    end
                    if self.selected == class and class.state == "selected" then
                        class.open_name_dialog()
                        return
                    end
                    if self.selected and self.selected ~= class then
                        local previous = self.selected
                        local deselect = previous.definition.deselect_sound
                        if deselect and audio.exists(deselect) then
                            audio.play(deselect, { bus = "ui" })
                        end
                        local duration = previous.show("back")
                        if duration > 0 then
                            self.transitions[#self.transitions + 1] = {
                                class = previous,
                                remaining = duration,
                            }
                        else
                            previous.show("unselected")
                        end
                    end

                    -- Re-selecting a character cancels any pending return to
                    -- its unselected pose.
                    for index = #self.transitions, 1, -1 do
                        if self.transitions[index].class == class then
                            table.remove(self.transitions, index)
                        end
                    end
                    self.selected = class
                    if definition.select_sound and audio.exists(definition.select_sound) then
                        audio.play(definition.select_sound, { bus = "ui" })
                    end
                    local duration = class.show("forward")
                    if duration > 0 then
                        self.selection_transition = class
                        self.transitions[#self.transitions + 1] = {
                            class = class,
                            remaining = duration,
                            complete = class.finish_selection,
                        }
                    else
                        class.finish_selection()
                    end
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
                local label_width = text.set(
                    label,
                    screen.option_style,
                    assert(locale.text(definition.label)),
                    0,
                    "left"
                )
                label:set_position(
                    definition.x + 28 + label_width / 2,
                    definition.y + definition.height / 2
                )
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
            option.visible = false
            if visual then
                visual:set_dc6(
                    definition.sheet,
                    manifest.palettes[definition.palette],
                    0,
                    option.checked and 1 or 0
                )
                visual:set_position(definition.x + 10, definition.y + definition.height / 2)
                visual:set_visible(false)
                label:set_visible(false)
            end
            self.creation_options[#self.creation_options + 1] = {
                control = option,
                visual = visual,
                label = label,
            }
        end

        local exit_definition = screen.controls.exit
        ui_button.create(self.root, self.controls, "exit", exit_definition, assert(locale.text(exit_definition.label)), {
            layer = "hud",
            on_activate = function()
                scenes.replace("main_menu")
            end,
        })
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,
    update = function(self, elapsed)
        self.cursor:update()

        for index = #self.transitions, 1, -1 do
            local transition = self.transitions[index]
            transition.remaining = transition.remaining - elapsed
            if transition.remaining <= 0 then
                table.remove(self.transitions, index)
                if transition.complete then
                    transition.complete()
                else
                    transition.class.show("unselected")
                end
            end
        end

        if self.selection_transition then
            return
        end
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
