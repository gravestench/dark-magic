# Evidence: skill cast start and do dispatch

Behavioral contract: [skill cast start and do dispatch](../../specs/skill-cast-start-and-do-dispatch.md).
Open items: [evidence gaps](../../plans/gaps/skill-cast-start-and-do-dispatch.md).

## Identity, reuse and limits

The original artifacts are `LOD113C-D2GAME` and `LOD113C-D2COMMON`, Windows x86 PE32, identified by the [source manifest](../source-manifest.md). Public addresses below are module-relative RVAs unless explicitly marked VA; function descriptions are analyst descriptions, not exported original names. Readers need their own matching game files and a PE-aware disassembler. No original binary, raw capture, private source path or reconstruction artifact is distributed here.

The 2026-10-12 record established the local validator/start/action/frame/scheduler instruction paths and inspected the original dispatch tables. Those compatible instruction readings retain their original date and method. Their broad verified labels did not establish unread target, activation, creator, resource or repeat-handler interiors. The current contract separates local decisions using supplied operation results from actual dependency effects; new locators do not renew old verification. No archived gameplay observation or verified original-profile archive table-cell record exists for this spec.

On 2026-10-15 missing paths were checked against available byte-matching coverage before reading instructions; none of the 19 targeted routines had such a reconstruction. Original-image decoding corroborated byte-encoded spans for D2Game 0xDF7B0, 0xE0660, 0xF14F0, 0xA1E80, 0xCCFE0, 0xA1880, 0xA28F0 and 0x9FF70. The image was hash-checked against LOD113C-D2GAME. Complete requested extents decoded, including embedded table bytes; decoding an extent alone is not proof that every byte is an instruction or that an unread callee chain is complete. The old core decision spans at 0xA19A0, 0x9FD00, 0x9F3B0 and 0xA0830 were also checked through original-image listings. The missing dependency paths below remain explicit.

## Function records and roots

| Artifact | RVA / export | Analyst description | Rules | Method, date and established boundary |
|---|---|---|---|---|
| LOD113C-D2GAME | 0x78550; mode table RVA 0xFA3F8 | Mode dispatcher | SCD-R01/SCD-R02 | Original table/decompile reading 2026-10-12; mode-table linkage to entry handlers. Mode setting consumes NRF-R14a; other preparation effects remain gated |
| LOD113C-D2GAME | 0x787C0, 0x78840 | Unit-target/point-target skill mode entry | SCD-R01 | Instruction reading 2026-10-12; complete plain-instruction caller recheck 2026-10-15 corrects route-specific target/entry/mode order and fatal-target boundary; latch/validator behavior retained |
| LOD113C-D2GAME | 0xDA750 | Entry-preparation request target | SCD-R01/SCD-Q10 | Located from both entry calls and instructions inspected 2026-10-15; encoded helper spans, imported operations and actual effects remain unverified. No cleanup meaning is adopted from the analyst label |
| LOD113C-D2GAME | 0xA0830 | Start validator | SCD-R02 | Instructions 2026-10-12, decision-span image corroboration 2026-10-15: current skill, target check, level/charge override, identity, party/town gates, start and player-reset branch |
| LOD113C-D2GAME | 0x9FD00 | Start dispatch | SCD-R03/SCD-R04/SCD-R05a/SCD-R05b | Instructions 2026-10-12, image corroboration 2026-10-15: validity/activation/town/line/index checks, callback, mana request and periodic purge |
| LOD113C-D2GAME | 0xA19A0 | Action dispatch | SCD-R06/SCD-R07/SCD-R08/SCD-R12/SCD-R13 | Instructions 2026-10-12, image corroboration 2026-10-15: checks/index selection, callback, missile request/result override, resource-request order |
| LOD113C-D2GAME | 0x79A10; frame table RVA 0x1079D0 | Player frame-event dispatcher | SCD-R09a | Instructions/table 2026-10-12; frame handler returns 2 cause call 0x79850 |
| LOD113C-D2GAME | 0x79170 | Skill-mode frame handler | SCD-R09a/SCD-R09c | Instructions 2026-10-12: current-skill gate, argument store, exceptional and ordinary decisions, finishing result |
| LOD113C-D2GAME | 0xA2060 | Frame action stage | SCD-R09b | Instructions 2026-10-12: bonus level and action arguments; result ignored before scheduler |
| LOD113C-D2GAME | 0x9F3B0 | Periodic scheduler | SCD-R11 | Instructions 2026-10-12, targeted read/image 2026-10-15: clamp, i32 numerator, signed division, multiply/add; ordinary/sentinel replacement requests |
| LOD113C-D2GAME | event table RVA 0xFBC34 → 0xA1E80 | Type-8 repeat handler | SCD-R11/SCD-Q09 | Prior table/decompile lead 2026-10-12; instructions/image 2026-10-15 investigated state/current-skill branches, not all eligibility callees |
| LOD113C-D2GAME | 0x9E420 | Direct action-table dispatcher | SCD-R14a | Instructions 2026-10-12: signed bound, non-null pointer, callback only; entire encoded decision/call span corroborated against original image 2026-10-15 |
| LOD113C-D2GAME | 0x40E10, 0x39DA0 | Missile callers of direct dispatcher | SCD-R14a | Instructions 2026-10-12: selected numeric field forwarded; enclosing missile eligibility not claimed |
| LOD113C-D2GAME | 0x9FC00, 0xE0660 | Target and party predicates | SCD-Q03 | New instruction read 2026-10-15; second routine's encoded span image-checked; missing hostile/party/owner callees remain |
| LOD113C-D2GAME | 0x9EC80 | Activation predicate | SCD-Q06 | New instructions 2026-10-15: non-player fast return, current skill, positive charges, level-based mana comparison and selector-116 special predicate; imported semantics incomplete |
| LOD113C-D2GAME | 0x9FA30 | Target-coordinate resolver | SCD-R04/SCD-R07/SCD-Q10 | New instructions 2026-10-15: target-unit coordinates or stored target point, success requires both nonzero; target resolution/coordinate imports not fully followed |
| LOD113C-D2GAME | 0xA1880, 0xA28F0 → 0x6F930 | Lob/ordinary missile request builders → creator | SCD-R07/SCD-Q10 | New instructions/image 2026-10-15: target and optional quantity gates before core creator; dispatcher ignores builder return. Core creation has its own contract |
| LOD113C-D2GAME | 0x9F850, 0xA1790, 0x9FF70 | Mana, quantity and delay helpers | SCD-R05a/SCD-R08/SCD-Q10 | New instructions 2026-10-15; delay encoded span image-checked. Actual charged-item/resource/state/event effects require further dependencies |
| LOD113C-D2GAME | 0xDF7B0, 0xE1930, 0xDFA90 | Animation preparation, exceptional frame predicate, finished predicate | SCD-R09a/SCD-Q01 | First two new instruction reads 2026-10-15, scheduler image-checked; finished-predicate local comparisons retain 2026-10-12 reading. Preparation/predicate side effects not fully verified |
| LOD113C-D2GAME | 0xF1730, 0xF17E0 → 0xF14F0 | Immediate wrappers/dispatch | SCD-R10/SCD-Q04 | Caller relationships and wrapper entry slices read 2026-10-15; dispatch body/image read; wrapper roots/target-selector callees incomplete |
| LOD113C-D2GAME | 0xADB30, 0xA37D0 | Other consuming forced-action callers | SCD-R10/SCD-Q04 | New instruction read 2026-10-15: state/event and skill gates, exact reached action arguments; dispatch roots incomplete, item-event label not adopted as fact |
| LOD113C-D2GAME | 0x91B70, 0x9E8E0 | Progression and active-state table users | SCD-R14b/SCD-Q09 | New instructions 2026-10-15: selectors/bounds/calls; progression has charge-state and seed operations; surrounding eligibility/state contracts incomplete |
| LOD113C-D2GAME | 0xCD630 → 0xCCFE0 | Parse target check → unit validation | SCD-R15 | New instructions 2026-10-15, validator image-checked: output clearing, length/kind gates, existence/inventory/area/range delegation |
| LOD113C-D2GAME | 0xCE7F0, 0xCE870, 0xCE950, 0xCE9D0, 0xCEA50, 0xCEAB0, 0xCEB10, 0xCEB70 → 0xCE620 | Skill-on-unit command callers/mode selection | SCD-Q02/SCD-Q07 | Caller set from 2026-10-12 reading, 0xCE620 newly read 2026-10-15: usability/mode/fallback/redirect calls identified, not all followed |

### Original dispatch tables

D2Game start table RVA 0x107338 (VA 0x6FD27338), 91 dwords; action table RVA 0x1074A8 (VA 0x6FD274A8), 191 dwords. Fresh read of the hash-matching image on 2026-10-15 confirms 64 non-null start slots (1–29,31–65) and 152 non-null action slots (1–152). Other slots within the respective checked bounds are null. This is original engine dispatch-table evidence, not a Skills archive-cell read. The player mode table is RVA 0xFA3F8, two dwords per mode; frame table RVA 0x1079D0 has one per mode. Their original-image reads corroborate the handler links used by SCD-R01/SCD-R09a.

Both callbacks receive game in ECX, unit in EDX, skill identity and level on the stack, returning signed EAX. Start dispatch has identity in EAX and stack game/unit/level/force. Action dispatch has unit in EAX and stack game/id/level/consume/forced/itemTarget. These calling conventions are investigation mappings, not required APIs.

### Imported semantic operations

Original 2026-10-12 import resolution used the D2Game import slots and corresponding D2Common exports. It supports the local caller roles below, not a complete new reading of every imported implementation.

| Semantic operation | LOD113C-D2COMMON export ordinal | RVA | Caller rules |
|---|---:|---|---|
| Current skill | 10511 | 0x303A0 | SCD-R02/SCD-R06/SCD-R09 |
| Skill identity | 10601 | 0x4E1D0 | SCD-R02/SCD-R09b |
| Bonus-inclusive level | 10306 | 0x505C0 | SCD-R02/SCD-R06/SCD-R09b; full integration SCD-Q10 |
| Find owned skill | 10630 | 0x4F070 | SCD-R06 |
| Get/set skill conditions | 10397 / 10137 | 0x4DE40 / 0x4DE60 | SCD-R06/SCD-R09a |
| Set current skill | 11103 | 0x303E0 | SCD-R02/SCD-R03/SCD-R06 |
| Charge marker | 10537 | 0x4E240 | SCD-R02/SCD-Q06 |
| Calculation front end | 10786 | 0x51BF0 | SCD-R08/SCD-R11; SCX-R11 supplies its separate contract |
| Room / town test | 10331 / 10057 | 0x2FE10 / 0x3C390 | SCD-R03/SCD-R06 |
| Line test | 10479 | 0x4F160 | SCD-R04; actual execution SCD-Q10 |
| Mode set | 11090 | 0x33920 | SCD-R01, NRF-R14a |
| Store frame-event argument | 10542 | 0x2F0C0 | SCD-R09a |
| Set target point | 10286 | 0x349D0 | SCD-R01; imported target resolved 2026-10-15 and valid-input branch read through byte-equivalent C reconstruction |
| Set selected target unit | 11129 | 0x30E80 | SCD-R01; imported target resolved 2026-10-15 and valid-input branch read through byte-equivalent C reconstruction |

Unresolved imported operations on newly inspected helper paths are not assigned meanings from labels. SCD-Q03/SCD-Q06/SCD-Q10 gate their actual integration until imports and owned callees are fully followed. In particular, the historical concern about D2Game thunk 0xA3B2 / D2Common ordinal 10507 and apparent argument count in caller 0xCE7F0 remains a networking investigation item; stack cleanup alone does not establish total argument count.

### Semantic layout map

The behavioral contract uses logical fields; these original layouts only permit independent investigation.

| Semantic field | Original representation | Rules |
|---|---|---|
| Current skill/list | Unit +0xA8, list +0x10 | SCD-R02/SCD-R09 |
| Selected command skill slots | list +0x08 / +0x0C | SCD-R15; left/right assignment unresolved |
| Skill identity | node record pointer +0x00; record word +0x00 | SCD-R02 |
| Conditions A/B | node flags +0x0C bits 0/1 | SCD-R06/SCD-R09a |
| Base level/charge marker/count | node +0x28/+0x34/+0x38 | SCD-R02 |
| Do-handled latch | Unit +0xC4 bit 6 | SCD-R01/SCD-R07/SCD-R09 |
| Sequence-event value | Unit +0x38 bits 8–31 | SCD-R08/SCD-R09a |
| Completion condition | Unit +0xC6 bit 0 | SCD-R09a |
| Game frame/pending events | Game +0xA8 / Unit +0xDC | SCD-R11 |
| Start/action/progression selectors | Skills +0x2C/+0x2E/+0x30..0x34 | SCD-R05a/SCD-R06/SCD-R14b |
| Animation/missile/period | Skills +0x10/+0x46/+0x128 | SCD-R07/SCD-R08/SCD-R11 |
| Alternate action/line/delay | Skills +0x16A/+0x18F/+0x190 | SCD-R04/SCD-R06/SCD-R08 |
| Skill flags | Skills +0x04 onward: decquant 0,lob 1,aura 5,periodic 6,InTown 8,TargetAlly 26,ItemCheckStart 33,usemanaondo 37 | Local branch consumers |
| Active-state/missile selectors | States +0x36; Missiles +0x3C/+0x4C | SCD-R14 |
| Stored target point and selected unit | Path +0x10/+0x12 are the point-setter's low-16-bit coordinate writes; Path +0x58 is selected unit, with target kind/id cached at +0x5C/+0x60 when non-null | SCD-R01; valid setter branches only, unrelated preparation effects remain SCD-Q10 |

Skill/missile/state descriptor roots are D2Common 0x61EC0/0x7C8D0/0x5F6E0, read on 2026-10-12. Names are consumed through GDF's current schema contract; detailed schema ownership remains there. No archive row/column is reproduced here.

## Gap assessment and correction rationale (2026-10-15)

SCD-Q01: animation scheduler reads frame marks, accepts mark codes 1–4 and calls D2Game 0x8EA80. Its table sends codes 1,2,4 through a branch incrementing the event argument, while code 3 uses a zero argument. Original bytes at 0xDF86A/0xDF87F corroborate the byte load/jump-table decode. D2Game 0xE1930 calls 0xE07A0, an imported condition with input 0x400, another imported predicate and possibly 0xE1610; its returned 1/2 alone does not describe those side effects. These narrower new readings investigate, but do not close, real animation production/exceptional integration.

SCD-Q03: 0x9FC00 can pass absent/self targets, invokes further eligibility routines for targetable-only skills, inspects target flags/mode and may clear the selected target before returning 1. 0xE0660 checks player caster, player/monster target, skill flag, monster owner and party relationships. Its encoded span agrees with the original image; hostile/targetability, owner and party callees are still required. The validator's supplied boolean contract is retained; actual target mutation/eligibility is explicitly gated.

SCD-Q04: caller relationships establish 0xF1730/0xF17E0 → 0xF14F0, but not complete gameplay roots. The latter saves latch at 0xF158B–0xF1595 and restores it at 0xF16E1–0xF1705. **Correction:** the prior text attributed that post-call latch decision to a caller argument; the image and instructions show restoration of the saved entry latch. Target setup/restore and target-selector helper eligibility still prevent promotion beyond asm-only. The 0xADB30/0xA37D0 bodies have state/event branches; the earlier “item-event” role was a label, not a proven root.

SCD-Q05: 0xA19A0 reads the start pointer during consumption without its start-dispatch bound and conditionally replaces the action index without checking ItemEffect. The read-only collection below assessed effective text-table prevalence; original-profile identity and compiled-field correspondence remain unestablished. Valid supplied selectors do not depend on those gaps. No out-of-bounds access is converted to an invented null result.

### Effective Skills collection attempt

On 2026-10-15, the available session inventory reported one member as `lod-113c` and the others as modified profiles. The version/profile registry was unavailable. The reported vanilla session's Storm loader returned `data/global/excel/skills.txt`: total 173535 bytes, collected completely in contiguous offset/length ranges (0,8192), (8192,32768), (40960,32768), (73728,32768), (106496,32768), (139264,32768), (172032,1503). Every success, offset and returned length was checked before joining decoded `base64` bodies; split rows were retained. Header-name lookup selected `skill`, `Id`, `srvstfunc`, `srvdofunc`, `ItemEffect`, with zero-based column positions 0,1,4,5,143. All 357 nonempty records had numeric explicit Id values; no positional convention replaced that column.

The complete returned text contained 112 nonempty start selectors, 274 nonempty action selectors and 72 nonempty ItemEffect selectors. Their numeric ranges were 1–65, 1–152 and 1–151 respectively. No nonempty selector in this returned file was negative, nonnumeric or outside its applicable start/action bound. Empty cells were kept distinct from explicit zero; compiled interpretation was not inferred. This is a complete negative scan of this effective text file, not proof about compiled runtime records or an independently authenticated original archive profile. Effective-file hash, physical supplying archive, loaded-module fingerprints and override inventory were not captured. No original-profile data label or gameplay confirmation is promoted. SCD-Q05 remains open for identity/compiled correspondence and malformed-access behavior.

SCD-Q06: 0x9EC80 checks positive charged count, otherwise derives a mana comparison from current-skill data and bonus level; selector 116 calls a separate imported special predicate. This differs from the validator's charges==0 test and is not a complete mana-cost contract. Imports/special-state and charged-skill dependencies remain unresolved; forced start and forced action controls remain independently verified locally.

SCD-Q09: 0xA1E80 has distinct carried-id -1 and positive skill branches, current-skill/state/stat-list checks and event-removal paths. 0x91B70 has progression charge-state selection and seed operations around direct action-table calls, while 0x9E8E0 selects an active-state callback. Reading their local table calls does not verify all progression/aura eligibility. These remain asm-only with exact gates rather than overclaiming repeat gameplay. The scheduler's separate repeat-kind input is now stated and discriminated by SCD-V15; the ordinary-kind outputs in SCD-V05 are unchanged.

SCD-Q10: 0xA1880/0xA28F0 can return 0 before 0x6F930 when target coordinates cannot resolve, and can perform optional quantity eligibility checks. **Correction:** a valid srvmissile branch sets latch/result=1 even when the request builder returns 0. SCD-V13 now separates that verified decision from actual creation. SCD-V10 retains game seed {1,0x29A} → {0x6AC6935F,0}, with successful creation-seeding entry and no later draws explicitly supplied. NRF-R17a verifies seeding rather than the complete creator; the [missile creation contract](../../specs/missile-creation-and-initialization.md) owns core request validation/creation and retains its own gates. SCD-V09 supplies a creator result rather than claiming real creation has no seed draw.

The mana helper 0x9F850 distinguishes current charged skill from ordinary deduction; quantity helper 0xA1790 inspects weapons/items and delegates quantity updates; delay 0x9FF70 interacts with state/stat-list/event operations. None is established by the dispatcher's request alone. Local call timing and argument lists remain verified with supplied operations; actual resource writes and complete draw sequences stay open. Preparation/reset/event-storage calls likewise require contracts rather than silently treated no-ops.

SCD-Q02/SCD-Q07: 0xCD630 clears outputs before length/type checking; 0xCCFE0 identifies inventory item, area and range boundaries. 0xCE620 calls usability/mode helpers, can choose fallback, redirects a target through state-dependent logic and branches into mode/action requests. Those chains were investigated without guessing left/right identity or deriving selected mode solely from Skills.anim. The original text omitted parsed-output initialization; the explicit gated rule now includes the observed clearing.

SCD-Q11: the old return-catalog question did not support a claim that callback returns are boolean. Return propagation was already established by the start dispatcher. SCD-R05b now states that existing supplied-return fact, exercised with 7 in SCD-V11; individual callback catalogs remain an outside-boundary question. This is not evidence that a particular shipped callback returns 7.

Other-build equivalence is outside the 1.13c evidence policy and contributes no parity inference. All observations remain pending. Every remaining gap was assessed against these source/dependency records; a partial local reading was not promoted into a complete integration claim.

## Animation observation

SCD-O01 is a pending procedure, not a capture. Use an owned vanilla LoD 1.13c session with a Sorceress at level 6 or higher, no charged items and sufficient mana, in the Blood Moor on Normal difficulty. Use Fire Bolt as an initial candidate, but inspect its effective Skills row first to establish whether it actually exercises the server-missile branch; class/skill name alone does not establish that branch. Choose an available ordinary skill without that branch for comparison, establish its actual row, and record both identities, levels and all callback effects. A branch requiring a distinct setup remains pending until that setup is established.

Observe D2Game 0x78840/0x787C0, 0x79170, 0xA2060 and 0xA19A0 on the matching image. Capture entry mode, current skill, latch before/after, event code/argument, condition A/B, action invocations and missile-helper result, then interrupt a separate use before its action event. Count actual events and distinguish ordinary versus exceptional route. Compare the observed inputs to SCD-R09a, not to a guessed animation-frame number. Record capture window, record count, drops and unobserved branches. An interrupted cast/empty capture does not establish that the path is never called.

## Periodic observation

SCD-O02 is pending. Use an owned vanilla LoD 1.13c Paladin at level 6 or higher with an available aura in the Blood Moor on Normal difficulty and enough resources for repeated uses. First inspect the effective Skills record to establish the aura/periodic flags and effective perdelay reference/value; do not infer eligibility solely from the skill's name. Record class, level, skill, row, equipment, activation and state prerequisites. If the chosen skill does not reach the scheduler, establish eligibility before treating absence as evidence.

Observe D2Game 0x9F3B0 and 0xA1E80. For each reached scheduler call record frame, effective level, perdelay result, repeat kind, due frame, payload and replaced events. Compare due against SCD-R11 with signed i32 arithmetic, varying the frame at initial activation. For a positive nonoverflowing sequence confirm due congruent to 1 modulo p. The repeat eligibility/predicate branches need their own captured inputs. Record counts/drops and actual coverage; a scheduling observation does not verify every repeat condition.

## Migration and review record

2026-10-15: existing scope/file boundary retained; original source narrative and fixture encodings archived privately with integrity metadata. Public detailed evidence now uses artifact identities and original RVAs/exports, while the behavioral contract uses semantic fields. The ten existing vector identities retain their numerical results; six additional vectors distinguish nonboolean propagation, bypass limits, request-versus-creation success, exceptional frame behavior, special repeat payload and direct dispatch. Supplied operations make resource state and argument/order assertions explicit. Static labels describe their stated local boundary; no new runtime label is claimed.

Self-review on 2026-10-15 checked original numerical results, supplied-stage effects and logs, rule/vector coverage, semantic layouts, source dates and remaining gates. Mechanical lint was run before self-review and is repeated after edit passes. The first fresh full blind-implementer review and its triage are recorded below; another fresh full review is required for the resulting changed contract. Remaining assessed gaps are in the linked gap file. No implementation comparison or plan was requested or created.

Historical mechanical checks on 2026-10-15: `python .agents/skills/spec-writing/tools/check_spec.py skill-cast-start-and-do-dispatch` returned 0 errors and 0 warnings for the first-review contract; the changed contract is checked again at handoff. No tests or gameplay actions were run.

## Full blind-implementer reader 1 and triage

The fresh full reader used the behavioral spec and its cited dependency contracts without binary/source/runtime access. Returned counts: MISMATCH 1, GUESS 0, GAP 0, DISCRIM 1, GATED 14, COVERAGE 12, HARDEST 3. These counts belong to that reviewed contract, not the revised one below. A second full reader is pending; no completed review or approval of the changed contract is claimed.

| Finding | Outcome |
|---|---|
| MISMATCH: SCD-V09 frame action used consume=0 while SCD-R09b uses 1 | Spec edit: V09 had explicitly substituted frame action and did not cite R09b, so its supplied consume=0 did not by itself contradict a numerical rule output. It nevertheless failed to exercise the real frame-stage controls. V09 now executes R09b, logs exact action arguments (20,3,1,0,0), and requires one or three delay evaluations after the respective reached actions. Mana remains 2560 because the supplied start slot is populated and usemanaondo is clear; no result was fitted to an observation |
| DISCRIM: SCD-V09 did not assert caller controls | Spec edit: exact invocation counts and consume/forced/itemTarget values are asserted in both cases; consume=0 additionally fails the required delay-evaluation counts. R09b is now cited. All initial latch, return, missile count and mana results are retained |

### Advisory coverage outcomes

| Finding | Outcome |
|---|---|
| COVERAGE: SCD-R02 missing current skill, invalid identity, party rejection | No change: R02 gives explicit ordered exits and whether reset occurs; V12(a) already exercises the zero/no-reset/no-callback result, and V03/V06/V12(e) distinguish reset-producing failures. Additional fixtures for the same no-reset result are advisory. Actual party production remains SCD-Q03; no new local decision is missing |
| COVERAGE: SCD-R03 absent objects, invalid identity, activation failure | Spec edit: V17(a)/(b) distinguish first/later absent-room results; V17(c) proves activation failure precedes the second room gate. Other absent-object/identity cases have the explicitly specified initial zero/no-reset result. No new actual room/activation integration is claimed |
| COVERAGE: SCD-R04 selectors 1,3,4,5 | No change: V07 tests mask selection, invalid selector, no-selector, missing target and caller reset. The remaining selectors are fixed bindings in the same mapping rather than separate result/control-flow recipes; no exhaustive fixed-binding fixtures added |
| COVERAGE: SCD-R05a negative or >=91 start selector | Spec edit: V17(d)/(e) require return 0 for -1 and 91, with no callback/charge/purge; this distinguishes rejected indices from null-slot success 1 |
| COVERAGE: SCD-R06 nonforced lookup, nonpositive level, other kinds | Spec edit: V12(f) proves fallback from current id 21 to owned id 20; V18(b) proves other-kind bypass with absent skills, failed activation and absent room. The already explicit nonpositive ownership-level rejection retains the same zero/no-action result; its producer remains a supplied skill-level contract |
| COVERAGE: SCD-R07 zero-coordinate fallback and ordinary creator coordinates | Spec edit: V13(b) requires four zero coordinate arguments after failed/zero-coordinate resolution. V09(a) now names the ordinary creator; V13(a) distinguishes nonzero displacement/extrapolation. The same coordinate arithmetic does not need a duplicate fixture for each creator binding |
| COVERAGE: SCD-R08 Sequence nonzero-event suppression and other kinds | Spec edit: V18(a)/(c) distinguish Sequence suppression from non-Sequence application while requiring evaluation in both; V18(b) requires other-kind resource skipping and delay evaluation |
| COVERAGE: SCD-R09a no current skill | Spec edit: V14(b) requires immediate return 2/completion without argument write or predicate/action/scheduling; the owned skill node's conditions remain unchanged |
| COVERAGE: SCD-R11 negative/overflowing frame arithmetic | Spec edit: V15(b) gives frame -10 → numerator -3 → quotient 0 → due 1; V15(c) gives frame 2147483647 → wrapped numerator -2147483642 → quotient -268435455 → due -2147483639. Actual event storage with unusual due frames remains SCD-Q10 |
| COVERAGE: SCD-R13 failed creator without seed step | Spec edit: V13 supplies a failed, draw-free creator and explicitly asserts unchanged game seed {1,0x29A}. This discriminates direct dispatcher draws without asserting complete creator eligibility/draw behavior; that integration remains gated |
| COVERAGE: SCD-R14a in-range null/index 0 | No change: V16 distinguishes direct invocation from the full dispatcher and tests the no-call outcome at invalid index 191. The known null slot 0 uses the same no-callback/no-effect outcome; exhaustive table-slot fixtures are advisory |
| COVERAGE: repeated SCD-R05a/SCD-R11 start/record bounds | No change beyond V17(d)/(e): start-index rejection is now discriminated. Record bounds are explicit and use GDF-R08a; repeating the same invalid-record/no-scheduling outcome does not establish a new mechanism |

### Existing gate outcomes

| Finding | Outcome |
|---|---|
| GATED: SCD-R09c | No change: retain unresolved animation production under SCD-Q01 |
| GATED: SCD-R10 | No change: retain asm-only caller eligibility/root boundary under SCD-Q04 |
| GATED: SCD-R14b | No change: retain asm-only progression/aura eligibility under SCD-Q09 |
| GATED: SCD-R15 | No change: retain asm-only networking/range/mode boundary under SCD-Q02/SCD-Q07 |
| GATED: SCD-Q01 | No change: actual animation and special predicate effects remain integration gaps |
| GATED: SCD-Q02 | No change: command identities/slot/stat/range ownership remains outside the supplied dispatch boundary |
| GATED: SCD-Q03 | No change: actual target/party decisions and mutations remain gated |
| GATED: SCD-Q04 | No change: immediate/other caller eligibility and roots remain gated |
| GATED: SCD-Q05 | No change: malformed accesses and authenticated original-profile compiled prevalence remain open |
| GATED: SCD-Q06 | No change: actual activation/charged-mana integration remains open |
| GATED: SCD-Q07 | No change: supplied mode-entry behavior does not establish command-selected mode |
| GATED: SCD-Q09 | No change: repeated gameplay/progression/aura integration remains open |
| GATED: SCD-Q10 | No change: supplied operations do not verify actual dependency effects/draws |
| GATED: SCD-Q12 | No change: pending gameplay observations limit runtime confirmation, not verified local supplied stages |

### Hardest-rule outcomes and source recheck

| Finding | Outcome |
|---|---|
| HARDEST: SCD-R02 ordered validation, charged override and reset | No change: the ordered pseudocode and V03/V06/V12 explicitly distinguish charged base level, no-reset exits and player-only reset. Actual predicate/level producers remain named gates |
| HARDEST: SCD-R06–SCD-R08 selection, missile override and consuming requests | Spec edit: both control flags are now expressly required to be nonzero in R06/R07; V13(b)/V18 add the missing distinct coordinate/resource outcomes, and invocation level is explicitly separate from the eligibility getter result |
| HARDEST: SCD-R09a–SCD-R11 frames, ignored results and scheduling | Spec edit: V09 now reaches the actual frame-action controls; V14(b) covers absent current skill, and V15(b)/(c) distinguish signed/truncated/wrapped scheduler arithmetic. The real animation/repeat producers remain gated |

A source recheck on 2026-10-15 first checked available byte-matching coverage for D2Game 0x9FD00; no byte-matching view was available. Its instructions and the already corroborated original-image span show two room getter calls: initial absence exits before activation and without skill/mode writes; with InTown clear a second lookup occurs after activation, and its absent/town result clears skill and requests player reset. R03 now states these separate lookups rather than describing one room value with two conflicting failure effects. V17 supplies per-call results to distinguish the branches; it makes no assertion that the two results differ in actual gameplay. Actual room lookup execution remains SCD-Q10.

The resulting reader-1 triage contract has 18 vector identities. Original numerical outcomes are retained; added cases/inputs/request assertions and the room-lookup clarification required the second fresh full reader recorded below.

## Full blind-implementer reader 2 and triage

The second fresh full reader recomputed all 18 vector identities/cases using the spec and cited dependencies without binary/source/runtime access. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 14, COVERAGE 12, HARDEST 3. There are no mismatch/guess/gap/discrimination entries to resolve. The subsequent entry-order source correction below changes the reviewed contract, so these zero counts are not final review completion; a third fresh full reader is required.

| Finding | Outcome |
|---|---|
| COVERAGE: SCD-R01 unit-target entry and fatal null target | Spec edit: V09(c)/(d) now assert both exact unit-target entry orders, stored-point preservation, pre-assert entry/mode requests, no numeric fatal return and no later work; source basis below |
| COVERAGE: SCD-R02 missing current skill, invalid identity, party rejection | No change: explicit ordered exits/reset semantics remain; the no-reset rejection result is exercised by V12(a), with reset-producing failures distinguished by V03/V06/V12(e). These additional rejection predicates have no unspecified local decision |
| COVERAGE: SCD-R03 missing game/unit or invalid identity | No change: explicit initial rejection precedes activation, writes no current-skill/mode state, and has the zero/no-reset/no-callback recipe discriminated by V17(a). Exhaustive absent-object/record-binding fixtures remain advisory |
| COVERAGE: SCD-R04 selectors 1,3,4,5 | No change: fixed mask bindings share the V07 selected-mask/collision path; the existing cases discriminate absent/invalid selector, missing target and caller reset without one fixture per fixed binding |
| COVERAGE: SCD-R06 missing/nonpositive owned skill and town reset | No change: R06 explicitly orders lookup/level/town decisions; V12(f) discriminates fallback lookup and V18(b) other-kind bypass. Zero/no-action and clear-skill/player-reset recipes are already represented by V12 and V06; actual level/room producers remain SCD-Q10 |
| COVERAGE: SCD-R07 target point sourced from a unit target | No change: dispatcher consumes a supplied resolved point, whose displacement/extrapolation/failure results are discriminated by V13. Actual unit-target-to-point resolution is the SCD-Q10 dependency, not a new local coordinate formula |
| COVERAGE: SCD-R10 no executable vector | No change: asm-only actual caller eligibility is already gated by SCD-Q04; no guessed executable fixture added |
| COVERAGE: SCD-R11 invalid skill identity/record | No change: explicit zero/no-evaluation/no-replacement result consumes GDF-R08a record validity; V05(d) already discriminates the no-scheduling recipe. No duplicate record-bound fixtures added |
| COVERAGE: SCD-R14a in-range null slot | No change: known original slot 0 is null; its no-call/no-effect result is the same result recipe in V16(b). Populated call versus full dispatch is distinguished by V16(a) |
| COVERAGE: SCD-R14b no executable vector | No change: progression/active-state eligibility remains asm-only under SCD-Q09; no guessed vector added |
| COVERAGE: SCD-R15 no executable vector | No change: networking/range/mode execution remains asm-only under SCD-Q02/SCD-Q07; no guessed vector added |
| COVERAGE: SCD-R02 other target/party combinations | No change: absent/self target, TargetAlly and supplied party result have explicit boolean gates in R02; actual predicate/target mutation remains SCD-Q03. No additional distinct local state-change recipe is missing |

GATED dispositions are unchanged for each of the same 14 entries: SCD-R09c (retain SCD-Q01), SCD-R10 (retain SCD-Q04), SCD-R14b (retain SCD-Q09), SCD-R15 (retain SCD-Q02/SCD-Q07), SCD-Q01 (animation integration), SCD-Q02 (network owner), SCD-Q03 (target integration), SCD-Q04 (caller eligibility), SCD-Q05 (malformed/profile boundary), SCD-Q06 (activation), SCD-Q07 (command-selected mode), SCD-Q09 (repeat/aura/progression), SCD-Q10 (actual operation effects), SCD-Q12 (runtime confirmation). Each keeps the corresponding outcome and blocked boundary from the reader-1 gate table and current gap file; no supplied-input assertion is newly gated by those references.

| Finding | Outcome |
|---|---|
| HARDEST: SCD-R02/SCD-R03/SCD-R05a charged level, two room calls, reset order | No change: the explicit ordered rules and V03/V06/V12/V17 already supply these decisions; actual producers retain their named gates |
| HARDEST: SCD-R06–SCD-R08 controls, missile override and resources | No change: both controls, invocation/getter levels, request result override and consuming effects are explicit and distinguished by V08/V12/V13/V18 |
| HARDEST: SCD-R09a–SCD-R11 latch and signed scheduling | No change to frame/scheduler behavior: V09/V14/V15 explicitly distinguish those choices. R01 entry order was corrected separately from new source evidence below |

### Entry-order source correction after reader 2

On 2026-10-15 both D2Game entry handlers were checked for available byte-matching coverage; neither 0x787C0 nor 0x78840 had a byte-matching view, so their complete 128-byte/112-byte plain instruction bodies were read. D2Game 0xDA750 was also checked and inspected; its encoded spans/imported interior remain unresolved, so it is represented only as the neutral entry-preparation request (game,unit), with actual effects gated by SCD-Q10.

The exact unit-target chain is 0x787C0 → 0xDA750 → D2Game import thunk 0xA4A2 (D2Common #11090, RVA 0x33920) → null-target test → D2Game thunk 0xACD0 (D2Common #11129, RVA 0x30E80) → 0xDFF10 → 0xE0750 → 0xDF7B0 → latch clear → 0xA0830. A null target takes assertion line 0x4BE and the fatal path after entry preparation and the mode request, before target setting, remaining preparation, latch clearing or validation. These are source-derived requests; no fault was executed.

The exact point-target chain differs: 0x78840 first calls D2Game thunk 0xAB2C (D2Common #10286, RVA 0x349D0), then 0xDA750, then mode-setting thunk 0xA4A2, then unit-target thunk 0xACD0 with null, followed by 0xDFF10 → 0xE0750 → 0xDF7B0 → latch clear → 0xA0830. Both new import targets were resolved against the original module exports, checked for available matching coverage and read through C reconstructions equivalent to their original machine bytes.

For valid inputs, #10286 stores the point coordinates and clears unit-target selection. #11129 stores selected target unit and, for a non-null target, its kind/identity; it does not clear stored point coordinates. **Corrections:** the earlier R01 list omitted the entry-preparation request, placed the mode request before point selection, and described clearing the other target uniformly on both routes. The caller chains and target setters establish the route-specific order and point preservation now stated. No broader cleanup meaning or full preparation effect is inferred. V09(a)/(b) keep their numerical results with corrected complete entry logs; V09(c)/(d) add nonfatal/fatal unit-target subcases without a new vector identity. That behavior/order change required fresh full reader 3, recorded below.

## Full blind-implementer reader 3 and triage

The third fresh full reader returned MISMATCH 1, GUESS 0, GAP 1, DISCRIM 0, GATED 35, COVERAGE 13 and HARDEST 3. These counts belong to the entry-order revision it read. The composition corrections and added subcases below change the reviewed contract, so a fourth fresh full reader is required. No review completion or user approval is inferred from earlier zero counts.

| Finding | Outcome |
|---|---|
| MISMATCH: SCD-V09(c) final player mode 10 after validator result 0 | Spec edit: V09 now executes the actual local validator. Its unit-target failure case supplies start callback return 0 and follows R02's reset(mode=5,force=1),so final mode is corrected to 5. The earlier mode-10 expectation came from a no-effect validator substitute and does not follow R02. This correction is supported by the already verified validator/reset branch,not a runtime observation |
| GAP: V09(a–c) replaced local R02 with a validator return/no-effects stub absent from section 3's dependency contract | Spec edit: all reached local rules now execute their stated behavior. V09(a)/(b) execute R02/R03/R05a successfully and log the actual start mana request. Their explicit initial mana is 2816,so one supplied 256 deduction yields the existing final 2560; validator return is 1 and mode stays 10. V09(c) begins with 2560 and executes start failure/reset with no deduction. V09(d) still terminates before validation. Outside-boundary preparation,level,predicate,resource and per-skill callback operations retain their explicit supplied contracts |

The earlier fixture triage did not enforce local-rule composition. R01/R12 and the vector preamble now explicitly retain effects of called local stages. A source-supported zero result from the validator is not a no-effect operation: it can reset a player. The analogous action-dispatch substitutes in V05/V14 were also removed. Both now execute the local action dispatcher with valid null slot 160 and no missile,compute return 0 before callback/resource/delay work,and assert that exact return before continued periodic scheduling. Their existing due frames and state outputs are unchanged. No new evidence label or actual dependency effect is promoted.

### Advisory coverage outcomes

| Finding | Outcome |
|---|---|
| COVERAGE: R02 absent skill/invalid identity/party rejection/self-target/TargetAlly | No additional cases: these ordered predicates and skip conditions are explicit; zero/no-reset rejection is represented by V12(a),reset-producing failures by V03/V06/V12(e),and V09(c) now executes the non-self/TargetAlly-clear party check with supplied result 0. Actual predicate production/mutation remains SCD-Q03 |
| COVERAGE: R03 absent game/unit/invalid identity | No change: the initial zero/no-reset/no-callback recipe is explicit and distinguished by V17(a); these extra object/record bindings add no new local state-change recipe |
| COVERAGE: R04 selectors 1,3,4,5 | No change: additional fixed mask bindings share V07's selected-mask/collision path; absent/invalid selector,missing target and caller reset are already distinguished |
| COVERAGE: R06 absent game/unit/invalid identity/town/missing or nonpositive owned skill/valid null action | No new cases: R06 gives ordered rejection/reset decisions; V12(f) tests lookup and V18(b) bypass. V01(b),V05 and V14(a) execute valid null slot 160 and now assert actual zero return/no callback/resource/delay. Other listed failures have the explicit zero/no-action or clear-skill/reset result recipes already represented by V12/V06 |
| COVERAGE: R07 absent missile record/coordinate overflow | Spec edit: V13(c) supplies i32 extrema and distinguishes displacement wrap to -1 and destination wrap to -2. This is the arithmetic consumer boundary,not proof that actual coordinate producers permit those inputs. Missing record takes the explicitly stated no-missile branch; no duplicate fixed-record fixture added |
| COVERAGE: R08 nonpositive delay/player versus monster consumption | Spec edit: V18(d) supplies monster kind 1 and asserts mana/quantity requests plus delay evaluation,but no player-only delay application. Its explicitly supplied resource operations only log,so resources remain unchanged. Delay zero already reaches the nonpositive branch in existing vectors; a negative result has the same no-application recipe |
| COVERAGE: R09a condition A set with frame predicate !=2 | Spec edit: V14(c) supplies predicate 1 with code 7/latch set and requires no action/scheduler/B write while retaining argument storage and completion |
| COVERAGE: R09b direct frame-action entry without current skill | Spec edit: V14(d) enters R09b directly and requires no action,scheduler or completion request; this distinguishes it from R09a's missing-skill completion route |
| COVERAGE: R11 invalid identity/record or negative period | No change: invalid record has the explicit zero/no-evaluation/no-replacement recipe; V05(d) already distinguishes no scheduling. Signed period clamping is tested by V05(a)'s value 3→5; negative values select that same clamp recipe |
| COVERAGE: R14a valid null slot | No change: known null index 0 has the no-call/no-effect recipe in V16(b),with actual populated invocation distinguished by V16(a). Exhaustive table bindings remain advisory |
| COVERAGE: R14b no executable vectors | No change: actual progression/active-state eligibility remains asm-only and SCD-Q09-gated; no guessed fixture added |
| COVERAGE: R15 no executable vectors | No change: actual command/range/mode behavior remains asm-only and SCD-Q02/SCD-Q07-gated; no guessed fixture added |
| COVERAGE: R05a/R12/R13 no additional branch finding | No additional coverage request to resolve. The actual-start and actual-zero-dispatch assertions above strengthen composition without changing these rules' evidence labels |

### Boundary and gate outcomes

The reader's 35 GATED entries include 18 supplied-input vector boundaries and two pending observations. They are not 35 unresolved behavioral rules or blanket gates on local verified assertions. Each disposition is recorded here against the existing exact blocked boundary.

| Finding | Outcome |
|---|---|
| GATED: SCD-Q01 | Retain actual animation/special predicate integration gate |
| GATED: SCD-Q02 | Retain networking owner/command-range boundary |
| GATED: SCD-Q03 | Retain actual target/party producer and mutation gate |
| GATED: SCD-Q04 | Retain immediate/other caller eligibility/root gate |
| GATED: SCD-Q05 | Retain malformed access/authenticated compiled-profile gap |
| GATED: SCD-Q06 | Retain actual activation/charged-mana integration gate |
| GATED: SCD-Q07 | Retain command-selected mode outside-boundary question |
| GATED: SCD-Q09 | Retain repeat/progression/aura integration gate |
| GATED: SCD-Q10 | Retain actual outside-boundary operation effects/draws only; local stage composition now executes the stated rules |
| GATED: SCD-Q11 | Retain outside-boundary handler return catalog; supplied nonboolean propagation remains verified |
| GATED: SCD-Q12 | Retain pending gameplay confirmation without adding a local implementation gate |
| GATED: SCD-R09c | Retain unresolved animation production under Q01 |
| GATED: SCD-R10 | Retain asm-only caller eligibility under Q04 |
| GATED: SCD-R14b | Retain asm-only progression/active-state conditions under Q09 |
| GATED: SCD-R15 | Retain asm-only command/range/mode integration under Q02/Q07 |
| GATED: SCD-O01 | Retain pending animation observation; no capture claimed |
| GATED: SCD-O02 | Retain pending periodic observation; no capture claimed |
| GATED: SCD-V01 | Retain supplied ownership/record inputs; local action stages execute |
| GATED: SCD-V02 | Retain supplied resource effects; start/action composition executes |
| GATED: SCD-V03 | Retain supplied predicates/reset operation; actual validator failure executes |
| GATED: SCD-V04 | Retain supplied activation/resources; actual start/action controls execute |
| GATED: SCD-V05 | Retain supplied level/expression/events; remove action-dispatch substitute and execute local zero branch |
| GATED: SCD-V06 | Retain supplied town/predicate/reset inputs; actual validator/start composition executes |
| GATED: SCD-V07 | Retain supplied collision/resolution; actual start/validator decisions execute |
| GATED: SCD-V08 | Retain supplied skill callbacks; local conditional selection executes |
| GATED: SCD-V09 | Retain only outside-boundary operation/callback inputs; remove validator substitute and execute local validation/start/frame/action stages |
| GATED: SCD-V10 | Retain successful creation-seeding entry and no-later-draw boundary; no whole-creator parity claimed |
| GATED: SCD-V11 | Retain supplied per-skill callback/resource/event operations; local nonboolean propagation executes |
| GATED: SCD-V12 | Retain declared predicate/level inputs; local validation/control decisions execute |
| GATED: SCD-V13 | Retain failed creator/resource/coordinate inputs; arithmetic and requests remain verified even for synthetic extreme coordinates |
| GATED: SCD-V14 | Retain special predicate/completion/event inputs; remove action-dispatch substitute and execute its local zero branch |
| GATED: SCD-V15 | Retain supplied expression/event storage; signed/wrapped due-frame arithmetic remains verified |
| GATED: SCD-V16 | Retain isolated direct-table boundary; no local gate added to that operation |
| GATED: SCD-V17 | Retain supplied per-call room/activation results; actual local ordering/bounds execute |
| GATED: SCD-V18 | Retain explicitly declared activation/resource/expression operations; local unit-kind/Sequence decisions execute |

| Finding | Outcome |
|---|---|
| HARDEST: V09/R01–R02 composition | Spec edit: execute local validator/start and retain all writes,including mode 5 reset on failure; make successful start charge and initial mana explicit |
| HARDEST: R06 gates | No behavior change: ordered conditions remain explicit; V05/V14 now reach the real null-slot branch,and V18(d) adds the distinct monster request recipe |
| HARDEST: R11/V15 division,wrap,event removal | No change: V15's exact signed/wrapped intermediate values and retained sentinel/ordinary events remain explicit; actual event storage stays gated |

Self-review after this triage recomputed V09's start deduction (2816-256=2560),failed-start reset (mode 5 without deduction),and fatal pre-validator branch; V05/V14's empty-slot zero branches; V13(c)'s i32 coordinate wraps; and V18(d)'s request-only supplied resource effects. The spec retains 18 vector identities. Changed composition,inputs,subcases and corrected V09(c) output require another fresh full reader.

## Full blind-implementer reader 4 and triage

The fourth fresh full reader recomputed all 18 vector families and cases. It returned MISMATCH 1, GUESS 0, GAP 1, DISCRIM 0, GATED 16 grouped entries, COVERAGE 13 and HARDEST 3. Its gate count groups the 11 questions, four gated rules and all supplied-stage vector limits; it is not a count of individually unverified assertions. The following corrections require fresh full reader 5: reader 4 did not return a zero mismatch count, so the delta-check stopping condition was not met.

| Finding | Outcome |
|---|---|
| MISMATCH: V09(a) claimed a created missile with no seed step | Spec edit: its declared successful creation-seeding entry now consumes NRF-R17a, with initial game seed {1,666}, one game step to {0x6AC6935F,0}, and supplied missile seed {0x6AC6935F,666}. No later initialization draws are supplied. The prior draw-free creator fixture did not exercise the cited successful-seeding contract. Mana 2560, mode 10, one missile, latch true and action counts remain unchanged. The exact producer-entry eligibility and full initialization remain Q10; this is not a gameplay observation |
| GAP: section 5 omitted return outputs | Spec edit: the output inventory now names start/action/validator, frame-handler and scheduler returns and their owning rules. Those returns were already explicit in the rules and vectors; no new return behavior is inferred |
| GATED: Q01–Q07 and Q09–Q12 | No change: retain each existing question and its exact blocked producer/integration/observation boundary in the gap file |
| GATED: R09c/R10/R14b/R15 | No change: retain the documented animation, caller, progression/state and command boundaries |
| GATED: V01–V18 supplied stages | No new blanket gate: each fixture retains its stated execution boundary; V09(a) now explicitly consumes the successful seeding entry, matching V10's existing boundary |

| Coverage entry | Outcome |
|---|---|
| R02 absent skill, invalid identity, party rejection and TargetAlly | No change: explicit zero/no-reset exits and bypass are in R02. V12(a) distinguishes no-reset rejection; V03/V06/V12(e) distinguish reset failures; V09(c) executes the non-self party check. No unshown result recipe is added by another fixed eligibility binding |
| R03 absent objects/first room, forced activation and second town room | No change: V17(a) tests first-room rejection; V12(c) forced activation bypass; V17(b) the second-lookup clear/reset recipe; V06 the town distinction. Missing objects and the second-town result do not introduce a different output recipe |
| R04 remaining mask bindings | No change: V07 distinguishes selected mask, absent selector/target, invalid selector and caller reset; the remaining masks are fixed bindings on that same path |
| R06 missing/nonpositive owned skill, town, negative index and other kinds 3–5 | No change: ordered rejection/reset and bounds are explicit. V12(f) tests lookup fallback; V18(b) bypasses those checks for another kind; V01/V05/V14 distinguish valid null-slot zero from invocation. Extra fixed kind/invalid-index examples add no new result recipe |
| R07 missing missile record and both zero axes | No change: missing record is the existing no-missile branch; V13(b) requires four zeros after failed/zero-coordinate resolution, while V13(a)/(c) distinguish valid and wrapped coordinates |
| R08 malformed start read and other positive-delay branches | No change: malformed reads remain Q05 with no invented safe result. V18(a)/(c)/(d) distinguish Sequence suppression, ordinary player application and monster request-only behavior; V13 supplies ordered positive delay |
| R09a actual finishing-predicate conditions | No change: real predicate execution remains Q01/Q10; its supplied true/false results and missing-skill shortcut are distinguished by V09/V14 |
| R11 invalid record and separate periodic/aura flags | No change: invalid identity/record has the explicit zero/no-evaluation recipe. V05 is periodic-only and V15 aura-only; V05(d) tests neither. These flag branches are already exercised |
| R14a negative index and null slot | No change: R14a gives the signed bounds and null condition. V16 already distinguishes invocation from no-call/no-effects; further fixed rejected/null indices use that result recipe |
| R10 other callers | No change: actual roots and eligibility remain asm-only/Q04, with no guessed executable fixture |
| R14b progression/state | No change: Q09 retains eligibility/state integration; no guessed vector |
| R15 command target check | No change: Q02/Q07 retain networking/range/mode boundaries; no guessed vector |
| R09c real event production | No change: Q01 retains production; supplied event decisions are already distinguished |

| Hardest entry | Outcome |
|---|---|
| R01/V09 entry order, fatal target and reset persistence | V09's seeding fixture corrected as above; route-specific logs, fatal boundary and reset-to-5 assertions remain unchanged |
| R06 with V08/V12/V18 control/ownership/unit kinds | No change: ordered rules and exact callback/resource logs already separate these decisions; actual producers retain named gates |
| R09a/R09b with V14 exceptional route and ignored zero | No change: V14 executes the local null-slot action dispatcher and requires scheduling after its computed zero result; no same-spec no-effect substitute remains |

Root self-review recomputed the one creation step from game {1,666}: low=0x6AC690C5+666=0x6AC6935F, high=0; NRF-R17a initializes missile {low,666}. Cases V09(b–d) create no missile and make no seed step. Section 5 now inventories the existing numeric returns. No rule behavior, label or gate was promoted by this correction; the added fixture seed assertion still requires the fresh full review recorded next.

## Full blind-implementer reader 5 and triage

The fifth fresh full reader returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 5, GATED 29, COVERAGE 10 and HARDEST 3. Every finding is assessed below. The five discriminating changes affect fixtures only; rules, ownership, supplied-operation contracts and evidence gates are unchanged. These changes are ready for a fresh delta reader; no delta-reader result is claimed here.

| Finding | Outcome |
|---|---|
| MISMATCH: none | No rule or expected numeric result correction requested |
| GUESS: none | No inferred input/default to resolve |
| GAP: none | No contract omission to resolve |
| DISCRIM: V05 perdelay arguments | Spec edit: V05(a–c) each requires exactly one eval perdelay(unit,id=20,level=3),with its declared result 3 or 8; the due frames and ignored local action result remain unchanged |
| DISCRIM: V15 perdelay arguments | Spec edit: V15(a–c) each requires exactly one eval perdelay(unit,id=20,level=3) returning 8 before sentinel removal/scheduling. The special event payload (-1,0) must not replace those evaluation arguments |
| DISCRIM: V16(a) callback arguments | Spec edit: populated slot 10 must receive (game,unit,20,3) exactly once; callback invocation alone is insufficient |
| DISCRIM: V09(a) ordinary creator coordinates | Spec edit: the sole ordinary creator receives displacement=(0,0),destination=(0,0),because its actual frame-action controls are forced=0,itemTarget=0 despite stored point (30,40). Successful seeding,one missile,mana 2560,mode 10,latch and call counts remain unchanged |
| DISCRIM: V08 original selector bound before ItemEffect | Spec edit: new V08(e) supplies srvdofunc=191,forced=1,itemTarget=1,ItemEffect=3 and requires return 0 before replacement,with no callback,missile,resource or delay work; mana 2560 and current skill remain unchanged. This valid replacement value does not resolve malformed replacement accesses under Q05 |

| Gated entry | Outcome |
|---|---|
| SCD-Q01 | Retain actual animation/special predicate integration boundary |
| SCD-Q02 | Retain networking command/range boundary |
| SCD-Q03 | Retain actual target/party producer and mutation boundary |
| SCD-Q04 | Retain other-caller roots/eligibility boundary |
| SCD-Q05 | Retain malformed access and authenticated compiled-profile gap; V08(e) tests the already verified original-selector bound only |
| SCD-Q06 | Retain actual activation/charged-mana integration boundary |
| SCD-Q07 | Retain command-selected mode boundary |
| SCD-Q09 | Retain aura/periodic/progression eligibility/state integration boundary |
| SCD-Q10 | Retain actual outside-boundary operation effects/draws; same-spec local rules still execute |
| SCD-Q11 | Retain per-skill return catalog boundary; supplied propagation remains verified |
| SCD-Q12 | Retain pending gameplay confirmation; no observation claimed |
| SCD-V01 | Retain supplied ownership/record boundary |
| SCD-V02 | Retain supplied resource-operation effects |
| SCD-V03 | Retain supplied predicates/reset-operation effects |
| SCD-V04 | Retain supplied activation/resource effects |
| SCD-V05 | Retain supplied level/expression/event operations; exact expression arguments now asserted |
| SCD-V06 | Retain supplied town/predicate/reset inputs |
| SCD-V07 | Retain supplied collision/resolution results |
| SCD-V08 | Retain supplied per-skill callbacks; new original-selector bound case remains local |
| SCD-V09 | Retain supplied outside-boundary operations and successful seeding entry; local validation/start/frame/action composition remains executed |
| SCD-V10 | Retain successful seeding entry/no-later-draw boundary |
| SCD-V11 | Retain supplied callback/resource/event effects |
| SCD-V12 | Retain supplied predicate/level results |
| SCD-V13 | Retain supplied failed creator/resource/coordinate inputs |
| SCD-V14 | Retain supplied special predicate/completion/event inputs; local zero action result still computed |
| SCD-V15 | Retain supplied expression/event storage; exact evaluation arguments now asserted |
| SCD-V16 | Retain isolated direct-table boundary; exact callback arguments now asserted |
| SCD-V17 | Retain supplied per-call room/activation results |
| SCD-V18 | Retain explicitly supplied activation/resource/expression operations |

| Coverage entry | Outcome |
|---|---|
| R02 absent skill/invalid identity/party rejection/self/TargetAlly | No additional fixture: R02 explicitly orders these exits and bypasses. V12(a) represents zero without reset; V03/V06/V12(e) represent reset failures; V09(c) executes the non-self party check. Further fixed predicate bindings add no unshown result recipe |
| R03 absent game/unit/invalid identity/second town result | No additional fixture: initial zero/no-reset is distinguished by V17(a),and second-room clear/reset by V17(b),with town behavior in V06. The listed extra bindings share those explicit output recipes |
| R04 remaining mask bindings | No additional fixture: V07 already distinguishes selected-mask collision,invalid selector,no resolved target and caller reset; the remaining masks bind the same result recipe |
| R06 town/missing owned skill/nonpositive owned level | No additional fixture: ordered rejection and clear/reset recipes are explicit; V12(f) exercises owned lookup,V06 clear/reset,and V18(b) the other-kind bypass. Extra failure bindings introduce no different result recipe |
| R07 ordinary four-zero creator coordinates | Resolved by V09(a)'s exact zero displacement/destination assertion despite a stored nonzero point |
| R09a ordinary code other than 1/2 | No additional fixture: V14(c) already requires no action/scheduler when predicate=1,code=7 and latch is set. Exhaustive ordinary-code bindings are advisory; R09a explicitly requires code 1/2 on its ordinary route |
| R09b nonordinary repeat kind | Not a reachable input to R09b: that rule always calls R11 with ordinary repeat kind. V15 separately enters R11 with special repeat kind; no fictitious frame-action input added |
| R11 invalid record/periodic-only/aura-only | No additional fixture: invalid identity/record has the explicit zero/no-evaluation recipe. V05 already exercises periodic-only,V15 aura-only,and V05(d) neither flag |
| R10 other callers | Retain asm-only/Q04 boundary; no guessed executable fixture |
| R14b/R15 progression/state and command paths | Retain Q09 and Q02/Q07 boundaries respectively; no guessed executable fixture |

| Hardest entry | Outcome |
|---|---|
| V09 composition | Preserve actual local validation/start/frame/action execution,successful seeding and route-specific reset/fatal behavior; add only exact ordinary-creator coordinates |
| V13 wrap/missile override/resource order | No change: signed wrapped coordinate intermediates,failed creator,forced result 1,ordered consuming requests and unchanged seed remain explicit |
| V15 signed scheduler | Preserve signed division,overflow and sentinel event replacement; add exact expression inputs so event payload cannot be substituted for evaluation identity/level |

Self-review checked all five fixture edits against the unchanged local rules and retained every prior numeric result. The spec still has 18 vector identities; V08(e) is an additional subcase. The gap file is unchanged because these edits neither close nor expand any evidence gap. No new source investigation,gameplay observation or test was performed for this triage.

## Delta blind-implementer reader 6

2026-10-15: a fresh smaller reader checked the actual contract, numeric foundations,
game-data fields and the cited expression front-end contract. It read no companion,
prior reader report, private source, implementation, runtime or web material and encountered
no file-access limitation. This was a delta check, qualified by full pass 5's zero mismatch
count and the complete triage above. Changed material was V05(a–c), V15(a–c), V16(a),
V09(a) and V08(e); no behavior, ownership or gate boundary changed after full pass 5.

Counts: **MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0**. Outcome for each list: no unresolved
entry. The reader independently recomputed every changed case, including due frames
106/105/97, the signed/wrapped special-repeat values 105/1/-2147483639, the one successful
creation seed step and the invalid original-selector early exit. It confirmed the exact
evaluation/callback arguments, four zero ordinary-creator coordinates, unchanged resource
values and no-work assertions. Each incorrect alternative identified in full pass 5 now
fails a stated assertion. Existing gates and earlier advisory outcomes remain unchanged.
This completes the applicable reader check for the current supplied-stage contract; it
establishes neither user approval, gameplay confirmation nor full behavior reproduction.

## Spec-review corrections, 2026-10-15

An independent review re-derived V01 through V18 from the rules (due frames, wrapped coordinates, seed steps and resource values all agreed) and re-read D2Game start dispatch `0x9FD00`, scheduler `0x9F3B0`, action dispatch `0xA19A0` and target-coordinate resolver `0x9FA30`. Start preconditions, line masks, start bound 91/null-slot return 1, scheduler clamp/division/payloads, ItemEffect replacement and the consuming-request order agree with R03, R04, R05a, R06, R08 and R11. Two discrepancies were found and corrected:

- **R07 partial coordinates.** The resolver `0x9FA30` always writes both coordinate outputs (target unit position, or the stored target point when the target is absent or the unit itself) and returns 1 only when both are nonzero. The action dispatcher's missile branch (`0xA1B3A`..`0xA1BD0`) ignores that return. It loads x, then y, and only when both are nonzero replaces them with displacement `(tx-ux,ty-uy)` and computes the destination. When either is zero it jumps to the creator call with the raw x/y still in the displacement registers and zero destination. The previous R07 stated four zeros for every non-both-nonzero case, and V13(b) expected `(0,0),(0,0)` for supplied `(0,40)`. The original bytes give displacement `(0,40)`, destination `(0,0)`. R07 now states the raw-pair path, V13(b) is corrected and V13(d) adds the `(30,0)` case. A fully zero resolution still passes four zeros, and the no-control path makes no resolution request.
- **R06 owned-skill fallback.** At `0xA1A7E`..`0xA1AC5`, a matching current skill whose bonus-inclusive level is <=0 does not return 0 directly; control falls to the owned-skill search and a second level request. R06 now states that fall-through. No vector previously exercised it.

Both edits change rule behavior and a vector expectation, so the reader-6 completion above no longer covers the resulting contract; a fresh full reader pass is required and has not run.
