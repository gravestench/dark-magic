-- Materialize one trusted durable character as a live Diablo II player.
--
-- Go owns save I/O, authentication, and command admission. The meaning of a
-- Diablo player entity belongs here: its components, initial mouse skills,
-- belt shape, collision size, and legacy composite appearance are mod policy.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local skills = require("d2legacy.data.skill")
local M = {}

local function finite(value)
    return type(value) == "number" and value == value
        and value ~= math.huge and value ~= -math.huge
end

local function present(value)
    return type(value) == "string" and value:match("%S") ~= nil
end

local function already_entered(character_id)
    for _, entity in ipairs(ecs.query({all={"d2legacy.player.identity"}})) do
        if ecs.get(entity, "d2legacy.player.identity"):get("character_id") == character_id then
            return true
        end
    end
    return false
end

function M.validate(command)
    local p = command.payload
    assert(type(p) == "table", "player entry payload must be a table")
    for _, field in ipairs({"character_id","player","name","class"}) do
        assert(present(p[field]), "player entry " .. field .. " is required")
    end
    assert(p.level >= 1 and p.health >= 0 and p.max_health >= p.health,
        "player entry progression or health is invalid")
    assert(p.mana >= 0 and p.max_mana >= p.mana, "player entry mana is invalid")
    assert(finite(p.x) and finite(p.y) and finite(p.world_width) and finite(p.world_height),
        "player entry world values must be finite")
    assert(p.world_width > 0 and p.world_height > 0 and p.x >= 0 and p.y >= 0
        and p.x < p.world_width and p.y < p.world_height, "player entry is outside the world")
    assert(p.act >= 1 and p.act <= 5 and p.level_id > 0, "player entry location is invalid")
end

local class_tokens={amazon="AM",sorceress="SO",necromancer="NE",paladin="PA",barbarian="BA",assassin="AI",druid="DZ"}

local function initial_skills(skills)
    local left, right, left_chosen, right_chosen = 0, 0, false, false
    for _, skill in ipairs(skills or {}) do
        assert(skill.id >= 0 and skill.level >= 1 and skill.list_row >= 0,
            "learned skill is invalid")
        assert(skill.left_allowed or skill.right_allowed, "learned skill is not assignable")
        if not left_chosen and skill.left_allowed then left, left_chosen = skill.id, true end
        if not right_chosen and skill.right_allowed then right, right_chosen = skill.id, true end
    end
    return left, right
end

function M.apply(command)
    local p = command.payload
    assert(not already_entered(p.character_id), "character already entered")
    local learned = skills.starting_for_class(p.class)
    local left, right = initial_skills(learned)
    local player = ecs.create({
        ["d2legacy.player.identity"]={character_id=p.character_id,player=p.player,name=p.name,class=p.class},
        ["d2legacy.player.progress"]={level=p.level,experience=p.experience},
        -- Legacy base attack rating is five points per Dexterity. Equipment and
        -- skills add their own removable sources later; the host supplies only
        -- the durable Dexterity fact and does not interpret the D2 formula.
        ["d2legacy.player.combat_stats"]={attack_rating=(p.dexterity or 0)*5,defense=p.defense or 0},
        ["d2legacy.combat.defense"]={base_physical_resist=p.physical_resistance or 0,
            base_fire_resist=p.fire_resistance or 0,physical_resist=p.physical_resistance or 0,
            fire_resist=p.fire_resistance or 0,max_fire_resist=75,physical_reduction_raw=0},
        ["d2legacy.player.vitals"]={health=p.health,max_health=p.max_health,mana=p.mana,max_mana=p.max_mana,
            mana_raw=p.mana*256,max_mana_raw=p.max_mana*256},
        -- D2 synthesizes an unarmed profile when no weapon contributes one.
        ["d2legacy.combat.melee_profile"]={range=2,physical_min=256,physical_max=512},
        ["d2legacy.player.appearance"]={cof=p.cof or "",token=assert(class_tokens[string.lower(p.class)],"unknown player class"),
            palette=present(p.palette) and p.palette or "data/global/Palette/units/pal.dat",
            weapon_class="HTH"},
        ["d2legacy.player.animation"]={direction=p.direction or 0,mode="NU"},
        ["d2legacy.world.position"]={x=p.x,y=p.y},
        ["d2legacy.world.velocity"]={x=0,y=0},
        ["d2legacy.world.facing"]={direction=p.direction or 0,directions=16},
        ["d2legacy.player.movement_mode"]={running=false},
        ["d2legacy.player.skill_assignment"]={left=left,right=right},
        ["d2legacy.player.skill_intent"]={side="",skill_id=0,target_x=p.x,target_y=p.y,target_id=""},
        ["d2legacy.player.belt"]={capacity=4},
        ["d2legacy.world.player_control"]={player=p.player},
        ["d2legacy.world.bounds"]={width=p.world_width,height=p.world_height},
        ["d2legacy.world.location"]={act=p.act,level_id=p.level_id},
        -- A player occupies two subtiles across in legacy D2: radius one.
        ["d2legacy.world.collider"]={radius=1},
        ["d2legacy.world.selectable"]={id="player:"..p.player,kind="player",label=p.name,
            owner=p.player,radius=0.75,priority=10},
    })
    for _, skill in ipairs(learned) do
        ecs.create({["d2legacy.player.learned_skill"]={owner=player,skill_id=skill.id,
            level=skill.level,list_row=skill.list_row,left_allowed=skill.left_allowed,
            right_allowed=skill.right_allowed}})
    end
end

function M.register()
    commands.register({kind="system.player.enter",authorities={"system","administrator"},
        validate=M.validate,apply=M.apply})
end

return M
