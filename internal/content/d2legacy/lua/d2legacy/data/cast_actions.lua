-- Resolve shared cast-action facts for every explicitly admitted skill.
--
-- Behavior-family decoders remain responsible for authority. This join owns
-- the Skills.txt fields shared by those families so SC/SQ action selection,
-- cast overlays, client cues, and sounds do not grow per-skill branches.

local records = require("engine.records/v1")

local M = {}

local function index(rows)
    local result = {}
    for _, row in ipairs(rows) do
        local id = tonumber(row.Id)
        if id then
            result[math.floor(id)] = row
        end
    end
    return result
end

local function integer_or_blank(row, field)
    local raw = row[field]
    if raw == nil or raw == "" then
        return 0
    end
    local value = tonumber(raw)
    assert(value and value == math.floor(value), "skill " .. tostring(row.Id) .. " has invalid " .. field)
    return value
end

function M.load(admitted)
    assert(type(admitted) == "table", "admitted skill definitions are required")
    local skills = index(assert(records.load("data/global/excel/skills.txt")))
    local result = {}
    for skill_id in pairs(admitted) do
        local row = assert(skills[skill_id], "admitted skill is absent from Skills.txt: " .. skill_id)
        local mode = row.anim or ""
        assert(mode == "" or #mode == 2, "skill " .. skill_id .. " has invalid animation mode")
        result[skill_id] = {
            skill_id = skill_id,
            animation_mode = mode,
            sequence_transition = row.seqtrans or "",
            sequence_number = integer_or_blank(row, "seqnum"),
            use_attack_rate = row.UseAttackRate == "1",
            start_sound = row.stsound or "",
            do_sound = row.dosound or "",
            cast_overlay = row.castoverlay or "",
            client_missile = row.cltmissile or "",
            client_missile_a = row.cltmissilea or "",
            client_missile_b = row.cltmissileb or "",
            client_missile_c = row.cltmissilec or "",
        }
    end
    return result
end

return M
