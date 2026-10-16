# Open Wounds

**Description**

Open Wounds gives a successful attack a chance to make its target bleed for a short time. The
bleeding strength depends on the attacker's level and the kind of target, and another successful
application refreshes the duration without stacking another bleed. It can finish off a monster;
against a player, the life drain stops at one hit point.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: OWD-R, OWD-Q, OWD-V, OWD-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-open-wounds.md)
Open evidence gaps: [gap file](../plans/gaps/combat-open-wounds.md)

## 1. Scope

This contract covers the Open Wounds item-event callback, its chance and magnitude, the
timed-state request and create/refresh policy, and the monster/player life steps that consume
the resulting regeneration value. It also bounds the monster tick restart and bleed-death
hand-off. Item-event registration/dispatch belongs to
[combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md); hit placement
belongs to [combat-damage-pipeline-order](combat-damage-pipeline-order.md); generic list
construction, attachment, expiry and removal callbacks belong to
[state-timed-state-lifecycle](state-timed-state-lifecycle.md). Poison and burn producers belong
to [combat-status-application](combat-status-application.md). Mana/stamina regeneration, death
sequence internals, network message formats and producers of other regeneration sources are
outside this boundary. Supplied list/stat effects make local arithmetic executable without
establishing the gated lifecycle integrations.

## 2. Trigger and ownership

The server computes this behavior. In the original data profile, item stat 135
`item_openwounds` registers selector 15 for events 5 `domeleedamage` and 6 `domissiledamage`
(OWD-R01). The event owner is the attacker; `other` is the defender. A missile supplies its
forwarded owner as attacker. An absent owner or defender makes this callback return 0.

DPO-R13a/DPO-R23 place the attacker event after the apply gate and before physical clamping,
leech, life subtraction and status application. The proc roll, state request and refresh or
creation requests run synchronously there. IEP-R07 visits matching entries newest registration
first and ignores callback results for traversal; different effects do not have a fixed
selector-number order. OWD-Q06 covers observation of their combined order.

Life changes occur in regeneration event type 3, reached from the server frame event dispatch.
The monster handler runs OWD-R09/OWD-R10; the player handler reschedules and runs OWD-R11
before the separate stamina and mana steps. DPO-R19c requests a monster restart after a
surviving, damaging hit (OWD-R12). State expiry uses event type 12, STL-R03/STL-R04. Equal-frame
and cross-group order remain gated (OWD-R13). The complete original dispatch chains are in the
[companion](../research/spec-evidence/combat-open-wounds.md#dispatch-and-function-evidence).

Active state, life and life-percent changes are candidates for server-to-client notification;
message identity and delivery are OWD-Q08, not a local packet contract.

## 3. Inputs and data

All arithmetic below is signed i32 with low-32-bit wrap unless marked otherwise. Signed
division truncates toward zero. Life and regeneration use raw 8.8 life units (NRF-R08a);
frames use NRF-R09a. Pseudocode names impose no language or API design.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| `A`, `D`, `G` | Attacker/event owner, defender, defender's game | Optional units; game/frame | The callback checks both units before reading chance; applier requires a target |
| `chance` | Attacker stat 135 `item_openwounds`, layer 0 | i32 | Existing effective percent value, not recomputed from equipped items here |
| `A.seed` | Attacker seed | Pair of u32 | NRF-R01 chance step |
| `eventId` | Item-event dispatcher | i32 | Event 6 additionally halves the player value |
| Attacker level | Stat 12 `level`, layer 0 | i32 | Curve input |
| Attacker identity | Unit type and id | u32 each | New-list owner; type 0 player, 1 monster |
| Defender identity/conditions | Type, class id, mode; `specialA`, `specialB`; `aliveExclusionCondition` | u32 type/id; i32 class/mode; booleans | Monster conditions are supplied neutral inputs (OWD-Q01); expiry condition follows STL-R11 |
| Defender environment | Game, allocation tag, current room, supplied `isTown` | Logical values | Allocation tag is passed unchanged; no room/town guard applies to players |
| Loaded monster definitions | Class count, optional MonStats record, signed i16 `MonStatsEx`, extended count and optional MonStats2 record | Counts i32; flags boolean | Boolean fields `MonStats.npc` (GDA-R08a) and `MonStats2.isAtt` (GDA-R09a) determine eligibility; a false flag tests 0 (GDA-R16) |
| State definitions | Count, optional state record; `curse`, `exp`, `aura` and life-state membership | i32 count; booleans/sets | Named loaded inputs; not regenerated by these rules |
| Existing state list | State, skill, level, expiry, owner type/id, timed condition, own stat entries, callback | i32 fields and logical references | RLI-R03 lookup; STL-R06a fields |
| Stat containers and definitions | Own/resolved arrays, selected-array condition, owner and `fMin` / `MinAccr` / `ValShift` | RLI section 3.7 | Existing-value reads consume RLI-R02a/RLI-R02b, not dynamic summation |
| `g`, life and maximum | Stat 74 `hpregen`, stat 6 `hitpoints`, stat 7 `maxhp`, layer 0 | i32 raw 8.8 | Unit reads select existing effective arrays; base read selects own array |
| Life-state test inputs | Unit/base-list presence, state-storage condition, active state words, life-mask words, state count | Logical records; u32 words, i32 count | OWD-R09b intersection test |
| Last life percent | Stat 352 `last_sent_hp_pct`, layer 0 | Read i32, use low u8 | Notification threshold in OWD-R09d |
| Curse resistance | Stat 109 `curse_resistance`, layer 0; state 57 `attract` | i32; active-state test | Alternate-profile branch OWD-R06b only |
| Pending frame events | Unit's ordered pending entries: kind, frame, running, pending-removal, target, parameters | Logical records | OWD-R09c; kind 3 regeneration, kind 12 expiry |
| Damage hand-off | Effective life; `record.total`, `record.lethal` | i32 raw 8.8; boolean | DPO-R19c / CSA-R01 consumed by OWD-R12 |
| Death inputs | States 2 `poison`, 54 `uninterruptable`, 62 `openwounds`; supplied registered-unit lookup by owner identity | List/state lookups, optional unit | OWD-R10's candidate hand-off; lookup inputs/results below |

`Stat(u,s,layer)` is RLI-R02a's unit read, with RLI-R02b's minimum adjustment. `BaseStat`
selects the base list's own entries with those same guards/adjustment; it does not traverse
attached lists. `MaxLife` reads stat 7 through the effective-array reader, returning 0 for
missing unit/base, definition count at most 7 or absent stat-7 definition. `SetUnitStat` is an
ordered unit replacing-write request; its stored effects remain RLI-Q03 / OWD-Q12. The
separate `SetListStat` consumes the list writer STL-R06c. The
selected effective arrays are supplied independently; their production is not part of a read.

`RegisteredUnitLookup(game,type,id)` is a supplied read-only operation in the death-tail
recipes. Its inputs are a present game and owner type 0..5 plus the u32 unit identity. It returns
the supplied matching registered unit or null when no such unit exists, with no mutation or
draw. Compatible original
lookup evidence is recorded in the companion; these recipes do not require a registry storage
design. No-list tail entry, malformed owner types and absent game remain OWD-Q10.

| Mapping / constant | Origin | Binding |
|---|---|---|
| Stats 6, 7, 12, 74, 109, 135, 352 and their names | Table data; code use | Names/IDs in `data/global/excel/itemstatcost.txt`, row `ID`, column `Stat`; original consumers fix these stat selections |
| Stat 135 events / selector 15 | Table data; code handler table | Row `ID=135`, `itemevent1=domeleedamage`, `itemeventfunc1=15`, `itemevent2=domissiledamage`, `itemeventfunc2=15`; data selects an engine-defined callback |
| Events 5/6 | Table identity; code dispatch | Zero-based `data/global/excel/events.txt` rows excluding header, `domeleedamage` / `domissiledamage`; original apply code fixes dispatched IDs |
| States 2, 52, 54, 57, 62, 92, 100, 115 | Table identity; code selection | `data/global/excel/states.txt`, explicit `id`, column `state`; fixed original code uses these IDs |
| State 62 profile | Table data | Row `id=62`, `state=openwounds`, `curse`, `exp`, `aura`, `plrstaydeath`, `monstaydeath`, `bossstaydeath` empty; `life=1` |
| Life-mask profile | Table data | `states.txt` rows `id=2 poison`, `id=52 preventheal`, `id=62 openwounds`: `life=1`; row 115 `burning`: `life` empty; no other row has raw `life=1` in the complete read |
| Stat 74 minimum | Table data | `itemstatcost.txt`, row `ID=74 hpregen`, `fMin` empty: minimum adjustment disabled |
| Curve slopes, added 40, duration 200, stat 74/state 62/skill 0/level 1 request | Engine code | Fixed original constants |
| Player divisors 4 and 2, special-monster divisor 2, roll modulus 100, percent threshold 4 | Engine code | Fixed original arithmetic |
| Regeneration 3, expiry 12; death notifications 10 then 9 | Engine code | Fixed event kinds; wider notification effects are owned by IEP-R09a |

Named table cells and data-profile limits are recorded in the
[table evidence](../research/spec-evidence/combat-open-wounds.md#named-table-cells).
Custom data binding/remapping is a separate capability; the original profile does not
establish it.

## 4. Rules (execution order)

**OWD-R01 Wiring** [verified in 1.13c code (static)]. Consume IEP-R03 with a supplied
stat-135/layer-0 change notification and the two event/selector fields above. A nonzero total,
including negative, creates one item-entry pair if absent: event 5 first, event 6 second,
therefore event 6 is first in the list. Both carry this stat/layer identity and select this
callback; callback value 2 is 0. Nonzero-to-nonzero leaves the pair unchanged; zero removes
both through IEP-R05. Several items share one pair and one current chance read. Automatic
equip/unequip notification production is IEP-Q01/IEP-Q03, not established by this supplied
notification rule. Apply placement consumes DPO-R13a/DPO-R23, not another item's callback.

**OWD-R02 Chance roll** [verified in 1.13c code (static)].

```
if A absent or D absent: return 0
chance = Stat(A, suppliedStatIdentity.id, suppliedStatIdentity.layer)
if chance <= 0: return 0
step(A.seed)                       // NRF-R01, store both words
roll = unsigned(A.seed.lo) mod 100
if roll >= chance: return 0         // signed comparison
continue to OWD-R03
```

There is exactly one attacker draw when `chance>0`, even for `chance>=100`. The callback
does not read the damage record or its second registered value. No early exit requests a
state application. Through item wiring, the supplied identity is stat 135, layer 0.

**OWD-R03 Level curve** [verified in 1.13c code (static)]. Read attacker stat 12 at layer 0;
clamp signed level to at least 1. All products/sums wrap i32. With slopes `{9,18,27,36,45}`:

```
f(L) = 0                                              if L <= 1
     = (L-1)*9                                        if L <= 15
     = (L-15)*18 + 14*9                               if L <= 30
     = 15*18 + (L-30)*27 + 14*9                       if L <= 45
     = 15*(18+27) + (L-45)*36 + 14*9                  if L <= 60
     = 15*(18+27+36) + (L-60)*45 + 14*9               otherwise
v = f(L)+40
```

No draw is made by the curve. Extreme supplied levels retain wrap; no universal positivity
or saturation is assumed beyond the nonoverflowing range.

**OWD-R04a Defender adjustment** [verified in 1.13c code (static)]. For a player, divide
`v` by 4, then by 2 only for event 6. For a monster, divide by 2 when either `specialA` or
`specialB` is true; no event test applies. Missing monster data gives false for both. Other
unit types leave `v` unchanged. Every division is signed, truncating toward zero; no draw.

**OWD-R04b Monster-condition producers** [unresolved]. Which monster creation paths set
`specialA` and `specialB` is OWD-Q01. These are supplied conditions with no inferred category
name or default for an existing monster. OWD-R04a's arithmetic is available with explicit inputs.

**OWD-R05 Timed state request** [verified in 1.13c code (static)]. After adjustment, request
`ApplyTimedState(source=A,target=D,skill=0,level=1,duration=200,stat=74,value=-v,state=62,
callback=default)` and return 1, ignoring the applier's return. Negation wraps i32. A rejected
application still returns 1 and retains the one consumed draw. No further direct draw.

**OWD-R06a Eligibility** [verified in 1.13c code (static)]. For a timed-state request, require
a target (absent target is fatal). Reject with null for signed state outside the definition
count or an absent state record. For a monster, reject invalid signed class, absent MonStats,
`npc=true`, a MonStatsEx link (signed i16) that is negative or at or above the extended
count, an absent selected MonStats2 record, or `isAtt=false`. Other unit types skip
the monster checks. Then test the state's supplied `curse` membership. In the recorded state-62
profile it is false, so continue directly to OWD-R07a. Rejection makes no state/list/event
request; the caller's draw and return 1 remain as OWD-R05.

**OWD-R06b Alternate curse branch** [asm-only]. A state-62 profile marked `curse` can reject
for curse resistance at least 100 or active state 57, adjust duration by
`duration-mulDiv(resistance,duration,100)` when resistance is nonzero, and select replacement
through the first primary list satisfying the supplied curse-list condition instead of state
lookup. The lookup's guarded search is read; the complete alternate-profile caller and its
application effects remain OWD-Q13 / OWD-Q12.
This branch is unreachable for the recorded original state-62 cells. It supplies no accepted
custom-profile application contract.

**OWD-R07a Refresh** [verified in 1.13c code (static)]. Look up the target state list using
RLI-R03. If absent, create through OWD-R08. If state or skill differs, go to replacement.
When the level equals the requested level, a nonzero duration requests new expiry
`low32(G.frame+duration)`, then another type-12 event for that expiry; return the same list.
The expiry setter follows STL-R06a: positive expiry sets timed, other values do not clear it.
Do not rewrite magnitude, owner, callback, skill, level or active-state membership. A lower
signed requested level returns null with no change; a higher level goes to replacement.
Two item procs have skill 0/level 1: the second refreshes 200 frames while retaining the first
magnitude and owner. Multiple expiry events may remain queued; freeing is OWD-Q09.

**OWD-R07b Replacement** [asm-only]. A state/skill mismatch or higher requested level
requests old-list detach, old-list free, then the creation sequence. The supplied-effect
recipe removes the old contribution and invokes its removal callback before creation;
actual detach/free/callback/storage effects remain OWD-Q09/OWD-Q12. No guessed list surgery
or guaranteed old-state cleanup is required by this gated rule.

**OWD-R08 Creation requests** [verified in 1.13c code (static)]. In order:

```
MarkForUpdate(D)
SetState(D,62,1)                               // RLI-R03, STL-R07
expiry=0; timed=false
if duration != 0:
    expiry=low32(G.frame+duration)
    Schedule(G,D,12,expiry)
    timed=true
owner=(source.type,source.id) if source present else (6,-1)
L=NewList(D.allocationTag,timed,expiry,owner,curse,exp,aura)
SetListState(L,62)
SetListSkill(L,0)
SetListLevel(L,1)
SetListStat(L,74,-v,0)
Attach(D,L,fullApplication)
SetListCallback(L,genericStatusCallback)
return L
```

For the original item request, expiry is frame+200, timed true and curse/exp/aura false.
Successful allocation supplies an initially empty list through STL-R06a. State/skill/level/
callback setters and the list-stat wrapper ignore null lists; the call order above is not
replaced by an early allocation-failure return. Actual attachment, contribution propagation,
stat writes and update marking remain OWD-Q12. Expiry routing/traversal consumes STL-R03/
STL-R04, with actual free and removal effects gated by OWD-Q09. The generic callback is
STL-R11; its state-clear/refresh requests do not establish the refresh dependencies' effects.

**OWD-R09a Monster life step** [verified in 1.13c code (static)]. On the monster event-3
handler, with dependency reads/effects explicitly supplied:

```
g=Stat(u,74,0)
if LifeStateAny(u): g=low32(g-BaseStat(u,74,0))   // OWD-R09b
if HasState(u,52) and g>=0: return
Schedule(G,u,3,low32(G.frame+1))
if g==0: CancelTick(u); return
life=Stat(u,6,0); max=MaxLife(u)
if g<0 and life<256:
    if room absent or isTown(room): return
life=low32(life+g)
if life>max: CancelTick(u); life=max
if life<1: life=0
SetUnitStat(u,6,life,0)
PercentUpdate(u)                                // OWD-R09d
if life!=0: return
if u.mode==0 or u.mode==12: return
continue to OWD-R10
```

Comparisons are signed. A schedule precedes the room guard and life write; no room/town
therefore skips the write while retaining the next tick. Current/max life are i32 raw 8.8;
there is no extra saturation before the signed comparisons. Actual stored stat effects are
OWD-Q12; complete tick restart/queue integration is OWD-Q05/OWD-Q09.

**OWD-R09b Life-state subtraction** [verified in 1.13c code (static)]. `LifeStateAny` returns
false for an absent unit/base, a base without state storage or absent active-state words.
Otherwise test intersection of the active-state words and supplied `life` mask over
`trunc(low32(stateCount+31)/32)` words, with signed division toward zero; a nonpositive word
count gives false. For positive word counts, both supplied arrays cover that many words.
Return true at the first nonzero intersection, else false. In the valid recorded count 185, six words are
tested. When true, OWD-R09a subtracts the own-array stat-74 read from the effective stat-74
read; with `fMin` false these are their unchanged stored values. It does not remove other
attached-source contributions or exclude the player's own regeneration in OWD-R11a.
Mask production/loaded correspondence remains a supplied dependency, not gameplay confirmation.

**OWD-R09c Cancel monster ticks** [verified in 1.13c code (static)]. Visit the unit's
pending frame-event chain, saving each successor. For every kind-3 entry, request removal:
an executing entry becomes pending removal; other matching entries become unavailable and
leave the queue/unit chain. Nonmatching entries remain. This removes the new next-frame
entry scheduled earlier in the same tick; the dispatcher later releases the executing one.
These are frame-event entries, distinct from STL-R11a's kind/key unit timers. No draw.

**OWD-R09d Life-percent update** [asm-only]. Compute old as the low byte of stat 352. For
a present unit, let `cur=Stat(u,6,0)>>8`, `m=MaxLife(u)>>8` (arithmetic shifts). If signed
`m>0` and `cur<m`, candidate `p` is the low byte of signed
`low32(cur<<7)/m`; otherwise `p=128`. An absent unit also supplies 128. If integer
`abs(old-p)>4`, request a life-percent notification carrying `p`, then stat 352 replacement
with `p`. The notification operation is unread (OWD-Q08); stored writes remain OWD-Q12.
Supplied notification/write effects allow arithmetic examples without identifying a message.

**OWD-R10 Bleed-death hand-off** [asm-only]. After a monster life step reaches 0 outside
modes 0/12, select a poison-state-2 list first, else an Open-Wounds-state-62 list. Request
clearing the unit's pending death auxiliary value to 0. With a selected list and valid
game/owner type 0..5, consume the supplied RegisteredUnitLookup result for its owner identity.
No-list tail entry, malformed owner types and missing game have no successful fallback specified here
(OWD-Q10). With active state 54, request state 92 on and return
without death/events. Otherwise invoke the death operation `(game,victim,killer,1)`, then
dispatch event 10 on victim/other killer with null record, then event 9 on killer/other
victim with null record when killer exists. No burning-state-115 list is selected. Complete death
execution remains OWD-Q10; IEP-R09a owns death-notification
placement with a supplied death outcome. No local draw is established on this readable slice.

**OWD-R11a Player life step** [verified in 1.13c code (static)]. The player event handler
always requests event 3 at `low32(frame+1)` first, forwarding its two incoming parameters.
Then skip all life/stamina/mana steps for an absent unit, `aliveExclusionCondition=true`, a
player in mode 0/17, a monster in mode 0/12 or another unit type. Otherwise run the life step,
then stamina, then mana. Only the life step is owned here:

```
g=Stat(u,74,0)
if g==0: return 1
life=low32(Stat(u,6,0)+g); max=MaxLife(u)
if life>max: life=max; RemoveHealthPot(u)        // OWD-R11b
if life<256: life=256
SetUnitStat(u,6,life,0)
PercentUpdate(u)
return 1
```

No life-state subtraction, preventheal check or town guard occurs in this life step. The
floor follows the maximum cap, so even a supplied maximum below 256 does not change the
floor. It makes no direct draw. First scheduling of a live player is OWD-Q05; this scheduled
handler's request/arithmetic does not depend on establishing that producer. Actual life
storage and subsequent percentage reads require OWD-Q12's effects.

**OWD-R11b Health-potion removal** [asm-only]. When the player life addition exceeds
maximum, seek state-100 `healthpot`; if found request detach and free before the life store.
No percentage update happens inside this removal request. Actual list/stat/callback effects
remain OWD-Q09/OWD-Q12; the one percentage step is after the life write in OWD-R11a.

**OWD-R12 Monster tick restart** [asm-only]. Consume CSA-R01's final lethal decision and
DPO-R19c. Positive effective life clears lethal; nonpositive sets it. For positive life,
nonzero total and monster defender: nonzero effective hpregen requests canceling its kind-3
frame events then scheduling kind 3 at frame+1 with three zero payload values; regardless
of hpregen, request per-hit monster update. Zero total/player/nonpositive life makes no such
request. The original per-hit update, scheduler mutation and integration remain DPO-Q24 /
OWD-Q05. A zero-damage proc does not obtain a restart from this stage; another scheduler
may exist. A supplied cancellation of an unrun current-frame tick skips that tick.

**OWD-R13 Bleed window** [asm-only]. With a newly created list at frame `F`, expiry `F+200`,
a restart for `F+1`, continuous tick scheduling and oldest-first equal-frame event order,
the candidate is bleed on `F+1..F+199`, with expiry before the `F+200` tick: 199 payments.
A refresh at `F'` gives the same candidate end `F'+199`, retaining prior payments. Actual
same-frame queue order, real attach/free/expiry and competing schedule/cancellation remain
OWD-Q09/STL-Q07. A 200-frame request alone does not verify 199 executed payments.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| OWD-R01 | Supplied item notification and dispatch placement | verified in 1.13c code (static) | IEP-Q01/IEP-Q03: automatic notification only | OWD-V12 |
| OWD-R02 | One attacker roll with nonpositive no-draw guard | verified in 1.13c code (static) | none for supplied current chance | OWD-V01 |
| OWD-R03 | Piecewise level curve | verified in 1.13c code (static) | none | OWD-V02 |
| OWD-R04a | Player/monster adjustment | verified in 1.13c code (static) | OWD-Q01: condition producers only | OWD-V03 |
| OWD-R04b | Monster-condition meaning/production | unresolved | OWD-Q01: production | none |
| OWD-R05 | Fixed timed-state request, ignore return | verified in 1.13c code (static) | OWD-Q12: stored list effects only | OWD-V01, OWD-V09 |
| OWD-R06a | State and monster eligibility | verified in 1.13c code (static) | none with supplied loaded records | OWD-V09 |
| OWD-R06b | Alternate curse branch | asm-only | OWD-Q13: lookup/application | none |
| OWD-R07a | Same-level refresh and weaker-request refusal | verified in 1.13c code (static) | OWD-Q09/OWD-Q12: actual expiry/list integration | OWD-V05, OWD-V10 |
| OWD-R07b | Replace requests | asm-only | OWD-Q09/OWD-Q12: detach/free | OWD-V10 |
| OWD-R08 | Ordered creation requests | verified in 1.13c code (static) | OWD-Q12: stored write/attach/update; OWD-Q09: expiry effects | OWD-V01 |
| OWD-R09a | Monster arithmetic/guards/requests | verified in 1.13c code (static) | OWD-Q12: stored write; OWD-Q05/OWD-Q09: event integration | OWD-V04, OWD-V06, OWD-V07, OWD-V13 |
| OWD-R09b | Life-state test and own-value subtraction | verified in 1.13c code (static) | none with supplied masks/arrays | OWD-V06 |
| OWD-R09c | Kind-3 frame-event cancellation | verified in 1.13c code (static) | none within supplied event chain | OWD-V13 |
| OWD-R09d | Candidate life-percent request/write | asm-only | OWD-Q08: notification; OWD-Q12: storage | OWD-V04 |
| OWD-R10 | Candidate bleed-death hand-off | asm-only | OWD-Q10: death effects, no-list/nonstandard inputs | OWD-V14 |
| OWD-R11a | Player event request and life arithmetic | verified in 1.13c code (static) | OWD-Q12: stored writes; OWD-Q08: notification | OWD-V08 |
| OWD-R11b | Candidate health-potion removal | asm-only | OWD-Q09/OWD-Q12: actual removal | none |
| OWD-R12 | Consumed monster restart | asm-only | OWD-Q05 / DPO-Q24: scheduler/update | OWD-V11 |
| OWD-R13 | Candidate 199-payment window | asm-only | OWD-Q09 / STL-Q07: queue/expiry integration | OWD-V04 candidate; OWD-O02 |

## 5. Outputs and side effects

| Output / request | Rule | Assertion boundary |
|---|---|---|
| Item event pair/dispatch calls | OWD-R01 | Supplied IEP notification/dispatcher |
| Attacker seed words; callback return 0/1 | OWD-R02, OWD-R05 | Direct local writes/return |
| Magnitude and timed-state request | OWD-R03, OWD-R04a, OWD-R05 | Direct local values/request |
| Eligibility null return or continuation | OWD-R06a | Supplied loaded state/monster definitions |
| New expiry; type-12 schedule request; existing-list return | OWD-R07a | Field/request; actual expiry/free gated |
| Detach/free requests; creation requests | OWD-R07b, OWD-R08 | Supplied effects; OWD-Q09/OWD-Q12 |
| Update, state 62 on, list owner/state/skill/level/stat/callback, full attach | OWD-R08 | Ordered requests; unresolved effects stay gated |
| Event 3 at next frame / cancel kind-3 entries | OWD-R09a, OWD-R09c, OWD-R11a, OWD-R12 | Schedule/cancel boundary above |
| Stat 6 layer-0 replacement request | OWD-R09a, OWD-R11a | Arithmetic exact; stored propagation OWD-Q12 |
| Life-state predicate and effective-minus-own hpregen | OWD-R09b | Supplied masks/arrays, no mutation |
| Candidate life-percent notification and stat 352 replacement | OWD-R09d | OWD-Q08/OWD-Q12 |
| Death auxiliary clear, owner lookup, state 92 on, death call, events 10/9 | OWD-R10 | Candidate hand-off; OWD-Q10 |
| Health-potion detach/free before cap/floor store | OWD-R11b | Candidate supplied effects |
| Separate stamina then mana call after player life | OWD-R11a | Only order; resource effects outside scope |
| Lethal result and monster per-hit update/restart | OWD-R12 | CSA-R01 / DPO-R19c; integration gated |
| Candidate list contribution removal/state-clear requests at expiry | OWD-R08, OWD-R13 | STL-R04/STL-R05/STL-R11; actual effects gated |

No unit creation, save write or packet format is defined here. Death's downstream effects and
actual network delivery require their own contracts.

## 6. Branches and edge cases

- Only a player value is event-dependent: quarter in melee, eighth through two truncating
  divisions for event 6. Hirelings use monster type; difficulty and player count are not inputs.
- Nonpositive chance makes no draw; chance at least 100 still consumes one. Missing attacker
  or defender returns 0. An ineligible monster retains the consumed draw and caller return 1.
- Level at most 1 gives 40. Extreme supplied levels and life additions use i32 wrap; valid
  ordinary game values do not establish a general no-overflow guarantee.
- Unknown monster conditions require explicit inputs. Missing monster data makes the tested
  conditions false; their wider category meaning is OWD-Q01.
- State/life-list existence and active state are separate. Supplied effective arrays may
  include regeneration from several sources; getters do not build those totals.
- An equal-level refresh retains first magnitude and owner. A lower request leaves the list
  unchanged. A higher request/skill mismatch uses the gated replacement path.
- Monster base regeneration is subtracted only when an active life-mask state is present.
  An attached positive source still offsets bleed. Player life never uses that subtraction.
- Monster preventheal stops this invocation before scheduling when net regeneration is
  nonnegative; it does not cancel independent preexisting future entries in that branch.
- With negative net regeneration and life below 256, town/no room skips the monster write.
  At or above 256 the loss can cross below one hit point before that guard applies next time.
- Zero monster net regeneration schedules then cancels kind 3; positive overflow above maximum
  also cancels. Exactly maximum without overflow does not take the `life>max` branch.
- Player zero regeneration reschedules without a life write; death/type/auxiliary guards
  also follow the initial reschedule. A player floor of 256 follows cap at maximum.
- Monster death attribution prefers poison, then Open Wounds, never burning. State 54 requests
  death delay instead of death/events. Actual death remains gated; malformed lookup inputs
  receive no inferred null-result fallback.
- State-62 `curse=false` avoids the alternate curse branch in the recorded profile; changing
  that cell does not establish the unread custom branch. Other-type creation requests do
  not supply a life consumer from the player/monster dispatches.

## 7. Open questions

| Id | Kind / status | Evidence that would settle it |
|---|---|---|
| **OWD-Q01** | outside-boundary / open | Trace which original monster creation paths set `specialA`/`specialB`. Their supplied-condition adjustment is verified; no category default is imposed |
| **OWD-Q05** | outside-boundary + integration / open | Establish other monster tick producers, first player scheduling and real restart integration. Local scheduled-handler arithmetic is available; DPO-Q24 retains per-hit update effects |
| **OWD-Q06** | observation / open | Observe combined Open Wounds/Crushing Blow/knockback registration and callback order on a hit. IEP-R01/IEP-R07 already fix order from a supplied list; no universal item-effect order is assumed |
| **OWD-Q08** | behavior + integration / open | Identify the life-percent notification and complete stat/state message delivery; OWD-R09d and actual delivery remain gated |
| **OWD-Q09** | integration / open | Complete real list detach/free/attach propagation and equal-frame queue order needed for expiry and 199 payments; STL-R04 traversal alone does not establish stored effects or order |
| **OWD-Q10** | integration / open | Complete death execution, no-list tail entry and nonstandard owner-input handling; compatible registered-unit lookup is read for valid game/type inputs, but owner preference/request order alone does not establish kill experience or death effects |
| **OWD-Q11** | outside-boundary / open | Determine whether another player site enforces preventheal for nonnegative regeneration; this life step has no such test |
| **OWD-Q12** | integration / open | Complete unit-write propagation (RLI-Q03), list write/attach/free (STL-R06c/STL-R06b/STL-R05) and RLI-Q02 update effects. Explicit supplied-effect arithmetic/request vectors remain usable |
| **OWD-Q13** | behavior / open | Complete the alternate-profile caller's request/branch verification in OWD-R06b; compatible first-matching primary-list lookup is read, but no original-profile state-62 curse execution is claimed |
| **OWD-Q14** | observation / pending | Reproduce OWD-O01..OWD-O05 on the evidence build and match the named assertions. No gameplay observation has been made |

All remaining gaps and dated attempts are in the linked
[gap file](../plans/gaps/combat-open-wounds.md). Gameplay observation is possible; a pending
observation does not gate separately verified static supplied-input assertions.

## 8. Test vectors

### 8.1 Executable vectors

Every recipe executes the cited local rules; its boundary is stated per family. No local rule
is replaced by a supplied implementation. External dependency requests are logged in order, one entry per requested operation;
requests a dependency operation makes internally (for example RLI-R03 SetState's own update
marking) are not logged as separate entries.
For explicitly supplied effects: allocation returns a fresh empty list; list setters persist
their named fields; state set stores active membership where storage exists; replacing writes
persist in own/selected effective arrays; full attach adds the supplied contribution to the
effective array; detach/free removes it and invokes a supplied callback when present;
scheduling records the named event (STL-R01) and inserts it at the head of the target
unit's pending-event chain (the original link step's order, recorded in the companion); the
dispatcher releases a running entry marked for removal once, after its handler returns; cancellation follows
OWD-R09c. Unit-stat writes update the selected array
before a later percentage read. Supplied notification/death/update operations only log calls.
These fixture effects are assertions of the recipe, not verified real dependency integrations.
Read definitions are present, stat-definition count 359, `fMin=false` unless stated, state count
185, and storage exists. Units use resolved effective arrays; their own arrays are supplied
separately. Their associated owner is the unit. Regeneration schedules have no override and
parameters 0,0 unless stated; the two incoming player-event parameters are 0,0.
Life-state masks use the named table profile. Existing effective/own values are explicitly
supplied where a recipe needs both.

Defaults: `G.frame=1000`; attacker player/type 0, id 1, chance 100, level 30, seed `{4,666}`;
event 5, registered identity `(135,0)`, second value 0. Defender is monster/type 1, mode 1,
class 0, MonStats count 1 with record 0 linked by MonStatsEx 0 to MonStats2 record 0/count 1,
`npc=false`, `isAtt=true`, monster data present,
`specialA=false`, `specialB=false`, no states/state-62 list; original state-62 record present
with curse/exp/aura false. Default callback is `genericStatusCallback`. Omitted resource
inputs are not read by the named bounded operation. Seeds in isolated nondrawing stages
are unchanged; `no draw` does not claim the upstream chance callback was executed.

The NRF-R01 step pairs used below are `{1,666}->{0x6AC6935F,0}` (mod 100 = 51),
`{3,666}->{0x4053B4E9,1}` (25), `{4,666}->{0xAB1A45AE,1}` (10),
`{5,666}->{0x15E0D673,2}` (99), and
`{0xAB1A45AE,1}->{0x62FD7EE7,0x475D97EE}` (63).

- **OWD-V01** Chance and request (OWD-R02, OWD-R03, OWD-R04a, OWD-R05, OWD-R08;
  boundary: supplied callback/application-request stage; isolated creation in (i))
  - Inputs: defaults, with the independent overrides below. Successful creation uses the
    preamble's explicit allocation/write/attach effects and default owner `(0,1)`.
  - Seeds / Draws: each case gives its entry seed and complete draw result.
  - Trace: chance guard, NRF-R01 step/remainder, level-30 curve f=396 then v=396+40=436, no monster
    halving, fixed request, then eligible creation requests in OWD-R08 order.
  - (a) Inputs: chance 0, seed `{1,666}`. Draws: none. Outputs: return 0, seed unchanged;
    no request, state change or event.
  - (b) Inputs: chance 25, seed `{4,666}`. Draws: attacker mod 100=10, final seed
    `{0xAB1A45AE,1}`. Outputs: request `(A,D,0,1,200,74,-436,62,default)`; logged order
    update, state 62 on, schedule 12 at 1200, allocation, list state/skill/level/stat, attach,
    callback; supplied list timed true, curse/exp/aura false, expiry 1200, owner `(0,1)`,
    state 62, skill 0/level 1, hpregen -436, generic callback; callback return 1.
  - (c) Inputs: chance 25, seed `{3,666}`. Draws: attacker mod 100=25, final
    `{0x4053B4E9,1}`. Outputs: equality fails, return 0; no application request.
  - (d) Inputs: chance -5, seed `{5,666}`. Draws: none. Outputs: return 0, seed unchanged,
    no application request.
  - (e) Inputs: chance 100, seed `{5,666}`. Draws: attacker mod 100=99, final
    `{0x15E0D673,2}`. Outputs: request `(A,D,0,1,200,74,-436,62,default)`, state 62 on,
    expiry event 12 at 1200, new timed list with owner `(0,1)`, state 62, skill 0/level 1,
    hpregen -436, generic callback, curse/exp/aura false; callback return 1.
  - (f) Inputs: defender absent, default seed `{4,666}`. Draws: none. Outputs: return 0,
    seed unchanged and no application request.
  - (g) Inputs: chance 100, seed `{5,666}`, present callback units, state-definition count 0.
    OWD-R06a rejects state 62 and returns null. Draws: attacker mod 100=99, final `{0x15E0D673,2}`.
    Outputs: request `(A,D,0,1,200,74,-436,62,default)`; no stored state/list/event;
    callback return 1, independently of null applier result.
  - (h) Inputs: chance 5, seed `{4,666}`. Draws: attacker unsigned mod 100=10, final
    `{0xAB1A45AE,1}`. Outputs: return 0 and no application request; the one draw is consumed.
    Treating that low word as signed before the remainder would give -86, incorrectly succeeding.
  - (i) Inputs: isolated OWD-R08 creation entry, deliberately bypassing the chance callback;
    source absent, target D/default original state-62 profile, duration 200, skill 0/level 1,
    stat 74/value -436 (v 436), frame 1000 and no prior list. Seeds: no draw; Draws: none.
    Trace: source fallback selects owner `(6,-1)`. Outputs: update, state 62 on, schedule
    event 12 at 1200, allocate a timed list with owner `(6,-1)` and curse/exp/aura false,
    set state 62/skill 0/level 1/stat 74=-436, full attach, generic callback, then return L.
    This creates the supplied list with expiry 1200; no callback return 0/1 is asserted here.
  - (j) Inputs: defender player/type 0, mode 1, event 6, chance 25, seed `{4,666}`. Draws:
    attacker mod 100=10, final `{0xAB1A45AE,1}`. Trace: v 436, player 436/4=109, event 6
    109/2=54. Outputs: request `(A,D,0,1,200,74,-54,62,default)`; creation as in (b) with
    hpregen -54, expiry 1200, owner `(0,1)`; callback return 1.
  - Discriminates: (c) rejects inclusive success; (e) rejects skipping a certain-chance draw;
    (g) rejects returning the applier's result. Case (b)'s log rejects attach-before-stat and
    callback-before-attach or list-allocation-before-expiry-scheduling request order; (j)
    rejects a callback that ignores its event argument or takes it from the second registered
    value (either gives -109); (h)
    rejects a signed remainder from the stepped low word; (i) rejects inventing another
    owner or dropping direct creation solely because its source is absent.
  - runtime_observed: no.

- **OWD-V02** Level bands (OWD-R03; boundary: isolated curve)
  - Inputs: independent levels -3,0,1,2,10,15,16,30,31,45,46,60,61,99.
  - Seeds: no draw; Draws: none.
  - Trace: level 30 gives `(30-15)*18+14*9=396`, +40=436; level 46 gives
    `15*18+15*27+36+126=837`, +40=877; level 99 gives
    `15*(18+27+36)+39*45+126=3096`, +40=3136.
  - Outputs: respectively v 40,40,40,49,121,166,184,436,463,841,877,1381,1426,3136;
    no state/list/seed mutation.
  - (a) Inputs: separate isolated curve entry, level 47721889. Seeds: no draw; Draws: none.
    Trace: `(47721889-60)*45=2147482305`; adding `15*(18+27+36)+14*9=1341` gives
    f=2147483646, then `low32(f+40)` is signed -2147483610. Outputs: v=-2147483610,
    no mutation. This is a supplied extreme input, not a claim that ordinary game levels reach it.
  - Discriminates: each band boundary and its first following level reject a constant-slope
    curve or inclusive-boundary shift; low levels reject extrapolating below 1. Case (a)
    rejects retaining the wider result 2147483686 instead of the required signed i32 result.
  - runtime_observed: no.

- **OWD-V03** Target adjustment (OWD-R04a; boundary: isolated adjustment)
  - Inputs: independent cases below; `specialA`/`specialB` false unless explicitly true.
  - Seeds: no draw; Draws: none.
  - Trace: signed truncation; for player event 6, divide by 4 before dividing by 2.
  - (a) Inputs: v 436, monster,event 5. Outputs:436.
  - (b) Inputs: v 436, monster,event 5,specialA alone true. Outputs:218.
  - (c) Inputs: v 436, monster,event 5,specialB alone true. Outputs:218.
  - (d) Inputs: v 436, monster,event 5, unrelated condition true and both specials false.
    Outputs:436.
  - (e) Inputs: v 436, monster,event 6. Outputs:436.
  - (f) Inputs: v 436, player,event 5. Outputs:109.
  - (g) Inputs: v 436, player,event 6. Outputs:109 then 54.
  - (h) Inputs: v 436, unit type 2,event 5. Outputs:436.
  - (i) Inputs: v 49, player,event 5. Outputs:12.
  - (j) Inputs: v 49, player,event 6. Outputs:12 then 6.
  - (k) Inputs: v 49, monster,event 5,specialA true. Outputs:24.
  - (l) Inputs: v 40, player,event 6. Outputs:10 then 5.
  - (m) Inputs: v -49, player,event 6. Outputs: -12 then -6, with signed truncation toward zero.
  - (n) Inputs: v 436, monster,event 5,specialA and specialB both true. Outputs:218.
  - Discriminates: (n) rejects halving once per true condition (109); (e) rejects halving all missile defenders; (b)/(c)/(d) distinguish either
    special condition from unrelated flags; odd positive values reject rounding divisions upward;
    (m) rejects floor division of a negative value (-13 then -7).
  - runtime_observed: no.

- **OWD-V04** Monster life and percentage (OWD-R09a, OWD-R09d;
  boundary: supplied monster life/percentage stage, plus separately gated window recipe)
  - Inputs: mode 1, room outside town, life 256000,max 256000, effective hpregen -436,
    own hpregen 0, active state 62 with attached -436 list, no state 52, old percent 128;
    first frame 1001. Keep these regeneration/state inputs for the repeated ticks.
  - Seeds: no draw; Draws: none.
  - Trace: g=-436; schedule event 3 for 1002 before first store; first life 255564 gives
    cur 998,m 1000,p 127 (`127744/1000`); difference 1 does not cross threshold 4.
  - Outputs: first stored life 255564, percent 128 unchanged, no notification. After 25 ticks,
    life 245100 (`256000-25*436`). On tick 14: life 249896, cur 976, p 124, difference
    exactly 4, no notification/write, percent remains 128. The only percentage write is
    on tick 19: life 247716,cur 967,p 123, difference 5, notification carrying 123 before
    stat 352=123. Final percent remains 123 after tick 25.
    Under the supplied 199-payment candidate only, loss 86764, exact remaining
    life 169236; loss is 338.921875 HP (338.92 when displayed to two decimal places).
  - Discriminates: tick 14's no-write at difference 4 and the only write at tick 19 reject a
    threshold of at least 4; stored raw life rejects rounding whole life before accumulation.
    The tick-19 percentage also distinguishes reading life before the store. The 199-payment recipe is OWD-R13
    candidate arithmetic, not proof of the original equal-frame dispatch order.
  - runtime_observed: no.

- **OWD-V05** Refresh retains magnitude/owner (OWD-R07a;
  boundary: two sequential supplied callback/application stages)
  - Inputs: first frame 1000, level 10, no old list; second frame 1100, level 30, explicit
    first-stage list state 62/skill 0/level 1, owner `(0,1)`, magnitude -121, expiry 1200.
  - Seeds: first `{4,666}`; second explicitly receives `{0xAB1A45AE,1}`.
  - Draws: first attacker mod 100=10 ->`{0xAB1A45AE,1}`; second mod 100=63 ->
    `{0x62FD7EE7,0x475D97EE}`.
  - Trace: first v 121 creates; second v 436 request has equal skill 0/level 1, so expiry only.
  - Outputs: first one list stat 74=-121, expiry 1200 and event 12 at 1200; second same list
    and owner, stat 74=-121, expiry 1300 and timed true, another event 12 at 1300; no allocation,
    state-set, stat-replacement or callback replacement on second proc. Both callbacks return 1.
  - (a) Inputs: separate timed-state applier entry at frame 1000, duration 0, skill 0/level 1,
    state 62/stat 74/value -436, source A/target D. Existing L has state 62/skill 0/level 1,
    hpregen -300, owner `(0,1)`, timed true, expiry 1100, generic callback; active state 62,
    queue contains exactly event 12 at 1100. Seeds: no draw; Draws: none. Trace: eligibility
    succeeds, same-level refresh skips the duration branch. Outputs: applier returns L,
    expiry 1100/timed true/hpregen -300/owner/callback unchanged, same single event 12 at 1100;
    no expiry-set, schedule, state-set, stat-write or allocation request. The item callback
    is not executed; it would supply duration 200 instead.
  - (b) Inputs: reversed sequence: first frame 1000, level 30, seed `{4,666}`, no old list;
    second frame 1100, level 10, explicit first-stage list state 62/skill 0/level 1, owner
    `(0,1)`, magnitude -436, expiry 1200, seed `{0xAB1A45AE,1}`. Draws: first mod 100=10,
    second mod 100=63. Trace: first v 436 creates; second v 121 has equal skill 0/level 1.
    Outputs: first list stat 74=-436, expiry 1200, event 12 at 1200; second same list and
    owner, stat 74=-436, expiry 1300, another event 12 at 1300; no allocation, state-set,
    stat-replacement or callback replacement. Both callbacks return 1.
  - Discriminates: stronger second magnitude must not replace the first or stack another list;
    (b) rejects keeping the stronger, weaker or latest magnitude instead of the first;
    (a) rejects unconditionally scheduling expiry or clearing/replacing it for duration 0.
  - runtime_observed: no.

- **OWD-V06** Own regeneration versus attached contributions (OWD-R09a, OWD-R09b, OWD-R11a;
  boundary: supplied masks/arrays and life requests)
  - Inputs: independently life 256000,max 256000,mode 1,outside-town room,no state 52,old percent 128.
    Monster cases have state storage and mask profile of section 3; fMin false. Each case's
    effective/own hpregen values are supplied existing array entries, not a getter-computed sum.
  - Seeds: no draw; Draws: none.
  - Trace: monster g=effective-own only with a life-mask intersection; player never subtracts.
  - (a) Inputs: monster,own+50,attached state 62 -436, effective -386, active 62.
    Outputs: LifeStateAny true, BaseStat 50, g=-436, life 255564.
  - (b) Inputs: monster,own+50,state 62 -436 and poison-state 2 -200, effective -586,active 2/62.
    Outputs: LifeStateAny true, BaseStat 50, g=-636, life 255364.
  - (c) Inputs: monster,own+50,burning-state 115 -100,effective -50,active 115 only.
    Outputs: LifeStateAny false,g=-50,life 255950.
  - (d) Inputs: monster,own 0,unflagged attached+50,state 62 -436,effective -386,active 62.
    Outputs: LifeStateAny true,BaseStat 0,g=-386,life 255614.
  - (e) Inputs: player,own+50,state 62 -436,effective -386,active 62,aliveExclusionCondition false.
    Outputs: no LifeStateAny subtraction,g=-386,life 255614.
  - Outputs: every case requests its next tick and its listed life write; no percentage
    notification/write. The explicit life numbers concretize the formula `256000+g`.
  - Discriminates: (a)/(d)/(e) reject subtracting all positive sources, retaining own monster
    regeneration or applying the monster subtraction to players; (c) rejects treating burning
    as a life-mask state. None validates real attached-list propagation.
  - runtime_observed: no.

- **OWD-V07** Monster town/no-room guard (OWD-R09a, OWD-R10;
  boundary: supplied monster life stage and candidate death request)
  - Inputs: mode 1,g=-109 (effective -109,own 0,active 62/list -109),max 256000,old percent 0,
    initial life 300/frame 1001, state 62 owner `(0,1)`, no poison/state 54. Supplied lookup finds
    owner `(0,1)`; actual death effects are gated. Keep life written by the preceding tick.
  - Seeds: no draw; Draws: none.
  - Trace: life 300>=256 bypasses room guard, stores 191,p 0. On next tick 191<256 checks room.
  - (a) Inputs: town room throughout. Outputs: schedule each next frame; life 191 then stays 191;
    no later life/percent write, no death request.
  - (b) Inputs: first tick outside town, then no room. Outputs: same first store 191; later ticks
    reschedule without writing life or death request.
  - (c) Inputs: outside-town room throughout. Outputs: life 191,82,0 (third raw sum -27 floors 0);
    no percent write. Candidate death auxiliary 0, death request with found owner, event 10 then 9.
  - (d) Inputs: separate entry, town room, life exactly 256, g=-109. Trace: 256<256 is false,
    so the guard is bypassed. Outputs: schedule, store 147, percent `(0*128)/1000=0` unchanged
    (old 0), no notification/write and no death request.
  - (e) Inputs: separate entry, town room, life 191, effective hpregen +109, own 0, no active
    states. Trace: g>=0 with no state 52 schedules; the guard needs g<0. Outputs: schedule,
    store 300, percent 0 unchanged, no notification/write.
  - Discriminates: (d) rejects a `life<=256` guard; (e) rejects applying the town guard for
    nonnegative g. Town guard is before negative addition only when old life<256; it does not
    clamp every monster to 256 or skip the initial loss. Room rejection retains the new schedule.
  - runtime_observed: no.

- **OWD-V08** Player floor and scheduled guards (OWD-R11a;
  boundary: supplied player event/life stage)
  - Inputs: player mode 1, aliveExclusionCondition false,life 600,max 25600,effective hpregen -109,
    own 0,active 62/list -109,old percent 1; successive frames. Stamina/mana only log their calls.
  - Seeds: no draw; Draws: none.
  - Trace: life 491,382,273, then raw 164 floors 256; next raw 147 floors 256. Each life-step
    percentage calculation gives 1 (cur 1,m 100,128/100=1).
  - Outputs: initial event 3 reschedule on every invocation; successive life stores
    491,382,273,256,256, with every later continuing negative tick storing 256; no percent write.
    Life helper returns 1, then stamina and mana are called in that order.
  - (a) Inputs: separate entry, player mode 1, life 256, g 0, old percent 1. Outputs: reschedule,
    no life/percent write, life-helper return 1, then stamina/mana.
  - (b) Inputs: separate entry, player mode 17,life 600,g -109. Outputs: reschedule before the
    guard; no life/stamina/mana call or life/percent write.
  - (c) Inputs: separate player entry, mode 1, life 600, maximum 25600, effective hpregen 109,
    own 0, active state 52, town room, old percent 1, aliveExclusionCondition false. Outputs:
    reschedule, store 709, percentage 2 differs by 1 so no notification/write; life helper returns 1,
    then stamina/mana. No preventheal rejection occurs.
  - (d) Inputs: separate player entry, mode 1, life 191, maximum 25600, effective hpregen -109,
    own 0, active 62, town room, old percent 1, aliveExclusionCondition false. Outputs:
    reschedule, raw sum 82 then floor 256, store 256, percent 1 unchanged; no town rejection.
  - (e) Inputs: separate player entry, mode 1, life 600, maximum 25600, effective hpregen -109,
    old percent 1, aliveExclusionCondition true. Outputs: event 3 reschedule before the guard;
    no life/stamina/mana call and no life/percent write.
  - (f) Inputs: separate player entry, mode 1, aliveExclusionCondition false, life 100,
    maximum 128, effective/own hpregen 200, old percent 128, active state 100 only;
    a primary healthpot list HP with state 100, no stat entries and no callback. Supplied
    detach removes HP; supplied free releases it without further calls. Draws: none.
    Trace: sum 300 exceeds maximum 128; cap to 128, request healthpot detach/free through
    OWD-R11b, then floor to 256. Outputs: reschedule; ordered detach(HP),free(HP),life write 256;
    percentage 128 because maximum>>8 is 0, no notification/percent write; life helper returns 1,
    then stamina/mana. Stored removal and notification remain their existing gates.
  - (g) Inputs: isolated player-event handler entry with a present unit of type 2, mode 1,
    aliveExclusionCondition false, frame 1000 and no pending events; parameters 0,0.
    Seeds: no draw; Draws: none. Outputs: first request event 3 for this unit at 1001,
    no override/parameters 0,0; then return before life, stamina or mana calls and without
    life/percent writes. This is an isolated handler guard, not original type-2 dispatch.
  - (h) Inputs: separate player entry, mode 1, aliveExclusionCondition false, life 25500,
    maximum 25600, effective/own hpregen 100, old percent 128, active state 100 only; a primary
    healthpot list HP with state 100, no stat entries and no callback. Draws: none. Trace: sum
    25600 is not above the maximum. Outputs: reschedule; no detach/free request; life write
    25600; percentage 128 (cur 100 is not below m 100), no notification/percent write; life
    helper returns 1, then stamina/mana.
  - Discriminates: rejects monster floor 0, stopping rescheduling at g 0/dead mode, applying
    the monster preventheal gate in (c), applying its below-one-HP town guard in (d), or
    ignoring the independent aliveExclusionCondition guard in (e). Case (f) rejects floor-before-cap,
    which would store 128, and observes removal requests before the final life store.
    Case (g) rejects running resource steps for other unit types after rescheduling.
    Case (h) rejects removing the healthpot list when the sum only reaches the maximum.
  - runtime_observed: no.

- **OWD-V09** Eligibility does not change callback success (OWD-R06a, OWD-R05;
  boundary: supplied callback/eligibility stage)
  - Inputs: independent default callback entries, original state 62 data; overrides below.
  - Seeds: `{4,666}` per case; Draws: one attacker mod 100=10, final `{0xAB1A45AE,1}` per case.
  - Trace: v 436 for monster or 109 for player; callback requests application and ignores null.
  - (a) Inputs: npc true. Outputs: application null, no state/event/list, callback return 1.
  - (b) Inputs: npc false,isAtt false. Outputs: application null,no state/event/list,callback 1.
  - (c) Inputs: npc false,isAtt true and valid linked records. Outputs: application list,
    state 62 on,event 12 at 1200,list owner `(0,1)`,expiry 1200,skill 0/level 1,hpregen -436,
    timed true,curse/exp/aura false,generic callback; callback 1.
  - (d) Inputs: MonStatsEx=-1. Outputs: unavailable extended record,application null,
    no state/event/list; callback 1.
  - (e) Inputs: defender player. Outputs: skip monster records, application list with
    state 62,hpregen -109,expiry 1200,owner `(0,1)`,skill 0/level 1,timed true,curse/exp/aura false,
    generic callback; state-on and event 12 at 1200 requested; callback 1.
  - (f) Inputs: npc false, MonStatsEx=1, extended count 1 (record 0 has isAtt true). Outputs:
    link at the count is unavailable, application null, no state/event/list; callback 1.
  - (g) Inputs: npc false, MonStatsEx=0, extended count 1, selected MonStats2 record absent.
    Outputs: application null, no state/event/list; callback 1.
  - Discriminates: rejected application still consumes the chance draw and returns callback 1;
    player rejection must not depend on absent monster records. (f) rejects testing only a
    negative link (or `>` instead of `>=` the count); (g) rejects reading `isAtt` without the
    record-presence test.
  - runtime_observed: no.

- **OWD-V10** Weaker level and replacement (OWD-R07a, OWD-R07b;
  boundary: supplied callback/application stage; replacement stored effects remain gated)
  - Inputs: independent default callbacks; existing list state 62,expiry 1100,hpregen -300,
    owner `(0,1)`, with skill/level overrides below. Default v 436 request is skill 0/level 1.
  - Seeds: `{4,666}` per case; Draws: one attacker mod 100=10, final `{0xAB1A45AE,1}` per case.
    The applier portion alone makes no draw.
  - Trace: compare state/skill before level. Replacement requests detach/free before creation.
  - (a) Inputs: list skill 0,level 3. Outputs: applier null; old list expiry 1100,hpregen -300
    unchanged; no event/create request; callback 1.
  - (b) Inputs: list skill 0,level 0. Outputs: detach/free old list, then update/state-on,
    schedule 12 at 1200,new timed list expiry 1200,hpregen -436,state 62,skill 0/level 1,owner `(0,1)`,
    generic callback,curse/exp/aura false; callback 1. Supplied removal callback precedes creation.
  - (c) Inputs: list skill 5,level 1. Outputs: same independent replacement request sequence,
    new timed list expiry 1200,hpregen -436,state 62,skill 0/level 1,owner `(0,1)`,generic callback,
    curse/exp/aura false; callback 1.
  - Requests made inside the supplied removal callback, and whether the old list's event 12
    at 1100 stays queued, are not asserted (OWD-R07b; OWD-Q09/OWD-Q12). The asserted log is
    detach, free, then the OWD-R08 creation sequence.
  - Discriminates: equal requested/list level does not prevent replacement for skill mismatch;
    lower requested level does not refresh expiry. Actual old-contribution removal remains gated.
  - runtime_observed: no.

- **OWD-V11** Consumed hit restart (OWD-R12;
  boundary: supplied apply-finalization requests; real scheduler/update integration gated)
  - Inputs: independent reached finalization entries at frame 1000; default monster life 200000,
    total 4096,effective hpregen -436; explicit pending kind 3 frame 1000 and kind 12 frame 1200;
    each case starts lethal=true except (e), which starts false.
  - Seeds: no draw in this stage; Draws: none.
  - Trace: final life determines lethal; only living monster/nonzero total enters update.
  - (a) Inputs: defaults. Outputs: lethal false,remove all kind 3,kind 12 retained,
    event 3 at 1001 with three zero payload values, then per-hit update request.
  - (b) Inputs: total 0. Outputs: lethal false,no cancellation/schedule/update request.
  - (c) Inputs: player defender. Outputs: lethal false,no cancellation/schedule/update request.
  - (d) Inputs: hpregen 0. Outputs: lethal false,no cancellation/schedule; per-hit update requested.
  - (e) Inputs: life 0. Outputs: lethal true,no cancellation/schedule/update request.
  - (f) Inputs: effective hpregen +50. Outputs: as (a): lethal false, remove all kind 3,
    kind 12 retained, event 3 at 1001 with three zero payload values, then per-hit update.
  - (g) Inputs: total -4096. Outputs: as (a).
  - Discriminates: (f) rejects restarting only for negative hpregen; (g) rejects testing
    total>0 instead of total!=0. Hpregen 0 skips scheduling but does not skip per-hit update; zero total does.
    Supplied removal of the unrun frame 1000 event rejects retaining a same-frame payment.
  - runtime_observed: no.

- **OWD-V12** Supplied registration/dispatch (OWD-R01;
  boundary: IEP notification and dispatcher, not equip-notification production)
  - Inputs: present unit u,empty event list,stat 135/layer 0 change 0->25; supplied definition
    event 5/selector 15 and event 6/selector 15. Dispatcher receives game,event,owner u,other v,
    record R and both registered values; supply a logging callback for dispatch identity only.
  - Seeds: no draw in registration or supplied logger; Draws: none.
  - Trace: create 5 then 6 through head insertion; nonzero-to-nonzero retains pair; zero removes.
  - Outputs: head first event 6 then 5, both item kind,owner identity `(135,0)`,values
    `((135,0),0)`,Open Wounds selector. Dispatch 5 logs one call with owner u,other v,record R,
    identity `(135,0)`,second value=0; Dispatch 7 logs none. Next notification 25->60 adds no entry;
    effective stat read supplied as 60 supplies the next callback's chance. Notification 60->0
    removes both. Dispatcher result is the logger's supplied 0; no Open Wounds roll is executed
    by this registration/argument-only logger fixture.
  - Discriminates: rejects appending rather than head insertion, duplicating entries at 60
    or dispatching mismatched events.
  - runtime_observed: no.

- **OWD-V13** Cancel versus return-before-schedule (OWD-R09a, OWD-R09c;
  boundary: supplied monster frame-event chain)
  - Inputs: independent frame 1001 entries,monster mode 1,pending running kind 3 node N0 frame 1001
    and kind 12 E frame 1200, no other entries; state storage present,own hpregen 0,active 62,
    no state 52 except (c)/(d),outside-town room,old percent 128.
  - Seeds: no draw; Draws: none.
  - Trace: next kind 3 N1 frame 1002 is added before the g 0/cap cancellation, but preventheal
    with nonnegative g returns before scheduling. The running N0 is deferred until return.
  - (a) Inputs: effective hpregen 0. Outputs: schedule N1,then remove N1,mark N0 pending removal,
    retain E,no life write; after dispatcher releases N0,no pending kind 3.
  - (b) Inputs: life 255900,max 256000,effective hpregen 136 (attached+436 and state 62-300).
    Outputs: schedule N1,sum 256036>max,remove N1,mark N0,retain E,store 256000; tick stops
    after dispatcher releases N0. Old percent 128 remains 128,no notification.
  - (c) Inputs: life 255900,max 256000,effective hpregen 136,state 52 active.
    Outputs: return without schedule or life/percent write; no explicit CancelTick call;
    dispatcher releases running N0, retains E, no new kind 3 was created.
  - (d) Inputs: effective hpregen 0, state 52 active; initial pending chain `[P,N0,E]`,
    with an additional nonrunning kind-3 P at frame 1100. Trace: g=0 satisfies the nonnegative
    preventheal guard. Outputs: no schedule, CancelTick call, life read/write or percent update;
    dispatcher releases running N0 and retains both P and E. This is a return before the
    zero-g branch; it does not cancel the independent future P.
  - (e) Inputs: effective hpregen 0, state 52 absent; initial pending chain is
    `[P,N0,E]`, where P is another nonrunning kind-3 event at frame 1100, N0 is the running
    kind-3 event at 1001 and E is the unrelated kind-12 event at 1200. Trace: schedule N1
    kind 3/frame 1002 at the head, then cancel visits `[N1,P,N0,E]`. Outputs: remove N1 and
    P, mark N0 pending removal, retain E; no life write. After dispatcher releases N0,
    E is the only pending event. No existing future kind-3 event survives cancellation.
  - (f) Inputs: no active states, effective hpregen 256, own hpregen 0, life 128000,
    maximum 256000, old percent 64; original pending chain `[N0,E]`. Draws: none.
    Trace: g=256, schedule N1 at 1002, life 128256 is below maximum; percent is
    `((128256>>8)*128)/1000 = 501*128/1000 = 64`. Outputs: no CancelTick request,
    life write 128256, percent 64 unchanged/no notification; after dispatcher releases
    running N0, pending chain is exactly `[N1,E]`, frames 1002 and 1200. The next tick survives.
  - (g) Inputs: no active states, effective hpregen 100, own hpregen 0, life 255900,
    maximum 256000, old percent 128; original pending chain `[N0,E]`. Draws: none.
    Trace: g=100, schedule N1 at 1002, sum 256000 is not above the maximum. Outputs: no
    CancelTick request, life write 256000, percent 128 (cur 1000 is not below m 1000)
    unchanged/no notification; after dispatcher releases N0, pending chain is exactly `[N1,E]`.
  - Discriminates: (g) rejects a `>=` maximum cancellation test. Matching cancellation removes even the next event scheduled in this call;
    preventheal is a prior return, not schedule-then-cancel or removal of unrelated expiry.
    Case (d) rejects testing g>0 instead of g>=0, including a wrong final queue that loses P;
    (e) rejects canceling only the running/new events; (f) rejects canceling positive
    regeneration before reaching maximum or skipping its life addition.
  - runtime_observed: no.

- **OWD-V14** Candidate bleed-death requests (OWD-R10;
  boundary: supplied lookup/death/dispatch operations, not actual death effects)
  - Inputs: independent reached monster tick tail after storing life 0,mode 1,no state 54,
    no poison list,state 62 list owner `(0,1)`; supplied registered lookup finds `(0,1)`.
  - Seeds: no local draw on this bounded tail; Draws: none from supplied operations.
  - Trace: poison-list preference before Open Wounds, clear auxiliary, lookup, state 54 test,
    death request, victim event 10, optional killer event 9.
  - (a) Inputs: defaults. Outputs: auxiliary 0,killer `(0,1)`,death `(G,u,killer,1)`,
    event 10 owner u/other killer then 9 owner killer/other u,both record null.
  - (b) Inputs: poison list owner `(1,77)` also present and lookup finds that unit.
    Outputs: auxiliary 0,killer `(1,77)`,death `(G,u,killer,1)`,then 10 and 9 as above with this killer.
  - (c) Inputs: state 54 active. Outputs: auxiliary 0,lookup finds `(0,1)`,state 92 on;
    no death request or event 10/9.
  - (d) Inputs: mode 12 at the prior OWD-R09a tail. Outputs: no auxiliary clear,lookup,
    death request,state 92 request or event 10/9 after the already supplied life store.
  - (e) Inputs: owner not registered,lookup null. Outputs: auxiliary 0,death `(G,u,null,1)`,
    event 10 owner u/other null/record null; no event 9.
  - (f) Inputs: poison-state 2 list owner `(1,77)` only, no state-62 list; lookup finds that
    unit. Outputs: auxiliary 0, killer `(1,77)`, death `(G,u,killer,1)`, then 10 and 9 with
    this killer.
  - (g) Inputs: burning-state 115 list owner `(1,77)` only, no state-2 or state-62 list.
    Outputs: no registered-unit lookup for `(1,77)` and no killer `(1,77)`; the remaining
    no-list tail is OWD-Q10 and not asserted.
  - Discriminates: (f) rejects requiring a state-62 list; (g) rejects falling back to the
    burning list. Poison owns credit when both lists exist; state 54 suppresses death/events,
    null killer suppresses only event 9, and already-dead mode suppresses the entire hand-off.
  - runtime_observed: no.

### 8.2 Observation vectors

None has been performed. These are proposed observations, with detailed original-function
locations and reproducible procedures in the
[companion](../research/spec-evidence/combat-open-wounds.md#proposed-observations).

- **OWD-O01**: match one attacker draw/callback on a melee and missile hit, event 5/6,
  registered stat identity and return 1; distinguish placement before leech.
- **OWD-O02**: measure list creation/refresh, tick start, natural-regeneration exclusion and
  payment frames through expiry, including a zero-damage proc. The 199-payment prediction is gated.
- **OWD-O03**: test a hittable ineligible monster for rejection after one draw and callback return 1;
  if no eligible gameplay setup exists, record that limit without claiming a match.
- **OWD-O04**: observe a bleed death, owner preference and downstream kill credit; current
  experience/death effects are not established by a supplied hand-off fixture.
- **OWD-O05**: observe a hostile player defender losing life to the 256 raw floor without dying.

## 9. Acceptance and dependencies

Partial acceptance covers the verified local assertions in OWD-R01, OWD-R02, OWD-R03,
OWD-R04a, OWD-R05, OWD-R06a, OWD-R07a, OWD-R08, OWD-R09a, OWD-R09b, OWD-R09c and
OWD-R11a with the executable recipe boundaries above. Passing supplied-effect/candidate
recipes does not establish actual write/attach/free/notification/death/scheduler integration.
Full reproduction additionally requires the gated assertions in the Id map to be resolved
and attributable matching observations; it cannot be inferred from 14 recipe families alone.

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md)
NRF-R01, NRF-R05, NRF-R08a and NRF-R09a for numeric operations;
[core-record-layouts-and-ids](core-record-layouts-and-ids.md) RLI-R03/RLI-R02a/RLI-R02b for
state and stat operations; IEP-R01/IEP-R02/IEP-R03/IEP-R05/IEP-R07 for supplied registration
and dispatch, with IEP-Q01/IEP-Q03 notification limits; DPO-R13a/DPO-R23 for proc placement,
DPO-R19c/DPO-Q24 and CSA-R01 for final lethal/restart; STL-R01/STL-R03/STL-R04/STL-R06a/
STL-R07/STL-R11 for schedule, traversal, primitives, state and removal requests, preserving
STL-R05/STL-R06b/STL-R06c/STL-Q07, unit-write RLI-Q03 and update RLI-Q02 gates.
[core-game-data-table-records](core-game-data-table-records.md) GDA-R08a/GDA-R09a/GDA-R12/
GDA-R13/GDA-R16 supply named fields and flag-test results and [core-game-data-fields](core-game-data-fields.md)
GDF-R02/GDF-R03 stat and state identity conventions (id equals row index); the specific
stat and state ids used here are this spec's section 3 identities with their named-cell records. Explicit room/
town inputs use the PVP-R06 condition's supplied result without asserting its unread internals.

**Supporting references:** SCD-R06/SCD-R09 and MMC-R01/MMC-R03/MMC-R14 identify original hit
roots; they add no requirement to reproduce those whole systems for a supplied proc callback.
The source manifest and named-cell records support evidence identity, not runtime parity.

**Depended on by:** DPO and IEP delegate the Open Wounds callback here; CSA poison/burn share
the monster life consumer and candidate attribution. NRF's callback draw-site inventory cites
this effect. Reverse references do not create prerequisites to local arithmetic.

## 10. Related specs

- [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md): registration,
  selector lookup, traversal and death notification placement.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): apply placement and restart.
- [combat-status-application](combat-status-application.md): poison/burn life contributions.
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): lifecycle and removal requests.
- [combat-crushing-blow](combat-crushing-blow.md): separate effect sharing item-event triggers.
- [combat-knockback](combat-knockback.md): separate effect whose callback order follows registration.
- [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md): town and attribution.
- [skill-cast-start-and-do-dispatch](skill-cast-start-and-do-dispatch.md) and
  [missile-movement-and-collision](missile-movement-and-collision.md): hit dispatch roots.
- [core-game-data-fields](core-game-data-fields.md): data names and identities.
- [core-game-data-table-records](core-game-data-table-records.md): loaded record fields and flag tests.
- [core-record-layouts-and-ids](core-record-layouts-and-ids.md): stat reads and state-list operations.

## Appendix A. Provenance

Evidence build is original Lord of Destruction Expansion 1.13c. D2Game/D2Common static readings
dated 2026-10-13 support the bounded callback, applier and life steps; targeted helper reading
and original-image instruction corroboration dated 2026-10-15 support the life-state test and
getter boundary. Named cells reuse complete effective loader reads dated 2026-10-15; extraction
does not renew the read or prove compiled-data correspondence. No gameplay observation has
been performed. Public original-module/function references, data limits, recorded correction rationale
and proposed procedures are in the [companion](../research/spec-evidence/combat-open-wounds.md),
using artifact identities from the [source manifest](../research/source-manifest.md).
