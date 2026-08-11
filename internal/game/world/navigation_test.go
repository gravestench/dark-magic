package world

import "testing"

func openTestMap(width, height int) *Map {
	return &Map{WidthSubtiles: width, HeightSubtiles: height, flags: make([]Flags, width*height)}
}

func TestFindPathRoutesAroundWall(t *testing.T) {
	m := openTestMap(10, 10)
	for y := 0; y < 9; y++ {
		m.flags[y*10+4] = Flags{BlockWalk: true}
	}
	path, err := m.FindPath(PathRequest{Start: Point{2, 2}, Goal: Point{7, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(path) < 2 {
		t.Fatalf("path=%#v", path)
	}
	for _, point := range path {
		flags, _ := m.FlagsAt(int(point.X), int(point.Y))
		if flags.Blocked() {
			t.Fatalf("path crosses wall at %#v", point)
		}
	}
}

func TestFindPathDoesNotCutBlockedCorner(t *testing.T) {
	m := openTestMap(4, 4)
	m.flags[0*4+1] = Flags{BlockWalk: true}
	m.flags[1*4+0] = Flags{BlockWalk: true}
	if _, err := m.FindPath(PathRequest{Start: Point{0, 0}, Goal: Point{2, 2}}); err == nil {
		t.Fatal("path cut sealed diagonal corner")
	}
}

func TestFindPathStopsAroundOccupiedTarget(t *testing.T) {
	m := openTestMap(8, 8)
	m.flags[4*8+4] = Flags{BlockWalk: true}
	path, err := m.FindPath(PathRequest{Start: Point{1, 4}, Goal: Point{4, 4}, StopRadius: 1.5})
	if err != nil {
		t.Fatal(err)
	}
	end := path[len(path)-1]
	if cellDistance(navCell{int(end.X), int(end.Y)}, navCell{4, 4}) > 1.5 {
		t.Fatalf("end=%#v", end)
	}
}

func TestFindPathRespectsEntityRadius(t *testing.T) {
	m := openTestMap(7, 7)
	for x := 0; x < 7; x++ {
		if x != 3 {
			m.flags[3*7+x] = Flags{BlockWalk: true}
		}
	}
	if _, err := m.FindPath(PathRequest{Start: Point{3, 1}, Goal: Point{3, 5}, Radius: 0.75}); err == nil {
		t.Fatal("wide entity fit through one-cell gap")
	}
}
