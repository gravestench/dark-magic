# Party gold sharing

**Description**

When a player in a party picks up gold, the game can divide the pile among the party members on the same map. The pile is divided evenly by the number of living members there, rounding down, and the player who picked it up keeps the remainder along with any share another member could not carry. Each player can hold only a limited amount of gold that grows with character level, and gold nobody can hold is handed back for dropping near the picker. Unlike experience sharing, gold sharing has no distance limit, and in some cases, such as when the picker has no party, the picker keeps the whole pile.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: GPS-R, GPS-Q, GPS-V, GPS-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/player-party-gold-sharing.md)
Open evidence gaps: [gap file](../plans/gaps/player-party-gold-sharing.md)

## 1. Scope

This contract begins when a player picks up a gold pile of a supplied amount. It covers the owner and party bypass, the party count, the even split and remainder, each recipient's capacity against the level-based gold cap, the add requests and the returned excess. Actual stat storage and excess-gold recreation have integration gates; drop ownership, visibility and pickup timers remain a bounded unanswered question. [Party experience sharing](player-party-experience-sharing.md) owns kill experience and the same-level party iterator consumed here (PXL-R04a). Pickup validation, item removal and notifications outside the amount handoff are not defined here.

## 2. Trigger and ownership

D2Game computes these server results in its pickup and party paths. Picking up a gold item reaches the gold pickup caller, which selects the picker-only or party split, requests each capacity-limited add and receives the excess, then requests excess-gold recreation when the excess is positive. The companion records the dispatch chain and output-request locations. Operations run once per gold pickup, rather than every frame. Complete stat propagation, recreation results and message delivery remain unestablished.

## 3. Inputs and data

Integer notation consumes foundations section 1. Unless specified otherwise values are signed 32-bit integers; low32 retains 32 bits, i32 interprets them as signed, and trunc rounds toward zero. Stat reads use layer 0 and RLI-R02a/RLI-R02b. Getters do not establish aggregation or mutation.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| picker, recipient | Optional units: type, mode, identity | Logical units; type/mode i32 | Player type 0; modes 0/17 are the player death/dead modes |
| gold | Stat 14 `gold`; amount; picker; ordered party records; ownerExceptionCondition and owner lookup result | i32 / logical records | Owner exception requires resolved player |
| level for the cap | Stat 12 `level`, effective read | i32 | Read through RLI-R02a/RLI-R02b; no independent clamp |
| party eligibility | member.aliveExclusionCondition; player data, party id, party record, ordered member identities | Boolean; unsigned 16-bit party id; logical records | 65535 means no party; exclusion producer PXL-Q04, owner-condition producer Q01. `aliveExclusionCondition` is the shared name from [core-game-data-fields](core-game-data-fields.md) section 3, the same condition MSD-R05/MSD-R21 test; it implies no party-specific meaning |
| rooms | Room presence, room level id | Level id i32 | Same-level enumeration and recipient room test |
| dependency operations | Add/replace stat, recreate excess gold | Supplied operations with ordered logs | Requests distinct from stored effects; Q02 |

| Mapping or constant | Origin | Binding |
|---|---|---|
| Stats 12/14 and death/dead mode identities | Table names GDF-R02/GDF-R05/GDF-R06; selections engine code | Code selects fixed numeric ids/modes; cited data records supply names |
| Gold multiplier 10000 | Engine code | Fixed by original code |
| ItemTypes row identity 4, ItemType `Gold`, Code `gold` | GDF-R07 | Original-profile item-type mapping; pickup selects the gold path. Recreation separately selects the item code `gld ` in R04 |

No vector requires a shipped table cell; caps, levels and amounts below are supplied inputs.

## 4. Rules (execution order)

**GPS-R01** Amounts and requests [verified in 1.13c code (static)]. Gold bypasses party sharing when ownerExceptionCondition and its owner lookup yields a player, or picker has no usable party. Otherwise count n with PXL-R04a's same-level iterator and an alive callback: exclusion/type/death-mode checks, no distance or extra state-7 test. n<=1 or failed party-record lookup selects picker-only.

For n>1, share=trunc(amount/n). Visit ordered party identities. A found other member requires a room on picker's map level. This pass checks the picker's mode is neither 0 nor 17, not recipient liveness/exclusion. Calculate accepted capacity below and request recipient gold add; set remaining=i32(low32(remaining-accepted)). Finally calculate picker's capacity on remaining and request its add. Picker-only uses the whole amount. Return excess; positive excess requests R04 recreation.

Capacity: proposed=i32(low32(gold+amount)); initially accepted=amount, excess=0. If signed proposed>cap, accepted=i32(low32(cap-gold)), excess=i32(low32(amount-accepted)). Add guards compute i32(low32(gold+accepted)); negative result or, for a player, signed result>cap requests replacement with 0; otherwise request addition of accepted. These are requests, not a mutated stat-store guarantee. For ordinary nonnegative gold<=cap without overflow, accepted=min(amount, cap-gold).

**GPS-R02** Gold cap [verified in 1.13c code (static)]. Missing unit/list or unavailable level definition yields 0. Otherwise read effective level through RLI-R02a/RLI-R02b and return i32(low32(level*10000)), without an independent level clamp.

**GPS-R03** Owner condition meaning [unresolved]. Supplied condition/lookup tests are established; whether a particular player-dropped pile carries them remains Q01.

**GPS-R04** Storage and recreation [unresolved]. Actual additive/replacing stat propagation remains Q02. Positive excess reaches a path selecting item code `gld ` and chunks no larger than 2000000000 near picker room/position. Complete placement, creation, RNG and success/failure effects remain Q02; a request does not assert created items.

**GPS-R05** Drop ownership/timers [unresolved]. Complete ownership, visibility, pickup eligibility and timer policies remain Q03. Located pickup/validation is not a complete absence proof.

**GPS-R06** Direct random draws [verified in 1.13c code (static)]. Party counting, local supplied-input split/capacity calculations and add/recreation request routing perform no direct random draw. Dependency results are supplied. Actual stat propagation and excess-gold recreation interiors are excluded; their draws are not established here.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| GPS-R01 | Gold requests | verified in 1.13c code (static) | Q02: storage/recreation | GPS-V01, GPS-V02 |
| GPS-R02 | Gold cap | verified in 1.13c code (static) | none with supplied getter | GPS-V02 |
| GPS-R03 | Owner meaning | unresolved | Q01: pile producer | none |
| GPS-R04 | Storage/recreation | unresolved | Q02: actual effects/RNG | GPS-V02 request only |
| GPS-R05 | Drop policy | unresolved | Q03: complete behavior | none |
| GPS-R06 | Direct draw boundary | verified in 1.13c code (static) | Q02: excluded interiors | GPS-V03 |

## 5. Outputs and side effects

| Result or requested effect | Rules |
|---|---|
| Party count, share, accepted/excess amounts and supplied cap results | R01/R02 |
| Recipient/picker stat-14 add or zero-replacement requests; excess return | R01; actual storage Q02 |
| Positive-excess recreation request with picker/game/amount | R01/R04; creation/placement/draws Q02 |
| No direct seed update within local stage | R06 |

No persistence or delivered-message guarantee is established for gated interiors.

## 6. Branches and edge cases

A resolved player owner exception or unusable picker party selects picker-only before counting; a count of one or a failed party-record lookup also selects picker-only. The divisor counts alive members without any distance test, while recipients need only presence and a room on the picker's map level, so a dead member can receive a share; the recipient pass requires the picker not to be in mode 0 or 17. Division truncates toward zero, including negative supplied amounts; the remainder and unaccepted shares go to the picker. Capacity uses wrapping i32; a negative add total requests zero replacement. The cap has no level clamp, so a negative effective level gives a negative cap. Excess requests do not prove creation or draws. Classic/custom-data equivalence is not established beyond Expansion 1.13c and supplied-input operations.

## 7. Open questions

| Id | Kind / status | Assertion and evidence needed |
|---|---|---|
| **GPS-Q01** | outside-boundary / open | Gold ownerExceptionCondition producer: trace writers/pile creation and distinguish a player drop; supplied test and owner lookup available. The collection-exclusion producer is PXL-Q04 |
| **GPS-Q02** | integration / open | Actual stat-14 add/replacement propagation (RLI-Q03/STL-R06c) and excess-gold recreation: placement, creation, success/failure and RNG. Supplied arithmetic/requests available; read specific interiors and reproduce attributable cases |
| **GPS-Q03** | behavior / open | Complete drop ownership, visibility, pickup eligibility and timer policies. Root/output anchors, complete callee reading and distinguishing vectors needed |

Every verified behavioral rule still lacks matching gameplay observations for this contract. This observation gap is separate from integration gates and appears in the [gap file](../plans/gaps/player-party-gold-sharing.md).

## 8. Test vectors

### 8.1 Executable vectors

All vectors are supplied-input stage fixtures with Seeds: no draw; Draws: none; runtime_observed: no. Named dependency operations log requests in order; none mutates original storage. Defaults: own/effective values agree. Party inputs have present same-level rooms and alive mode 1. Every case starts fresh.

- **GPS-V01** Gold remainder (R01; boundary: supplied-input stage)
  - Inputs: amount 1000; same-level alive P/A/B, each gold 0, cap 100000; other order A/B.
  - Seeds: no draw.
  - Draws: none.
  - Trace: n=3, share 333, remaining 1000-333-333=334.
  - Outputs: add(A, 333), add(B, 333), add(P, 334), excess 0, no recreation.
  - Discriminates: discarding remainder gives P=333.
  - runtime_observed: no

- **GPS-V02** Gold cap/excess/liveness (R01/R02/R04; boundary: supplied-input stage)
  - Inputs: amount 1001, same-level alive P/A, effective level 10, caps 100000, P gold 0, A gold 99900, no owner exception, record P/A; operation log.
  - Seeds: no draw.
  - Draws: none.
  - Trace: n=2, share 500, A accepts 100, remaining 901, picker capacity then excess.
  - Outputs: baseline add(A, 100), add(P, 901), excess 0; storage Q02.
  - Discriminates: unused shares return to picker; count/recipient liveness differs.
  - (a) Inputs: P gold 99950. Outputs: add(A, 100), add(P, 50), excess 851, recreation(P, 851); no created-item/draw assertion.
  - (b) Inputs: record P only. Outputs: n=1, picker-only add(P, 1001), excess 0.
  - (c) Inputs: amount 1000, record P/A/B, P/A alive, B mode 17, all gold 0/cap 100000. Outputs: n=2, share 500, add(A, 500), add(B, 500), add(P, 0), excess 0.
  - (d) Inputs: amount 1000, record P/A/B, P mode 17, A/B alive, all gold 0/cap 100000. Outputs: n=2, share 500, no other adds, add(P, 1000), excess 0.
  - (e) Inputs: isolated cap, present unit/list, ordinary NRF getter inputs, effective level 10, own level 20. Outputs: 100000. Independent missing-unit/missing-list cases return 0.
  - (f) Inputs: isolated capacity amount 100, gold 2147483647, cap 100000. Outputs: signed proposed=-2147483549, accepted 100, excess 0; add guard requests replace(gold, 0) for negative total.
  - (g) Inputs: amount=1000, ordinary alive P/A party with gold 0/cap 100000, ownerExceptionCondition=true and supplied owner lookup returns player Q distinct from picker P. Draws: none. Trace: resolved player selects picker-only before counting. Outputs: add(P,1000), excess=0, no A request.
  - (h) Inputs: amount=1000, ordinary alive same-level P/A party, each gold 0/cap 100000, ownerExceptionCondition=true, supplied owner lookup returns a monster. Draws: none. Trace: exception requires player type; count n=2 and share=500. Outputs: add(A,500), add(P,500), excess=0.
  - (i) Inputs: amount=1001, player picker P gold 0/cap 100000, no usable player party data, no owner exception. Draws: none. Trace: unusable party selects picker-only. Outputs: add(P,1001), excess=0.
  - (j) Inputs: amount=1000, record P/A/B, P/A ordinary same-level alive players, B has no room, all gold 0/cap 100000. Draws: none. Trace: P/A give divisor n=2; recipient A accepts 500; B's missing room omits its request; picker retains 500. Outputs: add(A,500), add(P,500), excess=0, no B request.
  - (k) Inputs: isolated picker-only allocation stage, supplied amount=-100, player P gold=200/cap=100000. Draws: none. Trace: proposed=100<=cap, accepted=-100/excess=0; add guard sees nonnegative total 100. Outputs: add(P,-100), excess=0, no recreation request. No original negative-valued gold item is claimed.
  - (l) Inputs: isolated gold-cap calculation, player P with usesResolvedStats=true, own stat 12/layer 0=10, resolved stat 12/layer 0=-1, definition count 13, level definition available with minimum adjustment disabled. Draws: none. Trace: effective read selects -1, low32(-1*10000)=-10000. Outputs: cap=-10000; no implicit zero-level clamp.
  - (m) Inputs: isolated party allocation stage, supplied amount=-1000, ordered same-level alive P/A/B with valid party/rooms, each gold=1000/cap=100000, no owner exception. Draws: none. Trace: n=3, trunc(-1000/3)=-333; A/B each accept -333 with nonnegative resulting gold 667; remaining=-1000-(-333)-(-333)=-334, and picker's resulting gold is 666. Outputs: ordered add(A,-333), add(B,-333), add(P,-334), excess=0, no recreation request. No original negative-valued gold item is claimed. Discriminates: floor division would allocate -334/-334/-332; picker-only would omit both recipient requests.
  - runtime_observed: no

- **GPS-V03** Local seeds (R06; boundary: supplied-input stage)
  - Inputs: the GPS-V02(a) gold run (picker P gold 99950, recipient A gold 99900, amount 1001, caps 100000, record P/A, no owner exception). The five seed-owning objects of PXL-V15 keep its initial seeds M={lo:1, hi:2}, V={lo:3, hi:4}, P={lo:5, hi:6}, H={lo:7, hi:8}, game={lo:9, hi:10}. Supplied dependencies only log; the run stops at the recreation request.
  - Seeds: no draw.
  - Draws: none.
  - Trace: count, split, capacity and request routing reach the recreation request without a local random operation; dependency interiors excluded.
  - Outputs: all ten seed words unchanged, direct draw count 0; no claim inside recreation or stat propagation.
  - Discriminates: a local split roll changes seed or draw count.
  - runtime_observed: no

### 8.2 Observation vectors

All remain pending; the [companion procedures](../research/spec-evidence/player-party-gold-sharing.md#observation-procedures) state setup and measurements.

- **GPS-O01** Gold near/far/dead/other-level/cap controls: expect GPS-V01/GPS-V02 amounts; measure storage/excess creation separately.

## 9. Acceptance and dependencies

Partial acceptance covers verified supplied-input assertions in V01-V03: owner/party bypass, alive count and even split with remainder, recipient room and picker-mode gates, capacity/excess and add guards, the level-based cap, the excess request and no direct local draw. Synthetic logs do not establish original mutation. Full reproduction additionally requires resolving R03/R04/R05 at their stated boundaries, complete dependency-backed stored gold effects and excess recreation, and attributable observations for all verified behavioral rules. A verified-subset pass is partial acceptance. User approval is explicit and separate.

**Depends on:** PXL-R04a same-level party iterator (solo-callback guards, party record lookup, ordered registry/map-level visiting); RLI-R02a/RLI-R02b stat reads, with RLI-Q03/STL-R06c actual-write limits. GDF-R02/GDF-R05/GDF-R06/GDF-R07 supply named stat/mode/item-type identities. Supporting producer/schema records do not establish complete integration. Owners of the cited RLI ids: [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** no known consumer. Monster scaling cites R01 only for the shared name of the exclusion condition.

## 10. Related specs

- [Party experience sharing](player-party-experience-sharing.md): same-level iterator, collection-exclusion producer and experience split.
- [Monster scaling](monster-scaling-by-players-and-difficulty.md): MSD-R21's iterator and alive-callback record, reused as evidence for the party count.
- [Timed state lifecycle](state-timed-state-lifecycle.md): propagation integration limits.

## Appendix A. Provenance

Evidence build: Lord of Destruction Expansion 1.13c. Original D2Game/D2Common identities are in the [source manifest](../research/source-manifest.md). Compatible static readings of the pickup dispatch retain 2026-10-11/12 dates; complete caller, split, capacity, add and cap checks retain their 2026-10-15 dates and limits in the [public companion](../research/spec-evidence/player-party-gold-sharing.md). No new gameplay observation is claimed. The companion maps semantic operations to original locations, records corrections and pending procedures. Status remains draft and does not imply approval.
