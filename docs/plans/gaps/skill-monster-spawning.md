# Gaps: skill-monster-spawning

Spec: [Monster skill spawning and calls for help](../../specs/skill-monster-spawning.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|-----|------|---------|-------------------------------|------------------|-----------------------------------|----------------|
| MSP-Q01 | integration | MSP-R03, MSP-R04, MSP-V02 | Actual creator, initialized unit, optional overlay/final callback; supplied fallback schedule available | Full creator/initialization/state contracts | 2026-10-29: full local minion/nest bodies, creator wrapper and main creator entry read; imports and field getters resolved; downstream creation/registration incomplete | Follow creator and selected-unit initialization/state through stored effects with integration vectors |
| MSP-Q02 | behavior | MSP-R05 | Complete target effect/delayed spawning | Target predicate, state/list producer and removal callback contract | 2026-10-29: original action 133 -> predicate/state-producer/removal-callback chain located and caller body read; consumer effects unverified | Read predicate/state-producer/removal-callback through actual target/creation writes |
| MSP-Q03 | behavior | MSP-R06 | Complete help response/timing/recipient contract | Consumer of target-associated request | 2026-10-29: original action 128 -> request consumer located and caller body read; frame expression and state request identified but consumer incomplete | Follow consumer/event to affected creatures with exact timing/eligibility |
| MSP-Q04 | behavior | MSP-R04 | Production of prepared class/mode/position and nest event/room/collision effects | Both starts and preparation dependencies | 2026-10-29: slots 49/62 and four field getters identified; start 49 entry and nest event/room sequence inspected; full producer chain incomplete | Read both starts and all preparation-owned callees to semantic outputs |
| MSP-Q05 | integration | MSP-R01, MSP-O01 | Actual AI/frame dispatch; cells and narrow compiled selectors available | Actual entry/AI handoff | 2026-10-29: complete effective text, original-image slots and five loaded id/start/action records agree; no actual callback capture | Capture actual AI/frame entry |
| MSP-Q06 | observation | MSP-R03, MSP-O01 | Runtime-confirmation only | Naturally triggered exact argument/result capture | 2026-10-29: original profile/data and selected compiled records collected; no matching spawn/help encounter completed | Perform owned-game procedure and record exact successful assertions |

2026-10-29 continuation: R07/R08 close supplied-record spawn-tuple production
and persisted cache writes; R09 closes the post-preparation Nest request
schedule; R10 closes Impregnate's eligible-target state request and ignored
producer return; R11 closes Cry Help's calculation/deadline/request subset
and overlay-call order. These supplied stages remain available independently
of Q01 creator/owner/list/overlay effects, Q02 actual infestation lifecycle,
Q03 owner-linked command propagation/responses, Q04 full start 49 event/room
preparation and actual producer values, and Q05 AI scheduling. Expanded full review and the clean final focused delta review are complete, as recorded in the companion. No runtime confirmation follows from these closures.
