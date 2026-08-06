local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local dc6 = require("darkmagic.ui.dc6")

return {
    create = function(self)
        self.root = render.create("transition")
        self.background = dc6.frontend_background(
            self.root,
            "transition",
            "data/global/ui/FrontEnd/TitleScreen.dc6",
            "data/global/Palette/sky/pal.dat"
        )
    end,
    update = function(self, elapsed)
        if input.pressed("confirm") then scenes.replace("main_menu") end
    end,
}
