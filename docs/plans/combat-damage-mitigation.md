# Physical flat-reduction ordering implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development after approval. Steps use checkbox syntax for tracking.

**Status:** Implemented and independently reviewed 2026-11-02. Required broad validation is not fully green; see the delivery record below. The approved re-target prerequisite exception applies only to this slice.
**Issue:** Refs #383 (G8 combat ordering); does not close #383.
**Goal:** Apply an existing raw physical reduction before percentage resistance in the shared production damage path.
**Architecture:** Keep Diablo policy in `policy/mitigation.lua`. Retain the current defense component, stat-source pipeline, `damage.resolve` consumer and health storage. Add no state, capability, component field, flag or parallel resolver.
**Tech Stack:** Go-hosted authoritative Lua; existing Lua suite harness.
**Spec:** `docs/specs/combat-damage-mitigation.md`, DMI; numeric dependency `docs/specs/core-numeric-and-rng-foundations.md`, NRF-R05/NRF-R08a.
**Base inspected:** `7dcf7da5c29c1cdd9dc0ba347be6ea0742b2c21e` (current main when prepared).
**Depends on:** No missing engine mechanism for this supplied-input ordering correction. The local implementation-planning skill requires a 1.13c re-target plan before gameplay plans; no such plan was found. The maintainer explicitly approved the narrowly scoped prerequisite exception on 2026-11-02; it does not waive the prerequisite for later gameplay plans.

## Design and approaches

The selected direction is implementing the existing mitigation spec under #383. Inspection found several separate behavior changes, so this plan proposes its first independently testable ordering correction.

1. **Recommended first slice:** Change only physical flat-reduction placement. The existing raw reduction is already produced by stat sources and used in the shared melee/missile damage resolver. This proves one concrete combat-ordering correction without changing combat state or patch selection.
2. **Broader alternative:** Implement the complete twelve-entry mitigation stage, including classification, bypass, pierce, scaled reductions, absorb and exact integer branches. That requires new inputs and integration across owners and is a larger design than the first slice.

The bounded operation is `max(0, floor((raw - reduction_raw) * (100 - selected_resistance) / 100))`, preserving the current resistance selection and final clamp. It corrects ordering only. The spec's full signed-output behavior, exact i32 overflow, helper branches and resistance-selection rules are deliberately deferred. The floor matches truncation for the positive, non-overflowing selected vector; this is not a substitute for NRF-R05 over its full domain.

No source conversion changes: `physical_reduction_raw` is already in raw damage units. DMI-V01 supplies stat 34 as 5 whole points; the fixture must therefore supply `5 << 8 = 1280` to this existing raw field. Never shift that field a second time.

## Global constraints

- Evidence build: Lord of Destruction Expansion 1.13c.
- Life and damage magnitudes use signed 8.8: raw 256 represents 1.0 (NRF-R08a).
- The full spec uses i32 two's complement, wrapping arithmetic, signed comparisons and division truncating toward zero. This slice claims only DMI-V01's positive, non-overflowing arithmetic, where the existing numeric operations agree.
- DMI-R10: mitigation itself draws no random number; preserve the random owner and stream state.
- Keep checkpoint/restore and replay behavior deterministic. No snapshot schema change.
- Preserve behavior outside physical reduction ordering, including current caps, bypass omissions, absorb policy, duration handling and health quantization. List their remaining gaps explicitly.
- Do not claim 1.14d equivalence, original-game runtime confirmation, full DMI acceptance or completion of #383.
- Do not commit, push or post to GitHub without a separate request.

## Coverage table

All DMI-R01 through DMI-R10 are labeled **verified in 1.13c code (static)** in the spec. Rows below distinguish the exact subset delivered from verified behavior deferred by scope; deferral is not an evidence gate.

| Rule | Owning task / implementation test | Assertion delivered or disposition |
|---|---|---|
| DMI-R01 regular-monster predicate | Deferred | No class/hireling classification changes. |
| DMI-R02 flat-reduction setup | Task 1 fixture mapping | DMI-V01 supplied raw reduction 1280 only; no implementation claim for stat-34 loading, shifts, scaling or overflow. |
| DMI-R03 event-11 placement | Deferred | Existing event registration/dispatch is not implemented here. |
| DMI-R04 class tests | Deferred | No undead/demon lookup changes. |
| DMI-R05 bypass | Deferred | No bypass inputs or policy added. |
| DMI-R06 per-type order | Tasks 1-2, `DMI_V01_physical_reduction_before_resistance` | Physical 5120, raw reduction 1280, resistance 25 produces 2880. Other eleven entries and negative stored outputs remain deferred. |
| DMI-R07 effective resistance | Task 1 uses supplied selected resistance 25 | Its selection is unchanged; caps, pierce, difficulty and sanctuary changes remain deferred. |
| DMI-R08 absorb | Deferred | No absorb behavior changes. |
| DMI-R09 mulDiv argument order | Task 1 limited DMI-V01 assertion | 75 * 3840 / 100 = 2880 in the narrow positive branch only. General NRF-R05 implementation and DMI-V13/V26 are deferred. |
| DMI-R10 no draws | Task 2 restored continuation and source review | Preserve no-call structure. DMI-V14 original-seed assertions are deferred; deterministic restoration alone does not prove no draws. |

DMI-V01's `physical` assertion is the selected mitigation-boundary assertion; the helper exposes no absorb accumulator, so no such assertion is claimed. The production fixture can also observe immediate total 2880, but that total belongs to DPO-R12a, not DMI acceptance. Existing player whole-health quantization remains a Dark Magic boundary, not original 1.13c life-storage parity.

## Review focus

- Raw units: 5 whole reduction points become 1280 raw exactly once; test both the pure helper and stat-source consumer.
- Unchanged channels: the existing mixed physical/fire fixture must retain fire 500 while physical changes from 730 to 735.
- Non-positive remainder: retain the current final clamp, including zero and over-reduction; signed-record parity remains deferred.
- Health quantization: inspect mitigated raw output separately from applied player whole-health damage.
- Restore: use the same supplied reduction and resistance after restore; no random state or hidden Lua state may change.

### Task 1: Correct physical reduction placement with the exact vector

**Files:**
- Modify `internal/content/d2legacy/lua/d2legacy/policy/mitigation.lua` (`M.apply`, physical branch only).
- Modify `internal/content/d2legacy/lua/d2legacy/policy/mitigation_test.lua`.

**Interfaces:** Preserve `M.apply(raw, channel, defense, monster_target) -> raw_number`.

- [x] Add `DMI_V01_physical_reduction_before_resistance` using the existing module suite. Supply physical=5120, physical_reduction_raw=1280, physical_resist=25, no bypass, no scaled reduction and no absorb. Assert physical=2880. The helper accepts supplied inputs and does not establish the original attacker-class predicate or callback walk.
- [x] Run `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/policy/mitigation$' -count=1`; record the new vector failing with actual 2560 versus expected 2880.
- [x] Move flat physical subtraction before percentage resistance. Retain the current resistance clamp and final `max(0, ...)`; add no field or helper abstraction.
- [x] Update `clamps_resistance_and_preserves_integer_order`: raw=1000, reduction_raw=20, resistance=25 now yields 735 rather than 730. Preserve its elemental expectations.
- [x] Add compatibility checks for zero reduction (1000, 0, 25 -> 750), zero resistance (1000, 20, 0 -> 980), exact reduction (1000, 1000, 25 -> 0) and over-reduction (1000, 1280, 25 -> 0). These are engine regression checks, not new original-game vectors.
- [x] Re-run the focused mitigation suite and record its actual result.

### Task 2: Prove the shared authority consumer and restored continuation

**Files:**
- Modify `internal/content/d2legacy/lua/d2legacy/policy/damage_test.lua`.
- Use existing `test.restore_checkpoint()` and `test.expect_checkpoint_parity(1)` actions; no harness modification.

**Interfaces:** Preserve `damage.resolve(target, rolled_bundle, ecs) -> result_record` and the existing stat-source/defense fields.

- [x] Update existing assertions in `derived_mitigation_and_lethal_damage_vectors`: physical helper result 735; mixed result physical=735, fire=500, damage_raw=1235, remaining_health_raw=3765 for a monster starting at 5000. No other expected value changes without tracing its source.
- [x] Add a separate authority case named `DMI_V01_production_physical_order`. Enter a player with health=max_health=50, add normal stat sources physical_resist=25 and physical_reduction_raw=1280, then step the derived-stats system so its stored defense values enter the latest checkpoint. Assert those exact stored defense inputs before calling the production `damage.resolve` with physical=5120.
- [x] Assert rolled.physical=5120 and mitigated.physical=2880. With current whole-health storage, assert damage_raw=2816 and remaining_health_raw=9984 (39 whole health). Clearly distinguish those engine compatibility values from the spec's raw mitigation result.
- [x] After the derived-stat step has created its checkpoint, run the damage assertion callback once, without stepping afterward. Call `test.restore_checkpoint()` to restore that latest pre-damage checkpoint, reacquire the player entity and assert health=50 plus the same defense inputs, then invoke the same assertion callback again. Both timelines must give mitigated=2880 and health=39. Then `test.step(1)` to checkpoint the post-damage state and `test.expect_checkpoint_parity(1)` to compare subsequent continued/restored checksums. This proves fixture-driven production-resolver continuation, not command-log replay of the direct test callback and not unchanged original RNG seeds.
- [x] Run the focused suites, `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/policy/(mitigation|damage)$' -count=1`, including the added restore check. Record actual failures and fixes.

### Task 3: Validate and review the delivered boundary

**Files:**
- Update this plan's execution record and `docs/plans/gaps/combat-damage-mitigation.md` implementation comparison only.

- [x] Run `go test ./...` (the `make test` recipe): exit 1, failures in eight untouched packages; gameplay and all d2legacy adapters passed.
- [x] Run `go test ./internal/acceptance -run 'Test(Retired|LegacyRenderer|NoAccidental|DependencyDirection|CommandRemains|GameplayOwnership)'` (the architecture recipe).
- [x] Attempt `go test -race -p 1 -parallel 1 -timeout 20m ./...` (the race recipe).
- [x] Run available Lua checks: syntax PASS; repeated/seeded hardening Go recipe PASS (100.699s). make/stylua are unavailable, so formatting and the full make target were not run.
- [x] Perform a separate fresh review of production reachability, unit conversion, mixed-channel assertions, replay/restore and all claims. The local skill requests subagent-driven execution/review; use the applicable skills at execution time.
- [x] Refresh only the delivered ordering row in the implementation comparison. Keep all other gaps open and record the exact checked commit and commands.
- [x] Prepare every PR-template section with `Refs #383`, evidence limits, checks and AI use. Do not submit or commit unless asked.

## Gated and deferred work

Evidence observations DMI-O01..DMI-O03 remain pending; they limit runtime-confirmation claims, not the verified supplied-input arithmetic. DMI-Q05's unknown reductionScale producer blocks no local supplied-input assertion and is irrelevant when the supplied scale is zero. Dependency-backed event and original stat-array integration are not tested here.

Verified but outside this slice: the complete twelve-entry loop, physical cap 50, elemental caps, hireling treatment, negative stored outputs, all overflow/rounding branches, scaled reduction, difficulty/pierce/sanctuary rules, bypass, percent/flat absorb and callback integration. DMI-V02..DMI-V27 are not claimed as passing. In particular, retaining the final clamp does not reproduce DMI-V08, and retaining the current cap does not reproduce DMI-V03c.

## Execution record

Before edits, at base `7dcf7da5c29c1cdd9dc0ba347be6ea0742b2c21e`:

- `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/policy/(mitigation|damage)$' -count=1` — PASS, package reported 0.508s.
- At that pre-edit point, broad checks had not run; current results are below.
- At that pre-edit point no implementation had changed. No commit, push or GitHub write has been performed.

### Delivered working-tree change and validation (2026-11-02)

The physical branch of `policy/mitigation.lua:M.apply` now subtracts the existing raw
flat reduction before resistance, preserving current resistance selection and final clamp.
DMI-V01's selected physical assertion is implemented: `(5120 - 1280) * 75 / 100 = 2880`.
The helper and production-authority tests pass; reverting the ordering makes both fail.
The authority test verifies stat-source derivation and reruns the same production resolver
assertions after restoring the pre-damage checkpoint. The later checkpoint comparison proves
continued/restored state equality, not command-log replay of the direct callback.
Both tasks and the whole code/test change received independent review with no actionable
findings; the coordinator also inspected the complete code/test diff.

Checks ran on this uncommitted diff over `7dcf7da5c29c1cdd9dc0ba347be6ea0742b2c21e`, not on a submitted commit:

| Command / check | Actual result |
|---|---|
| `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/policy/mitigation$' -count=1` | RED before implementation (2560 vs 2880); GREEN afterward, 0.568s |
| `go test ./internal/mod/d2legacy -run 'TestLuaSuites/d2legacy/policy/(mitigation\|damage)$' -count=1` | PASS on final restored files, 4.960s; separate old-order sensitivity run made the authority regression fail |
| `go test ./...` | FAIL, exit 1; `internal/mod/d2legacy` passed in 79.771s and all reported d2legacy adapters passed |
| `go test ./internal/acceptance -run 'Test(Retired\|LegacyRenderer\|NoAccidental\|DependencyDirection\|CommandRemains\|GameplayOwnership)'` | PASS, 1.116s |
| `go run ./internal/dev/tools/lua_syntax internal/content/d2legacy/lua internal/content/ds1editor/lua` | PASS, exit 0 |
| `go test -race -p 1 -parallel 1 -timeout 20m ./...` | Stopped before tests: CGO is disabled; no C compiler found |
| Lua formatting / make targets | make and stylua unavailable; equivalent Go recipes used where possible; no formatting pass claimed |
| Owned game-data/runtime checks | Not run: neither MPQ environment variable is configured; no original-game observation claimed |

The full-suite failures were in untouched packages: `internal/app/envconfig`,
`internal/app/realm`, `internal/app/serverapp`, `internal/dev/tools/bik_view`,
`internal/dev/tools/d2legacy_lua_test`, `internal/dev/tools/env_config`,
`internal/dev/tools/mpq2file`, and `internal/runtime/lua`. Observed failures concerned
owner-only/file-mode requirements, executable-lookup wording, backslash-versus-slash paths,
and a render decode-duration assertion reporting zero. No independent clean-checkout broad
baseline was run; these are observed untouched-path failures, not a claim that every cause
has been proven. Exact failed tests and full output are retained in this plan's ignored
SDD workspace alongside the completed PR-template draft and reviews.

No full DMI acceptance, original runtime confirmation, 1.14d equivalence or closure of #383
is claimed. All remaining scope/evidence limits above stand. No commit or push was made.

Repeated Lua hardening also passed: with `DARK_MAGIC_LUA_TEST_TIERS=fast,integration`,
`DARK_MAGIC_LUA_TEST_REPEAT=3`, and `DARK_MAGIC_LUA_TEST_ORDER_SEED=8675309`,
`go test -count=1 ./internal/mod/d2legacy -run 'TestLua(Suites|HarnessContract)'` returned
exit 0 in 100.699s. This is the repeated/seeded Go portion of hardening; the complete
make target and formatting remain unrun because make/stylua are absent. Final
`git diff --check` passed (Git emitted CRLF normalization notices only).
