# Evidence: attack hit roll

Spec: [Attack hit roll](../../specs/combat-hit-roll.md)

## Binary identity and method

The [source manifest](../source-manifest.md) identifies `LOD113C-D2GAME`, `LOD113C-D2COMMON` and `LOD113C-FOG`, original LoD 1.13c Windows x86. Original-image rehash 2026-10-15 matched D2Common SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`, size 679936. Earlier static records are reused within their dated boundaries; matching current specimens does not prove historical loaded-DLL identity. Historical data profiles lack complete archive/override hashes.

Descriptions are analyst labels. Original binary locations, calling conventions and layouts below are investigation evidence, not required implementation layouts. Detailed private source matching records, commands, session addresses and original fixture encodings are archived privately with the complete original files and integrity manifest.

## New readings and corrections (2026-10-15)

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2COMMON` | 0x32890 | 10672 | Total defense | HIT-R01a/HIT-R01b | Disassembly, all 896 bytes decoded against matching original image; direct arithmetic and Holy Shield request, dependency results supplied |
| `LOD113C-D2COMMON` | 0x31EA0 | 10621 | Player attack base | HIT-R02a | Disassembly checked against original 240-byte extent |
| `LOD113C-D2COMMON` | 0x39430 | 10910 | Parameterized selected stat read | HIT-R02b/c/d/e/g/HIT-R09a | Full disassembly; forwards bounded definition and layer to 0x38A80, owned RLI-R02a/RLI-R02b |
| `LOD113C-D2COMMON` | 0x1250 | none | Skill record lookup | HIT-R01b | Reconstruction checked against original machine code; signed bounds then row selection |
| `LOD113C-D2COMMON` | 0x51BF0 | none | Expression front-end | HIT-R01b | Reconstruction checked against original machine code; context(unit, skill, level), SCX-R11 interpreter boundary |
| `LOD113C-D2COMMON` | 0x4F640 | 10804 | Mastery consumer | HIT-R02f | Targeted full listing; special 0x4F200 and weapon selection remain unresolved |
| `LOD113C-D2GAME` | 0xA00B0 | none | State applier | HIT-R09b | Full 704-byte disassembly read; delegates attach/write/free effects to STL primitives with current gates |
| `LOD113C-D2GAME` | 0x372C0 | none | Selected weapon | HIT-R02f | Full 368-byte listing read; imported slot/equipment producers not accepted as defaults |

HIT-R01a/HIT-R02a arithmetic gates are closed by original-image opcode checks, not by recency or a new label. In total defense, 0x3298C bytes 8D3C28 add the two defense percentages; 0x32ADD bytes 0FAFFE multiply percent by base. At 0x31F42 bytes 8D5430F9 add tohit+dexterity-7, then the next instruction adds four Dexterities-28: the result is tohit+5*dexterity-35. Signed division/magic sequences were followed to returns. The total-defense negative branch divides its product by -100; subtraction of product/100 is the same signed result for the supported arithmetic.

Correction: the earlier statement that null total-defense input safely returns 0 was wrong. The initial null branch reaches 0x3290E, then 0x32919/0x3291D unconditionally reload/dereference the unit. The revised null failure claim does not change any existing numeric vector. The Holy Shield skill lookup call at 0x32A95 targets 0x1250, not the earlier 0x2250 locator. At 0x32ACC evaluation calls 0x51BF0; matched front-end confirms SCX context and no direct random draw. Expression-specific producers/draws remain owned by SCX, not erased by supplying a result. State-list reads 350/351 call 0x38B00, the own-array operation under RLI-R02b; ordinary unit readers use 0x38A80, selected-array operation.

The stat-getter question is closed: 0x39430 rejects invalid signed stat ids, missing list or missing definition; normalizes layer to `u16` and forwards to the same selected-array 0x38A80 that ordinary reads use. It dereferences its unit before list checks, so no null-unit default is established for that wrapper. Local callers use non-null units.

HIT-R09b remains asm-only: reading the complete applier resolves its request skeleton but does not establish all indirect writes. Compatible STL primitive contracts support the state-effect boundary; CSA-R05 is a cold-applier pattern, not the owner of this generic applier. The local hook establishes supplied request assertions; attach/replacing/free propagation gates remain. The earlier timed-list description is retained below as historical evidence, not promoted to verified stored effects.

Correction of a historical mastery locator: the D2Common + 0x4F640 instruction calling the context helper names a section-relative label corresponding to original RVA 0x303A0, not 0x2F3A0. The actual producer remains unread/gated; no mastery result is promoted. The generic applier record's zero field at+0x08 is called effectIdentity in the contract; its larger gameplay meaning is not established. D2Game thunk 0xAE3E resolves D2Common ordinal 10725/RVA 0x37D60, which writes list+0x1C; thunk 0xAE50 resolves ordinal 11081/RVA 0x37D40, which writes list+0x20 with the supplied level. Both setters and the identity getter ordinal 10555/RVA 0x37D50 were read as reconstruction checked against original bytes on 2026-10-15. These field writes do not justify a skill identity name or close the complete applier gate.

## Gap assessment

Every question from the prior contract was assessed. Named stat/state/mode/physical-field/flag questions reuse GDF/NRF current contracts or dated table evidence. Arithmetic decoding and getter identity closed with the readings above. Holy Shield calculation consumes SCX-R11 instead of a blanket unknown-RNG gate. Reaction/life, other patch versions and community comparisons are outside the existing boundary. Other passive-stat production, skill caller and PvP caller searches retain their actual partial coverage; no six-rung absence claim is made. Mastery, special monster, classifier, equipment and owner/precondition questions received the targeted reads above and remain precisely gated. No gameplay campaign was part of this migration; historical observations and pending observations retain their dates/scope. The gap file records every remaining behavior, integration, observation or outside-boundary item.

## Table data

On 2026-10-15 the running member reported the lod-113c profile. Its effective Storm-loader read of data/global/excel/itemstatcost.txt requested off 0, length 0 and returned 40109/40109 bytes: complete header/table coverage. SHA-256 of effective bytes: `ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`. Columns were selected by exact headers ID and Stat. Each explicit ID matched the zero-based data row excluding header, with a unique cell:117=item_preventheal;123=item_demon_tohit;124=item_undead_tohit;342=passive_mastery_melee_th;350=modifierlist_skill;351=modifierlist_level. These names are verified in 1.13c data at that effective-profile scope. No new compiled-bin correspondence, complete override inventory, physical supplying archive or historical loaded-DLL identity was established; reported profile identity alone is not proof of pristine data. These cells do not verify formulas or gameplay.

Historical 2026-10-13 States.txt row 52 name preventheal and the six ItemStatCost names were read through the game's loader. Original per-range/total-byte/header-selection metadata was not preserved publicly, so only their recorded narrow named-cell scope is reused. The fresh ItemStatCost read supplies complete coverage for its six names; no fresh States read is claimed. Earlier 2026-10-12 named class/monster/armor cells and formula correspondence remain in the historical record below; archive hashes and full effective-file coverage were not captured. Synthetic vectors need supplied fields, not those historical data identities.

## Historical evidence (2026-10-11 through 2026-10-13)

The following preserved records describe the prior reading and actual observation scope. Historical labels, open questions, review counts and correction narratives are evidence history, not the revised contract's current status. The current contract and new correction rationale above govern. Binary-layout numbers are preserved here for independent reproduction. Private source filenames were mapped to original module/RVA locations. Listing positions and command transcripts remain in the separately preserved private audit.

## 2. Trigger and ownership

**Module.** D2Game (server) computes the hit roll, the block roll, avoidance and the result word; D2Common (shared) computes the total-defense term, the player attack-rating base and the block chance. The coverage map places the owning folder at D2Game `UNIT\SUnitDmg`; no assert string naming a source file was read in the cited functions, so the folder is a lead.

**Dispatch chain A: melee attack (representative root).** Every link is a cited RVA; the bodies are in Appendix A.

1. Frame loop -> per-frame skill event -> do dispatch rva 0xA19A0 ([skill-cast-start-and-do-dispatch](../../specs/skill-cast-start-and-do-dispatch.md) SCD-R09, SCD-R06): `RVA 0xa19a0` `call eax` with `eax` read from the do-function table at D2Game `.data` + 0x84A8 (VA 0x6FD274A8), `ecx` = game record, `edx` = attacking unit.
2. Do-function slot 46 -> rva 0x93270 ([skill-leveled-damage-tohit-and-elements](../../specs/skill-leveled-damage-tohit-and-elements.md) SLD-R13 provenance for the slot number). `RVA 0x93270`: the defender is the unit stored in the attacker's path record (`PathRecord +0x58`, read through D2Common ordinal 10392, rva 0x348F0, reconstruction checked against original bytes); call rva 0x92A60 with `eax` = attacker, stack `(game, defender, skill id, x)` where `game` is the `ecx` of step 1.
3. Rva 0x92A60 (`RVA 0x92a60`): `eax` = defender, `edi` = attacker, stack `(game, bonus, 0)`, `bonus` = the skill's to-hit term plus stat 325 (SLD-R13; an input here) -> **rva 0xDE5A0** (result word; SBA-R05, SBA-R06, SBA-R12).
4. Rva 0xDE5A0 -> hostility gate rva 0xE0A60 (`RVA 0xde5a0`; SBA-R12) -> range test D2Common ordinal 11138, rva 0x326C0 (SBA-R12) -> **hit roll rva 0xDDE90** (HIT-R01 to HIT-R08) -> **block and avoidance roll rva 0xDB790** (SBA-R01 to SBA-R04) -> state test ordinal 10494, rva 0x33CD0 (SBA-R05).
5. Rva 0xDDE90 -> total defense ordinal 10672, rva 0x32890 (`RVA 0xdde90`; HIT-R01); player attack-rating base ordinal 10621, rva 0x31EA0 (HIT-R02a); defense/attack-rating adjustment rva 0xDB1F0 (HIT-R02c, HIT-R02d, HIT-R02g); melee-only term rva 0x372C0 and ordinal 10804, rva 0x4F640 (HIT-R02f); monster-type matrix test rva 0x1620 (HIT-R02b); percent helper rva 0x14D0 (HIT-R03); post-hit hook rva 0xDCE70 (HIT-R09).
6. Rva 0xDB790 -> block chance ordinal 10212, rva 0x31D20 (`RVA 0xdb790`; SBA-R01) -> moving test ordinal 10143, rva 0x313F0 (SBA-R02) -> **avoidance rva 0xDB030** (SBA-R04) -> weapon-avoid list rva 0xDA540 (`RVA 0xdb030`), token helper ordinal 10431, rva 0x441E0 , draw helper rva 0x1180 , attacker assert rva 0xDF780 (reconstruction checked against original bytes).

Twenty-nine other D2Game functions also call rva 0xDE5A0 (original D2Game RVAs `0x273c0`, `0x27820`, `0x27a10`, `0x289a0`, `0x28be0`, `0x44c20`, `0x44d70`, `0x451b0`, `0x4c2a0`, `0x4c720`, `0x4cb00`, `0x4cc50`, `0x4dac0`, `0x4f7d0`, `0x77240`, `0x77790`, `0x92940`, `0x92bb0`, `0x93040`, `0x93680`, `0x95690`, `0x98ba0`, `0x98e90`, `0x99d20`, `0xa1170`, `0xa11b0`, `0xa2c40`, `0xb41b0`, `0xb60c0`); each passes its own `(game, bonus, range)` and was not walked to its root (SBA-Q08). The monster melee attack observed at runtime (HIT-O01) entered rva 0xDDE90 from inside rva 0xDE5A0 (return address 0xDE646); its caller above 0xDE5A0 was not traced.

**Dispatch chain B: missile hitting a unit.**

1. Frame loop -> missile step rva 0x3F320 ([missile-movement-and-collision](../../specs/missile-movement-and-collision.md) MMC-R03) -> contact rva 0x3EDE0 (MMC-R11, MMC-R14; [missile-hit-and-damage-calculation](../../specs/missile-hit-and-damage-calculation.md) MHD-R01).
2. Rva 0x3EDE0 (`RVA 0x3ede0`): if the target is non-null and `MissilesRecord.ToHit (+0x18B)` is non-zero, **hit roll rva 0xDDE90** with `eax` = missile owner, stack `(target, Stat(missile, 19 tohit), isMissile = 1)`; a zero `ToHit` skips the roll and continues as a hit (SBA-R07a).
3. On a hit: damage fill rva 0x3A4E0 -> hand-off rva 0x3AFF0 (MHD-R11a, MHD-R11c) -> **apply rva 0x3AE10** (`RVA 0x3aff0`; MHD-R12 gives the argument roles `(game, missile, target)` with the damage record in `edi`).
4. Rva 0x3AE10 -> owner lookup rva 0xE03A0 (`RVA 0x3ae10`; SBA-R13) -> state test ordinal 10494  -> knockback draw rva 0x1180 (SBA-R11) -> **block and avoidance roll rva 0xDB790**  with `eax` = (`DamageRecord +0x08` != 0), `ecx` = owner, `esi` = target, stack `(game, 1)` -> own code-to-flag map (SBA-R06).

Rva 0xDE380 also calls rva 0x3AE10 (`RVA 0xde380`) and rva 0x3B040 also calls rva 0xDDE90; neither was walked (SBA-Q08).

**Dispatch chain C: avoidance only.** Rva 0xA10B0 (`RVA 0xa10b0`) calls **avoidance rva 0xDB030** with `eax` = defender, stack `(nDamageParam = 1)`, `edi` = the owner resolved by rva 0xE0A20 (not read; it supplies the attacker for the SBA-R10 assert; SBA-Q08), only when the result word `DamageRecord +0x04` is non-zero; it never rolls shield block and maps codes itself (SBA-R06). It is reached from rva 0x92BB0 and from the area callback rva 0xA1210 (flag-1 apply route, [combat-damage-pipeline-order](../../specs/combat-damage-pipeline-order.md) DPO-R08d); their roots were not walked (SBA-Q08).

**When it runs.** Once per attack contact, inside the attacker's skill do-function at its do frame (chain A) or inside the missile's movement step on the frame of contact (chain B), on the server; before the damage record is built and applied (the result word this spec produces is the input of [combat-damage-pipeline-order](../../specs/combat-damage-pipeline-order.md)). No timer event and no per-frame cadence belong to this spec; the only event scheduled is the timed-state expiry posted by the post-hit hook's applier (HIT-R09b).

**Client side.** None of the cited functions sends a message. The client learns the outcome through the unit-update and damage messages the damage pipeline sends after apply (names belong to a `net-` spec, none written); a block is visible as the defender's block mode (player mode 9 `Block`, monster mode 6 `block`, GDF-R05, GDF-R06), set by [combat-hit-recovery](../../specs/combat-hit-recovery.md).

## Historical named inputs

The 2026-10-12 loader record named these individual cells: CharStats.BlockFactor
(Barbarian) 25, CharStats.ToHitFactor (Barbarian) 20, MonStats.A1TH (`fallen1`) 101,
MonStats.Level (`fallen1`) 1, MonLvl.L-TH (level 1) 8, and Armor.block (Buckler) 0.
Their original archive/override hashes and complete read coverage were not captured;
these are reused historical named-cell records, not fresh original-profile verification.
The observations below retain the exact inputs and measured scope. Earlier full contract
copies and private listing-indexed working notes remain in the integrity-preserved audit;
the current main contract, function records and correction rationale state supported rules.

### 8.2 Observation vectors and reproduction procedure

Priorities: run vs walk block (`/3` only on Run and Town Walk), Dex < 15 (no block), player running (hit roll skipped), moving defender with Dodge but no Evade, a pure-elemental missile against a shield. Common setup for an owned 1.13c LoD game, single player, Normal difficulty: a level-1 character in Act I (Blood Moor, level id 2), attacked by Fallen (`fallen1`, monster level 1); read the defender's `Unit +0x20`/`+0x24` before each hit roll and the return values of the cited functions (HIT-O01 and SBA-O07 to SBA-O09 were reproduced this way).

- **HIT-O01** (HIT-R05, HIT-R06; HIT-R01a and HIT-R02e consistent). Observed 2026-10-12 (Appendix A): Fallen attacker, `AR = (101 * 8) / 100 = 8` (`MonStats.A1TH (fallen1) = 101`, `MonLvl.L-TH (level 1) = 8`), `bonus` 0, `isMissile` 0, `aLvl` 1; Barbarian defender, `DEF = 4 + 20 / 4 = 9` (stat 31 = 4, Dex 20, pct 0, stat 33 = 0; inferred from the chance), `dLvl` 1: ratio = 800 / 17 = 47; chance = (47 * 1 * 2) / 2 = 47. Reading: 116 hit-roll calls, every hit with roll <= 46 and every miss with roll >= 47 (roll = `edx` at return); the defender's seed never changed across a miss. Procedure to repeat: trace rva 0xDDE90 with a deref of the first stack argument (the defender) for 60 s of melee.

## Appendix A. Provenance

Historical readings: 2026-10-11, 2026-10-12 and 2026-10-13. Evidence is original LoD 1.13c Windows x86, read by disassembly or reconstruction checked against original bytes. Analyst labels are not original exported names.

**Function table.**

| Original module location | Form | Label | RVA | Ordinal | Rules |
|-------------|------|-------|-----|---------|-------|
| `D2Game.dll+0xdde90` | asm | `ProcessMonsterSkillRoll` (wrong) | 0xDDE90 | n/a | HIT-R01 to HIT-R08, SBA-R07a, SBA-R08, HIT-R09 |
| `D2Game.dll+0xdb1f0` | asm | `ProcessMonsterBehavior_8b30` (wrong) | 0xDB1F0 | n/a | HIT-R02c, HIT-R02d, HIT-R02g |
| `D2Game.dll+0x01620` | asm | `CheckSkillTableBit` | 0x1620 | n/a | HIT-R02b |
| `D2Game.dll+0x01220` | C | `GetMonsterRecord` | 0x1220 | n/a | HIT-R02b |
| `D2Game.dll+0x014d0` (read in NRF-R05) | asm | `CalculateScaledDivision` | 0x14D0 | n/a | HIT-R03 |
| `D2Game.dll+0x22a80` | asm | `CheckPlayerDataFlags` (wrong) | 0x22A80 | n/a | HIT-R02d |
| `D2Game.dll+0x372c0` | asm | `GetUnitPrimaryWeapon` | 0x372C0 | n/a | HIT-R02f |
| `D2Game.dll+0xdce70` | asm | `ProcessSkillAction_af90` (wrong) | 0xDCE70 | n/a | HIT-R09a |
| `D2Game.dll+0xa00b0` (historical initial-eligibility slice) | asm | `CreateSkillPath` (wrong) | 0xA00B0 | n/a | HIT-R09b |
| `D2Game.dll+0x3aff0` (historical apply-handoff call slice) | asm | `ProcessObjectTargetWithStats` | 0x3AFF0 | n/a | section 2 |
| `D2Game.dll+0xa19a0` (historical action-table call slice) | asm | `ExecuteSkillWithMissiles` | 0xA19A0 | n/a | section 2 |
| `D2Game.dll+0x93270` (historical target/hit-calculator call slice) | asm | `ProcessSkillHitDamage_1480` | 0x93270 | n/a | section 2 |
| `D2Game.dll+0x92a60` (historical hit-result call slice) | asm | `CalculateFullSkillDamage` | 0x92A60 | n/a | section 2 |
| `D2Game.dll+0xde380` (historical alternative apply-call slice) | asm | caller of 0x3AE10 | 0xDE380 | n/a | section 2 |
| `D2Common.dll+0x32890` | asm | `CalcTotalDefenseRating` | 0x32890 | 10672 (thunk 0xA832, slot 0x6FD18560) | HIT-R01 |
| `D2Common.dll+0x31ea0` | asm | `CalcCombatStatBonus` | 0x31EA0 | 10621 (thunk 0xA7F0, slot 0x6FD18534) | HIT-R02a |
| `D2Common.dll+0x4f640` | asm | `CleanupStatListModChain` (wrong) | 0x4F640 | 10804 (thunk 0xA80E, slot 0x6FD18548) | HIT-R02f |
| `D2Common.dll+0x386e0`, `RVA 0x382f0` | asm | `GetStatFromListByLayerChecked`, `STATS_CopyStatsByType` | 0x386E0, 0x382F0 | 10702 (thunk 0xA7D8, slot 0x6FD18524) | HIT-R02b, HIT-R02f, SBA-R04b |
| `D2Common.dll+0x38b70` (RLI-R02a) | asm | `GetUnitBaseStat` | 0x38B70 | 10973 (thunk 0xA42A, slot 0x6FD182B0) | stat reads |
| `D2Common.dll+0x39430` (not read) | asm | `GetMissileRecordField54` (wrong) | 0x39430 | 10910 (thunk 0xA802, slot 0x6FD18540) | stat reads 115, 116, 117, 119, 123, 124 (the stat-getter question) |
| `D2Common.dll+0x1a550` | asm | `FreeUnitStatList` (wrong) | 0x1A550 | 10064 (thunk 0xA580, slot 0x6FD18394) | HIT-R02c, HIT-R02d |
| `D2Common.dll+0x1a2d0` | asm | `SetStructField4C` (wrong) | 0x1A2D0 | 11104 (thunk 0xA7F6, slot 0x6FD18538) | HIT-R02c, HIT-R02d |
| `D2Common.dll+0x1a610` | asm | `GetUnitSeedHi` (wrong) | 0x1A610 | 10255 (thunk 0xA814, slot 0x6FD1854C) | HIT-R02g |
| `D2Common.dll+0x1a5b0` | asm | `SyncUnitPositionFromPath` (wrong) | 0x1A5B0 | 10239 (thunk 0xA508, slot 0x6FD18344) | HIT-R02g |
| `D2Common.dll+0x1ff40` | asm | `FindEquippedType33Item` | 0x1FF40 | n/a (internal) | SBA-R01a, SBA-R01e, HIT-R01b |
| `D2Common.dll+0x24430` | asm | `GetUnitDirection` (wrong; item-type equivalence test) | 0x24430 | 10744 (thunk 0xA430, slot 0x6FD182B4) | SBA-R01e, SBA-R04b, HIT-R02f |
| `D2Common.dll+0x348f0` | C | `GetPathRoom` (label; returns `arg +0x58`) | 0x348F0 | 10392 (thunk 0xA598, slot 0x6FD183A4) | section 2 |
| `D2Common.dll+0x51bf0` (not read; cited as the calc evaluator) | C | skill calc evaluator | 0x51BF0 | n/a (internal) | HIT-R01b |
| `Fog.dll+0x0fc20`, `RVA 0x11100` | C | `GetReturnAddress`, `CleanupAndAbort` | 0xFC20, 0x11100 | 10265 (thunk 0xA2DA), 10024 (thunk 0xA2C2) | fatal error 0xA2A path (SBA-R03, SBA-R04a, HIT-R09b) |
| D2Game do-function table (per SCD section 2) | data | `.data + 0x84A8` | VA 0x6FD274A8 | n/a | section 2 |

**Runtime records.** Member historical vanilla LoD profile, Lord of Destruction 1.13c, single player, Normal. Traces of 2026-10-12 were armed for 60 s with a cap of 200 records and both ended on the time limit; the 2026-10-13 row is a read of the records still held in the member's disarmed slot.

| Vector | Version | Member | Date | Trace RVA | Records | Result |
|--------|---------|--------|------|-----------|---------|--------|
| HIT-O01 | 1.13c | historical vanilla LoD profile | 2026-10-12 | D2Game+0xDDE90 (48-byte defender-argument read) | 116 (61 hits, 55 misses), caller D2Game+0xDE646 only, stack cleanup 12 bytes | match: HIT-R05, HIT-R06; 0 defender draws in the hit roll |
| HIT-O01, SBA-O08 (re-read) | 1.13c | historical vanilla LoD profile | 2026-10-13 | D2Game+0xDDE90 (disarmed, 116 records retained) | 12 records re-read: rolls 60, 41, 97, 7, 14, 60, 24, 85, 2, 73, 9, 85 with hits exactly where roll < 47; stack cleanup 12 bytes; caller 0xDE646; the defender seed went from `{0xBFC793C3, 0x29A}` before the hit of record 1 to `{0xDE8467A9, 0x4FFD63FE}` before record 2, which is exactly 1 LCG step (a blocked hit) | match |

Mismatches: none. Calling conventions measured: rva 0xDDE90 `ret 0Ch` with the attacker in `eax` and stack `(defender, bonus, isMissile)`; rva 0x31D20 `ret 8` with stack `(unit, applyScaling)`.

**Blind-implementer pass.** Run 1 (2026-10-13): GUESS 8, GAP 9; every item folded (section 3 rows for `Unit +0x04`, `StatList +0x48`, the item-type test, body slots 4 and 5 and the entry value of `DamageRecord +0x04`; SBA-R01a, SBA-R02, SBA-R03, SBA-R04, SBA-R06, SBA-R07a, SBA-R11, SBA-R13; section 5 failure rows; HIT-V08, SBA-V09, SBA-V14; HIT-Q02, SBA-Q11). Run 2 (2026-10-13, fresh reader): GUESS 6, GAP 5; every item folded (section 3 rows for the unit type and the attacker pointer; HIT-R02d null `MonsterData`; SBA-R01a gate order; SBA-R01b getter identity; SBA-R04b token-call order; HIT-V03, HIT-V06, HIT-V10, SBA-V09, SBA-V16 inputs). No third full review was completed for that revision.

**Per-claim ledger.**
- runtime-confirmed (1.13c): HIT-R05, HIT-R06, SBA-R01a (scaling branch), SBA-R03; the calling conventions above.
- verified in 1.13c code (static): HIT-R02b, HIT-R02c, HIT-R02d, HIT-R02e, HIT-R02g, HIT-R03, HIT-R04, HIT-R07, HIT-R08, SBA-R01b, SBA-R02, SBA-R04a, SBA-R04b, SBA-R04c, SBA-R05, SBA-R06, SBA-R07a, SBA-R08, HIT-R10, HIT-R09a, SBA-R10, SBA-R11; stat ids 2, 12, 19, 20, 31, 32, 33, 115, 116, 117, 119, 123, 124, 171, 16, 182, 179, 338, 339, 340, 348 as read from the getter arguments; item type 0x33; offsets +0x3C, +0x49, +0x18B, +0xAE, +0x1C (MonStats), +0x16 (MonsterData), +0x08 (damage record, role per NRF), +0x70 (game record).
- asm-only: HIT-R01b, SBA-R01d (role of `Game +0x70`), SBA-R01e, HIT-R09b, SBA-R12, SBA-R13.
- unresolved: HIT-R01a, HIT-R02a (one-instruction `db` runs, the arithmetic-decoding question); HIT-R02f, SBA-R01c, SBA-R04d, SBA-R04e, SBA-R07b, SBA-R09b.
- names: from core-game-data-fields, plus six names read from the 1.13c `ItemStatCost.txt` and `States.txt` through the member's own file loader on 2026-10-13 (rows 117, 123, 124, 342, 350, 351; state row 52).

**Caveats.**
- `db` runs read as instructions: `RVA 0x32890` (`8D 3C 28` = `lea edi,[eax+ebp]`, the `pct` sum) and (`0F AF FE` = `imul edi,esi`, the product) on the HIT-R01a path; `RVA 0x31ea0` (`8D 54 30 F9` = `lea edx,[eax+esi-7]`) on the HIT-R02a path; `RVA 0x326c0` (`8D 5C 10 01` = `lea ebx,[eax+edx+1]`) in SBA-R12; `RVA 0x24430` (`85 0C 16` = `test [esi+edx],ecx`; `8B 14 0E` = `mov edx,[esi+ecx]`) in the item-type test; the embedded table region of `RVA 0x313f0` is jump-table data (the player path is assembled); `RVA 0x4f640` (HIT-R02f loop); `RVA 0x31680` (SBA-R01c); `RVA 0x32890` (HIT-R01b); `RVA 0xe03a0`, `RVA 0x23ef0` and the prologue of `RVA 0x325c0` are whole `db` listings with the tool's decoding comments.
- Label corrections to earlier drafts: the missile owner resolver is rva 0xE03A0 (corrected from a section-relative analyst locator), not 0xDF3A0; the relation-flag helper is rva 0xF02A0 (corrected from a section-relative analyst locator), not 0xEF2A0; "attacker present" as the first argument of rva 0xDB790 on the missile path was wrong: it is `DamageRecord +0x08 != 0` (block-enable); `applyScaling` is `Game +0x70` of the game record passed down both chains, which closes the "if `arg1` is the game record" condition of the expansion-selector question.
- Status changes in this revision: HIT-R01a and HIT-R02a `verified` -> `unresolved` (assumption 3, the arithmetic-decoding question); SBA-R04d, SBA-R07b, SBA-R09b `not found in the original-image analysis` -> `unresolved` (no six-rung record existed); HIT-R09 split into HIT-R09a (verified) and HIT-R09b (asm-only) because the applier was not read to its writes; SBA-R04b's text corrected (a layer-0 entry assigns rather than maxes; layers are `u16`). New ids: SBA-R01e, SBA-R12, SBA-R13, SBA-V17, SBA-V18, SBA-O10 and an arithmetic-decoding question, closed by the 2026-10-15 readings.
- Label collisions: the analysis has a different 112-byte function labeled `CalcUnitBlockChance` at D2Game rva 0x21E10; the exported block function is D2Common rva 0x31D20 (ordinal 10212). Decompiler comments were not consulted in this revision; the earlier caveat that the comment "min level 2" contradicts the asm (`max(level, 1) * 2`) stands.
- The stat-list enumerator (ordinal 10702) returns entries from the list at `StatList +0x48` when `StatList +0x10` bit 31 is set and nothing otherwise (`RVA 0x386e0`, `RVA 0x382f0`); each entry is a dword key `(id << 16) | layer` and a dword value, delivered to callers as (`u16` layer, `i32` value) pairs of 8 bytes. The getter at rva 0x39430 (ordinal 10910) used for six stats was not read (the stat-getter question).
- Community formulas, 1.10f and D2MOO were not used; no 1.14d claim is made.

## Observations

Tool-neutral reproduction uses an owned matching LoD 1.13c game, single player Normal, Act I Blood Moor, with a level 1 Barbarian (Dex 20) attacked by Fallen as in the historical procedure above. Select a prepared character with the required stats; no save editing or action is implied by this record. Read the hit-roll return, roll and attacker/defender seeds at the original location listed in the historical procedure for 60 s of melee, and compare exact stats, source profile and trigger before interpreting missing calls. HIT-O01 was an actual historical observation only within its recorded case (ratio 47, level term 47, no defender step on a miss). Other hit-roll branches (caps, player attack-rating terms, defense adjustments and the post-hit request) remain unobserved under HIT-Q05; pending observations require a separately authorized gameplay campaign.

## Migration and review record

2026-10-15: existing file boundary preserved. Consumption search of implementation and plans outside gaps found no file/prefix consumer; historical aliases were removed, and unchanged rule/vector numbers retained for references. The complete originals/encodings/private provenance are preserved with an integrity manifest outside the repository. All 32 existing numeric vector results/seeds remain unchanged; four supplied-input vectors add Holy Shield, type-matrix, no-local-cap and missile/avoidance-only checks. No gameplay vector was newly observed. New evidence changes HIT-R01a/HIT-R02a/HIT-R01b/SBA-R01d/SBA-R04d to verified static at their stated supplied-input/request boundaries; it does not promote unresolved producers. Static old labels retain dated support and limits. Current separate blind-reader results will be recorded after that pass; historical review counts above are not a current review claim.

Writer self-review (2026-10-15): checked the current contract against the preserved numeric cases, input widths, caller argument order, request-only versus stored-effect boundaries, RNG ownership, public binary identities and the complete gap assessment. All 32 prior result/seed sets were preserved, including hexadecimal-to-decimal result-word presentation and the identical seed aliases. The four added vectors are supplied-input examples; none claims live gameplay confirmation. Every verified rule has a referenced executable vector. The final writer check was `python .agents/skills/spec-writing/tools/check_spec.py combat-shield-block-and-avoidance`: 0 errors, 0 warnings. No gameplay suite or implementation test ran. The separate reader pass remains required for review completion.

### Fresh blind reader, full pass 1 (2026-10-15)

The separate reader used only the current contract and cited dependencies, with no implementation, source-analysis or web access. Counts reported: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 5, GATED 15, COVERAGE 18, HARDEST 3. This is a full pass, not a delta check. The following outcomes address every reported entry. No reader result implies user approval or changes an evidence label.

| Finding | Outcome |
|---|---|
| DISCRIM: HIT-V08 missile mastery | Added c/d with a supplied nonzero 30 mastery result; missile skips the producer and retains AR 435, melee requests it once and obtains AR 536. |
| DISCRIM: HIT-V14 strict thresholds | Added exact AR 1048576 and pct 65536 boundary cases; direct amounts 346030/98304 distinguish premature divide-first amounts 346005/98250. |
| DISCRIM: HIT-V11 exclusion interval | Added immediately adjacent classes 703/710, which request application, contrasting retained excluded endpoints 704/709. |
| HARDEST: HIT-V08 producer order | Explicitly stated the already-recorded player call order: base, defense/flat adjustments, optional mastery and percent formation, percent scaling, negative swap. Added nonzero supplied mastery examples; actual producer remains gated. |
| COVERAGE: HIT-R01a large-base/large-modifier branches | Added HIT-V17 case a/b with exact divide-first final defense results. |
| COVERAGE: HIT-R01b ineligible branches | Added HIT-V18 case a-h, each returning contribution 0 with no evaluation request. |
| COVERAGE: HIT-R02a invalid class/helper input | Added HIT-V19 case a-c: omitted class term 315 versus fatal 4223. |
| COVERAGE: HIT-R02b mastery, false relation, non-monster | Added HIT-V08 case c/d and HIT-V16 case b/c. |
| COVERAGE: HIT-R02d nonpositive/negative defense | Added HIT-V12 case f-h: unchanged 500 for nonpositive stat and signed negative-defense result-250. |
| COVERAGE: HIT-R04 final negative clamp | Added HIT-V04 case e with minimum-i32 wrapping through both transfers, then final clamp; ratio 100 distinguishes a missing clamp. |
| COVERAGE: HIT-R06 zero denominator | Added isolated fault HIT-V20. |
| COVERAGE: HIT-R09a adjacent interval classes | Added HIT-V11 case d/e as above. |
| GATED: all 15 current questions | No gate was removed or default invented. The behavior, integration, observation and outside-boundary limits remain in the contract/gap file. |

The revised contract has 44 executable vectors. All 32 original result/seed sets remain intact; existing cases changed only by specified discriminating inputs with the same outputs, expanded traces or explicit rule attribution. No gameplay campaign or implementation test was run. A fresh full pass 2 is required because fixture inputs, outputs and acceptance coverage changed; the first reader's zero mismatch count does not complete review of this revision.

Writer recheck after pass 1 (2026-10-15): re-read D2Game RVA 0xDDE90 using its complete 560-byte instruction listing; disassembly was the available evidence form for this reading. Its calls to player base, defense/flat adjustment, optional selected-weapon/mastery, percent/type construction, ordered percentage helper and negative swap establish the clarified player sequence. The root-anchored function and its dependency boundaries remain those recorded above; no mastery producer is promoted. Recomputed every added case, including exact threshold amounts, minimum-i32 wrapping, final modifier branches, missile draw/flag composition and all new no-draw paths. Focused lint after the fixture and record edits returned 0 errors and 0 warnings. This recheck is writer self-review, not the required fresh reader pass 2.

### Fresh blind reader, full pass 2 (2026-10-15)

The second fresh full reader recomputed all 44 vector families from the public contract and
cited dependencies. Counts: **MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; GATED 68;
COVERAGE 10; HARDEST 3**. The gated list groups 15 questions, nine unresolved/asm-only
rules and 44 vector execution boundaries. These limits remain explicit; a supplied-input
case does not establish dependent integration or gameplay confirmation. There are no
mismatch, guess, gap or discrimination findings to resolve.

| Reader coverage entry | Outcome |
|---|---|
| HIT-R01a exact base 1048576 / modifier 65536 | No additional recipe: HIT-R03/HIT-V14 already distinguishes both strict arithmetic thresholds, and HIT-V07/HIT-V17 exercise both HIT-R01a arithmetic branches. Those exact HIT-R01a inputs are not claimed as an observed or executed case. |
| HIT-R01b qualifying skill types 1/3 | No change: these fixed bindings select the same contribution recipe as HIT-V15; HIT-V18 distinguishes the qualification failures. |
| HIT-R02d monsterDefenseHalf / hireling | No change: HIT-V12 distinguishes the halving result; repeating each qualifier does not add a result-distinct recipe. |
| HIT-R09a other post-hit qualifiers | No change: the same no-request result is exercised by HIT-V11's excluded classes; stat 117 defaults to zero in the other fixtures. |
| HIT-R02c plus HIT-R02d composition | No additional recipe: IgnoreDEF's zero remains zero through a positive fractional-defense adjustment, the same zero-result behavior already stated. |

| Reader hardest entry | Outcome |
|---|---|
| HIT-R01a/HIT-R03 ordered percentages and strict thresholds | No rule change: HIT-V14/HIT-V17 explicitly distinguish arithmetic order and threshold branches. |
| HIT-R04 minimum-i32 sequential transfers | No change: HIT-V04's wrapping case distinguishes both transfers and the final clamp. |

Root recomputation for SBA-V06d: seed `{5,666}` produces remainders 99, 73, 5 and final
seed `{0x134E1255,0x54152EE9}`; the third step was independently calculated with integer
arithmetic from the preceding seed `{0xC997B681,0x09201110}`. SBA-V20d's seed `{4,666}`
produces remainder 10 and the declared first-step seed, yielding code 16 / word 32768.
Both cases preserve all prior results and keep actual producer effects outside their
supplied-input boundaries. The contract still has 44 families. These new subcases require
fresh **full pass 3**; pass 2 does not complete review of the changed fixture set.

Public-provenance cleanup retains original binary references, dispatch chains, dated
observations, individual named cells, caveats and correction rationale. A duplicated old
contract and private listing-indexed notes were removed from this companion; the complete
pre-cleanup file and its SHA-256 remain in the private audit. Private listing indices and
capture-tool field names were replaced by their semantic source/capture boundaries. This
presentation cleanup supplies no new behavioral evidence and changes no label or gate.

### Fresh blind reader, full pass 3 (2026-10-15)

The third fresh full reader reported **MISMATCH 0; GUESS 1; GAP 0; DISCRIM 1;
GATED 15; COVERAGE 14; HARDEST 3** for the 44-family contract after pass 2.
The gated entries are the 15 current questions. Every reported entry has the
following outcome; no source reading, gameplay campaign, implementation test or
evidence-label promotion was performed during this triage.

| Reader finding | Outcome |
|---|---|
| COVERAGE: HIT-R01a null/zero/exact final thresholds | Added HIT-V17 case c for the null-unit memory fault and case d for zero total defense. No additional exact-threshold recipe: HIT-V14 already distinguishes the identical strict percentage thresholds, while HIT-V07/HIT-V17 exercise the final-defense direct/divide-first arithmetic. No exact-boundary HIT-R01a execution is claimed. |
| COVERAGE: HIT-R02b 128-entry limit | Added HIT-V16 case d with a supplied enumerator fixture offering 129 entries; the caller requests 128, the fixture returns the first 128, and typeBonus/pct are 128/158. Logged maximum and omitted layer 129 distinguish the consumer limit without claiming execution of the original enumerator. |
| COVERAGE: HIT-R02e non-player Stat 119 | Added HIT-V09 case b: raw AR 265 with percent 20 adds 53, yielding AR 318. Original zero-percent case a remains AR 265. |
| COVERAGE: HIT-R02g negative additions | Added HIT-V13 case f: supplied demon/undead bonuses -50/-40 add unchanged to AR 300, yielding 210. |
| COVERAGE: HIT-R09a no-hook qualifiers | No change: HIT-V11's excluded classes already exercise the no-request recipe; eligible class 5 with Stat 117 zero is supplied by HIT-V10 case a, and default monster-attacker hits also fail the player-attacker qualification. The single conjunction produces the same no-request result for each failed qualifier; isolated execution of every qualifier is not claimed. |
| COVERAGE: HIT-R02d boss/neutral-half predicates individually | No change: HIT-V12's player/hireling case already distinguishes the positive-value halving recipe (DEF 375 instead of 250). Boss and the supplied neutral half condition select that same arithmetic; their predicates remain explicit, with the neutral producer still under HIT-Q03. No isolated boss/neutral-half execution is claimed. |
| HARDEST: HIT-R02c/HIT-R02d arithmetic and eligibility | Retained HIT-V10's ignore-defense eligibility and HIT-V12's halving, cap, nonpositive and negative-defense recipes. The existing player call order and supplied neutral-condition boundary remain explicit; no rule or gate changed. |
| GATED: 15 questions | No change: all current behavior, integration, observation and outside-boundary questions remain open at their existing boundaries; supplied fixtures close none of them. |

The two maximum-count fixtures explicitly declare their own enumerator response
policy. They establish requested maxima and local consumption using supplied
results, not the original producer's execution or integration. The contract
continues to contain 44 vector families. All preceding numeric outputs and seeds,
including the root's pass-2 additions, remain unchanged; the added subcases have
independent inputs, seeds, ordered draws and exact outputs. Writer recomputation
gives equality code 0 / word 5 / seed B1 for each new comparison case, type bonus
128 / pct 158 for the maximum-128 case, failed weapon code 0 / v 1 / seed B1 for
the maximum-32 case, and word 261 for prior-flag preservation.

Pass 3 returned no mismatch and every entry has an explicit outcome. The edits
clarify equivalent wording and add fixtures only; no behavior, ownership or gate
changed. The next reader run may therefore be **delta pass 4** for SBA-R11 and the
added/reformatted cases: SBA-V06 e/f, SBA-V07 c, HIT-V09 a/b, SBA-V12 e, SBA-V15 e, HIT-V13 f, HIT-V16 d,
HIT-V17 c/d, SBA-V21 c/d/e and SBA-V24 c. It must recompute the changed/added cases and confirm
the GUESS and DISCRIM resolutions. This record does not claim that delta pass 4
has run or that user review is complete.

### Fresh blind reader, delta pass 4 (2026-10-15)

The fresh smaller reader checked the actual main file, numeric foundations, game-data
fields and required main dependency contracts. It read no companion, gap file, previous
report, private source, implementation, runtime or web material. The scope was SBA-R11 and
SBA-V06 e/f, SBA-V07 c, HIT-V09 a/b, SBA-V12 e, SBA-V15 a–e, HIT-V13 f, HIT-V16 d, HIT-V17 c/d, SBA-V21 c–e and SBA-V24 c. The equivalent
wording and fixture-only changes qualified for a delta check after full pass 3.

The initial report gave **MISMATCH 6; GUESS 0; GAP 0; DISCRIM 4**, all arising from its
incorrect decimal conversion/remainder for seed B. It claimed roll 54 rather than 10 in
SBA-V06 e/SBA-V07 c/SBA-V12 e/SBA-V15 a/SBA-V15 b/SBA-V24 c; its four discrimination claims concerned the three equality
cases and SBA-V15a's successful shield result. The coordinator independently checked the exact
integer step and requested a calculation audit without changing the spec.

For B, `4 * 0x6AC690C5 + 0x29A = 7165593006`; low32 is `0xAB1A45AE`, decimal 2870625710,
high is 1, and unsigned remainder modulo 100 is 10. The reader independently recomputed this
with exact integers, also confirming F's remainder 36, then recomputed all delta cases.
It explicitly retracted each of the six mismatch and four discrimination entries. The
equality cases do fail at 10<10; SBA-V15 a/b give words 28/16 and their stated seeds; SBA-V24 c gives
code 0 with v=1 after exactly 32 queried entries. No input, expected result, rule or evidence
label was altered to fit the erroneous calculation.

Corrected final counts: **MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0**. Every initial entry has
the arithmetic-error/retraction outcome above; the other cases also recompute to their
specified values, requests and seeds. Existing gates and advisory dispositions remain.
This completes the applicable reader check for this contract, without user approval,
gameplay confirmation or complete behavior reproduction.

### Specification review (2026-10-15)

An independent review recomputed every executable vector and the six seed aliases from
NRF-R01/NRF-R02 and found no numeric mismatch. It re-read the original disassembly for
D2Game 0xDB790, 0xDB030, 0xDE5A0, 0xDCE70, 0x3AE10, 0xDB1F0 and the cap/hook region of
0xDDE90, and D2Common 0x31D20 and 0x313F0: the movement penalty (effective 2/3/6, raw mode
not 2, signed division by 3), strict block/avoidance comparisons, defender draws and
assertion order, melee and missile code maps, the 16384 discard branch, the prevent-heal
request (classes 704..709 excluded, state 52, duration 120000, stat 31) and the defense
reduction branches agree with the current rules. Edits were presentation only: SBA-V03
case traces now state why modes 2 and 1 keep 75, negative values are spaced, and
implementation-gap rows cite current rule ids. No vector input, output, rule, label or gate
changed, so no reader rerun is required.

### Restructure (2026-10-15)

The approved review recommendation split this contract out of [combat-shield-block-and-avoidance](../../specs/combat-shield-block-and-avoidance.md) at the attack hit roll. This spec now owns contact defense, total defense and Holy Shield, attack rating and its adjustments, the percent step, negative swap, ratio, level term, cap, attacker draw, the post-hit prevent-healing request and attacker type differences. The shield-block spec keeps whether the roll runs and with which arguments, shield block, avoidance, result codes/flags, missile flags, preconditions and the attack's complete draw order. No plan had consumed the source spec, so both files renumbered cleanly. No rule behavior, label, value, vector input, seed or output changed. The former type-differences rule was divided into an attacker rule here (HIT-R10) and a defender rule there (SBA-R09a). The skip rule's ignore-defense sentence became a stated consequence in HIT-R02c, while the skips themselves stay in SBA-R07a; HIT-V10 now cites HIT-R08 instead of the skip rule, and HIT-V08/HIT-V09 also cite HIT-R10. The former single gameplay-coverage question was divided into HIT-Q05 here and SBA-Q06 there; HIT-Q05 records that only the ratio/level cases were observed. The shield producer for Holy Shield stays SBA-Q10 and other hit-roll callers stay SBA-Q08. The evidence, dispatch chains, function, runtime, ledger and caveat records and the review records above were copied from the source companion with ids mapped to the current numbering; table rows that concern only shield block or avoidance stay in that companion. Review and reader counts above describe the combined 44-family contract before this restructure; they are not a review of this file. Because ownership changed, a fresh full reader pass is required for this contract.
