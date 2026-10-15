# Evidence: Missile Creation and Initialization

Behavioral contract: [Missile Creation and Initialization](../../specs/missile-creation-and-initialization.md).
Open evidence: [gap file](../../plans/gaps/missile-creation-and-initialization.md).

## Binary identities and methods

Evidence build is original Lord of Destruction Expansion 1.13c Windows x86. Function references below resolve through [source manifest](../source-manifest.md) artifacts `LOD113C-D2GAME` and `LOD113C-D2COMMON`. Their original images were rehashed on 2026-10-15 and matched the manifest SHA-256/size identities. Artifact identity establishes applicability of original-image locations, not historical loaded-module or archive identity. All function descriptions below are analyst descriptions, not original exported names. Export ordinals identify the matching D2Common artifact.

The 2026-10-12 record read the creator and its listed setup helpers by disassembly, with selected reconstruction checked against original bytes for small helpers. That compatible reading remains support for its recorded arithmetic/control-flow slices. New analysis on 2026-10-15 first checked whether reconstruction checked against original bytes was available; the adjacent-room getter was read in that form. The newly inspected creator, registration, placement, registry, distance, base-list, owner-attribution, monster-event and synchronization slices had no matching reconstruction and were read as disassembly. Encoded instruction runs in registry/distance/base-list and selected caller/dispatcher slices were decoded independently from the fingerprint-matched original image; a decoding check does not close an unread callee or producer.

## Dispatch and function references

Root route retained from 2026-10-12: D2Game skill server do-function table at RVA `0x1074A8`, entry 1, selects basic attack `0xA2C40`; its ranged route calls positioned-missile request `0xA28F0`, which calls creator `0x6F930`. The caller zeroes a 0x5C-byte request. Existing [skill dispatch evidence](skill-cast-start-and-do-dispatch.md) and [PvP evidence](combat-pvp-damage-and-kill-credit.md) support their own dispatch/root boundaries; this record does not reverify every skill/object/AI caller or the complete server scheduler. The historical discovery found approximately 100 creator references and 23 skill do-table entries through one intermediate; those are discovery counts, not completeness or acceptance claims.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0x6F930` | none | One-missile creator | MCI-R01..R30 | Disassembly 2026-10-12; targeted full creator reread 2026-10-15 for local order, flags, requests, refusals and widths; dependency effects separately gated |
| LOD113C-D2GAME | `0xA2C40`, `0xA28F0` | none | Attack route and positioned request | Trigger; MCI-R01/R02 | Compatible 2026-10-12 dispatch/caller record; current SCD root contract |
| LOD113C-D2GAME | `0x1250`, `0x1100`, `0x1130` | none | Missile record lookup, unit X/Y | MCI-R01/R02/R08 | Compatible 2026-10-12 reconstructions checked against original bytes; dynamic coordinate contract NRF-R10b; static-source result remains supplied at the caller boundary |
| LOD113C-D2GAME | `0x1280` | none | Containing room selection | MCI-R03 | Full disassembly 2026-10-15, owner room before ordered candidates; no producer/sort inferred |
| LOD113C-D2COMMON | `0x2FE10`, `0x3BA20` | 10331, 10383 | Unit room; stored adjacent collection/count | MCI-R03/R13a | Room read reused 2026-10-12; collection getter reconstruction checked against original bytes 2026-10-15, nullable candidates handled by Game consumer |
| LOD113C-D2COMMON | `0x33CD0`, `0x38520`, `0x39140`, `0x38B00` | 10494, 10871, 10680, none | State test/list lookup/own-list read | MCI-R06 | 2026-10-12 local calls plus current NRF-R12a/b/c; own list is distinct from aggregate |
| LOD113C-D2COMMON | `0x2F370`, `0x3C000`, `0x2D2C0` | 11153, 10826, 10459 | Unit allocation, room level, act selection | MCI-R09 | Compatible 2026-10-12 reconstruction checked against original bytes/read; act thresholds 1,40,75,103,109,1024 identified in original constant region; MSD shares room-level getter |
| LOD113C-D2GAME | `0xE0110` | none | Unit creation seeder | MCI-R10/R30 | Current NRF-R03a/b/NRF-R17a, original local game-seed caller retained; no whole-creation budget inferred |
| LOD113C-D2COMMON | `0x36740` | 10614 | Seed initializer | MCI-R10/R26a | NRF-R03a/b; compatible original reconstruction record |
| LOD113C-D2GAME | `0x6F8A0` | none | Missile metadata/id and creation event | MCI-R12 | Local disassembly reread 2026-10-15; fresh-unit event writes/insertions only |
| LOD113C-D2COMMON | `0x6A390`, `0x6A350` | 10098, none | Allocate/free missile metadata | MCI-R12a | Compatible 2026-10-12 reconstruction checked against original bytes; zeroing/source sentinel; unchecked failure remains distinct |
| LOD113C-D2GAME | `0x8E680`, `0x8E540`, `0x8E880`, `0x8E6F0` | none | Event allocation/link, allocator, chain cleanup and removal | MCI-R12b/MCI-Q03 | 2026-10-15 local link/cleanup callers and compatible STL-R01 allocation evidence; full nonfresh/removal/consumer integration is not established |
| LOD113C-D2GAME | `0xE0B40` | none | Parallel persistent-creation entry setup | MCI-Q03 | Original-image decoding 2026-10-15 corroborates shared event fields/calls only; no new scheduler ownership inferred |
| LOD113C-D2GAME | `0xE0F40` | none | Type-specific placement caller | MCI-R13a | 2026-10-15 disassembly plus original-image type-3 branch/jump target; allocation, registry, room/update request order |
| LOD113C-D2COMMON | `0x36030`, `0x665C0` | 11003, 10158 | Dynamic-path allocation; update marking | MCI-R13a/MCI-Q11 | Directed 2026-10-15 path read; nested path/collision/room callees remain unverified; update integration NRF-Q17 |
| LOD113C-D2GAME | `0xE0310` | none | Unit registry insertion | MCI-R13b | Full original-image decode 2026-10-15, unsigned descending insertion/equality assertion; supplied bucket chain |
| LOD113C-D2COMMON | `0x6A710`, `0x6A6D0`, `0x6A690` | 10558, 10970, 10837 | Total-range, lifetime and activation setters | MCI-R14/R21 | Compatible 2026-10-12 disassembly, signed i16 clamps; caller arithmetic reread 2026-10-15 |
| LOD113C-D2COMMON | `0x6A240`, `0x6A270` | none, 10737 | Elapsed and remaining-frame reads | Scope boundary | Compatible 2026-10-12 read; later countdown Game `0x3F320` belongs to movement |
| LOD113C-D2COMMON | `0x34D40`, `0x35040`, `0x349D0`, `0x348B0` | 10488, 10336, 10286, 11164 | Speed/reset, target-unit, target-point, collision-mask stores | MCI-R16/R18a | Compatible 2026-10-12 reconstruction checked against original bytes; setter reread 2026-10-15/current NRF-R08b; caller order preserved |
| LOD113C-D2COMMON | `0x35640` | 11056 | Collision-footprint update request | MCI-R16/MCI-Q11 | Compatible direct-call/read; nested collision-map publication is gated |
| LOD113C-D2COMMON | `0x36230` | 10334 | Path planner entry/dispatch | MCI-R18b/MCI-Q04 | Directed disassembly 2026-10-15; caller `(path,missile,0)` and ignored return verified; reached handlers and mutations remain integration gap |
| LOD113C-D2COMMON | `0x6A1B0`, `0x34CE0`, `0x34D10` | 10024, 11022, 10378 | Source metadata, acceleration and max-speed writes | MCI-R19/R20 | Compatible 2026-10-12 disassembly/reconstruction; no 75% maximum scale |
| LOD113C-D2COMMON | `0x7F5E0` | 10080 | Integer distance to supplied point | MCI-R21 | Full 144-byte original-image decode 2026-10-15, every return and integer wrap/clamp; no callees |
| LOD113C-D2COMMON | `0x3AF00` | 11166 | Base-list setup | MCI-R22 | Disassembly and full 208-byte original-image decode 2026-10-15; empty fresh branch complete under supplied allocation success; old-list cleanup `0x3A9F0` remains lifecycle integration |
| LOD113C-D2COMMON | `0x6A750`, `0x6A2F0`, `0x6A2D0` | 11167, 10652, none | Skill clamp, level narrow write/sign-extended read | MCI-R23 | Compatible 2026-10-12 disassembly/reconstruction and creator call order |
| LOD113C-D2GAME | `0x6F620`, `0x37540` | none | Weapon activation / damage wrapper | MCI-R24/MCI-Q06 | Wrapper reread 2026-10-15; full equipment operation remains MDP's producer boundary |
| LOD113C-D2COMMON | `0x6BAA0`, `0x6B4A0` | 10413, 10484 | MDP builder/store | MCI-R24 | Current MDP-R01..R12 with explicit producer/stat-write gates; no local duplicate damage formula |
| LOD113C-D2COMMON | `0x33020`, `0x32CA0`, `0x38100`, `0x3A7D0`, `0x3A280` | 10210, none, none, none, none | Owner source assignment / list allocation, attach, list-stat writes | MCI-R25 | Non-null-owner path in `0x33020` and full `0x32CA0` caller read 2026-10-15; STL-R06a/b/c requests/limits; allocation failure before valid-bit write |
| LOD113C-D2GAME | `0x6F660` | none | Local-seed pierce calculation | MCI-R26 | Full disassembly 2026-10-15; input read kinds retained from 2026-10-12 and NRF-R12b/c |
| LOD113C-D2COMMON | `0x39430`, `0x38B70`, `0x38C20`, `0x38A80`, `0x3A740` | 10910, 10973, 10587, none, 10887 | Aggregate/own-base reads; unit-write request | MCI-R26/R27 | Read consumers NRF-R12b/c; actual replacing write remains NRF-Q20, not verified by getter |
| LOD113C-D2COMMON | `0x6A330`, `0x6A310` | 10623, 11035 | Metadata condition read/write | MCI-R27 | Compatible 2026-10-12 reconstruction checked against original bytes; bit 2 OR in creator |
| LOD113C-D2GAME | `0x21880`, `0x268E0` | none | Monster event-5 wrapper/dispatcher | MCI-R28 | Wrapper disassembly and full dispatcher original-image decode 2026-10-15; table-selected callback interiors not read |
| LOD113C-D2GAME | `0x6F740`, `0x7B0A0` | none | Local missile sender and synchronization caller | MCI-R29/MCI-Q08 | Sender disassembly; caller original-image decoding 2026-10-15 establishes call at `0x7B348`, not full sync/client/delivery chain |
| LOD113C-D2COMMON | `0x7C8D0` | none | Missiles loader schema | Inputs | Compatible 2026-10-12 descriptor reading; field name/width/bit identity, not complete compiler/default behavior |

### Import and call placement

Original import/export resolution retained from 2026-10-12 used the identified original PE tables, with two independent parsers agreeing. In the creator: Game thunk `0xA478` resolves Common #10331/`0x2FE10`; `0xA55C`→#10383/`0x3BA20`; `0xA66A`→#11003/`0x36030`; `0xAAF0`→#10334/`0x36230`; `0xAB80`→#10080/`0x7F5E0`; `0xADAE`→#11166/`0x3AF00`; `0xAC76`→#10210/`0x33020`; `0xB132`→#10413/`0x6BAA0`; `0xB114`→#10484/`0x6B4A0`; `0xA472`→#10887/`0x3A740`. Other resolved targets are named in the function table. Thunks are navigation aids, not interchangeable target identities.

Creator entry takes game in ECX and request in EDX and returns missile-or-null in EAX. The damage wrapper receives owner in EAX (copied to ESI), missile in EBX, original level in EDI and source on stack. It calls weapon activation with selector EAX=0, then builder with output scratch record, owner/source/missile/level, then store with owner/missile/damage result/level. Creation's callback receives ECX=missile and EDX=callbackArg; its numeric return is ignored. The placement wrapper's missile branch passes six path-creation arguments `(pool,room,X,Y,unit,0)`. Its jump-table type-3 target is Game `0xE1068`. Stack cleanup alone was not used to infer argument count.

## Semantic fixture and layout mapping

This maps original observation fields to the behavioral contract; internal layouts are not required implementation architecture.

| Original object / fields | Semantic terms / supporting rules |
|---|---|
| Request +0x00 flags; +0x04 owner; +0x08 source; +0x0C targetUnit; +0x10 missileId | Request routing, MCI-R01/R02/R04/R08 |
| Request +0x14/+0x18 source X/Y; +0x1C/+0x20 target X/Y | Start, retained target point; MCI-R02/R04/R08/R21 |
| Request +0x28 speed; +0x2C skill; +0x30 level; +0x34 loops; +0x40 elapsed; +0x44 activate; +0x48 toHit; +0x4C range | Semantic scalar inputs in section 3 |
| Request +0x54 callback; +0x58 opaque argument | MCI-R23; unused +0x24/+0x38/+0x3C/+0x50 were not read by creator |
| Unit +0 type; +4 class; +8 pool; +0x0C id; +0x10 mode; +0x14 missile metadata; +0x18 act; +0x1C environment; +0x2C path; +0x80 game | MCI-R09/R11/R12/R13 |
| Unit +0x20/+0x24 seed; +0x28 initialLo; +0x44 animation; +0x5C base list | MCI-R10/R14/R22 |
| Unit +0xC8 bit 25 expansion; bit 26 condition A; +0xC4 bit 4 condition B, bit 1 condition C, bit 3 condition D, bit 2 destroy condition; +0xD0 selector; +0x74 auxiliary | MCI-R09/R11/R12a/R17/R25a; wider roles Q09 |
| Unit +0x94/+0x98 owner type/id; +0xC8 bit 10 valid | Direct attribution, MCI-R25a/PVP-R07 |
| Unit +0xDC event chain; +0xE4 registry successor | Creation-event association and registry insertion, MCI-R12b/R13b |
| Game +0x1C pool; +0x70 expansion; +0x9C counter; +0xBC+4*act environment; +0xD0/+0xD4 seed | MCI-R09/R10/R11 |
| Path +2/+6 integer position words; +0x10/+0x12 target words; +0x1C room; +0x38 reset; +0x4C footprint; +0x50 collision mask | Position/target/reset/collision semantic inputs/outputs |
| Path +0x58/+0x5C/+0x60 tracking unit/type/id; +0x7C speed; +0x84 maximum; +0x88 acceleration | MCI-R16/R18/R20 |
| Missile metadata +8 activation i16; +0x0A skill i16; +0x0C level i16; +0x0E totalRange i16; +0x10 lifetime i16; +0x14 conditions; +0x18/+0x1C last-collision source | MCI-R12a/R14/R19/R21/R23/R27; fresh block0x34 bytes/source id-1 |
| Room +0x4C/+0x50 origin; +0x54/+0x58 extent; +0 adjacent count; +0x24 adjacent array | MCI-R03; Common getter writes stored array first and count second; no dynamic candidate discovery |
| Base list +0x10 sign bit extended; +0x58 state storage; +0x5C/+0x60 callback/context; +0x44 associated unit | MCI-R22a; fresh base0x64 bytes plus zeroed state arrays |

Original Missiles record stride is0x1A4, record collection/count at DataTables+0xB64/+0xB6C. Loader schema maps flags byte+4 bits 0/2/3/4/5 to LastCollide/Pierce/CanSlow/CanDestroy/ClientSend; +0x96/+0x98 signed Range/LevRange; +0x9A/+0x9B/+0x9C unsigned Vel/VelLev/MaxVel; +0x9E signed Accel; +0x136 Activate; +0x180/+0x181/+0x182 SubLoop/SubStart/SubStop; +0x183/+0x184 CollideType/Collision; +0x19C DamageRate. Light+0x130/RandStart+0x17C/HitShift+0x196 direct-read search limits are below. No shipped row is reproduced.

Collision mask code entries are the first dwords of nine 8-byte Game entries beginning RVA `0x10D92C`; second dwords are later collision callbacks owned by movement/contact. The registry per-type table is at Game `0xFB014`; bucket=low7(id). Creation event groups use Game per-type table RVA `0xFA3DC` and the event scheduler's persistent-group heads; a new entry has byte+0=0, byte+2 OR4, frame+4=-1, target+8=missile, target id/type +0x0C/+0x10, override/parameters +0x14/+0x18/+0x2C zero. Unit linkage is +0x24/+0x28 and group linkage +0x1C/+0x20. These locators permit independent checking of the verified fresh insertion, without claiming an event consumer or mapping it to the separate unit-timer chain.

## Bounded question assessment and new readings, 2026-10-15

Before new analysis each historical question, gated rule, missing executable fixture and pending observation was checked against the original records, source manifest and current NRF/GDF/STL/MMF/SCD/PVP/MSD/MDP contracts. Unchanged old evidence keeps its original dates/methods. Compatibility is tied to manifest artifacts, explicit supplied inputs and recorded data-profile scope; no earlier failed attempt was treated as closure. New targeted reads answer remaining local questions rather than repeat a whole skill/scheduler/gameplay sweep.

| Question / family | Compatible evidence and remaining bounded investigation | Result / exact remaining boundary |
|---|---|---|
| Names, historical Q01 | GDF identity conventions and complete 2026-10-15 loader payloads; resolve only the named cells below by header | Named text identities established; no archive/compiled global correspondence asserted |
| Owner pierce seed Q02 | NRF existing-stat reads and creator pierce helper; targeted immediate/output-anchor search for stat 328 in D2Game | Game `0x38F40` decrements a missile's count, not an established owner seed producer. Other stat 328 literals include assertion/string/message contexts and unrelated candidates; no complete owner-writer contract. Remains outside boundary, does not gate supplied seed |
| Event/hash Q03 | Current STL frame-event allocation vs separate unit timers; full creator-registration local reading, original-image registry tail; parallel Game `0xE0B40` decoding | Fresh event association and descending insertion established. Complete persistent-entry consumer/removal and nonfresh chain cleanup remain integration gap |
| Planner Q04 | Existing NRF/KNB/MMF/SFR directed path coverage; read Common `0x36230` entry, guards/dispatch/mutation routes and creator call | Local `(path,missile,0)` request/ignored return established. Original planner route calls multiple direction/collision/handler operations; their full outputs/draws not verified by sibling supplied paths |
| Distance, historical Q05 | No exact existing dependency consumer for Common `0x7F5E0`; no reconstruction checked against original bytes; decode all 144 bytes from manifest original | Decoding closed; exact wrapped distance and caller's retained-point unsigned quotient/saturation established; full upstream planner position stays Q04 |
| Damage Q06 | Stable/current MDP arithmetic and request subsets; creator damage wrapper reading | Caller order/inputs established; equipment, third-chance, preparation and actual stat propagation retain their MDP/NRF/STL gates. No global draw-budget promotion |
| Monster Q07 | Historical pipeline dispatcher read remains partial for event 5; full original240-byte dispatcher decoded; wrapper guards read | Branch chooses event 5 table at Game RVA `0x10F45C` with24-byte mod stride, nine slots, missing callbacks skipped. Handler bodies/selection profile/effects/draws remain incomplete |
| Notify Q08 | Old sender record; checked original caller decode at Game `0x7B348` and local sender eligibility | Call-site byte uncertainty closed. Actual sync root, recipient/cadence/eligibility and network delivery remain unverified; no forced game action added |
| Neutral fields Q09 | Original direct writes retain exact values/flags; no compatible reader contract gives wider meaning | Keep neutral semantic conditions and wider-reader question. This is not a gate on the established writes |
| Observations Q10 | Historical 2026-10-12 no reachable member; current authorized work permits table reads, not gameplay | No gameplay vector performed. All observation families retained with tool-neutral procedures |
| Placement/rooms Q11 | Adjacent getter reconstruction, creator selector and directed Common 512-byte allocator read; original placement type 3 branch decoded | Stored collection order/consumer selection established. Allocator calls path-type/collision getters and room-publication operations; complete callees and map/update effects remain gated. It zeroes allocated path without checking failure; no generic null fallback established |
| CollideType Q12 | Exact nine code-mask entries retained; creator reads unsigned byte without local upper bound | Valid0..8 accepted; loader/reachable selectors above 8 not established. No fabricated default or clamping rule |
| Lifecycle Q13 | New precise branch from former R22/R25/R26b/R27 gates; NRF-Q20/STL-R06b/c | Fresh empty-base setup and direct/request subsets verified; existing-base cleanup, attachment/propagation and indirect notifications remain gates |

### Correction rationale and affected claims

The following discrepancies are retained as history; the current contract gives the supported result.

1. The 2026-10-12 R08 pseudocode assigned target-unit coordinates back to the target point. Creator `0x6F930` actually loads those coordinates only into temporary gate registers, leaving its retained stack point unchanged. R21 reloads the retained point. MCI-R04/R08/R21 and V05/V16 now distinguish tracking/gate position from point-duration input. This is source-supported correction, not a fitted gameplay expectation.
2. The original R21 allowed a quotient >=0x80000000 to become a negative saturated lifetime. The actual unsigned divisor is a32-bit left shift by 4; a nonzero divisor is at least 16 and numerator is at most0xFFFF0000, so quotient <=268431360. The generic signed clamp remains, but this caller cannot reach its negative-q branch. Original-image distance/caller arithmetic supports V16's saturated positive maximum instead. No existing supported numeric vector was removed.
3. Original R13b described an ordered hash bucket without checked tail. The checked unsigned `jbe`/`jne` route at Game `0xE0350..0xE0397` establishes descending ids, insertion before smaller, equality fatal; V15 discriminates ascending order and silent replacement.
4. Original R22 said any old list was freed and all later stats simply landed there. Common `0x3AF00` calls cleanup only for an extended old base and then clears the reference; full cleanup is still unverified. Fresh-empty branch is separate. The stat writers request operations whose stored effects remain NRF-Q20/STL-R06c. V12/V18 assert exact count/stat requests rather than pretend those are independently verified stored values.
5. Original R25 conditioned all source-list work on stateCount>98 and implied owner-valid was unconditional. Common `0x32CA0` gates state membership/update by that count, then continues source-list lookup/allocation/write with a present base. Allocation failure returns before valid-bit OR. Direct owner fields are written before that failure. Creator then clears its condition D and continues. V18 preserves this partial-success boundary.
6. Original V06/O02 and R30 could be read as a whole-creation exactly-one-game-draw guarantee with Pierce clear. Current NRF/MDP/Blade evidence requires separating the reached unit-seeder step, local pierce seed, callback, actual path, critical/preparation and monster-mod budgets. The original seed result0x6AC6935F and every original pierce sequence remain. O02 now measures the seeder boundary, not an unverified whole-operation budget.
7. Original Q01 requested a particular physical patch archive. Effective Storm file reads establish the loaded archive-set result, not which physical MPQ supplied it. The named-cell record below makes that scope explicit. Text names do not verify runtime compiled values.
8. The original generic source/target shorthand did not distinguish R08's wrapped signed absolute test from R21's post-negation nonnegative clamp. INT_MIN survives the former but is zeroed by the latter. The 2026-10-15 instruction reading supports V19(d)/V16(f); arithmetic is not generalized from a different distance helper.

Original historical caveats: most analysis labels were misleading (for example an act mapper labelled as a gold-bank query and a collision setter labelled as a recipe matcher); names never supported identity alone. A decompiler lead misidentified CollideType as start Y; the original byte read and mode store govern. Earlier exact private source state, labels, fixture encodings and all original text have a SHA-256 checked private snapshot; they are not public source prerequisites. No rule was or is labelled not-found, and no six-rung absence search is claimed.

### Narrow negative-read record

The compatible 2026-10-12 source/data-anchor search found HitShift consumers in Common damage parameter functions `0x6A3E0`, `0x6A4B0`, `0x6A580`, `0x6B360`, and RandStart elsewhere outside the checked creator tree. Creator `0x6F930` and the specifically read non-damage setup helpers contained no direct field read of Light/RandStart/HitShift. This record supports only those instruction slices, not absence inside unread path/mod/callback handlers. MDP owns actual HitShift arithmetic and its own current gates. New migration did not turn the old bounded negative read into a global table-use claim.

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
| `itemstatcost.txt` | ID 19; index 19 | `ID` | `19` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 19; index 19 | `Stat` | `tohit` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 156; index 156 | `ID` | `156` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 156; index 156 | `Stat` | `item_pierce` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 161; index 161 | `ID` | `161` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 161; index 161 | `Stat` | `skill_handofathena` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 166; index 166 | `ID` | `166` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 166; index 166 | `Stat` | `skill_pierce` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 327; index 327 | `ID` | `327` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 327; index 327 | `Stat` | `damage_framerate` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 328; index 328 | `ID` | `328` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 328; index 328 | `Stat` | `pierce_idx` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 353; index 353 | `ID` | `353` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 353; index 353 | `Stat` | `source_unit_type` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 354; index 354 | `ID` | `354` | Named identity in MCI section 3; historical name question closed |
| `itemstatcost.txt` | ID 354; index 354 | `Stat` | `source_unit_id` | Named identity in MCI section 3; historical name question closed |
| `states.txt` | id 87; index 87 | `id` | `87` | Named identity in MCI section 3; historical name question closed |
| `states.txt` | id 87; index 87 | `state` | `slowmissiles` | Named identity in MCI section 3; historical name question closed |
| `states.txt` | id 98; index 98 | `id` | `98` | Named identity in MCI section 3; historical name question closed |
| `states.txt` | id 98; index 98 | `state` | `sourceunit` | Named identity in MCI section 3; historical name question closed |

The consumer fixes these numeric ids independently of the data's name mapping. These twenty named cells establish neither alternate remapping support nor a gameplay execution, and set no runtime_observed flag.

## Observation procedures and historical runtime status

No gameplay vectors have been observed. The 2026-10-12 record reported no reachable fleet member; that historical limitation does not assert that observation is impossible today. This regeneration did not attempt gameplay. All procedures use a legally owned vanilla 1.13c profile and matching original modules. Record class/level, difficulty, gear, row identity and actual supplied request values; stop if the setup cannot reach the needed branch naturally. No forced request/memory mutation is needed or authorized by these procedures.

<a id="mci-o01"></a>
### MCI-O01 speed/max/acceleration

Use a Normal level 1 Amazon with a bow and Magic Arrow or a throwing javelin, at the Blood Moor side of the Rogue Encampment exit. Record actual class/skill level and the requested missile row, including Vel/VelLev/MaxVel/Accel/CanSlow. Capture creator return and read path speed/max/acceleration for the returned missile before its next movement update. With no slow state, compute the R05→R07 value and compare maximum/acceleration independently; do not assume a named row has V03's synthetic values. A matching capture confirms only the recorded fields/branch.

<a id="mci-o02"></a>
### MCI-O02 reached seeder

Use the same Normal character and location, with callback/mod/weapon context recorded. Capture the game seed at Game `0xE0110` entry and after its step, and the missile's initial/final seed writes, while creator `0x6F930` supplies the game source seed. Expect exactly one step of that source and child `{new lo,666}` with initialLo copy, as MCI-V06. Separately log downstream damage/path/mod/callback draws before comparing whole creator entry/return; no whole-creation one-draw expectation is established by this procedure. Pierce-clear alone does not remove critical or other dependency draws.

<a id="mci-o03"></a>
### MCI-O03 range/loops/elapsed

Use a Normal character with an available skill whose naturally captured request sets loop or elapsed flags, recording its class/character level/skill level and selected row. Stand still in Blood Moor with no nearby target needed. At creator entry record Range/LevRange/SubLoop/SubStart/SubStop/Activate, flags and loop/elapsed/override values. Read metadata totalRange/lifetime/activation and animation accumulator after their writes and before later update. Compare the exact R14 arithmetic including independent clamps; if the requested branch is not naturally available, record inconclusive setup rather than synthesize an expected row.

<a id="mci-o04"></a>
### MCI-O04 pierce

Use a Normal Amazon able to fire ten ordinary arrows with nonzero Pierce chance, recording character/skill levels and gear/stat contributions. Stand in a safe Blood Moor area and fire separated shots in one direction. Before each creator call read owner aggregate 156/166 and own-base328; capture local initialization/draws and count-write request in Game `0x6F660`, then read the missile's actual count after creation. Compute each count from its recorded own-base seed and chance using V12's procedure. A changing owner seed needs a separately identified writer; repeated counts alone do not prove one. A mismatch in stored count requires lifecycle/callback investigation and is not fitted by changing the expected arithmetic.

<a id="mci-o05"></a>
### MCI-O05 distance

Use a Normal character with a naturally reachable aimed missile request; record class/level/gear/skill and source/start/owner/target positions. In open terrain locate a request whose post-nudge gate coordinate differs by 100 on one axis, and another comparable request differing by 99, with nonzero final speed. Capture creator's gate/return and seed state before the unit seeder. Expect100 to return null before seeding and 99 to pass this gate, subject to later allocation/path outcomes. Ordinary view range may prevent this setup: inability to produce the request is inconclusive, not a no-record confirmation. The original proposal for a forced target point is retained privately as history and is not an authorized gameplay mutation.

## Migration, preservation and review record, 2026-10-15

The complete original spec was saved privately before editing and SHA-256 compared byte-for-byte. No prior companion/gap file existed. Repository consumption searches for the filename and MCI prefix found no implementation plan outside gaps, implementation code or test consumer. Practical execution-ordered ids remain without permanent/alias boilerplate; all 30 parent rule families,12 original executable-vector families,5 observation families and 12 historical question topics are retained in the contract or the dated assessment/corrections above. New subrules separate verified requests/fresh branches from unresolved execution. Original numeric cases, seeds and pierce sequences are retained, with exact request-vs-stored distinctions where the lifecycle owner is gated.

Current document status remains draft: unresolved integration/behavior and possible pending observations remain. No implementation, implementation comparison, plan, game action, trace or memory write occurred. The lead coordinates fresh independent-reader review and publication. The self-review/lint record is below; no historical reader result is claimed for this new behavioral contract.

### Allocation guard clarification

The compatible Common `0x2F370` reconstruction was checked again in the self-review. The creator tests its return for null before seeding, but the allocator itself clears its allocation unconditionally before returning. MCI-R09/MCI-V20 therefore specify the creator's supplied-absent-result guard; they do not establish safe null fallback after low-level allocation failure. This distinction also applies independently to unchecked path/metadata/base-list allocations. The historical phrase “allocation failure returns null” was too broad.

### Self-review and checks, 2026-10-15

The semantic rewrite has all template sections and a Description, explicit numeric widths/wrap/truncation and seed owners, complete local stage returns, output inventory and partial/full acceptance gates. All verified behavioral rows name executable fixtures; all multi-case fixtures give independent overriding inputs, draws and exact outputs. The seeder/pierce numeric families, early seed boundary, 0x2-over-0x20 priority, own-list 161, owner-position coincidence and 99/100 boundary were specifically compared with the original record. Original binary fixture encodings/history remain in the private snapshot; public locators resolve to the original artifacts.

No plan/code/test consumer was found in the repository-wide filename/prefix search outside spec/research/gap directories. Original text SHA-256 and preservation inventories are recorded privately. Twenty named text cells were parsed by header from complete compatible payloads, with no new archive/body publication. Source-function availability was checked before new analysis; the output-anchor candidate `0x38F40` has no reconstruction checked against original bytes and only supports the distinction between a missile count decrement and an unknown owner producer. No all-rungs absence claim is made.

The initial draft mechanical check reported 10 errors/152 warnings (primarily dense numeric spacing and one stale link); the corrected three-file draft reported 0 errors/0 warnings. The same checker is rerun after the final self-review edits and its output/hash recorded in the private handoff. No implementation tests or gameplay observations were run. This self-review is not an independent blind-reader review or approval; that review is coordinated by the lead for the resulting contract.

The lead's preliminary source/contract check requested two clarifications before independent review: the main contract now names the missile-creation message, with its numeric protocol identity retained only in this companion; R24 explicitly maps MDP source to owner and MDP parent to request source. V18 supplies a distinct missile parent id 99 and stored level 4464 against original request level 70000, preserving and clarifying its existing caller-order discriminator. These do not remove or promote any integration gate.

## Full blind-reader pass 1, 2026-10-15

The fresh reader examined all 20 vector families and 77 lettered cases using the main contract and its cited dependencies. Reported counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 14, COVERAGE 8 grouped findings, HARDEST 3. This is a documentation hand-recomputation/review, not gameplay observation, execution testing or approval. All 30 original parent rule families, twelve original vector families and five observation families remain.

The blocking DISCRIM concerned MCI-V15's small positive identities, which did not distinguish unsigned from signed descending comparison. Added V15(c): new id 0x80000101 goes before 513/129 in bucket 1. Signed comparison would append it. Added V15(d) for the distinct empty-tail append outcome. Existing middle insertion and fatal equality cases remain unchanged.

| Advisory COVERAGE group | Outcome |
|---|---|
| R09 act thresholds | Added V21's supplied act-selection stage, covering both sides of each threshold and the >=1024 fallback; exact selected environment is asserted. This does not re-execute allocation/seeding |
| R11 modes/counter | Added V06(d) with CollideType 8 and mode 8. Existing V06(a)-(c) already distinguish ordinary increment, all-ones skip and zero rollover; no redundant counter cases |
| R12a/b/R13b existing metadata/nonfresh chain/append/high bit | Added V22 to retain existing metadata without allocation, and V15(c)/(d) unsigned high-bit head/append. No change for nonfresh event chains: Q03 explicitly gates their cleanup/lifecycle and adding a guessed stored outcome would cross it |
| R14 negative/overflow range | Added V07(f) negative range -2 with activation -5, and V07(g) wrapped multiplication yielding 18 rather than a widened/clamped 32767. No new formula or source claim |
| R16 collision selectors/target storage | Added V10(e) exact tracking reference/type/id and no point-target write, plus remaining distinct mask results in V10(f)-(i). Selectors 4/5 duplicate the already exercised 0/0x104 results; their explicit table mappings remain, and no repetitive fixture is needed. Above 8 remains Q12 |
| R17 false CanDestroy with already-set condition | Added V22(a)/(b): false CanDestroy preserves both initially true and initially false destroy conditions, with no write request. This distinguishes unconditional clear and unconditional set |
| R23 absent callback/negative skill | Added V11(e) with no callback and skill -1, exact empty callback log and clamped 0/narrowed 4464 metadata. Existing present-callback overwrite cases remain |
| R25a/R26a state count and helper guards | Added V18(e) count 98: no membership/update change but source-list requests still occur and final creator returns M. Added V12(g)-(i) individual owner/type/Pierce refusals, and(j) accepted monster owner with unchanged exact four-draw sequence. These isolated helper cases do not assert reachability after creator refusal |

HARDEST, all no change: R21/V16 already specifies every wrap, post-negation clamp, shifted unsigned quotient and exact output; R26a/V12 already separates own-base seed from aggregate chance and signed stop/cap with exact local seeds; R25a/V18 already preserves direct writes, early allocation failure and the final creator return. The added count 98 case extends the last discriminator without changing attribution ownership or gates.

GATED 14 consists of Q03/Q04/Q06/Q07/Q08/Q10/Q11/Q12/Q13, R22b/R25b/R28b/R29 and aggregate R30 accounting. These remain exact current limits; the reader's grouping does not downgrade R30's separately verified local draw accounting or promote any dependency effects. No rule label, ownership, numeric formula or gate boundary changed. The new supplied-stage assertions require a fresh full second reader pass; completion is not claimed from this pass.

## Full blind-reader pass 2, 2026-10-15

The fresh second reader reported MISMATCH 1, GUESS 0, GAP 0, DISCRIM 1, GATED 12 grouped outcomes, COVERAGE 12 grouped findings and HARDEST 3. The reader did not report a numeric vector-family or lettered-case count; the main contract at that review had 22 vector families, checked separately. These are review counts, not runtime results.

MISMATCH outcome: clarify R12a's ambiguous “Store id” as “Store the unit id.” The reader had interpreted an id write into existing missile metadata. The original Game `0x6F8A0` instruction writes EAX to the unit's +0x0C field, as already mapped in this companion; no metadata-id field or write exists in this contract. V22 already correctly asserted unit id 42 and unchanged metadata N/source id 55. Its numeric result is retained. The issue was wording ambiguity, not an arithmetic or fixture correction, and no new source replay was required.

DISCRIM outcome: V10 now adds separate cases (j)/(k) for CollideType 4→mask 0 and CollideType 5→mask 0x104. Although their outputs duplicate other selector results, the exact positional mappings were a stated claim; those slots now have explicit executable cases. The earlier FULL1 advisory disposition is retained above as history and is superseded by this required discriminator.

| Advisory COVERAGE group | Outcome |
|---|---|
| R01 successful validation | Added V01(f), valid player/id/record continuing to R02 with no draw/allocation at this isolated validation boundary |
| R08 no-target distance refusal | Added V05(i), no targetUnit and retained point 100 units from start, no nudge and early absent return with game seed unchanged |
| R13a no-room update skip | Added V14(c), supplied absent room-test result with present path: path/registry requests remain, update is skipped. Original room/publication production remains Q11 |
| R16 selectors 4/5 and above 8 | Required V10(j)/(k) added as described above. No change above 8: Q12 explicitly gates that domain |
| R21 dy greater than dx | Added V16(h), transposed deltas 4/6 yielding distance 8/lifetime 14. Always doubling dx yields lifetime 12 and fails |
| R22b existing-base cleanup | No change: R22b/Q13 explicitly gate cleanup/propagation; a guessed stored cleanup outcome would cross the lifecycle boundary |
| R25a already-active state and successful missing-list allocation | Added V18(f) preserving changed=false for active 98 while still requesting update, and V18(g) exact allocate→full-attach-request→list-write sequence and final creator return. Actual attachment effects remain gated |
| R26b stat 166 and chance-sum wrapping | Added V12(k) with chance 60 solely from stat 166, and V12(l) wrapping 2147483647+1 to -2147483648 before a one-draw failure; exact local seed/count requests asserted |
| R27 optional flags off | Added V18(h), no optional stat 19 or metadata-condition write, metadata 8 retained, mandatory final rate request and return still present |
| R28b monster callbacks | No change: Q07 gates selected handler effects/draws; wrapper order remains independently covered |
| R29 synchronization | No change: Q08 gates full synchronization/delivery; the creator's local request does not supply a delivery fixture |
| R30 aggregate draw accounting | No change: local game/local-pierce accounting is verified and covered; complete path/damage/mod/callback budgets remain with their named owners and gates |

HARDEST, all no change: R08's owner/start/tracking/retained-point distinctions, INT_MIN and zero-speed branch are explicit in V05/V19; R21's wrapped signed distance then unsigned shifted division is explicit in V16, with the new symmetric-branch fixture; R26's local seed, signed chance comparison and four-pass cap retain exact V12 draws/final seeds. No new behavioral choice was required.

The reader grouped GATED 12 as Q02/Q03/Q04/Q06/Q07/Q08/Q09/Q10/Q11/Q12/Q13 plus MDP gates. These retain their current assertion-level boundaries; grouping counts do not add a local gate to supplied-seed or neutral-field-write contracts. No rule label, proven arithmetic, ownership or gate changed. All existing identifiers/cases and hex values remain. The added fixtures and wording clarification require a fresh full third reader pass; independent review completion is not claimed yet.

## Full blind-reader pass 3, 2026-10-15

The fresh reader outlined the contract in semantic operations and recomputed all 22 executable families and 117 cases from inputs/rules, ignoring traces. Each lettered independent case counts once and each unlettered fixture once; per-family counts were 6,5,3,5,9,4,7,3,1,11,5,12,4,3,4,8,1,8,4,1,11,2. Final counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 15 unique limits, COVERAGE 6, HARDEST 3. The 15 limits are Q02/Q03/Q04/Q06/Q07/Q08/Q09/Q10/Q11/Q12/Q13 and R22b/R25b/R28b/R29; supplied assertions remain available.

| Advisory entry | Disposition |
|---|---|
| R02 explicit start | No change: V02(b) already sets flags 0x3 with explicit start (10,20), distinct from source (100,200), and requires start (10,20)/point (15,17). This rejects ignoring the explicit-start flag. |
| R03 signed-wrap rectangle bounds | No change: the signed low32 upper-bound expression is explicit with no additional arithmetic branch. V13 pins half-open bounds, first room and ordered neighbors; actual rectangle production remains Q11. |
| R12b event absence/nonempty chain | No change: the verified branch explicitly requires supplied successful event allocation and a fresh empty chain. Nonfresh cleanup/consumption remains Q03; unchecked dependency failures have no invented safe-null outcome. |
| R16 zero-speed target storage | No change: R16 target storage is unconditional on speed; only R18 moving setter/planner has that condition. V05(e) retains a distant target at zero speed and V10(e) pins the exact tracking write; their specified stage contracts compose without a new default. |
| R20 zero-speed acceleration/maximum | No change: R20 expressly runs even at speed zero, with no local conditional branch. V03 pins unshifted acceleration and unscaled maximum, while V14(b/c) pins skipped moving setter/planner at zero speed; no speed gate is permitted around R20. |
| R25a list precedence | No change: this consumes the shared state-list lookup. NRF-R12a and NRF-V33(a)/(g) already discriminate primary before secondary and secondary fallback. Copying that owner matrix would add no local caller decision. |

HARDEST, all no change: V12/R26a specifies local seed ownership, every dependent step, signed stopping and cap; V16/R21 specifies every wrap, clamp and unsigned division; V18/R24-R28 separates level/roles and ordered requests across supplied integrations. The reader independently checked V12's four remainders and final seed. Every applicable blocking finding is resolved; this completes independent documentation review while draft status and maintainer approval remain pending. Final triage changes no rule, input, vector, expected output, acceptance or execution boundary. Publication lint, exact scope and preservation checks are repeated; no implementation tests/gameplay were performed.
