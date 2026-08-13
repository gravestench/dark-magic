local test = require("d2legacy.tests/v1")

local fixtures = require("d2legacy.tests.support.fixtures")
local monster = {
    spawn_id = "near-target",
    seed = 1,
    x = 11,
    y = 12,
    act = 1,
    level_id = 1,
    definition = {
        id = "target",
        base_id = "",
        graphics_id = "",
        name_key = "Target",
        ai = "target",
        token = "TT",
        weapon_class = "HTH",
        components = {},
        life_min = 4096,
        life_max = 4096,
        level = 1,
        defense = 0,
        attack_rating = 0,
        physical_min = 0,
        physical_max = 0,
        experience = 0,
        treasure_class = "",
        collider_radius = 1,
        select_radius = 1,
        velocity = 0,
        think_interval = 100,
        aggro_radius = 0,
        attack_range = 2,
    },
}

return test.suite({
    profile = "authority",
    tier = "fast",
    records = {
        ["data/global/excel/skills.txt"] = {
            {
                Id = "0",
                skill = "Attack",
                skilldesc = "attack",
                leftskill = "1",
                general = "1",
                passive = "0",
            },
            {
                Id = "36",
                skill = "Fire Bolt",
                srvmissile = "firebolt",
                skilldesc = "firebolt",
                leftskill = "1",
                general = "0",
                passive = "0",
                etype = "fire",
                interrupt = "1",
                srvstfunc = "",
                srvdofunc = "",
                mana = "5",
                manashift = "7",
                emin = "3",
                emax = "6",
                HitShift = "8",
            },
        },
        ["data/global/excel/skilldesc.txt"] = {
            { skilldesc = "attack", ListRow = "0", IconCel = "0" },
            { skilldesc = "firebolt", ListRow = "1", IconCel = "1" },
        },
    },
    cases = {
        test.case("approach_admits_attack_animation_when_target_is_already_in_melee_range", function(t)
            t:system_command("system.player.enter", fixtures.amazon_entry, { tick = 1, sequence = 1 })
            t:step()
            t:run(function()
                local ecs = require("engine.ecs/v1")
                local player = test.only_entity_with("d2legacy.player.identity")
                ecs.create({
                    ["d2legacy.player.learned_skill"] = {
                        owner = player,
                        skill_id = 0,
                        level = 1,
                        list_row = 0,
                        left_allowed = true,
                        right_allowed = true,
                    },
                })
                local assignment = ecs.get(player, "d2legacy.player.skill_assignment")
                assignment:set("left", 0)
                test.expect(assignment:get("left"), "assigned left skill"):equals(0)
                ecs.set(player, "d2legacy.combat.attack_approach", {
                    skill_id = 0,
                    target_id = "monster:near-target",
                    request_tick = 2,
                    target_x = 11,
                    target_y = 12,
                })
            end)
            t:system_command("system.monster.spawn", monster, { tick = 2, sequence = 2 })
            t:step()
            t:check(function()
                local ecs = require("engine.ecs/v1")
                local approaches = test.entities_with("d2legacy.combat.attack_approach")
                local animations = test.entities_with("d2legacy.combat.attack_animation")
                local target = ecs.query({all={"d2legacy.monster.identity","d2legacy.world.selectable"}})[1]
                local target_position = ecs.get(target, "d2legacy.world.position")
                local player = test.only_entity_with("d2legacy.player.identity")
                local player_position = ecs.get(player, "d2legacy.world.position")
                local profile = ecs.get(player, "d2legacy.combat.melee_profile")
                test.expect(#approaches, "attack_approach count after admission"):equals(0)
                test.expect(#animations, "attack_animation count after admission"):equals(1)
                test.expect(
                    target_position:get("x") .. "," .. target_position:get("y"),
                    "target position after admission"
                ):deep_equals("11,12")
                test.expect(
                    player_position:get("x") .. "," .. player_position:get("y"),
                    "player position after admission"
                ):deep_equals("10,12")
                test.expect(profile:get("range"), "melee range"):equals(2)
                local attack = ecs.get(animations[1], "d2legacy.combat.attack_animation")
                test.expect(attack:get("start_tick"), "start_tick"):is_near(2, 0)
                test.expect(attack:get("impact_tick"), "impact_tick"):is_near(5, 0)
                test.expect(attack:get("complete_tick"), "complete_tick"):is_near(10, 0)
            end)
            t:step(8)
            t:check(function()
                local events = test.events("d2legacy.combat.attack_animation_event")
                test.expect(#events, "animation event count"):equals(3)
                local kinds = {}
                for _, event in ipairs(events) do
                    kinds[event:get("kind")] = true
                end
                test.expect(kinds["attack_started"], "attack_started event"):is_true()
                test.expect(kinds["attack_impact"], "attack_impact event"):is_true()
                test.expect(kinds["attack_completed"], "attack_completed event"):is_true()
            end)
        end),
    },
})
