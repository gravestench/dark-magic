# Original game source manifest

This manifest identifies original game artifacts used as behavioral evidence. It publishes
fingerprints and navigation metadata; it does not distribute the original files. Public
spec evidence cites these artifact ids so a reader with matching files can repeat the
analysis without access to the private reconstruction project.

## Lord of Destruction 1.13c — Windows x86

Binary evidence profile: original Lord of Destruction Expansion 1.13c for Windows.
Historical data-profile identity has the limits stated below; binary identity alone does
not establish an unmodified archive set.
All files below are 32-bit x86 PE images (`Machine=0x014C`, `PE32=0x010B`), with
little-endian numeric storage. Fingerprints were captured on 2026-10-14.

| Artifact id | Original filename | Bytes | SHA-256 |
|---|---|---:|---|
| `LOD113C-D2COMMON` | `D2Common.dll` | 679936 | `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10` |
| `LOD113C-D2GAME` | `D2Game.dll` | 1138688 | `24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53` |
| `LOD113C-FOG` | `Fog.dll` | 212992 | `96c05770ab06b71a7a2d6abe5dc7c2f943b30ac8fe367e44bdbad7e2643bb863` |
| `LOD113C-STORM` | `Storm.dll` | 372736 | `4b5fcaf87c98676e6be94f980b06b893c50f1505343cbbf038c5b331a2da2d5b` |
| `LOD113C-GAME` | `Game.exe` | 61440 | `74fe9c092a521f7710392548c82f81544e531107db2358617e83818874db40a2` |

| Artifact id | Preferred image base | File version resource | Product version resource |
|---|---|---|---|
| `LOD113C-D2COMMON` | `0x6FD50000` | Not available | Not available |
| `LOD113C-D2GAME` | `0x6FC20000` | Not available | Not available |
| `LOD113C-FOG` | `0x6FF50000` | Not available | Not available |
| `LOD113C-STORM` | `0x6FBF0000` | `1.08b` | Empty |
| `LOD113C-GAME` | `0x00400000` | `1, 0, 13, 60` | `1, 0, 13, 257` |

The build label identifies the source bundle. A DLL's resource version can differ from
that label; matching the exact file fingerprint establishes whether a cited location
applies. These five fingerprints agree across two stored original bundles. The four DLL
fingerprints also match the recorded executable hashes of the analysis images used for
instruction-listing corroboration. This correspondence identifies the analyzed specimens;
fingerprinting adds no gameplay confirmation.

The historical session record includes the matching `Game.exe` SHA-256. The historical
static stat-reader record includes the matching `D2Common.dll` SHA-256. Full fingerprints
of the DLLs loaded by the historical runtime session were not recorded; the fresh loader
fingerprint identifies the available original `Storm.dll` analysis specimen, rather than
proving the exact historical loaded DLL identity.

## Reading a function reference

An RVA is an address relative to its module's image base. For example, a `D2Common.dll`
RVA of `0x36760` has preferred virtual address `0x6FD86760`. A loaded base can differ.
An RVA is not a physical file offset: use a PE-aware disassembler or debugger to map it.
Export ordinals apply to the identified module fingerprint. Descriptive function names
in evidence tables are analyst labels unless explicitly identified as original exports.

To check identity, compute SHA-256 of the original file, for example with PowerShell
`Get-FileHash -Algorithm SHA256 -LiteralPath <path-to-owned-binary>`, and compare its
size and hash above. Open that matching image in a PE-aware analysis tool, navigate to
the cited RVA, and follow the branches/callees described by the evidence companion.
This requires the reader's own original files, not unpublished reconstructed source.

## Game-data identity and limits

The [numeric and RNG evidence](spec-evidence/core-numeric-and-rng-foundations.md),
[game-data field evidence](spec-evidence/core-game-data-fields.md) and
[mitigation evidence](spec-evidence/combat-damage-mitigation.md) record narrow original-profile
table reads by effective virtual filename, row convention, column and named value. The
companions retain their own dates, coverage and compiled/memory correspondence limits.
Those reads used the running game's `Storm.dll` loader. `D2Common.dll`,
`D2Game.dll` and `Fog.dll` support the cited code rules; `Storm.dll` supports the cited
string/header helpers and the table access method; `Game.exe` identifies the recorded
session executable.

Full hashes of the effective `.txt`/`.bin` files, the physical supplying MPQs, and a
complete override inventory were **not captured** for those historical reads. The named
cells remain the scope of that evidence. Matching executable fingerprints alone does
not establish matching archives or effective game data. A future data profile should
record effective-file hashes and known overrides, and name a physical archive only
when its contribution has been established.

Historical profile identification also used the reported vanilla edition/build and profile
family, with no known overrides reported; the executable hash alone is insufficient because
modded profiles can share `Game.exe`. A named text cell, compiled field and cached runtime
field establish only the correspondence recorded for that case, not general table equivalence.
