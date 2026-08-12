-- START HERE: this is the bundled d2legacy mod's presentation front door.
--
-- Dark Magic loads this file first. Its job is intentionally tiny: ask the
-- engine for a few versioned capabilities, load the presentation manifest,
-- register every scene name, and choose the first scene to show.
--
-- A useful rule while reading this mod:
--   require("engine.something/v1")  = engine/modding API capability
--   require("d2....")    = ordinary Lua from this example mod
--
-- Keeping that boundary visible is important. Lua gets useful operations; it
-- does not get raw renderer pointers, filesystem objects, or engine internals.

-- `local` means only this file can see this name. New mod code should prefer
-- locals so one file cannot accidentally overwrite another file's variables.
local render = require("engine.render/v1")
-- The scene capability owns the navigation stack: replacing root screens and
-- pushing/popping/toggling overlays.
local scenes = require("engine.scene/v1")
-- The data capability loads versioned, validated manifests from mod content.
local data = require("engine.data/v1")
-- This one is not an engine capability. It is another Lua file in this mod.
local registry = require("d2.bootstrap.scene_registry")

-- The manifest is data, not executable policy. `assert` makes a broken mod fail
-- here with a useful error instead of producing mysterious nil errors later.
-- The schema string is also versioned, so the loader knows what shape to expect.
local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2.presentation/v1"))

-- A Lua module returns one value. Here we return a component definition table.
-- Dark Magic reads this table and calls the lifecycle functions below.
return {
    -- Stable component identity. This is a name, not a filename dependency.
    id = "d2.boot",
    -- Version of this tiny boot-component contract.
    api = 1,

    -- `self` is this particular running instance. Values placed on `self` are
    -- instance state, so another instance does not have to share them.
    start = function(self)
        -- Create one retained render node owned by the boot component's scope.
        -- "transition" is the render layer. The returned value is a checked Lua
        -- handle; the native renderer resource behind it remains engine-owned.
        --
        -- This root is an empty hook. Screens hang their pictures below it.
        -- Keeping the hook here also makes ownership plain: boot creates it, so
        -- boot removes it when the component stops.
        self.root = render.create("transition")

        -- Registration teaches the scene manager every symbolic name it may
        -- open. It does NOT create every scene now. Think of it as filling in a
        -- phone book: "main_menu" -> that scene definition.
        registry.register_all(manifest)

        -- `replace` chooses the one root scene that should be active. The
        -- loading scene can then advance to title/menu scenes by the same API.
        scenes.replace("loading")
    end,

    stop = function(self)
        -- We do not manually free a native renderer object here. The checked
        -- handle belongs to this component's resource scope, so the engine owns
        -- native cleanup. Dropping our Lua reference also makes our intent clear.
        self.root = nil
    end,
}
