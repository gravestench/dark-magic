# Overlay scenes

Every overlay is independently bootable at the canonical 800x600 presentation
resolution. This lets presentation work proceed before its authoritative game,
item, quest, NPC, or network state capability exists.

## Implemented panels

`inventory`, `character`, `skills`, `automap`, `options`, `pause`, `help`,
`quests`, `party`, `stash`, `cube`, `hireling`, `vendor`, and `waypoint`.

## State-bound presentation shells

`quick_skills`, `belt`, `messages`, `move_gold`, `npc_interaction`,
`npc_dialogue`, `item_tooltip`, `ground_items`, `confirmation_dialog`, `death`,
`area_transition`, `player_trade`, `gambling`, `npc_services`, `hireling_hire`,
`chat`, and `overhead_labels`.

Run one with:

```sh
go run -tags ffmpeg ./cmd/darkmagic --start-scene <name>
```

Or capture it with the existing `START_SCENE=<name> make capture` workflow.
Riiablo is used only as a source for fixed panel-local facts; its mobile controls
and dynamic stage/aspect-ratio placement are not part of these desktop scenes.

Capture every registered frontend, lab, gameplay, and overlay scene in isolated
application runs with:

```sh
MPQ_DIRECTORY=/path/to/diablo-ii make capture-all
```

Artifacts are grouped under `captures/all-scenes/<scene>/`. Override
`CAPTURE_ALL_DIR` to choose another root, or
`CAPTURE_ALL_FIXTURE_CHARACTERS` to change the deterministic character fixture
count used by character-dependent scenes.
