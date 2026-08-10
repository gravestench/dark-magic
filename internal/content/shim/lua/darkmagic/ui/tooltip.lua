-- Origin-aware Diablo tooltip with a translucent black backing box.
--
-- Tooltips are a nice geometry lesson: callers give one anchor point `(x,y)`,
-- but that point might mean the tooltip's left edge, center, right edge, top, or
-- bottom. After measuring the text, this helper turns the requested anchor into
-- an actual top-left rectangle and clamps it inside the logical viewport.

local render = require("dm.render/v1")
local data = require("dm.data/v1")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local tooltip = {}

-- Convert one anchored coordinate into the rectangle's START coordinate.
local function anchored_start(value, size, origin)
    if origin == "left" or origin == "top" then
        -- Caller already gave the leading edge.
        return value
    elseif origin == "center" then
        -- Move left/up by half size so `value` becomes the center.
        return value - size / 2
    elseif origin == "right" or origin == "bottom" then
        -- Move by the whole size so `value` becomes the trailing edge.
        return value - size
    end
    error("invalid tooltip origin: " .. tostring(origin))
end

function tooltip.create(root, value, x, y, options)
    options = options or {}
    local layer = options.layer or "modal"

    -- Tooltip has two retained pieces: dark rectangle backing + bitmap text.
    local result = {
        background = render.create(layer, root),
        label = render.create(layer, root),
    }

    -- Lua can assign two locals from one comma-separated expression.
    local padding_x, padding_y = options.padding_x or 4, options.padding_y or 2

    function result:set_text(next_value)
        -- Avoid rebuilding identical retained text/backing.
        if self.value == next_value then return end
        self.value = next_value

        -- Text must be rendered/measured FIRST because tooltip box size depends
        -- on the actual bitmap glyph layout and wrapping.
        local text_width, text_height = text.set(
            self.label,
            options.style or "tooltip",
            next_value,
            options.max_width or 0,
            "center"
        )

        -- Add padding around the measured content.
        local width, height = text_width + padding_x * 2, text_height + padding_y * 2

        -- Convert caller's chosen anchor semantics into a top-left box.
        local left = anchored_start(x, width, options.origin_x or "center")
        local top = anchored_start(y, height, options.origin_y or "bottom")

        -- Clamp the entire tooltip into the logical viewport. This prevents an
        -- edge-of-screen item from putting its tooltip partly off-screen.
        left = math.max(0, math.min(manifest.resolution.width - width, left))
        top = math.max(0, math.min(manifest.resolution.height - height, top))

        self.background:fill_rect(width, height, 0, 0, 0, options.alpha or 200)
        self.background:set_position(left + width / 2, top + height / 2)

        -- Text is centered horizontally; vertically it starts after top padding.
        self.label:set_position(left + width / 2, top + padding_y + text_height / 2)
    end

    function result:set_visible(visible)
        -- Keep backing and label together so one cannot accidentally remain as a ghost.
        self.background:set_visible(visible)
        self.label:set_visible(visible)
    end

    -- Build content immediately, but tooltips begin hidden until a control asks
    -- to reveal them through its on_state callback.
    result:set_text(value)
    result:set_visible(false)
    return result
end

return tooltip
