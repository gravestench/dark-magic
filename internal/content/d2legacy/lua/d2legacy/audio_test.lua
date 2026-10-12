local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("owns_legacy_sound_record_selection", {
            test.run(function()
                local played
                test.mock_module("engine.records/v1", {
                    load = function()
                        return {
                            {
                                Sound = "menu",
                                FileName = "ui/one.wav",
                                IsUI = "1",
                                ["Volume Min"] = "255",
                                ["Volume Max"] = "255",
                                ["Group Size"] = "2",
                                ["Group Weight"] = "1",
                            },
                            {
                                Sound = "menu_alt",
                                FileName = "ui/two.wav",
                                IsUI = "1",
                                ["Volume Min"] = "255",
                                ["Volume Max"] = "255",
                                ["Group Size"] = "0",
                                ["Group Weight"] = "3",
                            },
                        }
                    end,
                }, { "load" })
                test.mock_module("engine.audio/v1", {
                    exists = function(path)
                        return path == "data/global/sfx/ui/two.wav"
                    end,
                    play = function(path, options)
                        played = { path = path, options = options }
                        return played
                    end,
                }, { "exists", "play" })
                require("d2legacy.audio").play_record("menu", 2)
                test.assert(
                    played.path == "data/global/sfx/ui/two.wav"
                        and played.options.volume == 1
                        and played.options.bus == "ui"
                )
            end),
        }),
        test.case("redirected_group_uses_target_row_position", {
            test.run(function()
                local played
                test.mock_module("engine.records/v1", {
                    load = function()
                        return {
                            {
                                Sound = "redirect_source", Redirect = "2", FileName = "source.wav",
                                ["Group Weight"] = "9",
                            },
                            { Sound = "unrelated", FileName = "neighbor.wav", ["Group Weight"] = "9" },
                            {
                                Sound = "target", FileName = "target-first.wav",
                                ["Group Size"] = "2", ["Group Weight"] = "1",
                            },
                            { Sound = "target_alt", FileName = "target-second.wav", ["Group Weight"] = "3" },
                            { Sound = "invalid_redirect", Redirect = "99", FileName = "fallback.wav" },
                        }
                    end,
                }, { "load" })
                test.mock_module("engine.audio/v1", {
                    exists = function() return true end,
                    play = function(path)
                        played = path
                        return path
                    end,
                }, { "exists", "play" })
                local audio = require("d2legacy.audio")
                audio.play_record("redirect_source", 2)
                test.expect(played):equals("target-second.wav")
                audio.play_record("invalid_redirect", 0)
                test.expect(played):equals("fallback.wav")
            end),
        }),
    },
})
