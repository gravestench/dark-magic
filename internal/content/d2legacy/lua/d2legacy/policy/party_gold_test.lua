local test = require("d2legacy.tests/v1")

local function member(id, gold, level, counts, receives)
    return {
        id = id,
        gold = gold,
        level = level,
        counts_for_divisor = counts,
        can_receive = receives,
    }
end

local function allocate(amount, members, safe)
    return require("d2legacy.policy.party_gold").allocate(amount, "picker", members, safe)
end

local function request_is(request, id, accepted, kind, amount)
    test.expect(request.id):equals(id)
    test.expect(request.accepted):equals(accepted)
    test.expect(request.kind):equals(kind)
    test.expect(request.amount):equals(amount)
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        test.case("three_living_members_give_rounding_remainder_to_picker", {
            test.run(function()
                local result = allocate(1000, {
                    member("picker", 0, 10, true, true),
                    member("alice", 0, 10, true, true),
                    member("bob", 0, 10, true, true),
                }, false)
                test.expect(result.count):equals(3)
                test.expect(#result.requests):equals(3)
                request_is(result.requests[1], "alice", 333, "add", 333)
                request_is(result.requests[2], "bob", 333, "add", 333)
                request_is(result.requests[3], "picker", 334, "add", 334)
                test.expect(result.excess):equals(0)
            end),
        }),
        test.case("unaccepted_share_returns_to_picker_then_excess_stays", {
            test.run(function()
                local members = {
                    member("picker", 0, 10, true, true),
                    member("alice", 99900, 10, true, true),
                }
                local baseline = allocate(1001, members, false)
                request_is(baseline.requests[1], "alice", 100, "add", 100)
                request_is(baseline.requests[2], "picker", 901, "add", 901)
                test.expect(baseline.excess):equals(0)

                members[1].gold = 99950
                local capped = allocate(1001, members, false)
                request_is(capped.requests[1], "alice", 100, "add", 100)
                request_is(capped.requests[2], "picker", 50, "add", 50)
                test.expect(capped.excess):equals(851)
            end),
        }),
        test.case("dead_recipient_does_not_count_but_can_receive", {
            test.run(function()
                local result = allocate(1000, {
                    member("picker", 0, 10, true, true),
                    member("alice", 0, 10, true, true),
                    member("bob", 0, 10, false, true),
                }, false)
                test.expect(result.count):equals(2)
                request_is(result.requests[1], "alice", 500, "add", 500)
                request_is(result.requests[2], "bob", 500, "add", 500)
                request_is(result.requests[3], "picker", 0, "add", 0)
                test.expect(result.excess):equals(0)
            end),
        }),
        test.case("production_never_distributes_more_than_the_physical_pile", {
            test.run(function()
                local result = allocate(1000, {
                    member("picker", 0, 10, true, true),
                    member("living", 0, 10, true, true),
                    member("dead-one", 0, 10, false, true),
                    member("dead-two", 0, 10, false, true),
                }, true)
                test.expect(result.count):equals(2)
                request_is(result.requests[1], "living", 500, "add", 500)
                request_is(result.requests[2], "dead-one", 500, "add", 500)
                request_is(result.requests[3], "dead-two", 0, "add", 0)
                request_is(result.requests[4], "picker", 0, "add", 0)
                test.expect(result.excess):equals(0)
            end),
        }),
        test.case("signed_negative_split_truncates_toward_zero", {
            test.run(function()
                local result = allocate(-1000, {
                    member("picker", 1000, 10, true, true),
                    member("alice", 1000, 10, true, true),
                    member("bob", 1000, 10, true, true),
                }, false)
                request_is(result.requests[1], "alice", -333, "add", -333)
                request_is(result.requests[2], "bob", -333, "add", -333)
                request_is(result.requests[3], "picker", -334, "add", -334)
                test.expect(result.excess):equals(0)
            end),
        }),
        test.case("signed_wrap_requests_zero_replacement_without_claiming_original_storage", {
            test.run(function()
                local result = allocate(100, {
                    member("picker", 2147483647, 10, true, true),
                }, false)
                request_is(result.requests[1], "picker", 100, "replace", 0)
                test.expect(result.excess):equals(0)
            end),
        }),
        test.case("supplied_level_controls_cap_arithmetic", {
            test.run(function()
                -- The original getter reads effective stat 12. This pure
                -- policy test verifies only the supplied level's cap math.
                local result = allocate(1000, {
                    member("picker", 99900, 10, true, true),
                }, false)
                request_is(result.requests[1], "picker", 100, "add", 100)
                test.expect(result.excess):equals(900)
            end),
        }),
        test.case("production_over_cap_balance_never_debits_the_picker", {
            test.run(function()
                local result = allocate(100, {
                    member("picker", 20000, 1, true, true),
                }, true)
                request_is(result.requests[1], "picker", 0, "add", 0)
                test.expect(result.requests[1].next_balance):equals(20000)
                test.expect(result.excess):equals(100)
            end),
        }),
    },
})
