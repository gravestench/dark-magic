# Issue #598: running stamina drain arithmetic

## Intent

Make the existing running-stamina rule produce the signed 32-bit drain amount recovered from original Diablo II 1.13c code. An admissible large `item_staminadrainpct` source currently gives the authoritative player tick and connected prediction a different result. The goal is one corrected result in both paths, while identifying the static-code evidence limit for 1.13c and the inference for 1.14d.

## Boundary and evidence

Issue #598 covers only the drain amount for valid prepared inputs: an unsigned 8-bit class `RunDrain`, signed 32-bit torso speed and percentage stat, and admitted nonnegative current/maximum raw stamina. `docs/specs/player-stamina-drain-while-running.md` SDR-R04/R06/R07/R08 and SDR-V07(c)/V10(d), backed by its companion original-code investigation, specify the local arithmetic. They do not verify original-game stored stamina or live event cadence; SDR-Q02/Q05 and 1.14d observations remain open.

The current `internal/content/d2legacy/lua/d2legacy/systems/stamina.lua` invokes `d2legacy.movement_rules/v1`.stamina, which calls `movement.AdvanceStamina` in the Go movement adapter. Connected prediction calls that same Go function. Both already use the same private `staminaDrain` formula; this issue changes its arithmetic without altering the Lua module signature, `StaminaTick`, `AdvanceStamina`, or prediction interface. Stamina recovery, town membership, walking transition details, stat propagation, malformed input policy, and 1.14d runtime verification remain outside scope.

## Approaches

1. **Correct the existing shared Go rule (chosen).** Apply signed 32-bit wrapping at the original arithmetic stages inside private `staminaDrain`. This preserves authority/prediction parity and introduces no second policy path or new interface. The formula already lives in this adapter; correcting it is a narrow compatibility repair, not an assertion that Go is the desired long-term Diablo policy owner.
2. Move the authoritative formula into Lua and retain a Go prediction mirror. This would duplicate the rule and create a new drift risk without a shared parity contract. It also changes ownership beyond the arithmetic bug.
3. Move all stamina and prediction policy into a shared Lua runtime. This best satisfies the general policy ownership rule but requires a larger client prediction integration and performance review, beyond issue #598.

`docs/ARCHITECTURE.md` and `docs/GAMEPLAY_OWNERSHIP.md` assign Diablo formulas to Lua and Go adapters to thin translation. The current movement implementation is a transitional tension with that rule, while `docs/research/MOVEMENT_COLLISION_AND_PATHFINDING.md` explicitly documents the shared pure stamina rule. This design keeps the existing seam for one scoped parity fix. Any ownership migration needs a separately scoped follow-up; #346 tracks broader movement and stamina parity but does not currently require that migration. This fix adds no new Go policy surface.

## Arithmetic

For the supplied valid inputs, compute in this order:

1. `w = i32(2 * RunDrain)` in raw 8.8 stamina units.
2. `m = i32(trunc(speed / 10) + 1)` and `w = i32(w * m)`. With no torso source, the supplied speed is zero and the multiplier is one in the current model.
3. If signed percentage `p` is nonzero, `product = i32(p * w)` before division; then `w = i32(w + trunc(product / -100))`.
4. If signed `w < 1`, use one raw unit. `AdvanceStamina` retains its existing drain, clamp, and force-walk decisions.

`i32` means retain the low 32 bits and interpret them as signed. Division truncates toward zero. The current Go adapter fields are `int64`, so convert at each named original-width stage rather than relying on 64-bit overflow. Do not silently clamp `p >= 100` before the signed product; SDR-V07(c) distinguishes that shortcut.

## Acceptance checks

- Direct shared-rule tests use `CurrentRaw=MaximumRaw=23552`, running/moving outside town: SDR-V07(c) `RunDrain=20`, speed 0, `p=107374183` yields drain 40, `CurrentRaw=23512`, `ForceWalk=false`; SDR-V10(d) synthetic speed `1000000000`, `p=0` wraps the weight product negative and floors to drain 1, `CurrentRaw=23551`, `ForceWalk=false`.
- Regression direct vectors: `p=33` yields 23525, `p=-33` yields 23499, speed 10 yields 23472, speed -9 yields 23512, `p=100` and `RunDrain=0` each yield 23551. These check truncation, torso multiplier, and one-unit floor.
- An authoritative Lua player-entry/tick test admits a passive source `item_staminadrainpct=107374183`, then resets `stamina_raw` to 2560 after the entry tick and advances one running movement tick outside town. It observes the resolved percentage and `stamina_raw=2520`; the current rule produces 2559.
- A connected prediction HUD with `CurrentRaw=2560`, `MaximumRaw>=2560`, `RunDrain=20`, `p=107374183`, running velocity, and outside-town level advances one tick and observes `staminaRaw=2520` through the existing prediction path.
- Existing movement, stamina, prediction, architecture, and formatting checks pass. The PR reports exact commands on its submitted commit and whether local real-asset and race suites ran.

## Risks and limits

The extreme percentage is valid through Dark Magic's current integer stat-source admission, but it is a synthetic discriminator, not a claim that an ordinary item grants that value. The extreme torso speed case is direct-rule-only; do not claim it is live equipment content. The 1.13c rule is recovered from static code with no successful original runtime storage observation, and 1.14d matching behavior is inferred. Keeping the existing Go shared rule maintains a known architecture debt; no broader migration is implied by this PR.
