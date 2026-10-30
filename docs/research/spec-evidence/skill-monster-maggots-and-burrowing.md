# Evidence: monster maggots and burrowing

Contract: [monster maggots and burrowing](../../specs/skill-monster-maggots-and-burrowing.md). Remaining work: [evidence gaps](../../plans/gaps/skill-monster-maggots-and-burrowing.md).

## Identity and reuse

Original Windows x86 Expansion 1.13c artifact `LOD113C-D2GAME` is identified by the [source manifest](../source-manifest.md). The inspected image was SHA-256 checked on 2026-10-29 and matched `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`. RVAs below are module-relative; analyst descriptions are not exported names.

The shared [dispatch evidence](skill-cast-start-and-do-dispatch.md) retains its 2026-10-12/15 method and limitations. It supplies the established dispatcher/table roots, not the interiors of these monster handlers. This new reading examined their selected original table slots and the local paths below. It does not reverify the shared AI/frame entry or generic creators. No compatible earlier candidate-specific companion or observation existed.

## Function records and dispatch chains

| Record | Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|---|
| MGB-F01 | LOD113C-D2GAME | 0x9FD00 -> table 0x107338, slots 43/44/45/51/52 -> 0xB31B0/0xB4760/0xB3120/0xB2FA0/0xB2F80 | none | Start dispatch and selected callbacks | MGB-R01/MGB-R03/MGB-R06 | Reused dispatcher evidence; freshly read table slots 2026-10-29 |
| MGB-F02 | LOD113C-D2GAME | 0xA19A0 -> table 0x1074A8, slots 84/86/87/94 -> 0xB65E0/0xB3060/0xB42D0/0xB2E90 | none | Action dispatch and selected callbacks | MGB-R02/MGB-R04–MGB-R06 | Reused dispatcher evidence; fresh original-image slots |
| MGB-F03 | LOD113C-D2GAME | 0xB2FA0, 0xB2F80 | none | Submerge/emerge starts | MGB-R03/MGB-R06 | Complete byte-equivalent C reading 2026-10-29; nullable branch fully read; non-null interaction meanings unverified |
| MGB-F04 | LOD113C-D2GAME | 0xB31B0 | none | MaggotEgg start | MGB-R06 | Complete byte-equivalent C reading; clears unit conditions then calls imported operation |
| MGB-F05 | LOD113C-D2GAME | 0xB65E0, 0xB3060, 0xB42D0 | none | Hatch, descend, lay | MGB-R02/MGB-R04–MGB-R06 | Complete plain-instruction bodies read for the local guards/schedule; successful callee integration remains open |
| MGB-F06 | LOD113C-D2GAME | 0xB4760, 0xB3120, 0xB2E90 | none | Other transition entry points | MGB-R06 | Instruction reading; some entries are inside a larger analysis extent, so original dispatch slots identify their actual entry; helper semantics incomplete |

Start callbacks receive game in ECX, unit in EDX, skill id and level as two stack arguments. Action callbacks have the same callback argument mapping. This mapping comes from the dispatch contract and the selected entry instructions, not from `ret 8` alone.

### Guard and start rationale

The three action entries load the skill id before their first call and use signed lower and exclusive upper bounds against the global Skills count. Each invalid-id exit returns 0. No unit access, helper call or RNG operation intervenes on that branch. MGB-R02 deliberately excludes valid-index/null-table-address behavior and successful paths.

At 0xB2FA0 and 0xB2F80, a null EDX skips the only unit-condition write, then returns 1. The first clears unit +0xC4 mask 0xE; the second sets it. Those masks' full observable interaction meanings are not inferred from the analyst labels, hence MGB-R06 stays unresolved. MaggotEgg start 0xB31B0 clears the same mask and invokes thunk 0xACA6. Import resolution identifies D2Common ordinal 10084, RVA 0x365A0. That body's path/collision calls were inspected, but their full downstream contracts were not established.

### Hatch scheduling rationale and supplied boundary

At 0xB65E0 a selection helper 0x9EE50 must return a nonnegative class, and an imported unit-related gate 0xAABA must return non-null. The handler clears unit +0x3C, evaluates Skills field +0x138 via thunk 0xAC9A, then stores the result as desired count. These preparation effects and loaded-data evaluation are outside MGB-R04's supplied-input stage.

The stage's first call to 0xEDDB0 carries class in EAX, caster in EDX, selected search argument in ECX, stack game/-1/0. Success is non-null EAX and sets one created-unit condition (unit +0xC4 mask 0x04000000), whose semantic effect remains unverified. In the repeat loop the call receives the same class/caster, ECX=8, stack game/current band/0. Success increments the success count, failure increments the band; each repeat increments its attempt count. Exit tests are success count >= desired, band >= 5 and attempt count >= signed low32(desired*2). The initial call is not charged to this last budget. At loop completion the handler calls 0xDEEE0 and returns 1. That final helper's effect is gated and excluded from the stage; the local schedule does not assert actual creation or total RNG use.

Selection helper 0x9EE50 and creator 0xEDDB0 are unresolved integration boundaries, rather than empty operations. MGB-V03 supplies only the returned identities and an explicit request log. This permits testing the count/band ordering without inventing geometry, creature semantics or helper effects. The created-unit condition and pre/post-stage writes are not included in the verified behavioral assertion.

### Laying and transition follow-through

0xB42D0 calls selection 0x9EE50, preparation 0xE07A0, target import 0xA598, coordinate helpers 0x1100/0x1130, direction imports 0xA7A8/0xAD1E/0xAC88, room import 0xA478 and creator 0xEF650. The last is a small request-builder into 0xEEF10. That original creator's entry, MonStats bounds, placement-mode selection and nearby-room preparation were inspected; its remaining placement/seeding/registration calls do not have a complete consumed contract here. This targeted follow-through establishes why merely seeing a creator call is insufficient to verify an egg.

0xB3060 reads a calculation and current/max life, performs percentage arithmetic and a life setter only on its positive-calculation path. The layout and arithmetic leads are retained, but the preceding animation/collision helper at thunk 0xACA0 and complete transition semantics remain MGB-Q03. 0xB4760 queries coordinates, invokes 0xE19C0, then changes interaction conditions and invokes imported collision mutation. Those calls were located from output writes; none is labeled verified merely from a name.

## MGB-D01: effective Skills text cells

On 2026-10-29 an observed LoD Expansion 1.13c data profile (reported original Game.exe hash `74fe9c092a521f7710392548c82f81544e531107db2358617e83818874db40a2`) read the effective `data/global/excel/skills.txt` through its loaded Storm archive set. An independent inventory of all external/loose overrides and loaded-module hashes was not captured. Physical supplying MPQ and effective-file hash were not captured. These are narrow text-cell observations, not compiled correspondence or gameplay confirmation.

All three reads succeeded: offset 0 returned 65536 of total 173535 bytes; offset 65536 returned 65536; offset 131072 returned 42463. Returned bytes were joined contiguously, preserving split rows. Full header and table coverage was established. Header lookup selected `skill` (index 0), `Id` (1), `srvstfunc` (4), `srvdofunc` (5). Rows use exact `skill` names; explicit `Id` cells identify the numeric skill rather than positional row indexing.

| Row | Column | Raw cell | Interpretation / claim |
|---|---|---|---|
| MaggotEgg | srvstfunc | `43` | start selector; MGB-R01 |
| MaggotEgg | srvdofunc | `84` | action selector; MGB-R01 |
| MagottUp | srvstfunc | `44` | start selector; MGB-R01 |
| MagottDown | srvstfunc | `45` | start selector; MGB-R01 |
| MagottDown | srvdofunc | `86` | action selector; MGB-R01 |
| MagottLay | srvdofunc | `87` | action selector; MGB-R01 |
| Submerge | srvstfunc | `51` | start selector; MGB-R01 |
| Submerge | srvdofunc | `94` | action selector; MGB-R01 |
| Emerge | srvstfunc | `52` | start selector; MGB-R01 |

### Compiled selector corroboration

On the same 2026-10-29 owned original-profile game, read-only live memory followed D2Game IAT slot RVA 0xF829C to the DATATBLS global and its loaded table structure. Structure +0xB98 supplied the Skills records and +0xBA0 reported count 357. Record stride was 0x23C. Each narrow 48-byte read checked the word skill identity at +0 and words start/action at +0x2C/+0x2E. All reads succeeded. Exact (Id, start, action) triples were (159,43,84), (161,44,0), (162,45,86), (163,0,87), (176,51,94), (180,52,0). These explicit identities agree with the named text cells and original dispatch slots. No dynamic calculation, actual AI choice or callback execution is established. Loaded-module hashes were not captured; original-profile registry/launch identity and the narrow correspondence are the measured scope. MGB-Q04 retains AI/frame entry only.

## Search limits and tools

Root ladder rung 1 used original dispatch slots and the shared dispatch companion. Rung 2 found invalid-id return and callback condition/creator writes; targeted helper reading followed those writes. Byte-equivalent reconstruction availability was checked before every new entry reading. No all-six-rungs absence claim is made. Original-image disassembly supplied the required instruction listings.

## Observation procedure

MGB-O01: use an owned original LoD Expansion 1.13c game, a character able to survive the Maggot Lair on Normal, and no table overrides. Locate a living maggot with room for an egg and a separate egg that can hatch. Record actual skill id/level and the relevant loaded selector before using their names as evidence. Observe the selected callback plus class selector, placement call and final-caster call; capture coordinates, input count and every returned unit identity. Repeat near an obstacle, and record which requests failed. For burrowing, record unit interaction/collision/animation state before descent and after emergence, rather than infer invisibility from a condition mask. Do not kill the egg before its hatch action or treat absence after such interruption as failed creation.

Callback-entry vectors require the exact supplied entry arguments, not a normal player skill animation. Natural gameplay may never pass invalid ids or absent units through the preceding dispatcher; MGB-V01/MGB-V02 remain static guard cases and were not manufactured by injected calls or memory writes. No matching maggot gameplay capture was collected in this pass. MGB-Q05 concerns naturally reachable positive hatch scheduling and successful transitions; the impossible-entry limitation does not invalidate their static branch proofs.

## Review record

Initial handoff record, 2026-10-29: self-review checked the supplied-input boundaries, required per-rule vectors, binary/private separation, source identities, current labels and dependency gates. Mechanical lint of this trio with the two sibling candidate trios returned exit 0, 0 errors and 0 warnings before the reader edits below. At that handoff, fresh reader completion was pending. The final applicable full pass 2 below completes the stopping rule for the resulting contract; neither the initial handoff nor either reader approves the spec.

### Fresh blind-implementer full pass 1 — 2026-10-29

The independent smaller reader received the contract, foundations and cited dependencies; no binary, source, gameplay or implementation inspection supplied additional rules. Reported counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 0, COVERAGE 3, HARDEST 3. The report labeled GATED count 6 but enumerated seven ids: MGB-Q01–MGB-Q05, MGB-R05, MGB-R06. The count discrepancy does not change any gate; all seven enumerated ids are already documented limits.

| Finding | Outcome | Rationale / resulting change |
|---|---|---|
| GAP: MGB-V03 intermediate successes/tries/band absent from section 5 | No change | The spec template explicitly permits named intermediate rule results in executable vectors. These counters are algorithm values, not extra gameplay writes or persisted effects; section 5 already names the ordered requests |
| COVERAGE: initial success reaches desired before repeats | Spec edit | Added MGB-V03(e), exact one-request output, distinguishing an unconditional first repeat |
| COVERAGE: negative desired count | No change | The same `desired <= 0` branch is already exercised by MGB-V03(d); negative counts do not select a different request/result branch |
| COVERAGE: positive desired doubles to a negative signed retry bound | Spec edit | Added MGB-V03(f): desired 1073741824 doubles in low 32 bits to signed -2147483648, permitting the initial request but no repeated request |
| HARDEST: separate success count, retry count and band | No change | MGB-R04 defines their independent initialization/update conditions; MGB-V03(a) separates failure-driven band changes from success-driven progress, and (b)/(c) separate stopping limits |
| HARDEST: signed identity bounds | No change | MGB-R02 expressly uses signed lower and exclusive upper comparisons, with both edges in MGB-V01 |
| HARDEST: absent-caster behavior across two start callbacks | No change | MGB-R03 names only Submerge/Emerge starts and MGB-V02 independently covers both; no other null-caster behavior is generalized |

The two additions were recomputed from unchanged MGB-R04 and checked against its original local loop: after the first request the band/retry counters initialize to 1/0, and signed exit checks suppress repeats in both new cases. No rule, source label or gameplay gate changed. Lint after these edits returned exit 0, 0 errors and 0 warnings. A fresh full pass 2 is required because executable cases and expected outputs changed.

### Fresh blind-implementer full pass 2 — 2026-10-29

The fresh smaller reader's final applicable report returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, COVERAGE 0, GATED 7, HARDEST 3. All seven gates remain the explicitly documented MGB-Q01–MGB-Q05 and MGB-R05/MGB-R06. The three HARDEST entries again concerned independent success/retry/band counters, signed bounds, and the two absent-caster starts. Each receives no change: the unchanged rule text defines those distinctions, the final vectors exercise the applicable distinct branches, and the first-pass triage already gives the specific rationale. No remaining behavioral ambiguity or discriminating-vector finding was reported.

This full pass applies to the contract including MGB-V03(e)/(f), has no unresolved GUESS/GAP/DISCRIM and completes the blind-implementer stopping rule. No further contract edit follows it. Draft status and gameplay/integration gates remain; neither reader completion nor the zero mismatch count establishes full behavior or user approval.

## Successful transition and newborn-state follow-through — 2026-10-29

This subsequent pass investigates actual successful helper effects rather than repeating unchanged retry arithmetic. The reader passes above remain scoped to their preceding contract. New interaction assignments, cell changes, life-request integration and newborn marking require a fresh applicable reader. The original `LOD113C-D2GAME`/`LOD113C-D2COMMON` fingerprints and 1.13c Windows x86 identity apply; no other build was used.

Compatible source/read boundaries were checked in the current MRH, MTP, RLI, SCX and PVP companions before reuse. Their prior dates and limits remain those of their own records. New source reading checked byte-matching availability: Submerge/Emerge starts, dynamic position accessors and the pure room resolver had a byte-matching reconstruction; selected collision/life/creator-wrapper entries did not. Original-image table cells confirmed the exact selected collision helper branches. No original table, binary or raw game capture is distributed here.

| Record | Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|---|
| MGB-F07 | LOD113C-D2GAME | 0xB2FA0/0xB2F80, 0xB3120/0xB4760 | none | Interaction trio clear/set at starts | MGB-R03/MGB-R07/MGB-R08 | Complete byte-equivalent reconstruction for first pair; original instruction bodies for maggot pair; direct assignments distinguished from wider condition meaning |
| MGB-F08a | LOD113C-D2COMMON | 0x34BB0/0x34B80 | 10867/10750; Game thunks 0xA454/0xA3B8 | Dynamic integer X/Y | MGB-R07/MGB-R08 | Complete byte-equivalent reconstruction on valid path; NRF-R10b owns prepared coordinate meaning |
| MGB-F08b | LOD113C-D2COMMON | 0x2FE10 | 10331; Game thunk 0xA478 | Current room | MGB-R07/MGB-R08 | Complete normal-path instructions for monster dynamic path; returns path room or null |
| MGB-F08c | LOD113C-D2COMMON | 0x4BCF0 -> selected 0x4BDD4 -> 0x4BBB0 -> 0x1330 | 11073 for 0x4BCF0; Game thunk 0xAD30 | Descent cross-condition clearing | MGB-R07 | Selected shape-5 branch and full five-cell mutation helper read; complete byte-matching room resolver transcription read first |
| MGB-F08d | LOD113C-D2COMMON | 0x4BE90 -> selected 0x4BEA8 -> 0x4B970/0x4B930 -> 0x1330 | 10565 for 0x4BE90; Game thunk 0xAD2A | Emergence cross-condition setting | MGB-R08 | Selected shape-1 branch, both complete mutation helpers and pure room resolver read |
| MGB-F09 | LOD113C-D2GAME | 0xB3060 -> Common 0x51BF0 / 0x38B70 / 0x390F0 / 0x3A740 | Game thunks 0xAC9A/0xA42A/0xA3A0/0xA472 | Descent calculation/read/write request | MGB-R09 | Complete caller body; exact compatible SCX/RLI contracts and MRH import map reused; actual replacing propagation remains shared gate |
| MGB-F10 | LOD113C-D2GAME | 0xB65E0/0xB42D0; consumer 0x777F0 | none | Newborn skip-kill-experience condition | MGB-R04/MGB-Q02 | Direct OR write read; exact bit-26 consumer correspondence reused from PVP-R09a and its original-image wrapper evidence |
| MGB-F11 | LOD113C-D2GAME | 0xEDDB0 -> 0xEEF10 | none | Near-actor creator request builder | MGB-R04/MGB-Q01 | Complete wrapper body and actual position/room accessors read; deep creator remains gated |

### Interaction assignments and their limits

The three logical interaction conditions A/B/C correspond to original unit +0xC4 bits 1/2/3. The same neutral names are already used by MTP-R12. Submerge start clears mask 0xE; Emerge start ORs mask 0xE, then both return 1 without a type/id/level check. MagottDown start clears the same trio before its footprint helper; a successful MagottUp relocation sets it before the footprint helper. These are actual condition assignments with complete local write evidence. They do not establish every targeting/combat/physics consumer or a claim of invisibility. Those wider meanings remain MGB-Q03. No byte positions or original unit layout are required by the behavioral abstraction.

### Actual five-cell spatial changes

At 0xB3120, the valid monster path reads Y through Common ordinal 10750 and X through 10867, obtains current room through 10331 and calls ordinal 11073 with room/X/Y/shape 5/mask 0x1000. The original dispatch table at Common RVA 0x4BDF0 has six entries; slot 5 selects RVA 0x4BDD4. That branch calls 0x4BBB0 with the same point/mask and performs no extra mutation. The complete helper clears the mask, in order, at left, center, right, up and down; it never touches diagonals.

For each point, Common 0x1330 first checks the supplied room rectangle, then its stored adjacent-room list in order, using signed lower-inclusive/upper-exclusive bounds. Its complete byte-matching transcription has no nested call. The mutation helper independently resolves each of the five points. Missing room, collision descriptor or grid skips only that point. A valid cell is selected from its containing grid and ANDed with the complement of mask 0x1000, preserving every other condition. The spec calls this one logical value the **burrow-cell condition** without claiming that it alone controls traversability, vision or missile contact. Valid room rectangles and grid coverage are inputs; malformed metadata/overflow is not assigned invented behavior.

At 0xB4760, successful relocation is followed by a new read of resulting position/room and a call to ordinal 10565 with shape 1/mask 0x1000. The original table at Common RVA 0x4BF9C, slot 1, selects 0x4BEA8. The complete selected branch calls 0x4B970 to OR the condition at the same five-cell cross, then 0x4B930 to OR mask 0x1000 at center once more. The second center update is idempotent. The two full setters use the same room/grid guards and preserve unrelated conditions. This closes the actual cell effects at the reached callback boundary without asserting the entire relocation operation or every collision consumer is verified.

Relocation Game 0xE19C0 is the exact MTP-R10 operation: game/unit/current room/current point, nearby search enabled and ordinary marker. MTP-R10 owns local nearby/path/update requests; its collision/search/synchronization gates remain MTP-Q02. The MagottUp caller returns 0 if relocation returns zero and performs no following trio/cell changes. On nonzero return it uses the *resulting* point, not the original request point. MGB-V06 supplies post-relocation spatial state and an explicit result; it does not claim failed relocation must have been side-effect free. A complete separate relocation investigation is not duplicated here.

### Actual descent expression and existing-stat reads

The valid MagottDown action sets its do-handled latch, calls Common ordinal 10569/RVA 0x2EEC0 and inspects the resulting animation frame values before the life stage. The complete accessor is a frame-information reader, not the collision helper suggested by an earlier label: for a monster it returns arithmetic-shifted unit animation fields. The preceding sequence-rate/animation bookkeeping retains MGB-Q03; its global consumers are not replaced by a guessed animation model.

From the life stage the callback evaluates Skills +0x138 (`calc1`) through the exact Common ordinal 10786 front end in context (caster, skill id, level). The matched front end and SCX contracts already read for MRH are compatible. A nonpositive result branches directly to callback result 1, with no current/max-life request. A positive result reads **current life first**, through ordinal 10973 with stat 6/layer 0, then maximum through ordinal 10913's hard-coded stat 7/layer 0. RLI-R02a/RLI-R02b apply to both selected existing arrays; this is not an assumed base-life read.

At the percentage call, EAX is the positive calculation result, EDX current life and ECX 100. Thus the multiplier basis differs from ZakarumHeal: `mulDiv(P,current,100)`, add current with low-32-bit wrap, then signed upper clamp to maximum. No lower floor occurs before the replacing write at ordinal 10887. The callback's final result is 1 regardless of that operation's normal return. Complete replacing propagation/notification remains STL-R06c/STL-Q07 and RLI-Q03. MGB-V07 executes actual supplied compiled bytes and getter arrays rather than supplying arbitrary calc/read returns; it asserts only the replacing request at the unresolved shared boundary.

### Newborn experience marking and creation-mode correction

The hatch and laying success paths OR original unit +0xC4 with 0x04000000. The compatible PVP companion's original wrapper evidence at Game 0x777F0/0x77811 reads that exact bit 26 before calling experience routing; PVP-R09a names it `skipKillExperience`. The earlier unnamed-condition gate is therefore closed for the hatch stage's direct assignment. Each non-null hatch result receives the actual assignment before its success count increments. No result means no such write. This does not establish all reward routes or actual creator-produced units when the creator itself remains supplied. The laying success has the same located write; its complete direction/creator path is still MGB-Q02.

The earlier neutral term `initialSearchValue` was too weak: full creator wrapper 0xEDDB0 stores that argument as the creation mode in the same request field as the generic creator wrapper. Its output tuple comes from the mode producer now read under MSP-R07; the first hatch attempt uses the selected mode, repeated attempts use fixed mode 8. The request's separate placement selector remains -1 then bands 1–4, without an invented search geometry. Current MGB-R04/V03 use `initialMode` and `(band,mode)`; all numeric retry traces are unchanged. This correction follows the request field's consumer, not a label guess. The compatible MSP preparation record has completed its owner's fresh review while retaining its profile/AI-producer limits; no full original-profile class producer or deep creation is inferred here.

The complete wrapper 0xEDDB0 obtains the actor's current room and integer position, fills the creator descriptor with game/room/class/mode/point/placement selector/final option, and calls 0xEEF10. It has no direct RNG step. The core creator still owns seeding, collision search, allocation, registration and later effects. The final hatch call 0xDEEE0 is a death-processing request with the caster as both victim and attacker and trailing control 1, followed by local return 1. Complete death/mode/reward effects remain the original death owner's integration boundary, rather than inferred from that request or the newborn marker.

### Current continuation status

MGB-R03 now closes non-null Submerge/Emerge assignments, MGB-R04 closes the actual newborn experience-condition write and correct mode binding, MGB-R07/MGB-R08 close the selected actual grid-cell helper effects, and MGB-R09 closes the actual descent expression/getter-to-write-request stage. These are successful helper/effect assertions beyond invalid-entry guards and supplied retry counts. Deep creation/final death, laying direction/creator integration, wider interaction/animation/physics, actual replacing propagation and naturally triggered observations remain explicit.

At the continuation draft handoff, the preceding reader completion applied only to the earlier contract and the new behavior still required a fresh smaller reader. The final applicable full/focused records below now complete that requirement. Self-review checked coordinates and room-edge resolution, one missing grid versus whole-cross abort, exact selected table slots, preservation of other conditions, post-relocation point/control use, current versus maximum life basis, signed wrap/no-floor behavior, expression/getter ownership, newborn marking and the distinction between direct cell effects and wider physics. No matching gameplay capture was collected in this continuation.

For the pending MGB-O01 continuation, record the logical interaction trio and mapped burrow-cell condition before/after a naturally reached descent/emergence. Read the center, four orthogonal neighbors and at least one diagonal, including a room-edge case where practicable. Record relocation's actual return and resulting point separately from its earlier request. At the descent action, capture actual `calc1` result, selected current/max-life read values, replacing request and later stored value as separate assertions. For a hatched unit, record the skipKillExperience condition immediately after its creator result before later state changes. These procedures seek matching observed cases; none was executed by the new static reading.

### Continuation fresh blind-implementer full pass 1 — 2026-10-29

The fresh smaller reader used the latest behavioral contract and cited dependencies without obtaining additional behavioral rules from original binaries, implementation or gameplay. Counts were MISMATCH 0, GUESS 0, GAP 0, DISCRIM 2, COVERAGE 4, HARDEST 3 and reported GATED 9. The nine distinct question references represented by its grouped gate bullets are MGB-Q01–MGB-Q05, MTP-Q02, SCX-Q01, STL-Q07 and RLI-Q03; its references to MGB-R05/MGB-R06 name the already gated laying/remaining-transition rules, not additional uncertainty invented by this review. Supplied valid compiled bytes remain executable independently of SCX-Q01, and no runtime confirmation is inferred.

| Finding | Outcome | Rationale / resulting change |
|---|---|---|
| DISCRIM: MGB-V06's unchanged room lets a stale-room implementation pass | Spec edit | V06 now requests room A but supplies resulting room D. A has no collision descriptor or adjacent fallback, while D/B have the required grids; expected changes occur specifically in D/B. The existing changed-point test is retained |
| DISCRIM: MGB-V07 did not require the maximum clamp to change a candidate | Spec edit | Added V07(e): current 2560, positive percentage 10, maximum 2600; candidate 2816 must request raw 2600 |
| COVERAGE: strict upper-exclusive room edge | Spec edit | Added V05(d), placing the right cross point at x=102, exactly A's exclusive upper edge; it must update an adjacent grid instead of A |
| COVERAGE: first-match order for overlapping adjacent rooms | Spec edit | The same V05(d) supplies ordered [B,C] with separate overlapping grids. Only B's point is cleared; C remains true. This is a synthetic selector fixture, not an assertion about generated room overlap |
| COVERAGE: missing collision descriptor versus missing grid | No change | Both guards skip the one point and continue the other cross points. V05(b) already distinguishes this per-point continuation result from aborting the entire footprint; an absent descriptor produces the same callback result and remaining-cell recipe. No new shared-helper semantics are assumed |
| COVERAGE: new post-relocation room | Spec edit | Resolved by the V06 room-discrimination change above, retaining its post-operation snapshot boundary |
| HARDEST: independent hatch successes/retries/bands and signed wrap | No change | MGB-R04 gives each update/test separately; V03(a)/(b)/(c)/(e)/(f) distinguish successful progress, failure-driven band change, repeat budget, initial completion and signed doubled-bound wrap |
| HARDEST: per-cell room/grid guards | Spec edit for edge/order coverage; otherwise no change | R07 already defines independent room resolution and skip behavior; V05(d) strengthens edge/order evidence while V05(a)/(b)/(c) retain cross-room, partial-grid and absent-room results |
| HARDEST: SCX/RLI/mulDiv/wrap/clamp/request boundary | Spec edit for maximum-clamp discrimination; otherwise no change | V07(e) supplies the missing changed-clamp case. Existing cases compute actual expression/getter results, current-life basis and signed wrap; actual replacing propagation stays the named dependency gate |

The changed outputs were recomputed from unchanged rules: V05(d)'s x=102 lies outside A and resolves first to B; V06 must use D/B after relocation rather than the descriptor-less A; V07(e) gives increase 256, candidate 2816 and final request 2600. No rule, evidence label or gate changed. The resulting contract requires the fresh focused review of V05(d), V06 and V07; the full pass above does not approve the edited fixtures.

### Continuation fresh focused pass 2 — 2026-10-29

The fresh smaller reader checked the changed MGB-V05(d), MGB-V06 and MGB-V07 material against the unchanged rules and dependency contracts. Its final applicable report returned MISMATCH 0, GUESS 0, GAP 0 and no outstanding DISCRIM finding. The report's “5/5 requested mutations distinguished” describes five successful checks of plausible incorrect alternatives, not five new findings: inclusive A upper edge, selecting the second overlapping neighbor, using the old room, using the old point, and omitting the maximum clamp are all rejected by the exact revised outputs.

The reader confirmed that x=102 lies outside A and selects first neighbor B while leaving C unchanged; the emergence stage must use the post-relocation D/B cross because old room A has no descriptor, while the old point fails the diagonal/left-cell assertions; descent gives increase 256, candidate 2816, clamped request 2600 and callback return 1 even when the supplied writer returns 0. Existing supplied-relocation limits MTP-Q02/MGB-Q03 and actual replacing-storage STL/RLI gates remain; the focused pass makes no new gameplay or propagation claim.

| Finding | Outcome | Rationale |
|---|---|---|
| HARDEST (1): MGB-V05(d)'s strict upper edge and first matching adjacent room | No change | The explicit rectangle ends place x=102 outside A; the stated [B,C] order chooses B, and separate grid assertions require B=false/C=true. The reader independently recomputed those outputs and rejected both incorrect alternatives |

This is the final applicable continuation review. MISMATCH is zero and no unresolved GUESS/GAP/DISCRIM remains. All full-pass and focused-pass findings have recorded outcomes; no subsequent contract, vector, source label or gate edit is needed. The files remain draft with their explicit creation, other-transition, physics/storage and observation limits. Review completion does not infer user approval or runtime confirmation.
