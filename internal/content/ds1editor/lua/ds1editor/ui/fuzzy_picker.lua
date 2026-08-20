local input = require("engine.input/v1")
local render = require("engine.render/v1")
local text = require("ds1editor.ui.text")

local picker = {}
local visible_rows = 9

-- Score an ordered fuzzy match while rewarding compact and early character runs.
local function score(value, query)
    value, query = value:lower(), query:lower()
    if query == "" then return 0 end
    local position, result, previous = 1, 0, 0
    for index = 1, #query do
        local found = value:find(query:sub(index, index), position, true)
        if not found then return nil end
        result = result + 20 - math.min(found, 19)
        if previous > 0 and found == previous + 1 then result = result + 12 end
        previous, position = found, found + 1
    end
    local exact = value:find(query, 1, true)
    return result + (exact and 100 - math.min(exact, 99) or 0)
end

-- Return a stable best-first copy so filtering never mutates the caller's asset list.
local function ranked(items, query)
    local result = {}
    for _, value in ipairs(items) do
        local rank = score(value, query)
        if rank then result[#result + 1] = {value=value, score=rank} end
    end
    table.sort(result, function(left, right)
        return left.score == right.score and left.value:lower() < right.value:lower() or left.score > right.score
    end)
    return result
end

-- Construct one keyboard-driven modal picker and retain all nodes for cheap refreshes.
function picker.create(root, options)
    local width, height = options.width or 800, options.height or 600
    local modal_width, modal_height = math.min(680, width - 80), math.min(450, height - 80)
    local center_x, center_y = width / 2, height / 2
    local value = {items=assert(options.items), on_select=assert(options.on_select), open=false, query="", selected=1}
    value.root = render.create("modal", root)
    local backdrop = render.create("modal", value.root)
    backdrop:fill_rect(modal_width, modal_height, 5, 9, 14, 246)
    backdrop:set_position(center_x, center_y)
    value.title = render.create("modal", value.root)
    value.query_node = render.create("modal", value.root)
    value.help = render.create("modal", value.root)
    text.set(value.title, "font_lab_heading", options.title or "SELECT", modal_width - 40, "center")
    value.title:set_position(center_x, center_y - modal_height / 2 + 44)
    text.set(
        value.help,
        "font_lab_caption",
        "Type to filter   Up/Down choose   Enter select   Esc close",
        modal_width - 40,
        "center"
    )
    value.help:set_position(center_x, center_y + modal_height / 2 - 28)
    value.rows = {}
    for index = 1, visible_rows do
        value.rows[index] = render.create("modal", value.root)
        value.rows[index]:set_position(center_x, center_y - 105 + (index - 1) * 30)
    end
    -- Rebind the fixed visible row pool after query or keyboard selection changes.
    function value:refresh(matches)
        self.matches = matches or ranked(self.items, self.query)
        self.selected = math.max(1, math.min(self.selected, math.max(1, #self.matches)))
        text.set(
            self.query_node,
            "font_lab_color",
            "[gold]>[/] [white]" .. self.query .. "_[/]",
            modal_width - 60,
            "left"
        )
        self.query_node:set_position(center_x, center_y - 145)
        local first = math.max(1, self.selected - visible_rows + 1)
        for row = 1, visible_rows do
            local match = self.matches[first + row - 1]
            local selected = first + row - 1 == self.selected
            local color = selected and "[green]" or "[white]"
            local content = color .. (match and match.value or "") .. "[/]"
            text.set(self.rows[row], "font_lab_color", content, modal_width - 60, "left")
        end
    end
    -- Reset transient search state each time the modal becomes visible.
    function value:show()
        self.open, self.query, self.selected = true, "", 1
        self:refresh()
        self.root:set_visible(true)
    end
    -- Hide the modal without discarding its reusable retained nodes.
    function value:close()
        self.open = false
        self.root:set_visible(false)
    end
    -- Consume modal keyboard input before the underlying editor can act on it.
    function value:update()
        if not self.open then return false end
        local entered = input.text()
        if entered ~= "" then self.query, self.selected = self.query .. entered, 1 end
        if input.pressed("backspace") and #self.query > 0 then self.query, self.selected = self.query:sub(1, -2), 1 end
        self.matches = ranked(self.items, self.query)
        if input.pressed("up") then self.selected = math.max(1, self.selected - 1) end
        if input.pressed("down") then self.selected = math.min(#self.matches, self.selected + 1) end
        self:refresh(self.matches)
        if input.pressed("cancel") then self:close(); return true end
        if input.pressed("confirm") and self.matches[self.selected] then
            local selected = self.matches[self.selected].value
            self:close(); self.on_select(selected)
        end
        return true
    end
    value.root:set_visible(false)
    return value
end

return picker
