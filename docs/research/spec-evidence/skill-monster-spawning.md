# Evidence: monster skill spawning and calls for help

Contract: [monster skill spawning and calls for help](../../specs/skill-monster-spawning.md). [Evidence gaps](../../plans/gaps/skill-monster-spawning.md) retain unverified integrations.

## Identity and reuse

Windows x 86 original Expansion 1.13c `LOD113C-D2GAME` is identified by the [source manifest](../source-manifest.md). SHA-256 of the read image was checked on 2026-10-29 and matched `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`. All locations below are RVAs; descriptions are analyst labels.

Reuse of [shared dispatch evidence](skill-cast-start-and-do-dispatch.md) preserves its 2026-10-12/15 dates and gates. New local readings do not establish its unresolved AI/frame producers. No earlier candidate-specific companion existed. Compatible text-data evidence is the exact 2026-10-29 loader read described in MGB-D01; its original identity and coverage limits apply without renewal.

## Function records and roots

| Record | Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|---|
| MSP-F01 | LOD113C-D2GAME | 0xA19A0 -> table 0x1074A8:91 -> 0xB5E50;128 -> 0xC310;133 -> 0xC5D0;135 -> 0xCDC0 | none | Action dispatcher and selected actions | MSP-R01–MSP-R06 | Reused caller; fresh original-image table slots 2026-10-29 |
| MSP-F02 | LOD113C-D2GAME | 0x9FD00 -> table 0x107338:49 -> 0xB5FB0;62 -> 0xBF90 | none | Start selection | MSP-R01/MSP-R04 | Reused caller; new table/entry instruction reading; complete preparation effects unresolved |
| MSP-F03 | LOD113C-D2GAME | 0xCDC0 | none | Minion-spawn action | MSP-R02/MSP-R03/MSP-R04 | Full local body read; original-image disassembly 0xCDC0–0xCF00 corroborates encoded spans, both creator calls and exits |
| MSP-F04 | LOD113C-D2GAME | 0xB5E50 | none | Nest action | MSP-R02/MSP-R04 | Full local instruction body; original-image disassembly 0xB5E50–0xB6010 corroborates encoded spans and part of start 49 |
| MSP-F05 | LOD113C-D2GAME | 0xC5D0 -> 0xA00B0 / callback 0xC000 | none | Impregnation request | MSP-R02/MSP-R05 | Local caller read; full callback/iteration integration unverified |
| MSP-F06 | LOD113C-D2GAME | 0xC310 -> 0x212F0 | none | Help request | MSP-R02/MSP-R06 | Local caller read; target-associated effect implementation unverified |
| MSP-F07 | LOD113C-D2GAME | 0xEF650 -> 0xEEF10 | none | Generic creator request builder and core | MSP-R04 | Complete request builder read; main creator entry/class/placement setup inspected; full creation/seeding/registration chain incomplete |

Selected callbacks use ECX game, EDX unit and stack skill id/level, supported by the reused dispatcher and entry instructions. The spec's supplied-operation notation does not prescribe that binary interface.

## Guard and minion fallback rationale

Each selected action checks its signed skill id against 0 and the global Skills count before its first call or gameplay write. The selected invalid paths return 0; original-image listings corroborate the encoded MinionSpawner/Nest guards. No unread helper is inside this invalid-id slice.

0xCDC0 obtains current skill via thunk 0xAABA. It reads four current-skill values through thunks 0xAD12,0xACB2,0xAD3C,0xAB44. The values are forwarded as class, x, y, mode respectively in the creator request. The caller obtains a room through 0xA478 and finds the coordinate room via local 0x1280. The complete local room helper body uses signed bounds and an imported adjacent-room list; its external list producer remains unresolved.

With room non-null, the first call 0xCE65 -> 0xEF650 carries ECX class, EDX mode; stack game/room/x/y/-1/0. Empty EAX causes the second call 0xCE83 -> 0xEF650 with identical class/mode/game/room/x/y and final stack 4/0. Its result is retained in ESI. Either nonempty result joins at 0xCE9A; both empty go to 0xCE8E. This proves the supplied-result stage's precise order/selection. It does not establish any real creator side effects.

After selection, 0xCE9A sets created unit +0xC4 mask 0x04020000, invokes thunk 0xAC76 with the created unit and original caster, then optionally calls 0xA83E when signed skill state field +0xE6 is in range. Actual initialization/state semantics and final callback effects require complete dependency evidence. The creator request-builder maps class/mode/context/coordinates/placement arguments into its request and calls 0xEEF10; that body's MonStats bound/lookup, room/placement preparation and later local calls show why a returned pointer alone cannot stand for a complete initialized monster.

| D2Game thunk | Original resolved target | Supported caller reading / limit |
|---|---|---|
| 0xAABA | LOD113C-D2COMMON ordinal 10511, RVA 0x303A0 | Current skill; compatible shared dispatcher mapping |
| 0xAD12 | LOD113C-D2COMMON ordinal 10980, RVA 0x4E030 | Byte-equivalent C read: returns current-skill field +0x18 on non-null input |
| 0xACB2 | LOD113C-D2COMMON ordinal 10649, RVA 0x4E000 | Byte-equivalent C read: returns field +0x1C |
| 0xAD3C | LOD113C-D2COMMON ordinal 10179, RVA 0x4DFD0 | Byte-equivalent C read: returns field +0x20 |
| 0xAB44 | LOD113C-D2COMMON ordinal 10631, RVA 0x4DFA0 | Byte-equivalent C read: returns field +0x24 |
| 0xAC76 | LOD113C-D2COMMON ordinal 10210, RVA 0x33020 | Owner-link request; nonnull second argument is original caster, not creation mode |
| 0xA83E | LOD113C-D2COMMON ordinal 10689, RVA 0x32C10 | Overlay request; complete stat-list/notification effects not verified here |

The four field-getter bodies assert on a null argument. MSP-R03 supplies prepared values after a non-null current-skill node; it does not infer the assert-path behavior or the fields' production from the getter labels. No complete preparation contract follows from resolving their imports.

## Targeted remaining paths

Nest 0xB5E50 traverses a caster event list and conditionally invokes 0x8E6F0 before 0xE2240, then obtains prepared current-skill values, checks coordinate room/coordinates, calls a collision operation and creator 0xEF650, and applies a created-unit mask and optional overlay. Local original bytes settle those encoded calls; full event-list removal/reset/collision/creator effects remain MSP-Q01/MSP-Q04. Start49 0xB5FB0 selects a monster class through 0x9EE50 and writes current-skill preparation fields. The full source path, not merely its four getter values, is still required. Start62 0xBF90 is a separately selected producer and cannot silently reuse start 49's interpretation.

Impregnate 0xC5D0 sets the caster action latch on a valid id, calls preparation 0xE07A0, resolves a selected target, rejects self/absent targets and calls predicate 0xC110. It builds a target-associated request containing skill id/level, sentinel -1, state selector 110 and callback 0xC000, then calls effect producer 0xA00B0 and returns 1. The target predicate, iterator and callback interiors remain unresolved; no delayed monster creation timing is invented. These are located paths, not a six-rung not-found claim.

Cry Help 0xC310 resolves a selected target, requires caster type 1, computes max(`calc1`,1), adds it to the current game frame, constructs a request with target type/id and calls 0x212F0 before an optional target-state request via 0xA83E. Actual event ownership, timing semantics and responder behavior require the request consumer; the skill name does not prove a spatial radius or direct creation.

## MSP-D01: effective Skills text and compiled selectors

The full 2026-10-29 loader-read profile, exact ranges, coverage checks, header-selected parsing and identity limits are in [MGB-D01](skill-monster-maggots-and-burrowing.md#mgb-d01-effective-skills-text-cells). Exact skill names select the following individual cells; `Id` is explicit, not positional. No whole row/table is distributed.

| Row | Column | Raw cell | Claim |
|---|---|---|---|
| Nest | srvstfunc | `49` | MSP-R01 |
| Nest | srvdofunc | `91` | MSP-R01 |
| EvilHutSpawner | srvstfunc | `49` | MSP-R01 |
| EvilHutSpawner | srvdofunc | `91` | MSP-R01 |
| MinionSpawner | srvstfunc | `62` | MSP-R01 |
| MinionSpawner | srvdofunc | `135` | MSP-R01 |
| Impregnate | srvdofunc | `133` | MSP-R01 |
| Cry Help | srvdofunc | `128` | MSP-R01 |

On the same date/profile, read-only live DATATBLS inspection used the narrow procedure in MGB-D01. Each explicit id/start/action record read succeeded: (167,49,91), (320,49,91), (302,62,135), (300,0,133), (290,0,128). These agree with the named text cells and original table slots. No calculation execution, start-prepared values, AI choice or actual spawn was established. Loaded-module hashes, effective-text hash, complete override inventory and physical MPQ provenance remain uncaptured.

## Observation procedure

MSP-O01: on an owned original LoD Expansion 1.13c game, use a character able to survive a nest encounter on Normal, then a separate Act V minion-spawner encounter. First verify the exact selected loaded skill and start/action entry. Record current-skill prepared class/mode/position, room and every creator argument/result. Observe both open placement and obstructed placement; a fallback is confirmed only when the first request actually returned empty and the second request occurred. Record later initialization/state and stored creature identity separately from the local request stage.

For Impregnate, establish the eligible target via the selected predicate, capture the target-associated request and its actual later callback/creation. For Cry Help, record selected target, computed frame/request and subsequent actual consumer, rather than count nearby monsters as proof of a help event. Exact successful expected effects remain gated pending those consumer contracts. No spawn/help gameplay capture was collected in this pass. The unrelated healing observation supplies no spawning evidence.

Invalid-id MSP-V01 may be unreachable through normal dispatcher validation; its static guard proof is not replaced by an injected call. MSP-V02 is a supplied-operation vector, and no manufactured memory/call result is claimed as gameplay confirmation.

## Review status

Initial handoff record, 2026-10-29: self-review checked exact fallback arguments/order, stage result versus callback result, helper boundaries, cell/compiled evidence, source identities and all explicit gates; mechanical lint of this trio with the two sibling candidate trios returned exit 0, 0 errors and 0 warnings. Fresh blind-reader completion was pending at that handoff. The prior full pass 2 below completed the initial fallback-only contract; the expanded preparation/request contract completed the fresh full review and final focused delta review recorded below. All unresolved integrations remain draft; passing a reader or lint does not establish gameplay evidence or user approval.

### Fresh blind-implementer full pass 1 — 2026-10-29

An independent smaller reader used this contract and cited dependencies without binary, source, gameplay or implementation inspection. Counts were MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, COVERAGE 0, HARDEST 3. The report labeled GATED count 7 while enumerating nine ids: MSP-Q01–MSP-Q06 and MSP-R04–MSP-R06. All nine enumerated ids are existing documented gates; the raw count discrepancy has no effect on their scope.

| Finding | Outcome | Rationale / resulting change |
|---|---|---|
| DISCRIM: equal class/mode inputs cannot reject swapping those arguments | Spec edit | MSP-V02 now supplies distinct class 7 and mode 3 and requires those exact values in every request. Both the initial and fallback logs reject swapping them |
| HARDEST: fallback order and complete request argument preservation | Spec edit for argument discrimination; otherwise no change | The distinct argument values fix the reported weakness; MSP-R03 and MSP-V02 already distinguish first success, fallback success and both-fail selection |
| HARDEST: signed id guard | No change | MSP-R02 states signed lower and exclusive upper bounds; MSP-V01 exercises both rejected boundaries before any helper access |
| HARDEST: distinguishing supplied requests from gated creation effects | No change | Scope, stage preamble, MSP-R03 and acceptance explicitly exclude real creator/initialization effects and identify MSP-Q01; the vector selected result is distinct from the callback result |

All exact logs were recomputed with the changed inputs from unchanged MSP-R03. No source label, rule behavior or gate changed. Mechanical lint of all three candidate trios after these edits returned exit 0, 0 errors and 0 warnings. Fresh full pass 2 is required because vector inputs and expected request arguments changed.

Before full pass 2, MSP-V02(c)'s two-request output was spelled out with the same exact values instead of referring to case (b)'s log. This removes cross-case output inheritance without adding or changing an assertion, input, result or gate. The fresh pass receives the fully spelled-out final vector.

### Fresh blind-implementer full pass 2 — 2026-10-29

The fresh smaller reader's final applicable report returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, COVERAGE 0, GATED 9, HARDEST 3. The nine enumerated gates are MSP-Q01–MSP-Q06 and MSP-R04–MSP-R06, matching the normalized enumeration of the first pass. The earlier report's raw count 7 and nine-id enumeration remain recorded above. No further behavioral edit was recommended or made.

| Finding | Outcome | Rationale |
|---|---|---|
| HARDEST: guard ordering | No change | MSP-R02 rejects invalid signed ids before helper calls, writes and draws, with both bounds exercised by MSP-V01 |
| HARDEST: exact fallback and argument preservation | No change | MSP-R03 and final MSP-V02 independently specify first success, fallback success and both-fail logs, preserving distinct class/mode and all context/position/options |
| HARDEST: selected stage result versus callback/real-creator effects | No change | The supplied-input boundary, MSP-R03, vector preamble and acceptance distinguish the selected identity from the outer callback result and gate actual creation/finalization under MSP-Q01 |

MISMATCH is zero and no unresolved GUESS/GAP/DISCRIM remains. The final applicable full pass completes the blind-implementer stopping rule for the unchanged resulting contract. All source labels, vectors, gates and draft status remain. This evidence-only record establishes neither complete gameplay behavior nor user approval.

## Successful preparation followup

On 2026-10-29 this writer checked the existing MSP evidence, gaps, completed
fallback-stage reviews and preserved earlier audit before reading missing
slices. Exact numeric matched-source manifest checks found no entries for
D2Game 0xBF90/0xB5FB0/0xC310/0xC5D0/0xC110/0x212F0/0x9EE50/
0xA00B0 and D2Common 0x1A3D0/0x33020/0x32C10. Applicable instruction
listings were read; original-image disassembly independently corroborated
the new caller branches, helper bodies and embedded callback address. Source
forms, exact commands, excluded diagnostic partial ranges and archive paths
remain in the separate private audit. No binary or data file was edited.

### Common spawn-tuple production and fallback

D2Game 0x9EE50 pushes seven arguments into thunk0xAED4, resolved from
the original PE tables to D2Common ordinal 11086 RVA0x1A3D0. Argument meanings
established by caller and consumer are actor, monster-spawn preference, skill
identity, level, output mode, output x and output y. The helper's analyst
collision label is not its identity. Its whole body was read. With nonzero
preference and a monster actor, it selects the actor's MonStats record;
class is signed word+0x20, x/y add signed bytes+0x22/+0x23 to position,
and signed mode byte+0x24 is clamped to 1 outside 0..15. The skill-record
branch instead selects signed summon+0xBC and mode+0xBF, rejecting invalid
summon bounds before output writes and retaining unshifted actor coordinates.
There is no use of the forwarded level inside these branches and no RNG.

Original descriptor construction pairs MonStats names spawn/spawnx/spawny/
spawnmode (strings at0x93014/0x9300C/0x93004/0x92FF8) with offsets0x20/
0x22/0x23/0x24 at D2Common0x5AC52..0x5AD13. The selected Skills summon/
summode fields pair with+0xBC/+0xBF in descriptor construction0x63D3A..
0x63DBE. These local field-name associations do not claim actual loaded
values for any original named monster. The position getter bodies0x1050/
0x1080 have byte-corresponding source views, which were read on the followup:
dynamic paths read unsigned 16 components, static-position types read their
stored32-bit components; missing dynamic path returns 0. The initial exploratory
instruction read preceded noticing those source views; no independent
verification credit is assigned to that order error.

The D2Game wrapper checks primary result as signed >=0 and <MonStats count.
Only an invalid primary result takes the supplied actor AI-context fallback
field+0x3C; fallback must be strictly positive and <count, else result-1.
Already-written output mode/x/y are retained. If lower production returned
before writing fields, the wrapper does not initialize them; that error-path
producer context remains unverified and no default tuple is invented. MSP-R07
is explicitly a supplied-valid-record/output-stage contract, not universal
error recovery or actual AI record construction.

### Persisted start cache

Start62 table slot points exactly at D2Game0xBF90; that entry ends at0xBFFE,
before the independent callback entry0xC000. Start62 obtains current skill,
calls shared preparation with monster preference 1, then writes returned class,
x,y,mode via thunks0xAC94/0xACDC/0xAD0C/0xAB20, without testing class.
It returns 1. Start49 at0xB5FB0 uses the same preference and writes, but
rejects negative returned class before all four stores. Successful stores are
followed by room/collision, actor preparation and event cancellation paths;
those are outside R08's cache checkpoint and remain Q04.

The four setters resolve respectively to D2Common ordinal 10243/+0x4DF60,
10554/+0x4DF20,10748/+0x4DEE0,11031/+0x4DEA0. Their byte-corresponding source
views were read first: present skill nodes receive the supplied32-bit value
at+0x18/+0x1C/+0x20/+0x24 respectively. The matching getters and creation
consumer already read in the prior record identify these semantic fields as
class,x,y,mode. Null-node assertion paths are outside R08's present-node
precondition. The stores are persisted state, not just hypothetical call logs.

### Nest request schedule and minion owner correction

The original Nest action0xB5E50 was reread with its existing original-image
corroboration. After its earlier event/preparation phase, usable coordinate
room and nonzero x/y permit collision thunk0xAD30 with room,x,y,1,0x100,
then one creator0xEF650 with selector 1 and final option 0. Collision return
is not tested; creator empty returns the action's failure exit, and no fallback
is present. R09 ends at the selected creator result and does not credit
the later created-unit mask/overlay writes. Real collision and creator effects
remain integration gates, independently of the exact request schedule.

The prior minion followup described0xAC76(created,preparedMode). Independent
stack tracking corrects its second argument to the original caster: prepared
mode was stored at a distinct local slot and only forwarded into creation.
The actual call reaches D2Common ordinal 10210/+0x33020; on nonnull second
argument it reads that unit's type/id and calls0x32CA0. That helper writes
the created unit's owner identity fields, then may create/update an owner
stat list and notification state. Both bodies were read, but the deeper
list/stat/notification effects remain Q01; no complete owner-link integration
is promoted here. This correction prevents interpreting an integer mode as
a unit pointer and does not change MSP-R03's creator arguments.

### Impregnate request and removal operation

The source-relative alias L_9F0B0 in the original action listing identifies
actual D2Game RVA0xA00B0, not RVA0x9F0B0. Original-image call decoding
settles the target. The prior iterator wording was incorrect: this is the
shared target-associated state/list producer. Its whole instruction body was
read, including state eligibility, existing-list matching, duration and list
construction, but its imported list/notification contracts are not all closed.

At0xC5D0, valid lookup sets the caster latch, refreshes/resolves target,
rejects self/absent targets and applies predicate0xC110. After eligibility,
the caller clears a nine-word request and stores actor,target,skill,level,
duration 0, primary-stat sentinel-1, primary value 0,state 110 and removal
callback0xC000. The producer return is ignored and action returns 1. This
establishes R10's supplied eligible-target request, not actual state creation.

Predicate0xC110's reachable entry requires a monster target, zero result
from0xDFA90, excludes normalized classes 546/551, rejects existing state 110,
and rejects one alignment result. Exact normalization/relationship imports
remain Q02; labels alone are not used. Callback0xC000 was original-image
decoded from its true entry, despite being included inside the broader
0xBF90 listing. It first requests state removal, obtains associated skill
identity, defaults spawn class 551/mode 1 and conditionally uses skill summon/
summode. Its actor flag/death-mode guard, room predicate and creation wrapper
0xEDDB0 determine whether a spawn is requested. Actual removal timing,
stat-list lifecycle and that callback's stored effects remain Q02. Thus the
action's duration 0 is not evidence of immediate expiry or immediate spawning.

### Cry Help request subset and actual propagation lead

The exact action0xC310 ends at0xC3F7; following missile logic in the larger
listing is a separate entry and is not part of this reading. After a selected
distinct target and monster actor, the action evaluates selected `calc1` through
thunk0xAC9A (D2Common ordinal 10786/+0x51BF0, existing SCX-R11), clamps the
signed result below 1 to 1, and adds the game frame with low32 wrapping.
It supplies propagation kind 1,target type/id and that deadline to 0x212F0,
then optionally requests the target overlay and returns 1. Original stack
tracking establishes request+8=kind,+0xC=type,+0x10=id,+0x14=deadline;
the next copied word is not established by this caller and remains Q03.
Vectors deliberately assert only this known request subset, not invented
values for unwritten context.

The whole0x212F0 body and local0x20F30 were read and original-image decoded.
The consumer follows an actor-owner relationship, finds a monster owner and
iterates that owner's associated creature ids; resolvable creatures obtain
a command-list entry and five payload words are copied into it. This is not
a radius scan or an immediate create operation. Owner-record semantics,
command allocation/lifecycle and actual responder processing remain Q03.
This reading localizes the missing effects rather than inferring them from
the skill name.

The optional operation0xA83E resolves D2Common ordinal 10689/+0x32C10.
Its body was read: overlay bounds are at shared count+0xBC0, and it writes
stat 178 into a stat list then requests refresh. It is not a plain state-bit
setter (the state count is a different field). The old optional-state wording
is corrected to optional overlay; complete list/refresh effects stay Q01/Q03.
Minion/Nest use the selected creation overlay field+0xE6 and permit selector 0;
Cry Help uses field+0x4E and requires selector>0. The new R11 states only
the latter caller's strict branch and logged request, not a new overlay owner.

### Review and observation limit for the continuation

The initial expanded handoff required a fresh full blind reader for R07-R11/V03-V07; that full review and the final focused delta completion are recorded below. Earlier completed
fallback-stage reviews above remain historical and do not review these new
assertions. No original-game observation was collected in this continuation.
All exact data-profile and original binary identity limits above are retained.

### Expanded preparation full review 1 — 2026-10-29

Fresh smaller-model full review of R07-R11/V03-V07 and the retained contract:
MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, advisory COVERAGE 7, HARDEST 3.
DISCRIM V06 final latch alone did not reject setting it after the effect call:
edited the supplied effect operation to capture actor latch at entry and
assert true there in each independent producer-return case. The discriminating
assertion now rejects a late write, without changing R10's existing order.

All advisory coverage entries were triaged without changing the contract:
R07 nonmonster with preference set uses the same skill-record alternative
whose output construction is exercised by V03(e); no new computation/result
branch, so no change. R07 invalid Skills.summon early output context remains
explicitly outside the supplied valid-record/output contract under Q04;
inventing an initialized tuple for that path would be incorrect, so no change.
R07 negative or count-bound fallback rejects to -1, the rejection result already
exercised by V03(d); positive fallback is V03(c), so no change. R08's start
negative-class/persisted-write alternatives are already independently covered
by V04(a-d), no change. R09 request/coordinate alternatives are covered by
V05(a-e), no change. R10 actual ineligible-target production remains Q02,
not inferred from V06's eligible-target request, no change. R11 negative or
above-count overlay values select the same skip operation/result as the zero
and equality cases in V07(b/c), no change. R11 calculation result zero uses the same clamp
to one as the negative case V07(a); positive one is V07(c), no change.
Wrapping addition is already exercised by V07(b), no change. The raw seven
coverage groups include these related subcases; no unreported execution claim
is added by grouping them here.

HARDEST primary failure retaining already prepared coordinates/mode: no
change, R07 and V03(c/d) state exact retained values. HARDEST start-specific
negative class and persisted cache writes: no change, R08/V04 distinguish
all four cases and assert persisted tuples. HARDEST signed clamp, wrap,
ignored help result and overlay order/bounds: no change, R11/V07 carry exact
calculation/request logs and outcomes. The V06 assertion change requires a
fresh reader on the resulting contract; prior fallback-only reviews remain
historical. No source reanalysis or other spec edit was required.

### Expanded preparation focused delta review 2 — final applicable record

Fresh independent smaller-reader focused delta check of the V06 producer-entry
latch observer and assertion returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0.
Both independent producer returns (empty and listL) were recomputed: exact
request, producer-entry latch true, final latch true and final result 1.
A variant setting the latch only after the producer enters captures false and
fails the new assertion. No further edit was recommended or made. Q02's real
producer/eligibility/lifecycle effects remain gated. The three HARDEST entries
were carried forward from the full review with their recorded outcomes; this
focused pass did not pretend to re-review unchanged material.

The full expanded review's only actionable discriminator finding is resolved,
with no outstanding MISMATCH/GUESS/GAP/DISCRIM in the final applicable delta.
This completes the requested fresh-reader stopping rule for the resulting
contract. Earlier fallback-only and expanded-review records retain their
historical scope. The spec remains draft with its existing creation, target,
event, responder, data/producer and observation gates. Review completion is
not gameplay confirmation or user approval. No contract, vector or gate
changed after this final pass.
