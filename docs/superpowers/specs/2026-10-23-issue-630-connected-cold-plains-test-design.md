# Issue #630: observe one connected frame and finish live assertions before offline replay

## Intent and boundary

The owned-data connected Cold Plains population test must compare a client's installed map, player projection, and HUD from the same frame, then finish its live crossing and population assertions while the client still reconciles each authoritative correction. This issue changes the test synchronization and ordering only. It does not alter transport, 64-tick event history, the semantic gap guard, map activation, monster population, or game rules.

## Evidence and root cause

`moveConnectedAcrossColdPlains` advances one client frame, then reads `clientsession.Session.View()`. The QUIC watch worker can publish a newer HUD between those operations. A frame with level 2 installed into the map and replica could then be compared with a `View()` already reporting level 3. This mechanism is consistent with the observed `connected map = 2, want 3` failure, though that run did not capture the exact interleaving. `clientWorld.lastHUD` records the HUD passed through `presentLocalPrediction` by the completed frame. A new presentation revision activates its map before that HUD is stored; a reused revision retains the already installed map. Checking `lastHUD` therefore aligns the assertion with installed state.

Later, `TestConnectedColdPlainsPopulationMatchesAuthority` calls `verifyProductionPopulationReplay` before returning across the seam. That helper composes a fresh offline fixture and replays checkpoints synchronously while the listen host continues ticking at 25 Hz. The client only advances in `fixture.advanceGame`. Event views retain 64 ticks; after a long replay, the next client frame can correctly reject a `FromTick` beyond `lastEventViewTick+1`. Observed candidate failures such as `after tick 19 (starts at 84)` occurred at the return-crossing call after replay verification. This timing hazard is consistent with those failures, though the exact elapsed tick path was not instrumented. The gap guard is behaving as designed.

The test passed five consecutive isolated runs on merged main, consistent with the observed intermittent timing. The code path and event-window tick arithmetic establish the proposed mechanisms more directly. Earlier clean-main and candidate failures are recorded in #630 and #631. No original Diablo II observation is relevant to these test/transport timing invariants; `docs/NETWORKING.md` and the current implementation govern them.

## Approaches

1. **Use the frame-installed HUD and perform offline replay verification last (chosen).** `moveConnectedAcrossColdPlains` reads `app.clientWorld.lastHUD` after each `advanceGame` call, so map and HUD come from one installed correction. Retain the replay snapshot and existing population assertions, but perform `verifyProductionPopulationReplay` only after both connected return crossings and stable resident-ID checks, when no more client frames need processing. This changes no production contract and removes the client starvation introduced by test verification.
2. Increase event history or skip gap validation. This would mask missed reliable cues and change production behavior without evidence of a transport defect.
3. Pause the host or disconnect/reconnect around replay verification. This changes the scenario being tested and may hide a real connected-path failure.

## Acceptance and verification

- Preserve the existing connected test's checks: first crossing, client population equal to an acknowledged host checkpoint, visible monster data, offline replay equivalence, return crossing, and stable resident IDs. The replay check moves after live assertions but still uses the captured replay.
- The prior owned-data `TestConnectedColdPlainsPopulationMatchesAuthority` failures provide red evidence. Run it repeatedly with owned 1.14d data after the test-only edit and run the full owned-data `clientapp` package order. CI without MPQs skips this test; report that explicitly.
- Run focused connected transition, event-window, and offline recovery tests, then ordinary no-data repository and CI gates on the submitted commit.
- Independent correctness and `refactor-for-readability` style reviews check that the test still covers the same behavior and that no production gap check is weakened.

## Limits

This addresses a frame observation race and a test-induced client starvation hazard. It does not prove that a game client can stop reconciling for longer than 64 authority ticks and catch up without losing semantic cues; that would be a separate production resilience issue if required by a real user scenario. Do not claim original-game parity from this test-only change.
