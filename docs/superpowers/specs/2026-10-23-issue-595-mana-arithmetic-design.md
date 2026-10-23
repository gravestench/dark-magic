# Issue #595: player mana tick arithmetic

## Intent and boundary

Make the existing authoritative player mana tick match the verified 1.13c PMR-R04/R05/R07a arithmetic for admitted, nonnegative current and maximum mana. This advances parity for percent and flat recovery without claiming a 1.14d observation. The production path stays in `systems/mana.lua`; the existing player-entry, stat-source, aura-suppression and checkpoint architecture remain its inputs. Issue #595 is the acceptance owner.

## Approaches

1. **Private arithmetic in the mana owner (chosen).** Add local signed-32 wrapping and PMR's `mulDiv(factor, base, 100)` branch behavior in `systems/mana.lua`. The current system stays the sole writer of mana vitals. This is small and avoids changing unrelated stat consumers. The original `mulDiv` operation order is preserved for this call site.
2. **Shared numeric policy module.** Publish a reusable `mulDiv` API for every future Diablo policy consumer. That would require an independent full NRF-R05 contract and tests for general 32-bit arguments, including the wide branch, without a second current consumer. Defer that work to a linked issue when it is needed.

## Behavior

For a live admitted player, read raw maximum/current mana, CharStats-derived frames, percentage bonus and flat recovery as the system does today. When not suppressed, compute `base=max(1,trunc(maximum/(25*frames)))`. Compute `factor=i32(100+bonus)` and use NRF-R05's signed `mulDiv(factor,base,100)` operation, including its divide-first branch and low-32 product behavior. Add flat recovery with signed 32-bit wrap. For valid nonnegative resource inputs, apply PMR-R07a's upper-then-lower signed delta clamp, then add the delta once and update the existing raw/display vitals. Suppression skips base/percent while retaining flat recovery. Authored zero ManaRegen continues to use the existing 7,500-tick fallback.

## Checks and limits

Use the existing authority-profile Lua suite so a real player entry, derived stat source, tick and checkpoint path exercise the arithmetic. PMR-V14(a) must produce 6,550,000 raw mana rather than 6,553,700. V14(b/c) cover signed recovery and factor wrap with nonnegative current/maximum. Existing ordinary recovery, negative percent, suppression/flat, and zero-record fallback are regressions. Run the focused suite and repository-required checks on the final commit. Complete PMR event scheduling, potion-list cleanup, integrated stat propagation, invalid class construction, negative current/maximum synthetic request vectors V14(d/e), and original 1.14d confirmation remain outside issue #595. Original 1.13c supplied-input static evidence supports the arithmetic; it does not by itself prove target-version runtime parity.
