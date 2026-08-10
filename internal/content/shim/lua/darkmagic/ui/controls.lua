-- Reusable retained-mode control manager for shim and mod-authored screens.
--
-- THIS FILE IS THE INPUT BRAIN FOR MOST SHIM WIDGETS.
--
-- The important design is that a "button" is not a mysterious native object.
-- A control is just a Lua table describing:
--   * where it is;
--   * whether it can be used;
--   * what role it has;
--   * callbacks to run when something interesting happens.
--
-- The manager owns interaction rules: focus, pointer hit testing, text entry,
-- activation, ranges, pointer capture, and accessibility state. A visual widget
-- such as button.lua or slider.lua owns only how those states LOOK.
--
-- That separation means a mod can reuse the same interaction rules with totally
-- different art.

-- Versioned engine input capability. It reports logical actions such as
-- "confirm" or "left" instead of forcing this Lua to know physical key codes.
local input = require("dm.input/v1")

-- `M` is the public module table returned at the bottom of the file.
local M = {}

-- `Manager` is the table used as the method prototype for manager instances.
local Manager = {}

-- This is Lua's lightweight object pattern. When `instance.some_method` is not
-- found directly on the instance, Lua looks in Manager because of __index.
Manager.__index = Manager

-- A control is eligible for keyboard/controller focus only when all four rules
-- are true. Breaking the condition across lines is just for readability.
local function eligible(manager, control)
    return control.focusable and control.enabled and control.visible
        and control.scope == manager.active_scope
end

-- Hit testing answers one question: "does point (x,y) belong to this control?"
local function contains(control, x, y)
    -- Hidden/disabled controls and controls from another modal focus scope are
    -- intentionally treated as if they are not under the pointer at all.
    if control.visible == false or control.enabled == false
        or control.scope ~= control.manager.active_scope then
        return false
    end

    -- Most controls are rectangles, but a caller may provide a custom test for
    -- irregular art or overlapping actors. Only literal true counts as a hit.
    if control.hit_test then
        return control.hit_test(control, x, y) == true
    end

    -- Rectangles are half-open: left/top edges count, right/bottom edges do not.
    -- That avoids two neighboring cells both claiming the exact shared border.
    return x >= control.x and y >= control.y
        and x < control.x + control.width and y < control.y + control.height
end

-- Convert all interaction facts into one small semantic visual state string.
local function state_of(manager, control, x, y)
    -- Order matters. A hidden control should not also look disabled, and a
    -- currently pressed control wins over hover/focus styling.
    if control.visible == false then return "hidden" end
    if control.enabled == false then return "disabled" end
    if manager.pressed == control then return "pressed" end
    if contains(control, x, y) then return "hover" end
    if manager.focus == control then return "focused" end
    return "normal"
end

-- Sliders and scrollbars share "range" behavior. Returning `control and ...`
-- safely produces nil/false if there is no focused control.
local function is_range(control)
    return control and (control.role == "scrollbar" or control.role == "slider")
end

function Manager:add(control)
    -- Fail early when a mod constructs an impossible control. These assertions
    -- turn later mysterious input bugs into immediate, source-local errors.
    assert(type(control.id) == "string" and control.id ~= "", "control id is required")
    assert(type(control.x) == "number" and type(control.y) == "number", "control position is required")
    assert(type(control.width) == "number" and control.width > 0, "positive control width is required")
    assert(type(control.height) == "number" and control.height > 0, "positive control height is required")
    assert(self.by_id[control.id] == nil, "duplicate control id: " .. control.id)

    -- Apply defaults. `a or b` means "use a if it has a value, otherwise b."
    control.role = control.role or "button"

    -- These three use `~= false` so omitted/nil means enabled/visible/focusable.
    -- That is different from `== true`, which would make omission mean false.
    control.enabled = control.enabled ~= false
    control.visible = control.visible ~= false
    control.focusable = control.focusable ~= false

    -- Scopes isolate modal groups. A dialog can use scope="dialog" so arrows
    -- cannot wander into controls visually sitting behind the dialog.
    control.scope = control.scope or self.active_scope

    -- Give helper functions a path back to the manager that owns this control.
    control.manager = self

    -- Lua's `condition and a or b` is a compact conditional expression here.
    control.state = control.enabled and "normal" or "disabled"

    -- Lua arrays are 1-based. `#table + 1` appends a new item.
    self.controls[#self.controls + 1] = control

    -- Keep a second lookup table so callers can find a control by ID instantly.
    self.by_id[control.id] = control

    -- The first eligible control automatically becomes initial focus.
    if self.focus == nil and eligible(self, control) then self.focus = control end
    return control
end

function Manager:add_checkbox(control)
    -- Checkbox behavior is just a small adapter around an ordinary control.
    control.role = "checkbox"
    control.checked = control.checked == true

    -- Save any caller-supplied activation callback before replacing it.
    local activate = control.on_activate
    control.on_activate = function(current)
        -- Toggle the value first...
        current.checked = not current.checked
        -- ...tell presentation/game code about the new value...
        if current.on_change then current.on_change(current, current.checked) end
        -- ...then preserve the caller's original activation hook if it had one.
        if activate then activate(current) end
    end
    return self:add(control)
end

-- The next helpers look odd because Lua strings are byte strings. UTF-8
-- characters may use several bytes. A text cursor should move by CHARACTERS,
-- not accidentally land in the middle of a multi-byte character.
--
-- UTF-8 continuation bytes are 128..191. A byte outside that range starts a new
-- character, which is enough for the cursor bookkeeping used here.
local function character_count(value)
    local count = 0
    for index = 1, #value do
        local byte = value:byte(index)
        if byte < 128 or byte >= 192 then count = count + 1 end
    end
    return count
end

-- Convert a human character position (0,1,2,...) to Lua's 1-based byte index.
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
    -- One byte past the string is the insertion point "at the end."
    return #value + 1
end

-- Keep only the first N UTF-8 characters without slicing through a character.
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
    -- Split the string at the cursor byte and stitch left + inserted + right.
    local byte = byte_at_character(value, cursor)
    return value:sub(1, byte - 1) .. inserted .. value:sub(byte)
end

local function delete_before(value, cursor)
    if cursor <= 0 then return value, cursor end
    -- Find the byte span belonging to the previous full UTF-8 character.
    local first = byte_at_character(value, cursor - 1)
    local last = byte_at_character(value, cursor)
    -- Return TWO values: the new string and the cursor moved one character left.
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
    -- `math.huge` acts like "no practical maximum" when max_length is omitted.
    control.max_length = control.max_length or math.huge

    -- Clamp the starting cursor into the actual character range. This nested
    -- min/max is a common clamp pattern: never below 0, never past text length.
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

    -- Clamp the starting value into [min,max].
    control.value = math.max(control.min, math.min(control.max, control.value or control.min))
    return control
end

function Manager:add_scrollbar(control)
    return self:add(prepare_range(control, "scrollbar"))
end

function Manager:add_slider(control)
    return self:add(prepare_range(control, "slider"))
end

-- Tiny methods can stay one line when the meaning is obvious from the name.
function Manager:get(id) return self.by_id[id] end

function Manager:set_enabled(id, enabled)
    local control = assert(self.by_id[id], "unknown control: " .. id)
    control.enabled = enabled == true

    -- Focus/capture must never remain stuck on a control that can no longer act.
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
    -- Friendly API: callers may pass the control table OR its string ID.
    if type(target) == "string" then control = assert(self.by_id[target], "unknown control: " .. target) end
    -- Returning false instead of forcing invalid focus lets callers recover.
    if control ~= nil and not eligible(self, control) then return false end
    self.focus = control
    return true
end

function Manager:set_scope(scope)
    assert(type(scope) == "string" and scope ~= "", "focus scope is required")
    self.active_scope = scope

    -- A drag/click that began in an old modal scope must not leak into the new one.
    self.pointer_capture = nil

    -- If old focus is no longer legal, choose the first legal control in scope.
    if not self.focus or not eligible(self, self.focus) then
        self.focus = nil
        self:move_focus(1)
    end
end

function Manager:move_focus(delta)
    if #self.controls == 0 then self.focus = nil return nil end

    -- Some menus intentionally stop at their first/last entry instead of wrapping.
    if self.wrap_focus == false then
        local focusable = {}
        local current = 0

        -- Build a temporary list containing only controls focus can actually use.
        for _, control in ipairs(self.controls) do
            if eligible(self, control) then
                focusable[#focusable + 1] = control
                if control == self.focus then current = #focusable end
            end
        end

        if #focusable == 0 then self.focus = nil return nil end

        if current == 0 then
            -- If nothing was focused, choose an end based on travel direction.
            current = delta < 0 and #focusable or 1
        else
            -- Clamp rather than wrap.
            current = math.max(1, math.min(#focusable, current + delta))
        end
        self.focus = focusable[current]
        return self.focus
    end

    -- Wrapping mode searches the original control list. First find the numeric
    -- index of current focus (or leave start=0 if there isn't one).
    local start = 0
    for index, control in ipairs(self.controls) do
        if control == self.focus then start = index break end
    end

    -- Search at most one full circuit. The modulo expression turns values that
    -- run past either end back into Lua's 1..#controls range.
    for step = 1, #self.controls do
        local index = ((start - 1 + delta * step) % #self.controls) + 1
        local candidate = self.controls[index]
        if eligible(self, candidate) then self.focus = candidate return candidate end
    end

    self.focus = nil
    return nil
end

local function set_value(control, value, snap)
    -- Range controls never store a value outside their declared bounds.
    value = math.max(control.min, math.min(control.max, value))

    if snap and control.step and control.step > 0 then
        -- Convert distance from min into a step number and round to nearest step.
        local steps = math.floor(((value - control.min) / control.step) + 0.5)
        value = math.min(control.max, control.min + steps * control.step)
    end

    -- Avoid useless redraws/callbacks if nothing actually changed.
    if value == control.value then return false end
    control.value = value
    if control.on_change then control.on_change(control, value) end
    return true
end

local function set_pointer_value(control, x, y)
    if control.pointer_to_value then
        -- A visual widget can define custom geometry (for example accounting for
        -- thumb size). The manager still owns clamping/snapping/callback rules.
        set_value(control, control.pointer_to_value(control, x, y), true)
        return
    end

    -- Generic fallback maps pointer position to a 0..1 fraction of the track.
    local fraction
    if control.orientation == "vertical" then
        fraction = (y - control.y) / control.height
    else
        fraction = (x - control.x) / control.width
    end
    set_value(control, control.min + (control.max - control.min) * fraction, true)
end

function Manager:activate(control)
    -- Safe no-op for missing/disabled/hidden targets. This makes keyboard and
    -- pointer code able to call one helper without repeating all three guards.
    if not control or not control.enabled or not control.visible then return false end
    if control.on_activate then control.on_activate(control) end
    return true
end

local function edit_text(control)
    -- Remember the previous value so on_change only fires when text changed.
    local old = control.value

    -- `input.text()` returns text typed this frame, separate from key actions.
    local entered = input.text()
    if entered ~= "" then
        -- Optional filters can reject/transform input (character-name rules etc.).
        if control.filter then entered = control.filter(entered) or "" end

        -- Max length is measured in characters, not UTF-8 bytes.
        local available = control.max_length - character_count(control.value)
        if available > 0 and entered ~= "" then
            entered = first_characters(entered, available)
            control.value = insert_at(control.value, control.cursor, entered)
            control.cursor = control.cursor + character_count(entered)
        end
    end

    -- Lua can receive multiple return values directly into multiple variables.
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
    -- FIRST: keyboard/controller navigation and range adjustment.
    local focused_range = is_range(self.focus)
    local adjustable_range = focused_range and self.focus.max > self.focus.min

    -- Vertical range controls consume Up/Down to adjust the value. Otherwise
    -- those same actions move focus through ordinary controls.
    if adjustable_range and self.focus.orientation == "vertical" then
        if input.pressed("down") then set_value(self.focus, self.focus.value + self.focus.step) end
        if input.pressed("up") then set_value(self.focus, self.focus.value - self.focus.step) end
    else
        if input.pressed("down") then self:move_focus(1) end
        if input.pressed("up") then self:move_focus(-1) end
    end

    -- Horizontal range controls similarly consume Left/Right. Textboxes keep
    -- Left/Right for their own cursor editing later in this function.
    if adjustable_range and self.focus.orientation == "horizontal" then
        if input.pressed("right") then set_value(self.focus, self.focus.value + self.focus.step) end
        if input.pressed("left") then set_value(self.focus, self.focus.value - self.focus.step) end
    elseif self.focus and self.focus.role ~= "textbox" and (not focused_range or not adjustable_range) then
        if input.pressed("right") then self:move_focus(1) end
        if input.pressed("left") then self:move_focus(-1) end
    end

    -- SECOND: find what the pointer is over.
    local x, y = input.cursor()
    local hovered = nil
    local hovered_priority = -math.huge
    for index, control in ipairs(self.controls) do
        if contains(control, x, y) then
            -- Explicit visual priority lets overlapping character-create actor
            -- bounds follow authored draw order. Without a custom value, the
            -- later-added control wins because its list index is larger.
            local priority = control.hit_priority or index
            if priority >= hovered_priority then
                hovered = control
                hovered_priority = priority
            end
        end
    end

    -- Pointer hover also moves focus, which keeps mouse and keyboard state in sync.
    if hovered and eligible(self, hovered) then self.focus = hovered end

    -- THIRD: pointer capture.
    --
    -- Diablo II-style buttons depress on mouse-down but ACTIVATE on mouse-up,
    -- only if release is still inside the same control. Remembering the pressed
    -- control is called pointer capture.
    if input.pressed("pointer_primary") then
        self.pointer_capture = hovered
    end

    -- Range controls are special: while held, they continue tracking the pointer
    -- even if it moves beyond the original hit rectangle.
    if self.pointer_capture and is_range(self.pointer_capture) and input.down("pointer_primary") then
        set_pointer_value(self.pointer_capture, x, y)
    end

    -- `pressed` is visual state for this frame. Reset it, then derive it from
    -- either pointer capture or a held keyboard/controller Confirm action.
    self.pressed = nil
    if self.pointer_capture and input.down("pointer_primary") then
        self.pressed = self.pointer_capture
    elseif self.focus and input.down("confirm") then
        self.pressed = self.focus
    end

    -- Release completes a click/drag transaction.
    if input.released("pointer_primary") then
        local captured = self.pointer_capture
        self.pointer_capture = nil
        self.pressed = nil

        -- Ordinary buttons require the release to remain inside. Ranges get one
        -- final exact value sample before release completes.
        if captured and contains(captured, x, y) then
            if is_range(captured) then
                set_pointer_value(captured, x, y)
            else
                self:activate(captured)
            end
        end
    end

    -- Keyboard/controller confirm activates the focused control on press.
    if input.pressed("confirm") then self:activate(self.focus) end

    -- Text editing runs only for a focused textbox, so normal Left/Right
    -- navigation cannot accidentally edit an unfocused field.
    if self.focus and self.focus.role == "textbox" then edit_text(self.focus) end

    -- LAST: turn all the interaction facts into semantic visual states. Only
    -- call `on_state` when the state changes; retained widgets need not redraw
    -- identical art every frame.
    for _, control in ipairs(self.controls) do
        local next_state = state_of(self, control, x, y)
        if control.state ~= next_state then
            control.state = next_state
            if control.on_state then control.on_state(control, next_state) end
        end
    end
end

function Manager:accessibility()
    -- Produce a VALUE-ONLY snapshot. Callers receive descriptive facts, not
    -- mutable references to the manager's internal ownership state.
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

function M.new(options)
    options = options or {}

    -- `setmetatable(instance, Manager)` plus Manager.__index = Manager gives the
    -- returned plain table methods such as `manager:add(...)` and `:update()`.
    return setmetatable({
        -- Ordered list supports deterministic focus traversal and hit priority.
        controls = {},
        -- ID lookup supports convenient direct access.
        by_id = {},
        focus = nil,
        pressed = nil,
        pointer_capture = nil,
        active_scope = options.active_scope or "default",
        -- Omitted means wrap; only explicit false disables wrapping.
        wrap_focus = options.wrap_focus ~= false,
    }, Manager)
end

return M