local a = manager:accessibility()
assert(activated == "" and visual_state == "pressed" and a[2].focused and a[2].role == "button")
