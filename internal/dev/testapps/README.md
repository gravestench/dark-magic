# Manual test applications

These repository-private programs exercise engine slices without becoming
shipped products:

- `audio_file` performs manual audio-file playback checks.
- `scene_demo` runs the standalone interactive scene harness.

The asset-backed composite laboratory is an integrated engine scene rather than
a second renderer. Start it with `go run ./cmd/client
--start-scene=composite_lab`; see the root README for recipe flags and controls.
- `shell` runs the renderer-free Charmbracelet frontend for the shared Lua
  shell used by future client, game-server, and realm targets.
- Item-generation diagnostics now belong to the authoritative `d2legacy` Lua
  mod. Add future probes beside that policy instead of reviving native Go rule
  executables.

Run one from the repository root, for example:

```shell
go run ./internal/dev/testapps/scene_demo
```
