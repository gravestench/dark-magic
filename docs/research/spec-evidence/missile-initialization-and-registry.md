# Evidence: Missile Initialization and Registry

Behavioral contract: [Missile Initialization and Registry](../../specs/missile-initialization-and-registry.md).
Open evidence: [gap file](../../plans/gaps/missile-initialization-and-registry.md).
Sibling contract for the same creator's earlier stages: [Missile Creation and Seeding evidence](missile-creation-and-seeding.md).

## Binary identities and methods

Evidence build is original Lord of Destruction Expansion 1.13c Windows x86. Function references below resolve through [source manifest](../source-manifest.md) artifacts `LOD113C-D2GAME` and `LOD113C-D2COMMON`. Their original images were rehashed on 2026-10-15 and matched the manifest SHA-256/size identities. Artifact identity establishes applicability of original-image locations, not historical loaded-module or archive identity. All function descriptions below are analyst descriptions, not original exported names. Export ordinals identify the matching D2Common artifact.

The 2026-10-12 record read the creator and its listed setup helpers by disassembly, with selected reconstruction checked against original bytes for small helpers. That compatible reading remains support for its recorded arithmetic/control-flow slices. New analysis on 2026-10-15 first checked whether reconstruction checked against original bytes was available. The newly inspected creator, registration, placement, registry, distance, base-list, owner-attribution and monster-event slices had no matching reconstruction and were read as disassembly. Encoded instruction runs in registry/distance/base-list and selected dispatcher slices were decoded independently from the fingerprint-matched original image; a decoding check does not close an unread callee or producer.

## Dispatch and function references

These stages run inside creator `0x6F930` after it has numbered the missile; the dispatch root from the skill server do-function table to the creator is recorded in the [sibling companion](missile-creation-and-seeding.md#dispatch-and-function-references).

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0x6F930` | none | One-missile creator | MIR-R01a..R17b; earlier stages MCI-R01..R13 | Disassembly 2026-10-12; targeted full creator reread 2026-10-15 for local order, flags, requests, refusals and widths; dependency effects separately gated |
| LOD113C-D2GAME | `0x6F8A0` | none | Missile metadata/id and creation event | MIR-R01 | Local disassembly reread 2026-10-15; fresh-unit event writes/insertions only |
| LOD113C-D2COMMON | `0x6A390`, `0x6A350` | 10098, none | Allocate/free missile metadata | MIR-R01a | Compatible 2026-10-12 reconstruction checked against original bytes; zeroing/source sentinel; unchecked failure remains distinct |
| LOD113C-D2GAME | `0x8E680`, `0x8E540`, `0x8E880`, `0x8E6F0` | none | Event allocation/link, allocator, chain cleanup and removal | MIR-R01b/MIR-Q01 | 2026-10-15 local link/cleanup callers and compatible STL-R01 allocation evidence; full nonfresh/removal/consumer integration is not established |
| LOD113C-D2GAME | `0xE0B40` | none | Parallel persistent-creation entry setup | MIR-Q01 | Original-image decoding 2026-10-15 corroborates shared event fields/calls only; no new scheduler ownership inferred |
| LOD113C-D2GAME | `0xE0F40` | none | Type-specific placement caller | MIR-R02a | 2026-10-15 disassembly plus original-image type-3 branch/jump target; allocation, registry, room/update request order |
| LOD113C-D2COMMON | `0x2FE10` | 10331 | Unit room | MIR-R02a | Room read reused 2026-10-12; also used by MCI-R03 |
| LOD113C-D2COMMON | `0x36030`, `0x665C0` | 11003, 10158 | Dynamic-path allocation; update marking | MIR-R02a/MIR-Q02 | Directed 2026-10-15 path read; nested path/collision/room callees remain unverified; update integration RLI-Q02 |
| LOD113C-D2GAME | `0xE0310` | none | Unit registry insertion | MIR-R02b | Full original-image decode 2026-10-15, unsigned descending insertion/equality assertion; supplied bucket chain |
| LOD113C-D2COMMON | `0x6A710`, `0x6A6D0`, `0x6A690` | 10558, 10970, 10837 | Total-range, lifetime and activation setters | MIR-R03/R10 | Compatible 2026-10-12 disassembly, signed i16 clamps; caller arithmetic reread 2026-10-15 |
| LOD113C-D2COMMON | `0x6A240`, `0x6A270` | none, 10737 | Elapsed and remaining-frame reads | Scope boundary | Compatible 2026-10-12 read; later countdown Game `0x3F320` belongs to movement |
| LOD113C-D2COMMON | `0x34D40`, `0x35040`, `0x349D0`, `0x348B0` | 10488, 10336, 10286, 11164 | Speed/reset, target-unit, target-point, collision-mask stores | MIR-R05/R07a | Compatible 2026-10-12 reconstruction checked against original bytes; setter reread 2026-10-15/current NRF-R08b; caller order preserved |
| LOD113C-D2COMMON | `0x35640` | 11056 | Collision-footprint update request | MIR-R05/MIR-Q02 | Compatible direct-call/read; nested collision-map publication is gated |
| LOD113C-D2COMMON | `0x36230` | 10334 | Path planner entry/dispatch | MIR-R07b/MIR-Q04 | Directed disassembly 2026-10-15; caller `(path,missile,0)` and ignored return verified; reached handlers and mutations remain integration gap |
| LOD113C-D2COMMON | `0x6A1B0`, `0x34CE0`, `0x34D10` | 10024, 11022, 10378 | Source metadata, acceleration and max-speed writes | MIR-R08/R09 | Compatible 2026-10-12 disassembly/reconstruction; no 75% maximum scale |
| LOD113C-D2COMMON | `0x7F5E0` | 10080 | Integer distance to supplied point | MIR-R10 | Full 144-byte original-image decode 2026-10-15, every return and integer wrap/clamp; no callees |
| LOD113C-D2COMMON | `0x3AF00` | 11166 | Base-list setup | MIR-R11 | Disassembly and full 208-byte original-image decode 2026-10-15; empty fresh branch complete under supplied allocation success; old-list cleanup `0x3A9F0` remains lifecycle integration |
| LOD113C-D2COMMON | `0x6A750`, `0x6A2F0`, `0x6A2D0` | 11167, 10652, none | Skill clamp, level narrow write/sign-extended read | MIR-R12 | Compatible 2026-10-12 disassembly/reconstruction and creator call order |
| LOD113C-D2GAME | `0x6F620`, `0x37540` | none | Weapon activation / damage wrapper | MIR-R13/MIR-Q06 | Wrapper reread 2026-10-15; full equipment operation remains MDP's producer boundary |
| LOD113C-D2COMMON | `0x6BAA0`, `0x6B4A0` | 10413, 10484 | MDP builder/store | MIR-R13 | Current MDP-R01..R12 with explicit producer/stat-write gates; no local duplicate damage formula |
| LOD113C-D2COMMON | `0x33020`, `0x32CA0`, `0x38100`, `0x3A7D0`, `0x3A280` | 10210, none, none, none, none | Owner source assignment / list allocation, attach, list-stat writes | MIR-R14 | Non-null-owner path in `0x33020` and full `0x32CA0` caller read 2026-10-15; STL-R06a/b/c requests/limits; allocation failure before valid-bit write |
| LOD113C-D2GAME | `0x6F660` | none | Local-seed pierce calculation | MIR-R15 | Full disassembly 2026-10-15; input read kinds retained from 2026-10-12 and RLI-R02a/RLI-R02b |
| LOD113C-D2COMMON | `0x36740` | 10614 | Seed initializer | MIR-R15a; MCI-R10 | NRF-R03a/b; compatible original reconstruction record |
| LOD113C-D2COMMON | `0x39430`, `0x38B70`, `0x38C20`, `0x38A80`, `0x3A740` | 10910, 10973, 10587, none, 10887 | Aggregate/own-base reads; unit-write request | MIR-R15/R16 | Read consumers RLI-R02a/RLI-R02b; actual replacing write remains RLI-Q03, not verified by getter |
| LOD113C-D2COMMON | `0x6A330`, `0x6A310` | 10623, 11035 | Metadata condition read/write | MIR-R16 | Compatible 2026-10-12 reconstruction checked against original bytes; bit 2 OR in creator |
| LOD113C-D2GAME | `0x21880`, `0x268E0` | none | Monster event-5 wrapper/dispatcher | MIR-R17 | Wrapper disassembly and full dispatcher original-image decode 2026-10-15; table-selected callback interiors not read |
| LOD113C-D2COMMON | `0x7C8D0` | none | Missiles loader schema | Inputs | Compatible 2026-10-12 descriptor reading; field name/width/bit identity, not complete compiler/default behavior |

### Import and call placement

Original import/export resolution retained from 2026-10-12 used the identified original PE tables, with two independent parsers agreeing. In the creator's stages owned here: Game thunk `0xA66A`→#11003/`0x36030`; `0xAAF0`→#10334/`0x36230`; `0xAB80`→#10080/`0x7F5E0`; `0xADAE`→#11166/`0x3AF00`; `0xAC76`→#10210/`0x33020`; `0xB132`→#10413/`0x6BAA0`; `0xB114`→#10484/`0x6B4A0`; `0xA472`→#10887/`0x3A740`. The room thunks used by the earlier stages are in the [sibling companion](missile-creation-and-seeding.md#import-and-call-placement). Other resolved targets are named in the function table. Thunks are navigation aids, not interchangeable target identities.

The damage wrapper receives owner in EAX (copied to ESI), missile in EBX, original level in EDI and source on stack. It calls weapon activation with selector EAX=0, then builder with output scratch record, owner/source/missile/level, then store with owner/missile/damage result/level. Creation's callback receives ECX=missile and EDX=callbackArg; its numeric return is ignored. The placement wrapper's missile branch passes six path-creation arguments `(pool,room,X,Y,unit,0)`. Its jump-table type-3 target is Game `0xE1068`. Stack cleanup alone was not used to infer argument count.

## Semantic fixture and layout mapping

This maps original observation fields to the behavioral contract; internal layouts are not required implementation architecture. Request field offsets are in the [sibling mapping](missile-creation-and-seeding.md#semantic-fixture-and-layout-mapping).

| Original object / fields | Semantic terms / supporting rules |
|---|---|
| Request +0x54 callback; +0x58 opaque argument | MIR-R12 |
| Unit +0x0C id; +0x14 missile metadata; +0x18 act; +0x1C environment; +0x2C path | MIR-R01/R02 |
| Unit +0x44 animation; +0x5C base list | MIR-R03/R11 |
| Unit +0xC4 bit 1 condition C, bit 3 condition D, bit 2 destroy condition | MIR-R01a/R06/R14a; wider roles MCI-Q02 |
| Unit +0x94/+0x98 owner type/id; +0xC8 bit 10 valid | Direct attribution, MIR-R14a/PVP-R07 |
| Unit +0xDC event chain; +0xE4 registry successor | Creation-event association and registry insertion, MIR-R01b/R02b |
| Path +2/+6 integer position words; +0x10/+0x12 target words; +0x1C room; +0x38 reset; +0x4C footprint; +0x50 collision mask | Position/target/reset/collision semantic inputs/outputs |
| Path +0x58/+0x5C/+0x60 tracking unit/type/id; +0x7C speed; +0x84 maximum; +0x88 acceleration | MIR-R05/R07/R09 |
| Missile metadata +8 activation i16; +0x0A skill i16; +0x0C level i16; +0x0E totalRange i16; +0x10 lifetime i16; +0x14 conditions; +0x18/+0x1C last-collision source | MIR-R01a/R03/R08/R10/R12/R16; fresh block 0x34 bytes/source id-1 |
| Base list +0x10 sign bit extended; +0x58 state storage; +0x5C/+0x60 callback/context; +0x44 associated unit | MIR-R11a; fresh base 0x64 bytes plus zeroed state arrays |

Missiles record field offsets and loader flag bits are in the sibling mapping. Collision mask code entries are the first dwords of nine 8-byte Game entries beginning RVA `0x10D92C`; second dwords are later collision callbacks owned by movement/contact. The registry per-type table is at Game `0xFB014`; bucket=low7(id). Creation event groups use Game per-type table RVA `0xFA3DC` and the event scheduler's persistent-group heads; a new entry has byte+0=0, byte+2 OR4, frame+4=-1, target+8=missile, target id/type +0x0C/+0x10, override/parameters +0x14/+0x18/+0x2C zero. Unit linkage is +0x24/+0x28 and group linkage +0x1C/+0x20. These locators permit independent checking of the verified fresh insertion, without claiming an event consumer or mapping it to the separate unit-timer chain.

## Bounded question assessment and new readings, 2026-10-15

Before new analysis each historical question, gated rule, missing executable fixture and pending observation was checked against the original records, source manifest and current NRF/GDF/STL/MMF/SCD/PVP/MSD/MDP contracts. Unchanged old evidence keeps its original dates/methods. Compatibility is tied to manifest artifacts, explicit supplied inputs and recorded data-profile scope; no earlier failed attempt was treated as closure. New targeted reads answer remaining local questions rather than repeat a whole skill/scheduler/gameplay sweep. Rows for questions owned by the sibling contract are in its [assessment](missile-creation-and-seeding.md#bounded-question-assessment-and-new-readings-2026-10-15).

| Question / family | Compatible evidence and remaining bounded investigation | Result / exact remaining boundary |
|---|---|---|
| Owner pierce seed MIR-Q07 | NRF existing-stat reads and creator pierce helper; targeted immediate/output-anchor search for stat 328 in D2Game | Game `0x38F40` decrements a missile's count, not an established owner seed producer. Other stat 328 literals include assertion/string/message contexts and unrelated candidates; no complete owner-writer contract. Remains outside boundary, does not gate supplied seed |
| Event/hash MIR-Q01 | Current STL frame-event allocation vs separate unit timers; full creator-registration local reading, original-image registry tail; parallel Game `0xE0B40` decoding | Fresh event association and descending insertion established. Complete persistent-entry consumer/removal and nonfresh chain cleanup remain integration gap |
| Planner MIR-Q04 | Existing NRF/KNB/MMF/SFR directed path coverage; read Common `0x36230` entry, guards/dispatch/mutation routes and creator call | Local `(path,missile,0)` request/ignored return established. Original planner route calls multiple direction/collision/handler operations; their full outputs/draws not verified by sibling supplied paths |
| Distance, historical question | No exact existing dependency consumer for Common `0x7F5E0`; no reconstruction checked against original bytes; decode all 144 bytes from manifest original | Decoding closed; exact wrapped distance and caller's retained-point unsigned quotient/saturation established; full upstream planner position stays MIR-Q04 |
| Damage MIR-Q06 | Stable/current MDP arithmetic and request subsets; creator damage wrapper reading | Caller order/inputs established; equipment, third-chance, preparation and actual stat propagation retain their MDP/NRF/STL gates. No global draw-budget promotion |
| Monster MIR-Q08 | Historical pipeline dispatcher read remains partial for event 5; full original 240-byte dispatcher decoded; wrapper guards read | Branch chooses event 5 table at Game RVA `0x10F45C` with 24-byte mod stride, nine slots, missing callbacks skipped. Handler bodies/selection profile/effects/draws remain incomplete |
| Observations MIR-Q09 | Historical 2026-10-12 no reachable member; current authorized work permits table reads, not gameplay | No gameplay vector performed. All observation families retained with tool-neutral procedures |
| Placement/rooms MIR-Q02 | Directed Common 512-byte allocator read; original placement type 3 branch decoded | Allocator calls path-type/collision getters and room-publication operations; complete callees and map/update effects remain gated. It zeroes allocated path without checking failure; no generic null fallback established. Adjacent-room collection production is MCI-Q01 |
| CollideType MIR-Q03 | Exact nine code-mask entries retained; creator reads unsigned byte without local upper bound | Valid 0..8 accepted; loader/reachable selectors above 8 not established. No fabricated default or clamping rule |
| Lifecycle MIR-Q05 | New precise branch from former base-list, attribution, pierce-request and stat-request gates; RLI-Q03/STL-R06b/c | Fresh empty-base setup and direct/request subsets verified; existing-base cleanup, attachment/propagation and indirect notifications remain gates |

### Correction rationale and affected claims

The following discrepancies are retained as history; the current contract gives the supported result. The retained-point/tracking correction and the INT_MIN distinction between MCI-R08 and MIR-R10 are in the [sibling correction record](missile-creation-and-seeding.md#correction-rationale-and-affected-claims).

1. The original point-distance lifetime rule allowed a quotient >=0x80000000 to become a negative saturated lifetime. The actual unsigned divisor is a 32-bit left shift by 4; a nonzero divisor is at least 16 and numerator is at most 0xFFFF0000, so quotient <=268431360. The generic signed clamp remains, but this caller cannot reach its negative-q branch. Original-image distance/caller arithmetic supports MIR-V09's saturated positive maximum instead. No existing supported numeric vector was removed.
2. The original registry rule described an ordered hash bucket without checked tail. The checked unsigned `jbe`/`jne` route at Game `0xE0350..0xE0397` establishes descending ids, insertion before smaller, equality fatal; MIR-V03 discriminates ascending order and silent replacement.
3. The original base-list rule said any old list was freed and all later stats simply landed there. Common `0x3AF00` calls cleanup only for an extended old base and then clears the reference; full cleanup is still unverified. Fresh-empty branch is separate. The stat writers request operations whose stored effects remain RLI-Q03/STL-R06c. MIR-V13/MIR-V12 assert exact count/stat requests rather than pretend those are independently verified stored values.
4. The original attribution rule conditioned all source-list work on stateCount>98 and implied owner-valid was unconditional. Common `0x32CA0` gates state membership/update by that count, then continues source-list lookup/allocation/write with a present base. Allocation failure returns before valid-bit OR. Direct owner fields are written before that failure. Creator then clears its condition D and continues. MIR-V12 preserves this partial-success boundary.

Original historical caveats: most analysis labels were misleading (for example a collision setter labelled as a recipe matcher); names never supported identity alone. The original byte read governs the CollideType mask selection. No rule was or is labelled not-found, and no six-rung absence search is claimed.

<a id="named-cells-2026-10-15"></a>
## Named cells, 2026-10-15

This record is the part of the single 2026-10-15 named-cell pass that supports this contract; its payload provenance is identical to the [sibling record](missile-creation-and-seeding.md#named-cells-2026-10-15). The pass parsed the complete effective-file payloads already collected on 2026-10-15 through the running game's Storm loader. It reused the established answering LoD 1.13c profile, distinct from four modified profiles; the refreshed version-catalog attempt had failed. The source reported a prior file override disabled, but no complete current override inventory, loaded DLL fingerprints or physical supplying MPQs were captured. Manifest hashes identify analysis images, not the loaded archive set. No new gameplay action, hook, trace or memory write was performed. Existing hooks were untouched. This is a new narrow cell resolution from compatible same-date bytes, not a renewed gameplay or full-profile verification.

Each whole-file request used offset 0/length 0; decoded payload length equalled returned and total bytes. Columns were selected by header name; indices below exclude the header and begin at zero, and the explicit ID/id uniquely selects the same index. Raw string cells are retained; decimal ID cells convert to the stated integer. Hashes identify only the effective text files and do not prove equality to historical compiled/prepared records. Each named cell is verified in 1.13c data within those profile limits.

| Effective virtual file | Offset / requested length | Returned / total bytes | SHA-256 | Coverage |
|---|---|---|---|---|
| `data/global/excel/itemstatcost.txt` |0 /0 (whole file)|40109 /40109|`ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`|Complete header/file; only named cells below|
| `data/global/excel/states.txt` |0 /0 (whole file)|20920 /20920|`898806c2b69e55e316ecfed48dde95aa2933829b94fdad234bdabc27bdbc23e1`|Complete header/file; only named cells below|

| File | Row identity / zero-based index | Column | Raw cell | Supports |
|---|---|---|---|---|
| `itemstatcost.txt` | ID 19; index 19 | `ID` | `19` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 19; index 19 | `Stat` | `tohit` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 156; index 156 | `ID` | `156` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 156; index 156 | `Stat` | `item_pierce` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 166; index 166 | `ID` | `166` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 166; index 166 | `Stat` | `skill_pierce` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 327; index 327 | `ID` | `327` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 327; index 327 | `Stat` | `damage_framerate` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 328; index 328 | `ID` | `328` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 328; index 328 | `Stat` | `pierce_idx` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 353; index 353 | `ID` | `353` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 353; index 353 | `Stat` | `source_unit_type` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 354; index 354 | `ID` | `354` | Named identity in MIR section 3; historical name question closed |
| `itemstatcost.txt` | ID 354; index 354 | `Stat` | `source_unit_id` | Named identity in MIR section 3; historical name question closed |
| `states.txt` | id 98; index 98 | `id` | `98` | Named identity in MIR section 3; historical name question closed |
| `states.txt` | id 98; index 98 | `state` | `sourceunit` | Named identity in MIR section 3; historical name question closed |

The consumer fixes these numeric ids independently of the data's name mapping. These sixteen named cells establish neither alternate remapping support nor a gameplay execution, and set no runtime_observed flag.

## Observation procedures and historical runtime status

No gameplay vectors have been observed. The 2026-10-12 record reported no reachable fleet member; that historical limitation does not assert that observation is impossible today. This regeneration did not attempt gameplay. All procedures use a legally owned vanilla 1.13c profile and matching original modules. Record class/level, difficulty, gear, row identity and actual supplied request values; stop if the setup cannot reach the needed branch naturally. No forced request/memory mutation is needed or authorized by these procedures.

<a id="mir-o01"></a>
### MIR-O01 range/loops/elapsed

Use a Normal character with an available skill whose naturally captured request sets loop or elapsed flags, recording its class/character level/skill level and selected row. Stand still in Blood Moor with no nearby target needed. At creator entry record Range/LevRange/SubLoop/SubStart/SubStop/Activate, flags and loop/elapsed/override values. Read metadata totalRange/lifetime/activation and animation accumulator after their writes and before later update. Compare the exact MIR-R03 arithmetic including independent clamps; if the requested branch is not naturally available, record inconclusive setup rather than synthesize an expected row.

<a id="mir-o02"></a>
### MIR-O02 pierce

Use a Normal Amazon able to fire ten ordinary arrows with nonzero Pierce chance, recording character/skill levels and gear/stat contributions. Stand in a safe Blood Moor area and fire separated shots in one direction. Before each creator call read owner aggregate 156/166 and own-base 328; capture local initialization/draws and count-write request in Game `0x6F660`, then read the missile's actual count after creation. Compute each count from its recorded own-base seed and chance using MIR-V13's procedure. A changing owner seed needs a separately identified writer; repeated counts alone do not prove one. A mismatch in stored count requires lifecycle/callback investigation and is not fitted by changing the expected arithmetic.

## Migration and self-review records

The 2026-10-15 migration, preservation, allocation-guard clarification and self-review records of the combined contract are kept intact in the [sibling companion](missile-creation-and-seeding.md#migration-preservation-and-review-record-2026-10-15); they applied to every rule now in this contract. The allocation-guard clarification there covers the unchecked metadata, path and base-list allocations of MIR-R01a/R02a/R11a.

## Full blind-reader pass 1, 2026-10-15

The fresh reader examined all 20 vector families and 77 lettered cases of the combined contract. Reported counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 14, COVERAGE 8 grouped findings, HARDEST 3 (full header in the [sibling record](missile-creation-and-seeding.md#full-blind-reader-pass-1-2026-10-15)). This is a documentation hand-recomputation/review, not gameplay observation, execution testing or approval.

The blocking DISCRIM concerned the registry vector's small positive identities, which did not distinguish unsigned from signed descending comparison. Added MIR-V03(c): new id 0x80000101 goes before 513/129 in bucket 1. Signed comparison would append it. Added MIR-V03(d) for the distinct empty-tail append outcome. Existing middle insertion and fatal equality cases remain unchanged.

| Advisory COVERAGE group | Outcome |
|---|---|
| MIR-R01a/b/R02b existing metadata/nonfresh chain/append/high bit | Added MIR-V01 to retain existing metadata without allocation, and MIR-V03(c)/(d) unsigned high-bit head/append. No change for nonfresh event chains: MIR-Q01 explicitly gates their cleanup/lifecycle and adding a guessed stored outcome would cross it |
| MIR-R03 negative/overflow range | Added MIR-V04(f) negative range -2 with activation -5, and MIR-V04(g) wrapped multiplication yielding 18 rather than a widened/clamped 32767. No new formula or source claim |
| MIR-R05 collision selectors/target storage | Added MIR-V07(e) exact tracking reference/type/id and no point-target write, plus remaining distinct mask results in MIR-V07(f)-(i). Selectors 4/5 duplicate the already exercised 0/0x104 results; their explicit table mappings remain, and no repetitive fixture is needed. Above 8 remains MIR-Q03 |
| MIR-R06 false CanDestroy with already-set condition | Added MIR-V01(a)/(b): false CanDestroy preserves both initially true and initially false destroy conditions, with no write request. This distinguishes unconditional clear and unconditional set |
| MIR-R12 absent callback/negative skill | Added MIR-V11(e) with no callback and skill -1, exact empty callback log and clamped 0/narrowed 4464 metadata. Existing present-callback overwrite cases remain |
| MIR-R14a/R15a state count and helper guards | Added MIR-V12(e) count 98: no membership/update change but source-list requests still occur and final creator returns M. Added MIR-V13(g)-(i) individual owner/type/Pierce refusals, and (j) accepted monster owner with unchanged exact four-draw sequence. These isolated helper cases do not assert reachability after creator refusal |

HARDEST, all no change: MIR-R10/MIR-V09 already specifies every wrap, post-negation clamp, shifted unsigned quotient and exact output; MIR-R15a/MIR-V13 already separates own-base seed from aggregate chance and signed stop/cap with exact local seeds; MIR-R14a/MIR-V12 already preserves direct writes, early allocation failure and the final creator return. The added count 98 case extends the last discriminator without changing attribution ownership or gates. The GATED disposition is in the sibling record.

## Full blind-reader pass 2, 2026-10-15

The fresh second reader reported MISMATCH 1, GUESS 0, GAP 0, DISCRIM 1, GATED 12 grouped outcomes, COVERAGE 12 grouped findings and HARDEST 3 for the combined contract (full header and GATED disposition in the [sibling record](missile-creation-and-seeding.md#full-blind-reader-pass-2-2026-10-15)). These are review counts, not runtime results.

MISMATCH outcome: clarify MIR-R01a's ambiguous “Store id” as “Store the unit id.” The reader had interpreted an id write into existing missile metadata. The original Game `0x6F8A0` instruction writes EAX to the unit's +0x0C field, as already mapped in this companion; no metadata-id field or write exists in this contract. MIR-V01 already correctly asserted unit id 42 and unchanged metadata N/source id 55. Its numeric result is retained. The issue was wording ambiguity, not an arithmetic or fixture correction, and no new source replay was required.

DISCRIM outcome: MIR-V07 now adds separate cases (j)/(k) for CollideType 4→mask 0 and CollideType 5→mask 0x104. Although their outputs duplicate other selector results, the exact positional mappings were a stated claim; those slots now have explicit executable cases. The earlier pass-1 advisory disposition is retained above as history and is superseded by this required discriminator.

| Advisory COVERAGE group | Outcome |
|---|---|
| MIR-R02a no-room update skip | Added MIR-V02(c), supplied absent room-test result with present path: path/registry requests remain, update is skipped. Original room/publication production remains MIR-Q02 |
| MIR-R05 selectors 4/5 and above 8 | Required MIR-V07(j)/(k) added as described above. No change above 8: MIR-Q03 explicitly gates that domain |
| MIR-R10 dy greater than dx | Added MIR-V09(h), transposed deltas 4/6 yielding distance 8/lifetime 14. Always doubling dx yields lifetime 12 and fails |
| MIR-R11b existing-base cleanup | No change: MIR-R11b/MIR-Q05 explicitly gate cleanup/propagation; a guessed stored cleanup outcome would cross the lifecycle boundary |
| MIR-R14a already-active state and successful missing-list allocation | Added MIR-V12(f) preserving changed=false for active 98 while still requesting update, and MIR-V12(g) exact allocate→full-attach-request→list-write sequence and final creator return. Actual attachment effects remain gated |
| MIR-R15b stat 166 and chance-sum wrapping | Added MIR-V13(k) with chance 60 solely from stat 166, and MIR-V13(l) wrapping 2147483647+1 to -2147483648 before a one-draw failure; exact local seed/count requests asserted |
| MIR-R16 optional flags off | Added MIR-V12(h), no optional stat 19 or metadata-condition write, metadata 8 retained, mandatory final rate request and return still present |
| MIR-R17b monster callbacks | No change: MIR-Q08 gates selected handler effects/draws; wrapper order remains independently covered |

HARDEST for this contract, no change: MIR-R10's wrapped signed distance then unsigned shifted division is explicit in MIR-V09, with the new symmetric-branch fixture; MIR-R15's local seed, signed chance comparison and four-pass cap retain exact MIR-V13 draws/final seeds. No new behavioral choice was required.

## Full blind-reader pass 3, 2026-10-15

The fresh reader recomputed all 22 executable families and 117 cases of the combined contract from inputs/rules, ignoring traces. Final counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 15 unique limits, COVERAGE 6, HARDEST 3 (full header and per-family counts in the [sibling record](missile-creation-and-seeding.md#full-blind-reader-pass-3-2026-10-15)).

| Advisory entry | Disposition |
|---|---|
| MIR-R01b event absence/nonempty chain | No change: the verified branch explicitly requires supplied successful event allocation and a fresh empty chain. Nonfresh cleanup/consumption remains MIR-Q01; unchecked dependency failures have no invented safe-null outcome. |
| MIR-R05 zero-speed target storage | No change: MIR-R05 target storage is unconditional on speed; only MIR-R07's moving setter/planner has that condition. MCI-V06(e) retains a distant target at zero speed and MIR-V07(e) pins the exact tracking write; their specified stage contracts compose without a new default. |
| MIR-R09 zero-speed acceleration/maximum | No change: MIR-R09 expressly runs even at speed zero, with no local conditional branch. MIR-V08 pins unshifted acceleration and unscaled maximum, while MIR-V02(b/c) pins skipped moving setter/planner at zero speed; no speed gate is permitted around MIR-R09. |
| MIR-R14a list precedence | No change: this consumes the shared state-list lookup. RLI-R03 and RLI-V05(a)/(g) already discriminate primary before secondary and secondary fallback. Copying that owner matrix would add no local caller decision. |

HARDEST, all no change: MIR-V13/MIR-R15a specifies local seed ownership, every dependent step, signed stopping and cap; MIR-V09/MIR-R10 specifies every wrap, clamp and unsigned division; MIR-V12/MIR-R13..R17a separates level/roles and ordered requests across supplied integrations. The reader independently checked MIR-V13's four remainders and final seed. Every applicable blocking finding was resolved; this completed independent documentation review of the combined contract while draft status and maintainer approval remain pending. Final triage changed no rule, input, vector, expected output, acceptance or execution boundary.

## Restructure (2026-10-15)

This contract was created by splitting the combined missile creation contract at the point where the creator has numbered the missile, following the approved review recommendation for oversized specs. The combined spec was not consumed by any plan, code or test, so this contract was numbered cleanly in execution order: 24 rule ids, 9 questions, 13 executable vectors and 2 observations. It received metadata and the creation event, placement and registry insertion, range/lifetime/activation, missing path, path setup, destroy condition, moving speed/planner, last-collision source, acceleration/maximum speed, point-distance lifetime, base list, callback/skill/level, damage requests, owner attribution, pierce, stat requests/return and the monster-mod event. Validation, aim, speed, allocation, seeding, identity, the client-notification boundary and draw accounting remain in [missile-creation-and-seeding](missile-creation-and-seeding.md). The combined path/room question was split (path, collision-room and update integration here as MIR-Q02; adjacent-room production stays MCI-Q01), and the observation question was split (MIR-Q09 here). The speed/maximum vector came here as MIR-V08 because its discriminating assertions are maximum speed and acceleration; its speed values remain MCI-R05/R07 assertions. The neutral-condition question stayed with the sibling as MCI-Q02 and is cited by MIR-R01a/R06. No rule behavior, label, value or vector result changed. Function, layout, assessment, correction, named-cell, observation and review records were divided by owning rule with their dates retained; the migration and self-review records stay in the sibling companion. A fresh blind-implementer pass of the split contracts has not been run.

## Publication review — 2026-10-23

The fresh independent reader of the split MCI/MIR contracts read both complete specs and
gap files, sampled their companion identity/correction/restructure/review records, and
assessed the full behavioral boundary. Representative MIR-V04 range/activation,
MIR-V05 independent clamps, MIR-V09 distance/lifetime and MIR-V13 pierce-seed cases were
recomputed without a mismatch or an invented integration result. The dated 2026-10-15
restructure statement above remains historical; this 2026-10-23 review supplies the
post-split blind pass for issue #499. It does not claim an all-case replay or gameplay.

An independent evidence reader checked the six split public documents, manifest identities,
dated records, names/limits, links and defined IDs. A separate architecture reader checked
MIR's metadata/event/registry/range/path/damage-request/owner/pierce/stat/return ownership
against MCI, movement, damage-parameter and contact specs. Both approved the handoff along
with the blind reader. The issue body now names the six current files.

MIR-Q01–Q09 and sibling MCI-Q01–Q04 retain their exact open evidence boundaries, including
event, path, stat and damage integration plus original-game observation. No behavioral rule,
vector output, evidence label or gate changed. No new original-image investigation,
protected game-data capture, implementation comparison, gameplay test or 1.14d
confirmation was added. Both contracts remain draft.
