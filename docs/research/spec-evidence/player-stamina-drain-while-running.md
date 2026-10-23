# Evidence: player stamina drain while running

Behavioral contract: [Player stamina drain while running](../../specs/player-stamina-drain-while-running.md). Remaining assertions: [gap file](../../plans/gaps/player-stamina-drain-while-running.md).

## Original artifacts and methods

Evidence build is Windows x86 Lord of Destruction Expansion 1.13c. Artifact identities are in the [source manifest](../source-manifest.md): `LOD113C-D2GAME`, `LOD113C-D2COMMON`, `LOD113C-FOG`, `LOD113C-STORM` and `LOD113C-GAME`. Original D2Game/D2Common analysis images were rehashed on 2026-10-15 and exactly matched manifest SHA-256/size. This establishes current instruction applicability, not historical loaded-DLL or archive identity. Descriptions below are analyst labels, not original export names.

Compatible readings from 2026-10-12 are reused for the computing path, table descriptors, frame caller and mode requests. Shared NRF/GDF evidence dated 2026-10-14 and current STL/OWD/MMF/DUR/MCI/MHD/MMC records dated 2026-10-15 were assessed before collecting new evidence. Reuse retains its original date and assertion limits. Getter evidence does not close writer integration; frame-event allocation is distinct from Common unit-timer cancellation. Private reconstruction history, complete original document bytes and binary fixture encodings are preserved separately with integrity recorded privately.

New readings first checked whether a matching reconstruction existed; matching reconstruction was used for Common `0x3A710` and `0x719A0`. Other new applicable paths used disassembly. Encoded instruction runs were independently decoded from the fingerprinted original images, including full extents of Game `0x77BB0` (256 bytes), `0x79A10` (80), `0x79B90` (640), `0x8F070` (208), `0x8ECC0` (272), `0xE0750` (80), `0x8E6F0` (288), `0x78610` (256), `0xE0B40` (144), Common `0x1E1B0` (352), `0x3A740` (64), and `0x38550` (240). Embedded tables at function tails were treated as data, not as disassembled instructions. These checks support the stated local decisions/requests, not unread downstream operations.

## Function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0x2E050` | none | Server frame caller | SDR-R01a | Compatible 2026-10-12 frame/output anchor; call at `0x2E0D5` reaches frame pump |
| LOD113C-D2GAME | `0x8F070`, `0x8ECC0` | none | Frame pump / player-group dispatcher | SDR-R01b/SDR-R01c | Disassembly plus full original-image decoding 2026-10-15; supplied traversal/callback boundary |
| LOD113C-D2GAME | `0x8E6F0` | none | Frame-event removal | SDR-R01b/SDR-Q01 | Full disassembly/image 2026-10-15; deferred running removal, saved-successor update and unlink/free-chain writes |
| LOD113C-D2GAME | `0xE0750`, `0xE0B40` | none | Pending removal / persistent registration | SDR-R01d/SDR-Q01 | Full directed disassembly/image 2026-10-15; local type scans/requests/registration fields; full population invariant not established |
| LOD113C-D2GAME | `0x8E680`, `0x8E540` | none | Frame-event association/allocation | SDR-R01d/SDR-Q01 | Reuse MIR-R01b and STL-R01 dated 2026-10-15; association is not whole recurring lifecycle |
| LOD113C-D2GAME | `0x79A10` | none | Player-mode dispatcher | SDR-R01e | Disassembly/image 2026-10-15; supported table domain and result-2 mode-end request |
| LOD113C-D2GAME | `0x79B90` | none | Movement-mode consumer | SDR-R01f | Compatible 2026-10-12 caller; full source/image examined 2026-10-15; Run gate/Walk request/path continuation slice only |
| LOD113C-D2GAME | `0x77BB0` | none | Running-stamina computation | SDR-R02 through SDR-R10 | Compatible 2026-10-12 reading; full source/image checked 2026-10-15 including ordered mutation/read requests |
| LOD113C-D2GAME | `0x78610` | none | Player-mode transition | SDR-R11/SDR-Q04 | Full local caller and jump-table decoding 2026-10-15; full dependency effects remain gated |
| LOD113C-D2COMMON | `0x2FE10`, `0x3C390`, `0x2D320` | 10331, 10057, none | Room getter / town wrapper / classifier | SDR-R02 | Compatible PVP/MSD/NRF room records; source reread and all 109 town table bytes checked 2026-10-15 |
| LOD113C-D2COMMON | `0x1E670`, `0x1E1B0` | 11139, none | Body-slot getter / lazy page | SDR-R05b/SDR-R05c | Source and full page original-image decode 2026-10-15; existing-page and supplied-allocation-success slices |
| LOD113C-D2COMMON | `0x719A0` | 10695 | Items-record lookup | SDR-R06 | Matching reconstruction read 2026-10-15; unsigned bound, fatal absent collection, computed record |
| LOD113C-D2COMMON | `0x787C0` | none | Items field descriptors | Inputs/SDR-R06 | Compatible 2026-10-12 speed descriptor reading; field name/offset, not complete compiler |
| LOD113C-D2COMMON | `0x39430`, `0x38B70`, `0x38A80` | 10910, 10973, none | Unit stat-read entries / selected-array reader | SDR-R07/SDR-R10a | Compatible 2026-10-12 wrappers plus RLI-R02a/RLI-R02b 2026-10-14 getter contract; no aggregation/propagation |
| LOD113C-D2COMMON | `0x3A710`, `0x3A330` | 10551, none | Add request / additive list writer | SDR-R09/SDR-Q02 | Matching wrapper first, source and original-image target check 2026-10-15; caller target corrected; full writer propagation remains gated |
| LOD113C-D2COMMON | `0x3A740`, `0x3A280`, `0x38550` | 10887, 10261, none | Replace request / inner writer / changed-stat tracking | SDR-R10b/SDR-Q02/SDR-Q03 | Source and wrapper/tracking image decode 2026-10-15; original PE export 10261 independently resolved to 0x3A280; actual entry mutation/propagation/delivery not claimed |
| LOD113C-D2COMMON | `0x303E0`, `0x34C60`, `0x36230`, `0x33920`, `0x35210` | 11103, 10800, 10334, 11090, 10498 | Skill-context reset / continuation read / planner / mode setter / path preparation | SDR-R11/SDR-Q04 | Compatible reconstruction and RLI-R04a/MCI planner limits; fresh import resolution 2026-10-15; incomplete full transition |
| LOD113C-D2GAME | `0x782F0`, `0xDA750`, `0x79850`, `0xE07A0`, `0xE1930`, `0x439A0`, `0x78430`, `0xD01D0` | none | Stop, world/mode-end, preparation/path/optional state/arrival dependencies | SDR-Q04/SDR-Q07 | Directed caller/source assessment; interiors not established as inert; full effects remain gated |
| LOD113C-D2COMMON | `0x2EB70` | 10920 | Client-record getter | SDR-Q07 | Fresh import resolution; movement history remains outside local drain |
| LOD113C-FOG | `0x1CD50`, `0x1CDA0` | 10045, 10047 | Allocate / reallocate | SDR-R05c/SDR-Q07 | Common allocation thunks and import/export tables resolved; supplied successful results only; allocator internals not promoted |

## Dispatch, requests and semantic mapping

Root chain: Game `0x2E050 → 0x8F070 → 0x8ECC0 → type table 0xFBB74[0] → 0x79A10 → mode table 0x1079D0[3] → 0x79B90 → 0x77BB0`. Drain call is at `0x79BC4`; false-result Walk request to `0x78610` is at `0x79BD5`; path step `0xE1930` follows at `0x79BDE`. This output anchor and both dispatch tables identify the routine independently of misleading labels.

Pump calls player persistent pass at `0x8F0B6`, with due-mode selector zero, then due bucket at `0x8F0C7`, selector one. There are **five** persistent group heads, manager `+0xA04..+0xA14`, and five arrays of 64 buckets, not four. Player uses group 0; other pump targets are `0x8E9A0`, `0x8EBC0`, `0x8EF20`, `0x8EDD0`. Their full consumer behavior is outside this spec. The pump's selected current index uses signed remainder 64; fixture frames are nonnegative. Group mapping at Game `0xFA3DC` gives player 0, monster 1, object 3, missile 2, item 4, tile -1, absent 0. The complete byte ranges of all 15 event slots, 20 mode slots and seven group entries were read from the matching image on 2026-10-15. Event slot 1 points to mode-end processing `0x79850`; event slot 2 is null. Mode slots 2/3/6/19 point to `0x79B90`, mode 0 to `0x77940`; all twenty supported mode entries are nonnull. Mode indexing itself has no guard.

Persistent dispatch sets running before the call, clears it after, removes only pending-removal entries and resumes the manager's saved successor. Due dispatch requires exact frame equality then removes the entry after callback. The removal routine `0x8E6F0` has no calls: running entries are merely marked pending-removal and returned; actual removal updates saved successor if it pointed to the removed entry, unlinks manager and unit chains and puts the entry on the event free chain. This closes the specific unread-free-routine concern. It does not close arbitrary callback mutation, duplicate population or initial player registration. Common `0x37110` unit-timer cancellation is a different subsystem and supplies no proof here.

Pending removal `0xE0750` first scans type 0, then type 1, saving each unit-chain successor before calling `0x8E6F0`. Persistent registration `0xE0B40` scans/removes type-0 entries, calls `0x8E680`, initializes type 0/frame -1/no callback/parameters 0, sets its persistent condition, and inserts the returned entry at the selected group head. MIR-R01b establishes target/identity and unit-chain association supplied by `0x8E680`; those writes are not inferred from their absence in the caller. Complete arbitrary nonfresh/reentrant lifecycle stays SDR-Q01.

Town path is Common #10331/`0x2FE10 → #10057/0x3C390 → 0x2D320`. For a player, path `+0x1C` supplies room. Present room traverses `+0x10 → +0x58 → +0x1D0` without null guards. Classifier decrements id and compares unsigned with 108; code table `0x2D348` contains zero at indices 0/39/74/102/108 and one elsewhere. Two jump entries map zero to true/one to false. Full 109-byte image read corroborates all five towns and all non-town entries; this is engine code data, not a copied game table.

| Original location | Semantic meaning |
|---|---|
| Unit +0/+4/+0x10/+0x2C/+0x5C/+0x60/+0xDC | type/class/mode/path/own base list/inventory/event chain |
| Game +0xA8/+0xB8; manager +0xA18 | frame/event manager/saved successor |
| Event +0/+2/+4/+8/+0x14/+0x18/+0x2C | type/condition word/frame/target/A/B/direct callback |
| Event +0x1C/+0x20; +0x24/+0x28 | manager successor/predecessor; unit successor/predecessor |
| Event +2 masks 1/4/8 | running/persistent/pending removal; allocator free marker is separate mask 2 |
| DataTables +0xBC4/+0xBC8, record stride 0xC4, byte +0x42 | CharStats collection/count/RunDrain |
| inventory magic 0x01020304, +0x14/+0x18; page +8/+9/+0x0C | validity/page collection/count; dimensions/slots |
| Common 0x9D378 bytes 13,1 | Body-page size descriptor, independently checked 2026-10-15 |
| Common globals 0x9FB98/0x9FB94, stride 0x1A8, record +0xD8 | Items collection/count/speed |
| unit skill context +0xA8 → +0x10; path +0x28 | reset context / continuation condition; full role remains SDR-Q04 |
| client record +0xA0/+0xA4/+0xA8+8*i/+0xAC+8*i, 20 entries | movement history index/time/positions, not drain inputs |

Drain takes unit in EDI, returns in EAX. Movement takes game in ECX/unit in EDX; its path result is retained independently of the drain's EAX. Original stat calls have unit/stat/value-or-delta/layer semantic order; stat reads have unit/stat/layer. Game thunks resolve through original PE tables: `0xA478→Common #10331/0x2FE10`, `0xA808→#10057/0x3C390`, `0xA7AE→#11139/0x1E670`, `0xA53E→#10695/0x719A0`, `0xA802→#10910/0x39430`, `0xA42A→#10973/0x38B70`, `0xA4D2→#10551/0x3A710`, `0xA472→#10887/0x3A740`. These were re-resolved on 2026-10-15. Common body-page calls `0x9222/0x9252` resolve through slots `0x890D0/0x890AC` to Fog #10045/`0x1CD50` and #10047/`0x1CDA0`.

Mode-transition imports are Game `0xA496→#11103/0x303E0`, `0xAC70→#10800/0x34C60`, `0xAAF0→#10334/0x36230`, `0xA4A2→#11090/0x33920`, `0xB018→#10498/0x35210`; movement state check is `0xA68E→#10494/0x33CD0`, and client-record getter `0xA4C0→#10920/0x2EB70`. State 13 was historically named blaze; it adds a call at actual Game `0x439A0`, not the symbol-like `0x429A0` previously published. Full preparation/state/arrival effects remain SDR-Q07. The frame caller invokes `0x2CA20/0x29BE0/0x2CBE0` before event pumping and `0x2D9B0/0x9D850/0x9D5B0/0x2C0F0` after; those locations alone establish no writer order. OWD-R11a supplies the independently verified separate resource-call order.

## Arithmetic, getter and storage corrections

The complete drain instruction path on 2026-10-15 confirms u8 RunDrain doubled, signed speed/10 with truncation, low32 weight product, low32 percent product divided signed by -100, low32 addition, signed minimum 1, Add→Read→conditional Replace→return ordering. No whole-operation RNG claim is made for unread surrounding callbacks. Items speed descriptor at Common `0x787C0`, with string `speed` at RVA `0x8F3C8`, maps the dword field +0xD8; the 2026-10-12 descriptor assessment was partial about type-register lifetime, while the drain itself unambiguously reads a dword. GDF's CharStats descriptor maps RunDrain to byte +0x42. Runtime cells do not prove either meaning by themselves.

The matching wrapper view for #10551 uses a dependency symbol whose numeric suffix suggested `0x39330`. Fresh original bytes at Common `0x3A727` are a relative tail jump to actual **`0x3A330`**. Source at that actual entry confirms own-array lookup/insertion, low32 addition, zero removal, propagation call and conditional tracking. The old location `0x39330` is inside unrelated arithmetic. This correction changes the public callee locator, not the requested delta. Entry-management and propagation callees are not completely verified here, so SDR-Q02 remains. An absent list returns from this additive writer, but no general stored-result contract follows.

Common #10887/`0x3A740` loads own list and forwards to #10261/`0x3A280`, then calls `0x38550` only when the actual inner result is nonzero and unit type is player. This inner zero/nonzero test is separate from the drain's unconditional disregard of the outer wrapper return. Direct inner control flow compares new-minus-old, can remove zero, and calls propagation; current STL-R06c/RLI-Q03 retains full entry-management/propagation gates. `0x38550` checks extended-list/stat eligibility and inserts a stat identity into a tracking collection. It is not itself an established network sender. No requirement for an opaque optional notification hook is derived from the unread path. The original PE export directory was independently resolved on 2026-10-15 to corroborate ordinal 10261 at `0x3A280`.

RLI-R02a/RLI-R02b selects an existing own or resolved array and performs exact lookup/minimum handling. It neither sums attached lists nor constructs those arrays. Therefore a request to the own list cannot establish the following selected value, final stored zero or client update. The current vectors explicitly supply post-add reads and a fixture persistence model. Exact original numeric arithmetic outputs 23512/23472/24305/21503/23525/23551 and 589 invocations are preserved at that qualified boundary. The depletion fixture retains -10 versus 0 as alternative supplied reads and requests zero in both; it does not assert an actual intermediate clamp. The old unqualified “never increases here” overflow statement lacked support under propagation and 32-bit wrap and is not retained as a stored-value guarantee.

The unchecked original mode table does not support an invented bounds clamp. The current domain is supported modes 0..19, with malformed modes left without a safe portable result. The complete transition remains gated: selected mode 2 becomes town walk 6 or walk 2; Common path preparation, skill reset and continuation read run; zero continuation calls Game `0x782F0` and returns before mode write/registration, otherwise `0xDA750`, mode setter, `0xE0750`, `0xE0B40` run. The local call sequence is evidence, not a complete mode-result contract.

## Named table cells

Historical 2026-10-12 reads through the owned reported original 1.13c game's Storm loader: CharStats 0+3360/3360 and BodyLocs 0+160/160 were complete; ItemStatCost 15500..18700, Armor 0..14000 and States 0..2600 were partial. Only named values/names were retained. Partial absence is never treated as proof that a row is missing. Effective hashes, archive suppliers, complete overrides and exact loaded-module hashes were not captured. Historical live compiled CharStats correspondence is recorded separately below.

Reuse: GDF's complete 2026-10-14 ItemStatCost 40109-byte/States 20920-byte/PlrMode 302-byte records and 2026-10-15 corroboration supply their named mappings within their recorded profile limits. DUR's 2026-10-15 BodyLocs header/full-file read covers 160/160 bytes with effective hash `6b3a6ac0bb1cef5e7f216f083ef85082acedf22c4bbfacf5b254fb36ecd8bbd5`; row 3 Code=tors. That recent read's catalog refresh failed; it corroborates the cell without independently establishing complete vanilla identity. Historical whole-file BodyLocs name evidence is retained separately, not retroactively hashed.

Fresh read-only 2026-10-15 collection selected the answering reported original LoD 1.13c profile separately from four reported modified profiles. The version catalog did not answer; historical original-profile identity remains the qualified basis. No new launch, gameplay, hook or memory mutation occurred. Each effective file was requested at offset 0/length 60000 through that game's Storm loader. Complete returned bodies were decoded; columns selected by header, rows by class/name/explicit ID. Full CharStats and ItemStatCost coverage was obtained; Armor's first 60000 bytes include complete header and all three named rows, not complete table. Exact supplying MPQs, override inventory, effective-file hashes and current loaded-DLL identities were not captured for these fresh reads. No new compiled correspondence or gameplay label is promoted.

| Effective virtual path | Offset / returned / total | Coverage |
|---|---|---|
| data/global/excel/charstats.txt | 0 / 3360 / 3360 | complete header/file; Expansion separator excluded for class indices |
| data/global/excel/armor.txt | 0 / 60000 / 76370 | complete header/three selected rows; partial file; no absence claims |
| data/global/excel/itemstatcost.txt | 0 / 40109 / 40109 | complete header/file; explicit ID matches zero-based selected data-row index |

| File / row identity | Column | Raw cell | Interpretation / supports |
|---|---|---|---|
| CharStats class Amazon, prepared index 0 | RunDrain | 20 | Decimal u8 input; SDR-R04/SDR-V04 |
| CharStats class Barbarian, prepared index 4 | RunDrain | 20 | Decimal u8 input; SDR-R04/SDR-V01 |
| CharStats class Assassin, prepared index 6 | RunDrain | 15 | Decimal u8 input; SDR-R04/SDR-V03 |
| Armor name Quilted Armor | speed | 0 | Decimal i32 input; SDR-R06/SDR-V04 |
| Armor name Ring Mail | speed | 5 | Decimal i32 input; SDR-R06/SDR-V03 |
| Armor name Full Plate Mail | speed | 10 | Decimal i32 input; SDR-R06/SDR-V02 |
| ItemStatCost explicit ID 10, row 10 | Stat | stamina | Name only; fixed code selector independently read |
| ItemStatCost explicit ID 154, row 154 | Stat | item_staminadrainpct | Name only; fixed code selector independently read |
| BodyLocs data row 3, header excluded/index 0 first | Code | tors | Reused DUR corroboration; torso name |

The current arithmetic fixtures provide every numeric input explicitly, so profile qualification does not gate their supplied-input assertions. Text values alone do not establish compiled runtime values. Historical compiled Barbarian/Assassin reads below corroborate only those named fields. Header coverage or a table name never establishes selector meaning, configurable ids or full loader behavior.

## Historical runtime records

The 2026-10-12 session reported LoD 1.13c, single-player Normal, a level-1 Barbarian named Seam, dead in mode 17 and undriven. No save was edited, no member launched, and no gameplay action was performed. Loaded-DLL fingerprints were not recorded; relocation/address details remain in the private audit.

| Target / assertions | Date / location | Records | Actual result / limits |
|---|---|---|---|
| Player type-0 mode dispatch, SDR-R01b/SDR-R01d | 2026-10-12; Game 0x79A10, 8 seconds | 0 | Inconclusive. Dead mode 17 and absent established trigger/capture eligibility do not confirm event removal, cadence or any drain assertion |
| Drain, SDR-R04 through SDR-R10 | 2026-10-12; Game 0x77BB0 | not armed | Entry prologue was refused before arming; forced override was not used. No drain observation |
| Compiled class inputs only | 2026-10-12; Common table root 0x99E1C, CharStats collection/count | memory reads, no gameplay records | Count=7. Barbarian bytes +0x40..+0x43 were 06 09 14 08: WalkVelocity=6/RunVelocity=9/RunDrain=20/LifePerLevel=8. Assassin bytes were 06 09 0F 08: corresponding RunDrain=15. These agreed with the loader text reads, not an executed drain |

No gameplay vector is runtime-confirmed. In particular, the zero-record result does not prove that death removed an event, and the refused drain trace does not prove a function was never called.

## Observation procedures

All procedures use owned original-profile LoD 1.13c, Normal single player, controlled naturally reachable character/gear/stats. Identify the effective data profile and actual module fingerprints. Use game-relative frame observations and raw stamina rather than the integer panel alone. Observe movement consumer Game `0x79B90`, the drain call/output and actual storage/mode writes with suitable tools. Validate trigger, address, record cap/window and capture loss before interpreting absence. A refused entry or zero records is inconclusive; do not override an existing hook or alter live state to force confirmation.

- **SDR-O01:** level-1 Barbarian, no torso/percent source, start full raw stamina 23552 outside town in Blood Moor, run continuously on open ground. Record first running/drained frame, adjacent raw values, other writers, all 589 invocations and final Walk request/actual mode separately. Conditional expectation is 40 raw units each, 32 at 588, zero request at 589; actual stored outcome/cadence must be measured.
- **SDR-O02:** same controlled setup with Full Plate Mail/strength requirement met, speed=10. Expect 80, 32 at 294, zero request at 295 under no-other-writer conditions.
- **SDR-O03:** same setup with Ring Mail speed=5; compare adjacent drain amounts with O01, expect 40 rather than 60/80.
- **SDR-O04:** observe reachable running inside Rogue Encampment/level 1 with positive stamina, and a naturally reachable zero-stamina running case if available. Verify drain returns before stat work. Other resource changes are not counted as drain; failure to reach zero-running is not confirmation.
- **SDR-O05:** identify actual total stat-154 getter result 50, then 100, using legitimately obtainable gear/conditions. Do not assume an affix name supplies it. With base 40/no torso weight expect 20 then 1 raw unit per invocation. If required inputs cannot be reached, record the limitation.
- **SDR-O06:** outside town, naturally deplete stamina then request Run; record transition request, chosen mode and stored mode separately. This confirms only the reached player-mode integration case if valid, not the whole drain or zero-subtraction contract.

## Question assessment and migration, 2026-10-15

| Boundary | Existing evidence assessed / new attempt | Result |
|---|---|---|
| Event recurrence, SDR-Q01 | STL frame scheduler; MCI fresh association; full pump/player walker/removal/pending-removal/register source/image reads | Encoded instructions and running-removal behavior resolved locally. Creation/join/duplicate/arbitrary reentry invariant still open |
| Storage, SDR-Q02 | NRF getter correction, STL-R06c/RLI-Q03; matching Add wrapper first; actual image target; additive/replacing source entries | Ordered requests verified; actual Add target corrected to 0x3A330. Entry management/propagation/selected-value production still open |
| Delivery, SDR-Q03 | Replace wrapper/tracking source and original image | Direct tracking call recognized; no established message/delivery contract |
| Transition, SDR-Q04 | Local transition source/image/jump table and resolved imports; MCI planner/NRF mode setter | Request/stop-vs-setter sequence retained, full path/transition integration open |
| Gameplay/data, SDR-Q05 | Historical failed captures and compiled fields; GDF/DUR records; new qualified named-cell reads | No gameplay campaign or promotion; named cells corroborated with explicit identity/coverage limits |
| Resource interference, SDR-Q06 | OWD-R11a player resource consumer and old frame caller | Life→stamina→mana order reused; running gate/interleaving and actual stamina writer effects remain outside-boundary |
| Movement/page dependencies, SDR-Q07 | Full movement caller, preparation candidate; body getter/page image, allocation import resolution, MMF/MCI/MHD/MMC boundaries | Existing page and supplied-success lazy page settled; full preparation/path/history/allocation lifetime remain open |

Consumption search covered both the spec filename and SDR prefix across the repository. No implementation code, tests or implementation plan outside gap records consumed it. Current ids use clean execution order without aliases; the same twelve parent behavior/vector families and seven question families remain. No implementation comparison, gameplay, code changes or source modifications were performed.

Historical reader passes dated 2026-10-13 retain their original scope, not approval of this contract. Pass 1 reported GUESS 24/GAP 11. Findings addressed group selection, event conditions/type/direct-call arguments, unit-chain registration, mode domain, town table/linked-record guards, class-record computed-address guard, i32 amount, unsigned item bound, page dimensions/growth, item globals, raw stat reads, unresolved Add/Replace, transition town mapping, minimum duration, regeneration order, state-13 name and surrounding movement/bookkeeping effects. Pass 2 reported GUESS 5/GAP 2: saved-successor/reentrancy, allocator association, transition's early-stop branch, replacement/tracking return condition and lazy page growth were folded into that version; no new reading was performed for that pass. Their unsupported inferences are corrected above. Complete original wording/bytes are privately preserved; no historical MISMATCH/DISCRIM/COVERAGE/HARDEST counts were recorded.

Current presentation moves binary locations/layouts/dispatch details here, keeps semantic arithmetic/requests/vectors in the contract and qualifies actual storage, recurrence and mode claims. New local evidence changes old asm-only registration/transition claims to precise unresolved integration, and verifies the supplied traversal/page/zero-request/conditional-duration slices. All runtime flags remain no. Mechanical lint/self-review and the fresh-reader handoff are recorded after the final writer pass; fresh-reader completion and maintainer approval are not inferred.

Writer self-review on 2026-10-15 checked the preserved twelve vector families, seven question families and six observation families against the resulting contract, including all original exact raw arithmetic outputs and qualified 589-frame duration. The complete original 56345-byte document was archived privately and its SHA-256 rechecked. Public files were checked as UTF-8/LF; no broad technical-token spacing rewrite was used. The exact command `python .agents/skills/spec-writing/tools/check_spec.py player-stamina-drain-while-running` returned 0 errors and 0 warnings after writer edits. This is document lint, not a gameplay or implementation test. A fresh FULL1 reader pass is required on this resulting semantic contract; historical reader counts do not satisfy it.

### Campaign full reader pass 1 and triage, 2026-10-15

The fresh reader received the semantic contract and its cited dependency specs, without original-code or implementation access. The input included the two ordinary-language Description sentences. Counts were MISMATCH 0, GUESS 1, GAP 2, DISCRIM 1, GATED 21, COVERAGE 7, HARDEST 3. GATED counted seven questions, two unresolved rule slices and twelve supplied-stage vector families; those limits remain explicit rather than being treated as defects or approvals.

| Finding | Outcome |
|---|---|
| GUESS 1 / GAP 1: V05(d) outer Replace return was not distinguished from the inner change result; tracking output missing | Edited R10b/input table and V05 boundaries. Cases (a)..(d) stop at a logged supplied outer-wrapper request; (d)'s outer result 2 is ignored by drain without asserting its interior. New isolated wrapper cases (e)/(f)/(g) supply inner results 0/1/2 and assert respectively no Track / one Track / one Track. Actual storage and tracking effects remain SDR-Q02/SDR-Q03 |
| GAP 2 / DISCRIM 1: V08(g) did not observe running during callback or after it | Added callback-entry running=true, immediate post-callback running=false before removal, exact ordered observation log for E/F, and unvisited new-head N remaining false. Omitting either write now fails the fixture |
| COVERAGE 1: R01b running flag | Resolved by V08(g)'s added observations above |
| COVERAGE 2: R01c nonzero valid type / type>=15 skip | Added V08(i) type 1 selecting the established mode-end handler, with ignored event-handler return. No change to type>=15 skip: V08(d) already supplies type 255 without a direct callback and asserts no call; its direct-callback case separately proves precedence. Type 2's null-handler case was also already present |
| COVERAGE 3: R01e absent-unit mode selection / supported null handler | Added V08(j): absent unit selects mode 0 and passes absent target to its supplied handler. No supported-null-mode fixture: the original twenty supported entries are all nonnull; the guard is retained without inventing a null entry or malformed-mode fallback |
| COVERAGE 4: R02 full absent-path drain | Added V06(d) with present player/absent dynamic path, non-town result, supplied positive read 23512, Add -40 and drain 1. It asserts no safe whole-movement null-path outcome |
| COVERAGE 5: R04 RunDrain 0/255 | No change: V10(c) already supplies RunDrain=0 and proves the floor. Value 255 has the same u8 read/doubling recipe as V01/V03 and no distinct branch or saturation; R04 and section 3 state its width. No exhaustive u8-input coverage is claimed |
| COVERAGE 6: R05b out-of-range body location | No change: the drain always supplies fixed location 3. Existing invalid-inventory and mismatched-page cases already demonstrate the getter's no-item return and continued unweighted drain. The preserved generic helper guard does not create a variable-location producer in this consumer |
| COVERAGE 7: R10b changed/unchanged routing | Resolved by explicit supplied inner results 0/1/2 in V05(e)..(g); these are request-routing assertions, not integrated stored-change claims |
| HARDEST 1: V08 traversal across supplied stages | Running observations and exact logs added as above; preamble and case-local execution boundaries remain explicit. No list-population/reentry integration inferred |
| HARDEST 2: V05 replacement-result/tracking interpretation | Separate outer ignored result and inner nonzero routing now have independent cases and named inputs/outputs; actual producer effects remain gated |
| HARDEST 3: V07 wrap/truncation order | No change: R07 explicitly wraps multiplication before signed division and wrapped addition. V07(a)/(b) distinguish truncation direction; (c) gives low32 product 24 and exact final amount 40, distinguishing wide multiplication |

Root review additionally identified that V09's Walk assertion requires the movement consumer. Its boundary is now repeated supplied post-preparation movement stages, with one drain and PathStep per invocation, one Walk request at invocation 589, and 589 total PathStep requests. Stored mode/cadence/other-writer gates are unchanged. The Common inner replacing-writer ordinal was independently resolved to #10261/`0x3A280`; the public source-profile prose now uses descriptive original LoD 1.13c wording, retaining the literal session classification only privately. These are source-attribution/presentation corrections, not new gameplay claims.

Every FULL1 finding has an outcome above. Because vectors, routing precision and an execution boundary changed, a fresh full reader pass 2 is required; pass 1 is not the completion review of this resulting contract. No new gameplay, implementation tests, implementation comparison or source mutation accompanied triage.

### Campaign full reader pass 2 and triage, 2026-10-15

The fresh full reader recomputed all executable cases from the resulting contract and dependencies. Counts: MISMATCH 0, GUESS 0, GAP 1, DISCRIM 2, GATED 28, COVERAGE 6, HARDEST 3. The GATED count comprises seven questions, nine rules whose integrated effects have gates, and twelve supplied-stage vector families. This counting convention differs from pass 1's unresolved-rule-only subtotal; it does not identify seven newly introduced gates.

| Finding | Outcome |
|---|---|
| GAP 1: vector fixture stamina absent from section 5 | Added a section-5 output row explicitly naming synthetic fixture stamina and its read/zero-assignment model. It is a supplied example output, not an asserted original write; SDR-Q02 remains unchanged |
| DISCRIM 1: V05 Walk request omitted game/unit identity | Added shared semantic log expansion: G is the supplied game, U is the current unit; WalkRequest(m) includes exactly G/U/m. A request for a different game or unit fails even with correct mode. Unit Add/Read/Replace, Drain and PathStep identities are explicit as well |
| DISCRIM 2: V08(f) Walk request omitted game/unit identity | The same shared notation applies to V08(f) and all other shorthand logs. This fixes argument assertions without prescribing an original calling convention or API |
| COVERAGE 1: R01c selected nonnull handler types 2..14 | No duplicate table matrix: V08(b) exercises original type 0/non-null selection, V08(i) type 1/non-null selection, V08(d) type 2/null and type 255/out-of-range. Every other present table entry follows the same verified type-bound/lookup/argument-forwarding recipe, with no distinct selector branch. No per-type handler interior is claimed |
| COVERAGE 2: R01e null handler / supported nonmovement modes beyond 0 | No change: all twenty original supported entries are nonnull, so no null entry is fabricated. V08(j)'s mode-0 supplied handler covers nonmovement selection through the same table-index/pass-through recipe as other nonmovement modes; their interiors remain outside this stage. Modes 2/3/6/19 independently exercise the movement gate |
| COVERAGE 3: R05b out-of-range body location | No change for the same fixed-consumer reason recorded in pass 1: running drain always supplies location 3. Existing invalid-inventory/page-mismatch cases produce the no-item/unweighted-continuation result. No extra variable-location producer or exhaustive helper matrix is introduced |
| COVERAGE 4: R06 valid torso with absent items collection | Added V12(g): in-range torso class 0/count 10 with absent collection terminates fatally before later percent/stat/Walk requests; no numeric return. This discriminates the stated fatal path from out-of-range class's safe absent-record result |
| COVERAGE 5: R09 absent-own-list forwarding | Added V01(b), an isolated nonzero-add wrapper fixture with absent own list and exact ListAdd(absent,10,-40,0) request. It asserts neither getter/drain result nor stored stamina |
| COVERAGE 6: R10b nonplayer with nonzero inner result | Added V05(h), isolated wrapper/type 1/inner result 1: ListReplace is forwarded but no Track request. No nonplayer running-drain or actual stored result is asserted |
| HARDEST 1: R01b/V08(g) running/removal/saved successor | No change: pass-1 edits explicitly observe running at callback entry, immediately after callback before removal, and final F/new-head state; exact E/Remove/F ordering remains supplied-list behavior with lifecycle gates |
| HARDEST 2: R07/V07 low32 product then signed truncation | No change: rule order and exact V07 product 24/adjustment 0/output 40, plus positive/negative 33 cases, already distinguish wrap and truncation |
| HARDEST 3: R10b inner/outer/drain returns | No change to established separation: V05(d) supplies only ignored outer result; V05(e)..(h) execute isolated inner routing with explicit logs; actual inner-result producer/storage/tracking remain gated |

No behavior gap was closed by these fixture/assertion changes. Rule labels, existing question/gap boundaries, historical observations and the ordinary-language Description remain unchanged. The additional cases and output/argument assertions require fresh full reader pass 3. No additional original-code campaign, gameplay or implementation test was performed during this triage.

### Campaign full reader pass 3 and triage, 2026-10-15

The third fresh full reader recomputed the supplied arithmetic and expected outcomes. Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1, GATED 11 grouped entries, COVERAGE 7, HARDEST 3. GATED groups were recurrence, storage, tracking/delivery, complete transitions, observations/profile, other writers, preparation/allocation, V05 inner/outer boundaries, V06/V08 isolated stages, V01(b) own-list forwarding and general supplied persistence/inert callback/path boundaries. Grouping changed the count, not the underlying gates. The reader did not supply a case tally and made no source, gameplay or implementation-test claim.

| Finding | Outcome |
|---|---|
| DISCRIM 1: direct callback result 2 was tested only at isolated selector, not event-list caller | Added V08(k): persistent list E/F, E's direct D returns 2 and F's supplied type-0 handler returns 1. Both are visited/retained with running cleared; no ModeEnd or Remove request. This separates list-pass callback-result disregard from mode-dispatch result handling |
| COVERAGE 1: R01e supported mode with null handler | No change: all twenty supported mode table entries are nonnull. Event-slot null coverage is separate; no null mode entry or malformed-mode fallback is invented |
| COVERAGE 2: R01b saved successor itself changed during callback/removal | Added V08(l) with explicit supplied callback effects removing F and changing saved successor from F to J. The list pass next visits J, skips F, retains E/J and ends with absent saved successor. The callback mutation is a supplied fixture input; no arbitrary lifecycle producer is verified |
| COVERAGE 3: R03 valid class index into absent collection | No change: this is outside the valid prepared-collection input boundary, with no promised safe failure in R03. The original checks the computed record address, not the collection base independently. No portable safe-null or exact fatal result is invented for malformed storage |
| COVERAGE 4: R05b out-of-range location / other dimension mismatches | No change: running drain supplies location 3, while the generic guard remains stated. Existing 12 by 1 page mismatch exercises the no-item/no-resize recipe; a height mismatch uses the same specified dimension predicate and result. No exhaustive malformed-page matrix is claimed |
| COVERAGE 5: R05c allocation failure | No change: successful allocation is an explicit supplied input; failure/lifetime integration remains SDR-Q07. An invented failure result would cross that boundary |
| COVERAGE 6: R06 in-range class/present collection/absent record | No change: the consumer tests the optional lookup result, and V12(f) already exercises absent-record/unweighted continuation. The original indexed collection has no established per-row hole mechanism; a computed zero address is not promoted into a new table-hole producer or general corruption contract |
| COVERAGE 7: R12 positive w other than 40 | No change: the conditional duration uses the same ceiling/max recipe for supplied positive w. V02 independently supplies amount 80, while V09 discriminates ceiling using remainder 32 at invocation 588. No per-amount repeated-arithmetic matrix or exhaustive-duration coverage is claimed |
| HARDEST 1: V08 mutable traversal/running/due-persistent/result routing | V08(k)/(l) add the two distinguishing pass-level logs. Existing callback-entry/post-callback running observations and supplied lifecycle boundaries remain unchanged |
| HARDEST 2: V05 inner result versus wrapper/drain return | No change: V05(d) stops at ignored outer return; V05(e)..(h) separately assert inner zero/nonzero/player guard logs without original storage effects |
| HARDEST 3: V07 wrap before signed division | No change: R07 explicitly orders low32 multiplication before truncating signed division; V07's wrapped product 24 and positive/negative percent cases retain exact outcomes |

Only V08(k)/(l) and their distinguishing assertion were added after full pass 3. Rule behavior, inputs outside those cases, labels, ownership, acceptance and gate boundaries remain unchanged. All findings have an outcome above, so the next fresh reader may check this delta under the workflow's fourth-pass procedure. No new source reading, gameplay, implementation comparison or implementation test was performed.

### Fresh delta reader pass 4 and completion review, 2026-10-15

The fresh independent reader checked V08(k)/(l), their shared defaults/log notation and governing R01b/R01c/R01e contracts. Unchanged arithmetic/cases retain the three full reviews above; only those two new scenarios and their distinguishing assertion changed after FULL3. Delta counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0, GATED 1, COVERAGE 0, HARDEST 3. No case tally was claimed.

All three HARDEST entries receive no change: V08(k) explicitly retains both persistent entries and continues to F despite direct D returning 2, with no ModeEnd/Remove; V08(l) supplies the mutation and observes J rather than stale F, ending with no saved successor; R01b and V08(g)/(k)/(l) specify callback-time running, post-callback clearing, removal conditions and traversal order. The one reviewed gate remains R01b/Q01: actual recurring population and arbitrary callback mutation producers are unestablished, while supplied-list consumer results are available. No callback fixture verifies its real producer.

Root review read all three documents, all additions and every reader disposition; it separately checked the complete removal image, five-group pump call sequence, additive forwarding target and retained raw arithmetic. The resulting contract has no outstanding blocking reader finding. Document review is complete; draft status, original storage/lifecycle/transition/delivery/profile/observation gaps and maintainer approval remain. No gameplay, implementation tests or new implementation comparison ran. Publication repeats document checks on committed bytes.

### Delta reader passes 5 and 6 after the dependency restructure, 2026-10-15

The 2026-10-15 restructure moved no SDR rule; it remapped cross-spec ids: former GDF R14a to GDF-R05a, former GDF R03a to GDA-R06a, former GDF R11a to GDF-R02, former NRF R12b/R12c to RLI-R02a/RLI-R02b, former NRF Q20 to RLI-Q03, former NRF Q03 to NRF-Q02, former MCI R12b to MIR-R01b and former MCI Q03 to MIR-Q01 (split maps). Root review checked each target in the current dependency specs. Every remap matched the split maps and resolved to the needed contract; none was a renumbering collision. RLI-R02a/RLI-R02b text is identical to the former NRF R12b/R12c.

**Delta pass 5** (fresh Sonnet reader; current SDR, NRF, GDF, GDA, RLI, STL, MIR specs). Counts: MISMATCH 0, GUESS 2, GAP 1, DISCRIM 3, GATED 4, COVERAGE advisory (missing-entry reads, resolved-versus-own selection, fMin on stat 10), HARDEST 3. All foreign ids resolved, including OWD-R11a, STL-R01 and RLI-R04a. All dependent vectors recomputed unchanged.

| Finding | Outcome |
|---|---|
| GUESS 1: R07/R10a do not say unit read `Stat(unit,...)` or own-array list read `GetStat(list,...)` | Spec edit (wording/binding, no behavior change). SDR-R07, SDR-R10a and the percent, post-add and stat-array input rows now name the RLI-R02a unit read and then RLI-R02b. Verified 2026-10-15 by original-image disassembly: D2Common `0x38B70` (#10973) and `0x39430` (#10910) load the unit's list (+0x5C), check the id against the definition count, and call `0x38A80`. That function selects the +0x48 array when list flag bit 31 is set and the +0x24 array otherwise, then applies the fMin/owner-type/MinAccr replacement. This is RLI-R02a's unit read plus RLI-R02b |
| GUESS 2: whether fMin/MinAccr apply to stat 10/154 | No change. The drain consumes the RLI-R02b result unchanged, and vectors supply getter results. Minimum behavior is owned by RLI-R02b/RLI-V03 |
| GAP 1: GDF-R02's table does not list 154 item_staminadrainpct | No change. This predates the restructure (former GDF R11a had the same scope). GDF-R02 fixes ID = row index, and vectors use numeric 154. The name is a companion named cell, so it is naming only |
| DISCRIM 1: own-array list read still passes | Resolved by the GUESS 1 edit. RLI-V02 discriminates the unit read from the list read in the owning contract |
| DISCRIM 2: missing 154 entry treated as error | No change. Owned by RLI-R02a; RLI-V02 discriminates the missing entry (returns 0, no minimum) |
| DISCRIM 3: R01d registration by reusing the MIR-R01b association | No change. SDR-R01d is unresolved under SDR-Q01, has no vector by design, and expressly says MIR-R01b does not establish recurring dispatch |
| HARDEST 1-3 (unit/list read; minimum on stamina; reach of MIR-R01b) | First resolved by the edit; second and third as GUESS 2 and DISCRIM 3 |

Also: the Depends-on line now links RLI, GDA and MIR ids to their owning specs and adds NRF-Q02 (link/wording only).

**Delta pass 6** (fresh Sonnet reader on the edited contract). Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 1 advisory, GATED 7 questions plus R01d/R11, COVERAGE 2, HARDEST 3. The reader confirmed that all three no-change items and the GUESS 1/DISCRIM 1 resolution hold. It recomputed V01-V07 and V09-V12 with no difference. Outcomes:
- Advisory DISCRIM (no SDR vector separates the unit read from the list read, or applies the minimum to stat 10): no change, because RLI-V02(a)/(c) and RLI-V03 discriminate these in the owning contract.
- The stat-154 name note: as GAP 1 above.
- COVERAGE (non-player stat-154 owner; class 7 with a present unit): no change, because they are outside the running-player consumer and V11 covers out-of-range classes.
- HARDEST (V07c wrap, V05 tracking guard, V08 traversal): no change, because each is already stated and has a discriminating vector (FULL3 triage).

Final applicable run: MISMATCH 0, with no unresolved GUESS, GAP or DISCRIM. Lint: 0 errors, 0 warnings.

### Focused publication review of the current contract, 2026-10-23

A fresh blind implementer read the current SDR contract and its cited dependency specs without this companion or historical review results. The reader independently recomputed SDR-V01–SDR-V12's numeric outputs, conditional duration and local draw/order claims and found no arithmetic, RNG or ordering contradiction. In particular, V07(c)'s `107374183*40` retains low32 value 24; at supplied 25 fps and no other writer, 23552 raw stamina with drain 40 reaches 32 after 588 invocations and depletes on 589. It identified a fixture ambiguity in V05(b)/(c) and missing discrimination for R06's wrapped weight product. The writer clarified those two supplied getter results as overrides and added synthetic V10(d). A separate blind delta review recomputed V10(d)'s multiplier 100000001, wrapped signed product -294967256 and post-floor drain 1, then approved the fixture wording and rule attribution without a blocker.

An independent publication reviewer checked the three-file handoff, all cited dependency IDs, provenance labels and Q01–Q07 boundaries, and approved the qualified static/supplied-input publication. That review found 25 relative links resolving at the time of review; committed-byte checks are reported in the PR. The earlier 2026-10-15 full/delta passes remain dated evidence, not a claim of new source reading. No new original-game observation, binary reading, actual storage/cadence integration, gameplay implementation or 1.14d equivalence was established. Draft status remains appropriate.
