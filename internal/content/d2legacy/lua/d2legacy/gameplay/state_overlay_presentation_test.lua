local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("state_and_overlay_records_select_shared_recipes", {
            test.run(function()
                local adapter = require("d2legacy.gameplay.state_overlay_presentation")
                local overlay_path = table.concat({ "data", "global", "overlays" }, "/") .. "/"
                local resolved = adapter.from_rows({
                    {
                        state = "syntheticcurse",
                        overlay1 = "curse_loop",
                        overlay2 = "curse_back",
                        castoverlay = "curse_cast",
                        removerlay = "curse_remove",
                        onsound = "curse_on",
                        offsound = "curse_off",
                    },
                }, {
                    {
                        overlay = "curse_loop",
                        Filename = "Loop",
                        Frames = "24",
                        AnimRate = "12",
                        NumDirections = "1",
                        Trans = "3",
                        Xoffset = "2",
                        Yoffset = "-4",
                        Height1 = "14",
                        Height2 = "0",
                        Height3 = "-14",
                        Height4 = "-60",
                        InitRadius = "1",
                        Radius = "6",
                        Red = "255",
                        Green = "64",
                        Blue = "64",
                    },
                    { overlay = "curse_back", Filename = "Back", Frames = "8", AnimRate = "16", Trans = "0" },
                    { overlay = "curse_cast", Filename = "Cast", Frames = "10", AnimRate = "20", Trans = "3" },
                    { overlay = "curse_remove", Filename = "Remove", Frames = "6", AnimRate = "12", Trans = "3" },
                }, "syntheticcurse")
                test.assert(#resolved.active == 2, [=[#resolved.active == 2]=])
                test.assert(
                    resolved.active[1].path == overlay_path .. "Back.dcc"
                        and resolved.active[1].layer == "back"
                        and resolved.active[1].loop,
                    [=[overlay2 resolves as a looping back layer]=]
                )
                test.assert(
                    resolved.active[2].path == overlay_path .. "Loop.dcc"
                        and resolved.active[2].blend == "screen"
                        and resolved.active[2].offset_x == 2
                        and resolved.active[2].offset_y == -4
                        and adapter.height_offset(resolved.active[2], 1) == 14
                        and adapter.height_offset(resolved.active[2], 2) == 0
                        and adapter.height_offset(resolved.active[2], 3) == -14
                        and adapter.height_offset(resolved.active[2], 4) == -60
                        and adapter.height_offset(resolved.active[2], 0) == 0
                        and resolved.active[2].light.initial_radius == 1
                        and resolved.active[2].light.radius == 6
                        and resolved.active[2].light.red == 255
                        and resolved.active[2].light.green == 64
                        and resolved.active[2].light.blue == 64
                        and resolved.active[2].duration_seconds == 2,
                    [=[overlay1 preserves its authored attachment and light recipe]=]
                )
                test.assert(
                    resolved.applied.path == overlay_path .. "Cast.dcc"
                        and not resolved.applied.loop
                        and resolved.removed.path == overlay_path .. "Remove.dcc"
                        and resolved.applied_sound == "curse_on"
                        and resolved.removed_sound == "curse_off",
                    [=[cast/removal overlays remain one-shot recipes]=]
                )
            end),
        }),
        test.case("states_without_overlays_remain_semantic_only", {
            test.run(function()
                local adapter = require("d2legacy.gameplay.state_overlay_presentation")
                local resolved = adapter.from_rows({ { state = "freeze" } }, {}, "freeze")
                test.assert(resolved and #resolved.active == 0, [=[resolved and #resolved.active == 0]=])
                test.assert(resolved.applied == nil and resolved.removed == nil, [=[no overlay recipe is invented]=])
            end),
        }),
    },
})
