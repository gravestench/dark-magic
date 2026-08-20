local data = require("engine.data/v1")
local render = require("engine.render/v1")

local theme = assert(data.load("darkmagic/ds1-editor/ui/theme.json"))
local text = {}

-- Build the shared renderer options used by both measurement and rasterization.
local function font_options(width, alignment, style)
    style = style or {}
    return {
        red=255, green=255, blue=255, alpha=255,
        transform=style.transform and assert(theme.palette_transforms[style.transform]) or "",
        max_width=width or 0,
        align=alignment or style.align or "center",
    }
end

-- Bind one editor-owned bitmap font to a retained text node with optional wrapping and alignment.
local function render_font(node, font, value, width, alignment, style)
    return node:set_text(
        font.table,
        font.sheet,
        assert(theme.palettes[font.palette]),
        tostring(value or ""),
        font_options(width, alignment, style)
    )
end

-- Resolve a semantic style and preserve its inline color scope for measurement or drawing.
local function styled_value(style_name, value)
    local style = assert(theme.text_styles[style_name], "unknown editor text style: " .. tostring(style_name))
    local font = assert(theme.fonts[style.font], "unknown editor font: " .. tostring(style.font))
    value = tostring(value or "")
    if style.text_color then value = "[" .. style.text_color .. "]" .. value end
    return style, font, value
end

-- Return exact prospective bitmap texture bounds without creating a retained render node.
function text.measure(style_name, value, width, alignment)
    local style, font, content = styled_value(style_name, value)
    return render.measure_text(
        font.table,
        font.sheet,
        assert(theme.palettes[font.palette]),
        content,
        font_options(width, alignment, style)
    )
end

-- Apply a named semantic style so screens do not depend on concrete font files or palette paths.
function text.set(node, style_name, value, width, alignment)
    local style, font, content = styled_value(style_name, value)
    return render_font(node, font, content, width, alignment, style)
end

return text
