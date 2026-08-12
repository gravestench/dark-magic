-- Manifest-driven Diablo II software cursor and scene integration.
--
-- This file teaches THREE important modding ideas at once:
--
-- 1. A cursor is ordinary presentation. It is a retained render node following
--    the logical pointer position; native/platform cursor internals stay hidden.
-- 2. Several scene instances may briefly exist at once, so cursor ownership has
--    to be explicit or an old scene can leave a ghost pointer visible.
-- 3. Lua can DECORATE scene lifecycle callbacks. `wrap` below adds cursor policy
--    around a scene without requiring that scene to copy cursor code.

local render = require("engine.render/v1")
local input = require("engine.input/v1")

local M = {}

-- Module-level set of every cursor object currently known to this helper.
-- A Lua table can act like a set by using objects as keys and `true` as values.
local instances = {}

local function register(cursor)
    instances[cursor] = true
end

local function unregister(cursor)
    -- Assigning nil removes a table entry. The guard also makes nil harmless.
    if cursor then instances[cursor] = nil end
end

-- Internal constructor used by both M.new and the automatic scene decorator.
local function create(parent, definition, palettes)
    -- Older/simple manifests expose one sheet/frame directly. Convert that shape
    -- into the same mode table used by newer animated/multi-state cursors.
    local legacy_mode = {
        sheet = definition.sheet,
        frame = definition.frame,
        hotspot = definition.hotspot,
    }

    -- Prefer explicitly authored modes. Otherwise manufacture one `default` mode.
    local modes = definition.modes or { default = legacy_mode }
    local default_mode = definition.default_mode or "default"

    -- This table is the Lua-side cursor object. It stores cheap state and a
    -- checked render handle; the actual graphics resource remains engine-owned.
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
        -- Cursor art should sit above normal HUD/modal art.
        cursor.node:set_z(1000)
    end

    local function apply_mode(self, name)
        -- Avoid decoding/resetting the same visual mode every update frame.
        if self.mode == name then return end

        local mode = assert(modes[name], "unknown cursor mode: " .. tostring(name))
        self.mode = name
        self.hotspot = mode.hotspot or { x = 0, y = 0 }

        -- Headless execution still updates logical cursor state even without art.
        if not self.node then return end

        if mode.fps then
            -- Animated cursor mode. Lua can ignore a returned value by assigning
            -- it to `_`, a conventional "I intentionally do not use this" name.
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
            -- Static cursor mode. DC6 decoding returns dimensions AND authored
            -- offsets because the cropped frame is positioned around an anchor.
            self.width, self.height, self.offset_x, self.offset_y = self.node:set_dc6(
                mode.sheet,
                palettes[definition.palette],
                definition.direction,
                mode.frame or 0
            )
        end
    end

    function cursor:set_mode(name)
        -- Validate immediately instead of storing a typo that fails much later.
        assert(modes[name], "unknown cursor mode: " .. tostring(name))
        self.requested_mode = name
        apply_mode(self, name)
    end

    function cursor:set_visible(visible)
        -- Only literal true becomes true. This makes nil/false both mean hidden.
        self.visible = visible == true
        if self.node then self.node:set_visible(self.visible) end
    end

    function cursor:update()
        -- Compact conditional expression:
        -- if primary pointer is held AND a `pressed` mode exists, use it;
        -- otherwise use whichever regular mode the caller requested.
        local mode = input.down("pointer_primary") and modes.pressed and "pressed" or self.requested_mode
        apply_mode(self, mode)

        if not self.node then return end

        local x, y = input.cursor()

        -- Input coordinates describe the logical pointer HOTSPOT. The retained
        -- node wants its image CENTER. Therefore:
        --   pointer - hotspot + DC6 frame offset + half decoded dimensions
        self.node:set_position(
            x - self.hotspot.x + self.offset_x + self.width / 2,
            y - self.hotspot.y + self.offset_y + self.height / 2
        )
    end

    -- Initialize immediately. Without this, a newly created retained node could
    -- flash for one frame at its default origin before the scene's first update.
    apply_mode(cursor, default_mode)
    cursor:update()
    register(cursor)
    return cursor
end

function M.new(parent, definition, palettes)
    -- Public constructor is intentionally boring: one obvious entry point over
    -- the internal constructor used by the decorator too.
    return create(parent, definition, palettes)
end

-- Make ONE cursor authoritative for this frame and hide all other registered
-- instances. Why hide all of them instead of trusting old scenes to update?
-- Because a blocking overlay can prevent the scene underneath from receiving
-- another callback at all, so that lower cursor cannot hide itself.
function M.focus(cursor, visible)
    for instance in pairs(instances) do
        local active = instance == cursor and visible ~= false
        instance:set_visible(active)
        if active then instance:update() end
    end
end

-- Decorate a scene with default software-cursor lifecycle/focus semantics.
--
-- Existing scenes that construct `self.cursor` keep that authored cursor.
-- Simpler scenes get an automatic `self.__darkmagic_shell_cursor` instead.
-- The private-looking name is deliberate: it avoids changing the meaning of a
-- scene's own public-ish `self.cursor` field.
--
-- Options:
--   hidden = true                       never show a cursor in this scene
--   visible_when = function(self) ...  dynamic rule, e.g. hide during a movie
function M.wrap(scene, definition, palettes, options)
    -- Decorators may be applied through several registry paths. Mark the table so
    -- accidentally wrapping the same definition twice remains harmless.
    if scene.__darkmagic_cursor_wrapped then return scene end
    scene.__darkmagic_cursor_wrapped = true

    options = options or {}

    -- Functions are ordinary Lua values. Save the originals before replacing
    -- lifecycle fields so our wrapper can delegate to them later.
    local original_create = scene.create
    local original_enter = scene.enter
    local original_update = scene.update
    local original_destroy = scene.destroy

    local function cursor_for(self)
        -- Prefer a cursor the scene explicitly authored; fall back to our private
        -- automatically created one.
        return self.cursor or self.__darkmagic_shell_cursor
    end

    local function ensure_cursor(self)
        -- Any of these conditions means there is nothing for the decorator to do.
        if options.hidden or self.cursor or self.__darkmagic_shell_cursor then return end

        -- Parent=nil keeps the automatic pointer independent of arbitrary scene
        -- root transforms. The SCENE SCOPE still owns the checked render handle.
        self.__darkmagic_shell_cursor = create(nil, definition, palettes)
    end

    local function visible_for(self, focused)
        -- A held item itself becomes the pointer in Diablo II. Item presenters
        -- copy authoritative held-state into this flag; hiding the hand here
        -- prevents an item image and hand image from being drawn simultaneously.
        local visible = focused ~= false and not options.hidden
            and self.__darkmagic_item_held ~= true

        if visible and options.visible_when then
            -- Require literal true from optional dynamic policy.
            visible = options.visible_when(self) == true
        end
        return visible
    end

    scene.create = function(self)
        -- Do NOT auto-create the shell cursor during create. Some existing scenes
        -- create their authored cursor later in enter(). Navigation always enters
        -- after creating, so waiting prevents a useless orphan cursor.
        if original_create then original_create(self) end
    end

    scene.enter = function(self)
        if original_enter then original_enter(self) end
        ensure_cursor(self)

        -- Scene transitions run Enter synchronously before the next visible
        -- frame. Transfer cursor ownership immediately so the previous scene's
        -- pointer cannot flash for a frame during push/replace.
        M.focus(cursor_for(self), visible_for(self, true))
    end

    scene.update = function(self, elapsed, focused, input_allowed, world_view)
        -- Preserve the scene's exact update arguments. The decorator does not
        -- reinterpret gameplay/input routing; it only observes focus afterward.
        if original_update then original_update(self, elapsed, focused, input_allowed, world_view) end
        M.focus(cursor_for(self), visible_for(self, focused))
    end

    scene.destroy = function(self)
        -- Remove both possible ownership paths from our module-level registry.
        -- Native render-resource cleanup still belongs to the scene scope.
        unregister(self.cursor)
        unregister(self.__darkmagic_shell_cursor)

        if original_destroy then original_destroy(self) end

        -- Drop our private Lua reference after original teardown has had a chance
        -- to inspect scene state.
        self.__darkmagic_shell_cursor = nil
    end

    return scene
end

return M
