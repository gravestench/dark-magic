# Evidence: movement modifiers and freeze

Spec: [Movement modifiers and freeze](../../specs/state-movement-modifiers-and-freeze.md).
Artifacts: [original source manifest](../source-manifest.md).
Remaining limits: [evidence gaps](../../plans/gaps/state-movement-modifiers-and-freeze.md).

## Identity and reuse

Evidence profile: Lord of Destruction Expansion 1.13c, Windows x86, little-endian. Artifact ids
`LOD113C-D2COMMON` and `LOD113C-D2GAME` resolve to the manifest's fingerprints. Available original
files were rehashed on 2026-10-15 and matched those entries. This establishes specimen correspondence,
not gameplay behavior. Function descriptions below are analyst descriptions, not exported names.

The 2026-10-12 inline instruction readings are retained for their recorded slices. Their source
record included original-instruction reading and independent original listing checks of encoded
instruction regions in the rate dispatcher and freeze applier. Earlier draft read dates were not
recorded; no new date is assigned to them. The old blanket assertion that no checked reconstruction
existed was inconsistent with its own cold-callback record; the callback's checked reconstruction
record is preserved. Private source history, exact original documents, layouts and collection
commands are preserved in a separately hashed audit outside the repository.

Before new collection, the current NRF, GDF, CSA, STL, HRC, DMI, DPO and monster-scaling records
were compared for the same original artifacts, inputs and outputs. Same-function coverage is reused
with its dates: CSA cold/freeze and helper readings 2026-10-13/14, NRF helper/direction readings
2026-10-12 through 2026-10-14, HRC shared rate readings 2026-10-12 with targeted corroboration
2026-10-15, and monster initialization's full independent listing 2026-10-14. Their actual local
boundaries are preserved; none establishes an unread propagation or animation-preparation path.
Data profile and override limits remain those of the linked GDF/DMI named-cell records.

New 2026-10-15 reading was restricted to unresolved input/consumer boundaries and correction
candidates. Checked C was available for Common 0x50CE0 and was read before deriving its partial
producer behavior. Other newly read functions in the table below lacked that view and were read
in disassembly. Common 0x230A0/0x719A0 do have checked reconstructions but were not read here;
their item-type production stays gated. No game action, table collection or gameplay observation
was performed. Relocation and formatting alone do not promote any label.

## Dispatch and function references

The [CSA dispatch chain](combat-status-application.md#dispatch-chain), retaining its 2026-10-13/14
readings, anchors status entry: skill do-function table at Game RVA 0x1074A8, selected index 1,
leads through 0xA2C40 to build/enqueue 0xDDDE0 and dequeue 0xDED30 (also 0xA20C0); the missile
hit-function table at Game RVA 0x10DA48 reaches 0x3AE10 and callback 0xA10B0. Dequeue/contact/
callback routes reach hit application 0xDE0C0. That function calls cold 0xDC780 with damage-record
cold length, then freeze 0xDDAC0 with freeze length. Freeze fallback calls 0xDC780. The appliers
install callbacks Game 0xDB4C0 and 0xDBD90; actual expiry/removal-to-callback integration consumes
STL-R03/STL-R05 and remains gated.

New cold list creation and cold expiry call the imported Common #10819 at RVA 0x33110. The
[NRF animation-preparation record](core-numeric-and-rng-foundations.md#fixed-point-time-movement-and-object-animation)
also identifies Common 0x32220/0x501E0; their complete remapping/preparation remains NRF-Q17 and
MMF-Q11. The dispatcher calls 0x501E0, then animation preparation 0x32670, then predicates and
rate branches. The local consumer boundary starts after supplied effective inputs/preparation;
a call chain alone does not verify their effects.

Creation is reused from [monster scaling](monster-scaling-by-players-and-difficulty.md#creation-chain):
Game 0xE1D90 -> 0xE0110 seeding, type-1 table -> 0xB0320 -> preset lookup 0xCA5F0 -> 0xAFDB0.
The reached fixed-stat slice is after state-list/stat setup and before later life, property and
equipment setup. Player initialization's reached local fixed-stat slice is Game 0xEB2C0; complete
player creation is not newly claimed by its three requests.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2COMMON | 0x33110 | 10819 | Ordered rate dispatcher | MMF-R02/MMF-R03/MMF-R04a/MMF-R05a/MMF-R05b/MMF-R06a/MMF-R06f | 2026-10-12 instruction reading retained; 2026-10-15 complete applicable consumer branch reading and independent listing corroboration of encoded head; preparation/predicate producers gated |
| LOD113C-D2COMMON | 0x323E0 | none | Diminishing stat selection | MMF-R01 | NRF-R07, compatible original reading retained |
| LOD113C-D2COMMON | 0x30D50 / 0x2FB10 | none / none | Base velocity / movement animation base | MMF-R03/MMF-R04a | 2026-10-12 retained; complete applicable disassembly 2026-10-15 confirms signed monster word, player byte and unguarded invalid record reads |
| LOD113C-D2COMMON | 0x34D40 | 10488 | Path-speed/target/reset writer | MMF-R03 | NRF-R08b checked reconstruction, retained; ordinal resolved from original export table 2026-10-15 |
| LOD113C-D2COMMON | 0xCEB0 / 0xD240 / 0xCF40 | none | Step/ramp and direction production | MMF-R06d/MMF-R06e | NRF-R11a/R11b records: direction constant-table read 2026-10-14; full path production not inferred |
| LOD113C-D2COMMON | 0x30B80 / 0x2EF90 / 0x2EFC0 / 0x2F010 / 0x2FAC0 | none | Cast/block/get-hit/knockback/base-only predicates | MMF-R06f | 2026-10-12 retained; HRC-R12a/R14 supplies three rate branches; skill/sequence integration gated |
| LOD113C-D2COMMON | 0x30C70 / 0x30BD0 | none | Movement / attack predicates | MMF-R06g | Complete disassembly 2026-10-15 identifies table and selected-skill branches; dependent skill production not all read |
| LOD113C-D2COMMON | 0x501E0 / 0x32670 / 0x32EF0 | none / 11122 / 10932 | Effective-input preparation / animation preparation / attack preparation | MMF-R06g/MMF-R05a | NRF/HRC reuse and caller inspection; actual effects remain MMF-Q08/MMF-Q11 |
| LOD113C-D2COMMON | 0x30530 | 10747 | Eligibility by kind/class | MMF-R05b | Complete disassembly 2026-10-15; entry-to-return class comparisons, no callee |
| LOD113C-D2COMMON | 0x1E1B0 / 0x23EF0 / 0x24430 | none / 11006 / 10744 | Inventory grid / acceptance / type relationship | MMF-Q08 | Complete relevant disassembly 2026-10-15; encoded regions corroborated for latter two; allocator/type-provider callees remain integration gaps |
| LOD113C-D2COMMON | 0x33C20 / 0x50CE0 | none / 10639 | Active-state mask intersection / player base override | MMF-R05a/MMF-Q08 | First's role reused from NRF/PVP; original listing corroboration 2026-10-15; second checked C read, selected-weapon/calculation callees unverified here |
| LOD113C-D2GAME | 0xEB2C0 | none | Player initialization fixed-stat requests | MMF-R06b | 2026-10-12 retained; narrow ordered request slice checked 2026-10-15 |
| LOD113C-D2GAME | 0xAFDB0 | none | Monster initialization fixed-stat requests | MMF-R06c | MSD full original listing 2026-10-14 retained; entry-to-fixed-request slice assessed 2026-10-15 |
| LOD113C-D2COMMON | 0x3A740 | 10887 | Unit stat-write request entry | MMF-R06b/MMF-R06c | Game thunk 0xA472 resolved 2026-10-15; forwards replacing write to 0x3A280; actual propagation is STL-R06c |
| LOD113C-D2GAME | 0xDE0C0 / 0xDC780 / 0xDDAC0 / 0xDAE20 | none | Applied-hit handoff / cold / freeze / cold effect | MMF-R07/MMF-R08a/MMF-R08b/MMF-R08d/MMF-R11/MMF-R12/MMF-R13a/MMF-R13d/MMF-R14a | CSA-R01/R05-R09 original readings 2026-10-13/14 retained |
| LOD113C-D2GAME | 0xDA7B0 / 0xDC0B0 / 0xDB3C0 / 0xDBD00 | none | Length strip / mitigation / resistance / entry arithmetic | MMF-R09/MMF-R10a/MMF-R10b/MMF-R10c | DPO-R10 and DMI-R01/R02/R05-R09 current records; not reread or expanded here |
| LOD113C-D2GAME | 0xDB4C0 | none | Cold callback | MMF-R13c | Checked reconstruction, CSA-R13 dated 2026-10-14; callback-local only |
| LOD113C-D2GAME | 0xDBD90 | none | Freeze callback | MMF-R13c | 2026-10-12 retained; complete disassembly and complete independent original listing 2026-10-15; cleanup call/effects separated |
| LOD113C-D2GAME | 0xDA750 | none | Associated-record cleanup | MMF-Q12 | Complete disassembly assessed 2026-10-15; encoded regions/free integration not verified, no stored cleanup claim |
| LOD113C-D2GAME | 0x13B0 / 0x22A80 | none | Extended-monster flag / condition-word intersection | MMF-R08c/MMF-R14a | CSA-R02/R06 retained 2026-10-14; GDF-R06a identifies deadCol |
| LOD113C-D2COMMON | 0x1A550 / 0x1A2D0 | 10064 / 11104 | Boss / fixed hireling category | MMF-R08c | CSA-R02 original full disassembly 2026-10-14 retained |
| LOD113C-D2COMMON | 0x33CD0 / 0x34510 / 0x37DB0 | 10494 / 10945 / none | State test / state request / membership leaf | MMF-R08c/MMF-R13a/MMF-R14a | NRF-R12a/STL-R07; GDF-R11b selects state 101 for block |
| LOD113C-D2COMMON | 0x603C0 | 10237 | Difficulty row getter | MMF-R11/MMF-R12 | CSA checked reconstruction; domain 0..2 |
| LOD113C-D2COMMON | 0x38520 / 0x39140 / 0x38B70 | 10871 / 10680 / 10973 | State-list / own-list / unit selected-array lookup | MMF-R13d/MMF-R13b | NRF-R12a/R12b/R12c retained; not contribution mutation |
| LOD113C-D2COMMON | 0x38100 / 0x38060 / 0x37D70 | 11013 / 10785 / 10911 | New list / state / callback | MMF-R13a | STL-R06a retained checked reconstruction |
| LOD113C-D2COMMON | 0x38000 / 0x38020 | 10412 / 10161 | Expiry read/write | MMF-R13d | STL-R06a retained checked reconstruction |
| LOD113C-D2COMMON | 0x3A7D0 / 0x3A280 / 0x3A7B0 | 10807 / 10261 / 10188 | Attach / replacing stat writes | MMF-R13b/MMF-R13e | STL-R06b/R06c gated execution, CSA caller requests established |
| LOD113C-D2COMMON | 0x33AB0 / 0x33920 | 11061 / 11090 | Stay-death / mode write | MMF-R13c | STL-R12a/NRF-R14a retained local consumers |
| LOD113C-D2GAME | 0x8EA80 / 0x8E6F0 | none | Schedule / timer-node removal | MMF-R13c/MMF-R13e | STL-R01 / CSA-Q02 limits retained |

Resolved Game import thunks are retained in the [CSA table](combat-status-application.md#function-references):
0xA3CA -> Common #10945/0x34510; 0xA68E -> #10494/0x33CD0; 0xA5D4 -> #10871/0x38520;
0xA6A0 -> #11013/0x38100; 0xA5B6 -> #10785/0x38060; 0xA75A -> #10911/0x37D70;
0xA60A -> #10807/0x3A7D0; 0xA76C -> #10261/0x3A280; 0xA790 -> #10412/0x38000;
0xA778 -> #10161/0x38020; 0xA7BA -> #10237/0x603C0; 0xA580 -> #10064/0x1A550;
0xA7F6 -> #11104/0x1A2D0; 0xA59E -> #10819/0x33110; 0xA7E4 -> #11061/0x33AB0;
0xA42A -> #10973/0x38B70. Freeze callback 0xA4A2 -> #11090/0x33920. Initialization uses
0xA472 -> #10887/0x3A740. These are original module-relative thunks, not function identities
inferred from analyst names. Original export-table assessment 2026-10-15 confirms additional ordinals above.

## Semantic-to-original mapping

These locations identify evidence, not required layouts or calling conventions.

| Semantic term | Original 1.13c location |
|---|---|
| unit kind/class/id/mode | Unit +0x00/+0x04/+0x0C/+0x10 |
| seed {lo,hi} | Unit +0x20/+0x24 |
| path / animationRate / actionRate / selected animation record | Unit +0x2C / word+0x4C / dword+0x3C / +0x50, selected rate at +0x0C |
| stat list / inventory / unit game / event chain | Unit +0x5C / +0x60 / +0x80 / +0xDC |
| path speed/targetSpeed/resetMarker | Path +0x7C/+0x84/+0x38; step/ramp fields follow NRF companion |
| game frame/difficulty | Game +0xA8/+0x6D |
| modeConditionA/B | Game +0x6A byte nonzero / +0x74 dword nonzero; wider meanings not inferred |
| aliveExclusionCondition | Unit +0xC6 bit 0 |
| coldLength/freezeLength | DamageRecord +0x30/+0x34 |
| WalkVelocity/RunVelocity | CharStats byte +0x40/+0x41 |
| Velocity and derived ordinary/run animation bases | MonStats signed words +0x32/+0x36/+0x38 |
| cold effect/delay | MonStats signed byte+0x168+difficulty / unsigned byte+0x4F+selected difficulty |
| boss / linked MonStats2 / deadCol | MonStats flag bit 6 of+0x0C / signed word+0x18; MonStats2 flag bit 19 of+0x04 |
| freeze/cold divisor | DifficultyLevels +0x14/+0x18, record stride 0x58 |
| list expiry/state/callback | StatList +0x18/+0x14/+0x24 |
| monsterFlag8 | MonsterData word+0x16 AND 8 |
| timer node kind/successor | byte+0x00 / +0x24 |
| attack pair | grid 0 from inventory helper 0x1E1B0, entry storage+0x0C, item fields+0x10/+0x14; no body-slot names inferred |
| override predicate / player condition / numerator | active-state words intersect DataTables mask+0x164; Unit +0xC8 bit 3; numerator Unit +0x48 AND 0xFFFFFF00 |

Cold's native attacker/defender roles are eax/esi with length on stack; freeze uses ecx/eax with
length on stack. The callbacks use ecx unit/edx state, with removed-list argument on stack.
The semantic spec uses roles and explicit fields; private original fixtures preserve their
binary encodings. The divisor is not a damage-record field.

## Rule rationale and corrections

### Movement and action consumer boundary

The 2026-10-12 reading already established first-match order, ordinary numerical families and
signed movement versus unsigned action division. The 2026-10-15 original instruction assessment
confirms Common 0x33110's movement branch at 0x3331B: path is tested before reading stat 67 or
writing the rate. With a path, add D(4) and floor 25, multiply the movement animation base by that
same percentage, signed divide 100/clamp, write Unit+0x4C, then multiply the shifted velocity
base, signed divide 100, call Common 0x34D40. Low-half `imul` occurs before division, so the
main contract now specifies exact wrapping products and includes a discriminating overflow case.
The invalid class path in 0x30D50/0x2FB10 reads through a null record; it is not a defined zero base.
These are source-supported clarifications/corrections, not new runtime measurements.

Action branch at Common 0x333DE first requests 0x32EF0(unit,1), then reads stat 68 and D(0). Its
paired-item condition calls 0x30530. With the supplied accepted items, both 0x24430(item,45) tests
must pass; it adds low32(S1+S2), signed-corrects halving toward zero, subtracts S1 and adds that
to the running rate. Actual player/mode 18 subtracts 30 before clamp 15..175. A nonzero optional
base override replaces B, then low32 multiply, unsigned /100 and cap 32767 write both rate fields.
The new numeric pair cases exercise this consumer independently; no unread item lookup is promoted.

Common 0x30530's complete no-callee path establishes player classes 4/6 and valid monster classes
417/418. Complete 0x1E1B0 inspection reveals allocation/reallocation rather than proving named hand
slots; 0x23EF0 tests item condition bits 0x100/0x4000 ; 0x24430 consumes a prepared item-type matrix,
then a secondary type from the item record. Their provider functions are not all read, so no
semantic meaning for type 45 or the pair slots is invented. The checked reconstruction of 0x50CE0
establishes only a player with condition bit 3 can yield an override, using selected weapon and
helper 0x267F0 or default divisor 19, applied to a masked numerator. The helper chain remains gated.

Movement/attack predicates 0x30C70/0x30BD0 were read in full. They consume per-mode engine tables
and sometimes selected-skill properties; these are not equivalent to an unconditional list of
mode numbers. This prompted explicit supplied predicates and the MMF-Q11 integration gate.
Common 0x33110's encoded head was corroborated in the original independent listing; later newly
used consumer branches are ordinary disassembly instructions. Earlier complete instruction-region
corroboration dated 2026-10-12 remains retained, not relabeled as a new full producer reading.

### Fixed initialization requests

The earlier text called monster baselines class-dependent and unread. Matching same-function
coverage in the monster-scaling companion identifies creation and provides a complete original
listing dated 2026-10-14. Targeted 2026-10-15 entry-to-fixed-request assessment at Game 0xAFDB0
finds consecutive layer 0 calls to Game thunk 0xA472: stat 68 value 100, stat 67 value 75, stat 69 value 100.
The primitive resolves to Common #10887/0x3A740. Player 0xEB2C0 requests the same order, all 100.
The initial all 100 player numeric family is retained; its order is made explicit. Monster 75
is a source-supported correction; later equipment/property writes are not ruled out, and these
requests do not establish original stored totals. No additional RNG is in either local fixed-write slice.

### Cold, freeze and lifecycle

CSA-R02/R05-R09/R13 settles the four freeze checks, invalid-record ColdEffect fallback, deadCol
suppression and all local application/refresh/draw requests. The old conditional possibility that
check 1 was a state test is removed: it is Common #10494/0x33CD0 for state 54. The other checks are
boss #10064/0x1A550, Game 0x22A80 mask 8, and hireling category #11104/0x1A2D0. No evidence identifies
flag 8 with a broader champion/unique label. Game 0x13B0(class,19) selects MonStats2 through
MonStatsEx and tests deadCol, closing its local identity question. Corpse behavior stays outside.

The original text overstated a globally unique state list, summed contributions, unchanged E
lookup and realized lifecycle effects. Current NRF-R12a finds the first matching primary list,
then secondary; CSA-R05 still looks up E on each eligible application, but never rewrites an
existing cold list's stats. Current NRF-R12b/R12c selects existing arrays; it does not sum or
mutate them. STL-R05/R06b/R06c leaves actual propagation/removal unresolved. The spec therefore
preserves verified requests separately from supplied fixture storage and actual lifecycle effects.

A complete independent original listing of Game 0xDBD90 corroborated its encoded regions on
2026-10-15: dead/stay-death guard, state-off request, null-unit fatal guard, call 0xDA750, imported
mode 1 write, then guarded monster game/data/record delay selection and kind 2 removal/scheduling
requests. The Game 0xDA750 disassembly reveals linked associated-record removal/free calls but
encoded-region and free effects were not fully verified here. Its effects remain MMF-Q12 rather
than an unqualified cleanup claim. The callback-local no-direct-draw statement excludes dependency
effects. Normal aidel is selected only when both neutral game conditions are false; otherwise
selected difficulty is used. This preserves the known branch without guessing names for game fields.

Cold zero divisor skips division (CSA-R05); freeze zero divisor is fatal (CSA-R08). This agrees
with the old detailed rules and fixes the old blanket uncertainty in the ledger, rather than
changing the retained cold/freeze numeric vectors. Freeze duration can be 0, list expiry=current
frame, but STL-R01 queues that request next frame. The new assertion distinguishes the list value
from scheduled execution.

### Resistance and absorb

The old resistance-helper fragment was asm-only, used player-only classic penalties and did not
identify length absorb entries. Current DMI-R01/R05/R07 establishes regular-monster/hireling
classification, immunity/pierce exception, Expansion penalties from data, Classic 20/50 penalties,
floor/caps and positive bypass. DMI's code-defined entries 5/6 have reduction slot 0 and no absorb
bindings, so length absorption does not run. DMI-R06/R09 consumes NRF-R05 `mulDiv(100-r,len,100)`;
it is not universally equivalent to multiplying in an unbounded integer or always low32 before
division. These same-artifact current dependency records supersede the partial helper fragment.
All retained small length outputs remain identical; MMF-V26 separates the corrected eligibility/
penalty behavior, using already recorded named difficulty cells. No new DMI source reading was claimed.

## Evidence-gap assessment

Every topic from the ten original questions and every gated leaf was assessed before rewriting.
This table records dispositions rather than a new acceptance contract.

| Topic | Assessment / targeted work | Current disposition |
|---|---|---|
| Cold/freeze length production | DPO-R06b and DPO-R07 now identify cold stat and conversion producers; complete arbitrary skill/missile producers not inferred | MMF-Q01 outside-boundary, none locally for supplied lengths |
| Four freeze checks | Same-function CSA-R02/R07 fully resolves local identities and precedence | Closed locally; MMF-R08b/MMF-R08c; flag 8 wider classification stays neutral |
| Divisor/ColdEffect/aidel data | Reuse GDF named Normal/Nightmare/Hell divisor cells; no arbitrary per-class cold/delay cell read | Divisors settled, MMF-Q03 supplies precise actual-cell limit |
| Monster baseline stats | Reused MSD root/full listing and targeted fixed-write assessment, original thunk resolution | Fixed requests settled by MMF-R06c; MMF-Q04 is later outside producer coverage |
| Run baseline/path/direction | Common movement consumer checked; NRF-R11a/R11b supplies complete isolated step and scale | MMF-R06d/MMF-R06e verified supplied boundary; full creator/caller path remains MMF-Q05 |
| Stat sums/overflow | NRF selected-array contract contradicts summation inference; caller low32 arithmetic traced | Arithmetic explicit; stored contribution integration MMF-Q06, not guessed |
| Shatter meaning/names | GDF names, CSA deadCol helper and draw establish local requests; no death-path sweep authorized | MMF-Q07 outside-boundary, no local draw gate |
| Pair slots/type/override | Complete caller arithmetic/helper conditions read; inventory allocator and item-type/override providers identified but not completely verified | MMF-R05b supplied arithmetic verified, actual production/preparation MMF-Q08 |
| Division-by-zero observation | Static NRF helper already establishes fault; no gameplay in migration | MMF-Q09 observation only |
| Class flag 19 record/column | Same-function CSA-R06 and GDF-R06a establishes MonStats2.deadCol | Closed locally; corpse consequence remains separate |
| Derived monster animation bases | Current GDF-R05b already records exact producer dependency and limits | GDF-Q01 integration remains; supplied signed words usable |
| Unread dispatcher/predicate callees | Complete relevant movement/attack predicates read; selected skill, remap and preparation not all verified | MMF-R06g/MMF-Q11 explicit production gate |
| Callback lifecycle/cleanup | Reused CSA cold callback and full freeze callback listing; cleanup dependency assessed | MMF-R13e/MMF-Q12; local callback log available independently |
| Verified rules without executable vectors | Retain 22 families, concretize symbolic ones, add paired/step/fallback/resistance/callback families | MMF-V01 through MMF-V28; all unobserved |
| Pending original observations | Four procedures retained and expanded to cover current verified branches | MMF-Q09; no runtime promotion |

Historical absence claims had no retained six-rung search record. They are not renewed as
`not found in d2-decomp`. Bounded positive source/dependency assessments above replace those
claims; the exact historical prose remains in the private preserved original. Further source
collection must target the named gate and complete its applicable boundary, not repeat old searches
because of their age.

## Data evidence reuse

No new data-only rule is introduced. [GDF named-cell records](core-game-data-fields.md#named-cell-records)
provide the effective file/profile identity, header/row convention, coverage and raw cells for
DifficultyLevels Normal/Nightmare/Hell cold/freeze divisors 1/2/4. GDF stat/state/mode records supply
the names/identities; its column-loader records establish field mapping. [DMI table records](combat-damage-mitigation.md#table-data-records)
provide named Expansion ResistPenalty cells 0/-40/-100 and compiled correspondence limits. Their
original dates and loader/profile/override limitations are retained. No file hash or physical MPQ
origin is invented here, and no supplied W/E/aidel/base fixture is labeled a shipped class cell.

## Observation procedures

Every procedure is pending. Use an owned vanilla Expansion 1.13c game whose artifacts and effective
data match the intended profile, and independently read actual stats/table inputs. Observe original
entry and output locations from the function table. Record build/data identity, class/level,
difficulty, scene, inputs, requests versus actual stored effects, final outputs and seed sequence.
These procedures describe future reproduction, not permission for actions or claims of completion.

<a id="mmf-o01"></a>
### MMF-O01

Use a disposable level 20 player, initially unequipped, on a clear flat outdoor area. Read its
class WalkVelocity and actual speed stats; choose controlled equipment/skill effects to vary item
FRW and additive movement independently. Compare Walk/Run/Town Walk only in appropriate accessible
areas, recording selected effective mode/base, movement percentage, path speed/step and stored rate.
Capture newly created player/monster fixed-stat requests separately from later modified totals.
For attack/default/cast/block/get-hit/knockback, record the actual selected branch and rate inputs;
compare the exact vector arithmetic, including caps and zero minima. Paired-item/override cases
require actual slot/type/preparation inputs before they can confirm that integration. Compare axis
and diagonal step against their exact supplied direction components. Terrain, class, level and
rate inputs stay fixed within each comparison. Extreme negative or overflow cases require a
separate authorized disposable setup; ordinary equipment cannot be assumed to reach them.

<a id="mmf-o02"></a>
### MMF-O02

Use a disposable level 20 player and a controlled cold attack in a clear outdoor area; record the
attacker/defender class and both post-build lengths. Repeat on each selected actual monster class
and difficulty with its read cold effect, divisors, boss/deadCol, flag 8 and category. Compare neither
item defense, Half Freeze and Cannot Be Frozen while controlling resistance/pierce. Record strip,
effective resistance, adjusted duration, selected list expiry and actual scheduler frame. Reapply
before expiry with shorter/equal/longer candidate ends and differing E; distinguish own-list values,
unit resolved totals and each AI request. At expiry record callback entry, removal-before-callback,
state/mode/AI requests and actual stored/restored effects separately. Select cases where delay 0 and
Normal/difficulty column differ, recording both game conditions. Dead/stay-death cases require a
separate controlled scene and do not establish mask producers from supplied masks alone.

<a id="mmf-o03"></a>
### MMF-O03

On a disposable level 20 player in a clear outdoor area, hold class, level, terrain and item speed
stats fixed while comparing unchilled and chilled movement/attack/default action inputs. Record
actual resolved 67/68/69 and rate/path outputs, rather than infer aggregation from visible speed.
For selected cold-only and player cold-plus-freeze hits, capture defender seed before each cold
call, each new pair, modulo 100 and requested state 107. Compare monster deadCol true/false with
actual read record identity; a shatter request does not confirm later corpse behavior. Keep
incoming cold/freeze lengths distinct so the two-call order can be measured.

<a id="mmf-o04"></a>
### MMF-O04

Use an isolated disposable original-game session with a separately authorized means of reaching
FRW=-150. Read the actual selected stat result immediately before the diminishing helper, then
record whether division faults before rate/speed writes. Keep ordinary player/class/path inputs
valid; a null or missing-definition early result is a different case. Save the fault location and
last valid outputs. Do not infer runtime confirmation from the static zero denominator or force
this setup in a session used for other play.

## Migration record - 2026-10-15

The existing bounded movement/action/cold/freeze contract was rewritten to the semantic template.
Ids were found unconsumed by implementation plans/code/tests; practical ids were retained and aliases
removed. All 22 vector families and 4 observation families were retained. Symbolic player W uses
supplied 6, attack base supplied 256, Half Freeze divisor supplied 4; these are fixture concretizations,
not changed formulas or claims of per-class archive values. All existing small numeric results are
preserved. MMF-V21's player two-call case now records its exact second seed and two off requests;
this completes its original draw-count expectation. Original binary fixture encodings and private
history were preserved before editing. The ten implementation-comparison findings remain historical
and were not rerun.

Source-supported corrections and narrowed boundaries are detailed above: monster fixed requests,
player request order, movement null-path/overflow, resistance/pierce/Expansion policy, no length
absorb, first-list lookup, E lookup versus stored list values, and requests versus realized effects.
Labels for MMF-R05b/MMF-R06c/MMF-R06d/MMF-R06e/MMF-R08c/MMF-R10b/MMF-R10c now reflect the named verified
supplied slices. MMF-R13b is unresolved actual propagation; MMF-R06g/MMF-R13e expose unread integration
rather than imply full execution. The document remains draft; no observation label was promoted.

Mechanical lint (`.agents/skills/spec-writing/tools/check_spec.py state-movement-modifiers-and-freeze`)
completed on 2026-10-15 with 0 errors and 0 warnings. Self-review checked semantic inputs/outputs,
all 22 retained numeric families, added cases, request/storage boundaries, public references,
all ten historical comparison findings and the private preserved originals. No implementation tests
or gameplay checks were run. Fresh reader outcomes are recorded below; a new full review is
required for the changes following FULL1. No historical review is presented as review of this
rewritten contract. Remaining gaps are in the
[gap file](../../plans/gaps/state-movement-modifiers-and-freeze.md#open-evidence-gaps).

## Fresh blind review FULL1 - 2026-10-15

A fresh reader examined the complete contract and cited dependency contracts, recomputing the
reported 27 vector families and 71 cases without implementation, private source, gameplay or web
access. Counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 1, GATED 16, COVERAGE 14, HARDEST 3. The report
was preserved privately as a transcription before fixes. These findings do not approve the spec.

| Finding | Disposition |
|---|---|
| GAP 1: standalone diminishing and eligibility returns missing from outputs | Added MMF-R01 signed return/fault and MMF-R05b boolean return to section 5; no additional write or gameplay effect claimed |
| DISCRIM 1: MMF-V12(a) lacks a ratio separating fixed and derived stat 67 | Added supplied WalkVelocity 6/RunVelocity 9; ratio gives 150 while the asserted request remains 100; no internal data-read count asserted |
| GATED 16 | Retained MMF-R04b/MMF-R06g/MMF-R13b/MMF-R13e/MMF-R14b, the ten current questions, and the supplied/isolated execution boundaries; no guessed integration added |
| HARDEST 1: MMF-R13c/MMF-V19 callback/timer order | No change: MMF-V19 explicitly separates application and supplied callback logs, normal-delay zero-to-15 selection and 1021/1035 requests. CSA-V11(a) is the owning saved-successor discriminator: removal clears a successor, so the required later kind-2 removal rejects rereading that successor after removal. Actual removal/callback integration remains gated |
| HARDEST 2: MMF-R13d/MMF-V20 retained expiry/new AI/cold draw | No change: MMF-V20(a) asserts the unchanged-expiry cold draw, and (c) asserts retained list end 1050 separately from new AI request 1021. Equality is discriminated in owning CSA-V08(d)/CSA-V11(c); no new propagation assumption |
| HARDEST 3: MMF-R14a/MMF-V21 seeds/two calls | No change: MMF-V21(c) gives both ordered draws, exact second seed and two player state-off requests; its persisted supplied list context makes the second equal-expiry call independently executable |

Coverage entries were advisory, and each has a separate disposition. Dependency vectors below
are read in the owning [CSA](../../specs/combat-status-application.md) or
[DMI](../../specs/combat-damage-mitigation.md) contract; their bindings and operations are not
copied into a parallel authority here.

| Coverage finding | Disposition |
|---|---|
| 1. MMF-R02 upper percentage/no ceiling | No change: existing MMF-V09(d) uses pct 16777216 with exact wrapped outputs; a ceiling of 175 would produce speed 2688 instead of 0 and fails that vector |
| 2. MMF-R03 monster velocity/invalid record | Added MMF-V09(e)/(f) using monster Velocity 8 and exact speed 3072, plus isolated invalid-record fault case (g) rejecting zero fallback |
| 3. MMF-R04a ordinary/run monster bases | Added MMF-V09(e)/(f): supplied ordinary/run values 200/100 yield distinct rates 300/150 at pct 150; derived-value production stays GDF-Q01 |
| 4. MMF-R05a extreme wrapped product | Added MMF-V10(f): low32(1073741824*100)=0, so both rates 0 reject unbounded multiplication then cap |
| 5. MMF-R05b averaging in an ineligible class | Added MMF-V23(e): class 0, eligibility false, qualified supplied items 60/20, final rate 256 rejects applying the adjustment that would yield 204 |
| 6. MMF-R06f remaining overlaps/early returns | Added MMF-V22(k) excluded null/type 4/type 5 entries and (l)-(q) adjacent precedence cases with different rates, path results or actionRate/log writes. Existing (g) covers cast before block. No internal query count is asserted |
| 7. MMF-R07 invalid/unavailable record fallback | Added MMF-V13(d)/(e): class 1 at count 1 and negative class -1 at count 1 exercise the upper/lower bounds; both give E=-50, end 1010 and one draw, rejecting E=0 defaults. Valid in-range class records are not mocked as unavailable |
| 8. MMF-R08b null/other type fatal | Added MMF-V28: null faults before requests; type 2 makes freezeNew(1050) requests before the fatal type assertion, without asserting post-fatal stored state |
| 9. MMF-R08c other hireling classes | No change: CSA-V04(c) discriminates every selected helper class and neighboring false classes; CSA-V10(i) class 271 and (d) class 359 exercise the nonzero category through fallback. MMF-V25 already supplies category 2/class 338. All nonzero categories take the same local routing branch |
| 10. MMF-R10a positive resistance bypass | No change: shared DMI-V07(a)/(c) discriminates zeroing positive resistance while retaining negative resistance; DMI-V24 discriminates the player bypass branch. MMF consumes that same per-entry operation with code-established length bindings and zero slot 0/no absorb |
| 11. MMF-R10b cap/nonimmune pierce | No change: DMI-V03(a)/(b)/(d) discriminates cap 75/95, and DMI-V04(b) uses regular monster resistance 99/pierce 10 to return 89 rather than immune 100. The common consumer owns these branches |
| 12. MMF-R11 negative length | Added MMF-V13(f): -1 returns before cold minimum/requests/draw, rejecting an unconditional minimum 1 |
| 13. MMF-R13d equal candidate expiry | No change: CSA-V08(d) rejects a cold equality write while still drawing; CSA-V11(c) rejects a freeze equality write while retaining its expiry and scheduling kind 12. They own the consumed refresh operations |
| 14. MMF-R14a threshold 19/20 | No change: CSA-V09(h)/(i) supplies seeds {0,20}/{0,19}, yielding exact rolls 20/off and 19/on, discriminating strict less-than 20 at the owning draw operation |

This revision adds no source reading, table cell, gameplay observation or label promotion. It adds
output inventory entries, the discriminating initialization input and vector cases for distinct
local results not already exercised. All earlier numeric expectations remain intact. Because
inputs, vector assertions and the output inventory changed, fresh full review FULL2 is required;
FULL1's zero mismatches does not complete review of this revision.

Root review of the added cases found that the initial MMF-V13(e) fixture supplied class 0/count 1
while declaring its record unavailable. The consumed GDF-R05a getter must select that in-range
record, so the fixture was inconsistent with its dependency. Before FULL2, it was replaced with
class -1/count 1, exercising the signed lower bound alongside case (d)'s upper bound. Expected
E=-50, end 1010, own-list values, draw and final seed are unchanged; no new predicate, rule or gate
was introduced. Every pre-FULL1 output assertion remains preserved.

## Fresh blind review FULL2 - 2026-10-15

A second fresh reader examined all 28 vector families and every lettered case, without reporting
a numeric case total. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 4, GATED 16, HARDEST 3. Its
COVERAGE heading reported 7, but the report contained five grouped bullets; this discrepancy is
retained rather than inventing two entries. The complete transcribed report and pre-fix documents
were preserved privately before edits. Findings do not approve the spec.

| Finding | Disposition |
|---|---|
| DISCRIM 1: movement final values do not reject reversed write/request order | MMF-V01 now records observable animationRate assignment before path-speed request in one ordered log. Numeric outputs remain 106/768; this tests output order, not internal query counts |
| DISCRIM 2: MMF-V06 default divisor 1 does not reject player division | Supplies cold divisor 4/freeze divisor 2; correct player fallback still ends 1060. Incorrect division would end 1015 or 1030, respectively |
| DISCRIM 3: MMF-V17 does not distinguish mulDiv operand order | Added MMF-V17(b), length 65537/r33: correct NRF-R05 operands (67,65537,100) multiply first and return 43909; swapping them selects divide-first and returns 43885 |
| DISCRIM 4: log-only MMF-V19 removal cannot reject an unsaved successor | No new removal fixture: main MMF-V19 now explicitly limits itself to supplied traversal logs and cites consumed CSA-R09/CSA-V11(a) for the saved-successor discriminator with mutating removal. Actual original removal/lifecycle stays gated; this fixture does not claim to establish it |
| GATED 16 | Retained the five unresolved rule leaves, the ten current questions and GDF-Q01; actual integration remains unverified |
| HARDEST 1: MMF-R06f/MMF-V22 overlapping predicates/different writes/HRC dependencies | No change: MMF-V22(g)/(l)-(q) now has distinct final rates, path writes, actionRate or preparation logs for each adjacency; HRC-R12a/R14 owns the three delegated formulas. Supplied predicates remain independent of actual producer integration |
| HARDEST 2: MMF-R13d/MMF-V20 expiry/new-duration AI | No change: explicit existing expiry, new candidate, unchanged list stats and final seed distinguish cold; freeze keeps end 1050 while requesting AI 1021. Equal/later dependency cases are named below |
| HARDEST 3: MMF-R13c/MMF-V19/MMF-V27 callback gates and requests | Existing supplied callback inputs/logs remain explicit; added MMF-V27(e) fixes a null freeze callback's state-off-before-fatal sequence. Actual cleanup/mode-bound processing/timer/lifecycle effects retain their gates |

| Grouped coverage finding | Disposition by subbranch |
|---|---|
| 1. MMF-R13d cold equality; freeze equality/later | No change: CSA-V08(d) discriminates no cold equality write while drawing; CSA-V11(c) covers freeze old expiry equal to candidate (no write but schedule), and old expiry 1005/candidate 1010 (write 1010), rejecting either a missing extension or equality write |
| 2. MMF-R10a nonpositive entry; positive bypass | No change: DMI-V08(b) supplies initially negative field and requires stored 0 before resistance; DMI-V19 includes a stripped zero cold-length entry retaining 0. DMI-V07(a)/(c) and DMI-V24 distinguish positive bypass from preserved negative resistance. These are the owning shared per-entry branches, with length bindings separately fixed in DMI's table |
| 3. MMF-R11 nonmonster with divisor greater than 1 | MMF-V06 now supplies cold 4/freeze 2 and keeps player length 60/end 1060; CSA-V10(f) independently exercises player fallback under Nightmare divisors 2, also without division |
| 4. MMF-R01 null unit/null list/invalid definition | Added isolated MMF-V08(b)/(c)/(d): null unit faults, null list returns 0, stat 96 at definition count 96 returns 0 despite stored 50. This uses an invalid upper-bound definition, not a fabricated missing in-range record |
| 5. MMF-R13c live cold clear; null freeze fatal | Live player/monster state 11/107 clears and no dead-gate call are already discriminated by CSA-V15(a)/(d)/(g); no duplicate cold callback added. New MMF-V27(e) asserts state1-off request before null fatal, without post-fatal stored-state claims |

NRF-R05's exact branch conditions were checked in the existing foundations contract before adding
MMF-V17(b). No original-source reread, new data collection, behavior correction, gate change or
runtime promotion was needed. Every prior numeric expected output remains unchanged; MMF-V01
adds an ordering assertion, MMF-V06 adds nonunit divisor inputs, MMF-V08/MMF-V17 group their existing
numeric cases before the added cases, and MMF-V19 narrows the claim of its log-only discriminator.
New inputs/assertions and execution clarification require a fresh full review FULL3; FULL2's zero
mismatches does not complete review of the resulting contract.

## Fresh blind review FULL3 - 2026-10-15

A third fresh reader checked all 28 families, every lettered subcase and grouped independent
inputs, without reporting a numeric case count. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1,
GATED 43, HARDEST 3. GATED counted the five unresolved rules, ten current questions and 28 supplied
stage families. The final corrected COVERAGE count was 10: the initial header said 11, but the
reader withdrew its last entry about MMF-R06f after finding no additional gap. Both the correction
and the complete report transcription were preserved privately before fixes.

| Finding | Disposition |
|---|---|
| DISCRIM 1: MMF-V09(b) lacks a distinct RunVelocity | Added only supplied CharStats.RunVelocity=9 alongside WalkVelocity=6. Correct speed remains 2104; using the run base would give 3156. All other case inputs/outputs remain unchanged |
| GATED 43 | No change: rule, question and supplied-stage limits remain explicit; the different count reflects this reader counting the 28 families individually, not new evidence gates |
| HARDEST 1: MMF-V22 dispatcher precedence | No change: adjacent supplied predicate cases assert different rate, actionRate, path or preparation-log results; HRC formulas remain delegated and actual input production remains gated |
| HARDEST 2: MMF-V21(c) two player cold calls and unchanged expiry | No change: persisted supplied first-call list, exact two-step seed and two state-off requests already make the second call explicit; equal expiry suppresses expiry writes, not the draw |
| HARDEST 3: MMF-V17(b) asymmetric mulDiv | No change: explicit caller arguments (67,65537,100), both branch tests and exact 43909-versus-43885 result establish the operand-order discriminator |

| Corrected coverage finding | Disposition |
|---|---|
| 1. MMF-R05b wrapping pair sum/adjustment | Added MMF-V23(f)/(g): INT_MAX pair wraps its sum before averaging, while INT_MIN/INT_MAX wraps the adjustment; both finish at rate 38 rather than the respective unbounded alternatives 256/448 |
| 2. MMF-R06a negative/wrapped unsigned product | Added MMF-V11(b)/(c): negative B=-1 gives u32 product 4294967196 and capped 32767; large positive B=1073741824 wraps the product to 0 and stores 0. These exercise the default caller independently of the attack caller or bare division helper |
| 3. MMF-R06d ramp, target and negative-speed clamps | No change: NRF-V22(a) discriminates five-tick ramp/target clamp and cleared acceleration; (b) retains acceleration at the negative-speed clamp; (e) distinguishes target equality; (f) establishes target-before-negative order. MMF-R06d consumes NRF-R11a directly |
| 4. MMF-R09 negative partner halving | No change: DPO-V21 supplies coldLength=-3/freezeLength=20/stat 118=1, requiring -1/10 and rejecting floor division or skipping the negative partner; DPO-R10 owns the consumed strip |
| 5. MMF-R10a positive-resistance bypass | No change: DMI-V07(a)/(c) and DMI-V24 already discriminate positive bypass versus negative-resistance preservation and player selection; same owner/vector citations as prior dispositions |
| 6. MMF-R11 wrapped cold end | Added MMF-V13(g): frame 2147483640 plus duration 20 requests expiry -2147483636, not an unbounded/saturated positive value; list fields and requested frames are supplied-stage assertions only |
| 7. MMF-R12 wrapped freeze end | Added MMF-V18(d): same wrapped list expiry with independently wrapped kind-2 request -2147483635; no actual queue execution claim |
| 8. MMF-R13d freeze extension/equality | No change: CSA-V11(c) explicitly covers old expiry 1020 retained, equal 1010 with no write, and earlier 1005 replaced by 1010, including required kind-12 requests; CSA-R08 owns the consumed refresh |
| 9. MMF-R13d first matching state list | No change: NRF-V33(a) supplies two state 11 lists across primary/secondary chains and requires the primary P11; (g) selects secondary S11 when primary has none. NRF-R12a owns lookup and MMF-R13d consumes its returned list; no replacement selection policy added locally |
| 10. MMF-R14a exact roll 20 | No change: CSA-V09(h)/(i) supplies exact 20/off and 19/on under the same cold-shatter operation; no duplicate seed binding |

There is no remaining eleventh coverage finding. The reader's withdrawn MMF-R06f entry causes
no edit. No rule, ownership, arithmetic, gate, label or observation changed in this follow-up;
all edits to the behavioral main are fixture inputs/cases. Prior numeric expectations are retained.

The exact next review scope is DELTA4: changed MMF-V09(b); added MMF-V11(b)/(c), MMF-V13(g),
MMF-V18(d), MMF-V23(f)/(g); and confirmation that DISCRIM 1 and the six addressed coverage subbranches
are resolved without weakening the documented gates. MMF-V11's prior cases are now grouped as (a)
with unchanged inputs/outputs. A fresh delta reader must recompute every listed case. This record
is a handoff for that review, not a completion or approval claim.

## Fresh blind review DELTA4 - 2026-10-15

A fresh smaller reader independently checked seven changed cases: MMF-V09(b), MMF-V11(b),
MMF-V11(c), MMF-V13(g), MMF-V18(d), MMF-V23(f) and MMF-V23(g). Counts on changed material:
MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0. No finding required a further change.

No rule behavior, input source, ownership or gate boundary changed after FULL3. The applicable
reader stopping condition is met. Actual input preparation, stored propagation, lifecycle and
gameplay observations retain their stated gaps; the document remains draft and maintainer
approval remains outstanding.

## Independent review spot-check - 2026-10-15

A separate reviewer re-derived sampled rules from the original 1.13c instructions before comparing
them with the spec; no source was edited and no gameplay was performed. Sampled for surprising
constants/exceptions: MMF-R01, MMF-R02, MMF-R04a, MMF-R05a, MMF-R05b, MMF-R06a, MMF-R06c, MMF-R06f,
MMF-R07, MMF-R08b, MMF-R08d, MMF-R11, MMF-R12, MMF-R13c, MMF-R13d, MMF-R14a; random others:
MMF-R03, MMF-R06b, MMF-R09. All agree with the current contract.

- Common 0x33110: encoded head decoded by hand; its first three calls land on 0x501E0, 0x32670 and
  0x30B80, then 0x323E0 with index 2, matching the companion chain. Predicate order 0x30B80, 0x2EF90,
  0x2EFC0, 0x2F010, 0x30C70, 0x30BD0, 0x2FAC0, default; movement path test precedes stat 67 reads
  and both writes; floor 25; signed products/divides for movement, unsigned for action; Sequence
  test reads actual type/mode (18, minus 30); clamp 15..175; block admits types 0/1/3.
- Common 0x323E0 table (.rdata, stride 12): kind 1 with (A, stat) = (120,93), (120,99), (120,105),
  (120,102), (150,96) for indices 0..4.
- Common 0x2FB10 / 0x30D50 / 0x30530: player base 101 only for mode 3 else 213; monster words +0x36
  ordinary/+0x38 for mode 15; Walk byte and monster signed word shifted 8; invalid record reads
  through null; eligibility 4/6 and 417/418, null false.
- Game 0xAFDB0 / 0xEB2C0: layer-0 writes 68=100, 67=75, 69=100 and 68/67/69=100 respectively.
- Game 0xDC780 / 0xDDAC0 / 0xDAE20 / 0xDBD90 / 0xDA7B0 and checked C of 0xDB4C0: request orders,
  -50 fallback, cold zero-divisor skip, freeze zero-divisor fatal, extend-only signed comparisons,
  freeze kind-12 at max end, kind-2 at frame+adjusted+1, unconditional cold draw with threshold 20,
  freeze callback zero delay to 15 and Normal-column condition, strip stats 153/118 with signed halving.

Recomputed vectors: MMF-V01-V04, V06, V08(a), V09(a)-(f), V10(a)-(f), V11, V12(a)/(b), V13(a)/(b)/(g),
V17(a)/(b), V18(a)/(b)/(d), V19, V20, V21 (including second seed {3459745073,1647453129}, roll 73),
V22(a)-(j)/(q), V23(a)/(b)/(f)/(g), V24(a), V27(b)-(e), V28(b): no mismatch. Edits in this pass were
notation only (run-together tokens and one missing `=` in MMF-V08(a)); no rule, input, output or
gate changed, so no fresh reader rerun is required.
