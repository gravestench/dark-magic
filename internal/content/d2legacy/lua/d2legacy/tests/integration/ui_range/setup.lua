local ui = require("d2legacy.ui.controls")
manager = ui.new()
slider = manager:add_slider({
    id = "volume",
    x = 10,
    y = 10,
    width = 100,
    height = 20,
    min = 0,
    max = 100,
    step = 10,
    value = 40,
})
