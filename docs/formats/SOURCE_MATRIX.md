# Diablo II legacy format source matrix

This document records the external evidence used by Dark Magic's clean-room research into DT1, DS1, DC6, DCC, COF, composite animation, and DRLG map generation.

The goal is to recover observable format and runtime behavior, not to translate another engine. Facts extracted from implementation sources must be restated independently and corroborated where possible.

Gameplay-system, save, item, stat, and timing sources are tracked separately in [../research/SYSTEMS_SOURCE_MATRIX.md](../research/SYSTEMS_SOURCE_MATRIX.md). Sources that inform both binary formats and gameplay may appear in both matrices with the role and inspected paths made explicit.

## Confidence labels

- **verified**: direct binary observation or multiple independent sources agree.
- **high**: strong historical documentation or reverse-engineered original runtime behavior.
- **medium**: one mature implementation with no known conflict.
- **uncertain**: inference, incomplete field meaning, or conflicting sources.

## Sources

| Source | Version / commit | License | Role | Material used |
| --- | --- | --- | --- | --- |
| Paul Siramy DT1 documentation | historical site, `paul.siramy.free.fr/_divers/dt1_doc/` | historical documentation | format documentation / original modding research | DT1 layout, tile/block concepts, DS1/DT1 relationship, coordinate semantics. Original site is intermittently unavailable and should be mirrored only through lawful archival references. |
| Paul Siramy Win_DS1 / DS1 editor docs and tools | historical | historical documentation / source distribution | reverse-engineering evidence | DS1 version behavior, layer order, object/path/substitution semantics, map coordinate terminology. |
| Paul Siramy MergeDCC / animation extractor material | historical | source distribution; verify per archive | reverse-engineering evidence | DCC/DC6 decoding edge cases and animation composition history. |
| The Phrozen Keep | historical forum corpus | per-post copyright | historical corroboration | old file-format discussions, COF/composite behavior, level TXT semantics, renderer limitations, and quest-record research. Forum claims are never treated as authoritative without corroboration. |
| ThePhrozenKeep/D2MOO | `3b21043b99e987bad41cf0f7b49f1f246db52d5c` | MIT | strongest reverse-engineered original runtime source | 1.10f DRLG, tile lookup/rarity, packed DS1 cell semantics, collision, level TXT structures, coordinate transforms, RNG, and server-authoritative quest controllers for all five acts. |
| jaenster/libd2 | `e6cdc4927c6180be8dd309b0423b470f64f1fc6c` inspected 2026-08-10 | MIT for clean-room source; embedded Blizzard-derived blobs/maps are explicitly excluded | clean-room 1.14d implementation and high-value independent corroboration | Deterministic DRLG for all acts and difficulties, maze/outdoor/preset rules, RNG/call order, room/tile materialization, collision, warps, objects/monsters, DS1/DT1 parsing, pathfinding, and verification methodology. Treat source behavior as implementation evidence; reproduce claims with owned-asset/original-runtime probes before making them Dark Magic contracts. |
| collinsmith/riiablo | `59aa8d5021c457e8d2891bb23c9e3a0ef724c6b3` | Apache-2.0 | independent implementation / corroboration | raw DT1/DS1/DC6/DCC/COF parsing, DCC cell reconstruction, composite resource resolution, component updates, palette transforms. |
| OpenDiablo2/OpenDiablo2 | current `master` inspected 2026-08-09 | GPL-3.0 | independent implementation / corroboration only | COF parsing, composite path construction, timing model, direction mapping, DC6 parsing. No code, comments, or mechanically translated algorithms may be copied into Dark Magic. |
| eezstreet/OpenD2 | current `master` inspected 2026-08-09 | GPL-3.0 | historical corroboration only | additional independent renderer/UI/file-format behavior where useful. No implementation code may be copied. |
| Dark Magic codec modules | tagged versions in Dark Magic `go.mod` | repository-specific; currently maintained as independent modules | current implementation under test | identifies what the current API already exposes and what future codec work belongs outside the engine repository. |

## Primary source paths inspected

### D2MOO

- `source/D2CMP/include/D2CMP.h`
- `source/D2Common/include/D2Seed.h`
- `source/D2Common/src/D2Seed.cpp`
- `source/D2Common/include/DataTbls/LevelsTbls.h`
- `source/D2Common/src/DataTbls/LevelsTbls.cpp`
- `source/D2Common/include/Drlg/D2DrlgRoomTile.h`
- `source/D2Common/src/Drlg/DrlgRoomTile.cpp`
- `source/D2Common/src/Drlg/DrlgDrlg.cpp`
- `source/D2Common/src/Drlg/DrlgPreset.cpp`
- `source/D2Common/src/Drlg/DrlgMaze.cpp`
- `source/D2Common/src/Drlg/DrlgOutdoors.cpp`
- `source/D2Common/src/Drlg/DrlgTileSub.cpp`
- `source/D2Common/include/D2QuestRecord.h`
- `source/D2Game/include/QUESTS/Quests.h`
- `source/D2Game/src/QUESTS/Quests.cpp`
- `source/D2Game/src/QUESTS/ACT1/` through `ACT5/`
- `doc/Coordinates.md`

### Riiablo

- `core/src/main/java/com/riiablo/map5/Dt1.java`
- `core/src/main/java/com/riiablo/map5/Dt1Decoder.java`
- `core/src/main/java/com/riiablo/map5/Dt1Decoder7.java`
- `core/src/main/java/com/riiablo/map5/Tile.java`
- `core/src/main/java/com/riiablo/map5/Block.java`
- `core/src/main/java/com/riiablo/map5/Ds1.java`
- `core/src/main/java/com/riiablo/map5/Ds1Decoder.java`
- `core/src/main/java/com/riiablo/file/Dc6.java`
- `core/src/main/java/com/riiablo/file/Dc6Decoder.java`
- `core/src/main/java/com/riiablo/file/Dcc.java`
- `core/src/main/java/com/riiablo/file/DccDecoder.java`
- `core/src/main/java/com/riiablo/file/Cof.java`
- `core/src/main/java/com/riiablo/file/Animation.java`
- `core/src/main/java/com/riiablo/engine/client/CofResolver.java`
- `core/src/main/java/com/riiablo/engine/client/CofLayerLoader.java`
- `core/src/main/java/com/riiablo/engine/client/CofLayerCacher.java`
- `core/src/main/java/com/riiablo/engine/client/CofAlphaHandler.java`
- `core/src/main/java/com/riiablo/engine/client/CofTransformHandler.java`
- `core/src/main/java/com/riiablo/engine/server/PlayerItemHandler.java`

### OpenDiablo2

- `d2common/d2fileformats/d2cof/cof.go`
- `d2common/d2fileformats/d2cof/cof_dir_lookup.go`
- `d2common/d2fileformats/d2dc6/dc6.go`
- `d2core/d2asset/composite.go`

### libd2

- `docs/VERIFICATION.md`
- `packages/drlg/src/drlg/rng.zig`
- `packages/drlg/src/drlg/DrlgLogic.zig`
- `packages/drlg/src/drlg/DrlgRoom.zig`
- `packages/drlg/src/drlg/preset.zig`
- `packages/drlg/src/drlg/maze/`
- `packages/drlg/src/drlg/outdoors/`
- `packages/drlg/src/drlg/tilegen.zig`
- `packages/drlg/src/drlg/materialize.zig`
- `packages/drlg/src/collision.zig`
- `packages/drlg/src/objects.zig`
- `packages/formats/src/ds1.zig`
- `packages/formats/src/dt1.zig`

## Conflict policy

When sources disagree, use this order:

1. lawful observation of original assets;
2. Paul Siramy / contemporary format research;
3. D2MOO reverse-engineered 1.10f runtime behavior;
4. libd2 behavior backed by its documented retail-engine verification;
5. agreement among multiple independent implementations;
6. one implementation;
7. inference.

Do not hide disagreements. Mark them in the relevant format document and add a probe in `CODEC_FOLLOWUPS.md` when a binary-level fact needs confirmation.

## Known conflicts requiring probes

- **COF rate field:** Riiablo's later parser treats the four bytes following the COF bounding box as a 32-bit animation rate. OpenDiablo2 exposes the final byte as speed and treats the preceding three bytes as unknown. The Dark Magic codec must preserve all four bytes until owned-asset probes establish the observed value distribution and playback interpretation.
- **DT1 header version wording:** historical tools often describe DT1 as version `7.6`; later parsers commonly expose `7` as the first 32-bit field and the second field as flags/version-like metadata. Probe representative owned DT1 files and preserve both 32-bit words rather than collapsing the distinction prematurely.
- **Renderer limits vs format limits:** historical 256x256 DC6 and small DCC decode-buffer restrictions are original renderer/tool constraints, not sufficient evidence of binary-format maxima.
