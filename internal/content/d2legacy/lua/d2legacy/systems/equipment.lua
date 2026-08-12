-- Project the active equipped hand into the player's melee/composite facts.
--
-- Containers own where the weapon is. Combat owns the resolved profile it
-- consumes. This tiny projection system is the deliberate bridge between the
-- two domains and falls back to Diablo's unarmed HTH profile.

local ecs=require("engine.ecs/v1")
local M={}

local function layout_for(entities,owner)
    for _,entity in ipairs(entities) do
        local layout=ecs.get(entity,"d2legacy.items.layout")
        if layout and layout:get("owner")==owner then return entity,layout end
    end
end

local function active_weapon(entities,layout_entity,set)
    for _,slot in ipairs({"rarm","larm"}) do
        for _,entity in ipairs(entities) do
            local item=ecs.get(entity,"d2legacy.item.identity")
            local placement=ecs.get(entity,"d2legacy.item.placement")
            local melee=ecs.get(entity,"d2legacy.item.melee")
            if item and placement and melee and item:get("owner"):id()==layout_entity:id()
                and placement:get("container")=="equipment" and placement:get("slot")==slot
                and placement:get("weapon_set")==set then return melee end
        end
    end
end

function M.register()
    ecs.system({id="d2legacy.player.equipment_melee_profile",phase="pre_simulation",
        query={any={"d2legacy.world.player_control","d2legacy.items.layout","d2legacy.item.identity"}},
        read={"d2legacy.world.player_control","d2legacy.items.layout","d2legacy.item.identity",
            "d2legacy.item.placement","d2legacy.item.melee"},
        write={"d2legacy.combat.melee_profile","d2legacy.player.appearance"},
        update=function(_,entities)
            for _,player in ipairs(entities) do
                local control=ecs.get(player,"d2legacy.world.player_control")
                if control then
                    local layout_entity,layout=layout_for(entities,control:get("player"))
                    local weapon=layout and active_weapon(entities,layout_entity,layout:get("active_weapon_set")) or nil
                    local profile=ecs.get(player,"d2legacy.combat.melee_profile")
                    local appearance=ecs.get(player,"d2legacy.player.appearance")
                    if profile and appearance then
                        profile:set("range",weapon and weapon:get("range") or 2)
                        profile:set("physical_min",weapon and weapon:get("physical_min") or 256)
                        profile:set("physical_max",weapon and weapon:get("physical_max") or 512)
                        appearance:set("weapon_class",weapon and weapon:get("weapon_class") or "HTH")
                    end
                end
            end
        end})
end
return M
