local enter = [[{
"character_id":"hero","player":"alice","name":"Hero","class":"Amazon",
"level":1,"experience":0,"dexterity":20,"defense":0,"fire_resistance":20,
"health":10,"max_health":10,"mana":20,"max_mana":20,
"expansion":true,"hardcore":false,"cof":"","palette":"units",
"direction":0,"mode":"NU","x":10,"y":12,
"world_width":100,"world_height":80,"act":1,"level_id":1,"skills":[]
}]]

return {
    tests = {
        derived_mitigation_and_lethal_damage_vectors = {
            {submit_system = {tick = 1, sequence = 1, kind = "system.player.enter", payload = enter}},
            {step = 1},
            {run = function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({all={"d2legacy.player.identity"}})[1]
                ecs.create({["d2legacy.stat.source"]={
                    target=player, source_id="shield:fire", stat="fire_resist", value=30,
                }})
                ecs.create({["d2legacy.stat.source"]={
                    target=player, source_id="armor:physical", stat="physical_resist", value=25,
                }})
                ecs.create({["d2legacy.stat.source"]={
                    target=player, source_id="armor:flat", stat="physical_reduction_raw", value=20,
                }})
            end},
            {step = 1},
            {run = function()
                local ecs = require("engine.ecs/v1")
                local mitigation = require("d2legacy.policy.mitigation")
                local damage = require("d2legacy.policy.damage")
                local player = ecs.query({all={"d2legacy.player.identity"}})[1]
                local defense = ecs.get(player, "d2legacy.combat.defense")
                assert(defense:get("fire_resist") == 50)
                assert(defense:get("physical_resist") == 25)
                assert(mitigation.apply(1000, "fire", defense) == 500)
                assert(mitigation.apply(1000, "physical", defense) == 730)
                local remaining, lethal, applied = damage.apply(player, 4096, ecs, "fire")
                assert(applied == 2048 and remaining == 512 and not lethal)
                remaining, lethal, applied = damage.apply(player, 1024, ecs, "fire")
                assert(applied == 512 and remaining == 0 and lethal)
            end},
        },
    },
}
