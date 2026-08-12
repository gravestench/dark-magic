local enter = [[{
"character_id":"hero","player":"alice","name":"Hero","class":"Amazon",
"level":1,"experience":0,"dexterity":20,"defense":0,
"health":50,"max_health":50,"mana":20,"max_mana":20,
"expansion":true,"hardcore":false,"cof":"","palette":"units",
"direction":0,"mode":"NU","x":10,"y":12,
"world_width":100,"world_height":80,"act":1,"level_id":1,"skills":[]
}]]

return {
    tests = {
        player_entry_and_movement_use_authoritative_systems = {
            {submit_system = {tick = 1, sequence = 1, kind = "system.player.enter", payload = enter}},
            {step = 1},
            {submit = {tick = 2, sequence = 1, player = "alice", kind = "player.move",
                payload = [[{"x":1,"y":1,"running":true}]]}},
            {step = 1},
            {run = function()
                local ecs = require("engine.ecs/v1")
                local players = ecs.query({all={
                    "d2legacy.player.identity", "d2legacy.world.velocity",
                    "d2legacy.player.animation",
                }})
                assert(#players == 1)
                local velocity = ecs.get(players[1], "d2legacy.world.velocity")
                local animation = ecs.get(players[1], "d2legacy.player.animation")
                local facing = ecs.get(players[1], "d2legacy.world.facing")
                local mode = ecs.get(players[1], "d2legacy.player.movement_mode")
                local expected = 15 * 0.7071067811865476
                assert(math.abs(velocity:get("x") - expected) < 0.000000001)
                assert(math.abs(velocity:get("y") - expected) < 0.000000001)
                assert(animation:get("mode") == "RN")
                assert(facing:get("direction") == 4)
                assert(mode:get("running") == true)
            end},
            {submit = {tick = 3, sequence = 2, player = "alice", kind = "player.move",
                payload = [[{"x":0,"y":0,"running":false,"target":{"x":20,"y":16}}]]}},
            {step = 1},
            {run = function()
                local ecs = require("engine.ecs/v1")
                local player = ecs.query({all={"d2legacy.player.identity"}})[1]
                local animation = ecs.get(player, "d2legacy.player.animation")
                local facing = ecs.get(player, "d2legacy.world.facing")
                assert(animation:get("mode") == "WL")
                assert(facing:get("direction") == 15)
            end},
        },
    },
}
