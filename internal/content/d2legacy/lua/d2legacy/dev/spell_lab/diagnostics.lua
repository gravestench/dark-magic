-- Read-only Spell Lab legend over the production game-world scene.

local ecs = require("engine.ecs/v1")
local render = require("engine.render/v1")
local text = require("d2legacy.ui.text")

local M = {}

local function label(parent, value, y)
    local node = render.create("hud", parent)
    text.set(node, "font_lab_color", value, 780, "center")
    node:set_position(400, y)
    node:set_z(1500001)
    return node
end

local function controlled_player()
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.world.player_control" } })) do
        if ecs.get(entity, "d2legacy.world.player_control"):get("player") == "local-player" then
            return entity
        end
    end
    return nil
end

function M.create(scene)
    local state = { root = render.create("hud", scene.root), last_detail = "" }
    state.backdrop = render.create("hud", state.root)
    state.backdrop:fill_rect(1, 1, 0, 0, 0, 190)
    state.backdrop:set_scale(800, 78)
    state.backdrop:set_position(400, 39)
    state.backdrop:set_z(1500000)
    state.title = label(
        state.root,
        "[gold]SPELL LAB[/]  [white]CLICK A HUD SKILL ICON TO SELECT | LEFT/RIGHT CLICK THE WORLD TO CAST[/]",
        18
    )
    state.detail = label(state.root, "[gray]WAITING FOR AUTHORITATIVE PLAYER ADMISSION[/]", 50)
    return state
end

function M.update(state)
    local player = controlled_player()
    if not player then
        return
    end
    local vitals = ecs.get(player, "d2legacy.player.vitals")
    local assigned = ecs.get(player, "d2legacy.player.skill_assignment")
    local learned = 0
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.learned_skill" } })) do
        if ecs.get(entity, "d2legacy.player.learned_skill"):get("owner"):id() == player:id() then
            learned = learned + 1
        end
    end
    local value = string.format(
        "[blue]%d EXACT-ID SKILLS[/]  [green]LEFT %d[/]  [gold]RIGHT %d[/]  [white]MANA %d / %d[/]",
        learned,
        assigned:get("left"),
        assigned:get("right"),
        vitals:get("mana"),
        vitals:get("max_mana")
    )
    if value ~= state.last_detail then
        state.last_detail = value
        text.set(state.detail, "font_lab_color", value, 780, "center")
    end
end

function M.destroy(state)
    if state and state.root and state.root:exists() then
        state.root:destroy()
    end
end

return M
