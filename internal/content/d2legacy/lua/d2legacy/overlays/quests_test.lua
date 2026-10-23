local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("owner_progress_refreshes_while_open_and_keeps_unknown_distinct", {
            test.run(function()
                local progress
                local latest, destroyed, z = "", 0, 0
                local function component(values)
                    return {
                        get = function(_, key)
                            return values[key]
                        end,
                    }
                end
                local function node()
                    return {
                        set_dc6 = function()
                            return 256, 256
                        end,
                        set_position = function() end,
                        set_z = function(_, value)
                            z = value
                        end,
                        destroy = function()
                            destroyed = destroyed + 1
                        end,
                    }
                end
                test.mock_module("engine.ecs/v1", {
                    query = function()
                        return { "peer", "self" }
                    end,
                    get = function(entity, name)
                        if name == "d2legacy.player.identity" then
                            return component({ player = entity == "self" and "alice" or "bob" })
                        end
                        if entity == "self" and progress then
                            return component(progress)
                        end
                        return component({ quest_id = 1, completed = true, rewarded = true })
                    end,
                })
                test.mock_module("engine.network/v1", {
                    status = function()
                        return { player_id = "alice" }
                    end,
                })
                test.mock_module("engine.render/v1", {
                    create = node,
                    assets_available = function()
                        return true
                    end,
                })
                test.mock_module("engine.input/v1", {
                    pressed = function()
                        return false
                    end,
                })
                test.mock_module("engine.scene/v1", { toggle_overlay = function() end })
                test.mock_module("engine.locale/v1", {
                    text = function(key)
                        return key
                    end,
                })
                test.mock_module("engine.data/v1", {
                    load_manifest = function()
                        return {
                            palettes = { fixture = "fixture" },
                            sounds = { button = "fixture" },
                            screens = {
                                quests = {
                                    panel = {
                                        sheet = "fixture",
                                        palette = "fixture",
                                        frames = { 0, 1, 2, 3 },
                                        x = 0,
                                        y = 0,
                                    },
                                    tabs = { sheet = "fixture", x = 0, y = 0, step = 1 },
                                    sockets = { positions = {} },
                                    unavailable = { x = 0, y = 0, width = 300 },
                                    close = { label = "close", x = 0, y = 0 },
                                },
                            },
                        }
                    end,
                })
                test.mock_module("d2legacy.quest_catalog/v1", {
                    quests = function()
                        return {}
                    end,
                })
                test.mock_module("d2legacy.ui.controls", {
                    new = function()
                        return { update = function() end }
                    end,
                })
                test.mock_module("d2legacy.ui.button", { create = function() end })
                test.mock_module("d2legacy.ui.text", {
                    create = function(_, _, value)
                        latest = value
                        return node()
                    end,
                })
                local overlay = require("d2legacy.overlays.quests")
                local instance = {}
                progress = { quest_id = 1, remaining = 0, unsettled_rooms = 2, population_available = true }
                overlay.enter(instance)
                test.expect(latest):equals("Den of Evil: 0 known living monsters; 2 chambers not yet surveyed.")
                test.expect(z):equals(1)
                overlay.update(instance)
                test.expect(destroyed):equals(0)
                progress.unsettled_rooms, progress.remaining = 0, 1
                overlay.update(instance)
                test.expect(latest):equals("Den of Evil: 1 monster remaining.")
                test.expect(destroyed):equals(1)
                progress.population_available = false
                overlay.update(instance)
                test.expect(latest):equals("Den of Evil: population information unavailable.")
                progress.completed, progress.remaining = true, 0
                overlay.update(instance)
                test.expect(latest):equals("Den of Evil: cleared. Return to Akara for one skill point.")
                progress.rewarded = true
                overlay.update(instance)
                test.expect(latest):equals("Den of Evil: rewarded. Akara granted one skill point.")
            end),
        }),
    },
})
