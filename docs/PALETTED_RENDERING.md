# Paletted rendering evaluation

Dark Magic currently expands DC6 and DCC color indexes through their selected
palette before uploading RGBA textures. Profiling showed that avoiding generic
per-pixel conversion and duplicate animation surfaces provides a direct benefit
without changing composition semantics.

An indexed GPU representation would reduce texture residency for uncomposited
DC6/DCC frames from four bytes per pixel to roughly one byte per pixel plus a
palette texture. It is not yet the correct default because several current paths
produce RGBA output before rendering:

- normalized DC6 animations place cropped frames on a shared transparent canvas;
- COF character animation composites multiple DCC components with shadows,
  transparency, and draw effects;
- frontend logo flames use non-alpha blend modes;
- the headless backend and golden fixtures consume ordinary Go images.

The implementation would therefore require an indexed texture resource, palette
resource binding, a backend shader path, indexed transparency rules, and either
shader-based COF composition or an RGBA fallback. That is a renderer-backend
feature rather than a safe cache optimization.

The decision is to retain the RGBA path for composed presentation assets while
preserving `ResourcePalette` and the indexed codec data needed for a later world
and character renderer prototype. Adoption should be reconsidered when DCC/COF
composition moves onto the GPU, and accepted only when captured profiles show a
meaningful residency reduction without changing palette, blend, or headless
composition results.
