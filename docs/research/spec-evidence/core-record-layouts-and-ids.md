# Evidence: record layouts and ids

Spec: [Record layouts and ids](../../specs/core-record-layouts-and-ids.md).
Artifacts: [original source manifest](../source-manifest.md).
Limits: [evidence gaps](../../plans/gaps/core-record-layouts-and-ids.md).
Related records: [numeric and RNG foundations companion](core-numeric-and-rng-foundations.md).

## Identity and verification method

Evidence profile: Lord of Destruction Expansion 1.13c, Windows x86, little-endian.
Artifact ids resolve to original-file fingerprints in the source manifest. All function
descriptions below are analyst-assigned labels, not original symbol names or required APIs.
RVA is relative to the module image base; a PE-aware tool maps it to file/live locations.
Ordinals belong to that identified module/build. A dash means no ordinal is recorded here,
not proof that the function is unexported. Data addresses are distinguished from function entries.

Static readings on 2026-10-12/13 used disassembly and reconstructed C checked against
original machine code. The 2026-10-14 targeted getter reading and foundation migration
added independent instruction listings for ambiguous regions and narrow original-image data
reads. The source manifest establishes correspondence between the analysis images and
fingerprinted originals. The original files are sufficient to repeat the analysis;
unpublished reconstruction files and collection tools are not required sources. Detailed
internal history and original binary fixture encodings are preserved in a separate private
audit. No binaries, whole game tables, private saves or raw captures are published here.

## Restructure (2026-10-15)

This spec was split out of the numeric and RNG foundations spec under the user-approved review
recommendation to split oversized specs. The record structures, unit type codes, the damage record,
the shared name correspondence, the existing-array stat reads with their minimum adjustment, the
state-list lookup/test/set routing and the unit mode setter with its mode-bound traversal moved here,
with their vectors, open questions, gap rows and the records below. Random draws, seeding, arithmetic,
fixed-point/time units and the prepared object animation draw stayed in the numeric foundations spec.
No rule's behavior, label, value or vector result changed; ids were renumbered for this unconsumed
spec. The question covering animation and bookkeeping integration and the gameplay-confirmation
question were each divided by the rules they gate (bookkeeping here, animation preparation in the
numeric spec). Records dated before this split were written against the combined spec; their counts
(vector families, labeled cases, leaf rules) describe the combined spec at that date. Rows moved from
whole-run records are grouped below under the run they belong to; the rest of each run's record stays
in the [numeric companion](core-numeric-and-rng-foundations.md). Bare vector or rule labels inside
quoted finding names name the finding as the reader reported it at that date.
No source, table or gameplay reading was performed for this restructure.

<a id="focused-regeneration-and-evidence-reuse-2026-10-15"></a>
## Focused regeneration and evidence reuse (2026-10-15)

These parts of the numeric foundation regeneration concern the rules now in this spec. The
2026-10-12/13 and 2026-10-14 records remain support for exactly their recorded original 1.13c
slices; their dates, labels and observation limits are unchanged. No new table or gameplay
collection was performed by that pass.

The current lifecycle contract confirms the previously recorded distinction: STL-R07
establishes active/changed membership, while STL-R07a's disguise result, the update queue,
STL-R05 free, STL-R06b attach and STL-R06c replacing-write effects remain gated. Those
gates still delimit RLI-Q02/RLI-Q03; the compatible shared-state reading does not turn
a supplied free or bit-only fixture into complete lifecycle execution.

The hit-recovery damage-class width is already i32 in the damage record (section 3.1); its
corrected full-width interpretation requires no record-width change.

| Question | Reused record and actual date | 2026-10-15 assessment / remaining assertion |
|---|---|---|
| RLI-Q01 | 2026-10-14 condition consumers and semantic mappings | Neutral predicates remain usable as supplied; wider meanings remain with the subsystem owner |
| RLI-Q02 | 2026-10-14 update/disguise inspections; current STL-R07/R07a | Active/changed membership is separate from unresolved disguise and update-queue effects |
| RLI-Q03 | 2026-10-14 setter/attach/free entry inspections; current STL-R05/R06b/R06c | Actual propagation/free/attach effects remain gated under their owner; getter and ordered-request fixtures remain available |
| RLI-Q04 | Historical observations dated 2026-10-12/13; procedures reviewed 2026-10-14 | No gameplay/helper execution was newly observed; all executable vectors remain unconfirmed |

<a id="campaign-reader-passes-2026-10-15"></a>
## Campaign reader passes (2026-10-15)

Three full blind-implementer passes on the combined spec returned MISMATCH 0; counts, gated
groupings and the remaining entries are in the numeric companion's records of
[pass 1](core-numeric-and-rng-foundations.md#campaign-full-reader-pass-1-and-advisory-triage-2026-10-15),
[pass 2](core-numeric-and-rng-foundations.md#campaign-full-reader-pass-2-and-local-order-assertions-2026-10-15)
and [pass 3](core-numeric-and-rng-foundations.md#campaign-full-reader-pass-3-and-completion-2026-10-15).
The advisory entries for the rules now in this spec received these outcomes:

| Pass | Advisory branch | Outcome and concrete reason |
|---|---|---|
| 1 | RLI-R03 null/storage guards | No change: RLI-V04 (b)/(i) show the zero state-test result, RLI-V05 (b) a null lookup result, and (c) no stored bit effect with a missing base while update is still requested. Additional null/storage permutations reuse those guard/return/request recipes; their distinct preconditions remain explicit in RLI-R03 |
| 1 | RLI-R03 SetState on another unit kind | No change: RLI-V05 (e)/(j)/(l) demonstrate enable/clear/unchanged-value forwarding. RLI-R03 explicitly applies no unit-type filter; repeating the same call/bit operation for another fixed type adds no new result recipe |
| 1 | RLI-R04b missing base or retained extended list | No change: RLI-V07 (b) demonstrates early return without visits, while (a) demonstrates reloading the current head. A retained flagged extended list can repeat indefinitely as the rule and section 6 state; it is not a successful finite fixture, and no new dependency mutation or termination default is supplied |
| 2 | RLI-R03 FindStateList null unit, missing list/storage and multiple matches | No change: fatal null-input is NRF-V04 (c); null lookup is RLI-V05 (b); primary-before-secondary duplicate selection is RLI-V05 (a), with secondary fallback (g). The rule explicitly returns the first matching list and states its distinct guards; extra absent-storage/duplicate fixed identities add no new traversal/result recipe |
| 2 | RLI-R03 HasState null/type/storage/range guards | No change: RLI-V04 (b)/(i) demonstrate rejected types returning 0, (c)/(g)/(h) range rejection, (e)/(f) accepted types and bit masks. Other guard failures share that zero-result recipe without changing the explicit per-operation guards |
| 2 | RLI-R03 SetState valid null unit / type-independent routing | No change: RLI-V05 (c) demonstrates the bit request with no stored effect followed by update; (e)/(j)/(l) demonstrate set, clear and unchanged nonzero argument. A valid null unit retains the stated update fault/integration boundary, not a successful default. RLI-R03 explicitly has no type filter; repeating a type binding adds no new operation recipe |
| 2 | RLI-R04b missing base list | No change: RLI-V07 (b) already demonstrates early return with no visits/requests; the missing-base guard selects that same recipe |
| 2 | RLI-R04b retained flagged extended-list restart | No change: RLI-V07 (a) explicitly observes reloading the current head. A retained flagged extended list can make that loop nonterminating, as RLI-R04b and section 6 state. No extra fixture mutation or termination policy is invented |
| 3 | RLI-R03 null unit, missing storage and negative SetState on null unit | No change: RLI-V04 (b)/(i) supply the zero state-test guard result, RLI-V05 (b) the null lookup result, and (c) the no-stored-bit-effect plus update request. RLI-V05 (k) directly checks the negative-state no-request guard; RLI-R03 applies it before any valid-state bit/update requests, including on a null unit. Valid null-unit update handling retains its explicit fault/integration boundary rather than a successful default |
| 3 | RLI-R02a positive count without matching identity | No change: RLI-V02 (f) directly supplies positive count 2 with entries (34,1)/(36,0), requests (34,0), and expects 0 with no fallback |
| 3 | RLI-R04a type 2/3/4 equal and changed modes | No change: RLI-V06 (b)/(c)/(g)/(i) exercise ordinary equality/change routing, unrestricted requested mode and the monster-only exception's limit; (d)/(h) isolate type 5 refusal. Types 2/3/4 use the ordinary path, with no additional type-specific local arithmetic or request recipe |
| 3 | RLI-R04b missing base / flagged negative state | No change: RLI-V07 (b) supplies early return without visits/requests; (c) rejects placing free under the strict lower state bound, and (d) rejects placing it under the upper bound. A negative state follows the same lower-bound failure and independent free recipe. The rule still explicitly requires a base list before traversal |

HARDEST entries of these passes for this spec were the mode-bound head restart (RLI-V07, passes 1
and 2) and the minimum-condition matrix (RLI-V03, pass 3); each received no change because the
vector already states its observed visit order or condition matrix explicitly.

## Independent review spot check (2026-10-15)

An independent review read the following original 1.13c instructions before comparing them with
the rules: D2Common `0x38B70`, `0x38A80`, `0x382B0` (RLI-R02a/RLI-R02b); `0x33CD0`, `0x37DB0`,
`0x34510`, `0x38730`, `0x38520`, `0x37F80` (RLI-R03); `0x33920` (RLI-R04a); `0x3AE60`
(RLI-R04b). Undecoded byte runs at `0x38A80`, `0x3AE60` and `0x38730` were decoded by hand from
their opcode bytes. Every arithmetic, guard and order statement of those rules was confirmed. The
mode setter `0x33920` tail-requests D2Common `0x335F0`, whose prepared type-2 slice is NRF-R13;
that is a call-site anchor, not a complete frame-loop chain. The full review record, including the
numeric rules it read, is in the
[numeric companion](core-numeric-and-rng-foundations.md#independent-review-spot-check-and-corrections-2026-10-15).

## Original function references

<a id="stat-reader-verification-2026-10-14"></a>
### Stat-reader verification (2026-10-14)

This targeted reading owns RLI-R02a/RLI-R02b existing-array retrieval, not array construction,
attached-list aggregation, replacing writes or gameplay. The recorded D2Common analysis
hash matches the manifest's `LOD113C-D2COMMON` fingerprint.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2COMMON | `0x38B70` | 10973 | Unit stat getter | RLI-R02a | Entry-to-return disassembly; unit/list/id guards |
| LOD113C-D2COMMON | `0x38A80` | — | Selected-array getter/minimum | RLI-R02a, RLI-R02b | Independent entry-to-return instruction listing |
| LOD113C-D2COMMON | `0x382B0` | — | Exact signed-key search | RLI-R02a | Checked reconstruction; complete 53-byte helper |
| LOD113C-D2COMMON | `0x39140` | 10680 | List stat getter | RLI-R02a | Entry-to-return disassembly; own-array route |
| LOD113C-D2COMMON | `0x38B00` | — | Own-array getter/minimum | RLI-R02a, RLI-R02b | Entry-to-return disassembly |
| LOD113C-D2COMMON | `0x73330` | — | Stat-definition descriptors | RLI-R02b, RLI-R01a | Selected original instructions/strings checked independently |
| LOD113C-FOG | `0x26798` | 10259 | Bit-mask data export | RLI-R03, RLI-R02b | Import/export resolution and narrow image read; data address, not function |

Root: mitigation at D2Game RVA `0xDC0B0` calls D2Common #10973 for flat reduction,
resistance and absorb; poison/burn reapplication calls #10680. Their output-producing
callers establish getter identity beyond a name match.

| Semantic term | Original mapping |
|---|---|
| unit.statList | Unit `+0x5C` |
| usesResolvedStats / hasStateStorage | StatList `+0x10` bit 31, one original extended-list condition |
| ownStats | StatList `+0x24` pointer, signed 16-bit count `+0x28` |
| resolvedStats | StatList `+0x48` pointer, signed 16-bit count `+0x4C` |
| associated owner | StatList `+0x44`, owner type dword `+0x00` |
| exact identity | Signed i32 key `(id<<16)+u16(layer)`; key then signed value in each pair |
| definitions | DataTables `+0xBCC` array, `+0xBD4` count, stride `0x144` |
| fMin / MinAccr / ValShift | Record `+0x04` bit 10 / signed dword `+0x2C` / byte `+0x18` |

At `0x38A80+0x00..0x13`, bit 31 selects resolved or own array; call `+0x15` reaches
`0x382B0`, with array in EAX and key in EDI. Missing result/address returns zero at
`+0x1A..0x2B`. Stored value is read at `+0x32..0x37` even with null optional metadata.
`+0x39..0x5F` tests fMin, extended list, associated owner and type 0/1. `+0x61..0x6D`
compares against **unshifted** MinAccr and shifts only the replacement, modulo 32 bits.
Return is `+0x71`. `0x38B00` always selects ownStats, then follows the same minimum path.
Wrapper null-unit result is zero. Its record guard tests the computed address; valid
loaded-table inputs do not promise safety for corrupt table bases or entry memory.

No RNG, list walk, summation, general shift or maximum clamp occurs here. Correction
dated 2026-10-14: the earlier aggregation reading was incorrect. One existing array
supplies one value, optionally replaced by shifted MinAccr. Associated list owner,
not the unit argument's type, controls the minimum. Producer behavior remains separate.

D2Common import slot RVA `0x890B0` resolves Fog #10259 / RVA `0x26798`; array byte
`+8` is 4. Descriptor instructions `0x73330+0xEE..0x13D` map `valshift` to `+0x18`
and `minaccr` to `+0x2C`. Descriptor/string RVAs `0x73632/0x8DF9C` identify `fmin`,
conversion `0x1A`, bit 10 at destination `+4`. Lowercase descriptor names correspond
to archive headers; narrow compiled values corroborate these selected mappings below.

<a id="states-and-mode"></a>
### States and mode

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2COMMON | `0x38520`, `0x37F80` | 10871 for `0x38520` | State-list wrapper/finder | RLI-R03 | Disassembly; primary then secondary chains |
| LOD113C-D2COMMON | `0x33CD0`, `0x37DB0` | 10494 for `0x33CD0` | State-test wrapper/bit read | RLI-R03 | Disassembly; guards and bit result |
| LOD113C-D2COMMON | `0x34510`, `0x38730` | 10945 for `0x34510` | State-set routing/bit writes | RLI-R03 | Disassembly/independent listing 2026-10-14; disguise/update integration gated |
| LOD113C-D2COMMON | `0x1770`, `0x33C20` | — | State-record/disguise intersection | RLI-Q02 | Checked reconstruction / disassembly; dependency inspection only |
| LOD113C-D2COMMON | `0x33920` | 11090 | Mode setter | RLI-R04a | Disassembly; guard, equal exception, ordered writes/requests |
| LOD113C-D2COMMON | `0x3AE60`, `0x3A9F0` | 10196 for `0x3AE60` | Mode-bound list traversal/free | RLI-R04b | Independent traversal listing / checked free reconstruction; free effects dependency-owned |
| LOD113C-D2COMMON | `0x665C0`, `0x2FE10` | 10158 for `0x665C0` | Update/region queue context | RLI-Q02 | Selected disassembly/listing inspection, not full network contract |

Relevant branch/output rationale:

- State chains start at Unit `+0x5C`; primary/secondary heads are StatList `+0x3C/+0x40`,
  state `+0x14`, expiry `+0x18`, bit buffer `+0x58`. Same bit-31 condition enables
  extended-list state storage and resolvedStats. Valid SetState routes its bit request
  and unconditional update request; full disguise/update effects stay RLI-Q02.
- Equal mode returns 0; only equal monster mode 1 skips update/modeChanged. No equal
  case processes lists/resets. Changed mode requests update, marks changed, stores mode,
  processes lists, then resets. At D2Common `0x3AEBB`, traversal reloads primary head
  after each processed flagged list; unflagged nodes follow saved successor. A retained
  flagged extended list may repeat indefinitely; RLI-V07 supplies finite removal.

<a id="stat-table-verification-2026-10-14"></a>
## Named game-data records (2026-10-14)

Historical reads used the running game's archive loader on the reported vanilla LoD
1.13c edition/build and profile family, with no known overrides reported. Other modded
profiles were excluded. The `LOD113C-GAME` executable hash identifies the recorded
session executable; it does not independently establish an unmodified archive set,
because modded profiles can share `Game.exe`. Supplying physical MPQs, archive precedence,
effective-file SHA-256 hashes and a complete override inventory were **not captured**.
Code hashes do not fingerprint game data. Reproduction must record its own effective
files/overrides before comparing these narrow cells, under the [manifest's data-identity
limits](../source-manifest.md#game-data-identity-and-limits). No gameplay action, memory
write, hook or save edit was performed.

### Stat minimum metadata

Virtual file `data/global/excel/ItemStatCost.txt`: requested offset 0 length 65536;
returned all 40109/40109 bytes. TSV columns selected by header, rows by Stat and explicit
decimal ID. IDs agree with zero-based data-row position excluding header in this profile;
that correspondence does not prove editable IDs control compiler assignment.

| Stat / ID | Columns | Raw cells | Interpretation / supported claim |
|---|---|---|---|
| hitpoints / 6 | fMin; MinAccr; ValShift | empty; empty; `8` | Shift metadata exists without minimum flag; no general read scaling |
| maxhp / 7 | fMin; MinAccr; ValShift | `1`; `1`; `8` | RLI-R02b/RLI-V03 metadata; owner/value setup is supplied fixture |
| normal_damage_reduction / 34; magic_damage_reduction / 35 | fMin; MinAccr; ValShift for each | All cited cells empty | Narrow compiled correspondence below |
| item_absorbfire / 143; item_absorblight / 145; item_absorbmagic / 147; item_absorbcold / 149 | fMin; MinAccr; ValShift for each | All cited cells empty | No minimum adjustment on these loaded flat-absorb definitions |

Virtual file `data/global/excel/ItemStatCost.bin`: total 116320 bytes. Successful
requested/returned ranges (offset:length) 0:4, 1948:648, 11020:4212, 46012:3888.
Header 359 records; explicit compiled IDs match cited row indices. Record start
`4+ID*324`; tested bit 10, signed minimum `+0x2C`, shift byte `+0x18`:

| Compiled ID | fMin | MinAccr | ValShift |
|---|---|---|---|
| 6 | false | 0 | 8 |
| 7 | true | 1 | 8 |
| 34, 35, 143, 145, 147, 149 (each) | false | 0 | 0 |

Coverage suffices for these cells, not the full binary table. It proves this correspondence,
not general parser defaults, live loaded data, physical MPQ origin or gameplay.

### Difficulty divisors and identities

Virtual file `data/global/excel/difficultylevels.txt` read fully 0..586/586 bytes;
columns selected by header, rows by Name. Narrow retained cells:

| Row Name | Column | Raw cell | Interpretation |
|---|---|---|---|
| Nightmare | MonsterColdDivisor | `2` | Signed cold divisor 2 |
| Nightmare | MonsterFreezeDivisor | `2` | Signed freeze divisor 2 |
| Hell | MonsterFreezeDivisor | `4` | Signed freeze divisor 4 |

These divisor cells support the status appliers CSA-R05 and CSA-R08; the status application companion's link reaches
them through a pointer anchor in the numeric foundations companion.

States text coverage was 20920/20920 bytes, 185 data rows, with zero-based identity
correspondence checked excluding header. ItemStatCost text above contained 359 rows.
RLI-R01b consumes specific stat/state/mode/ItemTypes/monster identity records in the
[game-data evidence companion](core-game-data-fields.md), and RLI-R01a consumes its
per-table descriptor/getter records. This named dependency does not assert that the
numeric pass reread every descriptor. MonStats cached/text tail discrepancies remain
recorded there; no universal text-to-compiled class correspondence or custom remapping
capability follows. These are data-only records, not gameplay confirmation.

<a id="foundation-migration-source-record-2026-10-14"></a>
## Foundation migration source record (2026-10-14)

These are the rows of the foundation migration attempt for the questions now in this spec. Ambiguous
instruction regions were independently checked before bounded labels were assigned. Full gameplay
paths/dependencies were not promoted. The remaining rows are in the
[numeric companion](core-numeric-and-rng-foundations.md#foundation-migration-source-record-2026-10-14).

| Question | Original-artifact attempt | Remaining boundary |
|---|---|---|
| RLI-Q01 | Tested predicates/masks retained with neutral names | Wider gameplay meanings not invented |
| RLI-Q02 | Update/region and disguise slices inspected | Complete dependency effects |
| RLI-Q03 | Attach/write entries `0x3A7D0/0x3A280/0x3A740` and callees inspected | Actual lifecycle-owned propagation remains unverified |
| RLI-Q04 | Historical readings reviewed | No gameplay campaign performed |

These bounded searches are not absence proofs. Supplied-input arithmetic/request vectors
do not close integrations; exact blocked assertions remain in the gap file.

## Additional public correction rationale

- Equal-mode exception and head-restart traversal were clarified from their original
  branches, preserving existing results.

<a id="focused-evidence-reconciliation-2026-10-14"></a>
## Focused evidence reconciliation (2026-10-14)

- **RLI-R01b / named game-data records:** profile wording now distinguishes the reported
  edition/build and profile family from the session executable hash, following the
  manifest. A complete override inventory was not captured. This clarifies the evidence
  limits without changing any retained named cell, interpretation or data label.

<a id="split-reader-pass-1-2026-10-14"></a>
## Split revision: blind-implementer pass 1 (2026-10-14)

Entries of this full run for the rules now in this spec; counts and the other entries are in the
[numeric companion](core-numeric-and-rng-foundations.md#split-reader-pass-1-2026-10-14). Source reread
(2026-10-14, original instructions; independent listings of the region already existed, so the decoded
bytes were checked against them):

- D2Common `0x3AE60`: for a flagged list the state test (`0x3AE8C..0x3AE9F`) guards only the state-bit
  clear on the unit (call to `0x38730`) and the update request (call to `0x665C0`); the free at
  `0x3AEB6` is reached after that block whenever list flag bit 31 is clear.

| Entry | Outcome |
|---|---|
| GUESS 5-6 (R14b target and nesting) | Spec edit: unit state-bit clear and update under the state test; free independent (`0x3AE60`) |
| DISCRIM 3 (state types) | RLI-V04 (e), (f) |
| DISCRIM 4-5 (mode setter) | RLI-V06 (h), (i) |
| DISCRIM 11 (secondary chain) | RLI-V05 (g) |

All earlier expected values are unchanged.

<a id="split-reader-pass-2-2026-10-14"></a>
## Split revision: blind-implementer pass 2 (2026-10-14)

Entries of this full run for the rules now in this spec; counts and the other entries are in the
[numeric companion](core-numeric-and-rng-foundations.md#split-reader-pass-2-2026-10-14). Source rereads
for the GUESS entries (2026-10-14, original instructions):

- D2Common `0x34510` (export 10945): returns with no call when the signed state is below 0 or not
  below the state count; otherwise calls the bit write `0x38730` and then `0x665C0`. There is no
  unit-type test. `0x38730` returns without a write for a null unit, a null list or a list without
  flag bit 31.
- D2Common `0x33920`: on a mode change the update request (`0x3393D`) precedes the changed-flag and
  mode writes; on an equal mode the update request precedes the changed-flag write.
- D2Common `0x3AE60`: the expiry flag is cleared at `0x3AEC5`, after the walk ends.

| Entry | Outcome |
|---|---|
| GUESS 1 (SetState bounds, type filter) | RLI-R03 states signed bounds and no type test; RLI-V05 and RLI-V07 state type 1 |
| GUESS 5 (call-time fields) | RLI-R04a and RLI-R04b state that requests observe fields written so far; RLI-V06 (c), (f) and RLI-V07 (a) assert the observed values |
| GAP 5 (unit types) | RLI-V07 and RLI-V05 state type 1 |
| DISCRIM 1 (state bounds) | RLI-V04 (g), (h) |
| DISCRIM 5 (call-time snapshots) | RLI-V06 (c), (f); RLI-V07 (a) |
| DISCRIM 11 (owner type 3) | RLI-V03 (b) |

Every suggested rejecting input was recomputed under both behaviors before it was added. Earlier
expected values are unchanged.

<a id="split-reader-pass-3-2026-10-14"></a>
## Split revision: blind-implementer pass 3 (2026-10-14)

Entries of this full run for the rules now in this spec; counts and the other entries are in the
[numeric companion](core-numeric-and-rng-foundations.md#split-reader-pass-3-2026-10-14). Source rereads
for the GUESS entries (2026-10-14, original instructions):

- D2Common `0x38520` (export 10871) returns null for a null stat list or a list without flag bit 31,
  then calls `0x37F80`, which returns null for state 0 and otherwise compares the state by equality
  along the primary chain (when flag bit 31 is set), then the secondary chain. There is no range test.
  The lifecycle contract's summary of this lookup omits the null-list and flag-bit-31 tests; that is
  reported to the coordinator for its owner, not changed here.
- D2Common `0x38730`: a nonzero value argument ORs the state bit; zero ANDs its complement. The
  mode-bound traversal passes 0.
- D2Common `0x33CD0`: state 0 passes the lower bound (`test` and `jl`), so `HasState(0)` reads bit 0.

| Entry | Outcome |
|---|---|
| GUESS 1 (identity reduction) | RLI-R02a defines modulo 65536 as the unsigned low 16 bits |
| GUESS 2 (lookup bounds) | RLI-R03: no range test in FindStateList; STL discrepancy reported |
| GUESS 3 (set argument) | RLI-R03 defines the `on` argument and the clear request's 0 |
| DISCRIM 3 | RLI-V04 (i) |
| DISCRIM 4 | RLI-V04 (j); state 0 is legal for HasState |
| DISCRIM 5 | RLI-V07 (c), (d) |
| DISCRIM 6 | RLI-V06 (f) observes mode 2 at the walk's requests |
| DISCRIM 10 | RLI-V02 (g) |
| DISCRIM 11 | RLI-V03 (b) owner type 2 |

Each added input was recomputed under both the claimed and the rejected behavior. Earlier expected
values are unchanged.

<a id="split-delta-check-1-2026-10-14"></a>
## Split revision: delta check 1 (2026-10-14)

Entries of this delta check for the rules now in this spec; counts and the other entries are in the
[numeric companion](core-numeric-and-rng-foundations.md#split-delta-check-1-2026-10-14). Source reread
(2026-10-14): D2Common `0x34510` pushes its value argument unchanged to the bit write `0x38730`, which
tests it for nonzero.

| Entry | Outcome |
|---|---|
| GUESS 1 (SetState argument) | RLI-R03: the argument is carried unchanged; RLI-V05 (l) |
| GAP 2 (unstated observed fields) | Preamble: only stated observed fields are asserted; RLI-V06 (f) unchanged |
| DISCRIM 1 | RLI-V05 (h), (i) |
| DISCRIM 2 | RLI-V05 (j) |
| DISCRIM 3 | RLI-V05 (k) |

Each added input was recomputed under both behaviors. Earlier expected values are unchanged. Delta
checks 2 and 3 had no entries for these rules; the split-revision blind-implementer pass completed
with delta check 3.

## Historical reader and preservation checks

The targeted getter review (2026-10-14) and the historical whole-foundation reader passes covered the
stat-read, state and mode rules as part of the combined spec; their counts and dispositions are in the
[numeric companion](core-numeric-and-rng-foundations.md#reader-and-preservation-checks-2026-10-14).
A blind-implementer pass of this spec as a separate document has not been run since the restructure.
