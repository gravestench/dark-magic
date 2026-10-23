# Evidence: Missile Creation and Seeding

Behavioral contract: [Missile Creation and Seeding](../../specs/missile-creation-and-seeding.md).
Open evidence: [gap file](../../plans/gaps/missile-creation-and-seeding.md).
Sibling contract for the same creator's later stages: [Missile Initialization and Registry evidence](missile-initialization-and-registry.md).

## Binary identities and methods

Evidence build is original Lord of Destruction Expansion 1.13c Windows x86. Function references below resolve through [source manifest](../source-manifest.md) artifacts `LOD113C-D2GAME` and `LOD113C-D2COMMON`. Their original images were rehashed on 2026-10-15 and matched the manifest SHA-256/size identities. Artifact identity establishes applicability of original-image locations, not historical loaded-module or archive identity. All function descriptions below are analyst descriptions, not original exported names. Export ordinals identify the matching D2Common artifact.

The 2026-10-12 record read the creator and its listed setup helpers by disassembly, with selected reconstruction checked against original bytes for small helpers. That compatible reading remains support for its recorded arithmetic/control-flow slices. New analysis on 2026-10-15 first checked whether reconstruction checked against original bytes was available; the adjacent-room getter was read in that form. The newly inspected creator, distance and synchronization slices had no matching reconstruction and were read as disassembly. Encoded instruction runs in the distance and selected caller slices were decoded independently from the fingerprint-matched original image; a decoding check does not close an unread callee or producer.

## Dispatch and function references

Root route retained from 2026-10-12: D2Game skill server do-function table at RVA `0x1074A8`, entry 1, selects basic attack `0xA2C40`; its ranged route calls positioned-missile request `0xA28F0`, which calls creator `0x6F930`. The caller zeroes a 0x5C-byte request. Existing [skill dispatch evidence](skill-cast-start-and-do-dispatch.md) and [PvP evidence](combat-pvp-damage-and-kill-credit.md) support their own dispatch/root boundaries; this record does not reverify every skill/object/AI caller or the complete server scheduler. The historical discovery found approximately 100 creator references and 23 skill do-table entries through one intermediate; those are discovery counts, not completeness or acceptance claims.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0x6F930` | none | One-missile creator | MCI-R01..R13; later stages MIR-R01a..R17b | Disassembly 2026-10-12; targeted full creator reread 2026-10-15 for local order, flags, requests, refusals and widths; dependency effects separately gated |
| LOD113C-D2GAME | `0xA2C40`, `0xA28F0` | none | Attack route and positioned request | Trigger; MCI-R01/R02 | Compatible 2026-10-12 dispatch/caller record; current SCD root contract |
| LOD113C-D2GAME | `0x1250`, `0x1100`, `0x1130` | none | Missile record lookup, unit X/Y | MCI-R01/R02/R08 | Compatible 2026-10-12 reconstructions checked against original bytes; dynamic coordinate contract NRF-R10b; static-source result remains supplied at the caller boundary |
| LOD113C-D2GAME | `0x1280` | none | Containing room selection | MCI-R03 | Full disassembly 2026-10-15, owner room before ordered candidates; no producer/sort inferred |
| LOD113C-D2COMMON | `0x2FE10`, `0x3BA20` | 10331, 10383 | Unit room; stored adjacent collection/count | MCI-R03; MIR-R02a | Room read reused 2026-10-12; collection getter reconstruction checked against original bytes 2026-10-15, nullable candidates handled by Game consumer |
| LOD113C-D2COMMON | `0x33CD0`, `0x38520`, `0x39140`, `0x38B00` | 10494, 10871, 10680, none | State test/list lookup/own-list read | MCI-R06 | 2026-10-12 local calls plus current RLI-R03/RLI-R02a/RLI-R02b; own list is distinct from aggregate |
| LOD113C-D2COMMON | `0x2F370`, `0x3C000`, `0x2D2C0` | 11153, 10826, 10459 | Unit allocation, room level, act selection | MCI-R09 | Compatible 2026-10-12 reconstruction checked against original bytes/read; act thresholds 1,40,75,103,109,1024 identified in original constant region; MSD shares room-level getter |
| LOD113C-D2GAME | `0xE0110` | none | Unit creation seeder | MCI-R10/R13 | Current NRF-R03a/NRF-R03b/NRF-R12a, original local game-seed caller retained; no whole-creation budget inferred |
| LOD113C-D2COMMON | `0x36740` | 10614 | Seed initializer | MCI-R10; MIR-R15a | NRF-R03a/b; compatible original reconstruction record |
| LOD113C-D2GAME | `0x6F740`, `0x7B0A0` | none | Local missile sender and synchronization caller | MCI-R12/MCI-Q03 | Sender disassembly; caller original-image decoding 2026-10-15 establishes call at `0x7B348`, not full sync/client/delivery chain |
| LOD113C-D2COMMON | `0x7C8D0` | none | Missiles loader schema | Inputs | Compatible 2026-10-12 descriptor reading; field name/width/bit identity, not complete compiler/default behavior |

Functions reached after identity (metadata, event, placement, registry, setters, planner, distance, base list, damage, attribution, pierce, stat requests and monster event) are listed in the [sibling companion](missile-initialization-and-registry.md#dispatch-and-function-references).

### Import and call placement

Original import/export resolution retained from 2026-10-12 used the identified original PE tables, with two independent parsers agreeing. In the creator's stages owned here: Game thunk `0xA478` resolves Common #10331/`0x2FE10`; `0xA55C`→#10383/`0x3BA20`. The remaining creator thunks are recorded with the stages that call them in the [sibling companion](missile-initialization-and-registry.md#import-and-call-placement). Other resolved targets are named in the function table. Thunks are navigation aids, not interchangeable target identities.

Creator entry takes game in ECX and request in EDX and returns missile-or-null in EAX. Stack cleanup alone was not used to infer argument count.

## Semantic fixture and layout mapping

This maps original observation fields to the behavioral contract; internal layouts are not required implementation architecture.

| Original object / fields | Semantic terms / supporting rules |
|---|---|
| Request +0x00 flags; +0x04 owner; +0x08 source; +0x0C targetUnit; +0x10 missileId | Request routing, MCI-R01/R02/R04/R08 |
| Request +0x14/+0x18 source X/Y; +0x1C/+0x20 target X/Y | Start, retained target point; MCI-R02/R04/R08; MIR-R10 |
| Request +0x28 speed; +0x2C skill; +0x30 level; +0x34 loops; +0x40 elapsed; +0x44 activate; +0x48 toHit; +0x4C range | Semantic scalar inputs in section 3 of both contracts |
| Request +0x54 callback; +0x58 opaque argument | MIR-R12; unused +0x24/+0x38/+0x3C/+0x50 were not read by creator |
| Unit +0 type; +4 class; +8 pool; +0x0C id; +0x10 mode; +0x18 act; +0x1C environment; +0x80 game | MCI-R09/R11 |
| Unit +0x20/+0x24 seed; +0x28 initialLo | MCI-R10 |
| Unit +0xC8 bit 25 expansion; bit 26 condition A; +0xC4 bit 4 condition B; +0xD0 selector; +0x74 auxiliary | MCI-R09/R11; wider roles MCI-Q02 |
| Game +0x1C pool; +0x70 expansion; +0x9C counter; +0xBC+4*act environment; +0xD0/+0xD4 seed | MCI-R09/R10/R11 |
| Room +0x4C/+0x50 origin; +0x54/+0x58 extent; +0 adjacent count; +0x24 adjacent array | MCI-R03; Common getter writes stored array first and count second; no dynamic candidate discovery |

Original Missiles record stride is 0x1A4, record collection/count at DataTables+0xB64/+0xB6C. Loader schema maps flags byte+4 bits 0/2/3/4/5 to LastCollide/Pierce/CanSlow/CanDestroy/ClientSend; +0x96/+0x98 signed Range/LevRange; +0x9A/+0x9B/+0x9C unsigned Vel/VelLev/MaxVel; +0x9E signed Accel; +0x136 Activate; +0x180/+0x181/+0x182 SubLoop/SubStart/SubStop; +0x183/+0x184 CollideType/Collision; +0x19C DamageRate. Light+0x130/RandStart+0x17C/HitShift+0x196 direct-read search limits are below. No shipped row is reproduced.

## Bounded question assessment and new readings, 2026-10-15

Before new analysis each historical question, gated rule, missing executable fixture and pending observation was checked against the original records, source manifest and current NRF/GDF/STL/MMF/SCD/PVP/MSD/MDP contracts. Unchanged old evidence keeps its original dates/methods. Compatibility is tied to manifest artifacts, explicit supplied inputs and recorded data-profile scope; no earlier failed attempt was treated as closure. New targeted reads answer remaining local questions rather than repeat a whole skill/scheduler/gameplay sweep. Rows for questions now owned by the sibling contract are in its [assessment](missile-initialization-and-registry.md#bounded-question-assessment-and-new-readings-2026-10-15).

| Question / family | Compatible evidence and remaining bounded investigation | Result / exact remaining boundary |
|---|---|---|
| Names, historical question | GDF identity conventions and complete 2026-10-15 loader payloads; resolve only the named cells below by header | Named text identities established; no archive/compiled global correspondence asserted |
| Notify MCI-Q03 | Old sender record; checked original caller decode at Game `0x7B348` and local sender eligibility | Call-site byte uncertainty closed. Actual sync root, recipient/cadence/eligibility and network delivery remain unverified; no forced game action added |
| Neutral fields MCI-Q02 | Original direct writes retain exact values/flags; no compatible reader contract gives wider meaning | Keep neutral semantic conditions and wider-reader question. This is not a gate on the established writes |
| Observations MCI-Q04 | Historical 2026-10-12 no reachable member; current authorized work permits table reads, not gameplay | No gameplay vector performed. All observation families retained with tool-neutral procedures |
| Adjacent rooms MCI-Q01 | Adjacent getter reconstruction and creator selector | Stored collection order/consumer selection established; producer of the collection and its order remains outside this boundary. Path allocation and publication are MIR-Q02 |

### Correction rationale and affected claims

The following discrepancies are retained as history; the current contract gives the supported result. Corrections to stages now owned by the sibling contract are in its [correction record](missile-initialization-and-registry.md#correction-rationale-and-affected-claims).

1. The 2026-10-12 aim pseudocode assigned target-unit coordinates back to the target point. Creator `0x6F930` actually loads those coordinates only into temporary gate registers, leaving its retained stack point unchanged. MIR-R10 reloads the retained point. MCI-R04/R08, MIR-R10, MCI-V06 and MIR-V09 now distinguish tracking/gate position from point-duration input. This is source-supported correction, not a fitted gameplay expectation.
2. The original seeding vector, seeder observation and draw-accounting rule could be read as a whole-creation exactly-one-game-draw guarantee with Pierce clear. Current NRF/MDP/Blade evidence requires separating the reached unit-seeder step, local pierce seed, callback, actual path, critical/preparation and monster-mod budgets. The original seed result 0x6AC6935F and every original pierce sequence remain. MCI-O02 now measures the seeder boundary, not an unverified whole-operation budget (MCI-R13, MCI-V07).
3. The original names question requested a particular physical patch archive. Effective Storm file reads establish the loaded archive-set result, not which physical MPQ supplied it. The named-cell record below makes that scope explicit. Text names do not verify runtime compiled values.
4. The original generic source/target shorthand did not distinguish MCI-R08's wrapped signed absolute test from MIR-R10's post-negation nonnegative clamp. INT_MIN survives the former but is zeroed by the latter. The 2026-10-15 instruction reading supports MCI-V04(d)/MIR-V09(f); arithmetic is not generalized from a different distance helper.

Original historical caveats: most analysis labels were misleading (for example an act mapper labelled as a gold-bank query and a collision setter labelled as a recipe matcher); names never supported identity alone. A decompiler lead misidentified CollideType as start Y; the original byte read and mode store govern. Earlier exact private source state, labels, fixture encodings and all original text have a SHA-256 checked private snapshot; they are not public source prerequisites. No rule was or is labelled not-found, and no six-rung absence search is claimed.

### Narrow negative-read record

The compatible 2026-10-12 source/data-anchor search found HitShift consumers in Common damage parameter functions `0x6A3E0`, `0x6A4B0`, `0x6A580`, `0x6B360`, and RandStart elsewhere outside the checked creator tree. Creator `0x6F930` and the specifically read non-damage setup helpers contained no direct field read of Light/RandStart/HitShift. This record supports only those instruction slices, not absence inside unread path/mod/callback handlers. MDP owns actual HitShift arithmetic and its own current gates. New migration did not turn the old bounded negative read into a global table-use claim. Supports MCI-R13.

<a id="named-cells-2026-10-15"></a>
## Named cells, 2026-10-15

This pass parsed the complete effective-file payloads already collected on 2026-10-15 through the running game's Storm loader. It reused the established answering LoD 1.13c profile, distinct from four modified profiles; the refreshed version-catalog attempt had failed. The source reported a prior file override disabled, but no complete current override inventory, loaded DLL fingerprints or physical supplying MPQs were captured. Manifest hashes identify analysis images, not the loaded archive set. No new gameplay action, hook, trace or memory write was performed. Existing hooks were untouched. This is a new narrow cell resolution from compatible same-date bytes, not a renewed gameplay or full-profile verification.

Each whole-file request used offset 0/length 0; decoded payload length equalled returned and total bytes. Columns were selected by header name; indices below exclude the header and begin at zero, and the explicit ID/id uniquely selects the same index. Raw string cells are retained; decimal ID cells convert to the stated integer. Hashes identify only the effective text files and do not prove equality to historical compiled/prepared records. Each named cell is verified in 1.13c data within those profile limits.

| Effective virtual file | Offset / requested length | Returned / total bytes | SHA-256 | Coverage |
|---|---|---|---|---|
| `data/global/excel/itemstatcost.txt` |0 /0 (whole file)|40109 /40109|`ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`|Complete header/file; only named cells below|
| `data/global/excel/states.txt` |0 /0 (whole file)|20920 /20920|`898806c2b69e55e316ecfed48dde95aa2933829b94fdad234bdabc27bdbc23e1`|Complete header/file; only named cells below|

| File | Row identity / zero-based index | Column | Raw cell | Supports |
|---|---|---|---|---|
| `itemstatcost.txt` | ID 161; index 161 | `ID` | `161` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 161; index 161 | `Stat` | `skill_handofathena` | Named identity in MCI section 3; historical name question closed |
| `states.txt` | id 87; index 87 | `id` | `87` | Named identity in MCI section 3; historical name question closed |
| `states.txt` | id 87; index 87 | `state` | `slowmissiles` | Named identity in MCI section 3; historical name question closed |

The original record resolved twenty named cells from these two payloads; the four above support this contract and the remaining sixteen are in the [sibling record](missile-initialization-and-registry.md#named-cells-2026-10-15). The consumer fixes these numeric ids independently of the data's name mapping. These named cells establish neither alternate remapping support nor a gameplay execution, and set no runtime_observed flag.

## Observation procedures and historical runtime status

No gameplay vectors have been observed. The 2026-10-12 record reported no reachable fleet member; that historical limitation does not assert that observation is impossible today. This regeneration did not attempt gameplay. All procedures use a legally owned vanilla 1.13c profile and matching original modules. Record class/level, difficulty, gear, row identity and actual supplied request values; stop if the setup cannot reach the needed branch naturally. No forced request/memory mutation is needed or authorized by these procedures.

<a id="mci-o01"></a>
### MCI-O01 speed/max/acceleration

Use a Normal level 1 Amazon with a bow and Magic Arrow or a throwing javelin, at the Blood Moor side of the Rogue Encampment exit. Record actual class/skill level and the requested missile row, including Vel/VelLev/MaxVel/Accel/CanSlow. Capture creator return and read path speed/max/acceleration for the returned missile before its next movement update. With no slow state, compute the MCI-R05→R07 value and compare maximum/acceleration (MIR-R09) independently; do not assume a named row has MIR-V08's synthetic values. A matching capture confirms only the recorded fields/branch.

<a id="mci-o02"></a>
### MCI-O02 reached seeder

Use the same Normal character and location, with callback/mod/weapon context recorded. Capture the game seed at Game `0xE0110` entry and after its step, and the missile's initial/final seed writes, while creator `0x6F930` supplies the game source seed. Expect exactly one step of that source and child `{new lo,666}` with initialLo copy, as MCI-V07. Separately log downstream damage/path/mod/callback draws before comparing whole creator entry/return; no whole-creation one-draw expectation is established by this procedure. Pierce-clear alone does not remove critical or other dependency draws.

<a id="mci-o03"></a>
### MCI-O03 distance

Use a Normal character with a naturally reachable aimed missile request; record class/level/gear/skill and source/start/owner/target positions. In open terrain locate a request whose post-nudge gate coordinate differs by 100 on one axis, and another comparable request differing by 99, with nonzero final speed. Capture creator's gate/return and seed state before the unit seeder. Expect 100 to return null before seeding and 99 to pass this gate, subject to later allocation/path outcomes. Ordinary view range may prevent this setup: inability to produce the request is inconclusive, not a no-record confirmation. The original proposal for a forced target point is retained privately as history and is not an authorized gameplay mutation.

## Migration, preservation and review record, 2026-10-15

The complete original spec was saved privately before editing and SHA-256 compared byte-for-byte. No prior companion/gap file existed. Repository consumption searches for the filename and MCI prefix found no implementation plan outside gaps, implementation code or test consumer. Practical execution-ordered ids remain without permanent/alias boilerplate; all 30 parent rule families, 12 original executable-vector families, 5 observation families and 12 historical question topics are retained in the contract or the dated assessment/corrections above. New subrules separate verified requests/fresh branches from unresolved execution. Original numeric cases, seeds and pierce sequences are retained, with exact request-vs-stored distinctions where the lifecycle owner is gated. (These counts describe the combined contract before the restructure record below.)

Current document status remains draft: unresolved integration/behavior and possible pending observations remain. No implementation, implementation comparison, plan, game action, trace or memory write occurred. The lead coordinates fresh independent-reader review and publication. The self-review/lint record is below; no historical reader result is claimed for this new behavioral contract.

### Allocation guard clarification

The compatible Common `0x2F370` reconstruction was checked again in the self-review. The creator tests its return for null before seeding, but the allocator itself clears its allocation unconditionally before returning. MCI-R09/MCI-V08 therefore specify the creator's supplied-absent-result guard; they do not establish safe null fallback after low-level allocation failure. This distinction also applies independently to unchecked path/metadata/base-list allocations (MIR-R01a/R02a/R11a). The historical phrase “allocation failure returns null” was too broad.

### Self-review and checks, 2026-10-15

The semantic rewrite has all template sections and a Description, explicit numeric widths/wrap/truncation and seed owners, complete local stage returns, output inventory and partial/full acceptance gates. All verified behavioral rows name executable fixtures; all multi-case fixtures give independent overriding inputs, draws and exact outputs. The seeder/pierce numeric families, early seed boundary, 0x2-over-0x20 priority, own-list 161, owner-position coincidence and 99/100 boundary were specifically compared with the original record. Original binary fixture encodings/history remain in the private snapshot; public locators resolve to the original artifacts.

No plan/code/test consumer was found in the repository-wide filename/prefix search outside spec/research/gap directories. Original text SHA-256 and preservation inventories are recorded privately. Twenty named text cells were parsed by header from complete compatible payloads, with no new archive/body publication. Source-function availability was checked before new analysis; the output-anchor candidate `0x38F40` has no reconstruction checked against original bytes and only supports the distinction between a missile count decrement and an unknown owner producer. No all-rungs absence claim is made.

The initial draft mechanical check reported 10 errors/152 warnings (primarily dense numeric spacing and one stale link); the corrected three-file draft reported 0 errors/0 warnings. The same checker is rerun after the final self-review edits and its output/hash recorded in the private handoff. No implementation tests or gameplay observations were run. This self-review is not an independent blind-reader review or approval; that review is coordinated by the lead for the resulting contract.

The lead's preliminary source/contract check requested two clarifications before independent review: the main contract now names the missile-creation message, with its numeric protocol identity retained only in this companion; MIR-R13 explicitly maps MDP source to owner and MDP parent to request source. MIR-V12 supplies a distinct missile parent id 99 and stored level 4464 against original request level 70000, preserving and clarifying its existing caller-order discriminator. These do not remove or promote any integration gate.

## Full blind-reader pass 1, 2026-10-15

The fresh reader examined all 20 vector families and 77 lettered cases using the main contract and its cited dependencies. Reported counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 14, COVERAGE 8 grouped findings, HARDEST 3. This is a documentation hand-recomputation/review, not gameplay observation, execution testing or approval. All 30 original parent rule families, twelve original vector families and five observation families remain. These counts cover the combined contract before the restructure; dispositions for ids now owned by the sibling contract are in its [review history](missile-initialization-and-registry.md#full-blind-reader-pass-1-2026-10-15), including the blocking DISCRIM on registry insertion.

| Advisory COVERAGE group | Outcome |
|---|---|
| MCI-R09 act thresholds | Added MCI-V09's supplied act-selection stage, covering both sides of each threshold and the >=1024 fallback; exact selected environment is asserted. This does not re-execute allocation/seeding |
| MCI-R11 modes/counter | Added MCI-V07(d) with CollideType 8 and mode 8. Existing MCI-V07(a)-(c) already distinguish ordinary increment, all-ones skip and zero rollover; no redundant counter cases |

GATED 14 consisted of the questions now numbered MCI-Q01/Q03/Q04 and MIR-Q01/Q02/Q03/Q04/Q05/Q06/Q08/Q09, the unresolved rules now numbered MIR-R11b/R14b/R17b and MCI-R12, and aggregate MCI-R13 accounting. These remain exact current limits; the reader's grouping does not downgrade MCI-R13's separately verified local draw accounting or promote any dependency effects. No rule label, ownership, numeric formula or gate boundary changed. The new supplied-stage assertions required a fresh full second reader pass; completion was not claimed from this pass.

## Full blind-reader pass 2, 2026-10-15

The fresh second reader reported MISMATCH 1, GUESS 0, GAP 0, DISCRIM 1, GATED 12 grouped outcomes, COVERAGE 12 grouped findings and HARDEST 3. The reader did not report a numeric vector-family or lettered-case count; the main contract at that review had 22 vector families, checked separately. These are review counts, not runtime results. They cover the combined contract before the restructure. The MISMATCH (metadata id wording) and DISCRIM (collision selectors 4/5) concern ids now owned by the sibling contract and are recorded in its [review history](missile-initialization-and-registry.md#full-blind-reader-pass-2-2026-10-15).

| Advisory COVERAGE group | Outcome |
|---|---|
| MCI-R01 successful validation | Added MCI-V01(f), valid player/id/record continuing to MCI-R02 with no draw/allocation at this isolated validation boundary |
| MCI-R08 no-target distance refusal | Added MCI-V06(i), no targetUnit and retained point 100 units from start, no nudge and early absent return with game seed unchanged |
| MCI-R12 synchronization | No change: MCI-Q03 gates full synchronization/delivery; the creator's local request does not supply a delivery fixture |
| MCI-R13 aggregate draw accounting | No change: local game/local-pierce accounting is verified and covered; complete path/damage/mod/callback budgets remain with their named owners and gates |

HARDEST for this contract, no change: MCI-R08's owner/start/tracking/retained-point distinctions, INT_MIN and zero-speed branch are explicit in MCI-V06/MCI-V04. The other two HARDEST findings concern the sibling contract.

The reader grouped GATED 12 as the questions now numbered MCI-Q01/Q02/Q03/Q04 and MIR-Q01/Q02/Q03/Q04/Q05/Q06/Q07/Q08/Q09 plus MDP gates. These retain their current assertion-level boundaries; grouping counts do not add a local gate to supplied-seed or neutral-field-write contracts. No rule label, proven arithmetic, ownership or gate changed. All existing identifiers/cases and hex values remained. The added fixtures and wording clarification required a fresh full third reader pass; independent review completion was not claimed then.

## Full blind-reader pass 3, 2026-10-15

The fresh reader outlined the contract in semantic operations and recomputed all 22 executable families and 117 cases from inputs/rules, ignoring traces. Each lettered independent case counts once and each unlettered fixture once; per-family counts of the combined contract were 6,5,3,5,9,4,7,3,1,11,5,12,4,3,4,8,1,8,4,1,11,2. Final counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 15 unique limits, COVERAGE 6, HARDEST 3. The 15 limits are the questions now numbered MCI-Q01..Q04 and MIR-Q01..Q09 and the unresolved rules MIR-R11b/R14b/R17b and MCI-R12; supplied assertions remain available. Dispositions for sibling-owned ids are in its [review history](missile-initialization-and-registry.md#full-blind-reader-pass-3-2026-10-15).

| Advisory entry | Disposition |
|---|---|
| MCI-R02 explicit start | No change: MCI-V02(b) already sets flags 0x3 with explicit start (10,20), distinct from source (100,200), and requires start (10,20)/point (15,17). This rejects ignoring the explicit-start flag. |
| MCI-R03 signed-wrap rectangle bounds | No change: the signed low32 upper-bound expression is explicit with no additional arithmetic branch. MCI-V03 pins half-open bounds, first room and ordered neighbors; actual rectangle production remains MCI-Q01. |

Every applicable blocking finding was resolved; this completed independent documentation review of the combined contract while draft status and maintainer approval remain pending. Final triage changed no rule, input, vector, expected output, acceptance or execution boundary. Publication lint, exact scope and preservation checks were repeated; no implementation tests/gameplay were performed.

## Restructure (2026-10-15)

The combined missile creation contract was split at the point where the creator has numbered the missile, following the approved review recommendation for oversized specs. The spec was not consumed by any plan, code or test, so both resulting specs were renumbered cleanly in execution order. This contract (renamed from the combined file) keeps validation, start/room/aim, speed, the distance refusal, allocation/act, seeding, identity/mode, the client-notification boundary and the request's draw accounting: 13 rule ids, 4 questions, 9 executable vectors and 3 observations. Metadata, the creation event, placement and registry insertion, range/lifetime/activation, missing path, path setup, destroy condition, moving speed/planner, last-collision source, acceleration, point-distance lifetime, base list, callback, damage requests, attribution, pierce, stat requests and the monster-mod event moved to [missile-initialization-and-registry](missile-initialization-and-registry.md) with its own companion and gap file. The combined adjacent-room/path question was split: adjacent-room collection production stays here (MCI-Q01); path, collision-room and update integration moved (MIR-Q02). The observation question was split into MCI-Q04 and MIR-Q09. The seeding vector MCI-V07 keeps its MIR-R01a metadata assertions; the speed/maximum vector moved to MIR-V08 because its discriminating assertions are maximum speed and acceleration. No rule behavior, label, value or vector result changed. Function, layout, assessment, correction, named-cell, observation and review records above were divided by owning rule with their dates retained. A fresh blind-implementer pass of the split contracts has not been run.

## Publication review — 2026-10-23

A fresh independent blind reader assessed the complete behavioral boundary in both split
MCI and MIR specs and read both gap files in full. The reader sampled companion identity,
source-boundary, correction, restructure and review-history sections, then independently
recomputed representative speed and aim/distance cases in MCI-V04 and related MIR range,
lifetime and pierce-seed cases. No in-scope rule/vector mismatch or guessed result was found.
The 2026-10-15 restructure note above accurately described review status at that date;
this current review is the fresh pass over the split contracts required by issue #499.
It is not an all-case arithmetic replay or a gameplay observation.

An independent evidence reader reviewed the six public split documents, source-manifest
identities, dated 2026-10-12/15 evidence, named-cell/profile limits, question and vector
IDs, and local links. A separate architecture reader checked issue #499 and README routes,
MCI's validation/aim/speed/allocation/seed/identity boundary, MIR's later same-request
stages, and the delegated movement, parameter and contact contracts. All three approved
bounded 1.13c documentation publication. Issue #499's body was updated to name the six
current split files in place of the retired combined filenames.

MCI-Q01–Q04 and sibling MIR-Q01–Q09 remain open at their stated boundaries. No behavioral rule, vector output, evidence label or gate changed.
This review adds no original-image investigation, protected game-data capture,
implementation comparison, gameplay test or 1.14d confirmation. Both contracts remain
draft.
