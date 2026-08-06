local render = require("dm.render/v1")
local controls = require("darkmagic.ui.controls")

local M = {}

function M.text_entry(parent, definition, font, popup_palette, font_palette, prompt, initial, on_accept)
    local dialog = { open = true, nodes = {}, manager = controls.new() }
    dialog.root = render.create("modal", parent)
    dialog.root:set_position(definition.x + definition.width / 2, definition.y + definition.height / 2)
    if render.assets_available() then
        dialog.root:set_dc6(definition.sheet, popup_palette, 0, 0)
    else
        dialog.root:fill_rect(definition.width, definition.height, 20, 15, 10, 245)
    end
    local field = dialog.manager:add_text_field({
        id="value", scope="dialog", label=prompt, value=initial or "", max_length=definition.max_length or 255,
        x=definition.x + 20, y=definition.y + 70, width=definition.width - 40, height=30,
    })
    dialog.manager:add({id="ok", scope="dialog", label="OK", x=definition.x + 35,
        y=definition.y + 125, width=80, height=30, on_activate=function()
            if on_accept(field.value) ~= false then dialog:close() end
        end})
    dialog.manager:add({id="cancel", scope="dialog", label="Cancel", x=definition.x + 145,
        y=definition.y + 125, width=80, height=30, on_activate=function() dialog:close() end})
    dialog.manager:set_scope("dialog")
    if render.assets_available() then
        dialog.text = render.create("modal", parent)
        dialog.nodes[#dialog.nodes + 1] = dialog.text
        local function redraw()
            dialog.text:set_text(font.table, font.sheet, font_palette, prompt .. "\n" .. field.value, {
                red=210, green=180, blue=110, max_width=definition.width - 40, align="center"
            })
            dialog.text:set_position(definition.x + definition.width / 2, definition.y + 70)
        end
        field.on_change = redraw
        redraw()
    end
    function dialog:update() if self.open then self.manager:update() end end
    function dialog:close()
        self.open = false; self.root:set_visible(false)
        for _, node in ipairs(self.nodes) do node:set_visible(false) end
    end
    return dialog
end

return M
