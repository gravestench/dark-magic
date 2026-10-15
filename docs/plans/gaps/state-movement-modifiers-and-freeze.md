# Gaps: state-movement-modifiers-and-freeze

Spec: [Movement modifiers and freeze](../../specs/state-movement-modifiers-and-freeze.md).

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|-----|------|---------|-------------------------------|------------------|-----------------------------------|----------------|
| Derived monster animation bases | integration | MMF-R04b; GDF-Q01 | Actual ordinary/run prepared bases, not MMF-R04a supplied-word arithmetic | Complete GDF-R05b production | 2026-10-15 reused current GDF producer assessment; signed consumer words verified, not text columns | Owning derived-base producer through final writes |
| Full run/path production | integration | MMF-Q05; MMF-R03/MMF-R06d/MMF-R06e | Actual run baseline/trajectory; supplied velocity and step remain available | Caller modifier and direction/path inputs | 2026-10-15 Common movement consumer confirms same Walk base; reused NRF-R11a/R11b step/scale | Relevant original creators/callers through supplied inputs |
| Stored stat contribution/removal | integration | MMF-Q06; MMF-R13b/MMF-R13e | Actual chilled unit totals, restoration and callback invocation; local logs/selected reads remain available | Full attachment, replacing writes, array updates and removal | 2026-10-15 compared NRF-R12b/R12c getters and STL-R05/R06b/R06c; no summation/mutation established by reads | Complete owner propagation/removal boundary |
| Attack preparation and producers | integration | MMF-Q08; MMF-R05a/MMF-R05b | Actual slots/type qualification/base override/preparation effects and draws; supplied consumer available | Complete inventory/item-type, attack preparation and selected override chain | 2026-10-15 complete caller arithmetic and eligibility read; inventory/type/checked override assessed, provider/allocator callees not all verified | Rooted original provider paths and all local callees through relevant outputs |
| Dispatcher input preparation | integration | MMF-Q11; MMF-R06g | Actual effective inputs, selected record and skill/sequence predicates; supplied precedence/rates available | Complete remap/preparation and predicate-dependent skill paths | 2026-10-15 complete applicable dispatcher and movement/attack predicate readings; current HRC/NRF gates retained | Producer paths and original table/skill selection through selected values |
| Callback/timer/refresh execution | integration | MMF-Q12; MMF-R13c/MMF-R13e | Real cleanup/refresh/timer effects, not supplied callback logs/mode primitive | Complete cleanup and CSA-Q02/STL-Q07 dependency execution | 2026-10-15 full original freeze callback listing; cleanup disassembly assessed; CSA cold callback reused | Complete cleanup/free/removal/refresh and expiry-to-callback integration |
| Length producers | outside-boundary | MMF-Q01 | None locally; additional skill/missile cold/freeze producer contracts | Actual producer inputs outside movement | 2026-10-15 current DPO-R06b/R07 answers cold stat and conversion cases; no universal producer claim | Named producer contract for desired upstream case |
| Per-class cold effect / AI delay cells | outside-boundary | MMF-Q03 | None locally for supplied records; actual selected class inputs remain data-owned | Named actual class/column cells and compiled correspondence | 2026-10-15 reused GDF divisor/name records; no arbitrary class cold/delay cell claimed | Header/row-based effective loader reads of named needed cells |
| Later monster contributions | outside-boundary | MMF-Q04 | None locally for fixed initialization requests | Later class/property/equipment input production | 2026-10-15 MSD original same-function record plus fixed-write source slice resolves 68=100/67=75/69=100 | Named later producer through requested effects |
| Shatter corpse result | outside-boundary | MMF-Q07; MMF-R14b | None locally for draw/state request; death/corpse consumer result | Original consumer of active state 107 | 2026-10-15 CSA/GDF reuse resolves deadCol; no death-path discovery performed | Rooted death/corpse consumer contract |
| Gameplay confirmation | observation | MMF-Q09; MMF-O01/MMF-O02/MMF-O03/MMF-O04; MMF-V01 through MMF-V28 | Runtime-confirmation claims only; separate integration gaps still apply | Controlled matching observations of each verified behavioral rule | 2026-10-15 migration reused static records and retained four pending procedures; no gameplay performed | Permitted original 1.13c observations recording exact inputs, outputs, seeds and limits |

## Implementation gaps

The following ten findings are the preserved historical implementation comparison from the existing
record. They were not compared against current implementation during this migration. Its paths,
status statements and unresolved readings describe that historical comparison, not renewed evidence.
Only references have been updated to current ids, and a private-source absence phrase has been
normalized for public provenance. Current behavioral evidence and gates are in the section above
and the linked companion.

<!-- Gap list: original 1.13c behavior (docs/specs/state-movement-modifiers-and-freeze.md) vs current Dark Magic code. Input for a future implementation plan; not a plan. -->
# Gap items: state-movement-modifiers-and-freeze

Old-spec source of repo claims: `docs/research/MOVEMENT_COLLISION_AND_PATHFINDING.md`,
`internal/mod/d2legacy/adapter/movement/rules.go` (DM = Dark Magic). Old spec issue: #375.

| # | Rule (current spec) | Repo file/function or doc | Current repo behavior | Original behavior | Status |
|---|-----------------|---------------------------|-----------------------|-------------------|--------|
| 1 | MMF-R03, MMF-R06d/MMF-R06e (tick, 8.8, 16.16) | `adapter/movement/rules.go`; MOVEMENT_COLLISION_AND_PATHFINDING.md | 25 Hz, velocity 8.8 | 8.8 truncation confirmed; conversion to 16.16 path step unresolved | verified (8.8); 16.16 unresolved/gated |
| 2 | MMF-R03 | `rules.go`: `trunc(WalkVelocity * 256 * pct / 100)` | Same formula | Matches 1.13c for players (signed truncation toward zero; base = CharStats WalkVelocity byte << 8) | verified |
| 3 | MMF-R02 | `rules.go` percentage rule | Item FRW `item*150/(item+150)`, additive velocity stat, floor 25 | Matches (floor and ordering verified) | verified |
| 4 | MMF-R02/MMF-Q05, section 6 run/walk | `rules.go`: "run starts at `int(100*Run/Walk)`" (150 run baseline) | Run baseline = 100*Run/Walk | Not found: 1.13c function D2Common 0x33110 uses WalkVelocity and one pct for WL, RN, TW; RunVelocity (+0x41) only read by getter rva 0x2FD10 and charge-style skills (D2Game 0x26A80, 0x99620) | unresolved (DM claim unsupported) |
| 5 | MMF-R01 | `rules.go`: guard `item+150 != 0` | Guards divide by zero | No retail counterpart: retail divides by zero at v == -A | verified (derived, not runtime-observed) |
| 6 | MMF-R07 | `rules.go`: cold `velocitypercent = -50` | Fixed -50 for all targets | -50 is only the non-monster fallback; monsters use per-difficulty signed MonStats ColdEffect byte | verified |
| 7 | MMF-R04a/MMF-R05a | `rules.go` animation | WL/RN base rates 213/101, scaled by same pct as velocity | 213/101 constants match; same-pct scaling matches for movement modes only; attack/other modes use stats 0x44/0x45 with 15..175 clamp | verified |
| 8 | MMF-R08a through MMF-R12 | repo before this pass | No cold/freeze duration, immunity, freeze action-rate rules | Specified in rules 8-13 | verified (static) |
| 9 | Section 7 / Open questions | "Dependencies": broad cold/freeze content disabled until predicates and data values (divisors, ColdEffect, aidel) supplied from MPQ data | Disabled | Predicates unresolved; archive values not established in that comparison | gated |
| 10 | MMF-V07 | DM note: "do not reproduce in DM" divide by zero | DM avoids retail div-by-zero via guard | Retail divides by zero | verified (derived) |
