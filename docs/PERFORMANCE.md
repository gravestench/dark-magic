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
| Visible tile query | 6.20 µs, 6,496 B, 17 allocs | 1.72 µs, 2,041 B, 8 allocs | 72% faster, 69% fewer bytes |
| 512 redundant node updates | 59.8 µs, 343,823 B, 522 allocs, 512 changes | 21.7 µs, 81,920 B, 512 allocs, 0 changes | 64% faster; renderer traffic removed |
| Four pending video frames | 390 ns, 1,792 B, 3 allocs, 4 changes | 128 ns, 256 B, 1 alloc, 1 change | 67% faster; stale uploads removed |
| 24-system ECS tick | 2.42 µs, 1,920 B, 49 allocs | 2.38 µs, 1,728 B, 48 allocs | immutable schedule removes the per-tick copy |
| Node bounds culling | 7.75 ns | 4.26 ns | 45% faster |
| 800x600 NRGBA pixel preparation | 6.7 ms, 3.85 MB, 480,001 allocs | 3.25 ns, 0 B, 0 allocs | contiguous video buffers bypass conversion |
| Blood Moor update p50 | 3.84 ms | 3.85 ms | effectively unchanged |
| Blood Moor update p95 | 5.54 ms | 5.44 ms | 1.8% lower |
| Blood Moor frame p50 | 18.00 ms | 17.96 ms | effectively unchanged |
| Blood Moor frame p95 | 18.24 ms | 18.24 ms | frame-capped; unchanged |

The optimized gameplay run had a larger p99 cold-work outlier. It is deliberately
not presented as a gain: p99 remains sensitive to asynchronous asset completion
and host scheduling in this short capture. The steady-state p50/p95 comparison is
the useful result.

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
