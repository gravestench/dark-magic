-- Inventory overlay example.
--
-- This placeholder demonstrates the recommended blocking-modal lifecycle. The
-- authentic panel will be assembled from manifest and Inventory.txt data.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")

return {
    blocks_update_below = true,

    enter = function(self)
        self.root = render.create("modal")
        self.root:set_position(400, 300)
        self.root:fill_rect(520, 500, 18, 18, 18, 230)
    end,

    update = function(self, elapsed)
        if input.pressed("inventory") or input.pressed("cancel") then
            scenes.pop()
        end
    end,
}
