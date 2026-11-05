# Gaps: core-numeric-and-rng-foundations

Spec: [Numeric and RNG Foundations](../../specs/core-numeric-and-rng-foundations.md).
Evidence: [migration attempts and source records](../../research/spec-evidence/core-numeric-and-rng-foundations.md#foundation-migration-source-record-2026-10-14).

Status-application gaps (stored list effects, status timer-node and handler execution, status-path conditions
and status observations) are rows of the [status application gap file](combat-status-application.md). Gaps of
the record, stat-read, state and mode rules (wider runtime-condition meanings, state/update bookkeeping, list
free and attach integration and their observations) are rows of the
[record layouts and ids gap file](core-record-layouts-and-ids.md).

All rows were assessed on 2026-10-15 against their recorded source identities, methods,
actual boundaries and the current dependency contracts; see the
[reuse assessment](../../research/spec-evidence/core-numeric-and-rng-foundations.md#focused-regeneration-and-evidence-reuse-2026-10-15).
The 2026-10-14 attempts below retain their actual dates. Reuse does not renew verification,
complete a failed search or establish gameplay execution. The new seed-source/initializer
reading clarified NRF-R03a/NRF-R03b inputs and return scope; it did not locate a gameplay
no-parent creator, incoming game/player seed producer or a complete resource initializer.
IEP-R07/IEP-R13b/IEP-R14 to IEP-R18 provide known direct-draw slices only; NRF-Q04 continues
to block complete unread-handler/callee totals and full data bindings.

Restructure (2026-10-15): the rows were divided between this file and the record layouts and ids gap file
along the spec split; the animation/bookkeeping and gameplay-confirmation rows were each divided by the rules
they affect. No attempt, date or boundary changed.

## Open evidence gaps

| Gap | Kind | Affects | Blocked assertions / boundary | Missing evidence | Attempted (date, sources, result) | Would close it |
|---|---|---|---|---|---|---|
| Actual no-parent creator | outside-boundary | NRF-Q01, NRF-R03c, NRF-V05 | None locally: supplied clock/counter arithmetic verified; identifying a gameplay no-parent invocation remains outside the helper | Creator eligibility and actual samples | 2026-10-14: read D2Common `0x36760`, `0x265D` and `0x4DD0` fully; the seeder's null-parent branch is present; no gameplay invocation observed | Anchor a caller that supplies no parent and record its inputs |
| Live frame-rate policy | behavior | NRF-Q02, NRF-R09b, NRF-R09a, NRF-V12 | Whether the original live default can be changed after startup; supplied helper arithmetic/default constants remain eligible | Indirect initialization/rate-write caller or matching write observations | 2026-10-14: setter and recompute read; a raw D2Game code-section scan found 0 direct relative calls to `0x29800` or `0xF7970`; the import directories of the original D2Game, D2Client, Game, Fog, D2Launch and D2Multi found 0 imports of D2Game export 10026; prior pointer-slot evidence retained | Identify indirect consumer/startup invocation or valid rate-write captures |
| Complete coordinate/direction producers | integration | NRF-Q03, NRF-R10b, NRF-R11b, NRF-V15 | Complete producer routes/table recreation beyond the stated supplied path and two cited direction entries; getter extraction and velocity arithmetic are eligible | Static-position routes, all direction-entry/producer contracts | 2026-10-14: high-word getters read, direction writer `0xD240` and helper `0xCF40` read, all 1536 original constant-table bytes read; axis/diagonal scale confirmed | Specify each remaining producer's outputs and the complete direction-table contract under its owner |
| Callback-specific random totals | outside-boundary | NRF-Q04, NRF-R04c | Whole callback draw owner/count assertions; dispatcher with supplied callback contracts is separate | Per-handler path and data-to-handler binding | 2026-10-14: read bounded slot selector `0x9E290` and shared dispatcher `0x37030`; reviewed IEP/CBL contracts; dispatcher alone does not determine callback RNG | Accept handler-owned draw contracts for every required callback |
| Complete mode/search eligibility and draw totals | integration | NRF-Q05, NRF-R04b, NRF-V08 | Whole mode-change/stored-target operation eligibility and all callee draws; reached local draw sites remain eligible | Upstream helper contracts, data slot selection and other draws | 2026-10-14: independent instruction listings of `0x77240`, `0xB1820` and `0x335F0`; both monster direct sites and the searcher mask site read; source table records call boundaries | Trace upstream selection and named callee draw contracts from gameplay owners |
| Incoming game seed and resource initializer | outside-boundary | NRF-Q06, NRF-R03d, NRF-V06 | None for the supplied-seed one-step stage; default/new-game seed and complete resource initialization totals are unverified | Incoming seed producer and resource initializer contract | 2026-10-14: independent `0xCACC0` listing proves one direct step/copy/local/request; `0xCAB30` entry/calls and seed-writer candidates inspected | Accept producer/resource contracts; no inference from the chosen fixture seed |
| Animation preparation | integration | NRF-Q07, NRF-R06b, NRF-R13, RLI-R04a | Actual animation preparation/refresh and other reset writes, including the effects of the animation reset requested by RLI-R04a; the prepared object branch and supplied quotient inputs remain eligible | Complete animation dependency contracts | 2026-10-14: read `0x33110`, `0x32220`/`0x501E0`, the `0x335F0` object branch and the reconstruction-checked `0x3E980`; many animation context/callee effects remain outside the inspected prepared branch | Accept complete named animation owners and their gated paths |
| Player seed producer | outside-boundary | NRF-Q08, NRF-R12a, NRF-V16 | None for creation's player-skip assertion; initial/default/loaded player seed remains unverified | Player-load producer and reseeding routes | 2026-10-14: searched the original D2Game for calls to the seed-initialization import; inspected candidates `0x11490`, `0x101E0`, `0x104F0` and `0xB9A0`; `0x11490` is item creation, no complete player-load root found | Trace actual player load/init to its seed writes |
| Gameplay confirmation | observation | NRF-Q09, NRF-O01 to NRF-O03; verified local rules and their vectors | Runtime-confirmation claims only; static/supplied-input assertions remain eligible | Valid matching observations of the stated rule/vector cases | 2026-10-14: retained prior observation procedures and reviewed recorded idle/layout/table/global reads; no traces/actions/memory writes performed | Run a separately authorized gameplay/helper observation campaign and record actual matching coverage |


## Publication review (2026-10-23)

A fresh blind review found no new behavioral or integration gap beyond the NRF-Q01–NRF-Q09 rows above. NRF-V05(g)'s supplied clock providers distinguish the local counter-before-sampling order; they do not identify an actual no-parent gameplay creator or close the incoming-seed, callback, animation, live-rate or observation gates. No 1.14d comparison was performed.
## Implementation gaps

### Current audit — 2026-11-03

Baseline: local main `7e5e0138ee448c5ce42b3b3ca8b6ca2af7cd63b0`, read in the knockback-plan worktree. This is a source comparison for the [minimum numeric plan](../core-numeric-and-rng-foundations.md), not a test run or new original-game investigation. Open evidence gaps above are unchanged. Verified-but-deferred rules are not recategorized as unresolved.

| Rule/sub-id | Current repository comparison | Minimum-plan disposition |
|---|---|---|
| NRF-R01 | `internal/game/simulation/random.go:Stream` implements SplitMix64; `internal/runtime/lua/authority_rng_module.go` exposes named streams, not two-word LCG seeds. No shared NRF step exists. | Implement Lua-owned multiplier/step using generic exact native multiply/add |
| NRF-R02 | Current authority_random rejects nonpositive limits, uses registered streams/rejection sampling, and has no supplied-pair RandMask. `deterministic_module.go` hashes a keyed input instead. Neither is interchangeable with the spec. | Implement Rand/RandMask beside existing APIs; defer RandInRange |
| NRF-R03a | Named streams derive from world seed and purpose. No shared original child/unit seed initializer is established by these owners. | Deferred; supplied existing pairs suffice |
| NRF-R03b | Existing stream construction is not original InitTimerState/InitSeed ordering. | Deferred |
| NRF-R03c | No need to consume wall clocks/counters for bounded KNB fixtures; current stream constructor is not this operation. | Deferred verified supplied-clock arithmetic; Q01 creator still gated |
| NRF-R03d | Runtime config Seed initializes named streams, not the original supplied game-seed/resource request sequence. | Deferred; Q06 integration still gated |
| NRF-R04a | Current melee/damage/loot policies call purpose-named authority_random streams. A pure explicit-pair API can preserve argument isolation but cannot prove an original draw-site binding by itself. | Implement primitive API; leave KNB/HRC consumers to enforce their site ownership; other rows deferred |
| NRF-R04b | No corresponding shared original monster elemental/searcher seed-site implementation was established by the inspected random owners. | Deferred; Q05 remains gated |
| NRF-R04c | Dispatcher inspection cannot establish all callbacks' RNG totals. Current named-stream APIs do not supply the missing evidence. | Gated Q04 |
| NRF-R05 | `policy/action_rate.lua:trunc_div` is local and restricted to positive divisors; `policy/stat_resolution.lua` uses floor. Neither is the full ordered mulDiv contract. | Defer full mulDiv; implement generic wrap/division only for KNB notation |
| NRF-R06a | Local action-rate truncation and floor-based stat resolution do not establish all percent idioms. | Deferred; no caller migration |
| NRF-R06b | No shared unsigned multiply-high /100 policy identified in inspected numeric/RNG owners. | Deferred; do not add unused percent policy |
| NRF-R07 | `policy/stat_resolution.lua:resolve` resolves flat/percentage sources; it is not original diminishing math over RLI selected arrays. | Deferred; no new RLI dependency for this slice |
| NRF-R08a | Current raw damage uses 8.8 quantities while movement uses f64 positions; converting those stores is unnecessary for supplied path words. | Consume supplied 16.16 notation only; resource/storage migration deferred |
| NRF-R08b | Existing generic velocity/forced-motion data does not establish this original speed/marker setter. | Deferred |
| NRF-R09a | No frame-rate change is needed by pure local vector calls. Existing fixed-tick behavior is not migrated by this plan. | Deferred |
| NRF-R09b | Source inspection here does not resolve the original later live-rate producer. | Gated Q02 |
| NRF-R10a | KNB local endpoints accept path units directly; current forced-motion floats are not evidence of this unusual negative tile conversion. | Units documented; converters deferred |
| NRF-R10b | No supplied original-path high-word policy helper in inspected module/runtime arithmetic owners. World-space entity positions are a different representation. | Implement supplied-table getter through exact generic high_u16; Q03 other producers remain gated |
| NRF-R11a | Generic forced-motion admission normalizes a continuous ray; it is not the original acceleration/direction/fixed-point step update. | Deferred; leave movement unchanged |
| NRF-R11b | `policy/direction.lua` contains player presentation direction choices, not proof of the complete original fixed direction table. | Deferred; Q03 remains gated |
| NRF-R12a | Existing game Seed and monster bootstrap do not establish original unit-creation draw counts. | Deferred; Q08 player producer remains gated |
| NRF-R12b | No original item-secondary-seed contract established by named RNG construction. | Deferred |
| NRF-R13 | No prepared original object-animation reset branch established by inspected RNG/capability owners. | Deferred; Q07 remains gated |

### Shared implementation boundaries

- `internal/runtime/lua/authority_rng_module.go:maximumExactLuaInteger` documents the Lua 2^53-1 exact-integer limit. A direct `lo * 0x6AC690C5 + hi` in Lua cannot implement every u32 pair. A generic Go `engine.integer/v1` capability is planned; the multiplier and draw policy stay in Lua.
- `internal/mod/d2legacy/runtime.go:ConfigureModuleRuntime` is already shared by module test and authority compositions. `identity.go:authoritativeCapabilityVersions` and `identity_test.go:TestAuthoritativeCapabilityIdentityRejectsCompositionDrift` require the new capability to be pinned, not hidden behind a test loader.
- Existing ownership rules classify `internal/runtime/lua/` as a mod-neutral boundary and `d2legacy/policy/` as Diablo policy. No new random manager/store/package is required.
- Existing named-stream independence, snapshot/restore and registration tests in `internal/game/simulation/random_test.go`, runtime RNG tests in `authority_rng_module_test.go`, and keyed-purity tests in `deterministic_module_test.go` must retain their expectations. The new capability changes normal code/recipe identity, not archive schema or compatibility policy.
- No shared `policy/numeric.lua` or `engine.integer/v1` exists on the audited base. The plan creates them; no implementation is claimed by this audit. Broader rows above are comparisons of the relevant inspected owners, not exhaustive absence proofs across the repository.

## Implementation update — issue #675

The approved minimum slice is now implemented in this worktree: generic `engine.integer/v1`, Lua `numeric.step/rand/rand_mask/path_integer_position`, shared runtime composition and identity pinning, with focused tests for the selected NRF assertions and KNB scalar boundaries. Task-level red/green evidence and independent reviews are recorded in the ignored per-plan workspace; final repository verification is recorded in the implementation plan. The historical audit above still describes its explicitly named base commit.

This does not change original-game evidence status, resolve any open evidence row, implement deferred verified rules, create a durable unit-seed store, or establish physical knockback/1.14d parity. Current gameplay RNG implementations and their archive formats remain unchanged.
