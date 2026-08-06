local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local dc6 = require("darkmagic.ui.dc6")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.main_menu
local logo = screen.logo
local units_palette = manifest.palettes[logo.palette]

return {
    enter = function(self)
        self.root = render.create("hud")
        self.background = dc6.frontend_background(
            self.root,
            "hud",
            screen.background,
            manifest.palettes[screen.palette],
            manifest.layouts.frontend_tiles
        )
        if render.assets_available() then
            self.logo = {
                black_left = render.create("hud", self.root),
                black_right = render.create("hud", self.root),
                fire_left = render.create("hud", self.root),
                fire_right = render.create("hud", self.root),
            }
            self.logo.fire_left:set_blend(logo.fire_blend)
            self.logo.fire_right:set_blend(logo.fire_blend)
            self:configure_logo()
        end
    end,

    configure_logo = function(self)
        local function animate(node, path)
            dc6.anchored_frame(node, path, units_palette, logo.anchor.x, logo.anchor.y, 0)
            node:set_dc6_animation(path, units_palette, 0, logo.frames_per_second, "loop")
        end
        animate(self.logo.black_left, logo.black_left)
        animate(self.logo.black_right, logo.black_right)
        animate(self.logo.fire_left, logo.fire_left)
        animate(self.logo.fire_right, logo.fire_right)
    end,

    update = function(self, elapsed)
        if input.pressed("confirm") then
            scenes.replace("character_select")
        end
    end,
}
