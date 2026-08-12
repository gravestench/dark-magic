local a = manager:accessibility()
assert(
    a[1].focused == false
        and a[2].checked == true
        and a[3].value == "é"
        and a[4].value == 7
        and a[4].focused == true
)
