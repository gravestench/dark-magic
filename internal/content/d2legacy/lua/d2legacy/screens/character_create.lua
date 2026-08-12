-- Expansion character-creation scene.
--
-- This is one of the most sophisticated Lua scenes in d2legacy, but it is still
-- built from the same small ideas you have already seen:
--
--   * retained render nodes for pictures;
--   * controls.Manager for interaction;
--   * manifest/compat data for recovered layout facts;
--   * save capability for actual character creation/selection;
--   * a small Lua state machine for actor hover/selection animations.
--
-- The seven heroes are NOT seven special native UI objects. Each class is a Lua
-- table containing its definition, retained nodes, control, current state, and a
-- few small functions. That is the thread to follow through this file.
--
-- Reference engines are used only for observable asset/layout/timing behavior;
-- their widget implementations are not imported.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local saves = require("engine.save/v1")
local data = require("engine.data/v1")
local audio = require("engine.audio/v1")
local locale = require("engine.locale/v1")
local dc6 = require("d2legacy.ui.dc6")
local controls = require("d2legacy.ui.controls")
local ui_button = require("d2legacy.ui.button")
local ui_checkbox = require("d2legacy.ui.checkbox")
local ui_text_field = require("d2legacy.ui.text_field")
local cursor = require("d2legacy.ui.cursor")
local text = require("d2legacy.ui.text")
local compat = require("d2legacy.ui.compat")
local preload = require("d2legacy.ui.preload")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local screen = manifest.screens.character_create
local recovered = compat.frontend.character_create

-- The inline name/options panel must sit above the animated class actors but
-- below a few intentionally higher presentation pieces.
local panel_z = 120

-- Recovered frontend text coordinates describe the TOP of the text area, while
-- retained text nodes are positioned by CENTER. Measure first, then add half the
-- rendered height to get the correct center Y.
local function position_text_from_top(node, style, value, definition)
    local _, height = text.set(node, style, value, definition.width, "center")
    node:set_position(definition.x, definition.y + height / 2)
end

-- Many composite widgets have optional child render nodes. This tiny helper
-- makes `nil` safe so callers can apply visibility without four separate guards.
local function set_node_visible(node, visible)
    if node then node:set_visible(visible) end
end

-- A text field/checkbox is partly a CONTROL and partly several render nodes.
-- Hiding the form therefore has to hide both interaction and every visual piece.
local function set_form_control_visible(control, visible)
    control.visible = visible
    set_node_visible(control.node, visible)
    set_node_visible(control.background_node, visible)
    set_node_visible(control.value_node, visible)
    set_node_visible(control.label_node, visible)
end

-- Same idea for z-order: composite controls may expose several optional nodes.
local function set_form_control_z(control, z)
    if control.node then control.node:set_z(z) end
    if control.background_node then control.background_node:set_z(z) end
    if control.value_node then control.value_node:set_z(z + 1) end
    if control.label_node then control.label_node:set_z(z + 1) end
end

-- Make a shallow copy of `base`, then let `override` replace individual fields.
-- This is a common Lua way to combine defaults with profile-specific facts
-- WITHOUT mutating either source table shared by other modules.
local function merged_definition(base, override)
    local result = {}
    for key, value in pairs(base or {}) do result[key] = value end
    for key, value in pairs(override or {}) do result[key] = value end
    return result
end

local function leave_character_creation()
    -- Character select intentionally forwards an empty roster into character
    -- creation. Sending Exit back through that screen creates a two-scene loop,
    -- so an empty roster returns to the main menu instead.
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

        -- Scene-local mutable presentation/input state begins here.
        self.controls = controls.new()
        self.expansion = true
        self.hardcore = false
        self.form_controls = {}
        self.classes = {}
        self.transitions = {}
        self.selection_transition = nil

        if render.assets_available() then
            -- The campfire is independent looping artwork behind/in front of the
            -- class actors depending on z. Legacy draw mode 3 maps through compat
            -- rather than hard-coding renderer blend equations here.
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

            -- These three retained text nodes are reused as hover/selection changes.
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

        -- Store this helper on the scene because many class callbacks below need
        -- to replace the same two text nodes with the hovered/selected class copy.
        self.show_class_copy = function(definition)
            if not self.class_name then return end

            local class_id = string.lower(definition.class)
            position_text_from_top(
                self.class_name,
                screen.labels.class.style,
                assert(locale.text("d2legacy.character_class." .. class_id .. ".name")),
                recovered.class_name
            )
            position_text_from_top(
                self.class_description,
                screen.labels.description.style,
                assert(locale.text("d2legacy.character_class." .. class_id .. ".description")),
                recovered.description
            )
        end

        -- CLASS ACTORS -------------------------------------------------------
        --
        -- The original screen uses seven overlapping animated actors. One loop
        -- constructs seven small Lua `class` objects from manifest data.
        for _, definition in ipairs(screen.classes) do
            local fallback = assert(screen.stage[definition.class])
            local stage = assert(recovered.stage[definition.class])

            local placement = {
                -- Prefer cross-checked recovered anchor/z facts, retain manifest
                -- fallback hit rectangles for headless/no-asset execution.
                anchor = stage.anchor or fallback.anchor,
                z = recovered.draw_order[definition.class] or fallback.z,
                hit = fallback.hit,
                overlay_draw_mode = stage.overlay_draw_mode,
            }

            -- This table represents ONE class actor's presentation state.
            local class = { definition = definition, placement = placement }

            if render.assets_available() then
                class.node = render.create("hud", self.root)
                class.overlay = render.create("hud", self.root)
                class.node:set_z(placement.z)
                class.overlay:set_z(placement.z + 1)

                if placement.overlay_draw_mode then
                    class.overlay:set_blend(compat.draw_mode(placement.overlay_draw_mode))
                end

                -- Real decoded idle bounds are better hit geometry than guessed
                -- manifest rectangles. Convert animation-local bounds into this
                -- actor's shared screen anchor coordinate space.
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

            -- Switch this actor to one logical animation state:
            -- unselected, hover, forward, selected, or back.
            class.show = function(state)
                class.state = state

                local transitioning = state == "forward" or state == "back"

                -- Bracket lookup with a computed key lets one function read
                -- `hover_frames_per_second`, `selected_frames_per_second`, etc.
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
                        -- Some class states are two independently cropped layers.
                        -- anchored_composite puts both in one common anchor canvas.
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

                    -- Pause managed playback and reset a scene-driven shared clock
                    -- for every layer participating in this class state.
                    class.animation_elapsed = 0
                    dc6.pause_animations(class.animation_nodes)
                    dc6.synchronize_animations(class.animation_nodes, 0)
                end

                -- Transition scheduler below wants DURATION in seconds. Frame
                -- count / FPS gives that duration; zero FPS safely gives zero.
                return frames_per_second > 0 and frames / frames_per_second or 0
            end

            class.finish_selection = function()
                class.show("selected")
                self.selection_transition = nil
                self.show_class_copy(definition)
                self:set_form_visible(true)
                self:update_ok_state()
            end

            -- Every actor starts in the back/unselected idle pose.
            class.show("unselected")

            class.control = self.controls:add({
                id = string.lower(definition.class),
                label = definition.class,
                x = placement.hit.x,
                y = placement.hit.y,
                width = placement.hit.width,
                height = placement.hit.height,

                -- Actors overlap. Lower numeric base plus authored z means the
                -- visually frontmost actor wins pointer overlap resolution.
                hit_priority = -1000 + placement.z,

                on_state = function(_, state)
                    -- Hover/focus updates descriptive copy even before activation.
                    if state == "hover" or state == "focused" then
                        self.show_class_copy(definition)
                    elseif self.selected then
                        self.show_class_copy(self.selected.definition)
                    end

                    -- Never interrupt one-shot forward/back transition with a hover
                    -- idle change. Otherwise nonselected actors can freely swap
                    -- between hover and unselected idle presentation.
                    if self.selected ~= class and class.state ~= "back" and class.state ~= "forward" then
                        class.show((state == "hover" or state == "focused") and "hover" or "unselected")
                    end
                end,

                on_activate = function()
                    -- While one class is walking forward/back, freeze selection
                    -- input until the one-shot transition completes.
                    if self.selection_transition then return end

                    -- Clicking the ALREADY-selected hero deselects them: play
                    -- backward transition, hide form, and eventually return idle.
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

                    -- Selecting a DIFFERENT class may require the previous hero
                    -- to walk backward while the new one walks forward.
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

        -- INLINE CHARACTER FORM ---------------------------------------------
        -- Original panel shows name + Expansion + Hardcore after class selected.
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
                -- OK button eligibility changes whenever name changes.
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
        -- Original pointer checkbox should not become extra arrow-key focus stop.
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

        -- Show/hide the WHOLE inline form: manager interaction + every visual node.
        self.set_form_visible = function(current, visible)
            current.form_visible = visible
            for _, control in ipairs(current.form_controls) do
                current.controls:set_visible(control.id, visible)
                set_form_control_visible(control, visible)
            end
            if visible then current.controls:set_focus(current.name_field) end
        end

        -- STATIC EXIT / OK ---------------------------------------------------
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

        -- Manifest may not yet carry every recovered OK art fact, so supply a
        -- fallback definition and let compat.screen_control override recovered fields.
        local ok_fallback = {
            x = 630,
            y = 535,
            width = screen.controls.exit.width,
            height = screen.controls.exit.height,
            label = "d2legacy.dialog.ok",
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
                    -- Defensive validation repeats the visible enabled-state rule.
                    if not self.selected then return end

                    local name = self.name_field.value or ""
                    if #name < recovered.form.minimum_name_length then return end

                    -- ACTUAL save creation is performed by engine.save/v1. Lua passes
                    -- plain requested values and gets back opaque ID or error.
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

                    -- Select newly created save by opaque ID, then transition into loading.
                    assert(saves.select(id))
                    scenes.replace("game_loading")
                end,
            }
        )

        self.update_ok_state = function(current)
            -- OK is eligible only when class is selected AND name meets minimum.
            local valid = current.selected ~= nil
                and #(current.name_field.value or "") >= recovered.form.minimum_name_length
            current.controls:set_enabled("ok", valid)
        end

        -- Initial scene has no selected class, so form is hidden and OK disabled.
        self:set_form_visible(false)
        self:update_ok_state()
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)

        -- Interaction states may have been prepared during startup; this helper
        -- reuses that job or schedules them if necessary.
        preload.character_create_interactions()
    end,

    update = function(self, elapsed)
        if self.cursor then self.cursor:update() end

        -- Advance every class's SHARED layer clock from scene elapsed time.
        for _, class in ipairs(self.classes) do
            if class.animation_nodes then
                class.animation_elapsed = class.animation_elapsed + elapsed
                dc6.synchronize_animations(class.animation_nodes, class.animation_elapsed)
            end
        end

        -- Transition timers are removed while iterating. Walk BACKWARD so
        -- `table.remove` cannot shift an unvisited later item into a skipped index.
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

        -- Keep transition resources stable while a hero is walking forward/back.
        -- Pointer/keyboard controls resume as soon as selection_transition clears.
        if self.selection_transition then return end

        self.controls:update()

        -- Enter/Confirm from name field acts like OK when form is valid, without
        -- requiring keyboard focus to travel to the mostly pointer-authored OK art.
        if self.form_visible and input.pressed("confirm") and self.ok_button.enabled
            and self.controls.focus ~= self.ok_button then
            self.controls:activate(self.ok_button)
        end

        if input.pressed("cancel") then
            if self.form_visible and self.selected and self.controls.focus ~= self.selected.control then
                -- First Cancel from inside form returns focus to selected hero.
                self.controls:set_focus(self.selected.control)
            else
                -- Next Cancel (or Cancel with no selection) leaves scene safely.
                leave_character_creation()
            end
        end
    end,
}
