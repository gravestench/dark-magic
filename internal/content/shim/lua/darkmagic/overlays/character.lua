-- Character-stat overlay backed by an immutable engine-owned stat snapshot.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local saves = require("dm.save/v1")
local data = require("dm.data/v1")
local locale = require("dm.locale/v1")
local controls = require("darkmagic.ui.controls")
local button = require("darkmagic.ui.button")
local text = require("darkmagic.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.character
local offset_x, offset_y = screen.offset_x or 0, screen.offset_y or 0

local function dc6_at(root, sheet, palette, frame, x, y)
    local node = render.create("modal", root)
    local width, height = node:set_dc6(sheet, palette, 0, frame)
    node:set_position(x + width / 2, y + height / 2)
    return node
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
            { x = panel.x + offset_x, y = panel.y + offset_y }, { x = panel.x + offset_x + 256, y = panel.y + offset_y },
            { x = panel.x + offset_x + 256, y = panel.y + offset_y + 256 }, { x = panel.x + offset_x, y = panel.y + offset_y + 256 },
        }
        for index, frame in ipairs(panel.frames) do
            dc6_at(self.root, panel.sheet, palette, frame, positions[index].x, positions[index].y)
        end

        local heading_values = { name = character.name, class = character.class }
        for _, heading in ipairs(screen.headings) do
            text.create(self.root, screen.heading_style, heading_values[heading.field], heading.x + offset_x, heading.y + offset_y, 150)
        end
        for _, value in ipairs(screen.values) do
            text.create(self.root, screen.label_style, assert(locale.text(value.label)), value.label_x + offset_x, value.label_y + offset_y, 130)
            text.create(self.root, screen.value_style, displayed_value(character, value.field), value.x + offset_x, value.y + offset_y, 100)
        end

        local close = screen.close
        local close_placement = {
            sheet=close.sheet, palette=close.palette, up_frame=close.up_frame, down_frame=close.down_frame,
            x=close.x + offset_x, y=close.y + offset_y, width=close.width, height=close.height, label=close.label,
        }
        button.create(self.root, self.controls, "close", close_placement, assert(locale.text(close.label)), {
            layer = "modal",
            show_label = false,
            sound = manifest.sounds.button,
            tooltip = assert(locale.text(close.label)),
            on_activate = function() scenes.toggle_overlay("character", "left") end,
        })
    end,

    update = function(self)
        self.controls:update()
        if input.pressed("character") or input.pressed("cancel") then
            scenes.pop()
        end
    end,
}
