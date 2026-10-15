# Evidence: Blade Sentinel and Blade Fury

Behavioral contract: [Blade Sentinel and Blade Fury](../../specs/skill-assassin-blade-sentinel-and-fury.md).
Original artifact identities: [source manifest](../source-manifest.md).

## Binary identity and methods

Evidence build: original Lord of Destruction Expansion 1.13c, Windows x86. References below use
`LOD113C-D2GAME` (D2Game.dll, SHA-256 `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`)
and `LOD113C-D2COMMON` (D2Common.dll, SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`).
Their sizes/bases and matching-original correspondence are in the manifest. Descriptions
are analyst labels, not original exports. Module RVAs remain usable with matching owned binaries.

Reused compatible original-code readings date 2026-10-11 (earliest reading date unrecorded)
and 2026-10-12. New targeted disassembly and original-image decoding date 2026-10-15.
For new readings, availability of reconstructions checked against original bytes was checked before assembly: none of
the selected D2Game cast/follow/roller/sidecall targets had a reconstruction checked against original bytes. D2Common's
release-time getter/setter do have matching reconstructions and were read first. The full
688-byte damage-fill image was fingerprinted and decoded, including its previously encoded
physical/fire/magic slice. Reading that image resolves those instructions; it does not
upgrade unrelated MHD branches, collection profiles or runtime coverage.

## Dispatch and function records

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | 0xA19A0 | none | action dispatcher | R01/R02 | Reused 2026-10-12 plus SCD 2026-10-15 record: prepared srvdofunc selection and callback; final caller semantics consume SCD-R06/R12 |
| `LOD113C-D2GAME` | 0x96310 | none | Sentinel action 44 | R01/R10a | Original instructions 2026-10-11/12; full targeted reread 2026-10-15 through return 1 and both failure epilogues |
| `LOD113C-D2GAME` | 0x9EE50 | none | summon identity resolver wrapper | R01/Q01 | New 2026-10-15: calls Common ordinal selection and validates monster count; fallback monster data read. Supplied resolved value boundary; complete selector integration open |
| `LOD113C-D2GAME` | 0x95D50 | none | summon creation request | R01/R10/Q08 | Reused seed chain and 2026-10-15 reread: selection, room/location gates, creator 0xA3610 then summon setup/mode; not a full creation budget |
| `LOD113C-D2GAME` | 0xE1D90; 0xE0110 | none | generic unit creator; seeder | R10a/R11 | Reused 2026-10-12 game-seed call sites and current NRF-R03a/NRF-R17a: reached non-player seeding only |
| `LOD113C-D2GAME` | 0xE19C0 | none | placement/relocation-shaped side operation | R01/Q09 | New complete instruction-form read 2026-10-15; imports/path/room/collision/update effects remain integration gates; no whole-helper no-draw claim |
| `LOD113C-D2GAME` | 0x21290 | none | route command copy | R01/Q01 | New 2026-10-15 read: copies caller command words and sets monster command association; allocation/storage integration remains open |
| `LOD113C-D2GAME` | 0x8E6F0; 0x8EA80 | none | event unlink; scheduler | R01/R04/R10b | Compatible prior record and STL-R01: removes reached type 2 events, requested frame and actual clamp distinct; state 54/type 2 branch can draw |
| `LOD113C-D2GAME` | 0x95560 | none | Fury action 48 | R02/R06a/R11 | New full 2026-10-15 reread; creation return ignored, timer and result 1 assigned inside open gate, state 12 refresh reached from either branch |
| `LOD113C-D2GAME` | 0x91340; 0x91240 | none | selected missile; progression result | R02/Q03 | New 2026-10-15 instruction-form readings: prepared Skills selections, progression-family state path, expression requests; encoded progression bytes not independently resolved here, so P supplied |
| `LOD113C-D2GAME` | 0x95BC0 | none | start-table 26 candidate | R06b/Q02 | New original-image slice from entry through return at 0x95D43; all local branches located; imported lifecycle, mana and animation effects not fully read |
| `LOD113C-D2GAME` | 0xE2480 | none | animation event refresh-shaped side operation | R02/Q09 | New instruction-form reading 2026-10-15: removes event 0/1 and schedules animation event requests; encoded loop and imports unresolved, so effects/draw budget gated |
| `LOD113C-D2GAME` | 0x6F930 | none | missile creator | R11/R07 | Compatible 2026-10-12 reached seeder and current MCI-R10/R24; failures before seeding and downstream creation draws remain external |
| `LOD113C-D2GAME` | 0x40860 | none | owner-follow missile action 20 | R03/R05a/R09a | Full original disassembly 2026-10-15: owner condition, lookup, unavailable result; sync/frame 10/collision then return 1 or end then return 2 |
| `LOD113C-D2GAME` | 0x3A1D0; 0xDFA90 | none | owner-position sync request; unavailable decision | R03 | New 2026-10-15: room/position/path request and explicit kind/mode/completion tests; full relocation integration gated |
| `LOD113C-D2GAME` | 0xE03A0 | none | owner unit lookup | R03/Q04 | Encoded table/hash lookup candidate read 2026-10-15; supplied owner result, not a newly verified real lookup chain |
| `LOD113C-D2GAME` | 0x3F320 | none | plain missile step (collision pass requested by action 20) | R03 | Review 2026-10-15: 0x40860 calls it after the counter write with the game on the stack and ignores its result; interior consumed through MMC-R05 (MMC companion record), not reread here |
| `LOD113C-D2GAME` | 0x3EDE0 | none | contact/handoff; also R03's end request | R03/R04/R12 | Compatible MMC/MHD contact records; new 2026-10-15 full instruction-form read locates early exits and arming calls at miss/success branches. Review 2026-10-15: 0x40860's end path calls this function with missile in EDI, stack(game, target 0, force 1) |
| `LOD113C-D2GAME` | 0x3E590 | none | contact cooldown arming | R04 | New 2026-10-15 complete local sequence: allocation failure test, event 12, list state/callback, attach, active state request; imports by named STL contracts |
| `LOD113C-D2GAME` | 0x3A4E0 | none | damage fill base-roll slice | R12c/R08/Q06 | Full 688-byte original-image decode 2026-10-15; zeroing, physical, fire and magic-min argument corroborated at 0x3A4FD..0x3A55D; seven-slot slice through 0x3A622; later target/critical adjustments retain MHD gates |
| `LOD113C-D2GAME` | 0x39B40 | none | per-element roller | R12a/R12b/R08 | New complete 2026-10-15 read: max guard, min guard, initial swap, percentage, wrapped difference, missile Rand and add |
| `LOD113C-D2GAME` | 0x14D0; 0x11D0 | none | mulDiv; guarded Rand | R12 | Compatible NRF-R05/NRF-R02: pct first operand, endpoint second, 100 divisor; signed positive modulus only steps |
| `LOD113C-D2GAME` | 0xA6B70 | none | candidate neighboring AI operation | Q01 | New 2026-10-15 table-neighbor lead plus full instruction read: `calc4`-derived frame condition, one-time missile stage and stored-command coordinates. AI identity/root and full callees remain unproved |

### Original dispatch data

At D2Game RVA 0x1074A8, action table slot 44 contains preferred VA 0x6FCB6310 and
slot 48 contains 0x6FCB5560. At RVA 0x107338, start table slot 26 contains 0x6FCB5BC0.
At RVA 0x10D970, missile do table slot 20 contains 0x6FC60860. Original image cells read
2026-10-15; SCD's dispatcher bound is 191 entries for ordinary do selection, not a blanket
catalog of all skill execution. ItemEffect/forced handling consumes SCD-R06 unchanged.
The hit/damage function tables remain at RVA 0x10DA48/0x10DB68; no unrelated slot binding
is inferred from their locations.

The neighboring AI pointer array contains 0x6FCC7DC0 at RVA 0x10EDE0 and 0x6FCC6B70 at
0x10EDF0. The former is the independently investigated sentry candidate; this positional
relationship is a lead for Blade. The 2026-10-15 attempt did not establish the index consumer,
symbolic AI-name compiler or numeric 102 binding. No absence or complete six-rung search is claimed.

### Imported and internal mappings

| Semantic operation | Original D2Common ordinal | Target RVA | Support / remaining limit |
|---|---|---|---|
| layer-0 unit stat read |10973|0x38B70|NRF-R12b/R12c existing-stat contract; does not establish source contribution production |
| replacing unit stat request |10887|0x3A740|Sentinel level request; NRF-Q20/STL-R06c propagation gated |
| skill release-time store |10243|0x4DF60|Reconstruction checked against original bytes: non-null node stores supplied i32 at node+0x18; null fatal assertion |
| skill release-time read |10980|0x4E030|Reconstruction checked against original bytes: non-null node returns signed node+0x18; null fatal assertion |
| find owned skill node |10630|0x4F070|SCD-R06; supplied node result |
| remaining-lifetime counter setter |10970|0x6A6D0|New 2026-10-15 full disassembly: missile kind 3 with data stores signed i16 after clamp[-32768, 32767]; absent/wrong-kind/data returns without write; action 20 supplies 10 |
| active-state read |10494|0x33CD0|NRF-R12a, not associated-list lookup |
| caster room read (R01 placement) |10331|0x2FE10|Review 2026-10-15 thunk resolution; room-getter semantics reuse the KNB/HRC companion records |
| cooldown list allocation (R04) |11013|0x38100|Review 2026-10-15 thunk resolution; STL-R06a owns allocation |
| list state tag (R04) |10785|0x38060|Review 2026-10-15 thunk resolution; STL list-field contract; called with state 86 |
| list expiry callback (R04) |10911|0x37D70|Review 2026-10-15 thunk resolution; STL list-field contract; called with D2Game 0x38E60 |
| list attach (R04) |10807|0x3A7D0|Review 2026-10-15 thunk resolution; STL-R06b (gated); called with full=1 |
| state-on request (R04) |10945|0x34510|Review 2026-10-15 thunk resolution; NRF-R12a/STL-R07 routing; called with state 86, value 1 |

The original local call thunks are D2Game 0xA42A/0xA472/0xAC94/0xAD12/0xA3F4/0xAC7C/0xA68E
respectively for the first seven rows, then 0xA478/0xA6A0/0xA5B6/0xA75A/0xA60A/0xA3CA.
Thunk resolution names ordinals only; analyst labels attached to those targets in older
corpora are unrelated and are not relied on. Getter/setter descriptions in older reconstruction labels refer to unrelated
items; the actual read/store establishes the release-time semantics. These imported targets
belong to the manifest's D2Common artifact, not D2Game.

| Semantic field / call | Original layout or calling evidence | Rule |
|---|---|---|
| caster caster marker | Unit+0xC4 value 0x40 set/cleared | R01/R02 |
| summon summon marker | Unit+0xC4 value 0x08 cleared on returned summon | R01 |
| owner-reference condition / reference | Unit+0xC8 bit 10; owner kind/id at+0x94/+0x98 | R03 |
| frame / game seed | Game+0xA8; +0xD0/+0xD4 | R01/R02/R04/R10/R11 |
| child seed / initialLo | Unit+0x20/+0x24; +0x28 | R10a/R11 |
| placement request | 0xE19C0: ECX=casterX, EAX=casterY, EDX=casterRoom, stack(game, summon, 0, 0); caller ignores result | R01/Q09 |
| animation refresh request | 0xE2480: ESI=caster, stack(game, 1); older two-argument description omitted register unit | R02/Q09 |
| absolute-target missile request | 0x95560 zeroes 0x17 dwords, flags 0x20, owner/source caster, missile+0x10, target+0x1C/+0x20, skill/level+0x2C/+0x30 | R02 |
| missile remaining-lifetime counter | Common#10970/RVA0x6A6D0 stores signed word at MissileData+0x10; value 10 after synchronization, before collision | R03 |
| NextHit / NextDelay | prepared Missiles+0x188/+0x189, u8 | R04 |
| cooldown list | type 2, owner target kind/id, expiry frame+u8 delay, state 86, callback D2Game 0x38E60; attach full=1 | R04 |
| base record | zero 0x70 bytes; physical+0x08, fire+0x10, magic+0x20, lightning+0x1C, cold+0x24, poison+0x28, burn+0x14 | R12c |
| per-element arguments | roller EAX=maxStat, ECX=pctStat, stack(missile, minStat); mulDiv EAX=pct, EDX=endpoint, ECX=100 | R12a/R12b |

## Corrections and evidence applicability

The historical contract and its full internal provenance have been preserved privately.
This public record retains every material correction and usable original-artifact locator.

| Affected claims | Earlier discrepancy / evidence | Current resolution and limits |
|---|---|---|
| R01/R02 partial failure | Earlier prose called target failures no-effect despite preceding marker writes |2026-10-15 full cast readings show flag write precedes lookup; V03(f)/V09(b, c) retain it |
| R01/R02 timer/flags | Initial pre 2026-10-12 reading said Fury timer written regardless of gate and marker cleared on caster rather than summon | Historical 2026-10-12 correction:0x95560 jge skips creation/timer; 0x96310 clears returned summon bit 3. Retained |
| R02/R06a | Earlier 'return 1 if released' and only-release wording could imply successful creation |2026-10-15 full 0x95560 reading: creator result never tested; timer store and result 1 follow an attempt. V03's failed creator is discriminating |
| R10a/R11 | Earlier one game draw per cast / blocked cast draws nothing | Reached seeder still exactly one game step, child initialized not drawn. MCI failures may bypass it; downstream operations and state 12 refresh can draw. Removed unsupported whole-cast budget, preserving exact seed values |
| R03/R05a/R09a | Earlier owner exists shorthand omitted reference condition and unavailable kinds/modes; labels asm-only | New 0x40860 and 0xDFA90 reads establish reached decision; sync/collision/end imported effects remain supplied. Return 1/2 ignore dependency result |
| R04 | Earlier every contact arms / fixed delay implied full expiry | Original contact 0x3EDE0 early state/type/relation exits precede arming.0x3E590 arms miss and reached handoff only, tests allocation, schedules before tags/attach/state. STL stores due/past event next frame while list retains requested expiry. Actual attach/free remain gates |
| R07/R08 | Weapon ownership transfer and global no-change-in-flight inferred from labels and missing reads | MCI-R24/MDP own building/storage. New full image shows the base roller reads missile stats; later 0x3A4E0 code reads TARGET-dependent facts, not proof of a universal owner/equipment snapshot. Retain unchanged-supplied-stat result only; integrated capture/procs open |
| R12a/R12b | Earlier adjusted result always within min..max-1 | Signed guards apply before percentage; low32 changes can reverse/wrap endpoints. No post-percentage swap. V10(c) demonstrates pct-200 =>-5/-9, return-5, no draw |
| R12c/Q10 | Earlier physical/fire/magic-min argument unresolved because encoded bytes were unchecked |2026-10-15 matching original image decoded across all 688 bytes.0x3A4FD zeroing, 0x3A4FF max 22 read, 0x3A513 min 21, 0x3A531 Rand, 0x3A53A physical store, 0x3A548 fire helper and 0x3A54D magic min 52 corroborate full base slice. R12c becomes verified static; complete MHD fill/integration and observations remain open |
| R12 identities | Older poison/burn/mastery names absent from GDF's selected list | Existing compatible loader record names 57/58 poisonmindam/poisonmaxdam, 316/317 burningmin/burningmax, 332 passive_pois_mastery; TD-BLADE-STATS below. Fixed code ids unchanged |
| Q01/Q02/Q03/Q04 | Older not-found labels had only label searches or no loaded table | New selected original table slots, loader cells and bounded candidate reads add evidence; absent six-rung record is unresolved, not absence. Text records do not prove compiled runtime association |
| V08/V10/V11 | Prior examples left random final outputs/seeds symbolic | Preserved original ordinary inputs and numeric results; supplied entry seed 1/0 completes V10. V08 swapped 12/8 case fully computes both physical then inherited fire; V11 retains original lightning/swapped/equal cases and adds a seven-slot discrimination case |

The 2026-10-12 start-table correction remains: preferred VA 0x6FD27338 rather than the
earlier 0x6FD28338. ItemEffect only replaces an already bounds-accepted original do selector
under both nonzero flags, per SCD-R06. Old function labels were never original names.
Earlier blanket reconstruction-percentage statements are not evidence of today's availability.

## Named table data

### Source profile and coverage

Date 2026-10-15; effective paths below read through the running game's Storm loader.
The source member reports vanilla LoD launch key `lod-113c`; the recorded member inventory
distinguishes it from PD 2 profiles. The current version-console refresh failed. Earlier
recorded override inventory reported file redirection disabled, but no complete current
override inventory or pristine archive claim is made. Physical supplying MPQs and compiled
runtime correspondence are uncaptured. A successful text read is data evidence, not gameplay.

Skills reuses the same day's compatible complete loader record in the
[sentry evidence](skill-assassin-sentry-firing-and-range.md#td-skills17): requests
(offset, length)=(0, 60000),(60000, 60000),(120000, 53535), all returned exactly requested,
total 173535 bytes, SHA-256 `465c28e83657540e0d7ae5b0de826d5657be7fba1aa539a85c3ca3aeea34b498`.
New Missiles requests were(0, 60000),(60000, 60000),(120000, 60000),(180000, 12979), all exact,
total 192979 bytes, SHA-256 `89bee1fdeeb520d6eb3920437d237ba5a70bdf045770ed9808ca1a611fd932f0`.
New MonStats requests were offsets 0, 60000, 120000, 180000, 240000, 300000, 360000 with 60000
bytes each, then 420000/12634, all exact, total 432634 bytes, SHA-256
`a15e1cb876bd9af1e8f54199c9619226cf4835c45c799aed2d432901487ecb7c`.
Complete bodies were joined before parsing by header names; no preview-only or split-row
absence inference was used. Narrow cells only are published.

### TD-BLADE-SKILLS

Effective virtual file `data/global/excel/skills.txt`; explicit Id identity agrees with
zero-based data-row index excluding header for 257/266. Supports R01/R02/R06b and Q02/Q03/Q04.

| Row | Columns | Raw cells / interpretation |
|---|---|---|
| Blade Sentinel /Id 257 |srvstfunc; srvdofunc; summon; srvmissilea|blank; 44; bladecreeper; blade creeper |
| Blade Sentinel /Id 257 |`calc4`;`Param1`;`Param2`;`Param3`;`Param4`|`ln12`; 100; 12; 5; 3. Expression and parameters only; no independently verified lifetime formula |
| Blade Fury /Id 266 |srvstfunc; srvdofunc; srvmissilea|26; 48;`bladefragment1` |
| Blade Fury /Id 266 |`prgcalc1`;`Param3`;`Param4`|`par4`; 3; 5. With faithful prepared selection/evaluation, P5 would give duration 4; not a `Param4`-as-duration rule |
| Blade Fury /Id 266 |startmana; mana; lvlmana; manashift; usemanaondo|6; 8; 1; 5; 1. No completed mana deduction formula inferred |
| Blade Sentinel /Id 257; Blade Fury /Id 266 | srvmissile | blank; blank. Read 2026-10-15 from a local copy whose SHA-256 equals the loader record above (465c28e8...), same header-based column selection and row convention. The compiled selector these blanks produce is not read (Q04) |

### TD-BLADE-MISSILES

Effective virtual file `data/global/excel/missiles.txt`; named row identities with explicit Id.
Supports R03/R04's named source values and Q04. The Skill column identifies a damage/data
association, not by itself the actual creator selection.

| Row | Columns | Raw cells / interpretation |
|---|---|---|
| blade creeper /Id 392 |pSrvDoFunc; pSrvHitFunc; NextHit; NextDelay; Skill|20; 37; 1; 25; Blade Sentinel |
| blade creeper /Id 392 |LastCollide; CollideType; ToHit; AlwaysExplode; CollideKill|blank; 3; 1; blank; blank. Read 2026-10-15 from a local copy whose SHA-256 equals the loader record above (89bee1fd...). Supports V01's LastCollide-disabled fixture only as a text cell; compiled values Q04 |
| `bladefragment1` /Id 506 |pSrvDoFunc; Skill|1; Blade Fury |
| `bladefury1` /Id 505 |pSrvDoFunc; pSrvHitFunc; Skill|1; 52; Blade Fury. Distinct named row; not silently substituted for selected fragment |

### TD-BLADE-MONSTATS

Effective virtual file `data/global/excel/monstats.txt`; row Id `bladecreeper`, explicit
hcIdx 413, physical zero-based data-row position 414 excluding header because separator rows
affect position. Raw AI=`BladeCreeper`,`Skill1`=`Blade Sentinel`; numeric AI 102 does not follow
from the symbolic text name. Supports Q01's identified table lead only.

### TD-BLADE-STATS

Reused 2026-10-15 compatible complete `data/global/excel/itemstatcost.txt` loader record from
[game-data evidence](core-game-data-fields.md).
The complete single request(offset 0, length 40109) returned 40109/40109 bytes; decoded effective-file SHA-256 `ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`. Ids use zero-based data rows excluding header, as that record establishes. Selected Stat
cells:57=`poisonmindam`, 58=`poisonmaxdam`, 316=`burningmin`, 317=`burningmax`, 332=`passive_pois_mastery`.
These are verified names. R12's fixed code selectors establish their use independently;
the selected GDF table omits some of these names, which is not evidence of absence.

## Gap assessment and search limits (2026-10-15)

| Question / evidence gap | Compatible record and new bounded investigation | Result |
|---|---|---|
| Q01 route/lifetime/AI 102 | Preserved 2026-10-11/12 label-only attempt; new MonStats named record, original adjacent AI pointer cells, full 0xA6B70 candidate and 0x21290 command-copy reading | Candidate uses stored command points and `calc4`-derived expiry, but index consumer/root and full callees unresolved; no current-position or lifetime formula promoted |
| Q02 start/mana | New Skills start 26 cell, original start-table 26 pointer and original 0x95BC0 body through return | Locates state 12-list and timer branches and conditional 0x95560/0x9F850 calls; full lifecycle/mana/repeat semantics unresolved |
| Q03 progression vs `Param4` | New `prgcalc1`=`par4`,`Param4`=5 cells; 0x91240/0x91340 readings and SCX expression context | Text relation resolved; compiled selection/state/evaluation correspondence open; supplied P remains verified consumer input |
| Q04 delay/do mapping | New complete Missiles named cells and original missile slot 20; full follow/contact/arm local reread; current STL-R01/R06a/R06b/R05 compatibility assessment | Named delay 25 and action 20 cells closed. Actual compiled selection, collision/path and list attach/free/expiry still integration gaps |
| Q05 exact snapshot/owner bonuses | Current MCI-R24/MDP parameter owner records assessed; full 0x3A4E0 original image and complete element helper read | Label-based transfer inference removed. Base stored-stat stage verified; exact real capture/mutable stats/creator draws open |
| Q06 procs/removal/interruption | Full fill image establishes downstream target/critical stages; current MHD/HRC boundaries assessed; start candidate reaches timed state and mana call | Complete post-fill/proc/interruption effects remain in named dependency paths; no universal cancellation or unchanged damage conclusion |
| Q07 other builds | Evidence policy restricted to original 1.13c; no other build inspected | Outside boundary; no equivalence claim |
| Q08 initialization draws | Reused seeder chain; new spawner read plus MSD-R01..R19 and HRC mode context | Source step retained exactly; class/data/path-specific extra count requires instantiated creation inputs |
| Q09 sidecall roles/draws/flags | New complete instruction-form E19C0/E2480 reads; identified passed caster/summon/room/point and state 12 path; encoded animation loop/import gates assessed | Exact local request and retained marker behavior clarified; complete side-operation effects/draw counts open |
| Q10 encoded physical/fire/magic | New complete matching original 688-byte image decode and resolved stat/Rand/helper imports | Local base-roll uncertainty closed; R12c verified. Q10 now tracks pending gameplay confirmation, while full fill stays Q06 |
| All missing executable coverage | Preserved 11 original families; supplied inputs, exact outputs, seeds, draw order and discriminating failure/negative-percent/seven-slot cases added | Fixture-definition gaps closed; supplied operations do not promote actual integration |
| O01/O02/O03 | Preserved all original proposed observations; compatible prior record contains no actual captures; no gameplay campaign performed | All remain pending with bounded procedures below |

### Arithmetic record

V06/V07's exact first source step remains 0x10*0x6AC690C5+3=0x6_AC690C53,
game={0xAC690C53, 6}, child={0xAC690C53, 666}. This is a reached seeder record.
V08(a)'s fire draw from{1, 0} yields low 0x6AC690C5, high 0, result modulo 4=1, fire 6.
V08(b) first draws physical 9; second seed is{0x669A3799, 0x2C8909E7}, again modulo 4=1, fire 6.
V10(50) uses adjusted 7/13, modulus 6, result 3, return 10; V10(-30) uses 4/7,
modulus 3, result 0, return 4. Both finish{0x6AC690C5, 0}. V10(-200) has adjusted-5/-9,
signed modulus-4, return-5 with seed unchanged. V10(d) pct 70001 and endpoints 200/300 select divide-first, adjusted 140200/210300, modulus 70100, result 62685, return 202885. V10(e) endpoint maximum 2147483647 with pct 100 uses divide-first 2147483600 and wraps the adjusted maximum to-49; adjusted minimum 2 yields modulus-51, return 2, no draw. These are static hand computations, not captures.
The contract's complete seven-slot table records V11(e)'s ordered draws/final seeds.

## Observation procedures

No runtime record exists. The three recipes remain pending. Use an owned original 1.13c
installation matching the manifest, establish its effective data/overrides separately, and
record exact class, level, skill rank, synergies, equipment, difficulty, positions, frames and seed
pairs. A practical common setup is an Assassin level 30 with chosen ranks/prerequisites,
sufficient mana, Normal difficulty, and a stable non-town stationary target; level 30 is a
setup choice, not a behavior constant. A source/text table read never counts as gameplay.

**BSF-O01.** Stand so a Sentinel route crosses one stationary eligible monster outbound and
return; retain another comparable target off the route as a control. Record actual creature,
missile class and prepared NextHit/NextDelay, contact entry/filter outcomes, to-hit result,
list allocation/expiry, active state 86 and event 12 request/stored/dispatch frames. First
record a successful contact, then a controlled miss. Count contacts before/after each expiry
and on both legs. Expected verified boundary: blocked membership skips to-hit/arming;
reached miss arms before hitbymissile, using the same target state as a hit. Actual list
expiry and leg timing require Q01/Q04; a visual crossing alone is not a recorded contact.

**BSF-O02.** Hold Fury on a stable target from a known start frame with constant equipment
and sufficient mana; record start/action callback entries, progression result, release-time
values, creator results, marker, state 12, refresh requests and mana before/after each operation.
Capture equality and the first later frame explicitly. Expected local behavior:timer<frame,
attempt writes frame+duration and return 1 irrespective of creator result; blocked action
keeps timer but can request refresh. Do not predict held-input frequency, mana amount or
P from a text field until Q02/Q03 correspondence is established. Record external draws
separately from the reached seeding step.

**BSF-O03.** Choose two weapons with sharply different documented damage and no other
changes. Launch a missile with A, swap to B before contact, and separately record creation
parameter stores, missile stats at contact, missile seed, base-roll outputs and final target
damage. Repeat with A retained as a control and record target bonuses/resistances. Expected
verified boundary:identical supplied missile stats and seed give identical base rolls.
Whether real stored values or final bonuses remain identical is Q05/Q06; do not call an
inconclusive equipment trial confirmation of a universal launch snapshot.

## Migration and review record

2026-10-15: current-template behavioral regeneration preserves the file boundary, 12 parent
rule families, 11 executable-vector families, three observation recipes and the 10 historical
question topics. Repository search found no plan outside gaps, implementation code or test
reference to the spec filename/prefix; ids are unconsumed. Practical rule/vector ids were
retained without old-number aliases. Q10's static instruction uncertainty is settled above;
its historical topic/result remains here, with pending observation in the current contract.

Labels changed with evidence: R03/R05a/R09a now verified at the reached supplied-operation
decision boundary; R12c verified at the corroborated base-roll slice. R08 now verifies only
unchanged supplied stored inputs. R05c's unsupported absence label becomes unresolved;
R07/R10b's incomplete capture/budget claims become unresolved. Existing verified requests
and arithmetic retain their status within explicit integration gates. The document remains
draft. No gameplay observation, no implementation change and no cross-build claim.

All original private paths, source commits, reconstruction names, fixture encodings and dated
investigation detail were preserved in a separately hashed private audit. Public source
references above stand independently of it. Historical reviews do not review this contract.
Mechanical lint and self-review are recorded after the final edit pass. The root
workflow supplies the fresh independent reader; none is claimed complete here.

Final writer check 2026-10-15: targeted `check_spec.py` returned 0 errors/0 warnings. Self-review checked source/dependency boundaries, marker failure retention, creator-result handling, seed ownership, all 11 vector families/three recipes, exact ordinary values and seven-slot arithmetic, public/private separation, and unchanged historical implementation findings. No automated implementation test or gameplay action was run. This is a writer handoff, not independent-reader completion.


### Independent reader FULL1 (2026-10-15)

The reader recomputed all 11 executable-vector families, including the seven-roll V11(e)
case; no explicit total case count was reported. Counts: MISMATCH 0, GUESS 0, GAP 1,
DISCRIM 2, GATED 21, COVERAGE 6, HARDEST 3. This pass reviews its input snapshot;
fixture/input edits below require a fresh full pass and do not claim that review complete.

| Finding | Disposition |
|---|---|
| GAP: V04 external caster weapon A/B and ranges not declared in inputs/dependencies | Spec edit: section 3 declares external caster equipment context as a fixture-only input, explicitly outside R12's consumed arguments; Q05 retains real stat propagation. Original exact base-roll assertions remain unchanged |
| DISCRIM: V08(b) physical/fire both modulus 4 and result 1 allow reversed order | No new fixture: V08's discrimination text now names this limitation and cites the owning R12c BSF-V11(e), whose distinct moduli 4/5/6/7/8/9/10, component tuple and intermediate seeds reject the altered order. V08 retains its swap/arithmetic checks without claiming independent order discrimination |
| DISCRIM: V05(a) placement result unspecified; abort-on-failure could pass | Spec edit: supply placement failure 0/no fixture effects and require subsequent command/event/marker operations and final action return 1. This exercises the original0x96310 ignored result; no placement effect is promoted |
| COVERAGE: R01 invalid/absent skill row | No extra case: R01 states both guards before the marker; the unchanged-zero-return/no-marker/no-target-or-spawn state recipe is represented by V09(d). Remaining difference is earlier resolver-call reach, not another stored-output recipe; do not add a pure query-count fixture |
| COVERAGE: R02 missing skill node/negative selected missile | No extra case: R02 explicitly orders the two guards before progression/marker/target. V03(e) already represents the same retained-marker/unchanged-timer/return 0/no-target-or-creator effect recipe. Additional cases would distinguish earlier lookup reach only; real lookup production remains gated |
| COVERAGE: R03 false owner-reference/completion/unavailable kinds/modes beyond absent owner and monster mode 12 | No extra case: every listed condition selects the same end-request/return 2/no-sync-or-counter-write recipe already discriminated by V05(d). R03 gives each exact predicate; branch enumeration would add no different output recipe |
| COVERAGE: R04 absent target and remaining ordered filter failures | No extra case: absent-target arming has the no-allocation/event/attach/state-write recipe of V02(f); every remaining early filter failure shares V02(c)'s no-to-hit/no-arming result. MMC-R11 owns filter order and concrete predicates; no whole-contact zero-effects claim is added |
| COVERAGE: R12a nonpositive minimum with positive maximum | Spec edit: V11(f) supplies min=0/max=9/mastery 50 and requires zero lightning/no step before percentage. It rejects an implementation checking only the maximum guard while preserving all original cases |
| COVERAGE: R12b percentage identity-1 skips read | No extra case: R12b explicitly skips it; with the owning NRF stat-reader invalid-identity semantics, inventing a nonzero supplied value at identity-1 would not be a compatible fixture. Its numeric no-adjustment recipe is already exercised with percentage 0. A pure read-count assertion is unnecessary |
| HARDEST: V10/R12b operand order, mulDiv, truncation/wrap | No change: R12b states exact ordered NRF-R05 arguments and V10(a..e) discriminates negative truncation, post-adjustment order, operand order and large-operand wrapping |
| HARDEST: V11(e)/R12c seven-component seed order | No change: explicit seven-pair input, ordered draw/component/seed table and final tuple provide each step; R12c owns that order and excludes later fill stages |
| HARDEST: V03/R02 marker, target, timer gate and final dispatcher distinction | No change: R02/V03 provide action-callback effects/return; section 2 and SCD-R06/R12 explicitly distinguish the dispatcher's final result. Failed creation and blocked-state12 cases already discriminate the local decisions |

Every GATED entry is retained with no promotion:

| Finding | Disposition |
|---|---|
| GATED: BSF-Q01 | Retain creature AI/route/lifetime boundary |
| GATED: BSF-Q02 | Retain Fury start/mana/repetition boundary |
| GATED: BSF-Q03 | Retain compiled progression correspondence; supplied P available |
| GATED: BSF-Q04 | Retain actual missile/contact/list integration; local requests available |
| GATED: BSF-Q05 | Retain actual parameter/stat/equipment propagation and creator budget |
| GATED: BSF-Q06 | Retain final-hit/proc/interruption effects |
| GATED: BSF-Q07 | Retain other-build outside-boundary question |
| GATED: BSF-Q08 | Retain exact summon initialization draw budget |
| GATED: BSF-Q09 | Retain side-operation effects/draws and marker meanings |
| GATED: BSF-Q10 | Retain pending gameplay confirmation |
| GATED: BSF-V01 | Retain supplied contact/list/event fixture boundary |
| GATED: BSF-V02 | Retain supplied hit/list/filter fixture boundary |
| GATED: BSF-V03 | Retain supplied action dependencies and action-callback result |
| GATED: BSF-V04 | Retain unchanged supplied missile stats and external equipment context only |
| GATED: BSF-V05 | Retain cast/follow supplied-operation boundary, including placement failure and wrapped scheduling limit |
| GATED: BSF-V06 | Retain isolated reached seeder, not whole creation budget |
| GATED: BSF-V07 | Retain supplied creator with exactly stated seeder behavior |
| GATED: BSF-V08 | Retain supplied base-fill inputs, not final hit |
| GATED: BSF-V09 | Retain supplied cast storage/placement results |
| GATED: BSF-V10 | Retain isolated roller arithmetic and supplied stored reads |
| GATED: BSF-V11 | Retain supplied seven-slot base-roll boundary, not downstream fill |

All 1 GAP, 2 DISCRIM, 6 COVERAGE, 3 HARDEST and 21 GATED entries have the dispositions
above. MISMATCH/GUESS had no entries. No new behavior gate or evidence-label change was
introduced by this review triage. Fresh FULL2 remains pending.

Root evidence-link review 2026-10-15: corrected the SFR shared-loader anchor to `#td-skills17`; compatible current base-fill evidence remains reusable for MHD/MDP at the exact R12c slice above, without promoting their later target/critical/parameter stages or repeating unchanged original-image reads.

### Independent reader FULL2 (2026-10-15)

A fresh second full reader recomputed all 11 vector families and 38 lettered/base cases, including V11(e)'s seven ordered rolls; V05(d) also states an independent unavailable-owner alternative. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 21, COVERAGE 5, HARDEST 3. The complete report is retained privately. It confirmed the declared inputs/output inventory, failed-placement continuation and existing V11(e) order discriminator. Outcomes:

| Finding | Disposition |
|---|---|
| MISMATCH 0 / GUESS 0 / GAP 0 / DISCRIM 0 | No blocking finding; no contract edit required |
| GATED 21: Q01..Q10 and V01..V11 | Retain every documented integration, outside-build, observation and supplied-stage limit; the grouped count closes none |
| COVERAGE: R01 invalid/absent row | No additional case: the existing V09(d) recipe gives early return 0 without marker, target or spawn effects. Exact invalid/absent-row bindings remain unexercised; no new result recipe is claimed |
| COVERAGE: R02 missing node, negative selected missile, wrapped duration | No additional case: V03(e) represents the retained-marker/unchanged-timer early-return recipe; V03(a/d) represents the positive-duration attempt recipe. Low32 duration subtraction consumes the section 3/NRF integer contract. Extreme wrapped-duration and earlier guard bindings remain unexercised; this is not exhaustive input coverage |
| COVERAGE: R03 owner-reference false, completion, other unavailable kinds/modes | No additional case: all select the same end-request/return 2/no-sync-or-counter-write recipe as V05(d). The precise predicates stay explicit; no unstated live-owner default is introduced |
| COVERAGE: R04 absent target and remaining contact-filter failures | No additional case: R04 arming with absent target has the no-list/event/membership recipe of V02(f); early contact rejection uses V02(c). MMC owns the concrete ordered filters. These do not claim an absent-target whole-contact callback sequence |
| COVERAGE: R12b percentage identity -1 | No additional case: exact identity guard remains explicit and has the existing numeric no-adjustment recipe; compatible NRF invalid-stat reads do not supply an invented nonzero value for -1. Extra pure-query counts are not acceptance |
| HARDEST: V10/R12b operands, mulDiv, truncation, wrap | No change: exact operand order and V10(a..e) provide the branch/rounding/wrap arithmetic, with root independent recomputation agreeing |
| HARDEST: V11(e)/R12c seven ordered rolls | No change: distinct moduli, component values, intermediate seeds and final tuple discriminate caller order |
| HARDEST: V03/R02 marker, target, timer, creation and dispatcher result | No change: the failed creator, equality and refresh cases assert action effects/return; SCD owns the distinct final dispatcher result |

All findings have dispositions. The final applicable blind-reader condition is met; no rule, input, output, vector, ownership or gate changed after FULL2. The document remains draft with its existing gaps and pending observations. This record adds no gameplay confirmation, implementation test, implementation plan or maintainer approval.

### Specification review (2026-10-15)

An independent reviewer reread D2Game 0x96310, 0x95560, 0x40860, 0xDFA90 (unavailable-unit test), 0x3A1D0, 0x3E590, the 0x3EDE0
contact slice through the miss/success arming calls, 0x3A4E0 base slice, 0x39B40, 0x11D0,
0x14D0 and the 0xE1D90 seeder call (game seed pair passed for non-player kinds). R01, R02,
R03/R09a, R04, R06a, R08, R10a/R11 and R12a..c agree with these instructions; all executable
vectors V01..V11 were recomputed from the rules (seed, mulDiv and roller arithmetic by
script) and match.

| Affected claims | Earlier discrepancy / evidence | Current resolution and limits |
|---|---|---|
| Description | Double-encoded apostrophe in the opening paragraph | Text corrected; no behavioral change |
| R05a / Description | R05a and the Description asserted the missile follows the creature as verified. 0x40860/0x3A1D0 resolve and follow the missile's owner reference; which unit the real Sentinel missile names as owner is set by its creator, which is not read (Q01 candidate 0xA6B70 unproved) | R05a restated as owner-reference following, creature only under V05(c)'s supplied owner; gate Q01/Q04 added in the Id map and gap rows. V05(c) values unchanged |
| R03 dependencies | The collision pass (0x3F320) and end request (0x3EDE0 with no target, force 1) had no named dependency contract or function record | R03 now cites MMC-R05 (plain step), MMC-R11..R16 (contact end) and MMC-R01 (dispatcher removal on return 2); records added above. No behavior change |

These edits change R05a's gate and R03's cited ownership, so a fresh delta reader over
R03/R05a/V05 is required before the blind-implementer condition applies to this revision.

### Independent reader FULL3 (2026-10-15)

Run type: full (only two full runs were recorded before; the review changes to R03's
ownership and R05a's gate also require a full run). A fresh reader received the spec,
foundations, GDF and the section 9 dependency specs (SCD, SCX, MCI, MDP, MMC, STL, MSD, HRC,
MHD) and recomputed 38 lettered/base cases. Counts: MISMATCH 0, GUESS 6, GAP 13, DISCRIM 9,
GATED 5 grouped entries (R05a owner identity; R05b/R05c/R06b/R07/R09b/R10b; Q01..Q10; O01..O03
and stage-only vectors; partial gates on verified rules), COVERAGE 9, HARDEST 3. It found the
Description consistent with R05a's owner-reference hedge.

Triage readings, 2026-10-15, original disassembly of D2Game 0x96310, 0x95560, 0x40860,
0xDFA90 and the 0x3EDE0 entry, and the head of 0x95D50:

| Affected claims | Earlier discrepancy / evidence | Current resolution and limits |
|---|---|---|
| R01 creation request | R01 said summon creation used the resolved target. In 0x96310 the target resolver's outputs go to the command write only; the creation call at 0x95D50 receives (game, caster, 0, 0, skill row, skill id, level). 0x95D50 tests the two position arguments and takes another path when either is zero. The 0x9EE50 selection return is only tested (signed >=0); its out-value is not read again | R01 corrected: caster/skill/level with explicit position (0, 0); target and selection not passed. What the zero position selects inside creation stays Q08/Q09. V09(a) asserts it |
| R03 absent missile and order | 0x40860: null missile and false condition both jump to the end path (end with missile register as supplied, target 0, force 1) and return 2; the owner lookup is reached only when the condition is true. Sync, counter 10, then step are sequential | R03 states the absent-missile path, lookup only on true condition, and the explicit order. No change to existing outputs |
| R02 final dispatcher result | SCD-R07 adds a request and sets result 1 when srvmissile is valid. The srvmissile text is blank for both rows; the compiled selector is not read | Gate: §2 and Q04 name the compiled srvmissile selector; vectors assert callback returns only. No behavior assumed |
| R04 NextHit gate | 0x3EDE0 tests state 86 only when the prepared NextHit byte is nonzero | Already in R04; new V02(g) discriminates it |

| Finding | Disposition |
|---|---|
| GUESS R01 selection not passed to creation | Spec edit: R01 (verified above) |
| GUESS R02 creation start point | No change: R02 already makes every remaining creation option zero/absent; how the creator uses its source is MCI's |
| GUESS R02 no refresh on early returns | No change: R02 orders the duration and target returns before the gate; asm confirms. New V03(g) also asserts no refresh |
| GUESS V05(d) condition in absent-owner case | Spec edit: V05(d) states condition true in both cases |
| GUESS R03 sync/counter/step order | Spec edit: R03 says "in this order"; V05(c) asserts the ordered requests |
| GUESS V01 hit-callback/damage-fill not asserted | Spec edit: V01 states they follow MMC-R14 with supplied results and are not asserted |
| GAP R02/V03 srvmissile/final dispatcher result | Gate: Q04 and §2 edit; TD-BLADE-SKILLS records blank text cells |
| GAP V05(c) caster linkage | Spec edit: the caster is a separate live unit not named by the owner reference |
| GAP V05(d) condition value | Spec edit (with the GUESS above) |
| GAP V05(a) summon initial level/marker | Spec edit: initially level 1, summon marker true |
| GAP V03 creation fields | Spec edit: skill 266, level 10, supplied missile 506, target (130, 230) |
| GAP §5 MMC-owned requests | Spec edit: §5 row names them as MMC-R11/MMC-R14 outputs; R04 asserts only arming's position |
| GAP V04 external weapon not an input | No change: §3 declares it fixture-only context; the vector asserts R12 has no such input, Q05 gates propagation |
| GAP R03 missile room/previous counter, i16 width | No change: synchronization uses the owner's room; the write sets 10 regardless of the earlier value; integer width is a contract element, not a layout |
| GAP R01 marker semantics | No change: already Q09 |
| GAP §3 `Param3`/`calc4` cells | No change: not inputs to verified rules; R06b names `Param3` as a cell only; Q01/Q02 gate their use |
| GAP R04/V01 LastCollide second rejection | Spec edit: V01 supplies LastCollide 0; text cell blank recorded in TD-BLADE-MISSILES |
| GAP R02 absent node/negative missile, GAP R01 invalid row/absent caster | No change: same retained recipes as V03(e)/V09(d), per FULL1/FULL2 |
| DISCRIM V05(d) end result | Spec edit: end returns 0, return still 2 |
| DISCRIM V05(c) collision result | Spec edit: collision returns 0, return still 1 |
| DISCRIM V05 condition skipped | Spec edit: new V05(e) condition false with live owner gives end/return 2 |
| DISCRIM V03(e) duration==0 | Spec edit: new V03(g) P=0 |
| DISCRIM V03 refresh only on blocked branch | Spec edit: new V03(h) open gate with state 12 |
| DISCRIM V02(b) state 86 regardless of NextHit | Spec edit: new V02(g) NextHit 0 with state 86 on |
| DISCRIM V08/V11 mastery mapping | Spec edit: new V11(g), distinct masteries per slot, equal endpoints; mastery constants reconfirmed in 0x3A4E0 (357/330/331/332, burn 316/317 with 329) |
| DISCRIM V08(b) physical/fire order | No change: V11(e) discriminates, as V08 states |
| DISCRIM V01 LastCollide | Spec edit (with the GAP above) |
| COVERAGE R12a/b swap with nonzero percentage | Spec edit: new V10(f); percentage-before-swap gives -9/-5 and a draw |
| COVERAGE R09a condition false with live owner | Spec edit: V05(e) |
| COVERAGE R02 open gate with state 12 | Spec edit: V03(h) |
| COVERAGE R01 invalid row/absent caster; R02 absent node/negative missile/P wrap; R03 other unavailable kinds/modes; R04 absent target/LastCollide/AlwaysExplode; R12 negative maximum; R12b mulDiv 64-bit branches; R10a null seed/rejected class | No change: same output recipes as existing cases, or owned by MMC/NRF/MCI contracts (reasons as FULL1/FULL2) |
| HARDEST R03/V05/R05a | Resolved by the R03/V05 edits above |
| HARDEST R04 with V01/V02 | No change beyond V01/V02(g) edits; MMC/STL own the clamp |
| HARDEST R02/V03 dispatcher result | Gated under Q04 as above |

R01, R03, Q04 and vectors changed, so a fresh full run is required.

### Independent reader FULL4 (2026-10-15)

Run type: full (rule behavior of R01/R03 and the Q04 gate changed after FULL3, so no delta
check applies). A fresh reader recomputed 44 cases (41 lettered, 3 base), including V11(e)
and the new V02(g), V03(g, h), V05(e), V10(f), V11(g). Counts: MISMATCH 0, GUESS 7, GAP 7,
DISCRIM 11, GATED grouped (R05b/R05c/R06b/R07/R09b/R10b; Q01..Q10; STL-Q10, STL attach/free,
NRF-Q20, MMC-Q09; stage-only vectors and O01..O03), COVERAGE 22 items, HARDEST 3. Reader
coverage limit: it reported reading only SCD-R12 from SCD and not opening SCX, MCI, MDP, MSD,
HRC, MHD, GDF or the companion; it states none was needed to execute a vector.

This was the fourth reader run, the cap for this pass. No spec edit was made after it, so
the contract reviewed is the one FULL4 read. The blind-implementer pass is **not complete**:
the entries marked "open" below need the stated edit and a further reader run.

| Finding | Disposition |
|---|---|
| GUESS R01/R02 absent caster | No change: both rules write only a "present" caster's marker and continue; 0x96310/0x95560 test the caster before the write and fall through. Downstream target resolution with no caster is the resolver's |
| GUESS R03 whose completion condition | Open: proposed wording "the owner's completion condition" (0xDFA90 tests the owner unit) |
| GUESS V01/V02 ToHit | Open: proposed fixture ToHit 1 (text cell 1 recorded in TD-BLADE-MISSILES) |
| GUESS V02(c) which filter rejects | Open: proposed fixture "CollideType gate rejects the target type" with stated CT and owner hostility |
| GUESS V05/V09 skill, level, selection values | Open: proposed fixture skill 257, level 10, selection 7 |
| GUESS V05(a) summon room | Open: proposed summon in room B, placement asserted with caster room A |
| GUESS R02 selected missile source | No change: R02 consumes the supplied selection result; its compiled source is Q04 |
| GAP V01/V02 MMC fields (ToHit, CT, owner, Explosion, AlwaysExplode, CollideFriend) | Open: proposed explicit fixture defaults with the GUESS edits above |
| GAP V05/V09 skill/level; selection and summon room | Open (same as GUESS entries) |
| GAP absent caster/absent node vectors | No change: same recipes as V09(d)/V03(e), per FULL1..FULL3 |
| GAP opaque marker/condition meanings | No change: Q09 |
| GAP final dispatcher result | No change: gated by Q04 and §2 |
| GAP MMC-R14 miss return | No change: owned by MMC (MMC-V10); V02 asserts arming position only |
| DISCRIM V05(c) collision propagated when nonzero | Open: proposed second collision result 2, expected return 1 |
| DISCRIM V05(d) end result propagated when nonzero | Open: proposed second end result 1, expected return 2 |
| DISCRIM V05(a)/V09(a) type-2 removal scope | Open: proposed pending type-2 and type-12 events; only type 2 removed |
| DISCRIM V05(a) placement room | Open (summon room B) |
| DISCRIM V09(a) selection passed to creation | Open (selection 7; creation carries none) |
| DISCRIM V09/V03 selection 0 and selected missile 0 | Open: 0x96310 and 0x95560 accept zero (signed >=0 tests); proposed zero-value success cases |
| DISCRIM V03 unsigned timer comparison | Open: 0x95560 uses a signed compare; proposed release -1, frame 100, gate opens |
| DISCRIM V05(d) unavailable-mode table | Open: proposed owner player mode 12 and monster mode 17 (available), player mode 0/17 and monster mode 0 (unavailable), per 0xDFA90 |
| DISCRIM V11 swapped min/max stat ids | Open: proposed asymmetric-pair case with logged read order |
| DISCRIM V04 capture-at-creation | No change: stage limit stated; Q05 gates capture |
| DISCRIM V01 NextHit-independent rejection | No change: V02(g) rejects it, as V01's citation of R04 does not claim that branch |
| COVERAGE (22 items) | No change in this pass; branches with distinct results overlap the open DISCRIM edits above |
| HARDEST R04/V01/V02 | Open with the MMC fixture edits |
| HARDEST V10/R12b | No change: V10(a..f) state operands, branches and wraps |
| HARDEST V05(a)/R01 ignored values | Open with the V05/V09 edits |

Remaining for completion: apply the open fixture/vector edits (no rule behavior change is
proposed; each was checked against the cited instructions), then one fresh full reader run
including SCD-R06..R08 and the MMC fixture defaults.

### FULL4 open edits applied (2026-10-15)

Each open FULL4 entry was checked against original disassembly of D2Game 0x96310, 0x9FA30 (target resolver: X through its register out pointer, Y through its stack out pointer), 0x95560, the 0x91340 and 0x91240 entries, 0x40860, 0xDFA90, 0x3E590, 0x39B40 and the 0x3A4E0 roller call sequence before editing.

| FULL4 entry | Instruction evidence | Edit |
|---|---|---|
| GUESS R03 whose completion condition | 0x40860 passes the looked-up owner to 0xDFA90, which tests that unit's completion bit before its kind | R03: "the owner's completion condition" |
| GUESS V01/V02 ToHit; GAP MMC fields | Text cells in TD-BLADE-MISSILES (ToHit 1, CollideType 3) | Section 8.1 contact defaults: ToHit 1, CollideType 3, Explosion 0, AlwaysExplode 0, CollideFriend 0, CollideKill 0, LastCollide 0, present hostile owner |
| GUESS V02(c) filter | MMC-R11 type gate | V02(c) names a CollideType rejection (CollideType 1 with a monster target after FULL6) |
| GUESS/GAP V05/V09 skill, level, selection | 0x96310: the selection out-value is never reread; creation receives (game, caster, 0, 0, row, skill, level) | V05/V09 inputs skill 257, level 10, selection 7; V09(a) asserts creation carries neither the cast point nor the selection |
| GUESS V05(a) summon room; DISCRIM placement room | Placement room is the room getter applied to the caster | V05(a) summon in room B (at (7, 7) after FULL6); placement asserts caster room A and caster position |
| DISCRIM V05(c) collision result | 0x40860 ignores the 0x3F320 result and returns 1 | V05(c) adds collision result 2 |
| DISCRIM V05(d) end result | End result ignored, return 2 | V05(d) mode-12 case end returns 1 |
| DISCRIM type-2 removal scope | The summon event loop compares the event type byte with 2 before unlinking | V05(a) pending type 2 (due 150) and type 12 (due 160); only type 2 removed |
| DISCRIM V09/V03 zero selection or missile | 0x96310 and 0x95560 use signed nonnegative tests | V09(e) selection 0; V03(i) selected missile 0 |
| DISCRIM V03 unsigned timer | 0x95560 compares release time with frame using a signed jump | V03(j) timer -1 opens at frame 100 |
| DISCRIM V05(d) mode table | 0xDFA90: player modes 0/17 and monster modes 0/12 unavailable; every other kind unavailable | V05(f) mode/kind/completion matrix |
| DISCRIM V11 swapped min/max ids | 0x39B40 reads the maximum (register argument) first; 0x3A4E0 passes 22/21, 49/48, 53/52, 51/50, 55/54, 58/57, 317/316 | V11(d, f) read logs; new V11(h) fourteen-read order |

Lint after these edits: 0 errors, 0 warnings. R03 wording and vectors changed, so the next run was full.

### Independent reader FULL5 (2026-10-15)

Run type: full; first of at most four runs in this task. The reader (Sonnet) received the spec, foundations, GDF and the section 9 dependencies, with SCD-R06..R08 named explicitly. It recomputed 49 cases. Counts: MISMATCH 0, GUESS 4, GAP 7, DISCRIM 10, GATED grouped (R05b/R05c/R06b/R07/R09b/R10b, Q01..Q10, O01..O03, STL-Q10, STL attach/free, stage-only vectors), COVERAGE 14, HARDEST 3. New instruction evidence: 0x3E590 zero-extends NextDelay, schedules event 12 with three zero parameters and attaches with condition 1.

| Finding | Disposition |
|---|---|
| GUESS R02 absolute-target representation | No change: semantic fields listed; MCI owns the request representation |
| GUESS R03 counter write in the log | Spec edit: V05(c)/(f) assert the exact ordered log synchronize, counter write 10, collision |
| GUESS R04 callback/event/attach values | Spec edit: R04 states event 12 without callback and with zero auxiliaries, the opaque MMC-R16 contact-expiry callback and attach condition 1; V01/V02(a) assert them |
| GUESS R03 absent missile end | Spec edit: new V05(g) |
| GAP R01 selection/target arguments | No change: supplied operations, results only; no vector credits their arguments |
| GAP R02 selection/P arguments | Spec edit after FULL6: arguments verified and asserted |
| GAP expiry callback identity | No change: MMC-R16 opaque token; position asserted |
| GAP owner-reference condition type | No change: Q09 |
| GAP V01 hit handler not asserted | Spec edit: V01 asserts the hit-handler request after hitbymissile (handler supplied 0); later flow is MMC-owned |
| GAP V07 supplied seeding creator | No change: declared supplied-input boundary |
| GAP R12c other field copies | No change: explicitly not asserted |
| DISCRIM V05 monster mode 17 | Spec edit: V05(f7) |
| DISCRIM player completion / other kinds | Spec edit: V05(f8) player completion, V05(f9) missile kind 3 |
| DISCRIM V05(c) order | Spec edit (with the GUESS above) |
| DISCRIM arming before to-hit | Spec edit: V01/V02(a) logs start with the to-hit request |
| DISCRIM arming after the handler | Spec edit: V01 hit-handler request follows arming |
| DISCRIM attach argument / schedule parameters | Spec edit: asserted |
| DISCRIM NextDelay signedness | Spec edit: V02(h) NextDelay 200, expiry 300 |
| DISCRIM progression vs missile order | Spec edit: V03(k) selected missile -1 makes no progression request (0x95560 calls 0x91340 before 0x91240) |
| DISCRIM V09 row vs selection order | Spec edit: V09(f) invalid row makes no selection request (0x96310 bound check precedes 0x9EE50) |
| DISCRIM percentage identity -1 | Spec edit: V11(e) percentage read log 329, 357, 330, 331, 332, 329 and none for physical; identity -1 is not used by R12c callers (COVERAGE) |
| COVERAGE (14) | No change beyond the edits above: remaining branches share existing result recipes or are MMC/NRF/MCI-owned |
| HARDEST V05(a)/R01, V01/V02, V10(d, e) | No change beyond the edits above |

### Independent reader FULL6 (2026-10-15)

Run type: full (R04 argument wording and vectors changed). It recomputed 53 cases. Counts: MISMATCH 0, GUESS 10, GAP 10, DISCRIM 7, GATED grouped, COVERAGE 8 groups, HARDEST 3. New instruction evidence: 0x95560 calls 0x91340 with (skill identity, caster) and 0x91240 with (caster, skill identity, invocation level).

| Finding | Disposition |
|---|---|
| GUESS 8.8 vs raw component | No change: section 3 states raw integer vector values; R12 returns the roll |
| GUESS V04 weapon ranges | No change: context only |
| GUESS R01 target/selection arguments | No change: supplied operations |
| GUESS R01 absent caster | No change: as FULL4 |
| GUESS R02 progression arguments; DISCRIM V03 argument order | Spec edit: R02 names (caster, skill) and (caster, skill, level); V03(a) logs them |
| GUESS refresh literal 1 | No change: opaque literal, Q09 |
| GUESS owner resolution | No change: supplied lookup, Q04 |
| GUESS V05(b) requested vs stored | No change: same value; STL-Q10 |
| GUESS hit-handler arguments | No change: MMC-owned, not asserted |
| GAP summon position; DISCRIM placement position | Spec edit: summon at (7, 7); placement asserts the caster position |
| GAP caster level read | Spec edit: defaults state the layer-0 stat 12 read returns 30 |
| GAP creation start, to-hit arguments, contact return, missile 506 identity, weapon interface, i16 width, command shape, dispatcher result | No change: MCI/MMC-owned, fixture-only, gated by Q04/Q05, or width already stated |
| DISCRIM V02(c) CollideType 0 only | Spec edit: V02(c) uses CollideType 1 with a monster target; remaining type predicates are MMC-R11/MMC-V08 |
| DISCRIM V10(e) exact mulDiv | No change: V10(d)'s return 202885 rejects exact-precision mulDiv |
| DISCRIM V08 swapped physical ids | No change: V11(h) read log |
| DISCRIM V07 action-side seeding | No change: V03(a) asserts no draws with a non-seeding creator |
| DISCRIM placement success | Spec edit: V05(a) placement returns 0 and 1 |
| COVERAGE / HARDEST | No change |

### Independent reader FULL7 (2026-10-15)

Run type: full (R02 argument wording changed). It recomputed 53 cases. Counts: MISMATCH 0, GUESS 6, GAP 9, DISCRIM 6, GATED grouped, COVERAGE 16, HARDEST 3.

| Finding | Disposition |
|---|---|
| GUESS R01 row check vs selection order | Spec edit: R01 states the row check precedes the selection request (0x96310) |
| GUESS R02 release-time read position | No change: not observable; rule order stated |
| GUESS R03 kind vs completion order | No change: both give unavailable; equivalent outcome |
| GUESS V01 same or different missile; DISCRIM per-missile cooldown | Spec edit: V01's contact at 110 is a different missile with fresh fields |
| GUESS ToHit 0; R12b read position | No change: MMC-owned coverage; matches the rule |
| GAP creator shape, stat 19/handler arguments, local fixture inputs, stat names, last collision, step arguments, post-handler return, dispatcher result, V05(b) | No change: MCI/MMC-owned, fixture-declared, TD-BLADE-STATS, or gated by Q04/STL-Q10 |
| DISCRIM V03 refresh on early returns | Spec edit: V03(e, f, g, k) have state 12 true and no refresh (0x95560 early returns bypass the state 12 test) |
| DISCRIM owner/source swap | No change: both are the caster by rule; equivalent within this boundary |
| DISCRIM V04 snapshot | No change: R08 scope; Q05 |
| DISCRIM NextDelay clamp; V05(d, e) | No change: V02(h) rejects the clamp; the reader found no remaining alternative for V05 |
| COVERAGE / HARDEST | No change |

### Independent reader FULL8 (2026-10-15)

Run type: full (R01 order wording and vectors changed). This was the fourth and last run allowed in this task. It recomputed 53 cases. Counts: MISMATCH 0, GUESS 11, GAP 9, DISCRIM 11, GATED grouped, COVERAGE 20, HARDEST 3. Reader coverage limit: it opened only parts of NRF, GDF, SCD, MMC, STL and SCX and did not open MCI, MDP, MSD, HRC, MHD or the companion; it states that no vector needed them. No spec edit was made after this run, so the contract it read is the current one.

| Finding | Disposition |
|---|---|
| GUESS R01 absent caster | No change: the rule writes only a present caster's marker and continues to target resolution (0x96310 tests and skips), as FULL4 |
| GUESS placement zeros / absent caster room | No change: literal arguments; meanings Q09 |
| GUESS R02 node key | No change: SCD-R06 owned-skill lookup; supplied result |
| GUESS creation request shape | No change: MCI-owned request; semantic fields listed |
| GUESS state 12 test | No change: NRF-R12a active-state read, GDF-R12 |
| GUESS owner without a room | No change: supplied synchronization; Q04/Q09 |
| GUESS kind vs completion order | No change: equivalent outcome |
| GUESS miss/hit contact returns | No change: MMC-R14/R15 and MMC-V10 own them; not asserted |
| GUESS V05(b) | No change: STL-Q10 |
| GUESS stat-read adjustment | No change: fixtures supply returned layer-0 values |
| GAP P=6 | Gated: supplied P, Q03 |
| GAP missile 506 | No change: fixture id; TD-BLADE-MISSILES Id 506; Q04 |
| GAP weapon A ranges | No change: context only, not consumed |
| GAP fixture-only rooms/events | No change: declared fixtures |
| GAP absence outputs not in section 5 | No change: they are absences of section 5 outputs (creation, refresh, type 2 removal/request) |
| GAP creation argument list | No change: as the GUESS above |
| GAP SCD activation/mana requests | Gated: Q02/Q04; vectors assert callback returns only |
| GAP event removal key | No change: R01 removes the summon's pending type 2 events |
| GAP unopened dependencies | No change: the reader states no vector depends on them |
| DISCRIM V02(b, c) filter order | No change: MMC-R11 owns the order; MMC-V08(a) (state 86 rejects with last collision unchanged) and the MMC-V08 base case (last write before type rejection) discriminate it |
| DISCRIM NextHit==1 vs nonzero | **Open**: every BSF and MMC vector uses NextHit 0 or 1. Proposed edit: a V02 case with NextHit 2 (state 86 on rejects; a miss arms). 0x3E590 tests the byte as nonzero; the 0x3EDE0 state 86 entry test should be rechecked before editing. Needs a further reader run |
| DISCRIM state bit vs list | No change: V02(b) has state 86 on with no list, so a list test would not reject |
| DISCRIM arming after damage fill | No change: V01 places arming before the hit-handler request, which precedes the fill under MMC-R14; MMC-V10(e) orders handler then fill |
| DISCRIM handler on a miss | No change: MMC-V10(a, b) own the AlwaysExplode miss branch |
| DISCRIM owner/source swap | No change: equivalent within this boundary |
| DISCRIM frame+duration overflow | **Open**: R02 states i32(frame+duration) under section 3's low32 wrap, but no vector exercises it. Proposed edit: a V03 case with frame 2147483647, timer 2147483646, duration 5, expecting timer -2147483644 (0x95560 uses a plain add), then a further reader run |
| DISCRIM V04 snapshot | No change: R08 scope; Q05 |
| DISCRIM V08 order | No change: V11(e) |
| DISCRIM dispatcher always 1 | Gated: Q04 |
| DISCRIM negative percentage with a large endpoint | No change: NRF-R05 owns mulDiv branches and their vectors; R12b consumes it in the stated operand order |
| COVERAGE (20) | No change: advisory; shared result recipes or MMC/NRF/MCI-owned |
| HARDEST V01/V02, V05, V10(d, e) | No change: order and owners stated |

**Outcome.** Four reader runs (FULL5..FULL8) were made in this task, and every run returned MISMATCH 0. The blind-implementer pass is **not complete**: FULL8 leaves two DISCRIM entries open (NextHit values other than 0 or 1; the R02 timer wrap). Both need a vector edit and one further full reader run. Every other FULL8 entry is resolved or gated above. The document remains draft.
