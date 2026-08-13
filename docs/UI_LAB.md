# UI lab

The `ui_lab` scene exercises Diablo II UI controls through the same Lua,
retained-rendering, input, bitmap-font, audio, and asset paths used by normal
screens. It is a compatibility/diagnostic scene, not a parallel GUI toolkit.

Launch it directly with a directory containing supported game archives:

```sh
MPQ_DIRECTORY="~/my-mod,~/d2_english_mpq" go run -tags ffmpeg ./cmd/client --start-scene ui_lab
```

The scene now covers the reusable UI surface rather than only the original four
control-manager roles:

- authored DC6 button, including normal/focused/pressed behavior and tooltip;
- disabled button and focus exclusion;
- text-only / label button;
- authored `clickbox.dc6` checkbox frames;
- authored character-name `textbox.dc6` text field;
- semantic slider with keyboard adjustment, step snapping, pointer capture, and
  drag updates;
- authored Diablo text scrollbar using recovered `TextSlid.dc6` arrow, gutter,
  and thumb frames;
- selectable/paged list with distinct selection and repeated-activation hooks;
- mutually-exclusive tab / selection group;
- panel/container for grouped retained presentation;
- progress bar driven by the slider value;
- standalone tooltip with viewport clamping;
- focus-isolated modal confirmation dialog;
- bitmap text and the regular Diablo II software cursor.

The shared `d2legacy.ui.controls` manager owns focus, accessibility roles,
pointer capture, mouse-up activation, range dragging, text editing, and keyboard
behavior. Widget modules own presentation and widget-specific composition, so
shipping screens and mods use the same semantics shown by the lab.

Recovered presentation values live in `d2legacy.ui.compat`. In particular, the
text scrollbar uses the recovered `TextSlid.dc6` frame mapping: down/up hollow
arrows `8/9`, down/up filled arrows `10/11`, gutter `13`, and thumb/fill `14`.
The original executable also references `OptBar.dc6` and `OptBarC.dc6` for
options sliders. Asset inspection verifies `OptBarC.dc6` as a two-frame
255-by-37 bar and `OptSkull.dc6` as a one-frame 28-by-28 thumb. The in-game
sound and music controls use those authored assets and update validated
`engine.settings/v1` preferences in the `0..1` range. Riiablo independently uses
that range for music/effects volume, while AbyssEngine validates the same range
for its separate music, SFX, UI, and master channels. OpenDiablo2 supplies the
canonical paths and menu hierarchy; OpenD2 is consulted for retained panel and
input-order behavior but does not currently implement this options surface.

The frontend logo is also a useful draw-mode regression target: Diablo II draw
mode 3 uses `GL_ONE, GL_ONE_MINUS_SRC_COLOR`, represented by the retained
renderer as `screen`. This is intentionally distinct from generic additive
blending.
