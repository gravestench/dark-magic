-- Interpret the legacy Skills.txt and Missiles.txt rows for Fire Bolt.
--
-- Go decodes files and exposes immutable strings. This module owns the Diablo
-- rule that skill 36 plus the reviewed function/flag combination means a
-- straight fire projectile. Unknown combinations fail instead of being guessed.

local records = require("engine.records/v1")

local M = {}

local function find(rows, column, wanted)
    for _, row in ipairs(rows) do
        if row[column] == wanted then return row end
    end
    return nil
end

local function required_integer(row, column)
    local value = tonumber(row[column])
    assert(value and value >= 0 and value == math.floor(value),
        "Fire Bolt has invalid " .. column)
    return value
end

local function shifted(row, value_column, shift_column)
    return required_integer(row, value_column) * (2 ^ required_integer(row, shift_column))
end

local function integer_or(row, column, fallback)
    local value = tonumber(row[column])
    if value == nil then return fallback end
    return math.floor(value)
end

function M.load()
    local skills = assert(records.load("data/global/excel/skills.txt"))
    local skill = assert(find(skills, "Id", "36"), "Fire Bolt skill 36 is missing")
    assert(skill.skill == "Fire Bolt" and skill.srvmissile == "firebolt")
    assert(skill.etype == "fire" and skill.interrupt == "1")
    assert(skill.srvstfunc == "" and skill.srvdofunc == "")

    local missiles = assert(records.load("data/global/excel/Missiles.txt"))
    local missile = assert(find(missiles, "Missile", "firebolt"),
        "Fire Bolt missile is missing")
    assert(missile.Skill == "Fire Bolt" and missile.pSrvDoFunc == "1")
    assert(missile.CollideType == "3" and missile.CollideKill == "1")

    local velocity = required_integer(missile, "Vel")
    local lifetime = required_integer(missile, "Range")
    local animation_speed = integer_or(missile, "AnimSpeed", 16)
    local cel = assert(missile.CelFile and missile.CelFile ~= "" and missile.CelFile,
        "Fire Bolt missile has no CelFile")
    return {
        skill_id = 36,
        mana_cost_raw = shifted(skill, "mana", "manashift"),
        minimum_damage_raw = shifted(skill, "emin", "HitShift"),
        maximum_damage_raw = shifted(skill, "emax", "HitShift"),
        effect_delay = 1,
        complete_delay = 2,
        speed_per_tick = velocity / 25,
        lifetime_ticks = lifetime,
        maximum_range = velocity * lifetime / 25,
        collision_radius = required_integer(missile, "Size") / 2,
        damage_channel = "fire",

        -- These are presentation facts copied from the same immutable record.
        -- They travel with the projectile so render code remains a passive
        -- observer and never reinterprets Diablo data on its own.
        missile_id = "firebolt",
        dcc = "data/global/missiles/" .. cel .. ".dcc",
        palette = "data/global/palette/units/pal.dat",
        travel_sound = missile.TravelSound or "",
        hit_sound = missile.HitSound or "",
        directions = math.max(integer_or(missile, "NumDirections", 1), 1),
        frames_per_second = math.max(math.floor(animation_speed * 25 / 16), 1),
        loop = missile.LoopAnim == "1",
        offset_x = integer_or(missile, "Xoffset", 0),
        offset_y = integer_or(missile, "Yoffset", 0),
        offset_z = integer_or(missile, "Zoffset", 0),
    }
end

return M
