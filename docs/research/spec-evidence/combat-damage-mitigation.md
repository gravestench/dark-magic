# Damage mitigation evidence

Behavioral contract: [Damage mitigation](../../specs/combat-damage-mitigation.md).

This companion supplies original-binary provenance, semantic mappings and narrow data records
for the contract's current rule, question, vector and observation ids. The spec owns the
behavior and acceptance boundary. Original layouts below are evidence navigation aids, not
required implementation representations.

## Original artifacts and methods

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. Artifact ids resolve to exact
fingerprints, sizes and preferred image bases in the [source manifest](../source-manifest.md):
`LOD113C-D2GAME` (`D2Game.dll`), `LOD113C-D2COMMON` (`D2Common.dll`) and `LOD113C-FOG`
(`Fog.dll`). `LOD113C-STORM` identifies the available original loader analysis specimen;
`LOD113C-GAME` identifies the recorded session executable. The manifest's analysis-image hash
correspondence establishes the original specimens analyzed, not every byte loaded during the
historical runtime session. Full historical loaded-DLL fingerprints were not captured.

Static cleanup/audit reads were recorded on 2026-10-12 and full rereads on 2026-10-13, including
the dispatcher, release wrapper, stat-read guard and getter call scan. The original earlier
read's date was not recorded. Methods were instruction disassembly and reconstructions checked
against original bytes; two independent instruction decoders corroborated the per-type helper
where its prior listing was incomplete. The targeted 2026-10-14 getter verification is linked
below. Publication of these records does not constitute a new behavioral verification.

All addresses are module-relative RVAs unless explicitly marked otherwise. Export ordinals
belong to the identified artifact; a PE-aware reader must map an RVA to a file or loaded image.
Function descriptions are analyst labels, not original exported names; the bit-mask data
export's original name is separately identified below. Misleading
older labels were not used to identify behavior: callers, branches and output writes did so.
Detailed private investigation history and original binary fixture encodings were preserved
separately before this public rewrite; public verification does not require that archive.

## Function and data references

`—` means no export ordinal is asserted. Static methods and dates are those above unless a row
states otherwise. Boundary-crossing functions support the named dependency rather than becoming
new rules of this spec.

| Artifact | Function/data RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | `0xDC0B0` | — | shared scaling and mitigation entry | DMI-R01/R02/R03/R05/R06/R09/R10 | Disassembly: classification and reduction setup, dispatcher call, bypass, ordered loop and output total |
| `LOD113C-D2GAME` | `0xDBD00` | — | per-type mitigation | DMI-R06/R09/R10 | Instructions independently decoded against original image; resistance, reduction, percent arithmetic, absorb and field write |
| `LOD113C-D2GAME` | `0xDB3C0` | — | effective resistance | DMI-R07/R10 | Disassembly through return: pierce, penalty, floor, caps and sanctuary |
| `LOD113C-D2GAME` | `0xDAD00` | — | percent then flat absorb | DMI-R08/R09/R10 | Disassembly through return and both accumulator additions |
| `LOD113C-D2GAME` | `0x14D0` | — | ordered multiply/divide helper | DMI-R09/R10; NRF-R05 | Disassembly; shared arithmetic contract is NRF-R05 |
| `LOD113C-D2GAME` | `0xF44B0` | — | arithmetic helper callee | DMI-R10; NRF-R05 | Callee included in recorded arithmetic/no-draw audit; behavior owned by foundations |
| `LOD113C-D2GAME` | `0xDD450` | — | damage-record builder | DMI-R05 inputs; DPO-R05 | Disassembly of bypass flag production from attacker stats; record building is outside mitigation |
| `LOD113C-D2GAME` | `0xDAC10` | — | scale selector | DMI-R10; PVP-R02a | Disassembly; supplies whole-entry vector pairings' scale |
| `LOD113C-D2GAME` | `0xDA7B0` | — | status-length strip | DMI-R03/R10; DPO-R10 | Disassembly; follows event dispatch and precedes bypass/loop |
| `LOD113C-D2GAME` | `0x37030` | — | event-list dispatcher | DMI-R03; DPO-R23/IEP-R07 | Full instruction read on 2026-10-13; traversal/callback contract belongs to IEP |
| `LOD113C-D2GAME` | `0xA19A0` | — | skill do dispatcher | dispatch root; SCD-R06 | Disassembly of selector load and indirect call through do-function table |
| `LOD113C-D2GAME` | `0xA2C40` | — | weapon skill action | dispatch chain | Disassembly: calls queued damage build path |
| `LOD113C-D2GAME` | `0xA3220` | — | do-function slot 2 target | dispatch-table context | Original image table entry; no local behavioral claim |
| `LOD113C-D2GAME` | `0xA2B40` | — | do-function slot 3 target | dispatch-table context | Original image table entry; no local behavioral claim |
| `LOD113C-D2GAME` | `0xDDDE0` | — | queued damage build path | dispatch chain | Disassembly plus rel32 decoding of mitigation call |
| `LOD113C-D2GAME` | `0xDE0C0` | — | damage apply path | alternate entry; DPO-R08c/R08d | Calls shared entry when its scaling/mitigation flag is nonzero |
| `LOD113C-D2GAME` | `0x3AE10` | — | missile-contact caller | alternate entry; DPO-R08c | Supplies apply flag 1 |
| `LOD113C-D2GAME` | `0xDED30` | — | queued apply caller | dispatch count; DPO-R08d | Supplies apply flag 0, so does not mitigate a second time |
| `LOD113C-D2GAME` | `0x755C0` | — | direct shared-entry caller | alternate entry; PVP-R01a | Recorded direct path runs the shared entry twice per record |
| `LOD113C-D2GAME` | `0xF42D0` | — | direct shared-entry caller | alternate entry; DPO-R08c | Recorded direct caller; path ownership remains in pipeline/PvP specs |
| `LOD113C-D2COMMON` | `0x1A2D0` | 11104 | hireling-class category | DMI-R01/R10 | Full instruction read; null/type guards and five fixed class comparisons |
| `LOD113C-D2COMMON` | `0x1A5B0` | 10239 | undead category test (check A) | DMI-R04/R07/R10 | Full instruction read; class bounds and low/high undead bits |
| `LOD113C-D2COMMON` | `0x1A610` | 10255 | demon category test (check B) | DMI-R04/R10 | Full instruction read; class bounds and demon bit |
| `LOD113C-D2COMMON` | `0x603C0` | 10237 | difficulty-row getter | DMI-R07/R10; GDF-R02 | Reconstruction checked against original bytes; clamped index and row pointer |
| `LOD113C-D2COMMON` | `0x33CD0` | 10494 | unit state-test wrapper | DMI-R07/R10; NRF-R12a | Disassembly; calls state bit test |
| `LOD113C-D2COMMON` | `0x37DB0` | — | state bit test | DMI-R07/R10; NRF-R12a | Disassembly; reads state bitfield with exported mask table |
| `LOD113C-D2COMMON` | `0x38B70` | 10973 | unit stat getter wrapper | DMI-R02/R07/R08/R10; NRF-R12b | Guard read on 2026-10-13; full getter-boundary verification on 2026-10-14 in foundations |
| `LOD113C-D2COMMON` | `0x38A80` | — | existing-array stat getter | DMI-R02/R07/R08/R10; NRF-R12c | Original-byte call scan on 2026-10-13; full 2026-10-14 read in foundations |
| `LOD113C-D2COMMON` | `0x382B0` | — | exact stat-identity search | DMI-R10; NRF-R12c | Reconstruction checked against original bytes; leaf without calls or RNG step |
| `LOD113C-D2COMMON` | `0x39430` | 10910 | alternate stat-read wrapper | DMI-R10 | Recorded no-draw callee of strip/selector path; forwards to `0x38A80` |
| `LOD113C-D2COMMON` | `0x1A550` | 10064 | monster-class predicate | DMI-R10 | Recorded strip/selector callee in no-draw audit |
| `LOD113C-D2COMMON` | `0x1A720` | 10278 | monster bit-7 predicate | DMI-R10; PVP-R02a setup | Full instruction read on 2026-10-13; rejects non-monsters before class flag read |
| `LOD113C-FOG` | `0x1CCC0` | 10046 | event-node release wrapper | DMI-R03 dependency | Full 48-byte instruction read on 2026-10-13; forwards to `0x1B0C0` |
| `LOD113C-FOG` | `0x1B0C0` | — | release implementation | outside DMI contract | Internals not fully decoded; no release-implementation behavior is asserted here |
| `LOD113C-D2GAME` | `0x102AB0` | — | mitigation type-binding table (data) | DMI-R06 and section 3 | Original image: twelve `0x2C`-byte entries; historical read-only memory correspondence below |
| `LOD113C-D2GAME` | `0x1074A8` | — | skill do-function pointer table (data) | dispatch root; GDF-R08 | Original image: slots 1–3 identify action targets above |
| `LOD113C-FOG` | `0x26798` | 10259 | `gdwBitMasks` (original exported data name) | DMI-R04/R07/R10; GDF-R18 | Original image: 32 u32 masks, `mask[n] = 1 << n`; export name/ordinal/RVA correspondence checked in the original PE name table on 2026-10-14 |

### Import resolution

These original D2Game thunks distinguish caller locations from resolved target functions.
IAT values below are RVAs in `LOD113C-D2GAME`; export targets are in the artifact shown.
The two release/stat thunks were re-resolved on 2026-10-13.

| D2Game thunk RVA | D2Game IAT RVA | Resolved artifact, ordinal and target RVA | Use |
|---|---|---|---|
| `0xA7F6` | `0xF8538` | `LOD113C-D2COMMON` #11104, `0x1A2D0` | regular-monster flags |
| `0xA508` | `0xF8344` | `LOD113C-D2COMMON` #10239, `0x1A5B0` | check A |
| `0xA814` | `0xF854C` | `LOD113C-D2COMMON` #10255, `0x1A610` | check B |
| `0xA7BA` | `0xF8510` | `LOD113C-D2COMMON` #10237, `0x603C0` | difficulty row |
| `0xA68E` | `0xF8448` | `LOD113C-D2COMMON` #10494, `0x33CD0` | sanctuary state test |
| `0xA42A` | `0xF82B0` | `LOD113C-D2COMMON` #10973, `0x38B70` | layer-0 stat reads |
| `0xA802` | `0xF8540` | `LOD113C-D2COMMON` #10910, `0x39430` | strip/selector callee |
| `0xA580` | `0xF8394` | `LOD113C-D2COMMON` #10064, `0x1A550` | strip/selector callee |
| `0xA7EA` | `0xF8530` | `LOD113C-D2COMMON` #10278, `0x1A720` | scale-selector predicate |
| `0xA2EC` | `0xF8BB8` | `LOD113C-FOG` #10046, `0x1CCC0` | dispatcher release wrapper |

The Fog mask export is imported at D2Game IAT RVA `0xF8BE0` and D2Common IAT RVA `0x890B0`.

## Dispatch chain

1. The skill do dispatcher at D2Game `0xA19A0` loads `SkillsRecord.srvdofunc` at record
   `+0x2E`, indexes the function-pointer table at `0x1074A8`, and calls the selected target.
   GDF-R08 identifies the selector; SCD-R06 owns dispatch eligibility. Slot 1 points to
   `0xA2C40`; slots 2 and 3 point to `0xA3220` and `0xA2B40`.
2. Weapon action `0xA2C40` calls `0xDDDE0`. That queued build path calls builder `0xDD450`
   when damage-record flag `0x2` is clear, then calls shared entry `0xDC0B0`. Its original
   call instruction at `0xDDE2D` was independently decoded from rel32 bytes.
3. `0xDC0B0` computes regular-monster flags and reductions; calls scale selector `0xDAC10`
   and performs the scaling loop (PVP-R02a/DPO-R08a); calls event dispatcher `0x37030` with
   event 11; calls strip `0xDA7B0`; computes bypass; walks the type table calling `0xDBD00`;
   then writes the total (DPO-R12a).
4. Per-type helper `0xDBD00` calls resistance helper `0xDB3C0` at instruction `0xDBD27`,
   arithmetic helper `0x14D0` at `0xDBD65`, and absorb helper `0xDAD00` at `0xDBD78`.
   Its final store writes the selected record field. Absorb adds its two contributions to
   record `+0x48`. These writes anchor the computation independently of descriptive labels.

**Alternate paths.** Apply function `0xDE0C0` calls `0xDC0B0` when its scale/mitigation flag is
nonzero. Missile-contact function `0x3AE10` passes flag 1. Other direct callers recorded in
the pipeline/PvP contracts are `0x755C0` (twice per record) and `0xF42D0`. Queued apply
`0xDED30` passes flag 0, so a queued record is mitigated at build time, without a second pass
on application. DPO-R08c/DPO-R08d and PVP-R01a own the complete path identities and counts.

**Original calling conventions.** `0xDC0B0` takes four stack arguments `(game, attacker,
defender, record)`, returning with `ret 0x10`. `0xDBD00` receives the type entry in `eax`,
context in `ecx`, and bypass on the stack (`ret 4`), returning the mitigated value.
`0xDB3C0` receives entry in `ebx` and context in `edi`, returning resistance. `0xDAD00`
receives value in `eax`, context in `ecx`, entry in `ebx`, returning the post-absorb value.
`0x14D0` receives `(a,b,c)` in `(eax,edx,ecx)`. Dispatcher `0x37030` receives list owner in
`ebx` and stack `(game,event,other,record)`, with `ret 0x10`.

The shared entry's `eax` return is the six-damage-field sum before addition of lifeDrain to
the total; `0xDDDE0` ignores it. This is a recorded binary detail, not an assertion of
DPO-R12a or an added DMI acceptance rule. No function on the computing path establishes a
source-folder name by an assertion string; the neighboring damage pipeline establishes the
combat subsystem relationship.

## Semantic-to-original mapping

| Semantic term | Original encoding | Rule / dependency |
|---|---|---|
| physical/fire/lightning/cold/magic | DamageRecord `+0x08/+0x10/+0x1C/+0x24/+0x20` | DMI-R06 |
| coldLength/freezeLength/poisonLength/poison | record `+0x30/+0x34/+0x2C/+0x28` | DMI-R06; CSA-R01 |
| lifeDrain/manaDrain/staminaDrain | record `+0x38/+0x3C/+0x40` | DMI-R06; CSA-R01 |
| absorbHeal/total/reductionScale | record `+0x48/+0x4C/+0x54` | DMI-R08 / DPO-R12a / DMI-R02; scale producer is DMI-Q05 |
| bypassUndead/bypassDemons/bypassOther | record flags at `+0x00`, masks `0x100/0x200/0x400` | DMI-R05; produced by `0xDD450` from stats 103/104/106 under DPO-R05 |
| context game/difficulty/attacker/defender/record | context `+0x00/+0x04/+0x08/+0x0C/+0x18` | DMI-R01/R02/R06; ten dwords in shared-entry temporary frame |
| regularAttacker/regularDefender | context `+0x10/+0x14` | DMI-R01 |
| reduction slots 0/1/2 | context `+0x1C/+0x20/+0x24` | DMI-R02/R06 |
| type bindings | D2Game `0x102AB0`, twelve entries of `0x2C` bytes; i32 `-1` means absent stat | DMI-R06; engine-defined table |
| entry field/resist/max/pierce/%absorb/flat/slot/regular-attacker marker | entry `+0x00/+0x04/+0x08/+0x0C/+0x10/+0x14/+0x18/+0x1C` | DMI-R06/R07/R08 |
| other type-entry fields | `+0x20` scaling flag, `+0x24` name-string pointer, `+0x28` unused here | DPO-R08a; no local behavior assigned to unused field |
| game difficulty/expansion | Game `+0x6D` u8 / `+0x70` nonzero flag | DMI-R07 |
| selected difficulty row | D2Common #10237; stride `0x58`, ResistPenalty at `+0x00`; array/count at D2Common data-section offsets `+0x7B54/+0x7B58` | GDF-R02; DMI-R07 |
| unit type/class | Unit `+0x00/+0x04`; original type 0 player, 1 monster | DMI-R01/R04 |
| MonStats array/count | DataTables `+0xA78/+0xA80`, record stride `0x1A8` | DMI-R04 |
| lUndead/hUndead/demon | MonStatsRecord `+0x0C` bits 11/12/13; byte `+0x0D` tested with mask indices 3/4/5 | DMI-R04; GDF-R18 |
| unit stat list | Unit `+0x5C` | NRF-R12b; DMI-V27 getter integration |
| stat definition array/count | DataTables `+0xBCC/+0xBD4`, record stride `0x144`; packed key `(id << 16) + u16(layer)` | NRF-R12b/R12c; DMI stat reads |
| defender event list | Unit `+0x90` | DMI-R03; IEP-R07 |
| event entry event/flags/type | byte `+0x00`; flags `+0x02` (1 active callback, 2 removal mark); type `+0x04` (0 one-shot) | IEP-R07; consumed by DMI-R03 |
| callback values/callback/previous/next | node `+0x0C/+0x10`, `+0x14`, `+0x18/+0x1C` | IEP-R07 |
| fire-length field | record `+0x18` | CSA-R01/DPO-R10 setup; not read by mitigation loop |

The type table's used columns encode the spec's twelve semantic entries as follows. The
spec's table supplies the stat names; these are raw numeric bindings and record-field offsets.

| Entry / semantic field | Record offset | Resist | Max | Pierce | % absorb | Flat | Slot | Regular attacker |
|---|---|---|---|---|---|---|---|---|
| 0 physical | `0x08` | 36 | -1 | -1 | -1 | -1 | 1 | 0 |
| 1 fire | `0x10` | 39 | 40 | 333 | 142 | 143 | 2 | 0 |
| 2 lightning | `0x1C` | 41 | 42 | 334 | 144 | 145 | 2 | 0 |
| 3 cold | `0x24` | 43 | 44 | 335 | 148 | 149 | 2 | 0 |
| 4 magic | `0x20` | 37 | 38 | -1 | 146 | 147 | 2 | 0 |
| 5 coldLength | `0x30` | 43 | 44 | 335 | -1 | -1 | 0 | 0 |
| 6 freezeLength | `0x34` | 43 | 44 | 335 | -1 | -1 | 0 | 0 |
| 7 poisonLength | `0x2C` | 110 | -1 | 336 | -1 | -1 | 0 | 0 |
| 8 poison | `0x28` | 45 | 46 | 336 | -1 | -1 | 0 | 0 |
| 9 lifeDrain | `0x38` | -1 | -1 | -1 | -1 | -1 | 0 | 1 |
| 10 manaDrain | `0x3C` | -1 | -1 | -1 | -1 | -1 | 0 | 1 |
| 11 staminaDrain | `0x40` | -1 | -1 | -1 | -1 | -1 | 0 | 1 |

## Rule-linked rationale

| Rule | Original decisions, calculation and output evidence |
|---|---|
| DMI-R01 | D2Game `0xDC0B0` tests monster type and calls D2Common #11104 for attacker and defender, storing regular flags when the category is zero. D2Common `0x1A2D0` returns zero for null/non-monster; unsigned class comparisons give 1 for `0x10F`, 2 for `0x152`, 3 for `0x167`, 4 for `0x230..0x231`, otherwise zero. |
| DMI-R02 | `0xDC0B0` reads layer-0 stats 34/35 through #10973, shifts each left by 8, and scales only when shifted reduction and record scale are positive. The call order is `(scale,red,0x400)`. Slot 0 is zero; results are stored in context slots 1/2 without a clamp. |
| DMI-R03 | After scaling, `0xDC0B0` passes record, attacker, event `0xB`, game and defender owner to `0x37030`; the next step calls `0xDA7B0` without testing dispatcher `eax`. The strip and per-type loop therefore follow even a nonzero result. |
| DMI-R04 | D2Common `0x1A5B0/0x1A610` return false before reading flags for null, non-monster, negative/out-of-range class or zero computed record address. Check A combines byte `+0x0D` mask bits 3/4; B uses bit 5, mapping to MonStats low/high undead and demon. |
| DMI-R05 | After strip, `0xDC0B0` combines checks A/B with record masks `0x100/0x200/0x400`. For non-monsters only `0x400` applies. Class selection is computed once before the loop. |
| DMI-R06 | `0xDC0B0` advances through the twelve type entries, ending when an entry requires a regular attacker and that context flag is zero. `0xDBD00` loads the bound field, zeroes initially nonpositive input, otherwise gets resistance before testing bypass. It subtracts the selected reduction without a clamp unless bypassed; positive remaining values with nonzero resistance call `(100-r,v,100)`, with `r` capped at 100. Non-bypassed values call absorb even after becoming nonpositive. Final store writes the bound field. |
| DMI-R07 | `0xDB3C0` reads resistance if bound; subtracts nonzero attacker pierce unless a regular defender has resistance at least 100; applies difficulty penalty to non-regular defenders except bindings 36/37. Classic selection uses two decrement/equality branches, so only difficulties 1/2 subtract 20/50. Expansion adds row `+0x00` unchanged. A nonpositive result returns immediately with floor -100. Positive non-regular resistance uses cap 50 for binding 36, otherwise 75 or `min(75+maxStat,95)`, with no lower cap bound. Finally binding 36, attacker state 47 and check A can set positive resistance to zero. |
| DMI-R08 | `0xDAD00` returns immediately when percent-absorb binding is absent. Otherwise it caps percent at 40, performs `(pct,v,100)` only for positive percent, and adds the contribution to `record+0x48`. It then shifts flat stat by 8; only a positive result is used, limited to the current value even if that value is negative. The second contribution is added to `+0x48`; the post-subtraction value returns to the field writer. |
| DMI-R09 | Callers `0xDC0B0`, `0xDBD00` and `0xDAD00` supply `(scale,red,1024)`, `(100-r,v,100)` and `(pct,v,100)` respectively to `0x14D0`. NRF-R05 owns the helper's exact branching/rounding; DMI-V26 discriminates caller argument order separately from DMI-V13's arithmetic checks. |
| DMI-R10 | The recorded stage/read/arithmetic call audit found no LCG step or RNG-helper call in the boundary described below. Event callbacks and opaque release internals are not expanded into a new local contract. |

### Dispatcher contract resolution (2026-10-13)

The full read of D2Game `0x37030` (176 bytes) returns zero for null owner; otherwise it starts
result zero and walks the list at Unit `+0x90`. Nonmatching entries continue with the successor
loaded before the event test. A matching entry sets active-callback flag 1, calls its callback
at `+0x14`, clears flag 1, **reloads successor `+0x1C` after the callback**, and stores the
callback result. Callback arguments are game in `ecx`, requested event in `edx`, and stack
`(owner,other,record,value1,value2)`. Removal mark 2 or entry type zero causes unlinking and a
call to release thunk `0xA2EC`; final result is the last matching callback's result.

This resolved DMI-Q02's former sibling discrepancy: one description had read the next entry
before every callback and omitted the event argument. The targeted IEP-R07/DPO-R23 correction
made both owning contracts agree with the instructions; IEP-V08 distinguishes removal of the
next entry, while DPO-V26 states the complete semantic argument list. DMI-V19's placement and
continuation assertions consume those reconciled contracts. Numeric mitigation results and
static labels did not change, and no gameplay observation occurred. See the historical
[item-event correction](combat-item-event-triggers-and-procs.md#dispatcher-correction) and
[pipeline correction](combat-damage-pipeline-order.md#dispatcher-correction).

The release wrapper resolves to Fog #10046 at `0x1CCC0`. It forwards node, file/line metadata
and a conditional pool argument to Fog `0x1B0C0`, then returns 1. The inner release function was
not fully decoded. The recorded call follows unlinking; this evidence does not establish its
opaque internals or add release behavior to mitigation acceptance.

### Stat-reader contract resolution (2026-10-14)

The 2026-10-13 read of D2Common #10973 at `0x38B70` established null-unit/list returns zero,
signed stat-id bounds against the definition count, and a zero test of the computed definition
address. Valid inputs pass the definition address and packed stat/layer identity to `0x38A80`.
That guard read alone did not verify the helper or resolve the then-open DMI-Q01 integration.

On 2026-10-14, full entry-to-return reads of the unit/list wrappers, both getter helpers and
the exact-search leaf settled Q01 at the **existing-array getter boundary**. NRF-R12b/NRF-R12c
now carry verified static labels. The earlier description had assumed summation and a general
shift/clamp; the instructions select one existing array and apply a conditional
MinAccr/ValShift replacement. Original-profile reduction/flat-absorb definitions have fMin
clear, confirmed in named text and compiled cells. See the foundations
[getter evidence](core-numeric-and-rng-foundations.md#stat-reader-verification-2026-10-14) and
[table records](core-numeric-and-rng-foundations.md#stat-table-verification-2026-10-14).

DMI-V27 supplies ordered existing arrays and definitions: cases (a)/(c) give physical
2880/3840 and heal 0; case (b) gives fire 1792 and heal 768. It distinguishes resolved-array
selection, absent-entry zero and unchanged whole-point values before DMI's shift. Array
construction, attached-list aggregation, registration and gameplay are outside that fixture.
The 26 prior vectors were preserved; all 27 remain unobserved. The historical final reader
reported MISMATCH 0, GUESS 0, GAP 0 and recomputed all 27 vectors. Its review/preservation
scope is recorded in the [foundations companion](core-numeric-and-rng-foundations.md#reader-and-preservation-checks-2026-10-14).

### No-draw audit (DMI-R10)

The recorded audit covers D2Game `0xDC0B0`, `0xDBD00`, `0xDB3C0`, `0xDAD00`, `0xDAC10`,
`0xDA7B0`, arithmetic helpers `0x14D0/0xF44B0`, and the D2Common read/classification callees:
#10973 → `0x38B70` → `0x38A80` → `0x382B0`; #10494 → `0x33CD0` → `0x37DB0`;
#10239/#10255/#11104/#10237; #10910 → `0x39430` → `0x38A80`; #10064/#10278.
The 2026-10-13 original-byte scan added the getter call at `0x38A95` to `0x382B0`; that
leaf has no further call or LCG constant. The stage's zero-draw conclusion was
unchanged. Event callbacks are explicitly excluded: no claim is made that their bodies draw
nothing. The opaque release implementation is not promoted to a newly verified local helper.

### Scale selection evidence

PVP-R02a's verified selector at D2Game `0xDAC10` supplies `P=100` for the common setup's
regular-monster attacker/player defender and player attacker/regular-monster defender,
when the attacker is not a hireling and its PVP-R02b flag is clear. The 2026-10-13 full read
of D2Common #10278 at `0x1A720` rejects null/non-monster and invalid class before checking
MonStats flag bit 7; a player therefore fails this predicate and #11104's hireling test.
The older pipeline DPO-R08b label was asm-only while the PvP contract supplied this verified
selector slice; this evidence did not upgrade that separate label. Other pairings in DMI
vectors remain supplied-input mitigation-boundary cases where stated.

## Table-data records

These reads used the running game's Storm loader on 2026-10-13 over its loaded archive set.
The historical source profile was reported as original Lord of Destruction 1.13c, with LoD
catalog/family identity and no known data overrides. Its recorded executable fingerprint is
`LOD113C-GAME`. That hash alone is insufficient to identify vanilla data because modded
profiles can share the executable. A complete data/override inventory, effective-file hashes,
physical supplying MPQs and full historical live-DLL identities were not captured. The
manifest's current DLL fingerprints identify available original analysis specimens; they do
not retroactively supply those missing runtime identities.

Columns were selected by header name; complete returned bytes, rather than truncated previews,
were used for target cells. Only the named cells below are published. Archive reads establish
their data contents, not gameplay execution. These records' compiled limits apply to the
2026-10-13 reads; the later specific getter metadata reads are separately linked above.

**TD-1 — `data/global/excel/itemstatcost.txt`.** Total 40109 bytes. Read intervals: 0–500
(header), 500–3500, 3900–4900, 4900–5900, 11000–11300, 11300–12200, 12200–13200,
15200–15500, 15400–17100, 36500–36800, 37400–38700. Header and each target row complete;
not a full-table read. Row identity is column `Stat`; numeric value is column `ID` (second
header field). Whole-file positional row-index correspondence was not checked.

| `Stat` row | Raw `ID` |
|---|---:|
| `normal_damage_reduction` | 34 |
| `magic_damage_reduction` | 35 |
| `damageresist` | 36 |
| `magicresist` | 37 |
| `maxmagicresist` | 38 |
| `fireresist` | 39 |
| `maxfireresist` | 40 |
| `lightresist` | 41 |
| `maxlightresist` | 42 |
| `coldresist` | 43 |
| `maxcoldresist` | 44 |
| `poisonresist` | 45 |
| `maxpoisonresist` | 46 |
| `skill_bypass_undead` | 103 |
| `skill_bypass_demons` | 104 |
| `skill_bypass_beasts` | 106 |
| `item_poisonlengthresist` | 110 |
| `item_halffreezeduration` | 118 |
| `item_absorbfire_percent` | 142 |
| `item_absorbfire` | 143 |
| `item_absorblight_percent` | 144 |
| `item_absorblight` | 145 |
| `item_absorbmagic_percent` | 146 |
| `item_absorbmagic` | 147 |
| `item_absorbcold_percent` | 148 |
| `item_absorbcold` | 149 |
| `item_cannotbefrozen` | 153 |
| `passive_fire_pierce` | 333 |
| `passive_ltng_pierce` | 334 |
| `passive_cold_pierce` | 335 |
| `passive_pois_pierce` | 336 |

Interpretation: name/id identities consumed by DMI section 3 and GDF-R11. Code fixes stat
selectors in D2Game `0xDC0B0/0xDD450` and the type table `0x102AB0`; finding a numeric id
in data does not make those consumers dynamically remapped. `itemstatcost.bin` was not read
in TD-1; the later named getter-metadata correspondence is a separate 2026-10-14 record.

**TD-2 — `data/global/excel/states.txt`.** Total 20920 bytes. Reads: 0–300 (header),
5000–5600, 5800–6500. Coverage is header and target row, not the full table. Row by
`state=sanctuary`, column `id`, raw value `47`. Interpretation: state identity (GDF-R12);
D2Game `0xDB3C0` fixes selector `0x2F` for DMI-R07. `states.bin` was not read.

**TD-3 — `data/global/excel/difficultylevels.txt`.** Total 586 bytes, full read 0–586.
Rows by `Name`, with zero-based data-row indices excluding header; `ResistPenalty` is the
second header field. Named raw values:

| `Name` / data-row index | `ResistPenalty` | `HireableBossDamagePercent` |
|---|---:|---:|
| Normal / 0 | 0 | 50 |
| Nightmare / 1 | -40 | 35 |
| Hell / 2 | -100 | 25 |

Compiled `data/global/excel/difficultylevels.bin` was fully read (268 bytes): count dword 3,
then three `0x58`-byte records starting at offset 4. Signed dwords at record `+0x00` are
0/-40/-100 and at `+0x38` are 50/35/25, agreeing with the named text columns. D2Common
#10237 selects the row; D2Game `0xDB3C0` reads `+0x00` for DMI-R07. This establishes the
stated field correspondence, not general equivalence of every text/compiled cell.

**TD-4 — `data/global/excel/monstats.txt`.** Total 432634 bytes. Reads: 0–200 and
200–1100 (header), 48000–48900, 52000–52900, 156700–157300, 157200–158100, 174000–174900,
192000–192900, 193200–193900, 195500–196400, 202100–203000, 203500–204400, 314500–315400,
316700–317600, 317700–318400, 333000–333900. Header has `Id`/`hcIdx` as first two fields
and columns `lUndead`, `hUndead`, `demon`. Row identity is `Id`; values are the explicit
`hcIdx` column, without claiming positional row-index identity.

| `Id` row | Raw `hcIdx` |
|---|---:|
| `sandleaper1` | 78 (`0x4E`) |
| `roguehire` | 271 (`0x10F`) |
| `act2hire` | 338 (`0x152`) |
| `act3hire` | 359 (`0x167`) |
| `act5hire1` | 560 (`0x230`) |
| `act5hire2` | 561 (`0x231`) |
| `baaltentacle1` | 562 (`0x232`) |

Coverage is only header and these target rows. Their category flag cells were not recorded;
vectors supply class flags and a synthetic count, rather than asserting those actual rows'
flags or the original total row count. D2Common `0x1A2D0` fixes hireling class comparisons;
`0x1A5B0/0x1A610` consume category flag bits (DMI-R01/R04). `monstats.bin` was not read.

**TD-5 — `data/global/excel/events.txt`.** Total 420 bytes, full read. Column `event` at
zero-based data row 11 (excluding header) has raw value `absorbdamage`. IEP owns the
row-index-to-event-id convention. D2Game `0xDC0B0` fixes the dispatched event immediate
`0xB`; TD-5 supplies the name, rather than proving that callback dispatch occurred.

## Search and runtime records

Discovery followed the original do-function pointer table and callers into `0xDC0B0`, then
the per-type output store and absorb accumulator writes. Type-table references and the
fixed reduction/resistance/state/bypass selectors anchored the numerical branches. No rule
was recorded as an unsuccessful function search. Private search commands are not needed to
follow the public dispatch and output anchors above.

One read-only live memory corroboration was recorded on 2026-10-13. No executable vector was
run; no gameplay was launched, driven or instrumented by that evidence pass.

| Vector | Source profile | Date | Module + RVA | Records | Result / limitation |
|---|---|---|---|---|---|
| none; type-table correspondence only | reported original LoD 1.13c profile, subject to manifest identity limits | 2026-10-13 | D2Game `0x102AB0`, 528 bytes | one read | All 132 dwords agreed with the analysis image except the twelve name-string pointers at entry `+0x24`, relocated by a common delta. This corroborates the table, without proving gameplay or complete loaded-DLL identity. |

All DMI-V01–V27 retain `runtime_observed: no`. Loader cell reads are data evidence;
read-only table memory correspondence is not a gameplay observation. DMI-O01–O03 remain
proposed, and document status remains draft.

## Corrections and historical reviews

These summaries preserve correction rationale and the actual scope of earlier reviews. Their
counts describe the texts reviewed on their dates, not a fresh review of this public rewrite.

- **Absorb correction.** The earlier reading applied percent arithmetic and flat subtraction
  unconditionally. D2Game `0xDAD00` skips the whole step with no percent binding, skips percent
  at nonpositive percent and skips flat at nonpositive shifted flat stat. This matters for
  negative inputs; DMI-R08 records the corrected reading.
- **Classification labels.** Undead/demon tests were upgraded from unverified after full reads
  of D2Common #10239/#10255. The regular-monster predicate received its own labeled rule and
  vector after #11104 was read. The old reduction-slot names both denoted context
  `+0x1C..+0x24`; the current spec uses semantic slots.
- **Event dispatch scope, 2026-10-13.** DMI-R03 changed from asm-only to verified static when
  the original dispatcher and call site were read. It was narrowed to local placement,
  owner/other/record and ignored result, consuming dispatcher lifecycle from its owner.
  The initially discovered sibling conflict became Q02, later settled as recorded above.
- **Early document migration, 2026-10-13.** Formatting/semantic terminology changed with no
  source, archive or gameplay reread. Its checks preserved vector values and resolved 63
  local links; initial/final readers reported no GUESS or GAP. Those results apply only to
  that historical text.
- **Source/migration review, 2026-10-13.** The static pass supplied TD-1–TD-5, dispatcher
  release resolution, stat guards, the getter leaf-call addition and PVP scale-selection
  support. Gameplay was not attempted. Reader passes reported respectively
  `(MISMATCH,GUESS,GAP) = (1,8,6), (0,7,3), (0,8,7), (0,1,7), (0,0,5)`; documented gates
  and advisory coverage findings were triaged separately. Revisions clarified supplied-input
  versus shared-entry cases, synthetic class/count inputs, negative reductions/caps/absorb,
  non-monster bypass and magic-penalty exemption. A proposed shared-entry return assertion
  was removed because DPO-R12a defines total, not return value. The preservation check kept
  existing expected values; added vectors and cases were calculated expectations, not
  observed results. The then-current link check recorded 126 links and no broken target.
- **Acceptance/gate review, 2026-10-13.** Integration, observation and outside-boundary gaps
  were distinguished. V19's old final callback returned zero, so it could not demonstrate
  ignoring a nonzero final result. The corrected fixture returns 7 and observes coldLength 40
  before stripping to zero, while physical mitigation still yields 2880. V26 added the three
  caller-order discriminators: resistance 75001 versus swapped 75000; physical 75840 versus
  swapped 74998; fire/heal 60002/40001 versus swapped 60003/40000. A fresh reader recomputed
  all 26 then-current vectors with MISMATCH 0, GUESS 0, GAP 0; targeted discrimination review
  found no required correction. Source rules/labels did not change in that review.
- **Dispatcher correction review, 2026-10-13.** The targeted reader of reconciled sibling
  contracts returned MISMATCH 0, GUESS 0, GAP 0 and no required discrimination finding.
  It resolved Q02, without a full legacy-spec migration or gameplay execution.
- **Getter correction review, 2026-10-14.** Q01 settled at existing-array retrieval and V27
  was added, with the exact scope and historical reader result recorded above. Existing
  implementation comparison findings were not refreshed.

Earlier review records describing Q01/Q02 as open are superseded by their dated resolutions.
Current remaining gaps are [DMI-Q03/Q04 observations and DMI-Q05's outside-boundary scale
producer](../../plans/gaps/combat-damage-mitigation.md). Q05 blocks no supplied-input local
computation. Advisory unexercised branches did not expand getter fixtures into array-production
acceptance or imply exhaustive branch coverage. No new labels, vectors or gameplay results
are established by this provenance-only update.

## Observation procedures

These procedures are pending; none has been run. Use a debugger or instrumentation on owned
matching original artifacts and establish the effective data profile before interpreting a
new observation. Locations map to semantic fields above. Record actual inputs, entry/exit
values and relevant seeds, so the comparison uses measured inputs. Prepare controlled items
or a copy of a save where needed; exact stats must be measured rather than inferred from item
descriptions. Keep observations separate when all desired conditions cannot occur in one hit.

### DMI-O01 reproduction

Use a level-1-or-higher character of any class in a Normal single-player game, without a party,
in Act I Blood Moor. Prepare defender stats 34/36 for DMI-V01's physical case (5 and 25),
and separately prepare the maximum-resistance and absorb controls of DMI-V03/DMI-V05.
For absorb, apply V05's arithmetic using the measured fire input and fire bindings 142/143.
Measure defender stats 34/36/39/40/142/143 and attacker stat 333 at hit time. Let one regular
monster, such as a `fallen1`-family monster, hit once. Capture physical/fire at entry and exit
of D2Game `0xDBD00` for type entries 0/1, with record reached through context `+0x18`, and
absorbHeal at `0xDC0B0` exit. A hit without a positive fire component does not exercise the
fire/cap/absorb branches; obtain a measured suitable hit for those comparisons.

Expected output is DMI-R02 and DMI-R05–R08 applied to the measured values, including the
relevant table bindings and recorded expansion/difficulty. Repeat with one input changed
at a time: stat 34 0 versus 5, stat 40 0 versus 20, stat 142 0 versus 60. Verify the other
inputs and incoming damage for each comparison; differing rolls require recomputation from
the captured entry value. These controls target reduction, maximum-resistance and absorb
caps, not arbitrary final life change.

### DMI-O02 reproduction

Use any character able to enter expansion Hell, with defender stat 110 measured as zero and
stat 45 recorded, and a regular-monster attacker. Find an attack whose measured poisonLength
is positive at D2Game `0xDC0B0` entry. Record difficulty/expansion, attacker classification,
all relevant pierce/bypass/status inputs, and poisonLength (`record+0x2C`) on entry/exit.
When the hit also carries lifeDrain (`+0x38`), record its entry/exit values separately.

Under DMI-V16's conditions poisonLength exit is `mulDiv(200,entry,100)`; under DMI-V15's
conditions the regular-monster lifeDrain input has the same multiplier. The control is the
same attack in Normal with stat 110 zero, where poisonLength is unchanged under the vector's
other inputs. Compare each hit using its measured entry rather than assuming identical
rolls. This observes the static penalty-on-length/drain reading and addresses DMI-Q04.

### DMI-O03 reproduction

Use Normal difficulty and a level-1 character of any class with measured stat 34 at least 3,
prepared through items or a save copy. Let a `fallen1`-family regular monster hit once,
repeating until the measured physical input is below 768 for stat 34 = 3. Capture the full
incoming record, stat 36 and other contributing channels to isolate DMI-V08's conditions.
Read physical (`+0x08`) at entry/exit of D2Game `0xDBD00` for entry 0 and total (`+0x4C`)
at `0xDC0B0` exit, then life before/after later application.

With DMI-V08's other inputs, expected physical is `input-(Stat(34)<<8)` when negative, total
has that same negative value, and DPO-R17 leaves life unchanged for a nonpositive total.
For another measured stat-34 value, choose input below its shifted reduction. Repeat with
flat reduction zero as the control, recording incoming damage and other channels again.
This is the pending DMI-Q03 cross-check, not a runtime-confirmed result.

## Public provenance update (2026-10-14)

This pass translated the existing evidence into original artifact ids, module-relative
locations, rule-linked rationale and semantic mappings after the complete original records
were privately preserved. It retained narrow TD-1–TD-5 cells, dated correction rationale,
read-only memory scope and pending observation anchors. Behavioral rules, 27 vectors, labels,
ownership and gap boundaries remain those in the current spec; publication does not close
an evidence gap or promote confirmation status. No new gameplay or data-read campaign was
performed. Document checks for this rewrite are reported separately by the coordinating pass.

## Public provenance checks (2026-10-14)

Exact before/after comparison preserved the entire spec before Appendix A, all 27 vector
definitions, labels and gates, and the complete gap file including its prior implementation
comparison. Complete original specs, companions and gap files were preserved privately and
checked against the pre-edit snapshot; archive locations and SHA-256 values are private.

The five manifest binary hashes/sizes matched two stored original bundles. Read-only PE
checks mapped all 37 function/data table references into original image sections and matched
all 11 cited export targets. The original Fog export-name table also confirmed #10259's
name, gdwBitMasks, at RVA 0x26798. These are identity/navigation checks, not a new behavioral
verification or implementation of the opaque release path.

Combined checks for both migrated specs, companions and shared manifest resolved all 69
local links/anchors, including 17 inbound companion links. Public citation scans found no
private source paths, reconstruction filenames, commits, member identifiers or tool calls.
Focused cross-review found no remaining required provenance/contract issue; the complete
dispatch chain from the skill do-function table to mitigation and measured fire bindings in O01 were checked in the final
text. No implementation tests, vector recomputation, new table reads or gameplay actions
were performed. Existing labels, gaps, pending observations and draft status are unchanged;
this record neither approves the spec nor establishes runtime parity.
