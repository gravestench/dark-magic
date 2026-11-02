# Evidence: monster resurrection, healing and blood sucking

Contract: [monster resurrection, healing and blood sucking](../../specs/skill-monster-resurrect-and-heal.md). [Evidence gaps](../../plans/gaps/skill-monster-resurrect-and-heal.md) remain explicit.

## Identity, reuse and boundaries

Original Windows x86 Expansion 1.13c `LOD113C-D2GAME` is identified in the [source manifest](../source-manifest.md). The image's SHA-256 was checked on 2026-10-29 and matched `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`. All addresses below are RVAs unless explicitly VA. Descriptions are analyst labels, not original exported names.

The [shared dispatch companion](skill-cast-start-and-do-dispatch.md) supplies its 2026-10-12/15 dispatcher reading with its original limits. New 2026-10-29 table reads identify the monster callbacks; they do not complete AI/frame or downstream helper contracts. No earlier candidate-specific spec/companion observation was available to reuse. NRF-R01/R02/R05/R08a supply compatible numeric contracts, including the exact two-input multiplication order and draw behavior. Their prior verification is not renewed by this new caller reading.

## Function records and roots

| Record | Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|---|
| MRH-F01 | LOD113C-D2GAME | 0xA19A0 -> table 0x1074A8: 96 -> 0xB3770; 97 -> 0xB5BD0; 107 -> 0xB3B20; 127 -> 0xCF00 | none | Action dispatcher to healing/revival/blood actions | MRH-R01–MRH-R05 | Reused dispatcher; original-image slot read 2026-10-29 |
| MRH-F02 | LOD113C-D2GAME | 0xB3770 | none | Healing action | MRH-R02/MRH-R03 | Complete local plain-instruction body read 2026-10-29; supplied calculation/getter/write results separate local arithmetic from real effects |
| MRH-F03 | LOD113C-D2GAME | 0xB5BD0 | none | Resurrection action | MRH-R02/MRH-R04 | Instructions read; original-image disassembly 0xB5BD0–0xB5D40 corroborated encoded spans and targets; downstream revival/mode chain remains open |
| MRH-F04 | LOD113C-D2GAME | 0x9FD00 -> table 0x107338: 55 -> 0xB35A0; 60 -> 0xC760; 61 -> 0xC1B0 | none | Mosquito, blood and self-revival starts | MRH-R01/MRH-R04/MRH-R05 | Original slots read; 0xC1B0 entry has no separately indexed reconstruction extent, original instructions read directly; no absence inference |
| MRH-F05 | LOD113C-D2GAME | 0xCF00, 0xC760 | none | Suck Blood action/start | MRH-R02/MRH-R05/MRH-R09 | Action's local recipient/life chain completed on 2026-10-30 below; complete start/queued combat integration remains open |
| MRH-F06 | LOD113C-D2GAME | 0xB33B0 | none | Shared revival operation | MRH-R04 | Complete local instruction body followed; stat/path/registration/room imports and local mode helpers incomplete |
| MRH-F07 | LOD113C-D2GAME | 0xB35A0, 0xB3B20 | none | Mosquito start/action | MRH-R05/MRH-R10 | Complete original-image corroboration 2026-10-30; start's actual preparation writes established below; downstream action combat/repeat effects remain open |

The root table's callbacks receive game in ECX, unit in EDX, skill id and level on the stack. This follows the shared caller contract and entry reads. MRH-R03 starts later and uses semantic supplied values, not a required binary calling convention.

## Healing arithmetic rationale

0xB3770 rejects negative/too-large skill ids before any call. With a valid record it calls preparation 0xE07A0 and resolves a target via thunk 0xA598, rejecting self or absent targets. It sets unit +0xC4 mask 0x40 before evaluating upper Skills field +0x13C, then lower field +0x138 through thunk 0xAC9A. These preceding effects are MRH-Q01, not claimed as no-ops by the arithmetic contract.

The upper result uses a signed positive test and a cap at 100; nonpositive results become 0. The lower result is clamped to zero and then to upper. ESI holds upper-lower at the call to bounded helper 0x11D0, with ECX pointing to caster seed +0x20. The helper contract is NRF-R02; a nonpositive range does not draw. The returned roll adds to lower. The next calls obtain target maximum life and stat 6/layer 0. At 0xB3849–0xB3858, EAX is percentage, EDX maximum life, ECX=100 before 0x14D0. Thus operand order is mulDiv(percent,maximumLife,100), not interchangeable mathematical multiplication. Current life is added in 32 bits, raw values <=1 become raw 1, then a signed comparison limits to supplied maximum before stat 6/layer 0 setter. The routine returns 1 after that request.

| Semantic operation | D2Game thunk / local RVA | Resolved original target | Established use / limit |
|---|---|---|---|
| Skill calculation | 0xAC9A | LOD113C-D2COMMON ordinal 10786, RVA 0x51BF0 | Evaluation caller order identified; actual loaded calculation integration remains MRH-Q01 |
| Maximum-life result | 0xA3A0 | LOD113C-D2COMMON ordinal 10913, RVA 0x390F0 | Used as percentage basis and upper clamp; supplied result in MRH-R03 |
| Base-stat read | 0xA42A | LOD113C-D2COMMON ordinal 10973, RVA 0x38B70 | Target/stat 6/layer 0 request; supplied result in MRH-R03 |
| Stat write | 0xA472 | LOD113C-D2COMMON ordinal 10887, RVA 0x3A740 | Target/stat 6/layer 0/value; actual stat callback/storage effects unverified here |
| Bounded draw | 0x11D0 | LOD113C-D2GAME same RVA | NRF-R01/R02; caller seed and modulus followed |
| Percentage operation | 0x14D0 | LOD113C-D2GAME same RVA | NRF-R05; exact caller operand order followed |

The write operation is logged in vectors and is not credited as a verified stored-stat change. No unread local callee is hidden inside the verified arithmetic stage: only the two named numeric contracts execute there; the other operations are explicit supplied boundary values/results or logged requests.

## Revival and blood follow-through

The original-image listing of 0xB5BD0 identifies target predicate imports 0xA688/0xAD4E, target room lookup 0xA478, collision query 0xAC6A, path relation 0xA5FE, revival helper 0xB33B0, monster mode request sequence 0xF3CB0 -> 0x75070 -> 0xF3CB0 -> 0x75E10, and optional target-state request 0xA83E. This settles the encoded local branch/call reading but not those helper effects. 0xB33B0's complete local body restores life from stat 7 into stat 6 and has monster-specific interaction, registration and collision work. Those local observations do not establish full corpse eligibility/lifecycle or the semantics of all room/mode calls; MRH-R04 remains unresolved.

Self-start 0xC1B0 rejects absent/non-monster units, calls 0xB33B0, clears an interaction-condition mask and invokes thunk 0xA682 with a path and mask 0x100. A successful local return is insufficient to establish completed resurrection. Its full original-image entry was read directly because the selected entry lacked a separate analysis boundary; this is a located function, not a not-found result.

Suck Blood action 0xCF00 calls preparation, target lookup and combat-record selection 0xDAA90. Its conditional path invokes owner selection 0x21260, compares a combat-record quantity to target current life, evaluates a skill percentage, and requests recipient life through the same numeric helper before calling 0xDED30. The start at 0xC760 constructs a damage request through 0xA11B0, invokes 0xDCE00, evaluates multiple calculations and calls 0xDDDE0. Neither full combat path nor recipient selection has a supplied owned dependency contract here. Their naming remains a lead, not verification.

Mosquito start 0xB35A0 derives a temporary seed from caster/target identifiers and writes a skill-associated target/count result; action 0xB3B20 uses target lookup, repeated bounded draws, damage-building helpers, a percentage life request and later combat application. It is distinct from the Suck Blood family, and no single RNG count or recipient rule is generalized across them. These targeted instruction reads are retained as MRH-Q03's attempted slice; encoded entry bytes and complete damage/target helper chains remain unverified.

## MRH-D01: named Skills text cells

Reuse of the same 2026-10-29 complete effective-file read is documented in [MGB-D01](skill-monster-maggots-and-burrowing.md#mgb-d01-effective-skills-text-cells): source profile, exact contiguous byte ranges, full coverage, named-header parsing, archive-loader method and identity limits all apply. This record uses the same original LoD profile and file, not a renewed read or gameplay capture. Header indices for `calc1` and `calc2` were 186 and 188. Exact row names select the cells; numeric skill ids use the explicit `Id` field, never positional indexing.

| Row | Column | Raw cell | Supported claim |
|---|---|---|---|
| ZakarumHeal | srvdofunc | `96` | MRH-R01 |
| Bestow | srvdofunc | `96` | MRH-R01 |
| Resurrect | srvdofunc | `97` | MRH-R01 |
| Resurrect2 | srvdofunc | `97` | MRH-R01 |
| Self-resurrect | srvstfunc | `61` | MRH-R01 |
| Mosquito | srvstfunc | `55` | MRH-R01 |
| Mosquito | srvdofunc | `107` | MRH-R01 |
| Suck Blood | srvstfunc | `60` | MRH-R01 |
| Suck Blood | srvdofunc | `127` | MRH-R01 |
| ZakarumHeal | `calc1` | `15+5*lvl` | MRH-R01 text expression only |
| ZakarumHeal | `calc2` | `50` | MRH-R01 text expression only |

### Compiled selector corroboration

The same 2026-10-29 read-only loaded-table procedure described in MGB-D01 checked explicit id/start/action words for these records. Each 48-byte read succeeded; (Id, start, action) triples were (179,0,96), (182,0,96), (181,0,97), (327,0,97), (293,61,0), (206,55,107), (289,60,127). They agree with the named text cells and original dispatch slots. No calculation expression, target eligibility, AI choice or action execution was established by those narrow reads. Effective-file hash, physical supplying MPQ and loaded-module hashes remain uncaptured. MRH-Q04 retains actual AI/frame entry; no real calculation execution follows from reading text or selector words.

## Observation procedure

MRH-O01: on an owned original LoD Expansion 1.13c game, use a character able to survive the Travincal area on Normal and observe a healing-capable monster near a damaged eligible ally. First establish the actual skill's loaded selector/level and the ally's eligibility. Record entry, both actual calc results, caster seed before/after, target current/max raw life, setter request and final stored value. A vector match needs those inputs and the exact ordered draw and write, not just a visually increasing life bar. Different input values are a new attributable case; do not fit them to an existing vector. A resurrection procedure additionally records the corpse mode/room/collision state before action, successful revival and subsequent mode/state output, but exact expected outcomes remain MRH-Q02. Mosquito and Suck Blood must be observed separately.

On 2026-10-29 an original-profile level-1 Barbarian was in the starting town. A 5-second entry trace at D2Game 0xB3770, capped at 5 records, returned count 0 and dropped 0. No healing monster or eligible target was present or triggered. This is an inconclusive observation, not an executed heal, negative identity result or runtime-confirmed vector. The absent trigger prevents comparison with a healing vector. Matching gameplay remains pending and the stage vectors retain `runtime_observed: no`.

Natural gameplay is unlikely to enter a handler with an invalid skill id or absent caster after the shared dispatcher. MRH-V01 is a static entry-guard vector; no injected call or manufactured memory state is claimed as gameplay confirmation.

## Search and review limits

Dispatch-table root and local output-call anchors were identified. Targeted complete local instruction readings reached numeric writes and the helper boundaries above. This is not an all-six-rungs absence claim. New byte-equivalent reconstruction availability was checked before instruction reads; none covered the selected healing/revival/action entries. Initial handoff record, 2026-10-29: self-review checked exact operand order, clamp order, raw units, no-draw versus modulus-one behavior, vector boundaries, source identities and dependency gates; mechanical lint of this trio with the two sibling candidate trios returned exit 0, 0 errors and 0 warnings. Fresh blind-reader completion was pending at that handoff. The final applicable full pass 2 below completes the stopping rule; neither review nor lint approves the spec or removes its gates.

### Fresh blind-implementer full pass 1 — 2026-10-29

The independent smaller reader used the behavioral contract and cited dependencies without binary, source, gameplay or implementation inspection. Counts were MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, COVERAGE 2, HARDEST 3, GATED 7. The seven enumerated gates are MRH-Q01–MRH-Q05 and MRH-R04/MRH-R05, all already explicit in the contract.

| Finding | Outcome | Rationale / resulting change |
|---|---|---|
| DISCRIM: swapping the first two mulDiv operands leaves MRH-V04's result unchanged | No change to MRH-V04; separate spec edit | MRH-V04 expressly discriminates exact multiply-then-divide from original divide-first arithmetic, not operand swapping. That stated claim remains valid. Added MRH-V06(a) independently discriminates the specified operand order using percent 25 and maximum 100007: correct increase 25001 versus swapped increase 25000. A percent above the rule's 100 cap would not be a valid distinguishing percentage input |
| COVERAGE: current life plus increase wraps i32 | Spec edit | MRH-V06(b) supplies explicit isolated values producing raw -2147482696 before the floor; final raw write 1 rejects widened/saturating addition |
| COVERAGE: negative supplied maximum and signed upper clamp | Spec edit | MRH-V06(c) supplies a negative maximum and requires raw -1000 after floor-then-maximum order. It explicitly makes no claim that normal target preparation admits that input |
| HARDEST: mulDiv branches and truncation | Spec edit for operand discrimination; otherwise no change | The new case separates operand order, while MRH-V04 already separates divide-first from exact arithmetic; NRF-R05 supplies all thresholds and signed truncation |
| HARDEST: raw floor before maximum | No change | MRH-R03 states the order; MRH-V03(c) already tests zero maximum, and MRH-V06(c) adds a signed-negative example without changing the rule |
| HARDEST: guard before target operations | No change | MRH-R02 and MRH-V01 explicitly require rejection before calls, writes or draws, with an absent caster/target in the invalid-id cases |

The added cases were recomputed from unchanged MRH-R03/NRF-R05: 2500175/100 truncates to 25001, swapped divide-first trunc(100007/100)*25 is 25000, addition 2147484600 wraps to -2147482696, and signed -250+100 floors to 1 then clamps to maximum -1000. No source labels, rule behavior or gates changed. MRH-V06 is an isolated supplied arithmetic exercise, not a game-ready fixture or evidence that actual preparation admits every signed input. Mechanical lint of all three candidate trios after these edits returned exit 0, 0 errors and 0 warnings. Fresh full pass 2 is required because executable cases and acceptance references changed.

### Fresh blind-implementer full pass 2 — 2026-10-29

The fresh smaller reader's final applicable report returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 7, COVERAGE 2, HARDEST 3. Its seven gates remain MRH-Q01–MRH-Q05 and MRH-R04/MRH-R05, all already explicit. It recomputed the contract including MRH-V06; no behavioral correction follows this report.

| Finding | Outcome | Rationale |
|---|---|---|
| COVERAGE: NRF-R05's first operand > 65536 branch | No change | MRH-R03 normalizes percentage to 0–100 before passing it as the first operand. That helper branch is outside this caller's supplied-input domain; shared-helper branch coverage belongs to NRF-R05 |
| COVERAGE: not every bounded-draw result enumerated | No change | NRF-R02 owns bounded-draw generation; MRH-V02 tests the caller's modulus 30 and offset with an exact seed/result, MRH-V05 tests modulus 1 advancement, and MRH-V03 tests no-draw equality. Exhaustively enumerating helper outputs would duplicate its contract |
| HARDEST: id guard before target helpers | No change | MRH-R02 and MRH-V01 state and exercise the guard before helpers, writes and draws |
| HARDEST: mulDiv branch and operand order | No change | MRH-R03 fixes operand order, MRH-V04 distinguishes divide-first arithmetic, and MRH-V06(a) distinguishes swapped operands |
| HARDEST: raw floor before signed maximum clamp | No change | MRH-R03 specifies both ordered signed clamps; MRH-V03(c)/MRH-V06(c) distinguish zero and negative maxima |

MISMATCH is zero and no unresolved GUESS/GAP/DISCRIM remains. The advisory coverage findings are resolved as above; the blind-implementer stopping rule is complete for this unchanged resulting contract. Draft status, source labels, integration gates and pending gameplay observations remain. Reader completion establishes neither complete behavior nor user approval.

## Successful healing callback follow-through — 2026-10-29

This subsequent evidence pass addresses the initial MRH-Q01 target/preparation/calculation/getter chain rather than repeating unchanged arithmetic reviews. The preceding full passes remain historical records for the contract they read. MRH-R06/MRH-R07 and MRH-V07/MRH-V08 add normal-entry callback behavior; they require a fresh applicable reader. MGB/MSP contracts were not changed in this pass.

### Identity and compatible dependency reuse

The original D2Common image was freshly hashed and matched `LOD113C-D2COMMON` SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`. D2Game remains the hash-matching specimen identified above. New target-accessor and expression-front-end readings used byte-equivalent C where available; preparation and registry lookup had no such reconstruction and were read as instructions. Explicitly scoped original-image instruction listings corroborated every encoded span used for preparation, registry lookup and the replacing-write wrapper. Decompilation was a reading aid only; no shared program selection or binary was modified.

RLI-R02a/RLI-R02b's existing-array read evidence and SCX-R01–SCX-R11/SCX-R20/SCX-R21's compiled-execution/parameter evidence were checked for matching original artifact identity and assertion boundary before reuse. Their original dates/methods remain those in their own companions. This caller adds which target, stat, calculation field and context is supplied; it does not claim to reverify their implementation or close their producer/storage gaps.

| Record | Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|---|
| MRH-F08 | LOD113C-D2GAME | 0xB3770 -> 0xE07A0 -> 0xE03A0 | none | Healing target normalization and registry resolution | MRH-R06 | Complete normal-entry instruction path; original-image listings corroborate encoded branches/calls; invalid context diagnostics outside stated valid-entry boundary |
| MRH-F09a | LOD113C-D2COMMON | 0x348F0 | 10392; Game thunk 0xA598 | Selected target accessor | MRH-R06 | Complete byte-equivalent reconstruction: null path returns null, otherwise path selected target field |
| MRH-F09b | LOD113C-D2COMMON | 0x34910 | 11163; Game thunk 0xA592 | Cached target identity accessor | MRH-R06 | Complete byte-equivalent reconstruction on non-null selected-target path; reads cached type/id |
| MRH-F09c | LOD113C-D2COMMON | 0x35040 | 10336; Game thunk 0xA5FE | Selected target setter | MRH-R06 | Complete byte-equivalent reconstruction: clears pointer on null without changing cached pair; non-null setter copies unit type/id |
| MRH-F10a | LOD113C-D2COMMON | 0x51BF0 | 10786; Game thunk 0xAC9A | Skill calculation front end | MRH-R07 | Complete byte-equivalent reconstruction; SCX-R11 consumes pool bounds and exact (caster, skill, level) context; SCX owns interpreter/parameter execution |
| MRH-F10b | LOD113C-D2COMMON | 0x390F0 -> 0x38A80 | 10913; Game thunk 0xA3A0 | Maximum-life front end | MRH-R07 | Complete plain-instruction wrapper: hard-coded stat 7/layer 0, definition guard and shared selected-array getter; RLI-R02a/b own getter interior |
| MRH-F10c | LOD113C-D2COMMON | 0x38B70 -> 0x38A80 | 10973; Game thunk 0xA42A | Current-life read front end | MRH-R07 | Existing compatible RLI-R02a/b contract plus actual caller stat 6/layer 0 traced |
| MRH-F11 | LOD113C-D2COMMON | 0x3A740 -> 0x3A280 -> 0x39F70 | 10887; Game thunk 0xA472 | Replacing writer and propagation handoff | MRH-R07/MRH-Q01 | Wrapper encoded bytes corroborated; direct replacement body and propagation loop read; full stored aggregate/entry/callback effects remain STL-R06c/STL-Q07/RLI-Q03 |

The target helper descriptions deliberately differ from misleading analyst labels in some available listings: the matched code accesses a selected target, not a room or world coordinate. Original path fields +0x58/+0x5C/+0x60 map to selected target/cached type/cached id. Their numeric layout is investigation metadata only and is not a required behavioral representation.

### Actual target normalization

At 0xB37A6 the valid heal path calls 0xE07A0 with game in ECX and caster in EAX. The helper requires non-null game/caster/path on its normal path. It reads selected target through Common 0x348F0. If selected target is null, it returns without looking up the cached pair. Otherwise Common 0x34910 reads cached type/id and Game 0xE03A0 resolves them in that game's unit registry. That lookup has no calls on valid inputs: for types 0–4 it chooses the type-specific registry bucket and walks unit links to the matching id; type 5 has its own chain. The original-image type table at Game RVA 0xFB014 contains offsets 0x1120/0x1320/0x1520/0x1920/0x1720 for types 0–4, with type 5's chain at game +0x1B20. These anchors establish the lookup's game/type/id role, not an imposed registry storage design.

The preparation compares the resolved unit identity with the selected pointer. A mismatch clears selection via Common 0x35040(null), rather than selecting the resolved replacement. It also clears selection when the resolved unit is type 4 with mode 1 or 2. The null setter only changes the selected pointer; cached type/id remain unchanged. No stat or RNG operation occurs. Back in 0xB3770, the selected pointer is read again and self/absent targets return 0. Only after those checks does the caller set unit +0xC4 mask 0x40, the do-handled latch already mapped by the shared dispatch companion. This closes local preparation/selection effects for the explicit valid normal-entry boundary without treating every cached identity as a valid live target.

MRH-V07's canonical registry replacement case distinguishes stale-target clearing from retargeting, while its two excluded item modes and allowed item-mode-zero case establish the actual narrow mode test. No friendly/living/range/monster-only predicate exists in this handler's local target gate; upstream AI eligibility is not inferred from the helper label.

### Actual calculation and existing-stat reads

The callback requests the field at Skills +0x13C (`calc2`) first using (caster, skill id, callback level), clamps upper to 0–100, then requests +0x138 (`calc1`) in the same context before clamping lower. It requests the second calculation even when upper was normalized to zero. The complete matched Common 0x51BF0 body enforces unsigned pool bounds and forwards the selected byte slice plus that context to the shared evaluator. SCX-R11 and the requested opcode/parameter rules therefore supply actual expression execution rather than arbitrary calculation returns. Any expression RNG occurs before the caller's local bounded draw; no claim of zero expression draws is generalized from literal-only vectors.

After the local draw, Game thunk 0xA3A0 reaches Common 0x390F0. It returns zero for a null unit/list or unavailable stat-7 definition; otherwise it selects exactly definition 7 and key (7,0), then invokes the very same Common 0x38A80 getter covered by RLI-R02a/RLI-R02b. The next current-life call goes through ordinal 10973 with stat 6/layer 0 into that getter. Thus existing-array selection and conditional minimum adjustment are established through the exact owning contract. The initial phrase “base-life read” was too narrow: ordinal 10973 can read resolved stats, and the new current contract says current life instead. MRH-V08 supplies distinct own/resolved arrays so this correction affects an executable assertion rather than terminology alone.

MRH-V08's synthetic compiled bytes execute literal 50 (or zero), and literal 15 plus literal 5 times the supplied level token. The exact pool bytes and valid Skills context are supplied, so this verifies the complete caller-to-expression-to-read-to-request chain without assuming the unverified text compiler emitted those bytes or that the original loaded ZakarumHeal expression pool was captured. It deliberately leaves actual replacing storage supplied and asserts the actual stat-array reads before that boundary.

### Replacing writer: narrowed remaining gate

The callback pushes target, stat 6, computed raw value and layer 0 to Game thunk 0xA472 and unconditionally loads result 1 after normal return. Original Common 0x3A740 forwards the target's stat list into list writer 0x3A280, then can perform additional player changed-stat bookkeeping when the list writer reports a change. The callback does not inspect that result. MRH-V08 supplies a normal return of 0 and still requires callback result 1, without pretending that the real writer always fails or has no side effects.

0x3A280 locates or creates an own-stat entry, computes new-minus-old, writes/removes the own entry, and calls 0x39F70 to propagate the delta. The normal-path propagation listing checks the loaded stat definition and list conditions, chooses direct aggregate addition 0x389E0 or dependent-stat evaluation 0x39CE0, and can traverse parents. Direct addition can create/remove resolved entries and invoke a registered callback; dependent evaluation has additional recursive definition cases. Those complete entry-management, aggregate, callback and changed-stat effects are exactly the already unresolved STL-R06c/STL-Q07 and RLI-Q03 boundary. The present callback spec does not establish a parallel generic stat authority or infer complete stored healing from the earlier own-entry write. This is now the specific remaining healing storage gate; it no longer blankets local target normalization, expression-front-end requests or stat getters.

### Review and continuation status

The initial MRH-Q01 target/preparation and existing-getter unknowns are closed within MRH-R06/MRH-R07's explicit valid normal-entry and shared-contract boundaries. Actual arbitrary expression producers retain their own SCX gates, and the original-profile compiled ZakarumHeal formula/AI execution is not newly observed. MRH-Q01 is narrowed to actual replacing storage/propagation/notification, with the shared owner identified. Resurrection and blood-sucking gates were not expanded or promoted. No eligible healing invocation was observed in this continuation, so no gameplay confirmation is added.

MRH-R06/MRH-R07 and MRH-V07/MRH-V08 are new behavioral material and the base-versus-resolved input correction changes an asserted getter integration. The old full reader passes remain scoped to their preceding contract. Self-review checked selected-target versus registry identities, cached-pair persistence, current/max selected-array reads, evaluator context/order, token execution, local versus dependency draws, callback result versus storage effect, and complete named dependency boundaries. At the continuation draft handoff a fresh smaller reader was pending; the final applicable continuation pass below completes that requirement. Unchanged SCX/RLI contracts are reused with their existing review/evidence limits.

### Continuation fresh blind-implementer full pass 1 — 2026-10-29

The fresh smaller reader examined the resulting contract and cited dependencies, including the full callback extension, without obtaining additional behavioral rules from binary, source, gameplay or implementation inspection. It reported MISMATCH 0, GUESS 0, GAP 0 newly introduced, DISCRIM 0, GATED 0 newly improper, COVERAGE 1 and HARDEST 1. The qualified GAP/GATED counts do not erase existing documented uncertainty: MRH-Q01's actual replacing storage and the applicable SCX producer limits remain explicit, as do the separate resurrection/blood and observation gates.

The reader independently decoded MRH-V08's upper literal 50 and lower expression 15+5*level, selected the resolved maximum 25600/current 2560 through RLI, and recomputed the raw write requests 13056 and 2560. It confirmed that the supplied writer's normal 0 return is ignored by the callback, which returns 1. All executable MRH-V01–MRH-V08 cases recomputed correctly.

| Finding | Outcome | Rationale |
|---|---|---|
| COVERAGE: MRH-V07's item-mode-zero success stops at the MRH-R07 handoff, whereas MRH-V08 continues with a monster target | No change | MRH-R07 adds no further local target-kind branch. RLI-R02a/RLI-R02b own selected-array and owner-kind minimum semantics for the retained target; SCX owns the caster-context expression execution. The shared replacing writer's kind-specific stored/notification effects remain MRH-Q01. A second complete item fixture would repeat the local request sequence rather than exercise another local branch; no claim is made that item and monster getter values or real stored effects must match |
| HARDEST: synthetic compiled pool plus supplied writer proves the local callback-to-request path, not original compiled formulas or persistence/notification | No change | MRH-V08 expressly supplies valid synthetic bytes, actual existing arrays and a replacing-write operation at the named boundary. Its inputs, outputs, discrimination and acceptance exclude original compiler-produced bytes and actual replacing-storage/notification effects. MRH-R07 and MRH-Q01 make the same distinction; no evidence label is promoted to runtime confirmation |

This is the final applicable reader pass for the continuation contract. MISMATCH is zero, no unresolved new GUESS/GAP/DISCRIM or improperly hidden gate was reported, and every advisory finding has the outcome above. The blind-implementer stopping rule is complete without a subsequent contract, vector or gate edit. The spec remains draft with the stated storage, other-action and gameplay limits; reader completion does not infer user approval.

## Sustained storage and encounter investigation — 2026-10-30

This pass reuses the preceding 2026-10-29 callback, getter and expression records with their original limits. The original D2Common image was freshly fingerprinted and again matched `LOD113C-D2COMMON`. The generic replacing-write owner is still STL-R06c/STL-Q07 and RLI-Q03. These additional source and data records do not replace that authority or complete MRH-Q01.

### Life definition and direct propagation evidence

An owned original-profile LoD 1.13c process returned the full effective `data/global/excel/itemstatcost.txt` through its archive loader: requested offset 0/length 64000, returned 40109/total 40109 bytes. Header-based parsing selected row `hitpoints`, explicit `ID` cell `6`. Named cells `fCallback`, `keepzero`, `op`, `op stat1`, `op stat2` and `op stat3` were empty; `direct` was `1`, `maxstat` was `maxhp`. Empty cells are retained as empty text here, rather than guessed numeric compiler output. No data override was requested and the process reported the original LoD profile; the effective-file hash, physical supplying archive and loaded-module hashes were not captured.

A separate read-only compiled-record check followed Game's import slot 0xF829C to the shared data-table context, then its ItemStatCost pointer/count at +0xBCC/+0xBD4. The count was 359, record stride 0x144, record 6's id was 6, bytes +4/+5 were both 0x10 and bytes +0x50–+0x53 were all zero. The Common import slot at 0x890B0 resolves Fog's exported `gdwBitMasks`; its first six observed words were 1, 2, 4, 8, 16 and 32. These narrow compiled controls agree with the direct, non-callback life-definition path; the read does not capture a healing invocation or establish an original expression-pool producer. Its source-profile limits remain those above.

| Artifact / function | Analyst description | New reading / remaining boundary |
|---|---|---|
| LOD113C-D2COMMON 0x3A280, ordinal 10261 | Own-stat replacement | Complete direct body reread: equality returns 0; a changed nonzero value is stored before propagation of new-minus-old. Entry insertion/removal and special-list effects are not assumed away |
| LOD113C-D2COMMON 0x382B0 | Exact entry lookup | Compatible byte-equivalent reconstruction reread; searches the active sorted entries |
| LOD113C-D2COMMON 0x38270, 0x38640, 0x381A0 | Entry insertion index, insertion and removal | Complete local instruction reading; original-image listings corroborated encoded removal. Reallocation/memory-movement callees still delimit full entry-management closure |
| LOD113C-D2COMMON 0x39F70 -> 0x389E0 | Propagation and direct resolved-entry addition | Original-image listings corroborate both full local instruction bodies. The recorded life-definition control bytes choose direct addition rather than 0x39CE0's formula branch; the callback-bit test is false for the recorded definition |
| LOD113C-D2COMMON 0x38550 | Changed-stat-key insertion | Original-image branch/table reading: stat 6 selects the immediate return at 0x38625 through table 0x38630. This excludes this particular key-collection operation, not every possible life notification elsewhere |
| LOD113C-D2COMMON 0x09252 -> LOD113C-FOG ordinal 10047, RVA 0x1CDA0 | Reallocation wrapper | Import and export resolved; byte-equivalent Fog wrapper read. Its allocation worker 0x1BDD0 and deeper heap paths remain incomplete |
| LOD113C-D2COMMON 0x39CE0, 0x39530 | Dependent-stat traversal and formula evaluation | Complete available instruction text read as a lead; encoded portions and the full transitive formula paths are not newly verified |

MRH-R08/MRH-V09 establish one reached actual existing-entry case: replacement versus aggregation. In the direct existing-entry branch, the propagated amount is computed from the old **own** entry. With own life 1000, resolved life 2560 and a replacing request of 13056, the local delta is 12056 and the direct resolved addition produces 14616. Writing the request directly into the resolved entry would give a different result. No allocation/removal or supplied writer executes in this nonzero existing-entry case. This does not retroactively change MRH-V08, whose writer remains explicitly supplied. MRH-Q01's full shared replacing-write boundary remains open for other applicable list conditions, entry-management, propagation and later notification cases.

The player front end requires separate bookkeeping, not a semantic return-value promise. Common 0x3A740 consumes the internal list operation's changed/unchanged result. For a changed player write it passes stat/layer to 0x38550; the stat-6 immediate-return branch excludes that key. Its listing leaves EAX=6 after extracting the stat identity, whereas the nonplayer suffix retains the list operation's 1. The healing caller consumes neither value. An initial research draft asserted those raw register values as writer results; self-review removed that unnecessary external return requirement before the fresh reader. MRH-V09(c) now asserts actual player changed-key exclusion and the healing callback's result 1. It does not turn an unconsumed register residue into a required stat-writer API result.

The compatible entry-management source records added for MTP-R20 were checked for the same original artifact and helper identities. They support reuse of their actual reading method, not transfer of that rule's stat-353/354 contract to life 6. MRH's captured life-definition controls and the fixed stat-6 changed-key branch establish its own selected path. Generic STL-R06c remains unresolved; no global stat authority is promoted by either specialized integration case. The new MRH-R08/MRH-V09 material requires an applicable fresh reader after self-review.

### Natural encounter preparation

For the 2026-10-30 original-profile encounter attempt, the original game's `-act3` launch option established an Act III starting location, level 75. No eligible healing invocation or exact callback/write capture was obtained. Town inactivity establishes neither healing behavior nor handler absence; natural encounter capture remains pending.

### Revival dependency follow-through

Game's predicate thunk 0xA688 resolves Common ordinal 10648/RVA 0x1A520, whose full plain-instruction body accepts only a present monster in mode 0 or 12. Thunk 0xAD4E resolves ordinal 10341/RVA 0x34020, forwarding the target and a loaded state mask into 0x33C20; the latter's complete original-image listing establishes mask intersection against the target's active state storage. The mask's producer/meaning and the remaining revival effects still require their owning contracts.

Game's collision-mask thunk 0xA682 resolves Common ordinal 11056/RVA 0x35640. Its byte-equivalent reconstruction clears the old footprint using the old mask, stores the new path mask and reapplies the footprint using that mask. The footprint-refresh thunk 0xA5DA resolves ordinal 10704/RVA 0x35CF0; its complete local body reaches 0x35800/0x35360/0x359D0 for removal, footprint selection and reapplication. The final pattern writer 0xAD2A resolves ordinal 10565/RVA 0x4BE90. These additional reads identify the actual collision dependency chain; they do not by themselves establish complete revival or generic collision parity. MRH-Q02 remains open for those effects, registration, event and mode work.

### Preliminary full peer reading of the storage extension — 2026-10-30

The fresh full peer reader examined the resulting contract and cited dependencies. It inherited the writer's model, so this is a preliminary peer pass rather than the required smaller-model blind-implementer pass. Counts were MISMATCH 0 under the intended array selection, GUESS 1, GAP 1, DISCRIM 1, COVERAGE 2 and HARDEST 3. The five existing MRH question gates and applicable SCX producer limits were retained. No gameplay or full-gap completion follows from this report.

| Finding | Outcome | Rationale / change |
|---|---|---|
| GUESS: MRH-V09 did not explicitly set usesResolvedStats | Spec edit | Set it true for every independent case. With it false, the reader could instead obtain maximum/current 10000/1000 and request 5100, followed by resolved 6660. The explicit true input selects the intended actual getter result 25600/2560 |
| GAP: section 5 omitted the internal changed/unchanged decision and player changed-key exclusion | Spec edit | Added both observable stage outputs to section 5; the player collection is an explicit input and MRH-V09(c) starts/ends with it empty |
| DISCRIM: equality compared with own versus resolved life was indistinguishable in MRH-V09(b) | Spec edit | The new case has own life 13056, resolved life 2560 and actual computed request 13056. It requires internal unchanged, both different arrays unchanged and callback 1. Its normal percentage draw still advances C. Comparing against resolved life now selects a wrong changed decision |
| COVERAGE: wrapped delta/resolved addition and resolved storage above maximum | Spec edit | MRH-V09(d) supplies explicit synthetic signed arrays. Request 100 minus own -2147483600 wraps to delta -2147483596; adding resolved -1000 wraps to 2147482700, which is not capped to maximum 1100. These are executable supplied-array inputs, not claims about natural encounter preparation |
| COVERAGE: expression seed changes before the caller's bounded draw | No change | MRH-R07 explicitly consumes the seed after both completed evaluations; MRH-V08's ordered log requires evaluation before the caller draw. Execution of random expression tokens and their draws remains owned by the cited SCX contracts. The literal-only cases do not claim to exercise every possible expression binding or its seed mutation |
| HARDEST: exact multiplication order, wrap and clamp in MRH-R03/MRH-V04/MRH-V06 | No change | The earlier discriminating cases and NRF-R05 contract remain unchanged; the new storage case does not replace that arithmetic |
| HARDEST: selected reads, expression order and ignored writer result in MRH-R07/MRH-V08 | Spec edit for the missing selection input in V09; otherwise no change | Actual read selection is now explicit in the new vector. MRH-V08 retains the independent supplied writer returning 0 and ordered calculation/read/request assertions |
| HARDEST: own versus resolved life and player bookkeeping in MRH-R08/MRH-V09 | Spec edit | Distinct own/resolved equality and explicit player collection/exclusion assertions address those choices; the raw unused unit-writer register value remains investigation metadata only |

Self-review recomputed every changed case after these edits: MRH-V09(a)/(c) request 13056 and resolved 14616; (b) the same request equals own 13056 and leaves resolved 2560 unchanged; (d) normalized percent 100 gives increase 1100 and request 100, then the two wrapped storage operations above. Mechanical lint returned exit 0, 0 errors and 0 warnings. The public trio is frozen for the formal smaller-model full blind pass; earlier full passes remain historical for their own contracts. All five full MRH gates remain open and no matching healing observation has been added.

### Formal smaller-model full blind pass — 2026-10-30

The fresh smaller reader recomputed every MRH-V01–MRH-V09 case from the resulting contract and cited dependencies. Counts were MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, COVERAGE 0 and HARDEST 3. All five existing MRH question gates remain explicit; no gameplay or complete storage/revival/blood-sucking claim was added.

| Finding | Outcome | Rationale |
|---|---|---|
| HARDEST: MRH-R01/MRH-R02/MRH-R06 data associations and selected-target rejection | No change | Data cells are distinct from actual dispatch execution, and the valid normal-entry target/registry cases and rejection order are explicit in MRH-V01/MRH-V07 |
| HARDEST: MRH-R03/MRH-R07 upper/lower evaluation, draw/read order, exact multiplication, wrap and clamp | No change | The reader recomputed the existing ordered request and arithmetic vectors without ambiguity or mismatch; their exact dependency and supplied-stage limits remain stated |
| HARDEST: MRH-R08 own-entry equality, delta propagation and player bookkeeping | No change | The reader confirmed MRH-V09(b)'s distinct own/resolved equality and MRH-V09(d)'s wrapped storage result; the actual existing-entry case retains the full shared integration gate for other applicable effects |

This is the final applicable formal reader pass for MRH-R08/MRH-V09. All blocking lists and advisory coverage are zero and every HARDEST entry has the outcome above. The blind-implementer requirement is complete for this resulting contract. The sustained investigation continues against the remaining five full gates; the document remains draft.

## Suck Blood recipient and reached life writes — 2026-10-30

This follow-through reuses the preceding original-artifact identities, MRH-F08–MRH-F11 selection/expression/getter/writer records and MRH-R08's actual reached storage case with their original limits. New original-image decoding of the complete owner helper corroborated every encoded instruction. No process invocation was observed and no full MRH question was closed.

| Record | Artifact / original RVA | Ordinal | Analyst description | Rules / method and limit |
|---|---|---|---|---|
| MRH-F12 | LOD113C-D2GAME 0xCF00 -> 0xDAA90 | none | Suck Blood action and first combat-record lookup | MRH-R09; complete plain-instruction bodies read. Queue starts at caster +0xAC, nodes compare attacker type/id at +4/+8 and defender type/id at +0xC/+0x10, advancing +0x84; the first exact match returns its damage at +0x14 |
| MRH-F13 | LOD113C-D2GAME 0x21260 -> 0xE03A0 | none | Canonical monster-owner lookup | MRH-R09; full 0x30-byte owner body corroborated against original-image disassembly. Monster data +0x14 -> context +0x28 -> game/id/type at +0x28/+0x2C/+0x30. Absent required context or failed registry lookup yields no owner. E03A0 consumes the compatible MRH-F08 normal registry contract; context production remains MRH-Q04 |
| MRH-F14 | LOD113C-D2GAME 0xDED30 | none | Subsequent queued combat consumer | Existing DPO-R13a/R23–R26 and companion's 2026-10-13/14 reading reused with original limits. Complete local body reread for caller correspondence; no generic reach/apply/event/reflection authority is promoted |

After the skill/target guards, 0xCF00 sets caster +0xC4 mask 0x40 and obtains the first matching pending damage record. An absent record returns 0 with that latch already set. A present record whose result +4 has bit 1 clear skips directly to queued consumption. On the hit branch, only monster caster type 1 calls 0x21260; its null result falls back to that caster. The nonmonster branch leaves the recipient absent and is not silently generalized into that fallback contract. MRH-R09 explicitly establishes the monster-caster path while MRH-Q03 retains the full caller boundary.

The action reads recipient current life through Game thunk 0xA42A (Common #10973/0x38B70), recipient maximum through 0xA3A0 (Common #10913/0x390F0), then selected-target current life through 0xA42A. It signed-selects the lesser of the queued damage's physical field +8 and target current life. It evaluates selected Skills `calc1` reference +0x138 through 0xAC9A (Common #10786/0x51BF0), with the caster, unchanged callback skill identity and level. The result is not clamped to 0..100. Game 0x14D0 receives that calculation result as its first multiplicand, the selected physical/life basis as its second and divisor 100. The wrapped sum with recipient current life is floored below raw 1, then signed-clamped to recipient maximum. Thunk 0xA472 invokes the same actual replacing writer already established by MRH-F11/MRH-R08. There is no target-life subtraction at this local stage and no mutation of the queued damage record.

Only after that write does the action call 0xDED30 with original game/caster/selected target. The same call occurs without restoration on a non-hit record. The action returns 1 after normal consumer completion and ignores its register result. Thus an eventual reach/apply rejection cannot be used to move or retroactively suppress the earlier local life operation. The subsequent consumer may perform its own damage, events, reflection, reaction or queue removal under DPO's existing gates; the locally restored value is not asserted to be the final value after those effects.

MRH-V10 executes canonical owner lookup, supplied-byte expression evaluation, actual RLI reads and the actual reached existing-entry writer. Only the subsequent generic combat consumer is supplied. Case (a)'s request 3810 produces own 3810/resolved 5370; fallback case (b) requests caster 4250 and produces own 4250/resolved 5250. Case (c) proves a resolved monster owner is not filtered into the caster. Case (d) uses first matching miss before a later matching hit, and (e) has no matching identity. The negative calculation in (f) requests raw 1 and produces resolved 1561; applying the healing percentage clamp would instead request 2560. In (g) maximum 3000 caps the request while the actual replacing delta produces resolved 4560 without another clamp. These are explicit synthetic inputs, not claims of natural encounter preparation or original compiler output.

Additional source follow-through read the separate start at 0xC760 (an entry inside the available 0xC6B0 region), its 0xA11B0 hit-result operation, physical fill 0xDCE00 and queued completion 0xDDDE0. Existing SLD-R14–SLD-R17 own corresponding physical/element fill paths where actually reached, and DPO owns queued scaling/enqueue/consume. Complete start/combat integration remains MRH-Q03. The Mosquito pair 0xB35A0/0xB3B20 was newly corroborated in full against the original image, with actual current-skill field accessors and seed initialization read through their resolved Common exports. Those readings settle encoded local instructions but do not by themselves establish full Mosquito combat/repeat effects.

Spatial follow-through located Common #11138/0x326C0 -> 0x325C0/0x7FCD0/0x314A0 -> 0x30730/0x4AD20. Full local source bodies were read; 0x325C0/0x326C0 were corroborated from the original image and existing compatible SFR/SBA evidence was checked. The remaining equipment/line traversal and generic dependency limits are retained rather than supplied as an actual successful spatial result. MTP's null-reference-point query does not establish this line predicate. None of those partial reads renews the existing DPO/SFR/SBA reach contract or closes MRH-Q03.

Self-review recomputed the new local cases and retained all five full question gates. A fresh smaller reader is required for MRH-R09/MRH-V10; the earlier formal pass remains applicable to its unchanged MRH-R08/MRH-V09 contract.

### Fresh smaller full reading of Suck Blood extension — 2026-10-30

The fresh smaller full reader reported MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, COVERAGE 2 and HARDEST 3. The seven referenced gates were MRH-Q01–MRH-Q05 plus the unresolved complete-family MRH-R04/MRH-R05; no gate or rule label changes follow.

| Finding | Outcome | Rationale / change |
|---|---|---|
| DISCRIM: using target current life alone passed every new case | Vector edit | MRH-V10(h) now supplies physical 1000 below target current 5000. The correct basis is 1000, increase 250, request 2810 and actual own/resolved 2810/4370. Target life alone instead requests 3810 and produces resolved 5370. The exact unchanged seed/list/queue metadata and operation order are explicit |
| COVERAGE: physical damage below the target-life cap | Vector edit | The same MRH-V10(h) exercises this previously absent minimum branch without changing MRH-R09 |
| COVERAGE: absent monster data, absent AI context or absent context game fallback | No change | MRH-F13's complete original-image body directs each absent prerequisite to the same no-owner result before the caller's shared fallback-to-caster branch. MRH-V10(b) exercises that fallback after failed canonical lookup; MRH-R09 explicitly lists all absent prerequisites. Their producers remain gated and no successful malformed-context result is inferred. Separate fixtures would repeat the same recipient/read/storage continuation |
| HARDEST: MRH-R09/MRH-V10 first match, owner and storage before consumer | Vector edit only for the minimum discriminator | First matching miss, canonical player/monster owners, failed-owner fallback and the consumer's observation of already changed arrays remain explicit in the prior cases |
| HARDEST: MRH-R08/MRH-V09 own versus resolved equality and wrapped arithmetic | No change | This extension consumes the unchanged formally reviewed reached storage contract; the new physical-basis case recomputes its own delta and resolved sum |
| HARDEST: MRH-R07/MRH-V08–MRH-V09 expression/draw/read/write order | No change | The healing callback and its ordered vectors are unchanged; MRH-R09 preserves its distinct recipient-current, recipient-maximum, target-current, calculation and replacement sequence |

Self-review recomputed MRH-V10(h): 25*1000/100 = 250; 2560+250 = 2810; own delta 2810-1000 = 1810; resolved 2560+1810 = 4370. No draw or clamp changes that case. The resulting vector delta is frozen for fresh smaller review; the full reader found no rule mismatch, missing input or missing output.

### Final applicable Suck Blood delta reading — 2026-10-30

The fresh smaller delta reader reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0. It independently recomputed MRH-V10(h)'s basis 1000, increase 250, request 2810, own delta 1810 and resolved 4370; the target-life-only alternative produces 3810/5370 and is rejected. The seven referenced gates remain the five full questions and unresolved MRH-R04/MRH-R05.

The one COVERAGE note repeated absent-context fallback and retains the full pass's explicit no-change rationale: every absent prerequisite returns no owner before the shared caster fallback already exercised in MRH-V10(b). The three HARDEST entries repeat the prior first-record/recipient/storage-before-consumer, own-versus-resolved/wrap and ordered expression/draw/read/write concerns, each with the recorded no-change rationale except the now resolved minimum discriminator. No contract edit follows from this clean delta.

The full pass plus this final applicable delta completes the blind-implementer stopping rule for MRH-R09/MRH-V10. The prior MRH-R08/MRH-V09 formal pass remains applicable to its unchanged contract. All full gates remain open and no runtime confirmation is added.

## Mosquito complete start preparation — 2026-10-30

MRH-R10 consumes the original start-table slot 55 -> 0xB35A0 recorded by MRH-F04. Its complete 0x100-byte region was read and independently corroborated against the original image on 2026-10-30; the local encoded LEA and all call targets agree. This establishes the complete start callback, including its actual node writes. The separate action at 0xB3B20 retains MRH-Q03.

| Record | Artifact / original RVA | Export ordinal | Supported correspondence / boundary |
|---|---|---|---|
| MRH-F15 | LOD113C-D2GAME 0xB35A0 | none | MRH-R10 complete start: lookup/current-node guard, normalization, temporary seed, two calculations, draw and actual stores |
| MRH-F16 | Game thunk 0xAABA -> LOD113C-D2COMMON 0x303A0 | 10511 | Byte-equivalent reconstruction read: present unit's skill container at +0xA8, current node +0x10; absent container yields no node |
| MRH-F17 | Game 0xA616/0xA628 -> LOD113C-D2COMMON 0x36750/0x36740 | 10936/10614 | Byte-equivalent reconstruction read: temporary seed first initializes {1,666}, then immediately replaces its low word with the wrapped caster/target id sum while keeping high 666; compatible NRF-R03a supplies semantic initialization |
| MRH-F18 | Game 0xAC94/0xACDC/0xAD0C -> LOD113C-D2COMMON 0x4DF60/0x4DF20/0x4DEE0 | 10243/10554/10748 | Complete byte-equivalent reconstruction reads: actual stores to retained node +0x18/+0x1C/+0x20, respectively count, target id and target type in this callback; unrelated node fields remain untouched |

The start validates its skill before calling the current-node getter; a missing current node returns before target normalization. With a node, it calls the same 0xE07A0 normalization and 0xA598 selection accessor used by MRH-R06, rejects absent/self selection, initializes the stack-local seed, then evaluates record `calc1` +0x138 and `calc2` +0x13C in that order via 0xAC9A/Common #10786/0x51BF0. The second call uses the original callback skill identity saved on the stack, even though the first result replaces EBX. Both use caster and original callback level. The original Game 0x11D0 bounded draw receives the temporary seed in ECX and the wrapped upper-minus-lower range in ESI, not the caster's persistent seed. Its result adds the first calculation result, then signed values below 1 are floored to 1.

The three nonempty-node setters execute actual stores in count/id/type order, followed by result 1. The id and type are read from the retained target. The start contains no life write, damage call, event schedule or do-handled latch store. Its count is later read by the action, but this does not establish an exact attack cadence or complete action execution. Binary getter/setter labels suggesting item or inventory behavior were not used as semantic evidence; their bodies establish these slot operations.

MRH-V11(a) uses id sum 10 and temporary high 666. NRF-R01 gives low 734112332/high 4, whose modulus 3 is 2; literal endpoints 4 and 7 therefore store count 6 and player target id/type 9/0. Persistent caster/target seeds remain unchanged because these synthetic expressions draw nothing. Range 1 in (b) steps the temporary seed despite returning 0; equality in (c) does not step. Signed byte literals in (d) produce -5/-2, draw 2 and floor -3 to 1. Reversed 7/4 endpoints in (e) do not sort and consume no draw. Cases (f)/(g) distinguish current-node rejection before normalization from stale selection clearing with a present node. Temporary seed values are executable intermediate outputs, not an additional persistent game-state requirement.

The remaining revival work was also followed through the complete local event-reset operation 0xAF920, counter operation 0xCA490, mode selection/build/dispatch 0xF3CB0/0x75070/0x75E10, callback selector 0x749A0 and its reached local preparation helpers. Existing MTP event/collision source records and STL scheduler records were checked for exact correspondence. The full monster-mode dispatcher reaches further handlers and event, AI, animation and update operations; those unread interiors and the broader revival eligibility/final-state boundary remain MRH-Q02. No new revival behavior is promoted by locating that path.

All five full question gates remain open. MRH-R09/MRH-V10's completed full/delta review remains applicable to that unchanged contract; the new MRH-R10/MRH-V11 require a fresh smaller reader after self-review.

### Final applicable Mosquito-start full reading — 2026-10-30

The fresh smaller full reader reported MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, COVERAGE 3 and HARDEST 3. GATED 7 comprises MRH-Q01–MRH-Q05 and the unresolved complete-family MRH-R04/MRH-R05. No contract or gate change follows.

| Finding | Outcome | Rationale |
|---|---|---|
| COVERAGE: MRH-R09 absent-context owner fallback | No change | This repeats the prior full/delta advisory, whose complete MRH-F13 absent-prerequisite branches and shared fallback continuation were explicitly triaged above. MRH-V10(b) executes the same no-owner-to-caster continuation; no malformed-context successful result is assumed |
| COVERAGE: MRH-R10 invalid skill or absent selected record | No change | MRH-R10 explicitly requires the signed lookup bounds and present record before current-node access, normalization, calculations or writes. MRH-F15's complete start entry establishes the same rejection suffix with result 0. MRH-V01 exercises the common invalid-id guard semantics in its stated callers, while MRH-V11 exercises this start's later distinct rejection order. Those earlier vectors are not relabeled as Mosquito execution; another guard-only fixture would add no newly ambiguous decision or effect |
| COVERAGE: MRH-R10 id-sum, range-subtraction and count-addition wrap | No change | The three widths are explicit: low32 id sum, signed i32 low32 upper-minus-lower, and signed i32 low32 lower-plus-draw before the floor. NRF owns seed initialization and signed-positive range interpretation. MRH-F15's complete original instructions establish these widths; the temporary seed, reversed interval, no-step and floor cases already distinguish their ordering. MRH-V11 does not claim every wrapping operand combination was exercised, and no widening or saturation is allowed by the current rule |
| HARDEST: MRH-R10/MRH-V11 temporary seed, ordered evaluations, guard order and actual node writes | No change | The reader found no mismatch or missing value in all seven cases. Temporary and persistent seeds, range-1 stepping, zero/reversed intervals, negative-count floor and current-node-before-normalization rejection remain explicit |
| HARDEST: MRH-R09/MRH-V10 first match, canonical recipient and storage before consumer | No change | The completed Suck Blood full/delta pass remains applicable to this unchanged contract, including the below-target-life physical discriminator and explicit supplied consumer boundary |
| HARDEST: MRH-R08/MRH-V09 own/resolved storage and wrap | No change | Its completed formal pass and distinct own-entry equality/wrapped storage cases remain applicable; neither the subsequent Suck Blood extension nor Mosquito start modifies that rule |

This clean fresh full pass completes the applicable blind-implementer stopping rule for MRH-R10/MRH-V11. Each advisory and HARDEST outcome is recorded above; no behavioral edit requires a new delta pass. Earlier full/delta reviews remain applicable to their unchanged contracts. The trio remains draft with all five full question gates open and no runtime-confirmed gameplay case.

## Mosquito local action through actual life and repeat writes — 2026-10-30

MRH-R11 consumes action-table slot 107 -> Game 0xB3B20, its complete original-image corroboration preserved by MRH-F07 and the prior current-node/seed/expression/getter/writer records. This follows the complete skill-local action to its output stores while preserving separate spatial/apply/reaction boundaries. No generic producer is silently replaced with a verified successful result.

| Record | Artifact / original RVA | Ordinal | Supported correspondence / limit |
|---|---|---|---|
| MRH-F19 | LOD113C-D2GAME 0xB3B20 | none | MRH-R11 whole local action; original-image decoding corroborates entry, clear, local draws, life setter and final repeat writes |
| MRH-F20 | Game 0xAD12/0xACB2/0xAD3C -> LOD113C-D2COMMON 0x4E030/0x4E000/0x4DFD0 | 10980/10649/10179 | Complete byte-equivalent reconstructions read 2026-10-30: retained current-node count/id/type at +0x18/+0x1C/+0x20 |
| MRH-F21 | Game 0xACE8 -> LOD113C-D2COMMON 0x2F0C0 | 10542 | Complete byte-equivalent reconstruction read 2026-10-30; preserves Unit +0x38 low byte, stores argument low 24 bits in bits 8–31. Consumes compatible SCD sequence-event correspondence; no subsequent frame execution established |
| MRH-F22 | Game 0xA796/0xA766 -> LOD113C-D2COMMON 0x52100/0x51FF0; Game 0x9F1A0/0x9F210; Game 0xACE2 -> Common 0x4E900 | 10567/10297; internal fills; 10510 | Compatible SLD endpoint/fill/duration contracts and their original 2026-10-12/15 readings reused; caller argument/order and three extra draws established by MRH-F19 |
| MRH-F23 | Game 0xA754 -> LOD113C-D2COMMON 0x326C0; Game 0xDE0C0/0xDCF20 | 11138; internal consumers | Reach uses extra allowance 0. DPO scale-true apply and HRC reaction remain their exact shared dependencies/gates; MRH-V12 supplies these three boundaries explicitly |

The action obtains the current node before reading its saved id/type and invoking Game 0xE03A0 on the original game. It does not call 0xE07A0 or read the path's selected unit. The supplied reach request is (caster, resolved node target, 0). Only a nonzero reach result proceeds to snapshot the node +0x18 count.

Damage clears 0x1C dwords at the local record base. Before calling either fill, source-included physical minimum/maximum are requested with condition 1 and each arithmetic-shifted right by 8. The fills then run in physical/elemental order. Compatible SLD-R14 requests source inclusion 0 and SLD-R15 requests its own duration/min/max. The three additional Game 0x11D0 draws use caster +0x20 and the same whole-maximum-minus-whole-minimum range.

The original stack displacements distinguish the writes: at 0xB3C44, after pushing the first duration argument, the store at ESP+0x54 addresses damage +0x28 (poison magnitude), and receives twice the first whole-unit range outcome. After duration #10510 returns and restores its arguments, 0xB3C59's ESP+0x54 store addresses damage +0x2C (poison length). The next two draws write shifted outcomes to damage +0x3C/+0x40 (mana/stamina drain). At 0xB3C8E another pushed argument means ESP+0x30 addresses damage +4: the result word is set to 1, not the record flags. The whole record's semantic mapping consumes the existing DPO companion. These displacements were checked against the full original-image listing, independently of analyst function names.

Skill `calc3` +0x140 is evaluated in caster/id/level context. Caster current life is read, and Game 0x14D0 receives calculation, current record physical and 100 in that order. The current-life sum wraps, then the maximum-life wrapper is called and its signed comparison selects the upper cap. No raw lower-floor instruction exists in this action. The actual replacing writer precedes Game 0xDE0C0, which receives scale request 1. After apply returns, an actual target-life read exactly equal to 0 sets result bit 2, then Game 0xDCF20 reacts. Neither consumer is newly verified by its placement; MRH-V12 supplies them and asserts their boundary explicitly.

The retained count is the value read before preparation/apply/reaction. Its decremented positive branch writes the retained node through Common #10243/0x4DF60, reads current Skills `Param1` +0x148 and stores its low-24-bit sequence-event value through #10542, then reads current Unit +0x48 animation total, ANDs 0xFFFFFF00 and adds 0x100. Compatible HRC's original semantic map identifies +0x48 as total, separately from accumulator +0x44; no GH driver or animation cadence is inferred. Source field correspondence for `Param1` consumes GDA/SCX's Skills descriptor mapping, not a new named data-cell claim. Count 1 does not store zero. A consumer changing the unit's current node does not redirect the saved-node store, provided the retained node remains valid. Sequence low byte and total use their post-reaction values.

MRH-V12's actual endpoint/fill inputs avoid the source-equipment producer (`SrcDam` 0) and Kick override. Four local caster steps yield draws 351/0/2/2, physical 1375, poison magnitude 8/length 12 and mana/stamina 1536/1536. Literal `calc3` 25 yields increase 343 and request 2903; actual own delta 1903 produces resolved 4463 before apply. Cases separately establish count-1 no-repeat, maximum cap, no lower floor, saved-node/count versus post-reaction controls, failed reach before preparation and the post-apply zero-life lethal OR. All bytes and consumer mutations are explicit synthetic inputs, not original compiler or natural AI claims.

Self-review recomputed every arithmetic and storage branch and checked the original register/stack argument correspondence, field widths, draw ownership/order and consumer-before-repeat sequence. All five full question gates remain open. MRH-R08/MRH-V09, MRH-R09/MRH-V10 and MRH-R10/MRH-V11 retain their completed applicable reader passes; this new MRH-R11/MRH-V12 requires a fresh smaller reader.

A further self-review discriminator MRH-V12(h) uses caster seed {2,666}. The four seed pairs become {3582796836,0}, {2512981940,1494357220}, {1658350184,1048145591} and {1733152959,691685207}; draws are 36/2/2/0. This gives physical 1060, poison 12/length 12, mana 1536 and stamina 1024. The target's raw life is only 1000, but Mosquito does not cap this physical basis to target current life before its caster restoration: increase 265, request 2825, own delta 1825 and resolved 4385. This rejects both a Suck Blood-style target-life cap and exchanging the last two draw assignments, which coincided numerically in the default seed case. No generic combat effect is newly asserted by that fixture.


### Final applicable Mosquito-action full reading — 2026-10-31

The fresh smaller full reader independently recomputed MRH-V01–MRH-V12, including all eight MRH-V12 cases. It reported MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0 and COVERAGE 0. GATED 7 remains MRH-Q01–MRH-Q05 and unresolved complete-family MRH-R04/MRH-R05.

| Finding | Outcome | Rationale |
|---|---|---|
| HARDEST: MRH-V09 actual own/resolved storage | No change | The independently recomputed own-entry equality, wrapped delta and resolved accumulation agree; existing-list preconditions and full storage gate remain explicit |
| HARDEST: MRH-V10 recipient and physical restoration basis | No change | First-match selection, recipient fallback and actual life write before the supplied consumer remain consistent; the below-target-life physical case discriminates the restoration basis |
| HARDEST: MRH-V12 retained node and post-reaction controls | No change | Case (e) uses the saved node/count while reading sequence low byte and animation total after reaction. Case (h) distinguishes the physical basis and changed draw assignments. All eight cases recompute without missing inputs or incorrect outputs |

This final applicable full pass completes the blind-implementer stopping rule for MRH-R11/MRH-V12. The preceding self-review statement records the initial handoff, not a current pending review. No contract, vector, evidence label or gate changes follow; natural gameplay confirmation remains absent.

### Natural Fallen Shaman revival — 2026-10-31

MRH-O02 records a partial natural encounter on an owned process reporting the vanilla LoD Expansion 1.13c profile. Ordinary travel reached Cold Plains, level 3. A class-58 Fallen Shaman C had a registered minion chain whose identifiers, in traversal order, were 75, 74, 73, 72. Ordinary melee killed class-19 minion T, identifier 73, leaving monster mode 12. T's stored owner key resolved C. Bounded entry/return observations accompanied ordinary retreat; no gameplay call or memory write supplied the outcome.

The first decision window recorded one D2Game `0x890A0` AI invocation and four `0xB12B0` candidate callbacks. The callback for T changed the selection from absent to T, count 0 to 1 and selected distance to 20; the other callbacks left that selection unchanged. The AI entered with seed {2447118354,788880129} and left with {706359088,1100627028}. The four original RNG steps produce bounded values 87, 63, 33 and a final low byte 48. The second chance value 63 fails the captured Normal-difficulty threshold 45, so this window has no resurrection action/helper call. This is observed selection and a declined decision, not an unsuccessful revival attempt.

The next two-second retreat window recorded one AI invocation, four candidate callbacks, one action call at `0xB5BD0` from dispatcher `0xA1B36`, and one shared revival call at `0xB33B0` from `0xB5CD0`. The action received skill 181/level 1, C as caster and cached selected target T; it returned 1. The helper returned 1. C's AI seed stepped from {706359088,1100627028} to {215003117,1353485988}, producing chance values 48 and 17. The latter passes 45. This establishes this actual AI-to-action occurrence; it does not establish other families' entry or a general cadence.

The attributed D2Common replacing-writer entry `0x3A740`, caller Game `0xB33D9`, received T, stat 6, raw 768, layer 0. Its reached inner entry `0x3A280` had T's extended base list. Both own and resolved arrays initially lacked stat 6 while retaining stat 7=768. The own array had count/capacity 11/12, and the resolved array 14/16. Captured before-images and return deltas show insertion of (stat 6, layer 0, value 768) in both arrays, with counts 12 and 15 and unchanged backing pointers. This is an actual prepared-capacity insertion, not MRH-R08's existing-entry vector. It supplies evidence for the lifecycle owner's insertion boundary without establishing allocator failure, generic propagation or notifications.

The successful window held respectively 1/4/1/1/178/235 records for AI, candidate, action, helper, unit writer and inner writer, all with dropped count 0. Every record reports the same thread identity and depth 0. Sequence numbers are local to each entry, and no common timestamps or global call order were recorded. AI-to-action attribution therefore uses the source path and matching identities, not a nested recorded stack or cross-entry timeline. No calls were in flight at the end of the window. The helper's register-held target was not directly recorded at its entry; attribution uses the action's cached target, its reached helper call, the helper-specific writer caller and the identical registered target/list identity. Later paused reads found T in mode 8 with flags 0x0402000E and own/resolved life and maximum life 768. Those later reads are not immediate helper-return snapshots: the resolved count had subsequently become 12, with the cause of that later change unobserved, so they do not replace the inner writer's count-15 return state.

Static follow-through checked available byte-equivalent reconstructions before disassembly. Whole Game `0x890A0`, `0xB12B0`, `0xB1320`, `0xB17B0`, `0x21490`, `0x21260`, `0xF2120`, `0xF3E30`, `0xB5BD0` and `0xB33B0` were read; original-image decoding corroborated selected encoded search/registry instructions. The original Game fingerprint matches `LOD113C-D2GAME` in the manifest. Common `0x33D60` resolves through import ordinal 10541 to mask intersection `0x33C20`; the latter's encoded instruction was independently decoded. Compatible SFR-R10 distance and existing registry/skill-node contracts retain their prior evidence limits. These readings do not yet provide a complete revival rule/vector: optional quest attachment, complete mode/animation/damage/modifier paths and the remaining lifecycle cases still require their owned contracts.

Runtime loaded-DLL hashes, complete override inventory, effective-file hashes and physical supplying archives were not captured. Narrow prepared MonStats/Skills fields and the installed callback pointer were read, but installation is distinct from the actual invocation above. This observation promotes no rule label or document status and closes none of MRH-Q01–MRH-Q05 in full. In particular, it is not a healing observation under MRH-O01 and does not runtime-confirm MRH-R08 or any existing vector.

### Prepared life insertion consumer — 2026-11-01

The reviewed lifecycle handoff supplies STL-R06d/STL-R06e's actual direct propagation and sorted replacement/insertion with prepared storage. MRH-R08 now consumes those named contracts for stat 6 through the same unit wrapper used by the existing-entry case. This extends the consumer's admitted inputs to independent absence of the own and resolved life entries; it does not establish a new generic stat writer. The wrapper, Common ordinal 10887/RVA `0x3A740`, calls the list writer `0x3A280` and requests further changed-key bookkeeping only for a changed player-unit write. The list writer independently checks the base's owner kind. Both collection paths exclude stat 6 through `0x38550`, irrespective of the stat definition's collection bit. No numeric unit-wrapper return contract is added.

The source controls were checked before comparing the captured output. The dated 2026-10-30 named `hitpoints`/ID 6 text record and compiled-control record above retain their original profile and hash limits. A read-only 2026-11-01 check of the still-loaded original-profile process followed the imported data-table context afresh: ItemStatCost count 359, stride 0x144, record identity 6, control bytes +4/+5 both 0x10 and bytes +0x50–+0x53 all zero. This agrees exactly with the earlier narrow control record. It is later corroboration of the loaded definition, not a snapshot taken at the 2026-10-31 writer entry.

| Required semantic input | Code/contract correspondence | Captured insertion input / consequence |
|---|---|---|
| Extended base with state storage, nonsecondary | List flags +0x10: 0x80000000 set, 0x2000 clear; STL-R06d starts at the source itself | Full flags 0x80000000; direct addition starts at T's own base |
| No parent | List +0x34 | Zero; there is no second ancestor to update |
| Owner and target identity | List owner kind/id +0x08/+0x0C; unit base +0x5C; resolved-owner pointer +0x44 | Monster kind 1/id 73 and the same target/base identities; player-only wrapper suffix is skipped |
| Direct definition | Record bytes +0x51/+0x52 zero | No dependent-expansion/formula operation |
| No resolved callback or marker | Byte +5 mask 8 clear; byte +0x53 zero | No resolved callback/positive-marker operation, even though a callback pointer is present |
| No filtering/zero retention | Byte +4 mask 4 clear; byte +0x50 zero | Direct nonzero addition; no zero-result removal reached |
| Changed-key control | Byte +5 mask 16 set, but stat 6 is fixed-code excluded | No collected key; the captured owner is also nonplayer |
| Valid active arrays and prepared capacity | STL-R06e insertion and STL-R06d addition; original own +0x24/count +0x28/capacity +0x2A, resolved +0x48/count +0x4C/capacity +0x4E | Own 11/12 and resolved 14/16; both lack life, so one insertion fits each existing store |

The existing full readings of Common `0x3A740`, `0x3A280` and `0x39F70` were rechecked after checking byte-equivalent C availability; none has such a reconstruction. Their compatible original-image decoding remains the recorded evidence. STL owns the complete prepared insertion/shift and direct addition callees, including `0x382B0`, `0x38270`, `0x38640`, `0x389E0` and the shared memory-movement operation. No source list is detached or freed, no array resize is requested, and no dependent formula or indirect callback is admitted by this case. Lists attached below the base do not enter this parent-directed operation. Other list conditions, missing capacity, resize/failure and notification-producing definitions remain their original shared gates.

MRH-V13(a) independently derives two inserted values 2112. Case (b) instead derives old resolved 320 + own delta 2112 = 2432 while increasing only the own count. Case (c) derives delta 2112-500 = 1612 and increases only the resolved count. These synthetic cases discriminate independent entry absence and the own-versus-resolved delta basis before the capture replay is considered.

MRH-V13(d) supplies the actual recorded initial active arrays, count/capacity pairs and request 768. STL-R06e gives old own 0, delta 768 and a new own life entry. STL-R06d independently gives absent resolved 0 + 768 and a new resolved life entry. Source-derived expected counts are 12/12 and 15/16. Comparison with MRH-O02's unit-writer record 166 and inner-writer record 217 matches those complete active arrays, unchanged preexisting entries, capacities/backing pointers, remaining captured list metadata, target unit metadata/seed and inner changed result 1. It does not establish cross-entry temporal ordering, callback absence from a callback trace, later mode processing, delivery or complete revival. The source exclusion of callbacks is distinct from observing their absence.

### Final applicable prepared-insertion full reading — 2026-11-01

The fresh smaller full reader independently recomputed MRH-V01–MRH-V13 with the cited dependency contracts. It reported MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0 and COVERAGE 0. GATED 7 remains MRH-Q01–MRH-Q05 and unresolved complete-family MRH-R04/MRH-R05.

| Finding | Outcome | Rationale |
|---|---|---|
| HARDEST: MRH-V09 signed wrapping and own/resolved updates | No change | The own-entry comparison, wrapped delta and resolved accumulation independently recompute under the stated existing-entry inputs |
| HARDEST: MRH-V12 draws, actual storage and post-consumer writes | No change | Its eight cases retain the distinct draw assignments, actual life write before supplied combat consumers, retained node/count and post-reaction control reads |
| HARDEST: MRH-V13 independent absence and prepared insertion | No change | Cases (a)–(d) independently produce 2112/2112, 2112/2432, resolved delta 1612, and life 768 with count/capacity pairs 12/12 and 15/16. Sorted insertion and independent count changes require no missing input or guessed result |

This completes the applicable static blind-implementer stopping rule without a contract, vector, acceptance, gate or label edit. Review establishes consistency of the supplied-input contract; it does not establish that all of a vector's inputs and assertions were captured in gameplay.

An independent field-by-field audit supports MRH-O02's narrow storage observation and finds no numeric mismatch. Whole MRH-V13(d) remains `runtime_observed: no`: its inherited synthetic callback and empty callback log were not captured, a nonzero primary-list head does not establish the fixture's singleton child relation or valid unchanged child contents, absence of concurrent mutation was not measured, and unrelated metadata was recorded only within a 256-byte unit window and 128-byte base-list window. Unchanged captured pointers and bytes establish only those captured values. Later ItemStatCost corroboration is not an entry-time definition snapshot. Static exclusions of callbacks, child traversal, allocation, resize, deletion and RNG are source evidence rather than observations of those calls' absence. The fixture and its assertions remain intact; no assertion is removed or reinterpreted to obtain a whole-vector match. All five full MRH question gates, static rule labels and draft status remain unchanged.

### Revival collision and selected-mode follow-through — 2026-11-01

MRH-Q02's action-side query is Game import `0xAC6A` -> Common ordinal 10851/RVA `0x4CE90`, distinct from the revival helper's final collision setter at Common `0x4BE90`. After checking byte-equivalent reconstruction availability, the complete query and dispatch callees `0x4C570`, `0x4C140`, `0x4B360`, `0x4B510`, `0x4AA90` and `0x4AC40` were read. The room-coordinate lookup `0x01330` has a byte-equivalent reconstruction, which was read first. Original-image decoding corroborated the encoded point, edge and rectangle scans. The immutable dispatch data selects the single-point query for pattern 0, the cross query for patterns 1/3/5 and the inclusive 3-by-3 rectangle query for patterns 2/4; larger unsigned patterns return 1. These are source findings within the still-gated action, not a new successful eligibility vector or a complete collision contract.

The selected revival mode requires the linked extended monster record. Game `0x01320` bounds the monster index against MonStats count, reads its signed MonStatsEx word at `+0x18`, bounds that linked identity against MonStats2 count, and selects stride `0x134`. Game `0xF3CB0` reads `ResurrectMode` at extended record `+0x10B`, not at the same offset in MonStats. Common `0x56C10`'s named descriptors bind `ResurrectMode` and `ResurrectSkill` to `+0x10B/+0x10C`. Original-image decoding corroborates both the linked lookup and selector. The latter treats a zero skill word as absent; otherwise its zero-extended identity must be below the signed Skills count before selection through Common ordinal 10630/`0x4F070` -> `0x4E480` and storage through ordinal 11103/`0x303E0`. Thus a prepared word 0xFFFF is excluded by the count guard, rather than by a signed-negative test. The small store has a byte-equivalent reconstruction, which was read; the complete lookup and its wrapper were also read. Complete preparation/producer effects remain their existing shared gates.

A read-only check of the owned original-profile process on 2026-11-01 found MonStats count 734, MonStats2 count 609 and class 19's signed linked identity 19. The linked extended record had mode byte 8 and skill word 0xFFFF. A Storm-loader read of effective `data/global/excel/monstats2.txt`, range 0–32768, returned 32768 of 138542 bytes: the complete 126-column header and complete `Id=fallen1` row at zero-based ordinary row 19 were available. Header-name selection gives `ResurrectMode=S1` and empty `ResurrectSkill`. This is sufficient coverage for those named cells, not complete-file coverage. GDF-R06a identifies mode 8 as S1. The text cells, compiled values and code-defined consumers are separate evidence; no generic missing-column conversion or complete parser/linker is verified. Effective-file hash, supplying physical archive and loaded-DLL hashes were not captured.

For these later prepared inputs, the selector's return is 8. The earlier neutral-mode source exploration remains conditional work for mode 1 and does not establish the actual Fallen record's selected mode. Original immutable Game mode-8 rows select callbacks `0x75510`, `0x75CE0`, `0x76BF0` and preparation controls 0/0/1. Their complete bodies, Game `0x77240`'s damage-stat preparation and `0x268E0`'s modifier dispatch were read after the reconstruction check; reached handler, animation, ownership, allocation, notification and indirect callback interiors remain incomplete. These new reads do not establish complete output effects merely by resolving a table slot.

The 2026-11-01 correspondence is later than MRH-O02's 2026-10-31 capture, and ordinary Save and Exit/rejoin has since created a new game. It is not an entry-time record or proof of which operation produced the earlier mode 8. No revival rule/vector, evidence label, acceptance boundary or full question gate changes. MRH-Q02 still requires the actual selected mode's complete lifecycle/animation outputs and successful/failing integration vectors.

### Selected monster animation investigation — 2026-11-01

The animation lookup reached by Common `0x32670` was followed through `0x41F70` -> `0x43C30`/`0x43860` -> `0x41DD0`. Compatible whole original-image readings from 2026-10-31 retain that date; the byte-equivalent reconstructions of extended-monster lookup `0x013B0`, weapon-code mapping `0x437F0` and eight-character comparator `0x41D00` were checked and read on 2026-11-01. The original formatting import at Common `0x89248` resolves to USER32 `wsprintfA`, with the ordinary format `%s%s%s`. This resolves the formatting boundary's identity, rather than treating a function label as its implementation.

On the supplied untransformed monster-kind/class-19/mode-8 path, the tag builder uses MonStats token, MonMode token and the linked MonStats2 weapon code. The later loaded records contain `FA`, `S1` and `hth`; the only applicable loaded monster token-override entry selects mode 13, not mode 8. The lookup uppercases the concatenation to `FAS1HTH`, sums its unsigned character bytes and selects bucket 239. A read-only chase of that later loaded bucket found 29 entries of stride `0xA0`; two contiguous reads covering all 4640 entry bytes found the exact tag at entry 13, with frame count 24 and rate 256. This is record correspondence, not an observed execution of the lookup or animation initialization. No binary animation data or raw capture is published here.

The original classification rows also distinguish monster class ranges. For mode 8, the five dwords at Common `0x8C828` are 1/1/1/0/0 for classes below 410; the row at `0x8C968` is 1/0/1/0/0 for classes at or above 410. Common `0x30C70`'s direct movement field therefore admits the Fallen path, before the attack predicate. The later-class row instead continues to the skill-dependent attack test. Sharing a Game mode callback does not establish identical Common animation branches, and a missing optional resurrection-skill selection does not prove the current skill remains absent through intervening mode operations.

The supplied untransformed Fallen path consumes MMF-R02/MMF-R04a/MMF-R03's movement percentage, ordinary prepared animation base and path-speed operations. It does not use the selected skill-animation record's rate as that movement base. The later prepared MonStats values at `+0x32/+0x34/+0x36/+0x38` are 5/5/256/128. For example, explicitly supplied stat 67=75 and diminishing result D(4)=0 give rate 192 and path speed 960 under those existing shared numeric contracts; these are conditional source calculations, not observed setter results. In Common `0x335F0`, the selected monster initialization path also clears the animation condition and sequence, sets its accumulator from the monster-kind result of `0x32220`, and derives the animation total from the selected frame count. Reaching it, supplying the untransformed inputs and establishing the complete final mode effects remain MRH-Q02.

A fresh whole reading of Common `0x5A8A0`, after checking reconstruction availability, follows the ordinary/run prepared-value producer to `0x43C30` and `0x41DD0`. It gives a targeted remaining shared-evidence check under GDA-R08b/GDA-Q05 and MMF-R04b; it does not close those contracts here. Actual transformation selection, initialization-time production, the complete mode wrapper, state/list lifecycle, damage/modifier callbacks and subsequent animation execution retain their existing owners and gates. No rule, vector or runtime label is promoted by these supplied-input and later-record correspondences.
