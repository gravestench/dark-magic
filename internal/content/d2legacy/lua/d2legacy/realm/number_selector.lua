-- Compact numeric selector used by the authored Realm Create Game pane.

local render = require("engine.render/v1")
local text = require("d2legacy.ui.text")
local button = require("d2legacy.ui.button")

local M = {}

local function arrow(x, y, frames)
    return {
        x = x,
        y = y,
        width = 16,
        height = 16,
        sheet = "data/global/ui/BIGMENU/numberarrows.dc6",
        palette = "sky",
        up_frame = frames[1],
        down_frame = frames[2],
    }
end

function M.create(root, manager, id, definition, options)
    options = options or {}
    local selector = {
        value = definition.value or definition.minimum,
        minimum = assert(definition.minimum),
        maximum = assert(definition.maximum),
    }
    local value_node

    local function draw()
        if not value_node then
            return
        end
        text.set(value_node, options.style or "realm_lobby_gold", selector.value, definition.width, "center")
        value_node:set_position(definition.x + definition.width / 2, definition.y + definition.height / 2)
    end

    local function change(delta)
        local next_value = math.max(selector.minimum, math.min(selector.maximum, selector.value + delta))
        if next_value == selector.value then
            return
        end
        selector.value = next_value
        draw()
        if options.on_change then
            options.on_change(selector, next_value)
        end
    end

    if render.assets_available() then
        value_node = render.create(options.layer or "hud", root)
    end
    selector.up = button.create(root, manager, id .. "_up", arrow(definition.arrow_x, definition.y, { 0, 1 }), "", {
        layer = options.layer or "hud",
        show_label = false,
        on_activate = function()
            change(1)
        end,
    })
    selector.down =
        button.create(root, manager, id .. "_down", arrow(definition.arrow_x, definition.y + 12, { 2, 3 }), "", {
            layer = options.layer or "hud",
            show_label = false,
            on_activate = function()
                change(-1)
            end,
        })

    function selector:set_value(value)
        self.value = math.max(self.minimum, math.min(self.maximum, value))
        draw()
    end

    selector.node = value_node
    draw()
    return selector
end

return M
