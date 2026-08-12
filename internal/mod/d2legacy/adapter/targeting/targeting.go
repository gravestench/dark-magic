// Package targeting resolves d2legacy pointer hits against authoritative ECS
// entities. Authored DS1 stamps remain world facts; this package classifies the
// live things created from those facts.
package targeting

import (
	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const Component = "d2legacy.world.selectable"

type Hit struct {
	ID, Kind, Label, Owner string
	X, Y, Radius           float64
	Priority               int64
}

type Resolver struct{ engine *gameecs.Engine }

func New(engine *gameecs.Engine) *Resolver { return &Resolver{engine: engine} }

func (resolver *Resolver) HitAt(x, y float64) (Hit, bool) {
	if resolver == nil || resolver.engine == nil {
		return Hit{}, false
	}
	selectables, ok := akara.GetDynamicStore(resolver.engine.World(), Component)
	if !ok {
		return Hit{}, false
	}
	positions, ok := akara.GetDynamicStore(resolver.engine.World(), "d2legacy.world.position")
	if !ok {
		return Hit{}, false
	}
	hits := make(map[string]Hit)
	for _, entity := range selectables.Entities() {
		component, present := selectables.Get(entity)
		if !present {
			continue
		}
		position, present := positions.Get(entity)
		if !present {
			continue
		}
		id, _ := component.Get("id")
		kind, _ := component.Get("kind")
		label, _ := component.Get("label")
		owner, _ := component.Get("owner")
		radius, _ := component.Get("radius")
		priority, _ := component.Get("priority")
		px, _ := position.Get("x")
		py, _ := position.Get("y")
		hit := Hit{ID: id.(string), Kind: kind.(string), Label: label.(string), Owner: owner.(string), X: px.(float64), Y: py.(float64), Radius: radius.(float64), Priority: priority.(int64)}
		hits[hit.ID] = hit
	}
	items := make([]gameworld.Selectable, 0, len(hits))
	for _, hit := range hits {
		items = append(items, gameworld.Selectable{ID: hit.ID, Kind: hit.Kind, Label: hit.Label, X: hit.X, Y: hit.Y, Radius: hit.Radius, Priority: int(hit.Priority)})
	}
	selector, err := gameworld.NewSelector(items, 8)
	if err != nil {
		return Hit{}, false
	}
	selected, found := selector.Hit(x, y)
	if !found {
		return Hit{}, false
	}
	return hits[selected.ID], true
}

func Schema() akara.Schema {
	return akara.Schema{Name: Component, Version: 1, Fields: []akara.Field{{Name: "id", Kind: akara.FieldString}, {Name: "kind", Kind: akara.FieldString}, {Name: "label", Kind: akara.FieldString}, {Name: "owner", Kind: akara.FieldString}, {Name: "radius", Kind: akara.FieldFloat64}, {Name: "priority", Kind: akara.FieldInt64}}}
}
