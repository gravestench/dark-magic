package entryworld

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// Build generates and materializes the canonical Act I entry topology. It resolves both seams and the town spawn
// before publication so callers never receive a partially usable authority world.
func Build(
	ctx context.Context,
	content fs.FS,
	d2legacySource fs.FS,
	records Records,
	resolver gameworld.ObjectResolver,
	seed uint64,
	difficulty int,
) (*Prepared, error) {
	if err := validateDependencies(content, d2legacySource, records, resolver); err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	generated, err := d2mapgen.GenerateEntryWorld(ctx, d2legacySource, records, seed, difficulty)
	if err != nil {
		return nil, err
	}
	return prepareTopology(ctx, content, d2legacySource, records, resolver, generated)
}

// RestorePrepared rehydrates a saved topology without asking Lua to select new
// zones. Materialization and collision resolution still run before publication.
func RestorePrepared(
	ctx context.Context,
	content fs.FS,
	d2legacySource fs.FS,
	records Records,
	resolver gameworld.ObjectResolver,
	snapshot []byte,
) (*Prepared, error) {
	if err := validateDependencies(content, d2legacySource, records, resolver); err != nil {
		return nil, err
	}
	generated, err := d2mapgen.RestoreEntryWorld(snapshot)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return prepareTopology(ctx, content, d2legacySource, records, resolver, generated)
}

// prepareTopology is shared by fresh generation and snapshot restore so both paths reject missing maps and unsafe
// seams before publishing a playable world.
func prepareTopology(
	ctx context.Context,
	content fs.FS,
	d2legacySource fs.FS,
	records Records,
	resolver gameworld.ObjectResolver,
	generated d2mapgen.EntryWorld,
) (*Prepared, error) {
	worlds := make(map[int]*gameworld.Map, len(generated.Zones))
	for level := 1; level <= len(generated.Zones); level++ {
		zone := generated.Zones[level]
		if zone == nil {
			return nil, fmt.Errorf("d2legacy entry world: missing zone %d", level)
		}
		gameMap, err := materialize(ctx, content, zone, resolver)
		if err != nil {
			return nil, fmt.Errorf("materialize entry level %d: %w", level, err)
		}
		worlds[level] = gameMap
	}
	seams := make([]gametransition.Seam, 0, len(generated.Seams))
	for _, spec := range generated.Seams {
		seam, err := gametransition.ResolveSeam(spec, worlds[spec.FirstLevel], worlds[spec.SecondLevel])
		if err != nil {
			return nil, fmt.Errorf("join entry levels %d and %d: %w", spec.FirstLevel, spec.SecondLevel, err)
		}
		seams = append(seams, seam)
	}
	if len(seams) == 0 {
		return nil, errors.New("d2legacy entry world has no seams")
	}
	townX, townY, found := d2mapgen.ResolveTownEntry(ctx, d2legacySource, records, worlds[1])
	if !found {
		return nil, errors.New("d2legacy entry world: Act I town has no campfire entry")
	}
	spawns := map[int][2]float64{1: {townX, townY}}
	for _, seam := range seams {
		spawns[seam.Second.LevelID] = [2]float64{seam.Second.ArrivalX, seam.Second.ArrivalY}
	}
	return &Prepared{
		Worlds: worlds, Zones: generated.Zones, Spawns: spawns,
		Seams: seams, Topology: generated,
		Difficulty: int(generated.Zones[1].Request().Difficulty),
	}, nil
}

// validateDependencies rejects incomplete composition before generation starts. The single check intentionally keeps
// the public compatibility error unchanged while preventing nil boundaries from failing much deeper in map loading.
func validateDependencies(
	content fs.FS,
	d2legacySource fs.FS,
	records Records,
	resolver gameworld.ObjectResolver,
) error {
	if content == nil || d2legacySource == nil || records == nil || resolver == nil {
		return errors.New("d2legacy entry world: content, source, records, and object resolver are required")
	}

	return nil
}

// materialize drives the incremental world loader to completion. Checking both its sentinel and progress preserves
// compatibility with materializers that report completion through either supported mechanism.
func materialize(
	ctx context.Context,
	content fs.FS,
	zone *worldgen.Zone,
	resolver gameworld.ObjectResolver,
) (*gameworld.Map, error) {
	materializer, err := gameworld.NewMaterializer(content, zone, resolver)
	if err != nil {
		return nil, err
	}

	for {
		err = materializer.Step(ctx)
		if errors.Is(err, gameworld.ErrMaterializationComplete) {
			break
		}

		if err != nil {
			return nil, err
		}

		progress := materializer.Progress()
		if progress.Completed == progress.Total {
			break
		}
	}

	return materializer.Result()
}
