-- Commit decoded quest completion, NPC rewards, and difficulty progression.

local authority=require("engine.authority_command/v1")
local ecs=require("engine.ecs/v1")
local M={}

local function quest(id)
    return require("d2legacy.quest_catalog/v1").quest(id)
end

local function player_for(id)
    for _,entity in ipairs(ecs.query({all={"d2legacy.player.identity"}})) do
        if ecs.get(entity,"d2legacy.player.identity"):get("player")==id then return entity end
    end
    error("quest player does not exist")
end

local function active_npc(player)
    for _,entity in ipairs(ecs.query({all={"d2legacy.interaction.context"}})) do
        local context=ecs.get(entity,"d2legacy.interaction.context")
        if context:get("owner")==player then
            local target=ecs.get(context:get("target"),"d2legacy.interaction.target")
            return target and string.lower(target:get("npc")) or ""
        end
    end
    return ""
end

function M.complete(command)
    local quest=assert(quest(command.payload.quest_id),"unknown recovered quest")
    local player=player_for(command.payload.player)
    local stage=#quest.stages
    ecs.set(player,"d2legacy.quest.progress",{quest_id=quest.id,stage=stage,
        completed=true,rewarded=false})
end

function M.reward(command)
    local player=player_for(command.player)
    local progress=assert(ecs.get(player,"d2legacy.quest.progress"),"quest is not complete")
    local quest=assert(quest(progress:get("quest_id")),"unknown recovered quest")
    assert(progress:get("completed") and not progress:get("rewarded"),"quest reward is unavailable")
    local npc=active_npc(command.player)
    assert(quest.id==1 and npc=="akara","quest reward requires its recovered NPC relationship")
    local player_progress=ecs.get(player,"d2legacy.player.progress")
    player_progress:set("unspent_skill_points",player_progress:get("unspent_skill_points")+1)
    progress:set("rewarded",true)
    ecs.create({["d2legacy.quest.event"]={kind="quest_rewarded",tick=command.tick,
        player=command.player,quest_id=quest.id,npc=npc,reward="skill_point:1"}})
end

function M.advance_difficulty(command)
    local player=player_for(command.payload.player)
    local difficulty=ecs.get(player,"d2legacy.player.difficulty")
    local wanted=command.payload.difficulty
    assert(command.payload.completed_act==5,"difficulty progression requires completed Act V")
    assert(wanted==difficulty:get("current")+1 and wanted<=2,"invalid next difficulty")
    difficulty:set("highest_completed",difficulty:get("current"))
    difficulty:set("current",wanted)
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
