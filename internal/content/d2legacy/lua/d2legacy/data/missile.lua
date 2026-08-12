-- Interpret Missiles.txt presentation columns inside the d2legacy mod.
-- The engine exposes immutable decoded rows; paths, rates, and looping are
-- Diablo presentation policy and therefore stay here beside the consuming UI.

local records=require("engine.records/v1")
local M={}
local function integer(row,key,fallback) return math.floor(tonumber(row[key]) or fallback or 0) end
function M.all()
 local result={}
 for _,row in ipairs(records.load("data/global/excel/missiles.txt")) do
  local id,cel=row.Missile or row.missile,row.CelFile or row.celfile
  if id and id~="" and cel and cel~="" then
   local speed=integer(row,"AnimSpeed",integer(row,"animspeed",16))
   result[#result+1]={id=id,dcc="data/global/missiles/"..cel..".dcc",
    palette="data/global/palette/units/pal.dat",directions=math.max(integer(row,"NumDirections",integer(row,"numdirections",1)),1),
    frames_per_second=math.max(math.floor(speed*25/16),1),loop=integer(row,"LoopAnim",integer(row,"loopanim",0))~=0,
    travel_sound=row.TravelSound or row.travelsound or "",hit_sound=row.HitSound or row.hitsound or "",
    offset_x=integer(row,"Xoffset",integer(row,"xoffset",0)),offset_y=integer(row,"Yoffset",integer(row,"yoffset",0)),
    offset_z=integer(row,"Zoffset",integer(row,"zoffset",0))}
  end
 end
 table.sort(result,function(left,right) return left.id<right.id end)
 return result
end
return M
