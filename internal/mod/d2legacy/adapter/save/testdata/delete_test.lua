local s=require("d2legacy.save/v1")
assert(s.select("hero") and s.delete("hero"))
assert(#s.characters()==0 and s.selected()==nil)
