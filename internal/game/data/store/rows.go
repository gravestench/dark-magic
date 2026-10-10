package recordstore

// RowOrdinal is a zero-based position in the authored table. It is not an
// authored numeric ID, an act-local index, or a key in a derived lookup.
type RowOrdinal int

// SymbolicID is an authored text identifier such as a monster code. Its
// spelling is kept intact; a case-folded lookup key is a separate value.
type SymbolicID string

// NumericID is an explicit number stored in a record, not its position.
type NumericID int

// ActLocalID is meaningful only together with its act. Some tables author its
// Index column; others, such as MonPreset, derive it from per-act row order.
type ActLocalID struct {
	Act   int
	Index int
}

// DerivedIndex is a consumer's lookup convenience, not an authored identity.
type DerivedIndex string

// Row keeps the coordinates of one generic record beside its authored cells.
// Values belongs to the caller, just like the maps returned by Load.
type Row struct {
	Values     map[string]string
	Ordinal    RowOrdinal
	Table      string
	Provenance Provenance
}

// LoadRows returns ordered records with their table position and winning
// source. The source is resolved independently of any derived lookup key; a
// consumer must explicitly decide which authored columns are identifiers.
func (s *Store) LoadRows(path string) ([]Row, error) {
	rows, err := s.Load(path)
	if err != nil {
		return nil, err
	}

	table := s.canonicalPath(path)
	provenance, found := s.Source(table)
	if !found {
		provenance.Layer, provenance.Path = s.resolveSource(table)
	}

	result := make([]Row, len(rows))
	for ordinal, values := range rows {
		result[ordinal] = Row{
			Values:     values,
			Ordinal:    RowOrdinal(ordinal),
			Table:      table,
			Provenance: provenance,
		}
	}

	return result, nil
}
