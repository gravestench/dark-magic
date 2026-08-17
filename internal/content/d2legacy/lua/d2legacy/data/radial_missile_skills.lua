-- Decode exact-ID, targetless radial missile skills from owned 1.14d records.

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
    assert(value and value >= (minimum or 0) and value == math.floor(value), label .. " has invalid " .. column)
    return value
end

local function shifted(row, column, label)
    return integer(row, column, label) * (2 ^ integer(row, "HitShift", label))
end

local function shifted_mana(row, column, label)
    return integer(row, column, label) * (2 ^ integer(row, "manashift", label))
end

local function damage_gains(row, prefix, label)
    local result = {}
    for tier = 1, 5 do
        result[tier] = shifted(row, prefix .. tier, label)
    end
    return result
end

local channels = { fire = "fire", ltng = "lightning", cold = "cold", mag = "magic", pois = "poison" }

local function decode(skill, missiles)
    local id = integer(skill, "Id", "radial-missile skill")
    local label = skill.skill or ("skill " .. id)
    assert(skill.srvstfunc == "" and skill.srvdofunc == "22", label .. " has unsupported server functions")
    assert(skill.cltstfunc == "" and skill.cltdofunc == "25", label .. " has unsupported client functions")
    assert(
        skill.anim == "SC" and skill.range == "none" and skill.interrupt == "1",
        label .. " has unsupported cast policy"
    )
    local missile_id = assert(skill.srvmissilea ~= "" and skill.srvmissilea, label .. " has no radial missile")
    assert(skill.srvmissileb == missile_id and skill.srvmissilec == missile_id, label .. " radial missile slots differ")
    local missile = assert(missiles[string.lower(missile_id)], label .. " missile is missing")
    assert(missile.Skill == label and missile.pSrvDoFunc == "1", label .. " has unsupported missile function")
    assert(missile.CollideType == "3" and missile.CollideKill == "", label .. " has unsupported collision policy")
    assert(missile.LastCollide == "1" and missile.NextHit == "1", label .. " has unsupported repeat-contact policy")
    local channel = assert(channels[skill.EType], label .. " has unsupported damage channel " .. tostring(skill.EType))
    local cel = assert(missile.CelFile ~= "" and missile.CelFile, label .. " missile has no CelFile")
    local animation_speed = tonumber(missile.AnimSpeed) or 16
    return {
        behavior = "missile.radial",
        trajectory = "radial",
        skill_id = id,
        mana_cost_raw = shifted_mana(skill, "mana", label),
        mana_cost_per_level_raw = shifted_mana(skill, "lvlmana", label),
        minimum_mana_cost_raw = shifted_mana(skill, "minmana", label),
        minimum_damage_raw = shifted(skill, "EMin", label),
        maximum_damage_raw = shifted(skill, "EMax", label),
        minimum_damage_per_level_raw = damage_gains(skill, "EMinLev", label),
        maximum_damage_per_level_raw = damage_gains(skill, "EMaxLev", label),
        missile_count_base = integer(skill, "Param1", label, 1),
        missile_count_per_level = integer(skill, "Param2", label),
        effect_delay = 1,
        complete_delay = 2,
        requires_point_target = false,
        speed_per_tick = integer(missile, "Vel", label, 1) / 25,
        lifetime_ticks = integer(missile, "Range", label, 1),
        collision_radius = integer(missile, "Size", label, 1) / 2,
        destroy_on_contact = false,
        next_hit_delay = integer(missile, "NextDelay", label, 1),
        impact_radius = 0,
        impact_missile_id = "",
        impact_dcc = "",
        impact_palette = "",
        impact_lifetime_ticks = 0,
        impact_directions = 1,
        impact_frames_per_second = 1,
        impact_loop = false,
        impact_sound = "",
        damage_channel = channel,
        missile_id = missile_id,
        dcc = "data/global/missiles/" .. cel .. ".dcc",
        palette = "data/global/palette/units/pal.dat",
        travel_sound = missile.TravelSound or "",
        hit_sound = missile.HitSound or "",
        directions = math.max(tonumber(missile.NumDirections) or 1, 1),
        frames_per_second = math.max(math.floor(animation_speed * 25 / 16), 1),
        loop = missile.LoopAnim == "1",
        offset_x = tonumber(missile.xoffset) or 0,
        offset_y = tonumber(missile.yoffset) or 0,
        offset_z = tonumber(missile.zoffset) or 0,
    }
end

function M.load(supported_ids)
    assert(type(supported_ids) == "table", "supported radial-missile skill IDs are required")
    local skills = index(assert(records.load("data/global/excel/skills.txt")), "Id")
    local missiles = index(assert(records.load("data/global/excel/Missiles.txt")), "Missile")
    local result = {}
    for _, skill_id in ipairs(supported_ids) do
        local row = skills[tostring(skill_id)]
        if row then
            local definition = decode(row, missiles)
            assert(not result[definition.skill_id], "duplicate radial-missile skill ID")
            result[definition.skill_id] = definition
        end
    end
    return result
end

return M
