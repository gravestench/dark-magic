# Riiablo recovered game data

The four tables in this directory are preserved from the Apache-2.0-licensed
[Riiablo](https://github.com/collinsmith/riiablo) project. They document game
relationships that were originally compiled into Diablo II rather than shipped
as ordinary Excel tables.

Dark Magic keeps these source rows intact for provenance. Go code parses and
validates them into immutable records; Lua consumes only that normalized view.
The speech table supplies the join between a logical `Sounds.txt` identifier and
a localization string key. It does not contain Blizzard audio or localized text.
`ds1types.txt` names DS1 preset definitions and their level types; `obj.txt`
maps act-local DS1 object IDs to global `Objects.txt` IDs.

Source paths in the Riiablo repository:

- `assets/data/quests.txt`
- `assets/data/speech.txt`
- `assets/data/ds1types.txt`
- `assets/data/obj.txt`

Imported 2026-08-06. Modifications: none.
