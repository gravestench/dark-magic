-- The upward-opening desktop skill chooser used by the two HUD skill wells.
-- Learned-skill ownership comes from the ECS snapshot. This file only arranges
-- those facts into buttons; it never decides that a character knows a skill.
local render = require("dm.render/v1")
local locale = require("dm.locale/v1")
local tooltip = require("darkmagic.ui.tooltip")

local M = {}

local function skill_name(skill)
    return skill.name_key and locale.text(skill.name_key)
        or string.format("SKILL %d", skill.id)
end

local function position(side, row_index, column, row_width, definition)
    local y = definition.left.y - row_index * definition.height
    if side == "left" then
        return definition.selector_left + (column - 1) * definition.width, y
    end
    return definition.selector_right - row_width + (column - 1) * definition.width, y
end

local function group_rows(skills, side)
    local rows, order = {}, {}
    for _, skill in ipairs(skills) do
        local allowed = side == "left" and skill.left_allowed or skill.right_allowed
        if allowed and not skill.passive and skill.list_row >= 0 then
            if not rows[skill.list_row] then
                rows[skill.list_row] = {}
                order[#order + 1] = skill.list_row
            end
            rows[skill.list_row][#rows[skill.list_row] + 1] = skill
        end
    end
    table.sort(order)
    for _, row in ipairs(order) do
        table.sort(rows[row], function(a, b)
            return side == "left" and a.id < b.id or a.id > b.id
        end)
    end
    return rows, order
end

local function add_button(selector, side, skill, x, y)
    local node = render.create("hud", selector.root)
    node:set_dc6(skill.sheet, selector.palette, 0, skill.icon)
    node:set_position(x + selector.definition.width / 2, y + selector.definition.height / 2)
    node:set_visible(false)
    local tip = tooltip.create(selector.root, skill_name(skill), x + selector.definition.width / 2, y, {})
    local control = selector.controls:add({
        id = string.format("%s_skill_choice_%d", side, skill.id),
        label = skill_name(skill), x = x, y = y,
        width = selector.definition.width, height = selector.definition.height,
        visible = false,
        on_activate = function()
            selector.assign(side, skill.id)
            M.close(selector)
        end,
        on_state = function(_, state)
            tip:set_visible(state == "hover" or state == "focused" or state == "pressed")
        end,
    })
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
    return { root=render.create("hud", root), definition=definition, palette=palette,
        controls=controls, assign=assert(assign), buttons={}, open=nil, signature=nil }
end

-- Learned skills change rarely. Build retained nodes only when their IDs or
-- levels change, not every frame.
function M.set_skills(selector, skills)
	if selector.signature ~= nil or #skills == 0 then return end
    selector.signature = true
    build_side(selector, "left", skills)
    build_side(selector, "right", skills)
end

function M.close(selector)
    selector.open = nil
    for _, button in ipairs(selector.buttons) do
        button.control.visible = false
        button.node:set_visible(false)
        button.tip:set_visible(false)
    end
end

function M.toggle(selector, side)
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
