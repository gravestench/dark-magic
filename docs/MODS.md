# Mod loading and distribution

Dark Magic treats the engine, installed packages, enabled profiles, and live
session mod set as four different things. This separation lets the client start
with no mods, keeps a downloaded package from executing merely because it is
present, and gives network admission one exact identity to verify.

## Startup model

1. The native engine starts independently of game mods.
2. Product distribution code reconciles redistributable bundled packages into
   the local mod cache. The current distribution bundles `d2legacy`.
3. A user profile names enabled package IDs in deterministic order. A missing
   profile is initialized with `d2legacy`; an existing empty profile stays
   empty.
4. The resolver verifies every selected immutable blob, walks dependencies,
   rejects missing packages, cycles, and version mismatches, and emits a lock
   containing exact package digests.
   A non-empty set contains exactly one `game` package plus zero or more
   `extension` packages.
5. Every package is privately mounted at `mods/<id>/`. Only top-level
   `content_roots` declared by its manifest also enter the shared overlay.
   Dependencies activate before dependents. Shared resource lookup is the
   reverse: later profile entries and dependents override prerequisites.
6. Only the resolved lock is mounted and its declared entry components become
   eligible for activation. An empty lock starts the mod-neutral native shell.

This follows the useful parts of established systems without inheriting their
legacy constraints. OpenMW separates installed data paths from enabled content
lists and makes load order explicit. Godot loads independent ZIP/PCK resource
packs and explicitly warns that executable packs require a signature policy.
Unreal Game Features dependency-sort separately activatable features. OCI/TUF
use immutable digests, declared sizes, and signed metadata rather than trusting
a mutable filename.

Primary references:

- [OpenMW mod installation and content lists](https://openmw.readthedocs.io/en/openmw-0.49.0/reference/modding/mod-install.html)
- [Godot packs, patches, mods, and security concerns](https://docs.godotengine.org/en/stable/tutorials/export/exporting_pcks.html)
- [Unreal Game Features dependency and activation model](https://dev.epicgames.com/documentation/en-us/unreal-engine/API/Plugins/GameFeatures)
- [The Update Framework target descriptors](https://theupdateframework.github.io/specification/draft/)
- [OCI artifact digest identity](https://oras.land/docs/1.2/concepts/artifact/)

## Local layout

The default locations use the operating system's cache and configuration
directories:

```text
<user-cache>/dark-magic/modcache/
  index.json
  blobs/sha256/<digest>.zip
  quarantine/

<user-config>/dark-magic/mods.json
```

`DARK_MAGIC_MOD_CACHE` and `DARK_MAGIC_MOD_PROFILE` override these paths for
development and isolated tests. `index.json` is a mutable local name-to-digest
reference. Blob paths are immutable content identities. `mods.json` currently
uses this deliberately small shape:

```json
{
  "schema": "dark-magic.mod-profile/v1",
  "enabled": ["d2legacy"]
}
```

Use an empty `enabled` array to start without mods. The engine reconciles the
bundled `d2legacy` archive into the cache on every version, but does not rewrite
an existing user's enablement choice.

`client -mods none` (or `DARK_MAGIC_MODS=none`) is a one-run recovery override
that starts the mod-neutral shell without rewriting `mods.json`. A comma-
separated value temporarily selects another installed set in profile order.

The former `DARK_MAGIC_MOD_DIRECTORY` overlay is intentionally not part of
production startup: a mutable directory would let code and data differ from the
resolved lock. A future author mode may mount a workspace only while networking
and trusted persistence are disabled, and will label that session unverified.

## Package metadata and locks

Every package contains `mod.json` using `dark-magic.mod/v1`. It declares stable
ID, version, game/extension kind, engine API, redistributability, dependencies,
requested capabilities, shared content roots, and client/authority component
entry points. The local cache descriptor supplies the archive SHA-256 digest
and compressed byte size, avoiding a self-referential digest inside the archive.

## Package namespaces and shared content

Installing `d2legacy` produces this logical VFS shape even though the immutable
ZIP keeps ordinary package-relative paths:

```text
mods/d2legacy/mod.json
mods/d2legacy/boot.lua
mods/d2legacy/components/...
mods/d2legacy/lua/...

assets/...       declared shared content
data/...         declared shared content
locales/...      declared shared content
manifests/...    declared shared content
```

Thus two packages may both contain `boot.lua` without ambiguity. Code,
metadata, and ordinary Lua modules remain package-private. `components`, `lua`,
and `mods` are reserved and cannot be exported as shared roots. `d2legacy`
declares `assets`, `data`, `locales`, and `manifests`; the empty template
declares none.

Lua module ownership follows the same rule. `require("example.foo")` resolves
only from `mods/example/lua/example/foo.lua` (or the root-module equivalent),
never from another package that happens to contain the same relative path.
Stable package IDs and Lua namespace prefixes should therefore match.

For a dependency chain `base -> extension`, activation is `base, extension`
while shared lookup is `extension, base, external game data`. An extension may
therefore override `assets/ui/example.dc6` deliberately, but cannot replace
`mods/base/boot.lua`. A later enabled sibling extension wins over an earlier
one. Package locks record that deterministic order.

The resolved `dark-magic.mod-lock/v1` is the unit that sessions, replays,
checkpoints, and realm admission must pin. A user profile is never sufficient
network identity because its mutable names can resolve differently later.

## Trust and acquisition direction

There are three distinct trust paths:

- bundled: authenticated by the engine distribution and reconciled locally;
- local author package: explicitly installed and capability-approved by the
  user, but not treated as realm-trusted;
- realm package: accepted only through signed, expiring/versioned metadata that
  binds digest, size, dependencies, engine compatibility, capabilities, and
  redistributability.

P2P peers will remain untrusted blob/chunk mirrors. Downloads stay in bounded
quarantine until every chunk and the complete package match the signed realm
manifest, then promote atomically to the immutable cache. MPQs, Blizzard game
data, player saves, credentials, and private keys are never redistributable.

## Remaining composition extraction

The generic content stack and cache no longer select `d2legacy`, and the client
has a real empty-mod startup path. The current playable client still contains a
compiled first-party `d2legacy` adapter for Diablo-specific save, map, and
presentation bridges. The next slice replaces that direct assembly with a
distribution-registered game-adapter interface keyed by the resolved game
package contract. Go plugins are intentionally not the portability boundary;
portable mod behavior remains Lua plus data, while narrowly compiled adapters
are product registrations rather than engine imports.
