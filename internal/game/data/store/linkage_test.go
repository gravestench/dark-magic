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

// These fixtures deliberately give one row three different numeric-looking
// identities. A column value must never stand in for row or per-act position.
func TestValidateLinkKeepsNumericOrdinalAndActLocalTargetsDistinct(t *testing.T) {
	store := New(fstest.MapFS{
		"source.txt": &fstest.MapFile{Data: []byte("Act\tordinal\tnumeric\tlocal\n1\t2\t42\t1\n1\t42\t2\t2\n")},
		"target.txt": &fstest.MapFile{Data: []byte("Act\tId\tCode\n1\t99\tfirst\n2\t0\tother-act\n1\t42\tsecond\n")},
	})
	base := LinkSpec{Name: "identity-kind", SourceTable: "source.txt", TargetTable: "target.txt"}
	tests := []struct {
		name, sourceColumn string
		kind               IdentityKind
		wantValue          string
	}{
		{"row ordinal", "ordinal", IdentityRowOrdinal, "42"},
		{"explicit numeric ID", "numeric", IdentityNumeric, "2"},
		{"act-local position", "local", IdentityActLocal, "2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := base
			spec.SourceColumn = test.sourceColumn
			spec.Identity = test.kind
			switch test.kind {
			case IdentityNumeric:
				spec.TargetColumn = "Id"
			case IdentityActLocal:
				spec.SourceActColumn = "Act"
				spec.TargetActColumn = "Act"
			}
			diagnostics, err := store.ValidateLink(spec)
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 1 || diagnostics[0].RawValue != test.wantValue || diagnostics[0].RowOrdinal != 1 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
		})
	}

	_, err := store.ValidateLink(LinkSpec{Name: "ambiguous", SourceTable: "source.txt", SourceColumn: "ordinal",
		TargetTable: "target.txt", TargetColumn: "Id", Identity: IdentityRowOrdinal})
	if err == nil {
		t.Fatal("row-ordinal link accepted an explicit target ID column")
	}
}
