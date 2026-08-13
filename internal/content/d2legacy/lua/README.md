# Bundled d2legacy mod: start here

For the complete runtime, ECS, system, UI, and testing map, read
[`../ARCHITECTURE.md`](../ARCHITECTURE.md). A minimal copyable sibling lives at
[`../../mod_template`](../../mod_template/README.md).

The embedded `d2legacy` mod is Dark Magic's first-party Diablo II gameplay and presentation package **and living documentation for mod authors**. Its concise runtime namespace is `d2legacy.*`; the canonical package identity is `d2legacy`. If you are new to programming, that is intentional: you should be able to read this mod from the top, follow the comments, change something small, and gradually understand how a game is put together.

You do **not** need to understand all of Lua before starting. Read the comments first. Treat unfamiliar syntax like punctuation in a comic book: keep following the story, then come back to the punctuation later.

## The big idea

Dark Magic deliberately splits responsibilities between Lua and the engine.

Think of the engine as the part that owns dangerous/heavy machinery: files,
decoded game assets, rendering resources, audio devices, deterministic
scheduling, ECS storage, persistence primitives, networking, and other native
state. Lua is the director: through narrow, versioned capabilities it may
describe presentation, submit player intent, and implement authoritative mod
gameplay. The first-party `d2legacy` mod is the intended owner of Diablo II
rules. “Authoritative” does not mean “written in Go.”

A typical mod therefore looks like this:

```text
player input
    |
    v
Lua scene / widget
    |
    +--> read a safe snapshot
    +--> submit an intent or command
    +--> create/update presentation nodes
    |
    v
Dark Magic capability / authoritative Lua handler
    |
    v
deterministic ECS and registered state/resources
```

The `require("engine.something/v1")` calls are the modding API. The `/v1` matters: it is an explicit version boundary. A mod should not reach into Go packages or renderer internals.

Ordinary modules such as `require("d2legacy.ui.button")` are Lua code from the bundled mod. You can read them, copy their patterns, or replace them in another mod.

## Recommended reading order

Do **not** begin with the largest file. This order builds one idea at a time:

1. `../../boot.lua` — how a mod starts and how scene ownership begins.
2. `d2legacy/bootstrap/scene_registry.lua` — how names such as `main_menu` become registered scenes.
3. `d2legacy/bootstrap/overlay_routing.lua` — how a helper can wrap another scene without rewriting it.
4. `d2legacy/screens/loading.lua` — a small real scene using engine capabilities.
5. `d2legacy/screens/title.lua` — retained rendering, animation, audio, and navigation.
6. `d2legacy/ui/controls.lua` — the shared input/focus brain behind widgets.
7. `d2legacy/ui/button.lua` — how one visual widget plugs into that input brain.
8. `d2legacy/screens/main_menu.lua` — several small systems composed into a complete screen.
9. `d2legacy/ui/item_grid.lua` — presentation reads authoritative state and submits intent instead of mutating gameplay directly.
10. `d2legacy/gameplay/world.lua` — ECS components, snapshots, and presentation binding.
11. `d2legacy/screens/game_world.lua` — the world scene joins simulation, presentation, HUD, and overlays.
12. `d2legacy/ui/game_hud.lua` — a larger capstone example.
13. `d2legacy/screens/ui_lab.lua` — a playground showing the reusable widgets together.
14. `d2legacy/screens/composite_lab.lua`, `monster_lab.lua`, `missile_lab.lua`, `combat_lab.lua`, `dt1_lab.lua`, `ds1_lab.lua`, `mapgen_lab.lua`, and `warp_lab.lua` — asset-backed animation, tile, map, generated-zone, and spatial portal instruments. Combat Lab deliberately wraps `game_world.lua` so its collision, culling, depth ordering, composites, HUD, and authoritative combat cannot drift into a lab-only implementation.

After that, browse whichever overlay or widget resembles the thing you want to make.

## Structure

- `d2legacy/bootstrap` wires scene names and shared routing policy together.
- `d2legacy/screens` contains root navigation scenes: title screens, menus, character screens, and the game world.
- `d2legacy/overlays` contains panels that appear above another scene.
- `d2legacy/ui` contains reusable Lua presentation and interaction helpers.
- `d2legacy/gameplay` contains presentation snapshots and composite adapters;
  authoritative rules live in the purpose-named `commands`, `components`,
  `data`, `items`, `loot`, `mapgen`, `policy`, and `systems` directories.

## Tiny Lua glossary

These are the pieces you will see constantly:

- `local x = ...` creates a name that belongs only to this file or block. Prefer `local`; globals are hard to reason about in mods.
- `require("name")` loads another module and returns what that module exported.
- `{ ... }` creates a table. Lua tables can behave like objects, dictionaries, arrays, or configuration records.
- `function thing:method()` is method syntax. Inside it, `self` means `thing`.
- `function thing.method(self)` is the long form of the same basic idea.
- `if condition then ... end` runs code only when the condition is true.
- `for _, item in ipairs(items) do ... end` walks an array in order. `_` means "I do not need this value."
- `for key, value in pairs(table) do ... end` walks key/value entries where order is not important.
- `return value` gives a result back to the caller. At the bottom of a module it is what `require(...)` receives.
- `nil` means "no value." It is also how a table entry is removed.
- `assert(value, "message")` means "this must exist/be true; otherwise stop here with a useful explanation."
- `pcall(fn, ...)` calls something safely and returns whether it succeeded instead of immediately crashing the Lua call.
- `a or b` is commonly used for defaults: use `a` when it exists, otherwise `b`.
- `condition and a or b` is an older Lua-style compact conditional. Read it roughly as "if condition then a else b."

## Scenes: little state machines

A scene is usually just a Lua table containing some lifecycle callbacks:

- `create(self)` — construct scene-owned state/resources.
- `enter(self)` — the scene is becoming active.
- `update(self, elapsed, focused, input_allowed, world_view)` — advance presentation/interaction.
- `render(self)` — optional explicit rendering work.
- `exit(self)` — the scene is no longer active.
- `destroy(self)` — final cleanup or non-resource bookkeeping.

Use only the callbacks a scene needs.

Values stored on `self` are **instance state**. This matters because there can be old/outgoing and new/incoming instances of the same scene during a safe transition.

Checked render, audio, video, subscription, and callback handles created inside a scene/component scope are owned by that scope. The engine reclaims those handles when the scope closes. This is why many mod scenes do not manually free every render node.

## Root scenes versus overlays

Use `scenes.replace("name")` when moving from one root screen to another, such as title -> main menu.

Use `scenes.push(...)` / `scenes.pop()` for genuinely stack-ordered modal surfaces.

Gameplay panels normally use:

```lua
scenes.toggle_overlay(id, "left")
scenes.toggle_overlay(id, "right")
scenes.toggle_overlay(id, "full")
```

The slots let Dark Magic reason about which part of the game view remains visible. Left and right panels can coexist. A full overlay replaces both sides.

An overlay may set:

- `blocks_update_below` — whether lower scenes keep updating.
- `passes_input_below` — whether routed input may reach lower visible scenes.
- `world_view` — which part of the viewport the world should frame into (`left`, `right`, `center`, or `none`).

**Updating is not the same as owning input.** A world may continue simulating under a transparent panel while the panel owns keyboard/menu input.

## Retained rendering

Dark Magic uses retained render nodes. A Lua scene creates a node once and changes its properties later instead of redrawing every primitive manually every frame.

Typical pattern:

```lua
self.picture = render.create("hud", self.root)
self.picture:set_dc6(path, palette, 0, frame)
self.picture:set_position(x, y)
```

The node is a checked handle. The engine owns the native renderer object behind it.

Many APIs use **center positions**, while old Diablo II data often describes top-left positions or common animation anchors. Helpers such as `d2legacy.ui.dc6` exist so that conversion is explained and shared instead of repeated as mystery arithmetic.

## Controls versus visuals

`d2legacy.ui.controls` owns interaction rules such as:

- focus;
- hit testing;
- pointer capture;
- mouse-up activation;
- keyboard/controller navigation;
- text editing;
- slider/scrollbar adjustment;
- accessibility snapshots.

A visual widget such as `button.lua` owns how those states look.

That separation is deliberate. A mod can use the same interaction behavior with completely different art.

A control is mostly plain data plus callbacks. There is no hidden native button object.

Gameplay itself is pointer-first because Diablo II is pointer-first. Pointer
coordinates and button/hold events become movement, interaction, item, and
combat intents; keyboard events supply hotkeys, modifiers, text, and escape
behavior. A future controller adapter may synthesize navigation or a world
target, but Lua systems must not assume controller focus is the primary model.

## Presentation snapshots and authoritative gameplay

Presentation should not quietly become gameplay authority. That restriction is
about responsibility, not language: a Lua module registered as a trusted
`d2legacy` command handler or deterministic ECS system may own gameplay policy;
a panel or retained scene may not.

For example, an inventory panel does **not** directly change the authoritative item table. It reads a copied snapshot so it knows what to draw, then calls an engine capability to request a move.

```text
snapshot says sword is in backpack
        |
        v
Lua draws sword
        |
player clicks equipment slot
        |
        v
Lua submits move intent
        |
        v
d2legacy validates policy through controlled engine APIs
        |
        v
next snapshot shows the result
```

This pattern makes saves, multiplayer, replay, validation, and debugging much
safer. Authoritative Lua uses deterministic scheduling and RNG, declared ECS
access, restricted side effects, stable module/configuration identity, and
registered serialized state so it participates in the same guarantees as a Go
handler.

## Manifest-driven presentation

Asset paths, palette choices, timing, layout, localization keys, and other presentation facts usually belong in versioned mod manifests rather than being scattered through Lua.

Lua presentation code should mostly describe **composition and interaction**.
Authoritative `d2legacy` Lua owns Diablo gameplay policy. Go capabilities own
decoding, native rendering/audio resources, deterministic simulation
mechanisms, persistence primitives, and capability/resource enforcement.

`d2legacy.ui.compat` is a special compatibility catalog: it stores recovered Diablo II presentation facts that have been researched/corroborated. Keeping those facts separate from the widget implementation makes it easier to tell "this is observed D2 behavior" from "this is how Dark Magic chose to implement it."

## Headless-friendly code

You will often see:

```lua
if not render.assets_available() then
    return
end
```

That is not wasted code. Dark Magic can run presentation logic in tests without proprietary game assets mounted. Keeping logic usable without art makes mods easier to validate automatically.

## Preloading is preparation, not ownership

The preload helpers describe expensive asset work that can happen early. Lua does not receive an MPQ decoder or GPU object. It submits descriptions of what will probably be needed; the engine schedules safe background preparation and owner-thread GPU work.

This is a good example of the capability philosophy: expose the useful operation, not the dangerous implementation object.

## Style for living documentation

The bundled `d2legacy` package is also the example mod, so clarity is a feature.

- Use four spaces for indentation and no statement semicolons.
- Prefer `local` names.
- Put one field or argument per line when a call/table becomes difficult to scan.
- Prefer early returns for unavailable optional capabilities and modal states.
- Comment **why** a line or block exists, especially ownership, policy, coordinate systems, lifecycle rules, recovered format behavior, and compact Lua tricks.
- For clever code, assume the reader is bright but completely new to engineering.
- Do not remove a useful explanatory comment merely because an experienced programmer considers the code "obvious."
- Keep controls as plain data and update visuals through callbacks.
- Read presentation facts from manifests instead of introducing new path/layout literals in Lua or Go.

## Validation

Every Lua file can be syntax-checked independently with:

```text
luac -p path/to/file.lua
```

Engine-level behavior is exercised by Go acceptance tests, which boot the embedded mod through the same versioned capability boundary used by the client.

When changing comments only, it is also useful to compare executable Lua after stripping comments/whitespace. The goal of documentation-only work is for the program to behave exactly as before.
