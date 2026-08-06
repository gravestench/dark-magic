-- Character panel overlay example.
--
-- A blocking modal pauses updates to the world below it. The panel is anchored
-- to the left side, matching the eventual Diablo II character sheet placement.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")

return {
    blocks_update_below = true,

    create = function(self)
        self.root = render.create("modal")
        self.root:set_position(210, 300)
        self.root:fill_rect(380, 500, 28, 20, 12, 245)
    end,

    update = function(self)
        if input.pressed("character") or input.pressed("cancel") then
            scenes.pop()
        end
    end,
}
