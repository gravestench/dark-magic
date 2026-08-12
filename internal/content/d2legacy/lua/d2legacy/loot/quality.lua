-- Ordered ItemRatio quality checks: unique, set, rare, magic, superior,
-- normal, then low. All arithmetic is integer and consumes a named authority
-- RNG stream, so replay never depends on Lua or host-global random state.

local records=require("engine.records/v1")
local random=require("engine.authority_random/v1")
local M={}
local SCALE=128
local function integer(row,key,default) return math.floor(tonumber(row[key]) or default or 0) end
local function ratio(version,uber,class_specific)
 for _,row in ipairs(records.load("data/global/excel/itemratio.txt")) do
  if integer(row,"Version")==version and (integer(row,"Uber")~=0)==uber and (integer(row,"Class Specific")~=0)==class_specific then return row end
 end
end
local function denominator(row,prefix,monster_level,item_level,magic_find,modifier,diminishing,use_magic_find)
 local base=integer(row,prefix);assert(base>0,"invalid ItemRatio "..prefix)
 local divisor=integer(row,prefix.."Divisor");local value=(base-(divisor>0 and math.floor((monster_level-item_level)/divisor) or 0))*SCALE
 value=math.max(value,SCALE)
 if use_magic_find and magic_find>0 then local bonus=magic_find;if diminishing>0 then bonus=math.floor(magic_find*diminishing/(magic_find+diminishing)) end;value=math.floor(value*100/(100+bonus)) end
 local minimum=integer(row,prefix.."Min");if minimum>0 then value=math.max(value,minimum) end
 value=value-math.floor(value*(modifier or 0)/1024)
 return math.max(value,SCALE)
end
function M.roll(item,context,modifier)
 modifier=modifier or {unique=0,set=0,rare=0,magic=0}
 local row=assert(ratio(context.version or 100,item.uber or false,item.class_specific or false),"missing ItemRatio row")
 local checks={{"unique","Unique",250,true},{"set","Set",500,true},{"rare","Rare",600,true},{"magic","Magic",0,true},{"superior","HiQuality",0,false},{"normal","Normal",0,false}}
 for _,check in ipairs(checks) do local value=denominator(row,check[2],context.monster_level,item.level,context.magic_find or 0,modifier[check[1]] or 0,check[3],check[4]);if random.integer("d2legacy.loot.quality",value)<SCALE then return check[1] end end
 return "low"
end
return M
