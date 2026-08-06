local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")

return {
    enter = function(self)
        self.root = render.create("hud")
        self.root:set_position(400, 300)
        self.root:fill_rect(800, 600, 24, 12, 10, 255)
    end,

    update = function(self, elapsed)
        if input.pressed("confirm") then
            scenes.replace("character_select")
        end
    end,
}
