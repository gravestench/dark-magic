# Evidence: Missile Damage Parameters

Behavioral contract: [Missile Damage Parameters](../../specs/missile-damage-parameters.md).
Original artifact identities: [source manifest](../source-manifest.md).
Open evidence: [gap file](../../plans/gaps/missile-damage-parameters.md).

## Binary identity and methods

Evidence build is original Lord of Destruction Expansion 1.13c, Windows x86. References use
`LOD113C-D2COMMON`, D2Common.dll, 679936 bytes, SHA-256
`59fa5928522f566f2bf99675571206ad70df889c89d3d07fa87edf5083e06e10`, preferred base
`0x6FD50000`, and `LOD113C-D2GAME`, D2Game.dll, 1138688 bytes, SHA-256
`24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53`, preferred base
`0x6FC20000`. These are module-relative RVAs, not assumed loaded addresses. Descriptions
are analyst labels rather than original exported names. Original file hashes were checked
against the manifest before the new image readings on 2026-10-15.

Compatible instruction readings dated 2026-10-12 support the established column endpoints,
length, bypass consumer, creator chain and numeric fixtures. Their original private source
history and full original document were archived separately before rewriting. New readings
dated 2026-10-15 first checked availability of reconstructions matched against original
machine code. Matching reconstructions were read for scaling, item coefficient getters,
the two-handed wrapper, attack preparation, stored missile skill and Skills lookup. The
remaining selected missile/equipment functions had no matching reconstruction. Their
complete instruction-form bodies were read within the boundaries below. Previously encoded
builder opening, full builder selector table, mastery selector table, storage-copy loop and
Missiles descriptor instructions were corroborated from the matching original image.
Disassembling embedded tables as instructions supplies no meaning: pointer entries and
mastery byte indices were separately read as data.

Reuse assessment: current SCX-R34a/SCX-R34b owns the missile expression front end/resolver;
SLD-R02 owns the five bands and SLD-R01 through SLD-R11/SLD-R19 the skill endpoints. Their
2026-10-12/2026-10-15 records match the Common artifact and exact contexts used here.
NRF-R12b/NRF-R12c establishes read selection/minimum behavior only. NRF-Q20/STL-R06c leaves
write propagation unresolved. CSA-R02's fixed hireling category applies to the no-item branch.
The compatible SBA/SLD/DUR/MMF records locate equipment and preparation producers but do
not establish their complete production contracts. They therefore support dependency
identification rather than an integrated equipment result or a whole-operation no-draw claim.
The Blade companion's full Game 0x3A4E0 reading was assessed as downstream support only;
it establishes stored-stat consumption, not production or propagation of the stats here.

## Dispatch and functions

Creation roots remain owned by MCI. The compatible original chain is D2Game `0x6F930`
creation stage → `0x6F620` wrapper → thunk `0xB132`, IAT slot preferred VA `0x6FD18220`,
Common #10413/`0x6BAA0` builder; then thunk `0xB114`, slot `0x6FD18234`, Common
#10484/`0x6B4A0` store. The wrapper locally calls Game `0x37540` with zero selection before
the builder; actual weapon activation effects belong to MCI's caller boundary and are not
silently included in the supplied input fixtures. At the wrapper entry, source is EAX,
parent is EBX, missile EDI and level the stack argument. The builder receives five stack
arguments `(block,source,parent,missile,L)`; store receives `(source,missile,block,L)`.
Neither builder nor store supplies a meaningful success integer contract. These layouts
are evidence for semantic terms, not required implementation interfaces.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | 0x6F930; 0x6F620 | none | Creator and parameter wrapper | trigger | Compatible 2026-10-12 caller chain; wrapper instructions reread 2026-10-15; MCI owns pre-builder activation |
| `LOD113C-D2COMMON` | 0x6BAA0 | 10413 | Parameter builder | R04/R06/R07/R09/R10/R11 | Full 1888-byte original image plus instruction-form reading 2026-10-15; supplied external operations, not complete equipment/preparation effects |
| `LOD113C-D2COMMON` | 0x6B4A0 | 10484 | Metadata and stat requests | R12a/R12b | Full 1536-byte original image plus instruction-form reading 2026-10-15; requests and loop, not setter propagation |
| `LOD113C-D2COMMON` | 0x6A580; 0x6A4B0; 0x6B360; 0x6A3E0 | none recorded | Column physical min/max and element min/max | R01a | Compatible full 2026-10-12 instructions; all four 208-byte instruction forms reread 2026-10-15 for fallback/consumer argument checks |
| `LOD113C-D2COMMON` | 0x4DDB0 | none recorded | Five-band progression | R02 | Matching-reconstruction evidence reused through SLD-R02; no new progression claim |
| `LOD113C-D2COMMON` | 0x11E0 | none recorded | Argument-ordered percentage helper | R01a/R05 | NRF-R05 existing matching-reconstruction contract |
| `LOD113C-D2COMMON` | 0x69DC0 | none recorded | Column element length | R03 | Compatible 2026-10-12 full instructions; complete 176-byte instruction form reread 2026-10-15 |
| `LOD113C-D2COMMON` | 0x6B2B0; 0x9270; 0x6B160; 0x6A790 | 10231 for front end | Missile expression entry/interpreter/callback/resolver | R01b/Q02 | Reuse SCX-R34a/R34b and its current companion: front-end input/return coverage, resolver/owner-entry gates retained |
| `LOD113C-D2COMMON` | 0x6A2B0; 0x1250 | 10240; none | Stored signed missile skill; Skills lookup | R04 | Matching reconstructions read 2026-10-15: type 3/data guards, signed i16 return; signed skill bounds and selected record |
| `LOD113C-D2COMMON` | 0x52100; 0x51FF0; 0x50460; 0x50360; 0x4E900 | 10567; 10297; 10121; 11091; 10510 | Skill physical min/max, element min/max, duration | R04 | SLD current contracts reused; caller arguments newly traced; physical include-source0 and element/duration condition 1 |
| `LOD113C-D2COMMON` | 0x6AB20 | none recorded | Element mastery bonus | R05 | Full 192-byte original image and instruction-form read 2026-10-15, both selector tables decoded as data; stat reads and NRF percentage boundary |
| `LOD113C-D2COMMON` | 0x6B180 | none recorded | Critical decision | R08a | Complete 304-byte instruction-form read 2026-10-15; inline LCGs, signed comparisons, supplied third chance |
| `LOD113C-D2COMMON` | 0x6ABE0 | none recorded | Element stat additions | R11a | Complete 1376-byte instruction-form read 2026-10-15 through final write/return; NRF selected/minimum getter boundary |
| `LOD113C-D2COMMON` | 0x69C60 | none recorded | Selected field scaling | R11a | Matching reconstruction read first 2026-10-15; exact scaled field set and low32 signed /128 |
| `LOD113C-D2COMMON` | 0x31B90; 0x31B40 | 10641; none recorded | Player/monster selected weapon | R07b/Q04 | Complete instruction forms read 2026-10-15; local guards/requests only, inventory/skill/item dependencies remain open |
| `LOD113C-D2COMMON` | 0x24BF0; 0x229A0; 0x22970 | 10326; 10213; 10120 | Item predicate wrapper; attribute coefficients | R07a/R07b | Matching reconstructions read first 2026-10-15; table/type producer integration remains supplied |
| `LOD113C-D2COMMON` | 0x24C30; 0x1FB80; 0x24430 | 10711; 10051; 10744 | Alternate endpoint/type/classification dependencies | R07b/Q04 | Full instruction forms read 2026-10-15; compatible SLD/SBA/DUR/MMF limits retained, no complete production promotion |
| `LOD113C-D2COMMON` | 0x4F640 | 10804 | Weapon mastery dependency | R07b/R08b/Q04 | Full instruction form and compatible SBA record assessed 2026-10-15; context, special-case and enumeration producers remain open |
| `LOD113C-D2COMMON` | 0x32EF0 | 10932 | Source preparation | R11b/Q04 | Matching reconstruction read first 2026-10-15; reaches selection/ownership-transfer dependencies, actual effects/draw budget open |
| `LOD113C-D2COMMON` | 0x1A2D0 | 11104 | Fixed hireling category | R07a | Compatible CSA-R02 full instruction reading 2026-10-14; null/non-monster returns 0, five fixed classes |
| `LOD113C-D2COMMON` | 0x38B70; 0x38A80 | 10973; none recorded | Ordinary stat read; optimized selected-array leaf | R05/R07/R08/R09/R10/R11 | NRF-R12b/NRF-R12c reused; getter evidence does not establish setters |
| `LOD113C-D2COMMON` | 0x3A280; 0x3A740; 0x382F0 | none recorded; 10887; none recorded | List replace request; unit replace request; stat enumeration | R12a/R12b/Q06 | Storage caller resolutions/read 2026-10-15; full propagation/enumeration not established |
| `LOD113C-D2COMMON` | 0x7C8D0 | none recorded | Missiles descriptors/loader | R01/R03/R04/R06 | Existing GDF fields plus original 6592-byte loader image decode 2026-10-15; named descriptor links, not archive cell values |

### Instruction and data decisions

Builder `0x6BAA0..0x6BB82` clears 31 dwords before source/missile/type/class checks and
chooses column versus skill branch using signed linked skill and MissileSkill. All encoded
opening calls were resolved against the original rel32 bytes. Stored skill getter at
`0x6BC03` targets `0x6A2B0`, not the calculation entry `0x6B2B0`. The former returns signed
missile data+0x0A with type/data guards; the latter is the expression operation called by
column endpoints. Skills lookup `0x1250` rejects signed negative/out-of-count identities.
Physical requests at `0x6BC49`/`0x6BC57` pass 0; element min/max/duration at
`0x6BC74`/`0x6BC82`/`0x6BC90` pass 1. Duration's extra argument is ignored by its
consumer; only the two elemental endpoints use it as the mastery condition. Column mastery first updates minimum at
`0x6BBCE`, then maximum at `0x6BBE1`.

R07 begins at `0x6BCCA`. It passes player weapon-selection condition 1, performs critical
before share halving, and reads selected endpoint stats from the source unit, not from
the selected weapon. `0x6BD6C` requests mastery selector 1; critical helper's weapon
request uses selector 2. Attribute products are low32 before signed multiply-high /100
with sign correction. Parent branch starts `0x6BF70`, does not shift its endpoint stats,
does not set sourcePhysical, and does not read parent bypass stats. At `0x6C03E` the
missile stat 25 and signed max of retained stats 18/17 are added before the -90 floor.

The full original EType pointer table at `0x6C1BC` contains twelve preferred-address
entries targeting RVAs, in selector 1..12 order:
`0x6C0A7`, `0x6C0C1`, `0x6C0D4`, `0x6C0E4`, `0x6C0FB`, `0x6C112`,
`0x6C122`, `0x6C132`, `0x6C142`, `0x6C160`, `0x6C14B`, `0x6C0E4`.
Each target's field writes were read, not inferred from table position. Selector 12 shares
the cold triple; selector 10 jumps directly to the addition stage without a random draw.
The selector is zero-extended, tested for 0, decremented and bounded unsigned at 11.

Mastery `0x6AB20` uses byte indices at `0x6ABCC`, then five pointers at `0x6ABB8`.
The twelve indices are `0,1,4,2,3,4,4,4,4,4,4,2`, mapping type 12 to the same stat 331
branch as type 4. Stat 331's optimized getter and the ordinary stat 329/330/332 getter all
consume NRF's existing-stat/minimum contract. Unsupported indices return 0. The helper
returns only the percentage bonus; the builder separately adds it with wrapping arithmetic.

The complete element addition body `0x6ABE0..0x6B12A` reads layer-zero selected stats with
absent unit/list/definition returning 0. Poison stat 101 strictly positive skips both 59 and 326.
Scaling's reconstruction checked against original bytes touches offsets `0x0C,0x10,0x18,0x1C,0x20,0x24,0x28,0x2C,
0x30,0x34,0x38,0x40,0x44,0x4C,0x60,0x64,0x68`; notably not fireLength at 0x14 or
poisonLength at 0x3C. Source preparation occurs before the first source elemental read at
`0x6C16B`; it is not present on the parent path. Its matched body at `0x32EF0` reaches
`0x30530`, `0x31B90`, `0x20320`, `0x20270` and ownership transfer `0x3AC90`.
These dependencies prevent a universal unchanged-stat/no-draw claim.

Critical helper `0x6B180` reads a chance then performs unsigned `mul`/carry and unsigned
`div 100` on the new low word, retaining the high word in the seed. The first two comparisons
use signed `jl` for success; the third uses signed `jge` for failure. This establishes the
signed-chance interpretation independently of the unsigned remainder operation. Chance 141
and third chance test nonzero before stepping, so a negative nonzero chance still consumes
its draw and fails the signed comparison with a nonnegative remainder. A success exits
before reading any subsequent chance. The initial null-unit check only supplies chance 0;
the following seed access is unguarded, so a null unit does not return a successful integer.

R12 destination guard requires type 3. Metadata at missile data+0x14 retains old bits,
ORs bit 0 from parameter flag 1, bit 1 from parameter flag 4, and skips the metadata write
when data absent. List write calls target `0x3A280`; conditional unit writes target
`0x3A740`. Player-only notification calls at `0x38550` cannot be reached after the type 3
guard; they do not add a missile side effect. The original decoded copy loop at
`0x6BA50..0x6BA87` reads u16 layer and i32 value from each supplied eight-byte entry.
Enumeration eligibility is source/list presence and source list+0x10 sign bit set; it
requests stat 180/capacity 128 through `0x382F0`. Enumeration production remains Q06.

## Field bindings

Missiles descriptor identities are code evidence, not values from a data row. The selected
prepared record has stride 0x1A4; count/table are DataTables+0xB6C/+0xB64. Existing GDF-R07
and its companion own the loader schema limits. Relevant layout correspondence:

| Semantic field | Original prepared-record location | Evidence/consumer |
|---|---|---|
| MinDamage/MaxDamage | +0xB0/+0xB4 | descriptor strings and endpoint reads |
| physical min/max five steps | +0xB8..0xC8 / +0xCC..0xDC | descriptors, progression arguments |
| DmgSymPerCalc | +0xE0 | expression descriptor kind 0x19 and endpoint consumer |
| EType | +0xE4 | linked selector descriptor kind 0xD and builder |
| EMin/EMax | +0xE8/+0xEC | descriptor/endpoint reads |
| elemental min/max five steps | +0xF0..0x100 / +0x104..0x114 | descriptor/endpoint reads |
| EDmgSymPerCalc | +0x118 | expression kind 0x19 and endpoint consumer |
| ELen/ELevLen 1..3 | +0x11C/+0x120..0x128 | descriptor/length reads |
| CltSrcTown/SrcDamage/SrcMissDmg/Holy | +0x12C/+0x12D/+0x12E/+0x12F | adjacent descriptors; builder consumes only the latter three here |
| Skill/HitShift | +0x194/+0x196 | linked skill descriptor; signed 16/unsigned 8 consumers |
| ApplyMastery/`Half2HSrc`/MissileSkill | flags | GDF-R07 named flag descriptors; consumer mask requests |
| stored skill/level | missile data+0x0A/+0x0C | matching stored-skill getter; endpoint signed 16 fallback |
| Skills SrcDam/EType | +0x1A5/+0x1DC | builder u8 reads; SLD field contract |
| strength/dexterity coefficients | item record+0x106/+0x108 | matched getters return low16; caller sign-extends; raw 1 maps 100 |

On 2026-10-15 the original loader image showed descriptor at `0x7DBC9..0x7DBED` stores
string pointer preferred VA `0x6FDDAEFC`, kind 4, auxiliary 0, destination 0x12F, and callback 0.
The original string at RVA `0x8AEFC` is `Holy`. The loader prologue sets EAX 4/ESI 0; these
values are retained through that descriptor construction. This resolves the bypass byte's
column identity without assigning semantics from its spelling. Nearby `Half2HSrc` is a
different bit descriptor using auxiliary 14. Builder `0x6BC99..0x6BCCA` independently tests
the Holy byte's low three bits and sets parameter flags 0x100/0x200/0x400. No named archive
Holy value, effective file hash, physical MPQ attribution or compiled correspondence is
newly claimed. Dependency text names retain the manifest/GDF/SLD source-profile limits.

### Parameter layout mapping for independent binary inspection

Original block dword 0 holds flags 1 sourcePhysical, 2 critical, 4 secondaryMarker,
0x100/0x200/0x400 bypass conditions. Dwords 1/2 are physical endpoints; 3/4/5 fire endpoints
and fireLength; 6/7 lightning; 8/9 magic; 10/11/12 cold and coldLength; 13/14/15 poison and
poisonLength; 16 poisonCount; 17/18 life drain; 19/20 mana drain; 21/22 stamina drain;
23 stunLength;24/25/26 burn and burnLength;27/28/29/30 stat 121/stat 122/stat 120/physical
percent. Store calls map these fields exactly as the spec's ordered table. The duplicate
stat 315 requests occur at the fireLength and burnLength positions. This binary fixture
mapping supports inspection; it is not a required language/API/record layout.

## Corrections and question accounting

The pre-regeneration document had twelve parent rules, seven vector families, seven
questions and one observation. All twelve rule boundaries, seven supported numeric vector
families and the observation are retained. The complete pre-edit document and internal
investigation history were archived privately with integrity metadata. There was no
existing companion or gap file and no implementation-comparison section to lose. A
repository search for the file/prefix found only documentation references, with no plan
outside gaps, code or test consumption. Practical MDP ids remain; aliases were removed
under the unconsumed policy. No gameplay, source code, implementation plan or unrelated
shared file was changed by this writing pass.

| Question | Assessment and new evidence on 2026-10-15 | Current result |
|---|---|---|
| Q01 bypass-column identity | Original loader descriptor/retained register values and string image plus actual bit consumer | Closed: Holy; no archive-cell/gameplay assertion |
| Q02 calculation | Current SCX front-end/interpreter dependency assessment | Front-end consumer verified; resolver/owner-entry integration remains open |
| Q03 encoded builder/skill endpoints | Complete fingerprinted builder image, matched stored-skill/Skills lookup, current SLD contracts, exact five-call arguments/order | Local selection/request gap closed; source/Kick/expression production retains Q02/Q04 and SLD's gates |
| Q04 equipment/preparation | All ten originally named helpers individually assessed/read in proper source order, matched getter/preparation paths and current SBA/SLD/DUR/MMF evidence reused | Integration remains open for actual producer results/effects and draw budgets; supplied consumers verified |
| Q05 complete EType table | Original 12 pointer entries plus all target field writes | Closed: numeric routing complete including type 10 no-op and type 12 cold |
| Q06 duplicate 315/zero replacement | Complete original store including encoded copy loop, real replacing targets identified, NRF/STL current write gates checked | Request order verified; stored replacement/removal and enumeration remain open |
| Q07 gameplay | Compatible availability failure from 2026-10-12 assessed; migration requests no gameplay campaign | Observation remains pending, not impossible |

Specific corrections, with source rationale:

- The original R05 list omitted selector 12. Full mastery byte/pointer tables establish
  the same cold-stat 331 branch as selector 4. V10 now discriminates both.
- The original endpoint summary implied an absent missile always yielded a fallback level 0.
  All four entry guards show absence leaves the supplied level untouched; present missile
  with absent data gives 0. Existing numeric vectors are unaffected; V08 separates the cases.
- The original R04 opening and later selector entries relied on undecoded bytes/a decompiler
  lead. Full original image and data reads supply those instructions and table targets;
  selection is now verified with supplied endpoint operations and their dependency gates.
- The original R12 said the second 315 write replaces the element-type 1 length. The
  verified fact is two replacing requests, including a possible later zero. Propagation is
  still gated; V05/V06's stored-stat assertions are preserved as integration expectations,
  not promoted by their verified flag/seed/request assertions.
- V06's two-draw seed/numbers remain exactly as before. Its failure case now states the
  actual one-draw final seed. Additional third-draw arithmetic is derived from the same
  NRF recurrence, not from a gameplay observation.
- V07's original slot 3=300→150, slot 5=50 unchanged, -3→-1 and share 128 unchanged cases
  are all retained as semantic fire minimum/fireLength cases. The private original retains
  their original binary fixture encoding.

The legacy status ledger was static R01a/R02/R03/R06; asm-only R05/R07/R08/R09/R10/R11/R12;
unresolved R01b/R04. New labels promote only the explicitly bounded front-end, request,
arithmetic, routing, draw and metadata assertions supported above. Producer/storage subrules
remain unresolved. No old review is presented as a fresh review and no runtime label was added.

## Arithmetic and preservation record

V01: `7+4=11`, `3+11=14`, shifted 3584. V02: supplied 50 adds 7 before shifting 5376;
supplied 0 returns 3584. V03 retains the twelve outputs
`0,0,1,7,9,23,26,41,45,65,70,75`. V04 retains 25/25/60/63/84/88.
V05 retains mask 5→undead/beast, with stat 103/106 requests 1 and no 104 request.
V06 source `{1,666}` first widened product is `0x6AC6935F`, giving high 0/remainder 51;
second is `0x2C890AFD2F2ED81B`, giving remainder 31; third is
`0x13ADFE65280687C4`, remainder 12. Two-draw success's final seed remains
`{0x2F2ED81B,0x2C890AFD}`; one-draw failure ends `{0x6AC6935F,0}`.
V07 retains 300*64/128=150, -192/128=-1, fireLength 50 excluded, factor 128 skip.

V11 coefficient arithmetic uses wrapping products 5050/1275, signed quotients 50/12,
plus mastery 10 and source percent 3 gives 75. Halved share 64 yields 1280/2560.
V13 wraps percent additions then floors -106 to -90; it stores endpoints 13/24 unchanged
by that percent. V14's fire `(300+512)/2=406`, unshifted poison/drain/burn and poison
branch skips/count are separately asserted. These are exact hand traces and synthetic
fixtures; no new test code or gameplay trace was run.

## Critical observation

MDP-O01 is pending. With an owned matching original 1.13c game, use a character and skill
whose creation is confirmed to reach the parameter builder's source-contribution branch.
Use a safe stationary target area and record class, level, skill level, difficulty, equipped
items, selected weapon and the actual returned chance stats. Before one cast record source
seed at unit+0x20/+0x24; capture critical entry/return at Common 0x6B180 and subsequent
Common 0x6B4A0 conditional 141 request. Count/attribute any earlier calculation or equipment
draw separately. With stat 337=0 and stat 141=40, compare a seed yielding first 51/second 31
to the two-step recurrence in V06; with stat 141=0 and no third chance, compare the one-step
failure. Read the missile's selected stat 141 afterward only as a separately recorded
storage assertion, with definitions/list state and valid getter semantics identified.
Record module fingerprints, actual seeds/chances/returns, ordered requests, missing/truncated
records and capture limits. An uncontrolled seed/chance is a different input, not a fitted
V06 match. Other rules need their own measured assertions under Q07; O01 is not blanket
runtime confirmation.

## Regeneration record

2026-10-15: preserved the pre-edit document privately; assessed consumption, exact dependency
coverage and the seven old questions before collection. The targeted original binary and
matching-reconstruction readings above close the column-identity, encoded selection,
complete routing and missing local vector-definition gaps. Fifteen initial writer executable
families include all seven original families. Four open question families remain in the
linked gap file. Static supplied-consumer labels and unresolved producer/storage subrules
are distinguished; document status remains draft. No gameplay campaign or implementation
comparison was undertaken. Writer self-review checked all twelve parent rule boundaries,
the seven original numeric families, all seven question outcomes, the observation, exact
integer/draw semantics, source/dependency coverage, public/private separation and partial
acceptance gates. The spec checker returned 0 errors and 0 warnings on the final writer
pass; exact command was `python .agents/skills/spec-writing/tools/check_spec.py missile-damage-parameters`.
The private preservation record holds original/final integrity metadata. Fresh-reader
passes are recorded below as they occur.

## First fresh-reader review and disposition

The first blind reader on 2026-10-15 recomputed all fifteen then-current vector families;
no separate numeric case count was reported. Reported categories were MISMATCH 0, GUESS 1,
GAP 0, DISCRIM 2, GATED 16, COVERAGE 11 and HARDEST 3. The report actually listed ten
coverage bullets. Both the reported count 11 and the ten-item inventory are retained here;
no eleventh item is inferred. This pass was not empty and did not establish completion.
The following changes retain the verified/gated boundaries and require a second fresh pass.

### Source-backed branch clarification

The GUESS concerned whether a nonzero source share halved from 1 to 0 keeps source selection
for later elemental additions. The competing readings were a single remembered source choice
versus a fresh test of the modified share. The same fingerprinted Common builder establishes
two distinct decisions: source/parent physical choice at `0x6BCCA..0x6BF70`, share halving
and local store at `0x6BD5F..0x6BD68`, then a reload of that modified local at `0x6C160`.
Zero branches to `0x6C18D`, where the retained parent share and parent presence are tested.
The source element path pushes shift 8 and source at `0x6C170..0x6C175`; the parent path
pushes shift 0 and parent at `0x6C19D..0x6C1A2`. Both call element addition `0x6ABE0`, then
scale by their selected share. Full addition-helper reading already establishes how the
supplied unit and shift affect its writes; no new producer inference is needed.

Therefore a source share 1 halved to 0 does not revisit parent physical/critical processing,
but can use parent elemental additions/share with no source-preparation request. The
earlier section 6 sentence concerned physical precedence but did not explicitly distinguish
this later selection. R09/R11a/section 6 now distinguish them. V16 supplies source fire
stats 1000/2000 and parent 10/20, row 300/500, parent share 64: fire 155/260, source seed one
step, parent seed unchanged, sourcePhysical true, critical false and no preparation call.
This resolves the ambiguity from instructions; agreement or a synthetic result alone
would not establish that branch.

### Complete finding inventory

| Finding | Disposition and exact evidence/fixture |
|---|---|
| GUESS 1: physical versus elemental selection after halving 1 to 0 | Source decisions traced above; current R09/R11a/section 6 clarified and V16 added with distinct unit stats/seeds/request log |
| DISCRIM 1: V02's small operands permit reversal | V02(c) now supplies p 65537/value 100001 at the actual column percentage caller: correct bonus 65500655, endpoint 65600656, reversed endpoint 65637001; logged ordered arguments |
| DISCRIM 2: V10's small mastery permits reversal | V10(b) supplies mastery 65537/value 100001 to the actual mastery caller: same branch-sensitive bonus/augmented result and reversed failure; V10(a)'s six original small cases retained |
| COVERAGE 1: R01b negative fallback and invalid pool/reference | V02(d) separates direct-front-end level -1 to stored 10 from level 0 remaining 0; V02(e) covers reference equality 4, unsigned out-of-range and absent pool, no evaluation |
| COVERAGE 2: R03 signed length | V04 adds negative base/level results and signed overflow to INT_MIN without a clamp |
| COVERAGE 3: R04 mastery and skill bounds | V09(f) asserts ordered minimum 30 to 45 then maximum 40 to 60 mastery operations; V09(g) covers negative linked-skill column branch and stored-negative fallback to linked skill equal to count, rejected |
| COVERAGE 4: R05 zero mastery and operands | V10(b)/(c) cover ordered branch-sensitive operands and zero cold mastery with no percentage request |
| COVERAGE 5: R07a no weapon, alternate endpoints, coefficients | V17(a) fixed hireling/no-inventory dexterity contribution; (b) alternate 159/160 priority/no classification; (c) raw 1/default 100 and signed 65535 to -1; (d) negative coefficient truncation |
| COVERAGE 6: R08a first success and negative first chance | First success was already V06(c), retained with no later read; V06(e) adds negative first chance, one required failed draw and exact final seed |
| COVERAGE 7: R09 parent critical success and endpoint wrap | V12(c) asserts first parent success, wrapped endpoint products and exact parent/source seed ownership |
| COVERAGE 8: R10 wrapped sum and negative bonuses | V13(b)/(c) assert wrapping percent sum and signed maximum of two negative competing bonuses |
| COVERAGE 9: R11a parent and remaining groups | V16 discriminates modified-share parent selection; V18 enumerates all mapped stat additions in the parent/shift 0 stage and every scaled/excluded field group |
| COVERAGE 10: R12a preset markers and ineligible enumeration | V15(d) retains both preset markers, logs only the 31 base writes including both zero stat-315 requests, skips conditional flags/enumeration/copy |
| GATED: Q02/Q04/Q06/Q07 | All four question families retained with exact integration/observation boundaries; no supplied fixture closes a producer or gameplay gap |
| GATED: R07b/R08b/R11b/R12b | All four unresolved subrules retained; equipment/mastery/preparation/storage effects remain unverified |
| GATED: V02/V05/V06/V09/V11/V12/V14/V15 | All eight reported supplied-input families retain producer/write/observation limits; added cases do not imply integrated acceptance |
| HARDEST 1: V11 coefficients, halving and percent | Original 75 percent and 1280/2560 contributions retained; V16/V17 add zero-halving and alternate/sign boundaries |
| HARDEST 2: V14 additions, poison, selective scaling | Original 406 fire, poison 27/119, count 0/4 and drain/burn outputs retained; V18 covers remaining groups/parent stage |
| HARDEST 3: V15 duplicate stat 315, flags and layers | Original 31-request order, first stat 315=50/second stat 315=0, conditional 103/106/141 and layered 180 requests retained; cases split with explicit outputs and preset/ineligible case added |

The operand fixtures use NRF-R05's first-operand-large branch: second operand 100001 is not
above `0x100000`; first 65537 is above `0x10000`; divisor 100 is not above 4096, so division
occurs first. `655*100001=65500655`; reversed gives `1000*65537=65537000`. Added base 100001
gives 65600656 versus 65637001. This is independently applicable to each caller's established
argument placement, not merely a shared-helper test. V16's parameter flags are transient
parameter assertions; real equipment/mastery results and actual replacing writes remain
gated. Current executable inventory is eighteen families, with no newly claimed gameplay
observation. Mechanical lint and preservation checks were repeated after these edits;
the second fresh-reader result follows.

## Second fresh-reader review and disposition

The second blind reader on 2026-10-15 recomputed all eighteen families and reported 98
independent executable cases, excluding O01. Independent levels, selectors and alternatives
were counted separately. Its exact per-family counts were:

| Family | Cases | Family | Cases | Family | Cases |
|---|---:|---|---:|---|---:|
| V01 | 1 | V07 | 3 | V13 | 16 |
| V02 | 8 | V08 | 5 | V14 | 2 |
| V03 | 13 | V09 | 10 | V15 | 4 |
| V04 | 12 | V10 | 8 | V16 | 1 |
| V05 | 1 | V11 | 1 | V17 | 4 |
| V06 | 5 | V12 | 3 | V18 | 1 |

Reported categories were MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 4, COVERAGE 8 and
HARDEST 3. The reader covered column/fallback/progression/calc/shift/duration; pool and
evaluator routing; ordered skill/column selection, mastery and Holy; source/parent physical
and critical decisions; clamp/routing; modified-share elemental selection and scaling;
metadata/write requests and their explicit production gates. It found no supplied-case
arithmetic mismatch or unresolved interpretation, but its discrimination/coverage findings
still required the fixture changes below and a third full pass. The 98-case total describes
the reviewed revision, not an invented count for the expanded fixtures.

| Finding | Disposition |
|---|---|
| DISCRIM 1: V04's unspecified HitShift default did not distinguish an erroneous duration shift | V04 explicitly supplies HitShift 8 for its same unshifted durations 25/25/60/63/84/88 and signed cases; a duration shift now fails |
| COVERAGE 1: R01a's other column families | V02(f) adds four independent endpoint calls on one synthetic record with distinct bases/increments and distinct physical/element references; exact outputs 4096/7680/13824/18688 and reference requests 0/0/1/1 |
| COVERAGE 2: R03 absent row | V04 adds class 0/count 1/selected row absent/L10, exact return 0 with no field read or draw |
| COVERAGE 3: R06 Holy bit 1 and unrelated bits | V05(b) supplies Holy 2, only demon condition/stat 104 request; V05(c) supplies Holy 128, no bypass condition/conditional bypass request |
| COVERAGE 4: R07a monster inventory and non-item selected weapon | V17(e) supplies monster/class 271/inventory and selector result N/type player; critical still requests selector 2 for non-null N, then physical uses the no-item endpoint/dexterity branch with 768/1280 and percent 21; no item-path requests |
| COVERAGE 5: R08a absent contributor fault | V06(f) directly supplies absent unit; chance defaults 0, first seed access faults, no completed step/write/return or later chance request |
| COVERAGE 6: R09 present parent with zero share | V12(d) retains present P and nonzero parent stats but sets both shares 0; all physical contributions/conditions remain 0/false, both seeds unchanged, no parent request |
| COVERAGE 7: R11a both shares 0 | V18(b) starts with a nonzero supplied record and present parent stats, both shares 0; every field remains exactly supplied, no preparation/getter/scaling/critical request |
| COVERAGE 8: R12a absent source with otherwise eligible enumeration operations | V15(e) supplies absent source and logs the complete normal metadata/base/conditional writes while no eligibility/enumeration/copy operation is reached |
| GATED 1: Q02 | Calculation production remains open; the additional endpoint fixture supplies results and does not verify the bytecode producer or unresolved resolver |
| GATED 2: Q04 | Equipment/mastery/Kick/override/preparation production remains open; monster/non-item requests are supplied-result consumption only |
| GATED 3: Q06 | Replacing mutation and enumeration remain open; absent-source and Holy cases assert logs, not actual stored stats |
| GATED 4: Q07 | All observations remain pending; no additional fixture is a gameplay record |
| HARDEST 1: R07a selection/sign/truncation/halving | Existing V11/V16/V17 arithmetic and request distinctions retained; new V17(e) adds only the monster/non-item path |
| HARDEST 2: R08a short circuit and seed owner | All exact seeds and reached-draw assertions retained; absent-unit failure is a separate direct helper case |
| HARDEST 3: R11a repeated selection and differing scaling set | V16 and V18's exact selection/group outputs retained; no-work V18(b) distinguishes a different early exit without duplicating arithmetic machinery |

These are fixture additions to already established decisions, rather than new behavior or
evidence-label promotions. Existing matching-artifact source records support the guards
and field choices; no fresh decoding, data collection or gameplay was needed. V02(f)'s
four pre-percentage values are 14/25/36/49; percentages 20/20/50/50 add 2/5/18/24, producing
16/30/54/73 before shift 8. No calculation token semantics are inferred from those supplied
results. Eighteen families remain. The third fresh-reader result follows; the spec checker
and original-vector preservation checks were repeated after this edit pass.

## Third fresh-reader review and disposition

The third blind reader on 2026-10-15 covered all eighteen families and counted 55 cases
under its convention: each lettered variant once, a multi-selector/level list as one case,
and V14's separate stat-101 alternative separately. This differs from the second reader's
independent-alternative counting; the change from 98 to 55 is a counting convention, not
removal of assertions. Reported categories were MISMATCH 0, GUESS 0, GAP 0, DISCRIM 3,
GATED 7, COVERAGE 3 and HARDEST 3. The two meaningful discrimination findings below add
two targeted variants; no rule, gate, dependency, seed expectation or existing numeric
result changes.

| Finding | Disposition |
|---|---|
| DISCRIM 1: shifting base/progression separately versus their sum in V01 | No new numeric fixture: with the same masked shift and low32 wrapping this transformation is exactly equivalent. Proof below. V01's discrimination text now names band length/base selection and does not claim to reject this equivalent transformation |
| DISCRIM 2: V02(a)'s exact 50 percent permits a percentage after HitShift | Added V02(g), p 1/unshifted value 14/HitShift 8: percentage truncates to 0 before shifting, exact endpoint 3584; shifting first would add 35 and produce 3619 |
| DISCRIM 3: V09(g)'s rejected linked skill does not distinguish a negative stored-skill fallback | Kept the count-equality rejection case. Added V09(h), linked 2/count 3/stored -1, valid skill 2 with SrcDam 128/EType 5: requires five successful linked-skill requests and selected 10/20/30/40/50; choosing stored -1 rejects and fails |
| COVERAGE 1: R11a source lightning/magic/cold additions | No duplicate fixture matrix. The complete addition-helper evidence establishes one semantic mapping for stats 48..55, each with the same supplied-unit getter and supplied shift argument; the source caller passes 8, parent passes 0. V14 exercises source shift 8 and V18 each mapped non-fire field with parent shift 0. Their composition supports the declared uniform recipe; it does not claim every source field was separately observed |
| COVERAGE 2: R12a enumeration beyond 128 | No invented oversized fixture or local clamp. The caller requests capacity 128 and consumes the producer's returned entries in order; actually producing a conforming bounded enumeration remains R12b/Q06. Behavior for a fabricated over-capacity producer response is not established by supplied valid entries |
| COVERAGE 3: R01b actual expression draw | No duplicate evaluator/RNG fixture. SCX-R34a with SCX-R15 and SCX-V13 owns actual missile-context random draws/seed ownership; SCX-V34 covers front-end context. MDP's arithmetic fixture supplies evaluation results, while actual unresolved parameter/owner-entry production retains Q02 |
| GATED 1: Q02 | Unresolved calculation-produced results remain open; the new percentage fixture supplies p and does not verify its production |
| GATED 2: Q04 | Equipment/mastery/Kick/override/preparation results and effects remain open; the new negative stored-skill consumer fixture does not close external skill endpoint producers |
| GATED 3: Q06 | Stat replacement and enumeration production remain open; no greater-than-capacity return or final stored replacement was fabricated |
| GATED 4: Q07 | Gameplay observation remains pending, independently of static supplied-input assertions |
| GATED 5: R07b/R08b/R11b | The reported equipment/mastery/preparation producer group remains unresolved |
| GATED 6: R12b | The reported stat-mutation/enumeration group remains unresolved |
| GATED 7: V01 through V18 | All eighteen families retain their stated supplied-input/isolated boundaries and no-gameplay status; this report grouping does not gate already verified supplied arithmetic |
| HARDEST 1: R07a/V11/V17 | Selection, sign conversion, wrapped product/truncation and halving outputs/request distinctions retained; no new ambiguity |
| HARDEST 2: R08a/V06 | Short circuit, first/conditional draws and source/parent seed attribution retained; no new ambiguity |
| HARDEST 3: R11a/V16/V18 | Modified-share reselection and exact differing scaling groups retained; no new ambiguity |

For any i32 bit patterns `x,y` and effective shift count `n` obtained by masking with 31,
multiplication by `2^n` distributes
over addition modulo `2^32`. Thus
`low32(low32(x+y)<<n) = low32(low32(x<<n)+low32(y<<n))`.
No saturation, signed comparison, intervening division or differing shift count occurs in
the proposed V01 transformation. An overflow example cannot distinguish equal results;
no false distinction or new overflow behavior is introduced. This equivalence does not
move percentage division across the shift: V02(g)'s `trunc(1*14/100)=0` versus
`trunc(1*3584/100)=35` proves that different operation order is detectable.

V09(h)'s changed count makes linked skill 2 valid while stored -1 remains invalid; the
assertions are successful lookup 2, physical calls with include-source 0, elemental calls
with mastery 1, duration, exact selected values and shares 128/0. The five selected
endpoint and length values are supplied results, not newly observed skill
damage. No artifact reread or gameplay was required for these fixture changes: existing
caller/consumer readings already establish the operations. Three complete fresh-reader
rounds are now documented. A subsequent delta review is scoped to V02(g), V09(h), V01's
corrected discrimination statement and the finding resolutions above; its result is pending.

## Blind-implementer delta pass 4 — 2026-10-15

After three full passes, a fresh reader checked exactly the two added cases V02(g) and V09(h), plus the V01 discrimination-wording correction, under the delta-review rule. Counts for changed material only: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0. All three previous findings are resolved. The reader independently obtained 3584 versus the incorrect after-shift result 3619, the successful linked-skill-2 request/value sequence for stored -1/count 3, and the modular left-shift equivalence proof. No producer or stored-effect claim was inferred.

Rules, ownership and gates were unchanged after full pass 3. This clean applicable delta completes independent documentation review; no further behavior, input, vector, output, acceptance or boundary edit is made in this record. The full-pass counts retain their own case-count conventions rather than being added together. All original twelve parent families, seven legacy executable families, seven historical question outcomes and the observation procedure are preserved. Status remains draft, gameplay observations and exact integration gates remain pending, and maintainer approval has not been inferred. Mechanical lint, committed scope, technical-token and preservation checks are repeated for publication.
