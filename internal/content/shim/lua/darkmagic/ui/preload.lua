-- Small manifest-aware preload bundles for predictable scene transitions.
--
-- This module only describes CPU work. Go workers read MPQs and decode bitmap
-- data in the background; the renderer thread still creates every GPU texture.
local render = require("dm.render/v1")
local data = require("dm.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local preload = {}
local frontend_job
local character_interaction_job

local function add(requests, seen, key, request)
    if seen[key] then return end
    seen[key] = true
    table.insert(requests, request)
end

local function add_dc6(requests, seen, path, palette_name)
    if not path then return end
    add(requests, seen, "dc6\0" .. path .. "\0" .. tostring(palette_name), {
        kind = "dc6",
        path = path,
        palette = assert(manifest.palettes[palette_name], "unknown preload palette: " .. tostring(palette_name)),
    })
end

local function add_dc6_frame(requests, seen, path, palette_name, frame)
    if not path or frame == nil then return end
    add(requests, seen, table.concat({"dc6_frame", path, palette_name, frame}, "\0"), {
        kind = "dc6_frame", path = path, palette = assert(manifest.palettes[palette_name]),
        direction = 0, frame = frame,
    })
end

local function add_dc6_animation(requests, seen, path, palette_name)
    if not path then return end
    add(requests, seen, table.concat({"dc6_animation", path, palette_name}, "\0"), {
        kind = "dc6_animation", path = path, palette = assert(manifest.palettes[palette_name]),
        direction = 0, anchor = "offsets",
    })
end

local function add_dc6_composite(requests, seen, path, overlay, palette_name)
    if not path then return end
    if not overlay then return add_dc6_animation(requests, seen, path, palette_name) end
    add(requests, seen, table.concat({"dc6_composite", path, overlay, palette_name}, "\0"), {
        kind = "dc6_composite", path = path, overlay = overlay,
        palette = assert(manifest.palettes[palette_name]), direction = 0, anchor = "offsets",
    })
end

local function add_frontend_background(requests, seen, definition)
    local layout = manifest.layouts.frontend_tiles
    local frame = layout.first_frame
    for _ = 1, #layout.rows do
        for _ = 1, #layout.columns do
            add_dc6_frame(requests, seen, definition.background, definition.palette, frame)
            frame = frame + 1
        end
    end
end

local function add_control_assets(requests, seen, controls)
    for _, control in pairs(controls or {}) do
        local frames = control.up_frames or { control.up_frame }
        for _, frame in ipairs(frames) do add_dc6_frame(requests, seen, control.sheet, control.palette, frame) end
    end
end

local function add_character_states(requests, seen, create, states)
    for _, class in ipairs(create.classes) do
        for _, state in ipairs(states) do
            add_dc6_composite(
                requests, seen, class[state], class[state .. "_overlay"],
                create.class_palette
            )
        end
    end
end

-- Start warming both destinations reachable through Single Player. Calling
-- this more than once returns the original job instead of scheduling copies.
function preload.frontend()
    if not render.assets_available() then return nil end
    -- This immutable bundle is valid for the process lifetime. Requeueing it
    -- from main_menu after the startup gate completed needlessly revisited
    -- every decoded asset and every native texture cache entry.
    if frontend_job then return frontend_job end
    local requests, seen = {}, {}
    local title = manifest.screens.title
    local menu = manifest.screens.main_menu
    local select = manifest.screens.character_select
    local create = manifest.screens.character_create

    -- Queue every pre-gameworld screen while startup videos are playing. The
    -- order follows the common player path so a short/skipped movie still
    -- makes the next screen resident first.
    add_frontend_background(requests, seen, title)
    for _, side in ipairs({"left", "right"}) do
        add_dc6_composite(requests, seen, title.logo["black_" .. side], title.logo["fire_" .. side], title.logo.palette)
    end
    add_frontend_background(requests, seen, menu)
    for _, side in ipairs({"left", "right"}) do
        add_dc6_composite(requests, seen, menu.logo["black_" .. side], menu.logo["fire_" .. side], menu.logo.palette)
    end
    add_control_assets(requests, seen, menu.controls)

    -- Character creation is the first expensive interactive destination for a
    -- new profile. Queue its visible actors and pointer-triggered states before
    -- secondary menu pages so hover and selection cannot outrun the warmer.
    add_frontend_background(requests, seen, create)
    add_dc6_animation(requests, seen, create.campfire.sheet, create.campfire.palette)
    add_dc6(requests, seen, create.dialog.sheet, create.dialog.palette)
    add_control_assets(requests, seen, create.controls)
    add_control_assets(requests, seen, create.options)
    for _, class in ipairs(create.classes) do
        add_dc6_animation(requests, seen, class.unselected, create.class_palette)
    end
    add_character_states(requests, seen, create, {"hover", "forward"})
    add_character_states(requests, seen, create, {"selected", "back"})

    add_frontend_background(requests, seen, manifest.screens.tcpip)
    add_control_assets(requests, seen, manifest.screens.tcpip.controls)
    add_frontend_background(requests, seen, manifest.screens.credits)
    add(requests, seen, "cinematics-background", {
        kind = "dc6_combined", path = manifest.screens.cinematics.background,
        palette = assert(manifest.palettes[manifest.screens.cinematics.palette]), direction = 0,
    })
    add_dc6_animation(requests, seen, manifest.screens.game_loading.sheet, manifest.screens.game_loading.palette)
    add_dc6_animation(requests, seen, manifest.cursor.modes.default.sheet, manifest.cursor.palette)
    for _, mode in ipairs({"normal", "pressed", "hand"}) do
        local definition = manifest.cursor.modes[mode]
        add_dc6_frame(requests, seen, definition.sheet, manifest.cursor.palette, definition.frame)
    end

    add_frontend_background(requests, seen, select)
    for frame = 0, 1 do add_dc6_frame(requests, seen, select.selection, select.selection_palette, frame) end
    for _, frame in ipairs({select.scrollbar.up_frame, select.scrollbar.down_frame, select.scrollbar.thumb_frame}) do
        add_dc6_frame(requests, seen, select.scrollbar.sheet, select.scrollbar.palette, frame)
    end
    add_dc6(requests, seen, select.delete_dialog.sheet, select.delete_dialog.palette)
    add_control_assets(requests, seen, select.controls)

    -- Fonts are cached with their palette transform because applying PL2 data
    -- is part of CPU-side font construction.
    local styles = {
        "button_normal", "button_hover", "character_select_title",
        "character_select_metadata", "character_create_heading",
        "character_create_description", "character_create_option",
    }
    for _, style_name in ipairs(styles) do
        local style = assert(manifest.text_styles[style_name])
        local font = assert(manifest.fonts[style.font])
        local transform = style.transform and assert(manifest.palette_transforms[style.transform]) or ""
        local key = table.concat({"font", font.table, font.sheet, font.palette, transform}, "\0")
        if not seen[key] then
            seen[key] = true
            table.insert(requests, {
                kind = "font", table = font.table, sheet = font.sheet,
                palette = assert(manifest.palettes[font.palette]), transform = transform,
            })
        end
    end

    frontend_job = render.preload(requests)
    return frontend_job
end

-- Once the creation screen itself is visible, prepare interaction-only states
-- (hover, selection walks, and overlays) without delaying the transition.
function preload.character_create_interactions()
    if not render.assets_available() then return nil end
    -- frontend() owns these states now and starts them during startup videos.
    -- Reusing its job also prevents a completed bundle from being requeued when
    -- the character-creation scene is constructed.
    if frontend_job then return frontend_job end
    if character_interaction_job then
        local status = render.preload_status(character_interaction_job)
        if status and not status.done then return character_interaction_job end
    end
    local requests, seen = {}, {}
    local create = manifest.screens.character_create
    add_character_states(requests, seen, create, {"hover", "forward", "selected", "back"})
    character_interaction_job = render.preload(requests)
    return character_interaction_job
end

function preload.frontend_status()
    if not frontend_job then return nil end
    return render.preload_status(frontend_job)
end

return preload
