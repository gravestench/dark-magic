-- Primary expansion frontend menu.
--
-- This is the first "everything comes together" frontend example:
--   background helper + animated composite + localization + controls + buttons
--   + app capability + cursor + scene navigation + asset preloading.
--
-- The important pattern is COMPOSITION. main_menu.lua is not a giant GUI system;
-- it mostly wires several small, independently understandable helpers together.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")
local locale = require("engine.locale/v1")
local dc6 = require("d2legacy.ui.dc6")
local controls = require("d2legacy.ui.controls")
local button = require("d2legacy.ui.button")
local text = require("d2legacy.ui.text")
local compat = require("d2legacy.ui.compat")
local app = require("engine.app/v1")
local cursor = require("d2legacy.ui.cursor")
local preload = require("d2legacy.ui.preload")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local screen = manifest.screens.main_menu
local logo = screen.logo
local units_palette = manifest.palettes[logo.palette]

return {
    enter = function(self)
        self.root = render.create("hud")

        self.background = dc6.frontend_background(
            self.root,
            "hud",
            screen.background,
            manifest.palettes[screen.palette],
            manifest.layouts.frontend_tiles
        )

        if render.assets_available() then
            self.logo = {
                black_left = render.create("hud", self.root),
                black_right = render.create("hud", self.root),
                fire_left = render.create("hud", self.root),
                fire_right = render.create("hud", self.root),
            }

            -- Only flame layers use recovered D2 draw mode 3 (screen blend).
            self.logo.fire_left:set_blend(compat.draw_mode(3))
            self.logo.fire_right:set_blend(compat.draw_mode(3))

            -- Method syntax calls the function stored in this same scene table and
            -- automatically passes `self` as its first argument.
            self:configure_logo()
        end

        self:configure_controls()
        self:configure_labels()
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)

        -- This is intentionally AFTER visible menu construction. If startup did
        -- not already complete the immutable preload bundle, player "think time"
        -- on this screen becomes useful background preparation time.
		-- Build this scene first. Once it is visible, the menu becomes useful
		-- think-time for warming the destinations behind Single Player.
		preload.frontend()
    end,

    configure_labels = function(self)
        if not render.assets_available() then
            return
        end

        -- Label definitions are data-driven, so adding/changing copy does not
        -- require a hard-coded sequence of render calls here.
        for id, manifest_definition in pairs(screen.labels) do
            local definition = manifest_definition

            if id == "legal" then
                -- Legal/disclaimer placement is a recovered compatibility fact.
                -- Copy only presentation geometry/style from compat while keeping
                -- the manifest's localization key.
                local recovered = compat.frontend.main_menu.disclaimer
                definition = {
                    x = recovered.x,
                    y = recovered.y,
                    width = recovered.width,
                    align = recovered.align,
                    style = recovered.style,
                    key = manifest_definition.key,
                }
            end

            local label = render.create("hud", self.root)
            local text_value = assert(locale.text(definition.key))

            if id == "version" then
                -- Localized version string contains a format placeholder. App
                -- capability supplies the engine/application version value.
                text_value = string.format(text_value, app.version())
            end

            text.set(label, definition.style, text_value, definition.width, definition.align)
            label:set_position(definition.x, definition.y)
        end
    end,

    configure_logo = function(self)
        -- Same four-layer/shared-clock idea as title.lua.
        dc6.anchored_composite(
            { self.logo.black_left, self.logo.fire_left },
            { logo.black_left, logo.fire_left },
            units_palette,
            logo.anchor.x,
            logo.anchor.y,
            logo.frames_per_second,
            "loop"
        )

        dc6.anchored_composite(
            { self.logo.black_right, self.logo.fire_right },
            { logo.black_right, logo.fire_right },
            units_palette,
            logo.anchor.x,
            logo.anchor.y,
            logo.frames_per_second,
            "loop"
        )

        self.logo_elapsed = 0
        dc6.pause_animations(self.logo)
        dc6.synchronize_animations(self.logo, 0)
    end,

    configure_controls = function(self)
        self.controls = controls.new()

        -- Local helper closes over this scene's root/manager so each button row
        -- only needs to supply its semantic ID.
        local function add_control(id)
            -- Compatibility catalog overrides cross-checked original 800x600
            -- geometry/frame facts WITHOUT mutating the manifest table itself.
            local definition = compat.screen_control("main_menu", id, assert(screen.controls[id]))

            button.create(self.root, self.controls, id, definition, assert(locale.text(definition.label)), {
                layer = "hud",
                on_activate = function()
                    if definition.action == "exit" then
                        -- UI requests application exit through a capability; Lua
                        -- does not call OS window APIs directly.
                        app.request_exit()
                    else
                        -- Root menu destinations REPLACE the current root scene.
                        scenes.replace(definition.target or "character_select")
                    end
                end,
            })
        end

        -- Array order is meaningful because controls.Manager focus traversal uses
        -- insertion order when no custom navigation graph is supplied.
        for _, id in ipairs({ "single_player", "multiplayer", "credits", "cinematics", "exit" }) do
            add_control(id)
        end
    end,

    update = function(self, elapsed)
        if self.logo then
            self.logo_elapsed = self.logo_elapsed + elapsed
            dc6.synchronize_animations(self.logo, self.logo_elapsed)
        end

        if self.controls then
            self.controls:update()
        end

        if self.cursor then
            self.cursor:update()
        end

        if input.pressed("cancel") then
            scenes.replace("title")
        end
    end,
}
