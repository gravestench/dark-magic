local test = require("d2legacy.tests/v1")

local function aura(entity_id, target_id, state_id, period)
    return {
        entity_id = entity_id,
        target_entity_id = target_id,
        state_id = state_id,
        aura = true,
        aura_period_ticks = period,
    }
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("distinct_active_auras_alternate_one_visual_without_suppressing_other_states", {
            test.run(function()
                local cycle = require("d2legacy.gameplay.aura_overlay_cycle")
                local timed = { entity_id = 1, target_entity_id = 20, state_id = "curse", aura = false }
                local snapshots = {
                    timed,
                    aura(2, 20, "might", 50),
                    aura(3, 20, "prayer", 50),
                    aura(4, 30, "thorns", 50),
                }
                local selected, state = cycle.select(snapshots, nil, 0)
                test.assert(#selected == 3, [=[one aura per target plus non-aura states is selected]=])
                test.expect(selected[1].state_id):equals("curse")
                test.expect(selected[2].state_id):equals("might")
                test.expect(selected[3].state_id):equals("thorns")

                selected, state = cycle.select(snapshots, state, 1.9)
                test.expect(selected[2].state_id):equals("might")
                selected, state = cycle.select(snapshots, state, 0.2)
                test.expect(selected[2].state_id):equals("prayer")
                selected = cycle.select(snapshots, state, 2.0)
                test.expect(selected[2].state_id):equals("might")
            end),
        }),
    },
})
