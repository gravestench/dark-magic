-- Interpret rolled property codes into portable stat/effect facts.
-- Unsupported legacy functions remain explicit instead of being silently
-- flattened into the wrong stat operation.

local records=require("engine.records/v1")
local M={}
local catalog
local direct={[1]=true,[2]=true,[8]=true,[13]=true,[14]=true,[15]=true,[16]=true,[17]=true,[22]=true}
local function integer(row,key,default) return math.floor(tonumber(row[key]) or default or 0) end
local function load()
 if catalog then return catalog end;catalog={}
 for _,row in ipairs(records.load("data/global/excel/properties.txt")) do if row.code and row.code~="" then local steps={};for i=1,7 do local fn=integer(row,"func"..i);local stat=row["stat"..i] or "";if fn~=0 or stat~="" then steps[#steps+1]={fn=fn,stat=stat,set=integer(row,"set"..i)~=0,value=integer(row,"val"..i)} end end;catalog[row.code]=steps end end
 return catalog
end
function M.apply(prefixes,suffixes)
 local stats,effects,unsupported={},{},{}
 for _,affix in ipairs({prefixes,suffixes}) do for _,value in ipairs(affix) do for _,modifier in ipairs(value.modifiers) do
  local previous=0;for _,step in ipairs(assert(load()[modifier.code],"unknown property "..modifier.code)) do local fn=step.fn;if fn==3 or fn==9 then fn=previous elseif fn~=0 then previous=fn end
   if direct[fn] then stats[#stats+1]={code=step.stat,parameter=modifier.parameter,value=modifier.value,set=step.set,fn=fn}
   elseif fn==5 then effects[#effects+1]={kind="minimum_damage",value=modifier.value}
   elseif fn==6 then effects[#effects+1]={kind="maximum_damage",value=modifier.value}
   elseif fn==7 then effects[#effects+1]={kind="damage_percent",value=modifier.value}
   elseif fn==10 then effects[#effects+1]={kind="skill_tab",value=modifier.value,class=math.floor(modifier.parameter/3),tab=modifier.parameter%3}
   elseif fn==11 then effects[#effects+1]={kind="proc",skill_id=modifier.parameter,chance=modifier.minimum,level=modifier.maximum}
   elseif fn==19 then effects[#effects+1]={kind="charged_skill",skill_id=modifier.parameter,charges=modifier.minimum,level=modifier.maximum}
   elseif fn==20 then effects[#effects+1]={kind="indestructible",value=modifier.value}
   elseif fn==23 then effects[#effects+1]={kind="ethereal",value=modifier.value}
   elseif fn~=0 then unsupported[#unsupported+1]={property=modifier.code,fn=fn,stat=step.stat,parameter=modifier.parameter,value=modifier.value} end
  end
 end end end
 table.sort(stats,function(a,b)return a.code~=b.code and a.code<b.code or a.parameter<b.parameter end)
 return stats,effects,unsupported
end
return M
