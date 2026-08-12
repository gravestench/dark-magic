-- Startup warmup + cinematic sequence.
--
-- This is a very useful scene for understanding ASYNCHRONOUS ENGINE WORK from
-- Lua's point of view. Lua does not create worker threads or upload textures.
-- It receives a preload job handle/status and a diagnostics snapshot, then waits
-- until both CPU preparation and queued GPU warmup are finished.
--
-- Only after that gate does the scene start the cinematic sequence. This avoids
-- making background asset decoding compete with video decoding for the same CPU
-- and graphics-owner resources.
--
-- Video follows the same capability philosophy: Lua asks `engine.video/v1` to play
-- a path and receives a checked playback handle. Decoder processes, temporary
-- files, native resources, and cleanup remain engine-owned details.

local render = require("engine.render/v1")
local input = require("engine.input/v1")
local scenes = require("engine.scene/v1")
local video = require("engine.video/v1")
local data = require("engine.data/v1")
local preload = require("d2legacy.ui.preload")
local compat = require("d2legacy.ui.compat")
local loading_graphic = require("d2legacy.ui.loading_graphic")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local startup = manifest.startup
local screen = assert(manifest.screens.loading)
local game_loading = assert(manifest.screens.game_loading)

return {
    enter = function(self)
        -- Ask the renderer capability to start preparing the frontend bundle.
        self.preload_job = preload.frontend()

        -- A nil job is a valid headless/no-assets case, so warming is a boolean
        -- derived from whether a real job exists.
        self.warming = self.preload_job ~= nil

        -- Transition root is a simple black/letterbox-style stage beneath both
        -- the loading graphic and embedded movie presenter.
        self.root = render.create("transition")
        self.root:set_z(-1)
        self.root:set_position(screen.x, screen.y)
        self.root:fill_rect(screen.width, screen.height,
            screen.fill.red, screen.fill.green, screen.fill.blue, screen.fill.alpha)

        if self.warming and render.assets_available() then
            -- Reuse the same progressive loading-screen art that game loading
            -- uses. Here its "progress" represents frontend warmup readiness.
            self.warm_animation = render.create("transition", self.root)
            self.warm_frames = loading_graphic.attach(
                self.warm_animation,
                compat.frontend.game_loading.sheets or { game_loading.sheet },
                manifest.palettes[game_loading.palette]
            )
            self.warm_animation:set_position(0, 0)

            -- We do not know the maximum pending GPU bytes in advance. Remember
            -- the largest value observed so far and use it as a progress baseline.
            self.warm_pending_peak = 0
        end

        -- `index` points into startup.sequence. Zero means no cinematic tried yet.
        self.index = 0

        -- Headless/no-assets mode has nothing to warm, so begin sequence immediately.
        if not self.warming then self:advance() end
    end,

    -- Advance to the next playable startup movie, or to title when sequence ends.
    advance = function(self)
        if self.playback then
            -- Explicitly stop the previous checked playback before replacing its handle.
            self.playback:stop()
            self.playback = nil
        end

        -- `while` lets unavailable/failing entries be skipped until one actually starts.
        while self.index < #startup.sequence do
            self.index = self.index + 1

            if video.available() then
                -- Playback can fail because a codec/file is unavailable. Protected
                -- call turns that error into values so manifest failure policy decides.
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

        -- Guard prevents several update branches from replacing title repeatedly.
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
            -- CPU preload status and renderer diagnostics are VALUE snapshots.
            local status = preload.frontend_status()
            local diagnostics = render.diagnostics()

            -- CPU is ready when there is no job/status or every request completed.
            local cpu_ready = status == nil or status.done
            -- GPU is ready when owner-thread warm texture queue is empty.
            local gpu_ready = diagnostics.pending_warm_textures == 0

            local cpu_progress = status and status.total > 0 and status.completed / status.total or 1

            self.warm_pending_peak = math.max(self.warm_pending_peak or 0, diagnostics.pending_warm_bytes)

            local gpu_progress = self.warm_pending_peak > 0
                and 1 - diagnostics.pending_warm_bytes / self.warm_pending_peak or 1

            -- The visual bar reserves first 75% for CPU preparation and last 25%
            -- for GPU warm uploads. This is PRESENTATION math, not worker scheduling.
            local progress = cpu_ready and (0.75 + 0.25 * gpu_progress) or (0.75 * cpu_progress)
            loading_graphic.seek(self.warm_animation, self.warm_frames, progress)

            -- Stay in warmup until BOTH kinds of work are done.
            if not cpu_ready or not gpu_ready then return end

            self.warming = false

            if self.warm_animation then
                -- This one node is no longer needed once cinematics begin, so
                -- destroy its checked retained handle explicitly now rather than
                -- waiting for the whole scene to leave.
                self.warm_animation:destroy()
                self.warm_animation = nil
            end

            self:advance()
            return
        end

        -- Skip action means "advance to next startup entry," not "jump directly
        -- to title"; repeated skip input can therefore step through the sequence.
        if startup.skippable and input.pressed("skip") then
            self:advance()
            return
        end

        -- If no playback handle exists, try to advance/start one.
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
