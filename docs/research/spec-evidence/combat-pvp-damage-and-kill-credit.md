# PvP damage and kill-credit evidence

This companion supports [the behavioral contract](../../specs/combat-pvp-damage-and-kill-credit.md). Original binaries are identified through the [source manifest](../source-manifest.md). Function descriptions are analyst labels unless explicitly identified as original exports. RVAs are module-relative; preferred VAs aid navigation and are not runtime-base assumptions. Private source collection details and complete originals are preserved separately.

## Original function references

| Artifact | Function RVA / original location | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | 0xdc0b0, VA 0x6fcfc0b0 | none | ApplyDamageModifiers | PVP-R01, PVP-R03, PVP-R04 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xde0c0 | none | ProcessCombatDamage | PVP-R01, PVP-R05a, PVP-R06 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xded30 | none | queued apply entry (`ProcessUnitCombatInteraction`; dequeues, calls 0xde0c0 with flag 0) | PVP-R01, PVP-R01a, PVP-Q10 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xddde0, VA 0x6fcfdde0 | none | build-time caller of 0xdc0b0 (`CalculateUnitDamage`) | PVP-R01, PVP-R01a, PVP-Q10 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdab90 | none | enqueue (`AllocGameResource`) | PVP-R01a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xa2c40, 0xa20c0 | none | basic attack do-function (`ExecuteWeaponCombatAction`); do table entry 1 | PVP-R01a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | VA 0x6fd2da48 | none | missile SrvHitFunc table (59 slots) | PVP-R01a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xa10b0, 0xa1210, 0x92cc0 | none | flag-1 apply via callbacks (`ProcessMissileHitEffects`, `CopyDataAndProcess`, unlabeled 0x92cc0) | PVP-R01a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0x4f450, 0x99290 | none | flag-0 self-apply, no stage | PVP-R01a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0x755c0 | none | direct caller of 0xdc0b0, then 0xde0c0 with flag 1 | PVP-R01 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xf42d0 | none | direct caller of 0xdc0b0, then 0xde0c0 with flag 0 | PVP-R01 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdd450 | none | record builder (`CalculateFullDamageStats`) | PVP-R01 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdbd00 | none | per-type resist/absorb (`ProcessMonsterBehavior_9160`) | PVP-R01 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdba40 | none | leech routine (`ProcessMonsterBehavior_8ea0`) | PVP-R08a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdac10, VA 0x6fcfac10 | none | scale selector (`ProcessMonsterBehavior_8190`) | PVP-R02a, PVP-R13 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | VA 0x6fd22ab0 | none | scale table (12 x 44 bytes, flag at entry `+0x20`) | PVP-R03 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xe0a60, VA 0x6fd00a60 | none | hostility predicate (analyst label `GetUnitOwner`) | PVP-R05a, PVP-R07 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xde5a0 | none | GetUnitCombatFlags | PVP-R05a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0x3ede0 | none | ProcessObjectInteraction | PVP-R05a, PVP-R06 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xa0370 | none | ValidateSkillTargetUnit | PVP-R05a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xf02a0 | none | CheckPlayerInteractionFlags (relation mask test) | PVP-R05a, PVP-R05b | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xe03a0 | none | FindUnitByTypeAndId | PVP-R07, PVP-R12 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0x6f930 | none | CreateMissileUnit | PVP-R07 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0x3ae10 | none | ProcessObjectTargetEffect | PVP-R07 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0x95d50 | none | SpawnHirelingAtTarget | PVP-R05c | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xeef10 | none | SpawnSummonedMonster | PVP-R05c | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xe0460 | none | SetupUnitSkillState | PVP-R05c | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdb1f0 | none | hit chance routine | PVP-R08b | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | VA 0x6fd11130 (rva 0xf1130); checks 0x6fd1115d, 0x6fd11174, 0x6fd11138-0x6fd1114d, 0x6fd1120d, 0x6fd1122c-0x6fd11253; write sites 0x6fd102f0, 0x6fd105b0 | none | declare-hostility handler | PVP-R09a, PVP-R09b, PVP-R09c | Historical decompiler lead 2026-10-11/12; actual instruction read 2026-10-15 below |
| `LOD113C-D2GAME` | 0xf10f0 | none | RemovePlayerFromParty | PVP-R09a, PVP-R09b | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdeee0 | none | ProcessUnitDeath | PVP-R10a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0x777f0 | none | TryApplyTargetInteraction | PVP-R10a, PVP-R13 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdea00 | none | DistributeKillExperience | PVP-R10a, PVP-R10b, PVP-R11a, PVP-R13 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdb5e0 | none | FindKillerPlayerUnit | PVP-R10a, PVP-R10b, PVP-R13 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0x1f2d0, 0xb65e0 | none | minion-spawn writers of victim bit 26 (`ApplyMonsterBitMaskEffects`, `ComputeSkillCalcValue_3500`) | PVP-R10a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdc030 | none | experience amount helper (`CalculateDamageWithLevel`) | PVP-R11a, PVP-R11b | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xddcb0 | none | hireling experience call (`ApplyPlayerExperienceGain`) | PVP-R11a, PVP-R11b, PVP-R13 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xde8f0 | none | DistributePartyExperience | PVP-R11a, PVP-R13 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xdfa60 | none | GetPlayerPartyId | PVP-R11a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xb2940 | none | ProcessArenaKillEvent | PVP-R11a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xa3090, 0x48fc0, 0x509e0 | none | bit-31 writers (`ExecuteSkillWithHirelingUpdate`, `ProcessMonsterResurrection`, `ApplySkillStateWithCalc_6690`) | PVP-R02b, PVP-R04, PVP-R13, PVP-Q01 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0xe0570 | none | party-membership test (`CheckUnitsInSameParty`) | PVP-R12, PVP-R13, PVP-Q07 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2GAME` | 0x9beb0 | none | FindPlayerPartyId | PVP-R12 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2COMMON` | 0x603C0 | 10237 | D2Common ord 10237 (difficulty record) | PVP-R02a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2COMMON` | 0x1a2d0 | 11104 | D2Common ord 11104 (hireling-class) | PVP-R02a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2COMMON` | 0x1a550 | 10064 | D2Common ord 10064 (MonStats `+0x0C` bit-6 test; formerly "monster-mask test") | PVP-R02a, PVP-R02b, PVP-R08b | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2COMMON` | 0x1a720 | 10278 | D2Common ord 10278 (MonStats `+0x0C` bit-7 test; formerly "mask-0x1c test") | PVP-R02a, PVP-R02b | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2COMMON` | 0x84d30 | 10666 | D2Common ord 10666 (tier-equal test) | PVP-R05a, PVP-R12 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2COMMON` | 0x391f0 | none | D2Common stat read of state 0x69 | PVP-R05a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2COMMON` | 0x2eb70 | 10920 | D2Common ord 10920 (relation list) | PVP-R05a | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2COMMON` | 0x2FE10 / 0x3C390 | 10331 / 10057 | D2Common ords 10331 / 10057 (room, town test) | PVP-R06 | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2COMMON` | 0x33DC0 / 0x384F0 / 0x380E0 / 0x38080 | 10927/10195, 11131/10473 | D2Common ords 10927/10195, 11131/10473 (stat-list redirect) | PVP-R10b, PVP-R10c | Historical 2026-10-11/12 disassembly/read record; limitations below |
| `LOD113C-D2COMMON` | 0x33020 | 10210 | D2Common ord 10210 (missile owner store) | PVP-R07 | Historical 2026-10-11/12 disassembly/read record; limitations below |

The historical table conflated several targets. Current exact imports below replace those locators; the historical reading is not silently treated as a new verification. D2Common 10331 is room getter 0x2FE10 and 10057 is town predicate 0x3C390. The redirect imports are not 0x2FE10.


## Import/export correspondence and targeted readings - 2026-10-15

Original D2Game.dll SHA-256 24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53 and D2Common.dll SHA-256 59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10 were rechecked against the available analysis specimens. This establishes current artifact identity; it does not identify DLLs loaded by an earlier gameplay session. PE import slots and matching target export tables establish:

| D2Game thunk | Import ordinal | D2Common target RVA | Role and support |
|---|---|---|---|
| 0xA3E2 | 10587 | 0x38C20 | Unit own-array stat read; initial declaration checks use 0x38B00, not effective-array selection |
| 0xA42A | 10973 | 0x38B70 | Unit effective stat read; repeated target level check consumes NRF-R12b/c |
| 0xA478 | 10331 | 0x2FE10 | Unit room getter; PVP-R06/PVP-R09a |
| 0xA808 | 10057 | 0x3C390 | Town predicate; PVP-R06/PVP-R09a |
| 0xA4C0 | 10920 | 0x2EB70 | Player data access; relation helpers; rejects absent/non-player unit by assertion |
| 0xA580 | 10064 | 0x1A550 | MonStats.boss test; PVP-R02a/PVP-R08b |
| 0xA7EA | 10278 | 0x1A720 | MonStats.primeevil test; PVP-R02a |
| 0xA7F6 | 11104 | 0x1A2D0 | Five-class hireling predicate; PVP-R02a |
| 0xA6F4 | 10666 | 0x84D30 | Same-tier predicate; PVP-R05a/PVP-R12 |
| 0xA7BA | 10237 | 0x603C0 | Difficulty record; PVP-R02a; compatible DPO import record |
| 0xA7A2 | 10927 | 0x33DC0 | Active-state intersection, calls 0x33C20; PVP-R10b |
| 0xA7B4 | 10195 | 0x384F0 | Flagged list search, calls 0x37F40; PVP-R10b |
| 0xA7FC | 11131 | 0x380E0 | Selected list owner-type value; PVP-R10b |
| 0xA84A | 10473 | 0x38080 | Selected list owner-id value; PVP-R10b |
| 0xA802 | 10910 | 0x39430 | Stat read used by defense adjustment; SBA owns complete slice |

D2Game import slot RVA 0xF8D10 identifies KERNEL32.GetTickCount. Its supplied u32 samples are an external time input, not a gameplay RNG draw. Current matches list reconstruction checked byte-for-byte against the image for Common 0x380E0/0x38080; those were read before assembly. All other newly read listed predicates/resolvers use disassembly where no such reconstruction exists. Original-image instruction decoding corroborates Game 0xDB5E0, 0xF10F0..0xF12BC, 0xF0090/0xF00E0, 0xF02A0/0xF02F0/0xF0640, 0xDAC10, and Common 0x33C20/0x33DC0/0x384F0/0x37F40/0x391F0/0x84D30, including previously encoded instruction runs. Game 0xDC0B0's factor arithmetic is traced from the selector through its table loop to the field stores; unchanged surrounding mitigation consumes DPO/DMI contracts and retains their dates.

<a id="redirect-helper-reading"></a>

### Complete flag-list helper reading - 2026-10-15

This narrow record applies to manifest artifact LOD113C-D2COMMON, SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`, 679936 bytes. Common ordinal 10195 resolves to RVA `0x384F0` (48-byte function), which calls `0x37F40` (64-byte function). Both complete disassemblies and their original-image instruction bytes were read. The wrapper dereferences its supplied unit before any null guard, so absent unit faults; absent base list or a base list without the extended/sign condition returns null. With an extended base it passes that base and the supplied flag mask to the helper.

The helper chooses the secondary list head when requested mask `0x2000` is set, otherwise the primary head; it clears that selector bit and visits the chosen chain in order. It returns the first list whose flags intersect the remaining requested mask, or null at the end. The base itself is not a candidate. Thus supplied `0x800` and `0x20` each select the first matching primary-chain list. There is no write or RNG step. This does not establish an alternate caller's eligibility, replacement policy, list propagation or other side effects. PvP's separate ordinal 10927 active-state intersection is not a generic base-list-existence predicate.

## Dispatch and output anchors

1. Compatible shared damage chain: server frame loop Game 0x79A10 -> frame-event table 0x1079D0 -> skill handler 0x79170 -> do entry 0xA2060 -> do dispatcher 0xA19A0 -> do table 0x1074A8 slot 1 -> basic attack 0xA2C40 -> build completion 0xDDDE0 -> scale/mitigation 0xDC0B0 -> enqueue 0xDAB90 -> queued apply 0xDED30 -> apply 0xDE0C0. This reuses SCD/DPO's dated root record; it was not rerun as a fresh full frame-loop analysis.
2. Game 0xDC0B0 -> selector 0xDAC10 -> field-selection table RVA 0x102AB0, 12 entries of 44 bytes, scale flag at entry +0x20 -> selected DamageRecord write. Existing 2026-10-12 caller/xref analysis found four direct stage sites: 0xDDDE0, 0xDE0C0, 0x755C0 and 0xF42D0. Queued/basic and direct-route counts remain those recorded; no universal exhaustive call graph beyond that recorded search is claimed.
3. Game 0xDE0C0 -> hostility 0xE0A60 -> directional relation 0xF02A0 or Common #10666/0x84D30 -> Common tier reader 0x391F0. The early result-bit writes at Game 0xDE0F1 and 0xDE130 are the output anchors. State-list lookup Common 0x38520 ->0x37F80 visits primary then secondary lists; own stat reader 0x38BD0 consumes NRF-R12b/c.
4. Missile contact Game 0x3AE10 -> apply 0xDE0C0 with enable 1; the resolved creator supplies attacker. Missile creation/owner store Game 0x6F930 -> Common #10210/0x33020 are historical supporting reads, not proof that every contact passes missing-owner preconditions. SBA-R22 retains contact integration limits.
5. Player-relation command handler Game 0xCC8E0 (7-byte command) -> relation-operation dispatcher 0xF0A70 -> table RVA 0xFA3B0, slot 4 at 0xFA3C0 containing 0x6FD11130 -> declaration entry 0xF1130. Alternate 6-byte command handler 0xCC860 reaches the same dispatcher. The dispatcher looks up the target id in the game's player hash and passes the acting player separately. Instruction and table checks establish acting player in the handler's second register argument and target in its first stack argument; these storage details do not constrain a reconstruction's API.
6. Declaration 0xF1130 has room/town/level guards, relation-node requests 0xF02F0 twice, relation read 0xF02A0, timestamp read 0xF00E0 and store 0xF0090, then accepted activation 0xF0FF0 and cleanup 0x386F0. Its rejection write at 0xF12B2 is the initial rejection output anchor. Relation flag mutation is 0xF0640, notifications 0xF05B0. The full activation/party/cleanup/message callees are not completely verified here; R09b/R09c/Q03 retain that integration.
7. Death root (call sites verified 2026-10-15 in the original D2Game image): anchor 1's frame loop reaches basic attack Game 0xA2C40 (reused [DPO dispatch record](combat-damage-pipeline-order.md#dispatch-and-output-anchors)). 0xA2C40 calls build completion 0xDDDE0 at 0xA2D8F and, on the branch where its thunk 0xAE7A test returns 0, queued apply 0xDED30 at 0xA2DC7. 0xDED30 calls apply 0xDE0C0 at 0xDEDF8 and result consumer 0xDCF20 at 0xDEE9C with the copied result record. 0xDCF20 calls death 0xDEEE0 at 0xDD053 when record flag word +4 has bit value 2 set. 0xDEEE0 calls wrapper 0x777F0 at 0xDEF81, the function's only caller. 0x777F0 calls experience router 0xDEA00 at 0x77811 when both units are present and victim word +0xC4 bit 26 is clear. 0xDEA00 calls credited-player resolver 0xDB5E0 at 0xDEA48, its only caller, after null, unit-type 0/1 and positive-value guards. 0xDB5E0 calls monster owner lookup 0xE03A0 at 0xDB635, 0xDB68B and 0xDB6EE, and reaches the four Common targets above. Other direct callers of 0xDEEE0 (16 further sites) are outside this kill-credit root. Router 0xDEA00 requests hireling update 0xDDCB0, party distribution 0xDE8F0 or solo amount 0xDC030 followed by level update 0xDE6D0. Scoring call 0xB2940 is separate after the wrapper; its missing-context guard precedes score writes. Full scoring and complete pet-list search are not covered by supplied-input vectors.

## Semantic field correspondence

| Semantic term | Original encoded field/condition | Rules |
|---|---|---|
| unit type/class/id | Unit +0/+4/+0xC | PVP-R02a/PVP-R07/PVP-R10a |
| monster link | Unit +0x14 -> MonsterData +0x28 -> owner block; block game/type/id at+0x28/+0x30/+0x2C | PVP-R07/PVP-R10a/PVP-R12 |
| missile link valid/type/id | Unit +0xC8 mask 0x400, +0x94/+0x98 | PVP-R07 |
| scaleCondition / skipKillExperience / alignmentCreationCondition | Unit +0xC4 masks 0x80000000/0x04000000/0x00020000 | PVP-R02a/PVP-R10a/PVP-R05c |
| relation nodes | PlayerData+0x38; node otherId+0,flags+4,timestamp+8,next+0x10; eligibility mask 0x08 | PVP-R05a/PVP-R09a |
| game presence / party sentinel | Unit+0x80; party id 0xFFFF | PVP-R12 |
| damage scaled fields | physical+8,fire+0x10,lightning+0x1C,magic+0x20,cold+0x24,poison+0x28,drains+0x38/+0x3C/+0x40 | PVP-R03 |
| unscaled fields | poison/cold/freeze lengths+0x2C/+0x30/+0x34; burn+0x14/+0x18,stun+0x44 | PVP-R03 |
| result/override | result+4,hit 1/lethal 2; recordflags+0 mask 0x1000 | PVP-R06 |
| boss/primeevil/inTown | MonStats+0x0C bit 6/bit 7/bit 10; town reads byte+0x0D &mask[2] | PVP-R02a/PVP-R06 |
| difficulty scale | DifficultyLevels+0x38; game difficulty+0x6D | PVP-R02a |
| defense exclusions | MonsterData+0x16 masks 0x0A/0x02; SBA neutral exclusion/half terms | PVP-R08b |
| alignment | state 0x69 list, own stat 0xAC/layer 0 | PVP-R05a |
| credit-state predicate | unit base+0x5C; extended sign bit+0x10; active mask+0x58; supplied mask DataTables+0x144; count+0xC4 | PVP-R10b |
| credit list | primary head extended+0x3C, secondary+0x40 excluded; next+0x2C; flags+0x10 mask 0x800; owner type/id+8/+0xC | PVP-R10b |
| pet count / scoring context | DataTables+0xBF0 / Game+0x1D28 | PVP-R11a |

These encodings explain the evidence and are not implementation requirements. Original binary fixtures and exact earlier text are retained privately with SHA-256.

## Corrections and evidence limits

### Credit redirect

The 2026-10-11/12 main text described #10927 as a generic stat-list existence test, conflated redirect accessor targets with the room getter, and said a victim's list wins without stating null erasure. The current original-image reading establishes active-state intersection separately from flag-list search. Game 0xDB63E initializes selected list 0;0xDB654 assigns the killer list;0xDB66F assigns the qualifying victim result unconditionally;0xDB671 tests that final selection. Thus a qualifying victim with no list clears the killer selection and retains the one-hop owner. A present final list whose owner lookup fails instead removes credit. PVP-V18's two existing outcomes remain P1/P2; new PVP-V24 distinguishes both missing cases and primary-list order. PXL-R02a independently described the same null-erasure branch; the correction here is supported by original bytes rather than that sibling's wording alone. PVP-R10c retains wider producer uncertainty, not unread owner accessors.

### Scale factor and named predicates

Original text left MonStats bits 6/7 names unresolved and assumed every selected P<=65536. GDF's dated descriptor/flag records identify boss and primeevil; they do not make scaleCondition a table flag. Game 0xDAC10 can return the signed difficulty field unchanged, and 0xDC0B0 has the second operand-threshold branch. R04 therefore uses NRF-R05(P,v,100), retaining every existing numeric result and adding synthetic P = 70001/-50/0/missing-record cases. No original table value is invented. The one-run queued/basic correction was originally recorded 2026-10-12 after a missed encoded call in 0xDDDE0 was located; that correction and historical per-route search limits remain valid.

### Alignment and requests

Common 0x391F0 reads state 105 for players as well as monsters. Missing base/list cases return 0; other unit types require a present base before returning 2. The same-tier helper returns 1 for identical non-null units before comparing tiers. The original generalization to all non-player/non-monster units omitted that base-list guard; V26 adds that distinction. Game 0xE0460 requests stat-list creation/attachment and propagation with unread/incompletely integrated effects. R05c now asserts the creation caller's selected tier request and direct Align 1 condition write; Q11 explicitly blocks automatic stored-tier/propagation acceptance. Old V14 expected tier numbers 2/1/0/0 remain as request values, not an invented observation of an attached list.

### Declaration verification

The original record said the handler was absent from the earlier source view and used only a decompiler reading. The handler entry 0xF1130 is inside the listing beginning at 0xF10F0; original bytes corroborate the complete handler instruction range. Command lookup/table dispatch establishes acting/target roles. R09a retains verified local numeric/room/time decisions and explicit request boundaries. The 60000-ms timestamp addition is in the declaration handler itself, not the preceding remove-from-party entry 0xF10F0. Initial level/room rejection writes 1; time refusal instead sends requests and leaves that rejection output unchanged. Full enable/disable, party removal, cleanup and message effects remain Q03; R09b moves from recovered to unresolved integration on located source, and R09c remains unresolved for complete command behavior.


### Relation raw-result fixture correction

The original PVP-R05a/PVP-V13 already identified eligibility mask `0x08` and reported a boolean hostile result, so no original numeric vector output changes here. Common relation consumption in Game `0xF02A0` returns `node.flags AND 8`: raw 0 or 8, not normalized 1. During this migration's new PVP-V25 draft, its equality case initially supplied an impossible raw result 1/request 1; before stable handoff it was corrected to raw 8/request 8 with the node's eligibility flag set. This affects only that new supplied-input case. Existing old boolean cases remain true/false as written; the declaration comparison must use the raw result. The complete original snapshot and pre-handoff edits are retained privately.

### Shared effects, RNG and builds

The original PVP-R08c/Q07 said block/avoid, Crushing Blow, Open Wounds, monster drops and ears were not found. No six-rung search record supported the absence label. Current SBA/CBL/OWD contracts own the named combat effects; Q07 retains complete combined integration and leaves loot/ears outside this contract without an absence claim. No new broad search or gameplay attempt was performed merely because the old record is dated.

The historical 2026-10-12 RNG audit reported zero direct draws at Game 0xE0570/0x509E0/0xA3090/0xDAC10 and credit/experience functions. It is not proof of arbitrary callback or dependency interiors. R13/V22 now assert the owned local decisions and explicitly supplied no-draw operations. Original seed-invariance intent remains at that bounded execution; a universal complete-attack assertion is withdrawn with this rationale. No 1.14d equivalence was tested historically or now, and it is outside the evidence-build policy; the former confirmed_1.14d vector column was uniformly untested and does not establish a 1.13c observation.

## Compatible data and dependency records

Named stat/state identities consume GDF-R11a/GDF-R12 and their table records for effective itemstatcost.txt/states.txt. Class names/hcIdx consume GDF-R17a; field/flag correspondence consumes GDF-R02/R05a/R07/R18, and PetType count ownership consumes GDF-R19b. The underlying records retain their actual 2026-10-12/13/14 dates, effective-loader profile and archive/hash/override limits; this migration does not claim a new loader read or pristine-archive/compiled correspondence. The count identity corrects the old unknown global-field name without explaining why the consumer uses >7.

SBA-R02c/R02d own defense adjustment, DPO/DMI stage/guard placement, IEP event traversal, LSE leech effectiveness and PXL experience amount/credit boundaries. These supporting records retain their labels/gates. The linked specs are not automatically full integration evidence. Pending PVP-O01 has no measured matching case. Old runtime tooling limited to a later patch was an investigation limitation, not proof that original 1.13c observation is impossible.

<a id="gap-assessment-and-migration"></a>

## Gap assessment and migration - 2026-10-15

| Question | Assessment and outcome |
|---|---|
| PVP-Q01 | Historical three-writer pattern search remains partial; no repeat with unchanged conditions. Actual hit-time carriers remain outside-boundary; supplied condition cases retain 17/100 |
| PVP-Q02 | Reused GDF name/flag/count records; targeted redirect imports/helpers correct predicate and selected-list behavior. Wider flags/mask/list producer meanings remain open |
| PVP-Q03 | Newly located embedded handler, command/table caller and timestamp helpers establish roles/local checks; complete side-effect callees remain incomplete, so no integrated declaration closure |
| PVP-Q04 | Positive supplied victim experience remains valid static routing; player-victim gameplay observation pending |
| PVP-Q05 | Targeted tier reader/helper shows player/list behavior and base guard; actual player writers remain producer question |
| Resolved difficulty/town identity | Resolved difficulty-field and inTown identity folded into current rules; dated resolution retained here |
| PVP-Q07 | Shared owners replace stale absence claim; full combined integration remains; loot/ears outside-boundary and not established |
| Later-build equivalence | Historical 1.14d untested limit retained above; outside current 1.13c build, no main-contract question |
| Resolved owner-link identity | Historical owner-link double-dereference and mask 0x08 correction preserved in semantic correspondence; current NRF agrees. No current conflict remains |
| PVP-Q10 | Historical 2026-10-12 encoded build-time call resolves placement; plain-melee observation remains pending |
| PVP-Q11 | New precise integration gate for existing alignment writer's unverified stored allocation/attachment/propagation |
| PVP-Q12 | Pending gameplay coverage for all statically verified rules; no gameplay actions authorized in this migration |
| PVP-Q13 | Precise existing limits: complete hireling search and present-context scoring/award integration; supplied request routing and absent-context guard available |

No non-gap plan, implementation or test consumer of this file/PVP prefix was found. Existing rule/vector ids remain to keep sibling references stable; no aliases remain in the main contract, and retired question ids are not reused. The complete main/gap originals and SHA-256 were captured before edits. All 22 original vector family ids and their supported exact numeric outcomes remain; V14 explicitly becomes request-only, V22 loses its unsupported complete-attack scope, V17 adds actual defense outputs, and V23 through V27 add distinguishing supplied fixtures. All original observation statuses remain untested. Historical six-row implementation comparison is preserved in the gap file without rerunning or claiming it current. No implementation code, tests, plan, game action, hook, Git action or shared document was changed by this writer.

## Writer review

Writer self-review on 2026-10-15 checked all 25 leaf rules, the retained parent headings, 27 executable vector families, ten current gap rows, both dependency directions, semantic input/output coverage and explicit supplied-operation boundaries. The original 22 vector families retain their supported numeric outcomes; exact scope corrections and new-case correction are recorded above. Main text uses no original internal locations. Public locators and hashes were rebuilt from the pre-spacing text/source records after an overbroad draft spacing pass was discarded; exact identifier checks are preserved privately. Final mechanical check: the spec checker reported 0 errors and 0 warnings for this three-document revision. The private preservation record retains all 31 existing rule/heading identities, all 22 original vector identities and the supported outcomes; five supplied-input families were added. No implementation test or gameplay observation ran. Independent blind-reader review is coordinated separately and is not yet claimed complete.

## Blind-implementer full pass 1 - 2026-10-15

A fresh smaller-model reader used the current main contract and cited main dependency contracts, without original source, implementation, runtime or web access. It independently recomputed all 27 vector families and every listed subcase; it did not supply a numeric subcase count. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 7, GATED 6 grouped entries, COVERAGE 1 summary, HARDEST 3. This run identified fixture-discrimination gaps and does not complete review of the resulting revision.

| Finding | Disposition |
|---|---|
| DISCRIM 1: positive selected fields omitted/swapped | Added PVP-V11(b), with asymmetric positive lightning/cold/magic/poison/manaDrain/staminaDrain and a distinct asserted result for all nine scaled fields. Original exclusion case remains |
| DISCRIM 2: both scale thresholds exceeded | Added PVP-V23(b): factor 70001, magnitude 1048577, v-first result 733960485; wrong P-first result 734003900 explicitly rejected |
| DISCRIM 3: town exception ignores invalid/missing class | Added PVP-V15(e), independent negative-class and unavailable-record runs, each with inTown supplied true but actual predicate false; both retain hit and clear lethal |
| DISCRIM 4: defender owner lookup failure | Added PVP-V16(b), failed first lookup retains M2 rather than null and stops further lookups |
| DISCRIM 5: missile link-valid condition skipped | Added PVP-V16(c), false link-valid condition retains missile X and makes no lookup even though stored owner fields could resolve |
| DISCRIM 6: alignment secondary fallback | No duplicated lookup fixture: R05a now explicitly consumes NRF-R12a; PVP-V26 cites NRF-V33(g), which returns the matching secondary list after a nonmatching primary, and NRF-V33(a), which prefers an existing primary match. Both use the same state-list lookup regardless of nonzero state identity; PvP supplies 105. Existing PvP fixtures separately test that caller's fixed state/stat identities and returned tier |
| DISCRIM 7: signed timestamp comparison | Added PVP-V25(f): unsigned stored 2147483649/current 2147483647 refuses; signed comparison would wrongly allow. No second sample, store or activation is permitted |
| GATED 1: actual scale/skip/mask/list producer roles | No change: PVP-Q01/PVP-Q02 already preserve these outside-boundary inputs without gating supplied consumer values |
| GATED 2: command validation/stored relations/messages/cleanup | No change: PVP-Q03 preserves complete declaration integration; local checks and requests remain independent |
| GATED 3: player tier producers/alignment propagation | No change: PVP-Q05/PVP-Q11 retain actual producers and stored effects; lookup/request assertions remain available |
| GATED 4: combined effects/credit | No change: PVP-Q07 preserves complete combined integration under the named effect owners |
| GATED 5: gameplay observations | No change: PVP-Q04/PVP-Q10/PVP-Q12 remain pending; no gameplay operation ran |
| GATED 6: full hireling search/scoring/awards | No change: PVP-Q13 retains these dependency-backed effects; supplied routing is not their verification |
| COVERAGE 1: summary of all families | No change: the entry summarized existing coverage and reported no additional missing branch |
| HARDEST 1: victim null erasure versus missing owner | No change: R10b and V24 distinguish erased list selection (retain original owner) from present list with failed owner lookup (no credit) |
| HARDEST 2: unsigned timestamp/equality/second tick | Added signed-boundary discriminator V25(f); existing V25(b) retains equality and second-sample assertions. No rule or gate changed |
| HARDEST 3: list presence versus active membership and absent identity | No change: V26/R05a/R12 explicitly distinguish these boundaries; the consumed secondary lookup case is identified above |

These changes add cases/assertions and an explicit reference to the existing shared lookup contract; rule behavior, evidence labels and gates are unchanged. All original vector cases and expected values remain. The contract still has 27 vector families. Fresh full pass 2 is required for this resulting revision. No new original-source campaign, implementation test or gameplay observation was performed for these fixture additions.

## Blind-implementer full pass 2 - 2026-10-15

The fresh reader independently recomputed 27 vector families and 92 explicit runs. Reported counts: MISMATCH 1, GUESS 0, GAP 0; its DISCRIM header said 7 but it enumerated eight distinct findings. The ledger therefore records and disposes all eight below; the original header discrepancy is preserved privately rather than silently treated as seven. GATED had one grouped entry covering the current questions; COVERAGE had one summary; HARDEST had three. The mismatch was a wrong dependency citation, not a numeric result disagreement.

| Finding | Disposition |
|---|---|
| MISMATCH 1: non-monster leech effectiveness owner | Corrected R08a and section 9 from LSE-R10 (capped-add helper) to LSE-R03 (effectiveness). The latter explicitly supplies 100 for non-monster defenders. The value 100 and all prior leech-input outcomes are unchanged; no source rereading was needed |
| DISCRIM 1: duplicate directional nodes | Added PVP-V13(g): first B-node flags 0, later B-node flags 8 returns false, distinguishing the already stated first-match rule from any-match scanning |
| DISCRIM 2: hireling/boss priority | Added PVP-V10(e): both hirelings plus defender boss and difficulty factor 37 still chooses 25/640, before a difficulty-field read |
| DISCRIM 3: defender chain capped at two | Added PVP-V16(d): three successful hops resolve M3 through M2/M1 to player A and log all three lookups |
| DISCRIM 4: unrelated result flags lost | Added PVP-V15(f): incoming result 7 becomes 5 at town rejection, retaining hit and unrelated bit 4 |
| DISCRIM 5: recursive same-party owner resolution | Added PVP-V21(f): one-hop M1 versus player A, tiers 1/1, returns false; an extra hop would reach A and wrongly pass identity |
| DISCRIM 6: credit mask partial final word | Added PVP-V24(d): stateCount 33 scans two words, with only state 32 intersecting, returning true |
| DISCRIM 7: timestamp sum overflow | Added PVP-V25(g): supplied second tick 4294967280 plus 60000 wraps to 59984 and is the exact store request |
| DISCRIM 8: hireling product wrapping | Added PVP-V20(b): 50000000*86 wraps to 5032704 before division, share 19659; wide unwrapped result 16796875 is rejected |
| GATED 1: PVP-Q01/PVP-Q02/PVP-Q03/PVP-Q04/PVP-Q05/PVP-Q07/PVP-Q10/PVP-Q11/PVP-Q12/PVP-Q13 | No change: the grouped entry repeats the current producer, integration and pending-observation boundaries. Existing rules/Id map/questions already state their exact available supplied-input subsets |
| COVERAGE 1: existing coverage summary | No change: this is a summary of recomputed coverage, not another missing branch |
| HARDEST 1: PVP-V23(b), competing thresholds | No change: exact operand order and both unequal computed results are already explicit; new hireling-priority case tests a different selector branch |
| HARDEST 2: PVP-V25(f), unsigned time | No change to that case: unsigned comparison and refusal log are explicit. V25(g) separately adds the requested wrapping discriminator |
| HARDEST 3: PVP-V24(a-c), null erasure/missing owner | No change: selected-list absence retains the initial owner while a present list with failed owner lookup removes credit, with distinct lookup logs |

All eight additions are focused cases within existing families. Original cases and supported values remain; the Id map, open questions and gap file are unchanged. The only rule/dependency text correction is LSE-R03's owner reference; no rule arithmetic, ownership boundary, evidence label or gate changes. The contract remains 27 vector families. A fresh full pass 3 is required after these new fixtures and the source-reference correction. No implementation test, original-source campaign or gameplay observation ran in this triage.

## Declaration level-source clarification - 2026-10-15

Full pass 3 exposed that R09a's earlier phrase "target level checked again" did not distinguish the actual source arrays. Reusing the already complete Game handler instruction record: initial actor check at `0xF1158` and target check at `0xF116F` call thunk `0xA3E2`, import #10587, Common `0x38C20`; the later target check at `0xF11B0` calls thunk `0xA42A`, #10973, Common `0x38B70`. The import identities were already resolved against the exact images. A narrow complete read of Common `0x38C20` (96-byte function; no checked source reconstruction listed) confirms unit/list/id/definition guards followed by `0x38B00`, which always selects the own array. Common `0x38B70` instead selects through `0x38A80`, whose own/resolved behavior is the existing NRF-R12b contract. Compatible full helper/minimum records remain dated 2026-10-14; this does not renew array-construction or aggregation proof. The currently available Common image was rehashed to the manifest's unchanged 679936-byte SHA-256 `59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`.

R09a and its input/fixture preamble now name initial own-array and later effective reads separately. Existing declaration fixtures supply their same level in both arrays, retaining every earlier result. PVP-V25(i) supplies immutable target own 9/effective 8 with usesResolvedStats=true and minimum adjustment disabled: initial checks pass, the later check returns early. No setter, changing getter return, concurrent mutation or guessed node-creation effect is invented. This is a semantic input/source clarification and therefore requires another fresh full reader even after full pass 3's zero numeric mismatches.

## Blind-implementer full pass 3 - 2026-10-15

The fresh reader recomputed 27 vector families and 101 explicit cases. Counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 7, GATED 9 grouped entries, COVERAGE 9 summaries, HARDEST 3. Every finding is disposed below. A zero mismatch count does not complete review after the source/input clarification and fixture corrections.

| Finding | Disposition |
|---|---|
| GAP 1: in-range absent MonStats record contradicts GDF-R05a | Replaced only PVP-V15(e)'s class 0/count 734/unavailable-record run with class 734/count 734, a supported out-of-range boundary. Negative-class run remains. Both still return result 1 with unchanged physical 2560. The invalid fixture was not evidence that an ordinary in-range record can be absent |
| DISCRIM 1: hireling against player | Added PVP-V10(f): class 271/scaleCondition false still selects 17 and physical 435 |
| DISCRIM 2: override plus missing defender room | Added PVP-V15(g): actual hostility false, override true, no room, result 7 ->4; override bypasses hostility alone |
| DISCRIM 3: same present player | Added PVP-V13(h) with an eligible synthetic self-node and expected hostility false. The initial draft also imposed a pure-read call constraint; full-pass-4 root review removes that constraint as recorded below, while retaining the semantic result |
| DISCRIM 4: raw 8/request 0 | Added PVP-V25(h), which selects the disable branch without tick/store requests; complete effects remain Q03 |
| DISCRIM 5: owner-link game versus supplied redirect game | Added PVP-V24(e), separate registries X/G mapping list owner 42 to different players; lookup logs and credited P2 require redirect through G |
| DISCRIM 6: defender self-owner | Added PVP-V16(e), one successful self lookup then stop and retain M |
| DISCRIM 7: repeated target level below 9 | Verified the existing handler's distinct imports and narrowly read the initial own-array wrapper. R09a/input/preamble now distinguish own/effective views; PVP-V25(i) supplies own 9/effective 8 immutably. Fresh full 4 is required, rather than a delta |
| GATED 1: PVP-Q01/PVP-Q02 actual condition/skip/mask/list producers | No change: these outside-boundary producers do not gate supplied local consumer arithmetic |
| GATED 2: PVP-Q03 full declaration integration/meaning | No change: stored relations, cleanup, message delivery and complete command semantics remain gated |
| GATED 3: PVP-Q04 player-kill observation | No change: actual award observation remains pending |
| GATED 4: PVP-Q05 player alignment writer | No change: actual producer is distinct from the supplied list reader |
| GATED 5: PVP-Q07 combined effects/credit | No change: full combined integration remains; ears/loot stay outside this contract |
| GATED 6: PVP-Q10 plain-melee observation | No change: static run count is distinct from pending measured gameplay |
| GATED 7: PVP-Q11 alignment propagation | No change: caller-selected tier values do not establish attachment/propagation |
| GATED 8: PVP-Q12 general gameplay | No change: no new gameplay observation ran |
| GATED 9: PVP-Q13 hireling/scoring/awards | No change: supplied lookup/request routing does not verify full dependency effects |
| COVERAGE 1: factor category | No change: summary of existing covered factors, not a new missing branch |
| COVERAGE 2: arithmetic category | No change: summary of existing wrap/threshold cases, not a new missing branch |
| COVERAGE 3: fields category | No change: summary of covered selected/excluded fields |
| COVERAGE 4: hostility category | No change: summary of existing relation/tier cases |
| COVERAGE 5: guards category | No change: summary of covered rejection/override cases |
| COVERAGE 6: owners category | No change: summary of covered depth/fallback cases |
| COVERAGE 7: declarations category | No change: summary of covered declaration decisions |
| COVERAGE 8: credit category | No change: summary of covered player/redirect/null cases |
| COVERAGE 9: routing category | No change: summary of covered supplied award/scoring decisions |
| HARDEST 1: R04 thresholds/wrap/operand order/truncation | No change: V02/V12/V23 already supply exact distinct values and both threshold precedence |
| HARDEST 2: R10b null erasure/lookup failure | No change to existing cases: V24(a-c) state distinct retained-owner/no-credit results; new game-selector case tests another assertion |
| HARDEST 3: R09a unsigned/equality/ticks/wrapping | No change to those cases: V25(b/f/g) retain exact comparisons and request values. The own/effective source clarification is independently recorded above |

This revision retains 27 vector families, all original supported numbers and unchanged labels/gates. It corrects one impossible supplied fixture without changing its result, adds seven focused discriminators, and clarifies the declaration input sources with original evidence. The Id map/questions/gap file remain unchanged. No implementation test, gameplay operation or broad source campaign ran; only the narrowly required own-array wrapper/source applicability check was added. Fresh full pass 4 is required for the resulting contract.

## Blind-implementer full pass 4 and root review - 2026-10-15

The fresh reader recomputed 27 vector families and 112 cases. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 5, GATED 8 grouped entries, COVERAGE 0, HARDEST 3. The root independently identified one excess assertion in V13(h). All entries have dispositions below. This run accepts the clarified source inputs for its reviewed revision; the subsequent fixture additions and assertion removal still require fresh review.

| Finding | Disposition |
|---|---|
| DISCRIM 1: small-branch signed product wrap | Added PVP-V12(b): P 65535/v 65535 wraps to signed -131071, then returns -1310 rather than wide 42948362 |
| DISCRIM 2: extended base is not a redirect candidate | Added PVP-V24(f): creditEligible true, extended base marked redirect/owner P6, no primary match; credit remains one-hop P1 rather than P6 |
| DISCRIM 3: initial actor own/effective distinction | Added PVP-V25(j): actor own 8/resolved 9 and target 9/9 rejects before node requests; no changing getter output is supplied |
| DISCRIM 4: disable bypasses repeat target-level guard | Added PVP-V25(k): initial own 9/9, target effective 8, raw 8/request 0 enters disable; the enable-only repeat-level guard must not prevent this route |
| DISCRIM 5: absent killer/victim | Added PVP-V27(b), two fresh wrapper runs, each returning before award routing with no requests |
| Root finding: pure-read/helper constraints in PVP-V13(h) | Removed the empty relation/tier-read log acceptance assertion and rejection of an equivalent self-rejecting helper; removed that constraint from its Inputs/Trace/Outputs/explanation and updated the earlier companion disposition. The eligible synthetic self-node and final hostility=false remain. Original instruction order is evidence; equivalent pure-read/API structure is not an added behavioral requirement |
| GATED 1: PVP-Q01/PVP-Q02 producer roles | No change: actual scale/skip/mask/list producers remain outside-boundary |
| GATED 2: PVP-Q03 declaration integration | No change: complete stored relation/cleanup/delivery behavior remains gated |
| GATED 3: PVP-Q04 player-kill observation | No change: pending gameplay observation remains separate from supplied routing |
| GATED 4: PVP-Q05 player alignment | No change: actual writer remains unestablished |
| GATED 5: PVP-Q07 combined effects/credit | No change: complete combined integration remains under its owners |
| GATED 6: PVP-Q11 alignment propagation | No change: stored creation/list effects remain gated |
| GATED 7: PVP-Q12 general observation | No change: no gameplay confirmation was added; PVP-Q10's existing plain-melee observation also remains pending |
| GATED 8: PVP-Q13 hireling/scoring/awards | No change: supplied request decisions remain distinct from integrated effects |
| HARDEST 1: PVP-V25(i), own/effective placement | No change to that case: the four immutable arrays and final early return are explicit. V25(j)/(k) add distinct initial-actor and disable controls |
| HARDEST 2: PVP-V24 priority/null/game contexts | No change to existing cases: null erasure and separate games are explicit. V24(f) adds only the missing base-versus-primary discriminator |
| HARDEST 3: PVP-V23(b), threshold precedence | No change: both exact, unequal results and operand order remain explicit |

There were no coverage findings in this run. All additions/removal are fixtures or acceptance assertions; rule behavior, source inputs, ownership boundaries, labels, Id map, open questions and gap file are unchanged. The contract still has 27 vector families. With full pass 4 at MISMATCH 0 and every entry disposed, fresh DELTA pass 5 may review exactly PVP-V12(b), PVP-V24(f), PVP-V25(j)/(k), PVP-V27(b), the revised PVP-V13(h), and their unchanged rule/input/dependency context. Assertion removal is included in that rerun. This is not a completed delta or user approval. No source campaign, implementation test, gameplay operation or Git action ran in this triage.

## Blind-implementer delta pass 5 - 2026-10-15

A fresh smaller reader checked seven changed cases: PVP-V12(b), PVP-V24(f), PVP-V25(j),
PVP-V25(k), both PVP-V27(b) runs and revised PVP-V13(h). Counts on changed material:
MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0. No changed-material finding required another edit.
The semantic self-node result is checked without a pure-read or helper-call constraint.

Rule behavior, input sources, ownership and gate boundaries are unchanged since full pass 4.
The applicable reader stopping condition is met. The document remains draft with its
existing evidence and observation gaps; gameplay confirmation and maintainer approval
remain outstanding.
