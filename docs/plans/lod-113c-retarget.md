# Bounded 1.13c Retarget Plan

> **For agentic workers:** Use the applicable Superpowers planning and verification skills. This is a validation-boundary plan, not authorization to implement new gameplay. Steps use checkbox syntax for tracking.

**Goal:** Let #378 validate verified 1.13c rules through existing headless profiles without advertising or admitting playable 1.13c sessions.

**Architecture:** Reuse the existing Lua module/ECS test profiles to call production-owned policy at explicit supplied-input boundaries. Keep the current authority startup, game rules, named RNG streams, session admission and persistence contracts intact. Shared numeric and combat implementations remain separately owned prerequisites.

**Tech Stack:** Existing Go test runner, embedded d2legacy Lua, gopher-lua module/ECS profiles.

**Spec:** Engineering boundary supporting the approved [knockback specification](../specs/combat-knockback.md); this document establishes no additional original-game behavior.
**Issue:** Refs [#378](https://github.com/gravestench/dark-magic/issues/378). This scope is planning and validation preparation only; it does not close the issue or authorize a wider version migration.
**Status:** Retarget scope approved by the maintainer on 2026-11-03. The maintainer subsequently authorized the [minimum numeric-foundation implementation](core-numeric-and-rng-foundations.md) under #675. The entire NRF specification and a playable 1.13c target remain outside that authorization.
**Baseline:** Local main `7e5e0138ee448c5ce42b3b3ca8b6ca2af7cd63b0`, with the knockback planning documents in this worktree.

## Scope decision and alternatives

The chosen approach is headless verification using existing profiles. It allows static rule vectors to run without making unsupported claims about an entire game session. The current runner already provides this separation, so the minimum retarget step requires no new runtime feature, target flag, profile, generic compatibility layer or global RNG replacement.

An opt-in playable 1.13c target would also preserve the default, but would require a supported-behavior boundary across bootstrap, data, combat, admission and persistence. It is larger than the current request and deferred. Replacing the default target would conflict with preserving existing supported behavior and is excluded.

The word “retarget” here means selecting the evidence and validation boundary for the new rule tests. It does not change the product's supported target. In particular, do not set `game_rules.target` to `lod-1.13c` merely to run a local arithmetic test.

## Current code and resulting file boundary

Repository-relative paths below are audit entry points, not an instruction to modify them.

| Existing owner | What it already provides | Decision |
|---|---|---|
| `internal/mod/d2legacy/lua_suite_test.go:startLuaSuiteProfile` | `module` and `ecs` branches configure a bounded runtime without calling full `StartWithConfig`; `authority` calls full startup | Reuse these branches; no new profile |
| `internal/content/d2legacy/lua/d2legacy/tests/v1.lua:M.suite` | Explicit profile, tier, records, initial data and named cases | Declare existing metadata; do not invent evidence metadata that the harness does not validate |
| `internal/mod/d2legacy/lua_suite_test.go:runLuaAction` | Stepping, commands and checkpoint actions require the authority profile | Preserve this separation; module/ECS tests cannot claim authority recovery |
| `internal/content/d2legacy/lua/d2legacy/policy/game_rules.lua:build` | Defaults to and admits only `lod-1.14d` | Preserve unchanged |
| `internal/app/clientapp/session_bootstrap.go:sessionInitialData` and `internal/mod/d2legacy/adapter/entryworld/bootstrap.go:Prepared.InitialData` | Production bootstrap selects `lod-1.14d` | Preserve unchanged |
| `internal/game/simulation/random.go:Stream/RandomStreams` | Replay-stable named SplitMix64 streams | Preserve unchanged; these are not NRF-R01/R02 |
| `internal/game/simulation/runtime_recipe.go` and `internal/game/session/compatibility.go` | Runtime identity and admission/reconnect compatibility | No new identity fields or protocol changes for module fixtures |

This preparation supplies the validation boundary for the numeric foundation. The broader implementation plan remains with [issue #378](https://github.com/gravestench/dark-magic/issues/378); its local planning edits are outside the numeric implementation PR. No original-game spec is edited to fit the repository.

## Global constraints

- Evidence target for the new KNB vectors: Lord of Destruction Expansion 1.13c. Existing 1.14d records and observations retain their attribution.
- Use only verified rule assertions. Keep KNB-R09b/R10c and the spec's dependent effects gated.
- Preserve exact widths, truncation, seed words, draw owners and ordering specified by KNB and consumed numeric contracts.
- Synthetic fixtures contain supplied values; they are not observations, original save files, compiled-table correspondence or proof of a full data loader.
- Pure policy arguments and outputs may be transient tables. Any state that affects future authoritative ticks requires the existing ECS/registered-store ownership and recovery contracts before production use.
- No module-global mutable seeds, second simulation owner, copied RNG inside knockback, new test-only gameplay command, or alternate combat dispatcher.
- No game data, raw captures or private saves are committed.

## Acceptance levels

| Level | Accepted evidence | What remains unproved |
|---|---|---|
| Local rule | Named KNB vector calls actual policy in `module`; real shared arithmetic/RNG implementation where required | Session startup, automatic dispatch, live gameplay, full RNG budget |
| Supplied-input stage | Exact operations, outputs and logs at the spec's boundary; `module` or `ecs` only where existing capabilities suffice | Real execution/effects of supplied clip, planner, mode/state or downstream callbacks |
| Authoritative integration | Owning production transaction and explicit durable stores exercised with replay/restore under `authority` | Original-runtime confirmation and unresolved gameplay behavior |
| Playable 1.13c support | Not delivered by this plan | Requires a separately scoped and approved product target/admission/data/migration change |

Passing a local vector is useful acceptance in its own right. It does not require claiming the authoritative integration level. Conversely, restoring a fixture table manually is not a session checkpoint test.

## Minimum dependency order for #378

| Work | Minimum prerequisite | Unblocks |
|---|---|---|
| This retarget preparation | Review of this bounded scope; existing runner/profile contracts | Removes the supposed need for a new playable target before local KNB validation |
| Shared exact numeric operations | A separately approved foundation plan for NRF-R01/R02 and the wrapping/truncation/coordinate operations consumed by KNB; use its exact vectors | KNB item/missile draws and endpoint arithmetic; never substitute current SplitMix64 |
| Item stat and mode/state consumers | Required verified RLI operations and supplied record facts through their owning implementation | The corresponding KNB item, reaction and monster-start assertions |
| Shared reaction selection | Verified HRC selector through its owner | KNB fallback/conversion and selected-reaction tests with supplied terminal effects |
| Production combat/seed/path integration | Required SBA/MHD/DPO stages and durable owners, each scoped separately | KNB Task 6, actual invocation and replay/restore acceptance |

Do not make completion of every related spec a prerequisite to every local test. A missing dependency's real effects can stay supplied only where the vector explicitly permits that. The real NRF implementation is required for chance-vector seed assertions; a fixed-output random stub would not establish them.

The maintainer authorized the minimum verified numeric foundation consumed by #378, approved its execution and authorized creation of its owning issue #675; see the [bounded numeric plan](core-numeric-and-rng-foundations.md) for rule-level scope and verification. This does not approve the entire foundation specification or silently expand #378 into a wider foundation migration.

## Review focus

1. A module vector must not relax `game_rules.build` or relabel the production target.
2. A chance test must use the real shared two-word generator, not a mock or named SplitMix64 stream.
3. A supplied callback log must not be described as proof of its downstream effects.
4. A module/ECS fixture must not be presented as authoritative checkpoint/replay acceptance.
5. Existing 1.14d fixtures, saves, recipe digests and tests retain their behavior and attribution.

## Preparation tasks

### Task 1: Record the supported validation boundary

**Files:** this plan.
**Consumes:** approved KNB spec, existing runner branches and current product bootstraps.
**Produces:** the four acceptance levels and dependency order above.

- [ ] Review the module/ECS/full-authority startup branches and the narrower-profile action guards named above.
- [ ] Accept this scope only if no new target flag, playable admission, default change, source-data relabeling or persistence migration is needed for local vectors.
- [ ] Keep all remaining physical motion, item timing and patch-parity questions in the KNB gates, rather than inventing behavior here.

### Task 2: Correct the knockback planning holds

**Files:** `docs/plans/combat-knockback.md`, `docs/plans/gaps/combat-knockback.md`.
**Consumes:** Task 1's acceptance boundary.
**Produces:** accurate per-task readiness and a dated correction to the prior broad retarget hold.

- [ ] Link this plan as the validation-boundary prerequisite; remove the claim that a playable 1.13c retarget must land first.
- [ ] Retain real numeric/stat/selector dependencies for local tests. Keep production caller/identity/store integration as a later gate, not an implicit part of retarget preparation.
- [ ] Preserve all 19 KNB rule entries, all 36 vector references and existing evidence gates. Preserve the original audit baseline and date this refinement separately.

### Task 3: Verify the preparation without expanding its scope

**Files:** the three planning documents only; no harness or production source changes planned.
**Consumes:** Tasks 1–2.
**Produces:** reviewed documents and an honest verification record.

- [ ] Check relative document links and named source entry points against this checkout; run `git diff --check` and check new files too.
- [ ] Confirm every KNB rule ID and vector ID remains covered and that only the three planning documents changed.
- [ ] When validating the runner before later code work, run `go test -count=1 ./internal/mod/d2legacy -run '^TestLuaHarnessContract$'`. Its existing “authority actions reject narrower profiles” case checks module stepping refusal. This is an existing harness check, not proof of KNB rules, production preservation or 1.13c support.
- [ ] At later implementation time, run the relevant policy/shared-owner tests and required repository gates on the submitted revision. Do not add a mirror test merely to prove that this documentation repeats current source.

No new runtime implementation is required to complete this preparation if review confirms the existing bounded profiles suffice. Do not create empty adapters or policy scaffolding just to make the retarget step appear to contain code.

## Handoff and limits

Review this scope before implementing dependent code. The next useful deliverable is a bounded numeric-foundation implementation plan, followed by that owner's tested implementation. Production combat integration and #378 closure remain later work. This task changes documents only; no unit test, native-game experiment, compatibility migration or gameplay implementation is claimed by writing them.
