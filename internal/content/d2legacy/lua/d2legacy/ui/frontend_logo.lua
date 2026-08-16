-- Shared four-layer frontend logo composition.
--
-- Title, main menu, and Realm login all use the same black/fire halves and one
-- synchronized animation clock. Keeping that recipe here prevents those
-- screens from drifting as presentation work continues.

local render = require("engine.render/v1")
local dc6 = require("d2legacy.ui.dc6")
local compat = require("d2legacy.ui.compat")

local M = {}

function M.create(root, definition, palettes, layer)
    if not render.assets_available() then return nil end
    local logo = {
        nodes = {},
        elapsed = 0,
    }
    logo.black_left = render.create(layer or "hud", root)
    logo.black_right = render.create(layer or "hud", root)
    logo.fire_left = render.create(layer or "hud", root)
    logo.fire_right = render.create(layer or "hud", root)
    logo.nodes = {
        logo.black_left,
        logo.black_right,
        logo.fire_left,
        logo.fire_right,
    }
    logo.fire_left:set_blend(compat.draw_mode(3))
    logo.fire_right:set_blend(compat.draw_mode(3))
    local palette = assert(palettes[definition.palette], "unknown frontend logo palette")
    dc6.anchored_composite(
        { logo.black_left, logo.fire_left },
        { definition.black_left, definition.fire_left },
        palette, definition.anchor.x, definition.anchor.y,
        definition.frames_per_second, "loop"
    )
    dc6.anchored_composite(
        { logo.black_right, logo.fire_right },
        { definition.black_right, definition.fire_right },
        palette, definition.anchor.x, definition.anchor.y,
        definition.frames_per_second, "loop"
    )
    dc6.pause_animations(logo.nodes)
    dc6.synchronize_animations(logo.nodes, 0)
    return logo
end

function M.update(logo, elapsed)
    if not logo then return end
    logo.elapsed = logo.elapsed + (elapsed or 0)
    dc6.synchronize_animations(logo.nodes, logo.elapsed)
end

return M
