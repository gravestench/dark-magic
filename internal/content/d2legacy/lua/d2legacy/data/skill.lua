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
return M
