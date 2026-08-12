-- Commit Diablo monster death once when health reaches zero.
--
-- The corpse remains as an entity because Diablo rules and skills may consume
-- it later. Collision, targeting, and AI are removed; presentation switches to
-- the death mode; XP is credited to the killer when it is a player.

local ecs = require("engine.ecs/v1")
local loot = require("d2legacy.loot.generate")
local M = {}

local function by_id(entities, wanted)
    for _, entity in ipairs(entities) do
        local selected = ecs.get(entity, "d2legacy.world.selectable")
        if selected and selected:get("id") == wanted then return entity end
    end
end

local function emit(structural, kind, context, identity, killer, credited, xp, treasure, drops)
    structural:create({["d2legacy.monster.death_event"]={kind=kind,tick=context.tick,
        monster_id=identity:get("spawn_id"),killer_id=killer,credited_id=credited,
        xp=xp,loot_seed=identity:get("seed"),treasure_class=treasure,drops=drops}})
end

function M.register()
    ecs.system({id="d2legacy.monster.death",phase="effects",
        query={any={"d2legacy.monster.stats","d2legacy.combat.melee_event","d2legacy.combat.event"}},
        read={"d2legacy.monster.stats","d2legacy.monster.identity","d2legacy.monster.death",
            "d2legacy.combat.melee_event","d2legacy.combat.event","d2legacy.world.selectable",
            "d2legacy.player.progress","d2legacy.monster.appearance"},
        write={"d2legacy.monster.death","d2legacy.monster.death_event","d2legacy.monster.ai",
            "d2legacy.world.velocity","d2legacy.world.collider","d2legacy.world.selectable",
            "d2legacy.player.progress","d2legacy.monster.appearance"},
        update=function(context, entities, structural)
            local killers = {}
            for _, event_entity in ipairs(entities) do
                local event = ecs.get(event_entity,"d2legacy.combat.melee_event") or ecs.get(event_entity,"d2legacy.combat.event")
                if event and event:get("remaining_health_raw") == 0 then killers[event:get("target_id")] = event:get("attacker_id") end
            end
            for _, monster in ipairs(entities) do
                local stats, identity = ecs.get(monster,"d2legacy.monster.stats"), ecs.get(monster,"d2legacy.monster.identity")
                if stats and identity and stats:get("health") <= 0 and not ecs.get(monster,"d2legacy.monster.death") then
                    local selected = ecs.get(monster,"d2legacy.world.selectable")
                    local monster_id = selected and selected:get("id") or "monster:"..identity:get("spawn_id")
                    local killer = killers[monster_id] or ""
                    local credited, xp = killer, stats:get("experience")
                    local killer_entity = by_id(entities, killer)
                    local progress = killer_entity and ecs.get(killer_entity,"d2legacy.player.progress")
                    if progress and xp > 0 then progress:set("experience",progress:get("experience")+xp) end
                    local generated=loot.encode(loot.roll(identity:get("treasure_class"),{version=100,monster_level=stats:get("level"),magic_find=0}))
                    structural:set(monster,"d2legacy.monster.death",{tick=context.tick,killer_id=killer,
                        credited_id=credited,xp=xp,loot_seed=identity:get("seed"),
                        treasure_class=identity:get("treasure_class"),drops=generated,active=false,corpse_usable=true})
                    local appearance = ecs.get(monster,"d2legacy.monster.appearance")
                    if appearance then appearance:set("mode","DT") end
                    local velocity = ecs.get(monster,"d2legacy.world.velocity")
                    if velocity then velocity:set("x",0); velocity:set("y",0) end
                    structural:remove(monster,"d2legacy.monster.ai")
                    structural:remove(monster,"d2legacy.world.collider")
                    structural:remove(monster,"d2legacy.world.selectable")
                    for _, kind in ipairs({"monster_killed","monster_loot","monster_quest_kill","monster_death_presented"}) do
                        emit(structural,kind,context,identity,killer,credited,xp,identity:get("treasure_class"),generated)
                    end
                end
            end
        end})
end

return M
