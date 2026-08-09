-- Reusable retained-mode control manager for shim and mod-authored screens.
--
-- Controls are plain Lua tables. The manager owns focus, pointer hit testing,
-- text entry, activation, and accessibility state; callers remain responsible
-- for rendering control visuals in their on_state/on_change callbacks.
local input = require("dm.input/v1")

local M = {}
local Manager = {}
Manager.__index = Manager

local function eligible(manager, control)
    return control.focusable and control.enabled and control.visible
        and control.scope == manager.active_scope
end

local function contains(control, x, y)
    return control.visible ~= false and control.enabled ~= false
        and control.scope == control.manager.active_scope
        and x >= control.x and y >= control.y
        and x < control.x + control.width and y < control.y + control.height
end

local function state_of(manager, control, x, y)
    if control.visible == false then return "hidden" end
    if control.enabled == false then return "disabled" end
    if manager.pressed == control then return "pressed" end
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
    control.scope = control.scope or self.active_scope
    control.manager = self
    control.state = control.enabled and "normal" or "disabled"
    self.controls[#self.controls + 1] = control
    self.by_id[control.id] = control
    if self.focus == nil and eligible(self, control) then self.focus = control end
    return control
end

function Manager:add_checkbox(control)
    control.role = "checkbox"
    control.checked = control.checked == true
    local activate = control.on_activate
    control.on_activate = function(current)
        current.checked = not current.checked
        if current.on_change then current.on_change(current, current.checked) end
        if activate then activate(current) end
    end
    return self:add(control)
end

local function character_count(value)
    local count = 0
    for index = 1, #value do
        local byte = value:byte(index)
        if byte < 128 or byte >= 192 then count = count + 1 end
    end
    return count
end

local function byte_at_character(value, character)
    if character <= 0 then return 1 end
    local seen = 0
    for index = 1, #value do
        local byte = value:byte(index)
        if byte < 128 or byte >= 192 then
            if seen == character then return index end
            seen = seen + 1
        end
    end
    return #value + 1
end

local function first_characters(value, count)
    if count <= 0 then return "" end
    local seen = 0
    for index = 1, #value do
        local byte = value:byte(index)
        if byte < 128 or byte >= 192 then
            seen = seen + 1
            if seen > count then return value:sub(1, index - 1) end
        end
    end
    return value
end

local function insert_at(value, cursor, inserted)
    local byte = byte_at_character(value, cursor)
    return value:sub(1, byte - 1) .. inserted .. value:sub(byte)
end

local function delete_before(value, cursor)
    if cursor <= 0 then return value, cursor end
    local first = byte_at_character(value, cursor - 1)
    local last = byte_at_character(value, cursor)
    return value:sub(1, first - 1) .. value:sub(last), cursor - 1
end

local function delete_at(value, cursor)
    if cursor >= character_count(value) then return value end
    local first = byte_at_character(value, cursor)
    local last = byte_at_character(value, cursor + 1)
    return value:sub(1, first - 1) .. value:sub(last)
end

function Manager:add_text_field(control)
    control.role = "textbox"
    control.value = control.value or ""
    control.max_length = control.max_length or math.huge
    control.cursor = math.max(0, math.min(character_count(control.value), control.cursor or character_count(control.value)))
    return self:add(control)
end

function Manager:add_scrollbar(control)
    control.role = "scrollbar"
    control.min = control.min or 0
    control.max = control.max or 1
    assert(control.max >= control.min, "scrollbar max must be at least min")
    control.step = control.step or 1
    control.orientation = control.orientation or "horizontal"
    assert(control.orientation == "horizontal" or control.orientation == "vertical", "invalid scrollbar orientation")
    control.value = math.max(control.min, math.min(control.max, control.value or control.min))
    return self:add(control)
end

function Manager:get(id) return self.by_id[id] end

function Manager:set_enabled(id, enabled)
    local control = assert(self.by_id[id], "unknown control: " .. id)
    control.enabled = enabled == true
    if self.focus == control and not control.enabled then self:move_focus(1) end
end

function Manager:set_scope(scope)
    assert(type(scope) == "string" and scope ~= "", "focus scope is required")
    self.active_scope = scope
    if not self.focus or not eligible(self, self.focus) then
        self.focus = nil
        self:move_focus(1)
    end
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
        if eligible(self, candidate) then self.focus = candidate return candidate end
    end
    self.focus = nil
    return nil
end

local function set_value(control, value)
    value = math.max(control.min, math.min(control.max, value))
    if value == control.value then return end
    control.value = value
    if control.on_change then control.on_change(control, value) end
end

function Manager:activate(control)
    if not control or not control.enabled or not control.visible then return false end
    if control.on_activate then control.on_activate(control) end
    return true
end

local function edit_text(control)
    local old = control.value
    local entered = input.text()
    if entered ~= "" then
        if control.filter then entered = control.filter(entered) or "" end
        local available = control.max_length - character_count(control.value)
        if available > 0 and entered ~= "" then
            entered = first_characters(entered, available)
            control.value = insert_at(control.value, control.cursor, entered)
            control.cursor = control.cursor + character_count(entered)
        end
    end
    if input.pressed("backspace") then
        control.value, control.cursor = delete_before(control.value, control.cursor)
    end
    if input.pressed("delete") then control.value = delete_at(control.value, control.cursor) end
    if input.pressed("home") then control.cursor = 0 end
    if input.pressed("end") then control.cursor = character_count(control.value) end
    if input.pressed("left") then control.cursor = math.max(0, control.cursor - 1) end
    if input.pressed("right") then control.cursor = math.min(character_count(control.value), control.cursor + 1) end
    if old ~= control.value and control.on_change then control.on_change(control, control.value) end
end

function Manager:update()
    if input.pressed("down") then self:move_focus(1) end
    if input.pressed("up") then self:move_focus(-1) end

    if self.focus and self.focus.role == "scrollbar" then
        if input.pressed("right") then set_value(self.focus, self.focus.value + self.focus.step) end
        if input.pressed("left") then set_value(self.focus, self.focus.value - self.focus.step) end
    elseif self.focus and self.focus.role ~= "textbox" then
        if input.pressed("right") then self:move_focus(1) end
        if input.pressed("left") then self:move_focus(-1) end
    end

    local x, y = input.cursor()
    local hovered = nil
    for _, control in ipairs(self.controls) do
        if contains(control, x, y) then hovered = control end
    end
    if hovered and eligible(self, hovered) then self.focus = hovered end

    self.pressed = nil
    if hovered and input.down("pointer_primary") then
        self.pressed = hovered
    elseif self.focus and input.down("confirm") then
        self.pressed = self.focus
    end

    if input.pressed("pointer_primary") then
        if hovered and hovered.role == "scrollbar" then
            local fraction
            if hovered.orientation == "vertical" then
                fraction = (y - hovered.y) / hovered.height
            else
                fraction = (x - hovered.x) / hovered.width
            end
            set_value(hovered, hovered.min + (hovered.max - hovered.min) * fraction)
        else
            self:activate(hovered)
        end
    end
    if input.pressed("confirm") then self:activate(self.focus) end

    if self.focus and self.focus.role == "textbox" then edit_text(self.focus) end

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
            scope = control.scope,
            checked = control.checked,
            value = control.value,
            cursor = control.cursor,
            min = control.min,
            max = control.max,
        }
    end
    return result
end

function M.new()
    return setmetatable({ controls = {}, by_id = {}, focus = nil, active_scope = "default" }, Manager)
end

return M
