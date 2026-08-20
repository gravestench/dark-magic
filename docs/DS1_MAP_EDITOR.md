# DS1 Map Editor

The Map Editor is a standalone authoring tool for opening an existing DS1, editing its
floor, wall, and shadow source records, previewing the unsaved result with the
normal Dark Magic world pipeline, and safely saving a mod overlay copy.

## Run it

Mount Diablo II assets as usual, choose a writable directory for your mod
overlay, then start the tool:

```sh
MPQ_DIRECTORY="~/d2_english_mpq" \
go run ./tools/ds1-editor \
  --mods none \
  --window-size 1600x1000 \
  --open data/global/tiles/act1/town/townE1.ds1 \
  --output "~/my-d2-mod"
```

`--open` is optional; it opens a mounted DS1 immediately and is useful for
repeatable profiling or for a project-specific editor shortcut.

`--output` defaults to a user configuration directory and is the only place the
tool may save. The mounted game
data is read-only; saving `data/global/tiles/act1/foo.ds1` writes only
`~/my-d2-mod/data/global/tiles/act1/foo.ds1`. The final write is atomic.
The editor also rejects an output root or destination beneath a directory
listed in `MPQ_DIRECTORY`, so unpacked original assets cannot be overwritten
by mistake.

The tool always uses native resolution: its render surface matches the current
drawable window exactly. It does not upscale a fixed 800×600 renderer;
resizing rebuilds the editor chrome around the active unsaved document.

The application is its own composition and content package. It starts the
generic Lua capability host, but registers only display, data, VFS, input,
render, scene, and map-editor modules. It neither installs nor executes the
`d2legacy` Lua package. Editor Lua lives under the private `ds1editor.*`
namespace, while its generated native-pixel DC6 component atlas and palette are
exported only beneath `darkmagic/ds1-editor/` in the layered VFS. Diablo II
archives remain a read-only data/font/tile dependency.

Run `make ds1-editor-assets` to rebuild the embedded DC6 sheets and frame
manifest from the generated source sheets in `tools/ds1-editor/assets/source`.

## Everyday workflow

1. Press `F` or select **OPEN**, then search for a DS1.
2. Choose **FLOOR**, **WALL**, or **SHADOW**. The **LIBRARY** panel groups the
   concrete tile thumbnails by every DT1 declared by that DS1; scroll either
   list to browse it.
3. Use **PAINT** to draw, **PICK** to sample a map cell, **ERASE** to clear the
   active layer, and **PAN** (or right-drag) to navigate. Hold `Space` while
   primary-dragging for temporary pan; the arrow keys pan without changing the
   selected tool.
4. Scroll to zoom around the pointer. `Home` and `End` make small zoom
   adjustments.
5. Save with **SAVE** or `Ctrl+S`. Use `Ctrl+Z` and `Ctrl+Y` to undo and redo
whole gestures.

**PICK** resolves the same rarity-selected physical DT1 record used by the map
renderer. It selects that DT1 and thumbnail in **LIBRARY**, outlines the cell in
a layer-specific color, and opens **SELECTED**. That view edits the authored DS1
type/style/sequence, Prop1 byte, and hidden bit with undo. It also shows the
resolved DT1 record, dimensions, rarity, roof/direction metadata, and aggregate
subtile collision flags. DT1 metadata itself remains inspect-only because the
tool does not write mounted tile libraries.

The first preview is built asynchronously. A complete earlier preview remains
visible while a changed DS1 is materialized. Completed strokes report their
dirty DS1 cells; only intersecting layer/depth chunks are discarded and
rasterized again, while every unaffected retained texture remains in place.

## What this version intentionally protects

- Opened DS1 fields are retained in their codec model; painting changes only
  the selected source record family and layer.
- A brush stroke is one undo entry, including a drag that revisits a cell.
- Tiles are selected by logical `(orientation, style, sequence)`, not a DT1
  record offset. This keeps maps correct when a matching DT1 has variants.
- Object and substitution records remain untouched. Their full authoring UI,
  custom map creation, and path/object tools are the next editor milestones.

## Autotiling status

**AUTO** is opt-in and off by default. When enabled, the editor learns directed
neighbor relationships already present in the open DS1. It considers only
logical identities supplied by the selected DT1, prefers the explicitly picked
tile on ties, and may adapt the painted cell plus immediately adjacent cells
that already belong to that same tileset family. This gives conservative edge
continuation without rewriting unrelated terrain. The learned graph is a local
authoring aid, not proof of semantic compatibility; transition, roof, and
collision-sensitive work should still be reviewed in **SELECTED**.
