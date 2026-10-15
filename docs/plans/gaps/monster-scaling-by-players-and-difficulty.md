# Gaps: monster-scaling-by-players-and-difficulty

Spec: [Monster scaling by player count and difficulty](../../specs/monster-scaling-by-players-and-difficulty.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|-----|------|---------|-------------------------------|------------------|-----------------------------------|----------------|
| MSD-Q01 | outside-boundary | MSD-R05, MSD-R10, MSD-R21 | none locally; the rules use the tested values of `game.selectorByte`, `game.selectorFlag` and `notAliveCondition` | Producer meanings of the game-creation byte and flag bit 21, and of the unit condition | 2026-10-14: game creation `0x2C2D0` read for the writes (creation argument byte; flag bit 21; expansion from bit 20); callers not read ([companion](../../research/spec-evidence/monster-scaling-by-players-and-difficulty.md#search-records)) | Read the game-creation callers and the writers of the unit condition |
| MSD-Q02 | behavior | MSD-R22a, MSD-R22b | NoDrop weights whose truncated value depends on intermediate rounding; exact-binary cases such as MSD-V19 remain available | The floating-point precision control in effect on the original server, and results for inputs outside the covered domain (zero divisor, out-of-range truncation) | 2026-10-14: listing of `0x12534..0x125FB` and `0xF5724`; no precision-control write on the path | Read every precision-control write in the hosting process and its server-thread start |
| MSD-Q03 | behavior | MSD-R09, MSD-R12, MSD-R13, MSD-O05 | Claims that values stay fixed after players join or leave; creation values are unaffected | Exhaustive writer search for stats 100, 6, 7 and 13 on monsters | 2026-10-12: no recompute path found; 2026-10-14: stat-100 replacing-write search found only creation, undecoded regions and other write operations not covered | Complete the writer search over all stat-write operations and decoded regions |
| MSD-Q04 | integration | MSD-R17, MSD-R18, MSD-R19 | A complete replay of a monster's creation draws, including item seeds, indirect calls and mode-update draws; MSD-R01 to MSD-R16 draws with supplied inputs remain available | Item-creation path conditions, two indirect call targets, mode-update eligibility (NRF-Q11) | 2026-10-12: chains read in disassembly; 2026-10-14: item-creation call sites located, stat-list callback inspected | Trace `0xAFCE0`/`0x11980`/`0x11490`, resolve the indirect calls of `0x3AF00` and `0x3A7D0`, and settle NRF-Q11 |
| MSD-Q05 | outside-boundary | MSD-R20 | none locally for supplied base values; real damage values need the damage-setup owner | Base damage and attack-rating production, `damageTermSkip` and the stat write used | 2026-10-14: listing of `0x772BA..0x774D2`; skip test `0x391F0` and write forwarder `0x3A7B0` identified, not read through | Specify the monster damage setup under its owner |
| MSD-Q06 | outside-boundary | MSD-R02, MSD-R03 | none locally for supplied `cnt` and `v`; real values need the preparation contract | Preparation of the MonStats2 per-slot component counts and variant-limit byte | 2026-10-14: MonStats2 loader descriptors decoded; `TotalPieces` bound; the slot bytes not bound to a text column | Read the MonStats2 component-column preparation |
| MSD-Q07 | observation | MSD-O01 to MSD-O06 | Runtime-confirmation claims only | Gameplay observations on an owned 1.13c game | 2026-10-14: migration made no gameplay observation; none earlier recorded | Run the companion's observation procedures |
| MSD-Q08 | outside-boundary | MSD-R23 | none locally; the bypass is a supplied condition | Which drop callers pass the bypass and why | 2026-10-14: bypass behavior read at `0x125FF..0x12657`; callers not surveyed | Survey the callers of `0x12380` |
| MSD-Q09 | outside-boundary | none | none locally | Whether gold amount or magic find depends on player count | 2026-10-12: not found in the scaling and drop functions read; no full search | Search the gold-drop and quality paths under the treasure owner |
| MSD-Q10 | outside-boundary | MSD-R16 | none locally for supplied rows followed by another class; tables whose class rows reach the end of storage | Whether the prepared MonEquip storage guarantees a following row of another class | 2026-10-14: listing of `0x92F28..0x9302E` shows no row-count bound; loader storage end not read | Read the MonEquip loader's allocation and the table compiler's trailing storage |
| MSD-Q11 | behavior | MSD-R02, MSD-R03, MSD-V02 | The appearance path of a class whose MonStats2 row has `TotalPieces <= 2`, with real storage; supplied variant count 0 remains available | Initialization of a level record's preset entry storage | 2026-10-12: preset lookup read, storage initialization not read; 2026-10-14: blind pass 2 identified the dependence | Read the allocation and initialization of the level preset records |

## Implementation gaps

Comparison of the original behavior with the current Dark Magic code, kept as an input for a
future implementation plan; not a plan. Code named: lua/d2legacy/policy/player_count.lua,
commands/spawn_monster.lua, commands/monster_death.lua, loot/treasure_class.lua,
policy/experience_gain.lua, data/difficulty_levels.lua, data/monster.lua. Test:
internal/mod/d2legacy/party_nodrop_test.go. Values in that code were recovered from 1.10f D2MOO.
The original issue is #373 (parent #344).

| Rule | Repo item | Current repo behavior | Original behavior | Status |
|---|---|---|---|---|
| MSD-R05 | policy/player_count.lua | `/players` override (1..8) else live count; checkpointed under `d2legacy.player_count`. Override is a debug rule, not an original-game fact | live count of players | verified (static) |
| MSD-R06, MSD-R11, MSD-R12 | commands/spawn_monster.lua, data/monster.lua | `bonus = 50*(effective-1)` percent (0..350) applied at spawn; whole life capped at 2^23-1, stored 8.8 | same | verified (matches) |
| MSD-R06, MSD-R13 | policy/experience_gain.lua | same bonus percent applied to stored experience at spawn | same, applied to stat 13 | verified (matches) |
| MSD-R21 | loot/treasure_class.lua, commands/monster_death.lua | `party = nearby + 1`; `elig = min(party, effective)`; `eligible = floor((effective-elig)/2)+elig`, capped by spawn count | no clamp of elig to E (only min(party,8)); no distance ("nearby" is same-level, same-party, alive); CONTRADICTION: party 5, E 3 gives 4 not 3 | verified (static); "level" asm-only |
| MSD-R22a | loot/treasure_class.lua | converts chance back to weight | `N' = trunc(T*p/(1-p))`, `p=(N/(N+T))^eligible` | verified; x87 asm-only |
| MSD-R14 (DifficultyLevels) | data/difficulty_levels.lua | columns ResistPenalty, MonsterFreezeDivisor, MonsterColdDivisor, DeathExpPenalty, LifeStealDivisor, ManaStealDivisor | all real columns; no life/XP/NoDrop column | verified |
| tests | internal/mod/d2legacy/party_nodrop_test.go | two players in one party; checks count chain and credited player; no 3+ players, minion kills, or distance | n/a | n/a |
| Section 7 | hireling Normal-level handling | not applied yet; out of scope | n/a | n/a |
| Section 7 | docs/research/MULTIPLAYER_PARTIES_PVP_AND_TRADE.md | proximity rules shared with party XP, quests, gold | NoDrop "nearby" is same-level and alive | verified |
| Done list | issue #373 | implementation should use live player and party context for life, XP, NoDrop, with tests changing count/difficulty; runtime vectors for counts 1,2,8 and 1.14d spot-check still open | n/a | n/a |
