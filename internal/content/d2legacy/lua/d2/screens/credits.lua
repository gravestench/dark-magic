-- Localized expansion credits viewer.
--
-- This scene shows how a mod can READ optional user-owned game data through the
-- virtual filesystem capability without embedding that proprietary text in the
-- Dark Magic repository.
--
-- Lua receives text, normalizes line endings, converts a tiny original formatting
-- convention into semantic color tags, and paginates it. Filesystem mounting,
-- archive provenance, and real host paths remain engine-owned behind engine.vfs/v1.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")
local locale = require("engine.locale/v1")
local vfs = require("engine.vfs/v1")
local dc6 = require("d2.ui.dc6")
local cursor = require("d2.ui.cursor")
local styled_text = require("d2.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2.presentation/v1"))
local screen = manifest.screens.credits

-- Split text into deterministic fixed-line pages.
local function pages(text, count)
    -- Three locals in one statement: completed pages, current page lines, line count.
    local result, page, lines = {}, {}, 0

    -- Windows text often uses CRLF (\r\n); older data may contain lone CR.
    -- Normalize both forms to plain `\n` so the loop below has one separator.
    text = text:gsub("\r\n", "\n"):gsub("\r", "\n")

    -- Appending one newline ensures the pattern also sees the final line when
    -- the source file does not end with a newline.
    for line in (text .. "\n"):gmatch("(.-)\n") do
        -- Leading `*` is an original formatting marker, not visible text.
        if line:sub(1, 1) == "*" then
            line = "[red]" .. line:sub(2)
        elseif line ~= "" then
            line = "[gold]" .. line
        end

        -- Lua can assign array append AND increment in one multi-assignment.
        page[#page + 1], lines = line, lines + 1

        if lines == count then
            -- Finish this page, then reset current-page array and line counter.
            result[#result + 1], page, lines = table.concat(page, "\n"), {}, 0
        end
    end

    -- If source did not divide evenly into pages, keep the final partial page.
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

        -- VFS path is a logical content path, not a host OS path handed to Lua.
        local text = vfs.read_text(screen.text)
        if not text then
            -- Missing optional credits data gets a localized friendly fallback.
            text = assert(locale.text("d2.credits.unavailable"))
        end

        self.pages, self.page = pages(text, screen.lines_per_page), 1

        if render.assets_available() then
            self.text = render.create("hud", self.root)
            self:draw_page()
        end

        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,

    draw_page = function(self)
        -- Current numeric page selects one prebuilt string from the page array.
        styled_text.set(self.text, screen.text_style, self.pages[self.page], 700, "center")
        self.text:set_position(400, 300)
    end,

    update = function(self)
        self.cursor:update()

        -- Remember prior page so retained text is rebuilt only if navigation changed it.
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
