# Player Mana Regeneration: Gaps

Spec: [Player mana regeneration](../../specs/player-mana-regeneration.md).
Evidence and attempts: [Public companion](../../research/spec-evidence/player-mana-regeneration.md#gap-assessment-and-targeted-investigation).

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| Initial event | integration | PMR-Q01, PMR-R01b | First player event type 3 posting | Creation/join root | 2026-10-14: read scheduling helper and limited player creation/join call search; root unresolved | Complete root-to-scheduler path |
| Recurring chain/order | integration | PMR-Q02, PMR-R01b | Whole chain and inter-group ordering | Complete dispatcher lifecycle | 2026-10-14: reused player-table records and current STL-R01/R02; STL-R02a still gated | Resolve queue ordering/lifetime |
| Potion cleanup | integration | PMR-Q03, PMR-R06b, PMR-V06 | Actual contribution/list/callback/state 106 effects; supplied calls remain available | Complete lifecycle/potion callback | 2026-10-14: reread cleanup caller and assessed current STL-R05; effects unresolved | Read full detach/free and callback paths |
| Stat propagation | integration | PMR-Q04, PMR-R07b | Stored totals and client notification; arithmetic/request independently available | Complete add/propagation/sync | 2026-10-14: checked wrapper and max getter; underlying propagation/sync not completed | Trace mutation through writes and synchronization |
| Monster producer | outside-boundary | PMR-Q05, PMR-R02b | Natural type 1 entry only | Actual producer | 2026-10-14: reused immediate-dispatch lead; accepted type case alone insufficient | Identify root entering player handler |
| Gameplay observations | observation | PMR-Q06, PMR-O01, PMR-O02, PMR-O03 | Confirmation beyond narrow dead exit | Cadence/scheduler/living arithmetic/potion captures | 2026-10-12 three-entry capped sample retained; 2026-10-14 no new gameplay campaign | Valid attributable matching observations |
| Null scheduler target | integration | PMR-Q08, PMR-R01a | Actual scheduler effect for null target, not request | Guarded insertion path | 2026-10-14 entry slice inspected; current NRF state-storage guard resolves that separate question | Complete null-target scheduling path |
| Unit teardown | outside-boundary | PMR-Q09, PMR-R01b | Chain cancellation on unit removal | Teardown producers | 2026-10-14 assessed lifecycle cancellation; all teardown roots not followed | Trace removal to event cancellation |

No implementation comparison was requested. Pending observations limit confirmation, not verified supplied-input arithmetic or requests.
