local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "ecs",
    tier = "fast",
    cases = {
        test.case("semantic_cues_tolerate_absent_optional_events", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.create({ ["d2legacy.combat.melee_event"] = { kind = "hit_resolved" } })
                local world = require("d2legacy.gameplay.world")
                local cues = world.semantic_cues()
                test.assert(#cues == 1, [=[#cues == 1]=])
                test.assert(
                    cues[1].cue_type == "combat" and cues[1].kind == "hit_resolved",
                    [=[cues[1].cue_type == "combat" and cues[1].kind == "hit_resolved"]=]
                )
                test.assert(
                    #world.semantic_cues({ [cues[1].entity_id] = true }) == 0,
                    [=[#world.semantic_cues({ [cues[1].entity_id] = true }) == 0]=]
                )
            end),
        }),
        test.case("connected_roster_excludes_compatibility_hero_but_keeps_local_and_peer_mirrors", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local world = require("d2legacy.gameplay.world")

                local function player(owner, class, token, x)
                    return ecs.create({
                        ["d2legacy.player.identity"] = {
                            character_id = owner .. "-character",
                            player = owner,
                            name = class,
                            class = class,
                        },
                        ["d2legacy.player.appearance"] = {
                            cof = "",
                            token = token,
                            palette = "data/global/Palette/units/pal.dat",
                            weapon_class = "HTH",
                        },
                        ["d2legacy.player.animation"] = { direction = 0, mode = "NU" },
                        ["d2legacy.world.position"] = { x = x, y = 10 },
                        ["d2legacy.world.facing"] = { direction = 0, directions = 16 },
                        ["d2legacy.world.location"] = { act = 1, level_id = 2 },
                    })
                end

                local compatibility = player("player-2", "Assassin", "AI", 10)
                player("player-2", "Barbarian", "BA", 18)
                player("player-1", "Assassin", "AI", 10)

                local snapshots = world.player_snapshots("player-2", true, compatibility)
                test.assert(#snapshots == 2, [=[#snapshots == 2]=])
                local by_token = {}
                for _, snapshot in ipairs(snapshots) do by_token[snapshot.token] = snapshot end
                test.assert(by_token.BA and by_token.BA.x == 18, [=[by_token.BA and by_token.BA.x == 18]=])
                test.assert(by_token.AI and by_token.AI.x == 10, [=[by_token.AI and by_token.AI.x == 10]=])
            end),
        }),
    },
})
