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

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local startup = manifest.startup
local screen = assert(manifest.screens.loading)

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
            if not cpu_ready or not gpu_ready then return end

            self.warming = false
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
