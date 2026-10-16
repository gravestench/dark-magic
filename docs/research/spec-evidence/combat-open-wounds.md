# Open Wounds: evidence

Contract: [combat-open-wounds](../../specs/combat-open-wounds.md).
Original artifacts: [source manifest](../source-manifest.md), `LOD113C-D2GAME`,
`LOD113C-D2COMMON` and `LOD113C-FOG`. Function descriptions below are analyst labels, not
exported-name guarantees. RVAs are module-relative and require matching fingerprints.

## Evidence scope and dates

The original 2026-10-13 reading covers the callback, its request construction, the generic
state applier's bounded paths, monster/player life steps and their dispatch anchors. These
are reused at their original dates and limits; citation relocation is not a fresh verification.
The complete prior document, internal matching details, binary fixture encodings, old search
commands and review text were preserved privately with SHA-256 before migration.
This public record retains original binary locators, interpretations and correction rationale.

The manifest identifies current analysis specimens. It does not independently fingerprint
every historically loaded DLL or prove pristine archives. Gameplay was not attempted in the
historical pass or this migration. No vector is runtime-confirmed; no mismatch observation exists.

## Dispatch and function evidence

| Artifact | Function / table RVA | Export ordinal / import thunk | Analyst description | Supported local rules | Method and boundary |
|---|---|---|---|---|---|
| `LOD113C-D2GAME` | `0xAD410` | n/a | Open Wounds callback | R02-R05 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0xAC940` | n/a | Piecewise level curve | R03 | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2GAME` | `0x22A80` | n/a | Monster-data condition mask | R04a | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0xA00B0` | n/a | Generic timed-state request applier | R06-R08 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0x1320` | n/a | MonStats-linked extended-record lookup | R06a | 2026-10-13 disassembly reading; decoding limits below |
| `LOD113C-D2GAME` | `0x76740` | n/a | Monster regeneration/life/death tail | R09a-R09d, R10 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0x79B10`, `0x77CB0` | n/a | Player event handler and life helper | R11a, R11b | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0xDE0C0` | n/a | Hit application | R01, R12, section 2 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0x37030` | n/a | Item/unit-event dispatcher | R01, R10 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0xA2C40`, `0xDED30`, `0x3AE10` | n/a | Basic skill do, queued apply and missile apply caller | section 2 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0x2E050` | n/a | Server frame event call | section 2 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0x8F070` | n/a | Frame-event group runner | section 2, R13 | 2026-10-13 disassembly reading; decoding limits below |
| `LOD113C-D2GAME` | `0x8EBC0`, `0x8ECC0` | n/a | Monster/player queue runners | section 2, R13 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0x769A0` | n/a | Monster unit-event dispatch | section 2 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0xFBC14`, `0xFBB74`, `0xFA3DC` | n/a | Monster/player unit-event tables and unit-type queue map | section 2 | 2026-10-13 original-image/import record |
| `LOD113C-D2GAME` | `0x1077A8` | n/a | Item-event callback table | R01 | 2026-10-13 original-image/import record |
| `LOD113C-D2GAME` | `0x8EA80` | n/a | Frame-event scheduler | R07a, R08, R09a, R11a, R12, R13 | 2026-10-13 disassembly reading; decoding limits below |
| `LOD113C-D2GAME` | `0x8E680`, `0x8E6F0` | n/a | Frame-event allocation/removal | R09c, R13 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME / LOD113C-D2COMMON` | `0x74870`, `0x77820`, `0x3ADD0` | D2Common 11093 (thunk `0xAA2A`) | Monster/player expiry handlers and shared expiry walk | R08, R13, Q09 | 2026-10-13 disassembly reading |
| `LOD113C-D2GAME` | `0x9F7A0` | n/a | Generic state removal callback | R08 | 2026-10-13 disassembly reading; one or more noted callees not read |
| `LOD113C-D2GAME` | `0xE03A0` | n/a | Registered-unit lookup | R10 | 2026-10-13 disassembly reading; byte decoding not checked; decoding limits below |
| `LOD113C-D2GAME` | `0xDEEE0` | n/a | Death request target; body not read | R10 | 2026-10-13 locator only; body not read |
| `LOD113C-D2GAME` | `0x14D0` | n/a | Signed multiply/divide helper | R06b (NRF-R05) | 2026-10-13 disassembly reading |
| `LOD113C-D2COMMON` | `0x39430` | D2Common 10910 (thunk `0xA802`) | Effective unit-stat reader wrapper | R02 | 2026-10-13 disassembly reading |
| `LOD113C-D2COMMON` | `0x38B70` | D2Common 10973 (thunk `0xA42A`) | Effective unit-stat reader and selected-array read | R03, R06b, R09a, R11a, R12 | 2026-10-13 disassembly reading |
| `LOD113C-D2COMMON` | `0x390F0` | D2Common 10913 (thunk `0xA3A0`) | Maximum-life stat-7 reader | R09a, R11a | 2026-10-13 disassembly reading |
| `LOD113C-D2COMMON` | `0x3A740` | D2Common 10887 (thunk `0xA472`) | Unit replacing-write wrapper | R09a, R11a | 2026-10-13 disassembly reading |
| `LOD113C-D2COMMON` | `0x33D80` | D2Common 10372 (thunk `0xB078`) | Life-state mask predicate wrapper/helper | R09b | 2026-10-13 disassembly reading; decoding limits below |
| `LOD113C-D2COMMON` | `0x38C20` | D2Common 10587 (thunk `0xA3E2`) | Own-array stat reader wrapper/core | R09b | 2026-10-13 disassembly reading; one or more noted callees not read |
| `LOD113C-D2COMMON` | `0x33B40` | D2Common 10033 (thunk `0xAE5C`) | Curse-mask membership test | R06a | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2COMMON` | `0x33A70` | D2Common 10226 (thunk `0xAE9E`) | Exp-mask membership test | R08 | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2COMMON` | `0x5F6E0`, `0x5DAD0` | n/a | States loader and flag-mask builder | R06a, R08, R09b | 2026-10-13 disassembly reading; decoding limits below |
| `LOD113C-D2COMMON` | `0x38520` | D2Common 10871 (thunk `0xA5D4`) | Find state list | R07a, R10, R11b | 2026-10-13 disassembly reading |
| `LOD113C-D2COMMON` | `0x33CD0` | D2Common 10494 (thunk `0xA68E`) | Has active state | R06b, R09a, R10 | 2026-10-13 disassembly reading |
| `LOD113C-D2COMMON` | `0x34510` | D2Common 10945 (thunk `0xA3CA`) | State request routing | R08, R10 | 2026-10-13 disassembly reading |
| `LOD113C-D2COMMON` | `0x665C0` | D2Common 10158 (thunk `0xA4DE`) | Unit update-mark request target | R08 | 2026-10-13 locator only; body not read |
| `LOD113C-D2COMMON` | `0x38100` | D2Common 11013 (thunk `0xA6A0`) | Stat-list allocation | R08 | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2COMMON` | `0x38060` | D2Common 10785 (thunk `0xA5B6`) | Set list state identity | R08 | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2COMMON` | `0x37D60`, `0x37D40` | D2Common 10725 (thunk `0xAE3E`), 11081 (thunk `0xAE50`) | Set list skill and level | R08 | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2COMMON` | `0x38040`, `0x37D50`, `0x37D30` | D2Common 11132 (thunk `0xAEAA`), 10555 (thunk `0xAE62`), 10515 (thunk `0xAEBC`) | Read list state, skill and level | R07a | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2COMMON` | `0x38020` | D2Common 10161 (thunk `0xA778`) | Set list expiry | R07a | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2COMMON` | `0x3A7B0` | D2Common 10188 (thunk `0xA6E2`) | List-stat replacing-write wrapper | R08 | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2COMMON` | `0x3A7D0` | D2Common 10807 (thunk `0xA60A`) | Full/filtered attach target | R08 | 2026-10-13 disassembly reading |
| `LOD113C-D2COMMON` | `0x37D70` | D2Common 10911 (thunk `0xA75A`) | Set list callback | R08 | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2COMMON` | `0x3A5F0`, `0x3AD40` | D2Common 11111 (thunk `0xABB6`), 11108 (thunk `0xABAA`) | Detach/free list request wrappers | R07b, R11b | 2026-10-13 reconstruction checked against original bytes; one or more noted callees not read |
| `LOD113C-D2COMMON` | `0x384F0` | D2Common 10195 (thunk `0xA7B4`) | Curse-list lookup target; body not read | R06b | 2026-10-13 locator only; body not read |
| `LOD113C-D2COMMON` | `0x38080`, `0x380E0` | D2Common 10473 (thunk `0xA84A`), 11131 (thunk `0xA7FC`) | Read list owner id/type | R10 | 2026-10-13 reconstruction checked against original bytes |
| `LOD113C-D2COMMON` | `0x2FE10`, `0x3C390` | D2Common 10331 (thunk `0xA478`), 10057 (thunk `0xA808`) | Unit-room lookup and room-town test | R09a | 2026-10-13 disassembly reading |
| `LOD113C-FOG` | `0x26798` | Fog 10259 | Bit-mask data export, GDA-R16 | R06a, R08, R09b | 2026-10-13 import record; current GDA-R16 data-export resolution |

The ordinal columns distinguish D2Game import-thunk RVAs from resolved D2Common functions.
The callback table stores a pointer, rather than exporting an original callback name.
`n/a` means no export ordinal is asserted. Unrelated analyst names in the original investigation were treated only as search labels;
control flow, table identity and output writes establish the descriptions published here.

### Original dispatch chains

1. Melee: D2Game skill do table `0x1074A8`, slot 1 -> `0xA2C40` -> build `0xDDDE0` and queued
   apply `0xDED30` (alternate `0xA20C0`) -> `0xDE0C0` -> dispatcher `0x37030` -> callback table
   `0x1077A8`, slot 15 -> `0xAD410` -> applier `0xA00B0`. DPO-R13a/DPO-R23 and IEP-R07 hold
   the consumed semantic contracts. Result bit 1 and event suppression bit `0x20` determine
   attacker events; queued flag 0 selects event 5. The proc precedes clamp/leech/life/status.
2. Missile: movement/contact roots MMC -> `0x3EDE0` -> `0x3AFF0` -> `0x3AE10` -> apply
   `0xDE0C0` with flag 1. Record flag `0x20` must be set and `0x80` clear for attacker event 6.
   The forwarded owner is the event owner and seed/level source, not the missile.
3. Frame roots: D2Game `0x2E1E0` / `0x526F0` -> `0x2E050` -> timer runner `0x8F070`.
   Monster queue runner `0x8EBC0` -> `0x769A0` -> unit-event table `0xFBC14`, slot 3 -> `0x76740`.
   Player queue runner `0x8ECC0` -> table `0xFBB74`, slot 3 -> `0x79B10` -> `0x77CB0`.
   `0x79B10` subsequently invokes stamina `0x77A50` and mana `0x77950` (outside local life step).
4. Scheduling: `0x8EA80` uses unit-type queue map `0xFA3DC={0,1,3,2,4,-1}`. Candidate group
   order is missiles, players, monsters, objects, items; an immediate list precedes each ordinary
   bucket. Monster dispatch map `0x76A00` routes type 3 directly to its table. Partial image checks
   do not establish all queue mutation/order behavior; STL-R02a/STL-Q07 retain that gate.
5. Expiry type 12: monster slot 12 -> `0x74870`, player slot 12 -> `0x77820`; both call D2Common
   #11093 at `0x3ADD0` with unit/current frame. STL-R04 supplies checked traversal; actual free
   #11108 / `0x3AD40` -> `0x3A9F0`, detach #11111 / `0x3A5F0` -> `0x3A3B0`, attach #10807 /
   `0x3A7D0` and writer #10188 / `0x3A7B0` -> `0x3A280` retain STL-Q07 integration gates.
6. Generic removal callback `0x9F7A0` requests unit-timer cancellation `0x37110`, StayDeath
   #11061 / `0x33AB0`, state off and refreshes. STL-R11/STL-R11a verify request order. These
   unit timers are distinct from frame-event removal `0x8E6F0`; shared numeric kind values do
   not make them interchangeable.

### Semantic field and operation correspondence

The following layout facts support independent original-image examination; they impose no
storage or API on a new implementation.

| Semantic value | Original evidence encoding | Consumers |
|---|---|---|
| Owner/defender type, class, id, mode | Unit +0/+4/+0xC/+0x10 | OWD-R03/R04/R06/R08/R10/R11 |
| Attacker seed | Unit +0x20/+0x24 | OWD-R02 |
| Monster conditions specialA / specialB | MonsterData word +0x16 masks `0x4` / `0x8`; either passes mask `0xC` at Game `0x22A80` | OWD-R04a; producers OWD-Q01 |
| `aliveExclusionCondition` | Unit +0xC6 bit 0 | OWD-R11a; STL-R11 |
| Target game / allocation tag | Unit +0x80 / +0x08 | OWD-R06/R08; allocator's first argument differs from CSA's game allocation pool |
| Death auxiliary | Unit +0xB0 | OWD-R10 clear; wider meaning unclaimed |
| Base stat list / active states | Unit +0x5C; list +0x58 with list flag bit 31 | OWD-R09b |
| Own/effective stat arrays | List +0x24 / +0x48, selected by extended condition | RLI-R02a/RLI-R02b; reads do not traverse attachments |
| List owner/state/skill/level/expiry/callback | +0x08/+0x0C, +0x14, +0x1C, +0x20, +0x18, +0x38 | OWD-R07/R08/R10 |
| Item list condition flags | timed `2`, curse `0x20`, exp `0x800`, aura `8` | OWD-R08 |
| Original stat identity encoding | high 16 bits stat, low 16 bits layer; stat 135 layer 0 `0x870000` | IEP registration / OWD-R02 |
| Item event list/entry | Unit +0x90; event byte +0, kind +4, identity +8, values +0xC/+0x10, callback +0x14, links +0x18/+0x1C | OWD-R01; IEP |
| Item callback ABI | five stack arguments owner, other, record, `value1`, `value2`; event in edx; ret 0x14 | OWD-R02/R05 at Game `0xAD410` |
| Applier request | nine dwords source,target,skill,level,duration,stat,value,state,callback; request pointer in eax | OWD-R05 at `0xA00B0` |
| Frame-event chain/entry | Unit +0xDC; entry type +0, flags +2, frame +4, target +8, type +0x10, args +0x14/+0x18, bucket links +0x1C/+0x20, unit links +0x24/+0x28, override +0x2C | OWD-R09c/R13 |
| Running/pending-removal frame event | flags bit 0 / bit 3 | OWD-R09c; removal `0x8E6F0` |
| Game frame / queue | Game +0xA8 / +0xB8; heads +4, newest heads +0x504, immediate +0xA04, saved next +0xA18 | OWD-R13 candidate ordering |
| State definitions/count/masks | DataTables +0xBC/+0xC4; mask k at +0xCC+4k; curse k11, exp k30, life k32 | OWD-R06/R08/R09b; GDA-R13 |
| Monster definition eligibility | MonStats +0xC bit 8 npc; signed word +0x18 MonStatsEx; MonStats2 +4 bit 9 isAtt | OWD-R06a; GDA-R08a/GDA-R09a/GDA-R16 |
| Final damage hand-off | total +0x4C; result +4 bit 2 lethal | OWD-R12; DPO/CSA |
| Unit life replacing write | D2Common #10887 / `0x3A740` -> `0x3A280` | OWD-R09/R11; unit effects RLI-Q03 / OWD-Q12; list writer STL-R06c is separate |
| Life-percent notification | D2Game `0x7A260` | OWD-R09d, unread OWD-Q08 |
| Owner lookup / death | D2Game `0xE03A0` / `0xDEEE0` | OWD-R10; real death remains gated; valid registered lookup uses the supplied result and the exact source record below |

### Compatible registered-unit lookup (2026-10-14 evidence reused)

The original Open Wounds record left Game `0xE03A0` decoding unchecked. The current
[item-event evidence](combat-item-event-triggers-and-procs.md#targeted-gap-investigations-2026-10-14)
records complete disassembly plus original-image corroboration of that 112-byte function and
the 20-byte unit-bucket table at `0xFB014`, dated 2026-10-14. Its valid game/type/id lookup
walks the appropriate bucket and returns the matching unit or null. The matching artifact
and lookup inputs apply here, independently of the reanimation caller that previously
needed it. That closes only the lookup slice of OWD-Q10; it does not read death `0xDEEE0`
or prove downstream credit. Its reused input boundary requires a valid game and owner type
0..5; the function does not provide a successful null-result fallback for arbitrary invalid
owner types. No new lookup read was performed during this migration.

### New bounded helper check (2026-10-15)

The former life-mask/getter question's unread-search limit is closed for the supplied-input life-state/getter boundary.
D2Common #10372 at `0x33D80` loads the life mask and calls `0x33C20`. A targeted complete
disassembly read established null/base/storage guards, signed word-count loop and first
intersection return. Original matching `LOD113C-D2COMMON` image bytes at `0x33C60` are
`8B 14 0F`, decoding as `mov edx,[edi+ecx]`; this verifies the formerly unchecked instruction
used in the mask intersection. The image SHA-256 was checked against the manifest.

The word count is not an unadjusted arithmetic shift. Instructions first add 31 to the
signed state count with 32-bit wrap, sign-extend with `cdq`, keep the sign word's low five
bits (`and edx,31`), add that bias to the wrapped numerator, then arithmetic-shift by five.
This is signed truncation toward zero of the wrapped numerator divided by 32. A signed
nonpositive result skips scanning. Positive scan counts require sufficient supplied mask
and active-state array coverage; no valid original mask storage is inferred for arbitrary
corrupt counts. This retains the exact six-word result for state count 185.

The base-own reader #10587 / `0x38C20` calls `0x38B00`; complete binary search/minimum
behavior consumes the checked RLI-R02a/RLI-R02b boundary and its original source records.
A targeted wrapper read confirmed the unit/id/definition/base guards and own-array path;
no additional per-read summation occurs. #10910 / `0x39430` likewise forwards to the effective
reader `0x38A80` after nonnull-owner callback entry; max-life #10913 / `0x390F0` forwards stat 7
when its base/definition guards pass. New reading used disassembly because these wrappers and
the life-mask helper had no available reconstruction checked against original bytes.

OWD-R09b is now `verified in 1.13c code (static)` only for supplied masks/arrays and the
subtraction. It does not verify mask production, attached contribution propagation or gameplay.
OWD-V06's old g/life results are unchanged. Frame-event removal `0x8E6F0` was read to its return
to check applicability against the different STL unit-timer mechanism; it preserves the existing
OWD-R09c result and does not close global equal-frame scheduler ordering.

The linked MonStatsEx helper at Game `0x1320` was also read completely because its original
record left a byte-encoded instruction without independent corroboration. The matching
`LOD113C-D2GAME` image contains `0F BF 44 02 18`, the signed-word load of MonStatsEx before
the second bounds test. This supports OWD-R06a's supplied linked-record selection, not
compiled-table publication. No byte-checked reconstruction was available for this helper.

### Compatible curse-list lookup (2026-10-15 evidence reused)

The [complete PvP lookup record](combat-pvp-damage-and-kill-credit.md#redirect-helper-reading)
resolves D2Game thunk `0xA7B4`, import slot
`0x6FD1850C`, to Common #10195 / `0x384F0`. That complete 48-byte function requires a
nonnull unit, returns null for missing base or a nonextended base, then calls complete
64-byte helper `0x37F40`. The matching original Common image corroborates every instruction.
With requested flag mask `0x20`, it walks the primary chain and returns the first list with
any matching flag; the base is not a candidate. A request containing `0x2000` would instead
select the secondary chain and clear that selection bit before testing, which is not this
call's input. Neither helper mutates state or draws randomness; null unit faults.

This record is reused for exactly the generic applier's flag-`0x20` lookup, preserving its
2026-10-15 date and original artifact identity. OWD-Q13 no longer calls the lookup unread;
the complete alternate-profile caller verification and actual attachment/removal/storage
remain gated. The original artifact, function and method here suffice independently of private analysis files.

### Historical instruction-reading limits

The 2026-10-13 States mask builder Common `0x5DAD0` had two instructions decoded by hand;
current STL-R12b still leaves producer byte/definition correspondence gated. The frame runner
Game `0x8F070` had ten direct call destinations corroborated against the image. Scheduler
`0x8EA80` had direct destinations `0x8E810`, `0xA68E`, `0xE2240`, `0x8E680` and frame-clamp
bytes checked, with other decoded control flow not independently established. These limited
checks support locators and existing request slices, not the unresolved equal-frame queue
order or malformed-input domains. The old expiry-walk decoding limit is narrowed through
STL-R04's complete traversal contract; actual free effects remain STL-R05.

Fog #10259 is a data export at `0x26798`, not a function. Current GDA-R16 / RLI mask evidence
supplies the resolved original-image array identity; the data address is recorded in the table above. That lookup does not verify state-mask production.

## Named table cells

On 2026-10-15 the campaign read complete effective files through the game's loaded archive
loader on a responding `lod-113c` profile. This migration extracts named cells from those
cached full responses; it is not another loader read. The historical vanilla launch profile
is reused, with no fresh complete override inventory, supplying physical MPQ, loaded DLL
fingerprints or compiled-table correspondence. The text cells below are `verified in 1.13c data`
at that scope; they do not confirm gameplay or prove the live compiled values match.

| Effective virtual path | Read coverage | Effective SHA-256 |
|---|---|---|
| `data/global/excel/itemstatcost.txt` | offset 0, returned/total 40109/40109 bytes | `ee7fa82b48e0c4f27a566334b2da13cf1c83f5c4c0439552084a231cb5938375` |
| `data/global/excel/states.txt` | offset 0, returned/total 20920/20920 bytes | `898806c2b69e55e316ecfed48dde95aa2933829b94fdad234bdabc27bdbc23e1` |

Headers select columns by name. Explicit `ID` / `id` values below also equal the zero-based
data-row index excluding the header in these files. Only selected facts are published.

| File | Row identity / index | Column | Raw cell | Interpretation / claim |
|---|---|---|---|---|
| `itemstatcost.txt` | explicit id 6; row 6 | `Stat` | `hitpoints` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 7; row 7 | `Stat` | `maxhp` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 12; row 12 | `Stat` | `level` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 74; row 74 | `Stat` | `hpregen` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 74; row 74 | `fMin` | empty | Empty flag disables the named profile condition; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 109; row 109 | `Stat` | `curse_resistance` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 135; row 135 | `Stat` | `item_openwounds` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 135; row 135 | `itemevent1` | `domeleedamage` | Event name linked by the table loader; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 135; row 135 | `itemeventfunc1` | `15` | Decimal true/selector value; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 135; row 135 | `itemevent2` | `domissiledamage` | Event name linked by the table loader; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 135; row 135 | `itemeventfunc2` | `15` | Decimal true/selector value; OWD inputs / OWD-R01/R06/R08/R09b |
| `itemstatcost.txt` | explicit id 352; row 352 | `Stat` | `last_sent_hp_pct` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 2; row 2 | `state` | `poison` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 2; row 2 | `life` | `1` | Decimal true/selector value; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 52; row 52 | `state` | `preventheal` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 52; row 52 | `life` | `1` | Decimal true/selector value; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 54; row 54 | `state` | `uninterruptable` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 57; row 57 | `state` | `attract` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 62; row 62 | `state` | `openwounds` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 62; row 62 | `curse` | empty | Empty flag disables the named profile condition; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 62; row 62 | `exp` | empty | Empty flag disables the named profile condition; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 62; row 62 | `aura` | empty | Empty flag disables the named profile condition; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 62; row 62 | `life` | `1` | Decimal true/selector value; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 62; row 62 | `plrstaydeath` | empty | Empty flag disables the named profile condition; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 62; row 62 | `monstaydeath` | empty | Empty flag disables the named profile condition; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 62; row 62 | `bossstaydeath` | empty | Empty flag disables the named profile condition; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 92; row 92 | `state` | `death_delay` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 100; row 100 | `state` | `healthpot` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 115; row 115 | `state` | `burning` | Name/identity correspondence; OWD inputs / OWD-R01/R06/R08/R09b |
| `states.txt` | explicit id 115; row 115 | `life` | empty | Empty flag disables the named profile condition; OWD inputs / OWD-R01/R06/R08/R09b |

The complete States read contains raw `life=1` only for ids 2,52,62. The scan establishes this
text-profile fact, not loaded-mask production. Event-name identities for 5/6 consume the compatible complete original-loader record in
[Crushing Blow table evidence](combat-crushing-blow.md#table-data-verification): 2026-10-13
22:57:43 UTC, `data/global/excel/events.txt`, offset 0/requested length 65536, returned/total
420/420 bytes, complete header/file; header-selected column `event`, zero-based rows excluding
header: row 5 raw `domeleedamage`, row 6 raw `domissiledamage`. Its vanilla LoD 1.13c launch
profile is distinct from modified profiles. Historical effective-file hash, physical MPQ,
complete override inventory and compiled correspondence were not captured. The named text
cells are data evidence only. The selected ItemStatCost strings alone do not assign event numbers. Code dispatch and selector semantics remain separate binary evidence.

## Corrections and compatibility assessment

The 2026-10-13 document recorded these corrections to its 2026-10-12 predecessor:

- Monster life tick `0x76740` is monster table slot 3. Players use `0x79B10` -> `0x77CB0`,
  flooring at 256 without monster life-state subtraction, preventheal or town tests. The town
  immunity statement was narrowed to monsters below one HP; players never die from this step.
- A life-state monster with own hpregen +50 and bleed -436 uses g=-436, not -386; -386 still
  applies when +50 is attached or for the player step. OWD-V06 retains all five distinctions.
- State 62 is not a curse. OWD-R06 was split into eligibility and unreachable original-profile
  curse branch; `curse` correspondence and the exact named cells have separate data support.
- State 54 is `uninterruptable`; state 92 is `death_delay`. Poison owner precedes Open Wounds
  owner for the candidate bleed death; the death auxiliary is cleared before the lookup/hand-off.
- Correct original module RVAs are death `0xDEEE0`, frame-event removal `0x8E6F0` and
  scheduler `0x8EA80`; old reconstruction labels were relative to the text section instead.
- The 200-frame request became a candidate 199-payment expectation, not runtime confirmation.
  The old 200-payment result was replaced in that historical revision; no new payment result
  is introduced by this migration. The 199 result remains explicitly gated.
- Refresh keeps first magnitude. Replacement's candidate free invokes the old callback before
  new creation. The exact constructor allocation tag is the target's own tag, unlike several
  status appliers' game-pool argument. These facts do not verify stored attach/free propagation.

Migration on 2026-10-15 reuses current NRF/IEP/DPO/CSA/STL contracts at their recorded boundaries:

- Old references called existing-value reads "summed reads" and attributed writes to RLI-R03.
  RLI-R02a/RLI-R02b actually select already existing own/resolved arrays. RLI-Q03 retains
  the unit replacing-write propagation gap; STL-R06c owns the separate list-stat writer. Explicit fixture effects preserve old
  numeric recipes while actual storage remains OWD-Q12. Automatic item notification likewise
  remains IEP-Q01/IEP-Q03, independently of supplied IEP-R03 registration.
- The unknown monster bits no longer default clear for an arbitrary existing monster. Only
  fixture inputs and absent MonsterData provide false conditions; category producers remain Q01.
- STL-R04 closes the old unchecked expiry-walk traversal question, but STL-R05/STL-R06b/
  STL-R06c and STL-R02a retain real attach/free/storage/equal-frame order. Q09 therefore
  remains an integration question. STL-R11's generic callback supplies ordered requests,
  not unverified passive/animation bookkeeping defaults.
- Closed Q02 (curse mask), Q04 (room/town input boundary) and Q07 (no cross-build equivalence
  scope) are historical records here; Q03's helper issue is resolved as above. Remaining
  original Q01/Q05/Q06/Q08/Q09/Q10/Q11 retain their identities. Q12 makes write integration
  explicit, Q13 retains complete alternate curse-caller verification, Q14 groups pending observations.
- Old claims that all life additions could never overflow were too broad for arbitrary i32
  inputs. The main states original wrapping arithmetic without changing any ordinary vector value.

### Gaps assessed and targeted follow-ups

| Question / assertion | Reused evidence and migration outcome | Remaining smallest check |
|---|---|---|
| OWD-Q01 monster conditions | Original Game `0x22A80` tests both conditions; setter leads near `0x23D5E` and `0x232A4` do not prove category meaning; no default imposed | Trace those creation setters to actual categories |
| Life-mask/base-read question | Targeted wrapper/mask reading and matching original-byte check above; NRF getter/search contract now complete | Closed for supplied masks/arrays; no mask-producer claim |
| OWD-Q05 tick producers/restart | DPO-R19c keeps scheduler/per-hit update gated; CSA-R11 names poison/burn rescheduling. Historical candidate schedulers at Game `0xEA10`, `0xEA70`, `0x336D0`, `0x37E20`, `0x6E270`, `0xAE6D0`, `0xAEA40`, `0xAEDC0`, `0xCFBC0`, `0xDA350` were not read | Trace relevant producer or initial player event; supplied handler stays independent |
| OWD-Q06 callback order | Current IEP-R01/IEP-R07 fixes order from an explicit entry list; equipment notification timing is not newly verified | Matching combined-effect observation / producer order |
| OWD-Q08 percent/message | Old `0x76740`/`0x77CB0` arithmetic and `0x7A260` request recorded; sender unread | Read sender and actual stat/state delivery path |
| OWD-Q09 expiry/window | Current STL traversal/callback request contracts apply; attach/free and equal-frame queue order remain unresolved | Read missing lifecycle/queue mutation slices; observe refresh/expiry |
| OWD-Q10 death | Compatible original-unit-lookup source record dated 2026-10-14 checked Game `0xE03A0` and table `0xFB014`; `0xDEEE0` body still unread; IEP-R09a supplies notification placement only | Complete death effects and no-list/nonstandard owner-input handling within chosen ownership |
| OWD-Q11 external player prevention | Local player life step contains no preventheal branch; wider sites not read | Player regeneration producer contract, outside local supplied step |
| OWD-Q12 stored effects | Current unit-write RLI-Q03, list STL-R05/R06b/R06c and update RLI-Q02 explicitly preserve gaps; fixture effects provide no evidence of real propagation | Complete actual propagation/update paths |
| OWD-Q13 alternate curse branch | Compatible 2026-10-15 Common #10195 / `0x384F0` -> `0x37F40` lookup record reused for flag `0x20`; state-62 original data excludes this branch | Complete alternate caller/request verification if supporting another profile |
| OWD-Q14 observation | Five original recipes, no game action/capture in either pass | Authorized attributable original-game observations |

Repeating the old failed breadth searches with unchanged inputs would add no evidence. Current
shared contracts narrow their applicability; remaining questions stay open instead of claiming
fresh verification from age or presentation changes. No source was modified or reconstructed.

## Proposed observations

These procedures are unperformed proposals. They use module-relative locations and logical
values, with no private debugger/member/session requirement. Use owned matching original files.

- **OWD-O01**: Normal single player, character with total Open Wounds chance 100, no life leech,
  ordinary monster outside town. At Game `0xAD410` entry/return, record attacker/defender,
  event 5, stat identity 135/layer 0, second value 0, attacker seed pair and return. Basic melee
  attack once; repeat with bow/normal arrow for event 6. Require one NRF-R01 step and return 1.
  Compare apply/next leech placement if a separate leech-capable case is observed; no such draw
  is expected from the initial no-leech setup. Handler registers/stack map is above.
- **OWD-O02**: level 30 gives v = 436 against an unadjusted monster. Record its own/effective hpregen
  and active states first; capture game frame, life, state-62 list magnitude/owner/expiry and event-3
  calls at Game `0x76740` from proc frame F through F+210. Candidate payments F+1..F+199, none
  from this list at F+200. Repeat with refresh at F+100 (candidate end F+299), positive natural
  own regeneration and a zero-total-damage proc. Distinguish scheduler start, subtraction,
  actual expiry/free and equal-frame order rather than assuming all follow from duration.
- **OWD-O03**: use a hittable monster with recorded MonStats `npc=true` or MonStats2 `isAtt=false`.
  Require one chance draw and callback return 1, but no state/list/event requests. If original
  gameplay cannot hit such a target, record unreachable setup; do not manufacture a match.
- **OWD-O04**: weak monster with life below candidate `199*v`, no poison and an Open Wounds
  owner still registered. Stop hitting after the proc; capture zero-life monster tick, owner lookup,
  Game `0xDEEE0` request and dispatcher `0x37030` events 10 then 9. Compare player experience
  before/after separately; that downstream effect is a hypothesis requiring evidence. Repeat a
  poison-plus-Open-Wounds setup only with separately established list inputs to test preference.
- **OWD-O05**: two hostile characters, attacker level 30/chance 100, melee request v = 109, defender
  low life/no other regeneration. Record life writes from Game `0x77CB0` for 210 frames. Require
  decreasing life to 256, never a lower stored life from this step, and cessation after expiry.

## Historical review record and migration review

The original 2026-10-13 blind reader run 1 reported 17 GUESS and 19 GAP; run 2 reported
6 GUESS and 6 GAP. Added inputs/queue/expiry/callback/lookup distinctions were folded into
that document, and Q01/Q05/Q08/Q09/Q10 remained open. A third run was not performed then.
Those partial historical rounds do not satisfy the current campaign's fresh-reader gate.
Their complete original wording and private fixture encodings remain in the verified private
archive. Public corrections, identities and unresolved boundaries are retained above.

The 2026-10-15 writer self-review checked all 20 labeled rule leaves against their request,
arithmetic or explicit gate boundaries, all 14 vector families, every remaining question and
both dependency directions. Main rules use semantic values/operations; original binary
encodings are confined to this record and the private archive. It preserves all earlier
numeric recipes and five observation identities without claiming the two partial historical
reader rounds are a fresh campaign review. No implementation, gameplay, test, hook or Git
operation was performed by this writer.

OWD-V06 concretizes the old `256000+g` shorthand as 255564,255364,255950,255614,255614 for
the unchanged g values -436,-636,-50,-386,-386. These are additional explicit assertions of
the existing arithmetic, not corrections to old results. New bounded cases are OWD-V01(g),
where the real eligibility guard rejects state 62 with definition count 0 but the callback
still returns 1, and OWD-V08(c)/(d), distinguishing the player's lack of the monster
preventheal/town guards. Original curve outputs, seed steps/remainders, refresh magnitudes,
life sequence, percentage threshold, kill-owner choices and candidate 199-payment total are
unchanged. The exact subcase inputs, local execution boundaries and supplied effects are
explicit in the main; they do not verify real propagation or death.

The mechanical spec/companion/gap check reports 0 errors and 0 warnings on this stable
migration. The separately preserved before snapshot and migration comparison retain its
integrity hash, all original numeric cases and original fixture encodings. This is writer
self-review and documentation lint, not implementation-fixture execution, gameplay evidence,
independent-reader completion, approval or a status promotion. Fresh campaign review is pending.

### Campaign full reader pass 1 (2026-10-15)

A fresh smaller reader used the main contract and dependency contracts, recomputing 14
families and 70 cases/subcases. Reported counts: MISMATCH 1, GUESS 0, GAP 1, DISCRIM 0,
GATED 5 grouped limits, COVERAGE 6, HARDEST 3. The five grouped gates named monster-condition
production, lifecycle/payment-window integration, scheduling/restart, death/nonstandard inputs
and gameplay. This grouping does not remove the separately explicit notification, write or
alternate-profile gates.

| Finding | Outcome and reason |
|---|---|
| MISMATCH: tick 19 alone cannot distinguish threshold greater than 4 from at least 4 | No numeric correction: the entire existing V04 asserts only one notification in 25 ticks, at 19. At least 4 would instead notify at 14. The main now makes tick 14 explicit: life 249896, cur 976, p 124, difference 4, no write, old percent stays 128. The reported issue concerns an abbreviated discrimination explanation, not an erroneous old output |
| GAP: IEP-R13b does not specify a generic registered-unit lookup contract | Spec edit: removed that ownership citation. Section 3 explicitly supplies a read-only registered-unit lookup result for present game/valid type/id; V14 supplies its exact outcomes. The compatible original lookup source record remains evidence here. Death effects and no-list/nonstandard-input handling remain OWD-Q10; reanimation is not a required implementation dependency |
| COVERAGE: absent attacker | No change: R02's same absent-unit guard returns 0 with unchanged seed/no request for either absent input; V01(f) already demonstrates that result recipe for absent defender |
| COVERAGE: negative defender-adjustment value | Spec edit: V03(m) supplies v=-49/player/event 6 and requires -12 then -6. This distinguishes signed truncation from floor division without changing any previous positive result |
| COVERAGE: invalid class, absent MonStats, absent/invalid extended record or absent state record | No change: these explicit R06a eligibility guards all reach the same null/no-request outcome. V09(a)/(b)/(d) distinguish the existing rejection recipe and retained callback draw/return; duplicating every unavailable-record binding adds no distinct effect or arithmetic recipe |
| COVERAGE: percent difference 4, nonpositive maximum, current at least maximum | Spec edit for difference 4 as above. V13(b) already stores maximum and then computes cur=m=1000, p128 unchanged. Nonpositive m remains R09d's explicit gated candidate branch, with no accepted notification behavior promoted; no fixture is needed to claim that incomplete operation verified |
| COVERAGE: death tail without either selected list | Explicit gate: Q10 and the R10 execution boundary now name no-list tail entry. V14(e) retains a selected list but supplies a missing registered owner; that does not establish the no-list caller branch. No fabricated owner or successful-null fallback was added |
| COVERAGE: aliveExclusionCondition, remaining death modes, health-potion removal | Spec edit: V08(e) supplies the independent aliveExclusionCondition=true guard and requires initial reschedule followed by no life/stamina/mana calls. Other stated death-mode constants share V08(b)'s early-return recipe. R11b remains asm-only with actual removal gated; no accepted removal contract is inferred from a fixture |
| HARDEST: actual lifecycle/payment count | No change: Q09/STL-Q07 explicitly gate the real effects and 199-payment candidate; supplied effects do not settle them |
| HARDEST: monster tick/death integration | No change: Q05/DPO-Q24 and Q10 explicitly preserve scheduler/update/death boundaries |
| HARDEST: matching gameplay setup | No change: Q14 and all observation procedures remain pending, with no table-to-gameplay promotion |

There were no GUESS or DISCRIM entries needing separate disposition. All prior numeric
results and evidence labels remain unchanged after this review. V03(m) and V08(e) add two
bounded cases, while V04 makes an existing 25-tick assertion explicit. The main's lookup
execution/gate clarification requires another fresh full reader pass; this round does not
establish review completion or approval. Mechanical lint and arithmetic/preservation checks
were rerun with 0 errors and 0 warnings; no implementation tests or gameplay were run.

### Campaign full reader pass 2 (2026-10-15)

A second fresh smaller reader recomputed 14 families and 72 cases/subcases. Reported counts:
MISMATCH 0, GUESS 0, GAP 0, DISCRIM 6, GATED 10 grouped limits, COVERAGE 2 summary entries,
HARDEST 3. Its ten gate groups were automatic notification, monster-condition production,
alternate curse application, detach/free, stored attachment, life write/sender, death/no-list,
health-potion removal, monster restart and the 199-payment candidate. All existing individual
question and dependency gates remain authoritative.

| Finding | Outcome |
|---|---|
| DISCRIM 1: chance draw could use signed low-word remainder | Added V01(h): seed `{4,666}` steps to `{0xAB1A45AE,1}`, unsigned remainder 10 with chance 5 returns 0/no request. A signed remainder would be -86 and incorrectly succeed |
| DISCRIM 2: curve has no supplied overflow input | Added V02(a): level 47721889 produces f = 2147483646 then signed i32 v=-2147483610 after adding 40. A wider unwrapped result is 2147483686. This is a bounded extreme input, not an original-profile attainable level claim |
| DISCRIM 3: refresh duration 0 could still update/schedule | Added V05(a) at the isolated generic-applier boundary. Equal skill/level and duration 0 return the existing list, leaving expiry 1100, magnitude -300 and its single queued expiry event unchanged. The upstream item callback, which always supplies 200, is not executed |
| DISCRIM 4: preventheal guard could test positive rather than nonnegative g | Added V13(d), active state 52 and g = 0, requiring return before schedule, cancellation or life/percent work; a preexisting nonrunning kind-3 P remains alongside unrelated expiry E after the running event returns |
| DISCRIM 5: cancellation could remove only running/new kind-3 events | Added V13(e) with an additional preexisting nonrunning future kind-3 entry. Cancellation removes it and the newly scheduled entry, marks the executing entry, retains the unrelated kind-12 entry and leaves only that expiry event after dispatcher return |
| DISCRIM 6: player maximum cap and one-HP floor order | Added V08(f), supplied life 100/g = 200/maximum 128: sum 300 caps 128, requests supplied health-potion detach/free, then floors 256. The log requires removal requests before life store 256; floor-before-cap would store 128. Actual removal/write/sender effects retain their gates |
| COVERAGE: existing branch coverage summary | No additional cases: existing seed/no-draw, target adjustments, eligibility, refresh, monster/player life, percentage threshold, scheduling and hand-off recipes retain their inputs/results |
| COVERAGE: boundary-coverage summary | The six concrete boundary deficiencies are resolved by the six discriminating cases above; no further branch was requested by the summary entries |
| HARDEST: V04's 25 sequential ticks and candidate 199-payment window | No change beyond explicit tick 14 from pass 1: local arithmetic and the existing gate on real payment count remain clear |
| HARDEST: V05/V09/V10 callback, eligibility, refresh and replacement | No change beyond duration 0 case above: explicit isolated/supplied boundaries and actual replacement gates distinguish local requests from dependency effects |
| HARDEST: V14 owner preference/state 54/null/event order | No change: supplied lookup outcomes and existing no-list/death gates remain explicit; no accepted death effect is added |

There were no MISMATCH, GUESS or GAP entries to resolve. All old numeric outputs, rule
labels, ownership and gates remain unchanged. Six additional bounded cases expand the recipes
without adding rule behavior. A short negative reanimation comparison was removed from the
supplied lookup definition; its positive contract is unchanged. Fresh full pass 3 is required
after these fixture additions. Lint and private arithmetic/preservation checks were rerun;
no implementation test, gameplay, source reread or status promotion occurred.

### Campaign full reader pass 3 (2026-10-15)

A third fresh smaller reader recomputed all 14 families and 78 cases/subcases, counting each
sequential recipe once while computing every step. Counts: MISMATCH 0, GUESS 0, GAP 0,
DISCRIM 4, GATED 7 grouped limits, COVERAGE 3, HARDEST 3. The grouped gates named monster
conditions, automatic notification, tick producers/update, notifications, lifecycle/equal-frame
order, death/no-list/nonstandard inputs, and writes/alternate profile/gameplay. They preserve
every separately stated question and dependency limit.

| Finding | Outcome and reason |
|---|---|
| DISCRIM 1: a defender-only callback null guard might pass without an absent-attacker case | No change: deleting only the attacker guard while retaining the consumed safe RLI-R02a Stat read is observably equivalent. Stat(null)=0, then chance<=0 returns 0 before drawing/requesting anything. A proposed positive chance cannot override this null getter. The rule still records its actual two-unit guard; no observable pure-read call count or fabricated null-unit stat value is added |
| DISCRIM 2: direct creation with absent source has no owner-fallback assertion | Added V01(i), explicitly entering R08 after eligibility and bypassing the callback. Source absent produces owner `(6,-1)`, while target, ordered creation requests, state 62/stat 74=-436/skill 0/level 1/expiry 1200 and return L are explicit. This does not contradict the callback's absent-attacker rejection |
| DISCRIM 3: monster positive g below maximum has no life/queue discriminator | Added V13(f), life 128000 plus g 256 gives 128256 below maximum 256000. Percentage 64 is unchanged; no cancellation occurs and next tick N1 survives alongside expiry E after the running event returns |
| DISCRIM 4: other unit types in the player-event handler could still run life/resource steps | Added V08(g), present type 2/mode 1/aliveExclusionCondition false: request next event first, then no life/stamina/mana calls. It is isolated handler execution, not a claim about original type 2 dispatch |
| COVERAGE: absent-unit PercentUpdate fallback 128 | No change: R09d already states this gated candidate fallback. No accepted sender/delivery behavior is established by another constant-input fixture |
| COVERAGE: state mismatch replacement | No impossible fixture: RLI-R03's same-state lookup returns only a list with the requested state. A different-state list selected by the alternate curse-flag lookup belongs to R06b/Q13's existing gate, not the accepted noncurse refresh path. Skill mismatch and higher level remain independently exercised by V10 |
| COVERAGE: invalid/missing eligibility record enumeration | No change: these guards remain explicit and use the existing V09 null/no-request recipe; exhaustive unavailable-record bindings add no distinct local result |
| HARDEST: V02 band slopes and i32 wrap | No change: exact piecewise arithmetic and the independently computed extreme-level case already distinguish them |
| HARDEST: V04 sequential raw life/strict percent threshold and gated 199 window | No change: first/14th/19th/25th values are explicit; actual 199-payment integration remains gated |
| HARDEST: V13 scheduling, cancel traversal, executing-event deferral and preventheal | No change beyond positive-g case above: explicit before/after pending chains and future-event retention/removal distinguish the existing local paths |

All four discrimination findings and every advisory entry have a recorded disposition.
No previous numeric result, rule behavior, label, ownership or gate changed. Three new bounded
cases expand the fixture set; no source was reread, implementation test run or gameplay action
taken. Private arithmetic/preservation checks and mechanical lint passed with 0 errors and
0 warnings. As the third full pass has MISMATCH 0 and complete triage, the fixture-only
changes require a fresh delta review of V01(i), V08(g) and V13(f). Review completion and
approval remain pending.

### Campaign delta reader pass 4 (2026-10-15)

A fresh smaller reader recomputed the three added cases OWD-V01(i), OWD-V08(g) and
OWD-V13(f) from the current rules and dependency contracts. Counts: MISMATCH 0, GUESS 0,
GAP 0, DISCRIM 0. Each case's inputs, exact outputs and discriminator were checked; there
were no changed-material findings needing correction.

The reader retained the stored-creation gates OWD-Q09/OWD-Q12, the isolated type-2 handler
boundary, and the life-write/event integration gates OWD-Q12 and OWD-Q05/OWD-Q09. No rule,
ownership or gate changed after the third full review. The applicable reader stopping
condition is met. Gameplay confirmation, resolution of the remaining evidence gaps and
maintainer approval remain outstanding; this review does not promote the draft status.

### Restructure reference check and reader passes 5-8 (2026-10-15)

After the 2026-10-15 dependency restructure (record/stat/state material moved from NRF to
core-record-layouts-and-ids, RLI; table-record fields moved from core-game-data-fields to
core-game-data-table-records, GDA; shared name `aliveExclusionCondition`), every cited id was
checked against the current dependency specs. Remap corrections: the stat-reader input row
cited the old NRF section 3.8 and now cites RLI section 3.7; the Dependencies paragraph had a
sentence spliced into it and now names each owning file; bare GDA-R08/GDA-R09 (families) are
now GDA-R08a/GDA-R09a; GDF-R02/GDF-R03 are cited for the identity convention only; section 10
lists the two new owners. Run 5 was a delta check of the remapped references, run 6 a delta
check of its fixes, run 7 a full run (required after the OWD-R06a wording change), run 8 a
delta check of the run-7 additions.

New readings used in triage (LOD113C-D2GAME original disassembly, 2026-10-15):
- Frame-event link step `0x8E680`, called from scheduler `0x8EA80`: the new entry's unit-chain
  next is the unit's current chain head (+0xDC), the old head's previous points to it, its
  previous is cleared, and the unit head is set to it. This is head insertion; the 8.1
  preamble now states it instead of attributing a position to STL-R01, which gives none.
- MonStatsEx helper `0x1320` and its call site in applier `0xA00B0`: class and signed-i16
  link are each rejected when negative or at or above the respective count; the helper adds
  no null test, and the caller rejects a null result before testing `isAtt`. OWD-R06a now
  states the link bounds and the absent-record rejection explicitly (no behavior change).

| Run | Kind | MISMATCH | GUESS | GAP | DISCRIM | Other |
|---|---|---|---|---|---|---|
| 5 | delta (remapped references) | 0 | 2 | 5 | 0 | GATED listed |
| 6 | delta (run-5 fixes) | 0 | 2 | 4 | 4 | |
| 7 | full | 0 | 5 | 4 | 8 | GATED 17, COVERAGE 14, HARDEST 3 |
| 8 | delta (run-7 additions) | 0 | 0 | 0 | 0 | all ten listed findings confirmed resolved |

| Finding | Outcome and reason |
|---|---|
| R5 GUESS: SetState's own update marking as a log entry | Spec edit: 8.1 preamble logs one entry per requested operation; requests made inside a dependency operation are not logged |
| R5 GUESS: insertion position in the unit chain (V13(e),(f)) | Spec edit: preamble states head insertion, verified in the link step above |
| R5 GAP: bare GDA-R08/GDA-R09 do not resolve | Spec edit: GDA-R08a/GDA-R09a (section 3, section 9, companion field table) |
| R5 GAP: GDA-R16 does not define npc/isAtt | Spec edit: fields cited to GDA-R08a/GDA-R09a; GDA-R16 kept only for "a false flag tests 0" |
| R5 GAP: GDF-R02/GDF-R03 do not list the specific ids | Spec edit: section 9 cites them for the row-index convention; specific ids are section 3 identities with named-cell records |
| R5 GAP: RLI-Q03 wording versus "unit-write propagation" | No change: RLI-Q03 covers replacing stat-write integration and consumers requesting replacing writes cite it |
| R5 GAP: no unexecutable vector | No change: nothing to resolve |
| R6 GUESS 1-2: V10 replacement callback log and old-state inputs | Spec edit: V10 note leaves callback-internal requests and old event retention unasserted under OWD-R07b/OWD-Q09/OWD-Q12 |
| R6 GAP 1 / DISCRIM 2: in-range link with absent record | Spec edit: OWD-R06a absent-record rejection (asm above); V09(g) added |
| R6 GAP 2-3: V01(e), V09, V05 list requests by reference to R08/R07a; V05 expiry-set not a log entry | No change: outputs are fully determined by R08/R07a order and asserted expiry values |
| R6 GAP 4: head insertion unverified | Resolved by the link-step reading above |
| R6 DISCRIM 1: only negative link tested | Spec edit: V09(f) link equal to the count |
| R6 DISCRIM 3: V10 callback log | Gated as above |
| R6 DISCRIM 4: only V13(f) separates head from tail | No change: (f) asserts the exact chain |
| R7 GUESS: V10 callback logging / old event 1100 | Spec edit: V10 note (as R6) |
| R7 GUESS: R10 list selection and owner identity | No change: R10 states poison first, else state 62, and the selected list's owner |
| R7 GUESS: R09d stat 352 layer | No change: section 3 gives layer 0; R09d is gated |
| R7 GUESS: release of a pending-removal running entry | Spec edit: preamble states one release after the handler returns |
| R7 GAP: PVP-R06 room/town predicate | No change: room and `isTown` are supplied inputs in section 3 |
| R7 GAP: GDA-R16 for npc/isAtt | No change: cited only for false-flag result |
| R7 GAP: default defender id; NewList allocation tag | No change: no output depends on them |
| R7 DISCRIM: both specials true | Spec edit: V03(n) |
| R7 DISCRIM: magnitude retention order | Spec edit: V05(b) |
| R7 DISCRIM: `total>0`, negative-only hpregen | Spec edit: V11(f),(g) |
| R7 DISCRIM: `>=` maximum (monster) | Spec edit: V13(g) |
| R7 DISCRIM: `>=` maximum (player healthpot) | Spec edit: V08(h) |
| R7 DISCRIM: `life<=256` guard; guard for nonnegative g | Spec edit: V07(d),(e) |
| R7 DISCRIM: callback ignores event argument | Spec edit: V01(j) |
| R7 DISCRIM: burning-list fallback / poison-only | Spec edit: V14(f),(g); no-list tail remains OWD-Q10 |
| R7 COVERAGE (14 entries) | No change: remaining branches either share an existing null/no-request outcome, are gated (R09d, R10 no-list, R11b, R12/R13 candidates) or were covered by the DISCRIM additions above |
| R7 HARDEST: V13 chain model; V10/V05(a) logging; V14 death tail | No change beyond the preamble and V10 note: the chains and gates are explicit |

The applicable stopping condition is met by run 8 (MISMATCH 0, no unresolved GUESS, GAP or
DISCRIM). No gameplay observation was made; status remains draft and approval is pending.
