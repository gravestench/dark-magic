-- Decode the complete Necromancer golem family from the pinned Expansion
-- Skills, SkillDesc, PetType, and modifier-skill records. Exact IDs admit
-- content; formulas select generic stat and reaction policies.

local records = require("engine.records/v1")
local M = {}

local function index(rows, column)
    local result = {}
    for _, row in ipairs(rows) do
        local key = row[column]
        if key and key ~= "" then
            result[string.lower(key)] = row
        end
    end
    return result
end

local function integer(row, column, label, minimum)
    local value = tonumber(row[column])
    assert(value and value == math.floor(value) and value >= (minimum or 0), label .. " has invalid " .. column)
    return value
end

local function referenced(skills, name, label)
    return assert(skills[string.lower(name)], label .. " references unknown skill " .. name)
end

local function skill_id(skills, name, label)
    return integer(referenced(skills, name, label), "Id", label .. " modifier")
end

local function assert_localized_synergies(description, names, label)
    assert(description and description["dsc3texta1"] == "Sksyn", label .. " has unsupported synergy heading")
    local expected = { "Golem Mastery", "Summon Resist" }
    for _, name in ipairs(names) do
        expected[#expected + 1] = name
    end
    for index, name in ipairs(expected) do
        local key = description["dsc3texta" .. (index + 1)]
        assert(type(key) == "string" and key ~= "", label .. " has missing localized synergy key for " .. name)
    end
end

local function common(row, pet, skills, descriptions, label)
    assert(row.srvdofunc == "56" or row.srvdofunc == "57", label .. " has unsupported summon function")
    assert(row.pettype == "golem" and row.petmax == "1", label .. " has unsupported pet limit")
    assert(row.summon ~= "" and row.summode == "S1", label .. " has incomplete summon materialization")
    assert(row.range == "none" and row.InGame == "1" and row.InTown == "1", label .. " has unsupported use flags")
    assert(row.leftskill == "", label .. " must remain right-assignment only")
    assert(
        row.passivestat1 == "velocitypercent" and row.passivecalc1 == "skill('Golem Mastery'.dm34)",
        label .. " has unsupported mastery velocity"
    )
    assert(row.calc1 and row.calc1:find("skill%('Golem Mastery'%.ln12%)"), label .. " has no mastery life modifier")
    local mastery = referenced(skills, "Golem Mastery", label)
    local resist
    for _, candidate in pairs(skills) do
        if candidate.passivestat1 == "passive_summon_resist" and candidate.passivecalc1 == "dm12" then
            assert(not resist, label .. " has ambiguous summon resistance")
            resist = candidate
        end
    end
    resist = assert(resist, label .. " has no summon resistance modifier")
    local shift = integer(row, "manashift", label)
    return {
        behavior = "summon.golem",
        skill_id = integer(row, "Id", label),
        mana_cost_raw = integer(row, "mana", label) * (2 ^ shift),
        mana_cost_per_level_raw = integer(row, "lvlmana", label, -1000) * (2 ^ shift),
        minimum_mana_cost_raw = integer(row, "minmana", label) * (2 ^ shift),
        effect_delay = 1,
        complete_delay = 2,
        requires_point_target = row.TargetItem ~= "1",
        requires_item_target = row.TargetItem == "1",
        success_only = row.stsuccessonly == "1",
        summon_monster_id = row.summon,
        summon_mode = row.summode,
        category = row.pettype,
        category_group = tonumber(pet.group) or 0,
        category_base_max = math.max(tonumber(pet.basemax) or 0, 0),
        unsummon = pet.unsummon == "1",
        warp_with_owner = pet.warp == "1",
        range_limited = pet.range == "1",
        limit_policy = "fixed_one",
        mastery_skill_id = integer(mastery, "Id", label .. " mastery"),
        mastery_life_base = integer(mastery, "Param1", label .. " mastery"),
        mastery_life_per_level = integer(mastery, "Param2", label .. " mastery"),
        mastery_velocity_minimum = integer(mastery, "Param3", label .. " mastery"),
        mastery_velocity_maximum = integer(mastery, "Param4", label .. " mastery"),
        mastery_attack_base = integer(mastery, "Param5", label .. " mastery"),
        mastery_attack_per_level = integer(mastery, "Param6", label .. " mastery"),
        summon_resist_skill_id = integer(resist, "Id", label .. " resistance"),
        summon_resist_minimum = integer(resist, "Param1", label .. " resistance"),
        summon_resist_maximum = integer(resist, "Param2", label .. " resistance"),
        clay_skill_id = skill_id(skills, "Clay Golem", label),
        blood_skill_id = skill_id(skills, "BloodGolem", label),
        iron_skill_id = skill_id(skills, "IronGolem", label),
        fire_skill_id = skill_id(skills, "FireGolem", label),
        clay_attack_per_hard_level = integer(referenced(skills, "Clay Golem", label), "Param8", label),
        blood_life_per_hard_level = integer(referenced(skills, "BloodGolem", label), "Param8", label),
        iron_defense_per_hard_level = integer(referenced(skills, "IronGolem", label), "Param8", label),
        defense_level_source = "hard_modifier",
        fire_damage_per_hard_level = integer(referenced(skills, "FireGolem", label), "Param8", label),
        life_percent_per_level = 0,
        damage_percent_per_level = 0,
        slow_minimum = 0,
        slow_maximum = 0,
        slow_duration_ticks = 750,
        life_steal_minimum = 0,
        life_steal_maximum = 0,
        owner_share_percent = 0,
        owner_healing_share_percent = 0,
        thorns_base = 0,
        thorns_per_level = 0,
        fire_absorb_minimum = 0,
        fire_absorb_maximum = 0,
        fire_minimum = 0,
        fire_maximum = 0,
        fire_minimum_bands = {},
        fire_maximum_bands = {},
        granted_skill_name = "",
        granted_skill_base = 0,
        granted_skill_per_level = 0,
        granted_skill_cap = 0,
        granted_aura_radius_base = 0,
        granted_aura_radius_per_level = 0,
        granted_aura_period_ticks = 0,
        granted_aura_minimum = 0,
        granted_aura_maximum = 0,
        granted_aura_minimum_bands = {},
        granted_aura_maximum_bands = {},
        item_allowed_types = { "weap", "armo" },
        item_excluded_types = { "bow", "xbow" },
        item_required_material_flag = 2,
    }
end

local function decode(row, pet, skills, descriptions)
    local label = row.skill or ("skill " .. tostring(row.Id))
    local definition = common(row, pet, skills, descriptions, label)
    if row.srvdofunc == "57" then
        assert(
            row.srvstfunc == "20" and row.TargetItem == "1" and row.stsuccessonly == "1",
            label .. " has unsupported item-target transaction"
        )
    else
        assert(
            (row.srvstfunc or "") == "" and (row.TargetItem or "") == "",
            label .. " has unsupported point-target transaction"
        )
    end

    if row.aurastat1 == "item_slow" then
        assert(
            row.aurastatcalc1 == "dm34" and row.auraevent1 == "damagedinmelee" and row.auraeventfunc1 == "27",
            label .. " has unsupported slow reaction"
        )
        definition.life_percent_per_level = integer(row, "Param1", label)
        definition.damage_percent_per_level = integer(row, "Param2", label)
        definition.slow_minimum = integer(row, "Param3", label)
        definition.slow_maximum = integer(row, "Param4", label)
        assert_localized_synergies(
            descriptions[string.lower(row.skilldesc)],
            { "BloodGolem", "IronGolem", "FireGolem" },
            label
        )
    elseif row.auraeventfunc1 == "23" then
        assert(
            row.auraevent1 == "domeleedamage" and row.auraeventfunc2 == "26" and row.auraeventfunc3 == "26",
            label .. " has unsupported life-sharing reactions"
        )
        definition.damage_percent_per_level = integer(row, "Param4", label)
        definition.life_steal_minimum = integer(row, "Param1", label)
        definition.life_steal_maximum = integer(row, "Param2", label)
        definition.owner_share_percent = integer(row, "Param3", label)
        definition.owner_healing_share_percent = integer(row, "Param6", label)
        assert_localized_synergies(
            descriptions[string.lower(row.skilldesc)],
            { "Clay Golem", "IronGolem", "FireGolem" },
            label
        )
    elseif row.aurastat1 == "thorns_percent" then
        assert(
            row.aurastatcalc1 == "ln12" and row.aurastat2 == "fade" and row.aurastatcalc2 == "16",
            label .. " has unsupported thorns aura"
        )
        definition.thorns_base = integer(row, "Param1", label)
        definition.thorns_per_level = integer(row, "Param2", label)
        definition.defense_level_source = "effective_self"
        assert_localized_synergies(
            descriptions[string.lower(row.skilldesc)],
            { "Clay Golem", "BloodGolem", "FireGolem" },
            label
        )
    elseif row.aurastat1 == "fireresist" then
        assert(
            row.aurastatcalc1 == "100 - dm12"
                and row.aurastat2 == "item_absorbfire_percent"
                and row.aurastatcalc2 == "dm12"
                and row.aurastat3 == "firemindam"
                and row.aurastatcalc3 == "edmn"
                and row.aurastat4 == "firemaxdam"
                and row.aurastatcalc4 == "edmx",
            label .. " has unsupported fire effects"
        )
        definition.fire_absorb_minimum = integer(row, "Param1", label)
        definition.fire_absorb_maximum = integer(row, "Param2", label)
        definition.fire_minimum = integer(row, "EMin", label) * 256
        definition.fire_maximum = integer(row, "EMax", label) * 256
        for band = 1, 5 do
            definition.fire_minimum_bands[band] = integer(row, "EMinLev" .. band, label) * 256
            definition.fire_maximum_bands[band] = integer(row, "EMaxLev" .. band, label) * 256
        end
        assert(
            row.sumskill1 == "holy fire" and row.sumsk1calc == "min(ln56,30)",
            label .. " has unsupported granted aura"
        )
        definition.granted_skill_name = row.sumskill1
        definition.granted_skill_base = integer(row, "Param5", label)
        definition.granted_skill_per_level = integer(row, "Param6", label)
        definition.granted_skill_cap = 30
        local granted = referenced(skills, row.sumskill1, label)
        assert(
            granted.srvdofunc == "66"
                and granted.aura == "1"
                and granted.periodic == ""
                and granted.aurafilter == "42883"
                and granted.EType == "fire",
            label
                .. " has unsupported granted periodic aura: srv="
                .. tostring(granted.srvdofunc)
                .. " aura="
                .. tostring(granted.aura)
                .. " periodic="
                .. tostring(granted.periodic)
                .. " filter="
                .. tostring(granted.aurafilter)
                .. " etype="
                .. tostring(granted.EType)
        )
        definition.granted_aura_radius_base = integer(granted, "Param1", label .. " granted aura")
        definition.granted_aura_radius_per_level = integer(granted, "Param2", label .. " granted aura")
        definition.granted_aura_period_ticks = integer(granted, "perdelay", label .. " granted aura")
        definition.granted_aura_minimum = integer(granted, "EMin", label .. " granted aura") * 256
        definition.granted_aura_maximum = integer(granted, "EMax", label .. " granted aura") * 256
        for band = 1, 5 do
            definition.granted_aura_minimum_bands[band] = integer(granted, "EMinLev" .. band, label .. " granted aura")
                * 256
            definition.granted_aura_maximum_bands[band] = integer(granted, "EMaxLev" .. band, label .. " granted aura")
                * 256
        end
        assert_localized_synergies(
            descriptions[string.lower(row.skilldesc)],
            { "Clay Golem", "BloodGolem", "IronGolem" },
            label
        )
    else
        error(label .. " has unsupported golem effect shape")
    end
    return definition
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported golem skill IDs are required")
    local rows = assert(records.load("data/global/excel/skills.txt"))
    local by_id, by_name = index(rows, "Id"), index(rows, "skill")
    if #supported_ids == 0 then
        return {}
    end
    local pets = index(assert(records.load("data/global/excel/PetType.txt")), "pet type")
    local descriptions = index(assert(records.load("data/global/excel/SkillDesc.txt")), "skilldesc")
    local result = {}
    for _, id in ipairs(supported_ids) do
        local row = by_id[tostring(id)]
        if row then
            local pet = assert(pets[string.lower(row.pettype or "")], row.skill .. " references unknown PetType")
            local definition = decode(row, pet, by_name, descriptions)
            assert(not result[definition.skill_id], "duplicate golem skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
