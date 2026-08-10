-- Shared adapter for Diablo II's progressive loading-screen artwork.
--
-- The DC6 is authored like an animation, but loading screens choose its frame
-- from actual progress. Keeping that little rule here prevents startup and
-- game-world loading from quietly drifting into different behavior.
local M = {}

function M.attach(node, candidates, palette)
    local errors = {}
    for _, path in ipairs(candidates) do
        local ok, frames_or_error = pcall(function()
            return node:set_dc6_animation(path, palette, 0, 10, "once", "offsets")
        end)
        if ok then
            node:animation_pause()
            node:animation_seek(0)
            return frames_or_error, path
        end
        errors[#errors + 1] = path .. ": " .. tostring(frames_or_error)
    end
    error("unable to decode a verified Diablo II loading screen:\n" .. table.concat(errors, "\n"))
end

function M.seek(node, frames, progress)
    if not node or not frames then return end
    local fraction = math.max(0, math.min(1, progress or 0))
    node:animation_seek(fraction * math.max(0, frames - 1) / 10)
end

return M
