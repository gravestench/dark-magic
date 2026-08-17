-- Read-only diagnostics layered over the production game-world scene.

local render = require("engine.render/v1")
local text = require("d2legacy.ui.text")

local M = {}
local fade_seconds = 0.10

local function label(parent, value, y)
    local node = render.create("hud", parent)
    text.set(node, "font_lab_color", value, 780, "center")
    node:set_position(400, y)
    node:set_z(1500001)
    return node
end

local function controlled_player(ecs)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.player_control" } })) do
        local control = ecs.get(entity, "d2legacy.world.player_control")
        if control:get("player") == "local-player" then
            return entity
        end
    end
    return nil
end

local function snapshot(ecs, entity)
    local position = ecs.get(entity, "d2legacy.world.position")
    local velocity = ecs.get(entity, "d2legacy.world.velocity")
    local location = ecs.get(entity, "d2legacy.world.location")
    local animation = ecs.get(entity, "d2legacy.player.animation")
    return {
        x = position:get("x"),
        y = position:get("y"),
        velocity_x = velocity:get("x"),
        velocity_y = velocity:get("y"),
        level_id = location:get("level_id"),
        mode = animation:get("mode"),
    }
end

local function update_fade(state, elapsed)
    local remaining = math.max(0, state.fade_remaining or 0)
    local alpha = math.floor(255 * math.min(1, remaining / fade_seconds))
    state.fade:fill_rect(1, 1, 0, 0, 0, alpha)
    state.fade:set_visible(alpha > 0)
    state.fade_remaining = math.max(0, remaining - (elapsed or 0))
end

function M.create(scene)
    local state = { crossings = 0, fade_remaining = 0, ecs = require("engine.ecs/v1") }
    state.backdrop = render.create("hud", scene.root)
    state.backdrop:fill_rect(1, 1, 0, 0, 0, 190)
    state.backdrop:set_scale(800, 82)
    state.backdrop:set_position(400, 41)
    state.backdrop:set_z(1500000)
    state.title = label(scene.root, "[gold]WARP LAB[/]  [white]CLICK THE BLUE OR RED WARP TO TRAVEL[/]", 18)
    state.detail = label(scene.root, "[gray]WAITING FOR AUTHORITATIVE PLAYER ADMISSION[/]", 48)
    state.help = label(
        scene.root,
        "[white]Production movement, collision, interaction, relocation, animation, and camera authority[/]",
        70
    )
    state.fade = render.create("hud", scene.root)
    state.fade:fill_rect(1, 1, 0, 0, 0, 0)
    state.fade:set_scale(800, 600)
    state.fade:set_position(400, 300)
    state.fade:set_z(2000000)
    state.fade:set_visible(false)
    return state
end

function M.update(state, _, elapsed)
    local player = controlled_player(state.ecs)
    if not player then
        update_fade(state, elapsed)
        return
    end
    local current = snapshot(state.ecs, player)
    if state.level_id and current.level_id ~= state.level_id then
        state.crossings = state.crossings + 1
        state.fade_remaining = fade_seconds
    end
    state.level_id = current.level_id
    text.set(
        state.detail,
        "font_lab_color",
        string.format(
            "[blue]LEVEL %d[/]  [gold]POSITION %.1f, %.1f[/]  [green]VELOCITY %.1f, %.1f[/]  "
                .. "[white]MODE %s[/]  [red]CROSSINGS %d[/]",
            current.level_id,
            current.x,
            current.y,
            current.velocity_x,
            current.velocity_y,
            current.mode,
            state.crossings
        ),
        780,
        "center"
    )
    update_fade(state, elapsed)
end

function M.destroy(state)
    if not state then
        return
    end
    for _, node in ipairs({ state.backdrop, state.title, state.detail, state.help, state.fade }) do
        if node and node:exists() then
            node:destroy()
        end
    end
end

return M
