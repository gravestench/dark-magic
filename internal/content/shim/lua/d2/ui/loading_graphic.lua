-- Shared adapter for Diablo II's progressive loading-screen artwork.
--
-- The original DC6 is shaped like an animation: many ordered frames. A loading
-- screen is different from a normal animation, though: we do NOT want time to
-- decide the frame. Real loading progress decides the frame.
--
-- Startup warmup and game loading both need that exact rule, so it lives here
-- once instead of being implemented twice and slowly drifting apart.

local M = {}

-- Attach a progressive loading animation to an already-created render node.
-- `candidates` is a list because different game archive layouts may expose an
-- equivalent verified asset at different paths.
function M.attach(node, candidates, palette)
    local errors = {}

    for _, path in ipairs(candidates) do
        -- `pcall` means "protected call." If decoding this candidate throws a
        -- Lua error, capture that error as a VALUE and try the next candidate.
        local ok, frames_or_error = pcall(function()
            -- 10 FPS is the authored animation rate. We pause immediately below;
            -- the rate is still useful because animation_seek uses seconds.
            return node:set_dc6_animation(path, palette, 0, 10, "once", "offsets")
        end)

        if ok then
            -- Stop automatic time-based playback: progress will drive the clock.
            node:animation_pause()
            -- Make the initial state explicit instead of waiting for one update.
            node:animation_seek(0)
            -- Return two values: number of frames and the path that worked.
            return frames_or_error, path
        end

        -- Preserve every failure. If all candidates fail, the eventual error is
        -- far more helpful than "last path failed" with no context.
        errors[#errors + 1] = path .. ": " .. tostring(frames_or_error)
    end

    error("unable to decode a verified Diablo II loading screen:\n" .. table.concat(errors, "\n"))
end

-- Choose the loading-screen frame from a normalized progress value.
function M.seek(node, frames, progress)
    -- This helper is intentionally safe to call before a graphic exists.
    if not node or not frames then return end

    -- Clamp progress into 0..1. `progress or 0` also makes nil mean "not begun."
    local fraction = math.max(0, math.min(1, progress or 0))

    -- There are `frames - 1` intervals from first frame to last. Divide by the
    -- authored 10 FPS because animation_seek takes time in seconds rather than a
    -- raw frame number.
    node:animation_seek(fraction * math.max(0, frames - 1) / 10)
end

return M
