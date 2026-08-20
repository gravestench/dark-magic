package recordstore

import (
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
)

// TestValidateLinkReportsIdentityAndPinnedProvenance protects the diagnostic
// fields tooling uses to locate an unresolved authored reference.
func TestValidateLinkReportsIdentityAndPinnedProvenance(t *testing.T) {
	source, err := content.New(content.Layer{Name: "patch", FS: fstest.MapFS{
		"data/global/excel/skills.txt": &fstest.MapFile{
			Data: []byte("skill\tsrvmissile\nFire Bolt\tfirebolt\nBroken Skill\tmissing\n"),
		},
		"data/global/excel/missiles.txt": &fstest.MapFile{Data: []byte("Missile\nfirebolt\n")},
	}})
	if err != nil {
		t.Fatal(err)
	}

	pinned, _, err := Pin(source)
	if err != nil {
		t.Fatal(err)
	}

	diagnostics, err := pinned.ValidateLink(LinkSpec{Name: "skill-server-missile",
		SourceTable: "data/global/excel/skills.txt", SourceColumn: "srvmissile", SourceKeyColumn: "skill",
		TargetTable: "data/global/excel/missiles.txt", TargetColumn: "Missile", Identity: IdentitySymbolic})
	if err != nil {
		t.Fatal(err)
	}

	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	diagnostic := diagnostics[0]
	if diagnostic.RowOrdinal != 1 || diagnostic.SourceLine != 3 || diagnostic.AuthoredKey != "Broken Skill" ||
		diagnostic.Column != "srvmissile" || diagnostic.RawValue != "missing" ||
		diagnostic.SourceLayer != "patch" || diagnostic.SourcePath != "data/global/excel/skills.txt" ||
		diagnostic.Identity != IdentitySymbolic || diagnostic.Severity != "error" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

// TestValidateLinkReportsAmbiguousTargetAndAllowsEmptySentinel distinguishes a
// permitted missing link from an ambiguous non-empty target identity.
func TestValidateLinkReportsAmbiguousTargetAndAllowsEmptySentinel(t *testing.T) {
	files := fstest.MapFS{
		"data/global/excel/source.txt": &fstest.MapFile{Data: []byte("id\tlink\none\t\ntwo\t7\n")},
		"data/global/excel/target.txt": &fstest.MapFile{Data: []byte("id\n7\n7\n")},
	}

	source, err := content.New(content.Layer{Name: "fixture", FS: files})
	if err != nil {
		t.Fatal(err)
	}

	pinned, _, err := Pin(source)
	if err != nil {
		t.Fatal(err)
	}

	diagnostics, err := pinned.ValidateLink(LinkSpec{Name: "numeric-fixture",
		SourceTable: "data/global/excel/source.txt", SourceColumn: "link", SourceKeyColumn: "id",
		TargetTable: "data/global/excel/target.txt", TargetColumn: "id", Identity: IdentityNumeric, AllowEmpty: true})
	if err != nil {
		t.Fatal(err)
	}

	diagnosticIsAmbiguous := len(diagnostics) == 1 &&
		diagnostics[0].Message == "target identity is ambiguous" &&
		diagnostics[0].Identity == IdentityNumeric
	if !diagnosticIsAmbiguous {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}
