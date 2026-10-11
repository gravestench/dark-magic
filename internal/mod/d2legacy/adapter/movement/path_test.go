package movement

import (
	"reflect"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

func TestRouteCornersPreservesRouteGeometry(t *testing.T) {
	for _, trial := range []struct {
		name    string
		current gameworld.Point
		path    []gameworld.Point
		want    []gameworld.Point
	}{
		{name: "empty"},
		{name: "stationary", path: []gameworld.Point{{0, 0}}, want: []gameworld.Point{{0, 0}}},
		{name: "direct", path: []gameworld.Point{{0, 0}, {1, 0}}, want: []gameworld.Point{{0, 0}, {1, 0}}},
		{name: "axis", path: []gameworld.Point{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, want: []gameworld.Point{{0, 0}, {2, 0}, {3, 0}}},
		{name: "diagonal", path: []gameworld.Point{{0, 0}, {1, 1}, {2, 2}, {3, 3}}, want: []gameworld.Point{{0, 0}, {2, 2}, {3, 3}}},
		{name: "corner", path: []gameworld.Point{{0, 0}, {1, 0}, {2, 0}, {2, 1}, {2, 2}}, want: []gameworld.Point{{0, 0}, {2, 0}, {2, 1}, {2, 2}}},
		{name: "reverse", path: []gameworld.Point{{0, 0}, {1, 0}, {0, 0}, {-1, 0}, {-2, 0}}, want: []gameworld.Point{{0, 0}, {1, 0}, {-1, 0}, {-2, 0}}},
		{name: "fractional_start_on_segment", current: gameworld.Point{0.25, 0}, path: []gameworld.Point{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, want: []gameworld.Point{{0, 0}, {2, 0}, {3, 0}}},
		{name: "fractional_start_off_segment", current: gameworld.Point{0, 0.25}, path: []gameworld.Point{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}}, want: []gameworld.Point{{0, 0}, {1, 0}, {3, 0}, {4, 0}}},
		{name: "small_turn", path: []gameworld.Point{{0, 0}, {1, 1e-14}, {2, 0}, {3, 0}, {4, 0}}, want: []gameworld.Point{{0, 0}, {1, 1e-14}, {2, 0}, {3, 0}, {4, 0}}},
		{name: "duplicate", path: []gameworld.Point{{0, 0}, {0, 0}, {1, 0}, {2, 0}}, want: []gameworld.Point{{0, 0}, {0, 0}, {1, 0}, {2, 0}}},
	} {
		t.Run(trial.name, func(t *testing.T) {
			if got := routeCorners(trial.path, trial.current); !reflect.DeepEqual(got, trial.want) {
				t.Fatalf("route corners = %v, want %v", got, trial.want)
			}
		})
	}
}
