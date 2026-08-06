local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")

return {
    enter = function(self)
        self.root = render.create("world")
    end,

    update = function(self, elapsed)
        if input.pressed("inventory") then
            scenes.push("inventory")
        end
    end,
}
