-- Decode explicitly supported targetless, timed self-state skills.

local records = require("engine.records/v1")
local skill_modifiers = require("d2legacy.data.skill_modifiers")
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

local function required_integer(row, column, label)
    local value = tonumber(row[column])
    assert(value and value >= 0 and value == math.floor(value), label .. " has invalid " .. column)
    return value
end

local function shifted(row, value_column, shift_column, label)
    return required_integer(row, value_column, label) * (2 ^ required_integer(row, shift_column, label))
end

local function integer_or(row, column, fallback)
    local value = tonumber(row[column])
    if value == nil then
        return fallback
    end
    return math.floor(value)
end

local function shifted_or(row, column, shift_column, fallback, label)
    local value = tonumber(row[column])
    if value == nil then
        return fallback
    end
    assert(value >= 0 and value == math.floor(value), label .. " has invalid " .. column)
    return value * (2 ^ required_integer(row, shift_column, label))
end

local function damage_gains(row, prefix, label)
    local result = {}
    for tier = 1, 5 do
        result[tier] = shifted_or(row, prefix .. tier, "HitShift", 0, label)
    end
    return result
end

local function synergy_ids(expression, skills_by_name, marker, label)
    local result = {}
    for name in string.gmatch(expression or "", "skill%('([^']+)'%.blvl%)") do
        local skill = assert(skills_by_name[string.lower(name)], label .. " has an unknown synergy " .. name)
        result[#result + 1] = required_integer(skill, "Id", label)
    end
    assert(#result > 0 and string.find(expression, marker, 1, true), label .. " has an unsupported formula")
    return result
end

local function same_ids(left, right)
    if #left ~= #right then
        return false
    end
    for index, value in ipairs(left) do
        if value ~= right[index] then
            return false
        end
    end
    return true
end

local reactions = {
    ["damagedinmelee\0" .. "2"] = "melee_damage_freeze",
    ["attackedinmelee\0" .. "3"] = "melee_attack_cold",
    ["hitbymissile\0" .. "1"] = "missile_hit_return",
}

local function missile_definition(row, missiles, label)
    local missile_id = row.srvmissilea
    if not missile_id or missile_id == "" then
        return nil
    end
    local missile = assert(missiles[missile_id], label .. " has an unknown reaction missile")
    assert(
        missile.Skill == label
            and missile.pSrvDoFunc == "1"
            and missile.CollideType == "3"
            and missile.CollideKill == "1",
        label .. " has an unsupported reaction missile"
    )
    local velocity = required_integer(missile, "Vel", label)
    local lifetime = required_integer(missile, "Range", label)
    local cel =
        assert(missile.CelFile and missile.CelFile ~= "" and missile.CelFile, label .. " missile has no CelFile")
    local animation_speed = integer_or(missile, "AnimSpeed", 16)
    return {
        trajectory = "straight",
        missile_id = missile_id,
        speed_per_tick = velocity / 25,
        lifetime_ticks = lifetime,
        maximum_range = velocity * lifetime / 25,
        collision_radius = required_integer(missile, "Size", label) / 2,
        destroy_on_contact = true,
        next_hit_delay = 0,
        impact_radius = 0,
        impact_radius_per_level = 0,
        impact_missile_id = "",
        impact_dcc = "",
        impact_palette = "",
        impact_lifetime_ticks = 0,
        impact_directions = 1,
        impact_frames_per_second = 1,
        impact_loop = false,
        impact_transparency_mode = 0,
        impact_sound = "",
        knockback_value = 0,
        damage_channel = "cold",
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

local function decode(row, skills_by_name, states_by_name, missiles)
    local id = required_integer(row, "Id", "self-state skill")
    local label = row.skill or ("skill " .. id)
    assert(row.srvstfunc == "" and row.srvdofunc == "18", label .. " has unsupported server functions")
    assert(
        row.aurastate and row.aurastate ~= "" and row.aurastat1 == "skill_armor_percent",
        label .. " has an unsupported state or stat"
    )
    assert(row.aurastatcalc1 == "ln12", label .. " has an unsupported stat formula")
    assert(string.find(row.auralencalc or "", "*par7", 1, true), label .. " has an unsupported synergy formula")
    local reaction = assert(
        reactions[(row.auraevent1 or "") .. "\0" .. (row.auraeventfunc1 or "")],
        label .. " has an unsupported reactive event"
    )
    local duration_synergies = synergy_ids(row.auralencalc, skills_by_name, "ln34", label)
    local armor_state = assert(states_by_name[row.aurastate], label .. " has an unknown armor state")
    local state_group = required_integer(armor_state, "group", label)
    assert(state_group > 0, label .. " armor state has no exclusive group")
    local definition = {
        behavior = "state.self-timed-reactive",
        skill_id = id,
        mana_cost_raw = shifted(row, "mana", "manashift", label),
        effect_delay = 1,
        complete_delay = 2,
        state_id = row.aurastate,
        stat = "defense",
        stat_operation = "percent",
        stat_base = required_integer(row, "Param1", label),
        stat_per_level = required_integer(row, "Param2", label),
        duration_base = required_integer(row, "Param3", label),
        duration_per_level = required_integer(row, "Param4", label),
        duration_synergy_per_level = required_integer(row, "Param7", label),
        duration_synergy_skill_ids = duration_synergies,
        exclusive_group = "state-group:" .. state_group,
        reaction = reaction,
        reaction_state_id = "",
        reaction_chill_state_id = "",
        reaction_stat = "",
        reaction_stat_value = 0,
        reaction_chill_stat = "velocitypercent",
        reaction_chill_stat_value = -50,
        reaction_duration_base = 0,
        reaction_duration_per_level = { 0, 0, 0, 0, 0 },
        reaction_duration_synergy_percent = 0,
        reaction_duration_synergy_skill_ids = {},
        reaction_disables_action = false,
        minimum_damage_raw = 0,
        maximum_damage_raw = 0,
        minimum_damage_per_level_raw = { 0, 0, 0, 0, 0 },
        maximum_damage_per_level_raw = { 0, 0, 0, 0, 0 },
        damage_synergy_skill_ids = {},
        damage_synergy_percent_per_level = 0,
        damage_channel = "cold",
        reaction_missile = missile_definition(row, missiles, label),
        reaction_overlay = row.cltoverlaya or "",
    }
    if reaction == "melee_damage_freeze" then
        assert(string.find(row.calc1 or "", "*par8", 1, true), label .. " has an unsupported reaction formula")
        local reaction_synergies = synergy_ids(row.calc1, skills_by_name, "ln56", label)
        assert(same_ids(duration_synergies, reaction_synergies), label .. " has mismatched reactive synergies")
        definition.reaction_state_id = assert(states_by_name.freeze, label .. " requires the freeze state").state
        definition.reaction_chill_state_id = assert(states_by_name.cold, label .. " requires the cold state").state
        definition.reaction_duration_base = required_integer(row, "Param5", label)
        definition.reaction_duration_per_level = { required_integer(row, "Param6", label), 0, 0, 0, 0 }
        for tier = 2, 5 do
            definition.reaction_duration_per_level[tier] = definition.reaction_duration_per_level[1]
        end
        definition.reaction_duration_synergy_percent = required_integer(row, "Param8", label)
        definition.reaction_duration_synergy_skill_ids = reaction_synergies
        definition.reaction_disables_action = true
        assert(not definition.reaction_missile, label .. " freeze reaction unexpectedly has a missile")
    else
        assert(row.EType == "cold", label .. " has an unsupported reaction element")
        definition.minimum_damage_raw = shifted(row, "EMin", "HitShift", label)
        definition.maximum_damage_raw = shifted(row, "EMax", "HitShift", label)
        definition.minimum_damage_per_level_raw = damage_gains(row, "EMinLev", label)
        definition.maximum_damage_per_level_raw = damage_gains(row, "EMaxLev", label)
        definition.damage_synergy_skill_ids, definition.damage_synergy_percent_per_level =
            skill_modifiers.hard_level_sum_percent(row, "EDmgSymPerCalc", "Param8", skills_by_name, label)
        definition.reaction_state_id = assert(states_by_name.cold, label .. " requires the cold state").state
        definition.reaction_chill_state_id = definition.reaction_state_id
        definition.reaction_stat = "velocitypercent"
        definition.reaction_stat_value = -50
        definition.reaction_duration_base = required_integer(row, "ELen", label)
        for tier = 1, 5 do
            definition.reaction_duration_per_level[tier] = integer_or(row, "ELevLen" .. tier, 0)
        end
        assert(definition.reaction_duration_base > 0, label .. " has no cold length")
        if reaction == "melee_attack_cold" then
            assert(not definition.reaction_missile, label .. " melee reaction unexpectedly has a missile")
        else
            assert(definition.reaction_missile, label .. " ranged reaction has no missile")
        end
    end
    if definition.reaction_missile then
        definition.reaction_missile.skill_id = id
    end
    return definition
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported self-state skill IDs are required")
    local rows = assert(records.load("data/global/excel/skills.txt"))
    local states_by_name = index(assert(records.load("data/global/excel/states.txt")), "state")
    local missiles = index(assert(records.load("data/global/excel/Missiles.txt")), "Missile")
    local skills = index(rows, "Id")
    local skills_by_name = {}
    for _, row in ipairs(rows) do
        if row.skill and row.skill ~= "" then
            skills_by_name[string.lower(row.skill)] = row
        end
    end
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        -- Narrow module fixtures do not have to reproduce every declared
        -- retail row. The mounted-data coverage report is the strict target
        -- completeness gate; any row present here is still decoded by ID.
        if row and row.srvdofunc == "18" then
            local definition = decode(row, skills_by_name, states_by_name, missiles)
            assert(not result[definition.skill_id], "duplicate self-state skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
