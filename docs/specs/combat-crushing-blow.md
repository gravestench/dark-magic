# Crushing Blow

**Description**

Crushing Blow gives an attack a chance to take away a portion of the target's remaining life
before the attack's normal damage. Its effect becomes smaller as the target loses life, and
it removes a smaller portion against players, hirelings, certain monsters and ranged hits.
Physical resistance can reduce the loss.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: CBL-R, CBL-Q, CBL-V, CBL-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-crushing-blow.md)
Open evidence gaps: [gap file](../plans/gaps/combat-crushing-blow.md)

## 1. Scope

This contract covers the Crushing Blow item-event callback: registration choices for stat 136
`item_crushingblow`, event placement, the chance draw, defender divisor, amount, requested life
replacement, damage-result death flag and overlay request. Registration and damage application
are consumed at the supplied-notification and supplied-apply-stage boundaries of IEP-R03 and
DPO-R13a/DPO-R23. Automatic equipment notification, missile eligibility/owner lookup, stat-write
integration, client messages and classification producers retain their owners/questions below.
Open Wounds, ordinary damage, leech and other callbacks are separate behaviors.

## 2. Trigger and ownership

D2Game computes the callback on the server. A supplied stat-change notification for stat
136/layer 0 selects CBL-R01's registrations. Producing those notifications for every actual
unit/equipment change remains IEP-Q01/IEP-Q03 and CBL-Q11.

After the apply gate and requested mitigation, DPO-R13a dispatches attacker event 5 then defender
event 1 for queued melee, or eligible attacker event 6 then defender event 2 for non-queued
apply. IEP-R07 invokes the callback once per matching registration reached. This is before the
physical clamp, leech, ordinary life subtraction and status application (DPO-R14 to DPO-R19).
A supplied non-queued caller can raise event 6 without being a missile; the event still selects
the doubled divisor. Original missile owner/eligibility belongs to MHD-R01/MHD-R12.

The callback sends no client message. Requested stat writes and update marking do not establish
message selection or delivery (CBL-Q08). The [dispatch evidence](../research/spec-evidence/combat-crushing-blow.md#dispatch-evidence)
records both original roots and their remaining producer limits.

## 3. Inputs and data

Names describe logical operations, not required APIs or layouts. `Stat` consumes the existing
selected-array getter RLI-R02a/RLI-R02b: it adds no contributions during a read. The chance read
uses the registered stat identity; other reads explicitly select the ids below. `ReplaceLife`
is a replacing unit-stat request, not an established stored effect (CBL-Q10/RLI-Q03).
`low32` and signed division use the foundations' notation; divisions truncate toward zero.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| Attacker and defender | Callback owner and other | optional unit references | absent either returns 0 |
| Chance identity/value | Registration `(136,0)`, stat `item_crushingblow` through RLI-R02a/RLI-R02b | signed i32 percent | nonpositive skips draw |
| Attacker seed | Attacker-owned NRF-R01 state | two u32 values `{lo,hi}` | advanced for chance |
| Event id | Supplied item event | i32: 5 `domeleedamage`, 6 `domissiledamage` | only id 6 doubles divisor |
| Defender type/class | Logical type and class | u32 type; i32 class | type 0 player, 1 monster |
| Hireling membership | GDF-R08 class identities; engine membership test | boolean | monster classes 271 `roguehire`, 338 `act2hire`, 359 `act3hire`, 560 `act5hire1`, 561 `act5hire2`; false for non-monster |
| Boss record/value | Loaded MonStats class record/count, GDA-R08a/GDA-R04a | optional record, boolean `boss` | false for missing/out-of-range class record |
| `monsterDivisorFlag` | Supplied monster-data condition | boolean | false for non-monster/absent monster data; assignment remains CBL-Q01 |
| Player count | Defender stat 100 `monster_playercount`, layer 0 | signed i32 | caller floors to 1, no upper cap |
| Bonus table | CBL-R05 engine constants | nine i32 entries indexed 0..8 | `{0,0,50,100,150,200,250,300,350}` |
| Life | Defender stat 6 `hitpoints`, layer 0 | signed i32 raw 8.8 | current supplied getter result |
| Resistance | Defender stat 36 `damageresist`, layer 0 | signed i32 percent | positive values reduce amount; ceiling 100 |
| Damage result | Optional shared record | `hit`, `suppressEvents`, `death` booleans; other fields retained | dispatch tests and death write |
| Apply mode/conditions | Supplied DPO apply stage | `queuedMelee`/`nonQueued`; `itemEventsEnabled`, `suppressAttackerMissileEvent` booleans | CBL-R02a gates |
| Missile conditions | Supplied missile/data | `enablesItemEvents`, `suppressesAttackerEvent` booleans | false for absent data/non-missile; OR transfer preserves incoming flags |
| Registration/data | IEP-R01/R02/R03/R05/R07 and GDA-R12 | ordered entries, event/selector fields | item kind, identity, handler and callback values |
| Overlay id/count | Fixed request id 147 (`0x93`); supplied `overlayCount` | signed i32 | LSE-R13 guard; original count identity remains CBL-Q04 |
| Overlay list inputs | LSE-R13/STL-R06a | pool/type/id, optional base list, primary chain and conditions | selection/creation requests; stored effects remain CBL-Q09 |
| Getter inputs | [record layouts](core-record-layouts-and-ids.md) section 3.7, RLI-R02a/RLI-R02b | definitions/minimum metadata, existing ordered arrays | missing definitions/entries/list follow getter guards |
| Write operations | Replacing unit write; LSE-R13 attach/list-write/update operations | ordered requests | actual effects separately gated |

### 3.1 Origin and binding of mappings and constants

Numeric identities in this spec describe the original 1.13c data profile. Their names identify
the behavioral inputs; custom remapping would be a separate implementation capability.
An identity listed in a table may also be selected explicitly by engine code, so changing
the table alone does not establish a changed code binding.

Paths below are relative to `data/global/excel/` in the loaded 1.13c archive set. Cell
checks dated 2026-10-13 for `itemstatcost.txt` and `events.txt` are recorded in
[Table-data verification](../research/spec-evidence/combat-crushing-blow.md#table-data-verification).
Other table-field mappings use the cited shared specs; no new per-monster cell values are
claimed here. Code consumers remain supported by the companion's source records.

| Mapping or constant | Origin | Original binding and source |
|---|---|---|
| Chance stat: `item_crushingblow` = 136 | Table data | `itemstatcost.txt`, row `Stat = item_crushingblow`, column `ID = 136`. The handler reads the stat identity supplied by registration, not a literal chance-stat id (R01, R03, IEP-R03). |
| Life: `hitpoints` = 6 | Table data for name/id; engine code for selection | `itemstatcost.txt`, row `Stat = hitpoints`, column `ID = 6`; R06 reads and R07 requests a write to original id 6. |
| Resistance: `damageresist` = 36 | Table data for name/id; engine code for selection | `itemstatcost.txt`, row `Stat = damageresist`, column `ID = 36`; R06 explicitly reads original id 36. |
| Player count: `monster_playercount` = 100 | Table data for name/id; engine code for selection | `itemstatcost.txt`, row `Stat = monster_playercount`, column `ID = 100`; R04a explicitly reads original id 100. Its value at creation belongs to MSD-R09. |
| Overlay stat: `unit_dooverlay` = 178 | Table data for name/id; engine code for selection | `itemstatcost.txt`, row `Stat = unit_dooverlay`, column `ID = 178`; the overlay helper requests original id 178 (R08; stored effects gated). |
| Event names and ids | Table data for names/indices; engine code for dispatch | `events.txt`, column `event`, zero-based data rows excluding the header: 1 `damagedinmelee`, 2 `damagedbymissile`, 5 `domeleedamage`, 6 `domissiledamage`. R02a's dispatch selects these ids; R04a doubles the divisor for id 6. |
| Event registration choices | Table data | `itemstatcost.txt`, row `Stat = item_crushingblow`: `itemevent1 = domeleedamage`, `itemevent2 = domissiledamage`; resolved through the event-name lookup (GDA-R12, IEP-R03). |
| Crushing Blow handler selection | Table data selecting an engine-defined handler | Same row: `itemeventfunc1 = 16`, `itemeventfunc2 = 16`. Data selects 16; the engine handler mapping supplies its behavior (IEP-R02, IEP-R12). |
| Class `boss` value | Table data | `monstats.txt`, defender class row identified by `Id`/`hcIdx`, column `boss` (GDF-R08, GDA-R08a). Resolved from that record; R04a defines its divisor effect. |
| Hireling classes | Table data for class identities; engine code for membership test | `monstats.txt`, rows `Id = roguehire`, `act2hire`, `act3hire`, `act5hire1`, `act5hire2`, column `hcIdx` (GDF-R08). The original helper recognizes the five listed class ids; a new class is not automatically a hireling. |
| Divisors 4/8/10, missile factor 2, resistance ceiling 100 | Engine code | Fixed by R04a/R06, not loaded from game-table columns. |
| Player-count bonus table and formula | Engine code | The constant table and `(n - 2) * 50` branch are fixed by R05; player count itself is an input. |
| Overlay id `0x93` | Engine code | Fixed request argument in R08. Its overlay-table row identity remains unestablished; no named overlay row is claimed. |
| `monsterDivisorFlag` assignment and meaning | Unresolved | A runtime condition is read by code; the assigning behavior/monster kinds remain Q01/R04b. No table column is established. |
| Overlay-count table binding | Unresolved | R08 reads a bound, but its originating table remains Q04. |

## 4. Rules (execution order)

**CBL-R01 Wiring** [verified in 1.13c code (static)]. At an IEP-R03 supplied notification for
stat 136/layer 0, nonzero new value creates one pair only if no item registration for that
identity exists: first event 5, then event 6, each selector 16 and callback values `(stat
identity,0)`. Head insertion orders them event 6 then event 5. An existing registration leaves
the list unchanged; zero removes all item registrations for the identity through IEP-R05.
Several contributing items share the pair. Negative totals register, although CBL-R03 refuses
a nonpositive chance. Automatic notification remains CBL-Q11.

IEP-R07 owns traversal, executing/deferred-removal conditions and the last-handler return (0
when no match). Apply ignores that return. Newer registrations run first; runtime confirmation
of a concrete Crushing Blow/Knockback/Open Wounds sequence remains CBL-Q03. Registration and
dispatch make no direct draw; invoked callbacks own their draws.

**CBL-R02 Dispatch point** (heading; split into sub-ids).

- **CBL-R02a** [verified in 1.13c code (static)]. At the DPO-R13a supplied post-gate,
  post-mitigation stage, consume DPO-R23:

  ```
  if result.suppressEvents or not result.hit: no item events here
  elif applyMode == nonQueued:
      if itemEventsEnabled:
          if not suppressAttackerMissileEvent:
              Dispatch(6, owner=attacker, other=defender, result)
          Dispatch(2, owner=defender, other=attacker, result)
  else:
      Dispatch(5, owner=attacker, other=defender, result)
      Dispatch(1, owner=defender, other=attacker, result)
  continue at physical clamp
  ```

  Dispatches share the mutable record and ignore dispatcher returns. The outer result test
  precedes these calls; a callback mutation does not restart it. No direct draw.
- **CBL-R02b** [verified in 1.13c code (static)]. At supplied record preparation, queued
  melee replaces all damage flags with enabled true, attacker-event suppression false, all
  other flags clear. Missile contact ORs each true missile condition into its corresponding
  record condition and retains unrelated/incoming flags. A false missile condition does not
  clear a true incoming condition. No direct draw. Original condition production, owner
  resolution and full contact eligibility remain CBL-Q04/CBL-Q05.

**CBL-R03 Chance roll** [verified in 1.13c code (static)].

```
if attacker absent or defender absent: return 0
chance = Stat(attacker, registered chance identity)
if chance <= 0: return 0
step(attacker.seed)                         # NRF-R01, both u32 values written
roll = unsigned(attacker.seed.lo) modulo 100
if signed(roll) >= chance: return 0
continue at CBL-R04a
```

Exactly one attacker draw occurs for positive chance, including at least 100. Early exits
make no life/overlay request and leave the damage result unchanged.

**CBL-R04 Divisor by defender** (heading; split into sub-ids).

- **CBL-R04a** [verified in 1.13c code (static)].

  ```
  d = 4
  if defender.type == player: d = 10
  elif defender.type == monster:
      if IsHireling(defender): d = 10
      else:
          if Boss(defender): d = 8
          elif monsterDivisorFlag: d = 8
          n = Stat(defender, 100, 0)
          if n < 1: n = 1
          f = PCB(n)                         # CBL-R05
          if f != 0: d = low32(d + mulDiv(f, d, 100))
  if eventId == 6: d = low32(d * 2)
  ```

  Comparisons are signed; wrapped results are i32. Hirelings skip boss/condition/count
  tests. A true boss predicate skips the separate condition read. Other types retain 4
  before doubling.
- **CBL-R04b** [unresolved]. Assignment/classification of `monsterDivisorFlag` remains
  CBL-Q01. CBL-R04a's conditional read and supplied effect are verified; no champion/unique
  meaning or default for present monster data is established.

**CBL-R05 Player-count bonus** [verified in 1.13c code (static)]. The caller supplies i32
`n >= 1`. Return `T[n]` when `n < 9`, otherwise signed `low32((n-2)*50)`. Index 0 is not
reached by this caller; negative indexing is outside this helper's caller boundary. Counts
8/9 give 350; 10 gives 400. Multiplication can wrap negative. No draw.

**CBL-R06 Amount** [verified in 1.13c code (static)].

```
life = Stat(defender, 6, 0)
q = signedDivide(life, d)
r = Stat(defender, 36, 0)
if r >= 100: r = 100
if r > 0: q = low32(q - mulDiv(r, q, 100))
```

All values are i32. `mulDiv` takes operands in that order (NRF-R05). Nonpositive resistance
leaves the amount unchanged. A zero divisor or unrepresentable signed quotient faults before
the resistance read, life request, death write and overlay request, with no normal callback
return. No direct draw after CBL-R03. Divide-first precision can leave a nonzero amount at
resistance 100 (CBL-V11).

**CBL-R07 Life request and death flag** [verified in 1.13c code (static)].

```
new = signed(low32(life - q))
if new <= 0: new = 0
ReplaceLife(defender, stat=6, layer=0, value=new)
if new <= 0 and result exists: result.death = true
```

The request occurs even for an unchanged value; the death flag is never cleared here. Actual
array/contribution changes, insertion/no-op behavior, propagation and player notification
remain CBL-Q10/RLI-Q03. A supplied fixture write can persist the value independently. Positive
life with divisor at least 4 and nonoverflowing arithmetic leaves positive requested life;
extreme counts do not establish that divisor condition (CBL-V15).

**CBL-R08 Overlay request and successful return** [verified in 1.13c code (static)]. After
the life request/death write, if signed `q > 0`, call LSE-R13's overlay operation with
`(defender, id=147)`; otherwise skip it. Then return 1 without testing the
overlay operation's return. No additional draw. LSE-R13 supplies the count guard, first
matching overlay-list selection or creation, pending-condition write, stat-178 request and
update request. CBL-Q09 gates actual attach/list-write/update effects; CBL-Q04 leaves the
original count/row identity open. Supplied inputs permit local requests and return.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| CBL-R01 | Supplied notification selects one pair | verified in 1.13c code (static) | CBL-Q11: automatic notification | CBL-V12 |
| CBL-R02 | Dispatch heading | split | see sub-ids | see sub-ids |
| CBL-R02a | Apply events/order | verified in 1.13c code (static) | DPO-Q23/DPO-Q30: upstream apply integration; supplied stage eligible | CBL-V13 |
| CBL-R02b | Queued replacement/missile OR transfer | verified in 1.13c code (static) | none locally; producers CBL-Q04/CBL-Q05 | CBL-V14 |
| CBL-R03 | Guards and one attacker draw | verified in 1.13c code (static) | none at supplied getter boundary | CBL-V01 |
| CBL-R04 | Divisor heading | split | see sub-ids | see sub-ids |
| CBL-R04a | Type/condition/count divisor | verified in 1.13c code (static) | none at supplied predicate boundary | CBL-V02 to CBL-V08, CBL-V15 |
| CBL-R04b | Monster-condition producer | unresolved | CBL-Q01: original assignment/classification | none |
| CBL-R05 | Bonus table/wrapped multiplication | verified in 1.13c code (static) | none | CBL-V04, CBL-V05, CBL-V08, CBL-V15 |
| CBL-R06 | Signed fraction and resistance | verified in 1.13c code (static) | none at supplied getter boundary | CBL-V06, CBL-V10, CBL-V11, CBL-V15 |
| CBL-R07 | Requested life/floor/death flag | verified in 1.13c code (static) | CBL-Q10/RLI-Q03: stored replacing-write effects | CBL-V02, CBL-V09, CBL-V15 |
| CBL-R08 | Overlay call and return 1 | verified in 1.13c code (static) | CBL-Q09: stored overlay/update effects; supplied count eligible | CBL-V01 to CBL-V11, CBL-V15; LSE-V11 |

## 5. Outputs and side effects

| Output | Rule | Detail |
|---|---|---|
| Registration pair/traversal | CBL-R01 | identity `(136,0)`, events 6 then 5; IEP-R05/IEP-R07 removal |
| Dispatch/record preparation | CBL-R02a/CBL-R02b | owners/order and queued replacement or missile OR transfer |
| Attacker seed | CBL-R03 | one step after both unit guards when chance positive |
| Replacing-write request | CBL-R07 | defender stat 6/layer 0, value `new`; stored effects gated |
| Result death flag | CBL-R07 | set when `new <= 0` and result exists; other flags retained |
| Overlay operation and requests | CBL-R08 | defender/id 147; LSE-R13 list/pending/update requests |
| Return | CBL-R03/CBL-R08 | early 0, completed success 1; arithmetic fault has no normal return |

No direct client message, timed-state/timer scheduling, unit creation/destruction or save write
occurs in this callback. Overlay list creation belongs to LSE-R13.

## 6. Branches and edge cases

- Event 6 doubles every type's divisor, including a supplied non-missile route; other ids do
  not double. Queued dispatch does not require the enabled condition.
- Players/hirelings use 10 before doubling and skip count scaling. Ordinary PvP damage scaling
  does not scale this direct current-life calculation. Difficulty is not a callback input.
- Non-hireling monsters use 4 or 8 then count scaling. Missing class data gives boss false;
  missing monster data gives the separate condition false.
- Counts below 1 become 1. High counts are uncapped; multiplication, divisor add/doubling and
  life subtraction wrap. Divisor zero faults; a negative divisor can increase requested life
  or wrap its subtraction to a value floored to zero. Creation life bounds do not prove safe
  arithmetic for arbitrary supplied counts.
- Resistance above 100 clamps, nonpositive resistance does not amplify loss. Divide-first
  precision can retain a nonzero amount even at resistance 100.
- Zero/negative life with ordinary positive divisor floors requested life to 0, sets death
  when result exists and skips overlay. Tiny life can give amount 0: unchanged request,
  successful return, draw retained, no overlay.
- Missing result skips only death write. Null attacker/defender returns before reads/draws.
  Missing stat storage/definitions/entries follows RLI-R02a; actual write effects stay gated.
- Multiple callbacks use registration order, not a fixed Crushing Blow priority; callback
  mutation/deferred-removal behavior follows IEP-R07.

## 7. Open questions

| Id | Kind / status | Evidence that would settle it |
|---|---|---|
| **CBL-Q01** | outside-boundary / open | Monster kinds/creation paths assigning the separate divisor condition. Trace classification through setters and record condition with divisor; no local gate on supplied values. CBL-R04b remains unresolved. |
| **CBL-Q03** | observation / pending | Attributed gameplay sequence with Crushing Blow, Knockback and Open Wounds registered in known order; static traversal is IEP-R07. |
| **CBL-Q04** | outside-boundary / open | Overlay-count loader/row identity and original missile-condition/record producers, owned by LSE-Q03/MHD. Supplied values remain usable. |
| **CBL-Q05** | outside-boundary / open | Complete MHD-R01/MHD-R12 owner lookup/contact eligibility. The hand-off uses its resolved owner; supplied attacker remains eligible locally. |
| **CBL-Q07** | outside-boundary / open | Other non-queued callers' record preparation/eligibility. Their owners establish reachability; supplied event-6 behavior is fixed here. |
| **CBL-Q08** | outside-boundary / open | Networking owner traces stat/overlay sender, player notification and delivery; this callback sends no message directly. |
| **CBL-Q09** | integration / open | Stored overlay attach/list writes (STL-R06b/STL-R06c), update effects (RLI-Q02) and propagation. LSE-R13 establishes local search/creation and ordered requests. |
| **CBL-Q10** | integration / open | Actual replacing unit-stat write/propagation under RLI-Q03; requested value/order and death flag are independently verified. |
| **CBL-Q11** | integration / open | IEP-Q01/IEP-Q03 automatic notification production/order; supplied IEP-R03 operation remains verified. |
| **CBL-Q12** | observation / pending | Matching gameplay observations of verified assertions, including CBL-O01 to CBL-O03; hand calculations/table reads do not confirm gameplay. |

The [gap file](../plans/gaps/combat-crushing-blow.md) records attempts and blocked boundaries.
[Public gap assessment](../research/spec-evidence/combat-crushing-blow.md#gap-assessment)
distinguishes compatible evidence reuse from new readings.

## 8. Test vectors

### 8.1 Executable vectors

Every case starts fresh from its family's inputs/defaults; cases never inherit another case's
outputs. Unless stated, execute the supplied-input callback CBL-R03 to CBL-R08 with actual
NRF-R01/NRF-R05/RLI-R02a/RLI-R02b and supplied write/overlay boundaries. Life is raw 8.8.

- Attacker present, type player; defender present, type monster, non-hireling class 0,
  class/monster data present, boss false, separate divisor condition false, stat 100=1,
  stat 36=0, life 256000. Chance stat 136=100, event 5. Result present: hit true,
  suppressEvents/death false, unrelated flags retained.
- Getter definitions available through stat 178 (`statDefinitionCount=179`), `fMin=false`;
  selected existing arrays contain stated entries only; `usesResolvedStats=true`. These
  arrays are supplied, not constructed from items.
- `ReplaceLife` is a supplied fixture operation logging `(defender,6,0,new)` and persisting
  that value in the fixture's selected array. It does not execute the original setter or
  settle CBL-Q10. Thus each value called stored/read-back below is a fixture assertion only.
- A supplied overlay-boundary operation records `(defender,147)` after the life request/death
  write and returns 37; the callback ignores that value. It does not execute the helper's
  stored effects. LSE-V11 independently discriminates LSE-R13 search/allocation/ordered
  requests with supplied attach/list-write/update operations. These callback fixtures test
  the overlay call's arguments, condition and order, not CBL-Q09 integration.
- Default seed `{4,666}` draws `(attacker,100,10)`, ending `{0xAB1A45AE,1}`; completed success
  returns 1. Default result flags stay unchanged. Defender seed never changes locally.
- Additional NRF-R01 steps: `{1,666}` gives `{0x6AC6935F,0}`, roll 51; `{3,666}` gives
  `{0x4053B4E9,1}`, roll 25; `{5,666}` would give `{0x15E0D673,2}`, roll 99.
  The chance -5 case with that last seed makes no draw.

- **CBL-V01** Chance guards/threshold (CBL-R03, CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: default life 256000, chance/seed per case; each starts fresh.
  - Seeds: per case; defender seed unchanged.
  - Trace: on success divisor 4 gives amount 64000, requested life 192000.
  - (a) Inputs: chance 0, seed `{1,666}`. Draws: none. Outputs: return 0, seed unchanged, life/flags unchanged, no request.
  - (b) Inputs: chance 25, seed `{4,666}`. Draws: `(attacker,100,10)`. Outputs: seed `{0xAB1A45AE,1}`, life request 192000 then overlay `(defender,147)`, return 1, flags unchanged.
  - (c) Inputs: chance 25, seed `{3,666}`. Draws: `(attacker,100,25)`. Outputs: seed `{0x4053B4E9,1}`, return 0, life/flags unchanged, no write/overlay request.
  - (d) Inputs: chance 120, seed `{1,666}`. Draws: `(attacker,100,51)`. Outputs: seed `{0x6AC6935F,0}`, life request 192000 then overlay, return 1, flags unchanged.
  - (e) Inputs: chance -5, seed `{5,666}`. Draws: none. Outputs: return 0, seed/life/flags unchanged, no request.
  - (f) Inputs: attacker absent, defender otherwise default. Draws: none. Outputs: return 0, life/result unchanged, no write/overlay request.
  - (g) Inputs: defender absent, attacker otherwise default. Draws: none. Outputs: return 0, attacker seed `{4,666}`/result unchanged, no write/overlay request.
  - Discriminates: equality fails, nonpositive chance skips draws, chance above 100 still draws; null unit guards precede reads/draws.
  - runtime_observed: no
- **CBL-V02** Ordinary monster (CBL-R04a, CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defaults; life 256000 (`0x3E800`, 1000 HP).
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)`.
  - Trace: `d=4`, `n=1`, bonus 0; amount `256000/4=64000`, new `256000-64000=192000`.
  - Outputs: life request 192000 (fixture 750 HP) then overlay `(defender,147)`; default seed/return/flags.
  - (b) Inputs: defaults except `monsterDivisorFlag=true`, `boss=false`. Trace: divisor 8, bonus 0; amount 32000, new 224000. Outputs: life request 224000 then overlay; default seed/return/flags.
  - Discriminates: fraction uses current life; overlay defender/id follows the life request.
  - runtime_observed: no
- **CBL-V03** Event-6 divisor (CBL-R04a, CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defaults except event 6.
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)`.
  - Trace: divisor 4 doubles to 8; amount 32000, new 224000.
  - Outputs: life request 224000 (fixture 875 HP) then overlay; default seed/return/flags.
  - Discriminates: event 6 doubles divisor, halving this loss.
  - runtime_observed: no
- **CBL-V04** Boss/count scaling (CBL-R04a, CBL-R05, CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defaults except boss true, count 8.
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)`.
  - Trace: divisor 8; bonus 350; `mulDiv(350,8,100)=2800/100=28`; divisor 36; amount `256000/36=7111`, new 248889.
  - Outputs: life request 248889 then overlay; default seed/return/flags.
  - Discriminates: scaling acts on the boss divisor before life division, with truncation.
  - runtime_observed: no
- **CBL-V05** Count and event scaling (CBL-R04a, CBL-R05, CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defaults except count 3, event 6.
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)`.
  - Trace: bonus 100; `mulDiv(100,4,100)=4`; divisor 8 then doubled to 16; amount 16000, new 240000.
  - Outputs: life request 240000 (fixture 937.5 HP) then overlay; default seed/return/flags.
  - Discriminates: both count and event scaling affect divisor.
  - runtime_observed: no
- **CBL-V06** Player resistance (CBL-R04a, CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defender player, life 128000 (`0x1F400`, 500 HP), other defaults.
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)` each case.
  - Trace: divisor 10, initial amount 12800.
  - (a) Inputs: resistance 50. Trace: reduction `mulDiv(50,12800,100)=6400`; final amount 6400. Outputs: life request 121600 (fixture 475 HP), overlay, default seed/return/flags.
  - (b) Inputs: resistance 150. Trace: clamp 100, reduction 12800, amount 0. Outputs: unchanged-value life request 128000, no overlay, default seed/return/flags.
  - (c) Inputs: resistance -20. Trace: no reduction, amount 12800. Outputs: life request 115200 (fixture 450 HP), overlay, default seed/return/flags.
  - (d) Inputs: resistance 50, player count 3. Trace: player divisor stays 10, initial amount 12800, reduction 6400, final amount 6400. Outputs: life request 121600 (fixture 475 HP), overlay, default seed/return/flags.
  - (e) Inputs: resistance 0, event 6. Trace: player divisor 10 doubles to 20, amount 6400, new 121600. Outputs: life request 121600 (fixture 475 HP), overlay, default seed/return/flags.
  - Discriminates: resistance clamps but does not amplify for negatives; zero amount still writes/returns 1.
  - runtime_observed: no
- **CBL-V07** Hireling short circuit (CBL-R04a, CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defender monster class 338 `act2hire`, boss true, count 8, life 51200 (`0xC800`, 200 HP), event 6; other defaults.
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)`.
  - Trace: hireling divisor 10 skips boss/count and doubles to 20; amount 2560, new 48640.
  - Outputs: life request 48640 (fixture 190 HP) then overlay; default seed/return/flags.
  - Discriminates: boss/count cannot affect hireling divisor.
  - runtime_observed: no
- **CBL-V08** Count floor/table edge (CBL-R04a, CBL-R05, CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defaults except count per case.
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)` each case.
  - (a) Inputs: count 0. Trace: floor 1, bonus 0, divisor 4, amount 64000. Outputs: request 192000 then overlay; default seed/return/flags.
  - (b) Inputs: count 9. Trace: bonus `(9-2)*50=350`, increment 14, divisor 18, amount 14222. Outputs: request 241778 then overlay; default seed/return/flags.
  - (c) Inputs: count 10. Trace: bonus 400, increment 16, divisor 20, amount 12800. Outputs: request 243200 then overlay; default seed/return/flags.
  - (d) Inputs: count 8. Trace: table bonus 350, divisor 18, amount 14222. Outputs: request 241778 then overlay; default seed/return/flags.
  - (e) Inputs: count 2. Trace: table bonus 50, `mulDiv(50,4,100)=2`, divisor 6, amount `256000/6=42666`, new 213334. Outputs: request 213334 then overlay; default seed/return/flags.
  - Discriminates: 8/9 coincide; 10 uses formula; 0 floors to 1.
  - runtime_observed: no
- **CBL-V09** Tiny/zero/negative life (CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defaults except life per case.
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)` each case.
  - (a) Inputs: life 3. Trace: amount 0, new 3. Outputs: unchanged life request 3, no overlay, default seed/return/flags.
  - (b) Inputs: life 4. Trace: amount 1, new 3. Outputs: life request 3 then overlay, default seed/return/flags.
  - (c) Inputs: life 0. Trace: amount 0, new 0. Outputs: request 0, death true, hit true, suppressEvents false, no overlay; default seed/return.
  - (d) Inputs: life -256. Trace: amount -64, new -192 floors to 0. Outputs: request 0, death true, hit true, suppressEvents false, no overlay; default seed/return.
  - (e) Inputs: life 4, result.death initially true. Trace: amount 1, new 3. Outputs: life request 3 then overlay, death remains true, hit true, suppressEvents false; default seed/return.
  - (f) Inputs: life 0, result absent. Trace: amount 0, new 0. Outputs: life request 0, no result-flag write, no overlay; default seed/return.
  - Discriminates: death uses floored new life and is never cleared here; absent result skips only the flag write; unchanged write requested; nonpositive amount skips overlay.
  - runtime_observed: no
- **CBL-V10** Divide-first resistance (CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defaults except life 16777216 (`0x1000000`, 65536 HP), resistance 30.
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)`.
  - Trace: amount 4194304; `mulDiv(30,4194304,100)=(4194304/100)*30=1258290`; amount 2936014; new 13841202.
  - Outputs: life request 13841202 then overlay; default seed/return/flags.
  - Discriminates: multiply-first reduction differs; caller argument order selects this precision branch.
  - runtime_observed: no
- **CBL-V11** Resistance-100 remainder (CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defaults except life 16777216, resistance 100.
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)`.
  - Trace: amount 4194304; reduction `(4194304/100)*100=4194300`; final amount 4; new 16777212.
  - Outputs: life request 16777212 then overlay; default seed/return/flags.
  - Discriminates: resistance 100 retains positive loss in divide-first branch, unlike CBL-V06(b).
  - runtime_observed: no
- **CBL-V12** Supplied stat notification (CBL-R01; boundary: supplied-input stage)
  - Inputs: unit `u` has no registrations; available fixture definition `(136,0)` with event 5/selector 16 and event 6/selector 16. Supplied callback records invocation, reads the current supplied chance and returns 17; no other callbacks.
  - Seeds: no draw; recorder consumes none.
  - Trace: notify value 25 and supply chance 25; dispatch event 5 with owner `u`, other `v`, record `R`, game `G`; dispatch event 7; notify value 60 and supply chance 60, dispatch event 5 again; notify value 0.
  - Outputs: first notification creates `[event6,event5]`, both item kind/identity `(136,0)` and values `((136,0),0)`. First event 5 calls once with `(G,5,u,v,R,(136,0),0)`, reads chance 25, returns 17; event 7 calls none/returns 0. Value 60 adds no registration, retains identity/order, and the next event-5 call reads 60. Value 0 removes both. Seeds/record unchanged by recorder.
  - Discriminates: no pair per item/value increase; unrelated event skips; head insertion order and current getter value retained. Negative notification registration is independently discriminated by the consumed IEP-V02(e), not duplicated here.
  - runtime_observed: no
- **CBL-V13** Supplied apply stage (CBL-R02a; boundary: supplied-input stage)
  - Inputs: post-gate/post-mitigation stage, attacker `a`, defender `b`, record `R` hit true, suppressEvents/death false. Dispatch recorders return 23 for attacker, 31 for defender and do not mutate `R` unless a case overrides that behavior; stop at clamp entry.
  - Seeds: no draw; supplied recorders consume none.
  - Trace: evaluate result/route gates, dispatch in order and continue to clamp despite nonzero returns.
  - (a) Inputs: queuedMelee, enabled true, suppression false. Outputs: `(5,a,b,R)` then `(1,b,a,R)` then clamp.
  - (b) Inputs: nonQueued, enabled true, suppression false. Outputs: `(6,a,b,R)` then `(2,b,a,R)` then clamp.
  - (c) Inputs: nonQueued, enabled true, suppression true. Outputs: `(2,b,a,R)` then clamp.
  - (d) Inputs: nonQueued, enabled false, suppression false. Outputs: no dispatch, then clamp.
  - (e) Inputs: four fresh cases: each mode (`queuedMelee`, `nonQueued`) with each result pair (hit true/suppressEvents true, hit false/suppressEvents false); enabled true, suppression false in all four. Outputs: no dispatch, then clamp.
  - (f) Inputs: queuedMelee, enabled false, suppression true; default hit true/suppressEvents false. Outputs: `(5,a,b,R)` then `(1,b,a,R)` then clamp despite both nonzero dispatcher returns.
  - (g) Inputs: queuedMelee, enabled true, suppression false; attacker recorder logs event 5 while hit true, then clears `R.hit` and returns 23; defender recorder logs its received hit value and returns 31. Outputs: `(5,a,b,R,hit=true)` then `(1,b,a,R,hit=false)` then clamp; final `R.hit=false`, suppressEvents/death false, other fields/seeds unchanged.
  - Outputs: except the explicit mutation in (g), all result/record fields and seeds are unchanged. Cases (a), (b), (f) and (g) reach the attacker dispatch and can invoke a supplied matching attacker Crushing Blow registration; (c), (d) and all four (e) cases do not. These recorders do not execute its callback effects.
  - Discriminates: owners swap, attacker precedes defender, non-queued attacker suppression retains defender event; queued dispatch ignores the two record conditions, nonzero dispatcher return does not stop apply, and the outer result gate is not rechecked after the attacker callback.
  - runtime_observed: no
- **CBL-V14** Record-condition transfer (CBL-R02b; boundary: supplied-input stage)
  - Inputs: supplied record includes unrelated flag `X`; conditions per case. No callback executes.
  - Seeds: no draw.
  - Trace: queued replacement differs from missile OR transfer.
  - (a) Inputs: queuedMelee, enabled false, suppression true, X true. Outputs: enabled true, suppression false, X false.
  - (b) Inputs: missile contact, enabled false, suppression false, X true; both missile conditions true. Outputs: enabled true, suppression true, X true.
  - (c) Inputs: missile contact, enabled true, suppression true, X true; both missile conditions false. Outputs: enabled true, suppression true, X true.
  - (d) Inputs: missile contact, incoming enabled false, incoming suppression false, X true; missile enablesItemEvents true, suppressesAttackerEvent false. Outputs: enabled true, suppression false, X true.
  - Discriminates: OR retains existing conditions; queued replacement clears unrelated flags.
  - runtime_observed: no
- **CBL-V15** Extreme supplied player counts (CBL-R04a, CBL-R05, CBL-R06, CBL-R07, CBL-R08; boundary: supplied-input stage)
  - Inputs: defaults except count/life per case; arithmetic fixtures do not assert that ordinary creation produces these counts.
  - Seeds: `{4,666}`. Draws: `(attacker,100,10)` before arithmetic in each case.
  - (a) Inputs: count 2147483647, life 256000. Trace: bonus `low32((2147483647-2)*50)=-150`; `mulDiv(-150,4,100)=-6`; divisor -2, amount -128000, requested life 384000. Outputs: life request 384000, no overlay, flags unchanged, seed `{0xAB1A45AE,1}`, return 1.
  - (b) Inputs: count 2147483647, life 2147483392 (`0x7FFFFF00`). Trace: divisor -2, amount -1073741696; subtraction wraps to -1073742208, floors to 0. Outputs: life request 0, death true, hit true, suppressEvents false, no overlay; seed `{0xAB1A45AE,1}`, return 1.
  - (c) Inputs: count 1610612736, life 256000. Trace: bonus -1073741924; small-operand product `low32(-1073741924*4)=-400`, scaled result -4; divisor 0. Completed prefix: chance draw and life read. Outputs: signed-division fault before resistance read, life/overlay request or death write; no normal return. Completed draw produced `{0xAB1A45AE,1}` before fault. No post-fault recovery/state assertion.
  - (d) Inputs: count 515396075, life -2147483648 (`INT_MIN`). Trace: bonus `low32((515396075-2)*50)=-126`; `low32(-126*4)=-504`, signed division by 100 truncates to -5; divisor -1. Completed prefix: chance draw and life read. Outputs: signed-quotient-overflow fault before resistance read, life/overlay request or death write; no normal return. Completed draw produced `{0xAB1A45AE,1}` before fault. No post-fault recovery/state assertion.
  - Discriminates: widening count multiplication or assuming positive divisor rejects the exact wrap/fault outcomes.
  - runtime_observed: no

### 8.2 Observation vectors

None has been performed. The [observation procedures](../research/spec-evidence/combat-crushing-blow.md#observation-procedures)
identify setup, measurements, original locations and limits.

| Id | Rules | Observation target |
|---|---|---|
| **CBL-O01** | CBL-R01, CBL-R02a, CBL-R02b, CBL-R03 | Callback event/owner, one attacker seed step, life change before ordinary subtraction for melee/eligible missile |
| **CBL-O02** | CBL-R08 | Positive-amount overlay call and separately actual list/stat/update effect; zero amount skips call |
| **CBL-O03** | CBL-R04a, CBL-R05 | Plain-monster divisor at counts 1/8; champion/unique condition recorded without assuming classification |

## 9. Acceptance and dependencies

| Boundary | Rules / vectors | Acceptance |
|---|---|---|
| Supplied callback | CBL-R03, CBL-R04a, CBL-R05 to CBL-R08; CBL-V01 to CBL-V11, CBL-V15 | Guards, draw/no-draw/seeds, wrapped arithmetic, requested life, death flag, overlay call condition/arguments/order and return or arithmetic fault |
| Supplied notification/apply | CBL-R01, CBL-R02a, CBL-R02b; CBL-V12 to CBL-V14 | Registration identity/order/count, local transfer/dispatch; no automatic notification/full missile-owner claim |
| Stored integration | CBL-Q09, CBL-Q10, CBL-Q11 | Complete write/attach/update/notification contracts from owners |
| Producers/observation | CBL-R04b, CBL-Q01, CBL-Q03 to CBL-Q05, CBL-Q07, CBL-Q08, CBL-Q12 | Original producer/reachability and attributed observations; supplied fixtures do not settle them |

Passing verified subsets is partial reproduction. Full original-route reproduction needs its
applicable integration/producers and observations; no user approval is inferred.

**Depends on:** [numeric foundations](core-numeric-and-rng-foundations.md) NRF-R01/NRF-R05 for
RNG/arithmetic, NRF-R08a for units; [record layouts](core-record-layouts-and-ids.md)
RLI-R02a/RLI-R02b for existing reads; [game-data fields](core-game-data-fields.md)
GDF-R02/GDF-R08 and [game-data table records](core-game-data-table-records.md)
GDA-R12/GDA-R08a/GDA-R04a for named values/record selection; [item events](combat-item-event-triggers-and-procs.md)
IEP-R01/R02/R03/R05/R07 for registration/dispatch; [damage pipeline](combat-damage-pipeline-order.md)
DPO-R13a/DPO-R23 for placement; [leech/drain](combat-leech-and-drain.md) LSE-R13 for
overlay requests, consuming STL-R06a/STL-R06b/STL-R06c and RLI-Q02 with their gates. Actual unit
replacing writes remain RLI-Q03.

**Depended on by:** [damage pipeline](combat-damage-pipeline-order.md) for Crushing Blow effects;
[item events](combat-item-event-triggers-and-procs.md) IEP-R12a for selector 16's catalog entry;
[numeric foundations](core-numeric-and-rng-foundations.md) NRF-R04c/NRF-Q04, whose per-handler draw contracts include this callback's single draw (CBL-R03).
Reverse consumers/supporting citations are not automatic planning prerequisites.

## 10. Related specs

- [Monster scaling](monster-scaling-by-players-and-difficulty.md): MSD-R09 produces player count, not required for a supplied count.
- [Missile hit/damage](missile-hit-and-damage-calculation.md): MHD-R01/MHD-R12 owns original contact/owner setup.
- [Missile movement](missile-movement-and-collision.md): MMC-R01/MMC-R03/MMC-R14 supplies the contact root.
- [Cast dispatch](skill-cast-start-and-do-dispatch.md): SCD-R06/SCD-R09b supplies the melee root.
- [Open Wounds](combat-open-wounds.md) and [Knockback](combat-knockback.md): separate item-event effects.
- [Mitigation](combat-damage-mitigation.md): ordinary physical resistance differs from this life-fraction scaling.
- [State/list lifecycle](state-timed-state-lifecycle.md): list allocation/attach/write owner consumed by overlay.

## Appendix A. Provenance

Evidence build: original Expansion 1.13c, Windows x86. D2Game, D2Common, Fog and Storm artifacts
are identified by the [source manifest](../research/source-manifest.md). The
[public companion](../research/spec-evidence/combat-crushing-blow.md#provenance) maps roots,
imports and decisions to original functions independently of private research files.

Static instruction readings dated 2026-10-13 are reused within recorded boundaries;
2026-10-14/15 shared getter/item-event/overlay contracts supply current boundaries. Targeted
2026-10-15 instruction/image checks establish wrapping and distinguish local requests from
unverified stored effects. Named cells were read through the original loader on 2026-10-13;
effective-file hashes, complete overrides, physical supplying archives and compiled-table
equivalence were not captured for that record. No gameplay observation confirms a rule.
The Id map and gap file retain pending boundaries.
