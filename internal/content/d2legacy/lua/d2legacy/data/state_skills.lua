-- Decode explicitly supported targetless, timed self-state skills.

local records = require("engine.records/v1")
local M = {}

local function index(rows, column)
    local result = {}
    for _, row in ipairs(rows) do
        local key = row[column]
        if key and key ~= "" then
            result[key] = row
        end
    end
    return result
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported self-state skill IDs are required")
    local skills = index(assert(records.load("data/global/excel/skills.txt")), "Id")
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        -- Narrow module fixtures do not have to reproduce every declared
        -- retail row. The mounted-data coverage report is the strict target
        -- completeness gate; any row present here is still decoded by ID.
        if row then
            local id = assert(tonumber(row.Id), "self-state skill ID is required")
            id = math.floor(id)
            assert(not result[id], "duplicate self-state skill ID")
            result[id] = {
                behavior = "state.self-timed",
                state_id = "skill." .. tostring(id),
                -- Param1 is the base duration for this focused family. Exact
                -- 1.14d state/stat semantics remain explicitly evidence-gated.
                duration = math.max(1, math.floor(tonumber(row.Param1) or 600)),
            }
        end
    end
    return result
end

return M
