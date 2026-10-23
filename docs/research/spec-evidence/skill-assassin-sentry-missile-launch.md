# Evidence: Assassin sentry missile launch

Spec: [Assassin sentry missile launch](../../specs/skill-assassin-sentry-missile-launch.md).
Artifacts: [original source manifest](../source-manifest.md).
Limits: [evidence gaps](../../plans/gaps/skill-assassin-sentry-missile-launch.md).
Think-side records: [Assassin sentry firing and range evidence](skill-assassin-sentry-firing-and-range.md).

## Identity and evidence reuse

Lord of Destruction Expansion 1.13c, Windows x86. Module references resolve to the
manifest's `LOD113C-D2GAME` and `LOD113C-D2COMMON` fingerprints. All descriptive function
names here are analyst labels, not original exported names. An RVA is module relative,
not a physical file offset. Original image fingerprints were checked against the manifest
before the new branch/table inspections. No game files, rows or raw captures are published.

These records were written for the combined sentry firing and range contract and moved
here with their rules on 2026-10-15 (see the restructure record below). The 2026-10-12
speed-stage reading retains its date and limits. Before the 2026-10-15 reading, the
existing NRF/GDF/SCD/SCX/SLD contracts were assessed and reused: NRF-R03a for seed
initialization and SCD/SCX/SLD for action dispatch, calculation bounds and their precise
unresolved producers. No gameplay actions, new hooks or other-build analysis were performed.

## Dispatch and function records

| Artifact | Function or table RVA | Analyst description / rules | Method and actual boundary |
|---|---|---|---|
| LOD113C-D2GAME | 0x1074A8; 0x43B60 | action table 17 and burst | SCD's 2026-10-15 table/root record plus new full handler reading/original-image decode: latch before validation, missile bounds, `calc1`, per-index resolution and two nonzero-coordinate gates, conditional creator and return 1 |
| LOD113C-D2GAME | 0x41F50 | burst initializer | Byte-byte-equivalent C reconstruction read 2026-10-15; callback cap and seed/path requests from entry to return; imported setter/path effects remain Q04 |
| LOD113C-D2GAME | 0x6F930 | shared creator speed | Reused 2026-10-12 speed-stage record; targeted branch region reread 2026-10-15 confirms explicit-speed, state/list gate and scales. Full creator/output/flight integration remains Q03/Q04 |

The 2026-10-15 review resolved the calculation import used by the action:

| D2Game thunk RVA | LOD113C-D2COMMON export / RVA | Caller role |
|---|---|---|
| 0xAC9A | #10786 / 0x51BF0 | calculation evaluation for `calc1` (SCX contract); the think side's `calc4` use is recorded with SFR-R02 |

The initializer import mappings were resolved on 2026-10-15 and their getter/setter leaves read in source order:

| D2Game thunk RVA | LOD113C-D2COMMON export / RVA | Actual callback operation and method |
|---|---|---|
| 0xAD06 | #10534 / 0x6A290 | signed i16 missile frame count; byte-equivalent C reconstruction |
| 0xAD42 | #10558 / 0x6A710 | total-frame word setter with signed i16 clamp; complete disassembly |
| 0xAC7C | #10970 / 0x6A6D0 | current-frame word setter with signed i16 clamp; complete disassembly |
| 0xAAC6 | #10983 / 0x34AB0 | unsigned stored-point X at Path+0x10; byte-equivalent C reconstruction |
| 0xA628 | #10614 / 0x36740 | InitSeed writes value and 666; reused NRF contract and byte-equivalent C reconstruction |
| 0xAAFC | #10647 / 0x34F80 | path kind request; complete disassembly, original path flag/table effects remain Q04 |
| 0xAB14 | #10168 / 0x34F50 | low-byte duration is capped at 77 and stored in two path byte fields; complete original-image decode corroborated the low-byte/cap/writes on 2026-10-15; broader path integration remains Q04 |
| 0xAAF0 | #10334 / 0x36230 | path preparation; complete producer integration remains Q04 |

The original 'pathX' wording conflated the initializer's stored-point getter with the dynamic
current-X getter. Export 10983 reads Path+0x10, matching SCD's stored-point field, rather
than dynamic current-X+0x02. V01/V02 retain supplied stored-point 1000 and seed outcomes
1000/1001/1002; equality with caster currentX is no longer asserted. The creator's actual
stored-point production remains gated. The duration setter narrows the callback frame
argument to u8 before capping at 77; a negative signed frame count therefore cannot be
modeled as min(f,77). The callback's total/current frame clamp still tests signed f>=78.

The initializer's imported operations read total frame count, set total/current frame
counts, read path stored-point X, initialize seed, set path kind/duration and prepare path. The
byte-equivalent callback contains no direct LCG step; only the seed-initializer request
is local. Actual underlying path allocation/motion is unverified. Likewise action 17's
target resolver at D2Game RVA 0x9FA30 and missile selector0x91340 are named supplied
operations; action-level request handling is separated from their complete production.

### Semantic layout correspondence

These locations identify evidence; they are not implementation requirements.

| Semantic term | Original representation | Rules |
|---|---|---|
| `calc1` | Skills+0x138 | R02 |
| action selector | Skills+0x2E signed i16 | R01/R02 |
| seed | Unit+0x20/+0x24 | R02 |
| missile initialization callback/index | creation request+0x54/+0x58 | R02 |
| explicit speed/condition | creation flags bit 2; already-scaled bit 4; value+0x28 | R03 |
| velocity fields / slow match | Missiles+0x9A/+0x9B u8; flags byte+4 and global mask byte | R03 |

## Corrections and affected assertions

### Burst gates

Full action 17 image decoding confirms that `calc1` gives loop attempts, not guaranteed
created missiles. Target resolver0x9FA30 must succeed and both coordinates must be
nonzero before0x6F930 is requested; failed creation does not abort the loop or change
return 1. V01/V02 retain their original successful-three-bolt numeric outcomes with
success/placement explicitly supplied. V03 adds resolver/coordinate/failure cases.
The callback reads each missile's OWN path stored-point X; equality to casterX is conditional on
the supplied creation placement, rather than an unconditional creator guarantee.

Full speed and footprint integration remain gated or asm-only at their stated
boundaries; the shared 2026-10-15 label record stays in the think-side companion.

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
| Charged Bolt /38 | Id; srvdofunc | 38; 17 | explicit skill identity 38 selects text action 17, R01 |
| PrimeBolt /186 | Id; srvdofunc | 186; 17 | identity 186 selects text action 17, R01 |
| BoltSentry /306 | Id; srvdofunc | 306; 17 | identity 306 selects text action 17, R01 |
| ImpBolt /322 | Id; srvdofunc | 322; 17 | identity 322 selects text action 17, R01 |
| BoltSentry /306 | `calc1` | `ln12` + skill('Shock Field'.blvl)/3 | raw expression text only; no compiled evaluation claim |
| BoltSentry /306 | `calc4` | `par8` + skill('Lightning Sentry'.blvl)/4 | raw expression text only; no compiled evaluation claim |

Full file coverage establishes these are the four rows whose raw srvdofunc cell equals 17.
This is `verified in 1.13c data`, not gameplay execution or compiled-table correspondence.
Actual MonStats Skill1/aip values remain SFR-Q03; expression compiler output and skill binding remain Q02.
No unrelated monster-skill behavior is added by listing the other text selectors.

## Gap assessment and search limits (2026-10-15)

| Original question / gap | Compatible evidence and targeted follow-up | Current disposition |
|---|---|---|
| Q01 footprint; R05 | Reused generic creator context; callback C confirms path 10 request but not steering/collision | Missile/path owner integration and O01 remain |
| Q02 original rows; R01 | SCD table bounds reused; complete narrow-purpose Skills loader read above | do 17 text membership closed; `calc1`/compiled/runtime association still open (aip/Skill1/`calc4`: SFR-Q03) |
| Q03 speed; R03/R04 | Reused speed region and new targeted flags/state/list-scale reading | Kept asm-only/full output and travel integration open; V04 numerical supplied-stage cases |
| R02 overbroad burst result | Complete handler image and callback C | Request/initialization assertions available; creation/path effects explicit Q04 |

## Observation procedures

No observation has been collected. Use an owned original 1.13c installation matching
the manifest and establish its effective data separately. Common controls: Assassin level 30,
one chosen trap rank with sufficient prerequisites, no equipment granting skill/speed changes,
Normal difficulty, same act/level/room, same owner state and stable stationary target.
Record exact rank, synergy/base levels, data values, positions and relevant seeds at entry;
level 30 is a practical setup, not a behavior constant. Coordinate units consume NRF-R10.

- **STT-O01:** Use a charged-bolt sentry and one stationary target; retain each per-index
  target-resolution, creation success and initialized missile identity. Record initial
  X/seed/frame count and successive direction/position readings until expiry. Expected:
  V01/V02 callback values only for actually initialized missiles. Steering/angle series
  and count losses are measured independently; do not infer a fixed angular spread.

## Blind-implementer records for these rules (2026-10-15)

Each reader run below reviewed the combined pre-split contract. Its counts, stop-condition
statements, mixed findings and change lists stay in the [think-side companion](skill-assassin-sentry-firing-and-range.md);
the asm checks and findings that concern only the rules of this spec moved here unchanged,
grouped by the run that produced them.

### Blind-implementer FULL1 findings (2026-10-15)

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| COVERAGE: R02 invalid selected missile | Spec edit: V03(e/f) | Negative and count-equal missile identities return 0 after latch set, without expression/creation request |
| COVERAGE: R03 signed/wrapping speed | Spec edit: V04(e/f) | Explicit raw 1073741824 wraps its 75 product to -1073741824 before /100; negative level -9 tests truncation toward zero at /8. Both remain supplied arithmetic of an asm-only integrated rule |
| HARDEST: R02 attempts/stored point/signed frame versus u8 duration | No rule change | Existing source-backed gates, distinct current X700/stored X40 and -1/-256 frame cases remain; invalid missile additions exercise earlier exits |

### Blind-implementer FULL2 findings (2026-10-15)

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| DISCRIM: V03 creator failure is last | Spec edit: V03 primary sequence | Count 5 adds index 4 success after index 3 failure. Assert creation requests 3 then 4, final return 1 and M initialization seed 44; stopping on failure omits the later observable request/effect |
| COVERAGE: R02 null caster distinct from null initializer | Spec edit: V03(g) reached handler only | Absent caster with valid supplied selector/position inputs and count 0 returns 1 without latch write or creation. This does not claim outer dispatcher accepts absent unit |
| COVERAGE: R03 slow mask matches but state 87 inactive | Spec edit: V04(g) | Explicit raw 2560 with present list/value 50 stays 1920; applying slow solely from mask would produce 960 |
| HARDEST: R02 conditional attempts/stored point/frame narrowing | No rule change | Existing independent callback current X700/stored X40 and signed-frame cases remain; primary V03 now proves continuation after creator failure |

### Blind-implementer FULL3 findings (2026-10-15)

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GAP: R02 path-preparation output missing from section 5 | Spec edit: section 5 burst-output row | Explicitly lists the already specified callback path-preparation request, without adding its unverified internal effects |
| DISCRIM: V01 equal caster/stored-point X | No numeric change; explicit coverage clarification | V03 M has current X700 and stored-pointX40,index 4; seed 44 rejects current-X seed 704. Independent callback index 2 gives 42 rather than 702. V01 pins burst count/supplied successful values and does not claim to discriminate equal source coordinates |
| DISCRIM: V02 equal caster/stored-point X | No numeric change; explicit coverage clarification | Same V03 distinct-input discriminator; V02 retains its original 1000/1001/1002 and cap 77 results. No additional duplicate seed-source case is needed |
| COVERAGE: R02 path preparation after initialization | Spec edit: V03 effect-order assertion | M seed overwrite 44 precedes path-kind 10,path-duration 77,then path-preparation request; independent seed 42 case has the same request order. Actual path-preparation internals remain Q04 |
| COVERAGE: R03 false slow-mask match with state/list present | Spec edit: V04(h) | Explicit raw 2560 with false mask,state 87 active,list/value 50 present stays 1920; ignoring mask would yield 960 |
| HARDEST: R02 continuation/stored point/signed-frame-u8 duration | No rule change | V03's later success after failure,distinct current/stored X and negative-frame cases remain exact; the new effect-order assertion finishes the output inventory without promoting path internals |

### Blind-implementer FULL9 findings (2026-10-15)

Asm checks before editing (original disassembly, 2026-10-15):
- Initializer 0x41F50 (byte-equivalent C reconstruction and asm): reads the frame count; if >=0x4E it sets the
  held value to 0x4D and calls the total (#10558) then current (#10970) setters; then seed init
  from stored-point X (#10983) + index; path kind 10 (#10647); path duration (#10168) receives the
  CLAMPED held value; then path preparation (#10334). Duration is min(u8(g), 77) with g the
  post-clamp value, so f=300 gives 77, not 44.
- Action 17 handler 0x43B60 (asm): `calc1` evaluation #10786 takes (unit, `calc1`, skill id, level),
  where skill id and level are the handler's two stack arguments from dispatch (SCD-R06 invokes
  action(game, unit, id, level)). The missile identity comes from 0x91340(skill id, unit) and is
  bounds-checked against the Missiles record count. The resolver's success return is tested
  before both coordinate words.
- Creator 0x6F930 speed region: the explicit-velocity condition is creation-flags bit 4, not a
  value test. The slow branch calls the state test #10494(owner, 87), obtains the owner's
  state-87 stat list (thunk 0xA5D4, #10871 / 0x38520: unit list then state-list lookup) and reads stat 161 layer 0 from that list (thunk 0xA712, #10680 / 0x39140).
  The slow scale precedes the nonzero 75% scale.

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GUESS: R02 duration from original or clamped f | Spec edit: R02; V03(i) | Asm uses clamped g; f=300 gives frames 77 and duration 77 |
| GUESS: R02 frame-write position | Spec edit: R02 "in this order"; V03(h) | Total then current frame write precede the seed overwrite |
| GUESS: R03 which stat list | Spec edit: R03; V04(j) | Owner's state-87 list, not the base list |
| GUESS: R03 explicit velocity 0 | Spec edit: R03 | The explicit condition is a flag; value 0 is used when it is set |
| GAP: `calc1` level argument | Spec edit: R02; V03 inputs | The action's supplied level (3); evaluation (S, 20, 3) |
| GAP: R02 missile selection source and count | Spec edit: R02 | Supplied missile-selection operation; invalid when <0 or >= Missiles count |
| GAP: V03 index 0 failed-resolution coordinates | Spec edit: V03 inputs and Discriminates | Fails while returning (5, 5); no creation request |
| DISCRIM: R02 clamp threshold | Spec edit: V03(h) | f=78 gives 77 |
| DISCRIM: V04 75% before slow | Spec edit: V04(i) | Explicit 3 gives 1 then 0 (wrong order gives 1) |
| DISCRIM: V04 stat source | Spec edit: V04(j) | State list 50 versus base list 100 gives 960 |
| HARDEST: R02 | Resolved by R02 edits | Level, missile source, clamp and duration order now explicit |

### Blind-implementer FULL10 findings (2026-10-15)

Asm checks before editing (original disassembly, 2026-10-15):
- Action 17 handler 0x43B60: one creation-request block is cleared once and filled with flags,
  owner (caster), missile identity, caster position (two getters) as source, skill id and the
  action's level, callback and index. The resolver call receives the caster and two output
  pointers that are the block's target-coordinate fields; no index, skill or level is passed. The
  loop test on the `calc1` count is signed (`jle` before the loop, `jl` at its end).

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GUESS: R02 resolution arguments | Spec edit: R02; V03 | Caster only; five requests in V03 |
| GUESS: R02 creation arguments | Spec edit: R02; V03 outputs | Source caster position, target resolved coordinates, skill/level, owner, callback, index |
| GAP: null initializer unit only in section 6 | Spec edit: R02 | Absent initialized unit does nothing |
| GAP: V04 already-scaled/slow-mask sources | No change | Section 3 speed row lists them; their producers are gated Q03/Q04 |
| DISCRIM: V04 explicit value 0 | Spec edit: V04(k) | Result 0, not 5760 |
| DISCRIM: V03 creation arguments | Spec edit: V03 outputs | As above |
| DISCRIM: V03 negative `calc1` | Spec edit: V03(j) | -1 returns 1 with no requests |
| HARDEST: R02/V03 | Resolved by R02/V03 edits | Arguments now asserted |

### Blind-implementer FULL11 findings (2026-10-15)

Asm checks before editing (original disassembly, 2026-10-15):
- D2Common #10534 (0x6A290, byte-equivalent C reconstruction) reads missile data word +0x0E,
  the field written by #10558 (0x6A710, "total"); #10970 (0x6A6D0) writes +0x10 ("current"). The
  initializer's f is therefore the total frame count.
- Action 17 handler 0x43B60: skill identity <0, >= the Skills count, or a null record pointer
  returns 0 before the missile selection.
- Creator 0x6F930: the slow-branch state test and list lookup use the creation request's owner
  field (request+4), which the action-17 handler fills with the caster.

| Finding | Outcome | Reason / exact assertion |
|---|---|---|
| GUESS: R02 f total or current | Spec edit: R02; V03(k) | Total; total 50/current 100 gives no write and duration 50 |
| GUESS: R02 invalid skill bounds | Spec edit: R02 | <0, >= count or absent record |
| GUESS: callback on creator failure | Spec edit: R02 | Only actually initialized missiles; invocation is the creator's (Q04) |
| GUESS: absent caster with n>0 | Spec edit: R02 | Supplied position/resolution results for the absent caster (Q04) |
| GUESS: R03 owner and minimum adjustment | Spec edit: R03 | Request's owner unit; RLI-R02b applicability left open under Q03 (R03 stays asm-only) |
| GUESS: R03 missing stat 161 | Spec edit: R03 | RLI-R02a reads 0, speed 0, 75% skipped |
| GAP: dispatcher result versus handler return | Spec edit: R02 | Handler returns; SCD-R12 owns the final result |
| GAP: R03 condition origins | No change | Gated Q03/Q04; section 3 lists them as supplied booleans |
| GAP: V03 base versus bonus level | Spec edit: V03 Discriminates | Removed the unexercisable "base level" claim; the level is the dispatcher's (SCD-R06) |
| DISCRIM: action return from last creator | Spec edit: V03(l) | Single failing creation returns 1 |
| DISCRIM: f=77 written by a `>=77` clamp | Spec edit: V03 outputs | No frame-count write for M |
| HARDEST: R02/V03 | Resolved by R02/V03 edits | Frame source, callback, dispatcher boundary now explicit |

### Blind-implementer FULL12 result and open findings (2026-10-15)

| Finding | Proposed outcome | Reason / what would resolve it |
|---|---|---|
| GUESS: R02 "cap 48" in the constants row | Pending wording: the constants table groups cap 48 with action 17; say it is SFR-R03c's room cap and that n is uncapped | Text clarification |
| GAP: R03 slow-mask source | No change | R03 is asm-only under Q03/Q04 |
| GAP: absent caster with n>0 | No change | R02 states supplied results for the absent caster (Q04) |
| DISCRIM: V03/V01 action identity/level equal to Skill1/instance level | Pending: add an action with skill 38, level 5 differing from Skill1 20/level 3 | R02 uses the dispatcher's arguments |
| DISCRIM: seed hi word | Pending: enter seed {6, 0x1234}, expect hi 0x29A (NRF-R03a) | Vector strengthening |
| DISCRIM: V04 positive VelLev term | Pending: Vel 30, VelLev 8, L 9 gives (30+9)<<8, 75% = 7488 | Supplied arithmetic of asm-only R03 |
| HARDEST: R02/V03 | No change | Each condition has a sub-case |

## Restructure (2026-10-15)

This spec was created by dividing [Assassin sentry firing and range](skill-assassin-sentry-firing-and-range.md)
by trigger, following the spec review's split recommendation approved by the user. It holds
the action-17 Skills text selectors, the burst and its initializer, the shared creation speed
stage and the gated missile footprint, with their vectors, questions, the burst observation
recipe and the records above. The think, budget, target search and range stay in the source
spec. Neither spec is consumed, so both were renumbered in order without aliases. The source
questions on original data values, producer integration and pending observation were divided:
this spec keeps the `calc1`/Vel/selector values, the missile-selection, target-resolution and
creation producers, and its own pending observation.

No rule behavior, label, value or vector result changed. Section 3 now lists the action-17
constants apart from the think's room cap 48, which bears on FULL12's pending "cap 48" wording
item; that item still needs a fresh reader check. The FULL12 pending items for these rules
(the action identity/level case, the seed high word and the positive VelLev term) remain open,
and a fresh FULL reader run of this spec is still required.

## De-duplication against the missile creator (2026-10-15)

STT-R03 restated the generic creation speed stage that missile-creation-and-seeding owns as
MCI-R05 (initial speed), MCI-R06 (slow-missiles scale) and MCI-R07 (global speed scale). The
rule now consumes those three rules and keeps only its local consequences: the owner read by
the stage is the burst's caster, a state-87 list without stat 161 reads 0 and then skips the
75% scale, and the open RLI-R02b minimum question (STT-Q03). The two texts agreed on every
value (explicit/scaled velocity, unsigned Vel/VelLev, wrapping, slow scale before the 75% scale,
zero speed skipping the 75% scale). STT-V04 keeps all eight cases unchanged and now names
MCI-R05/MCI-R06/MCI-R07; MCI-V04 remains the owner's vector. The STT-R03 label stays asm-only
because it covers this spec's local binding, not the owner's verified arithmetic. No rule
behavior, label, value or vector result changed.

## Publication review — 2026-10-23

A fresh independent blind reader reviewed the complete post-split SFR and STT specs,
their gap files and the source/review sections of both companions. It recomputed the
STT creation-attempt, failed-creator, signed-frame clamp/narrowing and slow-before-75%
speed cases, and representative SFR timing/range/budget/selection cases. No blocking
mismatch was found, and the reader approved bounded draft publication. The 2026-10-15
FULL12 pending suggestions above remain dated history; not all proposed discriminator
cases were added. This fresh review accepts the current explicit static/supplied-input
boundary without claiming exhaustive branch coverage or a renewed original-image read.

An independent evidence reader checked both three-file sets' artifact identities,
effective-text data limits, links and open gaps. A separate architecture reader checked
the think/action-17 split, MCI speed ownership, later MMC flight/collision ownership
and #418 gameplay boundary. All three approved bounded 1.13c publication. The import
summary now says the initializer reads total frames, matching the existing #10534
record and STT-R02/V03(k); no original instruction was newly inspected. STT-V04 now
names stat 161=50 as a supplied final list-read result, leaving real minimum-adjustment
applicability under STT-Q03. Issue #496 now names both three-file handoffs.

STT-Q01–Q05, STT-O01 and the sibling SFR gates remain open. No STT behavioral rule,
vector output or original-source evidence label changed; the historical implementation
comparison was not renewed. No new protected game-data capture, gameplay test or 1.14d
confirmation was added. Both specs remain draft.
