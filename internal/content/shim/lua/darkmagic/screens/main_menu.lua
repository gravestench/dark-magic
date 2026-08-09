-- Primary expansion frontend menu.
--
-- This scene is the most complete example of manifest-driven composition:
-- tiled DC6 art, synchronized animation layers, localized bitmap-font labels,
-- reusable controls, scoped audio, cursor rendering, and scene navigation.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local dc6 = require("darkmagic.ui.dc6")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local text = require("darkmagic.ui.text")
local compat = require("darkmagic.ui.compat")
local app = require("dm.app/v1")
local cursor = require("darkmagic.ui.cursor")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
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
            -- Diablo II draw mode 3 is screen blending (ONE,
            -- ONE_MINUS_SRC_COLOR), not ordinary additive blending.
            self.logo.fire_left:set_blend(compat.draw_mode(3))
            self.logo.fire_right:set_blend(compat.draw_mode(3))
            self:configure_logo()
        end
        self:configure_controls()
        self:configure_labels()
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,

    configure_labels = function(self)
        if not render.assets_available() then
            return
        end
        for id, definition in pairs(screen.labels) do
            local label = render.create("hud", self.root)
            local text_value = assert(locale.text(definition.key))
            if id == "version" then
                text_value = string.format(text_value, app.version())
            end
            text.set(label, definition.style, text_value, definition.width, definition.align)
            label:set_position(definition.x, definition.y)
        end
    end,

    configure_logo = function(self)
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

        local function add_control(id)
            -- The compatibility catalog carries the cross-checked original
            -- 800x600 geometry/frame facts; navigation/localization remain in
            -- the presentation manifest so mods can still replace behavior.
            local definition = compat.screen_control("main_menu", id, assert(screen.controls[id]))
            button.create(self.root, self.controls, id, definition, assert(locale.text(definition.label)), {
                layer = "hud",
                on_activate = function()
                    if definition.action == "exit" then
                        app.request_exit()
                    else
                        scenes.replace(definition.target or "character_select")
                    end
                end,
            })
        end

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
