-- Turn immutable host/import values into authoritative ECS item state once.
-- The input module is not a mutable back door: every call returns a deep copy,
-- and all future gameplay observes or changes only the entities created here.

local ecs=require("engine.ecs/v1")
local M={}

function M.load()
    local available,initial=pcall(require,"engine.initial_data/v1")
    if not available then return end
    local data=initial.get("d2legacy.items")
    if not data or not data.owner then return end
    local layout=ecs.create({["d2legacy.items.layout"]={owner=data.owner,
        inventory_width=data.inventory_width or 0,inventory_height=data.inventory_height or 0,
        stash_width=data.stash_width or 0,stash_height=data.stash_height or 0,
        cube_width=data.cube_width or 0,cube_height=data.cube_height or 0,
        belt_capacity=data.belt_capacity or 0,active_weapon_set=data.active_weapon_set or 0,
        vendor_width=data.vendor_width or 0,vendor_height=data.vendor_height or 0,
        carried_gold=data.carried_gold or 0,stashed_gold=data.stashed_gold or 0}})
    for _,item in ipairs(data.items or {}) do
        ecs.create({
            ["d2legacy.item.identity"]={owner=layout,id=item.id,code=item.code,width=item.width,height=item.height,
                body_slots=item.body_slots or "",belt_eligible=item.belt_eligible or false,
                base_cost=item.base_cost or 0,applied_services=item.applied_services or ""},
            ["d2legacy.item.placement"]={container=item.container or "world",x=item.x or 0,y=item.y or 0,
                slot=item.slot or "",belt_slot=item.belt_slot or 0,weapon_set=item.weapon_set or 0,page=item.page or 0},
            ["d2legacy.item.presentation"]={inventory_dc6=item.inventory_dc6 or "",world_dc6=item.world_dc6 or "",
                world_animated=item.world_animated or false,composite=item.composite or "",weapon_class=item.weapon_class or ""},
            ["d2legacy.item.melee"]={range=item.melee_range or 0,physical_min=item.physical_min or 0,
                physical_max=item.physical_max or 0,weapon_class=item.melee_weapon_class or ""},
        })
    end
    for vendor,terms in pairs(data.trade_terms or {}) do
        ecs.create({["d2legacy.vendor.terms"]={vendor=string.lower(vendor),
            buy_multiplier=terms.buy_multiplier or 0,sell_multiplier=terms.sell_multiplier or 0,
            max_buy=terms.max_buy or 0}})
    end
    local interactions=initial.get("d2legacy.interactions") or {}
    local targets={}
    for _,target in ipairs(interactions.targets or {}) do
        targets[target.id]=ecs.create({["d2legacy.interaction.target"]={id=target.id,npc=target.npc,vendor=target.vendor or "",categories=target.categories or "",services=target.services or "",x=target.x,y=target.y,radius=target.radius}})
    end
    local no_target=ecs.create()
    ecs.create({["d2legacy.interaction.context"]={owner=interactions.owner or data.owner,target=targets[interactions.initial_target or ""] or no_target}})
end
return M
