-- Commit decoded quest completion, NPC rewards, and difficulty progression.

local authority=require("engine.authority_command/v1")
local ecs=require("engine.ecs/v1")
local game_rules=require("d2legacy.policy.game_rules")
local M={}

local function quest(id)
    return require("d2legacy.quest_catalog/v1").quest(id)
end

local function player_for(id)
    for _,entity in ipairs(ecs.query({all={"d2legacy.player.identity"}})) do
        if ecs.get(entity,"d2legacy.player.identity"):get("player")==id then return entity end
    end
    return nil
end

function M.complete(command)
    local quest=assert(quest(command.payload.quest_id),"unknown recovered quest")
    local player=assert(player_for(command.payload.player),"quest player does not exist")
    local progress=ecs.get(player,"d2legacy.quest.progress")
    if progress and progress:get("quest_id")==quest.id and progress:get("completed") then return end
    local stage=#quest.stages
    ecs.set(player,"d2legacy.quest.progress",{quest_id=quest.id,stage=stage,
        completed=true,rewarded=false})
end

function M.reward(command)
    local player=player_for(command.player)
    if not player then return false end
    local progress=ecs.get(player,"d2legacy.quest.progress")
    if not progress or progress:get("quest_id")~=1 or not progress:get("completed") or progress:get("rewarded") then return false end
    local context,target=require("d2legacy.interactions.context").active(command.player)
    local target_location=context and (ecs.get(context:get("target"),"d2legacy.world.location")
        or ecs.get(context:get("target"),"d2legacy.world.room_resident"))
    local location=ecs.get(player,"d2legacy.world.location")
    local vitals=ecs.get(player,"d2legacy.player.vitals")
    if not context or string.lower(target:get("npc"))~="akara" or not target_location
        or target_location:get("level_id")~=1 or not location
        or location:get("act")~=1 or location:get("level_id")~=1
        or not vitals or vitals:get("health")<=0 or ecs.get(player,"d2legacy.player.death") then return false end
    local npc="akara"
    local player_progress=ecs.get(player,"d2legacy.player.progress")
    player_progress:set("unspent_skill_points",player_progress:get("unspent_skill_points")+1)
    progress:set("rewarded",true)
    ecs.create({["d2legacy.quest.event"]={kind="quest_rewarded",tick=command.tick,
        player=command.player,quest_id=1,npc=npc,reward="skill_point:1"}})
    return true
end

function M.claim_available(owner,tick)
    return M.reward({player=owner,tick=tick})
end

-- Completing Act V unlocks the next difficulty for the character only; the
-- active game keeps the difficulty fixed by its immutable rules.
function M.advance_difficulty(command)
    local player=assert(player_for(command.payload.player),"quest player does not exist")
    local difficulty=ecs.get(player,"d2legacy.player.difficulty")
    local completed=game_rules.difficulty()
    local wanted=command.payload.difficulty
    assert(command.payload.completed_act==5,"difficulty progression requires completed Act V")
    assert(wanted==completed+1 and wanted<=2,"invalid next difficulty")
    difficulty:set("highest_completed",math.max(difficulty:get("highest_completed"),completed))
end

function M.register()
    authority.register({kind="system.quest.complete",authorities={"system","administrator"},
        validate=function(command) assert(type(command.payload.quest_id)=="number" and type(command.payload.player)=="string") end,
        apply=M.complete})
    authority.register({kind="quest.claim_reward",authorities={"player"},
        validate=function() end,apply=M.reward})
    authority.register({kind="system.difficulty.advance",authorities={"system","administrator"},
        validate=function(command) assert(type(command.payload.difficulty)=="number" and type(command.payload.player)=="string") end,
        apply=M.advance_difficulty})
end

return M
