local save=require("d2legacy.save/v1")
local c=save.characters()[1]
assert(c.appearance.cof=="hero.cof" and c.appearance.palette=="units.dat")
assert(c.appearance.direction==3 and c.appearance.components.HD=="head.dcc")
assert(c.stats.strength==25 and c.stats.health==70 and c.stats.max_health==80)
c.appearance.components.HD="mutated"
assert(save.characters()[1].appearance.components.HD=="head.dcc")
