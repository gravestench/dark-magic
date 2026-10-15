# Evidence: numeric and RNG foundations

Spec: [Numeric and RNG foundations](../../specs/core-numeric-and-rng-foundations.md).
Artifacts: [original source manifest](../source-manifest.md).
Limits: [evidence gaps](../../plans/gaps/core-numeric-and-rng-foundations.md).

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
fingerprinted originals. Tables below distinguish fully read local paths from partially
inspected candidates and dependency calls. The original files are sufficient to repeat
the analysis; unpublished reconstruction files and collection tools are not required sources.

The provenance rewrite on 2026-10-14 changed citation presentation only. It did not
renew code verification, change labels or claim gameplay confirmation. Detailed internal
history and original binary fixture encodings were preserved in a separate private audit.
The linked spec contains current inputs, rules, vectors, acceptance and gates. No binaries,
whole game tables, private saves or raw captures are published here.

## Focused regeneration and evidence reuse (2026-10-15)

The existing contract, companion, gap attempts, original artifact manifest and current
dependency contracts were assessed before new source reading. The 2026-10-12/13 and
2026-10-14 records remain support for exactly their recorded original 1.13c slices;
their dates, labels and observation limits are unchanged. Original-image fingerprint
correspondence is the one already established in the manifest. Effective table-profile
identity is not inferred from those code hashes. No new table or gameplay collection
was performed by this numeric pass.

The current lifecycle contract confirms the previously recorded distinction: STL-R07
establishes active/changed membership, while STL-R07a's disguise result, the update queue,
STL-R05 free, STL-R06b attach and STL-R06c replacing-write effects remain gated. Those
gates still delimit NRF-Q17/NRF-Q20; the compatible shared-state reading does not turn
a supplied free or bit-only fixture into complete lifecycle execution.

The reviewed item-event contract supplies narrower reusable records than a universal
callback total: IEP-R07's dispatcher has no direct draw; IEP-R13b takes one owner modulo
100 draw only after its stated guards; IEP-R14 to IEP-R18 have no direct callback draw.
Their downstream apply/reaction/state/write operations keep their own contracts and
gates. These records retain their 2026-10-12/13 and targeted 2026-10-14 dates in the
[item-event companion](combat-item-event-triggers-and-procs.md#original-function-references).
They refine the boundary of NRF-Q10 without establishing unread callbacks, original
item-to-handler bindings, later reanimation or complete callback/callee draw totals.

Current damage-pipeline, shield/block, knockback and hit-recovery owner contracts agree
with NRF-R04a's reached-site seed owners. The hit-recovery damage-class width is already
i32 in section 3.2 here; its corrected full-width interpretation requires no numeric
foundation arithmetic or record-width change. Consumer eligibility and complete sequence
remain with those owners. These dependency comparisons are not fresh binary verification.

| Question | Reused record and actual date | 2026-10-15 assessment / remaining assertion |
|---|---|---|
| NRF-Q02 | 2026-10-14 complete unit-seeder/sampled-time chain | Optional source-seed semantics clarified below; actual null-source gameplay creator remains outside the supplied-clock helper |
| NRF-Q03 | 2026-10-14 rate helper, recompute, call/import/pointer searches | Same indirect startup/live-rate question remains; repeating the same negative direct-call query adds no evidence |
| NRF-Q04 | 2026-10-14 getters, direction writer/helper and constant-table read | Supplied getter/velocity results remain verified; no complete static-position or direction producer contract was added |
| NRF-Q09 | 2026-10-14 condition consumers and semantic mappings | Neutral predicates remain usable as supplied; wider meanings remain with the subsystem owner |
| NRF-Q10 | 2026-10-14 numeric dispatcher/slot inspection; item-event records above | Known direct-draw slices are reusable; unread callback/callee totals and full data bindings remain outside this primitive contract |
| NRF-Q11 | 2026-10-14 reached monster/searcher listings; current HRC-R15 | Complete mode-event eligibility/additional draws remain gated; no full zero-draw event is inferred |
| NRF-Q13 | 2026-10-14 one-step game stage and resource entry/call inspection | Incoming seed and resource initialization remain outside the supplied-seed stage |
| NRF-Q17 | 2026-10-14 prepared animation/update/disguise inspections; current STL-R07/R07a | Active/changed membership is separate from unresolved disguise, update-queue and complete animation effects |
| NRF-Q19 | 2026-10-14 player-loader candidate inspection | Actual player-load/reseed root remains unverified; creation's player skip remains available |
| NRF-Q20 | 2026-10-14 setter/attach/free entry inspections; current STL-R05/R06b/R06c | Actual propagation/free/attach effects remain gated under their owner; getter and ordered-request fixtures remain available |
| NRF-Q21 | Historical observations dated 2026-10-12/13; procedures reviewed 2026-10-14 | No gameplay/helper execution was newly observed; all executable vectors remain unconfirmed |

The remaining questions match the previous searches' source identities and assertion
boundaries. They do not justify repeating identical failed queries or reading unrelated
gameplay paths. This pass added only the following missing source-argument/return-scope
check, rather than claiming all gaps closed.

### Seed-source and initializer contract check (2026-10-15)

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2COMMON | `0x36740`, `0x36750` | 10614, 10936 respectively | Supplied-value and initial seed writes | NRF-R03a, NRF-R03b | Complete C reconstruction checked against original bytes; seed writes only, not a universal caller-return claim |
| LOD113C-D2GAME | `0xE0110` | — | Unit-seeding source argument | NRF-R03a, NRF-R03b | Complete local seeder disassembly read; optional source-pair test, direct source step and tail request to #10614 |
| LOD113C-D2GAME | `0xE1D90` | — | Generic creation source-seed handoff | NRF-R17a | Entry through seeding slice, specifically `0xE1E72..0xE1E99`; source handoff corroborated by original-image decode `0xE1E38..0xE1E9C`; later creator paths not newly verified |

The initializer leaves write `{supplied value,666}` and `{1,666}`. Their matched
reconstructions have void prototypes; that alone does not prove every caller ignores an
incidental machine-register value. NRF-R03b therefore specifies writes and makes no
numeric return assertion. It no longer uses the undefined phrase “argument/return identity.”

At `D2Game.dll+0xE0110`, the optional source in ESI is dereferenced directly as two seed
words; it is not a Unit address. At `0xE1E74`, the generic creator forms `Game+0xD0`,
passes it to the seeder at `0xE1E7C`, and does not consume a numeric seed-initializer
result before its item-type test. The child seed remains `Unit+0x20`, initial copy
`Unit+0x28`. Thus the semantic input is an optional source seed: a parent's seed is one
valid source, and creation uses the game seed. This corrects the earlier “parent unit”
wording without changing any source owner, step, copy or vector value. The seeder's
initializer calls are #10936 at D2Game thunk `0xA616` and #10614 at thunk `0xA628`;
their resolved targets are the D2Common leaves above.

### Contract preservation and self-review

The private before snapshot preserves all three documents with SHA-256 records. No id,
old case, numeric input, draw owner/order, expected value, evidence label or integration
gate was removed or renumbered. The vector preamble now assigns illustrative `on=1`
explicitly and states that local NRF callees execute, while supplied dependency operations
apply only their stated fixture effects. The ordered parent-selection/seed-write log is
an observation of local execution, not an injected parent-selection operation. Arithmetic
and return/input inventories are explicit, including the unsigned quotient result.

Self-review checked every verified rule against its preserved vectors and execution
boundary, the source-argument correction against the source slice above, and all eleven
remaining gap rows against compatible attempts and current dependencies. No additional
result-distinct missing numeric branch was identified in this pass. Historical reader
counts below apply to their historical revisions. A fresh reader must assess this
regenerated contract before its review is complete. Status remains draft; gameplay
confirmation and user approval are not inferred.

Mechanical and preservation checks on 2026-10-15 found identical defined id sets, all
31 leaf rule-label pairs unchanged, all 25 executable-vector blocks unchanged, and all
three observation definitions unchanged. The spec check reported 0 errors and 46 warnings;
all 46 warnings are inherited historical cross-references in the status-application
companion. None points to an id added, removed or changed by this pass. The three documents
owned by this numeric regeneration have 0 errors and 0 warnings. Resolving those historical
status references belongs to that companion's regeneration; they were left unchanged here.

### Campaign full reader pass 1 and advisory triage (2026-10-15)

A fresh smaller reader used the main contract and dependency contracts, with no additional
access limits, and independently recomputed 25 vector families and 159 labeled cases.
Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 12, COVERAGE 11, HARDEST 3.
The twelve grouped gated entries were NRF-Q02, NRF-Q03, NRF-Q04, NRF-Q10, NRF-Q11,
NRF-Q13, NRF-Q17, NRF-Q19, NRF-Q20, NRF-Q21, NRF-R04c and NRF-R09b. Other explicit
limits, including NRF-Q09, remain in force; this grouping does not remove a question.

Every advisory coverage entry received the following outcome. The one new recipe is
derived from the already verified NRF-R07 supplied-value arithmetic, not a fresh source
reading or a newly asserted failure default. All retained cases and expected results
remain unchanged.

| Advisory branch | Outcome and concrete reason |
|---|---|
| NRF-R02 negative RandInRange modulus | No change: NRF-V02 (b) demonstrates the signed negative-modulus no-step guard; (d) demonstrates the offset-form guard returns the supplied offset without a step. Combining those same operations adds no distinct arithmetic recipe |
| NRF-R05 negative nonzero divisor | No change: NRF-V04 (a) and NRF-V13 (c) distinguish signed truncation from floor division; NRF-V04 (c), (d), (h), (j) and (l) demonstrate the divide-first branch recipes. A negative divisor uses the stated signed comparison/division rules rather than an additional saturation or absolute-value operation |
| NRF-R07 null unit and zero denominator | Spec edit: NRF-V05 (l) supplies a present value -120 so the nonzero-value branch reaches denominator 0 and faults, rather than the missing/zero-value return shown in (g)/(h). The null-unit fatal/fault recipe is already exercised by NRF-V11 (c); NRF-R07 explicitly states its distinct null-unit precondition |
| NRF-R09a fps 0 | No change: the zero-divisor fault recipe is now NRF-V05 (l), while NRF-V14 (a) to (d) test rate arithmetic/recompute. NRF-R09a already states fps 0 faults. No claim about fields retained or written before that fault is added; fault fixtures assert only abnormal completion |
| NRF-R10b null path | No change: NRF-V11 (c) supplies the fatal null-input recipe; NRF-V29 (g) exercises this getter's independent extraction result. NRF-R10b states the null-path assertion, without a post-fault state guarantee |
| NRF-R11 tick wrap | No change: NRF-V27 (c) demonstrates low-32-bit increment wrap, NRF-V22 (a) the five-tick ramp and (c) the no-acceleration unchanged-tick branch. The tick uses the same explicit low32 increment and signed comparison, not a new clock policy |
| NRF-R12a null/storage guards | No change: NRF-V15 (b)/(i) show the zero state-test result, NRF-V33 (b) a null lookup result, and (c) no stored bit effect with a missing base while update is still requested. Additional null/storage permutations reuse those guard/return/request recipes; their distinct preconditions remain explicit in NRF-R12a |
| NRF-R12a SetState on another unit kind | No change: NRF-V33 (e)/(j)/(l) demonstrate enable/clear/unchanged-value forwarding. NRF-R12a explicitly applies no unit-type filter; repeating the same call/bit operation for another fixed type adds no new result recipe |
| NRF-R14b missing base or retained extended list | No change: NRF-V31 (b) demonstrates early return without visits, while (a) demonstrates reloading the current head. A retained flagged extended list can repeat indefinitely as the rule and section 6 state; it is not a successful finite fixture, and no new dependency mutation or termination default is supplied |
| NRF-R17a unavailable MonStats record | No change: NRF-V17 (d)/(f) demonstrate the common pre-allocation rejection result, unchanged seed and no seeder request. Missing record reaches that same refusal; no new allocation/seed recipe is added |
| NRF-R18 prepared mode >=8 | No change: NRF-V11 (c) provides the fatal-assertion recipe, while NRF-V32 tests the valid prepared object branch. The upper-bound assertion is explicit in NRF-R18; negative indexing and full animation preparation remain outside its valid-record slice. No per-mode record, preceding write or post-fault field value is invented |

HARDEST entries were NRF-V04's ordered mulDiv branches, NRF-V27's unsigned FILETIME
wrapping and NRF-V31's head restart. Each received no change: its explicit operand order,
width/rounding trace or observed visit log already resolves the difficulty. No MISMATCH,
GUESS, GAP or DISCRIM required a separate resolution. NRF-V05 (l) expands the contract's
executable case set to 160 labeled cases within the same 25 vector families, so a fresh
full pass is required before the campaign reader review is complete. The previous
mechanical/preservation record above describes the pre-advisory revision; all old cases
and values remain preserved after this addition.

### Campaign full reader pass 2 and local-order assertions (2026-10-15)

A fresh smaller reader used only the main contract and its dependencies. Counts:
MISMATCH 0, GUESS 0, GAP 0, DISCRIM 2, GATED 22, COVERAGE 12, HARDEST 3.
The 22 grouped gated entries were NRF-R04c, NRF-R09b, NRF-Q02, NRF-Q03, NRF-Q04,
NRF-Q09, NRF-Q10, NRF-Q11, NRF-Q13, NRF-Q17, NRF-Q19, NRF-Q20, NRF-Q21 and the
nine supplied-input stage vectors NRF-V16, NRF-V17, NRF-V20, NRF-V24, NRF-V28,
NRF-V30, NRF-V31, NRF-V32 and NRF-V33. Their verified local assertions remain
available; this grouping does not gate the entire vector or remove any existing limit.

| Discrimination finding | Outcome |
|---|---|
| NRF-V24 could pass with secondary initial write after the game step | Spec edit: observed local assertions require secondary `{1,666}` write, second game step, initial-copy write, then final secondary write. All final pairs/copies and both game steps are unchanged |
| NRF-V28 could pass with local initial write after the game step | Spec edit: observed local assertions require local `{1,666}` write, game step, game-copy write, final local write, then resource request. All final values/request arguments are unchanged |

Both edits observe the actual NRF-R17b/NRF-R03d local writes and steps already stated
in the rules. They introduce no supplied implementation of those local rules and no
source of seed values beyond the existing fixture inputs. Their source support is the
compatible seeding and game-stage readings dated 2026-10-14, not new binary analysis.

The reader grouped twelve coverage entries, with several guard subbranches. Every
reported branch is addressed below; no additional case was needed.

| Advisory branch or grouped subbranches | Outcome and concrete reason |
|---|---|
| NRF-R07 table flag clear / zero-value bypass | No change: all five original constant-table flags are 1, so flag-clear is outside the valid original table profile. NRF-V05 (g)/(h) already give the zero-value return, while (l) rejects extending that bypass to a nonzero value with zero denominator |
| NRF-R09a fps 0 fault | No change: NRF-V05 (l) supplies the division-by-zero recipe and NRF-V14 supplies this helper's normal/recompute arithmetic. NRF-R09a states the fault; no retained-field or write-order assertion is inferred across it |
| NRF-R10b null path fatal | No change: NRF-V11 (c) supplies the fatal null-input recipe; NRF-V29 (g) demonstrates the independent getter result. The stated null-path assertion supplies no post-fault state |
| NRF-R11 wrapped speed ramp / step products | No change: NRF-V04 (k) distinguishes retaining the narrow product before division; NRF-V26 (f) retains low-32-bit shifted output; NRF-V27 (c)/(e) demonstrate wrapped increments/sums. NRF-V22 exercises this rule's ramp and signed product/shift recipes. No additional overflow policy is introduced |
| NRF-R12a FindStateList null unit, missing list/storage and multiple matches | No change: fatal null-input is NRF-V11 (c); null lookup is NRF-V33 (b); primary-before-secondary duplicate selection is NRF-V33 (a), with secondary fallback (g). The rule explicitly returns the first matching list and states its distinct guards; extra absent-storage/duplicate fixed identities add no new traversal/result recipe |
| NRF-R12a HasState null/type/storage/range guards | No change: NRF-V15 (b)/(i) demonstrate rejected types returning 0, (c)/(g)/(h) range rejection, (e)/(f) accepted types and bit masks. Other guard failures share that zero-result recipe without changing the explicit per-operation guards |
| NRF-R12a SetState valid null unit / type-independent routing | No change: NRF-V33 (c) demonstrates the bit request with no stored effect followed by update; (e)/(j)/(l) demonstrate set, clear and unchanged nonzero argument. A valid null unit retains the stated update fault/integration boundary, not a successful default. NRF-R12a explicitly has no type filter; repeating a type binding adds no new operation recipe |
| NRF-R14b missing base list | No change: NRF-V31 (b) already demonstrates early return with no visits/requests; the missing-base guard selects that same recipe |
| NRF-R14b retained flagged extended-list restart | No change: NRF-V31 (a) explicitly observes reloading the current head. A retained flagged extended list can make that loop nonterminating, as NRF-R14b and section 6 state. No extra fixture mutation or termination policy is invented |
| NRF-R17a unavailable MonStats record | No change: NRF-V17 (d)/(f) demonstrate the same pre-allocation null result, unchanged game seed and no seeder request |
| NRF-R18 prepared mode upper-bound fatal | No change: fatal assertion is NRF-V11 (c); valid prepared-object arithmetic is NRF-V32. NRF-R18 explicitly states the mode>=8 assertion; no invalid record indexing or preceding/post-fault write assertion is supplied |

HARDEST entries were NRF-R03c's FILETIME/counter wrapping, NRF-R05's threshold order
and NRF-R14b's primary-head restart. All received no change: NRF-V27, NRF-V04 and
NRF-V31 already state their inputs, arithmetic or observed visit order explicitly.
No rule, gate, evidence label or old case/result changed. The two local-order assertion
edits require a third fresh full reader pass. The case set remains 25 vector families
and 160 labeled cases. Inherited status-application cross-reference warnings remain
outside this numeric change's ownership and were addressed by that companion's
regeneration. This companion retains its public identifier-mapping destination; that
destination describes historical subjects and their current owners. Defined ids, all
31 leaf labels, all old case values and all observation definitions remain unchanged.
Exact old vector text is preserved apart from the one added fault case and the two
added local-order assertion/discrimination clauses.
After the linked destination was restored by its owner, the numeric spec check reported
0 errors and 0 warnings on 2026-10-15.

### Campaign full reader pass 3 and completion (2026-10-15)

A fresh smaller reader independently recomputed all 25 vector families and 160 labeled
cases from the main contract and dependencies. Counts: MISMATCH 0, GUESS 0, GAP 0,
DISCRIM 0, GATED 22, COVERAGE 11, HARDEST 3. The gated grouping is the same as
full pass 2: eleven questions NRF-Q02, NRF-Q03, NRF-Q04, NRF-Q09, NRF-Q10,
NRF-Q11, NRF-Q13, NRF-Q17, NRF-Q19, NRF-Q20 and NRF-Q21; unresolved NRF-R04c
and NRF-R09b; and the nine supplied-input stage vectors NRF-V16, NRF-V17, NRF-V20,
NRF-V24, NRF-V28, NRF-V30, NRF-V31, NRF-V32 and NRF-V33. These are the same
explicit assertion boundaries, not new blanket gates.

All advisory entries received no change, for the following reasons. No fixture,
expected output, behavioral assertion, ownership or gate changed after this full pass.

| Advisory entry and all reported branches | Disposition |
|---|---|
| NRF-R02 negative modulus / additional masks | No change: NRF-V02 (b) rejects stepping for a negative modulus and (d) checks the offset guard; NRF-V01 (b)/(e) and NRF-V02 (f)/(g) distinguish mask arithmetic, including zero/one and a non-power-of-two argument. Further masks reuse that same new-low-word mask recipe |
| NRF-R06a low32 product wrap | No change: NRF-V04 (k) explicitly distinguishes the narrow product from widened arithmetic before division; NRF-V13 (a)/(c) distinguish the signed truncating quotient operations. NRF-R06a composes those already explicit width/rounding operations |
| NRF-R07 indices 0/2/3, unflagged table and zero | No change: indices 0/2/3 have the same fixed flag/A values as index 1; NRF-V05 (a) to (e)/(i)/(j) cover that arithmetic, and (f) distinguishes the different A of index 4. All original table flags are 1, so an unflagged row is outside the valid original table profile. The zero-result bypass recipe is (g)/(h); (l) rejects bypassing a nonzero value merely because its denominator is zero |
| NRF-R09a fps 0 | No change: NRF-V05 (l) supplies the zero-divisor fault recipe; NRF-V14 exercises this helper's normal/recompute results. NRF-R09a states the fault without asserting retained fields or preceding writes |
| NRF-R10b null path | No change: NRF-V11 (c) supplies the fatal null-input recipe; NRF-V29 (g) directly checks this getter's extraction result. The null-path assertion has no post-fault state guarantee |
| NRF-R12a null unit, missing storage and negative SetState on null unit | No change: NRF-V15 (b)/(i) supply the zero state-test guard result, NRF-V33 (b) the null lookup result, and (c) the no-stored-bit-effect plus update request. NRF-V33 (k) directly checks the negative-state no-request guard; NRF-R12a applies it before any valid-state bit/update requests, including on a null unit. Valid null-unit update handling retains its explicit fault/integration boundary rather than a successful default |
| NRF-R12b positive count without matching identity | No change: NRF-V25 (f) directly supplies positive count 2 with entries (34,1)/(36,0), requests (34,0), and expects 0 with no fallback |
| NRF-R14a type 2/3/4 equal and changed modes | No change: NRF-V16 (b)/(c)/(g)/(i) exercise ordinary equality/change routing, unrestricted requested mode and the monster-only exception's limit; (d)/(h) isolate type 5 refusal. Types 2/3/4 use the ordinary path, with no additional type-specific local arithmetic or request recipe |
| NRF-R14b missing base / flagged negative state | No change: NRF-V31 (b) supplies early return without visits/requests; (c) rejects placing free under the strict lower state bound, and (d) rejects placing it under the upper bound. A negative state follows the same lower-bound failure and independent free recipe. The rule still explicitly requires a base list before traversal |
| NRF-R17a missing MonStats record | No change: NRF-V17 (d)/(f) exercise the common pre-allocation rejection result, unchanged game seed and absent seeder request. The unavailable-record predicate reaches that same refusal |
| NRF-R18 upper mode assertion / nonobject dispatch | No change: the fatal-assertion recipe is NRF-V11 (c); NRF-V32 exercises the valid prepared object branch. NRF-R18 explicitly limits the random branch to prepared type 2 and states its upper mode assertion. Other types' animation behavior and full preparation remain outside that supplied-object slice; no new nonobject result or invalid indexing is inferred |

HARDEST entries also received no change: NRF-V04 already states signed thresholds,
branch order and divide fault; NRF-V27 states wrapped time/counter inputs and all three
scramble results; NRF-V26 states the minimum-condition matrix, unshifted comparison and
low32 replacement. Their difficulty did not reveal a missing behavioral choice.

The campaign blind-implementer review is complete for this revision: three full passes
returned MISMATCH 0, every earlier required finding has its recorded resolution, and
the final applicable pass has no GUESS, GAP or DISCRIM. Advisory branches retain the
explicit outcomes above. The document remains draft because the recorded in-scope
gates and pending gameplay observations remain open. Review completion does not imply
user approval, source re-verification beyond the recorded slices or complete parity.

### Independent review spot check and corrections (2026-10-15)

An independent review read the following original 1.13c instructions before comparing them
with the rules: D2Game `0x1180`, `0x1450`, `0x1480` (NRF-R01/NRF-R02); `0x14D0` (NRF-R05);
`0xE0110` (NRF-R03a/NRF-R03c); `0xE1D90` entry through `0xE0EDC` (NRF-R17a); `0xE00C0` with
its D2Common imports #10679, #10414, #10193, #10411 and #10874 (NRF-R17b); `0x77240` slot
loop `0x77510..0x77630` (NRF-R04b); D2Common `0x36760`, `0x265D` (NRF-R03c); `0x323E0` and
the image table at `0x94608` (NRF-R07); `0x38B70`, `0x38A80`, `0x382B0` (NRF-R12b/NRF-R12c);
`0x33CD0`, `0x37DB0`, `0x34510`, `0x38730`, `0x38520`, `0x37F80` (NRF-R12a); `0x33920`
(NRF-R14a); `0x3AE60` (NRF-R14b); `0x335F0` object slice `0x3378D..0x3381E` and helper
`0x1100` (NRF-R18); `0xCEB0` (NRF-R11a); `0x4D910`, `0x4D950` (NRF-R10a); `0x34D40`
(NRF-R08b); D2Game `0x29800` (NRF-R09a) and D2Common `0x34B80` (NRF-R10b). Undecoded byte
runs at `0x38A80`, `0x3AE60`, `0x38730` and `0x3379C` were decoded by hand from their opcode
bytes. Every arithmetic, guard, order and seed-owner statement of those rules was confirmed,
including the direction-table entries 0 `(0,4096,0)` and 127 `(2896,2896,7)` at D2Common
`0x8D470`. Recomputed vectors NRF-V01, NRF-V20, NRF-V22 (c), NRF-V27, NRF-V30 (a) and
NRF-V32 (e) matched.

Corrections, each preserving behavior, arithmetic, owners and vector values:

- **NRF-R17a monster test.** The earlier record (`0xE1DA0..0xE1DDC`) described the tested
  byte as intersecting "an engine-held mask byte" and named the result neutrally. New
  evidence: the mask pointer at D2Game IAT `0xF8BE0` resolves to Fog #10259 / RVA `0x26798`,
  the constant bit table `1,2,4,8,...`; the test reads mask[1] = 2 against MonStats record
  byte `+0x0F`, which is bit 25 of the flag dword `+0x0C`. The MonStats descriptor array in
  D2Common `0x5AA40` (descriptor stored at stack `+0x240..+0x250`) names that field
  `enabled` (string at D2Common RVA `0x8E9B0`), conversion kind 0x1A, bit 25, offset `0x0C`,
  matching GDF-R05a's `enabled` flag. The section 3 input, NRF-R17a and NRF-V17 now name the
  MonStats `enabled` flag instead of the neutral `creationCondition`, which also removed a
  name collision with the different post-creation `creationCondition` of MSD-R19. NRF-Q09
  no longer lists this condition. Vector inputs and results are unchanged.
- **NRF-R04a rows.** The knockback row read "Skill/hit knockback ... knockback roll"; its
  consumer KNB-R01 is the item knockback roll, one attacker step with new low word `& 127`.
  The leech row cited DPO-R15, whose contract delegates the overlay Rand(2) to LSE-R13; the
  row now cites LSE-R13 reached through DPO-R15, and Depended on by lists LSE. Seed owners
  are unchanged.
- **Call-site anchors added.** D2Game `0xCACC0` (NRF-R03d) is called from `0x2C2D0` and
  `0x2C570`; the generic creator `0xE1D90` (NRF-R17a) has 22 recorded direct callers,
  including item creation `0x11490`; the mode setter `0x33920` (NRF-R14a) tail-requests
  D2Common `0x335F0`, whose prepared type-2 slice is NRF-R18. These are call-site anchors, not
  complete frame-loop chains.

A fresh delta reader check of the renamed NRF-R17a input is required under checklist
section 2 before the reader review again applies to this revision.

## Original function references

### Random primitives, seeding and arithmetic

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0x1160` | — | Two-word random step | NRF-R01 | Reconstruction checked against bytes; historical local step |
| LOD113C-D2GAME | `0x1180`, `0x11D0` | — | Positive-modulus draws | NRF-R01, NRF-R02 | Disassembly; signed guard, step, mask/modulo; `0x1180` listing checked again 2026-10-14 |
| LOD113C-D2GAME | `0x1450` | — | Always-step masked draw | NRF-R02 | Reconstruction checked against bytes; no modulus guard |
| LOD113C-D2GAME | `0x1480` | — | Offset random range | NRF-R02 | Disassembly; signed guard and offset addition |
| LOD113C-D2COMMON | `0x1100` | — | Object-reset draw helper | NRF-R02, NRF-R18 | Disassembly; positive-modulus helper |
| LOD113C-D2COMMON | `0x36740` | 10614 | Initialize seed from supplied value | NRF-R03a, NRF-R03b, NRF-R17b | Reconstruction checked against bytes; low/high writes |
| LOD113C-D2COMMON | `0x36750` | 10936 | Initial seed/timer values | NRF-R03a, NRF-R03b | Reconstruction checked against bytes; initial writes |
| LOD113C-D2GAME | `0xE0110` | — | Unit seeder | NRF-R03a, NRF-R03b, NRF-R03c, NRF-R17a | Disassembly; initial write, parent selection, counter and final seed/copy; order reread 2026-10-14 |
| LOD113C-D2COMMON | `0x36760` | 10995 | Parentless sampled-time scrambling | NRF-R03c | Full disassembly reading 2026-10-14; three rounds and result mask |
| LOD113C-D2COMMON | `0x265D`, `0x4DD0` | — | File-time conversion and unsigned division | NRF-R03c | Full local disassembly 2026-10-14; unsigned wrap, both divisor paths and low-word return |
| LOD113C-D2GAME | `0xCACC0` | — | Game seed stage | NRF-R03d | Independent listing 2026-10-14; step, copy/local initialization and resource request |
| LOD113C-D2GAME | `0xCAB30` | — | Resource initialization boundary | NRF-R03d, NRF-Q13 | Entry/calls only; not complete resource draw verification |
| LOD113C-D2GAME | `0x14D0` | — | Signed multiply/divide primitive | NRF-R05 | Disassembly; thresholds, precision branches and narrow IDIV fault; reread 2026-10-14 |
| LOD113C-D2GAME | `0xF44B0` | — | Wide signed-division runtime support | NRF-R05 | Disassembly reading recorded 2026-10-13 in the [pipeline arithmetic audit](combat-damage-pipeline-order.md#original-function-references); reused here, not independently reread in the 2026-10-14 numeric pass |
| LOD113C-D2GAME | `0xDD450` | — | Signed percent consumers | NRF-R06a | Disassembly; signed /100 and /128 idioms; full damage build owned elsewhere |
| LOD113C-D2COMMON | `0x33110` | 10819 | Unsigned /100 animation consumer | NRF-R06b | Independently decoded quotient slice 2026-10-14; input production/update remains NRF-Q17 |
| LOD113C-D2COMMON | `0x323E0` | — | Diminishing-return getter | NRF-R07 | Disassembly; valid-input stat selection and signed formula |

Relevant branches and writes:

- `D2Game.dll+0x1160` computes unsigned `lo*0x6AC690C5+hi` widened to 64 bits and
  stores both halves. `0x1180` tests signed `n<=0` before the step, then selects mask
  or remainder using `n & (n-1)`. `0x1450` always steps and masks with `m-1`.
  Modulo 100 and masking with 99 differ, as NRF-V01 now explicitly distinguishes.
- `D2Game.dll+0xE0110` calls D2Common #10936 before parent selection, advances a
  supplied parent once, then initializes the unit from its new low word through #10614.
  Without a parent it takes and increments counter RVA `0x111C50`, supplies twice the
  old counter modulo 32 bits to #10995, then writes unit seed and initial copy. The
  original counter starts zero. This does not locate the player-load seed producer.
- `D2Common.dll+0x265D` adds `0xFE624E212AC18000` modulo 64 bits and divides unsigned
  by 10000000 through `0x4DD0`; its return retains the quotient's low word. `0x36760`
  adds converted seconds, GetTickCount and the supplied argument modulo 32 bits, applies
  three rounds `x*0x19660D+0x3C6EF35F`, then clears bit 31. The original named OS
  imports are GetSystemTimeAsFileTime and GetTickCount. Sampled values are vector inputs.
- `D2Game.dll+0xCACC0` steps Game `+0xD0/+0xD4`, copies new low to `+0xEC`, initializes
  a local seed through #10614 and requests `0xCAB30`. Function-internal locations
  `0xCACCD..0xCACFF`, `0xCAD12` and `0xCAD21` show those writes/call. The incoming
  producer and resource initializer remain NRF-Q13; one direct step is not whole-game creation.
- `D2Game.dll+0x14D0` uses signed `0x10000` threshold/shift comparisons. Divide-first
  branches truncate before multiplication; the narrow product branch uses IDIV and
  faults for INT_MIN/-1. Operand order is asymmetric. No generic caller percent order
  follows. The [pipeline arithmetic audit](combat-damage-pipeline-order.md#original-function-references)
  records a 2026-10-13 disassembly reading of both `0x14D0` and its wide signed-division
  support at `0xF44B0`, explicitly mapped to NRF-R05. That earlier reading supplies the
  helper coverage; the 2026-10-14 numeric pass did not independently reread its internals.
- At D2Common RVA `0x331B7`, unsigned MUL by `0x51EB851F`, then SHR of high product
  by 5, gives unsigned /100. The contrasting signed sequence at `0x31F90` has arithmetic
  shifting and sign correction. The multiplier alone does not prove signedness.
- D2Common RVA `0x94608` holds five triplets `(flag,A,stat)`:
  `(1,120,93)`, `(1,120,99)`, `(1,120,105)`, `(1,120,102)`, `(1,150,96)`.
  `0x323E0` reads that table and the existing-array getter, then uses wrapped multiplication
  and signed division. No null-unit/index guard establishes portable malformed-input behavior.
  These constants come from the original image, not MPQ data.

### Reached draw-site ownership

These references establish local seed ownership, not every upstream test, callee draw or
complete gameplay total. All were disassembly readings; the stated independent listings
were added on 2026-10-14. Full caller behaviors belong to the spec's named gameplay owners.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0xDDE90` | — | Hit roll | NRF-R04a | Attacker seed at reached hit draw |
| LOD113C-D2GAME | `0xDD450` | — | Damage build | NRF-R04a | Attacker direct build draws |
| LOD113C-D2GAME | `0x770E0` | — | Monster damage doubling | NRF-R04a | Attacking-monster seed and nonzero Crit gate |
| LOD113C-D2GAME | `0xDBA40` | — | Leech overlay | NRF-R04a | Attacker Rand(2) |
| LOD113C-D2GAME | `0xDB910` | — | Hit-recovery veto | NRF-R04a | Defender mask draws |
| LOD113C-D2GAME | `0xDB790`, `0xDB030` | — | Block and avoidance | NRF-R04a | Defender direct sites |
| LOD113C-D2GAME | `0x77FA0` | — | Concentration gate | NRF-R04a | Defender seed and concentration-stat comparison |
| LOD113C-D2GAME | `0x3AE10` | — | Missile contact | NRF-R04a | Missile KnockBack gate before block |
| LOD113C-D2GAME | `0x39B40`, `0x3A4E0` | — | Missile damage and caller | NRF-R04a | Missile-owned ranges |
| LOD113C-D2GAME | `0x4F4D0` | — | Corpse-explosion damage | NRF-R04a | Corpse seed after corpse-state request |
| LOD113C-D2GAME | `0xAFDB0` | — | Rolled monster life | NRF-R04a (seed owner); MSD-R11, MSD-R12 | New-monster seed; roll, bonus and life requests are owned by MSD-R11/MSD-R12 and reuse this record |
| LOD113C-D2GAME | `0xB1400` | — | Invisible-candidate filter | NRF-R04a | Candidate seed; roll below 80 reaches hostility test |
| LOD113C-D2GAME | `0xDEBA0`, `0x12EE0` | — | Equipped/weapon durability | NRF-R04a | Defender equipped-item / attacker weapon sites respectively |
| LOD113C-D2GAME | `0x77240`, `0x75E10` | — | Monster-mode element slots and route | NRF-R04b | Independent reached-site listing; upstream/callee totals NRF-Q11 |
| LOD113C-D2GAME | `0xB1820` | — | Stored-target mask draw | NRF-R04b | Independent listing; searcher-owned reached type-3 path |
| LOD113C-D2GAME | `0x37030`, `0x9E290` | — | Item-event dispatch/slot lookup | NRF-R04c | 50-slot lookup and shared dispatch inspected; handler totals unverified |

Stun-cancel, cold-shatter and freeze-fallback site records (NRF-R04a rows naming CSA-R03,
CSA-R06 and CSA-R07) are in the [status application companion](combat-status-application.md#function-references).
Monster element draw
sites at D2Game RVAs `0x77545/0x77612` take the supplied monster seed; chance below
100 and selected type 10 govern the two local draws. Stored-target sites `0xB1968/0xB1971`
take searcher Unit `+0x20`. These addresses are instruction sites inside listed functions.

Correction dated 2026-10-13: `0x12EE0` performs one inline step, modulo 100 compared
with chance 4 or 10. Ten is a percent chance, not ten steps; there is no return-to-step
loop. The correction does not invent aggregate durability totals.

### Fixed-point, time, movement and object animation

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2COMMON | `0x34D40` | — | Path-speed setter | NRF-R08b | Reconstruction checked against bytes; speed/target/marker writes |
| LOD113C-D2GAME | `0x29800` | 10026 | Frame-rate setter | NRF-R09a | Reconstruction checked against bytes; quotient and two writes; reread 2026-10-14 |
| LOD113C-D2GAME | `0xF7970` | — | Frame-time recompute | NRF-R09a | Reconstruction checked against bytes; existing fps and quotient write; reread 2026-10-14 |
| LOD113C-D2GAME | `0x2E1E0` | — | Milliseconds/frame reader | NRF-R09b | Disassembly search context, not startup policy |
| LOD113C-D2COMMON | `0x4D910`, `0x4D950` | — | Coordinate divide and inverse | NRF-R10a | Disassembly / checked reconstruction respectively; signed correction and low-word inverse |
| LOD113C-D2COMMON | `0x34B80`, `0x34BB0` | — | Path integer X/Y getters | NRF-R10b | Reconstructions checked against bytes; high-word zero extension |
| LOD113C-D2COMMON | `0xCEB0` | — | Velocity ramp and step writer | NRF-R11a | Disassembly; ordered clamps and fixed-point products |
| LOD113C-D2COMMON | `0xD240`, `0xCF40` | — | Direction writer/helper | NRF-R11b | Disassembly and full constant-table read 2026-10-14; scale/two entries only |
| LOD113C-D2COMMON | `0x335F0` | — | Animation reset/object rate | NRF-R18 | Independent listing/dispatch bytes 2026-10-14; prepared object branch |
| LOD113C-D2COMMON | `0x3E980` | — | Object-record retrieval | NRF-R18 | Reconstruction checked against bytes; bounds/stride, no inferred column names |
| LOD113C-D2COMMON | `0x32220`, `0x501E0` | — | Animation preparation/type routing | NRF-Q17 | Disassembly/listing inspection; ordinary objects stay objects, complete preparation gated |

- Monster life setup (record reused by MSD-R12, which owns the rule) requests maximum stat 7
  then current stat 6 through D2Common #10887 (`0x3A740`), after signed upper clamp and `hp<<8`.
  These raw 8.8 write requests do not verify actual replacing-write propagation (NRF-Q20).
  NRF-R08a keeps only the 8.8 format this record illustrates.
- Path speed/target are `+0x7C/+0x84`, reset marker `+0x38=15` only for changed speed.
  Position getters MOVZX Path high words `+0x02/+0x06`, not sign extension.
  Velocity writer `0xCEB0` uses direction `+0x6A/+0x6E`, step `+0x72/+0x76`, speed
  `+0x7C`, target `+0x84`, acceleration `+0x88`, counter `+0x8C`. Target clamp precedes
  negative clamp; product low words/arithmetic shifts establish NRF-R11a rounding.
- `0xD240` stores `0xCF40` results. The helper quantizes slope into three-dword entries
  at D2Common RVA `0x8D470`, with swaps/signs. A 1536-byte read covered 128 entries;
  retained facts are entry 0 `(0,4096,0)` and entry 127 `(2896,2896,7)`. Full producer/
  table reproduction remains NRF-Q04. These are image constants, not archive cells.
- D2Game globals are RVAs `0x10E3CC` (fps), `0x111F70` (milliseconds/frame), initialized
  25/40; recompute pointer occurs at RVA `0xFF018`. Caller searches did not establish
  indirect startup policy. Helper inputs 24/30 do not imply a changed live-game rate.
- D2Common dispatch table RVA `0x33900` selects prepared type-2 object's slice
  `0x3378D..0x3381E`. It reads record total `+0xD8+4*mode`, start byte `+0x129+mode`,
  signed base rate `+0xF8+2*mode` and fixed-rate condition `+0x175`; writes frame before
  the draw and rate at Unit `+0x4C`. Correction dated 2026-10-14: randomness changes
  rate, not starting frame. Complete reset preparation remains NRF-Q17.

<a id="stat-reader-verification-2026-10-14"></a>
## Stat-reader verification (2026-10-14)

This targeted reading owns NRF-R12b/R12c existing-array retrieval, not array construction,
attached-list aggregation, replacing writes or gameplay. The recorded D2Common analysis
hash matches the manifest's `LOD113C-D2COMMON` fingerprint.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2COMMON | `0x38B70` | 10973 | Unit stat getter | NRF-R12b | Entry-to-return disassembly; unit/list/id guards |
| LOD113C-D2COMMON | `0x38A80` | — | Selected-array getter/minimum | NRF-R12b, NRF-R12c | Independent entry-to-return instruction listing |
| LOD113C-D2COMMON | `0x382B0` | — | Exact signed-key search | NRF-R12b | Checked reconstruction; complete 53-byte helper |
| LOD113C-D2COMMON | `0x39140` | 10680 | List stat getter | NRF-R12b | Entry-to-return disassembly; own-array route |
| LOD113C-D2COMMON | `0x38B00` | — | Own-array getter/minimum | NRF-R12b, NRF-R12c | Entry-to-return disassembly |
| LOD113C-D2COMMON | `0x73330` | — | Stat-definition descriptors | NRF-R12c, NRF-R16a | Selected original instructions/strings checked independently |
| LOD113C-FOG | `0x26798` | 10259 | Bit-mask data export | NRF-R12a, NRF-R12c | Import/export resolution and narrow image read; data address, not function |

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

### States, mode and creation

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2COMMON | `0x38520`, `0x37F80` | 10871 for `0x38520` | State-list wrapper/finder | NRF-R12a | Disassembly; primary then secondary chains |
| LOD113C-D2COMMON | `0x33CD0`, `0x37DB0` | 10494 for `0x33CD0` | State-test wrapper/bit read | NRF-R12a | Disassembly; guards and bit result |
| LOD113C-D2COMMON | `0x34510`, `0x38730` | 10945 for `0x34510` | State-set routing/bit writes | NRF-R12a | Disassembly/independent listing 2026-10-14; disguise/update integration gated |
| LOD113C-D2COMMON | `0x1770`, `0x33C20` | — | State-record/disguise intersection | NRF-Q17 | Checked reconstruction / disassembly; dependency inspection only |
| LOD113C-D2COMMON | `0x33920` | 11090 | Mode setter | NRF-R14a | Disassembly; guard, equal exception, ordered writes/requests |
| LOD113C-D2COMMON | `0x3AE60`, `0x3A9F0` | 10196 for `0x3AE60` | Mode-bound list traversal/free | NRF-R14b | Independent traversal listing / checked free reconstruction; free effects dependency-owned |
| LOD113C-D2COMMON | `0x665C0`, `0x2FE10` | 10158 for `0x665C0` | Update/region queue context | NRF-Q17 | Selected disassembly/listing inspection, not full network contract |
| LOD113C-D2GAME | `0xE1D90`, `0x6F930` | — | Generic/missile seeding stages | NRF-R17a | Disassembly; class refusals and successful local seeding |
| LOD113C-D2GAME | `0xE00C0` | — | Secondary item seed | NRF-R17b | Disassembly; initial seed, game draw, copy and final writes |
| LOD113C-D2COMMON | `0x23BE0`, `0x23B70`, `0x23BC0`, `0x23C10`, `0x23C70` | 10679 / 10414 / 10193 / 10411 / 10874 respectively | ItemData seed/copy access and allocator | NRF-R17a, NRF-R17b | Reconstructions checked against bytes; zero allocation/sentinel and separate seed |

Status application records (cold, freeze, poison, burn and stun appliers, gate helpers, cold
expiry callback, status hand-off, expiry setters, difficulty getter, stay-death and timer
requests) are in the [status application companion](combat-status-application.md#function-references),
which owns CSA-R01 to CSA-R13.

Relevant branch/output rationale:

- State chains start at Unit `+0x5C`; primary/secondary heads are StatList `+0x3C/+0x40`,
  state `+0x14`, expiry `+0x18`, bit buffer `+0x58`. Same bit-31 condition enables
  extended-list state storage and resolvedStats. Valid SetState routes its bit request
  and unconditional update request; full disguise/update effects stay NRF-Q17.
- Equal mode returns 0; only equal monster mode 1 skips update/modeChanged. No equal
  case processes lists/resets. Changed mode requests update, marks changed, stores mode,
  processes lists, then resets. At D2Common `0x3AEBB`, traversal reloads primary head
  after each processed flagged list; unflagged nodes follow saved successor. A retained
  flagged extended list may repeat indefinitely; V31 supplies finite removal.
- Creation rejects unsigned player class>=7 and negative/out-of-range monster class
  before successful seeding. Player skips seeder, other successful types draw once;
  item allocates zeroed 0x74-byte ItemData/sentinel -1 and draws again for its distinct
  secondary seed. Unit seed `+0x20/+0x24`, initial copy `+0x28`; ItemData seed `+0x04/+0x08`,
  copy `+0x10`. Separate missile seeding draws once; later initialization totals are excluded.

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
| maxhp / 7 | fMin; MinAccr; ValShift | `1`; `1`; `8` | NRF-R12c/V26 metadata; owner/value setup is supplied fixture |
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

These divisor cells support the status appliers CSA-R05 and CSA-R08; they remain here because
the status application companion links this record.

States text coverage was 20920/20920 bytes, 185 data rows, with zero-based identity
correspondence checked excluding header. ItemStatCost text above contained 359 rows.
NRF-R16b consumes specific stat/state/mode/ItemTypes/monster identity records in the
[game-data evidence companion](core-game-data-fields.md), and NRF-R16a consumes its
per-table descriptor/getter records. This named dependency does not assert that the
numeric pass reread every descriptor. MonStats cached/text tail discrepancies remain
recorded there; no universal text-to-compiled class correspondence or custom remapping
capability follows. These are data-only records, not gameplay confirmation.

## Historical read-only observations

The following layout/constant reads are not reproduced gameplay vectors. Server state
was read without driving an attack, installing a new trace, writing memory or editing a
save. Historical live DLL hashes were not captured except where the source identity
record explicitly establishes them; newly fingerprinted originals don't automatically
prove all historical live-module bytes matched. Source profile is vanilla LoD 1.13c.

| Assertion | Date | Module + RVA / original field | Records | Result and limit |
|---|---|---|---|---|
| NRF-O02 globals / NRF-R09a | 2026-10-13 | D2Game `0x10E3CC/0x111F70` | 2 reads | 25/40 matched; no 60-second frame count |
| NRF-O01 layout / NRF-R01,R03a | 2026-10-12 | Game `+0xD0/+0xD4`; server Unit `+0x20/+0x24/+0x28` | 3 game-seed / 2 player reads | Distinct low/high, high=666, copy=low; idle player unchanged, arithmetic not exercised |
| NRF-V05 constants / NRF-R07 | 2026-10-12 | D2Common RVA `0x94608` | 1 read | Five triplets above matched, formula not executed |

Game reads at frames 3714, approximately 4922 and 6600; player reads approximately
3000 frames apart. Client player seed did step while idle, outside these server rules.
No attributed gameplay mismatch was recorded. None of these readings runtime-confirms a rule.

<a id="observation-procedures-retained"></a>
## Tool-neutral observation procedures

All three gameplay procedures remain pending. Use an owned matching original installation
and a read-capable debugger/instrumentation system. Record module fingerprints, effective
data profile, date, inputs, sample times/frames, seed pairs and outputs. Add RVA to loaded
module base to account for relocation. Original offsets below locate evidence; they are
not required implementation layouts. Keep captures private and publish narrow measurements.
No attack, movement or status execution is claimed by the following prospective procedures.

### NRF-O01: one step per reached draw

On a level-1 single-player Normal character, identify server type-0 Unit with Game pointer
`+0x80`. Read seed `+0x20/+0x24`; a roughly 3000-frame idle control should leave it
unchanged. Outside town capture one bare-handed attack's initial/final seed and reached
attacker sites (`D2Game.dll+0xDDE90/+0xDD450`). Advance NRF-R01 from the initial pair
until reaching final; compare draw count/order with hit-roll and damage-build contracts'
actual reached branches. A miss ends after the hit roll. Exclude client-copy/concurrent
draws. Historical idle reads did not perform this one-swing arithmetic test.

### NRF-O02: frame-rate measurement

In Normal town sample Game `+0xA8`, external wall time and D2Game RVAs
`0x10E3CC/0x111F70`. Sample frame/wall time again after 60 seconds. Original 25-fps
profile expects 1500 frames within the two sampling-frame uncertainties, globals 25/40.
Record pauses/blocking that invalidate the interval. Only globals were historically read;
frame counting stays pending, and startup/later-rate policy remains NRF-Q03.

### NRF-O03: coordinate unit

In Act I town locate player Path through Unit `+0x2C`, read integer words `+0x02/+0x06`.
Choose an independently known n-tile automap displacement along one axis, walk until
stopped, reread and expect 5n path units along that axis. Record tile count/positions and
obstructions. Getter functions are D2Common `0x34B80/0x34BB0`. This does not establish
every static-position adapter or direction producer (NRF-Q04).

### Cold extension procedure

The cold-extension procedure belongs to the status application owner as CSA-O01; see the
[status application companion](combat-status-application.md#csa-o01-cold-extends-only).

<a id="foundation-migration-source-record-2026-10-14"></a>
## Foundation migration source record (2026-10-14)

The local readings above established supplied-clock arithmetic, game-seed stage, reached
monster/searcher owners, ColdEffect defaults, high-word getters, speed setter, direction
scale/two entries, unsigned quotient, state-bit routing, mode head restart, prepared
object-rate branch and status gate helpers. Ambiguous instruction regions were independently
checked before bounded labels were assigned. Full gameplay paths/dependencies were not promoted.

| Question | Original-artifact attempt | Remaining boundary |
|---|---|---|
| NRF-Q02 | Unit seeder and sampled-time chain read fully | No-parent gameplay creation not observed |
| NRF-Q03 | Relative-call scans to D2Game `0x29800/0xF7970`, pointer/global search, imports in D2Game,D2Client,Game,Fog,D2Launch,D2Multi | Zero direct E8 calls and zero imports of D2Game #10026 in those six; indirect startup caller unknown |
| NRF-Q04 | Getters, direction writer/helper and 128-entry image read | Other/static-position producers and full direction normalization unknown |
| NRF-Q09 | Tested predicates/masks retained with neutral names | Wider gameplay meanings not invented |
| NRF-Q10 | Shared event dispatcher and 50-slot lookup | Per-handler totals owned elsewhere |
| NRF-Q11 | Independent `0x77240/0xB1820` reached-site listings | Upstream eligibility/additional callee totals |
| NRF-Q13 | `0xCACC0` stage, `0xCAB30` entry/calls, seed candidates | Incoming producer/resource initialization |
| NRF-Q17 | Update/region, disguise, animation and timer slices inspected | Complete dependency effects/handler execution |
| NRF-Q19 | D2Game candidates `0x11490/0x101E0/0x104F0/0xB9A0` inspected | `0x11490` is item creation, no complete player-load seed anchor; search not exhausted |
| NRF-Q20 | Attach/write entries `0x3A7D0/0x3A280/0x3A740` and callees inspected | Actual lifecycle-owned propagation remains unverified |
| NRF-Q21 | Historical readings and O01..O04 reviewed | No gameplay campaign performed |

The historical rate-caller search also inspected D2Client.dll, D2Launch.dll and
D2Multi.dll. Their fingerprints were not captured in the public manifest; that search's
negative results are limited to the historical reported profile, not a reproducible
claim about every file bearing those version labels.

These bounded searches are not absence proofs. Supplied-input arithmetic/request vectors
do not close integrations; exact blocked assertions remain in the gap file.

## Additional public correction rationale

- Cold expiry callback is D2Game `0xDB4C0`, ColdEffect helper `0xDAE20`; mistaken earlier
  locations/labels were replaced by caller/output evidence.
- Stun duration 300 caps at 250. V21's positive cold contrasts (now CSA-V07 and CSA-V08) each use a defender draw;
  adjacent stun cases' no-draw prose does not apply to those contrasts.
- Equal-mode exception and head-restart traversal were clarified from their original
  branches, preserving existing results. Named freeze roles remove native ABI ambiguity.
- Caller-specific mulDiv ordering remains caller-owned. Modulus presentation errors were
  corrected to `n & (n-1)` from the instruction sequence, not a changed engine formula.
- Proposed V32 modulus correction was rejected: `0xD58D2424 % 4095=3531`, raw 34251
  clamps to 32767. Independent arithmetic agreed; the correct expectation stayed unchanged.

<a id="reader-and-preservation-checks-2026-10-14"></a>
## Historical reader and preservation checks (2026-10-14)

Readers used specs/dependencies only, not source/game/code/web. Targeted getter review's
first pass returned MISMATCH0/GUESS0/GAP0, GATED3, COVERAGE0 required/1 advisory,
HARDEST3. A high-ID guard control was added, then a fresh final pass returned zero
mismatches/guesses/gaps and no required finding. It covered getter/mitigation integration,
not the later whole-foundation migration.

Migration counts below apply only to the historical reviewed revisions; earlier runs do
not approve later edits. GATED counts vary with grouping of dependencies/stage fixtures.

| Numeric pass | MISMATCH | GUESS | GAP | GATED | COVERAGE required / advisory | HARDEST | Disposition |
|---|---|---|---|---|---|---|---|
| 1 | 1 | 1 | 0 | 11 | 1 / 3 | 3 | Named freeze roles, explicit equal-mode, mask/modulo and edge controls; malformed diminish outside preconditions |
| 2 | 0 | 1 | 0 | 17 | 5 reported / 0 reported | 3 | Modulus presentation fixed, initial seeding order/fixed-rate writes asserted, burn/freeze/class controls added |
| 3 | 0 | 0 | 0 | 11 | 1 / 5 | 3 | Cold equality discriminator, clock/counter wrap, early gate/setter controls |
| 4 | 0 | 0 | 0 | 30 | 2 / 1 | 3 | fps 24 truncation/recompute and changed-mode order assertions |
| 5 | 0 | 0 | 0 | 13 | 1 reported / 4 | 3 | Required item belonged to GDF decimal overflow; numeric fault/poison controls and source-backed stun correction |
| 6, final applicable | 0 | 0 | 0 | 28 | 0 / 6 reported | 3 | No required findings; integration/observation gates remain |

Final advisory outcomes: below-13 hireling duration was covered by the then-current status fixtures; other
nonzero freeze helpers reach V30's same cold fallback (now CSA-V10), V35 covers helper results (now CSA-V04);
negative object-rate clamp shares zero branch; invalid prepared mode is outside valid
0..7; denominator-zero, fps-zero and unavailable monster Velocity are stated exceptional
paths, not promised successful fixtures. No defaults or runtime promotion were added.
Hardest final items were mulDiv faults/branches, stat selection/minimum and status request/
timer/draw flow; explicit inputs/boundaries already addressed their complexity.

Historical combined audits preserved 26 numeric plus 18 data fixtures verbatim in the
then-existing archives, resolved 120 foundation/evidence/gap/index links and checked
semantic-location/fence/whitespace hygiene. Consumer audit preserved 129 fixtures apart
from equivalent named-role Freeze notation, resolving 282 links. These totals describe
the historical revision, not this provenance rewrite. Complete historical archives now
remain private; public references, corrections and observation scopes remain above.

Final applicable behavioral review found no required fix; that is neither user approval
nor gameplay confirmation. Status remains draft. No implementation plan/code/test or
gameplay campaign resulted from these documentation readings. Current provenance-only
preservation/link checks are reported after execution, not asserted by this historical record.

## Public provenance checks (2026-10-14)

This focused update applied the public-provenance procedure without changing the behavioral
contract. Exact text comparison preserved everything before Appendix A, including all 35
executable vector definitions, rule labels and gates. The complete prior companion was
preserved outside the public repository and checked for exact equality with the pre-edit
snapshot; archive location and SHA-256 are recorded privately.

Original-file hashes and sizes matched both stored source bundles and the manifest. The four
DLL hashes matched the recorded analysis-image hashes. A read-only PE check mapped all 104
unique table references to original image sections (executable for function entries), matched
all 26 listed export targets, and matched seven additional prose export mappings. These are
identity/navigation checks, not new entry-to-output behavioral verification.

All 48 checked local Markdown destinations/anchors resolved, including 13 inbound companion
links; all five compatibility anchors remain. Public/private citation scans found no private
paths, reconstruction filenames, commits, member identifiers or tool calls in the public
manifest/companion. An independent focused reader found no remaining provenance issue after
the template, source-identity and link corrections. No implementation tests, gameplay actions,
new table reads or evidence-label promotions were performed. Pending observations and existing
gates remain unchanged; this record does not approve the spec or establish runtime parity.

## Focused evidence reconciliation (2026-10-14)

- **NRF-R05 / NRF-V04:** the earlier numeric record described what its own pass had
  not reread, while the pipeline arithmetic audit recorded reading D2Game `0xF44B0`
  on 2026-10-13. These are different reading scopes, not conflicting quotient results.
  The pipeline table explicitly links that helper and `0x14D0` to NRF-R05; its recorded
  disassembly coverage is reused for the original 1.13c image identified above. The
  numeric table and rationale now cite that coverage without claiming a fresh helper
  reading. Arithmetic, operand order, fault behavior and NRF-V04's widened case are
  unchanged. Reuse does not establish any new caller integration or gameplay result.
- **Timed-status summary (now the CSA section 4 preamble):** the previous timer summary omitted STL-R01's special requested frame
  `-1` immediate path, although both the consumed lifecycle contract and this companion's
  scheduler rationale recorded it. The summary now states that exception before the
  ordinary next-frame adjustment. No timer implementation is duplicated here; handler
  execution and actual queue integration retain NRF-Q17. Existing vector inputs and
  results are unchanged.
- **NRF-R16b / named game-data records:** profile wording now distinguishes the reported
  edition/build and profile family from the session executable hash, following the
  manifest. A complete override inventory was not captured. This clarifies the evidence
  limits without changing any retained named cell, interpretation or data label.

Exact comparison against the pre-edit spec found only that timed-status summary change;
all 35 executable vector definitions, labels, gates and other spec text were preserved.
All 45 outgoing local Markdown destinations/anchors from the spec and companion resolved.
These are document checks, not new source, table or gameplay verification. Status remains
draft, and the existing evidence gaps remain open.

The fresh blind-implementer pass used the resulting spec and cited dependencies only,
hand-recomputing all 35 vectors of that revision. It returned MISMATCH 0, GUESS 0, GAP 0,
GATED 13 existing limits, COVERAGE 0 required / 1 advisory, and HARDEST 3. The advisory
item was the absence of a dedicated requested-frame `-1` example: no vector was added,
because this summary consumes STL-R01 and the numeric vectors claim ordinary scheduling,
not sentinel-path coverage. That branch remains advisory under the lifecycle owner.
The hardest items were NRF-V27's wrapped time arithmetic, NRF-V04's signed division
branches/fault and the then-current freeze timer vector's ordered requests (now CSA-V11). No change was needed: the existing
inputs, arithmetic, exceptional behavior and requested/queued-frame distinction specify
those decisions. Existing gates remain explicit; this reader result is not user approval,
fresh source verification or a gameplay observation.

<a id="status-and-life-split-2026-10-14"></a>
## Status and monster-life split record (2026-10-14)

This record covers phase 2 of the user-approved split of this spec. No source, table or gameplay
reading was performed; the moved rules keep their original evidence, now cited by their owners.

- **Moved to status application.** Timed stun, cold, freeze, poison and burn application, the
  status gate helpers, the cold expiry callback, the poison/burn pre-step, the status hand-off and
  final lethal decision, the freeze-to-cold routing, their vectors, the cold-extension observation
  and the status parts of NRF-Q09, NRF-Q17, NRF-Q20 and NRF-Q21 now belong to
  [combat-status-application](../../specs/combat-status-application.md). The
  [status application subject and current-owner mapping](combat-status-application.md#identifier-mapping)
  records their current owners.
- **Moved to monster scaling.** The monster life roll and bonus are MSD-R11; the life clamp and
  the ordered stat 7/stat 6 write requests are MSD-R12, with MSD-V11 and MSD-V12. NRF-R04a keeps the
  seed-owner row (the new monster's seed) and cites MSD-R11. NRF-R08a keeps the 8.8 and 16.16
  formats only. The sixth vector's life-roll and life-write cases are covered by MSD-V11 (b) and
  MSD-V12 (a), (c), whose outputs match the values it carried. After reader pass 2 that vector id is
  unused: the format rule defines no operation of its own and is exercised by NRF-V22, NRF-V29 (g)
  and MSD-V12.
- **Ids.** The spec is unconsumed, but remaining ids keep their numbers at the user's request;
  the removed numbers are left unused and no alias column is kept.
- **Section 3.** Rows used only by moved rules were dropped (status gates, status dependency
  operations, cold/freeze divisors, life-roll inputs, list expiry/callback identities, status
  condition names, ColdEffect, Velocity, MinHP/MaxHP, deadCol and the status state/stat/timer
  mapping). The damage record of section 3.2 stays as the shared semantic record consumed by the
  pipeline and status application. NRF-R04a gained a freeze-to-cold fallback row naming the defender
  seed owner and CSA-R07, which CSA's scope sentence cites.
- **Vector conversion.** The remaining vectors (NRF-V01 to NRF-V05, NRF-V07, NRF-V10, NRF-V11, NRF-V13 to
  NRF-V17, NRF-V20, NRF-V22, NRF-V24 to NRF-V33) take the template shape with Inputs, Seeds, Draws,
  Trace, Outputs, Discriminates and runtime_observed fields; each case states its own starting
  state. Every expected value is unchanged. Inputs were made explicit where a case relied on another
  case's state: NRF-V01 (d) states `{0x6AC6935F, 0}`; NRF-V17 (b) states the stepped game seed;
  NRF-V11 (a) states its prior seed; NRF-V05 (g) states an empty array for the zero read; NRF-V15
  (b) names its tested state ids; NRF-V16 (f) names one removable list (its outputs follow
  NRF-R14b and equal the prose result "free, clear valid states, then clear expiryFlag"). NRF-V31
  no longer lists an unused callback input. NRF-V29, NRF-V30 and NRF-V33 cases were relabelled
  one case per sub-bullet; NRF-V30 keeps only its monster-slot and searcher cases and NRF-V33 only
  its state lookup and set-routing cases.
- **Wording.** Words run into numbers were separated, and the gap file's private method details
  were restated as original-module readings.
- **Checks.** `check_spec.py core-numeric-and-rng-foundations combat-status-application
  monster-scaling-by-players-and-difficulty` was run after the edits; its result is reported to the
  user with this change. Errors that remain are in
  [combat-leech-and-status-effects](../../specs/combat-leech-and-status-effects.md), whose owner
  updates its references separately.
- **Reader passes.** The blind-implementer pass is complete for this revision: full runs 1 to 3 and
  delta checks 1 to 3 are recorded below; delta check 3 returned no findings. This is not user approval
  or gameplay confirmation; Status remains draft. No gaps were
  closed; open gaps are in the [gap file](../../plans/gaps/core-numeric-and-rng-foundations.md).
  No label changed. Status remains draft.

<a id="split-reader-pass-1-2026-10-14"></a>
## Split revision: blind-implementer pass 1 (2026-10-14)

A fresh smaller-model reader used the split spec and its cited dependencies only, in a full run.
Counts: MISMATCH 0, GUESS 7, GAP 8, DISCRIM 11, GATED 30, COVERAGE 28 (advisory).

Source rereads made for the GUESS entries (2026-10-14, original instructions; independent listings of
these regions already existed, so the decoded bytes were checked against them):

- D2Game `0xE0110`: the no-parent branch (`0xE0168..0xE0185`) stores the parentless value at the
  unit's initial copy and passes it to the seed initializer, exactly as the parent branch stores
  the stepped low word.
- D2Game `0x77240`, slot loop at `0x77510..0x77630`: the chance byte is tested unsigned; zero skips;
  a value of 100 or more skips the draw; otherwise the draw result is compared signed and the slot
  continues only when the result is below the chance. Element type 10 draws modulo 5 and adds 1.
- D2Common `0x3AE60`: for a flagged list the state test (`0x3AE8C..0x3AE9F`) guards only the state-bit
  clear on the unit (call to `0x38730`) and the update request (call to `0x665C0`); the free at
  `0x3AEB6` is reached after that block whenever list flag bit 31 is clear.
- D2Common `0x323E0`: reads the unit's stat-list field at entry with no unit test, so a null unit
  faults; a null list yields 0 before the definition test.

| Entry | Outcome |
|---|---|
| GUESS 1-3 (R04b comparison, conversion, zero chance) | Spec edit from the `0x77240` reread; NRF-V30 (e), (f) added |
| GUESS 4 (R07 null unit) | Spec edit: null unit faults (`0x323E0`) |
| GUESS 5-6 (R14b target and nesting) | Spec edit: unit state-bit clear and update under the state test; free independent (`0x3AE60`) |
| GUESS 7 (no-parent initial copy) | Spec edit in NRF-R03a from `0xE0110`; NRF-V27 (f) asserts it |
| GAP 1 (V14 expiry) | Assertion removed; expiry arithmetic belongs to the lifecycle owner |
| GAP 2 (V06 values) | No change: the interpretations are stated in Outputs |
| GAP 3-5 (consumer rules in R04a, V10, V20) | NRF-R04a now names each row's consumer rule; consumers listed under Depended on by, not Depends on; V10's divisor clause removed; V20 drops the request-order claim, which is DSC-R09's |
| GAP 6 (ItemData sentinel) | `itemData.sentinel` added to section 3; NRF-V24 asserts -1 |
| GAP 7 (fault assertions) | Preamble defines a fault assertion |
| GAP 8 (class values) | NRF-V17 (a) and NRF-V24 state class 0 |
| DISCRIM 1 (mulDiv) | NRF-V04 (h)-(j). The suggested both-large case with equal operands gives the same result on either branch, so (h) uses `b` 0x400000 to separate branch order |
| DISCRIM 2 (V05 rounding) | NRF-V05 (i), (j) |
| DISCRIM 3 (state types) | NRF-V15 (e), (f) |
| DISCRIM 4-5 (mode setter) | NRF-V16 (h), (i) |
| DISCRIM 6 (object class) | NRF-V17 (e) |
| DISCRIM 7 (corpse min) | NRF-V20 (c) |
| DISCRIM 8 (scale, negative step, idle tick) | NRF-V22 (c); with these inputs the shifted results are 5814 and -5815 (truncation would give -5814) |
| DISCRIM 9 (tick count, wrap) | NRF-V27 (d), (e) |
| DISCRIM 10 (negative rate) | NRF-V32 (f) |
| DISCRIM 11 (secondary chain) | NRF-V33 (g) |
| DISCRIM 12 (equal chance) | NRF-V30 (e) |
| GATED (30) | No change: existing open questions and integration gates |
| COVERAGE (28) | No vector added beyond the DISCRIM fixes: the remaining branches repeat a result an existing case shows, are documented fault paths, or belong to a consumer or lifecycle owner |

The coordinator's list numbered twelve DISCRIM entries against a count of 11; all twelve were
resolved. All earlier expected values are unchanged.

<a id="split-reader-pass-2-2026-10-14"></a>
## Split revision: blind-implementer pass 2 (2026-10-14)

A fresh smaller-model reader made a full run on the pass-1 revision. Counts: MISMATCH 0, GUESS 5,
GAP 8, DISCRIM 11, COVERAGE about 30 (advisory).

Source rereads for the GUESS entries (2026-10-14, original instructions):

- D2Common `0x34510` (export 10945): returns with no call when the signed state is below 0 or not
  below the state count; otherwise calls the bit write `0x38730` and then `0x665C0`. There is no
  unit-type test. `0x38730` returns without a write for a null unit, a null list or a list without
  flag bit 31.
- D2Common `0x323E0` reads the unit's stat list and calls the selected-array getter `0x38A80`
  (call at `0x3242A`), the getter that selects resolved values when flag bit 31 is set and applies the
  minimum adjustment; it is the unit read, not the own-array list read `0x38B00`.
- D2Common `0x335F0`, prepared object slice: when the fixed-rate byte is nonzero the 16-bit base rate
  is stored at once (within `0x3378D..0x337CD`) with no draw and no clamp; the clamps occur only on the random
  branch (`0x337CE..0x3381D`).
- D2Game `0xE1D90`: for type 1 the class is tested signed against 0 and the MonStats count, the
  record address must be nonzero, and the record's flag byte must intersect an engine-held mask byte
  (`0xE1DA0..0xE1DDC`); type 0 tests the class unsigned against 7 (`0xE1DE8`). All rejections return
  null before the allocation call at `0xE1DF8`. The mask byte's meaning is not established; the spec
  names the per-class result `creationCondition` (NRF-Q09).
- D2Common `0x33920`: on a mode change the update request (`0x3393D`) precedes the changed-flag and
  mode writes; on an equal mode the update request precedes the changed-flag write.
- D2Common `0x3AE60`: the expiry flag is cleared at `0x3AEC5`, after the walk ends.

| Entry | Outcome |
|---|---|
| GUESS 1 (SetState bounds, type filter) | NRF-R12a states signed bounds and no type test; NRF-V33 and NRF-V31 state type 1 |
| GUESS 2 (diminish read) | NRF-R07 states the unit read; NRF-V05 (k) |
| GUESS 3 (fixed-rate clamp) | NRF-R18 states no clamp; NRF-V32 (h) |
| GUESS 4 (rejection order) | NRF-R17a states tests before allocation. The reread also found the monster record availability and `creationCondition` tests, added to NRF-R17a and section 3 with NRF-V17 (f) |
| GUESS 5 (call-time fields) | NRF-R14a and NRF-R14b state that requests observe fields written so far; NRF-V16 (c), (f) and NRF-V31 (a) assert the observed values |
| GAP 1 (format vector) | The format vector is removed; NRF-R08a's Id-map row cites NRF-V22, NRF-V29 (g) and MSD-V12 |
| GAP 2 (V29 direction lookup) | NRF-V29 (h) states that only the multiplication is tested |
| GAP 3 (frames to seconds) | Assertion removed from NRF-V14 |
| GAP 4 (named data example) | No change: the data-only rule uses table records |
| GAP 5 (unit types) | NRF-V31 and NRF-V33 state type 1 |
| GAP 6 (outputs outside section 5) | Section 5 gains the ItemData sentinel write; the rest are named intermediates, no change |
| GAP 7 (V20 min) | NRF-V20 Inputs say `min` 0 unless a case states it |
| GAP 8 (mock description) | No change: the preamble states persistence and logging |
| DISCRIM 1 (state bounds) | NRF-V15 (g), (h) |
| DISCRIM 2 (narrow product) | NRF-V04 (k) uses divisor 3: with the suggested divisor 1 the widened and narrow results share their low 32 bits, so that input does not separate them |
| DISCRIM 3 (b-side tie) | NRF-V04 (l), `mulDiv(0x10000, 0x10000F, 0x10000)`: 1048576 by divide-first against 1048591 widened (the suggested input does not separate them) |
| DISCRIM 4 (resolved versus own read) | NRF-V05 (k) |
| DISCRIM 5 (call-time snapshots) | NRF-V16 (c), (f); NRF-V31 (a) |
| DISCRIM 6 (negative scale) | NRF-V22 (d) |
| DISCRIM 7 (exact target) | NRF-V22 (e) |
| DISCRIM 8 (non-10 element at chance 100) | NRF-V30 (g) |
| DISCRIM 9 (chance 200) | NRF-V30 (h) |
| DISCRIM 10 (modulus 1) | NRF-V32 (g) |
| DISCRIM 11 (owner type 3) | NRF-V26 (b) |
| COVERAGE (about 30) | No further vectors: the branches repeat results already shown, are documented faults, or belong to consumers or the lifecycle owner |

Every suggested rejecting input was recomputed under both behaviors before it was added. Earlier
expected values are unchanged. Rule behavior changed (NRF-R17a's monster tests) and statements were
added to NRF-R07, NRF-R12a, NRF-R14a, NRF-R14b and NRF-R18, so full reader run 3 is required.

<a id="split-reader-pass-3-2026-10-14"></a>
## Split revision: blind-implementer pass 3 (2026-10-14)

A fresh smaller-model reader made a full run on the pass-2 revision. Counts: MISMATCH 0, GUESS 5,
GAP 6, DISCRIM 13, COVERAGE 14 (advisory). This is the third consecutive full run with MISMATCH 0.

Source rereads for the GUESS entries (2026-10-14, original instructions):

- D2Common `0x38520` (export 10871) returns null for a null stat list or a list without flag bit 31,
  then calls `0x37F80`, which returns null for state 0 and otherwise compares the state by equality
  along the primary chain (when flag bit 31 is set), then the secondary chain. There is no range test.
  The lifecycle contract's summary of this lookup omits the null-list and flag-bit-31 tests; that is
  reported to the coordinator for its owner, not changed here.
- D2Common `0x38730`: a nonzero value argument ORs the state bit; zero ANDs its complement. The
  mode-bound traversal passes 0.
- D2Common `0x33CD0`: state 0 passes the lower bound (`test` and `jl`), so `HasState(0)` reads bit 0.
- D2Game `0xE0110`: the caller supplies the parent (or null); the seeder tests it only after the
  initial seed write.
- D2Game `0x77240`: after a taken slot, element type 10 is replaced by the selector result plus 1;
  every other type goes to the element dispatch unchanged.

| Entry | Outcome |
|---|---|
| GUESS 1 (identity reduction) | NRF-R12b defines modulo 65536 as the unsigned low 16 bits |
| GUESS 2 (lookup bounds) | NRF-R12a: no range test in FindStateList; STL discrepancy reported |
| GUESS 3 (set argument) | NRF-R12a defines the `on` argument and the clear request's 0 |
| GUESS 4 (parent selection) | NRF-R03a states the supplied parent and its test after the initial write |
| GUESS 5 (selector and other types) | NRF-R04b states one draw, `result + 1`, and no draw for other types |
| GAP 1 (scale) | Section 3 movement row lists `scale` |
| GAP 2 (slot inputs) | Section 3 monster elemental slot row |
| GAP 3 (corpse inputs) | Section 3 corpse-explosion row, supplied by DSC-R09 |
| GAP 4-5 (logs) | No change to the vectors; the preamble states that logs are asserted |
| GAP 6 (log format) | Preamble defines log entries and observed fields |
| DISCRIM 1 | NRF-V04 (m) |
| DISCRIM 2 | NRF-V07 (f) |
| DISCRIM 3 | NRF-V15 (i) |
| DISCRIM 4 | NRF-V15 (j); state 0 is legal for HasState |
| DISCRIM 5 | NRF-V31 (c), (d) |
| DISCRIM 6 | NRF-V16 (f) observes mode 2 at the walk's requests |
| DISCRIM 7 | NRF-V17 (g) |
| DISCRIM 8 | NRF-V22 (f) |
| DISCRIM 9 | NRF-V22 (g) |
| DISCRIM 10 | NRF-V25 (g) |
| DISCRIM 11 | NRF-V26 (b) owner type 2 |
| DISCRIM 12 | NRF-V30 (i) |
| DISCRIM 13 | NRF-V32 (i) |
| COVERAGE (14) | No further vectors: the branches repeat results already shown, are documented faults, or belong to consumers or the lifecycle owner |

Each added input was recomputed under both the claimed and the rejected behavior. Earlier expected
values are unchanged. The rule edits state behavior the code already had and earlier vectors did not
contradict; no rule's behavior, ownership or gate boundary changed, so the next reader run can be a
delta check.

<a id="split-delta-check-1-2026-10-14"></a>
## Split revision: delta check 1 (2026-10-14)

A fresh reader checked the pass-3 changes. Counts: MISMATCH 0, UNRESOLVED 0, GUESS 1, GAP 2, DISCRIM 3.

Source rereads (2026-10-14): D2Common `0x34510` pushes its value argument unchanged to the bit write
`0x38730`, which tests it for nonzero; D2Game `0xE0110` tests the unit first and, when it is null,
reports an assertion and calls the fatal exit before any write.

| Entry | Outcome |
|---|---|
| GUESS 1 (SetState argument) | NRF-R12a: the argument is carried unchanged; NRF-V33 (l) |
| GAP 1 (null unit at the seeder) | NRF-R03a: fatal assertion before any write |
| GAP 2 (unstated observed fields) | Preamble: only stated observed fields are asserted; NRF-V16 (f) unchanged |
| DISCRIM 1 | NRF-V33 (h), (i) |
| DISCRIM 2 | NRF-V33 (j) |
| DISCRIM 3 | NRF-V33 (k) |

Each added input was recomputed under both behaviors. Earlier expected values are unchanged; no rule's
behavior, ownership or gate boundary changed.

<a id="split-delta-check-2-2026-10-14"></a>
## Split revision: delta check 2 (2026-10-14)

A fresh reader checked the delta-check-1 changes. Counts: MISMATCH 0, UNRESOLVED 0, GUESS 0, GAP 1,
DISCRIM 1.

| Entry | Outcome |
|---|---|
| GAP (untestable "before any write") | NRF-R03a now states only the fatal assertion. Instruction order is kept here as a record: D2Game `0xE0110` tests the unit at entry and calls the fatal exit before its first seed write |
| DISCRIM (null-unit seeder fault) | NRF-V11 (c): null unit and null parent give a fault |

No rule's behavior, ownership or gate boundary changed; earlier expected values are unchanged.

<a id="split-delta-check-3-2026-10-14"></a>
## Split revision: delta check 3 (2026-10-14)

A fresh reader checked the delta-check-2 changes. Counts: MISMATCH 0, UNRESOLVED 0, GUESS 0, GAP 0,
DISCRIM 0. No change was needed. The blind-implementer pass for the split revision is complete (full
runs 1 to 3, delta checks 1 to 3). Status remains draft: in-scope gates and gameplay observations stay
open in the gap file.
