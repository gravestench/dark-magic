package world

import (
	"container/heap"
	"fmt"
	"math"
)

type PathRequest struct {
	Start, Goal Point
	Radius      float64
	StopRadius  float64
}

// FindPath performs deterministic eight-way A* over authoritative subtile
// collision. Goals may be occupied: StopRadius accepts the nearest reachable
// cell around an interaction footprint.
func (m *Map) FindPath(request PathRequest) ([]Point, error) {
	if m == nil || request.Radius < 0 || request.StopRadius < 0 || !finitePoint(request.Start.X, request.Start.Y) || !finitePoint(request.Goal.X, request.Goal.Y) {
		return nil, fmt.Errorf("world: invalid path request")
	}
	start := navCell{CollisionCell(request.Start.X), CollisionCell(request.Start.Y)}
	goal := navCell{CollisionCell(request.Goal.X), CollisionCell(request.Goal.Y)}
	if !m.walkableCell(start, request.Radius) {
		return nil, fmt.Errorf("world: path start is blocked")
	}
	frontier := &navQueue{}
	heap.Push(frontier, &navNode{cell: start, estimate: navHeuristic(start, goal)})
	cost := map[navCell]int{start: 0}
	parents := make(map[navCell]navCell)
	closed := make(map[navCell]bool)
	for frontier.Len() > 0 {
		current := heap.Pop(frontier).(*navNode)
		if closed[current.cell] {
			continue
		}
		closed[current.cell] = true
		if cellDistance(current.cell, goal) <= request.StopRadius {
			return reconstructPath(parents, start, current.cell), nil
		}
		for _, step := range navSteps {
			next := navCell{current.cell.x + step.x, current.cell.y + step.y}
			if closed[next] || !m.walkableCell(next, request.Radius) {
				continue
			}
			if step.x != 0 && step.y != 0 {
				if !m.walkableCell(navCell{current.cell.x + step.x, current.cell.y}, request.Radius) || !m.walkableCell(navCell{current.cell.x, current.cell.y + step.y}, request.Radius) {
					continue
				}
			}
			nextCost := cost[current.cell] + step.cost
			old, seen := cost[next]
			if seen && nextCost >= old {
				continue
			}
			cost[next], parents[next] = nextCost, current.cell
			heap.Push(frontier, &navNode{cell: next, cost: nextCost, estimate: nextCost + navHeuristic(next, goal)})
		}
	}
	return nil, fmt.Errorf("world: path target is unreachable")
}

func (m *Map) walkableCell(cell navCell, radius float64) bool {
	reach := int(math.Ceil(radius))
	for y := cell.y - reach; y <= cell.y+reach; y++ {
		for x := cell.x - reach; x <= cell.x+reach; x++ {
			if math.Hypot(float64(x-cell.x), float64(y-cell.y)) > radius+0.5 {
				continue
			}
			flags, inside := m.FlagsAt(x, y)
			if !inside || flags.Blocked() {
				return false
			}
		}
	}
	return true
}

// WalkableStep validates one adjacent collision-cell transition without
// allocating an A* frontier. Fixed-tick velocity movement only needs this
// local collision fact; route planning remains the caller of FindPath.
func (m *Map) WalkableStep(start, goal Point, radius float64) bool {
	if m == nil || radius < 0 || !finitePoint(start.X, start.Y) || !finitePoint(goal.X, goal.Y) {
		return false
	}
	from := navCell{CollisionCell(start.X), CollisionCell(start.Y)}
	to := navCell{CollisionCell(goal.X), CollisionCell(goal.Y)}
	dx, dy := to.x-from.x, to.y-from.y
	if absInt(dx) > 1 || absInt(dy) > 1 || !m.walkableCell(to, radius) {
		return false
	}
	if dx != 0 && dy != 0 {
		return m.walkableCell(navCell{from.x + dx, from.y}, radius) &&
			m.walkableCell(navCell{from.x, from.y + dy}, radius)
	}
	return true
}

type navCell struct{ x, y int }
type navStep struct{ x, y, cost int }

var navSteps = []navStep{{0, -1, 10}, {-1, 0, 10}, {1, 0, 10}, {0, 1, 10}, {-1, -1, 14}, {1, -1, 14}, {-1, 1, 14}, {1, 1, 14}}

func navHeuristic(a, b navCell) int {
	dx, dy := absInt(a.x-b.x), absInt(a.y-b.y)
	return 10*max(dx, dy) + 4*min(dx, dy)
}
func cellDistance(a, b navCell) float64 { return math.Hypot(float64(a.x-b.x), float64(a.y-b.y)) }

func reconstructPath(parents map[navCell]navCell, start, end navCell) []Point {
	cells := []navCell{end}
	for cells[len(cells)-1] != start {
		cells = append(cells, parents[cells[len(cells)-1]])
	}
	result := make([]Point, len(cells))
	for i := range cells {
		cell := cells[len(cells)-1-i]
		result[i] = Point{float64(cell.x), float64(cell.y)}
	}
	return result
}

type navNode struct {
	cell                  navCell
	cost, estimate, index int
}
type navQueue []*navNode

func (q navQueue) Len() int { return len(q) }
func (q navQueue) Less(i, j int) bool {
	if q[i].estimate != q[j].estimate {
		return q[i].estimate < q[j].estimate
	}
	if q[i].cost != q[j].cost {
		return q[i].cost > q[j].cost
	}
	if q[i].cell.y != q[j].cell.y {
		return q[i].cell.y < q[j].cell.y
	}
	return q[i].cell.x < q[j].cell.x
}
func (q navQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i]; q[i].index = i; q[j].index = j }
func (q *navQueue) Push(value any) {
	node := value.(*navNode)
	node.index = len(*q)
	*q = append(*q, node)
}
func (q *navQueue) Pop() any { old := *q; node := old[len(old)-1]; *q = old[:len(old)-1]; return node }
