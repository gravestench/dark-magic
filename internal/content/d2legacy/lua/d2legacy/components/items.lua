-- Authoritative item/container facts for one Diablo II owner.
--
-- Item identity, presentation recipe, and placement are separate components so
-- moving an item cannot accidentally rewrite what it is. Layout is one owner
-- entity; each item is another entity pointing back to that owner.

local ecs=require("engine.ecs/v1")
local M={}
local function component(name,fields) ecs.component({name=name,fields=fields}) end

function M.register()
    component("d2legacy.items.layout",{
        {name="owner",type="string"},{name="inventory_width",type="i64"},{name="inventory_height",type="i64"},
        {name="stash_width",type="i64"},{name="stash_height",type="i64"},{name="cube_width",type="i64"},
        {name="cube_height",type="i64"},{name="belt_capacity",type="i64"},{name="active_weapon_set",type="i64"},
        {name="vendor_width",type="i64"},{name="vendor_height",type="i64"},
        {name="carried_gold",type="i64"},{name="stashed_gold",type="i64"},
    })
    component("d2legacy.item.identity",{
        {name="owner",type="entity"},{name="id",type="string"},{name="code",type="string"},
        {name="width",type="i64"},{name="height",type="i64"},{name="body_slots",type="string"},
        {name="belt_eligible",type="bool"},{name="base_cost",type="i64"},{name="applied_services",type="string"},
    })
    component("d2legacy.item.placement",{
        {name="container",type="string"},{name="x",type="i64"},{name="y",type="i64"},{name="slot",type="string"},
        {name="belt_slot",type="i64"},{name="weapon_set",type="i64"},{name="page",type="i64"},
    })
    component("d2legacy.item.presentation",{
        {name="inventory_dc6",type="string"},{name="world_dc6",type="string"},{name="world_animated",type="bool"},
        {name="composite",type="string"},{name="weapon_class",type="string"},
    })
    component("d2legacy.item.melee",{
        {name="range",type="f64"},{name="physical_min",type="i64"},
        {name="physical_max",type="i64"},{name="weapon_class",type="string"},
    })
    component("d2legacy.items.bootstrap",{
        {name="owner",type="string"},{name="payload",type="string"},
    })
end
return M
