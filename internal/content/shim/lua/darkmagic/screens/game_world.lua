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
        self.character_stats = character and character.stats or nil
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
            local player = require("dm.player/v1")
            self.game_data = require("dm.game_data/v1")
            self.hud = game_hud.create(self.root, screen.hud, manifest.palettes, {
                request_running = player.request_running,
            })
        end

        -- VFS provenance and optional asset checks are examples of querying
        -- capabilities without receiving direct filesystem/native ownership.
        self.content_source = assert(vfs.source("boot.lua"))
        self.town_music_available = audio.exists("data/global/music/Act1/town1.wav")
        local world_width = self.world_dimensions and self.world_dimensions.width_subtiles or 4096
        local world_height = self.world_dimensions and self.world_dimensions.height_subtiles or 4096
        self.gameplay = self.gameplay_world.create(world_width, world_height, self.world, "local-player")
        self.initial_camera_x, self.initial_camera_y = nil, nil
    end,

    update = function(self, elapsed, focused, input_allowed, world_view)
        -- Transparent overlays may allow updates below them, but only the top
        -- scene receives input focus.
        if not input_allowed then
            return
        end
        -- Panels remain available while the fixed-step session is admitting
        -- the selected character; presentation binding must not block UI.
        if input.pressed("inventory") then
            scenes.push("inventory")
        elseif input.pressed("character") then
            scenes.push("character")
        elseif input.pressed("skills") then
            scenes.push("skills")
        elseif input.pressed("automap") then
            scenes.push("automap")
        elseif input.pressed("help") then
            scenes.push("help")
        elseif input.pressed("quests") then
            scenes.push("quests")
        elseif input.pressed("party") then
            scenes.push("party")
        elseif input.pressed("options") then
            scenes.push("options")
        elseif input.pressed("pause") or input.pressed("cancel") then
            scenes.push("pause")
        end
        if not self.gameplay_world.bind(self.gameplay) then
            return
        end
        if self.hud then
            local snapshot = self.gameplay_world.hud_snapshot(self.gameplay.hero, self.character_stats)
            if self.left_skill_id ~= snapshot.left_skill then
                self.left_skill_id = snapshot.left_skill
                self.left_skill = self.game_data.skill(snapshot.left_skill)
            end
            if self.right_skill_id ~= snapshot.right_skill then
                self.right_skill_id = snapshot.right_skill
                self.right_skill = self.game_data.skill(snapshot.right_skill)
            end
            snapshot.left_skill_detail = self.left_skill
            snapshot.right_skill_detail = self.right_skill
            game_hud.update(
                self.hud,
                snapshot
            )
        end
        local hero_x, hero_y = self.gameplay_world.position(self.gameplay.hero)
        local camera_x, camera_y = self.gameplay_world.position(self.gameplay.camera)
        if self.world then
            hero_x, hero_y = self.world:subtile_to_pixel(hero_x, hero_y)
            camera_x, camera_y = self.world:subtile_to_pixel(camera_x, camera_y)
        end
        if not self.initial_camera_x then
            self.initial_camera_x, self.initial_camera_y = camera_x, camera_y
        end
        local target_x = screen.hero.screen_x
        if world_view == "left" then
            target_x = manifest.resolution.width / 4
        elseif world_view == "right" then
            target_x = manifest.resolution.width * 3 / 4
        end
        local view_offset_x = target_x - screen.hero.screen_x
        if self.map then
            self.map:set_position(
                screen.map.screen_x + view_offset_x - (camera_x - self.initial_camera_x),
                screen.map.screen_y - (camera_y - self.initial_camera_y)
            )
        end
        self.hero:set_position(
            target_x + hero_x - camera_x,
            screen.hero.screen_y + hero_y - camera_y
        )
    end,

    destroy = function(self)
        self.gameplay_world.destroy(self.gameplay)
    end,
}
