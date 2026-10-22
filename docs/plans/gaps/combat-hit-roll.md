# Gaps: combat-hit-roll

Spec: [Attack hit roll](../../specs/combat-hit-roll.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| Mastery | behavior | HIT-Q01/HIT-R02f/HIT-R02b | Actual meleeExtra; supplied percent arithmetic available | Weapon selection and special mastery producer | 2026-10-15 D2Game + 0x372C0/D2Common + 0x4F640 targeted full listings; 0x4F200/0x303A0 and imported equipment selection remain incomplete | Complete exact producer chain, including early-return branch |
| Monster relation production | integration | HIT-Q02/HIT-R02b | Data-populated matrix/type index; supplied matrix arithmetic available | Named loader fields/equivalence producer | 2026-10-13 D2Game + 0x1620 relation consumer; 2026-10-15 current GDF/NRF reassessment and supplied bounds/sum vector; exact producer not established | Loader/matrix contract for named input |
| Monster condition writers | outside-boundary | HIT-Q03/HIT-R02c/HIT-R02d | none locally; upstream producer meanings of neutral conditions | Spawn/classification writers | 2026-10-13 masks 0x0A/0x02 and null-data behavior read; 2026-10-15 retained neutral booleans instead of invented boss labels | Identify exact writers and resulting logical conditions |
| Prevent-heal applier | integration | HIT-Q04/HIT-R09a/HIT-R09b | Actual timed contributions/callbacks; hook request available | Attach/write/free/refresh integration | 2026-10-15 complete D2Game + 0xA00B0 listing read and STL-R06a/b/c/STL-R11 primitives reused; CSA-R05 is a related cold-applier pattern; their propagation gates remain | Close owned dependency gates for this request |
| Pending gameplay | observation | HIT-Q05/HIT-O01/HIT-R01a/HIT-R02a/HIT-R07/HIT-R08/HIT-R09a | Runtime confirmation only; all supplied-input checks available | Caps, player attack-rating terms, defense adjustments and the post-hit request beyond the recorded ratio/level cases | 2026-10-12 historical observation matched the ratio/level cases only (HIT-O01); 2026-10-15 migration and restructure performed no gameplay campaign | Separately authorized matching original-game observations |

## Implementation gaps

The following findings are preserved from the existing comparison and were not re-audited during behavioral regeneration.

Gap list: original 1.13c behavior versus the implementation, as input for a future plan, not a plan. The Spec rule column names the current rule ids covering each historical subject; numbers quoted inside other cells are the historical numbering.

Source: old section 2 ("What Dark Magic already has"), the "Matches the repo's implemented assumption" note in old 3.0, and old sections 6-7. Original status of all 1.13c rules below: verified in 1.13c code (static) unless noted.

| # | Spec rule | Repo file/function/doc (as named in old spec) | Current repo behavior | Original (1.13c) behavior | Verified/gated |
|---|-----------|-----------------------------------------------|-----------------------|---------------------------|----------------|
| 1 | HIT-R01–HIT-R08 (hit/miss roll), especially HIT-R04–HIT-R08 | `combat.LegacyHitChance`; vectors in `internal/game/combat/hit_chance_test.go` | 1.10f-derived: `ratingFactor = 100*AR/(AR+DEF)`; `chance = 2*attackerLvl*ratingFactor/(attackerLvl+defenderLvl)`, clamp 5..95; `hit = random % 100 < chance`; negative defense moves magnitude into attack rating and vice versa, then both floor at 0. Dark Magic's Go/Lua code was not audited in the old pass. | Same core arithmetic (old steps 4-8 agree). | Verified (static) |
| 2 | HIT-R01/HIT-R01a (defense input) | same resolver | Does not model the stat 32 (missile) vs stat 33 (melee) defense split. | `DEF = CalcTotalDefenseRating + Stat(defender, isMissile ? 32 : 33)`. | Verified (static) |
| 3 | HIT-R02a–HIT-R02g/HIT-R03 (attack rating percent, ignore/reduce target defense) | same resolver | Does not model stat 119 / bonus percent, stat 115 (ignore target defense), stat 116 (reduce target defense), stat 179 per-monster-type bonus. | See spec rule 2; monsters skip the 0xDB1F0 step. | Verified (static); helper effects asm-only/unresolved |
| 4 | HIT-R03/HIT-R04 (order) | same resolver | Percent not applied to AR before the negative-defense swap (order differs from 1.13c). | `AR = AR + (AR*pct)/100` is applied before the negative-defense swap (rule 4). | Verified (static) |
