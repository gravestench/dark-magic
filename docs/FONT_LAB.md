# Font lab

The `font_lab` scene renders Diablo II bitmap fonts through the same manifest,
TBL/DC6 decoder, palette, PL2 transform, layout, and retained-rendering paths as
the normal interface. It is intended for visual diagnosis, not as a separate
font preview implementation.

Launch it directly with a directory containing supported game archives:

```sh
MPQ_DIRECTORY="~/my-mod,~/d2_english_mpq" go run -tags ffmpeg ./cmd/client --start-scene font_lab
```

Use Right, Down, Enter, or the primary mouse button to advance. Use Left or Up
to go back, and Escape to return to the main menu. The five pages cover:

1. semantic styles used by real screens;
2. decoded bitmap-font families under one common transform;
3. every named inline PL2 text-color slot;
4. the same glyphs and color slot under different contextual PL2 files;
5. alignment, wrapping, multiline layout, and inline-color continuity.

Inline colors can be scoped explicitly. `[red]danger[/]` colors only `danger`
and restores the caller's original tint afterward. `[/color]` and `[reset]`
are readable aliases for `[/]`. A color token without a reset remains active
for the rest of the string for compatibility with existing shim content.

When reporting a visual mismatch, include the page number and the visible row or
style label. That makes the suspect font, source palette, PL2 transform, color
slot, and layout path unambiguous.
