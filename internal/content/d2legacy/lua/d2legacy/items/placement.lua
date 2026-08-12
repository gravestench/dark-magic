-- Diablo II item footprint, body-slot, belt, held-item, and swap policy.

local ecs=require("engine.ecs/v1")
local M={}
local grids={inventory=true,stash=true,cube=true}
local slots={equipment=true,hireling=true,belt=true,quest_service=true}

local function has_token(csv,wanted)
    for token in string.gmatch(csv or "","[^,]+") do if token==wanted then return true end end
    return false
end
local function overlap(ax,ay,aw,ah,bx,by,bw,bh)
    return ax<bx+bw and bx<ax+aw and ay<by+bh and by<ay+ah
end
function M.grid_size(layout,container)
    if not grids[container] then return end
    return layout:get(container.."_width"),layout:get(container.."_height")
end
function M.overlaps(item_entity,item,destination,entities)
    local result={}
    for _,entity in ipairs(entities) do
        if entity:id()~=item_entity:id() then
            local other,placed=ecs.get(entity,"d2legacy.item.identity"),ecs.get(entity,"d2legacy.item.placement")
            if other and placed and other:get("owner"):id()==item:get("owner"):id()
                and placed:get("container")==destination.container
                and overlap(destination.x,destination.y,item:get("width"),item:get("height"),
                    placed:get("x"),placed:get("y"),other:get("width"),other:get("height")) then result[#result+1]=entity end
        end
    end
    return result
end
function M.slot_occupant(item_entity,item,destination,entities)
    for _,entity in ipairs(entities) do
        if entity:id()~=item_entity:id() then
            local other,placed=ecs.get(entity,"d2legacy.item.identity"),ecs.get(entity,"d2legacy.item.placement")
            if other and placed and other:get("owner"):id()==item:get("owner"):id()
                and placed:get("container")==destination.container then
                if destination.container=="belt" and placed:get("belt_slot")==destination.belt_slot then return entity end
                if destination.container~="belt" and placed:get("slot")==destination.slot
                    and (destination.container~="equipment" or destination.slot~="rarm" and destination.slot~="larm"
                        or placed:get("weapon_set")==destination.weapon_set) then return entity end
            end
        end
    end
end
function M.validate(layout,item_entity,item,d,entities,allow_one_overlap)
    assert(type(d)=="table" and type(d.container)=="string","item destination is required")
    d.x,d.y,d.slot,d.belt_slot,d.weapon_set,d.page=d.x or 0,d.y or 0,d.slot or "",d.belt_slot or 0,d.weapon_set or 0,d.page or 0
    local width,height=M.grid_size(layout,d.container)
    if width then
        assert(d.x>=0 and d.y>=0 and d.x+item:get("width")<=width and d.y+item:get("height")<=height,"item does not fit grid")
        local overlaps=M.overlaps(item_entity,item,d,entities)
        assert(#overlaps==0 or allow_one_overlap and #overlaps==1,"item footprint is occupied")
        return overlaps[1]
    end
    if d.container=="equipment" or d.container=="hireling" then
        assert(d.slot~="" and has_token(item:get("body_slots"),d.slot),"item cannot use body slot")
        if d.container=="equipment" and (d.slot=="rarm" or d.slot=="larm") then assert(d.weapon_set==0 or d.weapon_set==1,"invalid weapon set")
        else assert(d.weapon_set==0,"shared body slot cannot use weapon set") end
    elseif d.container=="belt" then
        assert(item:get("belt_eligible") and d.belt_slot>=0 and d.belt_slot<layout:get("belt_capacity"),"item cannot use belt slot")
    elseif d.container=="quest_service" then assert(d.slot~="","service slot is required")
    elseif d.container~="world" and d.container~="held" then error("unsupported item container") end
    local occupant=M.slot_occupant(item_entity,item,d,entities)
    assert(not occupant or allow_one_overlap,"item slot is occupied")
    return occupant
end
function M.set(component,d)
    for _,field in ipairs({"container","x","y","slot","belt_slot","weapon_set","page"}) do component:set(field,d[field]) end
end
M.is_grid=function(container) return grids[container] or false end
M.is_held_destination=function(container) return grids[container] or slots[container] or false end
return M
