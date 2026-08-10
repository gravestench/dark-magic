# Raylib renderer adapter

This package is the private native backend for the renderer-neutral retained
model in `internal/presentation/render`. It owns the window, Raylib frame loop,
GPU resources, native draw state, logical-resolution target, and final resource
destruction on the main thread.

Engine and Lua code never receive Raylib nodes, textures, cameras, or shaders.
They create checked retained nodes and resources through versioned capabilities;
the composer sends ordered changes to this adapter. The adapter keeps only the
private native state required to draw those changes.

## Ownership rules

- CPU decoding happens outside this package.
- GPU creation, updates, drawing, and destruction happen inside this package on
  the owner thread.
- Animation clocks are adapter-owned playback state, never callbacks attached to
  scene nodes.
- Hidden parents cull their complete subtree.
- Parent and child clips are intersected before drawing each node.
- Transform matrices are recomputed only after a local or ancestor change.
- Shutdown drains pending composition changes, releases backend nodes and palette
  effects, clears GPU caches, and only then closes the native context.

`BackendDiagnostics` exposes cumulative work counters without leaking native
handles. Profiling records these beside composer and texture-cache diagnostics.

## Texture upload policy

Raylib remains the native backend. Engine-produced contiguous `image.RGBA`
surfaces are wrapped as `UncompressedR8g8b8a8` and uploaded in one call. Do not
send these through raylib-go's generic `NewImageFromImage`: that helper crosses
the cgo boundary once per pixel. Non-contiguous or unusual Go images use the
generic conversion fallback, whose temporary native image is released
immediately after upload.

Background residency is optional and drains after visible presentation. It has
both byte and wall-clock budgets, whereas a texture demanded by a visible node
is never denied. Known immutable asset frames use generation-qualified semantic
keys so scene re-entry does not rescan all pixels merely to identify a texture.
