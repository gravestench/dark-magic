# DS1 Map Editor: Original Feature-Parity Specification

This document is the implementation inventory for bringing Dark Magic's standalone
DS1 editor to feature parity with Paul Siramy's original `win_ds1edit` while keeping
the useful modern behavior already built here.

The priority is authoring capability, correctness, and responsiveness. Ornate editor
chrome is not a parity requirement. The current styled UI may be replaced with plain,
non-stylized boxes, native-resolution bitmap fonts, and conventional widgets whenever
that makes the remaining features faster or safer to implement.

## Sources and research boundary

This inventory was produced from the public behavior documentation, not from the
original editor implementation:

- [Diablo II MAP Editor documentation](http://paul.siramy.free.fr/_divers/ds1/doc/index.html)
- [Advanced map-editing tutorial](http://paul.siramy.free.fr/_divers/ds1/doc/tut01/index.html)
- [Official DS1 Editor download page](http://paul.siramy.free.fr/_divers/ds1/dl_ds1edit.html)
- The `README.txt` and `Ds1edit.ini` distributed in the official
  `win_ds1edit_20111030.zip` release, which the web documentation explicitly delegates
  command-line, configuration, multiple-map, object, and path details to
- Every screenshot and diagram referenced by the main manual and advanced tutorial,
  plus the overview screenshots on the download page

No source file from the original editor was used to derive this specification.

The web manual's last recorded feature revision is March 5, 2006. The official
download is newer, so its bundled README and configuration are authoritative where
they add behavior not covered by the web page. The tutorial is explicitly unfinished:
it fully demonstrates warp and roof workflows, but its promised map-resizing and
walkability sections stop before completion. The release README supplies the missing
resize behavior.

## Status legend

- **Done**: usable in the current standalone DS1 editor.
- **Partial**: some engine or UI support exists, but it does not meet original parity.
- **Missing**: no complete authoring workflow exists yet.
- **Modern**: useful Dark Magic behavior not present in the original editor.

## Product rules

These rules apply even where the original editor behaved differently:

- Mounted MPQs and extracted original game-data roots are always read-only.
- Save writes a copy beneath the configured mod/output root and cannot escape it,
  including through symlinks.
- Opening and saving an untouched document must preserve all supported DS1 data.
- Destructive operations must participate in undo/redo. Original behavior such as
  expert editing with no undo is not worth reproducing.
- Pixel-art and bitmap UI assets render at native resolution. Layout may reflow, but
  the renderer must not scale a fixed virtual screen to fit the window.
- A paint, selection, object, or path gesture changes only affected document records
  and retained render chunks. It must not rebuild or republish the whole map.
- Long decoding, cataloging, and preview operations are asynchronous and cancellable.
- Every unavailable or ambiguous external-data dependency produces a useful editor
  diagnostic rather than a crash.

## What the screenshots establish

The images clarify several behaviors that are easy to miss in the prose:

- A tile selection is a set of records, not merely a selected cell. Different visible
  records in one cell may be independently included.
- Committed selected tiles are blue; the area currently being dragged is orange.
- Copy/cut previews color safe or identical destinations green and overwritten
  destination records red.
- Floors, walls, shadows, special tiles, objects, and NPC paths can all be visible at
  once, with selection and diagnostic overlays composited above the map.
- Walkability is shown on the 5 by 5 DT1 subtile diamond, not as one flag per DS1 cell.
  The complete mode distinguishes multiple collision meanings and includes a legend.
- Invisible or unresolved records remain selectable through a numbered/outlined
  placeholder instead of disappearing from the authoring view.
- Object authoring displays the sprite's feet separately from a movable editor-only
  label. Labels expose description, object index, type/ID, and path count.
- The tile chooser is a large paged matrix with layer jump buttons, category tabs,
  a composite preview, logical identity annotations, and DT1-source annotations.
- The path editor displays a table of path index, X, Y, and Action alongside editable
  polylines and an explicit OK/Cancel transaction.
- The expert tile editor is a layer-by-bit matrix. It annotates Hidden, Unwalkable,
  main index, sub-index, orientation, and partially understood priority/type fields.

## Current Dark Magic baseline

### Done

- A standalone `tools/ds1-editor` composition that does not start the game or the
  `d2legacy` Lua package.
- A private `ds1editor.*` Lua namespace and narrow `engine.map_editor/v1` capability.
- Read-only mounted assets and confined, atomic mod-overlay saves.
- Opening an existing DS1 through the shared codec while retaining untouched data.
- DT1 dependency resolution, logical tile identities, physical variant enumeration,
  thumbnails, metadata, and palette-aware previews.
- Floor, wall, and shadow point painting/erasing on the first layer of each family.
- Pan, right-drag pan, temporary Space pan, arrow-key pan, pointer-centered wheel zoom,
  fit/centering, native window resolution, and responsive relayout.
- Floor, wall, shadow, debug-grid, and collision visibility controls.
- Eyedropper sampling, a selected-record inspector, and editing of orientation,
  style, sequence, Prop1, two packed auxiliary fields, and Hidden.
- Coalesced gesture history with undo and redo.
- Incremental world-cell replacement and retained visible render chunks.
- Palette inference from the map act, Act I fallback, and explicit Act I-V override.

### Modern additions to preserve

- An opt-in adjacency model that learns conservative tile transitions from the open
  DS1 and can choose physical variants while painting.
- Explicit DT1 tileset and physical-variant browsing.
- A selected-cell composite preview with independent floor/wall/shadow/collision
  visibility.
- Redo, which the original editor did not provide.
- Safe save-as-copy semantics instead of renaming or overwriting the opened source.
- Native-resolution responsive layout rather than an 800 by 600 fixed display.

### Partial or missing foundations

- The Go capability can create a modern v18 document, but the Lua UI has no complete
  creation workflow.
- Only layer index zero is paintable from the Lua UI. A two-floor/four-wall DS1 is not
  yet fully editable.
- History mutations cover floor/wall/shadow records, not objects, paths,
  substitutions, dependencies, dimensions, or header fields.
- The current selected-record controls expose packed bytes, not the original
  per-layer expert bitfield matrix or multi-selection mixed values.
- New-document support exists below the UI but lacks a project-quality workflow.
- Incremental rendering exists, but large-map load and edit latency still require
  profiling and stricter invalidation boundaries.

## Complete original-editor parity backlog

### 1. Project, launch, and data-source workflows

- [ ] Open one DS1 by mounted path, host path, recent-file entry, or file picker.
- [ ] Open a multi-map project containing 1 to 100 DS1 entries.
- [ ] Give every project entry a DS1 path, `LvlTypes.txt` ID, and `LvlPrest.txt` DEF.
- [ ] Permit DEF `-1` and detect the `LvlPrest.txt` row from a unique filename.
- [ ] Diagnose ambiguous or missing automatic DEF matches.
- [ ] Switch directly among maps 1-10 in the active set.
- [ ] Switch among ten sets so maps 1-100 remain quickly addressable.
- [ ] Show/hide a project row containing current set, map index, and full path.
- [ ] Preserve the clipboard while switching maps so compatible cross-map paste works.
- [ ] Warn before cross-map paste when LvlType ID, LvlPrest DEF, or effective DT1
      catalogs differ.
- [ ] Import/export the original whitespace project-list format for interoperability.
- [ ] Provide a modern project file that can also store palette, viewport, open tabs,
      visibility, and editor-only preferences without altering DS1 files.
- [ ] Load loose mod-overlay files before archive files, matching Diablo II's
      `-direct -txt` precedence.
- [ ] Resolve fallback assets in patch, expansion, and base-game order.
- [ ] Offer a forced-DT1 mode with 1-32 explicit DT1 files for maps not represented in
      `LvlPrest.txt`.
- [ ] Prevent simultaneous forced-DT1 and LvlType/LvlPrest selection.
- [ ] Offer `force palette` Act I-V independently from a map's internal act.
- [ ] Offer an explicit act-mismatch override with a warning.
- [ ] Support a `no Vis debug` presentation option for clean screenshots.
- [ ] Export a dependency/debug report for loaded TXT rows, DS1 fields, and DT1 files.
- [ ] Persist decoded palettes and other safe derived caches so repeated launches do
      not repeat expensive color-table or catalog work.
- [ ] Restore last project, recent maps, output root, and non-document preferences.
- [ ] Investigate the release configuration's undocumented `workspace_enable` option;
      do not claim equivalent behavior until its user-visible semantics are known.

### 2. DS1 format and document lifecycle

- [ ] Read every DS1 version supported by the codec and identify its version in UI.
- [ ] Preserve version-specific unknown header and substitution fields losslessly.
- [ ] Decide per save whether to retain the source version or explicitly convert to
      v18; never perform an invisible format conversion.
- [ ] Expose width, height, act, substitution type, layer counts, dependency strings,
      object count, path count, and substitution-group count.
- [ ] Create a new v18 DS1 with dimensions, act, floor/wall layer counts, substitution
      mode, DT1 dependencies, LvlType ID, and optional LvlPrest DEF.
- [ ] Resize a single map by DS1 tile dimensions.
- [ ] Resize every map in a multi-map project only after explicit confirmation.
- [ ] When growing, append correctly initialized empty records on every declared layer.
- [ ] When shrinking, preview records, objects, and paths that will fall out of bounds.
- [ ] Let the user cancel, crop, or relocate affected objects/paths before commit.
- [ ] Explain the game's external size convention: `Levels.txt`/`LvlPrest.txt` values
      are DS1 dimensions minus one.
- [ ] Add/remove floor layers up to two and wall layers up to four where the target
      format supports them.
- [ ] Offer the original compatibility conversion that expands maps to two floor and
      four wall layers, but make it opt-in rather than automatic.
- [ ] Offer compact/minimized serialization as an explicit save option.
- [ ] Track dirty state per map and for the project as a whole.
- [ ] Confirm close/quit when one or more maps contain unsaved edits.
- [ ] Save one map, save all maps, and save as a new output-relative path.
- [ ] Keep the current read-only-source guarantee instead of the original editor's
      in-place save behavior.
- [ ] Add numbered backup snapshots inside the output root as an optional policy.
- [ ] Show the exact destination before saving and reveal it afterward.
- [ ] Validate the encoded result before atomically replacing an output copy.
- [ ] Provide document-scoped undo/redo from open to close, bounded by a configurable
      memory budget rather than temporary files.

### 3. Viewport, status, and general presentation

- [ ] Make Tile, Object, and Path explicit editing modes with distinct input routing.
- [ ] Always show current mode, map name, map index, cell/subtile coordinate, zoom,
      palette/gamma, and dirty state.
- [ ] Show the authored properties under the pointer in a compact status inspector.
- [ ] Pan with arrow keys, primary drag in Pan mode, secondary drag, and temporary
      Space drag.
- [ ] Optionally auto-scroll while the pointer rests at a viewport edge.
- [ ] Configure keyboard and edge-scroll speed independently.
- [ ] Keep pan and zoom as retained transforms; neither may rasterize the map again.
- [ ] Support deterministic zoom steps equivalent to 1:1, 1:2, 1:4, 1:8, and 1:16.
- [ ] Keep smooth pointer-centered wheel zoom as a modern option.
- [ ] Center on the record/subtile under the pointer.
- [ ] Configure the zoom applied by the center command.
- [ ] Recenter on the whole map with Home/Fit.
- [ ] Adjust display gamma upward and downward without altering source assets.
- [ ] Store the preferred gamma per project or palette.
- [ ] Capture the visible viewport without overwriting an existing screenshot; the
      original used numbered PCX files, while PNG should be the modern default.
- [ ] Capture the entire map at the current zoom and visibility state.
- [ ] Export full-map PNG in addition to an original-compatible bitmap option.
- [ ] Provide two night-preview qualities plus normal presentation.
- [ ] Make night preview visibly non-authoritative: it does not model ambient level,
      light radius, or full in-game lighting.

### 4. Layer visibility and diagnostic overlays

- [ ] Independently toggle Floor 1 and Floor 2.
- [ ] Independently toggle Wall 1 through Wall 4.
- [ ] Cycle shadows through off, opaque/as-is, white diagnostic, and translucent.
- [ ] Toggle NPC path overlays.
- [ ] Toggle special-tile priority between ordinary wall depth and top/roof depth.
- [ ] Show only one requested floor, wall, or shadow layer.
- [ ] Show every ordinary layer except one requested layer.
- [ ] Reset all layer visibility and display modes to defaults.
- [ ] Keep selection semantics based on visible record layers, not merely whole cells.
- [ ] Cycle the tile grid through off, above floors/below walls, and above all tiles.
- [ ] Cycle grid states forward and backward.
- [ ] Cycle collision display through off, simple, and complete.
- [ ] In simple collision mode, distinguish walking and jumping restrictions.
- [ ] In complete collision mode, expose all decoded DT1 subtile flags with a legend.
- [ ] Allow the collision legend to be shown or hidden independently.
- [ ] Overlay collisions on the exact 5 by 5 subtile diamond for each physical DT1
      record, including wall offsets and composition depth.
- [ ] Render unresolved and Hidden records as selectable diagnostic placeholders.
- [ ] Let the inspector reveal the original logical/raw record behind any diagnostic
      placeholder, matching the original right-click inspection behavior.
- [ ] Distinguish a missing DT1 identity from a deliberately Hidden DS1 record.
- [ ] Show Vis numbers/orientations and special-tile roles when Vis diagnostics are on.
- [ ] Keep diagnostic overlays out of saved DS1 data.

### 5. Tile selection model

- [ ] Represent selection as `(cell, record family, layer)` entries.
- [ ] Rectangle-drag to replace the current selection.
- [ ] Shift+rectangle to add visible records to the selection.
- [ ] Ctrl+rectangle to subtract visible records from the selection.
- [ ] Hide visible records in the dragged area without deleting them.
- [ ] Restore all editor-hidden records with one command.
- [ ] Permit visibility changes while an area drag is active.
- [ ] Draw committed selection, active area, hover, and clipboard preview as separate
      overlay lifetimes so none can become stale.
- [ ] Select every record on the map identical to visible records beneath the pointer.
- [ ] Add identical records to an existing selection.
- [ ] Subtract identical records from an existing selection.
- [ ] Search all layers and editor-hidden records after using only visible records at
      the clicked cell to define the identities.
- [ ] Select one floor/wall/shadow record in a stacked cell without forcing the other
      records in that cell into the selection.
- [ ] Expose select all, select none, invert selection, and selection count as modern
      conveniences.
- [ ] Keep selection changes out of document undo history unless they mutate records.

### 6. Tile clipboard, movement, deletion, and transforms

- [ ] Copy the selected records while preserving family, layer, relative offset, and
      every packed/unknown source field.
- [ ] Cut/move the selected records as one transaction.
- [ ] Follow the pointer with a native-resolution paste preview.
- [ ] Color destinations green when empty or identical.
- [ ] Color destinations red when a record will be overwritten.
- [ ] Explain each conflict in a preview inspector before drop.
- [ ] Cancel a pending copy/cut without changing the document.
- [ ] Paste repeatedly without rebuilding the clipboard.
- [ ] Paste between compatible maps in a multi-map project.
- [ ] Warn, remap explicitly, or refuse paste between incompatible catalogs; never
      silently produce distorted identities.
- [ ] Delete selected records across multiple cells, families, and layers.
- [ ] Make copy, cut, paste, delete, and bulk metadata edits single undoable commands.
- [ ] Add optional rotate/mirror operations only after DS1 orientation remapping rules
      are explicit and validated.

### 7. Tile library and ordinary tile editing

- [ ] Choose the active authored layer: Floor 1-2, Wall 1-4, or Shadow.
- [ ] Jump the tile chooser to the current record when the selected layer is non-empty.
- [ ] Otherwise open at the beginning of that layer/category.
- [ ] Show the Shadow layer only when compatible shadow records exist.
- [ ] Browse logical identity `(orientation/type, main/style, sub/sequence)`.
- [ ] Browse every physical DT1 variant for the logical identity.
- [ ] Group or filter tiles by floor, lower wall, upper wall/roof, special, shadow,
      animation, and other decoded DT1 categories.
- [ ] Show a composite preview beside the tile matrix.
- [ ] Annotate the DT1 source at the start of each source group.
- [ ] Support row, page-vertical, page-horizontal, and Home navigation.
- [ ] Search by logical numbers, DT1 filename, source index, category, and special-tile
      name.
- [ ] Place or replace the active record without changing unrelated records in the
      cell.
- [ ] Paint a chosen tile across a drag as one history entry.
- [ ] Erase only the active record family/layer.
- [ ] Preserve all physical variants and expose deterministic/randomized variant
      strategies.
- [ ] Preserve the modern auto-draw toggle, preview all cells it may affect, and make
      the affected set one atomic undoable mutation.
- [ ] Never edit or save a mounted DT1 from this tool.

### 8. Expert DS1 tile-field editing

- [ ] Provide a raw 32-bit and four-byte view for every selected tile record.
- [ ] Display one row per present layer: F1, F2, Shadow, W1-W4, and Substitution.
- [ ] Label known fields: Hidden, Unwalkable, main/style index, sub/sequence index,
      wall orientation, and known priority/type bits.
- [ ] Preserve and label unknown bits rather than hiding or normalizing them.
- [ ] For a single cell with no selection, edit all present layer records at that cell.
- [ ] For a multi-record selection, scope editing to selected records only.
- [ ] Display a uniform bit value distinctly from a mixed bit value.
- [ ] Mark a bit unavailable when the selection contains no record on that layer.
- [ ] Leave mixed values unchanged unless the user explicitly assigns zero or one.
- [ ] Apply a bulk bit assignment to every applicable selected record.
- [ ] Make expert edits undoable despite the original editor's limitation.
- [ ] Explain that the DS1 Unwalkable bit blocks the whole tile independently of DT1
      subtile collision flags.
- [ ] Provide safe presets for Hidden Vis, whole-tile Unwalkable, and clearing those
      bits while showing the exact underlying changes.

### 9. Object rendering and inspection

- [ ] Decode and render every DS1 object record without altering its type, ID,
      coordinates, flags, or serialized path association.
- [ ] Load object descriptions and presentation data from the effective loose/archive
      data set.
- [ ] Resolve Type 1 monsters/NPCs and Type 2 objects using the map act and configured
      object table sizes.
- [ ] Draw object animations through the normal Dark Magic sprite pipeline.
- [ ] Cycle animation presentation through hidden, frozen, and animated.
- [ ] Cycle object annotations through hidden, type/ID, animation speed, and name.
- [ ] Reload changed object-description data without reopening the DS1.
- [ ] Draw an editor-only label separately from the object's feet.
- [ ] Include description, object index, type/ID, and path count in the label.
- [ ] Allow a label to be moved temporarily to reveal the sprite beneath it without
      serializing the label position.
- [ ] Keep labels above walls so every object can be found in Object mode.
- [ ] Distinguish selectable and non-selectable in-game objects in the inspector.
- [ ] Display object coordinates in subtiles, direction when resolvable, and raw flags.

### 10. Object authoring

- [ ] Insert a new object at the pointer.
- [ ] Right-click an object or label to edit its Type and ID.
- [ ] Browse the object catalog by act, type, ID, name, animation, and direction.
- [ ] Permit Type 1 objects across acts as the original did.
- [ ] Constrain Type 2 objects to valid acts by default, with a project-level override
      matching `only_normal_type2`.
- [ ] Select one object or one label by click.
- [ ] Shift+click to add and Ctrl+click to remove.
- [ ] Prevent a mixed selection of labels and objects.
- [ ] Select all identical objects or labels under the pointer.
- [ ] Move selected objects with a preview and explicit drop (the original used
      Alt/AltGr plus primary click).
- [ ] Copy selected objects.
- [ ] Delete selected objects.
- [ ] Preserve or deliberately transform object paths during move/copy.
- [ ] Make insert, type/ID edit, move, copy, delete, and path association fully
      undoable and redoable.
- [ ] Validate map bounds and reserved terminal row/column concerns.
- [ ] Warn when two objects share coordinates and at least one owns paths, because the
      original workflow identifies this as unsafe.

### 11. NPC path authoring

- [ ] Enter Path mode only with exactly one object selected.
- [ ] Permit inspection for Type 1 and Type 2 records while warning that only Type 1
      monsters/NPCs use DS1 paths in the game.
- [ ] Draw the selected object's ordered path as points and connected lines.
- [ ] Draw other objects' paths as an optional context overlay.
- [ ] Display path index, X, Y, and Action in a table.
- [ ] Display the current default Action while editing.
- [ ] Start an `All New` transaction that replaces the current path only on commit.
- [ ] Add path points by clicking subtiles on the map.
- [ ] Set the default numeric Action for subsequently added points.
- [ ] Select and edit an existing point's position and Action.
- [ ] Insert, reorder, and delete individual points.
- [ ] Stop point placement without leaving the transaction.
- [ ] Commit with OK and restore the exact original path with Cancel.
- [ ] Make the committed path edit undoable/redoable in normal document history.
- [ ] Validate out-of-bounds path points, especially after resize.
- [ ] Preserve `NPCPathOrder` and object-to-path association through save.

### 12. Substitution layers and groups

- [ ] Display substitution records as their own layer when present.
- [ ] Include substitution rows in the expert field editor.
- [ ] Inspect and edit substitution type.
- [ ] Inspect and edit substitution-group rectangle X, Y, width, and height.
- [ ] Preserve/edit each group's extra unknown value without inventing semantics.
- [ ] Create, select, move, resize, duplicate, and delete substitution groups.
- [ ] Validate groups against document bounds.
- [ ] Include substitution mutations in undo/redo and incremental invalidation.
- [ ] Document which operations remain raw because the original manual does not define
      their higher-level game semantics.

### 13. Special tiles, Vis, warp, and roof workflows

- [ ] Identify special tiles and display their orientation/Vis number and role.
- [ ] Keep special tiles selectable even when they have no graphics.
- [ ] Warn before deleting special tiles used as entry, corpse, town portal, warp, or
      roof/pop-area markers.
- [ ] Offer a Hidden-Vis action that sets the DS1 Hidden bit on the correct wall layer.
- [ ] Warn when a graphic-less Vis is visible and would become a green tile in game.
- [ ] Warn when a Vis is too close to the terminal row/column or another unsafe border.
- [ ] Inventory Vis0-Vis7 used by the open DS1.
- [ ] Cross-check every used Vis against effective `Levels.txt` Vis and Warp columns.
- [ ] Validate that linked levels reference one another where required.
- [ ] Validate that every used Vis has a corresponding Warp definition.
- [ ] Inspect effective `LvlWarp.txt` direction, selection rectangle, base offset,
      exit-walk offset, LitVersion, and tile-count data.
- [ ] Overlay a warp's subtile base, activation rectangle, and exit-walk destination as
      demonstrated in the tutorial diagrams.
- [ ] Support the documented one-way-warp composition while clearly separating the
      DS1 special tile/object edit from external TXT changes.
- [ ] Detect roof/pop-area special tiles when known.
- [ ] Warn that functioning disappearing roofs also depend on compatible model tiles,
      special markers, and `LvlPrest.txt` Pops/PopPad values.

### 14. DT1 and external table integration

- [ ] Show the effective `LvlTypes.txt` row and its File1-File32 DT1 candidates.
- [ ] Show the effective `LvlPrest.txt` row, map file slot, dimensions, Dt1Mask, Pops,
      PopPad, and DEF.
- [ ] Explain which DT1 files are enabled by the 32-bit Dt1Mask.
- [ ] Provide a visual Dt1Mask editor/calculator for the writable mod overlay.
- [ ] Detect when a pasted logical identity cannot resolve in the destination catalog.
- [ ] Detect when copied roof/house tiles need an additional LvlTypes file and Dt1Mask
      bit, as demonstrated by the tutorial.
- [ ] Load effective `Levels.txt`, `LvlPrest.txt`, `LvlTypes.txt`, `LvlWarp.txt`,
      `Objects.txt`, and `ObjType.txt` through layered VFS precedence.
- [ ] Initially keep external TXT inspectors read-only; later write changes only to the
      configured mod overlay with explicit confirmation and independent history.
- [ ] Never patch TXT files inside MPQs.
- [ ] Refresh table-derived catalogs after a loose overlay file changes.

### 15. Validation and diagnostics

- [ ] Validate DS1 dimensions, layer counts, record counts, and format-version rules.
- [ ] Validate every logical tile identity against the effective DT1 catalog.
- [ ] Distinguish missing identity, missing DT1 file, corrupt DT1, palette failure, and
      unsupported orientation.
- [ ] Validate object type/ID ranges under current act/table configuration.
- [ ] Validate object/path ownership and duplicate-coordinate hazards.
- [ ] Validate special Vis and Warp table consistency.
- [ ] Validate cross-map clipboard compatibility before mutation.
- [ ] Validate resize data loss before mutation.
- [ ] Report unknown bits/fields without rejecting otherwise lossless documents.
- [ ] Provide a map-wide Problems panel with severity, cell/subtile coordinate, layer,
      object/path reference, explanation, and jump-to action.
- [ ] Export a plain-text/JSON diagnostic report suitable for bug reports.
- [ ] Keep validation off the frame loop; update only affected diagnostics after edits.

### 16. Input and discoverability

- [ ] Preserve conventional buttons and menus for every command; parity must not depend
      on memorizing the original keyboard-only interface.
- [ ] Provide a searchable command palette.
- [ ] Provide remappable shortcuts and a shortcut reference.
- [ ] Implement original-equivalent shortcuts where they do not conflict with desktop
      conventions: layer toggles, show-only/exclude, grid, collisions, paths, shadows,
      gamma, center, fit, screenshot, identical select, copy/cut/delete, and save.
- [ ] Show tooltips with action, shortcut, current state, and destructive consequence.
- [ ] Keep the Pan tool first and default.
- [ ] Make Space temporary pan from every non-modal authoring tool.
- [ ] Make Escape cancel the active drag, clipboard preview, modal, or transaction
      before it requests application close.
- [ ] Keep hover, selected, active-tool, disabled, and keyboard-focus states visually
      distinct without changing widget dimensions.
- [ ] Ensure every list and form is operable by mouse and keyboard.

### 17. Simple UI composition for the parity push

- [ ] Replace the current ornate shell with a small reusable widget layer if it slows
      feature work or obscures layout defects.
- [ ] Use opaque/flat boxes, one-pixel borders, bitmap-font labels, and a restrained
      semantic color palette.
- [ ] Implement measured text, button padding, checkbox, radio button, dropdown,
      tab, list, tree, numeric field, modal, tooltip, splitter, scrollbar, and status
      widgets before building more one-off composition code.
- [ ] Use a dockable/resizable layout: toolbar, tool rail, map viewport, inspector,
      Problems/console, and status bar.
- [ ] Let the inspector collapse or move below the map on narrow windows.
- [ ] Virtualize DT1 tiles, object catalogs, multi-map lists, and problem lists.
- [ ] Clip every list and preview to its owning panel using local coordinates.
- [ ] Keep all idle/hover/pressed/selected widget states the same measured dimensions.
- [ ] Remove or ignore generated ornate UI assets once no live composition references
      them; do not maintain two parallel widget systems.
- [ ] Revisit thematic pixel-art chrome only after functional parity and performance
      targets are met.

### 18. Performance acceptance criteria

- [ ] Opening a large town DS1 must keep the window responsive and show staged progress.
- [ ] Decode, dependency resolution, DT1 indexing, and initial preview must be separately
      timed so regressions have an owner.
- [ ] Reuse decoded palettes, DT1 catalogs, physical tile textures, and object animation
      data across maps that share them.
- [ ] Cull map chunks, objects, paths, labels, collisions, grids, and selection overlays
      outside the viewport.
- [ ] Panning within the same visible chunk range performs no content work.
- [ ] Zoom changes retained transforms and visibility, not DS1 materialization.
- [ ] A point paint replaces one authored world cell and invalidates only intersecting
      visual/collision/diagnostic chunks.
- [ ] A brush drag publishes affected chunks at most once per frame and once at commit.
- [ ] Selection-only changes never rematerialize the world.
- [ ] Object-only edits never rerasterize unchanged tile chunks.
- [ ] Path-only edits never rerasterize tile or object imagery.
- [ ] Visibility changes invalidate only the affected presentation layers.
- [ ] No map-sized pixel buffer is uploaded on an ordinary paint stroke.
- [ ] Background jobs carry generation IDs/cancellation so stale results cannot replace
      a newer document or palette.
- [ ] Profile whole-town load, first frame, pan, zoom, paint, collision toggle, object
      animation, and full-map screenshot as separate scenarios.
- [ ] Add performance counters to a diagnostics overlay: frame time, pending jobs,
      visible/dirty chunks, uploads, decoded DT1s, object sprites, and memory.

## Recommended implementation sequence

### Phase 0: simplify and make the renderer trustworthy

1. Replace the ornate composition with simple measured widgets and a dockable layout.
2. Add timing/counter instrumentation around decode, catalog, materialization,
   rasterization, upload, and Lua update.
3. Make paint update the authoritative source record and only the affected retained
   tile/collision chunks, with no map-wide preload or snapshot replacement.
4. Virtualize the physical tile library and remove per-item work outside visible rows.
5. Establish stable overlay lifetimes for hover, selected, drag area, and clipboard.

Exit criterion: large town maps pan and paint smoothly, and profiling proves an
ordinary point paint causes no unrelated tile rasterization or upload.

### Phase 1: complete tile records and selection

1. Generalize document mutations from one record to batch mutations across all two
   floor, four wall, shadow, and substitution layers.
2. Implement record-level rectangle and identical-tile selection.
3. Implement copy, cut, paste preview/conflicts, delete, hide/show, and bulk history.
4. Implement the complete tile picker and expert mixed-value bitfield editor.
5. Complete simple/full 5 by 5 collision and grid presentation modes.

Exit criterion: every tile-editing example in the original manual can be reproduced,
saved, undone, redone, and reopened without changing unrelated DS1 data.

### Phase 2: projects, multiple maps, creation, and resize

1. Add project state and 1-100 open-map navigation.
2. Add LvlType/LvlPrest selection and forced-DT1 workflows.
3. Add new-document and loss-previewed resize transactions.
4. Add compatible cross-map clipboard and optional output-root backups.

Exit criterion: the cottage-to-town and large maze workflows can be performed without
restarting the editor or writing to original assets.

### Phase 3: objects

1. Add editable object snapshots and batch history to the Go document model.
2. Add layered object resolution, animation, labels, and info modes.
3. Add insert, catalog/type edit, select, identical select, move, copy, and delete.
4. Add object validation and table refresh.

Exit criterion: every object workflow in the release README can be performed with
multi-step undo/redo.

### Phase 4: paths and substitutions

1. Add transactional path editing with point/action table and visual handles.
2. Preserve serialized path order and validate ownership/bounds.
3. Add substitution record and group editing with raw-field preservation.

Exit criterion: every documented path-editing screenshot can be reproduced, and a
saved/reopened document retains identical object/path associations.

### Phase 5: special-tile and external-data workflows

1. Add Vis/special-tile roles and safety diagnostics.
2. Add read-only inspectors for Levels, LvlWarp, LvlTypes, LvlPrest, Objects, and
   ObjType effective rows.
3. Add warp geometry overlays, link validation, DT1-mask assistance, and roof workflow
   diagnostics.
4. Only then consider explicit overlay-safe editing of external TXT files.

Exit criterion: the tutorial's trap-door, one-way tent warp, and disappearing cottage
workflows are fully explained and validated by the editor, with every external edit
written only to the mod overlay.

## Definition of original-editor parity

Parity is reached when all of the following are true:

- Every command and editing workflow documented by the main manual and release README
  has an equivalent discoverable UI action.
- The screenshots' selection, conflict, collision, object-label, tile-browser,
  bitfield, and path-editing information is available with equal or greater clarity.
- The advanced tutorial's DS1-side trap-door, Vis, tent-warp, and cottage/roof edits can
  be performed, and the required external table changes are inspected or validated.
- One project can safely operate on up to 100 DS1 files and copy compatible records
  between them.
- All supported layers, objects, paths, substitutions, special records, unknown bits,
  and version-specific fields survive an untouched open/save cycle.
- All authoring mutations are undoable/redoable and all saves remain confined to the
  configured output root.
- Large maps remain interactive during load and do not perform full-map work for local
  edits.

Visual theming is deliberately excluded from this definition. Once parity is stable,
the simple widget layer can receive a coherent pixel-art skin without changing editor
layout, input semantics, or document behavior.
