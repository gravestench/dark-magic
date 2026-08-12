return {
    initial_data_json = [[{
        "d2legacy.items": {
            "owner": "alice",
            "inventory_width": 10,
            "inventory_height": 4,
            "stash_width": 6,
            "stash_height": 8,
            "cube_width": 3,
            "cube_height": 4,
            "belt_capacity": 4,
            "active_weapon_set": 0,
            "vendor_width": 10,
            "vendor_height": 10,
            "carried_gold": 100,
            "stashed_gold": 200,
            "items": [{
                "id": "sword",
                "code": "ssd",
                "width": 1,
                "height": 3,
                "body_slots": "rarm,larm",
                "belt_eligible": false,
                "base_cost": 100,
                "inventory_dc6": "sword.dc6",
                "world_dc6": "flpsword.dc6",
                "world_animated": true,
                "container": "inventory",
                "x": 0,
                "y": 0,
                "slot": "",
                "belt_slot": 0,
                "weapon_set": 0,
                "page": 0,
                "melee_range": 2,
                "physical_min": 512,
                "physical_max": 1024,
                "melee_weapon_class": "1HS"
            }]
        }
    }]],
    tests = {
        materializes_initial_items_into_lua_owned_ecs = {
            {run = function()
                local ecs = require("engine.ecs/v1")
                assert(#ecs.query({all={"d2legacy.items.layout"}}) == 1)
                assert(#ecs.query({all={"d2legacy.item.identity"}}) == 1)
            end},
        },
        moves_items_through_lua_owned_policy = {
            {submit = {
                tick = 1,
                sequence = 1,
                player = "alice",
                kind = "item.move",
                payload = [[{"item_id":"sword","destination":{"container":"held"}}]],
            }},
            {step = 1},
            {run = function()
                local ecs = require("engine.ecs/v1")
                local items = ecs.query({all={
                    "d2legacy.item.identity", "d2legacy.item.placement",
                }})
                assert(#items == 1)
                local placement = ecs.get(items[1], "d2legacy.item.placement")
                assert(placement:get("container") == "held")
            end},
        },
    },
}
