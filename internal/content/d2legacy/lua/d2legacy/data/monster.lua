-- Join the legacy MonStats/MonStats2/MonLvl spreadsheet split into one runtime
-- monster definition. Go decodes rows; this mod decides what their columns mean.

local records=require("engine.records/v1")
local M={}
local function number(row,key,fallback) local value=tonumber(row[key]); if value==nil then return fallback or 0 end return math.floor(value) end
local function truth(row,key) local value=string.lower(row[key] or ""); return value=="1" or value=="true" end
local function find(rows,key,wanted) for _,row in ipairs(rows) do if row[key]==wanted then return row end end end
local function choose(row,normal,nightmare,hell,difficulty) if difficulty==2 then return number(row,hell) elseif difficulty==1 then return number(row,nightmare) end return number(row,normal) end
local function scale(base,ratio) return math.floor(base*ratio/100) end
local component_names={"HD","TR","LG","RA","LA","RH","LH","SH","S1","S2","S3","S4","S5","S6","S7","S8"}
function M.load(id,difficulty)
 local stats=assert(find(records.load("data/global/excel/monstats.txt"),"Id",id),"missing MonStats row")
 assert(truth(stats,"enabled") and truth(stats,"isSpawn") and not truth(stats,"npc"),"monster is not an ordinary spawn")
 local graphics_id=stats.MonStatsEx~="" and stats.MonStatsEx or stats.Id
 local graphics=assert(find(records.load("data/global/excel/monstats2.txt"),"Id",graphics_id),"missing MonStats2 row")
 local level=find(records.load("data/global/excel/monlvl.txt"),"Level",stats.Level)
 local life_min=choose(stats,"minHP","minHP(N)","minHP(H)",difficulty); local life_max=choose(stats,"maxHP","maxHP(N)","maxHP(H)",difficulty)
 local defense=choose(stats,"AC","AC(N)","AC(H)",difficulty); local attack=choose(stats,"A1TH","A1TH(N)","A1TH(H)",difficulty)
 local damage_min=choose(stats,"A1MinD","A1MinD(N)","A1MinD(H)",difficulty); local damage_max=choose(stats,"A1MaxD","A1MaxD(N)","A1MaxD(H)",difficulty)
 local experience=choose(stats,"Exp","Exp(N)","Exp(H)",difficulty)
 if not truth(stats,"noRatio") and level then
  life_min=scale(choose(level,"HP","HP(N)","HP(H)",difficulty),life_min); life_max=scale(choose(level,"HP","HP(N)","HP(H)",difficulty),life_max)
  defense=scale(choose(level,"AC","AC(N)","AC(H)",difficulty),defense); attack=scale(choose(level,"TH","TH(N)","TH(H)",difficulty),attack)
  damage_min=scale(choose(level,"DM","DM(N)","DM(H)",difficulty),damage_min); damage_max=scale(choose(level,"DM","DM(N)","DM(H)",difficulty),damage_max)
  experience=scale(choose(level,"XP","XP(N)","XP(H)",difficulty),experience)
 end
 local components={} for _,name in ipairs(component_names) do if truth(graphics,name) and graphics[name.."v"] and graphics[name.."v"]~="" then components[name]=string.upper(graphics[name.."v"]) end end
 local size=math.max(number(graphics,"SizeX",2),number(graphics,"SizeY",2))/2
 return {id=stats.Id,base_id=stats.BaseId or "",graphics_id=graphics_id,name_key=stats.NameStr or stats.Id,ai=stats.AI or "",
  token=string.upper(stats.Code or ""),weapon_class=string.upper(graphics.BaseW or graphics.BaseWeapon or "HTH"),components=components,
  life_min=life_min*256,life_max=life_max*256,level=number(stats,"Level",1),defense=defense,attack_rating=attack,
  physical_min=damage_min*256,physical_max=damage_max*256,experience=experience,
  treasure_class=difficulty==2 and stats["TreasureClass1(H)"] or difficulty==1 and stats["TreasureClass1(N)"] or stats.TreasureClass1 or "",
  collider_radius=size,select_radius=size,velocity=number(stats,"Velocity",5),think_interval=math.max(number(stats,"aidel",1),1),
  aggro_radius=number(stats,"aidist",35)>0 and number(stats,"aidist",35) or 35,attack_range=math.max(number(graphics,"MeleeRng",1),1),
  min_group=math.max(number(stats,"MinGrp",1),1),max_group=math.max(number(stats,"MaxGrp",1),1),rarity=math.max(number(stats,"Rarity",1),1)}
end
function M.all(difficulty)
 local result={}
 for _,row in ipairs(records.load("data/global/excel/monstats.txt")) do
  local ok,value=pcall(M.load,row.Id,difficulty or 0)
  if ok and next(value.components) then
   local parts={} for key,component in pairs(value.components) do parts[#parts+1]=key.."="..component end
   table.sort(parts); value.components=table.concat(parts,","); result[#result+1]=value
  end
 end
 table.sort(result,function(left,right) return left.id<right.id end)
 return result
end
return M
