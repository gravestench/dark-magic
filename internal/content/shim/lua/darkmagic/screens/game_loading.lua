-- Transition from the frontend shell into an interactive game session.
--
-- Engine-owned dependency work determines the target progress. Elapsed time is
-- used only to smooth the visual sweep so fast local loads remain legible.
local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local audio = require("dm.audio/v1")
local loading = require("dm.loading/v1")
local compat = require("darkmagic.ui.compat")
local loading_graphic = require("darkmagic.ui.loading_graphic")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.game_loading

return {
    enter = function(self)
        -- Frontend music spans all menus and stops only when game loading begins.
        audio.stop_group("frontend_music")
        self.displayed_progress = 0
        loading.begin(screen.dependencies)
        self.root = render.create("transition")
        self.root:set_position(screen.x, screen.y)
        self.root:fill_rect(screen.width, screen.height,
            screen.fill.red, screen.fill.green, screen.fill.blue, screen.fill.alpha)
        if render.assets_available() then
            self.animation = render.create("transition", self.root)
            self.frames, self.loading_sheet = loading_graphic.attach(
                self.animation,
                compat.frontend.game_loading.sheets or { screen.sheet },
                manifest.palettes[screen.palette]
            )
            self.animation:set_position(0, 0)
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
            -- set_dc6_animation is authored at 10 FPS above, so seek directly
            -- across the progressive loading frames as dependency progress rises.
            loading_graphic.seek(self.animation, self.frames, self.displayed_progress)
        end
        if status.state == "complete" and self.displayed_progress >= 1 then
            scenes.replace("game_world")
        end
    end,
}
