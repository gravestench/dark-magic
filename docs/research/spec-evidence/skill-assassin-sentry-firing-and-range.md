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
supplied dynamic positions, GDF-R03 for state 146 invis, and SCD/SCX/SLD for action dispatch,
calculation bounds and their precise unresolved producers. No universal highest-level
instance contract was found: selection/bonus production remains SFR-Q08, consistent with
SCX-Q07/SCD-Q10. No gameplay actions, new hooks or other-build analysis were performed.

## Dispatch and function records

| Artifact | Function or table RVA | Analyst description / rules | Method and actual boundary |
|---|---|---|---|
| LOD113C-D2GAME | 0xA7DC0 | reached sentry think, SFR-R01a..g/R06a | Original disassembly 2026-10-12; full entry-to-exit reread 2026-10-15 for parameter loads, ordered calls, second-budget return, event walk and skill-mode request. Actual AI101 entry remains unanchored |
| LOD113C-D2GAME | 0xA4A10 | budget, SFR-R02 | Original disassembly 2026-10-12; full reread 2026-10-15: owner/room early reject, signed lazy-init/counter tests, expression call, decrement and return. External producer effects remain Q08 |
| LOD113C-D2GAME | 0xF21E0; 0x75070; 0x75E10 | death-mode request construction | Reused 2026-10-12 disassembly: zero mode reaches request. Actual complete monster-mode execution is outside supplied budget stage |
| LOD113C-D2GAME | 0xF22C0; 0xF2380; 0x8E6F0 | delay, skill-mode, event-removal calls from think | Caller arguments/read 2026-10-15; full producer contracts are Q01/Q08, not verified by caller shape |
| LOD113C-D2GAME | 0xB1BC0 | stored-first target wrapper, SFR-R03a | Original 2026-10-12; complete 2026-10-15 entry/exit reading establishes shortcut, initial distances, chooser result, optional reach and distance writes |
| LOD113C-D2GAME | 0xB1820; jump table 0xB1A9C | stored-target resolver, SFR-R09/Q05 | Original 2026-10-12 direct-site audit; full listing assessed 2026-10-15. Four stored modes, hash buckets, selector 3 draw, temporary selector/state writes, selectors 5/6 and output validity identified; complete producers remain unverified |
| LOD113C-D2GAME | 0xB17B0; table 0xFB588 | selector dispatch, SFR-R03b | Full 2026-10-15 dispatch reading and original-image table cells: selector 6 `{kind0, callback0xB15E0}`, selector 7 `{kind0, callback0xB1580}`. Kind0 dispatches 0xB0720 |
| LOD113C-D2GAME | 0xB0720 | room walker, SFR-R03b | Complete listing assessed 2026-10-15: room lookup, adjacent-room-list request, per-room chain and early callback return. Encoded portions/room-list producer keep actual traversal unresolved |
| LOD113C-D2GAME | 0xB15E0 | two-slot candidate callback, SFR-R03c | Original 2026-10-12; complete 2026-10-15 reread: validity before distance/cap, slot split, strict improvement, collision 4 before writes, zero return |
| LOD113C-D2GAME | 0xB1580 | selector 7 callback | Complete 2026-10-15 reading: excludes the block's supplied unit, validates, caps at 48, requires class>1, strict improvement over the block's running best and collision 4 before writing the block's found unit/distance. The chooser supplies slot A's unit as the excluded unit (see the 2026-10-15 FULL9 triage); the earlier "slot B" wording was wrong |
| LOD113C-D2GAME | 0xB08A0 | candidate slot class | Full 2026-10-15 reading: player 14; bounded MonStats class index reads byte+0x4E; other/missing 0. Exact column name remains unknown; Align is +0x4C, so it is not substituted |
| LOD113C-D2GAME | 0xB1AB0 | chooser, SFR-R08 | Full 2026-10-15 reading and complete original-image decode of encoded instructions: player/empty slots, distance 5, temporary path setters/restoration, selector 7 and distance 20. Path helper effects remain Q05 |
| LOD113C-D2GAME | 0xB1400 | validity, SFR-R03e/g | Original 2026-10-12; full original-image decode 2026-10-15 confirms both opposite branch polarities and ordered state/candidate-seed checks |
| LOD113C-D2GAME | 0xDFA90 | unavailable condition | Full 2026-10-15 listing: null/`aliveExclusionCondition` or player mode 0/17, monster mode 0/12 returns 1; otherwise 0 |
| LOD113C-D2GAME | 0xE0A60 | final root decision, SFR-R03h | Original 2026-10-12, complete 2026-10-15 reread: one searcher owner step, repeated candidate-owner steps, self-root reject, player interaction 8 or inverted relation. Owner lookup0xE03A0 consumes PVP-R07 |
| LOD113C-D2GAME | 0xF02A0 | player interaction 8 request | Full listing assessed 2026-10-15: selected player relation entry flags masked by 8, or 0. Existing list read consumes PVP-R05a; actual relation production remains HOS-Q01 |
| LOD113C-D2GAME | 0xB08F0 | unit-to-point metric, SFR-R04 | Reused 2026-10-12 byte-equivalent C/disassembly: signed absolute offsets and octagonal truncation; supplied position boundary |
| LOD113C-D2GAME | 0xB0A20 | candidate-size distance, SFR-R10 | Reused 2026-10-12 disassembly: candidate in EBX, searcher stack argument; size subtraction on both axes then zero clamps and octagonal expression. Both room and stored callers supply this order |
| LOD113C-D2COMMON | #10589, 0x30F80; helpers0x013B0/0x01150 | size by kind | Reused import/hand decode 2026-10-12; new byte-equivalent C reconstruction read 2026-10-15 confirms player 2, signed monster byte 8, object nested dword+0xD0, missile byte+0x18A, item 1, other/null 0 |
| LOD113C-D2COMMON | 0x56C10; string0x8FA40 | MonStats2 SizeX identity | Reused 2026-10-12 field-descriptor disassembly and original string read: SizeX conversion byte at offset 8; MonStatsEx linkage through MonStats+0x18 |
| LOD113C-D2GAME | 0xB6280; 0xB5D50; 0xB6310 | summon context and Blade first-event contrast | Reused 2026-10-12 record: action 45 summon/mode update; Blade action 44 requests frame+1. This does not establish sentry first due frame |

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
| 0xA6F4 | #10666 /0x84D30 | root relation, inverted by final decision; new import resolution/read 2026-10-15, complete classification reader0x391F0 consumes the dated PVP-R05a record; actual tier-list production remains PVP-Q04 |
| 0xA808 | #10057 /0x3C390 | room-to-level predicate, reads level through0x2D320; current consumer does not guess wider flags |
| 0xA478 | #10331 /0x2FE10 | unit room; SCD's current import record |
| 0xA3F4 | #10630 /0x4F070 ->0x4E480 | skill-instance request; new thunk/export forwarding read; selection/bonuses remain SCX-Q07/SCD-Q10 |
| 0xA71E | #10589 /0x30F80 | size operation; original 2026-10-12 import record |
| 0x9454 | #10867 /0x34BB0 | dynamic integerX; NRF-R10b owns current getter contract |

### Semantic layout correspondence

These locations identify evidence; they are not implementation requirements.

| Semantic term | Original representation | Rules |
|---|---|---|
| difficulty | Game+0x6D u8 | R01/R06a |
| `aip1`..4 per difficulty | MonStats+0x56/+0x5C/+0x62/+0x68, each difficulty advances 2 bytes, signed i16 | R01/R06a |
| firing Skill1 | MonStats+0x170 signed i16 | R01/R02 |
| shot counter | AI-control+0x18 signed i32, initial-1 | R02 |
| `calc4` | Skills+0x144 | R02 |
| dynamic integer positions | Path+0x02/+0x06 u16 | R04/R07b/R10 |
| static positions | Path+0x0C/+0x10 i32 | R04/Q04 |
| seed | Unit+0x20/+0x24 | R01e/R03f/R09 |
| candidate eligibility A | Unit+0xC4 bit 2 | R03e/Q06 |
| `aliveExclusionCondition` | Unit+0xC6 bit 0 | R03e |
| owner reference | MonsterData+0x28 then owner link+0x28/+0x2C/+0x30 | R02/R03h |
| missile owner-resolution condition and identity | Unit+0xC8 bit 10; +0x94/+0x98 | R03h |
| stored-target selector/state | MonsterData+0x38/+0x34 | R09/Q05 |
| two slots | block+0/+4 unit/distanceA, +8/+12 unit/distanceB | R03c/R08 |

## Corrections and affected assertions

### State146 predicate conflict (SFR-Q07)

The 2026-10-12 R03f prose and V13(a) described a same-polarity hostility filter for
players and monsters. The same version's Q07 already contradicted that prose with
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
producer chains are not all verified, so SFR-Q07 now gates actual reach-result integration.
Local predicate handling is verified with supplied results. V11/V12/V21 retain their
numeric seed/roll outcomes; their old supplied 'hostile' condition becomes the precise
reach return. V13(a) now accepts invisible monster reach 0 when final passes; V13(c)
checks reach 1 rejection. No numeric fixture was fitted to gameplay observation.

The earlier R03g claim of 'no melee-distance code on this path' is therefore withdrawn:
the imported reach helper reads spatial separation. The separate fact that final
decision0xE0A60 has no distance test remains supported. It is not the unrelated
function0xDFA60; local assembly labels had previously made that distinction confusing.

### Think gates

Full think reading confirms that the second, decrementing budget call's return is tested
at RVA 0xA7ECF/0xA7ED1: nonzero stops before pending-event deletion/skill selection. Earlier pseudocode
omitted that return check. V23 supplies owner loss on the second call and leaves events
untouched; it supplements the retained V05/V16 successful sequences.

### Labels, units and draw totals

GDF-R03's dated state record establishes 146 invis; only the eligibility-conditionA name
remains SFR-Q06. NRF-R10b establishes supplied dynamic integer-part units and does not
establish every static/path producer. R07b is verified at that supplied getter boundary;
the original broader coordinate question remains Q04. Player range is enlarged by the
FULL size 2 per axis, monster by signed SizeX, not half a collision radius.

R03h and R09 lacked executable vectors in the older record. V29 and V33 now make their
supplied root decision and reached direct-site draw assertions executable. R09 consumes
NRF's exact RandMask(m)=newLo & (m-1), so argument 2 returns 0/1. Expression/calculation,
creation and mode operations can draw separately. No-direct-draw statements never claim
the total whole-think or creator draw count.

R03b/R03d/R03g/R08 and full speed/footprint/timing remain gated or asm-only at their
stated boundaries. Prior verified assertions are retained as local supplied-input
contracts where justified; complete imported producers are not promoted by relocation.
Unsubstantiated 'not found' labels are replaced by unresolved questions because no
complete six-rung absence record exists. All original evidence is preserved privately;
this companion retains usable original-artifact claims, rationale and dates.

## Gap assessment and search limits (2026-10-15)

| Original question / gap | Compatible evidence and targeted follow-up | Current disposition |
|---|---|---|
| Q01 first/post-shot timing; R05/R06b | Reused summon/action 44 contrast; reread reached think/budget and identified scheduler/mode requests | Still open: actual scheduling/mode-end producers and O01/O02 |
| Q02 AI101; R01g | New original-image absolute-pointer scan finds think pointer at D2Game RVA 0x10EDE0; no table-index consumer anchor established | Entry binding unresolved; not an absence claim |
| Q03 original MonStats/`calc4` values; R02/R06a | The 2026-10-15 complete Skills loader read is recorded in the [launch companion](skill-assassin-sentry-missile-launch.md#td-skills17); it includes the raw BoltSentry `calc4` text | aip/Skill1/`calc4` compiled/runtime association still open |
| Q04 coordinate/seed | NRF-R10a/b supplies getter units; NRF-R01/R02/R04 and original candidate-site record supply owner/modulus | Dynamic getter and candidate RNG settled; other position producers/extreme static domain remain |
| Other-build equivalence (question since removed) | Evidence policy permits only 1.13c; no other build read | Outside evidence build; no equivalence claim and no implementation gate |
| Q05 two slots; R03b/d/g/R08 | Full dispatcher/callback/class/chooser assessment; selector 6/7 original-image cells; encoded chooser corroborated | Supplied choices refined; real stored modes/room-list/path/collision/class-field producers remain |
| Q06 state/flag | GDF-R03 state 146 named invis; unavailable helper read | State name closed; candidate eligibilityA meaning/producer open |
| Candidate size (question since settled) | Reused original import/descriptor/metric/call-site evidence; size helper byte-equivalent C reconstruction corroborates type dispatch | Static size arithmetic retained; O03 still pending |
| Q07 polarity/hostility | Import resolution and complete b1400/326c0 image decode above | Opposite polarity/incorrect hostility interpretation corrected; actual range/separation/collision integration remains |
| R03h and R09 missing vectors | Reused local final decision and NRF reached draw contracts | V29/V33 close executable-fixture gap; complete producers still gated |
| All pending gameplay | No campaign authorized in migration; retained all four recipes | Q09 prevents runtime-confirmed status; not a local arithmetic gate |

No new six-rung search was claimed complete. The AI pointer search supplies a new data
lead, not the missing dispatch consumer or a verified identity 101. The room-class field
was checked against the existing GDF map (Align+0x4C is distinct); its name remains neutral.
The complete original-image selector 6/7 cells were read at0xFB5B8/0xFB5C0 and direct
dispatch instructions read through their callback selection; actual room/path producers
remain explicit missing boundaries rather than guessed 'nearest hostile' behavior.

## Review additions (2026-10-15)

A spec-standard review on 2026-10-15 reread these paths in original D2Game disassembly
and the original image. It changes no evidence label.

**AI table root for the think (narrows SFR-Q02).** D2Game RVA 0x9D0F0 selects an AI table
entry: when its special-index argument is nonzero (and, for 10..12, a D2Common test passes)
it returns the special table `0x10E668 + 16*index`. Otherwise it reads the monster record's
signed word at MonStats+0x1E, bounded to 0..147, and returns `0x10E788 + 16*index`,
defaulting to entry 0. The original-image cells of entry 101 at RVA 0x10EDD8 are
{0, 0x6FCC39B0, 0x6FCC7DC0, 0}: init RVA 0xA39B0 and think RVA 0xA7DC0. Entry 102 is
{0, 0x6FCC3990, 0x6FCC6B70, 0}. Installer 0x9D230 calls the entry's init field (+4) and
stores the think field (+8) in AI-control+4, using fallback 0x9CB00 when that field is null.
Executor 0x9D3B0 checks that pointer and calls it with game, unit and the parameter block.
This makes the earlier pointer-scan lead at 0x10EDE0 the think slot of table index 101.
Init 0xA39B0 writes AI-control+0x18 = -1, which is the counter that SFR-R02 lazily
initializes. It also copies a game field (+0xA8) to AI-control+0x14. Still open: the
MonStats column name for +0x1E and the sentry rows' values (Q03), the special-table
selector, and the event root that calls 0x9D3B0. The linked file lists no caller for it.
R01g therefore stays unresolved.

**Think skill-mode request shape (SFR-R01f).** The final call sequence pushes
(0, 0, target, Skill1) for D2Game 0xF2380 (`ret 0x10`, four stack arguments) and then the
skill instance for thunk 0xAE1A. That thunk resolves to D2Common #10136, RVA 0x4E180, a
null-safe getter of skill-node field +8, which is the skill's mode. Its result passes in
EAX to 0xF2380 with the unit and game in registers. The spec's ordered "select skill and
target, then request that skill's mode" therefore describes one compound request.
0xF2380 itself reaches current-skill set #11103 and mode request 0x95E10. The vectors' two
logged supplied operations are one decomposition of that request, not two original calls.

**Additional import identities** (resolved 2026-10-15 from D2Game thunks):

| D2Game thunk RVA | LOD113C-D2COMMON export / RVA | Caller role |
|---|---|---|
| 0xA3B8 | #10750 / 0x34B80 | dynamic integer Y for the distance operations (NRF-R10b) |
| 0xA68E | #10494 / 0x33CD0 | state test: think state 12, validity state 146 (RLI-R03) |
| 0xA3CA | #10945 / 0x34510 | state set-off request for state 12 (RLI-R03) |
| 0xA496 | #11103 / 0x303E0 | current-skill set/clear request on the missing-skill and owner-loss paths |
| 0xA400 | #10306 / 0x505C0 | bonus-inclusive skill level in R02 (argument 1); producer SFR-Q08 |
| 0xAC9A | #10786 / 0x51BF0 | calculation evaluation for `calc4`/`calc1` (SCX contract) |
| 0xAE1A | #10136 / 0x4E180 | skill-node mode getter for the R01f request |

Labels on these targets in the reconstruction are analyst labels and do not match these
roles, so roles come from the caller arguments only.

**SFR-R02 non-monster path (text clarification).** Budget entry 0xA4A10 sends an absent
unit, a non-monster unit, absent monster data, an absent owner reference, or an owner link
with a zero first field to the same clear-skill plus mode-0 path (0xA3A5C, returns 1) as a
missing owner, missing room or town room. The review added this to the R02 text. No vector
or expected output changed. The added sentence still needs a delta blind-reader check.

**Gap-file historical comparison.** The retained 2026-10-12 implementation comparison
says "hard candidate cap 49" and "retarget every think". The current contract rejects room
candidates with reduced distance above 48 (SFR-R03c, V06), and the stored-target shortcut
persists (SFR-R03a/R03d). The gap file now says that the spec governs.

**Reverse dependencies.** Blade Sentinel and Fury cites no SFR rule; its AI is table entry
102 (think 0xA6B70), not this think. The review moved it from "Depended on by" to Related
specs and corrected the Scope sentence. Death Sentry cites SFR-R02, SFR-R03a, SFR-R08 and
SFR-R10, but not the shared speed stage, so the reverse-dependency text now names those ids.

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
- **SFR-O03:** With the supplied or authenticated `aip4`=12 case, keep all eligibility,
  chance and size conditions matched. Place a player at raw offsets(13, 0)/(14, 0) and a
  monster with authenticated SizeX3 at(14, 3)/(15, 3), relative to the sentry. Record reduced
  distances, selection and shot outcome. Expected11/12 with only the 11 case passing
  strict range. Room-cap controls use reduced 48/49; do not conflate stored shortcut.

## Migration and review record

2026-10-15 migration preserves the existing shared-family file boundary. Searches for the
filename and SFR prefix in implementation/test paths and non-gap plans found no consumer;
sibling/gap citations alone do not make ids consumed. Practical ids are retained without
legacy aliases. All 24 original vector families were retained; symbolic cases now have
complete numerical inputs, seeds, ordered effects and exact outcomes. Additional cases
separate newly established gates and supplied dependency boundaries. Four
observation recipes remain; the assessment table accounts for all 11 original questions,
including the outside-build and settled candidate-size questions. The original spec/gap files and detailed
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
detector. SFR-R03h reuses this compatible contract; no repeated binary reading or
generic unread-root-relation gate is needed. HOS-Q01/PVP-Q04 still limit production
of relation/alignment storage, rather than the verified read of supplied existing lists.
The spatial reach import #11138/0x326C0 is a different helper and remains SFR-Q07.
V29 now uses realizable existing relation/tier inputs rather than an arbitrary nonzero
relation result; V29(f) discriminates asymmetric owner depth.

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
| DISCRIM: V32 missing B distance exactly 5 | Spec edit: V32(g) | B5 with no selector 7 result selects B5; an incorrect >=5 early return selects A20 |
| DISCRIM: V12 missing player roll exactly 80 | Spec edit: V12(b) | Seed {0,80} steps to {80,0}; roll 80 with reach 0 still accepts final 1, rejecting a >80 bypass |
| DISCRIM: V18/V26 all collision results 0 | Spec edit: V26(b) | Nearer valid T at distance 10 returns collision 1; existing slotA=A20 must survive. Ignoring rejection writes T10 |
| COVERAGE: R02 missing selected Skills record | Spec edit: V22(c) | Valid identity 20/count 400 with absent record returns lost 1/death before evaluation; this is distinct from invalid identity and missing instance |
| COVERAGE: R03a wrapper stored-target miss | Spec edit: V09(a) supplied wrapper stage | Absent stored/room/chooser result returns no unit,distance 2147483647,auxiliary 0. This does not implement or promote gated room/chooser producers |
| COVERAGE: R03c nonzero collision | Covered by V26(b) edit | Same assertion as blocking collision finding; no duplicate vector |
| COVERAGE: R03e missing units and town | Spec edit: V10(c..e) | Missing either unit returns 0; town invisible player returns 0 without advancing its candidate seed |
| COVERAGE: R03f exact roll 80 | Covered by V12(b) edit | Same assertion as blocking threshold finding |
| COVERAGE: R08 B distance 5 | Covered by V32(g) edit | Same assertion as blocking chooser finding; actual chooser path integration stays gated |
| COVERAGE: R10 other types/missing records/negative size | Spec edit: V20(a..h) | Negative SizeX enlarges offsets; invalid monster extension/missile identity returns size 0; object/missile/item dispatch gives distinct size. Kind5/null size-only cases assert no null-position distance |
| HARDEST: R02 lazy init/decrement/owner failure | No rule change | Ordered signed gates and stored counter are explicit; V05/V16/V22..V24 plus added absent-record case provide exact state/effect fixtures |
| HARDEST: R03f opposite polarity/candidate RNG | No rule change | Distinct player/monster returns and seed owner are explicit; V11..V13/V21/V28 and new exact 80 case distinguish the difficult threshold |

MISMATCH, GUESS and GAP have no entries requiring action. The 23 documented GATED
entries remain unchanged: R01g,R03b/R03d/R03g,R05/R06b/STT-R05/R08/STT-R03/STT-R04;
Q01/Q02/STT-Q01/Q03/STT-Q02/STT-Q03/Q04/Q05/Q06/Q07/Q08/STT-Q04/Q09/STT-Q05; and supplied-stage V32/STT-V04.
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
| DISCRIM: V11/V12/V21/V28 constant reach results allow reversed arguments | Spec edit: V11 asymmetric ordered operation | reach(S,T,0)=0 but reach(T,S,0)=1; correct validity 0 versus swapped-argument validity 1. This is R03f's caller-order discriminator. V12 retains threshold bypass, V21 validity-before-cap seed effect and V28 absent-state gate; those vectors do not each claim to discriminate argument order |
| DISCRIM: V26(b) collision has no reverse result | Spec edit: V26(b) asymmetric ordered operation | collision(T,S,4)=1 but collision(S,T,4)=0. Correct slotA=A20 survives; swapped arguments replace it with T10. Extra pure reads remain equivalent if the resulting slot/seed assertions match |
| COVERAGE: R01c state 12 inactive | Input clarification, no separate case | The preamble now explicitly states S state 12 inactive unless overridden; ordinary think fixtures exercise that branch. V08 retains the active/removal branch. No new callback-count assertion is needed |
| COVERAGE: R01e chance<=0 or>99 | No change: already exercised | V30 supplies Hell threshold 0 and fails its roll; V01/V05/V16 use threshold 100 and pass their stated rolls. V04/V17 discriminate comparison equality. Exhaustive threshold/roll combinations are not claimed |
| COVERAGE: R03a nonzero auxiliary result | Spec edit: V09(b) | Stored target/distance 20 with ordered reach result 1 returns auxiliary 1; reversed result 0 would fail that scalar assertion |
| COVERAGE: R03e searcher unavailable | Spec edit: V10(f) | Death-mode searcher rejects invisible player T before its candidate draw; T seed remains unchanged |
| COVERAGE: R03h self-owner stop | Spec edit: V29(g) | Candidate owner resolves to itself and terminates, preserving T root and final 1 with existing tier 1/1 inputs |
| COVERAGE: R08 B absent/A present | Spec edit: V32(h) | A20 and absent B selects A20, independently of the prior A-empty/B-present case; actual path integration remains gated |
| HARDEST: R02 lazy budget/decrement/owner failure | No rule change | Existing ordered gates and persisted values remain explicit; V05/V16/V22..V24 cover initialization, missing record/instance and owner failure, including the second budget call |
| HARDEST: R03f opposite polarity/candidate RNG | No rule change | Existing exact threshold/polarity and seed fixtures remain; V11's new asymmetric results additionally discriminate caller argument order |

MISMATCH,GUESS and GAP contain no entries. The same 23 GATED entries remain:
R01g,R03b/R03d/R03g,R05/R06b/STT-R05/R08/STT-R03/STT-R04;
Q01/Q02/STT-Q01/Q03/STT-Q02/STT-Q03/Q04/Q05/Q06/Q07/Q08/STT-Q04/Q09/STT-Q05; and V32/STT-V04.
None of these existing producer/observation limits was promoted by synthetic cases.
The five reported case-indentation defects were repaired without changing their
inputs, traces or expected outputs. All37 vector identities and the four observation
recipes remain; no source, table, gameplay or implementation test collection was added.

## Blind-implementer FULL3 triage (2026-10-15)

The third full reader examined 37 families and 110 cases, counting V14's five point
inputs as one bundled case. It reported MISMATCH 0,GUESS 0,GAP 1,DISCRIM 3,GATED 11,
COVERAGE 6,HARDEST 3. Its eleven GATED entries are the existing questions (now SFR-Q01..Q09 and
STT-Q01..Q05); the changed reporting granularity does not remove any rule or vector
gate. The coordinating owner preserves the complete report privately. This response
changes output inventory and fixtures only; no rule semantics, ownership, labels or gates
change. A fresh DELTA4 can therefore check the changes after three full MISMATCH 0 runs.

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| DISCRIM: V18 distance 49 cannot replace A20 even without cap | No numeric change; cite V06 | V06(a) empty slot rejects reduced 49; V06(b) empty slot accepts reduced 48. V18 now explicitly claims stable nearest ties, not independent cap discrimination |
| COVERAGE: R03c nearer collision-free candidate replaces occupied slot | Spec edit: V26(c) | Existing A20 is replaced by valid T10 with collision 0; always retaining the old slot fails. Reverse collision result 1 also preserves the caller-order distinction |
| COVERAGE: R03h missile-candidate owner-resolution branch | No change: premise corrected | R03h's missile owner branch is for the searcher; the candidate loop follows monster owners only. R03e rejects missile candidates on the sentry validation path. Shared missile-searcher resolution consumes PVP-R07, whose PVP-V15 covers failed/invalid missile owner links; no missile-candidate remapping is invented |
| COVERAGE: R07b dynamic candidate/missile getter | No change: owning getter already exercised | NRF-R10b and NRF-V15(g) give unsigned high-part X12/Y65535 from supplied 16.16 paths. SFR-R07b consumes that same type-independent getter contract; SFR-V31 checks sentry integer-part use, while R10/V20 cover candidate-size dispatch. No separate path representation or duplicate getter machinery is introduced |
| COVERAGE: R06a Nightmare field selection | No change: same indexed recipe | R06a states one difficulty-indexed signed-i16 selection over 0..2; V01/default 0 and V30/difficulty 2 distinguish selected fields and their signed use. Index 1 is an additional field binding of the same operation, not a separately inferred algorithm. Exhaustive data-population coverage is not claimed |
| HARDEST: R03f threshold/opposite polarity/candidate seed | No rule change | V11 asymmetric arguments,V12 exact 80,V13 opposite monster polarity and V21 persisted pre-cap seed make each distinct choice explicit |
| HARDEST: R03h asymmetric depth/self-stop/raw 8 | No rule change | PVP-R05a/R07 is the exact consumed contract; V29's directional 8,asymmetric chain and self-owner cases already distinguish these results |

MISMATCH and GUESS contain no findings. GATED remains the same eleven questions above,
with all existing rule/vector integration limits intact. The new callback assertion is
an ordered subsequence of effect requests; it imposes no count/order requirement on
equivalent extra pure getters. Numerical outputs 42/44,10,77,1920 are unchanged except
for the newly supplied successful occupied-slot replacement T10 and false-mask case.

### DELTA4 change list

- Section5: add path-preparation request to the existing STT-R02 output row.
- STT-V01/STT-V02/V18 Discriminates text only: cite existing STT-V03/V06 cases; no input/output change.
- V26(c): new independent collision-free T10 replacement of existing A20, with reverse
  ordered collision result 1 and exact final slots/seeds.
- STT-V03: add ordered callback effect subsequences for existing successful M/index 4 and
  independent index 2 initializers; seeds 44/42 and type 10/duration 77 remain unchanged.
- STT-V04(h): new false-mask/state-active/list-present raw 2560 case, expected 1920.
- Companion: this complete triage and changed-material list. Gap file unchanged.

## Blind-implementer DELTA4 result (2026-10-15)

A fresh fourth reader checked all changed material listed above, confirmed the assigned three file hashes and found no rule, ownership, label or gate changes. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0. It independently recomputed V26(c), both strengthened STT-V03 initializer assertions and STT-V04(h), checked the section 5 output entry, and verified the STT-V03/V06 discriminators cited by the STT-V01/STT-V02/V18 coverage clarifications. The reader reported this material by assertion groups rather than a numeric case count.

The STT-V03 seeds 44/42 reject current-X-derived 704/702; V06 rejects 49 and accepts 48; V26(c) replaces A20 by T10 only with the ordered zero collision result; STT-V03 requests seed overwrite, path kind 10, duration 77, then preparation; STT-V04(h) skips the slow scale and yields 1920. FULL3's one GAP and three DISCRIM findings are resolved. The final applicable reader condition is met, with existing producer and observation limits unchanged and status still draft. No implementation test, gameplay observation, implementation plan or maintainer approval is implied. The complete report is retained privately.

## Blind-implementer DELTA5 triage (2026-10-15)

Run type: delta check (FULL1..FULL3 and DELTA4 precede it). Changed material checked: the
review's R02 early-reject sentence, the Q02 text, the Scope sentence, the section 9/10
dependency lists and the merged STT-V03 Outputs bullet. Reported counts: MISMATCH 0, GUESS 5,
GAP 4 (one stating no consistency gap), DISCRIM 3, GATED 6, COVERAGE 5, HARDEST 3. The reader
recomputed V05, V16, V22, V23, V24 and STT-V03 with no difference and accepted changes 2..5.

Asm check before editing (original D2Game disassembly, 2026-10-15): budget 0xA4A10 tests,
in order, unit null, unit kind 1, MonsterData (+0x14) null, owner reference (+0x28) null,
owner link first field zero, then calls owner lookup 0xE03A0, room #10331 and town predicate
#10057; any failure reaches 0xA4A5C, which calls #11103(unit, 0) and mode request 0x95E10 with
the unit in the request block, including a null unit. The lazy-init and counter<=0 rejects
call 0xF21E0 with mode 0; its builder 0x95070 calls the same #11103 import (thunk 0xA496) with
the caller's unit before filling the mode-0 request. Both reject families therefore issue
current-skill clear then mode 0. They differ in one internal request-block byte (0x64 versus
0x64/0x65 by 0x94F20); that field's effect stays with Q08 and is not in the contract. The
counter test is signed `>=0` (any negative initializes); Skill1 is sign-extended and rejected
when <0 or >=count; the level request #10306 passes include-bonus 1; evaluation #10786 takes
(unit, `calc4`, Skill1, level).

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GUESS: R02 death request target for absent unit | Spec edit: R02 | Requests are issued with the supplied unit, absent included; their absent-unit handling is Q08. V24(h) |
| GUESS: R02 early-reject order / owner lookup when reference missing | Spec edit: R02 ordered gate list | Order read in asm; absent reference makes no owner lookup request. V24(h/i/j) |
| GUESS: R02 death request versus skill clear | Spec edit: R02 defines one death request (clear then mode 0) for every reject; V22/V24 traces state it | Asm: 0xF21E0's builder calls the same clear-skill import |
| GUESS: V24(c) owner absent meaning | Spec edit: V24(c) | Owner reference present, lookup returns absent; one lookup request asserted |
| GUESS: R02 bonus-inclusive level | Spec edit: R02 wording, V24(f) | Base 2 versus bonus-inclusive 3; evaluation (S, 20, 3) |
| GAP: section 3 lacks owner reference / town input | Spec edit: new section 3 row | Effects already in section 5 (current-skill clear/death-mode requests, lost) |
| GAP: no vector for new early-reject branches | Spec edit: V24(h/i/j) | Absent unit, player unit, missing reference |
| GAP: Q02 versus R01g wording | Spec edit: R01g text | States entry 101 holds the think; selection and event root unresolved. Label unchanged |
| GAP: consistency of changes 2..4 | No change | Reader found none |
| DISCRIM: V05/V16/V24 level or argument order stubbed | Spec edit: V24(f) | Rejects base level 2 and reordered evaluation arguments |
| DISCRIM: V22/V24 reject without skill clear | Spec edit: V22/V24 traces | Each mode 0 output is the ordered clear-then-mode-0 request pair |
| DISCRIM: V24(c..e) counter read before owner check | Spec edit: V24(g) | A pure early read is equivalent; lazy initialization before the owner gate is rejected by counter -1 kept and no evaluation |
| COVERAGE: R02 absent unit / non-monster / missing reference | Spec edit: V24(h/i/j) | Same as GAP above |
| COVERAGE: R02 negative counter other than -1 | Spec edit: V24(k) | -5 initializes to 2, decrements to 1 |
| COVERAGE: R02 negative Skill1 | Spec edit: V22(d) | Skill1 -1 rejects before instance request |
| HARDEST: missing reference versus lookup absent | Resolved by R02 order and V24(c/j) | |
| HARDEST: absent-unit request target | Resolved by R02 text and V24(h) | |
| HARDEST: death request includes skill clear | Resolved by R02 definition | |

GATED (R01g, Q01, Q02, Q03, Q05, Q08) is unchanged. These edits make R02's existing
binary behavior explicit and add cases; they change no label, ownership or gate. A fresh
DELTA6 checks them.

## Blind-implementer DELTA6 triage (2026-10-15)

Run type: delta check of the DELTA5 edits. Counts: MISMATCH 0, GUESS 2, GAP 3, DISCRIM 4,
GATED 3, COVERAGE 5, HARDEST 3. It recomputed V05, V16, V22(a..d), V23 and V24(a..k) with
no difference and confirmed every DELTA5 finding resolved except three partial ones listed
below. The asm reading above (Skills record check precedes the instance request; budget
counter lives in the AI parameter block, not the unit) supports the edits.

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GUESS: V24(h) counter of an absent unit | Spec edit: V24(h) | Counter belongs to supplied AI parameters; 3 unchanged |
| GUESS: V22(c) instance request with absent record | Spec edit: V22(c) | No instance, level or evaluation request (record test precedes instance request) |
| GAP: V24(h)/V23 absent-unit request log | No change beyond V24(h) wording | Only issue/order of the two requests is asserted; their handling is Q08 |
| GAP: V23 lacks death-request note | Spec edit: V23 trace | Clear then mode 0 for S |
| GAP: V24(i) owner reference unspecified | Spec edit: V24(i) | Present reference with succeeding lookup supplied |
| DISCRIM: V24(i) reference before kind | Spec edit: V24(i) | Lookup would succeed, so a reference-first order logs a lookup; asserted none |
| DISCRIM: V24(d/e) lazy init before room/town gates | Spec edit: V24(l/m) | Counter -1 kept, no instance/level/evaluation request |
| DISCRIM: V23 reject without clear | Spec edit: V23 trace | Ordered pair asserted |
| DISCRIM: V05/V16 level/argument order | No change | V24(f) is R02's discriminating case; V05/V16 claim think sequencing, not level selection |
| COVERAGE: Skill1 0 / count-1 | No change (advisory) | V22(a) rejects count, V05 accepts 20; not added |
| COVERAGE: missile or other non-monster kind | No change | Same kind!=monster branch as V24(i) |
| COVERAGE: decrement=true on reject paths | No change | Rejects precede the decrement; V23 already rejects with decrement=true |
| COVERAGE: lazy result 1 with decrement | No change | V24(k) covers persist-then-decrement arithmetic |
| COVERAGE: INT32_MIN counter | No change | Signed <0 test is covered by -1 and -5; no different branch |
| HARDEST: gate order | Resolved by V24(c/g/h/i/j/l/m) | |
| HARDEST: lazy init position | Resolved by V24(g/l/m) | |
| HARDEST: V24(h) log | Resolved by V24(h) wording | |

A fresh DELTA7 checks these edits.

## Blind-implementer DELTA7 triage (2026-10-15)

Run type: delta check. Counts: MISMATCH 0, GUESS 0, GAP 2, DISCRIM 3, GATED 3, COVERAGE 4,
HARDEST 3. All DELTA6 findings were confirmed resolved except the V24(i) claim wording.
Asm check: think 0xA7ECA calls the decrementing budget; nonzero at 0xA7ED1 jumps to the
epilogue 0xA7F47 and returns, with no schedule request. Budget 0xA4A10 branches on a null
owner-lookup result before the room request #10331.

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GAP: V23 call-sequence owner stub | Spec edit: V23 inputs | Lookup returns P on first call, absent on second |
| GAP: V24(c) room request after absent owner | Spec edit: V24(c) | No room request asserted |
| DISCRIM: V24(i) reference-before-kind claim unobservable | Spec edit: V24 Discriminates | Claim now names the lookup-before-kind alternative; pure reject order stated as unasserted and equivalent |
| DISCRIM: V23 delay request after reject | Spec edit: V23 outputs | No next-think delay request |
| DISCRIM: V22(a) instance request before bounds | Spec edit: V22(a) | No instance/level/evaluation request; record read is a pure getter |
| COVERAGE: Skill1 0/count-1; instance-absent with decrement; room-versus-town order | No change (advisory) | Same branches as existing cases or unobservable equivalents |
| COVERAGE: positive counter decrement | No change | V05/V16 cover it |
| HARDEST: pure-reject order | Resolved by Discriminates wording | |
| HARDEST: V23 stateful stub | Resolved by V23 inputs | |
| HARDEST: V24(h) no counter read | No change | R02 reads the counter only after all owner gates pass; (h) states it unchanged |

A fresh DELTA8 (the fourth reader run of this cycle) checks these edits.

## Blind-implementer DELTA8 result (2026-10-15)

Run type: delta check. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 3 (R01g/Q02,
Q08, Q03), COVERAGE 5, HARDEST 3. It recomputed V22(a..d), V23 and V24(a..m) with no
difference and confirmed all five DELTA7 findings resolved. Outcomes of the remaining entries:

| Finding | Outcome | Reason |
|---|---|---|
| COVERAGE: accepted Skill1 0 / count-1 | No change (advisory) | V22(a) rejects count; default 20 accepts; same comparison branch |
| COVERAGE: missile/item kind with owner reference | No change | Same kind-not-monster branch as V24(i) |
| COVERAGE: positive counter decrement; counter 1 to 0 | No change | V05/V16 cover both |
| COVERAGE: decrement wrap at INT32_MIN+1 | No change | Decrement runs only when counter>0, so no wrap is reachable |
| HARDEST: V23 call-sequence stub | No change | Inputs state the per-call results; R02 makes one lookup per call |
| HARDEST: V24(f) ordered subsequence | No change | Stated as subsequence with owner/room requests omitted |
| HARDEST: V22(b)/(c) versus R01b | No change | R02 and R01b order already explicit (R01a runs R02 first) |

The cosmetic V22 case order (a, d, b, c) is left as is to keep case ids stable. The final
applicable reader run meets the stop condition. Reader runs this cycle: four, all delta
(DELTA5..DELTA8). Status stays draft; no implementation test, gameplay observation or
approval is implied.

## Blind-implementer FULL9 triage (2026-10-15)

Run type: FULL (checklist section 2 steps 1-9). It was required because the 2026-10-15
review rewrote SFR-R02's behavior text and DELTA5..DELTA8 checked it only as deltas. This is
the first reader run of this follow-up cycle. Counts: MISMATCH 0, GUESS 12, GAP 9, DISCRIM 11,
GATED 22, COVERAGE 24, HARDEST 3. The reader recomputed every vector, including V22, V23 and
V24(a..m), and every LCG chain without difference. The complete report is retained privately.

Asm checks before editing (original D2Game disassembly, 2026-10-15):
- Wrapper 0xB1BC0: when the returned unit (stored or chooser path) is present it calls reach
  #11138(searcher, target, 0) into the auxiliary out-pointer; when absent it writes 0 there and
  distance 0x7FFFFFFF, with no reach call.
- Chooser 0xB1AB0 (encoded instructions decoded from the db listing): its arguments are
  (&slotA unit, &slotA distance, slotB unit, slotB distance), where slot A is the class>1 slot
  that 0xB15E0 writes at block+0. Order: present player searcher returns 0 (A, even if empty);
  slot B unit null returns 0; slot A unit null returns 1 (B); signed B distance >5 returns 0.
  Otherwise it saves path fields, sets the path target to slot A's unit (#10336), path kind 2,
  prepares the path (#10334), restores target, point (#10286) and kind, then reads #10800
  (D2Common 0x34C60, path field +0x28); nonzero returns 0 (A). For zero it runs selector 7 with
  block {found 0, distance 0x7FFFFFFF, excluded = slot A's unit}; a found unit with signed
  distance <20 replaces slot A and returns 0; otherwise it returns 1 (B). The previous spec text
  "excluding B" and the companion's 0xB1580 "slot B" wording were wrong; selector 7 improves on
  its own running best, not on slot A's distance. The meaning of path field +0x28 is not
  established, so R08 stays unresolved under Q05.
- Validity 0xB1400: searcher null/kind (0..1), then candidate null/kind; unavailable candidate
  then searcher; candidate room #10331 passed unchecked to the town predicate #10057; then
  eligibility A; then the state 146 branches as already specified.

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GUESS: V01/V05/V16/R01b think instance request missing from V01 log | Spec edit: V01 outputs | The ordered request subsequence now includes the selected-instance request; owner/room requests omitted |
| GUESS: R08 fixed 20 versus A distance | Spec edit: R08 | Fixed signed 20 |
| GUESS: R08 player searcher with empty A | Spec edit: R08; V32(k) | A player selects A even when empty: no unit, distance 2147483647 |
| GUESS: R08 whether selector 7 may return A, and nearer than what | Spec edit: R08; V32(j) | Excludes A's unit; running best starts at 2147483647. Also adds the path-result branch, V32(i) |
| GUESS: R03a auxiliary result when absent or on the room path | Spec edit: R03a; V09(d) | Reach on any returned target; absent gives 0 with no request |
| GUESS: R03e searcher kind | Spec edit: R03e | Both units must be player/monster; searcher checked first (pure order not asserted) |
| GUESS: R03e absent candidate room | Spec edit: R03e | The absent room is passed to the town predicate; its handling is Q08 |
| GUESS: R01f requested mode value | Spec edit: R01f | Mode value supplied by the selected skill instance; execution stays Q08 |
| GAP: outputs not in section 5 | Spec edit: section 5 | Intermediate values and request presence/absence/count are named as assertions, not stored effects |
| GAP: V23 call boundary | Spec edit: V23 inputs | The search is supplied; the only owner lookups are one per R02 call |
| GAP: reach "other arguments" given per vector | No change | Each vector states its ordered operation; the preamble default is reach 1 for any arguments |
| GAP: V29(f) Q owner and tier | Spec edit: V29(f) | Q is a monster with no owner link; other tiers follow the preamble |
| GAP: V30 Nightmare and `aip3`/`aip4` difficulty | Spec edit: V30(a/b) | Hell `aip4` 8 / `aip3` 9 gives due 109; Nightmare `aip1` 0 / `aip2` 5 gives due 105 |
| GAP: V01 post-shot mode and skill state | No change | Actual mode/skill execution is Q08; supplied effects log the shot only |
| DISCRIM: V01 removing all events | Spec edit: V01 | Events [2, 8] become [8] |
| DISCRIM: V09/V25 cap applied to stored path | Spec edit: V09(c) | Stored distance 60 returns T at the wrapper stage |
| DISCRIM: V11/V21 always rejecting roll<80 | Spec edit: V11(a) | Roll 51 with reach(S,T,0)=1 gives validity 1 |
| DISCRIM: V10 eligibility tested after draw | Spec edit: V10(g) | Invisible player with A=false; seed unchanged |
| DISCRIM: V30 `aip3`/`aip4` difficulty | Spec edit: V30(a) | Hell 8 rejects distance 10 |
| DISCRIM: V29(f) return value alone | No change | The vector asserts roots M/Q and no lookup of M's owner; these reject one-hop and repeated alternatives |
| DISCRIM: R02 Skill1 bounds | Spec edit: V24(n/o) | Skill1 0 and 399 accepted |
| DISCRIM: V33 Rand versus RandMask | No change: equivalent | Both step once and return lo&1 for argument 2 |
| HARDEST: R02 | No rule change | Order, death request and the logged-versus-pure boundary are explicit; V23 boundary clarified |
| HARDEST: R03f/R03h with PVP preamble | No rule change | V11(a) added; PVP-R05a/R07 remains the owning contract |

COVERAGE (24, advisory): new cases cover R06a Nightmare (V30(b)), STT-R02 f=78 and f>=256
(STT-V03(h/i)) and R02 Skill1 0/399 (V24(n/o)). The others get no change: they repeat branches
already shown (state 12 with a shot, slot B replacement, R03h tier combinations owned by PVP
vectors, octagonal wrap outside the stated domain, roll 79, missile Size>0 on both axes) or
lie inside gated STT-R03/Q08 boundaries.

R08's rule text changed (behavior of a gated rule), and the STT-R02, R03a, R03e and STT-R03 text was
made explicit. The next run is therefore FULL.

## Blind-implementer FULL10 triage (2026-10-15)

Run type: FULL, the second reader run of this follow-up cycle, on the FULL9-edited contract.
Counts: MISMATCH 0, GUESS 8, GAP 8, DISCRIM 14, GATED 31, COVERAGE 22, HARDEST 3. The reader
recomputed every vector, including all new FULL9 cases, without difference. It reported the
GATED list at finer granularity (one entry per rule/question pair); no gate changed.

Asm checks before editing (original D2Game disassembly, 2026-10-15):
- Unavailable condition 0xDFA90: null or completion bit returns 1; a player returns 1 for mode 0
  or 0x11; a monster for mode 0 or 0x0C; other kinds return 1. A player in mode 12 or a monster
  in mode 17 is therefore available.
- Validity, size and parameter readings already recorded support the remaining added cases
  (signed `aip1`/`aip4` loads, u8 missile Size, candidate-only size subtraction).

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GUESS: R01f mode value | Spec edit: section 3 row, preamble, V01 | Supplied instance mode value 9; V01 asserts mode request 9 |
| GUESS: V29 alignment to tier | No change: owned by PVP-R05a | PVP-R05a's tier reader (PVP-V24) defines the read; the preamble supplies stat 172 values |
| GUESS: think raw offset versus supplied distance | Spec edit: preamble | Think consumes the supplied distance; raw offset plus reduced value is a separate isolated R10 assertion |
| GUESS: V26(a) monster candidate inputs | Spec edit: preamble | Monster candidates share the ordinary inputs, no owner, SizeX 0, class 1 unless stated |
| GUESS: R03e dead modes per kind | Spec edit: R03e; V27(e/f/g) | Player 0/17, monster 0/12 |
| GUESS: R03h missile-searcher hop | No change: owned by PVP-R07 | PVP-V15 covers the missile owner link |
| GAP: `aliveExclusionCondition` source | No change | Section 3 already lists it as a supplied neutral boolean |
| GAP: R01f mode value input | Spec edit: section 3 row | Same as the GUESS above |
| GAP: stat 172 to tier | No change: owned by PVP-R05a | As above |
| GAP: V26(a)/V32 monster candidate inputs | Spec edit: preamble | As above |
| GAP: V32 selector 7 result inputs | Spec edit: V32 trace | Named result C is supplied directly; filtering stays gated |
| GAP: section 5 request arguments | Spec edit: section 5 rows | Select (Skill1, target, (0, 0)), mode value, resolution and creation requests with arguments |
| DISCRIM: V27 dead modes per kind | Spec edit: V27(e/f/g); V27 added to the root-default list | Player 12 and monster 17 valid; player 0 invalid |
| DISCRIM: V10 kind tests | Spec edit: V10(h/i) | Missile candidate and object searcher rejected |
| DISCRIM: R03c class threshold | Spec edit: V26(d) | Class 2 goes to slot A, class 0 to slot B |
| DISCRIM: think vectors and R10 variants | Spec edit: preamble | Each stated raw offset is an isolated R10 assertion, so half-size or Euclidean R10 fails V03/V15/V19 |
| DISCRIM: V32 fixed 20 | Spec edit: V32(l/m) | A30/C25 selects B; A10/C15 selects C |
| DISCRIM: R10 missile Size signedness | Spec edit: V20(i) | Size 200 gives distance 0, not 92 |
| DISCRIM: R10 searcher size | Spec edit: V20(j) | Searcher SizeX 5 ignored; distance 11, not 6 |
| DISCRIM: V08 state 12 removal before budget | Spec edit: V08(a) | Counter 0 dies with state 12 still active |
| DISCRIM: V01 mode and point arguments | Spec edit: V01 | select(20, T, (0, 0)); mode 9 |
| DISCRIM: V07 request order | Spec edit: V07 | Clear then mode 0 |
| DISCRIM: signed `aip1`/`aip4` | Spec edit: V30(c/d) | -1 values give due 111 and 107 |
| HARDEST: R08/V32 | No rule change | The path result stays a neutral supplied value under Q05; V32 asserts supplied branches only |
| HARDEST: R03h/V29 | No change: owned by PVP-R05a/R07 | The preamble supplies existing tier inputs |

COVERAGE (22, advisory): V10(h/i), V27(e..g) and V26(d) also cover the searcher-kind,
candidate-kind, dead-mode and class-0/2 branches. The remaining entries repeat branches shown by
existing cases or sit inside owned or gated contracts (PVP-R05a/R07 tier and relation
combinations, R08 selector-7 filters, STT-R03 combinations, NRF-R10b getter high-bit handling,
R04 negative differences already covered by absolute values in V14's domain); no change.

R03e's dead-mode wording, STT-R02's request arguments and the vector preamble changed, so the
next run is FULL.

## Blind-implementer FULL11 triage (2026-10-15)

Run type: FULL, the third reader run of this follow-up cycle, on the FULL10-edited contract.
Counts: MISMATCH 0, GUESS 8, GAP 7, DISCRIM 6, GATED 21, COVERAGE 24, HARDEST 3. Every vector
recomputed without difference.

Asm checks before editing (original disassembly, 2026-10-15):
- Think 0xA7DC0: Skill1 is sign-extended and passed to the instance request #10630 with no bounds
  check; state 12 active calls #10945(unit, 12, 0); the event walk visits every pending event and
  removes each one whose type byte is 2. The think's returned value is not part of this contract.

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GUESS: R01b Skill1 bounds when counter>=0 | Spec edit: R01b | Raw identity, no bounds check |
| GUESS: R01c removal request form | Spec edit: R01c | Set state 12 off (RLI-R03); no request when inactive |
| GAP: think return value | Spec edit: R01f | No result value is part of the contract |
| GAP: R04 consumer | Spec edit: R04 | Specified as an isolated operation; callers not claimed |
| GAP: V09/V32 supplied stages | No change | Already gated R03b/R08 under Q05 |
| GAP: TD-SKILLS17 row positions | No change | The table's first column gives "name / zero-based row" (for example Charged Bolt /38) |
| DISCRIM: V01 unconditional state removal | Spec edit: V01 | No state request asserted |
| DISCRIM: V01 removing only the first type 2 event | Spec edit: V01 | Events [2, 8, 2] become [8] |
| DISCRIM: V29(f) searcher depth by return value | Spec edit: V29(f) | M 1, Q 2, P 2: correct final 1; resolving to P gives 0 |
| DISCRIM: V13 monster reach argument order | Spec edit: V13(a) | Asymmetric operation; swapped arguments reject |
| HARDEST: R02/V24 | No rule change | Order and logged-versus-pure boundary stated |
| HARDEST: R03c/R03f/R03h with R08 | No rule change | Owned contracts and new asymmetric cases discriminate |

COVERAGE (24, advisory): STT-V03(k) also covers a total/current split. The others repeat existing
branches or fall inside owned (PVP, NRF-R10) or gated (R08, STT-R03, Q08) boundaries; no change.

R01b, R01c, STT-R02, STT-R03 and R04 text changed, so the next run is FULL. It is the fourth and
last reader run allowed in this task.

## Blind-implementer FULL12 result and open triage (2026-10-15)

Run type: FULL, the fourth and last reader run allowed in this follow-up task, on the
FULL11-edited contract. Counts: MISMATCH 0, GUESS 4, GAP 9, DISCRIM 11, GATED 19, COVERAGE 20,
HARDEST 3. Every executable vector and every LCG chain was recomputed without difference.

**The pass is NOT complete.** The run cap was reached with unresolved GUESS, GAP and DISCRIM
entries. No spec edit was made after this run, so the spec text matches the contract this reader
reviewed. Each entry's proposed outcome is listed below. An entry marked "pending" needs its edit
(asm-checked first when behavioral) and then a fresh FULL reader run. An entry marked "no change"
is a proposed outcome only, until a new run confirms the resulting contract.

| Finding | Proposed outcome | Reason / what would resolve it |
|---|---|---|
| GUESS: R10/V20(b) missing MonStats row versus missing MonStatsEx row | Pending: state both as size 0 after checking the size helper's bounds path; add a case | V20(b) pins only MonStatsEx=-1 |
| GUESS: R01f whose events | Pending wording: the sentry's own pending events | Text clarification |
| GUESS: negative delay due=frame+delay | No change | V30 pins -3 giving 97; the supplied scheduler appends due=frame+delay for any i16 |
| GAP: delay request event type and replace/append | No change | Scheduler is Q01/Q08; the preamble's supplied effect appends due=frame+delay |
| GAP: tier inputs (state 105, stat 172) | No change: owned | PVP section 3 and GDF-R02/GDF-R03 define the tier read; the reader did not open GDF |
| GAP: raw 0/8 player result versus PVP "true" | Pending: cite PVP-V12/PVP-V24 or the PVP rule that returns the masked flags | Cross-reference only |
| GAP: missile owner-resolution condition not in section 3 | Pending: add a section 3 row naming PVP-R07's owner-link-valid condition | Input inventory |
| GAP: Missiles.Size, object size and MonStatsEx inputs not in section 3 | Pending: add to the positions/size row | Input inventory |
| GAP: owner-lookup/room request arguments; absent-unit handling | No change | Q08 owns the producers; the requests' issue and order are the local contract |
| GAP: V32 supplied-only | No change | R08 is gated under Q05 |
| DISCRIM: abs dropped in R10/R04 | Pending: candidate raw(-13, 0) expecting 11; V14 points with unit (5, 5) and negative differences | Vector strengthening |
| DISCRIM: state 12 cleared before the missing-skill check | Pending: V07 case with state 12 active, instance absent; state stays active | Vector strengthening |
| DISCRIM: Skill1 lookup before budget | Pending: assert no instance request in V08(a)/V05 third think | Vector strengthening |
| DISCRIM: unsigned `aip3` | Pending: `aip3`=-1, no target, due 99 | Vector strengthening |
| DISCRIM: V29(b) flag mask | Pending: flags 0x0A give 8, flags 0x04 give 0 (PVP-R05a mask 8) | Vector strengthening |
| DISCRIM: collision before strict-improvement test | Pending: decide whether the collision request is logged; asm shows strict test first (0xB15E0), so a non-improving candidate makes no collision request | Needs a logged collision operation in V26 |
| DISCRIM: searcher-side dead modes | Pending: searcher monster modes 12/17 and completion cases | Vector strengthening |
| DISCRIM: V09(a) reach on missing target | Pending: log reach requests in the wrapper vectors; asm shows none when absent | Vector strengthening |
| HARDEST: R03f/R03h | No change | Owned PVP contracts; the GAP cross-reference above helps |
| HARDEST: R02 with R01a/R01f | No change | Order and two call sites are explicit |

COVERAGE (20, advisory): no change recorded; several overlap the pending DISCRIM cases.

Reader runs in this follow-up task: four, all FULL (FULL9..FULL12). Status stays draft. The
blind-implementer pass for this contract remains open until the pending outcomes are applied
and a fresh FULL run returns MISMATCH 0 with every GUESS, GAP and DISCRIM resolved or gated.

## Restructure (2026-10-15)

The spec review's split recommendation, approved by the user, divided this spec by trigger.
The action-17 burst, its Skills text selectors, the shared creation speed stage and the
missile footprint moved, with their vectors, questions and the burst observation recipe, to
[Assassin sentry missile launch](skill-assassin-sentry-missile-launch.md) (prefix STT). This
spec keeps the think, budget, target search, range and scheduling rules. Neither spec is
consumed, so both were renumbered in order without aliases; this companion and the gap file
use the current ids. The questions on original data values, producer integration and pending
observation were divided by the rules each part gates.

Records moved to the [launch companion](skill-assassin-sentry-missile-launch.md): the function
rows for the action-17 handler, its initializer and the creator speed stage; the initializer
import table and its stored-point correction; the burst correction; TD-SKILLS17; the launch
rows of the gap assessment; the burst observation procedure; and the launch-only asm checks
and findings of the FULL1..FULL12 reader runs. Run counts, change lists, mixed findings and
the stop-condition statements stay here and describe the combined pre-split contract (30 rule
ids, 37 vector families, four observation recipes). No rule behavior, label, value or vector
result changed. The constants rows are now separate in each spec, so the room cap 48 no longer
appears beside action 17; this bears on FULL12's pending "cap 48" wording item, which still
needs a fresh reader check. The other FULL12 pending items are unchanged, and the
blind-implementer pass for both specs remains open as recorded under FULL12.

## Publication review — 2026-10-23

A fresh independent blind reader read both complete post-split SFR/STT specs, their gap
files and the source/review sections of both companions. It recomputed representative
SFR strict-range, budget-to-zero, second-owner-failure, directional-root and chooser
cases and STT burst, callback-frame and speed cases without a blocking mismatch. This
reader approved bounded draft publication. The 2026-10-15 combined-contract FULL12
result above remains dated history: its suggested wording and discriminator additions
were not all implemented or rerun as that historical review protocol. They remain
visible leads for further coverage, while the fresh reader found no unmarked behavior
guess within the current supplied-input/static boundaries. This is not a gameplay
observation or an exhaustive branch-coverage claim.

An independent evidence reader checked both three-file sets' original source identities,
dated reuse, named text cells, public links and explicit Q/O gates. A separate
architecture reader checked SFR/STT/MCI/MMC ownership and #418's gameplay boundary.
All three approved the bounded 1.13c handoff. Issue #496 now names both three-file
sets. Split-stale BSF, GDA and coverage-index references were corrected. The launch
companion now names the initializer's total-frame read correctly, and STT-V04 states
that its stat-161 value is a supplied final list-read result.

SFR-Q01–Q09, SFR-O01–O03 and the sibling STT gates remain open. No SFR behavioral rule,
vector output or original-source evidence label changed; the historical implementation
comparison was not renewed. No new original-image investigation, protected game-data
capture, gameplay test or 1.14d confirmation was added. Both specs remain draft.
