-- In-game help overlay assembled from recovered Diablo II presentation facts.
--
-- This file is a good example of "weird old art becomes ordinary Lua layout."
-- The DC6 border is not one convenient rectangle, so Lua explicitly places its
-- pieces, then composes normal text and a reusable close button on top.
--
-- The implementation is Dark Magic's own; recovered/reference information tells
-- us observable frame/layout facts, not which source code structure to copy.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local data = require("engine.data/v1")
local locale = require("engine.locale/v1")
local controls = require("d2legacy.ui.controls")
local button = require("d2legacy.ui.button")
local text = require("d2legacy.ui.text")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local screen = manifest.screens.help

-- Create one DC6 frame at a top-left coordinate and return its dimensions so the
-- caller can place the next irregular piece relative to the actual decoded art.
local function frame(root, definition, index, x, y)
    local node = render.create("modal", root)
    local width, height = node:set_dc6(definition.sheet, manifest.palettes[definition.palette], 0, index)
    node:set_position(x + width / 2, y + height / 2)
    return width, height
end

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("modal")
        self.controls = controls.new()
        if not render.assets_available() then return end

        local border = screen.border

        if border.placements then
            -- Newer/profile-specific manifests may provide explicit piece
            -- placement. In that case the loop can be completely data-driven.
            for _, placement in ipairs(border.placements) do
                frame(self.root, border, placement.frame, placement.x, placement.y)
            end
        else
            -- Legacy fallback explains the irregular expansion sheet manually:
            -- frames 0/1 build left edge, 2-5 walk across top, 6 is lower-right.
            local w0, h0 = frame(self.root, border, 0, 0, 0)
            frame(self.root, border, 1, 0, h0)

            local x = w0
            local w2, h2 = frame(self.root, border, 2, x, 0)

            -- `middle_overlap` removes the pixels intentionally shared by adjacent pieces.
            x = x + w2 - border.middle_overlap

            -- `select(1, ...)` means "take only the first returned value." frame()
            -- returns width,height but these two steps need width only.
            local w3 = select(1, frame(self.root, border, 3, x, 0))
            x = x + w3
            local w4 = select(1, frame(self.root, border, 4, x, 0))
            x = x + w4
            frame(self.root, border, 5, x, 0)
            frame(self.root, border, 6, x, h2)
        end

        local heading = screen.heading
        local content = screen.content
        text.create(self.root, "panel_heading", assert(locale.text("d2legacy.help.title")), heading.x, heading.y, heading.width)

        -- Bullet content is localized. The small yellow dot is separate DC6 art.
        local bullets = {"run", "items", "stand", "map", "menu", "chat", "skills", "help"}
        for index, key in ipairs(bullets) do
            local y = 59 + (index - 1) * 20
            local dot = render.create("modal", self.root)
            local dw, dh = dot:set_dc6(screen.yellow_bullet, manifest.palettes.sky, 0, 0)
            dot:set_position(88 + dw / 2, y + 4 + dh / 2)
            text.create(self.root, "formal_large", assert(locale.text("d2legacy.help." .. key)), 100, y, content.text_width, "left")
        end

        -- These labels point at familiar HUD pieces. Each row is simply
        -- {display text, recovered x, recovered y}.
        local callouts = {
            {"Life Orb",65,451},{"New Stats",222,355},{"Stamina Bar",315,450},
            {"Run/Walk Toggle",264,480},{"Experience Bar",370,476},{"Mini-Panel",450,371},
            {"Belt",535,490},{"New Skill",578,355},{"Mana Orb",745,451},
            {"Left Mouse Button Skill",135,382},{"Right Mouse Button Skill",675,381},
        }

        for _, label in ipairs(callouts) do
            -- Selected presentation profile can scale recovered X and shift HUD Y.
            local x = label[2] * content.scale_x
            local y = label[3] + content.hud_offset_y
            text.create(self.root, "formal_large", label[1], x, y, 180)
        end

        local close = screen.close
        button.create(self.root, self.controls, "close", close, assert(locale.text(close.label)), {
            layer="modal", show_label=false, sound=manifest.sounds.button,
            tooltip=assert(locale.text(close.label)), on_activate=function() scenes.toggle_overlay("help", "full") end,
        })
    end,

    update = function(self)
        self.controls:update()
        if input.pressed("help") or input.pressed("cancel") then scenes.pop() end
    end,
}
