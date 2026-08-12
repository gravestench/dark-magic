-- Complete first authoritative loot-generation slice.
-- Treasure traversal, base-item resolution, and quality policy stay separate
-- so later affix/property modules can grow without turning this into a wall.

local treasure=require("d2legacy.loot.treasure_class")
local quality=require("d2legacy.loot.quality")
local items=require("d2legacy.loot.items")
local affixes=require("d2legacy.loot.affixes")
local properties=require("d2legacy.loot.properties")
local M={}
local function encode(value)
 if type(value)=="string" then return string.format("%q",value) end
 if type(value)=="number" or type(value)=="boolean" then return tostring(value) end
 if type(value)~="table" then return "null" end
 if #value>0 then local parts={} for _,entry in ipairs(value) do parts[#parts+1]=encode(entry) end;return "["..table.concat(parts,",").."]" end
 local keys={} for key in pairs(value) do keys[#keys+1]=key end;table.sort(keys,function(a,b)return tostring(a)<tostring(b)end)
 local parts={} for _,key in ipairs(keys) do parts[#parts+1]=string.format("%q:%s",tostring(key),encode(value[key])) end
 return "{"..table.concat(parts,",").."}"
end
function M.roll(treasure_class,context)
 local result={}
 for _,drop in ipairs(treasure.roll(treasure_class)) do local base=items.base(drop.code)
  if base then local item_quality=quality.roll(base,context,drop.quality);local prefixes,suffixes=affixes.roll(base,item_quality,context.monster_level,context.version);local stats,effects,unsupported=properties.apply(prefixes,suffixes)
   result[#result+1]={code=drop.code,quality=item_quality,level=context.monster_level,path=table.concat(drop.path," > "),inventory_file=base.inventory_file,world_file=base.world_file,width=base.width,height=base.height,base_cost=base.base_cost,prefixes=prefixes,suffixes=suffixes,stats=stats,effects=effects,unsupported=unsupported}
  else result[#result+1]={code=drop.code,quality="unresolved",level=context.monster_level,path=table.concat(drop.path," > ")} end
 end
 return result
end
M.encode=encode
return M
