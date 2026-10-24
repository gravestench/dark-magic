# Connected Cold Plains Acceptance Stability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the owned-data connected Cold Plains population acceptance test observe one installed client frame and complete live crossings before expensive offline replay verification.

**Architecture:** The test's crossing helper will select the HUD retained by `clientWorld` from the just-completed reconciliation frame, matching the map and controlled-player projection installed by that frame. The population test will keep its captured replay and all assertions, but move offline replay verification to the final step so a live listen host cannot outrun the client's 64-tick event history before another client frame.

**Tech Stack:** Go tests, QUIC listen session, `internal/app/clientapp`, owned 1.14d MPQs.

**Spec:** `docs/superpowers/specs/2026-10-23-issue-630-connected-cold-plains-test-design.md`

## Global Constraints

- Linked issue: #630. Change only connected Cold Plains acceptance-test observation/order and its issue documentation.
- Preserve all existing assertions, production event-gap validation, map activation, transport behavior, and game rules.
- Do not commit game data, captures, credentials, or private saves.
- Final correctness and formatting/style reviews use independent subagents; the style reviewer must use `refactor-for-readability`.

## Review Focus

- A newer QUIC HUD published after a frame must not be compared with that frame's older map; Task 1 helper uses `lastHUD`.
- A completed frame's installed HUD must still match its controlled-player projection and authoritative arrival; Task 1 retains `assertConnectedColdPlainsArrival`.
- Replay verification must use the replay captured at the acknowledged population checkpoint, even when run last; Task 1 retains the same `replay` value.
- Both return crossings and stable resident IDs must remain checked before replay work; Task 1 preserves and orders those assertions.
- A genuine semantic event gap must still fail; Task 1 leaves `remote_events.go` and its gap tests untouched, and verification runs those tests.

---

### Task 1: Align the connected test's observations with client frames

**Files:**
- Modify: `internal/app/clientapp/cold_plains_transition_test.go`
- Modify: `internal/app/clientapp/cold_plains_population_real_test.go`

**Interfaces:**
- Consumes: `fixture.advanceGame(t, 1)` and `app.clientWorld.lastHUD`, which `presentLocalPrediction` writes at the end of successful frame reconciliation.
- Produces: existing `moveConnectedAcrossColdPlains` helper and `TestConnectedColdPlainsPopulationMatchesAuthority` with unchanged assertion set and test names.

- [x] **Step 1: Record red evidence.** Use the previously observed owned-data failures from clean main (`connected map = 2, want 3` at first crossing) and the candidate (`remote presentation: semantic event window gap` during return after offline replay) as the failing existing acceptance test. Rerun the isolated test with owned 1.14d data for a current baseline; it may pass because the issue is intermittent, so do not invent a new deterministic failure.
- [x] **Step 2: Make the narrow test edit.** In `moveConnectedAcrossColdPlains`, read `hud := app.clientWorld.lastHUD` after each completed `fixture.advanceGame` instead of `app.network.client.View()` for arrival selection and assertion; keep `View()` only for timeout diagnostics. In `TestConnectedColdPlainsPopulationMatchesAuthority`, move `verifyProductionPopulationReplay(t, f, replay)` from before return crossing to after the final resident-ID comparison, immediately before return. Keep the captured `replay` and every other assertion.
- [x] **Step 3: Run green owned-data checks.** With `MPQ_DIRECTORY` and `DARK_MAGIC_TEST_MPQ_DIRECTORY` set to `/Users/dylanknuth/d2_english_mpq`, run `go test -count=5 ./internal/app/clientapp -run '^TestConnectedColdPlainsPopulationMatchesAuthority$'`, then `go test -count=1 ./internal/app/clientapp`. Record exact failures, if any; investigate rather than weakening assertions.
- [x] **Step 4: Run focused safety checks and commit.** `go test -count=1 ./internal/app/clientapp -run 'TestConnectedColdPlains|TestConnectedSemantic'` without assets, `gofmt` check, and `git diff --check`; commit only the two test files with `test(clientapp): align connected Cold Plains assertions with frames`.

### Task 2: Review and integrate

**Files:**
- Add: `docs/superpowers/specs/2026-10-23-issue-630-connected-cold-plains-test-design.md`
- Add: `docs/superpowers/plans/2026-10-23-issue-630-connected-cold-plains-test.md`
- Create: PR description using `.github/PULL_REQUEST_TEMPLATE.md`.

**Interfaces:**
- Consumes: Task 1's two test-only changes and exact red/green evidence.
- Produces: reviewed PR closing #630, merged after required CI passes.

- [x] **Step 1: Run independent correctness/evidence and `refactor-for-readability` formatting/style reviews.** Resolve concrete in-scope feedback; confirm assertions and production behavior are preserved.
- [ ] **Step 2: Commit the approved spec and plan.** Stage only the two issue #630 documents and commit them before the final verification run.
- [ ] **Step 3: Run final checks on the submitted commit.** Run `go test -count=1 ./...`, `make architecture`, `make test-lua-hardening`, `make build-client-backends`, owned-data connected repetition and full clientapp order, `gofmt` check, and `git diff --check origin/main...HEAD`. Run a focused race check if the test change warrants it. Report any existing unrelated failures precisely.
- [ ] **Step 4: Complete every PR-template section, push, create and attach the PR.** Re-run affected checks if the submitted commit changes.
- [ ] **Step 5: Verify all required CI jobs pass, then merge.** Fast-forward the clean `~/src/dark-magic` main checkout to the merge commit and report remaining limits.
