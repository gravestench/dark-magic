-- Decode explicitly admitted corpse-targeted summons from Expansion Skills,
-- PetType, and modifier-skill records. Exact IDs admit content; record shape
-- selects reusable materialization/stat policies inside the family.

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

local function integer(row, column, label, minimum)
    local value = tonumber(row[column])
    assert(value and value == math.floor(value) and value >= (minimum or 0), label .. " has invalid " .. column)
    return value
end

local function truth(row, column)
    return row[column] == "1"
end

local function referenced_skill(expression, label)
    local name = expression:match("skill%('([^']+)'%.lvl%)")
    assert(name, label .. " has unsupported modifier formula " .. expression)
    return name
end

local function summon_resist(skills, class, label)
    local found
    for _, row in ipairs(skills) do
        if row.charclass == class and row.passivestat1 == "passive_summon_resist" and row.passivecalc1 == "dm12" then
            assert(not found, label .. " has ambiguous summon-resistance modifier")
            found = row
        end
    end
    assert(found, label .. " has no summon-resistance modifier")
    return {
        skill_id = integer(found, "Id", label .. " summon resistance"),
        minimum = integer(found, "Param1", label .. " summon resistance"),
        maximum = integer(found, "Param2", label .. " summon resistance"),
    }
end

local function common(row, pet, skills, label)
    local shift = integer(row, "manashift", label)
    return {
        behavior = "summon.targeted-corpse",
        skill_id = integer(row, "Id", label),
        mana_cost_raw = integer(row, "mana", label) * (2 ^ shift),
        mana_cost_per_level_raw = integer(row, "lvlmana", label) * (2 ^ shift),
        minimum_mana_cost_raw = integer(row, "minmana", label) * (2 ^ shift),
        effect_delay = 1,
        complete_delay = 2,
        requires_corpse_target = true,
        summon_mode = row.summode,
        category = row.pettype,
        category_group = tonumber(pet.group) or 0,
        category_base_max = math.max(tonumber(pet.basemax) or 0, 0),
        unsummon = truth(pet, "unsummon"),
        warp_with_owner = truth(pet, "warp"),
        range_limited = truth(pet, "range"),
        summon_resist = summon_resist(skills, row.charclass, label),
        life_percent_base = 0,
        life_percent_per_level_after_three = 0,
        mastery_life_flat_per_level = 0,
        mastery_life_percent_per_level = 0,
        damage_percent_per_level_after_three = 0,
        mastery_damage_flat_per_level = 0,
        mastery_damage_percent_per_level = 0,
        attack_rating_per_combined_level = 0,
        defense_per_combined_level = 0,
        velocity_percent = 0,
        duration_base_ticks = 0,
        duration_per_level_ticks = 0,
        granted_skill_name = "",
        granted_skill_level_policy = "none",
    }
end

local function decode_fixed_pet(row, skills, skills_by_name, pet, label)
    assert(row.srvstfunc == "15" and row.srvdofunc == "31", label .. " has unsupported fixed-pet functions")
    assert(row.SelectProc == "2", label .. " has unsupported fixed-pet selection policy")
    assert(row.summon ~= "" and row.summode ~= "", label .. " has incomplete fixed-pet records")
    assert(row.petmax == "(lvl < 4) ?lvl:(2+lvl/3)", label .. " has unsupported pet-limit formula")
    assert(row.calc1 == "(lvl < 4) ? 0 : (par2 * (lvl - 3))", label .. " has unsupported life formula")

    local mastery_name = referenced_skill(row.passivecalc1, label)
    local mastery = assert(skills_by_name[mastery_name], label .. " references an unknown mastery skill")
    assert(
        row.passivestat1 == "maxhp"
            and row.passivecalc1
                == "skill('" .. mastery_name .. "'.lvl) * skill('" .. mastery_name .. "'.par1) * 256",
        label .. " has unsupported mastery-life formula"
    )

    local definition = common(row, pet, skills, label)
    definition.materialization = "fixed_pet"
    definition.limit_policy = "tiered"
    definition.summon_monster_id = row.summon
    definition.mastery_skill_id = integer(mastery, "Id", label .. " mastery")
    definition.mastery_life_flat_per_level = integer(mastery, "Param1", label .. " mastery")
    definition.life_percent_per_level_after_three = integer(row, "Param2", label)

    for slot = 1, 3 do
        local stat = row["aurastat" .. slot]
        local formula = row["aurastatcalc" .. slot]
        if stat == "damagepercent" then
            assert(formula == "((lvl < 4) ? 0 : ((lvl-3)*par3))", label .. " has unsupported damage formula")
            definition.damage_percent_per_level_after_three = integer(row, "Param3", label)
        elseif stat == "tohit" then
            assert(
                formula == "(lvl+skill('" .. mastery_name .. "'.lvl))*par4",
                label .. " has unsupported attack formula"
            )
            definition.attack_rating_per_combined_level = integer(row, "Param4", label)
        elseif stat == "armorclass" then
            assert(
                formula == "(lvl+skill('" .. mastery_name .. "'.lvl))*par5",
                label .. " has unsupported defense formula"
            )
            definition.defense_per_combined_level = integer(row, "Param5", label)
        elseif stat and stat ~= "" then
            error(label .. " has unsupported summon stat " .. stat)
        end
    end

    if row.passivestat2 and row.passivestat2 ~= "" then
        assert(
            row.passivestat2 == "item_normaldamage"
                and row.passivecalc2
                    == "skill('" .. mastery_name .. "'.lvl) * skill('" .. mastery_name .. "'.par2) + edmn",
            label .. " has unsupported mastery-damage formula"
        )
        definition.mastery_damage_flat_per_level = integer(mastery, "Param2", label .. " mastery")
    end

    if row.sumskill1 and row.sumskill1 ~= "" then
        assert(
            row.sumsk1calc == "skill('" .. mastery_name .. "'.lvl) + ((lvl < 4)?0:((lvl-2)/2))",
            label .. " has unsupported granted-skill formula"
        )
        definition.granted_skill_name = row.sumskill1
        definition.granted_skill_level_policy = "mastery_plus_half_after_two"
    end
    return definition
end

local function decode_revived_corpse(row, skills, skills_by_name, pet, label)
    assert(row.srvstfunc == "21" and row.srvdofunc == "58", label .. " has unsupported revive functions")
    assert(row.SelectProc == "3" and row.summon == "", label .. " has unsupported revive selection policy")
    assert(row.petmax == "lvl" and row.summode == "NU", label .. " has unsupported revive pet policy")
    local mastery_name = referenced_skill(row.calc1, label)
    local mastery = assert(skills_by_name[mastery_name], label .. " references an unknown mastery skill")
    assert(
        row.calc1 == "par1+skill('" .. mastery_name .. "'.lvl) * skill('" .. mastery_name .. "'.par3)",
        label .. " has unsupported revive-life formula"
    )
    assert(row.calc2 == "ln34", label .. " has unsupported revive-duration formula")
    assert(
        row.aurastat1 == "damagepercent"
            and row.aurastatcalc1 == "skill('" .. mastery_name .. "'.lvl) * skill('" .. mastery_name .. "'.par4)",
        label .. " has unsupported revive-damage formula"
    )
    assert(
        row.passivestat1 == "velocitypercent" and row.passivecalc1 == "par5",
        label .. " has unsupported revive-speed formula"
    )

    local definition = common(row, pet, skills, label)
    definition.materialization = "revived_corpse"
    definition.limit_policy = "skill_level"
    definition.requires_revivable_corpse = true
    definition.summon_monster_id = ""
    definition.mastery_skill_id = integer(mastery, "Id", label .. " mastery")
    definition.life_percent_base = integer(row, "Param1", label)
    definition.mastery_life_percent_per_level = integer(mastery, "Param3", label .. " mastery")
    definition.mastery_damage_percent_per_level = integer(mastery, "Param4", label .. " mastery")
    definition.velocity_percent = integer(row, "Param5", label)
    definition.duration_base_ticks = integer(row, "Param3", label)
    definition.duration_per_level_ticks = integer(row, "Param4", label)
    return definition
end

local function decode(row, skills, skills_by_name, pets)
    local id = integer(row, "Id", "corpse summon skill")
    local label = row.skill or ("skill " .. id)
    assert(row.TargetCorpse == "1" and row.range == "none", label .. " has unsupported target policy")
    assert(row.leftskill == "" and row.InGame == "1" and row.InTown == "", label .. " has unsupported use flags")
    assert(row.pettype ~= "", label .. " has no pet category")
    local pet = assert(pets[row.pettype], label .. " references an unknown PetType")
    if row.srvdofunc == "31" then
        return decode_fixed_pet(row, skills, skills_by_name, pet, label)
    end
    if row.srvdofunc == "58" then
        return decode_revived_corpse(row, skills, skills_by_name, pet, label)
    end
    error(label .. " has unsupported corpse-summon materialization function")
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported corpse-summon skill IDs are required")
    local rows = assert(records.load("data/global/excel/skills.txt"))
    local by_id, by_name = index(rows, "Id"), index(rows, "skill")
    local admitted = {}
    for _, skill_id in ipairs(supported_ids) do
        if by_id[tostring(skill_id)] then
            admitted[#admitted + 1] = skill_id
        end
    end
    if #admitted == 0 then
        return {}
    end
    local pets = index(assert(records.load("data/global/excel/PetType.txt")), "pet type")
    local result = {}
    for _, skill_id in ipairs(admitted) do
        local definition = decode(by_id[tostring(skill_id)], rows, by_name, pets)
        assert(not result[definition.skill_id], "duplicate corpse-summon skill ID")
        result[definition.skill_id] = definition
    end
    return result
end

return M
