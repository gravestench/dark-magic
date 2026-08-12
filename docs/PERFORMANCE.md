# Performance acceptance

Performance changes are accepted from matched workloads, not isolated intuition.
The checked budgets cover decoded and native residency, resource counts, decode
time, and rolling per-scene frame/update percentiles. Frame timing uses the most
recent 512 samples so long sessions remain bounded.

## August 2026 gameplay pass

Measurements below were captured on an Apple M1 with Go 1.25.12. Microbenchmark
figures are representative medians from five runs. The gameplay comparison used
the same Blood Moor fixture, pointer movement, 180 stable capture frames, release
assets, and CPU/heap profiling in both revisions.

| Slice | Before | After | Result |
| --- | ---: | ---: | ---: |
| Visible tile query | 6.20 µs, 6,496 B, 17 allocs | 1.41 µs, 0 B, 0 allocs | 77% faster; steady-state allocations removed |
| 512 redundant node updates | 59.8 µs, 343,823 B, 522 allocs, 512 changes | 21.7 µs, 81,920 B, 512 allocs, 0 changes | 64% faster; renderer traffic removed |
| Four pending video frames | 390 ns, 1,792 B, 3 allocs, 4 changes | 128 ns, 256 B, 1 alloc, 1 change | 67% faster; stale uploads removed |
| 24-system ECS tick | 2.42 µs, 1,920 B, 49 allocs | 1.02 µs, 0 B, 0 allocs | 58% faster; no-op tick allocations removed |
| Node bounds culling | 7.75 ns | 4.26 ns | 45% faster |
| 800x600 NRGBA pixel preparation | 6.7 ms, 3.85 MB, 480,001 allocs | 3.25 ns, 0 B, 0 allocs | contiguous video buffers bypass conversion |
| Blood Moor update p50 | 3.84 ms | 4.42 ms | instrumentation-inclusive |
| Blood Moor update p95 | 5.54 ms | 5.81 ms | instrumentation-inclusive |
| Blood Moor frame p50 | 18.00 ms | 18.05 ms | effectively unchanged |
| Blood Moor frame p95 | 18.24 ms | 18.25 ms | frame-capped; unchanged |

Short-run p99 remains sensitive to asynchronous asset completion and host
scheduling, so it is recorded in raw diagnostics but not presented as a gain.
The steady-state p50/p95 comparison is the useful end-to-end result.

The final diagnostics separate simulation (0.98 ms p95), Lua scene work
(5.02 ms p95), composition drain (0.002 ms on the sampled settled frame), and
native render traversal/draw submission (0.38 ms). Because native rendering is
well below one millisecond in the representative world, an additional custom
tile batch was measured as unjustified: Raylib already batches compatible
DrawTexturePro submissions, while a spatial-parent experiment increased node
visits and update p95 and was reverted.

Real `New_Bliz640x480.bik` playback decoded at 18–19 MB/s. A 319-frame startup
capture presented 99 video texture updates (about 126 MiB), spending 211 ms in
native uploads in total, or about 2.13 ms per presented update. Pending frames
are coalesced and contiguous NRGBA upload avoids the former per-pixel conversion.

## Reproduction

Run the microbenchmarks:

```shell
go test ./internal/presentation/maprender ./internal/game/ecs \
  ./internal/presentation/render ./internal/platform/raylib/renderer \
  -run '^$' -bench . -benchmem -count=5
```

Capture the real-asset acceptance matrix and enforce budgets:

```shell
MPQ_DIRECTORY=/path/to/diablo-ii make profile-acceptance
make profile-check
```

Raw profiles remain ignored under `profiles/acceptance`; the budget file and
this matched summary are the reviewable repository artifacts.
