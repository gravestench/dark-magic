-- Interpret Diablo II Skills, SkillDesc, and CharStats records.
--
-- The engine exposes rows as immutable strings. This module gives those
-- strings their Diablo meaning. Keeping that interpretation here means a
-- different mod can use the same record reader without inheriting classes,
-- starting skills, icon sheets, or mouse-button rules.

local records = require("engine.records/v1")

local M = {}

local icon_sheets = {
    ama = "AmSkillicon",
    sor = "SoSkillicon",
    nec = "NeSkillicon",
    pal = "PaSkillicon",
    bar = "BaSkillicon",
    dru = "DrSkillicon",
    ass = "AsSkillicon",
}

local function find(rows, key, wanted)
    for _, row in ipairs(rows) do
        if row[key] == wanted then
            return row
        end
    end
    return nil
end

local function descriptions_by_name()
    local result = {}
    for _, row in ipairs(records.load("data/global/excel/skilldesc.txt")) do
        local name = row.skilldesc
        if name and name ~= "" then
            result[name] = row
        end
    end
    return result
end

local function class_starting_skill(class)
    local wanted = string.lower(assert(class, "class is required"))
    for _, row in ipairs(records.load("data/global/excel/charstats.txt")) do
        if string.lower(row.class or "") == wanted then
            return row.StartSkill or ""
        end
    end
    return ""
end

local function is_initial_skill(skill, starting_skill)
    return skill.general == "1" or string.lower(skill.skill or "") == string.lower(starting_skill)
end

local function learned_skill(skill, description)
    local id = tonumber(skill.Id)
    local list_row = description and tonumber(description.ListRow)
    if not id or not list_row or list_row < 0 or skill.passive == "1" then
        return nil
    end
    return {
        id = math.floor(id),
        level = 1,
        list_row = math.floor(list_row),
        left_allowed = skill.leftskill == "1",
        right_allowed = true,
    }
end

-- load returns presentation facts for one numeric skill identifier.
function M.load(id)
    local skill = find(records.load("data/global/excel/skills.txt"), "Id", tostring(id))
    if not skill then
        return nil
    end

    local description = find(records.load("data/global/excel/skilldesc.txt"), "skilldesc", skill.skilldesc)
    if not description then
        return nil
    end

    local icon = tonumber(description.IconCel)
    if not icon or icon < 0 then
        return nil
    end

    local sheet = icon_sheets[string.lower(skill.charclass or "")] or "Skillicon"
    return {
        id = id,
        icon = math.floor(icon),
        sheet = "data/global/ui/SPELLS/" .. sheet .. ".DC6",
        name_key = description["str name"] or "",
        short_key = description["str short"] or "",
        list_row = math.floor(tonumber(description.ListRow) or 0),
        left_allowed = skill.leftskill == "1",
        passive = skill.passive == "1",
        weapon_selection = math.floor(tonumber(skill.weapsel) or 0),
    }
end

-- starting_for_class builds the authoritative initial learned-skill set.
-- Go imports the durable class name but does not interpret these D2 tables.
function M.starting_for_class(class)
    local starting_skill = class_starting_skill(class)
    local descriptions = descriptions_by_name()
    local result = {}

    for _, skill in ipairs(records.load("data/global/excel/skills.txt")) do
        if is_initial_skill(skill, starting_skill) then
            local learned = learned_skill(skill, descriptions[skill.skilldesc])
            if learned then
                result[#result + 1] = learned
            end
        end
    end

    table.sort(result, function(left, right)
        return left.id < right.id
    end)
    return result
end

-- learned_for_ids resolves an explicit development/evidence fixture through
-- the same owned Skills/SkillDesc records as ordinary starting skills. It does
-- not make an unknown, passive, or unassignable row executable.
function M.learned_for_ids(ids, level)
    assert(type(ids) == "table", "skill IDs must be a table")
    level = level or 1
    assert(level >= 1 and level == math.floor(level), "skill level must be a positive integer")
    local skills_by_id = {}
    for _, skill in ipairs(records.load("data/global/excel/skills.txt")) do
        local id = tonumber(skill.Id)
        if id then
            skills_by_id[id] = skill
        end
    end
    local descriptions = descriptions_by_name()
    local result, seen = {}, {}
    for _, id in ipairs(ids) do
        assert(type(id) == "number" and id >= 0 and id == math.floor(id), "skill ID must be a non-negative integer")
        assert(not seen[id], "duplicate skill ID " .. id)
        seen[id] = true
        local skill = assert(skills_by_id[id], "unknown skill ID " .. id)
        local learned = assert(learned_skill(skill, descriptions[skill.skilldesc]), "skill ID is not learnable " .. id)
        learned.level = level
        result[#result + 1] = learned
    end
    table.sort(result, function(left, right)
        return left.id < right.id
    end)
    return result
end

return M
