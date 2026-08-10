-- Transition from the frontend into an interactive game session.
--
-- `dm.loading/v1` owns dependency preparation. Lua does not load maps/codecs/data
-- directly here; it asks the engine to begin a named dependency set and reads a
-- small progress/status snapshot each frame.
--
-- There are TWO progress values on purpose:
--   status.progress          = real engine work progress (authority/fact)
--   self.displayed_progress  = smoothed presentation progress
--
-- The display may lag behind reality so a very fast local load does not flash a
-- loading screen for one unreadable frame. It may never run AHEAD of real work.

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
        -- Frontend music intentionally spans menu scenes but stops at the game boundary.
        audio.stop_group("frontend_music")

        self.displayed_progress = 0

        -- Begin engine-owned preparation described by manifest dependency IDs.
        loading.begin(screen.dependencies)

        self.root = render.create("transition")
        self.root:set_position(screen.x, screen.y)
        self.root:fill_rect(screen.width, screen.height,
            screen.fill.red, screen.fill.green, screen.fill.blue, screen.fill.alpha)

        if render.assets_available() then
            self.animation = render.create("transition", self.root)

            -- loading_graphic turns the progressive DC6 into a paused animation
            -- that can be driven by a 0..1 progress value instead of wall time.
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

        -- Convert elapsed seconds into maximum visual progress this frame.
        local step = elapsed / screen.sweep_seconds

        -- `math.min(real, displayed+step)` means the display moves smoothly toward
        -- real progress but can NEVER claim more work is done than the engine says.
        self.displayed_progress = math.min(status.progress, self.displayed_progress + step)

        if self.animation then
            loading_graphic.seek(self.animation, self.frames, self.displayed_progress)
        end

        -- Do not enter gameplay merely because real work finished; let the
        -- smoothed display finish its sweep too so the transition remains legible.
        if status.state == "complete" and self.displayed_progress >= 1 then
            scenes.replace("game_world")
        end
    end,
}
