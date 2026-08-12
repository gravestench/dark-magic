-- Diablo II map asset identity helpers. These functions never open assets or
-- allocate renderer resources, so authoritative generation remains headless.

local M = {}

local function trim(value)
    return (tostring(value or ""):gsub("^%s+", ""):gsub("%s+$", ""))
end

function M.path(value)
    local normalized = trim(value):gsub("\\", "/"):gsub("^/+", "")
    if normalized:lower():sub(1, 18) ~= "data/global/tiles/" then
        normalized = "data/global/tiles/" .. normalized
    end
    return normalized
end

function M.preset_files(row)
    local result = {}
    local declared = math.floor(tonumber(row.Files) or 0)
    local limit = declared > 0 and math.min(declared, 6) or 6
    for index = 1, limit do
        local value = trim(row["File" .. index])
        if value ~= "" and value ~= "0" then result[#result + 1] = value end
    end
    return result
end

function M.masked_tiles(level_type, mask)
    local result = {}
    mask = math.floor(tonumber(mask) or 0)
    for index = 1, 32 do
        local bit = 2 ^ (index - 1)
        if math.floor(mask / bit) % 2 == 1 then
            local value = trim(level_type["File " .. index] or level_type["File" .. index])
            if value ~= "" and value ~= "0" then result[#result + 1] = M.path(value) end
        end
    end
    return result
end

function M.basename(value)
    return M.path(value):match("([^/]+)$") or ""
end

return M
