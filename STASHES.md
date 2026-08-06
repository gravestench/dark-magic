# Historical stash inventory

These stashes were reviewed on 2026-08-05 and deliberately left unapplied:

- `stash@{0}` — **Legacy Lua runtime experiment** (`feature_modloader`): global
  Lua table conversion, old plugin exports, terminal mod, and tween callbacks.
  Superseded by the serialized capability runtime; retain only for archaeology.
- `stash@{1}` — **Renderer dependency pin** (`feature_rendering_and_gui`): a
  two-file `go.mod`/`go.sum` dependency adjustment. Review only if bisecting the
  historical renderer branch.
- `stash@{2}` — **Stats rewrite snapshot A** (`filter-branch: rewrite`): expanded
  legacy stats service and tests. It predates the internal-host architecture.
- `stash@{3}` — **Stats rewrite snapshot B** (`filter-branch: rewrite`): same
  file set and summary as snapshot A; treat it as a duplicate until hashes are
  compared during any future stats-port effort.

None should be applied wholesale. Any useful behavior should be ported as a
focused change with current capability and lifecycle tests.
