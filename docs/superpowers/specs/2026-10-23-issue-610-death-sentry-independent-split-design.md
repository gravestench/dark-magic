# Issue #610: independently truncate Death Sentry corpse damage channels

## Intent and evidence

Death Sentry currently computes fire as `floor(amount * pct / 100)` and physical as `amount - fire`. The merged Death Sentry corpse-damage specification, DSC-R08a and DSC-V03, records a 1.13c static-code result: both channels truncate their percentages of the original amount independently. With 101 raw damage and a 50% fire share, the specified result is 50 fire and 50 physical; the current remainder gives 50 fire and 51 physical. Applying this to the supported 1.14d runtime path is an inference from that 1.13c evidence, not an observed 1.14d game result.

## Boundary and approach

Change only the physical-channel calculation in `trap_skill.lua` to `floor(amount * (100 - pct) / 100)`, retaining the existing fire calculation. The accepted option is this local arithmetic repair plus a controlled production-path test. A shared damage-policy change or new corpse-explosion subsystem would affect attacks outside #610. For the controlled 50% fixture and supported valid-percent data, both products are nonnegative, so `floor` equals truncation toward zero. The decoder does not itself cap this percentage at 100.

Keep corpse selection, raw roll, percentage lookup, target radius, mitigation, event creation, and corpse consumption intact. This issue does not implement the specification's other duration/router, RNG, blast-ring, PvP, proc, or level-scaling behavior. It does not assert original 1.14d verification.

## Verification design

Use the authority trap test's existing real Death Sentry execution path. Give its synthetic corpse 202 raw maximum health and its synthetic Corpse Explosion skill a fixed 50% minimum and maximum roll with a 50% fire share, producing 101 raw damage before the split. Assert 50 physical and 50 fire on the event's attached `d2legacy.combat.damage_bundle`, plus aggregate rolled total 100 and target identity on `d2legacy.combat.event`, and corpse consumption. The odd-unit fixture must fail against the current remainder calculation and pass after the one-line change. Retain checkpoint parity coverage.

Run the focused trap suite, full Go suite, architecture and Lua checks required by the repository, then obtain independent code and evidence reviews. Report the exact checks on the submitted commit and whether owned game-data tests ran.

## Remaining uncertainty

The merged specification's 1.13c result is static evidence. Native 1.14d observation and the broader Death Sentry behaviors remain with the parent issue #420 and its linked work. Any behavior outside this two-channel split requires a separate issue.
