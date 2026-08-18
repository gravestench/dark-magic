-- Decode explicitly supported straight-missile skills from immutable records.
--
-- The supported ID list is implementation coverage, not a behavior switch.
-- Every selected row must satisfy this family's reviewed Skills/Missiles
-- contract. Unknown function/flag combinations fail instead of being guessed.

local records = require("engine.records/v1")
local skill_modifiers = require("d2legacy.data.skill_modifiers")

local M = {}

local function field(row, canonical, legacy)
    local value = row[canonical]
    if value == nil and legacy then
        value = row[legacy]
    end
    return value
end

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

local function required_integer(row, column, legacy, label)
    local value = tonumber(field(row, column, legacy))
    assert(value and value >= 0 and value == math.floor(value), label .. " has invalid " .. column)
    return value
end

local function shifted(row, value_column, shift_column, legacy_value, label)
    return required_integer(row, value_column, legacy_value, label)
        * (2 ^ required_integer(row, shift_column, nil, label))
end

local function shifted_or(row, value_column, shift_column, fallback, label)
    local value = tonumber(row[value_column])
    if value == nil then
        return fallback
    end
    assert(value >= 0 and value == math.floor(value), label .. " has invalid " .. value_column)
    return value * (2 ^ required_integer(row, shift_column, nil, label))
end

local function damage_gains(row, prefix, label)
    local result = {}
    for tier = 1, 5 do
        result[tier] = shifted_or(row, prefix .. tier, "HitShift", 0, label)
    end
    return result
end

local function integer_or(row, column, fallback)
    local value = tonumber(row[column])
    if value == nil then
        return fallback
    end
    return math.floor(value)
end

local function byte_or(row, column, fallback, label)
    local authored = row[column]
    if authored == nil or authored == "" then
        return fallback
    end
    local value = tonumber(authored)
    assert(value and value >= 0 and value <= 255 and value == math.floor(value), label .. " has invalid " .. column)
    return value
end

local function decode(skill, missile, missiles, skills_by_name, behavior)
    local skill_id = assert(tonumber(skill.Id), "straight-missile skill has no numeric ID")
    local label = skill.skill or ("skill " .. skill_id)
    local missile_id =
        assert(skill.srvmissile and skill.srvmissile ~= "" and skill.srvmissile, label .. " has no server missile")
    local damage_channel = field(skill, "EType", "etype")
    assert(
        type(damage_channel) == "string" and damage_channel ~= "" and skill.interrupt == "1",
        label .. " has an unsupported element or interrupt policy"
    )
    assert(skill.srvstfunc == "" and skill.srvdofunc == "", label .. " has unsupported server functions")
    assert(missile and missile.Missile == missile_id, label .. " missile is missing")
    assert(missile.Skill == label and missile.pSrvDoFunc == "1", label .. " has an unsupported missile function")
    assert(missile.CollideType == "3" and missile.CollideKill == "1", label .. " has an unsupported collision policy")

    local velocity = required_integer(missile, "Vel", nil, label)
    local lifetime = required_integer(missile, "Range", nil, label)
    local animation_speed = integer_or(missile, "AnimSpeed", 16)
    local cel =
        assert(missile.CelFile and missile.CelFile ~= "" and missile.CelFile, label .. " missile has no CelFile")
    local synergy_ids, synergy_percent =
        skill_modifiers.hard_level_sum_percent(skill, "EDmgSymPerCalc", "Param8", skills_by_name, label)
    local impact_radius = 0
    local impact_radius_per_level = 0
    local impact = nil
    local duration_base = 0
    local duration_per_level = 0
    local duration_synergy_ids = {}
    local duration_synergy_percent = 0
    if behavior == "missile.straight-impact-area" then
        assert(missile.pSrvHitFunc == "1", label .. " has unsupported impact function")
        impact_radius = required_integer(missile, "sHitPar1", nil, label)
        assert(impact_radius > 0, label .. " has no impact radius")
    elseif behavior == "missile.straight-freeze" then
        assert(missile.pSrvDmgFunc == "4", label .. " has unsupported damage function")
        duration_base = required_integer(skill, "ELen", nil, label)
        duration_per_level = required_integer(skill, "ELevLen1", nil, label)
        assert(
            duration_base > 0
                and duration_per_level > 0
                and required_integer(skill, "ELevLen2", nil, label) == duration_per_level
                and required_integer(skill, "ELevLen3", nil, label) == duration_per_level,
            label .. " has unsupported freeze-length bands"
        )
        duration_synergy_ids, duration_synergy_percent =
            skill_modifiers.hard_level_sum_percent(skill, "ELenSymPerCalc", "Param7", skills_by_name, label)
    elseif behavior == "missile.straight-impact-area-freeze" then
        assert(
            missile.pSrvHitFunc == "13" and missile.EType == "frze" and missile.HitFlags == "2",
            label .. " has unsupported area-freeze functions"
        )
        assert(skill.aurarangecalc == "ln12", label .. " has unsupported area-radius formula")
        impact_radius = required_integer(skill, "Param1", nil, label)
        impact_radius_per_level = required_integer(skill, "Param2", nil, label)
        assert(impact_radius > 0, label .. " has no impact radius")
        duration_base = required_integer(skill, "Param3", nil, label)
        duration_per_level = required_integer(skill, "Param4", nil, label)
        assert(duration_base > 0 and duration_per_level > 0, label .. " has no freeze duration")
        duration_synergy_ids, duration_synergy_percent = skill_modifiers.single_hard_level_percent_multiplier(
            skill,
            "auralencalc",
            "ln34",
            "Param7",
            skills_by_name,
            label
        )
    end
    if
        behavior == "missile.straight-impact-area"
        or behavior == "missile.straight-freeze"
        or behavior == "missile.straight-impact-area-freeze"
    then
        local impact_id = behavior == "missile.straight-impact-area-freeze" and missile.CltHitSubMissile1
            or missile.ExplosionMissile
        assert(impact_id and impact_id ~= "", label .. " has no impact missile")
        impact = assert(missiles[impact_id], label .. " impact missile is missing")
        assert(impact.Explosion == "1", label .. " impact missile is not an explosion")
    end
    return {
        behavior = behavior,
        trajectory = "straight",
        skill_id = math.floor(skill_id),
        mana_cost_raw = shifted(skill, "mana", "manashift", nil, label),
        mana_cost_per_level_raw = shifted_or(skill, "lvlmana", "manashift", 0, label),
        minimum_mana_cost_raw = shifted_or(skill, "minmana", "manashift", 0, label),
        minimum_damage_raw = shifted(skill, "EMin", "HitShift", "emin", label),
        maximum_damage_raw = shifted(skill, "EMax", "HitShift", "emax", label),
        minimum_damage_per_level_raw = damage_gains(skill, "EMinLev", label),
        maximum_damage_per_level_raw = damage_gains(skill, "EMaxLev", label),
        damage_synergy_skill_ids = synergy_ids,
        damage_synergy_percent_per_level = synergy_percent,
        effect_delay = 1,
        complete_delay = 2,
        requires_point_target = true,
        speed_per_tick = velocity / 25,
        lifetime_ticks = lifetime,
        maximum_range = velocity * lifetime / 25,
        collision_radius = required_integer(missile, "Size", nil, label) / 2,
        destroy_on_contact = true,
        next_hit_delay = 0,
        impact_radius = impact_radius,
        impact_radius_per_level = impact_radius_per_level,
        impact_missile_id = impact and impact.Missile or "",
        impact_dcc = impact and "data/global/missiles/" .. impact.CelFile .. ".dcc" or "",
        impact_palette = impact and "data/global/palette/units/pal.dat" or "",
        impact_lifetime_ticks = impact and required_integer(impact, "Range", nil, label) or 0,
        impact_directions = impact and math.max(integer_or(impact, "NumDirections", 1), 1) or 1,
        impact_frames_per_second = impact and math.max(math.floor(integer_or(impact, "AnimSpeed", 16) * 25 / 16), 1)
            or 1,
        impact_loop = impact and impact.LoopAnim == "1" or false,
        impact_transparency_mode = impact and integer_or(impact, "Trans", 0) or 0,
        impact_sound = impact and impact.TravelSound or "",
        effect_duration_base = duration_base,
        effect_duration_per_level = duration_per_level,
        effect_duration_synergy_skill_ids = duration_synergy_ids,
        effect_duration_synergy_percent_per_level = duration_synergy_percent,
        on_hit_state_id = (behavior == "missile.straight-freeze" or behavior == "missile.straight-impact-area-freeze")
                and "freeze"
            or "",
        on_hit_state_duration_policy = (
            behavior == "missile.straight-freeze" or behavior == "missile.straight-impact-area-freeze"
        )
                and "monster_cold"
            or "",
        on_hit_state_action_disabled = behavior == "missile.straight-freeze"
            or behavior == "missile.straight-impact-area-freeze",
        on_hit_state_exclusive_group = "",
        -- Preserve the target-authored byte without yet assigning binary-owned
        -- chance/result semantics to it.
        knockback_value = byte_or(missile, "KnockBack", 0, label),
        damage_channel = damage_channel,
        missile_id = missile_id,
        dcc = "data/global/missiles/" .. cel .. ".dcc",
        palette = "data/global/palette/units/pal.dat",
        travel_sound = missile.TravelSound or "",
        hit_sound = missile.HitSound or "",
        directions = math.max(integer_or(missile, "NumDirections", 1), 1),
        frames_per_second = math.max(math.floor(animation_speed * 25 / 16), 1),
        loop = missile.LoopAnim == "1",
        transparency_mode = integer_or(missile, "Trans", 0),
        offset_x = integer_or(missile, "Xoffset", 0),
        offset_y = integer_or(missile, "Yoffset", 0),
        offset_z = integer_or(missile, "Zoffset", 0),
    }
end

function M.load(supported_ids, behavior)
    assert(type(supported_ids) == "table", "supported straight-missile skill IDs are required")
    behavior = behavior or "missile.straight"
    assert(
        behavior == "missile.straight"
            or behavior == "missile.straight-impact-area"
            or behavior == "missile.straight-freeze"
            or behavior == "missile.straight-impact-area-freeze",
        "unsupported straight-missile behavior"
    )
    local skill_rows = assert(records.load("data/global/excel/skills.txt"))
    local skills = index(skill_rows, "Id")
    local skills_by_name = skill_modifiers.by_name(skill_rows)
    local missiles = index(assert(records.load("data/global/excel/Missiles.txt")), "Missile")
    local definitions = {}
    for _, skill_id in ipairs(supported_ids) do
        -- Narrow suite fixtures need not reproduce every admitted retail row.
        -- Mounted-data coverage and owned-archive boot remain the strict target
        -- completeness gates; any row carrying this family's missile contract
        -- is still decoded by exact ID. Name-only synergy fixtures are ignored.
        local skill = skills[tostring(skill_id)]
        if skill and skill.srvmissile and skill.srvmissile ~= "" then
            local definition = decode(skill, missiles[skill.srvmissile], missiles, skills_by_name, behavior)
            assert(not definitions[definition.skill_id], "duplicate straight-missile skill ID")
            definitions[definition.skill_id] = definition
        end
    end
    return definitions
end

return M
