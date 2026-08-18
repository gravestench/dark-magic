-- Resolve semantic timed states into record-authored overlay recipes.
--
-- Authority stores only the state ID and target relationship. Presentation
-- follows States.txt references into Overlay.txt, so neither this adapter nor
-- the renderer needs to recognize Frozen Armor, Enchant, or a specific curse.

local compat = require("d2legacy.ui.compat")

local M = {}
local catalog

local function index(rows, key)
    local result = {}
    for _, row in ipairs(rows or {}) do
        local value = row[key]
        if value and value ~= "" then
            result[value] = row
        end
    end
    return result
end

local function integer(row, key, fallback)
    local value = tonumber(row[key])
    return value and math.floor(value) or fallback
end

local function recipe(overlays, id, layer, loop)
    if not id or id == "" then
        return nil
    end
    local row = assert(overlays[id], "state references missing overlay " .. id)
    local filename = assert(row.Filename and row.Filename ~= "" and row.Filename, "overlay has no Filename " .. id)
    local frames = math.max(integer(row, "Frames", 1), 1)
    local frames_per_second = math.max(integer(row, "AnimRate", 16), 1)
    return {
        id = id,
        path = "data/global/overlays/" .. filename .. ".dcc",
        palette = "data/global/palette/units/pal.dat",
        directions = math.max(integer(row, "NumDirections", 1), 1),
        frames = frames,
        frames_per_second = frames_per_second,
        duration_seconds = frames / frames_per_second,
        loop = loop,
        layer = layer,
        predraw = row.PreDraw == "1",
        blend = compat.draw_mode(integer(row, "Trans", 0)),
        offset_x = integer(row, "Xoffset", 0),
        offset_y = integer(row, "Yoffset", 0),
    }
end

function M.from_rows(states, overlay_rows, state_id)
    local state = index(states, "state")[state_id]
    if not state then
        return nil
    end
    local overlays = index(overlay_rows, "overlay")
    local active = {}
    local back = recipe(overlays, state.overlay2, "back", true)
    local front = recipe(overlays, state.overlay1, "front", true)
    if back then
        active[#active + 1] = back
    end
    if front then
        active[#active + 1] = front
    end
    return {
        state_id = state_id,
        active = active,
        applied = recipe(overlays, state.castoverlay, "front", false),
        removed = recipe(overlays, state.removerlay, "front", false),
    }
end

local function load_catalog()
    if catalog then
        return catalog
    end
    local records = require("engine.records/v1")
    local states = assert(records.load("data/global/excel/States.txt"))
    local overlays = assert(records.load("data/global/excel/Overlay.txt"))
    catalog = { states = states, overlays = overlays, overlay_index = index(overlays, "overlay"), recipes = {} }
    return catalog
end

function M.resolve(state_id)
    local loaded = load_catalog()
    if loaded.recipes[state_id] == nil then
        loaded.recipes[state_id] = M.from_rows(loaded.states, loaded.overlays, state_id) or false
    end
    return loaded.recipes[state_id] or nil
end

function M.overlay(overlay_id, layer, loop)
    if not overlay_id or overlay_id == "" then
        return nil
    end
    local loaded = load_catalog()
    return recipe(loaded.overlay_index, overlay_id, layer or "front", loop == true)
end

return M
