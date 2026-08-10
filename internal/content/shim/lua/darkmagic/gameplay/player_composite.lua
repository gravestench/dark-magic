-- Translate authoritative player appearance state into legacy COF/DCC assets.
--
-- The ECS tells us WHAT the player is doing (NU/WL/RN), where they face, and
-- which equipment family they use. This adapter knows HOW Diablo II spells the
-- corresponding asset paths. Keeping that spelling here means neither input nor
-- simulation code needs to know about MPQ directory layouts.

local render = require("dm.render/v1")

local M = {}

local body_components = {
    HD = true, -- head
    TR = true, -- torso
    LG = true, -- legs
    RA = true, -- right arm
    LA = true, -- left arm
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
function M.unarmed(authority)
    local token = upper(authority.token)
    local mode = upper(authority.mode)
    local weapon_class = upper(authority.weapon_class)
    local cof = cof_path(token, mode, weapon_class)
    local info = assert(render.cof_info(cof))
    local components = {}

    for _, layer in ipairs(info.layers) do
        local component = upper(layer.type)
        if body_components[component] then
            components[component] = component_path(
                token,
                component,
                "LIT",
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
        -- Riiablo confirms player locomotion rate is coupled to authoritative
        -- velocity (32*walk speed, 16*run speed). NU's temporary fallback is
        -- replaced when AnimData.d2 becomes a typed runtime catalog.
        rate = mode == "WL" and 320 or (mode == "RN" and 240 or 128),
        -- A value-only change key prevents rebuilding the retained animation on
        -- every presentation frame. Position updates remain independent.
        key = table.concat({token, mode, weapon_class, tostring(authority.direction)}, ":"),
    }
end

return M
