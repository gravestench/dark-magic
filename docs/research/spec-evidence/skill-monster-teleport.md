# Evidence: monster teleport

Contract: [monster teleport](../../specs/skill-monster-teleport.md). Outstanding work: [gaps](../../plans/gaps/skill-monster-teleport.md).

## Identity and method

All new readings dated 2026-10-29 use original LoD Expansion 1.13c Windows x86. `LOD113C-D2GAME` is identified by the [manifest](../source-manifest.md): SHA-256 `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`, 1138688 bytes. The newly read original file matched that fingerprint. Available byte-matching coverage was checked before reading instructions; neither teleport entry has byte-equivalent C. Disassembly and Microsoft original-image listings establish the reads below. Private source locations, tool records and raw captures are separately preserved outside the repository.

The [dispatch companion](skill-cast-start-and-do-dispatch.md) supplies compatible 2026-10-12/15 action-dispatch evidence. Its dates and gated monster-frame/AI boundaries are retained; this pass does not renew or complete them. No earlier teleport-specific companion or observation was found in the repository search.

## Initial pass roots and function records (historical limits)

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | 0xA19A0; table 0x1074A8 | none established | Action dispatcher and action table | MTP-R01 to MTP-R07 | Reused dispatch contract; fresh original table read confirms slot 98 points to 0xB46B0 and slot 129 to 0xDBD0 |
| LOD113C-D2GAME | 0xB46B0 | none established | Direct teleport action | MTP-R04/MTP-R06 | Complete local entry/body read; called operation effects unresolved |
| LOD113C-D2GAME | 0xDBD0 | none established | Target-aware teleport action | MTP-R03/MTP-R05 to MTP-R07 | Original-image entry through 0xDD42; validation through first call establishes MTP-R03 |
| LOD113C-D2GAME | 0xDB00 | none established | Attachment cleanup requested by attached-caster path | MTP-R05/MTP-Q03 | Local body read; original label unrelated to proven behavior; state/mode/event callees unread |
| LOD113C-D2GAME | 0xD040 | none established | Attachment transition request | MTP-R07/MTP-Q03 | Call site located; full body/callees pending |
| LOD113C-D2GAME | 0x1280 | none established | Current/adjacent-room rectangle search | MTP-R04/MTP-Q02 | Local body read: signed half-open room tests and ordered adjacent-room loop; imported collection getter not verified here |
| LOD113C-D2GAME | 0xE19C0 | none established | Shared relocation operation | MTP-R04/MTP-R05/MTP-R07 | Body inspected; room/placement/state/event/network callees incomplete, no complete relocation contract claimed |

Descriptions are analyst descriptions. Import thunks are not original exports. No code identity is inferred from a name.

### Entry boundary and local decisions

The dispatch table's slot 129 explicitly points to 0xDBD0. A larger disassembly extent starts at 0xDB00 and includes it after a return/padding sequence. Therefore the first routine's analyst label is not the identity of the teleport entry. Original-image instructions at 0xDBD0 independently establish the entry and control flow.

MTP-R03: at 0xDBD0 identity is loaded and rejected when negative; signed count comparison rejects equality or above. Record stride is 0x23C. At 0xDC06 the computed record result is tested by the zero branch at 0xDC08, which also reaches return 0; no record field is read on that branch. An in-range absent record therefore describes malformed storage, not a normal GDA-R11a record. At 0xDC0E the word at record offset 0x80 is tested as signed, then sign-extended and compared with the state count. All rejection branches reach 0xDD37, clear the result and return without a call. The first relationship lookup begins at 0xDC2D/0xDC31. Thus the verified failure slices do not depend on interpreting any callee.

Selector 98 reads its path target coordinates using D2Game thunks 0xAac6/0xAad8, obtains a room through 0xA478 and calls 0x1280. A monster-only fallback follows data from the unit's monster data and a nested relationship record into 0xE03A0, then requests its room. It calls 0xE19C0 with both trailing controls zero. On nonzero return it calls thunk 0xA4DE, sets unit synchronization bit 0x10000 and returns 1. Complete meanings are not inferred from the bit or labels.

Selector 129 first calls 0xE0A20 (relationship lookup), 0xE07A0 (target refresh) and thunk 0xA598 (path target). A self-target becomes absent. With a caster relationship it calls 0x9FA30 for coordinates, then 0xE19C0 and 0xDB00, and returns 1 regardless of the first two results. Without that relationship and without target it calls 0xB46B0, forwarding its return. With a target it checks 0xE0A20 and 0xDFA90; both must return zero. It obtains position via 0x1100/0x1130 and room via 0xA478, requests 0xE19C0 with controls 1 and 0, and on success calls 0xD040 and returns 1. These are actual located calls, not complete callee verification.

### Semantic layout map

| Contract term | Original investigation mapping |
|---|---|
| callback context | game ECX, caster EDX, skill identity/level on stack; signed EAX return |
| loaded skill count/records | shared data root offsets 0xBA0/0xB98; record stride 0x23C |
| teleport state selector/state count | skill signed word 0x80 / shared root 0xC4 |
| path / synchronization condition | caster 0x2C / caster 0xC8 bit 0x10000 |

Offsets explain the evidence only; they are not required storage layouts. The signed consumed field read is established by the callback instructions. The shared GDA-R11a contract does not itself name that field. The continuation below resolves the historical MTP-Q01 through the original loader and selected live scalar reads.

## T01: effective Skills cells

The observed LoD Expansion 1.13c data profile reported original Game executable SHA-256 `74fe9c092a521f7710392548c82f81544e531107db2358617e83818874db40a2`. On 2026-10-29 its Storm loader read `data/global/excel/skills.txt` in contiguous ranges [0,60000), [60000,120000), [120000,173535). Success and returned counts 60000/60000/53535 establish complete file/header/row coverage. Joined effective-file SHA-256: `465c28e83657540e0d7ae5b0de826d5657be7fba1aa539a85c3ca3aeea34b498`. Columns were parsed by header names, rows selected by `skill`; `Id` is the explicit decimal identity column. Only cells needed here are published.

| Named row | Column | Raw cell | Interpretation / claim |
|---|---|---|---|
| MonTeleport | Id | `184` | Explicit decimal identity for C01 |
| MonTeleport | srvdofunc | `98` | Decimal selector; MTP-R01 |
| Imp Teleport | Id | `330` | Explicit decimal identity for C01 |
| Imp Teleport | srvdofunc | `129` | Decimal selector; MTP-R01 |
| Imp Teleport | aurastate | `attached` | Linked state name only; MTP-R02 |
| Baal Teleport | srvdofunc | `98` | Supporting relationship to the direct route |
| Baal Clone Teleport | srvdofunc | `98` | Supporting relationship to the direct route |

No known data override was reported by the reported vanilla data profile; a complete override inventory and physical supplying MPQ were not captured. The executable identity does not itself authenticate archives. The read proves effective text cells, not live compiled values or gameplay execution. No data row or full table is distributed publicly.

## Runtime setup and MTP-O01 procedure

2026-10-29: no teleport handler invocation or gameplay effect was captured in the initial observation attempt. This supplies no negative gameplay result and confirms no executable vector; matching gameplay remains pending.

For MTP-O01, use an original 1.13c Expansion character able to encounter imps, in Normal difficulty in Act V. Record class/level/equipment, location, game frame and target before each action. Observe an imp's teleport through action selector 129 and the direct selector-98 route separately. Trace the action entry, 0xE19C0 result and 0xDB00/0xD040 entry only as reached. Record path target, selected unit, both relationship lookup results, actual position/room before and after and final callback return. Distinguish the relationship-present path (final return 1 even if supplied operation results are zero), no-target delegation and target-predicate rejection. Natural gameplay alone cannot force every failure case; untriggered cases stay unobserved, and static gates must close before promoting a rule.

## Initial review and limits (historical)

The source pass establishes the exact no-call validation failures and named text cells. Complete AI handoff, target/relationship semantics, shared relocation, attachment transitions and gameplay records remain gated. At this initial source-pass stage the fresh review was pending. The final applicable review ledger below now records its completion; the evidence gates remain unchanged.


## C01: narrow compiled selector corroboration

Read-only inspection of another reported LoD Expansion 1.13c data profile on 2026-10-29 recorded the Game.exe manifest hash; loaded DLL hashes and a complete loose override inventory were not captured. The live Skills collection count was 357. Record indices 184 and 330 had stored identities 184 and 330, start selectors 0 and 0, and action selectors 98 and 129 respectively. These match T01's explicit `Id` and `srvdofunc` cells for MonTeleport and Imp Teleport. This supports only those scalar selector fields in the reported profile; it neither verifies numeric attachment state nor executes a teleport.

Investigation mapping: D2Game RVA 0xF829C references the shared data root, whose 0xB98/0xBA0 fields hold Skills storage/count; stride 0x23C, identity/start/action words at 0/0x2C/0x2E. Successful read-only live records covered 48 bytes at each selected record. Private collection and raw read records remain in the separately preserved audit.


## Mechanical and self-review record

On 2026-10-29 the three owned spec/companion/gap trios were checked together with the repository spec linter: 0 errors and 0 warnings. Self-review checked source identities, exact per-rule gates, supplied dependency boundaries, vector arithmetic and public/private separation. This mechanical/self-review entry preceded blind review; the final applicable ledger below is the current review status.




## Fresh blind review MTP1 — full, 2026-10-29

Counts: MISMATCH 2; GUESS 0; GAP 2; DISCRIM 1; GATED 4; COVERAGE 1; HARDEST 4. The reader used the contract/dependency material only. This is the first full review and does not approve the spec.

| Finding | Outcome and rationale |
|---|---|
| MISMATCH: unit-type dependency incorrectly named RLI-R01 | Edit: cite the owning record-layout contract section 3.3 Unit types rather than a stat-reader rule. |
| MISMATCH: GDA-R11a does not establish aurastate field-name association | Edit/gate: section 3 names only the signed consumed state field; MTP-Q01 explicitly requires original loader mapping and numeric linked correspondence. The verified consumer no longer depends on a guessed column name. |
| GAP: null prepared-record branch missing from companion | Edit: record original 0xDC06/0xDC08 zero-result branch, its no-field-access boundary and failure return. Original-image prefix rechecked 2026-10-29. |
| GAP: in-range absent record fixture inconsistent with normal complete storage | Edit: MTP-V01(c) now uses identity 0 with explicitly synthetic malformed absent storage. Normal GDA valid-index storage is distinguished in the rule/preamble. |
| DISCRIM: negative signed vs unsigned reinterpretation can reject equivalently | Edit: remove the false distinction. Test wrong clamp-to-zero against a valid record-0 context and empty observed access/entry logs. An equivalent unsigned bound implementation is expressly accepted. |
| COVERAGE: only failures allow always-reject implementation | Edit: add MTP-V02 with signed state 0, recording valid continuation through the first relationship-lookup entry and asserting no final callback return/effects. Instructions through 0xDC31 establish this boundary without interpreting the callee. |
| HARDEST: state/name binding | Edit/gate as MTP-Q01 above; no guessed state id. |
| HARDEST: target and relocation dependencies | No behavior change: MTP-Q02 and R04-R07 gates already name this missing output boundary. |
| HARDEST: attachment effects | No behavior change: MTP-Q03 already gates cleanup/transition. |
| HARDEST: gameplay confirmation | No behavior change: MTP-Q04 and observation record explicitly establish no gameplay capture. |
| GATED: four documented questions | No promotion: all four remain explicit. Only the supplied validation/first-entry boundary is verified. |

Rule/input/vector/acceptance changes require a fresh full reader. MTP1 is historical for the earlier contract; the subsequent MTP2 record below is the final applicable review.


## Historical blind review MTP2 — full, 2026-10-29

Normalized counts: MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; GATED 4 documented rule groups (R04-R07 under Q01-Q04); COVERAGE 0 uncovered deficiencies; HARDEST 4. The raw COVERAGE list described six successfully covered cases, not six missing cases. All reported comparisons were actionable-defect free.

| Finding or difficult boundary | Outcome and rationale |
|---|---|
| Synthetic in-range malformed record | No change: V01(c) explicitly supplies malformed absent record 0, while R03 and the fixture distinguish complete valid storage. Resolution and field access are separate. |
| Negative clamp versus unsigned equivalent bounds | No change: V01 rejects wrong clamp-to-zero using a valid record-0 context. Equivalent signed/unsigned bound formulations are explicitly not required to differ. |
| Signed state range endpoints | No change: V01 covers negative and count-equal rejection, and V02 covers valid state 0. The state field name/linked binding remains Q01. |
| Valid first lookup entry only | No change: V02 ends at the first relationship-lookup entry and asserts neither its result/effects nor a final callback result. |
| Existing gated rules/questions | No promotion: Q01-Q04 and R04-R07 remain explicit; review clarity does not verify missing effects or gameplay. |

That blind-implementer pass completed the earlier validation-only contract. The continuation below changes behavior rules and vectors; its fresh full review was pending at that stage; the current ledger below records completion. Status remains draft.

## Teleport continuation — 2026-10-29, current evidence

This continuation supersedes the earlier callee-unread limits for the exact local stages below. MTP1/MTP2 remain historical reviews; the initial observation attempt supplied no gameplay capture. No gameplay rule is runtime-confirmed. Original-image-checked reconstruction coverage was consulted before instruction follow-up. Most principal helpers have no byte-equivalent higher-level reconstruction; original-image listings corroborate their instructions. Decompiled views supplied leads only. The available reconstruction for Common 0x1330 is an assembly transcription; it was found after the initial instruction locator read and does not supply an independent higher-level proof.

### Newly closed code stages

| Module | RVA / original import | Established stage | Retained boundary |
|---|---|---|---|
| D2Game | 0xE0A20; 0xE03A0 | Relationship-present gate and game-registry kind/id lookup; valid kinds 0..5 | Malformed kinds excluded; no producer identity inferred |
| D2Game | 0xE07A0; 0x9FA30 | Saved-key refresh clears stale cache without replacement/key erasure; coordinates written before boolean result | Supplied valid game/path and registry |
| D2Game | 0xDFA90 | Player/monster kind, exclusion and mode life test | No broader death-state meaning |
| D2Game / D2Common | 0x1280 / 0x3BA20, ordinal 10383 | Current room then ordered adjacent candidates, signed half-open rectangle tests | Rectangle sums wrap; no collision claim |
| D2Game | 0xB46B0; 0xDBD0 | Direct, linked, no-target and distinct-target requests/results/order | Supplied downstream results do not establish placement |
| D2Game | 0xE19C0; 0xE1610 | Monster relocation request, nearby controls, low-word coordinates, update/marker; synchronization immediately stops when monster path condition is clear | General collision/transfer/client work remains MTP-Q02 |
| D2Common | 0xE1E0 ordinal 10533; 0xDF30; 0xDA40; 0xDC80; 0x1330 | Reached same-room suffix centers coordinates, projects, clears movement/progression/scratch fields and returns 1 | Collision calls already returned normally; room-update-request and containment-guard clear |
| D2Common | 0x665C0 ordinal 10158 | Actual queued condition, room changed marker, next pointer and update-head writes; repeat collection no-op | Valid queue storage; no client delivery |
| D2Game | 0xD040; 0xDB00; 0x18A0; 0x18E0 | Fixed-family attachment/cleanup decisions; state/link/event/AI/death requests; interaction writes | Full stat propagation, AI/death execution remain MTP-Q03 |
| D2Game | 0x8E680; 0x8E900; 0x8E6F0; allocator 0x8E540 | Successful supplied event construction, exact keyed removal and persistent-queue prepend | Allocation failure and later processing not closed |
| D2Common | 0x33020 ordinal 10210; 0x32CA0 | Immediate counterpart kind/id stores, including zeros for unlink, before optional stat integration | State-98/list-stat propagation and final relationship condition remain MTP-Q03 |
| D2Common | 0x32330 ordinal 11027 | Health getter order and arithmetic-shift-before-percent, zero whole maximum | Supplied RLI stat results |

Original imports were resolved through the original export table, not thunk labels: D2Common ordinals 10983/10764 -> 0x34AB0/0x34AA0 (stored target unsigned words); 10331 -> 0x2FE10 (current room); 10392 -> 0x348F0 (target cache); 11163 -> 0x34910 (saved target key); 10336 -> 0x35040 (set cache, null preserves key); 10509 -> 0x4D840 (nearby wrapper to 0x4D140); 10589 -> 0x30F80 (collision footprint); 10945 -> 0x34510 (state request); 10596 -> 0x2FCC0 (monster synchronization condition). Available original-image-checked reconstructions for these getters/wrappers were read first. The nearby call passes mask 7177 twice, pattern 0, scan limit 50 and controls 0/1; the public semantic request retains the relevant mask/limit/control contract without claiming its search effects.

Important corrections: cleanup at 0xDC83 forwards the invocation skill identity, not level (stack depth and original call operands checked). The fixed helpers use state 143 independently of the validated Skills field. The direct fallback uses the AI control saved lookup context/key, then accepts the found unit current room without another rectangle search; no unverified owner relationship is inferred. The death request is D2Game RVA 0xDEEE0. Path room-update-request bit 1 and synchronization bit 2 are distinct. Original Common constants at 0x8D464/0x8D468 establish the two zero scratch resets.

### T02 / C02: linked field correspondence

Original Common Skills-loader instructions map the name string at RVA 0x90F30 (`aurastate`) to descriptor type 0x14, destination offset 0x80 and state-name linker at preferred VA 0x6FDEFF98 (descriptor instructions 0x63670/0x6367B/0x63689/0x63694). The loaded type register was checked through its preceding assignments. An earlier different descriptor at the same destination offset belongs to SkillDesc `desctextb6`; it is not the Skills field mapping. Callback sign extension establishes the consumer width. This closes the code-field/name association rather than attributing it to GDA-R11a.

A 2026-10-29 LoD Expansion 1.13c data read reported the same original Game.exe baseline hash as T01. A complete Storm effective States file read returned 20920 bytes, SHA-256 `898806c2b69e55e316ecfed48dde95aa2933829b94fdad234bdabc27bdbc23e1`; parsing by header names found row `state=attached`, explicit `id=143`. No full data row/table is published.

Read-only live table metadata gave Skills count 357/stride 0x23C and States count 185/stride 0x3C. One corrected 136-byte read from Skills record 330 established stored identity 330, action selector 129 and signed word 0x80 = 143. An 8-byte States record 143 read established stored identity 143. A prior offset read that began 0x18 bytes into the Skills record was recognized as misbased and discarded; it supports no field assertion. This captures only selected scalars. Loaded DLL hashes, complete loose override inventory and physical supplying archive remain uncaptured; the profile executable identity does not authenticate archives or loaded modules. Original file fingerprints supply static code provenance separately.

### Continuation observation status

2026-10-29 continuation observation: a level-1 Barbarian with no skills was in Act I. No eligible imp/direct teleport action occurred, no function trace was collected and no gameplay assertion was promoted. The successful table reads above establish only their named values and narrow compiled correspondence.

### Current review/check ledger

The continuation adds MTP-R08 through MTP-R14 and executable MTP-V03 through MTP-V13, and promotes only explicit local stages of R04-R07. Self-review checked the health fixture against whole-health truncation (2559/23296 raw becomes 9/91, percent 9, despite raw fraction exceeding ten percent), same-room 16.16/projected arithmetic, identity-vs-level event key and supplied-effect boundaries. Fresh full smaller-reader review was pending at this handoff. The current ledger below records completion; prior MTP2 completion applies only to the earlier contract.

## Final applicable continuation review — full, 2026-10-29

The coordinator relayed the fresh smaller-reader full review: MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; COVERAGE 0 uncovered deficiencies; GATED 3 documented questions (MTP-Q02/MTP-Q03/MTP-Q04); HARDEST 3. All thirteen executable vectors were independently recomputed, including requested branches, argument identity, actual effect prefixes and arithmetic, and found coherent. No contract, vector or gate change was recommended or made.

| Finding or difficult boundary | Outcome and rationale |
|---|---|
| Mismatch, guess, gap and discrimination checks | No change: all actionable counts were zero. |
| Executable coverage | No change: all thirteen vectors were coherent, with no uncovered deficiency reported. |
| MTP-Q02 collision, cross-room transfer and nontrivial synchronization | No change: the exact request and reached-suffix boundaries remain verified; these integration effects remain gated. |
| MTP-Q03 link-stat propagation, allocation failure, event processing, AI and death execution | No change: local decisions, argument identity and effect prefixes remain verified; later effects remain gated. |
| MTP-Q04 matching gameplay observation | No change: data-read success does not establish a teleport gameplay capture. |
| Resolved MTP-Q01 | No change: original loader mapping and selected effective/compiled scalars retain their stated profile limits. |

The blind-implementer stopping rule is complete for this continuation. Earlier MTP1/MTP2 reviews and the initial pending handoff remain historical. Status remains draft; evidence gates and user approval are separate. Final mechanical verification after this ledger edit: teleport spec/companion/gap lint, 0 errors and 0 warnings.

## Collision, room transfer and attachment execution — 2026-10-30

The original D2Game and D2Common files were freshly fingerprinted to the existing manifest identities: D2Game SHA-256 `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`, 1138688 bytes; D2Common SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`, 679936 bytes. Compatible 2026-10-29 local proofs remain dated as such; these new readings close named portions of their remaining integration questions, without changing data-profile or gameplay identity by inference. Available original-image-checked higher-level reconstructions were read first. Other applicable instruction bodies were read and byte-encoded regions decoded against the original image. Decompiled views supplied leads and cross-checks only.

| Artifact | Function RVA / ordinal | Supported assertion | Exact new boundary |
|---|---|---|---|
| LOD113C-D2COMMON | 0x30F80 / 10589; 0x13B0 | MTP-R15 monster footprint | MonStats signed linked extended identity, bounds/records and signed byte footprint consumer |
| LOD113C-D2COMMON | 0x56C10, descriptor 0x56CAE..0x56CC6 | MTP-R15 field name | Original string RVA 0x8FA40 `SizeX`, type 4 and destination 8; consumer sign-extends that byte |
| LOD113C-D2COMMON | 0x4D840 / 10509 -> 0x4D140 | MTP-R15 search | Null reference point skips line-of-sight buffer; original point and complete radius-1-through-49 perimeter loops, stable minimum-score choice |
| LOD113C-D2COMMON | 0x4CE00; 0x4C570; 0x4C5C0; 0x4C530 | MTP-R15 shape queries | Point/cross/square and exact missing-grid/invalid-shape sentinels; unhandled cross-boundary codes return mask |
| LOD113C-D2COMMON | 0x4B3F0; 0x4AA90; 0x4AC40; 0x4B510 | MTP-R15 split/grid reads | Inclusive rectangle subdivision, recursive right/bottom remainders and actual masked cell reads |
| LOD113C-D2COMMON | 0x4BCF0; 0x4BE90; 0x4BB30; 0x4BB70; 0x4BBB0; 0x4B8F0; 0x4B930; 0x4B970 | MTP-R16 map writes | Pattern table, point/cross/square bit clears/sets, additional center/cross flags |
| LOD113C-D2COMMON | 0x4B280; 0x4B2F0; 0x4AB60; 0x4ABD0 | MTP-R16 rectangle writes | Actual OR/AND low-word cells and recursive portion handling, with unchecked rectangle backing storage explicit |
| LOD113C-D2COMMON | 0xDF30; 0xE1E0 / 10533; 0xDA40; 0xDC80 | MTP-R16 | Complete monster relocation, position guard, centering/projection, transfer call, containment assertion and movement reset |
| LOD113C-D2COMMON | 0xD760; 0x66700; 0x66800; 0x66410; 0x66630 | MTP-R17 | Actual resident/update-chain unlink/prepend and room fields; duplicate/bounds assertion |
| LOD113C-D2COMMON | 0x2FE10 / 10331; 0x2FE70; 0x391F0; 0x3B650; 0x3BA60 / 10561 | MTP-R17 | Current-room clear, rectangle/count access, state-105/list-stat-172 population predicate and checked decrement |
| LOD113C-D2GAME | 0xE1610 | MTP-R10 / MTP-V15 | Distinct rooms with empty client collections: monster auxiliary field clear and path synchronization clear; nonempty-client effects still Q02 |
| LOD113C-D2GAME | 0x9D230; 0x9D0F0; original table entry 0x10E768 | MTP-R18 | Reached selector-16 reset/init/install path; table entry {1, init 0x155C0, think 0x166D0, auxiliary 0} |
| LOD113C-D2GAME | 0x155C0; pairs 0xFCEE8 | MTP-R18 | One caster seed step then candidate pairs {2,0}/{4,0}; writes AI parameter selection |
| LOD113C-D2COMMON | 0x4F080 / 10708 via Game 0xA3D0 | MTP-R18 | First matching skill-record identity/entry-owner -1 after sentinel; skill level not tested |
| LOD113C-D2GAME | 0x1220; 0x1320 | MTP-R18 | Initializer ordinary/extended record prerequisites, with missing results returning before think install |
| LOD113C-D2GAME | 0x8F070 -> 0x8EBC0 -> 0x769A0; table cell 0xFBC28 -> 0x9E8E0 | MTP-R19 | Persistent monster queue processing, type-5 gate bypass and skill/state/active-action routing |
| LOD113C-D2GAME | 0x1AE0; action slot 147 at 0x1076F4 -> 0xC170 | MTP-R19 | State record getter and interior action entry; relationship lookup followed by position copy; return 1 |
| LOD113C-D2COMMON | 0xDAB0 / 11121 via Game 0xB456 | MTP-R19 | Full counterpart fixed-point copy, integer point-change condition and room-transfer call; no collision mutation or movement reset |
| LOD113C-D2GAME | 0x8E420; 0x8E540; 0x8E680; 0x8E6F0 | MTP-R12/MTP-R19 | Pool/free-node linkage, unit chain, processing-time deferred removal and actual free-pool return; allocation failure has unchecked downstream dereferences, no safe decline inferred |

Analyst descriptions are not original exported names. Original export ordinals identify imported functions; direct callees have no export identity asserted. The action-147 entry at 0xC170 is inside a larger listing beginning at 0xC110 and independently rooted by the action table; the unrelated first routine name does not identify it.

### Search and physical transfer findings

The search reads initial point first. Its first perimeter phase enumerates the two x edges for each ascending y; its second phase enumerates the two y edges for interior ascending x. Each ring is finished before selection, uses a strictly-lower Manhattan score replacement, and the loop limit is exclusive. All-blocked original point plus radii 1..49 produces 9801 candidate tests. Original instruction checks include 0x4D2B0..0x4D55C and return paths 0x4D562..0x4D5B9. The null reference-point argument avoids the separate allocation/ray traversal operations on this teleport route; they are not silently supplied operations.

Pattern-reading shape and path-writing pattern are different consumers. Read shapes 0/1 use one cell, 2 cross, 3 square. Write patterns 1/3 use cross plus center 4096/8192; 2/4 use square plus cross 4096/8192; 5 uses cross only; 0/default writes nothing. Original cells are u16 and ordinary set/clear is OR/AND, with no occupancy reference counting. Point and cross helpers check cell-storage presence; rectangle helpers do not check it after metadata is present. A missing square backing allocation is consequently not assigned the point helper sentinel or a safe no-op.

The position guard can refuse position stores after collision writes; relocation still continues and does not implement rollback. After a real room change the saved old room and new room are distinct; both resident and update chains are maintained independently. The population helper consumes the first state-105 list and own stat 172, not unit aggregate; the decrement checks positive count. These facts support actual cross-room MTP-V15 rather than a supplied relocation-return fixture. Empty-client synchronization exercises the reached nonclear synchronization branch without claiming uninspected monster creation-message effects.

### Attachment initializer and persistent following

The fixed selector-16 table points to initializer 0x155C0. It advances the caster seed once even if neither tested skill entry exists. The post-step parity chooses slot 3 first when odd and slot 5 first when even; the first match stores index 2 or 4, with third parameter 0. The selected-entry helper starts after the skill-chain sentinel, compares signed record identity and exact entry-owner -1, and does not test level. These effects belong to the initializer, preserving the earlier statement that the teleport caller itself makes no direct draw. General alternate-handler, ordinary-reset and subsequent think branches retain Q03 until their own effects are closed.

Persistent monster dispatch does not remove a frame -1 event merely after calling it. Removal during its processing sets a deferred condition and actual unlink/free happens after return. Type 5 dispatch reaches the skill record, its signed aurastate, the States record and signed srvactivefunc, and the selected action table. It does not test active membership. Action 147 follows a resolved counterpart even if that state membership bit is off, copies full fractional coordinates rather than centers, and can transfer rooms. Its copy helper itself performs no collision-map mutation or movement-counter reset. These assertions are independent of later AI decisions or general frame-group ordering.

### T03 / C03: attachment active-action identity

T03 reuses the complete effective States file from T02, read on 2026-10-29 with its original hash/profile limits. Header-based parsing of the named row `attached` gives raw `srvactivefunc` cell `147`; decimal conversion yields action selector 147. The existing GDA States loader descriptor maps that named field to the signed consumer. A fresh owned original-profile LoD Expansion 1.13c read on 2026-10-30 found States count 185 and read the complete 60-byte selected record at index 143: stored identity 143 and signed srvactivefunc 147. This C03 read corroborates only those selected scalar fields. The launch baseline reported the same Game executable hash as T01; actual loaded-module/override/archive identities were not yet captured in this record. It is not a handler/event execution observation.

### Source-unit stat propagation — 2026-10-30

The complete local source-unit list path was followed from Common 0x33020  -> 0x32CA0 through allocation 0x38100, attach 0x3A7D0, replacing writer 0x3A280, propagation 0x39F70, resolved-array delta 0x389E0, exact/insertion search 0x382B0/0x38270, insert 0x38640, remove 0x381A0 and tracking 0x38550. Unlink follows 0x3A3B0 and ordinary free 0x3A9F0. Available byte-equivalent higher-level forms for allocation, exact lookup and free were read first; other applicable instructions and original-image decoding supply the remaining evidence. This closes R20 for the ordinary primary state-98 lists this producer creates, with extended bases and original selected stat definitions. It does not silently extend that proof to an unrelated supplied secondary/extended/custom-callback state-98 list.

Stat353/354 replacement modifies the list own array, then propagates new-minus-old. Under the captured zero controls, the ordinary list starts at its attached base, and each ancestor resolved array receives the wrapped delta; zero results are removed. No dependent rewrite, callback or tracking insertion is selected. Fresh attach of an empty list contributes nothing; ordinary unlink subtracts own entries before release. Related state 98 active/changed bits and update marking are independent of state 143 in the outer callback. Disguise-true integration retains STL-R07a; V19 supplies a nondisguise state rather than inferring false from an empty text cell.

### T04 / C04: original source-unit stat definitions

A complete owned original-profile effective `data/global/excel/itemstatcost.txt` read on 2026-10-30 returned 40109 bytes, SHA-256 `ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`. Header-name parsing selected named `Stat=source_unit_type`, explicit decimal `ID=353`, and `Stat=source_unit_id`, `ID=354`. Their named callback/minimum/operation/keepzero text cells were empty; this is not evidence that the compiled values are zero.

A separate successful 648-byte live read selected compiled records 353/354 from count 359 with stride 324, stored identities 353/354. For each record, flag bytes at offsets 4/5 were both 0 and derived consumer controls at 0x50 through 0x53 were all 0. Thus the exact no-callback/no-dependent-rewrite/no-keepzero/no-positive-marker path is corroborated from compiled inputs, not guessed from blanks. Full raw definitions are not published. The same original-profile executable-baseline identity and still-uncaptured loaded-DLL/override/archive limits as C03 apply. These are selected data scalars, not execution of link/unlink or a gameplay confirmation.

### Client merge and normal AI restoration — 2026-10-30

Full D2Game 0xE1610 instructions establish an unsigned-reference sorted merge of the supplied old/current client collections. Current-only clients reach 0x7B0A0 unless their resolved player is the moved unit; old-only clients reach 0x8A3E0 with six-byte command 0x0A, unit-kind byte and little-endian unit identity. Equal references advance both without player lookup. Existing room equality returns before monster auxiliary/synchronization clears. The tail-current-only player getter 0x4AD10 and equivalent inline branches preserve a nonabsent stale cache while returning the actual registry result, clearing the cache only on missing resolution. These assertions close recipient decisions; the complete arriving-monster serializers remain Q02.

D2Game 0x8A3E0 has no byte-equivalent higher-level reconstruction available. Its full instruction body and original image decoding establish append at used+4, signed capacity comparison against 512, exact-fit acceptance, recycled-list removal, 520-byte allocation request when no recycled block exists, whole-block zeroing and pending-head/tail/next linkage before copying all bytes. The allocator thunk 0xA2E6 resolves FOG ordinal 10045/RVA 0x1CD50. The queue path assumes its normal successful result; no safe allocation failure or transport delivery was inferred. V20 uses an actual recycled block and exact-fit existing tail, so neither case supplies allocator effects. Complete sends to newly arriving clients still require monster serialization and subsequent output transport.

The original 0x7B0A0 dispatch tables independently identify monster index 1: first branch 0x7B233, second 0x7B45C. The second reaches 0xF35C0; the neighboring player branch beginning 0x7B393 and its owner/inventory sends are not monster obligations. The monster creation serializer 0x8CA20 and mode serializer 0xF35C0 were read through their full bodies, but their bitstream, appearance, mode and later send dependencies are not thereby completed. No public payload/return rule was inferred from their analyst names.

D2Game 0x9D0F0 reads the ordinary record signed word 0x1E for selector 0 and bounds it to 0..147. Original table entry 0x10EF28 for identity 122 is {flags 1, initializer 0x15530, thinker 0x1A550, auxiliary 0}. The prior attachment entry 16 also has auxiliary 0, so restoration from its installed thinker takes the actual reset/command-clear path, not an alternate operation. Initializer 0x15530 is an independent interior entry within the 0x15520 listing; the higher-level reconstruction covers only the first 0x15520 routine and does not prove this second routine. Its own instructions and original bytes set control parameter 0x14 to-1 and return. The general reset already zeroed parameters 0x18/0x1C; no draw occurs. Full 0x1A550 normal Imp thinker instructions were read, but actual downstream movement/skill-mode effects remain Q03; installing that thinker is distinct from executing it.

### C05 / C06: fresh state and imp selector scalars

A new owned original-profile LoD Expansion 1.13c process supplied successful complete selected record reads on 2026-10-30. C05 selected States index 98 from count 185/stride 60, stored identity 98, consumed byte 0x12=0, hence disguise bit 0 clear. This corroborates the nondisguise state-98 branch from an actual compiled value; it does not infer false from a blank text cell. C06 selected MonStats index 492 from count 734/stride 424, stored identity 492, signed extended identity 492 at 0x18 and ordinary AI identity 122 at 0x1E. Selected signed skill slots 3/5 are 283/282, consistent with the effective named `imp1` row captured separately; this record does not observe their execution. The previous process's heap addresses were not reused. Reported original executable baseline remains the existing Game identity; loaded-DLL hashes, complete override inventory and physical archive origin remain uncaptured for these data records. No teleport/attachment gameplay was observed by these reads.

### Sustained continuation review triage — 2026-10-30

The fresh smaller full reader reported MISMATCH 0, GUESS 0, DISCRIM 0 and two actionable GAP findings. Both were accepted: Section 5 now lists the current-only spawn/update request independently of its still-unverified serializer effects, and R21/V20 specify the actual six-byte wire command and exact queued bytes. No operation result was supplied to close those serializers, and Q02/Q03/Q04 retain their existing integration/observation meaning. The resulting delta awaits a fresh focused review; this is not a completed stopping-rule claim.

| Coverage advisory | Triage |
|---|---|
| R07 life/kind/mode exclusions and relocation 0 | V05 already covered monster mode 12, resolved relationship and success/delegation. Added independent alive-exclusion, object-kind, player mode 17, monster mode 0 and relocation 0 cases to distinguish their stop/attachment boundaries. |
| R09 items in mode 1/2 | Added V06(c): a matching item cache is cleared for either mode, preserving its key and producing stored fallback coordinates. Existing stale-reference case continues to distinguish replacement from clear. |
| R16 patterns 0/2–5, zero destination and guard paths | Added V15(c) exact independently isolated set-pattern maps; existing full V15 supplies pattern 1 and V09 supplies the reached position suffix. Zero-destination/position-guard interaction has no new vector in this delta; rule already states its clear-only/guard ordering, with no invented rollback or safe default. |
| R17 containing-room, failed search, player counter | Added V15(d/e) no-transfer and guard-refused search with exact retained fields/counter reset. Player population behavior is a shared callee branch outside the present-monster teleport scope; no new player teleport claim or fixture added. Existing V15(b) verifies the monster's own-list stat 172 predicate and actual population changes. |
| R21 cache fallback and current-only recipient | Added V20(d) actual saved-context missing-registry result/cache clear and remove bytes; V20(e) stops at exact spawn-operation entry before serialization. Existing V20(a/b/c) preserves equality/no-lookup, pending-block boundary and same-room early-return checks. Full arriving serialization stays Q02. |

The three HARDEST checks required no rule change: stable search-ring tie/grid behavior is explicit in R15/V14, complete collision/resident/update effects in V15, and reciprocal own/resolved list deltas and unlink restoration in R20/V19. Source was not expanded during this review fix. Mechanical validation of the resulting trio: 0errors,0warnings.

### Plain arriving-monster sender closure — 2026-10-30

The reviewed R21/R22 delta received a fresh focused smaller-reader result with no actionable mismatch, guess, gap, discrimination or coverage defect; the three existing questions remained gated. That result applies to the prior 22-rule/21-vector contract. The subsequent R23/R24 and V22/V23 additions below require their own fresh review.

| Artifact | RVA / ordinal | Actual supported boundary |
|---|---|---|
| LOD113C-D2GAME | 0x7B0A0 kind-1 first entry 0x7B233 and second entry 0x7B45C | Plain monster sends creation, stores stat 328, scans eight skill-send bits, sends state/message chains, then mode update |
| LOD113C-D2GAME | 0xDFA90; 0xAF610 | Alive-exclusion/kind/dead-mode predicate and all-sixteen-zero component test; mode 1 plus clear exclusion reaches creation |
| LOD113C-D2GAME | 0x8CA20 | Reached plain mode 1 creation bitstream and actual 14-byte output, including absent appearance/owner/flag 64 list bits |
| LOD113C-D2GAME | 0x7A940; 0x7ADF0 | Empty active-state bitmap emits 8-byte command 170; absent unit-message chain returns without message |
| LOD113C-D2GAME | 0xF35C0; 0x8BFC0 | Mode 1, absent current skill/target: actual 10-byte command 109, then own stat 328 read/increment/store |
| LOD113C-D2GAME | 0x77060 | Arrival health byte uses whole health shifted 7/divided by whole maximum, default 128; distinct from cleanup percentage |
| LOD113C-D2GAME | 0xE07A0; 0x49A50 | Existing target refresh reused; possible target requires recipient membership in its room; V22 absent target avoids wider target modes |
| LOD113C-D2COMMON | 0x303A0 /10511; 0x348F0 /10392 | Current-skill and cached-target getters; valid unit, missing skill container or target returns absent |
| LOD113C-D2COMMON | 0x34BB0 /10867; 0x34B80 /10750; 0x390F0 /10913; 0x38B70 /10973 | Actual path X/Y and existing stat 6/resolved max 7 reads, consumed under RLI's existing-array/minimum contract |
| LOD113C-D2COMMON | 0x3A740 /10887; 0x38C20 /10587; prior 0x3A280/0x39F70/0x389E0/0x38550 | Own stat 328 replacement/read/increment and wrapped own-to-resolved delta under C07; kind 1 avoids player tracking, definition byte 5=0 avoids changed-key tracking |
| LOD113C-D2COMMON | 0x33CA0 /11150; 0x37EC0 /10930 | Extended state-storage bitmap access and state 0/flag 64 first-list absence |
| LOD113C-FOG | 0x141E0 /10126; 0x14120 /10128; 0x13D80 /10127; 0x13B70 /10252 | Actual bitstream initialization, low-first packing/byte count and empty-word sentinel; no supplied bitstream outputs |

Available original-byte-matched higher-level forms were read first, including Common coordinate/skill getters and Fog initialization/byte count. Remaining bodies were read as instructions; encoded regions on reached creation/mode/state/bitwriter paths were decoded from original files. Original Fog was freshly fingerprinted SHA-256 96c05770ab06b71a7a2d6abe5dc7c2f943b30ac8fe367e44bdbad7e2643bb863,212992 bytes, matching LOD113C-FOG. Its original bit-mask table at 0x26898 has masks 0,1,3,7,15,31,63,127,255. No analysis target or binary was changed.

Original mode 1 table entry at D2Game 0xFA730 has no target/extra-coordinate controls and maps through the original mode-dispatch byte table to the neutral path. Empty components, special appearance, owner data and flag 64 payload encode four zero bits after mode 1's four-bit field, giving byte 1. The empty active-state sender still emits its terminator 255 and command header. Full pending output is 14+8+10=32 bytes; a 480-byte pending tail therefore fills exactly 512. The two stat 328 writes are ordered between these sends: replace with unsigned low16(X+Y), then read own value and replace with+1 after the mode packet. Original C07 controls select linear delta propagation with no callback/dependent rewrite/keepzero path; ordinary extended-base fixtures preserve unrelated entries. No transport delivery or runtime gameplay is inferred. Wider nonempty appearance/state/skill/unit-message serialization, alternative modes and active-current-skill mode packets remain Q02; actual think/death remain Q03.

### T07 / C07: sender stat 328 correspondence

Reused complete effective ItemStatCost text 40109 bytes/SHA-256 ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375, with prior reported profile limits. Header-name selection gives Stat=`pierce_idx`, explicit ID 328, Signed 1 and Send Bits 6; empty callback/minimum/operation cells are not treated as compiled zero. A fresh successful complete 324-byte selected compiled record read on the owned original-profile process on2026-10-30 found count 359/stride 324/index 328/storedId 328, flag bytes 4/5 =2/0 and controls 0x50..0x53 all 0. This confirms only the selected scalar definition inputs. Loaded-DLL identities, override inventory and supplying archive remain uncaptured here; it is no sender execution observation.

### Arrival full-review corrections — 2026-10-30

The fresh smaller full arrival reader reported MISMATCH 0, GUESS 1 and GAP 1 (one shared undefined wire-field finding), DISCRIM 2 and ten coverage advisories. Accepted the field finding: R24 now defines every creation byte explicitly, including emitted length 14 at byte 12, and the analogous emitted length 8 at byte 6 of the empty-state command. Accepted both discrimination findings: V22 now asserts the ordered packet-append/own-write/resolved-delta/read log rather than relying on identical final arrays/bytes, and V15(c)'s independent pattern 0 grid begins nonzero and must remain unchanged. No broader serializer/death/gameplay gate was closed. These corrections await a fresh smaller follow-up.

| Coverage advisory | Outcome |
|---|---|
| R07 player dead mode 0 | Added an independent mode 0 repetition to V05(h), preserving the player-kind saved-key fixture and early stop. |
| R09 absent dynamic target path | Added V06(d): actual output (0,0), cache retained, no stored-target fallback. |
| R10 no destination room / absent nearby room | Added V08(b/c), distinguishing no downstream request from the sole nearby-operation request and no marker/collection. |
| R12 family refusal / wildcard key 0 | Added V10(b/c), checking unchanged queues on refusal and both type 5 removals plus new zero-key insertion on wildcard. |
| R13 room suppression | Added V12(b), with queued false so the suppression guard is the actual reason for unchanged fields. |
| R15 original point already clear | Added V14(e), exactly one candidate test and no perimeter. |
| R19 deferred removal / missing counterpart | Added V17(b/c): processing-time removal defers then unlinks once before next invocation; missing registry counterpart retains event and all path fields. The override fixture tests actual removal/dispatcher behavior at its explicit first-event boundary. |
| R21 fresh 520-byte allocation | Added V20(f), successful supplied resource allocation only, actual block clear/link/copy; allocator internals remain outside that fixture. |
| R22 ordinary AI identity outside range | Added V21(b), signed -1/148 return entry 0 at the isolated selection-helper boundary, without asserting that entry's uninspected initializer. |
| R23 absent unit | Added V23(e), byte 128 before stat lookup. |

HARDEST checks required no rule changes: R15/V14 stable ring tie/grid result, R16/R17/V15 exact collision/room-chain/population effects and R20/V19 reciprocal own/resolved deltas and unlink restoration retain their explicit discriminators. No source expansion was performed for these fixes; added cases exercise already stated rules.

### Second arrival full-review triage — 2026-10-30

The second fresh smaller full reader reported one possible R16 mismatch, one related guess/ambiguity and one discriminator concern: patterns 3/4 could be read as adding 8192 after all effects of patterns 1/2. The reader's numeric interpretation would yield 12416 instead of V15(c)'s8320. Both original consumers were rechecked before editing: Common 0x4BCF0 and 0x4BE90 independently dispatch patterns 3/4 to cross/square supplied-mask operations followed by center/cross 8192, with no 4096 operation. Each supplemental operation in patterns 1–4 is guarded by the original supplied mask being nonzero. R16 now describes every pattern independently and states that patterns 3/4 do not also apply 4096. This also makes pattern 2's zero-mask guard explicit. V15(c) now explicitly rejects 12416 and checks zero-mask suppression on a nonzero grid. Prior length-field and ordered packet/stat findings were otherwise reported clean.

| Coverage advisory | Triage outcome |
|---|---|
| R06 delegated nonzero | V05(b) now independently repeats with delegated result 1, complementing its existing 0. |
| R09 static coordinates/self cache | Added V06(e/f): self-cache fallback to stored point and exact object static coordinates. |
| R10 alternate marker/skip search | Added V08(d), no footprint/search request and alternate-only marker with exact supplied-path boundary. |
| R15 footprint 0/grid boundary | V14(f) checks missing west neighbor 39; V14(d) now explicitly repeats footprint 0 as alias of 1. |
| R16 destination 0,0/out-of-range pattern | Added V15(f), actual clear-only map branch then centered zero position/reset; V15(c) repeats pattern 6 on nonzero unchanged grid. Existing owned 0xDF30 source was rechecked for this fixture; no new callee/source scope. |
| R17 player population/fallback room | Added V15(g) requiring fallback B after old-room search fails, with actual chain/counter/condition effects. Player transfer remains outside this present-monster rule's input scope, so no new player fixture or scope expansion; V15(b) already checks the applicable monster population predicate. |
| R19 invalid type 5 route | Added V17(d), skill 0 and count-equal357 rejected before action lookup while persistent event/position remain unchanged. |
| R20 unchanged replacement/zero deletion | Added isolated writer V19(b/c) for no-op replacement and exact own-zero deletion/resolved delta, without inferring outer link/state operations. |
| R21 moved client/successful saved-key refresh | Added V20(g/h), moved-unit recipient skipped and nonabsent different lookup result used without replacing cache. |
| R22 failed operation-pointer validity test | Added V21(c), explicitly synthetic dependency result; fatal after installer stores, no rollback or claim that the captured original pointer is invalid. |
| R24 stat 328 zero deletion | Added V22(b), isolated complete recipient at coordinates 0,0: first own/resolved entry removed, absent own read 0, then reinsert 1 after packet append, exact wire bytes and no allocation with existing capacity. |

HARDEST V14/V15/V22 retain explicit ring-tie/grid, collision/independent-chain/population and ordered packet/stat discriminators. No broad question was closed. The updated delta awaits a fresh smaller follow-up; lint is 0errors/0warnings.

### Additional bounded observation — 2026-10-30

An owned original-profile ActV route reached BloodyFoothills 110 through ordinary play. Bounded entry records at D2Game RVAs 0xB46B0,0xDBD0 and 0xE19C0 were configured for at most 128 records/600 seconds each. The saved reads returned zero records and zero dropped for all three, with no in-flight invocation at the end of the window. No eligible teleport was captured, and no negative behavior inference follows. Q04 remains open and no executable vector is runtime-observed.

Separately, selected live code slices at 0xB46B0/176bytes,0xDBD0/370bytes and 0xE19C0/576bytes exactly match the manifest-hashed original D2Game after HIGHLOW relocation normalization (zero, one and three adjustments respectively). This authenticates only those selected code slices; it does not authenticate the whole loaded DLL/file path, data/archive/override identity or a gameplay invocation.

### Third arrival full-review outcome fixes — 2026-10-30

The third fresh smaller full reader reported MISMATCH 0/DISCRIM 0, GUESS 1 and GAP 2 concerning an insufficiently explicit event-pool fault outcome and missing Section 5 fault/assertion inventory. The collision-pattern finding was resolved. The applicable source was rechecked before edits: Game 0x8E420 returns absent if the resource request returns absent normally; Game 0x8E540 then reads the valid existing head's next-pool link and writes it through the absent new-pool result at 0x8E579, without a guard. Original-image decoding confirms that exact reached write. R12 now states the invalid-memory-access fault boundary and lack of an ordinary callback result, with later exception handling/process termination outside the contract. Section 5 explicitly inventories this outcome and the AI installer's fatal assertion after selector/thinker/parameter stores; it also inventories R18's earlier state 54 assertion. Game 0x9D230 was fully rechecked: the pointer-validity call follows selector/thinker stores and the initializer, while state 54 is tested before reset.

Self-review also corrected V18's reachability: a matching event removed before allocation could return a free node to a pool, avoiding the asserted growth. The fixture now has only nonmatching E2, a present existing pool head and no free nodes; the removal scan preserves E2 and therefore does not make a node available. No removed matching-event effect is asserted at the fault. Prior state/link effects retain their stated normal-return dependency boundary. The new precise rule outcome requires another fresh full review; this is not a stopping-rule completion.

| Coverage advisory | Triage |
|---|---|
| R10 path relocation 0 | Added V08(e): nearby+path requests only, no later marker/update/synchronization, with dependency path mutations unasserted. |
| R12 host family 441 | V10 now independently repeats its accepted attachment stage with 441 and exactly the same outputs; no new downstream integration inferred. |
| R14 negative health | Added V11(d), percentage -1 and the existing death-request branch, distinguishing a nonnegative clamp. |
| R15 missing grid/square split | Added V14(g/h): absent metadata gives each documented sentinel; right square remainder reads B's nonzero cell. Unchecked backing-storage failure is not substituted for absent metadata. |
| R17 absent destination room | Added V15(h): guard false installs absent room after actual old-chain removal, preserving coordinates/progression and setting synchronization, with no new-room insertion. |
| R18 state 54 assertion | Added V16(e): assert before resets/initializer, parameters/control/seeds retained and no initializer draw. |
| R20 preexisting ordinary state 98 list | Added V19(d): actual list reuse, changed own/resolved deltas and update marking, with no allocation/free. |
| R21 current-only moved-unit reference | V20(g) independently reverses the collections and checks no spawn/queue; explicitly synthetic cached-reference input, not a normal monster-as-client-player profile claim. |

HARDEST V14/V15/V19/V22 retain their exact grid/ring, room-chain, own/resolved and ordered wire/stat assertions without further behavior changes. No Q02/Q03/Q04 closure or runtime-observed vector was added.

## Current applicable full review complete — 2026-10-30

The fresh smaller full review after the explicit event-pool fault outcome reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0, with three documented gates retained. Its sole coverage advisory said that the absent-destination-room reference V15(h) was missing. The frozen spec was checked directly: V15(h) is present at line 453 in the unchanged reviewed spec hash, with point(30,5), no fallback and guard clear; it asserts current room absent/prior A, synchronization true, queued false and actual old-chain removal. No rule/vector addition or reference correction is needed. The reported coverage concern is therefore resolved by verifying the already-existing case, with no uncovered deficiency remaining from this pass.

HARDEST V15/V20/V22 required no changes: the contract retains exact collision/independent room-chain/counter results, sorted recipient/cache/output-block behavior and packet/stat ordering with exact wire bytes. Prior pending-review statements and earlier findings above are historical; this is the current review result. The applicable blind-implementer review stopping rule is complete for the current 24 rules/23 vector IDs and their explicit fixture boundaries. Q02 wider arrival/transport, Q03 remaining list/AI/death effects and Q04 matching gameplay remain open; no draft-status or runtime promotion follows. This final update changes review/evidence metadata only.

## Component-bearing arrival continuation — 2026-10-30

New source closes the independently reached nonzero-component arrival branch. The preceding 24-rule/23-vector full review remains complete for its recorded version; R25/V24 below are substantive additions whose fresh smaller full review is recorded below. Q02 remains open for special appearance/modifier and owner payloads, nonempty states/skill/unit-message sends, other modes and transport; Q03/Q04 are unaffected. No foundation document was edited.

| Artifact | RVA / ordinal | Exact supported assertion |
|---|---|---|
| LOD113C-D2GAME | 0x8CA20, positive component branch 0x8CB1A..0x8CB85; thunk 0xADA2 | Components-present bit then sixteen count/value writes; width 1 below count 3, otherwise bit length of count-1; actual completed packet length/queue |
| LOD113C-D2COMMON | 0x84C30 /10305 ->0x13B0 | Valid monster/index 0..15 reads u8 at linked extended-record 0x15+index; missing record/nonmonster/out-of-range gives 0 |
| LOD113C-D2COMMON | 0x56C10 descriptor initialization | Type 23 fields HDv/TRv/LGv/Rav/Lav/RHv/LHv/SHv/S1v..S8v bind callback 0x53C80 and component parameters 0..15 |
| LOD113C-D2COMMON | 0x53C80 callback, ending 0x53E1D | The callback initializes/stores count byte 0x15+component and a 12-byte component-code array; this supplies field association, not a completed generic text/compiler handoff |
| LOD113C-D2COMMON | thunk 0x9240, original slot 0x6FDD90BC | FOG ordinal 10213/RVA 0x11F00 component-code dictionary lookup; full unsigned-key binary search read, unrelated trailing cleanup not promoted |
| LOD113C-FOG | 10128/RVA 0x14120 and 10127/RVA 0x13D80 | Compatible actual low-first writer/count reused; no supplied bitstream output or new transport contract |

The new count getter and descriptor/callback instruction bodies were read in full. Original D2Common bytes independently corroborate the getter, descriptor bindings and callback count stores; the file was freshly fingerprinted to the existing SHA-256 59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10. Descriptor strings are at RVAs 0x91FC4 down through 0x91F88 in four-byte steps, exactly the case-sensitive column names above. The original positive serializer instructions and Fog writer were already decoded in the preceding compatible arrival record and were reused rather than treating a count or packing result as supplied. The callback compiler's dictionary result determines code bytes but not the caller's consumed count identity; complete text-column selection, missing-column/default and arbitrary compiled correspondence remain under GDA-R03d's existing preparation boundary. No full compiler, compiled original count vector or current data/profile identity was inferred from a field descriptor.

R25 consumes unsigned prepared counts and raw component bytes. Widths are at least 1 even for counts 0/1, because the count<3 branch precedes bit-length calculation. Every component is serialized once presence is true; zero components/counts are not skipped. No clamp to count-1 occurs before the low-bit writer. The maximum u8 count needs 8 bits, so these sixteen component fields fit the existing local bitstream buffer. The same actual stat 328 changes and state/neutral messages execute after the enlarged creation packet; no additional gameplay draw, initializer or transport operation is reached.

V24's asymmetric counts/values yield widths[2,2,3,3,4,4,1,1,1,2,2,3,3,4,4,1], total 40 component bits. Mode 4 bits + presence 1 bit + these 40 bits + three zero suffix bits are 48 bits, six bytes, giving creation length 19. Independent low-first fixture calculation gives payload `D1795C3A8F1B`, total output 37 bytes and pending used 400->437. Its explicitly synthetic first index 3/count 3 variant gives first byte `F1` instead of `D1`, discriminating clamping without claiming normal index generation produced that input. No gameplay vector is runtime-observed.


### Component full review complete — 2026-10-30

The fresh smaller full review reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0. Its four gate references are the existing Q02/Q03/Q04 and GDA-R03d preparation boundary, not four newly discovered questions. The coverage advisory requests a missing linked extended record/count-lookup fallback case. R25 already states that this lookup returns count 0, and V24 exercises count 0 with width 1 and an emitted component bit. Once that supplied-input count is obtained, the serializer is identical; this advisory does not expose an undefined outcome or a different packing branch. No additional vector or contract change is needed. It does not establish arbitrary prepared-record correspondence or the upstream preparation pipeline.

HARDEST V15/V20/V24 required no changes: collision and independent room chains/counters remain explicit; recipient merge/cache and pending queue boundaries remain explicit; component widths, low-bit packing, no-clamp discriminator and ordered packet/stat effects were independently recomputed. The applicable blind-implementer stopping rule is complete for 25 rules/24 vector IDs at these stated boundaries. Wider Q02, remaining Q03 and Q04 remain open and status remains draft. Earlier pending component-review statements are historical.


## Special-appearance and owner arrival continuation — 2026-10-30

The same arriving-monster chain now closes its special-appearance and owner branches under valid stored inputs, without changing the full teleport scope. New R26/V25 extend the preceding reviewed 25-rule/24-vector contract and require a fresh smaller full review. Q02 remains open for nonempty active-state/stat payload, skill/unit-message sends, other modes and transport; Q03/Q04 are unchanged.

| Artifact | RVA / ordinal | Supported decision or write |
|---|---|---|
| LOD113C-D2GAME | 0x8CA20, 0x8CB85..0x8CE0A | Special section selected by first modifier nonzero or unit condition; five booleans, conditional signed-word identity, bounded modifier sequence, appearance word, associated-player field, independent saved-owner field |
| LOD113C-D2GAME | 0xE03A0 | Compatible valid-kind game registry lookup reused; stored association game/kind/id, not current client-player reference |
| LOD113C-D2GAME | 0x96600 | Full first-match category traversal; category 1..PetType count-1, entries compared by arriving monster id; first match returned, absent collection gives 0 |
| LOD113C-D2COMMON | 10920/RVA 0x2EB70 via Game thunk 0xA4C0 | Byte-equivalent reconstruction read first: requires present kind-0 player and returns player-data reference; 0x96600 additionally asserts if that data is absent |
| LOD113C-D2COMMON | GDA-R04b existing count identity | DataTables+0xBF0 is PetType count; no new loader/count/data observation was performed |

Original D2Game and D2Common were freshly fingerprinted to manifest SHA-256 24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53 and 59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10 respectively. No byte-equivalent C view exists for Game 0x96600; its entire instruction body was read, then independently decoded from original image bytes. The original 0x8CB85..0x8CE0A and registry body were independently decoded; compatible Fog packing/pending-queue proof is reused. Complete Game 0x7B0A0,0xF35C0,0xDFA90,0xE07A0,0xB46B0 and 0xE19C0 were rechecked for the changed appearance/owner inputs: they do not introduce an extra reached effect in the stated class-492/mode-1/cache-empty physical fixture. The direct destination finds its room normally, so its saved-association fallback is not reached.

Semantic mapping: monster-data byte+0x16 bits 2,3,1,4,6 supply appearance booleans A,B,C,D,E in that order. Bit 1 additionally gates signed word+0x26. Unit+0xC4 bit 9 or the first modifier byte at monster-data+0x1C selects the section. Modifiers use at most nine bytes starting +0x1C, stop at zero, and always append an explicit zero terminator. Word+0x14 follows, with no random-generation call. The optional association context at monster-data+0x28 supplies saved game/kind/id at context+0x28/+0x30/+0x2C; registry output must be kind 0. Common10920 gives that player's data; Game96600 reads data+0x44 category collection, collection+0 category-array base, category stride 12, entry id+4/next+0x14. Categories scan from 1 while signed index<count and return first match. Only result 7 emits the player's id+0x0C. No category name or list producer/allocation mechanism is inferred.

Separately, unit+0xC8 bit 10 and saved owner kind+0x94 equal 0 gate the independent saved-id field at+0x98. The original AND0x8FFFFFFF followed by a 31-bit write has the exact semantic result low28 bits plus three zero bits; it does not emit the high id bit. It performs no owner registry lookup. These are two distinct player-id wire fields. Explicit wire payload ordering belongs to R26; the internal offsets above map the evidence and do not prescribe an implementation layout.

V25's full success uses asymmetric booleans, signed identity -2, modifiers 5/129/zero, word 48879, player id 305419896 and different saved owner id 4063516280. Actual low-first calculation gives 133bits, payload `61F5FF2F080478F78D67452311CF8A4600`, creation 30 + empty-state 8 + neutral 10 =48 queued bytes, used400->448. Subcases independently cover first-modifier selection with the condition clear, earlier category 3 match suppressing category 7, all nine nonzero modifiers plus explicit terminator, and saved owner kind 1 suppressing only its own field. Ordered append/stat effects remain actual R24 operations, not supplied outputs. All values/lists are synthetic fixtures; no additional original-profile scalar, list generation, normal AI/death execution or gameplay observation is asserted.


### Special-appearance full-review delta — 2026-10-30

The fresh smaller full reader reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 1. V22's packet/stat log did not distinguish a wrong eager associated-player search whose result was discarded when the special section was absent. V22 now supplies a valid potentially matching association and explicitly requires its special-section registry/category-search log to be empty; the assertion remains scoped to this creation section rather than all teleport lookups. R26 already specifies the skip. No rule or integration boundary changed.

| Coverage advisory | Triage |
|---|---|
| Section present with appearance boolean C clear | V25(f) omits the conditional identity/read, with exact 117-bit payload and creation length 28. |
| First modifier zero while section condition set | V25(g) retains the section but emits only the zero terminator; exact payload differs from (f) although both have length 28. |
| Associated player absent/non-player/category 0 | V25(h) independently tests registry absence, wrong kind and absent player collection; all omit that player-id field, with distinct exact lookup/access limits and V25(c)'s output bytes. |
| Owner condition clear with section present | V25(e) independently repeats with saved kind 0/condition clear, retaining the special and associated-player fields but omitting the independent saved-id field. |

HARDEST R26/V25 required no rule changes: independent ordered appearance bits and owner fields, first matching category and bounded modifier terminator remain explicit. R24/V22 retains ordered packet/stat events and now explicitly distinguishes the skipped search. R21/V20 recipient/cache/pending-block boundaries remain unchanged. The added cases exercise already-read branches; no new source scope, data claim, gameplay observation or gate closure is asserted. This vector-only delta awaits a fresh smaller focused review; prior full component/appearance results remain historical evidence for their reviewed versions.


### Current special-appearance review complete — 2026-10-30

The fresh smaller focused review of the vector-only delta reported MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0 and COVERAGE 0, with the three existing Q02/Q03/Q04 gates retained. Independent recomputation confirmed V22's skipped association search and creation 14 + empty-state 8 + neutral 10, V25(f/g)'s 117-bit payloads/creation 28/pending 446, and V25(h/e)'s creation 26/pending 444. No further contract or vector edits were recommended.

HARDEST R26/V25, R24/V22 and R21/V20 required no changes: independent appearance/owner bits and first-category selection, exact ordered packet/stat effects and the recipient/cache/pending-queue boundaries remain explicit. The applicable blind-implementer stopping rule is complete for the current 26 rules/25 vector IDs and their stated valid-input boundaries. Earlier pending-review statements above are historical. This review/evidence update changes no behavior, vector, gate or status. Wider nonempty arrival serialization/other modes/transport, remaining AI/death effects and matching gameplay remain open; no full parity or runtime confirmation follows.


## Nonempty active-state arrival continuation — 2026-10-31

R27/V26/V27 extend the previously reviewed 26-rule/25-vector contract with the actual nonempty active-state sender and complete source-unit-list arrival. The instruction readings span 2026-10-30/31; the additional interpretations below use the already complete 2026-10-30 compiled captures. The new behavioral delta requires a fresh smaller full review. Q02 remains open for the creation flag-64 payload, oversized active-state buffer continuation, skill/unit-message sends, other mode/outer sends and transport; Q03/Q04 remain unchanged. No shared stat/state foundation was edited.

| Artifact | RVA / ordinal | Exact supported assertion |
|---|---|---|
| LOD 113C-D2GAME | 0x7B0A0 ->0x7A940 | Reached monster arrival sends actual state command between first stat-328 replacement and neutral update; state 93 clear/class 492 retains the established suffix |
| LOD 113C-D2GAME | 0x7A940, full body | Snapshot scan; suppression/length guard; state/list/first-sixteen-own-entry selection; actual signed/unsigned clamp, layer/value packing and pending append |
| LOD 113C-D2COMMON | 11150/RVA 0x33CA0 via Game 0xAFAC | Active-word pointer from valid extended state storage; no membership mutation |
| LOD 113C-D2COMMON | 10871/RVA 0x38520 ->0x37F80 via Game 0xA5D4 | Compatible RLI-R03 state-list selection: first primary then secondary; state 0 absent |
| LOD 113C-D2COMMON | 10658/RVA 0x383C0 via Game 0xA4C6 | Byte-equivalent reconstruction read first: copy min(signed own count,16) entries from own array, preserving order, return copied count |
| LOD 113C-FOG | 10252/RVA 0x13B70 via Game 0xA2F8 | Lowest set bit, or -1 for zero; full byte lookup table at 0x27B50 checked against original bytes |
| LOD 113C-FOG | data exports 10259/RVA 0x26798 and 10261/RVA 0x26818 | All thirty-two bit/complement masks checked; Game slots 0x6FD18BE0/0x6FD18BAC import these data exports |
| LOD 113C-FOG | 10126/RVA 0x141E0,10128/0x14120,10127/0x13D80 | Compatible actual low-first writer and rounded used-byte count; full writer rechecked for capacity limits |
| LOD 113C-D2GAME | 0x8A3E0 | Compatible actual pending-block append under R21, not a supplied delivery result |

The sender and remaining helper bodies were read in full and independently decoded from the original images where needed; the matching own-entry-copy reconstruction was read before its machine-code view. Original Game/Fog mask tables independently establish increasing-bit scan and snapshot-only clearing. GDA-R12's existing descriptor/consumer mapping supplies stat `Signed` at definition+0x04 bit 1, `Send Bits` at+0x08 and `Send Param Bits` at+0x09; GDA-R13 supplies States `nosend` at record+0x10 bit 0. RLI-R03 supplies the already established primary/secondary list finder, not an invented list result. No new loader, generic lifecycle or stat propagation authority is asserted.

The state bitmap pointer is obtained through Common 0x33CA0 from the extended base's+0x58; its contents are copied before clearing selected bits. The stated stage bounds that snapshot to eight words; the captured count 185 uses six. Compatible state-write/force-changed consumers distinguish the first active half from the following changed half; the sender takes the first half and does not mutate either. The sender uses the stat list's own array/count at+0x24/+0x28, not resolved values. Each copied eight-byte entry has u16 layer, u16 stat identity and i32 value. Positive copied count selects the list-present bit before entry filtering; even an all-filtered list emits the nine-bit 511 terminator. No Send Other test is made here. The two value-clamp branches apply only below width 32. The current rounded payload-byte count plus seven is compared to 219 before each surviving state; failure stops that word, and later words reach the same guard. These are exact caller decisions, not a guessed packet-size policy.

The full Fog writer was rechecked rather than assuming that the pre-state guard proves arbitrary writes fit. On an overflowing field it sets the error flag and rounded count to the 244-byte capacity without writing that field; the caller does not consume the error. Its resulting unwritten scratch tail is not established as zero and remains outside R27's successful-fit output contract under Q02. The original full teleport boundary remains unchanged. R27 closes a complete independently reachable successful sender path, not all oversized outputs.

### Additional scalar interpretation of C04/C05/C07 — 2026-10-31

The existing complete compiled reads were reinterpreted at the now-resolved consumers; no new live read or invocation was performed. C04's successful 648-byte capture contains both complete 324-byte stat records: stored identities 353/354, flag byte+0x04 both 0, Send Bits bytes+0x08 equal 5/32 and Send Param Bits bytes+0x09 both 0. Thus their active-state values use unsigned widths 5/32; clear Send Other does not suppress this sender. C05's complete 60-byte state-98 capture has byte+0x10 bit 0 clear, establishing `nosend=false` independently of the earlier disguise-byte interpretation. C07's complete 324-byte stat-328 capture has flag+0x04 bit 1 set, Send Bits+0x08=6 and Send Param Bits+0x09=0, establishing signed six-bit transmission. Counts remain states 185/stat definitions 359. Existing record identities, original-profile dates and loaded-module/archive/override limitations carry forward; old heap addresses are not reused. These scalar correspondences do not observe a sender, change a behavior label to runtime-confirmed or assert the synthetic vector list values were produced in gameplay.

V26 uses the actual R20-shaped ordinary state-98 source-unit list, owning stat 353=1 and 354=9. Its state payload is 81 bits, hex `62C306B109000000FFFF01`; actual command length is 18, and creation 14/state 18/neutral 10 append 42 bytes, pending 400->442. Bitmap and list contents are unchanged, and the stat-328 writes retain their actual ordered location. V27 independently distinguishes own versus base/resolved values; signed upper/lower and unsigned clamps; width-32 bypass; optional layer bits; positive copied count with all entries filtered; bitmap order/suppression/out-of-count padding; first-sixteen truncation; and the actual rounded-byte stop. Its synthetic 100-state case emits states 1..29: 28*59=1652 bits gives 207 rounded bytes, while 29*59=1711 gives 214 and reaches the threshold; final terminator gives 1719 bits/215 payload bytes and total 222. All cases fit the writer and use actual queue writes. No gameplay vector is runtime-observed.


### Nonempty-state full-review triage — 2026-10-31

The fresh smaller full review reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0. Its 17 gate references are existing local/dependency limits, not seventeen new questions; the local Q02/Q03/Q04 boundaries remain. The five coverage groups were triaged against the already-read consumer and dependencies, without new source scope or rule changes:

| Coverage advisory | Outcome |
|---|---|
| State 0 always returns absent from list lookup | V27(j) now emits state 0 followed by a zero list marker despite a present primary state-0 list, distinguishing both skipped-state and incorrect list-find behavior. |
| In-range absent prepared state record | V27(k) now uses an explicitly synthetic absent-address resolution, not a claimed normal sparse table; state 32 emits nothing while state 98 remains the positive control. |
| Seven/eight bitmap words | V27(l) now independently uses synthetic counts 224/256 and a sole active bit in the seventh/eighth word, giving exact distinct state payloads without changing original-profile count evidence. |
| A filtered state before a later survivor at the length cutoff | No new case: filtering emits no bits, and used-byte count cannot decrease. In V27(g), if state 30 is suppressed, state 31 encounters the same 214+7>=219 guard; the remaining word stops and later words also fail. If the guard were incorrectly applied before suppression, the same packet bytes, length, queue contents and unchanged unit state result on valid record storage. There is no callback or other exposed effect between those record checks. The contract already states the exact filter-before-guard order; this advisory exposes no additional result branch within the stated boundary. |
| Value widths above 32 and layer widths above 16 | V27(m) now uses width 40/value -2 and parameter width 20/layer 65535, distinguishing arithmetic value extension from zero extension and zero-extended layer bits from signed extension. |

HARDEST checks required no rule changes: R27's bitmap/filter/list/packing interactions and V27(g)'s bit-count versus rounded-byte stop remain explicit; R26's modifier bound, first-category selection and two independent id fields remain unchanged. The new vector-only delta requires a fresh smaller focused review. The full-review result applies to its reviewed version, and no current stopping-rule completion, gate removal, status promotion or gameplay observation is claimed for the new cases.


### Current nonempty-state full review complete — 2026-10-31

The second fresh smaller full review, covering the added V27(j–m) cases as part of the full contract, reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0, with the three existing Q02/Q03/Q04 gates retained. All four added groups were independently recomputed and matched their exact bytes. Its two coverage advisories require no behavioral edit. With zero state count, the copied word count is zero and the scan emits no state/list field; the ordinary end-255 framing still produces the same exact eight-byte command as V22's all-zero bitmap. Neither path performs a state-list lookup or exposes a bitmap/stat/seed mutation. The rule already states the count-dependent copy/scan; this is an empty-scan input with no additional output result to discriminate. This equivalence does not require an invented read of a zero-length bitmap. The filtered-state-before-cutoff advisory was independently endorsed as having the same bits and next surviving state's same failing threshold; V27(e/g) also exercise suppression and the rounded-byte cutoff separately.

HARDEST checks require no changes: R27/V27(g) retains rounded-byte cutoff, per-word stop and end framing; R27/V27(m) retains arithmetic extension to forty bits and zero extension of the u16 layer to twenty bits; R26/V25 retains conditional packing, two independent ids and first-category selection. Prior pending-review statements are historical. The applicable blind-implementer stopping rule is complete for the current 27 rules/27 vector IDs at their explicit boundaries. This update changes only review/evidence metadata; oversized/other recipient outputs, remaining AI/death/lifecycle effects and matching gameplay stay open, with no status promotion or runtime confirmation.


### Shared relationship / saved-owner input correction — 2026-10-31

A subsequent AI-consumer reading exposed an input-identity omission in the preceding reviewed model: the saved-owner wire condition/key and the relationship lookup's condition/key share storage. Fresh full reads of Game 0xE0A20 and Common 0x32CA0/0x33020, checked after looking for byte-equivalent C coverage, confirm the alias. Original-image decoding independently corroborates Game's reads of unit+0xC8 bit 10 and kind/id+0x94/+0x98, Common's writes of those same fields and OR/AND of 0x400, and the already-recorded 0x8CA20 wire consumer. The original Game/Common files again matched the manifest hashes. No shared lifecycle effect is promoted by this field-identity check.

R08, R20 and R26 now explicitly consume one relationship-present condition and one saved counterpart kind/id. R26's nested monster-data association context remains distinct from these unit fields; its two emitted player-id fields can still differ. Its saved-owner field is independently selected after the special section, but its underlying unit flag/key are shared with relationship lookup. The existing V25 bytes remain correct: its explicit input overrides the previously clear relationship condition and sets the player counterpart key. V26 also overrides that condition to set it, with kind 1/id 9; the saved-owner field remains absent through R26's kind guard, and the complete source-unit-state arrival bytes remain unchanged. R27's composition states that guard rather than inheriting a contradictory clear-condition precondition.

V25(i) adds an isolated relationship lookup on the same post-creation unit. It must resolve the full saved u32 key 4063516280, while the wire field has only its low28 bits 36984440; the serializer neither clears the shared condition nor replaces the stored key. This distinguishes a duplicate independent input and a masked-key overwrite without claiming that the selector-98 invocation calls the lookup. No packed byte, previous numeric result, gate or status changes. The current source-grounded input/composition correction and its probe require a fresh smaller review; preceding clean reviews apply to their recorded versions. Further AI/death behavior has not been promoted.


### Current shared-input alias full review complete — 2026-10-31

The fresh smaller full review of the source-grounded relationship/saved-owner correction reported MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0 and COVERAGE 0, with existing Q02/Q03/Q04 retained. It explicitly checked consistency of the input map, R08/R20/R26, V25(i), V26 and R27 composition, including that the full-key probe does not imply a selector-98 lookup call. No changes were recommended.

HARDEST checks required no changes: R27/V27(g)'s bitmap/capacity interaction, R26/V25's appearance/category and association-versus-saved-counterpart identity, and R21/V20's merge/cache/block-allocation contracts remain explicit. Earlier pending-review statements are historical. The applicable blind-implementer stopping rule is complete for the current 27 rules/27 vector IDs at their stated boundaries. This update changes only review metadata; no gate, status, numeric vector or runtime claim changes. Subsequent AI/death and remaining recipient/lifecycle work continues separately from this reviewed version.


## Actual attachment-thinker fallback continuation — 2026-10-31

R28/V28 close a subsequent attachment-thinker branch through actual normal-AI restoration, current-event deferral and next-frame node writes. They extend the preceding reviewed 27-rule/27-vector contract and require a fresh smaller full review. Q03 retains other thinker/mode/event/allocator/death branches and future execution; Q02/Q04 are unchanged. No shared lifecycle/record document was edited or promoted.

| Artifact | RVA | Supported reached operation |
|---|---|---|
| LOD 113C-D2GAME | 0x166D0, selected original path 0x166D0..0x16740/0x16939..0x1696D | Inline shared counterpart lookup before state-143 test; absent state/counterpart selects normal control 0, then continuation delay 1 |
| LOD 113C-D2GAME | 0x9D230 ->0x9D0F0 ->normal table entry 122/init 0x15530 | Compatible R22 reset/normal-122 initializer/install with old special 16 alternate absent; full installer body rechecked |
| LOD 113C-D2GAME | 0xF22C0, full body | Mode 1 bypasses Common thunk 0xA496 and Game 0x75E10; save each successor, remove type 2 only, then request type 2/now+1/no override/zero parameters |
| LOD 113C-D2GAME | 0x8E6F0 processing prefix | Flag word+2 mask 0x0001 causes OR 0x0008 and return, preserving links/current cursor; compatible removal body fully reread |
| LOD 113C-D2GAME | 0x8E540 ->0x8E680 | Existing pool's+0x7080 free node is popped via node+0x1C, cleared as twelve dwords, assigned target/kind/id and prepended to unit+0xDC chain |
| LOD 113C-D2GAME | 0x8EA80, full body | Compatible STL-R01 frame/type fields; actual node/calendar slot writes read through return; empty slot becomes head/tail singleton |
| LOD 113C-D2GAME | 0xFA3DC, index 1 | Original unit-kind mapping gives 1 for monster kind 1; source combines that group with 64 frame slots |

All listed new helper/thinker instruction bodies were read in full after checking for byte-equivalent C coverage; none has such a view. Compatible R22 initializer and RLI-R03 active-state/registry contracts retain their dates/limits. Full original-image listings of the continuation, scheduler, pool-pop and unit-link functions, plus the selected thinker entry/fallback and processing-removal prefix, independently corroborate the reached instructions. The original Game file was freshly fingerprinted to manifest SHA-256 24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53. The selected slot writes do not claim arbitrary calendar insertion/dispatch ordering under STL-R02a.

The think context is read on entry (control pointer, target and distance fields); the missing-state/counterpart branch exits before difficulty-dependent monster-template/skill/RNG operations. The normal reset consumes R22 with class 492/AI 122/old 16 and state 54 absent. The mode 1 guard is checked before constructing a mode-change request, so no 0x75E10 result is supplied or inferred. The continuation loop saves unit-next before calling removal. For the currently-processing type 2 event, removal tests processing mask 0x0001, sets pending-removal mask 0x0008 and returns before unlink/recycle; this case reuses the same proven processing/deferred mechanism as R19, with the actual type 2 caller read here.

The scheduler's type 2/state 54 special request is skipped because that state is absent. The node constructor uses a genuinely available pool head, not a supplied completed event. It pops the free node, clears all 48 bytes, writes target+8/kind+0x10/id+0x0C and repairs unit-next/previous+0x24/+0x28. The scheduler writes type+0, frame+4, parameters+0x14/+0x18 and override+0x2C; its empty-slot head/tail stores are through the selected calendar arrays+4/+0x504. These original offsets map the evidence only. No pool creation, heap call, callback or future event dispatch is established by this successful preallocated path.

V28 sets frame 100, currently-processing type 2 E0 in slot 36, nonmatching type 8 F in slot 8, free nodes E1/E2 and empty next slot 37. At thinker return the actual control is normal 122/[-1,0,0], E0 is still processing but deferred, E1 is the new unit head and singleton slot 37 node at frame 101, and the pool free head is E2. This boundary deliberately precedes outer cleanup of E0. State 143-on/missing-owner and stale-key/source-unit-list subcases retain the old state/link/stat data rather than inventing cleanup. No gameplay invocation was observed and no full normal-think/death closure follows.


### Attachment-thinker fallback full review and coverage triage — 2026-10-31

The fresh smaller FULL review of R28/V28 reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0, with GATED 3 (Q02/Q03/Q04). All V01–V28 outputs were recomputed, including the E1/E0/F unit links, E2 free head and calendar slot 37. It found two advisory COVERAGE entries. Invalid extended-link/record resolution has the same downstream footprint-0 collision result already exercised by V14(f), but the producing guard was not explicit in that fixture; V14(i) now checks negative/out-of-range links and a synthetic absent prepared address against a blocked east neighbor. An empty unit-event chain changes the actual unit-link outputs and eliminates the pending-removal write, so V28(d) checks a singleton E1 chain and empty removal/write logs. Neither addition changes a rule, input contract, execution boundary or gate; both require fresh reader checks before the resulting vector set's stopping rule is complete.

All three HARDEST entries need no contract change: V15's collision/room synchronization remains physically specified; V27(g)'s bit packing and rounded capacity threshold remain exact; V28's normal reset, processing-event deferral, available free-node pop and empty-calendar writes explicitly stop before outer cleanup. No RNG, heap growth or later thinker execution is claimed. The preceding clean full review applies to its recorded vector set; the two added cases are pending a fresh focused delta review. Q02/Q03/Q04 and draft status remain unchanged.


### Attachment-thinker coverage delta review complete — 2026-10-31

The fresh smaller focused DELTA review returned MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0 for the changed material. Both preceding coverage findings are resolved: V14(i)'s signed links -1/609 against count 609 and synthetic absent selected record yield footprint 0/collision 0 despite the east cell's 1, whereas footprint 2 would return 1; no pattern write is invoked. V28(d)'s empty unit chain/calendar/current cursor cause no removal visits or pending-removal writes, followed by the actual E1 singleton unit chain and slot-37 head/tail, free head E2, type 2/frame 101/zero parameters. The boundary remains thinker return before outer cleanup. No further edits were recommended.

R15/R28, dependency contracts, acceptance and Q02/Q03/Q04 remain unchanged. The full review's three HARDEST entries retain their recorded no-change outcomes: V15 collision/room effects, V27(g) packing/capacity and V28 deferred-event/free-node/calendar relations are explicit. Earlier pending-review statements are historical. The applicable blind-implementer stopping rule is complete for the current 28 rules/28 vector IDs. This completion records static review only; no gameplay observation or full AI/death/recipient closure follows.


## Present-counterpart attachment idle continuation — 2026-10-31

R29/V29 extend the reviewed 28-rule contract at the installed attachment thinker's owner-present branch. The original dispatch/installer/table identity remains R18; R28's missing-state/owner branch and its completed FULL/DELTA review remain intact. The new behavior requires a fresh smaller FULL review. Q02/Q03/Q04 and draft status remain: neither full subsequent AI nor death is closed.

| Original artifact | RVA / export | Supported reached operation |
|---|---|---|
| LOD 113C-D2GAME | 0x166D0, selected continuation 0x1674A..0x168E7 | Resolve fixed prepared templates 492/493, signed context distance guard 24, owner checks, second-parameter guard, difficulty-indexed distance/chance decisions and actual delay-20 continuation |
| LOD 113C-D2GAME | 0xDFA90 | Same kind/mode/alive-exclusion predicate as R07: monster modes 0/12 or player 0/17 reject; selected monster mode 1 passes |
| LOD 113C-D2COMMON | 10830 / 0x391F0 via Game 0xA514 | Compatible R17 modifier predicate: present extended base, first state-105 list through 0x38520/RLI-R03; absent list returns 0 before own stat-172 getter |
| LOD 113C-D2GAME | 0x1180 | NRF-R01/NRF-R02 caster Rand with modulus 100; unchanged primitive consumed, caller draw placement established here |
| LOD 113C-D2GAME | 0xF22C0 ->0x8E6F0/0x8E540/0x8E680/0x8EA80 | Reuse R28's fully read processing-event deferral and available-node/empty-slot stores with actual delay 20; no mode request when current mode is 1 |

The full 0x166D0 body was already read for R28 and reread here. No byte-equivalent C covers the listed thinker/owner/modifier bodies; complete disassembly and original-image decoded listings independently corroborate their selected instructions. The full 0xDFA90 and 0x391F0 bodies were read; the latter's compatible R17 source and RLI-R03 lookup contract supply the absent state-105 result without duplicating generic stat machinery. The original Game/Common files freshly match manifest hashes 24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53 and 59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10. The unselected positive-list/stat branch is not promoted by the absent-list vector.

The shared counterpart key resolves before state 143. With both present, DATATBLS+0xA78/+0xA80 resolves fixed record indices 492/493 (record stride 0x1A8, offsets 0x32EE0/0x33088); valid prepared records are explicit preconditions, and neither fixture threshold is an asserted original-profile scalar. Think-context+0x14 supplies signed distance, and control+0x18 is the second parameter. With distance at most 24, eligible counterpart and modifier zero, its nonzero second parameter takes the secondary-template decisions. The signed distance word is secondary+0x5C+2*difficulty; distance at least that word skips the first roll and chance read. Otherwise the caster draws before reading signed secondary+0x62+2*difficulty, and equality with that chance proceeds to the final decision. The final caster draw is compared to fixed 50; a result at least 50 reaches 0x168D7 ->0xF22C0 with delay 20. Lower final results enter mode 8 through 0xF2220; successful initial chance enters skill/target operations. Those other mode/skill effects remain Q03.

V29's distance 24/threshold 24 makes exactly one step from {22820111,5149} to {694638760,9518094}, returning 60; V29(b)'s distance 23 with chance 11 makes steps returning 11 then 60, with the chance read after the first step. Both physically create type 2/frame 120 in slot 56 from the available E1, defer processing E0 and leave E2 free. Attachment selector 16/parameters [0,2,0] are retained: no 0x9D230 reset/install is requested. The same source confirms no target-coordinate getter, skill lookup, mode change, state/stat mutation or direct client send on these reached paths. Future dispatch, nonprocessing removals, pool growth and general owner/stat cases retain their gates. No runtime invocation was observed.

V29(c) additionally checks final-decision equality: seed {0,50} steps to {50,0}, return 50, and still creates the actual delay-20 event with attachment control retained. This distinguishes a strict-greater final test. The main case keeps its unread difficulty-0 chance 101, so substituting that prepared value for fixed 50 takes a different branch. Difficulty-1 distance 90 would add a draw if incorrectly selected. These prepared values remain synthetic consumer inputs.


### Present-counterpart full review triage — 2026-11-01

The fresh smaller FULL review reported MISMATCH 0, GAP 0, GUESS 1, DISCRIM 1, seven advisory COVERAGE groups and three HARDEST entries. Its twelve GATED references describe existing Q02/Q03/Q04, stage-vector limits and shared STL-Q10; they are not twelve new local questions. The following outcomes apply individually.

- GUESS R29/V29: the reference to R07's kind/mode/exclusion eligibility could be read as importing its separate relationship-routing rejection. Original Game 0xDFA90 reads only kind/mode/exclusion; R29 now states these tests explicitly and excludes a lookup of H's key. V29 gives H a present relationship resolving J and asserts the sole registry lookup (G,1,9), so the broader routing test fails this valid input. No new owner effect is inferred.
- DISCRIM V17: action 147 always returns 1, including the absent-counterpart case. A synthetic action-147 return 0 would violate its actual source domain. Testing a return to retain that particular action is observationally equivalent within the action-147 branch, but the persistent dispatcher also admits supplied overrides. V17(e) now supplies a valid override returning 0 without removal, and asserts exact queue retention. Original Game 0x8EBC0 calls either route, clears processing and tests removal mask 8 without consulting the return; the new case rejects return-based outer retention without inventing an impossible action result. This is the local persistent dispatcher, distinct from STL-V16's ordinary fired-event removal boundary; no generic foundation change is claimed.
- COVERAGE R15 other invalid signed footprints: no change. The default branch returns the same 65535 for every footprint outside 0..3, without coordinates/map reads or mutation. V14(d)'s footprint 4 already selects that exact branch; a negative signed value selects no additional result branch.
- COVERAGE R19 return 0: V17(e) adds the valid override result-0/no-removal case. No action-147 result-0 fixture is asserted.
- COVERAGE R29 difficulties 1/2: V29(d) adds both selected-record-index cases. Difficulty 1 selects distance 90 and makes two draws (60,98); difficulty 2 selects signed distance -7 and makes one draw (60). Exact seeds and read logs distinguish incorrect index selection.
- COVERAGE R29 first draw below chance: no change; this enters the explicitly gated skill/target branch outside R29's verified idle predicate. No successful skill result is supplied.
- COVERAGE R29 final draw below 50: no change; this enters the explicitly gated mode-8 operation, whose output is not established by the delay-20 contract.
- COVERAGE R29 nonzero modifier or zero second parameter: no change; each leaves the stated R29 input domain for the explicitly gated skill/mode branch. The absent state-105 predicate and nonzero second parameter remain explicit preconditions.
- COVERAGE R29 negative signed distance/chance thresholds: V29(d)'s signed distance -7 and signed chance -9 now discriminate unsigned interpretation, with exact one/two-step outputs.

HARDEST V27(g) packing/capacity, V25 conditional wire fields and V29 conditional draws/frame 120/slot 56 need no contract changes: their bit counts, wire IDs and ordered decisions/stored event relations are already explicit. The full review's pending findings are resolved in the revised text, but the eligibility clarification and added vectors require a fresh smaller FULL review of this resulting 29-rule/29-vector set. The preceding R28 stopping rule remains historical; Q02/Q03/Q04 and draft status are unchanged.


### Present-counterpart full review complete — 2026-11-01

The fresh smaller FULL2 review returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0 and unresolved COVERAGE 0. Its twelve GATED references are the documented Q02/Q03/Q04 and stage/shared limits, including STL-Q10; no new local gap is implied. Every V01–V29 output was recomputed, including difficulty-1 rolls 60/98 ending at {552006998,289728526}, difficulty-2 signed threshold -7 skipping the first draw and returning 60, final equality 50, and V17(e)'s zero-return override retaining E while clearing processing. No changes were recommended.

All three HARDEST entries retain no-change outcomes: V27(g)'s packing/rounded capacity, V25's wire fields/length, and V29's signed thresholds, conditional seed steps and actual twenty-frame event writes are explicit. Earlier pending-review statements are historical. The applicable blind-implementer stopping rule is complete for the current 29 rules/29 vector IDs at their stated boundaries. This is static review only; draft status and Q02/Q03/Q04 remain. A later shared-lifecycle consumer integration changes only after its own source/contract comparison and applicable review.


## Reviewed ordinary lifecycle consumer integration — 2026-11-01

R20 now consumes the exact reviewed direct ordinary operations STL-R05a/STL-R06d/STL-R06e/STL-R06f. Shared source/evidence dates remain 2026-10-31, and local R20 caller/compiled-control records remain 2026-10-30; this integration does not renew their verification dates. The shared frozen spec fingerprint is da6c77d94f94ad3b388b2775a056fd79b7b4037b42683ef4129eafa90843f101. No shared document was edited. The consumer refactoring and explicit stage fixtures require a fresh smaller FULL review; the preceding clean R29 review remains historical.

| Consumer operation | Exact shared contract and preconditions | Local caller/evidence retained |
|---|---|---|
| Existing state-98 own 353 then 354 replacement | STL-R06e/R06d; ordinary primary source, extended nonsecondary/unfiltered ancestors, direct zero controls, prepared capacity/no resize | Common 0x32CA0 ->0x3A280 ->0x39F70; original C04 flags/control evidence and R20 selection/order remain local |
| New empty ordinary list full attachment | STL-R06f with parentless/callback-free STL-R05a; valid extended target base, empty own entries, no cycle or mode-bound condition | Common 0x32CA0 ->0x3A7D0, full-application request 1 after state 98 assignment; allocator entry/return is separate |
| Selected ordinary list unlink/subtraction | STL-R05a/R06d; ordinary primary/full applied, no callback, same direct controls/ancestry/prepared slots | Common 0x33020 finds first state 98 then calls 0x3A3B0 before its storage-release branch; key/state/update prefix remains local |

C04's original 2026-10-30 full compiled reads establish flags bytes +4/+5 zero and controls +0x50..+0x53 zero for stored definitions 353/354, count 359. They match the direct-definition inputs: no dependent formula/expansion, retain-zero, filtering, resolved notification, dependent marker or changed-key control. This preserves the existing capture identity/profile and loaded-module/archive limits; no new data read or gameplay invocation is claimed.

Compatible local caller bodies were reread after checking byte-equivalent C coverage. Common 0x33020/0x32CA0/0x3A3B0/0x3A7D0 have no such view; the compatible 0x3A9F0 view is byte-equivalent C. Their original-image corroboration retains its existing recorded date. The specific unlink boundary is important: 0x33020 calls 0x3A3B0 first, then tests the list's extended condition and may call 0x3A9F0; the latter again requests unlink before its own-array/list release calls. V19(e) stops on return from the first unlink, before that release path. Thus its readable own entries and still-set relationship condition are actual intermediate outputs, not a guessed post-free snapshot.

V19(d) now makes capacity 8, parentless nonsecondary/unfiltered base, ordinary full-applied source and absent callback explicit. Its exact existing outputs 1/9 own and 4/29 resolved require no allocation or resize. V19(e) begins at that linked state and executes the real cleanup prefix/unlink to resolved 3/20 and detached readable L98; it stops before release/final flag clear. V19(f) separately begins at the actual empty-list attach call with a valid returned list, capacity 0/absent own backing, so empty attachment makes no storage request and yields [L,P] without any contribution or state write. Neither new stage claims allocator, free, notification, extended-list or callback-body success.

The earlier complete V19 allocation/release fixture retains its stated normal-success boundary and recorded limits; it is not evidence that the shared handoff closes those internals. Other supplied list kinds, storage-resize/release behavior, indirect definitions, disguise clearing, subsequent AI/death and gameplay remain Q03/Q04 or the named shared gates. The local policy continues to choose identities 353/354, state 98, first-list selection, reciprocal counterpart values, membership/update and write/unlink order. The generic direct array/list mechanics are consumed from their owner rather than established as a second generic authority.


### Ordinary lifecycle consumer full review complete — 2026-11-01

The fresh smaller FULL review of R20's shared ordinary integration returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0 and uncovered-deficiency COVERAGE 0. Its eight GATED references describe existing Q02/Q03/Q04 and STL-R05/STL-R06b/STL-R06c/STL-Q07/STL-R07a limits, without adding local questions. No changes were recommended.

All three HARDEST entries need no edit: V19(d)'s actual replacing writes match the shared direct definition/ancestry/prepared-storage conditions; V19(e)'s unlink/subtraction ends with a detached readable list before release; V19(f)'s empty attach establishes parent/head/sibling/association stores before the write frontier. The reader confirmed those limits and the older full V19's separate allocation/release boundary. Earlier pending-review statements are historical. The applicable stopping rule is complete for the current 29-rule/29-vector contract and this consumer integration. No allocator/free/general lifecycle or full Q03 promotion follows; draft status, Q02/Q03/Q04 and the named shared gates remain.


## Far-counterpart skill-request prefix — 2026-11-01

R30/V30 extend the existing installed-thinker path through actual jitter, current-skill/condition/path writes and the precise mode-operation entry. They require a fresh smaller FULL review. The preceding R20/R29 stopping rule applies to its recorded version. Q03 still owns Game 0x75E10 execution, its return handling, cast/skill success, later AI parameter writes and death. Q02/Q04 and draft status remain unchanged; no generic mode/stat/allocator authority is added.

| Original artifact | RVA / export | Supported reached operation |
|---|---|---|
| LOD 113C-D2GAME | 0x166D0, 0x16785..0x16790 ->0x168EA..0x16904 | Present relationship/state 143, signed context distance above 24 branches before H eligibility; first-slot mode/skill and primary difficulty radius forwarded to 0xF2420 |
| LOD 113C-D2GAME | 0xF2420, full body | Current C X then two inlined caster draws interleaved with Y, signed radius subtraction, unit+0x80 associated game and absent-target handoff to 0xF2380 |
| LOD 113C-D2COMMON | 10867 /0x34BB0 via Game 0xA454; 10750 /0x34B80 via 0xA3B8 | Compatible NRF-R10b current integer dynamic coordinates; byte-equivalent reconstructed getters retained |
| LOD 113C-D2GAME | 0xF2380 ->0x75070 ->0x74F20 ->0x1320 | Valid mode 8 clears current selection, builds request and selects 100/101 from linked mode-8 condition |
| LOD 113C-D2COMMON | 10708 /0x4F080 via Game 0xA3D0; 11103 /0x303E0 via 0xA496 | Byte-equivalent views establish first record-id/entry-owner -1 match after sentinel, then current-skill store at unit skill-container+0x10 |
| LOD 113C-D2COMMON | 11005 /0x346A0 via Game 0xAD18 | Byte-equivalent view: nonnull path with supplied u8 control 1 stores 0 at path+0x93 (valid control stores control-minus-one) |
| LOD 113C-D2GAME | 0xF2380 prefix through call 0xF23EB ->0x75E10 entry | Unit+0xC4 mask 0x40 set; request actor/mode/target/X/Y/selector and final control 0 are passed, with no consumed result |

All newly owned prefix callees were checked for byte-equivalent C coverage first. Common 0x303E0/0x346A0/0x4F080 have such views and were read in full; Game 0xF2380/0xF2420/0x75070/0x74F20/0x1320 do not, and their complete bodies were read as instructions. Complete original-image decoded listings independently corroborate those five Game bodies. Compatible coordinate getters, installed-thinker identity and NRF-R01/NRF-R02 seed arithmetic retain their earlier dates/boundaries. The original Game SHA remains manifest 24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53. No shared program/source/consumer was changed.

For mode 8, original table 0xFBC50+12*8 contains {0,0,1}; 0x74F20 therefore uses the linked extended record from 0x1320 and tests byte+0x105 bit 0 through the existing imported Fog bit-mask table (Game slot 0xF8BE0, Fog export 10259/RVA 0x26798; compatible R27 verification). A zero/missing linked result returns zero. In the valid-record fixture the stored condition alone selects request byte+0x15=100 or 101. Its wider producer/meaning is not asserted. The request's logical options begin zero/absent; internal padding is not a behavioral contract. The matched current-skill setter writes the supplied entry/absence only when the valid skill container is present. The new path byte+0x93 is distinct from R16's auxiliary dwords+0x72/+0x76 and progression counters+0x24/+0x28; the unit condition mask 0x40 is distinct from C8's teleport/relationship conditions.

### C08 — dated compiled first-slot/radius correspondence

On 2026-11-01, the complete successful 424-byte C06 capture from 2026-10-30 was reinterpreted at additional code-consumed fields, with no new live read: stored row 492, signed first skill word+0x170=330, signed first-slot mode byte+0x180=8, signed difficulty radius words+0x5C/+0x5E/+0x60=14 each. This is compiled-record scalar correspondence only; existing process, launch-profile, archive and loaded-module limits remain C06's limits. Raw captures remain private. GDA-R08a supplies the `Skill1`/`aip2` field bindings. Selected original Common loader 0x5AA40 descriptors name `Sk1mode` (string RVA 0x927C4), descriptor kind 23/slot 0/callback 0x54130; that callback's slot-indexed stores identify mode byte+0x180. Its complete body was read, but full mode-token conversion/other producer behavior is not newly promoted. Code consumer 0xF2380/0x75070 independently establishes that the consumed byte is mode 8. The fixed template index 492 belongs to the thinker code; skill/mode/radius values are prepared bindings. No new named text-cell value or gameplay execution is claimed.

V30 uses current point (100,100) and radius 14: the two low-word modulo-28 results are 27 and 4, yielding (113,90), final seed {694638760,9518094}. The ordered current-skill stores are absence then the first skill-330/owner -1 entry; the skill-request condition is set and path byte changes 17 to 0 before the exact mode-8 request. Radius-zero/condition-set and absent-skill variants preserve their explicit consumer boundaries. The output boundary is entry to 0x75E10, not normal return from it: no fallback scheduling, actual mode/cast or caller's later parameter=-1 store is supplied. Further complete mode execution remains source work; no gameplay invocation was observed.


R30 self-review retains only consumer-visible semantic mode fields: mode/actor/target/X/Y, selector 100/101, numeric mode parameter 0, auxiliary mode control 0 and separate finalization control 0. Original request+0x18 supplies the numeric parameter, byte+0x1C the auxiliary control; zeroed internal padding is not asserted. V30 additionally makes selected entry A level 0 versus later B level 9, rejecting a level test the actual lookup never performs. V30(d)'s synthetic distinct associated game checks the source of the exact handoff independently of the invocation context; no natural two-game association is inferred. Original Common selected descriptor instructions 0x5D017..0x5D03D and full 0x54130 image listing corroborate the Sk1mode-to-byte association, without promoting token-conversion callees. The supplied mode-selection condition's producer/use stays Q03.


### Far-counterpart skill prefix full review complete — 2026-11-01

The fresh smaller FULL review returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0 and COVERAGE 0. Its one specifically listed stage gate is R30/V30's Q03 mode execution, return/cast, later AI and death boundary; existing Q02/Q03/Q04 elsewhere remain. No contract fix was needed. The dispatch brief had conflated R30's fixed primary record 492 with R29's secondary template 493. The source-grounded contract already correctly uses 492 for first skill 330/mode 8/radius 14 and 493 for R29's distance/chance thresholds; the brief was wrong and no data binding was changed to fit it.

HARDEST V30's conditional RNG/first matching skill/request, V27(g)'s packing/capacity and V25's appearance/association/independent IDs need no changes: exact reads/draws/stores/arguments and byte framing remain explicit. Earlier pending-review statements are historical. The applicable stopping rule is complete for the current 30 rules/30 vector IDs. This records static review only, with no mode/cast success, gameplay observation or document/gap promotion. Subsequent source work follows the actual mode-operation call chain while preserving the reviewed boundary until another complete effect is established.


## Actual selected imp start-validator continuation — 2026-11-01

MTP-R31/V31 closes the selected start-validator invocation reached by the mode-8 initializer, not the preceding mode operation. Original `LOD113C-D2GAME` RVA 0x75510 reaches 0xA0830. The existing SCD-R02 through SCD-R05a order is reused with its original evidence limits; this continuation resolves the selected producers instead of supplying their results. No gameplay invocation or complete mode success was observed. Current public scope/gates remain unchanged.

### C09 — compiled skill-330 selected fields

On 2026-11-01, one successful full 572-byte compiled record read on an owned original-profile LoD 1.13c process had stored identity 330, table count 357, stride 0x23C and stable pointer/count headers before and after the read. This corroborates srvstfunc 0 at +0x2C, srvdofunc 129 at +0x2E, lineofsight 0 at +0x18F and bonus group 0 at +0x1DC. The selected flags give InTown clear and periodic clear; raw bytes +4..+8 are [0,4,10,128,0]. Named column correspondence reuses T01's original dated complete effective Skills text evidence; compiled storage is not a fresh complete archive read. The same record's UseAttackRate flag is set, relevant to still-open wider mode integration, not a claim that an animation ran. Header stability and baseline profile do not establish whole loaded-module identity, complete loose-file inventory or physical archive origin. Raw records remain private.

### C10 — loaded effective-level bound scalar

On 2026-11-01, the same owned paused original-profile process supplied two successful four-byte reads: the pointer stored at `LOD113C-D2COMMON` RVA 0xA0B50 and the pointed signed scalar 99. Original Common export #10306/RVA 0x505C0 dereferences this scalar as the effective-level upper clamp. The record establishes that one consumed scalar, not its producer/default across profiles or execution of the level reader. No game function was invoked. Full raw read results and artifact integrity remain separately preserved.

### Selected code/effect map

| Semantic operation | Original 1.13c anchor and read method | Established selected branch |
|---|---|---|
| validator | D2Game 0xA0830, compatible SCD reading plus selected original-image decisions | Current node, eligibility, bonus-inclusive level, charge override, identity, target/party, town, start dispatch and return order |
| identity / charge extractor | D2Common #10601/0x4E1D0 and #10537/0x4E240, original-image-checked reconstruction read first | Identity 330; marker -1 returns 0 without override or write |
| absent-target eligibility | D2Game 0x9FC00 ->0xE07A0; D2Common #10392/0x348F0 | Reuses R09's actual absent-cache refresh, no registry lookup; returns accepted 1 before other target predicates |
| level | D2Common #10306/0x505C0 ->0x4FCB0 ->0x4EBA0 ->0x37DB0; original-image disassembly, complete instruction readings | Missing resolved stat 127/0, state 134 inactive, additional bonus 0, monster stat 97/330, group 0 skip, stat 107/330; zero bonus and base 0/1 preserved by floor/cap 99 |
| resolved missing stat | D2Common #10973/0x38B70 through RLI-R02a | Missing entry returns 0 before any minimum adjustment; no new stat-definition minimum is supplied |
| room / actual town | D2Common #10331/0x2FE10 reused getter; #10057/0x3C390 ->0x2D320, complete instructions and original-image table | Room +0x10 ->level link +0x58 ->level identity +0x1D0; immutable selector table accepts exactly 1,40,75,103,109; 110 fails unsigned upper guard |
| activation | D2Game 0x9EC80, complete instructions | Monster/nonplayer returns 1 before any further call |
| start | D2Game 0x9FD00, complete instructions and original-image disassembly | Three total validator/dispatch room reads, two town tests; lineofsight 0 skips resolution; original start table RVA 0x107338 slot 0 is null; no callback/mana; periodic clear skips event-chain access; result 1 |

Original-image-checked reconstructed functions were read before disassembly where available. The town selector's entire 109-byte original table and original entry guards were checked; this is code-fixed town classification, not a guessed Levels column. Physical cache getter calls include each refresh's initial cache test as well as the two direct caller queries; the semantic vector records the two refresh operations and two direct queries without requiring an internal API layout. Charge-marker -1, missing bonus entries and absent target deliberately select complete actual branches, not supplied booleans. A zero base level is accepted by this start slice; later SCD-R06 action-level acceptance is distinct.

### Boundaries and self-review

V31(a/b/c) tests actual accepted/no-write level-1 and level-0 returns and the distinct town-109 current-skill-clear/return-0 branch. No current mode, flag, animation, path, state/stat, event/client store or seed is changed by the accepted validator. This does not erase prior mode effects: the fixture begins at validator entry, and earlier mode-8 damage/animation work is still Q03. Subsequent action-129/actual teleport, callback/event execution, later AI/death and matching gameplay observation remain Q03/Q04. Shared non-direct recomputation for damage stats 19/21/22 is not consumed as a completed direct-control contract. This new rule awaits a fresh smaller FULL reader; prior 30-rule stopping-rule statements remain historical for this 31-rule revision.


### Actual selected start-validator fresh FULL review complete — 2026-11-01

The fresh smaller FULL reader returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 3, COVERAGE 7 advisory groups and HARDEST 3. The three gates are existing Q02, Q03 and Q04, not new defects. All executable vectors, including V31's selected start-validator path, were recomputed. An interim suspicion about target eligibility preceding the level read was withdrawn after rechecking SCD-R02's explicit order; it is not an outstanding finding. The selected-map stat-getter ordinal was corrected to #10973 at the unchanged RVA 0x38B70 against the owning RLI record; no contract behavior changed.

| Advisory group | Outcome |
|---|---|
| R04 absent initial room / wrapped rectangle upper bounds | No change. Absent initial room takes the already tested unmatched-room/fallback outcome of V07; it adds no new return category. Wrapped i32 bounds consume NRF's arithmetic and retain the same half-open match/no-match decisions; no new room-result default is supplied. These exact edge inputs were not newly observed. |
| R12 rejected counterpart family / non-monster family getter | No change. They reach the same empty-effect refusal already tested by V10(b)'s family guard; neither can issue later state/link/AI requests. This is not a claim every guard operand has a separate vector. |
| R14 multiplication wrap / division fault | No change. Each i32 raw health shifted by eight lies in [-8388608,8388607]; multiplication by 100 stays in [-838860800,838860700], so cannot wrap. A nonzero shifted denominator cannot cause signed division overflow with that numerator range; the zero-denominator return is V11(c). No impossible wrap/fault fixture is invented. |
| R15 remaining cross-query boundary codes / unchecked square backing | No change. Existing blocked/clear and grid-boundary fixtures test the declared search outcomes. Unchecked backing outside valid storage has no invented safe result and is not a newly closed execution path. |
| R17 population underflow / increment wrap / player counter | No change. V15(b) already tests actual monster decrement/increment versus V15's no-counter branch. Increment is NRF i32 wrapping, with no different control branch. The player path is outside R17's stated present-monster stage. The fatal nonpositive-old-counter diagnostic remains an unexercised advisory; this pass does not claim exhaustive diagnostic coverage or a normal result for it. |
| R19 other type-5 refusal guards, including absent records | No change. V17(d) tests the rejected route's no-action/retained-event outcome and V17(e) separately tests return-independent retention. Other named refusal operands reach that same outcome; normal GDA storage guarantees in-range records, so absent in-range records are synthetic malformed cases, not a new ordinary result. |
| R27 skipped nosend state | No change: already explicitly exercised by V27(e), state 9 suppressed between retained states 7 and 32. The original bitmap remains unchanged and exact output order/bytes exclude 9. |

HARDEST R15/V14–V15 (footprints, perimeter order, grids and room writes), R26/V25 (independent conditional low-first fields and lengths), and R31/V31 (eligibility, level/charge, town, activation and null dispatch) all receive no change: their existing order, numeric/byte assertions and stage boundaries are explicit and the reader recomputed them without a difference. No guesses, gaps or discrimination defects remain to resolve. The applicable stopping rule is complete for the current 31 rules/31 vector IDs. Prior pending-reader statements above are historical. This is static review only; document status stays draft and Q02/Q03/Q04 remain. Wider actual mode/attack-animation and non-direct damage-stat integration continues separately without treating this validator return as a completed cast.


### Targeted old-population assertion check — 2026-11-01

The R31 full review's R17 advisory received one targeted source check rather than relying on the earlier no-change rationale. The stated valid-chain/storage input domain does not exclude counter 0. Complete D2Common RVA 0x66700 instructions show resident unlink and clearing the removed unit's next-resident field before the population predicate calls RVA 0x3B650. The latter has an original-image-checked reconstruction, read first; its signed count test is strictly positive, and the zero/negative path reports the fatal assertion before the decrement store. Original manifest-hashed Common decoded listings for 0xD760/0x66700/0x3B650 were read through their complete bodies. No game was faulted or called.

The exact source order is prior-room save at 0xD7E0, call removal at 0xD7E6, unlink predecessor at 0x667A0 and clear the removed next field at 0x667A6, then modifier/own-list predicate and checked count call at 0x667C6. Counter check at 0x3B65B branches past the assertion only when positive; counter write at 0x3B67F is therefore not reached for 0. Update-chain removal at 0x667CC, queued-condition clear at 0x667DF, current-room clear at 0x667E8, and selected-room/synchronization stores at 0xD7F3/0xD7F6 occur only afterward and are not reached. The state-105/stat-172 predicate reuses R17's compatible RLI-backed reads; no new generic stat effect is asserted.

Added only V15(i), an isolated synthetic inconsistent count with otherwise valid chains, and made this existing fatal branch's write order explicit in R17 and section 5. Prior no-change advisory is historical, not the final disposition. It asserts the three retained writes, unchanged count/update/current-room fields and absence of all later insertion/synchronization work; no impossible normal return or rollback is supplied. Other R31 full-review advisory dispositions and Q02/Q03/Q04 remain unchanged. Because the rule's order is now explicit rather than vector-only, a fresh smaller FULL review is required for the resulting 31-rule contract; the clean R31 review remains the previous revision's checkpoint.


### Targeted population-assertion fresh FULL review complete — 2026-11-01

The fresh smaller FULL reader of the resulting contract returned MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0, with three existing gates Q02/Q03/Q04, ten advisory coverage groups and three hardest items. V15(i)'s exact retained prior-room/resident-unlink prefix and absence of all later updates were executable without a missing decision. The initial R31 no-change disposition for this old-counter assertion is superseded by the targeted source check, explicit R17 order and V15(i). No other rule, gate or vector was changed in this follow-up.

| Advisory group | Outcome |
|---|---|
| R04 failed alternate-context lookup | No change. This is the documented unmatched destination/absent request operand, not another callback return category; supplied relocation still decides the local result. Existing match/fallback and zero/nonzero-result fixtures retain their stated narrower coverage; this exact missing-registry combination is unexercised. |
| R12 ineligible counterpart family | No change. It is the same empty-effect family-refusal outcome as V10(b); the earlier review records already distinguish family rejection from successful relocation and attachment requests. No separate test of every equivalent family guard operand is added. |
| R13 no current room | No change. It has the same no-write/no-draw outcome as the suppression/already-queued refusals; no new collection or return category is introduced. |
| R14 multiplication overflow / negative shifted maximum | No change. Overflow is impossible for the shifted i32 range proved above. A negative nonzero denominator follows the same NRF signed division without another branch; V11 already tests negative percentage, zero denominator, death and restoration results. No additional raw-health variant is added. |
| R15 missing grid/cells and other boundary codes | No change. Existing rejection/clear/grid-boundary fixtures exercise the declared search result categories; unchecked malformed backing receives no safe default. This pass does not claim exhaustive cell-layout coverage. |
| R17 increment wrap / absent old room | No change. Increment consumes NRF i32 wrapping, with no new branch. Absent old room omits old unlink/decrement and retains the existing selected-room install/insertion operation; successful insertion and failed destination are already covered. The genuinely different fatal-decrement outcome is now separately V15(i). |
| R19 invalid state/action selection after valid skill | No change. These explicit guards share V17(d)'s no-action/retained-event result; V17(e) separately tests that a zero override return does not remove the event. No every-guard-operand coverage claim is made. |
| R21 remove append to a non-full tail | No change: V20(d) already specifies an existing output tail used 0, the six exact remove bytes and final used 6, without allocation. |
| R27 every scan/filter combination | No change. Advisory combinations add no independently specified result branch beyond the explicit valid/invalid/nosend/filter/capacity cases. V27(e/g) preserves the relevant state order, bitmap nonmutation and cutoff behavior; no exhaustive Cartesian product is claimed. |
| R30 negative radius | No change. A negative radius takes the same nonpositive-modulus no-step branch exercised by V30(b)'s zero radius, with ordinary NRF i32 coordinate arithmetic. The rule explicitly retains both calls and signed subtraction; no original-profile negative cell is asserted. |

HARDEST R17/V15(i)'s fatal write prefix, R27/V27(g)'s rounded-byte capacity/state order, and R29/V29's signed thresholds/conditional draws/actual continuation scheduling each receive no change: the reader found their explicit operations and assertions coherent and no mismatch. All actionable lists are empty; the current 31-rule/31-vector-ID applicable stopping rule is complete. Historical pending-review notes above remain historical. This records static review, not a gameplay match, complete teleport integration or status/approval promotion.


## Actual current-skill arriving-monster branch — 2026-11-01

R32/V32 extends the same current-only recipient path: R21 ->D2Game 0x7B0A0 kind-1 entry ->creation/state sends ->0xF35C0 ->0xF34C0. No compatible completed rule previously covered the present-current-skill mode sender; R24's neutral path remains intact. Available reconstruction coverage was checked first: neither Game 0xF35C0 nor 0xF34C0 has a matching entry. Whole instruction bodies were read; the complete 0x100-byte original 0xF34C0 body was decoded and read to corroborate its encoded instructions. The original manifest-hashed Game SHA remains 24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53. Existing original-image-checked 0x7B0A0/creation/bitwriter/pending-queue records retain their dates and limits; changed mode/current-node decisions were rechecked.

At 0x8CAC9..0x8CB05 the original creation writer explicitly preserves mode 8 as four bits; the remainder of the stated all-empty appearance sections follows the already closed R24 path. This makes its single payload byte 8 without altering its 14-byte length. The selected full original-image prefix was read; an initial capture ended one byte short of its final CALL encoding and was extended through 0x8CB05 before using that call. No new unresolved fragment supports a rule.

Original 0xF35C0 refreshes/cache-reads at 0xF3642/0xF364B, before current-skill lookup at 0xF368D. Present current skill branches to 0xF39F5, calls 0xF34C0 at 0xF39FD, then returns; it never reaches the neutral packet's 0xF38A6..0xF38BD own-stat-328 read/increment/store sequence. The new sender gets current node, identity through Common #10601/RVA 0x4E1D0 and level through #10306/0x505C0. R31's fully read selected level recipe supplies the actual 0/1 result, not a supplied level operation; its validator/activation/town/charge stages are not invoked. Its absent target argument bypasses all target-room/client-membership inspection.

At 0xF353E it calls Game thunk 0xAAD8 ->Common #10764/RVA 0x34AA0, the stored unsigned target-Y getter. It then calls 0xAAC6 ->Common #10983/RVA 0x34AB0 for stored target X. These compatible matched getter readings retain their original limits; both imports were freshly resolved against original exports. Sender writes at 0xF354D..0xF3582 establish all 17 output bytes, including zero-extension of low16 skill identity into a four-byte field, low8 level and final zero word. Queue call 0xF3589 ->0x8A3E0 consumes R21's actual pending append with sufficient tail capacity. No allocation, current-skill change, spell execution or transport occurs on this selected path. The present-target/unit-target command at 0xF3597 remains outside R32/Q02 rather than assigned the point-command format.

V32 begins at an isolated synchronization call with V22's resulting physical room/chain state already present, mode 8/current skill 330/base 1 or 0, and a deliberately different stored point (113,90) from current point (12,5). It does not rerun selector 98 using the changed stored point. This makes the consumer boundary explicit without claiming its physical or mode producer. Creation/state/point-command lengths are 14/8/17, queue 400->439. Only the first stat 328 replacement123->17/delta-106 occurs; no neutral increment follows. Exact packet bytes and effect/read order discriminate a wrong current-coordinate source, neutral fallthrough, missing level lookup or executing the skill instead of reporting it. These are static vectors, not gameplay/transport observations. This closes one selected arriving branch while Q02's remaining sends/targets/transport, Q03's full mode/stat/AI/death effects and Q04 remain. New R32 behavior requires a fresh smaller FULL review; preceding 31-rule reviews remain historical.


### Current-skill arrival fresh FULL review complete — 2026-11-01

The fresh smaller FULL review of the resulting 32-rule/32-vector-ID contract returned MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0, with existing gates Q02/Q03/Q04, six advisory coverage entries and three hardest items. No contract fix was needed after that pass. The sender/creation instruction addresses in the new evidence map were cross-checked against original listings and corrected without changing any behavioral element. Pending-reader notes above are historical; no observation or document promotion follows.

| Advisory | Outcome |
|---|---|
| R03 valid positive aurastate reaches lookup | No change: V04 supplies selector 129/record 330/aurastate 143 and reaches the relationship/target/coordinate sequence; V02 separately isolates the accepted zero selector. Positive selectors are not absent sentinels. |
| R04 no match and no alternate fallback | No change. The absent destination operand is explicitly defined; room/fallback and zero/nonzero local result categories are already tested. This exact input combination remains advisory, not claimed as an additional gameplay case. |
| R13 no current room | No change. This gives the same no-write/no-draw result as the tested suppression/already-queued guards, with no new effect or return category. |
| R15 negative footprint | No change. Like the invalid footprint 4 in V14(d), a negative signed footprint uses the same out-of-0..3 default query result 65535. There is no distinct negative-footprint shape or safe-success default to add. |
| R16 containing guard after collision writes | No change. The no-rollback ordering and guard refusal are explicit; V15(e) exercises the underlying no-containing-room guard result while R16's isolated collision cases test mask effects. The combined diagnostic/guard combination remains unexercised advisory; this review adds no claim that all physical failure combinations were observed or tested. |
| R17 selected-room bounds assertion | No change. With stable consistent rectangles, R17's actual search selects a new room only because it contains the current point. A separate failure fixture would require inconsistent data or a change between the search and assertion; no normal search result is replaced by an invented out-of-bounds room. The assertion remains explicit, and V15(i) separately tests the independently reachable nonpositive-counter fatal prefix. |

HARDEST V15's collision/position/independent chains/synchronization, V25's low-first conditional appearance/owner packing and V27's bitmap/list/rounded-byte guard each receive no change: the existing logical operations, exact bytes and write order remain explicit and the reader found no mismatch. All actionable lists are empty and the applicable stopping rule is complete for the current 32-rule contract, including R32's selected complete arrival branch and existing stage boundaries. Q02's remaining sends/targets/transport, Q03's full mode/stat/AI/death integration and Q04 gameplay matching remain open. No callback success or mode production is supplied to turn the new reporting packet into a cast assertion.


## Creation-stat payload continuation — 2026-11-02

R33/V33 consumes the remaining positive creation-list branch at original D2Game 0x8CE0A..0x8CF73 inside 0x8CA20, reached by the existing R21/R24 arriving-monster root. Prior complete creation/bitwriter/queue readings retain their dates; the selected positive tail and original instructions were rechecked on 2026-11-02. No matching reconstruction covers 0x8CA20. D2Common #10658/RVA 0x383C0's original-image-checked reconstruction was read before disassembly; #10930/0x37EC0 has no matching reconstruction and its complete instruction body was read. Fresh original decoded bodies corroborate both imported helpers and the complete selected tail. Original Game/Common/Fog fingerprints match the manifest identities already recorded above; no game process was invoked.

Game thunk 0xAFBE resolves Common #10930/0x37EC0 with unit, state 0 and mask 0x40. The base must exist and its flags' sign bit must be set; the mask does not select the secondary head. The finder visits primary links and accepts the first equal state whose flags intersect 0x40, independently of active membership. This is not RLI-R03's ordinary state-0 lookup. Its local condition is called creation-stat-list condition here, without assigning wider gameplay meaning. Game 0xA4C6 resolves Common #10658/0x383C0: signed own count at +0x28, own array +0x24, first min(count,16) contiguous entries copied before caller filtering. Valid nonnegative counts/readable arrays are the stated domain; no malformed-pointer or negative-count fallback is inferred.

The original Game import at RVA 0xF8BE0 resolves Fog data export #10259/RVA 0x26798. Its first original DWORD is 1, establishing the +0x04 definition-bit test as GDA-R12's Send Other, rather than a guessed mask. Caller 0x8CE50 loads unsigned stat/layer fields and raw value; ids at least table count, missing definitions, Send Other clear and Send Bits (+8) zero skip. Send Param Bits is +9. Signed and ValShift are not conversion controls in this branch. Existing R27/Fog low-first packing and R21 pending append are reused, not a second generic serialization authority.

The original branches distinguish no list (0x8CE1E ->0x8CF36, one zero write at 0x8CF3F) from a found empty/all-filtered list (0x8CF28/0x8CF1D ->zero write 0x8CF31, then fallthrough zero write 0x8CF3F). Accepted entries emit the first presence bit at 0x8CEC2, ids/layers/raw values at 0x8CECF/0x8CEEA/0x8CEFA and terminator 511 at 0x8CF14..0x8CF3F, with no extra zero. Byte count at 0x8CF49 adds 13, stores its low byte in the header and passes the full count to pending append at 0x8CF62. The local payload capacity is 244 bytes; oversized continuation is not promoted.

V33 uses supplied synthetic definition controls/own arrays, preserving first matching list and first sixteen raw entries rather than first sixteen accepted entries. Its 49-bit positive payload is [129,10,218,22,252,255,1] and complete creation length 20; absent produces length 14 while present-empty/all-filtered/first-sixteen-filtered produces length 15. The isolated complete serializer returns before the caller's stat/state/mode sends, so no linked-list application, stat-328 or cast success is supplied. Q02's remaining writer/sender/transport cases, Q03 and Q04 remain. New behavior awaits the next fresh smaller FULL review; preceding 32-rule stopping-rule records are historical.


## Actual selected mode-action event — 2026-11-02

R34/V34 follows the existing event root rather than inferring a skill trigger from its label. Compatible STL-R02 owns exact-frame dispatch and later removal: Game 0x8EBC0 invokes 0x769A0 for a nonoverride monster event. New selected readings establish the kind-0 binding, standard class-492/mode-8 callback and its actual action/animation effects, ending before outer event cleanup. Matching reconstruction availability was checked before instruction reads; none covers Game 0x769A0/0x752E0/0x75CE0/0xA2060 or Common 0x4F070/0x4E480. Whole bodies and complete original decoded selected listings were read. Compatible original action-dispatch, teleport, level, town, mana and nonsequence advancement readings retain their dates and bounds. Original Game/Common fingerprints match the existing manifest identities.

Original Game table RVA 0xFBC14 entries 0/1 contain RVAs 0x752E0/0x752A0. Router 0x769A0 kind 0 takes its local state-1 gate, through Game 0xA68E ->Common #10494/0x33CD0 and RLI-R03. With state clear it selects 0x752E0. With that state active and the same mode-8/alive-exclusion-clear actor, 0xDFA90 returns 0 and no callback occurs. The outer dispatcher and typed router make no animation-counter adjustment before this call. Standard mode table RVA 0xFA518 supplies 0x75510/0x75CE0/0x76BF0/1 for mode 8; class 492 selects this standard row even when its record's special-table condition is set, because it misses the class-specific cases in 0x749A0.

0x752E0 calls the row's middle callback. At 0x75CE0 the present node's flags through Common #10397/0x4DE40 have condition A clear, reaching 0xA2060 before Common #10853/0x32460. It does not store the event argument into the sequence-event field or test the unit latch on this selected branch. 0xA2060 obtains current node through #10511/0x303A0, identity through #10601/0x4E1D0 and inclusive level through #10306/0x505C0, invokes 0xA19A0 with consume 1/forced 0/itemTarget 0, ignores its return and invokes 0x9F3B0. R31 supplies only the selected actual level/town recipes, not another validator invocation. Existing R03/R04/R06/R10/R15–R17 close actual action-129 ->selector-98 physical success for the clear-relationship/cache-empty fixture.

For the zero-level case, the actual owned lookup is a different function from R18's first-id/owner search: Game 0xA3F4 ->Common #10630/0x4F070 ->0x4E480. The 16-byte export wrapper has no matching reconstruction and its encoded instructions were independently decoded against the original image; the complete 0x50-byte finder was read. On the stated one-entry chain it finds A when id matches and node +0x3C is zero; its level remains zero and the dispatcher rejects. Wider best-entry preferences are not promoted by this fixture. All actual found-node inputs are explicit; no successful lookup is supplied.

Game 0x9F850's complete original body returns 1 for a nonplayer before any current-node/stat/resource call, making the reached monster mana request inert. Quantity clear skips 0xA1790. The consuming-delay front end is Game 0xAC9A ->Common #10786/0x51BF0; its original-image-checked reconstruction was read first. SCX-R11 rejects the unsigned empty reference before bytecode. 0x9F3B0 periodic/aura-clear branch returns before calculation or event access. These actual selected operations close their local integration, not the helpers' player/charged or periodic branches.

### C11 — prepared animation record scalars

A 2026-11-01 owned original-profile read captured the complete loaded AnimData hash bucket 1 (57 records/9120 bytes) with stable pointer/count headers before and after. A separate Storm read of effective `data/global/animdata.d2` established total size 570304, preceding bucket-0 count 54 and bucket-1 range [8644,17768), including its count and all 57 records; every record byte matched the loaded bucket. Named key IPS1HTH is record 24 at file offset 12488. Selected scalars are frame count 16, stored base rate 256 and frame-7 action mark 1. These are individual binary-record fields; raw records/maps stay private. There is no whole-file hash, physical-archive identity, actual lookup/animation invocation or full loaded-module identity. Record selection and final rate production remain their existing gates. V34's complete action map is a supplied synthetic fixture, not a published game-data row.

### C12 — additional code-consumed skill-330 scalars

The complete successful 572-byte compiled Skills330 capture from 2026-11-01 (C09) was reinterpreted on 2026-11-02 without a new memory read: signed word +0x46 is -1 and DWORD +0x190 is 0xFFFFFFFF. Game 0xA19A0 consumes these as the missile selector and consuming-delay calculation reference. The same captured flag bytes establish decquant/aura/periodic/usemanaondo clear under GDA-R11a's flag mapping. The source/profile/header-stability and archive/module limits remain C09's original limits. The name/schema association for additional missile/calculation columns retains SCD-Q13's outside-boundary limit; the consumer's prepared fields, values and operations do not depend on inventing a loader correspondence.

### Nonsequence advancement and boundary

Common #10853/0x32460 has no matching reconstruction; the whole instruction body and complete original 352-byte decoded listing were read, including its encoded nonsequence spans. Counter is unit +0x44, signed step +0x4C, total +0x48, action code +0x4E and prepared map at animation record +0x10. Wrap requests Common #10031/0x32220; its compatible full reading calls 0x501E0, which with the input-remap condition clear (unit +0xC8 bit 3) preserves effective kind 1 and returns start frame 0 before inventory selection. The related original Game scheduler 0xDF7B0 writes game frame shifted eight to this same counter; it is not silently replaced by an assumed elapsed animation value in V34(b).

V34 starts with explicit prepared animation data and a reached event, not a supplied completed mode. Main counter 1536 becomes 1792/code 1; scheduler-produced value 25600 becomes 1280 after six wraps, retaining mark 1 from five full sweeps; zero becomes 256/code 0. Zero action level still yields callback 1/advancement, whereas active state 1 suppresses the callback. Exact V15 physical success is integrated only for level 1 and no target/link branch. Outer event-processing/free effects, earlier 0x75E10 damage/rate/animation production, other node/AI/death branches and gameplay matching remain Q03/Q04. No event-release or generic animation owner is duplicated. New R33/R34 behavior awaits one fresh smaller FULL review of the combined resulting contract.

### 2026-11-02 creation-stat/action-event full review, first pass

The fresh full pass over all 34 rules and vectors returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 3 (Q02–Q04), COVERAGE 5 advisory groups and HARDEST 3 groups. It independently recomputed the selected-list 49-bit payload and the positive, level-zero, suppression and multi-wrap action-event cases.

The blocking discrimination finding was resolved by adding V05(a)'s explicit X, then Y, then room getter log before relocation. This asserts the existing R07 order without changing the rule. Coverage triage: R04 wrapped upper-bound and absent-initial-room cases retain the existing rectangle/room-selection contract; no new successful or refusal result class beyond the existing room cases is claimed. R09's distinct dynamic target with absent path uses the same zero-valued dynamic getter contract already exercised at the getter boundary; no new result branch or ordering is introduced. R10/R16 alternate-marker and post-write guard refusal retain the independently stated marker and refusal-prefix boundaries; no atomic rollback is inferred. R33 now has V33(e/f) for in-range absent definitions and absent state storage, with exact existing empty/no-list outputs. R34 now has V34(f) for marks 2–4 and last-recognized replacement across full and final sweeps. All five advisory groups have an explicit outcome; only the new meaningful assertions were added.

HARDEST triage: no rule change for R15/V14–V15 because ordered perimeter/collision inputs distinguish ring ties and map masks; no change for R16/R17/V15 because independent chains and writes-before-fatal prefixes are explicit; no change for R33/V33 and R34/V34 algorithm boundaries because copy-before-filter and action-before-animation are explicit and the recomputed values agree. The added mark case further discriminates scan ordering. All gates and the before-outer-event-cleanup boundary remain unchanged. A fresh applicable review is pending for the added assertions; the first full pass is historical, not completion for this revision.

### 2026-11-02 creation-stat/action-event full review, second pass

The second fresh full review returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 2, GATED 3 (Q02–Q04), COVERAGE 3 advisory groups and HARDEST 3. It recomputed all vectors, including physical result 819200/360448 and 112/144 projections, the 9801-point maximum search count, the six-wrap animation result 1280 and the previous RNG vectors.

Both blocking ordering findings were resolved with assertions of existing rules: V03(a) now requires collection to finish while the marker remains false, followed by the marker write; V15 now explicitly logs stored X, stored Y, then current room before search. Self-review separately found that V34(f)'s first recognized mark equaled its final mark, weakening its claimed first-versus-last discriminator. Adding frame 1=2 leaves the expected last mark 3 unchanged but makes first-mark retention yield 2; every full sweep now sees 2,3,2,4 and the final sweep sees 2,3. This is an input/discrimination correction, with no rule or boundary change.

Coverage triage: no change for R17's defensive new-room assertion branches: selected-room bounds are established by the same unchanged point's containing-room search, while duplicate residency contradicts the fixture's valid unique membership; the fatal-result class itself is exercised by V15(i), without claiming that it covers these different fatal prefixes. No change for R19's additional selector/record refusals: V17(d) already exercises refusal before the action while preserving the persistent event, and the additional guards reach the same no-action retention result rather than a new result class; their exact guard order remains stated. No change for R23's negative maximum and h>=m saturation: the contract specifies the same byte 128 already exercised by saturation cases, without introducing a different write, draw or downstream branch. These remain advisory coverage limits, not newly verified gameplay cases.

HARDEST outcomes all no change: perimeter tie order/count is explicit in R15/V14; the V15 collision/room-chain prefix is explicit and now also asserts direct getter order; R34's nested action and scans have exact effects and before-cleanup boundary, with the stronger mark discriminator just described. A fresh full pass is required for these vector changes. No question or acceptance boundary was removed.

### 2026-11-02 creation-stat/action-event final full review

The third fresh full pass returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 3 (Q02–Q04), COVERAGE 1 advisory and HARDEST 3. It found no vector-output difference or missing input/output/decision, and accepted the strengthened getter, collection/marker and last-recognized-mark assertions. The current 34-rule contract satisfies the applicable blind-review stopping rule; earlier pending-review statements record historical revisions.

The sole advisory concerned R14 multiplication overflow or division fault. No change: for signed i32 raw health, `sar(H,8)` lies in [-8388608,8388607], so multiplying by 100 lies in [-838860800,838860700], wholly inside signed i32. The numerator cannot equal -2147483648, so the signed-overflow division case is also unreachable; zero denominator returns 0 before division as V11(c) already demonstrates. The numeric dependency's general fault contract remains intact, without inventing an impossible fixture.

HARDEST outcomes are no change: R27's bit counts, rounded-byte stopping condition and framing are explicit in V27; R17's independent chains and fatal decrement prefix are explicit in V15, including the underflow case; R34/V34 supplies the actual selected action integration, distinct zero-action versus suppressed-callback results, and exact sequential multi-sweep mark replacement before outer cleanup. All three remaining questions retain their previous full scope. No gameplay observation or complete mode/AI/death result is claimed by this review.
