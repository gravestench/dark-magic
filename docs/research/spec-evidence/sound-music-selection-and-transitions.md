# Evidence: Music selection and transitions

Spec: [Music selection and transitions](../../specs/sound-music-selection-and-transitions.md)

## Identity and method

Build: original Lord of Destruction Expansion 1.13c Windows x86. Original artifacts resolve through the [source manifest](../source-manifest.md): LOD113C-D2CLIENT, LOD113C-D2COMMON and LOD113C-D2SOUND. The client image was freshly hashed on 2026-10-29 and equals the manifest SHA-256 dd8bc6025de921216a97c17f97cd1a50fbb85926e838ec60e13451448836d906. The relevant shared and sound functions were read under those profile identities; loaded runtime DLL hashes were not captured.

New reading used byte-matched reconstruction for the helpers identified below and disassembly for the unmatched controller. Original-image native disassembly corroborated the encoded instruction sequences in the audio dispatch. An independent Ghidra analysis of an immutable copy of the same hash confirmed controller instructions and its incoming caller. Its pseudocode was a lead only: the stack/register argument recovery conflates current tick with another live register, so the behavioral contract follows the original caller and instructions instead.

## Function and dispatch records

All descriptions are analyst descriptions, not original exported names.

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2CLIENT | 0x44b50 | none | game-loop audio dispatch caller | MUS-R02b through MUS-R06 | original-image call sites 0x44bcb and 0x44ea0 reach 0x5f020; complete game loop not verified |
| LOD113C-D2CLIENT | 0x5f020 | none | client audio update dispatch | MUS-R02b through MUS-R06 | disassembly plus original-image byte/instruction corroboration; calls ambient at 0x5f02f, music at 0x5f03b, loads at 0x5f047, voices at 0x5f052; increments clock at 0x5f065 |
| LOD113C-D2CLIENT | 0x45660 | none | music/environment request controller | MUS-R02b through MUS-R06 | complete instruction reading; Ghidra listing corroboration; only the specified supplied-input stages are verified |
| LOD113C-D2CLIENT | 0x5dfd0 | none | sound and environment table initialization | MUS-R01, MUS-Q01 | descriptor and consumer reading; parser/import closure incomplete |
| LOD113C-D2CLIENT | 0x87d20 | none | find first voice handle for sound id | MUS-Q01 | byte-matched reconstruction read; returns zero if no matching list entry |
| LOD113C-D2CLIENT | 0x87af0 | none | detect a non-stopping voice in fixed id interval | MUS-Q01 | byte-matched reconstruction read; ids 0x1231..0x125a and state not 2; identity-to-track binding not inferred |
| LOD113C-D2CLIENT | 0x87de0 | none | get position for voice handle | MUS-Q01 | byte-matched reconstruction read; forwards sound object to D2Sound ordinal 10032 through thunk 0xd30c; position units not established |
| LOD113C-D2CLIENT | 0x88080 | none | test matching voice state equals 1 | MUS-Q01 | byte-matched reconstruction read; no claim that entry evidence establishes audible playback |
| LOD113C-D2CLIENT | 0x45520 | none | reset music state and bookmark storage | MUS-Q01 | byte-matched reconstruction read; original label describes unrelated video state and is not used as proof |
| LOD113C-D2CLIENT | 0x45570, 0x454e0 | none | schedule and adjust interruption state | MUS-R06 | disassembly read; root producers and all cue semantics remain gated |
| LOD113C-D2CLIENT | 0x888b0, 0x88930 | none | mark music voices for stop/fade | MUS-R05b | disassembly and original-image corroboration read; nested voice helpers and backend integration not complete |
| LOD113C-D2CLIENT | 0x88a70, 0x88dd0 | none | enqueue sound and update voices | MUS-R05b | instruction paths read; complete lifecycle owned by playback contract and remains gated |
| LOD113C-D2CLIENT | 0x5da50 | none | checked sound-record lookup | MUS-R04 integration | disassembly read; signed id bounds and record stride 0x8e |
| LOD113C-D2CLIENT | 0x23e20, 0xa5440 | none | area-entry speech trigger and unit eligibility | MUS-R06 | instruction reading; quest records, cue producers and nested dispatch not complete |
| LOD113C-D2COMMON | 0x2fe10 | 10331 | obtain room through unit path | MUS-Q01 | disassembly read; original misleading label not semantic evidence |
| LOD113C-D2COMMON | 0x3c000 | 10826 | obtain level identity from room hierarchy | MUS-Q01 | byte-matched reconstruction read |
| LOD113C-D2COMMON | 0x6ccc0 | 10014 | checked level-indexed record lookup | MUS-Q01 | byte-matched reconstruction read; positive index below count; record stride 0x220; field binding is a separate check |
| LOD113C-D2SOUND | 0x9af0 | 10001 | return master volume setting | MUS-R03 | byte-matched reconstruction read; producer corroboration below |
| LOD113C-D2SOUND | 0x9ac0 | 10011 | return music volume setting | MUS-R03 | byte-matched reconstruction read; producer corroboration below |

Controller import thunks: D2Client 0xc1f0 -> D2Common #10331 RVA 0x2fe10; 0xc1de -> #10826 RVA 0x3c000; 0xc214 -> #10014 RVA 0x6ccc0. D2Client 0xd270 -> D2Sound #10001 RVA 0x9af0; 0xd26a -> #10011 RVA 0x9ac0. Import and export tables of the original images supplied these bindings. D2Client 0xd30c -> D2Sound #10032 RVA 0x6850. Its byte-matched reconstruction was subsequently read: it obtains a buffer cursor through a nested callback and shifts the unsigned result right by two; missing buffer or unsuccessful callback returns zero. The nested callback/API binding, stream format and units remain part of MUS-Q01 rather than a verified dependency. The sound-table parser calls Fog #10208, #10210 and #10207; complete parser verification is not claimed.

## Semantic mapping and verified stage boundary

The dispatch reads the unsigned audio tick from client RVA 0x11c268 and supplies it as the stack argument to the controller. The controller caches area identity at 0x11c34c, entry tick at 0x11c340, requested song at 0x11c348 and bookmark tick at 0x11c350. These locations explain the evidence, not required implementation layouts.

MUS-R02b: controller 0x456d2..0x45709 initializes the candidate, loads the environment's first signed dword if valid, and filters a nonzero candidate against the inclusive signed range. The loader descriptor at 0x5e3a6 names Song (string RVA 0xd557c), puts it at environment offset zero and uses field type 2. Environment rows have stride 0x58. The descriptor plus consumer establishes that this is the Song candidate; a name alone does not establish an audible track. The controller gets its environment index from byte offset 0x21c of the level-indexed record. The subsequent loader recheck below binds this byte to Levels.SoundEnv; full effective parser/runtime correspondence remains open.

MUS-R03: 0x45725..0x4576a compares current request and candidate. Elapsed uses a 32-bit subtraction and unsigned comparison with 75. The current request and bookmark tick are written at 0x4575e and 0x45764 only when the candidate differs and at least one change-permitting condition holds. This verified stage ends before stop/submit helpers. Supplied eligibility query results do not verify their producers or the audible effect.

MUS-R04: 0x458de..0x4595e first requires nonzero request, current tick strictly greater than bookmark tick plus 125, a found handle and active voice. After the supplied position query, 0x45917..0x45958 initializes a zero result, applies three sequential signed interval tests and writes the bookmark and current tick. Sound-record offsets 0x54,0x58,0x5c map to Block 1, Block 2, Block 3 through descriptors in 0x5dfd0. Their text/runtime unit correspondence and resume effect remain gated. The stage's exact comparisons are verified even if its inputs are supplied independently of that unresolved integration.

MUS-R02a: the isolated loop at 0x5e954..0x5e990 initializes lower bound from sound count and upper bound to zero, visits prepared environment records, and applies signed minimum/maximum updates only for positive Song values. This loop has no nested call. Original-image instruction corroboration supports the same comparisons and output writes. The separately allocated bookmark storage and full parser/error paths remain MUS-Q01; their unresolved behavior does not gate this prepared-input loop.

MUS-R05b unresolved paths include request submission at controller 0x45849..0x458aa, stop/fade request operations 0x888b0/0x88930 and voice update 0x88dd0. MUS-R06 includes timed interruption state in 0x45570/0x454e0 and controller 0x45788..0x45849. No copied design recommendation establishes crossfading, seamless looping or which quest selects which track.

## T01 Effective song cells

Source profile: fleet-reported vanilla LoD launch key lod-113c, family LoD, version 1.13c, no data overrides reported. Game executable profile hash matches the manifest Game.exe value. Runtime loaded DLL hashes and physical supplying archive were not captured; this limits equivalence claims. Date: 2026-10-29.

Storm-loader virtual path: data/global/excel/soundenviron.txt. Requested offset 0, length 60000. Success; returned 5741 bytes, total size 5741; complete coverage of the returned file. Base64 bytes were decoded and parsed as tab-separated values by exact header names, retaining complete rows. Effective-file SHA-256: d69131977215f8d3ac91326ffa708fad24eeddb61efae3965cf4b4d9d41de90c. Rows use explicit Handle and Index fields; named identities and explicit indices are cited, not inferred positional numbering.

| Handle | Explicit Index | Column | Raw cell | Interpretation | Supports |
|---|---|---|---|---|---|
| ESOUNDENVIRON_TOWN_1 | 1 | Song | 4673 | decimal integer 4673 | MUS-R01 named cell |
| ESOUNDENVIRON_WILDERNESS | 2 | Song | 4679 | decimal integer 4679 | MUS-R01 named cell |

These are text cells only. The record proves neither compiled table equivalence, selected asset filename, actual execution, stream loop flag nor a quest-state choice. The independent supplied-input vectors do not presume these cells are the live compiled values.

## Observation procedures

MUS-O01: use an original Expansion 1.13c single-player character in Normal difficulty, any class/level and no special equipment. Start in Rogue Encampment and walk across the Blood Moor boundary. Measure the level identity, audio-update tick, area-entry tick, environment Song field, accepted range and current request on the same member. Collect the candidate decision and both state writes, including cases just below and at the 75-tick threshold. Record actual eligibility query results. A heard track alone cannot verify the stored request or exact delay. The route must provide a valid new level context; an entry capture in a menu does not establish gameplay behavior.

MUS-O02: in the same original game, remain in a level with nonzero requested song and active audio voice. Read the actual three block values, backend position, bookmark and bookmark tick through at least one qualifying update after 125 ticks. Compare the calculation to MUS-R04. Separately establish the backend position unit and the resume seek input before asserting an audible loop or seeking behavior. Do not mutate position or call the music routine to manufacture a match.

## Runtime records and limitations

| Vector | Source profile | Date | Module + RVA | Records | Result |
|---|---|---|---|---:|---|
| MUS-O01, MUS-O02 | lod-113c vanilla reported profile | 2026-10-29 | D2Client + 0x45660 intended | 0 | no observation: member reached character-creation menu, then repeated state/input requests timed out; trace admission rejected before arming |

During the initial attempt, no trace was armed and no assertion was runtime-confirmed. Automatic review rejected invasive function tracing while that owned member was unresponsive; additional explicit permission was requested. No workaround instrumentation was attempted. Graceful close timed out; the exclusively owned container slot was then stopped successfully. The archive read preceded the timeouts and is retained only as table evidence.

## Search coverage and remaining evidence

The existing audio research was examined as a lead. Its 1.14d missile observations and proposed architecture are outside this evidence profile. Root/output/data anchors located music at 0x45660 from original-image calls rather than a music label search (which returned no matching label). Sound environment strings and loader descriptors anchored data bindings. Raw-image relative-call scan and Ghidra incoming-reference evidence corroborated 0x5f020. This is successful location discovery, not an exhausted six-rung absence claim. All open integration paths are listed in the [gap file](../../plans/gaps/sound-music-selection-and-transitions.md).

## Review record

Initial self-review: verified the supplied-input boundary, clock signedness, strict bookmark cadence, inclusive request range, independent vectors, data coverage and explicit integration gates. Initial mechanical lint: 0 errors, 15 warnings. The preserved-zero statement describes current behavior; reconstruction and independent disassembly method names are public verification methods, not private filenames or tool sessions. These warnings were reviewed and retained.

Fresh blind-implementer full pass 1 (2026-10-29): MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 3, COVERAGE 10 as reported, HARDEST 3. The reader recomputed all eleven vectors and independently outlined MUS-R02b through MUS-R04 from the contract alone. Review completion is limited to this supplied-input contract; it does not close MUS-Q01 through MUS-Q03.

Advisory coverage outcomes: no change for zero/below-bound/upper-equality candidates, because the signed inclusive range operation and missing-environment zero result are explicit and the filter vectors already distinguish inclusive acceptance, out-of-range rejection and zero-producing absence. No change for zero-current-request, query-B-zero and wrapped elapsed: these separate branches are specified explicitly; the vectors cover the two outcome classes of request permission and absence of an extra timestamp write. No change for false bookmark eligibility, wrapped threshold and first-block final selection: their control flow and wrapping are explicit, and cadence, block interval, zero fallback and sequential overwrite each have discriminating vectors. These remain coverage limits; the review does not assert exhaustive branch testing. The reader's reported COVERAGE count of 10 groups the three filter, three request and three bookmark entries (nine listed cases); this record retains the reported count and actual listed scope rather than inventing a tenth case.

HARDEST outcomes: no change for sequential bookmark overwrite (MUS-V11 explicitly demonstrates it), request permission OR conditions (MUS-V05/MUS-V06 and rule text supply them), or strict unsigned bookmark cadence (MUS-V08/MUS-V09 and explicit arithmetic supply it). No behavioral assertion, gate or vector changed in this triage, so no new full reader was required. No user approval is inferred.

Final mechanical validation on 2026-10-29: 0 error(s), 16 warning(s). Warnings were reviewed: temporal field names describe current retained state, and source-reading method names describe public verification methods rather than private reconstruction paths. No implementation or user approval is asserted.

## Authorized runtime follow-up — 2026-10-29

The maintainer explicitly authorized fleet function traces on an owned, responsive 1.13c member. A new exclusively allocated vanilla lod-113c member reached single-player gameplay with the disposable level-1 Expansion Barbarian fixture in Rogue Encampment. State reads were responsive before arming. Runtime loaded-image hashes and difficulty were not separately captured; the reported launch profile remains the identity limit. No save contents are published.

| Vector / purpose | Source profile | Date | Module + RVA | Records read | Result |
|---|---|---|---|---:|---|
| MUS-O01 controller identity | vanilla lod-113c reported profile | 2026-10-29 | D2Client + 0x45660, caller return +0x5f040 | 12 initial records, dropped 0 | matching dispatch identity and consecutive tick inputs only; initial remaining records were not retained before recorder-slot reuse |
| MUS-O01 stationary paired state | same newly owned member | 2026-10-29 | D2Client + 0x45660 | 220 complete records, dropped 0 | audio ticks 5132 through 5351; four-byte stack cleanup; paired state snapshots inside that interval are consistent with unchanged request behavior, not an atomic vector capture |
| MUS-O02 backend position query | same newly owned member | 2026-10-29 | D2Sound + 0x6850, caller return D2Client +0x45917 | 7 complete records, dropped 0 | controller reached position query; all captured object windows returned the requested 80 bytes; no global bookmark-write capture |
| eligibility query B | same newly owned member | 2026-10-29 | D2Sound + 0x9ac0 | 10 complete records, dropped 0 | four records have caller return D2Client +0x45718; all ten return 50; this does not establish a Boolean-only return or global's wider meaning |

The complete controller window contains each tick from 5132 to 5351 once. Within it, paired live-state reads at audio tick 5187 and 5230 both had requested song 4673, cached area identity 1 and bookmark tick 5166. Thus the request and bookmark tick remained the same across the two reads, consistent with MUS-R03's unchanged-request case. These reads do not capture the globals on each function entry and return, so they do not establish every invocation's before/after state or the exact inputs of MUS-V07.

A checked live row for requested song 4673 had Block 1 = 3002253, Block 2 = -1, Block 3 = -1. The per-song stored bookmark read was 3002253. The seven backend-return positions were 2221458, 2332585, 2443746, 2554906, 2666012, 2777136 and 2888292; every one is nonnegative and below 3002253, so MUS-R04's supplied-input calculation would select 3002253. The matching stored value is corroboration only: the bookmark reads and position calls were not paired atomically with the global write and bookmark tick, so MUS-V09 and MUS-O02 are not marked matched. Backend position units, resume behavior and audible transitions remain unverified.

A trace of the client position-wrapper entry at 0x87de0 was refused for an unrecognized prologue. No force option was used. The backend entry at 0x6850 accepted normal tracing. An initial eligibility-A capture was replaced by later slot reuse before its records were read; no observation is claimed for it. The observation collector's windows accept register or stack-derived pointers and do not capture arbitrary controller globals on entry/return; this is the current completeness limitation.

No behavioral rule label or vector runtime flag is promoted. All captures were capped, the retained records' loss counters were checked, tracing was stopped or had already stopped at its cap, and the exclusively owned member accepted graceful quit. The earlier approval blocker is resolved; MUS-Q03 remains an observation-completeness gap. A broader listing of shared-host character saves was rejected by automatic review for unrelated private-save exposure; no workaround was attempted. Only the owned member's disposable fixture was inspected.

Follow-up review disposition: this adds limited observation evidence and resolves the permission blocker without changing behavior, input contracts, vectors, expected results, labels or gate boundaries. The existing fresh-reader result remains applicable.


## Static producer recheck — 2026-10-29

This recheck closes previously neutral producer identities without claiming full stream integration. D2Common level loader RVA 0x6e130 has original descriptors for Rain at record byte 0x5, Mud at 0x6 and SoundEnv at 0x21c. The SoundEnv descriptor is at 0x6f81c..0x6f839 and names the original string at RVA 0x8fac8. Its loader call at 0x6f8e3 publishes the level table pointer and count at 0x6f8e8. The original initialized global pointer and the getter at 0x6ccc0 link that table to the controller lookup; this is not inferred from the getter label. Descriptor identity/direct conversion reuse GDA-R03a/GDA-R03b; complete generic compilation remains GDA-R02/GDA-R03d. Applicable unmatched instructions and an independent original-image listing were read.

D2Sound RVA 0x9ad0 sets the global returned by 0x9af0 after writing the registry value named Master Volume; 0x9aa0 sets the global returned by 0x9ac0 after writing Music Volume. RVA 0x9b30 initializes these values to 100 and 50 respectively and retains those defaults unless a successful read supplies an unsigned value at most 100. The original strings and instruction listing corroborate these identities. Thus controller queries A/B are stored volume integers, not abstract backend-ready booleans. MUS-R03 still tests only zero versus nonzero; these producer identities do not establish downstream device behavior.

The current supplied prepared-record initialization is MUS-R02a with MUS-V12/MUS-V13. The earlier candidate-filter identifier is now MUS-R02b throughout this unconsumed contract. Prior eleven-vector review remains historical; a fresh review is required for the resulting thirteen-vector contract.


## Fresh blind-implementer full pass 2 — 2026-10-29

MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 3, COVERAGE 10, HARDEST 3. All thirteen vectors were independently recomputed, including the positive-only prepared-record loop and no-positive diagnostic result. All named discriminators rejected their stated alternatives.

COVERAGE outcomes: no change for available zero/below-bound/upper-equality candidates; zero current request/query-B-zero/wrapped elapsed; false bookmark eligibility/wrapped cadence/first-block final selection; or empty ordered set distinct from nonpositive values. These branches are explicit, with existing vectors distinguishing the corresponding acceptance, request-write, cadence, sequential comparison and no-repair outcomes. They remain advisory coverage limits, not claims of exhaustive testing. HARDEST outcomes: no change for the positive-only min/max loop, request permission OR and paired write, or strict wrapped cadence and sequential bookmark tests; each is explicit and recomputed successfully. GATED outcomes: retain MUS-Q01 through MUS-Q03. This review applies to the current thirteen-vector contract; it does not establish full stream behavior or approval.


## Interruption producer expansion — 2026-10-29

Original-image direct-call scanning found interruption scheduler 0x45570 called from event routing at 0x23200 and client scripted-effect entry regions 0x53070, 0x530d0, 0x53220 and the additional entry near 0x53450 within the 0x533f0 listing. These applicable unmatched instruction listings were read after checking reconstruction availability. The 0x23200 path selects class-specific cues and has explicit event-code special cases; no quest meaning is inferred from those numeric event codes. Scripted entries have local-player/room/object-lookup conditions or run before a missile-creation request. This materially narrows the missing root search, but object resolution, event-code producers and later interruption playback/resumption still require closure.

The scheduler has six stack inputs and an additional register selector used for a selected block bookmark. It attempts a current-song bookmark before requesting music voices stop, then writes the timed interruption state. Its deadline base is a separate client clock global (RVA 0x1197f8), not the audio-update counter supplied to the area controller. Exact clock correspondence must be established before treating interruption delays as audio ticks. A direct-call scan found no direct caller of adjustment helper 0x454e0; indirect/table references were not exhausted, so no absence claim is made. These are additional attempted paths under MUS-Q02, not verified interruption rules or new audible observations.


## Ordinary submission caller closure — 2026-10-29

MUS-R05a is the complete ordinary submission branch at D2Client 0x45849..0x458aa after interruption control permits entry. Current requested song must be nonzero and both settings nonzero. A first handle lookup through matched 0x87d20 suppresses submission whenever any handle exists. Otherwise the caller reads the per-song bookmark, queries matched 0x87af0 (non-stopping voices in the fixed original music interval), selects request flag 2 only for no such voice plus zero bookmark, calls 0x888b0 to request all music voices stop, then 0x88a70 with cue/absent emitter/delay zero/selected flags/bookmark playback value. It ignores the returned handle and writes the entering audio tick at RVA 0x11c344. Full instructions and original-image listing were rechecked. No RNG call occurs locally.

The verified boundary logs these operations with supplied returning results and no other coordinator-state changes; actual voice lookup population, stopping, admission, bookmark seek units and eventual audible start remain STP-R03b/STP-R04b and MUS-Q01. Interruption-to-ordinary entry remains MUS-Q02. Thus the former unresolved lifecycle rule is now MUS-R05b, while MUS-R05a defines caller orchestration with MUS-V14 through MUS-V18. A fresh full reader is required after this behavior addition.


## Fresh blind-implementer full pass 3 — 2026-10-29

MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 3, COVERAGE 11 advisory cases, HARDEST 4 rule groups. The reader independently recomputed all eighteen vectors and V18(a)/(b), twenty input cases total. All outputs and named discriminators matched.

COVERAGE outcomes: retain the ten previously recorded advisory partitions with their pass-2 no-change rationale. No change for the additional combined music-range-present/nonzero-bookmark case: V15 and V16 independently distinguish each zero-flag condition, and the rule explicitly uses flags 2 only when both fresh-start conditions hold. HARDEST outcomes: no change for range initialization, request permission/paired write, strict wrapped bookmark cadence/sequential tests or ordinary submission query/flag/stop/playback/attempt-tick order. Each is explicit and independently recomputed. GATED outcomes: retain Q01 through Q03. The current eighteen-vector supplied-stage contract meets the reader stopping rule; no complete audible or runtime confirmation or user approval is inferred.
