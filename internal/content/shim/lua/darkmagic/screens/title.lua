local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local dc6 = require("darkmagic.ui.dc6")
local cursor = require("darkmagic.ui.cursor")
local audio = require("dm.audio/v1")

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
        if render.assets_available() then
            self.logo = {
                black_left = render.create("transition", self.root),
                black_right = render.create("transition", self.root),
                fire_left = render.create("transition", self.root),
                fire_right = render.create("transition", self.root),
            }
            local logo = screen.logo
            local palette = manifest.palettes[logo.palette]
            self.logo.fire_left:set_blend(logo.fire_blend)
            self.logo.fire_right:set_blend(logo.fire_blend)
            local function animate(node, path)
                dc6.anchored_animation(node, path, palette, logo.anchor.x, logo.anchor.y,
                    logo.frames_per_second, "loop")
            end
            animate(self.logo.black_left, logo.black_left)
            animate(self.logo.black_right, logo.black_right)
            animate(self.logo.fire_left, logo.fire_left)
            animate(self.logo.fire_right, logo.fire_right)
        end
        self.cursor = cursor.new(self.root, manifest.cursor, manifest.palettes)
    end,
    update = function(self, elapsed)
        self.cursor:update()
        if input.pressed("skip") then
            if audio.exists(manifest.sounds.select) then audio.play(manifest.sounds.select) end
            scenes.replace("main_menu")
        end
    end,
}
