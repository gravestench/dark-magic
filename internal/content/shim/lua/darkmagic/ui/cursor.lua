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
    local legacy_mode = {
        sheet = definition.sheet,
        frame = definition.frame,
        hotspot = definition.hotspot,
    }
    local modes = definition.modes or { default = legacy_mode }
    local default_mode = definition.default_mode or "default"
    local cursor = {
        node = nil,
        width = 0,
        height = 0,
        definition = definition,
        visible = true,
        requested_mode = default_mode,
        mode = nil,
        offset_x = 0,
        offset_y = 0,
    }

    if render.assets_available() then
        cursor.node = render.create("cursor", parent)
        cursor.node:set_z(1000)
    end

    local function apply_mode(self, name)
        if self.mode == name then return end
        local mode = assert(modes[name], "unknown cursor mode: " .. tostring(name))
        self.mode = name
        self.hotspot = mode.hotspot or { x = 0, y = 0 }
        if not self.node then return end
        if mode.fps then
            local _
            _, self.width, self.height, self.offset_x, self.offset_y = self.node:set_dc6_animation(
                mode.sheet,
                palettes[definition.palette],
                definition.direction,
                mode.fps,
                mode.loop or "loop",
                "offsets"
            )
        else
            self.width, self.height, self.offset_x, self.offset_y = self.node:set_dc6(
                mode.sheet,
                palettes[definition.palette],
                definition.direction,
                mode.frame or 0
            )
        end
    end

    function cursor:set_mode(name)
        assert(modes[name], "unknown cursor mode: " .. tostring(name))
        self.requested_mode = name
        apply_mode(self, name)
    end

    function cursor:set_visible(visible)
        self.visible = visible == true
        if self.node then self.node:set_visible(self.visible) end
    end

    function cursor:update()
        local mode = input.down("pointer_primary") and modes.pressed and "pressed" or self.requested_mode
        apply_mode(self, mode)
        if not self.node then return end
        local x, y = input.cursor()
        self.node:set_position(
            x - self.hotspot.x + self.offset_x + self.width / 2,
            y - self.hotspot.y + self.offset_y + self.height / 2
        )
    end

    -- Position immediately so the cursor never flashes at the retained-node
    -- origin before the first scene update.
    apply_mode(cursor, default_mode)
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
-- screens/overlays that do not get a separate shell-owned cursor automatically.
-- The automatic cursor deliberately does NOT populate self.cursor: several
-- existing scenes use that field as part of their own initialization/lifecycle
-- state, so changing it from the decorator would alter scene behavior.
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

    local function cursor_for(self)
        return self.cursor or self.__darkmagic_shell_cursor
    end

    local function ensure_cursor(self)
        if options.hidden or self.cursor or self.__darkmagic_shell_cursor then return end
        -- Keep the shell cursor independent of scene-root transforms. The scene
        -- scope still owns and destroys the render node automatically.
        self.__darkmagic_shell_cursor = create(nil, definition, palettes)
    end

    local function visible_for(self, focused)
        local visible = focused ~= false and not options.hidden
        if visible and options.visible_when then
            visible = options.visible_when(self) == true
        end
        return visible
    end

    scene.create = function(self)
        -- Do not auto-create here. Some existing screens create their authored
        -- cursor in enter(); navigation always calls Enter after Create, so
        -- waiting avoids creating an orphan cursor that enter would replace.
        if original_create then original_create(self) end
    end

    scene.enter = function(self)
        if original_enter then original_enter(self) end
        ensure_cursor(self)
        -- Navigation calls Enter synchronously before the new scene is exposed
        -- to the next frame. Apply cursor ownership now so a push/replace cannot
        -- flash the previous scene's pointer for one frame.
        M.focus(cursor_for(self), visible_for(self, true))
    end

    scene.update = function(self, elapsed, focused, input_allowed, world_view)
        if original_update then original_update(self, elapsed, focused, input_allowed, world_view) end
        M.focus(cursor_for(self), visible_for(self, focused))
    end

    scene.destroy = function(self)
        -- An authored scene cursor and the automatic shell cursor are distinct
        -- ownership paths. Unregister whichever exist without mutating the
        -- scene's own cursor field or using it as shell lifecycle state.
        unregister(self.cursor)
        unregister(self.__darkmagic_shell_cursor)
        if original_destroy then original_destroy(self) end
        self.__darkmagic_shell_cursor = nil
    end

    return scene
end

return M
