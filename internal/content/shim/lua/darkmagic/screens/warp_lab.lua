-- Warp Lab proves the complete portal interaction loop with ordinary runtime
-- capabilities. Two copies of one real DS1 stamp are placed farther apart than
-- the viewport. Clicking the visible portal publishes an ECS intent; a fixed-
-- tick system walks the actor into range and teleports authoritative position.

local input = require("dm.input/v1")
local render = require("dm.render/v1")
local text = require("darkmagic.ui.text")
local vfs = require("dm.vfs/v1")
local composite = require("darkmagic.gameplay.player_composite")

local lab = {}
local palette = "data/global/palette/ACT1/pal.pl2"
local character_palette = "data/global/palette/ACT1/pal.dat"

local function label(root, value, y, style)
    local node = render.create("hud", root)
    local _, height = text.set(node, style or "font_lab_caption", value, 760, "center")
    node:set_position(400, y + height / 2)
    node:set_z(1000000)
    return node
end

local function choose_stamp()
    -- MPQ lookup is case-insensitive, but directory enumeration preserves the
    -- archive's spelling. Enumerate from the stable root and filter ourselves.
    local assets = vfs.list("data/global/tiles", ".ds1") or {}
    local candidates = {}
    for _, path in ipairs(assets) do
        local normalized = path:lower()
        if normalized:match("/act1/caves/cave1%.ds1$") then candidates[#candidates + 1] = path end
    end
    -- A mounted archive with unusual casing may not match the preferred name.
    -- The lab still remains argument-free and deterministic: use the first
    -- available Act I stamp instead of accepting composition-root flags.
    for _, path in ipairs(assets) do
        if path:lower():match("/act1/") then candidates[#candidates + 1] = path end
    end
    -- Some original DS1 headers contain the author's absolute development
    -- paths. The dependency resolver intentionally rejects those stale names.
    -- Probe deterministic Act I candidates until one declares mounted DT1s.
    for _, path in ipairs(candidates) do
        local ok, tiles = pcall(render.ds1_dependencies, path)
        if ok and #tiles > 0 then return path, tiles end
    end
    return nil, nil
end

local function nearest_open(map, width, height, wanted_x, wanted_y)
    for radius = 0, math.max(width, height) do
        for y = math.max(1, wanted_y - radius), math.min(height - 2, wanted_y + radius) do
            for x = math.max(1, wanted_x - radius), math.min(width - 2, wanted_x + radius) do
                if not map:blocked(x, y) then return x + 0.5, y + 0.5 end
            end
        end
    end
    return width / 2, height / 2
end

local function portal_picture(parent, color, caption)
    local root = render.create("world", parent)
    local glow = render.create("world", root)
    glow:fill_rect(46, 76, color[1], color[2], color[3], 72)
    glow:set_position(0, -35)
    local core = render.create("world", root)
    core:fill_rect(22, 58, color[1], color[2], color[3], 180)
    core:set_position(0, -29)
    local name = render.create("world", root)
    text.set(name, "font_lab_caption", caption, 180, "center")
    name:set_position(0, -88)
    return root
end

local function build_stamp(self, parent)
    for _, chunk in ipairs(self.chunks.chunks) do
        local node = render.create("world", parent)
        node:set_ds1_chunk(self.path, self.tiles, palette, chunk.index)
        node:set_position(chunk.x + chunk.width / 2 - self.canvas_width / 2,
            chunk.y + chunk.height / 2 - self.canvas_height / 2)
        node:set_z(chunk.depth)
    end
end

local function project(self, x, y)
    local px, py = self.map:subtile_to_pixel(x, y)
    return px - self.canvas_width / 2, py - self.canvas_height / 2
end

function lab:start_world()
    local world = require("dm.world/v1")
    self.map = world.load(self.path, self.tiles)
    self.chunks = render.ds1_chunks(self.path, self.tiles, palette)
    self.canvas_width, self.canvas_height = self.map:canvas()
    local dimensions = self.map:dimensions()

    -- A whole unused viewport separates the two authored stamps. Their local
    -- coordinates remain identical; only this explicit world-space placement
    -- distinguishes the western and eastern copies.
    self.stamp_gap = dimensions.width_subtiles + 80
    local portal_x, portal_y = nearest_open(self.map, dimensions.width_subtiles,
        dimensions.height_subtiles, math.floor(dimensions.width_subtiles / 2),
        math.floor(dimensions.height_subtiles / 2))
    local spawn_x, spawn_y = nearest_open(self.map, dimensions.width_subtiles,
        dimensions.height_subtiles, math.max(1, math.floor(portal_x - 10)), math.floor(portal_y + 4))

    self.west = render.create("world", self.world_root)
    self.east = render.create("world", self.world_root)
    build_stamp(self, self.west); build_stamp(self, self.east)
    local origin_x, origin_y = project(self, 0, 0)
    local east_x, east_y = project(self, self.stamp_gap, 0)
    self.east:set_position(east_x - origin_x, east_y - origin_y)

    self.fixture = self.fixture_module.create(
        {x=portal_x, y=portal_y},
        {x=portal_x + self.stamp_gap, y=portal_y},
        {x=spawn_x, y=spawn_y})

    self.portal_a = portal_picture(self.world_root, {45, 115, 255}, "WESTERN PORTAL")
    self.portal_b = portal_picture(self.world_root, {130, 65, 255}, "EASTERN PORTAL")
    self.actor = render.create("world", self.world_root)
    self.actor:set_scale(1.35, 1.35)
    self.actor:set_visible(false)
    self.ready = true
    text.set(self.status, "font_lab_color", "[green]READY[/]  [white]click the portal to publish an interaction intent[/]", 760, "center")
end

function lab:update_actor(mode, elapsed)
    local position = self.ecs.get(self.fixture.player, "dm.world.position")
    local x, y = position:get("x"), position:get("y")
    local recipe = composite.recipe({token="AM", mode=mode, weapon_class="HTH",
        direction=12, palette=character_palette})
    self.playback = self.playback or composite.new_playback(recipe)
    composite.advance(self.playback, recipe, elapsed or 0)
    if recipe.key ~= self.actor_key then
        self.actor:set_cof_animation(recipe.cof, recipe.palette, recipe.direction,
            recipe.components, "loop", recipe.rate, self.playback.seconds)
        self.actor:set_visible(true)
        self.actor_key = recipe.key
    end
    local px, py = project(self, x, y)
    self.actor:set_position(px, py)
    self.actor:set_z(self.map:entity_depth(x, y))
    self.world_root:set_position(400 - px, 300 - py)
    return x, y
end

function lab:update_portals()
    local portals = {
        {node=self.portal_a, entity=self.fixture.portal_a, id="warp-lab:a"},
        {node=self.portal_b, entity=self.fixture.portal_b, id="warp-lab:b"},
    }
    for _, item in ipairs(portals) do
        local position = self.ecs.get(item.entity, "dm.world.position")
        local x, y = position:get("x"), position:get("y")
        local px, py = project(self, x, y)
        item.node:set_position(px, py)
        item.node:set_z(self.map:entity_depth(x, y))
        local player = self.ecs.get(self.fixture.player,"dm.world.position")
        local sx, sy = self.map:subtile_to_screen(x, y,
            player:get("x"), player:get("y"), 400, 300)
        local cx, cy = input.cursor()
        if input.pressed("pointer_primary") and (cx-sx)^2+(cy-sy)^2 <= 42^2 then
            self.fixture_module.intent(self.fixture, item.id)
        end
    end
end

function lab:create()
    -- Keep capability lookup inside scene activation. Boot/catalog tests can
    -- register every scene without installing the gameplay ECS capability.
    self.ecs = require("dm.ecs/v1")
    self.fixture_module = require("darkmagic.dev.warp_lab.fixture")
    self.root = render.create("hud")
    self.world_root = render.create("world", self.root)
    self.title = label(self.root, "WARP LAB", 12, "font_lab_heading")
    self.status = label(self.root, "[gold]LOADING ACT I STAMP...[/]", 52, "font_lab_color")
    self.detail = label(self.root, "", 566, "font_lab_color")
    self.path, self.tiles = choose_stamp()
    if not self.path then error("Warp Lab needs a mounted Act I DS1 asset") end
    self.job = render.preload({{kind="ds1_chunks", path=self.path, tiles=self.tiles, palette=palette}})
end

function lab:update(elapsed)
    if not self.ready then
        local status = render.preload_status(self.job)
        if not status or not status.done then return end
        render.preload_release(self.job); self.job = nil
        if status.failed > 0 then error(tostring(status.errors[1] or "Warp Lab preload failed")) end
        self:start_world()
    end
    local has_intent = self.ecs.get(self.fixture.player, "dm.lab.warp.intent") ~= nil
    local x, y = self:update_actor(has_intent and "WL" or "NU", elapsed)
    self:update_portals()
    local state = self.ecs.get(self.fixture.player, "dm.lab.warp.state")
    text.set(self.detail, "font_lab_color", string.format(
        "[white]%s[/]   [gold]position %.1f, %.1f[/]   [blue]warps %d[/]",
        state:get("event"), x, y, state:get("warp_count")), 760, "center")
end

function lab:destroy()
    if self.job then render.preload_release(self.job); self.job = nil end
    if self.fixture_module then self.fixture_module.destroy(self.fixture) end
end

return lab
