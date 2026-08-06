local input = require("dm.input/v1")

local M = {}
local Manager = {}
Manager.__index = Manager

local function contains(control, x, y)
    return control.visible ~= false and control.enabled ~= false
        and x >= control.x and y >= control.y
        and x < control.x + control.width and y < control.y + control.height
end

local function state_of(manager, control, x, y)
    if control.visible == false then return "hidden" end
    if control.enabled == false then return "disabled" end
    if contains(control, x, y) then return "hover" end
    if manager.focus == control then return "focused" end
    return "normal"
end

function Manager:add(control)
    assert(type(control.id) == "string" and control.id ~= "", "control id is required")
    assert(type(control.x) == "number" and type(control.y) == "number", "control position is required")
    assert(type(control.width) == "number" and control.width > 0, "positive control width is required")
    assert(type(control.height) == "number" and control.height > 0, "positive control height is required")
    assert(self.by_id[control.id] == nil, "duplicate control id: " .. control.id)
    control.role = control.role or "button"
    control.enabled = control.enabled ~= false
    control.visible = control.visible ~= false
    control.focusable = control.focusable ~= false
    control.state = control.enabled and "normal" or "disabled"
    self.controls[#self.controls + 1] = control
    self.by_id[control.id] = control
    if self.focus == nil and control.focusable and control.enabled and control.visible then
        self.focus = control
    end
    return control
end

function Manager:get(id) return self.by_id[id] end

function Manager:set_enabled(id, enabled)
    local control = assert(self.by_id[id], "unknown control: " .. id)
    control.enabled = enabled == true
    if self.focus == control and not control.enabled then self:move_focus(1) end
end

function Manager:move_focus(delta)
    if #self.controls == 0 then self.focus = nil return nil end
    local start = 0
    for index, control in ipairs(self.controls) do
        if control == self.focus then start = index break end
    end
    for step = 1, #self.controls do
        local index = ((start - 1 + delta * step) % #self.controls) + 1
        local candidate = self.controls[index]
        if candidate.focusable and candidate.enabled and candidate.visible then
            self.focus = candidate
            return candidate
        end
    end
    self.focus = nil
    return nil
end

function Manager:activate(control)
    if not control or not control.enabled or not control.visible then return false end
    if control.on_activate then control.on_activate(control) end
    return true
end

function Manager:update()
    if input.pressed("down") or input.pressed("right") then self:move_focus(1) end
    if input.pressed("up") or input.pressed("left") then self:move_focus(-1) end
    local x, y = input.cursor()
    local hovered = nil
    for _, control in ipairs(self.controls) do
        if contains(control, x, y) then hovered = control end
    end
    if hovered and hovered.focusable then self.focus = hovered end
    if input.pressed("pointer_primary") then self:activate(hovered) end
    if input.pressed("confirm") then self:activate(self.focus) end
    for _, control in ipairs(self.controls) do
        local next_state = state_of(self, control, x, y)
        if control.state ~= next_state then
            control.state = next_state
            if control.on_state then control.on_state(control, next_state) end
        end
    end
end

function Manager:accessibility()
    local result = {}
    for _, control in ipairs(self.controls) do
        result[#result + 1] = {
            id = control.id,
            role = control.role,
            label = control.label or control.id,
            enabled = control.enabled,
            visible = control.visible,
            focused = self.focus == control,
        }
    end
    return result
end

function M.new()
    return setmetatable({ controls = {}, by_id = {}, focus = nil }, Manager)
end

return M
