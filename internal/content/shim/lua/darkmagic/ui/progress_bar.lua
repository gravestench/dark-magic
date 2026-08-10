-- Reusable retained-mode progress/value bar.
--
-- Unlike a slider, this widget is DISPLAY ONLY. Nothing here registers an input
-- control. A caller gives it values, and it converts those values into a filled
-- rectangle plus optional text.
--
-- Keeping passive display widgets separate from interaction is useful: a health
-- bar should not accidentally become focusable just because a slider also uses
-- a 0..1 value internally.

local render = require("dm.render/v1")
local text = require("darkmagic.ui.text")

local M = {}

local function fill(node, width, height, color)
    -- Renderer rectangles must have positive dimensions. Even a zero-percent bar
    -- gets a 1px backing texture; visibility below decides whether it is shown.
    node:fill_rect(
        math.max(1, width),
        math.max(1, height),
        color[1], color[2], color[3], color[4] or 255
    )
end

function M.create(root, definition, label, options)
    definition = definition or {}
    options = options or {}

    local x = assert(definition.x, "progress bar x is required")
    local y = assert(definition.y, "progress bar y is required")
    local width = assert(definition.width, "progress bar width is required")
    local height = assert(definition.height, "progress bar height is required")
    local minimum = definition.min or 0
    local maximum = definition.max or 1
    assert(maximum >= minimum, "progress bar max must be at least min")

    local result = {
        min = minimum,
        max = maximum,
        -- Clamp the initial value immediately so invalid caller data cannot make
        -- the fill extend outside the declared bar.
        value = math.max(minimum, math.min(maximum, definition.value or minimum)),
        root = root,
    }

    if render.assets_available() then
        result.background = render.create(options.layer or "hud", root)
        result.fill = render.create(options.layer or "hud", root)

        fill(result.background, width, height, options.background_color or {36, 30, 24, 255})
        result.background:set_position(x + width / 2, y + height / 2)

        if label and label ~= "" then
            result.label = render.create(options.layer or "hud", root)
            local _, label_height = text.set(
                result.label,
                options.label_style or "font_lab_caption",
                label,
                width,
                "left"
            )
            result.label:set_position(x + width / 2, y - 6 - label_height / 2)
        end

        if options.show_value then
            result.value_node = render.create(options.layer or "hud", root)
        end
    end

    local function refresh()
        local range = result.max - result.min

        -- Normalize arbitrary [min,max] into 0..1. A zero-sized range is treated
        -- as empty instead of dividing by zero.
        local fraction = range > 0 and (result.value - result.min) / range or 0
        local fill_width = math.max(1, width * fraction)

        if result.fill then
            fill(result.fill, fill_width, height, options.fill_color or {151, 126, 75, 255})

            -- Fill grows from LEFT to RIGHT, so its center shifts as width changes.
            result.fill:set_position(x + fill_width / 2, y + height / 2)

            -- Hide zero progress even though the backing texture is at least 1px.
            result.fill:set_visible(fraction > 0)
        end

        if result.value_node then
            -- Custom formatters receive both the raw value and normalized fraction.
            local value = options.format_value and options.format_value(result.value, fraction)
                or string.format("%d%%", math.floor(fraction * 100 + 0.5))
            text.set(result.value_node, options.value_style or "font_lab_caption", value, width, "center")
            result.value_node:set_position(x + width / 2, y + height / 2)
        end
    end

    function result:set_value(value)
        self.value = math.max(self.min, math.min(self.max, value))
        refresh()
    end

    function result:set_visible(visible)
        -- A temporary array is convenient here because there are only four
        -- optional child nodes and nil entries are simply skipped by the guard.
        for _, node in ipairs({self.background, self.fill, self.label, self.value_node}) do
            if node then node:set_visible(visible) end
        end
    end

    refresh()
    return result
end

return M
