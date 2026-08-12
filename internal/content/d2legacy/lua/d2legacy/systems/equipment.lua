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
                and placement:get("weapon_set")==set then return entity,item,melee end
        end
    end
end

local function sync_attack_rating_source(entities,structural,player,weapon_entity,item,weapon)
    local wanted=weapon and weapon:get("attack_rating") or 0
    local source_id=weapon_entity and "equipment:"..item:get("id") or ""
    local found=false
    for _,entity in ipairs(entities) do
        local source=ecs.get(entity,"d2legacy.stat.source")
        if source and source:get("target"):id()==player:id()
            and string.sub(source:get("source_id"),1,10)=="equipment:" then
            if source:get("source_id")==source_id and wanted~=0 then
                source:set("value",wanted); found=true
            else structural:destroy(entity) end
        end
    end
    if wanted~=0 and not found then
        structural:create({["d2legacy.stat.source"]={target=player,
            source_id=source_id,stat="attack_rating",value=wanted}})
    end
end

function M.register()
    ecs.system({id="d2legacy.player.equipment_melee_profile",phase="pre_simulation",
        query={any={"d2legacy.world.player_control","d2legacy.items.layout","d2legacy.item.identity","d2legacy.stat.source"}},
        read={"d2legacy.world.player_control","d2legacy.items.layout","d2legacy.item.identity",
            "d2legacy.item.placement","d2legacy.item.melee","d2legacy.stat.source"},
        write={"d2legacy.combat.melee_profile","d2legacy.player.appearance","d2legacy.stat.source"},
        update=function(_,entities,structural)
            for _,player in ipairs(entities) do
                local control=ecs.get(player,"d2legacy.world.player_control")
                if control then
                    local layout_entity,layout=layout_for(entities,control:get("player"))
                    local weapon_entity,item,weapon
                    if layout then weapon_entity,item,weapon=active_weapon(entities,layout_entity,layout:get("active_weapon_set")) end
                    local profile=ecs.get(player,"d2legacy.combat.melee_profile")
                    local appearance=ecs.get(player,"d2legacy.player.appearance")
                    if profile and appearance then
                        profile:set("range",weapon and weapon:get("range") or 2)
                        profile:set("physical_min",weapon and weapon:get("physical_min") or 256)
                        profile:set("physical_max",weapon and weapon:get("physical_max") or 512)
                        appearance:set("weapon_class",weapon and weapon:get("weapon_class") or "HTH")
                        sync_attack_rating_source(entities,structural,player,weapon_entity,item,weapon)
                    end
                end
            end
        end})
end
return M
