# Numeric and RNG Foundations

## Description

The game uses shared rules to calculate numbers, track time and produce random results. These rules determine
how fractions are rounded, how random results advance, whose random sequence each roll uses, and how new
characters, monsters and items receive their own random sequences. Using the same rules keeps combat, movement
and other game systems consistent.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: this spec; records and ids: [core-record-layouts-and-ids](core-record-layouts-and-ids.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: NRF-R, NRF-Q, NRF-V, NRF-O; policy in [README conventions](README.md#conventions)
Evidence: [source, table and observation records](../research/spec-evidence/core-numeric-and-rng-foundations.md)
Open evidence gaps: [remaining boundaries and observations](../plans/gaps/core-numeric-and-rng-foundations.md)

## 1. Scope

This spec defines the shared numerical primitives: random stepping and draw forms; seed ownership at identified
draw sites; seed initialization during creation; integer rounding helpers; fixed-point and time units; coordinate
conversions and velocity updates; and the prepared object animation draw. Record layouts, unit type, state and
mode identities, reads of existing stat arrays, state-list operations, mode-change routing and the shared damage
record belong to [core-record-layouts-and-ids](core-record-layouts-and-ids.md).
Draw eligibility and complete subsystem sequences belong to the cited gameplay owners. Timed cold, freeze,
poison, stun and burn application belongs to [combat-status-application](combat-status-application.md); the
monster life roll and life write belong to
[monster-scaling-by-players-and-difficulty](monster-scaling-by-players-and-difficulty.md) (MSD-R11, MSD-R12).
Building resolved stat arrays, complete timer/list internals, animation selection, networking and the frame loop
retain their dependency boundaries. Client random sequences are outside the server contract.

Notation: integers use two's complement; `i16`, `u16`, `i32`, `u32`, `i64`, `u64` name their signedness and
width. Signed `/` truncates toward zero; `>>` is arithmetic and `>>>` logical. Unsigned `mod` returns an
unsigned remainder. Except explicit widened products, addition, subtraction and multiplication keep the low 32
bits. Signed divide overflow or division by zero is an unguarded arithmetic fault. `low32(x)` keeps the low 32
bits, interpreted in the result's stated type. A fatal assertion ends execution with no defined subsequent
state. Illustrative operation names impose no language, API or storage design.

## 2. Trigger and ownership

- D2Game owns server random draws and unit-creation seeding. D2Common owns integer conversions, the prepared
  object animation draw and path updates. Calls run when the owning gameplay operation reaches them; primitives
  do not execute independently.
- Status application ([combat-status-application](combat-status-application.md)) and the monster life roll and
  write (MSD-R11, MSD-R12) consume these primitives; their placement within the hit and within monster creation
  is owned there.
- Velocity updates run within the movement step. Unit creation performs its seeding before the new unit's later
  initialization draws. The prepared object animation draw runs inside the animation reset that the mode setter
  (RLI-R04a) requests.
- Pure helpers emit no client message. Seed, path and animation changes reach clients through their owning
  lifecycle and networking systems.
- Full dispatch/output anchors and the source mapping of each semantic operation are in the linked companion.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| entity identity | `unit.type`, `unit.classId` | i32 type/class | Original-profile type codes are listed in [core-record-layouts-and-ids](core-record-layouts-and-ids.md) section 3.3; creation seeding tests player (0) and monster (1) classes (NRF-R12a) |
| seeds | `unit.seed`, `game.seed`, `itemData.seed`, parent or supplied seed | pair of u32 words `{lo,hi}` | Each is a distinct random state; seeding copies are `unit.initialLo`, `game.seedCopy`, `itemData.initialLo` |
| game context | `unit.game`, `game.frame`, `game.difficulty`, `game.expansion`, allocation context | association; i32 frame; u8 difficulty; i32 expansion | Difficulty 0 Normal, 1 Nightmare, 2 Hell; nonzero expansion selects expansion behavior |
| stat list | `unit.statList` | logical record or null | Read by NRF-R07 through RLI-R02a and RLI-R02b; fields and array ordering are in core-record-layouts-and-ids section 3.7 |
| item data sentinel | `itemData.sentinel` | i32 | Set to -1 when ItemData is allocated (NRF-R12a); a neutral name whose wider role is RLI-Q01 |
| position | `path.x`, `path.y`, integer position | 16.16 bit patterns; exposed integer part u16 | Fixed-point positions; static entities use a separate position representation |
| movement | `path.dirX`, `dirY`, `stepX`, `stepY`, `speed`, `targetSpeed`, `accel`, `tick`, `scale`, reset marker | direction i32 at scale 4096; steps 16.16; speeds 8.8; other fields i32 | Supplied movement state for NRF-R08b/NRF-R11a |
| monster inputs | MonStats record count; per-class MonStats `enabled` flag | i32; boolean | Bound and gate monster classes at creation seeding (NRF-R12a). The `enabled` flag is the named MonStats boolean of GDA-R08a; creation tests only that flag. This is distinct from the post-creation `creationCondition` of MSD-R19. Its gameplay role beyond this test belongs to the monster owner |
| dependency operations | resource initialization | supplied operation contract | The ordered request of NRF-R03d is distinct from dependency-produced stored effects |
| monster elemental slot | slot chance, element type, slot eligibility; stored-target type 3 branch reached | u8 chance; u8 element type; supplied booleans | Supplied entry to NRF-R04b's direct draw sites; eligibility and preparation are NRF-Q05 |
| corpse-explosion draw inputs | corpse unit and seed, sentry unit and seed, `min`, range | units; seeds `{lo,hi}`; i32 | Supplied by DSC-R04 to the NRF-R04a corpse row |
| object animation inputs | resolved type/mode, record, per-mode total, starting-frame byte, base rate, fixed-rate condition | i32 total; u8 start; i16 base rate; u8 condition | Prepared animation inputs to NRF-R13; selection and complete animation reset remain NRF-Q07 |
| process/time inputs | parentless counter, Windows FILETIME, tick count, fps and milliseconds/frame | u32 counter/ticks; u64 FILETIME; i32 rates | Wall-clock samples can be supplied explicitly; time sampling makes actual seeds variable |
| draw arguments | modulus `n`, range offset `off`, mask argument `m` | i32 modulus/offset; u32 mask argument | Select NRF-R02's draw form; the source seed is supplied separately |
| arithmetic arguments | `a`, `b`, `c`; percent operands `v`, `s`; unsigned quotient input `x`; diminishing index | i32 except u32 `x` | NRF-R05/NRF-R06 arguments and the valid 0..4 diminishing-table index for NRF-R07 |
| unit-seeder source | optional source seed, including a parent's seed or the game seed | seed pair or null | NRF-R03a steps this pair directly; it does not require a parent unit record |

| Mapping / constant | Origin | Binding |
|---|---|---|
| LCG multiplier `0x6AC690C5`, seed high word 666, parentless scramble constants, frame-rate default 25, fixed-point and direction scales | Engine code / binary constants | Fixed by original code; custom alternatives are separate capabilities |
| Diminishing table below | Engine constant table | Fixed original table selects data-defined stat identities |

### 3.1 Diminishing table

| Index | flag | A | Stat identity |
|---|---|---|---|
| 0 | 1 | 120 | 93 `item_fasterattackrate` |
| 1 | 1 | 120 | 99 `item_fastergethitrate` |
| 2 | 1 | 120 | 105 `item_fastercastrate` |
| 3 | 1 | 120 | 102 `item_fasterblockrate` |
| 4 | 1 | 150 | 96 `item_fastermovevelocity` |

### 3.2 Stats

Stat identities and names are GDF-R02. Directly consumed here: life 6/7, mana 8 and stamina 10 as 8.8
quantities (NRF-R08a); diminishing sources 93/99/105/102/96, read through RLI-R02a (NRF-R07). Other seed-site
stats belong to the cited consumers.

## 4. Rules (execution order)

These independent primitives are listed in stable id order. Ordering within an operation is explicit.

**NRF-R01** **LCG state and step** [verified in 1.13c code (static)]. A seed is `{lo,hi}`, both u32:

```
x = u64(lo) * 0x6AC690C5 + u64(hi)
lo = low32(x); hi = u32(x >>> 32)
```

The widened value cannot overflow u64. Both words are written; every positive-modulus draw uses exactly one
step, including a modulus of 1. Inline copies have the same arithmetic. Idle layout readings are corroboration,
not an observed arithmetic vector.

**NRF-R02** **Draw forms** [verified in 1.13c code (static)]. Modulus `n` is i32;
masks and results are u32:

```
Rand(seed,n):
    if n <= 0: return 0                       // no step
    step(seed)
    if (n & (n - 1)) == 0: return seed.lo & (n - 1)
    return seed.lo mod n
RandInRange(seed,n,off):
    if n <= 0: return off                     // no step
    return low32(Rand(seed,n) + off)
RandMask(seed,m):
    step(seed); return seed.lo & low32(m-1)    // always steps
```

`RandMask` does not validate a power-of-two mask argument: `m=0` returns all new lo bits; `m=1` returns 0.
Ordinary callers supply powers of two. Modulo 100 and masking with 99 differ. Equal seeds mean equal complete
word pairs.

**NRF-R03** **Seed initialization.**

- **NRF-R03a** [verified in 1.13c code (static)]. `InitSeed(seed,v)` writes `{u32(v),666}`;
  `InitTimerState(seed)` writes `{1,666}`. A new unit's seeder requires a non-null unit (a null unit is a fatal assertion) and receives a source seed pair or
  null from its caller. It first initializes the unit's seed to `{1,666}`, then tests that source: a
  non-null source selects the supplied-seed path, null the no-parent path. Both paths overwrite the seed. With a source, step that source once, copy its
  new lo to `unit.initialLo`, and initialize the child seed from that value. With no parent, take NRF-R03c's
  value `r`, write `unit.initialLo = r` and initialize the child seed to `{r,666}`. The child's seed is not
  stepped on either path.
- **NRF-R03b** [verified in 1.13c code (static)]. The seeder performs `InitTimerState` before selection and
  `InitSeed` after selection; both are seed writes, not separate random draws. The initializer contracts specify
  those writes and their supplied values; they make no numeric return assertion.
- **NRF-R03c** [verified in 1.13c code (static)]. Without a parent, take `c=counter`, set `arg=low32(2*c)`, and
  store `counter=low32(c+1)` before sampling time. The process-load counter is 0. Compute
  `t=low32(u64(FILETIME-116444736000000000) / 10000000)` (unsigned wrapped subtraction and unsigned division);
  take u32 tick count `k`; set `x=low32(k+t+arg)`. Repeat three times `x=low32(x*0x19660D+0x3C6EF35F)`; return
  `x & 0x7FFFFFFF`. Supplied samples make this deterministic; real wall-clock sampling is variable. No gameplay
  LCG step occurs on this path. Whether a gameplay creator supplies no parent is outside this arithmetic
  boundary (NRF-Q01).
- **NRF-R03d** [verified in 1.13c code (static)]. Game-seed initialization takes an existing supplied game seed,
  initializes a local seed to `{1,666}`, steps the game seed once, copies its new lo to `game.seedCopy`, and
  initializes the local seed to `{game.seed.lo,666}`. It then requests resource initialization with game
  allocation context, game resource destination, local seed, difficulty, expansion and the new lo. The incoming
  game-seed producer and resource initializer's effects/draws remain NRF-Q06; this rule proves the one direct
  step and request arguments, not all creation draws.

**NRF-R04** **Whose seed each subsystem draws from.** These rows fix only the seed owner and draw form at the
identified direct draw sites. Eligibility, the complete sequence and the executable vectors belong to the
consumer rule named in each row; those consumers depend on this table, not the reverse. Callees can add separate
draws.

- **NRF-R04a** [verified in 1.13c code (static)].

| Reached draw site | Seed owner | Draw form | Consumer rule (sequence and vectors) |
|---|---|---|---|
| Melee hit roll | attacker | modulo 100 | HIT-R08 |
| Damage-build physical, critical/deadly, elemental, drain, poison, unconditional modulus 1 and conversion modulus 5 | attacker | the corresponding damage-build forms | DPO-R02, DPO-R03, DPO-R21; order DPO-R29 |
| Monster damage doubling | attacking monster | modulo 100 | DPO-R22 |
| Leech overlay choice | attacker | Rand(2) | LSE-R12 (reached through DPO-R15) |
| Flagged-monster stun cancellation | attacker | modulo 100 | CSA-R03 |
| Attacking player's weapon durability | attacker | one modulo 100 (a chance of 10 is one draw, not ten) | DPO-R24 |
| Item knockback roll | attacker | one step, new low word & 127 | KNB-R01 |
| Block, dodge, avoid, weapon avoidance, evade | defender | Rand(100) or inline modulo 100 | SBA-R03, SBA-R04, SBA-R08 |
| Hit-recovery veto | defender | RandMask(2) / RandMask(4) | HRC-R06 |
| Concentration gate | defender | Rand(100) | HRC-R10b |
| Cold shatter | defender | modulo 100 | CSA-R06 |
| Freeze-to-cold fallback | defender | the fallback's cold application takes the cold shatter draw | CSA-R07 |
| Equipped-item durability | defender | 0-3 durability draws | DPO-R24 |
| Missile knockback | missile | Rand(100) | KNB-R02a, SBA-R11 |
| Missile damage | missile | Rand(max-min) per element | MHD-R10 |
| Death Sentry corpse-explosion damage | corpse | min+Rand(max-min) | DSC-R04 |
| Monster rolled life | new monster | the roll and bonus | MSD-R11 |
| Invisible player candidate hostility filter | candidate | Rand(100) | SFR-R03f |
| New unit and item secondary seed | game | one step each | NRF-R12a, NRF-R12b |

Unknown record-chain behavior at a listed site does not change its seed owner.

- **NRF-R04b** [verified in 1.13c code (static)]. At reached monster-mode elemental slots, the chance draw and
  random-element selector draw use that monster's seed. The slot chance is an unsigned byte. Chance 0 skips the
  slot with no draw. Chance 1..99 draws `r = Rand(100)`; the slot is taken when `r < chance` and skipped
  otherwise. Chance 100 or more takes the slot without that draw. A taken slot of element type 10 then makes
  one `Rand(5)` draw and uses element `result + 1` (1..5); a taken slot of any other element type uses its own
  element type with no further draw. In the reached stored-target type 3 branch, RandMask(2) steps the
  searcher's seed once, even if subsequent conditions ignore the result. Upstream eligibility, loaded slot
  preparation and additional callee draws are NRF-Q05, separate from these supplied-entry direct sites. Object
  animation uses NRF-R13.
- **NRF-R04c** [unresolved]. Item-event handlers can draw on seeds selected by each handler. The dispatcher
  consumes no random number itself under IEP-R07; the event id or callback slot alone does not establish the total
  count/owner. Handler-owned counts remain NRF-Q04; supplied callbacks with known draw behavior are a distinct
  integration fixture.

**NRF-R05** **mulDiv** [verified in 1.13c code (static)]. For i32 arguments:

```
if c == 0: return 0
if b > 0x100000:
    if c > (b >> 4): return low32(i64(a)*b / c)
    return low32((b/c)*a)
if a > 0x10000:
    if c > (a >> 4): return low32(i64(a)*b / c)
    return low32((a/c)*b)
return i32(low32(a*b)) / c
```

Comparisons are signed; divisions truncate. Divide-first branches lose precision and argument order can change
results. Unguarded overflow such as INT_MIN/-1 faults. A caller must specify its operand order rather than infer
symmetry from the mathematical product.

**NRF-R06** **Other percent idioms.**

- **NRF-R06a** [verified in 1.13c code (static)]. Signed division by 100 truncates. For an i32 product
  `t=low32(v*s)`, division by 128 is `(t+((t>>31)&127))>>7`, also truncating. A bare arithmetic shift floors
  instead and differs for negative nonmultiples.
- **NRF-R06b** [verified in 1.13c code (static)]. For u32 `x`, `high32(u64(x)*0x51EB851F) >>> 5` equals
  floor(x/100); the animation-rate consumer applies this unsigned helper to its already formed product. It is
  not the signed multiply-high sequence, which includes a sign correction. This rule owns the division identity,
  not animation-rate input production (NRF-Q07).

**NRF-R07** **Diminishing returns** [verified in 1.13c code (static)]. With valid index 0..4, read the unit's
stat list directly: a null unit faults (there is no unit guard and no Stat null-unit return on this path). Obtain
`v` as RLI-R02a's unit read of `T[index].stat` at layer 0 from that list (`resolvedStats` when
`usesResolvedStats` is true, otherwise `ownStats`), with RLI-R02b's minimum adjustment; a null stat list or
invalid/unavailable definition yields 0. If table flag 1 and v!=0, return `i32(low32(A*v))/low32(A+v)`; otherwise return v. Signed
division truncates; v=-A divides by zero. No index or unit-null guard exists. No draw.

**NRF-R08** **Fixed-point and speed formats.**

- **NRF-R08a** **Fixed-point formats** [verified in 1.13c code (static)]. Life (stats 6 and 7), mana (8),
  stamina (10) and damage magnitudes use signed 8.8: raw 256 represents 1.0. Path positions use 16.16 bit
  patterns whose high word is the integer part; step components use signed 16.16. Reads retain raw values except
  RLI-R02b's conditional minimum replacement; not every getter scales values. The monster life clamp and its
  ordered life write requests are MSD-R12.
- **NRF-R08b** [verified in 1.13c code (static)]. For a supplied path, set speed and targetSpeed to the
  supplied 8.8 i32 value. When it differs from current speed, first set resetMarker=15. Equal speed still
  updates targetSpeed; a null path returns without effects. No draw. Production of a correctly scaled speed
  belongs to movement modifiers.

**NRF-R09** **Frame rate and time units.**

- **NRF-R09a** [verified in 1.13c code (static)]. SetFrameRate(fps) computes signed 1000/fps and stores fps and
  milliseconds/frame. Recompute reads stored fps and rewrites milliseconds/frame. fps=0 faults. Static defaults
  are 25 and 40; all absolute event times/list expiries and durations are frames. At the original 25 fps
  profile, seconds=frames/25. The frame-loop policy is outside this helper.
- **NRF-R09b** [unresolved]. Whether any caller changes the original default after startup is NRF-Q02. The
  helper accepts other rates; absence of a direct call in the server image does not establish invariant live
  rate.

**NRF-R10** **Coordinate units.**

- **NRF-R10a** [verified in 1.13c code (static)]. Convert each signed path coordinate c to tile as trunc(c/5),
  then subtract 1 when c<0. Negative exact multiples also subtract 1 (c=-10 gives -3). The inverse returns
  low32(5*tile) per axis. One tile is 5 path units. Other coordinate-space adapters and distance metrics belong
  to their consumers.
- **NRF-R10b** [verified in 1.13c code (static)]. The two path-position getters expose unsigned 16-bit integer
  X/Y parts of a supplied 16.16 path. A null path asserts fatally. The supplied path uses NRF-R10a units. This
  extraction does not sign-extend the high word or establish a signed domain for all stored positions.
  Establishing units for every static-position/position-producing route remains NRF-Q03; the getter itself
  neither converts coordinates nor draws.

**NRF-R11** **Velocity to step vector.**

- **NRF-R11a** [verified in 1.13c code (static)]. With a supplied valid path:

```
k = scale if scale > 0 else 1024
if accel != 0:
    tick = low32(tick+1)
    if tick >= 5:
        speed = low32(speed+accel)
        if speed > targetSpeed: speed=targetSpeed; accel=0
        elif speed < 0: speed=0                // keep accel
        tick=0
s = i32(low32(speed*k)) >> 6
stepX = i32(low32(dirX*s)) >> 12
stepY = i32(low32(dirY*s)) >> 12
```

All comparisons are signed. Target-speed test precedes negative-speed test. No draw.
- **NRF-R11b** [verified in 1.13c code (static)]. Original prepared direction components use scale=4096: an axis
  component of 4096 represents 1; an equal-axis diagonal uses 2896, 2896 from the fixed lookup. Direction
  production quantizes a slope to a 128-entry engine constant table and applies swaps/signs. This rule fixes
  scale and the two cited entries; reconstructing every table entry and every path creator's normalization is
  NRF-Q03. NRF-R11a accepts explicit direction components independently.

**NRF-R12** **Creation seeding.**

- **NRF-R12a** [verified in 1.13c code (static)]. At creation's seeding boundary the class tests run before any
  allocation or seeder request; a rejected creation returns null with no allocation, seed write or draw. Reject
  player class>=7 by unsigned comparison (negative also rejects). Reject a monster when class<0,
  class>=MonStatsCount, its MonStats record is unavailable, or the record's `enabled` flag is false. Other
  type 2..5 classes are not tested there. Player skips the seeder: 0 game-seed steps. Other successfully allocated types
  call the unit seeder with game seed: 1 step and unit seed/copy from its new lo. Item then allocates zeroed
  ItemData with `itemData.sentinel`=-1 and invokes NRF-R12b: 2 total game-seed steps, with a distinct secondary seed.
  Separate missile creation calls the seeder once, not the generic creator. Later monster/item/missile
  initialization draws belong to their gameplay owners; these counts are only creation seeding, not whole
  creation totals. Player loader seeding remains NRF-Q08.
- **NRF-R12b** [verified in 1.13c code (static)]. For an item with ItemData: initialize its secondary seed
  to `{1,666}`, step the game seed once, copy the game's new lo to itemData.initialLo, read that copy and
  initialize ItemData.seed={copy,666}. The item's unit seed and unit.initialLo are unchanged. The secondary
  helper requires type 4 and existing ItemData for the asserted copy write; a wrong type/missing data is not a
  successful fixture.

**NRF-R13** **Object animation reset draw** [verified in 1.13c code (static)]. At the prepared object branch,
use a valid prepared resolved mode 0..7 (the upper bound mode>=8 asserts fatally; invalid negative indexing is
outside the valid-record boundary) and a valid loaded object record. Set animation.total to the supplied
per-mode total, animation.frame to startingFrameByte<<8, and take signed 16-bit baseRate. If
fixedRateCondition!=0, store the 16-bit baseRate unchanged, with no draw and no clamp. Otherwise `roll=Rand(object.seed,baseRate>>3)` (signed
arithmetic shift); compute `rate=low32(baseRate-(baseRate>>4)+roll)`, clamp signed rate<=0 to 0 and rate>=32767
to 32767, and store animation.rate as 16-bit. After the rate calculation animation.frame remains
startingFrameByte<<8. A nonpositive shifted modulus makes no step. The dispatch table selects this random branch
only for prepared object type 2; prepared players/monsters never enter it. Resolving transformed type/mode and
other animation reset writes remains NRF-Q07.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| NRF-R01 | Two-word LCG step | verified in 1.13c code (static) | none | NRF-V01 |
| NRF-R02 | Draw forms and zero-modulus controls | verified in 1.13c code (static) | none | NRF-V01, NRF-V02, NRF-V03 |
| NRF-R03a | Initialization and parent path | verified in 1.13c code (static) | none at supplied-parent boundary | NRF-V04, NRF-V16 |
| NRF-R03b | Initial write before final seed write | verified in 1.13c code (static) | none | NRF-V04 |
| NRF-R03c | Parentless counter and sampled-time arithmetic | verified in 1.13c code (static) | none for supplied clock samples; NRF-Q01 is observation/outside-boundary | NRF-V05 |
| NRF-R03d | One game step and resource request | verified in 1.13c code (static) | NRF-Q06: upstream seed/resource integration only | NRF-V06 |
| NRF-R04a | Direct-site seed owners | verified in 1.13c code (static) | none at the reached-site boundary; eligibility belongs to the cited owners | NRF-V07 (corpse owner); every other row's vectors belong to its consumer rule |
| NRF-R04b | Reached monster-slot and searcher draws | verified in 1.13c code (static) | NRF-Q05: upstream eligibility/additional callee totals | NRF-V08 |
| NRF-R04c | Callback-owned draw totals | unresolved | NRF-Q04: per-handler totals and owners | none |
| NRF-R05 | mulDiv signed branches and operand order | verified in 1.13c code (static) | none | NRF-V09 |
| NRF-R06a | Signed truncating percent idioms | verified in 1.13c code (static) | none | NRF-V10 |
| NRF-R06b | Unsigned multiply-high /100 | verified in 1.13c code (static) | none for supplied x | NRF-V15 |
| NRF-R07 | Diminishing formula and getter | verified in 1.13c code (static) | none at existing-array getter boundary | NRF-V11 |
| NRF-R08a | Signed 8.8 and 16.16 formats | verified in 1.13c code (static) | none | Format definition, no operation of its own; exercised by NRF-V14 and NRF-V15 (g) for 16.16 and by MSD-V12 for 8.8 life writes |
| NRF-R08b | Speed setter | verified in 1.13c code (static) | none for supplied path/value | NRF-V15 |
| NRF-R09a | Frame-rate helper and original defaults | verified in 1.13c code (static) | none for helper; live-rate policy is NRF-Q02 | NRF-V12 |
| NRF-R09b | Later live-rate changes | unresolved | NRF-Q02: caller/live policy | none |
| NRF-R10a | Path/tile conversions | verified in 1.13c code (static) | none | NRF-V13 |
| NRF-R10b | Integer parts of supplied path | verified in 1.13c code (static) | NRF-Q03: other producers/static routes only | NRF-V15 |
| NRF-R11a | Velocity ramp and step arithmetic | verified in 1.13c code (static) | none for supplied path/direction | NRF-V14 |
| NRF-R11b | Direction scale and two fixed entries | verified in 1.13c code (static) | NRF-Q03: complete direction producer/table identity beyond cited entries | NRF-V15 |
| NRF-R12a | Creation seeding direct counts | verified in 1.13c code (static) | NRF-Q08: player loader producer only; later creation draws outside these counts | NRF-V16, NRF-V17 |
| NRF-R12b | Item secondary seed | verified in 1.13c code (static) | none at successful item-data boundary | NRF-V17 |
| NRF-R13 | Prepared object branch rate draw | verified in 1.13c code (static) | NRF-Q07: record/type/mode preparation and other reset writes | NRF-V18 |

## 5. Outputs and side effects

| Output / effect | Owning rule |
|---|---|
| Changed lo/hi only on stepped seed; derived random result | NRF-R01/NRF-R02/NRF-R04 |
| Child seed/initialLo, counter, game seed/copy/local seed, item secondary seed/copy; ItemData allocation with `itemData.sentinel` -1 | NRF-R03/NRF-R12 |
| Resource-initialization request with seed, difficulty, expansion and copy | NRF-R03d |
| Signed mulDiv/percent/diminishing result, or unsigned quotient result; no side effects or draws | NRF-R05/NRF-R06/NRF-R07 |
| Path speed/target/reset marker; ramp state/step components | NRF-R08b/NRF-R11a |
| fps/milliseconds globals; coordinate results | NRF-R09a/NRF-R10 |
| Prepared object animation frame/total/rate and optional seed step | NRF-R13 |

Requested dependency effects become actual stored/queued effects only under that dependency's acceptance. No
pure helper sends messages, creates units or persists saves.

Seed initializers, the unit-seeding stage and the item-secondary-seeding stage assert writes and draws rather
than a numeric return. Creation vectors assert a created unit or rejection only within NRF-R12a's supplied
successful-allocation boundary; they do not assert the complete creator's later result or initialization.

## 6. Branches and edge cases

- Draws with nonpositive signed moduli skip the step; mask draws always step. Overflow wraps except explicit
  widening or divide faults.
- Type-specific seeding and missing record behavior are explicitly per rule. The prepared object animation draw
  has its own upper-bound assertion on the prepared mode.
- Unknown condition names convey only the tested branch, not inferred gameplay semantics. Custom remapping and
  complete new-format support are outside original-profile bindings.
- Callback-owned RNG, producer routes and unresolved dependency effects do not acquire verification from
  supplied-return fixtures.

## 7. Open questions

| Question | Kind / status | Evidence that would settle it |
|---|---|---|
| **NRF-Q01** | outside-boundary / observation pending | Identify an actual gameplay creator that passes no parent and observe its sampled inputs; supplied-clock arithmetic NRF-R03c is verified |
| **NRF-Q02** | behavior / unresolved | Establish the startup/runtime callers of the rate helper or valid captures of rate writes; helper arithmetic is verified |
| **NRF-Q03** | integration / partially established | Complete coordinate/direction producer contracts for static routes and all direction-table entries; supplied path getters, scale and cited entries are verified |
| **NRF-Q04** | outside-boundary / unresolved | Complete per-handler callback/callee draw contracts and data-to-handler bindings under the item-event owner; IEP-R07, IEP-R13b and IEP-R14 to IEP-R18 establish only their stated direct-draw slices, not a blanket seed total |
| **NRF-Q05** | integration / open | Complete upstream monster-slot/stored-target eligibility and extra callee draw contracts; reached direct sites NRF-R04b verified |
| **NRF-Q06** | outside-boundary / integration open | Incoming game-seed producer and resource initializer contract; one direct step/request NRF-R03d verified |
| **NRF-Q07** | integration / open | Animation preparation, rate-input production and other animation reset writes (NRF-R06b, NRF-R13), including the effects of the animation reset requested by RLI-R04a. Supplied inputs and the prepared object branch remain eligible. Changed-state, disguise and update-queue side effects are RLI-Q02; status-path refresh and timer-node effects are CSA-Q02 |
| **NRF-Q08** | outside-boundary / open | Player-load seed producer and all reseeding paths; creation seeding skips players |
| **NRF-Q09** | observation / pending | Matching gameplay/helper captures for the verified rules and NRF-O01 to NRF-O03; no gameplay observation has been made |

Attempts and exact blocked assertion boundaries are in the [gap file](../plans/gaps/core-numeric-and-rng-foundations.md). Pure helpers have no independently visible effect; caller captures can observe their inputs/results. Archive cell reads and recorded layout/global reads are not completed executable gameplay observations.

## 8. Test vectors

### 8.1 Executable vectors

Defaults for every vector unless its Inputs state otherwise:

- Fixtures: stat definition count 359 with available named entries; valid non-null associations; no
  unmentioned lists, states or conditions; successful allocation. Every case starts fresh from its own Inputs and
  the vector's Inputs.
- Local execution: execute each cited verified NRF operation and its in-boundary NRF callees. In particular,
  seeding executes the seed initializers and LCG step, and diminishing executes the selected-array getter of
  RLI-R02a/RLI-R02b. Parent/source selection and seed writes are observed local actions, not supplied replacement
  operations. NRF-V04's parent-path fixture supplies `parent.seed` as the source seed; NRF-V16/NRF-V17 supply
  `game.seed`.
- Operations: the supplied resource initialization operation logs its semantic arguments in call order and
  applies exactly the stored effect the vector states; it draws nothing. A log entry names the operation and its arguments. Logs (request order,
  parent selection) are asserted outputs. This makes local arithmetic and ordering executable without accepting
  dependency integration. NRF-V11 uses the getters of RLI-R02a/RLI-R02b themselves, with no dependency mock.
- Seeds: every seed is a full `{lo,hi}` pair; `0x29A` is 666. A starting game seed `{1, 0x29A}` is a chosen
  input, not a claim about new-game startup. "No draw" means every seed equals its input. Reference steps
  (NRF-R01): `{1, 0x29A}` steps to `{0x6AC6935F, 0}` (lo 1791398751); `{0x6AC6935F, 0}` steps to
  `{0x2F2ED81B, 0x2C890AFD}` (lo 791599131).
- Faults: an Outputs entry "fault" means the operation does not return normally and no defined state follows;
  a test checks only that the operation does not complete with a result.
- Assertions: numeric assertions belong to the rule ids a vector names; actual dependency effects remain the Id
  map's named gates. These fixtures make no whole-game seed-total claim.
- Every vector has `runtime_observed: no`.

- **NRF-V01** Step and draw forms (NRF-R01, NRF-R02; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `S = {1, 0x29A}` unless the case states another
  - Draws: per case
  - Trace: `x = 1 * 0x6AC690C5 + 0x29A = 0x6AC6935F` gives `S = {0x6AC6935F, 0}`; from `{0x6AC6935F, 0}`,
    `x = 0x2C890AFD_2F2ED81B` gives `{0x2F2ED81B, 0x2C890AFD}`.
  - Outputs: per case
  - (a) `Rand(S, 100)`. Draws: (S, 100, 51). Outputs: 51; `S = {0x6AC6935F, 0}`.
  - (b) `RandMask(S, 0x80)`. Draws: (S, mask 0x80, 0x5F). Outputs: 0x5F; `S = {0x6AC6935F, 0}`.
  - (c) `Rand(S, 2)`. Draws: (S, 2, 1). Outputs: 1; `S = {0x6AC6935F, 0}`.
  - (d) Inputs: `S = {0x6AC6935F, 0}`; `Rand(S, 100)`. Draws: (S, 100, 31). Outputs: 31;
    `S = {0x2F2ED81B, 0x2C890AFD}`.
  - (e) `RandMask(S, 100)`. Draws: (S, mask 100, 67). Outputs: `0x6AC6935F & 99` = 67; `S = {0x6AC6935F, 0}`.
  - Discriminates: (e) returns 67 where the modulo form returns 51, rejecting modulo for the mask form; (d)
    returns 31 from the new low word, rejecting a result taken from the old low word (51).
  - runtime_observed: no

- **NRF-V02** Nonpositive modulus and modulus one (NRF-R02; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `S = {1, 0x29A}` in every case
  - Draws: per case
  - Trace: a nonpositive modulus returns before the step; a mask draw always steps.
  - Outputs: per case
  - (a) `Rand(S, 0)`. Draws: none. Outputs: 0; `S` unchanged.
  - (b) `Rand(S, -5)`. Draws: none. Outputs: 0; `S` unchanged.
  - (c) `Rand(S, 1)`. Draws: (S, 1, 0). Outputs: 0; `S = {0x6AC6935F, 0}`.
  - (d) `RandInRange(S, 0, 7)`. Draws: none. Outputs: 7; `S` unchanged.
  - (e) `RandInRange(S, 100, 7)`. Draws: (S, 100, 51). Outputs: 58; `S = {0x6AC6935F, 0}`.
  - (f) `RandMask(S, 0)`. Draws: (S, mask 0, 1791398751). Outputs: 1791398751 (0x6AC6935F);
    `S = {0x6AC6935F, 0}`.
  - (g) `RandMask(S, 1)`. Draws: (S, mask 1, 0). Outputs: 0; `S = {0x6AC6935F, 0}`.
  - Discriminates: (c) rejects skipping the step for modulus 1; (a), (b) and (d) reject stepping before the
    guard; (f) rejects validating the mask argument.
  - runtime_observed: no

- **NRF-V03** Power-of-two modulus uses the new low word (NRF-R02; boundary: isolated rule)
  - Inputs: `Rand(S, 64)`
  - Seeds: `S = {0x6AC6935F, 0}`
  - Draws: (S, 64, 27)
  - Trace: one step gives lo 0x2F2ED81B; `0x2F2ED81B & 63` = 27.
  - Outputs: 27; `S = {0x2F2ED81B, 0x2C890AFD}`
  - Discriminates: the old low word gives `0x6AC6935F & 63` = 31; 27 rejects masking before the step.
  - runtime_observed: no

- **NRF-V04** Seed initialization and parent path (NRF-R03a, NRF-R03b; boundary: isolated rule)
  - Inputs: per case
  - Seeds: per case
  - Draws: per case
  - Trace: NRF-R03a's writes in NRF-R03b's order.
  - Outputs: per case
  - (a) Inputs: `seed = {9, 10}`; `InitSeed(seed, 5)`. Draws: none. Outputs: `seed = {5, 0x29A}`.
  - (b) Inputs: non-null child with `child.seed = {9, 10}`, `child.initialLo = 77`; non-null parent with
    `parent.seed = {1, 0x29A}`. Draws: one parent step (NRF-R01), new lo 0x6AC6935F. Outputs, as an ordered log:
    (1) InitTimerState writes `child.seed = {1, 666}`; (2) the parent is selected while `child.seed` is still
    `{1, 666}`; (3) the parent step writes `parent.seed = {0x6AC6935F, 0}` with `child.seed` still `{1, 666}`;
    (4) `child.initialLo = 0x6AC6935F`, then InitSeed writes `child.seed = {0x6AC6935F, 666}`. The child seed is
    never stepped.
  - (c) Inputs: the seeder called with a null unit and a null parent. Draws: none. Outputs: fault.
  - Discriminates: (b)'s ordered log rejects moving the initial child write after parent selection or stepping,
    although the final values would match; child hi 666 rejects copying the parent's hi.
  - runtime_observed: no

- **NRF-V05** Parentless supplied clock samples (NRF-R03c, NRF-R03a; boundary: isolated rule)
  - Inputs: per case; supplied FILETIME, tick count and counter
  - Seeds: `no draw` (no gameplay LCG step)
  - Draws: none
  - Trace: NRF-R03c: `arg = low32(2 * counter)`, counter incremented before sampling, converted seconds plus
    ticks plus `arg`, three scramble rounds, low 31 bits.
  - Outputs: per case
  - (a) Inputs: FILETIME 116444736000000000 (Unix epoch), tick count 0, counter 0. Outputs: `arg` 0; counter 1;
    seconds 0; initial `x` 0; rounds 1013904223, 1196435762, 3519870697; return 1372387049.
  - (b) Inputs: FILETIME 116444736000000000, tick count 0, counter 1. Outputs: `arg` 2; counter 2; initial `x` 2;
    rounds 1017233273, 1975575172, 811535379; return 811535379.
  - (c) Inputs: FILETIME 116444735999999999 (epoch minus 1), tick count 0, counter 4294967295. Outputs: `arg`
    4294967294; counter 0; unsigned 64-bit difference 18446744073709551615; quotient 1844674407370; low 32 bits
    2133437386; initial `x` 2133437384; rounds 310795399, 3209588794, 701219409; return 701219409.
  - (d) Inputs: FILETIME 116444736000000000, tick count 1000, counter 0. Outputs: `arg` 0; counter 1; initial `x`
    1000; rounds 2678429223, 4219084122, 2266909937; return 119426289.
  - (e) Inputs: FILETIME 116444736000000000, tick count 4294967295, counter 1. Outputs: `arg` 2; counter 2;
    initial `x` = `low32(4294967295 + 0 + 2)` = 1; rounds 1015568748, 1586005467, 2165703038; return 18219390.
  - (f) Inputs: FILETIME 116444736000000000, tick count 0, counter 0; a non-null child unit seeded with no parent
    (NRF-R03a). Outputs: `child.initialLo` 1372387049; `child.seed = {1372387049, 666}`; counter 1; the child
    seed is not stepped.
  - (g) Inputs: counter 7; supplied FILETIME and tick-count providers log the counter on entry, return
    116444736000000000 and 0 respectively, and have no other effect. Outputs: `arg` 14; counter 8 before
    either sample; both providers record counter 8; seconds 0; initial `x` 14; rounds 1037207573,
    2355444336, 1741392655; return 1741392655. No gameplay LCG seed changes.
  - Discriminates: (a) and (b) reject one scramble round and a plain (undoubled) counter; (c) rejects signed
    time subtraction, a rounded quotient and doubling the incremented counter; (d) rejects omitting the tick
    count and (e) a widened (unwrapped) sum; (f) rejects leaving `initialLo` unwritten on the no-parent path;
    (g) rejects sampling either clock value before writing the incremented counter.
  - runtime_observed: no

- **NRF-V06** Game seed hand-off (NRF-R03d; boundary: supplied-input stage)
  - Inputs: existing `game.seed = {1, 0x29A}`, difficulty 2, expansion 1, allocation context P, resource
    destination D; the supplied resource initializer logs its arguments and takes no draw
  - Seeds: `game.seed = {1, 0x29A}`
  - Draws: (game, step, lo 0x6AC6935F)
  - Trace: local seed `{1, 0x29A}`; one game step; copy; local seed reinitialized from the new lo.
  - Outputs: `game.seed = {0x6AC6935F, 0}`; `game.seedCopy` 0x6AC6935F; local seed `{0x6AC6935F, 0x29A}`;
    logged request `(P, D, local seed, 2, 1, 0x6AC6935F)`.
  - Ordered local assertions: (1) InitTimerState writes local seed `{1, 666}` while `game.seed = {1, 666}`;
    (2) the game step changes its pair to `{0x6AC6935F, 0}` while local seed is still `{1, 666}`;
    (3) write `game.seedCopy = 0x6AC6935F`; (4) InitSeed writes local seed `{0x6AC6935F, 666}`;
    (5) request resource initialization with the stated arguments and that initialized local seed.
  - Discriminates: the local seed and request lo reject taking the pre-step low word (1); this vector does not
    establish the incoming seed producer or the resource initializer's totals (NRF-Q06). The ordered local
    assertions reject moving the initial local-seed write after the game step.
  - runtime_observed: no

- **NRF-V07** Corpse draw owner (NRF-R04a; boundary: supplied-input stage)
  - Inputs: corpse-explosion draw site with supplied range and `min` 0 unless a case states `min`; corpse unit;
    sentry unit
  - Seeds: per case; `sentry.seed = {5, 0x29A}`
  - Draws: per case
  - Trace: `min + Rand(corpse.seed, range)`.
  - Outputs: per case
  - (a) Inputs: `corpse.seed = {1, 0x29A}`, range 10240. Draws: (corpse, 10240, 2911). Outputs: 2911;
    `corpse.seed = {0x6AC6935F, 0}`; `sentry.seed = {5, 0x29A}`.
  - (b) Inputs: `corpse.seed = {1, 0x29A}`, range 0. Draws: none. Outputs: 0 (= `min`); both seeds unchanged.
  - (c) Inputs: `corpse.seed = {1, 0x29A}`, `min` 5, range 10240. Draws: (corpse, 10240, 2911). Outputs: 2916;
    `corpse.seed = {0x6AC6935F, 0}`.
  - Discriminates: (a)'s unchanged sentry seed rejects drawing from the caster; (b) rejects stepping for an
    empty range; (c) rejects dropping `min`. The complete skill, including the order of its other requests, is
    DSC-R04's.
  - runtime_observed: no

- **NRF-V08** Reached monster-slot and searcher draw owners (NRF-R04b; boundary: supplied-input stage)
  - Inputs: per case; prepared local draw sites only, excluding upstream eligibility and downstream lookups and
    draws (NRF-Q05); `other.seed = {5, 0x29A}`
  - Seeds: per case
  - Draws: per case
  - Trace: NRF-R04b's chance draw, selector draw and searcher mask draw.
  - Outputs: per case; `other.seed` unchanged in every case
  - (a) Inputs: eligible monster slot, chance 75, element type 10, `monster.seed = {1, 0x29A}`. Draws: (monster,
    100, 51); (monster, 5, 1). Outputs: chance passes; element 2; `monster.seed = {0x2F2ED81B, 0x2C890AFD}`.
  - (b) Inputs: chance 50, element type 10, `monster.seed = {1, 0x29A}`. Draws: (monster, 100, 51). Outputs: slot
    skipped; `monster.seed = {0x6AC6935F, 0}`.
  - (c) Inputs: chance 100, element type 10, `monster.seed = {1, 0x29A}`. Draws: (monster, 5, 1). Outputs: no
    chance draw; element 2; `monster.seed = {0x6AC6935F, 0}`.
  - (d) Inputs: reached stored-target type 3 branch, `searcher.seed = {1, 0x29A}`. Draws: (searcher, mask 2, 1).
    Outputs: `searcher.seed = {0x6AC6935F, 0}`; downstream lookup outputs are supplied and not asserted.
  - (e) Inputs: chance 51, element type 10, `monster.seed = {1, 0x29A}`. Draws: (monster, 100, 51). Outputs:
    51 is not below 51, slot skipped; `monster.seed = {0x6AC6935F, 0}`.
  - (f) Inputs: chance 0, element type 10, `monster.seed = {1, 0x29A}`. Draws: none. Outputs: slot skipped;
    seed unchanged.
  - (g) Inputs: chance 100, element type 3, `monster.seed = {1, 0x29A}`. Draws: none. Outputs: slot taken with no
    selector draw; seed unchanged.
  - (h) Inputs: chance 200, element type 3, `monster.seed = {1, 0x29A}`. Draws: none. Outputs: slot taken; seed
    unchanged.
  - (i) Inputs: separately, chance 100 with element type 4 and with element type 11, `monster.seed = {1, 0x29A}`.
    Draws: none. Outputs: each slot taken with its own element type; seed unchanged.
  - Discriminates: (c) rejects drawing the chance at 100; (b) rejects drawing the selector after a failed
    chance; (e) rejects `r <= chance`; (f) rejects drawing for a zero chance; (g) and (i) reject a selector draw for element types other than 10 (including a range test such as `>= 10`); (h)
    rejects a signed chance byte (200 would read as -56 and skip the slot); (a)'s element 2 rejects using the
    raw selector result; the unchanged `other.seed` rejects any other owner.
  - runtime_observed: no

- **NRF-V09** mulDiv branches, operand order and fault (NRF-R05; boundary: isolated rule)
  - Inputs: per case; every argument i32
  - Seeds: `no draw`
  - Draws: none
  - Trace: test `b > 0x100000`, then `a > 0x10000`, then `c > (large operand >> 4)`; in (c) and (e)
    `0x200000 >> 4 = 0x20000`, in (d) `0x200001 >> 4 = 0x20000`.
  - Outputs: per case
  - (a) `mulDiv(-7, 100, 3)`: narrow branch, `-700 / 3`. Outputs: -233.
  - (b) `mulDiv(0x3000, 100, 100)`: narrow branch, product 0x12C000. Outputs: 0x3000 (12288).
  - (c) `mulDiv(0x200000, 300, 100)`: a-large, `100 <= 0x20000`, divide first: `20971 * 300`. Outputs: 6291300.
  - (d) `mulDiv(3, 0x200001, 100)`: b-large, divide first: `20971 * 3`. Outputs: 62913.
  - (e) `mulDiv(0x200000, 300, 0x30000)`: a-large, `0x30000 > 0x20000`, widened `629145600 / 196608`.
    Outputs: 3200.
  - (f) `mulDiv(5, 7, 0)`. Outputs: 0.
  - (g) `mulDiv(-2147483648, 1, -1)`: neither signed threshold holds; narrow product -2147483648 divided by -1.
    Outputs: arithmetic-overflow fault; no value is returned.
  - (h) `mulDiv(0x200000, 0x400000, 0x30000)`: both operands large; the b test comes first, `0x30000 <= 0x40000`,
    divide first: `(0x400000 / 0x30000) * 0x200000 = 21 * 2097152`. Outputs: 44040192.
  - (i) `mulDiv(3, 0x200001, 0x30000)`: b-large, `0x30000 > 0x20000`, widened `6291459 / 196608`. Outputs: 32.
  - (j) `mulDiv(0x200017, 0x10000, 0x20001)`: a-large, `c` equals `a >> 4` (0x20001), so the strict test fails;
    divide first: `16 * 65536`. Outputs: 1048576.
  - (k) `mulDiv(0x10000, 0xFFFFF, 3)`: neither threshold holds; narrow product `low32(0xFFFFF0000)` = -65536;
    `-65536 / 3`. Outputs: -21845.
  - (l) `mulDiv(0x10000, 0x10000F, 0x10000)`: b-large, `c` equals `b >> 4` (0x10000), so the strict test fails;
    divide first: `(0x10000F / 0x10000) * 0x10000 = 16 * 65536`. Outputs: 1048576.
  - (m) `mulDiv(3, 0x100000, 7)`: `b` equals 0x100000, so the b test fails; narrow `3145728 / 7`. Outputs:
    449389.
  - Discriminates: (c) rejects the exact product result 6291456 and (d) the exact 62914; (e) rejects divide-first
    (`(0x200000 / 0x30000) * 300` = 3000); (a) rejects floor division (-234); (g) rejects a wrapped return; (h)
    rejects testing `a` first (that branch widens to 44739242); (i) rejects divide-first on the b branch (30); (j)
    rejects `>=` in the threshold test (widened 1048579); (k) rejects widening the narrow branch (widened low 32
    bits 1431633920); (l) rejects `>=` in the b threshold test (widened 1048591); (m) rejects `b >= 0x100000` (divide-first 449388).
  - runtime_observed: no

- **NRF-V10** Truncating percent idioms (NRF-R06a; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: `(t + ((t >> 31) & 127)) >> 7`; signed `/ 100`.
  - Outputs: per case
  - (a) Product `t` -129, divided by 128: `-129 + 127 = -2`, `-2 >> 7`. Outputs: -1.
  - (b) Product `t` 129, divided by 128: `129 + 0`, `>> 7`. Outputs: 1.
  - (c) Signed `-150 / 100`. Outputs: -1.
  - Discriminates: (a) rejects a bare arithmetic shift (`-129 >> 7` = -2); (c) rejects floor division (-2).
  - runtime_observed: no

- **NRF-V11** Diminishing returns (NRF-R07; boundary: isolated rule with the actual getter)
  - Inputs: non-null unit whose list has `usesResolvedStats` false and `ownStats = [(99, 0, v)]` with active
    count 1; every definition available with `fMin` false, `MinAccr` 0, `ValShift` 0. Index 1 is flag 1, A 120,
    stat 99.
  - Seeds: `no draw`
  - Draws: none
  - Trace: `i32(low32(A * v)) / low32(A + v)`, signed and truncating.
  - Outputs: per case
  - (a) Index 1, `v` 10: `1200 / 130`. Outputs: 9.
  - (b) Index 1, `v` 30: `3600 / 150`. Outputs: 24.
  - (c) Index 1, `v` 60: `7200 / 180`. Outputs: 40.
  - (d) Index 1, `v` 120: `14400 / 240`. Outputs: 60.
  - (e) Index 1, `v` -60: `-7200 / 60`. Outputs: -120.
  - (f) Index 4 (flag 1, A 150, stat 96), `ownStats = [(96, 0, 100)]`: `15000 / 250`. Outputs: 60.
  - (g) Index 1, `ownStats` empty with active count 0: the read gives 0, returned unchanged. Outputs: 0.
  - (h) Index 1, unit stat list null. Outputs: 0.
  - (i) Index 1, `v` 100: `12000 / 220`. Outputs: 54.
  - (j) Index 1, `v` -50: `-6000 / 70`. Outputs: -85.
  - (k) Index 1, list `usesResolvedStats` true, `ownStats = [(99, 0, 60)]`, `resolvedStats = [(99, 0, 10)]`, each
    active count 1: the unit read gives 10; `1200 / 130`. Outputs: 9.
  - (l) Index 1, `v` -120: present entry passes through the getter; numerator -14400, denominator 0. Outputs:
    arithmetic divide-by-zero fault; no value is returned and no post-fault state is asserted.
  - Discriminates: (k) rejects the list read of `ownStats` (it would give 40); (i) rejects rounding to nearest
    (55) and (j) rejects floor division (-86); (a) rejects
    returning `v` unchanged; (e) rejects an absolute or unsigned computation; (f)
    rejects index 1's A for index 4 (12000 / 220 = 54). Case (l) rejects returning zero or the stored value when
    the nonzero stored value makes `A + v` zero.
  - runtime_observed: no

- **NRF-V12** Frame-rate helper (NRF-R09a; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: milliseconds/frame = signed `1000 / fps`.
  - Outputs: per case
  - (a) `SetFrameRate(25)`. Outputs: fps 25, milliseconds/frame 40.
  - (b) `SetFrameRate(30)`. Outputs: fps 30, milliseconds/frame 33.
  - (c) `SetFrameRate(24)`. Outputs: fps 24, milliseconds/frame 41.
  - (d) Inputs: stored fps 24, milliseconds/frame 999; Recompute. Outputs: fps 24, milliseconds/frame 41.
  - Discriminates: (c) rejects rounding to 42; (d) rejects keeping the stored 999. Recorded live 25/40 reads do
    not exercise these calls.
  - runtime_observed: no

- **NRF-V13** Path and tile conversion (NRF-R10a; boundary: isolated rule)
  - Inputs: per case; signed path coordinate `c` or tile `t`
  - Seeds: `no draw`
  - Draws: none
  - Trace: `trunc(c / 5)`, then subtract 1 when `c < 0`; inverse `low32(5 * t)`.
  - Outputs: per case
  - (a) `c` 12: `12 / 5`. Outputs: 2.
  - (b) `c` -1: `trunc(-1 / 5)` = 0, then subtract 1. Outputs: -1.
  - (c) `c` -10: -2, then subtract 1. Outputs: -3.
  - (d) Inverse `t` 2. Outputs: 10.
  - (e) Inverse `t` -3. Outputs: -15.
  - (f) `c` 0: `trunc(0 / 5)` = 0, no subtraction. Outputs: 0.
  - Discriminates: (c) rejects floor division (-2) at an exact negative multiple; (b) rejects plain truncation
    (0); (f) rejects subtracting for `c <= 0` (it would give -1).
  - runtime_observed: no

- **NRF-V14** Speed ramp (NRF-R11a; boundary: isolated rule)
  - Inputs: per case; `dirX` 0x1000, `dirY` -0x800, `scale` 0 (so `k` 0x400), `targetSpeed` 0x150
  - Seeds: `no draw`
  - Draws: none
  - Trace: NRF-R11a per call.
  - Outputs: per case
  - (a) Inputs: `speed` 0x100, `accel` 0x40, `tick` 4; six successive calls on the same path. Outputs: call 1:
    `tick` 5, `speed` 0x140, `tick` 0; `s` 0x1400; steps 0x1400 / -0xA00. Call 2: `tick` 1, speed unchanged, same
    steps. Calls 3 to 5: `tick` 2, 3, 4. Call 6: `speed` 0x180 > 0x150 gives `speed` 0x150, `accel` 0, `tick` 0;
    `s` 0x1500; steps 0x1500 / -0xA80.
  - (b) Inputs: `speed` 0x20, `accel` -0x40, `tick` 4; one call. Outputs: sum -0x20 gives `speed` 0, `accel`
    -0x40 retained, `tick` 0; steps 0 / 0.
  - (c) Inputs: `scale` 2048, `speed` 0x101, `accel` 0, `tick` 3, `dirX` 2896, `dirY` -2896; one call. Outputs:
    `tick` 3 and `speed` 0x101 unchanged; `s = (257 * 2048) >> 6` = 8224; `stepX` = `23816704 >> 12` = 5814;
    `stepY` = `-23816704 >> 12` = -5815.
  - (d) Inputs: `scale` -5, `speed` 0x100, `accel` 0, `tick` 0, `dirX` 0x1000, `dirY` 0; one call. Outputs: `k`
    1024; `s` 0x1000; `stepX` 0x1000; `stepY` 0.
  - (e) Inputs: `speed` 0x110, `accel` 0x40, `tick` 4; one call. Outputs: `tick` 5 reaches the ramp; `speed`
    0x150 equals `targetSpeed`, so no clamp: `accel` stays 0x40; `tick` 0; `s` 0x1500; steps 0x1500 / -0xA80.
  - (f) Inputs: `targetSpeed` -10, `speed` 5, `accel` -10, `tick` 4; one call. Outputs: sum -5 is above -10, so
    `speed` -10 and `accel` 0 (the target test precedes the negative test); `tick` 0; `s` = `-10240 >> 6` = -160;
    steps -160 / 80.
  - (g) Inputs: `scale` 1, `speed` -1, `accel` 0, `tick` 0; one call. Outputs: `s` = `-1 >> 6` = -1; `stepX` =
    `-4096 >> 12` = -1; `stepY` = `2048 >> 12` = 0.
  - Discriminates: (a)'s call 2 rejects ramping on every call; call 6 rejects keeping `accel` after the target
    clamp; (b) rejects clearing `accel` at the negative clamp; (c) rejects the default scale when `scale` is
    positive, truncating division for the negative step (-5814), and advancing `tick` when `accel` is 0; (d) rejects using a negative scale (`s` would be -20); (e) rejects
    `>=` in the target test (it would clear `accel`); (f) rejects the negative clamp first (it would give 0); (g)
    rejects truncating division by 64 (`s` would be 0).
  - runtime_observed: no

- **NRF-V15** Prepared integer and movement boundaries (NRF-R06b, NRF-R08b, NRF-R10b, NRF-R11b; boundary: isolated rule)
  - Inputs: per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: per cited rule.
  - Outputs: per case
  - (a) Unsigned `x` 4294967295 (NRF-R06b). Outputs: product high word 1374389534; logical shift by 5 gives
    42949672 = floor(x / 100).
  - (b) Unsigned `x` 99. Outputs: 0.
  - (c) Unsigned `x` 100. Outputs: 1.
  - (d) Speed setter (NRF-R08b): path `speed` 256, `targetSpeed` 512, reset marker 7; set 256. Outputs: marker
    7; `speed` 256; `targetSpeed` 256.
  - (e) Speed setter: path `speed` 256, `targetSpeed` 512, reset marker 7; set 300. Outputs: marker 15; `speed`
    300; `targetSpeed` 300.
  - (f) Speed setter: null path; set 300. Outputs: no effect.
  - (g) Path getters (NRF-R10b): bit patterns `x` 0x000C8000, `y` 0xFFFF0000. Outputs: X 12, Y 65535.
  - (h) Only the NRF-R11a multiplication with supplied components is tested here; the direction lookup itself is
    gated by NRF-Q03. Direction entries (NRF-R11b) with NRF-R11a: axis entry (0, 4096) and diagonal entry (2896, 2896) at
    scale 4096; `speed` 256, `scale` 1024, `accel` 0, so `s` 4096. Outputs: axis steps (0, 4096); diagonal
    steps (2896, 2896).
  - Discriminates: (d) rejects skipping `targetSpeed` on equal speed and (e) omitting the marker on change; (g)
    rejects sign-extending the high word (-1); (a) rejects the signed sequence. (h) asserts only these entries
    and the scale, not the full direction producer.
  - runtime_observed: no

- **NRF-V16** Creation seeding counts and class refusals (NRF-R12a, NRF-R03a; boundary: supplied-input stage)
  - Inputs: per case; MonStatsCount 1; class 0 MonStats record available with `enabled` true unless a
    case states otherwise; successful allocation
  - Seeds: per case
  - Draws: per case
  - Trace: NRF-R12a's class tests, then one unit-seeder game step for every non-player type.
  - Outputs: per case
  - (a) Inputs: `game.seed = {1, 0x29A}`; create a type 1 monster of class 0. Draws: (game, step, lo
    0x6AC6935F). Outputs: `game.seed = {0x6AC6935F, 0}`; unit seed `{0x6AC6935F, 0x29A}`; `initialLo`
    0x6AC6935F. A type 2 object of class 0, a type 5 tile of class 0, or a separate type 3 missile creation of
    class 0, each from `game.seed = {1, 0x29A}`, gives the same outputs.
  - (b) Inputs: `game.seed = {0x6AC6935F, 0}`; create a type 1 monster of class 0. Draws: (game, step, lo
    0x2F2ED81B). Outputs: `game.seed = {0x2F2ED81B, 0x2C890AFD}`; unit seed `{0x2F2ED81B, 0x29A}`; `initialLo`
    0x2F2ED81B.
  - (c) Inputs: `game.seed = {1, 0x29A}`; create a type 0 player of class 0. Draws: none. Outputs:
    `game.seed = {1, 0x29A}`; the player seed is not written.
  - (d) Inputs: `game.seed = {1, 0x29A}`; separately, player class -1, player class 7, monster class -1 and
    monster class 1. Draws: none. Outputs: each returns null before any allocation or seeder request;
    `game.seed = {1, 0x29A}`; no unit seed or copy is written.
  - (e) Inputs: `game.seed = {1, 0x29A}`; separately, a type 2 object of class -1 and one of class 99999. Draws:
    (game, step, lo 0x6AC6935F). Outputs: each is seeded: `game.seed = {0x6AC6935F, 0}`; unit seed
    `{0x6AC6935F, 0x29A}`; `initialLo` 0x6AC6935F.
  - (f) Inputs: `game.seed = {1, 0x29A}`; create a type 1 monster of class 0 whose `enabled` flag is false.
    Draws: none. Outputs: null before any allocation or seeder request; `game.seed = {1, 0x29A}`.
  - (g) Inputs: `game.seed = {1, 0x29A}`; create a type 0 player of class 6. Draws: none. Outputs: the player
    unit is created (non-null); `game.seed = {1, 0x29A}`; the player seed is not written.
  - Discriminates: (e) rejects applying a class range test to objects; (f) rejects omitting the record
    condition; (c) rejects seeding players; (d)'s player class 7 rejects a `> 7` bound and (g)'s class 6 a `>= 6` bound, player class -1 a
    signed comparison, and monster class 1 an inclusive count bound. Items are NRF-V17.
  - runtime_observed: no

- **NRF-V17** Item seeds (NRF-R12a, NRF-R12b; boundary: supplied-input stage)
  - Inputs: successful item creation of type 4, class 0
  - Seeds: `game.seed = {1, 0x29A}`
  - Draws: (game, step, lo 0x6AC6935F); (game, step, lo 0x2F2ED81B)
  - Trace: step 1 seeds the unit; allocate zeroed ItemData with sentinel -1; initialize the secondary seed to
    `{1, 0x29A}`; step 2 seeds ItemData.
  - Outputs: `itemData.sentinel` -1; `game.seed = {0x2F2ED81B, 0x2C890AFD}`; `unit.seed = {0x6AC6935F, 0x29A}`, `unit.initialLo`
    0x6AC6935F (unchanged by step 2); `itemData.initialLo` 0x2F2ED81B; `itemData.seed = {0x2F2ED81B, 0x29A}`;
    exactly 2 game steps. Later item-generation draws are excluded.
  - Ordered local assertions after ItemData allocation: (1) InitTimerState writes `itemData.seed = {1, 666}`
    while `game.seed = {0x6AC6935F, 0}`; (2) the game step changes that pair to
    `{0x2F2ED81B, 0x2C890AFD}` while `itemData.seed` is still `{1, 666}`; (3) write
    `itemData.initialLo = 0x2F2ED81B`; (4) InitSeed writes `itemData.seed = {0x2F2ED81B, 666}`.
  - Discriminates: the distinct secondary seed rejects sharing the unit seed; the unchanged unit copy rejects
    overwriting it in step 2. The ordered local assertions reject moving the initial secondary-seed write after
    the second game step, even though final seed values would match.
  - runtime_observed: no

- **NRF-V18** Prepared object animation rate draw (NRF-R13; boundary: supplied-input stage)
  - Inputs: per case; resolved type 2, mode 1, valid record, total 4096, starting-frame byte 3; no other
    animation preparation is asserted
  - Seeds: per case
  - Draws: per case
  - Trace: `n = baseRate >> 3`; `rate = low32(baseRate - (baseRate >> 4) + roll)`, clamped to 0..32767.
  - Outputs: per case
  - (a) Inputs: `baseRate` 256, fixed-rate condition 0, `object.seed = {1, 0x29A}`. Draws: (object, 32, 31).
    Outputs: frame 768; total 4096; rate `256 - 16 + 31` = 271; `object.seed = {0x6AC6935F, 0}`.
  - (b) Inputs: `baseRate` 256, fixed-rate condition 1, `object.seed = {1, 0x29A}`. Draws: none. Outputs: frame
    768; total 4096; rate 256; seed unchanged.
  - (c) Inputs: `baseRate` 7, fixed-rate condition 0, `object.seed = {1, 0x29A}`. Draws: none (`n` 0). Outputs:
    rate 7; seed unchanged.
  - (d) Inputs: `baseRate` -1, fixed-rate condition 0, `object.seed = {1, 0x29A}`. Draws: none (`n` -1).
    Outputs: rate `-1 - (-1)` = 0; seed unchanged.
  - (e) Inputs: `baseRate` 32767, fixed-rate condition 0, `object.seed = {2, 666}`. Draws: (object, 4095, 3531).
    Outputs: `object.seed = {0xD58D2424, 0}`; raw rate `30720 + 3531` = 34251, clamped to 32767; frame 768.
  - (f) Inputs: `baseRate` -100, fixed-rate condition 0, `object.seed = {1, 0x29A}`. Draws: none (`n` -13).
    Outputs: raw rate `-100 - (-7)` = -93, clamped to 0; stored 16-bit rate 0; seed unchanged.
  - (g) Inputs: `baseRate` 8, fixed-rate condition 0, `object.seed = {1, 0x29A}`. Draws: (object, 1, 0).
    Outputs: rate `8 - 0 + 0` = 8; `object.seed = {0x6AC6935F, 0}`.
  - (h) Inputs: `baseRate` -1, fixed-rate condition 1, `object.seed = {1, 0x29A}`. Draws: none. Outputs: stored
    rate -1 (16-bit 0xFFFF), unclamped; seed unchanged.
  - (i) Inputs: `baseRate` 256, fixed-rate condition 2, `object.seed = {1, 0x29A}`. Draws: none. Outputs: rate
    256; seed unchanged.
  - Discriminates: (a)'s frame 768 rejects applying the roll to the starting frame; (e) rejects omitting the
    upper clamp; (c) rejects stepping for a zero modulus; (f) rejects storing the unclamped 16-bit pattern
    0xFFA3; (g) rejects skipping the step for modulus 1; (h) rejects clamping on the fixed-rate branch; (i) rejects testing the condition for equality with 1. Actual animation preparation remains NRF-Q07.
  - runtime_observed: no

### 8.2 Observation vectors

All are pending gameplay observations. Read-only layout and global reads recorded in the companion do not
complete these vectors.

- **NRF-O01** Measure server attacker seed before/after one melee swing and compare with the owning hit/damage contracts' exact draw count; idle control should not step that server seed. [Procedure](../research/spec-evidence/core-numeric-and-rng-foundations.md#observation-procedures-retained).
- **NRF-O02** Measure frame progress over 60 seconds and original rate globals; expect 1500 frames within sampling error and 25 fps/40 ms. The recorded read covers only the globals. [Procedure](../research/spec-evidence/core-numeric-and-rng-foundations.md#observation-procedures-retained).
- **NRF-O03** Measure integer path displacement for a known map-tile displacement; expect 5 units per tile along either axis. [Procedure](../research/spec-evidence/core-numeric-and-rng-foundations.md#observation-procedures-retained).

## 9. Acceptance and dependencies

Partial acceptance comprises each verified assertion in NRF-V01 to NRF-V18 within its stated supplied-input
boundary, excluding unresolved NRF-R04c/NRF-R09b and the Id map's named integration effects. Full reproduction
additionally requires resolving every in-scope gate, accepting actual dependency-backed stored/queued effects, and
matching observations for the behavioral rules. Passing mocked requests or a pure helper never proves full
creation, combat or animation parity. All vectors remain runtime_observed:no.

**Depends on:** [core-record-layouts-and-ids](core-record-layouts-and-ids.md), RLI-R02a/RLI-R02b existing-array stat read and minimum adjustment consumed by NRF-R07; [core-game-data-fields](core-game-data-fields.md), GDA-R08a MonStats `enabled` flag tested by NRF-R12a and GDF-R02 stat identities. Missing complete animation producer contracts are NRF-Q07; they are not hidden prerequisites to pure numeric fixtures. Owners of the cited HIT, GDA ids: [combat-hit-roll](combat-hit-roll.md) (HIT); [core-game-data-table-records](core-game-data-table-records.md) (GDA).

**Depended on by:** gameplay specs consuming shared arithmetic, units, seeds or draw forms, including the consumer rules named in NRF-R04a's table (SBA, DPO, LSE, CSA, KNB, HRC, MHD, DSC, MSD and SFR rules), which hold the sequences and vectors for their draw sites, [combat-status-application](combat-status-application.md) (seed step), [monster scaling](monster-scaling-by-players-and-difficulty.md) (draw forms, mulDiv, 8.8 format, creation seeding), [damage mitigation](combat-damage-mitigation.md) (mulDiv), [damage pipeline](combat-damage-pipeline-order.md), [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (notation), movement, creation and the related combat/skill specs below. A reverse consumer relationship is not an automatic delivery prerequisite.

## 10. Related specs

- [core-record-layouts-and-ids](core-record-layouts-and-ids.md): record layouts, unit/state/mode identities, the damage record, stat reads and state/mode operations; NRF-R07 reads through RLI-R02a, and its mode setter requests the animation reset whose object draw is NRF-R13.
- [core-game-data-fields](core-game-data-fields.md): the names of the stat and MonStats identities cited here.
- [combat-status-application](combat-status-application.md): the stun, cold, freeze, poison and burn appliers and the status hand-off; NRF-R04a's stun-cancel, shatter and freeze-fallback rows name their seed owners.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): 8.8 life and the damage-build draw order.
- [combat-damage-mitigation](combat-damage-mitigation.md): `mulDiv` percents (NRF-R05).
- [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md): `Diminish` (NRF-R07), the speed setter (NRF-R08b) and 16.16 positions.
- [combat-hit-recovery](combat-hit-recovery.md): stat 0x63, the defender-seed draws and the monster mode-change draws of NRF-R04b.
- [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md): stats 338-340, the `mod 100` draws of NRF-R04a.
- [monster-scaling-by-players-and-difficulty](monster-scaling-by-players-and-difficulty.md): the monster life roll (MSD-R11) and life clamp and write (MSD-R12); NRF-R04a names the new monster's seed as owner.
- [combat-leech-and-drain](combat-leech-and-drain.md): the leech overlay draw of NRF-R04a.
- [combat-knockback](combat-knockback.md): the missile knockback draw of NRF-R04a.
- [missile-movement-and-collision](missile-movement-and-collision.md): the velocity update NRF-R11a and missile seeding.
- [skill-assassin-sentry-firing-and-range](skill-assassin-sentry-firing-and-range.md): the candidate-seed roll of NRF-R04a and the stored-target draw of NRF-R04b.
- [skill-assassin-death-sentry-corpse-damage](skill-assassin-death-sentry-corpse-damage.md): the corpse-seed roll of NRF-R04a.
- [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md): unit creation steps the game seed (NRF-R12a).
- [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md): IEP-R07's draw-free dispatcher, IEP-R13b's direct owner draw and IEP-R14 to IEP-R18's no-direct-draw callback slices support NRF-R04c's boundary; complete handler totals remain NRF-Q04.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. The original
`D2Common.dll`, `D2Game.dll` and `Fog.dll` specimens are identified by exact SHA-256
fingerprints in the [source manifest](../research/source-manifest.md). The manifest also
records the loader binary and the session executable, with their correspondence limits.

The [public evidence companion](../research/spec-evidence/core-numeric-and-rng-foundations.md)
provides module-relative function locations, export ordinals where present, rule-linked
rationale and observation procedures. Static readings were recorded on 2026-10-12/13,
2026-10-14 and 2026-10-15 using original instruction listings and reconstructions
checked against those instructions where recorded. Function descriptions are analyst labels.

Recorded memory reads establish only their stated scope. Gameplay observations remain
pending, and the spec remains draft. Public verification requires matching original game
files, not access to the private reconstruction project.
