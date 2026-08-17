-- Translate authoritative player appearance state into legacy COF/DCC assets.
--
-- The ECS tells us WHAT the player is doing (NU/WL/RN), where they face, and
-- which equipment family they use. This adapter knows HOW Diablo II spells the
-- corresponding asset paths. Keeping that spelling here means neither input nor
-- simulation code needs to know about MPQ directory layouts.

local render = require("engine.render/v1")

local M = {}

local function upper(value)
    return string.upper(assert(value))
end

-- Simulation uses a readable clockwise eight-way direction. Legacy DC/COF
-- files store directions in a peculiar interleaved order. Riiablo performs
-- this same conversion before both selecting a component direction and reading
-- its COF priority row; skipping it can pair a visible facing with the wrong
-- arm/head ordering table.
local cof_directions = {
    -- These are Riiablo DC.Direction.toRealDir and the eight sampled values
    -- from OpenDiablo2 Dir64ToCof. Simulation's readable directions are
    -- SOUTH/WEST/NORTH/EAST/DOWN/LEFT/UP/RIGHT; COF priority rows are not.
    [8] = { 1, 3, 5, 7, 0, 2, 4, 6 },
    [16] = { 2, 6, 10, 14, 0, 4, 8, 12, 1, 3, 5, 7, 9, 11, 13, 15 },
}

-- COF rows are sequential angle buckets; DCC files use the legacy interleave.
local dcc_directions = {
    [8] = { 4, 0, 5, 1, 6, 2, 7, 3 },
    [16] = { 4, 8, 0, 9, 5, 10, 1, 11, 6, 12, 2, 13, 7, 14, 3, 15 },
}

local function cof_direction(direction, count, space)
    if space == "encoded" then
        return direction
    end
    local lookup = cof_directions[count]
    if not lookup then
        return direction
    end
    return assert(lookup[direction + 1], "semantic direction is out of range")
end

local function cof_path(token, mode, weapon_class)
    return string.format("data/global/chars/%s/COF/%s%s%s.cof", token, token, mode, weapon_class)
end

local function component_path(token, component, appearance, mode, weapon_class)
    return string.format(
        "data/global/chars/%s/%s/%s%s%s%s%s.dcc",
        token,
        component,
        token,
        component,
        appearance,
        mode,
        weapon_class
    )
end

-- Resolve equipment overrides. For every remaining COF layer, the resolver
-- later probes Diablo II's default "lit" appearance. The COF layer supplies
-- each component's own weapon-class suffix; those suffixes are not always the
-- same as the HTH selector in the COF filename.
local function equipped_appearance(items)
    local appearance, hand_classes = {}, {}
    if not items then
        return appearance, nil
    end
    for _, item in ipairs(items.items or {}) do
        local hand = item.slot == "rarm" or item.slot == "larm"
        local active = not hand or item.weapon_set == items.active_weapon_set
        if item.container == "equipment" and active then
            for component, code in pairs(item.composite or {}) do
                appearance[upper(component)] = upper(code)
            end
            if hand and item.weapon_class and item.weapon_class ~= "" then
                hand_classes[item.slot] = upper(item.weapon_class)
            end
        end
    end
    -- Snapshot iteration order must never decide the composite. Diablo's main
    -- hand is the stable selector; the off hand is only the fallback when the
    -- main hand has no weapon-class recipe.
    return appearance, hand_classes.rarm or hand_classes.larm
end

local function resolve_appearance(authority, equipped, equipped_weapon_class)
    local token = upper(authority.token)
    local mode = upper(authority.mode)
    local weapon_class = equipped_weapon_class or upper(authority.weapon_class)
    local cof = cof_path(token, mode, weapon_class)
    local info = assert(render.cof_info(cof))
    local timing = render.animdata_info(token .. mode .. weapon_class)
    local components = {}

    for _, layer in ipairs(info.layers) do
        local component = upper(layer.type)
        local appearance = equipped[component]
        if appearance then
            components[component] = component_path(token, component, appearance, mode, upper(layer.weapon_class))
        else
            -- OpenDiablo2 tries the default LIT variant for every authored COF
            -- layer. This matters for class-specific body pieces: Necromancer
            -- S1/S2 are real forearm/overlay DCCs, while his empty SH layer has
            -- no file and must simply be skipped.
            local candidate = component_path(token, component, "LIT", mode, upper(layer.weapon_class))
            if render.asset_exists(candidate) then
                components[component] = candidate
            end
        end
    end

    local direction = cof_direction(authority.direction, info.directions, authority.direction_space)

    -- Expansion player locomotion overrides AnimData speed with fixed WL/RN
    -- bases scaled by the same effective velocity percentage as path motion.
    -- AnimData remains authoritative for frame count and frame events.
    local rate = timing and timing.speed or (mode == "WL" and 320 or (mode == "RN" and 240 or 128))
    if authority.class and (mode == "WL" or mode == "RN") then
        local movement_rules = require("d2legacy.movement_rules/v1")
        rate = movement_rules.animation_rate(
            authority.class,
            mode == "RN",
            authority.velocitypercent or 0,
            authority.item_fastermovevelocity or 0
        )
    end

    return {
        cof = cof,
        palette = authority.palette,
        direction = direction,
        dcc_direction = dcc_directions[info.directions] and dcc_directions[info.directions][direction + 1] or direction,
        components = components,
        rate = rate,
        frames = timing and timing.frames or info.frames,
        events = timing and timing.events or info.events,
        mode = mode,
        -- A value-only change key prevents rebuilding the retained animation on
        -- every presentation frame. Position updates remain independent.
        key = table.concat({ token, mode, weapon_class, tostring(direction), cof, tostring(rate) }, ":")
            .. ":"
            .. table.concat(
                (function()
                    local values = {}
                    for component, path in pairs(components) do
                        values[#values + 1] = component .. "=" .. path
                    end
                    table.sort(values)
                    return values
                end)(),
                ":"
            ),
    }
end

function M.resolve(authority, items)
    local equipped, equipped_weapon_class = equipped_appearance(items)
    return resolve_appearance(authority, equipped, equipped_weapon_class)
end

-- Development/probe entry point. The caller supplies the same validated recipe
-- shape item authority normally publishes, without pretending those values are
-- authoritative gameplay state.
function M.recipe(authority, appearance, weapon_class)
    return resolve_appearance(authority, appearance or {}, weapon_class)
end

-- Compatibility name for callers that deliberately want the empty-equipment recipe.
function M.unarmed(authority)
    return M.resolve(authority, nil)
end

-- Describe the expensive, CPU-side half of a composite update. The generic
-- preloader can execute this away from the Lua/render thread and queue complete
-- frames for bounded texture upload. Callers keep their previous animation
-- visible until this request reports ready.
function M.preload_request(composite)
    return {
        kind = "cof_animation",
        path = composite.cof,
        palette = composite.palette,
        direction = composite.direction,
        components = composite.components,
    }
end

-- Create presentation playback state for one authoritative animation mode.
-- Facing and equipment changes reuse this object, so they do not restart time.
function M.new_playback(composite)
    return {
        mode = composite.mode,
        frame = 1,
        remainder = 0,
        seconds = 0,
        authority_seconds = 0,
        rate = composite.rate,
    }
end

local function preserve_frame_phase(playback, composite)
    if playback.rate and playback.rate > 0 and playback.rate ~= composite.rate and playback.remainder > 0 then
        local previous_frame_seconds = 256 / (playback.rate * 25)
        local next_frame_seconds = 256 / (composite.rate * 25)
        local phase = playback.remainder / previous_frame_seconds
        playback.remainder = phase * next_frame_seconds
        -- The renderer seeks by elapsed seconds at the new rate. Convert to an
        -- equivalent time within the retained loop so it observes the same
        -- frame/phase rather than applying the new rate to the mode's history.
        playback.seconds = ((playback.frame - 1) + phase) * next_frame_seconds
    end
    playback.rate = composite.rate
end

-- Advance across every crossed frame, not merely the final one. A long frame
-- can therefore propagate attack/missile/sound/skill markers without dropping
-- intermediate events. Gameplay consequences still require authoritative
-- command/system handling; this stream synchronizes presentation consumers.
function M.advance(playback, composite, elapsed)
    local crossed = {}
    if elapsed <= 0 or composite.rate <= 0 or composite.frames <= 0 then
        return crossed
    end
    preserve_frame_phase(playback, composite)
    local frame_seconds = 256 / (composite.rate * 25)
    playback.seconds = playback.seconds + elapsed
    playback.authority_seconds = (playback.authority_seconds or 0) + elapsed
    playback.remainder = playback.remainder + elapsed
    while playback.remainder >= frame_seconds do
        playback.remainder = playback.remainder - frame_seconds
        playback.frame = (playback.frame % composite.frames) + 1
        local event = composite.events[playback.frame]
        if event and event ~= 0 then
            crossed[#crossed + 1] = { frame = playback.frame, event = event }
        end
    end
    return crossed
end

-- The first network sample seeks from mode start. Later samples integrate the
-- authoritative time delta at the currently projected rate, preserving phase
-- when FRW/chill changes without pretending the latest rate applied to the
-- mode's entire history.
function M.synchronize(playback, composite, seconds)
    local crossed = {}
    if composite.rate <= 0 or composite.frames <= 0 then
        return crossed
    end
    seconds = math.max(0, seconds or 0)
    if playback.synchronized and seconds >= playback.authority_seconds then
        local result = M.advance(playback, composite, seconds - playback.authority_seconds)
        playback.authority_seconds = seconds
        return result
    end
    preserve_frame_phase(playback, composite)
    local frame_seconds = 256 / (composite.rate * 25)
    local current = math.floor(seconds / frame_seconds)
    playback.seconds = seconds
    playback.authority_seconds = seconds
    playback.remainder = seconds - current * frame_seconds
    playback.frame = (current % composite.frames) + 1
    playback.synchronized = true
    return crossed
end

return M
