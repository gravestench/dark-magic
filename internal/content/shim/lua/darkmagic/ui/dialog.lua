-- Reusable modal text-entry dialog.
--
-- The dialog owns a modal render subtree and an isolated control focus scope.
-- Its callback returns false to keep the dialog open after validation fails.
local render = require("dm.render/v1")
local controls = require("darkmagic.ui.controls")
local label_button = require("darkmagic.ui.label_button")
local text = require("darkmagic.ui.text")

local M = {}

function M.text_entry(parent, definition, _, popup_palette, _, prompt, initial, on_accept)
    local dialog = {
        open = true,
        nodes = {},
        manager = controls.new(),
    }

    -- Prefer the authored popup art, while retaining a diagnostic fallback for
    -- tests and installations where the optional assets are unavailable.
    dialog.root = render.create("modal", parent)
    dialog.root:set_position(definition.x + definition.width / 2, definition.y + definition.height / 2)
    if render.assets_available() then
        dialog.root:set_dc6(definition.sheet, popup_palette, 0, 0)
    else
        dialog.root:fill_rect(definition.width, definition.height, 20, 15, 10, 245)
    end
    local field = dialog.manager:add_text_field({
        id = "value",
        scope = "dialog",
        label = prompt,
        value = initial or "",
        max_length = definition.max_length or 255,
        x = definition.x + 20,
        y = definition.y + 70,
        width = definition.width - 40,
        height = 30,
    })
    local ok_control = label_button.create(parent, dialog.manager, {
        id = "ok",
        x = definition.x + 35,
        y = definition.y + 125,
        width = 80,
        height = 30,
    }, "OK", {
        layer = "modal",
        scope = "dialog",
        on_activate = function()
            if on_accept(field.value) ~= false then
                dialog:close()
            end
        end,
    })
    local cancel_control = label_button.create(parent, dialog.manager, {
        id = "cancel",
        x = definition.x + 145,
        y = definition.y + 125,
        width = 80,
        height = 30,
    }, "Cancel", {
        layer = "modal",
        scope = "dialog",
        on_activate = function()
            dialog:close()
        end,
    })
    dialog.nodes[#dialog.nodes + 1] = ok_control.visual
    dialog.nodes[#dialog.nodes + 1] = cancel_control.visual
    dialog.manager:set_scope("dialog")

    -- Text is redrawn only when the field changes, rather than every frame.
    if render.assets_available() then
        dialog.text = render.create("modal", parent)
        dialog.nodes[#dialog.nodes + 1] = dialog.text
        local function redraw()
            text.set(dialog.text, "dialog_text", prompt .. "\n" .. field.value, definition.width - 40, "center")
            dialog.text:set_position(definition.x + definition.width / 2, definition.y + 70)
        end
        field.on_change = redraw
        redraw()
    end

    function dialog:update()
        if self.open then
            self.manager:update()
        end
    end

    function dialog:close()
        self.open = false
        self.root:set_visible(false)
        for _, node in ipairs(self.nodes) do
            node:set_visible(false)
        end
    end

    return dialog
end

-- Build a focus-isolated yes/no confirmation. The callback receives true only
-- for explicit confirmation; cancel input and the secondary button report false.
function M.confirm(parent, definition, _, popup_palette, _, message, yes_label, no_label, on_decide)
    local dialog = {
        open = true,
        nodes = {},
        manager = controls.new(),
    }
    dialog.root = render.create("modal", parent)
    dialog.root:set_position(definition.x + definition.width / 2, definition.y + definition.height / 2)
    if render.assets_available() then
        dialog.root:set_dc6(definition.sheet, popup_palette, 0, 0)
    else
        dialog.root:fill_rect(definition.width, definition.height, 20, 15, 10, 245)
    end

    function dialog:close(decision)
        if not self.open then
            return
        end
        self.open = false
        self.root:set_visible(false)
        for _, node in ipairs(self.nodes) do
            node:set_visible(false)
        end
        on_decide(decision == true)
    end

    local yes_control = label_button.create(parent, dialog.manager, {
        id = "yes",
        x = definition.yes.x,
        y = definition.yes.y,
        width = definition.yes.width,
        height = definition.yes.height,
    }, yes_label, {
        layer = "modal",
        scope = "dialog",
        on_activate = function()
            dialog:close(true)
        end,
    })
    local no_control = label_button.create(parent, dialog.manager, {
        id = "no",
        x = definition.no.x,
        y = definition.no.y,
        width = definition.no.width,
        height = definition.no.height,
    }, no_label, {
        layer = "modal",
        scope = "dialog",
        on_activate = function()
            dialog:close(false)
        end,
    })
    dialog.nodes[#dialog.nodes + 1] = yes_control.visual
    dialog.nodes[#dialog.nodes + 1] = no_control.visual
    dialog.manager:set_scope("dialog")

    if render.assets_available() then
        local message_node = render.create("modal", parent)
        text.set(message_node, "dialog_text", message, definition.width - 40, "center")
        message_node:set_position(definition.x + definition.width / 2, definition.y + 45)
        dialog.nodes[#dialog.nodes + 1] = message_node
    end

    function dialog:update()
        if self.open then
            self.manager:update()
        end
    end
    return dialog
end

return M
