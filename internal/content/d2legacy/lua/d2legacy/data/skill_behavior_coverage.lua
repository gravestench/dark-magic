-- Load the target-locked, explicitly reviewed skill-to-family declarations.
--
-- A declaration admits one exact skill ID to one reusable implementation
-- family. Similar Skills.txt function IDs or missile rows never admit another
-- skill implicitly.

local data = require("engine.data/v1")

local M = {}
local schema = "d2legacy.skill_behavior_coverage/v1"
local target = "diablo-ii-lod-1.14d-expansion"

function M.load()
    local manifest = assert(data.load("manifests/skill-behavior-coverage.v1.json"))
    assert(manifest.schema == schema, "skill behavior coverage schema is unsupported")
    assert(manifest.version == 1, "skill behavior coverage version is unsupported")
    assert(manifest.target == target, "skill behavior coverage target is unsupported")

    local result = { by_id = {}, by_family = {} }
    for _, declaration in ipairs(assert(manifest.implementations)) do
        local skill_id = assert(tonumber(declaration.skill_id), "skill behavior declaration ID is required")
        skill_id = math.floor(skill_id)
        local family = assert(declaration.family, "skill behavior declaration family is required")
        assert(skill_id >= 0 and not result.by_id[skill_id], "duplicate skill behavior declaration")
        assert(
            type(declaration.evidence_status) == "string" and declaration.evidence_status ~= "",
            "skill behavior declaration evidence is required"
        )
        result.by_id[skill_id] = declaration
        result.by_family[family] = result.by_family[family] or {}
        table.insert(result.by_family[family], skill_id)
    end
    return result
end

return M
