# Gaps: Player hostility declaration

Spec: [player-hostility-declaration](../../specs/player-hostility-declaration.md). Evidence: [companion](../../research/spec-evidence/player-hostility-declaration.md#gap-assessment-and-migration).

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| Complete declaration integration | integration | HOS-Q01/HOS-R04/HOS-R03/HOS-R01 | Complete stored/dependency-backed effects only; named supplied-input slice available | Complete producer/consumer path at the stated unresolved boundary | 2026-10-15 applicability assessment: Embedded handler and command/table roles located; local numeric/time decisions verified. Full node creation, enable/disable, party/skill cleanup and delivery remain incomplete | Trace the remaining named path through its actual output writes and add a distinguishing fixture |
| Gameplay confirmation | observation | HOS-Q02/HOS-O01/HOS-R02 | Runtime-confirmation claims only; supplied static assertions available | Attributable original 1.13c execution | 2026-10-15 applicability assessment: All prior/runtime columns untested; no new gameplay action. Static assertion acceptance remains separate | Observe the named local assertion with matched inputs and reached boundary |

## Implementation gaps

The preserved historical comparison for these rules is row 4 of the [PvP damage and kill-credit gap file](combat-pvp-damage-and-kill-credit.md#implementation-gaps), which compares HOS-R02's level 9 and 60000-ms values with an earlier research note. It was not rerun or updated during the restructure and is not a current implementation claim.

## Restructure (2026-10-15)

The declaration integration row moved here from the [PvP damage and kill-credit gap file](combat-pvp-damage-and-kill-credit.md) when the hostility-declaration request became its own specification. The gameplay-confirmation row is the declaration part of that file's general gameplay-confirmation row, with its attempt record unchanged. No gap was closed or reassessed.
