-- Permanent offers belong to each admitted owner's item graph. A purchase copies
-- the representative, preserving supply and checkpointed purchase identities.
-- Only living, undiscounted Normal-town purchases are admitted.
local ecs = require("engine.ecs/v1")
local find = require("d2legacy.items.find")
local contexts = require("d2legacy.interactions.context")
local rules = require("d2legacy.policy.game_rules")
local stats = require("d2legacy.gameplay.stat_sources")
local M = {}

function M.supported(owner)
    if rules.difficulty() ~= 0 then
        return false
    end
    for _, entity in
        ipairs(
            ecs.query({ all = { "d2legacy.player.identity", "d2legacy.player.progress", "d2legacy.player.vitals" } })
        )
    do
        if ecs.get(entity, "d2legacy.player.identity"):get("player") == owner then
            local location = ecs.get(entity, "d2legacy.world.location")
            local quest = ecs.get(entity, "d2legacy.quest.progress")
            return location
                and location:get("act") == 1
                and location:get("level_id") == 1
                and ecs.get(entity, "d2legacy.player.vitals"):get("health") > 0
                and not ecs.get(entity, "d2legacy.player.death")
                and not (quest and quest:get("quest_id") ~= 1 and (quest:get("completed") or quest:get("rewarded")))
                and stats.resolve(ecs.query({ all = { "d2legacy.stat.source" } }), entity, "item_reducedprices", 0)
                    == 0
        end
    end
    return false
end

function M.trading(owner, vendor)
    local context, target = contexts.active(owner)
    if not context or context:get("dialogue_page") ~= "trade" or not M.supported(owner) then
        return false
    end
    return string.lower(target:get("vendor")) == vendor
        and vendor == "akara"
        and string.find("," .. target:get("categories") .. ",", ",misc,", 1, true) ~= nil
end

function M.ensure(owner)
    local layout = find.layout(owner)
    if not layout then
        return
    end
    local rows = require("d2legacy.data.vendor_consumables").load()
    for index, row in ipairs(rows) do
        local id = "akara-offer:" .. row.code
        if not find.item(layout, id) then
            ecs.create({
                ["d2legacy.item.identity"] = {
                    owner = layout,
                    id = id,
                    code = row.code,
                    width = row.width,
                    height = row.height,
                    base_cost = row.base_cost,
                    identified = true,
                    belt_eligible = row.belt_eligible,
                },
                ["d2legacy.item.placement"] = { container = "vendor", slot = "misc", x = index - 1 },
                ["d2legacy.item.presentation"] = { inventory_dc6 = row.inventory_dc6 },
                ["d2legacy.vendor.offer"] = { vendor = "akara", price = row.price, purchases = 0 },
            })
        end
    end
end

function M.purchase(layout, entity, item, destination)
    local offer = ecs.get(entity, "d2legacy.vendor.offer")
    local number = offer:get("purchases") + 1
    local identity = item:snapshot()
    identity.owner = layout
    identity.id = "akara-purchase:" .. item:get("code") .. ":" .. tostring(number)
    ecs.create({
        ["d2legacy.item.identity"] = identity,
        ["d2legacy.item.placement"] = destination,
        ["d2legacy.item.presentation"] = ecs.get(entity, "d2legacy.item.presentation"):snapshot(),
    })
    offer:set("purchases", number)
end

return M
