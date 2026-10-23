local rules = require("d2legacy.movement_rules/v1")
local walk, run = rules.rates("Amazon", 23, 0)
assert(walk / 25 * 65536 == 30224 and run / 25 * 65536 == 42512)
for _, elapsed in ipairs({0.04, 0.08}) do
    local x, y, moving = rules.velocity(10, 20, "Amazon", {running = true, target = {x = 10.3, y = 20}}, 0, 0, elapsed)
    assert(math.abs(x * elapsed - 0.3) < 1e-12 and y == 0 and moving)
end

for _, level in ipairs({1, 40, 75, 103, 109}) do
    assert(rules.is_town(level), "original 1.13c town level " .. level)
end
for _, level in ipairs({0, 2, 45, 80, 108, 114, 110, -1}) do
    assert(not rules.is_town(level), "outside-town level " .. level)
end
