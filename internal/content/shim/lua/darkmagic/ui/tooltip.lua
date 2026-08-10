-- Origin-aware Diablo tooltip with a translucent black backing box.
local render = require("dm.render/v1")
local data = require("dm.data/v1")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local tooltip = {}

local function anchored_start(value, size, origin)
    if origin == "left" or origin == "top" then
        return value
    elseif origin == "center" then
        return value - size / 2
    elseif origin == "right" or origin == "bottom" then
        return value - size
    end
    error("invalid tooltip origin: " .. tostring(origin))
end

function tooltip.create(root, value, x, y, options)
    options = options or {}
    local layer = options.layer or "modal"
    local result = {
        background = render.create(layer, root),
        label = render.create(layer, root),
    }
    local padding_x, padding_y = options.padding_x or 4, options.padding_y or 2
    function result:set_text(next_value)
        if self.value == next_value then return end
        self.value = next_value
        local text_width, text_height = text.set(
            self.label,
            options.style or "tooltip",
            next_value,
            options.max_width or 0,
            "center"
        )
        local width, height = text_width + padding_x * 2, text_height + padding_y * 2
        local left = anchored_start(x, width, options.origin_x or "center")
        local top = anchored_start(y, height, options.origin_y or "bottom")
        left = math.max(0, math.min(manifest.resolution.width - width, left))
        top = math.max(0, math.min(manifest.resolution.height - height, top))

        self.background:fill_rect(width, height, 0, 0, 0, options.alpha or 200)
        self.background:set_position(left + width / 2, top + height / 2)
        self.label:set_position(left + width / 2, top + padding_y + text_height / 2)
    end

    function result:set_visible(visible)
        self.background:set_visible(visible)
        self.label:set_visible(visible)
    end

    result:set_text(value)
    result:set_visible(false)
    return result
end

return tooltip
