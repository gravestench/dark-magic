local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")

return {
    enter = function(self)
        self.root = render.create("hud")
    end,

    update = function(self, elapsed)
        if input.pressed("confirm") then
            scenes.replace("game_world")
        end
    end,
}
