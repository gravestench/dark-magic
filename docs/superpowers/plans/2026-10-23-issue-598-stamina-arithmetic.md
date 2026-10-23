# Issue #598 Running Stamina Drain Arithmetic Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Match the original 1.13c signed 32-bit running stamina drain arithmetic in the existing shared authority and prediction rule.

**Architecture:** Correct the private `staminaDrain` calculation used by `movement.AdvanceStamina`; keep all public interfaces and tick ordering. Test the pure rule, the authoritative Lua player tick, and connected client prediction so all three observe the same corrected result.

**Tech Stack:** Go movement adapter and client prediction; embedded Lua authority test harness.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-598-stamina-arithmetic-design.md`

## Global Constraints

- Work only on issue #598's drain amount for valid prepared inputs. Do not change recovery, town membership, stat admission, movement cadence, or public interfaces.
- `RunDrain` is unsigned 8-bit in source data; torso speed and stamina drain percentage are signed 32-bit. Apply signed 32-bit wrapping at each specified multiply and add, then truncate division toward zero and floor final drain at one.
- 1.13c evidence is static code; original stored stamina and 1.14d runtime parity remain unverified.
- Never commit Diablo II game data, captures, credentials, or private saves.

## Review Focus

- Percentage product overflow: a valid large positive `item_staminadrainpct` must use wrapped signed 32-bit multiplication before dividing.
- Torso multiplier overflow: a synthetic signed 32-bit speed must wrap the weight product before the one-unit floor.
- Signed percentage division: positive and negative percentages must truncate toward zero at the original division stage.
- Signed torso division: positive 10 and negative 9 must use the original multiplier behavior.
- Minimum drain: `p=100` and `RunDrain=0` must each drain one raw unit while running outside town.

---

### Task 1: Pin the pure rule and correct its arithmetic

**Files:**
- Modify: `internal/mod/d2legacy/adapter/movement/rules_test.go`
- Modify: `internal/mod/d2legacy/adapter/movement/rules.go`

**Interfaces:**
- Consumes: `AdvanceStamina(tick StaminaTick) ResolvedStamina`, with existing `StaminaTick` fields.
- Produces: The same interface and existing private `staminaDrain(tick StaminaTick) int64`, now using original-width stages.

- [ ] **Step 1: Write the failing table test** `TestAdvanceStaminaMatchesSigned32DrainArithmetic` in `rules_test.go`. Each case has `CurrentRaw=MaximumRaw=23552`, `Running=true`, `Moving=true`, `InTown=false`, and checks both `CurrentRaw` and `ForceWalk=false`. Cases: `{RunDrain:20,p:107374183,want:23512}`, `{RunDrain:20,ArmorRunDrain:1000000000,want:23551}`, `{RunDrain:20,p:33,want:23525}`, `{RunDrain:20,p:-33,want:23499}`, `{RunDrain:20,ArmorRunDrain:10,want:23472}`, `{RunDrain:20,ArmorRunDrain:-9,want:23512}`, `{RunDrain:20,p:100,want:23551}`, `{RunDrain:0,want:23551}`. Here `p` maps to `StaminaDrainPercent`.
- [ ] **Step 2: Confirm red.** Run `go test ./internal/mod/d2legacy/adapter/movement -run TestAdvanceStaminaMatchesSigned32DrainArithmetic -count=1`; expect at least the large percentage and large torso cases to fail under the current `int64` products.
- [ ] **Step 3: Correct `staminaDrain` only.** Compute `w=int32(2*RunDrain)`, `m=int32(ArmorRunDrain/10+1)`, then `w=int32(w*m)`; for nonzero percentage compute `product=int32(int32(StaminaDrainPercent)*w)` and `w=int32(w+product/-100)`; return `max(1,int64(w))`. Go's signed integer division truncates toward zero. Keep the conversion stages explicit and do not alter `AdvanceStamina`.
- [ ] **Step 4: Confirm green.** Run `go test ./internal/mod/d2legacy/adapter/movement -run 'TestAdvanceStamina' -count=1`; expect pass.
- [ ] **Step 5: Commit** the two files with `git commit -m "fix: match 1.13c running stamina drain arithmetic"`.

### Task 2: Verify authority and prediction use the corrected rule

**Files:**
- Modify: `internal/content/d2legacy/lua/d2legacy/systems/player_stamina_test.lua`
- Modify: `internal/app/clientapp/client_prediction_test.go`

**Interfaces:**
- Consumes: unchanged `movement.AdvanceStamina` from Task 1, `fixtures.player_entry`, and `newPositionPrediction`/`advanceTick`.
- Produces: regression coverage only; no production interface changes.

- [ ] **Step 1: Add authoritative Lua case** `large_stamina_drain_percentage_uses_signed32_product`: enter a player at `level_id=2` with passive source `{id="wrap",stat="item_staminadrainpct",value=107374183}`; step entry, reset `stamina_raw` to `2560`, submit a running `player.move`, then step one tick. Assert resolved movement stat `item_staminadrainpct=107374183` and `stamina_raw=2520`. Use the existing ECS and fixture style in `player_stamina_test.lua`.
- [ ] **Step 2: Add connected prediction test** `TestPredictionStaminaUsesSigned32Drain`: seed an Amazon HUD at level 2 with `StaminaRaw=2560`, `MaxStaminaRaw>=2560`, `RunDrain=20`, percentage `107374183`, and a pending running move; call `newPositionPrediction` and `advanceTick` for the command tick, then assert `prediction.staminaRaw==2520`. Use the existing `client_prediction_test.go` HUD and intent pattern.
- [ ] **Step 3: Confirm both tests use the desired path.** Run `DARK_MAGIC_LUA_TEST_TIERS=fast go test ./internal/mod/d2legacy -run TestLuaSuites -count=1` and `go test ./internal/app/clientapp -run TestPredictionStaminaUsesSigned32Drain -count=1`; expect pass. To demonstrate red without reverting production, the Task 1 red result is the arithmetic proof; verify that these tests invoke the real shared rule rather than a stub.
- [ ] **Step 4: Commit** the two test files with `git commit -m "test: cover signed32 stamina drain in authority and prediction"`.

### Task 3: Verify, review, and prepare the PR

**Files:**
- No additional source files.

**Interfaces:**
- Consumes: Task 1 and 2 commits.
- Produces: reviewable PR for issue #598 with exact submitted-commit evidence.

- [ ] **Step 1: Run** `go test ./internal/mod/d2legacy/adapter/movement ./internal/app/clientapp -count=1`, `make test-lua`, `go test ./...`, `make architecture`, `make test-lua-format`, and `git diff --check origin/main...HEAD`; record exact outputs on the final commit.
- [ ] **Step 2: Request independent subagent reviews** of arithmetic/evidence and scope/architecture. Address findings, rerun affected checks, and repeat review until approved.
- [ ] **Step 3: Fill every PR template section**, including `Closes #598`, source and test links, exact checks and results, explicit local game-data/race status, inference limits, and AI use. Have a subagent review the body.
- [ ] **Step 4: Push and create the PR**; wait for all CI jobs, then merge only after green checks and a stable head. If CI fails, diagnose, fix, rerun checks, and repeat.
