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

local function sync_sources(entities,structural,player,prefix,stat,wanted)
    local found={}
    for _,entity in ipairs(entities) do
        local source=ecs.get(entity,"d2legacy.stat.source")
        if source and source:get("target"):id()==player:id()
            and string.find(source:get("source_id"),prefix,1,true)==1 then
            local value=wanted[source:get("source_id")]
            if value and value~=0 then
                source:set("value",value); found[source:get("source_id")]=true
            else structural:destroy(entity) end
        end
    end
    for source_id,value in pairs(wanted) do
        if value~=0 and not found[source_id] then
            structural:create({["d2legacy.stat.source"]={target=player,
                source_id=source_id,stat=stat,value=value}})
        end
    end
end

local function equipped_defense(entities,layout_entity)
    local wanted={}
    for _,entity in ipairs(entities) do
        local item=ecs.get(entity,"d2legacy.item.identity")
        local placement=ecs.get(entity,"d2legacy.item.placement")
        local armor=ecs.get(entity,"d2legacy.item.armor")
        if item and placement and armor and item:get("owner"):id()==layout_entity:id()
            and placement:get("container")=="equipment" and armor:get("defense")~=0 then
            wanted["equipment:defense:"..item:get("id")]=armor:get("defense")
        end
    end
    return wanted
end

function M.register()
    ecs.system({id="d2legacy.player.equipment_melee_profile",phase="pre_simulation",
        query={any={"d2legacy.world.player_control","d2legacy.items.layout","d2legacy.item.identity","d2legacy.stat.source"}},
        read={"d2legacy.world.player_control","d2legacy.items.layout","d2legacy.item.identity",
            "d2legacy.item.placement","d2legacy.item.melee","d2legacy.item.armor","d2legacy.stat.source"},
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
                        local attack={}
                        if weapon and weapon:get("attack_rating")~=0 then
                            attack["equipment:attack:"..item:get("id")]=weapon:get("attack_rating")
                        end
                        sync_sources(entities,structural,player,"equipment:attack:","attack_rating",attack)
                        sync_sources(entities,structural,player,"equipment:defense:","defense",
                            equipped_defense(entities,layout_entity))
                    end
                end
            end
        end})
end
return M
