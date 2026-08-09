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
    if control.visible == false or control.enabled == false
        or control.scope ~= control.manager.active_scope then
        return false
    end
    if control.hit_test then
        return control.hit_test(control, x, y) == true
    end
    return x >= control.x and y >= control.y
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

local function is_range(control)
    return control and (control.role == "scrollbar" or control.role == "slider")
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

local function prepare_range(control, role)
    control.role = role
    control.min = control.min or 0
    control.max = control.max or 1
    assert(control.max >= control.min, role .. " max must be at least min")
    control.step = control.step or 1
    control.orientation = control.orientation or "horizontal"
    assert(control.orientation == "horizontal" or control.orientation == "vertical", "invalid " .. role .. " orientation")
    control.value = math.max(control.min, math.min(control.max, control.value or control.min))
    return control
end

function Manager:add_scrollbar(control)
    return self:add(prepare_range(control, "scrollbar"))
end

function Manager:add_slider(control)
    return self:add(prepare_range(control, "slider"))
end

function Manager:get(id) return self.by_id[id] end

function Manager:set_enabled(id, enabled)
    local control = assert(self.by_id[id], "unknown control: " .. id)
    control.enabled = enabled == true
    if self.focus == control and not control.enabled then self:move_focus(1) end
    if self.pointer_capture == control and not control.enabled then self.pointer_capture = nil end
end

function Manager:set_visible(id, visible)
    local control = assert(self.by_id[id], "unknown control: " .. id)
    control.visible = visible == true
    if self.focus == control and not control.visible then self:move_focus(1) end
    if self.pointer_capture == control and not control.visible then self.pointer_capture = nil end
end

function Manager:set_focus(target)
    local control = target
    if type(target) == "string" then control = assert(self.by_id[target], "unknown control: " .. target) end
    if control ~= nil and not eligible(self, control) then return false end
    self.focus = control
    return true
end

function Manager:set_scope(scope)
    assert(type(scope) == "string" and scope ~= "", "focus scope is required")
    self.active_scope = scope
    self.pointer_capture = nil
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

local function set_value(control, value, snap)
    value = math.max(control.min, math.min(control.max, value))
    if snap and control.step and control.step > 0 then
        local steps = math.floor(((value - control.min) / control.step) + 0.5)
        value = math.min(control.max, control.min + steps * control.step)
    end
    if value == control.value then return false end
    control.value = value
    if control.on_change then control.on_change(control, value) end
    return true
end

local function set_pointer_value(control, x, y)
    if control.pointer_to_value then
        set_value(control, control.pointer_to_value(control, x, y), true)
        return
    end
    local fraction
    if control.orientation == "vertical" then
        fraction = (y - control.y) / control.height
    else
        fraction = (x - control.x) / control.width
    end
    set_value(control, control.min + (control.max - control.min) * fraction, true)
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
    local focused_range = is_range(self.focus)
    local adjustable_range = focused_range and self.focus.max > self.focus.min
    if adjustable_range and self.focus.orientation == "vertical" then
        if input.pressed("down") then set_value(self.focus, self.focus.value + self.focus.step) end
        if input.pressed("up") then set_value(self.focus, self.focus.value - self.focus.step) end
    else
        if input.pressed("down") then self:move_focus(1) end
        if input.pressed("up") then self:move_focus(-1) end
    end

    if adjustable_range and self.focus.orientation == "horizontal" then
        if input.pressed("right") then set_value(self.focus, self.focus.value + self.focus.step) end
        if input.pressed("left") then set_value(self.focus, self.focus.value - self.focus.step) end
    elseif self.focus and self.focus.role ~= "textbox" and (not focused_range or not adjustable_range) then
        if input.pressed("right") then self:move_focus(1) end
        if input.pressed("left") then self:move_focus(-1) end
    end

    local x, y = input.cursor()
    local hovered = nil
    local hovered_priority = -math.huge
    for index, control in ipairs(self.controls) do
        if contains(control, x, y) then
            -- Explicit visual priority lets overlapping character-create actor
            -- bounds follow their authored draw order. Otherwise preserve the
            -- historical last-added-wins behavior by using the list index.
            local priority = control.hit_priority or index
            if priority >= hovered_priority then
                hovered = control
                hovered_priority = priority
            end
        end
    end
    if hovered and eligible(self, hovered) then self.focus = hovered end

    -- Diablo II buttons capture on mouse-down but dispatch their action only on
    -- mouse-up, and only when the release remains inside the same control. Range
    -- controls retain capture while dragging so their thumb can track the mouse.
    if input.pressed("pointer_primary") then
        self.pointer_capture = hovered
    end

    if self.pointer_capture and is_range(self.pointer_capture) and input.down("pointer_primary") then
        set_pointer_value(self.pointer_capture, x, y)
    end

    self.pressed = nil
    if self.pointer_capture and input.down("pointer_primary") then
        self.pressed = self.pointer_capture
    elseif self.focus and input.down("confirm") then
        self.pressed = self.focus
    end

    if input.released("pointer_primary") then
        local captured = self.pointer_capture
        self.pointer_capture = nil
        self.pressed = nil
        if captured and contains(captured, x, y) then
            if is_range(captured) then
                set_pointer_value(captured, x, y)
            else
                self:activate(captured)
            end
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
    return setmetatable({
        controls = {},
        by_id = {},
        focus = nil,
        pressed = nil,
        pointer_capture = nil,
        active_scope = "default",
    }, Manager)
end

return M