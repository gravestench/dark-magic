-- Retained diagnostic graphics layered over the production game-world scene.
-- F6 hides this entire module; F3/F4/F5 remain the shared world diagnostics.

local ecs_snapshot = require("d2legacy.dev.combat_lab.snapshot")
local input = require("engine.input/v1")
local render = require("engine.render/v1")
local text = require("d2legacy.ui.text")

local M = {}
local fixed_one = 256

-- Font rasterization is cached, but asking it to rebuild identical text every
-- frame is still needless work in a performance lab. Remember each last value.
local function set_text(owner, key, node, value, width)
    owner.text_values = owner.text_values or {}
    if owner.text_values[key] == value then return end
    owner.text_values[key] = value
    text.set(node, "font_lab_color", value, width or 780, "center")
end

local function label(parent, value, x, y, width, align)
    local node = render.create("hud", parent)
    text.set(node, "font_lab_color", value, width or 760, align or "center")
    node:set_position(x or 400, y)
    node:set_z(1500001)
    return node
end

local function create_panel(scene)
    local panel = {root=render.create("hud", scene.root)}
    panel.backdrop = render.create("hud", panel.root)
    panel.backdrop:fill_rect(1, 1, 0, 0, 0, 190)
    panel.backdrop:set_scale(800, 116)
    panel.backdrop:set_position(400, 58)
    panel.backdrop:set_z(1500000)
    panel.title = label(panel.root, "", 400, 14, 780)
    panel.player = label(panel.root, "", 400, 40, 780)
    panel.target = label(panel.root, "", 400, 66, 780)
    panel.events = label(panel.root, "", 400, 92, 780)
    return panel
end

local function world_marker(parent)
    local marker = {root=render.create("world", parent)}
    marker.horizontal = render.create("world", marker.root)
    marker.horizontal:fill_rect(18, 2, 255, 64, 64, 230)
    marker.vertical = render.create("world", marker.root)
    marker.vertical:fill_rect(2, 18, 255, 64, 64, 230)
    marker.caption = render.create("world", marker.root)
    marker.root:set_z(950000)
    return marker
end

local function destroy_marker(marker)
    if marker and marker.root and marker.root:exists() then marker.root:destroy() end
end

local function event_text(event)
    if not event then return "[gray]NO COMBAT EVENTS YET[/]" end
    if event.kind == "hit_resolved" then
        return string.format("[white]TICK %d[/]  HIT [blue]%s[/] -> [gold]%s[/]  %s",
            event.tick, event.attacker_id, event.target_id,
            event.hit and "[green]HIT[/]" or "[red]MISS[/]")
    end
    if event.kind == "damage_applied" then
        return string.format("[white]TICK %d[/]  [gold]%g PHYSICAL[/]  [blue]%g HP LEFT[/]",
            event.tick, event.damage / fixed_one, event.remaining_health / fixed_one)
    end
    return string.format("[white]TICK %d[/]  [gold]%s[/]  %s -> %s",
        event.tick, event.kind:upper(), event.attacker_id, event.target_id)
end

local function nearest_monster(player, monsters)
    if not player or not player.position then return nil, nil end
    local nearest, nearest_distance
    for _, monster in ipairs(monsters) do
        if monster.position then
            local dx = monster.position.x - player.position.x
            local dy = monster.position.y - player.position.y
            local distance = math.sqrt(dx * dx + dy * dy)
            if not nearest_distance or distance < nearest_distance then
                nearest, nearest_distance = monster, distance
            end
        end
    end
    return nearest, nearest_distance
end

local function update_panel(state, snapshot)
    local player = snapshot.player
    set_text(state.panel, "title", state.panel.title, string.format(
        "[gold]COMBAT LAB[/]  [white]LAST EVENT TICK %d[/]  [green]F6 DATA ON[/]  [white]F3 COLLISION  F4 TILES  F5 ORIGINS[/]",
        snapshot.tick))
    if not player then
        set_text(state.panel, "player", state.panel.player, "[red]WAITING FOR AUTHORITATIVE PLAYER ADMISSION[/]")
        set_text(state.panel, "target", state.panel.target, "[gray]HOSTILE POPULATION PENDING[/]")
        set_text(state.panel, "events", state.panel.events, "[gray]NO COMBAT EVENTS YET[/]")
        return
    end
    local profile, animation = player.profile or {}, player.animation or {}
    local phase = player.attack and "ATTACK"
        or (player.approach and "APPROACH" or "READY")
    set_text(state.panel, "player", state.panel.player, string.format(
        "[blue]%s[/]  POS [white]%.1f,%.1f[/]  MODE [green]%s[/]  PHASE [gold]%s[/]  RANGE %.1f  DAMAGE %g-%g",
        player.id, player.position.x, player.position.y, animation.mode or "?", phase,
        profile.range or 0, (profile.physical_min or 0)/fixed_one,
        (profile.physical_max or 0)/fixed_one))
    local nearest, distance = nearest_monster(player, snapshot.monsters)
    if nearest then
        local stats, ai = nearest.stats or {}, nearest.ai or {}
        set_text(state.panel, "target", state.panel.target, string.format(
            "[red]%d HOSTILES[/]  NEAREST [gold]%s[/] @ %.1f,%.1f  DIST [white]%.1f[/]  HP [green]%g/%g[/]  AI [blue]%s[/]",
            #snapshot.monsters, nearest.label, nearest.position.x, nearest.position.y, distance,
            (stats.health or 0)/fixed_one, (stats.max_health or 0)/fixed_one,
            ai.state or "?"))
    else
        set_text(state.panel, "target", state.panel.target, "[red]NO LIVE HOSTILES IN CURRENT LEVEL[/]")
    end
    set_text(state.panel, "events", state.panel.events, event_text(snapshot.events[#snapshot.events]))
end

local function update_markers(state, scene, snapshot)
    if not scene.world or not scene.map then return end
    local seen = {}
    for _, monster in ipairs(snapshot.monsters) do
        if monster.position and not monster.death then
            local key = tostring(monster.entity_id)
            local marker = state.markers[key]
            if not marker then marker=world_marker(scene.map.root);state.markers[key]=marker end
            local x, y = scene.world:subtile_to_pixel(monster.position.x, monster.position.y)
            marker.root:set_position(x-scene.world_canvas_width/2, y-scene.world_canvas_height/2)
            local stats, ai = monster.stats or {}, monster.ai or {}
            set_text(marker, "caption", marker.caption, string.format(
                "[red]%s[/] [white]%g/%g[/] [blue]%s[/]",
                monster.label, (stats.health or 0)/fixed_one,
                (stats.max_health or 0)/fixed_one, ai.state or "?"), 220)
            marker.caption:set_position(0, -34)
            seen[key] = true
        end
    end
    for key, marker in pairs(state.markers) do
        if not seen[key] then destroy_marker(marker);state.markers[key]=nil end
    end
end

function M.create(scene)
    local state = {visible=true, markers={}, panel=create_panel(scene)}
    return state
end

function M.update(state, scene)
    if input.pressed("debug_combat") then
        state.visible = not state.visible
        state.panel.root:set_visible(state.visible)
        for _, marker in pairs(state.markers) do marker.root:set_visible(state.visible) end
    end
    if not state.visible then return end
    local level_id = scene.world_level_id or 2
    local snapshot = ecs_snapshot.read(level_id)
    update_panel(state, snapshot)
    update_markers(state, scene, snapshot)
end

function M.destroy(state)
    if not state then return end
    for _, marker in pairs(state.markers) do destroy_marker(marker) end
    if state.panel and state.panel.root:exists() then state.panel.root:destroy() end
end

return M
