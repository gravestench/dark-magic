local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local dc6 = require("darkmagic.ui.dc6")

local frontend = "data/global/ui/FrontEnd/"
local units_palette = "data/global/Palette/units/pal.dat"

return {
    enter = function(self)
        self.root = render.create("hud")
        self.background = dc6.frontend_background(
            self.root,
            "hud",
            frontend .. "gameselectscreenEXP.dc6",
            "data/global/Palette/sky/pal.dat"
        )
        self.logo_frame = 0
        self.logo_elapsed = 0
        if render.assets_available() then
            self.logo = {
                black_left = render.create("hud", self.root),
                black_right = render.create("hud", self.root),
                fire_left = render.create("hud", self.root),
                fire_right = render.create("hud", self.root),
            }
            self.logo.fire_left:set_blend("additive")
            self.logo.fire_right:set_blend("additive")
            self:update_logo()
        end
    end,

    update_logo = function(self)
        local frame = self.logo_frame
        dc6.anchored_frame(self.logo.black_left, frontend .. "D2logoBlackLeft.DC6", units_palette, 400, 120, frame)
        dc6.anchored_frame(self.logo.black_right, frontend .. "D2logoBlackRight.DC6", units_palette, 400, 120, frame)
        dc6.anchored_frame(self.logo.fire_left, frontend .. "D2logoFireLeft.DC6", units_palette, 400, 120, frame)
        dc6.anchored_frame(self.logo.fire_right, frontend .. "D2logoFireRight.DC6", units_palette, 400, 120, frame)
    end,

    update = function(self, elapsed)
        self.logo_elapsed = self.logo_elapsed + elapsed
        if self.logo and self.logo_elapsed >= 1 / 15 then
            self.logo_elapsed = self.logo_elapsed - 1 / 15
            self.logo_frame = (self.logo_frame + 1) % 30
            self:update_logo()
        end
        if input.pressed("confirm") then
            scenes.replace("character_select")
        end
    end,
}
