# Game Data Fields — original-game evidence

Behavioral contract: [Game Data Fields](../../specs/core-game-data-fields.md).
Open gaps: [evidence gaps](../../plans/gaps/core-game-data-fields.md).
Artifact identities: [original game source manifest](../source-manifest.md).

The linked spec is the current contract and governs rule labels, vectors and gates. The 2026-10-14 provenance publication preserved historical findings without adding behavior verification or runtime confirmation. The focused 2026-10-15 refresh reuses compatible readings, clarifies consumed contracts and records narrow text-cell corroboration; it adds no gameplay confirmation. Complete investigation history and original binary-layout fixtures were preserved privately before replacement; they are not needed to follow these public references.

## Source identity and method

Historical static reads of original Lord of Destruction Expansion 1.13c occurred on 2026-10-12/13, with targeted rereads on 2026-10-14. Analysts read reconstructions checked against original machine code where available, and disassembly otherwise. Loader descriptor arrays were decoded from original D2Common instructions/stack stores on 2026-10-13, resolving column strings in read-only data; two machine-code-matched loader reconstructions agreed. Selected undecoded instructions were corroborated against independent original-image listings on 2026-10-14, including the conversion engine. This did not establish every indirect conversion handler.

The manifest pins available original binaries and their correspondence with analysis images. Newly recorded fingerprints do not prove the DLL identity of the historical live process. The historical session recorded the manifest's Game.exe hash and reported the vanilla 1.13c profile. Full live DLL fingerprints, effective table-file hashes, physical supplying archives and a complete override inventory were not captured. No known overrides were reported; that is not an inventory. An executable hash alone cannot establish unmodified game data.

All function descriptions below are analyst labels, not original exported-name claims. Ordinals identify exports in the pinned image. RVAs are relative to the module base, not file offsets. Import thunks and resolved targets are distinct entries. Historical table/memory correspondence is separate from gameplay execution.

## Function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method/boundary |
|---|---|---|---|---|---|
| `LOD113C-D2COMMON` | `0x66160` | 10943 | All-table initialization root | GDF-R01a, GDF-R19b | Dispatch read; reaches loaders and intermediate initializers |
| `LOD113C-D2COMMON` | `0x5D570` | — | Monster-table initializer | GDF-R05a, GDF-Q01 | Reaches monster loader and animation writer |
| `LOD113C-D2COMMON` | `0x5A730` | — | Treasure initialization | GDF-R10a, GDF-Q10 | Generated classes precede named treasure loading |
| `LOD113C-D2COMMON` | `0x603F0` | — | Difficulty table loader | GDF-R01a, GDF-R02 | Reconstruction checked against original machine code and original descriptor decode; publishes pointer/count, requires count 3 |
| `LOD113C-D2COMMON` | `0x60920` | — | Experience table loader | GDF-R01a, GDF-R04 | Reconstruction checked against original machine code and descriptor decode; no count publication requested |
| `LOD113C-D2COMMON` | `0x60A80` | — | Class table loader | GDF-R03a, GDF-Q13 | Original descriptor decode and disassembly; localized-name output separate |
| `LOD113C-D2COMMON` | `0x5AA40` | — | Monster table loader | GDF-R05a, GDF-R18 | Original descriptor decode and disassembly |
| `LOD113C-D2COMMON` | `0x56C10` | — | Extended monster table loader | GDF-R06a, GDF-Q06 | Descriptor decode; indirect callbacks incomplete |
| `LOD113C-D2COMMON` | `0x7C8D0` | — | Missile table loader | GDF-R07, GDF-R18 | Original descriptor decode |
| `LOD113C-D2COMMON` | `0x61EC0` | — | Skill-description and skill loaders | GDF-R08a, GDF-R18 | Descriptor decode; skilldesc precedes skills |
| `LOD113C-D2COMMON` | `0x73330` | — | Stat-definition loader | GDF-R09, GDF-R11a | Original descriptor decode |
| `LOD113C-D2COMMON` | `0x5F6E0` | — | State-definition loader | GDF-R12, GDF-R13 | Original descriptor decode |
| `LOD113C-D2COMMON` | `0x19330` | — | Monster-mode loader | GDF-R15b | Descriptor decode and publication stores |
| `LOD113C-D2COMMON` | `0x19650` | — | Player-type and player-mode loader | GDF-R14b | Descriptor decode; combined allocation and subset |
| `LOD113C-D2COMMON` | `0x72510` | — | Item-type loader | GDF-R16b | Named output, stride and stores; root caller not established historically |
| `LOD113C-D2COMMON` | `0x5F310` | — | Pet-type loader | GDF-R19b | Signed count clamp and >=256 abort; root caller read |
| `LOD113C-D2COMMON` | `0x5EF40` | 10849 | Table load and preparation coordinator | GDF-R01a, GDF-R20a, GDF-Q06, GDF-Q12 | Independent generation/final-read flags; indirect integration gated |
| `LOD113C-D2COMMON` | `0x5D770` | 10903 | Compiled table writer | GDF-R20a | Reconstruction checked against original machine code; writes rows, does not parse text |
| `LOD113C-D2COMMON` | `0x9900` | — | Resource allocation/read helper | GDF-R20a | Resource path read; complete lifecycle not promoted |
| `LOD113C-D2COMMON` | `0x215D` | — | Compiled-output stream wrapper | GDF-R20a | Disassembly; output direction corroborated by callee |
| `LOD113C-D2COMMON` | `0x2056` | — | Stream-write helper | GDF-R20a | Disassembly of output direction; complete writer integration gated |
| `LOD113C-FOG` | `0x12770` | 10208 | Tab-delimited text parser | GDF-R01a, GDF-R01d, GDF-Q12 | Disassembly; separator/count path |
| `LOD113C-FOG` | `0x11F80` | 10210 | Parsed-row count getter | GDF-R01a | Reconstruction checked against original machine code |
| `LOD113C-FOG` | `0x12A60` | 10207 | Descriptor conversion engine | GDF-R01b, GDF-Q06, GDF-Q12 | Disassembly and original-image listing; direct present-cell cases only |
| `LOD113C-FOG` | `0x12290` | — | Header index construction | GDF-R01d, GDF-Q12 | Hash/full comparison; missing/duplicate behavior gated |
| `LOD113C-FOG` | `0x12510` | — | Transformed-name insertion | GDF-Q06, GDF-Q12 | Search attempt; complete name-link contract unresolved |
| `LOD113C-FOG` | `0x121A0` | 10217 | Transformed-name lookup | GDF-Q06, GDF-Q12 | Search attempt; full defaults/conversions unresolved |
| `LOD113C-STORM` | `0x3AC10` | 502 | Header hash | GDF-R01d | Flag 0 selects ASCII case normalization |
| `LOD113C-STORM` | `0x3B1E0` | 509 | Case-insensitive comparison wrapper | GDF-R01d | Follows comparison implementation |
| `LOD113C-STORM` | `0x54BC` | — | Comparison branch selector | GDF-R01d | Reconstruction checked against original machine code; ASCII branch only |
| `LOD113C-STORM` | `0x7660` | — | ASCII case-insensitive comparator | GDF-R01d | Disassembly; locale alternatives not promoted |
| `LOD113C-STORM` | `0x3B230` | 508 | Case-sensitive comparison wrapper | GDF-R01a | Separator spelling comparison |
| `LOD113C-STORM` | `0x55A0` | — | Bounded case-sensitive comparator | GDF-R01a | Ten bytes including Expansion terminator |
| `LOD113C-D2COMMON` | `0x5A8A0` | — | Derived monster animation writer | GDF-Q01 | Writer identified; complete animation-selection producers unread |
| `LOD113C-D2COMMON` | `0x43C30` | 10406 | Animation selection dependency | GDF-Q01 | Identified callee; complete producer contract not traced |
| `LOD113C-D2COMMON` | `0x41DD0` | — | Animation selection dependency | GDF-Q01 | Identified callee; complete producer contract not traced |
| `LOD113C-D2COMMON` | `0x33110` | 10819 | Animation-rate state selection | GDF-R11b | Selects state 101; full rate calculation outside boundary |
| `LOD113C-D2COMMON` | `0x37DB0` | — | Extended-list state membership test | GDF-R11b | Extended-list gate and membership bits; consumes NRF-R12a |
| `LOD113C-D2COMMON` | `0x33CD0` | 10494 | Public state-test wrapper | GDF-R11b, GDF-R19a | Unit types, state-count bound and membership callee |
| `LOD113C-D2COMMON` | `0x31D20` | 10212 | Block calculation field consumer | GDF-R03a, GDF-R18 | Class bounds, unsigned BlockFactor and NoShldBlock reads; full block behavior outside boundary |
| `LOD113C-D2COMMON` | `0x32890` | 10672 | Defense calculation stat selector | GDF-R09, GDF-R11b | Definition-count bound and stat 171/layer 0; consumes NRF-R12b/c |
| `LOD113C-D2COMMON` | `0x38A80` | — | Stat read dependency | GDF-R11b | Owned by NRF-R12b; not independently reverified here |
| `LOD113C-D2COMMON` | `0x5E1E0` | — | Localized class display-name writer | GDF-Q13 | Writer identified; imported localization/string path incomplete |
| `LOD113C-D2COMMON` | `0x5A030` | — | Treasure text loader and selected conversion | GDF-R10a, GDF-R10b, GDF-Q10 | Descriptor decode and copy block; probability/counter lifecycle incomplete |
| `LOD113C-D2COMMON` | `0x547D0` | — | Generated item-type treasure classes | GDF-Q10 | Increments group counter; lifecycle incomplete |
| `LOD113C-D2COMMON` | `0x53780` | — | Treasure runtime allocation | GDF-Q10 | Allocation/zeroing identified; probability producers incomplete |
| `LOD113C-D2COMMON` | `0x53C80` | — | Extended-monster descriptor callback | GDF-Q06, GDF-Q12 | Short-code parsing and per-slot counts/values; lookup integration incomplete |
| `LOD113C-D2GAME` | `0x1220` | — | Monster record getter | GDF-R05a, GDF-R19a | Reconstruction checked against original machine code; bounds and stride |
| `LOD113C-D2GAME` | `0x1560` | — | Extended-monster mode availability | GDF-R06b, GDF-R19a | Disassembly/image corroboration; signed link, range/null guards, byte/mask return |
| `LOD113C-D2GAME` | `0x1420` | — | Skill record getter | GDF-R08a, GDF-R17b, GDF-R19a | Signed bounds and stride |
| `LOD113C-D2GAME` | `0x1A40` | — | Level-scaled skill parameter getter | GDF-R08b, GDF-R19a | Index/record/level guards then wrapping Param3+(level-1)*Param4 |
| `LOD113C-D2GAME` | `0x77FA0` | — | Skill interruption selector | GDF-R17b, GDF-R18 | Compares srvdofunc 67/76 before interrupt bit |
| `LOD113C-D2GAME` | `0xDBA40` | — | Drain-effectiveness consumer | GDF-R05a | Unsigned difficulty-selected Drain; full leech behavior outside boundary |
| `LOD113C-D2GAME` | `0xDE0C0` | — | Monster town flag consumer | GDF-R18 | Monster getter then byte +0x0D/mask[2]; full damage behavior outside boundary |
| `LOD113C-D2GAME` | `0xDCF20` | — | Second monster town flag consumer | GDF-R18 | Same selected-record/flag chain; full combat behavior outside boundary |
| `LOD113C-D2GAME` | `0xDE5A0` | — | State-54 call site | GDF-R11b | Passes state 54/unit to imported state wrapper |
| `LOD113C-D2GAME` | `0x770E0` | — | Monster critical-chance field consumer | GDF-R05a | Crit field attribution; full doubling behavior belongs to NRF |
| `LOD113C-D2COMMON` | `0x9288` | — | Text-parser import thunk | GDF-R01a | Fog #10208 at RVA 0x12770 |
| `LOD113C-D2COMMON` | `0x928E` | — | Row-count import thunk | GDF-R01a | Fog #10210 at RVA 0x11F80 |
| `LOD113C-D2COMMON` | `0x9294` | — | Conversion import thunk | GDF-R01b | Fog #10207 at RVA 0x12A60 |
| `LOD113C-FOG` | `0xCE5A` | — | Header-hash import thunk | GDF-R01d | Storm #502 at RVA 0x3AC10 |
| `LOD113C-FOG` | `0xCE6C` | — | Header-compare import thunk | GDF-R01d | Storm #509 at RVA 0x3B1E0 |
| `LOD113C-FOG` | `0xCE72` | — | Separator-compare import thunk | GDF-R01a | Storm #508 at RVA 0x3B230 |
| `LOD113C-D2GAME` | `0xA68E` | — | State-test import thunk | GDF-R11b | D2Common #10494 at RVA 0x33CD0 |

### Dispatch and imported paths

D2Common `0x66160` reaches Difficulty `0x603F0`, CharStats `0x60A80`, Experience `0x60920`, Missiles `0x7C8D0`, Skills/SkillDesc `0x61EC0`, ItemStatCost `0x73330`, States `0x5F6E0`, MonMode `0x19330`, player types/modes `0x19650` and PetType `0x5F310`. Monster initialization is `0x66160 → 0x5D570 → 0x5AA40 / 0x56C10`, with derived animation writer `0x5A8A0`. Treasure initialization is `0x66160 → 0x5A730`, generated classes `0x547D0`, then loader `0x5A030` and allocation `0x53780`. ItemTypes `0x72510` is anchored by named output/stride/stores; its root caller was not established historically. These are initialization paths, not per-frame dispatch.

Loaders pass table name, descriptor array, opaque context, optional count destination and record stride to D2Common `0x5EF40`. Resource helper is `0x9900`. Text parsing uses D2Common thunks `0x9288 → Fog #10208 / 0x12770`, `0x928E → Fog #10210 / 0x11F80`, and `0x9294 → Fog #10207 / 0x12A60`. Generation uses writer `0x5D770`, including output wrapper `0x215D → 0x2056`. Complete resource/conversion/writer integration remains gated.

Header binding: Fog `0x12290` uses thunk `0xCE5A → Storm #502 / 0x3AC10` with case-normalizing flag 0, then `0xCE6C → Storm #509 / 0x3B1E0 → 0x54BC → ASCII comparator 0x7660`. Locale alternatives remain outside the verified slice. Separator parsing uses `0xCE72 → Storm #508 / 0x3B230 → 0x55A0`, a case-sensitive ten-byte comparison including the terminator of `Expansion`. This does not establish removal of differently cased separators.

State path: D2Game `0xDE5A0 → thunk 0xA68E → D2Common #10494 / 0x33CD0 → 0x37DB0`. The wrapper bounds the state identity by the state count and admits unit types 0, 1 and 3. Membership reads the unit stat list, extended-list flag and state-bit word/mask. NRF-R12a owns that consumed operation. D2Common `0x33110` selects state 101 before membership. D2Common `0x32890` selects fixed stat 171/layer 0 then stat reader `0x38A80`; NRF-R12b/c owns the read result.

### Branch and correction rationale

- D2Common `0x5EF40` independently reads generation flag RVA `0xA145C` and final compiled-read flag RVA `0x99E20`. Generation requests text, prepares intermediate rows, writes them through `0x5D770`, then discards them. Final selection independently requests `.bin` or `.txt`. A missing selected resource aborts; no compiled-to-text fallback exists on the traced path. Generation followed by final text reading performs a separate second decode. The historical parser identity and compiled-first fallback descriptions were corrected by these reads.
- Text allocation zeroes `count * stride` bytes before Fog conversion. Numeric handler RVA `0x13169` inside Fog `0x12A60` accumulates wrapping i32 decimal digits and optionally negates (loop `0x13177` through `0x13193`), without saturation. Dispatch table `0x13450` selects direct numeric/flag paths for kinds 2/3/4/6/0x1A. Flag handler `0x131CB` sets on parsed nonzero and clears on zero while preserving other bits. Direct four-code handler `0x12F62` pads with spaces; bounded-string handler `0x12F26` truncates by auxiliary byte count and terminates. Kind 0xA additionally registers code links in a first pass; it differs from direct kind 9.
- Missing-field switches, name/code links, expressions and descriptor callbacks were inspected but not fully resolved. Zero-initialization and present empty-numeric zero do not establish a universal missing-column default. Fog insertion `0x12510` / lookup `0x121A0` do not close those gaps. MonStats2 callback resolves to D2Common `0x53C80`; its slots 0..15, comma-separated short codes and per-slot count/value writes are descriptor-specific, not a general meaning of kind 0x17.
- Animation path `0x66160 → 0x5D570 → 0x5AA40 → 0x5A8A0` reads BaseId, uses selectors `0x43C30 / 0x41DD0`, scales Velocity/Run, clamps signed results to 0..32767 and stores words +0x36/+0x38. No loader column names those words. Complete selection/production remains GDF-Q01.
- PetType loader `0x5F310` and its root caller establish count ownership. It clamps nonpositive counts to zero and rejects signed counts >=256. Count 20 alone could not identify PlrMode. CharStats `0x60A80 → 0x5E1E0` identifies the localized-name writer, but imported localization/string semantics remain GDF-Q13. Mode publication stores were identified separately.
- D2Common `0x32890` uses count >171, definition displacement `171 * 0x144 = 0xD86C`, packed stat key `0xAB0000`: stat 171 `skill_armor_percent`, not 172 `alignment`. State 54 `uninterruptable` is a state, not a stat. `0x33110 → 0x37DB0` tests state 101 `holyshield`.
- D2Game `0x77FA0 → 0x1420` compares Skills +0x2E `srvdofunc` with 67/76 before reading interrupt. They are handler selectors, not skill ids. Byte +7 bit 7 is bit 31, distinct from leftskill bit 30. Missile GetHit/SoftHit are distinct from ClientSend.
- MonStats byte +0x0D/mask[2] means inTown bit 10, not MonStats2. Byte +0x0F/mask[2] means NoShldBlock bit 26. Mask-table displacement +8 selects dword mask[2], low byte 4. D2Game `0xDBA40` reads Drain unsigned. MonMode row 13 is `knockback`; earlier `combat-knockback` was a document-renaming artifact.

## Descriptor and consumer evidence map

These original layouts connect named fields to consumers; they do not require reproducing original memory storage. Current semantic requirements and labels remain in the spec. A descriptor is five little-endian dwords: name pointer, conversion kind, auxiliary value, destination offset (or callback slot), and link/callback context. Conversion kind 0 terminates; a name `end` does not terminate a nonzero-kind descriptor.

Direct kinds established here are 1 bounded byte string, 2 i32, 3 low16, 4/6 low8, 9 four-byte code, 0x1A bit update with auxiliary bit index. Other kinds identify links/expressions/callbacks whose full conversion remains gated. Consumer signedness follows its read instruction. A schema identity alone proves neither a cell value nor callback execution.

| Collection | D2Common loader RVA | Stride | Original publication/container location | Rules |
|---|---|---|---|---|
| DifficultyLevels | 0x603F0 | 0x58 | pointer RVA 0xA0B54; count 0xA0B58, required 3 | GDF-R02 |
| CharStats | 0x60A80 | 0xC4 | DataTables +0xBC4 / count +0xBC8 | GDF-R03a/b |
| Experience | 0x60920 | 0x20 | pointer RVA 0xA0B50; no separate count requested | GDF-R04 |
| MonStats | 0x5AA40 | 0x1A8 | DataTables +0xA78 / count +0xA80 | GDF-R05a |
| MonStats2 | 0x56C10 | 0x134 | DataTables +0xA90 / count +0xA98 | GDF-R06a/b |
| Missiles | 0x7C8D0 | 0x1A4 | DataTables +0xB64 / count +0xB6C | GDF-R07 |
| Skills/SkillDesc | 0x61EC0 | 0x23C for Skills | Skills +0xB98 / +0xBA0; SkillDesc +0xB8C / +0xB94 | GDF-R08a/b |
| ItemStatCost | 0x73330 | 0x144 | DataTables +0xBCC / count +0xBD4 | GDF-R09 |
| States | 0x5F6E0 | 0x3C | DataTables +0xBC / count +0xC4 | GDF-R13 |
| ItemTypes | 0x72510 | 0xE4 | DataTables +0xBF8 / count +0xBFC | GDF-R16b |
| TreasureClassEx text | 0x5A030 | 0x2E0 | distinct runtime conversion below | GDF-R10a/b |
| PetType count | 0x5F310 | not needed for count slice | DataTables +0xBF0, static count RVA 0xA0AC8 | GDF-R19b |
| MonMode | 0x19330 | 0x34 | count RVA 0xA13A4; bases 0xA13A8/AC/B0 | GDF-R15b |
| Player types/modes | 0x19650 | 0x34 | count 0xA13C4; bases 0xA13C8/CC; mode subset 0xA13D0 | GDF-R14b |

### Difficulty field offsets

| Offset | Column | Cited by |
|--------|--------|----------|
| +0x00 | `ResistPenalty` | combat-damage-mitigation ("first dword", expansion resist penalty) |
| +0x04 | `DeathExpPenalty` | player-party-experience-and-loot-sharing |
| +0x08 | `UberCodeOddsNormal` | none |
| +0x0C | `UberCodeOddsGood` | none |
| +0x10 | `MonsterSkillBonus` | monster-scaling-by-players-and-difficulty |
| +0x14 | `MonsterFreezeDivisor` | state-movement-modifiers-and-freeze, combat-status-application (freeze divisor); CSA-R08 |
| +0x18 | `MonsterColdDivisor` | state-movement-modifiers-and-freeze, combat-status-application (cold divisor); CSA-R05 |
| +0x1C | `AiCurseDivisor` | none |
| +0x20 | `UltraCodeOddsNormal` | none |
| +0x24 | `UltraCodeOddsGood` | none |
| +0x28 | `LifeStealDivisor` | combat-leech-and-status-effects (life divisor) |
| +0x2C | `ManaStealDivisor` | combat-leech-and-status-effects (mana divisor) |
| +0x30 | `UniqueDamageBonus` | none |
| +0x34 | `ChampionDamageBonus` | none |
| +0x38 | `HireableBossDamagePercent` | combat-pvp-damage-and-kill-credit, combat-damage-pipeline-order (hireling-vs-monster scale `P`) |
| +0x3C | `MonsterCEDamagePercent` | none |
| +0x40 | `StaticFieldMin` | none |
| +0x44 / +0x48 / +0x4C / +0x50 / +0x54 | `GambleRare` / `GambleSet` / `GambleUnique` / `GambleUber` / `GambleUltra` | none |

### Class field offsets

| Offset | Column | Type | Cited by |
|--------|--------|------|----------|
| +0x3C | `ToHitFactor` | 2 | combat-shield-block-and-avoidance (per-class attack-rating bonus) |
| +0x40 | `WalkVelocity` | 4 | state-movement-modifiers-and-freeze; NRF-R08b, NRF-R11a |
| +0x41 | `RunVelocity` | 4 | state-movement-modifiers-and-freeze |
| +0x49 | `BlockFactor` | 4 | combat-shield-block-and-avoidance (per-class block byte; `movzx ecx, byte ptr [eax + 0x49]`, RVA 0x31D20) |

### Monster field offsets

| Offset | Column | Type | Cited by |
|--------|--------|------|----------|
| +0x00 | `Id` (prepared identity; not guaranteed to equal array index) | 0x11 | all monster specs (class id, GDF-R17); global cached correspondence is limited below |
| +0x02 | `BaseId` | 0x14 | combat-knockback, combat-hit-recovery (class 0x4E test) |
| +0x0C | flag dword; bits in GDF-R18 | 0x1A | GDF-R18 |
| +0x18 | `MonStatsEx` (index of the MonStats2 row; read as signed 16-bit, `movsx`, by D2Game RVA 0x1560) | 0x14 | combat-hit-recovery, combat-knockback (route to `MonStats2Record`) |
| +0x32 | `Velocity` | 3 | state-movement-modifiers-and-freeze (monster velocity word) |
| +0x34 | `Run` | 3 | none |
| +0x3A | `MissA1` | 0x14 | none; bounds the descriptor gap at derived words +0x36/+0x38 |
| +0x4C | `Align` | 4 | monster-scaling-by-players-and-difficulty (byte "noScaling") |
| +0x4F / +0x50 / +0x51 | `aidel` / `aidel(N)` / `aidel(H)` | 4 | state-movement-modifiers-and-freeze |
| +0x56 + 6k + 2d | `aip(k+1)` for difficulty d = 0/1/2 (`aip1` +0x56, `aip2` +0x5C, `aip3` +0x62, `aip4` +0x68, `aip5` +0x6E, `aip6` +0x74, `aip7` +0x7A, `aip8` +0x80; the `(N)`/`(H)` columns follow each at +2/+4) | 3 | skill-assassin-sentry-firing-and-range, skill-assassin-death-sentry-corpse-damage |
| +0xA0 / +0xA1 / +0xA2 | `Drain` / `Drain(N)` / `Drain(H)` | 4 | combat-leech-and-status-effects (drain effectiveness, read **unsigned**: `movzx ecx, byte ptr [edx + ecx + 0xA0]` with `ecx` = `Game +0x6D`, D2Game RVA 0xDBA40; earlier text said signed, corrected; now LSE-R03) |
| +0xA3 / +0xA4 / +0xA5 | `ToBlock` / `ToBlock(N)` / `ToBlock(H)` | 4 | none |
| +0xA6 | `Crit` | 4 | NRF-R04a (monster damage-doubling roll, D2Game rva 0x770E0) |
| +0xAA / +0xAC / +0xAE | `Level` / `Level(N)` / `Level(H)` | 3 | monster-scaling-by-players-and-difficulty |
| +0xB0 / +0xB2 / +0xB4 | `MinHP` / `(N)` / `(H)` | 3 | monster-scaling; NRF-R04a |
| +0xB6 / +0xB8 / +0xBA | `MaxHP` / `(N)` / `(H)` | 3 | monster-scaling; NRF-R04a |
| +0xBC / +0xBE / +0xC0 | `AC` / `(N)` / `(H)` | 3 | monster-scaling |
| +0xD4 / +0xD6 / +0xD8 | `Exp` / `(N)` / `(H)` | 3 | monster-scaling |
| +0x168 / +0x169 / +0x16A | `ColdEffect` / `(N)` / `(H)` | 4 | state-movement-modifiers-and-freeze (read signed) |
| +0x170 / +0x172 ... +0x17E | `Skill1` / `Skill2` ... `Skill8` (16-bit skill index) | 0x14 | skill-assassin-sentry-firing-and-range (`nSkill[0]` = `Skill1`), skill-assassin-death-sentry-corpse-damage (`nSkill[1]` = `Skill2`) |
| +0x198..+0x19F | `Sk1lvl`..`Sk8lvl` | 4 | none |

### Extended monster fields

| Offset | Column(s) | Type | Cited by |
|--------|-----------|------|----------|
| +0x04 | flag dword: `noGfxHitTest` bit 0, `noMap` 1, `noOvly` 2, `isSel` 3, `alSel` 4, `noSel` 5, `shiftSel` 6, `corpseSel` 7, `revive` 8, `isAtt` 9, **`small` 10**, **`large` 11**, `soft` 12, `critter` 13, `shadow` 14, `noUniqueShift` 15, `compositeDeath` 16, `inert` 17, `objCol` 18, `deadCol` 19, `unflatDead` 20 | 0x1A | combat-knockback (size class bits 10 and 11) |
| +0x14 | `HitClass` | 4 | none (candidate melee hit-class source, combat-hit-recovery HRC-R07: unresolved there) |
| +0xF0 | mode-availability dword: `mDT` bit 0, `mNU` 1, `mWL` 2, **`mGH` 3**, `mA1` 4, `mA2` 5, `mBL` 6, `mSC` 7, `mS1` 8, `mS2` 9, `mS3` 10, `mS4` 11, `mDD` 12, **`mKB` 13**, `mSQ` 14, `mRN` 15 | 0x1A | combat-knockback, combat-hit-recovery (bits 3 and 13); bit n = monster mode n (GDF-R15); D2Game RVA 0x1560 tests `byte [rec + 0xF0 + (mode >> 3)] & mask[mode & 7]` and returns 0 when the monster class is outside `[0, DataTables +0xA80)`, when the signed `MonStatsEx` word is negative or `>= DataTables +0xA98`, or when the MonStats2 record address is 0 |
| +0xF4..+0x103 | `dDT`..`dRN` (direction count per mode, one byte each, mode order as `+0xF0`: `dDT` +0xF4, `dGH` +0xF7, `dKB` +0x101, `dRN` +0x103) | 4 | combat-knockback (`dKB` +0x101, not consulted) |

### Missile fields

| Offset | Column | Type | Cited by |
|--------|--------|------|----------|
| +0x04 | flag dword: `LastCollide` 0, `Explosion` 1, `Pierce` 2, `CanSlow` 3, `CanDestroy` 4, `ClientSend` 5, **`GetHit` 6**, **`SoftHit` 7**, `ApplyMastery` 8, `ReturnFire` 9, `Town` 10, `SrcTown` 11, `NoMultiShot` 12, `NoUniqueMod` 13, `Half2HSrc` 14, `MissileSkill` 15 | 0x1A | combat-hit-recovery (bits 6, 7) |
| +0x78 / +0x7C | `dParam1` / `dParam2` | 2 | combat-knockback |
| +0x94 | `HitClass` | 4 | combat-hit-recovery (missile hit-class low nibble) |
| +0xAE | `KnockBack` | 4 | combat-knockback; NRF-R04a |
| +0x187 | `CollideFriend` | 4 | combat-pvp-damage-and-kill-credit (contact with a non-hostile unit) |
| +0x188 | `NextHit` | 4 | skill-assassin-blade-sentinel-and-fury |
| +0x189 | `NextDelay` | 4 | skill-assassin-blade-sentinel-and-fury |
| +0x18B | `ToHit` | 4 | combat-shield-block-and-avoidance SBA-R15a (flag enabling the missile hit roll) |

### Skill fields

| Offset | Column | Type | Cited by |
|--------|--------|------|----------|
| +0x04 | flag dword; bit 31 = **`interrupt`** (byte +0x07 bit 7: D2Game RVA 0x77FA0 `mov al, byte ptr [eax + 7]; test byte ptr [ecx + 0x1C], al` with `ecx` = `gdwBitMasks`, `mask[7]` = 0x80) | 0x1A | combat-hit-recovery (skill record byte +7 bit 7) |
| +0x2C | `srvstfunc` | 3 | skill-assassin-blade-sentinel-and-fury (start function index) |
| +0x2E | `srvdofunc` | 3 | skill-assassin-blade-sentinel-and-fury, skill-assassin-sentry-firing-and-range (do-function index); GDF-R17 |
| +0x50 | `aurafilter` | 2 | skill-assassin-death-sentry-corpse-damage |
| +0x64 | `aurarangecalc` | 0x19 | skill-assassin-death-sentry-corpse-damage |
| +0x138 / +0x13C / +0x140 / +0x144 | `calc1` / `calc2` / `calc3` / `calc4` | 0x19 | skill-assassin-sentry-firing-and-range, skill-assassin-death-sentry-corpse-damage |
| +0x148..+0x164 (step 4) | `Param1`..`Param8` | 2 | skill-assassin-death-sentry-corpse-damage (`Param3` +0x150, `Param4` +0x154; see note) |
| +0x16A | `ItemEffect` | 3 | skill-assassin-blade-sentinel-and-fury (item-effect override) |
| +0x198 / +0x19C | `ToHit` / `LevToHit` | 2 | skill-leveled-damage-tohit-and-elements |
| +0x1A0 | `ToHitCalc` | 0x19 | skill-leveled-damage-tohit-and-elements |
| +0x1A4 / +0x1A5 | `HitShift` / `SrcDam` | 4 | skill-leveled-damage-tohit-and-elements |
| +0x1A8 / +0x1AC | `MinDam` / `MaxDam` | 2 | skill-leveled-damage-tohit-and-elements |
| +0x1B0..+0x1C0 (step 4) | `MinLevDam1`..`MinLevDam5` | 2 | skill-leveled-damage-tohit-and-elements |
| +0x1C4..+0x1D4 (step 4) | `MaxLevDam1`..`MaxLevDam5` | 2 | skill-leveled-damage-tohit-and-elements |
| +0x1D8 | `DmgSymPerCalc` | 0x19 | skill-leveled-damage-tohit-and-elements |
| +0x1DC | `EType` | 0xD | skill-assassin-death-sentry-corpse-damage (element field), skill-leveled-damage-tohit-and-elements |
| +0x1E0 / +0x1E4 | `EMin` / `EMax` | 2 | skill-leveled-damage-tohit-and-elements |
| +0x1E8..+0x1F8 (step 4) | `EMinLev1`..`EMinLev5` | 2 | skill-leveled-damage-tohit-and-elements |
| +0x1FC..+0x20C (step 4) | `EMaxLev1`..`EMaxLev5` | 2 | skill-leveled-damage-tohit-and-elements |
| +0x210 | `EDmgSymPerCalc` | 0x19 | skill-leveled-damage-tohit-and-elements |
| +0x214 | `ELen` | 2 | skill-leveled-damage-tohit-and-elements |
| +0x218 / +0x21C / +0x220 | `ELevLen1` / `ELevLen2` / `ELevLen3` | 2 | skill-leveled-damage-tohit-and-elements |
| +0x224 | `ELenSymPerCalc` | 0x19 | skill-leveled-damage-tohit-and-elements |

Experience (GDF-R04): Amazon +0x00, Sorceress +0x04, Necromancer +0x08, Paladin +0x0C, Barbarian +0x10, Druid +0x14, Assassin +0x18, ExpRatio +0x1C, all kind 2. This does not establish the consumer interpretation of row zero.

Class context: class +0x20 (kind 1, limit 15); str/dex/int/vit/stamina/hpadd +0x30..+0x35; ManaRegen +0x3A; RunDrain +0x42; LifePerLevel/StaminaPerLevel/ManaPerLevel +0x43..+0x45; LifePerVitality/StaminaPerVitality/ManaPerMagic +0x46..+0x48; StatPerLevel +0x50 (kind 4), StartSkill +0xAC (linked kind 0x14). Earlier display-name bytes are post-load output. Skill Param3/Param4 are +0x150/+0x154; getter `0x1A40` guards index, absent record and nonpositive level before reading them.

### Stat-definition fields

Cited fields: `itemevent1` +0x48 (0x14), `itemeventfunc1` +0x4C (3), `itemevent2` +0x4A (0x14), `itemeventfunc2` +0x4E (3) (combat-knockback, combat-item-event-triggers-and-procs). Flag dword +0x04: `send other` 0, `signed` 1, `damagerelated` 2, `itemspecific` 3, `direct` 4, `updateanimrate` 9, `fmin` 10, `fcallback` 11, `saved` 12, `csvsigned` 13. Also `valshift` +0x18 (4), `minaccr` +0x2C (2) (NRF-R12c; [descriptor/consumer evidence](core-numeric-and-rng-foundations.md#stat-reader-verification-2026-10-14)), `maxstat` +0x32 (0x14), and for orientation `send bits` +0x08 (4), `send param bits` +0x09 (4), `divide` +0x0C (2), `multiply` +0x10 (2), `add` +0x14 (2), `op` +0x54 (4), `op base` +0x56 (0x14), `op stat1..3` +0x58/+0x5A/+0x5C (0x14). The descriptor names stat while the file header names Stat; the verified ASCII binding path explains this. Missing/duplicate behavior remains gated.

### State-definition fields

`state` +0x00 (row index), `overlay1..4` +0x02/+0x04/+0x06/+0x08, `castoverlay` +0x0A, `removerlay` +0x0C, `pgsvoverlay` +0x0E (0x14), flag dword +0x10 (bits: `nosend` 0, `aura` 1, `hide` 2, `transform` 3, `pgsv` 4, `active` 5, `remhit` 6, `damblue` 7, `damred` 8, `attblue` 9, `attred` 10, `curse` 11, `curable` 12, `plrstaydeath` 13, `monstaydeath` 14, `bossstaydeath` 15, `disguise` 16, `restrict` 17, `blue` 18, `armblue` 19, `rfblue` 20, `rcblue` 21, `rlblue` 22, `rpblue` 23, `stambarblue` 24, `armred` 25, `rfred` 26, `rcred` 27, `rlred` 28, `rpred` 29, `exp` 30, `shatter` 31, `life` 32, `udead` 33, `green` 34, `nooverlays` 35, `noclear` 36, `bossinv` 37, `meleeonly` 38, `notondead` 39; bits above 31 continue in the dword at +0x14), `stat` +0x18 (0x14), `setfunc` +0x1A (3), `remfunc` +0x1C (3), `group` +0x1E (3), `colorpri` +0x20 (4), `colorshift` +0x21 (4), `light-r` +0x22 (4), `light-g` +0x23 (4), `light-b` +0x24 (4), `onsound` +0x26 (0x14), `offsound` +0x28 (0x14), `itemtype` +0x2A (0xF), `itemtrans` +0x2C (0xD), `gfxtype` +0x2D (4), `gfxclass` +0x2E (3), `cltevent` +0x30 (0x14), `clteventfunc` +0x32 (3), `cltactivefunc` +0x34 (3), `srvactivefunc` +0x36 (3), `skill` +0x38 (0x14), `missile` +0x3A (0x14).

### Item-type fields

Loader fields (`itemtypes`, D2Common rva 0x72510; record size 0xE4; count slot `DataTables +0xBFC`): `code` +0x00 (type 0xA), `equiv1` +0x04 (0xF), `equiv2` +0x06 (0xF), `repair` +0x08 (4), `body` +0x09 (4), `bodyloc1` +0x0A (0xD), `bodyloc2` +0x0B (0xD), `shoots` +0x0C (0xF), `quiver` +0x0E (0xF), then one type-4 byte each: `throwable` +0x10, `reload` +0x11, `reequip` +0x12, `autostack` +0x13, `magic` +0x14, `rare` +0x15, `normal` +0x16, `charm` +0x17, `gem` +0x18, `beltable` +0x19, `maxsock1` +0x1A, `maxsock25` +0x1B, `maxsock40` +0x1C, `treasureclass` +0x1D, `rarity` +0x1E; `staffmods` +0x1F (0xD), `costformula` +0x20 (4), `class` +0x21 (0xD), `storepage` +0x22 (0xD), `varinvgfx` +0x23 (4), `invgfx1`..`invgfx6` +0x24 step 0x20 (string, 31).

### Treasure conversion

D2Common `0x5A030` text descriptors: treasure class +0x00 (string limit 32), picks +0x20 (i32), group/level +0x24/+0x26 (words), magic/rare/set/unique +0x28/+0x2A/+0x2C/+0x2E (words), nodrop +0x34 (i32), `item1`..10 +0x38 step 0x40 (string limit 64), `prob1`..10 +0x2B8 step 4 (i32). Source words +0x30/+0x32 have no descriptor names.

Selected conversion block RVA `0x5A556` adds nonzero group to generated counter RVA `0xA1460`, stores low16 at runtime +0x00; zero group leaves destination unchanged. Level copies to +0x02; Picks to +0x10, replacing zero with one; NoDrop to +0x14. Quality words copy to +0x1A/+0x1C/+0x1E/+0x20; reserved source words +0x30/+0x32 copy to +0x22/+0x24 in order. Runtime probability sums +0x08/+0x0C are computed, not columns. Complete probability/group lifecycle and runtime entry production remain GDF-Q10.

### Lookup and flag context

Descriptor link-context DataTables slots: monster names +0xA7C, extended monster names +0xA94, skill names +0xB9C (separate skill link +0x70), stat names +0xBD0, missile names +0xB68, events +0x78, item-type codes +0xBF4, elements +0x1C, body locations +0x0C, classes +0x04, store pages +0x14, overlays +0xBB8, sounds +0x198, transforms +0x9B4. Associations do not establish complete linking/callback semantics.

MonStats +0x0C bits: isSpawn 0, isMelee 1, noRatio 2, opendoors 3, SetBoss 4, BossXfer 5, boss 6, primeevil 7, npc 8, interact 9, inTown 10, lUndead 11, hUndead 12, demon 13, flying 14, killable 15, switchai 16, nomultishot 17, neverCount 18, petIgnore 19, deathDmg 20, genericSpawn 21, zoo 22, placespawn 23, inventory 24, enabled 25, NoShldBlock 26, noaura 27, rangedtype 28.

Skills +0x04 bits: decquant 0, lob 1, progressive 2, finishing 3, passive 4, aura 5, periodic 6, prgstack 7, InTown 8, Kick 9, InGame 10, repeat 11, stsuccessonly 12, stsounddelay 13, weaponsnd 14, immediate 15, noammo 16, enhanceable 17, durability 18, UseAttackRate 19, TargetableOnly 20, SearchEnemyXY 21, SearchEnemyNear 22, SearchOpenXY 23, TargetCorpse 24, TargetPet 25, TargetAlly 26, TargetItem 27, AttackNoMana 28, ItemTgtDo 29, leftskill 30, interrupt 31. Next dword: TgtPlaceCheck 32, ItemCheckStart 33, ItemCltCheckStart 34, general 35, scroll 36, usemanaondo 37, warp 38. This does not establish every flag consumer.

MonMode prepares name +0x00 (limit 31), token +0x20, and direction bytes DT/NU/WL/GH/A1/A2/BL/SC/S1/S2/S3/S4/DD/KB/SQ/RN at +0x24..+0x33. Player modes prepare name/token. Code is not a prepared field. Neither mode text file has direction columns; missing-column behavior remains gated.

## Original data locations

| Artifact | Data RVA | Export ordinal | Analyst description | Rules |
|---|---|---|---|---|
| `LOD113C-D2COMMON` | `0x99E1C` | 11170 | DataTables pointer export | GDF-R19a |
| `LOD113C-FOG` | `0x26798` | 10259 | 32-dword mask table, mask[n] = 1 << n | GDF-R18 |
| `LOD113C-D2COMMON` | `0xA145C` | — | Generation flag | GDF-R20a |
| `LOD113C-D2COMMON` | `0x99E20` | — | Final compiled-read flag | GDF-R20a |
| `LOD113C-D2COMMON` | `0xA1460` | — | Generated treasure group counter | GDF-Q10 |
| `LOD113C-D2COMMON` | `0xA0B54` | — | Difficulty pointer | GDF-R02 |
| `LOD113C-D2COMMON` | `0xA0B58` | — | Difficulty count | GDF-R02 |
| `LOD113C-D2COMMON` | `0xA0B50` | — | Experience pointer | GDF-R04 |
| `LOD113C-D2COMMON` | `0xA0AC8` | — | PetType count | GDF-R19b |
| `LOD113C-D2COMMON` | `0xA13A4` | — | MonMode count | GDF-R15b |
| `LOD113C-D2COMMON` | `0xA13A8` | — | MonMode pointer | GDF-R15b |
| `LOD113C-D2COMMON` | `0xA13AC` | — | MonMode pointer copy | GDF-R15b |
| `LOD113C-D2COMMON` | `0xA13B0` | — | MonMode pointer copy | GDF-R15b |
| `LOD113C-D2COMMON` | `0xA13C4` | — | Combined player type/mode count | GDF-R14b |
| `LOD113C-D2COMMON` | `0xA13C8` | — | Combined player type/mode pointer | GDF-R14b |
| `LOD113C-D2COMMON` | `0xA13CC` | — | Combined player type/mode pointer copy | GDF-R14b |
| `LOD113C-D2COMMON` | `0xA13D0` | — | Player-mode subset pointer | GDF-R14b |

Fog masks are imported at D2Game IAT RVA `0xF8BE0` and D2Common IAT RVA `0x890B0`. DataTables is imported at D2Game IAT RVA `0xF829C`. These slots resolve to the references above, not separate container definitions.

<a id="table-verification-2026-10-14"></a>
## Table verification (2026-10-14)

Historical reads used the running original-profile game's own archive loader. Reported vanilla 1.13c profile and recorded Game.exe identity are in the manifest. No game action, trace or memory write occurred. Physical supplying archives, effective-file hashes and complete overrides were not captured; no known overrides were reported.

All cells below were selected by decoded header name after joining contiguous returned byte ranges. Read chunks were at most 60000 bytes, or four bytes for count headers. Returned offsets/lengths and reported totals established coverage. Data-row indexing excludes the header and begins at 0. Exact Expansion separators are excluded for prepared-row correspondence; no missing row is inferred from a preview.

| Effective virtual path (data/global/excel/) | Requested offset+returned ranges | Total bytes | Coverage |
|--------------------------------------------|----------------------------------|-------------|----------|
| itemstatcost.txt | 0+40109 | 40109 | full header/file |
| difficultylevels.txt | 0+586 | 586 | full header/file |
| states.txt | 0+20920 | 20920 | full header/file |
| plrmode.txt | 0+302 | 302 | full header/file |
| monmode.txt | 0+238 | 238 | full header/file |
| itemtypes.txt | 0+8775 | 8775 | full header/file |
| skills.txt | 0+60000; 60000+60000; 120000+53535 | 173535 | full header/file |
| monstats.txt | 0+60000; 60000+60000; 120000+60000; 180000+60000; 240000+60000; 300000+60000; 360000+60000; 420000+12634 | 432634 | full header/file |
| `monstats2`.txt | 0+60000; 60000+60000; 120000+18542 | 138542 | full header/file |
| charstats.txt | 0+3360 | 3360 | full header/file |
| experience.txt | 0+7116 | 7116 | full header/file |
| missiles.txt | 0+60000; 60000+60000; 120000+60000; 180000+12979 | 192979 | full header/file |
| treasureclassex.txt | 0+60000; 60000+28253 | 88253 | full header/file |
| pettype.txt | 0+1288 | 1288 | full header/file |

### Compiled correspondence

| Effective virtual path (data/global/excel/) | Requested offset+returned ranges | Total bytes | Coverage / limits |
|--------------------------------------------|----------------------------------|-------------|-------------------|
| itemstatcost.bin | 0+60000; 60000+56320 | 116320 | full file; interpreted named cases only |
| difficultylevels.bin | 0+268 | 268 | full file; interpreted named cases only |
| states.bin | 0+11104 | 11104 | full file; interpreted named cases only |
| plrmode.bin | 0+1044 | 1044 | full file; interpreted named cases only |
| monmode.bin | 0+836 | 836 | full file; interpreted named cases only |
| itemtypes.bin | 0+23488 | 23488 | full file; interpreted named cases only |
| skills.bin | 0+60000; 60000+60000; 120000+60000; 180000+24208 | 204208 | full file; interpreted named cases only |
| monstats.bin | 0+60000; 60000+60000; 120000+60000; 180000+60000; 240000+60000; 300000+11220 | 311220 | full file; interpreted named cases only |
| `monstats2`.bin | 0+60000; 60000+60000; 120000+60000; 180000+7576 | 187576 | full file; interpreted named cases only |
| charstats.bin | 0+1376 | 1376 | full file; interpreted named cases only |
| experience.bin | 0+3236 | 3236 | full file; interpreted named cases only |
| missiles.bin | 0+60000; 60000+60000; 120000+60000; 180000+60000; 240000+47284 | 287284 | full file; interpreted named cases only |
| treasureclassex.bin | 0+60000; 60000+60000; 120000+60000; 180000+60000; 240000+60000; 300000+60000; 360000+60000; 420000+60000; 480000+60000; 540000+60000; 600000+27812 | 627812 | full file; interpreted named cases only |
| pettype.bin | 0+4 | 4484 | count header only |

These reads are archive-file reads, not live runtime memory reads. No binary payload, table, or complete row is copied into this repository. The interpreting layouts below are evidence details only:
- Four-byte compiled row count, followed by original fixed-size records; verified against 0x5EF40 selected compiled path.
- ItemStatCost stride 0x144, States 0x3C, Skills 0x23C: every compiled identity word matches all 359/185/357 corresponding explicit text ID/id/Id cells (zero mismatch).
- CharStats stride 0xC4: all seven class strings at +0x20 match text order with Expansion excluded; Sorceress BlockFactor +0x49=20, ManaRegen +0x3A=120, ToHitFactor +0x3C raw 0xFFFFFFF1=-15.
- MonStats stride 0x1A8: the 13 cited class rows have compiled Id words equal to their hcIdx; classes 78/560 MonStatsEx word +0x18=78/560, Velocity +0x32=10/8, Run +0x34=10/14; flag words +0x0C=0x02018003/0x0205840B. MonStats2 stride 0x134 Id words at 78/560 agree.
- ItemTypes stride 0xE4: original rows 4/51 Code bytes gold/shld.
- Skills selectors at +0x2E: Charge Id 107 srvdofunc 67, Whirlwind 151→76, SerpentCharge 352→67 match text cells.
- Difficulty stride 0x58: named vector fields and cold/freeze divisors match text; no new gameplay preparation is observed.
- PetType compiled header count 20 (total 4484 bytes) agrees with its 20 text rows (1288 bytes). Count-header-only coverage makes no per-record correspondence claim.
- MonMode and PlrMode .txt each have only name/token/code headers. Their schema-declared direction fields have no text columns; missing-column conversion remains gated rather than inferred from spacing.

### Named cell records

Each row is one named cell, interpreted only for its supported claim. Name-table records below are a dictionary's cited identities, not copies of game-table rows. TD-STAT/STATE explicit-id correspondence and filtered indexing are checked across complete files; this does not establish a configurable compiler binding.

| Record family | Effective virtual file under `data/global/excel/` |
|---|---|
| `TD-STAT` | `itemstatcost.txt` |
| `TD-STATE` | `states.txt` |
| `TD-DIFFICULTY` | `difficultylevels.txt` |
| `TD-CLASS` | `charstats.txt` |
| `TD-PLAYER-MODE` | `plrmode.txt` |
| `TD-MONSTER-MODE` | `monmode.txt` |
| `TD-ITEMTYPE` | `itemtypes.txt` |
| `TD-MONSTER-SKILL`, MonStats row | `monstats.txt` |
| `TD-MONSTER-SKILL`, Skills row | `skills.txt` |

| Record | Row identity / indexing | Column | Raw cell | Interpretation / correspondence | Supports |
|--------|-------------------------|--------|----------|----------------------------------|----------|
| TD-STAT | ItemStatCost ID 0, zero-based row 0 | Stat | strength | stat name; ID raw 0; prepared identity 0 | GDF-R11a |
| TD-STAT | ItemStatCost ID 2, zero-based row 2 | Stat | dexterity | stat name; ID raw 2; prepared identity 2 | GDF-R11a |
| TD-STAT | ItemStatCost ID 6, zero-based row 6 | Stat | hitpoints | stat name; ID raw 6; prepared identity 6 | GDF-R11a |
| TD-STAT | ItemStatCost ID 7, zero-based row 7 | Stat | maxhp | stat name; ID raw 7; prepared identity 7 | GDF-R11a |
| TD-STAT | ItemStatCost ID 8, zero-based row 8 | Stat | mana | stat name; ID raw 8; prepared identity 8 | GDF-R11a |
| TD-STAT | ItemStatCost ID 10, zero-based row 10 | Stat | stamina | stat name; ID raw 10; prepared identity 10 | GDF-R11a |
| TD-STAT | ItemStatCost ID 12, zero-based row 12 | Stat | level | stat name; ID raw 12; prepared identity 12 | GDF-R11a |
| TD-STAT | ItemStatCost ID 13, zero-based row 13 | Stat | experience | stat name; ID raw 13; prepared identity 13 | GDF-R11a |
| TD-STAT | ItemStatCost ID 14, zero-based row 14 | Stat | gold | stat name; ID raw 14; prepared identity 14 | GDF-R11a |
| TD-STAT | ItemStatCost ID 16, zero-based row 16 | Stat | item_armor_percent | stat name; ID raw 16; prepared identity 16 | GDF-R11a |
| TD-STAT | ItemStatCost ID 19, zero-based row 19 | Stat | tohit | stat name; ID raw 19; prepared identity 19 | GDF-R11a |
| TD-STAT | ItemStatCost ID 20, zero-based row 20 | Stat | toblock | stat name; ID raw 20; prepared identity 20 | GDF-R11a |
| TD-STAT | ItemStatCost ID 21, zero-based row 21 | Stat | mindamage | stat name; ID raw 21; prepared identity 21 | GDF-R11a |
| TD-STAT | ItemStatCost ID 22, zero-based row 22 | Stat | maxdamage | stat name; ID raw 22; prepared identity 22 | GDF-R11a |
| TD-STAT | ItemStatCost ID 23, zero-based row 23 | Stat | secondary_mindamage | stat name; ID raw 23; prepared identity 23 | GDF-R11a |
| TD-STAT | ItemStatCost ID 24, zero-based row 24 | Stat | secondary_maxdamage | stat name; ID raw 24; prepared identity 24 | GDF-R11a |
| TD-STAT | ItemStatCost ID 29, zero-based row 29 | Stat | lastexp | stat name; ID raw 29; prepared identity 29 | GDF-R11a |
| TD-STAT | ItemStatCost ID 31, zero-based row 31 | Stat | armorclass | stat name; ID raw 31; prepared identity 31 | GDF-R11a |
| TD-STAT | ItemStatCost ID 32, zero-based row 32 | Stat | armorclass_vs_missile | stat name; ID raw 32; prepared identity 32 | GDF-R11a |
| TD-STAT | ItemStatCost ID 33, zero-based row 33 | Stat | armorclass_vs_hth | stat name; ID raw 33; prepared identity 33 | GDF-R11a |
| TD-STAT | ItemStatCost ID 34, zero-based row 34 | Stat | normal_damage_reduction | stat name; ID raw 34; prepared identity 34 | GDF-R11a |
| TD-STAT | ItemStatCost ID 35, zero-based row 35 | Stat | magic_damage_reduction | stat name; ID raw 35; prepared identity 35 | GDF-R11a |
| TD-STAT | ItemStatCost ID 36, zero-based row 36 | Stat | damageresist | stat name; ID raw 36; prepared identity 36 | GDF-R11a |
| TD-STAT | ItemStatCost ID 37, zero-based row 37 | Stat | magicresist | stat name; ID raw 37; prepared identity 37 | GDF-R11a |
| TD-STAT | ItemStatCost ID 38, zero-based row 38 | Stat | maxmagicresist | stat name; ID raw 38; prepared identity 38 | GDF-R11a |
| TD-STAT | ItemStatCost ID 39, zero-based row 39 | Stat | fireresist | stat name; ID raw 39; prepared identity 39 | GDF-R11a |
| TD-STAT | ItemStatCost ID 40, zero-based row 40 | Stat | maxfireresist | stat name; ID raw 40; prepared identity 40 | GDF-R11a |
| TD-STAT | ItemStatCost ID 41, zero-based row 41 | Stat | lightresist | stat name; ID raw 41; prepared identity 41 | GDF-R11a |
| TD-STAT | ItemStatCost ID 42, zero-based row 42 | Stat | maxlightresist | stat name; ID raw 42; prepared identity 42 | GDF-R11a |
| TD-STAT | ItemStatCost ID 43, zero-based row 43 | Stat | coldresist | stat name; ID raw 43; prepared identity 43 | GDF-R11a |
| TD-STAT | ItemStatCost ID 44, zero-based row 44 | Stat | maxcoldresist | stat name; ID raw 44; prepared identity 44 | GDF-R11a |
| TD-STAT | ItemStatCost ID 45, zero-based row 45 | Stat | poisonresist | stat name; ID raw 45; prepared identity 45 | GDF-R11a |
| TD-STAT | ItemStatCost ID 46, zero-based row 46 | Stat | maxpoisonresist | stat name; ID raw 46; prepared identity 46 | GDF-R11a |
| TD-STAT | ItemStatCost ID 48, zero-based row 48 | Stat | firemindam | stat name; ID raw 48; prepared identity 48 | GDF-R11a |
| TD-STAT | ItemStatCost ID 49, zero-based row 49 | Stat | firemaxdam | stat name; ID raw 49; prepared identity 49 | GDF-R11a |
| TD-STAT | ItemStatCost ID 50, zero-based row 50 | Stat | lightmindam | stat name; ID raw 50; prepared identity 50 | GDF-R11a |
| TD-STAT | ItemStatCost ID 51, zero-based row 51 | Stat | lightmaxdam | stat name; ID raw 51; prepared identity 51 | GDF-R11a |
| TD-STAT | ItemStatCost ID 52, zero-based row 52 | Stat | magicmindam | stat name; ID raw 52; prepared identity 52 | GDF-R11a |
| TD-STAT | ItemStatCost ID 53, zero-based row 53 | Stat | magicmaxdam | stat name; ID raw 53; prepared identity 53 | GDF-R11a |
| TD-STAT | ItemStatCost ID 54, zero-based row 54 | Stat | coldmindam | stat name; ID raw 54; prepared identity 54 | GDF-R11a |
| TD-STAT | ItemStatCost ID 55, zero-based row 55 | Stat | coldmaxdam | stat name; ID raw 55; prepared identity 55 | GDF-R11a |
| TD-STAT | ItemStatCost ID 60, zero-based row 60 | Stat | lifedrainmindam | stat name; ID raw 60; prepared identity 60 | GDF-R11a |
| TD-STAT | ItemStatCost ID 62, zero-based row 62 | Stat | manadrainmindam | stat name; ID raw 62; prepared identity 62 | GDF-R11a |
| TD-STAT | ItemStatCost ID 64, zero-based row 64 | Stat | stamdrainmindam | stat name; ID raw 64; prepared identity 64 | GDF-R11a |
| TD-STAT | ItemStatCost ID 67, zero-based row 67 | Stat | velocitypercent | stat name; ID raw 67; prepared identity 67 | GDF-R11a |
| TD-STAT | ItemStatCost ID 68, zero-based row 68 | Stat | attackrate | stat name; ID raw 68; prepared identity 68 | GDF-R11a |
| TD-STAT | ItemStatCost ID 69, zero-based row 69 | Stat | other_animrate | stat name; ID raw 69; prepared identity 69 | GDF-R11a |
| TD-STAT | ItemStatCost ID 74, zero-based row 74 | Stat | hpregen | stat name; ID raw 74; prepared identity 74 | GDF-R11a |
| TD-STAT | ItemStatCost ID 85, zero-based row 85 | Stat | item_addexperience | stat name; ID raw 85; prepared identity 85 | GDF-R11a |
| TD-STAT | ItemStatCost ID 93, zero-based row 93 | Stat | item_fasterattackrate | stat name; ID raw 93; prepared identity 93 | GDF-R11a |
| TD-STAT | ItemStatCost ID 95, zero-based row 95 | Stat | lastblockframe | stat name; ID raw 95; prepared identity 95 | GDF-R11a |
| TD-STAT | ItemStatCost ID 96, zero-based row 96 | Stat | item_fastermovevelocity | stat name; ID raw 96; prepared identity 96 | GDF-R11a |
| TD-STAT | ItemStatCost ID 99, zero-based row 99 | Stat | item_fastergethitrate | stat name; ID raw 99; prepared identity 99 | GDF-R11a |
| TD-STAT | ItemStatCost ID 100, zero-based row 100 | Stat | monster_playercount | stat name; ID raw 100; prepared identity 100 | GDF-R11a |
| TD-STAT | ItemStatCost ID 102, zero-based row 102 | Stat | item_fasterblockrate | stat name; ID raw 102; prepared identity 102 | GDF-R11a |
| TD-STAT | ItemStatCost ID 103, zero-based row 103 | Stat | skill_bypass_undead | stat name; ID raw 103; prepared identity 103 | GDF-R11a |
| TD-STAT | ItemStatCost ID 104, zero-based row 104 | Stat | skill_bypass_demons | stat name; ID raw 104; prepared identity 104 | GDF-R11a |
| TD-STAT | ItemStatCost ID 105, zero-based row 105 | Stat | item_fastercastrate | stat name; ID raw 105; prepared identity 105 | GDF-R11a |
| TD-STAT | ItemStatCost ID 106, zero-based row 106 | Stat | skill_bypass_beasts | stat name; ID raw 106; prepared identity 106 | GDF-R11a |
| TD-STAT | ItemStatCost ID 110, zero-based row 110 | Stat | item_poisonlengthresist | stat name; ID raw 110; prepared identity 110 | GDF-R11a |
| TD-STAT | ItemStatCost ID 111, zero-based row 111 | Stat | item_normaldamage | stat name; ID raw 111; prepared identity 111 | GDF-R11a |
| TD-STAT | ItemStatCost ID 115, zero-based row 115 | Stat | item_ignoretargetac | stat name; ID raw 115; prepared identity 115 | GDF-R11a |
| TD-STAT | ItemStatCost ID 116, zero-based row 116 | Stat | item_fractionaltargetac | stat name; ID raw 116; prepared identity 116 | GDF-R11a |
| TD-STAT | ItemStatCost ID 118, zero-based row 118 | Stat | item_halffreezeduration | stat name; ID raw 118; prepared identity 118 | GDF-R11a |
| TD-STAT | ItemStatCost ID 119, zero-based row 119 | Stat | item_tohit_percent | stat name; ID raw 119; prepared identity 119 | GDF-R11a |
| TD-STAT | ItemStatCost ID 121, zero-based row 121 | Stat | item_demondamage_percent | stat name; ID raw 121; prepared identity 121 | GDF-R11a |
| TD-STAT | ItemStatCost ID 131, zero-based row 131 | Stat | thorns_percent | stat name; ID raw 131; prepared identity 131 | GDF-R11a |
| TD-STAT | ItemStatCost ID 141, zero-based row 141 | Stat | item_deadlystrike | stat name; ID raw 141; prepared identity 141 | GDF-R11a |
| TD-STAT | ItemStatCost ID 142, zero-based row 142 | Stat | item_absorbfire_percent | stat name; ID raw 142; prepared identity 142 | GDF-R11a |
| TD-STAT | ItemStatCost ID 143, zero-based row 143 | Stat | item_absorbfire | stat name; ID raw 143; prepared identity 143 | GDF-R11a |
| TD-STAT | ItemStatCost ID 144, zero-based row 144 | Stat | item_absorblight_percent | stat name; ID raw 144; prepared identity 144 | GDF-R11a |
| TD-STAT | ItemStatCost ID 145, zero-based row 145 | Stat | item_absorblight | stat name; ID raw 145; prepared identity 145 | GDF-R11a |
| TD-STAT | ItemStatCost ID 146, zero-based row 146 | Stat | item_absorbmagic_percent | stat name; ID raw 146; prepared identity 146 | GDF-R11a |
| TD-STAT | ItemStatCost ID 147, zero-based row 147 | Stat | item_absorbmagic | stat name; ID raw 147; prepared identity 147 | GDF-R11a |
| TD-STAT | ItemStatCost ID 148, zero-based row 148 | Stat | item_absorbcold_percent | stat name; ID raw 148; prepared identity 148 | GDF-R11a |
| TD-STAT | ItemStatCost ID 149, zero-based row 149 | Stat | item_absorbcold | stat name; ID raw 149; prepared identity 149 | GDF-R11a |
| TD-STAT | ItemStatCost ID 153, zero-based row 153 | Stat | item_cannotbefrozen | stat name; ID raw 153; prepared identity 153 | GDF-R11a |
| TD-STAT | ItemStatCost ID 164, zero-based row 164 | Stat | skill_concentration | stat name; ID raw 164; prepared identity 164 | GDF-R11a |
| TD-STAT | ItemStatCost ID 171, zero-based row 171 | Stat | skill_armor_percent | stat name; ID raw 171; prepared identity 171 | GDF-R11a |
| TD-STAT | ItemStatCost ID 172, zero-based row 172 | Stat | alignment | stat name; ID raw 172; prepared identity 172 | GDF-R11a |
| TD-STAT | ItemStatCost ID 179, zero-based row 179 | Stat | attack_vs_montype | stat name; ID raw 179; prepared identity 179 | GDF-R11a |
| TD-STAT | ItemStatCost ID 182, zero-based row 182 | Stat | armor_override_percent | stat name; ID raw 182; prepared identity 182 | GDF-R11a |
| TD-STAT | ItemStatCost ID 329, zero-based row 329 | Stat | passive_fire_mastery | stat name; ID raw 329; prepared identity 329 | GDF-R11a |
| TD-STAT | ItemStatCost ID 330, zero-based row 330 | Stat | passive_ltng_mastery | stat name; ID raw 330; prepared identity 330 | GDF-R11a |
| TD-STAT | ItemStatCost ID 331, zero-based row 331 | Stat | passive_cold_mastery | stat name; ID raw 331; prepared identity 331 | GDF-R11a |
| TD-STAT | ItemStatCost ID 333, zero-based row 333 | Stat | passive_fire_pierce | stat name; ID raw 333; prepared identity 333 | GDF-R11a |
| TD-STAT | ItemStatCost ID 334, zero-based row 334 | Stat | passive_ltng_pierce | stat name; ID raw 334; prepared identity 334 | GDF-R11a |
| TD-STAT | ItemStatCost ID 335, zero-based row 335 | Stat | passive_cold_pierce | stat name; ID raw 335; prepared identity 335 | GDF-R11a |
| TD-STAT | ItemStatCost ID 336, zero-based row 336 | Stat | passive_pois_pierce | stat name; ID raw 336; prepared identity 336 | GDF-R11a |
| TD-STAT | ItemStatCost ID 337, zero-based row 337 | Stat | passive_critical_strike | stat name; ID raw 337; prepared identity 337 | GDF-R11a |
| TD-STAT | ItemStatCost ID 338, zero-based row 338 | Stat | passive_dodge | stat name; ID raw 338; prepared identity 338 | GDF-R11a |
| TD-STAT | ItemStatCost ID 339, zero-based row 339 | Stat | passive_avoid | stat name; ID raw 339; prepared identity 339 | GDF-R11a |
| TD-STAT | ItemStatCost ID 340, zero-based row 340 | Stat | passive_evade | stat name; ID raw 340; prepared identity 340 | GDF-R11a |
| TD-STAT | ItemStatCost ID 348, zero-based row 348 | Stat | passive_weaponblock | stat name; ID raw 348; prepared identity 348 | GDF-R11a |
| TD-STAT | ItemStatCost ID 352, zero-based row 352 | Stat | last_sent_hp_pct | stat name; ID raw 352; prepared identity 352 | GDF-R11a |
| TD-STAT | ItemStatCost ID 357, zero-based row 357 | Stat | passive_mag_mastery | stat name; ID raw 357; prepared identity 357 | GDF-R11a |
| TD-STATE | States id 1, zero-based row 1 | state | freeze | state name; id raw 1; prepared identity 1 | GDF-R12 |
| TD-STATE | States id 2, zero-based row 2 | state | poison | state name; id raw 2; prepared identity 2 | GDF-R12 |
| TD-STATE | States id 11, zero-based row 11 | state | cold | state name; id raw 11; prepared identity 11 | GDF-R12 |
| TD-STATE | States id 12, zero-based row 12 | state | inferno | state name; id raw 12; prepared identity 12 | GDF-R12 |
| TD-STATE | States id 15, zero-based row 15 | state | concentrate | state name; id raw 15; prepared identity 15 | GDF-R12 |
| TD-STATE | States id 21, zero-based row 21 | state | stunned | state name; id raw 21; prepared identity 21 | GDF-R12 |
| TD-STATE | States id 42, zero-based row 42 | state | concentration | state name; id raw 42; prepared identity 42 | GDF-R12 |
| TD-STATE | States id 47, zero-based row 47 | state | sanctuary | state name; id raw 47; prepared identity 47 | GDF-R12 |
| TD-STATE | States id 54, zero-based row 54 | state | uninterruptable | state name; id raw 54; prepared identity 54 | GDF-R12 |
| TD-STATE | States id 65, zero-based row 65 | state | dodge | state name; id raw 65; prepared identity 65 | GDF-R12 |
| TD-STATE | States id 66, zero-based row 66 | state | avoid | state name; id raw 66; prepared identity 66 | GDF-R12 |
| TD-STATE | States id 68, zero-based row 68 | state | evade | state name; id raw 68; prepared identity 68 | GDF-R12 |
| TD-STATE | States id 86, zero-based row 86 | state | justhit | state name; id raw 86; prepared identity 86 | GDF-R12 |
| TD-STATE | States id 101, zero-based row 101 | state | holyshield | state name; id raw 101; prepared identity 101 | GDF-R12 |
| TD-STATE | States id 104, zero-based row 104 | state | corpse_nodraw | state name; id raw 104; prepared identity 104 | GDF-R12 |
| TD-STATE | States id 105, zero-based row 105 | state | alignment | state name; id raw 105; prepared identity 105 | GDF-R12 |
| TD-STATE | States id 107, zero-based row 107 | state | shatter | state name; id raw 107; prepared identity 107 | GDF-R12 |
| TD-STATE | States id 115, zero-based row 115 | state | burning | state name; id raw 115; prepared identity 115 | GDF-R12 |
| TD-STATE | States id 131, zero-based row 131 | state | shrine_resist_fire | state name; id raw 131; prepared identity 131 | GDF-R12 |
| TD-STATE | States id 133, zero-based row 133 | state | shrine_resist_poison | state name; id raw 133; prepared identity 133 | GDF-R12 |
| TD-STATE | States id 146, zero-based row 146 | state | invis | state name; id raw 146; prepared identity 146 | GDF-R12 |
| TD-DIFFICULTY | DifficultyLevels Normal, row 0 | MonsterColdDivisor | 1 | decimal integer, compiled word agrees | GDF-R02 / CSA consumers (CSA-R05, CSA-R08) |
| TD-DIFFICULTY | DifficultyLevels Normal, row 0 | MonsterFreezeDivisor | 1 | decimal integer, compiled word agrees | GDF-R02 / CSA consumers (CSA-R05, CSA-R08) |
| TD-DIFFICULTY | DifficultyLevels Nightmare, row 1 | MonsterColdDivisor | 2 | decimal integer, compiled word agrees | GDF-R02 / CSA consumers (CSA-R05, CSA-R08) |
| TD-DIFFICULTY | DifficultyLevels Nightmare, row 1 | MonsterFreezeDivisor | 2 | decimal integer, compiled word agrees | GDF-R02 / CSA consumers (CSA-R05, CSA-R08) |
| TD-DIFFICULTY | DifficultyLevels Hell, row 2 | MonsterColdDivisor | 4 | decimal integer, compiled word agrees | GDF-R02 / CSA consumers (CSA-R05, CSA-R08) |
| TD-DIFFICULTY | DifficultyLevels Hell, row 2 | MonsterFreezeDivisor | 4 | decimal integer, compiled word agrees | GDF-R02 / CSA consumers (CSA-R05, CSA-R08) |
| TD-DIFFICULTY | DifficultyLevels Hell, row 2 | ResistPenalty | -100 | decimal i32, compiled word agrees | GDF-V01 |
| TD-DIFFICULTY | DifficultyLevels Hell, row 2 | UberCodeOddsNormal | 20 | decimal i32, compiled word agrees | GDF-V01 |
| TD-DIFFICULTY | DifficultyLevels Hell, row 2 | UltraCodeOddsNormal | 30 | decimal i32, compiled word agrees | GDF-V01 |
| TD-DIFFICULTY | DifficultyLevels Hell, row 2 | LifeStealDivisor | 3 | decimal i32, compiled word agrees | GDF-V01 |
| TD-DIFFICULTY | DifficultyLevels Normal, row 0 | ManaStealDivisor | 1 | decimal i32, compiled word agrees | GDF-V01 |
| TD-DIFFICULTY | DifficultyLevels Hell, row 2 | HireableBossDamagePercent | 25 | decimal i32, compiled word agrees | GDF-V01 |
| TD-CLASS | CharStats Amazon, file data row 0; filtered class index 0 | class | Amazon | original class order, compiled string agrees | GDF-R03b |
| TD-CLASS | CharStats Sorceress, file data row 1; filtered class index 1 | class | Sorceress | original class order, compiled string agrees | GDF-R03b |
| TD-CLASS | CharStats Necromancer, file data row 2; filtered class index 2 | class | Necromancer | original class order, compiled string agrees | GDF-R03b |
| TD-CLASS | CharStats Paladin, file data row 3; filtered class index 3 | class | Paladin | original class order, compiled string agrees | GDF-R03b |
| TD-CLASS | CharStats Barbarian, file data row 4; filtered class index 4 | class | Barbarian | original class order, compiled string agrees | GDF-R03b |
| TD-CLASS | CharStats Druid, file data row 6; filtered class index 5 | class | Druid | original class order, compiled string agrees | GDF-R03b |
| TD-CLASS | CharStats Assassin, file data row 7; filtered class index 6 | class | Assassin | original class order, compiled string agrees | GDF-R03b |
| TD-CLASS | CharStats Sorceress, file row 1/class 1 | BlockFactor | 20 | decimal, compiled value agrees | GDF-V12 |
| TD-CLASS | CharStats Sorceress, file row 1/class 1 | ManaRegen | 120 | decimal, compiled value agrees | GDF-V12 |
| TD-CLASS | CharStats Sorceress, file row 1/class 1 | ToHitFactor | -15 | decimal, compiled value agrees | GDF-V12 |
| TD-PLAYER-MODE | plrmode zero-based data row 0 | Name | Death | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 0 | Token | DT | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 0 | Code | DT | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 1 | Name | Neutral | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 1 | Token | NU | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 1 | Code | NU | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 2 | Name | Walk | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 2 | Token | WL | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 2 | Code | WL | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 3 | Name | Run | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 3 | Token | RN | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 3 | Code | RN | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 4 | Name | Get Hit | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 4 | Token | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 4 | Code | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 5 | Name | Town Neutral | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 5 | Token | TN | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 5 | Code | TN | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 6 | Name | Town Walk | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 6 | Token | TW | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 6 | Code | TW | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 7 | Name | Attack1 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 7 | Token | A1 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 7 | Code | A1 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 8 | Name | Attack2 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 8 | Token | A2 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 8 | Code | A2 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 9 | Name | Block | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 9 | Token | BL | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 9 | Code | BL | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 10 | Name | Cast | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 10 | Token | SC | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 10 | Code | SC | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 11 | Name | Throw | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 11 | Token | TH | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 11 | Code | TH | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 12 | Name | Kick | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 12 | Token | KK | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 12 | Code | KK | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 13 | Name | Skill1 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 13 | Token | S1 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 13 | Code | S1 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 14 | Name | Skill2 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 14 | Token | S2 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 14 | Code | S2 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 15 | Name | Skill3 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 15 | Token | S3 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 15 | Code | S3 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 16 | Name | Skill4 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 16 | Token | S4 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 16 | Code | S4 | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 17 | Name | Dead | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 17 | Token | DD | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 17 | Code | DD | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 18 | Name | Sequence | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 18 | Token | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 18 | Code | SQ | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 19 | Name | Knock back | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 19 | Token | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-PLAYER-MODE | plrmode zero-based data row 19 | Code | KB | text identity/code; Name+Token prepared correspondence separately | GDF-R14a |
| TD-MONSTER-MODE | monmode zero-based data row 0 | name | death | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 0 | token | DT | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 0 | code | DT | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 1 | name | neutral | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 1 | token | NU | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 1 | code | NU | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 2 | name | walk | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 2 | token | WL | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 2 | code | WL | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 3 | name | gethit | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 3 | token | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 3 | code | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 4 | name | `attack1` | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 4 | token | A1 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 4 | code | A1 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 5 | name | `attack2` | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 5 | token | A2 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 5 | code | A2 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 6 | name | block | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 6 | token | BL | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 6 | code | BL | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 7 | name | cast | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 7 | token | SC | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 7 | code | SC | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 8 | name | `skill1` | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 8 | token | S1 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 8 | code | S1 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 9 | name | `skill2` | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 9 | token | S2 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 9 | code | S2 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 10 | name | `skill3` | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 10 | token | S3 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 10 | code | S3 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 11 | name | `skill4` | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 11 | token | S4 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 11 | code | S4 | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 12 | name | dead | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 12 | token | DD | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 12 | code | DD | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 13 | name | knockback | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 13 | token | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 13 | code | KB | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 14 | name | sequence | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 14 | token | xx | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 14 | code | xx | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 15 | name | run | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 15 | token | RN | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-MONSTER-MODE | monmode zero-based data row 15 | code | RN | text identity/code; Name+Token prepared correspondence separately | GDF-R15a |
| TD-ITEMTYPE | ItemTypes Gold, zero-based row 4 | ItemType | Gold | original index/name; Code prepared correspondence agrees | GDF-R16a |
| TD-ITEMTYPE | ItemTypes Gold, zero-based row 4 | Code | gold | original index/name; Code prepared correspondence agrees | GDF-R16a |
| TD-ITEMTYPE | ItemTypes Any Shield, zero-based row 51 | ItemType | Any Shield | original index/name; Code prepared correspondence agrees | GDF-R16a |
| TD-ITEMTYPE | ItemTypes Any Shield, zero-based row 51 | Code | shld | original index/name; Code prepared correspondence agrees | GDF-R16a |
| TD-MONSTER-SKILL | MonStats `sandleaper1`, file data row 78 | hcIdx | 78 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats `vulture1`, file data row 110 | hcIdx | 110 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats `willowisp1`, file data row 118 | hcIdx | 118 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats `batdemon1`, file data row 136 | hcIdx | 136 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats firebeast, file data row 230 | hcIdx | 230 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats iceglobe, file data row 231 | hcIdx | 231 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats `frogdemon1`, file data row 247 | hcIdx | 247 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats roguehire, file data row 271 | hcIdx | 271 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats act2hire, file data row 338 | hcIdx | 338 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats act3hire, file data row 359 | hcIdx | 359 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats `trappedsoul1`, file data row 403 | hcIdx | 403 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats act5hire1, file data row 561 | hcIdx | 560 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats act5hire2, file data row 562 | hcIdx | 561 | decimal class index; prepared Id agrees for this cited row | GDF-R17a |
| TD-MONSTER-SKILL | MonStats `sandleaper1`, hcIdx 78 | Velocity | 10 | prepared values for cited rows corroborate original named fields | GDF-V02 / GDF-V17 |
| TD-MONSTER-SKILL | MonStats `sandleaper1`, hcIdx 78 | Run | 10 | prepared values for cited rows corroborate original named fields | GDF-V02 / GDF-V17 |
| TD-MONSTER-SKILL | MonStats `sandleaper1`, hcIdx 78 | MonStatsEx | `sandleaper1` | prepared values for cited rows corroborate original named fields | GDF-V02 / GDF-V17 |
| TD-MONSTER-SKILL | MonStats `sandleaper1`, hcIdx 78 | inTown | (empty) | prepared values for cited rows corroborate original named fields | GDF-V02 / GDF-V17 |
| TD-MONSTER-SKILL | MonStats `sandleaper1`, hcIdx 78 | NoShldBlock | (empty) | prepared values for cited rows corroborate original named fields | GDF-V02 / GDF-V17 |
| TD-MONSTER-SKILL | MonStats act5hire1, hcIdx 560 | Velocity | 8 | prepared values for cited rows corroborate original named fields | GDF-V02 / GDF-V17 |
| TD-MONSTER-SKILL | MonStats act5hire1, hcIdx 560 | Run | 14 | prepared values for cited rows corroborate original named fields | GDF-V02 / GDF-V17 |
| TD-MONSTER-SKILL | MonStats act5hire1, hcIdx 560 | MonStatsEx | act5hire1 | prepared values for cited rows corroborate original named fields | GDF-V02 / GDF-V17 |
| TD-MONSTER-SKILL | MonStats act5hire1, hcIdx 560 | inTown | 1 | prepared values for cited rows corroborate original named fields | GDF-V02 / GDF-V17 |
| TD-MONSTER-SKILL | MonStats act5hire1, hcIdx 560 | NoShldBlock | (empty) | prepared values for cited rows corroborate original named fields | GDF-V02 / GDF-V17 |
| TD-MONSTER-SKILL | Skills Charge, Id 107 | srvdofunc | 67 | decimal handler selector, not skill id | GDF-R17b context |
| TD-MONSTER-SKILL | Skills Whirlwind, Id 151 | srvdofunc | 76 | decimal handler selector, not skill id | GDF-R17b context |
| TD-MONSTER-SKILL | Skills SerpentCharge, Id 352 | srvdofunc | 67 | decimal handler selector, not skill id | GDF-R17b context |

Additional count records: ItemStatCost 359, States 185, Skills 357 ordinary text rows; MonStats 735 text rows including one separator → 734 ordinary; MonStats2 610 including separator → 609; CharStats 8 including separator → 7; ItemTypes 104 including separator → 103; PlrMode 20; MonMode 16; PetType 20. These are archive-derived counts, not gameplay observations.

### Global MonStats correspondence discrepancy

The text file's hcIdx sequence matches ordinary-row indices for all 734 rows. The cached .bin record Id field does not: 11 tail records differ. Examples (only named cells / interpreted fields):
- text data row 724 Id cr_lancer8, hcIdx 723; .bin physical record index 723 has Id 617.
- text data row 725 Id `overseer6`, hcIdx 724; .bin record index 724 has Id 723.
- text data row 734 Id `willowisp8`, hcIdx 733; .bin record index 733 has Id 732.
The Id name cr_lancer8 occurs twice in text, at data rows 618 and 724, hcIdx 617 and 723. This is a lead for the first-pass unique-name behavior, not proof that it caused the cached-file mismatch. Archive compilation history and live post-load mutation have not been observed. Array index, explicit hcIdx cell, linked-name identity and record Id field remain distinct. Do not change the cited gameplay class indices to fit these record words. GDF-Q10 retains global correspondence as an integration gap; all 13 cited class indices corroborated normally.


## Historical live-memory correspondence (2026-10-12/13)

These were read-only observations on the reported original 1.13c session, with D2Common at its preferred base. Full live DLL hashes were not captured. No getter/flag-test/preparation execution was traced, no gameplay action or memory write occurred, and unrelated traces were left untouched. The old broad runtime labels were narrowed to actual correspondence in the current contract.

| Date | Vector/context | Location and coverage | Recorded result | Limit |
|---|---|---|---|---|
| 2026-10-12 | GDF-V01 | D2Common difficulty pointer RVA 0xA0B54; three records, +0x00..+0x2C | Selected fields agreed with named cells | +0x30..+0x54 unchecked then |
| 2026-10-13 | GDF-V01 | Same pointer/count; 264 bytes, three stride-0x58 records | All 66 dwords agreed with same-day DifficultyLevels text, 586 bytes | No preparation execution |
| 2026-10-13 | GDF-V02/V08 | Follow DataTables export; monster base +0xA78/count +0xA80; index 78 first 64 bytes | Id/BaseId/MonStatsEx 78; flags 0x02018003; Velocity/Run 10/10 | One record; no getter/flag-call trace |
| 2026-10-13 | GDF-V02/V17 | Same array, index 560 first 64 bytes | Id/BaseId/MonStatsEx 560; flags 0x0205840B; Velocity/Run 8/14 | One record; an initial address-addition error was discarded |
| 2026-10-13 | GDF-V12 | DataTables class base +0xBC4/count +0xBC8; 1372 bytes, seven records | class strings Amazon/Sorceress/Necromancer/Paladin/Barbarian/Druid/Assassin; Sorceress ManaRegen 120, ToHitFactor -15, WalkVelocity 6, RunVelocity 9, BlockFactor 20 | File cells were not read in this memory step; later archive evidence separate |
| 2026-10-13 | GDF-R19a/b | Count slots in 672-byte window starting DataTables +0xA78 plus state-count window +0xC4 | MonStats 734, MonStats2 609, Skills 357, CharStats 7, ItemStatCost 359, States 185, PetType-slot count 20 | Counts prove neither ordering nor writer execution |
| 2026-10-13 | GDF-Q01/O02 | Monster records 78/560, words +0x36/+0x38 | 78: 224/112 with Velocity/Run 10/10; 560: 256/128 with Velocity/Run 8/14 | Derived writer identified later; complete production unresolved |

The 2026-10-13 archive work read complete DifficultyLevels/PlrMode/MonMode files through the running loader, States with insufficient preview for names, and sizes for larger files. Names/identities were also compared with same-installation files extracted on 2026-10-12; matching sizes and limited byte comparisons were recorded. Physical archive identity and effective-file hashes were not retained as a complete profile. Full-coverage effective reads of 2026-10-14 above supply the current public named-cell record; historical size equality is not full equivalence.

<a id="observation-procedures"></a>
## Observation procedures

These procedures remain pending. Record matching executable/module identity, data profile, semantic inputs, measurement timing and actual coverage. None was executed for this provenance publication.

- **GDF-O01 — mode changes.** Single-player Normal, any class level 1 or above, with an established knockback source (e.g. knockback weapon or Barbarian Bash). Stand adjacent to a suitable melee target, apply knockback, measure its mode in the first animation frames. Separately arrange a documented hostile knockback source against the player and measure player mode. Expected identities: monster 13, player 19 during knockback. Original Unit mode is +0x10; sample before neutral/walk resumes. Record successful knockback and source eligibility. The old specific hostile-monster suggestion was not validated; setup must be established by the owning knockback contract. Mode-table rows do not establish execution.
- **GDF-O02 — derived animation.** Follow D2Common RVA 0x99E1C to DataTables, base +0xA78, class index times 0x1A8; compare words +0x36/+0x38 with Velocity/Run cells for at least five distinct classes. Existing cases are above. To attribute production, observe D2Common writer 0x5A8A0 during initialization with actual inputs/output stores; selectors 0x43C30/0x41DD0 remain unresolved producer dependencies. Record matches alone do not settle GDF-Q01.
- **GDF-O03 — PetType count.** Start a fresh original-profile process; observe root 0x66160 reaching D2Common loader 0x5F310. Measure decoded count, count-guard outcome and published count RVA 0xA0AC8 / DataTables +0xBF0. Original 20 has archive/memory correspondence; the producer has not been observed executing. Do not infer ownership from PlrMode's equal count.

## Gap attempts and retained limits

The historical 2026-10-14 pass inspected compiler/import paths, image listings, missing-field switch, name-link searches, descriptor callback, monster initialization/writer, treasure generation/allocation and localized-name writer. It settled the selected ASCII binding path, state/stat/handler identities, PetType ownership and mode stores described above. It did not close indirect conversion/default behavior (GDF-Q06/Q12), animation integration (GDF-Q01), treasure probability/counter/global correspondence (GDF-Q10), localization (GDF-Q13) or gameplay observation (GDF-Q11).

Earlier 2026-10-13 searches inspected candidate word stores, stride/field combinations, count stores and the player-mode loader, rejecting hits in unrelated item/path records. They did not complete every search rung or identify animation/count writers then. Later writer identification supersedes that uncertainty without proving complete production. No absence claim or new search campaign is made here. Supplied-record behavior remains distinct from producer integration. The linked gap file identifies blocked assertions.

## Historical corrections, vector preservation and review

The 2026-10-14 semantic migration kept the dictionary boundary, 20 rule families and numeric results while splitting mixed evidence labels. Original binary-layout fixtures and earlier whole-spec records are preserved privately. The current GDF-V01..V22 use semantic fields and do not require original addresses/strides/packed keys. Correspondence flags were narrowed to measured assertion scope; supplied-input vectors did not become gameplay-observed.

Original vector results retained: Difficulty -100/20/30/4/2/3/1/25; town/shield 4/0/0/0; stat/state names 99/171/54/146; mode name/token/code distinctions; treasure Picks/NoDrop and quality values; skill parameter 35/guarded zeros; handler 67/76 versus 68; record range failures; mode masks 32/8/0; descriptor counts 22/8; stat 171/layer 0; Sorceress 20/120/-15; identity-only Experience/States; missile 64/128/0/0; Gold/gold and Any Shield/shld; class 560/link/Velocity/Run; interrupt 128/0. Exact inputs, operations and results remain in the spec.

Historical reader corrections added discriminators for conversion-kind termination, reordered headers, independent second text decode, index/level/overflow guards, byte limits and four-code truncation, missing-resource aborts, unsigned BlockFactor 200, asymmetric mode flags, in-range absent-record guards, low16 group wrapping, PetType 255/256, reserved-word copy 17/29 and decimal overflow. These are synthetic inputs, not new table-cell claims. This publication did not rerun fixtures or game reading.

Historical review on 2026-10-14: an early reader reported three required discrimination findings; a later pass reported no GDF discrimination finding while forwarding foundation findings. Decimal overflow follow-up required another reader. Dictionary pass 3 reported four required findings; pass 4 reported one; pass 5 reported zero required/advisory findings. Final applicable pass 5: MISMATCH 0, GUESS 0, GAP 0, GATED 7, COVERAGE 0 required / 0 advisory; current GDF-V01..V22 assertions checked. It met the then-applicable reader stopping rule while retaining gates/draft status. This is historical review evidence, not approval or gameplay confirmation.

The historical combined documentation audit checked fixture/provenance preservation, main-contract location separation, local destinations/anchors and fence/whitespace consistency. Exact internal commands/session details remain private. That migration performed no implementation, plan, game action, trace, memory write, commit or push. This provenance publication changes evidence presentation/references only; unchanged contract labels and gates remain authoritative.

## Public provenance checks (2026-10-14)

Exact before/after comparison preserved the entire spec before Appendix A, all 22 vector
definitions, labels and gates, and the complete gap file. All 277 named-cell record rows
are identical to the pre-edit evidence. Complete original specs, companions and gap files
were preserved privately and checked against their pre-edit snapshot; archive locations
and SHA-256 values are recorded in the private audit.

The five manifest binary hashes/sizes matched two stored original bundles. Read-only PE
checks mapped 85 function/data locations to original image sections, including every one
of the 81 artifact/RVA table rows and four additional jump-table/import-slot locations.
All 17 cited export targets matched. These checks establish artifact navigation and
identity, not new behavioral verification or complete compiler execution.

Combined checks for both migrated specs, companions and the shared manifest resolved all
69 local links/anchors, including 17 inbound companion links. Public citation scans found
no private source paths, reconstruction filenames, commits, member identifiers or tool calls.
A focused cross-review found no remaining required provenance/contract issue. This pass ran
no implementation tests, vector recomputation, new table reads or gameplay actions. Existing
labels, gaps, pending observations and draft status remain unchanged; review does not imply
user approval or runtime parity.

<a id="named-cell-corroboration-2026-10-15"></a>
## Named-cell corroboration (2026-10-15)

The answering source reported the LoD 1.13c launch profile, selected separately from four reported modified profiles. Its vanilla-profile identification reuses the established historical record; refreshing the version catalog failed. A fresh complete override inventory, loaded DLL fingerprints and physical supplying MPQs were not captured. The records report a prior file override disabled; this does not establish a complete current inventory. The source manifest identifies analysis artifacts, not the current loaded DLL set. Existing hooks were untouched; no gameplay action or memory write occurred.

These coordinated read-only collections used the running game’s Storm loader. Offset 0 with requested length 0 requested each complete effective file. Decoded payload lengths equaled returned and total bytes. Columns were parsed by their actual header names; row indices exclude the header and start at zero. Each cited explicit ID/id uniquely selects the stated row and equals its row index. The hashes identify these effective text files; they do not establish physical archive origin, original-profile overrides or equality with the 2026-10-14 text/compiled files. No new compiled or live prepared-record correspondence was read.

| Effective virtual file | Offset / requested length | Returned / total bytes | SHA-256 | Coverage |
|---|---|---|---|---|
| `data/global/excel/states.txt` | 0 / 0 (whole file) | 20920 / 20920 | `898806c2b69e55e316ecfed48dde95aa2933829b94fdad234bdabc27bdbc23e1` | complete header and file; named cells below only |
| `data/global/excel/itemstatcost.txt` | 0 / 0 (whole file) | 40109 / 40109 | `ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375` | complete header and file; named cells below only |
| `data/global/excel/itemtypes.txt` | 0 / 0 (whole file) | 8775 / 8775 | `d399dc5d0f3a3e1a4f1eb611dabef8311b7025392c416f58dc26baff93af3550` | complete header and file; named cells below only |

All cells below are raw text. Decimal ID/id cells convert to the stated integer identity; Stat/state/ItemType/Code strings are retained as written. Each record is verified in 1.13c data within the reported-profile limits above. It corroborates the named mapping only; selector meaning, code-fixed uses, custom remapping, compiled preparation and gameplay execution are separate claims.

| File | Row identity / zero-based index | Column | Raw cell | Interpretation / supports |
|---|---|---|---|---|
| `states.txt` | id 1; row 1 | `id` | `1` | state identity/name; GDF-R12 |
| `states.txt` | id 1; row 1 | `state` | `freeze` | state identity/name; GDF-R12 |
| `states.txt` | id 15; row 15 | `id` | `15` | state identity/name; GDF-R12 |
| `states.txt` | id 15; row 15 | `state` | `concentrate` | state identity/name; GDF-R12 |
| `states.txt` | id 21; row 21 | `id` | `21` | state identity/name; GDF-R12 |
| `states.txt` | id 21; row 21 | `state` | `stunned` | state identity/name; GDF-R12 |
| `states.txt` | id 42; row 42 | `id` | `42` | state identity/name; GDF-R12 |
| `states.txt` | id 42; row 42 | `state` | `concentration` | state identity/name; GDF-R12 |
| `states.txt` | id 54; row 54 | `id` | `54` | state identity/name; GDF-R12 |
| `states.txt` | id 54; row 54 | `state` | `uninterruptable` | state identity/name; GDF-R12 |
| `states.txt` | id 65; row 65 | `id` | `65` | state identity/name; GDF-R12 |
| `states.txt` | id 65; row 65 | `state` | `dodge` | state identity/name; GDF-R12 |
| `states.txt` | id 66; row 66 | `id` | `66` | state identity/name; GDF-R12 |
| `states.txt` | id 66; row 66 | `state` | `avoid` | state identity/name; GDF-R12 |
| `states.txt` | id 68; row 68 | `id` | `68` | state identity/name; GDF-R12 |
| `states.txt` | id 68; row 68 | `state` | `evade` | state identity/name; GDF-R12 |
| `states.txt` | id 92; row 92 | `id` | `92` | state identity/name; GDF-R12 |
| `states.txt` | id 92; row 92 | `state` | `death_delay` | state identity/name; GDF-R12 |
| `states.txt` | id 101; row 101 | `id` | `101` | state identity/name; GDF-R12 |
| `states.txt` | id 101; row 101 | `state` | `holyshield` | state identity/name; GDF-R12 |
| `itemstatcost.txt` | ID 7; row 7 | `ID` | `7` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 7; row 7 | `Stat` | `maxhp` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 93; row 93 | `ID` | `93` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 93; row 93 | `Stat` | `item_fasterattackrate` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 95; row 95 | `ID` | `95` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 95; row 95 | `Stat` | `lastblockframe` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 96; row 96 | `ID` | `96` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 96; row 96 | `Stat` | `item_fastermovevelocity` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 99; row 99 | `ID` | `99` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 99; row 99 | `Stat` | `item_fastergethitrate` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 102; row 102 | `ID` | `102` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 102; row 102 | `Stat` | `item_fasterblockrate` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 105; row 105 | `ID` | `105` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 105; row 105 | `Stat` | `item_fastercastrate` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 164; row 164 | `ID` | `164` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 164; row 164 | `Stat` | `skill_concentration` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 352; row 352 | `ID` | `352` | stat identity/name; GDF-R11a |
| `itemstatcost.txt` | ID 352; row 352 | `Stat` | `last_sent_hp_pct` | stat identity/name; GDF-R11a |
| `itemtypes.txt` | Gold; row 4 | `ItemType` | `Gold` | cited original-profile item-type name/code; GDF-R16a |
| `itemtypes.txt` | Gold; row 4 | `Code` | `gold` | cited original-profile item-type name/code; GDF-R16a |
| `itemtypes.txt` | Any Shield; row 51 | `ItemType` | `Any Shield` | cited original-profile item-type name/code; GDF-R16a |
| `itemtypes.txt` | Any Shield; row 51 | `Code` | `shld` | cited original-profile item-type name/code; GDF-R16a |

The supplied GDF-V03 identities 99 and 54 are corroborated here; the 171 and 146 cell records remain the compatible 2026-10-14 evidence. State 101 corroborates GDF-R11b’s named state, while the code-fixed test still rests on its static consumer chain. ItemTypes name/code evidence does not verify link-building. The separate 32-byte animation-file header read establishes file presence only and supplies no GDF-Q01 animation-selection contract. Element-table reads likewise do not explain a selector handler or close a compiler conversion gap.

## Focused refresh and affected dependencies (2026-10-15)

All seven open gap rows were assessed against the matching source-manifest artifacts, the existing 2026-10-12/13/14 records and current foundations/expression contracts before considering new reading. Earlier source records retain their actual methods, dates and read boundaries. No old failed lookup/default or localization attempt was rerun, and no newly available reconstruction was treated as new proof. The animation selectors remain incomplete, indirect/compiler and missing-column behavior remain incomplete, global cached/text MonStats correspondence remains disputed, treasure probability/counter production remains with its gameplay owner, localized class names remain gated, and attributable execution remains pending. The new cells/hash coverage does not settle any of these questions.

The dependency review found two stale references: GDF-R19a and Acceptance had attributed code-built records to NRF-R16, whose current subrules consume this dictionary’s field/name correspondence; they now point to the foundations semantic-record sections and actual gameplay owners. Acceptance had called NRF-R01 wrapping notation although that rule owns the LCG; the supporting notation reference now targets foundations section 1. Neither change alters a field, conversion, call order, label or gate. GDF-V11 explicitly normalizes the NRF-R12a mask into its retained true/false predicates. Its supplied player/list/state-count inputs give raw state-54 mask 4194304 and absent state-101 result 0 under the existing static dependency; this is not a newly observed return.

All GDF-R/Q/V/O identities, the 22 existing vectors, existing dates, numerical expectations and seven gap boundaries remain. The main contract’s retrospective correspondence sentences were shortened and their original records remain above: GDF-V01 has 2026-10-13 compiled correspondence and 2026-10-14 corroboration; GDF-V02/V08/V12 and PetType retain their dated records. Synthetic fixture encodings and the complete pre-refresh three-file snapshot are preserved with SHA-256 privately. No implementation comparison, implementation or plan change, table-handler completion or gameplay campaign was performed.


## Focused self-review (2026-10-15)

The mechanical spec/companion/gap check reported 0 errors and 0 warnings. The separate preservation comparison checked all 36 rule definitions and their labels, all 22 executable-vector identities, and all 277 earlier named-cell records. Only the dependency description of GDF-R19a changes among normalized rule bodies; other rule bodies are unchanged apart from spacing/identifier formatting. Vector correspondence prose changes in GDF-V01/V02/V08/V12 retain the dated records above. GDF-V11 adds explicit dependency inputs and mask-to-predicate assertions while retaining its selected identities and true/false outputs.

The self-review mapped every verified behavioral rule to its existing supplied-input vector and each data-only rule to named-cell records, checked the dependency return formats and seven precise gap boundaries, and found no additional result-distinct discriminator needed for the unchanged local operations. The new text-cell records carry their own date, full-file coverage, header-based selection, narrow values, hashes and profile limits. This is a documentation check and self-review, not a fixture execution, gameplay observation, fresh independent-reader result or approval.

## Independent review spot check (2026-10-15)

An independent review read original 1.13c instructions before comparing them with the
rules: Fog `0x12A60` numeric handler `0x1316F..0x13193`, its kind dispatch bytes at
`0x13450` and jump table `0x13438` (GDF-R01b); D2Game `0x1560` (GDF-R06b), `0x1A40`
(GDF-R08b), `0x77FA0` selector compare (GDF-R17b/GDF-R18); D2Common `0x32890` (GDF-R11b),
`0x5A030` conversion block `0x5A556..0x5A5B0` (GDF-R10b), `0x5F310` (GDF-R19b), `0x5EF40`
(GDF-R20a), `0x603F0` (GDF-R02) and the class-record bounds in `0x31D20` (GDF-R03a). All
stated guards, widths, masks and orders were confirmed: decimal accumulation with a final
negation; kinds 2/3/4/6 store unconditionally while kind 5 skips an empty cell (kind 5 is
not one of the direct kinds this contract claims); availability mask `1 << (mode & 7)`;
`Param3 + (level-1)*Param4`; selectors 67/76 before mask[7]; count `> 171` before stat
171; nonzero group plus counter stored as low16 and Picks 0 replaced by 1; PetType signed
`>= 256` fatal and nonpositive published as 0; independent generation and final-form flags
with no fallback. GDF-R20a's text request substitutes `levels` for the table name
`leveldefs` (strings at D2Common RVAs `0x8F8E8`/`0x8FA90`); that table is outside this
dictionary's named collections and is reported for the owner rather than added here.
Recomputed GDF-V05, GDF-V06, GDF-V09 and GDF-V19 matched. The review also separated words
run into numbers in the compiled-correspondence bullets, count records and the MonStats
discrepancy examples above; no value changed.

## Focused blind-reader pass FULL1 (2026-10-15)

The reader examined the main contract and only the allowed table-evidence ranges. Reported counts: MISMATCH 0, GUESS 0, GAP 3, DISCRIM 3, GATED 6 (GDF-Q01/Q06/Q10/Q11/Q12/Q13), COVERAGE 4, HARD 3. This was a fresh reader result, not a gameplay observation or approval. Its findings supersede the initial self-review's conclusion that no further discriminator was needed.

All three inventory gaps were addressed: section 3 now lists the mode argument to GDF-R06b; section 5 lists the wrong-count Difficulty abort from GDF-R02 and the Experience publication/no-separate-count result from GDF-R04. No behavior, field width or gate changes.

The three discrimination findings were addressed with two distinct supplied fixtures. GDF-V08 now separates collection index 1 from stored identities 2/1 and field values 11/22, rejecting identity-based lookup in place of indexed retrieval. GDF-V09 separates monster index 1 from its linked MonStatsEx index 2, with opposite mGH flags; mode 3 must yield 8 instead of 0. GDF-V17 retains its original identity case and cites GDF-V08's index discriminator, avoiding a duplicate fixture.

Coverage suggestions were triaged individually:

- Byte-count versus Unicode-character limits: GDF-V19 now retains [0xC3] from supplied bytes [0xC3, 0xA9] at byte limit 1. This is supported by GDF-R01b's byte contract and makes no locale/encoding claim.
- Difficulty variants: GDF-V17 now supplies distinct Drain/Drain(N)/Drain(H) values 7/13/29 and asserts selection for difficulties 0/1/2 under GDF-R05a. This adds no leech or table-cell claim.
- True NoShldBlock: GDF-V02 now supplies inTown false/NoShldBlock true and asserts raw outputs 0/4 and predicates false/true under GDF-R18.
- Intermediate/final text-record zeroing: deferred at the original decoder/compiler execution boundary. GDF-V21 starts with distinct supplied decoded collections and proves load-selection/publication control flow. Their supplied fields cannot discriminate whether the original compiler zeroed allocation before conversions. Adding a fake local producer operation or silently treating supplied decoding as original compilation would cross GDF-R20b's GDF-Q06/Q12 preparation gate. The static zeroing evidence remains above; no new local zeroing-execution assertion is claimed. GDF-V21 now states this limit explicitly.

The three HARD cases required no changes: GDF-V21's independent generation/final selection and second text decode; GDF-V19's signed wrapping and low-width stores; GDF-V11's stat/state identity, raw mask and nonzero predicate. All 22 vector families, earlier cases/results, rule labels, question identities, dates and acceptance gates remain. The added inventories and executable assertions require a full second reader pass.

## Focused blind-reader pass FULL2 (2026-10-15)

The reader examined the actual main contract and only the allowed table-evidence ranges. Reported counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 9, COVERAGE 1, HARD 3. This is documentation review, not gameplay observation or approval.

The required discrimination finding concerned GDF-V05's equal Rare/Set/Unique values of 983, which could hide swapped quality copies. The original case and all existing variants remain. One supplied quality-copy fixture now gives Magic/Rare/Set/Unique distinct values 101/202/303/404 and requires those same-named destination values, with all other original inputs and outputs unchanged. Any swapped quality pair changes the result. This follows the existing GDF-R10b copy contract; it adds no original-table cell or producer claim.

The coverage suggestion noted that GDF-V01's reordered-header case gives nonzero values to only two of the 22 fields. It remains advisory: the case already discriminates header-name binding from input position, the direct conversion recipe is separately exercised, and the named per-field descriptor/data records support the remaining mapping. Twenty more cases repeating the same binding/conversion recipe would add no distinct behavioral branch or result. No additional exhaustive mapping fixtures were added.

The three HARD cases required no changes: GDF-V09's collection index to extended link to returned mask, GDF-V21's independent generation/final selection and separate second text decode, and GDF-V19's wrapping, low-width stores and raw-byte limits remain explicit.

The nine GATED outcomes cover GDF-R01c/R03c/R05b/R15b/R16b/R20b and GDF-O01/O02/O03. They retain the existing GDF-Q01/Q06/Q10/Q11/Q12/Q13 assertions and execution boundaries; this count uses a different reporting unit from FULL1's six question identities and closes no gap. All 36 rule labels, 22 vector families, earlier cases/results and seven gap rows remain. The new assertion requires FULL3, the third full campaign reader pass.

## Focused blind-reader pass FULL3 (2026-10-15)

The fresh third full reader reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0 after recomputing all 22 vector families. These were documented hand recomputations, not implementation tests or gameplay execution. No behavior or fixture corrections were required.

GATED 6 retains the six existing question boundaries:

- GDF-Q01: complete derived monster animation-rate production.
- GDF-Q06: complete name/code links and expression/callback conversions.
- GDF-Q10: treasure generation and global compiled correspondence.
- GDF-Q11: attributable gameplay/runtime observation.
- GDF-Q12: missing-column/default and complete linked-field behavior, including absent direction columns.
- GDF-Q13: localized class display names.

The reader's four COVERAGE observations were retained with these explicit dispositions:

- GDF-R20a pre-conversion zeroing is explicitly untested by the supplied decoded outcomes in GDF-V21. The existing static zeroing record is not an execution claim; no fake producer fixture was added.
- Original decoder/writer failure integration remains GDF-R20b's stated preparation boundary. GDF-V21 discriminates selected-resource absence and supplied successful outcomes; it does not establish all decoder/writer failure behavior.
- GDF-R06b's upper mode domain is sampled by mode 13, with modes 3/4 also covered. Its signed linked-index guards, byte/mask selection and result distinction are explicit. No exhaustive repetition across modes 8..15 was required or claimed.
- GDF-R15b has no nonzero supplied-direction fixture. The current vectors cover the stated zero-valued preparation case; missing direction-column defaults remain GDF-Q12. No complete original preparation claim or observation is inferred from that case.

The three HARDEST checks were GDF-R20a/GDF-V21's independent generation/final selection and publication/abort ordering; GDF-R06b/GDF-V09's collection index versus linked identity and returned mask; and GDF-R01b/GDF-V19's wrapping, narrow stores, flags, code padding and raw byte strings. Their explicit assertions needed no changes.

No unnecessary fixture expansion followed this pass. All 36 rule labels, 22 vector families and earlier cases/results, six question identities, seven gap rows, evidence labels and draft status remain. The main metadata's trailing hard-break spaces were removed solely to satisfy publication formatting. The mechanical check after this record reported 0 errors and 0 warnings. This review does not close the stated evidence gates, establish gameplay parity or imply user approval.
