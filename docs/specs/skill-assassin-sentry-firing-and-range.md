# Assassin sentry firing and range

**Description**

Assassin sentries repeatedly look for a target and decide whether to fire. Their attack range accounts for the target's size, and their limited shot supply runs down as they shoot; a sentry whose supply runs out dies. What the fired missiles do once launched is a separate behavior.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: SFR-R, SFR-Q, SFR-V, SFR-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-assassin-sentry-firing-and-range.md)
Open evidence gaps: [gap file](../plans/gaps/skill-assassin-sentry-firing-and-range.md)

## 1. Scope

Defines the think side of the existing shared Assassin trap-family boundary: reached sentry thinking, budget, stored/room target selection and candidate-size distance, up to the shot request. Death Sentry reuses the named budget, target and size helpers. The charged-bolt burst and the shared missile-speed stage reached after the shot request belong to [Assassin sentry missile launch](skill-assassin-sentry-missile-launch.md). Missile flight/collision, damage/hit resolution and the full skill/monster AI catalogs belong to their owners. Summon/AI entry, animation completion and dependency integration are gated below; no caller-independent shot interval is inferred.

## 2. Trigger and ownership

D2Game computes the think, budget and target decisions; D2Common supplies sizes, positions, skill selection/calculation and spatial predicates. The server AI event calls the think; its exact AI101 binding remains Q02. A think checks budget, skill, state, target, strict range and chance before requesting skill mode; the skill action that mode reaches (for charged-bolt sentries, the burst in STT-R02) is outside this boundary. First and later scheduling remain Q01; no ordering against unrelated frame systems or named client message is established. [Binary chains](../research/spec-evidence/skill-assassin-sentry-firing-and-range.md#dispatch-and-function-records) identify the reached entry and its output paths.

## 3. Inputs and data

Arithmetic is signed i32 with low32 wrapping and signed division truncating toward zero unless stated. Positions/size arithmetic fixtures use differences whose absolute values and doubled sums remain nonnegative signed i32. Extreme static-coordinate overflow outside that domain remains Q04. Units/time/seeds consume NRF-R01/NRF-R02/NRF-R08/NRF-R09a/NRF-R10. Supplied dependency results describe local request/return boundaries and do not verify those producers.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| sentry, candidate, searcher, owner, room, target | runtime units/lookup results | optional objects; kinds player 0, monster 1, missile 3 | Eligibility and ownership |
| sentry owner reference, owner room town test | the monster's stored owner reference; room-to-level town predicate | optional reference; boolean | R02 owner gate order |
| difficulty and AI parameters | difficulty 0..2; MonStats.`aip1`..4 and N/H variants; GDA-R08a | difficulty u8; parameters signed i16 | chance, failed-roll delay, absent-target delay, range |
| firing skill and record count | MonStats.Skill1; Skills collection | signed i16 identity; signed i32 count | lookup/shot action |
| skill instance and level | skill selection and bonus-level operations | optional node; i32 level | expression context and mode request |
| shot mode value | the selected skill instance's mode, supplied by the skill-node operation | i32 | R01f mode request value |
| budget | AI shot counter | signed i32; initially -1 | lazy result and remaining shots |
| calculation | Skills.`calc4` | compiled reference and i32 result | budget; SCX-R11/SLD-R19 |
| states | 12 inferno, 146 invis | active-state booleans; GDF-R03 | state clear, candidate gate |
| unavailable/eligibility conditions | `aliveExclusionCondition`, mode; candidate eligibility A | boolean; mode i32 | structural rejects; A producer Q06 |
| positions and size | integer path parts; candidate size operation | dynamic u16 parts; size signed i32, monster SizeX signed i8 | octagonal offsets |
| stored-target selector and results | searcher state and resolution | selector i32; optional unit/distance | shortcut and reached selector 3 draw |
| room lists, slot class, chooser results | ordered traversal, class byte, selector 7 result | ordered units; u8 class; i32 distance | two-slot selection |
| spatial/root operations | reach(searcher, candidate, 0), collision(candidate, searcher, mask), owner resolution, sameTierNeutral, directional interaction | PVP-R05a/R07 existing-list contracts; interaction mask 8 | exact consumer routing; Q05/Q07/Q08 |
| current frame and AI events | game clock and unit event storage | i32 frame; pending event type 2 | rescheduling/removal requests |

| Mapping / constant | Origin | Binding |
|---|---|---|
| state 12/146, mode 0 and player 17/monster 12 tests | engine code; GDF-R03/GDF-R05a/GDF-R06a supply names where established | fixed original-code tests; names do not imply remapping |
| room cap 48, threshold 80, event 2 | engine code | fixed original constants; event broader naming remains limited to this behavior |
| aip fields, Skill1, `calc4`, SizeX, Missiles.Size | table data | supplied prepared-record values; original row values require Q03 records |
| unknown monster slot class and candidate condition A | unresolved | numeric/boolean supplied inputs; no guessed name |

## 4. Rules (execution order)

R01 subrules give think order; R02 and R03 subrules are called at their stated stages. R04/R10 are distance operations, followed by timing/context boundaries. R01/R03/R06/R07 are grouping prefixes, not additional rules.

**SFR-R01a** Initial budget check [verified in 1.13c code (static)]. Run R02 with decrement=false. A lost result stops this think before skill lookup, target search or its direct chance draw.

**SFR-R01b** Missing selected skill [verified in 1.13c code (static)]. After a surviving budget check, request the unit's selected instance of MonStats.Skill1, passing the raw sign-extended identity with no bounds check (bounds are checked only inside R02's lazy initialization). If absent, request clearing its current skill and monster death mode 0, then stop. Instance selection and bonus-level production consume supplied results; their actual integration is Q08.

**SFR-R01c** Prepare and search [verified in 1.13c code (static)]. If state 12 inferno is active, request setting state 12 off (RLI-R03) before R03a's target search; when it is inactive no state request is made. Search on every reached think, while still allowing the stored-target shortcut. State testing and set-off request routing consume RLI-R03; broader lifecycle effects remain that dependency's gates.

**SFR-R01d** Strict active range [verified in 1.13c code (static)]. No returned target, or signed returned distance >= the difficulty-selected signed i16 `aip4`, requests the next think after signed i16 `aip3` ticks and stops. A smaller distance proceeds. The consumer uses the supplied search distance; R10 owns its arithmetic.

**SFR-R01e** Chance and failed-roll delay [verified in 1.13c code (static)]. After a target passes range, draw exactly Rand(sentry seed, 100), consuming NRF-R01/NRF-R02. Compare the result as signed i32 with sign-extended `aip1`. A result >= `aip1` requests the next think after sign-extended `aip2` ticks and stops. There is no direct chance draw on an earlier exit. Target and expression operations retain their own draws.

**SFR-R01f** Spend and request the shot [verified in 1.13c code (static)]. Run R02 with decrement=true. If it reports lost, stop. Otherwise remove pending unit AI events of type 2 (events of other types stay), request selecting the firing skill with the returned target and zero point coordinates, then request entering the mode value that the selected skill instance supplies. These are ordered requests; actual animation, event storage and resulting skill execution remain Q08. No next-think delay is set on this local success branch. The think produces no result value used by this contract; every outcome is one of the stated requests (its executor is Q02).

**SFR-R01g** AI binding [unresolved]. Original AI table entry 101 holds this think operation, but which monster records select that entry (the table-index column, the special-table override) and the event root that runs it are not established. Q02 gates entry from that identity, not the reached-think supplied-input stages.

**SFR-R02** Lazy shot budget [verified in 1.13c code (static)]. A "death request" here is the ordered pair: request clearing the unit's current skill, then request mode 0 for that unit. Every reject below issues it with the supplied unit (absent when the unit is absent) and returns lost=1; how those request operations treat an absent unit is theirs (Q08). The owner checks run in this order and stop at the first failure: unit absent; unit not a monster; monster owner reference absent; owner lookup returns absent; owner room absent; owner room is a town room. An absent owner reference makes no owner lookup request. Then read the signed i32 counter. If counter<0 (any negative value, not only -1): reject a Skill1 identity <0 or >=count, or an absent Skills record, with a death request; obtain the selected skill instance, rejecting absence the same way. Request the level of that instance with bonuses included (not the base level) and evaluate that skill's `calc4` with (sentry, Skill1, that level), persisting its i32 result as the counter. If counter<=0, issue a death request and return lost=1. If positive and decrement is nonzero, subtract 1 with i32 arithmetic and persist it. Return lost=0. No direct RNG draw occurs; owner/room/mode operations and expression evaluation have their declared effects and draws. A counter becoming zero after decrement survives until the next check.

**SFR-R03a** Stored target precedes room walk [verified in 1.13c code (static)]. Request stored-target resolution first. A returned target is used with its returned R10 distance without the room-walk cap; no nearby-room candidate callback runs. Otherwise initialize two empty nearest slots with distance 2147483647 and enter the room search. Whenever a target is returned, from either the stored-target path or the room search, the wrapper also requests the reach predicate(searcher, target, 0) and returns its result as the auxiliary result; the think does not use that auxiliary result. Missing target returns no unit, distance 2147483647 and auxiliary result 0, with no reach request.

**SFR-R03b** Nearby-room enumeration [unresolved]. Selector 6 selects a room-list walker and R03c callback. In supplied room order, it visits each room's unit chain in chain order. Callback zero continues; nonzero stops. R03c returns zero even after replacing a slot, so selection examines the remaining chain. Exact room-list production, linked-list validity and complete real traversal remain Q05.

**SFR-R03c** Two nearest slots [verified in 1.13c code (static)]. For each candidate, execute R03e/R03f/R03h validity before computing R10 distance. Reject invalid candidates, then distances>48. Slot class>1 selects slot A; class<=1 selects slot B. A candidate must be strictly nearer than the slot's existing distance. Only then request the collision predicate(candidate, searcher, 4); a nonzero result rejects it. A zero result replaces the selected slot's unit and distance. Equal distances keep the earlier candidate. Player slot class is 14; a valid monster reads its supplied unsigned class byte, other/missing records yield 0. Its table-column meaning remains Q05.

**SFR-R03d** Search frequency [asm-only]. The think reaches target search on each call surviving budget and skill checks. Its stored-target shortcut remains active; this alone does not establish a newly chosen enemy or an unlocked target after every shot.

**SFR-R03e** Structural candidate rejection [verified in 1.13c code (static)]. Reject an absent searcher or a searcher whose kind is not player/monster, then an absent candidate or a candidate whose kind is not player/monster (both units must be player or monster). Reject candidate then searcher when their unavailable-for-combat condition is true: null unit, `aliveExclusionCondition`, a player in mode 0 or 17, or a monster in mode 0 or 12 (a player in mode 12 or a monster in mode 17 is available). Request the candidate's room and pass that result, absent included, to the room town predicate; a nonzero result rejects. How the predicate treats an absent room is that operation's (Q08). Then reject a false candidate eligibility condition A. There is no direct draw before these exits. Eligibility condition A's producer/meaning remains Q06; supplied condition handling is exact.

**SFR-R03f** Invisible-candidate reach gate [verified in 1.13c code (static)]. If candidate state 146 invis is absent, go directly to R03h with no local draw or reach test. For an invisible player draw Rand(candidate seed, 100); a roll>=80 bypasses the reach test. A lower roll requests reach(searcher, candidate, 0) and rejects a ZERO result. An invisible monster makes no local draw, requests the same ordered predicate and rejects a NONZERO result. Surviving candidates run R03h. The reach operation is not a supplied hostility relation: it consumes range/separation/collision operations whose complete integration remains Q07. Candidate draws persist even if a later distance/collision test rejects.

**SFR-R03g** External validity operations [unresolved]. Room/town, owner lookup, reach, root relation and player interaction operations must provide their original results. Their local consumers are R03e/R03f/R03h. The reach operation reads spatial distance and range, while R03h itself contains no distance check. Q05/Q07/Q08 retain the complete producer contracts; no general hostile/alive predicate substitutes for them.

**SFR-R03h** Final owner-root decision [verified in 1.13c code (static)]. Start searcher root at searcher. For a monster with an owner reference, resolve one owner step and replace the root only if found; for a missile with owner-resolution condition set, do the corresponding one-step resolution. Starting at candidate, repeatedly resolve a monster owner while present and different from the current unit, stopping on missing owner or self-owner. Equal roots return 0. Two player roots request interaction(rootSearcher, rootCandidate, 8) and return that result unchanged. Otherwise request sameTierNeutral(rootSearcher, rootCandidate), returning 1 for zero and 0 for nonzero. No local draw or distance read occurs. Unresolvable owner references retain the current root. This consumes PVP-R07's asymmetric owner resolution and PVP-R05a's directional relation/tier-read contract: player results are raw 0 or 8; other results are 0 or 1. Existing supplied relation nodes/alignment lists are read by those verified contracts; their production remains HOS-Q01/PVP-Q04. There is no depth cap or multi-unit owner-cycle detector.

**SFR-R04** Unit-to-point octagonal distance [verified in 1.13c code (static)]. With supplied unit position (ux, uy) and point (x, y), compute wrapping i32 differences and their i32 absolute values; on the nonnegative coordinate-difference domain compute trunc(i32(2*max(dx, dy)+min(dx, dy))/2). There is no direct draw. Dynamic integer positions consume NRF-R10b; static-position production remains Q04. This operation does not subtract a size. Its callers are not claimed by this contract; it is specified as an isolated operation.

**SFR-R05** First shot scheduling [unresolved]. The summon action requests creating a sentry and a mode update. The first AI event's due frame and therefore cast-to-first-shot latency are Q01; no interval timer in the reached think establishes that latency.

**SFR-R06a** Difficulty-selected AI parameters [verified in 1.13c code (static)]. For difficulty 0, 1, 2 consume the matching base/Nightmare/Hell `aip1`..4 values as signed i16: `aip1` chance threshold, `aip2` failed-roll delay, `aip3` absent/out-of-range delay, `aip4` strict active range. No parameter is clamped by this local consumer. These are field roles, not assertions of any particular original row value.

**SFR-R06b** Post-shot scheduling [unresolved]. Skill-mode completion supplies later thinking. Its exact due-frame production remains Q01; the shot branch itself requests no `aip2`/`aip3` reschedule.

**SFR-R07a** Range arithmetic [verified in 1.13c code (static)]. The think range and room candidate cap use R10's size-reduced octagonal distance, not Euclidean distance. Only room candidates are capped at 48. R01d compares the supplied returned distance strictly below `aip4`.

**SFR-R07b** Coordinate domain [verified in 1.13c code (static)]. Dynamic player/monster/missile positions expose unsigned 16-bit integer parts of a supplied 16.16 path, in NRF-R10a/R10b units: 5 integer units per map tile. This describes getter inputs/results and does not establish every path/static-coordinate producer.

**SFR-R08** Integrated selection [unresolved]. Stored-target resolution, room traversal, both nearest slots and the chooser determine the returned target. For supplied chooser inputs, checked in this order: a present player searcher selects A, even when A is empty; empty B selects A; empty A selects B; signed B distance>5 selects A. Otherwise the chooser temporarily points the searcher path at A's unit, requests path preparation, restores the saved path values, then reads a neutral path result value; nonzero selects A without a selector 7 search. For zero, it runs selector 7 with a fresh best (no unit, distance 2147483647) that excludes A's unit; each other candidate passes validity, reduced distance<=48, slot class>1, strictly below the running best, then collision(candidate, searcher, 4)=0 to become the best. No found unit, or a signed found distance>=20 (a fixed 20, not A's distance), selects B; otherwise the found unit and distance replace A and A is selected. A selected empty slot returns no unit and distance 2147483647. The path result's meaning, actual path preparation/restoration and selector-7 traversal remain Q05.

**SFR-R09** Reached stored-target random branch [verified in 1.13c code (static)]. When the searcher's stored-target selector 3 branch is reached, draw exactly RandMask(searcher seed, 2), consuming NRF-R02/NRF-R04b. The mask argument is 2, so the returned bit result is 0 or 1. Subsequent stored-target routing/state writes are Q05; this rule owns the reached draw and seed owner only.

**SFR-R10** Candidate-size search distance [verified in 1.13c code (static)]. Read positions of candidate and searcher. Obtain the candidate size s: player 2; monster signed i8 MonStats2.SizeX through MonStats.MonStatsEx, with missing row size 0; object supplied size field; missile u8 Missiles.Size, missing record 0; item 1; other/null 0. Compute dx=max(i32(abs(xC-xS)-s), 0), dy=max(i32(abs(yC-yS)-s), 0), then trunc(i32(2*max(dx, dy)+min(dx, dy))/2). Only candidate size is subtracted, in full on EACH axis. Negative supplied SizeX enlarges offsets. Player/monster candidates use this at both stored-target and room callback sites; the stored-target route bypasses the room cap. No direct draw.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SFR-R01a | Initial budget check | verified in 1.13c code (static) | none | SFR-V05, SFR-V16, SFR-V22 |
| SFR-R01b | Missing selected skill | verified in 1.13c code (static) | SFR-Q08: actual skill/mode producers | SFR-V07 |
| SFR-R01c | Prepare and search | verified in 1.13c code (static) | SFR-Q05: complete search integration | SFR-V08, SFR-V09 |
| SFR-R01d | Strict active range | verified in 1.13c code (static) | none for supplied target/distance | SFR-V01, SFR-V02, SFR-V03, SFR-V15, SFR-V19 |
| SFR-R01e | Chance and failed-roll delay | verified in 1.13c code (static) | none | SFR-V04, SFR-V17 |
| SFR-R01f | Spend and request the shot | verified in 1.13c code (static) | SFR-Q08: actual event/skill-mode execution | SFR-V01, SFR-V05, SFR-V16, SFR-V23 |
| SFR-R01g | AI binding | unresolved | SFR-Q02: AI101 entry | none |
| SFR-R02 | Lazy shot budget | verified in 1.13c code (static) | SFR-Q08: actual lookup/mode integration; supplied results available | SFR-V05, SFR-V16, SFR-V22, SFR-V23, SFR-V24 |
| SFR-R03a | Stored target precedes room walk | verified in 1.13c code (static) | SFR-Q05: actual stored-target resolution | SFR-V09, SFR-V25 |
| SFR-R03b | Nearby-room enumeration | unresolved | SFR-Q05: real room enumeration | SFR-V26 |
| SFR-R03c | Two nearest slots | verified in 1.13c code (static) | SFR-Q05: actual slot-class/collision producers | SFR-V06, SFR-V18, SFR-V21, SFR-V26 |
| SFR-R03d | Search frequency | asm-only | SFR-Q05: stored-target production/lifetime | SFR-V09 |
| SFR-R03e | Structural candidate rejection | verified in 1.13c code (static) | SFR-Q06: eligibility producer; SFR-Q08: room predicate | SFR-V10, SFR-V27 |
| SFR-R03f | Invisible-candidate reach gate | verified in 1.13c code (static) | SFR-Q07: actual reach result; local handling verified | SFR-V11, SFR-V12, SFR-V13, SFR-V21, SFR-V28 |
| SFR-R03g | External validity operations | unresolved | SFR-Q05/SFR-Q07/SFR-Q08: complete validity integration | none |
| SFR-R03h | Final owner-root decision | verified in 1.13c code (static) | HOS-Q01/PVP-Q04: relation/alignment producers; supplied existing lists available | SFR-V29 |
| SFR-R04 | Unit-to-point octagonal distance | verified in 1.13c code (static) | none on supplied nonnegative offsets | SFR-V14 |
| SFR-R05 | First shot scheduling | unresolved | SFR-Q01: cast-to-think integration | SFR-O01 |
| SFR-R06a | Difficulty-selected AI parameters | verified in 1.13c code (static) | none | SFR-V04, SFR-V17, SFR-V30 |
| SFR-R06b | Post-shot scheduling | unresolved | SFR-Q01: mode-end to think due frame | SFR-O02 |
| SFR-R07a | Range arithmetic | verified in 1.13c code (static) | none for supplied positions/size | SFR-V01, SFR-V02, SFR-V03, SFR-V15, SFR-V19 |
| SFR-R07b | Coordinate domain | verified in 1.13c code (static) | none for supplied dynamic path | SFR-V31 |
| SFR-R08 | Integrated selection | unresolved | SFR-Q05: complete chooser path effects | SFR-V32 |
| SFR-R09 | Reached stored-target random branch | verified in 1.13c code (static) | none at reached draw; SFR-Q05 upstream/integration | SFR-V33 |
| SFR-R10 | Candidate-size search distance | verified in 1.13c code (static) | none on supplied positions/size; static positions Q04 | SFR-V01, SFR-V02, SFR-V03, SFR-V06, SFR-V15, SFR-V18, SFR-V19, SFR-V20 |

## 5. Outputs and side effects

| Effect | Rules |
|---|---|
| budget initialized/decremented, lost result 0/1 | SFR-R02 |
| current-skill clear/death-mode requests; state 12 removal request | SFR-R01b/SFR-R02; SFR-R01c |
| next-think delay requests; type 2 pending-event removal; skill-select request (Skill1, target, point (0, 0)) and mode request (instance mode value) | SFR-R01d/SFR-R01e; SFR-R01f |
| chosen target/distance, two nearest-slot replacements, validity results, auxiliary reach result | SFR-R03a/SFR-R03c/SFR-R03e/SFR-R03f/SFR-R03h/SFR-R08 |
| Point-distance and candidate-size distance i32 returns | SFR-R04/SFR-R10 |
| candidate or sentry/searcher seed changes at direct draw sites | SFR-R01e/SFR-R03f/SFR-R09 |

Vectors also assert values that are not stored effects: intermediate results of the rules above (resolved owner roots, candidate sizes, reduced offsets, the NRF-R10a tile conversion) and the presence, absence, arguments and count of requests to supplied operations (lookups, room callbacks, evaluations). These are checked through the supplied operations' logs and return values. Dependency operations retain their declared stored effects and draws, including calculation draws and mode changes. No persistence write or named client message is established here.

## 6. Branches and edge cases

Owner loss, absent owner room and town all stop the budget path. A negative counter initializes lazily; zero/negative evaluated budget dies immediately; positive decrement to zero permits that shot. Difficulty selects signed parameter variants without clamps. Chance thresholds<=0 always fail and >99 always pass the direct 0..99 draw. Range is strict; nearest ties are stable within their slot; stored targets bypass the room cap. Invisible players use their own seed and opposite reach-result polarity from invisible monsters. No difficulty/PvP special arithmetic exists in these local stages beyond supplied parameters and root interaction.

## 7. Open questions

Attempted evidence and precise blocked boundaries are in the [gap file](../plans/gaps/skill-assassin-sentry-firing-and-range.md).

| Id | Status / kind | Remaining evidence |
|---|---|---|
| **SFR-Q01** | unresolved / integration | First AI scheduling and skill-mode completion to later-think scheduling; O01/O02 |
| **SFR-Q02** | unresolved / integration | AI table entry 101 already holds this think and its installer/executor path is read (companion); remaining: the MonStats column feeding the table index, when the special-AI table overrides it, and the event root that reaches the executor |
| **SFR-Q03** | unresolved / integration | Original relevant `calc4`/aip/Skill1/SizeX values and their compiled/runtime association |
| **SFR-Q04** | unresolved / outside-boundary | Complete static/other position production and extreme static-coordinate overflow; supplied dynamic getter units consume NRF-R10 |
| **SFR-Q05** | unresolved / behavior and integration | Real stored-target eligibility/state writes, room-list production, class-column meaning and chooser path effects; collision/relation producer contracts |
| **SFR-Q06** | unresolved / outside-boundary | Candidate eligibility condition A meaning and producer; state 146 name is GDF-R03 |
| **SFR-Q07** | unresolved / integration | Complete actual reach result from weapon/range, separation and collision dependencies; local opposite polarities are established |
| **SFR-Q08** | unresolved / integration | Owner/room/skill-level, mode/event storage and root interaction/relation producers; the burst's target resolution and creation are STT-Q04 |
| **SFR-Q09** | unresolved / observation | No gameplay vector has been observed on the original game; observation is possible and pending |

## 8. Test vectors

### 8.1 Executable vectors

All records are synthetic supplied inputs, not archive rows. Each independent case starts with game frame 100; monster sentry S at (1000, 1000), mode 1, counter 3, owner P present in a non-town room; valid firing skill id 20, count 400, present instance, level 3; `aip1`=100, `aip2`=7, `aip3`=11, `aip4`=12 for all three difficulties. Selected skill/level, owner/room, mode/event and reach/collision operations log requests/results. Pure getters and spatial/relation predicates are omitted from asserted call-count logs. Their routes in Trace explain the original decisions; an equivalent extra pure read is not rejected when values, seed changes and effects agree. Logs assert effectful requests, expression draws and their meaningful order. Supplied mode requests persist the requested mode; event removal removes only type 2; scheduling appends due=frame+delay. Skill selection persists skill/target, and the supplied mode request logs one shot. These supplied effects verify local ordering, not Q08 integration. State removal persists active-state membership only.

Ordinary candidate T is a player at the case's raw offset from S, mode 1, eligibility A=true, not unavailable, non-town room, state 146 absent, size 2, slot class 14; final roots differ; supplied directional relation node on P for T has eligibility flags 8. A monster candidate in these fixtures (any named monster T, B or other candidate) has the same mode, eligibility, room, state, tier and collision/reach inputs, no owner link, SizeX 0 and slot class 1 unless a case states otherwise. Existing tier inputs for each unit are a present extended base list with state 105 list and own stat 172/layer 0 = 1; no other lists, minimum adjustment inactive. Collision 0 and reach 1 unless overridden. For validity-only gate vectors V11..V13/V21/V27/V28, searcher owner resolution returns absent and leaves the monster S root; the different player/monster roots with tiers 1/1 produce final result 1 through PVP-R05a. No budget check runs in those gate fixtures. A supplied stored-target resolver returns no target unless specified. Room-callback vectors receive candidates directly; think vectors receive the stated target/distance, so they do not execute room traversal. Local R02 calls execute R02; calculation operation returns stated `calc4`=3 without draws. Supplied dependency operations persist their stated writes, log in order and draw nothing unless a case says otherwise. Unmentioned state stays at entry; logs start empty. Direct draw seed for S and T is {lo=1, hi=0x29A} unless overridden. No-draw cases leave both unchanged. Unless overridden, a think fixture receives supplied search result T at raw offset (3, 0), distance 1. The think consumes the supplied returned distance (R01d). Wherever a think vector also states a raw offset and a reduced value, that is a separate isolated R10 assertion: R10(candidate T, searcher S) on that offset must return the stated returned distance, with S's own size irrelevant. The selected skill instance supplies mode value 9. S state 12 is initially inactive unless a case overrides it. Every vector states only its reached local or supplied-input assertions; integration gates remain in the Id map.

- **SFR-V01** Strictly inside range (SFR-R01d, SFR-R01e, SFR-R01f, SFR-R07a, SFR-R10; boundary: supplied-input stage)
  - Inputs: T raw(13, 0); returned distance 11; counter 3; `aip4`=12; pending unit AI events [2, 8, 2].
  - Seeds: S={1, 0x29A}; T unchanged.
  - Draws: S modulus 100 ->51; new S={0x6AC6935F, 0}.
  - Trace: s=2; reduced(11, 0); distance 11<12; 51<100.
  - Outputs: counter 2; pending events [8]; no state request (state 12 inactive); ordered request subsequence budget(false), selected-instance request for Skill1 20, search, budget(true), remove AI events of type 2, select(skill 20, target T, point (0, 0)), mode request with value 9 (one shot); no delay. Owner/room requests inside each budget call are omitted from this subsequence.
  - Discriminates: Exact distance 11 and one shot reject premature raw-distance rejection; the retained type 8 event rejects removing every pending event, and the second type 2 event rejects removing only the first; the absent state request rejects an unconditional state 12 removal.
  - runtime_observed: no

- **SFR-V02** Range equality (SFR-R01d, SFR-R07a, SFR-R10; boundary: supplied-input stage)
  - Inputs: T raw(14, 0); returned distance 12.
  - Seeds: no draw.
  - Draws: none.
  - Trace: reduced(12, 0); 12>=12.
  - Outputs: counter 3; schedule due 111; 0 shots.
  - Discriminates: Equality must not fire.
  - runtime_observed: no

- **SFR-V03** Diagonal range equality (SFR-R01d, SFR-R07a, SFR-R10; boundary: supplied-input stage)
  - Inputs: T raw(10, 10); returned distance 12.
  - Seeds: no draw.
  - Draws: none.
  - Trace: reduced(8, 8); (16+8)/2=12.
  - Outputs: counter 3; schedule due 111; 0 shots.
  - Discriminates: Euclidean reduced distance 11.31 would incorrectly fire.
  - runtime_observed: no

- **SFR-V04** Failed chance preserves budget (SFR-R01e, SFR-R06a; boundary: supplied-input stage)
  - Inputs: returned T, distance 1; `aip1`=51.
  - Seeds: S={1, 0x29A}.
  - Draws: S modulus 100 ->51; S={0x6AC6935F, 0}.
  - Trace: 51>=51.
  - Outputs: counter 3; schedule due 107; 0 shots; no decrement call.
  - Discriminates: Failed roll uses `aip2` rather than `aip3` and spends no budget.
  - runtime_observed: no

- **SFR-V05** Finite budget (SFR-R01a, SFR-R01f, SFR-R02; boundary: supplied-input stage)
  - Inputs: counter=-1; `calc4` result 2; two reached in-range successful thinks then one further check.
  - Seeds: S={1, 0x29A}.
  - Draws: S modulus 100 ->51 then 31; final S={0x2F2ED81B, 0x2C890AFD}.
  - Trace: initialize 2; decrement to 1 then 0; next check kills.
  - Outputs: 2 shot requests; counter 0; mode 0 on third think; no third chance draw.
  - Discriminates: Budget is lazy and death occurs at next check.
  - runtime_observed: no

- **SFR-V06** Room cap (SFR-R03c, SFR-R10; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: no draw.
  - Draws: none.
  - Trace: s = 2 ->distances 49 versus 48; first exceeds 48.
  - (a) Inputs: T raw(51, 0), empty slot A. Draws: none. Outputs: reduced distance 49; slot A empty.
  - (b) Inputs: T raw(50, 0), empty slot A. Draws: none. Outputs: reduced distance 48; slot A=T, distance 48.
  - Discriminates: Cap is 48 inclusive and independent of `aip4`.
  - runtime_observed: no

- **SFR-V07** Missing skill on later think (SFR-R01a, SFR-R01b; boundary: supplied-input stage)
  - Inputs: counter 3; selected instance absent.
  - Seeds: no draw.
  - Draws: none.
  - Trace: budget survives; skill lookup fails.
  - Outputs: ordered requests: current-skill clear, then mode 0; current skill absent; 0 searches/shot requests.
  - Discriminates: Later think reaches its own missing-skill path.
  - runtime_observed: no

- **SFR-V08** Clear inferno before search (SFR-R01c, SFR-R01d; boundary: supplied-input stage)
  - Inputs: state 12 active; search returns no target.
  - Seeds: no draw.
  - Draws: none.
  - Trace: remove state 12 before search; no target requests `aip3`.
  - Outputs: state 12 inactive; state removal before search observation, then schedule due 111; counter 3.
  - (a) Inputs: independent think; state 12 active; counter 0. Draws: none. Outputs: R02 death request (current-skill clear, then mode 0); state 12 remains active; no removal request, search or chance draw. Removing state 12 before the budget check would leave it inactive.
  - Discriminates: The search operation observes state 12 inactive; a search observing it active fails; (a) places the removal after the budget check.
  - runtime_observed: no

- **SFR-V09** Stored shortcut (SFR-R03a; boundary: supplied-input stage)
  - Inputs: stored result=T, distance 20; nearby candidate B distance 1; wrapper auxiliary reach result 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: stored success skips room callback; request reach(S, T, 0).
  - Outputs: returnT, distance 20, auxiliary 0; B seed unchanged; no room callbacks.
  - (a) Inputs: independent wrapper stage; stored resolver returns absent; supplied room/chooser result is no unit from empty slotsA/B, each distance 2147483647. Draws: none. Outputs: returned unit absent, distance 2147483647, auxiliary 0; all seeds unchanged. This asserts the verified wrapper's empty-result outputs only; actual enumeration/chooser execution remains Q05.
  - (b) Inputs: independent wrapper stage; stored result=T, distance 20. The supplied reach operation returns 1 for (S, T, 0), 0 for (T, S, 0), and 0 otherwise. Draws: none. Outputs: returned T, distance 20, auxiliary 1; all seeds unchanged.
  - (c) Inputs: independent wrapper stage; stored result=T, distance 60; reach 1 for (S, T, 0). Draws: none. Outputs: returned T, distance 60, auxiliary 1; no room callback; all seeds unchanged. Applying the room cap of 48 to the stored result would return no unit.
  - (d) Inputs: independent wrapper stage; stored resolver returns absent; supplied room/chooser result selects slot A=T, distance 30; reach returns 1 for (S, T, 0) and 0 otherwise. Draws: none. Outputs: returned T, distance 30, auxiliary 1. This asserts the wrapper's auxiliary request on the room-search path only; actual enumeration/chooser execution remains Q05.
  - Discriminates: Stored target wins despite nearer B and beyond active range; caller range is separate.
  - runtime_observed: no

- **SFR-V10** Structural kind and eligibility rejects (SFR-R03e; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: no draw.
  - Draws: none.
  - Trace: reject kind before helper calls in(a); reject eligibility in(b).
  - (a) Inputs: candidate object kind 2. Draws: none. Outputs: validity 0, all seeds unchanged.
  - (b) Inputs: monster candidate eligibility A=false. Draws: none. Outputs: validity 0, all seeds unchanged.
  - (c) Inputs: searcher absent; candidate T otherwise ordinary. Draws: none. Outputs: validity 0; all seeds unchanged.
  - (d) Inputs: candidate absent; searcher S otherwise ordinary. Draws: none. Outputs: validity 0; all seeds unchanged.
  - (e) Inputs: candidate player T has a town room and state 146 active; T seed={1, 0x29A}. Draws: none. Outputs: validity 0; T seed remains{1, 0x29A}; S unchanged.
  - (f) Inputs: searcher S is a monster in mode 0; candidate T is an otherwise valid invisible player with seed{1, 0x29A}. Draws: none. Outputs: validity 0; T seed remains{1, 0x29A}; S unchanged.
  - (g) Inputs: candidate player T in a non-town room with state 146 active, eligibility A=false, seed{1, 0x29A}. Draws: none. Outputs: validity 0; T seed remains{1, 0x29A}; S unchanged.
  - (h) Inputs: candidate is a missile (kind 3); searcher S ordinary. Draws: none. Outputs: validity 0.
  - (i) Inputs: searcher is an object (kind 2); candidate T ordinary. Draws: none. Outputs: validity 0; T seed unchanged.
  - Discriminates: Rejects cannot consume the invis draw; (g) rejects testing eligibility A after the invisible-player draw; (h) rejects a candidate-kind test that excludes only kind 2, and (i) rejects skipping the searcher-kind test.
  - runtime_observed: no

- **SFR-V11** Invisible player below threshold (SFR-R03f; boundary: supplied-input stage)
  - Inputs: T state 146 active. The supplied ordered reach operation returns 0 for (S, T, 0), 1 for (T, S, 0), and 0 for other arguments. Final result is 1 under the preamble's PVP inputs.
  - Seeds: T={1, 0x29A}; S unchanged.
  - Draws: T modulus 100 ->51; T={0x6AC6935F, 0}.
  - Trace: 51<80; reach(S, T, 0)=0 rejects.
  - Outputs: validity 0.
  - (a) Inputs: independent case; T state 146 active, seed{1, 0x29A}; the ordered reach operation returns 1 for (S, T, 0) and 0 for other arguments; final result 1 under the preamble's PVP inputs. Draws: T modulus 100 ->51; T={0x6AC6935F, 0}. Outputs: validity 1. Always rejecting an invisible player whose roll is below 80 would return 0.
  - Discriminates: Player zero-result rejection distinguishes polarity and seed owner. Swapping the reach arguments uses result 1 and incorrectly accepts final 1; the correct result is validity 0 with the stated stepped T seed. Extra pure reads do not affect acceptance.
  - runtime_observed: no

- **SFR-V12** Invisible player bypass and exact threshold (SFR-R03f; boundary: supplied-input stage)
  - Inputs: T state 146 active; reach result 0; final result 1 under the preamble's PVP inputs; independent cases below.
  - Seeds: per case; S unchanged.
  - Draws: exactly one T modulus-100 draw per case.
  - Trace: (a)84>=80 bypasses reach. (b)low32(0*0x6AC690C5+80)=80, high word 0; 80 mod 100=80, and equality bypasses reach.
  - (a) Inputs: T seed={6, 0x29A}. Draws: T modulus 100 ->84; T={0x80A76738, 2}. Outputs: validity 1; S unchanged.
  - (b) Inputs: T seed={0, 80}. Draws: T modulus 100 ->80; T={80, 0}. Outputs: validity 1; S unchanged.
  - Discriminates: Testing roll>80 instead of roll>=80 would run the supplied zero reach result in(b) and reject T. Both cases accept, preserving their exact stepped seeds.
  - runtime_observed: no

- **SFR-V13** Monster polarity and visible player (SFR-R03f; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: no draw.
  - Draws: none.
  - Trace: monster zero proceeds; visible player skips reach; monster nonzero rejects.
  - (a) Inputs: monster T invisible; the ordered reach operation returns 0 for (S, T, 0), 1 for (T, S, 0) and 0 otherwise. Draws: none. Outputs: final 1, validity 1. Swapped reach arguments would see 1 and reject.
  - (b) Inputs: player T visible, reach 0. Draws: none. Outputs: final 1, validity 1.
  - (c) Inputs: monster T invisible, reach 1. Draws: none. Outputs: validity 0.
  - Discriminates: Opposite monster/player zero handling is asserted explicitly.
  - runtime_observed: no

- **SFR-V14** Octagonal metric (SFR-R04; boundary: supplied-input stage)
  - Inputs: supplied unit(0, 0), points(10, 4), (3, 7), (5, 5), (0, 0), (1, 1).
  - Seeds: no draw.
  - Draws: none.
  - Trace: (20+4)/2=12; (14+3)/2=8; (10+5)/2=7; 0; 3/2=1.
  - Outputs: distances 12, 8, 7, 0, 1.
  - Discriminates: Odd minima truncate and diagonal differs from Euclidean.
  - runtime_observed: no

- **SFR-V15** Odd/even range boundary (SFR-R01d, SFR-R07a, SFR-R10; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: (a)S={1, 0x29A}; (b)no draw.
  - Draws: (a)S modulus 100 ->51; (b)none.
  - Trace: reduced(10, 3) gives 23/2=11; reduced(10, 4) gives 12.
  - (a) Inputs: T raw(12, 5), returned distance 11, S={1, 0x29A}. Draws: S modulus 100 ->51; S={0x6AC6935F, 0}. Outputs: one shot, counter 2, no delay.
  - (b) Inputs: T raw(12, 6), returned distance 12. Draws: none. Outputs: no shot, counter 3, due 111.
  - Discriminates: Exact odd truncation separates boundary.
  - runtime_observed: no

- **SFR-V16** Three-shot sequence (SFR-R01a, SFR-R01f, SFR-R02; boundary: supplied-input stage)
  - Inputs: counter=-1; `calc4`=3; three in-range successful thinks then fourth check.
  - Seeds: S={1, 0x29A}.
  - Draws: S modulus 100 ->51, 31, 12; final S={0x280687C4, 0x13ADFE65}.
  - Trace: counter -1 ->3 ->2 ->1 ->0; fourth check sees 0.
  - Outputs: 3 shots; counter 0; mode 0 on fourth; no fourth chance draw.
  - Discriminates: Reinitializing zero or killing on decrement would fail.
  - runtime_observed: no

- **SFR-V17** Chance equality pair (SFR-R01e, SFR-R06a; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: each S={1, 0x29A}.
  - Draws: each S modulus 100 ->51; new S={0x6AC6935F, 0}.
  - Trace: 51>=51; 51<52.
  - (a) Inputs: `aip1`=51, S={1, 0x29A}. Draws: S modulus 100 ->51; S={0x6AC6935F, 0}. Outputs: due 107, counter 3, no shot.
  - (b) Inputs: `aip1`=52, S={1, 0x29A}. Draws: S modulus 100 ->51; S={0x6AC6935F, 0}. Outputs: one shot, counter 2, no delay.
  - Discriminates: Strict threshold comparison.
  - runtime_observed: no

- **SFR-V18** Stable nearest tie and cap (SFR-R03c, SFR-R10; boundary: supplied-input stage)
  - Inputs: same slot class 14; A raw(22, 0), B(2, 22), C(50, 1), D(51, 2) in order; all pass validity/collision.
  - Seeds: no draw.
  - Draws: none.
  - Trace: reduced distances 20, 20, 48, 49; A writes; B ties; C farther; D capped.
  - Outputs: slot A=A, distance 20; C alone would fill distance 48.
  - Discriminates: First encountered wins; this occupied-slot sequence tests tie stability. It does not independently discriminate the cap because distance 49 cannot beat A20 even without a cap. V06(a) starts with an empty slot and rejects reduced 49; V06(b) independently accepts reduced 48, distinguishing the cap boundary.
  - runtime_observed: no

- **SFR-V19** Candidate full size (SFR-R01d, SFR-R10; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: (a/b)S={1, 0x29A}; (c)no draw.
  - Draws: (a/b)S mod 100 ->51; (c)none.
  - Trace: (a/b)reduced(11, 0); (c)(12, 0).
  - (a) Inputs: player raw(13, 2), returned 11, S={1, 0x29A}. Draws: S modulus 100 ->51; S={0x6AC6935F, 0}. Outputs: reduced(11, 0), one shot, counter 2.
  - (b) Inputs: monster SizeX=3, raw(14, 3), returned 11, S={1, 0x29A}. Draws: S modulus 100 ->51; S={0x6AC6935F, 0}. Outputs: reduced(11, 0), one shot, counter 2.
  - (c) Inputs: monster SizeX=3, raw(15, 3), returned 12. Draws: none. Outputs: reduced(12, 0), no shot, counter 3, due 111.
  - Discriminates: Full size per axis distinguishes half-size.
  - runtime_observed: no

- **SFR-V20** Axis clamp (SFR-R10; boundary: supplied-input stage)
  - Inputs: monster SizeX3, raw(2, 7).
  - Seeds: no draw.
  - Draws: none.
  - Trace: max(2-3, 0)=0; 7-3=4; (8+0)/2=4.
  - Outputs: distance 4.
  - (a) Inputs: monster SizeX=-3, raw(2, 7). Draws: none. Outputs: reduced(5, 10); distance 12.
  - (b) Inputs: monster class 0, MonStats count 1, present MonStats row with MonStatsEx=-1, raw(2, 7). Draws: none. Outputs: no extended row, size 0; distance 8 from reduced(2, 7).
  - (c) Inputs: missile identity 2, Missiles count 2, raw(2, 7). Draws: none. Outputs: no missile record, size 0; distance 8.
  - (d) Inputs: object with supplied object size field 3, raw(2, 7). Draws: none. Outputs: reduced(0, 4); distance 4.
  - (e) Inputs: missile identity 0, Missiles count 1, present record Size=3, raw(2, 7). Draws: none. Outputs: reduced(0, 4); distance 4.
  - (f) Inputs: item, raw(2, 7). Draws: none. Outputs: size 1, reduced(1, 6); distance 6.
  - (g) Inputs: isolated size operation only; unit kind 5. Draws: none. Outputs: size 0; no distance assertion.
  - (h) Inputs: isolated size operation only; unit absent. Draws: none. Outputs: size 0; no distance assertion.
  - (i) Inputs: missile identity 0, Missiles count 1, present record Size=200, raw(2, 7). Draws: none. Outputs: size 200 (u8); reduced(0, 0); distance 0. A signed i8 read (-56) would give reduced(58, 63) and distance 92.
  - (j) Inputs: searcher S a monster with SizeX 5; player candidate raw(13, 0). Draws: none. Outputs: candidate size 2 only; reduced(11, 0); distance 11. Also subtracting S's size would give 6.
  - Discriminates: Negative reduced axis must clamp before metric.
  - runtime_observed: no

- **SFR-V21** Validity draw precedes cap (SFR-R03c, SFR-R03f, SFR-R10; boundary: supplied-input stage)
  - Inputs: invisible player T raw(62, 0), reach 1, final 1; empty slot A.
  - Seeds: T={1, 0x29A}; S unchanged.
  - Draws: T mod 100 ->51; T={0x6AC6935F, 0}.
  - Trace: validity draw 51<80, reach 1 proceeds; distance 60>48.
  - Outputs: T seed advanced; slot A empty.
  - Discriminates: Cap-first code would leave candidate seed unchanged.
  - runtime_observed: no

- **SFR-V22** First-think missing skill (SFR-R01a, SFR-R02; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: no draw.
  - Draws: none.
  - Trace: lazy initialization fails before `calc4`/search. In every case "mode 0" is R02's death request: current-skill clear request, then mode 0 request, in that order.
  - (a) Inputs: counter=-1, Skill1=400, count 400. Draws: none. Outputs: lost 1, mode 0, counter-1, no instance/level/evaluation request, search, chance or shot.
  - (d) Inputs: counter=-1, Skill1=-1, count 400. Draws: none. Outputs: lost 1, mode 0, counter=-1; no instance, level or evaluation request, search, chance draw or shot.
  - (b) Inputs: counter=-1, Skill1=20, instance absent. Draws: none. Outputs: lost 1, mode 0, counter-1, no evaluation/search/chance/shot.
  - (c) Inputs: counter=-1, Skill1=20, skill count 400, selected Skills record absent. Draws: none. Outputs: lost 1, mode 0, counter=-1; no instance, level or expression evaluation request, search, chance draw or shot.
  - Discriminates: First-think failure belongs to budget helper.
  - runtime_observed: no

- **SFR-V23** Second budget failure (SFR-R01f, SFR-R02; boundary: supplied-input stage)
  - Inputs: first budget and range pass; the search is a supplied result, so the only owner lookups are the one inside each R02 call. The supplied owner lookup returns P on its first call (inside the first R02 call) and absent on its second (inside the second R02 call, after chance); events[2, 8].
  - Seeds: S={1, 0x29A}.
  - Draws: S mod 100 ->51; S={0x6AC6935F, 0}.
  - Trace: second budget requests death and returns 1; stop before purge/fire. "mode 0" is R02's death request: current-skill clear request, then mode 0 request for S, in that order.
  - Outputs: mode 0; events[2, 8]; counter 3; no selection/shot; no next-think delay request.
  - Discriminates: Failure return of decrementing call cannot be ignored.
  - runtime_observed: no

- **SFR-V24** Budget zero, negative, owner failures (SFR-R02; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a/b)persist eval then<=0; (f/k) persist a positive result; others fail before evaluation. Each "mode 0" output is R02's death request: current-skill clear request then mode 0 request for the stated unit, in that order. Each case is one direct R02 call with decrement=false unless stated.
  - (a) Inputs: counter=-1, `calc4` result 0. Draws: none. Outputs: lost 1, mode 0, counter 0, one evaluation.
  - (b) Inputs: counter=-1, `calc4` result-2. Draws: none. Outputs: lost 1, mode 0, counter-2, one evaluation.
  - (c) Inputs: owner reference present, owner lookup returns absent, counter 3. Draws: none. Outputs: lost 1, one owner lookup request, no room request, mode 0, counter 3, no evaluation.
  - (d) Inputs: owner room absent, counter 3. Draws: none. Outputs: lost 1, mode 0, counter 3, no evaluation.
  - (e) Inputs: owner room town, counter 3. Draws: none. Outputs: lost 1, mode 0, counter 3, no evaluation.
  - (f) Inputs: counter=-1; for the selected instance, level with bonuses included returns 3 and base level would return 2; `calc4` result 2. Draws: none. Outputs: lost 0; counter 2; ordered request subsequence (owner/room requests omitted): selected-instance request for Skill1 20, one level request with bonuses included (result 3), one evaluation of skill 20's `calc4` with (S, 20, 3); no mode request.
  - (g) Inputs: owner lookup returns absent, counter=-1. Draws: none. Outputs: lost 1, mode 0, counter=-1; no instance, level or evaluation request.
  - (h) Inputs: sentry unit absent, counter 3 (the counter belongs to the supplied AI parameters, not to the unit). Draws: none. Outputs: lost 1; current-skill clear and mode 0 requests issued with the absent unit (only their issue and order are asserted; their handling is Q08); no owner lookup request; counter 3 unchanged.
  - (i) Inputs: S is a player unit (kind 0) supplied with a present owner reference whose lookup would return P in its non-town room, counter 3. Draws: none. Outputs: lost 1; mode 0 for S; no owner lookup request; counter 3.
  - (j) Inputs: S is a monster with no owner reference, counter 3. Draws: none. Outputs: lost 1; mode 0 for S; no owner lookup request; counter 3.
  - (k) Inputs: counter=-5, `calc4` result 2, decrement=true. Draws: none. Outputs: lost 0; one evaluation; counter 1 (2 persisted, then 2-1).
  - (l) Inputs: owner room absent, counter=-1. Draws: none. Outputs: lost 1, mode 0, counter=-1; no instance, level or evaluation request.
  - (m) Inputs: owner room town, counter=-1. Draws: none. Outputs: lost 1, mode 0, counter=-1; no instance, level or evaluation request.
  - (n) Inputs: counter=-1, Skill1=0, count 400, present record and instance, `calc4` result 2. Draws: none. Outputs: lost 0; counter 2; one evaluation of skill 0's `calc4` with (S, 0, 3).
  - (o) Inputs: counter=-1, Skill1=399, count 400, present record and instance, `calc4` result 2. Draws: none. Outputs: lost 0; counter 2; one evaluation of skill 399's `calc4` with (S, 399, 3).
  - Discriminates: Signed nonpositive budget and owner gates. (n) rejects a `<=0` lower bound in place of `<0`; (o) rejects a `>=count-1` upper bound in place of `>=count`. (f) rejects a base-level evaluation (2) and a changed argument order; (g/l/m) reject lazy initialization before any owner, room or town gate; (h/i/j) reject an owner lookup before the unit, kind and reference checks, and (i) rejects an owner lookup issued before the kind check (the relative order of the pure reject tests is not observable and not asserted); (k) rejects initializing only when counter equals -1.
  - runtime_observed: no

- **SFR-V25** Stored target bypasses room cap (SFR-R03a, SFR-R01d; boundary: supplied-input stage)
  - Inputs: storedT reduced 60; returned target distance 60; `aip4`=61; `aip1`=100.
  - Seeds: S={1, 0x29A}.
  - Draws: S mod 100 ->51; S={0x6AC6935F, 0}.
  - Trace: wrapper returnsT without cap; 60<61.
  - Outputs: targetT; one shot; counter 2; no room callbacks.
  - Discriminates: Applying room cap to shortcut incorrectly prevents shot.
  - runtime_observed: no

- **SFR-V26** Two slots and nonzero-collision rejection (SFR-R03c; boundary: supplied-input stage)
  - Inputs: independent callback cases below. Candidate validity is nonzero under the preamble's existing PVP inputs; candidate state 146 is absent.
  - Seeds: no draw; S/T and all named candidate seeds retain their entry pairs.
  - Draws: none.
  - Trace: (a)A/B fill distinct slots; C ties A. (b)The nearer T passes validity, cap and strict improvement, but nonzero collision prevents the slot write.
  - (a) Inputs: empty slots; A player raw(22, 0), distance 20, class 14; B monster SizeX=0, raw(4, 0), distance 4, class 1; C player raw(2, 22), distance 20, class 14; collision 0 for all. Draws: none. Outputs: slotA=A, distance 20; slotB=B, distance 4.
  - (b) Inputs: slotA=A, distance 20; slotB empty, distance 2147483647; T player raw(12, 0), size 2, class 14; the ordered collision operation returns 1 for (T, S, 4), 0 for (S, T, 4), and 0 for other arguments. Draws: none. Outputs: T reduced distance 10; slotA remains A, distance 20; slotB remains empty, distance 2147483647; all seeds unchanged.
  - (c) Inputs: independent case; slotA=A, distance 20; slotB empty, distance 2147483647; T player raw(12, 0), size 2, class 14; ordered collision(T, S, 4)=0, reverse collision(S, T, 4)=1. Draws: none. Outputs: T reduced distance 10; slotA becomes T, distance 10; slotB remains empty, distance 2147483647; all seeds unchanged.
  - (d) Inputs: independent case; empty slots; monster D with slot class 2, SizeX 0, raw(4, 0), collision 0; then monster E with slot class 0, SizeX 0, raw(6, 0), collision 0. Draws: none. Outputs: slotA=D, distance 4; slotB=E, distance 6. A class test of `>2` or `==14` would put D in slot B; a class-0 candidate belongs to slot B.
  - Discriminates: A global nearest-only result loses A in(a); ignoring collision rejection or swapping its unit arguments replaces A with nearer T in(b). Acceptance asserts slots, distances and seeds, without an exact pure-query count.
  - runtime_observed: no

- **SFR-V27** Unavailable unit modes (SFR-R03e; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: no draw.
  - Draws: none.
  - Trace: unavailable candidate rejects before room/state tests.
  - (a) Inputs: player candidate mode 17. Draws: none. Outputs: validity 0.
  - (b) Inputs: monster candidate mode 12. Draws: none. Outputs: validity 0.
  - (c) Inputs: monster candidate mode 0. Draws: none. Outputs: validity 0.
  - (d) Inputs: player candidate mode 1, completion=true. Draws: none. Outputs: validity 0.
  - (e) Inputs: player candidate mode 12. Draws: none. Outputs: validity 1.
  - (f) Inputs: monster candidate mode 17, otherwise valid with the preamble's monster inputs. Draws: none. Outputs: validity 1.
  - (g) Inputs: player candidate mode 0. Draws: none. Outputs: validity 0.
  - Discriminates: Dead modes and completion are separate structural gates; (e/f) reject applying either dead mode to both kinds and (g) rejects skipping mode 0 for players.
  - runtime_observed: no

- **SFR-V28** Visible monster skips reach (SFR-R03f; boundary: supplied-input stage)
  - Inputs: monster T invis=false; reach result 1; final result 1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: state absent ->final directly.
  - Outputs: validity 1.
  - Discriminates: Monster rejection must be gated by invis.
  - runtime_observed: no

- **SFR-V29** Owner roots and raw directional return (SFR-R03h; boundary: integration of PVP-R05a/PVP-R07 with supplied existing lists/lookups)
  - Inputs: per case below. Existing tier storage follows the preamble; player relation nodes are supplied explicitly. Owner lookups return the stated unit or absent. All cases start independently.
  - Seeds: no draw.
  - Draws: none in every case.
  - Trace: resolve the searcher once and candidate repeatedly; equal roots reject; distinct player roots read directional flags AND8; other roots invert sameTierNeutral.
  - (a) Inputs: S owner P; T monster owner P. Draws: none. Outputs: return 0; resolved roots P/P; roots equal before relation/tier decision.
  - (b) Inputs: S owner P; T monster owner Q; P/Q players; P's relation node for Q has flags 8. Draws: none. Outputs: return 8; directional flags for P to Q are 8; reverse flags are 0 and do not affect result.
  - (c) Inputs: S/T different monsters with no owner link; own alignment values 1/1. Draws: none. Outputs: sameTierNeutral=0; final return 1.
  - (d) Inputs: S/T different monsters with no owner link; own alignment values 2/2. Draws: none. Outputs: sameTierNeutral=1; final return 0.
  - (e) Inputs: S monster owner lookup absent; T different monster with no owner; own alignment values 1/1. Draws: none. Outputs: roots remain S/T; final return 1.
  - (f) Inputs: S monster owner M (monster), M owner P; T monster owner N (monster), N owner Q; Q is a monster with no owner link; all supplied lookups succeed; own alignment values M 1, Q 2, P 2; other units' tier inputs follow the preamble. Draws: none. Outputs: searcher root M, candidate root Q; sameTierNeutral(M 1, Q 2)=0; final return 1; no lookup of M's owner. A searcher resolved past M to P gives sameTierNeutral(P 2, Q 2)=1 and final 0.
  - (g) Inputs: S/T different monsters; S has no owner link; T's present owner link resolves successfully to T itself; existing own alignment values 1/1. Draws: none. Outputs: roots remain S/T; sameTierNeutral=0; final return 1. The self-owner replacement stops rather than repeating indefinitely.
  - Discriminates: same-root rejection, raw player result 8, asymmetric depth and missing-owner fallback are exact.
  - runtime_observed: no

- **SFR-V30** Difficulty and signed parameter consumption (SFR-R01d, SFR-R01e, SFR-R06a; boundary: supplied-input stage)
  - Inputs: difficulty 2; `aip1` Normal100/Hell0; `aip2` Hell-3; `aip3` Hell9; `aip4` Hell12; returnedT distance 1.
  - Seeds: S={1, 0x29A}.
  - Draws: S mod 100 ->51; S={0x6AC6935F, 0}.
  - Trace: selectHell; 51>=0; frame 100+(-3)=97.
  - Outputs: counter 3; 0shots; due 97.
  - (a) Inputs: independent think; difficulty 2; `aip4` Normal 12/Hell 8; `aip3` Normal 11/Hell 9; returned T distance 10. Draws: none. Outputs: 10>=8; schedule due 109; counter 3; 0 shots; S seed unchanged. Reading Normal `aip4` 12 would proceed to the chance draw.
  - (b) Inputs: independent think; difficulty 1; `aip1` Normal 100/Nightmare 0; `aip2` Normal 7/Nightmare 5; returned T distance 1; S={1, 0x29A}. Draws: S modulus 100 ->51; S={0x6AC6935F, 0}. Outputs: 51>=0; schedule due 105; counter 3; 0 shots.
  - (c) Inputs: independent think; difficulty 0; `aip4`=-1; returned T distance 1. Draws: none. Outputs: 1>=-1; schedule due 111; counter 3; 0 shots. An unsigned read (65535) would proceed to the chance draw.
  - (d) Inputs: independent think; difficulty 0; `aip1`=-1; returned T distance 1; S={1, 0x29A}. Draws: S modulus 100 ->51; S={0x6AC6935F, 0}. Outputs: 51>=-1; schedule due 107; counter 3; 0 shots. An unsigned read (65535) would fire.
  - Discriminates: Wrong difficulty or clamp-to-one would change delay/result; (a) checks Hell `aip3`/`aip4` selection, (b) Nightmare selection and (c/d) signed `aip4`/`aip1`.
  - runtime_observed: no

- **SFR-V31** Dynamic integer positions (SFR-R07b, SFR-R10; boundary: supplied-input stage)
  - Inputs: supplied path X16.16=0x03E88000, Y=0x03E80000; player target integer(1013, 1000).
  - Seeds: no draw.
  - Draws: none.
  - Trace: integerS=(1000, 1000); raw(13, 0); size 2 ->distance 11.
  - Outputs: position(1000, 1000); distance 11; tile integer 200 by NRF-R10a.
  - Discriminates: Fractional path part does not round integer coordinate upward.
  - runtime_observed: no

- **SFR-V32** Supplied chooser cases (SFR-R08; boundary: supplied-input stage)
  - Inputs: independent cases below, with section defaults.
  - Seeds: no draw.
  - Draws: none.
  - Trace: player/empty/range branches, then the path result, then secondary distance<20. Unless a case states otherwise, the path result value read after restoration is 0 and A's unit is excluded from selector 7. A named selector 7 result (C) is supplied directly as the found unit and distance; its filtering is the gated part of R08.
  - (a) Inputs: player S, A20, B4. Draws: none. Outputs: selectedA, distance 20.
  - (b) Inputs: monster S, Aempty, B4. Draws: none. Outputs: selectedB, distance 4.
  - (c) Inputs: monster S, A20, B6. Draws: none. Outputs: selectedA, distance 20.
  - (d) Inputs: monster S, A20, B4, selector 7 none. Draws: none. Outputs: selectedB, distance 4.
  - (e) Inputs: monster S, A20, B4, selector 7 C19. Draws: none. Outputs: slotA=C19, selectedC, distance 19.
  - (f) Inputs: monster S, A20, B4, selector 7 C20. Draws: none. Outputs: selectedB, distance 4.
  - (g) Inputs: monster S, A20, B5; selector 7 result absent. Draws: none. Outputs: selectedB, distance 5; slotA remains A20.
  - (h) Inputs: monster S, slotA=A, distance 20; slotB absent, distance 2147483647. Draws: none. Outputs: selectedA, distance 20; slotA remains A20.
  - (i) Inputs: monster S, A20, B4; path result value 1. Draws: none. Outputs: selectedA, distance 20; no selector 7 search.
  - (j) Inputs: monster S, A20, B4; path result 0; the selector 7 walk offers only A's unit at reduced distance 10, valid, class 14, collision 0. Draws: none. Outputs: A's unit is excluded, so no unit is found; selectedB, distance 4. Excluding B instead of A would find A at 10 and select A10.
  - (k) Inputs: player S, slotA empty (distance 2147483647), B4. Draws: none. Outputs: selected A: no unit, distance 2147483647; no selector 7 search.
  - (l) Inputs: monster S, A30, B4, selector 7 result C25. Draws: none. Outputs: 25>=20, selectedB, distance 4; slotA remains A30. Comparing against A's distance (25<30) would select C25.
  - (m) Inputs: monster S, A10, B4, selector 7 result C15. Draws: none. Outputs: 15<20, slotA=C15, selectedC, distance 15. Comparing against A's distance (15>=10) would select B4.
  - Discriminates: B distance 5 must enter the secondary-result choice: selector 7 absent selects B5. A wrong B-distance>=5 early branch selects A20 instead. (i) rejects ignoring the path result, (j) rejects excluding B rather than A, and (k) rejects falling back to B for a player searcher. The distance 20 secondary boundary and two slots remain explicit supplied-stage checks.
  - runtime_observed: no

- **SFR-V33** Stored-target draw seed owner (SFR-R09; boundary: supplied-input stage)
  - Inputs: reached selector 3 random branch; S seed{1, 0x29A}; candidate T seed{6, 0x29A}.
  - Seeds: S={1, 0x29A}; T={6, 0x29A}.
  - Draws: S RandMask(2) ->1; new S={0x6AC6935F, 0}.
  - Trace: new lo 0x6AC6935F &1 =1.
  - Outputs: result 1; S advanced; T unchanged{6, 0x29A}; no assertion about later routing.
  - Discriminates: A candidate-seed draw would change T and leave S unchanged; the exact seed/output assertion rejects that owner.
  - runtime_observed: no

### 8.2 Observation vectors

All are pending; these initial recipes do not exhaust every rule. Procedures are [in the companion](../research/spec-evidence/skill-assassin-sentry-firing-and-range.md#observation-procedures).

- **SFR-O01** Measure cast tick to first shot at fixed trap rank. No exact latency is predicted until Q01 is resolved.
- **SFR-O02** Measure the first next-think tick after a shot. Exact mode-completion latency remains Q01.
- **SFR-O03** Compare the matched player raw(13, 0)/raw(14, 0) boundary at `aip4`=12 and monster SizeX3 raw(14, 3)/raw(15, 3) boundary; expect inside/outside per V01/V02/V19 after actual inputs are established.

## 9. Acceptance and dependencies

Partial acceptance reproduces all verified local request/return, arithmetic and direct-seed assertions in V01..V31 and V33, respecting each vector's supplied boundary and mixed-rule gates. V32 is an explicit supplied chooser check for the gated integrated R08; passing it does not promote R08. Full reproduction additionally requires Q01/Q02/Q03/Q05/Q06/Q07/Q08 producer contracts and original-game observations for every behavioral rule. Q04's unknown external position producers do not block supplied dynamic positions. O01, O02, O03 and further branch observations are pending; no partial pass claims complete parity. The launch that follows a shot request is accepted separately under [Assassin sentry missile launch](skill-assassin-sentry-missile-launch.md).

**Depends on:** [PvP relations and owners](combat-pvp-damage-and-kill-credit.md), PVP-R05a/PVP-R07 for the exact final predicate/owner/list contract and PVP-R12 no direct draw, retaining HOS-Q01/PVP-Q04 producer limits; [foundations](core-numeric-and-rng-foundations.md), NRF-R01/NRF-R02 seed stepping/draw forms, NRF-R04a/b direct-site seed owners, NRF-R10a/b coordinates, RLI-R02a/RLI-R02b list stat reads, RLI-R03 state request semantics and RLI-R04a mode boundary; [data fields](core-game-data-fields.md), GDA-R08a/GDA-R09a/GDA-R10/GDA-R11a fields, GDF-R03 states and GDF-R05a/GDF-R06a modes; [calculation expressions](skill-calc-expressions.md), SCX-R11/SCX-R15/SCX-R24 and their producer gates; [leveled skill calculations](skill-leveled-damage-tohit-and-elements.md), SLD-R19 calculation-front-end boundary. These contracts do not supply a fully verified highest-level-instance producer; SCX-Q07/SCD-Q10 and Q08 retain that limit. Owners of the cited GDA, RLI, HOS ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI); [player-hostility-declaration](player-hostility-declaration.md) (HOS).

**Depended on by:** [Death Sentry corpse damage](skill-assassin-death-sentry-corpse-damage.md), SFR-R02 budget, SFR-R03a/SFR-R08 target return and SFR-R10 size-by-kind. Supporting combat citations are not prerequisites for local supplied arithmetic.

## 10. Related specs

- [Assassin sentry missile launch](skill-assassin-sentry-missile-launch.md): the charged-bolt burst and shared speed stage reached after this think's shot request; it consumes no rule of this spec.
- [Blade Sentinel and Fury](skill-assassin-blade-sentinel-and-fury.md): a separate Assassin summon AI; it consumes no rule of this spec.
- [Damage pipeline](combat-damage-pipeline-order.md): later sentry hit damage.
- [Hit recovery](combat-hit-recovery.md): reactions and mode-owned draws.
- [Shield block and avoidance](combat-shield-block-and-avoidance.md): hit avoidance.

## Appendix A. Provenance

Lord of Destruction Expansion 1.13c, Windows x86. The [source manifest](../research/source-manifest.md) identifies D2Game/D2Common original images. Reused static readings retain their 2026-10-12 dates and limits; targeted disassembly, original-image branch/table checks and byte-matched size readings are dated 2026-10-15. The [companion](../research/spec-evidence/skill-assassin-sentry-firing-and-range.md) records exact functions and corrections. No gameplay observation was performed.
