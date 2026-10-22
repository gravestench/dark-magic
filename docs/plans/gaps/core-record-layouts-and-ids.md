# Gaps: core-record-layouts-and-ids

Spec: [Record Layouts and Ids](../../specs/core-record-layouts-and-ids.md).
Evidence: [migration attempts and source records](../../research/spec-evidence/core-record-layouts-and-ids.md#foundation-migration-source-record-2026-10-14).

Status-application gaps (stored list effects, status timer-node and handler execution, status-path conditions
and status observations) are rows of the [status application gap file](combat-status-application.md). Random,
seeding, arithmetic, time and animation-preparation gaps are rows of the
[numeric foundations gap file](core-numeric-and-rng-foundations.md).

All rows were assessed on 2026-10-15 against their recorded source identities, methods,
actual boundaries and the current dependency contracts; see the
[reuse assessment](../../research/spec-evidence/core-record-layouts-and-ids.md#focused-regeneration-and-evidence-reuse-2026-10-15).
The 2026-10-14 attempts below retain their actual dates. Reuse does not renew verification,
complete a failed search or establish gameplay execution.
Current STL-R07 establishes active/changed membership separately from STL-R07a disguise,
RLI-Q02 update-queue effects and STL-R05/STL-R06b/STL-R06c lifecycle integration.

Restructure (2026-10-15): these rows came from the numeric foundations gap file when the record, stat-read,
state and mode rules moved to this spec; the animation/bookkeeping and gameplay-confirmation rows were divided
by the rules they affect. No attempt, date or boundary changed.

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| Wider runtime-condition meanings | outside-boundary | RLI-Q01, RLI-R03, RLI-R04a, RLI-R01a, NRF-R12a | None locally: explicit tested predicates are available; wider roles/names of neutral conditions remain owner-specific | Producer/consumer meaning in the owning subsystem | 2026-10-14: tested masks and local consumers read, original field mappings retained in companion; no semantic name table exists for private runtime records | Trace each condition's producer and behavioral consumers in its owner |
| State and update bookkeeping | integration | RLI-Q02, RLI-R03, RLI-R04a, RLI-R04b | Disguise and update-queue side effects of state-set routing, mode changes and mode-bound traversal; local requests remain eligible | Complete state/update dependency contracts | 2026-10-14: read `0x38730` and `0x665C0`/`0x2FE10`; queue and disguise effects remain outside the inspected slices | Accept complete named bookkeeping owners and their gated paths |
| Actual free, attach and replacing-write integration | integration | RLI-Q03, RLI-R04a, RLI-R04b | Actual list-free effects requested by mode-bound traversal and the stored effects of replacing stat writes requested by consumers, through gated STL-R05/STL-R06b/STL-R06c and stat-update producers; supplied ordered calls, traversal and getter values remain eligible | Verified complete free/setter/attach/propagation contracts | 2026-10-14: independent listings of `0x3A280`, `0x3A7D0` and `0x3A740` inspected; branching callees/indirect effects remain under the STL owner; no blanket promotion from call-list inspection | Upgrade owning lifecycle contracts with complete callee reads and integration fixtures |
| Gameplay confirmation | observation | RLI-Q04; verified local rules and their vectors | Runtime-confirmation claims only; static/supplied-input assertions remain eligible | Valid matching observations of the stated rule/vector cases | 2026-10-14: retained prior observation procedures and reviewed recorded layout/table reads; no traces/actions/memory writes performed | Run a separately authorized gameplay/helper observation campaign and record actual matching coverage |

No implementation comparison was requested; this file contains evidence gaps only.
