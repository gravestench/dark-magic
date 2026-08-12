local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
return test.suite({
    profile = "authority",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "0",
                skill = "Attack",
                skilldesc = "attack",
                leftskill = "1",
                general = "1",
                passive = "0",
            },
            {
                Id = "36",
                skill = "Fire Bolt",
                srvmissile = "firebolt",
                skilldesc = "firebolt",
                leftskill = "1",
                general = "0",
                passive = "0",
                etype = "fire",
                interrupt = "1",
                srvstfunc = "",
                srvdofunc = "",
                mana = "5",
                manashift = "7",
                emin = "3",
                emax = "6",
                HitShift = "8",
            },
        },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "attack", ListRow = "0", IconCel = "0" },
            { skilldesc = "firebolt", ListRow = "1", IconCel = "1" },
        },
    },
    tests = {
        emits_animation_lifecycle_events = {
            {
                submit_system = {
                    tick = 1,
                    sequence = 1,
                    kind = "system.player.enter",
                    payload = fixtures.amazon_entry,
                },
            },
            { step = 1 },
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    ecs.create({
                        ["d2legacy.monster.identity"] = {
                            spawn_id = "event-target",
                            definition_id = "target",
                            base_id = "",
                            graphics_id = "",
                            seed = "1",
                            treasure_class = "",
                        },
                        ["d2legacy.monster.stats"] = {
                            level = 1,
                            health = 4096,
                            max_health = 4096,
                            defense = 0,
                            attack_rating = 0,
                            physical_min = 0,
                            physical_max = 0,
                            experience = 0,
                        },
                        ["d2legacy.world.position"] = { x = 12, y = 12 },
                        ["d2legacy.world.location"] = { act = 1, level_id = 1 },
                        ["d2legacy.world.collider"] = { radius = 1 },
                        ["d2legacy.world.selectable"] = {
                            id = "monster:event-target",
                            kind = "hostile",
                            label = "Target",
                            owner = "",
                            radius = 1,
                            priority = 1,
                        },
                    })
                end,
            },
            {
                submit = {
                    tick = 2,
                    sequence = 1,
                    player = "alice",
                    kind = "player.use_skill",
                    payload = {
                        side = "left",
                        target_x = 12,
                        target_y = 12,
                        target_id = "monster:event-target",
                    },
                },
            },
            { step = 11 },
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    local events = ecs.query({ all = { "d2legacy.combat.attack_animation_event" } })
                    assert(#events == 3)
                    local kinds, ticks = {}, {}
                    for _, entity in ipairs(events) do
                        local event = ecs.get(entity, "d2legacy.combat.attack_animation_event")
                        assert(
                            event:get("attacker_id") == "player:alice"
                                and event:get("target_id") == "monster:event-target"
                                and event:get("skill_id") == 0
                        )
                        kinds[event:get("kind")] = true
                        ticks[event:get("kind")] = event:get("tick")
                    end
                    assert(kinds.attack_started and kinds.attack_impact and kinds.attack_completed)
                    assert(ticks.attack_started < ticks.attack_impact and ticks.attack_impact < ticks.attack_completed)
                end,
            },
        },
    },
})
