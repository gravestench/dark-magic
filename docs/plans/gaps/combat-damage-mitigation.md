# Gaps: combat-damage-mitigation

Spec: [Damage mitigation](../../specs/combat-damage-mitigation.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|-----|------|---------|-------------------------------|------------------|-----------------------------------|----------------|
| Producer of `record.reductionScale` | outside-boundary | DMI-Q05; DMI-R02 input | none locally; upstream record-builder writer remains unspecified | The rule that writes the record's reduction scale on each record-building path | 2026-10-13: sibling-spec search found a supplied vector value but no producer rule; original D2Game store search found more than 40 candidate functions because the same field position occurs in unrelated structures, without identifying the damage-record writers. 2026-10-15: current sibling contracts still supply the value; [reuse assessment](../../research/spec-evidence/combat-damage-mitigation.md#regeneration-and-evidence-reuse-2026-10-15) | Owner of each record builder traces the store into the damage record and names it |
| Per-type observation pending | observation | DMI-O01; DMI-R02, DMI-R05 to DMI-R08; DMI-V01, DMI-V03, DMI-V05 | runtime-confirmation claim for O01 only; static numeric assertions remain eligible | Gameplay capture on an owned 1.13c game | 2026-10-13: not run; migration performs static and table-data reads only. Procedure written in the companion | Run DMI-O01 per runtime-protocol.md and record it in the companion |
| Penalty growth observation pending | observation | DMI-O02, DMI-Q04; DMI-R07 step 3; DMI-V15, DMI-V16 | runtime-confirmation claim for O02/Q04 only; static penalty arithmetic remains eligible | Gameplay capture on expansion Hell | 2026-10-13: not run; static reading and TD-3 (ResistPenalty 0/-40/-100, text and compiled table) recorded | Run DMI-O02 and record it |
| Negative physical field observation pending | observation | DMI-O03, DMI-Q03; DMI-R06; DMI-V08 | runtime-confirmation claim for O03/Q03 only; static negative-field arithmetic remains eligible | Gameplay capture with flat reduction above the physical roll | 2026-10-13: not run; static reading recorded | Run DMI-O03 and record it |

## Implementation gaps

Historical comparison of the original 1.13c behavior with the Dark Magic implementation,
preserved during the 2026-10-15 regeneration. The exact code snapshot for the earlier comparison
was not recorded; do not read these rows as a current implementation audit. Rule and vector
references use the spec's current ids. The comparison's status column records its earlier
assessment, including the old runtime gate in row 6; current static labels and observation
boundaries are defined by the spec and the open evidence gaps above. This publication did not
refresh the implementation comparison.

Repo files named in the comparison: `internal/content/d2legacy/lua/d2legacy/policy/mitigation.lua`, `policy/damage.lua`, `docs/research/COMBAT_DAMAGE_AND_DEATH.md` (stage map). The comparison treated repo values as implemented assumptions unless marked verified.

| # | Spec rule | Repo file/function at earlier comparison | Repo behavior recorded earlier | Original (1.13c) behavior | Original status |
|---|-----------|--------------------|-----------------------|---------------------------|-----------------|
| 1 | DMI-R06 (order flat reduction -> resist -> absorb); vector DMI-V01 | `policy/mitigation.lua` | Percent resist first, then subtracts flat `physical_reduction_raw`. DMI-V01: `5120*75/100 - 1280 = 2560` | Flat reduction first, then percent: `(5120-1280)*75/100 = 2880` | verified (static) |
| 2 | DMI-R07 caps | `policy/mitigation.lua` | Clamps physical resist to [-100, 100]; clamps element resist to [-100, 95], which matches only when the max-resist stat >= 20 | Non-monster defender: physical cap 50 (no max stat); element cap 75 + max stat, clamped at 95; with no max stat id cap is 75. No player cap above 95 | verified (static) |
| 3 | DMI-R07 step 3 difficulty penalty | `policy/mitigation.lua` | Missing | Applied to non-regular-monster defenders for elements (not stats 36/37) after pierce | verified (static) |
| 4 | DMI-R08 percent absorb cap | `policy/damage.lua` (`fire_absorb_percent`; locator corrected 2026-10-23) | Clamped to 100 | Capped at 40 | verified (static) |
| 5 | DMI-R08 flat absorb | `policy/mitigation.lua` | Absent | Present, after percent absorb, whole-HP (`<<8`) | verified (static) |
| 6 | Negative result of flat reduction (DMI-Q03) | `policy/mitigation.lua` | Not stated in the earlier spec text | Code stores `v - reduction` unclamped; the earlier spec text said to implement a clamp only after a probe | asm-only observation, runtime gated |
| 7 | Items 6+ of original contradiction list (life, crit, leech, poison) | n/a | Moved to combat-damage-pipeline-order spec | n/a | n/a |

Tracker items from the earlier spec's section 6 (vectors DMI-V01 to DMI-V05 on owned runtime; #383 update) were process content and were dropped; runtime capture of DMI-V01 to DMI-V05 remains open as DMI-O01 above.

### Current implementation assessment (2026-11-02)

Read against main `7dcf7da5c29c1cdd9dc0ba347be6ea0742b2c21e` for the proposed
[physical-ordering plan](../combat-damage-mitigation.md). This assessment supersedes
historical implementation claims above where they disagree. It changes no behavioral
rule, evidence label or observation gate.

| Rule | Current production owner and finding | Proposed disposition |
|---|---|---|
| DMI-R01 | `policy/damage.lua:resolve` passes `monster ~= nil`; `policy/mitigation.lua:apply` has no five-class hireling exclusion. | Deferred, not evidence-gated. |
| DMI-R02 | `components/shared.lua` exposes `physical_reduction_raw`; `systems/derived_stats.lua` resolves that raw value from existing stat sources. There is no whole-stat-34 conversion, magic reduction or reductionScale here. | Supply 1280 raw for V01's 5 whole points; preserve existing units. |
| DMI-R03 | `damage.resolve` calls per-channel mitigation directly; these functions provide no event-11 dispatch. | Deferred. |
| DMI-R04 | These functions take no class flags/count and perform no undead/demon classification. | Deferred. |
| DMI-R05 | No bypass flags are consumed by these functions. | Deferred. |
| DMI-R06 | Physical currently computes `max(0, floor(raw*(100-r)/100) - physical_reduction_raw)`. V01 therefore returns 2560 instead of 2880. Initial negative bundles are rejected; negative post-reduction values are clipped. | Correct physical flat-before-percent order only; retain final clamp as an explicitly incomplete signed-record boundary. |
| DMI-R07 | Physical resistance is clamped to [-100,100], including players. Elemental maximum fields are absolute effective caps, initialized/derived with base 75, not raw original max-resist bonus stats. `systems/derived_stats.lua` already applies difficulty penalties to player elemental resistance, so the historical assertion that penalties are simply missing is stale. Hireling penalty, pierce, magic player resistance and sanctuary still need work. | Preserve selection and existing producers; do not add 75 or apply a second difficulty penalty. |
| DMI-R08 | `policy/damage.lua` applies fire-percent absorb after channel mitigation with cap 100; no other percent/flat absorb inputs are consumed here. | Deferred; retain current behavior in this slice. |
| DMI-R09 | `mitigation.lua` uses `math.floor(value*(100-resistance)/100)`, not NRF-R05's signed/overflow/divide-first implementation. | V01's positive narrow arithmetic agrees; general arithmetic remains deferred. |
| DMI-R10 | `mitigation.lua` has no RNG dependency/call; `damage.resolve` has no draw although the separate `damage.roll` function does. | Preserve this structure; full original-seed assertion coverage remains deferred. |

Existing focused mitigation/damage Lua suites passed before any implementation edit.
That baseline does not prove the new ordering: current tests explicitly expect 730 for
raw 1000, reduction 20 and resistance 25, whereas the proposed order gives 735.
The mixed-channel fixture likewise expects physical 730, fire 500, total 1230 and
remaining health 3770; after this correction those become 735, 500, 1235 and 3765.
No rule or implementation gap is marked closed by preparing the plan.

### Delivered physical ordering subset (2026-11-02, uncommitted)

The working-tree change over `7dcf7da5c29c1cdd9dc0ba347be6ea0742b2c21e` corrects the
physical branch's flat-before-resistance order (DMI-R06 supplied positive subset).
DMI-V01's raw physical assertion now returns 2880 for raw input 5120, raw reduction
1280 and selected resistance 25. The unchanged component already stores raw reduction;
no additional shift is applied. Existing mixed-channel expectations now have physical
735, unchanged fire 500, total 1235 and remaining raw monster health 3765.
The new authority test covers stat-source inputs, the production resolver and restored
pre-damage continuation. [The plan's delivery record](../combat-damage-mitigation.md#delivered-working-tree-change-and-validation-2026-11-02)
records actual checks and remaining validation limits. The full repository gate is not green.

This closes only the selected ordering difference in historical implementation row 1.
DMI-R06's complete signed record/loop behavior remains incomplete, and every other
implementation/evidence gap above remains open. In particular, the final nonnegative
clamp, current caps, overflow handling and missing original-entry inputs are retained.
This is not full-spec acceptance, original-game runtime confirmation or issue closure.
