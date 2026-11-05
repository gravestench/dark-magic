# Numeric Foundation for Knockback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. The maintainer has authorized execution for issue #675; steps use checkbox syntax. Commits, pushes and PR publication remain unrequested.

**Goal:** Provide exact shared arithmetic and supplied-seed draws needed by #378's verified local rules without changing existing gameplay RNG or movement.

**Architecture:** A stateless, mod-neutral Go integer capability handles operations that Lua's floating-point numbers cannot safely perform at full width. First-party Lua owns the Diablo multiplier, seed step and draw policy. Explicit caller-owned seed pairs are arguments; no new random registry, durable store or unit-seeding path is introduced.

**Tech Stack:** Go, gopher-lua, existing versioned runtime capabilities and d2legacy module-profile suites.

**Spec:** [core-numeric-and-rng-foundations.md](../specs/core-numeric-and-rng-foundations.md), prefix NRF; integer notation in section 1. Consumer width assertions come from the approved [knockback spec](../specs/combat-knockback.md).
**Scope approval:** On 2026-11-03 the maintainer approved the retarget scope, requested this minimum numeric plan, then instructed implementation and authorized creating its focused issue. This authorizes the bounded implementation below, not the entire NRF spec. Commits, pushes and PR publication remain unrequested.
**Issue boundary:** [#675 — Add exact numeric primitives for verified 1.13c knockback rules](https://github.com/gravestench/dark-magic/issues/675), prerequisite for [#378](https://github.com/gravestench/dark-magic/issues/378). Implement only this issue's numeric slice; #378 remains open.
**Depends on:** Approved [bounded retarget plan](lod-113c-retarget.md). No record/stat/state, combat, animation, full movement or player-seed implementation must land to test these supplied-input primitives.
**Baseline:** Local main `7e5e0138ee448c5ce42b3b3ca8b6ca2af7cd63b0`; see the refreshed [implementation gaps](gaps/core-numeric-and-rng-foundations.md#implementation-gaps). No implementation or gameplay observations were performed during planning.

## Scope and preservation

Implement NRF-R01, the Rand and RandMask portions of NRF-R02, and the supplied path integer getter of NRF-R10b. Implement the generic word operations required by KNB-R08's integer notation. Keep RandInRange, all seed initialization/creation, other gameplay arithmetic and all production seed bindings out of this slice. Their absence does not prevent #378 from supplying an existing seed pair or path fixture.

The arithmetic API is general; it contains no Diablo identities or constants. The Lua module has no initializer that draws or registers streams. Existing melee, damage, loot, map generation, stamina and movement callers keep their current APIs and algorithms. No test expectations for their supported outputs are changed to accommodate the new generator.

The new capability and bundled policy bytes intentionally produce a new runtime identity. Register and pin the capability through the existing recipe owner; do not change recipe schema, protocol, save format or compatibility checks to force old identities to match. This preserves behavior on the new build, not cross-build checkpoint/network compatibility. Existing named-stream archives remain the same format. Unit-owned durable seeds and compatibility of future seed components belong to later integration, not this pure API.

## Global Constraints

- Evidence build: Lord of Destruction Expansion 1.13c. Every selected rule is **verified in 1.13c code (static)**; none is runtime-confirmed.
- Seed is `{lo,hi}`, both u32. `x = u64(lo) * 0x6AC690C5 + u64(hi)`; write both `lo = low32(x)` and `hi = u32(x >>> 32)`.
- Rand modulus is i32; masks/results are u32. Nonpositive modulus returns 0 without a step. Every positive modulus steps exactly once, including 1. RandMask always steps and uses `low32(m-1)`, including m=0 and m=1.
- Integers use two's complement. Except explicit widened products, addition, subtraction and multiplication keep the low 32 bits. Signed division truncates toward zero. Divide-by-zero and signed division overflow must not return a numeric result.
- Positions are supplied 16.16 bit patterns; integer getters return unsigned high 16 bits. A null path faults rather than returning zero. Five path units per tile is a unit definition, not a completed displacement measurement.
- Lua numbers cannot safely represent an arbitrary u64 product. Return its exact low/high u32 words separately; no packed seed/product is exposed as one Lua number.
- All future-tick state stays with existing explicit engine owners. No module-global seed, implicit unit seed initialization, wall clock, hidden counter or alternate session RNG. Checkpoint/restore and replay parity must still pass for existing supported behavior.

## Review Focus

1. Full-width multiply/add must preserve low bits and carry beyond Lua's exact integer range; Task 1 tests both words.
2. Nonpositive modulo draws skip mutation, while mask 0/1 and modulo 1 still step; Task 2 tests NRF-V02's selected cases.
3. Negative products and minimum-i32 absolute values wrap before signed truncation; Task 1 tests KNB-V34's arithmetic.
4. Supplied 16.16 high words are unsigned; Task 3 tests NRF-V15g and null-path refusal.
5. New primitives must not consume current named streams or alter production support; Task 4 checks identity composition and existing stream continuation.

## Files and responsibility

| Action / file | Responsibility |
|---|---|
| Create `internal/runtime/lua/integer_module.go` and `integer_module_test.go` | Stateless native word arithmetic and strict Lua scalar boundary; export `IntegerModule() Module` |
| Create `internal/content/d2legacy/lua/d2legacy/policy/numeric.lua` and `numeric_test.lua` | NRF step/draw policy and supplied path getter; no ECS ownership or production draw-site registration |
| Modify `internal/mod/d2legacy/runtime.go:ConfigureModuleRuntime` | Install IntegerModule once through the existing module composition used by headless and full authority |
| Modify `internal/mod/d2legacy/identity.go:authoritativeCapabilityVersions` | Pin `engine.integer` to `v1` |
| Extend `internal/mod/d2legacy/identity_test.go` | Assert the new capability is pinned and reject composition drift |
| Preserve `internal/game/simulation/random.go`, `internal/runtime/lua/authority_rng_module.go`, `deterministic_module.go` | Current named/stateful and keyed/stateless RNG remain unchanged |

Existing ownership prefixes already classify generic runtime adapters and Lua policy. Run the ownership gate; do not add a duplicate classification rule. No separate math package is required for these few bridge operations.

## Exact interfaces

`IntegerModule() Module` exports `engine.integer/v1`, `api=1`, with these pure functions. All arguments must be finite exact integers in their stated domain. Reject NaN, infinity, fractional, negative unsigned and out-of-range inputs before conversion; do not silently round or saturate them. These are capability-boundary safeguards, not claims about original game validation.

| Lua function | Argument domains | Return contract |
|---|---|---|
| `mul_add_u32(a,b,c)` | u32 × u32 + u32 | Two results `(lo,hi)` from exact u64 product plus c; the declared maximum cannot overflow u64 |
| `band_u32(a,b)` | u32, u32 | u32 bitwise AND |
| `sub_i32(a,b)` | i32, i32 | low32 subtraction interpreted as i32 |
| `mul_i32(a,b)` | i32, i32 | low32 multiplication interpreted as i32 |
| `abs_i32(a)` | i32 | wrapping absolute value; minimum i32 remains minimum i32 |
| `div_i32(a,b)` | i32, i32 | signed truncating quotient; raise a Lua error for divisor 0 or minimum-i32 divided by -1 |
| `add_u16(a,b)` | u16, i32 | low16 sum as u16 |
| `high_u16(bits)` | u32 | unsigned high 16 bits |

Do not hardcode 0x6AC690C5, 5, stat identities or coordinate policies in this Go module. Native helpers may be private implementation functions; the public boundary is the versioned Lua capability.

Lua `d2legacy.policy.numeric` exports:

- `step(seed) -> nil`: validate the supplied `{lo,hi}` pair, call the generic multiply/add with the Lua-owned multiplier, then store both returned words in that same supplied pair. No implicit initialization or copy to another seed.
- `rand(seed,n) -> u32`: validate arguments, skip for signed n<=0, otherwise step once and use the spec's power-of-two mask or unsigned remainder form.
- `rand_mask(seed,m) -> u32`: validate arguments, step once, then generic AND with `low32(m-1)`. Lua may compute this mask modulo 2^32 because the intermediate is within its exact range.
- `path_integer_position(path) -> x,y`: require a non-null supplied table with u32 `x_bits` and `y_bits`; return generic high_u16 of each. These field names describe the local supplied adapter, not a new ECS schema. It does not read a unit's current world position or invent an absent-path fallback. KNB's own source accessor supplies its separately verified absent-path behavior.

No RandInRange, InitSeed, unit factory, time sampler or velocity helper is exported in this slice. Knockback may adapt `numeric.rand`/`rand_mask` directly to its existing planned ops interface. Generic integer functions are reusable capability operations; do not reimplement their arithmetic in knockback.

## Complete NRF coverage and exclusions

S means **verified in 1.13c code (static)**; U means **unresolved**. “Deferred” means verified but outside this requested minimum, not evidence-gated. This table accounts for every rule/sub-id in the spec's Id map without pretending to implement the entire spec.

| Rule/sub-id | Evidence | Task / vector assertions or disposition |
|---|---|---|
| NRF-R01 | S | Tasks 1/2; NRF-V01 step words and actual Lua draws |
| NRF-R02 | S | Task 2; NRF-V01a–e, NRF-V02a/b/c/f/g, NRF-V03; RandInRange V02d/e explicitly deferred |
| NRF-R03a | S | Deferred: seed initializer/parent path; no task |
| NRF-R03b | S | Deferred: initializer ordering; no task |
| NRF-R03c | S | Deferred: supplied-clock arithmetic; actual no-parent creator remains Q01 |
| NRF-R03d | S | Deferred: game/resource initialization; Q06 integration remains gated |
| NRF-R04a | S | KNB item/missile rows inform caller-supplied pair choice; enforcement belongs to KNB-R01/R02a and HRC-R06 vectors, not a new foundation dispatcher. Other direct sites deferred |
| NRF-R04b | S | Deferred: monster slot/searcher draw sites; Q05 integration remains gated |
| NRF-R04c | U | Gated Q04: per-handler complete seed owners/totals; no task |
| NRF-R05 | S | Deferred: full mulDiv branch algorithm; Task 1 tests only generic signed division/product notation, not NRF-R05 acceptance |
| NRF-R06a | S | Deferred: percent-specific idioms; generic truncation alone does not establish this whole rule |
| NRF-R06b | S | Deferred: unsigned multiply-high percent operation |
| NRF-R07 | S | Deferred: diminishing returns and real stat getter |
| NRF-R08a | S | 16.16 format consumed only by Task 3/V15g; 8.8 resources deferred; no conversion or health-store migration |
| NRF-R08b | S | Deferred: speed setter |
| NRF-R09a | S | Deferred: frame-rate setter/default policy |
| NRF-R09b | U | Gated Q02: later live-rate changes; no task |
| NRF-R10a | S | Five-units-per-tile convention retained; conversion functions/V13 deferred because KNB accepts path coordinates, not tile conversion |
| NRF-R10b | S | Task 3; NRF-V15g getter assertion; null-path refusal from rule; Q03 complete producer integration remains gated |
| NRF-R11a | S | Deferred: velocity ramp and step calculation |
| NRF-R11b | S | Deferred: direction scale/entries; Q03 complete producer/table integration remains gated |
| NRF-R12a | S | Deferred: creation seeding; Q08 player seed producer remains gated |
| NRF-R12b | S | Deferred: item secondary seed |
| NRF-R13 | S | Deferred: prepared object animation; Q07 dependency effects remain gated |

The generic arithmetic tests also cite KNB-V23/V29/V34's exact scalar inputs to prove that the foundation can express its consumer's integer domain. They do not implement KNB endpoint generation or claim those complete KNB vectors passed.

## Task 1: Exact native word operations

**Files:** create integer_module.go / integer_module_test.go in the paths above.
**Consumes:** Go fixed-width integer semantics; no RNG/state owner.
**Produces:** IntegerModule and the eight scalar functions above.

- [x] Write failing `TestIntegerModuleNRFStepWords`: invoke the Lua capability through a real Runtime, using NRF-V01's multiply/add stages. `(1,0x6AC690C5,666)` returns `(0x6AC6935F,0)`; `(0x6AC6935F,0x6AC690C5,0)` returns `(0x2F2ED81B,0x2C890AFD)`. This tests arithmetic only; Task 2 tests actual policy sequencing.
- [x] Add `TestIntegerModuleKNBWidthBoundaries`: from KNB-V34, sub_i32(0,-2147483648) and abs_i32(-2147483648) each return -2147483648; mul_i32(3,-2147483547) returns -2147483345; dividing that by 2147483547 returns 0; add_u16(65535,3) returns 2. From KNB-V23, div_i32(-6,7) returns 0 rather than -1. Assert band_u32(0x6AC6935F,99)=67 and high_u16(0xFFFF0000)=65535.
- [x] Add `TestIntegerModuleRejectsInexactInputs`: table-driven fractional, NaN, infinity, unsigned -1, u32 4294967296 and i32 2147483648 calls must error. Divisors 0 and the minimum-i32/-1 pair must error without a numeric result. Also pin full generic precision with mul_add_u32(4294967295,4294967295,4294967295) returning `(0,4294967295)`; this is an engineering boundary test, not an original-game vector.
- [x] Run `go test -count=1 ./internal/runtime/lua -run '^TestIntegerModule'`; confirm the new tests fail for the missing capability before implementing it.
- [x] Implement IntegerModule with checked scalar domains, native fixed-width operations and two-word returns. Calculate signed wrapping via explicit unsigned conversions where necessary; do not expose a rounded u64 Lua number or rely on native divide overflow returning a value.
- [x] Re-run the same tests, including exact return arity and errors. No Diablo policy is added to Go.

## Task 2: Shared supplied-seed RNG policy

**Files:** create policy/numeric.lua and numeric_test.lua; wire IntegerModule in runtime.go and identity.go as part of this task, so the policy is tested through the actual shared module composition.
**Consumes:** Task 1's integer capability.
**Produces:** step, rand, rand_mask; caller-owned explicit seed pair only.

- [x] Add module/fast cases named `NRF-V01`, `NRF-V02-selected`, `NRF-V03` in the existing Lua suite format, with the exact starting pairs from the spec and fresh fixtures per variant. Assert both words and every result; do not mock the native capability or the RNG step.
- [x] Pin V01: from `{1,666}`, Rand(100)=51, RandMask(128)=95, Rand(2)=1, and RandMask(100)=67 independently, each ending `{0x6AC6935F,0}`. From `{0x6AC6935F,0}`, Rand(100)=31 and final pair `{0x2F2ED81B,0x2C890AFD}`. Also call step directly and assert these same two transitions.
- [x] Pin V02a/b: Rand(0/-5)=0 without mutation; V02c Rand(1)=0 with one step; V02f/g RandMask(0)=1791398751 and RandMask(1)=0, both with one step from `{1,666}`. V03 starts `{0x6AC6935F,0}`, Rand(64)=27 and writes the second transition. V02d/e are not executed or marked covered.
- [x] Run `go test -count=1 ./internal/mod/d2legacy -run '^TestLuaSuites$/^d2legacy/policy/numeric$'`; confirm failures reach missing numeric policy/capability before implementation.
- [x] Implement the three Lua functions, install IntegerModule in ConfigureModuleRuntime, and pin `engine.integer: v1` in authoritativeCapabilityVersions. Do not add a new runtime profile, startup seed or bootstrap target.
- [x] Re-run the focused suite and Task 1 tests. Keep an independent pair `{9,0}` unchanged while drawing the supplied pair; this proves argument isolation, not automatic attacker/missile binding. Future KNB tests enforce actual consumer ownership.

## Task 3: Supplied path integer getter

**Files:** modify numeric.lua / numeric_test.lua.
**Consumes:** Task 1 high_u16; NRF-R08a/R10b supplied 16.16 notation.
**Produces:** path_integer_position(path), no path writes.

- [x] Add failing module case `NRF-V15g`: supplied x_bits=0x000C8000, y_bits=0xFFFF0000 returns `(12,65535)` with path and every seed unchanged. State explicitly that V15a–f/h are not covered.
- [x] Add a rule-boundary case for null path: the operation errors without returning coordinates, rather than supplying `(0,0)`. This tests R10b, not KNB's different absent dynamic-source-path branch.
- [x] Run the focused numeric suite and confirm failure on the new getter assertions.
- [x] Implement the supplied-table getter through high_u16, with no tile conversion, coordinate policy, live entity lookup or position write.
- [x] Re-run numeric and integer suites. Do not change existing movement coordinate handling or use this helper to reinterpret current floating-point world positions.

## Task 4: Composition, preservation and deterministic continuation

**Files:** extend identity_test.go and numeric_test.lua; retain existing simulation/random and runtime RNG tests.
**Consumes:** Tasks 1–3 and existing capability identity checks.
**Produces:** verified shared composition and no new persistent state owner.

- [x] Extend TestAuthoritativeCapabilityIdentityRejectsCompositionDrift to assert the actual recipe pins engine.integer v1, and the actual module composition includes it. Continue rejecting missing/unpinned capabilities; do not weaken validation or hardcode an old digest.
- [x] Add `supplied_pair_continuation`: capture the two scalar words after NRF-V01's first step, pass an independent pair with those same words back to the same real policy and assert NRF-V01d/NRF-V03 continuation. This is pure value continuation, explicitly not an authoritative checkpoint test. No new ECS component/store is required by a pure argument API.
- [x] Run the identity test and existing TestAuthorityRandomModuleUsesCheckpointedNamedStreams, TestAuthorityRandomModuleRejectsUnregisteredPurpose, TestDeterministicDrawIsPureAndPurposeSeparated, and simulation random tests. New arithmetic must not advance registered streams. Preserve their current output expectations and archive format.
- [ ] Run `make test-lua-hardening`, `make test`, `make architecture`, `make test-race` on the submitted implementation revision. Report exact commands, failures and skips. Existing supported replay/restore coverage must pass; no new old-build identity compatibility is promised.
- [x] Request a fresh independent review of native precision, Lua policy ownership, capability pinning, partial vector claims and existing RNG preservation before merging. Report this as partial NRF acceptance and Refs #378, never full foundation completion or #378 closure.

Focused preservation commands for this task:

```sh
go test -count=1 ./internal/mod/d2legacy -run '^TestAuthoritativeCapabilityIdentityRejectsCompositionDrift$'
go test -count=1 ./internal/runtime/lua -run '^Test(AuthorityRandomModuleUsesCheckpointedNamedStreams|AuthorityRandomModuleRejectsUnregisteredPurpose|DeterministicDrawIsPureAndPurposeSeparated)$'
go test -count=1 ./internal/game/simulation -run '^Test(NamedStreamsAreIndependentAndRestorable|Uint64nIsBoundedAndZeroSafe|RandomStreamsRestoreExactNamedSequences|RandomStreamsRejectUnknownAndChangedRegistrations)$'
```

All must report PASS on the implementation revision. On failure, investigate the changed boundary; do not rewrite the existing RNG contracts to make this additive capability pass.

## Gated and deferred work

- NRF-Q01/Q06/Q08: actual parentless creator, incoming game/player seeds, resource initialization, player loading and reseeding remain gated. Supplied seed tests do not establish these producers.
- NRF-R04c/Q04 and Q05: full callback/monster-mode eligibility and callee draw totals remain gated. No dispatcher or blanket “zero extra draws” assertion is added.
- NRF-Q03: complete static-position routes and direction-table producers remain gated; the supplied high-word getter is independently eligible.
- NRF-R09b/Q02 and NRF-Q07: live frame-rate changes and complete animation/dependency effects remain gated. No timing, movement or animation implementation is scheduled.
- NRF-Q09 and NRF-O01–O03: runtime confirmation remains pending; no owned-game campaign or 1.14d parity claim is included.
- Verified but unneeded rules and RandInRange are deferred by scope, as the coverage table records. Do not call them unresolved, implement them “for completeness,” or mark their mixed vectors passed.
- Unit-owned durable seed storage, current combat caller replacement and KNB's RLI/HRC/SBA/MHD/DPO integration remain outside this foundation slice. They must land before production knockback, even after these numeric primitives pass.

## Handoff

Execution was approved and the bounded implementation is complete locally; broad verification limits are recorded below. Its implementation unblocks the numeric portion of #378, not every knockback dependency. Plan-authoring validation is limited to source/contract review, coverage, file links and whitespace; no new code, test run, capture, commit or publication is implied by this document.

## Execution verification — #675

Implemented in this worktree on 2026-11-03, with no commits or publication. Tasks 1–3 used missing-implementation RED followed by final-byte GREEN. Task 4's new regression assertions passed immediately against the prior tasks' implementation; no artificial RED was manufactured. Four fresh task reviews passed spec compliance and code quality.

| Check | Actual outcome |
|---|---|
| Integer bridge focused tests | PASS |
| Selected NRF Lua suite (11 cases after Task 4) | PASS |
| Actual capability composition/identity | PASS |
| Existing named/keyed RNG and stream continuation tests | PASS |
| Lua syntax, both source trees | PASS |
| Architecture/ownership direct Go target | PASS |
| Go formatting / whitespace | PASS |
| Repeated Lua fast/integration suites, 3 repeats, seed 8675309 | PASS, 37.422s |
| Full `go test -count=1 -json ./...` | FAIL: 76 failed-test events in 22 failing packages; 65 skipped-test events. Main d2legacy package passed, 44.90s. Includes sandbox rename denials, Windows permission assumptions, shell startup denial and a zero-time render diagnostic assertion |
| Representative baseline comparison | A clean archive of baseline HEAD reproduces envconfig/settings rename denials and shell startup denial. The isolated baseline render diagnostic check passed; its full-run timing failure remains an unverified broader-suite limitation |
| Workspace-local TEMP/TMP workaround | Removes the rename error in the selected envconfig case, which then fails its existing Unix mode expectation on Windows. No unrelated code changed |
| Race | Unavailable: actual race command reports that -race requires cgo (CGO_ENABLED=0); no C compiler on PATH |
| make / stylua format gate | Tools absent. Available Go equivalents and syntax ran; StyLua check remains unverified |
| Owned game data / original runtime | Not exercised; no data configured or observation campaign performed |

The broad/race/formatter gates are not claimed green. Exact commands and outputs are retained in the ignored per-plan validation workspace. All evidence refers to these uncommitted working-tree bytes, not a submitted commit. The original-game specs and their confidence labels are unchanged.

Final whole-change review: spec compliance and source quality PASS, no actionable findings. Repository verification remains incomplete; no merge/publish readiness is claimed.
