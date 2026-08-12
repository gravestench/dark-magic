-- Deterministically populate generated Diablo areas from decoded legacy rows.
-- The host supplies geometry candidates only; density, eligibility, monster
-- choice, group size, and definition interpretation are d2legacy policy.

local records=require("engine.records/v1")
local random=require("engine.authority_random/v1")
local commands=require("engine.authority_command/v1")
local monster=require("d2legacy.data.monster")
local spawn=require("d2legacy.commands.spawn_monster")
local M={}
local function find(rows,key,wanted) for _,row in ipairs(rows) do if row[key]==wanted then return row end end end
local function integer(row,key,fallback) return math.floor(tonumber(row[key]) or fallback or 0) end
local function candidates(level,difficulty)
 local prefix=difficulty==0 and "mon" or "nmon"; local count=math.max(integer(level,"NumMon",0),0); local out={}
 for index=1,math.min(count,25) do local id=level[prefix..index]; if id and id~="" then local ok,value=pcall(monster.load,id,difficulty); if ok then out[#out+1]=value end end end
 return out
end
local function weighted(values,total_roll) for _,value in ipairs(values) do if total_roll<value.rarity then return value end total_roll=total_roll-value.rarity end return values[#values] end
local function validate(command)
 local zone=command.payload
 assert(type(zone)=="table" and type(zone.level_id)=="number", "population zone is required")
 assert(type(zone.rooms)=="table", "population rooms are required")
end
local function apply(command)
 local zone=command.payload; if zone.level_id~=2 then return end
 local level=assert(find(records.load("data/global/excel/levels.txt"),"Id",tostring(zone.level_id)),"population level is missing")
 local density=integer(level,zone.difficulty==2 and "MonDen(H)" or zone.difficulty==1 and "MonDen(N)" or "MonDen",0)
 local values=candidates(level,zone.difficulty); if #values==0 then return end
 local total=0 for _,value in ipairs(values) do total=total+value.rarity end
 local created=0
 for _,room in ipairs(zone.rooms or {}) do
  local roll=random.integer("d2legacy.population.density",100000)
  if room.populate and (roll<density or created==0) then
   local definition=weighted(values,random.integer("d2legacy.population.family",total))
   local span=math.max(definition.max_group-definition.min_group+1,1); local count=definition.min_group+random.integer("d2legacy.population.group",span)
   for member=1,math.min(count,#room.points) do
    local point=room.points[member]; created=created+1
    spawn.materialize({tick=command.tick,payload={spawn_id="level:"..zone.level_id..":room:"..room.id..":monster:"..member,
      seed=random.integer("d2legacy.population.seed",2147483647),x=point.x,y=point.y,act=zone.act,level_id=zone.level_id,definition=definition}})
   end
  end
 end
end
function M.register()
 commands.register({kind="system.population.bootstrap",authorities={"system"},validate=validate,apply=apply})
end
return M
