local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("owns_legacy_sound_record_selection", {
            test.run(function()
                local played
                package.loaded["engine.records/v1"] = {
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
                }
                package.loaded["engine.audio/v1"] = {
                    exists = function(path)
                        return path == "data/global/sfx/ui/two.wav"
                    end,
                    play = function(path, options)
                        played = { path = path, options = options }
                        return played
                    end,
                }
                require("d2legacy.audio").play_record("menu", 2)
                test.assert(
                    played.path == "data/global/sfx/ui/two.wav"
                        and played.options.volume == 1
                        and played.options.bus == "ui"
                )
            end),
        }),
    },
})
