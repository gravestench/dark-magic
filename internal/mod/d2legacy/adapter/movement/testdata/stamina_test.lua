local rules = require("d2legacy.movement_rules/v1")
-- Numeric research vectors, not a player-spend transaction.
local cases = {
    -- class, level, old Vitality, current, passive %, old max, new max, new current
    { "Amazon", 1, 20, 21504, 0, 21504, 21760, 21760 },
    { "Sorceress", 1, 10, 18944, 0, 18944, 19200, 19200 },
    { "Necromancer", 1, 15, 20224, 0, 20224, 20480, 20480 },
    { "Paladin", 1, 25, 22784, 0, 22784, 23040, 23040 },
    { "Barbarian", 1, 25, 23552, 0, 23552, 23808, 23808 },
    { "Druid", 1, 25, 21504, 0, 21504, 21760, 21760 },
    { "Assassin", 1, 20, 24320, 0, 24320, 24640, 24640 },
    { "Amazon", 1, 21, 16210, 0, 21760, 22016, 16656 },
    { "Amazon", 1, 22, 0, 0, 22016, 22272, 256 },
    { "Barbarian", 1, 26, 18168, 0, 23808, 24064, 18619 },
    { "Barbarian", 1, 27, 0, 0, 24064, 24320, 256 },
    -- A running frame is inferred to precede allocation in these two cases.
    { "Barbarian", 12, 25, 29114, 30, 34278, 34611, 29652 },
    { "Barbarian", 1, 26, 20815, 0, 23808, 24064, 21294 },
    { "Assassin", 1, 22, 23340, 0, 24960, 25280, 23959 },
}
for _, case in ipairs(cases) do
    local class, level, vitality, current, passive, previous, maximum, expected = unpack(case)
    assert(rules.maximum_stamina(class, level, vitality, 0, 0, 0, passive) == previous)
    assert(rules.maximum_stamina(class, level, vitality + 1, 0, 0, 0, passive) == maximum)
    local _, _, _, quarters = rules.class_facts(class)
    local rescaled = rules.rescale_stamina(current, previous, maximum)
    assert(rules.allocate_vitality_stamina(class, current, previous, maximum) == expected, class .. " allocation order")
    if current < previous then
        assert(rescaled ~= expected, class .. " generic-only negative control")
        assert(
            rules.rescale_stamina(current + quarters * 64, previous, maximum) ~= expected,
            class .. " gain-before-rescale negative control"
        )
    end
end
assert(rules.rescale_stamina(0, 24064, 24320) == 0)
assert(rules.rescale_stamina(18168, 23808, 24064) == 18363)
assert(not pcall(rules.class_facts, "unknown"))
