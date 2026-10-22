# Missile Creation and Seeding

**Description**

When an attack, skill or another game action asks for a projectile, the server first checks its owner and starting place, works out where it aims and how fast it starts, and refuses a request whose aim is too far away. It then creates the projectile, gives it a new number and its own random seed taken from the game. The rest of the projectile's setup continues immediately afterwards; a failure there can leave the already numbered projectile behind.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: MCI-R, MCI-Q, MCI-V, MCI-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/missile-creation-and-seeding.md)
Open evidence gaps: [gap file](../plans/gaps/missile-creation-and-seeding.md)

## 1. Scope

This contract owns the first part of one server missile-creation request: refusals, start room and point, retained aim point, initial speed and its slow/global scaling, the distance refusal, unit allocation and base fields, act association, the unit seed, identity and mode, and the bounded draw accounting of the whole request. It also identifies the subsequent client-notification boundary. Everything after identity within the same request (missile metadata, the creation event, placement and registry insertion, range/lifetime/activation, path setup, callback, damage-setup requests, owner attribution, pierce count, stat requests, the monster-mod event and the request's final return) belongs to [missile-initialization-and-registry](missile-initialization-and-registry.md). Per-frame movement, collision, countdown and explosions belong to [missile-movement-and-collision](missile-movement-and-collision.md); computed damage belongs to [missile-damage-parameters](missile-damage-parameters.md); hit resolution belongs to [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md). Skill callers and message contents have their own contracts. Supplied dependency results permit the verified local stages below; they do not establish complete room-collection or synchronization integration.

## 2. Trigger and ownership

D2Game's server `MISSILES` subsystem runs synchronously when a skill, missile, object or AI action supplies a creation request. One identified route is the skill server do-function table, basic-attack entry, positioned-missile request, then this creator. Each invocation returns at most one missile. It is not a periodic operation and has no independently fixed position within a frame beyond its caller's execution. After MCI-R11 the same invocation continues directly into [MIR-R01a](missile-initialization-and-registry.md); the request's return value after that point is defined there (MIR-R04, MIR-R16). Later missile updates are owned by the movement contract.

D2Common supplies missile records, room operations, state primitives and unit allocation. The creator sends no client message directly. A later synchronization caller can select the missile-creation message; recipient selection and complete delivery remain MCI-Q03. Binary entry points, call chains and original request/layout mappings are in the companion.

## 3. Inputs and data

Runtime integers are i32 unless specified. A bit test uses the u32 request flags. All arithmetic wraps to 32 bits before the next operation unless a rule explicitly widens it. Signed division truncates toward zero. `low32` retains the low 32 bits. Positions are path units; speed is signed 8.8.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| owner, source, targetUnit | request unit references | unit or absent; type/id i32 | Owner must be player or monster; source supplies position; targetUnit supplies tracking aim |
| missileId, record count | request; loaded Missiles collection | i32 | Zero-based missile record index and signed bounds |
| flags | request conditions below | u32 | Select supplied start, aim and speed operations |
| srcX/srcY, tgtX/tgtY | request | i32 path units | Explicit start and absolute target or target offset |
| velOverride, level | request | i32; speed override whole units or 8.8 | Speed override and level-based speed scaling |
| owner/source/target position | unit position operation; NRF-R10b for dynamic paths | unsigned 16-bit integer path components for player/monster/missile; i32 for static positions | Source start and target/owner coincidence tests |
| owner room and candidates | current room and its ordered adjacent-room collection | room or absent; candidate count u32 | Supplied rectangle/candidate data for start-room selection |
| room rectangle / level | room x0,y0,width,height; level identity | i32 | Half-open signed rectangle tests and act association |
| game context | allocation pool, expansion condition, act environments, missile counter, game seed | logical pool/environments; boolean; u32 counter; two u32 seed words | Allocation, base fields, identity and seeding |
| unit allocation result | supplied unit allocation operation | success or absent as supplied | Absent result distinguished from an unchecked dependency fault |
| owner state/list data | state 87; ordered existing stat arrays | RLI-R03/RLI-R02a | State predicate and own-list slow percentage |
| client context | supplied synchronization eligibility/owner lookup | supplied operations/results | Complete production/execution remains MCI-Q03 |

| Request flag | Established local behavior |
|---|---|
| 0x1 | Explicit start point |
| 0x2 / 0x20 | Start-relative / absolute target point; 0x2 wins |
| 0x4 / 0x10 | Speed override / override already in 8.8 |

Other request flags are consumed by [missile-initialization-and-registry](missile-initialization-and-registry.md) section 3.

| Data field | Source / logical width | Consumer |
|---|---|---|
| `Vel`, `VelLev` | `data/global/excel/missiles.txt`, u8 | MCI-R05 |
| `CollideType` | Same file, u8 | MCI-R11 mode; collision-mask use is MIR-R05 |
| `CanSlow`, `ClientSend` | Same file, booleans; GDA-R10 | MCI-R06/R12 |
| `Light`, `RandStart`, `HitShift` | Same file; not consumed directly by the creator | MCI-R13; damage-owned `HitShift` use is MDP |

| Named identity | Read / request | Origin | Binding |
|---|---|---|---|
| state 87 `slowmissiles`; stat 161 `skill_handofathena` | Active-state test; layer-0 own state-list percentage | Named States/ItemStatCost cells in [table record](../research/spec-evidence/missile-creation-and-seeding.md#named-cells-2026-10-15) | Name/id mapping from data; consumer identities fixed in code |
| 75%, 100-unit rejection, high seed word 666, act thresholds | Rules below; NRF-R03a for seed initializer | Engine code/constants | Fixed by original code |
| missile fields and flags | Missiles schema identified by original loader; GDA-R03/GDA-R10 | Table data; flag consumers in code | Data supplies values; code fixes their interpretation |

Named table records establish text-cell identities only within their recorded profile. Synthetic row values below do not assert a shipped missile row or complete text/compiled correspondence.

## 4. Rules (execution order)

**MCI-R01** Request validation [verified in 1.13c code (static)]. Return absent with no allocation, counter write or draw for absent owner, signed missileId < 0 or >= count, owner type outside player/monster, or absent record after lookup.

**MCI-R02** Start point [verified in 1.13c code (static)]. Flag 0x1 selects `(srcX,srcY)`. Otherwise an absent source returns absent without draws; a present source supplies both coordinates. The owner need not be the source.

**MCI-R03** Start room [verified in 1.13c code (static)]. An absent owner room yields absent creation. Test the owner room first with signed half-open bounds `x0 <= X < low32(x0+width)` and `y0 <= Y < low32(y0+height)`. Otherwise visit its supplied adjacent rooms in stored order, skipping absent entries, and choose the first containing point. No match returns absent before allocation/seeding. The candidate getter returns the stored collection/count; it does not sort or construct rooms. Their producer is outside this consumer stage (MCI-Q01).

**MCI-R04** Retained target point [verified in 1.13c code (static)]. If flag 0x2, set point to `low32(start+request offset)` per axis; else if 0x20, use request target; otherwise use start. This retained point is distinct from the target-unit position used for R08's distance test.

**MCI-R05** Initial speed [verified in 1.13c code (static)]. With flag 0x4, use velOverride directly if 0x10, otherwise `low32(velOverride<<8)`. Without 0x4, compute `low32((trunc(i32(VelLev*level)/8)+Vel)<<8)`, wrapping the product and addition separately. Both table velocity fields are unsigned bytes.

**MCI-R06** Slow-missiles scale [verified in 1.13c code (static)]. When CanSlow is set, owner state 87 is active and its state-87 list exists, read layer-0 stat 161 from that list's own array, then set speed to `trunc(i32(value*speed)/100)`. The unit aggregate is not read. Missing state or list leaves speed unchanged. This consumes RLI-R03/RLI-R02a; it does not produce the state/list.

**MCI-R07** Global speed scale [verified in 1.13c code (static)]. If speed != 0, set it to `trunc(i32(speed*75)/100)`. R08, MIR-R05, MIR-R07 and MIR-R10 use this result. Maximum speed is set separately by MIR-R09 and does not receive this percentage.

**MCI-R08** Aim and distance refusal [verified in 1.13c code (static)]. Skip this entire rule when final speed is 0. Otherwise, with no targetUnit and retained point equal to start, increment both retained coordinates with 32-bit wrap. With a targetUnit different from owner whose position equals the owner's position, increment both retained coordinates and drop targetUnit; this test uses owner position, even when source/start differs. Keep targetUnit when it is the owner. For distance only, use the remaining targetUnit's current position when present, otherwise the retained point. Per axis form the wrapped signed delta from start, negate with wrap if negative, then reject if its signed value >= 100. Thus ordinary deltas 99 pass and 100 fail; an INT_MIN delta remains INT_MIN and does not fail this signed comparison. A surviving targetUnit is stored by MIR-R05; its snapshot does not replace the retained point used by MIR-R10.

**MCI-R09** Unit allocation/base fields [verified in 1.13c code (static)]. Request zeroed missile allocation from the game pool. An absent allocator result, if supplied to the creator, returns absent before seeding. The original allocation helper zeroes its allocation without a null check; low-level allocation failure therefore has no established safe null-return fallback. On success set type missile, class missileId, game/pool references, expansion condition equal to the game's expansion condition, creation condition A true, auxiliary reference empty, and act/environment from start-room level. Signed level thresholds select act 0 below 40, 1 below 75, 2 below 103, 3 below 109, 4 below 1024, else 0. These associations do not give condition A a wider meaning (MCI-Q02).

**MCI-R10** Unit seed [verified in 1.13c code (static)]. Consume NRF-R03a/NRF-R03b and NRF-R12a at the reached missile seeder: initialize the child seed to `{1,666}`, step the game seed once, store its new lo as the missile's initial seed copy, then set child seed to `{new game.lo,666}`. The child is not stepped. R01-R09's early refusals never reach this operation; this is not a whole-creation draw guarantee.

**MCI-R11** Identity/mode/conditions [verified in 1.13c code (static)]. Set creation condition B true and creation selector 11. Set id to `low32(game.missileCounter+1)`, replacing exactly 0xFFFFFFFF with 1, then store counter=id. Zero is not separately skipped. Set missile mode to the unsigned CollideType value. Meanings of B and selector 11 remain MCI-Q02.

**MCI-R12** Client notification [unresolved]. Later synchronization reaches a missile sender that requires a valid direct owner link, resolvable owner, valid missile record and ClientSend true before requesting missile-creation message. The creator only requests update marking (MIR-R02a, MIR-R14a). Actual recipient/cadence/eligibility and delivery remain MCI-Q03; no unconditional creation-message assertion follows from the mark.

**MCI-R13** RNG accounting [verified in 1.13c code (static)]. Before the unit seeder, the creator's own validations, room selection, coordinate/speed arithmetic and allocation request make no draws. At the seeder, R10 takes one game-seed step. The creator's other directly verified stages make no draws except MIR-R15a's 0..4 local steps. Callback, complete damage preparation, actual path building and monster-mod handlers retain their own draw budgets/gates (MIR-Q04/Q06/Q08); adding these known steps never proves a whole-creation total. Light/RandStart/HitShift are not directly read by the creator or the specifically checked non-damage setup helpers. Damage-owned HitShift use remains MDP, and unread helper interiors are not covered by that absence claim.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MCI-R01 | Validation | verified in 1.13c code (static) | none | MCI-V01 |
| MCI-R02 | Start | verified in 1.13c code (static) | none | MCI-V02 |
| MCI-R03 | Room selection | verified in 1.13c code (static) | Q01: candidate production only; supplied order available | MCI-V03 |
| MCI-R04 | Retained point | verified in 1.13c code (static) | none | MCI-V02 |
| MCI-R05 | Initial speed | verified in 1.13c code (static) | none | MCI-V04, MIR-V08 |
| MCI-R06 | Own-list slow scale | verified in 1.13c code (static) | none | MCI-V05 |
| MCI-R07 | Global scale | verified in 1.13c code (static) | none | MCI-V04, MCI-V05, MIR-V08 |
| MCI-R08 | Aim/distance | verified in 1.13c code (static) | none | MCI-V04, MCI-V06 |
| MCI-R09 | Allocation/base/act | verified in 1.13c code (static) | none | MCI-V07, MCI-V08, MCI-V09 |
| MCI-R10 | Game step/child seed | verified in 1.13c code (static) | none | MCI-V07 |
| MCI-R11 | Identity/mode | verified in 1.13c code (static) | none | MCI-V07 |
| MCI-R12 | Later client notification | unresolved | Q03: synchronization/delivery | none |
| MCI-R13 | Bounded draw accounting | verified in 1.13c code (static) | MIR-Q04/MIR-Q06/MIR-Q08: aggregate draw total; supplied callback budget external | MCI-V01, MCI-V07, MIR-V13 |

## 5. Outputs and side effects

| Output / effect | Rule |
|---|---|
| Return absent for early refusal or allocation absence; otherwise continue to MIR-R01a, whose stages define the final return | MCI-R01/R02/R03/R08/R09 |
| Unchecked dependency allocation/fault outcomes have no null fallback | MCI-R09 |
| Allocated missile, type/class/game/pool, act/environment, expansion/creation conditions A/B, selector 11, auxiliary empty, mode, id/counter | MCI-R09/R11 |
| Game seed step, missile initialLo and seed writes | MCI-R10 |
| Computed start room/point, retained point, surviving targetUnit and final speed handed to the initialization stage | MCI-R02..R08 |
| Later possible missile-creation message | MCI-R12 |

No save/persistence operation is directly requested. Synchronization effects retain the gate named above.

## 6. Branches and edge cases

Only players/monsters own requests here. Difficulty is not read; expansion changes only the named base condition in this creator. Source and owner may differ. Zero final speed skips nudge and distance refusal; the initialization stage's dependence on zero speed is defined in MIR-R05/R07/R09/R10. A raw override of 1 becomes 0 after scaling.

A targetUnit equal to owner is retained. Owner-position coincidence, rather than start-position coincidence, drops another targetUnit. Distance 99/100 and wrapped INT_MIN differ as R08 specifies. A supplied absent unit allocation returns before the seeder; a dependency assertion or failed unchecked allocation is not that branch. Seed and identity effects stay consumed when a later initialization stage returns absent (MIR-R04).

## 7. Open questions

The [gap file](../plans/gaps/missile-creation-and-seeding.md) records attempted evidence and precise blocked assertions.

- **MCI-Q01** | unresolved; outside-boundary | Production and order of the owner room's adjacent-room collection. Supplied room order supports the local consumer selection. Close with the owning room contract.
- **MCI-Q02** | unresolved; outside-boundary | Wider readers/meaning of the neutral creation conditions (including MIR-R01a's C/D and MIR-R06's destroy condition), auxiliary field and selector 11. Local writes are established; no local assertion depends on assigning wider meanings. Close by tracing their relevant consumers.
- **MCI-Q03** | unresolved; integration | Actual synchronization scheduling, recipient selection, sender eligibility and delivery. Checked sender call-site bytes do not verify the full caller chain. Close through the synchronization/message owner.
- **MCI-Q04** | unresolved; observation | No executable vector has a matching gameplay capture. Source/seed/speed effects are observable; confirmation remains pending, not impossible. No gameplay campaign is part of these records.

## 8. Test vectors

### 8.1 Executable vectors

Each vector is isolated at its named stage unless explicitly stated. All table values are supplied synthetic records. Unlisted request flags are 0; units and records are valid for the named stage. `no draw` means the local stage, excluding an invoked supplied operation's own budget. Supplied operations append ordered arguments to an initially empty call log; allocation supplies the expressly named success/absence, and supplied direct metadata/path writes persist for the vector. Fixtures never silently substitute an absent result for a fatal dependency failure. Each case starts anew and inherits only the vector's shared inputs. All vectors have runtime_observed: no.

- **MCI-V01** Early refusals (MCI-R01/R13; boundary: isolated rule)
  - Inputs: count 10, counter 41; otherwise valid player owner/id 1 and present record.
  - Seeds: game `{1,666}`.
  - Draws: per independent case below.
  - Trace: (a)-(e) reach their first failing validation before allocation/seeder; (f) passes all validation conditions.
  - (a)
    - Inputs: owner absent.
    - Draws: none.
    - Outputs: absent return; counter 41, game seed {1,666}; no allocation/callback.
  - (b)
    - Inputs: missileId -1.
    - Draws: none.
    - Outputs: absent return; counter 41, game seed {1,666}; no allocation/callback.
  - (c)
    - Inputs: missileId 10.
    - Draws: none.
    - Outputs: absent return; counter 41, game seed {1,666}; no allocation/callback.
  - (d)
    - Inputs: owner object.
    - Draws: none.
    - Outputs: absent return; counter 41, game seed {1,666}; no allocation/callback.
  - (e)
    - Inputs: record lookup absent.
    - Draws: none.
    - Outputs: absent return; counter 41, game seed {1,666}; no allocation/callback.
  - (f)
    - Inputs: owner player, missileId 1, count 10, record present, counter 41.
    - Draws: none.
    - Outputs: validation succeeds and control continues to R02; counter 41, game seed {1,666}; allocation/callback logs empty at this validation-stage boundary.
  - Discriminates: accepting equality with count or allocating before owner validation changes return/log. Case (f) rejects unconditional refusal of otherwise valid requests.
  - runtime_observed: no

- **MCI-V02** Start and point precedence (MCI-R02/R04; boundary: isolated rule)
  - Inputs: source position (100,200), request target (5,-3). Stage ends before R08.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: select start, then relative before absolute before start-default.
  - (a)
    - Inputs: flags 0.
    - Draws: none.
    - Outputs: start/point (100,200); nonzero-speed R08 would nudge to (101,201).
  - (b)
    - Inputs: flags 0x3; start (10,20).
    - Draws: none.
    - Outputs: start (10,20), point (15,17).
  - (c)
    - Inputs: flags 0x22.
    - Draws: none.
    - Outputs: start (100,200), point (105,197).
  - (d)
    - Inputs: flag 0x20; target (50,60).
    - Draws: none.
    - Outputs: start (100,200), point (50,60); nonzero-speed R08 would refuse its 140-unit Y distance.
  - (e)
    - Inputs: flags 0; source absent.
    - Draws: none.
    - Outputs: absent return.
  - Discriminates: choosing absolute first makes (c) point (5,-3).
  - runtime_observed: no

- **MCI-V03** Owner room first, ordered neighbors (MCI-R03; boundary: supplied-input stage)
  - Inputs: owner room A rectangle (0,0,10,10); ordered candidates [absent,B,C], B/C both rectangle (10,0,10,10).
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: half-open upper bound excludes 10 from A; visit candidates in given order.
  - (a)
    - Inputs: point (9,9).
    - Draws: none.
    - Outputs: select A.
  - (b)
    - Inputs: point (10,5).
    - Draws: none.
    - Outputs: select B.
  - (c)
    - Inputs: point (20,5).
    - Draws: none.
    - Outputs: absent creation; no allocation/seeder request.
  - (d)
    - Inputs: owner room absent; point (9,9).
    - Draws: none.
    - Outputs: absent creation; no allocation/seeder request.
  - Discriminates: inclusive bound chooses A in (b); sorting/reversing matching candidates may choose C.
  - runtime_observed: no

- **MCI-V04** Signed/wrapped speed and zero gate (MCI-R05/R07/R08; boundary: isolated stages)
  - Inputs: arithmetic stages only; valid records/units as needed.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: (a) trunc(-20/8)=-2; 8<<8=2048;153600/100=1536. (b) 75/100=0. (c) low32(2147483647*75)=2147483573; /100=21474835. (d) negating INT_MIN retains INT_MIN, signed comparison>=100 false; Y0 passes.
  - (a)
    - Inputs: Vel 10/VelLev 4/level -5/no override.
    - Draws: none.
    - Outputs: speed 1536.
  - (b)
    - Inputs: flags 0x14/override 1; start (0,0), point (400,400), targetUnit absent.
    - Draws: none.
    - Outputs: speed 0, R08 skipped; retained point (400,400), no refusal.
  - (c)
    - Inputs: flags 0x14/override 2147483647.
    - Draws: none.
    - Outputs: speed 21474835.
  - (d)
    - Inputs: final speed 1/start (0,0)/point (-2147483648,0)/targetUnit absent.
    - Draws: none.
    - Outputs: distance gate passes; retained point (-2147483648,0).
  - Discriminates: floor negative division, widened product or mathematical absolute value changes output.
  - runtime_observed: no

- **MCI-V05** Own slow list and signed truncation (MCI-R06/R07; boundary: supplied-input stage)
  - Inputs: initial speed 3072, CanSlow true; unit aggregate stat 161=90.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: (a) 3072*50/100=1536, *75/100=1152. (b) -30720/100=-307, -23025/100=-230. (c)/(d) skip first scale, 3072*75/100=2304. (e) 0 remains 0.
  - (a)
    - Inputs: active state 87, matching own list stat 161=50.
    - Draws: none.
    - Outputs: speed 1152.
  - (b)
    - Inputs: active state 87, matching own list stat 161=-10.
    - Draws: none.
    - Outputs: speed -230.
  - (c)
    - Inputs: state absent.
    - Draws: none.
    - Outputs: speed 2304.
  - (d)
    - Inputs: state active but matching list absent.
    - Draws: none.
    - Outputs: speed 2304.
  - (e)
    - Inputs: state active, matching own list value 0.
    - Draws: none.
    - Outputs: speed 0.
  - Discriminates: aggregate value 90 or floor(-307.2) fails (a)/(b).
  - runtime_observed: no

- **MCI-V06** Aim/gate boundaries (MCI-R08, with MIR-R05's target selection; boundary: supplied-input stage)
  - Inputs: final speed 2304, start (100,100), owner position (100,100), retained point (100,100), absent targetUnit; stage includes target selection of MIR-R05 only, not its collision operations.
  - Seeds: game `{1,666}`; this stage never reaches seeder.
  - Draws: per independent case below.
  - Trace: nudge absent/owner-coincident target, then compare the selected gate coordinates per axis.
  - (a)
    - Inputs: shared defaults.
    - Draws: none.
    - Outputs: point target (101,101), no tracking, pass.
  - (b)
    - Inputs: target T different from owner at (150,100).
    - Draws: none.
    - Outputs: tracking T, retained point (100,100), gate delta (50,0), pass.
  - (c)
    - Inputs: target T at (200,100).
    - Draws: none.
    - Outputs: absent return, game seed {1,666} unchanged.
  - (d)
    - Inputs: point (120,130), T different from owner at (100,100).
    - Draws: none.
    - Outputs: drop T; point target (121,131), pass.
  - (e)
    - Inputs: speed 0, T at (400,400).
    - Draws: none.
    - Outputs: tracking T, retained point (100,100), no nudge/gate, pass.
  - (f)
    - Inputs: T at (199,100).
    - Draws: none.
    - Outputs: tracking T, retained point (100,100), delta (99,0), pass.
  - (g)
    - Inputs: start (90,90), point (120,130), T different from owner at (100,100).
    - Draws: none.
    - Outputs: drop T; point target (121,131), pass.
  - (h)
    - Inputs: targetUnit is owner at (100,100).
    - Draws: none.
    - Outputs: keep tracking owner, retained point (100,100), pass.
  - (i)
    - Inputs: speed 2304, start (100,100), retained point (200,100), targetUnit absent; owner at (100,100).
    - Draws: none.
    - Outputs: no nudge; gate delta (100,0); absent return, game seed {1,666} unchanged.
  - Discriminates: 100-inclusive distance, start-position coincidence or dropping the owner changes (c)/(g)/(h). Case (i) rejects a distance gate incorrectly restricted to tracking-unit targets.
  - runtime_observed: no

- **MCI-V07** Allocation, seeding and identity (MCI-R09/R10/R11/R13, with MIR-R01a's fresh metadata fields; boundary: supplied-input stage)
  - Inputs: successful zeroed unit/metadata allocation, expansion true, CollideType 3, room level 80, no later stages beyond MIR-R01a's metadata/id writes.
  - Seeds: game `{1,666}`, child initially zeroed.
  - Draws: per independent case below.
  - Trace: child first `{1,666}`; game product plus 666 gives 1791398751; final child `{1791398751,666}`; room 80 selects act 2; counter increment skips exactly all-ones.
  - (a)
    - Inputs: counter 41.
    - Draws: one game step to {0x6AC6935F,0}; no child step.
    - Outputs: game {0x6AC6935F,0}; child {0x6AC6935F,666}, initialLo 0x6AC6935F; type missile, id/counter 42, mode 3, act 2, expansion/A/B true, selector 11, auxiliary empty, C/D false, source id -1.
  - (b)
    - Inputs: counter 0xFFFFFFFE.
    - Draws: one game step to {0x6AC6935F,0}; no child step.
    - Outputs: game {0x6AC6935F,0}; child {0x6AC6935F,666}, initialLo 0x6AC6935F; type missile, id/counter 1, mode 3, act 2, expansion/A/B true, selector 11, auxiliary empty, C/D false, source id -1.
  - (c)
    - Inputs: counter 0xFFFFFFFF.
    - Draws: one game step to {0x6AC6935F,0}; no child step.
    - Outputs: game {0x6AC6935F,0}; child {0x6AC6935F,666}, initialLo 0x6AC6935F; type missile, id/counter 0, mode 3, act 2, expansion/A/B true, selector 11, auxiliary empty, C/D false, source id -1.
  - (d)
    - Inputs: counter 41, CollideType 8; shared successful allocations and room level 80.
    - Draws: one game step to {0x6AC6935F,0}; no child step.
    - Outputs: game {0x6AC6935F,0}; child {0x6AC6935F,666}, initialLo 0x6AC6935F; type missile, id/counter 42, mode 8, act 2, expansion/A/B true, selector 11, auxiliary empty, C/D false, source id -1.
  - Discriminates: stepping child, seeding before successful allocation or skipping zero changes exact outputs. Case (d) rejects a fixed mode 3 while retaining the existing counter boundary cases. C/D false and source id -1 are MIR-R01a assertions.
  - runtime_observed: no

- **MCI-V08** Allocation failure seed boundary (MCI-R09; boundary: supplied-input stage)
  - Inputs: all validations/room/aim pass, unit allocation returns absent, game counter 41.
  - Seeds: game `{1,666}`.
  - Draws: none.
  - Trace: a supplied absent allocator return reaches the creator guard before base writes/seeder; it does not verify safe null-return behavior from the unchecked original allocator.
  - Outputs: absent return, counter 41/seed `{1,666}`, no metadata/event/path/callback requests.
  - Discriminates: seeding or incrementing before allocation changes seed/counter.
  - runtime_observed: no

- **MCI-V09** Act-selection boundaries (MCI-R09; boundary: isolated act-selection stage)
  - Inputs: supplied room level and corresponding game act environments E0..E4. This stage asserts only the act/environment association, with successful base allocation supplied by its caller.
  - Seeds: no draw.
  - Draws: none.
  - Trace: test signed level against 40,75,103,109,1024 in order; select the matching act and that same-index environment.
  - (a)
    - Inputs: room level 39.
    - Draws: none.
    - Outputs: act 0, environment E0.
  - (b)
    - Inputs: room level 40.
    - Draws: none.
    - Outputs: act 1, environment E1.
  - (c)
    - Inputs: room level 74.
    - Draws: none.
    - Outputs: act 1, environment E1.
  - (d)
    - Inputs: room level 75.
    - Draws: none.
    - Outputs: act 2, environment E2.
  - (e)
    - Inputs: room level 102.
    - Draws: none.
    - Outputs: act 2, environment E2.
  - (f)
    - Inputs: room level 103.
    - Draws: none.
    - Outputs: act 3, environment E3.
  - (g)
    - Inputs: room level 108.
    - Draws: none.
    - Outputs: act 3, environment E3.
  - (h)
    - Inputs: room level 109.
    - Draws: none.
    - Outputs: act 4, environment E4.
  - (i)
    - Inputs: room level 1023.
    - Draws: none.
    - Outputs: act 4, environment E4.
  - (j)
    - Inputs: room level 1024.
    - Draws: none.
    - Outputs: act 0, environment E0.
  - (k)
    - Inputs: room level -1.
    - Draws: none.
    - Outputs: act 0, environment E0.
  - Discriminates: every threshold's strict comparison and the high-level return to act 0; clamping all high levels to act 4 fails case (j).
  - runtime_observed: no

### 8.2 Observation vectors

- **MCI-O01** Initial speed (MCI-R05/R07, with MIR-R09's maximum/acceleration): capture one returned missile's supplied row and level plus final path values; expected arithmetic is MIR-V08, with MCI-R06 when applicable. [Procedure](../research/spec-evidence/missile-creation-and-seeding.md#mci-o01).
- **MCI-O02** Seeder boundary (MCI-R10/R13): record game seed immediately around the seeder and initial child writes, then account separately for downstream draws. Expect NRF-R03a's one game step and unstepped child seed. [Procedure](../research/spec-evidence/missile-creation-and-seeding.md#mci-o02).
- **MCI-O03** Distance rejection (MCI-R08): find a naturally reachable request with a 100-axis delta and compare 99 under equal relevant inputs. Expect early refusal at 100 and acceptance through this gate at 99. [Procedure](../research/spec-evidence/missile-creation-and-seeding.md#mci-o03).

## 9. Acceptance and dependencies

Partial acceptance consists of MCI-V01..V09's exact local arithmetic, guards, fields, returns and seeds at their supplied-input boundaries, including the explicit unchecked-failure distinction. Verified-rule rows name their vectors; synchronization effects are not inferred from these passing stages. Full behavior reproduction of a creation request additionally requires [missile-initialization-and-registry](missile-initialization-and-registry.md)'s acceptance for the stages after identity, MCI-Q03's in-scope integration and applicable original-game observations, including MCI-O01/O02. MCI-Q01/Q02 are upstream/wider meaning questions and impose no local supplied-input gate; MCI-Q04 blocks runtime-confirmation claims. A partial pass never establishes complete creation parity.

**Depends on:** [foundations](core-numeric-and-rng-foundations.md) NRF-R03a/NRF-R03b/NRF-R12a for exact seeds, NRF-R06a for percent format, NRF-R10b for dynamic coordinates and RLI-R03/RLI-R02a for existing state/stat reads; [game data](core-game-data-fields.md) GDA-R03/GDA-R10/GDA-R12/GDF-R02/GDF-R03 for schemas and identity conventions; [missile initialization and registry](missile-initialization-and-registry.md) MIR-R01a..R17b for the request's remaining stages and final return. Wider room producers require their named open integration contracts. Owners of the cited GDA, RLI ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** [missile initialization and registry](missile-initialization-and-registry.md) for the allocated, seeded and numbered missile, its start room/point, retained point, surviving targetUnit and final speed; [missile movement](missile-movement-and-collision.md) for initialized speed; creating skill contracts including [Blade skills](skill-assassin-blade-sentinel-and-fury.md). These consumers do not establish the upstream operations here.

## 10. Related specs

- [missile-initialization-and-registry](missile-initialization-and-registry.md): the same request's stages after identity and its final return.
- [missile-movement-and-collision](missile-movement-and-collision.md): later planner/movement/countdown/collision ownership.
- [missile-damage-parameters](missile-damage-parameters.md): damage initialization, element selectors and damage-owned draws.
- [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md): separate movement-state operations; it does not define this slow-list producer.
- [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md): concrete missile caller and caller-owned damage/RNG boundaries.

## Appendix A. Provenance

Original LoD Expansion 1.13c Windows x86, identified by [source manifest](../research/source-manifest.md) artifacts LOD113C-D2GAME and LOD113C-D2COMMON. Compatible 2026-10-12 static readings are retained at their recorded boundaries. Targeted 2026-10-15 source readings and original-image decoding establish room selection, distance arithmetic and local refusal distinctions; named text-cell records reuse complete 2026-10-15 loader payloads with profile limits. No gameplay vector is confirmed. Detailed original artifact/function locators, corrections, evidence assessment, observation procedures and review status are in the [companion](../research/spec-evidence/missile-creation-and-seeding.md); remaining evidence is in the [gap file](../plans/gaps/missile-creation-and-seeding.md).
