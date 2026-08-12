local a = manager:accessibility()
assert(slider.value == 50 and a[1].role == "slider" and a[1].focused == true)
