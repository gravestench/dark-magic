-- Small manifest-aware preload bundles for predictable scene transitions.
--
-- This module is a good example of a SAFE modding API boundary.
--
-- Lua does NOT get an MPQ reader, decoder worker, GPU texture, or background
-- thread. Instead Lua builds a plain list that DESCRIBES assets we expect to
-- need soon, then passes that list to `engine.render/v1`.
--
-- The engine is free to prepare CPU data in workers and upload GPU resources on
-- the graphics-owner thread. The mod only asks for useful work to happen.

local render = require("engine.render/v1")
local data = require("engine.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "d2legacy.presentation/v1"))
local preload = {}

-- Keep the returned job handles so repeated calls can reuse the same work.
-- These are cheap Lua/checked handles, not worker objects the mod owns.
local startup_job
local frontend_job
local character_interaction_job

-- Every helper below eventually funnels through this deduplicating function.
local function add(requests, seen, key, request)
    -- If the key is already in our set, another screen asked for the same exact
    -- prepared representation. One decode request is enough.
    if seen[key] then
        return
    end

    -- Lua tables can act as sets by using key -> true.
    seen[key] = true

    -- `table.insert` appends the plain request record to the ordered work list.
    table.insert(requests, request)
end

local function add_dc6(requests, seen, path, palette_name)
    if not path then
        return
    end

    -- `\0` is a NUL separator. Asset paths/palette names cannot accidentally
    -- blur together across it, so this makes a compact deterministic cache key.
    add(requests, seen, "dc6\0" .. path .. "\0" .. tostring(palette_name), {
        kind = "dc6",
        path = path,
        palette = assert(manifest.palettes[palette_name], "unknown preload palette: " .. tostring(palette_name)),
    })
end

local function add_dc6_frame(requests, seen, path, palette_name, frame)
    -- A nil path or frame means there is no valid frame request to schedule.
    if not path or frame == nil then
        return
    end

    add(requests, seen, table.concat({ "dc6_frame", path, palette_name, frame }, "\0"), {
        kind = "dc6_frame",
        path = path,
        palette = assert(manifest.palettes[palette_name]),
        direction = 0,
        frame = frame,
    })
end

local function add_dc6_animation(requests, seen, path, palette_name)
    if not path then
        return
    end

    add(requests, seen, table.concat({ "dc6_animation", path, palette_name }, "\0"), {
        kind = "dc6_animation",
        path = path,
        palette = assert(manifest.palettes[palette_name]),
        direction = 0,
        -- "offsets" asks the decoder/normalizer to preserve authored anchor
        -- offsets instead of pretending every cropped frame starts at 0,0.
        anchor = "offsets",
    })
end

local function add_dc6_composite(requests, seen, path, overlay, palette_name)
    if not path then
        return
    end

    -- A class state may have no overlay layer. In that simpler case, request the
    -- normal animation rather than inventing an empty second component.
    if not overlay then
        return add_dc6_animation(requests, seen, path, palette_name)
    end

    add(requests, seen, table.concat({ "dc6_composite", path, overlay, palette_name }, "\0"), {
        kind = "dc6_composite",
        path = path,
        overlay = overlay,
        palette = assert(manifest.palettes[palette_name]),
        direction = 0,
        anchor = "offsets",
    })
end

local function add_frontend_background(requests, seen, definition)
    local layout = manifest.layouts.frontend_tiles
    local frame = layout.first_frame

    -- We only care how MANY rows/columns exist here; `_` is the conventional
    -- Lua name for a loop variable whose actual value is intentionally ignored.
    for _ = 1, #layout.rows do
        for _ = 1, #layout.columns do
            add_dc6_frame(requests, seen, definition.background, definition.palette, frame)
            frame = frame + 1
        end
    end
end

local function add_control_assets(requests, seen, controls)
    -- `controls or {}` makes a missing optional control table behave like empty.
    for _, control in pairs(controls or {}) do
        -- Some buttons use several side-by-side frames; smaller buttons use one.
        local frames = control.up_frames or { control.up_frame }
        for _, frame in ipairs(frames) do
            add_dc6_frame(requests, seen, control.sheet, control.palette, frame)
        end
    end
end

local function add_character_states(requests, seen, create, states)
    -- Every class needs the same named set of actor states, so a two-level loop
    -- is clearer and safer than repeating seven classes times four states.
    for _, class in ipairs(create.classes) do
        for _, state in ipairs(states) do
            -- Lua string concatenation lets `state="hover"` select fields named
            -- `hover` and `hover_overlay` without hard-coding every pair.
            add_dc6_composite(requests, seen, class[state], class[state .. "_overlay"], create.class_palette)
        end
    end
end

local function add_logo(requests, seen, definition)
    for _, side in ipairs({ "left", "right" }) do
        add_dc6_composite(
            requests,
            seen,
            definition.logo["black_" .. side],
            definition.logo["fire_" .. side],
            definition.logo.palette
        )
    end
end

local function add_cursor(requests, seen)
    -- Pointer art is tiny, but a first-use decode stall follows the player's
    -- hand directly and is therefore unusually visible.
    add_dc6_animation(requests, seen, manifest.cursor.modes.default.sheet, manifest.cursor.palette)
    for _, mode in ipairs({ "normal", "pressed", "hand" }) do
        local definition = manifest.cursor.modes[mode]
        add_dc6_frame(requests, seen, definition.sheet, manifest.cursor.palette, definition.frame)
    end
end

local function add_text_styles(requests, seen, styles)
    -- Fonts are cached WITH their palette transform because applying PL2 data is
    -- part of the prepared CPU-side bitmap-font result.
    for _, style_name in ipairs(styles) do
        local style = assert(manifest.text_styles[style_name])
        local font = assert(manifest.fonts[style.font])
        local transform = style.transform and assert(manifest.palette_transforms[style.transform]) or ""
        local key = table.concat({ "font", font.table, font.sheet, font.palette, transform }, "\0")

        if not seen[key] then
            seen[key] = true
            table.insert(requests, {
                kind = "font",
                table = font.table,
                sheet = font.sheet,
                palette = assert(manifest.palettes[font.palette]),
                transform = transform,
            })
        end
    end
end

-- Warm only the title and main-menu working set before startup may continue.
-- Secondary destinations use later player think-time instead of inflating the
-- process before the first interactive screen appears.
function preload.startup()
    -- Headless tests deliberately run without Diablo assets. No assets means
    -- there is nothing useful to schedule, so a nil job is a valid result.
    if not render.assets_available() then
        return nil
    end

    if startup_job then
        return startup_job
    end
    local requests, seen = {}, {}
    local title = manifest.screens.title
    local menu = manifest.screens.main_menu

    add_frontend_background(requests, seen, title)
    add_logo(requests, seen, title)
    add_frontend_background(requests, seen, menu)
    add_logo(requests, seen, menu)
    add_control_assets(requests, seen, menu.controls)
    add_cursor(requests, seen)
    add_text_styles(requests, seen, { "button_normal", "button_hover", "frontend_version", "frontend_legal" })

    startup_job = render.preload(requests)
    return startup_job
end

-- Warm secondary frontend destinations after the main menu is already useful.
-- Character interaction animations are deliberately a third scene-local stage.
function preload.frontend()
    if not render.assets_available() then
        return nil
    end
    if frontend_job then
        return frontend_job
    end

    local requests, seen = {}, {}
    local select = manifest.screens.character_select
    local create = manifest.screens.character_create

    -- Prepare the character-creation shell and its seven visible idle actors,
    -- but do not decode four interaction-state families for every class here.
    add_frontend_background(requests, seen, create)
    add_dc6_animation(requests, seen, create.campfire.sheet, create.campfire.palette)
    add_dc6(requests, seen, create.dialog.sheet, create.dialog.palette)
    add_control_assets(requests, seen, create.controls)
    add_control_assets(requests, seen, create.options)

    for _, class in ipairs(create.classes) do
        add_dc6_animation(requests, seen, class.unselected, create.class_palette)
    end

    -- Less-common frontend destinations follow after the hot path.
    add_frontend_background(requests, seen, manifest.screens.tcpip)
    add_control_assets(requests, seen, manifest.screens.tcpip.controls)
    add_frontend_background(requests, seen, manifest.screens.credits)

    -- Cinematics use a multi-frame combined DC6 surface rather than the tiled
    -- frontend-background helper, so describe that representation directly.
    add(requests, seen, "cinematics-background", {
        kind = "dc6_combined",
        path = manifest.screens.cinematics.background,
        palette = assert(manifest.palettes[manifest.screens.cinematics.palette]),
        direction = 0,
    })

    add_dc6_animation(requests, seen, manifest.screens.game_loading.sheet, manifest.screens.game_loading.palette)

    -- Character selection is also part of the normal Single Player path.
    add_frontend_background(requests, seen, select)
    for frame = 0, 1 do
        add_dc6_frame(requests, seen, select.selection, select.selection_palette, frame)
    end
    for _, frame in ipairs({ select.scrollbar.up_frame, select.scrollbar.down_frame, select.scrollbar.thumb_frame }) do
        add_dc6_frame(requests, seen, select.scrollbar.sheet, select.scrollbar.palette, frame)
    end
    add_dc6(requests, seen, select.delete_dialog.sheet, select.delete_dialog.palette)
    add_control_assets(requests, seen, select.controls)

    add_text_styles(requests, seen, {
        "character_select_title",
        "character_select_metadata",
        "character_create_heading",
        "character_create_description",
        "character_create_option",
    })

    -- Hand the complete PLAIN-DATA work description to the engine capability.
    frontend_job = render.preload(requests)
    return frontend_job
end

function preload.character_create_interactions()
    if not render.assets_available() then
        return nil
    end
    if character_interaction_job then
        return character_interaction_job
    end

    local requests, seen = {}, {}
    local create = manifest.screens.character_create
    add_character_states(requests, seen, create, { "hover", "forward", "selected", "back" })
    character_interaction_job = render.preload(requests)
    return character_interaction_job
end

function preload.startup_status()
    if not startup_job then
        return nil
    end
    -- Again, callers get a STATUS SNAPSHOT through the capability rather than
    -- direct access to worker-thread state.
    return render.preload_status(startup_job)
end

return preload
