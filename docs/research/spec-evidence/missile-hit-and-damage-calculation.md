# Evidence: Missile Hit and Damage Calculation

Behavioral contract: [Missile Hit and Damage Calculation](../../specs/missile-hit-and-damage-calculation.md).
Original artifacts: [source manifest](../source-manifest.md).

## Identity, methods and reuse

Evidence build is original Lord of Destruction Expansion 1.13c, Windows x86. All D2Game locations refer to `LOD113C-D2GAME` (SHA-256 `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`); D2Common refers to `LOD113C-D2COMMON` (SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`). Manifest sizes, preferred bases and matching-original correspondence apply. Descriptions below are analyst labels, not original exported names.

The complete previous contract, provenance, exact binary fixtures, observations and questions were preserved privately before rewriting. Compatible 2026-10-12 code readings retain that date and their actual limits. The current NRF/GDF/PVP/SBA/HRC/KNB/IEP/STL/MDP/MCI contracts were checked for applicable input/output boundaries. Existing 2026-10-15 Blade evidence corroborates the full 688-byte fill image and per-element roller; this supports the base slice, not its previously gated target adjustment by itself. The missing target/critical and dispatch paths received targeted new reading on 2026-10-15.

New readings checked availability of reconstructions matched to original machine code before disassembly. No selected Game contact/fill/apply/lookup/filter/miss-entry target or Common last-target/type-enumeration/classification target had one. Game cooldown-removal callback and Common source-flags getter did, and those matching reconstructions were read first. Original-image decoding corroborated encoded instructions across the complete 832-byte contact dispatcher, 112-byte no-target entry, 112-byte creator lookup and existing 688-byte fill image. All branches were followed to relevant output/call/return, rather than treating a linear decoder's byte count alone as control-flow proof. Private analysis identities, availability details, original listings and preservation hashes remain outside the public repository.

These methods establish the stated supplied-input and request boundaries. They do not establish full stat/registry/map/list lifecycle integration or run any gameplay action. The previous failed loader attempt on 2026-10-12 is retained as an unsuccessful historical attempt, not a claim that no 1.13c game can now be reached.

## Function records and anchors

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | 0x3F320; 0x3EDE0 | none | movement contact caller; contact dispatcher | R13..R21/R24/R25 | 2026-10-12 roots plus 2026-10-15 complete dispatcher instruction/image reading. Movement owns caller final return; local fill/callback/apply/stop writes identify output end |
| `LOD113C-D2GAME` | 0xE03A0 | none | registered creator lookup | R13b/R26 | 2026-10-15 complete original-image corroboration: game guard, kind selection, hash/list traversal, first identity match; registry population supplied |
| `LOD113C-D2GAME` | 0x38FB0 | none | contact kind filter | R15c | 2026-10-15 compatible movement record: complete 64-byte original-image proof plus table at 0x38FDC; unsigned selector 0/1/2/other branches |
| `LOD113C-D2GAME` | 0x38F40 | none | pierce counter decision | R16 | 2026-10-12 compatible reading and new 2026-10-15 complete instruction reading; nonzero decrements by replacement request; no persistence promotion |
| `LOD113C-D2GAME` | 0xDDE90 | none | optional hit roll | R18 | 2026-10-15 narrow null-argument guard to 0xDE0A5: either missing attacker/target returns 0 before calls. Non-null formula consumes SBA's existing contract |
| `LOD113C-D2GAME` | 0x3E590; 0x38E60 | none | cooldown preparation; installed callback | R20 | 2026-10-15 complete preparation disassembly; matched callback reconstruction requests state off. STL/NRF own attachment/state effects |
| `LOD113C-D2GAME` | 0x37030 | none | contact item-event dispatcher | R20 | Compatible IEP-R07 full 2026-10-13 dispatcher record and current contract: event 0 on target, other missile, absent record, result ignored |
| `LOD113C-D2GAME` | 0x3E510 | none | separate no-target callback/remove entry | R22 | 2026-10-15 full instruction/image corroboration: callback result ignored; tail request to 0xE20A0; real callers unestablished |
| `LOD113C-D2GAME` | 0x3A4E0 | none | hit-time fill | R23 | Compatible base-read evidence and complete 2026-10-15 688-byte original image; new target/critical path from 0x3A646 through 0x3A790, all owned callees resolved |
| `LOD113C-D2GAME` | 0x39B40 | none | elemental roller | R23b | Compatible Blade 2026-10-15 complete reading plus checked call arguments; original endpoint guards before percentage, no second swap |
| `LOD113C-D2GAME` | 0x01620 | none | monster-type relation consumer | R23d | New 2026-10-15 complete instruction reading: positive/count bounds on both indices then matrix bit, not equality; supplied matrix producer |
| `LOD113C-D2GAME` | 0x11D0; 0x14D0 | none | guarded Rand; mulDiv | R23 | Reuse NRF-R01/R02/R05 with original operand order and existing code evidence; no formula inferred from labels |
| `LOD113C-D2GAME` | 0x3AFF0 | none | apply then armor request | R25b/R26 | 2026-10-15 full 80-byte instruction reading: apply return never gates armor; stat setter actual propagation gated |
| `LOD113C-D2GAME` | 0x3AE10 | none | direct missile apply | R26 | Compatible SBA/HRC/KNB record plus new full 480-byte reading through process/results calls; creator resolution, shared-record writes and ignored returns |
| `LOD113C-D2GAME` | 0xDB790; 0x770E0; 0xDE0C0; 0xDCF20 | none | defender check; owner double; process; result consumer | R26 | Named SBA/DPO/HRC boundary contracts; caller placement/arguments checked 2026-10-15, downstream effects retain their gates |
| `LOD113C-D2GAME` | 0xE0A60; 0xE20A0 | none | hostility; removal | R15d/R22 | PVP predicate and movement removal boundaries; local caller supplies these operations |
| `LOD113C-D2COMMON` | 0x6A130; 0x6A1B0 | 10227;10024 | last-target equality/setter | R15b | New complete 2026-10-15 disassembly; type 3/valid definition/LastCollide/data guards, saved kind/id, null-target id-1 |
| `LOD113C-D2COMMON` | 0x1A610; 0x1A5B0 | 10255;10239 | demon; undead predicate | R23d | Compatible SBA-R02g plus complete 2026-10-15 reading: monster valid row, definition bits 13 and 11/12; no draw |
| `LOD113C-D2COMMON` | 0x386E0; 0x382F0; 0x38220 | 10702;none;none | stat enumeration wrapper/copy/lower-bound | R23d | New complete 2026-10-15 instruction readings; aggregate-eligible wrapper, ordered entries, cap 128, no fMin arithmetic during copy |
| `LOD113C-D2COMMON` | 0x1A2D0 | 11104 | hireling classifier | R25b | Compatible GDF/DPO plus full2026-10-15 instruction reading: five class values 271/338/359/560/561 |
| `LOD113C-D2COMMON` | 0x6A050; 0x6A330 | 10677;10623 | definition hit-class; source flags | R26 | Complete2026-10-15 getter disassembly and matching source-flags reconstruction; zero on wrong-kind/missing-data where tested |
| `LOD113C-D2COMMON` | 0x4BE10 | 10232 | footprint-clear request | R24/Q11 | Compatible movement2026-10-15 caller/shape dispatcher reading; complete nested map mutation not established |

### Full dispatch chain and calling evidence

The movement owner locates per-frame action/root selection and callers of D2Game 0x3EDE0. At contact, EDI=missile, stack(game,target,force), cleanup 12. The local chain is0x3EDE0 ->0x3A4E0 (EBX=missile,stack target/record) -> optional damage callback ->0x3AFF0 ->0x3AE10 -> optional0xDE0C0 ->0xDCF20. The hand-off uses the same mutable record; no callback-return predicate intervenes after the damage callback.

Contact hit calls at 0x3EFC8 and 0x3F045 decode to `FF 14 85` absolute indexed calls through 0x10DA48: ECX=game,EDX=missile,stack target. The damage callback at 0x3F099 uses table 0x10DB68 with ECX=game,EDX=missile,stack target/record. The separate no-target entry's indirect call at 0x3E562 supplies the same game/missile and null target, then tail-jumps to 0xE20A0 without testing EAX. These argument positions are evidence mappings, not required APIs.

The final contact branch at 0x3F0B4 tests result bit 1; absent bit 1 and force 0 reaches0x3F106, returning 1 at 0x3F108. A clear stopMask bit 1 also reaches that return. The stopping path returns 2 at 0x3F0FA even with no room. Miss branch returns 2 at 0x3EFDA, except its callback bit 4 reaches the return 1 epilogue. The normal hit callback bit 4 also reaches 1. Therefore completed damage and rejected contacts can both return 1. Caller0x3F320's no-candidate branch calls contact then overwrites EAX with 2; some failure/expiry paths propagate it. The contact result alone does not prove every caller continues/removes identically.

### Original handler tables

Both counts and every table cell were read directly from the matching original image on 2026-10-15. Hit table RVA 0x10DA48 has 71 dwords, followed by count 71 at 0x10DB64. Damage table 0x10DB68 has 31 dwords, followed by count 31 at 0x10DBE4. Indices are signed i16; compare only after positive check. Null selected cells are unguarded calls, not skipped callbacks. This establishes tables and dispatch semantics, not text/compiled-row bindings.

| Table | Non-null slot -> handler RVA |
|---|---|
| hit | 1 -> 0x3b990, 2 -> 0x3e330, 3 -> 0x3bb60, 4 -> 0x3f220, 5 -> 0x3c700, 6 -> 0x3bd10, 7 -> 0x39fd0, 8 -> 0x39f70, 9 -> 0x3e0b0, 10 -> 0x3b140, 11 -> 0x3f120, 12 -> 0x3d020, 13 -> 0x3b3e0, 14 -> 0x3def0, 15 -> 0x3cdd0, 16 -> 0x3ec10, 17 -> 0x3eae0, 18 -> 0x3ea60, 19 -> 0x3e8a0, 20 -> 0x3dd60, 21 -> 0x3abd0, 22 -> 0x3db30, 23 -> 0x39da0, 24 -> 0x3b7c0, 25 -> 0x3da60, 26 -> 0x3c990, 27 -> 0x38ec0, 28 -> 0x3e740, 29 -> 0x3c850, 31 -> 0x3a920, 32 -> 0x3be70, 33 -> 0x38ea0, 35 -> 0x3e670, 36 -> 0x3ab30, 37 -> 0x38e90, 38 -> 0x3c5b0, 39 -> 0x3bc00, 40 -> 0x3c490, 43 -> 0x39e30, 44 -> 0x3b5f0, 45 -> 0x3c3b0, 47 -> 0x3d8c0, 48 -> 0x3c260, 50 -> 0x391a0, 51 -> 0x3c160, 52 -> 0x3c030, 53 -> 0x3f7d0, 54 -> 0x3bc80, 55 -> 0x396e0, 56 -> 0x3b210, 57 -> 0x3bbb0, 58 -> 0x3bed0, 59 -> 0x3d800 |
| damage | 1 -> 0x39ce0, 2 -> 0x39080, 3 -> 0x39670, 4 -> 0x38e70, 5 -> 0x3a790, 6 -> 0x39c80, 7 -> 0x39600, 8 -> 0x38ff0, 9 -> 0x39560, 10 -> 0x39040, 11 -> 0x39bd0, 12 -> 0x3a850, 13 -> 0x3aaa0, 14 -> 0x39a40 |

Hit null slots:0,30,34,41,42,46,49,60..70. Damage null slots:0,15..30. Slot0 is excluded by the gate despite being null. No claim is made that an original prepared row actually chooses every slot; that needs its own data/binding record.

### Resolved imports

All locations below belong to `LOD113C-D2COMMON`; the first location is the importing D2Game thunk, followed by original ordinal and resolved target RVA. Existing-stat/state primitives consume NRF's current contract. Setter/attach/map integration is explicitly open.

| Semantic operation | Game thunk | Common ordinal | Common target RVA | Boundary |
|---|---|---|---|---|
| ordinary stat read |0xA42A|10973|0x38B70|NRF-R12b/R12c|
| alternate non-null stat read |0xA802|10910|0x39430|SBA's checked wrapper/NRF values|
| replacing stat request |0xA472|10887|0x3A740|NRF-Q20/STL-R06c integration gate|
| state test |0xA68E|10494|0x33CD0|NRF-R12a|
| active state request |0xA3CA|10945|0x34510|STL-R07/NRF update gate|
| last target match |0xB27C|10227|0x6A130|Local guarded equality|
| last target setter |0xB0F0|10024|0x6A1B0|Local guarded stores|
| stat 180 enumeration |0xA7D8|10702|0x386E0|0x382F0 ->0x38220 existing-array copy|
| demon predicate |0xA814|10255|0x1A610|Valid monster definition flag|
| undead predicate |0xA508|10239|0x1A5B0|Valid monster definition flags|
| hireling class |0xA7F6|11104|0x1A2D0|Fixed five classes|
| missile hit class |0xB276|10677|0x6A050|u8 definition value|
| missile source flags |0xB108|10623|0x6A330|Matched guarded getter|
| room |0xA478|10331|0x2FE10|Supplied footprint request input|
| missile size |0xA71E|10589|0x30F80|Current GDF/movement getter boundary|
| footprint clearing |0xB258|10232|0x4BE10|Request only, nested map effects gated|
| allocate list |0xA6A0|11013|0x38100|STL-R06a initialization|
| list state |0xA5B6|10785|0x38060|STL-R06a field setter|
| list callback |0xA75A|10911|0x37D70|STL-R06a field setter|
| full attach |0xA60A|10807|0x3A7D0|STL-R06b gate|

### Semantic/binary mappings and decisive writes

| Semantic term | Original field/call | Supported rule |
|---|---|---|
| definition/selector | stride0x1A4; flags+4; hit/damage i16+0x0E/+0x10 |R13/R21/R25|
| creator | Unit+0xC8 bit 10;kind/id+0x94/+0x98 |R13/R26|
| registered chain | Original table 0xFB014 supplies game offsets0x1120,0x1320,0x1520,0x1920,0x1720 for kinds 0..4;id&127;kind 5 head+0x1B20;node id+0x0C,next+0xE4 |R13b|
| contact selection | missile Unit+0x10;filter table 0x38FDC |R15c|
| saved target | MissileData+0x18/+0x1C |R15b|
| flags A/B | MissileData+0x14 bits 0/1 -> DamageRecord.flags 32/128 |R26|
| cooldown | Missiles+0x188/+0x189; missile game+0x80; game frame+0xA8; callback0x38E60 |R20|
| record zero/base |28 dwords=0x70 bytes; physical+8;fire+0x10;magic+0x20;lightning+0x1C;cold+0x24;poison+0x28;burn+0x14 |R23a/b/f|
| record lengths/drains |coldLength+0x30,poisonLength+0x2C,manaBurn+0x3C,lifeLeech+0x38,staminaBurn+0x40,burnLength+0x18,stunLength+0x44 |R23c|
| critical/bypass |result word+4 OR0x2000;flags+0 OR0x100/0x200/0x400 |R23e|
| target type |MonStats stride0x1A8,type signed word+0x1C;demon/undead byte+0x0D bits 5/3/4 |R23d|
| physical pct |0x3A6FF floor-90;0x3A70E low32 IMUL;0x3A716..0x3A724 signed truncation /100 and add |R23d|
| record apply metadata |reductionScale+0x54,hitClass+0x60,reservedCondition byte+0x64 |R26|
| defender call |0xDB790:EAX=(physical!=0),ECX=creator,ESI=target,stack(game,1) |R26|
| direct process/results |0xDE0C0 and 0xDCF20 use creator,not missile,as attacker; processing flag 1 at first; mutable record common |R26|

The per-element roller receives EAX=max stat,ECX=percent stat,stack(missile,min stat). Its mulDiv uses EAX=pct,EDX=endpoint,ECX=100. Physical target adjustment uses its own wrapped product instead. Stat180 enumeration copies raw entry values from the aggregate array when the list has the aggregate-eligible condition; no minimum-value transform is applied during copying. The relation helper requires both positive indices strictly below typeCount and tests row targetType/column entryLayer; diagonal equality is not its predicate.

## Corrections and affected-claim review

| Affected assertions | Historical discrepancy and supporting/conflicting evidence | Resolution / remaining limits |
|---|---|---|
|R14,V08,V10/V11/V12|2026-10-12 prose called 1 rejected and 2 processed; source's final branches allowed successful damage then 1|2026-10-15 full dispatcher image establishes continue 1/removal 2 and exact branch conditions. V08 changes2->1; V10 positive-pierce changes2->1. Force/no-target/miss remain branch-specific; movement caller may override|
|R13b/Q14|Lookup wholly encoded, previously unresolved|2026-10-15 original image/table establish supplied registered-chain selection/first match. Invalid kind dereferences null rather than returning absent. Registry lifecycle remains outside-boundary; no creator lookup recursion invented|
|R15c/Q08|Unchecked encoded kind filter|Compatible new movement original-image/table record establishes 0/1/2/unsigned other, including negative raw selectors as other. Producer remains MCI|
|R21b/R25c/Q09|Indirect encoded calls/null behavior unresolved|Complete original-image/table checks establish signed gates/arguments and null call faults. Actual row binding and individual bodies remain outside-boundary|
|R22/Q10|No-target callback/remove entry asm-only, no known xref|Full image establishes callback result ignored then unconditional removal request. Missing real caller does not gate supplied-entry behavior; actual trigger remains outside-boundary|
|R23a/f/Q15|Physical/fire encoded base unresolved; Blade's base evidence previously diverged|Full matching 688-byte image corroborates zeroing, max 22,min 21,Rand,physical store and fire/magic arguments; unchanged numeric V14/V15 values retained. New complete target/critical read extends support beyond Blade's base-only slice|
|R23d|Text said a stat 180 layer matches numeric target type|Full 0x01620 predicate reads relation matrix with positive/count bounds; different layer can match and equal layer can fail. V19b discriminates. Matrix producer correspondence remains outside-boundary|
|R23d arithmetic|Unresolved encoded low32 physical percentage|Original image0x3A709..0x3A724 shows wrapped product then signed /100; V19d distinguishes NRF-R05/widened product. Critical doubles afterward|
|R23b range|A universal positive min..max range could be inferred|Original guards precede percentage; no post-adjust swap; V19c yields-5 with no draw for inverted adjusted endpoints|
|R20/V13|Creation/expiry phrased as an already active timed state|Full cooldown preparation checks allocation; schedule precedes tags/attach/state, event 0 follows even allocation failure. Zero-delay list requested expiry differs from scheduler stored frame. Actual lifecycle effects retain STL gates|
|R24/Q11|Unidentified footprint call called a stop|Caller request arguments and final result verified; map mutation remains gated under movement's incomplete nested-shape read|
|R25b/V17/V18|Actual armor/stat persistence asserted; lost owner implied no later writes|Setter actual propagation is unresolved in NRF/STL. Local replace request verified, armor follows even early no-owner direct apply. Original70/0/unchanged examples preserved as supplied-writer fixtures|
|R26|Direct contact could be conflated with DPO's queued reaction placement|Full call order uses creator as attacker; block/rewrite, class/suppression, owner double, source flags,scale,conditional process then unconditional results. V21 asserts mutation/order and final result flags|
|Q01/Q12|Stat names and null attacker behavior unresolved|Complete named-cell record below and narrow DDE90 null guard close those questions. ToHit absent creator returns miss 0 before any draw; downstream SBA's non-null formula retains its own gates|

Creation references Q02/Q04/Q06 now point to current MDP-R06/R04/R10. Other creation reference stubs, V01..V07 and O05 retain their destinations; they define no local duplicate behavior. Unconsumed status was established by repository search for filename and prefix in code/tests and implementation plans outside gaps; no consumption hit was found. Stable local ids are retained without a Formerly column; preserved original private records carry history/fixture encodings.

## Named table cells

Date2026-10-15; effective `data/global/excel/itemstatcost.txt` read through the recorded original LoD 1.13c launch profile's Storm loader. This reuses the coordinated complete body and profile in the current [pipeline table record](combat-damage-pipeline-order.md) and [Blade evidence](skill-assassin-blade-sentinel-and-fury.md#source-profile-and-coverage), rather than claiming a new collection. Request offset 0/default whole-file read returned 40109 of 40109 bytes, complete header/body. Effective SHA-256 is `ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`. The recorded inventory distinguishes this launch profile from the running PD2 entries; the version-console refresh failed, complete current override/archive inventory and compiled correspondence are uncaptured, and the earlier redirection-disabled record does not establish current pristine archives. Physical supplying MPQ unestablished. No gameplay confirmation follows from this text read.

Header-named columns ID/Stat were parsed from decoded bytes. Explicit ID equals zero-based data-row index excluding header for each cited cell. Each following entry is a named `ItemStatCost.ID/Stat` cell, labeled **verified in 1.13c data**; names support section 3, while code fixes the numeric consumers.

| ID/data-row index | Column Stat, raw cell |
|---|---|
| 25 | `damagepercent` |
| 56 | `coldlength` |
| 57 | `poisonmindam` |
| 58 | `poisonmaxdam` |
| 59 | `poisonlength` |
| 66 | `stunlength` |
| 120 | `item_damagetargetac` |
| 122 | `item_undeaddamage_percent` |
| 156 | `item_pierce` |
| 166 | `skill_pierce` |
| 180 | `damage_vs_montype` |
| 315 | `firelength` |
| 316 | `burningmin` |
| 317 | `burningmax` |
| 326 | `poison_count` |
| 327 | `damage_framerate` |
| 328 | `pierce_idx` |
| 332 | `passive_pois_mastery` |

The other named identities consume GDF/NRF's qualified records. Missile definition names/widths consume GDF-R07 and its loader descriptors; no actual Missiles row value is needed for the synthetic vectors. No partial-table absence or invented original row binding is asserted.

## Gap assessment and review boundary

All original local questions were assessed against existing compatible contracts and new targeted reads. Q01 names, Q08 filter, Q12 null attacker, and encoded instruction portions of Q09/Q14/Q15 are resolved to the supplied boundaries above. Q09's real callback/content binding, Q10 reachability, Q11 complete map writes, Q14 registry producer, Q15 type matrix producer and Q16/Q17 lifecycle integration remain explicit in the linked gap file. Creation questions stay with MDP. No six-rung absence claim is made. No gameplay observations, implementation comparison or other-version analysis were requested or performed.

Writer self-review covers semantic input/output placement, exact numbers, callback mutation/result handling, supplied-stage versus actual effects, code-origin bindings, original artifact locators, private preservation and open-gate scope. FULL1 independent review on 2026-10-15 recomputed all 16 local vector families and 41 leaf cases from the contract/dependencies, excluding creation reference stubs. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 8, COVERAGE 16, HARDEST 3. The triage below adds focused cases; a fresh full review is required for this resulting revision. Neither review counts nor evidence status imply maintainer approval.

## Observation procedures

No MHD-O01..MHD-O04 capture exists. The2026-10-12 unavailable-member attempt did not observe behavior. The following are planned tool-neutral procedures; equipment/skill selection must first be checked to produce the listed prepared inputs. An attempted procedure without eligible inputs or valid capture is inconclusive.

- **MHD-O01:** Use an owned vanilla 1.13c character of any class/level able to cast the selected damaging missile, on Normal outside town facing an ordinary monster. Select a row with known nonzero ranged element pairs and capture those current missile stats plus seed. At the fill entry/return (D2Game 0x3A4E0), record target predicates, every consumed stat/mastery, seed before/after and filled record before damage callback. Recompute under R23, including zero/equal/percentage cases. Expect one draw only for each positive wrapped range in the defined order. A skill name/visual lightning alone does not establish those inputs.
- **MHD-O02:** Use any character able to fire a prepared Pierce missile with owner stat 156 or 166 nonzero, with several ordinary monsters aligned outside town on Normal. Capture initial missile 328 and each eligible contact at 0x3EDE0/0x38F40; confirm target gates pass. Read requested and actual stat 328, stopMask, final contact return and movement caller result. Expect decrement request per nonzero counter, continue 1 for stopMask2, and stopping 2 at zero only when result/force requests it. Life/expiry/map effects require separate assertions.
- **MHD-O03:** Use any eligible character/level able to create a missile carrying stat 120 nonzero. On Normal outside town, compare ordinary monster class 5 with a hireling in a separately valid contact setup; record armor 31 before, missile 120, classifier and request at 0x3AFF0, then actual writer effect. Expect wrapped/floored request for ordinary monster and none for hireling/player. If no such original row/skill can be established, do not fabricate eligibility; record the missing content binding.
- **MHD-O04:** Use any class with a suitably slow missile outside town on Normal, a live ordinary monster target and a disposable original-game creator whose removal can be performed through gameplay. Record creator reference/registry lookup while missile is in flight and again at 0x3AE10. When creator lookup is absent, expect no record preset or process/results call. Observe callback and later armor requests separately; owner loss alone does not rule out their effects. The unavailable natural creator-removal setup remains inconclusive until established.

For every procedure record actual character/level/items, skill/row, build/profile, difficulty/location, frame, record count, inputs, expected/observed assertions and unobserved branches. No procedure changes a vector's runtime_observed flag until matching attributable observations exist.


## FULL1 advisory triage

The 2026-10-15 full reader found no blocking mismatch, guess, gap or discrimination failure. Each advisory received the outcome below; all added examples are synthetic supplied-input cases. The resulting contract has 16 local vector families and 59 independent leaf cases. No new producer, stat/map/list lifecycle or gameplay evidence is claimed.

| Coverage finding | Outcome | Rationale / cases |
|---|---|---|
| R13a absent/invalid definition | Add V08b/c | Distinct contact-entry fault before a return, with invalid signed class and absent missile fixtures. |
| R13b no match/kind 5 | Add V23d; no change for no match | Kind 5 must select the separate chain. V18 already supplies an empty registered chain and gets no creator; a nonempty nonmatching chain has the same absence outcome. |
| R15b new save/absent setter | Add V09d; no change for new save | V09c already proves a new target is saved before hostility rejection. Isolated null-target setter now explicitly retains kind and clears id. |
| R15c selectors 0/2 | Add V23e/f/g | Explicit reject-all and monster-only accept/refuse cases supplement existing player-only refusal and unsigned-other acceptance. |
| R15d friend pass/no owner | Add V09e; no change for no owner | False hostility with CollideFriend set proceeds. Existing V18/V23c use absent creator and reach fill without a hostility call. |
| R16 fallback/negative counter | Add V10c | Zero skill-pierce enables item-pierce read; nonzero negative counter decrements and suppresses stopping. |
| R18 successful hit | Add V11c | Supplied successful hit-roll call proceeds to event/fill; this stage does not claim zero real SBA hit-roll draws. |
| R19 no target/no force | No change | V08 already asserts this exact branch and final continue 1 without callback/fill. |
| R20 absent target | Add V12c | AlwaysExplode reaches the callback with absent target, no allocation, and null-owner event dispatch invoking no registered callback. |
| R21b other null slots | No change | V20d proves the eligible-null call fault; full original table identifies every other null. Their observable failure is the same recipe. |
| R22 doCall0/invalid definition | Add V22b/c | Disabled call avoids definition access and removes; enabled call faults before removal for invalid class. |
| R23a nonpositive/reversed | Add V19e/f; no change for max 0 | Existing zero absent physical stats already produce zero/no draw; new cases discriminate initial swapping and positive-max/zero-min guard. |
| R23b nonpositive/reversed/nonpositive range | Add V19e/f; no change for another nonpositive wrapped range | New fire cases cover swapping and min guard; V19c already proves nonpositive adjusted range skips Rand without a second swap. Other nonpositive ranges use the same Rand recipe. |
| R23d selection/bounds/cap | Add V19g/h/i | Supplied arrays/count/relation distinguish layer bounds, 128-entry cap and non-resolved-list refusal. Producer/stat propagation gates remain. |
| R25c negative selector | No change | Signed eligibility 0<d<31 is explicit; V16a/c distinguish skipped invalid selection from V16b eligible null fault. Negative skipped selector has the same no-call outcome. |
| R26 other early guards | No change | These all return before the common record writes/operations. V18 proves that unchanged-record/no-apply recipe; new contact-definition fault cases do not assert direct-apply faults. |

| Hardest finding | Outcome | Rationale |
|---|---|---|
| R23b/R23f | No rule change; extend V19e/f | Original endpoint guards and per-endpoint mastery precede shared-seed draws; V19a already discriminates seven-slot ordering, V14 mastery-before-roll, V19c no post-adjust swap. |
| R23d | No rule change; extend V19g/h/i | Existing V19b/d already separate relation from equality, clamp order and wrapped product; new cases supply array eligibility/bounds/cap inputs explicitly. |
| R21a/R24 | No change | V20a/b/c test complete result replacement and final return 1/2, including force not overriding callback mask 4. V10 tests pierce-suppressed final stopping. |

All eight documented gates receive **no change**: Q09 callback bodies/row bindings; Q10 real separate-entry callers; Q11 actual footprint writes; Q13 gameplay observations; Q14 registry producer; Q15 matrix producer; Q16 actual stat storage; Q17 actual cooldown lifecycle. Their classifications, blocked assertions and supplied-input limits remain in section 7 and the gap file; none was treated as a new reader defect. MISMATCH/GUESS/GAP/DISCRIM had zero entries and therefore required no individual correction.

## Full blind-reader pass 2 — 2026-10-15

The fresh reader recomputed all sixteen local executable families and 59 independent cases, excluding MDP reference stubs, from rules and supplied inputs without using traces as calculation inputs. Final counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 13 grouped limits, COVERAGE 18, HARDEST 3. The 13 limits comprise the eight local questions, PVP/SBA/callback/direct-apply dependency boundaries and the common supplied-fixture boundary. These retain their existing assertion-level gates, not a blanket downgrade of verified consumers.

| Advisory entry | Disposition |
|---|---|
| R13b no match/invalid kind | No change: V18 supplies an empty registered chain and consumes absence; a nonempty nonmatching chain has the same unchanged/no-owner recipe. R13b explicitly specifies the invalid-kind fault, distinct from absence; V08 and V22 already exercise the general controlled semantic-fault boundary without requiring unsafe pointers. |
| R15b LastCollide clear | No change: the guarded setter has no write when its flag is clear. V09 tests its writes/refusals with the flag set; the no-write recipe is expressly defined and no additional storage effect is inferred. |
| R15d no-creator hostility skip | No change: V18 and V23(c) already have no creator and reach filling/armor or hand-off. A false default hostility result must not reject those supplied no-owner cases. |
| R16 Pierce disabled/no eligible creator | No change: R16 retains stopMask 3 for these guards. V20(b) exercises non-piercing final stopping; V18 exercises missing creator and later work. No unknown counter update is supplied for a failed guard. |
| R18 present-creator miss | No change: the common supplied miss branch is V11(a)/(b), while V11(c) pins the present creator and successful result. Actual non-null hit-roll draws/failure belong to SBA; no duplicate hit-roll matrix is added. |
| R21a negative selector/h70 | No change: signed eligibility is explicit, negative takes the skipped-selector recipe exercised at 0/71 in V12; eligible 70 is a null call under the complete table and the same failure recipe exercised at 30 in V20(d). No null fallback is permitted. |
| R21b other null slots | No change: V20(d) discriminates unguarded eligible-null invocation; the checked complete table identifies the other null cells. Repeating the same semantic fault for each slot adds no new result handling. |
| R23a physical max nonpositive | No change: absent/default stat 22 is 0 in the zero-filled cases and suppresses the physical draw; V19(f) separately tests positive max with zero min. Both endpoint guards and their no-draw result are available. |
| R23b elemental max nonpositive | No change: unused element pairs have default max 0 and no draw in V14/V19; V19(f) separately exercises positive fire max/zero min. The fixed per-slot sequence remains pinned by V19(a). |
| R23c positive poison division | No change: V15(a) already gives poisonLength 200/count 3, result 66. V15(c) adds negative truncation; this advisory identifies an already exercised branch. |
| R23d absent resolved list/nonpositive type | No change: V19(i) explicitly refuses enumeration when the list is not eligible, retaining physical 100; V19(g) pins strict positive/count entry bounds. An absent list or nonpositive type uses the specified no-enumeration/zero-extra recipe; real list/matrix production remains gated. |
| R23d false relation | No change: V19(b) already supplies relation(2,2)=false for value 90 and relation(2,1)=true for value 10. Its output rejects equality matching and inclusion of the false-relation entry. |
| R25b other hireling classes | No change: the checked fixed five-class classifier and V17(c) establish the excluded-hireling no-write recipe; the remaining fixed class identities use the same classifier result, not different armor arithmetic. |
| R26 wrong missile kind | No change: R26 requires kind 3 before record preparation. The same unchanged-record/no-apply early-return recipe is V18, and the guard itself is explicit; no partial combat effect is invented. |
| R26 invalid owner reference/definition | No change: failed direct-apply requirements return before common record writes. V18 exercises that unchanged/no-process/results recipe. Contact-entry invalid-definition faults in V08 do not become direct-apply fault claims. |
| R26 state 54 | No change: reaction preparation consumes HRC-R01b/SBA-R20 and their owning conditions/vectors. MHD-V21 pins local shared-record order; adding the dependency matrix would duplicate those owners. |
| R20 callback removes another event | No change: event/list traversal and registered callback effects remain with IEP/STL and Q17. This caller orders the event request and ignores its final result; it does not invent mutation behavior for the delegated event dispatcher. |
| R13a Explosion through contact | No change: R13a defines initial result 0 for Explosion, and V12(b) exercises it with force; movement MMC-V10(d)/(e) already discriminates ordinary Explosion contact without and with result replacement. MHD-V20 owns callback replacement/final masks; no additional producer effect is implied. |

HARDEST, all no change: V19 explicitly orders all seven seed draws, resets the record and distinguishes endpoint, relation, wrapping and cap behavior; V20 asserts complete callback-result replacement and final continuation/removal returns; V21 asserts shared-record mutations and ordered block/class/double/process/results operations. Every applicable blocking finding is resolved on the resulting contract. No rule, input, vector, expected output, acceptance or boundary is changed in this final triage, so another reader run is not required. Independent documentation review is complete; draft status, exact integration/observation gaps and pending maintainer approval remain. Publication checks are repeated without implementation tests or gameplay.
