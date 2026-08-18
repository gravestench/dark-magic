-- Materialize a trusted population recipe as one authoritative monster.
-- Population planners may remain adapters while this module owns what a
-- Diablo monster entity means and which state it receives.

local commands = require("engine.authority_command/v1")
local ecs = require("engine.ecs/v1")
local random = require("engine.authority_random/v1")
local player_count = require("d2legacy.policy.player_count")
local M = {}

local function encode_components(values)
    local keys = {}
    for key in pairs(values or {}) do
        table.insert(keys, key)
    end
    table.sort(keys)
    local result = {}
    for _, key in ipairs(keys) do
        table.insert(result, key .. "=" .. values[key])
    end
    return table.concat(result, ",")
end

local function existing(id)
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.monster.identity" } })) do
        if ecs.get(entity, "d2legacy.monster.identity"):get("spawn_id") == id then
            return true
        end
    end
    return false
end

local function count_game_players()
    return math.max(#ecs.query({ all = { "d2legacy.player.identity" } }), 1)
end

local function percentage(value, percent)
    return math.floor(value * percent / 100)
end

function M.validate(command)
    local spawn, definition = command.payload, command.payload and command.payload.definition
    assert(type(spawn) == "table" and type(definition) == "table", "spawn definition is required")
    assert(type(spawn.spawn_id) == "string" and spawn.spawn_id ~= "", "spawn ID is required")
    assert(type(spawn.x) == "number" and type(spawn.y) == "number", "spawn position is required")
    assert(type(definition.id) == "string" and definition.id ~= "", "monster definition ID is required")
end

function M.components(command, game_player_count)
    local spawn, definition = command.payload, command.payload.definition
    -- Monster life endpoints are authored as whole values encoded in Diablo's
    -- 8.8 fixed-point unit. Draw a whole point, then restore the raw scale.
    -- The named engine stream is checkpointed, replayed, and unavailable to
    -- presentation code; no clock or process-global random state leaks in.
    assert(
        definition.life_min % 256 == 0 and definition.life_max % 256 == 0,
        "monster life endpoints must be whole 8.8 values"
    )
    local minimum, maximum = definition.life_min / 256, definition.life_max / 256
    local base_health = minimum + random.integer("d2legacy.monster.spawn.life", maximum - minimum + 1)
    local scaling = player_count.monster_spawn(game_player_count, definition.evil ~= false)
    local health = (base_health + percentage(base_health, scaling.life_bonus_percent)) * 256
    local experience = definition.experience + percentage(definition.experience, scaling.experience_bonus_percent)
    local result = {
        ["d2legacy.monster.identity"] = {
            spawn_id = spawn.spawn_id,
            definition_id = definition.id,
            base_id = definition.base_id or "",
            graphics_id = definition.graphics_id or "",
            seed = tostring(spawn.seed),
            treasure_class = definition.treasure_class or "",
        },
        ["d2legacy.monster.stats"] = {
            level = definition.level,
            spawn_player_count = scaling.effective_player_count,
            health = health,
            max_health = health,
            defense = definition.defense,
            attack_rating = definition.attack_rating,
            physical_min = definition.physical_min,
            physical_max = definition.physical_max,
            experience = experience,
        },
        ["d2legacy.combat.melee_profile"] = {
            range = definition.attack_range,
            physical_min = definition.physical_min,
            physical_max = definition.physical_max,
        },
        ["d2legacy.combat.knockback_target"] = {
            mode_supported = definition.knockback_mode == true,
            size_class = definition.knockback_size or "normal",
        },
        ["d2legacy.monster.appearance"] = {
            token = definition.token,
            mode = "NU",
            weapon_class = definition.weapon_class,
            name_key = definition.name_key,
            components = encode_components(definition.components),
            death_sound = definition.death_sound or "",
            overlay_height = definition.overlay_height or 0,
        },
        ["d2legacy.monster.ai"] = {
            behavior = definition.ai,
            state = "idle",
            target_id = "",
            next_think_tick = command.tick,
            think_interval = definition.think_interval,
            aggro_radius = definition.aggro_radius,
            attack_range = definition.attack_range,
            speed = definition.velocity,
        },
        ["d2legacy.world.position"] = { x = spawn.x, y = spawn.y },
        ["d2legacy.world.velocity"] = { x = 0, y = 0 },
        ["d2legacy.world.facing"] = { direction = 0, directions = 8 },
        ["d2legacy.world.location"] = { act = spawn.act, level_id = spawn.level_id },
        ["d2legacy.world.collider"] = { radius = definition.collider_radius },
        ["d2legacy.world.occupancy"] = { blocks_movement = true },
        ["engine.world.velocity_mover"] = {},
        ["d2legacy.world.selectable"] = {
            id = "monster:" .. spawn.spawn_id,
            kind = definition.evil == false and "friendly" or "hostile",
            label = definition.name_key,
            owner = "",
            radius = definition.select_radius,
            priority = 20,
        },
    }
    if definition.corpse_selectable == true then
        result["d2legacy.monster.corpse_selectable"] = {}
    end
    if definition.revivable == true then
        result["d2legacy.monster.revivable"] = {}
    end
    return result
end

function M.apply(command)
    local spawn = command.payload
    assert(not existing(spawn.spawn_id), "monster spawn already exists")
    ecs.create(M.components(command, count_game_players()))
end

-- Startup population is already trusted mod policy and does not need to forge
-- a synthetic transport command. Reusing this function keeps one monster
-- materializer for startup, admin spawning, replay, and tests.
M.materialize = M.apply

function M.register()
    commands.register({
        kind = "system.monster.spawn",
        authorities = { "system", "administrator" },
        validate = M.validate,
        apply = M.apply,
    })
end

return M
