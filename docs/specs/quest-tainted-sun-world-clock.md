# Tainted Sun world clock

**Description**

The Tainted Sun quest changes the outdoor world clock for Act II, and breaking the viper altar returns it to normal. The server keeps a separate clock for each act and tells players about changes in their current act. Quest progress can apply the eclipse again when a player joins a game.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: TSW-R, TSW-Q, TSW-V, TSW-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/quest-tainted-sun-world-clock.md)
Open evidence gaps: [gap file](../plans/gaps/quest-tainted-sun-world-clock.md)

## 1. Scope

This contract covers the quest-start, altar, retry and join stages that change Act II's clock; fixed clock states, normal/eclipse tick arithmetic, client clock assignment and server clock notification decisions; and their bounded RNG use. Quest flags, supplied callback results, timer delivery, player resolution and message delivery are explicit integration inputs. Complete quest progression, save encoding, other clock consumers, general lighting, colour interpolation, audio playback and other game builds are outside the reproduced boundary. Their unresolved connections are named below.

## 2. Trigger and ownership

D2Game owns the game-level quest and five per-act server clocks. Quest initialization installs join and event callbacks; join progress, an animation callback, supplied event keys 0x28/0x2C/0x2D and a timer callback reach the start stage. Event identities remain TSW-Q01. Object-operation dispatch reaches the altar stage. The normal server frame increments the game frame, then ticks acts 0 through 4 before the subsequent frame operations identified in the companion.

D2Common computes clock arithmetic and assignment. D2Client calls a separate client update wrapper and consumes a World-clock message through its command dispatcher. Its outer scheduling remains TSW-Q07. Start/reset recipients are filtered by act; periodic resynchronization also requires a resolved player and client state 4. Transmission and visual/audio consumers remain supplied integrations, TSW-Q10 and TSW-Q03. Logical message names impose no wire encoding.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| act clocks | game act collection | five optional acts, each with optional environment | An act differs from its environment; indices 0..4 denote Acts I..V |
| clock | environment | cycle/period/ticks/rate/index/flag/last resync: i32 | cycle 0..5; flag 0 normal, any nonzero flag eclipse |
| intensity/colour | environment presentation | intensity i32, three colour bytes | Only TSW-R07c's supplied intensity branches and explicit override are reproduced |
| quest | game quest 10, slot 11 | active/state/stage/transition markers u8; scheduled/applied/pending u8; neutral markers A/B/C u8; mode argument, unit marker, signed amulet count, source result and derived result i32; roster word u16 | Neutral marker writes are specified below; initialization zeroes the private markers |
| shared quest flags | game-level shared quest-flag record, distinct from any player's flags | slot-11 bit 13 | Written only by TSW-R01d; a player's own bit 13 is written only by TSW-R02b |
| quest 9 counter | game quest 9's private state, when present | i32, low32 addition | Written only by TSW-R02c |
| callback unit | animation/event callback input, optional | unit identity i32 | Absent unit records identity -1 |
| diagnostic switch | server-wide value, not per game | i32; nonzero enables | Producer remains TSW-Q10 |
| progress flags | the player's current-difficulty slots 11 and 10 | boolean bits 0,1,2,3,4,13,14 | Broader meanings remain TSW-Q02 |
| World-clock input | logical message | cycle/ticks i32, flag u8 | Client assignment values; binary encoding is evidence only |
| recipients | ordered player iteration and client list | player/client or absent; act u8; client state i32; client direct-player flag bit; cached player reference | Distinguishes a send request from completed delivery; lookup result is supplied |
| client handler eligibility | World-clock handler's current-player/current-act correspondence | boolean | Supplied to TSW-R10c; its evaluation remains TSW-Q10 |
| client source | supplied source, 0 when absent | i32 | Used by client act mapping and presentation; producer/level identity remains TSW-Q07 |
| event | supplied callback input | event key i32; auxiliary selector u16; auxiliary identity i32; optional unit | Gameplay meanings are not inferred from values |
| altar | player, optional object, item-presence results | object identity i32, object mode marker i32, object item-code field (four characters); flag/item tests boolean | Item search region is supplied, TSW-Q10 |
| game frame | server frame counter | u32, low32 increment | Advanced before act updates |
| timer | scheduler and delivery | schedule parameter, game quest counter i32; callback identity | Request and actual timed callback execution are distinct |
| dependency operations | iteration, notifications, mode/drop callbacks, presentation, send | ordered requests; supplied normal completion/results | Hidden effects/RNG remain TSW-Q08/TSW-Q10 |
| RNG | quest, game and unit seeds | each `{lo,hi}` u32 | Independent owners; NRF-R01/NRF-R02 own arithmetic |

These constants originate in **engine code** and are **fixed by original code**. They are not archive-table cells or required storage layouts. Row indices are 0..5; colour interpolation is outside this contract.

| Table | Begin values | Period values |
|---|---|---|
| normal | 320, 340, 0, 160, 180, 200 | 3, 3, 0, 1, 1, 2 |
| Act IV comparison | 340, 350, 0, 180, 190, 200 | 3, 3, 0, 1, 1, 2 |
| eclipse | 300, 0, 60, 120, 180, 240 | 3, 0, 1, 2, 2, 2 |

Rate indices 0/1/2 select 128/4/8. Client thresholds are 1,40,75,103,109,1024; entry 0 is not compared. Quest 10, progress slots 11/10, four-character item codes `vip ` and `hst `, and Quest-event flag values 16/1 are also fixed code selections. Item names are supporting descriptions, not new table mappings.

All clock additions, subtractions and products retain low32 and are interpreted as i32 for signed comparisons. Division truncates toward zero. Valid cycles are 0..5. Required act/environment and flag records exist unless a rule handles absence; assertion failures are not normal no-op results.

An intensity request carries a source and an act only. The intensity operation reads the environment's clock flag as stored at the moment of the request; "observes flag f" below names that stored value, not an argument. A colour request carries an act.

## 4. Rules (execution order)

### TSW-R01. Quest start stages

**TSW-R01a. Shared start** [verified in 1.13c code (static)]. Write transition marker 0 and stage marker 1; request the pre-start callback over supplied players. After normal completion, change state 0 to 1 and apply TSW-R07e with text `changing state to 1` plus a newline, retaining other states without that request. If Act II is absent, set pending 1 and return 0: retain earlier quest writes, leave clocks unchanged, request no start messages. Otherwise apply TSW-R04a, request TSW-R07a over players and return 1. This stage does not set applied or clear pending. Dependency-produced effects are TSW-Q10.

**TSW-R01b. Progress writer** [verified in 1.13c code (static)]. Per player, return 0 without writes if slot-11 bit 0 or 1 is set, or quest 10 is absent. Otherwise, unsigned state >=1 sets bit 2; additionally state 2 sets bit 3, or state 3 sets bit 4. Clear no other bits; return 0. Supplied flag operations persist in order.

**TSW-R01c. Upstream identity** [unresolved]. Supplied keys 0x2C/0x2D reach the schedule condition. A related callback compares an auxiliary selector with 0x15C and signed range 0x16A..0x174; one branch also tests identity 0xB1. Their meanings, including whether Lost City entry produces a relevant event, remain TSW-Q01. These comparisons establish no universal absence of a level-entry trigger.

**TSW-R01d. Join selection** [verified in 1.13c code (static)]. Test slot-11 bit 0, then bit 1. Either set: set slot-11 bit 13 in the game-level shared quest-flag record (the joining player's flags are unchanged) and return without applying. If bit 2 is clear, return without writes. Otherwise run TSW-R01a. Success sets applied 1; failure retains pending 1 and requests TSW-R07b. After either result overwrite `(stage,state)` with `(2,3)` if bit 4 is set, else `(2,2)` if bit 3 is set, else `(1,1)`. Failure does not undo this overwrite. Join selection requests no progress writer.

**TSW-R01e. Animation start slice** [verified in 1.13c code (static)]. First, on every call, write marker B=1 and unit marker=callback unit identity (-1 when absent). Active is tested once, at this point. If active is not exactly 1, skip to the final request. Otherwise, for unsigned state <=1, run TSW-R01a; set applied 1 on success or pending 1 on failure, then write state 3 and transition 0 in both cases. Then, still within active exactly 1 and for any state, if stage is 0, write transition 0 and stage 2 and request the pre-start callback over supplied players. Shared start writes stage 1, so this branch follows only a skipped start. Finally, on every path, request mode-update(callback unit, current mode argument value). The mode-update's effects are TSW-Q10.

**TSW-R01f. Event retry slice** [verified in 1.13c code (static)]. These tests run in order as independent checks, not an else-chain. Key 0x2C/0x2D, state 0, active exactly 1: invoke TSW-R01g. Event key 0x28 and pending exactly 1: run TSW-R01a; on success run TSW-R01b over supplied players then clear pending. Failure retains pending 1 and earlier writes. TSW-R01i then runs on every call. Supplied operation effects are TSW-Q10.

**TSW-R01g. Schedule** [verified in 1.13c code (static)]. Nonzero state or scheduled means no draw/request. Otherwise step the game quest structure's seed once through NRF-R01, select parameter `15+(new lo & 1)` (15 or 16), request scheduling TSW-R01h, then write scheduled 1. The scheduler stores a due value equal to the supplied game quest counter plus parameter with low32 arithmetic. Elapsed gameplay units/delivery remain TSW-Q10.

**TSW-R01h. Timer callback** [verified in 1.13c code (static)]. Clear scheduled first. If applied is nonzero, return 1 without applying. Otherwise run TSW-R01a. Failure sets pending 1 and returns 1; success sets applied 1, runs TSW-R01b over supplied players and returns 1. Success does not itself clear pending.

**TSW-R01i. Event auxiliary-identity slice** [verified in 1.13c code (static)]. Only when the event's auxiliary identity equals 0x28: if the roster word is nonzero, request roster-update(event unit identity, -1 when absent). Then, if state is exactly 2, write state 3 and apply TSW-R07e with text `changing state to 3` plus a newline. This test reads state after TSW-R01f's start stage. The event key does not gate this slice. Roster-update effects are TSW-Q10.

### TSW-R02. Altar stage

**TSW-R02a. Eligibility** [verified in 1.13c code (static)]. O denotes the operated object, which may be absent. A missing quest 10 returns 0 with no other work. With quest 10 found, test player slot-11 bit 1, then bit 0 if needed. Only if either is set, test slot-10 bit 0, then supplied `vip ` presence, then `hst ` presence, stopping at the first true result. True: request rejection with value 19 and player; return 0 without reset. If both primary bits are clear, skip all three secondary tests. Next, an existing object with nonzero mode marker returns 0 without reset; a missing object does not take that return. Otherwise TSW-R02b runs when active is nonzero, else TSW-R02c. All returns are 0.

**TSW-R02b. Active branch** [verified in 1.13c code (static)]. Requests and writes, in order:
1. Request object-source(O), store result as source result; request source-derived(source result), store result as derived result. Request mode-update(O,1). Write mode argument=2, marker B=1, unit marker=object identity (-1 when absent), marker A=1.
2. If active is still nonzero: run quest reset. With Act II present, clear applied, run TSW-R04b and request reset messages. With Act II absent, retain applied/clock and request no reset message. Then apply TSW-R07e with text `Finished quest`, and request quest-notification(4), then game-notification(6).
3. Write amulet count 0; request participant iteration(player). The supplied iteration sets the amulet count.
4. If active is still nonzero: if player slot-11 bit 0 is clear, set bit 1, set bit 13, then **clear** bit 14. Then request post-flag iterations A, B and C, each with O. Then request quest-record update(13).
5. Request source-derived(source result) again. Write O's item-code field to `vip `. Request amulet drop(O) amulet-count times, comparing the signed attempt index with the amulet count re-read before each attempt; a count <=0 requests none. Each nonzero drop result requests item-flag mark(item, 16, 1) and counts one success. Then request amulet-result(successes) and context request(4).
6. Draw once from the quest seed (TSW-R12b) and request `5+(new lo mod 5)` gold-pile drops(O). A missing game quest structure takes the fatal assertion path instead.
7. Request completion iteration, then a timer request with parameter 10 for TSW-R02d, whose due value is the supplied game quest counter plus 10 (low32), as in TSW-R01g. Return 0.

**TSW-R02c. Inactive branch** [verified in 1.13c code (static)]. Request mode-update(O,1); write marker A=1, mode argument=2 and amulet count 0; request participant iteration(player). Request object-source(O) and source-derived(result), storing neither. Write O's item-code field, run the counted drop loop and context request(4) as in TSW-R02b step 5, without amulet-result. Run TSW-R02b step 6. Then find quest 9: if present, add the success count to its counter (low32); return 0 either way. No environment reset, flag write, timer, marker B/unit marker write or completion iteration occurs.

**TSW-R02d. Altar timer callback** [verified in 1.13c code (static)]. If state is exactly 4, write transition 0 and stage 3 and request the pre-start callback over supplied players. Then, in every case, write marker C=0 and return 1. Timer delivery and elapsed units remain TSW-Q10.

Supplied operation effects, including actual drops, iteration results and amulet counts, remain TSW-Q08/TSW-Q10.

### TSW-R03. Ownership

**TSW-R03** [verified in 1.13c code (static)]. Quest state is game-level; each act owns its environment. Start/reset address only Act II and request messages only for players whose client act equals 1. Any eligible player's altar operation can reset that shared clock. Leaving an act does not reset it; the server continues updating existing acts. Join snapshots select the joiner's own act, TSW-R09a.

### TSW-R04. Boundary assignments

**TSW-R04a. Start** [verified in 1.13c code (static)]. Set index 1, rate 4, cycle 0, ticks 0, period 3. Request intensity with source 0/derived act 0; it observes the flag stored at entry, whatever its value. Set flag 1; recompute from eclipse row 0, writing period 3 and ticks `i32(4*300)=1200`; request intensity (source 0, act 0) again, which observes flag 1, then colour (act 0). Final clock `(cycle,period,ticks,rate,index,flag,last resync)` is `(0,3,1200,4,1,1,retained)`. The initial zero is not final ticks. Presentation remains TSW-Q03/TSW-Q10.

**TSW-R04b. Reset** [verified in 1.13c code (static)]. Set index 0, rate 128, cycle 2, ticks 0, period 0. Request intensity with source/act 0, observing the flag stored at entry, then clear flag. Make no colour request. Final tuple `(2,0,0,128,0,0,retained last resync)`. Immediate intensity need not equal its eventual normal value.

**TSW-R04c. Allocate** [verified in 1.13c code (static)]. Given successful allocation, zero initialization supplies index/flag/last resync 0; select rate 128 and normal row 2, giving cycle 2, period 0, ticks 0. Request initial intensity/colour with source/act 0; their results are supplied rather than assumed zero.

### TSW-R05. Wrap and transition

**TSW-R05** [verified in 1.13c code (static)]. After advance, signed ticks >=`i32(rate*360)` writes ticks 0. Next=cycle+1, wrapping 6 to 0. Compare using normal[next] if flag 0; otherwise Act IV comparison[next] for act argument 3, eclipse[next] for other acts. Only signed ticks **>**`i32(rate*compare begin)` transitions: set cycle=next; select assignment row from normal for flag 0, eclipse otherwise; write its period and `i32(rate*begin)` ticks. At most one transition runs. Flag nonzero/act 3 compares one table and assigns another. Wrap alone retains cycle/period.

### TSW-R06. Advance and client source

**TSW-R06a. Advance** [verified in 1.13c code (static)]. Add 1 first. Flag 0: act 3 adds 15; otherwise, if normal[cycle].period equals 2, add 1 more, and act 2 adds 8 more in that night branch. Totals are day 1, ordinary night 2, Act III night 10, Act IV 16. Nonzero flag always adds 1. Rate does not scale the delta. Then run TSW-R05.

**TSW-R06b. Supplied client wrapper** [verified in 1.13c code (static)]. Signed source <40 gives act 0; <75 act 1; <103 act 2; <109 act 3; <1024 act 4; otherwise act 0. Compare threshold entries 1..5 only. Save old period; call the core with this act; request intensity with source, then colour with derived act. Source 120 overwrites colour bytes `(245,240,255)` afterward. Return 1 if period changed, else 0, even if cycle changed. Source production remains TSW-Q07; supplied mapping is available independently.

### TSW-R07. Messages and intensity

**TSW-R07a. Start/reset requests** [verified in 1.13c code (static)]. Per Act II player, start requests World-clock `(5,0,1)` then Quest-event `(quest 10,flags 16)`. Server start is cycle 0/ticks 1200, distinct from the message. Reset requests only World-clock `(2,0,0)`. Other acts get neither request. Both per-player callbacks return 0. The act is the recipient client's act; client state is not tested. Before the act test, each callback requires a present player unit with a client; an absent unit, non-player unit or player without a client takes the fatal assertion path rather than being skipped. Delivery/sound remain TSW-Q10.

**TSW-R07b. Missing-act join** [verified in 1.13c code (static)]. Failed join apply requests Quest-event `(quest 10,flags 1)` to the joining player's client without act filtering; no start World-clock. The client is taken from player data only when the joining unit is present and is a player; otherwise the request carries no client. The send operation returns without queuing anything when it has no client. Earlier quest/pending writes remain in every case.

**TSW-R07c. Supplied intensity branches** [verified in 1.13c code (static)]. For act other than 3 and stored flag nonzero, subtract 8 if signed intensity>32, then floor at 32. Ignore period. For act other than 3, flag 0, source 120, assign 200 directly. Other formulas/rendering are TSW-Q03. Each call updates the stored intensity field from its current value, so start's two calls compound.

**TSW-R07d. Appearance** [unresolved]. Darkness/colour/audio/timing require TSW-Q03/TSW-Q10; a clock tuple does not establish appearance.

**TSW-R07e. Diagnostic broadcast** [verified in 1.13c code (static)]. At the sites named in TSW-R01a, TSW-R01i and TSW-R02b, if the diagnostic switch is nonzero, request one game-wide chat-style text message carrying the stated text. A zero switch requests nothing; text formatting alone has no visible effect. Recipient iteration and delivery are TSW-Q10.

### TSW-R08. Retry and repetition

**TSW-R08a. Direct retry** [verified in 1.13c code (static)]. Supplied act argument 1, found quest 10, pending exactly 1: apply TSW-R04a directly, clear pending, set applied 1; request no start messages. Other arguments/missing quest/pending!=1 do nothing. This stage does not check Act II for absence; valid input requires its act/environment. Outer dispatch/absence behavior remain TSW-Q10.

**TSW-R08b. Repetition/difficulty** [unresolved]. Selection consumes supplied current-difficulty flags. Saved/repeated progress production remains TSW-Q05.

### TSW-R09. Join and persistence

**TSW-R09a. Rebuild/snapshot** [verified in 1.13c code (static)]. Fresh clocks begin at TSW-R04c, advance on server frames and can be changed by TSW-R01d. Snapshot selects the joining client's current act and requests act-assignment first, World-clock with that act's current `(cycle,ticks,flag)` second. Values are sampled then; relative order to quest join needs supplied scheduling. This stage makes no environment save write.

**TSW-R09b. Persistence** [unresolved]. Read-stage save absence does not prove absence from every save path. Quest-bit encoding/load, repetition and wider clock persistence remain TSW-Q05.

### TSW-R10. Server/client execution

**TSW-R10a. Server frame** [verified in 1.13c code (static)]. Each normal frame increments the game frame, then visits acts 0..4 in order. Skip absent acts; run TSW-R10d for existing acts with the **act index**, independently of client source. Continue after no-notification results. The client threshold wrapper is not used here.

**TSW-R10b. Client cadence** [asm-only]. The identified client update calls TSW-R06b once, then requests refresh on nonzero return. Complete outer scheduling/source identity remain TSW-Q07. Idle progress does not prove every outer branch.

**TSW-R10c. Client assignment** [verified in 1.13c code (static)]. The supplied handler eligibility decides first: false makes no assignment and no request. When eligible, the client's current-act environment is the target and is required, and message cycle outside 0..5 or negative ticks takes the fatal assertion path before any clock write or request. Retain rate/index/last resync. If message ticks >`i32(rate*360)`, first use 0; equality is accepted. Assign cycle/ticks and period from the table selected by **message** flag. Request intensity (client source, its derived act), observing the old stored flag, then store message flag. Nonzero new flag recomputes period/ticks from its eclipse row using retained rate, then requests intensity again and colour. Supplied packet ticks are superseded there. Source 120 applies the colour override afterward, whether or not a colour request was made. Later handler refresh/consumers remain TSW-Q10.

**TSW-R10d. Server resync** [verified in 1.13c code (static)]. Existing act/absent environment returns 0. Otherwise save cycle/period, run core with server act index. Rate nonzero: signed `q=ticks/rate`; `d=i32(q-last resync)`; if d<0 set `d=i32(-d)`. Signed d>16 stores last resync=q and returns 1. Otherwise return 1 only if cycle or period changed, else 0. Rate 0 skips division; negating i32 minimum leaves it negative.

Return 1 requests a current-clock read(act index), which yields that act's post-update cycle/ticks/flag for the World-clock payload, then traverses supplied clients in list order, completing each client's resolution and requests before the next client. Resolve each client's player: with the client's direct-player flag (bit 0) set, use its cached player reference as is. Otherwise a zero cached reference skips the client without a lookup; a nonzero one requests the supplied player lookup(client), and a failed lookup writes the cached reference to zero. Skip any client whose resolved player is absent. For each resolved player request player-notification(game,player,0), then change-tracking(player,1), then request World-clock only for client state 4 and client act equal to the server act index. Thus other state/act clients can receive pre-filter requests but no clock message. Delivery/helper effects remain TSW-Q10. Cycle/period notification alone retains last resync. Drift testing includes wrap to 0.

### TSW-R11. Other consumers

**TSW-R11** [unresolved]. Complete consumers, including a possible stamina relationship, remain TSW-Q06. Bounded getter reading establishes no absence or stamina formula.

### TSW-R12. RNG

**TSW-R12a. Direct no-draw slice** [verified in 1.13c code (static)]. Eight audited bodies—direct retry, reset/start send callbacks, quest reset, shared start, join apply, join selection, quest initialization—contain no inline RNG step/direct helper call. Their own bodies leave game/unit/quest seeds unchanged. Called callbacks, allocation, presentation, delivery and drops are excluded; this is not a whole-scenario zero-draw guarantee.

**TSW-R12b. Quest-owned sites** [verified in 1.13c code (static)]. Schedule makes one quest-seed draw before requesting the timer. A reached altar inline site steps that same quest seed once after amulet requests, then requests `5+(new lo mod 5)` gold-pile drops, 5..9. No direct game/unit draw occurs at these sites. Intervening/helper RNG is TSW-Q08; the seed at the altar draw site is supplied explicitly.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| TSW-R01a | shared start | verified in 1.13c code (static) | TSW-Q10: callback/delivery effects only | TSW-V01, TSW-V07 |
| TSW-R01b | progress writer | verified in 1.13c code (static) | none | TSW-V20 |
| TSW-R01c | upstream identity | unresolved | TSW-Q01: gameplay event production | none |
| TSW-R01d | join selection | verified in 1.13c code (static) | TSW-Q05/TSW-Q10: saved flags/helper integration | TSW-V06, TSW-V07, TSW-V16 |
| TSW-R01e | animation slice | verified in 1.13c code (static) | TSW-Q10: mode-update/iteration effects | TSW-V21 |
| TSW-R01f | event retry slice | verified in 1.13c code (static) | TSW-Q01/TSW-Q10: event production/other effects | TSW-V21 |
| TSW-R01g | schedule | verified in 1.13c code (static) | TSW-Q10: actual timer/elapsed units | TSW-V18 |
| TSW-R01h | timer callback | verified in 1.13c code (static) | TSW-Q10: delivery/other effects | TSW-V21 |
| TSW-R01i | event auxiliary-identity slice | verified in 1.13c code (static) | TSW-Q01/TSW-Q10: event production/roster-update effects | TSW-V21 |
| TSW-R02a | altar eligibility | verified in 1.13c code (static) | none for supplied item tests | TSW-V08, TSW-V26 |
| TSW-R02b | altar active branch | verified in 1.13c code (static) | TSW-Q08/TSW-Q10: drops/callback effects | TSW-V02, TSW-V08, TSW-V26 |
| TSW-R02c | altar inactive branch | verified in 1.13c code (static) | TSW-Q08/TSW-Q10: drops/callback effects | TSW-V18, TSW-V26 |
| TSW-R02d | altar timer callback | verified in 1.13c code (static) | TSW-Q10: timer delivery | TSW-V26 |
| TSW-R03 | ownership | verified in 1.13c code (static) | none for supplied recipients | TSW-V09, TSW-V16 |
| TSW-R04a | start clock | verified in 1.13c code (static) | TSW-Q03/TSW-Q10: presentation | TSW-V01, TSW-V06, TSW-V07 |
| TSW-R04b | reset clock | verified in 1.13c code (static) | TSW-Q03: presentation | TSW-V02, TSW-V08 |
| TSW-R04c | allocation | verified in 1.13c code (static) | TSW-Q10: allocation/presentation | TSW-V16 |
| TSW-R05 | wrap/transition | verified in 1.13c code (static) | none | TSW-V03, TSW-V05, TSW-V10, TSW-V11, TSW-V12, TSW-V13, TSW-V14, TSW-V22 |
| TSW-R06a | advance | verified in 1.13c code (static) | none | TSW-V03, TSW-V04, TSW-V05, TSW-V15, TSW-V17 |
| TSW-R06b | client source | verified in 1.13c code (static) | TSW-Q07: source production; TSW-Q03/TSW-Q10: presentation | TSW-V23 |
| TSW-R07a | messages | verified in 1.13c code (static) | TSW-Q10: delivery/sound; TSW-Q11: observed client effect of the payload/server divergence only | TSW-V01, TSW-V02, TSW-V09 |
| TSW-R07b | pending request | verified in 1.13c code (static) | TSW-Q10: client resolution/delivery | TSW-V07 |
| TSW-R07c | intensity branches | verified in 1.13c code (static) | TSW-Q03: other formulas/rendering | TSW-V24 |
| TSW-R07d | appearance | unresolved | TSW-Q03/TSW-Q10: visual/audio integration | none |
| TSW-R07e | diagnostic broadcast | verified in 1.13c code (static) | TSW-Q10: switch producer/delivery | TSW-V01, TSW-V21, TSW-V26 |
| TSW-R08a | direct retry | verified in 1.13c code (static) | TSW-Q10: dispatch/absence | TSW-V21 |
| TSW-R08b | repetition/difficulty | unresolved | TSW-Q05: progress production | none |
| TSW-R09a | join snapshot | verified in 1.13c code (static) | TSW-Q10: outer order/delivery | TSW-V16 |
| TSW-R09b | persistence | unresolved | TSW-Q05: saving/loading | none |
| TSW-R10a | server frame | verified in 1.13c code (static) | TSW-Q10: dependency effects | TSW-V17, TSW-V19 |
| TSW-R10b | client cadence | asm-only | TSW-Q07: outer scheduling | none |
| TSW-R10c | client assignment | verified in 1.13c code (static) | TSW-Q10: refresh/delivery | TSW-V25 |
| TSW-R10d | server resync | verified in 1.13c code (static) | TSW-Q10: resolution/notifications/delivery | TSW-V17, TSW-V19, TSW-V22 |
| TSW-R11 | other consumers | unresolved | TSW-Q06: discovery | none |
| TSW-R12a | eight direct bodies | verified in 1.13c code (static) | TSW-Q08: reachable draws excluded | TSW-V18 |
| TSW-R12b | quest draw sites | verified in 1.13c code (static) | TSW-Q08: helper draws excluded | TSW-V18 |

## 5. Outputs and side effects

| Effect | Rule |
|---|---|
| shared-start return 0/1, progress-writer return 0, altar return 0, altar-timer return 1, client-wrapper period-change return 0/1 | TSW-R01a, TSW-R01b, TSW-R02a through TSW-R02d, TSW-R06b |
| quest state/stage/transition writes and retained failure state | TSW-R01a, TSW-R01d, TSW-R01e, TSW-R01i, TSW-R02d |
| markers A/B/C, mode argument, unit marker, source result, derived result, amulet count, quest 9 counter, object item-code field | TSW-R01e, TSW-R02b through TSW-R02d |
| per-player bits 2/3/4 and shared bit 13 | TSW-R01b, TSW-R01d |
| scheduled/applied/pending writes, timer requests and callback returns | TSW-R01d through TSW-R01h, TSW-R02b, TSW-R02d, TSW-R08a |
| rejection, mode-update, roster-update, object-source/source-derived, participant/post-flag/completion iterations, quest/game notifications, quest-record update, amulet/gold drop, item-flag mark, amulet-result and context requests; player bits 1/13 set and 14 clear | TSW-R01e, TSW-R01i, TSW-R02a through TSW-R02c |
| pre-start iteration and progress-writer iteration requests | TSW-R01a, TSW-R01e, TSW-R01f, TSW-R01h, TSW-R02d |
| schedule requests and their due values | TSW-R01g, TSW-R02b |
| clock rate/index/cycle/ticks/period/flag | TSW-R04a through TSW-R04c, TSW-R05, TSW-R06a, TSW-R10c |
| intensity, presentation requests, colour override | TSW-R04a through TSW-R04c, TSW-R06b, TSW-R07c, TSW-R10c |
| game frame, last resync, notification result | TSW-R10a, TSW-R10d |
| World-clock, Quest-event, act-assignment, diagnostic text requests | TSW-R07a, TSW-R07b, TSW-R07e, TSW-R09a, TSW-R10d |
| fatal assertion path for invalid sender recipients, a missing game quest structure, or an invalid client message cycle/ticks | TSW-R02b, TSW-R02c, TSW-R07a, TSW-R10c |
| current-clock read, pre-filter player-notification and change-tracking requests, player lookups, cached-reference clear | TSW-R10d |
| quest-seed writes, schedule parameter, gold-pile count | TSW-R01g, TSW-R02b, TSW-R02c, TSW-R12b |

The read stages make no environment save write; wider persistence remains gated. Message payloads belong to the rules; this table names effects only.

## 6. Branches and edge cases

Server act index and client derived source can differ. Clock flags use nonzero, while start active checks use equality to 1 and altar active uses nonzero. The animation slice writes marker B/unit marker and requests mode-update even when active is not 1. The event auxiliary-identity slice is independent of the event key, so one event can run a start and then change state 2 to 3. Negative source maps to act 0. Overflowed negative ticks use signed comparisons without saturation. Core wrap uses >=, transition uses >, client pre-clamp uses >. A cycle can change without a period change or client-wrapper return 1.

Missing Act II during shared start retains quest writes/pending; missing Act II during reset retains applied. Missing environment during server update returns 0. Invalid required assignment objects take assertion paths outside normal vectors; per-player senders abort the same way on an invalid recipient, whereas the failed-join request silently queues nothing without a client. A later failed delivery does not erase earlier local writes. Classic mode and other builds have no established equivalence here.

## 7. Open questions

The [gap file](../plans/gaps/quest-tainted-sun-world-clock.md#open-evidence-gaps) distinguishes behavioral, integration, observation and outside-boundary questions.

- **TSW-Q01** [unresolved]: establish event/auxiliary-selector producers and Lost City relationship with root-to-callback evidence. Supplied event selection remains available.
- **TSW-Q02** [unresolved]: establish gameplay meanings of neutral flags/markers from anchored readers/writers. Exact bit operations are available.
- **TSW-Q03** [unresolved]: complete presentation and observe darkness/colour/timing. Only two supplied intensity branches are reproduced; general lighting is outside scope.
- **TSW-Q04** [unresolved]: observe encoded messages and client handling. Static wire records are in the companion; no packet capture is available.
- **TSW-Q05** [unresolved]: establish quest persistence/repeated difficulty production and any other clock-saving path. Bounded join reading proves no universal absence.
- **TSW-Q06** [unresolved]: complete other-consumer discovery, including any stamina connection.
- **TSW-Q07** [unresolved]: establish client source identity and full outer cadence. Supplied threshold arithmetic/server act-index behavior do not depend on this answer.
- **TSW-Q08** [unresolved]: complete reachable RNG budgets for drop helpers, callbacks, other handlers, presentation and delivery. Direct-body/site budgets remain available.
- **TSW-Q09** [unresolved]: observe server notification/send decisions, transition/wrap and client/server divergence. Idle reads do not reproduce exact resync vectors.
- **TSW-Q10** [unresolved]: independently establish timer allocation/delivery, player lookup/notifications, remaining quest/altar/mode/roster/drop/iteration callbacks, the diagnostic switch's producer, presentation and transmission. Local requests/results are explicit; complete integrated effects remain gated.
- **TSW-Q11** [unresolved]: the start assignment (TSW-R04a) leaves the server at cycle 0 with eclipse flag 1, yet the start message (TSW-R07a) sends the fixed cycle 5. Is this divergence intended, and which later resync (TSW-R10d) or client step realigns the client? Both values are statically exact; only observation (TSW-O01) or a rooted client consumer can settle the effect.

## 8. Test vectors

### 8.1 Executable vectors

Each case is fresh unless stated. Default clock tuple `(cycle,period,ticks,rate,index,flag,last resync)` is `(2,0,0,128,0,0,0)`. Default quest `(active,state,stage,transition,scheduled,applied,pending)` is `(1,0,0,7,0,0,0)`; markers A/B/C, mode argument, unit marker, source result, derived result, amulet count and roster word are 0, the diagnostic switch is 0, event auxiliary identity is 0, game quest counter is 1000, and callback/event unit U has identity 77. Act II/environment and valid player P in Act II exist (P has a client whose act is 1 and state is 4); flags are clear, other acts absent unless a case names an act or clock for them (such a case supplies that act and environment), no intervening frame, empty log. A "resolved" client has its direct-player flag set and a cached reference to its player, so it makes no lookup request; an "unresolved" client has the flag clear and a zero cached reference, so it makes no lookup request and no other request. Altar logs list flag and item tests as entries in the request log. Core vectors run advance then transition with their supplied act. Stage vectors log dependency requests in order; supplied callbacks/presentation/sends return normally without unlisted state or seed changes. A call-time snapshot below records the semantic state observed immediately on entering that supplied operation; it adds no function argument and no new stored field. Flag writes persist within a case. Presentation leaves clock fields unchanged. These supplied effects do not verify integration.

Except explicit draw cases, Seeds: no draw; Draws: none within the local slice. Hidden helper draws are excluded. All complete vectors remain runtime_observed: no; partial idle readings are limited to the assertions in the companion.

- **TSW-V01** Shared start/payload distinction (TSW-R01a, TSW-R04a, TSW-R07a, TSW-R07e; boundary: supplied-input stage).
  - Inputs: defaults; pre-start iteration completes for P.
  - Seeds: no draw. Draws: none locally.
  - Trace: state 0 becomes 1; initial ticks 0 is replaced by 4*300=1200.
  - Outputs: return 1; quest state/stage/transition 1/1/0, applied/pending 0/0; clock `(0,3,1200,4,1,1,0)`. Log: pre-start(P) [call-time snapshot state 0/stage 1/transition 0], first intensity [call-time snapshot flag 0/cycle 0/period 3/ticks 0/rate 4/index 1], second intensity [snapshot flag 1/cycle 0/period 3/ticks 1200/rate 4/index 1], colour, World-clock(P,5,0,1), Quest-event(P,10,16).
  - Fresh Inputs: shared-start defaults except quest state 2. Outputs: return 1; state/stage/transition 2/1/0, applied/pending 0/0, clock `(0,3,1200,4,1,1,0)`. Request log: pre-start(P), first intensity [flag 0/cycle 0/period 3/ticks 0/rate 4/index 1], second intensity [flag 1/cycle 0/period 3/ticks 1200/rate 4/index 1], colour, World-clock(P,5,0,1), Quest-event(P,10,16). These are the same call-time snapshot meanings as above.
  - Fresh Inputs: shared-start defaults except diagnostic switch 1 (TSW-R07e). Outputs: as the first case; log pre-start(P), diagnostic text(`changing state to 1` plus newline), then the first case's remaining requests in order. With state 2 and switch 1 instead, no diagnostic text is requested. With state 1 and switch 1: Outputs return 1, state/stage/transition 1/1/0, the first case's clock and request log, and no diagnostic text. With last resync 17 instead: Outputs as the first case with clock `(0,3,1200,4,1,1,17)`. With state 0, switch 1 and Act II absent: Outputs return 0, state/stage/transition 1/1/0, pending 1, applied 0, no clock; log pre-start(P), diagnostic text(`changing state to 1` plus newline); no presentation or start message.
  - Discriminates: final ticks 1200 differs from payload ticks 0; reversed message order, storing flag 1 before first intensity, or resetting a nonzero quest state to 1 fails. A broadcast after the Act II test (Act II absent case), one for a retained nonzero state, or stage/transition/state writes on the wrong side of the pre-start request (its snapshot) fails.
  - runtime_observed: no.

- **TSW-V02** Reset retaining resync (TSW-R02b, TSW-R04b, TSW-R07a; boundary: supplied quest-reset stage).
  - Inputs: clock `(4,2,800,4,1,1,17)`, applied 1. Enter reset after supplied altar preparation.
  - Seeds: no draw. Draws: none locally.
  - Trace: clear applied, normal row 2; intensity sees incoming flag 1, then clear flag.
  - Outputs: applied 0; clock `(2,0,0,128,0,0,17)`; intensity [call-time snapshot flag 1/cycle 2/period 0/ticks 0/rate 128/index 0/last resync 17/applied 0], then World-clock(P,2,0,0); no colour or Quest-event request.
  - (b) As above, plus player P2 in act 0 with a client, after P in iteration order: Outputs the same clock/applied and log; no World-clock for P2.
  - (c) As the first case, but P's client state is 3, and Act I also exists with clock `(3,1,20000,128,0,0,0)`: Outputs the same Act II clock/applied and log including World-clock(P,2,0,0); Act I unchanged.
  - Discriminates: retained 17 rejects full reinitialization; clearing the flag before intensity fails its snapshot; (b) rejects an unfiltered reset send.
  - runtime_observed: no.

- **TSW-V03** Act IV advance (TSW-R06a; boundary: core tick).
  - Inputs: act 3, rate 128, ticks 100, cycle 2, period 0, flag 0.
  - Seeds: no draw. Draws: none.
  - Trace: 100+1+15=116, below next threshold 20480.
  - Outputs: cycle/period/ticks 2/0/116; other fields retained.
  - (b) Fresh act 3, flag 0, rate 128, cycle 2, period 0, ticks 20470. Trace: 20470+16=20486 > normal row 3 threshold 128*160=20480; assign normal row 3. Outputs cycle 3, period 1, ticks 20480.
  - Discriminates: +1/+2 fails; (b) rejects comparing a flag-0 Act IV clock with the Act IV comparison table (threshold 23040 would leave cycle 2/ticks 20486).
  - runtime_observed: no.

- **TSW-V04** Act III night (TSW-R06a; boundary: core tick).
  - Inputs: act 2, rate 128, cycle 5, period 2, ticks 30000, flag 0.
  - Seeds: no draw. Draws: none.
  - Trace: 30000+1+1+8=30010, below 40960.
  - Outputs: cycle/period/ticks 5/2/30010.
  - (b) Fresh act 2, flag 0, rate 128, cycle 5, period field 0 (supplied inconsistent with normal row 5), ticks 30000. Outputs cycle/period/ticks 5/0/30010.
  - Discriminates: +11 gives 30011; (b) rejects reading the stored period field instead of normal[cycle].period (that gives 30001).
  - runtime_observed: no.

- **TSW-V05** Eclipse rate-independent advance (TSW-R05, TSW-R06a; boundary: core tick).
  - Inputs: act 1, cycle 2, period 1, ticks 300, rate 4, index 1, flag 1.
  - Seeds: no draw. Draws: none.
  - Trace: 301 does not exceed 4*120=480.
  - Outputs: cycle/period/ticks 2/1/301, rate 4 retained.
  - Discriminates: scaling delta by rate gives 304.
  - runtime_observed: no.

- **TSW-V06** Fresh join rebuild (TSW-R01d, TSW-R04a, TSW-R09a; boundary: supplied stages in stated order).
  - Inputs: fresh allocation, P flags {2}; run allocation, quest join, snapshot without a frame.
  - Seeds: no draw. Draws: none locally.
  - Trace: bit 2 applies start, state overwrite selects 1, snapshot samples final clock.
  - Outputs: clock `(0,3,1200,4,1,1,0)`, applied 1, state/stage 1/1. Messages: World-clock(P,5,0,1), Quest-event(P,10,16), act-assignment(P,1), World-clock(P,0,1200,1).
  - Discriminates: start and snapshot payload differ. Outer join scheduling is supplied, not inferred.
  - runtime_observed: no.

- **TSW-V07** Join flags/failure (TSW-R01a, TSW-R01d, TSW-R07b; boundary: supplied join stage).
  - Inputs: fresh defaults, shared flags empty, overrides below.
  - Seeds: no draw. Draws: none locally.
  - Trace: primary bits precede bit 2; state overwrite follows either apply result.
  - (a) Flags {2}: Outputs started clock `(0,3,1200,4,1,1,0)`, applied 1, pending 0, state/stage 1/1, V01 messages.
  - (b) Flags {2,1}: Outputs default clock, state/stage 0/0, shared-record slot-11 bit 13 set, P's flags still {1,2}; no start/pending message.
  - (c) Flags {2,0}: Outputs as (b).
  - (d) Flags {}: Outputs default state/clock; no flag/message writes.
  - (e) Flags {2,3,4}, Act II absent, P act 0: Outputs state/stage/transition 3/2/0, pending 1, applied 0, no clock assignment; log pre-start(P), Quest-event(P,10,1), no World-clock.
  - (f) Flags {2,3}, bit 4 clear, Act II present: Outputs clock `(0,3,1200,4,1,1,0)`, applied 1/pending 0, state/stage/transition 2/2/0; player flags remain {2,3}; V01 start message/presentation requests in order.
  - (g) Flags {2}, Act II absent, P's player data has no client: Outputs pending 1, applied 0, state/stage/transition 1/1/0; log pre-start(P), Quest-event request with no client, which queues nothing.
  - (h) Flags {1} alone, and separately {0} alone: Outputs default clock, state/stage 0/0, shared-record bit 13 set, P's flags unchanged; no start/pending message.
  - Discriminates: (h) rejects testing bit 2 before the primary bits or requiring bit 2 for the bit-13 write; (e) rejects rollback, bit-3 precedence, and filtering the pending request by act; (f) requires the bit-3-only state/stage selection; (g) rejects aborting or undoing writes when no client exists.
  - runtime_observed: no.

- **TSW-V08** Altar short circuit/flags (TSW-R02a, TSW-R02b; boundary: supplied reset/flag slice).
  - Inputs: clock `(0,3,1200,4,1,1,0)`, applied 1, object mode marker 0, supplied `vip ` true; intervening callbacks complete without flag/state changes. Drop/later quest work is excluded.
  - Seeds: no draw before this slice's endpoint. Draws: none; later draw is V18.
  - Trace: primary bits decide whether secondary tests run.
  - (a) Flags {14}, bits 0/1 clear: Outputs no secondary tests, clock `(2,0,0,128,0,0,0)`, applied 0, reset World-clock(P,2,0,0); ordered set 1,set 13,clear 14 yields flags {1,13}; return 0 after supplied remaining work.
  - (b) Flags {1,14}, slot-10 bit 0 false: Outputs `vip ` check true, rejection(19,P), return 0; clock/applied/flags unchanged, no reset or `hst ` test.
  - (c) Flags {14}, object mode marker 2, and separately 1 and -1: Outputs return 0, unchanged clock/applied/flags, no reset/secondary tests.
  - (d) Flags {14}, Act II absent (so the preamble's clock is not present): Outputs no clock write or World-clock request, applied stays 1; flags {1,13} after supplied callbacks; return 0.
  - (e) Flags {1,14}, slot-10 bit 0 true: Outputs rejection(19,P), return 0, unchanged clock/applied/flags; neither item-presence test is requested and no reset message occurs.
  - (f) Flags {1,14}, slot-10 bit 0 false, supplied `vip ` false/`hst ` true: Outputs those three secondary tests in order, rejection(19,P), return 0, unchanged clock/applied/flags, no reset message.
  - (g) Eligibility-only boundary: flags {14}, operated object absent. Outputs no mode-marker read, no eligibility rejection/return, continuation to supplied preparation. Later object-dependent callback/reset/drop completion is TSW-Q10 and is not asserted by this case.
  - (m) Flags {1,14}, slot-10 bit 0 true, object mode marker 2: Outputs rejection(19,P), return 0, unchanged clock/applied/flags; the mode marker does not pre-empt the rejection.
  - (h) Flags {0,14}, bit 1 clear, slot-10 bit 0 true: Outputs primary bit-1 test false then bit-0 test true, slot-10 test true, rejection(19,P), return 0; unchanged clock/applied/flags, no item-presence/reset requests.
  - (i) Quest 10 absent, with no private-state inputs for that absent quest: Outputs return 0, input clock/player flags retained; no quest record read/write, player-flag lookup, preparation, reset, rejection or draw from this slice.
  - (j) Active 2, flags {14}, Act II present, other inputs and supplied effects from this vector's preamble: Outputs no secondary tests, clock `(2,0,0,128,0,0,0)`, applied 0, reset World-clock(P,2,0,0), ordered set 1/set 13/clear 14, flags {1,13}; return 0 after supplied remaining work. This slice still excludes drops/later quest work.
  - (k) Flags {1,14}, slot-10 bit 0 false, supplied `vip ` false and `hst ` false: Outputs primary bit-1 test true, then slot-10, `vip ` and `hst ` tests all false, no rejection; clock `(2,0,0,128,0,0,0)`, applied 0, reset World-clock(P,2,0,0); bit-0 test false, set 1, set 13, clear 14, flags {1,13}; return 0 after supplied remaining work.
  - (l) Flags {0,14}, slot-10 bit 0 false, supplied `vip ` false and `hst ` false: Outputs bit-1 test false, bit-0 test true, three secondary tests false, no rejection; clock `(2,0,0,128,0,0,0)`, applied 0, reset World-clock(P,2,0,0); the later bit-0 test is true, so no set/clear request and flags stay {0,14}; return 0 after supplied remaining work.
  - Discriminates: clear 14 rejects setting 14; (k) rejects rejecting whenever a primary bit is set; (l) rejects writing bits 1/13/14 regardless of bit 0; (d) rejects clearing applied before checking act; (e)/(f)/(h) reject omitted/incorrect flag/item tests; (g) rejects treating an absent object as a nonzero mode marker; (i) rejects doing quest work without the quest; (m) rejects testing the mode marker before eligibility; (j) rejects treating altar active as equality to 1.
  - runtime_observed: no.

- **TSW-V09** Act recipients (TSW-R03, TSW-R07a; boundary: supplied start).
  - Inputs: P act 1 then P2 act 0, each with a client; Act I clock `(3,1,20000,128,0,0,0)`, default Act II; no frame.
  - Seeds: no draw. Draws: none locally.
  - Trace: apply addresses Act II; recipient loop rejects P2.
  - Outputs: Act I unchanged; Act II `(0,3,1200,4,1,1,0)`; P gets V01 requests in order, P2 none.
  - (b) Fresh start or reset whose player iteration supplies P (act 1, with client) then P2 in act 0 with no client: Outputs P's sender requests, then the fatal assertion path at P2; no World-clock/Quest-event request for P2 and no normal return from that sender.
  - (c) Fresh start with P then P3, both act 1 with clients: Outputs clock as above; log after presentation World-clock(P,5,0,1), Quest-event(P,10,16), World-clock(P3,5,0,1), Quest-event(P3,10,16).
  - Discriminates: all-act reset/unfiltered send fails; (b) rejects silently skipping a clientless recipient in any act; (c) rejects batching all World-clock requests before all Quest-event requests.
  - runtime_observed: no.

- **TSW-V10** Strict transition exceeded (TSW-R05; boundary: core tick).
  - Inputs: act 0, cycle 3, period 1, ticks 23040, rate 128, flag 0.
  - Seeds: no draw. Draws: none.
  - Trace: 23041>128*180=23040; assign normal row 4.
  - Outputs: cycle 4, period 1, ticks 23040.
  - Discriminates: keeping advanced tick 23041 fails.
  - runtime_observed: no.

- **TSW-V11** Strict equality (TSW-R05; boundary: core tick).
  - Inputs: V10 except ticks 23039.
  - Seeds: no draw. Draws: none.
  - Trace: 23040 equals threshold, no transition.
  - Outputs: cycle 3, period 1, ticks 23040.
  - Discriminates: >= changes cycle to 4.
  - runtime_observed: no.

- **TSW-V12** Wrap without cycle change (TSW-R05; boundary: core tick).
  - Inputs: act 0, cycle 1, period 3, ticks 46079, rate 128, flag 0.
  - Seeds: no draw. Draws: none.
  - Trace: 46080>=128*360; wrap 0; next threshold 0 is not exceeded.
  - Outputs: cycle 1, period 3, ticks 0.
  - (b) Fresh act 0, flag 0, rate 128, cycle 5, period 2, ticks 40960. Trace: night advance 40962; below wrap 46080; next wraps 6 to 0, normal row 0 threshold 128*320=40960 exceeded. Outputs cycle 0, period 3, ticks 40960.
  - Discriminates: transition-before-wrap/resetting cycle fails; (b) rejects mishandling next cycle 6.
  - runtime_observed: no.

- **TSW-V13** Supplied zero and returned start state (TSW-R05, TSW-R04a; boundary: fresh core ticks).
  - Inputs: act 1, rate 4, index 1, flag 1, cycle 0, period 3.
  - Seeds: no draw. Draws: none.
  - Trace: advance then compare eclipse row 1 begin 0.
  - (a) Supplied ticks 0: Outputs advanced 1>0, cycle 1, period 0, ticks 0.
  - (b) Returned start ticks 1200: Outputs advanced 1201<1440 but >0, cycle 1, period 0, ticks 0.
  - Discriminates: (a) is a supplied fixture; it is not the returned start ticks value.
  - runtime_observed: no.

- **TSW-V14** Act IV compare/assign (TSW-R05; boundary: core tick).
  - Inputs: act 3, flag 1, rate 4, cycle 2, period 1, ticks 720.
  - Seeds: no draw. Draws: none.
  - Trace: 721>Act IV row 3 threshold 4*180=720; assign eclipse row 3 begin 120/period 2.
  - Outputs: cycle 3, period 2, ticks 480.
  - (b) Fresh act 3, flag 1, rate 4, index 1, cycle 2, period 1, ticks 650. Trace: 651 does not exceed Act IV row 3 threshold 720. Outputs cycle 2, period 1, ticks 651.
  - Discriminates: assigning comparison table gives ticks 720/period 1; (b) rejects comparing with the eclipse (480) or normal (640) table and adding the act-3 +15 under flag 1 (666).
  - runtime_observed: no.
- **TSW-V15** Advance branches (TSW-R06a; boundary: fresh core ticks).
  - Inputs: normal rate 128; fresh cases below; next threshold never reached.
  - Seeds: no draw. Draws: none.
  - Trace: one advance using supplied act and normal row period.
  - (a) cycle 5, period 2, ticks 30000, flag 0, acts 0/1/2/3/4: Outputs ticks 30002/30002/30010/30016/30002 respectively; cycle/period 5/2 retained.
  - (b) cycle 4, period 1, ticks 25000, flag 0, acts 2/3: Outputs ticks 25001/25016; cycle/period 4/1 retained.
  - (c) act 1, cycle 2, period 1, ticks 300, rate 4, index 1, flag 1: Outputs ticks 301, cycle/period 2/1 retained.
  - (d) cycle 5, period 2, ticks 300, rate 4, index 1, flag 1, acts 1/2/3: Outputs ticks 301 for each; cycle/period 5/2 retained (next thresholds 1200, or 1360 for act 3, not exceeded).
  - Discriminates: normal row chooses night; Act IV overrides it; (d) rejects applying night, Act III or Act IV bonuses to a nonzero flag.
  - runtime_observed: no.

- **TSW-V16** Allocation/own-act snapshot (TSW-R01d, TSW-R04c, TSW-R09a; boundary: supplied stages).
  - Inputs: allocate Act II, quest join, snapshot; no frame; for (c) Act I also exists with supplied clock `(3,1,20000,128,0,0,0)`, not reallocated.
  - Seeds: no draw. Draws: none locally.
  - Trace: allocation gives `(2,0,0,128,0,0,0)` before supplied presentation; join selects progress; snapshot selects own act.
  - (a) P act 1, flags {2}: Outputs Act II `(0,3,1200,4,1,1,0)`, applied 1; snapshot act-assignment(P,1), World-clock(P,0,1200,1), after V01 start messages.
  - (b) P act 1, flags {2,1}: Outputs Act II allocation tuple, shared-record bit 13 set, P's flags still {1,2}, applied 0; act-assignment(P,1), World-clock(P,2,0,0), no start message.
  - (c) P act 0, flags {2}: Outputs Act II started tuple, Act I unchanged; no start messages for P; act-assignment(P,0), World-clock(P,3,20000,0).
  - Discriminates: (c) rejects always snapshotting Act II.
  - runtime_observed: no.

- **TSW-V17** Exact 1000-frame static scenario (TSW-R06a, TSW-R10a, TSW-R10d; boundary: supplied server frames/client core calls).
  - Inputs: Act I default server/client clocks; other acts absent; 1000 normal server frames and 1000 client core calls with source/act 0; no external messages/start/reset; one resolved state-4 act-0 client.
  - Seeds: no draw. Draws: none in clock slices.
  - Trace: each adds 1; 1000<20480; q reaches 7, never distance >16 from 0.
  - Outputs: both cycle 2/period 0/ticks 1000; server last resync 0; game frame increases 1000; no World-clock/pre-filter request or cycle change.
  - Discriminates: rate-scaled delta/notification at distance <=16 fails. Idle measurements are partial corroboration and do not reproduce this exact case.
  - runtime_observed: no.

- **TSW-V18** Seed owners/budgets (TSW-R01g, TSW-R02c, TSW-R12a, TSW-R12b; boundary: direct bodies and supplied draw sites).
  - Inputs: game seed {7,666}, unit seed {9,666}; each case is fresh; direct bodies exclude helpers.
  - Seeds: each quest seed stated below. Draws: none in (a)/(d)/(f), one quest step in (b)/(c)/(e)/(g)/(h).
  - Trace: NRF-R01 step then the specified reduction; no scenario-wide budget inferred.
  - (a) Direct-body case, quest seed {0,666}: Outputs all game/unit/quest seeds unchanged by the eight bodies.
  - (b) state 0/scheduled 0, quest seed {0,666}: Draws quest modulus 2 result 0; parameter 15. Outputs seed {666,0}, schedule(timer,15) with due value 1000+15=1015 [call-time snapshot scheduled 0, quest seed {666,0}], then scheduled 1.
  - (c) state 0/scheduled 0, quest seed {1,666}: Draws quest modulus 2 result 1; parameter 16. Outputs seed {1791398751,0}, schedule(timer,16) with due value 1016, then scheduled 1.
  - (d) state 0/scheduled 1, quest seed {0,666}: Draws none; Outputs unchanged seed, no schedule, scheduled 1.
  - (e) Altar inline site after supplied amulet requests, seed at site {0,666}: Draws quest modulus 5 result 1. Outputs seed {666,0}, six gold-pile requests afterward; no game/unit step at this site.
  - (f) State 1/scheduled 0, quest seed {0,666}: Draws none; Outputs unchanged seed, no schedule, scheduled 0. This differs from an eligible state-0 request.
  - (g) Supplied inactive altar slice (TSW-R02c), active 0, default clock, applied 1, player flags {14}, object mode 0; preceding preparation/drop helpers complete with no clock/flag/seed writes and requested amulet count 0; quest 9 absent; seed at inline site {0,666}. Draws quest modulus 5 result 1. Outputs quest seed {666,0}, six gold-pile requests, default clock retained, applied 1 and flags {14} retained, no reset/World-clock request; return 0 after supplied remaining operations. No full drop/helper effect is asserted.
  - (h) state 0/scheduled 0, quest seed {0,667}: Draws quest modulus 2 result 1; parameter 16. Outputs seed {667,0}, schedule(timer,16) with due value 1016, then scheduled 1.
  - Discriminates: wrong seed owner, two steps, scheduling before the draw, ignoring nonzero state, or resetting an inactive quest fails; (h) rejects taking parity from the pre-step low word (0, giving 15). Helpers may draw independently.
  - runtime_observed: no.

- **TSW-V19** Drift/recipient filters (TSW-R10a, TSW-R10d; boundary: supplied server act update).
  - Inputs: act 0 default clock except ticks 2175; clients in order A resolved/state 4/act 0, B resolved/state 4/act 1, C resolved/state 3/act 0, D unresolved/state 4/act 0.
  - Seeds: no draw. Draws: none in clock/request slice.
  - Trace: ticks 2176, q=17, distance 17>16.
  - Outputs: cycle 2/period 0/ticks 2176, last resync 17, result 1. Exact log: current-clock read(act 0), then player-notification(game,A-player,0), change-tracking(A-player,1), World-clock(A,2,2176,0), player-notification(game,B-player,0), change-tracking(B-player,1), player-notification(game,C-player,0), change-tracking(C-player,1). No message for B/C/D and no operations for D. This forbids deferring A's clock request until after B/C notification pairs.
  - Fresh Inputs: ticks 2047, otherwise same. Outputs ticks 2048/q16/result 0/last 0, no log.
  - Separate Inputs: one normal frame from game frame 9; only acts 0 and 3 present, each with cycle 2/period 0/ticks 2175/rate 128/index 0/flag 0/last resync 0. Clients A resolved/state 4/act 0 and E resolved/state 4/act 3, in that order; any client source values are irrelevant. Trace: frame becomes 10; act 0 advances to 2176/q17 and act 3 to 2191/q17, each notifying. Outputs: both last-resync fields 17, both cycle/period 2/0; final game frame 10. Exact log: current-clock read(act 0), A notification, A tracking, World-clock(A,2,2176,0), E notification, E tracking; then current-clock read(act 3), A notification, A tracking, E notification, E tracking, World-clock(E,2,2191,0). Here notification/tracking mean the same two operations and arguments as above for that recipient. No other acts are updated.
  - Resolution Inputs: ticks 2175 as first case; clients in order F (direct flag clear, cached reference nonzero, supplied lookup returns F-player), G (direct flag clear, cached nonzero, lookup fails), H (direct flag clear, cached zero), I (direct flag set, cached reference I-player); all state 4/act 0. Outputs: G's cached reference becomes 0; F's, H's and I's are unchanged. Single ordered log after the current-clock read: lookup(F), F notification, F tracking, World-clock(F,2,2176,0), lookup(G), I notification, I tracking, World-clock(I,2,2176,0). No lookup for H or I.
  - Discriminates: before-tick sampling, >=16, filtering before pre-filter requests, batching A's send after B/C, reversed act visitation, or using a client-derived act for the server delta fails. The resolution case rejects looking up a zero cache, retaining a failed cache, looking up a direct-flag client, or resolving all clients before any notification.
  - runtime_observed: no.

- **TSW-V20** Progress exact bits (TSW-R01b; boundary: supplied flag stage).
  - Inputs: found quest 10, flags {}, fresh states 0/1/2/3/255; separate state 3 flags {1}, state 3 flags {0}, and missing-quest cases.
  - Seeds: no draw. Draws: none locally.
  - Trace: >=1 sets bit 2; exact 2/3 selects extra bit 3/4.
  - Outputs: flags {}/{2}/{2,3}/{2,4}/{2}; all return 0. State 3 flags {1} stays {1}; fresh state 3 flags {0} stays {0}; missing quest stays {}; each returns 0 without flag writes.
  - Discriminates: >=3 setting bit 4 fails at 255; no stale bits are cleared.
  - runtime_observed: no.

- **TSW-V21** Retry/timer/animation state (TSW-R01e, TSW-R01f, TSW-R01h, TSW-R01i, TSW-R07e, TSW-R08a; boundary: supplied start slices).
  - Inputs: fresh defaults; overrides below.
  - Seeds: no draw except case (g), whose quest seed is explicit. Draws: none locally except (g).
  - Trace: run one named slice; retain writes preceding failure.
  - (a) Animation with Act II absent: Outputs marker B 1/unit marker 77, state 3/stage 1/transition 0, applied 0/pending 1; log pre-start(P), mode-update(U,0); no start message and no second pre-start request, because stage is then 1.
  - (b) Event 0x28, pending 1, act present: Outputs started clock `(0,3,1200,4,1,1,0)`, pending 0/applied 0/state 1/stage 1; P bit 2 set; log pre-start, presentation, start messages, progress writer(P) [call-time snapshot pending 1/applied 0/state 1/stage 1], then the pending-clear write to 0.
  - (c) Timer scheduled 1/applied 1/pending 1: Outputs scheduled 0/applied 1/pending 1, return 1; no requests/clock writes.
  - (d) Timer scheduled 1/applied 0/pending 0, Act II absent: Outputs scheduled 0/pending 1/state 1/stage 1/transition 0, return 1; pre-start only.
  - (e) Direct retry argument 1/pending 1: Outputs started clock, pending 0/applied 1; quest state 0/stage 0/transition 7 retained; presentation only, no pre-start/progress/message request.
  - (f) Event 0x28/pending 1, Act II absent: Outputs pending 1/applied 0/state 1/stage 1/transition 0, default clocks retained; pre-start(P) only, no progress/message requests. With auxiliary identity 0, TSW-R01i makes no request.
  - (g) Fresh event key 0x2C or 0x2D independently, state 0/active 1/scheduled 0; quest seed {0,666}. Draws: one quest step, modulus 2 result 0, parameter 15. Outputs seed {666,0}, schedule(timer,15), scheduled 1, default clock and state/stage/transition 0/0/7 retained; no shared-start or message request. Event producers and timer delivery remain gated.
  - (h) Timer scheduled 1/applied 0/pending 0, Act II present: Outputs scheduled 0/applied 1/pending 0/state 1/stage 1/transition 0, started clock `(0,3,1200,4,1,1,0)`, P flags {2}, return 1. Log: pre-start(P), intensity, intensity, colour, World-clock(P,5,0,1), Quest-event(P,10,16), progress writer(P) [call-time snapshot applied 1/pending 0/scheduled 0/state 1/stage 1]. Fresh scheduled 1/applied 0/pending 1, Act II present instead: Outputs as (h) except pending 1 retained, with progress-writer snapshot pending 1.
  - (i) Fresh animation start slices with active 0 or 2 independently, state 0: Outputs marker B 1/unit marker 77, default state/stage/transition/applied/pending/clock retained; log mode-update(U,0) only. Fresh active 1/state 2 instead skips the start, then finds stage 0: Outputs state 2, stage 2, transition 0, marker B 1/unit marker 77, applied/pending/clock retained; log pre-start(P), mode-update(U,0). Fresh active 1/state 2/stage 1 with mode argument 2 and the callback unit absent: Outputs unit marker -1, marker B 1, stage/state/transition 1/2/7 retained; log mode-update(absent,2) only. The mode-update's effects are supplied.
  - (j) Fresh event key 0x29, pending 1/state 0/active 1: Outputs pending 1 and default state/stage/transition/applied/clock retained; no draw, schedule, shared-start, progress/message, roster or diagnostic request.
  - (k) Fresh direct-retry cases: argument 0/pending 1/quest present; argument 1/quest absent, with no private-state inputs for that absent quest; argument 1/pending 2/quest present. Outputs in each: default clock retained, no presentation/pre-start/progress/message request. In the two quest-present cases pending remains 1/2 respectively and state/stage/transition/applied stay 0/0/7/0; the absent quest has no private state to read or write.
  - (l) Fresh keys 0x2C/0x2D independently, active 2/state 0/scheduled 0, quest seed {0,666}: Outputs seed {0,666} unchanged, scheduled 0, active 2 and default state/stage/transition/applied/pending/clock retained; no draw/schedule/shared-start/progress/message request from the slice.
  - (m) Event auxiliary identity 0x28, each fresh: (m1) key 0x29, roster word 3, state 2, diagnostic switch 1: Outputs state 3, other quest fields retained; log roster-update(77), diagnostic text(`changing state to 3` plus newline). (m2) key 0x29, roster word 0, state 1: Outputs state 1 retained, no request. (m3) key 0x29, roster word 3, event unit absent, state 0: Outputs state 0 retained; log roster-update(-1) only. (m4) key 0x28, pending 1, state 2, roster word 0, switch 0, Act II present: Outputs started clock `(0,3,1200,4,1,1,0)`, pending 0, applied 0, state 3, stage 1, transition 0, P flags {2,3}; log pre-start(P), intensity, intensity, colour, World-clock(P,5,0,1), Quest-event(P,10,16), progress writer(P) [snapshot state 2/pending 1], pending-clear write, then the state-3 write. With auxiliary identity 0 instead, (m4) ends at state 2 with the same clock. (m5) key 0x29, roster word 0, diagnostic switch 1, state 3 and separately state 4: Outputs state retained, no request.
  - (n) Fresh event key 0x28, pending 2, auxiliary identity 0, Act II present: Outputs pending 2, default state/stage/transition/applied/clock retained; no pre-start, presentation, progress or message request.
  - (o) Fresh animation start slice, active 1, state 0, Act II present: Outputs marker B 1/unit marker 77, started clock `(0,3,1200,4,1,1,0)`, applied 1, pending 0, state 3, stage 1, transition 0. Log: pre-start(P), intensity, intensity, colour, World-clock(P,5,0,1), Quest-event(P,10,16), mode-update(U,0); no second pre-start request, because stage is then 1.
  - (p) Fresh animation start slice, active 1, state 1, Act II present: shared start retains state 1, then the slice writes state 3. Outputs marker B 1/unit marker 77, started clock `(0,3,1200,4,1,1,0)`, applied 1, pending 0, state 3, stage 1, transition 0. Log: pre-start(P), intensity, intensity, colour, World-clock(P,5,0,1), Quest-event(P,10,16), mode-update(U,0).
  - (q) Fresh event key 0x2C, active 1, state 1, pending 1, scheduled 0, quest seed {0,666}, auxiliary identity 0: Outputs pending 1, seed unchanged, scheduled 0, default clock, state/stage/transition 1/0/7 retained; no draw, schedule, start, progress or message request.
  - Discriminates: (p) rejects an animation start gated on state 0 only; (q) rejects treating key 0x2C as a pending retry or scheduling at nonzero state; (b) does not set applied and progress sees pending 1 before it is cleared; (c) clears scheduled before early return; (e) bypasses shared start writes/messages; (f) retains pending on failure; (g) schedules rather than applying immediately; (h) progress sees applied 1, and its pending-1 variant rejects clearing pending on timer success; (i)/(j)/(k)/(l) reject bypassing the exact eligibility guards, including treating schedule-active as nonzero; (a)/(i) reject skipping the always-made marker writes and mode-update, or the stage-0 branch; (m) rejects gating the auxiliary slice by event key, running it before the start stage, an else-chain, or a state>=2 test (m5); (n) rejects a nonzero-pending test for key 0x28; (o) rejects writing state 3/applied only on failure.
  - runtime_observed: no.

- **TSW-V22** Signed overflow/wrap resync (TSW-R05, TSW-R10d; boundary: fresh core/server updates).
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

- **TSW-V23** Client mapping/period return (TSW-R06b; boundary: supplied-source wrapper).
  - Inputs: fresh default clocks and sources -1,39,40,74,75,102,103,108,109,120,1023,1024; presentation supplied.
  - Seeds: no draw. Draws: none.
  - Trace: signed entries 1..5 mapping then core tick.
  - Outputs: acts 0,0,1,1,2,2,3,3,4,4,4,0; ticks 1 except sources 103/108 give 16; cycle 2/period 0/return 0 throughout. Source 120 final colour (245,240,255).
  - Ordered presentation requests for every listed source case: first intensity with that exact supplied source, then colour with its derived act from the output array. At entry to both requests the clock snapshot is cycle 2/period 0/flag 0/rate 128/index 0 and ticks 1, except sources 103/108 whose ticks are 16. Thus source 120 requests intensity(source 120) before colour(act 4), with ticks 1 at both entries; only after colour returns are the three colour bytes overwritten to (245,240,255). The requests carry these established semantic inputs; the snapshots are observations, not extra arguments.
  - Fresh Inputs: source 0/cycle 0/period 3/ticks 43520/rate 128/flag 0. Trace: 43521>row 1 threshold 43520. Outputs cycle 1/period 3/ticks 43520/return 0. Ordered requests: intensity(source 0), colour(act 0); both entry snapshots are cycle 1/period 3/ticks 43520/flag 0/rate 128/index 0.
  - Separate Inputs: source 0/cycle 0/period 3/ticks 1200/rate 4/index 1/flag 1. Trace: advance 1201, transition to eclipse row 1, period changes from 3 to 0. Outputs cycle 1/period 0/ticks 0/return 1; ordered intensity(source 0), colour(act 0), both with entry snapshots cycle 1/period 0/ticks 0/flag 1/rate 4/index 1.
  - Discriminates: environment-derived act/cycle-change return, always returning 0, swapped source/act request inputs, omitted requests or colour-before-intensity fails. Negative source maps to act 0 under either signed or unsigned comparison, so this vector does not distinguish them; the results are identical.
  - runtime_observed: no.

- **TSW-V24** Intensity calls (TSW-R07c; boundary: isolated supplied intensity branch).
  - Inputs: act 1/flag 1, fresh intensity 200/39/33/32/20; separate act 1/flag 0/source 120/intensity 20.
  - Seeds: no draw. Draws: none.
  - Trace: subtract 8 only above 32, then floor 32; normal source 120 assigns 200.
  - Outputs: 192/32/32/32/32; separate case 200. Separate act 1/flag 1/source 120/intensity 200: Outputs 192 (the flag-nonzero branch precedes the source-120 assignment). Separate act 1/stored flag 2/intensity 200: Outputs 192.
  - Discriminates: floor-before-subtract alone yields 31/25 at 39/33; period does not affect these branches.
  - runtime_observed: no.

- **TSW-V25** Client message/rate (TSW-R10c; boundary: supplied eligible handler/clock stage).
  - Inputs: default client clock/source 0, eligible player/act, World-clock(5,0,1); presentation supplied.
  - Seeds: no draw. Draws: none.
  - Trace: initially assign eclipse row 5 period 2/ticks 0; first intensity uses old flag 0; write flag 1; recompute 128*240=30720.
  - Outputs: clock `(5,2,30720,128,0,1,0)`; first intensity [call-time snapshot flag 0/cycle 5/period 2/ticks 0/rate 128/index 0], second intensity [snapshot flag 1/cycle 5/period 2/ticks 30720/rate 128/index 0], then colour; later refresh supplied.
  - Fresh Inputs: default clock, message (2,46080,0). Outputs cycle 2/period 0/ticks 46080/rate 128/flag 0; one intensity request, no recompute/colour. Message (2,46081,0) instead yields ticks 0. An ineligible player/act instead yields unchanged clock, no assignment requests. With eligibility true, message (6,0,0) or (2,-1,0) takes the fatal assertion path before any clock write or request. Source 120 with default clock and message (2,46080,0): Outputs cycle 2/period 0/ticks 46080/flag 0, one intensity request (source 120, act 4), no colour request, and final colour bytes (245,240,255).
  - Fresh Inputs: client clock `(5,2,30720,128,0,1,0)`, source 0, message World-clock(2,100,0). Trace: assign cycle 2, ticks 100, normal row 2 period 0; first intensity observes old flag 1; store flag 0; no recompute. Outputs clock `(2,0,100,128,0,0,0)`; one intensity(source 0, act 0) [snapshot flag 1/cycle 2/period 0/ticks 100], no colour request. Rejects recomputing whenever the stored flag is nonzero.
  - Fresh Inputs: default client clock, source 0, message World-clock(1,0,1). Trace: assign eclipse row 1 period 0 (normal row 1 would give 3), ticks 0; first intensity observes flag 0; write flag 1; recompute eclipse row 1, ticks 128*0=0. Outputs clock `(1,0,0,128,0,1,0)`; first intensity [snapshot flag 0/cycle 1/period 0/ticks 0], second intensity [snapshot flag 1/period 0/ticks 0], colour(act 0). Rejects choosing the period table by the old stored flag.
  - Fresh Inputs: client clock `(2,0,0,4,1,0,17)`, source 0, message World-clock(5,0,1). Trace: 0 does not exceed 4*360; assign eclipse row 5 period 2/ticks 0; first intensity observes flag 0; write flag 1; recompute 4*240=960. Outputs clock `(5,2,960,4,1,1,17)`; ordered intensity(source 0, act 0) [snapshot flag 0/cycle 5/period 2/ticks 0/rate 4/index 1], intensity(source 0, act 0) [snapshot flag 1/ticks 960], colour(act 0).
  - Discriminates: copying flag-1 packet ticks yields 0; copying server rate 4 yields 960 in the first case, while resetting a retained rate 4/index 1/last 17 to defaults fails the rate-4 case; >= pre-clamp fails at 46080; storing message flag before first intensity fails its old-flag snapshot.
  - runtime_observed: no.

- **TSW-V26** Altar ordered requests and timer callback (TSW-R02b, TSW-R02c, TSW-R02d, TSW-R07e, TSW-R12b; boundary: supplied altar operations).
  - Inputs: player flags {14}, slot-10 bit 0 clear, object O identity 55 with mode marker 0; object-source(O) returns s; the participant iteration sets the amulet count to 2; quest seed {0,666} at the draw site; other supplied operations return normally without other writes or draws.
  - Seeds: quest seed {0,666}. Draws: one quest step, modulus 5 result 1.
  - Trace: 666 mod 5 = 1, so 5+1 = 6 gold-pile drops.
  - (a) Active 1, applied 1, clock `(0,3,1200,4,1,1,0)`, diagnostic switch 1, quest 9 present with counter 4 (it stays 4), source-derived(s) returns d; first amulet drop returns item I1, second returns none. Outputs: source result s, derived result d, mode argument 2, marker B 1, unit marker 55, marker A 1, amulet count 2, applied 0, clock `(2,0,0,128,0,0,0)`, flags {1,13}, quest seed {666,0}, return 0. Log: test slot-11 bit 1 (false), test slot-11 bit 0 (false), object-source(O), source-derived(s), mode-update(O,1), intensity, World-clock(P,2,0,0), diagnostic text(`Finished quest`), quest-notification(4), game-notification(6), participant iteration(P), test slot-11 bit 0 (false), set 1, set 13, clear 14, post-flag iterations A(O), B(O), C(O), quest-record update(13), source-derived(s), O item-code field := `vip `, amulet drop(O) [call-time snapshot quest seed {0,666}], item-flag mark(I1,16,1), amulet drop(O), amulet-result(1), context request(4), six gold-pile drops(O) [first one's snapshot quest seed {666,0}], completion iteration, timer request(altar callback, 10) with due value 1010. With diagnostic switch 0 instead: the same outputs and log without the diagnostic text.
  - (b) Active 0, applied 1, default clock; the first amulet drop returns item I1, the second returns none; quest 9 present with counter 4. Outputs: marker A 1, mode argument 2, amulet count 2, marker B/unit marker/source result/derived result 0, quest 9 counter 5, clock/applied/flags unchanged, quest seed {666,0}, return 0. Log: test slot-11 bit 1 (false), test slot-11 bit 0 (false), mode-update(O,1), participant iteration(P), object-source(O), source-derived(s), O item-code field := `vip `, amulet drop(O), item-flag mark(I1,16,1), amulet drop(O), context request(4), six gold-pile drops(O). No reset, diagnostic, notification, flag, amulet-result, completion-iteration or timer request.
  - (c) As (b) but quest 9 absent: Outputs the same, with no counter write; return 0.
  - (d) Altar timer callback, state 4, stage 2, marker C 1: Outputs transition 0, stage 3, marker C 0, return 1; log pre-start(P). Separate state 3, stage 2, marker C 1: Outputs stage 2 retained, marker C 0, return 1, no request. Separate state 5, stage 2, marker C 1: Outputs stage 2/transition 7 retained, marker C 0, return 1, no request.
  - (e) As (a), except the supplied mode-update(O,1) sets active to 0. Outputs: active 0, applied 1, clock `(0,3,1200,4,1,1,0)` retained, flags {14}, source result s, derived result d, mode argument 2, marker B 1, unit marker 55, marker A 1, amulet count 2, quest seed {666,0}, return 0. Log: test slot-11 bit 1 (false), test slot-11 bit 0 (false), object-source(O), source-derived(s), mode-update(O,1), participant iteration(P), source-derived(s), O item-code field := `vip `, amulet drop(O), item-flag mark(I1,16,1), amulet drop(O), amulet-result(1), context request(4), six gold-pile drops(O), completion iteration, timer request(altar callback, 10). No reset, diagnostic, notification, flag test/write, post-flag iteration or quest-record update.
  - (f) As (a), except the supplied participant iteration sets active to 0 as well as the amulet count 2. Outputs: as (a) except active 0 and flags {14}. Log: as (a) with the entries from the second test of slot-11 bit 0 through quest-record update(13) removed.
  - (g) Active 1, applied 1, clock `(0,3,1200,4,1,1,0)`, diagnostic switch 0, player flags {14}; the participant iteration sets amulet count 2; the first amulet drop returns item I1 and also sets the amulet count to 1. Outputs: one amulet-drop attempt only, item-flag mark(I1,16,1), amulet-result(1), amulet count 1; otherwise the outputs of (a) with no diagnostic text.
  - Discriminates: (g) rejects capturing the amulet count once before the loop; (a) rejects adding successes to quest 9 on the active branch, omitting the second active test's flag/iteration block, counting attempted rather than successful drops, or scheduling before the gold drops; (b) rejects resetting or scheduling on the inactive branch, losing the quest-9 addition, or adding attempts (6) instead of successes (5); (a) with switch 0 rejects an unconditional diagnostic; (d) rejects a state-independent stage write or a state>=4 test; (e)/(f) reject dropping either active re-test.
  - runtime_observed: no.

### 8.2 Observation vectors

Pending [companion procedures](../research/spec-evidence/quest-tainted-sun-world-clock.md#observation-procedures) specify setup/capture limits.

| Id | Measure | Expected / comparison |
|---|---|---|
| **TSW-O01** | quest start/altar with another act as control | server returned clock, message payload, client assignment, reset tuple; appearance/trigger identity remain gated |
| **TSW-O02** | fresh joins before/during/after progress, repeated difficulty | flag selection, retained failure state, current snapshot; save encoding remains gated |
| **TSW-O03** | second player join and either player's altar operation | shared Act II clock, own-act snapshot, exact message order/delivery |
| **TSW-O04** | leave/return, server frames, same-act recipients | continued Act II advance, other-act preservation, separate drift/transition sends |

## 9. Acceptance and dependencies

Partial acceptance reproduces supplied-input assertions of TSW-V01 through TSW-V26 for all static leaves in the Id map. Arithmetic, writes, returns, direct draw sites and request order are available independently of event/save producers and downstream delivery/presentation. V08 excludes drops/later quest work and V26 supplies drop, iteration and callback results; V18 excludes helper draws; V17 is an exact static scenario. Passing these does not establish complete quest, packet, save, visual or RNG parity.

Full reproduction also requires gated assertions of TSW-R01c, TSW-R07d, TSW-R08b, TSW-R09b, TSW-R10b and TSW-R11; independently established integrations TSW-Q08/TSW-Q10; and applicable matching TSW-O01 through TSW-O04 and TSW-Q04/TSW-Q09 observations. Status/partial acceptance imply no user approval.

**Depends on:** NRF-R01/NRF-R02 in [numeric/RNG foundations](core-numeric-and-rng-foundations.md) for quest-seed arithmetic/reduction and its i32/low32 notation. Section 3's supplied integration operations remain TSW-Q10. GDF-R19a in [game-data fields](core-game-data-fields.md) supplies only the distinction between code-built records and loaded data, without quest flag meanings.

**Depended on by:** no known non-gap plan or executable consumer cites this contract. Potential consumers need their own contracts, TSW-Q06.

## 10. Related specs

- [Numeric/RNG foundations](core-numeric-and-rng-foundations.md): shared arithmetic and frame conventions.
- [Game-data fields](core-game-data-fields.md): names and code-built versus table-record distinction.

## Appendix A. Provenance

Evidence targets original Windows x86 Lord of Destruction Expansion 1.13c D2Common, D2Game, D2Client and Fog, identified in the [source manifest](../research/source-manifest.md). The [public companion](../research/spec-evidence/quest-tainted-sun-world-clock.md) records RVAs/ordinals, dated readings, correction rationale, partial observations and limits. Static readings dated 2026-10-12 retain their scope; targeted matched-reconstruction, instruction and original-image checks are dated 2026-10-15.

Read-only idle observations from 2026-10-12 corroborate code constants, selected fresh-clock fields and normal Act I day progress. They do not reproduce complete exact-frame, quest, transition or notification vectors. All executable vectors remain runtime_observed: no. No new gameplay observation is recorded.
