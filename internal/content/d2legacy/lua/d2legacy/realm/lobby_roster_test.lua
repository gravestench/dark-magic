local test = require("d2legacy.tests/v1")

local function load_roster()
    local calls = { actors = {}, recipes = {} }

    local function node()
        return {
            visible = true,
            animation_count = 0,
            set_clip = function(self, ...)
                self.clip = { ... }
            end,
            set_visible = function(self, visible)
                self.visible = visible
            end,
            set_cof_animation = function(self, ...)
                self.animation_count = self.animation_count + 1
                self.animation = { ... }
                if self.animation[1] == "Wide.cof" then
                    return 8, {}, 300, 100, 150, 90
                end
                return 8, {}, 120, 200, 60, 180
            end,
            set_scale = function(self, ...)
                self.scale = { ... }
            end,
            set_position = function(self, ...)
                self.position = { ... }
            end,
        }
    end

    test.mock_module("engine.render/v1", {
        assets_available = function()
            return true
        end,
        create = function()
            local actor = node()
            calls.actors[#calls.actors + 1] = actor
            return actor
        end,
    }, { "assets_available", "create" })
    test.mock_module("d2legacy.realm.common", {
        label = function()
            return node()
        end,
        set_label = function(label, value)
            label.text = value
        end,
    }, { "label", "set_label" })
    test.mock_module("d2legacy.characters.composite", {
        recipe = function(character)
            calls.recipes[#calls.recipes + 1] = character
            return {
                cof = character.class .. ".cof",
                palette = "units.dat",
                direction = 0,
                components = {},
                rate = 10,
            }
        end,
    }, { "recipe" })

    return require("d2legacy.realm.lobby_roster"), calls
end

return test.suite({
    name = "Realm live lobby roster",
    profile = "module",
    tier = "fast",
    cases = {
        test.case("renders_only_current_channel_members_as_character_composites", function(t)
            t:run(function()
                local module, calls = load_roster()
                local roster = module.create({})
                module.update(roster, {
                    {
                        member_id = "alice-session",
                        character = { character_id = "amazon", name = "Alice", class = "Amazon", level = 12 },
                    },
                    {
                        member_id = "bob-session",
                        character = {
                            character_id = "barbarian",
                            name = "Bob",
                            class = "Barbarian",
                            level = 8,
                            title = "Slayer",
                        },
                    },
                })

                test.expect(calls.recipes):has_length(2)
                test.expect(calls.recipes[1].character_id):equals("amazon")
                test.expect(roster.slots[1].actor.visible):equals(true)
                test.expect(roster.slots[1].label.text):equals("[white]Alice")
                test.expect(roster.slots[2].label.text):equals("[white]Bob\n[green]Slayer")
                test.expect(roster.slots[3].actor.visible):equals(false)
                test.expect(roster.slots[1].actor.scale[1]):equals(0.45)
                test.expect(roster.slots[1].actor.position[2]):equals(549)
            end)
        end),
        test.case("bottom_aligns_a_width_limited_composite_above_its_name", function(t)
            t:run(function()
                local module = load_roster()
                local roster = module.create({})
                module.update(roster, {
                    {
                        member_id = "wide-session",
                        character = { character_id = "wide", name = "Wide", class = "Wide", level = 1 },
                    },
                })

                test.expect(roster.slots[1].actor.scale[1]):equals(0.3)
                test.expect(roster.slots[1].actor.position[2]):equals(555)
            end)
        end),
        test.case("does_not_restart_stable_animations_and_removes_departed_members", function(t)
            t:run(function()
                local module, calls = load_roster()
                local roster = module.create({})
                local alice = {
                    member_id = "alice-session",
                    character = { character_id = "amazon", name = "Alice", class = "Amazon", level = 12 },
                }
                module.update(roster, { alice })
                module.update(roster, { alice })
                test.expect(calls.recipes):has_length(1)
                test.expect(roster.slots[1].actor.animation_count):equals(1)

                module.update(roster, {})
                test.expect(roster.slots[1].actor.visible):equals(false)
                test.expect(roster.slots[1].label.visible):equals(false)
                test.expect(roster.slots[1].presence_key):equals(nil)
            end)
        end),
    },
})
