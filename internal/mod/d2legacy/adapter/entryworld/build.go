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

// Build generates and materializes the canonical Act I entry pair. The operation resolves the seam and town spawn
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

	town, err := materialize(ctx, content, generated.Town, resolver)
	if err != nil {
		return nil, fmt.Errorf("materialize Act I town: %w", err)
	}

	moor, err := materialize(ctx, content, generated.Wilderness, resolver)
	if err != nil {
		return nil, fmt.Errorf("materialize Blood Moor: %w", err)
	}

	seam, err := gametransition.ResolveSeam(generated.Seam, town, moor)
	if err != nil {
		return nil, fmt.Errorf("join Act I town to Blood Moor: %w", err)
	}

	townX, townY, found := d2mapgen.ResolveTownEntry(ctx, d2legacySource, records, town)
	if !found {
		return nil, errors.New("d2legacy entry world: Act I town has no campfire entry")
	}

	return preparedWorld(generated, town, moor, seam, townX, townY, difficulty), nil
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

// preparedWorld assembles all seed-dependent artifacts only after generation, materialization, seam resolution, and
// spawn selection have succeeded. The wilderness arrival comes from the resolved seam rather than raw recipe tiles.
func preparedWorld(
	generated d2mapgen.EntryWorld,
	town *gameworld.Map,
	moor *gameworld.Map,
	seam gametransition.Seam,
	townX float64,
	townY float64,
	difficulty int,
) *Prepared {
	return &Prepared{
		Worlds: map[int]*gameworld.Map{
			generated.Seam.FirstLevel:  town,
			generated.Seam.SecondLevel: moor,
		},
		Zones: map[int]*worldgen.Zone{
			generated.Seam.FirstLevel:  generated.Town,
			generated.Seam.SecondLevel: generated.Wilderness,
		},
		Spawns: map[int][2]float64{
			generated.Seam.FirstLevel:  {townX, townY},
			generated.Seam.SecondLevel: {seam.Wilderness.ArrivalX, seam.Wilderness.ArrivalY},
		},
		Seam:       seam,
		Difficulty: difficulty,
	}
}
