-- Authoritative item movement and weapon-set selection commands.

local commands=require("engine.authority_command/v1")
local ecs=require("engine.ecs/v1")
local find=require("d2legacy.items.find")
local placement=require("d2legacy.items.placement")
local M={}
local PRICE_SCALE=1024

local function owner(command)
    local requested=command.payload.owner
    if command.authority=="player" then assert(requested==nil or requested=="" or requested==command.player,"cannot move another owner's items") end
    return requested and requested~="" and requested or command.player
end
function M.validate_move(command)
    local p=command.payload
    assert(type(p)=="table" and type(p.item_id)=="string" and p.item_id~="","item ID is required")
    assert(type(p.destination)=="table" and type(p.destination.container)=="string","item destination is required")
    assert(p.destination.container~="vendor","vendor stock requires a transaction")
    if p.place_held then assert(placement.is_held_destination(p.destination.container),"invalid held-item destination") end
    owner(command)
end
function M.apply_move(command)
    local p=command.payload
    local entities=ecs.query({any={"d2legacy.items.layout","d2legacy.item.identity"}})
    local layout_entity,layout=assert(find.layout(owner(command),entities))
    local item_entity,item=assert(find.item(layout_entity,p.item_id,entities))
    local current=assert(ecs.get(item_entity,"d2legacy.item.placement"))
    if p.place_held then assert(current:get("container")=="held","item is not held") end
    local displaced=placement.validate(layout,item_entity,item,p.destination,entities,p.place_held)
    placement.set(current,p.destination)
    if displaced then placement.set(ecs.get(displaced,"d2legacy.item.placement"),{container="held",x=0,y=0,slot="",belt_slot=0,weapon_set=0,page=0}) end
end
function M.validate_weapon_set(command)
    local p=command.payload
    assert(type(p)=="table" and (p.set==0 or p.set==1),"weapon set must be 0 or 1")
    owner(command)
end
function M.apply_weapon_set(command)
    local _,layout=assert(find.layout(owner(command)))
    layout:set("active_weapon_set",command.payload.set)
end

local function terms(vendor)
    vendor=string.lower(vendor)
    for _,entity in ipairs(ecs.query({all={"d2legacy.vendor.terms"}})) do
        local value=ecs.get(entity,"d2legacy.vendor.terms")
        if value:get("vendor")==vendor then return value end
    end
    error("unknown vendor")
end
local function held(layout_entity,id,entities)
    local entity,item=assert(find.item(layout_entity,id,entities))
    local placed=assert(ecs.get(entity,"d2legacy.item.placement"))
    assert(placed:get("container")=="held","item is not held")
    return entity,item,placed
end
local function vendor_items(layout_entity,category,excluded,entities)
    local values={}
    for _,entity in ipairs(entities) do
        local item,placed=ecs.get(entity,"d2legacy.item.identity"),ecs.get(entity,"d2legacy.item.placement")
        if item and placed and item:get("owner"):id()==layout_entity:id() and entity:id()~=(excluded and excluded:id() or -1)
            and placed:get("container")=="vendor" and placed:get("slot")==category then values[#values+1]={entity=entity,item=item,placed=placed} end
    end
    table.sort(values,function(a,b) return a.item:get("code")~=b.item:get("code") and a.item:get("code")<b.item:get("code") or a.item:get("id")<b.item:get("id") end)
    return values
end
local function arrange(layout,layout_entity,category,added,excluded,entities)
    local values=vendor_items(layout_entity,category,excluded,entities)
    if added then values[#values+1]=added; table.sort(values,function(a,b) return a.item:get("code")~=b.item:get("code") and a.item:get("code")<b.item:get("code") or a.item:get("id")<b.item:get("id") end) end
    local occupied={}
    for _,value in ipairs(values) do
        local found
        for page=0,1000 do for x=0,layout:get("vendor_width")-value.item:get("width") do for y=0,layout:get("vendor_height")-value.item:get("height") do
            local d={container="vendor",slot=category,page=page,x=x,y=y,belt_slot=0,weapon_set=0}
            local blocked=false
            for _,other in ipairs(occupied) do if other.page==page and x<other.x+other.w and other.x<x+value.item:get("width") and y<other.y+other.h and other.y<y+value.item:get("height") then blocked=true break end end
            if not blocked then found=d; occupied[#occupied+1]={x=x,y=y,w=value.item:get("width"),h=value.item:get("height"),page=page}; break end
        end if found then break end end if found then break end end
        assert(found,"vendor item cannot be arranged"); placement.set(value.placed,found)
    end
end
local function validate_vendor(command,buy)
    local p=command.payload
    assert(type(p.item_id)=="string" and p.item_id~="" and type(p.vendor)=="string" and p.vendor~="","item and vendor are required")
    if not buy then assert(type(p.category)=="string" and p.category~="" and not string.find(p.category,"/",1,true),"valid category is required") end
    owner(command); terms(p.vendor)
end
function M.apply_sell(command)
    local p=command.payload; local entities=ecs.query({any={"d2legacy.items.layout","d2legacy.item.identity"}})
    local layout_entity,layout=assert(find.layout(owner(command),entities)); local entity,item,placed=held(layout_entity,p.item_id,entities)
    local rule=terms(p.vendor); local price=math.floor(item:get("base_cost")*rule:get("buy_multiplier")/PRICE_SCALE)
    if rule:get("max_buy")>0 then price=math.min(price,rule:get("max_buy")) end
    arrange(layout,layout_entity,p.category,{entity=entity,item=item,placed=placed},nil,entities); layout:set("carried_gold",layout:get("carried_gold")+price)
end
function M.apply_buy(command)
    local p=command.payload; local entities=ecs.query({any={"d2legacy.items.layout","d2legacy.item.identity"}})
    local layout_entity,layout=assert(find.layout(owner(command),entities)); local entity,item=assert(find.item(layout_entity,p.item_id,entities)); local placed=assert(ecs.get(entity,"d2legacy.item.placement"))
    assert(placed:get("container")=="vendor","item is not vendor stock"); assert(not placement.slot_occupant(entity,item,{container="held",slot="",belt_slot=0,weapon_set=0},entities),"held item already exists")
    local rule=terms(p.vendor); local price=math.floor(item:get("base_cost")*rule:get("sell_multiplier")/PRICE_SCALE); assert(layout:get("carried_gold")>=price,"insufficient carried gold")
    local category=placed:get("slot"); arrange(layout,layout_entity,category,nil,entity,entities); placement.set(placed,{container="held",x=0,y=0,slot="",belt_slot=0,weapon_set=0,page=0}); layout:set("carried_gold",layout:get("carried_gold")-price)
end
function M.register()
    commands.register({kind="item.move",authorities={"player","administrator"},validate=M.validate_move,apply=M.apply_move})
    commands.register({kind="item.weapon_set",authorities={"player","administrator"},validate=M.validate_weapon_set,apply=M.apply_weapon_set})
    commands.register({kind="item.vendor_sell",authorities={"player","administrator"},validate=function(c) validate_vendor(c,false) end,apply=M.apply_sell})
    commands.register({kind="item.vendor_buy",authorities={"player","administrator"},validate=function(c) validate_vendor(c,true) end,apply=M.apply_buy})
end
return M
