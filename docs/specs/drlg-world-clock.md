# World clock

**Description**

Each act has its own outdoor world clock that runs through a repeating day and night cycle; Act III nights and the whole of Act IV run faster, and an eclipse switches the clock to a faster cycle of its own. The server advances every act's clock each frame and sends a fresh reading to players in that act when its copy has drifted or its time of day has moved on; a joining player receives the current reading for the act they are in. The client keeps its own copy, advances it, and accepts the server's readings.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: WCK-R, WCK-Q, WCK-V, WCK-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/drlg-world-clock.md)
Open evidence gaps: [gap file](../plans/gaps/drlg-world-clock.md)

## 1. Scope

This contract covers each act's world clock: its fixed clock states, the allocation, eclipse-start and reset assignments, normal/eclipse tick arithmetic, the client source wrapper, the two supplied intensity branches, server frame updates and resynchronization decisions, the join snapshot and client clock assignment. The quest stages that decide when Act II's clock is started or reset, and the start/reset messages they request, are owned by [Tainted Sun quest](quest-tainted-sun.md). Player resolution, timer and message delivery and presentation are explicit integration inputs. Other clock consumers, general lighting, colour interpolation, audio playback, save encoding and other game builds are outside the reproduced boundary. Their unresolved connections are named below.

## 2. Trigger and ownership

D2Game owns five per-act server clocks. The normal server frame increments the game frame, then ticks acts 0 through 4 before the subsequent frame operations identified in the companion. The start and reset assignments are requested by the quest stages of [Tainted Sun quest](quest-tainted-sun.md) (TSW-R01a, TSW-R02b, TSW-R05a); the join snapshot is requested when a client joins.

D2Common computes clock arithmetic and assignment. D2Client calls a separate client update wrapper and consumes a World-clock message through its command dispatcher. Its outer scheduling remains WCK-Q02. Periodic resynchronization requires a resolved player and client state 4. Transmission and visual/audio consumers remain supplied integrations, WCK-Q05 and WCK-Q01. Logical message names impose no wire encoding.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| act clocks | game act collection | five optional acts, each with optional environment | An act differs from its environment; indices 0..4 denote Acts I..V |
| clock | environment | cycle/period/ticks/rate/index/flag/last resync: i32 | cycle 0..5; flag 0 normal, any nonzero flag eclipse |
| intensity/colour | environment presentation | intensity i32, three colour bytes | Only WCK-R05a's supplied intensity branches and explicit override are reproduced |
| World-clock input | logical message | cycle/ticks i32, flag u8 | Client assignment values; binary encoding is evidence only |
| recipients | ordered client list | client; act u8; client state i32; direct-player flag bit; cached player reference | Distinguishes a send request from completed delivery; lookup result is supplied |
| client handler eligibility | World-clock handler's current-player/current-act correspondence | boolean | Supplied to WCK-R08b; its evaluation remains WCK-Q05 |
| client source | supplied source, 0 when absent | i32 | Used by client act mapping and presentation; producer/level identity remains WCK-Q02 |
| game frame | server frame counter | u32, low32 increment | Advanced before act updates |
| dependency operations | notifications, presentation, lookup, send | ordered requests; supplied normal completion/results | Hidden effects remain WCK-Q05 |

These constants originate in **engine code** and are **fixed by original code**. They are not archive-table cells or required storage layouts. Row indices are 0..5; colour interpolation is outside this contract.

| Table | Begin values | Period values |
|---|---|---|
| normal | 320, 340, 0, 160, 180, 200 | 3, 3, 0, 1, 1, 2 |
| Act IV comparison | 340, 350, 0, 180, 190, 200 | 3, 3, 0, 1, 1, 2 |
| eclipse | 300, 0, 60, 120, 180, 240 | 3, 0, 1, 2, 2, 2 |

Rate indices 0/1/2 select 128/4/8. Client thresholds are 1,40,75,103,109,1024; entry 0 is not compared.

All clock additions, subtractions and products retain low32 and are interpreted as i32 for signed comparisons. Division truncates toward zero. Valid cycles are 0..5. Required act/environment and flag records exist unless a rule handles absence; assertion failures are not normal no-op results.

An intensity request carries a source and an act only. The intensity operation reads the environment's clock flag as stored at the moment of the request; "observes flag f" below names that stored value, not an argument. A colour request carries an act.

## 4. Rules (execution order)

### WCK-R01. Boundary assignments

**WCK-R01a. Allocate** [verified in 1.13c code (static)]. Given successful allocation, zero initialization supplies index/flag/last resync 0; select rate 128 and normal row 2, giving cycle 2, period 0, ticks 0. Request initial intensity, then colour, with source/act 0; their results are supplied rather than assumed zero.

**WCK-R01b. Start** [verified in 1.13c code (static)]. Set index 1, rate 4, cycle 0, ticks 0, period 3. Request intensity with source 0/derived act 0; it observes the flag stored at entry, whatever its value. Set flag 1; recompute from eclipse row 0, writing period 3 and ticks `i32(4*300)=1200`; request intensity (source 0, act 0) again, which observes flag 1, then colour (act 0). Final clock `(cycle,period,ticks,rate,index,flag,last resync)` is `(0,3,1200,4,1,1,retained)`. The initial zero is not final ticks. Presentation remains WCK-Q01/WCK-Q05.

**WCK-R01c. Reset** [verified in 1.13c code (static)]. Set index 0, rate 128, cycle 2, ticks 0, period 0. Request intensity with source/act 0, observing the flag stored at entry, then clear flag. Make no colour request. Final tuple `(2,0,0,128,0,0,retained last resync)`. Immediate intensity need not equal its eventual normal value.

### WCK-R02. Advance

**WCK-R02. Advance** [verified in 1.13c code (static)]. Add 1 first. Flag 0: act 3 adds 15; otherwise, if normal[cycle].period equals 2, add 1 more, and act 2 adds 8 more in that night branch. Totals are day 1, ordinary night 2, Act III night 10, Act IV 16. Nonzero flag always adds 1. Rate does not scale the delta. Then run WCK-R03.

### WCK-R03. Wrap and transition

**WCK-R03** [verified in 1.13c code (static)]. After advance, signed ticks >=`i32(rate*360)` writes ticks 0. Next=cycle+1, wrapping 6 to 0. Compare using normal[next] if flag 0; otherwise Act IV comparison[next] for act argument 3, eclipse[next] for other acts. Only signed ticks **>**`i32(rate*compare begin)` transitions: set cycle=next; select assignment row from normal for flag 0, eclipse otherwise; write its period and `i32(rate*begin)` ticks. At most one transition runs. Flag nonzero/act 3 compares one table and assigns another. Wrap alone retains cycle/period.

### WCK-R04. Client source wrapper

**WCK-R04. Supplied client wrapper** [verified in 1.13c code (static)]. Signed source <40 gives act 0; <75 act 1; <103 act 2; <109 act 3; <1024 act 4; otherwise act 0. Compare threshold entries 1..5 only. The clock ticked is the environment of the act the caller supplies; the derived act is used only as the core's act argument and in the presentation requests. Save old period; call the core with this act; request intensity with source and this derived act, then colour with the same derived act. Source 120 overwrites colour bytes `(245,240,255)` afterward. Return 1 if period changed, else 0, even if cycle changed. Source production remains WCK-Q02; supplied mapping is available independently.

### WCK-R05. Intensity and appearance

**WCK-R05a. Supplied intensity branches** [verified in 1.13c code (static)]. For act other than 3 and stored flag nonzero, subtract 8 if signed intensity>32, then floor at 32. Ignore period. For act other than 3, flag 0, source 120, assign 200 directly. Other formulas/rendering are WCK-Q01. Each call updates the stored intensity field from its current value, so start's two calls compound.

**WCK-R05b. Appearance** [unresolved]. Darkness/colour/audio/timing require WCK-Q01/WCK-Q05; a clock tuple does not establish appearance.

### WCK-R06. Server execution

**WCK-R06a. Server frame** [verified in 1.13c code (static)]. Each normal frame increments the game frame, then visits acts 0..4 in order. Skip absent acts; run WCK-R06b for existing acts with the **act index**, independently of client source. Continue after no-notification results. The client threshold wrapper is not used here.

**WCK-R06b. Server resync** [verified in 1.13c code (static)]. An existing act whose environment is absent returns 0. Otherwise save cycle/period, run core with server act index. Rate nonzero: signed `q=ticks/rate`; `d=i32(q-last resync)`; if d<0 set `d=i32(-d)`. Signed d>16 stores last resync=q and returns 1. Otherwise return 1 only if cycle or period changed, else 0. Rate 0 skips the division and drift test and goes directly to the cycle/period test; negating i32 minimum leaves it negative.

Return 1 requests a current-clock read(act index), which yields that act's post-update cycle/ticks/flag for the World-clock payload, then traverses supplied clients in list order, completing each client's resolution and requests before the next client. Resolve each client's player: with the client's direct-player flag (bit 0) set, use its cached player reference as is. Otherwise a zero cached reference skips the client without a lookup; a nonzero one requests the supplied player lookup(client), and a failed lookup writes the cached reference to zero. Skip any client whose resolved player is absent. For each resolved player request player-notification(game,player,0), then change-tracking(player,1), then request World-clock only for client state 4 and client act equal to the server act index. Thus other state/act clients can receive pre-filter requests but no clock message. Delivery/helper effects remain WCK-Q05. Cycle/period notification alone retains last resync. Drift testing includes wrap to 0.

### WCK-R07. Join snapshot

**WCK-R07. Rebuild/snapshot** [verified in 1.13c code (static)]. Fresh clocks begin at WCK-R01a, advance on server frames and can be changed by TSW-R01d. Snapshot selects the joining client's current act and requests act-assignment first, World-clock with that act's current `(cycle,ticks,flag)` second. Values are sampled then; relative order to quest join needs supplied scheduling. This stage makes no environment save write.

### WCK-R08. Client execution

**WCK-R08a. Client cadence** [asm-only]. The identified client update calls WCK-R04 once, then requests refresh on nonzero return. Complete outer scheduling/source identity remain WCK-Q02. Idle progress does not prove every outer branch.

**WCK-R08b. Client assignment** [verified in 1.13c code (static)]. The supplied handler eligibility decides first: false makes no assignment and no request. When eligible, the client's current-act environment is the target and is required, and message cycle outside 0..5 or negative ticks takes the fatal assertion path before any clock write or request. Retain rate/index/last resync. If message ticks >`i32(rate*360)`, first use 0; equality is accepted. Assign cycle/ticks and period from the table selected by **message** flag. Request intensity (client source, its derived act), observing the old stored flag, then store the message flag value unchanged (a nonzero value other than 1 is stored as received). Nonzero new flag recomputes period/ticks from its eclipse row using retained rate, then requests intensity again and colour, both with the client source's derived act (the WCK-R04 threshold mapping). Supplied packet ticks are superseded there. Source 120 applies the colour override afterward, whether or not a colour request was made. Later handler refresh/consumers remain WCK-Q05.

### WCK-R09. Other consumers

**WCK-R09** [unresolved]. Complete consumers, including a possible stamina relationship, remain WCK-Q04. Bounded getter reading establishes no absence or stamina formula.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| WCK-R01a | allocation | verified in 1.13c code (static) | WCK-Q05: allocation/presentation | WCK-V10 |
| WCK-R01b | start clock | verified in 1.13c code (static) | WCK-Q01/WCK-Q05: presentation | TSW-V01, TSW-V03, TSW-V04 |
| WCK-R01c | reset clock | verified in 1.13c code (static) | WCK-Q01: presentation | TSW-V02, TSW-V05 |
| WCK-R02 | advance | verified in 1.13c code (static) | none | WCK-V01, WCK-V02, WCK-V03, WCK-V09, WCK-V11 |
| WCK-R03 | wrap/transition | verified in 1.13c code (static) | none | WCK-V01, WCK-V03, WCK-V04, WCK-V05, WCK-V06, WCK-V07, WCK-V08, WCK-V13 |
| WCK-R04 | client source | verified in 1.13c code (static) | WCK-Q02: source production; WCK-Q01/WCK-Q05: presentation | WCK-V14 |
| WCK-R05a | intensity branches | verified in 1.13c code (static) | WCK-Q01: other formulas/rendering | WCK-V15 |
| WCK-R05b | appearance | unresolved | WCK-Q01/WCK-Q05: visual/audio integration | none |
| WCK-R06a | server frame | verified in 1.13c code (static) | WCK-Q05: dependency effects | WCK-V11, WCK-V12 |
| WCK-R06b | server resync | verified in 1.13c code (static) | WCK-Q05: resolution/notifications/delivery | WCK-V11, WCK-V12, WCK-V13 |
| WCK-R07 | join snapshot | verified in 1.13c code (static) | WCK-Q05: outer order/delivery | WCK-V10 |
| WCK-R08a | client cadence | asm-only | WCK-Q02: outer scheduling | none |
| WCK-R08b | client assignment | verified in 1.13c code (static) | WCK-Q05: refresh/delivery | WCK-V16 |
| WCK-R09 | other consumers | unresolved | WCK-Q04: discovery | none |

## 5. Outputs and side effects

| Effect | Rule |
|---|---|
| client-wrapper period-change return 0/1 | WCK-R04 |
| clock rate/index/cycle/ticks/period/flag | WCK-R01a through WCK-R01c, WCK-R02, WCK-R03, WCK-R08b |
| client-wrapper derived act, intensity, presentation requests, colour override | WCK-R01a through WCK-R01c, WCK-R04, WCK-R05a, WCK-R08b |
| game frame, last resync, notification result | WCK-R06a, WCK-R06b |
| act-assignment and World-clock requests | WCK-R07, WCK-R06b |
| fatal assertion path for an invalid client message cycle/ticks | WCK-R08b |
| current-clock read, pre-filter player-notification and change-tracking requests, player lookups, cached-reference clear | WCK-R06b |

The join snapshot makes no environment save write; wider persistence remains gated under TSW-Q04. Message payloads belong to the rules; this table names effects only.

## 6. Branches and edge cases

Server act index and client derived source can differ. Clock flags are tested as nonzero. Negative source maps to act 0. Overflowed negative ticks use signed comparisons without saturation. Core wrap uses >=, transition uses >, client pre-clamp uses >. A cycle can change without a period change or client-wrapper return 1.

Missing environment during server update returns 0. Invalid required assignment objects take assertion paths outside normal vectors. A later failed delivery does not erase earlier local writes. Classic mode and other builds have no established equivalence here.

## 7. Open questions

The [gap file](../plans/gaps/drlg-world-clock.md#open-evidence-gaps) distinguishes behavioral, integration, observation and outside-boundary questions.

- **WCK-Q01** [unresolved]: complete presentation and observe darkness/colour/timing. Only two supplied intensity branches are reproduced; general lighting is outside scope.
- **WCK-Q02** [unresolved]: establish client source identity and full outer cadence. Supplied threshold arithmetic/server act-index behavior do not depend on this answer.
- **WCK-Q03** [unresolved]: observe server notification/send decisions, transition/wrap, the client's handling of a received World-clock message and client/server divergence. Idle reads do not reproduce exact resync vectors.
- **WCK-Q04** [unresolved]: complete other-consumer discovery, including any stamina connection.
- **WCK-Q05** [unresolved]: independently establish allocation, player lookup/notifications, presentation, the client handler's eligibility evaluation, outer join ordering and the transmission of clock messages. Local requests/results are explicit; complete integrated effects remain gated.

## 8. Test vectors

### 8.1 Executable vectors

Each case is fresh unless stated. Default clock tuple `(cycle,period,ticks,rate,index,flag,last resync)` is `(2,0,0,128,0,0,0)`. Default quest `(active,state,stage,transition,scheduled,applied,pending)` is `(1,0,0,7,0,0,0)`; markers A/B/C, mode argument, unit marker, source result, derived result, amulet count and roster word are 0, the diagnostic switch is 0, event auxiliary identity is 0, game quest counter is 1000, and callback/event unit U has identity 77. Act II/environment and valid player P in Act II exist (P has a client whose act is 1 and state is 4); flags are clear, other acts absent unless a case names an act or clock for them (such a case supplies that act and environment), no intervening frame, empty log. A "resolved" client has its direct-player flag set and a cached reference to its player, so it makes no lookup request; an "unresolved" client has the flag clear and a zero cached reference, so it makes no lookup request and no other request. Altar logs list flag and item tests as entries in the request log. A request "over supplied players" (pre-start, progress writer, start/reset senders) is logged as one entry per iterated player, in iteration order; a log that begins "after presentation" omits the earlier entries. In server start, reset and allocation logs, an unqualified "intensity" entry means intensity(source 0, act 0) and "colour" means colour(act 0), per WCK-R01a through WCK-R01c. When a server-resync case lists clients, the client list consists of exactly those clients, in that order; P's default client is not in it. Core vectors run advance then transition with their supplied act. Stage vectors log dependency requests in order; supplied callbacks/presentation/sends return normally without unlisted state or seed changes. A call-time snapshot below records the semantic state observed immediately on entering that supplied operation; it adds no function argument and no new stored field. Flag writes persist within a case. Presentation leaves clock fields unchanged. These supplied effects do not verify integration.

Except explicit draw cases, Seeds: no draw; Draws: none within the local slice. Hidden helper draws are excluded. All complete vectors remain runtime_observed: no; partial idle readings are limited to the assertions in the companion.

- **WCK-V01** Act IV advance (WCK-R02; boundary: core tick).
  - Inputs: act 3, rate 128, ticks 100, cycle 2, period 0, flag 0.
  - Seeds: no draw. Draws: none.
  - Trace: 100+1+15=116, below next threshold 20480.
  - Outputs: cycle/period/ticks 2/0/116; other fields retained.
  - (b) Fresh act 3, flag 0, rate 128, cycle 2, period 0, ticks 20470. Trace: 20470+16=20486 > normal row 3 threshold 128*160=20480; assign normal row 3. Outputs cycle 3, period 1, ticks 20480.
  - Discriminates: +1/+2 fails; (b) rejects comparing a flag-0 Act IV clock with the Act IV comparison table (threshold 23040 would leave cycle 2/ticks 20486).
  - runtime_observed: no.

- **WCK-V02** Act III night (WCK-R02; boundary: core tick).
  - Inputs: act 2, rate 128, cycle 5, period 2, ticks 30000, flag 0.
  - Seeds: no draw. Draws: none.
  - Trace: 30000+1+1+8=30010, below 40960.
  - Outputs: cycle/period/ticks 5/2/30010.
  - (b) Fresh act 2, flag 0, rate 128, cycle 5, period field 0 (supplied inconsistent with normal row 5), ticks 30000. Outputs cycle/period/ticks 5/0/30010.
  - Discriminates: +11 gives 30011; (b) rejects reading the stored period field instead of normal[cycle].period (that gives 30001).
  - runtime_observed: no.

- **WCK-V03** Eclipse rate-independent advance (WCK-R03, WCK-R02; boundary: core tick).
  - Inputs: act 1, cycle 2, period 1, ticks 300, rate 4, index 1, flag 1.
  - Seeds: no draw. Draws: none.
  - Trace: 301 does not exceed 4*120=480.
  - Outputs: cycle/period/ticks 2/1/301, rate 4 retained.
  - Discriminates: scaling delta by rate gives 304.
  - runtime_observed: no.

- **WCK-V04** Strict transition exceeded (WCK-R03; boundary: core tick).
  - Inputs: act 0, cycle 3, period 1, ticks 23040, rate 128, flag 0.
  - Seeds: no draw. Draws: none.
  - Trace: 23041>128*180=23040; assign normal row 4.
  - Outputs: cycle 4, period 1, ticks 23040.
  - Discriminates: keeping advanced tick 23041 fails.
  - runtime_observed: no.

- **WCK-V05** Strict equality (WCK-R03; boundary: core tick).
  - Inputs: WCK-V04 except ticks 23039.
  - Seeds: no draw. Draws: none.
  - Trace: 23040 equals threshold, no transition.
  - Outputs: cycle 3, period 1, ticks 23040.
  - Discriminates: >= changes cycle to 4.
  - runtime_observed: no.

- **WCK-V06** Wrap without cycle change (WCK-R03; boundary: core tick).
  - Inputs: act 0, cycle 1, period 3, ticks 46079, rate 128, flag 0.
  - Seeds: no draw. Draws: none.
  - Trace: 46080>=128*360; wrap 0; next threshold 0 is not exceeded.
  - Outputs: cycle 1, period 3, ticks 0.
  - (b) Fresh act 0, flag 0, rate 128, cycle 5, period 2, ticks 40960. Trace: night advance 40962; below wrap 46080; next wraps 6 to 0, normal row 0 threshold 128*320=40960 exceeded. Outputs cycle 0, period 3, ticks 40960.
  - Discriminates: transition-before-wrap/resetting cycle fails; (b) rejects mishandling next cycle 6.
  - runtime_observed: no.

- **WCK-V07** Supplied zero and returned start state (WCK-R03, WCK-R01b; boundary: fresh core ticks).
  - Inputs: act 1, rate 4, index 1, flag 1, cycle 0, period 3.
  - Seeds: no draw. Draws: none.
  - Trace: advance then compare eclipse row 1 begin 0.
  - (a) Supplied ticks 0: Outputs advanced 1>0, cycle 1, period 0, ticks 0.
  - (b) Returned start ticks 1200: Outputs advanced 1201<1440 but >0, cycle 1, period 0, ticks 0.
  - (c) Fresh cycle 5, period 2, ticks 1439 (other inputs as above): Trace: advance 1440 >= wrap threshold 4*360=1440, ticks 0; next 6 wraps to 0, eclipse row 0 threshold 4*300=1200 not exceeded. Outputs cycle 5, period 2, ticks 0.
  - Discriminates: (a) is a supplied fixture; it is not the returned start ticks value. (c) rejects a fixed wrap threshold 46080 (that keeps 1440, then transitions to cycle 0/period 3/ticks 1200).
  - runtime_observed: no.

- **WCK-V08** Act IV compare/assign (WCK-R03; boundary: core tick).
  - Inputs: act 3, flag 1, rate 4, cycle 2, period 1, ticks 720.
  - Seeds: no draw. Draws: none.
  - Trace: 721>Act IV row 3 threshold 4*180=720; assign eclipse row 3 begin 120/period 2.
  - Outputs: cycle 3, period 2, ticks 480.
  - (b) Fresh act 3, flag 1, rate 4, index 1, cycle 2, period 1, ticks 650. Trace: 651 does not exceed Act IV row 3 threshold 720. Outputs cycle 2, period 1, ticks 651.
  - (c) Fresh act 1, stored flag 2, rate 4, index 1, cycle 2, period 1, ticks 480. Trace: eclipse delta +1 gives 481 > eclipse row 3 threshold 4*120=480; assign eclipse row 3. Outputs cycle 3, period 2, ticks 480.
  - Discriminates: (c) rejects testing flag equal to 1 (normal handling keeps cycle 2/ticks 481 under threshold 640) and comparing a nonzero-flag clock with the normal table; assigning comparison table gives ticks 720/period 1; (b) rejects comparing with the eclipse (480) or normal (640) table and adding the act-3 +15 under flag 1 (666).
  - runtime_observed: no.
- **WCK-V09** Advance branches (WCK-R02; boundary: fresh core ticks).
  - Inputs: normal rate 128; fresh cases below; next threshold never reached.
  - Seeds: no draw. Draws: none.
  - Trace: one advance using supplied act and normal row period.
  - (a) cycle 5, period 2, ticks 30000, flag 0, acts 0/1/2/3/4: Outputs ticks 30002/30002/30010/30016/30002 respectively; cycle/period 5/2 retained.
  - (b) cycle 4, period 1, ticks 25000, flag 0, acts 2/3: Outputs ticks 25001/25016; cycle/period 4/1 retained.
  - (c) act 1, cycle 2, period 1, ticks 300, rate 4, index 1, flag 1: Outputs ticks 301, cycle/period 2/1 retained.
  - (d) cycle 5, period 2, ticks 300, rate 4, index 1, flag 1, acts 1/2/3: Outputs ticks 301 for each; cycle/period 5/2 retained (next thresholds 1200, or 1360 for act 3, not exceeded).
  - (e) act 0, flag 0, rate 128, cycle 0, period 3, ticks 1000: Outputs ticks 1001; cycle/period 0/3 retained (row 1 threshold 43520 not exceeded).
  - Discriminates: normal row chooses night; Act IV overrides it; (e) rejects treating period 3 (or any period >=2) as night, which gives 1002; (d) rejects applying night, Act III or Act IV bonuses to a nonzero flag.
  - runtime_observed: no.

- **WCK-V10** Allocation/own-act snapshot (TSW-R01d, WCK-R01a, WCK-R07; boundary: supplied stages).
  - Inputs: allocate Act II, quest join, snapshot; no frame; for (c) Act I also exists with supplied clock `(3,1,20000,128,0,0,0)`, not reallocated.
  - Seeds: no draw. Draws: none locally.
  - Trace: allocation gives `(2,0,0,128,0,0,0)` before supplied presentation; join selects progress; snapshot selects own act.
  - (a) P act 1, flags {2}: Outputs Act II `(0,3,1200,4,1,1,0)`, applied 1; snapshot act-assignment(P,1), World-clock(P,0,1200,1), after TSW-V01 start messages.
  - (b) P act 1, flags {2,1}: Outputs Act II allocation tuple, shared-record bit 13 set, P's flags still {1,2}, applied 0; act-assignment(P,1), World-clock(P,2,0,0), no start message.
  - (c) P act 0, flags {2}: Outputs Act II started tuple, Act I unchanged; no start messages for P; act-assignment(P,0), World-clock(P,3,20000,0).
  - Discriminates: (c) rejects always snapshotting Act II.
  - runtime_observed: no.

- **WCK-V11** Exact 1000-frame static scenario (WCK-R02, WCK-R06a, WCK-R06b; boundary: supplied server frames/client core calls).
  - Inputs: Act I default server/client clocks; other acts absent, including Act II; 1000 normal server frames and 1000 client core calls with source/act 0; no external messages/start/reset; one resolved state-4 act-0 client.
  - Seeds: no draw. Draws: none in clock slices.
  - Trace: each adds 1; 1000<20480; q reaches 7, never distance >16 from 0.
  - Outputs: both cycle 2/period 0/ticks 1000; server last resync 0; game frame increases 1000; no World-clock/pre-filter request or cycle change.
  - Discriminates: rate-scaled delta/notification at distance <=16 fails. Idle measurements are partial corroboration and do not reproduce this exact case.
  - runtime_observed: no.

- **WCK-V12** Drift/recipient filters (WCK-R06a, WCK-R06b; boundary: supplied server act update).
  - Inputs: act 0 default clock except ticks 2175; clients in order A resolved/state 4/act 0, B resolved/state 4/act 1, C resolved/state 3/act 0, D unresolved/state 4/act 0.
  - Seeds: no draw. Draws: none in clock/request slice.
  - Trace: ticks 2176, q=17, distance 17>16.
  - Outputs: cycle 2/period 0/ticks 2176, last resync 17, result 1. Exact log: current-clock read(act 0), then player-notification(game,A-player,0), change-tracking(A-player,1), World-clock(A,2,2176,0), player-notification(game,B-player,0), change-tracking(B-player,1), player-notification(game,C-player,0), change-tracking(C-player,1). No message for B/C/D and no operations for D. This forbids deferring A's clock request until after B/C notification pairs.
  - Fresh Inputs: ticks 2047, otherwise same. Outputs ticks 2048/q16/result 0/last 0, no log.
  - Separate Inputs: one normal frame from game frame 9; only acts 0 and 3 present, each with cycle 2/period 0/ticks 2175/rate 128/index 0/flag 0/last resync 0. Clients A resolved/state 4/act 0 and E resolved/state 4/act 3, in that order; any client source values are irrelevant. Trace: frame becomes 10; act 0 advances to 2176/q17 and act 3 to 2191/q17, each notifying. Outputs: both last-resync fields 17, both cycle/period 2/0; final game frame 10. Exact log: current-clock read(act 0), A notification, A tracking, World-clock(A,2,2176,0), E notification, E tracking; then current-clock read(act 3), A notification, A tracking, E notification, E tracking, World-clock(E,2,2191,0). Here notification/tracking mean the same two operations and arguments as above for that recipient. No other acts are updated.
  - Resolution Inputs: ticks 2175 as first case; clients in order F (direct flag clear, cached reference nonzero, supplied lookup returns F-player), G (direct flag clear, cached nonzero, lookup fails), H (direct flag clear, cached zero), I (direct flag set, cached reference I-player); all state 4/act 0. Outputs: G's cached reference becomes 0; F's, H's and I's are unchanged. Single ordered log after the current-clock read: lookup(F), F notification, F tracking, World-clock(F,2,2176,0), lookup(G), I notification, I tracking, World-clock(I,2,2176,0). No lookup for H or I.
  - Discriminates: before-tick sampling, >=16, filtering before pre-filter requests, batching A's send after B/C, reversed act visitation, or using a client-derived act for the server delta fails. The resolution case rejects looking up a zero cache, retaining a failed cache, looking up a direct-flag client, or resolving all clients before any notification.
  - runtime_observed: no.

- **WCK-V13** Signed overflow/wrap resync (WCK-R03, WCK-R06b; boundary: fresh core/server updates).
  - Inputs: rate 128/index 0/flag 0/act 0; overrides below.
  - Seeds: no draw. Draws: none.
  - Trace: low32 arithmetic precedes signed comparisons.
  - (a) cycle 2/period 0/ticks 2147483647: Outputs ticks -2147483648; no wrap/transition; cycle/period 2/0.
  - (b) Server cycle 1/period 3/ticks 46079/last 350; one resolved state-4 act-0 client: Outputs wrap ticks 0, cycle/period 1/3, q0, difference -350 becomes 350, last 0/result 1 and, after current-clock read(act 0) and the client's two pre-filter requests, World-clock(client,1,0,0).
  - (c) Server default except ticks 0/last -2147483648: Outputs ticks 1/q0; subtraction and negation leave d=-2147483648; result 0/last retained, no requests.
  - (d) Server synthetic rate 0/cycle 2/period 0/ticks -2/last resync 17: Trace: advance -1, signed -1 is below wrap threshold 0 and does not exceed transition threshold 0, then skip division. Outputs cycle 2/period 0/ticks -1, last resync 17, result 0, no requests. This supplied case does not claim rate 0 is selected by the fixed rate table.
  - (e) Server cycle 3/period 1/ticks 23040/last resync 170; sole resolved state-4 act-0 client P: Trace: advance 23041, transition to cycle 4/period 1/ticks 23040, q180/distance 10; cycle alone changes. Outputs last resync 170 retained (not 180), result 1; log current-clock read(act 0), player-notification(game,P-player,0), change-tracking(P-player,1), World-clock(P,4,23040,0).
  - (f) One normal server frame from game frame 9, only act 0 present but its environment absent, sole resolved state-4 act-0 client: Outputs game frame 10, server update return 0, environment still absent, no current-clock read, pre-filter notification or World-clock request.
  - (g) Server cycle 2/period 0/ticks -130/rate 128/last resync 15, sole resolved state-4 act-0 client: Trace: advance -129; no wrap/transition; q=-129/128 truncates to -1; d=-16 becomes 16, not >16. Outputs ticks -129, last resync 15 retained, result 0, no requests.
  - Discriminates: (g) rejects floor division (q=-2, distance 17, notify); unbounded absolute value notifies in (c); saturation fails (a); unconditional division, skipping the entire update (ticks -2), or forcing a zero-rate wrap (ticks 0) fails (d); checking only drift or period, or storing q on a cycle-only notification, fails (e); dereferencing an absent environment fails (f).
  - runtime_observed: no.

- **WCK-V14** Client mapping/period return (WCK-R04; boundary: supplied-source wrapper).
  - Inputs: for each source -1,39,40,74,75,102,103,108,109,120,1023,1024, a fresh case in which the caller supplies Act I with the default clock (that environment is the one ticked); presentation supplied, and the supplied colour request writes colour bytes (1,2,3).
  - Seeds: no draw. Draws: none.
  - Trace: signed entries 1..5 mapping then core tick.
  - Outputs: acts 0,0,1,1,2,2,3,3,4,4,4,0; ticks 1 except sources 103/108 give 16; cycle 2/period 0/return 0 throughout. Source 120 final colour (245,240,255).
  - Ordered presentation requests for every listed source case: first intensity with that exact supplied source and its derived act from the output array, then colour with the same derived act. At entry to both requests the clock snapshot is cycle 2/period 0/flag 0/rate 128/index 0 and ticks 1, except sources 103/108 whose ticks are 16. Thus source 120 requests intensity(source 120, act 4) before colour(act 4), and source 103 requests intensity(source 103, act 3) before colour(act 3), with ticks 1 at both entries; only after colour returns are the three colour bytes overwritten to (245,240,255). For every other listed source, including act-4 sources 109 and 1023, the colour bytes after return are (1,2,3), those left by the supplied colour request; no override is applied. Source 120's final (245,240,255) after a colour request that wrote (1,2,3) rejects applying the override before the colour request. The requests carry these established semantic inputs; the snapshots are observations, not extra arguments.
  - Fresh Inputs: source 0/cycle 0/period 3/ticks 43520/rate 128/flag 0. Trace: 43521>row 1 threshold 43520. Outputs cycle 1/period 3/ticks 43520/return 0. Ordered requests: intensity(source 0, act 0), colour(act 0); both entry snapshots are cycle 1/period 3/ticks 43520/flag 0/rate 128/index 0.
  - Separate Inputs: source 0/cycle 0/period 3/ticks 1200/rate 4/index 1/flag 1. Trace: advance 1201, transition to eclipse row 1, period changes from 3 to 0. Outputs cycle 1/period 0/ticks 0/return 1; ordered intensity(source 0, act 0), colour(act 0), both with entry snapshots cycle 1/period 0/ticks 0/flag 1/rate 4/index 1.
  - Discriminates: environment-derived act/cycle-change return, always returning 0, swapped source/act request inputs, omitted requests or colour-before-intensity fails. An intensity request carrying a fixed act 0 instead of the derived act fails at sources 40..1023. Applying the colour override to every act-4 source fails at 109/1023. Negative source maps to act 0 under either signed or unsigned comparison, so this vector does not distinguish them; the results are identical.
  - runtime_observed: no.

- **WCK-V15** Intensity calls (WCK-R05a; boundary: isolated supplied intensity branch).
  - Inputs: act 1/flag 1, fresh intensity 200/39/33/32/20; separate act 1/flag 0/source 120/intensity 20.
  - Seeds: no draw. Draws: none.
  - Trace: subtract 8 only above 32, then floor 32; normal source 120 assigns 200.
  - Outputs: 192/32/32/32/32; separate case 200. Separate act 1/flag 1/intensity -2147483648: Outputs 32 (no subtraction because it is not above 32; an unconditional low32 subtract would give 2147483640). Separate act 1/flag 1/source 120/intensity 200: Outputs 192 (the flag-nonzero branch precedes the source-120 assignment). Separate act 1/stored flag 2/intensity 200: Outputs 192.
  - Discriminates: floor-before-subtract alone yields 31/25 at 39/33; period does not affect these branches.
  - runtime_observed: no.

- **WCK-V16** Client message/rate (WCK-R08b; boundary: supplied eligible handler/clock stage).
  - Inputs: default client clock/source 0, eligible player/act, World-clock(5,0,1); presentation supplied.
  - Seeds: no draw. Draws: none.
  - Trace: initially assign eclipse row 5 period 2/ticks 0; first intensity uses old flag 0; write flag 1; recompute 128*240=30720.
  - Outputs: clock `(5,2,30720,128,0,1,0)`; first intensity [call-time snapshot flag 0/cycle 5/period 2/ticks 0/rate 128/index 0], second intensity [snapshot flag 1/cycle 5/period 2/ticks 30720/rate 128/index 0], then colour; later refresh supplied.
  - Fresh Inputs: default clock, message (2,46080,0). Outputs cycle 2/period 0/ticks 46080/rate 128/flag 0; one intensity request, no recompute/colour. Message (2,46081,0) instead yields ticks 0. An ineligible player/act instead yields unchanged clock, no assignment requests. With eligibility true, message (6,0,0), (-1,0,0) or (2,-1,0) takes the fatal assertion path before any clock write or request. With eligibility false, message (6,0,0) yields an unchanged clock, no request and no assertion. Source 120 with default clock and message (2,46080,0): Outputs cycle 2/period 0/ticks 46080/flag 0, one intensity request (source 120, act 4), no colour request, and final colour bytes (245,240,255).
  - Fresh Inputs: client clock `(5,2,30720,128,0,1,0)`, source 0, message World-clock(2,100,0). Trace: assign cycle 2, ticks 100, normal row 2 period 0; first intensity observes old flag 1; store flag 0; no recompute. Outputs clock `(2,0,100,128,0,0,0)`; one intensity(source 0, act 0) [snapshot flag 1/cycle 2/period 0/ticks 100], no colour request. Rejects recomputing whenever the stored flag is nonzero.
  - Fresh Inputs: default client clock, source 0, message World-clock(1,0,1). Trace: assign eclipse row 1 period 0 (normal row 1 would give 3), ticks 0; first intensity observes flag 0; write flag 1; recompute eclipse row 1, ticks 128*0=0. Outputs clock `(1,0,0,128,0,1,0)`; first intensity [snapshot flag 0/cycle 1/period 0/ticks 0], second intensity [snapshot flag 1/period 0/ticks 0], colour(act 0). Rejects choosing the period table by the old stored flag.
  - Fresh Inputs: client clock `(2,0,0,4,1,0,17)`, source 0, message World-clock(5,0,1). Trace: 0 does not exceed 4*360; assign eclipse row 5 period 2/ticks 0; first intensity observes flag 0; write flag 1; recompute 4*240=960. Outputs clock `(5,2,960,4,1,1,17)`; ordered intensity(source 0, act 0) [snapshot flag 0/cycle 5/period 2/ticks 0/rate 4/index 1], intensity(source 0, act 0) [snapshot flag 1/ticks 960], colour(act 0).
  - Fresh Inputs: client clock `(5,2,30720,128,0,1,0)`, source 0, message World-clock(5,999,1). Trace: 999 does not exceed 46080; assign cycle 5, ticks 999, eclipse row 5 period 2; first intensity observes old flag 1; store flag 1; recompute eclipse row 5, ticks 30720. Outputs clock `(5,2,30720,128,0,1,0)`; ordered intensity(source 0, act 0) [snapshot flag 1/cycle 5/period 2/ticks 999], intensity(source 0, act 0) [snapshot flag 1/ticks 30720], colour(act 0).
  - Fresh Inputs: default client clock, source 0, message World-clock(1,0,2). Trace: nonzero message flag selects eclipse row 1 period 0, ticks 0; first intensity observes flag 0; store flag 2 as received; recompute eclipse row 1, ticks 0. Outputs clock `(1,0,0,128,0,2,0)`; first intensity [snapshot flag 0/cycle 1/period 0/ticks 0], second intensity [snapshot flag 2/period 0/ticks 0], colour(act 0).
  - Fresh Inputs: default client clock, source 75, message World-clock(5,0,1). Trace: source 75 maps to act 2; assign eclipse row 5 period 2/ticks 0; recompute 128*240=30720. Outputs clock `(5,2,30720,128,0,1,0)`; ordered intensity(source 75, act 2) [snapshot flag 0], intensity(source 75, act 2) [snapshot flag 1], colour(act 2). Rejects a fixed act 0 in these requests.
  - Fresh Inputs: client clock `(2,0,0,4,1,0,0)`, source 0, message World-clock(2,1441,0). Trace: 1441 > 4*360=1440, use 0; normal row 2 period 0. Outputs clock `(2,0,0,4,1,0,0)`; one intensity(source 0, act 0), no colour. Message World-clock(2,1440,0) instead: Outputs clock `(2,0,1440,4,1,0,0)`, one intensity, no colour.
  - Discriminates: the (5,999,1) case rejects recomputing only when the stored flag changes from 0; the (1,0,2) case rejects selecting the eclipse table or recomputing only for flag exactly 1, and normalizing the stored flag to 1; the rate-4 pre-clamp cases reject a fixed 46080 threshold; copying flag-1 packet ticks yields 0; copying server rate 4 yields 960 in the first case, while resetting a retained rate 4/index 1/last 17 to defaults fails the rate-4 case; >= pre-clamp fails at 46080; storing message flag before first intensity fails its old-flag snapshot.
  - runtime_observed: no.

### 8.2 Observation vectors

Pending [companion procedures](../research/spec-evidence/drlg-world-clock.md#observation-procedures) specify setup/capture limits.

| Id | Measure | Expected / comparison |
|---|---|---|
| **WCK-O01** | leave/return, server frames, same-act recipients | continued Act II advance, other-act preservation, separate drift/transition sends |

## 9. Acceptance and dependencies

Partial acceptance reproduces supplied-input assertions of WCK-V01 through WCK-V16 for all static leaves in the Id map. Arithmetic, writes, returns and request order are available independently of downstream delivery/presentation. WCK-V11 is an exact static scenario. The start and reset assignments are also exercised through the quest-stage vectors TSW-V01 through TSW-V05 of [Tainted Sun quest](quest-tainted-sun.md); WCK-R01c has no vector of its own here. Passing these does not establish complete packet, save or visual parity.

Full reproduction also requires gated assertions of WCK-R05b, WCK-R08a and WCK-R09; independently established integrations WCK-Q05; and applicable matching WCK-O01 and WCK-Q03 observations. Status/partial acceptance imply no user approval.

**Depends on:** the i32/low32 notation of [numeric/RNG foundations](core-numeric-and-rng-foundations.md). TSW-R01d in [Tainted Sun quest](quest-tainted-sun.md) as the supplied join stage in WCK-V10. Section 3's supplied integration operations remain WCK-Q05. GDA-R04a in [game-data fields](core-game-data-fields.md) supplies only the distinction between code-built records and loaded data. Owners of the cited GDA ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA).

**Depended on by:** [Tainted Sun quest](quest-tainted-sun.md) requests WCK-R01b and WCK-R01c. No known non-gap plan or executable consumer cites this contract. Other consumers need their own contracts, WCK-Q04.

## 10. Related specs

- [Tainted Sun quest](quest-tainted-sun.md): the quest stages that start and reset Act II's clock and send the start/reset messages.
- [Numeric/RNG foundations](core-numeric-and-rng-foundations.md): shared arithmetic and frame conventions.
- [Game-data fields](core-game-data-fields.md): names and code-built versus table-record distinction.

## Appendix A. Provenance

Evidence targets original Windows x86 Lord of Destruction Expansion 1.13c D2Common, D2Game and D2Client, identified in the [source manifest](../research/source-manifest.md). The [public companion](../research/spec-evidence/drlg-world-clock.md) records RVAs/ordinals, dated readings, correction rationale, partial observations and limits. Static readings dated 2026-10-12 retain their scope; targeted matched-reconstruction, instruction and original-image checks are dated 2026-10-15.

Read-only idle observations from 2026-10-12 corroborate code constants, selected fresh-clock fields and normal Act I day progress. They do not reproduce complete exact-frame, transition or notification vectors. All executable vectors remain runtime_observed: no. No new gameplay observation is recorded.
