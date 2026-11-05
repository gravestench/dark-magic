local test = require("d2legacy.tests/v1")

local function expect_seed(seed, lo, hi)
    test.expect(seed.lo, "seed low word"):equals(lo)
    test.expect(seed.hi, "seed high word"):equals(hi)
end

return test.suite({
    profile = "module",
    tier = "fast",
    cases = {
        -- Original-game expectations are static 1.13c vectors, not runtime observations.
        test.case("NRF-V15g", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                -- Only V15g is covered here; V15a-f/h remain outside this case.
                local seed = { lo = 1, hi = 666 }
                local other = { lo = 9, hi = 0 }
                local path = { x_bits = 0x000C8000, y_bits = 0xFFFF0000, seed = seed }
                local x, y = numeric.path_integer_position(path)
                test.expect(x):equals(12)
                test.expect(y):equals(65535)
                test.expect(path.x_bits):equals(0x000C8000)
                test.expect(path.y_bits):equals(0xFFFF0000)
                test.expect(path.seed):equals(seed)
                expect_seed(seed, 1, 666)
                expect_seed(other, 9, 0)
            end),
        }),
        test.case("path_integer_position_rejects_null_path_without_coordinates", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                -- R10b's supplied-path rule, not KNB's absent source-path branch.
                local ok, result, y = pcall(numeric.path_integer_position, nil)
                test.expect(ok):is_false()
                test.expect(type(result)):equals("string")
                test.expect(y):is_nil()
            end),
        }),
        test.case("path_integer_position_requires_supplied_u32_words", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                local invalid = { "1", true, {}, 0.5, 0 / 0, math.huge, -math.huge, -1, 4294967296 }
                for _, value in ipairs(invalid) do
                    for _, field in ipairs({ "x_bits", "y_bits" }) do
                        local path = { x_bits = 0x000C8000, y_bits = 0xFFFF0000 }
                        path[field] = value
                        test.expect(pcall(numeric.path_integer_position, path)):is_false()
                        if value ~= value then
                            test.expect(path[field] ~= path[field]):is_true()
                        else
                            test.expect(path[field]):equals(value)
                        end
                        test.expect(path[field == "x_bits" and "y_bits" or "x_bits"])
                            :equals(field == "x_bits" and 0xFFFF0000 or 0x000C8000)
                    end
                end
                for _, path in ipairs({ false, 1, "path", {}, { x_bits = 0 }, { y_bits = 0 } }) do
                    test.expect(pcall(numeric.path_integer_position, path)):is_false()
                end
            end),
        }),
        test.case("path_integer_position_accepts_unsigned_word_boundaries", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                for _, vector in ipairs({
                    { bits = 0, expected = 0 },
                    { bits = 0x0000FFFF, expected = 0 },
                    { bits = 0x00010000, expected = 1 },
                    { bits = 0x80000000, expected = 32768 },
                    { bits = 0xFFFFFFFF, expected = 65535 },
                }) do
                    local path = { x_bits = vector.bits, y_bits = vector.bits }
                    local x, y = numeric.path_integer_position(path)
                    test.expect(x):equals(vector.expected)
                    test.expect(y):equals(vector.expected)
                    test.expect(path.x_bits):equals(vector.bits)
                    test.expect(path.y_bits):equals(vector.bits)
                end
            end),
        }),
        test.case("NRF-V01", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                local other = { lo = 9, hi = 0 }
                local variants = {
                    { draw = numeric.rand, argument = 100, result = 51 },
                    { draw = numeric.rand_mask, argument = 128, result = 95 },
                    { draw = numeric.rand, argument = 2, result = 1 },
                    { draw = numeric.rand_mask, argument = 100, result = 67 },
                }
                for _, variant in ipairs(variants) do
                    local seed = { lo = 1, hi = 666 }
                    test.expect(variant.draw(seed, variant.argument)):equals(variant.result)
                    expect_seed(seed, 0x6AC6935F, 0)
                    expect_seed(other, 9, 0)
                end

                local second = { lo = 0x6AC6935F, hi = 0 }
                test.expect(numeric.rand(second, 100)):equals(31)
                expect_seed(second, 0x2F2ED81B, 0x2C890AFD)
                expect_seed(other, 9, 0)

                local first_step = { lo = 1, hi = 666 }
                test.expect(numeric.step(first_step)):is_nil()
                expect_seed(first_step, 0x6AC6935F, 0)
                local second_step = { lo = 0x6AC6935F, hi = 0 }
                test.expect(numeric.step(second_step)):is_nil()
                expect_seed(second_step, 0x2F2ED81B, 0x2C890AFD)
                expect_seed(other, 9, 0)
            end),
        }),
        test.case("supplied_pair_continuation", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                local seed = { lo = 1, hi = 666 }
                test.expect(numeric.rand(seed, 100)):equals(51)
                expect_seed(seed, 0x6AC6935F, 0)

                -- Copy scalar values, not an authoritative checkpoint or seed owner.
                local lo, hi = seed.lo, seed.hi
                for _, vector in ipairs({
                    { modulus = 100, result = 31 }, -- NRF-V01d
                    { modulus = 64, result = 27 }, -- NRF-V03
                }) do
                    local continued = { lo = lo, hi = hi }
                    test.expect(numeric.rand(continued, vector.modulus)):equals(vector.result)
                    expect_seed(continued, 0x2F2ED81B, 0x2C890AFD)
                    expect_seed(seed, 0x6AC6935F, 0)
                    test.expect(lo):equals(0x6AC6935F)
                    test.expect(hi):equals(0)
                end

                test.expect(numeric.rand(seed, 100)):equals(31)
                expect_seed(seed, 0x2F2ED81B, 0x2C890AFD)
            end),
        }),
        test.case("NRF-V02-selected", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                -- V02d/e use RandInRange, outside this slice.
                for _, modulus in ipairs({ 0, -5 }) do
                    local seed = { lo = 1, hi = 666 }
                    test.expect(numeric.rand(seed, modulus)):equals(0)
                    expect_seed(seed, 1, 666)
                end
                local one = { lo = 1, hi = 666 }
                test.expect(numeric.rand(one, 1)):equals(0)
                expect_seed(one, 0x6AC6935F, 0)

                local mask_zero = { lo = 1, hi = 666 }
                test.expect(numeric.rand_mask(mask_zero, 0)):equals(1791398751)
                expect_seed(mask_zero, 0x6AC6935F, 0)
                local mask_one = { lo = 1, hi = 666 }
                test.expect(numeric.rand_mask(mask_one, 1)):equals(0)
                expect_seed(mask_one, 0x6AC6935F, 0)
            end),
        }),
        test.case("NRF-V03", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                local seed = { lo = 0x6AC6935F, hi = 0 }
                test.expect(numeric.rand(seed, 64)):equals(27)
                expect_seed(seed, 0x2F2ED81B, 0x2C890AFD)
            end),
        }),
        test.case("invalid_seed_words_are_rejected_even_for_nonpositive_modulus", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                local invalid = { "1", true, {}, 0.5, 0 / 0, math.huge, -math.huge, -1, 4294967296 }
                for _, value in ipairs(invalid) do
                    for _, field in ipairs({ "lo", "hi" }) do
                        for _, draw in ipairs({ numeric.step, numeric.rand, numeric.rand_mask }) do
                            local seed = { lo = 1, hi = 666 }
                            seed[field] = value
                            local ok, result = pcall(draw, seed, 0)
                            test.expect(ok):is_false()
                            test.expect(type(result)):equals("string")
                            -- NaN is unequal to itself; preserve its non-number property.
                            if value ~= value then
                                test.expect(seed[field] ~= seed[field]):is_true()
                            else
                                test.expect(seed[field]):equals(value)
                            end
                            test.expect(seed[field == "lo" and "hi" or "lo"]):equals(field == "lo" and 666 or 1)
                        end
                    end
                end
                for _, draw in ipairs({ numeric.step, numeric.rand, numeric.rand_mask }) do
                    test.expect(pcall(draw, nil, 0)):is_false()
                    test.expect(pcall(draw, 1, 0)):is_false()
                    test.expect(pcall(draw, "seed", 0)):is_false()
                    test.expect(pcall(draw, {}, 0)):is_false()
                    test.expect(pcall(draw, { lo = 1 }, 0)):is_false()
                    test.expect(pcall(draw, { hi = 666 }, 0)):is_false()
                end
            end),
        }),
        test.case("draw_arguments_require_exact_finite_in_domain_numbers_before_mutation", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                local invalid = { "1", true, {}, 0.5, 0 / 0, math.huge, -math.huge }
                for _, draw in ipairs({ numeric.rand, numeric.rand_mask }) do
                    for _, argument in ipairs(invalid) do
                        local seed = { lo = 1, hi = 666 }
                        test.expect(pcall(draw, seed, argument)):is_false()
                        expect_seed(seed, 1, 666)
                    end
                    local seed = { lo = 1, hi = 666 }
                    test.expect(pcall(draw, seed)):is_false()
                    expect_seed(seed, 1, 666)
                end
                local domains = {
                    { draw = numeric.rand, invalid = { -2147483649, 2147483648 } },
                    { draw = numeric.rand_mask, invalid = { -1, 4294967296 } },
                }
                for _, domain in ipairs(domains) do
                    for _, argument in ipairs(domain.invalid) do
                        local seed = { lo = 1, hi = 666 }
                        test.expect(pcall(domain.draw, seed, argument)):is_false()
                        expect_seed(seed, 1, 666)
                    end
                end
            end),
        }),
        test.case("draw_domain_endpoints_are_accepted", {
            test.run(function()
                local numeric = require("d2legacy.policy.numeric")
                local negative = { lo = 1, hi = 666 }
                test.expect(numeric.rand(negative, -2147483648)):equals(0)
                expect_seed(negative, 1, 666)
                local positive = { lo = 1, hi = 666 }
                test.expect(numeric.rand(positive, 2147483647)):equals(1791398751)
                expect_seed(positive, 0x6AC6935F, 0)
                local mask = { lo = 1, hi = 666 }
                test.expect(numeric.rand_mask(mask, 4294967295)):equals(1791398750)
                expect_seed(mask, 0x6AC6935F, 0)
                local zero = { lo = 0, hi = 0 }
                test.expect(numeric.step(zero)):is_nil()
                expect_seed(zero, 0, 0)
            end),
        }),
    },
})
