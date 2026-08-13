local test = require("d2legacy.tests/v1")

local category =
    [["category":{"id":"skeleton","group":1,"base_max":1,"replacement":"replace_oldest","warp_with_owner":true}]]

local function attach_payload(unit_id)
    return '{"unit_id":"'
        .. unit_id
        .. '","owner_id":"player:hero","ultimate_owner_id":"player:hero",'
        .. category
        .. "}"
end

local function create_actors()
    local ecs = require("engine.ecs/v1")
    local function actor(id, kind)
        ecs.create({
            ["d2legacy.world.selectable"] = {
                id = id,
                kind = kind,
                label = id,
                owner = "",
                radius = 1,
                priority = 1,
            },
        })
    end
    actor("player:hero", "player")
    actor("monster:skeleton-one", "friendly")
    actor("monster:skeleton-two", "friendly")
end

local function assert_restored_relationships()
    local ecs = require("engine.ecs/v1")
    local active, inactive = 0, 0
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.owned_unit" } })) do
        local relation = ecs.get(entity, "d2legacy.owned_unit")
        if relation:get("active") then
            active = active + 1
        else
            inactive = inactive + 1
        end
        test.assert(relation:get("owner_id") == "player:hero", [=[relation:get("owner_id") == "player:hero"]=])
        test.assert(relation:get("warp_with_owner"), [=[relation:get("warp_with_owner")]=])
    end
    test.assert(active == 1 and inactive == 1, [=[active == 1 and inactive == 1]=])

    local attribution = require("d2legacy.owned_units.attribution")
    local entities = ecs.query({ all = { "d2legacy.world.selectable" } })
    local credit = attribution.resolve(entities, "monster:skeleton-two")
    test.assert(credit.immediate_owner_id == "player:hero", [=[credit.immediate_owner_id == "player:hero"]=])
    test.assert(credit.ultimate_owner_id == "player:hero", [=[credit.ultimate_owner_id == "player:hero"]=])
end

local function fake_entity(id)
    return {
        id = function()
            return id
        end,
    }
end

local function assert_replacement_policy()
    local limits = require("d2legacy.owned_units.limits")
    local candidates = {
        { entity = fake_entity(7), category = "wolf", group = 2, active = true, created_tick = 1 },
        {
            entity = fake_entity(8),
            category = "skeleton",
            group = 1,
            active = true,
            created_tick = 2,
        },
        {
            entity = fake_entity(9),
            category = "skeleton",
            group = 1,
            active = true,
            created_tick = 3,
        },
    }
    local victims = limits.victims(candidates, {
        id = "skeleton",
        group = 2,
        base_max = 2,
        replacement = "replace_newest",
    })
    test.assert(#victims == 2, [=[#victims == 2]=])
    test.assert(
        victims[1].entity:id() == 7 and victims[2].entity:id() == 9,
        [=[victims[1].entity:id() == 7 and victims[2].entity:id() == 9]=]
    )
    test.assert(not pcall(function()
        limits.victims(candidates, {
            id = "skeleton",
            group = 1,
            base_max = 1,
            replacement = "reject",
        })
    end), "reject replacement policy accepted an over-limit unit")
end

local function assert_owned_unit_policy()
    assert_restored_relationships()
    assert_replacement_policy()
end

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/ownedunit" },
    cases = {
        test.case("limits_attribution_and_restore_are_authoritative", {
            { run = create_actors },
            test.submit_system({
                tick = 1,
                sequence = 1,
                kind = "system.owned_unit.attach",
                payload = attach_payload("monster:skeleton-one"),
            }),
            test.step(1),
            test.submit_system({
                tick = 2,
                sequence = 2,
                kind = "system.owned_unit.attach",
                payload = attach_payload("monster:skeleton-two"),
            }),
            test.step(1),
            test.restore_checkpoint(),
            { run = assert_owned_unit_policy },
        }),
    },
})
