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

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")
local dc6 = require("d2legacy.ui.dc6")
local cursor = require("d2legacy.ui.cursor")
local audio = require("engine.audio/v1")
local frontend_logo = require("d2legacy.ui.frontend_logo")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
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

        self.logo = frontend_logo.create(self.root, screen.logo, manifest.palettes, "hud")

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
        frontend_logo.update(self.logo, elapsed)

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
