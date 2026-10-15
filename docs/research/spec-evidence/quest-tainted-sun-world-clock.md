# Evidence: Tainted Sun world clock

Spec: [Tainted Sun world clock](../../specs/quest-tainted-sun-world-clock.md)
Open questions: [evidence gaps](../../plans/gaps/quest-tainted-sun-world-clock.md#open-evidence-gaps)

## Identity, methods and reuse

The evidence build is original Lord of Destruction Expansion 1.13c, Windows x86. Artifact ids below resolve through the [source manifest](../source-manifest.md): `LOD113C-D2COMMON`, `LOD113C-D2GAME`, `LOD113C-D2CLIENT`, and `LOD113C-FOG`. Addresses below are module-relative RVAs unless explicitly identified as preferred VAs. Descriptions are analyst labels, not original exported names. Original images and game assets are not distributed.

Static readings recorded on 2026-10-12 are reused for the quest chain, direct RNG audit, tables and server frame chain, with their original limits. The source draft's earliest reading had no recorded date or source identity; it remains an undated lead where the dated record did not cover it. Historical loaded DLL hashes were not captured, as the manifest explains. The available original D2Common/D2Game specimens match the published fingerprints; D2Client and Fog current specimen checks establish only those specimens, without retroactively fingerprinting a runtime session.

Targeted analysis on 2026-10-15 checked reconstructions verified against the original bytes first where available, otherwise disassembly, with original-image decoding for ambiguous byte runs and previously uncorroborated instruction slices. No reconstruction or original binary was changed. All source artifacts, full original document bytes, internal paths/revisions, commands, fixture encodings and tool/session history were separately preserved in a private integrity-checked audit. That audit is not required to interpret the public contract. The full original main and gap record predate this companion; historical evidence was not discarded.

The migration retains the quest-plus-clock boundary and all twelve rule families, nineteen existing vector families and four observation families. It adds discriminators for supplied flag updates, partial failure, overflow, client source mapping, intensity branches and assignment. General lighting was not added. No gameplay collection, implementation comparison or implementation work was performed. The earlier question about 1.14d remains untested and outside this evidence build; no cross-build claim is made. Repository references outside sibling specs/evidence/gaps did not consume this contract, so there is no historical-alias column.

## Original function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0xC7480` | — | quest initialization | TSW-R01, TSW-R12a | Dated initialization evidence reused; C view checked against original bytes read 2026-10-15 for installed callbacks, state 0, slot 11 and private state zeroing |
| LOD113C-D2GAME | `0xC6A30` | — | shared start | TSW-R01a, TSW-R12a | 2026-10-12 reading reused; original decode 2026-10-15 checks earlier writes, Act II test, call and return; callback interiors supplied |
| LOD113C-D2GAME | `0xC72C0`, `0xC7130` | — | join selector and apply | TSW-R01d, TSW-R07b | Disassembly 2026-10-12 reused; branch/state-order slices read 2026-10-15 |
| LOD113C-D2GAME | `0xC71A0`, `0xC73A0` | — | animation/event start slices | TSW-R01e, TSW-R01f, TSW-R01i, TSW-R07e | Targeted disassembly 2026-10-15; complete bodies re-read in the follow-up below; mode-update, roster-update and broadcast interiors excluded |
| LOD113C-D2GAME | `0xC7250`, `0xC6AF0` | — | schedule and timer callback | TSW-R01g, TSW-R01h | Dated draw/timer reading reused; targeted guard/return read 2026-10-15 |
| LOD113C-D2GAME | `0x1780`, `0x33AF0` | — | bounded random range and timer insertion | TSW-R01g, TSW-R12b | Disassembly read 2026-10-15; seed step/reduction; due-counter insertion only, not timer delivery |
| LOD113C-D2GAME | `0xC3750` | — | progress-bit writer | TSW-R01b | C view checked against original bytes read 2026-10-15, quest search, guards, exact bit choices, return 0 |
| LOD113C-D2GAME | `0xC6B30` | — | auxiliary quest callback | TSW-R01c | 2026-10-12 partial reading reused; 2026-10-15 instruction view confirms comparisons, not semantic selector identity |
| LOD113C-D2GAME | `0xC76B0` | — | altar operation | TSW-R02a, TSW-R02b, TSW-R02c, TSW-R12b | 2026-10-12 instruction-listing reading reused; original image decoded through `0xC7B3D` on 2026-10-15; callback/drop interiors excluded |
| LOD113C-D2GAME | `0xC5E90` | — | quest environment reset | TSW-R02b, TSW-R04b, TSW-R12a | Disassembly 2026-10-12 reused; 2026-10-15 verifies missing-act check before applied clear |
| LOD113C-D2GAME | `0xC5190`, `0xC5100` | — | start/reset per-player requests | TSW-R03, TSW-R07a, TSW-R12a | Dated reading reused; byte-run comparison at start sender checked in original image 2026-10-15 |
| LOD113C-D2GAME | `0xC5E30` | — | altar timer callback | TSW-R02d | Complete instruction read 2026-10-15 (follow-up); pre-start iteration interior supplied |
| LOD113C-D2GAME | `0xCBE70` | — | diagnostic text broadcast | TSW-R07e | Complete disassembly read 2026-10-15 (follow-up); message buffer built, recipient iteration `0x29820` with callback `0xCB580` supplied |
| LOD113C-D2GAME | `0x8A3E0` | — | per-client send queue | TSW-R07b | Entry instructions read 2026-10-15 (follow-up): absent client returns before any write; queue internals not verified |
| LOD113C-D2GAME | `0xC2150` | — | direct pending retry | TSW-R08a, TSW-R12a | Disassembly 2026-10-15; act argument/pending exact tests and no packet request; outer dispatch excluded |
| LOD113C-D2GAME | `0x9D610` | — | own-act assignment/current-clock snapshot | TSW-R09a | Dated reading reused; 2026-10-15 source view confirms message order; outer join callback scheduling excluded |
| LOD113C-D2GAME | `0x2E050`, `0x2CA20` | — | frame and per-act updates | TSW-R10a, TSW-R10d | 2026-10-12 root reading reused; original-image decode 2026-10-15 checks act loop, before-filter notifications and send request |
| LOD113C-D2COMMON | `0x37CA0` | 10337 | eclipse start assignment | TSW-R04a | Disassembly plus original-image checks 2026-10-15; call to row recomputation is material |
| LOD113C-D2COMMON | `0x37C20` | 10895 | normal reset assignment | TSW-R04b | Original-image check 2026-10-15 resolves byte runs and call/write order |
| LOD113C-D2COMMON | `0x37BA0` | — | fresh allocation/defaults | TSW-R04c | Original-image decode 2026-10-15 resolves zeroing/rate/row writes; allocator/presentation effects supplied |
| LOD113C-D2COMMON | `0x37570` | — | advance/wrap/transition | TSW-R05, TSW-R06a | 2026-10-12 reading reused; complete no-callee core read and original decode 2026-10-15 |
| LOD113C-D2COMMON | `0x37AF0` | 10135 | client source wrapper | TSW-R06b, TSW-R10b | Original-image decode 2026-10-15 resolves byte runs, thresholds, calls, colour override, period-only return |
| LOD113C-D2COMMON | `0x2D2C0` | — | source-to-act helper | TSW-R04, TSW-R10c | C view checked against original bytes and original-image decode supply the same threshold mapping; start/reset source 0 selects act 0; semantic source production excluded |
| LOD113C-D2COMMON | `0x37170` | — | recompute current row | TSW-R04a, TSW-R10c | Complete original-image decode 2026-10-15; no callees; writes period/ticks only |
| LOD113C-D2COMMON | `0x376F0` | — | intensity update | TSW-R07c | Disassembly and original decode 2026-10-15; only eclipse/non-Act-IV and normal-source-120 branches verified; float/Act-IV paths excluded |
| LOD113C-D2COMMON | `0x371C0` | — | colour interpolation request target | TSW-R04, TSW-R06b, TSW-R10c | Instruction view/original byte check 2026-10-15 establishes it writes colour, not rate or ticks; float runtime dependencies/full interpolation remain gated |
| LOD113C-D2COMMON | `0x37920` | 10358 | client set-cycle/clock | TSW-R10c | Complete disassembly/original decode 2026-10-15; clock writes, retained rate and conditional recomputation; presentation supplied |
| LOD113C-D2COMMON | `0x37660` | 11045 | server update/resync predicate | TSW-R10d | Dated reading reused, no-callee core traced through `0x37570`; original decode 2026-10-15 confirms signed arithmetic |
| LOD113C-D2COMMON | `0x373B0` | 11062 | current-clock getter | TSW-R09a, TSW-R10d | C view checked against original bytes read 2026-10-15; copies cycle/ticks/flag from environment |
| LOD113C-D2COMMON | `0x2D0F0`, `0x2D0B0`, `0x2D070` | 10174, 10003, 10593 | quest bit test/set/clear | TSW-R01b, TSW-R01d, TSW-R02a, TSW-R02b | C view checked against original bytes read 2026-10-15, resolved Fog operations below; gameplay names remain neutral |
| LOD113C-FOG | `0x14070`, `0x14090`, `0x14040` | 10119, 10118, 10120 | bit test/set/clear | TSW-R01b, TSW-R01d, TSW-R02a, TSW-R02b | Matched test/set reconstructions and clear disassembly, plus exact mask-data read 2026-10-15 |
| LOD113C-D2COMMON | `0x2EB70` | 10920 | valid-player data getter | TSW-R03, TSW-R07 | C view checked against original bytes read 2026-10-15; valid player type returns unit data; caller resolves client afterward |
| LOD113C-D2COMMON | `0x2FE10` | 10331 | client source getter | TSW-R06b, TSW-R10c | Disassembly read 2026-10-15; player path returns its supplied room-like source; broader source meaning remains TSW-Q07 |
| LOD113C-D2CLIENT | `0x44A80` | — | client update caller | TSW-R10b | Original-image decode 2026-10-15 confirms one wrapper call and refresh; outer cadence unresolved |
| LOD113C-D2CLIENT | `0xACE20`, `0xAE9C0` | — | command dispatcher/world-clock handler | TSW-R10c | Original-image decode 2026-10-15; selected row and handler-to-assignment argument order; later refresh/consumers supplied |

## Dispatch, imports and layouts

### Quest roots and supplied dependencies

The original quest initializer pointer is stored at D2Game data RVA `0x10DCF8`; initialization at `0xC7480` installs join `0xC72C0`, event `0xC73A0`, and auxiliary callback `0xC6B30`. Earlier quest-row evidence also identifies the quest table region at RVA `0x10DBF0`; the initializer-pointer cell, rather than a guessed row number, is the new exact anchor. The join initializer-callback event was recorded as 0xD on 2026-10-12. The animation start slice `0xC71A0` is retained with its dated chain and a supplied outer callback boundary. Its complete upstream animation dispatch is not newly established.

Shared start `0xC6A30` requests player iteration through `0xDFCC0` with pre-start callback `0xC52B0`, then tests the Act II pointer, calls the start import and iterates start sender `0xC5190`. Timer callback `0xC6AF0` and retry `0xC73A0` additionally iterate progress writer `0xC3750`. The pre-start callback's own quest notification calls, iteration implementation, optional diagnostic notification and unrelated callback interiors are supplied integration operations, not omitted verified behavior. The animation slice's final mode setter is D2Common #11090, RVA `0x33920`, reached through D2Game thunk `0xA4A2`; its full effects are TSW-Q10.

Object-operation pointer at D2Game RVA `0x107C18` selects altar `0xC76B0`. It finds quest 10, tests progress at `0xC7713..0xC7779`, tests object mode at `0xC7779..0xC778B`, and on the active path calls quest reset `0xC5E90` at `0xC792E`. Reset tests Act II before clearing applied. Altar flag operations are at `0xC7992..0xC79B9`. The room-item search target `0xE660`, preparation callbacks, item drop `0x11A70`, gold drop `0x6AFA0`, and later quest/iteration operations are supplied boundaries; their names or the presence of the item code do not establish complete item search/drop semantics. Inline seed sites at `0xC7855` and `0xC7AB5` are traced independently of those helpers. Rejection is requested through `0xDFFD0` at `0xC7B2F`. All normal returns decoded through `0xC7B3D` return 0.

| Caller thunk RVA | Imported original target | Rules / purpose |
|---|---|---|
| D2Game `0xAB98` | D2Common #10337, `0x37CA0` | start |
| D2Game `0xAB92` | D2Common #10895, `0x37C20` | reset |
| D2Game `0xAEFE` | D2Common #11062, `0x373B0` | current clock |
| D2Game `0xB2FA` | D2Common #11045, `0x37660` | server update |
| D2Game `0xA4C0` | D2Common #10920, `0x2EB70` | valid player data |
| D2Game `0xA568` | D2Common #10174, `0x2D0F0` | quest bit test |
| D2Game `0xA57A` | D2Common #10003, `0x2D0B0` | quest bit set |
| D2Game `0xA586` | D2Common #10593, `0x2D070` | quest bit clear |
| D2Client `0xCE50` | D2Common #10135, `0x37AF0` | client wrapper |
| D2Client `0xC556` | D2Common #10358, `0x37920` | client clock assignment |
| D2Client `0xC1F0` | D2Common #10331, `0x2FE10` | source getter |
| D2Common `0x92A6` | Fog #10119, `0x14070` | bit read |
| D2Common `0x92B2` | Fog #10118, `0x14090` | bit set |
| D2Common `0x92AC` | Fog #10120, `0x14040` | bit clear |

The bit wrappers form bit index `(slot<<4)+bit`; test normalizes the Fog mask result to 0/1. Fog set masks at RVA `0x26798` are low-byte 1,2,4,8,16,32,64,128; clear masks at `0x26818` are their low-byte complements. Set uses OR, clear uses AND. This establishes actual bit behavior, without assigning gameplay meanings to the slots. Other helpers remain integration gaps.

### Clock records and original constants

For the original layout, Game holds Act pointers at `+0xBC+4*a`; Act II is `+0xC0`. The **environment** is separately reached through Act `+4`. Shared entry points take an Act, not an environment. Environment fields are cycle `+0`, period `+4`, ticks `+8`, intensity `+0x0C`, colour `+0x18..+0x1A`, rate `+0x28`, index `+0x2C`, flag `+0x30`, last resync `+0x34`. Fresh allocation zeroes `0x38` bytes. These layouts locate evidence; they are not required storage designs.

Quest state/stage/transition are bytes `+0x0C/+0x0B/+0x14`; active `+9`; private state pointer `+0x18`, with scheduled/applied/pending at private `+1/+2/+3`. Game quest structure is `+0x10F4`, its seed `+0x18/+0x1C`; game seed `+0xD0/+0xD4`, unit seed `+0x20/+0x24`. A player's client is obtained through player-data `+0x9C`; act byte is client `+0x1AC`. Join's shared flag target is loaded from quest-structure `+0x0C`, after a separate `+8` non-null check. The neutral shared-flag name avoids the earlier draft's ambiguous pointer expression.

Normal/Act-IV/eclipse code tables begin at D2Common data RVAs `0x9B160/0x9B1A8/0x9B1F0`, with twelve-byte rows. Rate table is `0x9B238`; thresholds are `0x94668`. On 2026-10-15 exact original-image reads matched every begin/period/rate/threshold constant in spec section 3. These are engine-defined data, not archive-table cells and not gameplay execution. Colour words were inspected only as locations for the excluded interpolation; they were not turned into new contract values.

Client source access in `0x37AF0` and `0x37920` follows supplied source `+0x10`, then `+0x58`, then reads `+0x1D0`; absent source yields 0. This supports supplied value arithmetic, without treating an unresolved layout label as proof of a Levels row identity. The outer client source/cadence remains TSW-Q07.

### Server frame and client command roots

D2Game frame `0x2E050` increments Game `+0xA8` and calls `0x2CA20` at `0x2E0C4`, before subsequent calls at `0x2E0C9`, `0x2E0D0`, and later frame operations. `0x2CA20` loops five Act pointers; `0x2CA4C` calls thunk `0xB2FA` and, only for nonzero return, `0x2CA6B` reads clock through `0xAEFE`. D2Common `0x37660` calls core `0x37570`, then signed IDIV and low32 absolute-distance logic, retaining the i32-minimum anomaly. Period/cycle comparison does not rewrite resync. The server receives the loop act index, not the client threshold result.

Client resolution at `0x2CA9B..0x2CAD9` is supplied; resolved players reach notifications `0xD51A0` and D2Common #10201, RVA `0x30DC0`, through thunk `0xA9A6` at `0x2CAE6`, then the state/act filter `0x2CAEB..0x2CAFC`. Sending is requested at `0x2CB21` through D2Game `0x8A3E0`. That operation's transmission internals are not verified. Clients are linked at `+0x4A8`; state is `+4`; the pre-filter ordering is material even for recipients later rejected by the act/state checks.

D2Client outer update `0x44A80` calls the wrapper through `0xCE50` at `0x44AFE` and refreshes at `0x44B07` only on nonzero return. Original-image correspondence verifies the call, not full outer cadence. Command dispatcher `0xACE20` indexes twelve-byte command records at data RVA `0xDDE60`; World-clock row 0x53 begins `0xDE244`, handler `0xAE9C0`, length 10. The stream dispatcher calls that first handler slot after validation; the separate queued-command path `0xABC20` uses another callback slot and is not substituted for the stream dispatch. Handler `0xAE9C0` checks current-act/player correspondence, extracts cycle/ticks/flag, obtains source via #10331, and calls #10358. Later refresh/sound/presentation consumers remain supplied.

### Original message encodings

The original senders construct World-clock opcode `0x53`, ten bytes: opcode byte, little-endian cycle dword, little-endian ticks dword, flag byte. Start payload is `{5,0,1}`; reset `{2,0,0}`. Quest-event opcode `0x5D` is six bytes: opcode, quest byte `0x0A`, flags byte `0x10`, three zero bytes. Failed join uses `5D 0A 01 00 00 00`. The flag-1 start request precedes Quest-event for each eligible player. These are original wire evidence, not a new wire-format acceptance requirement in the main contract.

The original client row for `0x53` at `0xDE244` has length 10. Adjacent selected command rows `0x5D`/`0x5E` at `0xDE2BC`/`0xDE2C8` have lengths 6/38. This preserves the earlier size calibration and independently checks the current image's selected cells. The earlier reading associated the quest event with two client sounds; the named client leads are `D2Client+0x1E160` and `0x43830`. Their complete dispatch/audio effects remain TSW-Q10 and are not established by those labels or a table length. No packet capture was performed.

## Static corrections and affected-claim rechecks

### Final start ticks and client retained rate

The earlier draft asserted final start ticks 0 in its clock rule and start/join/reload vectors. It followed the initial store but omitted the downstream recomputation. Original-image checks on 2026-10-15 establish: D2Common `0x37CE5` initially stores 0; `0x37D07` sets flag 1; call `0x37D0E` targets `0x37170`. That complete no-callee helper selects eclipse row 0 at `0x37197`, multiplies rate 4 by begin 300 at `0x371AB`, and writes ticks 1200 at `0x371AE`. The subsequent intensity/colour routines' writes do not change these clock fields. The final server state is therefore ticks 1200, not 0. This is a source-backed correction, with no new runtime claim.

Rechecked assertions are TSW-R04a, shared start/join/direct-retry consumers, TSW-V01/V06/V07/V09/V16/V21, and the observation expectations. The independent server start sender still explicitly requests cycle 5/ticks 0/flag 1, so TSW-R07a's payload is retained. TSW-V13 retains the legitimate supplied-zero tick fixture and separately exercises returned start ticks 1200; both transition to row 1/ticks 0 on the next core call. Other threshold/rate/transition vectors are unchanged in arithmetic.

The client setter `0x37920` writes neither rate nor rate index. At `0x379EB..0x379F4` it writes the new flag and, if nonzero, calls the same recomputation helper, then intensity/colour. With retained client rate 128 and received cycle 5, final ticks are 128*240=30720; packet ticks 0 are superseded. This closes the previously unverified supplied-client-rate question and supplies TSW-V25. The stage with normal flag accepts ticks equal to rate*360 and pre-clamps only strictly greater values (`0x37996..0x3799A`). Full packet delivery/appearance is still gated.

### Altar clears bit 14

The earlier altar rule/vector said bit 14 was set with bits 1/13. Original altar call `0xC79B9` targets D2Game thunk `0xA586`, resolved to D2Common #10593 `0x2D070`, then Fog #10120 `0x14040`. Matched wrapper reading, clear routine AND operation and original mask values on 2026-10-15 prove **clear**, whereas prior calls at `0xC79A5/0xC79AF` use the set operation. Rechecked TSW-R02b, TSW-V08 and outputs now require set 1,set 13,clear 14. The vector begins with bit 14 set so a false set/no-op cannot pass. Flag meanings remain neutral under TSW-Q02.

### Boundaries, historical labels and locator corrections

Mixed table/runtime/code labels were split into one label per leaf. Idle memory values and table correspondence are retained as narrow corroboration; no complete executable vector receives runtime_observed: yes. In particular the exact 1000-frame/no-resync/unchanged-cycle TSW-V17 and drift send TSW-V19 were not captured. No source age was used to renew verification.

Earlier broad negatives about Lost City triggers, saving and other consumers had no six-rung search record. Current TSW-R01c/TSW-R09b/TSW-R11 are unresolved within the actually searched boundaries, rather than claiming global absence. Current event/flag/timer/delivery inputs are neutral and supplied; local arithmetic/writes are not globally gated by unknown producers.

Several earlier locator strings confused assembly label offsets with module RVAs. Import resolution establishes start/reset thunks `0xAB98/0xAB92`, flag thunks `0xA568/0xA57A/0xA586`, timer insertion `0x33AF0`, player iteration `0xDFCC0`, item drop `0x11A70`, and gold drop `0x6AFA0`. Original calls and imports, rather than label text, are the public locators. Original full text/locator strings remain privately preserved. These corrections do not independently verify the excluded helper interiors.

### Quest-stage follow-up (2026-10-15)

An independent static re-derivation of the quest stages against D2Game 1.13c reported omissions in the animation, event, altar, sender and resync slices. The reviewer re-read every changed path on 2026-10-15: the disassembly view where it has no undecoded byte runs, and independent decoding of the original D2Game image (preferred base `0x6FC20000`) for `0xC76B0..0xC7B3D`, `0xC5E30`, `0xC5E90`, `0x2CA9B..0x2CB38` and the entry of `0x8A3E0`. Imports were resolved through the original import tables. No reconstruction or binary was changed, and no gameplay was observed.

Several callee addresses in the incoming report were offset by `0x1000` and are corrected here from the original image: participant iteration `0xC3630` (reported `0xC4630`), post-flag iterations `0xC36B0`/`0xC5ED0`/`0xC5090` (reported `0xC46B0`/`0xC6ED0`/`0xC6090`), completion iteration `0xC5260` (reported `0xC6260`), altar timer callback `0xC5E30` (reported `0xC6E30`) and player lookup `0xE03A0` (reported `0xE13A0`). The report's animation finding named the `0xC7130` path; its instructions are in `0xC71A0`, while `0xC7130` is join apply. `0xC5090` starts after the `int3` at `0xC508F`, although the reconstruction groups it with `0xC5030`. D2Game `0x33920`, called by the altar, is not D2Common #11090, which also sits at RVA `0x33920` in D2Common.

Layout mapping for the new semantic names: private markers M4/M5/M6 are bytes at private `+4/+5/+6`; M8, MC, M10 and M98 are dwords at `+8/+0xC/+0x10/+0x98`; the amulet count is private `+0x9C`. The roster word is the u16 at quest `+0x9C`. Event key is event `+0x18`, auxiliary identity `+0x14` and event unit `+0x0C`; the animation argument's unit is at `+4`. Unit identity is unit `+0x0C`. The object item-code field is object `+0xB8`. Quest 9's counter is its private `+0x24`. Client direct-player flag is bit 0 of client `+0x3D4`, cached player reference `+0x174`, lookup keys `+0x16C/+0x170/+0x1A8`. The diagnostic switch is D2Game data RVA `0x111C20`.

| Rule | Evidence (D2Game RVAs) |
|---|---|
| TSW-R01e | `0xC71AE` writes M5=1; `0xC71B1..0xC71C0` writes MC from argument unit `+0x0C` or -1. `0xC71C3` active==1, else jump to `0xC7211`. `0xC71C8` unsigned state>1 skips the start at `0xC71CD`; results write pending/applied, then state 3/transition 0 at `0xC71DE/0xC71E2`. `0xC71E6` stage==0 writes transition 0/stage 2 and iterates `0xDFCC0` with `0xC52B0` at `0xC720B`. `0xC721D` calls thunk `0xA4A2` (D2Common #11090) with the argument unit and M8 on every path. The constant null test on the callback at `0xC71F2` makes the fatal branch `0xC7228` unreachable |
| TSW-R01f, TSW-R01i | `0xC73B2..0xC73D6`: key 0x2C/0x2D, state 0, active 1 call schedule `0xC7250`. `0xC73DB..0xC7404`: key 0x28, pending 1, shared start, progress iteration `0xC3750`, pending clear; both fall through. `0xC7408` tests auxiliary identity 0x28; `0xC740F` u16 roster word >0 calls D2Game `0x33830` at `0xC742B` with quest `+0x1C` and the event unit's identity or -1. `0xC7430` state 2 writes 3 at `0xC7442`, formats with USER32 `wsprintfA` (slot `0x6FD18D58`) and calls `0xCBE70` at `0xC7462` only when the switch is nonzero |
| TSW-R01a, TSW-R07e | In shared start, the state-0 branch writes 1 at `0xC6A6D`, formats at `0xC6A71`, and calls `0xCBE70` at `0xC6A8D` when the switch is nonzero, before the Act II test. Format string at .rdata RVA `0xFCC2C` is `changing state to %d` plus newline; the altar passes `Finished quest` at RVA `0xFB2E0` directly. `0xCBE70` builds a message with opcode `0x26` and the copied text and calls `0x29820` with callback `0xCB580`; that broadcast and its recipients are supplied |
| TSW-R07a | Both `0xC5190` and `0xC5100` test the unit for null, type 0, and player-data `+0x9C` client; any failure reaches the assertion call (line `0x7CB`) and `-1` exit before the act test |
| TSW-R07b | `0xC7145..0xC7173`: the client is read from player data only for a present type-0 unit, else 0. `0x8A3E0` tests the client at `0x8A3EB` and returns at `0x8A48C` without writes when it is absent |
| TSW-R02a..c | `0xC76C3..0xC76FB` missing quest returns 0. Active (`0xC78E5`): `#10331` (thunk `0xA478`) to M10, `#10826` (thunk `0xA556`) to M98, `#11090`(O,1), M8=2, M5=1, MC, M4=1 (`0xC78E5..0xC7920`); active retest `0xC7923`, reset `0xC792E`, broadcast `0xC7947`, `0x35EB0`(4) `0xC7953`, `0x35E90`(6) `0xC795D`; amulet count 0 and `0xC3630` iteration `0xC7966..0xC797E`; active retest `0xC798A`, flag test/set/set/clear `0xC7997..0xC79B9`, iterations `0xC36B0/0xC5ED0/0xC5090` `0xC79CC..0xC79F2`, D2Game `0x33920`(quest `+0xE0`,13) `0xC7A07`; `#10826`(M10), `+0xB8`=`vip `, drop loop `0x11A70`(O,7,...) with `#10977`(item,16,1) per success `0xC7A0C..0xC7A6F`; `0xC2380`(successes) `0xC7A76`; `0x6B570`(4) `0xC7A7E`; missing game quest structure fatal `0xC7A90`; draw `0xC7AB5`; gold `0x6AFA0` loop; `0xC5260` iteration `0xC7B06`; `0x33AF0`(quest, `0xC5E30`, 10) `0xC7B1A`; return 0. Inactive (`0xC779D`): `#11090`(O,1), M4=1, M8=2, amulet count 0, `0xC3630` iteration, `#10331`/`#10826` unstored, `+0xB8`=`vip `, counted drop loop, `0x6B570`(4), draw `0xC7855` (missing structure jumps to fatal `0xC7A90`), gold loop, quest-9 search and `+0x24` addition at `0xC78CC..0xC78DB`; both exits return 0 |
| TSW-R02d | `0xC5E30`: state 4 writes transition 0/stage 3 and iterates `0xC52B0`; `0xC5E60` clears private `+6`; returns 1 |
| TSW-R10d | `0x2CA9B` tests client `+0x3D4` bit 0 and takes `+0x174` directly; else `0x2CAAC` skips a zero `+0x174`, calls `0xE03A0` at `0x2CAC6` and clears `+0x174` at `0x2CACF` on a zero result; `0x2CAD7` skips an absent player |

Discrepancies corrected: TSW-R01e previously treated the marker writes, stage-0 branch and final mode-update as supplied bookkeeping, and TSW-V21(i)'s active-1/state-2 case asserted every default field retained with no request. Instructions `0xC71E6..0xC720B` show that a zero stage becomes 2 with transition 0 and a pre-start request; the case now asserts that. TSW-R01f described the event's remaining effects as supplied; the auxiliary-identity slice is now TSW-R01i. The single altar rule became TSW-R02a..d with the ordered requests above in place of "supplied preparation/intervening callbacks". TSW-R07a gained the sender fatal path, TSW-R07b the no-client behavior, TSW-R10d the resolution steps, and TSW-R07e the diagnostic broadcast. TSW-Q11 records the cycle-0 server start versus the fixed cycle-5 message. Vector arithmetic of unchanged cases was not affected. The affected cases are TSW-V01 (switch case), V07(g), V09(b), V19 (resolution case), V21(a)/(i)/(m) and new V26.

## Historical partial observations

| Vector / assertion | Source profile | Date | Module + RVA | Records | Result / limits |
|---|---|---|---|---|---|
| constants in section 3 | Reported vanilla LoD 1.13c, idle single-player Normal, level-1 character in Act I town | 2026-10-12 | D2Common `0x9B160`, `0x9B1A8`, `0x9B1F0`, `0x9B238`, `0x94668` | all eighteen begin/period pairs, rate and threshold values; total memory-read call count unrecorded | All values matched. Code data correspondence only, no gameplay branch confirmation |
| fresh client clock fields | same profile | 2026-10-12 | client Act at D2Client data `0x11C3B8`, environment reached through Act+4 | exact read count unrecorded | cycle 2/period 0/rate 128/index 0/flag 0 matched defaults; ticks already advanced, so no observed allocation tick 0 |
| Act I normal day progress | same profile | 2026-10-12 | D2Common environment ticks; D2Client wrapper import slot `0xCE144` | paired observations, exact aggregate call count unrecorded | Client delta 300 in 11.95 seconds and delta 2545 in 101.8 seconds, approximately 25 ticks/s; import target `D2Common+0x37AF0`. No counted 1000-call scenario, transition, night/other-act/eclipse case or client resync captured |
| server normal day progress | same profile; server module relocated | 2026-10-12 | D2Game game-table global `0x111C24`, Game+0xA8 frame/Act pointers; import slot `0xF80C0` | ticks 4208,4371,4920; frame 4376 at second sample and 4922 at third; exact read count unrecorded | Server ticks advanced; target `D2Common+0x37660`. Samples not atomic; pairwise offsets 5/2 do not establish exact synchronized frame deltas |
| server last-resync fields | same profile | 2026-10-12 | environment+0x34 | values 0x11 then 0x22, record count unrecorded | Values 17/34 observed; no return-value/send trace. Do not promote TSW-R10d or V19's send assertion |

Historical loaded bases were D2Common `0x6FD50000`, D2Client `0x6FAB0000`, with relocated D2Game; public RVAs remove relocation dependence. Full loaded DLL hashes/data overrides were not recorded. The runtime record used read-only memory reads, with no writes and no instruction traces. No cycle change occurred during the client readings. Not observed: eclipse boundaries, darkness, exact client rate after a message, night/Act III/Act IV/flag-1 updates, clock wrap/transition, or the World-clock send. Table cells/live field values alone do not confirm those rules.

## Search and gap assessment

On 2026-10-15 the dated records above were checked before targeted reading. Quest callback installation, output writes, frame roots, current-clock imports and message-handler rows were compatible evidence anchors; new reads addressed specific missing downstream calls, flag imports and byte decoding. No age-only reread or new gameplay campaign was used.

- TSW-Q01: existing quest callbacks/auxiliary comparisons were reused; targeted initialization and event/animation reading confirms bounded event selection. Selector producers and complete entry triggers remain untraced. No six-rung absence claim.
- TSW-Q02: matched flag wrappers plus Fog operations/masks close exact test/set/clear effects, including bit 14 correction. Gameplay meanings still need consumers.
- TSW-Q03: intensity's two requested non-Act-IV branches were read from entry branch to write. Other float/Act-IV paths and rendered effects are excluded and unresolved; colour applier inspection establishes no clock/rate writes, not a complete general-lighting contract.
- TSW-Q04: original command row, lengths, send construction and supplied client assignment were checked. Packet capture remains pending.
- TSW-Q05/TSW-Q06: existing bounded join/getter evidence does not cover all save/consumer paths. No complete six-rung search or runtime save campaign exists; both stay unresolved rather than repeating identical unsuccessful searches.
- TSW-Q07: original-image decoding closes supplied source mapping and period-only return. Source getter/caller read does not establish every producer or outer scheduling root; cadence/source semantics remain open.
- TSW-Q08: dated eight-body direct audit and two draw sites reused; targeted seed-site decoding adds explicit boundaries. Item/gold/callback interior reachability remains open. No whole-path no-draw claim.
- TSW-Q09: original server predicate/send caller confirms static request/filter behavior. Historical idle reads did not capture its send or counted-frame vectors; observation stays pending.
- TSW-Q10: targeted reads expose concrete dependency crossings (timer insertion/delivery, pre-start/status/altar callbacks, player resolution, pre-filter notifications, mode setters, drop operations, presentation and transmission). Their inputs/results/order are supplied in the contract. Their interiors were not silently adopted as verified behavior.

## Observation procedures

These procedures are pending, not records of completed gameplay. Use an owned vanilla LoD 1.13c Expansion profile with matching binaries; record executable/DLL identity, effective data/overrides, difficulty, class/level, state/flags and exact counted frames. Collect entry/return values and the output writer, recording capture capacity, dropped records and relocation. Use a disposable copied character when setup needs progress changes. No memory-call result by itself establishes a normal gameplay branch.

- **TSW-O01:** Normal-difficulty character able to reach the Act II quest, without assuming a class/level restriction; record actual class/level/items/progress. Observe quest/event inputs, shared start return, final server clock and both outgoing messages, then client assignment. Observe eligible altar operation with mode/flag/item-test results; compare another act in the same game. Read immediately around every boundary and retain all callbacks/drop limits. Expected local tuples are V01/V02; appearance/event identity are questions, not fabricated expected captures.
- **TSW-O02:** Record flags before joins for fresh, started and completed progress and each tested difficulty. Capture join selection, missing-act failure if naturally reached, timer/retry order and own-act snapshot. Compare the supplied-order cases V06/V07/V16; do not infer save-file encoding from a visual match. Save/load tests require their own normal file-path/encoding evidence.
- **TSW-O03:** Two players with recorded current acts/progress; capture the second joiner's snapshot and either player's eligible altar operation. Record every recipient and ordered World-clock/Quest-event request plus actual client handling. Distinguish stage calls from successful network delivery and sound playback.
- **TSW-O04:** Record server clocks before leaving Act II, each counted frame, and after return. Hold other-act control inputs constant. Trace server update return and request at D2Game `0x2CB21`, with state/act/player eligibility; use V19 equality/exceeded thresholds and V22 wrap as separate cases. Capture cycle/period changes independently. The historical idle samples cannot substitute for this counted observation.

## Handoff checks

The main contract has exact local vectors, explicit supplied boundaries and draft status. New source-backed corrections close final-start ticks, client-rate preservation/recomputation, bit 14 clear, supplied threshold arithmetic and bounded intensity arithmetic. Observations, producer identities, complete save/consumer searches and helper integrations remain in the linked gap record. Historical implementation-gap entries are retained as dated records, without a new comparison.

Fresh reader FULL1, FULL2 and FULL3 and their dispositions are recorded below. A fresh applicable DELTA4 check is pending for the amended vectors; the governing rules and gates are unchanged. User approval and publication acceptance are not inferred from this evidence record.

## Fresh reader review: FULL1

On 2026-10-15 a fresh full reader checked all twenty-five executable vector families and their listed cases from rules/inputs, ignoring traces, without source or gameplay access. No subcase count was reported. Counts were MISMATCH 0, GUESS 0, GAP 4, DISCRIM 1, GATED 26, COVERAGE 12 reported in eight rendered topic groups, and HARDEST 3. The reported coverage count and grouped rendering are retained without inventing a different count. This full pass does not itself approve the handoff.

| Finding | Disposition |
|---|---|
| GAP: shared-start return absent from Outputs | Added TSW-R01a return 0/1 to section 5; V01 and failed V07 preserve the existing assertions |
| GAP: progress-writer return absent from Outputs | Added TSW-R01b return 0 to section 5, matching V20 |
| GAP: altar return absent from Outputs | Added the altar return 0 (TSW-R02a through TSW-R02c) to section 5, matching bounded V08 cases |
| GAP: client-wrapper return absent from Outputs | Added TSW-R06b period-change return 0/1 to section 5; V23 has both changed-period return 1 and unchanged-period return 0 |
| DISCRIM: A's clock request could be deferred until after B/C pre-filter pairs | V19 now requires the exact recipient-stage log: A notification, A tracking, A clock, B notification, B tracking, C notification, C tracking. It explicitly rejects batching the send after B/C |
| COVERAGE group: progress writer's bit-0 early return | Added fresh V20 state-3/flags-{0} case, no writes and return 0 |
| COVERAGE group: pending-event failure and 0x2C/0x2D schedule | Added V21(f) retained pending/earlier writes on missing-act failure, and V21(g) exact seeded schedule for each supplied event key |
| COVERAGE group: nonzero state with scheduled zero | Added V18(f), unchanged seed/scheduled and no request |
| COVERAGE group: timer success | Added V21(h), exact final applied/scheduled/pending, progress flags, return and request order |
| COVERAGE group: altar slot-10 rejection, staff presence, missing object, inactive work/draw | Added V08(e)/(f) for the two short-circuit results; V08(g) stops exactly after eligibility continuation for missing object, retaining the downstream TSW-Q10 gate; V18(g) supplies inactive context/helper effects and proves no reset plus one inline quest draw/six gold-pile requests. Full helper/drop effects remain gated |
| COVERAGE group: Act-IV and other normal intensity | No new formula or vector: those branches are already excluded from TSW-R07c and explicitly gated by TSW-Q03. Expanding general lighting would cross the current contract boundary |
| COVERAGE group: multiple-act frame order/index | Added separate V19 two-act frame with ticks 2176/2191 and exact act-0-before-act-3 message/notification order, distinguishing loop index from a client source |
| COVERAGE group: rate zero and cycle/period notification without drift | Added V22(d) synthetic zero-rate division skip, and V22(e) cycle change with retained period and zero drift. A standalone period-only change is unreachable from the valid-cycle core because its only period write accompanies a cycle transition; the original return predicate still checks both fields |
| HARDEST: low32 overflow, wrap before transition | No rule change: section 3, TSW-R05 and V22(a)/(b) already give widths/order/exact outputs. The added zero-rate case makes division skip explicit |
| HARDEST: signed drift, threshold and per-client order | Existing V22(c) distinguishes i32-minimum negation; V19 equality/exceeded cases retain strict threshold; the exact log and new cycle-only notification case resolve the ordering ambiguity |
| HARDEST: join failure, flag precedence, state/snapshot order | No new behavior: V07(e) already requires bit-4 precedence and retained failure writes; V06/V16 explicitly supply the quest-join-before-snapshot schedule. Full outer order remains TSW-Q10 |

The twenty-six GATED entries were ten open questions, six unresolved/asm-only leaves and ten supplied-stage vector families. They remain declared gates, rather than missing local values. The coverage additions use existing rules and arithmetic; this pass collected no new source or runtime evidence. Behavioral assertions/vector coverage changed, so a fresh full reader is required before review completion can be claimed.


## Fresh reader review: FULL2

On 2026-10-15 the next fresh full reader returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 5, GATED 20, COVERAGE 8 and HARDEST 3. Its required findings concern the state a supplied operation observes at entry. The revised vectors explicitly define call-time snapshots of semantic state; these are observations in the vector log, not invented function arguments or additional stored fields.

| Finding | Disposition |
|---|---|
| DISCRIM: V01 could set eclipse flag before first intensity | Added first-intensity snapshot flag 0 with the initial zero ticks and second-intensity snapshot flag 1 with recomputed ticks 1200 |
| DISCRIM: V02 could clear flag before intensity | Added intensity snapshot flag 1 after reset fields/applied clear and before flag clear; final flag remains 0 |
| DISCRIM: V21(b) could clear pending before progress callback | Progress-writer entry snapshot requires pending 1/applied 0/state 1/stage 1, followed by the pending-clear write |
| DISCRIM: V21(h) could set applied after progress callback | Progress-writer entry snapshot requires applied 1/pending 0/scheduled 0/state 1/stage 1 |
| DISCRIM: V25 could store new flag before first intensity | First-intensity snapshot requires old flag 0 and supplied ticks 0; second requires flag 1 and recomputed ticks 30720 |
| COVERAGE: altar bit-0-only and absent quest | Added V08(h) primary bit-1 false/bit-0 true/slot-10 true rejection; V08(i) absent quest returns 0 without quest/flag work. No private-state input is assigned to an absent quest |
| COVERAGE: animation predicate false | V21(i) supplies fresh active 0/2 at state 0 and active 1 at state 2; each leaves the start slice unchanged |
| COVERAGE: unrelated event key | V21(j) key 0x29 retains pending/clock/state with no start/schedule/draw from the slice |
| COVERAGE: Act-IV intensity | No change: this branch is already explicitly gated under TSW-Q03 and excluded from the two verified intensity branches |
| COVERAGE: invalid direct retry | V21(k) distinguishes wrong act argument, absent quest, and pending 2; no presentation or quest-start request. An absent quest has no private-state input/output |
| COVERAGE: invalid client cycle, negative ticks, missing environment | No invented fallback: TSW-R10c's supplied-input domain requires a valid cycle/nonnegative ticks/environment. Those invalid/assertion paths are outside its normal-result boundary and remain subject to TSW-Q10 if integrated behavior is requested |
| COVERAGE: existing server act without environment | V22(f) increments the normal frame, returns 0 from the act update, retains absence and makes no clock/notification request |
| COVERAGE: other altar remainder counts | No exhaustive matrix added: TSW-R12b consumes the same NRF-R01/NRF-R02 step and modulo-5 recipe; V18(e)/(g) supply exact seed/result/count and local context. There is no additional conditional branch by remainder |
| HARDEST: V19 interleaved acts/recipients | No rule change: its exact logs now distinguish per-client interleaving and act-index traversal; these assertions were retained |
| HARDEST: V21 retained writes/call-time state | Added explicit pending/applied snapshots and focused guard/failure cases; earlier failure writes remain exact |
| HARDEST: V22 overflow/resync | No rule change: low32 negation, wrap, strict threshold, zero-rate skip and cycle-only notification each retain exact fixtures |

The twenty GATED entries remain explicit integration, producer, observation or valid-input-boundary limits. No source/gameplay collection occurred in this response to review. The amended assertions/cases require a fresh full pass; this record does not claim review completion or user approval.


## Additional focused review after FULL2

An additional focused reader checked the amended V01/V02/V21(b..h)/V25 snapshots and added V08(h..i), V21(i..k), V22(f), with their shared defaults and governing rules. It returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 2, COVERAGE 0 and HARDEST 3. Its gates were later altar drop/helper effects (TSW-Q08/TSW-Q10) and other animation/event work (TSW-Q10); its hardest entries were the absent-quest altar, timer/progress snapshot and existing-act/absent-environment cases. These declared boundaries needed no new mechanics. This was focused evidence and did **not** replace the third required full reader.

## Fresh reader review: FULL3

On 2026-10-15 the third fresh full reader recomputed all twenty-five vector families and listed subcases. It returned MISMATCH 0, GUESS 0, GAP 0, DISCRIM 2, GATED 16, COVERAGE 6 and HARDEST 3. No subcase count was reported. The sixteen gates were the ten open questions and six unresolved/asm-only leaves.

| Finding | Disposition |
|---|---|
| DISCRIM: V22(d) zero-rate/ticks-0 fixture also passed an incorrect entire-update skip | Changed the supplied starting ticks to -2. The core must advance to -1, remain below both zero thresholds, skip division, retain last resync 17 and return 0. A full skip yields -2; a forced wrap yields 0. The former ticks-0/expected-0 fixture was arithmetically valid but did not discriminate those alternatives |
| DISCRIM: V23 could reverse intensity/colour or omit established request inputs | Every supplied source case now requires intensity with that source first, colour with its derived act second, and explicit post-core clock snapshots at both entries. Source 120 gives intensity(source 120) then colour(act 4), before the final colour override. The two additional period-return cases have their exact ordered requests/snapshots too |
| COVERAGE: shared start with nonzero state | Added V01 fresh state-2 case retaining state 2 while producing the same start clock, marker writes and ordered requests |
| COVERAGE: bit-1 progress guard | No change: V20 already supplies state 3/flags {1}, returns 0 and writes no flags |
| COVERAGE: join bit 3 without bit 4 | Added V07(f), flags {2,3}, exact stage/state 2/2, applied 1 and start clock |
| COVERAGE: schedule event active other than exactly 1 | Added V21(l), each supplied event key with active 2, unchanged quest seed/scheduled/clock and no request |
| COVERAGE: altar active nonzero other than 1 | Added V08(j), active 2, the same bounded reset/flag/request result as active 1, excluding later helper/drop effects |
| COVERAGE: other modulo-5 remainder values | No additional matrix: the unchanged NRF-R01/NRF-R02 reduction recipe and TSW-R12b formula have no remainder-specific branch; exact seed/result/count fixtures already distinguish their owner/order/budget |
| HARDEST: low32 wrap/signed comparisons | No rule/gate change; V22 retains explicit overflow, signed-negation, wrap and threshold outputs, with a stronger zero-rate discriminator |
| HARDEST: drift and recipient interleaving | No rule/gate change; V19 retains its exact per-client/per-act logs and V22 its distance/cycle-only cases |
| HARDEST: join failure/precedence/snapshot order | Existing V07(e), V06/V16 supplied scheduling and failure-state assertions remain exact; V07(f) adds the distinct bit-3-only selection |

Only vectors and their review record changed in this pass. No rule behavior, ownership, acceptance boundary or gate changed, and no new source/gameplay evidence was collected. After three full passes, the next applicable fresh review may be a delta check of the amended vectors and the findings above. Review completion/user approval remain unclaimed pending that check.

## Fresh delta review 4 and review completion

After the three full reviews, the fresh delta reader checked every amended case listed after FULL3: V01 nonzero state, V07(f), V08(j), V21(l), V22(d) and all V23 request/snapshot assertions, with their shared defaults and governing rules. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 3, COVERAGE 0, HARDEST 3. It independently rejected skipping the zero-rate update, reversed presentation calls, wrong source/act arguments and incorrect flag/state predicates. The earlier extra focused review remains separately recorded and is not counted as a full review.

The three reviewed gates receive no change: Q03 retains actual presentation effects, Q07 retains source production/outer client scheduling and Q10 retains supplied integration/delivery effects. All three HARDEST entries receive no change: V23 explicitly maps each source, updates the core and orders intensity(source) before colour(derived act), with exact entry snapshots; V22(d) advances -2 to -1 before skipping division; V07(f) preserves flags while selecting stage/state 2/2. No new rule, ownership, acceptance or gate change was made in this final triage.

Root review read the complete contract, companion and gap record, all additions and every disposition; it independently checked the original start/recompute/colour images and the relevant threshold, seed and drift arithmetic. Every blocking reader finding is resolved on the resulting contract. Document review is complete, with draft status, exact integration/producer/observation gaps and maintainer approval remaining. No new gameplay, implementation test or implementation comparison ran. Publication repeats document checks on committed bytes.

## Review status after the quest-stage follow-up

The 2026-10-15 quest-stage follow-up changed rule behavior (TSW-R01e, TSW-R01f, TSW-R01i, TSW-R02a..d, TSW-R07a, TSW-R07b, TSW-R07e, TSW-R10d), corrected TSW-V21(i), and added vector cases and TSW-V26. The review completion above applies to the earlier contract only. A fresh blind-implementer pass over the changed rules and vectors is pending; no reader has checked this revision.
