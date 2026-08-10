# Dark Magic Lua shim examples

The embedded shim is both Dark Magic's first-party presentation layer and the
canonical set of examples for mod authors. It uses only versioned capabilities
available through `require`, plus ordinary Lua modules under `darkmagic.*`.

## Structure

- `darkmagic/screens` contains root navigation scenes.
- `darkmagic/overlays` contains short-lived panels pushed above a root scene.
- `darkmagic/ui` contains reusable Lua helpers with no native ownership.

Asset paths, palette choices, timing, layout, and localization keys belong in
versioned shim manifests. Lua should describe composition and interaction; Go
capabilities should own decoding, rendering, audio devices, persistence,
simulation, and other native resources.

## Lifecycle and ownership

A scene may implement `create`, `enter`, `update`, `render`, `exit`, and
`destroy`. Use only the callbacks the scene needs. Values stored on `self` are
instance state. Checked render, audio, video, subscription, and callback handles
created during a component or scene callback belong to its resource scope and
are reclaimed when that scope closes.

Use `scenes.replace` for root-screen transitions, `scenes.push` for overlays,
and `scenes.pop` to dismiss the top overlay. An overlay can set
`blocks_update_below` to pause or continue scenes beneath it. The `focused`
argument to `update` distinguishes continued simulation from input ownership.
The `dm.input/v1` capability enforces that boundary. Nonfocused callbacks
normally observe no actions, text, or pointer coordinates even when they keep
updating. An overlay may explicitly set `passes_input_below = true`; callbacks
beneath it then receive only gameplay actions and the pointer position, never
UI actions or text. The same overlay declares `world_view = "left"`, `"right"`,
or `"center"`, which is supplied as the fourth update argument so the world can
frame the player inside the unobscured region. `input.owner()` still reports the
single `scene`, `debug`, or `none` UI owner for diagnostics; scenes should not
use it to coordinate navigation.

## Style

- Use four spaces for indentation and no statement semicolons.
- Put one field or argument per line when a call or table becomes difficult to
  scan.
- Prefer early returns for unavailable optional capabilities and modal states.
- Explain ownership, policy, coordinate systems, and non-obvious game-format
  behavior; avoid comments that merely repeat a function name.
- Keep controls as plain data and update their visuals through callbacks.
- Read presentation facts from manifests instead of introducing path or layout
  literals in Lua or Go.

Every Lua file can be syntax-checked independently with `luac -p`. Engine-level
behavior is exercised by the Go acceptance tests, which boot the embedded shim
through the same versioned capability boundary used by the client.
