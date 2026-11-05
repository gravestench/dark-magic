# Knockback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task after maintainer review and prerequisite delivery. Steps use checkbox syntax for tracking. Do not execute this plan as part of its authoring.

**Goal:** Implement the verified 1.13c knockback decisions and supplied-input path stages from the approved spec, with exact KNB vector coverage, while leaving unresolved integration gated.

**Architecture:** First-party d2legacy Lua owns the knockback policy. Existing engine stores own seeds, reaction facts and path state; existing shared combat, state and movement owners supply their operations. No parallel RNG, combat dispatcher, state registry, movement store or per-skill knockback system is introduced.

**Tech Stack:** Go engine/session/ECS, embedded Lua through gopher-lua, existing d2legacy Lua suite harness.

**Spec:** [combat-knockback.md](../specs/combat-knockback.md), prefix KNB.
**Issue:** [#378](https://github.com/gravestench/dark-magic/issues/378), parent #345, milestone 7. Refs #378; do not close it.
**Approval:** Maintainer approved the spec and planning on 2026-11-03, then approved Task 2 execution on 2026-11-05. Other tasks remain outside this implementation.
**Audit baseline:** refreshed 2026-11-05 against worktree/PR head `85c56f14e3f62cf54439c5fb4ad9fabe565f963f`, based on main `180c9f9bf08dc1bb429e258f8ba91895f4063d0b`. The initial 2026-11-03 comparison against `7e5e0138` is retained as history in the [gap audit](gaps/combat-knockback.md#implementation-gaps).

## Acceptance boundary and execution readiness

The deliverable is partial static acceptance of KNB-V01–KNB-V36 at the spec's isolated-rule and supplied-input-stage boundaries. It is not completed physical knockback, runtime confirmation, or 1.14d equivalence. The issue still asks for 1.14d chance/distance/speed/eligibility observations and blocked/interrupted behavior; none is settled by these 1.13c vectors.

**Execution readiness is per task.** The approved [bounded retarget scope](lod-113c-retarget.md) permits isolated-rule and supplied-input-stage validation. [PR #676](https://github.com/gravestench/dark-magic/pull/676) implements #675's numeric foundation at the audited head. Both applicable workflows and all five jobs passed, including full Go tests, race, Lua hardening/StyLua, architecture and three-platform mod loading. The PR merged on 2026-11-05 as `de6f9088225f9a022b3b4d9bd50d07e61a6d858c`, after an independent clean review and fresh verification of all five CI jobs on its unchanged head. The numeric dependency is now on main.

Remaining stat/state/data/selector implementations do not block every local test. In particular, Task 2 can consume an explicitly supplied state-54 predicate result, and Task 4 can consume a supplied clip result. Neither proves those producers or physical motion. Production transaction integration, durable seeds/results/paths and actual callers remain required for Task 6 and gameplay acceptance. A completed, tested local policy API is an allowed partial deliverable; empty helpers, a fake production pipeline or unimplemented future entrypoints are not. No evidence gate is relaxed.

**Scope refinement (2026-11-03):** The earlier blanket retarget hold was too broad for local validation. The existing `startLuaSuiteProfile` module/ECS branches bypass full authority startup. This correction changes planning readiness, not the approved spec, product support or evidence labels.

## Depends on

These are consumed contracts, not permission to implement other specs. Distinguish a supplied local input from its real gameplay producer.

| Dependency | Current delivery / source finding | Local acceptance | Production integration still needed |
|---|---|---|---|
| Retarget boundary | Approved headless scope, existing module/ECS runner | Available; no new target/profile | No playable 1.13c admission claimed |
| Numeric foundation | #676: `policy/numeric.lua`, `engine.integer/v1`, actual composition and identity tests; all CI green | Available on PR head for Tasks 1/2/4/5 | Delivered on main; later add caller-owned durable unit/missile seed integration |
| RLI stat getter | `policy/stat_resolution.lua:resolve` is a flat/percentage calculation, not RLI-R02a/R02b | Task 1's supplied read operation/current result can be tested at its callback boundary; do not claim the getter itself passed | Real selected-array/minimum getter before actual item callback binding |
| RLI state query/routing | Inspected timed-state owner tracks named source-tagged instances; no shared RLI bitset getter at this boundary | Task 2 may consume the supplied RLI-shaped state-54 result; Task 3 may log a supplied SetState request | Actual RLI-R03 lookup/routing; live state representation/binding and guarded membership effects |
| RLI mode setter | No matching equal-mode/setMode owner in the inspected policy modules | Task 5 request-only assertions can supply planner/preparation; full V13c retains its real equal-mode-branch dependency | RLI-R04a for V13c and mode integration; update/animation effects keep their gates |
| GDF/GDA monster/data facts | Raw missile byte already decoded/published; monster profile collapses size, exposes mKB, and keeps BaseId as a name | Supplied byte, independent flags and valid record/base facts are legal fixture inputs | Lossless flags, mGH, numeric base binding and valid source MonStats/inTown before actual monster reactions |
| HRC selector | Current damage resolver returns immediate-health/lethal facts; no shared legacy result-word selector in inspected owners | Full Task 3 as written waits for HRC-R04a; a future explicitly split gate/rewrite-only subset could precede it | One shared selector/GH-veto owner, not a second KNB implementation |
| SBA/MHD/DPO transaction | Projectile `apply_hit` calls current damage policy and emits events, without the complete specified legacy stage chain | Not a dependency of standalone Task 2 or supplied-clip Task 4 | Task 6: chance/block/avoidance/multiplier/application/reaction order and real result/seed stores |
| STL lifecycle | Existing timed-state system is not blanket evidence of original stat-list lifecycle | A local SetState request does not require full lifecycle implementation | Membership, disguise/update and dependency effects remain governed by their owners/gates |
| IEP item binding | Compiled registration and numeric event binding remain KNB-Q01/Q02 | No automatic registration needed for supplied item callback tests | Real IEP integration/timing stays gated |

Supporting citations and reverse consumers do not add delivery prerequisites. In particular, full movement/free-list/animation implementations are not prerequisites to testing a supplied clip/planner/mode operation; they are prerequisites only to claims about their actual effects. Pending runtime observations do not gate verified local arithmetic.

## Approved bounded delivery — Task 2 missile chance

**Approved execution:** Deliver Task 2 alone as bounded static acceptance. #676 is merged; branch `codex/378-missile-knockback-chance` starts at merged main `de6f9088`. There is no missing broad foundation to build first for this supplied-input slice.

The scope is KNB-R02a, the supplied-predicate guard of KNB-R02b, and their KNB-R10a local draw assertions. It covers seven vector IDs and fourteen cases: V04; V05; V06a/b; V07a/b; V12a/d only; V17's four chances; V18a/b. No item stat/size lookup, reaction selector, mode setter, clipping, table-loader change or new persistent store is needed.

The supplied state result is the already-computed answer for state 54, not an implementation of RLI HasState. A value of 0 means absent; a nonzero RLI mask means present. Use explicit numeric zero comparison: Lua treats 0 as truthy. Tests can supply 0 or the state-54 active-mask fixture, without claiming mask storage/getter production. Actual target-state lookup and its eligibility checks are a later integration boundary.

Keep these task IDs for coverage; numbering does not prescribe dependency order:

| Task | Readiness after numeric delivery | Remaining boundary |
|---|---|---|
| 2: missile chance | Implemented supplied-state API; fourteen exact cases use real numeric.rand | Production state query, durable missile seed and contact pipeline remain deferred |
| 4: endpoint generation | Independent second candidate; real integer helpers + supplied source/clip/path fixture | Actual collision and motion remain KNB-Q07 |
| 1: item callback | Also eligible locally with supplied stat read and size facts; more guards/cases than Task 2 | Actual stat/data/IEP binding remains separate |
| 3: reaction stage | Full task waits for shared HRC selector; do not copy selector policy into KNB | Production state/data/terminal effects separately gated |
| 5: monster start | Request-only subset is bounded; full task needs Task 3/4 composition for V09 and RLI equal-mode behavior for V13c | Real preparation/mode/motion effects remain Q06/Q07 |
| 6: transaction/recovery | Blocked on real shared transaction and durable owners | No test-only replacement pipeline |

This work references #378 as partial delivery; no child issue is required for the approved bounded slice. Do not close #378 for these fourteen local cases or report V12b/c as passed.

**Task 2 delivery (2026-11-05):** `policy/knockback.lua:missile` implements only the supplied-state chance API. All fourteen selected cases passed through the module/fast harness with the real numeric module; two additional successful-draw OR preservation checks and all eleven numeric cases also passed. The tests first failed with the policy module absent. No other task is implemented by this delivery. The callable local policy is not connected to projectile contact and makes no actual state-query, durable-state, physical-motion or runtime-observation claim. Broader verification and review results belong to the implementation handoff; the numeric PR's CI is not Task 2 CI.
## Global Constraints

- Evidence build: Lord of Destruction Expansion 1.13c.
- Result is u16; item stat is signed i32; missile chance and distance byte B are u8.
- Seeds are separate attacker, missile and defender pairs `{lo,hi}` of u32 words. NRF-R01 steps `u64(lo) * 0x6AC690C5 + u64(hi)`; both words are written. Item chance uses newLow & 127; missile chance uses unsigned remainder out of 100.
- Signed division truncates toward zero. Start/endpoint coordinates are u16; source kinds 2/4/5 retain full i32 coordinates. KNB-R08 subtraction, absolute value and multiplication wrap i32; endpoint addition wraps to low 16 bits.
- NRF-R10 supplies five path units per map tile. B=5/10 gives S=(B >> 1)+1=3/6 path units, not measured completed travel.
- No local selection/start/endpoint RNG instruction implies a zero total for supplied callees. Never advance defender or owner seed for a missile chance roll.
- All future-tick state belongs to existing explicit ECS/registered stores. No mutable module-global seed, result, mode or path state. Checkpoint/restore and replay parity must still pass.
- Generic forced-motion semantics and existing supported gameplay remain unchanged outside this spec. No 1.14d behavior is silently replaced with 1.13c assumptions.
- Never commit owned data, captures, private saves or investigation artifacts. Plan authoring alone does not authorize commit/push/GitHub updates. The maintainer separately authorized merging #676 and implementing Task 2 on a fresh branch; no Task 2 publication or merge is requested.

## Review Focus

1. Preserve every unrelated result bit and every non-owning seed; Task 1/2 vectors pin these.
2. Town rejection must precede lethal state-54 handling; Task 3 includes KNB-V27k.
3. Fallback before base conversion can re-add KB even when mKB is false; Task 3 includes KNB-V32.
4. Failed clipping retains the previous count and requested first waypoint, and never changes actual position; Task 4 includes KNB-V13b/V30.
5. A zero planner/setter return does not become start failure; Task 5 includes KNB-V13a/c and V33.

## File ownership and interfaces

Paths below are relative to the repository root. New policy files match the existing policy ownership rule in `docs/architecture/gameplay-ownership.tsv`; verify the architecture gate instead of adding a redundant rule.

| File | Responsibility |
|---|---|
| Create `internal/content/d2legacy/lua/d2legacy/policy/knockback.lua` | Guarded chance callbacks, reaction gating/fallback/conversion, monster start requests |
| Create `internal/content/d2legacy/lua/d2legacy/policy/knockback_path.lua` | Distance decoding, exact endpoint arithmetic, supplied clipping and local path writes |
| Create adjacent `knockback_test.lua`, `knockback_path_test.lua` | Named KNB vectors, exact seeds, ordered calls and untouched-field assertions |
| Extend existing shared-owner suites after prerequisites land | Real numeric/stat/state/selector integration and missile/queued ordering; no duplicate test-only dispatcher |
| Existing `data/monster.lua`, `data/missile_skills.lua`, `components/shared.lua`, `components/skill_actions.lua` | Current binding facts are audit entry points. Any required field expansion belongs to producer integration delivery; supplied-input local tests do not require it |
| Existing `systems/projectile.lua` and shared combat owners replacing its incomplete transaction | Future supported production callers of local policy; adapt their landed stage contracts before execution |
| Existing `gameplay/systems/forced_motion.lua` | Preserve unchanged: this is not the specified path-type-8 planner |

Lua signature notation below describes values, not new runtime registries. `context` is an explicit per-invocation record of supplied logical references, result word, seeds and facts from spec section 3. `ops` contains narrow operations supplied by the owning production caller. Records are transient arguments over caller-owned durable state, not a second authoritative copy.

- `knockback.item(context, ops) -> integer 0|1`: guarded owner/target/damage references; `ops.read_stat(owner, stat_id) -> i32` and `ops.rand_mask(seed, 128) -> u32`. Mutates only the supplied damage result and attacker seed.
- `knockback.missile(context, ops) -> nil`: at the supplied-predicate boundary, context contains `chance` (u8), `state54_result` (u32, 0 absent/nonzero present), `missile_seed` (`{lo,hi}`) and `damage.result` (u16). `ops.rand(seed,100) -> u32` delegates to the real `numeric.rand`. Mutates only that missile pair and damage result. Real target lookup/HasState invocation is not performed or accepted by this local API; no new persistent state representation is introduced.
- `knockback.react(context, ops) -> nil`: optional room/source/target, valid records, state predicates, mode/result facts. `ops.set_state(target,92,1)`, `ops.select_reaction(context) -> branch`, and `ops.request_reaction(branch,context)` are shared state/HRC operations. The latter owns GH veto when GH wins; KNB does not implement HRC-R06.
- `knockback.start_monster(context, ops) -> integer 0|1`: availability, optional unit, mode/base/path; ordered `ops.set_type(path,8)`, `ops.set_distance(path,B)`, `ops.plan(target,0)`, `ops.prepare_mode_entry(target)`, `ops.set_mode(target,13)`.
- `knockback_path.generate(path, source, explicit_destination, ops) -> integer 0|1`: path start/count/waypoint/planned/reference fields; optional source with kind and supplied position/path presence; `ops.clip(start,requested) -> point`. Caller owns path storage. No actual-position write.
- `knockback_path.step(byte) -> integer`: unsigned decoding only.

The signatures are local module interfaces decided by this plan. Shared prerequisites need not adopt those API names: use narrow caller adapters to their landed APIs; do not require another owner to rename its interface. Adapter paths/signatures must be recorded in this section before execution. No generic service-locator or full host/runtime object is passed.

## Coverage

S = **verified in 1.13c code (static)**; U = **unresolved**. Tests use the literal vector IDs as case names inside the adjacent Lua suites. The spec's exact case variants, seeds, draws, logs and final fields are the fixture authority, not the historical gap table.

| Rule/sub-id | Label | Task / implementation test | Assertions / limits |
|---|---|---|---|
| KNB-R01 | S | 1; knockback_test KNB-V01/V02/V03/V14/V15/V16/V31/V35 | Guards, supplied stat identity, large precedence, threshold, OR, seed ownership |
| KNB-R02a | S | 2; KNB-V04/V05/V06/V07/V17/V18 | Strict unsigned comparison, zero guard, unchanged flags, exact words |
| KNB-R02b | S | 2; KNB-V12a/d | Supplied state-54 result skips draw/OR without clearing existing KB; actual RLI query/eligibility is outside this local acceptance |
| KNB-R03b | S | 3; KNB-V27a–k | Town membership and valid source record; town-before-state |
| KNB-R05a | S | 3; KNB-V12b/c, V27k | Request SetState(92,1) only on reached lethal state-54 branch; actual disguise/update gated Q11 |
| KNB-R03a | S | 3; KNB-V26 | Null/type and monster mode-0/12 refusal; no subtype-equivalence claim |
| KNB-R06a | S | 3; KNB-V10/V19/V32 | Clear only KB, OR GH, retain remaining flags |
| KNB-R06c | S | 3; KNB-V09/V20/V32 | After fallback, valid base 78 and no freeze; supplied base integration |
| KNB-R04 | S | 3; KNB-V11 | Consumed shared selector: monster death precedes KB; player block precedes death; no death effects |
| KNB-R06b | S | 3; KNB-V19/V28 | Shared monster/player priorities, one selected call; downstream player effects gated |
| KNB-R06d | S | 5; KNB-V13a/c/V21/V33 | Start guards/ordered requests/ignored returns; dependent effects gated Q06/Q07 |
| KNB-R07a | S | 4/5; path_test KNB-V08/V22, knockback_test V09 | B=5/10, S=3/6; no completed-distance claim |
| KNB-R08 | S | 4; path_test KNB-V08/V23/V29/V30b/V34/V36 | Supplied widths/wrap/truncation/reference and explicit destination |
| KNB-R09a | S | 4; path_test KNB-V13b/V23/V30/V34/V36 | Exact local writes/count retention/reference clearing; supplied clip only |
| KNB-R09b | U | gated | Q07: no actual collision or motion implementation |
| KNB-R10a | S | 1/2/5; V14/V16/V17/V21 | Direct local draws only; full callee RNG total gated Q06 |
| KNB-R10b | S | 6; KNB-V24 | Actual shared stage invokes chance before supplied later stages; dependent effects unproved |
| KNB-R10c | U | gated | Q01/Q02: no compiled registration or item timing |
| KNB-R10d | S | 6; KNB-V25 | Apply before reached GH-veto, chance first, KB skips GH-veto; item timing remains gated |

The section-3 named text mappings for stat 81, selectors 7 and state 92 are supporting bindings with recorded 1.13c data evidence, not new gameplay rules. Use the named ItemStatCost/States evidence records through GDF/GDA prerequisite binding tests. KNB-V35b must still accept supplied synthetic stat identity 82; do not hardcode 81 into the callback. A text-cell check never establishes compiled registration.

## Test fixture rules

Use spec section 8.1 defaults and all variants exactly. Tests must assert returns, full result words, every participating seed word, ordered call arguments, requested/planned/first-waypoint/count/reference and unchanged actual position. All unlisted seeds/fields stay unchanged. Operation spies perform only their stated supplied effects.

Use real `d2legacy.policy.numeric.rand/rand_mask` for chance tests, not a mocked integer stream or a knockback-local LCG. Task 2 supplies the state-54 result as an input; assert numeric 0 does not veto and the nonzero mask does. This tests KNB's supplied-result guard, not RLI state storage/getter behavior. Use supplied log-only clip/planner/prepare/reaction operations where the vector says so. V13c uses the shared RLI equal-mode branch with a supplied update observer; it must not reset animation or process lists. V25c's supplied veto makes exactly the named defender mask-4 draw and asserts no HRC threshold calculation.

Focused command for Task 1/2/3/5:
`go test -count=1 ./internal/mod/d2legacy -run '^TestLuaSuites$/^d2legacy/policy/knockback$'`

Focused command for Task 4:
`go test -count=1 ./internal/mod/d2legacy -run '^TestLuaSuites$/^d2legacy/policy/knockback_path$'`

These suites use the existing module/fast profile and harness discovery; no new runner or unregistered authoritative test command.

## Task 1: Item chance at supplied callback entry

**Files:** extend policy/knockback.lua and policy/knockback_test.lua after Task 2; create only if absent, preserving any previously delivered entrypoint.
**Consumes:** real shared NRF-R01/R02 and supplied current-stat read via narrow ops. Actual RLI-R02a/R02b implementation is a separate production integration dependency, not established by these supplied callback vectors.
**Produces:** item(context,ops), signature above.

- [ ] Write failing cases KNB-V01/V02/V03/V14/V15/V16/V31/V35. In particular, seed {1,0} yields low 0x6AC690C5 and mask result 69: small passes, normal/large fail; both flags use large. Stat 0/-1 and missing references do not read/draw beyond their stated guards. V35 retains result 16 as 24 and reads supplied identity 82 in its independent case.
- [ ] Run the focused knockback suite; confirm failure reaches the missing/new item behavior rather than a broken dependency.
- [ ] Implement item with supplied large/small flags kept separate, default threshold 64, positive-value guard, one attacker mask draw and u16 OR. Consume the supplied current-stat read; do not implement a second RLI getter or substitute stat_resolution.resolve's percentage formula.
- [ ] Re-run the focused suite and shared numeric tests; all named variants must pass. Any actual getter tests belong to that owner's later integration and must not be reported as passed by a stat-read spy.
- [ ] Review only this task's diff; record partial callback acceptance. Do not register selector 7 or change melee event timing.

## Task 2: Missile chance and uninterruptible guard

**Files:** create `internal/content/d2legacy/lua/d2legacy/policy/knockback.lua` and adjacent `knockback_test.lua` if delivered first. Add only the missile entrypoint; later tasks extend these files.
**Consumes:** supplied `state54_result`, chance byte and result word, plus real `numeric.rand` through narrow ops. No actual state getter, stat reader, monster profile or HRC selector is required for this bounded slice.
**Produces:** missile(context,ops).

- [x] Add failing KNB-V04/V05/V06/V07/V12a/d/V17/V18. Seed {2,0} produces remainder 70: chance 33 fails, 75 passes. Seed {1,0} yields 85: 85 fails, 86/100/255 pass. Seed {5,0} yields low 0x15E0D3D9/high 2, remainder 33: 33 fails, 34 passes. V06b must advance {1,15} to {0x6AC690D4,0}.
- [x] Run the focused suite and confirm failures on these new cases.
- [x] Implement `state54_result ~= 0` then zero-byte guards, one missile Rand(100), strict unsigned comparison and OR. Preserve existing result 12 in V07b/V12d and both non-owning seeds. Use `engine.integer/v1.band_u32` to test bit 8; OR can be expressed by adding 8 only when that bit is absent, so no new integer capability is needed. The inputs remain u16/u8/u32 as declared.
- [x] Re-run the focused suite and shared numeric tests, and review exact seed/call logs. Cover all fourteen cases under the seven named vector IDs; V12b/c are excluded. No actual state-query, producer data binding, durable seed or contact-pipeline acceptance is implied.
- [x] Report partial #378 static acceptance only. No production chance insertion before the shared missile transaction and state/seed ownership integration are delivered; keep current combat and generic forced motion unchanged.

## Task 3: Reaction gates, fallback and conversion through shared priority

**Files:** modify knockback.lua / knockback_test.lua.
**Consumes:** supplied valid records/state results plus the real shared HRC selector.
**Produces:** react(context,ops); no duplicate HRC selector or GH-veto algorithm.

- [ ] Add failing KNB-V10/V11/V12b/c/V19/V20/V26/V27/V28/V32 and V09's selection assertion. Use the shared selector with log-only terminal operations.
- [ ] Assert V27k leaves result 10/mode 1/path type 2/B 5/count 0/planned (7,7)/actual (100,100) and all seeds unchanged, with no SetState call. Assert V11d selects block rather than death; V28 selects block for player result 24 but KB(13) for monster result 24.
- [ ] Run the focused suite and confirm each new boundary fails before implementation.
- [ ] Implement town gate before state 54 before kind/dead-mode routing. Apply u16 fallback before valid base-78/unfrozen conversion, then delegate priority and selected effects once.
- [ ] Verify V32a result 20/block; V32b result 12/KB(13); V32c invalid base result 4/GH. V12c logs only SetState(target,92,1), not death.
- [ ] Re-run focused knockback and shared HRC/state tests. Actual disguise, queue, ordinary death, avoidance/block and player-mode effects retain their gates.

## Task 4: Integer endpoint and supplied clip writes

**Files:** create knockback_path.lua / knockback_path_test.lua.
**Consumes:** shared wrapping/truncating numeric operations, caller-owned path, supplied clip.
**Produces:** step(byte), generate(path,source,explicit_destination,ops).

- [ ] Write failing KNB-V08/V13b/V22/V23/V29/V30/V34/V36 with literal path fixtures. V08 requests (103,100); V23 includes (103,100)/(100,103)/(44,44)/(100,100); V29 static x=65536 requests (97,100); V36 missing dynamic path uses source (0,0) and requests (103,103).
- [ ] Include V34 u16 endpoint wrap to (2,100), i32 product wrap yielding (100,100), and minimum-i32 absolute-value behavior yielding (0,0). Use exact integer operations, never floating-point ray normalization.
- [ ] Run the focused path suite and confirm the new endpoint/write failures.
- [ ] Implement unsigned step decoding and wrapped arithmetic using shared numeric helpers. Store requested waypoint before clip; store supplied planned point and clear reference; retain count/requested waypoint on zero-length clip; otherwise first waypoint becomes clipped and count becomes 1.
- [ ] Re-run focused path tests. V30a keeps count 2/requested waypoint (103,100) despite planned (100,100); V30b sends explicit (120,120) to clip and stores returned (101,102). Every case keeps actual position unchanged.
- [ ] Review for accidental collision/movement implementation. Leave forced_motion.admit and its existing tests unchanged.

## Task 5: Monster start requests and composed base exception

**Files:** modify knockback.lua / knockback_test.lua; consume knockback_path.lua from Task 4.
**Consumes:** supplied mode availability/valid base, shared setter; caller-supplied planner and entry operation.
**Produces:** start_monster(context,ops).

- [ ] Add failing KNB-V09/V13a/c/V21/V33. V21 either missing availability returns 0 without writes; V33 mode 0 returns 1 without writes. Also cover the rule's missing-unit branch using the same supplied availability and unchanged-field checks, without claiming an additional original observation.
- [ ] Run focused suites and confirm these failures.
- [ ] Implement both availability gates, missing-unit/mode-0 success, then type 8/B=5 or 10, plan(target,0), prepare, setMode(13), return 1. Ignore planner/setter returns; do not invent cleanup.
- [ ] Verify V09 composes selection/start/generator to result 12, mode 13 and requested (106,100). Verify V13a continues after plan return 0. V13c logs update while modeChanged is false, then sets it true and returns handler 1 despite setter 0, with no list/reset calls.
- [ ] Re-run focused policy/path and shared mode tests. Assert no direct draw without pretending supplied mode/entry operations have universally zero draws.

## Task 6: Shared stage-order acceptance and determinism

**Files:** extend the landed shared missile/queued combat owner's test suites; adapt supported callers to knockback.lua in that same owner. Current audit entry point is systems/projectile.lua:apply_hit. No second dispatcher file is authorized.
**Consumes:** Tasks 1–5 plus SBA/MHD/DPO/HRC prerequisite transaction and durable seed/result/path storage.
**Produces:** real shared stage invocation with local KNB policies, exact KNB-V24/V25 acceptance and bounded replay/restore evidence.

**Readiness condition:** before execution, record the actual landed caller/test paths and adapter signatures in the file/interface section. This is blocked by absent prerequisite implementations, not an invitation to guess a transaction interface or ship unused modules.

- [ ] Add failing KNB-V24 to the actual shared transaction: chance 86/seed {1,0} gives result 8 and missile {0x6AC690C5,0} before ordered [block,avoidance,multiplier,apply,reaction]; each supplied observer sees those facts and untouched owner {2,0}/defender {9,0}.
- [ ] Add failing KNB-V25a/b/c at the same boundary: a logs [apply,KB(13)] without GH veto; b logs [apply,GH-veto,GH]; c logs [apply,GH-veto,defenderDraw(mask4,1),GH] with the exact final missile/defender words. These are supplied-stage assertions, not full HRC-veto tests.
- [ ] Run the owning focused tests; confirm failures before connecting the local policies. Insert calls only at the already verified shared transaction boundaries. Automatic item registration and physical forced-motion emission remain disabled.
- [ ] Re-run these tests with the complete KNB policy/path suites. No test-only dispatcher may substitute for the production shared call path.
- [ ] In the owning session acceptance suite, checkpoint caller-owned seed/result/path facts between two explicit supplied local-stage invocations; compare restored versus uninterrupted final canonical authority and ordered outputs. Replay the same invocation inputs and compare the same facts. Do not add an externally admitted test-only command or assert unverified actual movement.
- [ ] Run relevant shared-session tests and full repository gates listed below. Fresh review must confirm there is one policy caller authority, all modules are reached, no gated emission was enabled, and unchanged generic movement behavior still passes.

## Gated

No implementation task is assigned to these assertions.

| Boundary | Gate | Evidence needed to unlock |
|---|---|---|
| Automatic item stat/selector binding and callback timing | R10c; Q01/Q02 | Matching compiled record/loader correspondence, event-name numeric conversion, registration and actual timing relative to block/CB/deadly strike |
| Actual clip endpoint selection and physical motion, including blocked outcome | R09b; Q07 | Complete checked collision/callees, count initialization, player mode-19/planner/tick contracts and attributed observations |
| Full monster mode-dispatch refusal/preparation effects and complete RNG count | Q06; dependent portion of R06d/R10a | Outer return handling and owned entry/mode/elemental/animation contracts, including HRC-R15 |
| Player displacement/avoidance/block terminal effects | Q07; HRC-R04b/R11 | Owning player-mode/reaction/movement contracts, not merely the local selection |
| Full hireling/summon/NPC equivalence | Q08 | Actual subtype producers/records/mode paths; local kind routing alone is insufficient |
| Movement conflicts, networking and client correction | Q10 / #379 | Owning movement/network contracts and evidence; do not change current replacement policy here |
| Disguise and room-update effects of state 92 | Q11 | STL-R07a/STL-Q07 and RLI-Q02 completion; local SetState request remains eligible |
| Completed-distance measurement and runtime-confirmation claims | Q04/Q12; O01–O06 | Matching owned 1.13c captures naming actual inputs/assertions; five units/tile is not a completed-distance measurement |
| 1.14d parity and #378 closure | Issue acceptance | Separate target-patch observations; passing a 1.13c static vector does not establish 1.14d behavior |
| HRC speed 4096 as an emitted generic forced-motion speed | Outside local KNB writes; Q06/Q07 | HRC-owned rate contract plus complete movement/units/integration evidence; do not copy the number into current requests |

## Existing tests and compatibility

- `data/monster_catalog_test.lua`, `commands/spawn_monster_scaling_test.lua` and `adapter/movement/catalog_real_test.go:TestOwnedExpansion114dMovementAndKnockbackRecords` cover the current profile/data representation. The current decoder rejects both small/large flags; KNB-V03 requires large precedence at a supplied-record callback. Resolve lossless source facts through the prerequisite data owner, not by deleting the vector or broadly weakening decoder validation in #378.
- `data/missile_skills_test.lua` and `missile_skill_real_test.go` preserve raw KnockBack values; extend eligible bindings through their owner without relabeling 1.14d tests as 1.13c proof.
- `gameplay/systems/movement_test.lua`, `systems/player_death_test.lua` and `systems/point_movement_skill_test.lua` assert existing generic forced-motion/replacement/interruption outcomes. They do not contradict KNB-R08 because they test a different, still-gated physical integration. Preserve them.
- Existing named RNG streams and archived checkpoints must not be globally reseeded/reinterpreted. Shared prerequisite delivery owns any versioned seed/path facts and compatibility policy; this plan consumes it and tests its own continuation only.

## Verification and review

During implementation, each task follows red/green checks above. After all tasks and prerequisites, run on the submitted revision:

1. Both focused Lua policy suites and the actual shared combat/session suites, with explicit vector IDs.
2. `make test-lua-hardening` (fast/integration suites repeated 3 times, order seed 8675309), `make test`, `make architecture`, `make test-race`.
3. Relevant owned-data decoding checks only when the matching data is available; list exact patch/source and skips. These do not replace original-runtime observations.
4. Fresh independent review of production reachability, partial acceptance, integer widths, seed ownership, replay/restore and gates. Fill every section of the PR template; use Refs #378, never Closes #378 for this partial delivery.

Plan-authoring verification checks document links, rule/vector coverage and whitespace only. No gameplay implementation, unit tests, owned-data runs or runtime experiments were performed during authoring. Do not describe future commands as passed.

## Handoff

Task 2 is implemented at its approved supplied-state boundary, with fourteen selected vector cases and two OR preservation regression cases. #676 is merged and the bounded slice uses a fresh branch from merged main. Tasks 1/3/4/5/6 remain unimplemented here. Numeric primitives are delivered and CI-verified on that PR; no new broad foundation is required for the supplied-state missile chance stage. Record/stat/selector dependencies remain specific to their consuming assertions, while actual producer/transaction/durable-state integration remains separate. These delivery holds are not new evidence gaps or authorization to expand #378. Preserve full-rule coverage and all original evidence gates, and report only the assertions a partial delivery actually reaches.
