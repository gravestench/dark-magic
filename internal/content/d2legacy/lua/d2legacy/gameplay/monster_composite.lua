-- Turn a copied monster appearance snapshot into a legacy COF/DCC recipe.
--
-- Simulation never sees paths, textures, or retained render nodes. It says
-- "this is monster FA, walking east, using these MonStats2 body pieces". This
-- adapter is the one small place that knows how those facts are spelled inside
-- the original MPQs.

local render = require("engine.render/v1")
local player_composite = require("d2legacy.gameplay.player_composite")

local M = {}

local function upper(value) return string.upper(assert(value)) end

local function decode_components(encoded)
    local result = {}
    for pair in string.gmatch(encoded or "", "[^,]+") do
        local component, visual = string.match(pair, "^([^=]+)=([^=]+)$")
        if component and visual then result[upper(component)] = upper(visual) end
    end
    return result
end

local function component_path(token, component, visual, mode, weapon_class)
    return string.format("data/global/monsters/%s/%s/%s%s%s%s%s.dcc",
        token, component, token, component, visual, mode, weapon_class)
end

-- Velocity is an authoritative world-space fact; facing memory is deliberately
-- presentation-owned. A stopped monster keeps looking the way it last moved,
-- but that visual memory cannot alter combat, collision, or replay checksums.
function M.facing(previous, velocity_x, velocity_y)
    if velocity_x == 0 and velocity_y == 0 then return previous or 0 end
    local angle = math.atan(velocity_y, velocity_x)
    local bucket = math.floor((angle + math.pi / 8) / (math.pi / 4)) % 8
    -- Clockwise screen/world buckets -> the readable eight-way direction space
    -- shared by player_composite. The resolver performs the legacy interleave.
    return ({3, 4, 0, 5, 1, 6, 2, 7})[bucket + 1]
end

function M.resolve(snapshot)
    local token, mode = upper(snapshot.token), upper(snapshot.mode)
    local weapon_class = upper(snapshot.weapon_class)
    local cof = string.format("data/global/monsters/%s/COF/%s%s%s.cof",
        token, token, mode, weapon_class)
    -- A mod can advertise a semantic mode without shipping its optional legacy
    -- animation. Missing presentation data must not terminate simulation.
    if not render.asset_exists(cof) then return nil end
    local info = assert(render.cof_info(cof))
    local authored = decode_components(snapshot.components)
    local components = {}
    for _, layer in ipairs(info.layers) do
        local component = upper(layer.type)
        local visual = authored[component] or "LIT"
        local path = component_path(token, component, visual, mode, upper(layer.weapon_class))
        if render.asset_exists(path) then components[component] = path end
    end
    local timing = render.animdata_info(token .. mode .. weapon_class)
    local values = {}
    for component, path in pairs(components) do values[#values + 1] = component .. "=" .. path end
    table.sort(values)
    local logical_to_cof = {
        [8] = {1, 3, 5, 7, 0, 2, 4, 6},
        [16] = {2, 6, 10, 14, 0, 4, 8, 12, 1, 3, 5, 7, 9, 11, 13, 15},
    }
    local direction = snapshot.direction or 0
    if logical_to_cof[info.directions] then
        direction = logical_to_cof[info.directions][direction + 1]
    end
    return {
        cof = cof,
        palette = "data/global/palette/units/pal.dat",
        direction = direction,
        components = components,
        rate = timing and timing.speed or 128,
        frames = timing and timing.frames or info.frames,
        events = timing and timing.events or info.events,
        mode = mode,
        key = table.concat({cof, tostring(direction), table.concat(values, ":")}, "|"),
    }
end

M.preload_request = player_composite.preload_request
M.new_playback = player_composite.new_playback
M.advance = player_composite.advance

return M
