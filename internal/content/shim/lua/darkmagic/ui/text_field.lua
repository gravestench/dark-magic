-- Authored Diablo II text-entry composition.
--
-- controls.lua already owns the hard input work: UTF-8 cursor movement,
-- backspace/delete, typed text, max length, focus, and callbacks. This file
-- supplies the Diablo-looking BOX, label, text placement, and visible cursor.

local render = require("dm.render/v1")
local data = require("dm.data/v1")
local text = require("darkmagic.ui.text")
local compat = require("darkmagic.ui.compat")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local M = {}

-- Pick the historically appropriate text-box artwork from a friendly semantic kind.
local function sheet_for(defaults, kind)
    if kind == "name" then return defaults.name_sheet end
    if kind == "ip" then return defaults.ip_sheet end
    return defaults.generic_sheet
end

-- Character names intentionally accept a narrower set than arbitrary text.
-- This filter walks bytes because the allowed set is ASCII only.
local function name_filter(value)
    local result = {}
    for index = 1, #value do
        local character = value:sub(index, index)
        if character:match("[A-Za-z0-9_-]") then
            result[#result + 1] = character
        end
    end
    return table.concat(result)
end

-- Draw a simple underscore cursor INTO the displayed string.
--
-- `cursor` is measured in UTF-8 characters, while Lua string slicing uses byte
-- positions. Continuation bytes are 128..191, so a byte outside that range marks
-- a new character start. This mirrors the bookkeeping in controls.lua.
local function marker_at(value, cursor)
    if cursor <= 0 then return "_" .. value end

    local seen = 0
    for index = 1, #value do
        local byte = value:byte(index)
        if byte < 128 or byte >= 192 then
            if seen == cursor then
                return value:sub(1, index - 1) .. "_" .. value:sub(index)
            end
            seen = seen + 1
        end
    end

    -- Cursor after the last character.
    return value .. "_"
end

function M.create(root, manager, id, definition, label, options)
    definition = definition or {}
    options = options or {}

    local defaults = compat.widgets.text_field
    local x = assert(definition.x, "text field x is required")
    local y = assert(definition.y, "text field y is required")

    local palette_name = definition.palette or options.palette or "fechar"
    local palette = assert(manifest.palettes[palette_name], "unknown text field palette")
    local sheet = definition.sheet or sheet_for(defaults, definition.kind)

    local background
    local value_node
    local label_node

    -- Width/height may come from manifest fallback or from actual decoded art.
    local width = definition.width
    local height = definition.height

    if render.assets_available() then
        background = render.create(options.layer or "hud", root)
        local decoded_width, decoded_height = background:set_dc6(sheet, palette, 0, definition.frame or 0)

        -- Character-name entry is a specifically authored frontend control. When
        -- real assets exist, their decoded dimensions are authoritative.
        if definition.kind == "name" then
            width, height = decoded_width, decoded_height
        else
            width = width or decoded_width
            height = height or decoded_height
        end

        -- The source x/y is top-left. The historical background_y value is a
        -- small authored vertical correction before converting to center space.
        background:set_position(
            x + decoded_width / 2,
            y + defaults.background_y + decoded_height / 2
        )

        -- Text gets its own node so changing characters does not rebuild the box art.
        value_node = render.create(options.layer or "hud", root)

        if label and label ~= "" then
            label_node = render.create(options.layer or "hud", root)
            local label_width = definition.label_width or width
            local _, label_height = text.set(
                label_node,
                options.label_style or defaults.label_style,
                label,
                label_width,
                "left"
            )
            label_node:set_position(
                x + label_width / 2,
                y + defaults.label_y + label_height / 2
            )
        end
    end

    -- Headless fallback geometry still gives input tests a real hit rectangle.
    width = width or 272
    height = height or 32

    local function draw_value(current)
        if not value_node then return end

        local shown = current.value

        -- Focused field shows the underscore cursor at the manager's character position.
        if current.state == "focused" then
            shown = marker_at(current.value, current.cursor or 0)
        end

        -- Leave room for the authored inset and a small right margin.
        local value_width = math.max(1, width - defaults.text_x - 4)
        local _, text_height = text.set(
            value_node,
            options.text_style or defaults.text_style,
            shown,
            value_width,
            "left"
        )

        value_node:set_position(
            x + defaults.text_x + value_width / 2,
            y + defaults.text_y + text_height / 2
        )
    end

    -- Preserve caller callback while inserting our redraw step.
    local changed = options.on_change

    local control = manager:add_text_field({
        id = id,
        label = label or id,
        x = x,
        y = y,
        width = width,
        height = height,
        value = definition.value or "",
        max_length = definition.max_length,

        -- Compact conditional: a custom filter wins. Otherwise name fields get
        -- our ASCII name filter; other field kinds accept the manager's text as-is.
        filter = definition.filter or (definition.kind == "name" and name_filter or nil),

        enabled = options.enabled,
        scope = options.scope or definition.scope,

        on_change = function(current, value)
            draw_value(current)
            if changed then changed(current, value) end
        end,

        -- Focus changes matter visually because focused text displays the cursor.
        on_state = function(current)
            draw_value(current)
        end,
    })

    draw_value(control)

    -- Expose child handles so a composite form can hide/reorder the complete field.
    control.background_node = background
    control.value_node = value_node
    control.label_node = label_node
    return control
end

return M
