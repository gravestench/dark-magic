local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("owns_legacy_sound_record_selection", {
            test.run(function()
                local played
                test.mock_module("engine.records/v1", {
                    revision = function()
                        return nil
                    end,
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
                }, { "load", "revision" })
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
                    revision = function()
                        return nil
                    end,
                    load = function()
                        return {
                            {
                                Sound = "redirect_source",
                                Redirect = "2",
                                FileName = "source.wav",
                                ["Group Weight"] = "9",
                            },
                            { Sound = "unrelated", FileName = "neighbor.wav", ["Group Weight"] = "9" },
                            {
                                Sound = "target",
                                FileName = "target-first.wav",
                                ["Group Size"] = "2",
                                ["Group Weight"] = "1",
                            },
                            { Sound = "target_alt", FileName = "target-second.wav", ["Group Weight"] = "3" },
                            { Sound = "invalid_redirect", Redirect = "99", FileName = "fallback.wav" },
                        }
                    end,
                }, { "load", "revision" })
                test.mock_module("engine.audio/v1", {
                    exists = function()
                        return true
                    end,
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
        test.case("caches_names_until_record_revision_changes", {
            test.run(function()
                local fixture = {
                    rows = {
                        { Sound = "cue", FileName = "first.wav" },
                        { Sound = "CUE", FileName = "duplicate.wav" },
                        { Sound = "other", FileName = "other.wav" },
                    },
                    revision = "1",
                    loads = 0,
                    fail = false,
                }
                test.mock_module("engine.records/v1", {
                    revision = function()
                        return fixture.revision
                    end,
                    load = function()
                        fixture.loads = fixture.loads + 1
                        if fixture.fail then
                            return nil, "unavailable"
                        end
                        return fixture.rows
                    end,
                }, { "load", "revision" })
                test.mock_module("engine.audio/v1", {
                    exists = function()
                        return true
                    end,
                    play = function(path)
                        return path
                    end,
                }, { "exists", "play" })
                local audio = require("d2legacy.audio")
                test.expect(audio.play_record("cue")):equals("first.wav")
                test.expect(audio.play_record("CUE")):equals("first.wav")
                test.expect(audio.play_record("other")):equals("other.wav")
                test.expect(fixture.loads):equals(1)
                fixture.rows = { { Sound = "cue", FileName = "second.wav" } }
                fixture.revision = "2"
                test.expect(audio.play_record("cue")):equals("second.wav")
                test.expect(fixture.loads):equals(2)
                fixture.revision, fixture.fail = "3", true
                test.assert(not pcall(audio.play_record, "cue"))
                fixture.fail = false
                test.expect(audio.play_record("cue")):equals("second.wav")
                test.expect(fixture.loads):equals(4)
                fixture.rows, fixture.revision = {}, "4"
                local ok, err = pcall(audio.play_record, "cue")
                test.assert(not ok and err:find("unknown sound record: cue", 1, true))
                fixture.rows, fixture.revision = { { Sound = "cue", FileName = "uncached.wav" } }, nil
                test.expect(audio.play_record("cue")):equals("uncached.wav")
                test.expect(audio.play_record("cue")):equals("uncached.wav")
                test.expect(fixture.loads):equals(7)
            end),
        }),
    },
})
