package stats

import (
	"math"
	"reflect"
	"testing"
)

func fixtureSource(id SourceID, kind SourceKind, owner string, entries ...Entry) Source {
	return Source{ID: id, Kind: kind, Lifetime: LifetimeDurable, Owner: OwnerRef{Kind: string(kind), ID: owner}, Entries: entries}
}

func TestAuthorityKeepsSourceProvenanceAndParameters(t *testing.T) {
	authority := NewAuthority()
	strength := Key{ID: 1}
	skillOne := Key{ID: 97, Parameter: 1}
	skillTwo := Key{ID: 97, Parameter: 2}
	base := fixtureSource("base", SourceBase, "hero", Entry{Key: strength, Value: 20})
	boots := fixtureSource("item:boots", SourceItem, "boots", Entry{Key: strength, Value: 5}, Entry{Key: skillOne, Value: 1})
	charm := fixtureSource("item:charm", SourceCharm, "charm", Entry{Key: skillTwo, Value: 3})

	if _, err := authority.Apply("hero", Mutation{Replace: []Source{base, boots, charm}}); err != nil {
		t.Fatal(err)
	}
	assertEffective(t, authority, "hero", strength, 25)
	assertEffective(t, authority, "hero", skillOne, 1)
	assertEffective(t, authority, "hero", skillTwo, 3)

	snapshot := authority.Snapshot("hero")
	if got := []SourceID{snapshot.Sources[0].ID, snapshot.Sources[1].ID, snapshot.Sources[2].ID}; !reflect.DeepEqual(got, []SourceID{"base", "item:boots", "item:charm"}) {
		t.Fatalf("source order = %v", got)
	}
	if _, err := authority.Apply("hero", Mutation{Remove: []SourceID{"item:boots"}}); err != nil {
		t.Fatal(err)
	}
	assertEffective(t, authority, "hero", strength, 20)
	assertEffective(t, authority, "hero", skillOne, 0)
}

func TestAuthorityReplacementIsAtomicAndIdempotent(t *testing.T) {
	authority := NewAuthority()
	key := Key{ID: 1}
	source := fixtureSource("base", SourceBase, "hero", Entry{Key: key, Value: 10})
	revision, err := authority.Apply("hero", Mutation{Replace: []Source{source}})
	if err != nil || revision != 1 {
		t.Fatalf("first apply revision=%d err=%v", revision, err)
	}
	revision, err = authority.Apply("hero", Mutation{Replace: []Source{source}})
	if err != nil || revision != 1 {
		t.Fatalf("idempotent apply revision=%d err=%v", revision, err)
	}
	_, err = authority.Apply("hero", Mutation{Replace: []Source{fixtureSource("item", SourceItem, "item", Entry{Key: key, Value: 5}), {ID: "broken"}}})
	if err == nil {
		t.Fatal("expected invalid batch to fail")
	}
	assertEffective(t, authority, "hero", key, 10)
	if authority.Snapshot("hero").Revision != 1 {
		t.Fatal("failed batch changed the revision")
	}
}

func TestAuthorityRejectsOverflow(t *testing.T) {
	authority := NewAuthority()
	key := Key{ID: 1}
	_, err := authority.Apply("hero", Mutation{Replace: []Source{
		fixtureSource("a", SourceBase, "hero", Entry{Key: key, Value: math.MaxInt64}),
		fixtureSource("b", SourceItem, "item", Entry{Key: key, Value: 1}),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := authority.Effective("hero", key); err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestSessionStateIsCanonicalAndRestorable(t *testing.T) {
	first := NewAuthority()
	second := NewAuthority()
	sources := []Source{
		fixtureSource("z", SourceItem, "z", Entry{Key: Key{ID: 2}, Value: 2}),
		fixtureSource("a", SourceBase, "a", Entry{Key: Key{ID: 1}, Value: 1}),
	}
	_, _ = first.Apply("hero", Mutation{Replace: sources})
	_, _ = second.Apply("hero", Mutation{Replace: []Source{sources[1], sources[0]}})
	left, err := first.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	right, err := second.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	if string(left) != string(right) {
		t.Fatalf("canonical state differs:\n%s\n%s", left, right)
	}
	restored := NewAuthority()
	if err := restored.RestoreState(left); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Snapshot("hero"), restored.Snapshot("hero")) {
		t.Fatalf("restored snapshot differs: %#v", restored.Snapshot("hero"))
	}
}

func TestRestoreRejectsNonCanonicalStateWithoutMutation(t *testing.T) {
	authority := NewAuthority()
	key := Key{ID: 1}
	_, _ = authority.Apply("hero", Mutation{Replace: []Source{fixtureSource("base", SourceBase, "hero", Entry{Key: key, Value: 10})}})
	before, _ := authority.SnapshotState()
	if err := authority.RestoreState([]byte(`{"version":1,"entities":[{"entity":"z","revision":1,"sources":[]},{"entity":"a","revision":1,"sources":[]}]}`)); err == nil {
		t.Fatal("expected non-canonical restore to fail")
	}
	after, _ := authority.SnapshotState()
	if string(before) != string(after) {
		t.Fatal("failed restore changed live state")
	}
}

func assertEffective(t *testing.T, authority *Authority, entity EntityID, key Key, want int64) {
	t.Helper()
	got, err := authority.Effective(entity, key)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("effective %v = %d, want %d", key, got, want)
	}
}
