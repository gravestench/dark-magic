local test = require("d2legacy.tests/v1")

local function node()
    return {
        tint = nil,
        set_tint = function(self, red, green, blue)
            self.tint = { red, green, blue }
        end,
    }
end

local function tint_is(actor, red, green, blue)
    test.assert(actor.tint ~= nil, "actor tint was not applied")
    test.expect(actor.tint[1]):equals(red)
    test.expect(actor.tint[2]):equals(green)
    test.expect(actor.tint[3]):equals(blue)
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("hero_remote_player_and_monster_use_green_then_reset_with_active_states", {
            test.run(function()
                local presentation = require("d2legacy.gameplay.state_tint_presentation")
                local hero, remote, monster = node(), node(), node()
                local actors = { ["1"] = hero, ["2"] = remote, ["3"] = monster }
                local states = {
                    { target_entity_id = 1, state_id = "poison", level_id = 2 },
                    { target_entity_id = 1, state_id = "might", level_id = 2, aura = true },
                    { target_entity_id = 2, state_id = "poison", level_id = 2 },
                    { target_entity_id = 3, state_id = "poison", level_id = 2 },
                }
                local resolve = function(id)
                    return { green = id == "poison" }
                end
                presentation.apply(actors, states, resolve, 2)
                tint_is(hero, 128, 255, 128)
                tint_is(remote, 128, 255, 128)
                tint_is(monster, 128, 255, 128)

                presentation.apply(actors, { states[2] }, resolve, 2)
                tint_is(hero, 255, 255, 255)
                tint_is(remote, 255, 255, 255)
                tint_is(monster, 255, 255, 255)

                local reused = node()
                actors["3"] = reused
                presentation.apply(actors, { states[4] }, resolve, 1)
                tint_is(reused, 255, 255, 255)
            end),
        }),
    },
})
