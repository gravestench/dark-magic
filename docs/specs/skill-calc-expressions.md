# Skill Calc Expressions

**Description**

Skill formulas turn a skill's level, its listed values, and the character's bonuses into numbers the game can use. These small calculations decide values such as mana costs, effect strength, radius, and how much related skills add, and they can look up other skill values, the character's stats, or another skill's level. Skills and missiles share the same calculator, so every formula follows the same integer rounding rules.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: SCX-R, SCX-Q, SCX-V, SCX-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-calc-expressions.md)
Open evidence gaps: [gap file](../plans/gaps/skill-calc-expressions.md)

## 1. Scope

This contract defines evaluation of compiled calculation expressions: their input bytecode format, stack operations, pool storage, skill, missile and additional-family function bindings, and skill parameter resolution. The boundary starts with a supplied compiled expression or a requested skill parameter and ends with its integer result and any random-seed change. Skill damage and duration endpoints, the to-hit value, and the application of synergy percentages and elemental mastery to them belong to [skill-leveled-damage-tohit-and-elements](skill-leveled-damage-tohit-and-elements.md); this spec's damage, duration and to-hit parameter selectors return those results. Text grammar and precedence, selection of calculation columns by particular skills, the complete missile damage resolver, and server consumption of the result belong to their respective callers. Compiler output properties, missile resolver integration, and the additional expression family remain explicitly gated where evidence is incomplete. The compiled-expression encoding is specified in section 4.1.

## 2. Trigger and ownership

D2Common performs shared skill and missile calculations when a caller requests a named calculation column or parameter, including a damage endpoint's synergy calculation. It selects the expression pool, supplies the unit and skill/missile context, and calls the common Fog expression evaluator. Parameter tokens and function-call tokens synchronously request the corresponding value before evaluation continues. A nested calculation is completed before its caller resumes. These operations have no independent frame cadence; skill execution and tooltip callers determine when they run. No network message is emitted by this boundary. The dispatch chain and return sites are mapped in the companion.

## 3. Inputs and data

All operation names are illustrative. Integer operations use signed 32-bit wrapping arithmetic unless a width or the foundations contract states otherwise; division truncates toward zero. Shift counts use the low five bits. `mulDiv(a,b,c)` means NRF-R05 in the stated argument order. No data row is required by the synthetic vectors.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| expression and declared length | selected compiled calculation pool | bytes; signed length to evaluator | Instructions in section 4.1; declared length may be shorter than readable backing bytes |
| calculation reference | compiled calculation column | unsigned 32-bit byte offset; `0xFFFFFFFF` is empty | Selects an expression within its family pool |
| expression pools | skill, secondary skill/description, missile, additional family | readable byte sequence or absent; unsigned byte count | Independent pools; references are relative to the selected pool |
| parameter operation and function bindings | evaluation environment | optional parameter operation; ordered `(operation, argumentCount)` sequence | Requests and returned i32 values; a supplied operation can record calls |
| skill context | caller | unit or null, skill identity i32, level i32 | Used by skill parameter/function requests |
| missile context | caller | missile unit or null, owner or null, missile identity i32, level i32 | Missile and owner are distinct seed/skill owners |
| additional-family context | caller | two opaque values | Parameter operation returns zero; family identity/functions are SCX-Q05 |
| skill identity and record | loaded Skills definitions | i32 identity/count, record or absent | Bounds and record availability |
| damage, duration and to-hit results | SLD-R08 through SLD-R12 endpoints for the same `(unit,skill,L)` | i32 | Returned by SCX-R25a/SCX-R25b; their Skills fields and arithmetic are SLD inputs |
| raw skill parameters | `Skills.Param1` through `Param8` | i32 | Raw values and four linear/diminishing pairs |
| recursive calculation columns | `Skills.auralencalc`, `calc1` through `calc4`, `aurarangecalc`, `aurastatcalc1` through `aurastatcalc6`, `passivecalc1` through `passivecalc5`, `petmax`, `skpoints` | compiled references | Selected by SCX-R24 |
| mana fields | `Skills.mana`, `lvlmana`, `manashift` | signed 16-bit, signed 16-bit, unsigned 8-bit | Mana progression and scale |
| passive selection inputs | five ordered `Skills.passivestat` slots and aligned `passivecalc` columns | signed 16-bit stat identities; compiled references | First matching slot supplies SCX-R28's expression |
| skill-entry inputs | selected entry or absent, hard level, own-skill condition, bonus levels | level/bonus i32; condition boolean | Supplied selection/bonus stage; full producer integration is SCX-Q07 |
| skill level cap | caller's loaded Experience cap value | i32 | Clamp input; live compiled correspondence is SCX-Q03 |
| unit stats | 12 `level`, 19 `tohit` | i32 | Names/identities consume GDF-R02; code-fixed selection is separate below |
| stat function inputs | stat id, qualifier, stat count, selected read results | i32 | Qualifier routing and special to-hit read |
| missile resolver operation | supplied parameter/damage operation | request plus i32 result | Complete missile computation is outside this local stage; SCX-Q04 |
| missile-slot inputs | selected skill missile identity; missile linear base and increment | signed 16-bit fields | Unresolved slot/name integration; recovered arithmetic in SCX-R27 |
| random seeds | context unit or missile unit | `{lo,hi}` unsigned 32-bit words | Persistent seed state; NRF-R01/NRF-R02 |

| Mapping or constant | Origin | Binding |
|---|---|---|
| opcodes, stack limit 64, function indexes/counts, parameter selectors | engine code | fixed by original code; format and dispatch are section 4 |
| calculation-column reference and pool binding | loader/compiler code | data selects a compiled expression in an engine-selected pool |
| cap source | Experience first record, Amazon field | resolved from loaded data; the text cell `Experience.MaxLvl.Amazon = 99` is [verified data](../research/spec-evidence/skill-calc-expressions.md#cap-data); live compiled correspondence remains SCX-Q03 |
| Skills input names/formats | table schema; section 3's declared inputs and [local field correspondence](../research/spec-evidence/skill-calc-expressions.md#semantic-mapping-of-original-evidence), with GDA-R11a covering its listed subset | Named loaded fields supply the local arithmetic; incomplete missile-slot correspondence remains SCX-Q04 |
| stat names and numeric identities | original-profile ItemStatCost under GDF-R02 and the named-cell record | Names resolved from data; stat consumers and selector cases are fixed in code |
| passive selector stat pairs (342,345), (343,346), (344,347) | engine code | fixed matching identities; no data-name claim for these six ids |
| pool offsets and opaque contexts | original compiled input plus caller inputs | only expression bytes/references are an explicit format boundary; engine memory layouts are evidence details |

## 4. Rules (execution order)

### 4.1 Compiled-expression input format and evaluation

This subsection defines the required bytecode compatibility boundary. Each instruction starts with one unsigned opcode byte. Multi-byte operands are little-endian. The operand rules below are part of the input format; the original evaluator's object layouts and calling conventions are not. The executable vectors supply complete readable operands. For deliberately truncated input, only the checks specified below are established; reading beyond accessible storage has no defined integer result (SCX-Q08).

**SCX-R01** Entry and termination [verified in 1.13c code (static)]. `Eval(code, len, paramFn, funcTable, funcCount, ctx)` returns an `i32`. If `len < 1`: return 0. Otherwise read one opcode byte at a time from `code`. Opcodes 0, 2, 3 and every value above 22 **terminate**: if the stack holds at least one value, return the top value; else return 0. If the read position reaches `len` without a terminating opcode, return **0** regardless of the stack. Compiler terminator production is gated by SCX-R12; callers can supply terminated bytecode directly.

**SCX-R02** Stack [verified in 1.13c code (static)]. 64 `i32` slots and a depth counter starting at 0. A push when depth is 64 is **dropped** (value discarded, depth unchanged, evaluation continues). A pop at depth 0 yields 0 and leaves depth 0. Binary operators pop `b` (top) first, then `a`, and push `a op b`.

**SCX-R03** Literals [verified in 1.13c code (static)]. Opcode 7: one operand byte, sign-extended; 8: two bytes little-endian, sign-extended; 9: four bytes. Before reading, if `(operand position - code start) >= len`, return 0 (only the operand's first byte is bounds-checked). Push the value; advance past the operand.

**SCX-R04** Parameter references [verified in 1.13c code (static)]. Opcode 4: one operand byte, **zero-extended**; 5: two bytes, **sign-extended**; 6: four bytes. If `paramFn` is null, or the operand position check of SCX-R03 fails: return 0 from the whole evaluation. Otherwise push `paramFn(operand, ctx)`.

**SCX-R05** Comparisons [verified in 1.13c code (static)]. Signed `i32` compare, push 1 or 0: 10 `a < b`, 11 `a > b`, 12 `a <= b`, 13 `a >= b`, 14 `a == b`, 15 `a != b`.

**SCX-R06** Arithmetic [verified in 1.13c code (static)]. 16 `a + b`, 17 `a - b`, 18 `a * b` (32-bit wrap, low 32 bits of signed multiply), 19 `a / b` signed `idiv` truncating toward zero; if `b == 0` push **0** and continue (`INT_MIN / -1` is not guarded: hardware fault).

**SCX-R07** Power (opcode 20) [verified in 1.13c code (static)]. `e = pop, base = pop`; if `e <= 0`: push 1; else push `base^e` by repeated 32-bit wrapping multiplication.

**SCX-R08** Negate (opcode 21) [verified in 1.13c code (static)]. Push `-pop` (two's complement; `-INT_MIN = INT_MIN`).

**SCX-R09** Conditional (opcode 22) [verified in 1.13c code (static)].
```
f = pop ; t = pop
if depth > 0: c = pop ; result = (c != 0) ? t : f
else:          result = f                  // no third value: the false branch
push result
```
In source order the condition is pushed first, then the true value, then the false value.

**SCX-R10** Function call (opcode 1) [verified in 1.13c code (static)]. Operand: one byte `k` (unsigned). If `k >= funcCount`: push 0. Else `n = funcTable[k].argCount`:
```
n == 0: r = fn(0, 0, 0, ctx)
n == 1: x1 = pop;                         r = fn(x1, 0, 0, ctx)
n == 2: x2 = pop; x1 = pop;               r = fn(x1, x2, 0, ctx)
n == 3: x3 = pop; x2 = pop; x1 = pop;     r = fn(x1, x2, x3, ctx)
other:  r = 0 (no pops)
push r
```
`x1` is the first argument in source order (deepest on the stack).

### 4.2 Storage and skill functions

**SCX-R11** Calculation references and pool bounds [verified in 1.13c code (static)]. An empty source cell or a compiler result of length at most zero stores `0xFFFFFFFF`; otherwise compiled bytes are appended to the selected pool and the start offset is stored. A front end returns zero without evaluation when its pool is absent or the unsigned reference is not less than the pool byte count. Otherwise it evaluates from that byte offset with declared length equal to the remaining pool bytes. The skill and secondary skill/description front ends both supply `(unit, skill, level)`, the skill parameter operation, and SCX-R13's bindings. Compiler grammar/output validity is SCX-R12; storage behavior with a supplied compiler result is independent of that gap.

**SCX-R12** Compiler output properties [recovered; unverified]. The recovered compiler behavior appends a zero terminator and folds an expression without parameters/functions to one literal and a terminator. Skill identifiers use four-character, space-padded `skillcalc` codes; named functions correspond to SCX-R13. Quoted skill, missile, and stat names use the loaded name tables; stat qualifiers `accr` and `mod` select 1 and 2, others select 0. These producer assertions, literal-width selection, and missile name resolution are SCX-Q01. They do not gate execution of explicitly supplied bytecode.

**SCX-R13** Skill function bindings [verified in 1.13c code (static)]. The seven bindings are index 0 `min` (2 arguments), 1 `max` (2), 2 `rand` (2), 3 `skill` (2), 4 `miss` (2), 5 `stat` (2), and 6 `sklvl` (3). Binding an operation does not verify its separately gated producer integration.

**SCX-R14** Signed minimum and maximum [verified in 1.13c code (static)]. `min(a,b)` returns `a` if `a < b`, otherwise `b`; `max(a,b)` returns `a` if `a > b`, otherwise `b`.

**SCX-R15** Inclusive random range [verified in 1.13c code (static)]. A null context returns zero. With a context, if `a >= b`, return `a` without a draw. Compute `n = low32(b-a+1)` as signed; if `n <= 0`, return `a` without a draw. Otherwise step the context unit's seed once using NRF-R01 and reduce the new low word by NRF-R02: a power-of-two `n` uses a mask, other `n` use unsigned remainder. Return `low32(a+r)`. With a present context, an absent unit faults when the draw is needed; it does not fault on the no-draw branches.

**SCX-R16a** Skill function with supplied entry/bonus/cap [verified in 1.13c code (static)]. For `skill(s,p)`, begin at level zero. A present unit with a selected entry for `s` uses that entry's hard level and, only when the own-skill condition is true, adds its supplied bonus levels. Clamp a present entry's level with `min(max(L,0),cap)`. A missing unit or entry retains zero. Return the skill parameter result for `(unit,p,s,L)`. Cap initialization is SCX-Q03 and entry/bonus production is SCX-Q07; supplied values verify this combination independently.

**SCX-R16b** Skill-entry and bonus production [unresolved]. The skill function requests a selected entry and, for an own-skill entry, bonus levels before applying SCX-R16a. Selecting among duplicate/disabled entries and calculating every bonus source are SCX-Q07; no default producer is established here.

**SCX-R17** Missile function request [verified in 1.13c code (static)]. `miss(m,p)` with null context returns zero. Otherwise request the missile parameter operation with missile unit absent, owner equal to the skill context's unit, selector `p`, missile identity `m`, and level equal to the current skill level. Return that operation's result unchanged. The supplied-operation request/return contract is verified; complete missile resolver results are SCX-Q04.

**SCX-R18a** Stat-function routing [verified in 1.13c code (static)]. `stat(id,q)` returns zero for an absent context/unit, a negative id, or `id >= statCount`. Stat 19 requests the special to-hit operation regardless of `q`. Other ids request layer zero using the own-array list read for `q=1`, the selected-array list read for `q=2`, and the ordinary unit stat read for every other qualifier. These reads consume RLI-R02a/RLI-R02b, including their minimum adjustment; qualifier 2 and the ordinary unit read select the same array for the same list. Return the selected read's result unchanged.

**SCX-R18b** Qualifier and to-hit result integration [unresolved]. Complete production of the special to-hit result remains SCX-Q02. This does not gate SCX-R18a's guards, routing, or forwarding with supplied operation results.

**SCX-R19** Nested level request [verified in 1.13c code (static)]. `sklvl(s,p1,p2)` with null context returns zero. Otherwise first resolve `(unit,p1,currentSkill,currentLevel)` to `L2`, then resolve `(unit,p2,s,L2)` and return the second result unchanged. No level clamp is inserted between the calls.

### 4.3 Skill parameter resolution

**SCX-R20** Parameter dispatch [verified in 1.13c code (static)]. For `(unit,id,skill,L)`, an invalid skill identity, unavailable record, or selector outside 0 through `0x48` returns zero. Otherwise use the cases below. `P1` through `P8` mean `Param1` through `Param8`; `L` is the supplied level. Selector numeric identities are fixed in code; name-to-code resolution is SCX-R12.

**SCX-R21** Linear and raw parameters [verified in 1.13c code (static)]. Ids 0, 2, 4, 6: if `L <= 0`: 0; else `(L-1) * P2 + P1`, `(L-1) * P4 + P3`, `(L-1) * P6 + P5`, `(L-1) * P8 + P7` (i32 wrap). Ids 8..0x0F: `P1..P8` raw. Id 0x10: `L`.

**SCX-R22** Diminishing parameters [verified in 1.13c code (static)]. Ids 1, 3, 5: if `L <= 0`: 0; else `DM(L, P1, P2)`, `DM(L, P3, P4)`, `DM(L, P5, P6)`. Id 7: `DM(L, P7, P8)` **without** the `L <= 0` guard. `DM`:
```
f = (L * 110) / (L + 6)            // i32 idiv, truncating
v = (f * (hi - lo)) / 100          // i32 multiply, signed /100 truncating toward zero
r = lo + v ; if r > hi: r = hi
```
`L = -6` divides by zero (fault); `L = 0` gives `lo`.

**SCX-R23** Mana parameters [verified in 1.13c code (static)]. With `m = lvlmana * (L-1) + mana` (16-bit fields sign-extended) and `L <= 0` giving 0: id 0x15 `(m << manashift) >> 8`; id 0x16 `((m * 25) / 2 << manashift) >> 8` (`/2` truncating); id 0x2A `m << manashift`. `>>` is arithmetic.

**SCX-R24** Calc-column references [verified in 1.13c code (static)]. Ids 0x36..0x48 evaluate another compiled column of the same skill in the skill pool with the same context `(unit, skill, L)`: 0x36 `auralencalc`, 0x37..0x3A `calc1`..`calc4`, 0x3B `aurarangecalc`, 0x3C..0x41 `aurastatcalc1`..`6`, 0x42..0x46 `passivecalc1`..`5`, 0x47 `petmax`, 0x48 `skpoints`. Bounds as SCX-R11 (an empty column gives 0). The nesting depth is not limited by the code.

**SCX-R25a** Damage and duration parameter selection [verified in 1.13c code (static)]. Request the elemental minimum or maximum endpoint (SLD-R08 through SLD-R10) or the elemental duration (SLD-R11) with the same unit, skill, and level, and return as follows. The mastery-disabled selectors pass mastery condition 0 and the mastery-enabled selectors pass 1; the endpoint then applies or omits SLD-R10's mastery addition.

| Selector | Result |
|---|---|
| `0x11`, `0x12` | elemental minimum/maximum with mastery disabled, arithmetic shift right 8 |
| `0x26`, `0x27` | the same minimum/maximum raw values |
| `0x31`, `0x32` | elemental minimum/maximum with mastery enabled, arithmetic shift right 8 |
| `0x34`, `0x35` | mastery-enabled minimum/maximum raw values |
| `0x13`, `0x33` | elemental duration; the extra selector flag does not change the result |

**SCX-R25b** To-hit parameter selection [verified in 1.13c code (static)]. Selector `0x14` requests the skill to-hit value of SLD-R12 for the same unit, skill, and level and returns it unchanged, with no shift or other postprocessing. SLD-R12 owns the arithmetic, including its zero results for an invalid or unavailable skill, `L <= 0`, or an absent pool or invalid nonempty reference, and its linear fallback for an empty `ToHitCalc`.

**SCX-R26** Unit level parameters [verified in 1.13c code (static)]. Selector `0x28` returns zero for an absent unit, otherwise RLI-R02a/RLI-R02b's stat 12 `level` at layer zero. Selector `0x29` returns zero for an absent unit or selected skill entry; otherwise return `min(max(hardLevel,0),cap)`, without adding bonus levels. Supplied-entry behavior is verified; entry production and cap initialization remain SCX-Q07/SCX-Q03.

**SCX-R27** Missile-slot parameters [asm-only]. Selectors `0x1A` through `0x25` and `0x2B` through `0x30` select a signed 16-bit missile identity from the skill and request its damage or duration operation. All return zero at nonpositive level except `0x23` through `0x25`; those three calculate `signed16(missileIncrement)*L + signed16(missileBase)` with wrapping arithmetic. Selectors `0x1A`, `0x1B`, `0x1D`, `0x1E`, `0x20`, `0x21` arithmetic-shift their selected helper result right by 8. Slot naming, the complete per-selector routing, and damage/duration result contracts remain SCX-Q04. No final missile result is accepted from this rule.

**SCX-R28** Passive selector calculations [verified in 1.13c code (static)]. Selectors `0x17`, `0x18`, `0x19` correspond respectively to stat pairs `(342,345)`, `(343,346)`, `(344,347)`. With a valid skill and `L > 0`, inspect the five passive stat slots in order. At the first slot equal to either identity in the pair, evaluate the aligned passive calculation in the skill pool using `(unit,skill,L)` and return its result. No match, invalid skill, nonpositive level, absent pool, or invalid reference returns zero. No later matching slot is evaluated.

### 4.4 Other expression families

**SCX-R29a** Missile calculation front end and requests [verified in 1.13c code (static)]. A present missile unit whose type is not missile (3) returns zero. A negative supplied missile identity uses the missile unit's class identity; a negative level uses its stored signed 16-bit missile level, or zero when missile data is absent. Those fallback requests require a present missile unit. The front end uses the missile pool and SCX-R11's pool check with context `(missile,owner,missileId,level)`. Its four bindings are signed minimum (2 arguments), signed maximum (2), random range (2), and owner skill-level function (2). Missile random range follows SCX-R15 with the missile seed; an absent missile returns zero. The owner skill function combines owner entry/bonus/cap as SCX-R16a. Parameter tokens forward the same missile context to the missile resolver. Complete resolver and owner-entry integration remain gated by SCX-Q04/SCX-Q07.

**SCX-R29b** Missile parameter results [unresolved]. The complete missile resolver's parameter cases, damage helpers, name resolution, and negative-fallback fault cases are SCX-Q04/SCX-Q01/SCX-Q08. Supplied-operation routing in SCX-R29a does not establish these results.

**SCX-R30a** Additional-family pool and zero parameters [verified in 1.13c code (static)]. The additional front end uses its independent pool with the same unsigned reference bound and remaining-length evaluation as SCX-R11, passing a two-value opaque context. Every parameter token resolves to zero. These properties hold independently of which table owns the pool or what its functions compute.

**SCX-R30b** Additional-family owner and functions [unresolved]. The family has four function bindings. Its owning data columns, their trigger, and complete function results are SCX-Q05; no integration result using a function token is accepted here.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| SCX-R01 | Termination | verified in 1.13c code (static) | none | SCX-V01, SCX-V06, SCX-V19 |
| SCX-R02 | Stack limits/underflow | verified in 1.13c code (static) | none | SCX-V07, SCX-V08 |
| SCX-R03 | Literal encoding | verified in 1.13c code (static) | none | SCX-V01 |
| SCX-R04 | Parameter encoding/request | verified in 1.13c code (static) | none | SCX-V09 |
| SCX-R05 | Signed comparisons | verified in 1.13c code (static) | none | SCX-V10 |
| SCX-R06 | Arithmetic/division | verified in 1.13c code (static) | none | SCX-V02, SCX-V03 |
| SCX-R07 | Power | verified in 1.13c code (static) | none | SCX-V04 |
| SCX-R08 | Negate | verified in 1.13c code (static) | none | SCX-V08 |
| SCX-R09 | Conditional | verified in 1.13c code (static) | none | SCX-V05 |
| SCX-R10 | Function calls | verified in 1.13c code (static) | none | SCX-V11, SCX-V20 |
| SCX-R11 | Pool references/storage | verified in 1.13c code (static) | none | SCX-V12 |
| SCX-R12 | Compiler output producer | recovered; unverified | SCX-Q01: compiler-produced output properties only | none |
| SCX-R13 | Skill bindings | verified in 1.13c code (static) | none | SCX-V11 |
| SCX-R14 | Minimum/maximum | verified in 1.13c code (static) | none | SCX-V11 |
| SCX-R15 | Random ranges | verified in 1.13c code (static) | none | SCX-V13 |
| SCX-R16a | Supplied-entry level combination | verified in 1.13c code (static) | SCX-Q03/SCX-Q07: live cap/entry/bonus integration only | SCX-V14 |
| SCX-R16b | Entry/bonus production | unresolved | SCX-Q07: entry/bonus production | none |
| SCX-R17 | Missile request forwarding | verified in 1.13c code (static) | SCX-Q04: complete missile result only | SCX-V21 |
| SCX-R18a | Stat routing | verified in 1.13c code (static) | SCX-Q02: special to-hit producer only | SCX-V22 |
| SCX-R18b | Selected getter integration | unresolved | SCX-Q02: special to-hit result integration | none |
| SCX-R19 | Two parameter requests | verified in 1.13c code (static) | none | SCX-V23 |
| SCX-R20 | Parameter dispatch bounds | verified in 1.13c code (static) | none | SCX-V15, SCX-V24 |
| SCX-R21 | Raw/linear parameters | verified in 1.13c code (static) | none | SCX-V15 |
| SCX-R22 | Diminishing parameters | verified in 1.13c code (static) | none | SCX-V16 |
| SCX-R23 | Mana scaling | verified in 1.13c code (static) | none | SCX-V17 |
| SCX-R24 | Recursive columns | verified in 1.13c code (static) | none | SCX-V12, SCX-V25 |
| SCX-R25a | Damage/duration selectors | verified in 1.13c code (static) | none | SCX-V18, SCX-V26 |
| SCX-R25b | To-hit selector | verified in 1.13c code (static) | none | SCX-V27 |
| SCX-R26 | Unit/hard level | verified in 1.13c code (static) | SCX-Q03/SCX-Q07: live hard-level lookup/cap integration only | SCX-V14, SCX-V22 |
| SCX-R27 | Missile slot results | asm-only | SCX-Q04: final missile slot results | none |
| SCX-R28 | First passive match | verified in 1.13c code (static) | none | SCX-V28 |
| SCX-R29a | Missile front end/request | verified in 1.13c code (static) | SCX-Q04/SCX-Q07: produced parameter/entry results only | SCX-V13, SCX-V29 |
| SCX-R29b | Missile result integration | unresolved | SCX-Q04: missile result integration; SCX-Q08: malformed fallback | none |
| SCX-R30a | Additional pool/zero params | verified in 1.13c code (static) | none | SCX-V30 |
| SCX-R30b | Additional functions/owner | unresolved | SCX-Q05: owner/function results | none |

## 5. Outputs and side effects

| Output/effect | Rules |
|---|---|
| Returned i32 calculation or parameter, including forwarded damage, duration and to-hit results; zero/sentinel/fault cases as specified | SCX-R01 through SCX-R30b within their stated gates |
| Ordered parameter/function/dependency requests and their forwarded return values | SCX-R04, SCX-R10, SCX-R16a, SCX-R17, SCX-R18a, SCX-R19, SCX-R24, SCX-R25a, SCX-R25b, SCX-R28, SCX-R29a |
| Compiled bytes appended to the selected pool and reference stored, or empty sentinel stored | SCX-R11 with supplied compiler result |
| Context unit's seed changed once for a needed skill random draw; missile seed changed once for missile random draw | SCX-R15, SCX-R29a |

Evaluation makes no stat writes, creates no unit or event, and sends no message. The temporary evaluation stack is local to one evaluation; recursive calls have separate stacks. Pool construction is a loading effect; ordinary evaluation does not mutate its pool. Returned results can be used by other systems without this contract specifying those systems' effects.

## 6. Branches and edge cases

Empty calculations, absent pools, invalid references, invalid selectors/skills, missing callbacks, and evaluation running off its declared end return zero as specified. Damage, duration and to-hit selectors inherit the invalid-skill and level results of SLD-R01a. A dropped 65th push leaves the earlier stack intact. Conditional branches are already evaluated values; the conditional opcode performs no short circuit.

Unguarded signed division overflow faults. Diminishing level -6 faults only on the unguarded fourth pair. A random request with a present context and absent unit faults only when it needs a draw. Truncated operands and missing function-index bytes have the limits in SCX-Q08; they must not be assigned an invented successful value. Recursive calculations have no depth limit or cycle detection within this code path. No local difficulty or PvP branch is established; those callers supply the inputs.

## 7. Open questions

Each question's attempted evidence and exact blocked boundary is in the [gap file](../plans/gaps/skill-calc-expressions.md).

| Id | Status and boundary | Evidence that would settle it |
|---|---|---|
| **SCX-Q01** | behavior: compiler output, naming, literal width, and missile name resolution remain recovered | Trace compiler and owned encoding/name-resolution helpers through returned output; text precedence remains outside this evaluator's acceptance boundary |
| **SCX-Q02** | integration: special to-hit result | Finish the special to-hit path, including its unit eligibility and class adjustment; qualifier-one/two and ordinary reads consume RLI-R02a/RLI-R02b |
| **SCX-Q03** | integration: live compiled cap correspondence; none locally with a supplied cap | Corroborate the loaded first Experience record against its named text cell; loader binding and the text value are established separately |
| **SCX-Q04** | integration: complete missile resolver/slot results | Complete per-selector routing, slot column naming, and damage/duration dependencies; supplied request/return vectors remain available |
| **SCX-Q05** | behavior: additional family's owner and four function results | Trace the pool writer, its data columns, all four bindings, and their returned values |
| **SCX-Q06** | outside-boundary: other builds; none locally | No other build is evidence for this 1.13c contract |
| **SCX-Q07** | integration: selected skill entry and all bonus sources | Complete lookup/bonus dependencies and establish duplicate/disabled-entry semantics; supplied-entry arithmetic remains available |
| **SCX-Q08** | behavior: malformed backing storage/negative missile fallback with no missile | Establish caller-readable extents or observe the fault boundary; no safe integer result is specified |
| **SCX-Q09** | observation: all verified behavioral rules lack gameplay confirmation | Reproduce named vector assertions on owned matching 1.13c files; pending observations do not gate static supplied-input assertions |

## 8. Test vectors

### 8.1 Executable vectors

Defaults: each case starts independently; the unit is absent unless explicitly supplied, and no stats, states, or pool is read unless named. Hex byte sequences are exact compiled inputs in section 4.1's format; their declared length equals their byte count unless overridden. Each evaluator starts with an empty stack and complete readable operands. Synthetic skill records have valid identity 0 of count 1, all unspecified numeric fields zero, empty optional calculation references, and no Kick flag; damage, duration and to-hit fields named in a case are SLD inputs, used with the SLD endpoint contracts. In a supplied-input arithmetic case, a stated calculation result overrides that column's empty-reference default: use nonempty reference 0 and a supplied evaluation operation that logs the forwarded context and returns the named result without draws. Only cases that explicitly supply pool bytes execute bytecode for that result. All vectors use the 1.13c contract; no other build has been tested. Cases marked supplied-input stage use deterministic operations that record requests, return the named values, take no draw unless stated, and persist only explicit seed/pool changes. A recorded call is an output assertion; supplying a result does not verify its producer.

- **SCX-V01** Literal widths and subtraction (SCX-R01, SCX-R03, SCX-R06; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: decode signed operands; pop right operand first; return stack top at zero terminator
  - Outputs: per case
  - Discriminates: subtraction case rejects reversed operand order; signed word case rejects zero extension
  - runtime_observed: no
  - (a) Inputs: `07 05 07 03 11 00`. Draws: none. Outputs: `5-3 = 2`.
  - (b) Inputs: `08 FF FF 00`. Draws: none. Outputs: -1.
  - (c) Inputs: `09 00 01 00 00 00`. Draws: none. Outputs: 256.

- **SCX-V02** Zero divisor (SCX-R06; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: pop b=0 then a=7; zero-divisor branch pushes zero
  - Outputs: per case
  - Discriminates: rejects returning the numerator, faulting on zero, or terminating the whole calculation at a zero divisor
  - runtime_observed: no
  - (a) Inputs: `07 07 07 00 13 00`. Draws: none. Outputs: 0.
  - (b) Inputs: `07 07 07 00 13 07 05 10 00`. Draws: none. Outputs: 5; the division pushes zero, the next literal pushes 5, and addition returns 5.

- **SCX-V03** Signed truncation and multiplication wrap (SCX-R06; boundary: isolated rule)
  - Inputs: `07 F9 07 02 13 00`
  - Seeds: `no draw`
  - Draws: none
  - Trace: -7 divided by 2 truncates toward zero
  - Outputs: per case
  - Discriminates: rejects floor rounding (-4) and retaining a product beyond 32 bits
  - runtime_observed: no
  - (a) Inputs: `07 F9 07 02 13 00`. Draws: none. Outputs: -3.
  - (b) Inputs: `09 00 00 01 00 09 00 00 01 00 12 00`. Draws: none. Outputs: 0; 65536*65536 wraps modulo 2^32.
  - (c) Inputs: `09 00 00 00 80 07 FF 13 00`. Draws: none. Outputs: signed division fault for -2147483648 / -1; no integer return. The zero-divisor fallback does not cover this overflow.

- **SCX-V04** Power (SCX-R07; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: pop exponent then base; nonpositive exponent returns one
  - Outputs: per case
  - Discriminates: rejects exponent/base reversal and returning zero for exponent zero
  - runtime_observed: no
  - (a) Inputs: `07 02 07 0A 14 00`. Draws: none. Outputs: 1024.
  - (b) Inputs: `07 05 07 00 14 00`. Draws: none. Outputs: 1.
  - (c) Inputs: `07 05 07 FF 14 00`. Draws: none. Outputs: 1.
  - (d) Inputs: `09 00 00 01 00 07 02 14 00`. Draws: none. Outputs: 0; positive exponent 2 multiplies 65536 by itself with low-32-bit wrap.

- **SCX-V05** Conditional and missing condition (SCX-R09; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: pop false 20 then true 10; use third value only if present
  - Outputs: per case
  - Discriminates: rejects selecting top for true and treating a missing condition as true
  - runtime_observed: no
  - (a) Inputs: `07 01 07 0A 07 14 16 00`. Draws: none. Outputs: 10.
  - (b) Inputs: `07 00 07 0A 07 14 16 00`. Draws: none. Outputs: 20.
  - (c) Inputs: `07 0A 07 14 16 00`. Draws: none. Outputs: 20.

- **SCX-V06** No terminator (SCX-R01; boundary: isolated rule)
  - Inputs: `07 05`, declared length 2
  - Seeds: `no draw`
  - Draws: none
  - Trace: push 5; reach declared end
  - Outputs: 0
  - Discriminates: rejects returning top on ordinary run-off
  - runtime_observed: no

- **SCX-V07** Stack limit (SCX-R02; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: 64 values are retained; a full-stack push changes neither value slots nor depth
  - Outputs: per case
  - Discriminates: distinct overflow value 9 is discarded; returning 9 fails
  - runtime_observed: no
  - (a) Inputs: 65 repetitions of `07 01`, then `00`. Draws: none. Outputs: 1.
  - (b) Inputs: 64 repetitions of `07 01`, then `10 00`. Draws: none. Outputs: 2.
  - (c) Inputs: 64 repetitions of `07 01`, then `07 09 00`. Draws: none. Outputs: 1.

- **SCX-V08** Empty pops and negate (SCX-R02, SCX-R08; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: empty pop gives zero; negate uses wrapping signed arithmetic
  - Outputs: per case
  - Discriminates: rejects retaining an operand after negate and failure on empty binary operation
  - runtime_observed: no
  - (a) Inputs: `10 00`. Draws: none. Outputs: 0.
  - (b) Inputs: `15 00`. Draws: none. Outputs: 0.
  - (c) Inputs: `07 05 15 00`. Draws: none. Outputs: -5.
  - (d) Inputs: `09 00 00 00 80 15 00`. Draws: none. Outputs: -2147483648; negating INT_MIN wraps to INT_MIN.
  - (e) Inputs: `07 05 10 00`. Draws: none. Outputs: 5; the right operand is 5, the missing left operand is zero, and evaluation continues.
  - (f) Inputs: `10 07 05 10 00`, then independently `15 07 05 10 00`. Draws: none. Outputs: 5 in each case. Empty binary addition or empty negate produces zero and continues to the later literal/addition, rather than aborting with zero.

- **SCX-V09** Parameter widths and missing operation (SCX-R04; boundary: supplied-input stage)
  - Inputs: per case; present parameter operation logs input and returns it
  - Seeds: `no draw`
  - Draws: none
  - Trace: byte selector is unsigned; word selector is signed
  - Outputs: per case
  - Discriminates: rejects signed byte selectors and evaluating through an absent operation
  - runtime_observed: no
  - (a) Inputs: `04 80 00`, present operation. Draws: none. Outputs: 128; log [128].
  - (b) Inputs: `05 80 FF 00`, present operation. Draws: none. Outputs: -128; log [-128].
  - (c) Inputs: separately either preceding expression, absent operation. Draws: none. Outputs: 0; no call.
  - (d) Inputs: `06 FF FF FF FF 00`, present operation. Draws: none. Outputs: -1; log [-1], from the signed 32-bit selector.
  - (e) Inputs: readable bytes `04`, declared length 1, present operation. Draws: none. Outputs: 0; no parameter call because the first operand byte is outside declared length.
  - (f) Inputs: complete readable backing bytes `05 80 FF 00`, declared length 2, present operation. Draws: none. Outputs: 0; parameter log [-128]. The first operand byte is within declared length, so both readable word bytes are consumed and the callback runs; evaluation then runs past declared length without reading a terminator. This does not define a result for inaccessible backing bytes.

- **SCX-V10** Signed comparisons (SCX-R05; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: pop b then a; use signed comparison and push boolean 0 or 1
  - Outputs: per case
  - Discriminates: negative case rejects unsigned comparison; equality distinguishes strict and non-strict operators
  - runtime_observed: no
  - (a) Inputs: `07 02 07 05 0A 00`. Draws: none. Outputs: 1.
  - (b) Inputs: `07 02 07 05 0B 00`. Draws: none. Outputs: 0.
  - (c) Inputs: `07 05 07 05 0C 00`. Draws: none. Outputs: 1.
  - (d) Inputs: `07 FF 07 01 0A 00`. Draws: none. Outputs: 1.
  - (e) Inputs: `07 05 07 05 0D 00`, then independently `07 02 07 05 0D 00`. Draws: none. Outputs: 1,0 respectively (`>=`).
  - (f) Inputs: `07 05 07 05 0E 00`, then independently `07 02 07 05 0E 00`. Draws: none. Outputs: 1,0 respectively (`==`).
  - (g) Inputs: `07 05 07 05 0F 00`, then independently `07 02 07 05 0F 00`. Draws: none. Outputs: 0,1 respectively (`!=`).

- **SCX-V11** Skill binding indexes (SCX-R10, SCX-R13, SCX-R14; boundary: isolated binding selection; supplied operations in cases d through h)
  - Inputs: per case; seven skill bindings
  - Seeds: `no draw`
  - Draws: none
  - Trace: select the indexed operation; pop its argument count in source order and push its return, or push zero without pops for an invalid index
  - Outputs: per case
  - Discriminates: rejects swapped binding indexes, incorrect forwarded arguments, and returning a retained literal for an invalid index
  - runtime_observed: no
  - (a) Inputs: `07 03 07 09 01 00 00`. Draws: none. Outputs: 3.
  - (b) Inputs: `07 03 07 09 01 01 00`. Draws: none. Outputs: 9.
  - (c) Inputs: `07 03 07 09 01 09 00`. Draws: none. Outputs: 0; stack before terminator [3,9,0].
  - Cases (d) through (h) use context C and supplied semantic operations, each logging its operation name, all three argument slots, and context. The supplied `rand`, `skill`, `miss`, `stat`, and `sklvl` operations return 102,103,104,105,106 respectively, without a draw. These cases assert binding selection and forwarding only; they do not execute the operations' separately specified arithmetic or gated producers.
  - (d) Inputs: `07 03 07 09 01 02 00`. Draws: none. Outputs: 102; exactly one log entry `(rand,3,9,0,C)`.
  - (e) Inputs: `07 03 07 09 01 03 00`. Draws: none. Outputs: 103; exactly one log entry `(skill,3,9,0,C)`.
  - (f) Inputs: `07 03 07 09 01 04 00`. Draws: none. Outputs: 104; exactly one log entry `(miss,3,9,0,C)`.
  - (g) Inputs: `07 03 07 09 01 05 00`. Draws: none. Outputs: 105; exactly one log entry `(stat,3,9,0,C)`.
  - (h) Inputs: `07 02 07 03 07 04 01 06 00`. Draws: none. Outputs: 106; exactly one log entry `(sklvl,2,3,4,C)`.
  - (i) Inputs: `07 03 07 09 01 09 10 10 00`. Draws: none. Outputs: 12; invalid index pushes zero without consuming 3/9, then two additions consume the retained logical values. No function call occurs.

- **SCX-V12** Empty references, bounds, storage (SCX-R11, SCX-R24; boundary: supplied-input stage)
  - Inputs: per case; pool present with `07 05 00`, size 3; evaluation logs requests
  - Seeds: `no draw`
  - Draws: none
  - Trace: unsigned reference check precedes evaluation; a recursive empty column uses the same rule
  - Outputs: per case
  - Discriminates: rejects interpreting the empty sentinel as a negative valid reference and evaluating at the exclusive upper bound
  - runtime_observed: no
  - (a) Inputs: calculation reference `0xFFFFFFFF`. Draws: none. Outputs: 0; no evaluation.
  - (b) Inputs: parameter `0x37`, valid skill with empty `calc1`. Draws: none. Outputs: 0; no evaluation.
  - (c) Inputs: reference 3. Draws: none. Outputs: 0; no evaluation.
  - (d) Inputs: reference 0. Draws: none. Outputs: 5; evaluation request has declared length 3.
  - (e) Inputs: load an empty cell, then a nonempty cell with supplied compiler bytes `07 09 00` and length 3, starting empty pool. Draws: none. Outputs: stored references `0xFFFFFFFF`, 0; pool `07 09 00`.
  - (f) Inputs: load a nonempty cell with supplied compiler length 0, then independently length -1, starting empty pool in each case. Draws: none. Outputs: stored reference `0xFFFFFFFF`; pool remains empty in each case; no compiler bytes appended.
  - (g) Inputs: absent skill pool, reference 0. Draws: none. Outputs: 0; no evaluation.
  - (h) Inputs: present unit U, skill 0, level 3; independent skill pool `AA 07 05 00` and secondary skill/description pool `AA 07 09 00`, each size 4; request reference 1 through each respective front end. Draws: none. Outputs: 5 and 9 respectively; evaluation requests use the selected pool's bytes after offset 1, remaining length 3, and context `(U,0,3)` in each case. Leading `AA` is not executed.

- **SCX-V13** Inclusive random range and seed owner (SCX-R15, SCX-R29a; boundary: isolated rule)
  - Inputs: per case; context present
  - Seeds: per case
  - Draws: per case
  - Trace: NRF-R01: 1*0x6AC690C5+0x29A gives new low 0x6AC6935F, high 0; reduce by stated range
  - Outputs: per case
  - Discriminates: rejects exclusive upper bounds, a draw on equal/reversed endpoints, and using the missile owner seed
  - runtime_observed: no
  - (a) Inputs: `rand(3,3)` and separately `rand(5,2)`, unit seed {1,0x29A}. Draws: none. Outputs: 3 and 5 respectively; seed unchanged.
  - (b) Inputs: skill `rand(1,6)`, unit seed {1,0x29A}. Draws: unit, modulus 6, remainder 3. Outputs: 4; unit seed {0x6AC6935F,0}.
  - (c) Inputs: skill `rand(0,7)`, unit seed {1,0x29A}. Draws: unit, modulus 8, remainder 7. Outputs: 7; unit seed {0x6AC6935F,0}.
  - (d) Inputs: missile `rand(1,6)`, missile seed {1,0x29A}, owner seed {0,0}. Draws: missile, modulus 6, remainder 3. Outputs: 4; missile seed {0x6AC6935F,0}; owner seed unchanged.
  - (e) Inputs: skill `rand(1,6)`, absent context. Seeds: none. Draws: none. Outputs: 0.
  - (f) Inputs: skill `rand(-2147483648,2147483647)`, present context/unit, seed {1,0x29A}. Draws: none. Outputs: -2147483648; wrapped range length 0; seed unchanged.
  - (g) Inputs: skill `rand(3,3)`, present context, absent unit. Seeds: none. Draws: none. Outputs: 3; no fault, since no draw is needed.
  - (h) Inputs: skill `rand(1,6)`, present context, absent unit. Seeds: none. Draws: no completed seed step. Outputs: fault when accessing the absent unit's seed; no integer return.

- **SCX-V14** Skill versus hard level (SCX-R16a, SCX-R26; boundary: supplied-input stage)
  - Inputs: present unit; selected own-skill entry hard level 20, supplied bonus 3, cap 30; per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: own skill gives 20+3=23, hard selector uses 20; missing entry retains level zero
  - Outputs: per case
  - Discriminates: rejects adding bonuses to the hard-only selector and clamping a missing skill to one
  - runtime_observed: no
  - (a) Inputs: `skill(s,0x10)`. Draws: none. Outputs: 23; parameter request `(unit,0x10,s,23)`.
  - (b) Inputs: hard-level selector `0x29`. Draws: none. Outputs: 20.
  - (c) Inputs: missing selected entry; `skill(s,0)` with valid s. Draws: none. Outputs: 0; parameter request at L=0 returns zero.
  - (d) Inputs: own-skill condition false, otherwise initial inputs; `skill(s,0x10)`. Draws: none. Outputs: 20; parameter request `(unit,0x10,s,20)`; bonus 3 is not added.
  - (e) Inputs: hard level -5, bonus 0, cap 30; independently call `skill(s,0x10)` and hard selector `0x29`. Draws: none. Outputs: 0 in each; skill function parameter request has L=0.
  - (f) Inputs: hard level 40, bonus 3, cap 30; independently call `skill(s,0x10)` and hard selector `0x29`. Draws: none. Outputs: 30 in each; skill function parameter request has L=30.
  - (g) Inputs: absent unit; independently call `skill(s,0x10)`, selector `0x28`, and selector `0x29`, with valid s. Draws: none. Outputs: 0 in each; skill function requests `(null,0x10,s,0)`; the two level selectors make no stat/entry request.

- **SCX-V15** Linear/raw/level selectors (SCX-R20, SCX-R21; boundary: isolated rule)
  - Inputs: valid skill; P1=10, P2=3; per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: linear result (L-1)*3+10 when L>0; raw parameter has no level guard
  - Outputs: per case
  - Discriminates: rejects using L rather than L-1 and applying the level guard to raw parameters
  - runtime_observed: no
  - (a) Inputs: selector 0 at L=1, 5, 0 separately. Draws: none. Outputs: 10,22,0 respectively.
  - (b) Inputs: selector 8 at L=0. Draws: none. Outputs: 10.
  - (c) Inputs: selector 0x10 at L=5. Draws: none. Outputs: 5.

- **SCX-V16** Diminishing result and fourth-pair guard (SCX-R22; boundary: isolated rule)
  - Inputs: all four pairs have lo=10, hi=110; per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: f=110L/(L+6); value=10+f; clamp only above 110
  - Outputs: per case
  - Discriminates: rejects omitting the final upper clamp and applying the nonpositive guard to the fourth pair
  - runtime_observed: no
  - (a) Inputs: first diminishing selector at L=1. Draws: none. Outputs: 25 (f=15).
  - (b) Inputs: first diminishing selector at L=20. Draws: none. Outputs: 94 (f=84).
  - (c) Inputs: first diminishing selector at L=1000. Draws: none. Outputs: 110 (f=109, provisional 119).
  - (d) Inputs: fourth diminishing selector 7 at L=0. Draws: none. Outputs: 10.
  - (e) Inputs: first diminishing selector 1 at L=0. Draws: none. Outputs: 0.
  - (f) Inputs: fourth diminishing selector 7 at L=-1. Draws: none. Outputs: -12; f=-110/5=-22, r=10-22=-12, with no lower clamp to lo.
  - (g) Inputs: fourth diminishing selector 7 at L=-6. Draws: none. Outputs: signed division-by-zero fault at L+6; no integer return. This differs from the evaluator opcode's guarded zero divisor.

- **SCX-V17** Mana scale (SCX-R23; boundary: isolated rule)
  - Inputs: mana=10, lvlmana=1, manashift=8; per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: m=10+(L-1); at L = 5 m = 14; m*25/2=175 before shift
  - Outputs: per case
  - Discriminates: rejects dividing/rounding at a different stage and confusing raw and shifted values
  - runtime_observed: no
  - (a) Inputs: selector 0x15 at L = 1 and L = 5 separately. Draws: none. Outputs: 10,14 respectively.
  - (b) Inputs: selector 0x16 at L = 5. Draws: none. Outputs: 175.
  - (c) Inputs: selector 0x2A at L = 5. Draws: none. Outputs: 3584.
  - (d) Inputs: each selector 0x15,0x16,0x2A at L=0, then independently L=-1. Draws: none. Outputs: 0 for all six calls.
  - (e) Inputs: mana=-1,lvlmana=0,manashift=7,L=1; independently selectors 0x15,0x16,0x2A. Draws: none. Outputs: -1,-6,-128 respectively. For 0x16, -25/2 truncates to -12 before shift 7 and arithmetic right shift 8; rounding that division downward would yield -7.

- **SCX-V18** Elemental endpoint and parameter scale (SCX-R25a with the SLD-R08/SLD-R09 endpoint; boundary: supplied-input stage)
  - Inputs: EMin=10; all five EMinLev values 0; HitShift = 8; L = 1; mastery disabled; supplied EDmgSymPerCalc result 50
  - Seeds: `no draw`
  - Draws: none
  - Trace: D = 10 << 8=2560; mulDiv(50,2560,100)=1280; D = 3840; shifted selector 3840>>8=15
  - Outputs: EMinD = 3840; selector 0x11=15; selector 0x26=3840
  - Discriminates: rejects returning whole points from an unshifted selector
  - runtime_observed: no

- **SCX-V19** Alternate terminators (SCX-R01; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: after pushing 5, a terminating opcode returns top; zero declared length skips all bytes
  - Outputs: per case
  - Discriminates: rejects requiring only opcode 0 as terminator
  - runtime_observed: no
  - (a) Inputs: `07 05 02`, then independently `07 05 03`, `07 05 FF`. Draws:none. Outputs:5 in each.
  - (b) Inputs:`07 05 00`, declared length 0. Draws:none. Outputs:0.

- **SCX-V20** Function argument order and unsupported arity (SCX-R10; boundary: supplied-input stage)
  - Inputs: per case; function count 1; supplied function index 0 logs its three argument slots and returns 100*x1+10*x2+x3
  - Seeds: `no draw`
  - Draws: none
  - Trace: pop deepest-to-first source order; unused slots zero; unsupported arity consumes nothing
  - Outputs: per case
  - Discriminates: noncommutative return rejects reversal; unsupported arity log rejects accidental invocation
  - runtime_observed: no
  - (a) Inputs: `07 02 07 03 07 04 01 00 00`, arity 3. Draws:none. Outputs:234; log[(2,3,4)].
  - (b) Inputs: `07 02 01 00 00`, arity 0. Draws:none. Outputs:0; log[(0,0,0)].
  - (c) Inputs: `07 02 01 00 00`, arity 4. Draws:none. Outputs:0; no call; stack before terminator[2,0].
  - (d) Inputs: `07 02 01 00 00`, arity 1. Draws: none. Outputs: 200; log [(2,0,0)].
  - (e) Inputs: `07 02 01 00 10 00`, arity 4. Draws: none. Outputs: 2; no function call. Unsupported arity leaves 2 before pushing zero, so the following addition returns 2.

- **SCX-V21** Missile request return (SCX-R17; boundary: supplied-input stage)
  - Inputs: skill context(unit U,skill 0,level 7); miss(m=12,p=9); supplied resolver logs request and returns 73
  - Seeds: `no draw`
  - Draws: none
  - Trace: request absent missile, owner U, selector 9, missile 12, level 7; forward 73
  - Outputs: per case
  - Discriminates: rejects using skill identity as missile identity or dropping resolver return
  - runtime_observed: no
  - (a) Inputs: initial fields. Draws: none. Outputs: 73; exactly one request with the stated missile/owner/selector/identity/level.
  - (b) Inputs: absent context, otherwise initial request arguments. Draws: none. Outputs: 0; no missile resolver request.

- **SCX-V22** Stat routing and level read (SCX-R18a, SCX-R26; boundary: supplied-input stage)
  - Inputs: present context/unit; statCount359; ordinary read returns 37, qualifier-one11, qualifier-two37, special to-hit91; all log requests; per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: guard before dispatch; stat 19 bypasses qualifier selection; unit-level selector uses ordinary stat 12
  - Outputs: per case
  - Discriminates: own-array return differs from selected-array returns; the special to-hit override differs from both
  - runtime_observed: no
  - (a) Inputs: stat(34,0), stat(34,1), stat(34,2), stat(34,9) independently. Draws:none. Outputs:37,11,37,37 respectively; each selected getter logged once with id 34/layer 0.
  - (b) Inputs: stat(19,2). Draws:none. Outputs:91; special to-hit call only.
  - (c) Inputs: stat(359,0). Draws:none. Outputs:0; no call.
  - (d) Inputs: selector 0x28 with ordinary stat 12 supplied 37. Draws:none. Outputs:37; log[(12,0)].
  - (e) Inputs: stat(34,0) with absent context, then independently present context/absent unit; stat(-1,0) with present context/unit. Draws: none. Outputs: 0 in each; no getter or special to-hit call.

- **SCX-V23** Nested level has no clamp (SCX-R19; boundary: supplied-input stage)
  - Inputs: context(U,skill 0,level 4); sklvl(s=9,p1=8,p2=16); supplied parameter operation returns -3 for the first request and -30 for second
  - Seeds: `no draw`
  - Draws: none
  - Trace: first request(U,8,0,4)->-3; second(U,16,9,-3)->-30
  - Outputs: per case
  - Discriminates: rejects clamping intermediary level or returning the first result
  - runtime_observed: no
  - (a) Inputs: initial fields. Draws: none. Outputs: -30; ordered requests `(U,8,0,4)`, `(U,16,9,-3)`.
  - (b) Inputs: absent context, otherwise initial request arguments. Draws: none. Outputs: 0; no parameter request.
  - (c) Inputs: present context with unit null, current skill 0/level 4; `sklvl(9,8,16)`; supplied operation returns -3 then -30. Draws: none. Outputs: -30; ordered requests `(null,8,0,4)`, `(null,16,9,-3)`; a null unit alone does not suppress forwarding.

- **SCX-V24** Dispatch bounds (SCX-R20; boundary: isolated rule)
  - Inputs: per case; available skill 0 of count 1; parameter operation logs downstream requests
  - Seeds: `no draw`
  - Draws: none
  - Trace: reject invalid skill/selector before cases
  - Outputs: per case
  - Discriminates: rejects inclusive skill count and selector overflow
  - runtime_observed: no
  - (a) Inputs: skill 1,id 0,L = 1; then skill 0,id 0x49,L = 1 independently. Draws:none. Outputs:0 and no downstream request in each.

- **SCX-V25** Recursive context (SCX-R24; boundary: supplied-input stage)
  - Inputs: context(U,skill 0,level 5); selector 0x37; `calc1` reference 0; pool `04 10 00`; parameter selector 0x10 returns level
  - Seeds: `no draw`
  - Draws: none
  - Trace: nested evaluation retains level 5; parameter token 16 returns 5
  - Outputs: 5; nested request context(U,0,5); parameter request selector 16
  - Discriminates: rejects resetting recursive level to zero
  - runtime_observed: no

- **SCX-V26** Endpoint routing (SCX-R25a; boundary: supplied-input stage)
  - Inputs: context(U,skill 0,level 2); supplied elemental min disabled=768, max disabled=1024, min enabled=1280,max enabled=1536; duration=17; operations log requests
  - Seeds: `no draw`
  - Draws: none
  - Trace: select endpoint and mastery condition, then arithmetic shift only for shifted selectors
  - Outputs: selectors0x11,0x12,0x26,0x27,0x31,0x32,0x34,0x35,0x13,0x33 return 3,4,768,1024,5,6,1280,1536,17,17 respectively; each log contains its exact selected endpoint/condition
  - Discriminates: rejects interchanged min/max or missing mastery condition
  - runtime_observed: no

- **SCX-V27** To-hit fallback and unchanged calc return (SCX-R25b with the SLD-R12 value; boundary: supplied-input stage)
  - Inputs: valid skill; L = 3; ToHit = 10, LevToHit = 4; per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: empty reference uses 10+(3-1)*4=18; nonempty invokes supplied evaluator
  - Outputs: per case
  - Discriminates: rejects always using fallback or postprocessing the calculation result
  - runtime_observed: no
  - (a) Inputs: empty to-hit reference. Draws:none. Outputs:18; no evaluation.
  - (b) Inputs: present unit U; valid reference 0; supplied evaluator returns-7. Draws:none. Outputs:-7; evaluation context(U,0,3).
  - (c) Inputs: empty to-hit reference; skill identity 1 of count 1, then independently valid skill/L=0 and valid skill/L=-1. Draws: none. Outputs: 0 in each; no evaluation.
  - (d) Inputs: present unit U,valid skill/L=3,ToHit=10,LevToHit=4; nonempty reference 1 with present pool `00` of size 1. Draws: none. Outputs: 0; no evaluation and no linear fallback of 18. Independently absent pool/reference 0 gives the same result and no evaluation.

- **SCX-V28** First matching passive column (SCX-R28; boundary: supplied-input stage)
  - Inputs: valid skill; unit U; L2; selector 0x17; passive stat slots [100,345,342,200,201]; aligned valid calcs return[10,20,30,40,50]; evaluator logs selected column
  - Seeds: `no draw`
  - Draws: none
  - Trace: pair(342,345) matches slot 2 before slot 3; evaluate only aligned second calc
  - Outputs: per case; evaluations use context(U,0,L)
  - Discriminates: rejects preferring the first identity of the pair over the earliest slot or evaluating every match
  - runtime_observed: no
  - (a) Inputs: initial fields. Draws: none. Outputs: 20; only `passivecalc2` evaluated.
  - (b) Inputs: initial fields except passive stat slots [100,101,102,200,201]. Draws: none. Outputs: 0; no evaluation.
  - (c) Inputs: initial matching slots, L=0 then independently L=-1. Draws: none. Outputs: 0; no evaluation in either case.

- **SCX-V29** Missile context and remaining pool length (SCX-R29a; boundary: supplied-input stage)
  - Inputs: missile unit type 3,class 12,stored missile level 7; owner U; supplied missile identity -1, level -1; pool `04 09 00`, reference 0; supplied resolver returns 73
  - Seeds: `no draw`
  - Draws: none
  - Trace: fallback class 12/level 7; evaluate three bytes; token 9 forwards context
  - Outputs: per case; case (a) asserts resolver request(missile,owner U,selector 9,missile 12,level 7) and evaluator length 3
  - Discriminates: rejects using owner identity/level in missile context
  - runtime_observed: no
  - (a) Inputs: initial fields. Draws: none. Outputs: 73; resolver request as above.
  - (b) Inputs: initial fields except unit type 0. Draws: none. Outputs: 0; no evaluation or resolver request.
  - (c) Inputs: missile unit type 3, owner U; missile identity 12, level 7; missile pool `07 00 07 10 01 03 00`, reference 0; supplied owner entry for skill 0 with hard level 20, own-skill condition true, bonus 3, cap 30. Draws: none. Outputs: 23; binding index 3 requests skill parameter `(U,0x10,0,23)`; no missile resolver request. Entry/bonus/cap are supplied, so this does not verify their producers.

- **SCX-V30** Additional-family zero parameter (SCX-R30a; boundary: isolated rule)
  - Inputs: additional pool `04 09 00`, size 3; reference 0; opaque context(A,B)
  - Seeds: `no draw`
  - Draws: none
  - Trace: parameter token 9 resolves 0 and terminator returns 0
  - Outputs: per case
  - Discriminates: rejects forwarding the parameter to a skill/missile resolver
  - runtime_observed: no
  - (a) Inputs: initial fields. Draws: none. Outputs: 0; no function call.
  - (b) Inputs: absent additional pool, reference 0; then independently initial pool with reference 3 and reference `0xFFFFFFFF`. Draws: none. Outputs: 0; no evaluation or function call in all three cases.

- **SCX-V31** Cold and qualifier source selection (SCX-R18a; the mastery amount is SLD-R10's, also asserted by SLD-V14; boundary: integration of shared getter contract)
  - Inputs: present unit and list, definition count 359; stat 331 definition available with fMin false, MinAccr 0, ValShift 0. List usesResolvedStats true, owner type 0, ownStats [(331,0,10)], resolvedStats [(331,0,30)], each active count 1. EType 4, damage input D = 1000. Stat function context present.
  - Seeds: `no draw`
  - Draws: none
  - Trace: cold wrapper selects stat 331 and the selected-array getter; m = 30; mulDiv(30,1000,100) = 300. Qualifier one reads own value 10; qualifier two reads selected value 30.
  - Outputs: mastery amount 300; stat(331,1) = 10; stat(331,2) = 30; no list changes.
  - Discriminates: forcing cold to the own array gives 100, and treating qualifier two as an own-array read gives 10; both fail.
  - runtime_observed: no

### 8.2 Observation vectors

- **SCX-O01** Constant versus parameterized compiled expression: inspect the selected pool reference, expression bytes, and evaluator return for a known calculation. Expected execution follows SCX-R01 through SCX-R11; compiler folding/termination remains SCX-R12's gated assertion. Procedure: [expression observation](../research/spec-evidence/skill-calc-expressions.md#expression-observation). No matching observation is recorded.

## 9. Acceptance and dependencies

Partial static acceptance consists of SCX-V01 through SCX-V31's explicitly verified assertions, using supplied dependency operations where specified. This includes bytecode evaluation, pool checks/storage, dispatch/return routing, raw/linear/diminishing/mana calculations, passive matching, and damage/duration/to-hit selector routing. It does not establish compiler-produced output, skill-entry/bonus production, special to-hit results, complete missile results, or additional-family function results; the forwarded damage, duration and to-hit values carry the acceptance limits of their SLD rules. SCX-R12, SCX-R16b, SCX-R18b, SCX-R27, SCX-R29b, and SCX-R30b remain gated exactly as listed in the Id map. Complete reproduction requires closing those assertion gates and the applicable integration cases, followed by matching observations for SCX-Q09. Passing the supplied-input subset establishes no complete parity claim.

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), NRF-R01/NRF-R02 for ordered seed steps and reductions, RLI-R02a/RLI-R02b for ordinary unit stat reads; [core-game-data-fields](core-game-data-fields.md), GDA-R11a for its listed Skills-field subset and GDF-R02 for named stat identities; [skill-leveled-damage-tohit-and-elements](skill-leveled-damage-tohit-and-elements.md), SLD-R08 through SLD-R11 for the elemental endpoints and duration returned by SCX-R25a and SLD-R12 for the to-hit value returned by SCX-R25b, with SLD's own gates. Section 3 declares the additional local formula inputs, whose original field correspondence is recorded in the companion; incomplete missile-slot correspondence remains SCX-Q04. Missing dependency results are inputs only at the explicitly supplied stage; this contract does not promote them. Owners of the cited GDA, RLI ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** skill calculations consumed by [Assassin sentry firing/range](skill-assassin-sentry-firing-and-range.md) and [Death Sentry corpse damage](skill-assassin-death-sentry-corpse-damage.md); [leveled damage, to-hit, and elements](skill-leveled-damage-tohit-and-elements.md) evaluates its synergy and to-hit calculations through SCX-R01 through SCX-R11 and SCX-R13 through SCX-R28. Their callers own which fields/flags are requested. The [damage pipeline](combat-damage-pipeline-order.md) owns subsequent damage consumption.

## 10. Related specs

- [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md): numeric and ordinary stat-read contracts.
- [core-game-data-fields](core-game-data-fields.md): original-profile field/stat names and loader identities.
- [skill-leveled-damage-tohit-and-elements](skill-leveled-damage-tohit-and-elements.md): owns the damage, duration, to-hit, synergy and mastery arithmetic that the parameter selectors return; its gates do not close this contract's gates.
- [skill-assassin-sentry-firing-and-range](skill-assassin-sentry-firing-and-range.md): skill formula consumers.
- [skill-assassin-death-sentry-corpse-damage](skill-assassin-death-sentry-corpse-damage.md): corpse-damage formula consumer.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): consumes returned damage outside this boundary.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. The [source manifest](../research/source-manifest.md) identifies the original D2Common and Fog binaries. Compatible instruction readings dated 2026-10-12 are reused within their stated boundaries; targeted instruction and matched reconstruction readings dated 2026-10-15 cover the additional findings recorded in the [public companion](../research/spec-evidence/skill-calc-expressions.md). No gameplay observation is claimed. A narrow Experience text cell was read through the original loader on 2026-10-15; its exact profile/correspondence limits are in the companion. Other field/stat name evidence is consumed through the foundations. The document remains draft because the named behavior/integration gates and gameplay observations remain open.
