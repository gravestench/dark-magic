# Evidence: damage pipeline order

Spec: [Damage pipeline order](../../specs/combat-damage-pipeline-order.md)

## Identity and reuse

Evidence build: original Lord of Destruction Expansion 1.13c, Windows x86. Artifact ids `LOD113C-D2GAME`, `LOD113C-D2COMMON` and `LOD113C-FOG` refer to the [source manifest](../source-manifest.md). Prior readings retain their 2026-10-12/13 dates, scope and limitations. Newly fingerprinted analysis files identify the available original specimens; historical loaded DLL fingerprints, effective-file hashes, physical archive origins and complete override inventory were not captured. No identity check confirms gameplay.

Original spec/evidence/gap files were archived privately and compared byte-for-byte before edits on 2026-10-14. No plan outside gaps, code or test cited the spec filename or DPO prefix in the consumption search. Current ids are retained without aliases. The single ordered-path boundary remains intact. The original full investigation and binary fixture encodings remain privately preserved; source references below identify the original modules independently.

## Semantic correspondence

Damage fields map to original byte offsets: flags 00, result 04(u16), physical 08, physicalPercent0C, fire 10, burnMagnitude 14, burnLength 18, lightning1C, magic 20, cold 24, poisonMagnitude 28, poisonLength2C, coldLength 30, freezeLength 34, lifeDrain 38, manaDrain3C, staminaDrain 40, stunLength 44, absorbHeal 48, total4C, flatReductionScale 54, hitClass 60, markCondition 64(u8), conversionType 65(u8), conversionPercent 68, overlayId6C. Record size is0x70. These encodings are investigation details, not a required layout.

Queue nodes are0x88 bytes: game 00, attacker type/id 04/08, defender type/id 0C/10, damage 14, next 84; unit queue headAC. Unit type/class/id/mode 00/04/0C/10, monster data 14, seed 20/24, path2C, stat list 5C, inventory 60, game 80, event list 90, last attacker +0x9C/+0xA0, invalid-condition byte +0xC6 mask 1, last-attacker-presenceC8 mask 0x20000, pending eventsDC. Path target 58, room1C; game difficulty6D, Expansion 70, frame +0xA8. Monster-data 16 mask 8 is the neutral special-stun condition; its producer remains DPO-Q28. Public vectors use semantic values, while the privately archived originals preserve raw encodings.

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
| `LOD113C-D2COMMON` | 0x33CD0, 0x37DB0 | 10494 (thunk 0xA68E) | (state test and its bit test; GDF-R11) (analyst label) | every `HasState` | disassembly reading; recorded 2026-10-13; `D2Common.dll RVA 0x33CD0`, `RVA 0x37DB0` |
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

- **DPO-O01** **lethal boundary (DPO-R12a, DPO-R17).** Measure: the defender's stat 6 (`hitpoints`, raw) immediately before and after one hit, and the `total` of the applied record. Expected: `after = before - total` when that is `>= 0x100`, else 0, with no dependence on the predicted-lethal bit. Procedure: a level-1 character of any class in Act I (Blood Moor); let one Fallen hit the character while it stands still (the character is the defender; the Fallen's attack is a basic melee attack); read the character's stat 6 on the server unit the frame before the hit and the frame after, and the record's `total` at the entry of the apply (resource/status apply, supplied apply record); repeat until a hit lands with `before - total` in each of the ranges `>= 0x100`, `1..0xFF` (death) and `<= 0` (death). The build-side pre-flag is read as bit 0x2 of `result` at the same entry.
- **DPO-O02** **crit draws and single doubling (DPO-R03a).** Measure: the attacker's seed `seed` before and after the build (damage builder entry and return) and `physical` with result bit 0x2000 at return. Expected: the number of seed steps equals 1 (physical) + one per crit source with chance `> 0` up to and including the first success + element draws + 1; `physical` is doubled at most once. Procedure: an Amazon with `passive_critical_strike` (stat 0x151) from Critical Strike at a known chance and a weapon with no other crit source, attacking Fallen with a basic attack in the Blood Moor; repeat for configurations (0x151 only), (0x8D only, from an item with `item_deadlystrike`), (both); count steps per hit.
- **DPO-O03** **player mana leech burns the defender (DPO-R18).** Measure: the monster defender's stat 8 (`mana`) before and after one hit by a player with mana leech `M` percent and no `manaDrain` roll (player attackers never roll it). Expected in Normal (`ManaStealDivisor = 1`, GDF-R02): a decrease of `(M << 6)` raw units when the leech ran, for example `M = 5` → 320 raw (1.25 mana), set to 0 if that leaves less than 0x100. Procedure: any class wearing one item with `manadrainmindam` (stat 0x3E) = 5 and no other leech, hitting a monster that has mana (a Fallen Shaman); read the Shaman's stat 8 on the server unit before and after.


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

Final dependency reconciliation removes dependency references absent from the current numeric-foundation contract and uses current NRF-R01/R02/R05/R06a/R08a/R12a/R12b contracts. Current DPO parent ids replace historical subpart references for R31/R32. No original arithmetic or retained expected vector value changes.

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
| DPO-R08a negative/zero scale | no change | V09 already asserts negative cold unchanged and has zero default fields;PVP-V12 also records zero/negative supplied-value results. |
| DPO-R08c/d remaining entry profiles | no change | PVP-V09 records unflagged unchanged, flagged once, unscaled local profiles and direct-then-flagged twice. V27 covers the queued profile;the direct-then-unflagged route uses the already covered one-run shared stage plus unflagged preservation. Upstream caller integration remains gated. |
| DPO-R09 callback mutations | no change | Mutation/traversal belongs to IEP-R07 and IEP-V08;DPO-V26 asserts this caller's position and shared-record arguments. Callback-specific execution/draws remain DPO-Q11. |
| DPO-R12a regular-monster/hireling totals | spec edit | New V30 separates stored total from prediction for regular monster, hireling and player, with exact totals, result flags and queued record. |
| DPO-R13a no match/other apply gates | no change | V10 covers accepted/rejected prefix, result-bit clearing, town rejection, dead return and failed reach;V29 covers matched queue consumption. Remaining predicates select documented no-effect/return families already exercised;predicate acquisition is gated by Q23/Q30. |
| DPO-R16 negative absorb/state/dead | no change | V12 covers the no-write/return 0 family through a direct dead case and positive/capped healing. Other rejection predicates lead to that same result;state meaning consumes LSE-R12. |
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
| DPO-R16 nonpositive absorb/state 92 | no change | V12's direct dead case asserts return 0 with no write, the same result as these guards; positive and capped healing results are also exercised. State 92's meaning belongs to LSE-R12. |
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

ItemTypes has no explicit ID column in this record. These zero-based file data-row indices exclude the header; the existing loader/name convention is consumed separately from GDF-R16.

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
| R08c/d other entry profiles | No change: PVP-V09 covers direct/flagged and local entry profiles; V27 covers queued placement. Upstream integration remains explicitly gated |
| R09 event 11 with P=100 | No change: V27 explicitly supplies P=100 and requires defender event 11 on eligible hits |
| R11 demon / null / non-monster variants | No change: V28 covers classified and non-monster bypass results and both loop lengths. Remaining fixed predicates use those same bypass/result recipes |
| R12a wrapping sum / ineligible enqueue | No change: wrapping is the shared NRF arithmetic; V03/V27 already require the missed basic attack to build/enqueue and reach queue completion without scaling/prediction |
| R13a no-match / apply rejection / late mode test | Added V29(a) at the isolated post-event tail: a still-hit player now in mode 0 must retain the queue and request neither reflection, reaction nor removal. Other no-effect/rejection families remain covered by V10/V29 and their predicate integration remains gated |
| R14/R15 negative/in-range physical | No change: V11 case 2 already supplies physical below life and asserts it reaches leech unchanged. Signed min is the same operation for negative input; leech interiors remain LSE-owned |
| R16 state 92 / invalid/dead no-op | No change: V12's direct dead case covers return 0 with no write; the other guards select that same result, and LSE owns state meaning |
| R18 divisor quirk | No change: LSE-R17 owns the term rewrite and its divisor; this stage consumes the resulting current term. V13 discriminates this consumer's raw subtraction/threshold |
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

**Wording fixes.** DPO-R06b states that the poison-count division applies to the accumulated length (0xDD450 divides the stored sum). DPO-R15 cites LSE-R02/R04/R06/R09/R10/R13/R15 for the leech effects it consumes, and states the draw condition of LSE-R13. Section 3 garbled source cells and the unexplained asterisk were repaired; section 6 no longer cites the removed sub-id R31c.

These corrections change rule behavior (Correction 1) and add a vector case, so a fresh full blind-implementer pass is required for the resulting contract; delta pass 4 above no longer covers it.
