-- Select and materialize MagicPrefix/MagicSuffix records.
-- Eligibility, frequency, group exclusion, and property ranges are Diablo item
-- policy. Named streams make selection and value rolls replayable.

local records=require("engine.records/v1")
local random=require("engine.authority_random/v1")
local M={}
local function integer(row,key,default) return math.floor(tonumber(row[key]) or default or 0) end
local function truth(row,key) local v=string.lower(row[key] or "");return v=="1" or v=="true" end
local function contains(values,wanted) for _,value in ipairs(values) do if value.name==wanted then return true end end end
local function types(row,prefix) local out={} for i=1,7 do local value=row[prefix..i];if value and value~="" then out[#out+1]=value end end return out end
local function item_matches(item,wanted) return item.type==wanted or item.type2==wanted end
local function eligible(row,item,options,used)
 if not truth(row,"spawnable") or integer(row,"frequency")<=0 or integer(row,"level")>options.level then return false end
 local maximum=integer(row,"maxlevel");if maximum>0 and options.level>maximum then return false end
 if options.version<100 and integer(row,"version")>=100 or options.quality=="rare" and not truth(row,"rare") then return false end
 local group=integer(row,"group");if group~=0 and used[group] then return false end
 for _,value in ipairs(types(row,"etype")) do if item_matches(item,value) then return false end end
 for _,value in ipairs(types(row,"itype")) do if item_matches(item,value) then return true end end
 return false
end
local function modifier(row,index)
 local code=row["mod"..index.."code"];if not code or code=="" then return end
 local low,high=integer(row,"mod"..index.."min"),integer(row,"mod"..index.."max");if low>high then low,high=high,low end
 return {code=code,parameter=integer(row,"mod"..index.."param"),minimum=low,maximum=high,
  value=low+random.integer("d2legacy.loot.affix_value",high-low+1)}
end
local function choose(table_name,kind,item,options,limit,used,selected,total_selected)
 for _=1,limit do
  if total_selected[1]>=options.max_total then break end
  if random.integer("d2legacy.loot.affix_slot",2)==1 then
   local candidates,total={},0
   for _,row in ipairs(records.load("data/global/excel/"..table_name)) do
    if eligible(row,item,options,used) and not contains(selected,row.Name) then candidates[#candidates+1]=row;total=total+integer(row,"frequency") end
   end
   table.sort(candidates,function(a,b)return a.Name<b.Name end)
   if total>0 then local roll=random.integer("d2legacy.loot.affix_choice",total)
    for _,row in ipairs(candidates) do local weight=integer(row,"frequency");if roll<weight then
     local value={name=row.Name,kind=kind,group=integer(row,"group"),level_requirement=integer(row,"levelreq"),modifiers={}}
     for index=1,3 do local m=modifier(row,index);if m then value.modifiers[#value.modifiers+1]=m end end
     selected[#selected+1]=value;total_selected[1]=total_selected[1]+1;if value.group~=0 then used[value.group]=true end;break
    end;roll=roll-weight end
   end
  end
 end
end
function M.roll(item,quality,level,version)
 if quality~="magic" and quality~="rare" then return {},{} end
 local options={quality=quality,level=level,version=version or 100,max_total=quality=="rare" and 6 or 2}
 local prefixes,suffixes,used,total={},{},{},{0}
 choose("magicprefix.txt","prefix",item,options,quality=="rare" and 3 or 1,used,prefixes,total)
 choose("magicsuffix.txt","suffix",item,options,quality=="rare" and 3 or 1,used,suffixes,total)
 return prefixes,suffixes
end
return M
