# Empty Dark Magic mod template

This sibling of `d2legacy` is a deliberately small, copyable starting point. It
contains a valid boot component, one optional managed component, a namespace for
ordinary Lua modules, and the directories where larger mods normally grow.

It is reference content, not another first-party game. Installed packages are
kept in the content-addressed mod cache and enabled by the user's mod profile;
merely placing a package in the cache does not execute it. Copy this directory
into a new mod package and replace every `mod_template` identity before adding
it to a profile.

## Rename checklist

1. Rename `lua/mod_template/` to your stable mod namespace.
2. Replace `mod_template.*` IDs and `require` paths.
3. Give the package its own README, manifests, locale, and tests.
4. Keep capability imports versioned, such as `engine.scene/v1`.
5. Declare only the shared `content_roots` the mod actually owns. Never export
   the reserved `components`, `lua`, or `mods` roots.
6. Add host composition for the new package; do not edit d2legacy to smuggle a
   second mod into its runtime.

## Layout

```text
mod_template/
  mod.json                    package identity, compatibility, and entry points
  boot.lua                    presentation lifecycle entry point
  components/
    example.lua               optional independently managed component
  lua/mod_template.lua        ordinary namespace root module
  lua/mod_template/
    bootstrap/README.md       composition/registry destination
    commands/README.md        authoritative command destination
    components/README.md      ECS schema destination
    data/README.md            external-record interpretation destination
    policy/README.md          pure rules destination
    screens/README.md         root scene destination
    systems/README.md         deterministic ECS system destination
    tests/README.md           fixtures and cross-domain test destination
    ui/README.md              reusable presentation destination
```

`boot.lua` and top-level `components/*.lua` return managed component
definitions. Files below `lua/<namespace>/` are ordinary modules loaded with
`require("<namespace>....")`; they are not started merely because they exist.

At runtime the complete package is private beneath `mods/mod_template/`.
Manifest `content_roots` are the only package directories additionally layered
at the shared VFS root. Dependency order controls activation; reverse
dependency/profile order controls shared-resource precedence. Keep code private
and use shared roots only for intentional assets, data, locales, or manifests.
`require("mod_template.foo")` is bound to
`mods/mod_template/lua/mod_template/foo.lua`; another package cannot shadow it.

The empty boot component intentionally has no engine capability dependencies.
Add scene registration only when the mod has a scene to show. This makes the
fresh template discoverable and lifecycle-valid in headless tooling.

For the full production pattern, read
[`../d2legacy/ARCHITECTURE.md`](../d2legacy/ARCHITECTURE.md) and the repository
[mod loading guide](../../../docs/MODS.md).
