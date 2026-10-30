# Screen-fade evidence

Spec: [Screen fades](../../specs/draw-screen-fades.md)

## Identity and reading, 2026-10-29

The original client SHA-256 was freshly read and matches `LOD113C-D2CLIENT` in the [source manifest](../source-manifest.md). All descriptions below are analyst descriptions, not original exported names.

| Artifact | RVA / export | Supported rules | Method and boundary |
|---|---|---|---|
| LOD113C-D2CLIENT | 0x10E90 | SFD-R01 cancel | Reconstruction checked against original bytes: single timer clear |
| LOD113C-D2CLIENT | 0x10EA0 | SFD-R01 start | Reconstruction checked against original bytes: GetTickCount plus 500, low32 store |
| LOD113C-D2CLIENT | 0x10EC0 | SFD-R02, SFD-R03 | Complete instruction reading through graphics call; no reconstruction checked against original bytes available |
| LOD113C-D2CLIENT | 0xD1CE import thunk | SFD-R03 | Resolves D2Gfx ordinal 10045, RVA 0xB9F0 |
| D2Gfx.dll 1.13c | 0xB9F0 / ordinal 10045 | SFD-R03 | Reconstruction checked against original bytes: signed rectangle guards and interface forwarding; backend behavior unresolved |

Client IAT slot 0xCEF5C (preferred VA 0x6FB7EF5C) resolves KERNEL32.GetTickCount by name. Its external time result is the semantic clock input. Original bytes at client RVA 0x10EC9 are `FF 15 5C EF B7 6F`: opcode FF /2, absolute indirect call through that slot. This corroborates the disassembly's byte-emitted instruction rather than trusting its annotation. Other instructions in the bounded computing function are ordinary decoded instructions. The independently fingerprinted original D2Gfx.dll is 77824 bytes, SHA-256 `50284647a28e8839028e8acb021a6ebe48c006e431139969ccd0fe126e6e8811`, measured 2026-10-29. This companion records that additional identity without editing the shared manifest; no backend-wide claim is made.

## Dispatch and output anchors

Client draw path 0x44230 calls 0x8B110 when its view selector is not 3. The tail of 0x8B110 calls 0xC0EA, 0x805A0, then 0x10EC0. Only that reached tail is established; the complete upstream frame path and its eligibility have not been read. Unit-processing function 0xA7E20 calls 0x10EA0 at 0xA7F1F after several unit tests; those tests and imported helpers are not assigned gameplay meanings here. Cancel's callers were not fully resolved. These limits are SFD-Q01, not claims of verified complete dispatch.

Output anchor: client retained timer is preferred VA 0x6FBCC554 (module-relative 0x11C554). Width/height are globals 0x6FB8BC48/0x6FB8BC4C; view selector is 0x6FBCC414. The draw function compares current time and deadline unsigned, then computes low32 remaining-time times 255 and uses signed division by 500 (multiply-high reciprocal 0x10624DD3, arithmetic shift 5, sign correction). Half width uses sign correction before arithmetic shift, implementing truncation toward zero. Selector 1 subtracts the half from full width; selector 2 uses the half for the left bound. Graphics arguments are left, zero top, right, height minus 47, zero color selector and strength. D2Gfx ordinal 10045 rejects right below left or bottom below top, then forwards to interface slot 0xBC. The backend call is outside this supplied-request contract (SFD-Q02).

## Lead separation

Candidate 0x524D0 was read as a lead. It writes RGB 255/64/48 when its guards permit and has unit-color callers 0x52EB0 and 0xA9A20. It does not compute the timer rectangle. Its non-sentinel strength path calls floating arithmetic helpers and was not verified here. No rule in this spec claims its behavior.

## Observation procedure

Use an owned unmodified Expansion 1.13c game, any class and difficulty, with the selected client character present. First identify the gameplay action that reaches client 0x10EA0 under SFD-Q01 rather than guessing from labels. Capture current time and timer immediately on start; capture entries of 0x10EC0 and outgoing graphics requests at 0xD1CE, with corresponding view dimensions and selector, across the following 500 milliseconds. Check all six arguments, timer clearing and the unsigned expiration condition. Repeat with each half-view arrangement. No prescribed equipment, level or location has been established for this trigger; the procedure is gated until eligibility is known.

## Runtime attempt, 2026-10-29

An initial owned vanilla Expansion 1.13c attempt reached the title screen but established no entry observation at client 0x10EC0. It provides no gameplay-vector match or evidence of absence and does not satisfy gameplay-trigger eligibility. All vectors remain runtime_observed no. No table data was read or claimed.

## Review status

Self-review preserves supplied-stage versus gameplay/pixel boundaries. Fresh full pass 3 below completes the blind-implementer gate for the fifteen-vector contract. Document status remains draft for the stated integration and observation limits; approval is not inferred.

### Final full reader pass 3, 2026-10-29

Fresh independent full review: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 3, COVERAGE 1, HARDEST 3. All fifteen vectors recomputed correctly. The existing three questions retain their boundaries. Advisory negative odd widths with half selectors require no change: both half variants still fail the signed rectangle check for negative widths, the no-request result already discriminated by V12; positive odd widths V05/V06 distinguish half selection. No new valid-rectangle output category is uncovered. HARDEST R03 arithmetic/inclusive bounds, R02 unsigned time and V15 signed wrapped product need no change: the rules give each operation and the vectors distinguish signedness, truncation, overflow and equality. No behavior, input, vector, output, ownership, gate or acceptance assertion changed after this final full pass. The historical full passes and corrections follow.

## Blind reader pass 1, 2026-10-29

Fresh reader counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 4, COVERAGE 7, HARDEST 3. All seven initial vectors recomputed correctly. GATED findings preserve existing upstream dispatch, backend pixels, runtime capture, and observation-trigger limits; no behavior was inferred for them.

COVERAGE triage: start wrap adds SFD-V08 (zero) and SFD-V09 (nonzero). Inactive zero adds SFD-V10 and names time-sampling requests in Outputs. Unsigned high-bit comparison adds SFD-V11. Invalid horizontal bound adds SFD-V12. Out-of-range strength adds SFD-V13. Other-selector branch needs no change: selector 0 in SFD-V04 already exercises the default full-width branch. Invalid-rectangle deadline retention needs no change: SFD-V07 already asserts unchanged deadline with no request. All seven advisory entries therefore have a vector or an explicit no-change rationale.

HARDEST triage: R03 mixed signed/low32 arithmetic and bounds needs no rule change because section 3 states widths and truncation, R03 states sequencing, and new vectors expose bounds/range edges. R02 absolute unsigned expiration needs no rule change because it explicitly rejects elapsed-duration comparison, with SFD-V11 added to demonstrate it. V05 odd-width half plus fractional strength needs no change because separate right=401 and strength=127 assertions distinguish both choices. New vectors/acceptance and time-sampling output require a fresh full reader; that pass remains pending.

## Blind reader pass 2, 2026-10-29

Fresh full reader of the thirteen-vector contract: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 3, COVERAGE 2, HARDEST 3. Numerical outputs matched. DISCRIM V02 noted that an unnecessary clock read on cancel would pass; the resulting contract explicitly declares the supplied clock's sampling log, states that cancel does not sample, and asserts an empty sampling log in V02. Existing V10's inactive no-sampling assertion is now backed by the same declared observation boundary. Original cancel `0x10E90` is a timer clear without a clock call; this clarification adds no guessed time behavior.

COVERAGE inclusive bounds adds V14, a zero-area valid rectangle. COVERAGE wrapped arithmetic adds V15, whose unsigned remaining-time product has signed low32 value -255 and hence strength zero. R03's inclusive guard and low32 multiplication are unchanged. HARDEST R03 arithmetic/rectangle composition, R02 absolute unsigned comparison and V13 unclamped strength require no further behavior change: their widths, order and discriminating outputs are explicit, now also tested by V14/V15. All findings are resolved by edits or these no-change reasons. The new assertions require another fresh full pass; this record alone does not complete that pass.
## Upstream relocation leads, 2026-10-29

A new read follows the timer start backward without upgrading SFD-Q01. Original client command-table row 0x15 at RVA 0xDDF5C points to 0xAF640 and specifies length 11 (table origin calibrated against the world-clock companion's command 0x53 at 0xDE244). Handler 0xAF640 passes the packet's final byte as the third argument of 0xA7E20. That argument must be nonzero to reach timer start. The two other direct callers examined, 0x1B540 and 0xC7290, pass zero. At 0xA7E20, additional tests require the selected unit to be the current player, its old room to exist, player mode to differ from 0 and 17, and unit flag 0x10000 to be clear. A current-skill record test can suppress the argument. Those mode/flag/skill meanings and complete outer roots remain unresolved, so these are investigative conditions rather than new behavioral rules.

Client imports resolve to D2Common ordinal 10331 at 0x2FE10 (room lookup), ordinal 10511 at 0x303A0 (current-skill pointer retrieval), and ordinal 10858 at 0x4E1C0 (record-pointer dereference). The first was read in instructions, the other two in reconstructions checked against original bytes. Original server 0xCBB10 constructs the matching length-11 command and receives its final byte from its second stack argument. At 0x788B0, unit flags 0x10000 select final byte 1; another position-update branch sends zero. Candidate writers of the former flag include 0x6BD60 after successful relocation. Complete writer/dispatch/callee verification has not been completed; portal, waypoint and level-door actions are leads, not established triggers. No unsupported label was used to close SFD-Q01.

## Bounded runtime attempt, 2026-10-29

A newly owned vanilla Expansion 1.13c member loaded a normally created level-1 Barbarian on Normal difficulty with starter gear. Direct reads corroborated zero timer state, dimensions 800 by 600, view selector 0 and the imported clock function's implementation. Client draw observation captured 128 records with zero dropped records, but these records lacked simultaneous input windows. The packet handler and graphics forwarding recorder each captured zero calls within their capped windows. Walking around the town did not reach the intended natural relocation; consequently zero packet records do not test relocation eligibility or establish absence of the behavior. No vector was promoted.

A further attempted capture did not yield complete simultaneous input windows. Partial read-only values and ordinary draw returns do not prove a complete vector. The incomplete records remain preserved separately; no runtime label or acceptance assertion changes follow from this attempt.
An earlier capture is reused with its original limits: eight corrected draw records contained complete timer snapshots of zero before and after, return zero and zero dropped records. Graphics recorded zero requests while armed across the draw range. Initial register-derived windows were invalid after a scratch register changed, and an independent clock-call capture did not have established overlap. Thus the corrected subset corroborates inactive state only: absence of a time sample in SFD-V10 remains unobserved, and neither that vector nor SFD-R02 is promoted. The raw initial and corrected records are preserved separately without overwriting them. The subsequent bounded campaign adds no complete timer, clock or active-rectangle vector.

## Four-snapshot follow-up, 2026-10-29

The follow-up captured four read-only snapshots: deadline, current clock value, view dimensions and view selector. On a newly owned vanilla Expansion 1.13c member, the original instruction operands and resolved field addresses were checked before capture. An initial 16-record window and a later 512-record window each had zero dropped records and all four snapshots complete. Every complete record showed deadline zero before and after, dimensions 800 by 600, selector 0 and return zero. The clock snapshot was a separate read of the clock source; it does not prove that the fade callback requested a clock sample.

Normal play reached Blood Moor from town and returned through a normally created town portal. The two 512-record draw windows collected around the outward crossing had zero drops but no complete input snapshots after the incoming scratch register changed; they are discarded for vector matching. The later complete 512-record window followed the portal return and remained inactive throughout. No observed command-handler or graphics request establishes an eligible active fade, and the windows do not cover every instant of either transition. SFD-Q01 remains open.

A separate bounded attempt to bracket inactive draw records with clock-call timestamps remained incomplete. The clock recorder collected 512 zero-drop records, but the corresponding draw snapshots were missing after another register change. The corresponding draw window lacked valid snapshots when the register changed. No missing-window record or uncorrelated zero-call stream supports an absence claim. SFD-V10's no-clock-sampling assertion remains unobserved; all vectors retain runtime_observed no and all three rules retain their static labels.

Protected raw records are preserved privately with SHA-256 `f1d8ba6f7cc1af66356710f8c1a3296b0b2aef793ac712104b50debe0decd34e`. This evidence-only follow-up changes no behavioral contract, vector, acceptance assertion or gate; the full reader pass 3 therefore remains applicable.
