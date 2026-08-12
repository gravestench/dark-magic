-- Resolve TreasureClassEx recursively into terminal item codes.
--
-- Negative Picks means authored guaranteed copies. Positive Picks performs a
-- weighted draw including NoDrop. Recursion is bounded and cycle checked so a
-- malformed mod record fails the authoritative tick instead of hanging it.

local records=require("engine.records/v1")
local random=require("engine.authority_random/v1")
local M={}
local MAX_DEPTH=64
local catalog
local function integer(row,key,default) return math.floor(tonumber(row[key]) or default or 0) end
local function load()
 if catalog then return catalog end
 catalog={}
 for _,row in ipairs(records.load("data/global/excel/treasureclassex.txt")) do
  local name=row["Treasure Class"]
  if name and name~="" then
   assert(not catalog[name],"duplicate treasure class "..name)
   local value={name=name,picks=integer(row,"Picks"),no_drop=integer(row,"NoDrop"),entries={},
    quality={unique=integer(row,"Unique"),set=integer(row,"Set"),rare=integer(row,"Rare"),magic=integer(row,"Magic")}}
   for index=1,10 do local code=row["Item"..index];if code and code~="" then value.entries[#value.entries+1]={code=code,weight=integer(row,"Prob"..index)} end end
   catalog[name]=value
  end
 end
 return catalog
end
local function copy(values) local result={} for index,value in ipairs(values) do result[index]=value end return result end
local function copy_quality(value) return {unique=value.unique,set=value.set,rare=value.rare,magic=value.magic} end
local function expand(name,path,active,drops)
 local values=load();local class=assert(values[name],"unknown treasure class "..tostring(name))
 assert(#path<MAX_DEPTH,"maximum treasure-class depth exceeded")
 assert(not active[name],"treasure-class cycle at "..name)
 assert(class.picks~=0,"treasure class has zero Picks: "..name)
 active[name]=true;path=copy(path);path[#path+1]=name
 local chosen={}
 if class.picks<0 then
  local left=-class.picks
  for _,entry in ipairs(class.entries) do for _=1,entry.weight do if left<=0 then break end;chosen[#chosen+1]=entry;left=left-1 end end
 else
  local total=class.no_drop;for _,entry in ipairs(class.entries) do assert(entry.weight>=0,"negative treasure weight");total=total+entry.weight end
  assert(total>0,"treasure class has no weighted outcomes")
  for _=1,class.picks do local roll=random.integer("d2legacy.loot.treasure_class",total)
   if roll>=class.no_drop then roll=roll-class.no_drop;for _,entry in ipairs(class.entries) do if roll<entry.weight then chosen[#chosen+1]=entry;break end;roll=roll-entry.weight end end
  end
 end
 for _,entry in ipairs(chosen) do
  if values[entry.code] then expand(entry.code,path,active,drops)
  else drops[#drops+1]={code=entry.code,path=copy(path),quality=copy_quality(class.quality)} end
 end
 active[name]=nil
end
function M.roll(name) if not name or name=="" then return {} end;local drops={};expand(name,{}, {},drops);return drops end
return M
