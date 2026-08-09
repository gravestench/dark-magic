# UI lab

The `ui_lab` scene exercises Diablo II UI controls through the same Lua,
retained-rendering, input, bitmap-font, audio, and asset paths used by normal
screens. It is a compatibility/diagnostic scene, not a parallel GUI toolkit.

Launch it directly with a directory containing supported game archives:

```sh
MPQ_DIRECTORY="~/my-mod,~/d2_english_mpq" go run -tags ffmpeg ./cmd/darkmagic --start-scene ui_lab
```

The scene currently covers every control-manager role plus the shared visual
helpers they depend on:

- authored DC6 button, including normal/focused/pressed behavior and tooltip;
- disabled button and focus exclusion;
- authored `clickbox.dc6` checkbox frames;
- authored character-name `textbox.dc6` text field;
- bounded scrollbar/slider interaction;
- bitmap text and tooltip rendering;
- mouse hit testing, keyboard/controller focus, activation callbacks, text input,
  and the regular Diablo II cursor.

Recovered presentation values live in `darkmagic.ui.compat` and reusable widget
modules rather than in this lab. Shipping scenes and mods therefore exercise the
same implementation that the lab displays.

The frontend logo is also a useful draw-mode regression target: Diablo II draw
mode 3 uses `GL_ONE, GL_ONE_MINUS_SRC_COLOR`, represented by the retained
renderer as `screen`. This is intentionally distinct from generic additive
blending.
