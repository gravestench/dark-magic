# Evidence: monster melee specials

Contract: [monster melee specials](../../specs/skill-monster-melee-specials.md). Remaining evidence: [gaps](../../plans/gaps/skill-monster-melee-specials.md).

## Identity, reuse and method

All new readings dated 2026-10-29 target original LoD Expansion 1.13c Windows x86. Original `LOD113C-D2GAME` was freshly hash-checked against the [manifest](../source-manifest.md): 1138688 bytes, SHA-256 `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`. `LOD113C-D2COMMON` supplies resolved imports under the manifest's identity and shared record-operation evidence. Its loaded runtime correspondence is not inferred from the D2Game file hash.

The matching index was checked before new callback readings; no principal melee, swarm, jump, eat, whip or stomp entry has byte-byte-equivalent C. Available original-code listings were read. D2Game 0xB2FC0 was additionally corroborated by complete original-image disassembly. Imported current-skill/condition/argument field operations have byte-equivalent C, read before assigning their meanings; analyst labels were not treated as identities. Full unrelated asserted-failure paths for absent arguments are outside the supplied-present-node stage.

Compatible shared dispatch, numeric, field and expression evidence is reused with its own dates/limits through [skill dispatch](skill-cast-start-and-do-dispatch.md), [numeric foundations](core-numeric-and-rng-foundations.md), [record operations](core-record-layouts-and-ids.md) and [skill calculations](skill-calc-expressions.md). No prior melee-special companion/observation was found. Private paths, source forms, raw records and investigation commands are preserved separately outside the public repository.

## Root and output records

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | 0x9FD00 / 0xA19A0; tables 0x107338 / 0x1074A8 | none established | Skill start/action dispatch | all | Reused SCD contracts; original table bytes freshly read for listed associations |
| LOD113C-D2GAME | 0xB41B0 / 0xB4660 | none established | Start 42 / action 83 | MML-R04 | Action local body read; start/callee closure remains open |
| LOD113C-D2GAME | 0xB3EB0 / 0xB5600 | none established | Start 50 / action 92 | MML-R05 | Local bodies read, damage/target/launch helpers incomplete |
| LOD113C-D2GAME | 0xB3570 / 0xB6790 | none established | Start 64 / action 109 | MML-R06 | Local bodies read, animation/effect/follow-up helpers unresolved |
| LOD113C-D2GAME | 0xB6380 / 0xB60C0 | none established | Start 47 / action 89 | MML-R07 | Action multi-phase body read; start and path/callee closure incomplete |
| LOD113C-D2GAME | 0xB3890 / 0xB2FC0 | none established | Start 48 / action 90 | MML-R02/MML-R08 | Start body read; complete action instructions and imported field operations support supplied condition/calculation stage |
| LOD113C-D2GAME | 0xB5D40 | none established | Action 108 corpse eating | MML-R03/MML-R09 | Actual interior entry in larger listing located from table; reached numeric stage read through output call/result |
| LOD113C-D2GAME | 0xD9C0 | none established | Action 131 whip | MML-R10 | Body read; class/chance/state/effect semantics incomplete |
| LOD113C-D2GAME | 0xC900 | none established | Action 134 stomp | MML-R11 | Body read; encoded clear and damage/area callees incomplete |
| LOD113C-D2GAME | 0x14D0 | none established | Ordered scaled multiply/divide | MML-R03 | Local instructions read; complete NRF-R05 contract reused for all branches and arithmetic faults |
| LOD113C-D2COMMON | 0x303A0 | 10511 | Current skill | MML-R02 | Byte-equivalent C read; present caster and absent skill-list/current-node cases |
| LOD113C-D2COMMON | 0x4DE40 / 0x4DE60 | 10397 / 10137 | Get/set current-skill condition word | MML-R02 | Byte-equivalent C read; direct valid-node read/store; setter replaces whole word |
| LOD113C-D2COMMON | 0x2F0C0 | 10542 | Store frame-event argument while retaining low sequence byte | MML-R02 | Byte-equivalent C read; field expression truncates argument to 24 bits through shifted u32 write |
| LOD113C-D2COMMON | 0x51BF0 | 10786 | Skill calculation front end | MML-R02/MML-R03 | Resolved import and SCX-R11 reuse; context/column choice established in callers, actual expression integration retains SCX limits |

All descriptions are analyst descriptions. The corpse-eating entry 0xB5D40 is inside a larger listing beginning at 0xB5BD0; the action-table entry, independent return/padding and actual entry instructions establish its separate boundary. It must not inherit the earlier routine's label or scope.

### Swarm action derivation

Action table slot 90 points to 0xB2FC0. Its signed skill-id bounds and null prepared-record branches return 0. Thunk 0xAABA resolves to D2Common ordinal 10511/0x303A0, obtaining current skill; absent current skill returns 0. Thunk 0xA79C resolves to ordinal 10397/0x4DE40, reading the node's condition word. Bit 2 is the neutral alternate-calculation condition. When set, thunk 0xA78A resolves to ordinal 10137/0x4DE60 and stores zero in the entire condition word, then selects skill `calc2` at 0x13C. Otherwise it selects `calc1` at 0x138 and makes no condition write. The chosen reference is passed to ordinal 10786 with caster, invocation skill identity and level, then its returned i32 is forwarded to ordinal 10542 with caster. The callback returns 1.

The argument setter's valid-input expression is `(old & 0xFF) + (u32(value) << 8)`. Thus the separate low-byte sequence value survives; the argument is precisely its low 24 bits, interpreted unsigned. Original offsets are investigation mappings: current-skill node conditions 0x0C; caster combined sequence/event value 0x38. The contract exposes logical fields instead. No unknown flag purpose is inferred from a label.

Start 48 refreshes target via 0xE07A0, obtains path target through thunk 0xA598 and requires a distinct target/path/current skill. It calls path configuration/build thunks 0xAB14, 0xAAFC, 0xAAF0 and predicate 0xAC70; the initial build failure requests an alternative configuration and rebuild. Success stores conditions 1 via ordinal 10137. These imported path operations and the later producer of condition 2 remain MML-Q06; the action's supplied-condition arithmetic does not establish them.

### Corpse health-stage derivation

Action 108 validates identity/record, refreshes a distinct target, requires monster type and target mode 12 and checks thunk 0xA7F6. It obtains a corpse quantity through thunk 0xA3A0, resolves corpse identity/room through thunks 0xA478/0xA44E and calls 0xE20A0 before reading caster health. Those operations are not statically closed here.

The reached numeric stage reads current health via thunk 0xA42A with stat 6/layer 0, caster maximum via 0xA3A0, and only for positive saved corpse capacity evaluates `calc1`. It calls 0x14D0 with EAX=percentage, EDX=corpse capacity, ECX=100: this fixes NRF-R05's operand order, which is not mathematically interchangeable for its divide-first branches. After the returned delta is added with i32 wrap to saved current health, signed comparison floors at raw 1, then caps by saved maximum. Thunk 0xA472 receives caster, stat 6, chosen value and layer 0. The callback returns 1. The numeric stage has no RNG operation except the delegated calculation.

The maximum-health query thunk 0xA3A0 resolves to D2Common ordinal 10913/RVA 0x390F0. Its body loads stat identity 7 and calls 0x38A80; that stat-array helper and actual query/write integration are not fully re-read here. The supplied capacity/current/maximum and logged write boundary explicitly avoids promoting that missing integration. Complete stat identity names come from the cited shared field contract, not the query's analyst label.

### Remaining local callees

Fire Hit 83: target refresh 0xE07A0, target getter thunk 0xA598, target/combat operations 0xDED30/0xDA6D0, ordinary melee 0x9FD00-equivalent caller helper at 0x9FD00 is not assumed: the actual body calls D2Game 0x9FD00? The instruction target in this family's listing is symbol L_9ED00, corresponding to RVA 0x9FD00; this overlaps the shared start-dispatch locator and requires original-image call-target adjudication before assigning an ordinary-melee semantic identity. This explicit locator conflict remains MML-Q02, not resolved by labels.

Quick Strike 92 shares target/combat operations and calls launch builder 0xA28F0; a negative skill missile can call monster fallback thunk 0xACB8 with control 1. Start 50 calls 0xDDDE0, with a zeroed damage record and control 0x80; complete damage semantics are gated.

Frenzy 109: predicate 0xB2E00, effect helper 0xB4520 and extra helper 0x28E70 remain unread to completion; current-skill value thunk 0xAD12 selects the extra path. The callback also requests start dispatcher 0x9FD00; complete target/effect paths remain gated. Start 64's distinct-target checks are read, with target-refresh effects not promoted.

Jump 89: path/skill node operations, event removal 0x8E6F0, unit lookup 0xE03A0, contact operation 0xDED30, movement preparation 0xE2240, skill damage helpers 0xDE5A0/0x9F210 and later combat write 0xDDDE0. The listing also includes a following entry after the action return; only dispatch-rooted entry extents are treated as this action.

Whip 131: target-state word at skill 0x82, target predicate 0xDFA90, monster class/base-type helpers 0x18A0/0x1850, calc front end and modulo helper 0x1180 with caster seed. The special branch compares base type 0x1C5. It then chooses helper 0xC6B0 or 0xD850 after chance/state decisions. Named family meaning and output effects remain unresolved.

Stomp 134: prepared missile signed word 0x48 must be positive, not merely nonnegative. Damage setup 0x9F1A0/0x9F210, radius calculation using skill 0x64 and area processor 0xA0C70 are located. The request uses filter 0x8583 and callback 0xA1210; complete filter/callback/target enumeration is MML-Q09.

## T01: named Skills cells

The complete effective-file identity/ranges and observed data-profile limitations are recorded in [teleport T01](skill-monster-teleport.md#t01-effective-skills-cells). The same 2026-10-29 Storm read covered all 173535 bytes of effective `data/global/excel/skills.txt`, SHA-256 `465c28e83657540e0d7ae5b0de826d5657be7fba1aa539a85c3ca3aeea34b498`. Parsing used the complete header, named `skill` row and header-selected columns, with explicit decimal `Id` where used. These text cells have no general compiled/live correspondence claim.

| Named row | Column | Raw cell | Supported claim |
|---|---|---|---|
| Fire Hit | srvstfunc / srvdofunc | `42` / `83` | MML-R01, two named cells |
| Quick Strike | srvstfunc / srvdofunc | `50` / `92` | MML-R01, two named cells |
| MonFrenzy | srvstfunc / srvdofunc | `64` / `109` | MML-R01, two named cells |
| Jump | srvstfunc / srvdofunc | `47` / `89` | MML-R01, two named cells |
| Swarm Move | srvstfunc / srvdofunc | `48` / `90` | MML-R01, two named cells |
| RegurgitatorEat | srvdofunc | `108` | MML-R01 |
| Overseer Whip | srvdofunc | `131` | MML-R01 |
| Siege Beast Stomp | srvdofunc | `134` | MML-R01 |
| Swarm Move | `calc1` / `calc2` | `11` / `19` | Text calculation leads only |

## Runtime and reproduction

2026-10-29: no melee-special callback invocation or gameplay effect was captured. The [observation record](skill-monster-teleport.md#runtime-setup-and-mtp-o01-procedure) supplies no negative gameplay result. Table reads and inconclusive observation attempts do not promote any rule here.

MML-O01: with an original Expansion 1.13c character in Normal difficulty, encounter a swarm-type monster using Swarm Move. Record character class/level/equipment and exact location, monster/current skill and invocation level. Observe start 48 and action 90, capturing current-skill conditions before/after, chosen compiled calculation reference/result and caster frame-event argument before/after. Compare actual captured values with MML-R02, separating path effects from the argument stage. Do not infer alternate-condition production from one ordinary case.

MML-O02: encounter a corpse-eating monster with an eligible monster corpse and a recorded caster health deficit. Record character, difficulty, location, corpse identity/mode, saved corpse maximum, caster current/maximum at each reached read, invocation skill/level and `calc1` result. Observe action 108 through the stat-6 setter request and final callback return. Compare health arithmetic independently of corpse lifecycle/eligibility gates. Preserve exact raw 8.8 values; a health-bar picture is insufficient for rounding/clamp verification. Synthetic extreme-percentage MML-V03 need not occur naturally and remains unobserved unless actual matching inputs exist.

## Review status

This initial source-record entry preceded lint and blind review. Their actual results and the current final applicable review ledger are recorded below.


## Mechanical and self-review record

On 2026-10-29 the three owned spec/companion/gap trios were checked together with the repository spec linter: 0 errors and 0 warnings. Self-review checked source identities, exact per-rule gates, supplied dependency boundaries, vector arithmetic and public/private separation. This mechanical/self-review entry preceded blind review; the final applicable ledger below is the current review status.


## Fresh blind review MML1 — full, 2026-10-29

Counts: MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; GATED 0 new defects (10 documented question gates retained); COVERAGE 1; HARDEST 2. The reader reported no additional gate defects; normalize that report as 0 new gate findings while retaining the contract's documented MML-Q01 to MML-Q10 and its attack/path/corpse/dependency limits. It is not a claim that the spec is ungated. All four original vector results were recomputed successfully.

| Finding | Outcome and rationale |
|---|---|
| COVERAGE: MML-R02 early returns not exercised | Edit: MML-V05 adds negative identity, identity-count equality, explicitly malformed missing record and absent-current-skill cases. The initial node conditions and caster event fields remain unchanged, evaluator log is empty, no local draw/write occurs and result is 0. A stored node can remain unselected in the absent-current case. |
| HARDEST: V03 operand-ordered mulDiv | No change: reader confirms 70001256; swapping the ordered helper operands yields 70000956, which the existing expected writer value rejects. |
| HARDEST: V04 lower bound before maximum cap | No change: reader confirms output write 0; reversed clamp order yields 1 and fails the existing assertion. |
| Documented gates | No promotion: MML-Q01 to MML-Q10 and complete integration limits remain explicit; no gated attack/path/corpse default is introduced. |

The added early-result branch/vector and acceptance list required another full reader; MML2 below is the final applicable review.


## Final applicable blind review MML2 — full, 2026-10-29

Normalized counts: MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; GATED 10 documented questions (Q01-Q10, with SCX compiler/parameter dependency limits also retained); COVERAGE 0 uncovered deficiencies; HARDEST 2. The raw COVERAGE list described five covered vectors, not five missing cases. All recomputed results matched, including the added refusal cases.

| Difficult boundary or retained limit | Outcome and rationale |
|---|---|
| Ordered mulDiv in V03 | No change: expected write 70001256 is correct; the operand-swapped alternative gives 70000956 and is rejected. |
| Lower bound before maximum cap in V04 | No change: expected write 0 is correct; the reversed clamp order gives 1 and is rejected. |
| Existing evidence and dependency gates | No promotion: Q01-Q10, full attack/path/corpse integration and SCX compiler/parameter limits remain explicit. |

The blind-implementer pass is complete for the resulting melee contract under the skill stopping rule. Status remains draft; runtime confirmation, evidence gates and user approval remain separate.
