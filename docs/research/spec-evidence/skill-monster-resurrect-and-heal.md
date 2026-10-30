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
| MRH-F05 | LOD113C-D2GAME | 0xCF00, 0xC760 | none | Suck Blood action/start | MRH-R02/MRH-R05 | Local bodies read; actual combat, owner and storage chain incomplete |
| MRH-F06 | LOD113C-D2GAME | 0xB33B0 | none | Shared revival operation | MRH-R04 | Complete local instruction body followed; stat/path/registration/room imports and local mode helpers incomplete |
| MRH-F07 | LOD113C-D2GAME | 0xB35A0, 0xB3B20 | none | Mosquito start/action | MRH-R05 | Instruction reading; encoded spans and downstream combat calls need further corroboration; not promoted |

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
