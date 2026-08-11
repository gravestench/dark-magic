# Foundation systems research checkpoint

This short checkpoint file exists to make the first systems-research PR easy to navigate.

Read in this order:

1. [GAME_SYSTEMS_RESEARCH_PROGRAM.md](GAME_SYSTEMS_RESEARCH_PROGRAM.md)
2. [GAME_SYSTEMS_INDEX.md](GAME_SYSTEMS_INDEX.md)
3. [SYSTEMS_SOURCE_MATRIX.md](SYSTEMS_SOURCE_MATRIX.md)
4. [GAME_DATA_LOADING_AND_LINKAGE.md](GAME_DATA_LOADING_AND_LINKAGE.md)
5. [TIMING_RNG_AND_DETERMINISM.md](TIMING_RNG_AND_DETERMINISM.md)
6. [CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md](CHARACTER_PERSISTENCE_AND_SAVE_FORMAT.md)
7. [ITEM_STATS_AND_AFFIXES.md](ITEM_STATS_AND_AFFIXES.md)
8. [ITEM_SYSTEM.md](ITEM_SYSTEM.md)
9. [CHARACTER_PROGRESSION.md](CHARACTER_PROGRESSION.md)

The foundation docs are intentionally marked **baseline**, not **validated**. Each one ends with a numbered verification backlog. Those probes are the work that should turn uncertain compatibility details into implementation contracts.

The next research tranche should follow the index dependency order: combat/damage, skills/missiles/states, monsters/AI, hirelings/owned units, then movement/collision/pathfinding. Existing code work does not need to stop while that research is being filled in.
