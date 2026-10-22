# Knockback evidence

Behavioral contract: [Knockback](../../specs/combat-knockback.md).
Open attempts and blocked assertions: [gap file](../../plans/gaps/combat-knockback.md).

## Original artifacts and dated coverage

Evidence build is Lord of Destruction Expansion 1.13c, Windows x86. `LOD113C-D2GAME`
and `LOD113C-D2COMMON` resolve to the original filenames, sizes and SHA-256 fingerprints in
the [source manifest](../source-manifest.md). All descriptions below are analyst descriptions,
not original export names. RVAs require a PE-aware mapping; preferred addresses are not
assumed loaded addresses. No original binaries, complete table rows or gameplay captures
are distributed here.

The original static investigation was read on 2026-10-11 and its RNG/order distinctions
were refined on 2026-10-12. Compatible coverage of chance arithmetic, root tables, mode
dispatch and field loaders is reused with those dates. The original full spec and gap file,
their private source locations, fixture encodings and investigation history are preserved in
a separate hash-verified private archive. Relocating that history does not renew its proof.

On 2026-10-15 targeted reading addressed state-import identity, town membership, start-handler
calls, coordinate widths, waypoint writes, player mode preparation and prior speed/RNG claims.
New readings checked availability of byte-matched reconstructed C first and used disassembly
where none existed. Original-image decoding checked the stated jump/lookup data. The
available original-image fingerprints match the manifest; this does not establish historic
loaded DLL or archive identities beyond the manifest's recorded limits.

## Function and dispatch map

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | `0xACE60` | none | Item chance callback | KNB-R01, KNB-R10a | Disassembly, 2026-10-11/12; complete local body reread 2026-10-15; stat lookup consumes RLI-R02a |
| `LOD113C-D2GAME` | `0x3AE10` | none | Missile damage setup | KNB-R02a, KNB-R02b, KNB-R10b | Reused 2026-10-11/12 chance/order slice; SBA-R11 and MHD-R12 own surrounding missile integration; DPO-R22 owns owner monster multiplication |
| `LOD113C-D2GAME` | `0xDCF20` | none | Reaction dispatcher | KNB-R03a, KNB-R03b, KNB-R04, KNB-R05a, KNB-R06a, KNB-R06b, KNB-R10d | Disassembly entry through selection read 2026-10-15; HRC-R04a/HRC-R05/HRC-R06 consumed for shared reaction contracts |
| `LOD113C-D2GAME` | `0x771D0` | none | Base-78 GH conversion | KNB-R06c | Full disassembly reread 2026-10-15; state test via resolved import |
| `LOD113C-D2GAME` | `0x75160` | none | Monster mode-13 start | KNB-R06d, KNB-R07a | Full disassembly reread 2026-10-15; planner/cleanup/mode callees explicit boundary effects |
| `LOD113C-D2GAME` | `0x75070`, `0x75E10` | none | Monster mode parameters and table dispatch | KNB-R06d; KNB-Q06 | Reused 2026-10-11/12 roots; HRC-R09a/HRC-R15 delimit downstream mode integration |
| `LOD113C-D2GAME` | `0x78430`, `0x78760` | none | Player table and mode-19 start | KNB-Q07 | Reused root, complete 0x78760 read 2026-10-15 through call of 0x78610; full displacement not promoted |
| `LOD113C-D2GAME` | `0x78610` | none | Shared player mode preparation | KNB-Q07 | New full disassembly inspection 2026-10-15; original-image decode corroborated jump encoding; complete mode/planner/motion integration remains open |
| `LOD113C-D2COMMON` | `0x67390` | none | Path-type-8 endpoint generator | KNB-R07a, KNB-R08, KNB-R09a | Full disassembly reread 2026-10-15; clipping consumed as explicit supplied result |
| `LOD113C-D2COMMON` | `0x66C20` | none | Segment collision clipping | KNB-R09b; KNB-Q07 | Partial inspection 2026-10-15; byte-run decoding/interior callees not completely verified; no full collision claim |
| `LOD113C-D2COMMON` | `0x36230` | 10334 | Planner dispatcher | KNB-R06d, KNB-Q07 | Reused dispatch to path type 8; full planner effects not asserted |
| `LOD113C-D2COMMON` | `0x2FE10` | 10331 | Unit room getter | KNB-R03b | Full disassembly 2026-10-15; dynamic path room or null, static path room for kinds 2/4/5 |
| `LOD113C-D2COMMON` | `0x3C390` | 10057 | Room-to-level predicate wrapper | KNB-R03b | Full disassembly 2026-10-15; null room returns false, room chain supplies level to 0x2D320 |
| `LOD113C-D2COMMON` | `0x2D320` | none | Fixed town-level predicate | KNB-R03b | Disassembly plus hash-checked original branch/byte table 2026-10-15 |
| `LOD113C-D2COMMON` | `0x33CD0` | 10494 | State getter | KNB-R02b, KNB-R05a, KNB-R06c | Resolved import, consumes RLI-R03's verified state predicate |
| `LOD113C-D2COMMON` | `0x35040` | 10336 | Set reference unit | KNB-R06d; KNB-Q07 | Import resolved; reference-unit semantics reused from compatible prior reading |
| `LOD113C-D2COMMON` | `0x34F80` | 10647 | Set path type | KNB-R06d | Import resolved; compatible prior setter reading |
| `LOD113C-D2COMMON` | `0x34880` | 10527 | Store distance byte | KNB-R07a | Import resolved; compatible prior setter reading |
| `LOD113C-D2COMMON` | `0x33920` | 11090 | Mode write | KNB-R06d | Resolved import; RLI-R04a owns mode/animation effects |
| `LOD113C-D2COMMON` | `0x34510` | 10945 | State-on/off request routing | KNB-R05a; KNB-Q11 | Compatible RLI-R03/STL-R07 coverage reused with 2026-10-12/14 dates; full body checked 2026-10-15; disguise/update integration retains owning gaps |
| `LOD113C-D2COMMON` | `0x38730`, `0x665C0` | none | State membership write and update marking | KNB-R05a; KNB-Q11 | Compatible STL-R07/STL-R07a/RLI-R03 state primitive records, 2026-10-12/14; 2026-10-15 body readings do not verify unread disguise or room-queue effects |
| `LOD113C-D2COMMON` | `0x39430` | 10910 | Current stat getter | KNB-R01 | C availability checked; full disassembly read 2026-10-15; callback's guarded non-null owner maps to RLI-R02a/RLI-R02b selected-array/minimum leaf through 0x38A80 |
| `LOD113C-D2GAME` | `0xDA750` | none | Mode entry preparation | KNB-R06d; KNB-Q06 | C availability checked; full 96-byte disassembly and fingerprinted original-image decode 2026-10-15 establish caller and linked-record processing, not timer identity; full effects remain Q06 |
| `LOD113C-D2GAME` | `0x01180` | none | Bounded draw | KNB-R02a | Reused reading and NRF-R02 contract |
| `LOD113C-D2GAME` | `0x013B0`, `0x01560`, `0x01320`, `0x018A0` | none | Monster record, mode and base lookup helpers | KNB-R06a, KNB-R06c, KNB-R07a | Reused compatible 2026-10-11/12 lookup coverage; no new subtype contract inferred |
| `LOD113C-D2COMMON` | `0x7C8D0`, `0x56C10`, `0x73330` | none | Missile, size and item-event field loaders | KNB-R01, KNB-R02a, KNB-R10c | Reused 2026-10-11/12 loader descriptors through GDA-R09/GDA-R10/GDA-R12; text-to-compiled correspondence not implied |
| `LOD113C-D2COMMON` | `0x33110`, `0x34D40` | 10819; see HRC/MMF | Knockback rate and path-speed write | KNB-Q06, KNB-Q07 | HRC-R14 and MMF-R03 dependency evidence; not a new knockback speed derivation |

### Root chains and semantic mapping

Item root: D2Game callback selector table RVA `0x1077A8`, slot 7 at `0x1077C4`, contains
the callback `0xACE60`. The table lookup/registration/dispatch chain is the public
[item-event companion](combat-item-event-triggers-and-procs.md) and IEP-R02/IEP-R03/IEP-R07.
The callback's supplied stat identity is looked up before its inline seed step. It tests
monster size via the MonStats2 helper `0x01320`, with large before small, then ORs result
byte bit 8. The stat identity argument is not proof that code fixes stat 81.

Missile root: D2Game `0x3AE10` presets hit, tests target state, sets GH/soft flags and performs
the `KnockBack` byte comparison. Its later calls reach defender block/avoidance, owner
multiplier `0x770E0`, direct application `0xDE0C0` only if hit remains set, then direct reaction
`0xDCF20` regardless of that hit bit. See SBA-R11, MHD-R12 and DPO-R22 for those stage
contracts. DPO-R26 instead owns queued-path reaction placement. Original
`Missiles.KnockBack` field offset is `0xAE`.

Reaction root: `0xDCF20` reads target room through import thunk `0xA478` -> D2Common
#10331 `0x2FE10`, then calls thunk `0xA808` -> #10057 `0x3C390`. That wrapper extracts the
level through room/level links and calls `0x2D320`. It is not a direct import of `0x2D320`.
The source permission test reads `MonStats`, not `MonStats2`; the original field is bit 10
of the MonStats flag word at `0x0C`. With a known valid source record, it proves KNB-R03b.
After this gate the state test precedes kind dispatch. Monster fallback rewrites the u16
result, then calls conversion `0x771D0`; player/monster selection follows HRC-R04a.

Monster root: `0xDCF20` builds mode parameters through `0x75070`, dispatches through `0x75E10`
and mode table RVA `0xFA498` (16-byte stride), whose mode-13 start is `0x75160`.
The start handler checks availability bits 3 and 13, then writes type 8 and distance 5/10,
calls planner #10334 `0x36230`, mode entry preparation `0xDA750`, and mode setter #11090 `0x33920`.
The planner's return is ignored. This proves ordered local calls; it does not prove all
effects or seed behavior of the entire mode dispatcher.

Path root: planner `0x36230` dispatches path type 8 to `0x67390`. The generator's context
supplies start and explicit point; the dynamic path supplies reference unit and distance.
Distance is byte `0x90`; reference unit `0x58`; planned endpoint words `0x10/0x12`;
first waypoint `0x9C`; count `0x28`. Source kinds 2/4/5 read full dwords `0x0C/0x10` from
their static path; dynamic access reads zero-extended words `0x02/0x06` or zero if missing.
The generator stores requested point, invokes `0x66C20`, stores clipped plan and clears
reference. Count 1 and clipped waypoint are written only on the nonidentical branch.
KNB-R09a explicitly supplies clip output; unread collision interior remains KNB-Q07.

Player root: reaction requests player mode 19 via the player dispatch `0x78430`, whose
mode-19 handler is `0x78760`. When not already mode 19 it sets reference unit via #10336
`0x35040` and calls `0x78610`. That helper contains type-8/distance-5 preparation and planner
calls, but the full mode-19 jump selection, path success gate, notifications and later ticks
have not been promoted to a verified player displacement contract. KNB-Q07 records this
targeted progress rather than treating the older unanswered question as complete absence.

### Resolved import corrections

The previous `0x968E` citation was an assembly label, not a module RVA. Actual D2Game thunk
RVA `0xA68E` resolves to D2Common #10494 `0x33CD0`, establishing the state getter already
verified under RLI-R03. State-54 missile/reaction gates can therefore use a static-code label.
Related thunks: `0xA5FE` -> #10336 `0x35040` reference unit; `0xAAFC` -> #10647 `0x34F80`
path type; `0xAD24` -> #10527 `0x34880` distance; `0xAAF0` -> #10334 `0x36230` planner;
`0xA4A2` -> #11090 `0x33920` mode; `0xA3CA` -> #10945 `0x34510` SetState routing;
`0xA802` -> #10910 `0x39430` stat getter. The lethal state-54 caller pushes 92 and 1, not a
proven ordinary death invocation. Compatible RLI-R03 and STL-R07 identify the request as
SetState(target,92,1), meaning state 92 on. Only disguise bookkeeping and update-queue
integration remain KNB-Q11 through STL-R07a/STL-Q07 and RLI-Q02.

### Compatible state-92 ownership reuse

The public [state lifecycle record](state-timed-state-lifecycle.md#function-references)
already covers D2Common #10945 `0x34510` -> `0x38730` state-bit routing and `0x665C0`
update marking, with state-write reading dated 2026-10-12 and original-image byte-tail
checks dated 2026-10-14. The [foundations contract](../../specs/core-numeric-and-rng-foundations.md)
RLI-R03 and [state lifecycle contract](../../specs/state-timed-state-lifecycle.md)
STL-R07/STL-R07a own the operation. Those compatible original-artifact records are reused
at their actual dates; a 2026-10-15 C-availability check followed by complete disassembly
readings of all three bodies confirms the caller mapping without establishing additional
disguise-helper or room-queue effects.

The current local assertion is the ordered SetState request, using the owner's guards and
membership contract. It is not ordinary death-mode execution. The previous broad local
question that treated the whole state operation as unread is closed for identity, routing
and membership; KNB-Q11 now tracks only the owning contracts' STL-Q07 disguise and RLI-Q02
update-queue integration limits. V12 retains a supplied request boundary rather than
duplicating the state primitive's fixtures or claiming an observed membership change.

### Stat getter and mode entry preparation, 2026-10-15

The item callback's actual stat import is D2Common #10910, RVA `0x39430`. Its full
disassembly was read after checking reconstruction availability.
It dereferences its unit argument, so the callback's prior non-null owner guard matters.
It returns zero for a null stat list or rejected signed stat bounds, zero-extends the layer
to u16, forms the original lookup identity as `(stat << 16) | layer`, and calls `0x38A80`.
That selected-array/minimum leaf consumes RLI-R02a/RLI-R02b. This does not claim a call to
the separate ordinary unit wrapper at `0x38B70`, whose null-unit behavior is not this
import's own guard. KNB-R01's supplied current stat and callback guard results are unchanged;
the source-to-shared-leaf mapping is now explicit.

The full 96-byte D2Game `0xDA750` body was read after checking C availability and corroborated
with a hash-checked original-image decode. It processes a unit-associated linked record list,
compares entries' unit kind/id with the supplied unit and unlinks matching records. These
records were not established to be timers. The contract therefore names the invocation
mode entry preparation, with its complete effects still KNB-Q06. The supplied acceptance
operation logs its call and does not assume timer cancellation or other unverified effects.

### Town table check, 2026-10-15

The fingerprinted D2Common image's function `0x2D320` subtracts 1, rejects unsigned values
above 108 and selects between return-1/return-0 branches. Branch destinations at `0x2D340`
resolve to `0x2D334` and `0x2D33A`. The complete 109-byte lookup at `0x2D348` has zero entries
at indices 0, 39, 74, 102 and 108; those select return 1. Thus the level set is exactly
1, 40, 75, 103 and 109, with outside-range false. The room getter and wrapper were then read
to close the caller-to-predicate chain. This is an engine-code table, not a copied game-data
column or a claim of a gameplay town observation.

## Item stat text cells, 2026-10-15

The effective virtual file was `data/global/excel/itemstatcost.txt`, read through the game's
own Storm loader. Requested offset 0, length 0 (complete file); returned 40109 of 40109 bytes.
The complete decoded file SHA-256 is
`ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375`.
Tab-separated headers selected named columns; the selected row is `Stat = item_knockback`,
explicit `ID = 81`, zero-based data-row index 81 excluding the header. Named cells only:

| Column | Raw cell | Interpretation |
|---|---|---|
| `ID` | `81` | Text identity 81 |
| `Stat` | `item_knockback` | Named stat row |
| `itemevent1` | `domeleedamage` | First text event name |
| `itemeventfunc1` | `7` | First callback selector |
| `itemevent2` | `domissiledamage` | Second text event name |
| `itemeventfunc2` | `7` | Second callback selector |

Label: verified in 1.13c data, within the source-profile limitations shared by
[the pipeline's fresh named-cell record](combat-damage-pipeline-order.md#effective-file-named-cell-corroboration-2026-10-15).
The member reported `lod-113c` and its historical profile supplies vanilla identification;
the current version catalog was unavailable. No fresh complete override inventory, newly
loaded DLL fingerprint set, physical supplying MPQ attribution, compiled-record correspondence
or event-name numeric binding was collected. The file hash identifies these effective text
bytes only. A read is not a gameplay observation, and these cells do not alone close
KNB-Q01/KNB-Q02 or IEP-R04 automatic registration.

## State 92 text name, 2026-10-15

The effective virtual file `data/global/excel/states.txt` was read through the game's Storm
loader at offset 0, length 0 (complete file): 20920 of 20920 bytes returned. The complete
decoded-file SHA-256 is `898806c2b69e55e316ecfed48dde95aa2933829b94fdad234bdabc27bdbc23e1`.
Header-selected columns `id` and `state` identify the unique explicit id 92, also zero-based
data-row index 92 excluding the header. The named raw cells are `id = 92` and
`state = death_delay`. Label: verified in 1.13c data, at this text-cell boundary.

The member reports `lod-113c`; vanilla profile identity is reused with the same fresh
override-inventory, physical supplying MPQ, compiled-correspondence and loaded-DLL limits
as the [pipeline named-cell record](combat-damage-pipeline-order.md#effective-file-named-cell-corroboration-2026-10-15).
No gameplay action, hook change or table mutation was performed. Code fixes the requested
numeric id 92; this named cell identifies its text name. Neither that name nor the request
proves a scheduled delay, a timer, or ordinary death-mode execution. The remaining state
integration gates are STL-Q07 disguise bookkeeping and RLI-Q02 update-queue effects.

## Corrections and affected assertions

The 2026-10-15 review checked each prior question, weak label and non-executable vector
against compatible records before targeted reading. Presentation does not itself promote
evidence. The original numeric item/missile examples and 3/6 endpoint arithmetic remain;
their execution assertions are now limited to what the evidence establishes.

| Discrepancy | Evidence and correction | Affected current assertions / remaining limits |
|---|---|---|
| State-test import identity inferred from a label | Resolved actual thunk/import to #10494 `0x33CD0`, matching RLI-R03 | KNB-R02b and KNB-R05a local gates verified; only state-operation disguise/update integration remains Q11 |
| Town set and wrapper unresolved | Complete predicate/table and getter/wrapper reading above | KNB-R03b and V27 close supplied valid-room/record town gate |
| All source coordinates described as 16-bit | `0x67390` reads static source dwords, dynamic words | KNB-R08/V29 preserve full static coordinates; V23 numbers unaffected |
| Zero clip described as no waypoint/zero completed displacement | Generator leaves count unchanged on identical endpoint and writes plan rather than actual position | KNB-R09a/V13/V30 expose count and stored requested point; full blocked outcome Q07 |
| No speed constant asserted from handler/generator search | HRC-R14's downstream rate branch sets path speed `0x1000` | Shared dependency supplies 4096; no absence label survives; full movement still Q07 |
| Mode changes/clipping described as consuming no RNG | Local direct draws absent; HRC-R15 reports conditional elemental draws in monster mode dispatch | KNB-R10a limits assertion to direct operations; complete event draw count Q06 |
| Both kinds described as death-first | HRC-R04a and `0xDCF20` player prefix put avoidance/block before death | KNB-R04/KNB-R06b/V28 use actual kind-specific priority |
| GH veto implied after both item/missile rolls for every hit | GH-veto call exists only when selection reaches GH; item timing still lacks compiled binding | KNB-R10d/V25 conditional; Q02 retained |
| State-54 lethal helper called death without its consumer read | Resolved #10945 `0x34510`, then reused compatible RLI-R03/STL-R07 state-on routing and membership records | KNB-R05a/V12 SetState(92,1) request; Q11 only disguise/update integration, not primitive identity or membership |
| Complete subtype equivalence inferred from no local branch | Local kind dispatch established, subtype records/producers not traced | KNB-R03a verified locally; Q08 outside-boundary |
| Early vectors gave rates or predicates only | Fixed seeds, exact outcomes, call logs and endpoint fixtures added | V01–V26 retain numeric facts; V27–V34 discriminate corrections and result-distinct branches; not observations |
| Mode-start helper described as event cleanup | Full 0xDA750 reading establishes linked-entry processing without timer identity | KNB-R06d/V13 name the neutral preparation request; complete effects remain Q06 |
| Zero-byte fixture described as a decoded blank | Archive blank decoding was not proved by this supplied-stage check | V07 now supplies compiled byte 0 directly; the zero-byte behavior and outputs are unchanged |
| Missile direct reaction cited through queued placement | MHD-R12 explicitly owns 0x770E0, hit-conditional 0xDE0C0 and subsequent unconditional 0xDCF20; DPO-R22 owns the owner multiplier | KNB-R10b/KNB-R10d and trigger/dependency citations now distinguish missile direct placement from DPO-R26 queued placement; no DPO behavior changed |

Q03 block/stun is answered by HRC-R04a/HRC-R05: monster KB beats block/GH, player block
preempts KB, stun is consulted only in GH branch. Base identity/name uses GDF-R08; state
identity uses GDF-R03. Coordinate scale reuses NRF-R10; completed measurement remains Q04.
The mode-start refusal Q06 is locally answered but its caller consequences remain open.
The client-conflict search never covered the full client; its old absence claim is replaced
by outside-boundary Q10. Collision inspection found insufficient verified interior coverage;
no generic six-rung absence claim is made.

## Pending observation procedures

No gameplay action or trace was performed for this regeneration. Every O id is a future
procedure, with class/level and stat/data values to be recorded from a controlled legally
owned LoD 1.13c test profile before acceptance. Use Normal difficulty, an unmodified level-20
Amazon, no mercenary, no passive avoidance or gear granting unrelated combat effects unless
the named case needs it. Record actual level/class/items/stats, effective data identity,
target class/base/size/mode/states, room/level, all seeds and a nonlethal hit's flags. If the
setup cannot supply the stated byte/stat/target safely, report that case unobserved rather
than changing a live data profile or inferring its outcome.

- **KNB-O01:** In a cleared outdoor area, stand at least two tiles from obstacles and attack
  a controlled small, normal and large monster with item stat 81 positive. Record the exact
  item and effective stat; use the same attack, target life and difficulty. Observe callback
  `D2Game+0xACE60` entry/return, attacker seed and result write over a predeclared 128 eligible
  callbacks per target. Check each masked draw against its threshold; separately count selected
  KB and completed motion. Empirical rate alone cannot verify seed order or exact probability.
- **KNB-O02:** Use an existing lawful controlled fixture whose effective missile byte is 1,
  33 or 75; record missile class and that named cell. Hit the same nonlethal interruptible
  target on open ground, recording `D2Game+0x3AE10` chance-stage seed/result over 100 eligible
  hits for each available byte. Compare each remainder with the byte, separating acceptance
  from later block/reaction. Unavailable byte cases remain pending.
- **KNB-O03:** Position the same attacker and a default monster, then base-78 monster, on
  an open straight line with at least two tiles behind target. Record generator
  `D2Common+0x67390` start/source/B/requested/clipped endpoint and waypoint count, then actual
  unit positions on every server tick until the mode finishes. Expected construction is
  step 3 or 6; completed movement is measured separately. This resolves Q04 only with valid
  coordinate attribution and distinguishes it from unresolved movement integration.
- **KNB-O04:** Repeat O03 beside a wall behind target, then with only part of the requested
  segment available. Record `D2Common+0x66C20` arguments/result, initial/final waypoint count,
  planner return, mode and every actual-position update. An identical clipped endpoint
  predicts generator return 0, but do not infer a cleared prior count or completed no-motion
  result without reading it. This is the distinguishing check for Q07.
- **KNB-O05:** Use a controlled state-54 target and compare nonlethal/lethal result requests
  at `D2Game+0xDCF20`; record state operation and distinguish it from death-mode dispatch.
  Separately compare GH-selected and KB-selected hits against `D2Game+0xDB910`; no GH-veto
  call is expected on KB selection. Record count and seed owner rather than assuming two draws.
- **KNB-O06:** Repeat an identical nonlethal supplied hit against a controlled hireling,
  summon and NPC monster, recording their actual records/modes and `D2Game+0xDCF20` routing.
  Compare mode-0/mode-12 monsters as controls. Local dispatch should return for dead modes;
  matching one subtype case does not establish every producer/record or full Q08 equivalence.

Every capture must establish matching loaded module/profile identity, eligibility, valid
record count and full requested fields. Zero records or incomplete captures do not promote
a rule. No 1.14d equivalence is claimed or investigated.

## Regeneration record, 2026-10-15

Scope is the existing knockback contract only. No implementation, gameplay campaign, plan,
original-source edit or new behavior boundary is included. The contract is unconsumed by
implementation plans/code/tests; identifiers referenced by sibling contracts are retained
without historical aliases. All existing 26 vector identities remain, with concrete fixtures;
ten additional vectors cover targeted corrections and result-distinct branches. Historical implementation comparisons
remain explicitly historical in the gap file and were not audited against current code.

New evidence closes the text-row question, state import identity, town level set and wrapper,
local coordinate-width and waypoint-write ambiguities, and block/stun selection through shared
contracts. Compiled binding/timing, dependent mode/player/collision/movement effects, state
disguise/update integration, subtype/client ownership and gameplay observations remain as precise
[gap rows](../../plans/gaps/combat-knockback.md#open-evidence-gaps). No full-behavior label
or gameplay confirmation is claimed. Publication and user approval remain pending.

### Fresh-reader full pass 1 and triage

The first fresh reader used the main contract and cited dependency contracts without private
source, implementation, runtime or web access. It recomputed all 30 then-current executable
vectors. Counts: **MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; GATED 44; COVERAGE 11; HARDEST 3**.
There are no mismatch/guess/gap/discrimination findings requiring resolution. The 44 gated
entries are Q01/Q02/Q04/Q06/Q07/Q08/Q10/Q11/Q12, R09b/R10c, HRC-R04b/HRC-R11/HRC-R15,
and the supplied/isolated boundaries of V01–V30. Outcome: retain every documented gate;
stage fixtures do not verify dependencies or gameplay. These are limits, not additional
missing decisions.

| Reader coverage entry | Outcome |
|---|---|
| KNB-R01 missing callback references / absent size record | Added V31's failed-reference guards and absent-record default-threshold draw |
| KNB-R03b town level 103 | No change: this is another constant selecting the same table branch; V27 already distinguishes membership, permission and both unsigned bounds |
| KNB-R04 lethal versus block, including player prefix | Added V11c/d with identical lethal+KB+block flags and different monster/player selection |
| KNB-R06a fallback retaining other flags | Added V32a, retaining block while replacing KB with GH |
| KNB-R06b avoidance / soft / plain branches | No change: their complete terminal effects belong to HRC and retain HRC-R04b/R11 limits; this local contract supplies selected operations and already distinguishes KB/GH/player-block priority. No extra effect is inferred for non-KB terminal branches |
| KNB-R06c invalid base / conversion after fallback | Added V32b/c to distinguish conversion order and invalid lookup |
| KNB-R06d missing unit / mode 0 | Added V33 for return-1 no-op success after availability checks. No additional absent-unit recipe: the same post-availability no-op branch has identical result/writes; automatic availability lookup for an absent unit is not supplied by this fixture |
| KNB-R08 low16 endpoint wrap / extreme i32 | Added V34a/b/c for endpoint wrap, multiply wrap and signed absolute/max behavior |
| KNB-R10a full mode RNG | No change: full mode-change draws are explicitly gated by Q06/HRC-R15, separate from direct local draws |
| KNB-R10b other supplied stage order | No change: V24 already asserts the complete ordered log `[block,avoidance,multiplier,apply,reaction]`, and every stage sees the stepped missile seed/result; reordering any pair fails that assertion |
| KNB-R10d GH defender draws | Added V25c with an explicitly supplied one-draw GH operation and ordered missile/defender draws; it does not duplicate HRC-R06 threshold ownership |

| Reader hardest entry | Outcome |
|---|---|
| KNB-R08 / V23 / V29 mixed widths, wrap and division | Added V34 to make the independently computable width cases concrete; the arithmetic contract is unchanged |
| KNB-R06b/KNB-R06c fallback and conversion order | Added V32's retained-bit and re-conversion cases; no change to rule order |
| KNB-R03b / V27 unsigned bounds and source permission | No change: V27 already supplies membership, zero/outside-range and allowed/disallowed source cases with exact continuation outputs |

Two additional fixture clarifications remove unsupported assumptions: V07 supplies a compiled
zero byte rather than asserting a blank-field decoder, and V09 explicitly replaces the default
no-write planner with the composed generator stage using the just-stored distance byte.
V13's helper log is renamed mode entry preparation without assuming timer identity. The
selected-array/minimum stat dependency citation now follows the actual import and leaf above.

These edits required a second fresh **full** reader pass. That contract had 34 vectors;
changed fixtures are V07, V09, V11, V13, V25 and new V31–V34. The R01 dependency citation,
R06d's neutral helper name, Id-map vector references and acceptance range were updated to
match. No gate or numeric rule was promoted at that handoff. Full pass 2's results follow;
pass 1 does not approve the changed contract.

### Fresh-reader full pass 2 and triage

The second fresh full reader recomputed all 34 vectors and used the same restricted public
contract/dependency scope. Counts: **MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; GATED 11;
COVERAGE 8; HARDEST 3**. The 11 gated entries group R09b/R10c and the nine local open
questions. This count groups boundaries differently from pass 1; it does not remove the
documented dependent-operation or observation limits. Outcome: retain those limits, with
Q11 narrowed by compatible state-owner evidence as described above. There are no mismatch,
guess, gap or discrimination findings requiring resolution.

| Reader coverage entry | Outcome |
|---|---|
| KNB-R01 successful OR preserving flags | Added V35a with result 16 becoming 24 |
| KNB-R01 non-81 supplied stat identity | Added V35b: supplied synthetic identity 82 is positive while 81 is zero, proving callback selection without claiming a text-row binding |
| KNB-R04 avoidance before lethal | No change: player avoidance's complete eligibility/effects belong to HRC-R04a/HRC-R04b; the local supplied selection contract retains that prefix and its gate, and V11/V28 cover the distinct KB/death/block outcomes |
| KNB-R06b avoidance branch | No change: same HRC-R04b eligibility/effects boundary; do not invent an avoidance default for local fixtures |
| KNB-R06b GH versus soft/plain | No change: these terminal non-KB paths consume HRC's selected-operation contract; their full effects are outside the KB stage. Existing V19 rejects selecting GH when KB is present |
| KNB-R06d null unit after availability | No change: the same return-1/no-local-write branch is exercised by V33's mode-0 case; this does not provide automatic availability for a null unit |
| KNB-R06d mGH true / mKB false | No change: both availability failures reach the identical return-0/no-write result already exercised by V21; duplicating the second primitive guard adds no result-distinct recipe |
| KNB-R08 missing dynamic source path | Added V36: reference unit remains present, coordinates default to zero, and requested endpoint is `(103,103)` rather than unused explicit point `(7,7)` |

| Reader hardest entry | Outcome |
|---|---|
| V34 integer wrapping | No change: wrapping subtraction/absolute/multiply and signed max/division are explicit, with exact independent a/b/c endpoint outputs |
| V32 fallback and re-conversion | No change: the trace and a/b/c outputs distinguish bit retention, conversion order and invalid lookup |
| V25 supplied defender-draw order | No change: c declares the supplied operation, ordered seed owners and exact log; it explicitly does not verify HRC threshold selection |

The Q11 ownership/gate correction and the new V35/V36 cases required fresh **full pass 3**.
Changed material is KNB-R05a's dependency ownership/gate, Q11's precise remaining boundary,
V12's SetState request name, new V35/V36, the state-92 text-name citation and the acceptance range V01–V36. Full pass 2 is
historical support for its reviewed contract, not approval of these changes.

### Fresh-reader full pass 3 and triage

The third fresh full reader used the same public contract/dependency scope and recomputed
all 36 vector families. Counts: **MISMATCH 0; GUESS 0; GAP 2; DISCRIM 4; GATED 13;
COVERAGE 5; HARDEST 3**. No mismatch or guessed decision was reported. The 13 documented
gate/boundary entries remain explicit: no stage fixture establishes unresolved collision,
compiled timing, dependent bookkeeping or gameplay confirmation.

| Reader gap | Outcome |
|---|---|
| KNB-R10b cited queued DPO-R08d/DPO-R26 for missile multiplication/direct reaction | Corrected to MHD-R12 for the actual missile path and DPO-R22 for owner multiplication. Queued DPO-R26 remains only where that path is named; trigger, R10d and dependency references agree |
| Section 5 omitted R03b/V27 continuation/stop result | Added the town-gate continuation/stop effect to the outputs table |

| Reader discrimination finding | Outcome |
|---|---|
| V06's absorbing zero seed cannot detect an omitted RNG step | Retained case a and added b with seed `{1,15}`, new `{0x6AC690D4,0}`, remainder 0 and result 8 |
| V13 did not distinguish forwarding a mode-setter return | Added legitimate already-mode-13 case c using RLI-R04a's actual equal-mode branch: update request, supplied-model modeChanged write, setter return 0, handler return 1. The update operation logs only and engine update integration stays RLI-Q02 |
| V21 only rejected omission of the mGH guard | Retained a and added b with mGH true/mKB false, requiring the same refusal and no-write result |
| V27 did not reject trusting a permission bit without a source record | Added i with a monster whose record lookup supplies no valid row despite a detached true permission flag; continuation false. Source/record presence is explicit in section 3. Added j for absent source, also false |

| Reader coverage entry | Outcome |
|---|---|
| KNB-R02a/KNB-R02b pre-existing KB preservation | No new recipe: a skipped chance stage has no write; V07 supplies a nonzero preserved result and V12b supplies result 8 on the reaction state gate. The same no-clear stage contract applies to an already-set KB bit; no success/failure path clears it |
| Missing town source/record | Added V27i/j and explicit source/record-presence inputs; no new behavior gate |
| mGH-only / missing-unit start branches | V21b now rejects mGH-only availability checks; missing-unit post-availability no-op shares V33's return-1/no-write branch and does not receive an invented automatic availability result |
| Avoidance / soft / plain reactions | No change: their full eligibility/effects are owned by HRC and retain existing gates; local KB/GH/block/death selection assertions remain separately executable |
| Static source kinds 4/5 | No change: they use exactly the same full-i32 source-coordinate branch as kind 2, already distinguished by V29/V34 |

The requested signed-division fault recipe has no reachable case in the declared numeric
domain. Wrapping absolute values are nonnegative or minimum i32; their signed maximum can
be negative only at minimum i32, never -1. Zero skips division. Thus neither division by
zero nor minimum-i32 divided by -1 occurs. No fault vector was fabricated; R08 now states
this bound explicitly while retaining every arithmetic result.

| Reader hardest entry | Outcome |
|---|---|
| KNB-R08 / V34 widths, wrapping and division | No change to arithmetic or existing fixtures; retained exact wrapping cases and clarified why no division fault recipe exists |
| KNB-R06a/KNB-R06c/KNB-R06b / V32 fallback and conversion | No change: existing a/b/c results distinguish retained flags, re-conversion order and invalid lookup |
| KNB-R05a / V12 NRF/STL state-92 boundaries | No change: V12 remains a supplied SetState request; RLI-R03/STL-R07 own membership, STL-Q07/RLI-Q02 retain their precise integration gates |

The dependency/source-presence clarifications and new subcases required fresh **full pass 4**,
not a delta check. The main contract remains 36 vector families. Changed material is the
MHD-R12/DPO-R22 versus queued-DPO-R26 dependency ownership; section-3 source-record and
modeChanged inputs; section-5 town continuation and supplied-mode-model boundary; R08's
numeric-domain explanation; V06b, V13c, V21b and V27i/j. No gate is promoted and no gameplay,
implementation or original-source work is included. The subsequent review records follow.

### Additional contract review 4: prior-report exposure

Review 4 returned **MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0; GATED 9; COVERAGE 2;
HARDEST 3**. That reader also viewed prior reader reports embedded in this companion.
Its result is therefore recorded as an additional contract review with prior-report
exposure, not an uncontaminated blind-implementer pass. It does not establish that the
contract independently passed a fresh blind reading. The nine gated entries summarize the
current open questions; they do not erase other documented dependency/stage limits.

| Review-4 advisory entry | Outcome |
|---|---|
| Coverage: player avoidance before lethal | No change: HRC owns avoidance eligibility/effects, with the existing gate retained |
| Coverage: soft/plain terminal branches | No change: these terminal effects are HRC-owned; no additional KB effect is inferred |
| Hardest: KNB-R08 / V34 widths | No change: the signed widths, wrapping operations and exact a/b/c outputs are explicit |
| Hardest: KNB-R06a/KNB-R06c / V32 conversion | No change: retained flags, fallback before re-conversion and invalid lookup are distinguished |
| Hardest: KNB-R10d / V25 GH-draw order | No change: supplied operation, seed owner, ordered draws and log are explicit; HRC threshold selection is not claimed |

### Fresh-reader full pass 5 and triage

A new full reader was restricted to the main contract and cited dependencies. It did not
receive this companion or previous reader reports, and did not use original-source,
implementation, runtime or web material. Reading the governing checklist instructions
provided workflow guidance, without prior-report or behavioral-evidence exposure.

Counts: **MISMATCH 0; GUESS 0; GAP 0; DISCRIM 5; GATED 38; COVERAGE 5; HARDEST 3**.
The initial GATED 0 report omitted documented limits and was corrected to 38 grouped entries:
nine current open questions; eight rule-limit rows R04/R05a/R06b/R06d/R09a/R09b/R10a/R10c;
and the 21 supplied-stage families V08/V09/V10/V11/V12/V13/V19/V20/V21/V23/V24/V25/V26/
V27/V28/V29/V30/V32/V33/V34/V36. The corrected count is used here. These entries remain
limits on actual dependency execution and full parity; local verified assertions are not
blanket-gated by being listed. No mismatch, missing decision or missing input/output was
reported.

| Reader discrimination finding | Outcome |
|---|---|
| V02's prior roll did not distinguish threshold 64 from 32 | Retained a and added b: player target, seed `{0,48}`, roll 48, threshold 64 accepts, result 8/return 1/new seed `{48,0}` |
| V03's prior roll did not distinguish large threshold 32 from default 64 | Retained a and added b with both size flags and the same roll 48: large-first threshold 32 rejects, result 0/return 0/new seed `{48,0}`; default 64 and small-first 128 would accept |
| V07 did not reject clearing an already-set KB bit on zero chance | Retained a and added b with result 12/chance 0; result 12 and missile seed `{1,0}` remain, with no draw |
| V12 did not reject clearing an already-set KB bit in the missile state gate | Retained a–c and added d: missile chance 255/state 54/result 12 remains 12 with no draw and no call |
| V27's gate-only cases did not distinguish town-before-state order | Retained a–j and added composed reaction case k: level 1/source absent/state 54/lethal result 10 rejects at town gate, log remains empty, result/mode/path/seeds remain unchanged and no SetState(92,1) request occurs |

| Reader coverage entry | Outcome |
|---|---|
| Normal/player threshold acceptance between 32 and 63 | Covered by new V02b's roll 48 acceptance |
| Large-first threshold rejection between 32 and 63 with both size flags | Covered by new V03b's roll 48 rejection |
| Zero-chance preservation of existing KB | Covered by new V07b's unchanged result 12 and no draw |
| Missile state-gate preservation of existing KB | Covered by new V12d's unchanged result 12 and no draw |
| Town rejection before lethal uninterruptible-state request | Covered by new composed V27k's empty call log and unchanged supplied state |

| Reader hardest entry | Outcome |
|---|---|
| KNB-R08 / V34 widths and wrapping | No change: signed widths, low16 endpoints, wrapping products and the signed-max edge are explicit with exact outputs |
| KNB-R06b / V11/V19/V28/V32 priority and conversion | No change: kind-specific priority and fallback/re-conversion order are explicit and distinguished by these fixtures |
| KNB-R06d/KNB-R09a / V13/V30 ignored returns and path writes | No change: V13c uses the legitimate equal-mode setter return 0 with handler return 1; V30 distinguishes requested waypoint/count retention from clipped-plan writes; supplied-operation gates remain |

All five discrimination findings have fixture edits, and every coverage/hardest entry has
an outcome. Only subcases V02b/V03b/V07b/V12d/V27k were added; original cases and numeric
results remain. No rule, ownership or gate changed, and the contract remains 36 vector
families. With full pass 5 reporting MISMATCH 0 and every finding triaged after the third
full run, the next check is eligible as **delta pass 6**, covering these five added cases
and their discrimination findings against the same main/dependency contracts. Neither the
exposed review 4 nor this triage approves the changed fixtures; delta results follow.

### Fresh-reader delta pass 6

2026-10-15: a fresh smaller reader checked the actual main contract, its cited behavioral
dependencies and the governing checklist. It accessed no companion, prior report, private
source, implementation, runtime or web material. The initially omitted checklist path was
provided explicitly; the reader then read section 2 and reported no remaining access limit.
Changed material was V02b/V03b/V07b/V12d/V27k; rules, ownership and gates were unchanged.

Counts: **MISMATCH 0; GUESS 0; GAP 0; DISCRIM 0**. Each list has no unresolved entry.
The independent check covered the two roll-48 threshold contrasts, both result-12 skip
preservations and the composed town-before-state rejection. These assertions reject every
incorrect alternative identified by full pass 5. Existing documented limits and the
earlier advisory dispositions remain. This completes the applicable reader check for the
current contract; it does not establish user approval, gameplay confirmation or full parity.

### Checks

Before full pass 1, the targeted specification lint ran with exit 0, **0 errors, 0 warnings**,
and scoped `git diff --check` exited 0 with no output. Self-review caught and corrected a new
draft's seed-2 missile remainder to 70 before that reader pass. No code tests or gameplay
actions were run. After the triage edits, targeted lint again exited 0 with **0 errors,
0 warnings**, and scoped `git diff --check` exited 0 with no output. The same targeted lint
and scoped diff check were repeated after pass-2 triage and state-owner corrections, again
with exit 0 and 0 lint errors/warnings. Those checks describe the then-current contract;
full pass 3's actual results and triage are recorded above. The next handoff requires the
targeted lint on the pass-3 edits, which subsequently exited 0 with **0 errors, 0 warnings**.
After the review-5 fixture additions, targeted lint exited 0 with **0 errors, 0 warnings**.
No original-source campaign, Git command, code test or gameplay action is part of this
triage pass. Delta pass 6's actual results are recorded above.

### Specification review (2026-10-15)

An independent review recomputed all 36 vector families and found no numeric mismatch. It
re-read the original disassembly for D2Game 0xACE60 (owner/target/record guards, signed
stat test, default 64, large 32 before small 128, attacker seed, mask 127, OR 8), 0x75160
(mGH then mKB availability, mode-0 guard, path type 8, distance 10 for base 78 else 5) and
the missile chance slice of 0x3AE10; all agree with the current rules. KNB-R10a now calls
HRC-R15 `unresolved`, matching that rule's current label. The implementation-gap rows had
mechanical rename artefacts (`item_combat-knockback`, `missile_combat-knockback` and similar);
they were restored to the implementation's actual names `item_knockback`/`missile_knockback`
and to plain "knockback", without changing any finding. No vector, rule or gate changed.
