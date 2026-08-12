local a = manager:accessibility()
assert(activated == "two" and visual_state == "hover" and a[2].focused)
