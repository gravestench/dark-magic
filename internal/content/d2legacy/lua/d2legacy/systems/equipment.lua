-- Project the active equipped hand into the player's melee/composite facts.
--
-- Containers own where the weapon is. Combat owns the resolved profile it
-- consumes. This tiny projection system is the deliberate bridge between the
-- two domains and falls back to Diablo's unarmed HTH profile.

local ecs = require("engine.ecs/v1")
local resolution = require("d2legacy.policy.stat_resolution")
local weapon_selection = require("d2legacy.policy.weapon_selection")
local M = {}

local function layout_for(entities, owner)
    for _, entity in ipairs(entities) do
        local layout = ecs.get(entity, "d2legacy.items.layout")
        if layout and layout:get("owner") == owner then
            return entity, layout
        end
    end
end

local function active_weapons(entities, layout_entity, set)
    local result = {}
    for _, slot in ipairs({ "rarm", "larm" }) do
        for _, entity in ipairs(entities) do
            local item = ecs.get(entity, "d2legacy.item.identity")
            local placement = ecs.get(entity, "d2legacy.item.placement")
            local melee = ecs.get(entity, "d2legacy.item.melee")
            if
                item
                and placement
                and melee
                and item:get("owner"):id() == layout_entity:id()
                and placement:get("container") == "equipment"
                and placement:get("slot") == slot
                and placement:get("weapon_set") == set
            then
                result[slot] = { entity = entity, item = item, melee = melee }
                break
            end
        end
    end
    return result
end

local function can_dual_wield(player)
    local identity = ecs.get(player, "d2legacy.player.identity")
    local class = identity and string.lower(identity:get("class")) or ""
    return weapon_selection.can_dual_wield(class)
end

local function sync_sources(entities, structural, player, prefix, stat, wanted)
    local found = {}
    for _, entity in ipairs(entities) do
        local source = ecs.get(entity, "d2legacy.stat.source")
        if
            source
            and source:get("target"):id() == player:id()
            and string.find(source:get("source_id"), prefix, 1, true) == 1
        then
            local value = wanted[source:get("source_id")]
            if value and value.value ~= 0 then
                source:set("value", value.value)
                source:set("operation", value.operation or "add")
                source:set("order", value.order or 0)
                found[source:get("source_id")] = true
            else
                structural:destroy(entity)
            end
        end
    end
    local source_ids = {}
    for source_id in pairs(wanted) do
        source_ids[#source_ids + 1] = source_id
    end
    table.sort(source_ids)
    for _, source_id in ipairs(source_ids) do
        local value = wanted[source_id]
        if value.value ~= 0 and not found[source_id] then
            structural:create({
                ["d2legacy.stat.source"] = {
                    target = player,
                    source_id = source_id,
                    stat = stat,
                    operation = value.operation or "add",
                    value = value.value,
                    order = value.order or 0,
                },
            })
        end
    end
end

local function placement_is_active(placement, set)
    if placement:get("container") ~= "equipment" then
        return false
    end
    local slot = placement:get("slot")
    if slot == "rarm" or slot == "larm" then
        return placement:get("weapon_set") == set
    end
    return true
end

local function local_percent(entities, item_entity, stat)
    local total = 0
    for _, entity in ipairs(entities) do
        local modifier = ecs.get(entity, "d2legacy.item.stat_modifier")
        if
            modifier
            and modifier:get("item"):id() == item_entity:id()
            and modifier:get("stat") == stat
            and modifier:get("operation") == "local_percent"
        then
            total = total + modifier:get("value")
        end
    end
    return total
end

local function equipped_defense(entities, layout_entity, set)
    local wanted = {}
    for _, entity in ipairs(entities) do
        local item = ecs.get(entity, "d2legacy.item.identity")
        local placement = ecs.get(entity, "d2legacy.item.placement")
        local armor = ecs.get(entity, "d2legacy.item.armor")
        if
            item
            and placement
            and armor
            and item:get("owner"):id() == layout_entity:id()
            and placement_is_active(placement, set)
            and armor:get("defense") ~= 0
        then
            local percent = local_percent(entities, entity, "defense")
            local base = armor:get("defense")
            if percent ~= 0 and armor:get("base_defense_max") > 0 then
                base = armor:get("base_defense_max") + 1
            end
            local defense = resolution.local_value(base, percent)
            wanted["equipment:defense:" .. item:get("id")] = { value = defense, operation = "add", order = 10 }
        end
    end
    return wanted
end

local function modifier_sources(entities, layout_entity, set, stat)
    local wanted = {}
    for _, entity in ipairs(entities) do
        local modifier = ecs.get(entity, "d2legacy.item.stat_modifier")
        if
            modifier
            and modifier:get("stat") == stat
            and modifier:get("operation") ~= "local_percent"
            and modifier:get("value") ~= 0
        then
            local item_entity = modifier:get("item")
            local item = ecs.get(item_entity, "d2legacy.item.identity")
            local placement = ecs.get(item_entity, "d2legacy.item.placement")
            if
                item
                and placement
                and item:get("owner"):id() == layout_entity:id()
                and placement_is_active(placement, set)
            then
                local source_id = "equipment:modifier:"
                    .. stat
                    .. ":"
                    .. item:get("id")
                    .. ":"
                    .. modifier:get("source_kind")
                    .. ":"
                    .. tostring(modifier:get("order"))
                    .. ":"
                    .. modifier:get("source_id")
                wanted[source_id] = {
                    value = modifier:get("value"),
                    operation = modifier:get("operation"),
                    order = modifier:get("order"),
                }
            end
        end
    end
    return wanted
end

function M.register()
    ecs.system({
        id = "d2legacy.player.equipment_melee_profile",
        phase = "pre_simulation",
        query = {
            any = {
                "d2legacy.world.player_control",
                "d2legacy.items.layout",
                "d2legacy.item.identity",
                "d2legacy.item.stat_modifier",
                "d2legacy.stat.source",
            },
        },
        read = {
            "d2legacy.world.player_control",
            "d2legacy.items.layout",
            "d2legacy.item.identity",
            "d2legacy.item.placement",
            "d2legacy.item.melee",
            "d2legacy.item.armor",
            "d2legacy.item.stat_modifier",
            "d2legacy.stat.source",
            "d2legacy.player.identity",
        },
        write = { "d2legacy.combat.melee_profile", "d2legacy.player.appearance", "d2legacy.stat.source" },
        update = function(_, entities, structural)
            for _, player in ipairs(entities) do
                local control = ecs.get(player, "d2legacy.world.player_control")
                if control then
                    local layout_entity, layout = layout_for(entities, control:get("player"))
                    local weapons = {}
                    local active_set = 0
                    if layout then
                        active_set = layout:get("active_weapon_set")
                        weapons = active_weapons(entities, layout_entity, active_set)
                    end
                    local profile = ecs.get(player, "d2legacy.combat.melee_profile")
                    local appearance = ecs.get(player, "d2legacy.player.appearance")
                    if profile and appearance then
                        local primary = weapons.rarm or weapons.larm
                        local secondary = can_dual_wield(player) and weapons.rarm and weapons.larm
                        local weapon = primary and primary.melee
                        local item = primary and primary.item
                        profile:set("range", weapon and weapon:get("range") or 2)
                        profile:set("physical_min", weapon and weapon:get("physical_min") or 256)
                        profile:set("physical_max", weapon and weapon:get("physical_max") or 512)
                        profile:set("primary_attack_rating", weapon and weapon:get("attack_rating") or 0)
                        profile:set("primary_weapon_attack_rate", weapon and weapon:get("attack_rate") or 0)
                        profile:set("primary_hand", weapons.rarm and "rarm" or weapons.larm and "larm" or "unarmed")
                        profile:set("secondary_range", secondary and secondary.melee:get("range") or 0)
                        profile:set("secondary_physical_min", secondary and secondary.melee:get("physical_min") or 0)
                        profile:set("secondary_physical_max", secondary and secondary.melee:get("physical_max") or 0)
                        profile:set("secondary_attack_rating", secondary and secondary.melee:get("attack_rating") or 0)
                        profile:set(
                            "secondary_weapon_attack_rate",
                            secondary and secondary.melee:get("attack_rate") or 0
                        )
                        profile:set("secondary_hand", secondary and "larm" or "")
                        profile:set("dual_wield", secondary ~= nil)
                        appearance:set("weapon_class", weapon and weapon:get("weapon_class") or "HTH")
                        local attack = {}
                        if weapon and weapon:get("attack_rating") ~= 0 then
                            attack["equipment:attack:" .. item:get("id")] = {
                                value = weapon:get("attack_rating"),
                                operation = "add",
                                order = 10,
                            }
                        end
                        sync_sources(entities, structural, player, "equipment:attack:", "attack_rating", attack)
                        local attackrate = {}
                        if weapon and weapon:get("attack_rate") ~= 0 then
                            attackrate["equipment:attackrate:" .. item:get("id")] = {
                                value = weapon:get("attack_rate"),
                                operation = "add",
                                order = 10,
                            }
                        end
                        sync_sources(entities, structural, player, "equipment:attackrate:", "attackrate", attackrate)
                        if layout_entity then
                            sync_sources(
                                entities,
                                structural,
                                player,
                                "equipment:defense:",
                                "defense",
                                equipped_defense(entities, layout_entity, active_set)
                            )
                            sync_sources(
                                entities,
                                structural,
                                player,
                                "equipment:modifier:attack_rating:",
                                "attack_rating",
                                modifier_sources(entities, layout_entity, active_set, "attack_rating")
                            )
                            sync_sources(
                                entities,
                                structural,
                                player,
                                "equipment:modifier:defense:",
                                "defense",
                                modifier_sources(entities, layout_entity, active_set, "defense")
                            )
                            sync_sources(
                                entities,
                                structural,
                                player,
                                "equipment:modifier:attackrate:",
                                "attackrate",
                                modifier_sources(entities, layout_entity, active_set, "attackrate")
                            )
                            sync_sources(
                                entities,
                                structural,
                                player,
                                "equipment:modifier:item_fasterattackrate:",
                                "item_fasterattackrate",
                                modifier_sources(entities, layout_entity, active_set, "item_fasterattackrate")
                            )
                        else
                            sync_sources(entities, structural, player, "equipment:defense:", "defense", {})
                            sync_sources(
                                entities,
                                structural,
                                player,
                                "equipment:modifier:attack_rating:",
                                "attack_rating",
                                {}
                            )
                            sync_sources(entities, structural, player, "equipment:modifier:defense:", "defense", {})
                            sync_sources(entities, structural, player, "equipment:modifier:attackrate:", "attackrate", {})
                            sync_sources(
                                entities,
                                structural,
                                player,
                                "equipment:modifier:item_fasterattackrate:",
                                "item_fasterattackrate",
                                {}
                            )
                        end
                    end
                end
            end
        end,
    })
end
return M
