package world

import "testing"

// TestIntegrateVelocityMatchesBoundedAxisSeparatedMovement pins collision sliding and radius-adjusted map bounds.
func TestIntegrateVelocityMatchesBoundedAxisSeparatedMovement(t *testing.T) {
	open, err := NewOpenMap(20, 20)
	if err != nil {
		t.Fatal(err)
	}

	position := IntegrateVelocity(open, Point{X: 1.2, Y: 1.2}, Point{X: -10, Y: 5}, Point{X: 20, Y: 20}, 1, .1)
	if position != (Point{X: 1, Y: 1.7}) {
		t.Fatalf("bounded position = %+v, want {1 1.7}", position)
	}

	blocked, err := NewOpenMap(20, 20)
	if err != nil {
		t.Fatal(err)
	}

	blocked.flags[CollisionCell(6)+CollisionCell(5)*blocked.WidthSubtiles] = Flags{BlockWalk: true}

	position = IntegrateVelocity(blocked, Point{X: 5, Y: 5}, Point{X: 10, Y: 10}, Point{X: 20, Y: 20}, 0, .1)
	if position != (Point{X: 5, Y: 6}) {
		t.Fatalf("axis-separated collision position = %+v, want {5 6}", position)
	}
}
