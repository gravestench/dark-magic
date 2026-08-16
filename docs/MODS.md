# Mod loading and distribution

Dark Magic is a Diablo II implementation, not a game-neutral launcher.
`d2legacy` is therefore an immutable, embedded base package that is present in
every runtime. Users enable zero or more extension packages above that base.
Installation, profile enablement, session selection, mounting, and component
activation remain separate operations.

## Product and startup contract

1. Distribution code identifies the embedded `d2legacy` tree by a canonical
   SHA-256 digest. It is never copied into the user's mod cache.
2. `mods.json` names enabled extension IDs in deterministic order. A missing or
   empty profile means “vanilla d2legacy,” not “no game.”
3. The resolver verifies selected extension archives, walks dependencies,
   rejects cached `game` packages, missing dependencies, cycles, version drift,
   and an attempt to list `d2legacy` as an extension.
4. The result contains exactly one distribution-owned base plus an ordered
   extension lock. Dependencies activate before dependents.
5. Every package is private at `mods/<id>/`. Only manifest-declared
   `content_roots` also join the shared asset/data overlay. Shared lookup reverses
   activation order, so a dependent or later extension wins intentionally.
6. Only declared client entrypoints run in a connected client. Declared
   authority entrypoints run in offline, listen, dedicated, and realm-managed
   game authorities. A connected client's frozen offline authority components
   are disabled.

`client -mods none` and `DARK_MAGIC_MODS=none` are one-run vanilla overrides.
They disable extensions without disabling `d2legacy` or rewriting the profile.
A comma-separated override temporarily selects installed extension IDs.
Profiles written by the earlier cache-owned-base implementation are migrated
once by removing their `d2legacy` entry; other extension choices are preserved.

## Local layout

Defaults use the operating system's cache and configuration directories:

```text
<user-cache>/dark-magic/modcache/
  index.json
  blobs/sha256/<digest>.zip
  quarantine/
  .mutation.lock

<user-config>/dark-magic/mods.json
```

`DARK_MAGIC_MOD_CACHE` and `DARK_MAGIC_MOD_PROFILE` override these paths. The
profile has this intentionally small extension-only shape:

```json
{
  "schema": "dark-magic.mod-profile/v1",
  "enabled": ["example_extension"]
}
```

The mutable index chooses the locally enabled descriptor for a package ID.
Immutable blobs are addressed by digest, so exact versions required by
simultaneous network sessions may coexist even when only one is the profile
default. Cache mutation is cross-process serialized. Downloads and bundled
author packages remain in quarantine until validation and atomic promotion.

The former mutable `DARK_MAGIC_MOD_DIRECTORY` overlay is deliberately absent
from production startup. A future author mode must label sessions unverified
and disable trusted networking/persistence rather than silently diverging from
the session recipe.

## Manifest, namespaces, and ownership

Every package contains `mod.json` with schema `dark-magic.mod/v1`. It declares
ID, version, `game`/`extension` kind, engine API, redistributability,
dependencies, shared content roots, and client/authority component entrypoints.
Only embedded `d2legacy` may be a `game`; cached packages must be extensions and
normally declare a compatible dependency on `d2legacy`.

The archive descriptor supplies compressed byte size and SHA-256 digest. Those
fields stay outside the archive to avoid a self-referential hash. A downloaded
archive is accepted only if its full bytes, size, safe ZIP structure, manifest
ID/version/redistributability, and extension kind all match the authenticated
descriptor.

The logical VFS is:

```text
mods/d2legacy/mod.json
mods/d2legacy/boot.lua
mods/d2legacy/lua/d2legacy/...

mods/example_extension/mod.json
mods/example_extension/boot.lua
mods/example_extension/lua/example_extension/...

assets/...       shared only when exported by content_roots
data/...
locales/...
manifests/...
```

`components`, `lua`, and `mods` are reserved private roots. A
`require("example_extension.foo")` can resolve only from the owning package's
`lua/example_extension/foo.lua`; a sibling cannot spoof it. Component IDs and
manifest entrypoints must begin with the owning package ID. Component
dependencies may target the same package or a dependency declared in its
manifest, never an undeclared sibling. Active package IDs may not overlap as
dot namespaces (`foo` and `foo.bar` cannot be loaded together).

For `d2legacy -> foundation -> feature`, activation is `d2legacy, foundation,
feature`, while shared-resource lookup is `feature, foundation, d2legacy,
external game data`. Lua code remains private even when an extension overrides
a deliberately shared asset.

## Canonical runtime recipe

Network compatibility is not a profile name and not merely a package lock. One
`dark-magic.runtime-recipe/v1` pins:

- the canonical embedded `d2legacy` ID, version, digest, and size;
- every extension's ordered ID, version, digest, size, and redistributability;
- engine API and game-session protocol versions;
- authoritative Lua and gameplay-configuration hashes; and
- the path-independent external game-asset-set digest; and
- the exact versions of every authoritative capability contract actually
  registered by the production runtime.

The complete recipe is stored inside the runtime identity used by allocation,
admission tickets, joins, reconnects, checkpoints, restored sessions, replay
headers, and realm character compatibility. Reordering extensions or changing
any pinned metadata changes identity. A joiner recomputes the recipe from its
mounted bytes before it submits a character or begins a session lease.
The client may download redistributable extension packages, but it never
downloads protected external game assets; its locally computed asset-set digest
must already match.

## Direct-host package transfer

A self-host advertises its recipe over the same authenticated TLS 1.3 QUIC
connection used by the game protocol. The client fetches missing
redistributable extension archives before profile admission, so time spent in
character UI or package acquisition cannot expire a live game membership.

Current transfer constraints are deliberate:

- only extensions present in the host's exact recipe can be read;
- non-redistributable packages are rejected;
- chunks use bounded reliable streams (32 KiB application chunks); QUIC still
  starts with conservative 1200-byte packets and performs path-MTU discovery;
- each connection receives a bounded burst and refill rate;
- a recipe is limited to 64 extensions, 256 MiB per archive, and 2 GiB total;
- clients stream directly into bounded quarantine, verify the complete SHA-256
  archive and manifest, and promote atomically; and
- Blizzard MPQs/game data, saves, credentials, and private keys are never part
  of this package protocol.

The host is an untrusted byte mirror with respect to installation: possession
of bytes never overrides the authenticated recipe or local verification. Direct
self-host transfer is the first P2P source. Join-time composition is
session-scoped: a failed admission or post-composition startup restores the
client's configured extension recipe rather than silently changing its local
profile or leaving the frontend on the host's recipe. Resumable/multi-source chunks,
realm-signed publication and revocation, cache reference counting, and eviction
remain M23 work.

## Executable-code trust and engine APIs

Extensions are executable Lua, not passive texture packs. The manifest no
longer advertises a `capabilities` permission list because the runtime did not
enforce such a list and a decorative security field was misleading. Current
protection includes package/module/component ownership, exact recipes, separate
client versus authority activation, deterministic authority APIs, and native
resource scopes.

Installing or joining a session with an extension still means trusting its Lua
within the modules exposed to that runtime domain. A future capability grant
format must be enforced by isolated per-package/domain runtimes before it may be
added to the manifest. It must not be inferred from imports or documented as a
security boundary before then.

## Tests

Mod-cache tests cover digest/manifest tampering, unsafe ZIPs, dependency order,
cross-process mutation, corrupt-blob recovery, exact multi-version sessions, and
quarantine promotion. Runtime tests cover namespace spoofing, stale
`package.loaded` invalidation, component ownership, authority activation, and
recipe identity. The QUIC suite covers strict operation shapes, transfer rate
limits, and real authenticated host-to-client acquisition.

See the [d2legacy architecture](../internal/content/d2legacy/ARCHITECTURE.md),
the [Modding API Guide](MODDING_API.md), and the
[networking guide](NETWORKING.md) for runtime/test and transport details.
