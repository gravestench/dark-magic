local test = require("d2legacy.tests/v1")

local jobs = {}
test.mock_module("engine.render/v1", {
    assets_available = function()
        return true
    end,
    preload = function(requests)
        jobs[#jobs + 1] = requests
        return #jobs
    end,
    preload_status = function(job)
        return { done = true, job = job }
    end,
}, { "assets_available", "preload", "preload_status" })

local function logo(prefix)
    return {
        palette = "units",
        black_left = prefix .. "-black-left",
        fire_left = prefix .. "-fire-left",
        black_right = prefix .. "-black-right",
        fire_right = prefix .. "-fire-right",
    }
end

local function screen(prefix)
    return {
        background = prefix .. "-background",
        palette = "sky",
        controls = {
            ok = { sheet = prefix .. "-button", palette = "units", up_frame = 0 },
        },
    }
end

local manifest = {
    palettes = { sky = "sky.pal", units = "units.pal", fechar = "fechar.pal" },
    palette_transforms = {},
    layouts = { frontend_tiles = { first_frame = 0, rows = { 1 }, columns = { 1 } } },
    fonts = { ui = { table = "font.tbl", sheet = "font.dc6", palette = "units" } },
    text_styles = {},
    cursor = {
        palette = "units",
        modes = {
            default = { sheet = "cursor-default" },
            normal = { sheet = "cursor", frame = 0 },
            pressed = { sheet = "cursor", frame = 1 },
            hand = { sheet = "cursor", frame = 2 },
        },
    },
    screens = {
        title = screen("title"),
        main_menu = screen("menu"),
        tcpip = screen("tcpip"),
        credits = screen("credits"),
        cinematics = { background = "cinematics-background", palette = "sky" },
        game_loading = { sheet = "game-loading", palette = "units" },
        character_select = {
            background = "select-background",
            palette = "sky",
            selection = "select-box",
            selection_palette = "sky",
            scrollbar = { sheet = "scrollbar", palette = "sky", up_frame = 0, down_frame = 1, thumb_frame = 2 },
            delete_dialog = { sheet = "delete-dialog", palette = "fechar" },
            controls = { ok = { sheet = "select-button", palette = "units", up_frame = 0 } },
        },
        character_create = {
            background = "create-background",
            palette = "sky",
            class_palette = "units",
            campfire = { sheet = "campfire", palette = "fechar" },
            dialog = { sheet = "create-dialog", palette = "fechar" },
            controls = { ok = { sheet = "create-button", palette = "units", up_frame = 0 } },
            options = {},
            classes = {
                {
                    unselected = "amazon-unselected",
                    hover = "amazon-hover",
                    forward = "amazon-forward",
                    selected = "amazon-selected",
                    back = "amazon-back",
                },
                {
                    unselected = "sorceress-unselected",
                    hover = "sorceress-hover",
                    forward = "sorceress-forward",
                    selected = "sorceress-selected",
                    back = "sorceress-back",
                },
            },
        },
    },
}
manifest.screens.title.logo = logo("title")
manifest.screens.main_menu.logo = logo("menu")
for _, name in ipairs({
    "button_normal",
    "button_hover",
    "frontend_version",
    "frontend_legal",
    "character_select_title",
    "character_select_metadata",
    "character_create_heading",
    "character_create_description",
    "character_create_option",
}) do
    manifest.text_styles[name] = { font = "ui" }
end

test.mock_module("engine.data/v1", {
    load_manifest = function()
        return manifest
    end,
}, { "load_manifest" })

local preload = require("d2legacy.ui.preload")

local function has_path(requests, path)
    for _, request in ipairs(requests) do
        if request.path == path or request.overlay == path then
            return true
        end
    end
    return false
end

return test.suite({
    name = "Frontend preload staging",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("stages_character_interactions_at_character_creation", function(t)
            t:run(function()
                local startup = preload.startup()
                test.expect(startup, "startup job"):equals(1)
                test.expect(preload.startup(), "reused startup job"):equals(startup)
                test.expect(#jobs, "jobs after startup"):equals(1)
                test.assert(has_path(jobs[1], "title-background"), "startup omits title")
                test.assert(has_path(jobs[1], "menu-background"), "startup omits main menu")
                test.assert(not has_path(jobs[1], "amazon-unselected"), "startup includes character actors")

                local frontend = preload.frontend()
                test.expect(frontend, "frontend job"):equals(2)
                test.expect(preload.frontend(), "reused frontend job"):equals(frontend)
                test.expect(#jobs, "jobs after frontend"):equals(2)
                test.assert(has_path(jobs[2], "amazon-unselected"), "frontend omits visible class idle")
                test.assert(not has_path(jobs[2], "amazon-hover"), "frontend includes interaction state")

                local interactions = preload.character_create_interactions()
                test.expect(interactions, "interaction job"):equals(3)
                test.expect(preload.character_create_interactions(), "reused interaction job"):equals(interactions)
                test.expect(#jobs, "all staged jobs"):equals(3)
                for _, state in ipairs({ "hover", "forward", "selected", "back" }) do
                    test.assert(has_path(jobs[3], "amazon-" .. state), "interaction job omits " .. state)
                end
                test.assert(not has_path(jobs[3], "amazon-unselected"), "interaction job repeats visible idle")
            end)
        end),
    },
})
