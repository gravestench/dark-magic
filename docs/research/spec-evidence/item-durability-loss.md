# Evidence: durability loss

Spec: [Durability loss](../../specs/item-durability-loss.md)

## Evidence identity and reuse

The evidence build is original Lord of Destruction Expansion 1.13c, Windows x86. Artifact ids below resolve through the [source manifest](../source-manifest.md): `LOD113C-D2GAME` is D2Game.dll, 1138688 bytes, SHA-256 `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`; `LOD113C-D2COMMON` is D2Common.dll, 679936 bytes, SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`. Both available analysis specimens were rehashed on 2026-10-15 and matched. RVAs are relative to the module, not file offsets; preferred bases are D2Game `0x6FC20000`, D2Common `0x6FD50000`. Descriptions below are analyst labels, not original exported names.

Compatible local instruction readings dated 2026-10-12 retain their dates and coverage. They used disassembly, resolving imports through original PE import/export tables; readable decompiler labels alone were not evidence. Their original complete record and binary fixtures were privately preserved before migration. Relocation does not renew verification. Historical loaded runtime DLL hashes were not captured; current analysis-image fingerprints do not retroactively establish those loaded identities. Effective data hashes and overrides were also not captured historically.

Reuse assessment on 2026-10-15: [DPO evidence](combat-damage-pipeline-order.md) and DPO-R24a/V17 establish queued placement, pre-apply hit gate and local draw order with supplied item results, but DPO-Q12/Q27 leave break/item producers open. [NRF evidence](core-numeric-and-rng-foundations.md) and RLI-R02a/RLI-R02b establish selected-array reads; RLI-Q03 explicitly does not establish replacing mutation. [GDF evidence](core-game-data-fields.md) establishes calc schema and limited type identities, not full equivalence or item generation. [SCD evidence](skill-cast-start-and-do-dispatch.md) supplies the start-table root and selector contract; [SCX evidence](skill-calc-expressions.md) owns calc execution with its own gates. These records were reused only for their stated boundaries and actual dates.

New readings on 2026-10-15 used reconstruction checked against original function bytes when available; otherwise disassembly. Original-image instruction decoding independently corroborated relevant embedded instruction runs in the loss entry, initial schedulers, timer scheduler/dispatcher, primary selector and equipment callbacks. No gameplay action or trace was performed during this migration. Unknown local callback effects remain gated rather than being inferred from labels.

## Dispatch and output anchors

Combat root (reused DPO chain): D2Game `0x79A10` → frame-event table `0x1079D0` → `0x79170` → action entry `0xA2060` → dispatch `0xA19A0` → action table `0x1074A8` slot 1 → basic attack `0xA2C40` → queued completion `0xDDDE0` → mitigation `0xDC0B0` / enqueue `0xDAB90` → queue consumer `0xDED30`, which calls apply `0xDE0C0` and then, as a sibling call, durability `0xDEBA0` → loss `0x12EE0` → replacing-write import D2Common #10887, `0x3A740`. At `0xDED30`, the pre-apply hit test gates apply, marks and durability; events 7/3 follow. DUR-R02 consumes DPO placement rather than claiming new verification of the whole graph.

Within `0xDEBA0`, player attacker processing precedes player defender processing. The attacker calls primary selection `0x372C0`; defender armor lookup uses D2Common #11139 → `0x1E670`, type test #10744 → `0x24430`. Both reach `0x12EE0`. The weighted binary table has count 7 and the seven ordered location/weight pairs at D2Game RVA `0x10E628`; compatible DPO memory read dated 2026-10-13 covered 100 bytes. A new original-image read on 2026-10-15 corroborated the constant table and current image identity. Weight eligibility precedes qualification; chance is one inline seed step modulo 100.

Skill root (narrow new binding): SCD's D2Game start table RVA `0x107338`, slot 7 at `0x107354`, contains handler `0x4CC50`. Complete applicable handler reading shows successful-hit handling, selected primary weapon `0x372C0`, qualification, calc evaluation with record fields `+0x13C`/`+0x140`, then `0x4C9E0` with chance/amount in that order. `0x4CC50` also calls damage-related operations; those interiors are not newly verified here. Named Skills rows and whole hit-path integration remain DUR-Q03. The caller invokes D2Game `0x9F210` before the local durability call and queued completion `0xDDDE0` afterward; the earlier broad description of prior damage application does not establish these unread interiors. The binding establishes a start handler, not an action-handler selector or an original skill name.

Deferred scan `0x78A90` is called from melee entry `0x29600` twice, throw paths `0x4C840`/`0x4C870`/`0xA1600`/`0xA1630`, and player mode end `0x79850`. It tests inventory node kind 3, weapon type, qualification, durability<=0 and broken clear; it requests Neutral once, calls break `0xD5740`, then reloads the head. These legacy caller records retain their 2026-10-12 date; no fresh attack-graph verification is claimed.

Timer root: `0xEA70` and quantity counterpart `0xEA10` request event 3 through scheduler `0x8EA80`, default callback 0. New complete dispatcher reading `0x8EDD0` and matching original bytes show explicit callbacks bypass the default table. The default table is RVA `0xFB110`; slot 3 at `0xFB11C` contains `0xDA480`. On a due default event it passes game/item to `0xDA480`, which calls `0xDA350` for `(rate 252,target72,maxDur)` and, conditionally, `(rate 253,target70,maxStack)`. This settles the formerly inferred default selector binding. Scheduler allocation, event lookup/removal and complete frame-loop delivery remain DUR-Q04; selector verification alone does not prove actual cadence.

## Function references and boundaries

All ordinals in this table belong to the identified D2Common artifact. The imported targets were historically resolved from original PE tables. For same-module operations, the original function RVA identifies the target. The 2026-10-15 read boundaries below include entry/control flow through the request sites, not unknown callback interiors.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | `0xDED30`, `0xDEBA0` | — | Queued durability placement / weighted choice | DUR-R02/R03/R04 | Reused 2026-10-12 and current DPO placement/consumer records |
| `LOD113C-D2GAME` | `0x12EE0` | — | Combat loss | DUR-R05/R06/R07/R08a | Complete disassembly 2026-10-15; inline chance, subtract, clamp, schedule/write/notify/break/refresh request sites |
| `LOD113C-D2COMMON` | `0x26D40` | 10865 | Qualification | DUR-R01/R14 | Reused 2026-10-12 disassembly; valid record, bytes `+0x113/+0x112`, maxDur, stat 152; getter contracts reused |
| `LOD113C-D2COMMON` | `0x24430`, `0x24C30` | 10744, 10711 | Type equivalence / throwable getter | DUR-R03/R04/R05/R14 | Original targets; producer integration remains DUR-Q06; no unread equivalence promotion |
| `LOD113C-D2COMMON` | `0x221C0` | 10718 | Fallback selector accessor | DUR-R08a/R08b | Complete byte-checked reconstruction read 2026-10-15; item assertions, unsigned word at ItemData `+0x30`; meaning still unknown |
| `LOD113C-D2COMMON` | `0x260F0` | 10463 | Maximum stack | DUR-R08a/R12a | Complete disassembly 2026-10-15; calls base accessor, adds stat 254, signed upper cap 511; selected-array stat dependency NRF |
| `LOD113C-D2COMMON` | `0x22B70` | — | Base stack accessor | DUR-R08a/R12a | Complete byte-checked reconstruction 2026-10-15; valid indexed item record stride `0x1A8`, reads i32 field `+0xE8` |
| `LOD113C-D2COMMON` | `0x787C0` | — | Item loader descriptors | DUR-R01/R12a | Reused 2026-10-12 loader record; 2026-10-15 relevant descriptor/string check maps maxstack to `+0xE8`, string RVA `0x8F364`; not a fresh whole-loader claim |
| `LOD113C-D2COMMON` | `0x72510` | — | ItemTypes loader | DUR-R05 | Reused 2026-10-12 descriptor record: Throwable byte `+0x10`; GDA-R14 limits retained |
| `LOD113C-D2COMMON` | `0x22E20` | 10048 | Stackable getter | DUR-R12a/R13a | Complete byte-checked reconstruction 2026-10-15; item record byte `+0x132`; record producer boundary remains explicit |
| `LOD113C-D2COMMON` | `0x38B70`, `0x3A740`, `0x38FF0` | 10973, 10887, 11116 | Stat read / replacing-write entry / maxDur | DUR-R01/R06/R08a/R09a/R12a/R13a/R13b | Read consumes RLI-R02a/RLI-R02b; replacing-write execution remains RLI-Q03, never inferred from getter |
| `LOD113C-D2COMMON` | `0x23940`, `0x23900` | 10202, 10977 | Item flag test / set | DUR-R07/R09a/R10/R12a | Legacy bit 100 call sites; stored callback effects remain gated |
| `LOD113C-D2COMMON` | `0x1E670`, `0x1E8C0` | 11139, 10698 | Body item / node kind | DUR-R04/R10 | Legacy call-site evidence; array/row and traversal integration DUR-Q06/Q07 |
| `LOD113C-D2COMMON` | `0x2EB70`, `0x33920` | 10920, 11090 | Player data / mode request | DUR-R06/R10/R12a | Client from player data `+0x9C`; Neutral mode request consumes RLI-R04 with gates |
| `LOD113C-D2GAME` | `0xD5740` | — | Break entry | DUR-R09a/R09b | Complete disassembly 2026-10-15; ordered callback/write requests; actual meanings/interior draws unresolved |
| `LOD113C-D2COMMON` | `0x1FC70`, `0x2F1B0`, `0x1E4F0`, `0x30DC0`, `0x2C910`, `0x665C0` | 10694, 10175, 10050, 10201, 10600, 10158 | Break preparation, predicate, completion, owner update, notification, finalization | DUR-R09a/R09b | Legacy resolved targets/call order; neutral descriptions distinguish operations, not verified effects |
| `LOD113C-D2GAME` | `0xD46B0`, `0xD51A0` | — | Conditional equipment transition / revalidation | DUR-R09a/R09b | Complete entry/control-flow inspections 2026-10-15; several subcalls unresolved, so no full effects/draw count |
| `LOD113C-D2GAME` | `0x78A90` | — | Deferred break scan | DUR-R10 | Reused complete local reading 2026-10-12; list-head reload and one mode request |
| `LOD113C-D2GAME` | `0x79100` | — | Weapon refresh entry | DUR-R08a/R13b | Complete disassembly 2026-10-15; subcalls `0x778A0`, `0x78DD0`, `0x77B20` remain unresolved integration |
| `LOD113C-D2GAME` | `0xEA70`, `0xEA10` | — | Initial durability/quantity scheduling | DUR-R11/R13b | Complete disassembly and original-byte decoding 2026-10-15; pending lookup `0x8E490` and scheduling request; actual scheduler lifecycle gated |
| `LOD113C-D2GAME` | `0x8EA80`, `0x8EDD0` | — | Scheduler / event dispatcher | DUR-R11/R12b | Complete applicable readings/original bytes 2026-10-15; default selector verified; allocator/removal and full delivery remain unresolved |
| `LOD113C-D2GAME` | `0xDA480`, `0xDA350` | — | Replenish entry / one-point regen | DUR-R12a | Complete disassembly 2026-10-15; rate, target, cap argument order, signed comparisons, owner/client-conditional refresh, floor 125 |
| `LOD113C-D2COMMON` | `0x224C0`, `0x51BF0` | 10438,10786 | Item owner identity / skill calc front end | DUR-R12a/R13a | Resolved from matching PE tables on 2026-10-15; owner production gated, calc front end consumes SCX-R11 |
| `LOD113C-D2GAME` | `0xD5690` | — | Item inventory refresh | DUR-R12a | Request site identified; actual interior effects/draws DUR-Q07 |
| `LOD113C-D2GAME` | `0x4CC50`, `0x4C9E0` | — | Skill hit caller / loss consumer | DUR-R13a | Complete applicable disassembly 2026-10-15; successful-hit request/calcs; whole producing handler dependencies not newly verified |
| `LOD113C-D2GAME` | `0xA1790`, `0xA16A0` | — | Repeated stack selector / decrement | DUR-R13b/R13c | Complete disassembly 2026-10-15; selector examines arm items, qualification/type/stat producers gated; decrement requests verified with supplied selected item |
| `LOD113C-D2GAME` | `0x372C0` | — | Primary selection | DUR-R03/R13c | Complete entry/control-flow inspection and original-byte decode 2026-10-15; skill/hand table and imports unresolved; no exact selected-item guarantee |
| `LOD113C-D2GAME` | `0x8B010` | — | Item-stat notification entry | DUR-R06/R08a/R09a/R12a/R13a/R13b | Reused call-site parameter evidence; wire/client implementation outside boundary, actual delivery gated |

## Semantic mapping and control-flow evidence

The original item broken bit is `0x100` in ItemData flags at `+0x18`. The selector is the unsigned 16 word at ItemData `+0x30`; its accessor supplies no semantic name. Item qualification tests record durability byte `+0x112` and no-durability byte `+0x113`, reads maxDur/stat 152, and rejects a missing record. Game expansion/frame are `+0x70/+0xA8`; unit seed is `+0x20/+0x24`. These are evidence mappings, not required storage layouts.

At loss entry `0x12EE0`, selector 0 returns immediately after durability-zero write/notification. The owner refresh `0x79100` occurs in selector-positive quantity<=1 branch only. That branch writes quantity 0 but passes old q in its stat 70 message. A selector-positive quantity>1 branch writes q-1 then maximum durability. Both reach the maxStack>0 conditional durability-replenish call. No quantity-rate scheduling occurs here.

Break `0xD5740`: item in EAX, owner in ECX, game in EDX. It calls #10977(item, 0x100, 1), #10694(owner inventory, item), #10175(item). If the latter returns 1, it passes stack arguments(game, 1, 1) to `0xD46B0` while owner/item remain in ESI/EDI. It calls `0xD51A0(game,owner,0)`, #10050(inventory, item), #10201(owner, 1), stat 72 write 0, notification 0, #10600(owner, item, 1, 1), then writes owner word `+0x6E=9`, owner `+0x70=owner`, calls #10158(owner), and ORs owner `+0xC4` with `0x400`. Unknown bookkeeping meanings stay in DUR-Q07. Local routing has no seed step, but reachable equipment subcalls and refresh entries are not established draw-free. Earlier reachability found potential draws through event scheduling/monster mode paths; this is neither proof of player-owner execution nor of zero draws.

`0xDA350` increments the target once; n>cap returns 1 before write/scheduling. On write, owning-unit lookup is performed through the item owner's id; only a found player with nonnull client reaches notification and the broken-item inventory refresh. Thus absent client suppresses refresh as well as notification. The first request uses 2500/rate+1; later requests floor at 125. `0x8EA80` inspection additionally shows requested frames<=current are normalized to current+1 before insertion, and its visible special draw route requires monster/event 2. Allocation/event-list integration remains unread, so neither actual negative-rate cadence nor a complete scheduler draw total is promoted.

`0x4C9E0` tests stackability before its local qualification branch. Its caller already qualifies the selected primary item. Its successful stack chance calls `0xA1790`, which selects again across arm items and then calls `0xA16A0`; the decremented item is not guaranteed identical to the chance-tested item. `0xA16A0` writes floored quantity, schedules rate 253, notifies, conditionally resets durability to maximum, refreshes weapon state, then returns original quantity unless negative subtraction was clamped. The helper returns zero after negative subtraction was clamped, otherwise wrapping n+1. Those are local requests with a supplied selected item; selection and actual effects remain gated.

## Table records

Historical record, 2026-10-12: effective `data/global/excel/itemstatcost.txt`, `itemtypes.txt`, `bodylocs.txt` were read through the owned 1.13c game's loader for names/row identities only. No ranges, effective hashes, full override inventory or original archive fingerprints were captured. Preserve that limited support; it is not a newly complete table-data record. DPO's compatible 2026-10-13 name reads and 2026-10-15 narrow table records separately support stat 72 and types 45/50, with their own limits.

Fresh read-only corroboration, 2026-10-15: the session inventory reported one `lod-113c` member and four modified profiles. The version/profile registry did not answer, so independent original-profile identity could not be completed. The reported 1.13c member's Storm loader returned each file in one requested range(offset 0, length 60000); all responses succeeded, returned lengths equaled total lengths, complete `base64` bodies were decoded and tab-separated headers selected by name. No gameplay/hook changes were made. File hashes identify these returned bytes only. Physical supplying archives, loaded-module fingerprints, override inventory and compiled-record correspondence remain unestablished; no original-profile or gameplay label is promoted by these reads.

| Effective virtual file | Returned / total bytes | SHA-256 of returned complete file | Selection / named cells |
|---|---:|---|---|
| `data/global/excel/itemstatcost.txt` |40109/40109|`ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`|Explicit ID column matches zero-based data rows 70, 72, 73, 152, 252, 253, 254; Stat cells below|
| `data/global/excel/itemtypes.txt` |8775/8775|`d399dc5d0f3a3e1a4f1eb611dabef8311b7025392c416f58dc26baff93af3550`|No explicit ID; zero-based data rows 45, 50, 51 excluding header, all before Expansion separator; ItemType/Code/`Equiv1` below|
| `data/global/excel/bodylocs.txt` |160/160|`6b3a6ac0bb1cef5e7f216f083ef85082acedf22c4bbfacf5b254fb36ecd8bbd5`|Zero-based data rows excluding header; Code cells below|

Each entry below is a single named cell, not a copied row/column:

| File / row identity | Column | Raw cell / interpretation | Supports |
|---|---|---|---|
| ItemStatCost ID 70 | Stat | quantity | R08a/R12a/R13b name |
| ItemStatCost ID 72 | Stat | durability | R01/R06/R08a/R09a/R12a/R13a/R13b name |
| ItemStatCost ID 73 | Stat | maxdurability | maxDur name |
| ItemStatCost ID 152 | Stat | item_indesctructible | R01/R14 name |
| ItemStatCost ID 252 | Stat | item_replenish_durability | R11/R12a name |
| ItemStatCost ID 253 | Stat | item_replenish_quantity | R12a/R13b name |
| ItemStatCost ID 254 | Stat | item_extra_stack | maxStack name |
| ItemTypes row 45 | ItemType | Weapon | Type identity corroboration |
| ItemTypes row 45 | Code | weap | Type identity corroboration |
| ItemTypes row 50 | ItemType | Any Armor | Type identity corroboration |
| ItemTypes row 50 | Code | armo | Type identity corroboration |
| ItemTypes row 51 | Code | shld | Shield code corroboration |
| ItemTypes row 51 | `Equiv1` | armo | Declared equivalence only, not equivalence execution |
| BodyLocs row 1 | Code | head | Location name corroboration |
| BodyLocs row 3 | Code | tors | Location name corroboration |
| BodyLocs row 4 | Code | rarm | Location name corroboration |
| BodyLocs row 5 | Code | larm | Location name corroboration |
| BodyLocs row 8 | Code | belt | Location name corroboration |
| BodyLocs row 9 | Code | feet | Location name corroboration |
| BodyLocs row 10 | Code | glov | Location name corroboration |

No real weapon/armor row value is required by the executable vectors: their record durability, stackability, throwability and maximums are synthetic supplied inputs. Numeric stat/type arguments and weighted locations are independently code-fixed consumers; finding their names in data does not imply remapping support. Loader descriptor field names are schema evidence, not named original item-cell values.

## Search records and unresolved boundaries

| Question | Attempt / date | Result and smallest remaining check |
|---|---|---|
| DUR-Q01 |2026-10-12 accessor/loader evidence; 2026-10-15 complete selector/base/cap accessors and targeted loader descriptor/string check|maxstack base column settled; u16 selector meaning not established. Trace its producer and original prepared correspondence without naming it from an accessor label|
| DUR-Q02 |2026-10-12 draw reachability; 2026-10-15 complete break/transition/revalidation/refresh entry inspections|No direct break-entry draw, unresolved subcalls remain. Read applicable player-owner subpaths or attributable break/refresh seed trace in a future campaign|
| DUR-Q03 |2026-10-12 handler pointer without table identity; 2026-10-15 original-image pointer scan, matching SCD start-table root, complete applicable skill caller|Start index 7 settled. Named original Skills binding/whole hit-path evidence still absent|
| DUR-Q04 |2026-10-12 stat-pair/type inference; 2026-10-15 original-image pointer scan plus complete dispatcher/scheduler entry|Default event 3→handler settled. Read lookup/allocator/removal/frame delivery owner for complete lifecycle; supplied requested frames remain available|
| DUR-Q05 |2026-10-12 immediate stat 72/252 pusher search, headers and relevant sites only; 2026-10-15 compatibility/scope assessment|27 functions checked, not complete death/writer investigation and not all six search rungs. No absence label justified; preserve outside-boundary unresolved question|
| DUR-Q06 |2026-10-12 primary getter left open; 2026-10-15 complete primary/repeated-stack selector entries and byte-checked stackable getter|Skill/hand/type/inventory producer callees remain incomplete. Supply results locally; complete exact producer chain to accept original selected item|
| DUR-Q07 |2026-10-12 break call order/unknown owner fields; 2026-10-15 local write/client/refresh requests against current RLI-Q03|Getter does not verify mutation. Complete propagation/notification/owner/refresh/traversal contracts; retain supplied request assertions|
| Other-build equivalence (question removed 2026-10-15 review) |2026-10-12 explicit 1.14d untested; 2026-10-15 evidence-build boundary assessment|No 1.14d reading/observation authorized; outside build|
| DUR-Q09 |2026-10-12 seed-delta observation; 2026-10-15 confound/attribution reassessment|No direct durability, damage or hit-recovery attribution; future attributable gameplay required, no migration gameplay|
| Named table correspondence |2026-10-15 compatible sibling reuse and complete read-only text reads with reported profile|Registry unavailable; historical original-profile/compiled identity not promoted. Authenticate profile and corroborate prepared records for needed named values|

Historical 27 immediate pushers: D2Game `0x06B54`, `0x0E410`, `0x0EA70`, `0x0EBB0`, `0x0F370`, `0x10A20`, `0x11070`, `0x11490`, `0x12EE0`, `0x137F0`, `0x13A30`, `0x13E60`, `0x36470`, `0x4C9E0`, `0x78A90`, `0xA16A0`, `0xAFCE0`, `0xBD9B0`, `0xBDE60`, `0xBE2A0`, `0xBF990`, `0xD36D0`, `0xD3CB0`, `0xD5690`, `0xD5740`, `0xE7560`, `0xE9600`. The historical author identified lowering sites `0x12EE0`/`0x4C9E0`/`0xD5740`, and raising helper `0xDA350` whose stat id arrives in a register. Headers/immediate sites do not cover indirect ids, full restore writers or a death dispatch root. Runtime rung was not performed for death; this record does not meet the six-rung absence standard.

## Observation records and procedures

| Vector | Source profile | Date | Module + RVA | Records | Result |
|---|---|---|---|---:|---|
| DUR-O04 | Reported owned vanilla 1.13c LoD, single player; loaded DLL/data identity limits above |2026-10-12|D2Game `0xDDE90`, caller `0xDE646`; D2Common `0x31D20` block chance|116 hit-roll records; 57 block-chance records; 61 resolved hit transitions|Partial seed-delta observation only; consistent, not attributable rule confirmation|

### DUR-O01

Pending. On an owned 1.13c Expansion game, use a player with a qualifying ordinary sword and a separate player fixture with one qualifying armor piece. Record class/level, difficulty, equipment and exact prepared stats. In a non-town area, perform 1000 eligible player-versus-monster attacks and monster-versus-player hits. At D2Game `0xDEBA0` and `0x12EE0`, measure both unit seeds before/after and record selected item, qualification, threshold, roll and result. Separate misses, blocks, other hit callbacks and any break. Compare local one-weapon and two-selection/optional-one-chance counts; do not infer them from whole-hit deltas alone.

### DUR-O02

Pending. Use the same fixtures, record starting durability and avoid reaching a break during each sample. Read chosen item and roll at `0x12EE0`, and durability at the replacing-write request. Count qualifying selected-item trials and losses separately for armor/ordinary weapon. Compare exact roll<10/roll<4 decisions per trial; near 10%/4% aggregate frequency is supporting only. Record effective game data and caps; do not substitute sample frequency for threshold evidence.

### DUR-O03

Pending. Record class/level, difficulty, a qualifying nonstack sword at durability 1, max>0, and its selector word. In a non-town area, trigger one successful local loss and inspect item at `0x12EE0` exit, before any later scan. Then trigger the next attack or mode-end scan at `0x78A90`, and observe `0xD5740`, equipment/owner stats and seeds around the callback chain. Distinguish stored 0, break request, flagged broken and actual usability. The complete effects/usability claim stays unresolved until read and observed directly.

### DUR-O04

Historical inputs: level 1 Barbarian, only equipped armor the starting Buckler; four Fallen attackers; owned LoD 1.13c, single player, 2026-10-12. Hit-roll capture at D2Game `0xDDE90` ran 60 seconds, cap 200, yielded 116 records, all caller `0xDE646`; each read the defender seed words. Block-chance capture at D2Common `0x31D20` yielded 57 records. Seed-transition search against NRF's step resolved all 61 hit transitions to 1 or 4 steps:34 blocked hits used 1, 27 unblocked used 4; misses used 0. A monster attacker has no attacker durability part. With only Buckler, weight 4, the expected local draws are modulo 7, mask 4, modulo 100.

No durability function was traced; chosen item, roll, durability changes and damage were not read. The additional 3 seed steps were attributed by inference. HRC-R06's get-hit veto can consume 0–2 defender draws; the count agreement assumes it consumed none in this window. No statically verified rule/vector is promoted. Reproduction: identify the same build/profile, equipment and low-level class in a non-town area with Fallen; capture both durability entries and selected item/seed/damage directly during unblocked hits, then correlate block/hit-recovery entries. Capture validity, record loss and exact entry/exit boundaries must be documented before attributing steps.

## Corrections, preservation and review

The complete pre-migration record is privately archived with integrity metadata. All 15 parent rule families, 13 vector families, nine questions, four observations and original numeric traces are preserved through current rules/vectors or explicit correction records here. No historical id aliases are retained because no qualifying implementation/plan consumption was found; useful ids remain unchanged, with sub-ids splitting mixed labels.

Corrections supported by 2026-10-15 readings:

- DUR-V10 originally claimed owner weapon refresh for selector 0. `0x12EE0` returns directly after zero durability write/notification there. The corrected selector 0 case requires no refresh, and an added selector-positive q=1 case demonstrates the actual refresh branch. The original durability 0/deferred-break numeric outcome remains; complete break effects remain gated.
- DUR-R13 originally described spending the chance-tested stack item. `0x4C9E0` invokes a second selector `0xA1790`; the current contract exposes supplied selected item and gates original selection. V13's distinct-item case discriminates that unsupported identity assumption.
- DUR-R12 originally placed broken-item inventory refresh outside client eligibility. `0xDA350` nests it inside successful owner-player-client notification; V12 now distinguishes client absent. No changed numerical regen values.
- DUR-R15's old “not found” label had only 27 immediate-pusher inspections, not all six rungs. It is unresolved; the historical unsuccessful death inference is retained above, without a universal no-loss claim.
- DUR-R08/R09/R12/R13 mixed evidence is split. DUR-R09 actual equipment/owner effects and all replacing-write persistence are explicitly gated; stat getter evidence is not mutation evidence. First/later rate 5 delays 501 and rate 25 delays 101/125 remain exact.
- DUR-R01/R14's earlier blanket statements about no loss/break/replenishment on every path are limited to qualification-gated local operations. The standalone stack consumer and quantity fallback do not justify a universal ban; the current contract follows their explicit branches. Ethereal equivalence is limited to direct local consumers and does not claim unread equipment callbacks ignore all other flags.
- Historical DPO-R24a once reversed weapon/armor thresholds. The 2026-10-12 loss entry showed armor 10, weapon 4, Expansion throwable 10; the sibling corrected those values then. Current contract agrees. An unrelated historical NRF “up to 10 draws” lead was corrected by sibling evidence to one step; no ten-draw loop is supported.

### Fresh reader FULL1, 2026-10-15

Full review covered all 13 vector families and 38 explicit cases. Counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 2, GATED 30, COVERAGE 11, HARDEST 3. The report used the behavioral contract and its declared dependencies; it supplied no new original-game evidence. Pre-fix files and the complete relayed report were preserved privately. Outcomes:

| Finding | Disposition |
|---|---|
| GAP1: V12 explicit callback had no supplied callback input | Spec edit: V12(h) supplies due event 3 with explicit callback K, whose only effect is logging K; requires no regen, stat writes or event requests |
| DISCRIM1: V10 unchanged A/B list allowed a forward-only scan | Spec edit: V10(a) supplies break A that marks A broken and prepends eligible C; exact log Neutral 1→break A→break C→break B rejects forward-only traversal while retaining A-before-B and once-only mode assertions |
| DISCRIM2: V12 explicit callback bypass was not distinguished | Spec edit: V12(h) logs exactly K and leaves durability 14, separating explicit callback execution from default regeneration |
| GATED30 | No change: these are the documented producer, mutation, callback, scheduler, observation and other-build limits; no supplied fixture promotes their actual integration |
| COVERAGE1: R01 non-item/absent-record guards | No change: each is a false qualification result with no draw/effects, the same output recipe as V06(b-e). This does not claim those exact guard bindings were exercised |
| COVERAGE2: R04 W<=0 | Spec edit: V06(h) supplies a player inventory with empty locations and requires W=0, no selected-item call and unchanged S1 |
| COVERAGE3: R05 neither/both type results | No change: neither type uses the existing failed-eligibility no-draw/no-request recipe; both true selects R05's armor threshold 10 and has the same positive-loss arithmetic as V02(c), irrespective of the weapon result. Downstream zero routing still uses the separately declared weapon flag; no type-equivalence producer is inferred |
| COVERAGE4: R07 already-broken nonweapon | Spec edit: V10(c) supplies broken armor reaching zero and requires selector-zero write/notification with no repeated break |
| COVERAGE5: R08a q<=0 old-quantity notification | No change: V10(b) discriminates old q from stored 0 on the q<=1 recipe. Supplying another q<=1 changes only the explicitly forwarded old quantity, with no additional arithmetic or branch; exhaustive negative bindings are not claimed |
| COVERAGE6: R09a predicate 0 | Spec edit: V09(b) supplies predicate 0 and gives the exact break log omitting only the conditional transition |
| COVERAGE7: R10 nonplayer/nonmatching/non-kind-3 | No change: each guard rejects without mode/break requests; the terminal scan in V10(a) already traverses marked, nonmatching entries without further effects. Remaining guards use the same rejected-entry/owner recipe; actual traversal integration stays DUR-Q07 |
| COVERAGE8: R11 negative rate | Spec edit: V12(i) supplies rate -5 and requires requested frame 501 from frame 1000; actual scheduler normalization/delivery remains gated |
| COVERAGE9: R12a unqualified stackable and cap-1 continuation | Spec edit: V12(j) supplies unqualified, unbroken stackable armor and requires quantity-only regeneration without refresh; V12(k) supplies 39/40 then a second due tick, requiring write 40/reschedule once, followed by no further write/event |
| COVERAGE10: R13a nonweapon/failed stack chance/failed nonstack qualification | Spec edit for distinct stack-failure branch: V13(g) consumes the failed chance draw and leaves quantity 40 unchanged without selecting/spending. No change for nonweapon/failed-qualification guard variants: they use the existing pre-draw no-request recipe, with R01's predicate cases in V06 and the local empty-stack guard in V13(f); exact additional guard bindings are not claimed |
| COVERAGE11: R13b absent item/negative quantity/wrap | Spec edit: V13(h-k) directly invoke R13b, bypassing outer chance/quantity guards. They distinguish absent return 0/no requests, quantity 0 and -2 clamped writes/return 0, and wrapping minimum-i32 decrement/result |
| HARDEST1: Cyclic weighted selection | No change: R04 and V03/V04 define start modulus, ordered weights, wrap, empty skips, strict remainder comparison and exact seeds. No missing decision was identified |
| HARDEST2: Fallback write/notification/refresh/timer order | Spec edit strengthens V10 scan discrimination as above; R08a and V10(b)/V11 retain exact old-value notification and ordered requests. Selector meaning/effect integration stays gated |
| HARDEST3: Regen return/client refresh | Spec edit supplies explicit callback and concrete cap/qualification cases in V12(h/j/k); existing d/e/f preserve full-return suppression and client-conditional refresh. No rule or integration gate changes |

These vector additions/refinements require another fresh full reader run. FULL1 found no arithmetic mismatch; it does not complete review of the edited contract.

### Fresh reader FULL2, 2026-10-15

Full review covered all 13 vector families and 50 explicit cases. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 13 primary entries, COVERAGE 8, HARDEST 3. The primary gated entries group the unresolved R08b/R09b/R13c/R15 boundaries and Q01-Q09; they include Q03/Q05's outside-consumer questions, Q08's separate-build limit and Q09/O01-O04's observation limits. The smaller grouped count does not close any earlier gate. The complete reader report and pre-edit files were preserved privately. Outcomes:

| Finding | Disposition |
|---|---|
| MISMATCH 0 / GUESS 0 / GAP 0 / DISCRIM 0 | No blocking finding in the reviewed 50 cases; no rule, input interpretation or gate correction required |
| GATED13 primary entries | No change: supplied fixtures do not establish original selector meaning, break effects/interior draws, original stack selection, death effects, producer/mutation/scheduler integration or gameplay confirmation. The separate 1.14d question blocks no 1.13c assertion |
| COVERAGE1: R01 invalid kind/missing record | No change: these reject with the same false/no-draw/no-effect recipe as V06(b-e). Exact additional predicate bindings are not claimed |
| COVERAGE2: R04 selection at every table location | No change: V03/V04 exercise the one cyclic table-loop recipe, including empty skips, weights 3/5/4/2 and strict subtraction/selection. Selecting another populated location substitutes its declared pair into that same recipe; full population and every selected-location binding are not claimed |
| COVERAGE3: R05 neither/both type results | No change: the FULL1 disposition still applies. Neither is pre-draw rejection; both uses the armor threshold already demonstrated by V02(c). The original type producer remains gated |
| COVERAGE4: R08a nonpositive quantity and negative/overflowed stack cap | Spec edit for distinct cap results: V11(c/d) directly compute negative and wrapping cap results, rejecting a new lower clamp, wide sum or unsigned comparison. No change for quantity<=0: V10(b) already separates the old-quantity notification from stored 0 on the same q<=1 recipe; additional forwarded raw values add no branch |
| COVERAGE5: R11 nonintegral rate truncation | Spec edit: V12(m/n) provide rates 6/-6 and require requested frames 1417/585, distinguishing rounding up and negative floor at the actual scheduling caller. Arithmetic still consumes the stated signed truncation convention |
| COVERAGE6: R12a zero durability rate with positive quantity rate | Spec edit: V12(l) supplies qualified, unbroken stackable armor with rate 252=0/rate 253=5. It requires durability regen result 0, quantity regen result 1 and the quantity write/notification/event request. V12(j) skipped qualification and did not previously witness this return 0 fallthrough |
| COVERAGE7: R13a nonweapon/failed qualification | No change: these are the same pre-draw no-request recipe described in FULL1, with R01's qualification results exercised by V06 and a local pre-draw rejection in V13(f). No additional producer binding is inferred |
| COVERAGE8: R15 unresolved without executable vector | No change: R15 is unresolved and states no verified death-time operation. Inventing a death vector would contradict its existing outside-boundary question and limited search evidence |
| HARDEST1: R04/V03-V04 weighted cyclic order | No change: rules, ordered table, strict comparison and exact seed traces already define the operation; no missing decision found |
| HARDEST2: R08a/V10-V11 fallback write/notification order | No rule change: existing request ordering, selector-zero return and old-value notification stay exact; V11(c/d) add only the cap-boundary assertions above |
| HARDEST3: R12a/V12 return/client refresh/cadence | No rule change: d/e/f/h/k retain full suppression, client eligibility, explicit callback and cap behavior; l adds the missing zero-result fallthrough, and m/n add truncation assertions |

The five added cases V11(c/d) and V12(l/m/n) were sent for fresh FULL3. FULL2's clean blocking counts apply to the 50 cases it read, not to those later additions. Those edits changed no rule, gate or evidence label.

### Fresh reader FULL3, 2026-10-15

Full review covered all 13 vector families and 55 explicit cases. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 8, GATED 30, COVERAGE 7, HARDEST 3. The complete report and pre-fix files were preserved privately. Outcomes:

| Finding | Disposition |
|---|---|
| MISMATCH 0 / GUESS 0 / GAP 0 | No arithmetic mismatch or missing-input/blocking gap was reported in the reviewed cases |
| DISCRIM1: no both-player effect log to prove attacker-before-defender | Spec edit: V01(d) supplies both player parts with successful loss draws, logs A's write/notification before B's write/notification, and gives exact defender seed states. Reversing the parts reverses observable request order even when final values agree |
| DISCRIM2: selector-zero no-scheduling assertion had zero rates | Spec edit: V10(a) supplies rate 252=5,rate 253=25 and positive base maxstack 100 while still requiring no timer request from selector 0 |
| DISCRIM3: positive-cap fallback had zero rates | Spec edit: V11(a) supplies distinct positive rates 5/25 and requires durability-rate event request 1501 after quantity/durability writes and notifications; using quantity rate requests 1101 and fails |
| DISCRIM4: stack chance target qualified, allowing an extra qualification guard | Spec edit: V13(d) explicitly enters the standalone consumer with stat 152=1 on the chance-tested stack item. The successful chance still spends supplied item B. This does not claim the qualifying original skill caller can supply that input |
| DISCRIM5: positive skill-loss no-event assertion had zero rate | Spec edit: V13(a) supplies positive rates 252=5 and 253=25 and still requires no scheduling after durability 7 |
| DISCRIM6: no nonweapon case to prove early type return | Spec edit: V13(l) supplies a nonweapon with otherwise qualifying 10/10,chance 50,amount 3 and seed S57. It requires no draw/request and unchanged durability/seed, despite a would-be successful roll 3 |
| DISCRIM7: no unqualified nonstack weapon case | Spec edit: V13(m) supplies weapon/nonstack with stat 152=1 and otherwise effective chance/amount. It requires the pre-draw no-effect return |
| DISCRIM8: equal-maximum selected durability branch | No change: V13(i/j/k) already supply durability 20/max 20 and explicitly forbid a durability write while requiring quantity writes/refresh. An unconditional max-durability write would fail their existing request logs |
| GATED30 | No change: these are the documented original producer/selector, break/mutation/refresh/scheduler and observation limits; fixtures do not close them or promote labels |
| COVERAGE1: R01 invalid kind/missing record | No change: the same false/no-draw/no-effect predicate recipe is documented and exercised by V06(b-e); exhaustive bindings are not claimed |
| COVERAGE2: R04 right-arm selection | No change: right-arm population uses the declared weight 4 in the same cyclic table-loop recipe as V03/V04; these cases already visit the right-arm slot and subtract the weight 4 at left arm. No new location-specific operation or producer is asserted |
| COVERAGE3: R05 neither/both type results | No change: prior FULL1/FULL2 dispositions remain; neither rejects before draw, both uses armor threshold 10 as in V02(c). Downstream weapon-flag routing remains explicit |
| COVERAGE4: R08a quantity<=0 | No change: V10(b) already distinguishes old quantity from stored 0 for the q<=1 branch; substituting another raw q in that same forwarded-value recipe adds no new operation |
| COVERAGE5: R10 nonplayer/nonmatching | No change: zero requests on rejected owners/entries and terminal marked-item traversal remain the existing recipe; actual inventory integration is still DUR-Q07 |
| COVERAGE6: R12a quantity-target cap | Spec edit: V12(o) supplies qualified armor with rate 252=0,quantity 9,maxStack10,maxDur40 and two due ticks. It requires quantity 10 then no further write/event, rejecting an incorrect maxDur argument even though the generic regen cap branch was already covered for durability |
| COVERAGE7: R13b equal maximum | No change: identical to DISCRIM8's already-discriminating V13(i/j/k) assertions |
| HARDEST1: R04 weighted seed order | Spec edit only strengthens caller ordering through V01(d); R04/V03/V04's cyclic arithmetic and seed contract are unchanged |
| HARDEST2: R12a return/fallthrough/cadence | Spec edit adds quantity cap argument discrimination in V12(o); all existing return 0/return 1,client,explicit-callback and requested-frame assertions remain unchanged |
| HARDEST3: R13b clamp/wrapping return | No change: direct V13(h-k) already distinguish absent return, negative clamp, wrapping decrement/return and equal-max suppression without the outer chance guard |

All three full runs returned MISMATCH 0 and every finding has a recorded disposition. The next fresh review is DELTA4 over changed/new cases V01(d),V10(a),V11(a),V12(o),V13(a/d/l/m), checking the findings above. There are four added cases,59 total on the prior counting basis. No rule, ownership, gate, acceptance boundary or evidence label changed; original-profile actual effects remain gated. No pure-query count was added: ordering assertions use write,notification and timer-request logs.

Migration self-review on 2026-10-15 checked scope, semantic fields, one-label sub-ids, request-versus-stored boundaries, exact seeds/thresholds/weights, first/later timer frames, stat/quantity notification order and client-conditional refresh. The preservation inventory retains all 15 parent rule families, 13 vector families, nine questions and four observations; substantive corrections are listed above rather than silently preserving unsupported claims. Preliminary review added the declared qualification, weighted selection, regen return, default-handler selection and stack-spend return to the output inventory without changing their rules or cases. It also classified the separate 1.14d question as outside-boundary, explicitly blocking no 1.13c acceptance or runtime-confirmation assertion. Mechanical lint after the FULL3 fixture fixes returned 0 errors and 0 warnings. A fresh blind-reader pass is managed separately against the final three-file hashes; this record makes no claim that an earlier unrelated reader approved this contract. No implementation comparison, plan, gameplay campaign, commit or publication is part of this migration.

### Fresh reader DELTA4, 2026-10-15

After three full passes with MISMATCH 0, a fresh fourth reader checked all eight changed cases: V01(d), V10(a), V11(a), V12(o), and V13(a/d/l/m). Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0. All named findings were resolved. The exact attacker-before-defender effect order, selector-zero suppression despite positive rates, fallback rate 252 versus 253 and event ordering, quantity cap across two ticks, skill no-scheduling and stack/no-qualification branches, and pre-draw guards were independently recomputed. Existing equal-maximum cases V13(i/j/k) remained unchanged.

No rule, ownership, gate or acceptance boundary changed after FULL3. The final applicable reader condition is met; documented gates and observation limits remain, and status stays draft. A punctuation-only repair made the two arrows in V12(o) explicit after the reader without changing inputs, request order or outputs. No gameplay observation, implementation test, implementation plan or user approval is implied. The complete delta report is retained privately.
