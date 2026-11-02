# Monster curse and aura skills

**Description**

The monster curse action makes one random choice from five fixed skills before checking the selected skill record. Which units receive the selected effect and exactly what happens to them still require investigation.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), [core-game-data-fields](core-game-data-fields.md)
Ids: MCA-R, MCA-Q, MCA-V, MCA-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-monster-curse-and-aura.md)
Open evidence gaps: [gap file](../plans/gaps/skill-monster-curse-and-aura.md)

## 1. Scope

The server skill actions selected by Skills.srvdofunc values 111, 112.
The boundary begins at the selected action entry and ends at its observable
return and effect requests. AI selection, skill dispatch, generic damage,
missile simulation and monster creation are separate contracts; their integration
here remains gated. This document does not establish other boss actions.

## 2. Trigger and ownership

D2Game computes these actions through the server do-function table selected by
Skills.srvdofunc. [Skill dispatch](skill-cast-start-and-do-dispatch.md) owns that
selection under SCD-R06, SCD-R07 and SCD-R08. AI selection is an outside-boundary producer: it does not gate the supplied action context. Frequency and relative frame ordering are not asserted by this contract.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning | Origin | Binding |
|---|---|---|---|---|---|
| Action selector | Skills.srvdofunc | integer | Selects one action in Scope | table data | selected by data from engine-defined handlers |
| Skill index | Dispatcher-supplied skill identity | signed 32-bit | Record lookup input | runtime input | supplied input |
| Skill level | Dispatcher-supplied level | signed 32-bit | Calculation input | runtime input | supplied input |
| Actor and game | Current action context | logical units and game | Owner, targets, difficulty and seed; exact consumption gated | runtime input | supplied input |
| Refresh frame and existing effect metadata | Game frame; selected existing list's state, skill, level, expiry, timed condition, source owner and removal callback under OWD-R07a/STL-R06a | signed i32 frame/identities/level/expiry; logical metadata | R18's equal-state/skill/level refresh; pending frame entries consume STL-R01 | runtime input | existing records, independent of a new allocation |
| Reached replacing-life storage | MRH-R08's extended root base, independent own/resolved arrays, stat-6 controls and prepared capacity; RLI-R02a/R02b's subsequent reads | signed i32 raw 8.8; valid sorted stat/layer entries | R19 binds R14/R16's actual unit writes to the named direct-storage contract; no supplied writer effects | runtime input | existing storage; captured definition controls retain the owning evidence limits |
| Fixed selected identities | Engine selection mapping | signed 32-bit | 66, 72, 82, 87, 91 in index order | engine code | fixed by original code |
| Actor seed | Actor random state | two unsigned 32-bit halves | Updated by MCA-R01 | runtime input | supplied input |
| Selected-record collection and count | Supplied loaded Skills collection and signed record count | logical collection; signed 32-bit count | V07 supplies an empty collection with count 0; contextual validation input used only after MCA-R01 ends, never read by its selection stage | runtime input | supplied context |
| Selected skill record and loaded counts | Supplied Skills record, signed skill/state counts | logical record, signed 32-bit counts | R04 checks selected identity and target-state range | runtime input | supplied record context |
| Target state | Skills.auratargetstate, selected record; [local field evidence](../research/spec-evidence/skill-monster-curse-and-aura.md#selected-field-names) | signed 16-bit | Target state passed in the request | table data | resolved from supplied loaded record |
| Radius/duration calculations | Selected Skills.aurarangecalc / auralencalc; [local field evidence](../research/spec-evidence/skill-monster-curse-and-aura.md#selected-field-names) | unsigned 32-bit references, signed 32-bit results | Requested in this order under SCX-R11 | table data | resolved from supplied loaded record |
| Parameter 7 | Selected Skills.Param7 | signed 32-bit | Copied into index-4 speed/attack-rate requests | table data | resolved from supplied loaded record |
| Do-handled latch | Actor animation condition, SCD-R01/SCD-R07/SCD-R09 | boolean | Set after valid record/state guards | runtime input | shared actor condition |
| Supplied expression/recipient operations | SCX-R11 and MCA-Q04 recipient boundary | logical logged operations | Expression return/draw behavior and normally returning recipient result supplied by fixture | runtime input | dependency operations |
| Selected linear parameters | Skills.Param1 through Param4 | signed 32-bit | R06/R07's observed-profile radius and duration | table data | narrow loaded-record correspondence |
| Candidate kind, mode and class | Recipient unit | kind/class integer; signed mode | R08/R10 checks | runtime input | supplied candidate |
| Candidate own base stat and hireling classification | RLI-R02a/RLI-R02b; CSA-R02/CBL-R04a | signed 32-bit stat; boolean classification | R09's reduction eligibility; own value is distinct from resolved totals | runtime input | named dependency results |
| Candidate owner/relation graph, target flags and availability | PVP-R05a/R07; SFR-R03e | logical graph, flags, mode/exclusion condition | R10's repeated eligibility requests | runtime input | shared read contracts |
| Monster target exclusion | Candidate monster data | boolean | R10's explicit local exclusion check; evidence mapping in companion | runtime input | supplied condition |
| Control-0 eligibility inputs | Loaded monster and extended records; candidate's existing states and conditions | signed class/base-link/counts; flags/booleans; u16 monster condition word | R12 derives eligibility from `mWL`, `boss`, `switchai`, active states 54/143, target flag bit 2 and monster condition mask 10 | runtime input | loaded definitions and existing candidate conditions |
| Candidate alignment | State-105 list's selected layer-0 stat 172 | signed 32-bit; PVP-R05a / RLI-R02a/R02b | R12's alignment-1 rejection; a nonextended list selects its own array | runtime input | named read contract |
| Target-refresh/coordinate context | MTP-R09's present game, actor/path, cached target and saved kind/id, registry and stored target words | logical references; unsigned target/dynamic coordinate words; static coordinates i32 | R13 obtains the actual center, including MTP-R09's cache-clear effects | runtime input | named coordinate consumer |
| Room and unit traversal context | Actor current room, rectangles, ordered room array/count, room unit heads and next-unit links | optional initial room; present visited room records; signed i32 rectangle/count; logical unit links | R13's normal readable-storage traversal; static coordinate storage exists for static kinds | runtime input | supplied world records, not a world-generation claim |
| Candidate coordinate storage | Dynamic path or prepared static-position record | dynamic Y/X integer words u16, static Y/X i32 | R13's Y-then-X reads; an absent dynamic path gives 0 per coordinate, while static kinds require readable static storage | runtime input | local wrappers plus NRF-R10b's path getters |
| Event callback context | IEP-R07: game, event, owner, other, shared incoming combat record, registered skill and level | logical references; signed i32 skill/level; i32 raw physical damage | R14/R15 consume these roles, independently of the action actor/recipient names | runtime input | dispatcher/event-entry values |
| Callback state/source context | Selected Skills.auratargetstate; RLI-R03 existing state membership/list; list source kind/id and present game/readable registry for kinds 0..5 | signed i16 state; signed count; logical existing list/source identities | The existing state list supplies calculation source context; missing registry match is passed to SCX; malformed source kind/storage has no asserted normal callback result | table data and runtime input | selected state and existing list source metadata |
| Callback calculation and request fields | Skills.`calc1`/`calc2`/`calc3`, Param1..8, ResultFlags, HitFlags, HitClass, prgoverlay | unsigned i32 references; i32 parameters/flags/class; u16 ResultFlags; signed i16 overlay | SCX-R11/R20/R21 evaluate; local loader/consumer mappings are in the companion; supplied records need not match captured cells | table data | loaded record values; consumer meaning fixed by code |
| Resource and overlay context | Existing other-unit stat 6 and effective stat 7 under RLI-R02a/R02b and LSE-R10; LSE-R13 overlay context | signed i32 raw 8.8; signed overlay count; existing lists | R14 computes replacing-life and optional overlay requests; R19 covers reached direct life writes, broader storage retains RLI-Q03/IEP-Q09 | runtime input | named existing-read and overlay contracts |
| Combat dependency context | DPO section 3 and RLI section 3.2 logical combat record; DPO apply and reaction contracts | complete logical record and normal dependency continuation | R15 prepares a distinct cleared record; dependency-owned changes and draws are not fabricated | runtime input | named combat/dependency contracts |
| Class-290 continuation context | Other's current MonStats record and Drain/Drain(N)/Drain(H); game difficulty; owner's monster-data control record and its game/kind/id lookup | difficulty 0..2; unsigned byte Drain; logical optional records/linked unit | R16 reads current post-combat physical and resolves an optional link; LSE-R02/GDA-R08a supply the named Drain field | table data and runtime input | named Drain field and existing control/registry records |
| Capped-add amount and sequential stat view | R16's computed distribution or isolated helper input; existing stat 6/stat 7 reads | signed i32 raw 8.8 and optional fixture-only replacing-view updates | Defines requested values and computed return; R19 covers actual reached direct storage, broader effects remain RLI-Q03 | runtime input | local consumer arithmetic; fixture updates explicitly supplied |
| Timed-application local request | Fresh source/target/skill/level/duration/stat/value/state/callback request copied by the curse recipient | semantic record; duration/value signed i32 | R17 distinguishes this per-candidate copy from R05's prepared request | runtime input | local prepared request copy |
| Curse-policy inputs | Selected state's curse membership under GDA-R13; target stat 109 under RLI-R02a/R02b; target active state 57 under RLI-R03 | boolean, signed i32 resistance and duration; state count | R17's recorded curse states use this branch; aura state 68 does not | table data and engine code | membership resolved from data; stat 109/state 57 consumers fixed by code |
| Existing-list selection inputs | Target's optional extended base list, ordered primary/secondary chains and each list's curse condition/state identity | logical existing lists and conditions, no imposed layout | R17 first selects a primary curse list or, for noncurse state, consumes RLI-R03 state-list lookup; stored construction remains gated | runtime input | existing list graph; curse-condition consumer fixed by code |
| Aura collection preparation | Action-111 actor/game, invoking Skills record and Param4, caller level; MTP-R09 coordinate context and actor current room | signed i32 Param4/level/coordinates; logical references | R20 sets the do-handled latch after coordinate success and computes collection extent before requesting storage | table data and runtime input | supplied invoking record; Param4 schema under GDA-R11a |
| Prepared aura collection | Existing collection context and independent readable candidate storage, capacity 15; actor, current room, center, extent and accepted counter 0 | logical context/storage; unsigned i32 room-array count; wrapping i32 coordinates/extent | R20 executes the actual fixed-filter collector without supplying predicate results or writes; initial allocation remains Q03 | runtime input | initialized context supplied at its collector-entry boundary, not a claim about its allocation |
| Aura candidate room and flags | Candidate kind/mode, dynamic or static coordinate storage, target flags bits 2/3; candidate current room and level identity | existing unit/room records, coordinates as in R13, i32 level identity | R20's geometry precedes kind/mode, flags and candidate-room exclusion | runtime input | readable existing records; room exclusions fixed by code |
| Independent aura request context | R11's source/target, invoking Skills record and Param1/Param2/Param3, caller level; R17/R18's existing state/list/game context | logical units/records; signed i32 parameters and level | Constructs the actual noncurse request and consumes the existing-list refresh; optional list return is an output, not a supplied operation result | table data and runtime input | supplied invoking record under GDA-R11a; independent request boundary, not a claim of R10 eligibility |
| Recipient event collection | Candidate's existing ordered entries under IEP-R01/R05/R06/R07, with kind, owner identity, event, callback and two values | logical entries; event u8; identities/values signed i32; executing/removal conditions | R21 removes skill entries for the selected target state and creates R05's fixed bindings; callbacks later consume selected skill and caller level | runtime input | existing event records; logical insertion/removal consumed from the named owner, not an allocator result fixture |




The selector mappings are code-table observations, not verified archive cell claims.
The effective text-file selector cells are established by MCA-R03. R06 establishes additional narrow loaded-field correspondence for its dated source profile; it does not establish arbitrary profile production or gameplay effects.

## 4. Rules (execution order)

**MCA-R01** Fixed random skill selection [verified in 1.13c code (static)]

For action selector 112 with a supplied non-null actor, advance that actor's
64-bit seed once using NRF-R01 and NRF-R02 in the numeric foundation. Reduce the
unsigned new low 32-bit half modulo five. Map indices 0, 1, 2, 3 and 4 to
skill identities 66, 72, 82, 87 and 91 respectively. These identities are fixed
by engine code, not obtained from the caller's skill index or a table row.
This draw and stored seed update happen before validating the selected record.
The selected identity is an intermediate result; this rule asserts no effect
application, recipient, final return, or caller-skill behavior.

**MCA-R02** Effect selection and application [unresolved]

Complete selector-111 storage production and stored aura effects remain Q03; R20 establishes preparation and actual collection through its supplied prepared-storage boundary, and R10/R11 establish subsequent local candidate tests and state request. Complete selector-112 stored state/stat/event effects remain Q04; R08/R09 establish its local predicates, R12 establishes the captured index-4 callback rejection, and R13 establishes actual outer traversal and its produced return at the normal readable-storage boundary. R04/R05 establish preparation and return forwarding independently.

**MCA-R03** Original text-file action bindings [verified in 1.13c data]

The named cells below are in effective `data/global/excel/skills.txt`, selected
by the `skill` column and read by the `srvdofunc` header. Origin: table data.
Binding: selected by data from engine-defined handlers. They establish text-file
contents. The named identity/start/action selectors additionally match the loaded records in the limited source profile recorded in the companion. R06 adds narrow observed field values.

| Row identity (`skill`) | `srvdofunc` raw decimal value |
|---|---|
| FetishAura | 111 |
| MonCurseCast | 112 |

Source: [table-data record](../research/spec-evidence/skill-monster-curse-and-aura.md#table-data-record).

**MCA-R04** Selected-record validation and calculation order [verified in 1.13c code (static)]

After R01, reject with result 0 if the selected signed skill identity is
negative, at least the signed loaded skill count, or its record is absent.
Otherwise read that selected record's signed 16-bit `auratargetstate`; reject
with result 0 if it is negative or at least the signed loaded state count.
These rejections make no calculation or recipient request and do not change
the do-handled latch. R01's preceding seed update is retained.

The target-state and two expression field-name mappings are established by the [local loader-and-consumer evidence](../research/spec-evidence/skill-monster-curse-and-aura.md#selected-field-names); GDA-R11a supplies the shared record/Param7 schema, not these locally documented field mappings.

On success set the actor's do-handled latch true, then request calculation of
the selected record's `aurarangecalc`, followed by `auralencalc`, both through
SCX-R11 with context (actor, selected skill identity, caller-supplied level).
Use their signed 32-bit returns as radius and duration without local clamping.
The caller's skill identity does not replace the randomly selected identity.
There is no local draw in R04; expression-owned draws retain SCX's labels and
gates. Supplied results establish this caller's order independently of expression
production and evaluation. No positive-level check is performed here.

**MCA-R05** Five effect requests and recipient handoff [verified in 1.13c code (static)]

With R04's selected record, state, radius and duration supplied, prepare one
effect request containing game, actor, selected skill identity, caller level,
duration and selected target state. Set monster-only eligibility false by
default; prepare no primary stat, no additional stats and no event bindings
unless the index's row below supplies them. All stat values are signed 32-bit;
stat identities and event-binding choices are fixed by engine code. Parameter
7 is the selected record's signed 32-bit `Param7` under GDA-R11a.

| R01 index / selected skill | Primary stat/value | Ordered additional stat/value pairs | Monster-only | Ordered event id / callback-selector bindings |
|---|---|---|---|---|
| 0 / 66 | 36 `damageresist` = -100 | none | false | none |
| 1 / 72 | 25 `damagepercent` = -50 | none | false | none |
| 2 / 82 | none | none | false | (5,4) |
| 3 / 87 | none | none | false | (1,5), (2,5) |
| 4 / 91 | 67 `velocitypercent` = Param7 | (25 `damagepercent`,-50), (36 `damageresist`,-50), (68 `attackrate`,Param7) | true | none |

These are exactly the prepared requests; the skill names or expected player
spell effects do not override this fixed mapping. The event ids and callback
selectors identify requested event bindings, not their callback arithmetic.
Unused stat/event entries produce no requested binding. This semantic request
does not prescribe storage, sentinel values or an array layout.

Request recipient processing with filter 3, actor, supplied radius, the shared
curse-recipient operation and this effect request. The recipient operation
consumes the semantic fields above; its eligibility, state/stat writes and event
registration effects remain MCA-Q04. Forward the normally returning recipient
processor's integer return unchanged as the action result. R05 makes no local
random draw. Dependency-owned draws/effects are outside this supplied-request
stage and cannot be inferred from a prepared request.

**MCA-R06** Narrow original-profile selected fields [verified in 1.13c data]

The dated [loaded-field record](../research/spec-evidence/skill-monster-curse-and-aura.md#narrow-loaded-skills-correspondence)
establishes these named Skills values in its limited source profile. Origin:
table data; binding: resolved from the observed loaded record, not hardcoded
in the action. All five loaded `Param7` values are zero; empty text cells are
not the evidence for zero. Each complete compiled radius expression selects
linear parameter 0; each complete duration expression selects linear parameter
2 under SCX-R01/R04/R11/R20/R21.

| Named skill / Id | auratargetstate | Param1 / Param2 | Param3 / Param4 |
|---|---|---|---|
| Amplify Damage / 66 | 9 amplifydamage | 3 / 1 | 200 / 75 |
| Weaken / 72 | 19 weaken | 9 / 1 | 350 / 60 |
| Life Tap / 82 | 58 lifetap | 4 / 1 | 400 / 60 |
| Decrepify / 87 | 60 decrepify | 6 / 0 | 100 / 15 |
| Lower Resist / 91 | 61 lowerresist | 7 / 1 | 500 / 50 |

The same dated loaded-record capture additionally establishes callback fields:
skills 82 and 87 each have ResultFlags=0, HitFlags=0 and HitClass=0. Skill 82
has prgoverlay=151 and `calc1` reference 1805 selecting linear parameter 4,
with Param5=50/Param6=0; its `calc2`/`calc3` are empty. Skill 87 has
prgoverlay=-1 and all three calculation references empty. These are narrow
captured values, not deductions from empty text cells. The companion supplies
the local loader/consumer mapping for ResultFlags, HitFlags, HitClass and
prgoverlay; GDA-R11a supplies the shared record and calculation schema.

A separate 2026-10-31 [loaded-state capture](../research/spec-evidence/skill-monster-curse-and-aura.md#loaded-state-membership-correspondence) establishes state count 185 and
curse membership true for selected states 9,19,58,60,61, and false for state
68. Stored state identities and definition fields agreed with the six-word
membership masks. These are narrow dated loaded observations with the
companion's profile limits, not gameplay or whole-table authentication.

**MCA-R07** Selected profile formula results [verified in 1.13c code (static)]

For R06's captured complete expressions and parameters, SCX-R21 gives radius
`Param1+(L-1)*Param2` and duration `Param3+(L-1)*Param4` when signed `L>0`.
For `L<=0`, each result is zero. Arithmetic wraps i32, with no local positive
clamp or random draw. R04 still evaluates radius first, duration second, using
the selected skill rather than the invocation identity. These derived results
establish evaluation of captured compiled expressions, not observation of a cast.

**MCA-R08** Fixed recipient kind/mode predicate [verified in 1.13c code (static)]

After the outer processor's coordinate, room and distance stages supply a
candidate, filter 3 admits players only when mode is outside {0,17}, and
monsters only when mode is outside {0,12}. All other kinds fail. No other
filter-bit predicate is active for this fixed filter. This predicate does not
establish the outer room/geometry result or replace the later R10 eligibility.
It makes no mutation or random draw.

**MCA-R09** Nonpositive resistance-stat adjustment [verified in 1.13c code (static)]

Transform a requested `(target,stat,value)` before the recipient's primary or
additional list-stat request. For stat identities {36,37,39,41,43,45}, a
nonpositive value on a nonplayer, nonhireling target reads the base list's own
layer-0 value through RLI-R02a/RLI-R02b. If that read is at least 100, return
signed `value/5`, truncating toward zero. Otherwise return value unchanged.
Other stats, positive values, players and recognized hirelings return value
unchanged without the base-stat request. No resolved-total substitution,
saturation, draw or write occurs. Hireling classification consumes CSA-R02 /
CBL-R04a's original class predicate, with their existing limits.

**MCA-R10** Aura candidate eligibility order [verified in 1.13c code (static)]

For each candidate supplied by action 111's collection, request the PVP-R05a /
PVP-R07 relationship read with `(actor,candidate,game)`. Nonzero skips the
candidate. A remaining null candidate or class outside unsigned ranges
141..145 and 396..400 is skipped. Then perform target eligibility: a monster
requires GDA-R09a `MonStats2.mWL` and no monster-data exclusion bit. All
candidates require target flags bits 1,2,3 all set, followed by a false
`Unavailable(candidate)` result. `Unavailable` returns true for absent units,
`aliveExclusionCondition`, player modes 0/17, monster modes 0/12, and every
other unit kind; otherwise false. This consumes only SFR-R03e's explicitly
defined unavailable-for-combat subcondition. The local predicate reads the
candidate's mode/exclusion conditions; it makes no actor-availability,
room/town or condition-A request. Finally request the same relationship read again with the same three
input identities; only nonzero continues.

The intervening local predicates do not mutate the owner/relation graph. Thus
unchanged supplied graph/tier/relation inputs cannot pass both opposite tests.
This is a conditional consequence, not an assertion that relationship inputs
can never change or that the whole aura is unused. Actual collection and full
state effects remain Q03. This stage makes no draw or effect request on a
skipped candidate. If the later relationship succeeds, request the state
application in R11.

**MCA-R11** Aura state request and local return [verified in 1.13c code (static)]

After R10 succeeds, use the invoking skill's signed Params1/2/3 and caller level
L to request `(source=actor,target=candidate,skill=invokingSkill,level=L,
duration=Param2*L,stat=25,value=Param1+(L-1)*Param3,state=68,
callback=default)`. Products/sums wrap i32; no local level clamp or draw occurs.
FetishAura's captured Params1/2/3 are 10,50,5. The request reaches the same
generic timed-state consumer as OWD-R06a/R06b/R07a/R07b/R08. Its applicable
policy, replacement, attach/write/free and refresh limits remain Q03;
the prepared request does not establish stored effects. The local
request-producing helper returns the normally returning applier's optional
list reference unchanged; the action-111 caller ignores that local result.
At R17's independent request boundary, the noncurse existing-list refresh
consumes R18's first paragraph and OWD-R07a. A non-null local/applier result
does not establish passing R10's two opposite relationship tests or a complete
action result. Original state-list creation and other stored effects remain Q03.

**MCA-R12** Original-profile zero primary rejection [verified in 1.13c code (static)]

For R01 index 4 with R06's selected Lower Resist record, the prepared primary
stat 67 has value 0. The actual recipient first requires a present monster
candidate and an alignment read unequal to 1. Then evaluate its fixed
control-0 eligibility in this order:

1. Start canonical class at the candidate's class. When that class has a valid
   loaded MonStats record, replace it with that record's signed canonical-base
   link. If the resulting identity is nonnegative and below the loaded
   MonStats count, equals 492, and candidate state
   143 is active, reject.
2. Reject active candidate state 54.
3. Require a valid candidate-class MonStats record, a valid signed MonStatsEx
   link and extended record, `MonStats2.mWL=true`, `MonStats.boss=false` and
   `MonStats.switchai=true` under GDA-R09a/GDA-R16.
4. When target flag bit 2 is false, require `Unavailable(candidate)=true`
   as defined in R10. A true flag bypasses that availability request.
5. Reject a nonzero candidate monster condition word AND 10; absent monster
   data supplies zero for this mask test. Otherwise control-0 eligibility passes.

Every rejection returns 0. These are the actual read-only guard operations,
not a supplied successful eligibility result. If they pass, transform primary
stat 67 through R09: it remains 0, so return 0 before the later relationship
check, state application, additional stat requests and event registration.
The guard chain reads eligibility and makes no state/list/stat/event mutation.
This conditional callback result does not establish which candidate reaches
the outer processor or the outer processor's final result. It does establish
that no callback in this captured index-4 profile reaches the prepared effects.

**MCA-R13** Actual fixed-filter recipient traversal and return [verified in 1.13c code (static)]

For R05's actual recipient operation, consume MTP-R09 to refresh the actor's
target cache and obtain center `(x,y)`. Its cache-clear writes, if any, persist.
Coordinate result 0 returns 0 from this processor. Otherwise get the actor's
current room and apply only MTP-R04's current-then-adjacent rectangle search
to find a room containing the center. This consumes the rectangle-search
subrecipe, without the direct-teleport caller's linked-monster fallback or
relocation. No containing room returns 0.

For a found room, read its stored ordered room array and signed count. The
found room is visited only if present in that array; it is not separately
prepended by this processor. Capture `rSquared=wrap32(r*r)` from R04's signed
radius `r`. Starting at index 0, visit array indices in order while index<count. In this fixed
filter-3 route there is no room-town request. For each visited room, whose
storage must be present/readable at this boundary, read rectangle
`(rx,ry,width,height)` and apply this exact engine prefilter using
signed comparisons and i32 wrap in every sum/subtraction:

```
if wrap32(x+r) >= rx:
    if wrap32(y+r) >= ry: admit room
    else: admit iff wrap32(y-r) <= wrap32(ry+height)
else:
    if wrap32(x-r) > wrap32(rx+width): reject room
    else if wrap32(y+r) >= ry: admit room
    else: admit iff wrap32(y-r) <= wrap32(ry+height)
```

These ordered checks are the required prefilter; do not replace them with a
conventional rectangle/circle-overlap test. For each admitted room, begin at
its current unit head. Read each candidate's Y then X: objects/items/tiles use
their prepared static i32 coordinates; players/monsters/missiles use the
unsigned high-16 integer parts of their dynamic path, or 0 per coordinate if
that path is absent. Compute `dx=wrap32(candidateX-x)` and
`dy=wrap32(candidateY-y)`, then signed wrapping
`dSquared=wrap32(dx*dx+dy*dy)`.
If `rSquared>=dSquared` by signed comparison, run R08; if R08 passes, invoke
the actual curse recipient with that candidate and R05's prepared request.
Ignore its normally returning result. Read the candidate's next-unit link
after that callback, then continue; an empty head/link ends the room's loop.
Callback-owned effects and mutations remain Q04 rather than a supplied
successful state/list result.

After normal traversal, including an empty room array or no accepted
candidates, return exactly 1. This rule makes no direct draw; MTP-R09's
refresh effects and each callback's effects remain attributed to their owners.
R05 forwards this produced 0 or 1 as the action result. This does not replace
V20's isolated test of forwarding an arbitrary fixture-supplied result.
Malformed unreadable room/unit storage or a nonreturning callback does not
acquire a normal integer result from this rule.

**MCA-R14** Registered selector-5 life callback [verified in 1.13c code (static)]

Use IEP-R07's owner, other, incoming record, registered skill and level roles.
The callback stage requires present game/readable state and registry context
as listed in section 3; a malformed kind/storage or nonreturning lookup is
not assigned a normal integer result.
Require present owner and other, each player or monster; require other to be
available under R10's exact Unavailable subcondition; require a present incoming
record with signed physical damage greater than 0. Otherwise return 0.
Look up the registered skill under GDA-R11a; absent record returns 0.
Normalize a negative level to 1, preserving level 0. Require the selected
signed target state to be in [0,stateCount), active on owner under RLI-R03,
and to have an existing state list; otherwise return 0. Obtain that list's
source kind/id, look up the source in the game registry, and evaluate selected
`calc1` through SCX-R11 using that source (including absent source), registered
skill and normalized level. No owner or other substitution is made for source.
Compute `amount=mulDiv(percent,incoming.physical,100)` in this argument order
under NRF-R05. Read other's current raw life, then effective maximum life.
If either is nonpositive, return 0 without a write or overlay request.
Otherwise set `n=wrap32(current+amount)`, raise signed n below 1 to 1, then
cap n at maximum (`n>=maximum` selects maximum). Request replacing other's
layer-0 stat 6 with n. A zero or negative amount does not suppress this
request; no state-92 life-add gate is made here.

After the replacing request returns normally, if selected Skills.prgoverlay
is strictly positive and less than overlayCount, consume LSE-R13's overlay
operation on other with that identity and zero auxiliary input. Otherwise
make no overlay request. Return exactly 1 after normal continuation, regardless
of a dependency return value. The incoming record is not locally modified.
The callback has no direct draw; calculations and overlay operations own theirs.
R19 establishes reached replacing-life storage through its named direct
contract. Other replacing-write/overlay storage and updates retain RLI-Q03/IEP-Q09, not an
inferred successful stored-life change. This callback is the fixed selector 5
requested for index 3 by R05; generic callback fixtures do not establish a
new registration binding.

**MCA-R15** Registered selector-4 combat callback preparation [verified in 1.13c code (static)]

Use the same normal readable state/registry context as R14.
Require present owner and other; owner must be player or monster and available
under the exact Unavailable subcondition. Require present incoming record with
positive signed physical damage and a valid registered Skills record; failed
guards return 0. Normalize negative level to 1, preserving zero. Validate the
selected target state, owner membership and existing list, then resolve its
source context exactly as R14. For a regular monster owner (monster and not
CSA-R02/CBL-R04a's hireling class), select `calc1`; otherwise select `calc2`
when other is a player, or `calc3` for every other present kind. Evaluate that
reference using source, registered skill and normalized level through SCX.

Clear every named logical field in RLI section 3.2/DPO section 3's combat
record; binary padding is not an output obligation. Set physical to
`mulDiv(percent,incoming.physical,100)` under NRF-R05. Set result to selected
Skills.ResultFlags OR 32, flags to selected Skills.HitFlags, and hitClass to
selected Skills.HitClass. All other named fields remain zero. This new record
is distinct from the incoming shared record.

Issue DPO's direct apply with game, attacker=other, defender=owner,
scaleRequest=true and the new record; after normal continuation issue the
reaction with the same game, attacker, defender and current new record.
Do not copy the incoming record's flags, result, other damage or lengths.
The actual DPO scale/apply and HRC reaction may mutate the new record and
units and own their draws; their integration gates remain in force.
After both calls, the class-290 continuation is R16; every other other-unit
kind/class returns 1. No result from either dependency controls this local
continuation. There is no direct draw; SCX and reached dependencies own their
draws. R05 fixes selector 4 on index 2; this rule establishes no
additional event registration.

**MCA-R16** Selector-4 class-290 life continuation [verified in 1.13c code (static)]

After R15's reaction, only other being monster class 290 enters this stage.
Obtain other's current monster-data MonStats record; absent record or its
unsigned Drain byte for game difficulty 0..2 being 0 ends with callback 1.
Read other's existing alignment under PVP-R05a/RLI-R02a/R02b; nonzero likewise
ends with 1. Resolve owner's optional linked unit from its monster control
record's game/kind/id (absent owner monster data/control/game yields absent
link). If a link is present and Unavailable, or owner is Unavailable, end with 1.

Use the current new-record physical after apply/reaction, not incoming physical.
If Drain is not 100, replace that value by `mulDiv(Drain,value,100)`; then
compute `remaining=mulDiv(20,value,100)`. Nonpositive remaining ends with 1.
The local capped-add operation used below differs from LSE-R09: for amount<=0
return 0 with no read/write. Otherwise read current stat 6 then LSE-R10's
maximum, set `n=wrap32(current+amount)`, and only if n>maximum cap n and change
returned amount to `wrap32(maximum-current)`. Request replacing layer-0 stat 6
with n and return the computed amount after normal continuation. There is no
minimum-1 clamp, alive/state gate, or positive current/maximum requirement
inside this local operation. R19 establishes the reached direct replacing
effects; other stored replacing effects retain RLI-Q03.

If link exists, call that capped-add on link with
`mulDiv(50,remaining,100)` and subtract its returned amount with wrapping.
Call capped-add on owner with remaining and subtract its returned amount.
If link exists and remaining is still positive, call capped-add on link again
with remaining; ignore this last return. Sequential reads consume whatever
existing stat values the write dependency actually produces; a request is not
proof those values changed. Finally consume LSE-R13 overlay 151 on owner,
then on link if present, and return 1. No local draw is made. Full combat,
replacing-write and overlay effects remain their named dependency gates.

**MCA-R17** Actual curse-recipient application prefix [verified in 1.13c code (static)]

For R05 indices 0 through 3, the actual recipient has monsterOnly=false.
After R13's distance/R08 filtering, transform its present primary through R09;
a transformed zero returns 0 before eligibility. With no primary, omit that
transformation. Then consume only R10's target-eligibility subrecipe:
monster mWL/exclusion checks when applicable, required target flags, target
Unavailable, and one positive PVP-R05a relationship test with actor/target/game.
Do not add the aura-only class ranges or its earlier opposite relationship
test. Failed eligibility returns 0 without an application request.

On success build a fresh local timed-application request for this candidate,
copying source=actor, target=candidate, selected skill, caller level, prepared
duration, primary identity/transformed value (or absent primary), selected
target state and default removal callback. This is distinct from R05's shared
prepared request. Consume OWD-R06a's existing target/state/monster eligibility.
An eligibility rejection returns null from the applier and 0 from this
recipient, without additional-stat/event requests.

The remaining eligible timed-request policy also applies to R11's aura
request at its independent request boundary; it does not infer a naturally
reachable aura candidate. For a selected state with curse membership, read
target stat 109. A signed value at least 100 returns null before changing
duration or requesting state 57. Otherwise, if resistance is nonzero, replace
the local request duration with wrapping i32
originalDuration-mulDiv(resistance,originalDuration,100), consuming NRF-R05
in that argument order. There is no clamp on negative resistance or duration.
Then test active state 57; true returns null, retaining the changed local
request duration. This mutation never changes the prepared duration used to
build the next candidate's fresh request.

If not rejected, a curse request with an absent or nonextended target base
list selects no existing list. With an extended base list, walk only its
primary chain in stored order and select the first list whose curse condition
is set, even if that list has a different state or skill. The base and
secondary chain are not candidates for this search. For a noncurse request,
skip stat 109 and state 57, and consume RLI-R03's state-list lookup for the
requested state instead. These choices lead to the named OWD-R07a refresh,
OWD-R07b replacement or OWD-R08 creation boundaries. A null applier return from this prefix makes the recipient return 0 without
additional-stat/event requests. Their actual stored
lifecycle effects remain Q04 and the shared owner gates; a selected list is
not a successful application result. No direct draw occurs in this prefix.
R12 still establishes original index-4 rejection before this application;
wider monster-only AI application/removal effects remain Q04.

**MCA-R18** Existing-list refresh and fixed curse-recipient continuation [verified in 1.13c code (static)]

After R17 selects an existing list, including at its independent noncurse
R11 request boundary, when its state, skill and signed level
equal the local request, consume OWD-R07a's refresh path. A nonzero local
duration stores `wrap32(game.frame+localDuration)` as that list's expiry
through STL-R06a, then schedules type 12 for the target at that expiry,
with no override and parameters 0,0 through STL-R01. Positive expiry sets
the list's timed condition; a nonpositive expiry does not clear it.
Duration 0 makes neither expiry write nor scheduling request. After normal
continuation return the same existing list from the applier. No list is
allocated, detached or attached on this path. Keep its own stat values,
source owner, callback, state, skill, level, chain relations and active state
membership unchanged. In particular the requested primary magnitude does
not replace the existing magnitude on an equal-level refresh.

For R05's fixed indices 0 and 1, this non-null applier result reaches the
actual recipient's success continuation: monsterOnly is false, all five
additional stat identities are absent, the auxiliary operation is disabled
and the event descriptors select no registration. Thus no additional stat,
AI, auxiliary, item-event cancellation or registration operation occurs,
and the recipient returns exactly 1. R13 continues traversal and returns 1
normally, which R05 forwards. There is no direct draw beyond R01's selection.
This establishes the existing-list refresh, not its original creation or
later expiration/removal. Replacement/creation and full stored effects
remain Q03 for the aura boundary and Q04 for the curse boundary; other
selected curse branches retain Q04, with R21 establishing the fixed-index-2/3
event suffix after this refresh. STL-Q10 retains malformed/wrapped scheduling
domains; no normal event result is inferred for them. A nonreturning
dependency does not receive an asserted normal continuation result.

**MCA-R19** Reached callback life storage [verified in 1.13c code (static)]

R14's replacing other-life operation and each R16 capped-add replacing
operation reach the same unit-facing writer as
[monster healing](skill-monster-resurrect-and-heal.md) MRH-R08, with the
current recipient, stat 6, layer 0 and the locally computed value. For a
player or monster recipient satisfying MRH-R08's extended root-base,
direct-definition, nonzero request/result and prepared-storage conditions,
consume that exact actual storage contract through STL-R06e/STL-R06d.
Do not replace it by a successful writer fixture or a direct assignment to
the resolved value. The unit's base is the list passed to that writer;
the delta is based on its old own value. This operation makes no new life
clamp. Changed-player bookkeeping excludes stat 6; no changed-stat key is
inserted. No numeric unit-writer return is required or consumed here.

R16's next current-life read uses RLI-R02a/R02b against the resulting arrays.
Thus a capped replacing request can leave resolved life different from the
request or above the maximum used by the local calculation. Its capped-add
return remains R16's computed amount; it is neither the unit writer's return
nor the resolved-life delta. The last linked-unit capped-add return is still
ignored. R14's callback result and R16's normal continuation remain as stated
in their own rules. No direct RNG draw, state/list-chain mutation or stat
notification is introduced by this reached writer.

This does not establish the original creation of the supplied lists/arrays,
other definition/list/ancestry domains, resize/deletion/failure, dependent
formulas, notifications, later updates or release. Those retain Q04 and the
shared owner gates. The earlier combat and later overlay operations keep
their existing boundaries; actual life writes do not supply their effects.

**MCA-R20** Aura preparation and actual prepared collection [verified in 1.13c code (static)]

Action 111 first consumes MTP-R09's target refresh and coordinates. Coordinate
result 0 returns action result 0 before changing the do-handled latch or
requesting collection storage. After success, set the latch true and, for a
present invoking Skills record, compute signed extent
`wrap32(Param4+wrap32(2*callerLevel))`. No positive-level or extent clamp is
made. Begin collection setup with actor, the returned center and this extent.
Its initial context-allocation request precedes the actor-current-room lookup;
later initialization consumes that room. Invalid record storage has no asserted normal result.
Initial allocation, later growth and cleanup remain Q03; no successful
allocation result is supplied or inferred by this rule.

The actual collector below is available with the section-3 prepared context,
initial capacity 15 and accepted counter 0. Its fixed choices are engine
bindings, not configurable filter inputs. The room/unit storage must be
readable and independent of the output storage. If current room is absent,
store output count 0 and return collector result 0, without reading geometry.
Otherwise, with current rectangle `(rx,ry,width,height)`, use only current
room when all four signed, wrapping comparisons hold:

```
wrap32(x-r) > rx
wrap32(y-r) > ry
wrap32(x+r) < wrap32(rx+width)
wrap32(y+r) < wrap32(ry+height)
```

Here `r` is the extent. If any comparison fails, use the current room's
stored ordered room array and unsigned count instead. Do not prepend current
room or remove duplicate entries. Visit that chosen sequence in order, using
R13's exact room-prefilter subrecipe. No room-level town test occurs here.

For each admitted room, start at its unit head. Save each candidate's next
link **before** the predicate below, then continue at that saved link.
Read candidate X then Y: players, monsters and missiles use unsigned dynamic
integer words, with 0 per coordinate for absent path; objects/items/tiles
require readable static coordinates. Compute wrapping signed squared distance
and squared extent; reject when distance is greater. Geometry precedes kind
rejection. Admit only players with mode outside 0/17 and monsters with mode
outside 0/12. A player identical to actor is rejected; a monster identical
to actor is not rejected by identity. No hireling or alive-exclusion-condition
test occurs in this collector.

Next require target flag bit 2, then bit 3. Obtain candidate's current room;
an absent room does not reject. For a present readable room, reject when its
level identity is one of the code-fixed values 1,40,75,103,109. No collision
test, custom predicate, relationship read or state/stat operation occurs.
Each accepted predicate increments the context's accepted counter. Append the
candidate identity to output storage in encounter order, retaining duplicates,
and increment the local accepted count. At accepted count 15, the fifteenth
entry has already been stored: request capacity growth to 30. Effects and
normal continuation across that request remain Q03. With at most 14 accepted
entries, no allocation/growth request is made; finish by storing and returning
the accepted count. Preserve unused slots, backing storage, capacity, world
records, unit state/stats and seeds in this no-growth domain.

The resulting actual collection supplies R10's candidates. It does not verify
initial storage production, cleanup/release, subsequent stored aura effects,
or an engine-wide relation-stability invariant. No direct draw occurs in this
preparation or collector path. Full action effects remain Q03.

**MCA-R21** Fixed curse-event cancellation and registration [verified in 1.13c code (static)]

For R05's fixed indices 2 and 3, an actual non-null existing-list refresh
through R17/R18 reaches this recipient continuation. Monster-only is false,
the five additional stat identities are absent and the auxiliary operation
is disabled. Thus there is no AI, additional-stat or auxiliary request here.

First consume IEP-R05 to remove every candidate-owned event entry with
skill kind and owner identity equal to the selected target state. This key
is the state, not the selected skill, invocation skill or source actor.
It occurs even when the timed list was refreshed rather than created.
Nonmatching kind/identity entries retain their order; executing matches
use IEP-R05's deferred-removal contract rather than an invented immediate free.

Then visit R05's fixed event bindings in their stated order. Resolve each
callback through the shared selector mapping under IEP-R02/IEP-R06 and
create the entry through IEP-R01 on the candidate. Each entry has its stated
event and callback, skill kind, owner identity equal to selected target state,
and callback values `(selected skill identity, caller level)` in that order.
Index 2 creates event 5/selector 4 once. Index 3 creates event 1/selector 5,
then event 2/selector 5; head insertion means event 2 precedes event 1 in
the resulting collection. No unused binding creates an entry. Registration
does not execute a callback or rewrite the refreshed list's source owner,
removal callback or magnitude.

After these operations return normally, produce recipient result 1.
R13's traversal consumes that actual result as already specified. No direct
draw is made. This consumes the verified logical entry creation/removal
contract, not a supplied successful registration or an independently verified
heap allocator. Physical allocation/release internals, failure and wider
state-list creation/replacement/later removal remain Q04 and the shared owner
limits; no complete action parity or natural gameplay producer is inferred.
Later dispatch consumes IEP-R07: selector 5's registered skill/level become
R14's callback inputs, and selector 4 uses R15/R16. Those rules and R19
retain their exact existing read/write and integration boundaries.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MCA-R01 | Fixed random skill selection | verified in 1.13c code (static) | none | MCA-V01, MCA-V02, MCA-V03, MCA-V04, MCA-V05, MCA-V06, MCA-V07 |
| MCA-R02 | Action effects | unresolved | MCA-Q03, MCA-Q04 block their respective handler effects; Q01 compiled integration | none |
| MCA-R03 | Named text-file selectors | verified in 1.13c data | none for named text cells and observed selector correspondence; MCA-Q01 gates other compiled fields | table-data record |
| MCA-R04 | Selected record/state guards and radius/duration order | verified in 1.13c code (static) | MCA-Q01 actual data production; supplied records/results available | MCA-V08 through MCA-V14, MCA-V16 through MCA-V19 |
| MCA-R05 | Fixed request construction and return forwarding | verified in 1.13c code (static) | MCA-Q04 recipient effects only; supplied request/return available | MCA-V08 through MCA-V12, MCA-V15, MCA-V18, MCA-V19, MCA-V20 |
| MCA-R06 | Narrow loaded selected fields | verified in 1.13c data | source-profile limits; Q01 broader production | loaded-field record |
| MCA-R07 | Captured expression evaluation | verified in 1.13c code (static) | Q02 observation only | MCA-V21 |
| MCA-R08 | Filter-3 kind/mode predicate | verified in 1.13c code (static) | Q04 later application; outer traversal is R13 | MCA-V22 |
| MCA-R09 | Own-base resistance adjustment | verified in 1.13c code (static) | named read/classification dependencies | MCA-V23 |
| MCA-R10 | Aura repeated relationship order | verified in 1.13c code (static) | Q03 collection and full state effects | MCA-V24 |
| MCA-R11 | Aura state request and local optional-list return | verified in 1.13c code (static) | Q03 creation/replacement and wider stored effects; independent existing-list refresh consumes R18/OWD-R07a | MCA-V25, MCA-V48 |
| MCA-R12 | Loaded index-4 callback rejection | verified in 1.13c code (static) | Q04 wider callback effects; Q02 observation | MCA-V26 |
| MCA-R13 | Actual recipient traversal and produced return | verified in 1.13c code (static) | Q04 callback effects; Q02 observation; readable world records and normal continuation | MCA-V27, MCA-V28, MCA-V29, MCA-V49, MCA-V52 |
| MCA-R14 | Selector-5 life/overlay requests | verified in 1.13c code (static) | Q04 replacing effects outside R19 and overlay effects; Q02 observation | MCA-V30, MCA-V31, MCA-V42 |
| MCA-R15 | Selector-4 combat preparation and handoff | verified in 1.13c code (static) | Q04 DPO/HRC effects; Q02 observation | MCA-V32, MCA-V33, MCA-V35 |
| MCA-R16 | Class-290 continuation and capped-add requests | verified in 1.13c code (static) | Q04 storage outside R19 and overlay/combat integration; Q02 observation | MCA-V34, MCA-V36 |
| MCA-R17 | Per-recipient local request, curse duration/guard and list selection | verified in 1.13c code (static) | Q04 stored lifecycle/AI effects; Q02 observation | MCA-V37, MCA-V38, MCA-V43 |
| MCA-R18 | Actual existing-list refresh and fixed index-0/1 success continuation | verified in 1.13c code (static) | Q03/Q04 creation/replacement/later expiry and other effects; Q02 observation; STL-Q10 malformed scheduling | MCA-V39, MCA-V48, MCA-V49 |
| MCA-R19 | Actual callback life writes and subsequent reads through the prepared direct domain | verified in 1.13c code (static) | Q04 outside MRH-R08/STL direct domain, combat/overlay integration; Q02 observation | MCA-V40, MCA-V41 |
| MCA-R20 | Aura preparation and actual fixed collection | verified in 1.13c code (static) | Q03 initial allocation/growth/cleanup and stored aura effects; Q02 observation | MCA-V44, MCA-V45, MCA-V46, MCA-V47 |
| MCA-R21 | Actual fixed-index-2/3 skill-event cancellation and registration | verified in 1.13c code (static) | Q04 wider lifecycle/allocator/failure effects; Q02 observation; logical event effects consume IEP | MCA-V50, MCA-V51 |

## 5. Outputs and side effects

| Output | Rule | Evidence boundary |
|---|---|---|
| Actor seed advanced once | MCA-R01 | Supplied actor; entry selection stage |
| Complete stored recipient/effect execution | MCA-R02 | Gated; local requests, predicates and normal traversal are independently established |
| Failure result 0 without expression/recipient calls | MCA-R04 | Invalid selected record or target state; prior R01 seed update retained |
| Actor do-handled latch true; ordered radius/duration requests | MCA-R04 | Valid selected record/state; expression effects belong to SCX |
| Prepared effect request and recipient handoff, forwarded return | MCA-R05 | Recipient writes and event effects excluded |
| Captured profile radius/duration results | MCA-R07 | Derived from complete captured expressions and named parameters |
| Kind/mode accept or reject | MCA-R08 | Candidate predicate only |
| Transformed stat value and optional own-base read | MCA-R09 | No write |
| Repeated relationship reads, skip or state-application request; R11 local optional-list result | MCA-R10, MCA-R11 | Complete action effects excluded; R18 establishes independent existing-list refresh, creation/replacement retains Q03 |
| Callback result 0 and absence of later effect requests | MCA-R12 | Captured index-4 profile; no final action-result inference |
| Target-cache refresh effects; room/unit/callback visit sequence; processor result 0 or 1 | MCA-R13 | Actual traversal at its declared input/normal-return boundary; callback-owned effects remain gated |
| Replacing other-life request; optional prgoverlay operation; callback 0/1 | MCA-R14 | Existing-state/stat input boundary; R19 establishes reached direct life writes, other replacing/overlay effects excluded |
| Cleared combat record, selected fields and ordered apply/reaction requests | MCA-R15 | Incoming record unchanged locally; actual combat/reaction effects retain their owners |
| Ordered capped-add and overlay requests, computed capped-add returns | MCA-R16 | Current post-combat record and sequential stat reads; R19 establishes reached direct writes, other storage excluded |
| Fresh local timed request, local duration replacement, selected existing list or rejected recipient 0 | MCA-R17 | Shared prepared duration unchanged; actual refresh/replacement/creation effects excluded |
| Existing list expiry/timed change, scheduling request, retained metadata and successful recipient/action 1 | MCA-R18 | Equal state/skill/level refresh; recipient/action 1 only for its fixed-index-0/1 curse suffix; stored scheduling consumes STL-R01; creation/replacement/later expiry excluded |
| Actual own/resolved life changes, unchanged other keys/relations and subsequent life reads | MCA-R19 | MRH-R08/STL prepared direct domain; no supplied replacing-write effects; combat and overlay gates retained |
| Aura do-handled latch, extent and collection-storage request; actual candidate sequence/counter/count | MCA-R20 | Supplied invoking record and prepared collection; no allocation result or actual release/effect inferred |
| Removal of matching skill-event entries; new ordered entries, callback values and recipient 1 | MCA-R21 | Actual existing-list refresh and named IEP logical collection contract; no physical allocator/release or natural event-producer proof |

## 6. Branches and edge cases

The selected-record/state guards and signed supplied calculation returns are
established by R04; the request construction is R05. Full effects, actual
profile production outside R06, action-111 storage production and stored
callback effects retain Q01/Q03/Q04. R13 establishes its actual
coordinate/room/unit traversal and return boundary; R08 establishes its
fixed candidate kind/mode predicate. These integration limits do not gate the
supplied record/result stages. There is no local positive-level, radius or
duration clamp. No Classic behavior is asserted.

## 7. Open questions

**MCA-Q01** [unresolved]: R06/R07 close selected target-state, Param1..4/Param7 and complete radius/duration expression correspondence for the dated captured profile. Broader source-profile identity, compiled state metadata masks and other callback data production remain unestablished. The local target-state/radius/duration field names are established; their names are not gated merely because the shared field inventory is partial. This is an integration gate
on original-profile record production, not the supplied-input arithmetic or
record preparation. R14/R15 additionally consume the named callback fields corroborated in the companion; R06 also establishes the dated selected curse-mask membership. Remaining metadata and complete profile identity are still open. AI selection remains outside this boundary.

**MCA-Q02** [unresolved]: Which executable cases match an owned, identified
Expansion 1.13c game? Pending observation does not gate the isolated verified branch,
but does prevent a claim of runtime confirmation.

**MCA-Q03** [unresolved]: Complete action-111 storage production, cleanup and state/list effects. R20 establishes preparation and actual no-growth collection at its supplied prepared-storage boundary; it does not establish the original allocation, growth or release of that storage. R10/R11 close the class, repeated relationship and state-request stages. An engine-wide relation-stability invariant remains unestablished. The shared timed-state consumer's relevant policy and stored effects retain OWD-R06b/R07b, STL-R05/R06b/R06c/STL-Q07 and RLI-Q02/Q03; no successful natural aura case is inferred from its prepared request.

**MCA-Q04** [unresolved]: Complete action-112 generic curse-list execution and callback effects, including their effects on ongoing traversal. R08/R09/R12 establish their stated predicates and captured zero-primary rejection; R13 establishes actual coordinate/room/unit traversal and its produced normal return through the named dependencies. R14-R16 establish local callback arithmetic, exact combat/resource roles and ordered requests; R19 establishes prepared direct life storage and subsequent reads through MRH-R08/STL; other stored combat, resource and overlay effects retain their named owners. R17 establishes the actual selected-state curse-policy prefix and per-candidate request copy; R18 establishes the existing equal-level refresh and fixed index-0/1 successful continuation. R21 establishes fixed-index-2/3 logical event cancellation/registration after refresh through IEP, not physical allocator/release or failure effects. Creation/replacement/later expiry, remaining stored effects and monster-only AI effects retain Q04. The remaining operations and shared OWD/STL/RLI owner gates are listed in the companion. R13 can complete with result 1 despite recipient result 0; no complete stored effect follows from that outer result.

- **MCA-V52** Negative-radius recipient traversal and room fallthroughs (MCA-R08, MCA-R12, MCA-R13; boundary: supplied prepared index-4 request and readable world through actual traversal return)
  - Inputs: enter R13 directly after preparation, with actor A monster 312, game G, absent target cache, stored target words(65535,65535), present dynamic path/current room F. Supply the complete R05 index-4 prepared request from V27 with source A, selected skill 91, caller level 1, state 61, primary stat 67/value 0, duration 500, and signed radius -7. This is a prepared-input traversal case, not an R06 original-profile formula result or full action. F rectangle(65525,65525,40,40) contains the target; its ordered room array is[RX,RY,RA,RB], count 4. Rectangles are RX(65545,65525,-5,40), RY(65525,65545,40,-5), RA(65525,65545,40,10), RB(65545,65545,10,10). RA chain N->O->absent, RB chain P->absent; RX/RY each have a readable nonempty head Q, whose links and coordinates must not be requested. N is a player/mode 1 with absent dynamic path; O is an object with prepared static coordinates(65535,65535); P is a player/mode 1 with present dynamic coordinates(65535,65535). All storage remains readable/fixed, with no external mutation. The actual index-4 recipient rejects N/P at its monster-only guard; no recipient result is supplied. Observe room gates, head reads, candidate Y/X results, kind checks, callbacks and next links.
  - Seeds: no draw; all seeds unchanged.
  - Draws: none.
  - Trace: MTP-R09 returns center(65535,65535); room search finds F. Capture squared radius 49. The prefilter uses x+r=y+r=65528 and x-r=y-r=65542. RX takes the outer else and rejects because 65542>65540. RY takes the first outer arm but its vertical fallthrough rejects because 65542>65540. RA takes that same vertical fallthrough and admits because 65542<=65555. RB takes the outer else, passes 65542<=65555 on the horizontal check, then its vertical fallthrough admits at the same bound. N supplies Y=0 then X=0 locally without dereferencing its absent path: dx=dy=-65535, each wrapped square=-131071, sum=-262142, which passes signed comparison to 49. N passes R08, then actual recipient returns 0 at its player-kind guard. O has distance 0 but R08 rejects its object kind. P has distance 0, passes R08, then actual recipient returns 0. Read each next link after its processing; complete traversal and return 1.
  - Outputs: room-gate log[(RX,false),(RY,false),(RA,true),(RB,true)]; unit-head reads[RA,RB], with no RX/RY head/coordinate/link read. Coordinate-result log[Y(N)=0,X(N)=0,Y(O)=65535,X(O)=65535,Y(P)=65535,X(P)=65535]; callback log[(N,0),(P,0)]; next-link log[callback(N)=0,next(N)=O,next(O)=absent,callback(P)=0,next(P)=absent]. Actual processor result 1, all input records and seeds unchanged. No relationship, timed-state, stat, event, allocation/release or other effect; no action result or natural world/negative-radius producer is asserted.
  - Discriminates: rejecting a negative radius removes the visits; conventional overlap filtering drops RA/RB. Omitting either rejecting fallthrough visits its forbidden Q. Treating an absent dynamic path as failure omits N, while unbounded or unsigned squared comparison also omits its callback. Dynamic-path treatment of O loses its asserted static-coordinate reads; accepting static kinds adds a forbidden callback. X-before-Y or prefetching a next link before its callback violates the exact log. Returning a recipient's 0 rather than completing traversal fails final 1.
  - runtime_observed: no

## 8. Test vectors

### 8.1 Executable vectors

- **MCA-V01** Select index 0 (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1. Stop the assertion boundary immediately after selection, before record validation.
  - Seeds: actor = {lo: 1, hi: 0}.
  - Draws: (actor, modulus 5, result 0).
  - Trace: 1 times 1791398085 plus 0 = 1791398085; new high = 0; new low modulo 5 = 0.
  - Outputs: actor seed = {lo: 1791398085, hi: 0}; selected skill identity = 66.
  - Discriminates: using caller skill 999 instead of the fixed engine mapping fails the selected-identity assertion; skipping the draw on an invalid caller index fails the seed assertion.
  - runtime_observed: no

- **MCA-V02** Select index 1 (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1. Stop the assertion boundary immediately after selection, before record validation.
  - Seeds: actor = {lo: 1, hi: 1}.
  - Draws: (actor, modulus 5, result 1).
  - Trace: 1 times 1791398085 plus 1 = 1791398086; new high = 0; new low modulo 5 = 1.
  - Outputs: actor seed = {lo: 1791398086, hi: 0}; selected skill identity = 72.
  - Discriminates: using caller skill 999 instead of the fixed engine mapping fails the selected-identity assertion; skipping the draw on an invalid caller index fails the seed assertion.
  - runtime_observed: no

- **MCA-V03** Select index 2 (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1. Stop the assertion boundary immediately after selection, before record validation.
  - Seeds: actor = {lo: 1, hi: 2}.
  - Draws: (actor, modulus 5, result 2).
  - Trace: 1 times 1791398085 plus 2 = 1791398087; new high = 0; new low modulo 5 = 2.
  - Outputs: actor seed = {lo: 1791398087, hi: 0}; selected skill identity = 82.
  - Discriminates: using caller skill 999 instead of the fixed engine mapping fails the selected-identity assertion; skipping the draw on an invalid caller index fails the seed assertion.
  - runtime_observed: no

- **MCA-V04** Select index 3 (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1. Stop the assertion boundary immediately after selection, before record validation.
  - Seeds: actor = {lo: 1, hi: 3}.
  - Draws: (actor, modulus 5, result 3).
  - Trace: 1 times 1791398085 plus 3 = 1791398088; new high = 0; new low modulo 5 = 3.
  - Outputs: actor seed = {lo: 1791398088, hi: 0}; selected skill identity = 87.
  - Discriminates: using caller skill 999 instead of the fixed engine mapping fails the selected-identity assertion; skipping the draw on an invalid caller index fails the seed assertion.
  - runtime_observed: no

- **MCA-V05** Select index 4 (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1. Stop the assertion boundary immediately after selection, before record validation.
  - Seeds: actor = {lo: 1, hi: 4}.
  - Draws: (actor, modulus 5, result 4).
  - Trace: 1 times 1791398085 plus 4 = 1791398089; new high = 0; new low modulo 5 = 4.
  - Outputs: actor seed = {lo: 1791398089, hi: 0}; selected skill identity = 91.
  - Discriminates: using caller skill 999 instead of the fixed engine mapping fails the selected-identity assertion; skipping the draw on an invalid caller index fails the seed assertion.
  - runtime_observed: no

- **MCA-V06** Multiply high and carry (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1; stop after selection before record validation.
  - Seeds: actor = {lo: 3, hi: 4294967295}.
  - Draws: (actor, modulus 5, result 3).
  - Trace: 3 times 1791398085 = 5374194255; adding 4294967295 gives 9669161550; split at 4294967296 gives low 1079226958, high 2; low modulo 5 = 3.
  - Outputs: actor seed = {lo: 1079226958, hi: 2}; selected skill identity = 87.
  - Discriminates: retaining old high, writing zero high, or omitting addition carry fails high = 2; reducing the unsplit product fails selected identity.
  - runtime_observed: no

- **MCA-V07** Unsigned reduction before rejected selected record (MCA-R01; boundary: supplied-input stage)
  - Inputs: action selector = 112; actor exists; caller skill index = 999; caller level = 1; selected skill collection has count 0, so skill 66 is invalid when validation eventually begins. A logical observer records selected-record validation and record accesses; its log is initially empty. Stop immediately after selection, before validation.
  - Seeds: actor = {lo: 2, hi: 0}.
  - Draws: (actor, modulus 5, result 0).
  - Trace: 2 times 1791398085 plus 0 = 3582796170; split gives low 3582796170, high 0; unsigned low modulo 5 = 0; fixed identity at index 0 is 66; no selected-record validation has yet occurred at this boundary.
  - Outputs: actor seed = {lo: 3582796170, hi: 0}; selected skill identity = 66; selected-record validation/access log empty. No assertion is made about the subsequent return or effects.
  - Discriminates: interpreting low as signed gives -712171126 and a negative signed remainder, failing index 0/identity 66; validating selected record before drawing either logs an early validation or rejects without advancing seed, failing the asserted boundary state.
  - runtime_observed: no

- **MCA-V08** Prepare selected index 0 (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry immediately after R01 with index 0, selected skill 66; actor/game present, latch false; caller identity 999, caller level 2; skill count 400 and selected record present; selected auratargetstate = 42, state count 100; selected Param7 = -12. Selected radius reference R and duration reference D are distinct. Calculation operations log context and return R=13, D=29 without draws. Recipient operation logs filter, actor, radius and request, draws nothing, changes no state and returns 1 normally.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is outside this vector.
  - Draws: none.
  - Trace: record/state guards pass; set latch; evaluate R then D using selected identity 66 and level 2; initialize default request and choose row 0; issue recipient request and forward 1.
  - Outputs: latch true; ordered calls = [radius(actor,66,2), duration(actor,66,2), recipient(filter=3,actor,radius=13,request)]; request game/actor are the supplied context, skill=66, level=2, duration=29, targetState=42, primaryStat=36, primaryValue=-100, additionalStats=[], monsterOnly=false, eventBindings=[]; no other stat/event binding; final result=1.
  - Discriminates: using caller identity 999 fails both evaluation contexts and request identity; swapping duration/radius fails the ordered log and radius 13/duration 29; selecting a neighboring row fails its exact stat/event request assertion.
  - runtime_observed: no

- **MCA-V09** Prepare selected index 1 (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry immediately after R01 with index 1, selected skill 72; actor/game present, latch false; caller identity 999, caller level 2; skill count 400 and selected record present; selected auratargetstate = 42, state count 100; selected Param7 = -12. Selected radius reference R and duration reference D are distinct. Calculation operations log context and return R=13, D=29 without draws. Recipient operation logs filter, actor, radius and request, draws nothing, changes no state and returns 1 normally.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is outside this vector.
  - Draws: none.
  - Trace: record/state guards pass; set latch; evaluate R then D using selected identity 72 and level 2; initialize default request and choose row 1; issue recipient request and forward 1.
  - Outputs: latch true; ordered calls = [radius(actor,72,2), duration(actor,72,2), recipient(filter=3,actor,radius=13,request)]; request game/actor are the supplied context, skill=72, level=2, duration=29, targetState=42, primaryStat=25, primaryValue=-50, additionalStats=[], monsterOnly=false, eventBindings=[]; no other stat/event binding; final result=1.
  - Discriminates: using caller identity 999 fails both evaluation contexts and request identity; swapping duration/radius fails the ordered log and radius 13/duration 29; selecting a neighboring row fails its exact stat/event request assertion.
  - runtime_observed: no

- **MCA-V10** Prepare selected index 2 (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry immediately after R01 with index 2, selected skill 82; actor/game present, latch false; caller identity 999, caller level 2; skill count 400 and selected record present; selected auratargetstate = 42, state count 100; selected Param7 = -12. Selected radius reference R and duration reference D are distinct. Calculation operations log context and return R=13, D=29 without draws. Recipient operation logs filter, actor, radius and request, draws nothing, changes no state and returns 1 normally.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is outside this vector.
  - Draws: none.
  - Trace: record/state guards pass; set latch; evaluate R then D using selected identity 82 and level 2; initialize default request and choose row 2; issue recipient request and forward 1.
  - Outputs: latch true; ordered calls = [radius(actor,82,2), duration(actor,82,2), recipient(filter=3,actor,radius=13,request)]; request game/actor are the supplied context, skill=82, level=2, duration=29, targetState=42, primaryStat=none, primaryValue=none, additionalStats=[], monsterOnly=false, eventBindings=[(5, 4)]; no other stat/event binding; final result=1.
  - Discriminates: using caller identity 999 fails both evaluation contexts and request identity; swapping duration/radius fails the ordered log and radius 13/duration 29; selecting a neighboring row fails its exact stat/event request assertion.
  - runtime_observed: no

- **MCA-V11** Prepare selected index 3 (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry immediately after R01 with index 3, selected skill 87; actor/game present, latch false; caller identity 999, caller level 2; skill count 400 and selected record present; selected auratargetstate = 42, state count 100; selected Param7 = -12. Selected radius reference R and duration reference D are distinct. Calculation operations log context and return R=13, D=29 without draws. Recipient operation logs filter, actor, radius and request, draws nothing, changes no state and returns 1 normally.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is outside this vector.
  - Draws: none.
  - Trace: record/state guards pass; set latch; evaluate R then D using selected identity 87 and level 2; initialize default request and choose row 3; issue recipient request and forward 1.
  - Outputs: latch true; ordered calls = [radius(actor,87,2), duration(actor,87,2), recipient(filter=3,actor,radius=13,request)]; request game/actor are the supplied context, skill=87, level=2, duration=29, targetState=42, primaryStat=none, primaryValue=none, additionalStats=[], monsterOnly=false, eventBindings=[(1, 5), (2, 5)]; no other stat/event binding; final result=1.
  - Discriminates: using caller identity 999 fails both evaluation contexts and request identity; swapping duration/radius fails the ordered log and radius 13/duration 29; selecting a neighboring row fails its exact stat/event request assertion.
  - runtime_observed: no

- **MCA-V12** Prepare selected index 4 (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry immediately after R01 with index 4, selected skill 91; actor/game present, latch false; caller identity 999, caller level 2; skill count 400 and selected record present; selected auratargetstate = 42, state count 100; selected Param7 = -12. Selected radius reference R and duration reference D are distinct. Calculation operations log context and return R=13, D=29 without draws. Recipient operation logs filter, actor, radius and request, draws nothing, changes no state and returns 1 normally.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is outside this vector.
  - Draws: none.
  - Trace: record/state guards pass; set latch; evaluate R then D using selected identity 91 and level 2; initialize default request and choose row 4; issue recipient request and forward 1.
  - Outputs: latch true; ordered calls = [radius(actor,91,2), duration(actor,91,2), recipient(filter=3,actor,radius=13,request)]; request game/actor are the supplied context, skill=91, level=2, duration=29, targetState=42, primaryStat=67, primaryValue=-12, additionalStats=[(25, -50), (36, -50), (68, -12)], monsterOnly=true, eventBindings=[]; no other stat/event binding; final result=1.
  - Discriminates: using caller identity 999 fails both evaluation contexts and request identity; swapping duration/radius fails the ordered log and radius 13/duration 29; selecting a neighboring row fails its exact stat/event request assertion.
  - runtime_observed: no

- **MCA-V13** Reject invalid selected state (MCA-R04; boundary: supplied-input stage)
  - Inputs: entry after R01; index 0, selected identity 66, actor present, latch false, caller level 2, skill count 400; selected record present with auratargetstate=-1; state count 100; expression and recipient logs empty and record reads observed. All later parameters are uninspected.
  - Seeds: no draw inside this post-selection stage; R01 already updated the seed.
  - Draws: none.
  - Trace: selected state is negative, so return before latch, expression or recipient operations.
  - Outputs: result=0, latch false, expression/recipient call logs empty; no additional seed change.
  - Discriminates: proceeding with an invalid state would set latch or issue an operation, violating the asserted unchanged latch/empty logs.
  - runtime_observed: no

- **MCA-V14** Reject invalid selected record (MCA-R04; boundary: supplied-input stage)
  - Inputs: entry after R01; index 0, selected identity 66, actor present, latch false, caller level 2, skill count 400; selected record absent; state count 100; expression and recipient logs empty and record reads observed. All later parameters are uninspected.
  - Seeds: no draw inside this post-selection stage; R01 already updated the seed.
  - Draws: none.
  - Trace: selected record is absent, so return before latch, expression or recipient operations.
  - Outputs: result=0, latch false, expression/recipient call logs empty; no additional seed change.
  - Discriminates: proceeding with an invalid record would set latch or issue an operation, violating the asserted unchanged latch/empty logs.
  - runtime_observed: no

- **MCA-V15** Forward recipient failure result (MCA-R05; boundary: supplied-input stage)
  - Inputs: prepared index-0 context from V08 as independent inputs: actor/game present, skill 66, level 2, state 42, radius 13, duration 29; primary(36,-100), no additional stats/events, monsterOnly=false; recipient operation logs this request, changes no state, draws nothing and returns 0 normally.
  - Seeds: no draw.
  - Draws: none.
  - Trace: issue recipient request and forward its return 0.
  - Outputs: recipient log contains exactly (filter 3,actor,radius 13,prepared request); final result 0; no local seed change.
  - Discriminates: unconditionally succeeding after recipient processing returns 1 and fails the final-result assertion.
  - runtime_observed: no

- **MCA-V16** Selected identity equals skill count (MCA-R01, MCA-R04; boundary: supplied-input stage)
  - Inputs: action selector = 112; game/actor present; caller skill identity 999, level 2; loaded skill count = 66; actor latch false; record-read, expression and recipient logs initially empty. A logical observer records access to individual skill records, separately from reading the loaded count. All state/effect fields are uninspected.
  - Seeds: actor = {lo: 1, hi: 0}.
  - Draws: (actor, modulus 5, result 0).
  - Trace: 1 times 1791398085 plus 0 gives seed {1791398085,0}; low modulo 5 selects index 0 and identity 66; 66 is at least skill count 66, so return before individual record access.
  - Outputs: actor seed = {lo: 1791398085, hi: 0}; result = 0; latch false; individual-record-read, expression and recipient logs empty.
  - Discriminates: allowing equality at the upper bound reaches individual record lookup; rejecting before drawing leaves the entry seed unchanged. Either fails the boundary assertions.
  - runtime_observed: no

- **MCA-V17** Target state equals state count (MCA-R04; boundary: supplied-input stage)
  - Inputs: entry immediately after R01; actor present, latch false; index 0, selected identity 66; caller level 2; skill count 400 and selected record present; selected auratargetstate = 100, loaded state count = 100; expression and recipient logs empty. Later parameter fields are uninspected.
  - Seeds: no draw inside this post-selection stage; the preceding R01 update is already complete.
  - Draws: none.
  - Trace: selected skill record is valid; selected state 100 is at least count 100; return before latch or expression/recipient operations.
  - Outputs: result = 0; latch false; expression and recipient logs empty; no additional seed change.
  - Discriminates: accepting equality at the state bound sets latch or issues an operation, violating the asserted unchanged latch and empty logs.
  - runtime_observed: no

- **MCA-V18** Preserve negative radius and duration (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry after R01; game/actor present; latch false; index 0, selected identity 66; caller identity 999, level 2; skill count 400, record present, selected target state 42, state count 100; selected Param7 = -12; distinct radius/duration references R,D. Calculation operations log their contexts, draw nothing, and return R=-3, D=-7. Recipient operation logs its arguments/request, changes no state, draws nothing and returns 1 normally.
  - Seeds: no draw inside this post-selection stage.
  - Draws: none.
  - Trace: valid record/state; set latch; request radius then duration and retain signed returns -3 and -7; prepare index-0 request and issue recipient processing without clamping or omitting it; forward 1.
  - Outputs: latch true; ordered log = [radius(actor,66,2), duration(actor,66,2), recipient(filter=3,actor,radius=-3,request)]; request game/actor are the supplied context, skill=66, level=2, targetState=42, duration=-7, primaryStat=36, primaryValue=-100, additionalStats=[], monsterOnly=false, eventBindings=[]; no other binding; final result=1; no local seed change.
  - Discriminates: clamping signed results to positive values, rejecting nonpositive radius/duration, or omitting the recipient operation fails the exact logged radius/duration and recipient call assertion. No actual recipient behavior is inferred from these supplied returns.
  - runtime_observed: no

- **MCA-V19** Preserve zero radius and duration (MCA-R04, MCA-R05; boundary: supplied-input stage)
  - Inputs: entry after R01; game/actor present; latch false; index 0, selected identity 66; caller identity 999, level 2; skill count 400, record present, selected target state 42, state count 100; selected Param7 = -12; distinct radius/duration references R,D. Calculation operations log their contexts, draw nothing, and return R=0, D=0. Recipient operation logs its arguments/request, changes no state, draws nothing and returns 1 normally.
  - Seeds: no draw inside this post-selection stage.
  - Draws: none.
  - Trace: valid record/state; set latch; request radius then duration and retain signed returns 0 and 0; prepare index-0 request and issue recipient processing without clamping or omitting it; forward 1.
  - Outputs: latch true; ordered log = [radius(actor,66,2), duration(actor,66,2), recipient(filter=3,actor,radius=0,request)]; request game/actor are the supplied context, skill=66, level=2, targetState=42, duration=0, primaryStat=36, primaryValue=-100, additionalStats=[], monsterOnly=false, eventBindings=[]; no other binding; final result=1; no local seed change.
  - Discriminates: clamping signed results to positive values, rejecting nonpositive radius/duration, or omitting the recipient operation fails the exact logged radius/duration and recipient call assertion. No actual recipient behavior is inferred from these supplied returns.
  - runtime_observed: no

- **MCA-V20** Forward a non-Boolean processor result (MCA-R05; boundary: supplied-input stage)
  - Inputs: prepared index-0 context as independent inputs: actor/game present, selected skill 66, caller level 2, targetState=42, radius=13, duration=29; primaryStat=36, primaryValue=-100, additionalStats=[], monsterOnly=false, eventBindings=[]; recipient operation logs this request, changes no state, draws nothing and returns signed 32-bit integer 37 normally.
  - Seeds: no draw.
  - Draws: none.
  - Trace: issue recipient processing with filter 3, actor, radius 13 and the prepared request; forward its normally returned integer 37 unchanged.
  - Outputs: recipient log contains exactly (filter=3,actor,radius=13,prepared request); final result=37; no local seed or state change.
  - Discriminates: Booleanizing a nonzero result gives 1 and fails final result 37. This supplied fixture checks caller forwarding and makes no claim that the actual processor produces 37.
  - runtime_observed: no

- **MCA-V21** Captured formula evaluation (MCA-R06, MCA-R07; boundary: isolated rule)
  - Inputs: independent captured selected records in R06; complete radius expression parameter 0 then terminator, duration expression parameter 2 then terminator; SCX skill context uses that selected identity. Test L=1,3,0,-1 for each of the five records. Unit has no requested stat or skill dependency.
  - Seeds: no draw.
  - Draws: none.
  - Trace: SCX-R21 returns the corresponding linear pair for positive L, otherwise 0; evaluate radius then duration.
  - Outputs: ordered (radius,duration) at L=1 / L=3: skill 66 (3,200)/(5,350); skill 72 (9,350)/(11,470); skill 82 (4,400)/(6,520); skill 87 (6,100)/(6,130); skill 91 (7,500)/(9,600). Every L=0/-1 case returns (0,0). No seed mutation or effect request.
  - Discriminates: invocation-record substitution using skill 212 would give duration 300 at L=1; using L instead of L-1 changes at least the nonzero-increment pairs; omitting the nonpositive-level guard fails L=0/-1.
  - runtime_observed: no

- **MCA-V22** Filter-3 kinds and modes (MCA-R08; boundary: isolated rule)
  - Inputs: independent candidates (player,mode 0),(player,17),(player,1),(monster,0),(monster,12),(monster,1),(object,1),(missile,1),(item,1),(tile,1),(player,-1),(monster,-1),(player,12),(monster,17). Fixed filter 3; no geometry, relation or state operation is entered by this isolated predicate.
  - Seeds: no draw.
  - Draws: none.
  - Trace: test kind, then the corresponding two excluded modes.
  - Outputs: false,false,true,false,false,true,false,false,false,false,true,true,true,true in input order; dependency log empty; no mutation.
  - Discriminates: permitting every live unit fails the four other-kind cases; using player death mode 17 for monsters incorrectly admits monster 12. Rejecting every negative mode fails the two -1 cases; sharing both kinds' excluded mode sets fails player 12/monster 17.
  - runtime_observed: no

- **MCA-V23** Own-base immune reduction (MCA-R09; boundary: isolated rule)
  - Inputs: independent alive nonhireling monster class 141, statCount512, present stat definition with fMin=false and base own layer 0 stat 36=100, resolved 36=20. Request stat 36 values -100,-1,-101 in independent cases. Variants: base 36=99/resolved 36=200 with value-100; player with value-100; recognized hireling class 271 with value-100; same initial monster with stat 38/value-100; stat 37/value-100 with own 37=100; initial monster stat 36/value 1. All unused dependencies log no requests.
  - Seeds: no draw.
  - Draws: none.
  - Trace: three initial cases read own 36=100 and divide by 5 toward zero; base 99 case reads own 36=99 and retains value; player/hireling/stat 38/positive cases do not read own stats; stat 37 reads own 37=100.
  - Outputs: returns -20,0,-20,-100,-100,-100,-100,-20,1 in input order. Exactly one own-base request in each of cases 1,2,3,4,8; none otherwise; no write.
  - Discriminates: reading resolved totals reverses the initial and base 99 branch choices; floor division gives -21 for -101; treating stat 38 as selected gives -20; applying the reduction to players/hirelings changes their returns.
  - runtime_observed: no

- **MCA-V24** Aura opposite relationship tests (MCA-R10; boundary: supplied-input stage)
  - Inputs: twelve independent supplied candidate/actor pairs, each consisting of distinct monster units with no owners and a present game. Candidate class 141, mode 1, aliveExclusionCondition=false, MonStats2.mWL=true, monster target exclusion=false, flags bits 1/2/3=true. Case(a) actor/candidate alignment tiers 1/1; case(b) tiers 0/0; case(c) as(b) but candidate class 146; case(d) as(b) but required candidate flag bit 3=false and actor mode 0/aliveExclusionCondition=true; case(e) as(b) but mWL=false; case(f) as(b) but monster target exclusion=true; case(g) as(b) but candidate mode 0; case(h) as(b) but candidate monster data absent, supplying exclusion=false. Cases(i/j/k/l) use(b) with candidate classes 396/400/395/401 respectively, retaining the supplied valid record/mWL conditions. Actor mode/exclusion conditions are otherwise arbitrary unused context. Existing alignment lists/read definitions satisfy PVP-R05a and RLI-R02a/R02b, with fMin=false. Graph and all read inputs remain unchanged through each stage.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) PVP-R05a returns true, skip before class/target eligibility; (b) first relationship false, class/flags/availability pass, second relationship false, skip; (c) first false then class outside both ranges, skip. Cases(d/e/f/g) first relationship false, then reject at flag/mWL/exclusion/Unavailable respectively, before the second relationship. Case(h) missing monster data does not reject the exclusion check; other target predicates pass and second relationship is false. Cases(i/j) pass the second class range and reach the second false relationship; cases(k/l) reject at the adjacent class bounds.
  - Outputs: relationship request counts 1,2,1,1,1,1,1,2,2,2,1,1, each with identical (actor,candidate,game) identities. In(d) the target-eligibility log is [mWL=true,exclusion=false,requiredFlags=false], then skip without candidate Unavailable or second relationship; actor availability is never requested. State-application log empty in all cases; no graph, seed or unit mutation.
  - Discriminates: removing the first opposite test adds eligibility reads in(a); accepting the second false result emits an effect in(b); extending the first class range through 146 adds the second relationship in(c). Moving the second relationship before required flags yields two relationship requests in(d), failing count 1; requiring actor availability would wrongly stop before its first relationship. Replacing mWL/exclusion/Unavailable by true adds a second relationship in(e/f/g); rejecting absent monster data removes the second relationship in(h). Omitting or shifting the second class range fails at least one of(i/j/k/l).
  - runtime_observed: no

- **MCA-V25** Aura request arithmetic (MCA-R11; boundary: isolated rule)
  - Inputs: isolated request-builder entry after its caller's tests, not a claimed naturally reachable successful candidate. Present invoking skill 177 record Params1/2/3=10/50/5, source actor A, target T, game G, default callback. Independent levels 3,0,-2,2147483647. Application operation logs the request and normally returns null; no stored effects are supplied or asserted.
  - Seeds: no draw.
  - Draws: none.
  - Trace: compute wrapping duration 50*L and value 10+(L-1)*5; request generic timed state 68/stat 25 with the unchanged source,target,skill,level and default callback.
  - Outputs: exact request (G,A,T,177,L,duration,25,value,68,default), with (duration,value)=(150,20),(0,5),(-100,-5),(-50,-2147483648) respectively; no other request or mutation. The builder return and whole action return are not asserted.
  - Discriminates: using Param4 rather than Param3 gives value 30 at L=3; a positive-level guard suppresses the L=0 request; saturation fails the wrapping case.
  - runtime_observed: no

- **MCA-V26** Captured zero-primary callback (MCA-R12; boundary: supplied-input stage)
  - Inputs: R05 index 4 request from the R06 selected 91 record, Param7=0, state 61, radius 7, duration 500, caller level 1, source A and game G. Case(a) candidate is a player; case(b) candidate monster class 141, mode 1, aliveExclusionCondition=false, MonStats/MonStats2 counts 800 with class 141 records, canonical class 141 and MonStatsEx 141, mWL=true, MonStats.boss=false, MonStats.switchai=true (GDA-R16), monster data present with condition word 0, target flags bit 2=true. State storage exists with count 200 and no active states; the existing state 105 alignment list is nonextended and has own layer 0 stat 172=0, a present definition with fMin=false and statCount512. Independent variants(c/d/e) use(b) but respectively activate state 54, set target flag bit 2=false, or replace the alignment own stat 172 value by 1. Variant(f) uses(b) with canonical-base link 492 and active state 143. Variant(g) uses(b) with an extended state-105 list, own stat 172=0 and selected resolved stat 172=1. A read observer logs only the named semantic read/check events below when consumed. Loaded-count validation inside a named record event is included in that event; hardware prefetches are not logged. `canonicalLink` is the first MonStats use; `switchRecord` and `modeRecord` are the later switch-AI eligibility checks on the candidate class and its linked extended record. Other state/list/event dependencies log calls but are not given successful results.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) monster-only kind guard rejects; (b) alignment 0, canonical 141, no state 54, mWL/nonboss/switchai pass, true target flag 2 bypasses Unavailable, conditionWord AND 10=0, so control 0 passes; transform stat 67/value 0 remains 0 and return. (c) reject at active state 54 before the loaded switch-AI checks; (d) flag 2=false requests Unavailable=false for its alive monster mode 1, so reject before condition-mask/primary checks; (e) alignment 1 rejects before canonical/state/control checks; (f) canonical 492 with active state 143 rejects before state 54/switch-record checks; (g) RLI-R02a reads selected resolved alignment 1 and rejects like(e), despite own alignment 0.
  - Outputs: callback result 0 in all seven cases. Primary-transform requests: none,(67,0),none,none,none,none,none in input order. Exact logical read/request logs are (a) [kind=player]; (b) [kind=monster,alignment=0,canonicalLink(141)=141,stateActive(54)=false,switchRecord(141)=valid,modeRecord(141)=valid,mWL=true,boss=false,switchai=true,targetFlag2=true,conditionWordAnd10=0,primaryTransform(67,0)=0]; (c) [kind=monster,alignment=0,canonicalLink(141)=141,stateActive(54)=true]; (d) the(b) prefix through switchai=true, then [targetFlag2=false,Unavailable(candidate)=false]; (e) [kind=monster,alignment=1]; (f) [kind=monster,alignment=0,canonicalLink(141)=492,stateActive(143)=true]; (g) [kind=monster,alignment=1]. No further named log event occurs after a rejection. Later relationship, timed-state, additional-stat, event-cancel and event-register logs are empty; no mutation. Outer recipient/action return is not asserted.
  - Discriminates: ignoring monster-only eligibility reaches the transform in(a); treating a zero primary as absent or applying later additional stats reaches a forbidden later request in(b). Skipping state 54/alignment 1 rejection adds a primary transform in(c/e); reversing the conditional availability gate admits(d). Evaluating switchRecord/modeRecord before state 54 adds forbidden reads in(c); evaluating either definition check after failing alignment adds forbidden reads in(e). Skipping the canonical 492/state 143 rejection adds state 54 and later checks in(f). Substituting own-array alignment for the selected-array read adds forbidden later checks in(g).
  - runtime_observed: no

- **MCA-V27** Actual index-4 traversal produces success (MCA-R01, MCA-R04 through MCA-R08, MCA-R12, MCA-R13; boundary: supplied world-record stage)
  - Inputs: action 112, invoking skill 212, caller level 1; actor A is a monster class 312 with a valid dynamic path, latch=false, position(95,95), stored target words(100,100), absent target cache and present game/empty registry. Loaded Skills count is 357 and selected record 91 has R06's captured fields. Supply a 5891-byte pool: bytes at 2065 are 04 02 00, at 2068 are 04 00 00, and every other byte is 0. These two complete expressions match the captured slices; no whole original-pool identity is claimed or needed. Supplied state count is 200. Initial room R0 rectangle(90,90,40,40) contains both A and the target point; its stored ordered room array is[R0,R1], count 2. R1 rectangle(130,90,40,40). R0 unit chain A->T->U->V->absent; R1 head P->absent. T,U,V,P are players in mode 1 with readable dynamic paths at(103,104),(108,100),(100,107),(133,100), respectively. Readable storage stays fixed; R12's actual player-kind rejection runs as the callback. No successful recipient result or mutation is supplied. An observer logs room gates, candidate Y/X reads, distance, predicate/callback and next-link reads.
  - Seeds: A={lo:1,hi:4}; other unit seeds unchanged.
  - Draws: (A,modulus 5,result 4), exactly once; calculations/traversal/callback make no further draw.
  - Trace: R01 selects 91, R04 sets latch and evaluates radius 7 then duration 500. R05 builds its exact index 4 request with state 61 and primary 67=0. MTP-R09 returns center(100,100), with no refresh write/registry lookup because cache is absent; room search finds R0 immediately. R13 captures `rSquared=49`. Its prefilter admits both R0 and R1, including R1 outside the radius's conventional rectangle overlap. In chain order, candidate distances are 50,25,64,49,1089. Only T and V pass distance and R08; each actual R12 callback rejects its player kind and returns 0. Normal traversal completes and R05 forwards 1.
  - Outputs: A.seed={lo:1791398089,hi:0}; selected identity 91 is intermediate; latch=true; final action result 1. Room-gate log[(R0,true),(R1,true)]. Candidate coordinate reads in order[Y(A)=95,X(A)=95,Y(T)=104,X(T)=103,Y(U)=100,X(U)=108,Y(V)=107,X(V)=100,Y(P)=100,X(P)=133]. Next-link sequence[next(A)=T,callback(T)=0,next(T)=U,next(U)=V,callback(V)=0,next(V)=absent,next(P)=absent]; callback log[T,V]. No relationship, timed-state, additional-stat or event operation; no state-membership, stat or list change; the stated A seed/latch writes remain.
  - Discriminates: using A's current position instead of target point changes accepted candidates; strict distance<49 drops V; returning the first callback 0 suppresses V/final 1; conventional rectangle-overlap filtering omits the R1 gate/distance visit. Prefetching a next link before its callback changes the asserted next-link log.
  - runtime_observed: no

- **MCA-V28** Actual coordinate failure after preparation (MCA-R01, MCA-R04 through MCA-R07, MCA-R13; boundary: supplied world-record stage)
  - Inputs: independent V27 inputs but A's stored target words are(100,0), cache remains absent. All visited storage is readable; no recipient result is supplied.
  - Seeds: A={lo:1,hi:4}; other seeds unchanged.
  - Draws: (A,modulus 5,result 4), exactly once.
  - Trace: selection/preparation still yields skill 91, latch=true, radius 7/duration 500 and the same request. MTP-R09 writes center outputs(100,0), returns 0, and the actual outer processor returns 0 before requesting actor room or room search.
  - Outputs: A.seed={lo:1791398089,hi:0}; latch=true; final action result 0. Actor-room, room-search, room-array, distance and callback logs empty; no state-membership/stat/list/event mutation or cache-clear write; the stated A seed/latch writes remain.
  - Discriminates: testing target coordinates before selection/preparation wrongly suppresses the seed/latch writes; treating one nonzero coordinate as success reaches forbidden room requests.
  - runtime_observed: no

- **MCA-V29** Actual containing-room failure and empty traversal (MCA-R01, MCA-R04 through MCA-R07, MCA-R13; boundary: supplied world-record stage)
  - Inputs: two independent V27 variants with target(100,100) and absent cache. Case(a) A's initial current room is absent. Case(b) current room R0 contains the point but its stored room array is empty/count 0. No recipient result is supplied.
  - Seeds: A={lo:1,hi:4} in each case; other seeds unchanged.
  - Draws: (A,modulus 5,result 4), once per case.
  - Trace: each case completes preparation/coordinate resolution. (a) The narrow MTP-R04 room search returns no match for absent initial room; processor returns 0 without a linked-monster fallback. (b) Room search finds R0, then array count 0 skips traversal and processor returns 1.
  - Outputs: A.seed={lo:1791398089,hi:0}, latch=true in both; final action results 0/1 in input order. Case(a) has no room-array/distance/callback request; case(b) reads only R0's empty array/count after room search. Both callback logs empty; no state-membership/stat/list/event change; the stated A seed/latch writes remain.
  - Discriminates: requiring at least one recipient changes(b) to 0; applying the direct-teleport caller's fallback after failed room search adds forbidden operations in(a).
  - runtime_observed: no

- **MCA-V30** Actual index-3 callback with zero calculation (MCA-R14; boundary: existing state/stat views through requests and return)
  - Inputs: event owner O is a player, mode 0; other T is a monster class 141, mode 1, aliveExclusionCondition=false. Incoming logical combat record has physical=512; its other fields are zero. Registered skill 87, level -2, skill count 357; selected captured record has target state 60, `calc1` reference 4294967295 and prgoverlay=-1. State count 200; O has active state 60 and an existing nonextended list with source kind monster/id 7; game registry has no such source. T's existing current life=1024, effective maximum=768 under RLI-R02a/R02b with fMin=false. The replacing operation logs its request, returns normally and makes no mutation or draw. Overlay count 200; all other logs initially empty.
  - Seeds: no draw; all unit seeds unchanged.
  - Draws: none.
  - Trace: both kinds pass; only T availability is tested. Normalize level to 1; state/list pass; source lookup returns absent. SCX-R11 returns 0 for empty reference without bytecode evaluation. `mulDiv(0,512,100)=0`; positive current/maximum pass; sum 1024 caps to 768. Request replacing T's life, omit negative prgoverlay, return 1.
  - Outputs: calculation-front-end log[(absent source,87,1,empty reference)]; bytecode-evaluation log empty; ordered resource reads[life(T)=1024,maxLife(T)=768]; replacing log[(T,stat 6,layer 0,768)]; callback result 1. No owner availability request, overlay request, incoming-record change or fixture-view mutation; no state/list/event change.
  - Discriminates: returning early for zero percentage omits the replacing request; using owner as target writes the wrong unit; rejecting owner mode 0 suppresses the valid local callback; failing the maximum clamp requests 1024. This does not establish stored-write success or a gameplay event.
  - runtime_observed: no

- **MCA-V31** Life callback level, signed arithmetic and overlay order (MCA-R14; boundary: supplied record/existing stat views through requests)
  - Inputs: independent cases with present player O and monster T class 141/mode 1, aliveExclusionCondition=false; valid registered skill 87 record/count 357, target state 60 active with existing list/count 200 and source S found in the game registry. Supplied selected record has `calc1` reference 0, Param5=50, Param6=0 and prgoverlay 12; complete three-byte pool[4,4,0]. T life 300/max 1000; definitions have fMin=false. Overlay count 20. Replacing and overlay operations log ordered requests, return normally, make no mutation/draw. Cases(a) level 0; (b) level-1; (c) as(b) but Param5=-100; (d) as(b) but T life 2147483600/max 2147483647; (e) as(b) but life 0; (f) as(b) but maximum 0. Incoming physical 512, other fields 0. Other dependencies have no requested fields/functions.
  - Seeds: no draw.
  - Draws: none.
  - Trace: normalized levels 0,1,1,1,1,1; SCX parameter 4 gives percentages 0,50,-100,50,50,50. Amounts0,256,-512,256,256,256. Cases(a-d) wrap/add, floor at 1 then cap; (e/f) reject after both resource reads.
  - Outputs: calculation contexts all(S,87,normalized level); results 0,50,-100,50,50,50. Ordered logs(a)[replace(T,6,0,300),overlay(T,12)]; (b)[replace(T,6,0,556),overlay(T,12)]; (c)[replace(T,6,0,1),overlay(T,12)]; (d)[replace(T,6,0,1),overlay(T,12)]; (e/f)[]; returns 1,1,1,1,0,0. Existing fixture views and incoming records unchanged; no other request/draw.
  - Discriminates: normalizing level 0 to 1 fails(a); clamping negative percentage before addition fails(c); saturating rather than wrapping addition fails(d); overlay-before-write reverses every successful log; treating nonpositive current/maximum as writable emits forbidden requests(e/f). Supplied record fields establish no original-profile registration or stored-write claim.
  - runtime_observed: no

- **MCA-V32** Actual index-2 callback combat roles (MCA-R15; boundary: existing state through combat handoff and normal callback continuation)
  - Inputs: owner O is available regular monster class 141/mode 1, aliveExclusionCondition=false, not a hireling; other T is a player. Registered skill 82, level 3, skill count 357; selected captured target state 58, ResultFlags 0, HitFlags 0, HitClass 0, `calc1` reference 1805, Param5=50/Param6=0. Supplied pool count 5891, bytes 1805..1807=[4,4,0], other bytes 0. O has active state 58/count 200 and an existing list with source player/id 9; registry resolves it to distinct source S. Incoming record physical 512, fire 77, flags 8192,result 1,coldLength12, every other named logical field 0. Direct-apply and reaction operations log complete argument roles/current record, return normally and have no fixture effects/draws; their interiors are not executed by this handoff fixture.
  - Seeds: no draw.
  - Draws: none.
  - Trace: guards pass, source S; regular-monster owner selects `calc1`; SCX parameter 4 at L3 gives 50; clear new record, physical=256,result 32,flags 0,hitClass0. Call apply then reaction; T is not monster 290, return 1.
  - Outputs: calculation context(S,82,3); ordered handoff[(apply,game,attacker=T,defender=O,scaleRequest=true,N),(reaction,game,attacker=T,defender=O,N)]. N is the same distinct logical record in both calls: physical 256,result 32,flags 0,hitClass0,every other field in RLI3.2/DPO3 zero. Incoming record retains its exact inputs. Callback result 1; no life/overlay tail request or local unit mutation.
  - Discriminates: swapping event owner/other reverses both combat roles; copying incoming fields leaves forbidden fire/flags/coldLength; treating callback dependency returns as final result or omitting reaction fails the ordered handoff/continuation. This fixture makes no claim about actual DPO/HRC stored effects.
  - runtime_observed: no

- **MCA-V33** Combat calculation selection and zero result (MCA-R15; boundary: callback preparation/handoff)
  - Inputs: independent V32-style valid state/source/availability context with registered skill 87, target state 60 active, source S, incoming physical 100 and all other fields 0. Supplied record ResultFlags 5,HitFlags 4096,HitClass 141, Params1..6=11,0,22,0,33,0. Complete pool[4,0,0,4,2,0,4,4,0]; `calc1`/`calc2`/`calc3` references 0/3/6. Cases owner/other/level: (a) regular monster 141/player/1; (b) player/player/1; (c) player/monster 141/1; (d) recognized hireling 271/player/1; (e) player/monster 141/0; (f) player/monster 141/-1. Every other is present and not monster 290. Apply/reaction fixture operations log, return normally, make no effect/draw. Skill/state counts 357/200 and existing source list as in V32.
  - Seeds: no draw.
  - Draws: none.
  - Trace: references selected 0,3,6,3,6,6; normalized levels 1,1,1,1,0,1; percentages 11,22,33,22,0,33. Clear N and preserve supplied selected flags/class mapping; zero computed damage still makes both calls.
  - Outputs: N.physical 11,22,33,22,0,33; N.result 37,flags 4096,hitClass141 in every case; all other named fields 0. Each ordered log is apply(attacker=other,defender=owner,scaleRequest=true,N) then reaction(same roles,N), and callback result 1. Calculation context uses S/87 and stated normalized level. Incoming record/units unchanged by this fixture; no tail request.
  - Discriminates: using `calc1` for every monster fails hireling(d); choosing by other kind before owner classification fails(a); positive-level rejection suppresses(e); failing negative-level normalization changes(f); skipping zero calculated physical omits(e)'s calls.
  - runtime_observed: no

- **MCA-V34** Class-290 capped-add distribution (MCA-R16; boundary: post-combat existing views with explicit replacing-view fixture)
  - Inputs: post-R15 normal reaction stage; owner O is an available monster 141 with readable control record resolving linked player Q in game registry; Q available. Other is monster 290 with present monster-data MonStats record, Normal Drain50 and existing alignment 0 under a nonextended state 105 list, fMin=false. Current new-record physical 1000; incoming physical 9999 is unused here. Existing O life 990/max 1000,Q life 950/max 1000. For this fixture only, each replacing request updates the supplied logical stat view; this does not verify the actual RLI-Q03 producer. LSE-R13 overlay operations are logged without effects/draws, overlay count 200 and valid context. All resource/state definitions are present with fMin=false. No other dependency effects/draws.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Drain scales current physical to 500; remaining 20%=100. First Q amount 50 yields 1000, returned 50; remaining 50. O amount 50 caps 1040 to 1000, returned 10; remaining 40. Second Q amount 40 caps 1040 to 1000, returned 0, ignored. Overlay O then Q; return 1.
  - Outputs: ordered requests[replace(Q,6,0,1000),replace(O,6,0,1000),replace(Q,6,0,1000),overlay(O,151),overlay(Q,151)]; capped-add returns 50,10,0. Supplied view final O/Q life 1000/1000; callback 1. New record/incoming record unchanged locally. Actual stored-life/overlay effects remain Q04.
  - Discriminates: incoming rather than current physical changes every amount; omitting Drain changes remaining; distributing a fixed half without subtracting actual capped return loses Q's second request; reading all lives before the first write changes the third computed capped return.
  - Independent capped-add subcases use no callback guards: (a) amount 0 then-1 makes no read/write and returns 0; (b) amount 1,current-10,max-5 requests-9 and returns 1; (c) amount 100,current 900,max 1000 requests 1000 and returns 100; (d) amount 1,current 2147483647,max 2147483647 requests-2147483648 and returns 1. Each subcase has supplied existing stat view/fMin=false and a normally returning logged replacing operation, no overlay/draw; no whole callback result is asserted. These distinguish adding an unestablished minimum-1 policy and replacing wrapping arithmetic with saturation; they do not distinguish equivalent equality-cap branches.
  - runtime_observed: no

- **MCA-V35** Reject nonpositive incoming physical before calculation (MCA-R15; boundary: callback guard stage)
  - Inputs: two independent complete V32 contexts with its available monster owner O, player other T, valid selected skill 82/state/source/pool and normally returning fixture dependencies, but incoming physical is 0 and -1 respectively. Incoming fire 77, flags 8192, result 1 and coldLength 12 remain as in V32. An observer logs skill-record access, formula evaluation, combat-record construction, apply, reaction and life/overlay requests.
  - Seeds: no draw; all seeds unchanged.
  - Draws: none.
  - Trace: present owner/other and owner availability pass; signed incoming physical is not positive, so return before looking up the registered skill or constructing a new record.
  - Outputs: callback results 0/0; every observer log empty in both cases; incoming records and units unchanged; no local seed mutation.
  - Discriminates: clamping incoming physical to a positive minimum or testing only negativity reaches forbidden calculation/combat handoffs for physical 0. Testing only zero incorrectly admits -1. Omitting the guard differs from V33(e), where positive incoming damage and a zero calculated result still issue both combat calls.
  - runtime_observed: no

- **MCA-V36** Class-290 skip and owner-only continuation (MCA-R16; boundary: post-combat stage)
  - Inputs: six independent post-R15 reaction stages, with present Normal game G, owner O a monster class 141/mode 1/aliveExclusionCondition=false, other a monster class 290 with readable monster-data/MonStats, state count 200 and a nonextended state-105 list with own layer-0 stat 172=0 and fMin=false. Current new-record physical 1000, O life 500/max 1000, definitions present/fMin=false; owner's monster-data control record absent, so optional link is absent. Overlay count 200, valid LSE-R13 context. Replacing requests update only the supplied logical stat view and return normally; overlay operations log without effects/draws. Case(a) Drain 0; (b) Drain 100; (c) Drain 50 but other alignment stat 172=1; (d) Drain 100 and current physical 0; (e) Drain 100 with owner control record present resolving player Q, Q mode 0/aliveExclusionCondition=false; (f) Drain 100, absent link, but owner mode 0. For (e), registry/control/game/kind/id are valid and Q has readable unit data; no Q resource values are consumed. An observer logs only the named Drain/alignment/link/Unavailable checkpoints below; leading kind/class/data reads are not observer events.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) Drain 0 ends before alignment/link checks. (b) Alignment 0, absent link and available owner pass; Drain 100 preserves physical 1000; remaining=200; owner capped-add requests 700 and returns 200; overlay owner. (c) Nonzero alignment ends before link resolution. (d) The same guards as(b) pass, but remaining 0 ends before resource/overlay operations. (e) Linked Q is unavailable, so end before requesting owner availability. (f) Link absent, then owner is unavailable, so end before computing remaining/resource operations.
  - Outputs: callback result 1 in every case. Ordered effect requests(a,c,d,e,f)[]; (b)[replace(O,6,0,700),overlay(O,151)]. Supplied O life remains 500 except(b), where it becomes 700; no other view changes. Guard logs(a)[Drain=0]; (b/d)[Drain=100,alignment=0,link=absent,Unavailable(O)=false]; (c)[Drain=50,alignment=1]; (e)[Drain=100,alignment=0,link=Q,Unavailable(Q)=true]; (f)[Drain=100,alignment=0,link=absent,Unavailable(O)=true]. No additional named check after a rejection. New/incoming combat records and seeds unchanged locally. Actual stored effects remain Q04.
  - Discriminates: ignoring Drain 0, alignment 1 or the availability rejections emits forbidden checks/requests. Reusing owner's earlier pre-combat availability in(f) emits a replacing request. Issuing overlays for nonpositive remaining fails(d). Requiring a linked unit suppresses(b); treating Drain 100 as 50 requests 600 rather than 700. This case does not claim that multiplying by 100 versus preserving the same value is observably different.
  - runtime_observed: no

- **MCA-V37** Curse policy and first-list selection (MCA-R17; boundary: eligible timed-request prefix before refresh/replacement/creation)
  - Inputs: independent requests after OWD-R06a's target/state eligibility, with source A, player target T, selected skill 66, level 1, duration 200, primary 36/value -100, state 9 and default callback. State count 185; state 9 curse membership true, state 68 false. T has an extended base with ordered primary lists [L68(state 68,curse=false),L1(state 19,curse=true),L2(state 9,curse=true)] and secondary list L3(state 9,curse=true). Its resolved stat 109 is supplied through a present definition/count 512 with fMin=false. Cases(a) resistance 50; (b) -50; (c) 0; (d) 100 and active state 57; (e) 50 and active state 57; (f) R11-style request with skill 177/level 4/primary 25/value 25/state 68, resistance 100 and active 57; (g) resistance 50 and duration -200; (h) absent target base/no chains, hence RLI supplies stat 109=0 and no active state 57. State 57 is otherwise inactive; state storage is readable. Each request is independent. An observer logs resistance reads, mulDiv calls, state-57 tests and list selection; downstream lifecycle operations are not executed by this prefix fixture.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a/b/c) curse branch, resistance below 100; nonzero values adjust duration, then state 57 false and first primary curse L1 selected. (d) reject before duration/state test; (e) adjust then reject at state 57. (f) noncurse skips resistance/state 57 and state lookup selects L68. (g) retain signed duration arithmetic then select L1. (h) absent base gives zero resistance/false membership and no existing curse list.
  - Outputs: local duration values 100,300,200,200,100,200,-100,200 respectively. Cases(d/e) return null from applier, with no list search or downstream request. Other cases end at the continuation boundary with selected lists L1,L1,L1,L68,L1,absent for (a,b,c,f,g,h); no final applier/recipient result is asserted for them. Resistance reads occur in all except(f). Ordered mulDiv logs(a/b/e/g) are (50,200,100),(-50,200,100),(50,200,100),(50,-200,100); none otherwise. State-57 tests occur in (a,b,c,e,g,h), not(d/f). Existing lists, stats, states, seeds and the parent prepared request are unchanged; only each stated local duration changes.
  - Discriminates: clamping negative resistance or duration fails(b/g); rejecting at state 57 before adjustment leaves(e)'s duration 200; using state lookup rather than first primary curse chooses L2 instead of L1. Searching secondary first chooses L3. Reading resistance or state 57 for a noncurse request adds forbidden operations in(f). Treating a rejected request as an absent-list creation frontier conflates(d/e) with(h).
  - runtime_observed: no

- **MCA-V38** Complete two-recipient resisted curse rejection (MCA-R01, MCA-R04 through MCA-R09, MCA-R13, MCA-R17; boundary: supplied readable world/state/stat records through actual normal action return)
  - Inputs: action 112/invoking skill 212/level 1; actor A monster class 312, latch=false, dynamic position(95,95), stored target words(100,100), absent target cache, present game and actor owner-control link absent. Loaded Skills count 357; selected record 66 has R06's captured fields. Supply pool count 5891: complete duration bytes at reference 1287 are [4,2,0], radius bytes at 1290 are [4,0,0], every other byte 0. State count 185, selected state 9 record present/curse membership true. R0 rectangle(90,90,40,40) contains actor and center, ordered room array[R0], unit chain A->T->U->absent. T and U are distinct players, mode 1/aliveExclusionCondition=false, dynamic coordinates(101,100)/(100,102), target flags bits 1/2/3 true. A/T/U each have an extended base with an existing nonextended state-105 list own layer-0 stat 172=1 and no owner redirect; definitions/count 512/fMin=false make those tier reads 1. T/U have active state 57 in readable membership storage and existing resolved stat 109=50/25 respectively, fMin=false. No other dependency field/function is requested; observer logs named application requests before/after their actual prefix, relationship, lifecycle and event operations. No successful application result or stored write is supplied.
  - Seeds: A={lo:1,hi:0}; all other seeds unchanged.
  - Draws: exactly (A,modulus 5,result 0), no further draw.
  - Trace: select 66, set latch, calculate radius 3 then duration 200, prepare primary 36=-100/state 9. Center resolves(100,100), R0 admitted. A distance 50 skips; T/U distances 1/4 pass inclusive squared radius 9 and R08. R09 preserves -100 for players. Each eligibility subrecipe checks the three true flags, available candidate and one positive relationship; distinct non-player/player pair with tiers 1/1 is not sameTierNeutral, so PVP-R05a returns true. Each fresh local request starts duration 200. Actual curse policy reads 50/25, changes that local duration to 100/150, then active state 57 rejects with null. Each actual recipient returns 0; outer traversal ignores both and normally returns 1.
  - Outputs: A.seed={lo:1791398085,hi:0}, latch=true, selected intermediate 66, final action result 1. Ordered application log[(source=A,target=T,skill=66,level=1,state=9,primary=36,value=-100,durationBefore=200,durationAfter=100,result=null),(source=A,target=U,skill=66,level=1,state=9,primary=36,value=-100,durationBefore=200,durationAfter=150,result=null)]; callback returns [0,0]. Prepared request duration remains 200 after both calls. Relationship log[(A,T,game)=true,(A,U,game)=true], one per candidate; no aura-only negative first test. First-list search, refresh/replacement/creation, additional-stat, event-cancel/register and AI-change logs empty. No state/list/stat mutation; only stated A seed/latch writes remain.
  - Discriminates: leaking the first local duration into the shared prepared request makes U start at 100 and finish at 75; using actor resistance changes both durations; moving state-57 rejection before duration adjustment leaves both at 200. Stopping at the first callback 0 omits U and changes the required final caller result/visit log. This is static execution over supplied valid records, not a claim that this player state was naturally produced.
  - runtime_observed: no

- **MCA-V39** Actual existing curse refresh completes successfully (MCA-R01, MCA-R04 through MCA-R09, MCA-R13, MCA-R17, MCA-R18; boundary: supplied readable world/state/stat records through stored expiry and normal action result)
  - Inputs: two independent V38 contexts, replacing its chain by A->T->absent and omitting U. T's dynamic coordinates are(100,100), mode 1 and all three required target flags true. Game frame 1000, valid empty pending frame queues. T has no active state 57 and active state 9. Its extended base primary chain is[N,L,L105], with no secondary chain. N is an ordinary noncurse list, state 9/skill 66, expiry 1700/timed=true, source owner monster/id 8, removal callback J, own layer-0 stat 36=-31 and no other own entries; its parent/association are the same base/T. L is an ordinary curse list, state 9/skill 66, expiry 1500/timed=true, source owner monster/id 7, removal callback K, own layer-0 stat 36=-73 and no other own entries; parent and association are T's base/T. L105 is the nonextended alignment list from V38 with own stat 172=1. Existing state-9 contribution and all effective stat views are readable and stay unchanged; T's resolved stat 109 is 50 in(a),0 in(b). A's alignment/owner/target/room/seed and the captured selected-66 record/expression pool are exactly V38's. Case(a) caller, N and L level 1; case(b) caller, N and L level 0. The curse conditions on N/L are independent existing list metadata under RLI section 3 and STL-R06a; this fixture does not assert a natural producer of the two same-state lists. The actual scheduler consumes STL-R01's normal nonnegative-frame domain; no successful applier return or stat-writer effect is supplied. Observers record expiry writes, scheduling, applier/recipient results and all later creation, stat, AI, auxiliary and item-event operations.
  - Seeds: A={lo:1,hi:0} independently; other seeds unchanged.
  - Draws: exactly (A,modulus 5,result 0), no further draw.
  - Trace: both select 66 and set A's latch. Case(a) evaluates radius 3/duration 200; T has distance 0, passes the actual player eligibility and positive relationship, and its local duration becomes 100. Case(b) evaluates radius/duration 0; the same T still passes inclusive distance 0, and resistance 0 leaves local duration 0. R17 skips noncurse N and selects L in both. A separate RLI-R03 state-9 lookup would select N; it is not performed after this curse selection. State/skill/level all equal, so(a) stores expiry 1100 and schedules type 12 there;(b) skips both operations. Actual applier returns L; fixed index-0 continuation has no additional stat/event work and returns 1. Outer traversal completes with 1.
  - Outputs: in both A.seed={lo:1791398085,hi:0},latch=true; selected identity 66 is intermediate; applier result=L,recipient result=1,final action result=1. Case(a) ordered log[expiryWrite(L,1100),schedule(T,type 12,1100,no override,0,0)], with final L.expiry=1100/timed=true and one pending event at frame 1100 in slot 12 under STL-R01. Case(b) expiry/schedule log empty, L.expiry=1500/timed=true and pending queues empty. In both N remains expiry 1700/timed=true, own 36=-31, source owner monster/id 8, callback J, state 9/skill 66 and the stated level; no write or lifecycle request targets N. The primary chain remains[N,L,L105]. L's own 36=-73, source owner monster/id 7, callback K, state 9/skill 66, stated level, parent/association/chain and active state-9 membership remain unchanged. No allocation, detach/free/attach, primary/additional stat write, state marking, AI operation, auxiliary operation, item-event cancellation or registration; no other local mutation/draw. Prepared duration remains 200/0 respectively. Original installation, later event firing/free and gameplay execution are unasserted.
  - Discriminates: repeating state-9 lookup after R17 selects earlier noncurse N instead of L, so(a) would write N.expiry=1100 and leave L.expiry=1500, violating both the exact expiry log and unchanged-N assertion; in(b) it would return N instead of the asserted L. Rewriting the requested -100 magnitude changes L's retained -73; replacing source/callback changes the retained owner/K; extending only rather than replacing expiry leaves 1500 instead of 1100 in(a). Treating local resistance adjustment as shared prepared-duration mutation changes prepared 200. An allocation or replacement on refresh adds forbidden operations. Scheduling even for duration 0 adds an event in(b); requiring positive level/radius rejects its actual successful point candidate. Returning the refreshed pointer as the action integer instead of the actual recipient/traversal result fails final 1.
  - runtime_observed: no

- **MCA-V40** Class-290 distribution with actual own/resolved writes (MCA-R16, MCA-R19; boundary: post-combat continuation through actual prepared life storage and later overlay requests)
  - Inputs: V34's post-R15 normal reaction context, Normal game, available owner O monster 141, other monster 290, valid MonStats Drain 50, alignment 0 from the specified nonextended state-105 list, valid owner control resolving linked available player Q, and current new-record physical 1000. Replace V34's view-update fixture with actual MRH-R08 storage: O and Q each have an associated extended root base, no parent, secondary=false, owner type 1/0 respectively. Their own/resolved arrays are independently sorted [(6,0,life),(7,0,1000)] with active counts 2/capacities 8, unchanged backing storage and empty changed-key collections. Initially O own life=900/resolved=990; Q own=900/resolved=950. Definition count 359, stat 6 has MRH-R08's captured direct controls, fMin=false/MinAccr 0/ValShift 8; stat 7 is present/fMin=false. No insertion, zero deletion or resize is required. Other's alignment definition/read context remains V34's. The actual replacing unit writer runs, with no supplied result or view mutation. Overlay storage remains outside this vector: after the life writes, record LSE-R13's operations on O then Q and require only normal continuation, without supplying an overlay result or inferring its stored effects.
  - Seeds: no direct draw in this continuation/read/write chain; any overlay-owned RNG is outside the assertions.
  - Draws: none in the closed life-storage chain.
  - Trace: Drain 50 turns physical basis 1000 into 500, then remaining=100. First linked add asks 50, reads Q resolved life 950/max 1000 and requests 1000; actual delta=1000-900=100, so Q own/resolved become 1000/1050, capped-add return 50, remaining 50. Owner add reads 990/max 1000, caps 1040 to 1000, returns 10; actual delta 100 makes O own/resolved 1000/1090, remaining 40. Final linked add reads the newly stored resolved 1050/max 1000, caps 1090 to 1000 and computes return-50. Its requested 1000 equals Q's own 1000, so the actual writer leaves both arrays unchanged. Ignore this last return and request overlay 151 on O then Q; callback normally returns 1.
  - Outputs: ordered life requests[replace(Q,6,0,1000),replace(O,6,0,1000),replace(Q,6,0,1000)]; capped-add returns[50,10,-50]. Ordered current/max reads[(Q,950,1000),(O,990,1000),(Q,1050,1000)]. Immediately after those actual writes, O own/resolved life=1000/1090 and Q=1000/1050; each array count remains 2/capacity 8 and backing store unchanged. All stat-7 values, changed-key collections, list relations, membership and combat records remain unchanged by this life chain. No stat callback, allocation, resize or direct draw. Later ordered overlay requests[(O,151),(Q,151)] and normal callback result 1 remain R16's continuation; subsequent overlay-owned mutations/RNG are unasserted. Earlier combat production, state installation, outer action result and gameplay confirmation are unasserted.
  - Discriminates: assigning the request directly to resolved life gives both 1000 and changes the final linked return to 0. Reading own life instead of resolved initially changes the first request to 950. Treating the local return as the resolved delta changes remaining after the first add. Treating negative last return as failure changes callback 1. The vector executes the named storage producer and does not reuse V34's supplied view updates.
  - runtime_observed: no

- **MCA-V41** Actual zero-calculation callback changes own/resolved life (MCA-R14, MCA-R19; boundary: existing selected-87 callback through actual prepared storage and result)
  - Inputs: V30's complete existing state/source/callback context, incoming physical 512, player owner O mode 0, available monster other T class 141, selected skill 87/level-2/count 357, state 60 active on O, existing source kind monster/id 7 with no registry match, empty calculation reference 4294967295, prgoverlay-1/count 200. Replace its logged-only life writer with actual MRH-R08 storage on T: associated extended root base, no parent, secondary=false, owner type 1; own sorted entries[(6,0,512),(7,0,768)], resolved[(6,0,1024),(7,0,768)], both counts 2/capacities 8 and empty changed-key collection. Definition count 359, stat 6 has MRH-R08's captured direct controls and fMin=false/MinAccr 0/ValShift 8; stat 7 present/fMin=false. No insertion, zero deletion or resize; no supplied writer result or storage effect.
  - Seeds: no draw; existing unit seeds unchanged.
  - Draws: none.
  - Trace: the actual empty-expression front end returns 0 after negative level normalizes to 1. Positive current 1024/max 768 pass; amount 0 leaves candidate 1024, which caps to 768. The actual replacing writer targets T's base own life: delta=768-512=256, own becomes 768, resolved becomes 1024+256=1280. Negative overlay selector is omitted; callback returns 1.
  - Outputs: replacing log[(T,6,0,768)], callback result 1; T own/resolved life 768/1280, counts 2/capacities 8/backing stores unchanged, all stat 7 values 768. Changed-key collection, other entries, list relations, states, events and O unchanged. No stat callback or overlay operation, allocation, resize, extra request or draw; incoming record unchanged. Original event registration/state installation and gameplay execution are unasserted.
  - Discriminates: suppressing writes for zero percentage leaves own 512/resolved 1024. Assigning the capped request directly to resolved life gives 768 rather than 1280; applying another maximum clamp inside the writer also fails 1280. Targeting owner O instead of other T fails the stated target and unchanged-O assertions. This establishes stored effects only in the exact named prepared direct domain.
  - runtime_observed: no

- **MCA-V42** Life callback rejects before calculation (MCA-R14; boundary: callback guards through normal return)
  - Inputs: twelve independent complete V30 contexts, with exactly one replacement per case: (a) owner absent; (b) other absent; (c) owner is an object with readable unit storage; (d) other is an object with readable unit storage; (e) other is the original monster but mode 0; (f) incoming record absent; (g) incoming physical 0; (h) incoming physical -1; (i) registered skill 87 has no loaded record; (j) selected target state -1; (k) selected target state 200, equal to state count; (l) owner has no active state 60. One further independent V30 context (m) has active state 60 in readable membership storage on an extended base, but no matching list in its readable chains. All other V30 inputs remain supplied, including present game, counts, readable valid registry and existing resource values; no natural producer of a supplied membership/list combination is asserted. Observers log source registry lookup, calculation front end, resource reads, replacing writes and overlays.
  - Seeds: no draw; all supplied unit seeds unchanged.
  - Draws: none.
  - Trace: each case fails its stated R14 guard and returns 0 before source resolution or calculation. Cases(g/h) reject the incoming physical value rather than evaluating an empty calculation; case(m) fails the required existing-list lookup.
  - Outputs: callback result 0 in all thirteen cases; every observer log empty. All supplied units, records, arrays, lists, states and events unchanged; no extra request or local mutation.
  - Discriminates: admitting any failed guard reaches at least V30's calculation-front-end request and fails the empty log. This differs from V31(e/f), which calculate and read resources before returning 0, and from V30, whose zero calculated amount still requests a write and returns 1. No actual state installation, stored writer or gameplay outcome is asserted.
  - runtime_observed: no

- **MCA-V43** Actual recipients reject eligibility before application (MCA-R01, MCA-R04 through MCA-R09, MCA-R13, MCA-R17; boundary: supplied readable world through actual normal action return)
  - Inputs: V38's complete two-player world/action/selected-record/expression/state/stat context, except T's target flag bit 1 is false and U's target flag bit 2 is false. Their other target flag bits remain true. Both players retain V38's readable active state 57 and resistance values, but those fields are not requested after these eligibility failures. Observers log actual recipient results, timed-application requests, relationship requests, resistance reads and lifecycle/event operations. No callback or applier result is supplied.
  - Seeds: A={lo:1,hi:0}; all other seeds unchanged.
  - Draws: exactly (A,modulus 5,result 0), no further draw.
  - Trace: select 66, set latch and calculate radius 3 then duration 200. The same room and distances as V38 admit T/U through R08; R09 preserves the fixed primary -100 for each player. T/U fail the required target flags before the positive relationship test or application request. Both actual recipients return 0; R13 ignores those results and finishes normally.
  - Outputs: A.seed={lo:1791398085,hi:0}, latch=true, selected intermediate identity 66, ordered recipient results[(T,0),(U,0)], final action result 1. Timed-application, relationship, resistance-read, lifecycle and event logs empty. Prepared duration remains 200; no candidate-local application request is built. All supplied candidate state/stat/list values remain unchanged; no further mutation or draw.
  - Discriminates: omitting either required flag check reaches a forbidden relationship request. Deferring eligibility until the applier reaches a forbidden application request, unlike V38's later policy rejection. Propagating either callback 0 as the outer result fails the final 1.
  - runtime_observed: no

- **MCA-V44** Aura preparation reaches allocation without supplying its result (MCA-R20; boundary: actual action entry through coordinate rejection or first collection-context allocation request)
  - Inputs: four independent action-111 contexts with present game G and readable collection pool, invoking skill 177/count 357 and present record, actor A a monster class 312 with present path, absent cached target, no owner redirect and stored target words X=100/Y=100; latch=false and actor current room R0. Cases(a) level 1/Param4=3; (b) level 0/Param4=-7; (c) level 2147483647/Param4=3; (d) as(a) but stored target X=0. Registry storage is readable but unused by the absent-cache refresh. Observers log coordinate outputs, latch changes, collection-setup inputs, actor-room lookup and initial context-allocation requests. No allocation operation is executed or its result supplied.
  - Seeds: no draw; all unit seeds unchanged.
  - Draws: none.
  - Trace: MTP-R09 writes X then Y. Cases(a-c) return coordinate success and set latch. Compute wrapped extents 3+2=5, -7+0=-7, and 3+(-2)=1; enter setup and request its initial context storage. Stop at that request before room lookup. Case(d) writes coordinates(0,100), returns coordinate result 0 and action result 0 before latch/setup.
  - Outputs: coordinate-output pairs(100,100),(100,100),(100,100),(0,100). Cases(a-c) latch=true, setup inputs(G,A,100,100,extent=5/-7/1) respectively and exactly one initial collection-context storage request through G's pool. Actor-room lookup log empty in all cases; no initialized collection or candidate result is asserted. Case(d) latch=false, setup/allocation logs empty and final action result 0. All other supplied state unchanged; no draw, skill expression, state/stat/event operation or release. Cases(a-c) assert no normal action result beyond the gated allocation request.
  - Discriminates: computing radius from a skill expression adds an unrequested operation; positive-level/extent rejection drops(b)'s request; saturation changes(c)'s extent; performing room lookup before initial allocation violates its empty log. Setting latch before coordinate validation changes(d).
  - runtime_observed: no

- **MCA-V45** Actual aura collection order and fixed filters (MCA-R20; boundary: supplied initialized collection and readable world through actual candidate-array writes and collector result)
  - Inputs: four independent prepared collectors with capacity 15, all output slots initially absent, accepted counter/output count 0, no custom predicate and separate output/world storage. Cases(a/b): center(100,100), extent 5, current R0 rectangle(90,90,40,40), level 76, stored room array[R1,R0], count 2; R1 rectangle(90,90,40,40), level 76. R0 head chain A->T->P->U->D->N->absent. A is the actor, monster 312/mode 1 at(100,100); T player/mode 1 at(103,104); P player/mode 1 at(100,100) with target flag bit 2 false; U monster 141/mode 1 at(99,100) whose current room is readable town RT with level 75; D monster 141/mode 12 at(100,100); N object with prepared static coordinates(100,100). All except P have target flag bits 2/3 true; P bit 3 true. Other dynamic paths are present; A/T/P/D's current room is R0. R1 has a separate monster Q mode 1 at(100,100), flags 2/3 true and current room R1. Case(b) changes only A's kind/class to player/class 4, mode 1; paths/flags stay the same. Case(c) has monster actor A at(95,95), center(95,95), extent 5; current R0 and R1 have the same rectangles/level 76 as(a), stored room array[R1,R0,R1], count 3. R0 chain A->M->absent, R1 head Q->absent; M/Q are monsters 141/mode 1 at(96,95)/(95,96), present paths, flags 2/3 true and current rooms R0/R1 respectively. Case(d) has absent current room and otherwise(a)'s supplied context; its unvisited world storage remains readable. Every named unit is distinct except repeated room-array entries; all storage/links remain fixed. Observers log chosen room sequence, ordered next-link and X/Y reads, predicate decisions, candidate writes, allocation and relationship/state/stat/event operations. No predicate result or writer effect is supplied.
  - Seeds: no draw; all seeds unchanged.
  - Draws: none.
  - Trace: (a/b) bounds 95/95/105/105 lie strictly inside R0, so only R0 is visited. Geometry distances are 0,25,0,1,0,0. A passes only as a monster; T passes inclusive distance 25; P fails flag 2; U fails its candidate-room level 75; D fails mode 12; N fails kind after its coordinates are read. (c) x-r=90 equals R0's left edge, so strict containment fails. Visit stored R1,R0,R1; all prefilters and distances pass, and monster actor A is not identity-rejected. (d) absent room returns immediately.
  - Outputs: selected room sequences(a/b)[R0],(c)[R1,R0,R1],(d)[]. Actual written candidate sequences(a)[A,T],(b)[T],(c)[Q,A,M,Q],(d)[]; accepted counters, output counts and collector results respectively 2,1,4,0. Capacities remain 15, backing storage unchanged and all unused slots remain absent. Coordinate log(a/b)[X(A)=100,Y(A)=100,X(T)=103,Y(T)=104,X(P)=100,Y(P)=100,X(U)=99,Y(U)=100,X(D)=100,Y(D)=100,X(N)=100,Y(N)=100]; (c)[X(Q)=95,Y(Q)=96,X(A)=95,Y(A)=95,X(M)=96,Y(M)=95,X(Q)=95,Y(Q)=96]; (d)[]. Each visited unit logs its next link before those coordinate reads. No R1 head/coordinate read in(a/b). No allocation/growth, relationship, state/stat/event or release operation; world records, unit state/stats and seeds unchanged. No downstream aura handler or final action result is asserted.
  - Discriminates: always using the stored array adds Q in(a/b); using non-strict containment drops R1 visits and duplicate Q in(c). Universal self-exclusion removes A in(a/c), while no player self-exclusion adds A in(b). Omitting candidate town or flag checks adds U/P. Swapping X/Y preserves squared distances but violates the exact read log. Deduplicating the collection drops(c)'s second Q. Geometry after kind rejection omits N's asserted coordinate reads.
  - runtime_observed: no

- **MCA-V46** Fifteenth accepted aura candidate reaches growth after its write (MCA-R20; boundary: prepared collector through normal no-growth return or first growth request)
  - Inputs: two independent prepared collectors with actor A monster 312, center(100,100), extent 5, current R0 rectangle(90,90,40,40)/level 76, capacity 15, empty candidate slots, accepted counter/output count 0 and separate readable world/output storage. R0 has an ordered chain of distinct monsters M1 through M14 in(a), and M1 through M15 in(b), ending absent. A is not in either chain. Every Mi is monster 141/mode 1 at(100,100), with present path, flags 2/3 true and current room R0. All inputs to fixed predicates are readable; no outside mutation, result or writer effects are supplied. Observe candidate writes, accepted counter, capacity, final count store/return and allocation requests. Do not execute a requested growth operation or supply its result.
  - Seeds: no draw; all seeds unchanged.
  - Draws: none.
  - Trace: strict containment chooses only R0; all candidates pass. (a) writes M1..M14 in order, counter/local count 14 remains below capacity 15, then stores output count 14 and returns 14. (b) first writes all M1..M15, counter/local count 15 equals capacity 15, then requests growth to capacity 30; stop before the allocation operation returns.
  - Outputs: (a) actual slots 1..14 contain M1..M14, slot 15 remains absent, counter/output count/result 14, capacity 15 and no allocation request. (b) actual slots 1..15 contain M1..M15, counter 15, capacity still 15 at the growth-request frontier, and one request for capacity 30 after M15's write. Output count remains its initial 0 because the final count store has not been reached; no collector return or grown storage is asserted in(b). Backing storage and supplied world/units unchanged in both; no relationship/state/stat/event or release operation, direct draw or other request.
  - Discriminates: growing before storing the fifteenth candidate omits(b)'s M15 slot write; growing only when a sixteenth candidate appears omits(b)'s request. Treating a growth request as successful allocation invents a capacity 30 store/normal collector result beyond the explicit gate. Growing on the fourteenth acceptance adds a forbidden request in(a).
  - runtime_observed: no

- **MCA-V47** Collector null-path, fixed guards and signed wrapping boundaries (MCA-R20; boundary: supplied initialized collection through actual no-growth writes/count/return)
  - Inputs: three independent prepared collectors with actor A a monster class 312, capacity 15, empty output slots, accepted counter/output count 0, no custom predicate and separate readable world/output storage. A is absent from each candidate chain; its present path supplies current R0. Every R0 has level 76, readable rectangle and empty stored room array/count 0. Case(a) center(65535,65535), extent 5, R0 rectangle(65525,65525,40,40), chain N->F->P0->P17->M0->absent. N is monster 141/mode 1 with absent dynamic path and target flags 2/3 true; its absent path also makes current room absent. F is a player/mode 1 at(65535,65535), flag 2 true/flag 3 false. P0/P17 are players in modes 0/17 respectively, and M0 is monster 141/mode 0, all at(65535,65535) with flags 2/3 true. F/P0/P17/M0 have present readable paths and current room R0. Case(b) center(100,100), extent -5, R0 rectangle(90,90,40,40), sole candidate M monster 141/mode 1 at(103,104), present path/current room R0, flags 2/3 true. Case(c) center(-2147483648,-2147483648), extent 5, R0 rectangle(-2147483648,-2147483648,40,40), sole candidate T monster 141/mode 1 at(65535,65535), present path/current room R0, flags 2/3 true. All storage is readable and fixed; these supplied coordinates/links do not assert a natural world producer or successful allocation. Observers log saved next links, semantic X/Y coordinate results, wrapped arithmetic, fixed guards, candidate writes, counter/count and allocation/effect operations; zero coordinate results for N do not mean reads through an absent path.
  - Seeds: no draw; all unit seeds unchanged.
  - Draws: none.
  - Trace: (a) strict containment chooses R0 and its prefilter admits it. N's coordinate results are X=0 then Y=0; dx=dy=-65535, each wrapped square=-131071 and wrapped distance=-262142, which passes signed comparison to 25. N passes monster mode and flags; absent current room does not reject, so append N. F has distance 0 but fails flag 3 after flag 2; P0/P17/M0 have distance 0 and fail their kind-specific modes before flag/room operations. (b) strict containment and prefilter still admit R0 with extent -5; squared extent 25 and M's distance 25 pass inclusively, then all fixed guards pass. (c) wrapped x-r/y-r=2147483643 are greater than rectangle origins; x+r/y+r=-2147483643 are less than right/bottom -2147483608, so strict containment chooses R0 despite the empty stored array. Prefilter admits it. Each candidate-minus-center delta wraps to -2147418113, each square wraps to -131071 and distance to -262142; signed comparison to 25 passes, followed by all fixed guards.
  - Outputs: chosen room sequence[R0] in every case. Actual candidate arrays(a)[N],(b)[M],(c)[T], accepted counters/output counts/collector results 1/1/1. Each backing store/capacity 15 and unused slot remains unchanged. Coordinate-result logs(a)[X(N)=0,Y(N)=0,X(F)=65535,Y(F)=65535,X(P0)=65535,Y(P0)=65535,X(P17)=65535,Y(P17)=65535,X(M0)=65535,Y(M0)=65535],(b)[X(M)=103,Y(M)=104],(c)[X(T)=65535,Y(T)=65535]. Each next link is saved before that candidate's coordinate results. Candidate-room checks occur only for N/M/T and give absent/R0/R0 respectively; N has no room-level lookup. No allocation/growth, relationship/state/stat/event or release operation, mutation of world/unit records or draw. No original storage producer, downstream aura effect or final action result is asserted.
  - Discriminates: rejecting absent paths or candidate rooms omits N; using unsigned/unbounded squared distance rejects N/T. Omitting flag 3 after testing flag 2 appends F. Omitting player mode 0/17 or monster mode 0 appends its named candidate. Rejecting or zero-clamping negative extent omits M. Nonwrapping containment arithmetic in(c) falls back to the empty stored room array and omits T. Applying the fatal-null path getter directly to N fails the collector's actual local zero-coordinate continuation.
  - runtime_observed: no

- **MCA-V48** Independent noncurse aura request refreshes an existing list (MCA-R11, MCA-R17, MCA-R18; boundary: independent request-producing helper through actual existing-list applier and local optional-list return)
  - Inputs: three independent R11 request-stage contexts; do not execute R10, the collector or the full action. Source A is a present monster, target T a distinct player/class 4/mode 1 with game G/frame 1000. Invoking skill 177/count 357 is present, with captured Params1/2/3=10/50/5; state 68/count 185 is present and curse membership false. Request primary stat is 25, state 68, default removal callback. T has a readable extended base associated with T, primary chain[C,L] and no secondary lists; both existing lists have that base as parent. C has state 9/skill 66/level 1, curse condition true, expiry 1700/timed=true, source owner player/id 8, callback J and own layer-0 stat 36=-31. L has state 68/skill 177, curse condition false, expiry 1500/timed=true, source owner monster/id 7, callback K and own layer-0 stat 25=-73. T has active states 9/57/68, readable state membership storage, base own stat 109=100 and resolved entries 25=-73,36=-31,109=100, all at layer 0; definitions/count 359/fMin=false are present. All other own entries absent; readable capacities/backing storage remain fixed. Case(a) caller level 4/L.level 4; (b) caller level 0/L.level 0; (c) caller level 3/L.level 4. G has valid initially empty pending queues in STL-R01's normal nonnegative-frame domain. Observers log the prepared request, resistance/state-57 queries, selected existing list, expiry writes, scheduling, applier/helper returns, and creation/stat/callback/event operations. No applier result, allocation or stored effect is supplied; no natural producer of these existing records or R10 eligibility is asserted.
  - Seeds: no draw; all unit seeds unchanged.
  - Draws: none.
  - Trace: R11 constructs durations 200/0/150 and requested stat-25 values 25/5/20 from levels 4/0/3. OWD-R06a's target/state checks pass, with no monster-target checks for T. Noncurse state 68 skips resistance/state 57 and selects L by RLI-R03's state lookup, not earlier curse C. Cases(a/b) have equal state/skill/level: (a) stores expiry 1200 and schedules type 12 there; (b) duration 0 skips both and returns L. Case(c)'s requested level 3 is lower than stored 4, so OWD-R07a returns null without refresh or replacement. The local R11 helper forwards each actual applier result unchanged.
  - Outputs: prepared requests(source=A,target=T,skill=177,level=4/0/3,duration=200/0/150,stat=25,value=25/5/20,state=68,callback=default). Resistance and state-57 query logs empty; selected existing list L in all cases. Actual applier results[L,L,absent] and local helper results[L,L,absent]. Case(a) ordered effect log[expiryWrite(L,1200),schedule(T,type 12,1200,no override,0,0)]; L.expiry 1200/timed=true and one pending type 12 event at 1200 under STL-R01. Cases(b/c) effect logs empty, L.expiry 1500/timed=true and pending queues empty. In every case L retains own stat 25=-73, source owner monster/id 7, callback K, state 68/skill 177 and stated existing level; C remains completely unchanged. Base own/resolved entries, list chains/relations, active membership, actor and seeds remain unchanged. No allocation, detach/attach/free, magnitude/stat write, removal-callback invocation, active-state write, AI operation, item-event cancellation/registration or draw. No R18 fixed-index-0/1 recipient result or final action result is asserted; later timer firing/removal, original installation, natural aura recipients and complete aura effects remain gated.
  - Discriminates: using first-curse selection chooses C rather than L. Reading resistance 100 or active 57 as a curse guard wrongly rejects(a/b); rewriting magnitude, owner or callback violates their retention. Extending-only expiry leaves 1500 instead of 1200 in(a). Requiring positive level or scheduling for zero duration fails(b). Refreshing or replacing on lower level mutates(c), whose actual applier/helper results are absent. Converting optional-list results to boolean success instead of forwarding the same identity fails the exact helper outputs. Treating this independent request as proof that a candidate passes R10 or that the full action returns a particular result exceeds the asserted boundary.
  - runtime_observed: no

- **MCA-V49** Continue actual traversal after a successful recipient (MCA-R01, MCA-R04 through MCA-R09, MCA-R13, MCA-R17, MCA-R18; boundary: supplied readable world through both actual existing-list refreshes and final action result)
  - Inputs: V38's complete action-112 context, actor A, two-player world and ordered chain A->T->U->absent, selected-66 record and complete expression pool, stored target coordinates, relationship/alignment inputs, flags, resource definitions and T/U resistance 50/25. Override active state 57 to false on both T and U; active state 9 is true on both. Game G.frame=1000, with valid initially empty pending queues under STL-R01. T/U each belong to G and have a readable extended base and independent existing nonextended state 105 alignment list with own stat 172=1 as in V38. T primary chain[L,T105], U primary chain[D,U105], with all parents/associations to their respective base/unit and no secondary lists. L/D both have state 9/skill 66/level 1, curse condition true and timed=true. L has expiry 1500, own layer-0 stat 36=-73, source owner monster/id 7 and callback K; D has expiry 1600, own layer-0 stat 36=-55, source owner monster/id 9 and callback H. T/U's existing base own/resolved stat 109 values are 50/25; all existing definition/read conditions are V38's. Other own entries are unchanged; all storage is readable and independent. No callback/applier result, allocation or stored write is supplied. Observers log application/recipient results, expiry writes, scheduling and next-link reads. Original list installation, later event execution and natural production of this world remain unasserted.
  - Seeds: A={lo:1,hi:0}; other unit seeds unchanged.
  - Draws: exactly (A,modulus 5,result 0), no further draw.
  - Trace: select 66, set latch, compute radius 3 then duration 200, prepare state 9/primary 36=-100. A's distance 50 skips; T/U distances 1/4 pass radius squared 9 and R08. Actual player eligibility and positive relationships pass as in V38. Each fresh local request starts duration 200; resistance 50/25 produces durations 100/150, inactive state 57 permits continuation, and the primary curse search selects L/D. Equal metadata refreshes actual expiries to 1100/1150 and schedules each target there. Both actual recipients return 1; after T returns 1, R13 reads its next link U and processes U. Normal traversal finishes 1 and R05 forwards 1.
  - Outputs: A.seed={lo:1791398085,hi:0}, latch=true, selected intermediate 66, ordered applier results[L,D], recipient results[(T,1),(U,1)], final action result 1. Ordered effects[expiryWrite(L,1100),schedule(T,type 12,1100,no override,0,0),expiryWrite(D,1150),schedule(U,type 12,1150,no override,0,0)]. L/D retain timed=true and each target has exactly its one pending type 12 event at 1100/1150. Traversal log[next(A)=T,recipient(T)=1,next(T)=U,recipient(U)=1,next(U)=absent]. L/D retain stated own magnitudes, owner identities, callbacks, state/skill/level and all chain relations; membership and other stats/records unchanged. Prepared duration stays 200. No allocation/detach/attach/free, magnitude/additional-stat write, callback invocation, AI/auxiliary operation or item-event cancellation/registration; no other mutation/draw. Original creation, later removal, broader curse effects and gameplay observation remain gated.
  - Discriminates: returning or stopping after T's nonzero recipient result omits U's actual expiry/schedule and final traversal log. V38 already distinguishes the zero-result case; this vector independently distinguishes a stop-only-on-success mistake without supplying successful callback returns. Treating T's local duration as the prepared duration changes U's requested duration and expiry. Reapplying requested primary -100 overwrites the retained old magnitudes.
  - runtime_observed: no

- **MCA-V50** Actual existing refresh replaces skill-event bindings (MCA-R01, MCA-R04 through MCA-R08, MCA-R13, MCA-R17, MCA-R18, MCA-R21; boundary: supplied world/state/event records through actual refresh, logical event collection and final action return)
  - Inputs: three independent action-112 contexts with invoking skill 212, caller level 1, present game G/frame 1000, actor A monster 312/mode 1 at(95,95), latch=false, present path/current room R0, absent cached target and stored target words(100,100). R0 rectangle(90,90,40,40), stored room array[R0]/count 1, chain A->T->absent. T is a distinct player/class 4/mode 1 at(101,100), flags 1/2/3 true and aliveExclusionCondition=false. A/T each have an extended base and nonextended state 105 list with own layer-0 stat 172=1, no owner redirect, valid definitions/count 359/fMin=false, and present readable registry/game context. Loaded skill count 357/state count 185. Case(a) selects record 87, state 60, captured parameters 6/0/100/15; (b) record 82, state 58, parameters 4/1/400/60. Each supplied selected record uses radius reference 0/duration reference 3 and complete pool[4,0,0,4,2,0], with R06's other named callback fields. These references supply complete equivalent expressions, not a new original-pool identity. T has active selected state, state 57 inactive, actual resolved stat 109=0 under the existing read contract, and primary chain[L,T105]. L is nonextended with curse condition true, selected state/skill, level 1, expiry 1500/timed=true, own entries empty, source owner monster/id 7, removal callback K, parent/association T's base/T. Valid pending queues are initially empty. T's existing event list in order[X,Y,Z]: X is skill kind/owner identity selected state/event 1/callback selector 5/values(99,9); Y is item kind/owner identity selected state/event 99/callback J/values(11,12); Z is skill kind/owner identity 9/event 99/callback H/values(13,14). All entries are nonexecuting, not one-shot and not pending removal; other owner collections empty. Case(c) is an independent copy of(a) except L.level=2, greater than caller level 1. No applier/registration result or allocator effect is supplied. Consume IEP-R01/R05's verified logical entry operations at normal continuation; physical allocation/release is not asserted. Observers log refresh, cancellation key/owner, entry creation fields, resulting logical order and actual returns.
  - Seeds: (a) A={lo:0,hi:3}; (b) A={lo:0,hi:2}; (c) A={lo:0,hi:3}; other seeds unchanged.
  - Draws: exactly (A,modulus 5,result 3) in(a), (A,modulus 5,result 2) in(b), and (A,modulus 5,result 3) in(c); no further direct or reached draw.
  - Trace: cases(a/b) select 87/82, set latch, evaluate radius 6/4 then duration 100/400; prepare no primary/additional stats or auxiliary operation. Actor distance 50 skips, T distance 1 passes each squared radius and R08. Actual player eligibility/relationship pass; zero resistance/inactive 57 retain duration; existing curse L has matching metadata. Refresh expiry to 1100/1400 and schedule T there, returning L. Actual recipient cancels T's skill entries keyed by state 60/58, removing X and retaining Y/Z. Case(a) creates E1(event 1,selector 5), then E2(event 2,selector 5); case(b) creates E5(event 5,selector 4). These are actual IEP logical creations, not supplied results. Return recipient 1; normal traversal/action finish 1. Case(c) computes(a)'s radius/duration but actual lower-level rejection returns no list before refresh or event processing; recipient 0 is ignored by the outer traversal, which returns 1.
  - Outputs: cases(a/b) A.seed={lo:3,hi:0} in(a),{lo:2,hi:0} in(b), latch=true; actual applier result L, recipient 1 and final action 1. Ordered effects(a)[expiryWrite(L,1100),schedule(T,type 12,1100,no override,0,0),removeSkillEvents(T,ownerIdentity=60),create(E1,T,event 1,selector 5,skill kind,ownerIdentity=60,values 87/1),create(E2,T,event 2,selector 5,skill kind,ownerIdentity=60,values 87/1)]; (b) expiry 1400/schedule 1400 precede removeSkillEvents(T,ownerIdentity=58), then create(E5,T,event 5,selector 4,skill kind,ownerIdentity=58,values 82/1). Final T event order(a)[E2,E1,Y,Z],(b)[E5,Y,Z]; X unavailable, Y/Z fields/order unchanged and other owners' event collections empty. L retains own empty entries, source monster/id 7, callback K, state/skill/level, timed=true and chain; only stated expiry changes. Target active membership and other stats unchanged; one pending type 12 event per case. No additional stat, AI/auxiliary operation, callback invocation or extra draw. Case(c) has A.seed={lo:3,hi:0}/latch=true, applier absent/recipient 0/final action 1; L.expiry=1500/timed=true, pending queues empty and event order[X,Y,Z] with every field unchanged; no cancellation, creation or other effect occurs. No original list installation, physical allocation/release/failure, later expiry or natural combat event production is asserted.
  - Discriminates: keeping old registrations merely because L was refreshed leaves X. Cancelling by selected skill 87/82 or invoking skill 212 instead of target state 60/58 also leaves X; removing all skill entries incorrectly removes Z, and ignoring kind removes Y. Registering on actor A leaves T's expected list absent. Appending rather than prepending changes(a)'s E2/E1 order. Substituting invocation 212 or stale X values 99/9 violates new callback values. Running callbacks during registration adds forbidden effects. Cancelling or registering before confirming a non-null applier changes(c)'s unchanged event list.
  - runtime_observed: no

- **MCA-V51** Registered selector-5 entries drive actual sequential life writes (MCA-R14, MCA-R19, MCA-R21 with IEP-R07; boundary: existing records produced by V50(a), explicit event-dispatch requests through actual callback/storage and dispatcher returns)
  - Inputs: V50(a)'s resulting T state 60/list L/source metadata, event order[E2,E1,Y,Z], selected record 87 callback fields and valid G/frame 1000; entries not executing/one-shot/removal-pending. Do not execute a new action or infer a combat producer. Issue IEP-R07 requested event 1, then event 2, each with owner=T, other=Q and the same incoming logical combat record I: physical 512 and every other field 0. Q is a distinct player/mode 1, aliveExclusionCondition=false, with associated extended root base, owner type player, no parent, secondary=false, sorted own[(6,0,200),(7,0,1000)], resolved[(6,0,300),(7,0,1000)], both counts 2/capacities 8 and empty changed-key collection. Stat6 has MRH-R08's captured direct controls, fMin=false/MinAccr=0/ValShift=8; stat 7 present/fMin=false, definition count 359. T has no source registry match for monster/id 7; game and registry are readable and source kind valid. Overlay count 200, captured prgoverlay=-1, selected `calc1` empty reference 4294967295. The actual dispatcher, callback, existing-stat reads and R19 writer execute; no callback/writer result, view update or new allocation is supplied. Observers log dispatch order/arguments, calculation front end, current/max reads, replacing requests and resulting arrays.
  - Seeds: no draw; all seeds unchanged.
  - Draws: none.
  - Trace: event 1 scans E2 then matches E1, invoking selector 5 with(game G,event 1,owner T,other Q,I,skill 87,level 1). Existing owner-state/list lookup resolves old source monster/id 7 to absent; the empty calculation returns 0. Q current 300/max 1000 and amount 0 request replacing life 300; actual own delta 300-200=100 makes own/resolved 300/400. Callback returns 1 and dispatcher returns 1 after scanning Y/Z without matches. Event2 then matches E2 and invokes the same callback roles/values with event 2. Q's actual next current read is 400/max 1000; zero amount requests replacing 400, delta 400-300=100 makes own/resolved 400/500. Remaining event 1/99 entries do not match; callback/dispatcher return 1. No entry removes itself or is pending removal.
  - Outputs: ordered callback log[(event 1,T,Q,I,87,1),(event 2,T,Q,I,87,1)]; calculation front-end contexts[(absent source,87,1,empty),(absent source,87,1,empty)], no bytecode evaluation. Ordered current/max reads[(Q,300,1000),(Q,400,1000)]; replacing requests[(Q,stat 6,layer 0,300),(Q,stat 6,layer 0,400)]. Actual final Q own/resolved life 400/500, stat 7 values 1000, counts 2/capacities 8/backing stores unchanged; changed-key collection remains empty. Callback and dispatcher results[1,1]. T and L, all event fields/order[E2,E1,Y,Z], pending expiry event, incoming I, source actor A and seeds unchanged; executing conditions cleared after each callback. No overlay operation, stat notification, state/list-chain change, registration/removal, allocation/resize/release or draw in this closed event/life chain. Natural combat dispatch production, original installation, later timer firing and wider stored effects remain gated.
  - Discriminates: dispatching with actor A as owner or Q as owner fails the existing state 60 lookup; swapping owner/other writes the wrong unit. Using invocation 212 or stale X values 99/9 selects the wrong callback context. Suppressing zero-calculation writes leaves Q own 200/resolved 300. Assigning each request directly to resolved life prevents 400/500 and changes the second current read. Reading both currents before the first actual write makes the second read 300. This consumes actual registered metadata and storage, not supplied callback success or stat-view updates.
  - runtime_observed: no

### 8.2 Observation vectors

**MCA-O01** Observe the selected action and its final return/effects on an
owned Expansion 1.13c game after establishing the missing bindings. Expected
values for gated effects are deliberately unspecified; the
[procedure](../research/spec-evidence/skill-monster-curse-and-aura.md#observation-procedure)
defines what evidence must be retained. This is a pending investigation, not a
matching observation.

## 9. Acceptance and dependencies

MCA-V01 through MCA-V07 establish MCA-R01 at the selection stage; MCA-V08 through MCA-V20 establish R04/R05 with their supplied records and operation results; V16 also includes R01. V21 through V26 establish captured formula, local predicate, aura request and zero-primary callback stages. V27 through V29 establish actual outer traversal/return using the named coordinate/room dependencies and actual index-4 recipient rejection. V30 through V36 establish callback arithmetic, exact roles and ordered requests at their explicit existing-view/dependency boundaries. V37 establishes the actual curse-policy/search prefix; V38 follows two real rejecting prefixes through the produced normal action result, without a supplied successful application. V39 establishes existing equal-level curse refresh, its actual expiry store/scheduling and fixed index-0 successful continuation without supplying an applier result. V40/V41 establish actual prepared own/resolved life writes and subsequent reads through MRH-R08/STL, without supplying writer effects; earlier combat and later overlay storage remain excluded. V42/V43 establish early callback and actual recipient rejection frontiers. V44 establishes aura preparation up to initial allocation; V45/V46 establish actual fixed collection with supplied prepared storage, including the fifteenth-entry growth-request frontier. V47 establishes the actual collector's absent-path/room, remaining fixed guard and signed wrapping cases. V48 establishes independent noncurse aura request construction, actual existing-list refresh and forwarded optional-list helper results; no fixed curse-recipient/action success is inferred for that aura stage. V49 independently follows two actual nonzero curse-recipient results through continued traversal and final action return. V50 establishes actual existing refresh followed by fixed skill-event replacement under IEP; V51 consumes those actual registered entries through explicit event requests and sequential R14/R19 life writes. These fixtures do not establish initial allocation, growth/release effects, complete stored callback/aura effects or gameplay confirmation.
Full reproduction requires resolving MCA-Q03, MCA-Q04 and MCA-Q01, specifying executable vectors for
MCA-R02, and retaining attributable observations under MCA-Q02. A partial result
does not establish parity for these actions.

**Depends on:** [skill dispatch](skill-cast-start-and-do-dispatch.md) SCD-R06, SCD-R07 and SCD-R08 for selecting
the action from the server selector and invoking its handler; successful effect
integration remains MCA-Q01. No damage, missile or monster-creation outcome is
credited through an unnamed dependency.

**Depended on by:** no established consumer.

[Skill calculations](skill-calc-expressions.md) SCX-R11 supplies the expression front-end contract and retains producer-specific gates; GDA-R11a in [game-data records](core-game-data-table-records.md) supplies the Param7 schema. SCD-R01/SCD-R07/SCD-R09 name the shared do-handled latch.

R07 consumes SCX-R01/R04/R20/R21 for the complete captured expressions.
R09 consumes [record/stat reads](core-record-layouts-and-ids.md) RLI-R02a/R02b and
[status gates](combat-status-application.md) CSA-R02's unitGate classification.
R10 consumes [ownership and relations](combat-pvp-damage-and-kill-credit.md)
PVP-R05a/R07, [search availability](skill-assassin-sentry-firing-and-range.md)
SFR-R03e's unavailable-for-combat subcondition, and GDA-R09a's mode availability.
The actual state consumer retains [Open Wounds](combat-open-wounds.md)
OWD-R06a/R06b/R07a/R07b/R08 and [timed state lifecycle](state-timed-state-lifecycle.md)
STL-R05/R06b/R06c/STL-Q07 plus RLI-Q02/Q03; a request does not close those
stored-effect gates. Event insertion/removal/traversal retain IEP-R01/R05/R07
in [item events](combat-item-event-triggers-and-procs.md), with callback arithmetic/requests specified by R14-R16 and complete stored
effects still Q04.

R13 consumes [monster target/room operations](skill-monster-teleport.md)
MTP-R09 for refresh/coordinates and only MTP-R04's rectangle-search subrecipe.
Dynamic coordinate integer words consume NRF-R10b; R13's squared arithmetic
uses NRF's wrapping operations. No teleport movement, linked-monster fallback,
supplied placement result or lifecycle propagation is inferred through these
coordinate/room references.

R14-R16 consume IEP-R07's event roles, RLI-R03's existing state membership,
SCX-R11/R20/R21's calculation references/parameters and NRF-R05's ordered
mulDiv. R14/R16 consume LSE-R10's effective life cap and LSE-R13's overlay
operation, preserving IEP-Q09/RLI-Q02/RLI-Q03 stored effects. R15 consumes
DPO-R08c/R08d/R13a/R14 through R19/R23 and DPO-R26's reaction handoff;
combat gates remain DPO-Q19/Q20/Q24/Q30. The reaction consumes
[hit recovery](combat-hit-recovery.md) HRC-R02a/R02b/R03a/R04a/R07a,
with HRC-R08/HRC-Q06 retaining terminal/update integration.
R16 consumes LSE's named MonStats.Drain difficulty selection and PVP/RLI
existing alignment; its distinct capped-add is specified locally. R17 consumes OWD-R06a and
OWD-R07a/R07b/R08 at their explicit boundaries, RLI-R03 for state 57 and
state-list lookup, RLI-R02a/R02b for stat 109, and GDA-R13 for the selected
state membership association. It establishes only this consumer prefix; it
does not upgrade the shared alternate-profile OWD-R06b contract.
R18 consumes OWD-R07a's same-state/skill/level refresh, STL-R06a's actual
expiry setter and STL-R01's scheduling contract in its valid frame domain.
It does not consume a supplied successful allocation or infer attach/stat
propagation through the shared prepared direct-stat slices.
R19 consumes [monster healing](skill-monster-resurrect-and-heal.md) MRH-R08's
explicitly reusable isolated unit-life writer, including its captured stat-6
controls and player changed-key exclusion, and through it STL-R06d/R06e.
RLI-R02a/R02b selects the resulting resolved life on subsequent unit reads.
The source of the initial arrays is not inferred; broader MRH-Q01/STL-Q07/
RLI-Q03 storage and MCA-Q04's combat/overlay gates remain.
R20 consumes MTP-R09 for target refresh/coordinates, NRF-R10b for the
coordinate words and GDA-R11a for Param4. It reuses only R13's exact
room-prefilter subrecipe; its room-source choice, X-then-Y ordering and
saved-next-link traversal remain as stated in R20. No teleport movement or
relationship result is inferred through those references. Collection storage
allocation/growth/release remains MCA-Q03; the compatible shared allocator
investigation retains STL-Q07 without supplying a successful result.
R21 consumes IEP-R01/R02/R05/R06 for the actual logical event collection
operations and shared callback selection; V51 consumes IEP-R07 for dispatch
roles, ordering and returned result. Physical allocator/release internals are
not reverified through those semantic contracts. Skill-specific callback
effects remain R14-R16/R19 and their existing gates.

## 10. Related specs

[Skill dispatch](skill-cast-start-and-do-dispatch.md) supplies the action context.
[Numeric foundations](core-numeric-and-rng-foundations.md) owns integer and seed
operations when their consumption is established.

## Appendix A. Provenance

Original Windows x86 Expansion 1.13c D2Game, identified by
[source manifest](../research/source-manifest.md), was examined through static
instruction listings on 2026-10-29 through 2026-11-02. Narrow loaded-field and
compiled-expression observations on 2026-10-30 retain the companion's source-profile limits. The
[companion](../research/spec-evidence/skill-monster-curse-and-aura.md) identifies the dispatch slots,
read boundary and unresolved callees. Named selector text cells have the companion record; no gameplay case has been confirmed.
