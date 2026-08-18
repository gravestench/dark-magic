-- Decode the complete player-usable Assassin trap family from pinned 1.14d
-- Expansion records. Exact IDs admit coverage; record shape chooses one of the
-- reusable device, projectile, channel, or periodic-state transactions.

local records = require("engine.records/v1")
local M = {}

local RAW = 256

local function index(rows, column)
    local result = {}
    for _, row in ipairs(rows) do
        local key = row[column]
        if key and key ~= "" and result[string.lower(key)] == nil then
            result[string.lower(key)] = row
        end
    end
    return result
end

local function lookup(rows, value, label)
    return assert(rows[string.lower(value or "")], label .. " references missing " .. tostring(value))
end

local function integer(row, column, label, minimum)
    local value = tonumber(row[column])
    assert(value and value == math.floor(value) and value >= (minimum or 0), label .. " has invalid " .. column)
    return value
end

local function optional_integer(row, column, fallback)
    local value = tonumber(row[column])
    return value and math.floor(value) or fallback
end

local function matched_integer(expression, pattern, fallback)
    local matched = string.match(expression or "", pattern)
    return matched and tonumber(matched) or fallback
end

local function truth(row, column)
    return row[column] == "1"
end

local function damage_bands(row, prefix, label)
    local result = {}
    local scale = 2 ^ integer(row, "HitShift", label)
    for band = 1, 5 do
        result[band] = integer(row, prefix .. band, label) * scale
    end
    return result
end

local function mana(row, label)
    local scale = 2 ^ integer(row, "manashift", label)
    return integer(row, "mana", label) * scale,
        optional_integer(row, "lvlmana", 0) * scale,
        optional_integer(row, "minmana", 0) * scale
end

local function skill_id(skills, name, label)
    return integer(lookup(skills, name, label), "Id", label .. " modifier")
end

local function referenced_hard_levels(expression, skills, label)
    local result = {}
    local seen = {}
    for name in string.gmatch(expression or "", "skill%('([^']+)'%.blvl%)") do
        local id = skill_id(skills, name, label)
        if not seen[id] then
            seen[id] = true
            result[#result + 1] = id
        end
    end
    return result
end

local function damage_synergy_terms(row, skills, label)
    local expression = row.EDmgSymPerCalc or ""
    if expression == "" then
        return {}
    end
    local ids = referenced_hard_levels(expression, skills, label)
    local terms = {}
    for _, id in ipairs(ids) do
        local name
        for candidate in string.gmatch(expression, "skill%('([^']+)'%.blvl%)") do
            if skill_id(skills, candidate, label) == id then
                name = candidate
                break
            end
        end
        local escaped = string.gsub(name, "([%%%-%+%*%?%[%]%^%$%(%)%.])", "%%%1")
        local tail = string.match(expression, "skill%('" .. escaped .. "'%.blvl%)(.*)")
        local parameter = tail and string.match(tail, "par(%d+)")
        if not parameter then
            local head = string.match(expression, "^(.-)skill%('" .. escaped .. "'%.blvl%)")
            parameter = head and string.match(head, "par(%d+)[^p]*$")
        end
        assert(parameter, label .. " has unsupported synergy term for " .. name)
        terms[#terms + 1] = { skill_id = id, percent = integer(row, "Param" .. parameter, label) }
    end
    return terms
end

local function assert_description(row, descriptions, label)
    local description = lookup(descriptions, row.skilldesc, label)
    assert(description["str name"] and description["str name"] ~= "", label .. " has no localized name key")
    if (row.EDmgSymPerCalc or "") ~= "" then
        assert(description.dsc3texta1 == "Sksyn", label .. " has unsupported localized synergy heading")
        local count = 0
        for column, value in pairs(description) do
            if string.match(column, "^dsc3texta%d+$") and column ~= "dsc3texta1" and value ~= "" then
                count = count + 1
            end
        end
        assert(count > 0, label .. " has no localized synergy entries")
    end
    return description
end

local function missile_definition(row, label)
    assert(
        row.pSrvDoFunc == "1" or row.pSrvDoFunc == "20" or row.pSrvDoFunc == "31",
        label .. " missile has unsupported server movement"
    )
    local cel = assert(row.CelFile and row.CelFile ~= "" and row.CelFile, label .. " missile has no CelFile")
    return {
        missile_id = row.Missile,
        speed_per_tick = optional_integer(row, "Vel", 0) / 25,
        lifetime_ticks = integer(row, "Range", label),
        collision_radius = optional_integer(row, "Size", 1) / 2,
        destroy_on_contact = truth(row, "CollideKill"),
        next_hit_delay = optional_integer(row, "NextDelay", 0),
        dcc = cel == "null" and "" or "data/global/missiles/" .. cel .. ".dcc",
        palette = "data/global/palette/units/pal.dat",
        travel_sound = row.TravelSound or "",
        hit_sound = row.HitSound or "",
        directions = math.max(optional_integer(row, "NumDirections", 1), 1),
        frames_per_second = math.max(math.floor(optional_integer(row, "AnimSpeed", 16) * 25 / 16), 1),
        loop = truth(row, "LoopAnim"),
        transparency_mode = optional_integer(row, "Trans", 0),
        offset_x = optional_integer(row, "Xoffset", 0),
        offset_y = optional_integer(row, "Yoffset", 0),
        offset_z = optional_integer(row, "Zoffset", 0),
    }
end

local function damage_definition(row, label)
    local scale = 2 ^ integer(row, "HitShift", label)
    local elemental = row.EMin and row.EMin ~= ""
    local minimum_column = elemental and "EMin" or "MinDam"
    local maximum_column = elemental and "EMax" or "MaxDam"
    local minimum_prefix = elemental and "EMinLev" or "MinLevDam"
    local maximum_prefix = elemental and "EMaxLev" or "MaxLevDam"
    local channel = elemental and row.EType or "physical"
    channel = channel == "ltng" and "lightning" or channel
    return {
        minimum_damage_raw = integer(row, minimum_column, label) * scale,
        maximum_damage_raw = integer(row, maximum_column, label) * scale,
        minimum_damage_per_level_raw = damage_bands(row, minimum_prefix, label),
        maximum_damage_per_level_raw = damage_bands(row, maximum_prefix, label),
        damage_channel = channel,
    }
end

local function merge(target, source)
    for key, value in pairs(source) do
        target[key] = value
    end
    return target
end

local function common(row, skills, descriptions, label)
    assert(row.InGame == "1" and row.charclass == "ass", label .. " is not a player Assassin skill")
    assert_description(row, descriptions, label)
    local base, per_level, minimum = mana(row, label)
    return {
        behavior = "trap.assassin-family",
        skill_id = integer(row, "Id", label),
        mana_cost_raw = base,
        mana_cost_per_level_raw = per_level,
        minimum_mana_cost_raw = minimum,
        damage_synergy_terms = damage_synergy_terms(row, skills, label),
        effect_delay = 1,
        complete_delay = 2,
        requires_point_target = row.range == "rng",
    }
end

local function decode_lobbed(row, missiles, skills, descriptions, label)
    assert(row.lob == "1" and row.srvmissile ~= "", label .. " has unsupported lob policy")
    local definition = common(row, skills, descriptions, label)
    definition.shape = "lobbed_payload"
    merge(definition, damage_definition(row, label))
    local air_row = lookup(missiles, row.srvmissile, label)
    merge(definition, missile_definition(air_row, label))
    local ground = lookup(missiles, assert(air_row.HitSubMissile1, label .. " has no payload"), label)
    local impact = lookup(missiles, assert(ground.ExplosionMissile, label .. " has no explosion"), label)
    definition.impact_radius = integer(row, "Param1", label)
    definition.impact_missile_id = impact.Missile
    definition.impact_dcc = "data/global/missiles/" .. impact.CelFile .. ".dcc"
    definition.impact_palette = "data/global/palette/units/pal.dat"
    definition.impact_lifetime_ticks = optional_integer(impact, "Range", 1)
    definition.impact_directions = math.max(optional_integer(impact, "NumDirections", 1), 1)
    definition.impact_frames_per_second = math.max(math.floor(optional_integer(impact, "AnimSpeed", 16) * 25 / 16), 1)
    definition.impact_loop = truth(impact, "LoopAnim")
    definition.impact_transparency_mode = optional_integer(impact, "Trans", 0)
    definition.impact_sound = impact.TravelSound or ""
    definition.destroy_on_contact = true
    definition.impact_on_expiry = true
    return definition
end

local function decode_field(row, missiles, skills, descriptions, label)
    assert(row.srvdofunc == "43" and row.srvmissilea ~= "", label .. " has unsupported field deployment")
    local definition = common(row, skills, descriptions, label)
    definition.shape = "persistent_field"
    merge(definition, damage_definition(row, label))
    local air = lookup(missiles, row.srvmissilea, label)
    local ground = lookup(missiles, assert(air.HitSubMissile1, label .. " has no field payload"), label)
    merge(definition, missile_definition(ground, label))
    definition.speed_per_tick = 0
    definition.destroy_on_contact = false
    definition.field_count_base = integer(row, "Param1", label)
    definition.field_count_levels_per_add = integer(row, "Param2", label)
    local synergy = assert(string.match(row.prgcalc1, "skill%('([^']+)'%.blvl%)"), label .. " has no count synergy")
    definition.field_count_synergy_skill_id = skill_id(skills, synergy, label)
    definition.field_count_synergy_divisor = matched_integer(row.prgcalc1, "blvl%)/(%d+)", 1)
    definition.deploy_delay = integer(row, "delay", label)
    return definition
end

local function decode_patrol(row, missiles, pets, skills, descriptions, label)
    assert(row.srvdofunc == "44" and row.summon ~= "" and row.pettype ~= "", label .. " has unsupported patrol")
    local definition = common(row, skills, descriptions, label)
    definition.shape = "returning_weapon_patrol"
    merge(definition, damage_definition(row, label))
    merge(definition, missile_definition(lookup(missiles, row.srvmissilea, label), label))
    definition.weapon_fraction = integer(row, "SrcDam", label)
    definition.duration_base = integer(row, "Param1", label)
    definition.duration_per_level = integer(row, "Param2", label)
    definition.category = row.pettype
    definition.category_group = optional_integer(lookup(pets, row.pettype, label), "group", 0)
    definition.category_base_max = integer(row, "petmax", label)
    definition.deploy_delay = integer(row, "delay", label)
    return definition
end

local function helper_projectile(helper, missiles, label)
    local name = helper.srvmissile ~= "" and helper.srvmissile or helper.srvmissilea
    assert(name and name ~= "", label .. " helper has no server missile")
    return missile_definition(lookup(missiles, name, label), label)
end

local function shot_synergy(row, helper, skills, label)
    local expression = row.calc4 ~= "" and row.calc4 or helper.calc4 or ""
    local name = string.match(expression, "skill%('([^']+)'%.blvl%)")
    if not name then
        return 0, 1
    end
    return skill_id(skills, name, label), matched_integer(expression, "blvl%)/(%d+)", 1)
end

local function decode_sentry(row, missiles, pets, skills, monsters, descriptions, label)
    assert(row.srvdofunc == "45" and row.summon ~= "" and row.sumskill1 ~= "", label .. " has unsupported sentry")
    local definition = common(row, skills, descriptions, label)
    definition.shape = "stationary_sentry"
    merge(definition, damage_definition(row, label))
    local helper = lookup(skills, row.sumskill1, label)
    definition.helper_skill_id = integer(helper, "Id", label .. " helper")
    definition.operation = truth(helper, "TargetCorpse") and "corpse_or_projectile" or "projectile"
    local projectile_helper = helper
    if definition.operation == "corpse_or_projectile" then
        projectile_helper = lookup(skills, row.sumskill2, label)
        definition.corpse_minimum_percent = integer(helper, "Param1", label .. " corpse helper")
        definition.corpse_maximum_percent = integer(helper, "Param2", label .. " corpse helper")
        definition.corpse_fire_percent = integer(helper, "calc3", label .. " corpse helper")
        definition.corpse_radius_base = integer(helper, "Param3", label .. " corpse helper") / 2
        definition.corpse_radius_per_level = integer(helper, "Param4", label .. " corpse helper") / 2
    end
    merge(definition, helper_projectile(projectile_helper, missiles, label))
    definition.projectile_helper_skill_id = integer(projectile_helper, "Id", label .. " projectile helper")
    definition.monster_id = row.summon
    local monster = lookup(monsters, row.summon, label)
    assert(monster.AI == "AssassinSentry" or monster.AI == "DeathSentry", label .. " has unsupported sentry AI")
    definition.fire_interval = math.max(integer(monster, "aip4", label), 1)
    definition.target_radius = math.max(integer(monster, "aip3", label), 1)
    definition.category = row.pettype
    definition.category_group = optional_integer(lookup(pets, row.pettype, label), "group", 0)
    definition.category_base_max = integer(row, "petmax", label)
    definition.range_limited = truth(lookup(pets, row.pettype, label), "range")
    definition.shots_base = integer(row, "Param1", label)
    definition.shot_synergy_skill_id, definition.shot_synergy_divisor = shot_synergy(row, helper, skills, label)
    definition.burst_count_base = optional_integer(helper, "Param1", 1)
    definition.burst_count_per_level = optional_integer(helper, "Param2", 0)
    local burst_name = string.match(helper.calc1 or "", "skill%('([^']+)'%.blvl%)")
    definition.burst_synergy_skill_id = burst_name and skill_id(skills, burst_name, label) or 0
    definition.burst_synergy_divisor = matched_integer(helper.calc1, "blvl%)/(%d+)", 1)
    return definition
end

local function decode_channel(row, missiles, skills, descriptions, label)
    assert(
        row.srvstfunc == "26" and row.srvdofunc == "48" and row["repeat"] == "1" and row.usemanaondo == "1",
        label .. " has unsupported repeat channel"
    )
    local definition = common(row, skills, descriptions, label)
    definition.shape = "repeat_weapon_missile"
    merge(definition, damage_definition(row, label))
    merge(definition, missile_definition(lookup(missiles, row.srvmissilea, label), label))
    definition.weapon_fraction = integer(row, "SrcDam", label)
    definition.minimum_start_mana_raw = integer(row, "Param3", label) * RAW
    definition.channel_period_ticks = integer(row, "Param4", label)
    definition.complete_delay = definition.channel_period_ticks + 1
    return definition
end

local function decode_periodic(row, missiles, skills, descriptions, label)
    assert(
        row.srvstfunc == "28" and row.srvdofunc == "54" and row.periodic == "1",
        label .. " has unsupported periodic state"
    )
    local definition = common(row, skills, descriptions, label)
    definition.shape = "periodic_weapon_state"
    merge(definition, damage_definition(row, label))
    definition.state_id = assert(row.aurastate, label .. " has no state")
    definition.duration_base = integer(row, "Param1", label)
    definition.duration_per_level = integer(row, "Param2", label)
    definition.period_ticks = integer(row, "Param3", label)
    definition.radius = integer(row, "Param4", label)
    definition.weapon_fraction = integer(row, "SrcDam", label)
    definition.attachment_missile_id = row.srvmissilea
    local attachment = lookup(missiles, row.srvmissilea, label)
    definition.attachment_submissile_id = attachment.CltSubMissile1 or ""
    definition.requires_point_target = false
    return definition
end

local function decode(row, indexed)
    local label = row.skill or ("skill " .. tostring(row.Id))
    if row.lob == "1" then
        return decode_lobbed(row, indexed.missiles, indexed.skills, indexed.descriptions, label)
    elseif row.srvdofunc == "43" then
        return decode_field(row, indexed.missiles, indexed.skills, indexed.descriptions, label)
    elseif row.srvdofunc == "44" then
        return decode_patrol(row, indexed.missiles, indexed.pets, indexed.skills, indexed.descriptions, label)
    elseif row.srvdofunc == "45" then
        return decode_sentry(
            row,
            indexed.missiles,
            indexed.pets,
            indexed.skills,
            indexed.monsters,
            indexed.descriptions,
            label
        )
    elseif row.srvstfunc == "26" then
        return decode_channel(row, indexed.missiles, indexed.skills, indexed.descriptions, label)
    elseif row.srvstfunc == "28" then
        return decode_periodic(row, indexed.missiles, indexed.skills, indexed.descriptions, label)
    end
    error(label .. " has unsupported trap-family record shape")
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported Assassin trap skill IDs are required")
    local skill_rows = assert(records.load("data/global/excel/skills.txt"))
    local indexed = {
        skills = index(skill_rows, "skill"),
        skill_ids = index(skill_rows, "Id"),
        missiles = index(assert(records.load("data/global/excel/Missiles.txt")), "Missile"),
        pets = index(assert(records.load("data/global/excel/PetType.txt")), "pet type"),
        monsters = index(assert(records.load("data/global/excel/MonStats.txt")), "Id"),
        monster_graphics = index(assert(records.load("data/global/excel/MonStats2.txt")), "Id"),
        descriptions = index(assert(records.load("data/global/excel/SkillDesc.txt")), "skilldesc"),
    }
    local result = {}
    for _, id in ipairs(supported_ids) do
        local row = indexed.skill_ids[tostring(id)]
        if row then
            local definition = decode(row, indexed)
            if definition.monster_id then
                local monster = lookup(indexed.monsters, definition.monster_id, row.skill)
                lookup(indexed.monster_graphics, monster.MonStatsEx, row.skill)
            end
            assert(not result[definition.skill_id], "duplicate Assassin trap skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
