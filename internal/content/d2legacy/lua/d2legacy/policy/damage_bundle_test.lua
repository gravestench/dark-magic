local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("preserves_poison_length_beside_typed_damage_rates", {
            test.run(function()
                local bundle = require("d2legacy.policy.damage_bundle")
                local rolled = bundle.normalize({ physical = 256, poison = 128 })
                local mitigated = bundle.normalize({ physical = 256, poison = 64 })
                local staged = bundle.stage_component(rolled, mitigated, 3)
                test.expect(staged.poison_rolled_raw):equals(128)
                test.expect(staged.poison_mitigated_raw):equals(64)
                test.expect(staged.poison_length_frames):equals(3)
                test.expect(bundle.stage_component(rolled, mitigated).poison_length_frames):equals(0)
            end),
        }),
    },
})
