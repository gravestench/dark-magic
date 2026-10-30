# Gaps: sound-table-and-playback-rules

Spec: [Sound table and playback rules](../../specs/sound-table-and-playback-rules.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| Admission and context lifecycle | behavior | STP-R03b, STP-Q01 | duplicate stopping/deferral, slot eviction/priority and wider seed-context lifecycle; supplied group/history stages remain available | complete remaining caller/callee and context-producer traces | 2026-10-29: group helper and history stage closed as STP-R03a/STP-R03c with original descriptor and native-byte corroboration; split admission lookup and D2Sound start-marker accessor also read; complete loading/slot/duplicate paths remain | close admission/eviction/loading callees and context assignment lifecycle |
| Device integration and clock units | integration | STP-R04b, STP-Q02 | full playback/loop/volume lifecycle, stream starts, invalid schedule arguments and retirement; supplied STP-R02/STP-R04a/STP-R05 remain available | real device operation and complete lifetime contracts | 2026-10-29: D2Sound ordinal 10037 wrapper and import/device boundary read; scheduler 88230 and matched local helpers fully read; clock units and later retirement not established | trace scheduled voice through updates, stopping/retirement and stream/device failures |
| Stage observations | observation | STP-R02, STP-R03a, STP-R03c, STP-R04a, STP-R05, STP-Q03, STP-O01 | runtime confirmation only | complete matching transition/group/history/device-wrapper observations | 2026-10-29: initial admission capture inspected 4 records; follow-up captures inspected 6 group returns and 4 buffer marker/result records with dropped 0, all partial against full vector assertions | capture same-entry group seed/history and all draws; scheduler/transition inputs and writes; exact HRESULT/lock/device operation sequence |
