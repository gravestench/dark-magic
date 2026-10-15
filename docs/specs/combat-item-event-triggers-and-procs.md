# Item Event Triggers and Procs

**Description**

Some item bonuses trigger extra effects when you attack, take damage, or kill an enemy.
The game checks the effects registered for that moment and runs matching effects in order.
Individual effects determine their targets, resource changes, and chance rolls.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: IEP-R, IEP-Q, IEP-V, IEP-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-item-event-triggers-and-procs.md)
Open evidence gaps: [gap file](../plans/gaps/combat-item-event-triggers-and-procs.md)

## 1. Scope

The generic unit-event mechanism used by item effects: registration from stat definitions,
entry ordering and removal, callback selection, dispatch sites, and the generic callbacks for
physical and lightning retaliation, damage to mana, life and mana after a kill, and Rest in
Peace. The catalog identifies all 31 populated selectors without specifying unread callbacks.
Crushing Blow, Open Wounds and knockback own their effects in their separate specs.
Skill-specific registrations, chance-to-cast execution, reanimation spawning, freeze, slow,
flee, and stat contribution arithmetic have separate ownership. The notification integration
boundary needed to register item effects remains explicit here.

## 2. Trigger and ownership

D2Game's server unit-event machinery computes these rules. A stat-total change reaches item
registration; an event dispatch reaches registered callbacks through the same selector mapping.
Combat dispatches occur during damage building, mitigation, application and melee completion.
Other roots are regeneration death, level change, missile contact, and a monster entering its
dead mode. Each dispatch runs when its caller reaches that point; this is not a periodic scan
of equipped items. IEP-R09a states relative order; the companion records every binary link.

The client sees downstream damage, resource, state and overlay updates. Their messages belong
to their downstream owners. Producer eligibility and multi-stat notification order remain
IEP-Q01 and IEP-Q03.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| Game, owner, other | Supplied event context | Logical references; units may be absent | Owner supplies event entries; other is callback-specific |
| Requested event | Dispatch caller | i32 | Compared with entry u8 event without narrowing request |
| Entry | Owner's ordered event collection | Event u8; kind, owner identity, callback values/results 32-bit | Kind is item, skill or one-shot |
| Stat identity | Stat id and layer | Each u16 | Distinct layers register separately |
| Stat definition | `ItemStatCost.itemevent1/2`, `itemeventfunc1/2`, `fCallback` | Events/selectors signed i16; callback flag boolean | Definition fields supplied through GDA-R12 |
| Totals | Old/new total; callback's current owner stat | Signed i32 | Existing-value reads consume RLI-R02a/RLI-R02b |
| Resources | Stats 6 `hitpoints`, 8 `mana`; effective maximum life/mana | Signed i32 raw 8.8 | Current and supplied effective maxima |
| Damage record | Shared logical damage record, RLI-R01 | Signed i32 components/total, named conditions | Same record visible to successive callbacks |
| Retaliation eligibility | Other's `retaliationEligible` | Boolean | Neutral condition; producer meaning IEP-Q02 |
| Demon eligibility | Victim demon classification, HIT-R02g | Boolean | Selector 18 only |
| Reanimation eligibility | Monster victim; `reanimationExcluded`; owner chain/player id | Boolean, units and identity | Registration slice IEP-R13b only |
| Seed | Event owner's seed | `{lo,hi}` u32 pair | Direct reanimation chance draw |
| State request environment | RLI-R03 state-definition count and target state storage | i32 count; optional logical base list/storage | Consumed state-request routing; actual update bookkeeping remains gated |
| Overlay request environment | LSE-R13 inputs: supplied `overlayCount`, owner's stat list and attached-list chain | i32 count; LSE section 3 list fields | Consumed overlay list search/creation and ordered requests; stored effects IEP-Q09; original count identity LSE-Q03 |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Selector range 0 through 49; populated 1 through 31 | Engine code/table | Fixed by original code |
| Item events and selectors | Table data | `data/global/excel/ItemStatCost.txt`, row `Stat`, columns named above, select engine-defined handlers; exact catalog cells IEP-Q08 |
| Event identities/names | Table data | `data/global/excel/events.txt`, zero-based rows excluding header; code fixes dispatched numbers separately |
| Skill event selectors | Table data | `data/global/excel/skills.txt`, `auraevent1..3`, `auraeventfunc1..3`, `passiveevent`, `passiveeventfunc`; exact skill bindings outside effect contract |
| One-shot event 13 | Engine code | Fixed code identity; absence of table row remains catalog evidence gate |
| State 172 `restinpeace` | Engine code use; table name | Code-fixed state selection; name correspondence consumes GDF-R03 |
| Overlay requests 151, 152 | Engine code | Fixed callback request ids; operation LSE-R13, stored effects IEP-Q09 |
| Resource conversion 256, percentage divisor 100 | Engine code | Fixed arithmetic; NRF-R08a/NRF-R05 |

Supplying a stat definition or entry does not establish the original-profile table binding.

## 4. Rules (execution order)

**IEP-R01 Entry creation** [verified in 1.13c code (static)]. A non-null owner gains the
supplied event, callback, kind, owner identity and two callback values at the head of its list.
A null owner gains nothing. The newest registration is visited first. No direct draw.

**IEP-R02 Selector lookup** [verified in 1.13c code (static)]. A selector less than zero,
at least 50, or selecting an empty callback creates no entry. Otherwise create through IEP-R01.
Selectors 0 and 32 through 49 are empty; 1 through 31 are populated. No direct draw.

**IEP-R03 Item registration** [verified in 1.13c code (static)]. With a supplied notification
and valid stat definition, do nothing unless the first signed event is positive. Otherwise:

```
if newTotal != 0:
    if an item entry already matches this stat identity: leave entries unchanged
    else:
        create first event through its selector, item kind, stat identity, values (stat identity, 0)
        if second event > 0:
            create second event with its selector and the same kind, identity and values
else:
    remove all item entries for this stat identity through IEP-R05
```

The identity includes layer. Several items contributing the same stat/layer share one
registration, whose callback reads the total. Nonzero-to-nonzero changes preserve entries;
negative totals register. A failed first-selector lookup does not prevent a valid second
event. A nonpositive first event prevents both. No direct draw. Notification production and
order are IEP-Q01/IEP-Q03; the supplied-notification operation is independently verified.

**IEP-R04 Registration notification integration** [unresolved]. Unit initialization supplies
item registration as its base stat list's change callback. The complete set of unit producers
and `fCallback` notification traversal/order are IEP-Q01/IEP-Q03. Automatic equip/unequip
registration requires those boundaries; IEP-R03 accepts explicit notifications independently.

**IEP-R05 Removal** [verified in 1.13c code (static)]. Remove every entry matching requested
kind and owner identity. A currently executing match remains until its callback returns and
is marked for removal then; other matches become unavailable immediately. Both events for an
item stat identity are removed together. No direct draw.

**IEP-R06 Shared callback selection** [verified in 1.13c code (static)]. Item and skill
registration resolve the same selector to the same callback. Item entries have item kind and stat identity; skill
entries have skill kind and the skill's owner identity. Skill events/selectors come from the
corresponding skill fields. Removing skill entries uses IEP-R05 with skill kind. This does
not specify those skills' effects or notification timing. No direct draw.

**IEP-R07 Dispatcher** [verified in 1.13c code (static)].

```
if owner absent: return 0
result = 0
entry = owner's first entry
while entry exists:
    successor = entry.next
    if entry.event == requestedEvent:
        mark entry executing
        result = callback(game, requestedEvent, owner, other, record, valueA, valueB)
        clear executing mark
        successor = entry.next
        if removal requested or entry is one-shot: remove entry
    entry = successor
return result
```

Nonmatches advance using the pre-comparison successor. Matches reload the successor after
the callback; removing the next entry therefore skips it. Return the last called callback's
32-bit result, or zero if none ran. Nonzero results do not stop traversal. The dispatcher
draws nothing and shares the same mutable record across calls and with its caller. Callback
guards own eligibility of absent other/record inputs. New head entries do not restart the walk.

**IEP-R08 Generic callback inputs** [verified in 1.13c code (static)]. The generic callbacks
specified here read the current owner stat at the entry's stat identity using RLI-R02a/RLI-R02b.
Their failed guards return zero without direct draw or effect. IEP-R14 through IEP-R18 define
the exact guards; unread callbacks receive no inferred guard/draw contract.
Effective maximum mana selects stat 9 and maximum life selects stat 7 from the owner's
existing stat collection through RLI-R02a/RLI-R02b. An absent owner/list or a definition count
not exceeding the selected id supplies zero. Selector 18 consumes HIT-R02g's demon predicate.

**IEP-R09a Combat and death dispatch sites** [verified in 1.13c code (static)].

| Trigger / condition | Event, owner, other, record | Relative order |
|---|---|---|
| Missile contact | 0, contacted unit, missile, absent | Before missile hit function |
| Damage building by player/hireling attacker, record event suppression clear | 3, defender, attacker, shared record | DPO-R23 |
| Scale/mitigation | 11, defender, attacker, shared record | Before per-type mitigation, DPO-R09 |
| Melee apply, suppression clear | 5 attacker/defender then 1 defender/attacker; shared record | DPO-R13 |
| Missile apply, suppression clear | 6 attacker/defender then 2 defender/attacker; shared record | Event 6 also needs missile-attacker-event suppression clear |
| Apply reports death, suppression clear | 10 victim/killer then 9 killer/victim; shared record | After death handling |
| Regeneration death | 10 victim/killer then 9 killer/victim if killer exists; absent record | OWD-R10 |
| Queued melee completion | 7 attacker/defender then 3 defender/attacker; copied record | After durability, DPO-R23 |
| Monster enters mode 12 `dead` | 13 monster, absent other and record | After mode change |

Normal queued melee stage order is 3, 11, 5, 1, optionally 10/9, then 7/3. Missile stage
order is 0, 11, 6/2, optionally 10/9, subject to guards. Sites consume IEP-R07; callback draws
belong to their individual contracts and DPO-R29.

**IEP-R09b Level-change dispatch sites** [unresolved]. Two level-change paths dispatch event
12 on a player with absent other and record. Exact eligibility/level-change coverage is
IEP-Q05. Explicit dispatch of event 12 through IEP-R07 is independently verified.

**IEP-R10 Undispatched events** [verified in 1.13c code (static)]. The recorded complete
dispatcher call-site enumeration contains no event 4 or 8 dispatch. Event-8 entries therefore
receive no call from these original server roots. A separately registered event-7 entry is
independent. No direct draw.

**IEP-R11 Suppressed damage events** [verified in 1.13c code (static)]. Damage apply skips
events 5/1 or 6/2, and death events 10/9, when `record.suppressItemEvents` is set. Retaliation
sets it; a retaliation kill therefore raises no recursive item/kill events. This consumes
DPO-R13 rather than defining damage arithmetic. No direct draw.

**IEP-R12a Callback catalog** [verified in 1.13c code (static)]. Selectors 1 through 31 are
populated in the shared table. The following descriptions identify contracts or catalog leads;
they do not verify unread effects.

| Selector | Contract / catalog role |
|---|---|
| 1 through 5 | Skill callbacks; effects outside this contract |
| 6 | Physical retaliation, IEP-R14 |
| 7 | [Knockback](combat-knockback.md) |
| 8, 9 | Flee / blindness leads |
| 10 | Lightning retaliation, IEP-R14 |
| 11, 12 | Other retaliation leads; no item-row binding verified here |
| 13 | Damage to mana, IEP-R15 |
| 14 | Freeze lead |
| 15, 16 | [Open Wounds](combat-open-wounds.md), [Crushing Blow](combat-crushing-blow.md) |
| 17, 18 | Mana after kill / life after demon kill, IEP-R16/IEP-R17 |
| 19 | Slow lead |
| 20, 21 | Chance-to-cast leads |
| 22 through 27 | Skill callbacks; effects outside this contract |
| 28, 29 | Life after kill / Rest in Peace, IEP-R17/IEP-R18 |
| 30, 31 | Chance-to-cast / reanimation leads; registration slice IEP-R13b |

**IEP-R12b Original item bindings and unread roles** [recovered; unverified]. The recorded
catalog associates stats 78/128 with selectors 6/10 on event 1; 114 with 13 on 1/2;
138/139/86/108 with 17/18/28/29 on 9; 81/112/113/134/135/136/150 with
7/8/9/14/15/16/19 on 5/6; 195 with 20 on 7/8, 196 with 20 on 9, 198 with 20 on 5/6,
201 with 21 on 1/2, 197/199 with 30 on 10/12, and 155 with 31 on 9. IEP-Q08 blocks accepting
the complete original-profile binding catalog without named-cell records. IEP-Q06 covers
unread effect and skill roles. Supplied selector behavior is independent of these bindings.

**IEP-R13a One-shot lifetime** [verified in 1.13c code (static)]. A one-shot entry is removed
after its first matching call, regardless of result. Nonmatching dispatches retain it. No
direct draw beyond the callback contract.

**IEP-R13b Reanimation registration slice** [verified in 1.13c code (static)]. Selector 31
requires present owner and monster victim, no victim reanimation exclusion, and positive owner
stat value `v`. Then draw once from the owner seed modulo 100. If the draw is strictly less
than `v`, follow the owner chain to its first player. If found, create an event-13 one-shot
entry on the victim, with the stat layer and player identity as callback values, and return 1.
Otherwise return zero. The draw is consumed even when no player is found; failed earlier
guards draw nothing. Later spawning is outside this registration rule.

**IEP-R14 Retaliation, selectors 6 and 10** [verified in 1.13c code (static)]. For event 1,
owner is struck holder and other its melee attacker. Return zero if either is absent,
`other.retaliationEligible` is false, or owner value `v <= 0`. Otherwise create a zeroed record,
mark hit success, item-event suppression and `retaliationReactionFlag`, and set only physical
damage (6) or lightning damage (10) to `wrap32(v * 256)`. Supply physical/lightning damage-type
treatment respectively. Request scaled damage apply with holder as attacker and other as
defender, then reaction on other using holder and the same record; return 1. No direct callback
draw. Apply consumes DPO-R08c/DPO-R14 through DPO-R19 and mitigation. Reaction integration is
IEP-Q04; record construction and ordered requests remain verified.

**IEP-R15 Damage to mana, selector 13** [verified in 1.13c code (static)]. Require player
owner, present record, positive total, positive owner value `v`, and mana below effective
maximum. Total is post-mitigation but before the final life-loss clamp, so overkill counts.
Compute `new = wrap32(mana + mulDiv(v,total,100))` with NRF-R05 operand order. Floor signed
nonpositive result at zero, then cap at maximum mana. Request replacing stat 8, then call
LSE-R13's overlay operation with `(owner, id=152)`, then return 1. Failed guards return zero.
No direct draw. LSE-R13 supplies the count guard, first matching overlay-list selection or
creation, pending-condition write, stat-178 request and update request; IEP-Q09 gates their
stored effects and LSE-Q03 the original count identity. Write integration is IEP-Q10.

**IEP-R16 Mana after kill, selector 17** [verified in 1.13c code (static)]. Require present
owner/victim, player owner, nonzero `v` and mana below maximum. Compute
`new = wrap32(mana + wrap32(v * 256))`, floor nonpositive result at zero, cap at maximum,
request replacing stat 8, then LSE-R13's overlay operation with `(owner, id=152)` as in
IEP-R15, return 1. Negative values qualify. Failed guards return zero. No direct draw;
integration is IEP-Q09/IEP-Q10.

**IEP-R17 Life after kill, selectors 28 and 18** [verified in 1.13c code (static)]. Require
present owner/victim; selector 18 additionally requires demon eligibility. Any owner type
qualifies. Require nonzero `v` and life below maximum. Compute
`new = wrap32(life + wrap32(v * 256))`, floor nonpositive result at zero, cap at maximum,
request replacing stat 6, then LSE-R13's overlay operation with `(owner, id=151)` as in
IEP-R15, return 1. Failed guards return zero. No direct
draw. A negative accepted value lowers life without callback-owned death processing.
Integration is IEP-Q09/IEP-Q10.

**IEP-R18 Rest in Peace, selector 29** [verified in 1.13c code (static)]. With present owner
and victim, request state 172 `restinpeace` enabled on victim, return 1; otherwise return zero.
Do not reread the proc stat. RLI-R03 owns state routing and its update integration gate.
What the state prevents is outside this rule. No direct draw.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| IEP-R01 | Newest entry first | verified in 1.13c code (static) | none | IEP-V01 |
| IEP-R02 | Selector guards | verified in 1.13c code (static) | none | IEP-V02 |
| IEP-R03 | Registration per stat/layer | verified in 1.13c code (static) | IEP-Q01/IEP-Q03: notification integration only | IEP-V02, IEP-V03 |
| IEP-R04 | Notification integration | unresolved | IEP-Q01/IEP-Q03: initialization/notification producers | none |
| IEP-R05 | Immediate/deferred removal | verified in 1.13c code (static) | none | IEP-V03, IEP-V08, IEP-V09, IEP-V10 |
| IEP-R06 | Identical selector mapping | verified in 1.13c code (static) | none with supplied registration | IEP-V10 |
| IEP-R07 | Dispatch traversal and result | verified in 1.13c code (static) | none with supplied callbacks | IEP-V01, IEP-V04, IEP-V08, IEP-V09 |
| IEP-R08 | Guards/current value and maximum selection | verified in 1.13c code (static) | IEP-Q10: stored writes only | IEP-V05, IEP-V06, IEP-V07, IEP-V14 |
| IEP-R09a | Combat/death dispatch sites | verified in 1.13c code (static) | DPO gates: upstream production only | IEP-V11 |
| IEP-R09b | Level-change roots | unresolved | IEP-Q05: path eligibility | none |
| IEP-R10 | No event-4/8 root | verified in 1.13c code (static) | none in enumerated roots | IEP-V11 |
| IEP-R11 | Suppress recursive events | verified in 1.13c code (static) | none at supplied apply outcome | IEP-V05, IEP-V11 |
| IEP-R12a | Populated selector catalog | verified in 1.13c code (static) | none | IEP-V02, IEP-V10 |
| IEP-R12b | Item/role bindings | recovered; unverified | IEP-Q08: cells; IEP-Q06: unread roles | none |
| IEP-R13a | One-shot lifetime | verified in 1.13c code (static) | none | IEP-V04 |
| IEP-R13b | Reanimation registration | verified in 1.13c code (static) | none with supplied owner chain/exclusion; spawning outside | IEP-V12 |
| IEP-R14 | Retaliation record/requests | verified in 1.13c code (static) | IEP-Q04: reaction integration; IEP-Q02: producer only | IEP-V05 |
| IEP-R15 | Damage-to-mana requests | verified in 1.13c code (static) | IEP-Q09/IEP-Q10: stored effects | IEP-V06 |
| IEP-R16 | Mana-after-kill requests | verified in 1.13c code (static) | IEP-Q09/IEP-Q10: stored effects | IEP-V07 |
| IEP-R17 | Life-after-kill requests | verified in 1.13c code (static) | IEP-Q09/IEP-Q10: stored effects | IEP-V07 |
| IEP-R18 | Rest-in-Peace request | verified in 1.13c code (static) | RLI-Q02: update bookkeeping only | IEP-V13 |

## 5. Outputs and side effects

| Output / side effect | Rule |
|---|---|
| Entry registration, ordering, immediate/deferred removal | IEP-R01 through IEP-R07, IEP-R13a |
| Ordered callback calls, shared record mutations, final result | IEP-R07 |
| Dispatch requests on named owner/other pairs | IEP-R09a, IEP-R09b, IEP-R10, IEP-R11 |
| Owner seed advance; victim one-shot entry | IEP-R13b |
| Retaliation record; ordered apply/reaction requests | IEP-R14 |
| Replacing resource-stat request, then consumed LSE-R13 overlay list/pending/stat-178/update requests (stored effects gated by IEP-Q09) | IEP-R15 through IEP-R17 |
| Victim state-enable request; consumed RLI-R03 active-bit and update-marking requests | IEP-R18 |

No save-format or direct network-message write is specified here. Downstream damage, state
and message owners retain their contracts.

## 6. Branches and edge cases

Melee/missile apply dispatch different attacker/defender events. Different layers register
separately; multiple items at one layer do not. Equip notification order can alter callback
order (IEP-Q03). Negative bonuses qualify for selectors 17/18/28; retaliation and damage to
mana require positive values. Full resources suppress both write and overlay requests.
Generic callbacks do not directly read difficulty or player count; retaliation's consumed
path can apply PvP/mitigation. No distinct Classic profile is established. Resource arithmetic
wraps signed 32-bit intermediates before the stated signed clamps; no saturation is inferred.

## 7. Open questions

| Question | Status / kind | Missing evidence and exact boundary |
|---|---|---|
| **IEP-Q01** | unresolved / integration | Creation paths installing/retaining notification, including hirelings/monsters; complete installation/notification chain |
| **IEP-Q02** | open / outside-boundary | Wider producers/meaning of retaliation eligibility; none locally with supplied condition |
| **IEP-Q03** | unresolved / integration | Multi-stat change notification order; needed to derive equip-created list order |
| **IEP-Q04** | open / integration | Complete reaction handling of the neutral retaliation condition; record/requests remain verified |
| **IEP-Q05** | unresolved / behavior | Eligibility of two level-change roots; supplied dispatch remains verified |
| **IEP-Q06** | open / outside-boundary | Unread callback effects and exact skill bindings; none locally for dispatcher/table |
| **IEP-Q08** | unresolved / behavior | Narrow original-profile records for complete item/event catalog and absent bindings |
| **IEP-Q09** | unresolved / integration | Stored overlay attach/list writes (STL-R06b/STL-R06c), update effects (RLI-Q02) and propagation, as LSE-Q01; LSE-R13 establishes local search/creation and ordered requests, and callback requests remain verified |
| **IEP-Q10** | unresolved / integration | Replacing-write effects; supplied values/requests remain verified; write integration also RLI-Q03 |
| **IEP-Q11** | pending / observation | Matching 1.13c gameplay observations; none performed |

Attempts and closure criteria: [gap file](../plans/gaps/combat-item-event-triggers-and-procs.md).

## 8. Test vectors

### 8.1 Executable vectors

Resources are raw 8.8, ordinary integers i32. Fixtures supply valid definitions and resolved
current/max resource values; existing stat reads consume RLI-R02a/RLI-R02b. Replacing-stat,
overlay (LSE-R13, logged as `overlay<id>`), apply, reaction and state operations log ordered
requests without executing gated dependencies; LSE-V11 discriminates the overlay operation's
own search/allocation/ordered requests. Where life loss is asserted its supplied apply outcome is explicit. Entry
mutation persists within one vector. Cases start fresh unless their inputs state otherwise.
Fixture callbacks log all seven inputs and return the stated result.

- **IEP-V01** Newest first (IEP-R01, IEP-R07; boundary: supplied-input stage)
  - Inputs: initially empty U; create stat 136/event 5 then 6, then stat 135/event 5 then 6. Callbacks return 0, change nothing. Dispatch 5 with G, other T, record R.
  - Seeds: no draw.
  - Draws: none.
  - Trace: each creation prepends; event-6 entries skipped.
  - Outputs: list `(135,6),(135,5),(136,6),(136,5)`; calls 135 then 136 with `(G,5,U,T,R,its stat identity,0)`; return 0; R unchanged.
  - (a) Inputs: owner absent for the same creation requests. Outputs: no entry created and no callback invoked.
  - Discriminates: appending would call 136 first.
  - runtime_observed: no.

- **IEP-V02** Registration gates (IEP-R02, IEP-R03, IEP-R12a; boundary: supplied-input stage)
  - Inputs: empty U, stat identity S, new total 5, first event 1/selector 6, second event 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: first event positive, no existing entry, selector populated.
  - Outputs: exactly one item entry with event 1, selector 6, identity S, values `(S,0)`.
  - (a) Inputs: fresh U, first event 0, second event 5/selector 6. Outputs: no entries.
  - (b) Inputs: fresh U, first event 1, selector each of -1, 0, 32, 49, 50, second event 0. Outputs: no entry in each case.
  - (c) Inputs: fresh U, first event 1/selector 0, second event 5/selector 6. Outputs: exactly the event-5 entry.
  - (d) Inputs: fresh U for each selector 1 through 31, first event 1, second event 0; no callback execution requested. Outputs: exactly one item entry for each selector, retaining that selector's callback identity, event 1, identity S and values `(S,0)`.
  - (e) Inputs: fresh U, first event 1/selector 6, second event 0, new total -5. Outputs: exactly one item entry with event 1, selector 6, identity S and values `(S,0)`.
  - Discriminates: (a) rejects independent second event; (c) rejects requiring first-selector success.
  - runtime_observed: no.

- **IEP-V03** Shared contribution lifetime (IEP-R03, IEP-R05; boundary: supplied-input stage)
  - Inputs: empty U, stat 138/layer 0, supplied first event 9/selector 17 and second event 0/selector 0, unless a case overrides them; distinct item contributors A and B each contribute 5. Add A, add B, remove A, remove B, supplying successive summed-total notifications 0 to 5, 5 to 10, 10 to 5, 5 to 0. Contributor identity does not enter the registration identity.
  - Seeds: no draw.
  - Draws: none.
  - Trace: only first change creates; final zero removes.
  - Outputs: counts `[1,1,1,0]`, same entry through middle changes.
  - (a) Inputs: fresh U; stat 138 at layers 1 and 2; each supplied definition has event 5/selector 6 then event 6/selector 10. Notify layer 1 from 0 to 5, then layer 2 from 0 to 7, then layer 1 from 5 to 0. Outputs: counts `[2,4,2]`; final entries from head are layer-2/event-6 then layer-2/event-5, values `(stat138 at layer2,0)`; both layer-1 entries removed, both layer-2 entries retained.
  - Discriminates: per-item creation gives two at total 10.
  - runtime_observed: no.

- **IEP-V04** One-shot lifetime (IEP-R07, IEP-R13a; boundary: supplied-input stage)
  - Inputs: monster U with one event-13 one-shot, values `(7,42)`, callback returns 9; dispatch 12 then 13 then 13; absent other/record.
  - Seeds: no draw.
  - Draws: none.
  - Trace: nonmatch retains; first match calls and removes.
  - Outputs: returns `[0,9,0]`; one callback with event 13/values `(7,42)`; final list empty.
  - (a) Inputs: fresh monster U with the same event-13 one-shot and values `(7,42)`, callback returns 0; dispatch 12, then 13, then 13, with absent other/record. Outputs: returns `[0,0,0]`; callback called exactly once with event 13/values `(7,42)`; final list empty.
  - Discriminates: removal on nonmatch or retention after result 9 fails; (a) rejects removing a one-shot only when its callback returns nonzero.
  - runtime_observed: no.

- **IEP-V05** Retaliation/suppression (IEP-R08, IEP-R11, IEP-R14; boundary: supplied-input stage)
  - Inputs: holder U stat 78 = 10; attacker T monster eligible for retaliation; selector 6; supplied apply outcome with physical resistance 0, no flat reduction, life 2560 to 0. Reaction logs only.
  - Seeds: no draw in callback/request boundary; no dependency draws supplied.
  - Draws: none.
  - Trace: `10*256=2560`; hit, suppression and reaction conditions set; apply before reaction; supplied outcome death.
  - Outputs: physical 2560, other components zero; ordered `[apply(U,T,scaled,R),reaction(T,U,R)]`, return 1; supplied life loss 2560 (10 HP); no apply item/kill events.
  - (a) Inputs: fresh case eligibility false. Outputs: return 0, no effect requests.
  - (b) Inputs: fresh case selector 10, owner stat 128 = 10; no dependency effects executed. Outputs: lightning 2560, physical and all other components zero; hit, suppression and reaction conditions set; ordered `[apply(U,T,scaled,R),reaction(T,U,R)]`, return 1.
  - (c) Inputs: separate fresh base cases with owner absent, other absent, owner value 0, or owner value -1. Outputs: in every case return 0, no apply/reaction requests and no direct draw.
  - (d) Inputs: fresh selector-6 case, owner stat 78 = 1, eligible present U/T; no dependency effects executed. Outputs: physical 256, all other components zero; hit, suppression and reaction conditions set; ordered `[apply(U,T,scaled,R),reaction(T,U,R)]`, return 1.
  - Discriminates: false condition rejects ignoring eligibility; death rejects recursive kill events.
  - runtime_observed: no.

- **IEP-V06** Damage-to-mana (IEP-R08, IEP-R15; boundary: supplied-input stage)
  - Inputs: player U, stat 114 = 20, mana 0x1000, maximum 0x6400, current life 0x1000; post-mitigation record total 0x3200, physical 0x1000, fire 0x2200, all other damage components zero. The supplied record total exceeds current life and differs from physical damage.
  - Seeds: no draw.
  - Draws: none.
  - Trace: `mulDiv(20,12800,100)=2560`; `4096+2560=6656`.
  - Outputs: requests `[replace mana 0x1A00,overlay152]`, return 1.
  - (a) Inputs: fresh case mana 0x6400. Outputs: no requests, return 0.
  - (b) Inputs: fresh case monster owner. Outputs: no requests, return 0.
  - (c) Inputs: separate fresh base cases with owner absent, record absent, total 0, total -1, stat value 0, or stat value -1. Outputs: in every case no requests, return 0.
  - (d) Inputs: fresh base case mana 0x6300. Outputs: `[replace mana 0x6400,overlay152]`, return 1.
  - (e) Inputs: player U, value 100, mana 2147483646, maximum 2147483647, present record total 2, physical 2 and other components zero. Outputs: `[replace mana 0,overlay152]`, return 1; `mulDiv(100,2,100)=2`, wrapped sum -2147483648 is floored at zero.
  - Discriminates: using only physical damage or clamping to current life gives mana 0x1333 instead of 0x1A00; guards, signed-wrap floor and upper cap are distinct cases.
  - runtime_observed: no.

- **IEP-V07** Kill resource callbacks (IEP-R08, IEP-R16, IEP-R17; boundary: supplied-input stage)
  - Inputs: player killer/victim present; selector 28, stat 86 = 5, life 0x6400, maximum 0x7D00.
  - Seeds: no draw.
  - Draws: none.
  - Trace: `5*256=1280`; `25600+1280=26880`.
  - Outputs: `[replace life 0x6900,overlay151]`, return 1.
  - (a) Inputs: fresh case life 0x7C80. Outputs: `[replace life 0x7D00,overlay151]`, return 1.
  - (b) Inputs: selector 18/stat 139 = 5, non-demon victim, same resources. Outputs: no requests, return 0.
  - (c) Inputs: monster killer, selector 17/stat 138 = 3, mana 0x1000, maximum 0x6400. Outputs: no requests, return 0.
  - (d) Inputs: player killer, selector 17/stat 138 = -3, mana 512, maximum 25600. Outputs: `[replace mana 0,overlay152]`, return 1.
  - (e) Inputs: player killer, selector 17/stat 138 = 3, mana 0x1000, maximum 0x6400. Outputs: `[replace mana 0x1300,overlay152]`, return 1.
  - (f) Inputs: player killer and victim present, selector 17; first fresh case stat 138 = 3, mana/maximum both 0x6400; second fresh case stat 138 = 0, mana 0x1000, maximum 0x6400. Outputs: no requests, return 0 in each case.
  - (g) Inputs: fresh base case selector 18/stat 139 = 5, demon victim. Outputs: `[replace life 0x6900,overlay151]`, return 1.
  - (h) Inputs: fresh base case selector 28/stat 86 = -5, life 512. Outputs: `[replace life 0,overlay151]`, return 1, no death-processing request.
  - (i) Inputs: separate fresh base cases stat value 0 or life 0x7D00. Outputs: no requests, return 0 in each case.
  - (j) Inputs: selector 17/stat 138 = 3 with mana 0x1000/maximum 0x6400; selector 18/stat 139 = 5 with life 0x6400/maximum 0x7D00 and demon victim; selector 28/stat 86 = 5 with those same life values. For each, player owner and victim are present in the starting fixture; separately make owner absent or victim absent. Outputs: no requests, return 0 in each of the six fresh cases.
  - Discriminates: cap, demon/player guards and nonzero rather than positive kill-bonus test.
  - runtime_observed: no.

- **IEP-V08** Successor reload/arguments (IEP-R05, IEP-R07; boundary: supplied-input stage)
  - Inputs: G,U,T,event 11; shared R physical 2560, otherwise zero. U list X,Y,Z, event 11/item kind, identities 100/200/300, values `(7,0)/(3,0)/(9,0)`. X removes identity 200 and returns 5; Y returns 3; Z returns 7; no record changes.
  - Seeds: no draw.
  - Draws: none.
  - Trace: X removes nonexecuting Y; reloaded successor Z; nonzero result does not stop.
  - Outputs: calls `(G,11,U,T,R,7,0)` then `(G,11,U,T,R,9,0)`, each sees physical 2560; Y uncalled; return 7; list X,Z, neither executing; R unchanged.
  - (a) Inputs: fresh list all events 5, dispatch 11. Outputs: no calls, return 0, list X,Y,Z/R unchanged.
  - (b) Inputs: absent owner. Outputs: no calls, return 0, R unchanged.
  - (c) Inputs: fresh list X,Z on event 11 with base identities/values; R physical 2560. Override X to set R.physical to 2816, create event-11 item entry H at the head with identity 400/values `(0,0)`, then return 5. H would return 99; Z logs observed physical and returns 7. Outputs: calls X then Z, H never called; Z sees physical 2816; return 7; final list H,X,Z, none executing; R.physical 2816 and other fields unchanged.
  - Discriminates: pre-call-only successor, swapped units, omitted event argument, first-nonzero stop all fail.
  - runtime_observed: no.

- **IEP-V09** Deferred self-removal (IEP-R05, IEP-R07; boundary: supplied-input stage)
  - Inputs: U list X,Z on event 9, distinct item identities; X removes its own identity, logs continued in-call presence, returns 4; Z returns 8.
  - Seeds: no draw.
  - Draws: none.
  - Trace: X executing, removal deferred; removed after call, then Z visited.
  - Outputs: X present during its callback; calls X,Z; final list Z; return 8; Z not executing.
  - (a) Inputs: fresh U list X,Z as in the base case; X requests its own removal and logs continued in-call presence, but returns 0; Z returns 8. Outputs: X present during its callback; calls X,Z with callback results `[0,8]`; final list Z; return 8; Z not executing.
  - Discriminates: immediate self-removal fails presence assertion; stop at 4 omits Z; (a)'s final list rejects making deferred removal conditional on a nonzero callback result.
  - runtime_observed: no.

- **IEP-V10** Shared selection (IEP-R02, IEP-R05, IEP-R06, IEP-R12a; boundary: supplied-input stage)
  - Inputs: empty U; item then skill registration, selector 6/event 1, distinct identities, respective values `(S,0)/(K,2)`; callback fixture returns 1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: both resolve same callback; skill entry newest.
  - Outputs: same callback identity, kinds skill/item; skill then item calls with own values forwarded; return 1.
  - (a) Inputs: fresh U with supplied event-1 entries H (skill) then I (item), both owner identity J; H values `(9,0)`, I values `(7,0)`, each callback returns 1. Remove skill kind/identity J, then dispatch event 1 with game G, other T and unchanged shared record R. Outputs: removal retains exactly I; exactly one call `(G,1,U,T,R,7,0)`; return 1; I retained, H uncalled, R unchanged.
  - Discriminates: assigning a different callback mapping by registration source or forcing item values for skill entries fails. Separate storage tables with identical mappings satisfy this semantic requirement.
  - runtime_observed: no.

- **IEP-V11** Reached stages (IEP-R09a, IEP-R10, IEP-R11; boundary: supplied-input stage)
  - Inputs: queued melee by player attacker A against monster defender D. The shared record's `suppressItemEvents` condition is false at build, apply and death dispatch; the reached build site therefore satisfies DPO-R23's player/hireling-attacker and clear-result conditions. At the reached death tail, death is explicitly reported and DPO-R23's death-event gate is supplied as allowed: the result event-suppression condition it tests is clear for this same record. The death report and allowed result gate are separate supplied outcomes. Queued completion is reached. Dispatch logs requests, no registered callbacks; only the listed reached dispatch sites execute in this boundary. Reached-stage and eligibility outcomes are supplied dependency results; their upstream producers do not execute in this fixture.
  - Seeds: no draw at reached dispatch sites.
  - Draws: none.
  - Trace: build, mitigation, apply pair, death pair, completion pair.
  - Outputs: `(event,owner,other)` log `(3,D,A),(11,D,A),(5,A,D),(1,D,A),(10,D,A),(9,A,D),(7,A,D),(3,D,A)`; no 4/8; completion uses copied record, earlier stages shared record.
  - (a) Inputs: supplied apply only, suppression set, death reported. Outputs: no apply item/kill dispatches.
  - (b) Inputs: reached missile contact by missile M on D, then mitigation and nonlethal missile apply from A to D; both suppression conditions clear. Outputs: `(0,D,M),(11,D,A),(6,A,D),(2,D,A)`; contact has absent record, later calls share the damage record; no death pair.
  - (c) Inputs: supplied nonlethal missile apply only, item-event suppression clear but missile-attacker-event suppression set. Outputs: exactly `(2,D,A)` with the supplied record; no event 6.
  - (d) Inputs: reached regeneration death for D with killer A. Outputs: `(10,D,A),(9,A,D)`, both with absent record.
  - (e) Inputs: reached regeneration death for D with absent killer. Outputs: exactly `(10,D,absent)` with absent record.
  - (f) Inputs: reached monster M transition to mode 12 dead. Outputs: exactly `(13,M,absent)` with absent record, after the mode transition.
  - (g) Inputs: isolated reached apply death tail for player A and monster D; death reported, but the same record's result event-suppression condition is set, so the supplied DPO-R23 death-event gate is blocked. Earlier apply/mitigation and queued completion do not execute. Outputs: empty dispatch log, specifically no events 10 or 9.
  - (h) Inputs: two isolated build-notification decisions against D, with eligibility supplied under DPO-R23. First: ordinary monster attacker M with non-hireling classification and clear result event-suppression condition. Second: player A with result event-suppression condition set. No other stages execute. Outputs: empty dispatch log in each fresh case, specifically no build event 3.
  - (i) Inputs: isolated reached flagged missile-apply item-event decision from player A to monster D; preceding apply gates succeed, hit succeeds, flagged events are enabled and missile-attacker-event suppression is false, but the record's result event-suppression condition `suppressItemEvents` is true. Earlier scale/mitigation and later death/completion do not execute. Outputs: empty dispatch log, specifically neither event 6 nor event 2.
  - (j) Inputs: isolated reached regeneration death-notification decision for monster D with known killer A; supplied regeneration outcome remains alive with life 2560, so no regeneration death occurs. Other regeneration work does not execute. Outputs: empty dispatch log, specifically no events 10 or 9.
  - Discriminates: reversed units/death order fail the base log; (g) rejects ignoring the death result gate, (h) rejects ignoring build eligibility, and (i) rejects treating the clear missile-attacker suppression condition as sufficient to dispatch event 6. Suppression controls reject recursive item/kill events.
  - runtime_observed: no.

- **IEP-V12** Reanimation registration (IEP-R13b; boundary: supplied-input stage)
  - Inputs: monster owner U, stat value 100/layer 7; nonexcluded monster victim; owner chain U then player P identity 42; victim list empty.
  - Seeds: U `{lo:1,hi:0}`.
  - Draws: one `(U,100,85)` under NRF-R01/NRF-R02.
  - Trace: stepped seed `{lo:0x6AC690C5,hi:0}`; 1791398085 mod 100 = 85 < 100; first player P.
  - Outputs: event-13 one-shot on victim, callback values `(7,42)`; stepped U seed; return 1; no spawn asserted.
  - (a) Inputs: fresh same seed/value/victim, no player in owner chain. Outputs: no entry, same stepped seed, return 0.
  - (b) Inputs: fresh seed, value 0. Outputs: no draw, seed unchanged, no entry, return 0.
  - (c) Inputs: fresh base case victim excluded from reanimation. Outputs: no draw, seed unchanged, no entry, return 0.
  - (d) Inputs: fresh base case victim a player. Outputs: no draw, seed unchanged, no entry, return 0.
  - (e) Inputs: fresh base case owner absent. Outputs: no draw, no entry, return 0.
  - (f) Inputs: fresh base seed and victim, value 85. Outputs: one `(U,100,85)` draw, seed `{0x6AC690C5,0}`, no entry, return 0 because equality fails the strict chance test.
  - (g) Inputs: player owner P identity 42 with seed `{1,0}`, value 100/layer 7, same victim. Outputs: one `(P,100,85)` draw, P seed `{0x6AC690C5,0}`, event-13 one-shot on victim with values `(7,42)`, return 1; no owner-chain lookup beyond P.
  - Discriminates: (a) rejects resolving player before draw; (b) rejects drawing for nonpositive value.
  - runtime_observed: no.

- **IEP-V13** Rest-in-Peace request (IEP-R18; boundary: supplied-input stage)
  - Inputs: present player U and monster T, selector 29, stat value unspecified/unread; supplied state-definition count 173, T has a base list with state storage. Execute RLI-R03 state-call routing; its active-bit writer and update marker log requests without executing stored-state/update-bookkeeping effects.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unit guards pass; callback requests state 172 enabled. Since 172 < 173, RLI-R03 routes the active-bit request then unconditionally requests update marking.
  - Outputs: callback requests `enable state172 on T`; consumed routing emits ordered `[active-bit(T,172,1),MarkForUpdate(T)]`; callback return 1, no stat read. No stored-state or update-queue result is asserted.
  - (a) Inputs: absent victim. Outputs: no requests, return 0.
  - (b) Inputs: absent owner. Outputs: no requests, return 0.
  - Discriminates: positive-stat check or state on owner fails. Consumed state-routing controls for absent base storage and invalid state identity belong to RLI-V05(c)/(d)/(k); this callback's selected identity and request do not change those shared rules.
  - runtime_observed: no.

- **IEP-V14** Effective maximum selection (IEP-R08; boundary: supplied-input stage)
  - Inputs: player U with existing own-stat collection, no resolved-stat selection or minimum replacement; stat 7/layer 0 = 32000, stat 9/layer 0 = 25600, definition count 359, valid definitions for 7 and 9.
  - Seeds: no draw.
  - Draws: none.
  - Trace: 359 exceeds each selected id; RLI-R02a/RLI-R02b read the corresponding existing values.
  - Outputs: maximum life 32000, maximum mana 25600.
  - (a) Inputs: fresh case definition count 9. Outputs: maximum life 32000, maximum mana 0.
  - (b) Inputs: owner has no stat list. Outputs: both maxima 0.
  - Discriminates: swapping stat ids reverses the distinct results; treating definition count 9 as including id 9 fails (a).
  - runtime_observed: no.

### 8.2 Observation vectors

- **IEP-O01** Equip/unequip shared proc stats; expect entries per stat/layer, retained until
  zero; measure multi-stat notification and entry order.
- **IEP-O02** Measure killing queued melee owner/other/event order; expect IEP-V11 subject
  to its guards.
- **IEP-O03** Measure event-8 registration during missile attacks; expect no event-8 dispatch
  from the enumerated original roots.
- **IEP-O04** Measure retaliation with known resistance/reduction; expect consumed mitigation
  result and no recursive item/kill events.

All pending. [Reproduction procedures](../research/spec-evidence/combat-item-event-triggers-and-procs.md#observation-procedures)
state the capture boundaries and required inputs.

## 9. Acceptance and dependencies

Partial acceptance is the supplied-input assertions of IEP-V01 through IEP-V14 for verified
rules in the Id map. Requests do not establish stored effects; table lookup does not verify
unread callbacks. Complete reproduction additionally needs IEP-Q01/IEP-Q03 notification,
IEP-Q05 level roots, IEP-Q08 data bindings, IEP-Q09/IEP-Q10 resource/overlay integration and
relied-on reaction/state effects. IEP-Q11 limits runtime-confirmation claims only. Outside
effect families are not prerequisites to reproducing the supplied-callback dispatcher.

**Depends on:** NRF-R01/NRF-R02 seed/draw; NRF-R05 percentage arithmetic; NRF-R08a resource Owners of the cited HIT, GDA, RLI ids: [combat-hit-roll](combat-hit-roll.md) (HIT); [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).
units; RLI-R03 state routing; RLI-R02a/RLI-R02b existing stat reads; GDA-R12 definition fields
and GDF-R03 state names. [Damage pipeline](combat-damage-pipeline-order.md),
DPO-R08c/DPO-R09/DPO-R13 through DPO-R19/DPO-R23/DPO-R29, owns stages, damage writes and
downstream draw order. [Mitigation](combat-damage-mitigation.md) owns retaliation reduction
and resistance. [Open Wounds](combat-open-wounds.md), OWD-R10, owns regeneration-death context.
Replacing-write effects retain RLI-Q03; update bookkeeping retains RLI-Q02.
[Shield block and avoidance](combat-shield-block-and-avoidance.md), HIT-R02g, supplies the
demon predicate for selector 18.
[Leech and drain](combat-leech-and-drain.md), LSE-R13, owns the overlay
operation; its stored effects retain IEP-Q09 and its count identity LSE-Q03.

**Depended on by:** damage-pipeline event traversal; Crushing Blow, Open Wounds and knockback
registration/ordering; foundations' callback-owned draw discussion. Reverse references are
not automatic delivery prerequisites.

## 10. Related specs

- [combat-crushing-blow](combat-crushing-blow.md): selector 16.
- [combat-open-wounds](combat-open-wounds.md): selector 15 and regeneration death.
- [combat-knockback](combat-knockback.md): selector 7.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): damage/event order.
- [combat-damage-mitigation](combat-damage-mitigation.md): retaliation mitigation.
- [combat-leech-and-drain](combat-leech-and-drain.md): overlay operation.
- [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md): shared primitives.
- [core-game-data-fields](core-game-data-fields.md): fields and names.

## Appendix A. Provenance

Original Windows x86 LoD Expansion 1.13c: D2Game/D2Common with Fog/Storm dependencies in the
[source manifest](../research/source-manifest.md). Compatible static readings dated
2026-10-12 and dispatcher correction dated 2026-10-13 retain their dates and boundaries.
Targeted 2026-10-14 readings, named-cell limits and completed reader review are recorded in the
[companion](../research/spec-evidence/combat-item-event-triggers-and-procs.md). No gameplay
observations have been made; the full item-binding catalog remains explicitly gated.
