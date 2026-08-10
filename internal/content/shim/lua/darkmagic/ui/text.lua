-- Shared Diablo-style bitmap text construction.
--
-- If you are new to game UI, this file demonstrates an important separation:
-- a screen asks for a SEMANTIC style such as `panel_heading`; the manifest says
-- which bitmap font, palette, transform, and color that style actually means.
--
-- That keeps filenames and palette details out of every screen and lets a mod
-- replace the whole visual language by changing data instead of rewriting Lua.

-- Engine capability that owns retained render nodes and native text resources.
local render = require("dm.render/v1")
-- Engine capability used here only to load the versioned presentation manifest.
local data = require("dm.data/v1")

-- Load once at module import. Every helper below reads the same validated table.
local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))

-- Public module table returned at the bottom.
local text = {}

-- Translate a friendly style name into two pieces of data:
--   style = color/transform/alignment-ish presentation choices
--   font  = the actual bitmap font table/sheet/palette definition
--
-- Returning two values is ordinary Lua; callers can receive both at once.
local function resolve(style_name)
    local style = assert(manifest.text_styles[style_name], "unknown text style: " .. tostring(style_name))
    local font = assert(manifest.fonts[style.font], "unknown font for text style: " .. tostring(style_name))
    return style, font
end

-- Low-level helper shared by both semantic styles and exact-font calls.
local function render_font(node, font, value, width, alignment, options)
    -- Callers may omit options entirely. An empty table makes all `options.foo`
    -- reads below safe.
    options = options or {}
    local color = options.color or {}

    -- UI callers sometimes pass numbers. Converting here means every caller
    -- does not need to remember `tostring` separately. Nil becomes empty text.
    value = tostring(value or "")

    if options.text_color then
        -- Dark Magic's bitmap text parser understands tags like `[gold]`.
        -- Prefixing one tag is a compact way to select a palette-authored slot.
        value = "[" .. options.text_color .. "]" .. value
    end

    -- `node:set_text` asks the retained renderer to construct/update the text
    -- image owned by this node. Lua supplies descriptive data, not a GPU object.
    return node:set_text(
        font.table,
        font.sheet,
        assert(manifest.palettes[font.palette], "unknown bitmap-font palette"),
        value,
        {
            -- Missing RGBA channels default to fully opaque white modulation.
            red = color.red or 255,
            green = color.green or 255,
            blue = color.blue or 255,
            alpha = color.alpha or 255,

            -- This expression reads as:
            --   if options.transform exists, validate/look it up;
            --   otherwise use the empty string (no transform).
            transform = options.transform and assert(
                manifest.palette_transforms[options.transform],
                "unknown bitmap-font palette transform: " .. tostring(options.transform)
            ) or "",

            -- Width zero means "no wrapping limit" to the renderer.
            max_width = width or 0,
            -- An explicit function argument wins, then an options default, then
            -- centered text as the final fallback.
            align = alignment or options.align or "center",
        }
    )
end

-- Apply one named semantic style.
--
-- The returned width/height are useful because retained nodes are positioned by
-- their center while many old UI facts are written as top/left coordinates.
function text.set(node, style_name, value, width, alignment)
    local style, font = resolve(style_name)
    return render_font(node, font, value, width, alignment, style)
end

-- Sometimes reverse-engineered presentation facts identify an exact original
-- font/palette pairing but there is no reusable semantic style for it yet. This
-- escape hatch still keeps actual asset paths inside the manifest.
function text.set_font(node, font_name, value, width, alignment, options)
    local font = assert(manifest.fonts[font_name], "unknown bitmap font: " .. tostring(font_name))
    return render_font(node, font, value, width, alignment, options)
end

-- Convenience helper for the common pattern "create a text node, fill it, place
-- it from a conventional top coordinate, and return the measured dimensions."
function text.create(root, style_name, value, x, y, width, alignment, layer)
    -- Use the caller's layer if supplied; otherwise modal is a safe panel default.
    local node = render.create(layer or "modal", root)

    -- Multiple return values let us capture both rendered dimensions.
    local rendered_width, rendered_height = text.set(node, style_name, value, width, alignment)

    -- `y` here is the TOP of the text box, while retained node positioning uses
    -- its CENTER, so move down by half the measured height.
    node:set_position(x, y + rendered_height / 2)

    -- Return the handle plus measurements so callers can compose more geometry.
    return node, rendered_width, rendered_height
end

return text
