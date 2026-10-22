local find = require("d2legacy.items.find")
local world = require("d2legacy.items.world")
local M = {}

function M.assert_selected(selector)
    local selected = require("d2legacy.gameplay.targeting").selectable_at(1, 1, 2)
    assert(selected and selected.kind == "item" and selected.id == selector, "ground item selection changed")
end

function M.create_destination_item(id)
    require("d2legacy.items.entity").create(find.layout("bob"), {
        id = id,
        code = "hp1",
        width = 1,
        height = 1,
        container = "inventory",
        x = 0,
        y = 0,
        generation = { quality = "normal", level = 1, facts = "unrelated-generation" },
    })
    local item = find.item(find.layout("alice"), "sword")
    assert(world.can_pick_up(world.player("bob"), item), "fixture cannot reach the ground item")
end

function M.create_late_drop_item(id)
    require("d2legacy.items.entity").create(find.layout("aaron"), {
        id = id,
        code = "hp1",
        width = 1,
        height = 1,
        container = "inventory",
        x = 0,
        y = 0,
    })
end

return M
