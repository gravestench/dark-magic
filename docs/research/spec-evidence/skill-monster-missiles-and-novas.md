# Evidence: monster missiles and novas

Contract: [monster missiles and novas](../../specs/skill-monster-missiles-and-novas.md). Open work: [gaps](../../plans/gaps/skill-monster-missiles-and-novas.md).

## Identity and reuse

New readings dated 2026-10-29 use original LoD Expansion 1.13c Windows x86 `LOD113C-D2GAME`, SHA-256 `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`, 1138688 bytes, matched afresh against the [manifest](../source-manifest.md). The byte-matching index contains none of the eight principal callback entries. Their instructions were read through the available original-code disassembly. Microsoft original-image disassembly corroborated the complete nova callback's byte-encoded spans; this is not a claim that other callbacks' byte-encoded spans were decoded independently.

Compatible shared dispatch, missile request/layout and calculation evidence is reused through [skill dispatch](skill-cast-start-and-do-dispatch.md), [missile creation](missile-creation-and-seeding.md), [missile initialization](missile-initialization-and-registry.md), [skill calculations](skill-calc-expressions.md) and [numeric foundations](core-numeric-and-rng-foundations.md). Their original dates and dependency limits remain in force. No earlier monster-projectile-specific observation or companion was found. Private source forms/paths, matching records, tool history and raw data are preserved separately outside the repository.

## Dispatch and local function records

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | 0xA19A0; action table 0x1074A8 | none established | Server action selection | all | Reused dispatch contract; new table bytes confirm associations below |
| LOD113C-D2GAME | 0xB56C0 | none established | Targeted spray callback (slot 88) | MMS-R04 | Local body read; direction/node imports and helpers incomplete |
| LOD113C-D2GAME | 0xB52A0 | none established | Nova callback (slot 99) | MMS-R02/MMS-R03 | Entire body read; original-image instructions 0xB52A0 to 0xB5442 corroborated, request emission verified with supplied dependencies |
| LOD113C-D2GAME | 0xB6A60 | none established | Sustained breath callback (slot 95) | MMS-R05 | Local body read; byte-encoded spans/dependency effects remain unresolved |
| LOD113C-D2GAME | 0xB4DE0 | none established | Spread turret callback (slot 105) | MMS-R06 | Body read; arithmetic/encoded spans and coordinate integration not complete |
| LOD113C-D2GAME | 0xB4D30 | none established | Tower callback (slot 106) | MMS-R06 | Plain instructions read; target and launch operations unresolved |
| LOD113C-D2GAME | 0xB49F0 | none established | Generic shot callback (slot 110) | MMS-R07 | Plain instructions read; selection and builders unresolved |
| LOD113C-D2GAME | 0xB4C20 / 0xB4B10 | none established | Monster variant callbacks (slots 148 / 149) | MMS-R08 | Bodies read; shared encoded condition and variant producers unresolved |
| LOD113C-D2GAME | 0xB58A0 / 0xB5450 | none established | Simple projectile / trap callbacks (slots 85 / 93) | MMS-R09 | Table and named-row anchors located; complete instruction/callee reading pending |
| LOD113C-D2GAME | 0x1100 / 0x1130 | none established | Unit X/Y wrappers | MMS-R03 | Byte-equivalent C read; dynamic-path imports consume NRF-R10b with supplied valid caster coordinates |
| LOD113C-D2GAME | 0xAC9A → LOD113C-D2COMMON 0x51BF0 | 10786 | Skill calculation thunk → front end | MMS-R03 | Import resolved from original images; SCX-R11 contract reused, front-end C available; local callers supply `calc1` then `calc2` and context |
| LOD113C-D2GAME | 0x6F930 | none established | Whole missile creator | MMS-R03 | MCI/MIR contract reused; local call sites do not inspect its return |

Descriptions are analyst descriptions. The originating callback is established by the original table, not its label. AI scheduling/monster frame production is still the dispatch contract's gated boundary.

### Nova derivation and semantic map

Selector 99's table pointer leads to 0xB52A0. Signed identity and missile checks, including null records, branch to return 0 before any coordinate/calculation/creator call. Successful continuation sets caster bit 0x40 at unit offset 0xC4 when caster is non-null. This is the neutral skill-action condition; no wider meaning is inferred.

The request is zeroed over 0x5C bytes. Call sites obtain X then Y, and invoke D2Common ordinal 10786 with (caster, chosen `calc1` reference, skill identity, level) then the corresponding `calc2` reference. Request flags are 0x1F. The caller sets request owner/source, missile, explicit source, level and loops, leaving request skill identity zero. Primary speed is missile record dword 0x38 shifted left 6, secondary speed dword 0x3C shifted left 6. The original-image instructions corroborate this order, including stack changes during the expression calls; reading stack stores without accounting for pushed arguments would put loops/speed in the wrong fields.

| Semantic field | Original request mapping |
|---|---|
| controls, owner, source, missile identity | 0x00, 0x04, 0x08, 0x10 |
| source X/Y, relative aim X/Y | 0x14/0x18, 0x1C/0x20 |
| speed, skill identity, level, loops | 0x28, 0x2C, 0x30, 0x34 |
| skill `calc1`/`calc2`; missile `Param1`/`Param2` | skill 0x138/0x13C; missile 0x38/0x3C |

The two constant arrays are at D2Game RVAs 0xFB468 and 0xFB4A8. The former supplies aim Y and the latter aim X. Reading all 16 signed dwords from each matching image gives the ordered 16 pairs in MMS-R03. The primary loop reads pair indices 0,2,4,6,8,10,12,14, increments its byte index by 8, and stops at 64. The secondary loop, only for clamped `calc2` greater than 1, begins at pair index 1, advances by `calc2` indices, and ends after the last index at most 15. Its machine step is signed i32 `calc2<<2`; the verified safe domain avoids overflow. Values outside that domain remain MMS-Q02.

No local call other than coordinate, expression and creator dependencies appears in this callback. Therefore the local stage can be executed with supplied dependency results without guessing their interiors. Actual creation state, path draws and compiler/parameter behavior retain MCI/MIR/SCX gates. An empty request log is not a simulation of a successful creator.

### Unresolved operation inventory

Spray slot 88: current skill thunk 0xAABA; node value thunks 0xAD12/0xACB2; coordinate resolver 0x9FA30; direction/offset thunks 0xA7A8, 0xAD1E, 0xAC88, 0xACA0. A direction correction table at 0xFB518 and byte table at 0xFB510 are located, but their imported consumers and complete table semantics are not verified. Creator at 0x6F930 is known separately.

Breath slot 95: current skill thunk 0xAABA; missile-record getter 0x1250; coordinate resolver 0x9FA30; cleanup 0xB6720; follow-ups 0xB32E0/0xB3270; skill-node thunk 0xAD12; state thunk 0xA68E; event removal/scheduling 0x8E6F0/0x8EA80. The callback reads current game frame and skill/node values; their semantics and full start selector 53 are open. A successful creator does not alone establish the later event path.

Turret slot 105: coordinate resolver 0x9FA30 and calc front end; sign/quadrant offsets, signed centering and count loop read but not fully adjudicated. Slot 106 uses target refresh 0xE07A0, target getter thunk 0xA598, speed thunk 0xACFA and builder 0xA2800. Generic slot 110 uses monster missile getter thunk 0xACF4 and builders 0xA28F0/0xA1880. Slots 148/149 add monster-data byte offsets 0x0E/0x0F respectively, after base missile validation. The common skill condition chooses those two builders; complete binding/variant production is not established by the unsigned byte load alone.

## T01: effective Skills cells

The complete 2026-10-29 original-profile Storm read, source-profile identity, ranges and effective-file hash are recorded in [teleport T01](skill-monster-teleport.md#t01-effective-skills-cells). It is the same 173535-byte effective `data/global/excel/skills.txt`, SHA-256 `465c28e83657540e0d7ae5b0de826d5657be7fba1aa539a85c3ca3aeea34b498`. Its complete header and rows were joined before parsing by header name and named `skill` row. `Id` is an explicit decimal column. Known override inventory/physical archive origin and compiled correspondence retain that record's limits; text cells do not prove gameplay.

| Named row | Column | Raw cell | Claim |
|---|---|---|---|
| AndrialSpray | srvdofunc | `88` | MMS-R01 |
| PrimePoisonNova | Id | `192` | Explicit decimal identity for C01 |
| PrimePoisonNova | srvdofunc | `99` | MMS-R01 |
| FetishInferno | srvdofunc | `95` | MMS-R01 |
| DesertTurret | srvdofunc | `105` | MMS-R01 |
| ArcaneTower | srvdofunc | `106` | MMS-R01 |
| MissileSkill1 | srvdofunc | `110` | MMS-R01 |
| DoomKnightMissile | srvdofunc | `148` | MMS-R01 |
| NecromageMissile | srvdofunc | `149` | MMS-R01 |
| PrimePoisonNova | `calc1` | `lvl-1` | Text formula lead; not compiled execution |
| PrimePoisonNova | `calc2` | `2` | Text formula lead; not compiled execution |
| PrimePoisonNova | srvmissilea | `primepoisoncloud` | Linked missile name only |
| UnHolyBolt | srvdofunc | `85` | MMS-R09 source lead |
| ShamanFire | srvdofunc | `85` | MMS-R09 source lead |
| GargoyleTrap | srvdofunc | `93` | MMS-R09 source lead |

## Observation and review

MMS-O01: use a vanilla Expansion 1.13c character able to encounter the poison-nova skill. Record character class/level/equipment, difficulty, location, caster identity and selected skill/level before triggering an eligible action. Trace 0xB52A0 and each reached 0x6F930 call with full semantic request and actual return. Capture both speed parameters and evaluated calc values, all request directions in order, final callback return, record cap/loss and module identity. Compare the exact captured inputs with MMS-R03; do not infer successful missiles from creator call count. A gameplay case with all failures is useful only if naturally obtainable; it must not be fabricated through memory calls.

2026-10-29: no missile/nova callback invocation or gameplay effect was captured. The observation attempt was inconclusive, not a negative gameplay result. No gameplay vector has `runtime_observed: yes`; matching eligible action captures remain pending.

This initial source-pass entry preceded lint and blind review. Actual results and the current review ledger are recorded below; review completion is never inferred from source writing.

## Additional bounded lead

DeathMaul is not part of the named candidate handler set. A named Skills text row Id 308 selects action 136 and missile name `death mauler`; the original action table points to D2Game 0xC400. No complete output path was investigated here. The coordinating issue has recorded this nearby handler for deduplication/boundary review; this observation neither adds an in-scope rule nor establishes a new independent spec.


## C01: narrow compiled nova selector corroboration

The 2026-10-29 read-only original-profile record documented in [teleport C01](skill-monster-teleport.md#c01-narrow-compiled-selector-corroboration) also covered Skills record index 192: stored identity 192, start 0, action 99, collection count 357. The effective text row PrimePoisonNova has explicit `Id = 192` and `srvdofunc = 99`, establishing correspondence for these scalar fields only. No calculation reference/result, numeric missile identity, speed parameter, loaded-DLL fingerprint or gameplay execution was captured by that record. MMS-Q01 therefore remains open for those other fields.


### Named nova speed-parameter mapping

A new 2026-10-29 original-image decode of the D2Common missile loader at 0x7C8D0 identifies the two fields as `Param1` and `Param2` (case-sensitive column names). Descriptor stores at 0x7CBEF/0x7CC08 assign original string 0x8B354 (`Param1`) to record offset 0x38; 0x7CC1A/0x7CC33 assign string 0x8B34C (`Param2`) to offset 0x3C. Both descriptors use field type 2, matching i32 loader fields. This narrows the source-name mapping only; no original-profile numeric speed cell is asserted. The original D2Common file freshly hashes to manifest `LOD113C-D2COMMON` SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`. Complete loader byte-equivalent C was unavailable; the new instructions and named string bytes were read from the hash-matching image.


## Mechanical and self-review record

On 2026-10-29 the three owned spec/companion/gap trios were checked together with the repository spec linter: 0 errors and 0 warnings. Self-review checked source identities, exact per-rule gates, supplied dependency boundaries, vector arithmetic and public/private separation. This mechanical/self-review entry preceded blind review; the current applicable review ledger below controls review status.


## Fresh blind review MMS1 — full, 2026-10-29

Counts: MISMATCH 0; GUESS 0; GAP 1; DISCRIM 1; GATED 12; COVERAGE 3; HARDEST 3. The reader recomputed the original vectors without a mismatch. The counts describe that contract, not approval.

| Finding | Outcome and rationale |
|---|---|
| GAP: valid records with absent caster lack source-coordinate outcome | Edit/gate: MMS-R03 now requires a present caster; MMS-Q09 gates the valid-record absent-caster coordinate/result boundary. No safe coordinate default is invented. Early refusals remain separately available. |
| DISCRIM: vectors do not require equality of every semantic request field | Edit: fixture preamble now requires complete MCI/MIR semantic-request equality, all unmentioned scalars 0, references/callbacks absent and other conditions clear; this expressly excludes binary padding. Applies to every request in V02-V04. |
| COVERAGE: other early refusal guards absent | Edit: MMS-V01 now has six independent cases: negative identity, count equality, malformed absent skill record, negative missile, missile-count equality and malformed absent missile record. Each asserts zero downstream calls and unchanged condition. Original guards were re-read. |
| COVERAGE: secondary spacing greater than 2 absent | Edit: MMS-V04 uses spacing 3 and expects secondary pair indices 1,4,7,10,13 with exactly 13 total calls. It rejects a hard-coded odd-index ring and a wrong starting index. |
| COVERAGE: loops only -3 and 2 | Edit within V04: loops 0 are forwarded unchanged. No additional local branch is asserted; the whole request equality catches a clamp-to-1 error. |
| HARDEST: direction pairs and spacing | Edit: V04 adds irregular stride 3; V02 still distinguishes swapped axes and phase speeds. The exact original pair arrays remain unchanged. |
| HARDEST: signed spacing clamp and ring split | Edit: present-caster boundary is explicit; V03 proves negative-to-1 clamp/no second ring and V04 proves nontrivial positive stepping. Extreme overflow stays MMS-Q02. |
| HARDEST: ordered early refusal | Edit: expanded V01 distinguishes each named guard and no downstream effects. Malformed storage is explicitly synthetic. |
| GATED: Q01-Q08, MCI/MIR effects, SCX compiled results, NRF position and SCD outer behavior | No promotion: these limits remain explicit; new absent-caster Q09 narrows the verified stage. |

Those stage-boundary/vector changes required another full reader. MMS2 below records that review; its remaining assertion clarification requires MMS3.


## Fresh blind review MMS2 — full, 2026-10-29

Counts: MISMATCH 1; GUESS 0; GAP 0; DISCRIM 0; COVERAGE 0 uncovered deficiencies; HARDEST 3. GATED: nine existing question gates Q01-Q09 plus the existing consumed-dependency limitations remain explicit; the reader did not enumerate a single combined gate count. All refusal results and V02-V04 request directions, speeds, loop values and call counts recomputed correctly.

| Finding or difficult boundary | Outcome and rationale |
|---|---|
| MISMATCH: V01(c) record-resolution versus no-access wording | Edit: V01 now explicitly distinguishes resolving a record as absent from reading a field of that record. Case (c) asserts no skill-record field read, no missile-record resolution and no source-coordinate/calculation/creation call after the absent skill-record resolution; it does not assert that resolution never occurred. |
| Same refusal terminology in other cases | Edit: (a)/(b) state no prepared-record resolution or field read before failed identity guards; (d)/(e) acknowledge the skill selector read and state no missile-record resolution/read; (f) acknowledges absent missile-record resolution then forbids field reads/downstream calls. The generic output names only source-coordinate, calculation and creation calls. Results, inputs, guards and numeric outputs are unchanged. |
| HARDEST: secondary index walk | No change: V04 spacing 3 produces indices 1,4,7,10,13 and exactly 13 total calls; all were recomputed correctly. |
| HARDEST: signed spacing and overflow | No change: V03 proves negative spacing clamps to 1, V04 proves positive stepping, and Q02 retains the extreme signed-step-overflow boundary. |
| HARDEST: early guard ordering | Assertion clarification above; no behavior change: identity bound, skill resolution, signed missile selector bounds and missile resolution remain in source order before coordinates. |
| Existing gates/dependency limits | No promotion: Q01-Q09 and the MCI/MIR/SCX/NRF/SCD boundaries remain explicit. |

At the MMS2 handoff, teleport MTP2 and melee MML2 had completed their final applicable passes, while the missile assertion clarification required MMS3. That pending state is historical; the final applicable MMS3 ledger below now controls current review status.


## Historical blind review MMS3 — full, 2026-10-29

Counts: MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; GATED 9 existing questions (Q01-Q09); COVERAGE 0 uncovered deficiencies; HARDEST 3. Upstream MCI/MIR/SCX/NRF/SCD limitations remain explicit dependencies and are not counted as additional local questions. The fresh reader independently recomputed V01(a)-(f), V02, V03 and V04, including exact resolution-versus-field-read distinctions, complete semantic request defaults, ordered directions, speeds, loop values, request counts and final returns. No changes were recommended.

| Difficult boundary or retained limit | Outcome and rationale |
|---|---|
| Early guard ordering | No change: identity bounds precede skill-record resolution; an absent resolved skill record permits no field read; signed missile bounds precede missile-record resolution; all failures precede source-coordinate/calculation/creation calls. The six cases now distinguish these stages exactly. |
| Signed spacing and overflow boundary | No change: nonpositive spacing clamps to 1 for the supplied present-caster stage; positive nonoverflowing spacing follows the exact step. Extreme signed-step overflow remains Q02 and absent-caster execution remains Q09. |
| Secondary index walk | No change: V02 spacing 2 and V04 spacing 3 reproduce the exact ordered secondary indices and total counts, with the separate secondary speed and zero-default request fields preserved. |
| Existing gates and dependency limits | No promotion: Q01-Q09 and consumed dependency limitations remain explicit. Review completion establishes an unambiguous partial contract, not missing runtime or full integration evidence. |

MMS3 completed the earlier nova request-stage contract. Its gate remains historical for the changed continuation below, whose completed fresh review is recorded at the end. Status remains draft; runtime confirmation, evidence gates and user approval remain separate.

## Missile/tower continuation — 2026-10-29

This pass reuses the completed nova instruction proof without repeating its guards. It extends positive position/expression/creation integration and closes the tower callback and 64-direction helper. Prior unresolved tower statements describe initial coverage only. All other callback limits remain unless specifically narrowed here. New source readings checked original-image-checked reconstruction coverage first, read the available higher-level forms before instructions and corroborated the newly interpreted byte-encoded builder spans against the original image. No eligible gameplay case was observed during this continuation; no runtime confirmation is added.

### Identity and newly read roots

Original D2Game was freshly hashed to `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`, 1138688 bytes; Common to `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`, 679936 bytes. Both match the source manifest identities.

| Artifact | RVA / ordinal | Supported assertions | Method / exact boundary |
|---|---|---|---|
| LOD113C-D2GAME | 0xB4D30, slot 106 | MMS-R06b | Entire original-image callback through 0xB4DD2; validation/condition/refresh/cache gate/speed/helper arguments/final 1 |
| LOD113C-D2GAME | 0xA2800 | MMS-R10 | Entire original-image builder through 0xA28E2, including zeroing and loop bounds; none of its creation returns are tested |
| LOD113C-D2COMMON | 0x7C4A0, ordinal 11169 via Game thunk 0xACFA/IAT 0xF8890 | MMS-R06b speed | Byte-equivalent higher-level form read first, original-image instructions corroborated; unsigned bytes 0x9A/0x9B and signed wrapped multiply/truncating divide/add |
| LOD113C-D2GAME / LOD113C-D2COMMON | Game 0x1100/0x1130, Common 0x34BB0 ordinal 10867 / 0x34B80 ordinal 10750 | MMS-R03/MMS-R10 coordinates | Byte-equivalent higher-level forms read; X/Y integer-word getter behavior consumes NRF-R10b; static kinds read their prepared coordinates |
| LOD113C-D2COMMON | 0x51BF0, ordinal 10786 | MMS-R03 expression integration | Byte-equivalent front end read; SCX-R11 pool/context and explicit-bytecode evaluation reused, not compiler output |
| LOD113C-D2GAME | 0xE07A0; Game thunk 0xA598 -> Common 0x348F0 ordinal 10392 | MMS-R06b refresh/cache | Compatible 2026-10-29 MTP-R09 proof reused with its valid game/path/registry stage and saved-key limits |
| LOD113C-D2GAME | 0x6F930 | MMS-V05/MMS-V06/MMS-V07 actual creator stages | Compatible MCI-R01 through R11 / MIR-R01a proofs consumed; prefix vectors stop before unread/dependency-gated later effects |

Some internal analyst labels inside these functions use shifted symbols; original function RVAs and direct-image branch targets above establish identity. No label text establishes an algorithm.

### Tower request mapping and concrete closure

At 0xB4D81 the callback writes unit condition bit 0x40, before refreshing target and before rejecting absent/self cache. Refresh preserves only its matching eligible cache; a different registry reference clears the cache and retains the saved key. Target coordinates are never read by slot 106. The target is an eligibility gate for a ring around the caster, rather than an aim argument. Signed missile identity is range checked but no nonabsent missile record is required in this caller. The speed helper returns 0 on absent record; creation validates it separately.

The builder zeroes 0x5C request bytes, sets controls 3 (explicit start and relative aim), reads the caster position once, and stores owner but leaves source reference zero. It stores missile/skill/level arguments at request offsets 0x10/0x2C/0x30. Nonzero incoming speed switches controls to 7 and stores speed at 0x28; the already-fixed-point bit 0x10 remains clear. Thus an incoming 11 is whole units, unlike nova speed 192. Zero speed leaves controls 3 and permits MCI-R05 row speed production; zero does not force final missile speed zero. Loop count remains 0. All other semantic request fields retain zero/absent defaults.

The loop reads 64 signed dwords from each original constant array, X at D2Game RVA 0xFB870 and Y at 0xFB770, increments its byte index by 4 and stops at 0x100. All 64 pairs were read. They exactly equal the sixteen listed first-quadrant pairs and their three successive quarter-turn transforms in MMS-R10; MMS-V06 lists all expanded pairs. The first is (30,0), indices 16/32/48 are (0,30)/(-30,0)/(0,-30), and the final pair is (29,-2). The creator is called once per pair and no return controls continuation. Callback 106 returns 1 only after the builder completes normally.

Missile row `Vel` and `VelLev` field naming/schema are reused from MCI-R05 and GDA-R03/GDA-R10, rather than inferred from the helper name. Level multiplication wraps before signed truncation by 8; subsequent addition wraps. The getter supplies whole units, and MCI scales overrides to 8.8 then applies the global percentage.

### Positive integration fixtures and retained effects

MMS-V05 uses explicit synthetic skill-pool bytes: parameter selector 16 (invocation level), literal 1 and subtraction, followed by a separate literal 2 expression. The calculation references select offsets 0 and 6 within nine bytes. SCX-R11 and SCX-R01/R03/R04/R06/R21 execute the expressions to 2/2 with context caster/skill 192/level 3; no text compiler or original-profile compiled formula is claimed. Real dynamic position words yield source (100,200). The first actual creator stage selects the supplied room, retains (100,202), scales fixed speed 192 to 144, allocates, seeds and numbers the supplied successful unit, then initializes its metadata. The fixture stops after MIR-R01a before creation-event/path/registry work; no final creator or nova callback return is asserted at this prefix.

MMS-V06 executes all 64 actual creator early refusals due to absent owner room. It does not substitute an arbitrary creator result: MCI-R01/R02 pass and R03 refuses each request. Thus no allocation or seed step occurs, and the caller still returns 1. A stale-cache case tests the post-condition-write refusal, and a negative-level/zero-speed case distinguishes truncation and override selection. MMS-V07 instead reaches a positive first allocation/seed/id/metadata prefix: 11 whole units becomes 2816 fixed-point then 2112, start(100,200) retains(130,200), and seed{1,666} advances once to{1791398751,0}; child{1791398751,666} is not stepped. It stops at the same metadata boundary and makes no whole-ring draw assertion.

No general path-building, full range/damage preparation, callback, monster-mod event, creation final success or client-message effect is newly asserted. Those remain the exact MCI/MIR dependency gates. Spread105 remains MMS-Q05; variants/selection, spray, breath and trap/simple callbacks retain their existing named questions. Original-profile numeric missile/parameter/formula capture remains MMS-Q01 and gameplay remains MMS-Q08.

### Current handoff ledger

Self-review checked all 64 direction pairs, owner/source distinction, whole-unit versus fixed-point override, target eligibility versus aim, negative truncation, context and bytecode offsets, game/child seed distinction and the first-prefix stopping boundary. Mechanical results follow the final edit pass. At this handoff, fresh full smaller-reader review of this changed contract was pending. Its completion is recorded below; MMS3 remains a historical review of the earlier request-only boundary.

## Final applicable continuation review — full, 2026-10-29

The coordinator relayed the fresh smaller-reader review: MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; COVERAGE 0 uncovered deficiencies; GATED 9 local questions (MMS-Q01 through MMS-Q09), with the documented shared dependency gates retained; HARDEST 3. The reader recomputed the vectors and found no actionable defect. No contract, vector or gate change was recommended or made.

| Finding or difficult boundary | Outcome and rationale |
|---|---|
| Mismatch, guess, gap, discrimination and coverage checks | No change: all actionable and uncovered counts were zero; the vectors were recomputed coherently. |
| MMS-R10 / MMS-V06 sixty-four aim pairs | No change: the ordered quadrant values and transforms are explicit; V06 lists the complete expanded order and checks all sixty-four whole requests. |
| MMS-R03 / MMS-V05 / MMS-V07 fixed-point versus whole-unit speed | No change: nova fixed-point 192 becomes 144, while tower whole-unit 11 becomes 2816 then 2112; flags and prefix boundaries distinguish both paths. |
| MMS-R03 / MMS-V02 through MMS-V04 secondary ring | No change: negative spacing suppresses the second ring, spacing 2 and 3 select the exact listed indices, and primary requests precede secondary requests. |
| Existing local questions and shared integration limits | No promotion: nine local questions and the precise shared dependency gates remain unchanged. Prefix success does not establish final creation, whole-ring draw totals or gameplay. |

The blind-implementer stopping rule is complete for this continuation. Earlier MMS1/MMS2/MMS3 reviews and the pending handoff remain historical. Status remains draft; evidence gates and user approval remain separate. Final mechanical verification after this ledger edit: missile/nova spec, companion and gap lint, 0 errors and 0 warnings.
