-- Read immutable Skills/SkillDesc rows and turn them into d2legacy UI facts.
-- Icon-sheet selection and left-button eligibility are Diablo conventions,
-- not generic engine behavior.

local records=require("engine.records/v1")
local M={}
local sheets={ama="AmSkillicon",sor="SoSkillicon",nec="NeSkillicon",pal="PaSkillicon",bar="BaSkillicon",dru="DrSkillicon",ass="AsSkillicon"}
local function find(rows,key,wanted) for _,row in ipairs(rows) do if row[key]==wanted then return row end end end
function M.load(id)
 local skill=find(records.load("data/global/excel/skills.txt"),"Id",tostring(id)); if not skill then return nil end
 local description=find(records.load("data/global/excel/skilldesc.txt"),"skilldesc",skill.skilldesc); if not description then return nil end
 local icon=tonumber(description.IconCel); if not icon or icon<0 then return nil end
 local sheet=sheets[string.lower(skill.charclass or "")] or "Skillicon"
 return {id=id,icon=math.floor(icon),sheet="data/global/ui/SPELLS/"..sheet..".DC6",
  name_key=description["str name"] or "",short_key=description["str short"] or "",
  list_row=math.floor(tonumber(description.ListRow) or 0),left_allowed=skill.leftskill=="1",passive=skill.passive=="1"}
end

-- starting_for_class interprets CharStats, Skills, and SkillDesc as one
-- authoritative initial loadout. Go only imports durable character facts; it
-- does not decide which D2 skills a newly admitted class knows.
function M.starting_for_class(class)
 local wanted=string.lower(assert(class,"class is required"))
 local start=""
 for _,row in ipairs(records.load("data/global/excel/charstats.txt")) do
  if string.lower(row.class or "")==wanted then start=row.StartSkill or ""; break end
 end
 local descriptions={}
 for _,row in ipairs(records.load("data/global/excel/skilldesc.txt")) do
  if row.skilldesc and row.skilldesc~="" then descriptions[row.skilldesc]=row end
 end
 local result={}
 for _,skill in ipairs(records.load("data/global/excel/skills.txt")) do
  if skill.general=="1" or string.lower(skill.skill or "")==string.lower(start) then
   local description=descriptions[skill.skilldesc]
   local id,row=tonumber(skill.Id),description and tonumber(description.ListRow)
   if id and row and row>=0 and skill.passive~="1" then
    result[#result+1]={id=math.floor(id),level=1,list_row=math.floor(row),
     left_allowed=skill.leftskill=="1",right_allowed=true}
   end
  end
 end
 table.sort(result,function(a,b) return a.id<b.id end)
 return result
end
return M
