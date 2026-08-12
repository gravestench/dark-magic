-- Materialize a trusted population recipe as one authoritative monster.
-- Population planners may remain adapters while this module owns what a
-- Diablo monster entity means and which state it receives.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local M = {}

local function encode_components(values)
    local keys = {}
    for key in pairs(values or {}) do table.insert(keys, key) end
    table.sort(keys)
    local result = {}
    for _, key in ipairs(keys) do table.insert(result, key .. "=" .. values[key]) end
    return table.concat(result, ",")
end

local function existing(id)
    for _, entity in ipairs(ecs.query({all={"d2legacy.monster.identity"}})) do
        if ecs.get(entity,"d2legacy.monster.identity"):get("spawn_id") == id then return true end
    end
    return false
end

function M.validate(command)
    local spawn, definition = command.payload, command.payload and command.payload.definition
    assert(type(spawn)=="table" and type(definition)=="table", "spawn definition is required")
    assert(type(spawn.spawn_id)=="string" and spawn.spawn_id~="", "spawn ID is required")
    assert(type(spawn.x)=="number" and type(spawn.y)=="number", "spawn position is required")
    assert(type(definition.id)=="string" and definition.id~="", "monster definition ID is required")
end

function M.apply(command)
    local spawn, definition = command.payload, command.payload.definition
    assert(not existing(spawn.spawn_id), "monster spawn already exists")
    -- Definitions carry reviewed 8.8 raw values. A deterministic midpoint is
    -- sufficient until spawn variance is moved onto its named Lua RNG stream.
    local health = math.floor((definition.life_min + definition.life_max) / 2)
    ecs.create({
        ["d2legacy.monster.identity"]={spawn_id=spawn.spawn_id,definition_id=definition.id,
            base_id=definition.base_id or "",graphics_id=definition.graphics_id or "",
            seed=tostring(spawn.seed),treasure_class=definition.treasure_class or ""},
        ["d2legacy.monster.stats"]={level=definition.level,health=health,max_health=health,
            defense=definition.defense,attack_rating=definition.attack_rating,
            physical_min=definition.physical_min,physical_max=definition.physical_max,
            experience=definition.experience},
        ["d2legacy.combat.melee_profile"]={range=definition.attack_range,
            physical_min=definition.physical_min,physical_max=definition.physical_max},
        ["d2legacy.monster.appearance"]={token=definition.token,mode="NU",
            weapon_class=definition.weapon_class,name_key=definition.name_key,
            components=encode_components(definition.components),death_sound=definition.death_sound or ""},
        ["d2legacy.monster.ai"]={behavior=definition.ai,state="idle",target_id="",
            next_think_tick=command.tick,think_interval=definition.think_interval,
            aggro_radius=definition.aggro_radius,attack_range=definition.attack_range,speed=definition.velocity},
        ["d2legacy.world.position"]={x=spawn.x,y=spawn.y},
        ["d2legacy.world.velocity"]={x=0,y=0},
        ["d2legacy.world.location"]={act=spawn.act,level_id=spawn.level_id},
        ["d2legacy.world.collider"]={radius=definition.collider_radius},
        ["d2legacy.world.selectable"]={id="monster:"..spawn.spawn_id,kind="hostile",
            label=definition.name_key,owner="",radius=definition.select_radius,priority=20},
    })
end

function M.register()
    commands.register({kind="system.monster.spawn",authorities={"system","administrator"},
        validate=M.validate,apply=M.apply})
end

return M
