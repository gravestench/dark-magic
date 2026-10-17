local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "authority",
    tier = "fast",
    covers = { "internal/game/item" },
    records = {
        ["data/global/excel/monstats.txt"] = {
            { Id = "akara", Code = "PS", MonStatsEx = "akara", npc = "1", NameStr = "Akara" },
        },
        ["data/global/excel/monstats2.txt"] = {
            { Id = "akara", BaseW = "hth", TR = "1", TRv = "lit", OverlayHeight = "2" },
        },
        ["data/global/excel/monlvl.txt"] = { { Level = "1" } },
    },
    initial_data = {
        ["d2legacy.interactions"] = {
            owner = "alice",
            targets = {
                {
                    id = "act1-akara",
                    npc = "Akara",
                    x = 21,
                    y = 46,
                    radius = 4,
                    act = 1,
                    level_id = 1,
                    selectable = true,
                },
                {
                    id = "room-object",
                    npc = "Room Object",
                    x = 4,
                    y = 5,
                    radius = 2,
                    resident_id = "level:2:room-object",
                    level_id = 2,
                    room_id = "7",
                },
            },
        },
        ["d2legacy.items"] = {
            owner = "alice",
            inventory_width = 10,
            inventory_height = 4,
            stash_width = 6,
            stash_height = 8,
            cube_width = 3,
            cube_height = 4,
            belt_capacity = 4,
            active_weapon_set = 0,
            vendor_width = 10,
            vendor_height = 10,
            carried_gold = 100,
            stashed_gold = 200,
            items = {
                {
                    id = "sword",
                    code = "ssd",
                    width = 1,
                    height = 3,
                    body_slots = "rarm,larm",
                    belt_eligible = false,
                    base_cost = 100,
                    inventory_dc6 = "sword.dc6",
                    world_dc6 = "flpsword.dc6",
                    world_animated = true,
                    container = "inventory",
                    x = 0,
                    y = 0,
                    slot = "",
                    belt_slot = 0,
                    weapon_set = 0,
                    page = 0,
                    melee_range = 2,
                    physical_min = 512,
                    physical_max = 1024,
                    melee_weapon_class = "1HS",
                },
            },
        },
    },
    cases = {
        test.case("guide_model_and_selection_share_one_target_lifetime", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local targets = ecs.query({ all = { "d2legacy.interaction.target", "d2legacy.world.selectable" } })
                test.expect(#targets):equals(1)
                local entity = targets[1]
                local target = ecs.get(entity, "d2legacy.interaction.target")
                local position = ecs.get(entity, "d2legacy.world.position")
                local location = ecs.get(entity, "d2legacy.world.location")
                test.expect(target:get("id")):equals("act1-akara")
                test.expect(position:get("x")):equals(target:get("x"))
                test.expect(position:get("y")):equals(target:get("y"))
                test.expect(location:get("act")):equals(1)
                test.expect(location:get("level_id")):equals(1)
                local appearance = ecs.get(entity, "d2legacy.monster.appearance")
                test.expect(appearance:get("token")):equals("PS")
                test.expect(appearance:get("mode")):equals("NU")
                test.expect(appearance:get("components")):equals("TR=LIT")
                test.expect(appearance:get("overlay_height")):equals(2)
                local selected = require("d2legacy.gameplay.targeting").selectable_at(21, 46, 1)
                test.expect(selected.id):equals("act1-akara")
                require("d2legacy.items.bootstrap").ensure_player("alice")
                test.expect(#ecs.query({ all = { "d2legacy.world.selectable", "d2legacy.interaction.target" } }))
                    :equals(1)
            end),
        }),
        test.case("materializes_initial_items_into_lua_owned_ecs", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                test.assert(
                    #ecs.query({ all = { "d2legacy.items.layout" } }) == 1,
                    [=[#ecs.query({ all = { "d2legacy.items.layout" } }) == 1]=]
                )
                test.assert(
                    #ecs.query({ all = { "d2legacy.item.identity" } }) == 1,
                    [=[#ecs.query({ all = { "d2legacy.item.identity" } }) == 1]=]
                )
            end),
        }),
        test.case("attaches_authored_world_targets_to_stable_room_residency", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local residents = ecs.query({ all = { "d2legacy.world.room_resident" } })
                test.expect(#residents):equals(1)
                local resident = ecs.get(residents[1], "d2legacy.world.room_resident")
                test.expect(resident:get("id")):equals("level:2:room-object")
                test.expect(resident:get("level_id")):equals(2)
                test.expect(resident:get("room_id")):equals("7")
            end),
        }),
        test.case("moves_items_through_lua_owned_policy", {
            test.submit({
                tick = 1,
                sequence = 1,
                player = "alice",
                kind = "item.move",
                payload = { item_id = "sword", destination = { container = "held" } },
            }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local items = ecs.query({
                    all = {
                        "d2legacy.item.identity",
                        "d2legacy.item.placement",
                    },
                })
                test.assert(#items == 1, [=[#items == 1]=])
                local placement = ecs.get(items[1], "d2legacy.item.placement")
                test.assert(placement:get("container") == "held", [=[placement:get("container") == "held"]=])
            end),
        }),
    },
})
