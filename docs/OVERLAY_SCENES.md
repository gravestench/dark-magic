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
