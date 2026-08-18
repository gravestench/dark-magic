-- Resolve semantic cast cues through Skills.txt and the shared Overlay.txt
-- adapter. The renderer never recognizes a skill ID, class, or spell name.

local records = require("engine.records/v1")
local overlays = require("d2legacy.gameplay.state_overlay_presentation")

local M = {}
local skills

local function load_skills()
    if skills then
        return skills
    end
    skills = {}
    for _, row in ipairs(assert(records.load("data/global/excel/skills.txt"))) do
        local id = tonumber(row.Id)
        if id then
            skills[math.floor(id)] = row
        end
    end
    return skills
end

function M.resolve(skill_id)
    local row = load_skills()[skill_id]
    if not row then
        return nil
    end
    return {
        overlay = overlays.overlay(row.castoverlay, "front", false),
        start_sound = row.stsound or "",
        effect_sound = row.dosound or "",
    }
end

return M
