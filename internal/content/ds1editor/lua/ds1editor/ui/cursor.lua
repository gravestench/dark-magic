local assets = require("ds1editor.ui.assets")
local input = require("engine.input/v1")
local render = require("engine.render/v1")

local cursor = {}

-- Layer the editor-owned software cursor above a scene and keep it synchronized with native coordinates.
function cursor.wrap(scene)
    local original_create, original_update, original_destroy = scene.create, scene.update, scene.destroy
    -- Attach one editor-owned cursor after the wrapped scene has its root.
    function scene:create(...)
        if original_create then original_create(self, ...) end
        self.editor_cursor = render.create("cursor", self.root)
        self.editor_cursor:set_z(10000)
        assets.apply(self.editor_cursor, "cursors-markers", "cursor_pointer")
        self.editor_cursor_mode = "cursor_pointer"
    end
    -- Mirror native pointer state after scene input so pressed feedback is current.
    function scene:update(...)
        if original_update then original_update(self, ...) end
        if not self.editor_cursor then return end
        local mode = input.down("pointer_primary") and "cursor_pressed" or "cursor_pointer"
        if mode ~= self.editor_cursor_mode then
            assets.apply(self.editor_cursor, "cursors-markers", mode)
            self.editor_cursor_mode = mode
        end
        local x, y = input.cursor()
        self.editor_cursor:set_position(x + 12, y + 12)
    end
    -- Release the cursor before the wrapped scene tears down its retained tree.
    function scene:destroy(...)
        if self.editor_cursor and self.editor_cursor:exists() then self.editor_cursor:destroy() end
        self.editor_cursor = nil
        if original_destroy then original_destroy(self, ...) end
    end
    return scene
end

return cursor
