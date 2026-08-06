local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")

return {
    enter = function(self)
        self.root = render.create("transition")
        self.root:set_z(1)
        self.root:set_position(400, 300)
        self.root:fill_rect(800, 600, 8, 8, 12, 255)
    end,

    update = function(self, elapsed)
        if input.pressed("confirm") then
            scenes.replace("main_menu")
        end
    end,
}
