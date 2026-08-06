local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local dc6 = require("darkmagic.ui.dc6")
local cursor = require("darkmagic.ui.cursor")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))

return function(name)
    local screen = assert(manifest.screens[name])
    return {
        create = function(self)
            self.root = render.create("hud")
            self.background = dc6.frontend_background(self.root, "hud", screen.background,
                manifest.palettes[screen.palette], manifest.layouts.frontend_tiles)
            self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
        end,
        update = function(self)
            self.cursor:update()
            if input.pressed("cancel") then scenes.replace("main_menu") end
        end,
    }
end
