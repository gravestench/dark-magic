# Gaps: skill-monster-melee-specials

Contract: [monster melee specials](../../specs/skill-monster-melee-specials.md). Evidence: [companion](../../research/spec-evidence/skill-monster-melee-specials.md).

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| MML-Q01 | integration | MML-R01 | Named-row compiled binding; supplied stages stay available | Compiled named records | 2026-10-29: complete effective Skills read and dispatch table; no compiled row capture | Corroborate prepared fields and loader linkage |
| MML-Q02 | behavior | MML-R04 | Fire Hit full attack/start effects | Complete target/combat preparation callees | 2026-10-29: action body read, final call identified as shared start dispatcher, not inferred melee helper | Original-image decode and complete output trace |
| MML-Q03 | behavior | MML-R05 | Quick Strike damage/fallback/launch | Start damage and monster missile builder closure | 2026-10-29: start/action bodies read; exact missing calls named | Resolve imports and trace each request/output |
| MML-Q04 | behavior | MML-R06 | Frenzy full event/damage/state behavior | Animation predicate and effect/follow-up | 2026-10-29: callbacks read, specific callees remain unresolved | Resolve predicate/effect paths through output writes |
| MML-Q05 | behavior | MML-R07 | Complete jump phases/path/contact | Start, path/collision/room/combat operations | 2026-10-29: action's multi-phase body read; following entry distinguished | Trace start and each phase through movement/damage writes |
| MML-Q06 | behavior | MML-R02, MML-R08 | Path preparation and upstream alternate condition; none for supplied action stage | Imported path effects and later condition producer | 2026-10-29: start retry path and action condition consumer read | Resolve path/phase producers and observe movement |
| MML-Q07 | integration | MML-R03, MML-R09 | Actual corpse eligibility/removal and stat reads/writes; arithmetic/request stage stays available | Predicate/lifecycle/stat operation closure | 2026-10-29: eating interior entry and health suffix read; NRF operand order established, integration incomplete | Complete dependencies and matching health/corpse records |
| MML-Q08 | behavior | MML-R10 | Full whip target/chance/state/effects | Target family, RNG branch and both helpers | 2026-10-29: class constant, modulo draw and helper selection located | Resolve eligibility/effects, add exact seed vectors |
| MML-Q09 | behavior | MML-R11 | Full stomp damage/radius/area | Damage helpers and target filter/callback | 2026-10-29: positive missile gate and area request located; local callees incomplete | Trace all damage/area operations and distinguishing vectors |
| MML-Q10 | observation | MML-O01, MML-O02, MML-V01, MML-V02, MML-V03, MML-V04 | Runtime-confirmation claims only | Matching owned gameplay captures | 2026-10-29: no matching callback invocation or gameplay effect captured; observation inconclusive | Eligible action and exact raw-value observations |
