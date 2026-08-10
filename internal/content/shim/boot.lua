local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
local data = require("dm.data/v1")
local registry = require("darkmagic.bootstrap.scene_registry")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))

return {
    id = "darkmagic.boot",
    api = 1,

    start = function(self)
        -- This root is an empty hook. Screens hang their pictures below it.
        -- Keeping the hook here also makes ownership plain: boot creates it, so
        -- boot removes it when the component stops.
        self.root = render.create("transition")

        -- Registration teaches the scene manager every name it may open. It does
        -- not open all those screens; the final replace opens only the first one.
        registry.register_all(manifest)
        scenes.replace("loading")
    end,

    stop = function(self)
        self.root = nil
    end,
}
