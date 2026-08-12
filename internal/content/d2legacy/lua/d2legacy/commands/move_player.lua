-- Apply one admitted keyboard or point-and-click movement sample.
-- Walk/run speeds, Diablo animation modes, facing, and action replacement are
-- mod policy. The Go input adapter only normalizes native input and path
-- waypoints into this transport-neutral command.

local commands=require("engine.authority_command/v1")
local ecs=require("engine.ecs/v1")
local M={}
local function finite(v) return type(v)=="number" and v==v and v~=math.huge and v~=-math.huge end
local function player(name)
 for _,e in ipairs(ecs.query({all={"d2legacy.world.player_control"}})) do
  if ecs.get(e,"d2legacy.world.player_control"):get("player")==name then return e end
 end
end
local function direction(x,y)
 local sx=x<0 and -1 or x>0 and 1 or 0; local sy=y<0 and -1 or y>0 and 1 or 0
 return ({["0,1"]=0,["-1,0"]=1,["0,-1"]=2,["1,0"]=3,["1,1"]=4,["-1,1"]=5,["-1,-1"]=6,["1,-1"]=7})[sx..","..sy] or 0
end
local function validate(c)
 local p=c.payload; assert(type(p)=="table","movement payload is required")
 assert(type(p.x)=="number" and p.x>=-1 and p.x<=1 and type(p.y)=="number" and p.y>=-1 and p.y<=1,"movement axes must be normalized")
 assert(type(p.running)=="boolean","movement mode is required")
 if p.target then assert(finite(p.target.x) and finite(p.target.y),"movement target must be finite") end
end
local function apply(c)
 local entity=player(c.player); if not entity then return end
 local p,x,y=c.payload,c.payload.x,c.payload.y
 local explicit=p.target~=nil or x~=0 or y~=0
 local approach=ecs.get(entity,"d2legacy.combat.attack_approach"); local attack=ecs.get(entity,"d2legacy.combat.attack_animation")
 if explicit then ecs.remove(entity,"d2legacy.combat.attack_approach");ecs.remove(entity,"d2legacy.combat.attack_animation")
 elseif approach or attack then return end
 if p.target then local position=ecs.get(entity,"d2legacy.world.position");x,y=p.target.x-position:get("x"),p.target.y-position:get("y");local length=math.sqrt(x*x+y*y);if length<=0.2 then x,y=0,0 else x,y=x/length,y/length end
 elseif x~=0 and y~=0 then x,y=x*0.7071067811865476,y*0.7071067811865476 end
 local speed=p.running and 15 or 10;local velocity=ecs.get(entity,"d2legacy.world.velocity");velocity:set("x",x*speed);velocity:set("y",y*speed)
 local mode=ecs.get(entity,"d2legacy.player.movement_mode");if mode then mode:set("running",p.running) end
 local animation=ecs.get(entity,"d2legacy.player.animation");if animation then local moving=x~=0 or y~=0;animation:set("mode",moving and (p.running and "RN" or "WL") or "NU");if moving then animation:set("direction",direction(x,y)) end end
end
function M.register() commands.register({kind="player.move",authorities={"player"},validate=validate,apply=apply}) end
return M
