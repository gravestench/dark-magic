package stats

// StatNumberType is a value type for a stat value
type StatNumberType int

// Stat value types
const (
	StatValueInt StatNumberType = iota
	StatValueFloat
)

// ValueCombineType is a rule for combining stat values
type ValueCombineType int

const (
	// StatValueCombineSum means that the values are simply summed
	StatValueCombineSum ValueCombineType = iota

	// StatValueCombineStatic means that values can be combined only if they
	// have the same number value, and that the combination does not alter
	// the number value. This is typically for things like static skill level
	// monster/skill index for on proc stats where it doesnt make sense to sum
	// the values
	// example 1:
	//	if
	//		Stat_A := `25% chance to cast level 2 Frozen Orb on attack`
	//		Stat_B := `25% chance to cast level 3 Frozen Orb on attack`
	// then
	// 		Stat_A can NOT be combined with Stat_B
	//		even though the skills are the same, the levels are different
	//
	// example 2:
	//	if
	//		Stat_A := `25% chance to cast level 20 Frost Nova on attack`
	//		Stat_B := `25% chance to cast level 20 Frost Nova on attack`
	// then
	// 		the skills and skill levels are the same, so it can be combined
	//		(Stat_A + Stat_B) == `50% chance to cast level 20 Frost Nova on attack`
	StatValueCombineStatic
)

// StatValue is something that can have both integer and float
// number components, as well as a means of retrieving a string for
// its values.
type StatValue struct {
	numberType  StatNumberType
	combineType ValueCombineType
	stringerFn  func(StatValue) string
}

func (s *StatValue) NumberType() StatNumberType {
	//TODO implement me
	panic("implement me")
}

func (s *StatValue) CombineType() ValueCombineType {
	//TODO implement me
	panic("implement me")
}

func (s *StatValue) Clone() StatValue {
	//TODO implement me
	panic("implement me")
}

func (s *StatValue) SetInt(i int) StatValue {
	//TODO implement me
	panic("implement me")
}

func (s *StatValue) SetFloat(f float64) StatValue {
	//TODO implement me
	panic("implement me")
}

func (s *StatValue) SetStringer(f func(StatValue) string) StatValue {
	//TODO implement me
	panic("implement me")
}

func (s *StatValue) Int() int {
	//TODO implement me
	panic("implement me")
}

func (s *StatValue) Float() float64 {
	//TODO implement me
	panic("implement me")
}

func (s *StatValue) String() string {
	//TODO implement me
	panic("implement me")
}

func (s *StatValue) Stringer() func(StatValue) string {
	//TODO implement me
	panic("implement me")
}
