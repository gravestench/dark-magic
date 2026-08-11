-- Reusable keyboard-first modal for choosing one value from a large catalog.
-- Labs supply plain strings and one callback; this module owns text entry,
-- subsequence ranking, focus, modal drawing, and input capture.

local render = require("dm.render/v1")
local input = require("dm.input/v1")
local text = require("darkmagic.ui.text")

local M = {}
local visible_rows = 9

local function fuzzy_score(value, query)
    value, query = value:lower(), query:lower()
    if query == "" then return 0 end
    local position, score, previous = 1, 0, 0
    for index = 1, #query do
        local found = value:find(query:sub(index, index), position, true)
        if not found then return nil end
        score = score + 20 - math.min(found, 19)
        if previous > 0 and found == previous + 1 then score = score + 12 end
        previous, position = found, found + 1
    end
    local exact = value:find(query, 1, true)
    if exact then score = score + 100 - math.min(exact, 99) end
    return score
end

local function ranked(items, query)
    local result = {}
    for _, value in ipairs(items) do
        local score = fuzzy_score(value, query)
        if score ~= nil then result[#result + 1] = {value=value, score=score} end
    end
    table.sort(result, function(left, right)
        if left.score ~= right.score then return left.score > right.score end
        return left.value:lower() < right.value:lower()
    end)
    return result
end

function M.create(root, options)
    options = options or {}
    local picker = {items=assert(options.items), on_select=assert(options.on_select), open=false, query="", selected=1}
    picker.root = render.create("modal", root)
    picker.backdrop = render.create("modal", picker.root)
    picker.backdrop:fill_rect(680, 450, 0, 0, 0, 224)
    picker.backdrop:set_position(400, 300)
    picker.title = render.create("modal", picker.root)
    picker.query_node = render.create("modal", picker.root)
    picker.help = render.create("modal", picker.root)
    text.set(picker.title, "font_lab_heading", options.title or "SELECT", 620, "center")
    picker.title:set_position(400, 115)
    text.set(picker.help, "font_lab_caption", "Type to filter   Up/Down: choose   Enter: select   Esc: close", 620, "center")
    picker.help:set_position(400, 500)
    picker.rows = {}
    for index = 1, visible_rows do
        picker.rows[index] = render.create("modal", picker.root)
        picker.rows[index]:set_position(400, 190 + (index - 1) * 30)
    end

    function picker:set_visible(visible)
        self.root:set_visible(visible)
    end

    function picker:refresh()
        self.matches = ranked(self.items, self.query)
        self.selected = math.max(1, math.min(self.selected, math.max(1, #self.matches)))
        text.set(self.query_node, "font_lab_color", "[gold]>[/] [white]" .. self.query .. "_[/]", 620, "left")
        self.query_node:set_position(400, 155)
        local first = math.max(1, self.selected - visible_rows + 1)
        for row = 1, visible_rows do
            local match = self.matches[first + row - 1]
            local value = match and match.value or ""
            local color = first + row - 1 == self.selected and "[green]" or "[white]"
            text.set(self.rows[row], "font_lab_color", color .. value .. "[/]", 620, "left")
        end
    end

    function picker:show()
        self.open, self.query, self.selected = true, "", 1
        self:refresh()
        self:set_visible(true)
    end

    function picker:close()
        self.open = false
        self:set_visible(false)
    end

    function picker:update()
        if not self.open then return false end
        local entered = input.text()
        if entered ~= "" then
            self.query, self.selected = self.query .. entered, 1
            self:refresh()
        end
        if input.pressed("backspace") and #self.query > 0 then
            self.query, self.selected = self.query:sub(1, #self.query - 1), 1
            self:refresh()
        end
        if input.pressed("up") then self.selected = math.max(1, self.selected - 1); self:refresh() end
        if input.pressed("down") then self.selected = math.min(#self.matches, self.selected + 1); self:refresh() end
        if input.pressed("cancel") then self:close(); return true end
        if input.pressed("confirm") and self.matches[self.selected] then
            local value = self.matches[self.selected].value
            self:close()
            self.on_select(value)
        end
        return true
    end

    picker:set_visible(false)
    return picker
end

return M
