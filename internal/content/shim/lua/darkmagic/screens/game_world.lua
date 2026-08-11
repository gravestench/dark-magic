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
local tooltip = require("darkmagic.ui.tooltip")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.game_world

local function install_current_world(self)
	local world_capability = self.world_capability
	if not world_capability then
		world_capability = require("dm.world/v1")
		self.world_capability = world_capability
	end
	local level_id = world_capability.current_level()
	if self.world_level_id == level_id then return false end
	local next_world, recipe = world_capability.current()
	assert(next_world and recipe, "session world is unavailable")
	if self.hero and self.hero:exists() then self.hero:destroy() end
	self.collision_node, self.collision_region_key = nil, nil
	self.tile_debug_node, self.tile_debug_region_key = nil, nil
	self.hero_origin = nil
	if self.map then chunked_map.destroy(self.map) end
	self.world, self.world_recipe, self.world_level_id = next_world, recipe, recipe.level_id
	recipe.palette = screen.map.palette
	recipe.world = next_world
	self.world_dimensions = next_world:dimensions()
	self.world_canvas_width, self.world_canvas_height = next_world:canvas()
	self.map = chunked_map.create(self.root, recipe, {
		z = screen.map.z,
		viewport_width = manifest.resolution.width,
		viewport_height = manifest.resolution.height,
		canvas_width = self.world_canvas_width,
		canvas_height = self.world_canvas_height,
	})
	self.hero = render.create("world", self.map.root)
	self.hero:set_visible(false)
	-- The previous render handle belonged to the destroyed map root. Force the
	-- cached appearance to attach to this new handle while preserving the
	-- playback clock, so crossing a seam does not rewind the animation.
	self.hero_composite_key = nil
	if self.gameplay then self.gameplay_world.set_collision(self.gameplay, next_world) end
	return true
end

local function place_region_node(self, node, x, y, width, height)
	node:set_position(x - self.world_canvas_width / 2 + width / 2,
		y - self.world_canvas_height / 2 + height / 2)
end

local function refresh_collision_overlay(self, hero_x, hero_y)
    if not self.collision_visible or not self.world or not self.map then return end
    local cell_x, cell_y = math.floor(hero_x / 5), math.floor(hero_y / 5)
    local key = string.format("%d:%d", cell_x, cell_y)
    if key == self.collision_region_key then return end
    self.collision_region_key = key
    if not self.collision_node then
        self.collision_node = render.create("world", self.map.root)
        self.collision_node:set_z(900000)
    end
    local radius = 30
    local x, y, width, height = self.collision_node:set_world_collision_region(
        self.world, math.floor(hero_x) - radius, math.floor(hero_y) - radius,
        math.floor(hero_x) + radius + 1, math.floor(hero_y) + radius + 1
    )
    place_region_node(self, self.collision_node, x, y, width, height)
end

local function refresh_tile_overlay(self, hero_x, hero_y)
	if not self.tile_debug_visible or not self.world or not self.map then return end
	local cell_x, cell_y = math.floor(hero_x / 5), math.floor(hero_y / 5)
	local key = string.format("%d:%d", cell_x, cell_y)
	if key == self.tile_debug_region_key then return end
	self.tile_debug_region_key = key
	if not self.tile_debug_node then
		self.tile_debug_node = render.create("world", self.map.root)
		self.tile_debug_node:set_z(900001)
	end
	local radius = 30
	local x, y, width, height = self.tile_debug_node:set_world_tile_region(
		self.world, math.floor(hero_x) - radius, math.floor(hero_y) - radius,
		math.floor(hero_x) + radius + 1, math.floor(hero_y) + radius + 1)
	place_region_node(self, self.tile_debug_node, x, y, width, height)
end

local function create_cross(layer, parent, z, r, g, b)
	local horizontal, vertical = render.create(layer, parent), render.create(layer, parent)
	horizontal:fill_rect(17, 3, r, g, b, 255); vertical:fill_rect(3, 17, r, g, b, 255)
	horizontal:set_z(z); vertical:set_z(z)
	return {horizontal=horizontal, vertical=vertical}
end

local function set_cross(marker, visible, x, y)
	marker.horizontal:set_visible(visible); marker.vertical:set_visible(visible)
	if visible then marker.horizontal:set_position(x, y); marker.vertical:set_position(x, y) end
end

local function destroy_cross(marker)
	if not marker then return end
	marker.horizontal:destroy()
	marker.vertical:destroy()
end

local function update_debug_legend(self)
	local active = self.collision_visible or self.tile_debug_visible or self.origins_visible
	if not active and self.debug_legend then
		self.debug_legend:destroy()
		self.debug_legend = nil
		return
	end
	if active and not self.debug_legend then
		self.debug_legend = render.create("hud", self.root)
		self.debug_legend:set_z(1000000)
	end
	if not self.debug_legend then return end
	text.set(self.debug_legend, "font_lab_color", string.format(
		"[gold]WORLD DEBUG  [white]F3 collision %s  F4 tiles %s  F5 origins %s",
		self.collision_visible and "[green]ON" or "[red]off",
		self.tile_debug_visible and "[green]ON" or "[red]off",
		self.origins_visible and "[green]ON" or "[red]off"), 760, "center")
	self.debug_legend:set_position(400, 18)
end

local function selectable_at(self, x, y)
	local spawned = self.targeting and self.targeting.selectable_at(x, y) or nil
	if spawned and spawned.owner ~= "local-player" then return spawned end
	return self.world and self.world:selectable_at(x, y) or nil
end

return {
    create = function(self)
        -- Ordinary Lua helper defining/binding ECS world presentation behavior.
        self.gameplay_world = require("darkmagic.gameplay.world")

        -- Root retained node for world-layer presentation owned by this scene.
        self.root = render.create("world")

        if render.assets_available() then
            -- IMPORTANT: load the map through TWO different views for two
            -- different jobs.
            --
            -- dm.world/v1 produces renderer-INDEPENDENT semantic map facts:
            -- dimensions, collision, subtile projection, future LOS/objects/etc.
			install_current_world(self)
        end

        -- Hero sprite/composite is presentation only. The authoritative hero is
        -- a separate ECS entity admitted by the session and bound later.
        -- Map bands and standing entities share one parent so their baseline
        -- depths can interleave. A parent-level z can never provide occlusion.
		if not self.hero then self.hero = render.create("world", self.map and self.map.root or self.root) end
		if render.assets_available() then self.world_hover_tip = tooltip.create(self.root, "", 0, 0, {origin_x="left",origin_y="top",alpha=190}) end

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
            self.player = player
            self.interaction = require("dm.interaction/v1")
			self.targeting = require("dm.targeting/v1")
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
		if render.assets_available() then install_current_world(self) end
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
		if input.pressed("debug_collision") then
			self.collision_visible = not self.collision_visible
			self.collision_region_key = nil
			if not self.collision_visible and self.collision_node then
				-- Diagnostic textures are disposable. Destroying the whole retained
				-- node avoids leaving a hidden texture/resource transition queued on
				-- the render thread when F3 is released.
				self.collision_node:destroy()
				self.collision_node = nil
			end
		end
		if input.pressed("debug_map_tiles") then
			self.tile_debug_visible = not self.tile_debug_visible
			self.tile_debug_region_key = nil
			if not self.tile_debug_visible and self.tile_debug_node then
				self.tile_debug_node:destroy()
				self.tile_debug_node = nil
			end
		end
		if input.pressed("debug_origins") then
			self.origins_visible = not self.origins_visible
			if not self.origins_visible then
				destroy_cross(self.hero_origin); self.hero_origin = nil
			end
		end
		update_debug_legend(self)
		refresh_collision_overlay(self, hero_x, hero_y)
		refresh_tile_overlay(self, hero_x, hero_y)

        -- Hero normally targets manifest center. Side overlays ask the world to
        -- frame into the unobscured half instead.
        local target_x = screen.hero.screen_x
        local target_y = screen.hero.screen_y
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
		if self.origins_visible and not self.hero_origin and self.map then self.hero_origin = create_cross("world", self.map.root, 900002, 255, 64, 255) end
		if self.hero_origin then set_cross(self.hero_origin, true, hero_screen_x, hero_screen_y) end

        if self.map then
            -- Absolute authoritative camera coordinates determine both chunk
            -- culling and placement. Re-entering the scene cannot accumulate drift.
            local _, map_error, effective_target_x, effective_target_y = chunked_map.update(
                self.map, camera_pixel_x, camera_pixel_y, target_x, target_y, world_view
            )
            if map_error then error("updating world map: " .. tostring(map_error)) end
            target_x = effective_target_x or target_x
            target_y = effective_target_y or target_y
        end

        -- Legacy gameplay is pointer-authored. Reverse-project only the visible
        -- world portion; HUD and obscured overlay halves retain their own clicks.
        local pointer_x, pointer_y = input.cursor()
        local in_world = pointer_y >= 0
            and pointer_y < (screen.world_input_bottom or manifest.resolution.height)
        if world_view == "left" then
            in_world = in_world and pointer_x < manifest.resolution.width / 2
        elseif world_view == "right" then
            in_world = in_world and pointer_x >= manifest.resolution.width / 2
        elseif world_view == "none" then
            in_world = false
        end
		if self.world_hover_tip then
			local hover = nil
			if self.world and in_world and not self.__darkmagic_item_held then
				local hover_x, hover_y = self.world:screen_to_subtile(pointer_x, pointer_y, camera_x, camera_y, target_x, target_y)
				hover = selectable_at(self, hover_x, hover_y)
			end
			self.world_hover_tip:set_visible(hover ~= nil)
			if hover then self.world_hover_tip:set_text(hover.label ~= "" and hover.label or hover.kind);self.world_hover_tip:set_position(pointer_x+16,pointer_y+18) end
		end
        if self.player and self.world and in_world and not self.__darkmagic_item_held
            and (input.pressed("pointer_primary") or (input.down("pointer_primary") and not self.pending_interaction)) then
            local target_world_x, target_world_y = self.world:screen_to_subtile(
                pointer_x, pointer_y, camera_x, camera_y, target_x, target_y
            )
			local selected = selectable_at(self, target_world_x, target_world_y)
			if selected and input.pressed("pointer_primary") then
				if selected.kind == "hostile" then
					self.pending_interaction = nil
					self.player.request_skill("left", selected.x, selected.y, selected.id)
				elseif selected.kind == "static-object" or selected.kind == "dynamic-object" then
					self.pending_interaction = selected
					self.player.request_move(selected.x, selected.y, 3.5)
				else
					self.pending_interaction = nil
					self.player.request_move(selected.x, selected.y, selected.radius or 0)
				end
			else
				self.pending_interaction = nil
				self.player.request_move(target_world_x, target_world_y)
			end
		end
		if self.pending_interaction then
			local selected = self.pending_interaction
			local dx, dy = hero_x - selected.x, hero_y - selected.y
			if dx * dx + dy * dy <= 16 and self.world:line_clear(hero_x, hero_y, selected.x, selected.y) then
				self.interaction.open_at(selected.x, selected.y)
				self.player.request_move(hero_x, hero_y)
				self.pending_interaction = nil
			elseif not self.player.movement_pending() then
				-- Authority could not find a route. Do not leave a ghost interaction
				-- that unexpectedly fires after some unrelated later movement.
				self.pending_interaction = nil
			end
        end
		if self.player and self.world and in_world and not self.__darkmagic_item_held
			and input.pressed("pointer_secondary") then
			local skill_x, skill_y = self.world:screen_to_subtile(
				pointer_x, pointer_y, camera_x, camera_y, target_x, target_y
			)
			local selected = selectable_at(self, skill_x, skill_y)
			self.player.request_skill("right", skill_x, skill_y, selected and selected.id or "")
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
