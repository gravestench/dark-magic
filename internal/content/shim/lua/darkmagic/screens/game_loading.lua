-- Transition from the frontend shell into an interactive game session.
--
-- Engine-owned dependency work determines the target progress. Elapsed time is
-- used only to smooth the visual sweep so fast local loads remain legible.
local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local audio = require("dm.audio/v1")
local loading = require("dm.loading/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.game_loading

return {
    enter = function(self)
        -- Frontend music spans all menus and stops only when game loading begins.
        audio.stop_group("frontend_music")
        self.displayed_progress = 0
        loading.begin(screen.dependencies)
        self.root = render.create("transition")
        self.root:set_position(400, 300)
        self.root:fill_rect(800, 600, 0, 0, 0, 255)
        if render.assets_available() then
            self.animation = render.create("transition", self.root)
            self.frames = self.animation:set_dc6_animation(
                screen.sheet,
                manifest.palettes[screen.palette],
                0,
                10,
                "once",
                "offsets"
            )
            self.animation:set_position(0, 0)
            self.animation:animation_pause()
            self.animation:animation_seek(0)
        end
    end,

    update = function(self, elapsed)
        local status = loading.status()
        if status.state == "failed" then
            error(status.error or "game dependencies failed to load")
        end
        local step = elapsed / screen.sweep_seconds
        self.displayed_progress = math.min(status.progress, self.displayed_progress + step)
        if self.animation then
            -- Seeking a paused renderer-owned animation makes progress
            -- deterministic and independent of the frame rate.
            self.animation:animation_seek(self.displayed_progress * (self.frames - 1) / 10)
        end
        if status.state == "complete" and self.displayed_progress >= 1 then
            scenes.replace("game_world")
        end
    end,
}
