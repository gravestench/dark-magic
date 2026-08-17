# Performance acceptance

Performance changes are accepted from matched workloads, not isolated intuition.
The checked budgets cover decoded and native residency, resource counts, decode
time, and rolling per-scene frame/update percentiles. Frame timing uses the most
recent 512 samples so long sessions remain bounded.

## Native renderer A/B workflow

The interactive client has one backend-neutral desktop contract and two
compile-time native implementations. Raylib remains the default and production
path. `-tags ebitengine` selects the experimental Ebitengine renderer/input
adapter; `-tags raylib` makes the default choice explicit in scripts and CI.
Gameplay, Lua scenes, retained composition, fixtures, profiling, and capture
remain the same code in both binaries.

Compile both clients without launching them:

```shell
make build-client-backends
```

Run the matched Blood Moor comparison against legally obtained Expansion 1.14d
assets:

```shell
MPQ_DIRECTORY=/path/to/diablo-ii make profile-render-backends
```

The target builds both binaries once, disables native audio in both, runs the
same `game_world` fixture with the same viewport and 300-frame settle window,
and writes raw profiles, diagnostics, captures, and
`profiles/render-backends/summary.md`. Override `RENDER_PROFILE_SCENE`,
`RENDER_PROFILE_SETTLE`, or `RENDER_PROFILE_DIR` to test another matched load.
Compare captures as well as timings: a faster backend that renders a different
scene has not won.

Current Ebitengine limitations are explicit. Native audio commands are drained
muted, the developer shell session has no native text overlay, final-display
`pal.dat` quantization is rejected, and per-node palettes use cached CPU
quantization rather than the Raylib GPU lookup shader. These keep the adapter
useful for graphics/input experimentation without pretending it is ready to
replace the Raylib client. The ordinary Raylib client keeps native audio unless
`--native-audio=false` is supplied.

The first real-asset smoke capture completed the `ui_lab` lifecycle and produced
an 800x600 Ebitengine screenshot through the same capture session. CI compiles
both native clients; it does not launch GUI benchmarks on shared runners.

### Initial matched result

The first post-parity Blood Moor run used the default 300-frame settle recipe
on Apple Silicon with Go 1.25.12. Raylib and Ebitengine submitted 150 and 149
draws respectively after Ebitengine gained the same ancestor-visibility,
inherited-clip, and offscreen-culling rules. The two 800x600 captures differed
in 0.75% of pixels, concentrated in time-dependent cursor/actor frames; static
terrain, HUD placement, and camera composition aligned. A focused UI Lab crop
covering both authored buttons was pixel-identical after correcting Ebitengine's
Diablo draw-mode-4 blend to preserve the `ONE_MINUS_SRC_ALPHA` destination
term.

| Backend | Samples | Frame p50 | Frame p95 | Update p50 | Update p95 | Final native render | Draws | Nodes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Raylib | 438 | 17.088 ms | 17.277 ms | 3.656 ms | 7.095 ms | 0.505 ms | 150 | 447 |
| Ebitengine | 480 | 16.666 ms | 16.811 ms | 3.679 ms | 6.594 ms | 0.399 ms | 149 | 451 |

This single simple-world run says the experimental adapter is competitive; it
does not establish a replacement decision. Repeat runs, dense combat, palette-
heavy creatures/effects, native audio, console, and final-palette parity remain
required before changing the default.

With dependencies already downloaded and separate empty Go build caches, the
same machine compiled the full Raylib client in 33.87 seconds and Ebitengine in
22.68 seconds. Immediate cached rebuilds took 0.51 and 0.39 seconds. Ebitengine
was about 33% faster for this fresh compilation sample; ordinary incremental
rebuild time was effectively equivalent.

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

Per-scene profiling forces collection at scene boundaries, so the remaining
152 ms maximum is instrumentation-inclusive.

## August 2026 frontend preload residency pass

The localization after-capture also exposed a separate residency problem:
startup decoded the title, main menu, every secondary frontend destination,
and all four interaction-state animation families for all seven character
classes before the title could appear. That retained about 357 MB beneath the
preloader and 487 MB in the profiled Go heap at the main menu.

Preloading now has three consumer-timed stages. Startup gates only on title and
main-menu assets. Once the main menu is useful, player think time warms
secondary destinations and the visible unselected character actors. Hover,
forward, selected, and back animations are scheduled only by character
creation. Each stage remains an idempotent request bundle behind the same
engine-owned cache and renderer residency authority.

| Settled main-menu slice | Whole-frontend startup | Staged preload | Result |
| --- | ---: | ---: | ---: |
| Profiled Go heap | 486.99 MB | 215.70 MB | 55.7% lower |
| Preloader-retained heap | 357.41 MB | 111.66 MB | 68.8% lower |
| Combined decoded-cache weight | 338.86 MB | 59.15 MB | 82.5% lower |
| Composed-cache weight | 262.37 MB | 30.34 MB | 88.4% lower |
| Cumulative decode time | 4.098 s | 0.910 s | 77.8% lower |
| Main-menu update p95 | 0.794 ms | 0.344 ms | remains sub-millisecond |
| Pending preload requests at exit | 0 | 0 | settled workload complete |

The staged capture remained on the real-asset main menu for 427 profiled frames
after navigation, long enough for every secondary request to complete. A Lua
contract test independently proves idempotence and prevents character
interaction paths from leaking back into startup or main-menu bundles.

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
