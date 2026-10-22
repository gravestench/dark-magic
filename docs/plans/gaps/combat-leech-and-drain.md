# Gaps: combat-leech-and-drain

Spec: [Leech and Drain](../../specs/combat-leech-and-drain.md)
Evidence: [companion records](../../research/spec-evidence/combat-leech-and-drain.md)

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|-----|------|---------|-------------------------------|------------------|-----------------------------------|----------------|
| Stored effects of requested writes | integration | LSE-Q01, LSE-R09, LSE-R13, LSE-R14, LSE-R15; LSE-V07 to LSE-V14 | Actual stored results of the unit replacing write, the unit additive write, the overlay list attach, the list stat write and the update mark; arithmetic, caps, returns and ordered requests with supplied operations remain eligible | Complete contracts for D2Common #10887, #10551's core `0x3A330`, attach (STL-R06b), list stat write (STL-R06c) and the update queue (RLI-Q02); NRF-R08a defines formats only | 2026-10-14: read D2Common `0x3A710` (forwards a nonzero value to `0x3A330`), `0x665C0` (requests queueing on the room's update list via `0x2FE10`) and the overlay helper's calls; that attempt did not read `0x3A330`, `0x2FE10` or attach/list-write internals. 2026-10-15: compared current NRF/STL and HRC/KNB evidence; compatible room-getter reading establishes room/level/town selection, not queue mutation or stored stat propagation. Requests remain verified, stored effects do not | Accept owning contracts for these writes with integration vectors |
| Disguise condition and form-state list producers | outside-boundary | LSE-Q02, LSE-R14 | None locally: the apparent-form decision for supplied `disguiseCondition`, list and `gfxtype` values is verified; which units are disguised and which states the list holds are unknown | The writer of the disguise condition (STL-R07a) and the loader of the DataTables form-state list | 2026-10-14: read D2Common `0x501E0` in full; searched the D2Common listings for stores to the list and its count, finding readers only. 2026-10-15: current STL-R07a/RLI-Q02 still leave the producer unresolved; no compatible producer proof to reuse | Identify the loader that builds the list and the disguise bookkeeping writer |
| Overlay count identity | outside-boundary | LSE-Q03, LSE-R13 | None locally: the bound test against a supplied count is verified; whether ids 151 and 152 pass in the original profile is not established | The table published at the counted DataTables slot and the names of overlays 151 and 152 | 2026-10-14: searched the D2Common listings for stores to the counted slot and its record pointer, finding only readers (a family of record getters and the overlay helper); crushing-blow evidence has the same open question. 2026-10-15: GDF definition fields and compatible item-event overlay records add no collection-publication or named-row proof | Read the loader that stores the slot, then read the named rows of that table |
| Client messages | outside-boundary | LSE-Q04 | None locally; message content belongs to a `net-` owner | The update sender carrying stat 6, stat 8 and the overlay stat | 2026-10-14: confirmed the overlay helper requests queueing for update (`0x665C0`); no sender trace performed. 2026-10-15: current RLI-Q02 and unresolved STL-R14/STL-Q01 provide no compatible sender/body/delivery contract | Trace the server update sender in a `net-` spec |
| Gameplay confirmation | observation | LSE-Q05, LSE-O01, LSE-O02; LSE-R01 to LSE-R15; LSE-V01 to LSE-V14 | Runtime-confirmation claims only; static and supplied-input assertions remain eligible | Valid matching captures of the observation vectors and of rule cases, including the poison term of the total (DPO-Q02) | 2026-10-13: a running member held a dead character and no hit could be produced; 2026-10-14: migration performs static and table reads only, no gameplay action. 2026-10-15: reviewed those records and current DPO-Q02; no matching capture exists and no gameplay action was performed | Run an authorized observation campaign and record matched coverage in the companion |

## Implementation gaps

Comparison of the original 1.13c behavior with the Dark Magic implementation, as previously
recorded. Findings are unchanged; rule and vector references use the spec's current ids.

Repo file named in the comparison: `policy/leech.lua` (and its test). Its header cites 1.10f D2MOO as
its source. All values there are implemented assumptions unless noted.

| # | Spec rule | Repo file/function | Current repo behavior | Original behavior | Original status |
|---|-----------|--------------------|-----------------------|-------------------|-----------------|
| 1 | LSE-R01 to LSE-R08 player-path leech (base, drain P, x64 scale, difficulty divisors, mulDiv, >>6) | `policy/leech.lua` | Agrees with the player/percent path: base, drain, x64, difficulty divisors for players only, and the defender mana quirk | Same arithmetic as LSE-R01 to LSE-R08 | verified in 1.13c (static) |
| 2 | LSE-R12 overlay draw | `policy/leech.lua` | Extra `GetRandomModulo(2)` draw only when both leech types fire, matching the Lua pick | One extra modulo-2 draw only when mana leech occurred and life leech non-zero | verified in 1.13c (static) |
| 3 | LSE-R14 monster-attacker leech branch | `policy/leech.lua` | Flat monster leech not compared ("unverified here"); not yet compared to `leech.lua` | Branch calls D2Common thunk (label L_097CC), caps by defender mana/stamina and attacker max life | asm-only; gated |
| 4 | DPO-R12a poison immediate total (consumed by LSE-V01) | `policy/leech.lua` (Lua immediate total) | Poison is NOT in Lua's immediate total | 1.13c's immediate total does include `[+0x28]` | verified in 1.13c (semantics unresolved) |
| 5 | Provenance | Dark Magic files compared | `policy/leech.lua` only | n/a | n/a |
| 6 | Done-checklist | "Monster-attacker leech compared to `leech.lua`" open; "LSE-V09 captured on an owned runtime; #383 links the spec" open | pending | n/a | gated |

## Restructure (2026-10-15)

Renamed from `combat-leech-and-status-effects` with its spec and companion. The rules restating
the pipeline total and stage order were removed in favour of DPO-R12a, DPO-R14 to DPO-R18 and
CSA-R01, and the remaining rules renumbered (details in the
[companion record](../../research/spec-evidence/combat-leech-and-drain.md#restructure-2026-10-15)).
Gap rows and implementation findings are unchanged apart from their rule references.
