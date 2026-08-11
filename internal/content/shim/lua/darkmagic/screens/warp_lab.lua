-- Warp Lab proves portal travel without inventing a second world renderer.
--
-- The two authored DS1 stamps below use the same sparse map adapter as the
-- game world. That matters: a portal test is not useful if its tiles, actor,
-- camera, culling, or depth ordering behave differently from real gameplay.

local input = require("dm.input/v1")
local render = require("dm.render/v1")
local text = require("darkmagic.ui.text")
local vfs = require("dm.vfs/v1")
local composite = require("darkmagic.gameplay.player_composite")
local chunked_map = require("darkmagic.presentation.chunked_map")

local lab = {}
local palette = "data/global/palette/ACT1/pal.pl2"
local character_palette = "data/global/palette/ACT1/pal.dat"
local warp_fade_seconds = 0.10

local function label(root, value, y, style)
    local node = render.create("hud", root)
    local _, height = text.set(node, style or "font_lab_caption", value, 760, "center")
    node:set_position(400, y + height / 2)
    node:set_z(1000000)
    return node
end

local function choose_stamps()
    local assets = vfs.list("data/global/tiles", ".ds1") or {}
    local candidates = {}
    for _, path in ipairs(assets) do
        if path:lower():match("/act1/caves/cave1%.ds1$") then candidates[#candidates + 1] = path end
    end
    for _, path in ipairs(assets) do
        if path:lower():match("/act1/") then candidates[#candidates + 1] = path end
    end
    local selected, first_directory = {}, nil
    for _, path in ipairs(candidates) do
        local ok, tiles = pcall(render.ds1_dependencies, path)
        if ok and #tiles > 0 then
            local directory = path:lower():match("^(.+)/[^/]+$")
            if #selected == 0 or directory ~= first_directory then
                selected[#selected + 1] = {path = path, tiles = tiles}
                first_directory = first_directory or directory
                if #selected == 2 then return selected end
            end
        end
    end
    return nil
end

local function nearest_open(map, width, height, wanted_x, wanted_y)
    local function footprint_open(x, y)
        for offset_y = -1, 1 do
            for offset_x = -1, 1 do
                if math.sqrt(offset_x * offset_x + offset_y * offset_y) <= 1.5
                    and map:blocked_position(x + offset_x, y + offset_y) then return false end
            end
        end
        return true
    end
    for radius = 0, math.max(width, height) do
        for y = math.max(1, wanted_y - radius), math.min(height - 2, wanted_y + radius) do
            for x = math.max(1, wanted_x - radius), math.min(width - 2, wanted_x + radius) do
                if footprint_open(x + 0.5, y + 0.5) then return x + 0.5, y + 0.5 end
            end
        end
    end
    return width / 2, height / 2
end

local function portal_picture(parent, token, caption)
    local root = render.create("world", parent)
    local animation = render.create("world", root)
    local prefix, lower = "data/global/objects/" .. token, token:lower()
    animation:set_cof_animation(prefix .. "/COF/" .. token .. "ONHTH.cof",
        "data/global/palette/units/pal.dat", 0, {
            HD = prefix .. "/HD/" .. lower .. "hdlitonhth.dcc",
            TR = prefix .. "/TR/" .. lower .. "trlitonhth.dcc",
        }, "loop", 256)
    animation:set_blend("screen")
    local name = render.create("world", root)
    text.set(name, "font_lab_caption", caption, 180, "center")
    name:set_position(0, -88)
    return root
end

local function local_position(stamp, global_x, global_y)
    return global_x - stamp.origin_x, global_y
end

local function canvas_position(stamp, global_x, global_y)
    local x, y = local_position(stamp, global_x, global_y)
    local px, py = stamp.map:subtile_to_pixel(x, y)
    return px - stamp.canvas_width / 2, py - stamp.canvas_height / 2, x, y
end

local function active_stamp(self, global_x)
    local first, second = self.stamps[1], self.stamps[2]
    if math.abs(global_x - first.origin_x) <= math.abs(global_x - second.origin_x) then return first end
    return second
end

local function planned_route(self, global_x, global_y, stop_radius)
    local position = self.ecs.get(self.fixture.player, "dm.world.position")
    local start_x, start_y = local_position(self.active, position:get("x"), position:get("y"))
    local goal_x, goal_y = local_position(self.active, global_x, global_y)
    local path, path_error = self.active.map:find_path(
        start_x, start_y, goal_x, goal_y, 1, stop_radius or 0)
    if not path then
        self.ecs.get(self.fixture.player, "dm.lab.warp.state"):set(
            "event", "route rejected: " .. tostring(path_error))
        return nil
    end
    local encoded = {}
    for _, point in ipairs(path) do
        encoded[#encoded + 1] = string.format("%.0f,%.0f", self.active.origin_x + point.x, point.y)
    end
    return table.concat(encoded, ";")
end

local function make_stamp_view(self, stamp)
    stamp.view = chunked_map.create(self.world_root, {
        world = stamp.map,
        palette = palette,
    }, {
        viewport_width = 800,
        viewport_height = 600,
        canvas_width = stamp.canvas_width,
        canvas_height = stamp.canvas_height,
    })
    stamp.actor = render.create("world", stamp.view.root)
    stamp.actor:set_scale(1.35, 1.35)
    stamp.actor:set_visible(false)
end

function lab:start_world()
    local world = require("dm.world/v1")
    for _, stamp in ipairs(self.stamps) do
        stamp.map = world.load(stamp.path, stamp.tiles)
        stamp.canvas_width, stamp.canvas_height = stamp.map:canvas()
        stamp.dimensions = stamp.map:dimensions()
    end

    self.stamps[1].origin_x = 0
    self.stamp_gap = self.stamps[1].dimensions.width_subtiles + 80
    self.stamps[2].origin_x = self.stamp_gap
    for _, stamp in ipairs(self.stamps) do make_stamp_view(self, stamp) end

    local west, east = self.stamps[1], self.stamps[2]
    local portal_x, portal_y = nearest_open(west.map, west.dimensions.width_subtiles,
        west.dimensions.height_subtiles, math.floor(west.dimensions.width_subtiles / 2),
        math.floor(west.dimensions.height_subtiles / 2))
    local east_x, east_y = nearest_open(east.map, east.dimensions.width_subtiles,
        east.dimensions.height_subtiles, math.floor(east.dimensions.width_subtiles / 2),
        math.floor(east.dimensions.height_subtiles / 2))
    local spawn_x, spawn_y = nearest_open(west.map, west.dimensions.width_subtiles,
        west.dimensions.height_subtiles, math.max(1, math.floor(portal_x - 10)), math.floor(portal_y + 4))

    self.fixture = self.fixture_module.create({x=portal_x, y=portal_y},
        {x=east.origin_x + east_x, y=east_y}, {x=spawn_x, y=spawn_y})
    west.portal = portal_picture(west.view.root, "TP", "BLUE TOWN PORTAL")
    east.portal = portal_picture(east.view.root, "PP", "RED TOWN PORTAL")
    self.ready = true
    text.set(self.status, "font_lab_color",
        "[green]READY[/]  [white]click ground to move; click a portal to interact[/]", 760, "center")
end

function lab:update_views(player_x, player_y)
    local active = active_stamp(self, player_x)
    for _, stamp in ipairs(self.stamps) do
        stamp.view.root:set_visible(stamp == active)
        if stamp == active then
            local local_x, local_y = local_position(stamp, player_x, player_y)
            local camera_x, camera_y = stamp.map:subtile_to_pixel(local_x, local_y)
            local _, map_error, target_x, target_y = chunked_map.update(
                stamp.view, camera_x, camera_y, 400, 300, "center")
            if map_error then error("updating Warp Lab map: " .. tostring(map_error)) end
            stamp.target_x, stamp.target_y = target_x or 400, target_y or 300
        end
    end
    self.active = active
    return active
end

function lab:update_actor(mode, elapsed, stamp)
    local position = self.ecs.get(self.fixture.player, "dm.world.position")
    local actor = self.ecs.get(self.fixture.player, "dm.lab.warp.actor")
    local x, y = position:get("x"), position:get("y")
    local recipe = composite.recipe({token="AM", mode=mode, weapon_class="HTH",
        direction=actor:get("direction"), palette=character_palette})
    if not self.playback or self.playback.mode ~= recipe.mode then
        self.playback = composite.new_playback(recipe)
    end
    composite.advance(self.playback, recipe, elapsed or 0)

    if recipe.key ~= self.actor_key and not self.actor_job then
        self.actor_job = render.preload({composite.preload_request(recipe)})
        self.pending_recipe = recipe
    end
    if self.actor_job then
        local status = render.preload_status(self.actor_job)
        if status and status.done then
            local pending = self.pending_recipe
            render.preload_release(self.actor_job)
            self.actor_job, self.pending_recipe = nil, nil
            if status.failed == 0 and pending.key == recipe.key then
                for _, candidate in ipairs(self.stamps) do
                    candidate.actor:set_cof_animation(pending.cof, pending.palette, pending.direction,
                        pending.components, "loop", pending.rate, self.playback.seconds)
                end
                self.actor_key = pending.key
            end
        end
    end

    local px, py, local_x, local_y = canvas_position(stamp, x, y)
    stamp.actor:set_position(px, py)
    stamp.actor:set_z(stamp.map:entity_depth(local_x, local_y))
    stamp.actor:set_visible(self.actor_key ~= nil)
    return x, y
end

function lab:update_portals()
    local portals = {
        {stamp=self.stamps[1], entity=self.fixture.portal_a, id="warp-lab:a"},
        {stamp=self.stamps[2], entity=self.fixture.portal_b, id="warp-lab:b"},
    }
    local activated, pointer_x, pointer_y = false, input.cursor()
    for _, item in ipairs(portals) do
        local position = self.ecs.get(item.entity, "dm.world.position")
        local x, y = position:get("x"), position:get("y")
        local px, py, local_x, local_y = canvas_position(item.stamp, x, y)
        item.stamp.portal:set_position(px, py)
        item.stamp.portal:set_z(item.stamp.map:entity_depth(local_x, local_y))
        if item.stamp == self.active then
            local player = self.ecs.get(self.fixture.player, "dm.world.position")
            local player_x, player_y = local_position(item.stamp, player:get("x"), player:get("y"))
            local sx, sy = item.stamp.map:subtile_to_screen(local_x, local_y,
                player_x, player_y, item.stamp.target_x, item.stamp.target_y)
            if input.pressed("pointer_primary") and (pointer_x-sx)^2+(pointer_y-sy)^2 <= 42^2 then
                -- Route to the actual portal cell. FindPath's stop radius is
                -- measured between quantized collision cells, while portal
                -- activation below uses continuous authored coordinates. Using
                -- the interaction radius in both domains can leave the actor
                -- one fractional step outside the real activation circle.
                local route = planned_route(self, x, y, 0)
                if route then self.fixture_module.intent(self.fixture, item.id, route) end
                activated = true
            end
        end
    end
    return activated
end

function lab:update_pointer_movement()
    local pointer_x, pointer_y = input.cursor()
    if pointer_y < 85 or pointer_y > 550 then return end
    local player = self.ecs.get(self.fixture.player, "dm.world.position")
    local player_x, player_y = local_position(self.active, player:get("x"), player:get("y"))
    local target_x, target_y = self.active.map:screen_to_subtile(pointer_x, pointer_y,
        player_x, player_y, self.active.target_x, self.active.target_y)
    local global_x = self.active.origin_x + target_x
    local route = planned_route(self, global_x, target_y, 0)
    if route then self.fixture_module.move(self.fixture, global_x, target_y, route) end
end

local function update_warp_fade(self, warp_count, elapsed)
    -- Authority moves first. Presentation notices the completed transaction and
    -- briefly covers the camera handoff, just as the legacy client does. This
    -- timer never delays, changes, or participates in the authoritative warp.
    if warp_count ~= self.observed_warp_count then
        self.observed_warp_count = warp_count
        self.warp_fade_remaining = warp_fade_seconds
    end
    local remaining = math.max(0, self.warp_fade_remaining or 0)
    local alpha = math.floor(255 * math.min(1, remaining / warp_fade_seconds))
    self.warp_fade:fill_rect(1, 1, 0, 0, 0, alpha)
    self.warp_fade:set_visible(alpha > 0)
    self.warp_fade_remaining = math.max(0, remaining - (elapsed or 0))
end

function lab:create()
    self.ecs = require("dm.ecs/v1")
    self.fixture_module = require("darkmagic.dev.warp_lab.fixture")
    self.root = render.create("hud")
    self.world_root = render.create("world", self.root)
    self.title = label(self.root, "WARP LAB", 12, "font_lab_heading")
    self.status = label(self.root, "[gold]LOADING ACT I STAMPS...[/]", 52, "font_lab_color")
    self.detail = label(self.root, "", 566, "font_lab_color")
    -- A one-pixel retained image scales to the viewport, so changing fade alpha
    -- does not rebuild a full 800x600 CPU bitmap every frame.
    self.warp_fade = render.create("hud", self.root)
    self.warp_fade:fill_rect(1, 1, 0, 0, 0, 0)
    self.warp_fade:set_scale(800, 600)
    self.warp_fade:set_position(400, 300)
    self.warp_fade:set_z(2000000)
    self.warp_fade:set_visible(false)
    self.observed_warp_count = 0
    self.stamps = choose_stamps()
    if not self.stamps then error("Warp Lab needs two resolvable Act I DS1 assets") end
    self.job = render.preload({
        {kind="cof_animation", path="data/global/objects/TP/COF/TPONHTH.cof",
            palette="data/global/palette/units/pal.dat", direction=0, components={
                HD="data/global/objects/TP/HD/tphdlitonhth.dcc",
                TR="data/global/objects/TP/TR/tptrlitonhth.dcc"}},
        {kind="cof_animation", path="data/global/objects/PP/COF/PPONHTH.cof",
            palette="data/global/palette/units/pal.dat", direction=0, components={
                HD="data/global/objects/PP/HD/pphdlitonhth.dcc",
                TR="data/global/objects/PP/TR/pptrlitonhth.dcc"}},
    })
end

function lab:update(elapsed)
    if not self.ready then
        local status = render.preload_status(self.job)
        if not status or not status.done then return end
        render.preload_release(self.job); self.job = nil
        if status.failed > 0 then error(tostring(status.errors[1] or "Warp Lab preload failed")) end
        self:start_world()
    end

    local position = self.ecs.get(self.fixture.player, "dm.world.position")
    local x, y = position:get("x"), position:get("y")
    local stamp = self:update_views(x, y)
    local moving = self.ecs.get(self.fixture.player, "dm.lab.warp.intent") ~= nil
        or self.ecs.get(self.fixture.player, "dm.lab.warp.move_intent") ~= nil
    x, y = self:update_actor(moving and "WL" or "NU", elapsed, stamp)
    local portal_activated = self:update_portals()
    if input.pressed("pointer_primary") and not portal_activated then self:update_pointer_movement() end
    local state = self.ecs.get(self.fixture.player, "dm.lab.warp.state")
    update_warp_fade(self, state:get("warp_count"), elapsed)
    text.set(self.detail, "font_lab_color", string.format(
        "[white]%s[/]   [gold]position %.1f, %.1f[/]   [blue]warps %d[/]",
        state:get("event"), x, y, state:get("warp_count")), 760, "center")
end

function lab:destroy()
    if self.job then render.preload_release(self.job); self.job = nil end
    if self.actor_job then render.preload_release(self.actor_job); self.actor_job = nil end
    for _, stamp in ipairs(self.stamps or {}) do
        if stamp.view then chunked_map.destroy(stamp.view) end
    end
    if self.fixture_module then self.fixture_module.destroy(self.fixture) end
end

return lab
