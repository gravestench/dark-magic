# Performance acceptance

Performance changes are accepted from matched workloads, not isolated intuition.
The checked budgets cover decoded and native residency, resource counts, decode
time, and rolling per-scene frame/update percentiles. Frame timing uses the most
recent 512 samples so long sessions remain bounded.

## August 2026 frontend localization pass

A matched real-asset capture reproduced the trademark/title-to-main-menu hitch
with per-scene profiling enabled. The before capture spent 3.77 seconds of CPU
time in `tbl_text.UnmarshalReaderAt` and 4.24 seconds in file syscalls. The three
English TBL files total only about 602 KiB, but the decoder's fine-grained hash
and key reads became thousands of random reads when forwarded directly to the
compressed MPQ filesystem.

Localization now reads each TBL sequentially once and decodes the buffered
bytes. A regression filesystem intentionally advertises `io.ReaderAt` and
asserts that locale loading never invokes it, preserving the archive boundary
independently of the concrete mounted filesystem.

| Slice | Before | After | Result |
| --- | ---: | ---: | ---: |
| Title-scene maximum update | 4,134.026 ms | 152.096 ms | 96.3% lower |
| TBL `UnmarshalReaderAt` CPU | 3.77 s | absent from profile | compressed-archive random-read path removed |
| Main-menu update p50 | 0.146 ms | 0.159 ms | effectively unchanged |
| Main-menu update p95 | 0.699 ms | 0.794 ms | remains sub-millisecond |

The after heap profile still retained about 357 MB under eager frontend asset
preloading (about 73% of the 487 MB profiled Go heap). That is a separate,
measured residency problem and is not claimed as fixed by the localization
change. Per-scene profiling forces collection at scene boundaries, so the
remaining 152 ms maximum is instrumentation-inclusive.

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

Capture the frontend transition with real assets:

```shell
MPQ_DIRECTORY=/path/to/diablo-ii go run -tags ffmpeg ./cmd/client \
  --profile-dir profiles/live-menu-transition --profile-scenes all
```

Skip the trademark screen, wait for the main menu to settle, and quit normally
so the CPU, heap, diagnostics, and per-scene artifacts are finalized.

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

## Deferred renderer opportunities

Revisit the following only after representative dense gameplay makes native
rendering or draw submission a measured bottleneck. The current Blood Moor
native-render stage is about 0.38 ms, so these would add complexity without a
material frame-time benefit today.

- Pack stable DT1, DC6, or DCC surfaces into padded texture atlases or texture
  arrays when profiles show texture switching or draw submission dominating.
  Preserve nearest filtering, transparent borders, semantic cache identity,
  palette behavior, and independent resource lifetimes.
- Instance repeated world tiles, missiles, particles, ground items, or creature
  layers when dense scenes produce thousands of compatible visible placements.
  Batch keys must include texture/atlas, palette, shader, blend mode, and clip.
- Separate static world geometry from dynamic entities so floors and unchanged
  walls can use chunk meshes or persistent instance buffers without delaying
  moving creatures, missiles, overlays, or animated objects.
- Add aggregate bounds and a quadtree or spatial render hierarchy when retained
  offscreen regions make traversal scale with explored world size. A simple
  spatial-parent experiment regressed the current workload, so a future design
  must update bounds incrementally and demonstrate fewer visited nodes.
- Consider indexed GPU sprite storage and palette lookup for large creature or
  effect populations only when RGBA residency or upload bandwidth exceeds its
  checked budget. Preserve deterministic decoded output and golden captures.
- Revisit decode-buffer pooling and explicit renderer acknowledgements if long
  cinematics make frame allocation a live-heap or GC bottleneck. Buffers cannot
  be safely reused until the owner thread has consumed the queued upload.

Suggested triggers for reopening this work are sustained native rendering above
2 ms p95, draw submission among the top CPU-profile entries, frame time growing
with offscreen retained nodes, repeated upload-budget failures, or a supported
dense combat fixture missing its frame-time budget. Every accepted redesign
should include matched dense-world and combat captures plus a simpler scene to
catch regressions caused by extra hierarchy or batching overhead.
