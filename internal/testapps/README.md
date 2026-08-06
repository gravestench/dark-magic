# Manual test applications

These repository-private programs exercise engine slices without becoming
shipped products:

- `audio_file` performs manual audio-file playback checks.
- `loot_roll` and `quality_roll` diagnose deterministic item generation.
- `scene_demo` runs the standalone interactive scene harness.
- `treasure_roll_lua` exercises the historical Lua treasure-roll fixture.

Run one from the repository root, for example:

```shell
go run ./internal/testapps/scene_demo
```

