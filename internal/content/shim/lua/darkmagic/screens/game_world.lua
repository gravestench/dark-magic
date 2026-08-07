-- Minimal interactive game-world orchestration scene.
--
-- Lua owns input and presentation flow while dm.simulation/v1 owns persistent,
-- deterministic gameplay state. Selected-character appearance and the HUD are
-- disposable presentation handles owned entirely by this scene.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local vfs = require("dm.vfs/v1")
local audio = require("dm.audio/v1")
local simulation = require("dm.simulation/v1")
local saves = require("dm.save/v1")
local data = require("dm.data/v1")
local game_hud = require("darkmagic.ui.game_hud")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.game_world

return {
    create = function(self)
        self.root = render.create("world")
        if render.assets_available() and screen.map then
            self.map = render.create("world", self.root)
            local width, height = self.map:set_ds1(
                screen.map.ds1,
                screen.map.dt1,
                screen.map.palette
            )
            self.map_width, self.map_height = width, height
            self.map:set_z(screen.map.z)
            local state = simulation.state()
            self.initial_camera_x, self.initial_camera_y = state.camera_x, state.camera_y
            self.map:set_position(screen.map.screen_x, screen.map.screen_y)
        end
        self.hero = render.create("world", self.root)
        local character = saves.selected()
        if character and character.appearance and render.assets_available() then
            self.hero:set_cof_animation(
                character.appearance.cof,
                character.appearance.palette,
                character.appearance.direction,
                character.appearance.components,
                "loop"
            )
            self.hero:set_scale(screen.hero.scale, screen.hero.scale)
        else
            self.hero:set_visible(false)
        end
        if render.assets_available() then
            self.hud = game_hud.create(self.root, screen.hud, manifest.palettes)
        end

        -- VFS provenance and optional asset checks are examples of querying
        -- capabilities without receiving direct filesystem/native ownership.
        self.content_source = assert(vfs.source("boot.lua"))
        self.town_music_available = audio.exists("data/global/music/Act1/town1.wav")
    end,

    update = function(self, elapsed, focused)
        -- Transparent overlays may allow updates below them, but only the top
        -- scene receives input focus.
        if not focused then
            return
        end
        if self.hud then
            game_hud.update(self.hud)
        end
        local speed = 160 * elapsed
        local dx, dy = 0, 0
        if input.down("left") then
            dx = dx - speed
        end
        if input.down("right") then
            dx = dx + speed
        end
        if input.down("up") then
            dy = dy - speed
        end
        if input.down("down") then
            dy = dy + speed
        end
        if dx ~= 0 or dy ~= 0 then
            simulation.move_hero(dx, dy)
        end

        local state = simulation.state()
        if self.map then
            self.map:set_position(
                screen.map.screen_x - (state.camera_x - self.initial_camera_x),
                screen.map.screen_y - (state.camera_y - self.initial_camera_y)
            )
        end
        self.hero:set_position(
            screen.hero.screen_x + state.hero_x - state.camera_x,
            screen.hero.screen_y + state.hero_y - state.camera_y
        )

        -- Panels are scene overlays rather than long-lived engine services.
        if input.pressed("inventory") then
            scenes.push("inventory")
        elseif input.pressed("character") then
            scenes.push("character")
        elseif input.pressed("skills") then
            scenes.push("skills")
        elseif input.pressed("automap") then
            scenes.push("automap")
        elseif input.pressed("options") then
            scenes.push("options")
        elseif input.pressed("pause") or input.pressed("cancel") then
            scenes.push("pause")
        end
    end,
}
