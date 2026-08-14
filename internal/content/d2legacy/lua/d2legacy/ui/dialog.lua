-- Reusable modal dialogs.
--
-- A dialog is NOT a special native window. It is ordinary retained presentation
-- plus its own controls.Manager. Giving it a dedicated focus scope means arrow
-- keys, Enter, and pointer interaction stay inside the modal instead of falling
-- through to controls behind it.
--
-- This file also shows a useful callback pattern: a caller provides the semantic
-- decision function while the reusable dialog owns only presentation/lifecycle.

local render = require("engine.render/v1")
local controls = require("d2legacy.ui.controls")
local label_button = require("d2legacy.ui.label_button")
local text = require("d2legacy.ui.text")
local text_field = require("d2legacy.ui.text_field")

local M = {}

-- Text-entry dialog.
--
-- Two `_` parameters are intentionally unused compatibility arguments. Naming a
-- parameter `_` tells readers "the call shape includes this, but this function
-- does not need it."
function M.text_entry(parent, definition, _, popup_palette, _, prompt, initial, on_accept)
    local dialog = {
        open = true,
        nodes = {},
        manager = controls.new(),
    }

    -- The root owns the popup's backing art. Child label-button visuals created
    -- below use the same scene scope even though they are siblings under parent.
    dialog.root = render.create("modal", parent)
    dialog.root:set_position(
        definition.x + definition.width / 2,
        definition.y + definition.height / 2
    )

    if render.assets_available() then
        -- Prefer the actual authored multi-frame popup art.
        dialog.root:set_dc6_combined(definition.sheet, popup_palette, 0, 0)
    else
        -- Headless/dev fallback preserves geometry and control behavior.
        dialog.root:fill_rect(definition.width, definition.height, 20, 15, 10, 245)
    end

    -- Use the same authored textbox composition demonstrated by UI Lab and used
    -- by character creation. The dialog still owns focus isolation, while the
    -- widget owns its DC6 backing, readable label/value styles, and cursor.
    local configured_field = definition.field or {}
    local field = text_field.create(parent, dialog.manager, "value", {
        id = "value",
        scope = "dialog",
        value = initial or "",
        max_length = configured_field.max_length or definition.max_length or 255,
        x = configured_field.x or definition.x + 20,
        y = configured_field.y or definition.y + 70,
        width = configured_field.width or definition.width - 40,
        height = configured_field.height or 30,
        kind = configured_field.kind,
        palette = configured_field.palette,
        sheet = configured_field.sheet,
        frame = configured_field.frame,
        label_align = configured_field.label_align,
    }, prompt, {
        layer = "modal",
        scope = "dialog",
        label_style = configured_field.label_style,
        text_style = configured_field.text_style,
    })
    dialog.field = field

    -- text_field.create returns each retained child so the composite dialog can
    -- hide the entire widget atomically when cancelled or accepted.
    for _, node in ipairs({ field.background_node, field.value_node, field.label_node }) do
        if node then dialog.nodes[#dialog.nodes + 1] = node end
    end

    local ok_control = label_button.create(parent, dialog.manager, {
        id = "ok",
        x = definition.x + 35,
        y = definition.y + 125,
        width = 80,
        height = 30,
    }, "OK", {
        layer = "modal",
        scope = "dialog",
        normal_style = "dialog_button_normal",
        hover_style = "dialog_button_hover",
        on_activate = function()
            -- The caller may return literal false to say validation failed and
            -- keep the dialog open. Any other return value accepts/closes it.
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
        normal_style = "dialog_button_normal",
        hover_style = "dialog_button_hover",
        on_activate = function()
            dialog:close()
        end,
    })

    -- Keep the child label visuals in one list so close() can hide all of them.
    dialog.nodes[#dialog.nodes + 1] = ok_control.visual
    dialog.nodes[#dialog.nodes + 1] = cancel_control.visual

    -- Only controls in the `dialog` scope are eligible for focus/activation now.
    dialog.manager:set_scope("dialog")

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

-- Build a yes/no confirmation using the same focus-isolation idea.
-- The caller receives true only for explicit confirmation; the secondary choice
-- and external cancel paths can report false.
function M.confirm(parent, definition, _, popup_palette, _, message, yes_label, no_label, on_decide)
    local dialog = {
        open = true,
        nodes = {},
        manager = controls.new(),
    }

    dialog.root = render.create("modal", parent)
    dialog.root:set_position(
        definition.x + definition.width / 2,
        definition.y + definition.height / 2
    )

    if render.assets_available() then
        dialog.root:set_dc6_combined(definition.sheet, popup_palette, 0, 0)
    else
        dialog.root:fill_rect(definition.width, definition.height, 20, 15, 10, 245)
    end

    function dialog:close(decision)
        -- Idempotence: a second close request after the first does nothing. This
        -- prevents duplicate decision callbacks from two inputs in one frame.
        if not self.open then
            return
        end

        self.open = false
        self.root:set_visible(false)
        for _, node in ipairs(self.nodes) do
            node:set_visible(false)
        end

        -- Normalize any truthy-ish input to a strict boolean decision.
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
