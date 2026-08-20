local data = require("engine.data/v1")
local render = require("engine.render/v1")

local manifest = assert(data.load("darkmagic/ds1-editor/ui/assets.json"))
local assets = {}

local function dimensions(group, name)
    local size = group.sizes and group.sizes[name]
    if size then return size[1], size[2] end
    return group.tile_size, group.tile_size
end

-- Resolve one semantic icon name without leaking DC6 frame numbers into screens.
function assets.definition(sheet, name, variant)
    local group = assert(manifest.sheets[sheet], "unknown editor UI sheet: " .. tostring(sheet))
    local frames = assert(group.frames[name], "unknown editor UI frame: " .. tostring(name))
    local frame = frames
    if type(frames) == "table" then
        local index = (math.max(1, variant or 1) - 1) % #frames + 1
        frame = frames[index]
    end
    local width, height = dimensions(group, name)
    return group.path, manifest.palette, frame, width, height
end

-- Return every compatible frame for deterministic native-pixel variation.
function assets.variants(sheet, name)
    local group = assert(manifest.sheets[sheet], "unknown editor UI sheet: " .. tostring(sheet))
    local value = assert(group.frames[name], "unknown editor UI frame: " .. tostring(name))
    local frames = type(value) == "table" and value or {value}
    local width, height = dimensions(group, name)
    return group.path, manifest.palette, frames, width, height
end

-- Create a retained icon node with the editor palette and its authored 32×32 frame.
function assets.create(parent, sheet, name, x, y, z, variant)
    local path, palette, frame = assets.definition(sheet, name, variant)
    local node = render.create("hud", parent)
    local width, height = node:set_dc6(path, palette, 0, frame)
    node:set_position(x, y)
    node:set_z(z or 0)
    return node, width, height
end

-- Reuse an existing node for a new semantic icon while preserving its layout ownership.
function assets.apply(node, sheet, name, variant)
    local path, palette, frame = assets.definition(sheet, name, variant)
    return node:set_dc6(path, palette, 0, frame)
end

return assets
