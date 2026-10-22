# Durability loss

**Description**

Weapons and armor can lose durability during combat. A player's weapon can wear down when attacking, while a hit on a player can wear down one equipped piece of armor. Weapons can defer their break handling, and eligible items with replenishment can recover points over time; the complete effects of breaking equipment remain partly unresolved.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: DUR-R, DUR-Q, DUR-V, DUR-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/item-durability-loss.md)
Open evidence gaps: [gap file](../plans/gaps/item-durability-loss.md)

## 1. Scope

This contract covers combat loss on a player attacker's selected weapon and a player defender's weighted armor choice, skill-driven weapon loss, local durability and quantity write requests, immediate and deferred break requests, and replenishment operations reached by these paths. It covers local effects of indestructibility, no-durability records and ethereal items. Vendor repair, initial creation and quiver consumption are outside scope. DPO-R24a owns placement among hit stages. Equipment selection, stat mutation, notifications and complete break effects have explicit integration gates; supplied-input operations remain separately available.

## 2. Trigger and ownership

D2Game server code computes loss, break routing and replenishment requests; D2Common supplies item predicates and stat operations. The queued combat path enters after damage application and marks and before item events 7/3, using the hit test taken before application (DPO-R24a). It runs attacker processing then defender processing once per reached entry. Clearing hit during application does not skip this stage.

The skill consumer receives a selected weapon and evaluated `Skills.calc2`/`calc3` after a successful skill hit. Its original start-table binding is established; named skill rows and the complete producing hit path remain DUR-Q03. Deferred scanning is called from melee/throw attack paths and player mode end. Default item timer event type 3 routes to replenishment; scheduler lifecycle integration remains DUR-Q04. Server requests item-stat notifications for the relevant player's client; wire encoding and client handling are outside scope. Full binary chains are in the companion.

## 3. Inputs and data

Stats are signed i32 layer-0 values read through RLI-R02a/RLI-R02b. Arithmetic uses signed i32 wrapping values and truncation toward zero under the NRF notation conventions (core-numeric-and-rng-foundations section 1: low-32-bit addition/subtraction/multiplication, signed `/` truncating) unless stated. Draws use NRF-R01/R02. `WriteStat` means a replacing-write request, not verified stored mutation (DUR-Q07 / RLI-Q03). Logical operation names impose no API or storage design.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| quantity | stat 70 `quantity` | i32 | Current stack |
| durability / maximum | stats 72 `durability`, 73 `maxdurability` | i32 | Current and cap; `maxDur` reads the latter |
| indestructibility | stat 152 `item_indesctructible` | i32 | Positive prevents qualification; spelling follows data |
| replenish rates | stats 252 `item_replenish_durability`, 253 `item_replenish_quantity` | i32 | Timer divisors |
| extra stack | stat 254 `item_extra_stack` | i32 | Added to base cap |
| item record | weapons/armor/misc `durability`, `nodurability`, `stackable`, `maxstack` | first three u8; maxstack i32 | Prepared qualification/stack inputs |
| type / throwable results | equivalence with Weapon `weap`, Any Armor `armo`; `ItemTypes.Throwable` | supplied booleans / u8 | Producer integration DUR-Q06 |
| item / owner | presence, kind, inventory, client availability | references / booleans | Eligibility and recipient |
| equipment | primary-weapon result; body-location items; scan entries/node kind | references / integer | Supplied selectors and traversal |
| broken marker | flag set by break operation | boolean | Repeat-break and replenishment gate |
| stack fallback selector | item-local selector word | u16 | Zero/nonzero only; meaning DUR-Q01 |
| stack cap | supplied base maxstack plus stat 254, capped above at 511 | i32 | `maxStack`; no lower clamp |
| frame / timer | current frame, pending item type-3 event, due event/explicit callback | i32 frames / logical state | Supplied scheduler inputs |
| skill | evaluated `calc2` chance, `calc3` amount, selected stack-spend item | i32 / reference | Skill consumer inputs |
| break / refresh operations | callback results, owner/inventory relationships | supplied operations | Log requests; effects/draws DUR-Q02/Q07 |
| expansion | game edition flag | boolean | Throwable combat threshold |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Listed stat names/ids | table data: effective `data/global/excel/itemstatcost.txt`, explicit ID and named Stat cells | Original code fixes numeric consumers; [data records and limits](../research/spec-evidence/item-durability-loss.md#table-records) |
| 45 Weapon/weap; 50 Any Armor/armo | table data: effective `data/global/excel/itemtypes.txt`, zero-based data row excluding header, ItemType/Code | Original code fixes type arguments; equivalence DUR-Q06 |
| Locations 1 head, 3 tors, 4 rarm, 5 larm, 8 belt, 9 feet, 10 glov | table data: effective `data/global/excel/bodylocs.txt`, zero-based data row excluding header, Code | Code-fixed arguments; inventory correspondence DUR-Q06 |
| Ordered (location, weight): (1, 3), (3, 5), (4, 4), (5, 4), (8, 2), (9, 2), (10, 2); count 7 | engine code / binary constant data | Fixed by original code; all eligible total 22 |
| Armor 10, weapon 4, Expansion throwable 10; decrement 1; numerator 2500; later floor 125; cap 511 | engine code | Fixed by original code |
| Broken marker; event 3; node kind 3; Neutral mode 1 | engine code; Neutral name GDF-R05a; mode set RLI-R04a | Fixed original selectors; effects gated below |
| Throwable, durability, nodurability, stackable, maxstack, `calc2`, `calc3` fields | table schema and code loader/consumer | Prepared-data inputs; GDA-R11a/SCX own calc schema/evaluation |
| Stack fallback selector meaning | unresolved | DUR-Q01 |

Named-cell evidence retains its recorded profile limits. Text-file cells do not establish prepared-record correspondence, type equivalence or gameplay execution. Synthetic fixtures supply prepared inputs explicitly.

## 4. Rules (execution order)

**DUR-R01** Qualification [verified in 1.13c code (static)]. A present item qualifies iff kind=item(4), record present, nodurability=0, record durability!=0, maxDur!=0 and stat 152<=0. No draw. Qualification-gated paths skip their chance roll on failure. The standalone R13a stack branch does not qualify again.

**DUR-R02** Placement/order [verified in 1.13c code (static)]. Consume DPO-R24a's pre-apply hit gate and placement. Call R03 before R04; no other direct draw. Dependency interiors can draw separately; no total hit-window count is asserted.

**DUR-R03** Attacker [verified in 1.13c code (static)]. Player kind 0 with present supplied primary weapon calls R05 with that item and attacker seed owner. Other kinds, including monsters/hirelings, do nothing. Original hand/skill selection is DUR-Q06.

**DUR-R04** Defender choice [verified in 1.13c code (static)]. For player defender with inventory, retain present armor-type items in the seven locations in section 3 order and sum weights W. W<=0 returns without draws. Otherwise draw defender modulo 7 for start index s, then modulo W for remainder r (NRF-R02 mask form for powers of two). Walk cyclically from s, skipping ineligible entries. At each eligible entry, r<weight selects it and calls R05 with defender seed owner; otherwise subtract weight and continue. Comparisons are signed i32; r starts in [0, W). Qualification is tested after selection, so indestructible/no-durability armor still contributes weight and consumes both draws. A shield participates when its supplied type test says armor.

**DUR-R05** Chance [verified in 1.13c code (static)]. Compute supplied armor/weapon type results; return without draw if neither or R01 fails. Threshold is 10 for armor, otherwise 4 if not throwable, otherwise return before draw in Classic, otherwise 10 in Expansion. Draw seed owner modulo 100 once; roll>=threshold returns unchanged, else enter R06/R07/R08. No ten-draw loop.

**DUR-R06** Positive remainder [verified in 1.13c code (static)]. On successful chance compute d=durability-1. If signed d>0, replace by maxDur when d>=maxDur; call R11; request WriteStat(item, 72, d, 0); request item-stat notification(72, d) for seed owner's player client; return. Non-player/absent owner has no recipient. Stored/scheduler/message integration is DUR-Q04/Q07.

**DUR-R07** Non-weapon zero [verified in 1.13c code (static)]. For d<=0, not weapon and broken marker clear, request R09 and return. An already-broken non-weapon continues R08. No further direct draw.

**DUR-R08** Zero fallback.

- **DUR-R08a** Local requests [verified in 1.13c code (static)]. Selector 0 requests durability 0 and notification(72, 0), then returns immediately. Otherwise read quantity q. If q>1, request quantity q-1, notify(70, q-1), request durability maxDur, notify(72, maxDur). If q<=1, request durability 0, notify(72, 0), request quantity 0, notify(70, q) carrying old q, then request owner weapon refresh. After either nonzero-selector branch call R11 iff maxStack>0. All stat requests use layer 0 and notifications use seed owner's player client. For supplied base cap and extra-stack value, maxStack=min(i32(base+extra),511), with no lower clamp. No immediate break or direct draw; effects DUR-Q07.
- **DUR-R08b** Selector meaning [unresolved]. Original semantic producer of the selector remains DUR-Q01. Supplying its u16 value determines R08a without establishing original item-category eligibility.

**DUR-R09** Break.

- **DUR-R09a** Request order [verified in 1.13c code (static)]. Given item, owner, game: request broken marker on; inventory break preparation(owner inventory, item); item break predicate(item). If predicate=1, request conditional equipment transition(game, owner, item, 1, 1). Then request equipment revalidation(game, owner, 0); inventory break completion(owner inventory, item); owner update(owner, 1); durability write(72, 0, layer 0); item-stat notification(72, 0) to owner's player client; equipment notification(owner, item, 1, 1); owner finalization(owner). These neutral callback names distinguish operations without assigning unknown effects. No direct draw in this routing entry.
- **DUR-R09b** Integration [unresolved]. Callback stat/equipment effects, owner bookkeeping, client delivery and interior draws remain DUR-Q02/Q07. A supplied log does not prove stored broken equipment behavior or a draw-free break.

**DUR-R10** Deferred scan [verified in 1.13c code (static)]. For player owner, walk supplied inventory entries. On node kind 3, obtain item and test weapon, R01 qualification, durability<=0, broken clear. Before the first match request Neutral mode 1; request R09 and restart from head after each break. Request mode only once per scan. No direct draw. Actual traversal, mode effects and break mutation/termination are DUR-Q07; fixtures supply their changes.

**DUR-R11** Initial scheduling [verified in 1.13c code (static)]. Read rate 252; zero or supplied pending item event 3 returns. Otherwise request event 3 for item at frame+trunc(2500/rate)+1, default callback, zero auxiliary arguments. No 125 floor or direct draw. This is a requested frame, with scheduler execution DUR-Q04. R13b uses the identical operation with rate 253. Creation/transfer callers are outside scope.

**DUR-R12** Replenishment.

- **DUR-R12a** Tick [verified in 1.13c code (static)]. If broken clear and R01 qualifies, run regen(252, 72, maxDur); return if nonzero. Otherwise, if stackable!=0, run regen(253, 70, maxStack). Broken items skip durability regeneration. regen(rateStat, target, cap): read rate; zero returns 0. Else n=Stat(target)+1; signed n>cap returns 1 without effects. Otherwise request write(target, n, 0); look up owning player/client; if found request notification(target, n), then inside that same successful client branch request item inventory refresh iff n>=1 and broken set. Request default event 3 at frame+max(trunc(2500/rate)+1, 125), return 1. No direct draw; actual effects/interior refresh draws DUR-Q07.
- **DUR-R12b** Default selection [verified in 1.13c code (static)]. A supplied due event 3 without explicit callback selects this handler and supplies game/event item. An explicit callback bypasses default selection. No direct selection draw; lifecycle DUR-Q04.

**DUR-R13** Skill loss.

- **DUR-R13a** Chance/amount [verified in 1.13c code (static)]. Supplied weapon, attacker, chance, amount: return unless weapon type. If stackable, quantity<=0 returns without draw; otherwise draw attacker modulo 100, request R13c stack-spend selection iff roll<chance, return. If nonstackable require R01, draw attacker modulo 100, return if roll>=chance. Else d=durability-amount; signed d>0 requests write(72, d, 0) and notification(72, d) for attacker's player client, without cap/replenish scheduling; d<=0 requests R09 immediately even for weapons. The qualifying caller passes `calc2` as chance and `calc3` as amount, consuming SCX-R11 evaluation; complete hit/selection producers DUR-Q03/Q06.
- **DUR-R13b** Supplied stack-spend item [verified in 1.13c code (static)]. An absent selected item returns 0 without effects. Otherwise compute wrapping i32 n=quantity-1 and remember whether signed n<0; if so replace n by 0. Request quantity write n, R11 scheduling using rate 253, then quantity notification. If durability differs from maxDur, request durability maxDur and notification. Request owner weapon refresh. Return 0 if subtraction was negative, otherwise wrapping i32 n+1; no direct draw. This does not prove the selected item equals the chance-tested weapon.
- **DUR-R13c** Stack selection [unresolved]. The repeated weapon/arm selector can choose a different item for R13b. Exact original selected item and producer integration remain DUR-Q06; supplied selected-item requests do not settle them.

**DUR-R14** Exceptions [verified in 1.13c code (static)]. Positive stat 152, nonzero nodurability, zero record durability or zero stat 73 prevents qualification and its loss roll on qualification-gated local paths. Ethereal identity is not tested in local qualification/chance/arithmetic consumers, so it does not alter their supplied-input results. This does not claim unread selection/stack/break interiors ignore all other flags. Only R04/R05/R13a's reached draws occur locally.

**DUR-R15** Death effects [unresolved]. A death-triggered durability operation and its effects have not been established. No universal no-loss-on-death assertion follows; restore/set writers outside this boundary remain DUR-Q05.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| DUR-R01 | Qualification | verified in 1.13c code (static) | DUR-Q06: producers | DUR-V06 |
| DUR-R02 | Placement/order | verified in 1.13c code (static) | none locally | DUR-V01 |
| DUR-R03 | Supplied attacker weapon | verified in 1.13c code (static) | DUR-Q06: selection | DUR-V01, DUR-V02 |
| DUR-R04 | Weighted choice | verified in 1.13c code (static) | DUR-Q06: equipment/type | DUR-V03, DUR-V04, DUR-V06 |
| DUR-R05 | One chance draw | verified in 1.13c code (static) | DUR-Q06: type/throwable | DUR-V02, DUR-V05 |
| DUR-R06 | Positive-loss requests | verified in 1.13c code (static) | DUR-Q04/Q07: effects | DUR-V07, DUR-V08 |
| DUR-R07 | Immediate armor break request | verified in 1.13c code (static) | DUR-Q02/Q07: break | DUR-V09 |
| DUR-R08a | Zero fallback requests | verified in 1.13c code (static) | DUR-Q01/Q04/Q07: selector/effects | DUR-V10, DUR-V11 |
| DUR-R08b | Selector meaning | unresolved | DUR-Q01: original binding | none |
| DUR-R09a | Break request order | verified in 1.13c code (static) | DUR-Q02/Q07: callbacks | DUR-V09 |
| DUR-R09b | Complete break effects | unresolved | DUR-Q02/Q07: effects/draws | none |
| DUR-R10 | Deferred scan | verified in 1.13c code (static) | DUR-Q07: traversal/mode/break | DUR-V10 |
| DUR-R11 | Initial due-frame request | verified in 1.13c code (static) | DUR-Q04: scheduler | DUR-V07, DUR-V12 |
| DUR-R12a | Tick arithmetic/requests | verified in 1.13c code (static) | DUR-Q07: writes/client/refresh | DUR-V12 |
| DUR-R12b | Default timer selection | verified in 1.13c code (static) | DUR-Q04: lifecycle | DUR-V12 |
| DUR-R13a | Skill chance/amount | verified in 1.13c code (static) | DUR-Q03/Q06: producers; DUR-Q02/Q07: effects | DUR-V13 |
| DUR-R13b | Supplied stack item | verified in 1.13c code (static) | DUR-Q04/Q07: effects | DUR-V13 |
| DUR-R13c | Actual stack selection | unresolved | DUR-Q06: item identity | none |
| DUR-R14 | Local exceptions | verified in 1.13c code (static) | DUR-Q06/Q07: integration | DUR-V06 |
| DUR-R15 | Death effects unestablished | unresolved | DUR-Q05: outside restore/death paths | none |

## 5. Outputs and side effects

| Output | Local contract | Rule |
|---|---|---|
| Qualification result | Boolean true exactly when all qualification conditions hold, otherwise false | DUR-R01 |
| Weighted selection | Selected eligible item passed to the loss check; no selected-item call when W<=0 | DUR-R04 |
| Regen helper return | 0 for zero rate; otherwise 1, including the already-full branch with no write/reschedule | DUR-R12a |
| Default event handler selection | Event 3 without explicit callback selects the replenish handler; an explicit callback bypasses it | DUR-R12b |
| Stack-spend helper return | 0 for absent item or negative decremented quantity; otherwise wrapping i32 n+1 after decrement, equivalent to prior quantity when no clamp occurred | DUR-R13b |
| Seed changes | Named owner's NRF steps in order | DUR-R03/R04/R05/R13a |
| Durability write request | Stat 72, layer 0, computed value, max or 0 | DUR-R06/R08a/R09a/R12a/R13a/R13b |
| Quantity write request | Stat 70, layer 0, decrement/floor/zero | DUR-R08a/R12a/R13b |
| Item-stat notification request | Named notification, stat/value, player-client recipient; R08a can send old quantity | DUR-R06/R08a/R09a/R12a/R13a/R13b |
| Broken/equipment/owner requests | Ordered operations; effects gated | DUR-R09a/R09b |
| Weapon/inventory refresh request | Stated branch; interior effects/draws gated | DUR-R08a/R12a/R13b |
| Neutral mode request | Once before first deferred break | DUR-R10 |
| Timer request | Item event 3, default callback, due frame, zero auxiliaries | DUR-R11/R12a |

No persistence, unit creation/destruction or wire-format contract is established. Supplied fixture writes persist within that fixture only; actual propagation/refresh/scheduling/delivery remains gated.

## 6. Branches and edge cases

- Player versus player can run both parts with separate seeds. Monster/hireling versus player runs defender only; monster versus monster has no local durability draw.
- Classic throwables return before combat draw; Expansion threshold 10. Difficulty and melee/missile add no branch inside these consumers; eligibility belongs upstream.
- Empty/ineligible armor locations give W=0; neck/rings are absent from this table. All seven eligible locations total 22.
- R06 clamps above maximum; R13a does not. Zero/negative skill amount can raise durability without cap.
- Already-broken non-weapons at zero follow R08a. Selector 0 returns without refresh/scheduling.
- Rate 0 does nothing. Negative initial rate can request a frame at/before current; actual scheduler execution is DUR-Q04. Later requests apply floor 125.
- A tick from cap-1 writes cap and reschedules; a tick from cap stops. Nonzero durability regen result, including already-full, suppresses quantity regen.
- Missing qualification record fails R01. Other helpers require valid items/inventories/prepared records; malformed producer inputs remain DUR-Q01/Q06/Q07 without guessed defaults.

## 7. Open questions

| Question | Status / kind | Missing evidence / boundary |
|---|---|---|
| **DUR-Q01** | open / integration | Semantic producer of fallback selector; base maxstack column identity is established, but effective data/prepared correspondence and original selector binding remain open. R08a accepts supplied selector/cap |
| **DUR-Q02** | open / integration | Break/refresh callback chains and player-owner interior draws; local routing is draw-free, complete effects are not |
| **DUR-Q03** | open / outside-boundary | Start selector 7 binding is established; original Skills rows and complete qualifying-hit/calc production need SCD/SCX evidence. None locally for supplied chance/amount |
| **DUR-Q04** | open / integration | Default type 3 binding established; scheduler lookup/insertion/removal, requested-frame normalization and due delivery remain. R11/R12 requests and supplied-event selection available |
| **DUR-Q05** | open / outside-boundary | Fully read restore/set writers and trace death roots before asserting presence/absence of death-time loss |
| **DUR-Q06** | open / integration | Primary hand/skill selector, repeated stack selection, inventory locations, type equivalence/throwable and prepared-record producers. Supplied consumers available |
| **DUR-Q07** | open / integration | Actual stat propagation RLI-Q03, client delivery, owner lookup, refresh, traversal/mode and break semantics. Local requests/arithmetic available |
| **DUR-Q09** | open / observation | Attribute seed delta directly at durability entries with chosen item/roll, damage and get-hit-veto draws. Every local rule still needs attributable gameplay observation |

Attempts and exact remaining gates are in the [gap file](../plans/gaps/item-durability-loss.md). Observation gaps block runtime-confirmation claims independently of integration gates.

## 8. Test vectors

### 8.1 Executable vectors

Each independent case starts with this synthetic supplied-input fixture: present Expansion game frame F=1000, no pending timers; player owner/attacker with inventory/client; monster defender. Item kind 4, record present, record durability 40, nodurability 0, stackable 0; stats 70=0, 72=20, 73=20, 152=0, 252=0, 253=0, 254=0; base stack 0, selector 0, broken false. Supplied weapon true, armor false, throwable false; primary result=item; all armor locations empty. Calling an item armor overrides armor=true and weapon=false; a sword/weapon uses armor=false and weapon=true. Reads return supplied values. Supplied replacing writes persist for subsequent fixture reads; notifications, timer/mode/refresh/break operations log calls and make no unstated writes/draws. Named break fixture effects are supplied changes, not R09b assertions. Timers log requested frames and expose explicitly supplied pending/due events; they do not execute an unknown scheduler. Local draws execute NRF arithmetic. No output includes unlisted local calls/writes/draws. Boundary: supplied-input stage unless stated.

Exact seeds: S1={1, 0x29A}, S57={57, 0x29A}, S12={12, 0x29A}, SX={0x12345678, 0x29A}. First-step states T1={0x6AC6935F, 0}, T57={0xC6363E77, 0x17}, T12={0x014ECBD6, 5}, TX={0x03BA0CF2, 0x0797CA94}. S1 second/third: U1={0x2F2ED81B, 0x2C890AFD}, V1={0x280687C4, 0x13ADFE65}. Cases do not inherit other cases' outputs.

- **DUR-V01** Placement/attacker seed (DUR-R02/R03; boundary: placement plus supplied consumer)
  - Inputs: independent cases a/b/c use default sword 20/20 and monster defender. (a)pre-apply hit true, supplied apply clears hit; marks, defender stage and events 7/3 log order. (b)attacker monster, primary result absent, pre-apply hit true. (c)pre-apply hit false. (d)both attacker and defender are players with inventories; attacker primary sword A is 20/20; defender's only armor is torso item B at 20/max 40; pre-apply hit true and supplied apply clears hit. Both items have zero replenish rates. Supplied writes/notifications log their item/value in the shared effect log alongside apply, marks and events 7/3.
  - Seeds: (a/b/c)attacker S1, defender S57; (d)attacker S57, defender {5,0x29A}.
  - Draws: (a)attacker modulo 100=51, no defender draw; (b/c)none; (d)attacker modulo 100=3, then defender modulo 7=5, modulo 5=3, modulo 100=5.
  - Trace: (a)apply→marks→attacker check(51>=4)→defender stage→events 7/3; (b)both non-player parts skip; (c)pre-apply gate skips apply/marks/durability. (d)attacker loss makes A=19; defender W=5 starts at feet, skips empty feet/glov/head, selects tors because 3<5, then chance 5<10 makes B=19. Defender states after its three draws are {0x15E0D673,2}, {0xC997B681,0x09201110}, {0x134E1255,0x54152EE9}.
  - Outputs: (a)exact order above, no writes, durability 20, attacker T1, defender S57; (b)no attacker check/draw, seeds S1/S57 unchanged; (c)no apply/marks/durability, seeds S1/S57 unchanged, remaining events owned by DPO. (d)exact effect log apply→marks→write(A,72,19,0)→notify(A,72,19)→write(B,72,19,0)→notify(B,72,19)→events 7/3; A/B durability 19, no timer requests; attacker T57, defender {0x134E1255,0x54152EE9}.
  - Discriminates: apply clearing hit must not remove durability; non-player must not run player weapon processing; reversing attacker and defender parts reverses A/B write/notification order and fails d even though the final seeds/stats could agree.
  - runtime_observed: no

- **DUR-V02** Thresholds (DUR-R03/R05)
  - Inputs: independent(a)default sword 20/20, S57; (b)sword, S12; (c)armor true/weapon false, S12.
  - Seeds: owner as named.
  - Draws: owner modulo 100:(a)3, (b)6, (c)6.
  - Trace: 3<4; 6>=4; 6<10; successful losses 20-1=19.
  - Outputs: (a)write/notify 19, seed T57; (b)no requests, durability 20, seed T12; (c)write/notify 19, seed T12. No timer.
  - Discriminates: swapped armor/weapon thresholds fail b/c.
  - runtime_observed: no

- **DUR-V03** Cyclic/equality choice (DUR-R04)
  - Inputs: player defender; qualifying helm head(weight 3) and armor tors(weight 5) only, both 20/20; monster attacker.
  - Seeds: defender S1.
  - Draws: modulo 7=2; mask modulo 8=3; modulo 100=12.
  - Trace: W=8; start rarm; pass empty larm, belt, feet, glov; head 3>=3 leaves 0; tors 0<5 selected; 12>=10.
  - Outputs: tors selected, no writes, both durability 20; selection seed U1, final V1.
  - Discriminates: occupied-slot count or non-strict head comparison chooses differently.
  - runtime_observed: no

- **DUR-V04** Armor table choice (DUR-R04)
  - Inputs: player defender, qualifying armor head, tors, larm, belt, feet, glov; rarm empty; W=18; all 20/20.
  - Seeds: independent(a)defender S1, (b)defender SX.
  - Draws: (a)modulo 7=2, modulo 18=9, modulo 100=12; (b)modulo 7=0, modulo 18=0, modulo 100=19.
  - Trace: (a)rarm empty; larm 9→5, belt→3, feet→1; glov 1<2; (b)head 0<3. Chance fails both.
  - Outputs: (a)glov, seed V1; (b)head, selection seed{0xC437E0CE, 0x018DED5D}, final{0x9A55CBE3, 0x51D75543}; no writes, durability unchanged.
  - Discriminates: fixed start/noncyclic order fails one choice.
  - runtime_observed: no

- **DUR-V05** Classic/Expansion throwable (DUR-R05)
  - Inputs: qualifying throwable weapon 20/20; (a)Classic, (b)Expansion; (c)independent Classic qualifying throwable item 20/20 whose supplied type results are armor=true and weapon=true.
  - Seeds: owner S57 independently for a/b; owner S12 for c.
  - Draws: (a)none; (b)modulo 100=3; (c)modulo 100=6.
  - Trace: Classic returns; Expansion 3<10, d=19; (c)armor selects threshold 10 without the throwable test, 6<10, d=19.
  - Outputs: (a)no requests, durability 20, seed S57; (b)write/notify 19, no timer, seed T57; (c)write/notify 19, no timer, seed T12.
  - Discriminates: unconditional throwable draw changes Classic seed; testing throwable/Classic before armor fails c (no draw, seed S12), and using the weapon threshold 4 for an item that is also armor fails c (6>=4, no write).
  - runtime_observed: no

- **DUR-V06** Qualification/ethereal (DUR-R01/R04/R14)
  - Inputs: (a)player defender with only tors armor, stat 152=1, 20/20; (b-e)isolated qualification with respectively nodurability 1, record durability 0, stat 73=0, item absent; (f/g)isolated armor chance, ethereal false/true, stat 152=0, otherwise defaults; (h)player defender with inventory and every armor location empty.
  - Seeds: (a/h)defender S1; (b-e)no draw; (f/g)owner S57.
  - Draws: (a)modulo 7=2, modulo 5=1, no third; (b-e/h)none; (f/g)modulo 100=3.
  - Trace: (a)W=5 despite indestructibility, select tors, qualification false; (b-e)false; (f/g)true, 3<10, d=19; (h)W=0 returns before selection.
  - Outputs: (a)seed U1, durability 20, no requests; (b-e)false, no requests; (f/g)identical write/notify 19, seed T57; (h)no selected-item call or requests, defender seed S1.
  - Discriminates: pre-weight qualification removes a's two draws; ethereal exception changes g; drawing a start index before the W test changes h's seed.
  - runtime_observed: no

- **DUR-V07** Schedule-before-write (DUR-R06/R11)
  - Inputs: isolated armor loss 15/40, rate 252=5, frame 1000, no pending event.
  - Seeds: owner S57.
  - Draws: modulo 100=3.
  - Trace: 3<10; d=14; delay 2500/5+1=501.
  - Outputs: ordered event(item, 3, 1501, default, zero auxiliaries)→write(72, 14, 0)→notify(72, 14); durability 14, seed T57.
  - Discriminates: scheduling after write or omitting+1 fails log/frame.
  - runtime_observed: no

- **DUR-V08** Maximum clamp (DUR-R06)
  - Inputs: isolated armor loss, stored 30, max 20, rate 0.
  - Seeds: owner S57.
  - Draws: modulo 100=3.
  - Trace: d29>=20→20.
  - Outputs: write/notify 20, no timer; seed T57.
  - Discriminates: subtract-only 29 fails.
  - runtime_observed: no

- **DUR-V09** Immediate break routing (DUR-R07/R09a)
  - Inputs: independent a/b share isolated armor loss 1/40, broken false; callbacks log only; player owner/client; supplied broken-marker write persists. Break predicate returns (a)1 or (b)0.
  - Seeds: owner S57 in each case.
  - Draws: modulo 100=3 in each; none directly in break entry.
  - Trace: d0, nonweapon, clear marker→R09a sequence; predicate controls only the conditional transition.
  - Outputs: (a)broken-on→inventory preparation→predicate→conditional transition(game, owner, item, 1, 1)→revalidation(game, owner, 0)→inventory completion→owner update(owner, 1)→write(72, 0, 0)→notify(72, 0)→equipment notification(owner, item, 1, 1)→owner finalization; fixture broken true, durability 0, seed T57. (b)broken-on→inventory preparation→predicate→revalidation(game, owner, 0)→inventory completion→owner update(owner, 1)→write(72, 0, 0)→notify(72, 0)→equipment notification(owner, item, 1, 1)→owner finalization; fixture broken true, durability 0, seed T57. Actual break effects/interior draws not asserted.
  - Discriminates: deferred armor break or reversed arguments fail the log; a requires the conditional transition and b forbids it.
  - runtime_observed: no

- **DUR-V10** Stored zero/deferred scan (DUR-R08a/R10)
  - Inputs: (a)isolated sword loss 1/30, quantity 0, selector 0, rate 252=5, rate 253=25, base maxstack 100 and stat 254=0; then separately invoke scan on player with ordered entries A/B. A/B and a supplied entry C initially outside the list are node kind 3, qualified weapons, durability 0, broken false. Supplied break A marks A broken and prepends C, giving list C/A/B; subsequent break calls mark their item broken without other list changes. (b)independent fallback stage, selector 1, quantity 1, maxStack 0, weapon at zero. (c)independent isolated armor loss 1/40, broken true, selector 0, quantity 0. (d)independent scan on a player with ordered node-kind-3 qualified unbroken weapons D (durability 5/max 30) and E (durability 1/max 30). (e)two independent scans, one on a monster owner and one on an object-type (unit type 2) owner, each whose supplied entry list holds A' = node kind 3, qualified weapon, durability 0, broken false.
  - Seeds: (a)owner S57 through loss, then no draw; (b)no draw; (c)owner S57; (d/e)no draw.
  - Draws: (a/c)modulo 100=3; (b/d/e)none.
  - Trace: (a)selector 0 writes zero, returns; scan A→break/prepend C→restart→C→break/restart→B→break/restart→end. (b)q<=1 uses old q in message. (c)d=0, already-broken nonweapon continues to selector-zero fallback. (d)D and E fail durability<=0; the walk ends with no match. (e)each non-player owner returns before walking.
  - Outputs: (a)write/notify durability 0, no immediate break/refresh/timer, seed T57; scan log Neutral 1→break A→break C→break B, one mode request, final list C/A/B with all three broken. (b)write/notify(72, 0)→write(70, 0)→notify(70, 1)→weapon refresh, no timer, quantity 0. (c)write/notify(72, 0), no break/refresh/timer, marker remains true, quantity 0, owner seed T57. (d)no mode request and no break request; D/E durability 5/1, unbroken. (e)for both owners, no mode request and no break request; A' durability 0, unbroken.
  - Discriminates: selector 0 must not refresh or schedule despite positive rates; mode once; a forward-only scan omits C or processes B before C; old quantity 1 differs from stored 0; c forbids a repeated immediate break; an unconditional Neutral request at scan start fails d; omitting the player-owner guard, or excluding only monsters, fails e (Neutral 1→break A').
  - runtime_observed: no

- **DUR-V11** Stack fallback (DUR-R08a)
  - Inputs: (a)Expansion throwable, selector 1, quantity 40, durability 1/max 30, base maxstack 100, stat 254=0, rate 252=5, rate 253=25, frame 1000, no pending event; b/c/d independently execute isolated maxStack calculation with (base, stat 254) respectively (600, 5), (-2, 0), (2147483647, 1).
  - Seeds: (a)owner S57; (b/c/d)no draw.
  - Draws: (a)modulo 100=3; (b/c/d)none.
  - Trace: (a)3<10; d0; q40>1→39; refill 30; maxStack>0 calls R11 with rate 252=5, requesting 1000+2500/5+1=1501; (b)600+5=605 exceeds 511; (c)-2+0=-2 is below 511; (d)wrapping i32 sum is -2147483648, also below 511.
  - Outputs: (a)exact log quantity write 39→notify(70,39)→durability write 30→notify(72,30)→event 3 request 1501; no refresh/break; seed T57; (b)maxStack 511, no writes/calls/seed changes; (c)maxStack -2, no writes/calls/seed changes; (d)maxStack -2147483648, no writes/calls/seed changes.
  - Discriminates: durability-only decrement, omitted scheduling or quantity-rate scheduling fails a; rate 253 would request 1101 instead of 1501, and scheduling before the writes fails the log. Omitting the upper cap fails b; adding a lower clamp fails c; a wide/saturating sum or unsigned cap comparison fails d.
  - runtime_observed: no

- **DUR-V12** Cadence/default selector (DUR-R11/R12a/R12b)
  - Inputs: all cases share armor 14/max 40, rate 252=5, frame 1000, except overrides below. Independent cases: (a)none; (b)rate 252=25; (c)rate 5, pending true; (d)armor 40/max 40, rate 5, stackable 1, quantity 3, base maxstack 10, stat 254=0, rate 253=5; (e)broken stackable, durability 0/max 40, quantity 3, base maxstack 10, stat 254=0, rate 253=5, owner player/client present; (f)same as e, client absent; (g)rates 252/253=0; (h)due event type 3 has explicit callback K, which logs K once and makes no writes, requests or draws; (i)rate 252=-5, no pending event; (j)stat 152=1, stackable 1, quantity 3, base maxstack 10, stat 254=0, rate 253=5, broken false; (k)durability 39/max 40; (l)qualified unbroken armor, rate 252=0, stackable 1, quantity 3, base maxstack 10, stat 254=0, rate 253=5; (m)rate 252=6; (n)rate 252=-6; (o)qualified unbroken armor, rate 252=0, stackable 1, quantity 9, base maxstack 10, stat 254=0, rate 253=5. Actions: a/b invoke initial scheduling then the stated due ticks; c/i/m/n invoke initial scheduling only; d/e/f/g/h/j/l execute one due tick at frame 1000; k/o execute supplied due ticks at 1000 and 1501. Except h, supplied due events are default type 3 without explicit callback.
  - Seeds: no draw; all unchanged.
  - Draws: none locally; supplied refresh logs none.
  - Trace: (a)initial 501; due 1501→15, due 2002→16; (b)initial 101, due 1101→15, later floor 125; (c)pending suppresses; (d)41>40 returns 1, suppresses quantity; (e/f)broken skips durability, quantity 3+1=4; (g)zero rate returns 0; (h)explicit K bypasses the default; (i)trunc(2500/-5)+1=-499, requested frame 501; (j)qualification false skips durability, quantity 3+1=4; (k)39+1=40 writes/reschedules, then 40+1=41 exceeds cap and stops; (l)durability regen returns 0, quantity regen adds 1; (m)trunc(2500/6)=416, delay 417; (n)trunc(2500/-6)=-416, delay -415; (o)durability regen returns 0, quantity 9+1=10 reaches its cap; next tick quantity 10+1=11 exceeds cap 10, despite maxDur 40.
  - Outputs: (a)requested 1501, 2002, 2503, tick write/notify 15 then 16, final durability 16; (b)requested 1101 then 1226, tick write/notify 15, final durability 15; (c)no event request; (d)no writes/reschedule, quantity 3; (e)quantity write 4→notify(70, 4)→inventory refresh→request 1501; (f)quantity write 4→request 1501, no notification/refresh; (g)no writes/events, durability 14; (h)log exactly K, no regen calls/writes/events, durability 14; (i)one event request at 501, no writes/notifications; (j)quantity write 4→notify(70, 4)→request 1501, no inventory refresh, durability 14; (k)first tick write/notify(72, 40)→request 1501, second tick no write/notification/event, final durability 40; (l)durability regen result 0, quantity regen result 1, quantity write 4→notify(70, 4)→request 1501, no inventory refresh, durability 14; (m)one event request at 1417, no writes/notifications; (n)one event request at 585, no writes/notifications; (o)first tick quantity write 10 -> notify(70,10) -> request 1501, second tick no write/notification/event; final quantity 10, durability 14 unchanged. The frames in i/n are requests, not asserted actual past delivery. Reached regen helper returns 1 in a/b/d/e/f/j/k, 0 for the zero-rate call in g, and the separate durability/quantity results 0/1 in l and at each o tick; c/h/i/m/n never enter regen. Default due events select R12a, while h selects K. Continued supplied rate 5 ticks add 1 every 501 until 40.
  - Discriminates: initial 125 floor fails b; quantity fallthrough fails d; refresh without client fails f; explicit callback must bypass default in h; clamping a negative requested frame inside R11 fails i/n; requiring durability qualification for quantity regen fails j; scheduling after an already-full tick or stopping before the write reaching cap fails k; ignoring the zero durability-regen return suppresses l's quantity write; rounding up fails m and flooring negative division fails n; using maxDur 40 as the quantity cap instead of maxStack 10 writes quantity 11 on the second o tick and fails.
  - runtime_observed: no

- **DUR-V13** Skill amount/stack selection (DUR-R13a/R13b)
  - Inputs: a/b/c/e share nonstack weapon 10/10, chance 50, amount 3. Independent overrides: (a)rate 252=5, rate 253=25; (b)durability 3/max 10; (c)amount=-3; (d)direct R13a consumer entry with stackable tested weapon quantity 40, durability 1/max 30, chance 50, amount 3, stat 152=1 (unqualified); selected spend item explicitly different B, quantity 8, durability 5/max 20, rate 253=5, frame 1000; (e)S1 instead of S57; (f)stackable tested weapon quantity 0, chance 50, amount 3; (g)stackable tested weapon quantity 40, durability 1/max 30, chance 50, amount 3. Cases h-k call R13b directly, bypassing the outer chance/quantity guards: (h)selected item absent; (i)selected item quantity 0, durability 20/max 20; (j)selected item quantity -2, durability 20/max 20; (k)selected item quantity -2147483648, durability 20/max 20. Cases l/m call R13a directly with nonstack item 10/10, chance 50, amount 3: (l)armor=true,weapon=false; (m)weapon=true,armor=false,stat 152=1 (unqualified). (n)qualifying caller entry with nonstack weapon 10/10 and supplied evaluated results for the selected skill `calc1`=2, `calc2`=50, `calc3`=3, `calc4`=1. Rates not explicitly overridden are 0.
  - Seeds: (a-d/l/m/n)attacker S57; (e-k)attacker S1.
  - Draws: (a-d/n)modulo 100=3; (e/g)modulo 100=51; (f/h/i/j/k/l/m)none.
  - Trace: (a)10-3=7; (b)3-3=0→break; (c)10-(-3)=13; (d)stack branch bypasses qualification and successful chance reaches selector B, 8-1=7, refill 5→20; (e/g)51>=50; (f)empty returns; (h)absent returns 0; (i)0-1=-1 clamps to 0; (j)-2-1=-3 clamps to 0; (k)wrapping decrement gives 2147483647, nonnegative; wrapping n+1 gives -2147483648; (l)weapon-type guard returns before the would-be roll 3; (m)qualification guard returns before the would-be roll 3; (n)caller passes chance=`calc2` 50 and amount=`calc3` 3, 3<50, 10-3=7.
  - Outputs: (a)write/notify 7, no schedule; (b)immediate break request(weapon, attacker, game), actual effects gated; (c)write/notify 13, no cap/schedule; (d)B quantity write 7→quantity timer request 1501→notify 7→B durability write 20→notify 20→owner weapon refresh; stack helper returns 8; tested weapon quantity 40/durability 1 unchanged; seed T57 for a-d. (e)no writes, seed T1; (f)no requests, seed S1; (g)no selector/spend/write/notification/refresh/event requests, tested item quantity 40/durability 1, seed T1; (h)return 0, no requests, seed S1; (i)return 0, quantity write 0→notify(70, 0)→owner weapon refresh, no durability write/event, durability 20, seed S1; (j)return 0, quantity write 0→notify(70, 0)→owner weapon refresh, no durability write/event, durability 20, seed S1; (k)return -2147483648, quantity write 2147483647→notify(70, 2147483647)→owner weapon refresh, no durability write/event, durability 20, seed S1; (l)no requests/draws, durability 10, seed S57; (m)no requests/draws, durability 10, seed S57; (n)write/notify 7, no schedule, seed T57. Supplied calc caller passes chance 50, amount 3 in that order for a/b; whole skill damage execution is not asserted.
  - Discriminates: one-point loss or adding replenish scheduling despite positive rate 252 fails a; deferred skill break fails b; cap fails c; same-item assumption or an added stack qualification guard fails d; skipping failed draw fails e/g; applying the outer empty-stack guard to direct R13b calls incorrectly suppresses i's write/refresh; saturating rather than wrapping the decrement fails k; omitting the weapon guard fails l and omitting nonstack qualification fails m; n rejects `calc1` as chance (3>=2, no write), `calc4` as chance (3>=1, no write), `calc4` as amount (write 9), `calc1` as amount (write 8) and swapped chance/amount (3>=3, no write).
  - runtime_observed: no

### 8.2 Observation vectors

- **DUR-O01** Pending: seed steps across 1000 player-versus-monster/monster-versus-player eligible hits; expect 1 weapon chance step and 2 selection plus 0/1 armor chance steps, accounting separately for other draws. [Procedure](../research/spec-evidence/item-durability-loss.md#dur-o01).
- **DUR-O02** Pending: chosen-item losses near 10% armor/4% weapon; frequency supports but does not establish exact thresholds. [Procedure](../research/spec-evidence/item-durability-loss.md#dur-o02).
- **DUR-O03** Pending: sword stored 0 followed by next scan's break and resulting equipment behavior. Continued usability before scan is an observation target, not an established break-effect assertion. [Procedure](../research/spec-evidence/item-durability-loss.md#dur-o03).
- **DUR-O04** Partial observation: on 2026-10-12, level 1 Barbarian with only starting Buckler, attacked by 4 Fallen: 61 hit transitions, 34 blocked used 1 defender step, 27 unblocked used 4, misses used 0. Additional 3 fit 2 selection plus 1 chance draw for W=4; chosen item, roll, damage and durability were not captured. Get-hit attribution remains DUR-Q09. [Record/reproduction](../research/spec-evidence/item-durability-loss.md#dur-o04). `runtime_observed: yes (seed deltas only, 1.13c, 2026-10-12)`; no rule is runtime-confirmed.

## 9. Acceptance and dependencies

Partial local acceptance requires every exact local arithmetic, seed, selection, return and ordered-request assertion of DUR-V01–DUR-V13 for verified sub-ids in the Id map. Supplied effects do not establish stored durability/quantity, broken equipment, interior draws or actual timer cadence. R08b/R09b/R13c/R15 have no verified full-result acceptance. DUR-Q01/Q02/Q04/Q06/Q07 block their named integrations; Q03/Q05 are outside producers and Q09 limits gameplay confirmation. Full 1.13c reproduction additionally requires resolving applicable 1.13c gates and attributable gameplay observations. Passing supplied subsets is partial acceptance.

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) NRF-R01/R02(seed/draw), section 1 notation (integer arithmetic); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) RLI-R02a/RLI-R02b(reads), RLI-R04a(mode set requested by DUR-R10) with its mode-bound list processing RLI-R04b and their gates; [core-game-data-fields](core-game-data-fields.md) GDF-R05a(player mode 1 Neutral identity); [core-game-data-table-records](core-game-data-table-records.md) GDA-R11a(calc fields), GDA-R14(ItemTypes Throwable field schema with its limits); [combat-damage-pipeline-order](combat-damage-pipeline-order.md) DPO-R24a(placement/pre-apply hit; armor type 50/weapon type 45 tests); [skill-cast-start-and-do-dispatch](skill-cast-start-and-do-dispatch.md) SCD-R05a(start-table dispatch; the selector 7 binding is this spec's own evidence, DUR-Q03); [skill-calc-expressions](skill-calc-expressions.md) SCX-R11(evaluated calcs with its limits). Replacing mutation remains RLI-Q03, not a getter guarantee. The inventory-entry scan of DUR-R10 is owned here and gated by DUR-Q07; no dependency owns it.

**Depended on by:** DPO-R24a consumes DUR-R05 loss requests; DPO-R24b complete break integration remains unresolved. Sibling citations add no acceptance obligation.

## 10. Related specs

- [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md): upstream hit/avoidance result.
- [combat-hit-recovery](combat-hit-recovery.md): HRC-R06 get-hit veto can confound defender seed deltas.
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): mutation/owner effects referred through RLI-Q03.

## Appendix A. Provenance

Evidence: original Lord of Destruction Expansion 1.13c, Windows x86. The [source manifest](../research/source-manifest.md) identifies D2Game/D2Common artifacts. [Companion](../research/spec-evidence/item-durability-loss.md) preserves 2026-10-12 static readings/partial observation, compatible dated sibling evidence, and 2026-10-15 targeted instruction/binary-table readings and read-only text-cell corroboration. Current image hashes match the manifest; loaded DLL/data identities for the observation have explicit limits. No rule is runtime-confirmed.
