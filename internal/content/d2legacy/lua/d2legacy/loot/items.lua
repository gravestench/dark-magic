-- Join a terminal treasure code to its immutable base-item record.

local records=require("engine.records/v1")
local M={}
local tables={{"weapons.txt","weapon"},{"armor.txt","armor"},{"misc.txt","misc"}}
local function integer(row,key,default) return math.floor(tonumber(row[key]) or default or 0) end
function M.base(code)
 for _,source in ipairs(tables) do for _,row in ipairs(records.load("data/global/excel/"..source[1])) do if row.code==code then
  return {code=code,kind=source[2],name_key=row.namestr or "",type=row.type or "",type2=row.type2 or "",
   level=integer(row,"level"),level_requirement=integer(row,"levelreq"),magic_level=integer(row,"magic lvl"),
   inventory_file=row.invfile or "",world_file=row.flippyfile or "",width=integer(row,"invwidth",1),height=integer(row,"invheight",1),
   base_cost=integer(row,"cost"),uber=(row.ubercode or "")~="",class_specific=(row.type or "")=="class"}
 end end end
 return nil
end
return M
