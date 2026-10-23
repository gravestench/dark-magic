# Tainted Sun quest

**Description**

The Tainted Sun quest puts Act II's outdoor world clock into an eclipse, and breaking the viper altar returns it to normal. The quest decides when the eclipse starts, whether it is applied again when a player joins a game with quest progress, and what operating the altar records for the players. Players whose current act is Act II are told when the eclipse starts or ends; the clock itself is described in its own specification.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: TSW-R, TSW-Q, TSW-V, TSW-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/quest-tainted-sun.md)
Open evidence gaps: [gap file](../plans/gaps/quest-tainted-sun.md)

## 1. Scope

This contract covers the quest-start, altar, retry and join stages that start or reset Act II's clock, the quest's own start/reset/failed-join messages and diagnostic text, and their bounded RNG use. The clock assignments these stages request (WCK-R01b start, WCK-R01c reset), clock arithmetic, server frame updates, the join snapshot and client clock handling are owned by [world clock](drlg-world-clock.md). Quest flags, supplied callback results, timer delivery, player resolution and message delivery are explicit integration inputs. Complete quest progression, save encoding, general lighting, audio playback and other game builds are outside the reproduced boundary. Their unresolved connections are named below.

## 2. Trigger and ownership

D2Game owns the game-level quest. Quest initialization installs join and event callbacks; join progress, an animation callback, supplied event keys 0x28/0x2C/0x2D and a timer callback reach the start stage. Event identities remain TSW-Q01. Object-operation dispatch reaches the altar stage. Each act's clock, and the server frame that advances it, belong to [world clock](drlg-world-clock.md) (WCK-R06a); this quest only requests Act II's start and reset assignments.

Start/reset recipients are filtered by the recipient client's act. Transmission and client sound consumers remain supplied integrations, TSW-Q06. Logical message names impose no wire encoding.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| act clocks | game act collection | five optional acts, each with optional environment | Act II's presence is tested; clock fields and tables are defined in [world clock](drlg-world-clock.md#3-inputs-and-data) |
| quest | game quest 10, slot 11 | active/state/stage/transition markers u8; scheduled/applied/pending u8; neutral markers A/B/C u8; mode argument, unit marker, signed amulet count, source result and derived result i32; roster word u16 | Neutral marker writes are specified below; initialization zeroes the private markers |
| shared quest flags | game-level shared quest-flag record, distinct from any player's flags | slot-11 bit 13 | Written only by TSW-R01d; a player's own bit 13 is written only by TSW-R02b |
| quest 9 counter | game quest 9's private state, when present | i32, low32 addition | Written only by TSW-R02c |
| callback unit | animation/event callback input, optional | unit identity i32 | Absent unit records identity -1 |
| diagnostic switch | server-wide value, not per game | i32; nonzero enables | Producer remains TSW-Q06 |
| progress flags | the player's current-difficulty slots 11 and 10 | boolean bits 0,1,2,3,4,13,14 | Broader meanings remain TSW-Q02 |
| recipients | ordered player iteration | player or absent; its client or absent; client act u8; client state i32 | Distinguishes a send request from completed delivery |
| retry act argument | supplied by the direct retry's outer dispatch | i32 | Used only by TSW-R05a; its producer remains TSW-Q06 |
| event | supplied callback input | event key i32; auxiliary selector u16; auxiliary identity i32; optional unit | Gameplay meanings are not inferred from values |
| altar | player, optional object, item-presence results | object identity i32, object mode marker i32, object item-code field (four characters); flag/item tests boolean | Item search region is supplied, TSW-Q06 |
| timer | scheduler and delivery | schedule parameter, game quest counter i32; callback identity | Request and actual timed callback execution are distinct |
| dependency operations | iteration, notifications, mode/drop callbacks, presentation, send | ordered requests; supplied normal completion/results | Hidden effects/RNG remain TSW-Q05/TSW-Q06 |
| RNG | quest, game and unit seeds | each `{lo,hi}` u32 | Independent owners; the quest seed is the game-level quest structure's seed, shared by TSW-R01g and the altar site and distinct from quest 10's private state; NRF-R01/NRF-R02 own arithmetic |

Quest 10, progress slots 11/10, four-character item codes `vip ` and `hst `, and Quest-event flag values 16/1 are fixed code selections; they originate in **engine code** and are **fixed by original code**. Item names are supporting descriptions, not new table mappings. The World-clock payloads named in TSW-R04a are fixed code values; their clock meaning is defined in [world clock](drlg-world-clock.md).

Due values and the quest-9 addition retain low32 as stated in their rules. Required act/environment and flag records exist unless a rule handles absence; assertion failures are not normal no-op results.

## 4. Rules (execution order)

### TSW-R01. Quest start stages

**TSW-R01a. Shared start** [verified in 1.13c code (static)]. Write transition marker 0 and stage marker 1; request the pre-start callback over supplied players. After normal completion, change state 0 to 1 and apply TSW-R04c with text `changing state to 1` plus a newline, retaining other states without that request. If Act II is absent, set pending 1 and return 0: retain earlier quest writes, leave clocks unchanged, request no start messages. Otherwise apply WCK-R01b, request TSW-R04a over players and return 1. This stage does not set applied or clear pending. Dependency-produced effects are TSW-Q06.

**TSW-R01b. Progress writer** [verified in 1.13c code (static)]. Per player, return 0 without writes if slot-11 bit 0 or 1 is set, or quest 10 is absent. Otherwise, unsigned state >=1 sets bit 2; additionally state 2 sets bit 3, or state 3 sets bit 4. Clear no other bits; return 0. Supplied flag operations persist in order.

**TSW-R01c. Upstream identity** [unresolved]. Supplied keys 0x2C/0x2D reach the schedule condition. A related callback compares an auxiliary selector with 0x15C and signed range 0x16A..0x174; one branch also tests identity 0xB1. Their meanings, including whether Lost City entry produces a relevant event, remain TSW-Q01. These comparisons establish no universal absence of a level-entry trigger.

**TSW-R01d. Join selection** [verified in 1.13c code (static)]. Test the joining player's slot-11 bit 0, then bit 1. Either set: set slot-11 bit 13 in the game-level shared quest-flag record (the joining player's flags are unchanged) and return without applying. If bit 2 is clear, return without writes. Otherwise run TSW-R01a. Success sets applied 1 and leaves pending unchanged; failure sets pending 1 and requests TSW-R04b. After either result overwrite `(stage,state)` with `(2,3)` if bit 4 is set, else `(2,2)` if bit 3 is set, else `(1,1)`. Failure does not undo this overwrite. Join selection requests no progress writer.

**TSW-R01e. Animation start slice** [verified in 1.13c code (static)]. First, on every call, write marker B=1 and unit marker=callback unit identity (-1 when absent). Active is tested once, at this point. If active is not exactly 1, skip to the final request. Otherwise, for unsigned state <=1, run TSW-R01a; set applied 1 on success or pending 1 on failure, then write state 3 and transition 0 in both cases. Then, still within active exactly 1 and for any state, if stage is 0, write transition 0 and stage 2 and request the pre-start callback over supplied players. Shared start writes stage 1, so this branch follows only a skipped start. Finally, on every path, request mode-update(callback unit, current mode argument value). The mode-update's effects are TSW-Q06.

**TSW-R01f. Event retry slice** [verified in 1.13c code (static)]. These tests run in order as independent checks, not an else-chain. Key 0x2C/0x2D, state 0, active exactly 1: invoke TSW-R01g. Event key 0x28 and pending exactly 1: run TSW-R01a; on success run TSW-R01b over supplied players then clear pending. Failure retains pending 1 and earlier writes. TSW-R01i then runs on every call. Supplied operation effects are TSW-Q06.

**TSW-R01g. Schedule** [verified in 1.13c code (static)]. Nonzero state or nonzero scheduled means no draw/request. Otherwise a missing game quest structure takes the fatal assertion path; with it present, step the game quest structure's seed once through NRF-R01, select parameter `15+(new lo & 1)` (15 or 16), request scheduling TSW-R01h, then write scheduled 1. The scheduler stores a due value equal to the supplied game quest counter plus parameter with low32 arithmetic. Elapsed gameplay units/delivery remain TSW-Q06.

**TSW-R01h. Timer callback** [verified in 1.13c code (static)]. Clear scheduled first. If applied is nonzero, return 1 without applying. Otherwise run TSW-R01a. Failure sets pending 1 and returns 1; success sets applied 1, runs TSW-R01b over supplied players and returns 1. Success does not itself clear pending.

**TSW-R01i. Event auxiliary-identity slice** [verified in 1.13c code (static)]. Only when the event's auxiliary identity equals 0x28: if the roster word is nonzero, request roster-update(event unit identity, -1 when absent). Then, if state is exactly 2, write state 3 and apply TSW-R04c with text `changing state to 3` plus a newline. This test reads state after TSW-R01f's start stage. The event key does not gate this slice. Roster-update effects are TSW-Q06.

### TSW-R02. Altar stage

**TSW-R02a. Eligibility** [verified in 1.13c code (static)]. O denotes the operated object, which may be absent. A missing quest 10 returns 0 with no other work. With quest 10 found, test player slot-11 bit 1, then bit 0 if needed. Only if either is set, test slot-10 bit 0, then supplied `vip ` presence, then `hst ` presence, stopping at the first true result. True: request rejection with value 19 and player; return 0 without reset. If both primary bits are clear, skip all three secondary tests. Next, an existing object with nonzero mode marker returns 0 without reset; a missing object does not take that return. Otherwise TSW-R02b runs when active is nonzero, else TSW-R02c. All returns are 0.

**TSW-R02b. Active branch** [verified in 1.13c code (static)]. Requests and writes, in order:
1. Request object-source(O), store result as source result; request source-derived(source result), store result as derived result. Request mode-update(O,1). Write mode argument=2, marker B=1, unit marker=object identity (-1 when absent), marker A=1.
2. If active is still nonzero: run quest reset. With Act II present, clear applied, run WCK-R01c and request reset messages. With Act II absent, retain applied/clock and request no reset message. Then, still inside this active test, apply TSW-R04c with text `Finished quest`, and request quest-notification(4), then game-notification(6).
3. Write amulet count 0; request participant iteration(player). The supplied iteration sets the amulet count.
4. If active is still nonzero: if player slot-11 bit 0 is clear, set slot-11 bit 1, set slot-11 bit 13, then **clear** slot-11 bit 14. Then request post-flag iterations A, B and C, each with O. Then request quest-record update(13).
5. Request source-derived(source result) again. Write O's item-code field to `vip `. Request amulet drop(O) amulet-count times, with an attempt index starting at 0 and increasing by 1 after each attempt, comparing the signed attempt index with the amulet count re-read before each attempt; a count <=0 requests none. Each nonzero drop result requests item-flag mark(item, 16, 1) and counts one success. Then request amulet-result(successes) and context request(4).
6. Draw once from the quest seed (TSW-R07b) and request `5+(new lo mod 5)` gold-pile drops(O). A missing game quest structure takes the fatal assertion path instead.
7. Request completion iteration, then a timer request with parameter 10 for TSW-R02d, whose due value is the supplied game quest counter plus 10 (low32), as in TSW-R01g. Return 0.

**TSW-R02c. Inactive branch** [verified in 1.13c code (static)]. Request mode-update(O,1); write marker A=1, mode argument=2 and amulet count 0; request participant iteration(player). Request object-source(O) and source-derived(result), storing neither. Write O's item-code field, run the counted drop loop and context request(4) as in TSW-R02b step 5, without amulet-result. Run TSW-R02b step 6. Then find quest 9: if present, add the success count to its counter (low32); return 0 either way. No environment reset, flag write, timer, marker B/unit marker write or completion iteration occurs.

**TSW-R02d. Altar timer callback** [verified in 1.13c code (static)]. If state is exactly 4, write transition 0 and stage 3 and request the pre-start callback over supplied players. Then, in every case, write marker C=0 and return 1. Timer delivery and elapsed units remain TSW-Q06.

Supplied operation effects, including actual drops, iteration results and amulet counts, remain TSW-Q05/TSW-Q06.

### TSW-R03. Ownership

**TSW-R03** [verified in 1.13c code (static)]. Quest state is game-level; each act owns its environment. Start/reset address only Act II and request messages only for players whose client act equals 1. Any eligible player's altar operation can reset that shared clock. Leaving an act does not reset it; the server continues updating existing acts. Join snapshots select the joiner's own act, WCK-R07.

### TSW-R04. Quest messages

**TSW-R04a. Start/reset requests** [verified in 1.13c code (static)]. Per Act II player, start requests World-clock `(5,0,1)` then Quest-event `(quest 10,flags 16)`. Server start is cycle 0/ticks 1200, distinct from the message. Reset requests only World-clock `(2,0,0)`. Other acts get neither request. Both per-player callbacks return 0. The act is the recipient client's act; client state is not tested. Before the act test, each callback requires a present player unit with a client; an absent unit, non-player unit or player without a client takes the fatal assertion path rather than being skipped. Delivery/sound remain TSW-Q06.

**TSW-R04b. Missing-act join** [verified in 1.13c code (static)]. Failed join apply requests Quest-event `(quest 10,flags 1)` to the joining player's client without act filtering; no start World-clock. The client is taken from player data only when the joining unit is present and is a player; otherwise the request carries no client. The send operation returns without queuing anything when it has no client. Earlier quest/pending writes remain in every case.

**TSW-R04c. Diagnostic broadcast** [verified in 1.13c code (static)]. At the sites named in TSW-R01a, TSW-R01i and TSW-R02b, if the diagnostic switch is nonzero, request one game-wide chat-style text message carrying the stated text; "plus a newline" means one line-feed character (0x0A) and no carriage return. A zero switch requests nothing; text formatting alone has no visible effect. Recipient iteration and delivery are TSW-Q06.

### TSW-R05. Retry and repetition

**TSW-R05a. Direct retry** [verified in 1.13c code (static)]. Supplied act argument 1, found quest 10, pending exactly 1: apply WCK-R01b directly, clear pending, set applied 1; request no start messages. Other arguments/missing quest/pending!=1 do nothing. This stage does not check Act II for absence; valid input requires its act/environment. Outer dispatch/absence behavior remain TSW-Q06.

**TSW-R05b. Repetition/difficulty** [unresolved]. Selection consumes supplied current-difficulty flags. Saved/repeated progress production remains TSW-Q04.

### TSW-R06. Persistence

**TSW-R06. Persistence** [unresolved]. Read-stage save absence does not prove absence from every save path. Quest-bit encoding/load, repetition and wider clock persistence remain TSW-Q04.

### TSW-R07. RNG

**TSW-R07a. Direct no-draw slice** [verified in 1.13c code (static)]. Eight audited bodies—direct retry, reset/start send callbacks, quest reset, shared start, join apply, join selection, quest initialization—contain no inline RNG step/direct helper call. Their own bodies leave game/unit/quest seeds unchanged. Called callbacks, allocation, presentation, delivery and drops are excluded; this is not a whole-scenario zero-draw guarantee.

**TSW-R07b. Quest-owned sites** [verified in 1.13c code (static)]. Schedule makes one quest-seed draw before requesting the timer. A reached altar inline site steps that same quest seed once after amulet requests, then requests `5+(new lo mod 5)` gold-pile drops, 5..9. No direct game/unit draw occurs at these sites. Intervening/helper RNG is TSW-Q05; the seed at the altar draw site is supplied explicitly.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| TSW-R01a | shared start | verified in 1.13c code (static) | TSW-Q06: callback/delivery effects only | TSW-V01, TSW-V04 |
| TSW-R01b | progress writer | verified in 1.13c code (static) | none | TSW-V08 |
| TSW-R01c | upstream identity | unresolved | TSW-Q01: gameplay event production | none |
| TSW-R01d | join selection | verified in 1.13c code (static) | TSW-Q04/TSW-Q06: saved flags/helper integration | TSW-V03, TSW-V04, WCK-V10 |
| TSW-R01e | animation slice | verified in 1.13c code (static) | TSW-Q06: mode-update/iteration effects | TSW-V09 |
| TSW-R01f | event retry slice | verified in 1.13c code (static) | TSW-Q01/TSW-Q06: event production/other effects | TSW-V09 |
| TSW-R01g | schedule | verified in 1.13c code (static) | TSW-Q06: actual timer/elapsed units | TSW-V07 |
| TSW-R01h | timer callback | verified in 1.13c code (static) | TSW-Q06: delivery/other effects | TSW-V09 |
| TSW-R01i | event auxiliary-identity slice | verified in 1.13c code (static) | TSW-Q01/TSW-Q06: event production/roster-update effects | TSW-V09 |
| TSW-R02a | altar eligibility | verified in 1.13c code (static) | none for supplied item tests | TSW-V05, TSW-V10 |
| TSW-R02b | altar active branch | verified in 1.13c code (static) | TSW-Q05/TSW-Q06: drops/callback effects | TSW-V02, TSW-V05, TSW-V10 |
| TSW-R02c | altar inactive branch | verified in 1.13c code (static) | TSW-Q05/TSW-Q06: drops/callback effects | TSW-V07, TSW-V10 |
| TSW-R02d | altar timer callback | verified in 1.13c code (static) | TSW-Q06: timer delivery | TSW-V10 |
| TSW-R03 | ownership | verified in 1.13c code (static) | none for supplied recipients | TSW-V06, WCK-V10 |
| TSW-R04a | messages | verified in 1.13c code (static) | TSW-Q06: delivery/sound; TSW-Q07: observed client effect of the payload/server divergence only | TSW-V01, TSW-V02, TSW-V06 |
| TSW-R04b | pending request | verified in 1.13c code (static) | TSW-Q06: client resolution/delivery | TSW-V04 |
| TSW-R04c | diagnostic broadcast | verified in 1.13c code (static) | TSW-Q06: switch producer/delivery | TSW-V01, TSW-V09, TSW-V10 |
| TSW-R05a | direct retry | verified in 1.13c code (static) | TSW-Q06: dispatch/absence | TSW-V09 |
| TSW-R05b | repetition/difficulty | unresolved | TSW-Q04: progress production | none |
| TSW-R06 | persistence | unresolved | TSW-Q04: saving/loading | none |
| TSW-R07a | eight direct bodies | verified in 1.13c code (static) | TSW-Q05: reachable draws excluded | TSW-V07 |
| TSW-R07b | quest draw sites | verified in 1.13c code (static) | TSW-Q05: helper draws excluded | TSW-V07 |

## 5. Outputs and side effects

| Effect | Rule |
|---|---|
| shared-start return 0/1, progress-writer return 0, altar return 0, altar-timer return 1 | TSW-R01a, TSW-R01b, TSW-R02a through TSW-R02d |
| quest state/stage/transition writes and retained failure state | TSW-R01a, TSW-R01d, TSW-R01e, TSW-R01i, TSW-R02d |
| markers A/B/C, mode argument, unit marker, source result, derived result, amulet count, quest 9 counter, object item-code field | TSW-R01e, TSW-R02b through TSW-R02d |
| per-player bits 2/3/4 and shared bit 13 | TSW-R01b, TSW-R01d |
| scheduled/applied/pending writes, timer requests and callback returns | TSW-R01d through TSW-R01h, TSW-R02b, TSW-R02d, TSW-R05a |
| player flag-test and item-presence-test requests, rejection, mode-update, roster-update, object-source/source-derived, participant/post-flag/completion iterations, quest/game notifications, quest-record update, amulet/gold drop, item-flag mark, amulet-result and context requests; player bits 1/13 set and 14 clear | TSW-R01e, TSW-R01i, TSW-R02a through TSW-R02c |
| pre-start iteration and progress-writer iteration requests | TSW-R01a, TSW-R01e, TSW-R01f, TSW-R01h, TSW-R02d |
| schedule requests and their due values | TSW-R01g, TSW-R02b |
| Act II start and reset assignment requests (their clock writes are WCK-R01b and WCK-R01c) | TSW-R01a, TSW-R02b, TSW-R05a |
| start/reset World-clock, Quest-event and diagnostic text requests | TSW-R04a, TSW-R04b, TSW-R04c |
| fatal assertion path for invalid sender recipients or a missing game quest structure | TSW-R01g, TSW-R02b, TSW-R02c, TSW-R04a |
| quest-seed writes, schedule parameter, gold-pile count | TSW-R01g, TSW-R02b, TSW-R02c, TSW-R07b |

The quest stages make no environment save write; wider persistence remains gated (TSW-R06). No return value is specified for join selection, the animation and event slices, or direct retry; their callers and any use of those results remain TSW-Q06. Message payloads belong to the rules; this table names effects only.

## 6. Branches and edge cases

Start active checks use equality to 1, while altar active uses nonzero. The animation slice writes marker B/unit marker and requests mode-update even when active is not 1. The event auxiliary-identity slice is independent of the event key, so one event can run a start and then change state 2 to 3.

Missing Act II during shared start retains quest writes/pending; missing Act II during reset retains applied. Per-player senders take the fatal assertion path on an invalid recipient, whereas the failed-join request silently queues nothing without a client. A later failed delivery does not erase earlier local writes. Classic mode and other builds have no established equivalence here.

## 7. Open questions

The [gap file](../plans/gaps/quest-tainted-sun.md#open-evidence-gaps) distinguishes behavioral, integration, observation and outside-boundary questions.

- **TSW-Q01** [unresolved]: establish event/auxiliary-selector producers and Lost City relationship with root-to-callback evidence. Supplied event selection remains available.
- **TSW-Q02** [unresolved]: establish gameplay meanings of neutral flags/markers from anchored readers/writers. Exact bit operations are available.
- **TSW-Q03** [unresolved]: observe the encoded start/reset World-clock, Quest-event and failed-join messages and the client's Quest-event handling. Static wire records are in the companion; no packet capture is available. Client handling of World-clock is WCK-Q03.
- **TSW-Q04** [unresolved]: establish quest persistence/repeated difficulty production and any other clock-saving path. Bounded join reading proves no universal absence.
- **TSW-Q05** [unresolved]: complete reachable RNG budgets for drop helpers, callbacks, other handlers, presentation and delivery. Direct-body/site budgets remain available.
- **TSW-Q06** [unresolved]: independently establish timer allocation/delivery, the remaining quest/altar/mode/roster/drop/iteration callbacks, the item search region, the direct retry's outer dispatch, the diagnostic switch's producer and the transmission of the quest's messages. Local requests/results are explicit; complete integrated effects remain gated.
- **TSW-Q07** [unresolved]: the start assignment (WCK-R01b) leaves the server at cycle 0 with eclipse flag 1, yet the start message (TSW-R04a) sends the fixed cycle 5. Is this divergence intended, and which later resync (WCK-R06b) or client step realigns the client? Both values are statically exact; only observation (TSW-O01) or a rooted client consumer can settle the effect.

## 8. Test vectors

### 8.1 Executable vectors

Each case is fresh unless stated. Default clock tuple `(cycle,period,ticks,rate,index,flag,last resync)` is `(2,0,0,128,0,0,0)`. Default quest `(active,state,stage,transition,scheduled,applied,pending)` is `(1,0,0,7,0,0,0)`; markers A/B/C, mode argument, unit marker, source result, derived result, amulet count and roster word are 0, the diagnostic switch is 0, event auxiliary identity is 0, game quest counter is 1000, and callback/event unit U has identity 77. Act II/environment and valid player P in Act II exist (P has a client whose act is 1 and state is 4); flags are clear, other acts absent unless a case names an act or clock for them (such a case supplies that act and environment), no intervening frame, empty log. A "resolved" client has its direct-player flag set and a cached reference to its player, so it makes no lookup request; an "unresolved" client has the flag clear and a zero cached reference, so it makes no lookup request and no other request. Altar logs list flag and item tests as entries in the request log. A request "over supplied players" (pre-start, progress writer, start/reset senders) is logged as one entry per iterated player, in iteration order; a log that begins "after presentation" omits the earlier entries. In server start, reset and allocation logs, an unqualified "intensity" entry means intensity(source 0, act 0) and "colour" means colour(act 0), per WCK-R01a through WCK-R01c. When a server-resync case lists clients, the client list consists of exactly those clients, in that order; P's default client is not in it. Core vectors run advance then transition with their supplied act. Stage vectors log dependency requests in order; supplied callbacks/presentation/sends return normally without unlisted state or seed changes. A call-time snapshot below records the semantic state observed immediately on entering that supplied operation; it adds no function argument and no new stored field. Flag writes persist within a case. Presentation leaves clock fields unchanged. These supplied effects do not verify integration.

Except explicit draw cases, Seeds: no draw; Draws: none within the local slice. Hidden helper draws are excluded. All complete vectors remain runtime_observed: no; partial idle readings are limited to the assertions in the companion.

- **TSW-V01** Shared start/payload distinction (TSW-R01a, WCK-R01b, TSW-R04a, TSW-R04c; boundary: supplied-input stage).
  - Inputs: defaults; pre-start iteration completes for P.
  - Seeds: no draw. Draws: none locally.
  - Trace: state 0 becomes 1; initial ticks 0 is replaced by 4*300=1200.
  - Outputs: return 1; quest state/stage/transition 1/1/0, applied/pending 0/0; clock `(0,3,1200,4,1,1,0)`. Log: pre-start(P) [call-time snapshot state 0/stage 1/transition 0], first intensity [call-time snapshot flag 0/cycle 0/period 3/ticks 0/rate 4/index 1], second intensity [snapshot flag 1/cycle 0/period 3/ticks 1200/rate 4/index 1], colour, World-clock(P,5,0,1), Quest-event(P,10,16).
  - Fresh Inputs: shared-start defaults except quest state 2. Outputs: return 1; state/stage/transition 2/1/0, applied/pending 0/0, clock `(0,3,1200,4,1,1,0)`. Request log: pre-start(P), first intensity [flag 0/cycle 0/period 3/ticks 0/rate 4/index 1], second intensity [flag 1/cycle 0/period 3/ticks 1200/rate 4/index 1], colour, World-clock(P,5,0,1), Quest-event(P,10,16). These are the same call-time snapshot meanings as above.
  - Fresh Inputs: shared-start defaults except diagnostic switch 1 (TSW-R04c). Outputs: as the first case; log pre-start(P), diagnostic text(`changing state to 1` plus newline), then the first case's remaining requests in order. With state 2 and switch 1 instead, no diagnostic text is requested. With state 1 and switch 1: Outputs return 1, state/stage/transition 1/1/0, the first case's clock and request log, and no diagnostic text. With last resync 17 instead: Outputs as the first case with clock `(0,3,1200,4,1,1,17)`. With state 0, switch 1 and Act II absent: Outputs return 0, state/stage/transition 1/1/0, pending 1, applied 0, no clock; log pre-start(P), diagnostic text(`changing state to 1` plus newline); no presentation or start message.
  - Fresh Inputs: shared-start defaults except incoming Act II clock `(3,1,500,128,0,1,0)`. Outputs as the first case (final clock `(0,3,1200,4,1,1,0)`); first intensity [call-time snapshot flag 1/cycle 0/period 3/ticks 0/rate 4/index 1], second intensity [flag 1/ticks 1200], colour, World-clock(P,5,0,1), Quest-event(P,10,16).
  - Fresh Inputs: shared-start defaults except P's client state 3. Outputs and request log exactly as the first case.
  - Discriminates: the incoming-flag-1 case rejects clearing the flag before the first intensity; the client-state-3 case rejects requiring client state 4 for start messages; final ticks 1200 differs from payload ticks 0; reversed message order, storing flag 1 before first intensity, or resetting a nonzero quest state to 1 fails. A broadcast after the Act II test (Act II absent case), one for a retained nonzero state, or stage/transition/state writes on the wrong side of the pre-start request (its snapshot) fails.
  - runtime_observed: no.

- **TSW-V02** Reset retaining resync (TSW-R02b, WCK-R01c, TSW-R04a; boundary: supplied quest-reset stage).
  - Inputs: clock `(4,2,800,4,1,1,17)`, applied 1. Enter reset after supplied altar preparation.
  - Seeds: no draw. Draws: none locally.
  - Trace: clear applied, normal row 2; intensity sees incoming flag 1, then clear flag.
  - Outputs: applied 0; clock `(2,0,0,128,0,0,17)`; intensity [call-time snapshot flag 1/cycle 2/period 0/ticks 0/rate 128/index 0/last resync 17/applied 0], then World-clock(P,2,0,0); no colour or Quest-event request.
  - (b) As above, plus player P2 in act 0 with a client, after P in iteration order: Outputs the same clock/applied and log; no World-clock for P2.
  - (c) As the first case, but P's client state is 3, and Act I also exists with clock `(3,1,20000,128,0,0,0)`: Outputs the same Act II clock/applied and log including World-clock(P,2,0,0); Act I unchanged.
  - (d) As the first case, but incoming clock `(4,2,800,128,0,0,17)`: Outputs applied 0; clock `(2,0,0,128,0,0,17)`; intensity [call-time snapshot flag 0/cycle 2/period 0/ticks 0/rate 128/index 0], then World-clock(P,2,0,0); no colour or Quest-event request.
  - Discriminates: (d) rejects skipping the reset when the flag is already 0; retained 17 rejects full reinitialization; clearing the flag before intensity fails its snapshot; (b) rejects an unfiltered reset send.
  - runtime_observed: no.

- **TSW-V03** Fresh join rebuild (TSW-R01d, WCK-R01b, WCK-R07; boundary: supplied stages in stated order).
  - Inputs: fresh allocation, P flags {2}; run allocation, quest join, snapshot without a frame.
  - Seeds: no draw. Draws: none locally.
  - Trace: bit 2 applies start, state overwrite selects 1, snapshot samples final clock.
  - Outputs: clock `(0,3,1200,4,1,1,0)`, applied 1, state/stage 1/1. Messages: World-clock(P,5,0,1), Quest-event(P,10,16), act-assignment(P,1), World-clock(P,0,1200,1).
  - Discriminates: start and snapshot payload differ. Outer join scheduling is supplied, not inferred.
  - runtime_observed: no.

- **TSW-V04** Join flags/failure (TSW-R01a, TSW-R01d, TSW-R04b; boundary: supplied join stage).
  - Inputs: fresh defaults, shared flags empty, overrides below.
  - Seeds: no draw. Draws: none locally.
  - Trace: primary bits precede bit 2; state overwrite follows either apply result.
  - (a) Flags {2}: Outputs started clock `(0,3,1200,4,1,1,0)`, applied 1, pending 0, state/stage 1/1, TSW-V01 messages.
  - (b) Flags {2,1}: Outputs default clock, state/stage 0/0, shared-record slot-11 bit 13 set, P's flags still {1,2}; no start/pending message.
  - (c) Flags {2,0}: Outputs as (b).
  - (d) Flags {}: Outputs default state/clock; no flag/message writes.
  - (e) Flags {2,3,4}, Act II absent, P act 0: Outputs state/stage/transition 3/2/0, pending 1, applied 0, no clock assignment; log pre-start(P), Quest-event(P,10,1), no World-clock.
  - (f) Flags {2,3}, bit 4 clear, Act II present: Outputs clock `(0,3,1200,4,1,1,0)`, applied 1/pending 0, state/stage/transition 2/2/0; player flags remain {2,3}; TSW-V01 start message/presentation requests in order.
  - (g) Flags {2}, Act II absent, P's player data has no client: Outputs pending 1, applied 0, state/stage/transition 1/1/0; log pre-start(P), Quest-event request with no client, which queues nothing.
  - (h) Flags {1} alone, and separately {0} alone: Outputs default clock, state/stage 0/0, shared-record bit 13 set, P's flags unchanged; no start/pending message.
  - (i) Flags {2,4}, bit 3 clear, Act II present: Outputs clock `(0,3,1200,4,1,1,0)`, applied 1/pending 0, state/stage/transition 3/2/0; player flags remain {2,4}; TSW-V01 start message/presentation requests in order.
  - (j) Flags {2}, quest state 2 before join, Act II present: shared start retains state 2, then the overwrite selects `(1,1)`. Outputs clock `(0,3,1200,4,1,1,0)`, applied 1/pending 0, state/stage/transition 1/1/0; TSW-V01 start message/presentation requests in order.
  - (k) Flags {2}, pending 1 before join, Act II present: Outputs clock `(0,3,1200,4,1,1,0)`, applied 1, pending 1 retained, state/stage/transition 1/1/0; TSW-V01 start message/presentation requests in order.
  - Discriminates: (k) rejects clearing pending on join success; (i) rejects requiring bit 3 with bit 4 for `(2,3)`; (j) rejects skipping the default `(1,1)` overwrite; (h) rejects testing bit 2 before the primary bits or requiring bit 2 for the bit-13 write; (e) rejects rollback, bit-3 precedence, and filtering the pending request by act; (f) requires the bit-3-only state/stage selection; (g) rejects aborting or undoing writes when no client exists.
  - runtime_observed: no.

- **TSW-V05** Altar short circuit/flags (TSW-R02a, TSW-R02b; boundary: supplied reset/flag slice).
  - Inputs: clock `(0,3,1200,4,1,1,0)`, applied 1, object mode marker 0, supplied `vip ` true; intervening callbacks complete without flag/state changes. Drop/later quest work is excluded.
  - Seeds: no draw before this slice's endpoint. Draws: none; later draw is TSW-V07.
  - Trace: primary bits decide whether secondary tests run.
  - (a) Flags {14}, bits 0/1 clear: Outputs no secondary tests, clock `(2,0,0,128,0,0,0)`, applied 0, reset World-clock(P,2,0,0); ordered set 1,set 13,clear 14 yields flags {1,13}; return 0 after supplied remaining work.
  - (b) Flags {1,14}, slot-10 bit 0 false: Outputs `vip ` check true, rejection(19,P), return 0; clock/applied/flags unchanged, no reset or `hst ` test.
  - (c) Flags {14}, object mode marker 2, and separately 1 and -1: Outputs return 0, unchanged clock/applied/flags, no reset/secondary tests.
  - (d) Flags {14}, Act II absent (so the preamble's clock is not present): Outputs no clock write or World-clock request, applied stays 1; flags {1,13} after supplied callbacks; return 0.
  - (e) Flags {1,14}, slot-10 bit 0 true: Outputs rejection(19,P), return 0, unchanged clock/applied/flags; neither item-presence test is requested and no reset message occurs.
  - (f) Flags {1,14}, slot-10 bit 0 false, supplied `vip ` false/`hst ` true: Outputs those three secondary tests in order, rejection(19,P), return 0, unchanged clock/applied/flags, no reset message.
  - (g) Eligibility-only boundary: flags {14}, operated object absent. Outputs no mode-marker read, no eligibility rejection/return, continuation to supplied preparation. Later object-dependent callback/reset/drop completion is TSW-Q06 and is not asserted by this case.
  - (m) Flags {1,14}, slot-10 bit 0 true, object mode marker 2: Outputs rejection(19,P), return 0, unchanged clock/applied/flags; the mode marker does not pre-empt the rejection.
  - (h) Flags {0,14}, bit 1 clear, slot-10 bit 0 true: Outputs primary bit-1 test false then bit-0 test true, slot-10 test true, rejection(19,P), return 0; unchanged clock/applied/flags, no item-presence/reset requests.
  - (i) Quest 10 absent, with no private-state inputs for that absent quest: Outputs return 0, input clock/player flags retained; no quest record read/write, player-flag lookup, preparation, reset, rejection or draw from this slice.
  - (j) Active 2, flags {14}, Act II present, other inputs and supplied effects from this vector's preamble: Outputs no secondary tests, clock `(2,0,0,128,0,0,0)`, applied 0, reset World-clock(P,2,0,0), ordered set 1/set 13/clear 14, flags {1,13}; return 0 after supplied remaining work. This slice still excludes drops/later quest work.
  - (k) Flags {1,14}, slot-10 bit 0 false, supplied `vip ` false and `hst ` false: Outputs primary bit-1 test true, then slot-10, `vip ` and `hst ` tests all false, no rejection; clock `(2,0,0,128,0,0,0)`, applied 0, reset World-clock(P,2,0,0); bit-0 test false, set 1, set 13, clear 14, flags {1,13}; return 0 after supplied remaining work.
  - (l) Flags {0,14}, slot-10 bit 0 false, supplied `vip ` false and `hst ` false: Outputs bit-1 test false, bit-0 test true, three secondary tests false, no rejection; clock `(2,0,0,128,0,0,0)`, applied 0, reset World-clock(P,2,0,0); the later bit-0 test is true, so no set/clear request and flags stay {0,14}; return 0 after supplied remaining work.
  - (n) Flags {} (bit 14 already clear): Outputs no secondary tests, clock `(2,0,0,128,0,0,0)`, applied 0, reset World-clock(P,2,0,0); ordered set 1, set 13, clear 14 yields flags {1,13}, bit 14 still clear; return 0 after supplied remaining work.
  - Discriminates: clear 14 rejects setting 14, and (n) rejects toggling it; (k) rejects rejecting whenever a primary bit is set; (l) rejects writing bits 1/13/14 regardless of bit 0; (d) rejects clearing applied before checking act; (e)/(f)/(h) reject omitted/incorrect flag/item tests; (g) rejects treating an absent object as a nonzero mode marker; (i) rejects doing quest work without the quest; (m) rejects testing the mode marker before eligibility; (j) rejects treating altar active as equality to 1.
  - runtime_observed: no.

- **TSW-V06** Act recipients (TSW-R03, TSW-R04a; boundary: supplied start).
  - Inputs: P act 1 then P2 act 0, each with a client; Act I clock `(3,1,20000,128,0,0,0)`, default Act II; no frame.
  - Seeds: no draw. Draws: none locally.
  - Trace: apply addresses Act II; recipient loop rejects P2.
  - Outputs: Act I unchanged; Act II `(0,3,1200,4,1,1,0)`; P gets TSW-V01 requests in order, P2 none.
  - (b) Fresh start or reset whose player iteration supplies P (act 1, with client) then P2 in act 0 with no client: Outputs P's sender requests, then the fatal assertion path at P2; no World-clock/Quest-event request for P2 and no normal return from that sender.
  - (c) Fresh start with P then P3, both act 1 with clients: Outputs clock as above; log after presentation World-clock(P,5,0,1), Quest-event(P,10,16), World-clock(P3,5,0,1), Quest-event(P3,10,16).
  - (d) Fresh start with P (act 1) then P4 (act 2) and P5 (act 4), each with a client: Outputs clock as above; P gets TSW-V01 requests in order; P4 and P5 get none.
  - Discriminates: all-act reset/unfiltered send fails; (d) rejects a "nonzero act" filter; (b) rejects silently skipping a clientless recipient in any act; (c) rejects batching all World-clock requests before all Quest-event requests.
  - runtime_observed: no.

- **TSW-V07** Seed owners/budgets (TSW-R01g, TSW-R02c, TSW-R07a, TSW-R07b; boundary: direct bodies and supplied draw sites).
  - Inputs: game seed {7,666}, unit seed {9,666}; each case is fresh; direct bodies exclude helpers.
  - Seeds: each quest seed stated below. Draws: none in (a)/(d)/(f), one quest step in (b)/(c)/(e)/(g)/(h).
  - Trace: NRF-R01 step then the specified reduction; no scenario-wide budget inferred.
  - (a) Direct-body case, quest seed {0,666}: Outputs all game/unit/quest seeds unchanged by the eight bodies.
  - (b) state 0/scheduled 0, quest seed {0,666}: Draws quest modulus 2 result 0; parameter 15. Outputs seed {666,0}, schedule(TSW-R01h timer callback,15) with due value 1000+15=1015 [call-time snapshot scheduled 0, quest seed {666,0}], then scheduled 1.
  - (c) state 0/scheduled 0, quest seed {1,666}: Draws quest modulus 2 result 1; parameter 16. Outputs seed {1791398751,0}, schedule(TSW-R01h timer callback,16) with due value 1016, then scheduled 1.
  - (d) state 0/scheduled 1, quest seed {0,666}: Draws none; Outputs unchanged seed, no schedule, scheduled 1.
  - (e) Altar inline site after supplied amulet requests, seed at site {0,666}: Draws quest modulus 5 result 1. Outputs seed {666,0}, six gold-pile requests afterward; no game/unit step at this site.
  - (f) State 1/scheduled 0, quest seed {0,666}: Draws none; Outputs unchanged seed, no schedule, scheduled 0. This differs from an eligible state-0 request.
  - (g) Supplied inactive altar slice (TSW-R02c), active 0, default clock, applied 1, player flags {14}, object mode 0; preceding preparation/drop helpers complete with no clock/flag/seed writes and requested amulet count 0; quest 9 absent; seed at inline site {0,666}. Draws quest modulus 5 result 1. Outputs quest seed {666,0}, six gold-pile requests, default clock retained, applied 1 and flags {14} retained, no reset/World-clock request; return 0 after supplied remaining operations. No full drop/helper effect is asserted.
  - (h) state 0/scheduled 0, quest seed {0,667}: Draws quest modulus 2 result 1; parameter 16. Outputs seed {667,0}, schedule(TSW-R01h timer callback,16) with due value 1016, then scheduled 1.
  - (i) state 2/scheduled 0, quest seed {0,666}, and separately state 255/scheduled 0: Draws none; Outputs unchanged seed, no schedule, scheduled 0.
  - (j) state 0/scheduled 2, quest seed {0,666}: Draws none; Outputs unchanged seed, no schedule, scheduled 2.
  - (k) Altar inline site as (e), seed at site {0,668}: Draws quest modulus 5 result 3. Outputs seed {668,0}, eight gold-pile requests. Seed at site {0,665} instead: modulus 5 result 0; Outputs seed {665,0}, five gold-pile requests.
  - Discriminates: (k) rejects reductions that agree with `5+(new lo mod 5)` only at result 1; (i)/(j) reject testing state or scheduled for equality to 1 rather than nonzero; wrong seed owner, two steps, scheduling before the draw, ignoring nonzero state, or resetting an inactive quest fails; (h) rejects taking parity from the pre-step low word (0, giving 15). Helpers may draw independently.
  - runtime_observed: no.

- **TSW-V08** Progress exact bits (TSW-R01b; boundary: supplied flag stage).
  - Inputs: found quest 10, flags {}, fresh states 0/1/2/3/255; separate state 3 flags {1}, state 3 flags {0}, and missing-quest cases.
  - Seeds: no draw. Draws: none locally.
  - Trace: >=1 sets bit 2; exact 2/3 selects extra bit 3/4.
  - Outputs: flags {}/{2}/{2,3}/{2,4}/{2}; all return 0. State 3 flags {1} stays {1}; fresh state 3 flags {0} stays {0}; missing quest stays {}; each returns 0 without flag writes.
  - Fresh state 1 with flags {2,3,4}: Outputs flags {2,3,4}, return 0.
  - Discriminates: >=3 setting bit 4 fails at 255; the state-1 case with flags {2,3,4} rejects clearing bits 3/4 at a lower state.
  - runtime_observed: no.

- **TSW-V09** Retry/timer/animation state (TSW-R01e, TSW-R01f, TSW-R01h, TSW-R01i, TSW-R04c, TSW-R05a; boundary: supplied start slices).
  - Inputs: fresh defaults; overrides below.
  - Seeds: no draw except case (g), whose quest seed is explicit. Draws: none locally except (g).
  - Trace: run one named slice; retain writes preceding failure.
  - (a) Animation with Act II absent: Outputs marker B 1/unit marker 77, state 3/stage 1/transition 0, applied 0/pending 1; log pre-start(P), mode-update(U,0); no start message and no second pre-start request, because stage is then 1.
  - (b) Event 0x28, pending 1, act present: Outputs started clock `(0,3,1200,4,1,1,0)`, pending 0/applied 0/state 1/stage 1; P bit 2 set; log pre-start, presentation, start messages, progress writer(P) [call-time snapshot pending 1/applied 0/state 1/stage 1], then the pending-clear write to 0.
  - (c) Timer scheduled 1/applied 1/pending 1: Outputs scheduled 0/applied 1/pending 1, return 1; no requests/clock writes. Fresh applied 2 instead: Outputs scheduled 0/applied 2/pending 1, return 1; no requests/clock writes.
  - (d) Timer scheduled 1/applied 0/pending 0, Act II absent: Outputs scheduled 0/pending 1/state 1/stage 1/transition 0, return 1; pre-start only.
  - (e) Direct retry argument 1/pending 1: Outputs started clock, pending 0/applied 1; quest state 0/stage 0/transition 7 retained; presentation only, no pre-start/progress/message request.
  - (f) Event 0x28/pending 1, Act II absent: Outputs pending 1/applied 0/state 1/stage 1/transition 0, default clocks retained; pre-start(P) only, no progress/message requests. With auxiliary identity 0, TSW-R01i makes no request.
  - (g) Fresh event key 0x2C or 0x2D independently, state 0/active 1/scheduled 0; quest seed {0,666}. Draws: one quest step, modulus 2 result 0, parameter 15. Outputs seed {666,0}, schedule(TSW-R01h timer callback,15), scheduled 1, default clock and state/stage/transition 0/0/7 retained; no shared-start or message request. Event producers and timer delivery remain gated.
  - (h) Timer scheduled 1/applied 0/pending 0, Act II present: Outputs scheduled 0/applied 1/pending 0/state 1/stage 1/transition 0, started clock `(0,3,1200,4,1,1,0)`, P flags {2}, return 1. Log: pre-start(P), intensity, intensity, colour, World-clock(P,5,0,1), Quest-event(P,10,16), progress writer(P) [call-time snapshot applied 1/pending 0/scheduled 0/state 1/stage 1]. Fresh scheduled 1/applied 0/pending 1, Act II present instead: Outputs as (h) except pending 1 retained, with progress-writer snapshot pending 1.
  - (i) Fresh animation start slices with active 0 or 2 independently, state 0: Outputs marker B 1/unit marker 77, default state/stage/transition/applied/pending/clock retained; log mode-update(U,0) [call-time snapshot marker B 1/unit marker 77] only. Fresh active 1/state 2 instead skips the start, then finds stage 0: Outputs state 2, stage 2, transition 0, marker B 1/unit marker 77, applied/pending/clock retained; log pre-start(P), mode-update(U,0). Fresh active 1/state 2/stage 1 with mode argument 2 and the callback unit absent: Outputs unit marker -1, marker B 1, stage/state/transition 1/2/7 retained; log mode-update(absent,2) only. The mode-update's effects are supplied.
  - (j) Fresh event key 0x29, pending 1/state 0/active 1: Outputs pending 1 and default state/stage/transition/applied/clock retained; no draw, schedule, shared-start, progress/message, roster or diagnostic request.
  - (k) Fresh direct-retry cases: argument 0/pending 1/quest present; argument 1/quest absent, with no private-state inputs for that absent quest; argument 1/pending 2/quest present. Outputs in each: default clock retained, no presentation/pre-start/progress/message request. In the two quest-present cases pending remains 1/2 respectively and state/stage/transition/applied stay 0/0/7/0; the absent quest has no private state to read or write.
  - (l) Fresh keys 0x2C/0x2D independently, active 2/state 0/scheduled 0, quest seed {0,666}: Outputs seed {0,666} unchanged, scheduled 0, active 2 and default state/stage/transition/applied/pending/clock retained; no draw/schedule/shared-start/progress/message request from the slice.
  - (m) Event auxiliary identity 0x28, each fresh: (m1) key 0x29, roster word 3, state 2, diagnostic switch 1: Outputs state 3, other quest fields retained; log roster-update(77), diagnostic text(`changing state to 3` plus newline). (m2) key 0x29, roster word 0, state 1: Outputs state 1 retained, no request. (m3) key 0x29, roster word 3, event unit absent, state 0: Outputs state 0 retained; log roster-update(-1) only. (m4) key 0x28, pending 1, state 2, roster word 0, switch 0, Act II present: Outputs started clock `(0,3,1200,4,1,1,0)`, pending 0, applied 0, state 3, stage 1, transition 0, P flags {2,3}; log pre-start(P), intensity, intensity, colour, World-clock(P,5,0,1), Quest-event(P,10,16), progress writer(P) [snapshot state 2/pending 1], pending-clear write, then the state-3 write. With auxiliary identity 0 instead, (m4) ends at state 2 with the same clock. (m5) key 0x29, roster word 0, diagnostic switch 1, state 3 and separately state 4: Outputs state retained, no request.
  - (n) Fresh event key 0x28, pending 2, auxiliary identity 0, Act II present: Outputs pending 2, default state/stage/transition/applied/clock retained; no pre-start, presentation, progress or message request.
  - (o) Fresh animation start slice, active 1, state 0, Act II present: Outputs marker B 1/unit marker 77, started clock `(0,3,1200,4,1,1,0)`, applied 1, pending 0, state 3, stage 1, transition 0. Log: pre-start(P), intensity, intensity, colour, World-clock(P,5,0,1), Quest-event(P,10,16), mode-update(U,0); no second pre-start request, because stage is then 1.
  - (p) Fresh animation start slice, active 1, state 1, Act II present: shared start retains state 1, then the slice writes state 3. Outputs marker B 1/unit marker 77, started clock `(0,3,1200,4,1,1,0)`, applied 1, pending 0, state 3, stage 1, transition 0. Log: pre-start(P), intensity, intensity, colour, World-clock(P,5,0,1), Quest-event(P,10,16), mode-update(U,0).
  - (q) Fresh event key 0x2C, active 1, state 1, pending 1, scheduled 0, quest seed {0,666}, auxiliary identity 0: Outputs pending 1, seed unchanged, scheduled 0, default clock, state/stage/transition 1/0/7 retained; no draw, schedule, start, progress or message request.
  - (r) Fresh event key 0x2C, active 1, state 0, scheduled 0, applied 1, pending 1, quest seed {0,666}, auxiliary identity 0: Draws one quest step, modulus 2 result 0, parameter 15. Outputs seed {666,0}, schedule(TSW-R01h timer callback,15) with due value 1015, scheduled 1, applied 1/pending 1 retained, default clock and state/stage/transition 0/0/7 retained; no shared-start, progress or message request.
  - (s) Fresh animation start slice, active 1, state 255, stage 0: the unsigned state test skips the start, then stage 0 is found. Outputs state 255, stage 2, transition 0, marker B 1/unit marker 77, applied/pending/clock retained; log pre-start(P), mode-update(U,0).
  - Discriminates: (r) rejects adding an applied or pending guard to scheduling; (s) rejects a signed state test (255 as -1 would run the start); (i) rejects writing marker B/unit marker after the mode-update request; (p) rejects an animation start gated on state 0 only; (q) rejects treating key 0x2C as a pending retry or scheduling at nonzero state; (b) does not set applied and progress sees pending 1 before it is cleared; (c) clears scheduled before early return; (e) bypasses shared start writes/messages; (f) retains pending on failure; (g) schedules rather than applying immediately; (h) progress sees applied 1, and its pending-1 variant rejects clearing pending on timer success; (i)/(j)/(k)/(l) reject bypassing the exact eligibility guards, including treating schedule-active as nonzero; (a)/(i) reject skipping the always-made marker writes and mode-update, or the stage-0 branch; (m) rejects gating the auxiliary slice by event key, running it before the start stage, an else-chain, or a state>=2 test (m5); (n) rejects a nonzero-pending test for key 0x28; (o) rejects writing state 3/applied only on failure.
  - runtime_observed: no.

- **TSW-V10** Altar ordered requests and timer callback (TSW-R02b, TSW-R02c, TSW-R02d, TSW-R04c, TSW-R07b; boundary: supplied altar operations).
  - Inputs: player flags {14}, slot-10 bit 0 clear, object O identity 55 with mode marker 0; object-source(O) returns s; the participant iteration sets the amulet count to 2; quest seed {0,666} at the draw site; other supplied operations return normally without other writes or draws.
  - Seeds: quest seed {0,666}. Draws: one quest step, modulus 5 result 1.
  - Trace: 666 mod 5 = 1, so 5+1 = 6 gold-pile drops.
  - (a) Active 1, applied 1, clock `(0,3,1200,4,1,1,0)`, diagnostic switch 1, quest 9 present with counter 4 (it stays 4), source-derived(s) returns d; first amulet drop returns item I1, second returns none. Outputs: source result s, derived result d, mode argument 2, marker B 1, unit marker 55, marker A 1, amulet count 2, applied 0, clock `(2,0,0,128,0,0,0)`, flags {1,13}, quest seed {666,0}, return 0. Log: test slot-11 bit 1 (false), test slot-11 bit 0 (false), object-source(O), source-derived(s), mode-update(O,1) [call-time snapshot mode argument 0/marker A 0/marker B 0/unit marker 0], intensity, World-clock(P,2,0,0), diagnostic text(`Finished quest`), quest-notification(4), game-notification(6), participant iteration(P), test slot-11 bit 0 (false), set 1, set 13, clear 14, post-flag iterations A(O), B(O), C(O), quest-record update(13), source-derived(s), O item-code field := `vip `, amulet drop(O) [call-time snapshot quest seed {0,666}], item-flag mark(I1,16,1), amulet drop(O), amulet-result(1), context request(4), six gold-pile drops(O) [first one's snapshot quest seed {666,0}], completion iteration, timer request(altar callback, 10) with due value 1010. With diagnostic switch 0 instead: the same outputs and log without the diagnostic text.
  - (b) Active 0, applied 1, default clock; the first amulet drop returns item I1, the second returns none; quest 9 present with counter 4. Outputs: marker A 1, mode argument 2, amulet count 2, marker B/unit marker/source result/derived result 0, quest 9 counter 5, clock/applied/flags unchanged, quest seed {666,0}, return 0. Log: test slot-11 bit 1 (false), test slot-11 bit 0 (false), mode-update(O,1) [call-time snapshot mode argument 0/marker A 0], participant iteration(P), object-source(O), source-derived(s), O item-code field := `vip `, amulet drop(O), item-flag mark(I1,16,1), amulet drop(O), context request(4), six gold-pile drops(O). No reset, diagnostic, notification, flag, amulet-result, completion-iteration or timer request.
  - (c) As (b) but quest 9 absent: Outputs the same, with no counter write; return 0.
  - (d) Altar timer callback, state 4, stage 2, marker C 1: Outputs transition 0, stage 3, marker C 0, return 1; log pre-start(P). Separate state 3, stage 2, marker C 1: Outputs stage 2 retained, marker C 0, return 1, no request. Separate state 5, stage 2, marker C 1: Outputs stage 2/transition 7 retained, marker C 0, return 1, no request. Separate state 4, stage 0, transition 7, marker C 1: Outputs transition 0, stage 3, marker C 0, return 1; log pre-start(P).
  - (e) As (a), except the supplied mode-update(O,1) sets active to 0. Outputs: active 0, applied 1, clock `(0,3,1200,4,1,1,0)` retained, flags {14}, source result s, derived result d, mode argument 2, marker B 1, unit marker 55, marker A 1, amulet count 2, quest seed {666,0}, return 0. Log: test slot-11 bit 1 (false), test slot-11 bit 0 (false), object-source(O), source-derived(s), mode-update(O,1), participant iteration(P), source-derived(s), O item-code field := `vip `, amulet drop(O), item-flag mark(I1,16,1), amulet drop(O), amulet-result(1), context request(4), six gold-pile drops(O), completion iteration, timer request(altar callback, 10). No reset, diagnostic, notification, flag test/write, post-flag iteration or quest-record update.
  - (f) As (a), except the supplied participant iteration sets active to 0 as well as the amulet count 2. Outputs: as (a) except active 0 and flags {14}. Log: as (a) with the entries from the second test of slot-11 bit 0 through quest-record update(13) removed.
  - (g) Active 1, applied 1, clock `(0,3,1200,4,1,1,0)`, diagnostic switch 0, player flags {14}; the participant iteration sets amulet count 2; the first amulet drop returns item I1 and also sets the amulet count to 1. Outputs: one amulet-drop attempt only, item-flag mark(I1,16,1), amulet-result(1), amulet count 1; otherwise the outputs of (a) with no diagnostic text.
  - (h) As (g), except amulet count is 3 before the altar operation and the participant iteration sets nothing: Outputs amulet count 0, no amulet-drop or item-flag mark request, amulet-result(0); otherwise the outputs of (g) without its amulet entries. Separately, the participant iteration sets amulet count -1: Outputs amulet count -1, no amulet-drop or item-flag mark request, amulet-result(0); otherwise as before.
  - Discriminates: (h) rejects omitting the amulet-count reset before the participant iteration and an unsigned attempt/count comparison (which would attempt drops for -1); the mode-update snapshots in (a)/(b) reject writing the mode argument or markers before that request; (g) rejects capturing the amulet count once before the loop; (a) rejects adding successes to quest 9 on the active branch, omitting the second active test's flag/iteration block, counting attempted rather than successful drops, or scheduling before the gold drops; (b) rejects resetting or scheduling on the inactive branch, losing the quest-9 addition, or adding attempts (6) instead of successes (5); (a) with switch 0 rejects an unconditional diagnostic; (d) rejects a state-independent stage write or a state>=4 test; (e)/(f) reject dropping either active re-test.
  - runtime_observed: no.

### 8.2 Observation vectors

Pending [companion procedures](../research/spec-evidence/quest-tainted-sun.md#observation-procedures) specify setup/capture limits.

| Id | Measure | Expected / comparison |
|---|---|---|
| **TSW-O01** | quest start/altar with another act as control | server returned clock, message payload, client assignment, reset tuple; appearance/trigger identity remain gated |
| **TSW-O02** | fresh joins before/during/after progress, repeated difficulty | flag selection, retained failure state, current snapshot; save encoding remains gated |
| **TSW-O03** | second player join and either player's altar operation | shared Act II clock, own-act snapshot, exact message order/delivery |

## 9. Acceptance and dependencies

Partial acceptance reproduces supplied-input assertions of TSW-V01 through TSW-V10 for all static leaves in the Id map. Writes, returns, direct draw sites and request order are available independently of event/save producers and downstream delivery/presentation. TSW-V05 excludes drops/later quest work and TSW-V10 supplies drop, iteration and callback results; TSW-V07 excludes helper draws. Clock tuples asserted by these vectors are produced by the world clock's assignments (WCK-R01a through WCK-R01c) and snapshot (WCK-R07). Passing these does not establish complete quest, packet, save, visual or RNG parity.

Full reproduction also requires gated assertions of TSW-R01c, TSW-R05b and TSW-R06; independently established integrations TSW-Q05/TSW-Q06; and applicable matching TSW-O01 through TSW-O03 and TSW-Q03/TSW-Q07 observations. Status/partial acceptance imply no user approval.

**Depends on:** NRF-R01/NRF-R02 in [numeric/RNG foundations](core-numeric-and-rng-foundations.md) for quest-seed arithmetic/reduction and its i32/low32 notation. WCK-R01b (start assignment) and WCK-R01c (reset assignment) in [world clock](drlg-world-clock.md), requested by TSW-R01a, TSW-R02b and TSW-R05a; WCK-R01a (allocation) and WCK-R07 (join snapshot) as the supplied stages around join in TSW-V03. Section 3's supplied integration operations remain TSW-Q06. GDA-R04a in [game-data fields](core-game-data-fields.md) supplies only the distinction between code-built records and loaded data, without quest flag meanings. Owners of the cited GDA ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA).

**Depended on by:** no known non-gap plan or executable consumer cites this contract. WCK-V10 in [world clock](drlg-world-clock.md) runs TSW-R01d as a supplied join stage.

## 10. Related specs

- [World clock](drlg-world-clock.md): the per-act clock, its start/reset assignments, server resynchronization, join snapshot and client handling.
- [Numeric/RNG foundations](core-numeric-and-rng-foundations.md): shared arithmetic and frame conventions.
- [Game-data fields](core-game-data-fields.md): names and code-built versus table-record distinction.

## Appendix A. Provenance

Evidence targets original Windows x86 Lord of Destruction Expansion 1.13c D2Game, D2Common and Fog, identified in the [source manifest](../research/source-manifest.md). The [public companion](../research/spec-evidence/quest-tainted-sun.md) records RVAs/ordinals, dated readings, correction rationale, partial observations and limits. Static readings dated 2026-10-12 retain their scope; targeted matched-reconstruction, instruction and original-image checks are dated 2026-10-15.

No gameplay observation of the quest stages is recorded. All executable vectors remain runtime_observed: no.
