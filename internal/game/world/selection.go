package world

import (
	"fmt"
	"math"
	"sort"
)

// Selectable is the renderer-neutral pointer footprint of something in the
// world. Presentation may highlight it, but authority decides what activating
// the selected ID means.
type Selectable struct {
	ID, Kind string
	X, Y     float64
	Radius   float64
	Priority int
}

// Selector is a small uniform-grid index. A pointer query only examines nearby
// buckets rather than walking every object in a large generated zone.
type Selector struct {
	cellSize, maxRadius float64
	cells               map[[2]int][]Selectable
}

func NewSelector(selectables []Selectable, cellSize float64) (*Selector, error) {
	if cellSize <= 0 || math.IsNaN(cellSize) || math.IsInf(cellSize, 0) {
		return nil, fmt.Errorf("world: selection cell size must be positive and finite")
	}
	result := &Selector{cellSize: cellSize, cells: make(map[[2]int][]Selectable)}
	seen := make(map[string]struct{}, len(selectables))
	for _, selectable := range selectables {
		if selectable.ID == "" || selectable.Kind == "" || selectable.Radius <= 0 || !finitePoint(selectable.X, selectable.Y) {
			return nil, fmt.Errorf("world: selectable requires ID, kind, position, and positive radius")
		}
		if _, exists := seen[selectable.ID]; exists {
			return nil, fmt.Errorf("world: duplicate selectable %q", selectable.ID)
		}
		seen[selectable.ID] = struct{}{}
		result.maxRadius = max(result.maxRadius, selectable.Radius)
		key := result.cell(selectable.X, selectable.Y)
		result.cells[key] = append(result.cells[key], selectable)
	}
	return result, nil
}

func finitePoint(x, y float64) bool {
	return !math.IsNaN(x) && !math.IsNaN(y) && !math.IsInf(x, 0) && !math.IsInf(y, 0)
}

func (selector *Selector) cell(x, y float64) [2]int {
	return [2]int{int(math.Floor(x / selector.cellSize)), int(math.Floor(y / selector.cellSize))}
}

// Hit returns the best footprint containing the pointer. Higher explicit
// priority wins, then the closest normalized distance, then stable ID order.
func (selector *Selector) Hit(x, y float64) (Selectable, bool) {
	if selector == nil || !finitePoint(x, y) {
		return Selectable{}, false
	}
	reach := int(math.Ceil(selector.maxRadius / selector.cellSize))
	center := selector.cell(x, y)
	candidates := make([]selectableHit, 0)
	for cy := center[1] - reach; cy <= center[1]+reach; cy++ {
		for cx := center[0] - reach; cx <= center[0]+reach; cx++ {
			for _, candidate := range selector.cells[[2]int{cx, cy}] {
				dx, dy := x-candidate.X, y-candidate.Y
				distance := (dx*dx + dy*dy) / (candidate.Radius * candidate.Radius)
				if distance <= 1 {
					candidates = append(candidates, selectableHit{candidate, distance})
				}
			}
		}
	}
	if len(candidates) == 0 {
		return Selectable{}, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority > candidates[j].Priority
		}
		if candidates[i].distance != candidates[j].distance {
			return candidates[i].distance < candidates[j].distance
		}
		return candidates[i].ID < candidates[j].ID
	})
	return candidates[0].Selectable, true
}

type selectableHit struct {
	Selectable
	distance float64
}

// Selectables turns authored DS1 objects into stable semantic footprints.
func (m *Map) Selectables() []Selectable {
	result := make([]Selectable, 0, len(m.Objects))
	for index, object := range m.Objects {
		if !object.Resolved {
			continue
		}
		kind := "static-object"
		if object.Type == ObjectTypeDynamic {
			kind = "dynamic-object"
		}
		result = append(result, Selectable{
			ID: fmt.Sprintf("ds1-object:%d:%d:%d", object.Type, object.ID, index), Kind: kind,
			X: float64(object.X), Y: float64(object.Y), Radius: 1.5,
		})
	}
	return result
}

func (m *Map) SelectableAt(x, y float64) (Selectable, bool) {
	if m == nil {
		return Selectable{}, false
	}
	m.selectorOnce.Do(func() { m.selector, m.selectorErr = NewSelector(m.Selectables(), 8) })
	if m.selectorErr != nil {
		return Selectable{}, false
	}
	return m.selector.Hit(x, y)
}

// LineClear samples the same subtile facts used by movement. Endpoints are not
// treated as occluders because the actor and selected target may occupy them.
func (m *Map) LineClear(fromX, fromY, toX, toY float64) bool {
	x0, y0 := CollisionCell(fromX), CollisionCell(fromY)
	x1, y1 := CollisionCell(toX), CollisionCell(toY)
	dx, dy := absInt(x1-x0), absInt(y1-y0)
	sx, sy := -1, -1
	if x0 < x1 {
		sx = 1
	}
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy
	for {
		if (x0 != CollisionCell(fromX) || y0 != CollisionCell(fromY)) && (x0 != x1 || y0 != y1) {
			flags, inside := m.FlagsAt(x0, y0)
			if !inside || flags.BlockLOS {
				return false
			}
		}
		if x0 == x1 && y0 == y1 {
			return true
		}
		twice := 2 * err
		if twice > -dy {
			err -= dy
			x0 += sx
		}
		if twice < dx {
			err += dx
			y0 += sy
		}
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
