# Evidence: Skill Calc Expressions and Synergies

Spec: [Skill Calc Expressions and Synergies](../../specs/skill-calc-expressions-and-synergies.md)

## Source identity and evidence reuse

The original Windows x86 Lord of Destruction Expansion 1.13c binaries are
`LOD113C-D2COMMON` and `LOD113C-FOG` in the [source manifest](../source-manifest.md).
Their preferred bases are 0x6FD50000 and 0x6FF50000 respectively; all function
locations below are module-relative RVAs. Descriptions are analyst labels unless
an export ordinal is explicitly stated. A private full original-document archive
preserves the earlier investigation and fixture history. Public claims below can
be checked using the fingerprinted original files without that archive.

The 2026-10-12 record read the Fog evaluator in original-image disassembly, the
compiler in a decompiler view, and D2Common front ends and consumers in instruction
listings or a byte-checked source reconstruction. Compatible instruction coverage is
reused with that date and its limits. The current manifest establishes specimen
correspondence, not a new behavior reading or gameplay observation. No original
table-data read or runtime confirmation was recorded in that document. Fields and
stat names consume the separate GDF evidence, preserving that evidence's dates and
profile limits.

Targeted readings on 2026-10-15 checked the unresolved to-hit/passive selectors,
stat routing and leaf candidates, the cap pointer writer, missile dependencies,
additional-family front end, and mastery decoding. New readings chose a listed
byte-checked source reconstruction when present and otherwise instruction listings.
Mastery's undecoded instruction runs were corroborated against the fingerprinted
original image. An unread owned dependency or unresolved original-byte decoding
does not acquire verification through its caller's verified arithmetic.

## Dispatch, returns, and original locations

| Artifact | Function RVA / export | Analyst description and supported rules | Method, date, boundary |
|---|---|---|---|
| LOD113C-FOG | 0x19E90, export #10253; local pop helper 0x19CB0; jump table 0x1A440 | evaluator, SCX-R01 through SCX-R10 | Original-image disassembly, 2026-10-12; opcode dispatch to return, local stack/pop behavior |
| LOD113C-FOG | 0x1A4A0, export #10254 | compiler producer, SCX-R12 | Decompiler lead, 2026-10-12; targeted instruction inspection 2026-10-15 still leaves owned encoding/token helpers and several byte-run paths unread |
| LOD113C-D2COMMON | 0x9270 import thunk -> Fog #10253 at 0x19E90 | resolved evaluator import, SCX-R01/SCX-R11 | 2026-10-12 instruction/import record; thunk is distinct from interpreter |
| LOD113C-D2COMMON | 0x51BF0 and 0x51B80 -> 0x9270 | skill and secondary skill/description front ends, SCX-R11 | Byte-checked source reconstruction / instructions, 2026-10-12; pool check, context construction, evaluator return |
| LOD113C-D2COMMON | 0x60850, 0x607D0; loader 0x61EC0 | compiler-result append/store and field binding, SCX-R11 | Original instruction record, 2026-10-12; compiler return length supplied, not compiler internals |
| LOD113C-D2COMMON | 0x51B60 -> 0x51070, export #10147 | parameter callback and skill resolver, SCX-R04/SCX-R20 through SCX-R28 | Instructions, 2026-10-12; selector dispatch; unresolved leaf results split in current rules |
| LOD113C-D2COMMON | data RVA 0x99E24, count at 0x99E5C | seven skill bindings, SCX-R13 | Original data/instruction reading, 2026-10-12; binding entries mapped to functions below |
| LOD113C-D2COMMON | 0x4F0D0, 0x4F0C0 | min/max, SCX-R14 | Original disassembly, 2026-10-12; signed compare to return |
| LOD113C-D2COMMON | 0x50CB0 -> 0x1180 | skill random range, SCX-R15 | Original disassembly, 2026-10-12; range checks, seed selection, draw/reduction and return; numeric primitive NRF-R01/NRF-R02 |
| LOD113C-D2COMMON | 0x51AF0 -> 0x4E480, 0x4FCB0, 0x51070 | skill(), SCX-R16a/SCX-R16b | Instructions, 2026-10-12; lookup and bonus request, conditional addition, clamp, parameter request; lookup/bonus candidates reread 2026-10-15, full dependencies remain SCX-Q09 |
| LOD113C-D2COMMON | 0x50C90 -> 0x6A790, export #10347 | miss(), SCX-R17 | Original disassembly, 2026-10-12; null context and exact forwarded request/return |
| LOD113C-D2COMMON | 0x50C30 -> 0x31EA0, 0x38C20, 0x39430, 0x38B70 | stat(), SCX-R18a/SCX-R18b | Original disassembly, 2026-10-12; full routing instruction read 2026-10-15 |
| LOD113C-D2COMMON | 0x38C20 export #10587 -> 0x38B00 | qualifier-one stat read | Instructions, 2026-10-15; unit/id/list guards, layer-zero identity formation, own-array leaf candidate |
| LOD113C-D2COMMON | 0x39430 export #10910 -> 0x38A80 | qualifier-two stat read | Instructions, 2026-10-15; selected-array leaf candidate; byte-run corroboration/dependency reconciliation remains explicit |
| LOD113C-D2COMMON | 0x38B70 export #10973 | ordinary stat read | Consumed NRF-R12b/NRF-R12c, whose own evidence record establishes array/minimum semantics; no fresh promotion here |
| LOD113C-D2COMMON | 0x51AB0 -> 0x51070 twice | sklvl(), SCX-R19 | Instructions, 2026-10-12; ordered requests and unchanged second return |
| LOD113C-D2COMMON | 0x1550 through 0x1730 | raw/linear parameter helpers, SCX-R21 | Instructions, 2026-10-12; nonpositive guard and wrapped progression |
| LOD113C-D2COMMON | 0x4DC30; callers 0x4E680,0x4E630,0x4E5E0 | diminishing helper/guards, SCX-R22 | Instructions, 2026-10-12; signed divisions, upper clamp; fourth pair's unguarded call |
| LOD113C-D2COMMON | 0x51070 | mana, nested calculation, damage/level/missile selector dispatch, SCX-R23 through SCX-R27 | Instructions, 2026-10-12; corresponding result contracts/gates in spec |
| LOD113C-D2COMMON | 0x4EA50 -> 0x9270 | to-hit selector result, SCX-R25b | Instructions, 2026-10-15; guard, empty-reference fallback, direct evaluator return |
| LOD113C-D2COMMON | 0x4E7A0 -> 0x9270 | passive selector first match, SCX-R28 | Instructions, 2026-10-15; first of five slots, fixed pairs, aligned calc and unchanged return |
| LOD113C-D2COMMON | 0x4DDB0 | five-band level helper, SCX-R29 | Original disassembly, 2026-10-12; each signed level branch and summed result |
| LOD113C-D2COMMON | 0x50460 export #10121; 0x50360 | elemental min/max, SCX-R30 | Original disassembly, 2026-10-12; shift before synergy, minimum-only gate, optional mastery after synergy |
| LOD113C-D2COMMON | 0x4F870; selector map 0x4F91C; targets 0x4F908 | mastery selection, SCX-R31 | Original disassembly, 2026-10-12; instruction/map corroboration against matching image 2026-10-15; cold wrapper reconciled with the shared stat-read contract |
| LOD113C-D2COMMON | 0x11E0 | argument-ordered percent primitive | 2026-10-12 reading agrees with consumed NRF-R05; current contract uses that shared rule |
| LOD113C-D2COMMON | 0x52100; 0x51FF0 export #10297 | physical min/max, SCX-R32a/SCX-R32b | Instructions, 2026-10-12; source-share/synergy/shift sequence; source and Kick leaf integration remain gated |
| LOD113C-D2COMMON | 0x4F930; 0x31B90; 0x1FB80; 0x38ED0; 0x38F60 | Kick and weapon/item source dependencies | Prior caller references, 2026-10-12; no complete result proof recorded, SCX-Q07 |
| LOD113C-D2COMMON | 0x4E900 | elemental duration, SCX-R33 | Instructions, 2026-10-12; three level bands, base, synergy and return |
| LOD113C-D2COMMON | 0x6B2B0 -> 0x9270; callback 0x6B160 -> 0x6A790 | missile front end, SCX-R34a/SCX-R34b | Original disassembly, 2026-10-12; type/fallback/pool/context path; complete resolver remains separate |
| LOD113C-D2COMMON | 0x69DB0,0x69DA0,0x6B140,0x6B430; bindings data RVA 0x99C30, count 0x99C50 | missile functions, SCX-R34a | Original disassembly/data, 2026-10-12; missile seed and owner skill request distinguished |
| LOD113C-D2COMMON | 0x6A790 -> 0x6B360,0x6A3E0,0x69DC0,0x6A580,0x6A4B0 | missile resolver/result candidates | Targeted instructions, 2026-10-15; dispatch and first three helpers inspected, remaining helpers/byte-run paths and field bindings leave SCX-Q05 open |
| LOD113C-D2COMMON | 0x23CE0 -> 0x9270, callback 0x21FF0; bindings data RVA 0x9D354,count 0x9D374 | additional family, SCX-R35a/SCX-R35b | Byte-checked source reconstructions, 2026-10-15; complete pool check, two-value context, callback return zero; function bindings/owner unresolved |
| LOD113C-D2COMMON | 0x60920, cap pointer data RVA 0xA0B50 | cap producer candidate, SCX-Q03 | Byte-checked source reconstruction, 2026-10-15; Experience loader stores the record pointer; Amazon-to-first-dword binding corroborated against original code/string bytes; compiled-record correspondence remains unresolved |

The expression dispatch root is the selected family front end, not a periodic
game callback: 0x51BF0/0x51B80, 0x6B2B0, or 0x23CE0 -> import thunk 0x9270 ->
Fog #10253 at 0x19E90. The evaluator's parameter cases invoke the supplied callback
before pushing its returned integer. Skill callback 0x51B60 forwards to resolver
0x51070. Function opcode 1 selects the identified binding table, calls its selected
operation, and pushes the returned integer. Each terminating return is in Fog
0x19E90; pool rejects return zero in the front end. This anchors both entry and
result without claiming which server skill chooses a given column.

Game-root anchor (spec review, 2026-10-15): one complete reached chain from a server
dispatch table is D2Game skill do-function table RVA `0x1074A8`, entry 38 -> D2Game
`0x93340` -> physical fill `0x9F1A0` -> import thunk `0xA766` -> D2Common #10297
`0x51FF0` -> skill front end `0x51BF0` -> thunk `0x9270` -> Fog #10253 `0x19E90`.
The same stage's elemental fill `0x9F210` reaches D2Common `0x50460` through thunk
`0xAE92`; `0x50460` calls `0x9270`
directly for EDmgSymPerCalc and then mastery `0x4F870`. The table, D2Game links and
`0x51FF0`/`0x50460` call sites were re-read in that review and are also recorded in the
[leveled-damage companion](skill-leveled-damage-tohit-and-elements.md#dispatch-and-functions).
The review also re-read the Fog evaluator from the original image: the 22-entry opcode
jump table at Fog `0x1A440` maps opcode 1 to the function call, 2 and 3 to the terminator
path, 4/5/6 to byte (zero-extended)/word (sign-extended)/dword parameters, 7/8/9 to literals,
10 through 15 to signed `<`,`>`,`<=`,`>=`,`==`,`!=`, 16 through 22 to add, subtract, multiply,
guarded divide, power, negate and conditional. This agrees with SCX-R01 through SCX-R10.
D2Common `0x4DC30` (diminishing), the unguarded fourth-pair call inside `0x51070` and that
routine's selector 0x16 mana arithmetic (`*25`, truncating `/2`, shift, arithmetic `>>8`)
agree with SCX-R22 and SCX-R23.

## Semantic mapping of original evidence

These are navigation details, not required storage or API designs.

| Semantic term | Original evidence representation |
|---|---|
| skill/secondary/missile/additional pool | DataTables pointers and sizes +0x40/+0x44, +0x4C/+0x50, +0x60/+0x64, +0x98/+0x9C |
| skill context | three original consecutive dwords: unit, skill identity, level |
| missile context | four original consecutive dwords: missile, owner, missile identity, level |
| skill record | stride 0x23C; DataTables array +0xB98, count +0xBA0 |
| hard level and own-skill condition | selected skill entry +0x28 and owner marker +0x34 equal to -1 |
| selected entry candidate | 0x4E480 walks original skill entries; filters disabled marker +0x3C; own marker influences preference; a complete shared entry-selection contract is still missing |
| to-hit base, per-level increment, optional calc | SkillsRecord +0x198,+0x19C,+0x1A0 correspond to ToHit, LevToHit, ToHitCalc; original instructions0x654A2..0x6551C and strings0x8AE18/0x907BC/0x907B0 corroborated 2026-10-15 |
| passive identities and aligned calculation | five signed words starting +0x98; aligned i32 references starting +0xA4 |
| fixed mastery selection | table at RVA 0x4F91C indexed by EType-1; jump targets at 0x4F908; unsupported entries select zero return |
| cold source | unit list +0x5C; definition count +0xBD4 >331; definition index 331; identity encoded as 0x14B0000; direct call 0x38A80 |
| ordinary mastery source | 329/330/332 at layer 0 through 0x38B70; percentage tail jump0x11E0 |
| physical source selection condition | 0x1FB80(unit,unit inventory association) returns 2; meaning/complete dependency remains open |
| skill cap value | first dword through pointer RVA 0xA0B50, written by 0x60920; not an immutable binary literal |
| unit seed | original unit seed fields beginning +0x20; semantic ownership is specified independently |

Skills descriptor 0x61EC0 binds Param1..8 at +0x148..+0x164, recursive calc
references at +0x60..+0x7C,+0xA4..+0xB4,+0xC0,+0x138..+0x144,+0x170, mana fields
at +0x188/+0x18A/+0x18C, physical damage fields at +0x1A4..+0x1D8, elemental
damage fields at +0x1DC..+0x210, and duration fields at +0x214..+0x224.
These field-name correlations retain the 2026-10-12 original descriptor reading;
they do not reproduce a game-data row or claim a fresh archive read.

## Targeted findings and corrections (2026-10-15)

1. SCX-R25 formerly grouped a verified selector with unread postprocessing for
   selector 0x14. Reading0x4EA50 shows no postprocessing: valid nonempty calc returns
   Fog's integer unchanged. Empty calc instead computes the base plus `(L-1)` times
   its increment. SCX-R25a/SCX-R25b now separate endpoint selection and this result.
   Original descriptor instructions0x654A2..0x6551C and string bytes at 0x8AE18/0x907BC/0x907B0 establish the field names ToHit, LevToHit, ToHitCalc; the naming gap is closed.
2. SCX-R28's earlier unresolved body at 0x4E7A0 is now traced from resolver dispatch
   to the evaluator return. The fixed stat pairs match either member, scanning
   passive slots in order; only the first match's aligned calc is evaluated.
3. SCX-R18's earlier asm-only label overstated uncertainty in the routing and
   understated the missing read contracts.0x50C30 has plain instructions that
   establish every guard and selection. SCX-R18a accepts that supplied-operation
   stage; SCX-R18b retains the special to-hit result integration gap. The same leaf mapping closes qualifier-one/two source semantics: own-array for qualifier one, selected-array for qualifier two. Targeted
   leaf reads found own-array 0x38B00 and selected-array 0x38A80; the shared NRF
   contract is the authority for already verified ordinary getters.
4. The mastery mapping in the earlier SCX reading agrees with the mapping in
   SLD-R10. At the time of that earlier comparison, SLD's decoding/integration
   limits remained unresolved; its current R10 reuses the compatible SCX/NRF evidence.
   New original-image inspection of 0x4F870 corroborates instruction and
   table bytes, including the cold-specific direct 0x38A80 call and ordinary
   0x38B70 calls for other selected elements. This settles the decoding ambiguity. The compatible NRF evidence identifies the exact leaf at 0x38A80 as the selected-array/minimum operation and 0x38B00 as the own-array/minimum operation. The cold wrapper therefore consumes NRF-R12b/NRF-R12c directly. It does not force the base/own array. No gameplay result was collected.
5. SCX-R35's pool check and parameter callback now have complete matched-source
   readings at 0x23CE0/0x21FF0. SCX-R35a verifies those operations. The four function
   results and pool owner remain SCX-R35b's behavior gap.
6. Cap pointer search found writer 0x60920 and the Experience loader; the pointer
   identifies loaded data rather than a static cap constant. Original image instructions and strings separately establish Amazon at first-record offset 0; a narrow original-loader text read supplies MaxLvl/Amazon99. Loaded compiled-record correspondence is still not inferred from the text read.
7. SCX-R16/SCX-R26 and SCX-R32 formerly mixed caller arithmetic with incomplete
   entry/bonus/source/Kick dependencies. Supplied-input assertions are preserved;
   their integration gates are separate. All 23 original numeric examples and
   seeds remain in the expanded executable vectors; added cases distinguish
   call order, forwarding, overflow pushes, and parameter selection.

## Remaining evidence assessment

The [gap file](../../plans/gaps/skill-calc-expressions-and-synergies.md) records
every remaining question and attempted slice. Compiler inspection newly reached
the append/fold paths but did not resolve owned parser/encoding helpers or every
byte-run decode; a recovered assertion is not promoted. Missile inspection reached
the resolver and the first min/max/length candidates, but uncorroborated byte runs,
additional elemental helpers, and skill-slot field correspondence leave the result
contract open. Cap/name gaps concern producer/identity questions; they do not block
supplied-value arithmetic. No full search ladder is claimed and no function is
labeled absent merely because it was missing from an earlier inventory.

No gameplay campaign was performed during regeneration. SCX-Q12 therefore remains
an observation gap for every static behavioral rule. All executable vectors are
hand-derived from the contracts and declare `runtime_observed: no`.

## Elemental observation

SCX-O01 is pending. Use a matching vanilla Expansion 1.13c Sorceress with a learned
elemental skill and a related damage-synergy skill, in Normal difficulty, standing
in town with stable equipment and no temporary buffs. Record character/skill
levels, learned synergy levels, relevant mastery stats, selected Skills fields,
and the caller's mastery flag before changing a synergy point. Measure raw min/max
returns at D2Common 0x50460/0x50360 for the same skill context before and after a
controlled synergy increase. Calculate each expectation from the recorded fields,
using SCX-R30/SCX-R31. The tooltip alone cannot establish raw rounding. Exact
character setup and expected numbers must be pinned before this becomes a
reproducible observation vector; no fabricated record is supplied here.

## Expression observation

SCX-O02 is pending. On the same identified vanilla profile, select one known
constant calculation cell and one parameterized cell whose effective table value
has first been established. In a matching original process, inspect the selected
compiled pool reference and complete expression bytes before requesting the
calculation. Record the exact skill context, declared remaining pool length,
function/parameter requests, evaluator return at Fog0x19E90, and any seed change.
Compare execution to SCX-R01 through SCX-R11. Separately compare compiler output
to SCX-R12's gated terminator/folding expectations; a compiler mismatch does not
permit changing evaluator expectations. Class/level/equipment are immaterial for
a constant cell; the parameterized cell's unit/skill state must be recorded.

## Review record

Regeneration 2026-10-15 preserves the existing file boundary and root rule numbers;
sub-ids separate evidence boundaries. The original format compatibility contract
remains explicitly in section 4.1. Original internal layouts and investigation
history were preserved privately; public source references and semantic mappings
are retained above. No implementation comparison, implementation plan, gameplay
test, or approval claim is part of this record. Self-review checked numeric examples, semantic inputs/outputs, dependency gates, and original-source mappings. All 23 original vector numbers and values are retained; the final draft has 36 executable vector ids and two pending observations. A fresh smaller
reader is coordinated by the parent task; completion is not claimed before it.

Mechanical review on 2026-10-15: the spec, companion, and gap file pass the
spec-writing linter with 0 errors and 0 warnings. Self-review also added the
zero-divisor continuation case in SCX-V02 so a premature return cannot pass that
assertion. This is documentation validation; no gameplay test was run. Fresh-reader
findings and any resulting changes remain to be recorded by the coordinating task.

### Shared getter reconciliation (2026-10-15)

The numeric-foundations companion's stat-reader records identify original D2Common
0x38A80 and 0x38B00 with the same fingerprint and inputs. Their recorded complete
entry-to-return evidence establishes selected-array and own-array reads respectively,
including the same conditional minimum adjustment. Reusing that compatible evidence
closes the leaf uncertainty for SCX-R18a's qualifiers and SCX-R31's cold path; no
repeated byte-run decoding is required solely to relabel the shared contract.
The 2026-10-12 SCX cold wording said only "directly from the unit stat list";
the SLD-R10 lead called it "base stat list". Original wrapper 0x4F870 calls the
selected-array leaf, so a forced-own-array interpretation would be incorrect.
SCX-V36 discriminates the source using different supplied own/resolved values.
The special to-hit producer remains SCX-Q02, since it includes additional arithmetic,
player eligibility, and class-record data beyond the common getter.

## Cap data

TD-CAP, 2026-10-15: the original game's loader returned bytes 0 through 4095
(4096 returned of total 7116) of effective `data/global/excel/experience.txt` on the
reported vanilla Lord of Destruction Expansion 1.13c profile. The full header and
the first data row were present. Header-based parsing selects row identity
`MaxLvl` (first data row; the `Level` column identifies the row) and column
`Amazon`, raw decimal cell `99`, interpreted as integer 99. This named text cell
is verified in 1.13c data. Coverage is sufficient for this cell only; the rest of
the file was not required and no absence claim is made.

The profile is the same reported vanilla edition/build family described in the
manifest's original-profile data limits. No new pristine-install proof, complete
override inventory, full effective-file hash, or loaded-DLL fingerprint set was
captured for this read; no physical supplying archive is claimed. It was an
effective-file read through the running game's loader, not a loose-file lookup
or a gameplay observation. Text-file evidence alone does not prove the loaded
compiled value.

The independent code binding is D2Common 0x60920: first descriptor string at
RVA 0x915B0 is Amazon, bound to record offset 0; subsequent class descriptors advance
by four bytes, and row stride is 32. The returned table pointer is stored at
data RVA 0xA0B50. Skill function 0x51AF0 and hard-level resolver 0x51070 dereference
that pointer's first dword as the clamp input. The string at 0x91558 names
experience. Thus the table row identity and the code's selected field have
independent evidence; SCX-Q03 remains the narrower loaded compiled-record
correspondence question. The synthetic clamp vectors do not require this live
correspondence and retain their explicit cap inputs.

### Additional-family targeted follow-up (2026-10-15)

A narrow original-image read of D2Common data RVA 0x9D354 through 0x9D377 resolves
the four bindings to function RVAs 0x23D60,0x23D50,0x267C0,0x27120, each with arity 2,
followed by count 4. A targeted direct-reference search for front end 0x23CE0 found
no direct caller in the inspected shared-module instruction records; that is not
an absence proof because indirect dispatch remains possible. These targets narrow
the next reading but do not establish the pool's owner or the functions' results.
SCX-R35b/SCX-Q06 remain unresolved.

### First full blind-reader triage (2026-10-15)

The full reader reported MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 17,
COVERAGE 20, and HARDEST 3. Its coverage list contains 17 grouped bullets despite
the reported count 20; the table below records every supplied group without
inventing three additional findings. No arithmetic mismatch or missing decision
was reported. This review adds hand-derived assertions from the existing contract;
it supplies no new original-source reading or gameplay observation, and changes
no rule label, producer default, or gate.

DISCRIM outcome: SCX-V11 originally checked only binding indexes 0 and 1 and an
invalid index. Direct-operation vectors SCX-V13/SCX-V14/SCX-V26/SCX-V27/SCX-V28
could still pass with wrong indexes 2 through 6. New SCX-V11(d–h) executes function
tokens for each of those indexes, asserting the selected semantic operation,
three forwarded argument slots, context, and distinct final evaluator result.
The supplied operations isolate this binding boundary; they do not invent random,
skill-entry, missile-resolver, or stat-result production.

| Coverage group | Outcome and reason |
|---|---|
| SCX-R03/SCX-R04: four-byte operands | SCX-V09(d) adds a signed four-byte parameter selector and exact call log. No literal change: SCX-V01(c) already executes a four-byte literal with result 256. |
| SCX-R05: `>=`, `==`, `!=` | SCX-V10(e–g) adds true and false results for all three opcodes. |
| SCX-R06/SCX-R08: multiply and INT_MIN negate wrap | SCX-V03(b) adds 65536 squared wrapping to zero; SCX-V08(d) adds INT_MIN negation retaining INT_MIN. |
| SCX-R10: arity one | SCX-V25(d) asserts argument slots (2,0,0), one call, and evaluator return 200. |
| SCX-R11: nonpositive compiler length and absent pool | SCX-V12(f–g) asserts the sentinel/no append at supplied lengths 0 and -1 and zero/no evaluation with an absent pool. Compiler output production remains SCX-Q01. |
| SCX-R15: null context, overflowing range, absent seed owner | SCX-V13(e–h) asserts null-context zero, wrapped-length no draw, absent-unit no-draw success, and the draw-needed fault. No arbitrary seed is supplied for an absent unit. |
| SCX-R16a/SCX-R26: non-own entry, clamps, missing unit | SCX-V14(d–g) adds omitted bonus, lower/upper clamps, and absent-unit forwarding/level-selector assertions. Entry/bonus/cap producers remain gated. |
| SCX-R18a/SCX-R19: missing context/unit, negative id, null nested unit | SCX-V27(e) adds routing guards; SCX-V28(b–c) distinguishes absent context from a present context containing a null unit. The latter still forwards both requests. |
| SCX-R21/SCX-R23: other linear pairs, nonpositive mana level | SCX-V17(d) adds zero/negative-level mana guards. No other-pair cases: SCX-R21 explicitly binds each pair to the same progression; SCX-V15 already distinguishes progression from raw/level selection. Repeated field bindings are not an exhaustive discriminator claim. |
| SCX-R25b: invalid skill and nonpositive level | SCX-V32(c) adds zero/no evaluation for invalid identity and levels 0 and -1. |
| SCX-R28: no passive match and nonpositive level | SCX-V33(b–c) adds zero/no evaluation for no matching slot and for levels 0 and -1 despite matching slots. |
| SCX-R30: minimum gate enabled only by first increment | SCX-V20(b) keeps the shifted minimum at 256 and changes only EMinLev1 to 1 at level 1; one synergy evaluation and result 384 distinguish the increment arm of the gate. |
| SCX-R31: lightning, cold selector 12, selector 5 | No change: the explicit fixed table sends 2 to ordinary stat 330, 12 to the same cold stat/path as 4, and 5 to ordinary stat 332. SCX-V21 exercises ordinary-source arithmetic and unsupported-selector rejection; SCX-V36 exercises cold selected-array routing. These remaining table bindings introduce no additional result-changing control or arithmetic branch, and neither vector claims exhaustive selector coverage. |
| SCX-R32a: nonzero source share and maximum endpoint | SCX-V22(b) supplies a negative source, asserting truncation toward zero and source inclusion before synergy/shift, with distinct minimum 2304 and maximum 4864. Actual source production remains SCX-Q07. |
| SCX-R33: first two duration bands | SCX-V23(c) adds levels 8,9,16 and results 120,125,160, covering both bands and their transition. |
| SCX-R34a: type rejection and owner-skill binding | SCX-V34(b–c) adds wrong-type rejection and bytecode index 3 selecting owner skill/level with supplied entry/bonus/cap. No complete resolver result is added: SCX-Q05/SCX-Q09 still gate their producers. |
| SCX-R35a: absent pool and invalid reference | SCX-V35(b) adds absent-pool, exclusive-upper-bound, and empty-sentinel rejection with no evaluation. |

HARDEST outcomes, all no change because no missing decision was identified:

- SCX-R30/SCX-R31 and SCX-V20–SCX-V22/SCX-V36: the existing rules explicitly
  order shift, minimum-only gate, synergy, and mastery, with NRF-R05's argument
  order. Existing exact results distinguish the arithmetic stages; the added
  increment-only case addresses the advisory branch without changing those rules.
- SCX-R18a/SCX-R31 and SCX-V27/SCX-V36: own-array versus selected-array reads
  consume NRF-R12b/NRF-R12c. SCX-V36's own 10 and selected 30 distinguish the
  cold/qualifier source. The special to-hit producer remains an explicit gate.
- SCX-R22/SCX-V16: signed divisions truncate toward zero, the upper clamp is
  explicit, and only the fourth pair omits the nonpositive-level guard. Existing
  cases give 110 at the upper clamp and distinguish fourth-pair 10 from first-pair
  zero at level 0; no additional implementation choice is needed.

GATED 17 comprises SCX-R12, SCX-R16b, SCX-R18b, SCX-R27, SCX-R32b,
SCX-R34b, SCX-R35b and SCX-Q01, SCX-Q02, SCX-Q03, SCX-Q05, SCX-Q06,
SCX-Q07, SCX-Q08, SCX-Q09, SCX-Q11, SCX-Q12. Each retains its existing
outcome and exact blocked boundary in the Id map and gap file. Outside-boundary
server-flag production and other builds do not block supplied-input consumer
assertions. Pending observations do not become static implementation gates.

Self-review preserves all 35 root rule ids and their sub-ids, all 36 executable
vector ids, and the 23 earlier numeric examples and seeds. New cases use explicit
inputs, no draw or a specified fault, exact returns, and call/seed effects at their
stated boundary. A fresh full reader is required for the changed vectors; this
triage does not claim that the resulting contract has completed that review.

### Shared named-cell corroboration, 2026-10-15

The coordinating reader collected the complete effective `data/global/excel/itemstatcost.txt`
through the owned game's Storm loader: offset 0, whole-file request, 40109/40109 bytes.
Header-based parsing selected explicit raw ID `332`, whose raw `Stat` cell is
`passive_pois_mastery`; its zero-based file data-row index excluding the header is 332.
The complete returned file SHA-256 is
`ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`.
The [shared collection record](combat-damage-pipeline-order.md#effective-file-named-cell-corroboration-2026-10-15)
states profile, version-console, override/archive and compiled-table limitations. This
corroborates the effective text name and replaces the unresolved-name wording only; it
changes no selector binding, integer result, behavioral label or integration gate.

### Fixture-input self-review before full reader 2, 2026-10-15

The coordinating review made fixture defaults explicit: unit absent unless supplied, and a
named supplied calculation result selects nonempty reference 0 with a logged evaluation
operation. This avoids inheriting the empty-reference default while expecting synergy.
V21 now explicitly supplies its present mastery unit U; V32(b) explicitly supplies U for
its asserted evaluation context. V22's elemental input names a supplied result rather than
an ambiguous reference/value. Existing numerical results, rule behavior, ownership and gates
are unchanged. These fixture-input clarifications are included in the forthcoming full
second reader pass, rather than relying on pass 1's earlier inputs.

## Fresh full blind-implementer pass 2, 2026-10-15

The fresh reader read the complete resulting contract, foundations and game-data contracts,
and recomputed all 36 vector families and their subcases. Counts: MISMATCH 0, GUESS 0,
GAP 0, DISCRIM 1, GATED 17, COVERAGE 17 grouped advisory branches, HARDEST 3.
This is a full review, not a delta check. No source or gameplay evidence is added by it.

DISCRIM V21: accepted. With damage 2560, synergy 50% and mastery 30%, both sequential
percentage orders happen to return 4992. The prior discriminator's 4608 rejects adding
base-only mastery after synergy, but does not reject moving the whole mastery stage
before synergy. V21(c) now uses damage 257, shift zero, synergy 50 and mastery 30:
synergy first gives 385, then mastery gives 500; reversed stages give 334, then 501.
The original 4992/3840 outcomes remain unchanged. No behavioral rule changes.

| Advisory branch group | Outcome |
|---|---|
| R03/R04 operand truncation | Added V09(e) for missing first operand and V09(f) for complete readable backing extending beyond declared length: the latter logs signed parameter -128 before returning 0 on run-off. Literal run-off already returns 0 in V06; inaccessible backing remains Q11, with no invented integer result |
| R06 INT_MIN/-1 | Added V03(c): complete bytecode reaches the unguarded signed division-overflow fault, unlike opcode division by zero's explicit zero result |
| R07 positive-exponent product wrap | Added V04(d): 65536 squared returns low-32-bit zero; the original positive/nonpositive exponent cases remain |
| R11 nonzero reference / secondary pool | Added V12(h): independent skill and secondary pools at reference 1 return 5/9 with remaining length 3 and the same supplied context. Selecting the wrong pool or starting at byte zero fails |
| R17 absent context | Added V26(b): return 0 with no resolver request, contrasted with the original supplied return 73 |
| R20 unavailable record / negative selector | No change: these select the same reject-with-zero/no-downstream-request recipe already exercised by V29's invalid identity and selector bound; distinct rejection predicates are stated, without claiming exhaustive predicate coverage |
| R22 negative / descending parameters / -6 fault | Added V16(f/g): the unguarded fourth pair returns -12 at level -1 without a lower clamp and faults at level -6. Descending endpoints use the same final upper-clamp recipe already shown by V16(c); no separate recipe is inferred |
| R23 non-8 shift / negative / overflow mana | Added V17(e): signed mana -1 and shift 7 produce -1/-6/-128; the intermediate signed division distinguishes floor rounding. Low-32-bit arithmetic and masked shifts consume the same existing numeric contract, already exercised by wrapping arithmetic and shift vectors; no additional overflow policy is added |
| R25a zero synergy / mastery with absent unit | V20(c) now logs both reached evaluations returning zero and unchanged endpoints 512. The absent-unit mastery guard is already exercised for this same endpoint by SLD-V14; SCX-V31 separately tests selector/request routing. No duplicated source getter behavior is introduced |
| R25b invalid nonempty reference / pool | Added V32(d): reject with zero/no evaluation, without falling back to the distinct linear result 18 |
| R28 selector pairs 0x18/0x19 | No change: the code-fixed pairs are stated. V33 discriminates earliest matching slot, either member of the pair, no match and nonpositive levels; the other fixed pairs repeat that selection recipe |
| R29 absent increment table | No change: SLD-V01 exercises absent-array return zero for the same shared progression primitive; SCX-V18 discriminates all nonempty band transitions |
| R30 zero synergy / D greater than 256 | Added V20(c) for zero-result evaluation and V21(c) for the positive D=257 minimum gate; original exact-threshold and nonzero-increment cases remain |
| R31 lightning / poison selectors | No change: SLD-V14 checks both numeric bindings for the same mastery helper, while SCX-V21/V36 cover ordinary and cold selected-array results. More fixed-binding cases would not add an unshown arithmetic/source recipe |
| R32a invalid skill / nonpositive level | No change: SLD-V09 and SLD-V02 exercise these original physical endpoint defaults and level-zero progression. SCX-V22 already distinguishes local physical rounding/source arithmetic. No nonpositive-level gate is inserted |
| R33 invalid skill / nonpositive level | No change: SLD-V07/V09 exercise this original duration producer's zero returns; SCX-V23 exercises its three bands and synergy. Repeating the same zero-result guard adds no different recipe |
| R34a absent-missile negative fallback | No change: Q11 explicitly retains the malformed fallback/fault boundary, and R34b is unresolved. No safe fallback result or new acceptance assertion is invented |

GATED outcomes (17 ids), all no change: Q01/Q02/Q03/Q05/Q06/Q07/Q09/Q11 and
R12/R16b/R18b/R27/R32b/R34b/R35b retain their exact producer, integration or malformed-input
limits. Q12 is observation-only and does not gate the verified supplied-input arithmetic.
The other-build Q08 is outside the 1.13c boundary, not another verified-profile dependency.
The reader found no new question. GUESS 0 and GAP 0 need no edit.

HARDEST outcomes (3): V21's percentage order is addressed by the new noncommuting-rounding
case; V36's cold source remains explicitly the NRF selected-array/minimum contract with
separate own/resolved values; V13 already separates seed owners, inclusive range and
no-draw/fault outcomes. The latter two need no change.

Self-review preserves all original numerical outcomes, rule labels, ownership and gates.
Changed cases are V03(c), V04(d), V09(e/f), V12(h), V16(f/g), V17(e), V20(c), V21(c),
V26(a/b) and V32(d), plus the corrected V21 discriminator. These are fixture/assertion edits,
so fresh full reader pass 3 is required before review completion. No gameplay or
implementation test ran.

## Fresh full blind-implementer pass 3, 2026-10-15

All 36 executable vector families and their subcases were recomputed against the resulting
contract and cited foundations/data contracts. Reported counts: MISMATCH 0, GUESS 0, GAP 1,
DISCRIM 3, GATED 18 (including an aggregate supplied-stage limit), COVERAGE 10, HARDEST 3.
This was a full review. The report is not source evidence or approval.

| Finding | Outcome |
|---|---|
| GAP: GDF-R08a omits some named local fields | Citation clarified. GDF-R08a does not list mana/lvlmana/manashift or every local passive/calc field. Section 3 already declares those semantic inputs and widths; this companion's original Skills descriptor record identifies their field correspondence. Section 3 and Dependencies now cite that existing local record and limit GDF-R08a to its listed subset. No field, width, rule ownership or supplied-input boundary changes. Missile-slot identity/result correspondence remains Q05 rather than becoming an accepted mapping |
| DISCRIM: V08 empty-pop final zero | Added V08(e/f): one-value addition returns 5, and empty addition/negation followed by literal 5 and addition also returns 5. Premature abort with zero fails; the original empty-operation returns remain unchanged |
| DISCRIM: V11(c) invalid-index pops | The existing Outputs already assert logical stack [3,9,0] before termination, so consuming arguments violates an explicit assertion. Added V11(i) as an end-to-end scalar check as well: two following additions return 12 with no function call; prematurely consumed arguments fail it |
| DISCRIM: V25(c) unsupported-arity pops | The existing Outputs already assert logical stack [2,0] and no call. Added V25(e) as an end-to-end scalar check: a following addition returns 2, rejecting a consumer that discarded the retained operand |

| Advisory COVERAGE group | Outcome |
|---|---|
| R21 remaining linear/raw selectors | No change: the fixed input-to-pair bindings share the progression/raw recipe already distinguished by V15; no additional arithmetic or control branch is claimed |
| R22 other diminishing pairs | No change: V16 now covers positive, capped, guarded-zero, unguarded-negative and fault recipes. Changing the fixed parameter pair does not introduce another recipe |
| R24 recursive columns besides `calc1` | No change: V30 checks forwarded recursive context; V12 checks reference/pool behavior. Other fixed column identities consume those same operations |
| R28 remaining selector pairs | No change: V33 discriminates first matching slot, either pair member and rejection. The fixed 0x18/0x19 pairs use that same search recipe |
| R30 invalid/nonpositive level / empty synergy | No change: the same original endpoint's SLD-V05/V09 check level/lookup rejection and unchanged base with no calculation. SCX-V20/V21 check the additional minimum gate and ordered percentages. No new rejection policy is introduced |
| R31 remaining selectors / unavailable unit/definition | No change: SLD-V14 checks fixed numeric bindings, missing-unit/list/definition guards for the same helper. SCX-V36 distinguishes its selected-array/minimum source. A further binding case adds no unshown source or arithmetic recipe |
| R32a invalid/nonpositive level / disabled source inclusion | No change: SLD-V09/V02/V04 exercise these same physical endpoints' sentinels, level-zero baseline and both source gates. SCX-V22 independently checks this consumer's source/rounding order |
| R33 guards / negative percentage | No change: SLD-V07/V09 cover this duration producer's zero-result guards. A negative nonzero percentage follows the same signed mulDiv/add branch as V23(b), without a separate clamp; the numeric contract already specifies its rounding and width |
| R34a min/max bindings / pool rejection | No change: the named min/max operations consume R14, distinguished by V11(a/b); R11/V12 own the common pool rejection recipe. V34 separately checks missile context and owner-skill selection. This does not claim exhaustive testing of every family/binding combination |
| R15 remaining absent-unit fault | No change: the reader itself identifies this as already covered by V13(h); no uncovered branch or missing default was reported |

GATED outcomes: all existing rule/question gates remain. The reader's aggregate stage-only
entry means supplied fixtures do not verify upstream producers or runtime behavior; it is
not a new acceptance restriction. Q08 remains outside this build. Pass 2 reported GATED 17
while its explicit in-scope id list contains 16 distinct ids and separately mentions Q08;
this counting discrepancy does not identify an additional gate or silently close one.
HARDEST outcomes (3), no rule change: R30/R31's minimum gate and ordered percentages are
now distinguished by V20/V21(c); R15's range/seed/fault cases are explicit in V13;
R22's signed truncation, upper-only clamp and fourth-pair negative/fault cases are explicit
in V16. The reader supplied no additional unresolved decision for these entries.

Only fixture cases/assertions V08(e/f), V11(i), V25(e) and the reference clarification
changed after this full pass. All original outcomes and rule behavior, ownership, labels
and gates remain. With MISMATCH 0 and every finding adjudicated, the next applicable
fresh run is a delta check under the fourth-run rule. It must confirm these additions
and the citation resolution before review completion. No gameplay or implementation test ran.

## Fresh delta blind-implementer pass 4, 2026-10-15

Changed material: V08(e/f), V11(i), V25(e), and the section-3/Dependencies citation
clarification for locally declared fields versus GDF-R08a's listed subset. The fresh
reader checked these against the current contract, cited foundations/data contracts
and existing companion mapping record. MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0.
No further contract edit follows this delta report. This was a changed-material check,
not a fourth full review or a new source reading.

The applicable reader review is complete at the declared supplied-input/static
boundary. All existing producer, integration, malformed-input and observation limits
remain; status stays draft and no human approval or complete parity is inferred.
