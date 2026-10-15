# Missile Movement and Collision

**Description**

A missile advances, loses lifetime and checks what it encounters during each server update.
Terrain can stop it before it reaches a unit, and its collision settings decide which units
it can contact. Some contacts let it keep flying; others end it or invoke a hit effect.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: MMC-R, MMC-Q, MMC-V, MMC-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/missile-movement-and-collision.md)
Open evidence gaps: [gap file](../plans/gaps/missile-movement-and-collision.md)

## 1. Scope

This contract covers the server dispatcher, plain movement caller, lifetime/collision gates,
ordered cell/candidate selection, contact eligibility, pierce, cooldown requests and termination
results. Hit, event and damage interiors belong to [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md)
(MHD) and combat contracts. Creation supplies speed, lifetime, activation, owner and path state
through [missile-creation-and-initialization](missile-creation-and-initialization.md) (MCI).
Other server do handlers, full generic path planning/collision-map maintenance, streams,
networking and client presentation are outside this boundary. Supplied dependency results
support the local consumers below without establishing unread producers.

## 2. Trigger and ownership

D2Game `MISSILES` owns dispatch/contact policy; do selector 1 forwards to the plain step.
D2Common `PATH`/`COLLISN` own movement and collision operations. Movement precedes lifetime,
collision gates and at most one selected unit contact per supplied plain update. Creation's
event association consumes MCI-R12b; the callback-free event walker calls this dispatcher.
Recurring queue production and relative system/missile order remain MMC-Q03. No direct network
message is produced here; delegated combat effects retain their owners' message boundaries.

## 3. Inputs and data

All integer operations consume NRF widths, signedness and low32 conventions. Semantic operation
names impose no language/API/storage design. Supplied values do not establish their producers.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| missile/record | Type 3 unit and selected Missiles row or absent | association | Dispatcher checks validity; contact does not recover safely |
| selectors | Missiles.pSrvDoFunc, pSrvHitFunc, pSrvDmgFunc | signed i16 | Engine handler indices |
| towns | Town, SrcTown; missile/owner room and level | booleans; i32 level | Null room not town; levels 1,40,75,103,109 are town |
| owner | valid owner-link condition, type/id, supplied game lookup | association | Direct creator before hostility's further resolution |
| flags | LastCollide, Explosion, Pierce, CanDestroy | booleans | Record policy |
| collision fields | CollideType, Collision, CollideKill, CollideFriend, NextHit, NextDelay, Size, ToHit, AlwaysExplode | u8 | CollideType also stored on missile |
| remaining/activationCounter | MCI-R14/MCI-R21 metadata | signed i16 frames | Activation compared after decrement |
| last collision | lastType/lastId, target type/id | i32 | Empty source id -1 |
| velocity | path presence, speed/targetSpeed/accel/tick, directions, step | i32; speed 8.8, directions scale 4096, step 16.16 | NRF-R11a/NRF-R11b |
| trace state | fixed position/step, accumulated bits, recording/restricted-blocking/trace conditions; specialRelocationCondition | u32 16.16; u16 bits; booleans | Consumer MMC-R05b; special relocation is false in its supplied missile boundary; producers MMC-Q01 |
| path selector | Engine-defined path-type selector; value 4 on the fresh missile-allocation branch | i32 selector | Selects the initial recording/restricted conditions in MMC-R05c; complete path production remains gated |
| movement operations | room-link update, path-step result, transition predicate/update | supplied operations/results | Actual effects MMC-Q04 |
| collision query | room, integer x/y, size, mask, returned bits | i32; u16 result | Stationary mask 32767; scan mask MMC-R08 |
| pathFootprintSize | Stored path footprint size | i32 | Stationary query input; distinct from Missiles.Size used by cell scan/candidate search/final footprint request |
| cells | entered cells in stored order; count | u16 x/y pairs; i32 count | Recorded list; no sorting |
| candidates | ordered adjacent rooms/unit lists, bounds, candidate type/mode/position/size | logical collections; i32 bounds/sizes; u16 dynamic coordinates | MMC-R09b |
| eligibility | candidateEligibleA/B; active state 86; alignment tier | booleans; i32 | Both conditions required; tier PVP-R05a |
| hostility | owner, candidate, directional relations/owner-resolution inputs | PVP-R05a/PVP-R07 inputs | Producer gates retained |
| stats | missile 19 tohit,328 pierce_idx; owner/source 156 item_pierce,166 skill_pierce; source own-base 328 | signed i32 layer 0 | NRF-R12b/NRF-R12c; own-base differs from aggregate |
| hand-off | hit result, event callbacks, hit/damage handlers/results | supplied operations/i32 results | MHD/SBA/IEP own interiors |
| cooldown | game.frame, allocation result, list/state/timer operations | i32 frame; u8 delay; optional list | STL-R01/STL-R06a/STL-R06b/STL-R07 |
| seeds | source/missile/game/local seed | two u32 each | Local pierce steps only local seed |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Stat names/ids 19,156,166,172,328; states 86 justhit,105 alignment | Named ItemStatCost/States cells, [records](../research/spec-evidence/missile-movement-and-collision.md#named-cell-records) | Names from data; numeric selectors fixed by code |
| Missile columns/flags | GDF-R07 loader descriptors | Data selects engine-defined policy |
| Handler counts 53/71/31, action bits 1/2/4, masks, scale 1024, trace cap 10 | Engine code/tables | Fixed by original code |
| Collision values 1,4,64,128,256 | Numeric consumers/mask table | Wider writer meanings MMC-Q05 |
| Candidate size | Player 2; monster signed class-size; missile unsigned Size; supplied size result | Code selects type/field | Actual data/list/map integration MMC-Q04 |

## 4. Rules (execution order)

Only do result 2 requests dispatcher removal. Result 1 continues; other results do not request
removal here. Contact result 1 means rejected, aborted or continuing; 2 requests termination.
It is not a damage-success result.

**MMC-R01** Dispatcher/towns [verified in 1.13c code (static)]. Absent missile/record or do
selector <=0 or >=53 returns without do/removal. Resolve missile room, falling back to
coordinate-containing-room selection, and direct owner. Present owner in town causes removal
when SrcTown is set or owner is player. Independently, Town clear and missile room town causes
removal. Otherwise call the selected do handler and request removal exactly for result 2.
Null room is not town. No pointer guard exists: slots 4 and 38..52 are null and fault if selected.
Actual destruction is MMC-Q04; other do interiors are outside scope.

**MMC-R02** Plain selector [verified in 1.13c code (static)]. Slot 1 calls plain step and
forwards its result unchanged, without an additional local effect/draw.

**MMC-R03** Invalid direct step [verified in 1.13c code (static)]. Invalid record at direct
plain-step entry calls Contact(null,force=1). Contact reads the absent row without a guard and
faults; no successful end/removal result follows. Ordinary dispatcher rejects invalid rows first.

**MMC-R04** Velocity.
- **MMC-R04a** [verified in 1.13c code (static)]. Entering speed 0 skips movement and its velocity
  update. Supplied moving stage consumes NRF-R11a with scale 1024: low32 products, arithmetic
  shifts, signed ramp comparisons and strict speed>targetSpeed clamp. Negative acceleration
  flooring speed to 0 remains stored, but later plain updates skip the moving stage.
- **MMC-R04b** [verified in 1.13c code (static)]. Prepared direction scale 4096 and cited
  axis/diagonal entries consume NRF-R11b. Creation speed/acceleration/maximum consume MCI-R07,
  MCI-R18a/MCI-R20. Actual direction production retains MCI-Q04/NRF-Q04; supplied arithmetic
  remains available independently.

**MMC-R05** Movement and bounded trace.
- **MMC-R05a** [verified in 1.13c code (static)]. With nonzero entering speed, request room-link
  update, path step(missile,1024), then conditional room-transition update when its predicate is
  set. Wrapper returns 2 for path-step 0, otherwise 1. On wrapper 2, call Contact(null,force=1)
  and forward its result before lifetime/scan. Actual room/path effects remain MMC-Q04.
- **MMC-R05b** [verified in 1.13c code (static)]. With specialRelocationCondition false and
  supplied CheckMove collision results,the bounded tracer resets count to 0. Copy
  requested step to temporary increment; arithmetic-shift both components right one repeatedly
  until X lies in [-65536,65536], then do the same until Y lies in that interval. Retain the
  original full endpoint. Advance temporary fixed position by that increment with low32 wrap
  until its u16 integer-cell pair equals the full endpoint's pair. Only changed cells call
  CheckMove(old,new). Its supplied query bits are ORed into accumulated u16 bits; ordinary
  blocking accepts iff bits=0, restricted blocking iff (bits&5)=0. A separate player/monster
  relocation branch is outside this missile consumer (MMC-Q04). Refusal excludes the new cell:
  return blocked=1, current count and centre of the last accepted temporary cell. Acceptance
  appends the new cell only with recording-condition, then increments count regardless of that
  condition. At count>=10 return blocked=0,count=10,original full endpoint without more tests.
  Normal arrival returns blocked=0/full endpoint/count. Unblocked return with count>0 sets
  traceCondition. Local results do not establish path-step return/publication (MMC-Q04).
- **MMC-R05c** [verified in 1.13c code (static)]. On the fresh type-3 path-allocation branch,
  path selector 4 selects the fixed entry that sets recording-condition and restricted-blocking
  condition true. This verifies those local initialization writes; complete later planning and
  flag retention remain MMC-Q01. Cell-list access without recording-condition takes a fatal
  assertion. Supplied recorded lists support MMC-R09a independently.

**MMC-R06** Lifetime [verified in 1.13c code (static)]. After movement sign-extend remaining,
compute i32 r=remaining-1, store clamp(r,-32768,32767). Test unclamped r: if r<=0 call
Contact(null,force=1) and forward its result; otherwise continue. Last ordinary frame moves but
never scans; counter 3 at f expires at f+2. Counter -32768 stores -32768 but tests -32769.
Surviving abort can repeat expiry on later updates. With absent missile metadata,the getter
supplies 0 and the setter ignores its request,so r=-1 and force-end runs without a stored
lifetime write. No local draw.

**MMC-R07** Collision gates [verified in 1.13c code (static)], after lifetime:

```
if CollideType==0: return 1
if room absent or path absent: return 2
m = accumulated moving bits
if current speed==0:
    m = Query(room,currentX,currentY,pathFootprintSize,32767)
    accumulatedBits=low16(m)
if m & 5: return 2
if CollideType==6: return 1
if remaining > activationCounter: return 1
if low16(m)==0: return 1
if selected filter absent: return 2
continue to cell scan
```

Result-2 gates invoke no contact,cooldown,event,hit handler or MMC-R15 footprint request.
Actual Collision footprint/query effects remain MMC-Q04; own-footprint presence is not assumed.

**MMC-R08** Selector table [verified in 1.13c code (static)].

| CollideType | Mask | Filter before MMC-R10 |
|---|---|---|
| 0 | 0 | absent; early return |
| 1 | 132 (0x84) | players or monsters with alignment tier 2 |
| 2,5 | 260 (0x104) | monsters |
| 3 | 388 (0x184) | players and monsters |
| 4 | 0 | absent |
| 6 | 4 | absent; earlier return |
| 7 | 64 (0x40) | missile with valid CanDestroy row; bypass MMC-R10 |
| 8 | 389 (0x185) | players and monsters |

Filters require present context/record/candidate; non-7 filters then invoke MMC-R10. Tier uses
PVP-R05a's matching state-105 list own stat 172, independently of active alignment bit. Values
above 8 index beyond the table without a guard; no successful fallback is specified (MMC-Q04).

**MMC-R09** Cell/candidate order.
- **MMC-R09a** [verified in 1.13c code (static)]. In recorded-cell order query
  Query(room,x,y,missile Size,selected mask). Zero bits skips cell. Otherwise search through
  MMC-R09b. First returned candidate calls Contact(target,force=0); forward its result
  immediately, even rejection. If no candidate and bits&4, call Contact(null,force=1), ignore
  its return and return 2. Even an aborting hit handler cannot preserve the missile on this
  barrier branch. Otherwise advance; exhaustion/count<=0 returns 1.
- **MMC-R09b** [verified in 1.13c code (static)]. Absent room or missile Size outside 1..3
  returns absent. Visit supplied adjacent rooms in stored order, then each accepted room's
  ordered units. Room eligibility uses signed wrapped arithmetic and these literal branches:

```
if i32(x+2) < left:
    if i32(x-2) > i32(left+width): reject room
if i32(y+2) < top:
    if i32(y-2) > i32(top+height): reject room
accept room
```

  Skip players in mode 0/17, monsters in mode 0/12, all types except player/monster/missile.
  Read candidate size, clamp >=3 to 3 and skip <=0. Let dx=abs(x-unitX),dy=abs(y-unitY) over
  u16 dynamic coordinates. Apply geometry below, then filter(candidate,unchanged context).
  Return first nonzero-filter candidate; exhaustion returns absent. Actual list/size/position
  producers retain MMC-Q04.

| Missile size,candidate size | Geometry |
|---|---|
| 1,1 | dx=dy=0 |
| 1,2 or 2,1 | dx+dy<=1 |
| 1,3 or 3,1 | dx<=1 and dy<=1 |
| 2,2 | dx+dy<=2 |
| 2,3 or 3,2 | (dx<=2 and dy<=1) or (dy<=2 and dx<=1) |
| 3,3 | dx<=2 and dy<=2 |

**MMC-R10** Search filter [verified in 1.13c code (static)]. Reject absent candidate, then
unless candidateEligibleA/B both true. NextHit with active state 86 rejects; matching LastCollide
rejects. Absent owner accepts; otherwise accept hostility(PVP-R05a), else only CollideFriend!=0.
Search never writes last collision. Type-7 CanDestroy filter bypasses this.

**MMC-R11** Contact filters [verified in 1.13c code (static)]. Resolve direct creator. For
present target test NextHit/state 86, then LastCollide equality; either rejection returns 1 with
no later operations. Next request last-collision update: LastCollide plus present metadata stores
target type/id; otherwise retain fields. Type gate rejects CT0, accepts only player for CT1,
only monster for CT2, any type for unsigned CT>=3 (also supplied negative selectors). Then present
owner requires hostility or CollideFriend; absent owner skips this test. Failures return 1.
Type/hostility-refused targets can become last collision. CT1 finds a tier-2 monster but rejects
it at contact, ending this frame's scan while continuing the missile.

**MMC-R12** Pierce.
- **MMC-R12a** [verified in 1.13c code (static)]. Consume MCI-R26a creation with player/monster
  source,type-3 missile and Pierce. MCI-R26a/MCI-R26b own the complete calculation: supply source
  current aggregate stats 156/166 for chance and source own-base stat 328 for its local seed.
  Consume its exact ordered local draws,count and layer-0 missile-328 write request; zero chance
  skips that request. Source/missile/game seeds are unchanged by this helper. MMC-V09 exercises
  the dependency contract's seed/count boundary; actual stored count remains MMC-Q07 and the
  owner own-base producer remains MMC-Q02.
- **MMC-R12b** [verified in 1.13c code (static)]. After gates, with owner,target,Pierce and
  current owner 166!=0 or current owner 156!=0, read missile 328. Read 166 first and 156 only
  if that result is zero. Nonzero n requests low32(n-1),
  layer 0 and sets stopMask=2; zero leaves stopMask=3. Other branches retain 3. Owner stats are
  read anew per contact; negative n is nonzero. No local draw. Actual persistence MMC-Q07.
- **MMC-R12c** [unresolved]. Layer-0 replacing requests are established; storage/propagation
  consumes STL-R06c/NRF-Q20, remains MMC-Q07. Supplied persistence supports the multi-contact
  vector without establishing actual stored counts.

**MMC-R13** Initial actions [verified in 1.13c code (static)]. act=0 for Explosion else 2;
CollideKill!=0 ORs mask 1. stopMask starts 3, overridden only by MMC-R12b. Mask 1 permits contact
ending,mask 2 selects damage,hit-function mask 4 aborts.

**MMC-R14** Hand-off [verified in 1.13c code (static)]. Target plus ToHit requests SBA roll
(attacker=direct owner,defender=target,bonus=missile 19,isMissile=1). On miss, run MMC-R16 then
item event 0 hitbymissile(owner=target,other=missile,record absent,auxiliaries 0); ignore event result.
Only if AlwaysExplode and 0<hit selector<71 call hit handler: result&4 returns 1; otherwise return 2.
No damage or footprint request on miss.

On target hit, or absent target with force or AlwaysExplode, run cooldown,event,then valid hit
handler. Handler return replaces act entirely; act&4 returns 1 immediately. Bounds do not guard
null pointers; selected null slots fault. Absent event owner invokes no callbacks (IEP-R07).
Target plus act&2 requests damage fill,then handler when 0<damage selector<31,then application;
ignore damage-handler/application numeric results for ending. No target,force=0,AlwaysExplode=0
skips those operations. Finish at MMC-R15. Handler/damage/proc/draw interiors retain MHD/IEP/SBA
contracts and gates (MMC-Q08).

**MMC-R15** Finish [verified in 1.13c code (static)]. If neither act&1 nor force,return 1.
If stopMask&1 clear,return 1. Otherwise present room requests footprint removal(room,currentX,
currentY,Size,64); absent room skips request. Return 2. Actual map/unit destruction MMC-Q04.
Force-end can continue on hit abort except MMC-R09a's barrier discards that result.

**MMC-R16** Cooldown requests [verified in 1.13c code (static)]. NextHit=0 or absent target
returns with no request. expiry=low32(game.frame+unsigned NextDelay). Allocate list(type=2,
owner=target type/id,expiry). Absent result returns before later operations. Success in order:
schedule target timer 12 at expiry,callback absent/auxiliaries 0; set list state 86; set contact-expiry
callback; attach to target with condition 1; request target state 86 enabled. No internal
existing-list check. Attach/state/expiry effects remain MMC-Q09. Once active,state 86 gates all
NextHit missiles,not just the missile that requested it.

**MMC-R17** Draw scope [verified in 1.13c code (static)]. Dispatcher guards,lifetime/gates,
local trace/search/contact decisions,cooldown caller,action selection and finish make no direct
draw. Creation pierce alone steps a local seed in this slice. Path/room/collision/removal interiors,
other do handlers,event callbacks,hit/damage/to-hit operations retain their draw boundaries.
No whole-update zero-draw assertion follows.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MMC-R01 | Dispatcher/towns | verified in 1.13c code (static) | MMC-Q03 queue; MMC-Q04 removal | MMC-V01 |
| MMC-R02 | Plain forwarding | verified in 1.13c code (static) | none locally | MMC-V01 |
| MMC-R03 | Invalid direct-step fault | verified in 1.13c code (static) | none locally | MMC-V01 |
| MMC-R04a | Supplied velocity | verified in 1.13c code (static) | MMC-Q04 path integration | MMC-V02 |
| MMC-R04b | Direction/creation dependency | verified in 1.13c code (static) | NRF-Q04/MCI-Q04 direction producer only | MMC-V02 |
| MMC-R05a | Moving caller | verified in 1.13c code (static) | MMC-Q04 path/room | MMC-V03 |
| MMC-R05b | Trace consumer | verified in 1.13c code (static) | MMC-Q04 queries/publication | MMC-V03 |
| MMC-R05c | Initial path-condition writes | verified in 1.13c code (static) | MMC-Q01 later planner/retention only | MMC-V03 |
| MMC-R06 | Lifetime | verified in 1.13c code (static) | none with supplied movement | MMC-V04 |
| MMC-R07 | Collision gates | verified in 1.13c code (static) | MMC-Q04 actual bits | MMC-V05 |
| MMC-R08 | Mask/filter table | verified in 1.13c code (static) | MMC-Q04 selector/producer | MMC-V05,MMC-V08 |
| MMC-R09a | Cell/final barrier result | verified in 1.13c code (static) | MMC-Q04 cells/queries | MMC-V06 |
| MMC-R09b | Candidate order/geometry | verified in 1.13c code (static) | MMC-Q04 lists/sizes | MMC-V07 |
| MMC-R10 | Search filter | verified in 1.13c code (static) | PVP producer limits | MMC-V08 |
| MMC-R11 | Contact filters | verified in 1.13c code (static) | PVP producer limits | MMC-V08 |
| MMC-R12a | Pierce count request | verified in 1.13c code (static) | MMC-Q07 storage; MMC-Q02 upstream only | MMC-V09 |
| MMC-R12b | Pierce decrement request | verified in 1.13c code (static) | MMC-Q07 storage | MMC-V09 |
| MMC-R12c | Replacing integration | unresolved | MMC-Q07 | MMC-V09 gated persistence |
| MMC-R13 | Initial bits | verified in 1.13c code (static) | none locally | MMC-V10 |
| MMC-R14 | Hand-off consumer | verified in 1.13c code (static) | MMC-Q08 interiors | MMC-V10 |
| MMC-R15 | End/footprint request | verified in 1.13c code (static) | MMC-Q04 map/removal | MMC-V10 |
| MMC-R16 | Cooldown request order | verified in 1.13c code (static) | MMC-Q09 attach/expiry | MMC-V11 |
| MMC-R17 | Direct draw scope | verified in 1.13c code (static) | dependency draws excluded | MMC-V09,MMC-V10,MMC-V11 |

## 5. Outputs and side effects

| Output / effect | Rules |
|---|---|
| Do/contact results,removal request,fatal invalid-record/null-slot path | MMC-R01/MMC-R02/MMC-R03/MMC-R09a |
| Search-filter result,selected candidate or absent,and ordered filter(candidate,context) invocations | MMC-R08/MMC-R09b/MMC-R10 |
| Room-link/path-step/conditional transition requests | MMC-R05a |
| Supplied ramp/step writes | MMC-R04a/NRF-R11a |
| Trace count/list/traceCondition/output position/accumulated bits | MMC-R05b |
| Initial recording/restricted-blocking conditions | MMC-R05c |
| Remaining lifetime,stationary accumulated-bit cache | MMC-R06/MMC-R07 |
| Last-collision identity | MMC-R11 |
| Local seed,count/decrement replacing requests | MMC-R12a/MMC-R12b |
| Hit-roll/event/hit/damage-fill/handler/application requests | MMC-R14 |
| Footprint-removal request | MMC-R15 |
| Allocation/timer/list-state/callback/attach/state-enable requests | MMC-R16 |
| Actual map/removal,replacing-stat/cooldown persistence/expiry | Gated MMC-Q04/MMC-Q07/MMC-Q09 |
| Delegated damage/procs/messages/draws | MMC-R14 dependencies,MMC-Q08 |

## 6. Branches and edge cases

- Speed 0 skips acceleration; movement flooring speed to 0 selects stationary query that update.
- Movement failure calls force-end before lifetime; silent gates omit it; a later no-candidate
  barrier calls force-end but discards its result.
- A found unit refused by contact ends the scan; later cells/candidates are not consulted.
- Absent owner skips hostility/pierce but does not guarantee damage under MHD/SBA guards.
- Invalid dispatcher row returns; invalid direct-step row faults; null slots are not fallbacks.
- Size 0 or >3 finds no unit but can reach barrier termination.
- Chance 0 skips initialization; negative chance draws once/count 0; chance >99 can succeed four
  times. Negative nonzero count decrements with low32 wrap.
- Cooldown allocation failure stops only that helper's later requests,not the surrounding contact.
- Dependencies retain difficulty/PvP/classic/expansion distinctions; no additional local branch.

## 7. Open questions

| Question | Kind / blocked assertions | Evidence that settles it |
|---|---|---|
| **MMC-Q01** | integration: final path conditions after planning | Initial writes are MMC-R05c; complete later planner/retention chain; supplied trace/list consumers available |
| **MMC-Q02** | outside-boundary: owner own-base 328 producer | Identify player/monster writers; supplied seed does not imply universally fixed rolls |
| **MMC-Q03** | integration: recurring queue/relative order | Registration/requeue/event chain; supplied update available |
| **MMC-Q04** | integration: actual path/query/cell/list/size/map/removal,out-of-table selectors | Complete nested path/collision/room/removal; supplied results do not establish effects |
| **MMC-Q05** | outside-boundary: collision-bit writer meanings | Terrain/player/monster map writers; numeric masks need no guessed names |
| **MMC-Q06** | observation: local rules unobserved | Controlled matching 1.13c records; other builds outside boundary |
| **MMC-Q07** | integration: replacing-stat storage/propagation | NRF-Q20/STL-R06c; arithmetic/requests available |
| **MMC-Q08** | integration: to-hit/event/hit/damage interiors | Applicable SBA/IEP/MHD/handler contracts retain gates |
| **MMC-Q09** | integration: attach/state/expiry/callback effects | STL-R05/STL-R06b and expiry chain; request order available |

Attempts and remaining limits are in the [gap file](../plans/gaps/missile-movement-and-collision.md).

## 8. Test vectors

### 8.1 Executable vectors

Each case starts independently at the stated supplied-input stage. Defaults: present type-3
missile/row/path/metadata,do=1,Town=1,SrcTown=0,CT=2,Size=1,remaining=20,activationCounter=20,
speed=0,non-town room R level 2,owner absent. NextHit/LastCollide/Pierce/Explosion/CollideKill/
ToHit/AlwaysExplode=0,hit/damage selectors=0,last=(type 3,id -1),stats=0/no minimum adjustment,
active states absent. Fixed position=(688128,688128),integer=(10,10),pathFootprintSize=1,
mask=260,bits=0,cells/room
lists empty. Path-step=1,transition predicate false,recording/restricted conditions true,
traceCondition false,specialRelocationCondition false,count=0. Eligible candidates: type 1,mode 1,size 1,both eligibility
conditions true. Game frame=1000. Values are synthetic supplied fixtures,not named table rows.

Named room/path/footprint/Remove,event/hit/damage,list/attach/state requests log names/arguments
in order and otherwise have no effect unless explicitly given persistent storage. Pure supplied
room/owner/stat/state/size reads do not append a log entry. Query,CheckMove,Search and filter
calls append only where their stage is being exercised; a supplied Search result bypasses its
interior. Metadata/local seeds persist. V01 includes the selected Do call in its dispatcher log;
other vectors start inside the selected handler and add only the outer Remove request for final 2.
Supplied callbacks/results have no draws; this does not establish their real draw budgets.
Unrelated unit/game seeds start {1,0},unchanged in this model. Gated stored effects are named
separately from verified local requests.

- **MMC-V01** Dispatch boundaries (MMC-R01/MMC-R02/MMC-R03; boundary: supplied-input stage)
  - Inputs: defaults,independent overrides below; a supplied plain result bypasses its interior.
  - Seeds: no draw.
  - Draws: none.
  - Trace: record/index guards,towns,do call,exact-2 removal.
  - (a) Inputs: do=53. Outputs: log [],no removal.
  - (b) Inputs: Town=0,missile level=1. Outputs: [Remove],no do call.
  - (c) Inputs: owner player in level 40,SrcTown=0. Outputs: [Remove],no do call.
  - (d) Inputs: owner monster in level 40,SrcTown=0,supplied plain result=1. Outputs: [Do1],no Remove.
  - (e) Inputs: supplied plain result=2. Outputs: [Do1,Remove].
  - (f) Inputs: do=4. Outputs: null indirect-call fault,no Remove before fault.
  - (g) Inputs: record absent at dispatcher. Outputs: log [],no step. Independent direct plain-step
    entry with absent row: [Contact(null,1)],record fault,no successful result/removal.
  - (h) Inputs: independent do=0 and do=-1. Draws: none. Outputs: each log [],no do/removal.
  - (i) Inputs: missile room absent,containing-room result absent,Town=0,owner absent,
    supplied plain result=3. Draws: none. Trace: neither room is town; forward result 3.
    Outputs: [Do1],no Remove. The plain wrapper does not rewrite supplied result 3 to 2.
  - Discriminates: bound 53,player source-town refusal,null-slot fault and entry distinction.
  - runtime_observed: no

- **MMC-V02** Supplied ramp/direction (MMC-R04a/MMC-R04b; boundary: supplied-input stage)
  - Inputs: speed=2048,accel=256,targetSpeed=3000,tick=0,dir=(4096,0); moving arithmetic invoked
    each call until speed 0; scale=1024 through NRF-R11a.
  - Seeds: no draw.
  - Draws: none.
  - Trace: changes on fifth moving calls; signed strict maximum test,low32 products/shifts.
  - Outputs: calls 1..4 speed=2048; 5=2304,10=2560,15=2816,20=3000 with accel=0,tick=0.
    Call 1 s=32768,step=(32768,0).
  - (a) Inputs: speed=100,accel=-60,targetSpeed=3000,tick=0,dir=(4096,0). Draws: none.
    Outputs: call 5 speed=40; call 10 speed=0,accel=-60,tick=0; plain call 11 skips movement,
    retaining all three values.
  - (b) Inputs: speed=256,accel=256,targetSpeed=512,tick=4,dir=(4096,0). Draws: none.
    Outputs: first call speed=512,accel=256,tick=0; after five more moving calls speed=512,
    accel=0,tick=0. Equality does not clear acceleration on first call.
  - Discriminates: fifth-call timing,strict maximum comparison,zero-speed outer gate.
  - runtime_observed: no

- **MMC-V03** Trace/movement failure (MMC-R05a/MMC-R05b; boundary: supplied-input stage)
  - Inputs: default trace position=(688128,688128),requested step=(131072,0),accumulated bits=0;
    CheckMove query at (11,10) returns 256,at (12,10) returns 1; recording/restricted conditions true.
  - Seeds: no draw.
  - Draws: none.
  - Trace: increment normalizes to (65536,0). First accepts because 256&5=0; second accumulates
    257 and refuses because 257&5=1.
  - Outputs: calls [CheckMove((10,10),(11,10)),CheckMove((11,10),(12,10))]; cells=[(11,10)],
    count=1,bits=257,blocked=1,output position=(753664,688128),traceCondition=false. This is the local
    tracer result; no actual path-step result/publication is asserted.
  - (a) Inputs: independent plain update,speed=2048,path-step supplied 0,transition=false,
    hit selector=1 returning 0 on force-end. Draws: none. Trace: wrapper 2,force-end result 2.
    Outputs: [RoomLink,PathStep(missile,1024),Contact(null,1),Event(0,null,missile,null,0,0),
    Hit(null)=0,Footprint(R,10,10,1,64),Remove]; remaining=20 unchanged,no scan.
  - (b) Inputs: independent tracer at (688128,688128),step=(1048576,0),every query=0,
    recording=true,accumulated bits=0. Draws: none. Trace: increment=65536; cap after ten cells.
    Outputs: cells=[(11,10),(12,10),(13,10),(14,10),(15,10),(16,10),(17,10),(18,10),(19,10),(20,10)],
    count=10,blocked=0,traceCondition=true,output=(1736704,688128); exactly ten CheckMove calls,
    cells 21..26 untested.
  - (c) Inputs: fresh type-3 path allocation branch,path selector 4,recording=false,
    restricted-blocking=false before selector write. Draws: none. Outputs: recording=true,
    restricted-blocking=true at that local write; no claim about later planner retention.
  - (d) Inputs: independent plain update,speed=2048,path-step=1,transition predicate=true,
    accumulated bits=0,remaining=20. Draws: none. Trace: transition request precedes decrement.
    Outputs: [RoomLink,PathStep(missile,1024),TransitionUpdate],remaining=19,final result 1,
    no Contact/Remove.
  - (e) Inputs: independent tracer position=(688128,688128),requested step=(0,131072),
    recording=false,restricted=true,all query bits=0. Draws: none. Trace: Y normalization gives
    increment (0,65536); accept two changed cells,then normal arrival. Outputs:
    [CheckMove((10,10),(10,11)),CheckMove((10,11),(10,12))],count=2,cells=[],bits=0,blocked=0,
    output=(688128,819200),traceCondition=true. Recording off does not suppress counting.
  - (f) Inputs: independent tracer position=(688128,688128),requested step=(65536,0),
    restricted=false,recording=true,first query bits=256. Draws: none. Trace: ordinary blocking
    refuses any nonzero accumulated result. Outputs: [CheckMove((10,10),(11,10))],count=0,
    cells=[],bits=256,blocked=1,output=(688128,688128),traceCondition=false.
  - Discriminates: unit-bit cell before terrain failure,blocked cell exclusion,full endpoint at
    cap,failure before lifetime; actual tracer-to-path integration remains gated.
  - runtime_observed: no

- **MMC-V04** Last frame/clamp (MMC-R06; boundary: supplied-input stage)
  - Inputs: defaults,speed=2048,remaining=3; path-step=1 each update,later accumulated bits=0,
    no hit handler. Treat successive frames as one case with retained remaining value.
  - Seeds: no draw.
  - Draws: none.
  - Trace: movement before decrements 2,1,0; force-end at zero uses act=2,stopMask=3.
  - Outputs: frames 1000/1001 store 2/1,result 1,no Remove. Frame 1002 stores 0,result 2,
    [RoomLink,PathStep(missile,1024),Contact(null,1),Event(0,null,missile,null,0,0),
    Footprint(R,10,10,1,64),Remove]; no cell query/search at frame 1002.
  - (a) Inputs: remaining=-32768,speed=0. Draws: none. Outputs: r=-32769,stored=-32768,
    force-end result 2,one Remove.
  - (b) Inputs: remaining=1,speed=0,hit selector=1 returns 4. Draws: none. Outputs: stored=0,
    force-end result 1,no footprint/Remove. A second call starting with remaining=0 and same
    callback stores -1,returns 1 again.
  - (c) Inputs: metadata absent,speed=0,valid row,no hit handler. Draws: none. Outputs: read
    remaining=0,r=-1,no lifetime write,force-end result 2,one Remove.
  - Discriminates: movement-before-expiry,unclamped comparison,abort persistence.
  - runtime_observed: no

- **MMC-V05** Gates/masks (MMC-R07/MMC-R08; boundary: supplied-input stage)
  - Inputs: independent post-decrement stages,default room/path,size=1,speed=0; m supplied below.
  - Seeds: no draw.
  - Draws: none.
  - Trace: stationary query logs Query(R,10,10,1,32767),caches low16(m); apply gates in order.
  - (a) Inputs: CT=6,m=256. Outputs: one query,result 1,no search/Remove.
  - (b) Inputs: CT=6,m=4. Outputs: one query,result 2,Remove; no Contact/event/footprint.
  - (c) Inputs: CT=4,m=64,remaining=8,activationCounter=8. Outputs: one query,result 2,Remove,
    no Contact.
  - (d) Inputs: CT=2,m=256,activationCounter=8,remaining=10 then independent 9 then 8,cells empty.
    Outputs: each result 1; first two omit cell-list access,third accesses empty list; selected mask=260.
  - (e) Inputs: CT=0,room absent. Outputs: result 1,no query/Remove. Independent CT=2,room
    absent: result 2,Remove,no Contact.
  - (f) Inputs: independent CT values 1,3,5,8; stationary m=256,remaining=8,activationCounter=8,
    cells=[(10,10)],scan query returns 0. Draws: none. Outputs: each begins
    Query(R,10,10,1,32767),then one scan query with mask respectively 132,388,260,389; result 1,
    no Search/Contact/Remove. Each stores accumulatedBits=256.
  - (g) Inputs: CT=2,room R present,path absent at this collision stage. Draws: none.
    Outputs: result 2,Remove,no query/Contact. This stage fixture does not assert safe absent-path
    entry into earlier movement code.
  - Discriminates: wall-bit gate before CT6,CT0 before room,activation equality,CT4 silent removal.
  - runtime_observed: no

- **MMC-V06** First cell/barrier result (MMC-R09a; boundary: supplied-input stage)
  - Inputs: scan cells A=(10,10),B=(11,10),CT=2,Size=1; Query(A)=256,Query(B)=256;
    Search(A)=eligible monster A1,Search(B) would return B1; supplied Contact(A1,0)=1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: A query/search/contact returns 1 before B.
  - Outputs: [Query(R,10,10,1,260),Search(A),Contact(A1,0)=1]; final result 1,no Remove,
    B untouched.
  - (a) Inputs: cells A,B; Query(A)=4,Search(A)=absent; hit selector=1 returns 4 on force-end.
    Draws: none. Trace: force-end returns 1; scan overwrites with 2. Outputs:
    [Query(R,10,10,1,260),Search(A),Contact(null,1),Event(0,null,missile,null,0,0),Hit(null)=4,Remove];
    final result 2,no footprint request,no B access.
  - Discriminates: returned candidate ends scan even on rejection; ignored abort still removes.
  - runtime_observed: no

- **MMC-V07** Candidate geometry/order (MMC-R09b; boundary: supplied-input stage)
  - Inputs: FindUnit at (10,10),missile size=1; rooms R1 then R2,origin=(0,0),extent=(100,100);
    R1 ordered candidates A size 1@(10,11),B size 2@(10,11),C size 3@(11,11); R2 has D size 1@(10,10).
    All type 1/mode 1; filter returns 1 for every invoked candidate.
  - Seeds: no draw.
  - Draws: none.
  - Trace: room accepted; A fails exact-cell test; B passes dx+dy=1.
  - Outputs: returned B,filter log=[B]; C and R2 not visited.
  - (a) Inputs: independent R1 candidates [A size 1@(10,11),C size 3@(11,11)],same room bounds,
    missile size 1. Draws: none. Outputs: C,filter log=[C].
  - (b) Inputs: independent R1 with E size 2@(11,11),missile size 1. Draws: none.
    Outputs: absent,filter log=[].
  - (c) Inputs: missile size 2,candidate size 3@(12,11),one R1. Draws: none. Outputs: candidate,
    filter log=[candidate]. Independent size 2 candidate@(12,12): absent,filter log=[].
  - (d) Inputs: missile size 3,candidate size 3@(12,12),one R1. Draws: none. Outputs: candidate,
    filter log=[candidate]. Independent missile size 0 or 4: absent,no room/filter visits.
  - (e) Inputs: R1 origin=(0,0),extent=(1,1),cell=(10,10),missile size 1,one mode 1/size 1
    candidate@(10,10). Draws: none. Outputs: room accepted,candidate returned; both upper-bound
    checks skipped. Independent mode 12 monster at that cell: absent,no filter call.
  - (f) Inputs: missile size 1,cell=(10,10),rooms R1 then R2 with default bounds; R1 has only
    X type 1/mode 1/size 1@(10,10),R2 has only Y of the same type/mode/size/position;
    filter(X)=0,filter(Y)=1. Draws: none. Outputs: filter log=[X,Y],returned Y after rejecting X.
    Independent same fixture with filter(Y)=0: filter log=[X,Y],returned absent after exhaustion.
  - (g) Inputs: one default-bounds R1,one type-1/mode-1 candidate C,filter(C)=1,cell=(10,10).
    Independent (missile size,candidate size,candidate position) cases are
    (2,1,(11,10)),(2,2,(12,10)),(3,1,(11,11)),(3,2,(11,12)). Draws: none.
    Trace: corresponding tests give Manhattan 1,Manhattan 2,square offsets 1/1,and cross offsets
    1/2. Outputs: each returns C with filter log=[C].
  - Discriminates: list order over later exact-cell proximity,size shapes,literal room branches,
    dead-mode skip before filter.
  - runtime_observed: no

- **MMC-V08** Search/contact mismatch (MMC-R08/MMC-R10/MMC-R11; boundary: supplied-input stage)
  - Inputs: CT=1,LastCollide=1,target T(type 1,id 7),tier 2 from matching state 105 own stat 172=2,
    active alignment bit false,state 86 absent,owner absent,last=(1,8); eligible conditions true.
  - Seeds: no draw.
  - Draws: none.
  - Trace: search accepts tier 2; contact early gates pass,last becomes (1,7),players-only gate rejects.
  - Outputs: search=1,Contact(T,0)=1,last=(1,7),no cooldown/event/hit/damage/footprint/Remove,
    no later-cell scan.
  - (a) Inputs: CT2,NextHit1,state 86 active,last=(1,8). Draws: none. Outputs: search 0,direct
    contact 1,last unchanged,no event.
  - (b) Inputs: CT2,NextHit0,LastCollide1,last=(1,7). Draws: none. Outputs: search 0,direct
    contact 1,no last write/event/hit.
  - (c) Inputs: CT2,state 86 absent,LastCollide1,last=(1,8),owner P,hostility=false,CollideFriend0.
    Draws: none. Outputs: search 0; independent direct contact 1 after last=(1,7). Independent
    CollideFriend1 with same entering last=(1,8): search 1,contact passes gates.
  - (d) Inputs: CT7,type 3 candidate with valid CanDestroy1 row,eligibility conditions false,
    state 86 active,NextHit1. Draws: none. Outputs: search 1,bypassing shared filter. Independent
    CanDestroy0: search 0.
  - (e) Inputs: CT=2,candidate T type 1/mode 1,state 86 absent,owner absent; independent
    (candidateEligibleA=false,candidateEligibleB=true) and (true,false). Draws: none.
    Outputs: each search 0 before cooldown/last/hostility tests; no metadata write.
  - (f) Inputs: direct contact stage,CT=2,target T(type 1,id 7),metadata absent,LastCollide=1,
    owner absent,Explosion=1,CollideKill=0,ToHit=0,NextHit=0. Draws: none. Trace: last-collision
    compare/write guards skip absent metadata; type gate accepts. Outputs:
    [Event(0,T,missile,null,0,0)],result 1,no last-collision write,no damage/footprint/Remove.
  - Discriminates: CT1 search/contact mismatch,last write before later rejection,early rejection
    before mutation,CT7 bypass.
  - runtime_observed: no

- **MMC-V09** Local pierce seed/storage boundary (MMC-R12a/MMC-R12b/MMC-R12c/MMC-R17; boundary: supplied-input stage)
  - Inputs: source type 0,missile type 3,Pierce1,source own-base328=0; chance=source 156+166 as
    independently supplied below; source/missile/game seeds {1,0}.
  - Seeds: local starts {0,666}.
  - Draws: ordered reached steps (local,modulus 100): {lo 666,hi 0}->66;
    {lo 3365183618,hi 277}->18; {lo 3979650335,hi 1403592407}->35;
    {lo 101162930,hi 1659881787}->30.
  - Trace: NRF-R01 full product/high carry each step; strict signed comparison with chance.
  - Outputs: chance 50 or 66 draws once,requests count 0; chance 67 draws four,requests count 4.
    Low words respectively 0x29A,0xC894A082,0xED34A51F,0x06079FB2. All three other seeds unchanged.
  - (a) Inputs: chance 0,entering missile 328=9. Draws: none. Outputs: no count request,local
    caller leaves entering 9 untouched. Independent chance=-1: one draw 66,request count 0.
  - (b) Inputs: five supplied sequential contacts with present direct owner P(type 0,id 1),
    valid missile owner link resolving P,and five present eligible type 1 targets T1..T5;
    supplied hostility(P,each target)=true. Entering count 4,current owner 166=67,156=0,
    Pierce1,CollideKill1,Explosion1,no hit handler; supplied replacing operation persists counts.
    Draws: none. Trace: four decrements set stopMask2; fifth sees 0/stopMask3. Outputs: replacing
    log [328=3,328=2,328=1,328=0],contact results [1,1,1,1,2],Remove only fifth. Actual engine
    persistence remains gated MMC-R12c.
  - (c) Inputs: independent present direct owner P(type 0,id 1),valid missile owner link resolving P,
    present eligible type 1 target T,supplied hostility(P,T)=true,count 4,current owner 156=166=0,
    Pierce1,CollideKill1,Explosion1.
    Draws: none. Outputs: no count read/write,result 2,Remove despite entering count 4.
  - (d) Inputs: independent present direct owner P(type 0,id 1),valid missile owner link resolving P,
    present eligible type 1 target T,hostility(P,T)=true,Pierce1,CollideKill1,Explosion1,
    owner 166=0,owner 156=67,missile 328=-1. Draws: none. Trace: zero skill component falls back to
    item component; nonzero negative count decrements. Outputs: request missile 328=-2/layer 0,
    stopMask=2,result 1,no footprint/Remove; actual stored count remains gated.
  - (e) Inputs: present eligible target T(type 1,id 7),count=4,Explosion=1,CollideKill=1;
    independent (owner absent,Pierce=1) and (present linked owner P,hostility(P,T)=true,
    owner 166=67,Pierce=0). Draws: none. Trace: each misses a pierce precondition. Outputs:
    no count read/write,[Event(0,T,missile,null,0,0),Footprint(R,10,10,1,64),Remove],result 2.
  - Discriminates: chance comparison,local seed owner/full carry,current owner gate and explicit
    persistence boundary.
  - runtime_observed: no

- **MMC-V10** Hand-off/final results (MMC-R13/MMC-R14/MMC-R15/MMC-R17; boundary: supplied-input stage)
  - Inputs: default target T(type 1,id 7),owner absent,NextHit0,no event callbacks; all supplied
    operations have no draw. Hit selector 1 exists only where explicitly stated below.
  - Seeds: no local draw; dependency results supplied.
  - Draws: none in this supplied model.
  - Trace: initial act,to-hit branch,event/hit replacement,damage mask,finish,outer dispatcher.
  - (a) Inputs: ToHit1,missile 19=25,hit=false,AlwaysExplode0,CollideKill0. Outputs:
    [HitRoll(null,T,25,1)=false,Event(0,T,missile,null,0,0),Remove]; final 2,no hit/damage/footprint.
  - (b) Inputs: ToHit1,missile 19=25,hit=false,AlwaysExplode1,hit selector 1 returns 4. Outputs:
    [HitRoll(null,T,25,1)=false,Event(0,T,missile,null,0,0),Hit(T)=4]; final 1,no damage/footprint/Remove.
  - (c) Inputs: ToHit0,Explosion0,CollideKill0,hit/damage selector 0. Outputs:
    [Event(0,T,missile,null,0,0),Fill(T),Apply(T)]; final 1,no Remove.
  - (d) Inputs: Explosion1,CollideKill1,hit selector 0. Outputs:
    [Event(0,T,missile,null,0,0),Footprint(R,10,10,1,64),Remove]; final 2,no damage.
  - (e) Inputs: Explosion1,CollideKill1,hit selector 1 returns 2,damage selector 1 returns 99.
    Outputs: [Event(0,T,missile,null,0,0),Hit(T)=2,Fill(T),Damage1(T)=99,Apply(T)]; final 1,
    no footprint/Remove. Handler replaces initial stop mask; damage result ignored.
  - (f) Inputs: target absent,force 1,Explosion0,hit selector 1 returns 0. Outputs:
    [Event(0,null,missile,null,0,0),Hit(null)=0,Footprint(R,10,10,1,64),Remove]; no damage,final 2.
    Independent absent target,force 0,AlwaysExplode0,CollideKill0: final 1,log [].
  - (g) Inputs: direct contact stage,target absent,force=1,room absent,Explosion=0,hit selector=0.
    Draws: none. Outputs: [Event(0,null,missile,null,0,0),Remove],final result 2,no footprint request.
  - Discriminates: miss abort permits continuing,main hit replaces act,damage result ignored,
    force/absent-target skip and final dispatcher removal are asserted.
  - runtime_observed: no

- **MMC-V11** Cooldown requests/failure (MMC-R16/MMC-R17; boundary: supplied-input stage)
  - Inputs: target T(type 1,id 7),NextHit1,NextDelay15,frame 1000,allocation returns fresh L,
    no state 86 at entry.
  - Seeds: no draw.
  - Draws: none.
  - Trace: expiry=1000+15=1015; allocation gates later operations.
  - Outputs: [Allocate(type 2,owner 1/7,1015)=L,Schedule(T,12,1015,null,0,0),ListState(L,86),
    ListCallback(L,contactExpiry),Attach(T,L,1),State(T,86,true)]. Stored attach/expiry not asserted.
  - (a) Inputs: targetT,NextHit1,NextDelay15,frame 1000,allocation absent. Draws: none.
    Outputs: [Allocate(type 2,owner 1/7,1015)=null],no later requests.
  - (b) Inputs: NextHit0,targetT. Draws: none. Outputs: []; independent NextHit1,target absent: [].
  - (c) Inputs: second missile NextHit1,targetT with supplied persisted active state 86=true.
    Draws: none. Outputs: search rejects/direct contact 1 before arming; actual persistence/expiry
    at 1015 remain MMC-Q09.
  - (d) Inputs: target T(type 1,id 7),NextHit=1,NextDelay=0,frame=1000,allocation returns fresh L.
    Draws: none. Outputs: [Allocate(type 2,owner 1/7,1000)=L,Schedule(T,12,1000,null,0,0),
    ListState(L,86),ListCallback(L,contactExpiry),Attach(T,L,1),State(T,86,true)]. Requested expiry
    is 1000; STL-R01 owns actual scheduling/clamping and no stored expiry is asserted.
  - Discriminates: allocation failure omits schedule/state; supplied active state gates other
    missiles independently of unverified attachment.
  - runtime_observed: no

### 8.2 Observation vectors

No gameplay observation was performed. Table reads do not set runtime_observed.

| Id | Measurement | Procedure |
|---|---|---|
| **MMC-O01** | Pierce seed/local steps/count request/readback,chance 60/70/zero | [MMC-O01](../research/spec-evidence/missile-movement-and-collision.md#mmc-o01) |
| **MMC-O02** | Arrow path conditions and actual cell list | [MMC-O02](../research/spec-evidence/missile-movement-and-collision.md#mmc-o02) |
| **MMC-O03** | Unit-before-wall order and barrier's final result | [MMC-O03](../research/spec-evidence/missile-movement-and-collision.md#mmc-o03) |

## 9. Acceptance and dependencies

Partial acceptance is verified local arithmetic/writes,request order and final results in
MMC-V01..MMC-V11. Supplied path/results/bits/cells/lists/sizes/hit operations do not establish
original producers. Persistent count/cooldown assertions remain gated even when their supplied
model passes. Full reproduction also requires MMC-R12c and MMC-Q01/Q03/Q04/Q07/Q08/Q09
integration,plus matching observations for every verified local rule (MMC-Q06). Outside-boundary
MMC-Q02/MMC-Q05 do not gate supplied consumers. Partial acceptance establishes no parity or approval.

**Depends on:** NRF-R01/NRF-R03a seed,NRF-R11a/NRF-R11b velocity/direction,
NRF-R12a/NRF-R12b/NRF-R12c state/stat reads,NRF-Q20 writes; GDF-R07/GDF-R11a/GDF-R12 names;
MCI-R03 room,MCI-R07/MCI-R18a/MCI-R20 speed,MCI-R12b event,MCI-R14/MCI-R21 lifetime,MCI-R26a
pierce; PVP-R05a/PVP-R07 hostility/tier/owner; SBA-R01..SBA-R08 to-hit; IEP-R07/IEP-R09a event;
STL-R01/STL-R06a/STL-R06b/STL-R06c/STL-R07 timer/list/attach/stat/state;
MHD-R18/MHD-R20/MHD-R21/MHD-R23/MHD-R25/MHD-R26 hand-off/interiors. All retain their gates;
request logs do not close integration.

**Depended on by:** MHD timing/termination,MCI post-creation,Blade Sentinel/Fury cooldown.
Reverse citations are not delivery prerequisites.

## 10. Related specs

- [missile-creation-and-initialization](missile-creation-and-initialization.md): creates consumed inputs.
- [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md): owns hit/damage interiors.
- [missile-damage-parameters](missile-damage-parameters.md): owns creation damage values.
- [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md): owns to-hit/defenses.
- [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md): owns event callbacks.
- [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md): owns hostility/attribution.
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): owns timed lists/expiry.
- [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md): shares velocity primitives.
- [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md): consumes NextHit.
- [combat-knockback](combat-knockback.md): owns later displacement.

## Appendix A. Provenance

Original Lord of Destruction Expansion 1.13c Windows x86 D2Game/D2Common analysis artifacts are
identified in the [source manifest](../research/source-manifest.md). Compatible readings from
2026-10-12 and dependency records through 2026-10-15 retain dates/limits. Targeted instruction
and original-image/table checks on 2026-10-15 support the corrections in the
[companion](../research/spec-evidence/missile-movement-and-collision.md). Effective-loader records
retain profile/override limits. No gameplay observation or other-build equivalence is established.
