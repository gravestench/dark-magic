# Evidence: skill leveled damage, to-hit and elements

Contract: [Skill Leveled Damage, To-Hit and Elements](../../specs/skill-leveled-damage-tohit-and-elements.md).
Open evidence: [gap file](../../plans/gaps/skill-leveled-damage-tohit-and-elements.md).
Original artifacts: [source manifest](../source-manifest.md), `LOD113C-D2COMMON` and `LOD113C-D2GAME`.

## Identity, reuse and method

The target is original Lord of Destruction Expansion 1.13c, Windows x86. Function descriptions below are analyst descriptions, not established original exported names. Module-relative RVAs and ordinal bindings apply to the manifest fingerprints. The original static record was read on 2026-10-12 using original instruction listings, byte-matched reconstructions where listed, export/import resolution, and read-only image/listing cross-checks. Its dispatch, field, arithmetic, routing and RNG findings are reused with that date and coverage. The current original-image fingerprints correspond to the analysis specimens as recorded in the shared manifest; this identity check adds no gameplay observation.

Targeted readings on 2026-10-15 checked the physical maximum entry, physical source consumers and item wrappers, Kick caller bypass, elemental callers/fills, routing, and to-hit hand-off. The available matching-source index was checked first; those newly read routines had no listed byte-checked reconstruction, so instruction listings were used. The physical maximum entry and routing's encoded indirect jump were additionally decoded from the hash-checked original images. The mastery decoding and common stat-read contracts reuse the compatible SCX/NRF records, including their actual dates and boundaries. New source evidence supports consumer stages; it does not establish complete equipment, Kick or hit-resolution production.

No gameplay observation was made. A coordinated read of narrow text cells through the original game's loader on 2026-10-15 is recorded below, with profile and compiled-data limits. The prior complete specification and private investigation details were preserved separately before migration; public references do not require that archive.

## Dispatch and functions

The server skill do-function table is D2Game RVA `0x1074A8` (preferred VA `0x6FD274A8`). Its `Skills.srvdofunc` entries anchor these paths:

- 38 -> D2Game `0x93340` -> physical fill `0x9F1A0` -> elemental fill `0x9F210` -> area dispatch `0xA0C70` with callback `0x92CC0`.
- 53 -> D2Game `0x92CF0` -> physical fill `0x9F1A0` -> elemental fill `0x9F210` -> area dispatch `0xA0C70` with callback `0x92CC0`.
- 46 -> D2Game `0x93270` -> `0x92A60` -> to-hit import `0xACD6` and elemental fill `0x9F210`. The to-hit branch calls melee result `0xDE5A0`, which conditionally calls hit roll `0xDDE90` with the unchanged bonus.

These table entries and the computing output writers, rather than function labels, establish identity. The 2026-10-12 record also found physical-fill callers at table entries 21/33/42/51 (`0x431C0`, `0x92770`, `0x95FE0`, `0x957F0`) through cross-references; their bodies were not read for this contract and no equivalent caller behavior is claimed.

| Artifact | RVA / export | Analyst description | Rules and supported boundary |
|---|---|---|---|
| LOD113C-D2COMMON | `0x4DDB0`, internal | Five-band progression | R02/R03/R08; byte-checked reconstruction read 2026-10-12; signed branches, wrapped products and return |
| LOD113C-D2COMMON | `0x52100`, #10567 | Physical minimum | R01a/R03 through R07; instructions 2026-10-12; source/Kick consumer reread 2026-10-15 |
| LOD113C-D2COMMON | `0x51FF0`, #10297 | Physical maximum | R01b/R03 through R07; instructions 2026-10-12; complete original-image entry/return decoding 2026-10-15 |
| LOD113C-D2COMMON | `0x31B90`, #10641; `0x1FB80`, #10051 | Weapon and source-classification producers | Q02; supplied-result consumer anchors; targeted instructions 2026-10-15 identify further inventory/item dependencies, not a complete production contract |
| LOD113C-D2COMMON | `0x38ED0`, `0x38F60`, internal | Item source endpoint wrappers | R04b; instructions 2026-10-15, guards and selected-array stat 23/24 reads |
| LOD113C-D2COMMON | `0x38B70`, #10973; `0x38A80`, internal | Ordinary/selected-list stat reads | R04b/R10/R13; consumed RLI-R02a/RLI-R02b, whose original evidence owns leaf search/minimum behavior |
| LOD113C-D2COMMON | `0x4F930`, internal | Kick producer | R07/Q08; caller references from 2026-10-12, targeted listing 2026-10-15; producer arithmetic remains outside boundary |
| LOD113C-D2COMMON | `0x50460`, #10121; `0x50360`, #11091 | Elemental minimum/maximum | R01a/R08/R09/R10; instructions 2026-10-12; caller/order checks 2026-10-15 |
| LOD113C-D2COMMON | `0x4F870`, internal | Elemental mastery amount | R10; original instruction/map corroboration 2026-10-15 (see [absorbed expression-spec records](#records-absorbed-from-the-expression-spec-2026-10-15)) and shared NRF getter semantics |
| LOD113C-D2COMMON | `0x4E900`, #10510 | Elemental duration | R01a/R11; instructions 2026-10-12; three signed bands, no shift, calculation result addition |
| LOD113C-D2COMMON | `0x4EA50`, #10653 | Skill to-hit value | R01a/R12; instructions 2026-10-12 and 2026-10-15; linear fallback or unchanged evaluator result |
| LOD113C-D2COMMON | `0x51BF0`, #10786 | Skill calculation front end | R05/R19; byte-checked reconstruction 2026-10-12; SCX-R11 owns shared pool/context contract |
| LOD113C-D2COMMON | `0x9270`, internal; `0x50B60`, internal | Evaluator and parameter callback | R09/R11/R12/R19; boundary crossing consumes the named SCX contracts/gates |
| LOD113C-D2COMMON | `0x11E0`, internal | Argument-ordered percentage primitive | R05/R09/R10/R11; byte-checked reconstruction 2026-10-12; consumes NRF-R05 |
| LOD113C-D2COMMON | `0x61EC0`, internal | Skills descriptor loader | Input field identity; 2026-10-12 original descriptor/string reading, consumed GDA-R11 |
| LOD113C-D2GAME | `0x9F1A0`, internal | Physical fill | R14; instructions 2026-10-12 and 2026-10-15; min/max requests, source condition zero, range draw, additive physical write |
| LOD113C-D2GAME | `0x9F210`, internal | Elemental fill | R15; instructions 2026-10-12 and 2026-10-15; duration/min/max requests, range draw, router, unchanged d return |
| LOD113C-D2GAME | `0x9EFE0`, internal | Router/random element | R16/R17; instructions/table read 2026-10-12; encoded jump corroborated against original image 2026-10-15 |
| LOD113C-D2GAME | `0x11D0`, internal | Alternate random modulo | R14/R15; 2026-10-12 reading; consumes NRF-R01/NRF-R02 |
| LOD113C-D2GAME | `0x93340`, `0x92CF0`, internal | Two caller record-construction stages | R18; instructions 2026-10-12, stage/order reread 2026-10-15; no proof of other caller eligibility/effects |
| LOD113C-D2GAME | `0x92A60` -> `0xDE5A0` -> `0xDDE90`, internal | To-hit sum and conditional forwarding | R13; consumer instructions 2026-10-15; actual hit outcome consumes SBA |

### Import bindings

The original D2Game import thunks resolve through the original import/export tables:

| D2Game thunk RVA | Resolved original target | Use |
|---|---|---|
| `0xA796` | D2Common #10567 / `0x52100` | physical minimum |
| `0xA766` | D2Common #10297 / `0x51FF0` | physical maximum |
| `0xACE2` | D2Common #10510 / `0x4E900` | duration |
| `0xAE92` | D2Common #10121 / `0x50460` | elemental minimum |
| `0xAEC2` | D2Common #11091 / `0x50360` | elemental maximum |
| `0xACD6` | D2Common #10653 / `0x4EA50` | to-hit |
| `0xAC9A` | D2Common #10786 / `0x51BF0` | calculation |
| `0xA42A` | D2Common #10973 / `0x38B70` | stat read |

Bindings reuse 2026-10-12 import/export resolution; #10973 was independently re-resolved on 2026-10-15. D2Common's source calls #10641/#10051 are same-module direct calls to `0x31B90`/`0x1FB80`; the ordinals identify their exported entries rather than implying an import on that path.

## Semantic mapping and rationale

The public contract does not require the original storage layout. These mappings allow independent comparison with the original image:

| Semantic item | Original binary representation |
|---|---|
| Skill record lookup | DataTables records +0xB98/count +0xBA0; record stride 0x23C |
| Physical fields | `MinDam` +0x1A8, `MaxDam` +0x1AC; min increments +0x1B0 through +0x1C0, max +0x1C4 through +0x1D4; synergy +0x1D8 |
| Flags/source/shift | Kick at record +0x04 bit 9, tested through the shared flag-mask table; HitShift +0x1A4, SrcDam +0x1A5 |
| Elemental fields | EType +0x1DC; EMin/EMax +0x1E0/+0x1E4; min increments +0x1E8 through +0x1F8, max +0x1FC through +0x20C; synergy +0x210 |
| Duration fields | ELen +0x214; increments +0x218/+0x21C/+0x220; synergy +0x224 |
| To-hit fields | ToHit +0x198, LevToHit +0x19C, ToHitCalc +0x1A0 |
| Calculation pool | DataTables +0x40 pointer, +0x44 unsigned size; evaluator receives remaining length and the unit/skill/level context |
| Unit source/seed | Inventory association +0x60; unit list +0x5C; seed low/high +0x20/+0x24 |
| Physical/fire/lightning/magic/cold/poison | DamageRecord +0x08/+0x10/+0x1C/+0x20/+0x24/+0x28 |
| Burn magnitude/length | +0x14/+0x18 |
| Poison/cold/freeze durations | +0x2C/+0x30/+0x34 |
| Life leech/mana burn/stamina burn/stun duration | +0x38/+0x3C/+0x40/+0x44 |
| Flags/result/hit class | +0x00/+0x04/+0x60; result flags are a word |

Physical/duration/elemental endpoints receive stack arguments `(unit,skill,L,condition)` and clean 16 bytes; to-hit receives `(unit,skill,L)` and cleans 12. The band helper uses level in EAX and increments in ECX. Physical fill receives skill in EAX, unit in EBX, and stack `(record,L)`; elemental fill uses unit in EDI with the same other placement. Calling-convention names are unnecessary: these placements follow callers and callee reads, not cleanup alone.

For physical damage the source selection precedes the band/base addition, percentage addition, and shift. Item wrappers `0x38ED0`/`0x38F60` test unit/list/definition availability and select key `0x170000`/`0x180000` on the nonzero-condition branch. The physical callers pass 1 and 0x18 respectively, both nonzero. These wrappers directly call `0x38A80`, the selected-array/minimum leaf; the item's own-array storage is not forced. Ordinary source reads use #10973. The weapon selector `0x31B90` calls inventory/skill/item helpers `0x20700`, `0x20830`, `0x23EF0`, `0x24430`, `0x30530`, `0x4DE40`; source classification `0x1FB80` reaches `0x1E1B0`, `0x22310`, `0x22E70`, `0x24BF0`. Those production paths are outside this consumer boundary and were not all read. Classification is therefore a neutral numeric input, not a proven two-handed-wield condition.

Physical maximum's hash-checked original instructions at `0x51FF0` through `0x5203E` establish signed lookup rejection, return 2 at `0x52016`, and maximum Kick selection zero at `0x52036`. Physical minimum's established instructions return 1 and use selection 1. Both return directly from the Kick call before ordinary source/synergy/shift. The targeted `0x4F930` listing identifies stat-dependent player/non-player branches; their complete arithmetic is not made part of this contract. R07 verifies only the caller's dispatch/unchanged return using supplied results.

Elemental `0x50460`/`0x50360` shift first, evaluate synergy when allowed, and only then call `0x4F870` and add its return. The minimum tests signed shifted damage against 0x100 and first minimum increment against zero before evaluation; maximum omits that test. SCX's [mastery and shared-getter reconciliation](skill-calc-expressions.md#shared-getter-reconciliation-2026-10-15) corroborates the selector table at `0x4F91C`, target table at `0x4F908`, ordinary reads of 329/330/332 and cold's direct `0x38A80` stat-331 read. That record and NRF's same-fingerprint getter evidence settle the leaf interpretation; no unresolved blanket mastery gate remains.

Routing uses an unsigned comparison of `EType-1` with 11 and an indirect jump through D2Game `0x9F0E4`. The 2026-10-12 image read established the table and writes; original-image decoding on 2026-10-15 corroborated the encoded dispatch instruction and targets. Random type 10 advances the supplied unit seed with the foundation LCG and uses the original image table at D2Game RVA `0xFB970` (preferred VA `0x6FD1B970`), choices `{1,2,4,5}`. The fill preserves d across routing and returns it, including when type 9 uses d+length in the record.

The caller stage `0x93340` zeroes 0x70 bytes, fills physical then elemental, sets flags/result 1/1, and passes the area callback. `0x92CF0` sets flags 2 after clearing and before the same fills. R18's downstream no-physical/no-critical assertions consume DPO's contract; they are not an independent reading of that damage build. Calculation operations run inside endpoint requests before each fill's local range draw, so their SCX-owned draws can interleave with the local sequence. V15 makes this boundary observable without inventing a parameter producer.

At `0x92A60`, the to-hit return is added to stat 0x145 (325) before the call to `0xDE5A0`. That resolver reads the same bonus and pushes it unchanged at the call to `0xDDE90`; target/attacker checks, eligibility and hit/avoidance effects are outside the hand-off proof. No eligibility guarantee or outcome is inferred from this argument path.

## Corrections and affected-claim checks

The migration retains all thirteen original vector numerical outcomes and original rule ids. Repository searches for the filename and SLD prefix found no implementation-code/test or non-gap implementation-plan consumer; sibling citations alone do not require aliases. Existing ids were already ordered and remain usable. Binary fixture encodings remain in the separate preserved record and map to semantic fields above.

- The 2026-10-12 R01b unresolved label concerned the encoded maximum prologue. The 2026-10-15 same-fingerprint original-image decode resolves that uncertainty; return 2 and maximum Kick selection 0 are unchanged.
- The earlier R10 cold wording said “base stat list.” SCX's original wrapper/map decoding and NRF's leaf contract show a selected-array read, including minimum adjustment. This corrects a potential own-array interpretation. V14 separates own value 10 from selected value 30, producing mastery amount 300 rather than 100. Existing numerical vectors did not exercise cold mastery and keep their outcomes. R10 becomes statically verified through this exact dependency boundary.
- Earlier R04b wording identified the item's “own stat,” without proving storage-array behavior; the newly read item wrappers call the same selected-array leaf. R04b's supplied-result selection is now verified, while Q02 retains the unverified equipment producers. No equipment mechanics are inferred from labels.
- Earlier R07 was asm-only because the Kick producer was unread. Its caller bypass/argument/return behavior is now separately verified with supplied Kick results, while Q08 explicitly retains actual producer integration outside scope. This narrower verification does not close the broader physical-source/Kick production gaps, now Q02/Q08 (which absorbed the expression spec's equivalent question in the 2026-10-15 restructure).
- Earlier R13 had an unread forwarding path inside melee resolution. Targeted 2026-10-15 reading establishes the unchanged argument route, so the hand-off becomes verified. SBA continues to own whether that path executes and its result.
- Q04 now consumes the written SCX contract, including calculation RNG and precise unresolved producer gates. A local range draw is not the complete endpoint-dependent draw count. V01 through V13 use the same no-calculation-draw inputs as before; V15 additionally distinguishes interleaving.
- The 2026-10-15 text reads supply the previously absent stat names and named element-code rows. They do not establish compiled selector-index correspondence, so Q07 remains narrow and open. No static rule is promoted by a text read.

### Spec-review correction, 2026-10-15

An independent review of D2Game `0x92CF0` (the do-function 53 record stage) found an omitted write between the elemental fill and the area hand-off. After `0x9F210` returns, the routine loads the dword at selected Skills record +0x134 (record base = Skills collection + skill*0x23C, register preserved across both fills) and, when it is nonzero, stores it to combat-record +0x60 (hit class); a zero value skips the store. Do-function 38's stage `0x93340` goes straight from the elemental fill to the flag writes and hand-off, with no such store. The previous R18 text described do-function 53 as clear, flag 2, physical fill, elemental fill, which omitted this replacement; V13's do-function 53 case did not state hit class. R18 now states the replacement under a neutral field name, V13 gives value 0 (hit class 32) and value 7 (hit class 7) cases, and SLD-Q10 records that the Skills column feeding +0x134 is not identified in GDF's descriptor records. No existing numeric outcome changed. This changes a rule's behavior and a vector, so a fresh delta reader check of R18/V13 is required before the reader review is again complete; it has not run.

The same review independently re-derived V01 through V20 from the rules, re-read physical maximum `0x51FF0`, elemental minimum `0x50460`, mastery helper `0x4F870` (selector bytes 0,1,4,2,3,4,4,4,4,4,4,2 for EType 1..12; targets push 329/330/332, the cold path or zero), router `0x9EFE0`, fills `0x9F1A0`/`0x9F210` and stage `0x93340`; all agree with R01b, R04a/b, R07, R08 through R10, R14 through R17 and the do-function 38 part of R18.

## Named table cells

The following coordinated collection on 2026-10-15 used a running member reporting vanilla launch key `lod-113c`; separate PD2 profiles were excluded. Current member inventory was available, while the fresh local versions-console check refused connection. The selection reused the historical vanilla profile with no complete fresh override/archive inventory; a prior patch inventory reported `mpq_override` disabled. This supports the named effective text cells, with the explicit limit that a newly verified pristine archive set and compiled-table correspondence were not established. The code fingerprints do not remove that limit. Storm's effective-file read does not identify the physical supplying MPQ.

Both requests used offset 0 and length 0 (whole effective file). The complete returned `base64` bodies were decoded, lengths checked, hashes calculated, and headers parsed by name. No gameplay action, memory write or save change was involved.

| Record | Virtual file / coverage | Named cell and raw value | Interpretation and limit |
|---|---|---|---|
| TD-STAT325 | `data/global/excel/itemstatcost.txt`; returned/total 40109/40109 bytes, range [0,40109) | explicit `ID` 325, `Stat` = `progressive_tohit` | Also zero-based data row 325 excluding header; effective text identity for code-fixed 325, not fresh compiled correspondence |
| TD-STAT332 | same complete file and headers | explicit `ID` 332, `Stat` = `passive_pois_mastery` | Also zero-based data row 332 excluding header; text name only |
| TD-ETYPE1 | `data/global/excel/elemtypes.txt`; returned/total 195/195 bytes, range [0,195) | zero-based data row 1, `Elemental Type` = `Fire`, `Code` = `fire` | Positional text identity; no explicit numeric ID column |
| TD-ETYPE2 | same complete file/headers | row 2, `Elemental Type` = `Lightning`, `Code` = `ltng` | Same correspondence limit |
| TD-ETYPE3 | same complete file/headers | row 3, `Elemental Type` = `Magic`, `Code` = `mag` | Same correspondence limit |
| TD-ETYPE4 | same complete file/headers | row 4, `Elemental Type` = `Cold`, `Code` = `cold` | Same correspondence limit |
| TD-ETYPE5 | same complete file/headers | row 5, `Elemental Type` = `Poison`, `Code` = `pois` | Same correspondence limit |
| TD-ETYPE12 | same complete file/headers | row 12, `Elemental Type` = `Freeze`, `Code` = `frze` | Same correspondence limit |

These cell claims are **verified in 1.13c data** within the reported effective profile and stated correspondence limits. ItemStatCost full effective-file SHA-256 is `ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`; ElemTypes full effective-file SHA-256 is `91addf64ff87abb19dcceba77f6da6ed84e9e4194d30eeb1d69d36edd55cdc15`. No table or full row is reproduced. Other stat names and Skills descriptors consume GDF's existing data/loader evidence with its own dates and limits.

## Records absorbed from the expression spec (2026-10-15)

The 2026-10-15 restructure removed the damage-builder rules from the expression spec because they
duplicated this contract (see [Restructure record](#restructure-2026-10-15)). Its dated
original-source records for those builders moved here unchanged apart from ids. They are
independent readings of the same functions listed under [Dispatch and functions](#dispatch-and-functions);
where both records exist, each keeps its own date and boundary.

| Artifact | Function RVA / export | Analyst description and supported rules | Method, date, boundary |
|---|---|---|---|
| LOD113C-D2COMMON | 0x4DDB0 | five-band level helper, SLD-R02 | Original disassembly, 2026-10-12; each signed level branch and summed result |
| LOD113C-D2COMMON | 0x50460 export #10121; 0x50360 | elemental min/max, SLD-R08 through SLD-R10 | Original disassembly, 2026-10-12; shift before synergy, minimum-only gate, optional mastery after synergy |
| LOD113C-D2COMMON | 0x4F870; selector map 0x4F91C; targets 0x4F908 | mastery selection, SLD-R10 | Original disassembly, 2026-10-12; instruction/map corroboration against matching image 2026-10-15; cold wrapper reconciled with the shared stat-read contract |
| LOD113C-D2COMMON | 0x11E0 | argument-ordered percent primitive, SLD-R05/SLD-R09/SLD-R10/SLD-R11 | 2026-10-12 reading agrees with consumed NRF-R05; current contract uses that shared rule |
| LOD113C-D2COMMON | 0x52100; 0x51FF0 export #10297 | physical min/max, SLD-R01a/SLD-R01b, SLD-R03 through SLD-R07 | Instructions, 2026-10-12; source-share/synergy/shift sequence; source and Kick leaf integration remain gated |
| LOD113C-D2COMMON | 0x4F930; 0x31B90; 0x1FB80; 0x38ED0; 0x38F60 | Kick and weapon/item source dependencies | Prior caller references, 2026-10-12; no complete result proof recorded, SLD-Q02/SLD-Q08 |
| LOD113C-D2COMMON | 0x4E900 | elemental duration, SLD-R11 | Instructions, 2026-10-12; three level bands, base, synergy and return |

| Semantic term | Original evidence representation |
|---|---|
| fixed mastery selection | table at RVA 0x4F91C indexed by EType-1; jump targets at 0x4F908; unsupported entries select zero return |
| cold source | unit list +0x5C; definition count +0xBD4 >331; definition index 331; identity encoded as 0x14B0000; direct call 0x38A80 |
| ordinary mastery source | 329/330/332 at layer 0 through 0x38B70; percentage tail jump0x11E0 |
| physical source selection condition | 0x1FB80(unit,unit inventory association) returns 2; meaning/complete dependency remains open |

Targeted finding 4 of the expression companion (2026-10-15), with ids in current numbering:

4. The mastery mapping in the earlier SCX reading agrees with the mapping in
   SLD-R10. At the time of that earlier comparison, SLD's decoding/integration
   limits remained unresolved; its current R10 reuses the compatible SCX/NRF evidence.
   New original-image inspection of 0x4F870 corroborates instruction and
   table bytes, including the cold-specific direct 0x38A80 call and ordinary
   0x38B70 calls for other selected elements. This settles the decoding ambiguity. The compatible NRF evidence identifies the exact leaf at 0x38A80 as the selected-array/minimum operation and 0x38B00 as the own-array/minimum operation. The cold wrapper therefore consumes RLI-R02a/RLI-R02b directly. It does not force the base/own array. No gameplay result was collected.

The [shared getter reconciliation](skill-calc-expressions.md#shared-getter-reconciliation-2026-10-15)
record stays in the expression companion because it also covers that spec's stat-function qualifiers.

### Elemental synergy observation procedure

For SLD-O02 (pending), the expression spec's procedure adds: use a matching vanilla Expansion 1.13c Sorceress with a learned
elemental skill and a related damage-synergy skill, in Normal difficulty, standing
in town with stable equipment and no temporary buffs. Record character/skill
levels, learned synergy levels, relevant mastery stats, selected Skills fields,
and the caller's mastery flag before changing a synergy point. Measure raw min/max
returns at D2Common 0x50460/0x50360 for the same skill context before and after a
controlled synergy increase. Calculate each expectation from the recorded fields,
using SLD-R09/SLD-R10. The tooltip alone cannot establish raw rounding. Exact
character setup and expected numbers must be pinned before this becomes a
reproducible observation vector; no fabricated record is supplied here.

## Remaining source boundaries and searches

The 2026-10-12 search used dispatch-root anchors at the skill table and fill cross-references; output anchors at router record writes; data anchors for Skills columns and record stride; column-name strings and loader cross-references; and decompile/listing aids for encoded instructions. No runtime search was completed. No claim is labeled “not found.”

The 2026-10-15 follow-up reused those identified roots instead of repeating discovery. It read the previously missing maximum entry, source/item consumer path, to-hit argument forwarding, and original router dispatch bytes. The remaining questions are not absent-function claims: equipment/Kick producers are located, with further callees outside this spec; other source-enabled callers were not expanded into the boundary. The prior caller list identifies D2Game `0x39E30`, `0x92360`, `0x98E90`, `0x9E330`, `0xB3B20`, `0xDCE00` as leads, not an enumerated source-inclusion guarantee. SCX's dependency gaps and the lack of compiled element-selector correspondence remain explicit. Gameplay confirmation was not requested as part of migration and remains pending.

## Observation procedures

All procedures require owned original 1.13c files matching the source profile and controlled synthetic/test characters. Record class, level, difficulty, exact Skills fields, unit stats, enabled conditions, requested arguments, seed before/after, ordered endpoint calls and final record. Relocate module-relative locations for the live process; account for dropped/truncated records and draw-producing calculations. A matching subset confirms only its named assertions.

### Level observation

SLD-O01: single player Normal; use any class with a physical-damage skill whose measured Skills record has nonzero physical level increments. Stand next to one eligible monster, cast once at each controlled level 8,9,16,17,22,23,28,29, and read D2Common #10567/#10297 returns plus D2Game `0x9F1A0` inputs/output. Derive exact expected endpoints from those recorded field values and R02/R06 before comparing the range draw and additive write. No specific skill-row values or matching trace are currently recorded.

### Synergy observation

SLD-O02: single player Normal, a controlled character/eligible nearby target, and a measured skill record using both physical and elemental synergy calculations. Record the exact computed percentages and source inputs; choose a percentage/base combination where shifting before versus after percentage division differs. Compare raw endpoint results at D2Common `0x52100`/`0x51FF0` and `0x50460`/`0x50360`, then compare with R05/R09. Record every expression draw before attributing later seed steps. A suitable exact row/fixture has not yet been established. An elemental-only variant is in the [elemental synergy observation procedure](#elemental-synergy-observation-procedure).

### Random element observation

SLD-O03: single player Normal, a character using a measured `EType` 10 skill and an eligible nearby target. Read the caster seed, duration/min/max requests at `0x9F210`, then the router at `0x9EFE0` and written element/duration fields. Match the range step when positive and the following element step against R17, including the 50-frame fallback only for nonpositive supplied length. No exact original-game skill/fixture or trace has been collected.

### Mastery observation

SLD-O04: single player Normal, a Sorceress fire skill, a nearby eligible monster, and recorded fire mastery stat 329. Compare elemental helper returns with mastery disabled/enabled and otherwise identical measured inputs, or separate controlled characters with exact field/stat differences recorded. At D2Common `0x4F870`, compare the percentage amount with the synergy-increased damage passed by `0x50460`/`0x50360`. Also record cold's selected versus own values before attributing a future cold case. No matched observation is recorded; text reads do not qualify.

## Migration and review record

The 2026-10-15 migration preserves the existing behavior boundary and original numerical vector results, replaces binary-layout requirements with semantic fields, moves detailed public evidence here, and keeps private encodings/investigation history separately. Twenty executable vector groups cover every verified consumer rule; supplied results and dependency execution are explicitly distinguished. Header status remains draft. Self-review checked scope, original numerical outcomes/seeds, complete consumer source boundaries, getter semantics, precise gates, rule/vector coverage, dependency references, and public/private separation. The mechanical command `python .agents/skills/spec-writing/tools/check_spec.py skill-leveled-damage-tohit-and-elements` completed on 2026-10-15 with exit code 0 and `0 error(s), 0 warning(s)`. This is documentation validation, not execution of gameplay vectors or implementation tests. A separate fresh-reader pass is coordinated by the parent task.

### First full fresh-reader pass, 2026-10-15

The reader recomputed all 20 executable vector groups and reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0. GATED 7 identified exactly Q02, Q04, Q05, Q06, Q07, Q08 and Q09, consistent with the contract's supplied-input and integration boundaries. The three hardest checked cases were V13's draw order/record hand-off, V14's mastery plus synergy and cold selected-array source, and V19's wrapping bonus/forwarding.

The COVERAGE heading reported six advisories, but the supplied list contained five items: R05 nonempty synergy returning zero; R09 absent-pool/out-of-range skip and valid calculation returning zero; R11 duration calculation returning zero; R12 to-hit calculation returning zero; and R17 random indices 1/2 selecting types 2/4. These are optional additional branch combinations, not reported mismatches, guesses, evidence gaps or discrimination failures. R05/R09/R11 state the unchanged-value branch for zero; R09/R19 and SCX-R11 give the exact rejected-reference recipe; R12/SCX-R25b return the evaluator's signed result unchanged, including zero; R17 states the full ordered selection array and consumes NRF-R01 for the step. Existing executable groups and cited operation contracts establish those result recipes. No exhaustive branch-vector or gameplay-observation claim is made, and these advisories caused no contract changes.

### Separate operation-order finding and response, 2026-10-15

After that reader report, the coordinating review independently found that V14's original base 2560, synergy 50 and mastery 30 yield 4992 in either percentage order. The reader's DISCRIM 0 report is preserved above; it did not report this finding. The original trace specified the correct post-synergy input, but its final result alone did not distinguish the order.

An independent minimum-only rounding case was added under V14 without changing any existing outcome: EMin 257, HitShift 0, level 1, zero increments, supplied synergy 50, mastery enabled, EType 1 and supplied ordinary stat 329 value 30. The unit, skill lookup, pool/reference and evaluator-stub inputs are explicit. NRF-R05's signed thresholds put both calls on the narrow-product path: `50*257/100 = 128`, then `30*385/100 = 115`, producing 500. Reversing the operations gives `257+77=334`, then `334+167=501`, so the new final-result assertion distinguishes the order. Self-review checked the minimum gate (257 > 256), threshold/operand order, exact truncation, no-draw boundary and preservation of the earlier 4992/cold-source outcomes. A second full fresh-reader pass is required after this edit; this record does not claim that it has run.

### Second full fresh-reader pass, 2026-10-15

The fresh reader recomputed all 20 vector families and every case, including V14's new
rounding case. MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0; GATED 23, COVERAGE 4, HARDEST 3.
This was a full review. The seven existing questions and sixteen supplied-stage vector
boundaries made up the reader's GATED count; no additional gate or evidence promotion follows.

| Advisory branch | Outcome |
|---|---|
| R05 calculation returns zero | No change: V03 already checks the reached calculation request and nonzero consumer path; R05's zero amount preserves the existing value. This adds no different result recipe or clamp |
| R09 absent/invalid calculation pool/reference | No change: R19/V20 and SCX-R11 own and distinguish rejection without evaluation; V05/V06 separately exercise caller gating. A missing pool does not supply an invented expression result |
| R11 duration calculation returns zero | No change: V07 checks the unchanged base and the requested percentage addition; zero is the stated identity amount on that same signed recipe |
| R17 selection indices 1 and 2 | No change: R17 fixes the ordered selection array; V12 checks selection timing/default duration and two indices, while V16 checks the lightning/cold routing rows. This is not an exhaustive assertion for every selection/route combination |

The reader initially listed do-function-38 caller eligibility as a fifth coverage entry,
then retracted it because that eligibility is explicitly outside the vector boundary.
The final four-entry report is recorded above. All HARDEST entries need no change:
V12 has explicit adjacent seed states and optional range order; V13 separates the two
fill requests/draws and final flags/hand-off; V14 now distinguishes 500 from the reversed
order's 501 and separately distinguishes cold's selected array from its own array.

The applicable reader review is complete at the declared supplied-input/static boundary.
Existing producer, integration and observation gaps remain; header status stays draft.
No human approval, complete gameplay parity or runtime confirmation is inferred.

### Third full fresh-reader pass, 2026-10-15

This full review covered the contract after the R18/V13/Q10 correction. Counts: MISMATCH 0, GUESS 7, GAP 12, DISCRIM 13, GATED 13, COVERAGE 14, HARDEST 3. Every vector case was recomputed without a difference.

Triage was checked against the original D2Game/D2Common instructions on 2026-10-15. The supporting readings were as follows:

- D2Game `0x92CF0` rejects a failed lookup (negative identity, identity at least count, null record) with result 0 before any call. It clears 0x1C dwords (the whole 0x70-byte record), stores flags 2, then calls `0x9F1A0` and `0x9F210`. It then loads the dword at selected record +0x134 and, on a nonzero test, stores that dword to record +0x60. It hands off through `0xA0C70` with callback `0x92CC0` and returns 1.
- D2Game `0x93340` has the same lookup rejection with result 0. After the fills it ORs flags with 1 and resultFlags (word) with 1 on the cleared record, has no +0x134 access, and returns 1 after the hand-off.
- The fills `0x9F1A0` and `0x9F210` contain no level test. `0x9F210` pushes condition 1 for all three requests.
- Duration `0x4E900` reads only unit, skill and level; its fourth argument is unused.
- Physical minimum `0x52100` compares the classification with equality to 2.
- The selected-array leaf `0x38A80` selects its array from the list's resolved-stats flag. The ordinary unit read `0x38B70` and the item wrapper `0x38ED0` (count > 23 guard, stat 23 on a nonzero condition) both use that leaf.

| Entry | Outcome |
|---|---|
| GUESS 1 RandAlt undefined | Spec edit: R14/R15 name NRF-R02 `Rand` exactly |
| GUESS 2 fills at L <= 0 | Spec edit: R15 states there is no fill level check; V11 L0 case (fire 0, hitClass 32, no draw) |
| GUESS 3 replacement width | Spec edit: R18 whole-field replacement with a signed nonzero test (dword load/test/store); V13 cases `0x1234` and -5 |
| GUESS 4 read timing/source | Spec edit: R18 says the read happens after the elemental fill, from the same selected record, and is not a logged request |
| GUESS 5 duration condition | Spec edit: R11/R15 and the section 3 row say all three requests pass 1 and duration ignores it; V11/V15 request logs |
| GUESS 6, GUESS 7, GAP 2 array selection | Spec edit: R04b/R10 define the selection as RLI-R02a's unit-read array choice on the given list (never the own-array list read) plus RLI-R02b, and replace the "item wrappers" wording |
| GAP 1 inventory association | Spec edit: section 3 row "caster inventory reference" |
| GAP 3 mastery condition | Spec edit: section 3 row; unit presence is already in the stat-input row |
| GAP 4 section 5 omissions | Spec edit: request row adds source-stat, mastery-stat and stat-325 reads; the R18 row adds the stage result. `skillHitClassValue` is a record-field read under R18 |
| GAP 5 hand-off/return | Spec edit: R18 says the stage returns 1 after hand-off, assertions apply to the record as handed off, and the other dispatch arguments are outside the contract |
| GAP 6 L/skill/caster source | No change: section 3 already has level `L`, skill identity and caster seed as caller-supplied inputs |
| GAP 7 absent record in 53 | Spec edit: R18 early lookup rejection with result 0; V13 failed-lookup case |
| GAP 8, GAP 9 V19 resolver internals | Spec edit: V19 names the resolver guards (eligibility, reach) and the automatic-hit condition as supplied stubs owned by Q06, and separates the request's missile condition from the hit-roll's final argument. SBA was not added as a prerequisite, because these stubs are supplied inputs |
| GAP 10 resolver result use | Spec edit: R13 makes no assertion about how the result is used |
| GAP 11 V13 initial record | Spec edit: V13 starts the record with stale `0x55` in every field |
| GAP 12 Q10 evidence anchor | Evidence-link edit: the Q10 row links to the spec-review correction record |
| DISCRIM 1, 2, 4, 5, 6 (V13) | Spec edit: V13 adds a do-function 38 case with value 7 (hitClass 32), assertions on the record at hand-off, the stale-record clear, `0x1234`/-5, and EType 3 with value 7 |
| DISCRIM 3 flag-before-fill | No change: equivalent. The fills write only R14/R16 fields and never read flags, so setting flag 2 before or after them is unobservable at hand-off |
| DISCRIM 7, 8 V11/V15 order | Spec edit: exact request logs with condition arguments |
| DISCRIM 9 mulDiv argument order | Spec edit: large-value cases in V03 (`0x19B0000` vs `0x19AEA00`), V05 and V07 (105216 vs 105194), and V14 (21043 vs 21030) |
| DISCRIM 10 V06 gate | Spec edit: case with EMinLev2 1 and EMinLev1 0 gives `0x100` with no evaluation |
| DISCRIM 11 V17 classification | Spec edit: classifications 0 and 3 use the caster |
| DISCRIM 12 V14 mapping | Spec edit: the other mastery stats are 0 in the mapping cases. The order is already discriminated by the 500/501 case |
| DISCRIM 13 V19 addend scope | Spec edit: V08 stat 325 = -7 case still gives 60; V19 asserts the stat 325/layer 0 read |
| COVERAGE R17 k=1/2 | Spec edit: V12 cases d/e (seeds `{3,0x29A}`, `{4,0x29A}` give types 2 and 4) |
| COVERAGE R18 items | Covered by the V13 additions above |
| Other COVERAGE (R04a, R09, R10, R11, R12, R14, R15, R16, R13, R07, R05/R19) | No change: these are advisory. The rules state each branch's recipe, and existing vectors already discriminate those results or the branch does not change the recipe (R04a's SrcDam 128 case already distinguishes zero-extension) |
| HARDEST R18/V13 | Addressed by the R18/V13 edits above |
| HARDEST R04b/V17 | Addressed by the array-selection and inventory-reference edits |
| HARDEST R10/V14 | Addressed by the array-selection wording; the order is discriminated by 500/501 and the argument order by 21043/21030 |

These edits change rule statements (R15 level handling, R18 rejection/return), inputs and vectors, so another reader run is required.

### Fourth full fresh-reader pass, 2026-10-15

This was a full review, not a delta check, because the R15 level handling and the R18 rejection/return statements changed after the third run. Counts: MISMATCH 0, GUESS 4, GAP 8, DISCRIM 9, GATED 13, COVERAGE 17, HARDEST 3. Every vector case was recomputed without a difference.

A triage reading of D2Common `0x50460` on 2026-10-15 shows that the mastery block (unit and condition tests, then `0x4F870`) follows the synergy block whether or not the minimum gate rejected evaluation.

| Entry | Outcome |
|---|---|
| GUESS V02 shift-only injection | Spec edit: these cases apply R06 directly to a supplied v |
| GUESS V14 damage input 1000 / cold input | Spec edit: the mapping cases apply R10 directly to a supplied post-synergy e of 1000, and the cold case states input 1000 |
| GUESS V17 SrcDam/request granularity | Spec edit: SrcDam 128 and include-source 1. Each endpoint makes its own weapon/classification/stat requests (`0x52100` calls them itself); the six-request log is given; outputs are named as the selected source values x |
| GUESS R17 positive length kept | Spec edit: V16 type-10 case with length 30 keeps 30 (R17 already stated this) |
| GAP V19 resolver arguments/guards | No change: the resolver internals are supplied stubs and explicitly gated by Q06. Only the sum, wrap and forwarding are SLD assertions |
| GAP V13 hand-off form/eligibility/L | No change: R18 places the dispatch arguments and eligibility outside the contract. The DPO remark is limited to physical/critical draws, and the assertions stop at hand-off |
| GAP V03 source x origin | Spec edit: x is named as the supplied R04b selected endpoint, as in V04 |
| GAP V17 log granularity | Spec edit: as for the V17 GUESS above |
| GAP V09 other record fields | Spec edit: every other field, including hitClass, is preserved |
| GAP R16 optional outputs have no consumer | No change: section 5 lists them; their consumers are callers outside this boundary |
| GAP V20 "reference 1" log form | Spec edit: stated as an evaluation at offset 1 with remaining length 3 |
| GAP Q10 column identity | No change: already gated by Q10 |
| DISCRIM fill range always power of two | Spec edit: V10 range `0x401` case, r 101, physical `0x465` |
| DISCRIM type 10 always length 50 | Spec edit: V16 length 30 case |
| DISCRIM only positive percentages | Spec edit: V05 calculation -20 (6964), V07 calculation -20 (71), V14 stat 329 = -30 (adds -300) |
| DISCRIM other selectors mapped | Spec edit: V14 EType 6 adds 0 with no read |
| DISCRIM mastery skipped after gate rejection | Spec edit after the asm check: V14 gate-rejected minimum gives 332 |
| DISCRIM V04 SrcDam 128 only as w | Spec edit: endpoint output `0x1400` (sign-extended would give `0x1A00`) |
| DISCRIM physical fill at L0 | Spec edit: V10 L0 case gives physical `0x400` with no draw |
| DISCRIM V19 guard order | No change: the guard semantics belong to the hit-resolution contract (Q06); R13 asserts only the forwarded bonus when the roll is reached |
| DISCRIM Kick before lookup | Spec edit: V18 failed-lookup case gives 1/2 with no Kick request |
| COVERAGE (17 advisory) | No change beyond the vectors added above. Each remaining branch follows a stated recipe already distinguished by existing cases |
| HARDEST V14, V13, V19/V17 | Addressed by the V14/V17 clarifications above. V13 and V19 boundaries are stated |

**Outcome: the review is not complete.** The parent's cap of four reader runs is reached. All post-run-4 edits are vector clarifications or additions, plus the V09/V20 wording; no rule text changed after run 4. Their values were hand-computed by the spec writer, but no fresh reader has checked them. A delta check of V03, V04, V05, V07, V09, V10, V14, V16, V17, V18 and V20 against the findings above is still required before the blind-implementer pass can be called complete.

### Fifth reader run (delta check), 2026-10-15

A delta check, not a full review: no rule's behavior, input, ownership or gate boundary changed after the fourth full run. The fresh reader received the run-4 changes to V02, V03, V04, V05, V07, V09, V10, V14, V16, V17, V18 and V20 with the findings each resolves, plus R15/R18 with V11/V13. Counts: MISMATCH 0, GUESS 3, GAP 6, DISCRIM 6, GATED 9, COVERAGE 10, HARDEST 3. Every changed case was recomputed without a difference, and every listed run-4 finding was reported resolved except "other selectors mapped" (only EType 3 and 6 tested).

Triage reread original instructions on 2026-10-15: D2Game `0x92CF0` and `0x93340` load `EAX = 1` after the call to `0xA0C70` and return it, ignoring the hand-off's return; D2Common `0x4F870` decrements EType and rejects unsigned values above 11 with 0 (so EType 0 and 13 add 0), and its cold path requires the definition count to be greater than 0x14B (331).

| Entry | Outcome |
|---|---|
| GUESS V14 mapping-only entry | Spec edit: mastery condition 1, unit present, ordinary getter for 329/330/332 stated |
| GUESS V10 L0 starting record | Spec edit: record.physical `0x100` and seed stated |
| GUESS V13 physical include-source | Spec edit: exact five-request log with include-source 0 and condition 1 |
| GAP V17 request arguments/endpoint return | Spec edit: log `weapon(U,0)`, `classification(U,I)`, `stat(W,23/24,layer 0)`; endpoint returns `0x1E00`/`0x2800` with stated L1 and zero bases |
| GAP V04 SrcDam 64 / x -3 output | Spec edit: endpoint `0x1600` (flooring gives `0x1500`) |
| GAP V13 hand-off return | Spec edit after the asm check: supplied hand-off returns 0, stage result still 1 |
| GAP V14 read count | Spec edit: one stat 329 read per endpoint |
| GAP V09 no nonzero witness | Spec edit: every other field starts at `0x55` and is preserved |
| GAP V16/V11 type-9 resist / absent optional outputs | No change: derivable from R16; V16 exercises present outputs |
| DISCRIM V13 stage returns hand-off result | Spec edit: as GAP V13 |
| DISCRIM V09 fill zeroes other fields | Spec edit: as GAP V09 |
| DISCRIM V04 flooring division | Spec edit: as GAP V04 |
| DISCRIM V14 other selectors mapped | Spec edit after the asm check: EType 0, 7, 8, 9, 10, 11, 13 with every mastery stat 30 add 0, no read |
| DISCRIM V14 ordinary-getter own array; cold count bound | Spec edit: stat 329 own 10/resolved 30 adds 300; cold count 332 reads and adds 300 |
| DISCRIM V17 classification argument | Spec edit: as GAP V17 |
| COVERAGE (10 advisory) | No change: each branch follows a stated recipe; the R10 selector and array branches are now covered by the additions above |
| HARDEST V14, V13/R18, V17/R04b | Addressed by the clarifications above |

These are vector additions/clarifications only; rule text is unchanged, so the next run is a delta check of V04, V09, V10, V13, V14 and V17.

### Sixth reader run (delta check), 2026-10-15

Delta check of the run-5 edits to V04, V09, V10, V13, V14 and V17. Counts: MISMATCH 0, GUESS 1, GAP 3, DISCRIM 5, GATED 9, COVERAGE 6 groups, HARDEST 3. Every case was recomputed without a difference. The reader marked the V04, V09, V10, V13 and V17 findings resolved and V14 partially resolved (own-array rejection shown only for stat 329).

Triage reread D2Common `0x50460` on 2026-10-15: after the lookup and `L > 0` tests, the synergy block tests only the reference, the minimum gate and the pool/bounds before evaluating with context `(unit,skill,L)`; the unit is tested only in the later mastery block (condition, then unit). D2Game `0x92CF0`/`0x93340` return 0 on failed lookup before the record clear. RLI-R02a already states the `id >= statDefinitionCount` rejection for the ordinary getter.

| Entry | Outcome |
|---|---|
| GUESS V14 which array holds mapped values | Spec edit: V14 defaults state `usesResolvedStats` true, given values are resolved values (own 0), definition count 359 |
| GAP V14 no operation applies R10 to a supplied e | Spec edit: damage-input cases execute as an elemental maximum with EMax = damage input, HitShift 0, empty synergy, mastery condition 1; the return is input plus amount |
| GAP V17/V14 stat log meaning | Spec edit: the section 8.1 preamble says stat log entries trace the real RLI-R02a/RLI-R02b getter and are not supplied operations |
| GAP V14 absent unit with synergy | Spec edit after the asm check: the absent-unit case still makes its evaluation; R09 has no unit condition |
| DISCRIM V04 round-half-up | Spec edit: SrcDam 64/x 11 gives `0x1C00` (half-up gives `0x1D00`) |
| DISCRIM V04 unwrapped product | Spec edit: SrcDam 255/x `0x808081` gives `0x1800` (64-bit product gives `0x1700`) |
| DISCRIM V14 own-array for 330/332 | Spec edit: EType 1, 2 and 5 array cases each add 300 |
| DISCRIM V14 stat 332 count bound | Spec edit: count 332 adds 0, count 333 adds 300 (RLI-R02a bound) |
| DISCRIM V13 failed lookup clears record | Spec edit after the asm check: stale `0x55` record unchanged on failed lookup |
| COVERAGE (advisory) | No change: each branch follows a stated rule or NRF recipe; the additions above cover the R04a, R10 and R18 items that change a result |
| HARDEST V14, V13, V17 | Addressed by the clarifications above |

Vector additions and clarifications only; rule text is unchanged. The next run is a delta check of V04, V13 and V14 plus the section 8.1 preamble.

### Seventh reader run (delta check), 2026-10-15

Delta check of the run-6 edits to V04, V13, V14 and the section 8.1 preamble. Counts: MISMATCH 0, GUESS 2, GAP 2, DISCRIM 2, GATED 12, COVERAGE 7, HARDEST 3. Every case was recomputed without a difference, and every run-6 finding was reported resolved.

| Entry | Outcome |
|---|---|
| GUESS V04 origin of x in new cases | Spec edit: the independent cases state x is the supplied R04b selected endpoint (caster stat 21, absent weapon) |
| GUESS V14 definitions at overridden counts | Spec edit: every definition below the count is present |
| GAP V13 stage failed lookup only for identity >= count | Spec edit: identity -1 and absent record at identity 0 added; the `0x92CF0`/`0x93340` readings recorded under run 3 and run 6 show the signed-negative, count and null-record tests all return 0 before the clear |
| GAP reader read the expression spec's mastery rule by grep only | No change: a limit of the reader's check; R10 restated the amount that the expression spec then owned (R10 owns it since the 2026-10-15 restructure) |
| DISCRIM V13 partial lookup rejection | Spec edit: as the GAP above |
| DISCRIM V14 bounds for 329/330 | Spec edit: count 329/330 for stat 329 and 330/331 for stat 330 (RLI-R02a `id >= statDefinitionCount`) |
| COVERAGE (7 advisory) | No change: fMin adjustment, own-array lists, absent lists/entries and multi-field routing follow RLI-R02a/RLI-R02b and RLI-R01 recipes already discriminated elsewhere; duration ignores mastery by R11 |
| HARDEST V14, V13, V04 | No further change: the cases are now explicit |

Vector additions only; rule text unchanged. The next run is a delta check of V04, V13 and V14.

### Eighth reader run (delta check), 2026-10-15

Delta check of the run-7 edits to V04, V13 and V14. Counts: MISMATCH 0, GUESS 2, GAP 3, DISCRIM 0, GATED 10, COVERAGE 7, HARDEST 3. Every case was recomputed without a difference, and every run-7 finding was reported resolved. The reader tried sign-extension, flooring, half-up rounding, 64-bit products, share-after-shift, clearing or flagging before lookup, returning 1, and uniform cold bounds; each is rejected by an existing assertion.

| Entry | Outcome |
|---|---|
| GUESS V14 bound-case damage input | Spec edit (value the reader already used): damage input 1000 stated |
| GUESS V14 overridden count scope/defaults | Spec edit (as assumed): the stated count replaces 359 for every read in the case; definition defaults apply |
| GAP V04 gate cases lack an output | Spec edit (derivable value the reader computed): output `0x1700` with x 10 available |
| GAP V13 unreached hand-off stub / hit-class value | Spec edit (wording): neither is reached on failed lookup |
| GAP V13 identity of the stale storage | Spec edit (wording): the storage a successful stage would clear |
| COVERAGE (7 advisory) | No change: each branch follows R04a, RLI-R02a/RLI-R02b or NRF-R13 recipes already discriminated elsewhere; the eligibility branch is outside the contract |
| HARDEST V04, V14, V13 | No change: all cases are explicit and recomputed |

The post-run-8 edits state values and readings that the run-8 reader itself computed or assumed, so they change no output or assertion and introduce no new behavior; no fresh reader checked their wording, because this task's cap of four reader runs (runs 5 to 8) was reached.

**Outcome: the blind-implementer pass is complete at the supplied-input/static boundary.** Runs 1 to 4 were full reviews and runs 5 to 8 delta checks; no rule text changed after run 4. The final applicable run returned MISMATCH 0 and DISCRIM 0, and its GUESS and GAP entries are resolved by the edits above. Producer, integration and observation gaps (Q02, Q04 to Q10) remain; header status stays draft. No human approval is inferred.

### Restructure (2026-10-15)

The approved review split the former expressions-and-synergies spec, now
[skill-calc-expressions](../../specs/skill-calc-expressions.md). Its five-band progression,
elemental and physical damage builders, mastery and duration rules restated SLD-R01a through
SLD-R11 and were removed there; this spec owns them without renumbering any SLD id. Their unique
vector cases were folded into existing vectors as independent cases: increments `[1,2,3,4,5]` in
V01; the same-input physical/elemental rounding pair and the negative-source maximum case in V03
and V05; the zero-percent two-evaluation case in V06; and the `[10,5,2]` duration cases in V07.
The expression spec's mastery vector was already covered by V14. Its physical source/Kick question
merged into Q02 and Q08 (server caller flags into Q05), and its elemental-synergy observation
into O02. R10 now states the mastery amount itself; R19, Q04, the dependencies and Related specs
cite the renamed expression spec. No rule behavior, label, value or existing vector result changed.
The folded cases were hand-checked against R02 through R11 and the NRF-R05/NRF-R06a contracts;
no source reading, gameplay observation or fresh reader pass was performed for this restructure.
