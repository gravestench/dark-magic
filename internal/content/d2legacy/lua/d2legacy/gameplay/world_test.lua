local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "presentation",
    tier = "fast",
    tests = {
        semantic_cues_tolerate_absent_optional_events = {
            {
                run = function()
                    local ecs = require("engine.ecs/v1")
                    ecs.create({ ["d2legacy.combat.melee_event"] = { kind = "hit_resolved" } })
                    local world = require("d2legacy.gameplay.world")
                    local cues = world.semantic_cues()
                    assert(#cues == 1)
                    assert(cues[1].cue_type == "combat" and cues[1].kind == "hit_resolved")
                    assert(#world.semantic_cues({ [cues[1].entity_id] = true }) == 0)
                end,
            },
        },
    },
})
