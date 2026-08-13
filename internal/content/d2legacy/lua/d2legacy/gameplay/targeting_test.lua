local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "ecs",
    tier = "fast",
    cases = {
        test.case("selects_by_priority_distance_and_stable_id", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local function spawn(id, kind, x, priority)
                    ecs.create({
                        ["d2legacy.world.position"] = { x = x, y = 10 },
                        ["d2legacy.world.selectable"] = {
                            id = id,
                            kind = kind,
                            label = id,
                            owner = "owner:" .. id,
                            radius = 3,
                            priority = priority,
                        },
                    })
                end
                spawn("near", "npc", 10, 0)
                spawn("far", "npc", 12, 0)
                spawn("item", "item", 12, 5)
                spawn("distance-near", "npc", 20, 0)
                spawn("distance-far", "npc", 22, 0)
                spawn("stable-b", "npc", 30, 0)
                spawn("stable-a", "npc", 30, 0)
                local targeting = require("d2legacy.gameplay.targeting")
                local hit = targeting.selectable_at(10, 10)
                test.assert(
                    hit and hit.id == "item" and hit.kind == "item" and hit.label == "item",
                    [=[hit and hit.id == "item" and hit.kind == "item" and hit.label == "item"]=]
                )
                test.assert(
                    hit.owner == "owner:item" and hit.x == 12 and hit.y == 10 and hit.radius == 3,
                    [=[hit.owner == "owner:item" and hit.x == 12 and hit.y == 10 and hit.radius == 3]=]
                )
                test.assert(
                    targeting.selectable_at(20, 10).id == "distance-near",
                    [=[targeting.selectable_at(20, 10).id == "distance-near"]=]
                )
                test.assert(
                    targeting.selectable_at(30, 10).id == "stable-a",
                    [=[targeting.selectable_at(30, 10).id == "stable-a"]=]
                )
                test.assert(targeting.selectable_at(100, 100) == nil, [=[targeting.selectable_at(100, 100) == nil]=])
                test.assert(targeting.selectable_at(0 / 0, 10) == nil, [=[targeting.selectable_at(0 / 0, 10) == nil]=])
            end),
        }),
    },
})
