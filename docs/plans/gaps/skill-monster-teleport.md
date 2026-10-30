# Gaps: skill-monster-teleport

Contract: [monster teleport](../../specs/skill-monster-teleport.md). Evidence: [companion](../../research/spec-evidence/skill-monster-teleport.md).

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| MTP-Q02 | integration | MTP-R04 to MTP-R11 | Complete collision placement, general room transfer and nontrivial client synchronization; local requests/reached same-room stores are verified | Nearby search 4D140, collision footprint producers and collision-map operations, room transfer D760 through 66700/66800, nonclear synchronization E1610 and its client helpers | 2026-10-29 continuation: original bodies/imports read and corroborated; exact target/room selection, monster relocation caller, clear-sync guard and reached same-room suffix isolated. Supplied results do not prove these remaining effects | Close each named dependency with original output writes and distinguishing fixtures, or consume a verified owning contract |
| MTP-Q03 | integration | MTP-R05, MTP-R07, MTP-R12 | Complete link/state/stat propagation, allocation failure, event processing, AI and death execution | Link 33020/32CA0 state-98/list-stat integration; event allocator failure; subsequent persistent-event processing; AI 9D230; death DEEE0 | 2026-10-29 continuation: eligibility, state/link order, immediate saved-key prefix, successful event construction, keyed removal, interaction conditions and health-threshold requests closed. Later effects explicitly excluded | Trace remaining operations through their effects; use exact dependency contracts where separately owned |
| MTP-Q04 | observation | MTP-O01 | Gameplay confirmation only | Matching owned original-profile teleport/attachment action capture | 2026-10-29: initial observation supplied no handler invocation. Continuation Act I level-1 character yielded effective States and selected live scalar reads, but no eligible teleport or gameplay trace; observation inconclusive | Reach a natural eligible encounter and capture actual selected route, committed point/room, update/event effects and callback return |

## Resolved historical question

MTP-Q01: 2026-10-29 continuation original Skills loader descriptor maps signed consumed word to aurastate; complete effective States text and selected compiled Skills330/States143 scalar reads establish attached = 143 in the reported profile. Static module identity and runtime loaded-module/archive limits remain distinct in the companion.
