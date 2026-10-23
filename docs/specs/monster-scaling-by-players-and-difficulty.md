# Monster scaling by player count and difficulty

**Description**

When a monster appears, the game uses the difficulty, area and eligible player count to set
its level, life and experience value. More eligible players generally give ordinary monsters
more life and experience, and can reduce the chance that a treasure pick drops nothing.
Party members on the killer's map level can improve that drop chance even when they are far
from the kill. The monster's appearance, special properties and starting equipment also use
random choices whose order affects later random results.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: MSD-R, MSD-Q, MSD-V, MSD-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/monster-scaling-by-players-and-difficulty.md)
Open evidence gaps: [gap file](../plans/gaps/monster-scaling-by-players-and-difficulty.md)

## 1. Scope

This spec defines, for an ordinary monster unit created by the server: the effective player
count; the scaling count and the life and experience bonus percents; the difficulty index and
its Normal override; the monster level, including the expansion area level; the base minimum
life, maximum life, armor class and experience from MonStats and MonLvl; the life roll, the
life bonus, the life clamp and the ordered maximum-life and current-life writes; the armor-class,
experience, level and player-count stat writes; the difficulty bonus added to monster skill
levels; and the monster-seed draws made during creation, in execution order (preset build,
appearance, class draw, life roll, property chances, equipment roll), with the post-creation
mode update. It also defines the player-count term of the monster damage setup and, at a
treasure drop, the eligible-player count, the NoDrop weight conversion and the pick draw.
Out of scope, with their owners: the kill experience award and party share
([player-party-experience-sharing](player-party-experience-sharing.md));
seed stepping, draw forms, `mulDiv` and the 8.8 format
([core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md)); stat reads, writes
and state tests ([core-record-layouts-and-ids](core-record-layouts-and-ids.md)); the `unitGate`
status helper ([combat-status-application](combat-status-application.md), CSA-R02);
table fields ([core-game-data-table-records](core-game-data-table-records.md)) and stat, state, mode
and class identities ([core-game-data-fields](core-game-data-fields.md)); skill grants, property
application, item generation, the monster resistance and life-regeneration stats written in the
same creation step, the base monster damage values, treasure item selection and the gold drop.
Player units take another creation path and are outside this boundary.

## 2. Trigger and ownership

- Module: D2Game (server), `MONSTER` monster initialization; D2Common supplies the MonStats,
  MonLvl, Levels, MonProp, MonEquip and DifficultyLevels record lookups, the base-value
  helper, the area-level helper and the stat operations.
- Creation trigger: generic unit creation for unit type 1 steps the game seed to seed the new
  monster (MSD-R01), then calls the monster initializer, which looks up or builds the
  appearance preset (MSD-R02) and runs the monster stat initialization (MSD-R03 to MSD-R18).
  When the caller's creation condition is set, generic creation then runs the post-creation
  mode update (MSD-R19). This runs once per created monster, inside the creating server frame.
- Damage term: the monster damage setup reached on mode changes, including the post-creation
  mode update, evaluates MSD-R20.
- Drop trigger: with a nonzero prepared treasure total `T`, the treasure-class drop routine
  runs MSD-R21 to MSD-R23 once per pick; the killer is supplied by the caller. A zero total
  exits this treasure-class level before those stages.
- These stages request local stat writes and send no direct client message. Stored-write
  integration remains RLI-Q03; update bookkeeping remains RLI-Q02. Client delivery is outside
  this verified boundary; message contracts belong to `net-` specs.
- Dispatch chains, function references and instruction locations are in the
  [companion](../research/spec-evidence/monster-scaling-by-players-and-difficulty.md#dispatch-and-function-references).

## 3. Inputs and data

Operation names are illustrative. `Rand` and the inline step are NRF-R01/NRF-R02; `mulDiv` is
NRF-R05; `HasState` is RLI-R03; `GetStat(unit, id)` is the layer-0 read of RLI-R02a;
`SetStat(unit, id, value)` is a layer-0 replacing write request (stored effect RLI-Q03);
`unitGate` is CSA-R02. Signed `/` truncates toward zero; additions keep the low 32 bits.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| player units | the game's player-unit table | logical set; each entry's type is tested | Normally player (type 0) units; MSD-R05 also defines entries of other types |
| player-unit conditions | `HasState(u, 7)`; `u.aliveExclusionCondition`; `u.type`; `u.mode` (GDF-R05a, GDF-R06a) | boolean; i32 | `aliveExclusionCondition` is a unit condition the alive test treats as not alive (MSD-Q01); the shared name is defined in [core-game-data-fields](core-game-data-fields.md) section 3 and PXL-R04a and GPS-R01 test the same condition; modes as named |
| player-count override | one process-wide value shared by every game in the server process, set by the client player-count command | i32, 0..8, 0 at load | Read at each MSD-R05 evaluation; its producer is outside this boundary |
| `game.selectorByte`, `game.selectorFlag` | game-creation argument byte; game-creation flag | u8; i32 0 or 1 | Neutral names; roles in MSD-Q01 |
| `game.difficulty` | game difficulty | u8 | 0 Normal, 1 Nightmare, 2 Hell |
| `game.expansion` | game expansion condition | i32 | Nonzero is an expansion game |
| `game.seed`, `monster.seed`, `dropper.seed` | NRF-R01 seeds | two u32 | `dropper` is the unit whose treasure class drops |
| monster class | `monster.classId`; its MonStats row (GDA-R08a) | i32 | MonStats fields below |
| creation eligibility | MonStats record availability and its per-class creation eligibility under NRF-R12a | boolean | Tested before allocation or seeding; distinct from the post-initialization `creationCondition` below |
| table collections | MonStats, MonStats2, MonLvl, Levels, MonProp, MonEquip and DifficultyLevels collections and their counts | indexed logical records; i32 counts | Row availability and bounds used by the rules; DifficultyLevels uses GDA-R05's accepted three-row collection |
| preset storage | game presets by room level id; optional level record with `entryCount`; ordered entries with stored class, `variantCount` and variants | association; u8 counts; i32 class; variants of 16 u8 components | Supplied existing storage for MSD-R02/MSD-R03; initialization of an unregistered slot is MSD-Q11 |
| MonStats fields | `Align`, `noRatio`, `boss`, `Level`/`Level(N)`/`Level(H)`, `MinHP`, `MaxHP`, `AC`, `Exp` (each with `(N)`, `(H)`), `Skill1..8`, `Sk1lvl..Sk8lvl`, `MonStatsEx`, `MonProp` | u8; booleans; the GDA-R08a 16-bit words read sign-extended as i16 (a word of 0x8000 or more is negative); i16 links; signed bytes for skill levels | Difficulty `d` selects base, `(N)` or `(H)` |
| MonEquip first row | prepared per-class index of the first MonEquip row whose `monster` is this class | i16; -1 when none | Built at load (companion) |
| MonLvl row | MonLvl row `min(level, rowCount - 1)`; columns `HP`, `AC`, `XP` and `L-HP`, `L-AC`, `L-XP`, each with `(N)`, `(H)` | i32 cells | Base-value multipliers in percent |
| Levels row | Levels row by level id; `MonLvl2Ex`, `MonLvl3Ex` | i16 | Expansion area level for `d = 1`, `d = 2` |
| creation room | supplied room; `room.levelId` | association or null; i32 | Level id 0 when the room is null |
| MonStats2 preset inputs | `MonStatsEx` row: `TotalPieces`; prepared per-slot component counts `cnt[0..15]`; prepared variant-limit byte `v` | u8 each | `cnt` and `v` are prepared values, not text columns (MSD-Q06) |
| MonProp row | `MonProp` row: `prop1..prop6` and `chance1..chance6` in base, `(N)`, `(H)` blocks | i32 property link (negative means none); u8 chance | Property slots of MSD-R15 |
| MonEquip rows | `monster`, `level`, `oninit`, `item1..3`, `loc1..3`, `mod1..3` | i16; signed i16; u8; item code; i8; i8 | Equipment rows of MSD-R16, consecutive in table order |
| monster inventory | `monster.inventory` | association or null | Null means the monster has no inventory |
| `creationCondition` | caller-supplied creation condition of generic unit creation | boolean | Selects the post-creation mode update (MSD-R19); role outside this boundary |
| `MonsterSkillBonus` | DifficultyLevels (GDA-R05), row `clamp(game.difficulty, 0, rowCount - 1)` | i32 | Added to monster skill levels |
| treasure pick | `Picks`, `NoDrop` = `N` (GDA-R15a/GDA-R15b); prepared probability total `T` (classic total when `game.expansion == 0`, else expansion total) | i32 | `T` is produced outside this boundary |
| drop arguments | `killer` (unit or null); `dropper`; `noDropBypass` | association; boolean | `noDropBypass` is a caller-supplied condition (MSD-Q08) |
| killer owner | `owner(k)`: the owner unit of a non-player unit (for example a summon), or null | association or null | "Not a player" means a null result or a unit whose type is not 0 |
| killer party | `k.partyId` (0xFFFF means none; also none when `k` has no player data); `game.partyRecord(partyId)`, null when the game holds no record with that id; the record's member ids; each member's room level id | u16; association | The PXL-R04a enumeration terms; party identity is owned by the party system |
| damage-setup values | minimum damage, maximum damage and attack rating as they enter the term step (after the classic-game difficulty adjustment in a classic game); `damageTermSkip` | i32 in the units supplied by the damage setup; boolean | Supplied by the monster damage setup; units and producer are MSD-Q05 |
| stage results | `E`; `n`, `lifePct`, `xpPct`; `d`; `level`; `minHP`, `maxHP`, `ac`, `exp`; `hp`; `eligible`; `N'` | i32 | Produced in rule order by MSD-R05, MSD-R06, MSD-R07, MSD-R08, MSD-R10, MSD-R11, MSD-R21 and MSD-R22a; isolated vectors may supply the result at their stated entry |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Bonus tables `0, 0, 50, 100, 150, 200, 250, 300, 350` (life and experience, separate tables with equal contents) and the formulas for `n >= 9` | Engine code / image constants | Fixed by original code |
| Damage-term table `0, 0, 8, 16, 24, 32, 40, 48, 56` and formula `8n - 16` for `n >= 9` | Engine code / image constants | Fixed by original code |
| Life clamp `0x7FFFFF` at `0x800000`; 8.8 shift | Engine code | Fixed by original code (MSD-R12; 8.8 format NRF-R08a) |
| Classes 311 and 312 for the class draw; classes 195, 196, 294, 296 without a preset; preset entry limit 13; property chance modulus 100; MonEquip location range 1..10 and mod range 0..7 | Engine code | Fixed by original code; MonStats row identities per GDF-R08 |
| State 7 for the player-count exclusion; stat ids 6, 7, 12, 13, 19, 21, 22, 31, 100 | Engine code selects original-profile identities (GDF-R02, GDF-R03) | Fixed by original code |
| MonLvl, Levels, MonProp, MonEquip, MonStats and DifficultyLevels named columns | Engine loader descriptors bind table column names to prepared fields (companion records); actual cells are table data | Code fixes the consumed column and interpretation; cell values are resolved from data |

All numeric data values used by the vectors are supplied fixture values, not cited cells of the
original tables.

## 4. Rules (execution order)

MSD-R01 to MSD-R19 run in this order for one created monster. MSD-R20 runs in the damage setup.
MSD-R21 to MSD-R23 run per treasure pick. Each stat write is a `SetStat` request; the stored
effect of the replacing write is the RLI-Q03 integration gate.

**MSD-R01** **Creation seeding** [verified in 1.13c code (static)]. After the monster class,
record availability and per-class creation eligibility pass NRF-R12a, successful allocation
steps `game.seed` once and initializes
`monster.seed` to `{game.seed.lo, 666}` with `monster.initialLo = game.seed.lo` (NRF-R12a,
NRF-R03a). No other game-seed step precedes the monster's own draws below. A class that
NRF-R12a rejects creates no monster; no rule of this spec runs for it (NRF-V16 (d)).

**MSD-R02** **Appearance preset lookup and build** [verified in 1.13c code (static)]. The
initializer looks up the creation level's preset entry for `classId` before stat initialization:

```
if classId in {195, 196, 294, 296}: return null                  // no draw
levelRec = presets of game for room.levelId; if levelRec == null: return null
entry = first of levelRec's entries 0..entryCount-1 (registration order) whose stored class equals classId
if entry: return entry                                          // no draw
if levelRec.entryCount >= 13: return null
rec2 = MonStats2 row of MonStatsEx(classId); if rec2 == null or rec2.TotalPieces <= 2:
    return entry slot entryCount without registering it          // entryCount unchanged
register a new entry for classId; build(entry, monster.seed, request 3); return entry
build:
    existing = entry.variantCount
    req = min(3, (1 << v) - existing, 3 - existing); if req <= 0: return
    if existing == 0:
        for i in 0..15: variant0[i] = cnt[i] <= 1 ? 0 : Rand(monster.seed, cnt[i])
        variantCount = 1; req -= 1
    if req <= 0: return
    list = slots i in order with cnt[i] > 1; k = size(list)
    if k == 0: return
    if k == 1: A = B = list[0]
    else: a = Rand(monster.seed, k); A = list[a]; list[a] = list[k-1]
          B = list[Rand(monster.seed, k - 1)]
    repeat req times:
        variant = copy of variant0
        for attempt in 1..3:
            variant[A] = step-then-select(monster.seed, cnt[A])
            if B != A: variant[B] = step-then-select(monster.seed, cnt[B])
            if variant differs from every stored variant: break attempts
        store variant; variantCount += 1                       // a duplicate is kept after 3 attempts
```

`cnt[i] <= 1` is an unsigned test. `step-then-select(seed, c)` is one inline step, then
`lo & (c-1)` when `c` is a power of two, else `lo mod c` (unsigned), the same result as
`Rand(seed, c)` for `c > 0`. The build uses the seed of the first monster of that class on that
level; later monsters of the class on the level receive the stored entry with no build draw. An
unregistered slot leaves `entryCount` unchanged, so later monsters of that class repeat the same
lookup and receive the same unregistered slot; its variant count is 0 when the level record's
entry storage starts zeroed (MSD-Q11), which sends MSD-R03 to the per-slot path.

**MSD-R03** **Appearance** [verified in 1.13c code (static)]. With a non-null preset whose
variant count `c` (u8) is positive, one `Rand(monster.seed, c)` selects the stored 16-byte
variant copied to the monster's appearance. Otherwise, for slot i = 0..15 in order, the slot's
component count `c` is read; `c > 0` takes one draw `step-then-select(monster.seed, c)`, and
`c <= 0` stores 0 with no draw. The first monster pays MSD-R02's build draws before this draw.

**MSD-R04** **Class draw** [verified in 1.13c code (static)]. Class 311 takes one
`Rand(monster.seed, 4)` and stores the result over appearance byte 10 (zero-based) of the 16
bytes written by MSD-R03; class 312 takes one draw and stores it over bytes 10 and 11. Other classes take no draw. (A class outside
`0 <= classId < MonStatsCount` would take no draw; generic creation never reaches it, MSD-R01.)

**MSD-R05** **Effective player count** [verified in 1.13c code (static)]. No draw.

```
E = 0
for each unit u in the game's player-unit table:
    if HasState(u, 7): continue
    if u.aliveExclusionCondition: continue
    if u.type == 0 and u.mode in {0, 17}: continue
    if u.type == 1 and u.mode in {0, 12}: continue
    if u.type not in {0, 1}: continue
    E += 1
if 1 <= game.selectorByte <= 3 and override > E: E = override
```

The same function supplies `E` to MSD-R06 at creation and to MSD-R21 at each drop, read live.

**MSD-R06** **Scaling count and bonus percents** [verified in 1.13c code (static)]. No draw.

```
if MonStats.Align != 0: n = 1; lifePct = 0; xpPct = 0
else:
    n = max(1, E)
    lifePct = n < 9 ? T[n] : low32((n - 2) * 50)
    xpPct   = n < 9 ? T[n] : low32((5*n + 130) * 2)
T = [0, 0, 50, 100, 150, 200, 250, 300, 350]
```

Comparisons are signed. Neither percent depends on difficulty. In the same step the
`MonsterSkillBonus` of MSD-R14 is read from the DifficultyLevels row
`clamp(game.difficulty, 0, rowCount - 1)`; then, if `game.difficulty >= 3`, `game.difficulty` is set to 2.

**MSD-R07** **Difficulty index** [verified in 1.13c code (static)]. No draw.
`d = game.difficulty <= 0 ? 0 : min(game.difficulty, 2)`; then if `unitGate(monster) != 0`
(classes 271, 338, 359, 560, 561 under CSA-R02), `d = 0`. MSD-R08, MSD-R10 and MSD-R15 use `d`.
MSD-R14's bonus row is not affected by this override.

**MSD-R08** **Monster level** [verified in 1.13c code (static)]. No draw.

```
level = MonStats.Level[d]                                       // Level, Level(N), Level(H), i16
if game.expansion != 0 and d > 0 and not MonStats.noRatio and not MonStats.boss:
    id = room == null ? 0 : room.levelId
    level = (id <= 0 or id >= LevelsCount) ? 1 : (d == 1 ? Levels[id].MonLvl2Ex : Levels[id].MonLvl3Ex)
```

**MSD-R09** **Level and player-count stats** [verified in 1.13c code (static)]. Request
`SetStat(monster, 12, level)` then `SetStat(monster, 100, max(1, n))`. No draw.

**MSD-R10** **Base values** [verified in 1.13c code (static)]. No draw.
`sel = (game.selectorByte != 0 or game.selectorFlag != 0) ? 1 : 0`. The MonLvl row is
`r = min(level, rowCount - 1)` (signed); if `r < 0` or the row is unavailable, all four base
values are 0. Otherwise, for each value with its MonStats word `w` (index `d`) and MonLvl column:

| Value | MonStats word | MonLvl column for `sel = 0` / `sel = 1` |
|---|---|---|
| `minHP` | `MinHP[d]` | `HP[d]` / `L-HP[d]` |
| `maxHP` | `MaxHP[d]` | `HP[d]` / `L-HP[d]` |
| `ac` | `AC[d]` | `AC[d]` / `L-AC[d]` |
| `exp` | `Exp[d]` | `XP[d]` / `L-XP[d]` |

the value is `w` when `MonStats.noRatio` is set, else `mulDiv(w, MonLvl[r].column, 100)`.

**MSD-R11** **Life roll and bonus** [verified in 1.13c code (static)].

```
range = low32(maxHP - minHP + 1)
hp = low32(minHP + Rand(monster.seed, range))     // range <= 0: no step, adds 0
hp = low32(hp + mulDiv(lifePct, hp, 100))
```

One draw when `range > 0`, none otherwise (NRF-R02). The operand order of `mulDiv` is
`(lifePct, hp, 100)` (NRF-R05).

**MSD-R12** **Life clamp and write** [verified in 1.13c code (static)]. If `hp >= 0x800000`
(signed), `hp = 0x7FFFFF`. Then `raw = low32(hp << 8)`; request `SetStat(monster, 7, raw)`
followed by `SetStat(monster, 6, raw)`. Negative `hp` is not clamped. No draw. Values use the
8.8 format of NRF-R08a; actual stored-write integration is RLI-Q03.

**MSD-R13** **Armor class and experience** [verified in 1.13c code (static)]. Request
`SetStat(monster, 31, ac)`, then `xp = low32(exp + mulDiv(xpPct, exp, 100))` and request
`SetStat(monster, 13, xp)`. No draw.

**MSD-R14** **Difficulty skill-level bonus** [verified in 1.13c code (static)]. No draw in this
rule. For i = 1..8 in order: `skill = MonStats.Skill{i}` (i16), `lvl = MonStats.Sk{i}lvl` (signed
byte); if `skill < 0` or `lvl <= 0`, continue; otherwise request
`GrantSkill(monster, skill, low32(lvl + MonsterSkillBonus))`; a skill id of 0 is granted. The grant and the skill-mode
assignment that follows it are outside this boundary.

**MSD-R15** **Property chances** [verified in 1.13c code (static)]. If `MonStats.MonProp` is
negative or at least the MonProp row count, no slot runs. Otherwise, for slot i = 1..6 in the
`d` block (base, `(N)`, `(H)`):

```
if prop{i} < 0: break
if chance{i} == 0: request ApplyProperty(monster, slot i)        // no draw
else:
    step(monster.seed); roll = monster.seed.lo mod 100           // unsigned
    if roll < chance{i}: request ApplyProperty(monster, slot i)
```

Every nonzero chance takes its draw, including chances of 100 or more, which then always apply.
Zero to six draws. Property application is outside this boundary.

**MSD-R16** **Equipment roll** [verified in 1.13c code (static)]. Runs last in stat
initialization with the monster level from MSD-R08:

```
row = MonEquip first row of classId; if none: return
if row.oninit == 0: return
if monster.inventory == null: return
while row.monster == classId and row.level > level: row = next row     // signed comparison
while row.monster == classId:
    k = number of leading loc1..loc3 values in 1..10 (signed), stopping at the first outside
    if k > 0:
        step(monster.seed); j = (k & (k-1)) == 0 ? lo & (k-1) : lo mod k
        if location loc{j+1} of the inventory is free:
            m = mod{j+1}; if m < 0 or m >= 8: m = 0
            request CreateEquipmentItem(item{j+1}, level, m)
    row = next row
```

One draw per processed row with `k > 0`, including `k == 1`. Only the first row's `oninit` is
tested. The walk has no row-count bound; this contract covers data in which the class's rows are
followed by a row of another class (MSD-Q10). The location test and item creation are outside
this boundary.

**MSD-R17** **Equipment item seeding** [asm-only]. Each requested equipment item reaches generic
item creation, which steps `game.seed` twice under NRF-R12a/NRF-R12b. Its item-generation draws
use the new item seeds and belong to item generation (MSD-Q04).

**MSD-R18** **Other initialization callees** [asm-only]. The remaining calls of stat
initialization (stat-list setup, inventory setup, skill grants, flag reads and property
application) take no monster-seed or game-seed draw at creation, except through two unresolved
indirect calls in the stat-list operations (MSD-Q04).

**MSD-R19** **Post-creation mode update** [asm-only]. When `creationCondition` is set, the
mode update after initialization reaches the monster damage setup, which takes the monster-seed
draws of NRF-R04b at elemental mode slots; eligibility is NRF-Q05 (MSD-Q04).

**MSD-R20** **Damage player-count term** [verified in 1.13c code (static)]. No draw.

```
if damageTermSkip: term = 0
else:
    p = max(1, GetStat(monster, 100))
    term = (game.difficulty == 0 or p < 2) ? 0 : (p < 9 ? D[p] : low32(8*p - 16))
D = [0, 0, 8, 16, 24, 32, 40, 48, 56]
if game.expansion != 0 and term != 0:
    minDamage += trunc128(minDamage * term); maxDamage += trunc128(maxDamage * term)
    toHit += trunc128(toHit * term)
request writes of stat 21 = minDamage, 22 = maxDamage, 19 = toHit       // always, in this order
```

`trunc128(x)` is the truncating `/128` of NRF-R06a on the low-32-bit product. Production of the
base values, `damageTermSkip` and the classic-game difficulty adjustment are outside this rule
(MSD-Q05).

**MSD-R21** **NoDrop eligible count** [verified in 1.13c code (static)]. The
treasure-class level exits before the pick loop when its prepared total `T` is 0. Otherwise
this count is evaluated once per pick, only when `N != 0` and `Picks >= 0`, with `E` and the
party read again for each pick. No draw.

```
k = killer
if k != null and k.type != 0: k = owner(k)                     // owner lookup; may be null
party = (k == null or k.type != 0) ? 0 : PartyCount(k)
elig = party <= 1 ? 1 : min(party, 8)
eligible = elig + (E - elig) / 2                               // signed, truncates toward zero
if dropper != null and dropper.type == 1:
    eligible = min(eligible, max(1, GetStat(dropper, 100)))
PartyCount(k):
    if k has no room: return alive(k) ? 1 : 0
    if k.partyId == 0xFFFF: return alive(k) ? 1 : 0             // includes no player data
    if game.partyRecord(k.partyId) == null: return 0
    count members m of k's party found in the game whose room is non-null,
    whose room.levelId equals k's room.levelId and for which alive(m) holds
```

`alive(u)` fails for a unit with `aliveExclusionCondition`, for type 0 in mode 0 or 17, for type 1 in
mode 0 or 12, and for other types; unlike MSD-R05 it does not test state 7. Only PXL-R04a's enumeration stage applies; its experience
member filters (victim distance, zero position, the eight-member limit) are not part of this
count. There is no distance test, and `elig` is not clamped to `E`.

**MSD-R22** **NoDrop weight conversion**
- **MSD-R22a** [verified in 1.13c code (static)]. If `N != 0` and `eligible > 1`:

  ```
  r = double(N) / double(low32(N + T))
  p = r^eligible, computed as r multiplied by r, (eligible - 1) times, left to right
  q = 1.0 - p
  if q == 0.0: N' = 0
  else: N' = low32(truncate_toward_zero(((1.0 - q) / q) * T))
  ```

  Otherwise `N' = N`. No draw. The contract covers `N > 0`, `T > 0` and `N + T` without signed
  overflow; other inputs, including a zero divisor or a value outside the integer range at
  truncation, are MSD-Q02.
- **MSD-R22b** [unresolved]. The rounding precision of the intermediate products, quotients and
  differences of MSD-R22a is not established (MSD-Q02); results whose truncated value depends on
  that rounding are not predicted.

**MSD-R23** **Pick draw** [verified in 1.13c code (static)].

```
if Picks < 0: no NoDrop and no draw for this pick                // negative-pick selection is outside
else if noDropBypass: roll = Rand(dropper.seed, T); select item by roll
else:
    roll = Rand(dropper.seed, low32(T + N'))
    if roll < N': NoDrop; the pick is consumed
    else: select item by (roll - N')
```

Evaluated once per pick when the prepared total `T` is nonzero: one draw per non-negative pick
on the dropping unit's own seed; there is no separate treasure seed. Item selection is outside
this boundary.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MSD-R01 | One game-seed step seeds the monster | verified in 1.13c code (static) | none | MSD-V01 |
| MSD-R02 | Preset lookup; first-monster build draws | verified in 1.13c code (static) | MSD-Q06: real prepared `cnt`/`v` values; MSD-Q11: unregistered-slot variant count; supplied values eligible | MSD-V02, MSD-V21 |
| MSD-R03 | Appearance draw: one from a preset, else 0..16 | verified in 1.13c code (static) | none | MSD-V02, MSD-V03, MSD-V21 |
| MSD-R04 | Class 311/312 Rand(4) | verified in 1.13c code (static) | none | MSD-V04, MSD-V21 |
| MSD-R05 | Effective player count with override | verified in 1.13c code (static) | none | MSD-V05 |
| MSD-R06 | Scaling count; life/experience percents; difficulty rewrite | verified in 1.13c code (static) | none | MSD-V06 |
| MSD-R07 | Difficulty index; Normal override | verified in 1.13c code (static) | none | MSD-V07 |
| MSD-R08 | Monster level; expansion area level | verified in 1.13c code (static) | none | MSD-V08 |
| MSD-R09 | Stat 12 then stat 100 write requests | verified in 1.13c code (static) | RLI-Q03: stored-write integration, not the requested values or order | MSD-V09, MSD-V21 |
| MSD-R10 | Base values from MonStats and MonLvl | verified in 1.13c code (static) | none | MSD-V10 |
| MSD-R11 | Life roll and life bonus | verified in 1.13c code (static) | none | MSD-V11, MSD-V21 |
| MSD-R12 | Life clamp; stat 7 then stat 6 write requests | verified in 1.13c code (static) | RLI-Q03: stored-write integration, not clamp, shift or order | MSD-V12, MSD-V21 |
| MSD-R13 | Armor class and experience write requests | verified in 1.13c code (static) | RLI-Q03: stored-write integration, not the values or order | MSD-V13, MSD-V21 |
| MSD-R14 | Skill levels plus MonsterSkillBonus | verified in 1.13c code (static) | none for the requested level; the grant is outside | MSD-V14 |
| MSD-R15 | Property chance draws, 0..6 | verified in 1.13c code (static) | none | MSD-V15, MSD-V21 |
| MSD-R16 | Equipment roll: one draw per processed row | verified in 1.13c code (static) | MSD-Q10: tables whose class rows reach the end; supplied terminated rows eligible | MSD-V16, MSD-V21 |
| MSD-R17 | Equipment item seeding | asm-only | MSD-Q04 | none |
| MSD-R18 | No other creation draws except two indirect calls | asm-only | MSD-Q04 | none |
| MSD-R19 | Post-creation mode update draws | asm-only | MSD-Q04 | none |
| MSD-R20 | Damage player-count term and its expansion application | verified in 1.13c code (static) | MSD-Q05: real base values, skip condition and write integration; supplied values eligible | MSD-V17 |
| MSD-R21 | NoDrop eligible count | verified in 1.13c code (static) | none | MSD-V18 |
| MSD-R22a | NoDrop weight conversion sequence | verified in 1.13c code (static) | MSD-Q02: results that depend on intermediate rounding and inputs outside the covered domain | MSD-V19 |
| MSD-R22b | Intermediate rounding precision | unresolved | MSD-Q02 | none |
| MSD-R23 | Pick draw on the dropping unit's seed; bypass; negative picks | verified in 1.13c code (static) | none | MSD-V20 |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| `game.seed` stepped once; `monster.seed = {game.seed.lo, 666}`; `monster.initialLo` | MSD-R01 |
| Preset entry registered and built for the first monster of a class on a level; monster-seed draws | MSD-R02 |
| Appearance bytes; monster-seed draws | MSD-R03 |
| Appearance byte 10 (classes 311, 312) and byte 11 (class 312) overwritten by one drawn value; one monster-seed draw | MSD-R04 |
| `game.difficulty` rewritten to 2 when 3 or more | MSD-R06 |
| Stat 12 `level`, then stat 100 `monster_playercount` write requests | MSD-R09 |
| One monster-seed draw (when `range > 0`) | MSD-R11 |
| Stat 7 `maxhp`, then stat 6 `hitpoints` write requests, raw 8.8 | MSD-R12 |
| Stat 31 `armorclass`, then stat 13 `experience` write requests | MSD-R13 |
| Skill grant requests with bonused levels | MSD-R14 |
| Property application requests; 0..6 monster-seed steps | MSD-R15 |
| Equipment item creation requests; one monster-seed step per processed row | MSD-R16 |
| Game-seed steps of created equipment items | MSD-R17 |
| Monster-seed draws of the post-creation mode update | MSD-R19 |
| Stat 21 `mindamage`, 22 `maxdamage`, 19 `tohit` write requests | MSD-R20 |
| NoDrop outcome or item-selection roll; one `dropper.seed` draw per non-negative pick when `T != 0`; pick consumed on NoDrop | MSD-R23 |

No direct message is sent by these rules. Their stat-write requests do not establish stored
effects (RLI-Q03), update bookkeeping (RLI-Q02) or client delivery.

Named stage results may also be asserted: the preset entry or null (MSD-R02); effective count
`E` (MSD-R05); `n`, both bonus percents and the selected skill bonus (MSD-R06); `d`
(MSD-R07); level (MSD-R08); four base values (MSD-R10); rolled and bonused `hp` (MSD-R11),
clamped `hp` and raw life (MSD-R12); bonused experience (MSD-R13); damage term and adjusted
values (MSD-R20); party and eligible counts (MSD-R21); converted `N'` (MSD-R22a); and pick
roll or item-selection roll (MSD-R23). These are semantic results, not required public APIs.
Skill, property and equipment requests assert their arguments only; no result from those
outside operations is consumed here. The creation and drop procedures make no further
numeric-return assertion.

## 6. Branches and edge cases

- Unit type: the stat initialization returns without effect for a non-monster unit, a missing
  monster data record or MonStats row, or a class whose MonStats2 row (`MonStatsEx`) is
  unavailable; then no rule from MSD-R03 on runs and no stats are written.
- `Align != 0`: `n = 1` and both percents are 0; stat 100 is 1 and caps MSD-R21's `eligible` at 1.
- Difficulty: only `d`-indexed columns, the area level and the damage term change; the life and
  experience percents and the NoDrop conversion have no difficulty input. Classes under
  `unitGate` use Normal columns in every difficulty, while MSD-R14 still uses the game
  difficulty row.
- Empty life range (`maxHP < minHP`): no draw and `hp = minHP` before the bonus.
- Life at or above `0x800000` after the bonus clamps to `0x7FFFFF`; negative life is written
  unclamped as `low32(hp << 8)`.
- Negative or unavailable MonLvl row: all base values 0, so `range = 1`: one draw returning 0.
- Override: applies only when `game.selectorByte` is 1, 2 or 3; it can raise but never lower `E`.
- Zero living players: `E = 0` at creation gives `n = 1`; at a drop `eligible = elig + (0 - elig) / 2`.
- Party larger than `E` (for example when the override is off): `(E - elig) / 2` is negative, so
  `eligible` can fall below the party size (party 5, `E` 3 gives 4).
- Killer: a non-player killer is replaced by its owner; a missing or non-player owner gives
  `elig = 1`. Party members on another level, without a room, or not alive are not counted.
- `T = 0`: the treasure-class level exits before the NoDrop count or a pick draw, including
  when `noDropBypass` is set.
- `N = 0` with `T != 0`: no conversion and no eligibility computation; the draw is `Rand(T)`
  and never NoDrop.
- `Picks < 0`: no NoDrop and no draw for that pick.
- Preset sharing: only the first monster of its class on a level pays the build draws; later
  ones make one appearance draw from the stored entry.
- No recompute of stat 100, life or experience after a player joins or leaves is established
  (MSD-Q03).

## 7. Open questions

| Question | Kind / status | Evidence that would settle it |
|---|---|---|
| **MSD-Q01** | outside-boundary / open | Roles of `game.selectorByte`, `game.selectorFlag` and `aliveExclusionCondition`, from their game-creation and unit-state producers; the rules use their tested values only |
| **MSD-Q02** | behavior / unresolved | The floating-point precision control in effect when the conversion runs on the original server, and the conversion results for inputs outside MSD-R22a's covered domain |
| **MSD-Q03** | behavior / unresolved | An exhaustive writer search for stats 100, 6, 7 and 13 on monsters, including undecoded regions, to establish whether any path recomputes them after creation |
| **MSD-Q04** | integration / open | Reading the item-creation path called by the equipment roll, the two indirect calls of the stat-list operations and the post-creation mode-update chain, with NRF-Q05 for elemental-slot eligibility |
| **MSD-Q05** | outside-boundary / open | The monster damage setup's base damage and attack rating, `damageTermSkip` and the stat-write operation used for stats 19, 21, 22, under a monster damage owner |
| **MSD-Q06** | outside-boundary / open | Preparation of the MonStats2 per-slot component counts and the variant-limit byte from the component columns, under the game-data owner |
| **MSD-Q07** | observation / pending | Gameplay observations MSD-O01 to MSD-O06 on an owned 1.13c game; none were made |
| **MSD-Q08** | outside-boundary / open | Which drop callers pass `noDropBypass` and why; the rule uses the supplied condition |
| **MSD-Q09** | outside-boundary / open | Whether gold amount or magic find has a player-count term, under the treasure and gold owner; no search was completed |
| **MSD-Q10** | outside-boundary / open | Whether the prepared MonEquip storage guarantees a following row of another class after a class's last row; the equipment walk has no count bound |
| **MSD-Q11** | behavior / unresolved | Initialization of a level record's preset entry storage, which fixes the variant count of an unregistered slot; vectors supply 0 |

The [gap file](../plans/gaps/monster-scaling-by-players-and-difficulty.md) records what each
open question's attempt covered.

## 8. Test vectors

### 8.1 Executable vectors

Defaults unless a case overrides them: one monster of type 1 with a valid class and records;
supplied fixture data values; `game.expansion = 0`; `game.difficulty = 0`; `selectorByte = 0`;
`selectorFlag = 0`; override 0; `Align = 0`; `noRatio` and `boss` clear; `unitGate = 0` unless
a case names a nonzero-return class under CSA-R02 or supplies the gate result. The generic
creation fixture supplies a present, eligible MonStats record and successful allocation under
NRF-R12a; this eligibility does not set the separate post-initialization `creationCondition`.
DifficultyLevels is an accepted three-row collection; bonus cells are 0 unless specified.
Every vector is a supplied-input stage unless it states otherwise, and every case starts from its own
listed state; no case inherits another case's state or outputs. `M` is the monster seed
`{0x6AC6935F, 0x29A}` (the output of MSD-V01), used fresh by each case that names it.

Supplied operations: `SetStat` logs its calls in order and persists the stored raw value for the
rest of the vector; `GetStat` returns the supplied stored value; `HasState`, `unitGate`, the
inventory location test and `damageTermSkip` return supplied values and do not log; `GrantSkill`,
`ApplyProperty` and `CreateEquipmentItem` log their calls in order and persist nothing. Logged
arguments asserted: `GrantSkill(skill, level)`, `ApplyProperty(block d, slot i)` (the slot's
property fields are passed through and not asserted) and `CreateEquipmentItem(item, level, mod)`.
The preset level record, when present, holds the entries listed by each case in registration
order. All rules owned within a vector's stated boundary execute; the supplied dependency
operations above do not verify their producers or stored integration. In particular, the
fixture's persisted stat value is a supplied effect, not evidence that RLI-Q03 is resolved.
Outputs may assert named intermediate results. Each vector has `runtime_observed: no`.

- **MSD-V01** Creation seeding (MSD-R01; boundary: supplied-input stage)
  - Inputs: `game.seed = {1, 0x29A}`; valid monster class
  - Seeds: game `{1, 0x29A}`
  - Draws: one game-seed step (NRF-R01)
  - Trace: `x = 1*0x6AC690C5 + 0x29A = 0x6AC6935F`; game `{0x6AC6935F, 0}`; monster `{0x6AC6935F, 666}`
  - Outputs: `game.seed = {0x6AC6935F, 0}`; `monster.seed = {0x6AC6935F, 0x29A}`; `initialLo = 0x6AC6935F`
  - Discriminates: copying the whole game seed would give `hi = 0`; no step would give `lo = 1`
  - runtime_observed: no

- **MSD-V02** Preset build and appearance (MSD-R02, MSD-R03; boundary: supplied-input stage)
  - Inputs, shared: class 100 (not 195, 196, 294 or 296); room level id 5 with a level preset record; MonStats2 row of the class with `TotalPieces = 3`. In (a), (b), (d), (e) and (f) the record holds one entry, for class 50 (`entryCount = 1`), so class 100 registers as entry 1; in (c) it holds entry 0 for class 50 and entry 1 for class 100
  - (a) `cnt = 4, 3, 1`, other 13 slots 0; `v = 0`. Seeds: `M`. Draws: `Rand(4)` → `{0x2F2EDAB5, 0x2C890AFD}`, result 1; `Rand(3)` → `{0xF09B2846, 0x13ADFF7A}`, result 1; appearance `Rand(1)` → `{0xA3DF5D58, 0x645ADEBE}`, result 0. Trace: `req = min(3, 1, 3) = 1`; variant 0 only. Outputs: `entryCount = 2`; entry 1 for class 100 with variant `1, 1, 0...`; appearance = variant 0; three steps
  - (b) as (a) with `v = 1`. Seeds: `M`. Draws: `Rand(4)` = 1; `Rand(3)` = 1; `Rand(2)` → `{0xA3DF5D58, 0x645ADEBE}`, result 0, so A = slot 0 and the list becomes {1}; `Rand(1)` → `{0xC5CF3376, 0x44599814}`, result 0, so B = slot 1; re-roll A (`c = 4`) → `{0xB30291E2, 0x5281396D}`, 2; re-roll B (`c = 3`) → `{0xEE569C57, 0x4AA9E99F}`, 0; appearance `Rand(2)` → `{0x5C8B2892, 0x6368BA81}`, 0. Trace: `req = 2`; new variant `2, 0...` differs and is kept after one attempt. Outputs: variants `1, 1, 0...` and `2, 0, 0...`; appearance = variant 0; seven steps
  - (c) entry 1 (class 100) has variant count 2 and variants `1, 1, 0...` and `2, 0, 0...`. Seeds: `M`. Draws: appearance `Rand(2)` → `{0x2F2EDAB5, 0x2C890AFD}`, `lo & 1 = 1`. Outputs: no build; `entryCount` stays 2; appearance = `2, 0, 0...`; one step
  - (d) `cnt = 2`, other slots 0; `v = 2`. Seeds: `M`. Draws: `Rand(2)` → `{0x2F2EDAB5, 0x2C890AFD}`, 1; variant 1: step → `{0xF09B2846, 0x13ADFF7A}`, `lo & 1 = 0`; variant 2: steps → `{0xA3DF5D58, 0x645ADEBE}`, `{0xC5CF3376, 0x44599814}`, `{0xB30291E2, 0x5281396D}`, each 0; appearance `Rand(3)` → `{0xEE569C57, 0x4AA9E99F}`, 0. Trace: `req = 3`; `k = 1`, so A = B = slot 0 with no pick draw; variant 2 duplicates variant 1 on all three attempts and is kept. Outputs: slot-0 values `1`, `0`, `0` for variants 0, 1, 2 (other slots 0); variant count 3; appearance = variant 0; six steps
  - (e) `cnt = 2, 2`, other slots 0; `v = 2`. Seeds: `{1, 0x29A}`. Draws: `Rand(2)` → `{0x6AC6935F, 0}`, 1; `Rand(2)` → `{0x2F2ED81B, 0x2C890AFD}`, 1; `Rand(2)` → `{0x280687C4, 0x13ADFE65}`, 0 (A = slot 0); `Rand(1)` → `{0xB6AAB839, 0x10B1BFE6}`, 0 (B = slot 1); variant 1: (`{0xBFC793C3, 0x4C305F9E}`, 1), (`{0x2AB4C4AD, 0x4FFD63FF}`, 1) duplicate, then (`{0x9B880D20, 0x11CFF966}`, 0), (`{0xAFA41306, 0x40DEF89D}`, 0) kept; variant 2: (`{0xE778FC3B, 0x49422C1A}`, 1), (`{0xFADF7581, 0x608BA2E9}`, 1) duplicate, then (`{0x6C5F9F2E, 0x68A31D5B}`, 0), (`{0xC1567BC1, 0x2D33A727}`, 1) kept; appearance `Rand(3)` → `{0x3BA372AC, 0x50A3C577}`, 1. Outputs: variants `1, 1`, `0, 0`, `0, 1`; appearance = `0, 0`; thirteen steps
  - (f) `cnt = 2, 3, 5`, other slots 0; `v = 1`. Seeds: `{2, 0x29A}`. Draws: `Rand(2)` → `{0xD58D2424, 0}`, 0; `Rand(3)` → `{0x95C90FB4, 0x591214E4}`, 2; `Rand(5)` → `{0x62D86A68, 0x3E796EB7}`, 4; `Rand(3)` → `{0x674DD0BF, 0x293A4757}`, 0, so A = slot 0 and the list becomes {2, 1}; `Rand(2)` → `{0xF7415A52, 0x2B165909}`, 0, so B = slot 2; re-roll A (`c = 2`) → `{0x029AFA23, 0x6720D7C5}`, 1; re-roll B (`c = 5`) → `{0x8A2904B4, 0x011630E9}`, 3; appearance `Rand(2)` → `{0x90840F6D, 0x39A025CD}`, 1. Outputs: variants `0, 2, 4` and `1, 2, 3`; appearance = `1, 2, 3`; eight steps
  - (g) `entryCount = 1` (entry 0 for class 50); `v = 1` (not read); MonStats2 row with `TotalPieces = 2`; supplied unregistered-slot variant count 0; slot component counts as MSD-V03. Seeds: `M`. Draws: as MSD-V03. Outputs: `entryCount` stays 1; appearance `1, 1, 0, 0`, rest 0; three steps
  - (h) level record holds entry 0 for class 50 (`entryCount = 1`), with new entry 1's variant count supplied as 0; class 100 has `TotalPieces = 3`, all 16 component counts 0 and `v = 2`. Seeds: `M`. Draws: appearance `Rand(1)` → `{0x2F2EDAB5, 0x2C890AFD}`, result 0; no build draw. Trace: requested 3 variants; initial all-zero variant is stored, then `k = 0` ends the build. Outputs: `entryCount = 2`; entry 1 for class 100 has variant count 1 and an all-zero variant; appearance all zero; one step. Discriminates: continuing to generate variants with no eligible slots would give variant count 3.
  - Discriminates: (g) rejects registering an entry when `TotalPieces <= 2`; (b) rejects slot picks before variant 0 and the old-lo mask; (c) rejects rebuilding for a later monster; (d) rejects discarding a duplicate after three attempts or a fourth attempt; (e) rejects keeping the first duplicate; (f) rejects removing A by shifting, which would make B slot 1
  - runtime_observed: no

- **MSD-V03** Appearance without a preset (MSD-R03; boundary: supplied-input stage)
  - Inputs: class 195 (no preset); slot component counts 4, 3, 0, 1 for slots 0..3, other slots 0
  - Seeds: `M`
  - Draws: (slot 0, c 4) → `{0x2F2EDAB5, 0x2C890AFD}`, `lo & 3 = 1`; (slot 1, c 3) → `{0xF09B2846, 0x13ADFF7A}`, `lo mod 3 = 1`; slot 2 no draw; (slot 3, c 1) → `{0xA3DF5D58, 0x645ADEBE}`, `& 0 = 0`
  - Trace: zero-count slots store 0 without a step; count 1 still steps
  - Outputs: appearance bytes `1, 1, 0, 0`, rest 0; final seed `{0xA3DF5D58, 0x645ADEBE}`; three steps
  - Discriminates: skipping the step for count 1 would leave the seed at `{0xF09B2846, 0x13ADFF7A}`
  - runtime_observed: no

- **MSD-V04** Class draw (MSD-R04; boundary: isolated rule)
  - Inputs: (a) class 311; (b) class 312; (c) class 310
  - Seeds: `M`
  - Draws: (a), (b) `Rand(4)` → `{0x2F2EDAB5, 0x2C890AFD}`, result `lo & 3 = 1`; (c) none
  - Trace: power-of-two mask on the new lo
  - Inputs, shared: appearance bytes from MSD-R03 all 9
  - Outputs: (a) byte 10 = 1, others 9; (b) bytes 10 and 11 = 1, others 9; (c) bytes unchanged, seed unchanged
  - Discriminates: (c) rejects drawing for every class; (b) rejects writing only A for class 312
  - runtime_observed: no

- **MSD-V05** Effective player count (MSD-R05; boundary: supplied-input stage)
  - Inputs, shared: player units of type 0: P1 alive (mode 1, no state 7, condition clear), P2 in mode 17, P3 with state 7 in mode 1, P4 with `aliveExclusionCondition` in mode 1
  - Seeds: `no draw`
  - Draws: none
  - (a) `selectorByte = 0`, override 5. Trace: only P1 counts; override ignored. Outputs: `E = 1`
  - (b) `selectorByte = 2`, override 5. Outputs: `E = 5`
  - (c) `selectorByte = 4`, override 5. Outputs: `E = 1`
  - (d) `selectorByte = 1`, override 0, P1 in mode 0 instead. Outputs: `E = 0`
  - (e) `selectorByte = 1`, override 1, plus P5 alive like P1. Outputs: `E = 2`
  - (f) `selectorByte = 3`, override 5. Outputs: `E = 5`
  - (h) `selectorByte = 0`, plus P10 type 0 in mode 12 and P11 type 1 in mode 17. Outputs: `E = 3` (P1, P10, P11)
  - (g) `selectorByte = 0`, plus P6 type 1 in mode 0, P7 type 1 in mode 12, P8 type 1 in mode 1 and P9 type 2 in mode 1. Outputs: `E = 2` (P1 and P8)
  - Discriminates: (a) rejects counting dead or state-7 players; (c) rejects the override for every nonzero selector; (e) rejects replacing `E` by the override; (g) rejects counting only type 0 or counting other types; (h) rejects testing the other type's death modes
  - runtime_observed: no

- **MSD-V06** Scaling count and percents (MSD-R06; boundary: isolated rule)
  - Inputs: (a) `E = 1`; (b) `E = 2`; (c) `E = 8`; (d) `E = 0`; (e) `E = 8`, `Align = 1`; (f) supplied `E = 10`; (g) `game.difficulty = 3`, `E = 1`; (h) `game.difficulty = 4`, `E = 1`
  - Seeds: `no draw`
  - Draws: none
  - Trace: table for `n < 9`; formulas otherwise
  - Outputs: (a) n 1, life 0, xp 0; (b) n 2, 50, 50; (c) n 8, 350, 350; (d) n 1, 0, 0; (e) n 1, 0, 0; (f) n 10, life 400, xp 360; (g) n 1, 0, 0, `game.difficulty` set to 2; (h) n 1, 0, 0, `game.difficulty` set to 2
  - Discriminates: (f) rejects one formula for both percents; (e) rejects scaling a monster with `Align`; (h) rejects rewriting only the value 3
  - runtime_observed: no

- **MSD-V07** Difficulty index (MSD-R07; boundary: isolated rule)
  - Inputs: (a) `game.difficulty = 1`, class 100; (b) 2, class 271; (c) 0, class 560; (d) 4, class 100; (e) 1, class 338; (f) 2, class 359; (g) 2, class 561
  - Seeds: `no draw`
  - Draws: none
  - Trace: clamp, then `unitGate` override
  - Outputs: (a) `d = 1`; (b) `d = 0`; (c) `d = 0`; (d) `d = 2`; (e) `d = 0`; (f) `d = 0`; (g) `d = 0`
  - Discriminates: (b), (e), (f) and (g) reject ignoring the Normal override or honoring only some gate values; (d) rejects an unclamped index
  - runtime_observed: no

- **MSD-V08** Monster level (MSD-R08; boundary: supplied-input stage)
  - Inputs, shared: `Level, Level(N), Level(H) = 5, 40, 70`; Levels row 20 with `MonLvl2Ex = 55`, `MonLvl3Ex = 82`; Levels row 136 with `MonLvl3Ex = 60`; Levels row 1 with `MonLvl2Ex = 33`; Levels row 0 with `MonLvl3Ex = 77`; LevelsCount 137
  - Seeds: `no draw`
  - Draws: none
  - (a) classic, `d = 2`. Outputs: 70
  - (b) expansion, `d = 2`, room level id 20. Outputs: 82
  - (c) expansion, `d = 1`, level id 20, `boss` set. Outputs: 40
  - (d) expansion, `d = 2`, room null. Outputs: 1
  - (k) expansion, `d = 2`, room level id 0. Outputs: 1
  - (e) expansion, `d = 0`, level id 20. Outputs: 5
  - (f) expansion, `d = 1`, level id 20, `noRatio` set, `boss` clear. Outputs: 40
  - (g) expansion, `d = 2`, level id 137. Outputs: 1
  - (h) expansion, `d = 2`, level id 136. Outputs: 60
  - (i) expansion, `d = 1`, level id 20. Outputs: 55
  - (j) expansion, `d = 1`, level id 1. Outputs: 33
  - Discriminates: (c) and (f) reject the area level for either flag; (e) rejects it on Normal; (g), (h), (j) and (k) reject an off-by-one bound at either end (reading row 0 would give 77); (i) rejects reading `MonLvl3Ex` for Nightmare
  - runtime_observed: no

- **MSD-V09** Level and player-count stats (MSD-R09; boundary: supplied-input stage)
  - Inputs: (a) level 82, `E = 8`, `Align = 0`; (b) level 5, `E = 1`; (c) level 5, `E = 8`, `Align = 1`; (d) level 5, `E = 0`
  - Seeds: `no draw`
  - Draws: none
  - Trace: `n` from MSD-R06; request stat 12 first, then stat 100 = `max(1, n)`
  - Outputs: (a) logged requests (12, 82), (100, 8); (b) (12, 5), (100, 1); (c) (12, 5), (100, 1); (d) (12, 5), (100, 1)
  - Discriminates: the log order rejects writing stat 100 first; (c) and (d) reject writing raw `E`
  - runtime_observed: no

- **MSD-V10** Base values (MSD-R10; boundary: supplied-input stage)
  - Inputs, shared: MonLvl table of 100 rows; rows 0 and 99 have every cell 150 unless a case states otherwise
  - Seeds: `no draw`
  - Draws: none
  - (a) `MinHP[2] = 10`, `MaxHP[2] = 7`, `AC[2] = 20`, `Exp[2] = 33`; `sel = 0`, `d = 2`, level 200; row 99 `HP(H) = 250`, `AC(H) = 150`, `XP(H) = 300`. Trace: `mulDiv(10, 250, 100) = 25`; `1750/100 = 17`; `3000/100 = 30`; `9900/100 = 99`. Outputs: 25, 17, 30, 99
  - (b) words 10, 7, 20, 33 at `d = 0`; `selectorFlag = 1`, level 10; row 10 `L-HP = 150`, `L-AC = 100`, `L-XP = 200`. Trace: `1500/100 = 15`; `1050/100 = 10`; 20; 66. Outputs: 15, 10, 20, 66
  - (c) as (b) with `noRatio`. Outputs: 10, 7, 20, 33
  - (i) `d = 2`, `sel = 0`, level 10; words `MinHP[2] = 10`, `MaxHP[2] = 7`, `AC[2] = 30001`, `Exp[2] = 32767`; row 10 `HP(H) = 100`, `AC(H) = 175`, `XP(H) = 333`. Trace: `5250175/100 = 52501`; `10911411/100 = 109114`. Outputs: 10, 7, 52501, 109114
  - (j) as (i) with `selectorFlag = 1` and row 10 `L-HP(H) = 120`, `L-AC(H) = 130`, `L-XP(H) = 250` (row 10 still has the `(H)` cells of (i)). Trace: `1200/100 = 12`; `840/100 = 8`; `3900130/100 = 39001`; `8191750/100 = 81917`. Outputs: 12, 8, 39001, 81917
  - (k) as (i) with `noRatio`. Outputs: 10, 7, 30001, 32767
  - (l) as (k) (`noRatio` set) with the `MinHP[2]`, `MaxHP[2]`, `AC[2]` and `Exp[2]` words each 0xFFF6. Outputs: -10, -10, -10, -10
  - (d) as (b) with level -1. Trace: `r = -1`, no row read. Outputs: 0, 0, 0, 0
  - (g) as (c) (`noRatio` set) with level -1. Outputs: 0, 0, 0, 0
  - (h) as (b) with `selectorByte = 2` and `selectorFlag = 0`. Outputs: 15, 10, 20, 66
  - (e) as (b) with `selectorByte = 1` and `selectorFlag = 0`. Outputs: 15, 10, 20, 66
  - (f) `MinHP[0] = MaxHP[0] = 30000`, `AC[0] = Exp[0] = 0`; `sel = 0`, level 10; row 10 `HP = 100000`. Trace: `mulDiv(30000, 100000, 100)`: neither threshold applies, so `low32(3000000000) = -1294967296` and `/100 = -12949672`. Outputs: -12949672, -12949672, 0, 0
  - Discriminates: (a) rejects reading row 200; (d) rejects clamping a negative level to row 0 (which gives 15, 10, 30, 49); (g) rejects testing `noRatio` before the row; (b), (e) and (h) reject the base columns when either selector, or any nonzero selector byte, is set; (f) rejects the operand order `mulDiv(cell, word, 100)`, which gives 30000000; (i) and (j) reject a wrong armor-class column and rounding; (k) rejects scaling armor class under `noRatio`; (l) rejects an unsigned read (65526) of any of the four words
  - runtime_observed: no

- **MSD-V11** Life roll and bonus (MSD-R11; boundary: supplied-input stage)
  - Seeds: `M` for each case
  - (a) `minHP 10`, `maxHP 19`, `lifePct 50`. Draws: `Rand(10)` → `{0x2F2EDAB5, 0x2C890AFD}`, `791599797 mod 10 = 7`. Trace: `hp = 17`; `mulDiv(50, 17, 100) = 8`. Outputs: `hp = 25`
  - (b) `minHP = maxHP = 50`, `lifePct 0`. Draws: `Rand(1)` → `{0x2F2EDAB5, 0x2C890AFD}`, result 0. Outputs: `hp = 50`; seed `{0x2F2EDAB5, 0x2C890AFD}`
  - (c) `minHP 50`, `maxHP 49`, `lifePct 0`. Draws: none; seed stays `M`. Outputs: `hp = 50`
  - (d) `minHP = maxHP = 2000050`, `lifePct 50`. Draws: `Rand(1)` → `{0x2F2EDAB5, 0x2C890AFD}`, 0. Trace: `b > 0x100000` and `100 <= b >> 4`, so `(2000050 / 100) * 50 = 1000000`. Outputs: `hp = 3000050`
  - (e) `minHP = maxHP = 1048576`, `lifePct 50`. Draws: `Rand(1)` → `{0x2F2EDAB5, 0x2C890AFD}`, 0. Trace: narrow branch `52428800 / 100 = 524288`. Outputs: `hp = 1572864`
  - (f) `minHP = maxHP = 131072`, `lifePct 50`. Draws: `Rand(1)` → `{0x2F2EDAB5, 0x2C890AFD}`, 0. Trace: `b = 131072` is not above `0x100000` and `a = 50` is not above `0x10000`, so `6553600 / 100 = 65536`. Outputs: `hp = 196608`
  - Discriminates: (d) rejects the exact product 1000025; (c) rejects drawing on an empty range; the stepped seeds of (b), (d), (e) and (f) reject a range of `max - min` without `+ 1`; (f) rejects `mulDiv(hp, lifePct, 100)`, which divides first and gives 196572
  - runtime_observed: no

- **MSD-V12** Life clamp and write (MSD-R12, NRF-R08a; boundary: supplied-input stage)
  - Inputs: (a) `hp = 100`; (b) `hp = 450`; (c) `hp = 0x800000`; (d) `hp = -1`
  - Seeds: `no draw`
  - Draws: none
  - Trace: clamp at `0x800000`, shift left 8, low 32 bits
  - Outputs: (a) logged (7, `0x6400`), (6, `0x6400`); (b) (7, `0x1C200`), (6, `0x1C200`); (c) (7, `0x7FFFFF00`), (6, `0x7FFFFF00`); (d) (7, `0xFFFFFF00`), (6, `0xFFFFFF00`)
  - Discriminates: the log order rejects writing stat 6 first; (c) rejects storing `0x80000000`
  - runtime_observed: no

- **MSD-V13** Armor class and experience (MSD-R13; boundary: supplied-input stage)
  - Inputs: (a) `ac = 30`, `exp = 100`, `xpPct = 350`; (b) `ac = 0`, `exp = 33`, `xpPct = 50`; (c) `ac = 1`, `exp = 100050`, `xpPct = 50`
  - Seeds: `no draw`
  - Draws: none
  - Trace: (a) `mulDiv(350, 100, 100) = 350`; (b) `1650 / 100 = 16`; (c) narrow branch `5002500 / 100 = 50025`
  - Outputs: (a) logged (31, 30), (13, 450); (b) (31, 0), (13, 49); (c) (31, 1), (13, 150075)
  - Discriminates: (b) rejects rounding 16.5 up; (c) rejects `mulDiv(exp, xpPct, 100)`, which divides first and gives 150050
  - runtime_observed: no

- **MSD-V14** Skill-level bonus (MSD-R14; boundary: supplied-input stage)
  - Inputs: `Skill1 = 10`, `Sk1lvl = 3`; `Skill2 = -1`, `Sk2lvl = 5`; `Skill3 = 20`, `Sk3lvl = 0`; `Skill4 = 30`, `Sk4lvl = 1`; `Skill5 = 40`, `Sk5lvl = -2`; `Skill6 = 0`, `Sk6lvl = 2`; `Skill7 = -2`, `Sk7lvl = 4`; others -1; `game.difficulty = 2`; supplied `MonsterSkillBonus` Normal 0, Nightmare 2, Hell 3; `unitGate = 1`
  - Seeds: `no draw`
  - Draws: none
  - Trace: skip negative skill and nonpositive level; the bonus row (Hell) ignores the Normal override
  - Outputs: logged grants (10, 6), (30, 4), (0, 5), in that order
  - Discriminates: rejects the Normal row (grants at 3 and 1) for `unitGate` classes, granting level-0 or negative-level skills or a skill id of -2, and skipping skill id 0
  - runtime_observed: no

- **MSD-V15** Property chances (MSD-R15; boundary: supplied-input stage)
  - Seeds: `M` for each case
  - (a) `d = 0`; base block slots 1..6: (prop 5, chance 0), (prop 6, chance 50), (prop 7, chance 100), (prop -1, chance 0), (prop 8, chance 0), (prop 9, chance 0); `(N)` and `(H)` blocks all (prop 1, chance 0). Draws: slot 2 → `{0x2F2EDAB5, 0x2C890AFD}`, `lo mod 100 = 97`; slot 3 → `{0xF09B2846, 0x13ADFF7A}`, 30. Trace: slot 1 no draw; `97 < 50` false; `30 < 100` true; slot 4 stops the loop. Outputs: logged applications of slots 1 and 3; two steps
  - (b) `d = 0`; base slots: (prop 5, chance 97), then (prop -1) for slots 2..6. Draws: slot 1 → `{0x2F2EDAB5, 0x2C890AFD}`, 97. Trace: `97 < 97` false. Outputs: no application; one step
  - (d) `d = 0`; base slots: (prop 5, chance 50), (prop 6, chance 200), then (prop -1). Seeds: `{0xA3DF5D58, 0x645ADEBE}`. Draws: slot 1 → `{0xC5CF3376, 0x44599814}`, unsigned `lo mod 100 = 78`; slot 2 → `{0xB30291E2, 0x5281396D}`, 82. Trace: `78 < 50` false; `82 < 200` true. Outputs: logged application of slot 2 only; two steps
  - (c) `d = 2`; base block slots all (prop 3, chance 0); `(H)` block: (prop 4, chance 50), then (prop -1). Draws: `(H)` slot 1 → `{0x2F2EDAB5, 0x2C890AFD}`, 97. Outputs: no application; one step
  - Discriminates: (a) rejects drawing for chance 0 or skipping the draw for chance 100, and continuing past a negative property; (b) rejects `roll <= chance`; (a) slots 5 and 6 reject continuing after a negative property; (c) rejects reading the base block when `d = 2`, which would apply six properties with no draw; (d) rejects a signed remainder (-18, applied) and a signed chance byte
  - runtime_observed: no

- **MSD-V16** Equipment roll (MSD-R16; boundary: supplied-input stage)
  - Inputs: monster level 20; `monster.inventory` present; all locations free; rows in table order: R1 (class, level 30, `oninit` 1, locs 0, 0, 0); R2 (class, level 15, `oninit` 0, items `itemA2`, `itemB2`, `itemC2`, locs 4, 5, 0, mods 2, 9, 0); R3 (class, level 5, `oninit` 0, `item1` `itemA3`, locs 3, 0, 0, `mod1` 1); R4 (class, level 40, `oninit` 0, locs 0, 4, 4); R5 (class, level 40, `oninit` 0, `item1` `itemA5`, locs 6, 0, 0, `mod1` -1); R6 (another class)
  - Seeds: `M`
  - Draws: R2 `k = 2` → `{0x2F2EDAB5, 0x2C890AFD}`, `lo & 1 = 1`; R3 `k = 1` → `{0xF09B2846, 0x13ADFF7A}`, 0; R4 `k = 0`, none; R5 `k = 1` → `{0xA3DF5D58, 0x645ADEBE}`, 0
  - Trace: R1 skipped (30 > 20); later rows are not threshold-tested and their `oninit` is not read; R6 ends the walk
  - Outputs: logged item requests (`itemB2`, 20, mod 0 because 9 is out of range), (`itemA3`, 20, 1), (`itemA5`, 20, 0 because -1 is out of range); three steps
  - (b) monster level 10; `monster.inventory` present; location 4 occupied, others free; rows: S1 (class, level 10, `oninit` 1, items `itemA`, `itemB`, `itemC`, locs 1, 2, 3, mods 8, 0, 0); S2 (class, level -5, locs 4, 5, 0); S3 (class, level 50, `item1` `itemD`, locs 7, 0, 0, `mod1` 0); S4 (another class). Seeds: `M`. Draws: S1 `k = 3` → `{0x2F2EDAB5, 0x2C890AFD}`, `lo mod 3 = 0`; S2 `k = 2` → `{0xF09B2846, 0x13ADFF7A}`, `lo & 1 = 0`; S3 `k = 1` → `{0xA3DF5D58, 0x645ADEBE}`, 0. Trace: S1 is not skipped (10 is not above 10); S2 selects occupied location 4, so no item, but its draw is taken. Outputs: logged (`itemA`, 10, 0 because 8 is out of range), (`itemD`, 10, 0); three steps
  - Discriminates: R5 rejects threshold-testing every row; R3 rejects skipping the draw when `k == 1`; R2's mod rejects passing 9; (b) S1 rejects skipping a row whose level equals the monster level and accepting mod 8; S2 rejects skipping the draw for an occupied location
  - (c) monster level 10; `monster.inventory` present; all locations free; rows: T1 (class, level -5, `oninit` 1, items `itemX1`, `itemX2`, `itemX3`, locs 1, 2, 10, mods 0, 7, 0); T2 (class, level 0, locs 11, 1, 1); T3 (another class). Seeds: `{0x2F2EDAB5, 0x2C890AFD}`. Draws: T1 `k = 3` → `{0xF09B2846, 0x13ADFF7A}`, unsigned `lo mod 3 = 1`; T2 `k = 0` (location 11 stops the count), none. Outputs: logged (`itemX2`, 10, 7); one step
  - Discriminates (c): a negative first-row level is not skipped; location 10 counts and 11 stops the count; mod 7 is kept; a signed remainder (0) would select `itemX1`
  - (d) monster level 20; inventory present with all locations free; first row (class, level 0, `oninit` 0, `item1` `itemStopped`, locs 1, 0, 0, `mod1` 0), then a row of another class. Seeds: `M`. Draws: none. Trace: the first row's `oninit` stops the operation before processing its valid location. Outputs: no item request; seed unchanged. Discriminates: ignoring the first-row guard would draw once and request (`itemStopped`, 20, 0).
  - runtime_observed: no

- **MSD-V17** Damage player-count term (MSD-R20; boundary: supplied-input stage)
  - Inputs, shared: base `minDamage 10`, `maxDamage 20`, `toHit 100`; `damageTermSkip` false
  - Seeds: `no draw`
  - Draws: none
  - (a) expansion, difficulty 2, stat 100 = 4. Trace: `term = D[4] = 24`; `240/128 = 1`; `480/128 = 3`; `2400/128 = 18`. Outputs: term 24; logged (21, 11), (22, 23), (19, 118)
  - (b) as (a), classic, with 10, 20, 100 the values after the classic adjustment. Outputs: term 24; logged 10, 20, 100
  - (c) expansion, difficulty 0, stat 100 = 8. Outputs: term 0; logged 10, 20, 100
  - (d) expansion, difficulty 1, stat 100 = 10 (supplied). Trace: `term = 8*10 - 16 = 64`. Outputs: logged 15, 30, 150
  - (e) as (a) with `damageTermSkip` true. Outputs: term 0; logged 10, 20, 100
  - (f) as (a) with `minDamage -10`. Trace: `-240/128` truncates to -1. Outputs: logged -11, 23, 118
  - (g) expansion, difficulty 1, stat 100 = 8. Trace: `term = D[8] = 56`; `560/128 = 4`; `1120/128 = 8`; `5600/128 = 43`. Outputs: logged 14, 28, 143
  - (h) expansion, difficulty 1, stat 100 = 2. Trace: `term = D[2] = 8`; `80/128 = 0`; `160/128 = 1`; `800/128 = 6`. Outputs: term 8; logged (21, 10), (22, 21), (19, 106), in that order. Discriminates: treating two players as a zero-term case would leave 10, 20, 100.
  - Discriminates: (a) rejects `D[p - 1]` (term 16 gives 11, 22, 112); (b) rejects applying the term in classic games; (c) rejects a Normal term; (b), (c), (e) reject omitting the writes when the term is not applied; (f) rejects flooring (-12); (g) rejects the formula `8p - 16` (48) at `p = 8`. Every logged triple is stats 21, 22, 19 in that order
  - runtime_observed: no

- **MSD-V18** Eligible count (MSD-R21; boundary: supplied-input stage)
  - Inputs, shared: `E` is supplied as an MSD-R05 outcome and the party size as the PartyCount outcome of the listed members; a living player killer with a room, counted among its own party's members; party members listed are living, in the game and on the killer's level unless stated; a monster dropper with stat 100 = 8 unless stated; `N != 0`; `Picks >= 0`
  - Seeds: `no draw`
  - Draws: none
  - Trace: `eligible = elig + (E - elig) / 2`, then the dropper cap
  - Outputs: (party 1, `E` 1) 1; (1, 4) 2; (1, 8) 4; (2, 8) 5; (8, 8) 8; (5, 3) 4; (three living members, one on another level, `E` 3) party 2, 2; (party 4, `E` 8, stat 100 = 2) 2; (monster killer without owner, `E` 8) 4; (party 4, `E` 1) 3; (party of three: killer, one dead member and one member with a null room, `E` 8) party 1, 4; (summon killer whose owner is a player with party 2, `E` 8) 5; (party 1, `E` 8, dropper of type 2) 4; (party 1, `E` 8, null dropper) 4; (party 1, `E` 8, monster dropper with stat 100 = 0) 1; (party of two: killer and a member with state 7 in mode 1, `E` 8) party 2, 5
  - Additional case: ten living same-level party members including the killer, supplied `E = 8`, dropper of type 2. Seeds: `no draw`. Draws: none. Trace: `party = 10`, `elig = min(10, 8) = 8`, `eligible = 8 + (8 - 8) / 2 = 8`; no monster cap. Outputs: `elig = 8`, `eligible = 8`. Discriminates: omitting the party cap gives `10 + (8 - 10) / 2 = 9`, and a monster dropper's spawn cap cannot mask that error here.
  - Discriminates: (5, 3) rejects clamping `elig` to `E`; the other-level case rejects a party count without the level test; (party 4, `E` 1) rejects flooring (2); the dead or roomless members reject counting them; the summon case rejects `elig = 1` for owned summons; the type 2 and null droppers reject capping non-monsters; stat 100 = 0 rejects skipping the lower bound of 1; the state-7 member rejects reusing MSD-R05's exclusion (party 1, 4)
  - runtime_observed: no

- **MSD-V19** NoDrop conversion (MSD-R22a; boundary: isolated rule)
  - Seeds: `no draw`
  - Draws: none
  - (a) `N = 100`, `T = 100` (`r = 0.5` exactly). Trace: `p = 2^-e`; `p` and `q = 1 - p` are exact binary values. The quotient need not be exact; the listed truncations are stable in single and double precision. Outputs: eligible 1 → `N' = 100` (no conversion); 2 → 33; 3 → 14; 4 → 6; 5 → 3; 6 → 1; 7 → 0; 8 → 0
  - (b) `N = 60`, `T = 40` (`r = 0.6`). Trace: eligible 2: `p = 0.36`, `q = 0.64`, `0.5625 * 40 = 22.5`; eligible 3: `0.216/0.784 * 40 = 11.02`; eligible 4: `0.1296/0.8704 * 40 = 5.96`. Outputs: 22, 11, 5
  - (c) `N = 0`, `T = 100`, eligible 4. Outputs: `N' = 0`, no conversion
  - (d) `N = 60`, `T = 40`, eligible 1. Outputs: `N' = 60`, no conversion (converting in double precision would give 59)
  - Discriminates: (a) eligible 2 rejects `trunc(T * p)` (25); (b) rejects rounding up (23, 12, 6) and using `T / (N + T)` as `r` (eligible 2 gives 7); these values truncate the same way in single and double precision
  - runtime_observed: no

- **MSD-V20** Pick draw (MSD-R23; boundary: supplied-input stage)
  - Inputs, shared: `T = 100`, `Picks = 1`, `N = 100` before conversion; killer seed `{5, 0x29A}`; game seed `{7, 0x29A}`
  - Seeds: `dropper.seed = M` for each case; killer and game seeds unchanged in every case
  - (a) `N' = 33` (MSD-R22a with eligible 2). Draws: `Rand(133)` → `{0x2F2EDAB5, 0x2C890AFD}`, 23. Outputs: NoDrop; pick consumed
  - (b) `N' = 100` (eligible 1). Draws: `Rand(200)` → 197. Trace: `197 - 100 = 97`. Outputs: item selected by roll 97
  - (c) `N' = 33`, `noDropBypass`. Draws: `Rand(100)` → 97. Outputs: item selected by roll 97; no NoDrop test
  - (d) `Picks = -2`. Draws: none. Outputs: no NoDrop; seed stays `M`
  - (e) `T = 110`, supplied `N' = 23`. Draws: `Rand(133)` → 23. Trace: `23 < 23` false. Outputs: item selected by roll 0
  - (f) `T = 0`, `N = 100`, `Picks = 1`, first without and then with `noDropBypass`. Draws: none. Trace: the treasure-class level exits before eligibility or pick selection. Outputs: no NoDrop outcome or item-selection roll; `dropper.seed` stays `M` in both cases
  - Discriminates: (a) rejects a separate treasure seed and `roll >= N'`; (c) rejects including `N'` in the bypass range; (e) rejects `roll <= N'`; (f) rejects drawing when `T = 0`, even under bypass; the unchanged killer and game seeds reject drawing on another seed
  - runtime_observed: no

- **MSD-V21** Creation order across rules (MSD-R02, MSD-R03, MSD-R04, MSD-R06, MSD-R09, MSD-R11 to MSD-R16; boundary: supplied-input stage)
  - Inputs: room level id 5; MonLvl table of 100 rows (row 10 available); DifficultyLevels 3 rows (Normal `MonsterSkillBonus` 0); `MonProp` index 4 within the MonProp row count; class 311; its level record holds entry 0 for class 311 with variant count 2 and variants all 3 and all 5 (16 bytes each); `E = 2`; `d = 0`; `MonStats.Level = 10`; `noRatio` set; `MinHP 10`, `MaxHP 19`, `AC 5`, `Exp 40`; `Skill1..8 = -1`; MonProp base slot 1 (prop 2, chance 80), slot 2 prop -1; MonEquip first row (class, level 0, `oninit` 1, `item1` `itemE`, locs 1, 0, 0, `mod1` 0), then a row of another class; `monster.inventory` present, location 1 free
  - Seeds: `M`
  - Draws: appearance `Rand(2)` → `{0x2F2EDAB5, 0x2C890AFD}`, 1; class `Rand(4)` → `{0xF09B2846, 0x13ADFF7A}`, 2; life `Rand(10)` → `{0xA3DF5D58, 0x645ADEBE}`, `2749324632 mod 10 = 2`; property step → `{0xC5CF3376, 0x44599814}`, `lo mod 100 = 78`; equipment step → `{0xB30291E2, 0x5281396D}`, `& 0 = 0`
  - Trace: appearance = variant 1 (all 5), then byte 10 = 2; `n = 2`, `lifePct = xpPct = 50`; base values copied (`noRatio`); `hp = 12 + mulDiv(50, 12, 100) = 18`, raw `0x1200`; `xp = 40 + 20 = 60`; `78 < 80` applies the property
  - Outputs: appearance bytes all 5 except byte 10 = 2; logged writes (12, 10), (100, 2), (7, `0x1200`), (6, `0x1200`), (31, 5), (13, 60); logged `ApplyProperty(base block, slot 1)`; logged `CreateEquipmentItem(itemE, 10, 0)`; final seed `{0xB30291E2, 0x5281396D}`
  - Discriminates: drawing the class value before the appearance, the life roll after the property chance, or the equipment roll before the properties changes the life (any reorder gives a different roll and final assertions); the write log rejects any reorder of the creation writes
  - runtime_observed: no

### 8.2 Observation vectors

All are pending gameplay observations ([procedures](../research/spec-evidence/monster-scaling-by-players-and-difficulty.md#observation-procedures)).

- **MSD-O01** Life of one fixed monster type in one area at counts 1, 2 and 8 in each difficulty; expect MSD-R10 to MSD-R12 values from the supplied bounds and `lifePct` 0, 50, 350.
- **MSD-O02** Experience value of that monster at counts 1 and 8; expect +350% at 8.
- **MSD-O03** NoDrop rate of a fixed treasure class at counts 1, 2 and 8 with party
  members near, far on the same level, on another level and dead; expect no distance effect
  and MSD-V19's weights. For a discriminating dead-player case, use a real fixed treasure
  class with `N > 0` and `T > 0`, override 0, a monster whose stat 100 is at least 3, and
  three connected players: killer A, living unpartied B and C who dies but remains in the
  game. Repeat enough kills and picks to compare the NoDrop rates before and after C dies.
  The 1.13c static path predicts `E = 2`, `elig = 1`, `eligible = 1` after death; counting all
  connected slots would give `E = 3`, `eligible = 2`. Compute both rates from that class's
  actual `N` and `T`. As a supplied arithmetic illustration only, `N = T = 100` gives
  NoDrop weights 100 (50% per pick) and 33 (about 24.8%) respectively. A killer plus one
  dead player cannot distinguish these counts because signed halving leaves `eligible = 1`
  either way.
- **MSD-O04** A kill by a summoned minion: expect the owner's party to set `elig`.
- **MSD-O05** Stat 100, life and experience of a monster created before a player joins or leaves; record whether they change. MSD-Q03 leaves the result open.
- **MSD-O06** Party larger than `E`, if reachable; expect the MSD-R21 negative term.

## 9. Acceptance and dependencies

The behavior is reproduced when MSD-V01 to MSD-V21 pass and the observation vectors match.
Partial acceptance covers the verified assertions of MSD-V01 to MSD-V21 with supplied inputs:
draw counts, draw order, seed states, computed values and requested writes in order. Gated
assertions: the stored effect of the requested writes (RLI-Q03); results of MSD-R22a that depend
on intermediate rounding (MSD-Q02, MSD-R22b); real MonStats2 prepared preset inputs (MSD-Q06);
real damage-setup inputs (MSD-Q05); and the asm-only creation draws of MSD-R17 to MSD-R19
(MSD-Q04). Supplied MonEquip rows followed by another class are eligible; real tables whose
class rows reach the end require MSD-Q10's storage guarantee. A supplied unregistered preset
slot with variant count 0 is eligible; its actual initialized variant count requires MSD-Q11.
A complete replay of a monster's creation draws needs MSD-Q04 resolved. Full
reproduction also needs the observations of MSD-Q07.

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md):
NRF-R01/NRF-R02 step and draw forms, NRF-R03a and NRF-R12a/NRF-R12b creation seeding,
NRF-R04b mode-slot draws, NRF-R05 `mulDiv`, NRF-R06a `/128`, NRF-R08a 8.8 life format;
[core-record-layouts-and-ids](core-record-layouts-and-ids.md): RLI-R03 `HasState`,
RLI-R02a stat reads; [combat-status-application](combat-status-application.md):
CSA-R02 `unitGate`;
[core-game-data-table-records](core-game-data-table-records.md): GDA-R05 DifficultyLevels,
GDA-R08a MonStats, GDA-R15a/GDA-R15b treasure fields;
[core-game-data-fields](core-game-data-fields.md): GDF-R02 stats, GDF-R03 states,
GDF-R05a and GDF-R06a modes, GDF-R08 class identities;
[player-party-experience-sharing](player-party-experience-sharing.md): PXL-R04a
enumeration of the killer's party members with matching room level id. Both contracts state
that a missing party record yields no callback (companion).

**Depended on by:** [combat-crushing-blow](combat-crushing-blow.md) (stat 100, MSD-R09);
[core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) (supplied life bounds and
bonus). A reverse consumer is not a delivery prerequisite.

## 10. Related specs

- [player-party-experience-sharing](player-party-experience-sharing.md): the kill experience award that consumes stat 13, and the same-level party iterator also used by MSD-R21 (PXL-R04a).
- [combat-crushing-blow](combat-crushing-blow.md): reads stat 100 written by MSD-R09.
- [combat-hit-recovery](combat-hit-recovery.md): monster mode-change draws in the damage setup (MSD-R19).
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): where monster damage values are consumed.
- [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md): shared arithmetic, seeds and the life format.
- [core-record-layouts-and-ids](core-record-layouts-and-ids.md): stat reads and writes and state tests.
- [combat-status-application](combat-status-application.md): the `unitGate` helper under CSA-R02.
- [core-game-data-table-records](core-game-data-table-records.md): prepared table fields and lookups.
- [core-game-data-fields](core-game-data-fields.md): stat, state, mode and class identities.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. The original `D2Game.dll` and
`D2Common.dll` code cited by the companion is identified through the
[source manifest](../research/source-manifest.md). The client producer of the player-count
override is named only as an importer of a cited `D2Game.dll` export.
Static readings date from 2026-10-12 (disassembly) and 2026-10-14 (independent instruction
listings of the original images and loader column descriptors). The life roll and life write
evidence is shared with the foundations companion. No gameplay observation was made and no
archive table cell was read; all vectors use supplied values. The
[public evidence companion](../research/spec-evidence/monster-scaling-by-players-and-difficulty.md)
gives the function references, rationale, loader column records and observation procedures.
Function descriptions are analyst labels.
