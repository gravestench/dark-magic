-- Startup cinematic sequence.
--
-- The scene consumes only the versioned video capability and manifest policy.
-- Decoder selection, temporary files, native resources, and cleanup remain
-- engine-owned implementation details.
local render = require("dm.render/v1")
local input = require("dm.input/v1")
local scenes = require("dm.scene/v1")
local video = require("dm.video/v1")
local data = require("dm.data/v1")
local preload = require("darkmagic.ui.preload")
local compat = require("darkmagic.ui.compat")
local loading_graphic = require("darkmagic.ui.loading_graphic")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local startup = manifest.startup
local screen = assert(manifest.screens.loading)
local game_loading = assert(manifest.screens.game_loading)

return {
    enter = function(self)
        -- Do not compete with video decoding for CPU time or the graphics
        -- owner thread. Startup has an explicit warmup phase; cinematics begin
        -- only after CPU preparation and queued texture uploads are complete.
        self.preload_job = preload.frontend()
        self.warming = self.preload_job ~= nil
        self.root = render.create("transition")
        -- Keep the letterbox backdrop below the embedded presenter, which
        -- occupies z=0 on the transition layer.
        self.root:set_z(-1)
        self.root:set_position(screen.x, screen.y)
        self.root:fill_rect(screen.width, screen.height,
            screen.fill.red, screen.fill.green, screen.fill.blue, screen.fill.alpha)
        if self.warming and render.assets_available() then
            self.warm_animation = render.create("transition", self.root)
            self.warm_frames = loading_graphic.attach(
                self.warm_animation,
                compat.frontend.game_loading.sheets or { game_loading.sheet },
                manifest.palettes[game_loading.palette]
            )
            self.warm_animation:set_position(0, 0)
            self.warm_pending_peak = 0
        end
        self.index = 0
        if not self.warming then self:advance() end
    end,

    advance = function(self)
        -- Stop the previous checked handle before attempting the next entry.
        if self.playback then
            self.playback:stop()
            self.playback = nil
        end
        while self.index < #startup.sequence do
            self.index = self.index + 1
            if video.available() then
                local ok, playback = pcall(video.play, startup.sequence[self.index])
                if ok then
                    self.playback = playback
                    return
                end
                if startup.failure ~= "skip" then
                    error(playback)
                end
            end
        end
        if not self.finished then
            self.finished = true
            scenes.replace("title")
        end
    end,

    update = function(self)
        if self.finished then
            return
        end
        if self.warming then
            local status = preload.frontend_status()
            local diagnostics = render.diagnostics()
            local cpu_ready = status == nil or status.done
            local gpu_ready = diagnostics.pending_warm_textures == 0
            local cpu_progress = status and status.total > 0 and status.completed / status.total or 1
            self.warm_pending_peak = math.max(self.warm_pending_peak or 0, diagnostics.pending_warm_bytes)
            local gpu_progress = self.warm_pending_peak > 0
                and 1 - diagnostics.pending_warm_bytes / self.warm_pending_peak or 1
            local progress = cpu_ready and (0.75 + 0.25 * gpu_progress) or (0.75 * cpu_progress)
            loading_graphic.seek(self.warm_animation, self.warm_frames, progress)
            if not cpu_ready or not gpu_ready then return end

            self.warming = false
            if self.warm_animation then
                self.warm_animation:destroy()
                self.warm_animation = nil
            end
            self:advance()
            return
        end
        if startup.skippable and input.pressed("skip") then
            self:advance()
            return
        end
        if not self.playback then
            self:advance()
            return
        end

        local status = self.playback:status()
        if status.state == "complete" or status.state == "stopped" then
            self:advance()
        end
        if status.state == "failed" then
            if startup.failure == "skip" then
                self:advance()
            else
                error(status.error or "video playback failed")
            end
        end
    end,
}
