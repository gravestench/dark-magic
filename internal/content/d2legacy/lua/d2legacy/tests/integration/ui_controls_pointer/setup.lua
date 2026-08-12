local ui = require("d2legacy.ui.controls")
manager = ui.new()
activated = ""
visual_state = ""
manager:add({
    id = "one",
    label = "First",
    x = 0,
    y = 0,
    width = 20,
    height = 20,
    on_activate = function(c)
        activated = c.id
    end,
})
manager:add({
    id = "two",
    label = "Second",
    x = 30,
    y = 0,
    width = 20,
    height = 20,
    on_activate = function(c)
        activated = c.id
    end,
    on_state = function(_, state)
        visual_state = state
    end,
})
