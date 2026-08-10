-- Expansion character-creation scene.
--
-- This scene composes the original 800x600 frontend from recovered presentation
-- facts while keeping behavior behind Dark Magic's retained renderer and Lua
-- capabilities. Reference engines are used only for observable asset/layout/
-- timing behavior; their widget implementations are not imported.
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
local ui_checkbox = require("darkmagic.ui.checkbox")
local ui_text_field = require("darkmagic.ui.text_field")
local cursor = require("darkmagic.ui.cursor")
local text = require("darkmagic.ui.text")
local compat = require("darkmagic.ui.compat")
local preload = require("darkmagic.ui.preload")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.character_create
local recovered = compat.frontend.character_create
local panel_z = 120

local function position_text_from_top(node, style, value, definition)
    local _, height = text.set(node, style, value, definition.width, "center")
    node:set_position(definition.x, definition.y + height / 2)
end

local function set_node_visible(node, visible)
    if node then node:set_visible(visible) end
end

local function set_form_control_visible(control, visible)
    control.visible = visible
    set_node_visible(control.node, visible)
    set_node_visible(control.background_node, visible)
    set_node_visible(control.value_node, visible)
    set_node_visible(control.label_node, visible)
end

local function set_form_control_z(control, z)
    if control.node then control.node:set_z(z) end
    if control.background_node then control.background_node:set_z(z) end
    if control.value_node then control.value_node:set_z(z + 1) end
    if control.label_node then control.label_node:set_z(z + 1) end
end

local function merged_definition(base, override)
    local result = {}
    for key, value in pairs(base or {}) do result[key] = value end
    for key, value in pairs(override or {}) do result[key] = value end
    return result
end

local function leave_character_creation()
    -- Character select intentionally forwards an empty roster into character
    -- creation. Sending Exit back through that screen creates a two-scene
    -- loop, so an empty roster returns to the main menu instead.
    if #saves.characters() == 0 then
        scenes.replace("main_menu")
    else
        scenes.replace("character_select")
    end
end

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
        self.form_controls = {}
        self.classes = {}
        self.transitions = {}
        self.selection_transition = nil

        if render.assets_available() then
            -- OpenD2 draws the campfire last, with legacy draw mode 3. Its
            -- renderer's frontend animation Y transform maps (-80) to the
            -- normalized Dark Magic anchor y=319.
            self.campfire = render.create("hud", self.root)
            self.campfire:set_z(200)
            self.campfire:set_blend(compat.draw_mode(recovered.campfire.draw_mode))
            dc6.anchored_animation(
                self.campfire,
                screen.campfire.sheet,
                manifest.palettes[screen.campfire.palette],
                recovered.campfire.anchor.x,
                recovered.campfire.anchor.y,
                screen.campfire.frames_per_second,
                "loop"
            )

            self.heading = render.create("hud", self.root)
            self.class_name = render.create("hud", self.root)
            self.class_description = render.create("hud", self.root)
            self.heading:set_z(100)
            self.class_name:set_z(100)
            self.class_description:set_z(100)
            position_text_from_top(
                self.heading,
                screen.labels.heading.style,
                assert(locale.text(screen.labels.heading.key)),
                recovered.heading
            )
        end

        self.show_class_copy = function(definition)
            if not self.class_name then return end
            local class_id = string.lower(definition.class)
            position_text_from_top(
                self.class_name,
                screen.labels.class.style,
                assert(locale.text("darkmagic.character_class." .. class_id .. ".name")),
                recovered.class_name
            )
            position_text_from_top(
                self.class_description,
                screen.labels.description.style,
                assert(locale.text("darkmagic.character_class." .. class_id .. ".description")),
                recovered.description
            )
        end

        -- The original screen uses overlapping animated actors. Their fallback
        -- rectangles are retained for headless execution, while real-asset runs
        -- derive each interaction box from the decoded idle animation bounds.
        for _, definition in ipairs(screen.classes) do
            local fallback = assert(screen.stage[definition.class])
            local stage = assert(recovered.stage[definition.class])
            local placement = {
                anchor = stage.anchor or fallback.anchor,
                z = recovered.draw_order[definition.class] or fallback.z,
                hit = fallback.hit,
                overlay_draw_mode = stage.overlay_draw_mode,
            }
            local class = { definition = definition, placement = placement }

            if render.assets_available() then
                class.node = render.create("hud", self.root)
                class.overlay = render.create("hud", self.root)
                class.node:set_z(placement.z)
                class.overlay:set_z(placement.z + 1)
                if placement.overlay_draw_mode then
                    class.overlay:set_blend(compat.draw_mode(placement.overlay_draw_mode))
                end

                local x1, y1, x2, y2 = render.dc6_animation_bounds(
                    definition.unselected,
                    manifest.palettes[screen.class_palette],
                    0,
                    "offsets"
                )
                if x2 > x1 and y2 > y1 then
                    placement.hit = {
                        x = placement.anchor.x + x1,
                        y = placement.anchor.y + y1,
                        width = x2 - x1,
                        height = y2 - y1,
                    }
                end
            end

            class.show = function(state)
                class.state = state
                local transitioning = state == "forward" or state == "back"
                local frames_per_second = definition[state .. "_frames_per_second"]
                if not frames_per_second then
                    if transitioning then
                        frames_per_second = recovered.transition_frames_per_second
                    elseif state == "selected" then
                        frames_per_second = recovered.idle_front_frames_per_second
                    else
                        frames_per_second = recovered.idle_back_frames_per_second
                    end
                end
                local frames = definition[state .. "_frames"] or 0
                local loop = transitioning and "once" or "loop"

                if class.node then
                    local overlay_path = definition[state .. "_overlay"]
                    local palette = manifest.palettes[screen.class_palette]
                    if overlay_path then
                        class.animation_nodes = { class.node, class.overlay }
                        frames = dc6.anchored_composite(
                            class.animation_nodes,
                            { definition[state], overlay_path },
                            palette,
                            placement.anchor.x,
                            placement.anchor.y,
                            frames_per_second,
                            loop
                        )
                    else
                        class.animation_nodes = { class.node }
                        frames = dc6.anchored_animation(
                            class.node,
                            definition[state],
                            palette,
                            placement.anchor.x,
                            placement.anchor.y,
                            frames_per_second,
                            loop
                        )
                    end
                    class.node:set_visible(true)
                    class.overlay:set_visible(overlay_path ~= nil)
                    class.animation_elapsed = 0
                    dc6.pause_animations(class.animation_nodes)
                    dc6.synchronize_animations(class.animation_nodes, 0)
                end
                return frames_per_second > 0 and frames / frames_per_second or 0
            end

            class.finish_selection = function()
                class.show("selected")
                self.selection_transition = nil
                self.show_class_copy(definition)
                self:set_form_visible(true)
                self:update_ok_state()
            end

            class.show("unselected")
            class.control = self.controls:add({
                id = string.lower(definition.class),
                label = definition.class,
                x = placement.hit.x,
                y = placement.hit.y,
                width = placement.hit.width,
                height = placement.hit.height,
                -- Pointer behavior follows the recovered actor hit regions,
                -- while keyboard/controller focus remains available as a Dark
                -- Magic accessibility/compatibility extension.
                hit_priority = -1000 + placement.z,
                on_state = function(_, state)
                    if state == "hover" or state == "focused" then
                        self.show_class_copy(definition)
                    elseif self.selected then
                        self.show_class_copy(self.selected.definition)
                    end

                    if self.selected ~= class and class.state ~= "back" and class.state ~= "forward" then
                        class.show((state == "hover" or state == "focused") and "hover" or "unselected")
                    end
                end,
                on_activate = function()
                    if self.selection_transition then return end

                    -- Clicking the already-selected hero walks them back and
                    -- hides the dynamic name/options panel, matching OpenD2.
                    if self.selected == class then
                        local deselect = definition.deselect_sound
                        if deselect and audio.exists(deselect) then audio.play(deselect, { bus = "ui" }) end
                        self.selected = nil
                        self:set_form_visible(false)
                        self:update_ok_state()
                        local duration = class.show("back")
                        if duration > 0 then
                            self.selection_transition = class
                            self.transitions[#self.transitions + 1] = {
                                class = class,
                                remaining = duration,
                                complete = function()
                                    class.show("unselected")
                                    self.selection_transition = nil
                                end,
                            }
                        else
                            class.show("unselected")
                        end
                        return
                    end

                    if self.selected then
                        local previous = self.selected
                        local deselect = previous.definition.deselect_sound
                        if deselect and audio.exists(deselect) then audio.play(deselect, { bus = "ui" }) end
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

                    self.selected = class
                    self.show_class_copy(definition)
                    self:set_form_visible(true)
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

        -- Dynamic character form: OpenD2's panel origin is (320,490), with
        -- Expansion at +35 and Hardcore at +55. The panel stays inline rather
        -- than opening a modal character-name dialog.
        local name_definition = merged_definition({
            palette = screen.palette,
            width = 272,
            height = 32,
        }, recovered.form.name)
        self.name_field = ui_text_field.create(
            self.root,
            self.controls,
            "character_name",
            name_definition,
            "CHARACTER NAME",
            {
                layer = "hud",
                text_style = "formal_large",
                label_style = screen.option_style,
                on_change = function() self:update_ok_state() end,
            }
        )
        set_form_control_z(self.name_field, panel_z)
        self.form_controls[#self.form_controls + 1] = self.name_field

        local expansion_definition = merged_definition(screen.options.expansion, {
            x = recovered.form.expansion.x,
            y = recovered.form.expansion.y,
            width = compat.widgets.checkbox.width,
            height = compat.widgets.checkbox.height,
            checked = true,
        })
        self.expansion_control = ui_checkbox.create(
            self.root,
            self.controls,
            "expansion",
            expansion_definition,
            assert(locale.text(screen.options.expansion.label)),
            {
                layer = "hud",
                on_change = function(_, checked) self.expansion = checked end,
            }
        )
        self.expansion_control.focusable = false
        set_form_control_z(self.expansion_control, panel_z)
        self.form_controls[#self.form_controls + 1] = self.expansion_control

        local hardcore_definition = merged_definition(screen.options.hardcore, {
            x = recovered.form.hardcore.x,
            y = recovered.form.hardcore.y,
            width = compat.widgets.checkbox.width,
            height = compat.widgets.checkbox.height,
            checked = false,
        })
        self.hardcore_control = ui_checkbox.create(
            self.root,
            self.controls,
            "hardcore",
            hardcore_definition,
            assert(locale.text(screen.options.hardcore.label)),
            {
                layer = "hud",
                on_change = function(_, checked) self.hardcore = checked end,
            }
        )
        self.hardcore_control.focusable = false
        set_form_control_z(self.hardcore_control, panel_z)
        self.form_controls[#self.form_controls + 1] = self.hardcore_control

        self.set_form_visible = function(current, visible)
            current.form_visible = visible
            for _, control in ipairs(current.form_controls) do
                current.controls:set_visible(control.id, visible)
                set_form_control_visible(control, visible)
            end
            if visible then current.controls:set_focus(current.name_field) end
        end

        -- Static Exit/OK buttons use the character-create-specific selected
        -- button art, not the ordinary frontend medium button.
        local exit_definition = compat.screen_control("character_create", "exit", assert(screen.controls.exit))
        self.exit_button = ui_button.create(
            self.root,
            self.controls,
            "exit",
            exit_definition,
            assert(locale.text(exit_definition.label)),
            {
                layer = "hud",
                z = panel_z,
                on_activate = leave_character_creation,
            }
        )
        self.exit_button.focusable = false

        local ok_fallback = {
            x = 630,
            y = 535,
            width = screen.controls.exit.width,
            height = screen.controls.exit.height,
            label = "darkmagic.dialog.ok",
            sheet = "data/global/ui/FrontEnd/MediumSelButtonBlank.dc6",
            palette = screen.controls.exit.palette,
            up_frames = {0},
            down_frames = {1},
        }
        local ok_definition = compat.screen_control("character_create", "ok", ok_fallback)
        self.ok_button = ui_button.create(
            self.root,
            self.controls,
            "ok",
            ok_definition,
            assert(locale.text(ok_definition.label)),
            {
                layer = "hud",
                z = panel_z,
                enabled = false,
                on_activate = function()
                    if not self.selected then return end
                    local name = self.name_field.value or ""
                    if #name < recovered.form.minimum_name_length then return end
                    local id, err = saves.create_named(
                        name,
                        self.selected.definition.class,
                        self.expansion,
                        self.hardcore
                    )
                    if not id then
                        self.error = err
                        return
                    end
                    assert(saves.select(id))
                    scenes.replace("game_loading")
                end,
            }
        )

        self.update_ok_state = function(current)
            local valid = current.selected ~= nil
                and #(current.name_field.value or "") >= recovered.form.minimum_name_length
            current.controls:set_enabled("ok", valid)
        end

        self:set_form_visible(false)
        self:update_ok_state()
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
        preload.character_create_interactions()
    end,

    update = function(self, elapsed)
        if self.cursor then self.cursor:update() end

        for _, class in ipairs(self.classes) do
            if class.animation_nodes then
                class.animation_elapsed = class.animation_elapsed + elapsed
                dc6.synchronize_animations(class.animation_nodes, class.animation_elapsed)
            end
        end

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

        -- Keep transition resources stable while a hero is walking forward or
        -- back. Pointer/keyboard controls resume as soon as the animation ends.
        if self.selection_transition then return end

        self.controls:update()
        if self.form_visible and input.pressed("confirm") and self.ok_button.enabled
            and self.controls.focus ~= self.ok_button then
            self.controls:activate(self.ok_button)
        end
        if input.pressed("cancel") then
            if self.form_visible and self.selected and self.controls.focus ~= self.selected.control then
                self.controls:set_focus(self.selected.control)
            else
                leave_character_creation()
            end
        end
    end,
}
