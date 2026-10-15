# Crushing Blow evidence

Behavioral contract: [combat-crushing-blow](../../specs/combat-crushing-blow.md).
Original artifacts: [source manifest](../source-manifest.md), `LOD113C-D2GAME`,
`LOD113C-D2COMMON`, `LOD113C-FOG`, `LOD113C-STORM`; original Expansion 1.13c, Windows x86.
Descriptions of functions are analyst labels, not original exported names.

## Provenance

The 2026-10-13 investigation read the callback, apply dispatch and imported predicates through
original-code disassembly, and read callback slots/bonus constants from the original image.
The historical analysis images are represented by the manifest's D2Game/D2Common identities;
2026-10-15 rehashes of both available original bundles match those fingerprints. This does not
identify DLLs loaded by the earlier running session: complete loaded-DLL fingerprints were
not captured. New fingerprinting does not renew the original reading or identify historical
archive contents. Private source forms, locations, commands and complete original documents
are preserved separately with SHA-256 integrity records.

2026-10-15 targeted source and independent original-image instruction listings checked the
callback at D2Game RVA `0xACC30` through its returns and bonus helper `0xAF5F0`. No reconstruction
checked against the original machine code was available for these new reading targets.
The checked original analysis file has SHA-256
`24b293b86f95bea0228826c502ea04737ba21b91ca9832e52aac4b11fef14f53` and 1138688 bytes. The shared
getter, item-event, arithmetic and overlay records retain their own dates/limits; no gameplay
campaign occurred.

## Original function references

| Artifact | Function RVA | Export ordinal | Analyst description | Rules | Method / boundary |
|---|---|---|---|---|---|
| LOD113C-D2GAME | `0xACC30` | none | Crushing Blow callback | CBL-R03 to CBL-R08 | 2026-10-13 disassembly; 2026-10-15 entry-to-return disassembly and independent original-image instruction listing; local requests/returns, not downstream stored effects |
| LOD113C-D2GAME | `0xAF5F0`; table `0xFB614` | none | Player-count bonus | CBL-R05 | 2026-10-13 image/live constant read; 2026-10-15 full instruction listing: signed table/formula branch and low32 multiplication |
| LOD113C-D2GAME | `0x22A80` | none | Test supplied monster-data condition | CBL-R04a/CBL-R04b | 2026-10-13 reading reused; 2026-10-15 full 48-byte listing: null/type/data guards, word mask test, boolean return; no classification producer proof |
| LOD113C-D2GAME | `0x14D0` -> `0xF44B0` | none | Signed scaled division | CBL-R04a/CBL-R06 | Consumed NRF-R05, including multiply/divide-first width, order and fault limits |
| LOD113C-D2GAME | `0xDE0C0` | none | Apply-stage dispatch | CBL-R02a | 2026-10-13 dispatch reading; consumed DPO-R13a/DPO-R23 current contract, not all upstream predicates |
| LOD113C-D2GAME | `0xDED30` | none | Queued record copy/apply | CBL-R02b | 2026-10-13 flag replacement reading; DPO supplied queue/apply contract |
| LOD113C-D2GAME | `0xD9470`, `0x9E290`, `0x37190`, `0x370E0`, `0x37110`, `0x37030` | none | Registration, selector, insertion, lookup, removal, dispatch | CBL-R01 | Reused IEP-R01/R02/R03/R05/R07 dated 2026-10-12/13, targeted registration correction 2026-10-14; automatic notifications remain IEP-Q01/IEP-Q03 |
| LOD113C-D2GAME | table `0x1077A8`, slot 16 | none | Callback selector binding | CBL-R01 | Original-image and historical live pointer read maps selector 16 to `0xACC30`; cell names alone do not identify callback |
| LOD113C-D2GAME | `0x3EDE0` -> `0x3AFF0` -> `0x3AE10` | none | Missile contact hand-off | CBL-R02b/CBL-Q05 | 2026-10-13 chain, targeted 2026-10-15 `0x3AE10` reading; OR transfer and resolved-owner hand-off, wider contact eligibility is MHD-owned |
| LOD113C-D2COMMON | `0x39430` -> `0x38A80` | 10910 | Read registered stat identity | CBL-R03 | Consumed NRF-R12b/NRF-R12c, existing selected array/definition guards/minimum; no summation during getter |
| LOD113C-D2COMMON | `0x38B70` -> `0x38A80` | 10973 | Read unit stat | CBL-R04a/CBL-R06 | Same current getter owner; historic summed-read description corrected below |
| LOD113C-D2COMMON | `0x3A740` -> `0x3A280` | 10887 | Replacing unit-stat request | CBL-R07/CBL-Q10 | Caller request established; historical partial setter read does not verify full propagation/notification; NRF-Q20 remains |
| LOD113C-D2COMMON | `0x1A2D0` | 11104 | Five hireling-class membership | CBL-R04a | Reused 2026-10-13 reading: recognized classes 271/338/359/560/561; false for non-monster |
| LOD113C-D2COMMON | `0x1A550` | 10064 | Class boss predicate | CBL-R04a | Reused 2026-10-13 reading: supplied record or bounded class lookup, record byte bit 6; missing/out-of-range false |
| LOD113C-D2COMMON | `0x6A330` | 10623 | Missile runtime-condition getter | CBL-R02b | 2026-10-13 reconstruction checked against original machine code reused; 2026-10-15 source read: null/non-type-3/null-data gives 0, otherwise runtime condition word; no producer classification |
| LOD113C-D2COMMON | `0x32C10` | 10689 | Overlay operation | CBL-R08 | LSE-R14 dated 2026-10-14, checked against current 2026-10-15 boundary: guard/search/allocation/ordered requests; attach/list-write/update stored effects gated |
| LOD113C-D2COMMON | `0x37F40`, `0x38100`, `0x3A7D0`, `0x3A280`, `0x665C0` | internal | Overlay local search/allocation and boundary calls | CBL-R08/CBL-Q09 | LSE-R14/STL-R06a local search/allocation; attach STL-R06b, list write STL-R06c and update NRF-Q17 retain integration gates |

### Import resolution

The historical PE import/export resolution identifies the following D2Game thunk to D2Common
ordinal/target links. Current compatible getter/overlay records corroborate those same links;
location identity is distinct from complete effect verification.

| D2Game thunk RVA | D2Common ordinal | Target RVA | Use |
|---|---|---|---|
| `0xA802` | 10910 | `0x39430` | registered chance identity |
| `0xA42A` | 10973 | `0x38B70` | player count/life/resistance |
| `0xA472` | 10887 | `0x3A740` | replacing life request |
| `0xA7F6` | 11104 | `0x1A2D0` | hireling membership |
| `0xA580` | 10064 | `0x1A550` | boss predicate |
| `0xB108` | 10623 | `0x6A330` | missile conditions |
| `0xA83E` | 10689 | `0x32C10` | overlay request |

## Table-data verification

Read date: **2026-10-13**, completed **22:57:43 UTC**. The selected running member reported
the vanilla LoD 1.13c launch profile, distinct from responding modified-game profiles. The
record did not independently hash installed archives, identify physical supplying MPQs or
capture a complete override inventory. No override was reported by member identity; this is
not proof of absence. Claims apply to that effective virtual-file set.

Method: original Storm `SFileOpenFileEx` loader, separate reads at offset 0/length 65536,
complete decoded bodies parsed as tab-separated text with columns selected by header name.
No table or complete row is distributed. Historical effective-file hashes were not captured;
later GDF file hashes do not retroactively identify this record's bytes.

| Virtual path | Requested range | Returned / total bytes | Coverage |
|---|---|---|---|
| `data/global/excel/itemstatcost.txt` | 0 + 65536 | 40109 / 40109 | complete header/file and cited rows |
| `data/global/excel/events.txt` | 0 + 65536 | 420 / 420 | complete header/file and cited rows |

**Named stat cells:** all from `data/global/excel/itemstatcost.txt`; rows selected by `Stat`.
The raw decimal `ID` cells were converted to integers. The named rows' IDs also agree with
their zero-based data-row indices, excluding the header; that is a check on these cited rows,
not an inference that every table uses this convention.

| Row identity | Column | Raw cell | Interpretation | Supported claim |
|---|---|---|---|---|
| `Stat = hitpoints` | `ID` | `6` | integer 6 | section 3.1 life-stat identity; R06/R07 consume it |
| `Stat = damageresist` | `ID` | `36` | integer 36 | section 3.1 resistance identity; R06 consumes it |
| `Stat = monster_playercount` | `ID` | `100` | integer 100 | section 3.1 player-count identity; R04a consumes it |
| `Stat = item_crushingblow` | `ID` | `136` | integer 136 | section 3.1 chance-stat identity; R01 registration |
| `Stat = unit_dooverlay` | `ID` | `178` | integer 178 | section 3.1 overlay-stat identity; R08's label unchanged |
| `Stat = item_crushingblow` | `itemevent1` | `domeleedamage` | event name | R01's first registration choice |
| `Stat = item_crushingblow` | `itemeventfunc1` | `16` | integer selector 16 | R01's first handler selector, not proof of its meaning |
| `Stat = item_crushingblow` | `itemevent2` | `domissiledamage` | event name | R01's second registration choice |
| `Stat = item_crushingblow` | `itemeventfunc2` | `16` | integer selector 16 | R01's second handler selector, not proof of its meaning |

**Named event cells:** all from `data/global/excel/events.txt`, column `event`. This table
has no explicit ID column; the listed indices are zero-based data-row positions excluding the
header, consistent with the event-index convention documented in IEP section 2. Cells remain
strings; the index-to-event use in combat requires the separate code evidence.

| Data-row index | Raw cell | Interpretation | Supported claim |
|---|---|---|---|
| 1 | `damagedinmelee` | defender melee-damage event name | R02a's event-1 identity |
| 2 | `damagedbymissile` | defender missile-damage event name | R02a's event-2 identity |
| 5 | `domeleedamage` | attacker melee-damage event name | R01/R02a's event-5 identity |
| 6 | `domissiledamage` | attacker missile-damage event name | R01/R02a/R04a's event-6 identity |

Status of these named cells: **verified in 1.13c data**. They do not introduce new gameplay
rule ids or alter the existing Id map. `.txt` archive contents were read; compiled runtime
table equivalence was not newly established (GDF-R20a/GDF-R20b describe loader boundaries).
No attack, trace, memory write, save edit, launch or hook change was performed for these reads.
No gameplay vector becomes runtime-observed from this evidence.

### Separate code-consumer evidence

The 2026-10-13 callback reading and 2026-10-15 original-image listing show that `0xACC30`
extracts the chance stat identity from its callback value while explicitly selecting stat
100, stat 6, stat 36, event 6 and overlay 147. IEP-R03 supplies the registered identity;
selector-table resolution supplies the callback. The raw data cell alone cannot prove either
code mapping or compiled table correspondence.

## Dispatch evidence

These dated source chains establish reachability within their stated boundaries. All locations
are D2Game RVAs unless another module is named; table VAs are optional navigation aids.

### Melee chain

1. Server do-function dispatch `0xA19A0` reads `Skills.srvdofunc`, table `0x1074A8`; slot 1
   points to weapon do-function `0xA2C40`. Frame event `0xA2060` reaches this dispatcher
   (SCD-R06/SCD-R09b).
2. `0xA2C40` requests build `0xDDDE0` and queued apply `0xDED30`, on another branch through
   `0xA20C0` (DPO-R08d). The queue boundary is DPO-R12a.
3. `0xDED30` copies its queued record, replaces flags with 32 and calls `0xDE0C0` with
   defender in the return-value register and stack `(game,attacker,0,record)`.
4. After its gate/mitigation, `0xDE0C0` tests result mask 32 clear and mask 1 set. Apply flag
   zero dispatches event 5 on attacker/other defender, then event 1 on defender/other attacker,
   through `0x37030`, before clamp/leech/subtraction/status calls (DPO-R13a/DPO-R23).
5. `0x37030` walks the owner's event list (`Unit +0x90`). Matching entries call callback
   `Entry +0x14` with register `(game,event)` and stack `(owner,other,record,valueA,valueB)`;
   IEP-R07 owns traversal and mutation effects.
6. Supplied stat notification `0xD9470` resolves selector at `0x9E290`, creates through
   `0x37190`, and finds existing identity through `0x370E0`. ItemStatCost row 136 selects
   events 5/6 and selector 16. Callback table `0x1077A8` slot 16 points to `0xACC30`. Actual
   notification producers remain IEP-Q01/IEP-Q03; one pair is the supplied-notification rule.
7. `0xACC30` reads chance, draws, computes and requests life/overlay writes, with local result
   flag and return writes as detailed below. Imported getters/writes are resolved above.

### Missile contact chain

1. Per-missile frame dispatcher `0x3E3F0` selects do slot 1 `0x3F540`, which uses movement
   step `0x3F320`, cell scan and contact `0x3EDE0` (MMC-R01/MMC-R03/MMC-R14).
2. `0x3EDE0` reaches `0x3AFF0`, then `0x3AE10`.
3. `0x3AE10` requires type 3, owner link, successful owner lookup `0xE03A0` and a valid
   missile record. Its supplied condition getter D2Common #10623/RVA `0x6A330` returns runtime
   condition bits; bit 0 ORs flags 32 and bit 1 ORs flags 128. Incoming flags are retained.
   With the resulting hit bit set it passes its resolved owner as attacker and apply flag 1
   to `0xDE0C0`. Wider owner/eligibility/producer effects remain MHD-owned.
4. Nonzero apply flag runs `0xDC0B0` scale/mitigation first; with result mask 32 clear/hit set
   and flags 32 set, `0xDE0C0` dispatches attacker event 6 unless flags 128 is set, then
   defender event 2. The dispatcher/callback path is the same as the melee chain.

Other nonzero-flag callers share that branch; this does not prove their record eligibility
(CBL-Q07). Neither a local setter call nor update request proves client transmission (CBL-Q08).

## Local instruction rationale and corrections

### Getter and write boundaries

The 2026-10-13 draft said the getter summed own/attached contributions and credited NRF-R12c
with replacing-base propagation, entry insertion/no-op and player notification. Current
NRF-R12b/NRF-R12c instead selects an existing array and conditionally adjusts a minimum. It
neither sums contributions during the call nor establishes the setter's effects. The original
callback requests #10887 with stat 6, the computed value and layer 0; its succeeding death test
uses the local computed value, not a read-back. Accordingly CBL-R07 verifies that request and
local flag write, preserving every numerical requested value while gating actual stored
effects under CBL-Q10/NRF-Q20. Claims about `new + attached contribution` read-back are removed
as unsupported integration. Fixture persistence is explicitly supplied and proves no original
setter behavior. Notification production is likewise gated by IEP-Q01/IEP-Q03, rather than
equating a supplied notification with a real equip action. No client delivery follows merely
from these local calls (CBL-Q08).

### Overlay and successful return

The original CBL-R08 was asm-only because local overlay helpers were unread. LSE-R14's
2026-10-14 reading identifies the identical D2Common #10689/RVA `0x32C10`, with guard and first
primary-chain match, allocation through `0x38100`, attach request `0x3A7D0`, reloaded base-list
pending condition, list-stat request `0x3A280` and update request `0x665C0`. Its current contract
preserves unresolved stored integration. Reuse supports CBL-R08's local call with defender,
147 and zero, the operation's ordered requests, and caller return 1 regardless of helper
return; it does not resolve original overlay-count identity or stored attach/write/update.
Thus CBL-R08 is now verified static at that explicit boundary, with CBL-Q09/CBL-Q04 retained.
LSE-V11 covers helper selection/order; callback vectors cover the call arguments and condition.

### Record-condition transfer

CBL-R02b was asm-only and the draft separately assumed clear incoming flags. DPO's queued
record replacement and the dated contact reading establish the local transfer itself.
2026-10-15 rereading `0x3AE10` plus matched #10623/RVA `0x6A330` confirms bit-0/bit-1 tests
OR flags 32/128; false bits never clear incoming flags. The caller subsequently forwards its
resolved owner as attacker. CBL-R02b now verifies only this supplied transfer, with CBL-V14
distinguishing replacement from OR. Full missile eligibility/owner lookup and creation of
those runtime conditions stay MHD-owned, not inferred from this slice.

### Wrapped count, divisor and life arithmetic

The draft asserted the monster life clamp proved no overflow and that Crushing Blow could
not set death on positive life. That did not account for arbitrary signed player counts.
2026-10-15 full listings at D2Game `0xAF5F0` and `0xACC30`, corroborated by the matching original
image, establish:

- `0xAF5FD` subtracts 2, `0xAF600` uses a 32-bit multiply by 50.
- `0xACD09` adds the scaled term to the divisor, `0xACD12` doubles it for event 6.
- `0xACD20` sign-extends life and `0xACD21` performs signed division without a zero guard.
- `0xACD4B` subtracts resistance reduction; `0xACD4D` subtracts the amount from life before
  the signed-positive test and zero floor. No saturation or wider positive-count guard exists.

CBL-V15 supplies three result-distinct extremes. Count 2147483647 produces bonus -150 and
divisor -2, yielding requested life 384000 from 256000; life `0x7FFFFF00` instead wraps below
zero and requests zero/sets death. Count 1610612736 produces bonus -1073741924 and divisor
zero, faulting before the resistance read or any later request. These are arithmetic fixtures,
not claims about actual creation reachability. All existing ordinary numeric results stay
unchanged; post-fault recovery is outside the contract. Signed quotient overflow has the same
machine division limitation. FULL1 coverage review prompted a distinct supplied-count check:
515396075 produces wrapped bonus -126, product -504, scaled result -5 and divisor -1;
`INT_MIN / -1` faults at the same division instruction before the resistance read or writes
(CBL-V15(d)). This establishes reachability within the supplied signed-count domain, not that
ordinary monster creation produces the count.

## Gap assessment

Assessment date 2026-10-15; reused attempts keep their original dates and limits.

| Question / missing boundary | Evidence assessed / attempted | Result and smallest remaining check |
|---|---|---|
| CBL-Q01 monster-condition classification | 2026-10-13 consumer and setter leads; 2026-10-15 complete `0x22A80` read plus narrow setter slices at `0x24FC5`/`0x252D3` in functions `0x24E70`/`0x251F0` | Both slices OR the consumed bit, but their caller/classification/creation eligibility is unread; no monster-kind label follows. Trace their actual dispatch/caller inputs or attribute condition/divisor on observed spawns |
| Pair count/order (settled; folded into CBL-R01) | Current IEP-R01/R03/R05/R07, actual named ItemStatCost cells and callback table | Supplied notification identity/count/order supported; no new table or automatic-notification claim |
| CBL-Q03 relative order | IEP traversal and historical unperformed CBL-O01/O03 | Static newest-first available; no attributed gameplay sequence exists |
| CBL-Q04 overlay count and missile conditions | Existing LSE-Q03 search dated 2026-10-14 found readers only; 2026-10-15 shared GDF/overlay contracts, contact transfer and #10623 leaf read | Definition fields or another consumer do not establish collection publication or original rows; incoming conditions not assumed clear. Trace loader publication and owning missile condition producers |
| CBL-Q05 owner | Historical melee/missile chains, current MHD-R13/MHD-R26, direct `0x3AE10` hand-off | Resolved owner passed as attacker; wider lookup/contact eligibility remains MHD-owned. Consume completed owner contract rather than infer live identity from a stack argument |
| Build equivalence (evidence-build policy) | Current evidence-build policy and original-profile records | Closed: no other build is claimed |
| CBL-Q07 other non-queued routes | Current DPO-R13a/R23 and IEP-R14 retaliation suppression | Supplied flag/event behavior fixed, not all call sites' reachability. Complete each producer under its owner |
| CBL-Q08 messages | Existing local setter/overlay records, current LSE-Q04/NRF-Q17/STL networking limits | No sender/body/delivery contract; do not promote a request to delivery |
| CBL-Q09 overlay helpers | Compatible LSE-R14/STL-R06a read, retained STL-R06b/STL-R06c/NRF-Q17 gates | Local helper gap closes, original stored integration does not; complete owning attach/write/update contracts |
| CBL-Q10 replacing write | Historical partial #10887/`0x3A280` read; current NRF-Q20/STL lifecycle contracts | Local value/order eligible; propagation/callees/notification remain unverified |
| CBL-Q11 notifications | Current IEP-R03/R04 and IEP-Q01/IEP-Q03 | Supplied notification available; complete actual production/traversal order |
| CBL-Q12 observations | 2026-10-13 record: no gameplay traces; read-only slot/constants only | Authorized migration excludes gameplay; no runtime label/observed flag changes |
| Data support | Complete 2026-10-13 named cells above; current GDF names/classes/records | Reused exact cells and original coverage; new hashes do not identify historical bytes; no new unsupported named table value |

## Observation procedures

These procedures retain CBL-O01 to CBL-O03 and their original expected readings.
No observation has been performed. The spec keeps a summary and links here for reproduction.
Expected stored effects are observation targets, not established setter/overlay integration.
Record the supplied/getter life and the requested value separately from any actual read-back;
only matching attributed assertions can close the relevant observation or integration question.

### Recorded observation vectors and reproduction procedure

Priorities: the order of the extra draw relative to leech (CBL-O01), the overlay on a proc with a
positive amount (CBL-O02), the player-count change of the divisor (CBL-O03). Observation vectors
are not hand-executable: they need an owned 1.13c game and a way to record the named function
entries and unit fields; they confirm identity and order, the executable vectors confirm values.

- **CBL-O01** (CBL-R01, CBL-R02a, CBL-R02b, CBL-R03). What to measure: that one attacker draw
  happens inside apply before the leech draw, once per hit, on melee and on missile hits.
  Expected: on a melee hit with stat 136 = 100 the attacker's seed `{Unit +0x20, +0x24}` read at
  the callback's entry and again at its return differs by exactly one NRF-R01 step (the hit roll
  and damage-build draws happen before the entry and are not in this span), the callback is
  entered once per hit, and the defender's stat 6 at the return equals `life / d` less than at the
  entry, before the hit's own damage is subtracted; the callback is entered with `edx = 5` on a
  melee swing and `edx = 6` on an arrow or bolt. Procedure: any class
  at any level, Normal difficulty, single player, carrying one item with "+100% chance of crushing
  blow" (stat 136 = 100; a lower value works but needs repeats) and no life leech; stand next to a
  plain Act I monster outside town; record the attacker's seed and the monster's stat 6; perform
  one melee attack with the basic attack skill; at the callback entry (D2Game `0xACC30`) record
  the arguments (owner, other, record, `0x880000`, 0), the event id and the return value, and at
  the return record the monster's stat 6; then repeat with a bow or crossbow equipped and a normal
  arrow.
- **CBL-O02** (CBL-R08). What to measure: the overlay request on a proc with `q > 0`. Expected:
  after the callback returns 1 with `q > 0`, the defender holds an attached stat list flagged
  `0x80` carrying stat 178 = `0x93`, and D2Common `0x32C10` was entered with `(defender, 0x93, 0)`.
  Procedure: as CBL-O01, reading the defender's `Unit +0x5C` list and its attached lists (or the
  entry of `0x32C10`) right after the hit; repeat against a monster whose stat 36 is 100 or more
  (a physical-immune monster in Nightmare or Hell) and whose life is below `0x400000` raw
  (16384.0 HP, so that `q <= 0x100000` and CBL-V11's remainder cannot occur) and confirm no entry
  of `0x32C10` and no stat 178 write while the seed still advanced.
- **CBL-O03** (CBL-R04a, CBL-R05). What to measure: the divisor against player count and the
  `boss` flag. Expected: a plain monster at `/players 1` loses `life / 4` (CBL-V02); the same
  monster class spawned at `/players 8` loses `life / 18` (CBL-V08 (d)); a unique or champion
  monster of the same non-hireling class, with its `MonStats.boss` false, at `/players 1` loses `life / 8` when `MonsterData +0x16` bit 1 reads set for it and
  `life / 4` when it reads clear (the reading of the bit is recorded with the loss; the pair
  establishes the observed instance's condition/divisor relationship, not every producer or
  classification rule in CBL-R04b, which is why this expectation is not a single value).
  Procedure: as CBL-O01 in a single-player game;
  set `/players 1`, enter a new area, hit a plain monster once and read its stat 6 before and
  after the callback; leave, set `/players 8`, enter a new area so that stat 100 = 8 on new
  spawns, repeat; then repeat on a champion and on a unique monster, reading `MonsterData +0x16`
  on each.

## Semantic-to-binary mapping

The main spec uses logical operations and fields. These names describe behavior and do not
prescribe an API, programming language, storage layout or binary-compatible implementation.
The following mappings retain the original encodings for source review and historical fixtures.

| Semantic term | Original representation | Rules / vectors |
|---|---|---|
| `applyMode = queuedMelee` / `nonQueued` | apply flag argument 0 / any non-zero value (missile contact uses 1) | CBL-R02a, CBL-V13 |
| `result.hit` | result word mask `0x1` | CBL-R02a, CBL-V13 |
| `result.suppressEvents` | result word mask `0x20` | CBL-R02a, CBL-V13 |
| `result.death` | result word mask `0x2` | CBL-R07, CBL-V09 |
| `itemEventsEnabled` | damage flags mask `0x20` | CBL-R02a, CBL-R02b, CBL-V13 |
| `suppressAttackerMissileEvent` | damage flags mask `0x80` | CBL-R02a, CBL-R02b, CBL-V13 |
| `missile.enablesItemEvents` / `missile.suppressesAttackerEvent` | missile data dword bits 0 / 1 at `MissileData +0x14` | CBL-R02b |
| `monsterDivisorFlag` | word `MonsterData +0x16`, mask `0x2`; kind meaning unresolved | CBL-R04a, CBL-R04b, CBL-Q01 |
| `boss` | `MonStatsRecord +0x0C`, bit 6, mask `0x40` | CBL-R04a, CBL-V04 |
| `(stat id, layer)` identity | packed key `(id << 16) + layer`; stat 136/layer 0 is `0x880000`, stat 6/layer 0 is `0x60000` | CBL-R01, CBL-R07, CBL-V12 |
| Crushing Blow handler | D2Game callback slot 16, RVA `0xACC30` | CBL-R01 to CBL-R08, CBL-V12 |
| Overlay request `(defender,147)` | D2Common #10689/RVA `0x32C10`, stack `(defender,0x93,0)`; third argument is not read by the helper and is not a required semantic parameter | CBL-R08, CBL-V01 to CBL-V11 |
| Replacing unit-stat request | D2Common ordinal 10887, RVA `0x3A740` to `0x3A280`; complete stored effects remain gated | CBL-R07 |
| Overlay contribution / pending overlay update | attached stat-list flag `0x80`, base-list flag `0x100`, update helper RVA `0x665C0` | CBL-R08, CBL-Q09 |
| V13 (a)/(b) logical flags | flags `0x20`, result `0x0001`; apply flag 0 / 1 | CBL-V13 |
| V13 (c) logical flags | flags `0xA0`, result `0x0001`; apply flag 1 | CBL-V13 |
| V13 (d) logical flags | flags `0x00`, result `0x0001`; apply flag 1 | CBL-V13 |
| V13 (e) logical result | result `0x0021` (hit plus suppression) / `0x0000` (no hit) | CBL-V13 |
| Default vector result / V09 death result | `0x0001` (hit) / `0x0003` (hit plus death) | CBL-V01 to CBL-V11 |

## Historical observations and reviews

2026-10-13 read-only relocated memory checks found callback slots 15/16/17 at D2Game
`0x1077E4` equal to RVAs `0xAD410`/`0xACC30`/`0xACBA0`, and the nine constants at `0xFB614`
equal to the image. These identify pointer/constants only, not executed gameplay. No attack,
armed capture or matching gameplay observation occurred; no vector is runtime-observed.
Existing unrelated capture state was left untouched. Exact session/tool details and raw
historical annotations are privately preserved.

Historical reader run 1 reported 11 GUESS/15 GAP; run 2 reported 10 GUESS/7 GAP. Recorded
outcomes added explicit dispatch/registration inputs, flags, seeds, returned values and
helper questions. Neither historical report had a third run or establishes review completion
for this migrated contract. Their complete reports/original contracts are privately archived;
the relevant unresolved interpretations and corrections remain public above.

Other dated correction rationale retained from 2026-10-13: callback labels were not identity
proof; table source was corrected from erroneous `0x1B614` to `0xFB614`; boss missing-record
behavior and signed nonpositive-life branch were made explicit; the separate condition was
kept neutral instead of called champion/unique; R02/R04 were split by evidence label.
Crushing Blow and Open Wounds remain separate because this callback requests immediate life
replacement, while the latter creates a timed state consumed by regeneration.

## Migration record

2026-10-15 full migration retains the original file boundary, every existing CBL rule/question/
vector/observation identity, all original ordinary-case numeric values and seed pairs. Two
families add local flag-transfer and extreme-count discrimination; CBL-V01 adds null guards.
Original source/table dates, profile limits, correction rationale and original documents are
preserved publicly or in a SHA-256-verified private archive. No game assets or raw data tables
are published.

Labels changed: CBL-R02b asm-only -> verified static for supplied flag transfer; CBL-R08
asm-only -> verified static for local overlay operation/return through LSE-R14, with stored
integration gated. CBL-R07 retains verified static only for local request/flag behavior;
its unsupported stored effects are now explicitly gated. CBL-R04b remains unresolved. New
CBL-Q10/Q11/Q12 distinguish write integration, notification integration and observations.
Document status remains draft.

Self-review, mechanical checks and fresh reader outcomes are recorded as they run. No
implementation comparison, implementation test, gameplay action or approval is asserted.

### Migration self-review

2026-10-15 mechanical document check: 0 errors, 0 warnings. Preservation review retained all
12 existing rule/heading identities, 9 question identities, 13 vector identities and 3
observation identities; additions are CBL-Q10 to CBL-Q12 and CBL-V14/CBL-V15. Arithmetic
recomputation checked 21 result records and four seed examples against the existing ordinary
results and new explicit wrap/fault cases. All original ordinary numeric outputs remain.
Public/private separation, dependency ownership and every open-question disposition were
reviewed. No implementation test or gameplay observation ran. Fresh blind-reader review of
this resulting contract remains pending and is separate from the two historical reports.

## Blind-implementer FULL1

Fresh full review of the migrated contract checked 15 vector families and 40 explicit cases.
Counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 3, GATED 10, COVERAGE 4, HARDEST 3. This was a full
review, distinct from the historical reports above. The following dispositions apply to every
entry; fixture additions require a fresh full run on the resulting contract.

| List / entry | Outcome |
|---|---|
| MISMATCH / GUESS / GAP | No entries; no changes required |
| DISCRIM: V14 symmetric missile conditions permit swapped mapping | Added CBL-V14(d), incoming conditions both false and missile enabled true/suppression false; output true/false rejects the swap |
| DISCRIM: V12 positive notifications permit positive-only registration | No local duplicate: IEP-R03 is explicitly consumed, and IEP-V02(e) starts fresh with newTotal -5 and requires an entry. CBL-V12 points to that exact owner's discriminating case; the bound event pair is independently tested locally |
| DISCRIM: death assignment could clear an existing true value | Added CBL-V09(e), initial death true and positive requested life 3; flag must remain true |
| GATED: CBL-Q01/CBL-R04b classification | No change: original producer/classification remains unresolved; supplied condition effect is independently available |
| GATED: CBL-Q03 proc observation | No change: static traversal verified, attributed gameplay sequence pending |
| GATED: CBL-Q04 count/condition producers | No change: original publication/rows and missile producers remain outside-boundary; supplied inputs eligible |
| GATED: CBL-Q05 owner/contact | No change: full MHD integration stays its owner; supplied attacker callback remains executable |
| GATED: CBL-Q07 other routes | No change: producer reachability not inferred from shared non-queued branch |
| GATED: CBL-Q08 messages | No change: local requests do not establish sender/player notification/delivery |
| GATED: CBL-Q09 overlay stored integration | No change: LSE-R14 local requests retain STL-R06b/STL-R06c/NRF-Q17 integration gates |
| GATED: CBL-Q10 replacing write | No change: requested life/order/death flag verified; actual storage/propagation remains NRF-Q20 |
| GATED: CBL-Q11 notification | No change: supplied IEP-R03 notification does not settle IEP-Q01/IEP-Q03 production/order |
| GATED: CBL-Q12 gameplay | No change: no gameplay performed or runtime label promoted |
| COVERAGE: separate divisor condition true/boss false | Added CBL-V02(b), divisor 8/amount 32000/requested life 224000; this is a result-distinct supplied predicate branch |
| COVERAGE: other defender types | No change: the retained divisor-4 recipe produces the same amount/request as CBL-V02's existing default case; no additional kind-specific effect or draw occurs |
| COVERAGE: signed quotient overflow | Independently checked reachable count 515396075 and added CBL-V15(d) with INT_MIN/divisor -1; exact overflow fault prefix differs from the zero-divisor case |
| COVERAGE: absent result on lethal value | Added CBL-V09(f), life 0/result absent; life request remains while result-flag write is absent |
| HARDEST: V15 wrapped count/divisor/life chain | Existing formulas/precision rules remain unchanged; independent numeric recomputation agrees. Added only the distinct quotient-overflow fixture, retaining the fault-prefix boundary |
| HARDEST: V01 guards/LCG/threshold | No change: exact guard order, no-draw cases, seeds and strict comparison are stated and recomputed without guessing or mismatch |
| HARDEST: V12–V14 registration/dispatch/conditions | Added asymmetric flag fixture and exact negative-registration owner reference; no copying of generic traversal or change to supplied boundaries |

All old cases and numeric outputs remain. Added cases are CBL-V02(b), CBL-V09(e/f),
CBL-V14(d), CBL-V15(d). No rule behavior, ownership, evidence label or gate boundary changed.

## Blind-implementer FULL2

Fresh full review checked 15 vector families and 45 explicit cases. Counts: MISMATCH 0,
GUESS 0, GAP 0, DISCRIM 2, GATED 10, COVERAGE 1, HARDEST 3. The reader's process could not
access the private prompt/report path; the full method was supplied directly, with no prior
findings. Repository access worked and the review scope remained full. The exact report was
retained privately from the reader's final response.

| List / entry | Outcome |
|---|---|
| MISMATCH / GUESS / GAP | No entries; no changes required |
| DISCRIM: player count scaling could be incorrectly applied | Added CBL-V06(d): player/count 3/resistance 50 retains divisor 10 and life request 121600, rejecting a scaled player divisor of 20 |
| DISCRIM: queued dispatch could incorrectly require enabled/unsuppressed record | Added CBL-V13(f): queuedMelee with enabled false/suppression true still dispatches attacker 5 then defender 1, then clamp |
| GATED: CBL-Q01 assignment | No change: producer/classification unresolved; supplied condition is executable |
| GATED: CBL-Q03 callback-order observation | No change: static order available, attributed gameplay pending |
| GATED: CBL-Q04 count/missile producers | No change: original collection binding and condition producers stay outside-boundary |
| GATED: CBL-Q05 owner/eligibility | No change: complete original contact integration stays MHD-owned |
| GATED: CBL-Q07 other non-queued callers | No change: local route semantics do not prove all producer reachability |
| GATED: CBL-Q08 delivery | No change: sender/player notification/client delivery not inferred from requests |
| GATED: CBL-Q09 overlay storage | No change: attach/list-write/update integration retains its existing gates |
| GATED: CBL-Q10 stat storage | No change: replacing-write effects remain NRF-Q20; local request remains eligible |
| GATED: CBL-Q11 automatic notification | No change: IEP-Q01/IEP-Q03 production/order remain distinct from supplied notification |
| GATED: CBL-Q12 observations | No change: no gameplay performed or runtime label promoted |
| COVERAGE: other defender type | No change: retained divisor 4 has the same local amount/request recipe and no extra effect/draw as the existing default CBL-V02; no result-distinct branch requires another fixture |
| HARDEST: CBL-V15(c) zero divisor | No change: exact wrapped inputs, completed draw/life-read prefix and division-fault boundary are already explicit and independently recomputed |
| HARDEST: CBL-V15(d) quotient overflow | No change: reachable supplied count, divisor -1/INT_MIN and no later operations are explicitly stated; no recovery contract is invented |
| HARDEST: CBL-V13(e) outer gates across routes | No change: four fresh cases enumerate both modes and both result failures; inputs/seeds/outputs are explicit and no ambiguity/mismatch was reported |

All previous cases/results remain. Only CBL-V06(d) and CBL-V13(f) were added; CBL-V02 wording
now refers to semantic overlay defender/id because the unused binary argument belongs in the
companion mapping. No rule behavior, ownership, evidence label or gate boundary changed.
Fresh FULL3 review is required for the added fixtures.

## Blind-implementer FULL3

Fresh full review checked 15 families and 47 explicit cases. Counts: MISMATCH 0, GUESS 0,
GAP 0, DISCRIM 3, GATED 10, COVERAGE 5, HARDEST 3. Every entry is dispositioned below.
Resulting changes affect fixtures/assertions only; no rule behavior, ownership, label or gate
changed. Under the review stopping rule, a fresh delta check can review these changes after
this third mismatch-free full run and its complete triage.

| List / entry | Outcome |
|---|---|
| MISMATCH / GUESS / GAP | No entries; no changes required |
| DISCRIM: player event 6 could omit doubling | Added CBL-V06(e): life 128000/resistance 0/event 6 gives divisor 20, amount 6400, request 121600; monster-only doubling would request 115200 |
| DISCRIM: bonus-table entry 2 untested | Added CBL-V08(e): count 2 gives bonus 50/divisor 6/amount 42666/request 213334 |
| DISCRIM: outer result gate could be rechecked after attacker callback | Added CBL-V13(g): attacker clears shared hit, yet defender event 1 observes hit false and is still called before clamp |
| GATED: CBL-Q01 classification | No change: producer/classification unresolved; explicit supplied condition effect remains eligible |
| GATED: CBL-Q03 callback observation | No change: attributed gameplay sequence pending; static traversal retained |
| GATED: CBL-Q04 count/missile producers | No change: original count binding and producer identities outside-boundary |
| GATED: CBL-Q05 owner | No change: complete owner/contact integration remains MHD-owned |
| GATED: CBL-Q07 other non-queued callers | No change: producer reachability is not inferred from supplied route semantics |
| GATED: CBL-Q08 delivery | No change: no sender/player notification/delivery established from requests |
| GATED: CBL-Q09 overlay storage | No change: stored attach/list-write/update integration retains its gates |
| GATED: CBL-Q10 replacing storage | No change: NRF-Q20 retains stored effects; request remains eligible |
| GATED: CBL-Q11 notifications | No change: automatic production/order remains IEP-Q01/IEP-Q03 |
| GATED: CBL-Q12 gameplay | No change: no observation or runtime promotion |
| COVERAGE: table counts 2/4–7 | Count 2 added as CBL-V08(e), exercising the first nonzero table bonus. No exhaustive extra cases for fixed entries 4–7: the explicit nine-value binding uses the same indexed-table branch; counts 1/2/3/8 already sample that branch and the table/formula boundary |
| COVERAGE: event 6 for player/other types | Player added as CBL-V06(e). Other type is the existing divisor-4 recipe followed by the same unconditional doubling as CBL-V03; no extra kind-specific effect/draw or new arithmetic branch |
| COVERAGE: shared result mutation | Added CBL-V13(g) with the exact mutation and ordered log; defender still executes on the already-entered route |
| COVERAGE: missing/out-of-range class and other defender type | No change: missing/out-of-range boss record selects the existing false-predicate branch, supported by GDF-R19a; supplied false is covered by CBL-V02. Other type retains the same divisor-4/no-extra-effect recipe and draw count, rather than introducing a distinct calculation |
| COVERAGE: extreme count plus event 6; negative amount plus positive resistance | No change: event doubling is the unconditional low32 operation already tested across monster/player/hireling, while wrapped divisor/life and both fault paths are CBL-V15. No count- or divisor-sign branch is added by event 6. For negative amount/positive resistance, the positive-resistance branch is CBL-V06(a), and the signed narrow product/truncation/wrap recipe belongs to NRF-R05 and is discriminated by NRF-V04(a)/(k); negative-amount overlay suppression is CBL-V09(d)/CBL-V15(a). The combination adds no uncovered algorithm branch or independent operation |
| HARDEST: CBL-V15(c/d) fault prefixes | No change: zero-divisor and INT_MIN/-1 cases give exact completed draw/life-read prefixes and exclude later operations/recovery; independent arithmetic checks agree |
| HARDEST: CBL-R04a/R05 count chain | Existing CBL-V04/V05/V07/V08/V15 discriminate type/count/formula/wrap recipes; added CBL-V08(e) tests first nonzero table bonus and CBL-V06(e) player doubling. No rule change needed |
| HARDEST: CBL-R02a/R03 mutable dispatch and callback draw | Added CBL-V13(g) for mutable shared-record order. Callback guards/chance draw remain separately executable in CBL-V01; supplied dispatch recorders do not claim to execute those callback effects |
| Additional root finding: V13 reachability summary omitted (f) | Corrected fixture assertion to include (a)/(b)/(f)/(g) and explicitly exclude (c)/(d)/(e). The stated mutation exception now limits the shared unchanged-record assertion in (g) |

Changed-case list for the fresh delta check: new CBL-V06(e), CBL-V08(e), CBL-V13(g);
CBL-V13 input defaults, shared output/reachability summary and Discriminates assertion updated
only to describe the explicit mutation and existing queued case accurately. All prior cases
and numeric outputs remain. No further case was added for equivalent fixed-table samples or
cross-products of already specified arithmetic operations.

## Blind-implementer DELTA4

The fresh delta reader recomputed all three added cases, CBL-V06(e), CBL-V08(e) and
CBL-V13(g), and the changed V13 defaults, mutation exception and reachability assertions.
It received the exact changed-case list and the findings those changes address. Reported
counts: MISMATCH 0, GUESS 0, GAP 0, DISCRIM 0; changed cases checked: 3. No entry needs
another edit or gate. This was a delta check after the third full run returned MISMATCH 0
and all its entries received the dispositions above; no rule, ownership or gate changed.

The resulting contract retains 15 vector families and 50 explicit cases. The applicable
blind-reader stopping condition is met. Original classification, producer, stored-effect,
notification and gameplay gaps remain visible; the document remains draft and awaits user
review. This review does not establish gameplay execution, complete parity or approval.

## Specification review spot-check, 2026-10-15

An independent review re-read the original D2Game callback 0xACC30 from entry to its returns,
the hireling-class predicate D2Common 0x1A2D0 (classes 271, 338, 359, 560 and 561), the
missile-contact transfer 0x3AE10 (runtime-condition bit 0 ORs damage flag 32 and bit 1 ORs 128,
with no clearing), the apply-stage dispatch in 0xDE0C0 (result mask 32 clear and hit set;
queued 5 then 1; non-queued requires flag 32, skips 6 under flag 128, then 2), and the
nine-entry bonus table at RVA 0xFB614 from the manifest-matching image, `{0,0,50,100,150,200,
250,300,350}`. These agree with CBL-R02a, CBL-R02b, CBL-R03, CBL-R04a, CBL-R05 to CBL-R08.
Vectors CBL-V01 to CBL-V11 and CBL-V15 were recomputed from the rules, including the four
wrapped/fault count cases and the NRF-R01 seed pairs; no mismatch was found.

Hygiene correction: two settled entries were removed from the unconsumed contract's
open-question table: the registration pair count/order question (answered by CBL-R01 at the
supplied notification boundary) and the build-equivalence entry (the evidence-build policy,
not an open question).
Their gap-assessment rows above remain as dated assessment records. No rule, vector, label or
gate changed.
