-- Translate authoritative player appearance state into legacy COF/DCC assets.
--
-- The ECS tells us WHAT the player is doing (NU/WL/RN), where they face, and
-- which equipment family they use. This adapter knows HOW Diablo II spells the
-- corresponding asset paths. Keeping that spelling here means neither input nor
-- simulation code needs to know about MPQ directory layouts.

local render = require("dm.render/v1")

local M = {}

local default_appearance = {
    HD = "LIT", TR = "LIT", LG = "LIT", RA = "LIT", LA = "LIT",
}

local function upper(value)
    return string.upper(assert(value))
end

local function cof_path(token, mode, weapon_class)
    return string.format(
        "data/global/chars/%s/COF/%s%s%s.cof",
        token, token, mode, weapon_class
    )
end

local function component_path(token, component, appearance, mode, weapon_class)
    return string.format(
        "data/global/chars/%s/%s/%s%s%s%s%s.dcc",
        token, component, token, component, appearance, mode, weapon_class
    )
end

-- Resolve the deliberately small first slice: an unequipped player whose body
-- pieces use Diablo II's default "lit" appearance. Hands and shields are absent.
-- The COF layer supplies each component's own weapon-class suffix; those suffixes
-- are not always the same as the HTH selector in the COF filename.
local function equipped_appearance(items)
    local appearance, hand_classes = {}, {}
    if not items then return appearance, nil end
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
        local appearance = equipped[component] or default_appearance[component]
        if appearance then
            components[component] = component_path(
                token,
                component,
                appearance,
                mode,
                upper(layer.weapon_class)
            )
        end
    end

    return {
        cof = cof,
        palette = authority.palette,
        direction = authority.direction,
        components = components,
        -- AnimData.d2, not the COF header, owns player timing and frame events.
        -- The fallback only keeps modded archives with a missing record usable.
        rate = timing and timing.speed or (mode == "WL" and 320 or (mode == "RN" and 240 or 128)),
        frames = timing and timing.frames or info.frames,
        events = timing and timing.events or info.events,
        mode = mode,
        -- A value-only change key prevents rebuilding the retained animation on
        -- every presentation frame. Position updates remain independent.
        key = table.concat({token, mode, weapon_class, tostring(authority.direction), cof}, ":")
            .. ":" .. table.concat((function()
                local values = {}
                for component, path in pairs(components) do values[#values + 1] = component .. "=" .. path end
                table.sort(values)
                return values
            end)(), ":"),
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
function M.unarmed(authority) return M.resolve(authority, nil) end

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
    return { mode = composite.mode, frame = 1, remainder = 0, seconds = 0 }
end

-- Advance across every crossed frame, not merely the final one. A long frame
-- can therefore propagate attack/missile/sound/skill markers without dropping
-- intermediate events. Gameplay consequences still require authoritative
-- command/system handling; this stream synchronizes presentation consumers.
function M.advance(playback, composite, elapsed)
    local crossed = {}
    if elapsed <= 0 or composite.rate <= 0 or composite.frames <= 0 then return crossed end
    local frame_seconds = 256 / (composite.rate * 25)
    playback.seconds = playback.seconds + elapsed
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

return M
