-- Small manifest-aware preload bundles for predictable scene transitions.
--
-- This module only describes CPU work. Go workers read MPQs and decode bitmap
-- data in the background; the renderer thread still creates every GPU texture.
local render = require("dm.render/v1")
local data = require("dm.data/v1")

local manifest = assert(data.load_manifest("manifests/presentation.v1.json", "darkmagic.presentation/v1"))
local preload = {}
local frontend_job

local function add_dc6(requests, seen, path, palette_name)
    if not path or seen[path .. "\0" .. tostring(palette_name)] then return end
    seen[path .. "\0" .. tostring(palette_name)] = true
    table.insert(requests, {
        kind = "dc6",
        path = path,
        palette = assert(manifest.palettes[palette_name], "unknown preload palette: " .. tostring(palette_name)),
    })
end

local function add_control_assets(requests, seen, controls)
    for _, control in pairs(controls or {}) do
        add_dc6(requests, seen, control.sheet, control.palette)
    end
end

-- Start warming both destinations reachable through Single Player. Calling
-- this more than once returns the original job instead of scheduling copies.
function preload.frontend()
    if frontend_job or not render.assets_available() then return frontend_job end
    local requests, seen = {}, {}
    local select = manifest.screens.character_select
    local create = manifest.screens.character_create

    add_dc6(requests, seen, select.background, select.palette)
    add_dc6(requests, seen, select.selection, select.selection_palette)
    add_dc6(requests, seen, select.scrollbar.sheet, select.scrollbar.palette)
    add_dc6(requests, seen, select.delete_dialog.sheet, select.delete_dialog.palette)
    add_control_assets(requests, seen, select.controls)

    add_dc6(requests, seen, create.background, create.palette)
    add_dc6(requests, seen, create.campfire.sheet, create.campfire.palette)
    add_dc6(requests, seen, create.dialog.sheet, create.dialog.palette)
    add_control_assets(requests, seen, create.controls)
    add_control_assets(requests, seen, create.options)
    for _, class in ipairs(create.classes) do
        for _, field in ipairs({
            "unselected", "hover", "selected", "selected_overlay",
            "forward", "forward_overlay", "back", "back_overlay",
        }) do
            add_dc6(requests, seen, class[field], class.palette)
        end
    end

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

function preload.frontend_status()
    if not frontend_job then return nil end
    return render.preload_status(frontend_job)
end

return preload
