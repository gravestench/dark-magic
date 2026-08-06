local render = require("dm.render/v1")
local input = require("dm.input/v1")
local saves = require("dm.save/v1")
local scenes = require("dm.scene/v1")

return {
    create = function(self)
        self.characters = saves.characters()
        if #self.characters == 0 then
            scenes.replace("character_create")
            return
        end
        self.root = render.create("hud")
        self.root:set_position(400, 300)
        self.root:fill_rect(800, 600, 20, 14, 9, 255)
    end,
    update = function(self, elapsed)
        if input.pressed("cancel") then
            scenes.replace("main_menu")
        elseif input.pressed("confirm") and #self.characters > 0 then
            assert(saves.select(self.characters[1].id))
            scenes.replace("game_loading")
        end
    end,
}
