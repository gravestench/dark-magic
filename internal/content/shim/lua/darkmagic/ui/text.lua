-- Shared Diablo-style bitmap text construction.
--
-- Screens name semantic styles from the presentation manifest instead of
-- repeating font paths, palettes, and approximate colors. This keeps the
-- Blizzard-authored bitmap font and palette pairing intact while allowing a
-- mod to replace the complete visual vocabulary in one manifest override.
local render = require("dm.render/v1")
local data = require("dm.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))

local text = {}

-- Resolve a semantic style and its concrete bitmap-font definition. Failing
-- immediately produces a useful source-aware Lua error for malformed mods.
local function resolve(style_name)
    local style = assert(manifest.text_styles[style_name], "unknown text style: " .. tostring(style_name))
    local font = assert(manifest.fonts[style.font], "unknown font for text style: " .. tostring(style_name))
    return style, font
end

local function render_font(node, font, value, width, alignment, options)
    options = options or {}
    local color = options.color or {}
    value = tostring(value or "")
    if options.text_color then
        value = "[" .. options.text_color .. "]" .. value
    end
    return node:set_text(
        font.table,
        font.sheet,
        assert(manifest.palettes[font.palette], "unknown bitmap-font palette"),
        value,
        {
            red = color.red or 255,
            green = color.green or 255,
            blue = color.blue or 255,
            alpha = color.alpha or 255,
            transform = options.transform and assert(
                manifest.palette_transforms[options.transform],
                "unknown bitmap-font palette transform: " .. tostring(options.transform)
            ) or "",
            max_width = width or 0,
            align = alignment or options.align or "center",
        }
    )
end

-- Create centered-positioned text. The x coordinate is the center of the
-- requested text box, matching the retained renderer's node positioning.
-- Callers may still select left, center, or right alignment within that box.
function text.set(node, style_name, value, width, alignment)
    local style, font = resolve(style_name)
    return render_font(node, font, value, width, alignment, style)
end

-- Render one manifest-defined bitmap font without applying a semantic style.
-- This is useful when a recovered reference specifies an exact font/palette
-- pairing (for example OpenDiablo2's FontStyle42Units) but no Dark Magic style
-- exists yet. Asset paths still remain centralized in the presentation manifest.
function text.set_font(node, font_name, value, width, alignment, options)
    local font = assert(manifest.fonts[font_name], "unknown bitmap font: " .. tostring(font_name))
    return render_font(node, font, value, width, alignment, options)
end

function text.create(root, style_name, value, x, y, width, alignment, layer)
    local node = render.create(layer or "modal", root)
    local rendered_width, rendered_height = text.set(node, style_name, value, width, alignment)
    node:set_position(x, y + rendered_height / 2)
    return node, rendered_width, rendered_height
end

return text
