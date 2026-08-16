package recordstore

import (
	"fmt"
	"strings"
)

type IdentityKind string

const (
	IdentitySymbolic   IdentityKind = "symbolic_id"
	IdentityNumeric    IdentityKind = "numeric_id"
	IdentityRowOrdinal IdentityKind = "row_ordinal"
	IdentityActLocal   IdentityKind = "act_local_index"
)

type LinkSpec struct {
	Name            string
	SourceTable     string
	SourceColumn    string
	SourceKeyColumn string
	TargetTable     string
	TargetColumn    string
	Identity        IdentityKind
	AllowEmpty      bool
}

type LinkDiagnostic struct {
	LinkName     string       `json:"link_name"`
	SourceTable  string       `json:"source_table"`
	SourceLayer  string       `json:"source_layer,omitempty"`
	SourcePath   string       `json:"source_path,omitempty"`
	RowOrdinal   int          `json:"row_ordinal"`
	SourceLine   int          `json:"source_line"`
	AuthoredKey  string       `json:"authored_key,omitempty"`
	Column       string       `json:"column"`
	RawValue     string       `json:"raw_value"`
	TargetTable  string       `json:"target_table"`
	TargetColumn string       `json:"target_column"`
	Identity     IdentityKind `json:"identity_kind"`
	Severity     string       `json:"severity"`
	Message      string       `json:"message"`
}

// ValidateLink reports source-located missing or ambiguous relationships. It
// does not repair data or decide what a Diablo column means; d2legacy supplies
// that policy explicitly through LinkSpec.
func (s *Store) ValidateLink(spec LinkSpec) ([]LinkDiagnostic, error) {
	if strings.TrimSpace(spec.Name) == "" || strings.TrimSpace(spec.SourceTable) == "" ||
		strings.TrimSpace(spec.SourceColumn) == "" || strings.TrimSpace(spec.TargetTable) == "" ||
		strings.TrimSpace(spec.TargetColumn) == "" || !validIdentityKind(spec.Identity) {
		return nil, fmt.Errorf("recordstore: incomplete link specification")
	}
	sourceRows, err := s.Load(spec.SourceTable)
	if err != nil {
		return nil, err
	}
	targetRows, err := s.Load(spec.TargetTable)
	if err != nil {
		return nil, err
	}
	targets := make(map[string]int, len(targetRows))
	for _, row := range targetRows {
		targets[row[spec.TargetColumn]]++
	}
	provenance, _ := s.Source(spec.SourceTable)
	var diagnostics []LinkDiagnostic
	for ordinal, row := range sourceRows {
		raw := row[spec.SourceColumn]
		if raw == "" && spec.AllowEmpty {
			continue
		}
		matches := targets[raw]
		if matches == 1 {
			continue
		}
		message := "target is missing"
		if matches > 1 {
			message = "target identity is ambiguous"
		}
		diagnostics = append(diagnostics, LinkDiagnostic{
			LinkName: spec.Name, SourceTable: spec.SourceTable, SourceLayer: provenance.Layer,
			SourcePath: provenance.Path, RowOrdinal: ordinal, SourceLine: ordinal + 2,
			AuthoredKey: row[spec.SourceKeyColumn], Column: spec.SourceColumn, RawValue: raw,
			TargetTable: spec.TargetTable, TargetColumn: spec.TargetColumn, Identity: spec.Identity,
			Severity: "error", Message: message,
		})
	}
	return diagnostics, nil
}

func validIdentityKind(kind IdentityKind) bool {
	return kind == IdentitySymbolic || kind == IdentityNumeric || kind == IdentityRowOrdinal || kind == IdentityActLocal
}
