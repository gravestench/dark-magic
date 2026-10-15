# Evidence: shield block and avoidance

Spec: [Shield block and avoidance](../../specs/combat-shield-block-and-avoidance.md)

## Binary identity and method

The [source manifest](../source-manifest.md) identifies `LOD113C-D2GAME`, `LOD113C-D2COMMON` and `LOD113C-FOG`, original LoD 1.13c Windows x86. Original-image rehash 2026-10-15 matched D2Common SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`, size 679936. Earlier static records are reused within their dated boundaries; matching current specimens does not prove historical loaded-DLL identity. Historical data profiles lack complete archive/override hashes.

Descriptions are analyst labels. Original binary locations, calling conventions and layouts below are investigation evidence, not required implementation layouts. Detailed private source matching records, commands, session addresses and original fixture encodings are archived privately with the complete original files and integrity manifest.

## New readings and corrections (2026-10-15)

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2COMMON` | 0x32890 | 10672 | Total defense | SBA-R01a/SBA-R01b | Disassembly, all 896 bytes decoded against matching original image; direct arithmetic and Holy Shield request, dependency results supplied |
| `LOD113C-D2COMMON` | 0x31EA0 | 10621 | Player attack base | SBA-R02a | Disassembly checked against original 240-byte extent |
| `LOD113C-D2COMMON` | 0x39430 | 10910 | Parameterized selected stat read | SBA-R02b/c/d/e/g/SBA-R18a | Full disassembly; forwards bounded definition and layer to 0x38A80, owned NRF-R12b/R12c |
| `LOD113C-D2COMMON` | 0x1250 | none | Skill record lookup | SBA-R01b | Reconstruction checked against original machine code; signed bounds then row selection |
| `LOD113C-D2COMMON` | 0x51BF0 | none | Expression front-end | SBA-R01b | Reconstruction checked against original machine code; context(unit, skill, level), SCX-R11 interpreter boundary |
| `LOD113C-D2COMMON` | 0x4F640 | 10804 | Mastery consumer | SBA-R02f | Targeted full listing; special 0x4F200 and weapon selection remain unresolved |
| `LOD113C-D2COMMON` | 0x31680 | none | Special monster eligibility | SBA-R09c | Full listing inspected; table selection 0x84C70 and definition-name lookup 0x9240 remain unresolved |
| `LOD113C-D2COMMON` | 0x43860 | none | Equipment/mode classification | SBA-R12e | Full listing inspected; numerous equipment/item helpers remain an explicit unresolved producer boundary |
| `LOD113C-D2COMMON` | 0x1FF40 / 0x23EF0 / 0x24430 | none / none /10744 | Shield candidates/item predicates | SBA-R09e/SBA-R12b | Targeted full listings; slot/exclusion/type data integration remains unresolved |
| `LOD113C-D2GAME` | 0xA00B0 | none | State applier | SBA-R18b | Full 704-byte disassembly read; delegates attach/write/free effects to STL primitives with current gates |
| `LOD113C-D2GAME` | 0x372C0 | none | Selected weapon | SBA-R02f | Full 368-byte listing read; imported slot/equipment producers not accepted as defaults |
| `LOD113C-D2GAME` | 0xE0A20 | none | Linked owner wrapper | SBA-R22 | Full 64-byte disassembly; valid-link test then 0xE03A0; table lookup remains dependency boundary |

SBA-R01a/R02a arithmetic gates are closed by original-image opcode checks, not by recency or a new label. In total defense, 0x3298C bytes 8D3C28 add the two defense percentages; 0x32ADD bytes 0FAFFE multiply percent by base. At 0x31F42 bytes 8D5430F9 add tohit+dexterity-7, then the next instruction adds four Dexterities-28: the result is tohit+5*dexterity-35. Signed division/magic sequences were followed to returns. The total-defense negative branch divides its product by -100; subtraction of product/100 is the same signed result for the supported arithmetic.

Correction: the earlier statement that null total-defense input safely returns 0 was wrong. The initial null branch reaches 0x3290E, then 0x32919/0x3291D unconditionally reload/dereference the unit. The revised null failure claim does not change any existing numeric vector. The Holy Shield skill lookup call at 0x32A95 targets 0x1250, not the earlier 0x2250 locator. At 0x32ACC evaluation calls 0x51BF0; matched front-end confirms SCX context and no direct random draw. Expression-specific producers/draws remain owned by SCX, not erased by supplying a result. State-list reads 350/351 call 0x38B00, the own-array operation under NRF-R12c; ordinary unit readers use 0x38A80, selected-array operation.

The stat-getter question is closed: 0x39430 rejects invalid signed stat ids, missing list or missing definition; normalizes layer to `u16` and forwards to the same selected-array 0x38A80 that ordinary reads use. It dereferences its unit before list checks, so no null-unit default is established for that wrapper. Local callers use non-null units.

SBA-R09d reuses NRF's game.expansion semantic input and the previously complete caller-argument trace; the old blanket selector gate is removed. Classic gameplay remains unobserved. SBA-R12d now states the previously read consumer behavior (no cap); its upstream production question remains outside-boundary. The newly added supplied 101/roll 99 example distinguishes a local 95 cap without claiming any live character has that supplied value.

SBA-R18b remains asm-only: reading the complete applier resolves its request skeleton but does not establish all indirect writes. Compatible STL primitive contracts support the state-effect boundary; CSA-R05 is a cold-applier pattern, not the owner of this generic applier. The local hook establishes supplied request assertions; attach/replacing/free propagation gates remain. The earlier timed-list description is retained below as historical evidence, not promoted to verified stored effects.

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
3. Rva 0x92A60 (`RVA 0x92a60`): `eax` = defender, `edi` = attacker, stack `(game, bonus, 0)`, `bonus` = the skill's to-hit term plus stat 325 (SLD-R13; an input here) -> **rva 0xDE5A0** (result word; SBA-R13, SBA-R14, SBA-R21).
4. Rva 0xDE5A0 -> hostility gate rva 0xE0A60 (`RVA 0xde5a0`; SBA-R21) -> range test D2Common ordinal 11138, rva 0x326C0 (SBA-R21) -> **hit roll rva 0xDDE90** (SBA-R01 to SBA-R08) -> **block and avoidance roll rva 0xDB790** (SBA-R09 to SBA-R12) -> state test ordinal 10494, rva 0x33CD0 (SBA-R13).
5. Rva 0xDDE90 -> total defense ordinal 10672, rva 0x32890 (`RVA 0xdde90`; SBA-R01); player attack-rating base ordinal 10621, rva 0x31EA0 (SBA-R02a); defense/attack-rating adjustment rva 0xDB1F0 (SBA-R02c, R02d, R02g); melee-only term rva 0x372C0 and ordinal 10804, rva 0x4F640 (SBA-R02f); monster-type matrix test rva 0x1620 (SBA-R02b); percent helper rva 0x14D0 (SBA-R03); post-hit hook rva 0xDCE70 (SBA-R18).
6. Rva 0xDB790 -> block chance ordinal 10212, rva 0x31D20 (`RVA 0xdb790`; SBA-R09) -> moving test ordinal 10143, rva 0x313F0 (SBA-R10) -> **avoidance rva 0xDB030** (SBA-R12) -> weapon-avoid list rva 0xDA540 (`RVA 0xdb030`), token helper ordinal 10431, rva 0x441E0 , draw helper rva 0x1180 , attacker assert rva 0xDF780 (reconstruction checked against original bytes).

Twenty-nine other D2Game functions also call rva 0xDE5A0 (original D2Game RVAs `0x273c0`, `0x27820`, `0x27a10`, `0x289a0`, `0x28be0`, `0x44c20`, `0x44d70`, `0x451b0`, `0x4c2a0`, `0x4c720`, `0x4cb00`, `0x4cc50`, `0x4dac0`, `0x4f7d0`, `0x77240`, `0x77790`, `0x92940`, `0x92bb0`, `0x93040`, `0x93680`, `0x95690`, `0x98ba0`, `0x98e90`, `0x99d20`, `0xa1170`, `0xa11b0`, `0xa2c40`, `0xb41b0`, `0xb60c0`); each passes its own `(game, bonus, range)` and was not walked to its root (SBA-Q22). The monster melee attack observed at runtime (SBA-O07) entered rva 0xDDE90 from inside rva 0xDE5A0 (return address 0xDE646); its caller above 0xDE5A0 was not traced.

**Dispatch chain B: missile hitting a unit.**

1. Frame loop -> missile step rva 0x3F320 ([missile-movement-and-collision](../../specs/missile-movement-and-collision.md) MMC-R03) -> contact rva 0x3EDE0 (MMC-R11, MMC-R14; [missile-hit-and-damage-calculation](../../specs/missile-hit-and-damage-calculation.md) MHD-R13).
2. Rva 0x3EDE0 (`RVA 0x3ede0`): if the target is non-null and `MissilesRecord.ToHit (+0x18B)` is non-zero, **hit roll rva 0xDDE90** with `eax` = missile owner, stack `(target, Stat(missile, 19 tohit), isMissile = 1)`; a zero `ToHit` skips the roll and continues as a hit (SBA-R15a).
3. On a hit: damage fill rva 0x3A4E0 -> hand-off rva 0x3AFF0 (MHD-R25a, MHD-R25b) -> **apply rva 0x3AE10** (`RVA 0x3aff0`; MHD-R26 gives the argument roles `(game, missile, target)` with the damage record in `edi`).
4. Rva 0x3AE10 -> owner lookup rva 0xE03A0 (`RVA 0x3ae10`; SBA-R22) -> state test ordinal 10494  -> knockback draw rva 0x1180 (SBA-R20) -> **block and avoidance roll rva 0xDB790**  with `eax` = (`DamageRecord +0x08` != 0), `ecx` = owner, `esi` = target, stack `(game, 1)` -> own code-to-flag map (SBA-R14).

Rva 0xDE380 also calls rva 0x3AE10 (`RVA 0xde380`) and rva 0x3B040 also calls rva 0xDDE90; neither was walked (SBA-Q22).

**Dispatch chain C: avoidance only.** Rva 0xA10B0 (`RVA 0xa10b0`) calls **avoidance rva 0xDB030** with `eax` = defender, stack `(nDamageParam = 1)`, `edi` = the owner resolved by rva 0xE0A20 (not read; it supplies the attacker for the R19 assert; SBA-Q22), only when the result word `DamageRecord +0x04` is non-zero; it never rolls shield block and maps codes itself (SBA-R14). It is reached from rva 0x92BB0 and from the area callback rva 0xA1210 (flag-1 apply route, [combat-damage-pipeline-order](../../specs/combat-damage-pipeline-order.md) DPO-R08d); their roots were not walked (SBA-Q22).

**When it runs.** Once per attack contact, inside the attacker's skill do-function at its do frame (chain A) or inside the missile's movement step on the frame of contact (chain B), on the server; before the damage record is built and applied (the result word this spec produces is the input of [combat-damage-pipeline-order](../../specs/combat-damage-pipeline-order.md)). No timer event and no per-frame cadence belong to this spec; the only event scheduled is the timed-state expiry posted by the post-hit hook's applier (SBA-R18b).

**Client side.** None of the cited functions sends a message. The client learns the outcome through the unit-update and damage messages the damage pipeline sends after apply (names belong to a `net-` spec, none written); a block is visible as the defender's block mode (player mode 9 `Block`, monster mode 6 `block`, GDF-R14, GDF-R15), set by [combat-hit-recovery](../../specs/combat-hit-recovery.md).


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

Priorities: run vs walk block (`/3` only on Run and Town Walk), Dex < 15 (no block), player running (hit roll skipped), moving defender with Dodge but no Evade, a pure-elemental missile against a shield. Common setup for an owned 1.13c LoD game, single player, Normal difficulty: a level-1 character in Act I (Blood Moor, level id 2), attacked by Fallen (`fallen1`, monster level 1); read the defender's `Unit +0x20`/`+0x24` before each hit roll and the return values of the cited functions (SBA-O07 to SBA-O10 were reproduced this way).

- **SBA-O01** (formerly B1/B2; R09a, R10). Measure: block rate while walking vs running, same Barbarian with a Buckler (`Armor.block (Buckler) = 0`, `CharStats.BlockFactor (Barbarian) = 25`, Dex 20, level 1: chance 62 standing, 20 running). Procedure: stand among Fallen walking in place (mode 2) for 60 hits, then run back and forth (mode 3) for 60 hits; read the block-chance return (62 vs the block roll's effective 20) or count blocks. Expected: walking blocks about 62 of 100, running about 20 of 100; the block-chance function returns 62 in both cases (the `/3` is applied by the caller).
- **SBA-O02** (formerly B4; R09a). Measure: block rate with Dex < 15. Procedure: a character with Dex 10 (a fresh save edited to Dex 10, never a played save) carrying a shield; take 60 hits standing. Expected: no block ever; the block-chance function returns a negative value (`(10 - 15) * T / (2 * level)`).
- **SBA-O03** (formerly O1; R11, R12c, R13). Measure: standing melee, block then Dodge. Procedure: an Amazon with Dodge (stat 338 > 0) and a shield, standing; read the defender seed before each hit roll and the hit-roll return. Expected: after each hit, the defender's seed steps once on a block (code 1, mode 9 `Block` follows) and twice when the block fails and Dodge is rolled; never a Dodge animation on a blocked hit.
- **SBA-O04** (formerly O2; R12a). Measure: moving defender with Dodge but no Evade. Procedure: an Amazon with Dodge > 0 and Evade 0, no shield, walking while hit. Expected: no avoidance at all (no draw, no Dodge animation).
- **SBA-O05** (formerly O3; R12a, R14). Measure: moving missile defender with Avoid only. Procedure: an Amazon with Avoid > 0 and Evade 0, walking through Fallen Shaman fire bolts or Dark Ranger arrows. Expected: no avoidance.
- **SBA-O06** (formerly standing missile; R12c, R14). Measure: Avoid after a failed shield block. Procedure: the same Amazon with a shield, standing, hit by arrows (physical missiles). Expected: shield block first, then Avoid when the block failed.
- **SBA-O07** (R05, R06; R01a and R02e consistent). Observed 2026-10-12 (Appendix A): Fallen attacker, `AR = (101 * 8) / 100 = 8` (`MonStats.A1TH (fallen1) = 101`, `MonLvl.L-TH (level 1) = 8`), `bonus` 0, `isMissile` 0, `aLvl` 1; Barbarian defender, `DEF = 4 + 20 / 4 = 9` (stat 31 = 4, Dex 20, pct 0, stat 33 = 0; inferred from the chance), `dLvl` 1: ratio = 800 / 17 = 47; chance = (47 * 1 * 2) / 2 = 47. Reading: 116 hit-roll calls, every hit with roll <= 46 and every miss with roll >= 47 (roll = `edx` at return); the defender's seed never changed across a miss. Procedure to repeat: trace rva 0xDDE90 with a deref of the first stack argument (the defender) for 60 s of melee.
- **SBA-O08** (R09a, R09d). Observed 2026-10-12: rva 0x31D20 called 57 times with `applyScaling` 1, returning 62 = `((20 - 15) * (0 + 25)) / (1 * 2)` each time. Procedure: trace D2Common rva 0x31D20 during the same fight.
- **SBA-O09** (R11; R13 consistent). Observed 2026-10-12: 34 blocked hits each made exactly 1 defender step with roll 2..60; 27 unblocked hits had a first roll of 62..99 (a roll of 62 did not block: strict `<`); the block-chance function was called only after hits. Procedure: as SBA-O07 with the seed deref; count steps between consecutive hit rolls.
- **SBA-O10** (R16; DUR-R04, DUR-R05). Observed 2026-10-12: after every failed block the defender's seed moved 3 more steps (4 in total) before the next hit roll; the defender had no avoidance stats. Consistent with the durability step; not traced (SBA-Q20). Procedure: as SBA-O09; to settle, also trace rva 0x12EE0.
- **SBA-O11** (new; R14, SBA-Q23). Measure: shield block against a missile with zero physical damage. Procedure: a shield-bearing character standing in Fallen Shaman fire bolts (elemental only) for 60 impacts; trace D2Common rva 0x31D20. Expected: rva 0x31D20 is never called from rva 0xDB790 during those impacts (block-enable 0), while arrows from Dark Rangers do call it.


## Appendix A. Provenance

Historical readings: 2026-10-11, 2026-10-12 and 2026-10-13. Evidence is original LoD 1.13c Windows x86, read by disassembly or reconstruction checked against original bytes. Analyst labels are not original exported names.

**Function table.**

| Original module location | Form | Label | RVA | Ordinal | Rules |
|-------------|------|-------|-----|---------|-------|
| `D2Game.dll+0xdde90` | asm | `ProcessMonsterSkillRoll` (wrong) | 0xDDE90 | n/a | R01-R08, R15a, R16, R18 |
| `D2Game.dll+0xde5a0` | asm | `GetUnitCombatFlags` | 0xDE5A0 | n/a | R13, R14, R15a, R21 |
| `D2Game.dll+0xdb790` | asm | `RollBlockChance` | 0xDB790 | n/a | R09d, R10, R11, R16 |
| `D2Game.dll+0xdb030` | asm | `CalculateRandomChanceResult` | 0xDB030 | n/a | R12, R16, R19 |
| `D2Game.dll+0xda540` | asm | `ApplyItemProperty` (wrong) | 0xDA540 | n/a | R12b |
| `D2Game.dll+0xdb1f0` | asm | `ProcessMonsterBehavior_8b30` (wrong) | 0xDB1F0 | n/a | R02c, R02d, R02g |
| `D2Game.dll+0x01620` | asm | `CheckSkillTableBit` | 0x1620 | n/a | R02b |
| `D2Game.dll+0x01220` | C | `GetMonsterRecord` | 0x1220 | n/a | R02b |
| `D2Game.dll+0x014d0` (read in NRF-R05) | asm | `CalculateScaledDivision` | 0x14D0 | n/a | R03 |
| `D2Game.dll+0x01180` (read in NRF-R02) | asm | `GetRandomModulo` | 0x1180 | n/a | R12b, R12c, R20 |
| `D2Game.dll+0x22a80` | asm | `CheckPlayerDataFlags` (wrong) | 0x22A80 | n/a | R02d |
| `D2Game.dll+0x372c0` | asm | `GetUnitPrimaryWeapon` | 0x372C0 | n/a | R02f |
| `D2Game.dll+0xdce70` | asm | `ProcessSkillAction_af90` (wrong) | 0xDCE70 | n/a | R18a |
| `D2Game.dll+0xa00b0` (historical initial-eligibility slice) | asm | `CreateSkillPath` (wrong) | 0xA00B0 | n/a | R18b |
| `D2Game.dll+0xdf780` | C | `GetUnitField80_df780` | 0xDF780 | n/a | R19 |
| `D2Game.dll+0x3ae10` | asm | `ProcessObjectTargetEffect` | 0x3AE10 | n/a | R14, R20, R22 |
| `D2Game.dll+0x3ede0` (historical contact/hit-call slice) | asm | `ProcessObjectInteraction` | 0x3EDE0 | n/a | section 2, R15a |
| `D2Game.dll+0x3aff0` (historical apply-handoff call slice) | asm | `ProcessObjectTargetWithStats` | 0x3AFF0 | n/a | section 2 |
| `D2Game.dll+0xa10b0` (historical avoidance call/map slice) | asm | `ProcessMissileHitEffects` | 0xA10B0 | n/a | section 2, R14 |
| `D2Game.dll+0xe03a0` (`db`-run listing) | asm | `FindUnitByTypeAndId` | 0xE03A0 | n/a | R21, R22 |
| `D2Game.dll+0xe0a60` | asm | `GetUnitOwner` | 0xE0A60 | n/a | R21 |
| `D2Game.dll+0xe0a20` (not read) | asm | chain C owner resolver | 0xE0A20 | n/a | section 2, R22 |
| `D2Game.dll+0xf02a0` (not read here; per NRF 3.3) | asm | relation-flag test | 0xF02A0 | n/a | R21 |
| `D2Game.dll+0xa19a0` (historical action-table call slice) | asm | `ExecuteSkillWithMissiles` | 0xA19A0 | n/a | section 2 |
| `D2Game.dll+0x93270` (historical target/hit-calculator call slice) | asm | `ProcessSkillHitDamage_1480` | 0x93270 | n/a | section 2 |
| `D2Game.dll+0x92a60` (historical hit-result call slice) | asm | `CalculateFullSkillDamage` | 0x92A60 | n/a | section 2 |
| `D2Game.dll+0xde380` (historical alternative apply-call slice) | asm | caller of 0x3AE10 | 0xDE380 | n/a | section 2 |
| `D2Common.dll+0x32890` | asm | `CalcTotalDefenseRating` | 0x32890 | 10672 (thunk 0xA832, slot 0x6FD18560) | R01 |
| `D2Common.dll+0x31ea0` | asm | `CalcCombatStatBonus` | 0x31EA0 | 10621 (thunk 0xA7F0, slot 0x6FD18534) | R02a |
| `D2Common.dll+0x4f640` | asm | `CleanupStatListModChain` (wrong) | 0x4F640 | 10804 (thunk 0xA80E, slot 0x6FD18548) | R02f |
| `D2Common.dll+0x386e0`, `RVA 0x382f0` | asm | `GetStatFromListByLayerChecked`, `STATS_CopyStatsByType` | 0x386E0, 0x382F0 | 10702 (thunk 0xA7D8, slot 0x6FD18524) | R02b, R02f, R12b |
| `D2Common.dll+0x38b70` (NRF-R12b) | asm | `GetUnitBaseStat` | 0x38B70 | 10973 (thunk 0xA42A, slot 0x6FD182B0) | stat reads |
| `D2Common.dll+0x39430` (not read) | asm | `GetMissileRecordField54` (wrong) | 0x39430 | 10910 (thunk 0xA802, slot 0x6FD18540) | stat reads 115, 116, 117, 119, 123, 124 (the stat-getter question) |
| `D2Common.dll+0x1a550` | asm | `FreeUnitStatList` (wrong) | 0x1A550 | 10064 (thunk 0xA580, slot 0x6FD18394) | R02c, R02d |
| `D2Common.dll+0x1a2d0` | asm | `SetStructField4C` (wrong) | 0x1A2D0 | 11104 (thunk 0xA7F6, slot 0x6FD18538) | R02c, R02d |
| `D2Common.dll+0x1a610` | asm | `GetUnitSeedHi` (wrong) | 0x1A610 | 10255 (thunk 0xA814, slot 0x6FD1854C) | R02g |
| `D2Common.dll+0x1a5b0` | asm | `SyncUnitPositionFromPath` (wrong) | 0x1A5B0 | 10239 (thunk 0xA508, slot 0x6FD18344) | R02g |
| `D2Common.dll+0x31d20` | asm | `CalcUnitBlockChance` | 0x31D20 | 10212 (thunk 0xA7DE, slot 0x6FD18528) | R09 |
| `D2Common.dll+0x31680` | asm | `ITEMS_IsSpecialWeaponType` (wrong) | 0x31680 | n/a (D2Common-internal) | R09c |
| `D2Common.dll+0x1ff40` | asm | `FindEquippedType33Item` | 0x1FF40 | n/a (internal) | R09a, R09e, R01b |
| `D2Common.dll+0x1e1b0`, `RVA 0x23ef0` | asm | `INV_GetOrCreateGridSlot`, `IsItemNotRestricted` (`db` listing) | 0x1E1B0, 0x23EF0 | n/a (internal) | R09e |
| `D2Common.dll+0x24430` | asm | `GetUnitDirection` (wrong; item-type equivalence test) | 0x24430 | 10744 (thunk 0xA430, slot 0x6FD182B4) | R09e, R12b, R02f |
| `D2Common.dll+0x1e670` | asm | `AllocateFieldTableBuffer` (wrong; equipped item by body slot) | 0x1E670 | 11139 (thunk 0xA7AE, slot 0x6FD18508) | R12b |
| `D2Common.dll+0x313f0` | asm | `GetEntryValueById` (wrong; moving-mode test) | 0x313F0 | 10143 (thunk 0xA772, slot 0x6FD184E0) | R10 |
| `D2Common.dll+0x441e0` | asm | `GetUnitWeaponAnimToken` | 0x441E0 | 10431 (thunk 0xA838, slot 0x6FD18564) | R12b, R12e |
| `D2Common.dll+0x33cd0`, `RVA 0x37db0` (read in GDF-R11) | asm | state test wrapper, `STATS_TestStatFlag` | 0x33CD0, 0x37DB0 | 10494 (thunk 0xA68E, slot 0x6FD18448) | R13, R20 |
| `D2Common.dll+0x326c0` | asm | `IsUnitInAttackRange` | 0x326C0 | 11138 (thunk 0xA754, slot 0x6FD184CC) | R21 |
| `D2Common.dll+0x325c0` (`db` prologue) | asm | `GetUnitWeaponStyle` | 0x325C0 | 11133 (thunk 0xA77E, slot 0x6FD184E8) | R21 |
| `D2Common.dll+0x84d30` (not read) | asm | `AreUnitsInSameModifierTier` | 0x84D30 | 10666 (thunk 0xA6F4, slot 0x6FD1848C) | R21 |
| `D2Common.dll+0x348f0` | C | `GetPathRoom` (label; returns `arg +0x58`) | 0x348F0 | 10392 (thunk 0xA598, slot 0x6FD183A4) | section 2 |
| `D2Common.dll+0x51bf0` (not read; cited as the calc evaluator) | C | skill calc evaluator | 0x51BF0 | n/a (internal) | R01b |
| `Fog.dll+0x0fc20`, `RVA 0x11100` | C | `GetReturnAddress`, `CleanupAndAbort` | 0xFC20, 0x11100 | 10265 (thunk 0xA2DA), 10024 (thunk 0xA2C2) | fatal error 0xA2A path (R11, R12a, R18b) |
| D2Game do-function table (per SCD section 2) | data | `.data + 0x84A8` | VA 0x6FD274A8 | n/a | section 2 |

**Search records.** No rule is labeled `not found in the original-image analysis` in this revision. The three former `not found` rules were relabeled `unresolved` because their records were incomplete:
- SBA-R12d (avoidance cap): rung 1 root anchor: bodies of rva 0xDB030, 0xDB790, 0xDA540 (no clamp); rung 3 data anchor: the stat-identity search in D2Game RVAs 0xDB030 and 0xDA540 found only the stat-read pushes, no compare immediates near them. Rungs 2, 4, 5, 6 not run.
- SBA-R15b (never-miss flags): rung 1: no `SkillsRecord` read in rva 0xDDE90, 0xDE5A0, 0xDB790; the only skips are R15a. Rungs 2-6 not run.
- SBA-R17b (PvP chance): rung 1: no unit-type-0-vs-0 branch in rva 0xDDE90, 0xDB1F0, 0xDB790, 0xDB030 beyond R02c, R02d. Rungs 2-6 not run.

**Runtime records.** Member historical vanilla LoD profile, Lord of Destruction 1.13c, single player, Normal. Traces of 2026-10-12 were armed for 60 s with a cap of 200 records and both ended on the time limit; the 2026-10-13 row is a read of the records still held in the member's disarmed slot.

| Vector | Version | Member | Date | Trace RVA | Records | Result |
|--------|---------|--------|------|-----------|---------|--------|
| SBA-O07 | 1.13c | historical vanilla LoD profile | 2026-10-12 | D2Game+0xDDE90 (48-byte defender-argument read) | 116 (61 hits, 55 misses), caller D2Game+0xDE646 only, stack cleanup 12 bytes | match: R05, R06; 0 defender draws in the hit roll |
| SBA-O08 | 1.13c | historical vanilla LoD profile | 2026-10-12 | D2Common+0x31D20 | 57, caller D2Game+0xDB7B8 only | match: R09a returns 62 in every call with `applyScaling` 1 |
| SBA-O09, SBA-O10 | 1.13c | historical vanilla LoD profile | 2026-10-12 | D2Game+0xDDE90 (seed deltas between calls) | 61 transitions, each 1 or 4 steps | match: R11 (1 step on a block, strict `<`); SBA-O10 consistent with DUR-R04/R05 |
| SBA-O07, SBA-O09 (re-read) | 1.13c | historical vanilla LoD profile | 2026-10-13 | D2Game+0xDDE90 (disarmed, 116 records retained) | 12 records re-read: rolls 60, 41, 97, 7, 14, 60, 24, 85, 2, 73, 9, 85 with hits exactly where roll < 47; stack cleanup 12 bytes; caller 0xDE646; the defender seed went from `{0xBFC793C3, 0x29A}` before the hit of record 1 to `{0xDE8467A9, 0x4FFD63FE}` before record 2, which is exactly 1 LCG step (a blocked hit) | match |
| SBA-O01 to SBA-O06, SBA-O11 | 1.13c | historical vanilla LoD profile | 2026-10-13 | none armed | 0 | not attempted: the member's character (level-1 Barbarian) was dead (mode 17) in the Blood Moor with four neutral Fallen beside it, and this revision was not permitted to drive, launch or edit saves; the entry trace at rva 0xDB790 was not supported by the available instrumentation and was not enabled (SBA-Q19) |

Mismatches: none. Calling conventions measured: rva 0xDDE90 `ret 0Ch` with the attacker in `eax` and stack `(defender, bonus, isMissile)`; rva 0x31D20 `ret 8` with stack `(unit, applyScaling)`.

**Blind-implementer pass.** Run 1 (2026-10-13): GUESS 8, GAP 9; every item folded (section 3 rows for `Unit +0x04`, `StatList +0x48`, the item-type test, body slots 4 and 5 and the entry value of `DamageRecord +0x04`; R09a, R10, R11, R12, R14, R15a, R20, R22; section 5 failure rows; SBA-V16, SBA-V19, SBA-V25; SBA-Q16, SBA-Q26). Run 2 (2026-10-13, fresh reader): GUESS 6, GAP 5; every item folded (section 3 rows for the unit type and the attacker pointer; R02d null `MonsterData`; R09a gate order; R09b getter identity; R12b token-call order; SBA-V03, SBA-V14, SBA-V18, SBA-V19, SBA-V30 inputs). No third full review was completed for that revision.

**Per-claim ledger.**
- runtime-confirmed (1.13c): SBA-R05, SBA-R06, SBA-R09a (scaling branch), SBA-R11; the calling conventions above.
- verified in 1.13c code (static): SBA-R02b, R02c, R02d, R02e, R02g, R03, R04, R07, R08, R09b, R10, R12a, R12b, R12c, R13, R14, R15a, R16, R17a, R18a, R19, R20; stat ids 2, 12, 19, 20, 31, 32, 33, 115, 116, 117, 119, 123, 124, 171, 16, 182, 179, 338, 339, 340, 348 as read from the getter arguments; item type 0x33; offsets +0x3C, +0x49, +0x18B, +0xAE, +0x1C (MonStats), +0x16 (MonsterData), +0x08 (damage record, role per NRF), +0x70 (game record).
- asm-only: SBA-R01b, R09d (role of `Game +0x70`), R09e, R18b, R21, R22.
- unresolved: SBA-R01a, R02a (one-instruction `db` runs, the arithmetic-decoding question); R02f, R09c, R12d, R12e, R15b, R17b.
- names: from core-game-data-fields, plus six names read from the 1.13c `ItemStatCost.txt` and `States.txt` through the member's own file loader on 2026-10-13 (rows 117, 123, 124, 342, 350, 351; state row 52).

**Caveats.**
- `db` runs read as instructions: `RVA 0x32890` (`8D 3C 28` = `lea edi,[eax+ebp]`, the `pct` sum) and (`0F AF FE` = `imul edi,esi`, the product) on the R01a path; `RVA 0x31ea0` (`8D 54 30 F9` = `lea edx,[eax+esi-7]`) on the R02a path; `RVA 0x326c0` (`8D 5C 10 01` = `lea ebx,[eax+edx+1]`) in R21; `RVA 0x24430` (`85 0C 16` = `test [esi+edx],ecx`; `8B 14 0E` = `mov edx,[esi+ecx]`) in the item-type test; the embedded table region of `RVA 0x313f0` is jump-table data (the player path is assembled); `RVA 0x4f640` (R02f loop); `RVA 0x31680` (R09c); `RVA 0x32890` (R01b); `RVA 0xe03a0`, `RVA 0x23ef0` and the prologue of `RVA 0x325c0` are whole `db` listings with the tool's decoding comments.
- Label corrections to earlier drafts: the missile owner resolver is rva 0xE03A0 (corrected from a section-relative analyst locator), not 0xDF3A0; the relation-flag helper is rva 0xF02A0 (corrected from a section-relative analyst locator), not 0xEF2A0; "attacker present" as the first argument of rva 0xDB790 on the missile path was wrong: it is `DamageRecord +0x08 != 0` (block-enable); `applyScaling` is `Game +0x70` of the game record passed down both chains, which closes the "if `arg1` is the game record" condition of the expansion-selector question.
- Status changes in this revision: SBA-R01a and SBA-R02a `verified` -> `unresolved` (assumption 3, the arithmetic-decoding question); SBA-R12d, SBA-R15b, SBA-R17b `not found in the original-image analysis` -> `unresolved` (no six-rung record existed); SBA-R18 split into R18a (verified) and R18b (asm-only) because the applier was not read to its writes; SBA-R12b's text corrected (a layer-0 entry assigns rather than maxes; layers are `u16`). New ids: SBA-R09e, SBA-R21, SBA-R22, SBA-V31, SBA-V32, SBA-O11, the arithmetic-decoding question to SBA-Q26.
- Label collisions: the analysis has a different 112-byte function labeled `CalcUnitBlockChance` at D2Game rva 0x21E10; the exported block function is D2Common rva 0x31D20 (ordinal 10212). Decompiler comments were not consulted in this revision; the earlier caveat that the comment "min level 2" contradicts the asm (`max(level, 1) * 2`) stands.
- The stat-list enumerator (ordinal 10702) returns entries from the list at `StatList +0x48` when `StatList +0x10` bit 31 is set and nothing otherwise (`RVA 0x386e0`, `RVA 0x382f0`); each entry is a dword key `(id << 16) | layer` and a dword value, delivered to callers as (`u16` layer, `i32` value) pairs of 8 bytes. The getter at rva 0x39430 (ordinal 10910) used for six stats was not read (the stat-getter question).
- Community formulas, 1.10f and D2MOO were not used; no 1.14d claim is made.

## Observations

Tool-neutral reproduction uses an owned matching LoD 1.13c game, single player Normal, ActI Blood Moor. Use a level 1 Barbarian with Buckler, Dex 20 for shield observations; the named factors below give scaled 62 and running 20. For Dodge/Avoid/Evade cases use an Amazon with the stated passive stats and no unwanted competing passive. Select a prepared character with the required stats; no save editing or action is implied by this record. Stand or use the named Walk/Run mode, receive the named melee/physical/elemental contacts, and read result codes/flags/seeds at the original locations listed in the historical procedure. Compare exact mode, stats, source profile and trigger before interpreting missing calls. For O01/O02/O03 collect 60 accepted hits per setup; for O04/O05/O06 and O11 use 60 contacts with the passive/missile controls. A returned block chance alone does not measure the caller's movement penalty. O07..O10 were actual historical observations only within their recorded cases. Pending observations require a separately authorized gameplay campaign.

## Migration and review record

2026-10-15: existing file boundary preserved. Consumption search of implementation and plans outside gaps found no file/prefix consumer; historical aliases were removed, and unchanged rule/vector numbers retained for references. The complete originals/encodings/private provenance are preserved with an integrity manifest outside the repository. All 32 existing numeric vector results/seeds remain unchanged; four supplied-input vectors add Holy Shield, type-matrix, no-local-cap and missile/avoidance-only checks. No gameplay vector was newly observed. New evidence changes R01a/R02a/R01b/R09d/R12d to verified static at their stated supplied-input/request boundaries; it does not promote unresolved producers. Static old labels retain dated support and limits. Current separate blind-reader results will be recorded after that pass; historical review counts above are not a current review claim.

Writer self-review (2026-10-15): checked the current contract against the preserved numeric cases, input widths, caller argument order, request-only versus stored-effect boundaries, RNG ownership, public binary identities and the complete gap assessment. All 32 prior result/seed sets were preserved, including hexadecimal-to-decimal result-word presentation and the identical seed aliases. The four added vectors are supplied-input examples; none claims live gameplay confirmation. Every verified rule has a referenced executable vector. The final writer check was `python .agents/skills/spec-writing/tools/check_spec.py combat-shield-block-and-avoidance`: 0 errors, 0 warnings. No gameplay suite or implementation test ran. The separate reader pass remains required for review completion.


### Fresh blind reader, full pass 1 (2026-10-15)

The separate reader used only the current contract and cited dependencies, with no implementation, source-analysis or web access. Counts reported: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 5, GATED 15, COVERAGE 18, HARDEST 3. This is a full pass, not a delta check. The following outcomes address every reported entry. No reader result implies user approval or changes an evidence label.

| Finding | Outcome |
|---|---|
| GAP: SBA-V32 blank traces | Replaced all three traces with explicit initialization, layer assignment/matching, maximum and eligibility steps; outputs 30/-5/0 and no-draw boundaries preserved. |
| DISCRIM: SBA-V16 missile mastery | Added c/d with a supplied nonzero 30 mastery result; missile skips the producer and retains AR 435, melee requests it once and obtains AR 536. |
| DISCRIM: SBA-V29 strict thresholds | Added exact AR 1048576 and pct 65536 boundary cases; direct amounts 346030/98304 distinguish premature divide-first amounts 346005/98250. |
| DISCRIM: SBA-V24 exclusion interval | Added immediately adjacent classes 703/710, which request application, contrasting retained excluded endpoints 704/709. |
| DISCRIM: SBA-V25 assertion branches | Added enabled weapon-block and Evade draws with null attacker; both abort 2602 after the step despite a successful comparison being possible. |
| DISCRIM: SBA-V10 shield short circuit | Case a now supplies an enabled weapon-block fixture; it still returns shield flag 16 after one draw, with no fold/token request. Existing output/seed are unchanged. |
| HARDEST: SBA-V16 producer order | Explicitly stated the already-recorded player call order: base, defense/flat adjustments, optional mastery and percent formation, percent scaling, negative swap. Added nonzero supplied mastery examples; actual producer remains gated. |
| HARDEST: SBA-V26 draw/map composition | No change to a/b: their missile-first draws and words 28/16 recompute correctly and preserve the prior values. Added c/d for SoftHit and GetHit precedence; SBA-V30 case f separates nonzero KnockBack eligibility from failed knockback success. |
| HARDEST: SBA-V32 fold and malformed trace | Fixed as the GAP outcome above; fold-only and nonpositive eligibility endpoints remain explicit. |
| COVERAGE: R01a large-base/large-modifier branches | Added SBA-V37 case a/b with exact divide-first final defense results. |
| COVERAGE: R01b ineligible branches | Added SBA-V38 case a-h, each returning contribution 0 with no evaluation request. |
| COVERAGE: R02a invalid class/helper input | Added SBA-V39 case a-c: omitted class term 315 versus fatal 4223. |
| COVERAGE: R02b mastery, false relation, non-monster | Added SBA-V16 case c/d and SBA-V34 case b/c. |
| COVERAGE: R02d nonpositive/negative defense | Added SBA-V27 case f-h: unchanged 500 for nonpositive stat and signed negative-defense result-250. |
| COVERAGE: R04 final negative clamp | Added SBA-V04 case e with minimum-i32 wrapping through both transfers, then final clamp; ratio 100 distinguishes a missing clamp. |
| COVERAGE: R06 zero denominator | Added isolated fault SBA-V40. |
| COVERAGE: R09a shield absence | Added SBA-V41 case a with invalid class, to prove early return 0 before class/stat access. |
| COVERAGE: R09b positive special result | Added SBA-V41 case b with NoShldBlock false and supplied special result 1. |
| COVERAGE: R10 raw/effective disagreement | Added SBA-V42 case a-c, distinguishing the effective moving predicate and raw Walk exclusion. |
| COVERAGE: R11 disabled positive shield | Added SBA-V42 case d with positive block 30 but blockEnable 0 and unchanged seed. |
| COVERAGE: R12a other moving/nonpositive | Added SBA-V43 case a-c, covering monster Run and zero/negative Evade. |
| COVERAGE: R12b w4 and precedence | Added SBA-V44 case a/b with exact logged type-query order and short circuit. |
| COVERAGE: R13 Run/maps 2/16 | Run preset already exercised by SBA-V19, so no new Run assertion. Added supplied map-code 2 in SBA-V23 case d; SBA-V22 case a already returns 32768 for actual weapon code 16 and now explicitly cites R13. |
| COVERAGE: R14 nonzero KnockBack discard | Added SBA-V30 case f: failed missile knockback still disables discard when the definition value is nonzero. |
| COVERAGE: R18a adjacent interval classes | Added SBA-V24 case d/e as above. |
| COVERAGE: R19 weapon/Evade null attacker | Added SBA-V25 case c/d as above. |
| COVERAGE: R20 SoftHit/failed knockback | Added SBA-V26 case c/d for SoftHit/GetHit precedence and SBA-V30 case f for failed knockback. |
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
| R01a exact base 1048576 / modifier 65536 | No additional recipe: R03/V29 already distinguishes both strict arithmetic thresholds, and V15/V37 exercise both R01a arithmetic branches. Those exact R01a inputs are not claimed as an observed or executed case. |
| R01b qualifying skill types 1/3 | No change: these fixed bindings select the same contribution recipe as V33; V38 distinguishes the qualification failures. |
| R02d monsterDefenseHalf / hireling | No change: V27 distinguishes the halving result; repeating each qualifier does not add a result-distinct recipe. |
| R12a moving mode 2 bindings | No change: V12 shows moving Walk's zero result, V11 positive Run, and V43 other Run/nonpositive cases; fixed bindings share the stated predicate branch. |
| R14 avoidance-only map codes | Added V36d: actual weapon-block code 16 maps to 32768, rejecting use of the missile mapper. Code 2 is V36b; codes 1/4 are unreachable from the stated ordinary producer because shield is skipped and nDamageParam 1 selects Avoid. Code 8 shares the explicit clear/no-Evade mapping recipe. |
| R16 three failed standing checks | Added V10d: shield, weapon block and Dodge all fail, consuming exactly three defender draws and returning code 0 / word 5. |
| R18a other post-hit qualifiers | No change: the same no-request result is exercised by V24's excluded classes; stat 117 defaults to zero in the other fixtures. |
| R21/R22 failed preconditions | No invented fixture: their actual integration remains asm-only under Q22. |
| R09e shield selection | No invented fixture: Q24/Q26 retain the unresolved selection boundary. |
| R02c plus R02d composition | No additional recipe: IgnoreDEF's zero remains zero through a positive fractional-defense adjustment, the same zero-result behavior already stated. |

| Reader hardest entry | Outcome |
|---|---|
| R01a/R03 ordered percentages and strict thresholds | No rule change: V29/V37 explicitly distinguish arithmetic order and threshold branches. |
| R04 minimum-i32 sequential transfers | No change: V04's wrapping case distinguishes both transfers and the final clamp. |
| R14 missile flags, knockback and maps | Retained V26/V30 and added V36d's result-distinct avoidance-only weapon mapping. |

Root recomputation for V10d: seed `{5,666}` produces remainders 99, 73, 5 and final
seed `{0x134E1255,0x54152EE9}`; the third step was independently calculated with integer
arithmetic from the preceding seed `{0xC997B681,0x09201110}`. V36d's seed `{4,666}`
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
| GUESS: R20 dangling else | Reworded the existing nested conditions explicitly: state 54 absence guards both reaction flags and knockback; GetHit selects OR 4, otherwise SoftHit selects OR 16384. Existing state-present and GetHit/SoftHit precedence cases remain unchanged. This is equivalent wording, not a new guard or policy. |
| DISCRIM: success cases alone allow <= | Added V11 case c (Evade chance 10/roll 10), V22 case e (weapon chance 10/roll 10 with Dodge zero), and V10 case e (Dodge chance 10/roll 10 with shield/weapon disabled). Each returns code 0 / word 5 and seed B1, rejecting a successful equality result. Shield equality already fails in V19 case a, so no duplicate shield case was added. V44 uses the same weapon comparison tested by V22; no duplicate equality fixture or repeated selected-stat binding was introduced. |
| COVERAGE: R01a null/zero/exact final thresholds | Added V37 case c for the null-unit memory fault and case d for zero total defense. No additional exact-threshold recipe: V29 already distinguishes the identical strict percentage thresholds, while V15/V37 exercise the final-defense direct/divide-first arithmetic. No exact-boundary R01a execution is claimed. |
| COVERAGE: R02b 128-entry limit | Added V34 case d with a supplied enumerator fixture offering 129 entries; the caller requests 128, the fixture returns the first 128, and typeBonus/pct are 128/158. Logged maximum and omitted layer 129 distinguish the consumer limit without claiming execution of the original enumerator. |
| COVERAGE: R02e non-player Stat 119 | Added V17 case b: raw AR 265 with percent 20 adds 53, yielding AR 318. Original zero-percent case a remains AR 265. |
| COVERAGE: R02g negative additions | Added V28 case f: supplied demon/undead bonuses -50/-40 add unchanged to AR 300, yielding 210. |
| COVERAGE: R09a shield plus missing list/invalid class | Added V41 case c: zero stat reads retain class factor 20 and yield scaled block -150. Case d supplies a shield and invalid class -1, producing the distinct memory fault; the existing shield-absent case still returns zero before that access. |
| COVERAGE: R09b negative block/null/other types | Added V41 case e for raw monster block -3, rejecting a lower clamp. No additional null/other-type binding fixture: their stated zero-return/no-draw result is the same output recipe as V21's ineligible zero block. Their explicit input predicates remain in R09b, and no null/other-type execution is claimed. |
| COVERAGE: R12a Walk/other mode 2 bindings | No change: V12 exercises moving Walk's zero result, V11 positive Run and V43 other-unit Run/nonpositive Evade. Mode 2 selects the same Evade recipe; repeating every fixed mode/type binding adds no distinct arithmetic, return or draw sequence. |
| COVERAGE: R12b 32-entry limit | Added V44 case c with a supplied enumerator fixture offering 33 entries. The caller requests 32, so the final value-99 entry is omitted, v remains 1, roll 10 fails and the type-query log has exactly 32 entries. This checks the consumer maximum and fold, not original enumeration integration. |
| COVERAGE: R12c negative/equality | Added V10 cases e/f: equality fails after one draw; a negative Dodge chance makes no draw and leaves seed B unchanged. Avoid uses the same stated selected-chance comparison/eligibility recipe; no duplicate fixed-stat fixture was added. |
| COVERAGE: R18a no-hook qualifiers | No change: V24's excluded classes already exercise the no-request recipe; eligible class 5 with Stat 117 zero is supplied by V18 case a, and default monster-attacker hits also fail the player-attacker qualification. The single conjunction produces the same no-request result for each failed qualifier; isolated execution of every qualifier is not claimed. |
| COVERAGE: R20 existing flags | Added V26 case e: incoming word 256 becomes 257 on hit preset and 261 after GetHit OR 4, with no draw. This rejects replacing prior flags with a fresh word. |
| COVERAGE: R21/R22 rejected preconditions | No invented integration fixture: actual hostility/range/owner and remaining root integration remains asm-only under Q22, as retained in the contract and gap file. |
| COVERAGE: R02d boss/neutral-half predicates individually | No change: V27's player/hireling case already distinguishes the positive-value halving recipe (DEF 375 instead of 250). Boss and the supplied neutral half condition select that same arithmetic; their predicates remain explicit, with the neutral producer still under Q18. No isolated boss/neutral-half execution is claimed. |
| HARDEST: R02c/R02d arithmetic and eligibility | Retained V18's ignore-defense eligibility and V27's halving, cap, nonpositive and negative-defense recipes. The existing player call order and supplied neutral-condition boundary remain explicit; no rule or gate changed. |
| HARDEST: R12b fold/candidates/classification/draw | Retained V32's corrected folds and V44's candidate logs; V22 equality and V44's maximum fixture add the missing distinct assertions. Classification/equipment producers remain gated, and every successful/failed draw still has its original owner. |
| HARDEST: R14/R20 flags and ordering | Retained V26/V30/V36, including the root's pass-2 avoidance-only weapon mapping. Explicit R20 wording and V26's prior-flag case supplement those recipes without changing the missile-before-defender order or maps. |
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
changed. The next reader run may therefore be **delta pass 4** for R20 and the
added/reformatted cases: V10 e/f, V11 c, V17 a/b, V22 e, V26 e, V28 f, V34 d,
V37 c/d, V41 c/d/e and V44 c. It must recompute the changed/added cases and confirm
the GUESS and DISCRIM resolutions. This record does not claim that delta pass 4
has run or that user review is complete.

### Fresh blind reader, delta pass 4 (2026-10-15)

The fresh smaller reader checked the actual main file, numeric foundations, game-data
fields and required main dependency contracts. It read no companion, gap file, previous
report, private source, implementation, runtime or web material. The scope was R20 and
V10e/f, V11c, V17a/b, V22e, V26a–e, V28f, V34d, V37c/d, V41c–e and V44c. The equivalent
wording and fixture-only changes qualified for a delta check after full pass 3.

The initial report gave **MISMATCH 6; GUESS 0; GAP 0; DISCRIM 4**, all arising from its
incorrect decimal conversion/remainder for seed B. It claimed roll 54 rather than 10 in
V10e/V11c/V22e/V26a/V26b/V44c; its four discrimination claims concerned the three equality
cases and V26a's successful shield result. The coordinator independently checked the exact
integer step and requested a calculation audit without changing the spec.

For B, `4 * 0x6AC690C5 + 0x29A = 7165593006`; low32 is `0xAB1A45AE`, decimal 2870625710,
high is 1, and unsigned remainder modulo 100 is 10. The reader independently recomputed this
with exact integers, also confirming F's remainder 36, then recomputed all delta cases.
It explicitly retracted each of the six mismatch and four discrimination entries. The
equality cases do fail at 10<10; V26a/b give words 28/16 and their stated seeds; V44c gives
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
reduction branches agree with the current rules. Edits were presentation only: SBA-V07
case traces now state why modes 2 and 1 keep 75, negative values are spaced, and
implementation-gap rows cite current rule ids. No vector input, output, rule, label or gate
changed, so no reader rerun is required.
