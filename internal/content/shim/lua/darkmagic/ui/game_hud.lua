-- Authored in-game HUD composition for the supported 800x600 profile.
--
-- The manifest owns asset/frame/placement facts. This module owns only the
-- disposable render-node lifecycle and value-driven clipping behavior.
local render = require("dm.render/v1")

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
    local hud = { root = render.create("hud", root) }
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
    return hud
end

return M
