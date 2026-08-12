local ui = require("d2legacy.ui.controls")
manager = ui.new()
manager:add({ id = "outside", x = 0, y = 0, width = 10, height = 10 })
manager:add_checkbox({ id = "check", scope = "form", x = 20, y = 0, width = 10, height = 10 })
manager:add_text_field({
    id = "name",
    scope = "form",
    x = 40,
    y = 0,
    width = 10,
    height = 10,
    max_length = 3,
})
manager:add_scrollbar({
    id = "volume",
    scope = "form",
    x = 60,
    y = 0,
    width = 20,
    height = 10,
    min = 0,
    max = 10,
    value = 5,
    step = 2,
})
manager:set_scope("form")
