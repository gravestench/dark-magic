-- The upward-opening desktop skill chooser used by the two HUD skill wells.
--
-- This file is PRESENTATION ONLY. The list of learned skills arrives from an ECS
-- snapshot, including which side each skill may be assigned to. This code filters
-- and arranges those facts; it never grants a skill or decides that a character
-- is allowed to use one.

local render = require("engine.render/v1")
local locale = require("engine.locale/v1")
local tooltip = require("d2legacy.ui.tooltip")

local M = {}

local function skill_name(skill)
    -- Use localized game-data name when present. The fallback keeps diagnostics
    -- readable even if a localization key is unavailable.
    return skill.name_key and locale.text(skill.name_key)
        or string.format("SKILL %d", skill.id)
end

-- Compute one icon's top-left position in the upward-opening selector.
local function position(side, row_index, column, row_width, definition)
    -- Row 1 sits one icon-height above the normal HUD wells, row 2 two heights, etc.
    local y = definition.left.y - row_index * definition.height

    if side == "left" then
        -- Left selector grows rightward from its authored left edge.
        return definition.selector_left + (column - 1) * definition.width, y
    end

    -- Right selector is aligned to a right edge, so subtract the whole row width
    -- before adding each column offset.
    return definition.selector_right - row_width + (column - 1) * definition.width, y
end

local function group_rows(skills, side)
    -- `rows` maps recovered list_row -> array of skills.
    -- `order` separately remembers which row keys exist so we can sort them.
    local rows, order = {}, {}

    for _, skill in ipairs(skills) do
        -- Compact conditional chooses the side-specific authority flag.
        local allowed = side == "left" and skill.left_allowed or skill.right_allowed

        -- Passive skills do not belong in a selectable action well, and negative
        -- list rows mean "not placed in this selector layout."
        if allowed and not skill.passive and skill.list_row >= 0 then
            if not rows[skill.list_row] then
                rows[skill.list_row] = {}
                order[#order + 1] = skill.list_row
            end
            rows[skill.list_row][#rows[skill.list_row] + 1] = skill
        end
    end

    -- Table keys are not automatically ordered, so sort the recovered row IDs.
    table.sort(order)

    for _, row in ipairs(order) do
        -- Left and right selectors intentionally order skills in opposite ID
        -- directions so the rows grow toward the center in the desktop layout.
        table.sort(rows[row], function(a, b)
            return side == "left" and a.id < b.id or a.id > b.id
        end)
    end

    return rows, order
end

local function add_button(selector, side, skill, x, y)
    -- Skill icon is one retained DC6 node.
    local node = render.create("hud", selector.root)
    node:set_dc6(skill.sheet, selector.palette, 0, skill.icon)
    node:set_position(x + selector.definition.width / 2, y + selector.definition.height / 2)
    node:set_visible(false)

    -- Tooltip is a separate reusable composite that starts hidden.
    local tip = tooltip.create(selector.root, skill_name(skill), x + selector.definition.width / 2, y, {})

    -- Input remains an ordinary controls.Manager entry.
    local control = selector.controls:add({
        id = string.format("%s_skill_choice_%d", side, skill.id),
        label = skill_name(skill), x = x, y = y,
        width = selector.definition.width, height = selector.definition.height,
        visible = false,

        on_activate = function()
            -- `assign` is supplied by the HUD/game scene. In the real game it
            -- submits the assignment intent to authoritative player state.
            selector.assign(side, skill.id)
            M.close(selector)
        end,

        on_state = function(_, state)
            tip:set_visible(state == "hover" or state == "focused" or state == "pressed")
        end,
    })

    -- Keep all three related pieces together for bulk open/close operations.
    selector.buttons[#selector.buttons + 1] = { side=side, node=node, tip=tip, control=control }
end

local function build_side(selector, side, skills)
    local rows, order = group_rows(skills, side)

    for row_index, row in ipairs(order) do
        local entries = rows[row]
        local width = #entries * selector.definition.width

        for column, skill in ipairs(entries) do
            local x, y = position(side, row_index, column, width, selector.definition)
            add_button(selector, side, skill, x, y)
        end
    end
end

function M.create(root, definition, palette, controls, assign)
    -- A compact data-only constructor is enough here; expensive icon nodes are
    -- created later when the first learned-skill snapshot arrives.
    return { root=render.create("hud", root), definition=definition, palette=palette,
        controls=controls, assign=assert(assign), buttons={}, open=nil, signature=nil }
end

-- Learned skills change rarely. Build retained nodes only once for the current
-- selector data instead of recreating buttons every HUD frame.
function M.set_skills(selector, skills)
	if selector.signature ~= nil or #skills == 0 then return end
    selector.signature = true
    build_side(selector, "left", skills)
    build_side(selector, "right", skills)
end

function M.close(selector)
    selector.open = nil
    for _, button in ipairs(selector.buttons) do
        -- Hide BOTH interaction and presentation so invisible icons cannot still click.
        button.control.visible = false
        button.node:set_visible(false)
        button.tip:set_visible(false)
    end
end

function M.toggle(selector, side)
    -- If this side is already open, closing first means the toggle ends closed.
    local opening = selector.open ~= side
    M.close(selector)
    if not opening then return end

    selector.open = side
    for _, button in ipairs(selector.buttons) do
        if button.side == side then
            button.control.visible = true
            button.node:set_visible(true)
        end
    end
end

return M
