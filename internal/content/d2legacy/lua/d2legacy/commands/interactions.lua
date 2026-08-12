local commands=require("engine.authority_command/v1")
local ecs=require("engine.ecs/v1")
local M={}
local function owner(c) local p=c.payload.owner; if c.authority=="player" then assert(not p or p=="" or p==c.player,"cannot change another owner") end return p and p~="" and p or c.player end
local function context(name) for _,e in ipairs(ecs.query({all={"d2legacy.interaction.context"}})) do local v=ecs.get(e,"d2legacy.interaction.context"); if v:get("owner")==name then return v end end error("unknown interaction owner") end
local function target_by_id(id) for _,e in ipairs(ecs.query({all={"d2legacy.interaction.target"}})) do local v=ecs.get(e,"d2legacy.interaction.target"); if v:get("id")==string.lower(id) then return e,v end end end
local function player(name) for _,e in ipairs(ecs.query({all={"d2legacy.player.identity","d2legacy.world.position"}})) do if ecs.get(e,"d2legacy.player.identity"):get("player")==name then return ecs.get(e,"d2legacy.world.position") end end end
local function in_range(name,target) local p=player(name); if not p then return true end local dx,dy=p:get("x")-target:get("x"),p:get("y")-target:get("y"); return dx*dx+dy*dy<=target:get("radius")*target:get("radius") end
function M.open(c) local p=c.payload; local entity,target
 if p.at then local best,bestd for _,e in ipairs(ecs.query({all={"d2legacy.interaction.target"}})) do local v=ecs.get(e,"d2legacy.interaction.target"); local dx,dy=p.x-v:get("x"),p.y-v:get("y"); local d=dx*dx+dy*dy; if d<=2.25 and (not bestd or d<bestd) then entity,target,bestd=e,v,d end end
 else entity,target=target_by_id(assert(p.target,"target is required")) end
 assert(entity and in_range(owner(c),target),"interaction target is unavailable or out of range"); context(owner(c)):set("target",entity)
end
function M.close(c) context(owner(c)):set("target",ecs.create()) end
function M.register()
 commands.register({kind="interaction.open",authorities={"player","administrator"},validate=function(c) owner(c) end,apply=M.open})
 commands.register({kind="interaction.close",authorities={"player","administrator"},validate=function(c) owner(c) end,apply=M.close})
end
return M
