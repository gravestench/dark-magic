# Player stamina drain while running

**Description**

Running outside town spends stamina. How quickly it is used depends on your character, the armor worn on your torso and effects that reduce stamina drain.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: SDR-R, SDR-Q, SDR-V, SDR-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/player-stamina-drain-while-running.md)
Open evidence gaps: [gap file](../plans/gaps/player-stamina-drain-while-running.md)

## 1. Scope

This specifies the server's running-stamina consumer: dispatch eligibility, town exemption, class drain, torso weight, percentage reduction, minimum amount, ordered stat requests, post-add stamina test, walking request and conditional cadence. Supplied post-preparation inputs and stat-read results permit verified local behavior independently of unresolved stat propagation and recurring-event integration. Spending Vitality and producing regeneration are outside this boundary. Player mode changes own the complete transition, path processing and movement bookkeeping; stat lifecycle owns stored propagation; network behavior owns client delivery.

## 2. Trigger and ownership

D2Game computes the drain among its player-mode routines; no source-path assertion identifies the computing function's folder. Each server frame calls the frame-event pump. Its player group has a persistent list and a due-frame list. Type-0 events without a direct callback select player-mode dispatch. Supported modes 0..19 include Walk (2), Run (3), Town Walk (6) and Knock back (19), which select movement. At the supplied post-preparation stage only Run calls drain. Drain return 0 requests Walk, then path processing continues. The movement handler's own result controls enclosing mode-end dispatch.

A reached mode-setter path requests pending-event removal and persistent registration. Complete creation/join/nonfresh/reentrant lifecycle remains SDR-Q01/SDR-Q04. A supplied surviving persistent entry is visited once per frame by its list pass; this does not establish one live entry for every running player. Local drain precedes the path step. The separate resource event orders life, stamina and mana through OWD-R11a; stamina internals and relative order against running drain remain SDR-Q06. Drain directly sends no message; actual client updates are SDR-Q03/SDR-Q04.

## 3. Inputs and data

Integer additions, products and negation retain low 32 bits as signed i32; signed division truncates toward zero. Stamina is raw 8.8: 256 units represent one stamina point (NRF-R08a). Local lookup, arithmetic and request routing make no RNG draw; unresolved dependencies have no whole-operation draw guarantee.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| player/mode | unit presence, type 0; GDF-R14a | logical; mode u32, supported 0..19 | Run 3 permits call; Walk 2 requested on return 0 |
| game/current unit identity | supplied game context G and current unit U | logical identities | All unit/mode/path requests retain these exact identities; isolated wrapper cases may supply a nonplayer type without asserting a nonplayer drain |
| room/level | supplied dynamic path's room and valid linked level | optional room; level i32 | Absent path gives absent room; town ids 1,40,75,103,109 |
| class/count/record | prepared CharStats, GDF-R03a | i32 index/count; optional computed record | Valid index required; absent collection is not a general safe-null case |
| RunDrain | CharStats.RunDrain | u8 | Base amount is twice this value |
| inventory | absent, invalid or valid inventory | logical | Absence stops drain; invalid inventory gives no torso item |
| body page/allocations | prepared dimensions/slots; supplied successful allocation when required | logical page; dimensions u8 | 13 by 1; only slot 3 contributes; no safe allocator-failure fallback |
| torso/class/record | body slot 3; prepared items collection | class u32; optional record | Out-of-range class gives absent record; valid class needs present collection |
| speed | items record speed | i32 | Torso weight |
| percent p | stat 154 item_staminadrainpct, layer 0 | i32 getter result | NRF-R12b/NRF-R12c result used unchanged |
| post-add stamina | stat 10 stamina, layer 0 | i32 getter result | Read after addition request; positivity decides return |
| inner replacement result | supplied result from the list replacement operation | i32 | Unit wrapper requests tracking exactly when nonzero and unit is player; this is distinct from the whole wrapper's ignored return |
| stat definitions/arrays | NRF-R12b/NRF-R12c | existing ordered arrays/definitions | Select own/resolved arrays, without building/summing contributions |
| own stat list | supplied current-unit own-list presence/identity | logical list or absent | Mutation wrappers forward to this list; actual storage/propagation remains SDR-Q02 |
| events/callbacks | lists, type, frame, target, A/B, direct callback, running/removal conditions | logical; type u8, frame i32 | Supplied frame-event traversal boundary |
| handler result | selected mode handler | supplied i32 | Exactly 2 requests mode-end processing |
| movement/transition inputs | post-preparation unit, path/arrival results, continuation condition, room, skill context | supplied logical/numeric state | Full producers SDR-Q04/SDR-Q07 |
| cadence/writers | invocation count, rate and no intervening writer | supplied frames/i32 | Duration fixture supplies 25 fps; live invariant rate remains NRF-Q03 |

| Mapping/constant | Origin | Binding / named support |
|---|---|---|
| stat 10/154, mode 2/3 | Engine selects ids; GDF names are table data | Fixed consumers; ItemStatCost.Stat explicit IDs and PlrMode row names |
| torso location 3 | Engine code | Fixed selector; BodyLocs row 3 Code=tors corroborates name |
| factor 2, divisors 10/-100, floor 1 | Engine code | Fixed arithmetic constants |
| RunDrain 20/20/15 | Table data | CharStats rows Amazon/Barbarian/Assassin, column RunDrain; qualified records below; fixtures supply values |
| speed 0/5/10 | Table data | Armor rows Quilted Armor/Ring Mail/Full Plate Mail, column speed; qualified records below; fixtures supply values |
| town membership / mode handlers | Engine code | Fixed original selectors; no custom remapping or mode-index clamp established |

Named cells and profile limits are in the [companion](../research/spec-evidence/player-stamina-drain-while-running.md#named-table-cells). Code establishes record-field meaning separately from text-cell values.

## 4. Rules (execution order)

**SDR-R01** Dispatch to drain (subrules).

- **SDR-R01a** Frame entry [verified in 1.13c code (static)]. Each supplied frame pass calls the frame-event pump once. This establishes a call, not live-player registration or invariant elapsed time.
- **SDR-R01b** Player-group pass [verified in 1.13c code (static)]. Dispatch persistent player list without due-frame test, then current bucket with that test enabled. Only bucket entries whose frame equals currentFrame dispatch. Before each callback save successor and mark running; afterwards clear running. Remove dispatched bucket entries; remove persistent entries only when pending-removal is set. Continue at saved successor, including removal's changes to it. This local traversal accepts supplied lists/callback results; arbitrary callback lifecycle/population remain SDR-Q01.
- **SDR-R01c** Event selection [verified in 1.13c code (static)]. Direct callback receives `(game,target,type,A,B)`. Otherwise type<15 selects the fixed handler; absent handler or type>=15 skips call. Type 0 calls player-mode dispatch `(game,target,A,B)`; type 1 selects mode-end processing with the same arguments. List pass ignores callback results.
- **SDR-R01d** Recurring registration [unresolved]. Reached mode-setter path requests pending-event removal then persistent type-0 registration. Complete creation/join/nonfresh/reentrant population is SDR-Q01. MCI-R12b's fresh allocation association does not establish complete recurring dispatch.
- **SDR-R01e** Mode selection [verified in 1.13c code (static)]. Supplied supported mode 0..19 selects fixed handler; absent unit selects mode 0. Null handler skips call. Modes 2,3,6,19 select movement. Forward `(game,unit,A,B)` unchanged. Handler result exactly 2 requests mode-end processing with same arguments. Original indexing has no bounds test; no safe out-of-domain fallback is established.
- **SDR-R01f** Running gate [verified in 1.13c code (static)]. At supplied post-preparation stage, call drain only for present unit/mode 3. Return 0 requests mode 2 with that unit/game; then continue path processing. Drain result does not become the movement handler's final return. Full preparation/path/arrival/history effects are SDR-Q07; they are not assumed inert.

**SDR-R02** Town [verified in 1.13c code (static)]. Direct absent-unit drain returns 0. Otherwise obtain current room: absent dynamic path gives absent room, which is not town. For valid present room/level records, ids 1,40,75,103,109 are town and every other i32 id is not. Town returns 1 before class/inventory/stat work. Malformed linked records have no safe-null contract.

**SDR-R03** Class [verified in 1.13c code (static)]. Outside town return 0 for class<0, class>=classCount or absent computed record. No RunDrain read/stat request. Valid index into absent collection is not promised safe failure.

**SDR-R04** Base [verified in 1.13c code (static)]. Read RunDrain as u8; `w=2*RunDrain` in i32 raw stamina units. Add no fixed-point shift.

**SDR-R05** Inventory/torso (subrules).

- **SDR-R05a** Inventory absence [verified in 1.13c code (static)]. Absent inventory returns 0 before torso lookup, percent read or stat request.
- **SDR-R05b** Existing body page [verified in 1.13c code (static)]. Invalid inventory, location outside 0..12 or existing page dimensions unequal to 13 by 1 gives no item. With valid inventory/existing prepared slots read only slot 3. Getter absence allows unweighted drain; it differs from absent inventory.
- **SDR-R05c** Lazy page [verified in 1.13c code (static)]. With successful supplied allocations, absent page grows collection to one page with 13 by 1 dimensions/13 empty slots. Existing page lacking slots receives empty slots and those dimensions. No item created. Allocation/lifetime integration is SDR-Q07; no safe failure return established.

**SDR-R06** Weight [verified in 1.13c code (static)]. Look up present torso class by unsigned class<count. Out-of-range class gives absent record. Valid class requires present collection; missing collection is fatal rather than absent-record return. Present record: `m=low32(trunc(speed/10)+1)`; `w=low32(w*m)`. Absent item/record leaves w unchanged. Speeds 0/5 give multiplier 1, 10 gives 2, -10 gives 0.

**SDR-R07** Percent [verified in 1.13c code (static)]. Read stat 154/layer 0. If p!=0, `product=i32(low32(p*w))`; `w=i32(low32(w+trunc(product/-100)))`. Otherwise skip adjustment. Product wraps before division; p>=100 does not universally imply floor after overflow.

**SDR-R08** Minimum [verified in 1.13c code (static)]. If signed w<1 set w=1. Final amount is at least one raw unit.

**SDR-R09** Add request [verified in 1.13c code (static)]. Request `(unit,stat 10,delta=low32(-w),layer 0)`. Wrapper forwards nonzero delta to own list, including absent list. Ignore its return; test only following read. Actual addition/clamp/resolved-array production/propagation is SDR-Q02. No intermediate stored value asserted.

**SDR-R10** Post-add decision (subrules).

- **SDR-R10a** Positive read [verified in 1.13c code (static)]. After R09 read stat 10/layer 0 through NRF-R12b/NRF-R12c. Supplied signed result>0 returns 1 without zero replacement.
- **SDR-R10b** Zero request [verified in 1.13c code (static)]. Supplied post-add result<=0 requests `(unit,stat 10,value 0,layer 0)`, ignores the whole unit-replacement wrapper's return and returns 0. At the separate wrapper-routing stage, forward the replacement to own list. If the supplied inner list-replacement result is any nonzero i32 and unit type is player, request changed-stat tracking for that list/stat/layer; otherwise do not. The wrapper's final numeric return is not used by the drain and is not equated to its inner result. Actual storage, the producer of the inner result, tracking effects and client delivery remain SDR-Q02/SDR-Q03; no notification default prescribed.

**SDR-R11** Walking transition [unresolved]. R01f establishes mode-2 request. Transition consults town status to select Town Walk 6 or Walk 2, performs path/skill preparation and branches on continuation condition. Complete stop branch/mode setter/event removal/registration/path outcomes/messages remain SDR-Q04. Drain return 0 alone does not prove final mode. New Run refusal at zero belongs to player-mode transitions.

**SDR-R12** Conditional duration [verified in 1.13c code (static)]. Given one invocation per supplied frame, constant positive w, supplied arithmetic persistence, no overflow/intervening writer, raw stamina S depletes on `max(1,ceil(S/w))` invocations. At supplied 25 fps seconds=count/25. Actual recurrence is SDR-Q01, storage SDR-Q02, writer interference SDR-Q06. This conditional result verifies none of those integrations.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SDR-R01a | Frame call | verified in 1.13c code (static) | none | SDR-V08 |
| SDR-R01b | Supplied traversal | verified in 1.13c code (static) | SDR-Q01: full recurrence only | SDR-V08 |
| SDR-R01c | Event selection | verified in 1.13c code (static) | none | SDR-V08 |
| SDR-R01d | Registration integration | unresolved | SDR-Q01: lifecycle | none |
| SDR-R01e | Mode selection | verified in 1.13c code (static) | none | SDR-V08 |
| SDR-R01f | Run/Walk request | verified in 1.13c code (static) | SDR-Q07: full caller only | SDR-V05, SDR-V08, SDR-V11, SDR-V12 |
| SDR-R02 | Town | verified in 1.13c code (static) | none | SDR-V06 |
| SDR-R03 | Class | verified in 1.13c code (static) | none | SDR-V11 |
| SDR-R04 | Base | verified in 1.13c code (static) | none | SDR-V01, SDR-V03 |
| SDR-R05a | Inventory absence | verified in 1.13c code (static) | none | SDR-V12 |
| SDR-R05b | Existing torso | verified in 1.13c code (static) | none | SDR-V01, SDR-V02, SDR-V12 |
| SDR-R05c | Supplied-success page | verified in 1.13c code (static) | SDR-Q07: allocation only | SDR-V12 |
| SDR-R06 | Weight | verified in 1.13c code (static) | none | SDR-V02, SDR-V03, SDR-V10 |
| SDR-R07 | Percent | verified in 1.13c code (static) | none | SDR-V03, SDR-V04, SDR-V07 |
| SDR-R08 | Floor | verified in 1.13c code (static) | none | SDR-V04, SDR-V10 |
| SDR-R09 | Add request | verified in 1.13c code (static) | SDR-Q02: storage only | SDR-V01 |
| SDR-R10a | Supplied read test | verified in 1.13c code (static) | SDR-Q02: read production only | SDR-V01, SDR-V05 |
| SDR-R10b | Zero request | verified in 1.13c code (static) | SDR-Q02: storage; SDR-Q03: delivery | SDR-V05 |
| SDR-R11 | Full transition | unresolved | SDR-Q04: mode/path/messages | SDR-V05 (request only) |
| SDR-R12 | Conditional duration | verified in 1.13c code (static) | SDR-Q01/SDR-Q02/SDR-Q06: live duration only | SDR-V09 |

## 5. Outputs and side effects

| Output/effect | Rule / boundary |
|---|---|
| Final amount w; drain return 0/1 | SDR-R02 through SDR-R10 |
| Ordered stat-10/layer-0 Add, Read, conditional unit Replace and its own-list forwarding | SDR-R09/SDR-R10; storage SDR-Q02 |
| Conditional changed-stat tracking request | SDR-R10b; tracking/delivery SDR-Q03 |
| Mode-2 request followed by path continuation | SDR-R01f; actual transition SDR-R11/SDR-Q04 |
| Frame pump/selected callbacks/conditional mode-end requests | SDR-R01a/SDR-R01b/SDR-R01c/SDR-R01e |
| Running/removal conditions and successor progression | SDR-R01b; lifecycle SDR-Q01 |
| Conditional page growth/13 empty slots under supplied success | SDR-R05c; allocation SDR-Q07 |
| Conditional invocation count/seconds | SDR-R12; live duration gates above |
| Fixture-only stamina variable: supplied post-add read, then 0 when the supplied Replace model runs | Synthetic state of SDR-V01 through SDR-V07, SDR-V09, SDR-V10 and SDR-V12; supports SDR-R09/SDR-R10/SDR-R12 supplied examples only. This is not an asserted original storage write; actual storage remains SDR-Q02 |

Drain directly creates no unit, changes no state, schedules no event and sends no message. Surrounding preparation/history/transitions/registration/tracking have separate effects and gates.

## 6. Branches and edge cases

Town returns before invalid class/absent inventory. Outside town class failure or absent inventory returns 0 without stamina requests. Invalid inventory and out-of-range torso class instead allow unweighted drain. Malformed linked records, absent indexed collections and allocation failure have no invented safe fallback.

Modes 2/6/19 skip drain; supported table domain is 0..19. Unchecked indexing gives no portable malformed-mode result. Direct absent-unit drain returns 0 independently of the movement gate that skips it.

RunDrain=0 still floors to 1. Negative speed/percent are synthetic arithmetic inputs, not established shipped cells. Armor/percent products wrap. Negative requested delta does not prove globally monotonic stored stamina under overflow/unresolved propagation. Absent stat list gives getter 0; mutation still reaches its dependency. No difficulty/PvP/melee/missile/classic/expansion selector is read locally; evidence is Expansion 1.13c.

## 7. Open questions

| Question | Kind / exact boundary | Evidence that would settle it |
|---|---|---|
| **SDR-Q01** | integration: full recurring population, creation/join/nonfresh/reentrant registration | Complete lifecycle/one-entry-per-running-player invariant; local removal/association alone does not prove it |
| **SDR-Q02** | integration: Add/Replace storage, intermediate clamp, own/resolved-array production, final zero | Complete additive propagation and NRF-Q20/STL-R06c; distinguish request/storage/read |
| **SDR-Q03** | integration: changed-stat tracking to actual client stamina update | Complete tracking-to-message-to-client path |
| **SDR-Q04** | integration: full Walk transition/zero-stamina new-Run refusal | Complete preparation/stop branches/mode write/registration/messages |
| **SDR-Q05** | observation/data-profile qualification: gameplay unobserved; recent cells lack independent complete vanilla-profile identity | Matching owned 1.13c observations and profile/override identification |
| **SDR-Q06** | outside-boundary: other writers/relative scheduling; none locally for supplied read/no-writer fixture | Trace separate stamina resource step/Run gate/relative order; OWD-R11a establishes life→stamina→mana order only |
| **SDR-Q07** | integration: full preparation/path/history and page allocation/lifetime | Complete named dependencies; supplied post-preparation arithmetic/allocation success available |

Attempts are in the [gap file](../plans/gaps/player-stamina-drain-while-running.md).

## 8. Test vectors

### 8.1 Executable vectors

Defaults: present player, class 4/Barbarian, classCount=7, valid computed class record/RunDrain=20, mode 3, valid room level 2, valid present inventory/existing 13 by 1 page, no torso, initial stamina=23552 (92.0), p=0. Cases independently start with defaults plus overrides. Amounts are raw 8.8; 24320 represents 95.0 and 21504 represents 84.0. Event/mode/path operations log calls and perform no unspecified effects.

Call-log notation retains identities: G is the exact supplied game and U the exact current unit, called player in applicable cases. `WalkRequest(m)` expands to `WalkRequest(game=G,unit=U,mode=m)`; `PathStep` to `PathStep(game=G,unit=U)`; `Drain` to `Drain(unit=U)`. Unit `Add`, `Read` and `Replace` shorthand always target U, and list operations target the explicitly named list L. Each expected call asserts argument identity as well as values and order: a correct mode requested for any different game or unit fails. These are semantic logs, not required APIs or original calling conventions.

V01..V07/V10 use supplied fixture persistence: log Add, supply post-add getter result `i32(low32(initialStamina+delta))`, and let Replace set a fixture stamina variable to 0. V05 also supplies distinct read outcomes. V09 repeats this model. These are conditional fixture outputs, not verified original storage. Every vector has no local RNG draw; dependency interiors are supplied.

In V08 list cases, entries have type 0, target=player, A=11/B=22, no direct callback, frame=100, running=false and pending-removal=false unless overridden. Successors follow the listed order. Callback execution is supplied: it logs the event identity, makes no draw and changes no list/stat state except the explicit case override. Isolated selectors stop at their logged dependency call; they do not execute an unspecified movement handler. In V12, itemCount=10 and the items collection is present when an item is supplied.

V05(a)..(d) stop at the unit-replacement wrapper request: the wrapper is supplied as a logged dependency and its inner result/tracking are not executed or asserted. V05(e)..(h) separately execute the wrapper-routing stage with explicit inner results; these isolated cases assert request logs only, without a fixture stamina assignment. A tracking request is logged where stated; its actual effects remain gated. In the drain-stage cases the fixture stamina assignment is a supplied persistence effect, not an assertion that the original list writer stores zero.

- **SDR-V01** Base (SDR-R04/SDR-R05b/SDR-R09/SDR-R10a; boundary: supplied-input stage)
  - Inputs: defaults; read=23512; independently Add return=0 or 2.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: 2*20=40; no torso/p=0; delta=-40; read>0.
  - Outputs:
    - (a) Default drain stage, independently Add return=0 or 2: w=40; `[Add(10,-40,0),Read(10,0)=23512]`; drain=1, fixture stamina=23512 (91.84375); no Walk/direct message for either Add return.
    - (b) Isolated nonzero-add wrapper stage with current unit U/own list absent, stat 10/delta=-40/layer 0: exactly `ListAdd(absent,10,-40,0)` is forwarded. No getter/drain result, fixture stamina or actual stored effect asserted at this wrapper-only boundary.
  - Discriminates: rejects extra fixed-point shift/using Add return as continuation; (b) rejects silently suppressing the forwarded request when own list is absent.
  - runtime_observed: no
- **SDR-V02** Heavy torso (SDR-R05b/SDR-R06; boundary: supplied-input stage)
  - Inputs: defaults, Full Plate Mail speed=10 in torso, irrelevant helm speed=100, read=23472.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: multiplier=2; w=80; 23552-80=23472.
  - Outputs: w=80/delta=-80/read=23472/fixture stamina=23472 (91.6875)/drain=1; no Walk.
  - Discriminates: helm does not contribute; applying it disagrees with 80.
  - runtime_observed: no
- **SDR-V03** Medium/percent (SDR-R04/SDR-R06/SDR-R07; boundary: supplied-input stage)
  - Inputs: class 6/Assassin, RunDrain=15, Ring Mail speed=5 in torso, p=50, initial stamina=24320, read=24305; other defaults.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: base=30; trunc(5/10)+1=1; adjustment=-15; w=15.
  - Outputs: w=15/delta=-15/read=24305/fixture stamina=24305/drain=1; no Walk.
  - Discriminates: multiplier is 1, not fractional 1.5 or rounded 2.
  - runtime_observed: no
- **SDR-V04** Floor (SDR-R07/SDR-R08; boundary: supplied-input stage)
  - Inputs: class 0/Amazon, RunDrain=20, Quilted Armor speed=0, p=100, initial stamina=21504, read=21503; other defaults.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: 40*1; adjustment=-40; floor 0 to 1.
  - Outputs: w=1/delta=-1/read=21503/fixture stamina=21503/drain=1; no Walk.
  - Discriminates: 100 percent still requests one raw unit.
  - runtime_observed: no
- **SDR-V05** Depletion/order (SDR-R01f/SDR-R10a/SDR-R10b; boundary: supplied-input movement stage)
  - Inputs: initial stamina=30/w=40; other defaults; Walk/PathStep log only.
  - Seeds: `no draw`.
  - Draws: none locally.
  - Trace: Add -40 precedes supplied signed read; Replace precedes return 0/Walk.
  - Outputs:
    - (a) Read=-10: `[Add(10,-40,0),Read=-10,Replace(10,0,0),WalkRequest(2),PathStep]`; drain=0/fixture stamina=0.
    - (b) Read=0: same log with read=0; drain=0/fixture stamina=0.
    - (c) Read=7: `[Add(10,-40,0),Read=7,PathStep]`; drain=1/fixture stamina=7; no Replace/Walk.
    - (d) Read=-10/whole unit-replacement wrapper return=2: same log as (a), drain=0/fixture stamina=0. The supplied wrapper does not execute its interior; no inner-result or tracking assertion is made in this case.
    - (e) Isolated wrapper routing: present player/own list L, stat 10/value 0/layer 0, supplied inner result=0. Log `[ListReplace(L,10,0,0)]`; no Track request. No wrapper numeric return or actual stored result asserted.
    - (f) Same isolated inputs as (e), independently supplied inner result=1: log `[ListReplace(L,10,0,0),Track(L,10,0)]`; exactly one Track request; no wrapper numeric return or actual stored result asserted.
    - (g) Same isolated inputs as (e), independently supplied inner result=2: log `[ListReplace(L,10,0,0),Track(L,10,0)]`; exactly one Track request; no wrapper numeric return or actual stored result asserted.
    - (h) Isolated wrapper with current unit U/type 1/own list L, stat 10/value 0/layer 0, inner result=1: `[ListReplace(L,10,0,0)]`; no Track request. No wrapper numeric return, fixture stamina, nonplayer drain or actual stored result asserted.
  - Discriminates: pre-add 30 fails (a); assuming subtraction instead of read fails (c); whole-wrapper result does not cancel path. Cases (e)..(h) distinguish inner zero/nonzero routing and player-type guard, including nonzero other than 1, without conflating inner and whole-wrapper results. Actual stored zero/final mode unasserted.
  - runtime_observed: no
- **SDR-V06** Town/classifier (SDR-R02; boundary: isolated town/drain stage)
  - Inputs: town cases class=-1/inventory absent; initial stamina independently 0 or 23552; other defaults.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: town precedes class/inventory/stat work.
  - Outputs:
    - (a) Each id {1,40,75,103,109}: drain=1; no stat reads/requests/Walk; stamina stays supplied 0 or 23552.
    - (b) Classifier only, absent room or id 0/2/110/-1: town=false each; no whole-drain result asserted.
    - (c) Direct absent-unit drain: return=0 without room/stat request.
    - (d) Full supplied-input drain with present player/dynamic path absent, otherwise defaults, post-add read=23512: absent room is non-town; w=40; log `[Add(10,-40,0),Read(10,0)=23512]`; drain=1/fixture stamina=23512; no Replace/Walk. No later full-movement null-path result asserted.
  - Discriminates: zero stamina/invalid class do not defeat town exemption; invalid level ids do not become town.
  - runtime_observed: no
- **SDR-V07** Truncation/wrap (SDR-R07; boundary: supplied-input stage)
  - Inputs: defaults; independent p/read overrides below.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: wrap p*40, signed /-100, add 40, floor.
  - Outputs:
    - (a) p=33: product=1320/adjustment=-13/w=27/delta=-27/read=23525/fixture stamina=23525/drain=1.
    - (b) p=-33: product=-1320/adjustment=13/w=53/delta=-53/read=23499/fixture stamina=23499/drain=1.
    - (c) p=107374183: low32 product=24/adjustment=0/w=40/delta=-40/read=23512/fixture stamina=23512/drain=1.
  - Discriminates: (a)/(b) reject rounding away from zero; (c) rejects wide product/universal p>=100 floor.
  - runtime_observed: no
- **SDR-V08** Dispatch/walking (SDR-R01a/SDR-R01b/SDR-R01c/SDR-R01e/SDR-R01f; boundary: supplied-input dispatch stages)
  - Inputs: default unit/mode=2, A=11/B=22; supplied handler/path results; callbacks log arguments. Cases independent.
  - Seeds: `no draw`.
  - Draws: none locally; supplied callbacks make none.
  - Trace: execute only stated stage, inspect final calls/results.
  - Outputs:
    - (a) Frame stage: one FramePump request. Pump's player slice: PersistentPass then CurrentBucketPass once each; other groups outside fixture.
    - (b) Persistent [E], type=0/target=player/no direct callback/removal, supplied mode-handler result=1: PlayerMode(game,player,11,22); mode 2 selects movement; no Drain/Add/Walk/ModeEnd; stamina=23552; E remains/running=false.
    - (c) Supported-mode selector/movement result=2: Movement(game,player,11,22) then ModeEnd(game,player,11,22); mode=2/stamina=23552. Return=1 instead gives only Movement/no ModeEnd.
    - (d) Selector/direct D/type=255: only D(game,player,255,11,22), ignores D return=2. No D/type=255: no call. No D/type=2 (original null handler): no call.
    - (e) Due [E]/E.frame=100/currentFrame=101: no callback/removal. E.frame=101 instead: callback then Remove(E). Persistent pass/E.frame=100 still calls E.
    - (f) Supplied movement/mode independently 2/6/19: no Drain/Add/Walk; PathStep requested; stamina=23552 with supplied inert path. Mode=3/Drain=0: Drain→WalkRequest(2)→PathStep. Mode=3/Drain=1: Drain→PathStep. Absent unit skips Drain/requests PathStep; no safe full-handler null result asserted.
    - (g) Persistent [E,F], otherwise inert callbacks but E sets pending-removal: at E callback entry running=true/pending-removal=false; callback sets pending-removal=true. Immediately after callback and before Remove(E), running=false/pending-removal=true. Remove(E) follows; at F callback entry running=true/pending-removal=false; immediately after F callback running=false/pending-removal=false; F remains. Thus the observed order is `[CallbackE(running=true),AfterE(running=false),Remove(E),CallbackF(running=true),AfterF(running=false)]`. If E also supplies new head N with running=false/pending-removal=false, saved-successor continuation still visits F, never N in this pass; N's running remains false.
    - (h) Mode selector with mode=3 and supplied movement execution Drain=0, then WalkRequest(2), PathStep, and final movement result=2: log ends with ModeEnd(game,player,11,22). An independent final movement result=1 has the same Drain/Walk/Path prefix and no ModeEnd. This supplied final-result boundary does not verify arrival effects that produce it.
    - (i) Isolated event selector/type=1/no direct callback, target=player/A=11/B=22, supplied selected handler result=2: exactly ModeEnd(game,player,11,22); no second mode-end request and no PlayerMode call. The event list pass ignores this selected-handler result.
    - (j) Isolated mode selector/absent unit, A=11/B=22, supplied mode-0 handler result=1: exactly Mode0Handler(game,absent,11,22); no ModeEnd. No selected handler's interior is executed.
    - (k) Full supplied persistent-list pass [E,F]: E has direct callback D returning 2, F has no direct callback/type 0 and its supplied player-mode handler returns 1. Neither callback changes links or pending-removal. Calls are exactly `[D(G,U,0,11,22),PlayerMode(G,U,11,22)]`, in E/F order; no ModeEnd or Remove request. E and F each have running=true at their callback entry, then running=false immediately after; both remain linked in order [E,F] with pending-removal=false. D's return 2 does not alter traversal, retention or mode-end routing.
    - (l) Full supplied persistent-list pass [E,F,J]: at E callback entry running=true and saved successor=F. E's supplied callback logs removal of F and supplies the resulting list [E,J] and saved successor=J; this mutation is an explicit fixture input, not an established arbitrary callback producer. After E callback running=false; next callback is J with running=true, then running=false afterwards. Callback order is exactly E,J; F is not called, no dispatcher Remove request is made, E/J remain linked/running=false/pending-removal=false, and the saved successor at pass completion is absent. All supplied callback results are 1. Stamina stays 23552 under these inert stat dependencies.
  - Discriminates: due equality is bucket-only; D overrides invalid type; only exact selected-mode-handler result 2 causes ModeEnd, not Drain/event-callback result. Case (g) rejects omitting either running=true at callback entry or running=false after callback. Case (k) rejects result-driven mode-end processing/removal or early termination by the list pass. Case (l) rejects continuing at a stale pre-callback successor after the supplied mutation updates it.
  - runtime_observed: no
- **SDR-V09** Duration (SDR-R01f/SDR-R12; boundary: supplied-input repeated movement stage)
  - Inputs: defaults, one post-preparation movement-stage invocation/frame calling drain, constant w=40, no writer/overflow, fixture persistence above, 25 fps; mode stays supplied as 3 until the last invocation. Walk/PathStep log requests and have no unspecified effects.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: 23552-588*40=32; frame 589 gives -8 then Replace/drain=0.
  - Outputs: 589 movement/drain invocations; fixture stamina=32/drain=1 at 588, stamina=0/drain=0 at 589; one Replace/Walk at 589 followed by PathStep; no Walk at earlier invocations; 589 PathStep requests; 23.56 s. Actual storage/cadence/transition gated.
  - Discriminates: floor duration 588 fails positive frame-588 result.
  - runtime_observed: no
- **SDR-V10** Signed weight/zero base (SDR-R06/SDR-R08; boundary: supplied-input stage)
  - Inputs: defaults/synthetic torso; independent overrides below; supplies resulting post-add reads.
  - Seeds: `no draw`.
  - Draws: none.
  - Trace: signed quotient/multiplier/product then floor.
  - Outputs:
    - (a) speed=-10: multiplier=0/pre-floor w=0/final w=1/delta=-1/read=23551/fixture stamina=23551/drain=1.
    - (b) speed=-9: quotient=0/multiplier=1/w=40/delta=-40/read=23512/fixture stamina=23512/drain=1.
    - (c) RunDrain=0/speed=0: base=0/floor=1/delta=-1/read=23551/fixture stamina=23551/drain=1.
  - Discriminates: (b) distinguishes truncation/floor division; (c) rejects skipping minimum for zero base.
  - runtime_observed: no
- **SDR-V11** Class failures (SDR-R03/SDR-R01f; boundary: supplied-input movement stage)
  - Inputs: defaults; independent class=7, class=-1, or class=4/absent computed record; count=7; Walk/Path log only.
  - Seeds: `no draw`.
  - Draws: none locally.
  - Trace: class guard returns 0 before base/inventory/stat work.
  - Outputs: each drain=0/stamina=23552; no Add/Read/Replace; [WalkRequest(2),PathStep]; no final-mode assertion.
  - Discriminates: class==count fails; positive stamina does not prevent failure's Walk request.
  - runtime_observed: no
- **SDR-V12** Inventory distinctions (SDR-R05a/SDR-R05b/SDR-R05c/SDR-R06/SDR-R01f; boundary: supplied-input stage)
  - Inputs: defaults/independent overrides below; mode/path log only; stat fixture model where used.
  - Seeds: `no draw`.
  - Draws: none locally; supplied allocations make none.
  - Trace: inventory absence fails; getter absence continues; supplied-success lazy allocation initializes empty slots.
  - Outputs:
    - (a) Absent inventory: drain=0/stamina=23552; no stat work; [WalkRequest(2),PathStep].
    - (b) Invalid inventory: torso absent/w=40/delta=-40/read=23512/fixture stamina=23512/drain=1; no Walk.
    - (c) Absent page/supplied allocations succeed: one 13 by 1 page/13 empty slots; torso absent/w=40/delta=-40/read=23512/stamina=23512/drain=1.
    - (d) Existing page without slots/supplied allocation succeeds: dimensions 13 by 1/13 empty slots; w=40/delta=-40/read=23512/stamina=23512/drain=1.
    - (e) Existing 12 by 1 page/slots present: no item/no resize/page stays 12 by 1; w=40/delta=-40/read=23512/stamina=23512/drain=1.
    - (f) Torso unsigned class=itemCount: no record; w=40/delta=-40/read=23512/stamina=23512/drain=1.
    - (g) Present torso/class 0, itemCount=10, items collection absent: fatal termination in item-record lookup, with no numeric return and no later percent read/Add/stamina Read/Replace/Walk request. This differs from an out-of-range class yielding an absent record; fixture stamina remains the supplied initial 23552 because no supplied stat operation ran before termination.
  - Discriminates: invalid inventory differs from absence; mismatch does not permit read/resize; absent item record does not fail drain. Case (g) rejects treating an absent indexed items collection as a safe absent-record fallback.
  - runtime_observed: no

### 8.2 Observation vectors

Each has a [companion procedure](../research/spec-evidence/player-stamina-drain-while-running.md#observation-procedures). Control actual cadence/storage/other writers. No successful drain observation recorded.

- **SDR-O01** Unarmored: initial 23552/RunDrain=20/p=0; expect 40 per drained frame, 32 after 588, zero/Walk request at 589 under no-writer conditions.
- **SDR-O02** Heavy torso speed=10: expect 80/frame, 32 after 294, zero/Walk request at 295.
- **SDR-O03** Medium speed=5: expect 40/frame, conditional duration 589.
- **SDR-O04** Town level 1: no drain-owned stat work with positive or reachable zero-stamina running; distinguish regeneration.
- **SDR-O05** p=50: expect 20/frame; p=100: expect 1 with base=40/no torso weight.
- **SDR-O06** New Run request/zero stamina outside town: observe player-mode Walk result under SDR-Q04; no confirmed drain assertion.

## 9. Acceptance and dependencies

Partial acceptance requires verified arithmetic/supplied lookup/control/ordered requests in SDR-V01..SDR-V12 including every case. Fixture stamina is not original storage. SDR-R01d/SDR-R11 remain unresolved. Full reproduction needs SDR-Q01/SDR-Q02/SDR-Q03/SDR-Q04/SDR-Q07 integration, writer/scheduling relation SDR-Q06, profile qualification/matching observations SDR-Q05. A supplied subset/table read never establishes complete parity.

**Depends on:** [NRF](core-numeric-and-rng-foundations.md) NRF-R08a units, NRF-R09a rate, NRF-R12b/NRF-R12c existing-array reads, NRF-Q20 writes; [GDF](core-game-data-fields.md) GDF-R03a/GDF-R11a/GDF-R14a class/stat/mode names/fields; [STL](state-timed-state-lifecycle.md) STL-R06c write integration; [MCI](missile-creation-and-initialization.md) MCI-R12b fresh event association/MCI-Q03 limits. Complete player-mode/frame-event/additive/network contracts remain named gaps.

**Depended on by:** candidate player-mode behavior consumes drain return/Walk request; candidate regeneration accounts for this consumer; candidate stat-update behavior consumes resulting changes. These are relationships, not delivery prerequisites/completed contracts.

## 10. Related specs

- [combat-open-wounds](combat-open-wounds.md), OWD-R11a: life→stamina→mana order; no drain-interference proof.
- [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md): supplied velocity/state/stat integration.
- [item-durability-loss](item-durability-loss.md): body-slot lookup/qualified BodyLocs corroboration.
- [missile-movement-and-collision](missile-movement-and-collision.md): separate frame-event consumer/recurrence limits.

## Appendix A. Provenance

Original Windows x86 Lord of Destruction Expansion 1.13c. [Manifest](../research/source-manifest.md) artifacts LOD113C-D2GAME/LOD113C-D2COMMON identify code; LOD113C-STORM/LOD113C-GAME describe loader/profile limits. Compatible 2026-10-12 static readings and shared 2026-10-14/2026-10-15 evidence retain dates. Targeted disassembly/checked reconstruction/fingerprinted original-image instruction/table reads on 2026-10-15 support local requests/lookups. The 2026-10-12 runtime attempts matched no drain vector; compiled class fields corroborated inputs only. Fresh 2026-10-15 archive reads corroborate named text cells with profile limits. Detailed sources/corrections/observation scope are in the [companion](../research/spec-evidence/player-stamina-drain-while-running.md).
