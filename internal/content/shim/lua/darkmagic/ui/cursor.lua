-- Manifest-driven Diablo II software cursor and scene integration.
--
-- Cursor instances are retained render nodes owned by the scene scope that
-- created them. A small module-level registry lets the focused scene hide every
-- other instance, which matters when a blocking overlay prevents the scene
-- below it from receiving another update. This avoids stale/ghost pointers
-- without moving authored cursor assets into native platform code.
local render = require("dm.render/v1")
local input = require("dm.input/v1")

local M = {}
local instances = {}

local function register(cursor)
    instances[cursor] = true
end

local function unregister(cursor)
    if cursor then instances[cursor] = nil end
end

local function create(parent, definition, palettes)
    local cursor = {
        node = nil,
        width = 0,
        height = 0,
        definition = definition,
        visible = true,
    }

    if render.assets_available() then
        cursor.node = render.create("cursor", parent)
        cursor.node:set_z(1000)
        cursor.width, cursor.height = cursor.node:set_dc6(
            definition.sheet,
            palettes[definition.palette],
            definition.direction,
            definition.frame
        )
    end

    function cursor:set_visible(visible)
        self.visible = visible == true
        if self.node then self.node:set_visible(self.visible) end
    end

    function cursor:update()
        if not self.node then return end
        local x, y = input.cursor()
        self.node:set_position(
            x - definition.hotspot.x + self.width / 2,
            y - definition.hotspot.y + self.height / 2
        )
    end

    -- Position immediately so the cursor never flashes at the retained-node
    -- origin before the first scene update.
    cursor:update()
    register(cursor)
    return cursor
end

function M.new(parent, definition, palettes)
    return create(parent, definition, palettes)
end

-- Make one scene's pointer authoritative for this frame. Hiding every other
-- registered pointer is important for overlays with blocks_update_below=true:
-- the covered scene does not receive a focused=false update in that case.
function M.focus(cursor, visible)
    for instance in pairs(instances) do
        local active = instance == cursor and visible ~= false
        instance:set_visible(active)
        if active then instance:update() end
    end
end

-- Decorate a scene with default software-cursor ownership/focus semantics.
-- Existing screens that already construct self.cursor keep that instance;
-- screens/overlays that do not get one automatically. This keeps the helper
-- backwards compatible while making cursor visibility a shell-wide invariant.
--
-- Options:
--   hidden = true                       never show a cursor in this scene
--   visible_when = function(self) ...  dynamic visibility (e.g. video playing)
function M.wrap(scene, definition, palettes, options)
    if scene.__darkmagic_cursor_wrapped then return scene end
    scene.__darkmagic_cursor_wrapped = true

    options = options or {}
    local original_create = scene.create
    local original_enter = scene.enter
    local original_update = scene.update
    local original_destroy = scene.destroy

    local function ensure_cursor(self)
        if options.hidden or self.cursor then return end
        -- Keep the cursor independent of scene-root transforms. The scene scope
        -- still owns and destroys the render node automatically.
        self.cursor = create(nil, definition, palettes)
    end

    local function visible_for(self, focused)
        local visible = focused ~= false and not options.hidden
        if visible and options.visible_when then
            visible = options.visible_when(self) == true
        end
        return visible
    end

    scene.create = function(self)
        if original_create then original_create(self) end
        ensure_cursor(self)
    end

    scene.enter = function(self)
        if original_enter then original_enter(self) end
        ensure_cursor(self)
        -- Navigation calls Enter synchronously before the new scene is exposed
        -- to the next frame. Apply cursor ownership now so a push/replace cannot
        -- flash the previous scene's pointer for one frame.
        M.focus(self.cursor, visible_for(self, true))
    end

    scene.update = function(self, elapsed, focused)
        if original_update then original_update(self, elapsed, focused) end
        M.focus(self.cursor, visible_for(self, focused))
    end

    scene.destroy = function(self)
        unregister(self.cursor)
        if original_destroy then original_destroy(self) end
        self.cursor = nil
    end

    return scene
end

return M
