-- Minimal interactive game-world orchestration scene.
--
-- Lua owns input and presentation flow while Lua-defined Akara components own
-- deterministic gameplay state. Selected-character appearance and the HUD are
-- disposable presentation handles owned entirely by this scene.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local vfs = require("dm.vfs/v1")
local audio = require("dm.audio/v1")
local saves = require("dm.save/v1")
local data = require("dm.data/v1")
local game_hud = require("darkmagic.ui.game_hud")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.game_world

return {
    create = function(self)
        self.gameplay_world = require("darkmagic.gameplay.world")
        self.root = render.create("world")
        if render.assets_available() and screen.map then
            -- Decode a second, renderer-independent view of the authored map.
            -- This immutable handle is the future source of collision, LOS,
            -- objects, and navigation facts. The render node below remains
            -- disposable presentation and never becomes simulation state.
            local world = require("dm.world/v1")
            self.world = assert(world.load(screen.map.ds1, screen.map.dt1))
            self.world_dimensions = self.world:dimensions()

            self.map = render.create("world", self.root)
            local width, height = self.map:set_ds1(
                screen.map.ds1,
                screen.map.dt1,
                screen.map.palette
            )
            self.map_width, self.map_height = width, height
            self.map:set_z(screen.map.z)
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
        local world_width = self.world_dimensions and self.world_dimensions.width_subtiles or 4096
        local world_height = self.world_dimensions and self.world_dimensions.height_subtiles or 4096
        self.gameplay = self.gameplay_world.create(world_width, world_height, self.world, "local-player")
        local camera_x, camera_y = self.gameplay_world.position(self.gameplay.camera)
        if self.world then
            camera_x, camera_y = self.world:subtile_to_pixel(camera_x, camera_y)
        end
        self.initial_camera_x, self.initial_camera_y = camera_x, camera_y
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
        local hero_x, hero_y = self.gameplay_world.position(self.gameplay.hero)
        local camera_x, camera_y = self.gameplay_world.position(self.gameplay.camera)
        if self.world then
            hero_x, hero_y = self.world:subtile_to_pixel(hero_x, hero_y)
            camera_x, camera_y = self.world:subtile_to_pixel(camera_x, camera_y)
        end
        if self.map then
            self.map:set_position(
                screen.map.screen_x - (camera_x - self.initial_camera_x),
                screen.map.screen_y - (camera_y - self.initial_camera_y)
            )
        end
        self.hero:set_position(
            screen.hero.screen_x + hero_x - camera_x,
            screen.hero.screen_y + hero_y - camera_y
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

    destroy = function(self)
        self.gameplay_world.destroy(self.gameplay)
    end,
}
