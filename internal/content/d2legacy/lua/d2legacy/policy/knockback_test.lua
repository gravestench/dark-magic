local test = require("d2legacy.tests/v1")

local function missile_case(name, vector)
    return test.case(name, {
        test.run(function()
            local knockback = require("d2legacy.policy.knockback")
            local numeric = require("d2legacy.policy.numeric")
            local missile_seed = { lo = vector.seed_lo or 1, hi = vector.seed_hi or 0 }
            local damage = { result = vector.initial_result or 0, physical = 123 }
            local context = {
                chance = vector.chance,
                state54_result = vector.state54_result or 0,
                missile_seed = missile_seed,
                attacker_seed = { lo = 1, hi = 0 },
                defender_seed = { lo = 9, hi = 0 },
                damage = damage,
            }
            local log = {}
            local ops = {
                rand = function(seed, modulus)
                    test.expect(seed, "chance draw owns the missile pair"):equals(missile_seed)
                    local before = { lo = seed.lo, hi = seed.hi }
                    -- Observe the real numeric operation; no scripted RNG result is returned.
                    local roll = numeric.rand(seed, modulus)
                    log[#log + 1] = {
                        seed = "missile",
                        modulus = modulus,
                        roll = roll,
                        before = before,
                        after = { lo = seed.lo, hi = seed.hi },
                        result = damage.result,
                    }
                    return roll
                end,
            }

            test.expect(knockback.missile(context, ops), "missile return"):is_nil()

            local expected_seed = { lo = vector.final_lo, hi = vector.final_hi or 0 }
            local expected_log = {}
            if vector.roll ~= nil then
                expected_log[1] = {
                    seed = "missile",
                    modulus = 100,
                    roll = vector.roll,
                    before = { lo = vector.seed_lo or 1, hi = vector.seed_hi or 0 },
                    after = expected_seed,
                    result = vector.initial_result or 0,
                }
            end
            test.expect(log, "ordered real numeric draws"):deep_equals(expected_log)
            test.expect(context, "result, exact seeds and untouched supplied fields"):deep_equals({
                chance = vector.chance,
                state54_result = vector.state54_result or 0,
                missile_seed = expected_seed,
                attacker_seed = { lo = 1, hi = 0 },
                defender_seed = { lo = 9, hi = 0 },
                damage = { result = vector.result, physical = 123 },
            })
            test.expect(context.missile_seed, "caller-owned missile pair"):equals(missile_seed)
            test.expect(context.damage, "caller-owned damage record"):equals(damage)
        end),
    })
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        -- Static 1.13c KNB-R02a/R02b vectors at a supplied-predicate boundary.
        -- These do not query state storage or execute a production contact pipeline.
        missile_case("KNB-V04", {
            chance = 33,
            seed_lo = 2,
            roll = 70,
            final_lo = 0xD58D218A,
            result = 0,
        }),
        missile_case("KNB-V05", {
            chance = 75,
            seed_lo = 2,
            roll = 70,
            final_lo = 0xD58D218A,
            result = 8,
        }),
        missile_case("KNB-V06a", {
            chance = 1,
            seed_lo = 0,
            roll = 0,
            final_lo = 0,
            result = 8,
        }),
        missile_case("KNB-V06b", {
            chance = 1,
            seed_hi = 15,
            roll = 0,
            final_lo = 0x6AC690D4,
            result = 8,
        }),
        missile_case("KNB-V07a", {
            chance = 0,
            initial_result = 4,
            final_lo = 1,
            result = 4,
        }),
        missile_case("KNB-V07b", {
            chance = 0,
            initial_result = 12,
            final_lo = 1,
            result = 12,
        }),
        missile_case("KNB-V12a", {
            chance = 255,
            state54_result = 0x00400000,
            initial_result = 4,
            final_lo = 1,
            result = 4,
        }),
        missile_case("KNB-V12d", {
            chance = 255,
            state54_result = 0x00400000,
            initial_result = 12,
            final_lo = 1,
            result = 12,
        }),
        missile_case("KNB-V17a", {
            chance = 85,
            roll = 85,
            final_lo = 0x6AC690C5,
            result = 0,
        }),
        missile_case("KNB-V17b", {
            chance = 86,
            roll = 85,
            final_lo = 0x6AC690C5,
            result = 8,
        }),
        missile_case("KNB-V17c", {
            chance = 100,
            roll = 85,
            final_lo = 0x6AC690C5,
            result = 8,
        }),
        missile_case("KNB-V17d", {
            chance = 255,
            roll = 85,
            final_lo = 0x6AC690C5,
            result = 8,
        }),
        missile_case("KNB-V18a", {
            chance = 33,
            seed_lo = 5,
            roll = 33,
            final_lo = 0x15E0D3D9,
            final_hi = 2,
            result = 0,
        }),
        missile_case("KNB-V18b", {
            chance = 34,
            seed_lo = 5,
            roll = 33,
            final_lo = 0x15E0D3D9,
            final_hi = 2,
            result = 8,
        }),
        -- Additional KNB-R02a OR invariants; these are not new original-game observations.
        missile_case("successful_missile_chance_preserves_other_result_bits", {
            chance = 86,
            initial_result = 4,
            roll = 85,
            final_lo = 0x6AC690C5,
            result = 12,
        }),
        missile_case("successful_missile_chance_retains_existing_knockback_bit_and_draws", {
            chance = 86,
            initial_result = 12,
            roll = 85,
            final_lo = 0x6AC690C5,
            result = 12,
        }),
    },
})
