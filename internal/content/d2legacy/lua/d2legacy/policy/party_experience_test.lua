local test = require("d2legacy.tests/v1")

local function player(name, level, x, y)
    local ecs = require("engine.ecs/v1")
    return ecs.create({
        ["d2legacy.player.identity"] = {
            character_id = "character-" .. name,
            player = name,
            name = name,
            class = "Amazon",
        },
        ["d2legacy.player.progress"] = { level = level, experience = 0 },
        ["d2legacy.player.vitals"] = { health = 50, max_health = 50 },
        ["d2legacy.world.position"] = { x = x, y = y },
        ["d2legacy.world.location"] = { act = 1, level_id = 2 },
    })
end

local function victim(x, y)
    local ecs = require("engine.ecs/v1")
    local entity = ecs.create({ ["d2legacy.world.position"] = { x = x, y = y } })
    return ecs.get(entity, "d2legacy.world.position")
end

local function join(host, members)
    local party = require("d2legacy.policy.party")
    for _, member in ipairs(members) do
        party.invite(host, member, 1)
        party.accept(member, host)
    end
end

local function shares(credited, position, experience)
    local ecs = require("engine.ecs/v1")
    return require("d2legacy.policy.party_experience").shares(
        ecs.query({ all = { "d2legacy.player.identity" } }),
        credited,
        position,
        experience
    )
end

local function player_id(entity)
    return require("engine.ecs/v1").get(entity, "d2legacy.player.identity"):get("player")
end

return test.suite({
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("three_near_members_share_the_provisional_party_pool", {
            test.run(function()
                local credited = player("alice", 10, 100, 100)
                local second = player("bob", 10, 101, 100)
                local third = player("carol", 10, 102, 100)
                join("alice", { "bob", "carol" })

                local result = shares(credited, victim(100, 100), 1000)
                test.expect(#result):equals(3)
                for index, entity in ipairs({ credited, second, third }) do
                    test.expect(player_id(result[index].entity)):equals(player_id(entity))
                    test.expect(result[index].raw):equals(565)
                end
            end),
        }),
        test.case("victim_distance_boundary_changes_a_two_member_award", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local credited = player("alice", 10, 100, 100)
                local second = player("bob", 10, 180, 100)
                join("alice", { "bob" })

                local on_boundary = shares(credited, victim(100, 100), 1000)
                test.expect(#on_boundary):equals(2)
                test.expect(on_boundary[1].raw):equals(673)
                test.expect(player_id(on_boundary[2].entity)):equals(player_id(second))
                test.expect(on_boundary[2].raw):equals(673)

                ecs.get(second, "d2legacy.world.position"):set("x", 181)
                local past_boundary = shares(credited, victim(100, 100), 1000)
                test.expect(#past_boundary):equals(1)
                test.expect(player_id(past_boundary[1].entity)):equals(player_id(credited))
                test.expect(past_boundary[1].raw):equals(1000)
            end),
        }),
        test.case("sole_near_member_still_awards_the_out_of_range_credited_player", {
            test.run(function()
                local credited = player("alice", 10, 200, 100)
                player("bob", 10, 100, 100)
                join("alice", { "bob" })

                local result = shares(credited, victim(100, 100), 1000)
                test.expect(#result):equals(1)
                test.expect(player_id(result[1].entity)):equals(player_id(credited))
                test.expect(result[1].raw):equals(1000)
            end),
        }),
        test.case("unequal_levels_expose_provisional_lua_division", {
            test.run(function()
                local credited = player("alice", 71, 100, 100)
                player("bob", 50, 101, 100)
                player("carol", 20, 102, 100)
                join("alice", { "bob", "carol" })

                -- PXL-R04b's original float environment is unresolved. Lua
                -- double yields 3,457,134 here; stored float32 yields 3,457,133.
                local result = shares(credited, victim(100, 100), 4049741)
                test.expect(result[1].raw):equals(3457134)
            end),
        }),
        test.case("party_pool_uses_verified_1_13c_signed_32_bit_wrapping", {
            test.run(function()
                local credited = player("alice", 10, 100, 100)
                local members = { "bob", "carol", "dave", "erin", "frank", "grace", "hank" }
                for _, member in ipairs(members) do
                    player(member, 10, 100, 100)
                end
                join("alice", members)

                -- PXL-R04a wraps both products before truncating by 256.
                -- X=4,000,000 and n=8 produce pool -3,042,841, hence this
                -- ordinary finite raw share under the provisional quotient.
                local result = shares(credited, victim(100, 100), 4000000)
                test.expect(#result):equals(8)
                for _, entry in ipairs(result) do
                    test.expect(entry.raw):equals(-380355)
                end
            end),
        }),
        test.case("missing_member_state_and_death_exclude_party_awards", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local credited = player("alice", 10, 100, 100)
                local no_progress = player("bob", 10, 101, 100)
                local no_position = player("carol", 10, 102, 100)
                local no_location = player("dave", 10, 103, 100)
                local no_identity = player("erin", 10, 104, 100)
                local dead_health = player("frank", 10, 105, 100)
                local dead_marker = player("grace", 10, 106, 100)
                join("alice", { "bob", "carol", "dave", "erin", "frank", "grace" })
                ecs.remove(no_progress, "d2legacy.player.progress")
                ecs.remove(no_position, "d2legacy.world.position")
                ecs.remove(no_location, "d2legacy.world.location")
                ecs.remove(no_identity, "d2legacy.player.identity")
                ecs.get(dead_health, "d2legacy.player.vitals"):set("health", 0)
                ecs.set(dead_marker, "d2legacy.player.death", {
                    tick = 1,
                    killer_id = "monster:fixture",
                    credited_id = "monster:fixture",
                    hardcore = false,
                })

                local result = shares(credited, victim(100, 100), 1000)
                test.expect(#result):equals(1)
                test.expect(player_id(result[1].entity)):equals("alice")
                test.expect(result[1].raw):equals(1000)
            end),
        }),
        test.case("other_level_and_far_members_do_not_join_the_xp_pool", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local credited = player("alice", 10, 100, 100)
                player("bob", 10, 181, 100)
                local other_level = player("carol", 10, 101, 100)
                join("alice", { "bob", "carol" })
                ecs.get(other_level, "d2legacy.world.location"):set("level_id", 3)
                local result = shares(credited, victim(100, 100), 1000)
                test.expect(#result):equals(1)
                test.expect(player_id(result[1].entity)):equals("alice")
                test.expect(result[1].raw):equals(1000)
            end),
        }),
        test.case("zero_x_of_either_operand_rejects_party_member", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local credited = player("alice", 10, 2, 100)
                local second = player("bob", 10, 0, 100)
                join("alice", { "bob" })
                local member_zero = shares(credited, victim(1, 100), 1000)
                test.expect(#member_zero):equals(1)
                test.expect(player_id(member_zero[1].entity)):equals("alice")

                ecs.get(second, "d2legacy.world.position"):set("x", 1)
                local victim_zero = shares(credited, victim(0, 100), 1000)
                test.expect(#victim_zero):equals(0)
            end),
        }),
        test.case("fractional_world_x_is_a_provisional_dark_magic_coordinate_rule", {
            test.run(function()
                local credited = player("alice", 10, 0.5, 100)
                player("bob", 10, 1, 100)
                join("alice", { "bob" })
                -- The original dynamic getter's integer X for 0.5 would be
                -- zero. This synthetic f64-world rule deliberately differs.
                local result = shares(credited, victim(0.5, 100), 1000)
                test.expect(#result):equals(2)
                test.expect(result[1].raw):equals(673)
            end),
        }),
        test.case("missing_context_falls_back_to_solo_but_empty_party_awards_nobody", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local credited = player("alice", 10, 100, 100)
                local second = player("bob", 10, 101, 100)
                local position = victim(100, 100)
                test.expect(shares(credited, position, 1000)):is_nil()
                join("alice", { "bob" })
                test.expect(shares(credited, nil, 1000)):is_nil()

                ecs.remove(credited, "d2legacy.world.position")
                test.expect(shares(credited, position, 1000)):is_nil()
                ecs.set(credited, "d2legacy.world.position", { x = 100, y = 100 })
                ecs.remove(credited, "d2legacy.world.location")
                test.expect(shares(credited, position, 1000)):is_nil()
                ecs.set(credited, "d2legacy.world.location", { act = 1, level_id = 2 })

                ecs.get(credited, "d2legacy.player.vitals"):set("health", 0)
                ecs.get(second, "d2legacy.player.vitals"):set("health", 0)
                local nobody = shares(credited, position, 1000)
                test.expect(#nobody):equals(0)
            end),
        }),
        test.case("current_membership_and_nonpositive_level_sum_control_awards", {
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local credited = player("alice", 0, 100, 100)
                local second = player("bob", 0, 101, 100)
                local third = player("carol", 0, 102, 100)
                join("alice", { "bob", "carol" })
                local position = victim(100, 100)
                test.expect(#shares(credited, position, 1000)):equals(0)

                ecs.get(credited, "d2legacy.player.progress"):set("level", 10)
                ecs.get(second, "d2legacy.player.progress"):set("level", 10)
                ecs.get(third, "d2legacy.player.progress"):set("level", 10)
                require("d2legacy.policy.party").leave("bob")
                local current = shares(credited, position, 1000)
                test.expect(#current):equals(2)
                test.expect(player_id(current[1].entity)):equals("alice")
                test.expect(player_id(current[2].entity)):equals("carol")
                test.expect(current[1].raw):equals(673)
            end),
        }),
        test.case("ninth_admitted_member_fails_before_any_award", {
            test.run(function()
                local credited = player("alice", 10, 100, 100)
                local names = { "bob", "carol", "dave", "erin", "frank", "grace", "hank" }
                for _, name in ipairs(names) do
                    player(name, 10, 100, 100)
                end
                join("alice", names)
                player("irene", 10, 100, 100)
                local state = require("engine.authority_state/v1")
                local current = require("d2legacy.policy.party").snapshot()
                local party_id = current.membership.alice
                current.membership.irene = party_id
                table.insert(current.parties[party_id].members, "irene")
                state.replace("d2legacy.party", "d2legacy.party/v1", current)

                local ok = pcall(shares, credited, victim(100, 100), 1000)
                test.expect(ok):equals(false)
            end),
        }),
    },
})
