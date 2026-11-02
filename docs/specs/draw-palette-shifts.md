# Monster palette shifts

**Description**

Monsters can reuse the same artwork with different colour tables. The client selects a table for the monster's current appearance and can prepare an alternate version that changes selected colours while preserving the others. This contract describes that selection and preparation; the effects that choose a monster's appearance remain partly unresolved.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [numeric foundations](core-numeric-and-rng-foundations.md), [record layouts](core-record-layouts-and-ids.md), [game-data table records](core-game-data-table-records.md)
Ids: PLS-R, PLS-Q, PLS-V, PLS-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/draw-palette-shifts.md)
Open evidence gaps: [gap file](../plans/gaps/draw-palette-shifts.md)

## 1. Scope

The bounded behavior is selection of a monster's indexed-colour replacement table and preparation of its alternate replacement tables. Inputs are prepared monster/state records, active state membership, prepared shared colour banks, a cached set of monster base tables, an optional global random-table collection, an optional alternate-colour mode and an appearance selector. State-colour recomputation and its priority over the monster-specific fallback are defined below. Loading those assets, complete recomputation triggers, and choosing the graphics selector from creation, unique modifiers or other appearance effects remain explicit integration questions. Room/environment colour and intensity belong to [room and ambient lighting](draw-room-and-ambient-lighting.md); light-map construction belongs to [light map and light radius](draw-light-map-and-light-radius.md). Screen fades and world-clock/quest decisions are separate contracts. No rule here establishes a global screen-palette change.

## 2. Trigger and ownership

D2Client's `GAME\PalShift` consumer is reached during component rendering when no replacement table has already been supplied by that caller. Its result supplies a table reference to that rendering operation. Alternate-table preparation is reached while a monster-specific palette asset is being loaded. These are client presentation operations; they schedule no server event, send no message and draw no random number within the verified stages. Their invocation frequency and complete asset/selector producer chains remain PLS-Q01/PLS-Q02.

## 3. Inputs and data

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| unit presence/type | supplied unit | optional; type identity per record foundations | Only monsters can return a table |
| monster class | supplied monster identity | nonnegative i32 in the initialized cache domain | Chooses its cached palette asset |
| cached base tables | monster palette cache | present/absent; 8 banks of 256 u8 entries | Each entry replaces one source colour index |
| appearance selector | supplied monster graphics appearance | i32 | For the verified bank domain: 0..7, or 8..37 with the random collection enabled |
| base appearance selector | supplied monster graphics appearance | i32 | Retained by the state-colour stage; complete later base/current updates remain PLS-Q02 |
| random collection enabled/tables | prepared optional global collection | boolean; 30 banks of 256 u8 entries | Global tables selected instead of monster-specific banks |
| alternate-colour mode | supplied client presentation setting | boolean | Permits alternate tables |
| monster records | prepared MonStats and MonStats2 collections | counts i32, MonStatsEx signed i16; optional linked record | Resolve class through its MonStatsEx link |
| transformation kind | selected extended monster record's `localBlood` field | u8 | Code compares with 2; original column binding and a named-cell/compiled correspondence are recorded in the companion |
| alternate tables | prepared alternate monster tables | 8 banks of 256 u8 entries | Output of PLS-R05 where eligible |
| colour translation | supplied prepared translation table | 256 u8 entries | Maps each base-table value to a candidate replacement |
| state definitions | prepared States collection | nonnegative i32 count; present records indexed from 0 | Each record supplies unsigned u8 `colorpri` and `colorshift` |
| active state membership | supplied monster's current membership through RLI-R03 | active/inactive per state identity | Membership consulted during state-colour recomputation |
| state colour request | supplied monster presentation field | i32, producer domain 0..255 | Zero permits the monster fallback; nonzero requests a shared colour bank |
| shared colour banks | supplied prepared shared collection | 255 distinct references, each to 256 u8 entries | Request c=1..255 selects bank c-1; original availability remains PLS-Q01 |

| Constant | Origin | Binding |
|---|---|---|
| monster type; special class identities 363 and 364 | engine code | Fixed by original code; no archive name mapping asserted |
| selector split 8, random count 30, bank length 256, bank count 8, transformation-kind test 2 | engine code | Fixed by original code |
| MonStatsEx linkage | prepared table records | Resolved from data through GDA-R04a/GDA-R08a; no named archive cell value asserted |
| transformation-kind field | `MonStats2.localBlood` column; original schema | Data supplies the value; original code fixes the column binding and compares its unsigned byte with 2 |

The supplied class is a valid cache index when a monster is present; the selector consumer does not establish a safe out-of-range class operation. PLS-R05 independently checks class/link bounds before preparing alternate tables. Banks and translation entries are logical values, not required storage layouts.

## 4. Rules (execution order)

**PLS-R01. Eligibility and cache** [runtime-confirmed (1.13c)]. An absent unit or any non-monster returns no replacement table. A monster whose cached palette asset is absent also returns no replacement table before reading its appearance selector or consulting the random collection. No writes or RNG draws occur. Runtime coverage is the present-player return case in PLS-V07; other branches retain static support.

**PLS-R02. Low selectors** [verified in 1.13c code (static)]. With PLS-R01 eligible, selectors 0 and 1 return the corresponding monster base bank only for class 363 or 364, inclusive; every other class returns no table. These selectors bypass both the random collection and alternate-colour mode. Selectors 2..7 proceed to PLS-R04. No draws or writes occur.

**PLS-R03. Random-table selection** [verified in 1.13c code (static)]. With PLS-R01 eligible and the selector outside unsigned 0..7, an enabled random collection takes priority over alternate-colour mode. Compute the i32 difference `selector - 8`; if it is negative or at least 30, take the fatal assertion path without returning a table. Otherwise return global random bank `selector - 8`. No random draw occurs here: the selector is supplied. With the random collection disabled this branch continues to PLS-R04; behavior for selectors outside the supplied valid bank domain is PLS-Q04, not an invented clamp.

**PLS-R04. Base versus alternate selection** [verified in 1.13c code (static)]. For eligible selectors 2..7, return the base bank of the same index unless all three conditions hold: alternate-colour mode enabled, an extended monster record resolves, and its transformation kind equals 2. When all hold, return the alternate bank of that index. Resolve the extended record by checking signed class against the MonStats count, then its signed i16 MonStatsEx against the MonStats2 count; an invalid identity gives no extended record. This selection performs no table mutation and no RNG draw. The base banks returned by PLS-R02 do not pass through this rule.

**PLS-R05. Prepare alternate tables** [verified in 1.13c code (static)]. Given a destination asset, base banks and translation table, resolve the extended record as in PLS-R04. Invalid class/link, absent extended record, or transformation kind other than 2 makes no destination write. Otherwise process all eight banks, each containing all 256 entries. First copy every base entry into its corresponding alternate entry; then for each entry compute `candidate = translation[base entry]`. If candidate is nonzero, replace that alternate entry with candidate; if candidate is zero, retain the copied base entry. Reads and writes use u8 values without signed conversion, rounding, saturation or RNG. Base banks are unchanged. The translation is applied once to the base value, never repeatedly to an already translated value.

**PLS-R06. Asset preparation and availability** [unresolved]. The palette-selection stage requires the class's cached base tables and, where selected, prepared alternate or global tables. The original asset-loading path requests monster-specific tables and optional global tables. Exact failure handling, complete initialization/cleanup lifetime and the relationship between successful reads and table availability require PLS-Q01. A supplied table does not prove the loader's behavior.

**PLS-R07. Appearance producers and caller priority.**

- **PLS-R07a. Full producer integration** [unresolved]. The graphics appearance selector can be written during monster creation and later pet/presentation updates. Complete creation/variant selection, later replacement of the base/current graphics selectors and the eligibility and timing of state-colour recomputation require PLS-Q02. Supplied-selector PLS-R01..PLS-R05 and supplied-state PLS-R07b/PLS-R07c do not establish those producers or triggers.
- **PLS-R07b. Recompute a monster's state colour** [verified in 1.13c code (static)]. For a present monster distinct from the local player, examine the stable supplied States collection in ascending state-identity order. Start with best priority 0 and no selected identity. An active state replaces the current selection only when its unsigned `colorpri` is strictly greater than the best priority; equal priorities retain the earlier identity, and priority 0 never wins. After the scan, a selected identity greater than 0 writes that state's unsigned `colorshift` into the monster's state colour request. No selection, an empty collection, or winning identity 0 writes request 0. A winning state with `colorshift` 0 also writes 0; the scan does not skip it in favour of a lower priority. This stage changes neither active membership nor the base/current graphics selector, and draws no random number. Its complete callback eligibility remains PLS-Q02.
- **PLS-R07c. State table before monster fallback** [verified in 1.13c code (static)]. At the component-rendering palette-choice stage, a nonzero state colour request c selects shared colour bank c-1 and bypasses PLS-R01..PLS-R04. This priority holds even when the monster cache is absent or its graphics selector would fail the global bounds check. Request 0 runs PLS-R01..PLS-R04 with their supplied inputs and retains their result, including no table or a fatal assertion. Choice changes no bank, selector, membership or seed. The shared-bank content/availability and the upstream request's production are separate from this supplied-input choice.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| PLS-R01 | Presence/type/cache eligibility | runtime-confirmed (1.13c) | none | PLS-V01, PLS-V07 |
| PLS-R02 | Special low selectors | verified in 1.13c code (static) | none | PLS-V02 |
| PLS-R03 | Global-table priority and bounds | verified in 1.13c code (static) | none within supplied collection | PLS-V03 |
| PLS-R04 | Alternate selection | verified in 1.13c code (static) | none within supplied records/tables | PLS-V04 |
| PLS-R05 | Sparse one-pass translation | verified in 1.13c code (static) | none within supplied records/tables | PLS-V05, PLS-V06 |
| PLS-R06 | Asset lifetime/loading | unresolved | PLS-Q01: actual availability/cleanup integration | none |
| PLS-R07a | Complete selector producers and recomputation triggers | unresolved | PLS-Q02: creation/variant/pet integration and update eligibility | none |
| PLS-R07b | State priority and colour-request write | verified in 1.13c code (static) | none within supplied stable states/membership; actual triggers PLS-Q02 | PLS-V08 |
| PLS-R07c | Caller state-table priority | verified in 1.13c code (static) | none within supplied request/shared banks; actual availability PLS-Q01 | PLS-V09 |

## 5. Outputs and side effects

| Output / side effect | Rule |
|---|---|
| no table, selected base/alternate/global table reference, or fatal assertion | PLS-R01 through PLS-R04 |
| eight alternate banks overwritten; base banks retained | PLS-R05 |
| state colour request overwritten | PLS-R07b |
| shared-bank reference or monster-fallback result | PLS-R07c |
| asset availability and complete selector production remain gated | PLS-R06, PLS-R07a |

The verified stages do not change gameplay state, create light sources, interpolate ambient RGB, schedule events, send messages or advance any seed.

## 6. Branches and edge cases

Both low selectors require an eligible class even when global random tables are enabled. A missing class cache wins over every selector mode. Zero translation means preserve, including a nonzero base colour. Translation index 255 is unsigned. Selectors 8 and 37 choose the first and last global banks; selector 38 asserts when that collection is enabled. A negative selector also reaches that bounds failure with the collection enabled. With that collection disabled, selectors beyond the eight-bank domain and malformed asset buffers remain PLS-Q04. The reproduced stages have no difficulty or PvP branch and target Expansion 1.13c only.

State priority is unsigned and ties retain the lowest active identity with the winning priority. Identity 0 can win the scan but then clears the colour request; the scan does not choose a lower-priority state instead. A nonzero colour request bypasses the monster cache and selector checks at the caller stage. The graphics appearance selector and the state colour request are distinct presentation fields.

## 7. Open questions

- **PLS-Q01** [unresolved, integration]: trace all asset-open/read/close and cache-publication/cleanup paths. Establish when optional global, alternate and shared state-colour tables exist, missing-file behavior and resource lifetime. This gates PLS-R06 and original table availability, not supplied-table arithmetic or supplied-bank choice.
- **PLS-Q02** [unresolved, behavior]: complete creation and later appearance-update paths through remaining variant/pet helpers and callback eligibility. Establish the full graphics-selector production and `MonStats2.Utrans` integration, and when state changes invoke PLS-R07b. State `colorpri`/`colorshift` recomputation and caller-table priority are verified in PLS-R07b/PLS-R07c; their supplied inputs do not settle PLS-R07a.
- **PLS-Q04** [unresolved, behavior]: establish valid selector production and actual buffer availability when global random tables are disabled and selector is outside 0..7. No safe fallback or bounds clamp is specified for that domain.
- **PLS-Q05** [unresolved, observation]: reproduce the consumer/preparation and state-priority/caller-choice assertions beyond PLS-V07's present-player return case and compare complete inputs and selected table/output writes. No visible-colour or complete palette-lifecycle confirmation is established.

## 8. Test vectors

### 8.1 Executable vectors

Every case is fresh. Boundary: supplied-input selection or supplied-input preparation stage. Shared inputs: present monster class 10; valid cache; MonStats count 400 with class 10 linking to MonStats2 identity 3; MonStats2 count 5 with selected record present, transformation kind 2. Selector 2, alternate mode off, random collection off. Base bank `B[k][i] = i` for k=0..7 and i=0..255; alternate bank `A[k][i] = (i+1) mod 256`. Global bank `G[k][i] = (i+k+2) mod 256` for k=0..29. All table references are distinct even when contents coincide. For preparation, destination A starts with these stated contents. All seeds: no draw; all Draws: none; outputs include seed preservation. Table selection changes no bank contents. Any case overriding presence/type has no missing unit fields to read.

- **PLS-V01** Cache first (PLS-R01).
  - Inputs: (a) absent unit; (b) present player; (c) monster class 10 with absent cache, selector 38 and random collection enabled; (d) independent present object, missile, item and tile units (type identities 2, 3, 4 and 5 through the record foundations).
  - Seeds: no draw.
  - Draws: none.
  - Trace: reject absent/non-monster before cache; reject missing cache before selector/global lookup.
  - Outputs: each returns no table; no bank writes or assertion.
  - Discriminates: (c) rejects checking global bounds before cache eligibility.
  - runtime_observed: no.

- **PLS-V02** Low-selector class interval (PLS-R02).
  - Inputs: independent class/selector pairs (362,0), (363,0), (364,1), (365,1), (363,1), (364,0); all valid caches and random/alternate modes enabled.
  - Seeds: no draw.
  - Draws: none.
  - Trace: only classes 363 and 364 pass the inclusive class interval.
  - Outputs: no table, B[0], B[1], no table, B[1], B[0] respectively; banks unchanged.
  - Discriminates: rejects applying the global collection or alternate banks to selectors 0/1 and off-by-one class bounds.
  - runtime_observed: no.

- **PLS-V03** Global priority and exact limits (PLS-R03).
  - Inputs: random collection enabled and alternate mode enabled; independent selectors 8, 37, 38, -1, 17.
  - Seeds: no draw.
  - Draws: none.
  - Trace: differences 0,29,30,-9,9; the first, second and fifth satisfy 0..29.
  - Outputs: G[0], G[29], fatal assertion, fatal assertion, G[9] respectively; no bank writes.
  - Discriminates: rejects alternate-bank priority, 8-based off-by-one indexing and inclusive upper bound 30.
  - runtime_observed: no.

- **PLS-V04** Alternate eligibility and linkage (PLS-R04).
  - Inputs: (a) shared inputs, alternate mode on; (b) same with mode off; (c) mode on, transformation kind 1; (d) mode on, class 10 links to -1; (e) mode on, class 10 links to 5; (f) mode on, valid link 3 but selected extended record absent; (g) mode on, valid class-10 cache retained but MonStats count 10, so class 10 equals the record count; (h) shared inputs with mode on and selector 7; (i) shared inputs with mode on and global collection enabled; (j) shared inputs with mode off, global collection enabled and selector 7. For all cases other present extended records have transformation kind 1.
  - Seeds: no draw.
  - Draws: none.
  - Trace: (a)/(h)/(i) follow class 10 to extended record 3, kind 2; (b)..(g)/(j) fail one eligibility condition. The enabled global collection does not intercept selectors 2..7.
  - Outputs: (a) A[2]; (b)..(g) B[2]; (h) A[7]; (i) A[2]; (j) B[7]; no writes.
  - Discriminates: rejects using class as extended identity, treating every nonzero kind as 2, bypassing mode, or accepting count equality.
  - runtime_observed: no.

- **PLS-V05** Sparse translation across all banks (PLS-R05).
  - Inputs: valid shared class/records; translation all zero except T[0]=9, T[1]=7, T[7]=8 and T[255]=3. Every base bank has its shared identity contents.
  - Seeds: no draw.
  - Draws: none.
  - Trace: copy each B into A; index 0 maps to 9, index 1 maps to 7 once, index 7 maps to 8, index 255 maps to 3; all other translation entries are zero.
  - Outputs: for every bank k=0..7, A[k][0]=9, A[k][1]=7, A[k][7]=8, A[k][255]=3; A[k][i]=i for all other i in 0..255. Every B[k][i]=i remains unchanged.
  - Discriminates: rejects writing zero for absent translations, repeated translation (1 would become 8), signed index 255, or processing only one bank.
  - runtime_observed: no.

- **PLS-V06** Preparation failure retains destination (PLS-R05).
  - Inputs: independent class/link/kind overrides: class -1; class 400; class 10 link -1; class 10 link 5; class 10 link 3 absent extended record; class 10 link 3 kind 1. Translation every entry 99.
  - Seeds: no draw.
  - Draws: none.
  - Trace: every case fails a guard before the initial copy.
  - Outputs: all A[k][i]=(i+1) mod 256 retained; all B[k][i]=i retained, for every bank and entry. No other write.
  - Discriminates: rejects copying the base banks before checking eligibility and accepting an index equal to its count.
  - runtime_observed: no.

- **PLS-V07** Present-player rejection (PLS-R01; boundary: supplied-input selection stage).
  - Inputs: present player, class identity 1 and appearance selector 2. This case supplies only unit presence/type plus those retained metadata values; the monster-record, cache and bank defaults above are not inputs to this early-return stage.
  - Seeds: no draw.
  - Draws: none.
  - Trace: the present unit is a player rather than a monster, so return before cache or selector selection.
  - Outputs: no replacement table, no fatal assertion; unit type player/class identity 1 and appearance selector 2 retained.
  - Discriminates: rejects treating a non-monster reached from component rendering as a fatal malformed input instead of returning no table.
  - runtime_observed: yes (1.13c, 2026-10-29, [present-player capture](../research/spec-evidence/draw-palette-shifts.md#runtime-observations)).

- **PLS-V08** State priority, ties and identity zero (PLS-R07b; boundary: state-colour recomputation only).
  - Inputs: present nonlocal monster, base/current graphics selectors both 2 and prior state colour request 17. Four present state records: identity 0 has priority 255/colour 33 and is inactive; identity 1 has priority 3/colour 6 and is active; identity 2 has priority 7/colour 9 and is active; identity 3 has priority 7/colour 10 and is active. Each case is fresh. (a) these inputs; (b) priority of state 1 becomes 8; (c) state 0 becomes active; (d) all four priorities become 0; (e) only states 1 and 3 are active and state 3's colour becomes 0; (f) only state 1 is active and its colour becomes 104; (g) collection count is 0 with no records or active states. Other palette/cache/bank defaults are not inputs to this stage.
  - Seeds: no draw.
  - Draws: none.
  - Trace: scan ascending identities and replace only on strictly greater unsigned priority; discard winning identity 0 after the scan; a selected zero colour remains zero.
  - Outputs: state colour requests (a)..(g) are 9, 6, 0, 0, 0, 104, 0 respectively. Base/current graphics selectors remain 2; active membership and all state definitions remain unchanged.
  - Discriminates: rejects last-wins ties, first-active selection, accepting identity 0, accepting priority 0, skipping a winning zero-colour state, or suppressing colour 104 for a nonlocal monster.
  - runtime_observed: no.

- **PLS-V09** Caller colour override precedes monster fallback (PLS-R07c; boundary: supplied-input component palette choice).
  - Inputs: shared consumer defaults and prepared shared banks S[0]..S[254], all reference-distinct from B/A/G. Fresh cases: (a) state colour request 9; (b) request 0; (c) request 9, valid monster cache, selector 38, global collection enabled; (d) request 9 and absent monster cache; (e) request 255; (f) request 0 and absent monster cache; (g) request 0, valid monster cache, selector 38, global collection enabled; (h) request 1; (i) request 0, valid monster cache, selector 37, global collection enabled. No state recomputation is requested by this vector.
  - Seeds: no draw.
  - Draws: none.
  - Trace: nonzero requests use c-1 before the monster consumer; only request 0 enters the fallback.
  - Outputs: (a) S[8], (b) B[2], (c) S[8] without a fatal assertion, (d) S[8], (e) S[254], (f) no table and normal return, (g) fatal assertion and no selected table, (h) S[0], (i) G[29]. All bank contents, selectors, request values and active membership are unchanged.
  - Discriminates: rejects ignoring the state table, invoking the monster fallback before the override, using c instead of c-1, or limiting a u8 request to a signed-byte range.
  - runtime_observed: no.

### 8.2 Observation vectors

**PLS-O01.** Observe normal component rendering, including the present-player PLS-V07 case, and compare monster selections with PLS-V01 through PLS-V04 using measured cache, selector, modes and linked records; then observe eligible asset preparation and compare every changed entry with PLS-V05. [Procedure and capture limits](../research/spec-evidence/draw-palette-shifts.md#observation-procedure). PLS-V07 is matched; remaining complete consumer/preparation cases are pending.

**PLS-O02.** Observe state-colour recomputation for a nonlocal monster, recording its active states, priorities, colours and request before/after. Compare the reached case with PLS-V08, then follow that request to the component palette-choice stage and compare the selected shared bank or fallback with PLS-V09. [State-priority procedure](../research/spec-evidence/draw-palette-shifts.md#state-priority-observation-procedure). No matching observation is recorded.

## 9. Acceptance and dependencies

Partial acceptance reproduces PLS-V01 through PLS-V09 for PLS-R01 through PLS-R05 and PLS-R07b/PLS-R07c with their supplied stage inputs. It does not establish complete producer integration, original asset bytes, asset availability/lifetime, malformed-domain behavior or runtime execution beyond the recorded PLS-V07 case. Full reproduction additionally settles PLS-Q01, PLS-Q02 and PLS-Q04 and matches applicable PLS-O01/PLS-O02 observations; no partial pass establishes full palette or lighting parity.

**Depends on:** [numeric foundations section 1](core-numeric-and-rng-foundations.md#1-scope) for integer-width, wrap and truncation conventions; GDA-R03a/GDA-R03b/GDA-R03c/GDA-R04a/GDA-R08a/GDA-R13 in [game-data table records](core-game-data-table-records.md) for schema identity, direct byte conversion, ASCII header matching, prepared collection ownership, MonStatsEx linkage and State fields; RLI-R03 in [record layouts](core-record-layouts-and-ids.md) for the active-state test, without consuming its gated state-update bookkeeping. The column bindings do not consume the table dependency's gated complete compiler integration. No clock/quest arithmetic is consumed by the supplied palette stages.

**Depended on by:** no known non-gap plan or executable consumer; component rendering is the original consumer described here.

## 10. Related specs

- [Light map and light radius](draw-light-map-and-light-radius.md): produces scene light rather than indexed monster colour replacement.
- [Room and ambient lighting](draw-room-and-ambient-lighting.md): owns environment and room RGB/intensity consumers.
- [World clock](drlg-world-clock.md): clock and presentation requests; its colour requests do not establish a palette remap.
- [Tainted Sun](quest-tainted-sun.md): quest-controlled eclipse decisions remain in their owner.

## Appendix A. Provenance

The [source manifest](../research/source-manifest.md) identifies original Windows x86 Expansion 1.13c D2Client, D2Common and supporting modules. The [companion](../research/spec-evidence/draw-palette-shifts.md) records instruction, dispatch-table and reconstruction readings on 2026-10-29, the 2026-10-30 schema/compiled-record binding and further loading/producer readings. Original-image instruction listings corroborate the selection/preparation reading; loaded instruction slices matched those originals before entry recording. PLS-V07's present-player return was observed on an owned vanilla 1.13c game on 2026-10-29. The other eight executable vector families remain runtime_observed: no. The named `localBlood` cell has effective text/compiled/published correspondence; complete visible-colour behavior remains unconfirmed.
