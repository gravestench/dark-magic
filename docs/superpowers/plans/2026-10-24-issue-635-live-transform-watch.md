# Issue #635 Live Transform Watch Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the self-hosted race test observe a transform datagram after movement while its hosting watch is live, and report the stalled QUIC phase on timeout.

**Architecture:** The QUIC server keeps atomic counters per active credential watch and exposes immutable snapshots; the client extends its current atomic network stats. The connected test reads `Session.reliableWorld` under `session.mu` to confirm the reliable movement correction, samples both sides, and waits for a later decoded datagram before canceling the same watch.

**Tech Stack:** Go, `testing`, `sync/atomic`, existing quic-go transport.

**Spec:** `docs/superpowers/specs/2026-10-24-issue-635-live-transform-watch-design.md`

## Global Constraints

- Linked issue #635 owns only the live-transform assertion and its transport diagnostics.
- Keep the ten-second shared test deadline, real QUIC path, packet handling, and wire bytes unchanged.
- Retain no server per-credential counter entry after its watch ends; log no credentials or payloads.
- Do not change other feature issues or commit game data.

## Review Focus

- A canceled watch must not appear active after its receiver goroutine exits: assert synchronized active/inactive state in Task 1.
- A malformed datagram must increment raw-read and decode-failure counters without incrementing decoded receipt: assert in Task 1.
- A server watch closing at the deadline must not erase the test's last phase sample: assert retained copy in Task 2.
- A datagram received before the correction must not satisfy the post-correction assertion: assert baseline increase in Task 2.
- The second client's independent reliable movement must remain checked: keep its assertion in Task 2.

---

### Task 1: Bounded QUIC phase counters

**Files:** Modify `internal/app/gameserver/sessionquic/server.go`, `watch.go`, `client.go`, `client_watch.go`; add focused tests in `internal/app/gameserver/sessionquic/transform_watch_stats_test.go`.

**Interfaces:** `Server.TransformWatchStats(credential gameserver.SessionCredential) (TransformWatchStats, bool)` returns an immutable active-watch snapshot with observed and sent ticks and observe/encode/send attempt/error counters. Extend `NetworkStats` with `DatagramsRead`, `TransformDecodeErrors`, and synchronized `TransformWatchActive`.

- [ ] **Step 1: Write RED tests.** Assert the server snapshot changes through observe/send and disappears after watch teardown; assert client raw/decode/decoded counters and active state on successful, malformed, and canceled receives. Use existing real QUIC fixture/transport helpers where possible.
- [ ] **Step 2: Run focused tests.** `go test -race -p 1 -parallel 1 -timeout 5m ./internal/app/gameserver/sessionquic -run 'TestTransformWatchStats|TestTransformReceiveStats' -count=1`; expect missing API or counter assertions to fail.
- [ ] **Step 3: Implement counters.** Use atomic fields in a per-watch stats cell keyed by credential under a short server mutex; allocate after `BeginWatch`, remove on completion. Record transform-loop `endpoint.Observe` outcomes and each encode/send-stage outcome and latest tick in `sendTransforms`; do not count the separate reliable watch loop as transform observation. Use client atomics for raw reads/decode failures and `datagramMu` for active state. Keep counters observational.
- [ ] **Step 4: Run focused tests and commit.** Same command with `-count=3`; expect pass. Commit only Task 1 files.

### Task 2: Keep the hosting watch alive through a post-correction transform

**Files:** Modify `internal/app/clientsession/live_gameworld_test.go`; add focused diagnostic/lifetime tests in the same package only if the existing connected test cannot prove the behavior cleanly.

**Interfaces:** The fixture retains `*sessionquic.Server` for server stats. The hosting call to `assertLiveMovement` runs a post-correction callback before its deferred `stopWatch`; the second-client call has no transform callback. `waitForTransformDatagram` takes a post-correction decoded baseline and phase diagnostics.

- [ ] **Step 1: Record a RED witness.** Under the old ordering, wait for `TransformWatchActive == false` while `ctx.Err() == nil`, then baseline `TransformsReceived` and require an increase. Run the focused live test and record the expected deadline failure; discard only this temporary synchronization after the witness. No arbitrary sleep.
- [ ] **Step 2: Implement the live-watch assertion.** Read `session.reliableWorld` under `session.mu` and wait for its tick and origin to advance before declaring hosting movement corrected; combined `Session.World` alone can advance from a datagram. Capture that reliable correction tick, server/client baselines, and phase start before waiting for a later decoded datagram. Poll while its watch remains live for `TransformsReceived > baseline`; keep a test-owned last server/client sample each tick. On timeout report baseline/latest counters, latest observed/sent ticks, correction and replay ticks, elapsed and deadline remaining at phase start. If the shared deadline has already elapsed, report that phase directly. Cancel only after assertion or failure unwinds.
- [ ] **Step 3: Run focused repetitions.** `go test -race -p 1 -parallel 1 -timeout 5m ./internal/app/clientsession -run '^TestConnectSelfHostedEntersLiveGeneratedGameworld$' -count=20`; expect pass with the existing reliable movement and HUD assertions intact. Add and run a deterministic focused helper test proving the saved last server/client sample survives accessor disappearance and watch cancellation.
- [ ] **Step 4: Run required local gates on the committed tree.** `make test`, `make test-race`, `make architecture`, `make lint`, and `git diff --check`; record exact outputs and owned-game-data availability. Commit Task 2, rerun applicable gates on HEAD, then request independent correctness and formatting/style reviews, including `refactor-for-readability`. Complete PR template and push. Require the initial full macOS race CI and one rerun of its full race job to pass before merge; if either fails, inspect the new phase diagnostics before changing the test.
