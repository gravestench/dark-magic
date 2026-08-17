package recordstore

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
)

type incompleteListSource struct {
	*content.FS
}

func (source incompleteListSource) List(root, suffix string) ([]string, error) {
	paths, err := source.FS.List(root, suffix)
	if err != nil {
		return nil, err
	}
	result := paths[:0]
	for _, path := range paths {
		if path != "data/global/excel/MonPreset.txt" {
			result = append(result, path)
		}
	}
	return result, nil
}

var _ interface {
	fs.FS
	List(string, string) ([]string, error)
	ResolveSource(string) (string, string, error)
} = incompleteListSource{}

func TestPinHashesEffectiveTablesAndFreezesTheirBytes(t *testing.T) {
	base := fstest.MapFS{
		"data/global/AnimData.d2":      &fstest.MapFile{Data: []byte("authoritative animation timing")},
		"data/global/excel/armor.txt":  &fstest.MapFile{Data: []byte("code\nbase\n")},
		"data/global/excel/skills.txt": &fstest.MapFile{Data: []byte("skill\nbase\n")},
		"data/global/ui/panel.dc6":     &fstest.MapFile{Data: []byte("presentation")},
	}
	patch := fstest.MapFS{
		"data/global/excel/armor.txt": &fstest.MapFile{Data: []byte("code\npatch\n")},
	}
	source, err := content.New(content.Layer{Name: "patch", FS: patch}, content.Layer{Name: "base", FS: base})
	if err != nil {
		t.Fatal(err)
	}
	pinned, generation, err := Pin(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := generation.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(generation.Files) != 3 || generation.Files[0].Path != AnimationDataPath ||
		generation.Files[1].Path != "data/global/excel/armor.txt" || generation.Files[1].Source != "patch" ||
		generation.Files[2].Source != "base" {
		t.Fatalf("generation files = %#v", generation.Files)
	}
	patch["data/global/excel/armor.txt"].Data = []byte("code\nchanged\n")
	pinned.Invalidate("data/global/excel/armor.txt")
	rows, err := pinned.Load("data/global/excel/armor.txt")
	if err != nil {
		t.Fatal(err)
	}
	if rows[0]["code"] != "patch" {
		t.Fatalf("pinned rows changed after source mutation: %#v", rows)
	}
	_, sameGeneration, err := Pin(source)
	if err != nil {
		t.Fatal(err)
	}
	if generation.ID == sameGeneration.ID {
		t.Fatal("changed effective table bytes did not create a new generation")
	}
	base[AnimationDataPath].Data = []byte("changed authoritative animation timing")
	pinnedAnimation, err := pinned.Read(AnimationDataPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(pinnedAnimation) != "authoritative animation timing" {
		t.Fatalf("pinned AnimData changed after source mutation: %q", pinnedAnimation)
	}
	_, animationChanged, err := Pin(source)
	if err != nil {
		t.Fatal(err)
	}
	if sameGeneration.ID == animationChanged.ID {
		t.Fatal("changed AnimData bytes did not create a new generation")
	}
	base["data/global/ui/panel.dc6"].Data = []byte("changed presentation")
	_, presentationOnly, err := Pin(source)
	if err != nil {
		t.Fatal(err)
	}
	if animationChanged.ID != presentationOnly.ID {
		t.Fatal("presentation-only bytes changed the authoritative generation")
	}
}

func TestPinIncludesWinningSourceProvenance(t *testing.T) {
	files := fstest.MapFS{"data/global/excel/skills.txt": &fstest.MapFile{Data: []byte("skill\nattack\n")}}
	first, err := content.New(content.Layer{Name: "patch", FS: files})
	if err != nil {
		t.Fatal(err)
	}
	second, err := content.New(content.Layer{Name: "expansion", FS: files})
	if err != nil {
		t.Fatal(err)
	}
	_, firstGeneration, err := Pin(first)
	if err != nil {
		t.Fatal(err)
	}
	_, secondGeneration, err := Pin(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstGeneration.ID == secondGeneration.ID {
		t.Fatal("different winning provenance produced the same generation")
	}
}

func TestPinnedStoreRetainsCaseInsensitiveMPQLookups(t *testing.T) {
	source, err := content.New(content.Layer{Name: "patch", FS: fstest.MapFS{
		"data/global/excel/MonPreset.txt": &fstest.MapFile{Data: []byte("Act\tPlace\n1\tzombie\n")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	pinned, _, err := Pin(source)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := pinned.Load("data/global/excel/monpreset.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0]["Place"] != "zombie" {
		t.Fatalf("rows = %#v", rows)
	}
	provenance, found := pinned.Source("DATA/GLOBAL/EXCEL/MONPRESET.TXT")
	if !found || provenance.Layer != "patch" || provenance.Path != "data/global/excel/MonPreset.txt" {
		t.Fatalf("provenance = %#v, %v", provenance, found)
	}
	pinned.Invalidate("DATA/GLOBAL/EXCEL/MONPRESET.TXT")
	if pinned.Loaded("data/global/excel/monpreset.txt") {
		t.Fatal("case-insensitive invalidation left the canonical table cached")
	}
}

func TestPinDiscoversRequiredTableMissingFromMPQListfile(t *testing.T) {
	layered, err := content.New(content.Layer{Name: "retail", FS: fstest.MapFS{
		"data/global/excel/MonPreset.txt": &fstest.MapFile{Data: []byte("Act\tPlace\n1\tzombie\n")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	pinned, generation, err := Pin(incompleteListSource{FS: layered})
	if err != nil {
		t.Fatal(err)
	}
	if len(generation.Files) != 1 || generation.Files[0].Path != "data/global/excel/MonPreset.txt" {
		t.Fatalf("generation files = %#v", generation.Files)
	}
	if _, err := pinned.Load("data/global/excel/monpreset.txt"); err != nil {
		t.Fatal(err)
	}
}
