-- Interpret Diablo II Sounds.txt records and ask the generic audio capability
-- to play the resulting asset. The engine owns decoding, mixing, buses, and
-- sound-handle lifetime; this mod owns redirects, grouped variants, path
-- conventions, and the meaning of legacy flags.

local engine_audio = require("engine.audio/v1")
local records = require("engine.records/v1")
local M = {}
local path = "data/global/excel/Sounds.txt"

local function number(row, ...)
    for index = 1, select("#", ...) do
        local value = row[select(index, ...)]
        if value ~= nil and value ~= "" then return tonumber(value) or 0 end
    end
    return 0
end

local function text(row, ...)
    for index = 1, select("#", ...) do
        local value = row[select(index, ...)]
        if value ~= nil and value ~= "" then return tostring(value) end
    end
    return ""
end

local function choose_group(rows, first, count, seed)
    local total = 0
    for index = first, math.min(#rows, first + count - 1) do
        total = total + math.max(1, number(rows[index], "Group Weight", "GroupWeight"))
    end
    local pick = seed % total
    for index = first, math.min(#rows, first + count - 1) do
        local weight = math.max(1, number(rows[index], "Group Weight", "GroupWeight"))
        if pick < weight then return rows[index] end
        pick = pick - weight
    end
    return rows[first]
end

local function asset_path(row)
    local file = text(row, "FileName", "filename"):gsub("\\", "/"):gsub("^/", "")
    local candidates = {file}
    if number(row, "IsLocal", "islocal") ~= 0 then candidates[#candidates + 1] = "data/local/sfx/" .. file end
    if number(row, "IsMusic", "ismusic") ~= 0 then candidates[#candidates + 1] = "data/global/music/" .. file end
    candidates[#candidates + 1] = "data/global/sfx/" .. file
    for _, candidate in ipairs(candidates) do
        if engine_audio.exists(candidate) then return candidate end
    end
    error("sound asset is unavailable: " .. file)
end

local function options(row, requested, seed)
    local minimum = math.max(0, number(row, "Volume Min", "VolumeMin"))
    local maximum = number(row, "Volume Max", "VolumeMax")
    if maximum <= 0 then maximum = 255 end
    minimum = math.min(minimum, maximum)
    local volume = minimum + (seed % (maximum - minimum + 1))
    local channel = string.lower(text(row, "Channel", "channel"))
    local music = number(row, "IsMusic", "ismusic") ~= 0
    local bus = "sfx"
    if music then bus = "music"
    elseif number(row, "IsUI", "isui") ~= 0 then bus = "ui"
    elseif number(row, "IsAmbientScene", "IsAmbientEvent") ~= 0 then bus = "ambience"
    elseif channel:find("voice") or channel:find("speech") or channel:find("vocal") then bus = "speech" end
    return {bus=bus, volume=volume/255, loop=number(row,"Loop","loop")~=0,
        stream=music or number(row,"Stream","stream")~=0, group=requested}
end

function M.play_record(name, seed)
    seed = math.max(0, math.floor(tonumber(seed) or 0))
    local rows = assert(records.load(path))
    local first
    for index, row in ipairs(rows) do
        if string.lower(text(row, "Sound", "sound")) == string.lower(name) then first = index; break end
    end
    assert(first, "unknown sound record: " .. tostring(name))
    local row = rows[first]
    local redirect = tonumber(text(row, "Redirect", "redirect"))
    if redirect and rows[redirect + 1] then row = rows[redirect + 1] end
    local group = number(row, "Group Size", "GroupSize")
    if group > 1 then row = choose_group(rows, first, group, seed) end
    return engine_audio.play(asset_path(row), options(row, name, seed))
end

return M
