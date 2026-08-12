-- Interactive bitmap-font and PL2 inspection scene.
--
-- DEVELOPMENT TOOLS CAN BE NORMAL MOD SCENES.
--
-- That idea is worth noticing. Font Lab does not use a secret editor renderer or
-- a second text implementation. It uses the SAME `d2legacy.ui.text` helper and
-- retained render capability as shipping screens. If Font Lab looks wrong, the
-- real UI probably has the same bug; the tool cannot accidentally hide it behind
-- a separate preview stack.
--
-- This scene is also a friendly example of DATA-DRIVEN pages. Most of the file is
-- an array of page definitions containing title/detail plus a small draw function.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local text = require("d2legacy.ui.text")
local cursor = require("d2legacy.ui.cursor")
local data = require("engine.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))

local font_lab = {}

-- This table maps each page ROOT node to the child nodes drawn for that page.
-- It lets us hide/show already-built pages instead of rebuilding text every visit.
local page_nodes = {}

-- Draw one text node from conventional top-left box coordinates.
local function put(root, style, value, left, top, width, alignment)
    local node = render.create("hud", root)

    if page_nodes[root] then
        -- Track node only when this root belongs to one of our page caches.
        table.insert(page_nodes[root], node)
    end

    local _, height = text.set(node, style, value, width, alignment or "left")

    -- text.set produces retained node content; node positioning is center-based.
    node:set_position(left + width / 2, top + height / 2)
    return node, height
end

local function heading(root, title, detail)
    put(root, "font_lab_heading", title, 40, 20, 720, "center")
    put(root, "font_lab_caption", detail, 40, 66, 720, "center")
end

-- Each row is a PAGE DESCRIPTION, not a class/native screen type.
local pages = {
    {
        title = "Semantic UI styles",
        detail = "The exact styles currently used by frontend and in-game screens",
        draw = function(root)
            -- {semantic style name, human explanation}
            local rows = {
                { "frontend_legal", "Formal12 / Static / Sky PL2 / gold" },
                { "button_normal", "Exocet10 / Units / neutral 0x646464 modulation" },
                { "character_select_title", "Font42 / Units / Sky PL2 / gold" },
                { "character_select_metadata", "Font16 / Units / Sky PL2 / white" },
                { "panel_label", "Font6 / Units / Act I PL2 / white" },
                { "character_create_heading", "Font30 / Units / Fechar PL2 / white" },
            }
            local y = 105
            for _, row in ipairs(rows) do
                put(root, "font_lab_caption", row[2], 55, y, 690, "left")
                put(root, row[1], "The quick brown fox 0123456789", 55, y + 18, 690, "left")
                y = y + 70
            end
        end,
    },
    {
        title = "Bitmap font families",
        detail = "Palette-authored baselines; Exocet10 is shown in its intended button treatment",
        draw = function(root)
            local rows = {
                { "button_normal", "Exocet10 (button)" },
                { "font_lab_font6", "Font6" },
                { "font_lab_font16", "Font16" },
                { "font_lab_font30", "Font30" },
                { "font_lab_font42", "Font42" },
                { "font_lab_formal10", "Formal10" },
                { "font_lab_formal11", "Formal11" },
                { "font_lab_formal12", "Formal12" },
            }
            local y = 102
            for _, row in ipairs(rows) do
                put(root, "font_lab_caption", row[2], 45, y, 120, "right")
                put(root, row[1], "Abc XYZ 0123", 180, y, 570, "left")
                y = y + 58
            end
        end,
    },
    {
        title = "PL2 text-color slots",
        detail = "Font16 / Units indices transformed through Sky/Pal.pl2",
        draw = function(root)
            -- Inline tags are interpreted by the shared bitmap text pipeline.
            local samples = {
                "[white]WHITE  [red]RED  [green]GREEN",
                "[blue]BLUE  [gold]GOLD  [grey]GREY",
                "[black]BLACK  [orange]ORANGE  [yellow]YELLOW",
                "[gold]Unique Item Name",
                "[blue]+20% Faster Cast Rate",
                "[green]Set Item Name",
                "[grey]Socketed Item Name",
            }
            local y = 125
            for _, sample in ipairs(samples) do
                put(root, "font_lab_color", sample, 70, y, 660, "center")
                y = y + 55
            end
        end,
    },
    {
        title = "Contextual PL2 comparison",
        detail = "Identical Font16 glyph indices and [gold] slot; only Pal.pl2 changes",
        draw = function(root)
            -- Same semantic text can look subtly different through different
            -- contextual palette transform tables.
            local rows = {
                { "font_lab_gold_sky", "Sky/Pal.pl2" },
                { "font_lab_gold_fechar", "Fechar/Pal.pl2" },
                { "font_lab_gold_act1", "Act1/Pal.pl2" },
            }
            local y = 145
            for _, row in ipairs(rows) do
                put(root, "font_lab_caption", row[2], 65, y, 180, "right")
                put(root, row[1], "GOLD: Diablo II", 270, y, 450, "left")
                y = y + 100
            end
            put(root, "font_lab_caption", "These rows should differ subtly, not lose glyph shading or edges.", 70, 470, 660, "center")
        end,
    },
    {
        title = "Alignment, wrapping, and color continuity",
        detail = "Font16 / Units / Sky PL2 inside fixed-width boxes",
        draw = function(root)
            put(root, "font_lab_caption", "LEFT", 70, 112, 200, "center")
            put(root, "font_lab_color", "[gold]Gold begins here and wraps across several words without resetting the transform.", 70, 142, 200, "left")
            put(root, "font_lab_caption", "CENTER", 300, 112, 200, "center")
            put(root, "font_lab_color", "[white]White centered text\n[red]then a red second line", 300, 142, 200, "center")
            put(root, "font_lab_caption", "RIGHT", 530, 112, 200, "center")
            put(root, "font_lab_color", "[blue]Blue right-aligned text\n[green]Green continuation", 530, 142, 200, "right")
        end,
    },
}

function font_lab.create(self)
    self.root = render.create("hud")

    -- A plain rectangle gives the lab a neutral backdrop independent of D2 UI art.
    self.background = render.create("hud", self.root)
    self.background:fill_rect(manifest.resolution.width, manifest.resolution.height, 18, 15, 13, 255)
    self.background:set_position(manifest.resolution.width / 2, manifest.resolution.height / 2)

    -- page_roots caches each page root after its first visit.
    self.page_roots = {}
    self.page = 1
    self:show_page(1)
    self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
end

function font_lab.show_page(self, requested)
    -- Wrap ANY integer request into 1..#pages. The +1 after modulo converts
    -- Lua-friendly zero-based modulo result back to one-based array indexing.
    local next_page = ((requested - 1) % #pages) + 1

    if self.page_roots[self.page] then
        -- Hide all tracked children from old cached page.
        for _, node in ipairs(page_nodes[self.page_roots[self.page]]) do
            node:set_visible(false)
        end
    end

    if not self.page_roots[next_page] then
        -- First visit: build this page ONCE.
        local root = render.create("hud", self.root)
        self.page_roots[next_page] = root
        page_nodes[root] = {}

        local page = pages[next_page]
        heading(root, page.title, page.detail)
        page.draw(root)
        put(root, "font_lab_caption", string.format("%d / %d   Left/Up: previous   Right/Down/Enter/Click: next   Esc: menu", next_page, #pages), 40, 568, 720, "center")
    end

    self.page = next_page

    -- Reveal already-built current page children.
    for _, node in ipairs(page_nodes[self.page_roots[next_page]]) do
        node:set_visible(true)
    end
end

function font_lab.update(self)
    self.cursor:update()

    -- Several logical actions intentionally mean "next" so the lab is easy to
    -- drive with mouse, keyboard, or controller.
    if input.pressed("right") or input.pressed("down") or input.pressed("confirm") or input.pressed("pointer_primary") then
        self:show_page(self.page + 1)
    elseif input.pressed("left") or input.pressed("up") then
        self:show_page(self.page - 1)
    elseif input.pressed("cancel") then
        scenes.replace("main_menu")
    end
end

return font_lab
