package worldobjects

import (
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/game/data/recovered"
	"github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestResolverUsesActLocalOrderingForBothObjectKinds(t *testing.T) {
	rows := fstest.MapFS{"data/global/excel/monpreset.txt": &fstest.MapFile{Data: []byte("Act\tPlace\n1\tfallen\n2\tskeleton\n1\tzombie\n")}}
	resolver, err := New(recovered.Snapshot{MapObjects: []recovered.MapObject{{Act: 1, ID: 3, ObjectID: 108, Description: "Malus"}}}, recordstore.New(rows))
	if err != nil {
		t.Fatal(err)
	}
	if id, description, found := resolver.ResolveStaticObject(1, 3); !found || id != 108 || description != "Malus" {
		t.Fatalf("static = %d, %q, %v", id, description, found)
	}
	if class, found := resolver.ResolveDynamicObject(1, 1); !found || class != "zombie" {
		t.Fatalf("dynamic = %q, %v", class, found)
	}
	if _, found := resolver.ResolveDynamicObject(2, 1); found {
		t.Fatal("act-local dynamic index leaked across acts")
	}
}
