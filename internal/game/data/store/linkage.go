package recordstore

import (
	"fmt"
	"strings"
)

// IdentityKind records how a caller interprets a link value without making the generic store own Diablo semantics.
type IdentityKind string

const (
	// IdentitySymbolic identifies authored string keys.
	IdentitySymbolic IdentityKind = "symbolic_id"
	// IdentityNumeric identifies authored numeric keys represented by their original text.
	IdentityNumeric IdentityKind = "numeric_id"
	// IdentityRowOrdinal identifies a target by its zero-based position in a table.
	IdentityRowOrdinal IdentityKind = "row_ordinal"
	// IdentityActLocal identifies a numeric key whose meaning is scoped to an act.
	IdentityActLocal IdentityKind = "act_local_index"
)

// LinkSpec defines one caller-owned relationship between generic source and target tables.
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

// LinkDiagnostic locates an invalid relationship in both authored source data and its winning content layer.
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
	if !validLinkSpec(spec) {
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

	targets := countTargetIdentities(targetRows, spec.TargetColumn)
	provenance, _ := s.Source(spec.SourceTable)

	var diagnostics []LinkDiagnostic
	for ordinal, row := range sourceRows {
		diagnostic, unresolved := unresolvedLinkDiagnostic(spec, provenance, targets, ordinal, row)
		if unresolved {
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	return diagnostics, nil
}

// validLinkSpec rejects only structurally incomplete policy while leaving authored names untouched for exact lookups.
func validLinkSpec(spec LinkSpec) bool {
	return strings.TrimSpace(spec.Name) != "" &&
		strings.TrimSpace(spec.SourceTable) != "" &&
		strings.TrimSpace(spec.SourceColumn) != "" &&
		strings.TrimSpace(spec.TargetTable) != "" &&
		strings.TrimSpace(spec.TargetColumn) != "" &&
		validIdentityKind(spec.Identity)
}

// countTargetIdentities preserves duplicate counts because ambiguity is distinct from a missing relationship.
func countTargetIdentities(rows []map[string]string, column string) map[string]int {
	identities := make(map[string]int, len(rows))
	for _, row := range rows {
		identities[row[column]]++
	}

	return identities
}

// unresolvedLinkDiagnostic applies empty-sentinel and uniqueness policy to one source row. Returning a separate flag
// keeps a valid zero-value diagnostic distinct from a relationship that callers must report.
func unresolvedLinkDiagnostic(
	spec LinkSpec,
	provenance Provenance,
	targets map[string]int,
	ordinal int,
	row map[string]string,
) (LinkDiagnostic, bool) {
	raw := row[spec.SourceColumn]
	if raw == "" && spec.AllowEmpty {
		return LinkDiagnostic{}, false
	}

	matches := targets[raw]
	if matches == 1 {
		return LinkDiagnostic{}, false
	}

	message := "target is missing"
	if matches > 1 {
		message = "target identity is ambiguous"
	}

	return LinkDiagnostic{
		LinkName:     spec.Name,
		SourceTable:  spec.SourceTable,
		SourceLayer:  provenance.Layer,
		SourcePath:   provenance.Path,
		RowOrdinal:   ordinal,
		SourceLine:   ordinal + 2,
		AuthoredKey:  row[spec.SourceKeyColumn],
		Column:       spec.SourceColumn,
		RawValue:     raw,
		TargetTable:  spec.TargetTable,
		TargetColumn: spec.TargetColumn,
		Identity:     spec.Identity,
		Severity:     "error",
		Message:      message,
	}, true
}

// validIdentityKind keeps diagnostics constrained to the stable serialized identity vocabulary.
func validIdentityKind(kind IdentityKind) bool {
	return kind == IdentitySymbolic || kind == IdentityNumeric || kind == IdentityRowOrdinal || kind == IdentityActLocal
}
