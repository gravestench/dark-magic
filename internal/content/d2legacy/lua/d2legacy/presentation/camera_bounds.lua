-- Camera bounds for finite map canvases.
--
-- A side panel changes where the player appears on screen. That moves the map
-- underneath the player, but it must never pull the finite map edge into view.
-- This helper keeps the map covering the complete physical viewport. It knows
-- nothing about gameplay positions, panels, or DT1 graphics.

local camera_bounds = {}

function camera_bounds.clamp_center(center, content_size, viewport_size)
    if content_size <= viewport_size then
        -- A canvas smaller than the window cannot cover it, so the least
        -- surprising presentation is to center the canvas.
        return viewport_size / 2
    end

    local smallest_center = viewport_size - content_size / 2
    local largest_center = content_size / 2
    return math.max(smallest_center, math.min(largest_center, center))
end

function camera_bounds.anchor_for_center(center, content_size, camera_position)
    return center - content_size / 2 + camera_position
end

return camera_bounds
