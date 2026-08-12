-- Move players through authored level seams during the authoritative tick.
--
-- Go supplies only collision-derived endpoints. This module owns every Diablo
-- rule: which level pairs are connected, how close a player must be, where the
-- player arrives, and that movement stops after crossing. Keeping the rule in
-- one small system also makes replay deterministic: no wall-clock or native
-- input state is consulted while the fixed-tick schedule is running.

local ecs = require("engine.ecs/v1")
local initial_available, initial = pcall(require, "engine.initial_data/v1")

local M = {}
local trigger_radius = 2

local function finite(value)
    return type(value) == "number" and value == value
        and value ~= math.huge and value ~= -math.huge
end

local function validate(seam)
    assert(type(seam) == "table", "world transition seam must be a table")
    for _, name in ipairs({
        "source_level", "destination_level", "source_x", "source_y",
        "arrival_x", "arrival_y", "world_width", "world_height",
    }) do
        assert(finite(seam[name]), "world transition " .. name .. " must be finite")
    end
    assert(seam.source_level ~= seam.destination_level, "world transition must change levels")
    assert(seam.world_width > 0 and seam.world_height > 0, "world transition bounds must be positive")
    assert(seam.arrival_x >= 0 and seam.arrival_x < seam.world_width, "world transition arrival x is outside bounds")
    assert(seam.arrival_y >= 0 and seam.arrival_y < seam.world_height, "world transition arrival y is outside bounds")
end

local function configured_seams()
    local configuration = initial_available and initial.get("d2legacy.world_transitions") or {}
    local seams = configuration.seams or {}
    for _, seam in ipairs(seams) do validate(seam) end
    return seams
end

local function matching_seam(seams, level, x, y)
    for _, seam in ipairs(seams) do
        if seam.source_level == level then
            local dx, dy = x - seam.source_x, y - seam.source_y
            if math.sqrt(dx * dx + dy * dy) <= trigger_radius then return seam end
        end
    end
    return nil
end

local function cross(entity, seam)
    local location = ecs.get(entity, "d2legacy.world.location")
    local position = ecs.get(entity, "d2legacy.world.position")
    local bounds = ecs.get(entity, "d2legacy.world.bounds")
    local velocity = ecs.get(entity, "d2legacy.world.velocity")
    location:set("level_id", seam.destination_level)
    position:set("x", seam.arrival_x)
    position:set("y", seam.arrival_y)
    bounds:set("width", seam.world_width)
    bounds:set("height", seam.world_height)
    velocity:set("x", 0)
    velocity:set("y", 0)
end

function M.register()
    local seams = configured_seams()
    ecs.system({
        id = "d2legacy.world.transition",
        -- Collision follows movement in the engine's stable phase order. A
        -- seam is geometry too, so crossing it belongs at this barrier.
        phase = "collision",
        query = { all = {
            "d2legacy.world.player_control", "d2legacy.world.location",
            "d2legacy.world.position", "d2legacy.world.bounds", "d2legacy.world.velocity",
        } },
        read = { "d2legacy.world.player_control" },
        write = {
            "d2legacy.world.location", "d2legacy.world.position",
            "d2legacy.world.bounds", "d2legacy.world.velocity",
        },
        update = function(_, entities)
            for _, entity in ipairs(entities) do
                local location = ecs.get(entity, "d2legacy.world.location")
                local position = ecs.get(entity, "d2legacy.world.position")
                local seam = matching_seam(seams, location:get("level_id"), position:get("x"), position:get("y"))
                if seam then cross(entity, seam) end
            end
        end,
    })
end

return M
