-- Commit Diablo monster death exactly once when health reaches zero.
--
-- The corpse entity remains because later skills and interactions may consume
-- it. AI, collision, and selection are removed. Experience goes to a player
-- killer, and deterministic loot facts are copied into durable death state.

local ecs = require("engine.ecs/v1")
local loot = require("d2legacy.loot.generate")
local attribution = require("d2legacy.owned_units.attribution")
local player_count = require("d2legacy.policy.player_count")
local party = require("d2legacy.policy.party")
local M = {}

local function selectable_by_id(entities, wanted)
    for _, entity in ipairs(entities) do
        local selected = ecs.get(entity, "d2legacy.world.selectable")
        if selected and selected:get("id") == wanted then
            return entity
        end
    end
    return nil
end

local function collect_killers(entities, structural)
    local killers = {}
    for _, entity in ipairs(entities) do
        local event = ecs.get(entity, "d2legacy.combat.event")
        if event then
            structural:set(entity, "d2legacy.combat.death_observed", {})
            if event:get("kind") == "unit_died" then
                killers[event:get("target_id")] = event:get("attacker_id")
            end
        end
    end
    return killers
end

local function monster_selectable_id(monster, identity)
    local selected = ecs.get(monster, "d2legacy.world.selectable")
    if selected then
        return selected:get("id")
    end
    return "monster:" .. identity:get("spawn_id")
end

local function credit_experience(entities, killer_id, amount)
    local killer = selectable_by_id(entities, killer_id)
    local progress = killer and ecs.get(killer, "d2legacy.player.progress")
    if not progress or amount <= 0 then
        return
    end
    progress:set("experience", progress:get("experience") + amount)
end

local function count_game_players(entities)
    local count = 0
    for _, entity in ipairs(entities) do
        if ecs.get(entity, "d2legacy.player.identity") then
            count = count + 1
        end
    end
    return math.max(count, 1)
end

local function no_drop_context(entities, credited_id, monster_player_count)
    local credited = selectable_by_id(entities, credited_id)
    local credited_identity = credited and ecs.get(credited, "d2legacy.player.identity")
    local nearby_party_members = 0
    if credited_identity then
        nearby_party_members = party.additional_living_members_in_same_level(credited_identity:get("player"), entities)
    end
    return player_count.no_drop(count_game_players(entities), nearby_party_members, monster_player_count)
end

local function roll_loot(entities, identity, stats, credited_id)
    local monster_player_count = math.max(stats:get("spawn_player_count"), 1)
    local context = no_drop_context(entities, credited_id, monster_player_count)
    local drops = loot.roll(identity:get("treasure_class"), {
        version = 100,
        monster_level = stats:get("level"),
        magic_find = 0,
        player_count = context,
    })
    return loot.encode(drops), context
end

local function death_values(context, monster, identity, killer, credited, experience, drops, counts)
    return {
        tick = context.tick,
        killer_id = killer,
        credited_id = credited,
        xp = experience,
        loot_seed = identity:get("seed"),
        treasure_class = identity:get("treasure_class"),
        drops = drops,
        game_player_count = counts.game_player_count,
        effective_player_count = counts.effective_player_count,
        nearby_party_member_count = counts.nearby_party_member_count,
        monster_player_count = counts.monster_player_count,
        no_drop_player_count = counts.no_drop_player_count,
        active = false,
        corpse_usable = ecs.get(monster, "d2legacy.monster.corpse_selectable") ~= nil,
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
                game_player_count = values.game_player_count,
                effective_player_count = values.effective_player_count,
                nearby_party_member_count = values.nearby_party_member_count,
                monster_player_count = values.monster_player_count,
                no_drop_player_count = values.no_drop_player_count,
            },
        })
    end
end

local function stop_monster(monster, structural)
    local appearance = ecs.get(monster, "d2legacy.monster.appearance")
    if appearance then
        appearance:set("mode", "DT")
    end

    local velocity = ecs.get(monster, "d2legacy.world.velocity")
    if velocity then
        velocity:set("x", 0)
        velocity:set("y", 0)
    end
    structural:remove(monster, "d2legacy.monster.ai")
    structural:remove(monster, "d2legacy.world.collider")
    local selectable = ecs.get(monster, "d2legacy.world.selectable")
    if selectable and ecs.get(monster, "d2legacy.monster.corpse_selectable") then
        selectable:set("kind", "corpse")
        selectable:set("owner", "")
        selectable:set("priority", 5)
    else
        structural:remove(monster, "d2legacy.world.selectable")
    end
    structural:remove(monster, "engine.world.velocity_mover")
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
    local ownership = attribution.resolve(entities, killer)
    local credited = ownership.ultimate_owner_id
    local experience = stats:get("experience")
    credit_experience(entities, credited, experience)

    local drops, counts = roll_loot(entities, identity, stats, credited)
    local values = death_values(context, monster, identity, killer, credited, experience, drops, counts)
    structural:set(monster, "d2legacy.monster.death", values)
    stop_monster(monster, structural)
    emit_events(structural, context, identity, values)
end

local function update(context, entities, structural)
    local killers = collect_killers(entities, structural)
    for _, monster in ipairs(entities) do
        commit_death(context, entities, structural, monster, killers)
    end
end

function M.register()
    ecs.system({
        id = "d2legacy.monster.death",
        phase = "effects",
        after = { "d2legacy.combat.reflect_melee" },
        query = {
            any = {
                "d2legacy.monster.stats",
                "d2legacy.combat.event",
                -- Player entities carry the selectable ID used for kill
                -- credit, but no monster/event component. Include them in the
                -- system snapshot so credited XP can reach player.progress.
                "d2legacy.world.selectable",
                "d2legacy.player.identity",
            },
            none = { "d2legacy.world.inactive", "d2legacy.combat.death_observed" },
        },
        read = {
            "d2legacy.monster.stats",
            "d2legacy.monster.identity",
            "d2legacy.monster.death",
            "d2legacy.monster.corpse_selectable",
            "d2legacy.combat.event",
            "d2legacy.world.selectable",
            "d2legacy.player.identity",
            "d2legacy.player.vitals",
            "d2legacy.world.location",
            "d2legacy.owned_unit",
            "d2legacy.player.progress",
            "d2legacy.monster.appearance",
            "d2legacy.combat.death_observed",
        },
        write = {
            "d2legacy.monster.death",
            "d2legacy.monster.death_event",
            "d2legacy.monster.ai",
            "d2legacy.world.velocity",
            "d2legacy.world.collider",
            "d2legacy.world.selectable",
            "engine.world.velocity_mover",
            "d2legacy.player.progress",
            "d2legacy.monster.appearance",
            "d2legacy.combat.death_observed",
        },
        update = update,
    })
end

return M
