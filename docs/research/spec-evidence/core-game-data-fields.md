# Game Data Fields — original-game evidence

Behavioral contract: [Game Data Fields](../../specs/core-game-data-fields.md).
Open gaps: [evidence gaps](../../plans/gaps/core-game-data-fields.md).
Artifact identities: [original game source manifest](../source-manifest.md).
Table layouts, getters, conversions and loading: [Game Data Table Records evidence](core-game-data-table-records.md).

The linked spec is the current contract and governs rule labels, vectors and gates. The 2026-10-14 provenance publication preserved historical findings without adding behavior verification or runtime confirmation. The focused 2026-10-15 refresh reuses compatible readings, clarifies consumed contracts and records narrow text-cell corroboration; it adds no gameplay confirmation. Complete investigation history and original binary-layout fixtures were preserved privately before replacement; they are not needed to follow these public references.

On 2026-10-15 the combined dictionary was split by record family; see [Restructure (2026-10-15)](#restructure-2026-10-15). Entries below cite current ids. Review records dated before the split describe the combined 36-rule spec and keep their counts.

## Source identity and method

Historical static reads of original Lord of Destruction Expansion 1.13c occurred on 2026-10-12/13, with targeted rereads on 2026-10-14. Analysts read reconstructions checked against original machine code where available, and disassembly otherwise. Loader descriptor arrays were decoded from original D2Common instructions/stack stores on 2026-10-13, resolving column strings in read-only data; two machine-code-matched loader reconstructions agreed. Selected undecoded instructions were corroborated against independent original-image listings on 2026-10-14, including the conversion engine. This did not establish every indirect conversion handler.

The manifest pins available original binaries and their correspondence with analysis images. Newly recorded fingerprints do not prove the DLL identity of the historical live process. The historical session recorded the manifest's Game.exe hash and reported the vanilla 1.13c profile. Full live DLL fingerprints, effective table-file hashes, physical supplying archives and a complete override inventory were not captured. No known overrides were reported; that is not an inventory. An executable hash alone cannot establish unmodified game data.

All function descriptions below are analyst labels, not original exported-name claims. Ordinals identify exports in the pinned image. RVAs are relative to the module base, not file offsets. Import thunks and resolved targets are distinct entries. Historical table/memory correspondence is separate from gameplay execution.

## Function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method/boundary |
|---|---|---|---|---|---|
| `LOD113C-D2COMMON` | `0x66160` | 10943 | All-table initialization root | GDF-R05b, GDF-R06b (through GDA-R01) | Dispatch read; reaches the mode-table loaders below |
| `LOD113C-D2COMMON` | `0x73330` | — | Stat-definition loader | GDA-R12, GDF-R02 | Original descriptor decode |
| `LOD113C-D2COMMON` | `0x5F6E0` | — | State-definition loader | GDF-R03, GDA-R13 | Original descriptor decode |
| `LOD113C-D2COMMON` | `0x19330` | — | Monster-mode loader | GDF-R06b | Descriptor decode and publication stores |
| `LOD113C-D2COMMON` | `0x19650` | — | Player-type and player-mode loader | GDF-R05b | Descriptor decode; combined allocation and subset |
| `LOD113C-D2COMMON` | `0x33110` | 10819 | Animation-rate state selection | GDF-R04 | Selects state 101; full rate calculation outside boundary |
| `LOD113C-D2COMMON` | `0x37DB0` | — | Extended-list state membership test | GDF-R04 | Extended-list gate and membership bits; consumes RLI-R03 |
| `LOD113C-D2COMMON` | `0x33CD0` | 10494 | Public state-test wrapper | GDF-R04, GDA-R04a | Unit types, state-count bound and membership callee |
| `LOD113C-D2COMMON` | `0x32890` | 10672 | Defense calculation stat selector | GDA-R12, GDF-R04 | Definition-count bound and stat 171/layer 0; consumes RLI-R02a/RLI-R02b |
| `LOD113C-D2COMMON` | `0x38A80` | — | Stat read dependency | GDF-R04 | Owned by RLI-R02a; not independently reverified here |
| `LOD113C-D2GAME` | `0x1420` | — | Skill record getter | GDA-R11a, GDF-R09, GDA-R04a | Signed bounds and stride |
| `LOD113C-D2GAME` | `0x77FA0` | — | Skill interruption selector | GDF-R09, GDA-R16 | Compares srvdofunc 67/76 before interrupt bit |
| `LOD113C-D2GAME` | `0xDE5A0` | — | State-54 call site | GDF-R04 | Passes state 54/unit to imported state wrapper |
| `LOD113C-D2GAME` | `0xA68E` | — | State-test import thunk | GDF-R04 | D2Common #10494 at RVA 0x33CD0 |

### Dispatch and imported paths

D2Common `0x66160` reaches Difficulty `0x603F0`, CharStats `0x60A80`, Experience `0x60920`, Missiles `0x7C8D0`, Skills/SkillDesc `0x61EC0`, ItemStatCost `0x73330`, States `0x5F6E0`, MonMode `0x19330`, player types/modes `0x19650` and PetType `0x5F310`. Monster initialization is `0x66160 → 0x5D570 → 0x5AA40 / 0x56C10`, with derived animation writer `0x5A8A0`. Treasure initialization is `0x66160 → 0x5A730`, generated classes `0x547D0`, then loader `0x5A030` and allocation `0x53780`. ItemTypes `0x72510` is anchored by named output/stride/stores; its root caller was not established historically. These are initialization paths, not per-frame dispatch.

State path: D2Game `0xDE5A0 → thunk 0xA68E → D2Common #10494 / 0x33CD0 → 0x37DB0`. The wrapper bounds the state identity by the state count and admits unit types 0, 1 and 3. Membership reads the unit stat list, extended-list flag and state-bit word/mask. RLI-R03 owns that consumed operation. D2Common `0x33110` selects state 101 before membership. D2Common `0x32890` selects fixed stat 171/layer 0 then stat reader `0x38A80`; RLI-R02a/RLI-R02b owns the read result.

### Branch and correction rationale

- D2Common `0x32890` uses count >171, definition displacement `171 * 0x144 = 0xD86C`, packed stat key `0xAB0000`: stat 171 `skill_armor_percent`, not 172 `alignment`. State 54 `uninterruptable` is a state, not a stat. `0x33110 → 0x37DB0` tests state 101 `holyshield`.
- D2Game `0x77FA0 → 0x1420` compares Skills +0x2E `srvdofunc` with 67/76 before reading interrupt. They are handler selectors, not skill ids. Byte +7 bit 7 is bit 31, distinct from leftskill bit 30. Missile GetHit/SoftHit are distinct from ClientSend.
- MonStats byte +0x0D/mask[2] means inTown bit 10, not MonStats2. Byte +0x0F/mask[2] means NoShldBlock bit 26. Mask-table displacement +8 selects dword mask[2], low byte 4. D2Game `0xDBA40` reads Drain unsigned. MonMode row 13 is `knockback`; earlier `combat-knockback` was a document-renaming artifact.

## Descriptor and consumer evidence map

These original layouts connect named fields to consumers; they do not require reproducing original memory storage. The complete descriptor map is in the [table-records evidence](core-game-data-table-records.md#descriptor-and-consumer-evidence-map); the mode-table entries owned here are:

| Collection | D2Common loader RVA | Stride | Original publication/container location | Rules |
|---|---|---|---|---|
| MonMode | 0x19330 | 0x34 | count RVA 0xA13A4; bases 0xA13A8/AC/B0 | GDF-R06b |
| Player types/modes | 0x19650 | 0x34 | count 0xA13C4; bases 0xA13C8/CC; mode subset 0xA13D0 | GDF-R05b |

MonMode prepares name +0x00 (limit 31), token +0x20, and direction bytes DT/NU/WL/GH/A1/A2/BL/SC/S1/S2/S3/S4/DD/KB/SQ/RN at +0x24..+0x33. Player modes prepare name/token. Code is not a prepared field. Neither mode text file has direction columns; missing-column behavior remains gated.

## Original data locations

| Artifact | Data RVA | Export ordinal | Analyst description | Rules |
|---|---|---|---|---|
| `LOD113C-D2COMMON` | `0xA13A4` | — | MonMode count | GDF-R06b |
| `LOD113C-D2COMMON` | `0xA13A8` | — | MonMode pointer | GDF-R06b |
| `LOD113C-D2COMMON` | `0xA13AC` | — | MonMode pointer copy | GDF-R06b |
| `LOD113C-D2COMMON` | `0xA13B0` | — | MonMode pointer copy | GDF-R06b |
| `LOD113C-D2COMMON` | `0xA13C4` | — | Combined player type/mode count | GDF-R05b |
| `LOD113C-D2COMMON` | `0xA13C8` | — | Combined player type/mode pointer | GDF-R05b |
| `LOD113C-D2COMMON` | `0xA13CC` | — | Combined player type/mode pointer copy | GDF-R05b |
| `LOD113C-D2COMMON` | `0xA13D0` | — | Player-mode subset pointer | GDF-R05b |

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
| `TD-CLASS` | `charstats.txt` |
| `TD-PLAYER-MODE` | `plrmode.txt` |
| `TD-MONSTER-MODE` | `monmode.txt` |
| `TD-ITEMTYPE` | `itemtypes.txt` |
| `TD-MONSTER-SKILL`, MonStats row | `monstats.txt` |
| `TD-MONSTER-SKILL`, Skills row | `skills.txt` |

| Record | Row identity / indexing | Column | Raw cell | Interpretation / correspondence | Supports |
|--------|-------------------------|--------|----------|----------------------------------|----------|
| TD-STAT | ItemStatCost ID 0, zero-based row 0 | Stat | strength | stat name; ID raw 0; prepared identity 0 | GDF-R02 |
| TD-STAT | ItemStatCost ID 2, zero-based row 2 | Stat | dexterity | stat name; ID raw 2; prepared identity 2 | GDF-R02 |
| TD-STAT | ItemStatCost ID 6, zero-based row 6 | Stat | hitpoints | stat name; ID raw 6; prepared identity 6 | GDF-R02 |
| TD-STAT | ItemStatCost ID 7, zero-based row 7 | Stat | maxhp | stat name; ID raw 7; prepared identity 7 | GDF-R02 |
| TD-STAT | ItemStatCost ID 8, zero-based row 8 | Stat | mana | stat name; ID raw 8; prepared identity 8 | GDF-R02 |
| TD-STAT | ItemStatCost ID 10, zero-based row 10 | Stat | stamina | stat name; ID raw 10; prepared identity 10 | GDF-R02 |
| TD-STAT | ItemStatCost ID 12, zero-based row 12 | Stat | level | stat name; ID raw 12; prepared identity 12 | GDF-R02 |
| TD-STAT | ItemStatCost ID 13, zero-based row 13 | Stat | experience | stat name; ID raw 13; prepared identity 13 | GDF-R02 |
| TD-STAT | ItemStatCost ID 14, zero-based row 14 | Stat | gold | stat name; ID raw 14; prepared identity 14 | GDF-R02 |
| TD-STAT | ItemStatCost ID 16, zero-based row 16 | Stat | item_armor_percent | stat name; ID raw 16; prepared identity 16 | GDF-R02 |
| TD-STAT | ItemStatCost ID 19, zero-based row 19 | Stat | tohit | stat name; ID raw 19; prepared identity 19 | GDF-R02 |
| TD-STAT | ItemStatCost ID 20, zero-based row 20 | Stat | toblock | stat name; ID raw 20; prepared identity 20 | GDF-R02 |
| TD-STAT | ItemStatCost ID 21, zero-based row 21 | Stat | mindamage | stat name; ID raw 21; prepared identity 21 | GDF-R02 |
| TD-STAT | ItemStatCost ID 22, zero-based row 22 | Stat | maxdamage | stat name; ID raw 22; prepared identity 22 | GDF-R02 |
| TD-STAT | ItemStatCost ID 23, zero-based row 23 | Stat | secondary_mindamage | stat name; ID raw 23; prepared identity 23 | GDF-R02 |
| TD-STAT | ItemStatCost ID 24, zero-based row 24 | Stat | secondary_maxdamage | stat name; ID raw 24; prepared identity 24 | GDF-R02 |
| TD-STAT | ItemStatCost ID 29, zero-based row 29 | Stat | lastexp | stat name; ID raw 29; prepared identity 29 | GDF-R02 |
| TD-STAT | ItemStatCost ID 31, zero-based row 31 | Stat | armorclass | stat name; ID raw 31; prepared identity 31 | GDF-R02 |
| TD-STAT | ItemStatCost ID 32, zero-based row 32 | Stat | armorclass_vs_missile | stat name; ID raw 32; prepared identity 32 | GDF-R02 |
| TD-STAT | ItemStatCost ID 33, zero-based row 33 | Stat | armorclass_vs_hth | stat name; ID raw 33; prepared identity 33 | GDF-R02 |
| TD-STAT | ItemStatCost ID 34, zero-based row 34 | Stat | normal_damage_reduction | stat name; ID raw 34; prepared identity 34 | GDF-R02 |
| TD-STAT | ItemStatCost ID 35, zero-based row 35 | Stat | magic_damage_reduction | stat name; ID raw 35; prepared identity 35 | GDF-R02 |
| TD-STAT | ItemStatCost ID 36, zero-based row 36 | Stat | damageresist | stat name; ID raw 36; prepared identity 36 | GDF-R02 |
| TD-STAT | ItemStatCost ID 37, zero-based row 37 | Stat | magicresist | stat name; ID raw 37; prepared identity 37 | GDF-R02 |
| TD-STAT | ItemStatCost ID 38, zero-based row 38 | Stat | maxmagicresist | stat name; ID raw 38; prepared identity 38 | GDF-R02 |
| TD-STAT | ItemStatCost ID 39, zero-based row 39 | Stat | fireresist | stat name; ID raw 39; prepared identity 39 | GDF-R02 |
| TD-STAT | ItemStatCost ID 40, zero-based row 40 | Stat | maxfireresist | stat name; ID raw 40; prepared identity 40 | GDF-R02 |
| TD-STAT | ItemStatCost ID 41, zero-based row 41 | Stat | lightresist | stat name; ID raw 41; prepared identity 41 | GDF-R02 |
| TD-STAT | ItemStatCost ID 42, zero-based row 42 | Stat | maxlightresist | stat name; ID raw 42; prepared identity 42 | GDF-R02 |
| TD-STAT | ItemStatCost ID 43, zero-based row 43 | Stat | coldresist | stat name; ID raw 43; prepared identity 43 | GDF-R02 |
| TD-STAT | ItemStatCost ID 44, zero-based row 44 | Stat | maxcoldresist | stat name; ID raw 44; prepared identity 44 | GDF-R02 |
| TD-STAT | ItemStatCost ID 45, zero-based row 45 | Stat | poisonresist | stat name; ID raw 45; prepared identity 45 | GDF-R02 |
| TD-STAT | ItemStatCost ID 46, zero-based row 46 | Stat | maxpoisonresist | stat name; ID raw 46; prepared identity 46 | GDF-R02 |
| TD-STAT | ItemStatCost ID 48, zero-based row 48 | Stat | firemindam | stat name; ID raw 48; prepared identity 48 | GDF-R02 |
| TD-STAT | ItemStatCost ID 49, zero-based row 49 | Stat | firemaxdam | stat name; ID raw 49; prepared identity 49 | GDF-R02 |
| TD-STAT | ItemStatCost ID 50, zero-based row 50 | Stat | lightmindam | stat name; ID raw 50; prepared identity 50 | GDF-R02 |
| TD-STAT | ItemStatCost ID 51, zero-based row 51 | Stat | lightmaxdam | stat name; ID raw 51; prepared identity 51 | GDF-R02 |
| TD-STAT | ItemStatCost ID 52, zero-based row 52 | Stat | magicmindam | stat name; ID raw 52; prepared identity 52 | GDF-R02 |
| TD-STAT | ItemStatCost ID 53, zero-based row 53 | Stat | magicmaxdam | stat name; ID raw 53; prepared identity 53 | GDF-R02 |
| TD-STAT | ItemStatCost ID 54, zero-based row 54 | Stat | coldmindam | stat name; ID raw 54; prepared identity 54 | GDF-R02 |
| TD-STAT | ItemStatCost ID 55, zero-based row 55 | Stat | coldmaxdam | stat name; ID raw 55; prepared identity 55 | GDF-R02 |
| TD-STAT | ItemStatCost ID 60, zero-based row 60 | Stat | lifedrainmindam | stat name; ID raw 60; prepared identity 60 | GDF-R02 |
| TD-STAT | ItemStatCost ID 62, zero-based row 62 | Stat | manadrainmindam | stat name; ID raw 62; prepared identity 62 | GDF-R02 |
| TD-STAT | ItemStatCost ID 64, zero-based row 64 | Stat | stamdrainmindam | stat name; ID raw 64; prepared identity 64 | GDF-R02 |
| TD-STAT | ItemStatCost ID 67, zero-based row 67 | Stat | velocitypercent | stat name; ID raw 67; prepared identity 67 | GDF-R02 |
| TD-STAT | ItemStatCost ID 68, zero-based row 68 | Stat | attackrate | stat name; ID raw 68; prepared identity 68 | GDF-R02 |
| TD-STAT | ItemStatCost ID 69, zero-based row 69 | Stat | other_animrate | stat name; ID raw 69; prepared identity 69 | GDF-R02 |
| TD-STAT | ItemStatCost ID 74, zero-based row 74 | Stat | hpregen | stat name; ID raw 74; prepared identity 74 | GDF-R02 |
| TD-STAT | ItemStatCost ID 85, zero-based row 85 | Stat | item_addexperience | stat name; ID raw 85; prepared identity 85 | GDF-R02 |
| TD-STAT | ItemStatCost ID 93, zero-based row 93 | Stat | item_fasterattackrate | stat name; ID raw 93; prepared identity 93 | GDF-R02 |
| TD-STAT | ItemStatCost ID 95, zero-based row 95 | Stat | lastblockframe | stat name; ID raw 95; prepared identity 95 | GDF-R02 |
| TD-STAT | ItemStatCost ID 96, zero-based row 96 | Stat | item_fastermovevelocity | stat name; ID raw 96; prepared identity 96 | GDF-R02 |
| TD-STAT | ItemStatCost ID 99, zero-based row 99 | Stat | item_fastergethitrate | stat name; ID raw 99; prepared identity 99 | GDF-R02 |
| TD-STAT | ItemStatCost ID 100, zero-based row 100 | Stat | monster_playercount | stat name; ID raw 100; prepared identity 100 | GDF-R02 |
| TD-STAT | ItemStatCost ID 102, zero-based row 102 | Stat | item_fasterblockrate | stat name; ID raw 102; prepared identity 102 | GDF-R02 |
| TD-STAT | ItemStatCost ID 103, zero-based row 103 | Stat | skill_bypass_undead | stat name; ID raw 103; prepared identity 103 | GDF-R02 |
| TD-STAT | ItemStatCost ID 104, zero-based row 104 | Stat | skill_bypass_demons | stat name; ID raw 104; prepared identity 104 | GDF-R02 |
| TD-STAT | ItemStatCost ID 105, zero-based row 105 | Stat | item_fastercastrate | stat name; ID raw 105; prepared identity 105 | GDF-R02 |
| TD-STAT | ItemStatCost ID 106, zero-based row 106 | Stat | skill_bypass_beasts | stat name; ID raw 106; prepared identity 106 | GDF-R02 |
| TD-STAT | ItemStatCost ID 110, zero-based row 110 | Stat | item_poisonlengthresist | stat name; ID raw 110; prepared identity 110 | GDF-R02 |
| TD-STAT | ItemStatCost ID 111, zero-based row 111 | Stat | item_normaldamage | stat name; ID raw 111; prepared identity 111 | GDF-R02 |
| TD-STAT | ItemStatCost ID 115, zero-based row 115 | Stat | item_ignoretargetac | stat name; ID raw 115; prepared identity 115 | GDF-R02 |
| TD-STAT | ItemStatCost ID 116, zero-based row 116 | Stat | item_fractionaltargetac | stat name; ID raw 116; prepared identity 116 | GDF-R02 |
| TD-STAT | ItemStatCost ID 118, zero-based row 118 | Stat | item_halffreezeduration | stat name; ID raw 118; prepared identity 118 | GDF-R02 |
| TD-STAT | ItemStatCost ID 119, zero-based row 119 | Stat | item_tohit_percent | stat name; ID raw 119; prepared identity 119 | GDF-R02 |
| TD-STAT | ItemStatCost ID 121, zero-based row 121 | Stat | item_demondamage_percent | stat name; ID raw 121; prepared identity 121 | GDF-R02 |
| TD-STAT | ItemStatCost ID 131, zero-based row 131 | Stat | thorns_percent | stat name; ID raw 131; prepared identity 131 | GDF-R02 |
| TD-STAT | ItemStatCost ID 141, zero-based row 141 | Stat | item_deadlystrike | stat name; ID raw 141; prepared identity 141 | GDF-R02 |
| TD-STAT | ItemStatCost ID 142, zero-based row 142 | Stat | item_absorbfire_percent | stat name; ID raw 142; prepared identity 142 | GDF-R02 |
| TD-STAT | ItemStatCost ID 143, zero-based row 143 | Stat | item_absorbfire | stat name; ID raw 143; prepared identity 143 | GDF-R02 |
| TD-STAT | ItemStatCost ID 144, zero-based row 144 | Stat | item_absorblight_percent | stat name; ID raw 144; prepared identity 144 | GDF-R02 |
| TD-STAT | ItemStatCost ID 145, zero-based row 145 | Stat | item_absorblight | stat name; ID raw 145; prepared identity 145 | GDF-R02 |
| TD-STAT | ItemStatCost ID 146, zero-based row 146 | Stat | item_absorbmagic_percent | stat name; ID raw 146; prepared identity 146 | GDF-R02 |
| TD-STAT | ItemStatCost ID 147, zero-based row 147 | Stat | item_absorbmagic | stat name; ID raw 147; prepared identity 147 | GDF-R02 |
| TD-STAT | ItemStatCost ID 148, zero-based row 148 | Stat | item_absorbcold_percent | stat name; ID raw 148; prepared identity 148 | GDF-R02 |
| TD-STAT | ItemStatCost ID 149, zero-based row 149 | Stat | item_absorbcold | stat name; ID raw 149; prepared identity 149 | GDF-R02 |
| TD-STAT | ItemStatCost ID 153, zero-based row 153 | Stat | item_cannotbefrozen | stat name; ID raw 153; prepared identity 153 | GDF-R02 |
| TD-STAT | ItemStatCost ID 164, zero-based row 164 | Stat | skill_concentration | stat name; ID raw 164; prepared identity 164 | GDF-R02 |
| TD-STAT | ItemStatCost ID 171, zero-based row 171 | Stat | skill_armor_percent | stat name; ID raw 171; prepared identity 171 | GDF-R02 |
| TD-STAT | ItemStatCost ID 172, zero-based row 172 | Stat | alignment | stat name; ID raw 172; prepared identity 172 | GDF-R02 |
| TD-STAT | ItemStatCost ID 179, zero-based row 179 | Stat | attack_vs_montype | stat name; ID raw 179; prepared identity 179 | GDF-R02 |
| TD-STAT | ItemStatCost ID 182, zero-based row 182 | Stat | armor_override_percent | stat name; ID raw 182; prepared identity 182 | GDF-R02 |
| TD-STAT | ItemStatCost ID 329, zero-based row 329 | Stat | passive_fire_mastery | stat name; ID raw 329; prepared identity 329 | GDF-R02 |
| TD-STAT | ItemStatCost ID 330, zero-based row 330 | Stat | passive_ltng_mastery | stat name; ID raw 330; prepared identity 330 | GDF-R02 |
| TD-STAT | ItemStatCost ID 331, zero-based row 331 | Stat | passive_cold_mastery | stat name; ID raw 331; prepared identity 331 | GDF-R02 |
| TD-STAT | ItemStatCost ID 333, zero-based row 333 | Stat | passive_fire_pierce | stat name; ID raw 333; prepared identity 333 | GDF-R02 |
| TD-STAT | ItemStatCost ID 334, zero-based row 334 | Stat | passive_ltng_pierce | stat name; ID raw 334; prepared identity 334 | GDF-R02 |
| TD-STAT | ItemStatCost ID 335, zero-based row 335 | Stat | passive_cold_pierce | stat name; ID raw 335; prepared identity 335 | GDF-R02 |
| TD-STAT | ItemStatCost ID 336, zero-based row 336 | Stat | passive_pois_pierce | stat name; ID raw 336; prepared identity 336 | GDF-R02 |
| TD-STAT | ItemStatCost ID 337, zero-based row 337 | Stat | passive_critical_strike | stat name; ID raw 337; prepared identity 337 | GDF-R02 |
| TD-STAT | ItemStatCost ID 338, zero-based row 338 | Stat | passive_dodge | stat name; ID raw 338; prepared identity 338 | GDF-R02 |
| TD-STAT | ItemStatCost ID 339, zero-based row 339 | Stat | passive_avoid | stat name; ID raw 339; prepared identity 339 | GDF-R02 |
| TD-STAT | ItemStatCost ID 340, zero-based row 340 | Stat | passive_evade | stat name; ID raw 340; prepared identity 340 | GDF-R02 |
| TD-STAT | ItemStatCost ID 348, zero-based row 348 | Stat | passive_weaponblock | stat name; ID raw 348; prepared identity 348 | GDF-R02 |
| TD-STAT | ItemStatCost ID 352, zero-based row 352 | Stat | last_sent_hp_pct | stat name; ID raw 352; prepared identity 352 | GDF-R02 |
| TD-STAT | ItemStatCost ID 357, zero-based row 357 | Stat | passive_mag_mastery | stat name; ID raw 357; prepared identity 357 | GDF-R02 |
| TD-STATE | States id 1, zero-based row 1 | state | freeze | state name; id raw 1; prepared identity 1 | GDF-R03 |
| TD-STATE | States id 2, zero-based row 2 | state | poison | state name; id raw 2; prepared identity 2 | GDF-R03 |
| TD-STATE | States id 11, zero-based row 11 | state | cold | state name; id raw 11; prepared identity 11 | GDF-R03 |
| TD-STATE | States id 12, zero-based row 12 | state | inferno | state name; id raw 12; prepared identity 12 | GDF-R03 |
| TD-STATE | States id 15, zero-based row 15 | state | concentrate | state name; id raw 15; prepared identity 15 | GDF-R03 |
| TD-STATE | States id 21, zero-based row 21 | state | stunned | state name; id raw 21; prepared identity 21 | GDF-R03 |
| TD-STATE | States id 42, zero-based row 42 | state | concentration | state name; id raw 42; prepared identity 42 | GDF-R03 |
| TD-STATE | States id 47, zero-based row 47 | state | sanctuary | state name; id raw 47; prepared identity 47 | GDF-R03 |
| TD-STATE | States id 54, zero-based row 54 | state | uninterruptable | state name; id raw 54; prepared identity 54 | GDF-R03 |
| TD-STATE | States id 65, zero-based row 65 | state | dodge | state name; id raw 65; prepared identity 65 | GDF-R03 |
| TD-STATE | States id 66, zero-based row 66 | state | avoid | state name; id raw 66; prepared identity 66 | GDF-R03 |
| TD-STATE | States id 68, zero-based row 68 | state | evade | state name; id raw 68; prepared identity 68 | GDF-R03 |
| TD-STATE | States id 86, zero-based row 86 | state | justhit | state name; id raw 86; prepared identity 86 | GDF-R03 |
| TD-STATE | States id 101, zero-based row 101 | state | holyshield | state name; id raw 101; prepared identity 101 | GDF-R03 |
| TD-STATE | States id 104, zero-based row 104 | state | corpse_nodraw | state name; id raw 104; prepared identity 104 | GDF-R03 |
| TD-STATE | States id 105, zero-based row 105 | state | alignment | state name; id raw 105; prepared identity 105 | GDF-R03 |
| TD-STATE | States id 107, zero-based row 107 | state | shatter | state name; id raw 107; prepared identity 107 | GDF-R03 |
| TD-STATE | States id 115, zero-based row 115 | state | burning | state name; id raw 115; prepared identity 115 | GDF-R03 |
| TD-STATE | States id 131, zero-based row 131 | state | shrine_resist_fire | state name; id raw 131; prepared identity 131 | GDF-R03 |
| TD-STATE | States id 133, zero-based row 133 | state | shrine_resist_poison | state name; id raw 133; prepared identity 133 | GDF-R03 |
| TD-STATE | States id 146, zero-based row 146 | state | invis | state name; id raw 146; prepared identity 146 | GDF-R03 |
| TD-CLASS | CharStats Amazon, file data row 0; filtered class index 0 | class | Amazon | original class order, compiled string agrees | GDF-R01 |
| TD-CLASS | CharStats Sorceress, file data row 1; filtered class index 1 | class | Sorceress | original class order, compiled string agrees | GDF-R01 |
| TD-CLASS | CharStats Necromancer, file data row 2; filtered class index 2 | class | Necromancer | original class order, compiled string agrees | GDF-R01 |
| TD-CLASS | CharStats Paladin, file data row 3; filtered class index 3 | class | Paladin | original class order, compiled string agrees | GDF-R01 |
| TD-CLASS | CharStats Barbarian, file data row 4; filtered class index 4 | class | Barbarian | original class order, compiled string agrees | GDF-R01 |
| TD-CLASS | CharStats Druid, file data row 6; filtered class index 5 | class | Druid | original class order, compiled string agrees | GDF-R01 |
| TD-CLASS | CharStats Assassin, file data row 7; filtered class index 6 | class | Assassin | original class order, compiled string agrees | GDF-R01 |
| TD-PLAYER-MODE | plrmode zero-based data row 0 | Name | Death | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 0 | Token | DT | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 0 | Code | DT | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 1 | Name | Neutral | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 1 | Token | NU | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 1 | Code | NU | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 2 | Name | Walk | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 2 | Token | WL | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 2 | Code | WL | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 3 | Name | Run | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 3 | Token | RN | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 3 | Code | RN | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 4 | Name | Get Hit | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 4 | Token | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 4 | Code | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 5 | Name | Town Neutral | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 5 | Token | TN | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 5 | Code | TN | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 6 | Name | Town Walk | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 6 | Token | TW | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 6 | Code | TW | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 7 | Name | Attack1 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 7 | Token | A1 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 7 | Code | A1 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 8 | Name | Attack2 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 8 | Token | A2 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 8 | Code | A2 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 9 | Name | Block | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 9 | Token | BL | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 9 | Code | BL | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 10 | Name | Cast | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 10 | Token | SC | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 10 | Code | SC | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 11 | Name | Throw | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 11 | Token | TH | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 11 | Code | TH | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 12 | Name | Kick | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 12 | Token | KK | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 12 | Code | KK | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 13 | Name | Skill1 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 13 | Token | S1 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 13 | Code | S1 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 14 | Name | Skill2 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 14 | Token | S2 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 14 | Code | S2 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 15 | Name | Skill3 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 15 | Token | S3 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 15 | Code | S3 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 16 | Name | Skill4 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 16 | Token | S4 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 16 | Code | S4 | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 17 | Name | Dead | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 17 | Token | DD | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 17 | Code | DD | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 18 | Name | Sequence | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 18 | Token | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 18 | Code | SQ | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 19 | Name | Knock back | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 19 | Token | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-PLAYER-MODE | plrmode zero-based data row 19 | Code | KB | text identity/code; Name+Token prepared correspondence separately | GDF-R05a |
| TD-MONSTER-MODE | monmode zero-based data row 0 | name | death | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 0 | token | DT | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 0 | code | DT | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 1 | name | neutral | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 1 | token | NU | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 1 | code | NU | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 2 | name | walk | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 2 | token | WL | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 2 | code | WL | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 3 | name | gethit | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 3 | token | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 3 | code | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 4 | name | `attack1` | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 4 | token | A1 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 4 | code | A1 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 5 | name | `attack2` | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 5 | token | A2 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 5 | code | A2 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 6 | name | block | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 6 | token | BL | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 6 | code | BL | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 7 | name | cast | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 7 | token | SC | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 7 | code | SC | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 8 | name | `skill1` | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 8 | token | S1 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 8 | code | S1 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 9 | name | `skill2` | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 9 | token | S2 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 9 | code | S2 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 10 | name | `skill3` | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 10 | token | S3 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 10 | code | S3 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 11 | name | `skill4` | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 11 | token | S4 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 11 | code | S4 | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 12 | name | dead | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 12 | token | DD | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 12 | code | DD | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 13 | name | knockback | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 13 | token | GH | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 13 | code | KB | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 14 | name | sequence | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 14 | token | xx | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 14 | code | xx | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 15 | name | run | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 15 | token | RN | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-MONSTER-MODE | monmode zero-based data row 15 | code | RN | text identity/code; Name+Token prepared correspondence separately | GDF-R06a |
| TD-ITEMTYPE | ItemTypes Gold, zero-based row 4 | ItemType | Gold | original index/name; Code prepared correspondence agrees | GDF-R07 |
| TD-ITEMTYPE | ItemTypes Gold, zero-based row 4 | Code | gold | original index/name; Code prepared correspondence agrees | GDF-R07 |
| TD-ITEMTYPE | ItemTypes Any Shield, zero-based row 51 | ItemType | Any Shield | original index/name; Code prepared correspondence agrees | GDF-R07 |
| TD-ITEMTYPE | ItemTypes Any Shield, zero-based row 51 | Code | shld | original index/name; Code prepared correspondence agrees | GDF-R07 |
| TD-MONSTER-SKILL | MonStats `sandleaper1`, file data row 78 | hcIdx | 78 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats `vulture1`, file data row 110 | hcIdx | 110 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats `willowisp1`, file data row 118 | hcIdx | 118 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats `batdemon1`, file data row 136 | hcIdx | 136 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats firebeast, file data row 230 | hcIdx | 230 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats iceglobe, file data row 231 | hcIdx | 231 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats `frogdemon1`, file data row 247 | hcIdx | 247 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats roguehire, file data row 271 | hcIdx | 271 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats act2hire, file data row 338 | hcIdx | 338 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats act3hire, file data row 359 | hcIdx | 359 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats `trappedsoul1`, file data row 403 | hcIdx | 403 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats act5hire1, file data row 561 | hcIdx | 560 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | MonStats act5hire2, file data row 562 | hcIdx | 561 | decimal class index; prepared Id agrees for this cited row | GDF-R08 |
| TD-MONSTER-SKILL | Skills Charge, Id 107 | srvdofunc | 67 | decimal handler selector, not skill id | GDF-R09 context |
| TD-MONSTER-SKILL | Skills Whirlwind, Id 151 | srvdofunc | 76 | decimal handler selector, not skill id | GDF-R09 context |
| TD-MONSTER-SKILL | Skills SerpentCharge, Id 352 | srvdofunc | 67 | decimal handler selector, not skill id | GDF-R09 context |

Additional count records: ItemStatCost 359, States 185, Skills 357 ordinary text rows; MonStats 735 text rows including one separator → 734 ordinary; MonStats2 610 including separator → 609; CharStats 8 including separator → 7; ItemTypes 104 including separator → 103; PlrMode 20; MonMode 16; PetType 20. These are archive-derived counts, not gameplay observations.

The global MonStats cached/text correspondence discrepancy behind GDA-Q03 is recorded in the [table-records evidence](core-game-data-table-records.md#global-monstats-correspondence-discrepancy); the cited GDF-R08 class indices corroborated normally.

<a id="observation-procedures"></a>
## Observation procedures

These procedures remain pending. Record matching executable/module identity, data profile, semantic inputs, measurement timing and actual coverage. None was executed for this provenance publication.

- **GDF-O01 — mode changes.** Single-player Normal, any class level 1 or above, with an established knockback source (e.g. knockback weapon or Barbarian Bash). Stand adjacent to a suitable melee target, apply knockback, measure its mode in the first animation frames. Separately arrange a documented hostile knockback source against the player and measure player mode. Expected identities: monster 13, player 19 during knockback. Original Unit mode is +0x10; sample before neutral/walk resumes. Record successful knockback and source eligibility. The old specific hostile-monster suggestion was not validated; setup must be established by the owning knockback contract. Mode-table rows do not establish execution.

## Gap attempts and retained limits

The historical 2026-10-14 pass inspected compiler/import paths, image listings, missing-field switch, name-link searches, descriptor callback, monster initialization/writer, treasure generation/allocation and localized-name writer. It settled the selected ASCII binding path, state/stat/handler identities, PetType ownership and mode stores described above. It did not close indirect conversion/default behavior (GDA-Q01/GDA-Q02), animation integration (GDA-Q05), treasure probability/counter/global correspondence (GDA-Q03), localization (GDA-Q04) or gameplay observation (GDA-Q06).

Earlier 2026-10-13 searches inspected candidate word stores, stride/field combinations, count stores and the player-mode loader, rejecting hits in unrelated item/path records. They did not complete every search rung or identify animation/count writers then. Later writer identification supersedes that uncertainty without proving complete production. No absence claim or new search campaign is made here. Supplied-record behavior remains distinct from producer integration. The linked gap file identifies blocked assertions.

## Historical corrections, vector preservation and review

The 2026-10-14 semantic migration kept the dictionary boundary, 20 rule families and numeric results while splitting mixed evidence labels. Original binary-layout fixtures and earlier whole-spec records are preserved privately. The current GDF-V01..GDF-V05 and GDA-V01..GDA-V17 use semantic fields and do not require original addresses/strides/packed keys. Correspondence flags were narrowed to measured assertion scope; supplied-input vectors did not become gameplay-observed.

Original vector results retained: Difficulty -100/20/30/4/2/3/1/25; town/shield 4/0/0/0; stat/state names 99/171/54/146; mode name/token/code distinctions; treasure Picks/NoDrop and quality values; skill parameter 35/guarded zeros; handler 67/76 versus 68; record range failures; mode masks 32/8/0; descriptor counts 22/8; stat 171/layer 0; Sorceress 20/120/-15; identity-only Experience/States; missile 64/128/0/0; Gold/gold and Any Shield/shld; class 560/link/Velocity/Run; interrupt 128/0. Exact inputs, operations and results remain in the spec.

Historical reader corrections added discriminators for conversion-kind termination, reordered headers, independent second text decode, index/level/overflow guards, byte limits and four-code truncation, missing-resource aborts, unsigned BlockFactor 200, asymmetric mode flags, in-range absent-record guards, low16 group wrapping, PetType 255/256, reserved-word copy 17/29 and decimal overflow. These are synthetic inputs, not new table-cell claims. This publication did not rerun fixtures or game reading.

Historical review on 2026-10-14: an early reader reported three required discrimination findings; a later pass reported no GDF discrimination finding while forwarding foundation findings. Decimal overflow follow-up required another reader. Dictionary pass 3 reported four required findings; pass 4 reported one; pass 5 reported zero required/advisory findings. Final applicable pass 5: MISMATCH 0, GUESS 0, GAP 0, GATED 7, COVERAGE 0 required / 0 advisory; current GDF-V01..GDF-V05 and GDA-V01..GDA-V17 assertions checked. It met the then-applicable reader stopping rule while retaining gates/draft status. This is historical review evidence, not approval or gameplay confirmation.

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
| `states.txt` | id 1; row 1 | `id` | `1` | state identity/name; GDF-R03 |
| `states.txt` | id 1; row 1 | `state` | `freeze` | state identity/name; GDF-R03 |
| `states.txt` | id 15; row 15 | `id` | `15` | state identity/name; GDF-R03 |
| `states.txt` | id 15; row 15 | `state` | `concentrate` | state identity/name; GDF-R03 |
| `states.txt` | id 21; row 21 | `id` | `21` | state identity/name; GDF-R03 |
| `states.txt` | id 21; row 21 | `state` | `stunned` | state identity/name; GDF-R03 |
| `states.txt` | id 42; row 42 | `id` | `42` | state identity/name; GDF-R03 |
| `states.txt` | id 42; row 42 | `state` | `concentration` | state identity/name; GDF-R03 |
| `states.txt` | id 54; row 54 | `id` | `54` | state identity/name; GDF-R03 |
| `states.txt` | id 54; row 54 | `state` | `uninterruptable` | state identity/name; GDF-R03 |
| `states.txt` | id 65; row 65 | `id` | `65` | state identity/name; GDF-R03 |
| `states.txt` | id 65; row 65 | `state` | `dodge` | state identity/name; GDF-R03 |
| `states.txt` | id 66; row 66 | `id` | `66` | state identity/name; GDF-R03 |
| `states.txt` | id 66; row 66 | `state` | `avoid` | state identity/name; GDF-R03 |
| `states.txt` | id 68; row 68 | `id` | `68` | state identity/name; GDF-R03 |
| `states.txt` | id 68; row 68 | `state` | `evade` | state identity/name; GDF-R03 |
| `states.txt` | id 92; row 92 | `id` | `92` | state identity/name; GDF-R03 |
| `states.txt` | id 92; row 92 | `state` | `death_delay` | state identity/name; GDF-R03 |
| `states.txt` | id 101; row 101 | `id` | `101` | state identity/name; GDF-R03 |
| `states.txt` | id 101; row 101 | `state` | `holyshield` | state identity/name; GDF-R03 |
| `itemstatcost.txt` | ID 7; row 7 | `ID` | `7` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 7; row 7 | `Stat` | `maxhp` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 93; row 93 | `ID` | `93` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 93; row 93 | `Stat` | `item_fasterattackrate` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 95; row 95 | `ID` | `95` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 95; row 95 | `Stat` | `lastblockframe` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 96; row 96 | `ID` | `96` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 96; row 96 | `Stat` | `item_fastermovevelocity` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 99; row 99 | `ID` | `99` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 99; row 99 | `Stat` | `item_fastergethitrate` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 102; row 102 | `ID` | `102` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 102; row 102 | `Stat` | `item_fasterblockrate` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 105; row 105 | `ID` | `105` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 105; row 105 | `Stat` | `item_fastercastrate` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 164; row 164 | `ID` | `164` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 164; row 164 | `Stat` | `skill_concentration` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 352; row 352 | `ID` | `352` | stat identity/name; GDF-R02 |
| `itemstatcost.txt` | ID 352; row 352 | `Stat` | `last_sent_hp_pct` | stat identity/name; GDF-R02 |
| `itemtypes.txt` | Gold; row 4 | `ItemType` | `Gold` | cited original-profile item-type name/code; GDF-R07 |
| `itemtypes.txt` | Gold; row 4 | `Code` | `gold` | cited original-profile item-type name/code; GDF-R07 |
| `itemtypes.txt` | Any Shield; row 51 | `ItemType` | `Any Shield` | cited original-profile item-type name/code; GDF-R07 |
| `itemtypes.txt` | Any Shield; row 51 | `Code` | `shld` | cited original-profile item-type name/code; GDF-R07 |

The supplied GDF-V01 identities 99 and 54 are corroborated here; the 171 and 146 cell records remain the compatible 2026-10-14 evidence. State 101 corroborates GDF-R04’s named state, while the code-fixed test still rests on its static consumer chain. ItemTypes name/code evidence does not verify link-building. The separate 32-byte animation-file header read establishes file presence only and supplies no GDA-Q05 animation-selection contract. Element-table reads likewise do not explain a selector handler or close a compiler conversion gap.

## Focused refresh and affected dependencies (2026-10-15)

All seven open gap rows were assessed against the matching source-manifest artifacts, the existing 2026-10-12/13/14 records and current foundations/expression contracts before considering new reading. Earlier source records retain their actual methods, dates and read boundaries. No old failed lookup/default or localization attempt was rerun, and no newly available reconstruction was treated as new proof. The animation selectors remain incomplete, indirect/compiler and missing-column behavior remain incomplete, global cached/text MonStats correspondence remains disputed, treasure probability/counter production remains with its gameplay owner, localized class names remain gated, and attributable execution remains pending. The new cells/hash coverage does not settle any of these questions.

The dependency review found two stale references: GDA-R04a and Acceptance had attributed code-built records to RLI-R01, whose current subrules consume this dictionary’s field/name correspondence; they now point to the foundations semantic-record sections and actual gameplay owners. Acceptance had called NRF-R01 wrapping notation although that rule owns the LCG; the supporting notation reference now targets foundations section 1. Neither change alters a field, conversion, call order, label or gate. GDF-V02 explicitly normalizes the RLI-R03 mask into its retained true/false predicates. Its supplied player/list/state-count inputs give raw state-54 mask 4194304 and absent state-101 result 0 under the existing static dependency; this is not a newly observed return.

All GDF-R/Q/V/O identities, the 22 existing vectors, existing dates, numerical expectations and seven gap boundaries remain. The main contract’s retrospective correspondence sentences were shortened and their original records remain above: GDA-V06 has 2026-10-13 compiled correspondence and 2026-10-14 corroboration; GDA-V11/GDA-V09/GDA-V07 and PetType retain their dated records. Synthetic fixture encodings and the complete pre-refresh three-file snapshot are preserved with SHA-256 privately. No implementation comparison, implementation or plan change, table-handler completion or gameplay campaign was performed.

## Focused self-review (2026-10-15)

The mechanical spec/companion/gap check reported 0 errors and 0 warnings. The separate preservation comparison checked all 36 rule definitions and their labels, all 22 executable-vector identities, and all 277 earlier named-cell records. Only the dependency description of GDA-R04a changes among normalized rule bodies; other rule bodies are unchanged apart from spacing/identifier formatting. Vector correspondence prose changes in GDA-V06/GDA-V11/GDA-V09/GDA-V07 retain the dated records above. GDF-V02 adds explicit dependency inputs and mask-to-predicate assertions while retaining its selected identities and true/false outputs.

The self-review mapped every verified behavioral rule to its existing supplied-input vector and each data-only rule to named-cell records, checked the dependency return formats and seven precise gap boundaries, and found no additional result-distinct discriminator needed for the unchanged local operations. The new text-cell records carry their own date, full-file coverage, header-based selection, narrow values, hashes and profile limits. This is a documentation check and self-review, not a fixture execution, gameplay observation, fresh independent-reader result or approval.

## Independent review spot check (2026-10-15)

An independent review read original 1.13c instructions before comparing them with the
rules: Fog `0x12A60` numeric handler `0x1316F..0x13193`, its kind dispatch bytes at
`0x13450` and jump table `0x13438` (GDA-R03b); D2Game `0x1560` (GDA-R09b), `0x1A40`
(GDA-R11b), `0x77FA0` selector compare (GDF-R09/GDA-R16); D2Common `0x32890` (GDF-R04),
`0x5A030` conversion block `0x5A556..0x5A5B0` (GDA-R15b), `0x5F310` (GDA-R04b), `0x5EF40`
(GDA-R01), `0x603F0` (GDA-R05) and the class-record bounds in `0x31D20` (GDA-R06a). All
stated guards, widths, masks and orders were confirmed: decimal accumulation with a final
negation; kinds 2/3/4/6 store unconditionally while kind 5 skips an empty cell (kind 5 is
not one of the direct kinds this contract claims); availability mask `1 << (mode & 7)`;
`Param3 + (level-1)*Param4`; selectors 67/76 before mask[7]; count `> 171` before stat
171; nonzero group plus counter stored as low16 and Picks 0 replaced by 1; PetType signed
`>= 256` fatal and nonpositive published as 0; independent generation and final-form flags
with no fallback. GDA-R01's text request substitutes `levels` for the table name
`leveldefs` (strings at D2Common RVAs `0x8F8E8`/`0x8FA90`); that table is outside this
dictionary's named collections and is reported for the owner rather than added here.
Recomputed GDA-V17, GDA-V14, GDA-V12 and GDA-V03 matched. The review also separated words
run into numbers in the compiled-correspondence bullets, count records and the MonStats
discrepancy examples above; no value changed.

## Focused blind-reader pass FULL1 (2026-10-15)

The reader examined the main contract and only the allowed table-evidence ranges. Reported counts: MISMATCH 0, GUESS 0, GAP 3, DISCRIM 3, GATED 6 (GDA-Q05/GDA-Q01/GDA-Q03/GDA-Q06/GDA-Q02/GDA-Q04), COVERAGE 4, HARD 3. This was a fresh reader result, not a gameplay observation or approval. Its findings supersede the initial self-review's conclusion that no further discriminator was needed.

All three inventory gaps were addressed: section 3 now lists the mode argument to GDA-R09b; section 5 lists the wrong-count Difficulty abort from GDA-R05 and the Experience publication/no-separate-count result from GDA-R07. No behavior, field width or gate changes.

The three discrimination findings were addressed with two distinct supplied fixtures. GDA-V09 now separates collection index 1 from stored identities 2/1 and field values 11/22, rejecting identity-based lookup in place of indexed retrieval. GDA-V12 separates monster index 1 from its linked MonStatsEx index 2, with opposite mGH flags; mode 3 must yield 8 instead of 0. GDA-V10 retains its original identity case and cites GDA-V09's index discriminator, avoiding a duplicate fixture.

Coverage suggestions were triaged individually:

- Byte-count versus Unicode-character limits: GDA-V03 now retains [0xC3] from supplied bytes [0xC3, 0xA9] at byte limit 1. This is supported by GDA-R03b's byte contract and makes no locale/encoding claim.
- Difficulty variants: GDA-V10 now supplies distinct Drain/Drain(N)/Drain(H) values 7/13/29 and asserts selection for difficulties 0/1/2 under GDA-R08a. This adds no leech or table-cell claim.
- True NoShldBlock: GDA-V11 now supplies inTown false/NoShldBlock true and asserts raw outputs 0/4 and predicates false/true under GDA-R16.
- Intermediate/final text-record zeroing: deferred at the original decoder/compiler execution boundary. GDA-V01 starts with distinct supplied decoded collections and proves load-selection/publication control flow. Their supplied fields cannot discriminate whether the original compiler zeroed allocation before conversions. Adding a fake local producer operation or silently treating supplied decoding as original compilation would cross GDA-R02's GDA-Q01/GDA-Q02 preparation gate. The static zeroing evidence remains above; no new local zeroing-execution assertion is claimed. GDA-V01 now states this limit explicitly.

The three HARD cases required no changes: GDA-V01's independent generation/final selection and second text decode; GDA-V03's signed wrapping and low-width stores; GDF-V02's stat/state identity, raw mask and nonzero predicate. All 22 vector families, earlier cases/results, rule labels, question identities, dates and acceptance gates remain. The added inventories and executable assertions require a full second reader pass.

## Focused blind-reader pass FULL2 (2026-10-15)

The reader examined the actual main contract and only the allowed table-evidence ranges. Reported counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 9, COVERAGE 1, HARD 3. This is documentation review, not gameplay observation or approval.

The required discrimination finding concerned GDA-V17's equal Rare/Set/Unique values of 983, which could hide swapped quality copies. The original case and all existing variants remain. One supplied quality-copy fixture now gives Magic/Rare/Set/Unique distinct values 101/202/303/404 and requires those same-named destination values, with all other original inputs and outputs unchanged. Any swapped quality pair changes the result. This follows the existing GDA-R15b copy contract; it adds no original-table cell or producer claim.

The coverage suggestion noted that GDA-V06's reordered-header case gives nonzero values to only two of the 22 fields. It remains advisory: the case already discriminates header-name binding from input position, the direct conversion recipe is separately exercised, and the named per-field descriptor/data records support the remaining mapping. Twenty more cases repeating the same binding/conversion recipe would add no distinct behavioral branch or result. No additional exhaustive mapping fixtures were added.

The three HARD cases required no changes: GDA-V12's collection index to extended link to returned mask, GDA-V01's independent generation/final selection and separate second text decode, and GDA-V03's wrapping, low-width stores and raw-byte limits remain explicit.

The nine GATED outcomes cover GDA-R03d/GDA-R06b/GDA-R08b/GDF-R06b/GDA-R14/GDA-R02 and GDF-O01/GDA-O01/GDA-O02. They retain the existing GDA-Q05/GDA-Q01/GDA-Q03/GDA-Q06/GDA-Q02/GDA-Q04 assertions and execution boundaries; this count uses a different reporting unit from FULL1's six question identities and closes no gap. All 36 rule labels, 22 vector families, earlier cases/results and seven gap rows remain. The new assertion requires FULL3, the third full campaign reader pass.

## Focused blind-reader pass FULL3 (2026-10-15)

The fresh third full reader reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0 after recomputing all 22 vector families. These were documented hand recomputations, not implementation tests or gameplay execution. No behavior or fixture corrections were required.

GATED 6 retains the six existing question boundaries:

- GDA-Q05: complete derived monster animation-rate production.
- GDA-Q01: complete name/code links and expression/callback conversions.
- GDA-Q03: treasure generation and global compiled correspondence.
- GDA-Q06: attributable gameplay/runtime observation.
- GDA-Q02: missing-column/default and complete linked-field behavior, including absent direction columns.
- GDA-Q04: localized class display names.

The reader's four COVERAGE observations were retained with these explicit dispositions:

- GDA-R01 pre-conversion zeroing is explicitly untested by the supplied decoded outcomes in GDA-V01. The existing static zeroing record is not an execution claim; no fake producer fixture was added.
- Original decoder/writer failure integration remains GDA-R02's stated preparation boundary. GDA-V01 discriminates selected-resource absence and supplied successful outcomes; it does not establish all decoder/writer failure behavior.
- GDA-R09b's upper mode domain is sampled by mode 13, with modes 3/4 also covered. Its signed linked-index guards, byte/mask selection and result distinction are explicit. No exhaustive repetition across modes 8..15 was required or claimed.
- GDF-R06b has no nonzero supplied-direction fixture. The current vectors cover the stated zero-valued preparation case; missing direction-column defaults remain GDA-Q02. No complete original preparation claim or observation is inferred from that case.

The three HARDEST checks were GDA-R01/GDA-V01's independent generation/final selection and publication/abort ordering; GDA-R09b/GDA-V12's collection index versus linked identity and returned mask; and GDA-R03b/GDA-V03's wrapping, narrow stores, flags, code padding and raw byte strings. Their explicit assertions needed no changes.

No unnecessary fixture expansion followed this pass. All 36 rule labels, 22 vector families and earlier cases/results, six question identities, seven gap rows, evidence labels and draft status remain. The main metadata's trailing hard-break spaces were removed solely to satisfy publication formatting. The mechanical check after this record reported 0 errors and 0 warnings. This review does not close the stated evidence gates, establish gameplay parity or imply user approval.

<a id="shared-unit-condition-name-2026-10-15"></a>
## Shared unit-condition name (2026-10-15)

The 2026-10-15 spec review found one original unit condition carried under several names in
consuming contracts. GDF section 3 now gives it one semantic name, `aliveExclusionCondition`.
No new reading was performed: every consuming companion below already maps its own name to the
same original location from its own dated reading, and each consumer keeps its test, label and
gate.

| Semantic name | Original location | Consuming readings (owning companion) |
|---|---|---|
| `aliveExclusionCondition` | Unit byte `+0xC6`, bit 0 (mask 1) | D2Game `0xDB4C0` alive computation (combat-status-application, CSA-R13); D2Game `0xDFA90` unavailable-unit test (skill-assassin-sentry-firing-and-range; owner test in skill-assassin-blade-sentinel-and-fury); D2Game callback `0x37810` of the player-count walk (monster-scaling-by-players-and-difficulty, MSD-R05); party member collection (player-party-experience-sharing, player-party-gold-sharing); timed-state removal callbacks (state-timed-state-lifecycle); cast-start finishing predicate (skill-cast-start-and-do-dispatch, SCD-R09a); open-wounds expiry (combat-open-wounds); damage apply (combat-damage-pipeline-order); leech and drain (combat-leech-and-drain); movement and freeze (state-movement-modifiers-and-freeze) |

Names in use on 2026-10-15, to be replaced by the settled name in each contract:

| Contract (spec, companion, gap file) | Name in use |
|---|---|
| combat-status-application, combat-leech-and-drain, state-movement-modifiers-and-freeze | `aliveExclusionCondition` (already the settled name) |
| state-timed-state-lifecycle, combat-open-wounds | `expiryDeathCondition`; the open-wounds companion also says "neutral expiry/death condition" |
| monster-scaling-by-players-and-difficulty | `notAliveCondition` |
| player-party-experience-sharing, player-party-gold-sharing | `excludedFromPartyCollection` |
| skill-cast-start-and-do-dispatch, skill-assassin-sentry-firing-and-range, skill-assassin-blade-sentinel-and-fury | "completion condition" |
| combat-damage-pipeline-order | "invalid condition"; its companion also says "invalid-condition byte" |

`aliveExclusionCondition` was chosen because it is already used by three contracts and states
only the established behavior: alive and availability tests treat the unit as not alive when it
is set. The other names each describe one consumer's use (expiry, party collection, skill
completion, an invalid unit) and would imply a meaning not established for the others. Its
producers (which code sets or clears it) were not traced in this pass; wider meaning remains with
the consumers' open questions (for example CSA-Q03, STL-Q06, MSD-Q01).

Applied later on 2026-10-15: every contract in the table now uses `aliveExclusionCondition` in its
spec, companion and gap file. Two dated blind-implementer records in the
skill-assassin-blade-sentinel-and-fury companion keep the quoted wording "completion condition"
as it was proposed at the time.

<a id="restructure-2026-10-15"></a>
## Restructure (2026-10-15)

The user approved the spec review's recommendation to split oversized specs. The combined
36-rule dictionary was split by record family:

- Stayed here: class, stat and state identities, the fixed stat/state code roles, player and
  monster mode identities with their mode-table preparation, item-type identities, monster and
  skill identities and the handler-selector comparison (11 rule ids), their five vectors and the
  mode-change observation GDF-O01. Mode-table preparation stays with the mode identities because
  its whole output is those tables' names, tokens and direction values.
- Moved to [Game Data Table Records](core-game-data-table-records.md) (prefix GDA): the column
  schema and text conversions, table loading and ownership, PetType count, difficulty, class,
  experience, monster, extended-monster, missile, skill, stat-definition, state-definition,
  item-type field and treasure rules, and the flag-test results (25 rule ids), with 17 vectors,
  the animation and PetType observations and the six existing questions.
- The observation question was split: GDA-Q06 covers the table records; a new GDF-Q01 covers
  this spec's observation boundary (GDF-O01).
- Vectors that span both halves went to the spec owning the asserted behavior: GDF-V02 (fixed
  stat/state selection) and GDF-V04 (item-type identity and Code field) stay here and also serve
  GDA-R12 and GDA-R14; GDA-V07 (class getter) and GDA-V10 (monster identity record) moved with
  their getters and cite GDF-R01/GDF-R08.

No rule's behavior, label, value or vector result changed. Ids were renumbered in each file in
execution order (the spec is not consumed by a plan). Function, data-location and named-cell
rows were filtered by the rules they support; rows supporting both specs appear in both
companions. Dated records were kept with ids updated to the current numbering. The descriptor
map, historical live-memory correspondence and MonStats discrepancy are in the GDA companion;
the 2026-10-15 named-cell corroboration is here. The same pass added the shared unit-condition
name (section 3 of the spec; record above). The gap file was split to match. No new reading,
table read, fixture execution or gameplay action was performed.

## Focused publication review of the split dictionary (2026-10-23)

A fresh blind implementer read the current split GDF contract and cited GDA/RLI/NRF dependencies without this companion or historical combined-contract reviews. It recomputed GDF-V01–GDF-V05 at their stated data-only and supplied-input boundaries: the state-54 mask is `1 << 22 = 4194304`, the 359-versus-171 stat-definition guard, padded four-byte `GH  ` token, all 16 supplied monster direction values, four-byte `gold`/`shld` codes, and the selector 67/76/68 branches. It also checked 96 stat and 21 state decimal/hex pairs and contiguous 20 player/16 monster mode ids, finding no mismatch or duplicate. This is internal contract review, not a new original table read or runtime observation. The review identified one clarity risk: V03's all-zero direction inputs could be mistaken for original MonMode text columns. The writer labeled those values synthetic and retained GDA-Q02's missing-column-default gate. A separate blind delta review checked that wording against GDF-R06b and GDA-Q02.

An independent publication review checked this spec, companion, gap file, source manifest and index. It confirmed that the earlier 2026-10-15 full blind passes reviewed the former combined 36-rule/22-vector contract, not the current 11-rule/5-vector GDF split; the current blind review above supplies the applicable pass. The live #470 body names only this dictionary's three artifacts. The separate GDA table-records handoff, created by the 2026-10-15 split, is now tracked by [issue #568](https://github.com/gravestench/dark-magic/issues/568); the [index](../../specs/README.md) points that sibling to its own issue. GDA-Q02/GDA-Q03 and GDF-Q01 remain open. No new binary reading, effective-file/compiled-profile equivalence, gameplay observation, implementation or 1.14d parity is claimed; draft status remains appropriate.
