# Evidence: Tainted Sun quest

Spec: [Tainted Sun quest](../../specs/quest-tainted-sun.md)
Open questions: [evidence gaps](../../plans/gaps/quest-tainted-sun.md#open-evidence-gaps)
Related evidence: [world clock companion](drlg-world-clock.md)

## Identity, methods and reuse

The evidence build is original Lord of Destruction Expansion 1.13c, Windows x86. Artifact ids below resolve through the [source manifest](../source-manifest.md): `LOD113C-D2COMMON`, `LOD113C-D2GAME`, `LOD113C-D2CLIENT`, and `LOD113C-FOG`. Addresses below are module-relative RVAs unless explicitly identified as preferred VAs. Descriptions are analyst labels, not original exported names. Original images and game assets are not distributed.

Static readings recorded on 2026-10-12 are reused for the quest chain, direct RNG audit, tables and server frame chain, with their original limits. The source draft's earliest reading had no recorded date or source identity; it remains an undated lead where the dated record did not cover it. Historical loaded DLL hashes were not captured, as the manifest explains. The available original D2Common/D2Game specimens match the published fingerprints; D2Client and Fog current specimen checks establish only those specimens, without retroactively fingerprinting a runtime session.

Targeted analysis on 2026-10-15 checked reconstructions verified against the original bytes first where available, otherwise disassembly, with original-image decoding for ambiguous byte runs and previously uncorroborated instruction slices. No reconstruction or original binary was changed. All source artifacts, full original document bytes, internal paths/revisions, commands, fixture encodings and tool/session history were separately preserved in a private integrity-checked audit. That audit is not required to interpret the public contract. The full original main and gap record predate this companion; historical evidence was not discarded.

The migration retains the quest-plus-clock boundary and all twelve rule families, nineteen existing vector families and four observation families. It adds discriminators for supplied flag updates, partial failure, overflow, client source mapping, intensity branches and assignment. General lighting was not added. No gameplay collection, implementation comparison or implementation work was performed. The earlier question about 1.14d remains untested and outside this evidence build; no cross-build claim is made. Repository references outside sibling specs/evidence/gaps did not consume this contract, so there is no historical-alias column.

## Restructure (2026-10-15)

On 2026-10-15 the combined Tainted Sun world-clock contract was split along its quest/clock seam, applying the approved review recommendation to split oversized specifications. The quest stages, altar, start/reset/failed-join messages, diagnostic text, retry, quest persistence and quest RNG stayed in [Tainted Sun quest](../../specs/quest-tainted-sun.md) (prefix TSW). The clock tables, allocation/start/reset assignments, advance/wrap/transition, client wrapper, intensity branches, server frame/resynchronization, join snapshot, client assignment and other consumers moved to [world clock](../../specs/drlg-world-clock.md) (prefix WCK). Neither contract was consumed, so ids were renumbered in each file without aliases. The integration question and its gap row were split between the two contracts (TSW-Q06 and WCK-Q05); the message-observation question kept the quest's own messages (TSW-Q03), and observation of the client's World-clock handling joined WCK-Q03. Vectors spanning both halves stayed with the contract owning the asserted behavior: the quest-stage vectors TSW-V01 through TSW-V05 also exercise WCK-R01b and WCK-R01c, and WCK-V10 runs the quest join stage TSW-R01d as a supplied stage. The historical implementation-gap rows were divided between the two gap files by subject, keeping their numbers and wording.

No rule behavior, label, value or vector result changed, and no new source or gameplay evidence was collected. The dated records in this companion are unchanged except that their id references use the current ids; records covering both halves appear in both companions. No reader run has yet checked the two resulting contracts.

## Original function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0xC7480` | — | quest initialization | TSW-R01, TSW-R07a | Dated initialization evidence reused; C view checked against original bytes read 2026-10-15 for installed callbacks, state 0, slot 11 and private state zeroing |
| LOD113C-D2GAME | `0xC6A30` | — | shared start | TSW-R01a, TSW-R07a | 2026-10-12 reading reused; original decode 2026-10-15 checks earlier writes, Act II test, call and return; callback interiors supplied |
| LOD113C-D2GAME | `0xC72C0`, `0xC7130` | — | join selector and apply | TSW-R01d, TSW-R04b | Disassembly 2026-10-12 reused; branch/state-order slices read 2026-10-15 |
| LOD113C-D2GAME | `0xC71A0`, `0xC73A0` | — | animation/event start slices | TSW-R01e, TSW-R01f, TSW-R01i, TSW-R04c | Targeted disassembly 2026-10-15; complete bodies re-read in the follow-up below; mode-update, roster-update and broadcast interiors excluded |
| LOD113C-D2GAME | `0xC7250`, `0xC6AF0` | — | schedule and timer callback | TSW-R01g, TSW-R01h | Dated draw/timer reading reused; targeted guard/return read 2026-10-15 |
| LOD113C-D2GAME | `0x1780`, `0x33AF0` | — | bounded random range and timer insertion | TSW-R01g, TSW-R07b | Disassembly read 2026-10-15; seed step/reduction; due-counter insertion only, not timer delivery |
| LOD113C-D2GAME | `0xC3750` | — | progress-bit writer | TSW-R01b | C view checked against original bytes read 2026-10-15, quest search, guards, exact bit choices, return 0 |
| LOD113C-D2GAME | `0xC6B30` | — | auxiliary quest callback | TSW-R01c | 2026-10-12 partial reading reused; 2026-10-15 instruction view confirms comparisons, not semantic selector identity |
| LOD113C-D2GAME | `0xC76B0` | — | altar operation | TSW-R02a, TSW-R02b, TSW-R02c, TSW-R07b | 2026-10-12 instruction-listing reading reused; original image decoded through `0xC7B3D` on 2026-10-15; callback/drop interiors excluded |
| LOD113C-D2GAME | `0xC5E90` | — | quest environment reset | TSW-R02b, WCK-R01c, TSW-R07a | Disassembly 2026-10-12 reused; 2026-10-15 verifies missing-act check before applied clear |
| LOD113C-D2GAME | `0xC5190`, `0xC5100` | — | start/reset per-player requests | TSW-R03, TSW-R04a, TSW-R07a | Dated reading reused; byte-run comparison at start sender checked in original image 2026-10-15 |
| LOD113C-D2GAME | `0xC5E30` | — | altar timer callback | TSW-R02d | Complete instruction read 2026-10-15 (follow-up); pre-start iteration interior supplied |
| LOD113C-D2GAME | `0xCBE70` | — | diagnostic text broadcast | TSW-R04c | Complete disassembly read 2026-10-15 (follow-up); message buffer built, recipient iteration `0x29820` with callback `0xCB580` supplied |
| LOD113C-D2GAME | `0x8A3E0` | — | per-client send queue | TSW-R04b | Entry instructions read 2026-10-15 (follow-up): absent client returns before any write; queue internals not verified |
| LOD113C-D2GAME | `0xC2150` | — | direct pending retry | TSW-R05a, TSW-R07a | Disassembly 2026-10-15; act argument/pending exact tests and no packet request; outer dispatch excluded |
| LOD113C-D2COMMON | `0x2D0F0`, `0x2D0B0`, `0x2D070` | 10174, 10003, 10593 | quest bit test/set/clear | TSW-R01b, TSW-R01d, TSW-R02a, TSW-R02b | C view checked against original bytes read 2026-10-15, resolved Fog operations below; gameplay names remain neutral |
| LOD113C-FOG | `0x14070`, `0x14090`, `0x14040` | 10119, 10118, 10120 | bit test/set/clear | TSW-R01b, TSW-R01d, TSW-R02a, TSW-R02b | Matched test/set reconstructions and clear disassembly, plus exact mask-data read 2026-10-15 |
| LOD113C-D2COMMON | `0x2EB70` | 10920 | valid-player data getter | TSW-R03, TSW-R04 | C view checked against original bytes read 2026-10-15; valid player type returns unit data; caller resolves client afterward |

## Dispatch, imports and layouts

### Quest roots and supplied dependencies

The original quest initializer pointer is stored at D2Game data RVA `0x10DCF8`; initialization at `0xC7480` installs join `0xC72C0`, event `0xC73A0`, and auxiliary callback `0xC6B30`. Earlier quest-row evidence also identifies the quest table region at RVA `0x10DBF0`; the initializer-pointer cell, rather than a guessed row number, is the new exact anchor. The join initializer-callback event was recorded as 0xD on 2026-10-12. The animation start slice `0xC71A0` is retained with its dated chain and a supplied outer callback boundary. Its complete upstream animation dispatch is not newly established.

Shared start `0xC6A30` requests player iteration through `0xDFCC0` with pre-start callback `0xC52B0`, then tests the Act II pointer, calls the start import and iterates start sender `0xC5190`. Timer callback `0xC6AF0` and retry `0xC73A0` additionally iterate progress writer `0xC3750`. The pre-start callback's own quest notification calls, iteration implementation, optional diagnostic notification and unrelated callback interiors are supplied integration operations, not omitted verified behavior. The animation slice's final mode setter is D2Common #11090, RVA `0x33920`, reached through D2Game thunk `0xA4A2`; its full effects are TSW-Q06.

Object-operation pointer at D2Game RVA `0x107C18` selects altar `0xC76B0`. It finds quest 10, tests progress at `0xC7713..0xC7779`, tests object mode at `0xC7779..0xC778B`, and on the active path calls quest reset `0xC5E90` at `0xC792E`. Reset tests Act II before clearing applied. Altar flag operations are at `0xC7992..0xC79B9`. The room-item search target `0xE660`, preparation callbacks, item drop `0x11A70`, gold drop `0x6AFA0`, and later quest/iteration operations are supplied boundaries; their names or the presence of the item code do not establish complete item search/drop semantics. Inline seed sites at `0xC7855` and `0xC7AB5` are traced independently of those helpers. Rejection is requested through `0xDFFD0` at `0xC7B2F`. All normal returns decoded through `0xC7B3D` return 0.

| Caller thunk RVA | Imported original target | Rules / purpose |
|---|---|---|
| D2Game `0xAB98` | D2Common #10337, `0x37CA0` | start |
| D2Game `0xAB92` | D2Common #10895, `0x37C20` | reset |
| D2Game `0xA4C0` | D2Common #10920, `0x2EB70` | valid player data |
| D2Game `0xA568` | D2Common #10174, `0x2D0F0` | quest bit test |
| D2Game `0xA57A` | D2Common #10003, `0x2D0B0` | quest bit set |
| D2Game `0xA586` | D2Common #10593, `0x2D070` | quest bit clear |
| D2Common `0x92A6` | Fog #10119, `0x14070` | bit read |
| D2Common `0x92B2` | Fog #10118, `0x14090` | bit set |
| D2Common `0x92AC` | Fog #10120, `0x14040` | bit clear |

The bit wrappers form bit index `(slot<<4)+bit`; test normalizes the Fog mask result to 0/1. Fog set masks at RVA `0x26798` are low-byte 1,2,4,8,16,32,64,128; clear masks at `0x26818` are their low-byte complements. Set uses OR, clear uses AND. This establishes actual bit behavior, without assigning gameplay meanings to the slots. Other helpers remain integration gaps.

### Quest records

For the original layout, Game holds Act pointers at `+0xBC+4*a`; Act II is `+0xC0`. The environment reached through Act `+4`, its fields and the clock tables are recorded in the [world clock companion](drlg-world-clock.md#clock-records-and-original-constants). These layouts locate evidence; they are not required storage designs.

Quest state/stage/transition are bytes `+0x0C/+0x0B/+0x14`; active `+9`; private state pointer `+0x18`, with scheduled/applied/pending at private `+1/+2/+3`. Game quest structure is `+0x10F4`, its seed `+0x18/+0x1C`; game seed `+0xD0/+0xD4`, unit seed `+0x20/+0x24`. A player's client is obtained through player-data `+0x9C`; act byte is client `+0x1AC`. Join's shared flag target is loaded from quest-structure `+0x0C`, after a separate `+8` non-null check. The neutral shared-flag name avoids the earlier draft's ambiguous pointer expression.

### Original message encodings

The original senders construct World-clock opcode `0x53`, ten bytes: opcode byte, little-endian cycle dword, little-endian ticks dword, flag byte. Start payload is `{5,0,1}`; reset `{2,0,0}`. Quest-event opcode `0x5D` is six bytes: opcode, quest byte `0x0A`, flags byte `0x10`, three zero bytes. Failed join uses `5D 0A 01 00 00 00`. The flag-1 start request precedes Quest-event for each eligible player. These are original wire evidence, not a new wire-format acceptance requirement in the main contract.

The original client row for `0x53` at `0xDE244` has length 10. Adjacent selected command rows `0x5D`/`0x5E` at `0xDE2BC`/`0xDE2C8` have lengths 6/38. This preserves the earlier size calibration and independently checks the current image's selected cells. The earlier reading associated the quest event with two client sounds; the named client leads are `D2Client+0x1E160` and `0x43830`. Their complete dispatch/audio effects remain TSW-Q06 and are not established by those labels or a table length. No packet capture was performed.

## Static corrections and affected-claim rechecks

### Final start ticks and client retained rate

The earlier draft asserted final start ticks 0 in its clock rule and start/join/reload vectors. It followed the initial store but omitted the downstream recomputation. Original-image checks on 2026-10-15 establish: D2Common `0x37CE5` initially stores 0; `0x37D07` sets flag 1; call `0x37D0E` targets `0x37170`. That complete no-callee helper selects eclipse row 0 at `0x37197`, multiplies rate 4 by begin 300 at `0x371AB`, and writes ticks 1200 at `0x371AE`. The subsequent intensity/colour routines' writes do not change these clock fields. The final server state is therefore ticks 1200, not 0. This is a source-backed correction, with no new runtime claim.

Rechecked assertions are WCK-R01b, shared start/join/direct-retry consumers, TSW-V01/TSW-V03/TSW-V04/TSW-V06/WCK-V10/TSW-V09, and the observation expectations. The independent server start sender still explicitly requests cycle 5/ticks 0/flag 1, so TSW-R04a's payload is retained. WCK-V07 retains the legitimate supplied-zero tick fixture and separately exercises returned start ticks 1200; both transition to row 1/ticks 0 on the next core call. Other threshold/rate/transition vectors are unchanged in arithmetic.

The client setter `0x37920` writes neither rate nor rate index. At `0x379EB..0x379F4` it writes the new flag and, if nonzero, calls the same recomputation helper, then intensity/colour. With retained client rate 128 and received cycle 5, final ticks are 128*240=30720; packet ticks 0 are superseded. This closes the previously unverified supplied-client-rate question and supplies WCK-V16. The stage with normal flag accepts ticks equal to rate*360 and pre-clamps only strictly greater values (`0x37996..0x3799A`). Full packet delivery/appearance is still gated.

### Altar clears bit 14

The earlier altar rule/vector said bit 14 was set with bits 1/13. Original altar call `0xC79B9` targets D2Game thunk `0xA586`, resolved to D2Common #10593 `0x2D070`, then Fog #10120 `0x14040`. Matched wrapper reading, clear routine AND operation and original mask values on 2026-10-15 prove **clear**, whereas prior calls at `0xC79A5/0xC79AF` use the set operation. Rechecked TSW-R02b, TSW-V05 and outputs now require set 1,set 13,clear 14. The vector begins with bit 14 set so a false set/no-op cannot pass. Flag meanings remain neutral under TSW-Q02.

### Boundaries, historical labels and locator corrections

Mixed table/runtime/code labels were split into one label per leaf. Idle memory values and table correspondence are retained as narrow corroboration; no complete executable vector receives runtime_observed: yes. In particular the exact 1000-frame/no-resync/unchanged-cycle WCK-V11 and drift send WCK-V12 were not captured. No source age was used to renew verification.

Earlier broad negatives about Lost City triggers, saving and other consumers had no six-rung search record. Current TSW-R01c/TSW-R06/WCK-R09 are unresolved within the actually searched boundaries, rather than claiming global absence. Current event/flag/timer/delivery inputs are neutral and supplied; local arithmetic/writes are not globally gated by unknown producers.

Several earlier locator strings confused assembly label offsets with module RVAs. Import resolution establishes start/reset thunks `0xAB98/0xAB92`, flag thunks `0xA568/0xA57A/0xA586`, timer insertion `0x33AF0`, player iteration `0xDFCC0`, item drop `0x11A70`, and gold drop `0x6AFA0`. Original calls and imports, rather than label text, are the public locators. Original full text/locator strings remain privately preserved. These corrections do not independently verify the excluded helper interiors.

### Quest-stage follow-up (2026-10-15)

An independent static re-derivation of the quest stages against D2Game 1.13c reported omissions in the animation, event, altar, sender and resync slices. The reviewer re-read every changed path on 2026-10-15: the disassembly view where it has no undecoded byte runs, and independent decoding of the original D2Game image (preferred base `0x6FC20000`) for `0xC76B0..0xC7B3D`, `0xC5E30`, `0xC5E90`, `0x2CA9B..0x2CB38` and the entry of `0x8A3E0`. Imports were resolved through the original import tables. No reconstruction or binary was changed, and no gameplay was observed.

Several callee addresses in the incoming report were offset by `0x1000` and are corrected here from the original image: participant iteration `0xC3630` (reported `0xC4630`), post-flag iterations `0xC36B0`/`0xC5ED0`/`0xC5090` (reported `0xC46B0`/`0xC6ED0`/`0xC6090`), completion iteration `0xC5260` (reported `0xC6260`), altar timer callback `0xC5E30` (reported `0xC6E30`) and player lookup `0xE03A0` (reported `0xE13A0`). The report's animation finding named the `0xC7130` path; its instructions are in `0xC71A0`, while `0xC7130` is join apply. `0xC5090` starts after the `int3` at `0xC508F`, although the reconstruction groups it with `0xC5030`. D2Game `0x33920`, called by the altar, is not D2Common #11090, which also sits at RVA `0x33920` in D2Common.

Layout mapping for the new semantic names: private markers A/B/C are bytes at private `+4/+5/+6`; mode argument, unit marker, source result and derived result are dwords at `+8/+0xC/+0x10/+0x98`; the amulet count is private `+0x9C`. The roster word is the u16 at quest `+0x9C`. Event key is event `+0x18`, auxiliary identity `+0x14` and event unit `+0x0C`; the animation argument's unit is at `+4`. Unit identity is unit `+0x0C`. The object item-code field is object `+0xB8`. Quest 9's counter is its private `+0x24`. Client direct-player flag is bit 0 of client `+0x3D4`, cached player reference `+0x174`, lookup keys `+0x16C/+0x170/+0x1A8`. The diagnostic switch is D2Game data RVA `0x111C20`.

| Rule | Evidence (D2Game RVAs) |
|---|---|
| TSW-R01e | `0xC71AE` writes marker B=1; `0xC71B1..0xC71C0` writes unit marker from argument unit `+0x0C` or -1. `0xC71C3` active==1, else jump to `0xC7211`. `0xC71C8` unsigned state>1 skips the start at `0xC71CD`; results write pending/applied, then state 3/transition 0 at `0xC71DE/0xC71E2`. `0xC71E6` stage==0 writes transition 0/stage 2 and iterates `0xDFCC0` with `0xC52B0` at `0xC720B`. `0xC721D` calls thunk `0xA4A2` (D2Common #11090) with the argument unit and mode argument on every path. The constant null test on the callback at `0xC71F2` makes the fatal branch `0xC7228` unreachable |
| TSW-R01f, TSW-R01i | `0xC73B2..0xC73D6`: key 0x2C/0x2D, state 0, active 1 call schedule `0xC7250`. `0xC73DB..0xC7404`: key 0x28, pending 1, shared start, progress iteration `0xC3750`, pending clear; both fall through. `0xC7408` tests auxiliary identity 0x28; `0xC740F` u16 roster word >0 calls D2Game `0x33830` at `0xC742B` with quest `+0x1C` and the event unit's identity or -1. `0xC7430` state 2 writes 3 at `0xC7442`, formats with USER32 `wsprintfA` (slot `0x6FD18D58`) and calls `0xCBE70` at `0xC7462` only when the switch is nonzero |
| TSW-R01a, TSW-R04c | In shared start, the state-0 branch writes 1 at `0xC6A6D`, formats at `0xC6A71`, and calls `0xCBE70` at `0xC6A8D` when the switch is nonzero, before the Act II test. Format string at .rdata RVA `0xFCC2C` is `changing state to %d` plus newline; the altar passes `Finished quest` at RVA `0xFB2E0` directly. `0xCBE70` builds a message with opcode `0x26` and the copied text and calls `0x29820` with callback `0xCB580`; that broadcast and its recipients are supplied |
| TSW-R04a | Both `0xC5190` and `0xC5100` test the unit for null, type 0, and player-data `+0x9C` client; any failure reaches the assertion call (line `0x7CB`) and `-1` exit before the act test |
| TSW-R04b | `0xC7145..0xC7173`: the client is read from player data only for a present type-0 unit, else 0. `0x8A3E0` tests the client at `0x8A3EB` and returns at `0x8A48C` without writes when it is absent |
| TSW-R02a..c | `0xC76C3..0xC76FB` missing quest returns 0. Active (`0xC78E5`): `#10331` (thunk `0xA478`) to source result, `#10826` (thunk `0xA556`) to derived result, `#11090`(O,1), mode argument=2, marker B=1, unit marker, marker A=1 (`0xC78E5..0xC7920`); active retest `0xC7923`, reset `0xC792E`, broadcast `0xC7947`, `0x35EB0`(4) `0xC7953`, `0x35E90`(6) `0xC795D`; amulet count 0 and `0xC3630` iteration `0xC7966..0xC797E`; active retest `0xC798A`, flag test/set/set/clear `0xC7997..0xC79B9`, iterations `0xC36B0/0xC5ED0/0xC5090` `0xC79CC..0xC79F2`, D2Game `0x33920`(quest `+0xE0`,13) `0xC7A07`; `#10826`(source result), `+0xB8`=`vip `, drop loop `0x11A70`(O,7,...) with `#10977`(item,16,1) per success `0xC7A0C..0xC7A6F`; `0xC2380`(successes) `0xC7A76`; `0x6B570`(4) `0xC7A7E`; missing game quest structure fatal `0xC7A90`; draw `0xC7AB5`; gold `0x6AFA0` loop; `0xC5260` iteration `0xC7B06`; `0x33AF0`(quest, `0xC5E30`, 10) `0xC7B1A`; return 0. Inactive (`0xC779D`): `#11090`(O,1), marker A=1, mode argument=2, amulet count 0, `0xC3630` iteration, `#10331`/`#10826` unstored, `+0xB8`=`vip `, counted drop loop, `0x6B570`(4), draw `0xC7855` (missing structure jumps to fatal `0xC7A90`), gold loop, quest-9 search and `+0x24` addition at `0xC78CC..0xC78DB`; both exits return 0 |
| TSW-R02d | `0xC5E30`: state 4 writes transition 0/stage 3 and iterates `0xC52B0`; `0xC5E60` clears private `+6`; returns 1 |
| WCK-R06b | `0x2CA9B` tests client `+0x3D4` bit 0 and takes `+0x174` directly; else `0x2CAAC` skips a zero `+0x174`, calls `0xE03A0` at `0x2CAC6` and clears `+0x174` at `0x2CACF` on a zero result; `0x2CAD7` skips an absent player |

Discrepancies corrected: TSW-R01e previously treated the marker writes, stage-0 branch and final mode-update as supplied bookkeeping, and TSW-V09(i)'s active-1/state-2 case asserted every default field retained with no request. Instructions `0xC71E6..0xC720B` show that a zero stage becomes 2 with transition 0 and a pre-start request; the case now asserts that. TSW-R01f described the event's remaining effects as supplied; the auxiliary-identity slice is now TSW-R01i. The single altar rule became TSW-R02a..d with the ordered requests above in place of "supplied preparation/intervening callbacks". TSW-R04a gained the sender fatal path, TSW-R04b the no-client behavior, WCK-R06b the resolution steps, and TSW-R04c the diagnostic broadcast. TSW-Q07 records the cycle-0 server start versus the fixed cycle-5 message. Vector arithmetic of unchanged cases was not affected. The affected cases are TSW-V01 (switch case), TSW-V04(g), TSW-V06(b), WCK-V12 (resolution case), TSW-V09(a)/(i)/(m) and new TSW-V10.

## Search and gap assessment

On 2026-10-15 the dated records above were checked before targeted reading. Quest callback installation, output writes, frame roots, current-clock imports and message-handler rows were compatible evidence anchors; new reads addressed specific missing downstream calls, flag imports and byte decoding. No age-only reread or new gameplay campaign was used.

- TSW-Q01: existing quest callbacks/auxiliary comparisons were reused; targeted initialization and event/animation reading confirms bounded event selection. Selector producers and complete entry triggers remain untraced. No six-rung absence claim.
- TSW-Q02: matched flag wrappers plus Fog operations/masks close exact test/set/clear effects, including bit 14 correction. Gameplay meanings still need consumers.
- TSW-Q03: original command row, lengths, send construction and supplied client assignment were checked. Packet capture remains pending.
- TSW-Q04/WCK-Q04: existing bounded join/getter evidence does not cover all save/consumer paths. No complete six-rung search or runtime save campaign exists; both stay unresolved rather than repeating identical unsuccessful searches.
- TSW-Q05: dated eight-body direct audit and two draw sites reused; targeted seed-site decoding adds explicit boundaries. Item/gold/callback interior reachability remains open. No whole-path no-draw claim.
- TSW-Q06/WCK-Q05: targeted reads expose concrete dependency crossings (timer insertion/delivery, pre-start/status/altar callbacks, player resolution, pre-filter notifications, mode setters, drop operations, presentation and transmission). Their inputs/results/order are supplied in the contract. Their interiors were not silently adopted as verified behavior.

## Observation procedures

These procedures are pending, not records of completed gameplay. Use an owned vanilla LoD 1.13c Expansion profile with matching binaries; record executable/DLL identity, effective data/overrides, difficulty, class/level, state/flags and exact counted frames. Collect entry/return values and the output writer, recording capture capacity, dropped records and relocation. Use a disposable copied character when setup needs progress changes. No memory-call result by itself establishes a normal gameplay branch.

- **TSW-O01:** Normal-difficulty character able to reach the Act II quest, without assuming a class/level restriction; record actual class/level/items/progress. Observe quest/event inputs, shared start return, final server clock and both outgoing messages, then client assignment. Observe eligible altar operation with mode/flag/item-test results; compare another act in the same game. Read immediately around every boundary and retain all callbacks/drop limits. Expected local tuples are TSW-V01/TSW-V02; appearance/event identity are questions, not fabricated expected captures.
- **TSW-O02:** Record flags before joins for fresh, started and completed progress and each tested difficulty. Capture join selection, missing-act failure if naturally reached, timer/retry order and own-act snapshot. Compare the supplied-order cases TSW-V03/TSW-V04/WCK-V10; do not infer save-file encoding from a visual match. Save/load tests require their own normal file-path/encoding evidence.
- **TSW-O03:** Two players with recorded current acts/progress; capture the second joiner's snapshot and either player's eligible altar operation. Record every recipient and ordered World-clock/Quest-event request plus actual client handling. Distinguish stage calls from successful network delivery and sound playback.

## Handoff checks

The main contract has exact local vectors, explicit supplied boundaries and draft status. New source-backed corrections close final-start ticks, client-rate preservation/recomputation, bit 14 clear, supplied threshold arithmetic and bounded intensity arithmetic. Observations, producer identities, complete save/consumer searches and helper integrations remain in the linked gap record. Historical implementation-gap entries are retained as dated records, without a new comparison.

Fresh reader runs FULL1 through FULL10 (FULL1-3 full, DELTA4 a delta check, FULL5 onward full) and their dispositions are recorded below; the latest status is at the end of this record. User approval and publication acceptance are not inferred from this evidence record.

## Fresh reader review: FULL1

On 2026-10-15 a fresh full reader checked all twenty-five executable vector families and their listed cases from rules/inputs, ignoring traces, without source or gameplay access. No subcase count was reported. Counts were MISMATCH 0, GUESS 0, GAP 4, DISCRIM 1, GATED 26, COVERAGE 12 reported in eight rendered topic groups, and HARDEST 3. The reported coverage count and grouped rendering are retained without inventing a different count. This full pass does not itself approve the handoff.

| Finding | Disposition |
|---|---|
| GAP: shared-start return absent from Outputs | Added TSW-R01a return 0/1 to section 5; TSW-V01 and failed TSW-V04 preserve the existing assertions |
| GAP: progress-writer return absent from Outputs | Added TSW-R01b return 0 to section 5, matching TSW-V08 |
| GAP: altar return absent from Outputs | Added the altar return 0 (TSW-R02a through TSW-R02c) to section 5, matching bounded TSW-V05 cases |
| GAP: client-wrapper return absent from Outputs | Added WCK-R04 period-change return 0/1 to section 5; WCK-V14 has both changed-period return 1 and unchanged-period return 0 |
| DISCRIM: A's clock request could be deferred until after B/C pre-filter pairs | WCK-V12 now requires the exact recipient-stage log: A notification, A tracking, A clock, B notification, B tracking, C notification, C tracking. It explicitly rejects batching the send after B/C |
| COVERAGE group: progress writer's bit-0 early return | Added fresh TSW-V08 state-3/flags-{0} case, no writes and return 0 |
| COVERAGE group: pending-event failure and 0x2C/0x2D schedule | Added TSW-V09(f) retained pending/earlier writes on missing-act failure, and TSW-V09(g) exact seeded schedule for each supplied event key |
| COVERAGE group: nonzero state with scheduled zero | Added TSW-V07(f), unchanged seed/scheduled and no request |
| COVERAGE group: timer success | Added TSW-V09(h), exact final applied/scheduled/pending, progress flags, return and request order |
| COVERAGE group: altar slot-10 rejection, staff presence, missing object, inactive work/draw | Added TSW-V05(e)/(f) for the two short-circuit results; TSW-V05(g) stops exactly after eligibility continuation for missing object, retaining the downstream TSW-Q06 gate; TSW-V07(g) supplies inactive context/helper effects and proves no reset plus one inline quest draw/six gold-pile requests. Full helper/drop effects remain gated |
| COVERAGE group: Act-IV and other normal intensity | No new formula or vector: those branches are already excluded from WCK-R05a and explicitly gated by WCK-Q01. Expanding general lighting would cross the current contract boundary |
| COVERAGE group: multiple-act frame order/index | Added separate WCK-V12 two-act frame with ticks 2176/2191 and exact act-0-before-act-3 message/notification order, distinguishing loop index from a client source |
| COVERAGE group: rate zero and cycle/period notification without drift | Added WCK-V13(d) synthetic zero-rate division skip, and WCK-V13(e) cycle change with retained period and zero drift. A standalone period-only change is unreachable from the valid-cycle core because its only period write accompanies a cycle transition; the original return predicate still checks both fields |
| HARDEST: low32 overflow, wrap before transition | No rule change: section 3, WCK-R03 and WCK-V13(a)/(b) already give widths/order/exact outputs. The added zero-rate case makes division skip explicit |
| HARDEST: signed drift, threshold and per-client order | Existing WCK-V13(c) distinguishes i32-minimum negation; WCK-V12 equality/exceeded cases retain strict threshold; the exact log and new cycle-only notification case resolve the ordering ambiguity |
| HARDEST: join failure, flag precedence, state/snapshot order | No new behavior: TSW-V04(e) already requires bit-4 precedence and retained failure writes; TSW-V03/WCK-V10 explicitly supply the quest-join-before-snapshot schedule. Full outer order remains WCK-Q05 |

The twenty-six GATED entries were ten open questions, six unresolved/asm-only leaves and ten supplied-stage vector families. They remain declared gates, rather than missing local values. The coverage additions use existing rules and arithmetic; this pass collected no new source or runtime evidence. Behavioral assertions/vector coverage changed, so a fresh full reader is required before review completion can be claimed.


## Fresh reader review: FULL2

On 2026-10-15 the next fresh full reader returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 5, GATED 20, COVERAGE 8 and HARDEST 3. Its required findings concern the state a supplied operation observes at entry. The revised vectors explicitly define call-time snapshots of semantic state; these are observations in the vector log, not invented function arguments or additional stored fields.

| Finding | Disposition |
|---|---|
| DISCRIM: TSW-V01 could set eclipse flag before first intensity | Added first-intensity snapshot flag 0 with the initial zero ticks and second-intensity snapshot flag 1 with recomputed ticks 1200 |
| DISCRIM: TSW-V02 could clear flag before intensity | Added intensity snapshot flag 1 after reset fields/applied clear and before flag clear; final flag remains 0 |
| DISCRIM: TSW-V09(b) could clear pending before progress callback | Progress-writer entry snapshot requires pending 1/applied 0/state 1/stage 1, followed by the pending-clear write |
| DISCRIM: TSW-V09(h) could set applied after progress callback | Progress-writer entry snapshot requires applied 1/pending 0/scheduled 0/state 1/stage 1 |
| DISCRIM: WCK-V16 could store new flag before first intensity | First-intensity snapshot requires old flag 0 and supplied ticks 0; second requires flag 1 and recomputed ticks 30720 |
| COVERAGE: altar bit-0-only and absent quest | Added TSW-V05(h) primary bit-1 false/bit-0 true/slot-10 true rejection; TSW-V05(i) absent quest returns 0 without quest/flag work. No private-state input is assigned to an absent quest |
| COVERAGE: animation predicate false | TSW-V09(i) supplies fresh active 0/2 at state 0 and active 1 at state 2; each leaves the start slice unchanged |
| COVERAGE: unrelated event key | TSW-V09(j) key 0x29 retains pending/clock/state with no start/schedule/draw from the slice |
| COVERAGE: Act-IV intensity | No change: this branch is already explicitly gated under WCK-Q01 and excluded from the two verified intensity branches |
| COVERAGE: invalid direct retry | TSW-V09(k) distinguishes wrong act argument, absent quest, and pending 2; no presentation or quest-start request. An absent quest has no private-state input/output |
| COVERAGE: invalid client cycle, negative ticks, missing environment | No invented fallback: WCK-R08b's supplied-input domain requires a valid cycle/nonnegative ticks/environment. Those invalid/assertion paths are outside its normal-result boundary and remain subject to WCK-Q05 if integrated behavior is requested |
| COVERAGE: existing server act without environment | WCK-V13(f) increments the normal frame, returns 0 from the act update, retains absence and makes no clock/notification request |
| COVERAGE: other altar remainder counts | No exhaustive matrix added: TSW-R07b consumes the same NRF-R01/NRF-R02 step and modulo-5 recipe; TSW-V07(e)/(g) supply exact seed/result/count and local context. There is no additional conditional branch by remainder |
| HARDEST: WCK-V12 interleaved acts/recipients | No rule change: its exact logs now distinguish per-client interleaving and act-index traversal; these assertions were retained |
| HARDEST: TSW-V09 retained writes/call-time state | Added explicit pending/applied snapshots and focused guard/failure cases; earlier failure writes remain exact |
| HARDEST: WCK-V13 overflow/resync | No rule change: low32 negation, wrap, strict threshold, zero-rate skip and cycle-only notification each retain exact fixtures |

The twenty GATED entries remain explicit integration, producer, observation or valid-input-boundary limits. No source/gameplay collection occurred in this response to review. The amended assertions/cases require a fresh full pass; this record does not claim review completion or user approval.


## Additional focused review after FULL2

An additional focused reader checked the amended TSW-V01/TSW-V02/TSW-V09(b..h)/WCK-V16 snapshots and added TSW-V05(h..i), TSW-V09(i..k), WCK-V13(f), with their shared defaults and governing rules. It returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 2, COVERAGE 0 and HARDEST 3. Its gates were later altar drop/helper effects (TSW-Q05/TSW-Q06) and other animation/event work (TSW-Q06); its hardest entries were the absent-quest altar, timer/progress snapshot and existing-act/absent-environment cases. These declared boundaries needed no new mechanics. This was focused evidence and did **not** replace the third required full reader.

## Fresh reader review: FULL3

On 2026-10-15 the third fresh full reader recomputed all twenty-five vector families and listed subcases. It returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 2, GATED 16, COVERAGE 6 and HARDEST 3. No subcase count was reported. The sixteen gates were the ten open questions and six unresolved/asm-only leaves.

| Finding | Disposition |
|---|---|
| DISCRIM: WCK-V13(d) zero-rate/ticks-0 fixture also passed an incorrect entire-update skip | Changed the supplied starting ticks to -2. The core must advance to -1, remain below both zero thresholds, skip division, retain last resync 17 and return 0. A full skip yields -2; a forced wrap yields 0. The former ticks-0/expected-0 fixture was arithmetically valid but did not discriminate those alternatives |
| DISCRIM: WCK-V14 could reverse intensity/colour or omit established request inputs | Every supplied source case now requires intensity with that source first, colour with its derived act second, and explicit post-core clock snapshots at both entries. Source 120 gives intensity(source 120) then colour(act 4), before the final colour override. The two additional period-return cases have their exact ordered requests/snapshots too |
| COVERAGE: shared start with nonzero state | Added TSW-V01 fresh state-2 case retaining state 2 while producing the same start clock, marker writes and ordered requests |
| COVERAGE: bit-1 progress guard | No change: TSW-V08 already supplies state 3/flags {1}, returns 0 and writes no flags |
| COVERAGE: join bit 3 without bit 4 | Added TSW-V04(f), flags {2,3}, exact stage/state 2/2, applied 1 and start clock |
| COVERAGE: schedule event active other than exactly 1 | Added TSW-V09(l), each supplied event key with active 2, unchanged quest seed/scheduled/clock and no request |
| COVERAGE: altar active nonzero other than 1 | Added TSW-V05(j), active 2, the same bounded reset/flag/request result as active 1, excluding later helper/drop effects |
| COVERAGE: other modulo-5 remainder values | No additional matrix: the unchanged NRF-R01/NRF-R02 reduction recipe and TSW-R07b formula have no remainder-specific branch; exact seed/result/count fixtures already distinguish their owner/order/budget |
| HARDEST: low32 wrap/signed comparisons | No rule/gate change; WCK-V13 retains explicit overflow, signed-negation, wrap and threshold outputs, with a stronger zero-rate discriminator |
| HARDEST: drift and recipient interleaving | No rule/gate change; WCK-V12 retains its exact per-client/per-act logs and WCK-V13 its distance/cycle-only cases |
| HARDEST: join failure/precedence/snapshot order | Existing TSW-V04(e), TSW-V03/WCK-V10 supplied scheduling and failure-state assertions remain exact; TSW-V04(f) adds the distinct bit-3-only selection |

Only vectors and their review record changed in this pass. No rule behavior, ownership, acceptance boundary or gate changed, and no new source/gameplay evidence was collected. After three full passes, the next applicable fresh review may be a delta check of the amended vectors and the findings above. Review completion/user approval remain unclaimed pending that check.

## Fresh delta review 4 and review completion

After the three full reviews, the fresh delta reader checked every amended case listed after FULL3: TSW-V01 nonzero state, TSW-V04(f), TSW-V05(j), TSW-V09(l), WCK-V13(d) and all WCK-V14 request/snapshot assertions, with their shared defaults and governing rules. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 3, COVERAGE 0, HARDEST 3. It independently rejected skipping the zero-rate update, reversed presentation calls, wrong source/act arguments and incorrect flag/state predicates. The earlier extra focused review remains separately recorded and is not counted as a full review.

The three reviewed gates receive no change: WCK-Q01 retains actual presentation effects, WCK-Q02 retains source production/outer client scheduling and TSW-Q06/WCK-Q05 retains supplied integration/delivery effects. All three HARDEST entries receive no change: WCK-V14 explicitly maps each source, updates the core and orders intensity(source) before colour(derived act), with exact entry snapshots; WCK-V13(d) advances -2 to -1 before skipping division; TSW-V04(f) preserves flags while selecting stage/state 2/2. No new rule, ownership, acceptance or gate change was made in this final triage.

Root review read the complete contract, companion and gap record, all additions and every disposition; it independently checked the original start/recompute/colour images and the relevant threshold, seed and drift arithmetic. Every blocking reader finding is resolved on the resulting contract. Document review is complete, with draft status, exact integration/producer/observation gaps and maintainer approval remaining. No new gameplay, implementation test or implementation comparison ran. Publication repeats document checks on committed bytes.

## Review status after the quest-stage follow-up

The 2026-10-15 quest-stage follow-up changed rule behavior (TSW-R01e, TSW-R01f, TSW-R01i, TSW-R02a..d, TSW-R04a, TSW-R04b, TSW-R04c, WCK-R06b), corrected TSW-V09(i), and added vector cases and TSW-V10. The review completion above applies to the earlier contract only. Because rule ownership changed (TSW-R02 split, new TSW-R01i/TSW-R04c), the next reader runs are full reviews, numbered FULL5 onward in the overall sequence.

## Fresh reader review: FULL5

On 2026-10-15 a fresh full reader (smaller model, spec plus foundations and game-data-fields only; companion used only for the observation-procedure anchor) recomputed every case of TSW-V01 through TSW-V10. Counts: MISMATCH 1, GUESS 8, GAP 12, DISCRIM 16, GATED 4 groups, COVERAGE 20 groups, HARDEST 3.

Triage re-read original code where a behavioral statement changed: D2Common `0x37570` (comparison table selection, night test indexed by cycle), `0x37CA0` and `0x376F0` (intensity takes source/act only and reads the stored flag), `0x37920` (fatal assertion on cycle outside 0..5 or negative ticks before any write), D2Game `0x33AF0` (due value = game quest structure counter + parameter), `0xC73A0` (pending exactly 1; state exactly 2), `0xC5E30` (state exactly 4), the original-image altar decode `0xC76B0..0xC7B3D` (primary/secondary test order, signed amulet loop re-reading the count, both active re-tests, success counting) and `0x2CA20` (per-client resolution then requests, before the next client).

| Finding | Disposition |
|---|---|
| MISMATCH WCK-V01 trace threshold 20480 vs 23040 | No change to outputs: `0x37570` compares flag 0 against the normal table, so 20480 is correct. Added WCK-V01(b) to pin it |
| GUESS shared vs player bit 13 (TSW-R01d, TSW-V04(b), WCK-V10(b)) | Spec edit: TSW-R01d and a new Inputs row name the game-level shared quest-flag record; TSW-V04(b)/WCK-V10(b) assert P's flags unchanged |
| GUESS second intensity flag argument; GUESS "incoming flag" with flag already 1 | Spec edit: section 3 states intensity carries source/act only and observes the stored flag; WCK-R01b/WCK-R01c/WCK-R05a/WCK-R08b reworded |
| GUESS resolution state of WCK-V11/WCK-V12/WCK-V13 clients | Spec edit: section 8.1 defaults define resolved/unresolved clients |
| GUESS TSW-V10 log omits TSW-R02a tests | Spec edit: TSW-V10(a)/(b) logs start with the two primary tests; defaults state flag/item tests are log entries |
| GUESS amulet count type/zero | Spec edit: TSW-R02b step 5 states the signed loop with re-read count; count <=0 requests none |
| GUESS WCK-R08b eligibility | Spec edit: new supplied handler-eligibility input; WCK-R08b evaluates it first |
| GUESS join/snapshot order (TSW-V03/WCK-V10) | No change: already supplied by the vectors and WCK-R07/WCK-Q05 |
| GAP offset-like marker names | Spec edit: renamed to markers A/B/C, mode argument, unit marker, source result, derived result; companion layout mapping keeps the binary offsets |
| GAP unnamed outputs in section 5 | Spec edit: Outputs rows added for each named request, schedule due value, current-clock read and change tracking |
| GAP supplied-operation arguments | No change: arguments the contract asserts are in rules/vectors; others are TSW-Q06/WCK-Q05 integration |
| GAP scheduled due value not computable | Spec edit: default game quest counter 1000; TSW-V07(b)/(c)/(h) assert due 1015/1016 |
| GAP TSW-V06(b) act of P2 | Spec edit: P2 in act 0 |
| GAP TSW-V04(g) absent/non-player joining unit | No change: TSW-R04b gives those inputs the same no-client request, which queues nothing; the observable outcome equals TSW-V04(g) |
| GAP TSW-V05(g) absent object downstream | No change: already gated under TSW-Q06 in the case text |
| GAP WCK-R08b eligibility/invalid message | Spec edit: eligibility input; invalid cycle/ticks fatal path added to WCK-R08b, Outputs and WCK-V16 |
| GAP WCK-R01a allocation request contents | No change: supplied presentation, WCK-Q01/WCK-Q05 |
| GAP TSW-R07b wording | Spec edit: `5+(new lo mod 5)` |
| GAP WCK-R05a other branches | No change: gated by WCK-Q01 |
| GAP WCK-R06b lookup arguments/placement | Spec edit: lookup(client), per-client completion; WCK-V12 resolution case uses one ordered log |
| DISCRIM WCK-V01/WCK-V09/WCK-V14 Act IV table with flag 0 | Spec edit: WCK-V01(b) |
| DISCRIM period field vs normal[cycle].period | Spec edit: WCK-V02(b) |
| DISCRIM WCK-V13(e) storing q on cycle-only notify | Spec edit: last resync 170, retained |
| DISCRIM schedule parity from pre-step lo | Spec edit: TSW-V07(h) seed {0,667} |
| DISCRIM primary set, all secondary false | Spec edit: TSW-V05(k) |
| DISCRIM flag block regardless of bit 0 | Spec edit: TSW-V05(l) |
| DISCRIM altar active re-tests | Spec edit: TSW-V10(e)/(f) |
| DISCRIM TSW-R01i state>=2 | Spec edit: TSW-V09(m5) |
| DISCRIM TSW-R01f pending nonzero | Spec edit: TSW-V09(n) |
| DISCRIM TSW-R01e success path | Spec edit: TSW-V09(o) |
| DISCRIM TSW-R01a diagnostic for state 1 | Spec edit: TSW-V01 state-1 case |
| DISCRIM TSW-R02d state>=4 | Spec edit: TSW-V10(d) state 5 |
| DISCRIM WCK-V16 rate/index/last reset | Spec edit: WCK-V16 rate-4 case |
| DISCRIM WCK-V12 lookups batched | Spec edit: single ordered log |
| DISCRIM TSW-V06(b) act before client | Spec edit: P2 act 0 |
| DISCRIM TSW-V04(g) absent/non-player | No change: equivalent outcome under TSW-R04b (see GAP above) |
| GATED TSW-Q07 uncited | Spec edit: TSW-Q07 added to TSW-R04a's Gates column for the observed effect only |
| COVERAGE (20 groups) | Advisory. Branches with new result-changing cases were covered above (TSW-R01e success, TSW-R05-adjacent pending guards, TSW-R02a proceed, TSW-R02b re-tests, WCK-R08b invalid, TSW-R02d other states). Remaining groups need no vector: they reuse already-discriminated arithmetic/guards or sit behind declared gates (WCK-Q01/TSW-Q04/WCK-Q02/TSW-Q06/WCK-Q05, observation WCK-O01) |
| HARDEST TSW-V10 / TSW-R02b | Resolved by the full logs, TSW-R02a tests in log, and re-test cases |
| HARDEST WCK-V12 / WCK-R06b | Resolved by client-resolution defaults and the single ordered log |
| HARDEST TSW-V09 slices | No further change; snapshots and new (m5)/(n)/(o) cases fix the guards |

Rules, inputs, vectors and gates changed, so a further fresh full reader is required.

## Fresh reader review: FULL6

On 2026-10-15 the next fresh full reader recomputed every case of TSW-V01 through TSW-V10, including all FULL5 additions. Counts: MISMATCH 0, GUESS 8, GAP 9, DISCRIM 11, GATED 28, COVERAGE 31, HARDEST 3. Triage re-checked D2Common `0x37920` (source-120 override after both flag paths), `0x37570` (nonzero flag skips every advance bonus), `0x37CA0` (no last-resync write), `0x376F0` (flag-nonzero branch precedes the source-120 assignment), D2Game `0x33AF0` with the altar call at `0xC7B1A` (parameter 10), the altar decode (eligibility before mode marker; success counter added to quest 9), and the join bodies `0xC72C0`/`0xC7130` (no progress-writer iteration).

| Finding | Disposition |
|---|---|
| GUESS WCK-R08b source-120 override with flag 0 | Spec edit: WCK-R08b states the override applies without a colour request; WCK-V16 source-120 case |
| GUESS WCK-V12/WCK-V13 act-0 existence | Spec edit: defaults say a case naming an act supplies it |
| GUESS TSW-V06 P2 client | Spec edit: both players have clients |
| GUESS P's client state | Spec edit: defaults give P a client in act 1, state 4 |
| GUESS WCK-V10(c) Act I clock | Spec edit: Act I exists with the supplied clock, not reallocated |
| GUESS join runs no progress writer | Spec edit: TSW-R01d states it (join bodies make no progress-writer iteration) |
| GUESS TSW-V05(g) absent object | No change: gated under TSW-Q06 in the case |
| GUESS schedule order without snapshot | No change: order only is the intended assertion |
| GAP current-clock read placement | Spec edit: WCK-V12 and WCK-V13(b) logs list current-clock read(act) |
| GAP altar timer due value | Spec edit: TSW-R02b step 7 and Outputs; TSW-V10(a) due 1010 |
| GAP quest initialization has no vector | No change: only its no-draw property (TSW-R07a, TSW-V07(a)) and zeroed defaults are claimed |
| GAP colour request arguments | No change: section 3 and WCK-R01b give colour(act); colour bytes are presentation (WCK-Q01) |
| GAP WCK-V13(b) recipient | Spec edit: World-clock(client,1,0,0) |
| GAP symbolic helper effects | No change: TSW-Q06/WCK-Q05 |
| GAP TSW-V05 partial logs | No change: TSW-V05 is a bounded slice; TSW-V10 carries the complete altar logs |
| GAP allocation requests | No change: supplied presentation, WCK-Q01/WCK-Q05 |
| GAP outline note | No change: reader found no layout dependency |
| DISCRIM WCK-V08 table/bonus under flag 1 | Spec edit: WCK-V08(b) |
| DISCRIM flag-nonzero bonuses | Spec edit: WCK-V09(d) |
| DISCRIM TSW-V01 last resync reset | Spec edit: TSW-V01 last-resync-17 case |
| DISCRIM TSW-V05 mode before eligibility | Spec edit: TSW-V05(m) |
| DISCRIM TSW-V10 attempts vs successes | Spec edit: TSW-V10(b) second drop fails, counter 5 |
| DISCRIM unconditional Finished-quest text | Spec edit: TSW-V10(a) switch-0 variant |
| DISCRIM TSW-V09(h) pending cleared on timer success | Spec edit: TSW-V09(h) pending-1 variant |
| DISCRIM TSW-V02 unfiltered reset | Spec edit: TSW-V02(b) |
| DISCRIM WCK-V14 signed vs unsigned negative source | No change: equivalent for every input (both map negatives to act 0); the WCK-V14 claim was reworded to say so |
| DISCRIM TSW-V03 skipping overwrite | No change: TSW-V04(e)/(f) are the owning discriminating cases |
| DISCRIM WCK-V15 act 3 / flag 1 with source 120 | Spec edit: flag-1/source-120 case gives 192; act 3 stays gated by WCK-Q01 |
| GATED (28) | No change: declared gates |
| COVERAGE (31) | Advisory. Cases added above cover the result-changing ones raised (flag-1 advance, Act IV not exceeded, source 120 handler); the rest reuse discriminated arithmetic or sit behind declared gates |
| HARDEST WCK-V12, TSW-V10, WCK-R03/WCK-R02 | No further change beyond the logs and WCK-V01(b)/WCK-V08(b)/WCK-V09(d) cases above |

Vectors and a rule's stated outputs changed, so another fresh reader is required.

## Fresh reader review: FULL7

On 2026-10-15 the next fresh full reader recomputed every case. Counts: MISMATCH 1, GUESS 9, GAP 9, DISCRIM 14, GATED 36, COVERAGE 20, HARDEST 3. Triage re-checked D2Game `0xC6A30` (transition/stage written before the pre-start request; state change and diagnostic after it, before the Act II test), `0xC7250` (schedule request precedes the scheduled write), the altar decode (inline draw after the amulet loop; mode marker tested as nonzero), `0x2CA20` (current-clock read once per notifying act before the client loop; act filter compares client act with the loop index), and D2Common `0x37660` (signed IDIV truncation).

| Finding | Disposition |
|---|---|
| MISMATCH WCK-V13(e) log missing current-clock read | Spec edit: log begins with current-clock read(act 0) |
| GUESS WCK-V01/WCK-V02 rate | Spec edit: rate 128 stated |
| GUESS TSW-R04a act source | Spec edit: act is the client's act; client state not tested; TSW-V06(b) discriminator reworded |
| GUESS TSW-V06(b) order before abort | Spec edit: P precedes P2; P's requests precede the fatal path |
| GUESS TSW-V02(b) client state | Spec edit: TSW-R04a says client state is not tested |
| GUESS WCK-R06b matching act | Spec edit: client act equal to server act index |
| GUESS current-clock read result | Spec edit: read(act index) yields post-update cycle/ticks/flag |
| GUESS mode argument read time | Spec edit: TSW-R01e uses the current value at the final request |
| GUESS TSW-R01a order | Spec edit: TSW-V01 pre-start snapshot state 0/stage 1/transition 0 |
| GUESS absent-object altar | No change: TSW-Q06 |
| GAP TSW-V07(a)/quest initialization | No change: only the no-draw property is claimed; the transition default is a quest-record field, not a private marker |
| GAP TSW-V07(g) quest 9 | Spec edit: quest 9 absent |
| GAP WCK-V10 allocation requests; GAP WCK-R04 colour/intensity values; GAP TSW-V10 symbolic results; GAP TSW-V05(g) | No change: supplied presentation/integration under WCK-Q01/TSW-Q06/WCK-Q05 |
| GAP WCK-R08b flag range | No change: section 3 already says any nonzero flag is eclipse |
| GAP WCK-R07 absent act clock | No change: section 3 requires the act/environment unless a rule handles absence; the snapshot has no absence branch in its contract |
| GAP constants evidence | No change: records exist in this companion |
| DISCRIM TSW-V01 diagnostic after Act II test | Spec edit: TSW-V01 Act II absent/switch 1 case |
| DISCRIM TSW-V01 pre-start write order | Spec edit: pre-start snapshot |
| DISCRIM TSW-V04 flags {1}/{0} alone | Spec edit: TSW-V04(h) |
| DISCRIM TSW-V03 overwrite order | No change: TSW-V04(e)/(f) discriminate |
| DISCRIM TSW-V05(c) mode marker values | Spec edit: markers 1 and -1 |
| DISCRIM TSW-V06 two Act II recipients | Spec edit: TSW-V06(c) |
| DISCRIM TSW-V07 parity in (b)/(c) | No change: TSW-V07(h) rejects it |
| DISCRIM TSW-V07 scheduled write order | Spec edit: schedule snapshot scheduled 0 |
| DISCRIM altar draw before amulets | Spec edit: TSW-V10(a) seed snapshots at the first amulet drop and first gold drop |
| DISCRIM animation state<=1 | Spec edit: TSW-V09(p) |
| DISCRIM key 0x2C with pending 1 | Spec edit: TSW-V09(q) |
| DISCRIM WCK-V16 table by old flag | Spec edit: WCK-V16 message (1,0,1) case |
| DISCRIM WCK-V03 normal table under flag | No change: WCK-V07/WCK-V08 discriminate |
| DISCRIM floor vs truncating q | Spec edit: WCK-V13(g) |
| GATED (36) | No change: declared gates |
| COVERAGE (20) | Advisory; result-changing items raised as DISCRIM were covered above, the rest reuse discriminated guards or sit behind declared gates |
| HARDEST TSW-V10, WCK-V12/WCK-V13, TSW-V09 | No further change |

Vectors and rule wording changed; a further fresh full reader is required.

## Fresh reader review: FULL8

On 2026-10-15 the next fresh full reader recomputed every case. Counts: MISMATCH 0, GUESS 7, GAP 9, DISCRIM 9, GATED 8 groups, COVERAGE 14, HARDEST 3. Triage re-checked D2Game `0xC72C0` (join receives its own quest record; no quest-10 search exists), D2Common `0x376F0` (reads and writes the stored intensity field; flag tested as nonzero), `0x37570` (next cycle 6 wraps to 0 before the comparison) and the altar decode (amulet count re-read before every attempt; the active branch never touches quest 9).

| Finding | Disposition |
|---|---|
| GUESS WCK-R04 intensity request's act | **Unresolved after the cap.** Not yet checked against `0x37AF0`; its available instruction view is undecoded byte runs and the original image was not reachable in this pass. No value was assumed in the spec; WCK-V14 still asserts only intensity(source) and colour(derived act) |
| GUESS intensity persistence | Spec edit: WCK-R05a says each call updates the stored intensity field, so start's two calls compound |
| GUESS pre-start for a clientless player | No change: pre-start iteration interior is TSW-Q06 |
| GUESS TSW-R01e active re-read | Spec edit: active tested once (`0xC71C3`) |
| GUESS WCK-R08b target environment | Spec edit: the client's current-act environment |
| GUESS TSW-R01d quest 10 absent | No change: join is installed on quest 10 and receives that record; absence is not an input |
| GUESS negative rate | No change: rates are the fixed 128/4/8 |
| GAP TSW-V07(a) inputs | No change: only the no-draw property is claimed |
| GAP rate index 2 unused | No change: fixed code constant with its data record; no rule selects it in this boundary |
| GAP intensity/colour initial values | No change: presentation under WCK-Q01 |
| GAP altar object identity/item-code field | Spec edit: added to the altar Inputs row |
| GAP TSW-V05(d) contradictory fixture | Spec edit: case states Act II absent means the preamble clock is not present |
| GAP absent-object altar arguments | No change: TSW-Q06 |
| GAP recipient act/state not tied to WCK-R04 | No change: server recipients use client act/state supplied inputs, independent of the client-side mapping (section 2) |
| GAP auxiliary selector unused by verified rules | No change: consumed only by unresolved TSW-R01c |
| GAP Quest-event flag/bit 13 meanings | No change: TSW-Q02/TSW-Q03 |
| DISCRIM TSW-V04 progress writer in join | No change: equivalent. After the overwrite the writer would add only bits already required for that state selection, so no join input distinguishes it; TSW-R01d states join requests no progress writer |
| DISCRIM TSW-V02 client state / other acts | Spec edit: TSW-V02(c) |
| DISCRIM WCK-V14 intensity act | Unresolved with the GUESS above |
| DISCRIM WCK-V15 flag == 1 | Spec edit: stored flag 2 case |
| DISCRIM TSW-V10 count captured once | Spec edit: TSW-V10(g) |
| DISCRIM TSW-V10(a) quest 9 on active branch | Spec edit: TSW-V10(a) supplies quest 9 counter 4, retained |
| DISCRIM WCK-V16 recompute on stored flag | Spec edit: WCK-V16 stored-flag-1/message-flag-0 case |
| DISCRIM TSW-R01a pending clear | No change: TSW-V09(b)'s progress-writer snapshot rejects it |
| DISCRIM cycle 5 to 0 | Spec edit: WCK-V06(b) |
| GATED | No change: declared gates |
| COVERAGE (14) | Advisory; no further vectors beyond the cases above |
| HARDEST TSW-V09, WCK-V12/WCK-V13, TSW-V10 | No further change |

The four-run cap for this revision was reached with FULL8. Its edits changed vectors and rule wording, so a fresh reader (FULL9) is still required. One item remains open: the client wrapper's intensity act (WCK-R04 / WCK-V14). Review completion is not claimed.

### Later outcome of the two open FULL8 entries (2026-10-15)

| Finding | Disposition |
|---|---|
| GUESS WCK-R04 intensity request's act | Spec edit after asm check: D2Common `0x37AF0` loads the derived act into the act register read by `0x376F0` (the core `0x37570` preserves it), so WCK-R04 now requests intensity(source, derived act) |
| DISCRIM WCK-V14 intensity act | Spec edit: every WCK-V14 case asserts intensity(source, derived act); the discriminator rejects a fixed act 0 at sources 40..1023 |

This changes a rule input, so the run after FULL9 must be full.

## Fresh reader review: FULL9

On 2026-10-15 a fresh full reader (smaller model; spec, foundations, game-data fields; companion for links only) recomputed every case of TSW-V01 through TSW-V10 on the revision before the WCK-R04 intensity-act edit. Counts: MISMATCH 0, GUESS 6, GAP 9, DISCRIM 13, GATED 24, COVERAGE 24, HARDEST 3. Its findings were triaged together with FULL10 below; that table shows each entry's outcome. The WCK-R04 input change made after FULL9 was started requires a further full run in any case.

## Fresh reader review: FULL10

On 2026-10-15 a second fresh full reader recomputed every case on the revision that includes the WCK-R04 intensity(source, derived act) edit. Counts: MISMATCH 0, GUESS 7, GAP 10, DISCRIM 7, GATED 7 groups, COVERAGE 25, HARDEST 3. Triage re-checked D2Common `0x37570` (night bonus tests normal-row period equal to 2), `0x37660` (rate 0 jumps from the division to the cycle/period comparison; absent environment returns 0), `0x37920` (pre-clamp uses the stored rate; the message flag is stored as passed and tested as nonzero for both table choice and recomputation; colour receives the derived act from the source-to-act helper; override after both paths), D2Game `0xC71A0` (state compared with 1 as an unsigned byte; marker B and unit marker written before the active test and the final mode-update) and the altar's original bytes at `0xC7A26..0xC7A71` (count read, `jle` skips for a count <=0, successes counted separately, count re-read with a signed `jl` before each attempt).

| Run | Finding | Disposition |
|---|---|---|
| FULL9 | GUESS WCK-R06b rate 0 continuation | Spec edit: rate 0 goes directly to the cycle/period test (`0x37660`). With rate 0 the wrap threshold is 0 and transitions need ticks >0, so no cycle change can arise; no vector added |
| FULL9 | GUESS/GAP pre-start one request vs per player; multi-player logs | Spec edit: vector defaults define one log entry per iterated player and "after presentation" log abbreviation |
| FULL9 | GUESS quest seed identity | Spec edit: section 3 RNG row names the game-level quest structure's seed |
| FULL9 | GUESS WCK-R08b colour act | Spec edit: recompute intensity and colour use the client source's derived act (`0x37920`) |
| FULL9, FULL10 | GUESS WCK-R08b raw vs normalized flag | Spec edit: flag stored as received; WCK-V16 (1,0,2) case |
| FULL9 | GUESS TSW-R01e marker writes before mode-update | Spec edit: TSW-V09(i) mode-update snapshot (rule already ordered them) |
| FULL9 | GAP quest initialization has no rule | No change: only its no-draw property is claimed (TSW-R07a) |
| FULL9 | GAP TSW-V05(g) unexecutable tail | No change: gated TSW-Q06, stated in the case |
| FULL9 | GAP TSW-V05(d) fixture | No change: corrected after FULL8 (preamble clock not present) |
| FULL9, FULL10 | GAP WCK-R07 absent environment / act >=5 | No change: section 3 requires the act/environment; acts are 0..4 |
| FULL9, FULL10 | GAP intensity outside two branches; initial intensity/colour values | No change: WCK-Q01 |
| FULL9, FULL10 | GAP WCK-R01a allocation failure; lookup success with absent player | No change: rule is stated for successful allocation (WCK-Q05); an absent resolved player is already skipped |
| FULL9 | GAP direct flag with zero cached reference | No change: the cached reference is used as is and an absent player is skipped |
| FULL9 | GAP marker C never set to 1 | No change: writer outside this boundary, TSW-Q02 |
| FULL9 | DISCRIM night test period >=2 | Spec edit: WCK-V09(e) |
| FULL9 | DISCRIM fixed wrap threshold | Spec edit: WCK-V07(c) |
| FULL9 | DISCRIM progress writer clearing stale bits | Spec edit: TSW-V08 state-1/{2,3,4} case |
| FULL9, FULL10 | DISCRIM TSW-V04 (2,3) needs bit 3 too | Spec edit: TSW-V04(i) |
| FULL9 | DISCRIM toggling bit 14 | Spec edit: TSW-V05(n) |
| FULL9 | DISCRIM amulet count reset omitted | Spec edit: TSW-V10(h) |
| FULL9 | DISCRIM WCK-V16 recompute only on 0-to-nonzero | Spec edit: WCK-V16 (5,999,1) case |
| FULL9 | DISCRIM WCK-V16 flag == 1 | Spec edit: WCK-V16 (1,0,2) case |
| FULL9, FULL10 | DISCRIM colour override for every act-4 source | Spec edit: WCK-V14 asserts no override at 109/1023 |
| FULL9 | DISCRIM extra applied/pending schedule guard | Spec edit: TSW-V09(r) |
| FULL9 | DISCRIM TSW-R01e active re-read after start | No change: equivalent within the contract. The stage branch is reachable after a run start only if a supplied callback rewrites stage to 0, and vector callbacks make no unlisted state changes; when the start is skipped, no request intervenes before the stage test. Marker order part: spec edit TSW-V09(i) snapshot |
| FULL9 | DISCRIM WCK-V11 drift threshold weak | No change: WCK-V12 ticks-2047 case rejects it |
| FULL9 | DISCRIM WCK-V01(a) flat +15 | No change: WCK-V09(a) rejects it |
| FULL10 | GUESS WCK-R06b absent-environment wording | Spec edit: "an existing act whose environment is absent returns 0" |
| FULL10 | GUESS WCK-R05a other branches | No change: WCK-Q01 |
| FULL10 | GUESS act 3 never night | No change: already stated; WCK-V09 shows it |
| FULL10 | GUESS TSW-V04(e)/(g) P act/client | No change: defaults already yield to case-stated values |
| FULL10 | GUESS callbacks that set state | No change: modelled as stated in TSW-V10(e)/(f)/(g) |
| FULL10 | GUESS TSW-R01d which player | Spec edit: the joining player's bits |
| FULL10 | GAP WCK-V12 includes P's default client? | Spec edit: listed resync clients are the whole list |
| FULL10 | GAP WCK-V11 Act II default | Spec edit: WCK-V11 says Act II absent too |
| FULL10 | GAP TSW-R01d quest 10 / shared record absent | No change: join is installed on quest 10 and receives it (FULL8 record); flag records are required inputs |
| FULL10 | GAP TSW-R01g missing game quest structure | Spec edit: section 3 lists it as a required record for TSW-R01g; no absence behavior is claimed |
| FULL10 | GAP TSW-V05(g) absent object | No change: TSW-Q06 |
| FULL10 | GAP schedule log omits callback identity | Spec edit: schedule(TSW-R01h timer callback, n) in TSW-V07/TSW-V09 |
| FULL10 | GAP section 5 lacks flag/item test requests | Spec edit: added to section 5 |
| FULL10 | DISCRIM default (1,1) overwrite skipped | Spec edit: TSW-V04(j) |
| FULL10 | DISCRIM signed state 255 | Spec edit: TSW-V09(s) (`0xC71A0` unsigned compare) |
| FULL10 | DISCRIM fixed pre-clamp rate | Spec edit: WCK-V16 rate-4 cases |
| FULL10 | DISCRIM unsigned amulet compare | Spec edit: TSW-V10(h) count -1 |
| FULL10 | DISCRIM mode argument/markers before mode-update | Spec edit: TSW-V10(a)/(b) snapshots |
| both | GATED | No change: declared gates |
| both | COVERAGE | Advisory; result-changing branches raised as DISCRIM are covered above; the rest are gated or reuse discriminated guards |
| both | HARDEST TSW-V10, WCK-V12/WCK-V13, TSW-V09 | No further change; the added snapshots and cases address the ordering points they name |

Vectors, a rule input (WCK-R04) and rule wording (WCK-R08b, WCK-R06b) changed, so the next run is FULL11.

## Fresh reader review: FULL11

On 2026-10-15 a fresh full reader recomputed every case on the revision with the FULL9/FULL10 edits. Counts: MISMATCH 0, GUESS 8, GAP 9, DISCRIM 15, GATED 5 groups, COVERAGE 25, HARDEST 3. Triage re-checked D2Game `0xC7250` (state and scheduled tested as nonzero bytes; a null game quest structure takes the assertion call before the draw), `0xC6AF0` (scheduled cleared first; applied tested as nonzero), `0xC7130` (failure writes pending 1 and requests the Quest-event; success writes applied 1 only), `0xC5E30` (state compared with 4 only; no stage test), `0x2CA20` (direct-player test is bit 0 of the client flag word), D2Common `0x37BA0` (allocation requests intensity with source/act 0, then colour with act 0), `0x37920` (negative and >=6 cycles and negative ticks each assert) and `0x37AF0` (the ticked environment is the supplied act's).

| Finding | Disposition |
|---|---|
| GUESS WCK-R04 which environment is ticked | Spec edit: the caller-supplied act's environment; derived act is only the core act argument and presentation act |
| GUESS TSW-R02b step 2 notifications inside the active test | Spec edit: "still inside this active test" (TSW-V10(e)/(f) already show it) |
| GUESS TSW-R02b step 4 slot | Spec edit: slot 11 named for each write |
| GUESS TSW-R01d pending on success | Spec edit: success leaves pending unchanged; TSW-V04(k) |
| GUESS WCK-R01a request order | Spec edit: intensity, then colour |
| GUESS absent object on active branch | No change: TSW-Q06, TSW-V05(g) stops at eligibility |
| GUESS WCK-R08b derived act mapping | Spec edit: the WCK-R04 threshold mapping |
| GUESS notification/tracking literal arguments | No change: literal values are the stated arguments; meanings are WCK-Q05 |
| GAP intensity values unasserted | No change: WCK-Q01; only the two supplied branches are reproduced |
| GAP WCK-V14 colour bytes depend on supplied request | No change: presentation is supplied by design; only the override is fixed |
| GAP WCK-R07 absent environment | No change: required input |
| GAP TSW-R05a act argument | Spec edit: section 3 retry act argument row |
| GAP TSW-V05/TSW-V10 "supplied remaining work" | No change: stage boundaries are stated in each case |
| GAP WCK-V11 initial game frame | No change: the assertion is the increase of 1000 |
| GAP marker C never set | No change: TSW-Q02 |
| GAP item search region | No change: TSW-Q06 |
| GAP derived act / current-clock flag type | Spec edit: derived act added to section 5; flag width is the World-clock input row's u8 |
| DISCRIM TSW-V01 flag cleared before first intensity | Spec edit: TSW-V01 incoming-flag-1 case |
| DISCRIM start requires client state 4 | Spec edit: TSW-V01 client-state-3 case |
| DISCRIM nonzero-act filter | Spec edit: TSW-V06(d) |
| DISCRIM core flag == 1 | Spec edit: WCK-V08(c) (stored flag 2) |
| DISCRIM WCK-V03 normal compare table | Spec edit: WCK-V08(c) also rejects it |
| DISCRIM WCK-V04 period kept on transition | No change: WCK-V01(b), WCK-V06(b), WCK-V07 reject it |
| DISCRIM TSW-R01e active re-read | No change: equivalent within the supplied-callback contract (FULL10 record) |
| DISCRIM timer applied == 1 | Spec edit: TSW-V09(c) applied 2 |
| DISCRIM schedule state/scheduled == 1 | Spec edit: TSW-V07(i)/(j) |
| DISCRIM altar timer stage test | Spec edit: TSW-V10(d) state 4/stage 0 |
| DISCRIM assertions before eligibility; negative cycle | Spec edit: WCK-V16 ineligible (6,0,0) and eligible (-1,0,0) |
| DISCRIM direct flag value 2 | No change: the contract input is the single direct-player bit; other bits of the original flag word are not inputs |
| DISCRIM reset with flag already 0 | Spec edit: TSW-V02(d) |
| DISCRIM WCK-V15 act 3 | No change: Act-IV intensity is WCK-Q01 |
| GATED | No change: declared gates |
| COVERAGE | Advisory; result-changing branches raised as DISCRIM are covered above |
| HARDEST TSW-V10, TSW-V09, WCK-V12/WCK-V13 | No further change |

Vectors and rule wording (TSW-R01d, TSW-R01g, TSW-R02b, WCK-R01a, WCK-R04, WCK-R08b) changed, and TSW-R01g gained an assertion branch, so the next run is FULL12, the last within this task's cap.

## Fresh reader review: FULL12

On 2026-10-15 a fresh full reader recomputed every case on the revision with the FULL11 edits. Counts: MISMATCH 0, GUESS 4, GAP 9, DISCRIM 9, GATED 22, COVERAGE 14, HARDEST 3. Triage re-checked the D2Game diagnostic format string (`changing state to %d` followed by one line feed; `Finished quest` has none), the altar loop bytes (attempt index zeroed, incremented by 1), D2Common `0x376F0` (the flag-nonzero branch subtracts only above 32, then floors) and the NRF-R01 step for seeds {0,665} and {0,668} (new lo equals the old hi).

| Finding | Disposition |
|---|---|
| GUESS WCK-V14 supplied environment | Spec edit: each WCK-V14 case supplies Act I with the default clock as the ticked environment |
| GUESS newline form | Spec edit: TSW-R04c names one line feed (0x0A), no carriage return |
| GUESS amulet attempt index | Spec edit: index starts at 0 and increases by 1 |
| GUESS recipient act | No change: TSW-R04a already says the act is the recipient client's act |
| GAP TSW-V07(a) | No change: only the no-draw property of the eight bodies is claimed (earlier records) |
| GAP intensity values; colour bytes | No change: WCK-Q01; presentation is supplied. WCK-V14 now fixes the supplied colour result |
| GAP returns of join, animation, event slices and direct retry | Open: section 5 assigns them no return value and no vector asserts one. Callers' use of these results is outside the stage boundary (TSW-Q06). Whether to state "no specified return" or read the returned values is left for the next pass |
| GAP TSW-V05(g) absent object | No change: TSW-Q06 |
| GAP state after a fatal assertion | No change: the fatal path is terminal; only the absence of a normal return is asserted |
| GAP neutral naming | No change: markers and keys are already neutral names or fixed constants without layouts |
| GAP client-list source/order; join/snapshot order | No change: supplied integration (WCK-Q05), stated in the cases |
| DISCRIM argument-free intensity/colour log entries | Spec edit: vector defaults define them as intensity(source 0, act 0) and colour(act 0) in server stage logs |
| DISCRIM WCK-V14 override before colour | Spec edit: supplied colour writes (1,2,3); source 120 ends at (245,240,255), others at (1,2,3) |
| DISCRIM WCK-V16 fixed act 0 in recompute | Spec edit: WCK-V16 source-75 case (act 2) |
| DISCRIM WCK-V15 unconditional subtract | Spec edit: WCK-V15 intensity -2147483648 case |
| DISCRIM WCK-V11 act order / >=16 | No change: WCK-V12's two-act frame case and its 2047/2175 pair reject both |
| DISCRIM unit act vs client act | No change: the contract has only the client act as recipient input |
| DISCRIM WCK-V03 compare table | No change: WCK-V07/WCK-V08 reject it |
| DISCRIM TSW-V07 single reduction sample | Spec edit: TSW-V07(k) results 3 and 0 |
| GATED | No change: declared gates |
| COVERAGE (14) | Advisory; no further vectors |
| HARDEST TSW-V10, TSW-V09, WCK-V12/WCK-V13 | No further change |

## Review status after FULL12

Four reader runs were made within this task (FULL9-FULL12, all full). Every run returned MISMATCH 0. FULL12's edits added vector cases and assertions and clarified TSW-R02b and TSW-R04c. Under the "When to stop" rule a further fresh run (FULL13) is required, and the task cap prevents it here. One FULL12 GAP is still open: the unspecified return values of the join, animation, event and direct-retry stages. Review completion is not claimed.

