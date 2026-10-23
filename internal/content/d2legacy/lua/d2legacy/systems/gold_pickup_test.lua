local test = require("d2legacy.tests/v1")
local fixtures = require("d2legacy.tests.support.fixtures")

local function enter(player, sequence, overrides, tick)
    local values = overrides or {}
    values.player = player
    values.character_id = player
    return test.submit_system({
        tick = tick or 1,
        sequence = sequence,
        kind = "system.player.enter",
        payload = fixtures.player_entry(values),
    })
end

local function party_command(tick, player, sequence, kind, payload)
    return test.submit({ tick = tick, sequence = sequence, player = player, kind = kind, payload = payload })
end

local function player_entity(owner)
    local ecs = require("engine.ecs/v1")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.player.identity" } })) do
        if ecs.get(entity, "d2legacy.player.identity"):get("player") == owner then
            return entity
        end
    end
end

local function set_gold(owner, amount)
    local ecs = require("engine.ecs/v1")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.items.layout" } })) do
        local layout = ecs.get(entity, "d2legacy.items.layout")
        if layout:get("owner") == owner then
            layout:set("carried_gold", amount)
            return
        end
    end
    error("missing gold layout for " .. owner)
end

local function open(tick, player, sequence)
    return test.submit({
        tick = tick,
        sequence = sequence,
        player = player,
        kind = "interaction.open",
        payload = { target = "gold:test" },
    })
end

local function gold(owner)
    local ecs = require("engine.ecs/v1")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.items.layout" } })) do
        local layout = ecs.get(entity, "d2legacy.items.layout")
        if layout:get("owner") == owner then
            return layout:get("carried_gold")
        end
    end
end

local function pile()
    local ecs = require("engine.ecs/v1")
    for _, entity in ipairs(ecs.query({ all = { "d2legacy.interaction.target" } })) do
        if ecs.get(entity, "d2legacy.interaction.target"):get("id") == "gold:test" then
            return entity, ecs.get(entity, "d2legacy.gold.pile")
        end
    end
end

return test.suite({
    profile = "authority",
    tier = "fast",
    initial_data = {
        ["d2legacy.interactions"] = {
            owner = "alice",
            targets = {
                {
                    id = "gold:test",
                    x = 12,
                    y = 12,
                    radius = 4,
                    act = 1,
                    level_id = 1,
                    selectable = true,
                    gold_amount = 1000,
                },
            },
        },
    },
    cases = {
        test.case("solo_pickup_changes_authoritative_gold_and_consumes_pile", {
            enter("alice", 1),
            test.step(1),
            test.run(function()
                local entity, amount = pile()
                test.expect(entity ~= nil):is_true()
                test.expect(amount:get("amount")):equals(1000)
                test.expect(gold("alice")):equals(0)
                test.expect(require("d2legacy.gameplay.targeting").selectable_at(12, 12, 1).id):equals("gold:test")
            end),
            open(2, "alice", 1),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity, amount = pile()
                test.expect(gold("alice")):equals(1000)
                test.expect(amount:get("amount")):equals(0)
                test.expect(ecs.get(entity, "d2legacy.world.inactive") ~= nil):is_true()
                test.expect(require("d2legacy.gameplay.targeting").selectable_at(12, 12, 1)):is_nil()
                test.expect(require("d2legacy.items.snapshot").for_owner("alice").carried_gold):equals(1000)
            end),
            open(3, "alice", 2),
            test.step(1),
            test.run(function()
                test.expect(gold("alice")):equals(1000)
            end),
        }),
        test.case("distant_party_members_receive_gold_without_a_range_limit", {
            enter("alice", 1),
            enter("bob", 2, { x = 80 }),
            enter("charlie", 3, { x = 90 }),
            test.step(1),
            party_command(2, "alice", 1, "party.invite", { target = "bob" }),
            party_command(2, "alice", 2, "party.invite", { target = "charlie" }),
            test.step(1),
            party_command(3, "bob", 1, "party.accept", { inviter = "alice" }),
            party_command(3, "charlie", 1, "party.accept", { inviter = "alice" }),
            test.step(1),
            open(4, "alice", 3),
            test.step(1),
            test.run(function()
                test.expect(gold("alice")):equals(334)
                test.expect(gold("bob")):equals(333)
                test.expect(gold("charlie")):equals(333)
                test.expect(pile() ~= nil):is_true()
            end),
        }),
        test.case("dead_member_does_not_count_but_receives_a_share", {
            enter("alice", 1),
            enter("bob", 2),
            enter("charlie", 3),
            test.step(1),
            party_command(2, "alice", 1, "party.invite", { target = "bob" }),
            party_command(2, "alice", 2, "party.invite", { target = "charlie" }),
            test.step(1),
            party_command(3, "bob", 1, "party.accept", { inviter = "alice" }),
            party_command(3, "charlie", 1, "party.accept", { inviter = "alice" }),
            test.step(1),
            test.run(function()
                require("engine.ecs/v1").get(player_entity("bob"), "d2legacy.player.vitals"):set("health", 0)
            end),
            open(4, "alice", 3),
            test.step(1),
            test.run(function()
                test.expect(gold("alice")):equals(0)
                test.expect(gold("bob")):equals(500)
                test.expect(gold("charlie")):equals(500)
            end),
        }),
        test.case("two_dead_recipients_cannot_create_gold", {
            enter("alice", 1),
            enter("bob", 2),
            enter("charlie", 3),
            enter("donna", 4),
            test.step(1),
            party_command(2, "alice", 1, "party.invite", { target = "bob" }),
            party_command(2, "alice", 2, "party.invite", { target = "charlie" }),
            party_command(2, "alice", 3, "party.invite", { target = "donna" }),
            test.step(1),
            party_command(3, "bob", 1, "party.accept", { inviter = "alice" }),
            party_command(3, "charlie", 1, "party.accept", { inviter = "alice" }),
            party_command(3, "donna", 1, "party.accept", { inviter = "alice" }),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                ecs.get(player_entity("charlie"), "d2legacy.player.vitals"):set("health", 0)
                ecs.get(player_entity("donna"), "d2legacy.player.vitals"):set("health", 0)
            end),
            open(4, "alice", 4),
            test.step(1),
            test.run(function()
                local _, amount = pile()
                test.expect(gold("alice")):equals(0)
                test.expect(gold("bob")):equals(500)
                test.expect(gold("charlie")):equals(500)
                test.expect(gold("donna")):equals(0)
                test.expect(amount:get("amount")):equals(0)
            end),
        }),
        test.case("other_level_member_does_not_split_gold", {
            enter("alice", 1),
            enter("bob", 2),
            test.step(1),
            party_command(2, "alice", 1, "party.invite", { target = "bob" }),
            test.step(1),
            party_command(3, "bob", 1, "party.accept", { inviter = "alice" }),
            test.step(1),
            test.run(function()
                require("engine.ecs/v1").get(player_entity("bob"), "d2legacy.world.location"):set("level_id", 2)
            end),
            open(4, "alice", 2),
            test.step(1),
            test.run(function()
                test.expect(gold("alice")):equals(1000)
                test.expect(gold("bob")):equals(0)
            end),
        }),
        test.case("rejoined_party_uses_current_membership", {
            enter("alice", 1),
            enter("bob", 2),
            test.step(1),
            party_command(2, "alice", 1, "party.invite", { target = "bob" }),
            test.step(1),
            party_command(3, "bob", 1, "party.accept", { inviter = "alice" }),
            test.step(1),
            party_command(4, "bob", 2, "party.leave", {}),
            test.step(1),
            party_command(5, "alice", 2, "party.invite", { target = "bob" }),
            test.step(1),
            party_command(6, "bob", 3, "party.accept", { inviter = "alice" }),
            test.step(1),
            open(7, "alice", 3),
            test.step(1),
            test.run(function()
                test.expect(gold("alice")):equals(500)
                test.expect(gold("bob")):equals(500)
            end),
        }),
        test.case("invalid_recipient_is_skipped_and_picker_remains_safe", {
            enter("alice", 1),
            enter("bob", 2),
            test.step(1),
            party_command(2, "alice", 1, "party.invite", { target = "bob" }),
            test.step(1),
            party_command(3, "bob", 1, "party.accept", { inviter = "alice" }),
            test.step(1),
            test.run(function()
                require("engine.ecs/v1").get(player_entity("bob"), "d2legacy.player.progress"):set("level", 0)
            end),
            open(4, "alice", 2),
            test.step(1),
            test.run(function()
                test.expect(gold("alice")):equals(1000)
                test.expect(gold("bob")):equals(0)
            end),
        }),
        test.case("invalid_picker_aborts_without_mutating_a_recipient", {
            enter("alice", 1),
            enter("bob", 2),
            test.step(1),
            party_command(2, "alice", 1, "party.invite", { target = "bob" }),
            test.step(1),
            party_command(3, "bob", 1, "party.accept", { inviter = "alice" }),
            test.step(1),
            test.run(function()
                require("engine.ecs/v1").get(player_entity("alice"), "d2legacy.player.progress"):set("level", 0)
            end),
            open(4, "alice", 2),
            test.step(1),
            test.run(function()
                local _, amount = pile()
                test.expect(gold("alice")):equals(0)
                test.expect(gold("bob")):equals(0)
                test.expect(amount:get("amount")):equals(1000)
            end),
        }),
        test.case("already_over_cap_balances_never_decrease", {
            enter("alice", 1),
            enter("bob", 2),
            test.step(1),
            party_command(2, "alice", 1, "party.invite", { target = "bob" }),
            test.step(1),
            party_command(3, "bob", 1, "party.accept", { inviter = "alice" }),
            test.step(1),
            test.run(function()
                set_gold("alice", 12000)
                set_gold("bob", 11000)
            end),
            open(4, "alice", 2),
            test.step(1),
            test.run(function()
                local _, amount = pile()
                test.expect(gold("alice")):equals(12000)
                test.expect(gold("bob")):equals(11000)
                test.expect(amount:get("amount")):equals(1000)
            end),
        }),
        test.case("capped_recipient_and_picker_leave_conserved_excess", {
            enter("alice", 1),
            enter("bob", 2),
            test.step(1),
            party_command(2, "alice", 1, "party.invite", { target = "bob" }),
            test.step(1),
            party_command(3, "bob", 1, "party.accept", { inviter = "alice" }),
            test.step(1),
            test.run(function()
                set_gold("alice", 9950)
                set_gold("bob", 9900)
            end),
            open(4, "alice", 2),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity, amount = pile()
                test.expect(gold("alice")):equals(10000)
                test.expect(gold("bob")):equals(10000)
                test.expect(amount:get("amount")):equals(850)
                test.expect(ecs.get(entity, "d2legacy.world.inactive")):is_nil()
            end),
            open(5, "alice", 3),
            test.step(1),
            test.run(function()
                local _, amount = pile()
                test.expect(gold("alice")):equals(10000)
                test.expect(gold("bob")):equals(10000)
                test.expect(amount:get("amount")):equals(850)
            end),
        }),
        test.case("unadmitted_and_remote_open_cannot_collect", {
            open(1, "intruder", 1),
            enter("alice", 1),
            test.step(1),
            test.run(function()
                require("engine.ecs/v1").get(player_entity("alice"), "d2legacy.world.position"):set("x", 90)
            end),
            open(2, "alice", 1),
            test.step(1),
            test.run(function()
                local _, amount = pile()
                test.expect(gold("alice")):equals(0)
                test.expect(amount:get("amount")):equals(1000)
            end),
        }),
        test.case("missing_picker_position_cannot_use_generic_admission_fallback", {
            enter("alice", 1),
            test.step(1),
            test.run(function()
                require("engine.ecs/v1").remove(player_entity("alice"), "d2legacy.world.position")
            end),
            open(2, "alice", 1),
            test.step(1),
            test.run(function()
                local _, amount = pile()
                test.expect(gold("alice")):equals(0)
                test.expect(amount:get("amount")):equals(1000)
            end),
        }),
        test.case("consumed_pile_stays_consumed_after_restore_and_later_join", {
            enter("alice", 1),
            test.step(1),
            open(2, "alice", 1),
            test.step(1),
            test.restore_checkpoint(),
            enter("bob", 1, nil, 3),
            test.step(1),
            test.run(function()
                local ecs = require("engine.ecs/v1")
                local entity, amount = pile()
                test.expect(amount:get("amount")):equals(0)
                test.expect(ecs.get(entity, "d2legacy.world.inactive") ~= nil):is_true()
                test.expect(gold("alice")):equals(1000)
                test.expect(gold("bob")):equals(0)
            end),
        }),
    },
})
