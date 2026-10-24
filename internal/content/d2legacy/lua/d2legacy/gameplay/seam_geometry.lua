-- Prepared seam geometry shared by presentation targeting and authoritative crossing.
-- This module reads the published seam facts but never mutates an entity.

local initial_available, initial = pcall(require, "engine.initial_data/v1")

local M = {}

-- Reject non-finite coordinates before they can enter crossing or pointer math.
local function finite(value)
    return type(value) == "number" and value == value and value ~= math.huge and value ~= -math.huge
end

-- Validate the directed seam contract once before either consumer uses it.
local function validate(seam)
    assert(type(seam) == "table", "world transition seam must be a table")
    for _, name in ipairs({
        "source_level",
        "destination_level",
        "source_x",
        "source_y",
        "arrival_x",
        "arrival_y",
        "world_width",
        "world_height",
        "trigger_radius",
    }) do
        assert(finite(seam[name]), "world transition " .. name .. " must be finite")
    end
    assert(seam.source_level ~= seam.destination_level, "world transition must change levels")
    assert(
        seam.source_direction == "north"
            or seam.source_direction == "east"
            or seam.source_direction == "south"
            or seam.source_direction == "west",
        "world transition source direction must be cardinal"
    )
    assert(seam.trigger_radius > 0, "world transition trigger radius must be positive")
    if seam.trigger_lane_min ~= nil or seam.trigger_lane_max ~= nil then
        assert(
            finite(seam.trigger_lane_min)
                and finite(seam.trigger_lane_max)
                and seam.trigger_lane_min == math.floor(seam.trigger_lane_min)
                and seam.trigger_lane_max == math.floor(seam.trigger_lane_max)
                and seam.trigger_lane_min <= seam.trigger_lane_max,
            "world transition trigger lanes must be ordered integers"
        )
    end
    assert(seam.world_width > 0 and seam.world_height > 0, "world transition bounds must be positive")
    assert(seam.arrival_x >= 0 and seam.arrival_x < seam.world_width, "world transition arrival x is outside bounds")
    assert(seam.arrival_y >= 0 and seam.arrival_y < seam.world_height, "world transition arrival y is outside bounds")
end

-- Read the prepared seams shared by the screen and authoritative system.
function M.configured_seams()
    local configuration = initial_available and initial.get("d2legacy.world_transitions") or nil
    local seams = configuration and configuration.seams or {}
    for _, seam in ipairs(seams) do
        validate(seam)
    end
    return seams
end

-- Express both horizontal and vertical seams as normal and tangential axes.
local function normal_and_tangent(direction, x, y)
    if direction == "east" or direction == "west" then
        return x, y
    end
    return y, x
end

-- Convert a point on those axes back to map coordinates.
local function point_from_axes(direction, normal, tangent)
    if direction == "east" or direction == "west" then
        return normal, tangent
    end
    return tangent, normal
end

-- The town source has a bounded strip; generated and reverse sources use circles.
function M.matching_seam(seams, level, x, y)
    for _, seam in ipairs(seams) do
        if seam.source_level == level then
            if seam.trigger_lane_min ~= nil then
                local normal, tangent = normal_and_tangent(seam.source_direction, x, y)
                local source_normal = normal_and_tangent(seam.source_direction, seam.source_x, seam.source_y)
                local lane = math.floor(tangent + 0.5)
                if
                    math.abs(normal - source_normal) <= seam.trigger_radius
                    and lane >= seam.trigger_lane_min
                    and lane <= seam.trigger_lane_max
                then
                    return seam
                end
            else
                local dx, dy = x - seam.source_x, y - seam.source_y
                if dx * dx + dy * dy <= seam.trigger_radius * seam.trigger_radius then
                    return seam
                end
            end
        end
    end
    return nil
end

-- Expose the authority's crossing test to callers that need a Boolean.
function M.at_trigger(level, x, y)
    return M.matching_seam(M.configured_seams(), level, x, y) ~= nil
end

-- Choose the first map edge reached by a ray from an in-map hero to an outside pointer.
local function first_boundary(hero_x, hero_y, dx, dy, width, height)
    local first_t, direction = math.huge, nil
    local function consider(t, side)
        if t > 0 and t <= 1 and t < first_t then
            first_t, direction = t, side
        end
    end
    if dx > 0 then
        consider((width - hero_x) / dx, "east")
    elseif dx < 0 then
        consider(-hero_x / dx, "west")
    end
    if dy > 0 then
        consider((height - hero_y) / dy, "south")
    elseif dy < 0 then
        consider(-hero_y / dy, "north")
    end
    return direction, hero_x + first_t * dx, hero_y + first_t * dy
end

-- Only a ray that exits through the prepared town opening gets a reachable in-map movement target.
function M.pointer_target(level, hero_x, hero_y, target_x, target_y, width, height)
    if
        level ~= 1
        or hero_x < 0
        or hero_y < 0
        or hero_x >= width
        or hero_y >= height
        or not (target_x < 0 or target_y < 0 or target_x >= width or target_y >= height)
    then
        return target_x, target_y
    end

    local dx, dy = target_x - hero_x, target_y - hero_y
    local length_squared = dx * dx + dy * dy
    local exit_direction, exit_x, exit_y = first_boundary(hero_x, hero_y, dx, dy, width, height)
    local _, exit_tangent = normal_and_tangent(exit_direction, exit_x, exit_y)

    for _, seam in ipairs(M.configured_seams()) do
        if seam.source_level == level and seam.source_direction == exit_direction then
            local _, seam_tangent = normal_and_tangent(exit_direction, seam.source_x, seam.source_y)
            if seam.trigger_lane_min ~= nil then
                -- A click just beyond the bridge shoulder still aims through the prepared opening.
                local margin = 2
                local minimum = seam.trigger_lane_min - margin
                local maximum = seam.trigger_lane_max + margin
                local exit_lane = math.floor(exit_tangent + 0.5)
                if exit_lane >= minimum and exit_lane <= maximum then
                    local hero_normal, hero_tangent = normal_and_tangent(exit_direction, hero_x, hero_y)
                    local target_normal, target_tangent = normal_and_tangent(exit_direction, target_x, target_y)
                    local source_normal = normal_and_tangent(exit_direction, seam.source_x, seam.source_y)
                    local plane = (source_normal - hero_normal) / (target_normal - hero_normal)
                    if plane > 0 and plane < 1 then
                        local lane = math.floor(hero_tangent + plane * (target_tangent - hero_tangent) + 0.5)
                        if lane >= minimum and lane <= maximum then
                            lane = math.max(seam.trigger_lane_min, math.min(seam.trigger_lane_max, lane))
                            return point_from_axes(exit_direction, source_normal, lane)
                        end
                    end
                end
            elseif math.abs(exit_tangent - seam_tangent) <= seam.trigger_radius then
                local along = ((seam.source_x - hero_x) * dx + (seam.source_y - hero_y) * dy) / length_squared
                if along > 0 and along < 1 then
                    local nearest_x, nearest_y = hero_x + along * dx, hero_y + along * dy
                    local off_x, off_y = seam.source_x - nearest_x, seam.source_y - nearest_y
                    if off_x * off_x + off_y * off_y <= seam.trigger_radius * seam.trigger_radius then
                        return seam.source_x, seam.source_y
                    end
                end
            end
        end
    end
    return target_x, target_y
end

return M
