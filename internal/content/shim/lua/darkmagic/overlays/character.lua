-- Character-stat overlay backed by an immutable engine-owned stat snapshot.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local saves = require("dm.save/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.character

local function dc6_at(root, sheet, palette, frame, x, y)
    local node = render.create("modal", root)
    local width, height = node:set_dc6(sheet, palette, 0, frame)
    node:set_position(x + width / 2, y + height / 2)
    return node
end

local function text_at(root, font, text, x, y, width)
    local node = render.create("modal", root)
    local rendered_width, rendered_height = node:set_text(
        font.table,
        font.sheet,
        manifest.palettes[font.palette],
        text,
        { red = 210, green = 180, blue = 110, max_width = width or 150, align = "center" }
    )
    node:set_position(x, y + rendered_height / 2)
    return node, rendered_width, rendered_height
end

local function displayed_value(character, field)
    if field == "level" then
        return character.level
    end
    local stats = character.stats or {}
    if field == "health_pair" then
        return string.format("%d / %d", stats.health or 0, stats.max_health or 0)
    elseif field == "mana_pair" then
        return string.format("%d / %d", stats.mana or 0, stats.max_mana or 0)
    elseif field == "stamina_pair" then
        return string.format("%d / %d", stats.stamina or 0, stats.max_stamina or 0)
    end
    return stats[field] or 0
end

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("modal")
        self.controls = controls.new()
        if not render.assets_available() then
            return
        end
        local character = assert(saves.selected(), "character panel requires a selected character")
        local panel = screen.panel
        local palette = manifest.palettes[panel.palette]
        local positions = {
            { x = panel.x, y = panel.y }, { x = panel.x + 256, y = panel.y },
            { x = panel.x + 256, y = panel.y + 256 }, { x = panel.x, y = panel.y + 256 },
        }
        for index, frame in ipairs(panel.frames) do
            dc6_at(self.root, panel.sheet, palette, frame, positions[index].x, positions[index].y)
        end

        local heading_font = manifest.fonts[screen.heading_font]
        local heading_values = { name = character.name, class = character.class }
        for _, heading in ipairs(screen.headings) do
            text_at(self.root, heading_font, heading_values[heading.field], heading.x, heading.y, 150)
        end
        local label_font = manifest.fonts[screen.label_font]
        local value_font = manifest.fonts[screen.value_font]
        for _, value in ipairs(screen.values) do
            text_at(self.root, label_font, assert(locale.text(value.label)), value.label_x, value.label_y, 130)
            text_at(self.root, value_font, tostring(displayed_value(character, value.field)), value.x, value.y, 100)
        end

        local close = screen.close
        local close_node = dc6_at(self.root, close.sheet, manifest.palettes[close.palette], close.up_frame, close.x, close.y)
        self.controls:add({
            id = "close", label = assert(locale.text(close.label)),
            x = close.x, y = close.y, width = close.width, height = close.height,
            on_activate = function() scenes.pop() end,
            on_state = function(_, state)
                close_node:set_dc6(close.sheet, manifest.palettes[close.palette], 0, state == "hover" and close.down_frame or close.up_frame)
            end,
        })
    end,

    update = function(self)
        self.controls:update()
        if input.pressed("character") or input.pressed("cancel") then
            scenes.pop()
        end
    end,
}
