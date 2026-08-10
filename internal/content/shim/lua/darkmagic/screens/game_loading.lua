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

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local screen = manifest.screens.game_loading

local function attach_loading_animation(node)
    local candidates = compat.frontend.game_loading.sheets or { screen.sheet }
    local errors = {}
    for _, path in ipairs(candidates) do
        local ok, frames_or_error = pcall(function()
            return node:set_dc6_animation(
                path,
                manifest.palettes[screen.palette],
                0,
                10,
                "once",
                "offsets"
            )
        end)
        if ok then
            return frames_or_error, path
        end
        errors[#errors + 1] = path .. ": " .. tostring(frames_or_error)
    end
    error("unable to decode a verified Diablo II loading screen:\n" .. table.concat(errors, "\n"))
end

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
            self.frames, self.loading_sheet = attach_loading_animation(self.animation)
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
            -- set_dc6_animation is authored at 10 FPS above, so seek directly
            -- across the progressive loading frames as dependency progress rises.
            self.animation:animation_seek(self.displayed_progress * (self.frames - 1) / 10)
        end
        if status.state == "complete" and self.displayed_progress >= 1 then
            scenes.replace("game_world")
        end
    end,
}
