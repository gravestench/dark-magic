-- Expansion trademark/title scene shown after startup cinematics.
--
-- This is a good first FULL rendering scene to read. It demonstrates:
--   * a tiled frontend background;
--   * a four-node animated logo composite;
--   * legacy blend-mode compatibility;
--   * one shared animation clock;
--   * a software cursor;
--   * persistent frontend music;
--   * root-scene navigation.
--
-- Notice that paths, palettes, anchors, timing, and sound names are DATA in the
-- manifest. Lua mainly says how those facts are composed and when state changes.

local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local dc6 = require("darkmagic.ui.dc6")
local cursor = require("darkmagic.ui.cursor")
local compat = require("darkmagic.ui.compat")
local audio = require("dm.audio/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.title

return {
    create = function(self)
        -- This is an ordinary interactive HUD scene, not a transition-only layer.
        self.root = render.create("hud")

        -- Reuse the helper that converts one multi-frame frontend DC6 into its
        -- grid of retained background nodes.
        self.background = dc6.frontend_background(
            self.root,
            "hud",
            screen.background,
            manifest.palettes[screen.palette],
            manifest.layouts.frontend_tiles
        )

        if render.assets_available() then
            -- The logo is FOUR independent retained animations: black + flame for
            -- left half, black + flame for right half.
            self.logo = {
                black_left = render.create("hud", self.root),
                black_right = render.create("hud", self.root),
                fire_left = render.create("hud", self.root),
                fire_right = render.create("hud", self.root),
            }

            local logo = screen.logo
            local palette = manifest.palettes[logo.palette]

            -- Old D2 draw mode 3 maps to Dark Magic's screen-style blend. Only
            -- the flame layers use it; black/logo body remains ordinary opaque art.
            self.logo.fire_left:set_blend(compat.draw_mode(3))
            self.logo.fire_right:set_blend(compat.draw_mode(3))

            -- Each pair is normalized into ONE common anchor-space union. That is
            -- why independently cropped DC6 frames do not jump relative to each other.
            dc6.anchored_composite(
                { self.logo.black_left, self.logo.fire_left },
                { logo.black_left, logo.fire_left },
                palette,
                logo.anchor.x,
                logo.anchor.y,
                logo.frames_per_second,
                "loop"
            )

            dc6.anchored_composite(
                { self.logo.black_right, self.logo.fire_right },
                { logo.black_right, logo.fire_right },
                palette,
                logo.anchor.x,
                logo.anchor.y,
                logo.frames_per_second,
                "loop"
            )

            -- Instead of allowing four managed animations to tick independently,
            -- pause them and let this scene drive ONE elapsed-time clock.
            self.logo_elapsed = 0
            dc6.pause_animations(self.logo)
            dc6.synchronize_animations(self.logo, 0)
        end

        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)

        if audio.exists(manifest.sounds.title_music) then
            -- Persistent music deliberately survives scene replacement from title
            -- to main menu. The named group lets game loading stop it later.
            audio.play_persistent(manifest.sounds.title_music, {
                bus = "music",
                loop = true,
                stream = true,
                group = "frontend_music",
            })
        end
    end,

    update = function(self, elapsed)
        if self.logo then
            self.logo_elapsed = self.logo_elapsed + elapsed
            dc6.synchronize_animations(self.logo, self.logo_elapsed)
        end

        self.cursor:update()

        if input.pressed("skip") then
            if audio.exists(manifest.sounds.select) then
                audio.play(manifest.sounds.select)
            end

            -- Root navigation replaces title instead of stacking main menu above it.
            scenes.replace("main_menu")
        end
    end,
}
