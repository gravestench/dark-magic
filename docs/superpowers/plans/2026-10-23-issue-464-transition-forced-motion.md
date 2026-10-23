# Issue #464 Transition Forced Motion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** End an active forced-motion transaction once, at its pre-transition position, for both seams and warps.

**Architecture:** Keep the ordering rule in `gameplay/transition.lua`, using the existing `forced_motion.interrupt` owner. Seam systems pass their structural writer; admitted warp commands pass their tick and an immediate ECS writer through operation dispatch.

**Tech Stack:** Lua first-party `d2legacy` policy, Go-hosted Lua integration suites, existing ECS checkpoint harness.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-464-transition-forced-motion-design.md`

## Global Constraints

- Work only on issue #464's active-transaction boundary; pending requests and combat knockback production are separate issues.
- Treat original 1.13c/1.14d transition behavior as unverified.
- Do not add a new outcome owner, persistent state, game data, or Go engine behavior.
- Run checks on the submitted commit and report exact outcomes.

## Review Focus

- A player with no active transaction must transition normally: retain the existing seam and warp assertions in the focused suites.
- A malformed destination must not emit an interruption: a direct `transition.apply` test in the seam suite must retain motion and emit no event.
- A stale warp request must leave active motion intact: extend the existing out-of-range warp case with an active component and zero events.
- A transition after some movement must report the actual pre-transition position and progress: assert both fields in seam and warp cases.
- A restored game must record the same outcome: checkpoint active motion before seam crossing; restore active motion and resubmit the warp command in a paired case; check one event after a further tick.

---

### Task 1: Seam transition outcome

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/world_transition_test.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/gameplay/transition.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/world_transition.lua`

**Interfaces:**
- Consumes: `forced_motion.interrupt(context, entity, structural)` from `d2legacy.gameplay.systems.forced_motion`.
- Produces: `transition.apply(context, entity, destination, structural)`; `structural` has colon-call `create(components)` and `remove(entity, component)` methods.

- [ ] **Step 1: Write the failing seam cases.** Enter at `(6,12)` so tick 1 does not cross the `(10,12)` seam; set active `knockback` with start `(5,12)`, target `(14,12)`, speed `25`, requested distance `9`, applied distance `1`, and start tick `1`. Step tick 2 to `(7,12)` and assert the active component now records distance `2`. The completed tick 2 checkpoint now contains active motion. Call `test.expect_checkpoint_parity(1)` before the crossing; it advances both original and restored timelines through tick 3. Assert one `interrupted` event with `target_id=player:alice`, start `(5,12)`, requested distance `9`, applied distance `3`, end `(8,12)`, and tick `3`; assert destination `(6,40)` in level 2, absent active component, and unchanged event count after tick 4. Add a direct invalid-destination case that calls `transition.apply` through `pcall` while motion is active and asserts no event or removal. Retain the existing ordinary seam case.
- [ ] **Step 2: Run the focused test to prove failure.** `go test ./internal/mod/d2legacy -run '^TestLuaSuites/d2legacy/systems/world_transition$' -count=1` must fail on the new outcome/removal assertion.
- [ ] **Step 3: Implement the shared transition hook and seam wiring.** In `transition.lua`, require the existing forced-motion module and change `M.apply` to the signature above. Validate and fetch required relocation components, call `forced_motion.interrupt` once, then relocate as before. In `world_transition.lua`, receive `context, entities, structural` and pass context/writer to `transition.apply`; add the helper's declared reads/writes.
- [ ] **Step 4: Run the focused test and confirm pass.** Use the Step 2 command and require exit 0.
- [ ] **Step 5: Commit.** Commit only the three Task 1 files with `feat: interrupt forced motion at level seams`.

### Task 2: Warp transition outcome

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/world_warp_test.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/commands/interactions.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/interactions/operations.lua`
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/world_warp.lua`

**Interfaces:**
- Consumes: `transition.apply(context, entity, destination, structural)` from Task 1.
- Produces: `operations.apply(owner, entity, command)` and operation handler arguments `(owner, entity, value, command)`; existing handlers may ignore the fourth argument.

- [ ] **Step 1: Write the failing warp cases.** After the player enters at `(10,12)`, set active `knockback` with start `(8,12)`, target `(16,12)`, speed `25`, requested distance `8`, applied distance `2`, and start tick `1`. Submit the tick 2 interaction and assert one `interrupted` event with `target_id=player:alice`, start `(8,12)`, requested distance `8`, applied distance `2`, end `(10,12)`, and tick `2`; assert destination `(20,20)` in level 2, removed transaction, and no second event after tick 3. Add a paired restored-path case: prepare active motion, step tick 2 without an interaction so its checkpoint contains the moving transaction at `(11,12)` with applied distance `3`, optionally call `test.restore_checkpoint()`, then submit the same tick 3 warp command on both paths. Both cases must assert the exact one event at `(11,12)` with applied distance `3` and tick `3`, destination, removal, and no duplicate after tick 4. Extend stale out-of-range interaction to assert zero events and retained active transaction. Keep the ordinary paired-warp test path.
- [ ] **Step 2: Run the focused test to prove failure.** `go test ./internal/mod/d2legacy -run '^TestLuaSuites/d2legacy/systems/world_warp$' -count=1` must fail on outcome/removal, not fixture setup.
- [ ] **Step 3: Thread the command and supply immediate writes.** Pass `command` from `interactions.open` to `operations.apply`, then to the operation handler. In `world_warp.lua`, make the warp handler receive that fourth argument and call Task 1's transition signature with `{tick=command.tick}` and a local writer whose colon-call `create` and `remove` delegate to `ecs.create` and `ecs.remove`. Keep admission and range policy untouched.
- [ ] **Step 4: Run the focused test and confirm pass.** Use the Step 2 command and require exit 0.
- [ ] **Step 5: Commit.** Commit only the four Task 2 files with `feat: interrupt forced motion at warps`.

### Task 3: Verify and deliver issue #464

**Files:**
- Modify: PR body only, using `.github/PULL_REQUEST_TEMPLATE.md`; update issue #464 only if a completion criterion remains open.

**Interfaces:**
- Consumes: Task 1 and 2 behavior and tests.
- Produces: a reviewed PR linked to #464 with exact verification results.

- [ ] **Step 1: Run focused suites and broad gates on the final commit.** Run both focused Go tests, `make test`, `make architecture`, `make test-lua-hardening`, `make test-race`, and `git diff origin/main...HEAD --check`; record exact results and any game-data skips.
- [ ] **Step 2: Request independent code review.** Check correctness, deterministic checkpoint behavior, scope, architecture, and parity claim limits; fix and rerun affected gates as needed.
- [ ] **Step 3: Rebase on current `origin/main`, rerun required gates on the rebased commit, and prepare the full PR template.** Distinguish engine invariant from unverified original behavior.
- [ ] **Step 4: Push, open/attach PR, wait for every required CI check to pass, then merge manually.** Do not request auto-merge; verify the merged SHA and update the issue record.
