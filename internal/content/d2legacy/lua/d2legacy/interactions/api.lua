local requests=require("engine.interaction/v1")
local ecs=require("engine.ecs/v1")
local M={api=1,open=requests.open,open_at=requests.open_at,close=requests.close}
local function split(value) local out={} for token in string.gmatch(value or "","[^,]+") do out[#out+1]=token end return out end
function M.snapshot()
 for _,entity in ipairs(ecs.query({all={"d2legacy.interaction.context"}})) do
  local context=ecs.get(entity,"d2legacy.interaction.context"); if context:get("owner")=="local-player" then
   local target=context:get("target"); local value=target and ecs.get(target,"d2legacy.interaction.target") or nil
   if not value then return {active=false,categories={},services={}} end
   return {active=true,target_id=value:get("id"),npc=value:get("npc"),vendor=value:get("vendor"),categories=split(value:get("categories")),services=split(value:get("services"))}
  end
 end
 return {active=false,categories={},services={}}
end
return M
