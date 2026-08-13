local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("default_selector_prefers_right_hand_without_alternating", {
            test.run(function()
                local select = require("d2legacy.policy.weapon_selection").select
                test.expect(select(0, "rarm", "larm", 0)):equals("rarm")
                test.expect(select(0, "rarm", "larm", 1)):equals("rarm")
            end),
        }),
        test.case("explicit_selectors_choose_left_sequence_or_unarmed", {
            test.run(function()
                local select = require("d2legacy.policy.weapon_selection").select
                test.expect(select(1, "rarm", "larm", 0)):equals("larm")
                test.expect(select(3, "rarm", "larm", 1)):equals("larm")
                test.expect(select(3, "rarm", "larm", 2)):equals("rarm")
                test.expect(select(4, "rarm", "larm", 0)):equals("unarmed")
            end),
        }),
        test.case("only_legacy_player_classes_expose_a_second_weapon", {
            test.run(function()
                local policy = require("d2legacy.policy.weapon_selection")
                test.expect(policy.can_dual_wield("Barbarian")):is_true()
                test.expect(policy.can_dual_wield("Assassin")):is_true()
                test.expect(policy.can_dual_wield("Amazon")):is_false()
            end),
        }),
    },
})
