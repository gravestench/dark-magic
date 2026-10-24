package clientapp

import (
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// Dropping any prepared level or inventing a fallback makes this fail.
func TestPreparedVelocityPathsIncludesOnlyPreparedMaps(t *testing.T) {
	town, moor, plains := &gameworld.Map{}, &gameworld.Map{}, &gameworld.Map{}
	for _, worlds := range []map[int]*gameworld.Map{{1: town, 2: moor, 3: plains, 4: nil}, {1: town, 2: moor}} {
		paths, err := preparedVelocityPaths(worlds)
		if err != nil {
			t.Fatal(err)
		}

		count := 0

		for level, world := range worlds {
			if world == nil {
				if _, ok := paths[int64(level)]; ok {
					t.Fatalf("nil level %d was published", level)
				}

				continue
			}

			count++

			if paths[int64(level)] != world {
				t.Fatalf("level %d did not retain its prepared map", level)
			}
		}

		if len(paths) != count {
			t.Fatalf("published %d maps, want %d", len(paths), count)
		}

		if worlds[3] == nil && paths[3] != nil {
			t.Fatal("two-zone topology gained a Cold Plains fallback")
		}
	}
}

// Unavailable geometry must fail registration rather than silently enable collision-free movement.
func TestPreparedVelocityPathsRejectsUnavailableGeometry(t *testing.T) {
	for _, worlds := range []map[int]*gameworld.Map{nil, {}, {1: nil}} {
		if _, err := preparedVelocityPaths(worlds); err == nil {
			t.Fatal("accepted unavailable geometry")
		}
	}
}
