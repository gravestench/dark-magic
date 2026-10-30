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
