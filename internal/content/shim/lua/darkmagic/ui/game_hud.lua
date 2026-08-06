-- Authored in-game HUD composition for the supported 800x600 profile.
--
-- The manifest owns asset/frame/placement facts. This module owns only the
-- disposable render-node lifecycle and value-driven clipping behavior.
local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")

local M = {}

local function dc6_at(root, sheet, palette, frame, x, y)
    local node = render.create("hud", root)
    local width, height = node:set_dc6(sheet, palette, 0, frame)
    node:set_position(x + width / 2, y + height / 2)
    return node, width, height
end

local function dc6_bottom(root, sheet, palette, frame, x, bottom)
    local node = render.create("hud", root)
    local width, height = node:set_dc6(sheet, palette, 0, frame)
    node:set_position(x + width / 2, bottom - height / 2)
    return node
end

local function add_globe(root, definition, sheet, overlap_sheet, palette)
    local liquid = dc6_at(root, sheet, palette, definition.frame, definition.x, definition.y)
    local visible_height = math.floor(definition.height * definition.fill + 0.5)
    liquid:set_clip(
        definition.x,
        definition.y + definition.height - visible_height,
        definition.width,
        visible_height
    )
    dc6_at(
        root,
        overlap_sheet,
        palette,
        definition.overlap_frame,
        definition.overlap_x,
        definition.overlap_y
    )
end

function M.create(root, definition, palettes)
    local hud = { root = render.create("hud", root), controls = controls.new(), running = false, menu_open = false }
    local palette = palettes[definition.palette]

    for _, part in ipairs(definition.panel_parts) do
        dc6_bottom(hud.root, definition.panel_sheet, palette, part.frame, part.x, part.bottom)
    end

    local globes = definition.globes
    add_globe(hud.root, globes.health, globes.sheet, globes.overlap_sheet, palette)
    add_globe(hud.root, globes.mana, globes.sheet, globes.overlap_sheet, palette)

    for _, bar in ipairs({ definition.stamina, definition.experience }) do
        local node = render.create("hud", hud.root)
        local color = bar.color
        node:fill_rect(
            math.floor(bar.width * bar.fill + 0.5),
            bar.height,
            color.red,
            color.green,
            color.blue,
            color.alpha
        )
        node:set_position(
            bar.x + math.floor(bar.width * bar.fill + 0.5) / 2,
            bar.y + bar.height / 2
        )
    end

    local skills = definition.skills
    for _, placement in ipairs({ skills.left, skills.right }) do
        dc6_at(hud.root, skills.sheet, palette, skills.frame, placement.x, placement.y)
    end

    local run = definition.run
    local run_node = dc6_at(hud.root, run.sheet, palette, run.walk_frame, run.x, run.y)
    hud.controls:add({
        id = "run",
        label = "Run/Walk",
        x = run.x,
        y = run.y,
        width = run.width,
        height = run.height,
        on_activate = function()
            hud.running = not hud.running
            run_node:set_dc6(run.sheet, palette, 0, hud.running and run.run_frame or run.walk_frame)
        end,
    })

    local minipanel = definition.minipanel
    local minipanel_node = dc6_at(hud.root, minipanel.sheet, palette, 0, minipanel.x, minipanel.y)
    minipanel_node:set_visible(false)
    for index, button in ipairs(minipanel.buttons) do
        local button_definition = button
        local node = dc6_at(
            hud.root,
            minipanel.button_sheet,
            palette,
            button_definition.frame,
            minipanel.button_x + (index - 1) * minipanel.button_step,
            minipanel.button_y
        )
        node:set_visible(false)
        local control = hud.controls:add({
            id = "minipanel_" .. button_definition.id,
            label = assert(locale.text(button_definition.label)),
            x = minipanel.button_x + (index - 1) * minipanel.button_step,
            y = minipanel.button_y,
            width = minipanel.button_width,
            height = minipanel.button_height,
            enabled = button_definition.enabled,
            visible = false,
            on_activate = function()
                if button_definition.scene then
                    scenes.push(button_definition.scene)
                end
            end,
        })
        control.node = node
    end

    local menu = definition.menu
    local menu_node = dc6_at(hud.root, menu.sheet, palette, menu.closed_frame, menu.x, menu.y)
    hud.controls:add({
        id = "minipanel_toggle",
        label = "Open/Close Mini-panel",
        x = menu.x,
        y = menu.y,
        width = menu.width,
        height = menu.height,
        on_activate = function()
            hud.menu_open = not hud.menu_open
            menu_node:set_dc6(menu.sheet, palette, 0, hud.menu_open and menu.open_frame or menu.closed_frame)
            minipanel_node:set_visible(hud.menu_open)
            for _, button in ipairs(minipanel.buttons) do
                local control = hud.controls:get("minipanel_" .. button.id)
                control.visible = hud.menu_open
                control.node:set_visible(hud.menu_open)
            end
        end,
    })
    return hud
end

function M.update(hud)
    hud.controls:update()
end

return M
