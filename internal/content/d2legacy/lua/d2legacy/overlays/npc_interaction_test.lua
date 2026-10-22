local test = require("d2legacy.tests/v1")

local function fixture(speech_failure)
    local state = { active = true, npc = "Fixture NPC", target_id = "fixture:npc", page = "menu", revision = 1 }
    local calls = { choices = {}, pops = 0, controls = {}, destroyed = 0, rendered_text = {} }
    local pressed = false
    local function node()
        return {
            fill_rect = function() end,
            set_position = function() end,
            set_z = function() end,
            destroy = function()
                calls.destroyed = calls.destroyed + 1
            end,
        }
    end
    test.mock_module("engine.render/v1", {
        create = node,
        assets_available = function()
            return true
        end,
    })
    test.mock_module("engine.input/v1", {
        pressed = function()
            return pressed
        end,
    })
    test.mock_module("engine.scene/v1", {
        pop = function()
            calls.pops = calls.pops + 1
        end,
    })
    test.mock_module("engine.locale/v1", {
        text = function(value)
            return value
        end,
    })
    test.mock_module("d2legacy.quest_catalog/v1", {
        dialog = function(line)
            calls.caption_reads = (calls.caption_reads or 0) + 1
            if speech_failure then
                test.expect(line):equals("missing-speech")
                if speech_failure == "error" then
                    error("fixture localization unavailable")
                end
                return nil
            end
            test.expect(line):equals("fixture:caption")
            return { text = "Synthetic caption" }
        end,
    })
    test.mock_module("engine.data/v1", {
        load_manifest = function()
            return { resolution = { width = 800, height = 600 } }
        end,
    })
    test.mock_module("d2legacy.ui.text", {
        set_font = function(_, _, value)
            calls.rendered_text[#calls.rendered_text + 1] = value
            return 100, 16
        end,
    })
    test.mock_module("d2legacy.ui.controls", {
        new = function()
            calls.controls = {}
            return {
                add = function(_, control)
                    calls.controls[control.id] = control
                end,
                update = function() end,
            }
        end,
    })
    test.mock_module("d2legacy.interactions.api", {
        dialogue = function()
            local snapshot = {}
            for key, value in pairs(state) do
                snapshot[key] = value
            end
            snapshot.choices = require("d2legacy.interactions.dialogue").choices("akara", state.page)
            return snapshot
        end,
        choose = function(snapshot, choice)
            calls.choices[#calls.choices + 1] =
                { choice = choice, revision = snapshot.revision, target = snapshot.target_id }
        end,
    })
    local overlay = require("d2legacy.overlays.npc_interaction")
    local instance = {}
    overlay.enter(instance)
    overlay.update(instance)
    return overlay, instance, state, calls, function(value)
        pressed = value
    end
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("unknown_connected_speech_displays_unavailable_text_and_keeps_back_control", function(t)
            t:run(function()
                local overlay, instance, state, calls = fixture("missing")
                state.page, state.revision, state.line = "text", 2, "missing-speech"
                overlay.update(instance)
                test.expect(calls.rendered_text[#calls.rendered_text - 1]):equals("d2legacy.npc.speech_unavailable")
                calls.controls.back.on_activate()
                test.expect(calls.choices[1].choice):equals("back")
                test.expect(calls.pops):equals(0)
            end)
        end),
        test.case("catalog_failure_keeps_escape_available", function(t)
            t:run(function()
                local overlay, instance, state, calls, cancel = fixture("error")
                state.page, state.revision, state.line = "text", 2, "missing-speech"
                overlay.update(instance)
                cancel(true)
                overlay.update(instance)
                test.expect(calls.caption_reads):equals(1)
                test.expect(calls.choices[1].choice):equals("back")
            end)
        end),
        test.case("menu_and_topic_controls_submit_choices_without_inventing_authority", function(t)
            t:run(function()
                local overlay, instance, state, calls = fixture()
                test.expect(calls.controls.trade):is_nil()
                calls.controls.talk.on_activate()
                calls.controls.talk.on_activate()
                test.expect(#calls.choices):equals(1)
                test.expect(state.page):equals("menu")
                state.page, state.revision = "topics", 2
                overlay.update(instance)
                test.expect(calls.controls.gossip ~= nil):is_true()
                calls.controls.gossip.on_activate()
                test.expect(calls.choices[2]):deep_equals({ choice = "gossip", revision = 2, target = "fixture:npc" })
                test.expect(calls.destroyed):equals(1)
            end)
        end),
        test.case("escape_dismisses_text_through_authority_and_close_waits_for_inactive_context", function(t)
            t:run(function()
                local overlay, instance, state, calls, cancel = fixture()
                state.page, state.revision, state.line = "text", 2, "fixture:caption"
                overlay.update(instance)
                overlay.update(instance)
                test.expect(calls.caption_reads):equals(1)
                cancel(true)
                overlay.update(instance, 0, false, false)
                test.expect(#calls.choices):equals(0)
                cancel(true)
                overlay.update(instance)
                test.expect(calls.choices[1].choice):equals("back")
                test.expect(calls.pops):equals(0)
                cancel(false)
                state.page, state.revision = "menu", 3
                overlay.update(instance)
                calls.controls.cancel.on_activate()
                test.expect(calls.pops):equals(0)
                state.active = false
                overlay.update(instance, 0, false, false)
                test.expect(calls.pops):equals(0)
                overlay.update(instance)
                test.expect(calls.pops):equals(1)
            end)
        end),
    },
})
