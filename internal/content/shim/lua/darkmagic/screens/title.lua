local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local dc6 = require("darkmagic.ui.dc6")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.title

return {
    create = function(self)
        self.root = render.create("transition")
        self.background = dc6.frontend_background(
            self.root,
            "transition",
            screen.background,
            manifest.palettes[screen.palette],
            manifest.layouts.frontend_tiles
        )
    end,
    update = function(self, elapsed)
        if input.pressed("confirm") then scenes.replace("main_menu") end
    end,
}
