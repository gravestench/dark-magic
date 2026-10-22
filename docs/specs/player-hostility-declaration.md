# Player hostility declaration

**Description**

A player can ask to declare hostility toward another player in the same game. The request is accepted only while the declaring player stands in town and both characters are at least level 9, and after a declaration, declaring again within about one minute is refused with a chat message. The setting the request controls is the one checked when that player's attacks try to hurt the other character, but the full effects of switching hostility on or off, such as leaving a shared party and the notices other players see, are still being confirmed.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: HOS-R, HOS-Q, HOS-V, HOS-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/player-hostility-declaration.md)
Open evidence gaps: [gap file](../plans/gaps/player-hostility-declaration.md)

## 1. Scope

This specification owns the server's hostility-declaration request: the acting and target roles passed by the player-relation command, the acting player's room and town gate, both players' level gates, the comparison with the existing directional relation, the repeated target-level guard, the re-declaration timer and its timestamp request, the selection of the enable or disable branch, and the meaning of the relation eligibility flag that the request controls. The hit-time consumption of that flag (the hostility predicate and the early damage gates) belongs to PVP-R05a and PVP-R06 in [PvP damage, hostility and kill credit](combat-pvp-damage-and-kill-credit.md). Party experience belongs to [party experience](player-party-experience-sharing.md). Stored enable/disable effects, party/skill cleanup, message delivery, the wire format and the complete party protocol are outside the verified slice (HOS-Q01).

## 2. Trigger and ownership

D2Game computes the declaration decisions. D2Common supplies the room getter, town predicate, player data and the stat-list readers described below. Declaration checks are reached by a server player-relation command selecting relation operation 4, with an acting player, a looked-up target player, and a requested relation value. They are request-driven, not a per-frame timer. The companion records the original dispatch/call chains. Only named requests are covered for relation notifications and chat; client delivery and the stored effects of the enable/disable branches are not established here. The relation eligibility flag the enable branch controls is read at hit time by PVP-R05a.

## 3. Inputs and data

All operation names are illustrative. Integer arithmetic uses the foundations' i32/u32, wrapping and truncation conventions. Boolean predicates consume their documented nonzero result unless an exact raw result is stated. Units and lists are logical objects, not required memory layouts.

| Input | Semantic source | Width/format | Meaning |
|---|---|---|---|
| players | supplied acting player and looked-up target player | identity; type/id i32 | Player type 0 |
| relations | acting player's ordered relation nodes: other unit id, flags, timestamp; player data | id i32, flags/timestamp u32 | Directional match; first matching node; missing match gives zero |
| declaration | acting/target own-array stat 12 level, target effective stat 12 level (RLI-R02a/RLI-R02b existing arrays), acting room/town result, requested relation value, existing relation raw result, stored timestamp and current tick samples | levels i32; request/timestamps u32 | Request 0 disables; nonzero enables. Existing eligibility result is raw 0 or 8, not normalized to 1. Timestamp comparison is unsigned, not elapsed-time subtraction |
| dependency operations | supplied relation-node creation, timestamp access, refusal chat, notifications, activation and party/skill cleanup | named results and ordered call log | Stored effects must be explicitly supplied; local call routing does not verify their implementations |

Implementation-relevant identities and constants:

| Mapping/constant | Origin | Binding |
|---|---|---|
| Stat 12 level | table data: effective data/global/excel/itemstatcost.txt columns ID/Stat, GDF-R02 | Original consumers select the stated numeric identity in code |
| Relation eligibility raw value 8; relation operation 4 | engine code | Fixed by original consumers |
| Level 9; delay 60000 ms | engine code | Fixed by original consumers |

Original bindings describe the evidence profile; support for custom data remapping is a separate capability.

## 4. Rules

**HOS-R01. Command roles and complete dispatch** [unresolved]. The recorded command handler passes its acting player and looked-up target to the declaration entry; that establishes HOS-R02's roles. Complete command validation, every relation operation and every write producer remain HOS-Q01. No wire-format contract or complete party protocol is defined here.

**HOS-R02. Declaration checks** [verified in 1.13c code (static)]. At the relation-operation entry, require the acting player's room to exist and be a town, and both acting/target initial own-array levels >=9, or write rejection=1 and return. These initial stat 12/layer 0 reads select ownStats even when usesResolvedStats is true, then apply RLI-R02b; absent unit/list or invalid/missing definition yields 0 under the stated getter guards. With those passed, request relation-node availability in both directions (creation effects supplied). Read the acting player's current directional eligibility raw result (node flags AND 8) for the target. If equal to requestedRelation, request relation notifications in both directions and return. If it is zero, read the target's effective stat 12/layer 0 through RLI-R02a/RLI-R02b: resolvedStats when usesResolvedStats is true, otherwise ownStats. A value <9 returns without changing the rejection output. These are different array-selection operations; their stored inputs can differ without a getter changing during the request. A nonzero enabling request reads the stored directional timestamp and currentTick: if stored > currentTick unsigned, request the refusal chat and both notifications, without a new timestamp or activation. Otherwise sample ticks again and request storing low32(secondTick+60000), then enter HOS-R03. A zero request enters the disable branch of HOS-R03. These checks consume no random value. Declaring is distinct from PVP-R06's hit-time gates.

**HOS-R03. Declaration side effects** [unresolved]. Accepted enable/disable branches request relation changes, notifications and party/skill cleanup. Full stored relation changes, removal effects and message delivery are integration HOS-Q01. HOS-R02's local guard, refusal and timestamp-request assertions remain available with explicit dependency results; they do not establish these integrated effects.

**HOS-R04. Relation meaning** [recovered; unverified]. The relation eligibility flag is interpreted as the player's hostile declaration. Its actual consumption by PVP-R05a is verified; complete declaration/update/UI semantics remain HOS-Q01. This name does not add a symmetric relation requirement.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| HOS-R01 | Complete command integration | unresolved | HOS-Q01 full command/write producers | none |
| HOS-R02 | Declaration guard/timestamp requests | verified in 1.13c code (static) | HOS-Q01 stored node/activation integration | HOS-V01, HOS-V02, HOS-V03 |
| HOS-R03 | Stored declaration side effects | unresolved | HOS-Q01 complete enable/disable effects | none |
| HOS-R04 | Complete hostile-declaration meaning | recovered; unverified | HOS-Q01 integration/name interpretation | none |

## 5. Outputs and side effects

| Output/write/request | Rule |
|---|---|
| Rejection output=1 on initial declaration rejection; relation-node, refusal-chat, notification and timestamp-store requests | HOS-R02; stored enable/disable effects HOS-R03/HOS-Q01 |
| Enter enable or disable branch | HOS-R02; its stored effects HOS-R03/HOS-Q01 |
| No direct seed mutation in the declaration checks | HOS-R02 |

Stored relation activation, party/skill cleanup and client delivery require their owning contracts and gates. No persistence format or event schedule is introduced by the verified local slice.

## 6. Branches and edge cases

Initial rejection writes 1; the repeated target-level guard and the timer refusal leave the rejection output unchanged. The initial level checks read the own arrays, while the repeated target check reads the effective array, so their inputs can differ. A nonzero existing relation result skips the repeated target-level guard, so disabling an existing relation is not blocked by it. A request equal to the existing raw result (raw 8, not 1) only requests notifications. Declaration tick addition wraps u32 and comparison remains ordinary unsigned comparison across wrap; equality of stored and current ticks permits enabling. Declaring is not a hit-time gate; damage permission is decided by PVP-R05a/PVP-R06.

## 7. Open questions

| Id | Kind/status | Missing evidence and affected boundary |
|---|---|---|
| **HOS-Q01** | integration / open | Complete relation-node creation, enable/disable, party/skill cleanup and command/message behavior from request to stored effects. HOS-R02's acting/target roles and local numeric/room/refusal checks are established; supplied operations do not close this gate |
| **HOS-Q02** | observation / pending | Reproduce HOS-R02's verified assertions on original 1.13c with attributable inputs and measured outputs. No executable vector below is gameplay-observed |

Detailed attempts and remaining limits are in the [gap file](../plans/gaps/player-hostility-declaration.md#open-evidence-gaps).

## 8. Test vectors

### 8.1 Executable vectors

Each vector/case starts afresh. Unless overridden: distinct present players A (acting) and B (target), type 0, level 9; A's room present and not town; rejection initially 0. Declaration fixtures use present player base lists with usesResolvedStats=true. An unqualified level n supplies ownStats[(12,0,n)] and resolvedStats[(12,0,n)]; the default is n=9. An explicit own/effective-level case supplies each array separately. No dependency operation mutates these arrays within a case. Definitions for cited stats exist with min-adjustment disabled; stat definition count is 359. Player A's eligible node has flags 8, and an ineligible node has flags 0. Unstated node-availability, timestamp, tick, chat, notification, activation and cleanup operations are supplied no-op/log operations with no draws; they record call arguments/order and persist nothing. Actual owned declaration decisions execute rather than being replaced by their expected values. These fixtures do not execute full declaration changes. All cases have Seeds: no draw; Draws: none; runtime_observed: no. Logical field names express the fixture values; the companion maps them to original binary evidence.

- **HOS-V01** Low-level declaration rejection (HOS-R02; boundary: supplied-input stage)
  - Inputs: Acting level 8, target 9, acting in town; rejection initially 0. All later operations log. Also reverse levels 9/8.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Initial level check fails before relation-node requests.
  - Outputs: rejection=1; no node/timestamp/activation/chat/notification request.
  - Discriminates: Both players must meet level 9.
  - runtime_observed: no.

- **HOS-V02** Declaration timing and request order (HOS-R02; boundary: supplied-input stage)
  - Inputs: Acting player A and target B both level 9; A in town; A's node for B flags 0, requestedRelation=1, rejection initially 0; node availability calls supplied successful/no stored change. Cases (a) stored = 1001, current = 1000; (b) stored = 1000, current = 1000, second tick = 1002; (c) acting room absent; (d) acting in field; (e) A's node flags 8 and requestedRelation=8, so existing raw result 8 equals request 8.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a) unsigned greater comparison refuses; (b) equality passes, second tick+60000=61002; (c/d)initial guard rejects; (e) equal result requests notifications only.
  - Outputs: (a) node requests in both directions, refusal chat, notifications in both directions; rejection 0, no timestamp/activation; (b) node requests in both directions, storeTimestamp(61002), enter enable branch; rejection 0; (c/d)rejection 1, no other calls; (e) node requests in both directions, notifications in both directions, no timestamp.
  - Discriminates: Timestamp equality is allowed; store uses second sample after comparison.
  - (f) Inputs: fresh acting/target/node inputs as case (a), but stored timestamp=2147483649 and current tick=2147483647. Draws: none. Trace: unsigned 2147483649 > 2147483647, so refuse; interpreting the timestamps as signed would reverse this decision. Outputs: node requests in both directions, refusal chat, notifications in both directions; rejection remains 0, no second tick sample, timestamp store or activation request.
  - (g) Inputs: fresh acting/target/node inputs as case (b), stored timestamp=0, first current tick=0, second tick=4294967280. Draws: none. Trace: unsigned equality permits enabling; second tick+60000=4295027280, whose low32 value is 59984. Outputs: node requests in both directions, storeTimestamp(59984), enter enable branch; rejection remains 0. This rejects storing the unbounded sum or clamping instead of wrapping.
  - (h) Inputs: fresh acting/target players, both own/effective levels 9, acting in town; existing directional node flags=8, requestedRelation=0, rejection initially 0. Draws: none. Trace: raw 8 differs from request 0; the existing nonzero result skips the repeat-level guard; a zero request selects the disable branch. Outputs: node-availability requests in both directions, enter disable branch; rejection remains 0; no tick sample or timestamp-store request. Stored disable effects remain HOS-Q01.
  - (i) Inputs: fresh acting player A and target B in the declaration fixture, A in town; A.ownStats=[(12,0,9)] and A.resolvedStats=[(12,0,9)]; B.ownStats=[(12,0,9)] and B.resolvedStats=[(12,0,8)]; both lists usesResolvedStats=true, definitions present/count 359, minimum adjustment disabled; A's directional node flags=0, requestedRelation=1, rejection initially 0. Draws: none. Trace: initial own-array checks read 9/9 and pass; after both node-availability requests, raw eligibility 0 triggers the effective target read, which selects resolved value 8 and returns. Outputs: rejection remains 0; two node-availability requests only, no tick sample, timestamp store, activation, disable, refusal-chat or notification request. All four arrays remain unchanged. This rejects using the initial own-array value for the later effective-level guard.
  - (j) Inputs: fresh acting player A and target B, A in town; A.ownStats=[(12,0,8)] and A.resolvedStats=[(12,0,9)]; B.ownStats=[(12,0,9)] and B.resolvedStats=[(12,0,9)]; usesResolvedStats=true on both, minimum adjustment disabled, definitions present/count 359; requestedRelation=1 and rejection initially 0. Draws: none. Trace: initial actor own-array level is 8, so reject before node requests despite effective level 9. Outputs: rejection=1; no node, tick, timestamp, enable/disable or notification request; arrays unchanged.
  - (k) Inputs: fresh acting/target players, A in town; A own/effective level 9; B.ownStats=[(12,0,9)], B.resolvedStats=[(12,0,8)], usesResolvedStats=true, minimum adjustment disabled and definitions present/count 359. Existing A-to-B node flags=8, requestedRelation=0, rejection initially 0. Draws: none. Trace: initial own-array levels 9/9 pass; nonzero existing raw result 8 skips the repeat effective-level guard even though it would read 8; request 0 enters the disable branch. Outputs: two node-availability requests, enter disable branch, rejection remains 0; no tick or timestamp-store request; arrays unchanged. This rejects applying the zero-existing-result level guard to disabling an existing relation; stored disable effects remain HOS-Q01.
  - runtime_observed: no.

- **HOS-V03** No direct RNG in declaration checks (HOS-R02; boundary: supplied-input stage)
  - Inputs: Execute each HOS-V02 case independently with its exact fixture. Every external operation is the stated no-draw logger; initialize involved game/unit seed pairs {1, 666}.
  - Seeds: no draw.
  - Draws: none.
  - Trace: Only owned declaration decisions and no-draw supplied operations execute.
  - Outputs: All supplied seeds remain {1, 666}; zero direct random-step requests. No claim about the stored enable/disable effects' seeds.
  - Discriminates: Reject adding a local random draw to the declaration checks without asserting unexecuted dependency behavior.
  - runtime_observed: no.

### 8.2 Observation vectors

**HOS-O01. Declaration cross-checks.** Pending original 1.13c observation; procedure in the [companion](../research/spec-evidence/player-hostility-declaration.md#observation-procedure-hos-o01). Use two characters on both sides of level 9, with the declaring player in town and in the field, and declare, re-declare within one minute, and re-declare after one minute. Measure the rejection result, refusal chat, notifications and the resulting directional relation flag. Expected local assertions are HOS-V01 and HOS-V02; stored enable/disable effects remain gated by HOS-Q01. No observation is recorded as matched. A later build cannot confirm a 1.13c rule.

## 9. Acceptance and dependencies

Partial static acceptance comprises the verified assertions of HOS-V01 through HOS-V03 at their named execution boundaries. HOS-V02's accepted enable/disable effects are excluded until HOS-Q01 closes. Exact guard writes, own/effective level selection, raw relation comparison, unsigned timer comparison, wrapped timestamp request, request order and local no-draw assertions are available with explicit inputs. Passing those assertions does not establish complete original-game parity.

Full reproduction also requires HOS-Q01 settled with a source-complete contract and distinguishing executable vectors, plus matching original 1.13c gameplay observations for HOS-R02 (HOS-Q02). Pending observations do not erase static evidence. User approval is separate.

**Depends on:** RLI-R02a/RLI-R02b for own and effective stat-array reads and minimum adjustment; GDF-R02 for the stat 12 identity. Complete implementations of dependencies are not asserted by supplied fixture outcomes. Owners of the cited RLI ids: [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** PVP-R05a, whose directional hostility predicate reads the relation eligibility flag this request controls (HOS-R04). A supporting citation alone does not create an execution prerequisite.

## 10. Related specs

- [PvP damage, hostility and kill credit](combat-pvp-damage-and-kill-credit.md): hit-time hostility predicate and early damage gates that consume the relation flag.
- [Party experience and loot sharing](player-party-experience-sharing.md): party membership that the accepted branches' cleanup affects.
- [Core numeric and RNG foundations](core-numeric-and-rng-foundations.md): stat-array reads and u32 wrapping.

## Appendix A. Provenance

Evidence build is original Lord of Destruction Expansion 1.13c. D2Game/D2Common original specimens are identified by the [source manifest](../research/source-manifest.md). Compatible static readings retain their 2026-10-11/12 dates and limits. Targeted 2026-10-15 disassembly/original-image checks support the declaration handler, command roles and own/effective level sources recorded in the [public companion](../research/spec-evidence/player-hostility-declaration.md). Shared data identities reuse the dated GDF records; no new table-cell or gameplay observation is claimed. No executable vector is runtime-confirmed here.
