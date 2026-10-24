# Issue #635: Live transform watch lifetime and diagnostics

## Intent and boundary

The macOS race gate intermittently reaches the final transform assertion in `TestConnectSelfHostedEntersLiveGeneratedGameworld` with zero received datagrams. This blocks unrelated feature PRs. The test must keep observing the live hosting-client correction stream until it has proved a compact transform datagram arrived after the reliable movement correction. On failure it must show which phase stopped: host observation, server encoding/send, transport receive, or client decode. This issue owns the test watch lifetime and narrowly scoped QUIC diagnostics. It does not change gameplay, wire protocol, production watch lifetimes, race coverage, or the existing ten-second overall deadline.

## Evidence and mechanism

Both attempts of PR #655's CI run 40115134927 failed at `live_gameworld_test.go:49` after about 10.5 seconds, with `TransformsReceived=0`; the PR changes no `clientsession` or `sessionquic` code. Issue #635 records the same failure on PR #633. Twenty focused local race repetitions on #655 passed, so the failure is intermittent. The test currently opens `session.Watch` inside `assertLiveMovement`, defers cancel, returns after a reliable movement correction, and only then checks the transform counter. Client `Watch` opens `WatchTransforms` under that same context; server `sendTransforms` follows the reliable watch stream context. Canceling the watch before checking means no later packet can satisfy the assertion. This is a demonstrated lifetime gap; CI logs do not establish whether the server actually sent a datagram in either failure.

## Approaches

1. **Keep the hosting watch open through the transform assertion (chosen).** Record the decoded counter after reliable movement, wait for it to increase while the watch is live, then cancel. This tests receipt after the correction without changing production behavior.
2. Use a separate transport-only test. This could isolate QUIC behavior but would not verify the connected game-world path that #635 owns.
3. Increase the ten-second deadline. This may hide a slow failure and cannot make a canceled receiver observe a packet.

## Failure diagnostics

Add bounded, read-only QUIC watch counters. On the server, retain per-credential counters only for an active watch: latest successful `Observe` tick, encode attempts/errors, and datagram sends/errors. On the client, extend existing atomic `NetworkStats` with raw datagrams read, decode failures, and whether a transform receiver is active. Record the latest observed and sent ticks on the server. Immediately after the reliable movement correction, the test takes server and client counter baselines and a monotonic phase start time. While the watch is live, it keeps a test-owned last sample on each wait tick, so a deadline that closes the watch cannot erase the evidence. A timeout reports baseline and last sample, elapsed time, deadline remaining at phase start, latest correction tick, and authority replay tick. This distinguishes no new host observation, an encode/send failure, sent but not read, and read but not decoded. If the shared deadline has already expired before this phase starts, report that phase directly. No payload, credential, save, or game data is logged. Counters do not alter packet handling or wire bytes. Remove the server's per-credential entry on watch completion; the test owns only its bounded last sample.

## Regression witness and acceptance

- RED: in a temporary test witness, preserve the old cancellation order, wait until the client transform receiver reports inactive, then record its decoded-count baseline and require an increase. This must time out because the receiver has completed; record that focused failure before implementing the fix. The final integration assertion retains the post-correction baseline but runs while the same watch remains active.
- GREEN: move that post-correction wait inside the hosting client's live watch. Keep the second player's reliable movement check and the hosting player's reliable correction/HUD check. An increase proves a decoded datagram arrived after the correction; it does not by itself prove the frame contains the movement position.
- Add focused server/client counter tests for successful send/read, malformed datagrams, active-watch teardown, and test-owned last-sample retention across cancellation; avoid sleeping merely to trigger a race.
- Run focused repeated `go test -race -p 1 -parallel 1`, complete `make test` and `make test-race` on the committed change; macOS PR CI must pass before merging. If zero transforms persist, use the new counters to identify and repair the stalled phase within #635.

## Builder's calls

Keep the lifetime fix test-local. Server diagnostics may be a small `sessionquic` watch stats value returned by an accessor keyed by `SessionCredential`; client counters extend the existing `NetworkStats`. Store no per-watch counters after the watch ends. No game data is needed.
