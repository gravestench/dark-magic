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

Comparison of the original 1.13c behavior with the Dark Magic implementation, as previously
recorded. Findings are unchanged; rule and vector references use the spec's current ids.
The comparison's status column records its earlier assessment, including the old runtime gate
in row 6; current static labels and observation boundaries are defined by the spec and the open
evidence gaps above. This regeneration did not inspect or update the implementation comparison.

Repo files named in the comparison: `internal/content/d2legacy/lua/d2legacy/policy/mitigation.lua`, `policy/damage.lua`, `docs/research/COMBAT_DAMAGE_AND_DEATH.md` (stage map). The comparison treated repo values as implemented assumptions unless marked verified.

| # | Spec rule | Repo file/function | Current repo behavior | Original (1.13c) behavior | Original status |
|---|-----------|--------------------|-----------------------|---------------------------|-----------------|
| 1 | DMI-R06 (order flat reduction -> resist -> absorb); vector DMI-V01 | `policy/mitigation.lua` | Percent resist first, then subtracts flat `physical_reduction_raw`. DMI-V01: `5120*75/100 - 1280 = 2560` | Flat reduction first, then percent: `(5120-1280)*75/100 = 2880` | verified (static) |
| 2 | DMI-R07 caps | `policy/mitigation.lua` | Clamps physical resist to [-100, 100]; clamps element resist to [-100, 95], which matches only when the max-resist stat >= 20 | Non-monster defender: physical cap 50 (no max stat); element cap 75 + max stat, clamped at 95; with no max stat id cap is 75. No player cap above 95 | verified (static) |
| 3 | DMI-R07 step 3 difficulty penalty | `policy/mitigation.lua` | Missing | Applied to non-regular-monster defenders for elements (not stats 36/37) after pierce | verified (static) |
| 4 | DMI-R08 percent absorb cap | `policy/mitigation.lua` (`fire_absorb_percent`) | Clamped to 100 | Capped at 40 | verified (static) |
| 5 | DMI-R08 flat absorb | `policy/mitigation.lua` | Absent | Present, after percent absorb, whole-HP (`<<8`) | verified (static) |
| 6 | Negative result of flat reduction (DMI-Q03) | `policy/mitigation.lua` | Not stated in the earlier spec text | Code stores `v - reduction` unclamped; the earlier spec text said to implement a clamp only after a probe | asm-only observation, runtime gated |
| 7 | Items 6+ of original contradiction list (life, crit, leech, poison) | n/a | Moved to combat-damage-pipeline-order spec | n/a | n/a |

Tracker items from the earlier spec's section 6 (vectors DMI-V01 to DMI-V05 on owned runtime; #383 update) were process content and were dropped; runtime capture of DMI-V01 to DMI-V05 remains open as DMI-O01 above.
