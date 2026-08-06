local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local audio = require("dm.audio/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.game_loading

return {
    enter = function(self)
        audio.stop_group("frontend_music")
        self.elapsed = 0
        self.root = render.create("transition")
        self.root:set_position(400, 300)
        self.root:fill_rect(800, 600, 0, 0, 0, 255)
        if render.assets_available() then
            self.animation = render.create("transition", self.root)
            self.frames = self.animation:set_dc6_animation(
                screen.sheet, manifest.palettes[screen.palette], 0, 10, "once", "offsets")
            self.animation:set_position(0, 0)
            self.animation:animation_pause()
            self.animation:animation_seek(0)
        end
    end,

    update = function(self, elapsed)
        self.elapsed = self.elapsed + elapsed
        local progress = math.min(self.elapsed / screen.duration_seconds, 1)
        if self.animation then
            self.animation:animation_seek(progress * (self.frames - 1) / 10)
        end
        if progress >= 1 then scenes.replace("game_world") end
    end,
}
