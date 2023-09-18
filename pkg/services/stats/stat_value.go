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
type StatValue interface {
	NumberType() StatNumberType
	CombineType() ValueCombineType

	Clone() StatValue

	SetInt(int) StatValue
	SetFloat(float64) StatValue
	SetStringer(func(StatValue) string) StatValue

	Int() int
	Float() float64
	String() string
	Stringer() func(StatValue) string
}

// Diablo2StatValue is a diablo 2 implementation of a stat value
type Diablo2StatValue struct {
	number      float64
	stringerFn  func(StatValue) string
	numberType  StatNumberType
	combineType ValueCombineType
}

// NumberType returns the stat value type
func (sv *Diablo2StatValue) NumberType() StatNumberType {
	return sv.numberType
}

// CombineType returns the stat value combination type
func (sv *Diablo2StatValue) CombineType() ValueCombineType {
	return sv.combineType
}

// Clone returns a deep copy of the stat value
func (sv Diablo2StatValue) Clone() StatValue {
	clone := &Diablo2StatValue{}

	switch sv.numberType {
	case StatValueInt:
		clone.SetInt(sv.Int())
	case StatValueFloat:
		clone.SetFloat(sv.Float())
	}

	clone.stringerFn = sv.stringerFn

	return clone
}

// Int returns the integer version of the stat value
func (sv *Diablo2StatValue) Int() int {
	return int(sv.number)
}

// String returns a string version of the value
func (sv *Diablo2StatValue) String() string {
	return sv.stringerFn(sv)
}

// Float returns a float64 version of the value
func (sv *Diablo2StatValue) Float() float64 {
	return sv.number
}

// SetInt sets the stat value using an int
func (sv *Diablo2StatValue) SetInt(i int) StatValue {
	sv.number = float64(i)

	return sv
}

// SetFloat sets the stat value using a float64
func (sv *Diablo2StatValue) SetFloat(f float64) StatValue {
	sv.number = f

	return sv
}

// Stringer returns the string evaluation function
func (sv *Diablo2StatValue) Stringer() func(StatValue) string {
	return sv.stringerFn
}

// SetStringer sets the string evaluation function
func (sv *Diablo2StatValue) SetStringer(f func(StatValue) string) StatValue {
	sv.stringerFn = f
	return sv
}
