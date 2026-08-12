-- Commit Diablo monster death exactly once when health reaches zero.
--
-- The corpse entity remains because later skills and interactions may consume
-- it. AI, collision, and selection are removed. Experience goes to a player
-- killer, and deterministic loot facts are copied into durable death state.

local ecs = require("engine.ecs/v1")
local loot = require("d2legacy.loot.generate")
local M = {}

local function selectable_by_id(entities, wanted)
    for _, entity in ipairs(entities) do
        local selected = ecs.get(entity, "d2legacy.world.selectable")
        if selected and selected:get("id") == wanted then return entity end
    end
    return nil
end

local function collect_killers(entities)
    local killers = {}
    for _, entity in ipairs(entities) do
        local event = ecs.get(entity, "d2legacy.combat.melee_event")
            or ecs.get(entity, "d2legacy.combat.event")
        if event and event:get("remaining_health_raw") == 0 then
            killers[event:get("target_id")] = event:get("attacker_id")
        end
    end
    return killers
end

local function monster_selectable_id(monster, identity)
    local selected = ecs.get(monster, "d2legacy.world.selectable")
    if selected then return selected:get("id") end
    return "monster:" .. identity:get("spawn_id")
end

local function credit_experience(entities, killer_id, amount)
    local killer = selectable_by_id(entities, killer_id)
    local progress = killer and ecs.get(killer, "d2legacy.player.progress")
    if not progress or amount <= 0 then return end
    progress:set("experience", progress:get("experience") + amount)
end

local function roll_loot(identity, stats)
    local drops = loot.roll(identity:get("treasure_class"), {
        version = 100,
        monster_level = stats:get("level"),
        magic_find = 0,
    })
    return loot.encode(drops)
end

local function death_values(context, identity, killer, experience, drops)
    return {
        tick = context.tick,
        killer_id = killer,
        credited_id = killer,
        xp = experience,
        loot_seed = identity:get("seed"),
        treasure_class = identity:get("treasure_class"),
        drops = drops,
        active = false,
        corpse_usable = true,
    }
end

local function emit_events(structural, context, identity, values)
    local kinds = {
        "monster_killed",
        "monster_loot",
        "monster_quest_kill",
        "monster_death_presented",
    }
    for _, kind in ipairs(kinds) do
        structural:create({
            ["d2legacy.monster.death_event"] = {
                kind = kind,
                tick = context.tick,
                monster_id = identity:get("spawn_id"),
                killer_id = values.killer_id,
                credited_id = values.credited_id,
                xp = values.xp,
                loot_seed = values.loot_seed,
                treasure_class = values.treasure_class,
                drops = values.drops,
            },
        })
    end
end

local function stop_monster(monster, structural)
    local appearance = ecs.get(monster, "d2legacy.monster.appearance")
    if appearance then appearance:set("mode", "DT") end

    local velocity = ecs.get(monster, "d2legacy.world.velocity")
    if velocity then
        velocity:set("x", 0)
        velocity:set("y", 0)
    end
    structural:remove(monster, "d2legacy.monster.ai")
    structural:remove(monster, "d2legacy.world.collider")
    structural:remove(monster, "d2legacy.world.selectable")
end

local function commit_death(context, entities, structural, monster, killers)
    local stats = ecs.get(monster, "d2legacy.monster.stats")
    local identity = ecs.get(monster, "d2legacy.monster.identity")
    local already_dead = ecs.get(monster, "d2legacy.monster.death")
    if not stats or not identity or stats:get("health") > 0 or already_dead then
        return
    end

    local monster_id = monster_selectable_id(monster, identity)
    local killer = killers[monster_id] or ""
    local experience = stats:get("experience")
    credit_experience(entities, killer, experience)

    local drops = roll_loot(identity, stats)
    local values = death_values(
        context,
        identity,
        killer,
        experience,
        drops
    )
    structural:set(monster, "d2legacy.monster.death", values)
    stop_monster(monster, structural)
    emit_events(structural, context, identity, values)
end

local function update(context, entities, structural)
    local killers = collect_killers(entities)
    for _, monster in ipairs(entities) do
        commit_death(context, entities, structural, monster, killers)
    end
end

function M.register()
    ecs.system({
        id = "d2legacy.monster.death",
        phase = "effects",
        query = {
            any = {
                "d2legacy.monster.stats",
                "d2legacy.combat.melee_event",
                "d2legacy.combat.event",
            },
        },
        read = {
            "d2legacy.monster.stats",
            "d2legacy.monster.identity",
            "d2legacy.monster.death",
            "d2legacy.combat.melee_event",
            "d2legacy.combat.event",
            "d2legacy.world.selectable",
            "d2legacy.player.progress",
            "d2legacy.monster.appearance",
        },
        write = {
            "d2legacy.monster.death",
            "d2legacy.monster.death_event",
            "d2legacy.monster.ai",
            "d2legacy.world.velocity",
            "d2legacy.world.collider",
            "d2legacy.world.selectable",
            "d2legacy.player.progress",
            "d2legacy.monster.appearance",
        },
        update = update,
    })
end

return M
