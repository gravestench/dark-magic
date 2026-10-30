-- Den progress consumes the population owner's complete room domain and retained
-- residents. Inactive is not dead, and an unsettled room has an unknown count.
-- Quest participation/reward live on the player; no party credit is inferred.
local ecs = require("engine.ecs/v1")
local state = require("engine.authority_state/v1")
local rules = require("d2legacy.policy.game_rules")
local M = {}

local function den_plan()
    local root = state.read("d2legacy.population.plan")
    if not root or not root.installed then return nil end
    for _, area in ipairs(root.zones) do
        if area.level_id == 8 then return area end
    end
end

local function population(entities)
    local plan = den_plan()
    if not plan or plan.difficulty ~= rules.difficulty() or #plan.rooms == 0 then return nil end
    local rooms, unsettled = {}, 0
    for _, room in ipairs(plan.rooms) do
        rooms[room.id] = true
        if not room.activated then unsettled = unsettled + 1 end
    end
    local kills = {}
    for _, entity in ipairs(entities) do
        local event = ecs.get(entity, "d2legacy.monster.death_event")
        if event and event:get("kind") == "monster_quest_kill" then
            kills[event:get("monster_id")] = true
        end
    end
    local living, dead, seen = 0, 0, {}
    for _, entity in ipairs(entities) do
        local resident = ecs.get(entity, "d2legacy.world.room_resident")
        local identity = ecs.get(entity, "d2legacy.monster.identity")
        if resident and identity and resident:get("level_id") == 8 then
            local id = resident:get("id")
            if id ~= identity:get("spawn_id") or seen[id] or not rooms[resident:get("room_id")] then return nil end
            seen[id] = true
            if ecs.get(entity, "d2legacy.monster.death") and kills[id] then
                dead = dead + 1
            else
                living = living + 1
            end
        end
    end
    -- A missing expected resident is not a death or a zero. The population
    -- owner retains both living entities and corpses across room deactivation.
    if living + dead ~= plan.created then return nil end
    return { remaining = living, unsettled = unsettled, clear = unsettled == 0 and living == 0 and dead > 0 }
end

local function participate(_, entities, structural)
    local counts = population(entities)
    if not counts or counts.clear then return end
    for _, entity in ipairs(entities) do
        local player = ecs.get(entity, "d2legacy.player.identity")
        local location = ecs.get(entity, "d2legacy.world.location")
        if player and location and location:get("level_id") == 8 and not ecs.get(entity, "d2legacy.quest.progress") then
            structural:set(entity, "d2legacy.quest.progress", {
                quest_id = 1, stage = 1, completed = false, rewarded = false,
                remaining = counts.remaining, unsettled_rooms = counts.unsettled, population_available = true,
            })
        end
    end
end

local function update(_, entities)
    local counts = population(entities)
    for _, entity in ipairs(entities) do
        local progress = ecs.get(entity, "d2legacy.quest.progress")
        if ecs.get(entity, "d2legacy.player.identity") and progress and progress:get("quest_id") == 1
            and not progress:get("completed") then
            progress:set("population_available", counts ~= nil)
            if counts then
                progress:set("remaining", counts.remaining)
                progress:set("unsettled_rooms", counts.unsettled)
                if counts.clear then
                    progress:set("completed", true)
                    progress:set("stage", #assert(require("d2legacy.quest_catalog/v1").quest(1)).stages)
                end
            end
        end
    end
end

function M.register()
    local read = { "d2legacy.player.identity", "d2legacy.world.location", "d2legacy.quest.progress",
        "d2legacy.world.room_resident", "d2legacy.monster.identity", "d2legacy.monster.death",
        "d2legacy.monster.death_event" }
    local query = { any = { "d2legacy.player.identity", "d2legacy.monster.identity", "d2legacy.monster.death_event" } }
    -- Entry is recorded after travel and before combat's death effects. Someone
    -- entering an already cleared Den never acquires a pending reward.
    ecs.system({ id = "d2legacy.quest.den.entry", phase = "collision", after = { "d2legacy.world.transition" },
        query = query, read = read, write = { "d2legacy.quest.progress" }, update = participate })
    ecs.system({ id = "d2legacy.quest.den", phase = "effects", after = { "d2legacy.monster.death" },
        query = query, read = read, write = { "d2legacy.quest.progress" }, update = update })
end

return M
