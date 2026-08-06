-- Localized expansion credits viewer.
--
-- Credit text remains in the user's game data. The shim normalizes its line
-- endings and paginates it without copying proprietary text into this project.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local vfs = require("dm.vfs/v1")
local dc6 = require("darkmagic.ui.dc6")
local cursor = require("darkmagic.ui.cursor")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen, font = manifest.screens.credits, manifest.fonts.exocet10

-- Split text into deterministic, fixed-line pages. A leading asterisk is an
-- original formatting marker and is not shown to the player.
local function pages(text, count)
    local result, page, lines = {}, {}, 0
    text = text:gsub("\r\n", "\n"):gsub("\r", "\n")
    for line in (text .. "\n"):gmatch("(.-)\n") do
        if line:sub(1, 1) == "*" then
            line = line:sub(2)
        end
        page[#page + 1], lines = line, lines + 1
        if lines == count then
            result[#result + 1], page, lines = table.concat(page, "\n"), {}, 0
        end
    end
    if #page > 0 then
        result[#result + 1] = table.concat(page, "\n")
    end
    return result
end

return {
    create = function(self)
        self.root = render.create("hud")
        self.background = dc6.frontend_background(
            self.root,
            "hud",
            screen.background,
            manifest.palettes[screen.palette],
            manifest.layouts.frontend_tiles
        )
        local text = vfs.read_text(screen.text)
        if not text then
            text = assert(locale.text("darkmagic.credits.unavailable"))
        end
        self.pages, self.page = pages(text, screen.lines_per_page), 1
        if render.assets_available() then
            self.text = render.create("hud", self.root)
            self:draw_page()
        end
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,
    draw_page = function(self)
        self.text:set_text(
            font.table,
            font.sheet,
            manifest.palettes[font.palette],
            self.pages[self.page],
            {
                red = 205,
                green = 190,
                blue = 155,
                max_width = 700,
                align = "center",
            }
        )
        self.text:set_position(400, 300)
    end,
    update = function(self)
        self.cursor:update()
        local previous = self.page
        if input.pressed("down") or input.pressed("right") or input.pressed("confirm") then
            self.page = math.min(#self.pages, self.page + 1)
        end
        if input.pressed("up") or input.pressed("left") then
            self.page = math.max(1, self.page - 1)
        end
        if self.text and previous ~= self.page then
            self:draw_page()
        end
        if input.pressed("cancel") then
            scenes.replace("main_menu")
        end
    end,
}
