# Hit recovery evidence

Behavioral contract: [Hit recovery and Faster Hit Recovery](../../specs/combat-hit-recovery.md).
Original artifacts: [source manifest](../source-manifest.md). All descriptions below are analyst labels,
not original exported function names. RVAs are module-relative; exports are ordinals of the
identified original Windows x86 Expansion 1.13c artifacts. No other game build supplies constants.

## Evidence dates and reuse

The original 2026-10-11 static instruction reading and 2026-10-12 audit are retained for their
actual covered slices. They used disassembly; readable decompiler output for the rate dispatcher
and elemental helper was corroboration/lead evidence, not independently verified source. Moving
those records does not renew their verification date. Original text and private fixture encodings
are preserved in a separately integrity-recorded private archive. Private investigation paths,
reconstruction filenames, commits, sessions and listings are not required to use this companion.

The 2026-10-15 regeneration assessed those records against the current shared contracts before
targeted source reading. The original D2Game/D2Common specimens agree with the source-manifest
fingerprints. The source availability index was checked before new reading: the animation end
predicate at D2Common+0x2EF30 has a matching C reconstruction, which was read completely; the
other newly read targets below have no C view checked against original bytes and were read as disassembly. Comments on that
predicate misname animation fields as life/death fields; operations, not comments, establish meaning.

No gameplay observation, save editing, memory write, hook change or implementation audit was
performed for this regeneration. Fresh table reads establish narrow text-cell evidence only.

## Dispatch and source references

| Artifact | Function RVA / ordinal | Analyst description and root | Rules / actual coverage |
|---|---|---|---|
| LOD113C-D2GAME | 0xDE5A0 | Melee resolution reaches result OR after block/avoid mapping | R01a; retained 2026-10-11 reading, consumed SBA-R13 |
| LOD113C-D2GAME | 0x3AE10 | Missile contact/owner/record stage, result preparation then block/avoidance | R01b/R01c; 2026-10-12 reading; request/rewrite slice re-read 2026-10-15, no C view checked against original bytes; SBA-R14/SBA-R20 carry producer gates |
| LOD113C-D2GAME | 0xDCF20 | Combat-result consumer, reached after applied damage | R02–R05/R07/R08/R11; retained 2026-10-11/12 reads; entry through protected dispatch and complete player cooldown/GH/soft tail re-read 2026-10-15 |
| LOD113C-D2GAME | 0xDB910 | Veto invoked by monster/player GH branch in 0xDCF20 | R05/R06; complete 304-byte body read 2026-10-15, agreeing with retained audit; no C view checked against original bytes |
| LOD113C-D2GAME | 0x1450, 0x1180 | Mask/modulus draw helpers called with defender seed | R06/R10b; NRF-R01/NRF-R02 and retained 2026-10-12 seed-owner audit |
| LOD113C-D2GAME | 0x1560 | Monster mode availability predicate | R06/R09a; compatible GDF/KNB definition/availability contract; not a complete monster wrapper |
| LOD113C-D2GAME | 0xDA660 | Called from hit-class synthesis guard in 0xDCF20 | R07a; entire 112-byte listing and matching original-image instruction decode, 2026-10-15; no callees |
| LOD113C-D2COMMON | 0x6A050 / #10677 | Missile class getter called in 0x3AE10 | R07a initial missile class; complete 80-byte body re-read 2026-10-15, no C view checked against original bytes; u8 load, missile/type/class/record guards |
| LOD113C-D2GAME | 0x771D0, 0x75070, 0x75E10 | Monster flag conversion, parameter build and mode wrapper | R04/R09/R15; retained 2026-10-11/12 reading; 0x75E10 entry through mode-3 skip re-read 2026-10-15, complete wrapper not claimed |
| LOD113C-D2GAME | 0x75210 | GH start selected by monster mode table | R09a; retained reading and complete 80-byte start re-read 2026-10-15 after availability check (no C view checked against original bytes); entry preparation precedes NRF-R14a mode write; reset/wrapper effects gated |
| LOD113C-D2GAME | 0x75320 | GH tick selected by mode table | R09b; complete short entry through return read from original image 2026-10-15; no separate source listing; calls 0xAE6F0, unread |
| LOD113C-D2GAME | 0x76BF0 | Shared monster mode end handler | R09b; source located; entry/class-selection slice assessed 2026-10-15; complete class/callback effects remain open |
| LOD113C-D2GAME | 0x78550 → 0x78340, 0x77FA0 | Player mode wrapper, validator and local action decision | R10; retained 2026-10-12 reading; 0x78340/0x77FA0 complete listings read 2026-10-15; unresolved byte tables/validator helpers prevent full wrapper claim |
| LOD113C-D2GAME | 0x1420 | Current-skill definition getter used by 0x77FA0 | R10a; 2026-10-12 trace establishes Skills stride 0x23C and srvdofunc field; GDF-R08 |
| LOD113C-D2GAME | 0x78150 → 0xDA750 | Player GH entry and neutral entry preparation | R10c; complete listings read 2026-10-15; full 96-byte 0xDA750 original-image decode checked; other entry helpers unread |
| LOD113C-D2COMMON | 0x33110 / #10819 | Per-mode rate dispatcher reached from animation preparation | R12a/R14; retained 2026-10-12 GH/block/KB branch reading and original engine-table check; no C view checked against original bytes at 2026-10-15 availability assessment |
| LOD113C-D2COMMON | 0x323E0 | Diminishing reader called by rate dispatcher | R12a/R14; complete body re-read 2026-10-15, no C view checked against original bytes; selected-array leaf 0x38A80, consumed NRF-R07/NRF-R12b/NRF-R12c |
| LOD113C-D2COMMON | 0x2EFC0, 0x2EF90, 0x2F010, 0x2FB10 | GH/block/KB predicates and base-mode rate | R12/R14; compatible retained branch readings; no newly inferred per-class record |
| LOD113C-D2COMMON | 0x501E0 | Ordered transformation-driven effective type/mode selection | R12b; complete listing assessed 2026-10-15, no C view checked against original bytes; helper/table chain incomplete, no GH remap promotion |
| LOD113C-D2COMMON | 0x335F0 / #10595 | Animation preparation following mode write | R13a; complete 816-byte listing read 2026-10-15; remap/helpers/byte region incomplete, actual GH initialization gated |
| LOD113C-D2COMMON | 0x2F090, 0x2F060 | Two isolated accumulator advances | R13d; complete 48-byte listings read 2026-10-15, no C view checked against original bytes; caller not thereby established |
| LOD113C-D2COMMON | 0x2EF30 / #10426 | Animation next-step/end predicate | R13d; complete C view checked against original bytes read 2026-10-15; no writes/RNG; server GH caller remains open |
| LOD113C-D2GAME | 0x77240 | Reached monster elemental-slot processing | R15; retained 2026-10-12 decompiler/seed audit; complete disassembly read 2026-10-15, unvalidated byte runs and callees remain; direct slot draws consume NRF-R04b |

### Root anchors for the reaction consumer

The table above names 0xDCF20 as the computing consumer; its dispatch roots are the
damage-pipeline and missile chains already recorded by their owners. Queued path
([DPO companion](combat-damage-pipeline-order.md), DPO-R26): frame loop D2Game 0x79A10 ->
frame-event table 0x1079D0 -> 0x79170 -> do entry 0xA2060 -> do dispatch 0xA19A0 -> do
table slot 1, basic attack 0xA2C40 -> queued completion 0xDDDE0 -> enqueue 0xDAB90; the
queue consumer 0xDED30 calls reaction 0xDCF20 after apply 0xDE0C0. Missile direct path
(MHD-R26, SBA chain B): missile step 0x3F320 -> contact 0x3EDE0 -> damage fill/hand-off
0x3A4E0/0x3AFF0 -> apply 0x3AE10, whose tail calls 0xDCF20 unconditionally. The two final
call edges (0xDED30 -> 0xDCF20 and 0x3AE10 -> 0xDCF20) were re-checked in the original
disassembly during the 2026-10-15 specification review; upstream links reuse the owners'
dated records and add no new behavioral claim.

### Imported operations and semantic mapping

These resolved original imports separate common-library primitives from game wrappers.
Their owned contracts are the dependency authority; a resolved address alone does not close effects.

| D2Common export | Resolved RVA | Semantic operation / rules |
|---|---|---|
| #10494 | 0x33CD0 | State predicate; NRF-R12a, HRC-R01/R03/R06 |
| #10871 | 0x38520 | State-list lookup; NRF-R12a, HRC-R05/R10 |
| #10913 | 0x390F0 | Fixed stat-7 selected/minimum wrapper, HRC-R06 |
| #10945 | 0x34510 | SetState routing, NRF-R12a/STL-R07, HRC-R03a |
| #11090 | 0x33920 | Mode write, NRF-R14a, HRC-R09a/R10c |
| #11103 | 0x303E0 | Current-skill clearing, HRC-R10c |
| #10677 | 0x6A050 | Guarded missile class byte, HRC-R07a |
| #10819 | 0x33110 | Animation rate dispatcher, HRC-R12a/R14 |
| #10595 | 0x335F0 | Animation preparation, HRC-R13a |
| #10426 | 0x2EF30 | End predicate, HRC-R13d |
| #10331 → #10057 | 0x2FE10 → 0x3C390 → 0x2D320 | Room/level/town chain, KNB-R03b, HRC-R02b |
| #10158 | 0x665C0 | Update marking request; full integration NRF-Q17, HRC-R08/Q10 |

The max-life wrapper 0x390F0 was checked for C view checked against original bytes (none), then read fully on
2026-10-15: null unit/list, unavailable definition count<=7 or missing definition yields 0;
otherwise fixed stat 7 uses 0x38A80. It is not an aggregate-array read. The selected/minimum
leaf semantics are NRF-R12b/NRF-R12c; 0x38B70 is not the called leaf here.

### Original field map

The main contract requires semantic values only. This map permits reproducing the reading
with matching original files; it does not prescribe a new storage layout.

| Original field / table | Semantic value / supported assertion |
|---|---|
| Damage+4 word | R |
| Damage+0x4C/+0x28/+0x24/+0x10/+0x1C dwords | total/poison/cold/fire/lightning |
| Damage+0x60 dword, +0x64 byte | hitClass i32, synthesis suppression u8 |
| Unit+0xB0 dword | Defender lastHitClass |
| Unit+0x20 two dwords | Seed lo/hi |
| Unit+0x10, +0xC4 bit 0 | Mode, modeChanged |
| Unit+0x4C word, +0x44/+0x48 dwords, +0x30 | Rate, accumulator, total, animation condition |
| Unit+0x50 record +8/+0x0C | Animation frame count/base rate B; actual selected GH record open |
| Unit+0xA8 current skill +0x10; Skills+0x2E word | Current skill and srvdofunc (function indices 67/76, not skill ids) |
| Skills flags +4 dword bit 31 | interrupt |
| Game+0xA8 | Current frame |
| MonsterData+0x54 | Reaction synchronization code (full observable effect remains gated) |
| MonStats2+0xF0 bit 3/bit 13; MonStats+2 signed word | mGH/mKB; BaseId |
| MonStats+0x36 signed word | Monster KB base rate; loader column unresolved GDF-Q01 |
| Missiles+4 bits 6/7, +0xAE byte, +0x94 byte | GetHit/SoftHit, KB chance, initial HitClass |

The divisor jump table at D2Game RVA 0xDBA04 (preferred VA 0x6FCFBA04) was read in the
2026-10-11/12 audit; 2026-10-15 complete veto reading confirms unsigned whole-argument-minus1
bound<=11 before dispatch. Nonzero high bits therefore take default 16. Fog bit masks
are `1<<i`, from original PE table evidence retained in the audit, not label names.

The monster mode table is D2Game RVA 0xFA498 (preferred VA 0x6FD1A498), 16-byte rows.
The GH row at RVA 0xFA4C8 was checked against the matching original image on 2026-10-15:
start 0x75210, tick 0x75320, end 0x76BF0, fourth dword 1. Tick 0x75320..0x7534C was read through
return: optional Unit+0xDC chain, next+0x24, node byte 0==0 calls 0xAE6F0, then return 1.
The adjacent function is not included in that complete-entry claim. The unread helper
and server driver prevent animation-end inference.

Rate table at D2Common RVA 0x94608 (preferred VA 0x6FDE4608) has 12-byte entries
{flag, A, stat}; its consumed values are owned by NRF-R07. The 2026-10-12 original-table
reading establishes FHR index 1/A120/stat 99 and FBR index 3/A120/stat 102. Per-mode predicates
and branch precedence agree with MMF-R06f. KB selects its requested mode, not the previous
walking/running mode, then invokes path-speed setter 0x34D40 (NRF-R08b) with 4096.

## Corrections and affected-claim assessment

| Discrepancy in prior text | Original evidence and corrected claim | Remaining boundary |
|---|---|---|
| Stun bypass described as an active-state test | 2026-10-15 reaction-branch reading confirms Common #10871 / 0x38520 FindStateList(21), not active-bit testing; NRF-R12a/STL distinguish list presence from membership | R05/V03 now require a matching stun list; state 21 alone does not bypass veto |
| Protected lethal branch called an ordinary death operation with id 92 | 2026-10-15 import resolution/reading identifies D2Common #10945/0x34510, request state 92, value 1; current NRF-R12a and STL-R07 own routing/membership | State 92 text name death_delay does not establish death/timer semantics; STL-Q07 disguise/NRF-Q17 update execution still open |
| Class described as a word and helper retaining only low nibble | 0xDCF20 entry re-read 2026-10-15 uses dword+0x60 and byte suppression; 0xDA660 retains entire source unless 0=>13; matches current NRF/DPO i32 contract | Existing small numeric results unchanged; V16 adds high-bit preservation |
| Element rotation was insufficiently verified | Full 112-byte helper and hash-matched image decode plus 64-byte repeated pair table at Game RVA 0xFB050 read 2026-10-15 | Helper algorithm closed at supplied-record boundary; no RNG, shared byte cursor increments; upstream producer remains separate |
| Town set unresolved | Compatible KNB-R03b proof reuses complete Common #10331/0x2FE10→#10057/0x3C390→0x2D320 and checked table; exact {1, 40, 75, 103, 109} | Source/room records supplied; no client claim |
| Missile rewrite left asm-only | Current SBA-R14 already verifies same request/rewrite stage; 0x3AE10 exact slice re-read 2026-10-15 agrees | Upstream owner/contact/block producers retain SBA gates |
| Player decision treated missing current skill as idle/walking; exception treated as skill ids | Full 0x77FA0 read 2026-10-15 agrees with current-skill fast allow, matching passed skill and srvdofunc 67/76; guard/roll order isolated |0x78340 validator byte tables and outer entry/helper effects unresolved; no global action-state equivalence |
| Player GH preparation said to remove all timed events | Full 0xDA750 image decode shows matching-unit linked-record removal/free; node semantics not established | Neutral entry-preparation request only; current-skill write is separate; actual cancellation/timer integration open |
| Block numeric predicate was conflated with completed animation | Complete cooldown branch re-read 2026-10-15 establishes strict signed delta>trunc(FBR/8)+15, soft early return, mode request then unconditional stat-95 write request carrying current frame | Actual mode start and persisted writer effect remain gated; V06f supplies returned 0 to distinguish request order; fixture persistence is explicitly supplied |
| GH duration was an unqualified ceiling formula; some data said absent | Complete 0x2F090/0x2F060 and matched 0x2EF30 establish isolated primitives; 0x335F0 branch facts/remap remain partial; loader establishes animation file presence | Actual GH driver, record/frame/start/end and breakpoints remain precise integration questions; no six-rung absence claim |
| Full mode change said to draw nothing for GH | The requested-mode-3 skip is established in 0x75E10; 0x77240 contains unvalidated byte runs/unread helpers; NRF-R04b verifies reached direct slot draws | Keep complete event draws unresolved; local rate/start no-direct-draw claims do not establish full event counts |

The block cooldown's stat-write call is a boundary, not a newly read persisted write. The
2026-10-15 caller slice at D2Game+0xDCF20 pushes layer 0, current game frame, stat 95 and
defender before its stat-writer thunk at D2Game+0xA472. The writer interior was not read in
this pass and a complete resolved target/write contract is not established by that caller
reading. HRC-R11 therefore verifies the request and order only; HRC-Q06 retains original
persisted execution. V06/V07/V21 explicitly supply a writer that persists the value, so their
fixture stamp 100 is not a claim that the original writer was verified.

During coordinating review the block-rate expression was found to contain an accidental
assignment-like D=3 token. The retained rate-branch reading and NRF-R07 establish diminishing
index 3 with A=120 and FBR stat 102. HRC-R14 now defines its result as D and explicitly forms
`wrap32((base+D)*B)`, then unsigned division by 100. V15's existing D=24 and rate 189 for FBR 30,
B=256 are unchanged. This corrects expression spelling, not the original arithmetic.

The main text retains V01–V15 numeric identities and results, with concrete seeds instead of
probability-only cases. V12's death interpretation is corrected by the resolved import; it
now asserts the state-on request/membership. V13's zero-draw claim is narrowed to the supplied
start boundary; mode 3 and marker results persist. V15 preserves exact path speed 4096.
V01 probabilities remain derivations with explicit stepped cases. Negative-rate wording now
follows wrapped product rather than claiming all negative factors saturate under every overflow.
Historical findings in the implementation-gap section are retained as historical, not newly audited.

### Remaining gap assessment

Every retained gate was checked against the dated records and current shared contracts first.
Additional targeted reads were chosen for unanswered portions: 0x501E0/0x335F0 for actual
remap/preparation; 0x75320 original entry/0x76BF0 class entry for server end; 0x78340/0x78150/
0xDA750 for validator/entry; 0x77240 and 0x75E10 skip for event draws. These readings add local
facts but do not read the missing helper/table chain. Other producer coverage remains a
bounded caller-search question, not a claim of absence. Client behavior was not searched as
a new campaign. Gameplay observations were intentionally excluded from migration and remain
pending. Exact blocked assertions and smallest next checks are in the [gap file](../../plans/gaps/combat-hit-recovery.md).

## Named table cells

These cells are **verified in 1.13c data**, not runtime-confirmed behavior. The 2026-10-15
answering member reported the vanilla LoD lod-113c profile; current member inventory was
checked and four PD 2 members excluded. Historical vanilla profile evidence was reused.
No fresh complete override inventory, current version catalog, physical supplying MPQ, loaded
DLL hash or compiled-record correspondence was captured. Existing hooks were untouched.
Thus text identities below have that profile scope, not a general claim of pristine archives
or runtime table equivalence. The loader was the game's own Storm virtual-file access.

| Effective virtual file | Date/range/coverage | Effective SHA-256 |
|---|---|---|
| data/global/excel/states.txt |2026-10-15, offset 0, length 0(full), returned 20920/20920 bytes |898806c2b69e55e316ecfed48dde95aa2933829b94fdad234bdabc27bdbc23e1 |
| data/global/excel/itemstatcost.txt |2026-10-15, offset 0, length 0(full), returned 40109/40109 bytes; compatible same-day collection reused |ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375 |
| data/global/AnimData.d2 |2026-10-15, offset 0, length 32, returned 32/570304 bytes |Not captured: presence/header-only read |

Both complete text files were parsed by named headers. IDs are unique and agree with
zero-based data-row index excluding the header. Each named cell below is a separate narrow
claim; no full rows/columns are reproduced. Raw decimal ID text converts directly to integer.

| File / column | Row identity / ID column raw cell | Name raw cell | Supports |
|---|---|---|---|
| States.txt / state |id=1 |freeze |R05/R06 |
| States.txt / state |id=15 |concentrate |R10a |
| States.txt / state |id=21 |stunned |R05 |
| States.txt / state |id=42 |concentration |R10a/R10b |
| States.txt / state |id=54 |uninterruptable |R01/R03 |
| States.txt / state |id=65 |dodge |R04b |
| States.txt / state |id=66 |avoid |R04b |
| States.txt / state |id=68 |evade |R04b |
| States.txt / state |id=92 |death_delay |R03a name only |
| States.txt / state |id=101 |holyshield |R14 |
| ItemStatCost.txt / Stat |ID=7 |maxhp |R06 |
| ItemStatCost.txt / Stat |ID=93 |item_fasterattackrate |NRF diminishing-table mapping |
| ItemStatCost.txt / Stat |ID=95 |lastblockframe |R11 |
| ItemStatCost.txt / Stat |ID=96 |item_fastermovevelocity |NRF diminishing-table mapping |
| ItemStatCost.txt / Stat |ID=99 |item_fastergethitrate |R12a |
| ItemStatCost.txt / Stat |ID=102 |item_fasterblockrate |R11/R14 |
| ItemStatCost.txt / Stat |ID=105 |item_fastercastrate |NRF diminishing-table mapping |
| ItemStatCost.txt / Stat |ID=164 |skill_concentration |R10b |
| ItemStatCost.txt / Stat |ID=352 |last_sent_hp_pct |R08 |

The animation read proves only effective virtual-file presence. It does not identify a GH
record, frame count, B, initial accumulator, duration, breakpoint or full-file hash. Other mode,
loader and class-name mappings consume their dated GDF records without claiming a fresh read.

## Pending observation procedures

No procedure below has run. Use legally owned matching original binaries and a controlled
vanilla Expansion 1.13c profile. For a baseline use Normal difficulty, a level 20 Amazon, no
mercenary, no passive avoidance, no gear granting unrelated combat effects, and an outdoor
cleared area with room behind the target. Record actual character/items/skills/effective stats,
target class/base/level/states/mode, room/level, damage components/flags/class, suppression, cursor,
all relevant seeds, and effective data identity. A case requiring unavailable controlled inputs
stays unobserved; do not silently replace the case. Do not alter hooks owned by another session.
Each capture establishes loaded-module correspondence, eligibility, full requested fields and
valid record count. Zero records/incomplete captures do not promote any rule. Observe local
returns/requests separately from completed mode/timing. Record unexpected results before
re-reading affected instructions; do not fit vectors to observations.

### HRC-O01

Against the same nonlethal target with measured max life 25600 raw, collect GH-veto calls at
D2Game+0xDB910 with totals 255, 1599, 1600, 3199, 3200, 6399, 6400 when an existing controlled
fixture permits them. Capture defender seed before/after, whole class, states, poison, M,
availability and return for at least 16 eligible calls in each available band. Validate each
record's exact NRF draws/early returns; collect separate missing-mGH controls. Aggregate pass
fractions are secondary, not proof of seed order. This confirms R06 only for observed cases.

### HRC-O02

With equal measured total 8000, M=25600, compare ordinary, matching-stun-list, frozen, and matching-stun-list plus freeze
targets at D2Game+0xDCF20/0xDB910. Record whether veto was called and the selected reaction.
For an existing lawful state 54 fixture, compare nonlethal and lethal result requests; capture
D2Common+0x34510 arguments 92/1 and active/changed membership, then separately observe any
disguise/update effect. Do not label the call ordinary death because of the text name.

### HRC-O03

Hold target/source and total fixed; use available controlled physical, critical, and mixed
cold/fire/lightning/poison hits. Record incoming k, suppression, global byte cursor, components
and outgoing k/lastHitClass at D2Game+0xDCF20/0xDA660. Correlate start index and full 32-bit
preservation with the veto's entire-class divisor. Separately trace a melee class producer
and the guarded missile byte getter D2Common+0x6A050. At least one rotation pair with two
positive components is needed to distinguish cyclic priority from fixed priority.

### HRC-O04

Record a controlled player's actual selected GH record, B, count, initial accumulator and
effective type/mode. With known FHR 0, 10, 30, 60, 120, and an available negative fixture, capture
D2Common+0x33110's rate write and all server ticks until mode completion. Compare rate
arithmetic first; measure duration separately. Trace the actual caller of any observed
accumulator advance/end test; the presence of D2Common+0x2EF30/0x2F090 alone is insufficient.
Repeat with a transformed player and a monster to settle actual remap/data integration.

### HRC-O05

With FBR 0 then 24, record player block requests and stat 95 around frame deltas 15/16 and 18/19
at D2Game+0xDCF20; record soft-bit controls. Distinguish request, outer return, stamp and actual
mode 9 entry. Separately record D2Common+0x33110 block/KB rate and path-speed writes for player
and monster, including exact speed 4096. Collect complete mode-event seed changes separately;
no-direct-draw rate evidence is not a full-event count.

### HRC-O06

For the baseline player and a live monster, record current mode, current/passed skill,
srvdofunc, interrupt, state 42 own 164, state 15, seed and local return at D2Game+0x77FA0. Compare
known active-skill and missing-skill contexts rather than assuming walking means no skill.
Capture outer 0x78340/0x78550 handling, neutral request, 0x78150 entry/current-skill clearing,
and actual mode completion separately. For monster GH, capture table-selected 0x75210 start,
0x75320 tick, 0x76BF0 end and any intervening helper calls. A current-mode 1 refusal needs the
neutral-request record before allowance. Complete elemental-event draw counts require the
whole actual wrapper/callee chain, not only the mode 3 skip.

## Regeneration and review record, 2026-10-15

Scope remains the existing HRC boundary. No source, implementation, test, plan, index, manifest or
GitHub write is part of this writer's changes. The contract is unconsumed by implementation
plans/code/tests; numeric IDs retained for sibling references have no old aliases. V01–V15
retain their identities with the source-proven interpretation corrections above; V16–V22 add
discriminating cases for new supplied-boundary facts. User approval remains pending.

Closed/narrowed gaps: town set; matching-stun-list versus active-bit semantics; monster entry-preparation order; state 92 request identity/membership; class width/helper rotation;
missile rewrite through current SBA; local player GH decision; cooldown request/stamp; isolated
animation primitives. No false six-rung absence or blanket zero-draw label is retained.
Open gaps remain in the linked gap file. The targeted document checker was run after the rewrite and again after self-review corrections. The latest completed run returned exit 0, 0 errors and 0 warnings. Earlier spacing warnings were fixed; historical private-project terminology was replaced with original-code evidence without changing the historical findings. Public method wording identifies the C view checked against original bytes and disassembly without private filenames. No code tests or gameplay actions were run.

Self-review independently recomputed all 22 vectors, including both LCG words, early exits, signed/unsigned rate division, whole-class selection, strict cooldown and dependency request order. Coordinating review then corrected the block diminishing-result expression and narrowed R11 to the actually read stat-write request boundary; original persisted writer execution remains gated. It found and corrected the stun-list distinction and omitted monster entry-preparation/update request logs before handoff. Fixture defaults now supply valid skill definitions, class/base records and no eligible mode-bound lists. Private original encodings remain archived; V01 through V15 identities and small numeric results are retained except the source-proven interpretation corrections above.

### Fresh-reader full pass 1 and triage

The first fresh full reader used only the main contract, its cited dependency contracts and
the review checklist. It did not read this companion, prior reports, implementation, original
source, runtime material or web sources. It recomputed the 22 vector families and returned:
**MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; GATED 31; COVERAGE 16; HARDEST 3**.

There are no mismatch, guess, gap or discrimination findings to resolve. The 31 gated entries
are the seven question IDs, 12 unresolved rules and 12 supplied-stage vector limits. Retain
each stated boundary; their count does not promote any dependent operation or gameplay claim.
The reader reported 16 aggregate coverage entries. The table below groups related entries for
triage without equating its row count to the reader's count.

| Reported coverage area | Outcome |
|---|---|
| R02a monster mode 12 | No change: V11c already exercises that exact guard. |
| R04a unavailable-KB fallback and base conversion | No change: these operations consume KNB-R06a/KNB-R06c and their KNB-V32 discrimination; do not duplicate their owner here. |
| R04a player lethal without earlier flags | No change: the same selected death-request recipe is covered by HRC-V21 case (a) and the player/monster selection contrast in KNB-V11. No different local terminal effect is established. |
| R04a player avoidance priority/effects | No change: selector termination is in V21c; complete avoidance effects remain R04b/HRC-Q06. |
| R04a monster special synchronization alternatives | No change: soft, exceptional classes and unavailable mode 6 select the same synchronization recipe as V21d; no additional result-distinct recipe is needed. |
| R05 active state 21 without list, without freeze | No change: V03d already distinguishes an active bit from a matching list with freeze as a discriminating veto control; removing freeze creates no new local bypass recipe. |
| R06 nondefault-divisor stochastic bands | No change: V01/V22 exercise the threshold/draw recipes; V05 independently discriminates whole-class divisor binding. This does not claim every divisor/band pair was enumerated. |
| R07a suppression, lightning/poison priorities | No change: the same skip/selection recipes apply; V16e covers synthesis skip and V16a/b discriminate cyclic priority. Individual code bindings remain explicit, without an exhaustive binding-coverage claim. |
| R09a mode-12 start | No change: it shares mode-0 V13b's return-1/no-op recipe, and V11c separately covers the ordinary-dispatch mode-12 guard. |
| R10a mode-4/nonzero-context and mode-1 roll-failure exception | No change: the local decision supplies the requested GH mode/nonzero context; V14 exercises the mode-5 refusal/allowance decision, and V18d exercises the explicit refusal-to-neutral exception recipe. |
| R13d second advance below threshold | Added V19g: 100+300=400 below 1024, so no adjustment; rejects always adjusting to -112. |
| R13d null end-predicate input | Added V19h: fatal failure with no numeric return or writes; rejects a fabricated boolean default. |
| R14 block-base fallback | No change: V15a versus V15c supplies the 50/100 base contrast; the other fallback conditions select the same base-50 recipe. |
| R14 nonpositive/capped KB | Added V15g: supplied signed monster base 0 stores rate 0 and still sets speed/targetSpeed 4096/resetMarker 15. The declared monster source is i16, so it cannot exceed 32767; player source is fixed 213. No fabricated greater-than-32767 KB fixture was added. |
| R15 non-GH draws | No change: full event draws remain gated; reached direct slot draws are owned by NRF-R04b, with eligibility/additional effects under NRF-Q11/HRC-Q06. |

| HARDEST entry | Outcome |
|---|---|
| R06 draw bands | No change: ordered draws, early return and final seeds are explicit in V01/V22. |
| R07 cyclic selector and entire class | No change: V16 states cursor, components, full class and exact writes, including rotation/high-bit discrimination. |
| R10 fast paths and state42-before-state15 | No change: V14/V18 state the fast allowances and the drawn-then-refused case explicitly. |

An additional wording check made R01b's nesting explicit: only within the state-54-absent
branch does GetHit take precedence over SoftHit; state 54 present skips both and KB chance.
This is equivalent wording, not a changed rule. V10 already discriminates both selectors
true and the protected skip.

The three new subcases preserve all 22 vector families and prior numeric results. They require
fresh **full pass 2** under the review stopping rule; full pass 1 does not approve the changed
contract. Changed material is V15g, V19g/V19h and the equivalent R01b wording. User approval
remains pending. No new source campaign, gameplay or implementation test was performed.

After this triage, the targeted specification checker returned exit 0 with **0 errors and
0 warnings**. The cases were independently recomputed as rate 0 with retained speed-setter
execution, accumulator 400 without adjustment, and a fatal null-input predicate failure.

### Fresh-reader full pass 2 and triage

The second fresh full reader examined the actual main file and dependency contracts. It
returned **MISMATCH 0; GUESS 0; GAP 2; DISCRIM 1; GATED 13; COVERAGE 10; HARDEST 3**.
Its GATED count is internally inconsistent: it reported 13 but enumerated 12 unresolved
rule IDs and seven question IDs, which are 19 distinct IDs. Both the reported count and
that discrepancy are retained here; no listed gate or stage limit is removed because of
the counting error.

| Blocking finding | Outcome |
|---|---|
| GAP 1: supplied block/avoidance code absent from input inventory | Added section 3's distinct semantic i32 result code supplied by the SBA-R14 stage, including its low-five-bit use in R01c. |
| GAP 2: local decisions/returns absent from output inventory | Added section 5's guard continue/stop, protected-branch stop, R06 veto/allow, R09a return 0/1, R10a allow/refuse and fatal outcomes; clarified terminal selection and the remaining rewrite code. These inventory entries refer to the existing rules and do not invent terminal effects. |
| DISCRIM 1: V20 could accept an erroneous SoftHit-false requirement | Added V20d with SoftHit true, GetHit false, KB 0, code 1 and R=9: result 16392/code 0 with no draw. |

| Reported coverage area | Outcome |
|---|---|
| Fallback/base conversion | No change: KNB-V32 discriminates the operations owned by KNB-R06a/KNB-R06c. |
| Nondefault-divisor stochastic bands | No change: V01/V22 discriminate the threshold/draw recipes; V05 discriminates their divisor binding. No exhaustive cross-product claim is made. |
| Nonzero synthesis suppression | No change: this selects the same no-synthesis recipe as V16e's guard skip, with the guard and supplied byte explicitly stated in R07a. |
| Monster-start mode 12 | No change: same return-1/no-op recipe as V13b's mode-0 start guard; ordinary dispatch's mode-12 guard is already V11c. |
| Mode-4/nonzero-context decision and roll at least 100 | No change: requested GH/nonzero context is already the local decision boundary. Rand(100) cannot produce a result at least 100, so no fabricated roll fixture is added. |
| End predicate with condition nonzero and total<=0 | Added V19i: present unit, condition 1, total 0, accumulator 900, rate 100 returns true without writes/draws, contrasting V19d's positive-total false result. |
| Negative monster KB base, caps and unsigned GH cap | No change: negative KB shares V15g's <=0 recipe; monster source i16 and fixed player 213 cannot exceed 32767. GH unsigned saturation is already V08a. No exhaustive binding claim is made. |
| Unsigned town wrap | No change: KNB-R03b/KNB-V27 own the wrapped bounds and zero/outside cases consumed by HRC-R02b. |
| Avoidance states 66/68 effects | No change: complete eligibility/start effects remain R04b/HRC-Q06. |
| Zero/negative M | Added V22f: player, total 256, M=0 and seed {2,0} allows with no draw or write. Valid no-minimum-adjustment definitions remain supplied. The case discriminates an invented 4096-raw max-life floor, which would instead draw and veto. |

| HARDEST entry | Outcome |
|---|---|
| R06 draw order | No change: V01/V22 state each threshold, early exit and exact stepped seed. |
| V14 player roll before state 15 | No change: V14e explicitly draws before its state-15 refusal. |
| R04a/V21 priority with KNB transformations | No change: R04a orders the consumed KNB conversions before terminal selection; V21 provides the player/monster priority contrast, and KNB-V32 discriminates conversion internals. |

The changes preserve all 22 vector families and all prior case results. New cases V20d,
V19i and V22f were independently recomputed from the rules. The input/output inventory
additions require fresh **full pass 3**, not a delta review. Full pass 2 is a record for
its reviewed contract and does not approve these edits. All source, gameplay and implementation
boundaries remain unchanged; no new source campaign, gameplay, tests or git actions were run.
After pass-2 triage, the targeted specification checker returned exit 0 with **0 errors and
0 warnings**. User approval remains pending.

### Fresh-reader full pass 3 and triage

The third fresh full reader returned **MISMATCH 0; GUESS 1; GAP 1; DISCRIM 1; GATED 13;
COVERAGE 9; HARDEST 3**. Its 13 GATED entries are grouped: unresolved rules include the
combined R13a/R13b/R13c entry, with question groups including Q05/Q10/Q11. This grouped
count does not replace the full rule/question inventory or remove any stated limit. It is
distinct from full pass 2's recorded count/enumeration inconsistency.

| Finding | Outcome |
|---|---|
| GUESS 1: state-92 name wording could imply absence of a death/timer mechanism | Changed section 3 to name-only evidence: death/timer semantics require separate evidence. No mechanism or absence is inferred from the text name. |
| GAP 1: mode 4 could mean the current rather than requested mode | Added requestedMode as an i32 input distinct from current mode; local player-decision fixtures explicitly supply requestedMode=4. R10a now states requested mode 4 with nonzero context, preserves separate current-mode-0/17 guards, and excludes other requested-mode branches apart from its stated early allowances. V14/V18 current-mode-5 outcomes are unchanged. |
| DISCRIM 1: V12 claimed to reject an invented timer without a timer fixture | Removed that timer-discrimination assertion. V12 discriminates ordinary death dispatch through the state request and unchanged current mode; full dependent effects remain gated. The state name supplies no timer semantics. No timer logger or unverified effect was invented. |

The reader reported nine aggregate coverage entries. Related areas are grouped below for
triage without treating the number of table rows as a revised reader count.

| Reported coverage area | Outcome |
|---|---|
| R01b prior flags | No change: retained-bit behavior is explicit and owned by SBA-R20, with SBA-V26 case (e) already discriminating it. |
| Null/other-kind guards | No change: the relevant stop-without-reaction recipe is in V11a; V13d separately exercises the missing-unit start return. Explicit guard meanings remain in their own rules. |
| Soft/plain effects | No change: complete effects remain R08/HRC-Q06 rather than an invented supplied default. |
| Fallback/base conversion | No change: KNB-V32 discriminates the owned conversion recipes consumed before HRC priority. |
| Multiple avoidance flags | No change: complete avoidance eligibility/effects remain R04b/HRC-Q06; priority/termination is stated separately. |
| Unavailable monster mode 6 | No change: it selects V21d's existing synchronization recipe. |
| Negative max life | No change: same signed-comparison/no-floor recipe as V22f; no separate lower floor or binding is inferred. |
| R10 zero context/requested mode 0 and absent-defender fast paths | No change: early allowances use the existing allow/no-draw recipe in V18a/b/f, while absent defender has the existing explicit initial-refusal recipe. No new result-distinct recipe is required. |
| Concentration greater than 99 | No change: the uncapped value still selects the draw-then-refusal recipe of V14a; it does not require an out-of-range Rand(100) result. |
| Monster-start mode 12 | No change: V13b's mode-0 no-op/return-1 recipe; ordinary mode-12 dispatch is separately V11c. |
| Negative KB base | No change: same <=0 recipe as V15g, including continued speed-setter execution. |

| HARDEST entry | Outcome |
|---|---|
| R06 draw order | No change: V01/V22 retain explicit thresholds, ordered draws, early returns and final seeds. |
| R10 fast paths | No change to behavior: V14/V18 and the newly explicit requested/current-mode input distinction identify the order without supplying extra outcomes. |
| R12/R14 signed/unsigned arithmetic and clamps | No change: wrapped product, u32 reinterpretation, division and zero/cap distinctions are explicit; V04/V08/V15 discriminate them. |

No new vector family or numeric result was added. The requested-mode input clarification
and removed V12 assertion require fresh **full pass 4**, conservatively a full review rather
than a delta check. Full pass 3 remains a record of its reviewed contract. No new source
campaign, gameplay, implementation tests or git actions were performed. User approval
remains pending.

After pass-3 triage, the targeted specification checker returned exit 0 with **0 errors and
0 warnings**. The 22 vector families retain their preceding numeric results.

### Fresh-reader full pass 4 and triage

The fourth fresh full reader accessed the actual main contract and dependency contracts only,
with no reported access limitation. It reported **MISMATCH 0; GUESS 0; GAP 0; DISCRIM 1;
GATED 19; COVERAGE 6; HARDEST 3**, covering 17 verified rules, 22 vector families and 108
cases in that reviewed contract. Its 19 GATED entries are the 12 unresolved rules and seven
question IDs. Retain all those limits and the supplied-stage boundaries.

| Blocking finding | Outcome |
|---|---|
| DISCRIM 1: V05's high total allows regardless of an incorrectly selected divisor | Preserved every existing V05 input/result as case (a). Added case (b), independently testing the same ten whole k values at total 2000, M=25600 and seed {1,0}. Divisor 64 allows without a draw; 32 allows after one mask-4 draw; 8 vetoes without a draw; 16 allows after two draws. Exact final seeds distinguish all four and expose low-nibble-only handling of k=53. |

| Reported coverage area | Outcome |
|---|---|
| Player avoidance flags 256/512 | No change: complete effects remain R04b/HRC-Q06. |
| Monster KB priority over block/GH | No change: KNB-V28/KNB-V19 already discriminate the consumed owner behavior; HRC-V21 preserves the local priority contrast. |
| Special-divisor threshold bands | Added V05b, independently recomputed with thresholds 400/800/1600, 800/1600/3200, 3200/6400/12800 and 1600/3200/6400. Its terminal returns and final seeds resolve this coverage entry as well as DISCRIM 1. |
| No positive component and no critical indication | Added V16g: R=1, k=0, suppression 0, cursor 0 and all components zero writes class/lastHitClass 13 and cursor 1 without a draw; contrasts the existing critical class 29 case. |
| Monster-start mode 12 | No change: shares V13b's existing return-1/no-op start recipe. |
| Requested mode 0/zero-context fast allowances | No change: these use the existing allow/no-draw recipe in V18; the explicit earlier guards and requested/current-mode distinction remain. |

| HARDEST entry | Outcome |
|---|---|
| R06/V22 draw order | V05b now reinforces the already explicit threshold/draw order through four different terminal traces; no algorithm change. |
| R07/V16 cycle | V16g adds the missing noncritical empty-component result without changing rotation, full-class retention or cursor behavior. |
| R04/V21 priority | No change: local priority and the consumed KNB owner transformations remain explicit, with KNB-V28/KNB-V19 discrimination. |

The additions preserve 22 families and every prior case result. Only fixtures and their
discrimination/coverage records changed; rule behavior, ownership and gate boundaries did not.
After this fourth full review with zero mismatches and all findings triaged, the next required
reader is **delta pass 5** over V05's case-(a) presentation, new V05b, new V16g and the findings
they resolve. Full pass 4 does not approve the added fixtures. No new source campaign,
gameplay, implementation tests or git actions were performed. User approval remains pending.

After pass-4 triage, the targeted specification checker returned exit 0 with **0 errors and
0 warnings**. V05b's zero/one/two-draw outcomes and V16g's class 13/cursor 1 result were
independently recomputed before the delta-review handoff.

### Fresh-reader delta pass 5

2026-10-15: a fresh smaller reader checked the actual main contract, numeric foundations,
game-data fields and required main dependency contracts. It read no companion, gap file,
previous report, private source, implementation, runtime or web material. This delta check
followed full pass 4's zero mismatch count and complete triage. Its changed scope was V05's
preserved case-(a) presentation, the ten independent V05b inputs and V16g; no rule behavior,
ownership or gate changed.

Counts: **MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0**. Each list has no unresolved entry.
The reader independently recomputed all added cases: divisor 64 allows with no draw,
32 allows after one mask-4 draw, 8 vetoes without a draw and 16 allows after mask-2 then
mask-4 draws, with the exact stated seeds. The result/seed combinations distinguish all
four mappings and whole-class 53 from its low nibble. V16g writes class/lastHitClass 13 and
cursor 1 without a draw, rejecting an unwarranted critical code. It also confirmed V05a's
original inputs/results remain intact. Existing gates and advisory dispositions remain.
This completes the applicable reader check; it implies neither maintainer approval,
gameplay confirmation nor complete behavior reproduction.

### Specification review (2026-10-15)

An independent review recomputed all 22 vector families, including the stepped seeds, and
found no numeric mismatch. It re-read the original disassembly for D2Game 0xDB910 (divisor
jump table, floor 256, poison test, RandMask(2)/RandMask(4) order, monster availability),
the 0xDCF20 block-cooldown slice (soft return, signed trunc(FBR/8)+15 strict comparison,
mode-9 request then stat-95 write request) and monster block class exceptions 243/333/705,
and D2Common 0x33110's block base 50/100, GH (D+50)*B/100 and KB speed 4096 branches; all
agree with the current rules. Edits added the root-anchor paragraph above, spaced
run-together values and case labels in V06/V08/V09/V11-V17/V20-V22 and R10a/R10b, and
replaced historical rule numbers in the implementation-gap Spec rule column. No vector
input, output, rule, label or gate changed, so no reader rerun is required.
