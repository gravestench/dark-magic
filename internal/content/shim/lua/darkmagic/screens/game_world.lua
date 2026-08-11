-- Minimal interactive GAME WORLD orchestration scene.
--
-- This is a capstone file: it does not try to implement every subsystem itself.
-- It CONNECTS capabilities and smaller Lua helpers:
--
--   dm.world/v1          -> immutable map/collision/projection facts
--   darkmagic.gameplay.world -> ECS binding + camera helper
--   dm.player/v1         -> authoritative player intents
--   dm.items/v1          -> item snapshots/intents
--   dm.game_data/v1      -> immutable skill metadata
--   game_hud.lua         -> disposable HUD presentation
--   dm.scene/v1          -> overlay navigation
--
-- The key lesson is orchestration, not ownership. This scene owns its render
-- handles and presentation-side helper state. It does NOT own the authoritative
-- player entity, item containers, save system, map decoder, or input device.

local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local vfs = require("dm.vfs/v1")
local audio = require("dm.audio/v1")
local saves = require("dm.save/v1")
local data = require("dm.data/v1")
local game_hud = require("darkmagic.ui.game_hud")
local player_composite = require("darkmagic.gameplay.player_composite")
local chunked_map = require("darkmagic.presentation.chunked_map")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.game_world

return {
    create = function(self)
        -- Ordinary Lua helper defining/binding ECS world presentation behavior.
        self.gameplay_world = require("darkmagic.gameplay.world")

        -- Root retained node for world-layer presentation owned by this scene.
        self.root = render.create("world")

        if render.assets_available() and screen.map then
            -- IMPORTANT: load the map through TWO different views for two
            -- different jobs.
            --
            -- dm.world/v1 produces renderer-INDEPENDENT semantic map facts:
            -- dimensions, collision, subtile projection, future LOS/objects/etc.
            local world = require("dm.world/v1")
            self.world = assert(world.load(screen.map.ds1, screen.map.dt1))
            self.world_dimensions = self.world:dimensions()
            self.world_canvas_width, self.world_canvas_height = self.world:canvas()

            -- Presentation receives the same recipe, but uploads only chunks near
            -- the camera. It remains a disposable picture of authoritative facts.
            self.map = chunked_map.create(self.root, screen.map, {
                z = screen.map.z,
                viewport_width = manifest.resolution.width,
                viewport_height = manifest.resolution.height,
                canvas_width = self.world_canvas_width,
                canvas_height = self.world_canvas_height,
            })
        end

        -- Hero sprite/composite is presentation only. The authoritative hero is
        -- a separate ECS entity admitted by the session and bound later.
        -- Map bands and standing entities share one parent so their baseline
        -- depths can interleave. A parent-level z can never provide occlusion.
        self.hero = render.create("world", self.map and self.map.root or self.root)

        local character = saves.selected()

        -- Save metadata still supplies a few HUD fields whose live simulation
        -- schema does not yet own them; gameplay.world carefully prefers live ECS
        -- values where they DO exist.
        self.character_stats = character and character.stats or nil

        -- Keep a stable but initially hidden node. Once the session admits the
        -- player, live ECS appearance selects the composite below. Save metadata
        -- is deliberately not a second source of runtime mode or direction.
        self.hero:set_visible(false)

        if render.assets_available() then
            -- Require gameplay-facing capabilities only in the asset-backed HUD
            -- path that actually needs them.
            local player = require("dm.player/v1")
            local items = require("dm.items/v1")
            self.items = items
            self.game_data = require("dm.game_data/v1")

            -- Instead of giving game_hud raw capabilities, pass the small
            -- operations it needs. This is dependency injection with plain Lua tables.
            self.hud = game_hud.create(self.root, screen.hud, manifest.palettes, {
                request_running = player.request_running,
                assign_skill = player.assign_skill,
                item_snapshot = items.snapshot,
                move_item = items.move,
            })
        end

        -- These are capability QUERIES: Lua gets provenance/availability facts,
        -- not a host filesystem object or native audio device.
        self.content_source = assert(vfs.source("boot.lua"))
        self.town_music_available = audio.exists("data/global/music/Act1/town1.wav")

        -- Use semantic map dimensions when loaded; otherwise generous fallback
        -- keeps headless/tests able to instantiate movement state.
        local world_width = self.world_dimensions and self.world_dimensions.width_subtiles or 4096
        local world_height = self.world_dimensions and self.world_dimensions.height_subtiles or 4096

        -- This creates world helper state and attempts to bind the session-owned player.
        self.gameplay = self.gameplay_world.create(world_width, world_height, self.world, "local-player")

    end,

    update = function(self, elapsed, focused, input_allowed, world_view)
        -- Scene system separates UPDATE from INPUT OWNERSHIP. A transparent panel
        -- may keep the world updating while routing only certain input below.
        --
        -- Held item art must hide when this world does not currently receive
        -- pointer input, otherwise an old item cursor could remain visible under a modal.
        if self.hud then game_hud.set_item_cursor_visible(self.hud, input_allowed) end

        if not input_allowed then
            return
        end

        -- Overlay routing happens even before the session player has finished
        -- binding. UI should remain responsive while a fixed tick admits player state.
        if input.pressed("inventory") then
            scenes.toggle_overlay("inventory", "right")
        elseif input.pressed("character") then
            scenes.toggle_overlay("character", "left")
        elseif input.pressed("skills") then
            scenes.toggle_overlay("skills", "right")
        elseif input.pressed("automap") then
            scenes.toggle_overlay("automap", "full")
        elseif input.pressed("help") then
            scenes.toggle_overlay("help", "full")
        elseif input.pressed("quests") then
            scenes.toggle_overlay("quests", "left")
        elseif input.pressed("party") then
            scenes.toggle_overlay("party", "full")
        elseif input.pressed("options") then
            scenes.toggle_overlay("options", "full")
        elseif input.pressed("pause") or input.pressed("cancel") then
            scenes.toggle_overlay("pause", "full")
        end


        -- W submits a fixed-tick selection command. Equipment does not move
        -- between fake presentation containers: authority merely changes which
        -- pair of hand slots is active, replayable, and visible.
        if self.items and input.pressed("swap_weapons") then
            local item_snapshot = assert(self.items.snapshot())
            self.items.select_weapon_set(item_snapshot.active_weapon_set == 0 and 1 or 0)
        end

        -- Authoritative session may not have admitted the player yet. Bind tries
        -- to find that entity and returns false instead of creating a fake duplicate.
        if not self.gameplay_world.bind(self.gameplay) then
            return
        end

        if render.assets_available() then
            local authority = self.gameplay_world.composite_snapshot(self.gameplay.hero)
            local item_snapshot = self.items and self.items.snapshot() or nil
            local composite = player_composite.resolve(authority, item_snapshot)
            if not self.hero_playback or self.hero_playback.mode ~= composite.mode then
                self.hero_playback = player_composite.new_playback(composite)
            end
            self.hero_animation_events = player_composite.advance(self.hero_playback, composite, elapsed)
            if composite.key ~= self.hero_composite_key and not self.hero_pending_job then
                -- Never decode a cold multi-layer character during this frame.
                -- Keep the previous complete character visible while workers
                -- prepare the newest authoritative appearance/facing request.
                self.hero_pending_job = render.preload({player_composite.preload_request(composite)})
                self.hero_pending_key = composite.key
                self.hero_pending_composite = composite
            end
            if self.hero_pending_job then
                local status = render.preload_status(self.hero_pending_job)
                if status.done then
                    local pending = self.hero_pending_composite
                    if status.failed == 0 and pending.key == composite.key then
                        self.hero:set_cof_animation(
                            pending.cof,
                            pending.palette,
                            pending.direction,
                            pending.components,
                            "loop",
                            pending.rate,
                            self.hero_playback.seconds
                        )
                        self.hero:set_scale(screen.hero.scale, screen.hero.scale)
                        self.hero:set_visible(true)
                        self.hero_composite_key = pending.key
                    end
                    self.hero_pending_job = nil
                    self.hero_pending_key = nil
                    self.hero_pending_composite = nil
                end
            end
        end

        if self.hud then
            -- Build one value-only presentation snapshot from live ECS state plus
            -- the limited saved metadata fields described above.
            local snapshot = self.gameplay_world.hud_snapshot(self.gameplay.hero, self.character_stats)

            -- Skill metadata is immutable. Cache detail records until assigned ID changes.
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

            -- Learned-skill snapshot carries authority fields (ID/level/allowed
            -- sides). Merge immutable display metadata into those COPIED tables so
            -- the selector can render names/icons without mutating live ECS components.
            for _, skill in ipairs(snapshot.learned_skills) do
                local detail = self.game_data.skill(skill.skill_id)
                if detail then
                    for key, value in pairs(detail) do skill[key] = value end
                    skill.level = skill.level or 1
                end
            end

            game_hud.update(
                self.hud,
                snapshot
            )

            -- Cursor decorator reads this scene field and hides the software hand
            -- when authoritative held item art is already acting as pointer.
            self.__darkmagic_item_held = self.hud.item_held
        end

        -- ECS positions are semantic SUBTILES until this exact presentation boundary.
        local hero_x, hero_y = self.gameplay_world.position(self.gameplay.hero)
        local camera_x, camera_y = self.gameplay_world.position(self.gameplay.camera)

        -- Hero normally targets manifest center. Side overlays ask the world to
        -- frame into the unobscured half instead.
        local target_x = screen.hero.screen_x
        if world_view == "left" then
            target_x = manifest.resolution.width / 4
        elseif world_view == "right" then
            target_x = manifest.resolution.width * 3 / 4
        end

        local hero_screen_x, hero_screen_y = target_x, screen.hero.screen_y
        local camera_pixel_x, camera_pixel_y = camera_x, camera_y
        if self.world then
            -- The hero is a sibling of map bands. Its local point is therefore
            -- map-canvas space; their shared parent applies the camera once.
            hero_screen_x, hero_screen_y = self.world:subtile_to_pixel(hero_x, hero_y)
            hero_screen_x = hero_screen_x - self.world_canvas_width / 2
            hero_screen_y = hero_screen_y - self.world_canvas_height / 2
            camera_pixel_x, camera_pixel_y = self.world:subtile_to_pixel(camera_x, camera_y)
            self.hero:set_z(self.world:entity_depth(hero_x, hero_y))
        end

        if self.map then
            -- Absolute authoritative camera coordinates determine both chunk
            -- culling and placement. Re-entering the scene cannot accumulate drift.
            chunked_map.update(
                self.map, camera_pixel_x, camera_pixel_y, target_x, screen.hero.screen_y, world_view
            )
        end

        -- Hero screen position is target anchor plus hero-to-camera relative offset.
        self.hero:set_position(hero_screen_x, hero_screen_y)
    end,

    destroy = function(self)
        -- gameplay.world owns only its presentation camera entity; its destroy
        -- helper intentionally leaves session-owned hero authority alone.
        self.gameplay_world.destroy(self.gameplay)
        chunked_map.destroy(self.map)
    end,
}
