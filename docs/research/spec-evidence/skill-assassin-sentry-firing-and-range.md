# Evidence: Assassin sentry firing and range

Spec: [Assassin sentry firing and range](../../specs/skill-assassin-sentry-firing-and-range.md).
Artifacts: [original source manifest](../source-manifest.md).
Limits: [evidence gaps](../../plans/gaps/skill-assassin-sentry-firing-and-range.md).

## Identity and evidence reuse

Lord of Destruction Expansion 1.13c, Windows x86. Module references resolve to the
manifest's `LOD113C-D2GAME` and `LOD113C-D2COMMON` fingerprints. All descriptive function
names here are analyst labels, not original exported names. An RVA is module relative,
not a physical file offset. Original image fingerprints were checked against the manifest
before the new branch/table inspections. No game files, rows or raw captures are published.

The original 2026-10-12 static record read the sentry think, budget, target callback,
candidate validator, stored-target resolver and distance paths in original disassembly;
the point-distance operation also had a byte-equivalent C view. Original import resolution
and the RNG audit of that date establish the reported direct-site seed owners. Those
readings retain their date and actual limits. The available original images correspond to
the manifest specimens; identity correspondence does not itself reverify behavior.

Before additional reading on 2026-10-15, the existing spec, gap comparison, source manifest
and current NRF/GDF/SCD/SCX/SLD contracts were assessed. Compatible records were reused:
NRF-R01/R02/R03a/R04a/b for draw forms/initialization/site ownership, NRF-R10a/b for
supplied dynamic positions, GDF-R12 for state 146 invis, and SCD/SCX/SLD for action dispatch,
calculation bounds and their precise unresolved producers. No universal highest-level
instance contract was found: selection/bonus production remains SFR-Q12, consistent with
SCX-Q09/SCD-Q10. No gameplay actions, new hooks or other-build analysis were performed.

## Dispatch and function records

| Artifact | Function or table RVA | Analyst description / rules | Method and actual boundary |
|---|---|---|---|
| LOD113C-D2GAME | 0xA7DC0 | reached sentry think, SFR-R01a..g/R06a | Original disassembly 2026-10-12; full entry-to-exit reread 2026-10-15 for parameter loads, ordered calls, second-budget return, event walk and skill-mode request. Actual AI101 entry remains unanchored |
| LOD113C-D2GAME | 0xA4A10 | budget, SFR-R02 | Original disassembly 2026-10-12; full reread 2026-10-15: owner/room early reject, signed lazy-init/counter tests, expression call, decrement and return. External producer effects remain Q12 |
| LOD113C-D2GAME | 0xF21E0; 0x75070; 0x75E10 | death-mode request construction | Reused 2026-10-12 disassembly: zero mode reaches request. Actual complete monster-mode execution is outside supplied budget stage |
| LOD113C-D2GAME | 0xF22C0; 0xF2380; 0x8E6F0 | delay, skill-mode, event-removal calls from think | Caller arguments/read 2026-10-15; full producer contracts are Q01/Q12, not verified by caller shape |
| LOD113C-D2GAME | 0xB1BC0 | stored-first target wrapper, SFR-R03a | Original 2026-10-12; complete 2026-10-15 entry/exit reading establishes shortcut, initial distances, chooser result, optional reach and distance writes |
| LOD113C-D2GAME | 0xB1820; jump table 0xB1A9C | stored-target resolver, SFR-R12/Q08 | Original 2026-10-12 direct-site audit; full listing assessed 2026-10-15. Four stored modes, hash buckets, selector 3 draw, temporary selector/state writes, selectors 5/6 and output validity identified; complete producers remain unverified |
| LOD113C-D2GAME | 0xB17B0; table 0xFB588 | selector dispatch, SFR-R03b | Full 2026-10-15 dispatch reading and original-image table cells: selector 6 `{kind0, callback0xB15E0}`, selector 7 `{kind0, callback0xB1580}`. Kind0 dispatches 0xB0720 |
| LOD113C-D2GAME | 0xB0720 | room walker, SFR-R03b | Complete listing assessed 2026-10-15: room lookup, adjacent-room-list request, per-room chain and early callback return. Encoded portions/room-list producer keep actual traversal unresolved |
| LOD113C-D2GAME | 0xB15E0 | two-slot candidate callback, SFR-R03c | Original 2026-10-12; complete 2026-10-15 reread: validity before distance/cap, slot split, strict improvement, collision 4 before writes, zero return |
| LOD113C-D2GAME | 0xB1580 | selector 7 callback | Complete 2026-10-15 reading: excludes supplied slot B unit, validates, caps at 48, requires class>1, strict improvement and collision 4 before slot A writes |
| LOD113C-D2GAME | 0xB08A0 | candidate slot class | Full 2026-10-15 reading: player 14; bounded MonStats class index reads byte+0x4E; other/missing 0. Exact column name remains unknown; Align is +0x4C, so it is not substituted |
| LOD113C-D2GAME | 0xB1AB0 | chooser, SFR-R10 | Full 2026-10-15 reading and complete original-image decode of encoded instructions: player/empty slots, distance 5, temporary path setters/restoration, selector 7 and distance 20. Path helper effects remain Q08 |
| LOD113C-D2GAME | 0xB1400 | validity, SFR-R03f/g | Original 2026-10-12; full original-image decode 2026-10-15 confirms both opposite branch polarities and ordered state/candidate-seed checks |
| LOD113C-D2GAME | 0xDFA90 | unavailable condition | Full 2026-10-15 listing: null/completion condition or player mode 0/17, monster mode 0/12 returns 1; otherwise 0 |
| LOD113C-D2GAME | 0xE0A60 | final root decision, SFR-R03i | Original 2026-10-12, complete 2026-10-15 reread: one searcher owner step, repeated candidate-owner steps, self-root reject, player interaction 8 or inverted relation. Owner lookup0xE03A0 consumes PVP-R07 |
| LOD113C-D2GAME | 0xF02A0 | player interaction 8 request | Full listing assessed 2026-10-15: selected player relation entry flags masked by 8, or 0. Existing list read consumes PVP-R05a; actual relation production remains PVP-Q03 |
| LOD113C-D2GAME | 0xB08F0 | unit-to-point metric, SFR-R04 | Reused 2026-10-12 byte-equivalent C/disassembly: signed absolute offsets and octagonal truncation; supplied position boundary |
| LOD113C-D2GAME | 0xB0A20 | candidate-size distance, SFR-R13 | Reused 2026-10-12 disassembly: candidate in EBX, searcher stack argument; size subtraction on both axes then zero clamps and octagonal expression. Both room and stored callers supply this order |
| LOD113C-D2COMMON | #10589, 0x30F80; helpers0x013B0/0x01150 | size by kind | Reused import/hand decode 2026-10-12; new byte-equivalent C reconstruction read 2026-10-15 confirms player 2, signed monster byte 8, object nested dword+0xD0, missile byte+0x18A, item 1, other/null 0 |
| LOD113C-D2COMMON | 0x56C10; string0x8FA40 | MonStats2 SizeX identity | Reused 2026-10-12 field-descriptor disassembly and original string read: SizeX conversion byte at offset 8; MonStatsEx linkage through MonStats+0x18 |
| LOD113C-D2GAME | 0x1074A8; 0x43B60 | action table 17 and burst | SCD's 2026-10-15 table/root record plus new full handler reading/original-image decode: latch before validation, missile bounds, `calc1`, per-index resolution and two nonzero-coordinate gates, conditional creator and return 1 |
| LOD113C-D2GAME | 0x41F50 | burst initializer | Byte-byte-equivalent C reconstruction read 2026-10-15; callback cap and seed/path requests from entry to return; imported setter/path effects remain Q12 |
| LOD113C-D2GAME | 0xB6280; 0xB5D50; 0xB6310 | summon context and Blade first-event contrast | Reused 2026-10-12 record: action 45 summon/mode update; Blade action 44 requests frame+1. This does not establish sentry first due frame |
| LOD113C-D2GAME | 0x6F930 | shared creator speed | Reused 2026-10-12 speed-stage record; targeted branch region reread 2026-10-15 confirms explicit-speed, state/list gate and scales. Full creator/output/flight integration remains Q05/Q12 |

The think's missing-skill path clears current skill through D2Common #11103 or its
recorded request equivalent before its death-mode request; the budget's lazy-init kill
wrappers and invalid-owner path are separate callers. No exact event or animation effect
is inferred merely from the request label. Full semantic import identities established
below remain navigation records; unknown imports are not silently assigned semantics.

### Import identities and semantic mapping

| D2Game thunk RVA | LOD113C-D2COMMON export / target RVA | Supported caller role / limit |
|---|---|---|
| 0xA754 | #11138 /0x326C0 | reach predicate(searcher, candidate, 0), not hostility; new import resolution 2026-10-15 |
| 0xAD78 | #10839 /0x314A0 | collision/range-line request(candidate, searcher, mask); resolution reused/new corroboration 2026-10-15, producer incomplete |
| 0xA6F4 | #10666 /0x84D30 | root relation, inverted by final decision; new import resolution/read 2026-10-15, complete classification reader0x391F0 consumes the dated PVP-R05a record; actual tier-list production remains PVP-Q05 |
| 0xA808 | #10057 /0x3C390 | room-to-level predicate, reads level through0x2D320; current consumer does not guess wider flags |
| 0xA478 | #10331 /0x2FE10 | unit room; SCD's current import record |
| 0xA3F4 | #10630 /0x4F070 ->0x4E480 | skill-instance request; new thunk/export forwarding read; selection/bonuses remain SCX-Q09/SCD-Q10 |
| 0xA71E | #10589 /0x30F80 | size operation; original 2026-10-12 import record |
| 0x9454 | #10867 /0x34BB0 | dynamic integerX; NRF-R10b owns current getter contract |

The initializer import mappings were resolved on 2026-10-15 and their getter/setter leaves read in source order:

| D2Game thunk RVA | LOD113C-D2COMMON export / RVA | Actual callback operation and method |
|---|---|---|
| 0xAD06 | #10534 / 0x6A290 | signed i16 missile frame count; byte-equivalent C reconstruction |
| 0xAD42 | #10558 / 0x6A710 | total-frame word setter with signed i16 clamp; complete disassembly |
| 0xAC7C | #10970 / 0x6A6D0 | current-frame word setter with signed i16 clamp; complete disassembly |
| 0xAAC6 | #10983 / 0x34AB0 | unsigned stored-point X at Path+0x10; byte-equivalent C reconstruction |
| 0xA628 | #10614 / 0x36740 | InitSeed writes value and 666; reused NRF contract and byte-equivalent C reconstruction |
| 0xAAFC | #10647 / 0x34F80 | path kind request; complete disassembly, original path flag/table effects remain Q12 |
| 0xAB14 | #10168 / 0x34F50 | low-byte duration is capped at 77 and stored in two path byte fields; complete original-image decode corroborated the low-byte/cap/writes on 2026-10-15; broader path integration remains Q12 |
| 0xAAF0 | #10334 / 0x36230 | path preparation; complete producer integration remains Q12 |

The original 'pathX' wording conflated the initializer's stored-point getter with the dynamic
current-X getter. Export 10983 reads Path+0x10, matching SCD's stored-point field, rather
than dynamic current-X+0x02. V07/V19 retain supplied stored-point 1000 and seed outcomes
1000/1001/1002; equality with caster currentX is no longer asserted. The creator's actual
stored-point production remains gated. The duration setter narrows the callback frame
argument to u8 before capping at 77; a negative signed frame count therefore cannot be
modeled as min(f,77). The callback's total/current frame clamp still tests signed f>=78.

The initializer's imported operations read current frame count, set total/current frame
counts, read path stored-point X, initialize seed, set path kind/duration and prepare path. The
byte-equivalent callback contains no direct LCG step; only the seed-initializer request
is local. Actual underlying path allocation/motion is unverified. Likewise action 17's
target resolver at D2Game RVA 0x9FA30 and missile selector0x91340 are named supplied
operations; action-level request handling is separated from their complete production.

### Semantic layout correspondence

These locations identify evidence; they are not implementation requirements.

| Semantic term | Original representation | Rules |
|---|---|---|
| difficulty | Game+0x6D u8 | R01/R06a |
| `aip1`..4 per difficulty | MonStats+0x56/+0x5C/+0x62/+0x68, each difficulty advances 2 bytes, signed i16 | R01/R06a |
| firing Skill1 | MonStats+0x170 signed i16 | R01/R02 |
| shot counter | AI-control+0x18 signed i32, initial-1 | R02 |
| `calc1`/`calc4` | Skills+0x138/+0x144 | R09a/R02 |
| action selector | Skills+0x2E signed i16 | R09 |
| dynamic integer positions | Path+0x02/+0x06 u16 | R04/R07b/R13 |
| static positions | Path+0x0C/+0x10 i32 | R04/Q06 |
| seed | Unit+0x20/+0x24 | R01e/R03g/R12/R09a |
| candidate eligibility A | Unit+0xC4 bit 2 | R03f/Q09 |
| completion condition | Unit+0xC6 bit 0 | R03f |
| owner reference | MonsterData+0x28 then owner link+0x28/+0x2C/+0x30 | R02/R03i |
| missile owner-resolution condition and identity | Unit+0xC8 bit 10; +0x94/+0x98 | R03i |
| stored-target selector/state | MonsterData+0x38/+0x34 | R12/Q08 |
| two slots | block+0/+4 unit/distanceA, +8/+12 unit/distanceB | R03c/R10 |
| missile initialization callback/index | creation request+0x54/+0x58 | R09a |
| explicit speed/condition | creation flags bit 2; already-scaled bit 4; value+0x28 | R11a |
| velocity fields / slow match | Missiles+0x9A/+0x9B u8; flags byte+4 and global mask byte | R11a |

## Corrections and affected assertions

### State146 predicate conflict (SFR-Q11)

The 2026-10-12 R03g prose and V14(a) described a same-polarity hostility filter for
players and monsters. The same version's Q11 already contradicted that prose with
opposite branches. On 2026-10-15 the original import chain resolves thunk0xA754 through
IAT RVA 0xF84CC to D2Common export 11138, RVA0x326C0. Full original-image decoding of
D2Game RVA 0xB1400 confirms monster call0xB1473 followed by zero-to-final branch0xB147A;
nonzero falls through reject0xB147C. Player call0xB1498 follows its candidate roll;
zero branches at0xB149F to reject, nonzero proceeds to final0xB14A1. This establishes
opposite polarity independently of the helper label or sibling-spec claims.

D2Common RVA 0x326C0 original-image decode shows no hostility lookup. Its null inputs return 0.
For candidate monster class 258/261 it has a special separation/range+8 success branch.
The ordinary branch obtains searcher range through0x325C0, adds the supplied third
argument plus 1, compares the0x7FCD0 separation result, and when close enough requests
collision 0x314A0 with mask0x804, returning 1 only when that result is 0. Range helper
0x325C0 branches through monster extended data/equipment dependencies; separation
0x7FCD0 uses sizes and an original short-distance table. Their complete weapon/range/separation/collision
producer chains are not all verified, so SFR-Q11 now gates actual reach-result integration.
Local predicate handling is verified with supplied results. V12/V13/V23 retain their
numeric seed/roll outcomes; their old supplied 'hostile' condition becomes the precise
reach return. V14(a) now accepts invisible monster reach 0 when final passes; V14(c)
checks reach 1 rejection. No numeric fixture was fitted to gameplay observation.

The earlier R03h claim of 'no melee-distance code on this path' is therefore withdrawn:
the imported reach helper reads spatial separation. The separate fact that final
decision0xE0A60 has no distance test remains supported. It is not the unrelated
function0xDFA60; local assembly labels had previously made that distinction confusing.

### Think and burst gates

Full think reading confirms that the second, decrementing budget call's return is tested
at RVA 0xA7ECF/0xA7ED1: nonzero stops before pending-event deletion/skill selection. Earlier pseudocode
omitted that return check. V25 supplies owner loss on the second call and leaves events
untouched; it supplements the retained V05/V17 successful sequences.

Full action 17 image decoding confirms that `calc1` gives loop attempts, not guaranteed
created missiles. Target resolver0x9FA30 must succeed and both coordinates must be
nonzero before0x6F930 is requested; failed creation does not abort the loop or change
return 1. V07/V19 retain their original successful-three-bolt numeric outcomes with
success/placement explicitly supplied. V34 adds resolver/coordinate/failure cases.
The callback reads each missile's OWN path stored-point X; equality to casterX is conditional on
the supplied creation placement, rather than an unconditional creator guarantee.

### Labels, units and draw totals

GDF-R12's dated state record establishes 146 invis; only the eligibility-conditionA name
remains SFR-Q09. NRF-R10b establishes supplied dynamic integer-part units and does not
establish every static/path producer. R07b is verified at that supplied getter boundary;
the original broader coordinate question remains Q06. Player range is enlarged by the
FULL size 2 per axis, monster by signed SizeX, not half a collision radius.

R03i and R12 lacked executable vectors in the older record. V31 and V37 now make their
supplied root decision and reached direct-site draw assertions executable. R12 consumes
NRF's exact RandMask(m)=newLo & (m-1), so argument 2 returns 0/1. Expression/calculation,
creation and mode operations can draw separately. No-direct-draw statements never claim
the total whole-think or creator draw count.

R03b/R03e/R03h/R10 and full speed/footprint/timing remain gated or asm-only at their
stated boundaries. Prior verified assertions are retained as local supplied-input
contracts where justified; complete imported producers are not promoted by relocation.
Unsubstantiated 'not found' labels are replaced by unresolved questions because no
complete six-rung absence record exists. All original evidence is preserved privately;
this companion retains usable original-artifact claims, rationale and dates.

## Named table data

### TD-SKILLS17

Date 2026-10-15; effective virtual file `data/global/excel/skills.txt`; running source
profile reported vanilla LoD launch key `lod-113c`, excluding the separate PD2 profiles.
Current profile selection reuses the day's recorded member inventory and vanilla identity.
The current versions-console refresh had failed; no complete current override inventory,
physical archive identity or pristine archive claim is made. Prior recorded override
inventory reported override redirection disabled. Code hashes alone do not establish data.

The game's Storm loader returned successful contiguous requests `(off, length)`
`(0, 60000)`, `(60000, 60000)`, `(120000, 53535)`, each returned exactly its requested
length; total 173535 bytes. Complete decoded bytes include the header and all target rows.
SHA-256 of the joined effective file:
`465c28e83657540e0d7ae5b0de826d5657be7fba1aa539a85c3ca3aeea34b498`.
Columns were parsed by their header names. Zero-based data-row positions exclude the
header; explicit Id cells agree for the named rows below. Only narrow cells are published.

| Row identity / zero-based row | Column | Raw cell | Interpretation / supported claim |
|---|---|---|---|
| Charged Bolt /38 | Id; srvdofunc | 38; 17 | explicit skill identity 38 selects text action 17, R09b |
| PrimeBolt /186 | Id; srvdofunc | 186; 17 | identity 186 selects text action 17, R09b |
| BoltSentry /306 | Id; srvdofunc | 306; 17 | identity 306 selects text action 17, R09b |
| ImpBolt /322 | Id; srvdofunc | 322; 17 | identity 322 selects text action 17, R09b |
| BoltSentry /306 | `calc1` | `ln12` + skill('Shock Field'.blvl)/3 | raw expression text only; no compiled evaluation claim |
| BoltSentry /306 | `calc4` | `par8` + skill('Lightning Sentry'.blvl)/4 | raw expression text only; no compiled evaluation claim |

Full file coverage establishes these are the four rows whose raw srvdofunc cell equals 17.
This is `verified in 1.13c data`, not gameplay execution or compiled-table correspondence.
Actual MonStats Skill1/aip values and expression compiler output/skill binding remain Q04.
No unrelated monster-skill behavior is added by listing the other text selectors.

## Gap assessment and search limits (2026-10-15)

| Original question / gap | Compatible evidence and targeted follow-up | Current disposition |
|---|---|---|
| Q01 first/post-shot timing; R05/R06b | Reused summon/action 44 contrast; reread reached think/budget and identified scheduler/mode requests | Still open: actual scheduling/mode-end producers and O01/O02 |
| Q02 AI101; R01g | New original-image absolute-pointer scan finds think pointer at D2Game RVA 0x10EDE0; no table-index consumer anchor established | Entry binding unresolved; not an absence claim |
| Q03 footprint; R08 | Reused generic creator context; callback C confirms path 10 request but not steering/collision | Missile/path owner integration and O03 remain |
| Q04 original rows; R09b | SCD table bounds reused; complete narrow-purpose Skills loader read above | do 17 text membership closed; calc/aip/compiled/runtime association still open |
| Q05 speed; R11a/b | Reused speed region and new targeted flags/state/list-scale reading | Kept asm-only/full output and travel integration open; V36 numerical supplied-stage cases |
| Q06 coordinate/seed | NRF-R10a/b supplies getter units; NRF-R01/R02/R04 and original candidate-site record supply owner/modulus | Dynamic getter and candidate RNG settled; other position producers/extreme static domain remain |
| Q07 other-build equivalence | Evidence policy permits only 1.13c; no other build read | Outside evidence build; no equivalence claim and no implementation gate |
| Q08 two slots; R03b/e/h/R10 | Full dispatcher/callback/class/chooser assessment; selector 6/7 original-image cells; encoded chooser corroborated | Supplied choices refined; real stored modes/room-list/path/collision/class-field producers remain |
| Q09 state/flag | GDF-R12 state 146 named invis; unavailable helper read | State name closed; candidate eligibilityA meaning/producer open |
| Q10 candidate size | Reused original import/descriptor/metric/call-site evidence; size helper byte-equivalent C reconstruction corroborates type dispatch | Static size arithmetic retained; O04 still pending |
| Q11 polarity/hostility | Import resolution and complete b1400/326c0 image decode above | Opposite polarity/incorrect hostility interpretation corrected; actual range/separation/collision integration remains |
| R03i andR12 missing vectors | Reused local final decision and NRF reached draw contracts | V31/V37 close executable-fixture gap; complete producers still gated |
| R09a overbroad burst result | Complete handler image and callback C | Request/initialization assertions available; creation/path effects explicit Q12 |
| All pending gameplay | No campaign authorized in migration; retained all four recipes | Q13 prevents runtime-confirmed status; not a local arithmetic gate |

No new six-rung search was claimed complete. The AI pointer search supplies a new data
lead, not the missing dispatch consumer or a verified identity 101. The room-class field
was checked against the existing GDF map (Align+0x4C is distinct); its name remains neutral.
The complete original-image selector 6/7 cells were read at0xFB5B8/0xFB5C0 and direct
dispatch instructions read through their callback selection; actual room/path producers
remain explicit missing boundaries rather than guessed 'nearest hostile' behavior.

## Observation procedures

No observation has been collected. Use an owned original 1.13c installation matching
the manifest and establish its effective data separately. Common controls: Assassin level 30,
one chosen trap rank with sufficient prerequisites, no equipment granting skill/speed changes,
Normal difficulty, same act/level/room, same owner state and stable stationary target.
Record exact rank, synergy/base levels, data values, positions and relevant seeds at entry;
level 30 is a practical setup, not a behavior constant. Coordinate units consume NRF-R10.

- **SFR-O01:** Place the Assassin and a surviving stationary eligible enemy so that the
  target is within the verified reduced range. Cast one sentry. Record cast, creation,
  first AI call and first shot frames. Repeat independent controlled ranks. Expected:
  local think gates match; first due frame is measured for Q01, not guessed.
- **SFR-O02:** With a positive known budget and an in-range eligible target, record one
  shot, its skill-mode completion and next think. Distinguish scheduled event due time
  from actual dispatch. Expected: no `aip2`/`aip3` local success delay; exact later frame
  remains Q01 until producer reading and capture agree.
- **SFR-O03:** Use a charged-bolt sentry and one stationary target; retain each per-index
  target-resolution, creation success and initialized missile identity. Record initial
  X/seed/frame count and successive direction/position readings until expiry. Expected:
  V07/V19 callback values only for actually initialized missiles. Steering/angle series
  and count losses are measured independently; do not infer a fixed angular spread.
- **SFR-O04:** With the supplied or authenticated `aip4`=12 case, keep all eligibility,
  chance and size conditions matched. Place a player at raw offsets(13, 0)/(14, 0) and a
  monster with authenticated SizeX3 at(14, 3)/(15, 3), relative to the sentry. Record reduced
  distances, selection and shot outcome. Expected11/12 with only the 11 case passing
  strict range. Room-cap controls use reduced 48/49; do not conflate stored shortcut.

## Migration and review record

2026-10-15 migration preserves the existing shared-family file boundary. Searches for the
filename and SFR prefix in implementation/test paths and non-gap plans found no consumer;
sibling/gap citations alone do not make ids consumed. Practical ids are retained without
legacy aliases. All24 original vector families remain V01..V24; symbolic cases now have
complete numerical inputs, seeds, ordered effects and exact outcomes. Additional cases
V25..V37 separate newly established gates and supplied dependency boundaries. Four
observation recipes remain; the assessment table accounts for all 11 original questions,
including outside-buildQ07 and settled sizeQ10. The original spec/gap files and detailed
private investigation encodings/history are preserved in a hash-checked private archive.

This record does not claim a fresh blind-implementer review. The owner coordinates that
review after the writer's stable lint, preservation and self-review handoff. The status
remains draft because behavior/integration questions and possible gameplay observations
remain. No implementation comparison was newly performed; the existing comparison is
retained as historical in the gap file.

### Compatible final-predicate owner contract (2026-10-15 reuse assessment)

The current [PvP evidence](combat-pvp-damage-and-kill-credit.md#dispatch-and-output-anchors)
contains the exact same-fingerprint Game RVA 0xE0A60 ->0xF02A0 or Common #10666/0x84D30
->0x391F0 chain, with complete newly read/image-corroborated helper records dated
2026-10-15 and historical 2026-10-11/12 support. PVP-R05a/R07 therefore supplies the
existing-list tier read, directional raw 0/8 interaction, one-hop searcher owner and
recursive candidate owner with missing/self-link stop. There is no multi-unit cycle
detector. SFR-R03i reuses this compatible contract; no repeated binary reading or
generic unread-root-relation gate is needed. PVP-Q03/PVP-Q05 still limit production
of relation/alignment storage, rather than the verified read of supplied existing lists.
The spatial reach import #11138/0x326C0 is a different helper and remains SFR-Q11.
V31 now uses realizable existing relation/tier inputs rather than an arbitrary nonzero
relation result; V31(f) discriminates asymmetric owner depth.

Writer checks on 2026-10-15: mechanical lint, full current-contract self-review and
private preservation checks completed before handoff. They verify documentation shape,
cross-reference/link resolution, all 24 vector identities, all four recipes, all 11 original
question assessments and archived original hashes. Exact two-/three-draw sequences were
recomputed using the NRF LCG for the new concrete budget fixtures. Extra pure predicate
read counts are not acceptance requirements; source routing remains explanatory Trace.
Fresh blind-implementer review remains pending root coordination. No implementation
tests were run, and no gameplay observation is implied.

## Blind-implementer FULL1 triage (2026-10-15)

The first full reader examined the canonical pre-fix contract: 30 rule ids (19 static
code, 1 data, 2 asm-only and 8 unresolved), 37 vector families and 74 executable cases.
Reported counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 3, GATED 23, COVERAGE 9, HARDEST 3.
These counts describe that reviewed revision, not the later supplemented case count.
The complete report and reviewed file hashes are preserved privately. No rule, label,
ownership boundary or evidence gate changes in this response. Fixture changes require
a fresh second full review; the first report alone does not complete review.

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| DISCRIM: V35 missing B distance exactly 5 | Spec edit: V35(g) | B5 with no selector 7 result selects B5; an incorrect >=5 early return selects A20 |
| DISCRIM: V13 missing player roll exactly 80 | Spec edit: V13(b) | Seed {0,80} steps to {80,0}; roll 80 with reach 0 still accepts final 1, rejecting a >80 bypass |
| DISCRIM: V20/V28 all collision results 0 | Spec edit: V28(b) | Nearer valid T at distance 10 returns collision 1; existing slotA=A20 must survive. Ignoring rejection writes T10 |
| COVERAGE: R02 missing selected Skills record | Spec edit: V24(c) | Valid identity 20/count 400 with absent record returns lost 1/death before evaluation; this is distinct from invalid identity and missing instance |
| COVERAGE: R03a wrapper stored-target miss | Spec edit: V10(a) supplied wrapper stage | Absent stored/room/chooser result returns no unit,distance 2147483647,auxiliary 0. This does not implement or promote gated room/chooser producers |
| COVERAGE: R03c nonzero collision | Covered by V28(b) edit | Same assertion as blocking collision finding; no duplicate vector |
| COVERAGE: R03f missing units and town | Spec edit: V11(c..e) | Missing either unit returns 0; town invisible player returns 0 without advancing its candidate seed |
| COVERAGE: R03g exact roll 80 | Covered by V13(b) edit | Same assertion as blocking threshold finding |
| COVERAGE: R09a invalid selected missile | Spec edit: V34(e/f) | Negative and count-equal missile identities return 0 after latch set, without expression/creation request |
| COVERAGE: R10 B distance 5 | Covered by V35(g) edit | Same assertion as blocking chooser finding; actual chooser path integration stays gated |
| COVERAGE: R11a signed/wrapping speed | Spec edit: V36(e/f) | Explicit raw 1073741824 wraps its 75 product to -1073741824 before /100; negative level -9 tests truncation toward zero at /8. Both remain supplied arithmetic of an asm-only integrated rule |
| COVERAGE: R13 other types/missing records/negative size | Spec edit: V22(a..h) | Negative SizeX enlarges offsets; invalid monster extension/missile identity returns size 0; object/missile/item dispatch gives distinct size. Kind5/null size-only cases assert no null-position distance |
| HARDEST: R02 lazy init/decrement/owner failure | No rule change | Ordered signed gates and stored counter are explicit; V05/V17/V24..V26 plus added absent-record case provide exact state/effect fixtures |
| HARDEST: R03g opposite polarity/candidate RNG | No rule change | Distinct player/monster returns and seed owner are explicit; V12..V14/V23/V30 and new exact 80 case distinguish the difficult threshold |
| HARDEST: R09a attempts/stored point/signed frame versus u8 duration | No rule change | Existing source-backed gates, distinct current X700/stored X40 and -1/-256 frame cases remain; invalid missile additions exercise earlier exits |

MISMATCH, GUESS and GAP have no entries requiring action. The 23 documented GATED
entries remain unchanged: R01g,R03b/R03e/R03h,R05/R06b/R08/R10/R11a/R11b;
Q01/Q02/Q03/Q04/Q05/Q06/Q08/Q09/Q11/Q12/Q13; and supplied-stage V35/V36.
Pending gameplay, original data/compiled correspondence and producer contracts have
not been closed by synthetic cases. No query-count requirement was added: the new
cases discriminate returned values, stored slots/counters/latch, seed words or effect
requests, while an equivalent extra pure predicate read remains permitted.

## Blind-implementer FULL2 triage (2026-10-15)

The second full reader recomputed all 37 vector families and reported MISMATCH 0,
GUESS 0,GAP 0,DISCRIM 3,GATED 23,COVERAGE 8,HARDEST 3. No numeric case count was reported.
The coordinating owner retains the complete report privately. The outcomes below
change fixtures and clarify the initial state 12 input; rule semantics, evidence labels,
ownership and gates remain unchanged. A fresh third full review is required.

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| DISCRIM: V12/V13/V23/V30 constant reach results allow reversed arguments | Spec edit: V12 asymmetric ordered operation | reach(S,T,0)=0 but reach(T,S,0)=1; correct validity 0 versus swapped-argument validity 1. This is R03g's caller-order discriminator. V13 retains threshold bypass, V23 validity-before-cap seed effect and V30 absent-state gate; those vectors do not each claim to discriminate argument order |
| DISCRIM: V28(b) collision has no reverse result | Spec edit: V28(b) asymmetric ordered operation | collision(T,S,4)=1 but collision(S,T,4)=0. Correct slotA=A20 survives; swapped arguments replace it with T10. Extra pure reads remain equivalent if the resulting slot/seed assertions match |
| DISCRIM: V34 creator failure is last | Spec edit: V34 primary sequence | Count 5 adds index 4 success after index 3 failure. Assert creation requests 3 then 4, final return 1 and M initialization seed 44; stopping on failure omits the later observable request/effect |
| COVERAGE: R01c state 12 inactive | Input clarification, no separate case | The preamble now explicitly states S state 12 inactive unless overridden; ordinary think fixtures exercise that branch. V09 retains the active/removal branch. No new callback-count assertion is needed |
| COVERAGE: R01e chance<=0 or>99 | No change: already exercised | V32 supplies Hell threshold 0 and fails its roll; V01/V05/V17 use threshold 100 and pass their stated rolls. V04/V18 discriminate comparison equality. Exhaustive threshold/roll combinations are not claimed |
| COVERAGE: R03a nonzero auxiliary result | Spec edit: V10(b) | Stored target/distance 20 with ordered reach result 1 returns auxiliary 1; reversed result 0 would fail that scalar assertion |
| COVERAGE: R03f searcher unavailable | Spec edit: V11(f) | Death-mode searcher rejects invisible player T before its candidate draw; T seed remains unchanged |
| COVERAGE: R03i self-owner stop | Spec edit: V31(g) | Candidate owner resolves to itself and terminates, preserving T root and final 1 with existing tier 1/1 inputs |
| COVERAGE: R09a null caster distinct from null initializer | Spec edit: V34(g) reached handler only | Absent caster with valid supplied selector/position inputs and count 0 returns 1 without latch write or creation. This does not claim outer dispatcher accepts absent unit |
| COVERAGE: R10 B absent/A present | Spec edit: V35(h) | A20 and absent B selects A20, independently of the prior A-empty/B-present case; actual path integration remains gated |
| COVERAGE: R11a slow mask matches but state 87 inactive | Spec edit: V36(g) | Explicit raw 2560 with present list/value 50 stays 1920; applying slow solely from mask would produce 960 |
| HARDEST: R02 lazy budget/decrement/owner failure | No rule change | Existing ordered gates and persisted values remain explicit; V05/V17/V24..V26 cover initialization, missing record/instance and owner failure, including the second budget call |
| HARDEST: R03g opposite polarity/candidate RNG | No rule change | Existing exact threshold/polarity and seed fixtures remain; V12's new asymmetric results additionally discriminate caller argument order |
| HARDEST: R09a conditional attempts/stored point/frame narrowing | No rule change | Existing independent callback current X700/stored X40 and signed-frame cases remain; primary V34 now proves continuation after creator failure |

MISMATCH,GUESS and GAP contain no entries. The same 23 GATED entries remain:
R01g,R03b/R03e/R03h,R05/R06b/R08/R10/R11a/R11b;
Q01/Q02/Q03/Q04/Q05/Q06/Q08/Q09/Q11/Q12/Q13; and V35/V36.
None of these existing producer/observation limits was promoted by synthetic cases.
The five reported case-indentation defects were repaired without changing their
inputs, traces or expected outputs. All37 vector identities and the four observation
recipes remain; no source, table, gameplay or implementation test collection was added.

## Blind-implementer FULL3 triage (2026-10-15)

The third full reader examined 37 families and 110 cases, counting V15's five point
inputs as one bundled case. It reported MISMATCH 0,GUESS 0,GAP 1,DISCRIM 3,GATED 11,
COVERAGE 6,HARDEST 3. Its eleven GATED entries are the existing question ids Q01..Q06,
Q08,Q09,Q11..Q13; the changed reporting granularity does not remove any rule or vector
gate. The coordinating owner preserves the complete report privately. This response
changes output inventory and fixtures only; no rule semantics, ownership, labels or gates
change. A fresh DELTA4 can therefore check the changes after three full MISMATCH 0 runs.

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GAP: R09a path-preparation output missing from section 5 | Spec edit: section 5 burst-output row | Explicitly lists the already specified callback path-preparation request, without adding its unverified internal effects |
| DISCRIM: V07 equal caster/stored-point X | No numeric change; explicit coverage clarification | V34 M has current X700 and stored-pointX40,index 4; seed 44 rejects current-X seed 704. Independent callback index 2 gives 42 rather than 702. V07 pins burst count/supplied successful values and does not claim to discriminate equal source coordinates |
| DISCRIM: V19 equal caster/stored-point X | No numeric change; explicit coverage clarification | Same V34 distinct-input discriminator; V19 retains its original 1000/1001/1002 and cap 77 results. No additional duplicate seed-source case is needed |
| DISCRIM: V20 distance 49 cannot replace A20 even without cap | No numeric change; cite V06 | V06(a) empty slot rejects reduced 49; V06(b) empty slot accepts reduced 48. V20 now explicitly claims stable nearest ties, not independent cap discrimination |
| COVERAGE: R03c nearer collision-free candidate replaces occupied slot | Spec edit: V28(c) | Existing A20 is replaced by valid T10 with collision 0; always retaining the old slot fails. Reverse collision result 1 also preserves the caller-order distinction |
| COVERAGE: R03i missile-candidate owner-resolution branch | No change: premise corrected | R03i's missile owner branch is for the searcher; the candidate loop follows monster owners only. R03f rejects missile candidates on the sentry validation path. Shared missile-searcher resolution consumes PVP-R07, whose PVP-V16 covers failed/invalid missile owner links; no missile-candidate remapping is invented |
| COVERAGE: R07b dynamic candidate/missile getter | No change: owning getter already exercised | NRF-R10b and NRF-V29(g) give unsigned high-part X12/Y65535 from supplied 16.16 paths. SFR-R07b consumes that same type-independent getter contract; SFR-V33 checks sentry integer-part use, while R13/V22 cover candidate-size dispatch. No separate path representation or duplicate getter machinery is introduced |
| COVERAGE: R09a path preparation after initialization | Spec edit: V34 effect-order assertion | M seed overwrite 44 precedes path-kind 10,path-duration 77,then path-preparation request; independent seed 42 case has the same request order. Actual path-preparation internals remain Q12 |
| COVERAGE: R11a false slow-mask match with state/list present | Spec edit: V36(h) | Explicit raw 2560 with false mask,state 87 active,list/value 50 present stays 1920; ignoring mask would yield 960 |
| COVERAGE: R06a Nightmare field selection | No change: same indexed recipe | R06a states one difficulty-indexed signed-i16 selection over 0..2; V01/default 0 and V32/difficulty 2 distinguish selected fields and their signed use. Index 1 is an additional field binding of the same operation, not a separately inferred algorithm. Exhaustive data-population coverage is not claimed |
| HARDEST: R03g threshold/opposite polarity/candidate seed | No rule change | V12 asymmetric arguments,V13 exact 80,V14 opposite monster polarity and V23 persisted pre-cap seed make each distinct choice explicit |
| HARDEST: R03i asymmetric depth/self-stop/raw 8 | No rule change | PVP-R05a/R07 is the exact consumed contract; V31's directional 8,asymmetric chain and self-owner cases already distinguish these results |
| HARDEST: R09a continuation/stored point/signed-frame-u8 duration | No rule change | V34's later success after failure,distinct current/stored X and negative-frame cases remain exact; the new effect-order assertion finishes the output inventory without promoting path internals |

MISMATCH and GUESS contain no findings. GATED remains the same eleven questions above,
with all existing rule/vector integration limits intact. The new callback assertion is
an ordered subsequence of effect requests; it imposes no count/order requirement on
equivalent extra pure getters. Numerical outputs 42/44,10,77,1920 are unchanged except
for the newly supplied successful occupied-slot replacement T10 and false-mask case.

### DELTA4 change list

- Section5: add path-preparation request to the existing R09a output row.
- V07/V19/V20 Discriminates text only: cite existing V34/V06 cases; no input/output change.
- V28(c): new independent collision-free T10 replacement of existing A20, with reverse
  ordered collision result 1 and exact final slots/seeds.
- V34: add ordered callback effect subsequences for existing successful M/index 4 and
  independent index 2 initializers; seeds 44/42 and type 10/duration 77 remain unchanged.
- V36(h): new false-mask/state-active/list-present raw 2560 case, expected 1920.
- Companion: this complete triage and changed-material list. Gap file unchanged.

## Blind-implementer DELTA4 result (2026-10-15)

A fresh fourth reader checked all changed material listed above, confirmed the assigned three file hashes and found no rule, ownership, label or gate changes. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0. It independently recomputed V28(c), both strengthened V34 initializer assertions and V36(h), checked the section 5 output entry, and verified the V34/V06 discriminators cited by the V07/V19/V20 coverage clarifications. The reader reported this material by assertion groups rather than a numeric case count.

The V34 seeds 44/42 reject current-X-derived 704/702; V06 rejects 49 and accepts 48; V28(c) replaces A20 by T10 only with the ordered zero collision result; V34 requests seed overwrite, path kind 10, duration 77, then preparation; V36(h) skips the slow scale and yields 1920. FULL3's one GAP and three DISCRIM findings are resolved. The final applicable reader condition is met, with existing producer and observation limits unchanged and status still draft. No implementation test, gameplay observation, implementation plan or maintainer approval is implied. The complete report is retained privately.
