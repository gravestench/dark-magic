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
        test.case("state_snapshots_preserve_target_relationship_and_events_copy_position", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.component({
                    name = "d2legacy.skill.aura_effect",
                    fields = {
                        { name = "emitter", type = "entity" },
                        { name = "target", type = "entity" },
                        { name = "source_id", type = "string" },
                        { name = "skill_id", type = "i64" },
                        { name = "skill_level", type = "i64" },
                        { name = "state_id", type = "string" },
                        { name = "refresh_delay", type = "i64" },
                    },
                })
                local world = require("d2legacy.gameplay.world")
                local target = ecs.create({
                    ["d2legacy.monster.appearance"] = { overlay_height = 4 },
                    ["d2legacy.world.position"] = { x = 12, y = 8 },
                    ["d2legacy.world.location"] = { act = 1, level_id = 2 },
                    ["d2legacy.world.facing"] = { direction = 5, directions = 16 },
                })
                local instance = ecs.create({
                    ["d2legacy.state.instance"] = {
                        target = target,
                        state_id = "syntheticcurse",
                        source_id = "skill:owner:1",
                        applied_tick = 4,
                        expires_tick = 40,
                        policy = "refresh_same_source",
                    },
                })
                local aura = ecs.create({
                    ["d2legacy.skill.aura_effect"] = {
                        emitter = target,
                        target = target,
                        source_id = "aura:owner:98",
                        skill_id = 98,
                        skill_level = 1,
                        state_id = "might",
                        refresh_delay = 50,
                    },
                })
                ecs.create({
                    ["d2legacy.state.event"] = {
                        kind = "state_applied",
                        tick = 4,
                        target = target,
                        state_id = "syntheticcurse",
                        source_id = "skill:owner:1",
                        expires_tick = 40,
                        reason = "apply",
                    },
                })
                local snapshots = world.state_snapshots()
                test.assert(#snapshots == 2, [=[#snapshots == 2]=])
                local by_state = {}
                for _, snapshot in ipairs(snapshots) do
                    by_state[snapshot.state_id] = snapshot
                end
                local timed = by_state.syntheticcurse
                test.assert(
                    timed.entity_id == instance:id()
                        and timed.target_entity_id == target:id()
                        and timed.x == 12
                        and timed.y == 8
                        and timed.level_id == 2
                        and timed.direction == 5
                        and timed.overlay_height == 4,
                    [=[active state snapshot follows its live ECS target]=]
                )
                local aura_snapshot = by_state.might
                test.assert(
                    aura_snapshot.entity_id == aura:id()
                        and aura_snapshot.target_entity_id == target:id()
                        and aura_snapshot.aura
                        and aura_snapshot.aura_period_ticks == 50
                        and aura_snapshot.x == 12
                        and aura_snapshot.y == 8,
                    [=[active aura snapshot uses the same presentation contract]=]
                )
                local cues = world.semantic_cues()
                local cue = cues[#cues]
                test.assert(
                    cue.cue_type == "state"
                        and cue.kind == "state_applied"
                        and cue.target == nil
                        and cue.target_entity_id == target:id()
                        and cue.x == 12
                        and cue.y == 8
                        and cue.overlay_height == 4,
                    [=[state event copies target identity and current position]=]
                )
            end),
        }),
        test.case("connected_roster_keeps_peers_and_excludes_the_authenticated_owner", {
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
                        ["d2legacy.player.movement_stats"] = {
                            velocitypercent = -5,
                            item_fastermovevelocity = 20,
                        },
                        ["d2legacy.world.position"] = { x = x, y = 10 },
                        ["d2legacy.world.facing"] = { direction = 0, directions = 16 },
                        ["d2legacy.world.location"] = { act = 1, level_id = 2 },
                    })
                end

                player("player-2", "Barbarian", "BA", 18)
                player("player-1", "Assassin", "AI", 10)

                local snapshots = world.player_snapshots("player-2", false)
                test.assert(#snapshots == 1, [=[#snapshots == 1]=])
                test.assert(snapshots[1].token == "AI", [=[snapshots[1].token == "AI"]=])
                test.assert(snapshots[1].class == "Assassin", [=[snapshots[1].class == "Assassin"]=])
                test.assert(
                    snapshots[1].velocitypercent == -5 and snapshots[1].item_fastermovevelocity == 20,
                    [=[snapshots[1].velocitypercent == -5 and snapshots[1].item_fastermovevelocity == 20]=]
                )
                test.assert(snapshots[1].x == 10, [=[snapshots[1].x == 10]=])
            end),
        }),
    },
})
