# Evidence: damage pipeline order

Spec: [Damage pipeline order](../../specs/combat-damage-pipeline-order.md)

## Identity and reuse

Evidence build: original Lord of Destruction Expansion 1.13c, Windows x86. Artifact ids `LOD113C-D2GAME`, `LOD113C-D2COMMON` and `LOD113C-FOG` refer to the [source manifest](../source-manifest.md). Prior readings retain their 2026-10-12/13 dates, scope and limitations. Newly fingerprinted analysis files identify the available original specimens; historical loaded DLL fingerprints, effective-file hashes, physical archive origins and complete override inventory were not captured. No identity check confirms gameplay.

Original spec/evidence/gap files were archived privately and compared byte-for-byte before edits on 2026-10-14. No plan outside gaps, code or test cited the spec filename or DPO prefix in the consumption search. Current ids are retained without aliases. The single ordered-path boundary remains intact. The original full investigation and binary fixture encodings remain privately preserved; source references below identify the original modules independently.

## Semantic correspondence

Damage fields map to original byte offsets: flags 00, result 04(u16), physical 08, physicalPercent0C, fire 10, burnMagnitude 14, burnLength 18, lightning1C, magic 20, cold 24, poisonMagnitude 28, poisonLength2C, coldLength 30, freezeLength 34, lifeDrain 38, manaDrain3C, staminaDrain 40, stunLength 44, absorbHeal 48, total4C, flatReductionScale 54, hitClass 60, markCondition 64(u8), conversionType 65(u8), conversionPercent 68, overlayId6C. Record size is0x70. These encodings are investigation details, not a required layout.

Queue nodes are0x88 bytes: game 00, attacker type/id 04/08, defender type/id 0C/10, damage 14, next 84; unit queue headAC. Unit type/class/id/mode 00/04/0C/10, monster data 14, seed 20/24, path2C, stat list 5C, inventory 60, game 80, event list 90, last attacker +0x9C/+0xA0, `aliveExclusionCondition` byte +0xC6 mask 1, last-attacker-presenceC8 mask 0x20000, pending eventsDC. Path target 58, room1C; game difficulty6D, Expansion 70, frame +0xA8. Monster-data 16 mask 8 is the neutral special-stun condition; its producer remains DPO-Q28. Public vectors use semantic values, while the privately archived originals preserve raw encodings.

## Dispatch and output anchors

D2Game RVA 0x79A10 -> frame-event table RVA 0x1079D0 -> RVA 0x79170 -> do entry0xA2060 -> dispatch0xA19A0 -> do table0x1074A8 slot 1 -> basic attack0xA2C40. SCD-R06/R09a/R14a own the upstream contract. Basic attack obtains target through D2Common #10392 ->0x348F0, computes hit0xDE5A0, before-state 0x922A0, builder0xDD450, after-state 0x91E30, queued completion0xDDDE0. DDDE0 calls scale/mitigation0xDC0B0 and enqueue0xDAB90. Queue consumer0xDED30 calls apply0xDE0C0, durability0xDEBA0, events, reflection0x98A70, reaction0xDCF20, removal0xDA6D0. DE0C0's resource setters and event/status calls are output anchors. The alternate early branch and monster/callback roots remain DPO-Q19/Q20.

Entry profiles for DPO-R08c/d: queued DDDE0 -> DC0B0 -> DED30 -> DE0C0(flag 0) has one eligible run; missile3AE10 / callbackA10B0 -> DE0C0(flag 1) one;755C0 -> DC0B0 -> DE0C0(flag 1) two; F42D0 -> DC0B0 -> DE0C0(flag 0) one;4F450/99290 -> DE0C0(flag 0) zero. These source profiles distinguish original callers without imposing binary entry names as APIs.

## Historical public records

The following dated records are reused, not fresh verification. Analyst labels, including known wrong labels, are navigation aids only. Original module-relative references and resolved ordinals identify the evidence.

### Historical dispatcher evidence

## Dispatcher correction

Targeted contract reconciliation, 2026-10-13. Evidence build: LoD Expansion 1.13c.
The 2026-10-13 reading covered D2Game+0x37030 and +0x37110 in full, the event-11 dispatch branch of +0xDC0B0, and the Fog+0x1CCC0 storage-release forwarder. The dispatcher/removal disassemblies had no undecoded instruction regions. No gameplay action or trace was collected.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | 0x37030 | none | ProcessResourceCallbacks (analyst label) | IEP-R07, DPO-R23, callback inputs and successor reload | disassembly reading; recorded 2026-10-13; D2Game.dll RVA 0x37030 |
| `LOD113C-D2GAME` | 0x37110 | none | RemoveSkillResourceBuffer (analyst label) | IEP-R05 consumed by IEP-V08's removal fixture | disassembly reading; recorded 2026-10-13; D2Game.dll RVA 0x37110 |
| `LOD113C-D2GAME` | 0xDC0B0 | none | ApplyDamageModifiers (analyst label) | DPO-R09 and DMI-R03 event-11 dispatch placement | disassembly reading; recorded 2026-10-13; D2Game.dll RVA 0xDC0B0 |
| `LOD113C-FOG` | 0x1CCC0 | FOG #10046, mapping preserved from mitigation evidence | DeallocateResourceBuffer (analyst label) | storage-release boundary; forwarder only | disassembly reading; recorded 2026-10-13; Fog.dll RVA 0x1CCC0 |

**Instruction findings.** In RVA 0x37030, the dispatcher loads the requested event into edx; it is
unchanged before the callback call at instruction site. ecx supplies the game. The five stack inputs
are owner, other unit, record, entry registrationValueA, entry registrationValueB. Thus the language-independent
callback contract has seven inputs: game, event ID, owner, other, record, registrationValueA, registrationValueB.

The dispatcher saves the current entry's successor before the event test. A non-matching entry
continues from that saved successor. A matching callback runs with its running flag set;
it clears the flag, then reloads the current entry's successor after the callback.
The saved result is replaced by each invoked callback's result. The loop does not stop on a
nonzero result. No match, empty list and null owner yield result 0.

The matching entry is unlinked after the callback when its removal flag is set or entry type
is 0. In RVA 0x37110, removal of an entry that is not running updates its predecessor's next
link and its successor's previous link immediately; removal of a running entry is deferred.
Therefore a callback removing its next entry through IEP-R05 changes the successor reloaded
by the dispatcher. IEP-V08 checks this with supplied registrations and callbacks.

**Anchors and ownership.** The root chain is the damage-pipeline spec's preserved dispatch
chain, reaching the shared entry; RVA 0xDC0B0's event-dispatch branch dispatch event 11 on the defender,
then continue to the strip without reading the result. Output anchors for this correction
are the callback invocation, saved dispatcher result, and event-list links in 37030/37110.
IEP-R07 owns traversal; DPO-R23 owns combat dispatch points and consumes that contract.

**Semantic mapping.** Owner list = Unit +0x90; entry eventId = byte +0x00; running and
removeAfterCall = +0x02 masks 1 and 2; entryType = +0x04; ownerKey = +0x08;
registrationValueA/registrationValueB = +0x0C/+0x10; callback = +0x14; previous/next = +0x18/+0x1C.
The record is passed by reference. DPO-V26's physical/coldLength fields map to +0x08/+0x30,
and stat 153 is the original stat 0x99 used by DPO-R10. These encodings are evidence only.

**Correction rationale.** The earlier IEP-R07 pseudocode used a successor read before the
callback for every entry. That is correct only for non-matches; matching callbacks reload it.
The earlier DPO-R23 and DPO-V26 callback descriptions omitted the event ID carried in edx.
The superseded texts are preserved in the private audit. No existing numeric vector result or rule ID was changed.
New IEP-V08 distinguishes the mutation and argument cases; DPO-V26 now names the event-ID
argument. Mitigation consumes the reconciled contracts, resolving its dispatcher disagreement.

**Limits.** The release import's allocator internals remain outside the traversal contract;
the Fog forwarder's downstream routine was not decoded in this pass. Actual item registration,
the stat resolver and callback-specific effects retain their own evidence labels. No runtime
confirmation or full-spec verification is implied. This is a targeted dispatcher revision;
unchanged legacy sections have not undergone a full format migration or fresh source audit.

The historical omitted-event-argument and successor-read discrepancy is described above; the current dispatcher contract is IEP-R07. Superseded contract text is preserved privately rather than serving as an additional public contract.

## Historical provenance

These are the original investigation records and dates, preserved without renewing their claims.

### Original function references


The table identifies original module-relative functions, resolved import ordinals, supported rules and the scope of each dated reading. Reconstruction-assisted readings mean the reconstructed instructions were checked against the original machine code; no private source distribution is required.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | 0x79A10, 0x79170 | n/a | (player frame-event dispatch and handler; names per SCD) (analyst label) | section 2 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x79A10`, `RVA 0x79170` (per SCD-R09a; not re-read here) |
| `LOD113C-D2GAME` | 0xA2060 | n/a | (frame handler; name per SCD) (analyst label) | section 2 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xA2060` |
| `LOD113C-D2GAME` | 0xA19A0 | n/a | (do dispatch; name per SCD) (analyst label) | section 2 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xA19A0`  |
| `LOD113C-D2GAME` | 0xA2C40 | n/a | `ExecuteWeaponCombatAction` (analyst label) | section 2, R01a, R08d, R33 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xA2C40` |
| `LOD113C-D2GAME` | 0xA20C0 | n/a | (second route to 0xDED30) (analyst label) | section 2 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xA20C0` |
| `LOD113C-D2GAME` | 0x93040 | n/a | `CheckCombatFlagState_1080` (analyst label) | section 2, R01a, Q20 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x93040` |
| `LOD113C-D2GAME` | 0xDE5A0 | n/a | `GetUnitCombatFlags` (analyst label) | R01a, R01b | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDE5A0` |
| `LOD113C-D2GAME` | 0xDDE90 | n/a | (to-hit roll; SBA) (analyst label) | R01a (call only) | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDDE90` |
| `LOD113C-D2GAME` | 0xDB790 | n/a | (block/avoid; SBA) (analyst label) | R01a (call only) | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDB790` |
| `LOD113C-D2GAME` | 0xE0A60 | n/a | (hostility; PVP-R05a) (analyst label) | R01a, R13 (call only) | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xE0A60` |
| `LOD113C-D2GAME` | 0x922A0, 0x91E30 | n/a | `ProcessSkillType1Handler`, `ProcessSkillTypeHandlers` (analyst label) | R33 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x922A0`, `RVA 0x91E30` |
| `LOD113C-D2GAME` | 0x91B70, 0xDAA90, 0xE07A0 | n/a | (do-function helpers; grep for RNG only) (analyst label) | Q19 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x91B70`, `RVA 0xDAA90`, `RVA 0xE07A0` |
| `LOD113C-D2GAME` | 0xDD450 | n/a | `CalculateFullDamageStats` (analyst label) | R02 to R07, R21, R22, R23, R31, R32a | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDD450` |
| `LOD113C-D2GAME` | 0xDC530 | n/a | `CalculateDamageRange` (analyst label) | R02 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDC530` |
| `LOD113C-D2GAME` | 0xDCD80 | n/a | `CalculateDamageWithModifiers` (analyst label) | R04, R05 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDCD80` |
| `LOD113C-D2GAME` | 0xDBED0 | n/a | `CalculateScaledPercentBonus` (analyst label) | R02a, R04a, R05, R06a | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDBED0` |
| `LOD113C-D2GAME` | 0x37540, 0x37260 | n/a | `TransferWeaponStatOwnership`, `GetStatValue_9100` (analyst label) | R32a | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x37540`, `RVA 0x37260` |
| `LOD113C-D2GAME` | 0xDA890, 0xDA8C0, 0xDAD70 | n/a | `GetGameValue`, `ProcessMonsterBehavior_83a0`, `ProcessSkillAction_82f0` (analyst label) | R31 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDA890`, `RVA 0xDA8C0`, `RVA 0xDAD70` |
| `LOD113C-D2GAME` | 0x770E0 | n/a | `ApplyMonsterStatMultiplier` (analyst label) | R22 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x770E0` |
| `LOD113C-D2GAME` | 0xDDDE0 | n/a | `CalculateUnitDamage` (analyst label) | R01a, R08c, R08d, R12a | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDDDE0` (`db` runs) |
| `LOD113C-D2GAME` | 0xDC0B0 | n/a | `ApplyDamageModifiers` (analyst label) | R08a, R08b, R09, R10, R11, R12a | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDC0B0` (loop in `db` runs) |
| `LOD113C-D2GAME` | 0xDAC10 | n/a | (scale `P` selector; not re-read) (analyst label) | R08b | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDAC10` |
| `LOD113C-D2GAME` | 0xDA7B0 | n/a | `GetUnitStatValue` (wrong) (analyst label) | R10 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDA7B0` |
| `LOD113C-D2GAME` | 0xDBD00 | n/a | (per-type mitigation; DMI) (analyst label) | R11 (call only) | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDBD00` |
| `LOD113C-D2GAME` | 0xDAB90 | n/a | `AllocGameResource` (analyst label) | R08d, R12a | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDAB90` (`db` runs) |
| `LOD113C-D2GAME` | 0xDED30 | n/a | `ProcessUnitCombatInteraction` (analyst label) | R13a, R20, R23, R24a, R25, R26, R32b | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDED30` |
| `LOD113C-D2GAME` | 0xDA6D0 | n/a | `RemoveUnitResourceLink` (analyst label) | R13a | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDA6D0` (`db` runs) |
| `LOD113C-D2GAME` | 0xDE0C0 | n/a | `ProcessCombatDamage` (analyst label) | R13 to R19, R23, R30 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDE0C0` |
| `LOD113C-D2GAME` | 0xDFA90 | n/a | `IsUnitInValidState` (analyst label) | R13a, R16 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDFA90` |
| `LOD113C-D2GAME` | 0x15D0 | n/a | `GetMonsterDataFlags` (analyst label) | R13a | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x015D0` |
| `LOD113C-D2GAME` | 0x268E0 | n/a | `ProcessPlayerPassiveAbilities` (wrong) (analyst label) | R30 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x268E0` |
| `LOD113C-D2GAME` | 0xDBA40 | n/a | (leech; LSE) (analyst label) | R15 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDBA40` (entry gate and draw position only) |
| `LOD113C-D2GAME` | 0xDB570 | n/a | `AddUnitStatCapped` (analyst label) | R16 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDB570` |
| `LOD113C-D2GAME` | 0xDA970, 0xDA930 | n/a | `ModifyUnitStat_7d90`, `ModifyUnitStat` (analyst label) | R18 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDA970`, `RVA 0xDA930` |
| `LOD113C-D2GAME` | 0xDCC20 | n/a | `ProcessMonsterBehavior_b600` (wrong) (analyst label) | R19a, R27 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDCC20`  |
| `LOD113C-D2GAME` | 0xDC780, 0xDDAC0 | n/a | `ApplyColdDebuffEffect`; (freeze applier) (analyst label) | R19a, R28 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDC780` , `RVA 0xDDAC0`  |
| `LOD113C-D2GAME` | 0xDCAB0, 0xDC940 | n/a | (poison, burn appliers; CSA-R10 to CSA-R12) (analyst label) | R19a (call only) | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDCAB0`, `RVA 0xDC940` |
| `LOD113C-D2GAME` | 0x8EA80, 0x8E6F0, 0x90F30 | n/a | (event scheduler; node removal; monster update; the last two unread) (analyst label) | R19c, R28b | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x8EA80`, `RVA 0x8E6F0`, `RVA 0x90F30` |
| `LOD113C-D2GAME` | 0x37030 | n/a | `ProcessResourceCallbacks` (analyst label) | R09, R13a, R23 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x37030` |
| `LOD113C-D2GAME` | 0x9E290 | n/a | (item-event slot lookup; IEP) (analyst label) | R23 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x9E290` |
| `LOD113C-D2GAME` | 0xDEBA0 | n/a | `ProcessCombatStunEffect` (wrong) (analyst label) | R24a | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDEBA0` |
| `LOD113C-D2GAME` | 0x12EE0 | n/a | `ProcessItemDurabilityCheck` (analyst label) | R24a | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x12EE0`  |
| `LOD113C-D2GAME` | 0x98A70 | n/a | (reflection; not re-read) (analyst label) | R25 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x98A70` |
| `LOD113C-D2GAME` | 0xDCF20 | n/a | (reaction; HRC) (analyst label) | R26 (call only) | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0xDCF20` |
| `LOD113C-D2GAME` | 0x755C0, 0xF42D0, 0x3AE10, 0xA10B0, 0x92BB0, 0x4F450, 0x99290 | n/a | (other callers of 0xDC0B0 / 0xDE0C0; per the 2026-10-12 reading) (analyst label) | R08c, R08d | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x755C0`, `RVA 0xF42D0`, `RVA 0x3AE10`, `RVA 0xA10B0`, `RVA 0x92BB0`, `RVA 0x4F450`, `RVA 0x99290` |
| `LOD113C-D2GAME` | 0x9E420 | n/a | `DispatchMissileHandler` (do-table user from missiles) (analyst label) | section 2 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x9E420` |
| `LOD113C-D2GAME` | 0x372C0 | n/a | `GetUnitPrimaryWeapon` (analyst label) | R02b, R03b, R24a, R31c | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x372C0` (lines 29-60 only) |
| `LOD113C-D2GAME` | 0x1180, 0x11D0, 0x1480 | n/a | `GetRandomModulo`, `GetRandomModuloAlt`, `GetRandomInRangeMod` (analyst label) | all draws (NRF-R02) | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x01180`, `RVA 0x011D0`, `RVA 0x01480` |
| `LOD113C-D2GAME` | 0x14D0, 0xF44B0 | n/a | `CalculateScaledDivision`, `__alldiv` (analyst label) | R07, R08a (NRF-R05) | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x014D0`, `RVA 0xF44B0` |
| `LOD113C-D2GAME` | 0x1220 | n/a | `GetMonsterRecord` (analyst label) | R13a, R22 | disassembly reading; recorded 2026-10-13; `D2Game.dll RVA 0x01220` |
| `LOD113C-D2COMMON` | 0x348F0 | D2Common 10392 (thunk rva 0xA598) | `GetPathRoom` (wrong: returns `+0x58`) (analyst label) | section 2 | reconstruction reading checked byte-for-byte against original machine code; recorded 2026-10-13; `D2Common.dll RVA 0x348F0` |
| `LOD113C-D2COMMON` | 0x2F9B0 | 10009 (thunk 0xA844) | `SortTextListByKey` (wrong) (analyst label) | R13a | reconstruction reading checked byte-for-byte against original machine code; recorded 2026-10-13; `D2Common.dll RVA 0x2F9B0` |
| `LOD113C-D2COMMON` | 0x603C0 | 10237 (thunk 0xA7BA) | `CalcMissileBreakpointValue` (wrong: difficulty record) (analyst label) | R08a, R08b | reconstruction reading checked byte-for-byte against original machine code; recorded 2026-10-13; `D2Common.dll RVA 0x603C0` |
| `LOD113C-D2COMMON` | 0x229A0, 0x22970 | 10213 (thunk 0xA82C), 10120 (thunk 0xA7D2) | `GetItemRecordField106`, `GetItemRecordField108_22970` (analyst label) | R02b | reconstruction reading checked byte-for-byte against original machine code; recorded 2026-10-13; `D2Common.dll RVA 0x229A0`, `RVA 0x22970` |
| `LOD113C-D2COMMON` | 0x26D40 | 10865 (thunk 0xA970) | `IsItemDurabilityDepleted` (analyst label) | R24a | reconstruction reading checked byte-for-byte against original machine code; recorded 2026-10-13; `D2Common.dll RVA 0x26D40` (not read) |
| `LOD113C-D2COMMON` | 0x1A2D0 | 11104 (thunk 0xA7F6) | `SetStructField4C` (wrong: hireling predicate) (analyst label) | R05, R11, R12a | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x1A2D0` |
| `LOD113C-D2COMMON` | 0x1A610 | 10255 (thunk 0xA814) | `GetUnitSeedHi` (wrong: demon bit) (analyst label) | R11, R31b | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x1A610` |
| `LOD113C-D2COMMON` | 0x1A5B0 | 10239 (thunk 0xA508) | `SyncUnitPositionFromPath` (wrong: undead bits) (analyst label) | R11, R31c | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x1A5B0` |
| `LOD113C-D2COMMON` | 0x391F0 | 10830 (thunk 0xA514) | `GetUnitModifierStatValue` (alignment read) (analyst label) | R31 | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x391F0` |
| `LOD113C-D2COMMON` | 0x2FE10 | 10331 (thunk 0xA478) | `HaveLightResBonus` (wrong: room getter) (analyst label) | R13a | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x2FE10` |
| `LOD113C-D2COMMON` | 0x3C390 | 10057 (thunk 0xA808) | `CheckSpecialBodyLocation` (wrong: town test) (analyst label) | R13a, R13b | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x3C390` (callee 0x2C320 unread) |
| `LOD113C-D2COMMON` | 0x326C0 | 11138 (thunk 0xA754) | `IsUnitInAttackRange` (analyst label) | R01b, R13a | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x326C0` |
| `LOD113C-D2COMMON` | 0x325C0 | 11133 (thunk 0xA77E) | `GetUnitWeaponStyle` (analyst label) | R01b | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x325C0` |
| `LOD113C-D2COMMON` | 0x33D10 | 10866 (thunk 0xAE7A) | `ExchangeNestedField0xA8_0x14` (analyst label) | Q19 | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x33D10` |
| `LOD113C-D2COMMON` | 0x390F0 | 10913 (thunk 0xA3A0) | `GetDataTablePropertyStat` (reads stat 7) (analyst label) | R16 | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x390F0` |
| `LOD113C-D2COMMON` | 0x39430, 0x38B70, 0x38A80 | 10910 (thunk 0xA802), 10973 (thunk 0xA42A) | `GetMissileRecordField54` (wrong), `GetUnitBaseStat`, `STATS_GetStatValue` (analyst label) | every `Stat` read | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x39430`, `RVA 0x38B70`, `RVA 0x38A80` |
| `LOD113C-D2COMMON` | 0x33CD0, 0x37DB0 | 10494 (thunk 0xA68E) | (state test and its bit test; GDF-R04) (analyst label) | every `HasState` | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x33CD0`, `RVA 0x37DB0` |
| `LOD113C-D2COMMON` | 0x3A740 | 10887 (thunk 0xA472) | `MultiplyValuesBy5` (wrong: SetStat) (analyst label) | R16, R17, R18, R31a | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x3A740` |
| `LOD113C-D2COMMON` | 0x32C10 | 10689 (thunk 0xA83E) | `AllocateUnitPath` (wrong: overlay stat) (analyst label) | R22, R32b | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x32C10` |
| `LOD113C-D2COMMON` | 0x4F640 | 10804 (thunk 0xA80E) | `CleanupStatListModChain` (wrong) (analyst label) | R02b, R03b | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x4F640` (unread) |
| `LOD113C-D2COMMON` | 0x1FB80 | 10051 (thunk 0xA7C0) | `GetPlayerDataField2C` (analyst label) | R02b | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x1FB80` (unread) |
| `LOD113C-D2COMMON` | 0x24430, 0x24C30, 0x1E670 | 10744 (0xA430), 10711 (0xA9AC), 11139 (0xA7AE) | `GetUnitDirection` (item-type test), `GetItemTypeProperty`, `AllocateFieldTableBuffer` (slot item) (analyst label) | R24a, R31c | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x24430`, `RVA 0x24C30`, `RVA 0x1E670` (unread) |
| `LOD113C-D2COMMON` | 0x321E0, 0x38520, 0x44350, 0x50D40 | 10862 (0xA826), 10871 (0xA5D4), 10889 (0xAE68), 10709 (0xACF4) | `GetUnitWeaponRange`, `SetStructField0x24`, `BinkBufferSetHWND`, `GetWeaponDamageType` (analyst label) | R32b, R33, Q19 | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x321E0`, `RVA 0x38520`, `RVA 0x44350`, `RVA 0x50D40` (unread) |
| `LOD113C-FOG` | 0x1CD50 | Fog 10045 (thunk 0xA2E6) | `AllocateAndValidateResourceSlot` (allocator) (analyst label) | R08d | disassembly reading; recorded 2026-10-13; `Fog.dll RVA 0x1CD50` (unread) |

Ordinals were resolved with the skill's PE import/export directory analysis (PE import and export directories of the 1.13c DLLs). Thunks rva 0xA712 (R33 list-stat read) and 0xA7D8 (R31d) were not resolved.

**Call-site census** (2026-10-13 disassembly search for direct-call targets and encoded call instructions in original D2Game): callers of rva 0xDDDE0: 33 function regions (0x0C6B0, 0x273C0, 0x27820, 0x27A10, 0x289A0, 0x28BE0, 0x44AC0, 0x44D70, 0x44ED0, 0x45000, 0x451B0, 0x4C2A0, 0x4C720, 0x4CB00, 0x4CC50, 0x4DAC0, 0x4F7D0, 0x77790, 0x92940, 0x92A60, 0x92BB0, 0x93040, 0x93680, 0x95690, 0x98BA0, 0x98E90, 0x99D20, 0xA0E80, 0xA2C40, 0xB3EB0, 0xB41B0, 0xB4520, 0xB60C0); callers of rva 0xDED30: 35 function regions; callers of rva 0xDC0B0: exactly `D2Game RVA 0x755C0` instruction site, `D2Game RVA 0xDDDE0` instruction site, `D2Game RVA 0xDE0C0` instruction site, `D2Game RVA 0xF42D0` instruction site; callers of rva 0xDE0C0: 25 function regions (DPO-R08c); callers of rva 0xDD450: 0x44ED0, 0x45000, 0x92360, 0x92A60, 0x93040, 0xA2C40, 0xDDDE0; callers of rva 0x9E420: 0x39DA0, 0x40E10 (missiles). The do table VA 0x6FD274A8 is referenced from `D2Game RVA 0xA19A0` instruction site and `D2Game RVA 0x9E420` instruction site.

**Blind-implementer passes** (checklist section 2; one subagent with only this file and the two core specs, 2026-10-13). Run 1: 16 GUESS, 10 GAP; every item folded (ratio normalisation and the physical `mulDiv` form, bypass flags only on the player/hireling path, enqueue at the list head, the predicted-lethal sum not stored, the attacker mode-0 tests, gate 1 versus the room test, the pre-apply bit-0 gate of the marks and durability, dispatcher node semantics and callback arguments, durability walk skipping non-armour slots, crit order versus NRF-R04a, `E` of the cold applier, the table of callee results in section 3, section 5 rows, DPO-V01, DPO-V19, DPO-V26; new DPO-V28, DPO-Q28). Run 2: 14 GUESS, 10 GAP; every item folded (the `rangeRoll`-to-`[+8]` form in DPO-R02b, the hireling predicate's type test read in `RVA 0x1A2D0`, the poison-length else branch, `mulDiv` branch order by reference to NRF-R05, event 0xB unconditional, raw 8.8 stat access, the leech draw only with non-zero leech inputs, one durability draw with no loop, the dequeue read in `RVA 0xDA6D0`, gate inputs added to the vector conventions, DPO-V10, DPO-V12, DPO-V16, DPO-V26 inputs; new DPO-Q29; the `db`-run basis of DPO-R22 and DPO-R28a added to the caveats). Not re-run after the run-2 folds, so the pass has not yet returned an empty list and the spec stays `draft`.

**Search records.** No rule makes a completed absence claim.


| Vector | Version | Member | Date | Trace RVA | Records | Result |
|---|---|---|---|---|---|---|
| none (identity: do table) | 1.13c | owned vanilla 1.13c instance | 2026-10-13 | memory read `D2Game.dll+0x1074A8`, 16 bytes | n/a | slot 0 = 0, slot 1 = relocated module base + RVA 0xA2C40, slot 2 = +0xA3220, slot 3 = +0xA2B40: section 2 step 3 confirmed |
| none (data: 12-type table) | 1.13c | owned vanilla 1.13c instance | 2026-10-13 | memory read `D2Game.dll+0x102AB0`, 528 bytes | n/a | field offsets, markers and scale flags as in section 3 |
| none (data: durability slot table) | 1.13c | owned vanilla 1.13c instance | 2026-10-13 | memory read `D2Game.dll+0x10E628`, 100 bytes | n/a | count 7 and the seven (slot, weight) pairs of section 3 |
| none (data: names) | 1.13c | owned vanilla 1.13c instance | 2026-10-13 | game's own Storm-loader read of `data/global/excel/itemstatcost.txt` (40109 bytes) and `itemtypes.txt` (8775 bytes) | n/a | rows 17, 18, 25, 56-66, 72, 101, 120-122, 178, 180, 315, 325, 326, 332, 350 of `ItemStatCost.txt` and rows 45, 50, 57 of `ItemTypes.txt` as named in section 3 |

Mismatches: none (nothing compared).

**Status ledger.** Verified in 1.13c code (static): DPO-R01a, R02a, R03a, R04a, R04b, R05, R06a, R06b, R07, R08a, R08c, R08d, R09, R10, R11, R12a, R13a, R14, R15, R16, R17, R18, R19a, R19b, R21a, R21b, R22, R23, R24a, R26, R27, R28a, R29. Asm-only: R01b, R02b, R03b, R08b, R13b, R19c, R20, R25, R28b, R30, R31, R32, R33. Unresolved: R12b, R24b. Runtime-confirmed: none. Not found in original 1.13c binary analysis: none. Upgraded in this revision after a full read of the function and its callees: R04b, R06b, R09, R10, R19a, R21b, R23 (each has a new vector). Names verified from the 1.13c data files (section 3 `*` rows): stats 17, 18, 25, 0x38, 0x39, 0x3A, 0x3B, 0x3D, 0x3F, 0x41, 0x42, 0x48, 0x65, 0x78, 0x7A, 0xB2, 0xB4, 0x13B, 0x145, 0x146, 0x14C, 0x15E and item type 0x39.

**Caveats.**
- Text changed under a kept id (readings of 2026-10-13 that contradict the earlier text): DPO-R01a: on the basic attack the build and its draws run on a miss (earlier: "block/avoid/dodge stop the pipeline before any damage field exists"); DPO-R08d and DPO-R12a: the enqueue is unconditional (earlier: "nothing is enqueued" on a miss); DPO-R08a: the `mulDiv` argument order is `(P, v, 100)` (earlier `(v, P, 100)`; the branch description was already correct); DPO-R29: the order gains the dispatch points of DPO-R30 and DPO-R33 and the hit/miss split; DPO-R16: the max-life read is ordinal 10913 (stat 7) and the state-0x5C gate is named; DPO-R24a: the first defender draw's modulus is the table count 7; DPO-R13a: the last-attacker write (ordinal 10009) and the exact gate predicates (`killable`, rva 0xDFA90) are added; DPO-R23: list owners and the miss-path dispatch of events 7/3 added. DPO-V03 was rewritten accordingly.
- Labels changed: R04b, R06b, R09, R10, R19a, R21b, R23 from asm-only to verified (functions and callees read in full; vectors DPO-V21 to V27 added). R01, R03, R13 split into sub-ids. No label was lowered.
- Historical instruction-decoding limits: portions of D2Game+0xDDDE0, +0xDAB90, +0xDA6D0, +0xDC0B0, +0x326C0, +0xDB570, +0x770E0 and +0xDDAC0 originally relied on uncorroborated instruction listings. The targeted original-image record dated 2026-10-14 below resolves only its listed portions. Remaining reach and callee uncertainties are explicitly gated; an annotation beside encoded bytes is not independent verification.
- Label collisions: every original analysis label in the function table marked "wrong" does something other than its name says (the reach check is the only label found apt). Stat 18 (`item_mindamage_percent`) is the min-end percent and stat 17 the max-end percent of the physical roll in rva 0xDC530 (stack slots re-read), consistent with their names.
- Unread callees on verified paths: rva 0x372C0 and ordinal 10804 (R03a treats the weapon chance as an input), ordinal 10865 and 10711 (R24a treats "qualifies" and "throwable" as inputs), rva 0xE0A60 and 0x2C320 (R13a's gates 1 and 2, meanings per PVP), the appliers' interiors (LSE). Each is named in the rule as an input or a sibling's rule.
- The 1.13c assert file-name pointer `CONST_BASE + 0x1FF9` pushed by the cold and freeze appliers was read at `D2Game.dll+0xF9FF9` on the member and did not decode as a path (the bytes read `...delete...`); the source-folder claim of section 2 is therefore left open.
- Cross-spec contradiction to settle in the foundations (not edited here): NRF-R04a describes rva 0x12EE0 as stepping up to 10 times in a loop; `RVA 0x12EE0` lines 71-82 make exactly one inline step and one `div` (DPO-R24a).
- Vector DPO-V28 was added in this revision so that DPO-R11 has an executable vector of its own; the DMI vectors remain the arithmetic tests.
- The 12-type table, the durability slot table and the do-table slot were read from the running member's memory, that is from the loaded 1.13c `D2Game.dll` image; they are static data of the binary and were not re-read from the original analysis `binary` file.

## Targeted reader review

2026-10-13: a fresh smaller-model reader reviewed IEP-R01/R05/R07 and IEP-V08, DPO-R09/R10/R23
and DPO-V26, and DMI-R03/V19 with the revised gate and acceptance boundaries. It used only
behavioral specs/dependency contracts; no original-binary analysis, implementation code, live-game tools or web.
Results: MISMATCH 0, GUESS 0, GAP 0, required COVERAGE 0, GATED 2, HARDEST 3.
The documented gates were DMI-Q01 (stat integration) and DPO-Q11 (callback-specific draws).
The hardest cases were IEP-V08's mutation walk, DPO-V26's scale/dispatch/strip sequence and
DMI-V19's dispatch/mitigation sequence; no undocumented decision remained.

One advisory branch lacked a vector in this reviewed set: removal of the currently running
entry through IEP-R05. No change: IEP-V08 claims removal of the next non-running entry,
not coverage of deferred self-removal. Its stated assertions discriminate the intended cases.
The reader confirmed that removing the historical mitigation-dispatch conflict's conflict gate is justified by the owning contracts
now agreeing. It did not approve the full specs or verify original-game execution.

Document checks: 59 unaffected existing vectors unchanged; unrelated item-event rule text
preserved; original provenance retained with relocated links; implementation comparison
unchanged; closed dispatcher gap removed; stat-reader gate retained. All 151 local links and
anchors across the seven changed files resolved, with balanced fences and no trailing whitespace.
All three behavioral specs remain draft. No implementation tests or gameplay observations ran.


## Freeze payload clarification (2026-10-14)

DPO-V25 now expresses freeze with named attacker, defender and length roles, matching the normalized CSA-R01 contract. Original DE0C0 sets EAX to the defender and ECX to the attacker before calling DDAC0; its first role is the target and its second role the source. Those registers do not impose a language-independent API parameter order. No target/source role, numerical output, draw, or event order changed. The superseded binary-layout fixture is preserved in the private audit.

## Observation procedures

### 8.2 Observation vectors and reproduction procedure

Priorities: the lethal boundary (DPO-O01), the single doubling (DPO-O02), the mana-burn quirk (DPO-O03). Each on an owned 1.13c game, Normal difficulty, single player, tool-neutral.

- **DPO-O01** **lethal boundary (DPO-R12a, DPO-R17).** Measure: the defender's stat 6 (`hitpoints`, raw) immediately before and after one hit, and the `total` of the applied record. Expected: `after = before - total` when that is `>= 0x100`, else 0, with no dependence on the predicted-lethal bit. Procedure: a level-1 character of any class in Act I (Blood Moor) attacks a monster with the basic melee attack (the character is the attacker on the player basic-attack path of section 2; the monster is the defender, so monster frame-event roots gated by DPO-Q20 are not needed); read the monster's stat 6 on the server unit the frame before the hit and the frame after, and the record's `total` at the entry of the apply (resource/status apply, supplied apply record); repeat until a hit lands with `before - total` in each of the ranges `>= 0x100`, `1..0xFF` (death) and `<= 0` (death). The build-side pre-flag is read as bit 0x2 of `result` at the same entry.
- **DPO-O02** **crit draws and single doubling (DPO-R03a).** Measure: the attacker's seed `seed` before and after the build (damage builder entry and return) and `physical` with result bit 0x2000 at return. Expected: the number of seed steps equals 1 (physical) + one per crit source with chance `> 0` up to and including the first success + element draws + 1; `physical` is doubled at most once. Procedure: an Amazon with `passive_critical_strike` (stat 0x151) from Critical Strike at a known chance and a weapon with no other crit source, attacking Fallen with a basic attack in the Blood Moor; repeat for configurations (0x151 only), (0x8D only, from an item with `item_deadlystrike`), (both); count steps per hit.
- **DPO-O03** **player mana leech burns the defender (DPO-R18).** Measure: the monster defender's stat 8 (`mana`) before and after one hit by a player with mana leech `M` percent and no `manaDrain` roll (player attackers never roll it). Expected in Normal (`ManaStealDivisor = 1`, GDA-R05): a decrease of `(M << 6)` raw units when the leech ran, for example `M = 5` → 320 raw (1.25 mana), set to 0 if that leaves less than 0x100. Procedure: any class wearing one item with `manadrainmindam` (stat 0x3E) = 5 and no other leech, hitting a monster that has mana (a Fallen Shaman); read the Shaman's stat 8 on the server unit before and after.


## Targeted original-image recheck (2026-10-14)

Original D2Game SHA-256 matched LOD113C-D2GAME. New instruction listings decoded these whole byte extents, corroborating the previously ambiguous byte-run interpretations; this adds instruction evidence, not new gameplay execution or a full callee-chain claim:

| Original function/extent | Checked decisions/writes | Affected assertions |
|---|---|---|
| RVA 0xDDDE0, length 0xB0 | hit/mask 0x8380 gates;already-built test;call0xDC0B0 at 0xDDE2D;sum and monster-only drain addition;strict signed comparison;unconditional call0xDAB90 at 0xDDE7F | DPO-R01a, R08c/d, R12a |
| RVA 0xDAB90, length 0x80 | player/monster defender gate;copy record;store old head as next and new node as head | DPO-R08d, R12a;allocator integration stays Q30 |
| RVA 0xDA6D0, length 0x80 | four identity comparisons;matching head/predecessor unlink;continue after removal | DPO-R13a;allocator release integration not inferred |
| RVA 0xDB570, length 0x70 | amount/dead/state 92 gates;max/current reads;cap/set;actual returned increase | DPO-R16 |
| RVA 0x770E0, length 0xF0 | seed step;unsigned byte compare at 0x77164 and jae;the six shifts;marker/overlay request split | DPO-R22 |
| RVA 0xDC0B0, length from function extent | all four unrolled scaling blocks and the three-iteration loop: positive/scale flag tests, large-value division branches and write;matching argument order P, value,100 | DPO-R08a;prior remaining path reading reused |
| RVA 0xDDAC0, length 0x1F0 | fallback at 0xDDC9D passes supplied freeze length to cold call0xDC780 at 0xDDCA0 | DPO-R28a |

The previously flagged instruction decodings agree with these original bytes. Compatible prior reading remains the basis for unaffected branches; no vector value, seed or expected output changed. Original public caveats above are historical and this record supersedes only the listed decoding uncertainty. Reach's uncorroborated addition and unread callees remain gated, not promoted.

New targeted disassembly reading of scale selector D2Game 0xDAC10 identified its branch order and source operands; local imports for its special class/flag tests remain unread, so DPO-R08b remains asm-only. New reflection reading0x98A70 established signed (percent+4)/8 for player/hireling, physical mulDiv(percent, incoming physical,100), result 0x4021 and hitClass0x8D. This corrects earlier plain division/blank-record wording; the zero-clear byte-run and dependent apply/reaction integration remain gated. No existing vector asserts reflection results. Complete weapon/source investigation was attempted but one requested original module locator was incorrect; that failed read provides no evidence and Q27 remains.

Migration self-review keeps every original rule/vector id and all original vector numeric traces/results with semantic fixture roles. Current contract removes historical aliases, resolved questions and private/binary layouts; original complete encodings remain privately archived. This exact regenerated contract still needs a fresh full blind-reader pass; historical reader results above do not satisfy it. The mechanical check's final result is recorded after cleanup. No implementation/gameplay test, commit or publication occurred.

### Final self-review checks (2026-10-14)

The current contract has 29 executable vectors. DPO-V01..V28 retain their original numeric traces/expected values with semantic fixture roles. DPO-V29 adds a draw-free queued-miss case with exact reaction/event/removal log; this gives the verified reaction-placement rule a direct executable assertion beyond V20's build trace. The final reader must recompute this new case as well as the retained cases.

Baseline `check_spec.py combat-damage-pipeline-order`:367 errors, 51 warnings. After semantic cleanup:0 local errors and 0 warnings; two stale sibling references to the removed resolved-question entry remained for shared reconciliation. No executable implementation tests or gameplay observations ran. Original-image instruction checks are source verification, not gameplay confirmation. The current contract still requires a fresh full blind-implementer pass before review completion.

The self-review made execution boundaries explicit without changing retained numeric expectations: V01 ends at final death-flag processing, before deferred/post-apply work; V10 ends at the gate/apply-event prefix, before clamp. The earlier V10 statement that later stages change nothing was too broad with unnamed life defaulting to zero: R19b would set lethal. Its verified assertion is the gate prefix, not final life/death. V10's no-events assertion names apply events, not internal event 11 of its supplied scaling operation. Full-path acceptance remains gated and these are partial stage vectors. Shared omitted-state/registration/Crit/inventory/qualification defaults are now explicit so no extra dependency draw is silently assumed. A fresh reader must evaluate these clarified boundaries.

Final dependency reconciliation removes dependency references absent from the current numeric-foundation contract and uses current NRF-R01/NRF-R02/NRF-R05/NRF-R06a/NRF-R08a/RLI-R03/RLI-R02a contracts. Current DPO parent ids replace historical subpart references for R31/R32. No original arithmetic or retained expected vector value changes.

Final mechanical run (2026-10-14): `python .agents/skills/spec-writing/tools/check_spec.py combat-damage-pipeline-order` returned 0 errors and 0 warnings. Source verification and documentation lint ran; no implementation tests or gameplay observations ran. Fresh full blind-reader review remains pending.

Vector-field completion: retained V01..V28 now list their independent Inputs, Seeds, ordered Draws and Outputs directly, so recomputation need not recover inputs from the hand trace. The original numeric expected assertions remain intact. Stage-local assertions are distinguished from logged dependencies and no new gameplay observation is claimed. V08 retains a physical-subtraction-only 100-percent check; V12 explicitly invokes final-death processing for its lethal-bit assertions.

V27 now supplies a concrete seed and endpoint values for the actual verified build arithmetic: physical endpoints 256/512 give 351 after one draw 95, then mandatory draw(1) gives 0. Both hits and misses consume these two steps. Original event/scale-count expectations are unchanged. Only unread dependency interiors are draw-free supplied operations;the build itself is not replaced by a stub. This completes the missing seed/draw specification and does not claim full integration across the documented gates.

## Fresh full blind-implementer pass 1 (2026-10-14)

The fresh reader recomputed all 29 submitted executable vectors using the behavioral contract and cited dependencies. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 24, COVERAGE 19, HARDEST 3. This is a full pass, not a delta check. The recorded gates are DPO-Q01, Q02, Q04, Q06, Q07, Q11, Q12, Q13, Q15, Q16 and Q18 through Q31. They remain open;review does not promote evidence or confirm gameplay.

The advisory coverage triage below distinguishes a missing result from repeated bindings of an already exercised operation. No-change rows do not claim exhaustive branch coverage.

| Finding | Outcome | Reason or exact edit |
|---|---|---|
| DPO-R01b reach variants | no change | Gated by DPO-Q23;no verified reach-producer assertion is added. |
| DPO-R02b weapon selectors/factors | no change | Gated producer integration DPO-Q27;the verified supplied-endpoint operation remains separate. |
| DPO-R04a positive lightning/magic | no change | Repeated elementRoll bindings, not an additional arithmetic branch. V05/V24 exercise positive maximum, mastery, prior clamp and skipped maximum;R04b specifies the field/stat mapping. |
| DPO-R05 mana/stamina skip flags and remaining classes | no change | V06 exercises the generic roll, skip, prior-double-count, player and nonzero-hireling paths. Other drain bindings and classifier values 2/3/4 use those same result recipes;no exhaustive binding claim is made. |
| DPO-R06a mastery/prior poison | no change | V07 exercises the positive/raw poison path;rangeRoll's mastery and prior operations are exercised in V24/V06. Nonzero parameter values introduce no new control branch;poison units remain explicit. |
| DPO-R07 unlisted conversion types | spec edit | V08 adds fixed type 11 fire/burn-length and type 12 cold/freeze-length cases. Type2 is covered by V08 random selection, type 3 by V20;reserved/out-of-range no-add cases have the same result as V08 type 7. |
| DPO-R08a negative/zero scale | no change | V09 already asserts negative cold unchanged and has zero default fields;PVP-V11 also records zero/negative supplied-value results. |
| DPO-R08c/d remaining entry profiles | no change | PVP-V08 records unflagged unchanged, flagged once, unscaled local profiles and direct-then-flagged twice. V27 covers the queued profile;the direct-then-unflagged route uses the already covered one-run shared stage plus unflagged preservation. Upstream caller integration remains gated. |
| DPO-R09 callback mutations | no change | Mutation/traversal belongs to IEP-R07 and IEP-V08;DPO-V26 asserts this caller's position and shared-record arguments. Callback-specific execution/draws remain DPO-Q11. |
| DPO-R12a regular-monster/hireling totals | spec edit | New V30 separates stored total from prediction for regular monster, hireling and player, with exact totals, result flags and queued record. |
| DPO-R13a no match/other apply gates | no change | V10 covers accepted/rejected prefix, result-bit clearing, town rejection, dead return and failed reach;V29 covers matched queue consumption. Remaining predicates select documented no-effect/return families already exercised;predicate acquisition is gated by Q23/Q30. |
| DPO-R16 negative absorb/state/dead | no change | V12 covers the no-write/return 0 family through a direct dead case and positive/capped healing. Other rejection predicates lead to that same result;state meaning consumes LSE-R11. |
| DPO-R18 negative drains | no change | Same nonpositive guard branch as the zero-term case in V13;no resource write or extra draw is a new result family. Positive/floored outcomes are already explicit. |
| DPO-R19c regeneration scheduling | no change | Gated scheduler/update integration DPO-Q24;no verified deferred-effect assertion is added. |
| DPO-R21b negative/wrapping burn | spec edit | V23 adds prior-4, ratio 128 producing x=-196, zero added term, and unchanged magnitude-4. This distinguishes clamping from a negative quotient. Wrapping is an unconditional numeric primitive, not an extra burn branch;the existing product/add order is retained. |
| DPO-R23 flags 128 attacker-event suppression | spec edit | New V31 compares flags 32 with 160 and asserts exact 6/2 versus2-only dispatch logs at the apply prefix. |
| DPO-R25 reflection | no change | Gated by DPO-Q04;the partial source correction is not accepted as complete reflection integration. |
| DPO-R26 reaction placement | no change | V29 already directly asserts reaction after 7/3 and before removal on miss;HRC owns the reaction interior and reflection remains gated. |
| DPO-R29 full-path draw integration | no change | Existing stage vectors do not establish full integration. Q11/Q22/Q25 and other explicit producer/dependency gates remain;no guessed callback/handler draw totals are supplied. |

HARDEST triage: V10 apply gates/result bits, no change because its prefix endpoint and every case's expected gate effects are explicit;V27 build draws/dispatch placement, no change because its actual build draws and supplied dependency interiors are separately stated;V01 predicted versus final lethal, no change because the masked strict prediction and raw below 256 final-death threshold are separately traced. The reader reported no ambiguity in these cases.

The required coverage edits change only V08, V23 and add V30/V31;Id-map references and the executable-vector range now include 31 vectors. Existing expected assertions, rule behavior, ownership, evidence labels and gates are unchanged. No new source campaign, implementation test or gameplay observation ran. Because vector inputs/assertions changed after full pass 1, a fresh full reader is required for the resulting contract;this pass alone does not complete its review.

## Fresh full blind-implementer pass 2 (2026-10-15)

The fresh reader examined all 31 executable vectors using the contract and cited dependencies. The contract has 48 rule/subrule Id-map rows, counted before and after these edits; no rule was added. Reported counts: MISMATCH 1, GUESS 0, GAP 0, DISCRIM 1, GATED 23, COVERAGE 15, HARDEST 3. This was a full review. Each reported finding is triaged below; a reader report is checked against the contract rather than treated as behavioral evidence. This pass collected no new source or gameplay evidence and changes no rule label.

| Finding | Outcome | Reason or exact edit |
|---|---|---|
| MISMATCH: DPO-V01(e), prediction for T=2816 | no change | The reported calculation `2816 & ~255 = 2560` is incorrect: 2816 is 0xB00, already a multiple of 256, so `0xB00 & ~0xFF = 0xB00 = 2816`. Defender life 2560 is 0xA00. R12a's strict `2816 > 2560` is true, preserving prediction set, final life 0 and lethal set. No input or expected output is corrected. |
| DISCRIM: DPO-V24 skipped lightning/magic mastery reads | spec edit | At its supplied-input boundary, Stat now logs every requested unit/id/value. Exact output log is ids 49,48,329,51,55,54,331,53 with supplied values. Reading mastery 330/357 adds an entry even when its value is 0, so the previously invisible incorrect alternative fails. This asserts the existing R04b read condition, not an additional gameplay effect. |
| Acceptance range omitted DPO-V29..V31 | spec edit | Section 9 now names DPO-V01..V31. This repairs stale acceptance wording; it does not extend any vector's supplied-input boundary. |

The reader-listed GATED entries are DPO-Q01, Q02, Q04, Q06, Q07, Q11, Q12, Q13, Q15, Q16, Q18, Q19, Q20, Q21, Q22, Q23, Q24, Q25, Q26, Q27, Q28, Q29 and Q30. Outcome for every entry: no change, because the corresponding question, Id-map gate and [gap-file row](../../plans/gaps/combat-damage-pipeline-order.md) already specify the blocked assertion or integration. An Affects reference does not gate the other supplied-input assertions. The separate observation-only DPO-Q31 remains open even though it was not counted in the reader's 23 entries. GUESS 0 and GAP 0 require no edit.

| Advisory COVERAGE finding | Outcome | Reason or exact edit |
|---|---|---|
| DPO-R04a positive lightning/magic roll | no change | V05/V24 exercise the positive-maximum range, mastery and prior-clamp results. Lightning/magic invoke the same elementRoll arithmetic; these are additional field bindings, not an unshown result branch. |
| DPO-R04b lightning/magic id pairing | no change | R04b states their max/min/mastery bindings. V24's exact read log now checks their maximum ids and skips their min/mastery reads; the positive endpoint/mastery recipe is exercised by fire/cold. This is not exhaustive positive-value coverage of every binding. |
| DPO-R05 monster mana/stamina and unsupported attacker | spec edit | V06 adds independent case 6: type 2 with positive drain/bypass stats leaves every drain and flags 0, seed S1 unchanged, no draw. This distinguishes the unsupported-type return from player additions. Positive monster mana/stamina reuse the life-drain roll recipe and flags 8/16 reuse the already exercised skip recipe; no additional arithmetic branch is claimed. |
| DPO-R07 fixed type 3; no-add types 6/8/9 | no change | V20 selects type 3 and asserts its magic addition after the same type-selection point. V08's type 7 exercises subtract-physical-with-no-add; 6/8/9 use that same result recipe. Fixed versus randomly selected entry does not change those handlers. |
| DPO-R11 null defender | no change | Null and non-monster defenders share the flag-1024 branch. V28 case 5 supplies the non-monster branch with bypass 1 and all twelve logged calls. Null introduces no new result recipe within the supplied mitigation-dispatch boundary. |
| DPO-R12a nonphysical fields and wrapping total | no change | V01/V30 distinguish stored total, predicted sum and strict comparison. The remaining named damage fields contribute to the same unconditional wrapping sum. Wrapping is owned by NRF-R06a; there is no additional local control branch or different sum recipe. This does not claim exhaustive field-binding coverage. |
| DPO-R13a invalid/dead/unkillable defender; no matching queue entry | no change | V10 case 6 covers unchanged early return before last-attacker write; the other defender rejections have that same effect. No-match also returns without effect; V29 covers the effectful match path. Predicate acquisition/queue integration retain Q23/Q30. No new result family requires another vector. |
| DPO-R16 nonpositive absorb/state 92 | no change | V12's direct dead case asserts return 0 with no write, the same result as these guards; positive and capped healing results are also exercised. State 92's meaning belongs to LSE-R11. |
| DPO-R18 negative drain | no change | V13's zero manaDrain exercises the nonpositive guard: unchanged resource and no draw. Negative values use this same branch; positive fractional and negative-remainder outcomes already have exact assertions. |
| DPO-R19a zero-length applier calls | spec edit | V25 adds an independent all-zero-term case with the exact five-call zero-argument log, no resource writes and no draw. It rejects omitting caller dispatch merely because a supplied applier receives zero; CSA owns the early returns inside those appliers. |
| DPO-R22 doubling lightning/cold/magic | no change | V16 covers successful six-field doubling with nonzero physical/fire/poison, equality failure, marker/overlay alternatives and unchanged drains/lengths. The other three fields use the same unconditional shift in that success branch; no additional result recipe is introduced. |
| DPO-R24a full armor location order | no change | V17 fixes table-index modulus 7, cyclic traversal across empty entries, strict remainder subtraction at weight equality, selected location 3 and qualification-dependent third draw. Populating the other named locations repeats this weighted-selection recipe; their fixed bindings remain in section 3. Full seven-location population is not claimed. |
| DPO-R26 queued hit reaction | no change | V29 explicitly logs reaction between 7/3 and removal. R26's call and ordering are identical for a hit that reaches that point; hit-only intervening reflection belongs to the explicitly gated R25/Q04 contract. V27 supplies reaction without asserting its interior or full reflection integration. |
| DPO-R28a nonzero monster cold effect/eligible freeze | no change | CSA-V09 exercises nonzero monster cold effects and CSA-V10(e)/CSA-V11 exercise eligible freeze with no cold fallback draw. These interiors belong to CSA; DPO-V19 separately discriminates the caller's additional cold/fallback draws. |
| DPO-R29 later full-path draw order | no change | V20 explicitly ends at build, V27 logs supplied dispatch placement and V29 logs queue-tail ordering. Complete callback, modifier, state-handler and reflection integration remains Q11/Q22/Q25/Q04 and the other named gates; no guessed later draw total is added. |

HARDEST outcomes (3): R12a/V01/V30, no change because prediction and stored total are separately defined and traced, including the rejected masking mismatch above; R13a/V10/V31, no change because gate-prefix endpoints, flags and event logs are explicit; R23/R29/V27, no change because the actual physical and modulus-1 build draws are separated from supplied draw-free dependency interiors and dispatch placement. The reader reported no missing decision for these entries.

The resulting contract retains 31 executable vector ids and 48 rule/subrule rows. Changed material is V06 case 6, V24's supplied stat-read log and exact read assertions, V25's zero-term case, and section 9's acceptance range. All other existing numeric results are preserved. Rules, ownership, labels, questions and gap boundaries are unchanged; no gap-file edit is needed. Because vector inputs/assertions and acceptance wording changed after this full review, full reader pass 3 is required before review completion can be claimed.

Validation on 2026-10-15: `python .agents/skills/spec-writing/tools/check_spec.py docs/specs/combat-damage-pipeline-order.md` returned `0 error(s), 0 warning(s)` after these vector and acceptance edits. This is document lint, not execution of the vectors or gameplay confirmation.

## Publication provenance cleanup, 2026-10-15

The public record retains original artifacts/functions, dates, findings, correction rationale and limitations. Private reconstruction filenames, internal source-line references, working priorities, superseded binary fixtures and session-specific relocation details remain in the separately preserved audit. This cleanup changes no behavioral assertion, rule label, vector input/output or gate.

## Effective-file named-cell corroboration, 2026-10-15

Read-only collection used the owned game's Storm loader on the currently reported `lod-113c` member, excluding separately reported modified-game members. The historical vanilla source-profile identity is reused; refreshing the local launch-version console failed with connection refused. The prior patch inventory reported file override disabled, but no complete current override/archive inventory, supplying physical MPQ or newly fingerprinted loaded-module identity was captured. These limits prevent claiming fresh pristine-installation identity. No gameplay action or trace was collected.

Both virtual files were requested at offset 0 with whole-file length, returned in full, decoded from their complete bodies and parsed by header name. Text-cell agreement does not establish matching compiled runtime tables. The following named cells corroborate section 3's existing name/id claims; no behavioral label, original binding or absence claim is promoted.

| Effective virtual file | Returned / total bytes | Coverage | SHA-256 of returned file |
|---|---|---|---|
| `data/global/excel/itemstatcost.txt` | 40109 / 40109 | Complete header and all cited rows | `ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375` |
| `data/global/excel/itemtypes.txt` | 8775 / 8775 | Complete header and all cited rows | `d399dc5d0f3a3e1a4f1eb611dabef8311b7025392c416f58dc26baff93af3550` |

ItemStatCost rows were selected by explicit decimal `ID`; the selected zero-based file data-row index, excluding the header, equals that ID. The raw `Stat` cells are:

| Explicit raw ID | Raw Stat cell |
|---|---|
| 17 | `item_maxdamage_percent` |
| 18 | `item_mindamage_percent` |
| 25 | `damagepercent` |
| 56 | `coldlength` |
| 57 | `poisonmindam` |
| 58 | `poisonmaxdam` |
| 59 | `poisonlength` |
| 61 | `lifedrainmaxdam` |
| 63 | `manadrainmaxdam` |
| 65 | `stamdrainmaxdam` |
| 66 | `stunlength` |
| 72 | `durability` |
| 101 | `skill_poison_override_length` |
| 120 | `item_damagetargetac` |
| 121 | `item_demondamage_percent` |
| 122 | `item_undeaddamage_percent` |
| 178 | `unit_dooverlay` |
| 180 | `damage_vs_montype` |
| 315 | `firelength` |
| 325 | `progressive_tohit` |
| 326 | `poison_count` |
| 332 | `passive_pois_mastery` |
| 350 | `modifierlist_skill` |

ItemTypes has no explicit ID column in this record. These zero-based file data-row indices exclude the header; the existing loader/name convention is consumed separately from GDF-R07.

| File data-row index | Raw ItemType cell | Raw Code cell |
|---|---|---|
| 45 | Weapon | `weap` |
| 50 | Any Armor | `armo` |
| 57 | Blunt | `blun` |

## Fresh full blind-implementer pass 3, 2026-10-15

All 31 vector families were recomputed: MISMATCH 0, GUESS 1, GAP 0, DISCRIM 0,
GATED 12 grouped, COVERAGE 20, HARDEST 3. The grouped gates restate the existing
assertion-level limits; no additional gate or evidence promotion follows this report.

GUESS V10: the reader interpreted the explanation of outputs saying "no events" as a
global statement conflicting with cases that dispatch 5/1. Clarified that the phrase applies
only where a case's Outputs use it, and means no apply-dispatch requests; empty registrations
mean no callbacks, not no dispatch request. Case inputs, call logs and numeric outputs are
unchanged. This resolves the wording ambiguity without supplying a new default.

| Advisory COVERAGE group | Outcome |
|---|---|
| R01a remaining avoidance/reach/to-hit variants | No change: additional avoidance values repeat the fixed mapping recipe; V03 covers successful, avoided and refused hand-off results. The producer interiors remain SBA-owned or gated |
| R02a physical ratio / large value | No change: NRF-R05 owns asymmetric mulDiv, exercised through DPO-V09 with a large second operand. The physical rule fixes ratio/value order and divisor 128; changing these bindings does not add a different division recipe |
| R04a positive maximum with reversed/equal adjusted range | No change: shared rangeRoll's no-draw branch is already exercised by V04 case 2. Elemental callers consume that same operation |
| R05 other skip flags / drain fields / classifier values | No change: V06 covers roll, skip, prior double-count, player/hireling and unsupported-type results. Remaining fixed bindings repeat those result recipes |
| R07 fixed 2/3 and reserved 6/8/9 | No change: V08's random type 2 and V20's type 3 execute the same selected handlers; type 7 already covers subtract-with-no-add used by the reserved types |
| R08a remaining scale fields | No change: V09 covers enabled positive, nonpositive, disabled length, large-value and P=100 branches. Other fields repeat the table bindings |
| R08c/d other entry profiles | No change: PVP-V08 covers direct/flagged and local entry profiles; V27 covers queued placement. Upstream integration remains explicitly gated |
| R09 event 11 with P=100 | No change: V27 explicitly supplies P=100 and requires defender event 11 on eligible hits |
| R11 demon / null / non-monster variants | No change: V28 covers classified and non-monster bypass results and both loop lengths. Remaining fixed predicates use those same bypass/result recipes |
| R12a wrapping sum / ineligible enqueue | No change: wrapping is the shared NRF arithmetic; V03/V27 already require the missed basic attack to build/enqueue and reach queue completion without scaling/prediction |
| R13a no-match / apply rejection / late mode test | Added V29(a) at the isolated post-event tail: a still-hit player now in mode 0 must retain the queue and request neither reflection, reaction nor removal. Other no-effect/rejection families remain covered by V10/V29 and their predicate integration remains gated |
| R14/R15 negative/in-range physical | No change: V11 case 2 already supplies physical below life and asserts it reaches leech unchanged. Signed min is the same operation for negative input; leech interiors remain LSE-owned |
| R16 state 92 / invalid/dead no-op | No change: V12's direct dead case covers return 0 with no write; the other guards select that same result, and LSE owns state meaning |
| R18 divisor quirk | No change: LSE-R15 owns the term rewrite and its divisor; this stage consumes the resulting current term. V13 discriminates this consumer's raw subtraction/threshold |
| R22 missing class / absent defender | No change: V16 covers no-double/no-draw and marker/overlay branches. Missing records use the stated no-double guard; absent defender skips the same optional marker/overlay requests without adding a new arithmetic recipe |
| R23 callback mutation/draws | No change: IEP-R07/V08 own mutation/traversal; callback-specific interiors retain Q11 |
| R24a Expansion throwable / other weights / qualifications | No change: V17 covers thresholds 4/10, Classic no-draw, cyclic weighted choice and qualification-controlled third draw. Other fixed item/location bindings repeat these recipes |
| R26 early queued returns | Covered by new V29(a) for the late retained-queue return; V10 already covers failed reach/removal. Reaction remains HRC-owned |
| R28a monster cold / freeze fallback | No change: CSA-V09/V10/V11 own these applier interiors; DPO-V19 covers the extra fallback draw placement |
| R29 full integration counts | No change: Q11/Q22/Q25 and other named dependency gates prevent inventing callback/handler/reflection draw counts |

HARDEST outcomes (3), all no rule change: V27 has separate hit/lethal/miss dispatch logs and
explicit supplied interiors; V30 separately asserts stored total versus monster-type prediction;
V24 logs exact stat identities/order and omitted reads. No missing decision was reported there.

V21's shorthand stat assignments were spaced explicitly to match its unchanged numeric trace;
V26's original entry-kind encoding is now named skill kind; a remaining bare binary address in
the reverse-consumer list was replaced by its semantic operation. These are wording/evidence
clarifications. The only new executable case is V29(a). With full pass 3 returning MISMATCH 0
and all findings triaged, the next fresh reader may use the delta rule for that case and the
V10 clarification. No behavioral rule, ownership, evidence label or gate changed. Review
completion remains pending that check.

## Fresh delta blind-implementer pass 4, 2026-10-15

Changed material only: V29(a)'s post-event retained-queue return and V10's explanation
of outputs saying "no events". Reported MISMATCH 0, GUESS 0, GAP 0 and DISCRIM 0.
The reader recomputed V29(a)'s mode-0 return before reflection, reaction and removal:
the queue remains, the three request logs stay empty, and life, result and seed remain
unchanged. It also distinguished V10's apply-dispatch requests from callback execution.
No further contract edit follows this report.

Access limitation: the fresh reader's file tools failed. It checked the exact current
changed cases and their relevant R13a/R23/R26, unit identity, result, seed and
apply-prefix definitions supplied as excerpts by the coordinator. This is a review of
that changed material, not a fourth full-file review or an independent source check.
Full pass 3 covers the unchanged contract. The final delta findings are resolved within
this stated access scope; gameplay, dependency integration and the existing questions
remain unconfirmed. No rule label, ownership or gate is promoted by reader review.

## Independent review spot-check and corrections, 2026-10-15

A separate reviewer re-read original D2Game disassembly (original 1.13c instruction listings, read-only) for sampled rules, deriving each reading before comparing it with the rule text. No gameplay, runtime or table read was performed; no label was promoted.

| Original function | Rules compared | Result |
|---|---|---|
| D2Game 0xDD450 (build) | DPO-R03a, R04a, R05, R06a, R06b, R07, R21a, R21b, R22 placement, R23 build event 3, R31 demon term | Confirmed, except the ratio-128 scaling noted below |
| D2Game 0xDC530 (physical roll) | DPO-R02a, R02b | R02a confirmed (ratio 128 skip; (value/128)*ratio for value above 0x100000); R02b corrected below |
| D2Game 0xDCD80 (element roll) | DPO-R04a, R04b, DPO-V24 read order | Confirmed: maximum, minimum, then mastery read only for shifted maximum at least 8 |
| D2Game 0x770E0 (monster double) | DPO-R22 | Confirmed; byte run `3A D0` decodes as `cmp dl, al`, followed by unsigned `jae` |
| D2Game 0xDDDE0 (queued completion) | DPO-R01a eligibility, R08c, R12a | Confirmed: build only when eligible and not already built; strict signed masked comparison; monster-type lifeDrain term |
| D2Game 0xDC0B0, 0xDA7B0 | DPO-R08a, R09, R10, R11, R12a stored total | Confirmed: inline mulDiv with value-first threshold; strip and bypass order |
| D2Game 0xDE0C0 (apply) | DPO-R13a gates and events, R14, R16 placement, R17, R18, R19a, R19b, R19c, R30 placement | Confirmed |
| D2Game 0xDDBA0, 0xDCC20 | DPO-R24a, R27 | Confirmed: modulus 7 start, weight remainder walk, attacker-seed stun roll below 90 cancels |

Seed arithmetic for S1/S5 steps 1-5 and every draw value quoted by DPO-V02, V04 to V08, V16 to V20, V24 and V27 was recomputed and matches.

**Correction 1 (DPO-R05, R06a, R06b, R21b).** Original discrepancy: these rules used `scale`, which returns the value unchanged at ratio 128. In 0xDD450 only the four elemental writes branch on ratio 128; the monster drain additions, poison magnitude, cold-length, stun-length and burn-length terms always execute `imul` by the ratio byte followed by signed division by 128. At ratio 128 these differ from `scale` when the term's magnitude is at least 2^24. New helper `scaleAlways` and DPO-V22's wrap case record the fixed-code behavior. Existing vector results are unchanged.

**Correction 2 (DPO-R02b, asm-only).** 0xDC530's unarmed branch raises a minimum at or below 0 to 1 and a maximum at or below 1 to 2 before the stat 111 addition and shift; the weapon strength/dexterity factors are sign-extended from 16 bits (`movsx`) before multiplication. The earlier text omitted both. Label unchanged.

**Correction 3 (DPO-R08b, DPO-Q01).** Original discrepancy: DPO-R08b restated the scale selector as asm-only with eligibility open, while sibling PVP-R02a holds the same selector as verified static with its predicates read. DPO-R08b now consumes PVP-R02a, and DPO-Q01 is narrowed to PVP-R02b's scale-condition producer. No P value changed.

**Wording fixes.** DPO-R06b states that the poison-count division applies to the accumulated length (0xDD450 divides the stored sum). DPO-R15 cites LSE-R01/LSE-R03/LSE-R05/LSE-R08/LSE-R09/LSE-R12/LSE-R14 for the leech effects it consumes, and states the draw condition of LSE-R12. Section 3 garbled source cells and the unexplained asterisk were repaired; section 6 no longer cites the removed sub-id R31c.

These corrections change rule behavior (Correction 1) and add a vector case, so a fresh full blind-implementer pass is required for the resulting contract; delta pass 4 above no longer covers it.

## Fresh full blind-implementer pass 5, 2026-10-15

Full run after the spot-check corrections (scaleAlways in DPO-R05/R06a/R06b/R21b, DPO-R02b, DPO-R08b, DPO-R15, DPO-V22 wrap case). The reader recomputed all 31 vectors. Counts: MISMATCH 0, GUESS 8, GAP 7, DISCRIM 24 (25 entries listed), GATED 46, COVERAGE 49, HARDEST 3.

Source checks made by the spec writer for the behavioral edits below (original D2Game 1.13c disassembly, read-only, 2026-10-15): hit result 0xDE5A0 (reach failure jumps to the return-0 exit even with hit pre-set; each avoidance bit is tested and OR-ed independently, then any nonzero value clears hit); build 0xDD450 conversion jump table (type 10 adds 1 to the draw(5) result and dispatches through the same table, so types 1..5 run their handlers; elemental writes multiply then signed-shift with no clamp after scaling); monster double 0x770E0 (overlay request pushes 0, 0x36, defender); durability selection 0xDEBA0 (eligibility is the armor-type test per body location; walk starts at the drawn index, selects when remainder < weight, else subtracts); item durability 0x12EE0 (neither armor type 50 nor weapon type 45, or not qualified, returns before the draw; armor type gives 10 even for a weapon type); apply 0xDE0C0 (failed hostility without override, or absent room, clears result bits 1/2 and returns; flagged events require record flags 32); queued completion 0xDDDE0 (0x8380 test precedes the hit test and the build/scale call; stored total is written only inside the scale/mitigation call); scale/mitigation 0xDC0B0 bypass decision (undead+256, demon+512, flag 1024 only when neither or for a non-monster/null defender).

| Entry | Outcome |
|---|---|
| GUESS R01a reach-failure return | Spec edit: reach failure returns 0 even with hit pre-set (0xDE5A0); V03 running-player reach case added |
| GUESS R01a multi-bit avoidance | Spec edit: each b mask maps independently and is OR-ed; V03 b 0x13 case added |
| GUESS V08 type 10 S5 | Spec edit: R07 states the selected type's handling runs; V08 asserts S5 poisonMagnitude 0x80, poisonLength 50, and fixes type-10 starting fields at 0 |
| GUESS R13a flagged/unflagged | Spec edit: gate 7 names the scale-request argument and states no dispatch on flagged apply without record flags 32 |
| GUESS R02b percent floor/unarmed order | No change: R02b is asm-only and gated by DPO-Q27; its text already orders the floor before the stat 17/18 additions and the unarmed floors before stat 111 |
| GUESS R24a walk start/eligibility/threshold | Spec edit: walk starts at and includes the start index; eligibility is the armor-type test, qualification is tested on the selected item; items neither armor nor weapon type return before the draw; armor type threshold 10 |
| GUESS R22 overlay argument | Spec edit: overlay request on defender with zero extra parameter; V16 output states (defender, 54, 0) |
| GUESS R07 percent above 100 | No change: R07 states converted = mulDiv(percent, physical, 100) with no cap and physical floored at 0; no alternative reading remains |
| GAP section 6 stale overflow/burn wording | Spec edit: overflow bullet names scale/scaleAlways wrapping; R18 bullet names manaDrain/staminaDrain |
| GAP R08c/d profiles only in companion | Spec edit: R08c defines queued, callback, direct-then-flagged, direct-then-unflagged and unscaled local profiles semantically |
| GAP calling-convention leakage in conventions/V02 | Spec edit: replaced with supplied weapon mastery percent and weapon crit chance; V02 says weapon mode 0 |
| GAP V16 defender seed / overlay argument | Spec edit: defender seed S5 asserted unchanged; overlay argument stated in R22 |
| GAP V12 absorb return not in section 5 | Spec edit: section 5 row for the ignored absorb-heal return value |
| GAP R15 omits LSE-R02 return | Spec edit: R15 cites LSE-R02's monster-defender return; Depends on lists the consumed LSE ids |
| GAP V19/R28a ColdEffect/shatter | No change: shatter comparison and ColdEffect values are CSA-owned (CSA-R05/R06, CSA-V09..V11); V19 asserts only the draw count of this placement |
| DISCRIM V01 mask | Vector: case (f) T 2561 |
| DISCRIM V02 source order | Vector: case E (10,50,0) |
| DISCRIM V02 roll equal to chance | Vector: case F (0,0,31) |
| DISCRIM V03 0x8380 vs hit | Vector: V30 case (e) result 0x0105 at queued completion |
| DISCRIM V03 result 4 monster-only | Vector: V03 non-running player case |
| DISCRIM V03 skip avoidance after pre-set | Vector: V03 running player b 2 case |
| DISCRIM V04/V24 mulDiv operands | Vector: V24 case 2 with endpoints 0x20000/0x40000 |
| DISCRIM V05 lightning/cold order | Vector: V05 case 3 |
| DISCRIM V05 clamp/scale order | Vector: V05 case 4 wrap (clamp-after gives 0) |
| DISCRIM V06 mana/stamina order and ids | Vector: V06 case 7 |
| DISCRIM V06 bypass mapping | Vector: V06 cases 9/10 |
| DISCRIM V06 monster bypass flags | Vector: V06 case 8 |
| DISCRIM V07 stat 332 | Vector: V07 case 4 |
| DISCRIM V07/V22 poison length ratio | Vector: V22 ratio case |
| DISCRIM V09 remaining fields | Vector: V09 case 4 |
| DISCRIM V10 override and room | Vector: V10 case 7 |
| DISCRIM V13 exact 256 | Vector: V13 exact-256 case |
| DISCRIM V15/V20 mandatory draw placement | Vector: V15 case 2 |
| DISCRIM V16 lightning/magic/cold | Vector: V16 inputs add those fields |
| DISCRIM V17 thresholds | Vector: V17 (a2) roll 6 and (b2) roll 5 |
| DISCRIM V17 start index scheme | Vector: V17 (b2) start index lands on an eligible location |
| DISCRIM V18 threshold | Vector: V18 rolls 89 and 90 |
| DISCRIM V26 event 11 only when P != 100 | No change: V27 supplies P 100 and asserts defender event 11 on hits |
| DISCRIM V27/V30 scale timing | Vector: V30 case (d) P 17 |
| DISCRIM V28 bypass predicates | Vector: V28 cases 7/8/9 |
| DISCRIM V30 stored-total fields | Vector: V30 case (f) |
| HARDEST R08c/d with V27/V30 | Resolved by the R08c profile definitions and V30 (d) |
| HARDEST R13a with V10/V31 | Resolved by the gate 7 wording |
| HARDEST R24a with V17 | Resolved by the R24a walk/eligibility text and V17 (a2)/(b2) |

COVERAGE entries are advisory; the added cases above cover R01a reach failure and multi-bit b, R04a lightning draw, R05 mana/stamina and other bypass flags, R06a stat 332, R08a remaining fields and R13a absent room. The rest: no change, because each remaining branch repeats a recipe a vector already fixes or belongs to a gated question or a dependency's vectors (as recorded for full pass 3). Rule behavior text changed (R01a, R07, R13a, R22, R24a, R15, R08c definitions), so a fresh full reader run follows.

`check_spec.py combat-damage-pipeline-order`: 0 errors, 0 warnings after these edits. No gameplay observation ran.

## Fresh full blind-implementer pass 6, 2026-10-15

Full run on the contract after pass 5's edits. The reader recomputed every case of all 31 vectors with a script. Counts: MISMATCH 0, GUESS 12, GAP 10, DISCRIM 28 (31 entries listed), GATED 15 rules plus questions and stage-only vectors, COVERAGE 26 groups, HARDEST 3.

Source checks by the spec writer for this pass (original D2Game 1.13c disassembly, read-only, 2026-10-15): build 0xDD450 (physical-supplied flag jumps past the physical roll, its ratio scaling and all crit sources to the elemental stage; the ratio scaling sits inside the physical roll before the crit tests; crit chances test signed nonpositive as skip; monster drains call the element roll with the prior field and add its scaled result to the prior, so a no-maximum prior doubles; player drains add stat 60/62 raw; poison magnitude overwrites with the scaled roll; length terms as stated); queued completion 0xDDDE0 (prediction only ORs result 2, never clears it); apply 0xDE0C0 (both event forms sit under the result-32/hit test; after the final death flag, lethal events 10/9 follow with no death-mode or kill-credit call between); monster double 0x770E0 (marker is an OR of 0x10).

| Entry | Outcome |
|---|---|
| GUESS R05 vs section 6 | Spec edit: section 6 states the drain prior doubles when the maximum is nonpositive; V06 case 14 |
| GUESS R02a/R03a ratio of supplied physical, scaling vs crit | Spec edit: R02a states supplied physical is not ratio-scaled and scaling precedes crit; V04 cases 4-6 |
| GUESS R02b percent arithmetic | No change: R02b is asm-only, gated by DPO-Q27 |
| GUESS R12a prediction clearing | Spec edit: prediction never clears result 2; V30 case (g) |
| GUESS R13a gate 7 both forms | Spec edit: "only with result 32 clear and hit set (for both forms)" |
| GUESS R22 OR vs assign | Spec edit: OR 0x10; V16 hitClass 0x01 variant. The Crit compare is already stated as unsigned byte |
| GUESS R16 invalid/dead predicate | No change: R16 already cites DPO-R13a gate 4's predicate |
| GUESS R01a vs SBA block flag | No change: SBA-R05 maps code 1 to 16 and its block flag to the hit clear that R01a's mask-16 step performs |
| GUESS V24 stat-read log scope | Spec edit: the log covers only the elemental stage's reads |
| GUESS V25 mana/stamina order | Spec edit: resource writes are logged in the same call log, mana first as asserted |
| GUESS V28 undead and demon | No change: R11 states each predicate; no case requires both |
| GUESS R08c/d profile per caller | No change: caller-to-profile integration is DPO-Q19/Q20 |
| GAP mitigation writes not in section 5 | Spec edit: section 5 row naming DMI-owned per-entry writes consumed by DPO-R16 |
| GAP death handling | Spec edit: section 5 states death here is the result flag, with no death-mode or kill-credit write before events 10/9 |
| GAP result mask 512 | Spec edit: section 3 names mask 512 and its outside-path producer (SBA-R06) |
| GAP packed encodings | No change: section 3 already defines the numeric masks as logical controls retained for exact vectors without imposing a layout |
| GAP hit-class/mark/overlay producers | No change: gated by DPO-Q18/Q21 |
| GAP weapon mode input | Spec edit: section 3 build-argument row (weaponMode 0 on the basic attack; other values skip only the weapon crit source) |
| GAP R24a PvP order | Spec edit: attacker step before defender step; both run in player-vs-player |
| GAP reaction/HRC state 92 loop | No change: HRC owns the reaction interior; a later hit reads state 92 through R16's stated gate |
| GAP O01 monster attacker | Companion edit: O01 procedure now uses a player basic attack against a monster defender |
| GAP harness needed | No change: not a gap; every vector was executable |
| DISCRIM V01 prediction OR final | No change: V14 case 2 asserts lethal cleared with prediction set and life 0x100 |
| DISCRIM V01/V12 whole-life rounding | Vector: V01 case (g) T 300 |
| DISCRIM V02 negative chance | Vector: V02 case G |
| DISCRIM V02 third source success | Vector: V02 case H |
| DISCRIM V03 b 4 / b 8 | Vector: V03 cases |
| DISCRIM V03 reach only when pre-set | Vector: V03 monster reach-0 case |
| DISCRIM V03 state 54 on monsters | Vector: V03 monster state-54 case |
| DISCRIM V04 percent floor | No change: R02b asm-only, gated by DPO-Q27 |
| DISCRIM V05/V24 element order | Vector: V05 case 5 |
| DISCRIM V05 elements with scaleAlways | Vector: V05 case 6 |
| DISCRIM V06 drains with scale | Vector: V06 case 11 |
| DISCRIM V06 player term scaled/assigned | Vector: V06 case 12 |
| DISCRIM V06 hireling classes / player flags | Vector: V06 case 13 for flags; hireling class list is GDF-R08's mapping |
| DISCRIM V07 poison ratio | Vector: V07 case 5 |
| DISCRIM V08 selected after subtraction | Spec edit: claim removed (selection order is not observable); Discriminates restated |
| DISCRIM V08 type-10 seed owner | Vector: defender seed S5 unchanged asserted |
| DISCRIM V08 set vs raise length | Vector: prior-70 cases for types 4/11/12 |
| DISCRIM V08 conversion operand order | Vector: large-percent case |
| DISCRIM V10 killable and modes | Vector: cases 9-11 |
| DISCRIM V10 room only after hostility failure | Vector: case 8 |
| DISCRIM V10 inTown monster | Vector: case 12 |
| DISCRIM V13 stamina below 256 | Vector: stamina 0x800 / drain 0x740 |
| DISCRIM V16 other fields doubled | Vector: manaDrain, staminaDrain, burnMagnitude, freezeLength inputs asserted unchanged |
| DISCRIM V16 nonzero hitClass overlay | Vector: hitClass 0x01 variant |
| DISCRIM V17 durability on miss | Vector: case (a3) |
| DISCRIM V21 freeze-only condition | Vector: (40,0) case |
| DISCRIM V21 two-stat precedence | Vector: (40,20) with stats 153 and 118 |
| DISCRIM V22 prior poison length | Vector: prior-length case |
| DISCRIM V27 build event 3 placement | Vector: V27 case (d) callback logs physical 351 then 176 |
| DISCRIM V28 end vs skip | No change: equivalent; entries 9-11 are the last entries and all monster-only, so ending and skipping give identical results |
| DISCRIM V30 head vs tail | Vector: V30 case (h) |
| HARDEST R05 | Resolved by the section 6 correction and V06 case 14 |
| HARDEST R13a/R08c/R08d | No further change beyond the gate 7 wording; profiles defined in pass 5 |
| HARDEST R24a/V17 | No change: reader executed the walk correctly |

COVERAGE entries are advisory; the added cases cover R01a b 4/8, R02a ratio scaling, R03a negative chance, R04a magic, R05 prior/flags, R06 ratio and prior length, R10 cold-only, R13a killable/town/mode rejections and R18 stamina. The remaining branches repeat recipes already fixed by vectors or sit behind named gates or dependency vectors. Rule text changed (R02a, R12a, R13a, R22, R24a, sections 3, 5, 6), so a fresh full reader run follows.

`check_spec.py combat-damage-pipeline-order`: 0 errors, 0 warnings. No gameplay observation ran.

## Fresh full blind-implementer pass 7, 2026-10-15

Full run on the contract after pass 6's edits. Counts: MISMATCH 1, GUESS 6, GAP 7, DISCRIM 13, GATED 15 rules plus questions and stage-only vectors, COVERAGE 24 groups, HARDEST 3.

Source checks by the spec writer (original D2Game 1.13c disassembly, read-only, 2026-10-15): physical roll 0xDC530 (the -90 floor applies to the shared percent before stats 18/17 are added; endpoint sums are not floored); build 0xDD450 (type-10 draw depends only on a nonzero conversion type; build event 3 follows the burn update and precedes conversion); scale/mitigation 0xDC0B0 (stored total written at the end of the stage); item durability 0x12EE0 (armor type selects threshold 10 before any throwable test); durability selection 0xDEBA0 (attacker step precedes defender step); apply 0xDE0C0 (last-attacker write precedes the requested scale run).

| Entry | Outcome |
|---|---|
| MISMATCH V05 case 5 cold | Spec correction: the spec writer's pass-6 trace omitted the minimum; cold is 0x100 + 452 = 0x2C4 |
| GUESS R02b floor placement | Spec edit (label unchanged, asm-only): floor applies to the shared percent before stats 18/17; endpoint sums not floored |
| GUESS R10/R11 order | Spec edit: R10 states it runs after event 11 and before the mitigation loop |
| GUESS R12a stored total eligibility | Spec edit: total stored only at the end of a scale/mitigation run |
| GUESS R24a armor and throwable | Spec edit: armor type takes threshold 10 without a throwable test |
| GUESS R13a first mode test | No change: R13a's first test concerns the attacker's mode 0 only; defender modes are gate 4 |
| GUESS R07 type 10 with percent 0 | Spec edit: draw occurs also when conversionPercent is 0 |
| GAP V18 "bit 3" | Spec edit: special-stun condition false |
| GAP V20 "stun flag" | Spec edit: special-stun condition false |
| GAP V17 Expansion flag | Spec edit: conventions state Expansion unless a case says Classic |
| GAP V17 DUR-R05 arithmetic | No change: loss arithmetic is DUR-owned; V17 asserts the logged request and draws |
| GAP V19 CSA inputs | No change: CSA-R05/R06/R07 own cold effect and eligibility |
| GAP harness logs not in section 5 | Spec edit: conventions state call and stat-read logs are harness observations, not game outputs |
| GAP V12(d) | Spec edit: (d) has explicit absorb 0x400 and total 0 with final life 0xC00 |
| DISCRIM V04 scaling operand order | Vector: V04 case 7 (65583 vs reversed 65536) |
| DISCRIM V09 case 1 operands | No change: case 2 rejects reversed operands, as the reader confirmed |
| DISCRIM V05/V24 mastery ids 330/357 | Vector: V24 case 3 |
| DISCRIM V20/V15 drains before elements | Vector: V15 case 3 |
| DISCRIM V22 override assigns | Vector: override-with-prior case (95) |
| DISCRIM V26 P=100 | Vector: V26 case 2 (P 100 still dispatches) |
| DISCRIM V27(d) event 3 before burn | Vector: V27(d) callback also logs burnMagnitude 2 |
| DISCRIM V10 gate 5/6 order | Vector: V10 case 4 scale call observes last attacker |
| DISCRIM V17 hit re-tested after apply | Vector: V17 (a4) |
| DISCRIM V17 PvP order | Vector: V17 (c) request log order |
| DISCRIM V03/V30 0x8380 bits | Vector: V30 (e2) results 0x0081, 0x0201, 0x8001 |
| DISCRIM V16 whole-word hitClass | Vector: V16 hitClass 0x100 variant |
| DISCRIM V16 overlay without doubling | Vector: V16 Crit 51 with hitClass 0x20 |
| HARDEST V03/R01a, R13a/V10, R24a/V17 | No change: each was implemented and recomputed correctly; the remaining difficulty is the number of ordered conditions, all now stated and exercised |

COVERAGE entries are advisory; the remaining branches repeat recipes already fixed by vectors or are behind named gates or dependency vectors. The run-cap reader run follows as a full run because rule text changed.

`check_spec.py combat-damage-pipeline-order`: 0 errors, 0 warnings. No gameplay observation ran.

## Fresh full blind-implementer pass 8, 2026-10-15 (run cap reached)

Full run on the contract after pass 7's edits; this is the fourth and last reader run allowed for this review round. Counts: MISMATCH 0, GUESS 9, GAP 6, DISCRIM 17, GATED 15 rules plus 24 questions, 3 observations and stage-only vectors, COVERAGE 20, HARDEST 3. The pass is **not complete**: GUESS, GAP and DISCRIM entries below remain unresolved because the run cap was reached. No contract edit was made after this run, so pass 8 reviewed the current contract text. Proposed outcomes are recorded for the next review round; each needs a fresh reader run once applied.

Source checks made while triaging (original D2Game 1.13c disassembly, read-only, 2026-10-15): hit result 0xDE5A0 (a zero to-hit return leaves hit clear and skips the avoidance call; result 0 follows); build 0xDD450 (stats 103/104/106 set flags 256/512/1024 by three independent tests; the nonzero poisonMagnitude test reads the value after scaleAlways is stored).

| Entry | Status | Proposed outcome |
|---|---|---|
| GUESS R01a to-hit return | open | Spec edit: nonzero to-hit return sets hit; zero leaves hit clear and skips avoidance (asm checked); V03 case to-hit 0 |
| GUESS R02b factor rounding | no change | R02b asm-only, gated by DPO-Q27 |
| GUESS R06b poisonMagnitude test | open | Spec edit: test uses the scaled stored magnitude (asm checked); vector at ratio 64 with roll 1 |
| GUESS R05 bypass flags | open | Spec edit: three independent tests (asm checked); V06 case stats 103 and 104 giving 0x300 |
| GUESS R11 undead definition | open | Spec edit: cite DMI-R04 (lUndead or hUndead); V28 hUndead-only case |
| GUESS R08a drain scaling for non-monsters | no change | R08a scales every positive scale-enabled value with no attacker-type gate; V09 shows it |
| GUESS V27 endpoint supply | no change | V27 is a supplied-endpoint stage per DPO-R02a; endpoint production is DPO-Q27 |
| GUESS R23 lethal-event condition | no change | R23 states result 2 from DPO-R19b with result 32 clear and no further test |
| GUESS R12a/R16 life and max-life reads | open | Spec edit: name the consumed max-life operation contract (NRF or GDF owner) for stat 7 |
| GAP max-life read undefined | open | As above |
| GAP harness logs | no change | Section 8 conventions declare them harness observations |
| GAP hitClass logical meaning | no change | Marker meaning is DPO-Q18/Q21; masks are retained logical controls (section 3) |
| GAP physical endpoint source in V27/V15 | no change | V27 supplies endpoints; V15 is physical-supplied; production is DPO-Q27 |
| GAP duplicate queue matches | open | Spec edit: section 3/5 queue-entry semantics for several matching entries; vector with two matches |
| GAP queued record field list | open | Spec edit: section 5 states the queued entry copies the whole record |
| DISCRIM V03 to-hit return ignored | open | Vector: to-hit 0 case |
| DISCRIM V06 monster skip flags 8/16 | open | Vector: flags 8 and 16 with mana/stamina ranges |
| DISCRIM V06 else-if bypass chain | open | Vector: stats 103 and 104 giving 0x300 |
| DISCRIM V08 type 10 with percent 0 | open | Vector: type 10, percent 0, S1 |
| DISCRIM V16 signed Crit compare | open | Vector: Crit 200 with roll 51 |
| DISCRIM V17 Expansion throwable | open | Vector: Expansion throwable with roll 4..9 |
| DISCRIM V17 empty defender inventory | open | Vector: player defender with W 0 |
| DISCRIM V12 state 92 | open | Vector: state 92 with positive absorbHeal |
| DISCRIM V13 negative drains | open | Vector: manaDrain -0x100 |
| DISCRIM nonzero final dispatcher results | open | Vector: callbacks returning nonzero with later stages unchanged |
| DISCRIM V28 hUndead | open | Vector: hUndead-only defender with flag 256 |
| DISCRIM V29/V10 several matching entries | open | Vector: two matches, head copied, both removed |
| DISCRIM V29 first attacker mode-0 return | open | Vector: hit with attacker mode 0 at queue entry |
| DISCRIM V10 town vs dead gate order | open | Vector: town room, defender mode 12, R 3 |
| DISCRIM V15/V23 ratio 0 normalization | open | Vector: ratio 0 cases |
| DISCRIM V23 wrapped burn product | open | Vector: burnMagnitude 0x1000000 at ratio 128 |
| DISCRIM V19 monster E != 0 seed owner | open | Vector: monster defender with nonzero E (CSA-owned interior; placement draw only) |
| HARDEST R24a/V17, R13a/R08c/V10, R12a/V30 | no change | Reader implemented and recomputed each with MISMATCH 0 |

Remaining to finish the blind-implementer pass: apply the open proposed outcomes above (verified against the original instructions first), then a fresh full reader run, repeating until MISMATCH 0 and every GUESS, GAP and DISCRIM is resolved or gated. Review completion is not claimed.

`check_spec.py combat-damage-pipeline-order`: 0 errors, 0 warnings. No gameplay observation ran.

## Pass-8 proposed outcomes applied, 2026-10-15

Every pass-8 entry marked open above was applied as proposed, after a fresh read-only check of the original D2Game 1.13c disassembly: hit result 0xDE5A0 (zero to-hit return leaves hit clear and skips avoidance); build 0xDD450 (ratio 0 becomes 128 at entry; build marker 32 ORed at entry; monster skip flags 4/8/16; three independent bypass tests; poison length test reads the stored scaled magnitude; burn product wraps before the `max(0, x)` test); monster double 0x770E0 (unsigned byte compare); apply 0xDE0C0 (gate order hostility/room, town, killable, invalid/dead; no dispatcher return is tested); absorb heal 0xDB570 (state 92 refusal; cap read through the life-cap getter LSE-R10 describes); burns 0xDA970/0xDA930 (nonpositive drains return before any write); durability 0xDEBA0/0x12EE0 (W 0 takes no draw; Expansion throwable threshold 10); cold applier 0xDC780 (the E != 0 draw steps the defender); queue consumer 0xDED30 (first match from the head copied whole as 0x1C dwords; attacker mode-0 return before apply retains the queue) and removal 0xDA6D0 (every matching entry unlinked).

Spec edits: DPO-R01a (to-hit return), R05 (independent bypass tests), R06b (scaled magnitude test), R11 (DMI-R04 class tests), R13a (head-most match, whole-record copy, removal of every match), R16 (LSE-R09/LSE-R10/LSE-R11 cap read), section 3 queue-entry and stat 7 rows, section 5 queue row, Depends on (DMI-R04, LSE-R10). Vectors: V03 to-hit 0; V06 cases 15-18; V07 case 6; V08 type 10 at percent 0; V10 case 13; V12 (e); V13 negative drains; V16 Crit 200; V17 (a5), (b3); V19 monster E -20; V23 ratio-0 and wrap cases; V27 (e); V28 case 10; V29 (b), (c). Rule text changed, so the next reader run is full.

## Fresh full blind-implementer pass 9, 2026-10-15

Full run (run 1 of at most 4 in this review task) on the contract after the pass-8 outcomes. Counts: MISMATCH 0, GUESS 7, GAP 8, DISCRIM 11, GATED 15 rules plus 24 questions, 3 observations and stage-only vectors, COVERAGE 32, HARDEST 3. Every pass-8 finding was confirmed resolved; the reader recomputed every new case.

Source checks for this triage (original D2Game 1.13c disassembly, read-only, 2026-10-15): 0xDD450 (type-5 /8 is a truncating signed divide; the stat-56 cold-length test reads cold after the element rolls and before conversion; flag 32 ORed at entry; base priors of physical, elements and poison are the roll bases; weapon mode nonzero skips source a); 0x770E0 (class id tested signed against the MonStats count; with no defender the doubling stays and the marker and overlay are both skipped); 0xDE0C0 (lethal 10/9 test result 32 and result 2 only; a monster attacker whose MonStats row is missing fails the town gate; 5/1 and 6/2 require hit); 0xDE5A0 (hostility precedes reach and the hit operations); 0xDA930 (stamina remainder exactly 0x100 survives).

| Entry | Outcome |
|---|---|
| GUESS R02b truncation and unarmed order | No change: R02b is asm-only, gated by DPO-Q27 |
| GUESS R07 type-5 rounding, negative percent | Spec edit: truncated toward zero; negative percent already follows `max(0, physical - converted)` |
| GUESS R22 marker without defender | Spec edit: doubling stays, neither marker nor overlay; V16 defender-absent variant |
| GUESS R23 lethal events on the flagged form | Spec edit: both forms; record flags not tested |
| GUESS R06b positive cold timing | Spec edit: cold after DPO-R04a, before conversion and monster double |
| GUESS V08 100-percent type | Spec edit: type 7 |
| GUESS R05/R22 MonStats row | Spec edit: drains read no row; valid row is class in [0, count) |
| GAP build marker 32 producer | Spec edit: DPO-R02a states the entry OR on hit and miss |
| GAP DPO-R31 inputs in defaults | Spec edit: conventions state class flags clear, no damage_vs_montype, weapon not Blunt |
| GAP companion armorclass record | Spec edit: asterisk removed; the name consumes GDF-R02 |
| GAP gates 1/2 null units, missing attacker row | Spec edit for the missing MonStats row (fails town gate); null-unit hostility/room behavior stays DPO-Q30 |
| GAP V17 unqualified and monster/monster inputs | Vector: explicit seeds and outputs |
| GAP DPO-R19c update in V01/V27 | Vector: V01 states the update is a supplied no-effect operation; V27 already does |
| GAP identity mitigation | No change: vectors supply zero reductions/resistances under DMI's contract |
| GAP V10 queued reach prose | Vector: case 14 with outputs |
| DISCRIM V03 progressive-to-hit argument | Vector: stat 0x145 = 5 call (attacker, defender, 5, 0) |
| DISCRIM V03 hostility last | Vector: hostility-0 case asserts no reach, to-hit or avoidance call |
| DISCRIM prior bases | Vector: V04 case 8, V05 case 7, V07 case 7 |
| DISCRIM V02 weapon mode | Vector: V02 case I |
| DISCRIM hireling classes | Vector: V06 case 19 |
| DISCRIM V10 hit-clear dispatch | Vector: case 15 |
| DISCRIM V13 stamina exact 256 | Vector: drain 0x700 |
| DISCRIM V16 mask 0xF0 | Vector: hitClass 0x10 and 0x80 |
| DISCRIM V17 weights/eligibility | Vector: (b4) W 10 with a location-4 weapon |
| DISCRIM V27 lethal placement | Vector: lethal call log includes the life write and appliers |
| DISCRIM V29 reflection on miss, V30 miss enqueue | Vector: V29 thorns 50 with no reflection request; V30 (i) |
| HARDEST R13a, R05, R24a | No change: each recomputed without mismatch |

COVERAGE entries are advisory; the added cases cover the branches above where a result changes. Rule text changed (R02a, R06b, R07, R13a, R22, R23), so the next run is full.

`check_spec.py combat-damage-pipeline-order`: 0 errors, 0 warnings. No gameplay observation ran.

## Fresh full blind-implementer pass 10, 2026-10-15

Full run (run 2 of at most 4 in this review task) on the contract after pass 9's edits. Counts: MISMATCH 0, GUESS 4, GAP 6, DISCRIM 13, GATED 15 rules plus 24 questions, 3 observations and stage-only vectors, COVERAGE 24, HARDEST 3. Every pass-9 change was recomputed without mismatch.

Source checks for this triage (original D2Game 1.13c disassembly, read-only, 2026-10-15): 0xDD450 (the physical-supplied flag jumps only past the roll, ratio scaling and crit; the stun term multiplies by the build ratio; conversion floors physical at 0 after subtracting the unfloored converted amount; cold-length stage precedes conversion; build event 3 requires a player attacker or a hireling); 0xDE0C0 (final-life test is `life <= 0`; hostility/room failure clears results 1 and 2; town gate precedes killable; flagged events need result 32 clear and hit); 0xDED30/0xDA6D0 (reach failure removes every match; reflection then reaction on the surviving-hit tail); 0xDEBA0 (weighted walk, as V17 b4/b5 trace).

| Entry | Outcome |
|---|---|
| GUESS RLI-R02b adjustment on stats 6/7/8/10 | Spec edit: vector conventions supply definitions for which RLI-R02b returns stored values |
| GUESS R28a monster with state 54 | Spec edit: R28a defers to CSA-R07's routing, naming the state-54 return |
| GUESS R02b unarmed order | No change: R02b is asm-only, gated by DPO-Q27 |
| GUESS physical-supplied scope | Spec edit: DPO-R02a states every other build stage still runs |
| GAP R02b vectors | No change: asm-only, DPO-Q27 |
| GAP hostility/reach/avoidance arguments | No change: PVP-R05a and SBA own those operations; reach arguments are DPO-R01b/DPO-Q23 |
| GAP R08c/R08d direct profiles | No change: caller profiles are gated by DPO-Q19/Q20; enqueue rejection is advisory coverage |
| GAP reaction on a surviving hit | Vector: V29 (d) reflection then reaction then removal |
| GAP Id-map traceability | Spec edit: V20 under R03a, V31 under R08c, V30 under R08d, V29 under R13a |
| GAP harness logs | No change: declared harness observations |
| DISCRIM V06 monster drain ratio | Vector: cases 20/21 at ratio 64 |
| DISCRIM V22 stun ratio | Vector: ratio 64 stun 20 |
| DISCRIM V08 physical floor | Vector: 150-percent case |
| DISCRIM V14 lethal test | Vector: case 3 life 0x80 |
| DISCRIM V10 reach-failure removal | Vector: case 14 with two matches and one non-match |
| DISCRIM V10 gate order | Vector: cases 16 and 17 |
| DISCRIM V10 result bits on hostility/room failure | Vector: case 18 with R=3 |
| DISCRIM V31 result 32 and hit | Vector: cases (c) and (d) |
| DISCRIM nonzero dispatcher results | Vector: V26 case 3; V27 (e)/(e2) with conversion, reaction and removal logged |
| DISCRIM V27 build event 3 attacker type | Vector: V27 (f) monster, (g) hireling |
| DISCRIM V22 length stage vs conversion | Vector: V22 build-order case |
| DISCRIM V17 weights | Vector: (b4) reseeded to r=2, new (b5) location 4 |
| DISCRIM V07 zero maximum with prior | Vector: case 8 |
| HARDEST R24a, R13a, R04-R06/R21 | No change: each recomputed without mismatch |

Rule text changed (R02a, R28a), so the next run is full.

`check_spec.py combat-damage-pipeline-order`: 0 errors, 0 warnings. No gameplay observation ran.

## Fresh full blind-implementer pass 11, 2026-10-15

Full run (run 3 of at most 4 in this review task) on the contract after pass 10's edits. Counts: MISMATCH 0, GUESS 4, GAP 6, DISCRIM 5, GATED 15 rules plus 25 question entries, 3 observations and stage-only vectors, COVERAGE 23, HARDEST 3.

Source checks for this triage (original D2Game 1.13c disassembly, read-only, 2026-10-15): 0xDED30 (the flags-32 replacement is inside the hit branch; a miss passes the copied flags to 7/3 and the reaction); 0xDE0C0 (clamp, leech, absorb, subtraction, burns, appliers and the death flag follow gate 7 without a hit test; only the modifier hook tests hit); 0x770E0 (tests attacker type 1, class bounds and Crit only, so hirelings double); 0xDD450 (stun, poison, burn-length and drain terms multiply by the ratio at 128); 0x12EE0 (signed `roll < threshold`).

| Entry | Outcome |
|---|---|
| GUESS R13a flags replacement on a miss | Spec edit: replacement only on hit; V29 asserts flags 34 on the miss record |
| GUESS stages after gate 7 without hit | Spec edit: DPO-R14 states R14..R19 run regardless of hit; V11 case 3 |
| GUESS R02b unarmed order | No change: asm-only, DPO-Q27 |
| GUESS R22 hirelings | Spec edit: any monster-type attacker including hirelings; section 6 hireling clause |
| GAP V29(a)/V12 seams | No change: stated isolated stage boundaries |
| GAP packed masks | No change: section 3 retains them as logical controls for exact vectors |
| GAP return contracts to callers | No change: caller use is DPO-Q19/Q20 |
| GAP block/avoid arguments | No change: SBA-R05 owns them; reach arguments DPO-R01b/Q23 |
| GAP registration values | No change: IEP-R07/IEP-V08 own pass-through |
| GAP monster data/Crit producer | No change: MonStats inputs in section 3 consume GDF |
| DISCRIM V17 threshold equality | Vector: (a6) roll 4, (a7) roll 10 |
| DISCRIM V17 weights 8/9/10 | Vector: (b6) three seeds selecting locations 8, 9, 10 |
| DISCRIM V04 ratio 128 skip | Vector: case 9 physical 2097247 |
| DISCRIM scale vs scaleAlways | Vector: V06 case 22, V07 case 9, V22 stun wrap, V23 length wrap |
| DISCRIM shared record | Vector: V27 (h) callback write of physical 200 reaches the apply |
| HARDEST R13a, R24a, attacker partitioning | No change: recomputed without mismatch |

Rule text changed (R13a, R14, R22, section 6), so the next run is full.

`check_spec.py combat-damage-pipeline-order`: 0 errors, 0 warnings. No gameplay observation ran.

## Fresh full blind-implementer pass 12, 2026-10-15 (task run cap reached)

Full run (run 4 of 4 in this review task) on the contract after pass 11's edits. Counts: MISMATCH 0, GUESS 5, GAP 4, DISCRIM 11 (V17 entry has three sub-alternatives), GATED 15 rules plus 24 questions, 3 observations and stage-only vectors, COVERAGE 36, HARDEST 3. Every pass-11 change was recomputed without mismatch. The pass is **not complete**: the entries below remain unresolved because the run cap was reached. No contract edit was made after this run. Proposed outcomes are recorded for the next round; each needs a fresh full reader run once applied.

Source checks made while triaging (original D2Game 1.13c disassembly, read-only, 2026-10-15): 0xDE5A0 (the Run pre-set tests defender type 0 and mode 3); 0xDA970 (mana remainder compared signed, so a negative remainder stores 0); 0xDD450 (poison override requires stat 101 > 0); 0x12EE0 (qualification is tested before the draw; armor type takes threshold 10 before any throwable test).

| Entry | Status | Proposed outcome |
|---|---|---|
| GUESS V27/V30 physical endpoints 256/512 | open | Spec edit: state them as post-shift endpoint values supplied to DPO-R02a (DPO-R02b production gated by DPO-Q27) |
| GUESS V25 burn calls in the call log | open | Spec edit: the call log holds the stat writes and the five applier calls only |
| GUESS V11 case 3 scale argument | open | Spec edit: scale request false, hostility 1 |
| GUESS R10 state owner | open | Spec edit: states 133/131 tested on the defender |
| GUESS V17 (a)/(c) item inputs | open | Spec edit: attacker weapon not armor type, not throwable, qualified; (c) defender item qualified |
| GAP V27/V30/V02 endpoint inputs | open | As the first GUESS |
| GAP V20 no-crit outputs | open | Vector: state physical, cold and coldLength of the no-crit variant |
| GAP V27 (f)/(g) life and total | open | Vector: state life 2209, lethal clear, total 351 |
| GAP V16 class id | open | Vector: name the attacker class; add hireling case (see DISCRIM) |
| DISCRIM V04 mulDiv at ratio 64 | open | Vector: 21=0x2000, 22=0x2001 at ratio 64 gives 1048576 |
| DISCRIM V03 monster in mode 3 | open | Vector: monster defender mode 3, to-hit 0, result 0 with one to-hit call (asm checked) |
| DISCRIM V10 dead modes per type | open | Vector: monster mode 17 and player mode 12 proceed |
| DISCRIM V13 mana negative remainder | open | Vector: mana 0x100, drain 0x200 gives 0 (asm checked) |
| DISCRIM V22 negative stat 101 | open | Vector: stat 101 -10 gives poisonLength 16 (asm checked) |
| DISCRIM V16 hireling double | open | Vector: class 271, Crit 60 doubles |
| DISCRIM V23 ratio-0 normalization scope | open | Vector: ratio 0 with an elemental, drain or poison roll |
| DISCRIM V26/V27 event-11 record copy | open | Vector: event-11 callback writes coldLength or physical, later stages observe it |
| DISCRIM V27 durability placement | open | Vector: qualifying weapon with an event-7 callback drawing on the attacker seed |
| DISCRIM V31 nonzero event-6 result | open | Vector: event-6 callback returning 1; event 2 still dispatched |
| DISCRIM V17 attacker-step order/threshold/Classic armor | open | Vector: unqualified weapon (no draw), armor-type item (threshold 10), Classic armor-type throwable (draw) (asm checked) |
| HARDEST R24a/V17, R13a/V10, R02a/R02b | no change | Recomputed without mismatch |

Remaining to finish the blind-implementer pass: apply the open proposed outcomes above (behavioral ones verified against the original instructions first), then a fresh full reader run, repeating until MISMATCH 0 and every GUESS, GAP and DISCRIM is resolved or gated. Review completion is not claimed.

`check_spec.py combat-damage-pipeline-order`: 0 errors, 0 warnings. No gameplay observation ran.
