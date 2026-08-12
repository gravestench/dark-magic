local s=require("d2legacy.save/v1")
local id=assert(s.create("hero","Hero","Amazon"))
assert(id=="hero" and s.select(id) and s.selected().name=="Hero")
