package models

// ItemProperty represents an item modifier and its associated function.
type ItemProperty struct {
	Code string `csv:"code"` // ItemProperty ID used as a reference in other data files.

	Func1 string `csv:"func1"` // Function ID for defining the property behavior.
	Func2 string `csv:"func2"` // Additional Function ID.
	Func3 string `csv:"func3"` // Additional Function ID.
	Func4 string `csv:"func4"` // Additional Function ID.
	Func5 string `csv:"func5"` // Additional Function ID.
	Func6 string `csv:"func6"` // Additional Function ID.
	Func7 string `csv:"func7"` // Additional Function ID.

	Stat1 string `csv:"stat1"` // Stat applied by the property (references ItemStatCost.txt).
	Stat2 string `csv:"stat2"` // Additional Stat.
	Stat3 string `csv:"stat3"` // Additional Stat.
	Stat4 string `csv:"stat4"` // Additional Stat.
	Stat5 string `csv:"stat5"` // Additional Stat.
	Stat6 string `csv:"stat6"` // Additional Stat.
	Stat7 string `csv:"stat7"` // Additional Stat.

	Set1 bool `csv:"set1"` // Boolean field. If true, set the stat value regardless of its current value.
	Set2 bool `csv:"set2"` // Additional Boolean field.
	Set3 bool `csv:"set3"` // Additional Boolean field.
	Set4 bool `csv:"set4"` // Additional Boolean field.
	Set5 bool `csv:"set5"` // Additional Boolean field.
	Set6 bool `csv:"set6"` // Additional Boolean field.
	Set7 bool `csv:"set7"` // Additional Boolean field.

	Val1 int `csv:"val1"` // Integer field used as a possible input parameter for additional function calculations.
	Val2 int `csv:"val2"` // Additional Integer field.
	Val3 int `csv:"val3"` // Additional Integer field.
	Val4 int `csv:"val4"` // Additional Integer field.
	Val5 int `csv:"val5"` // Additional Integer field.
	Val6 int `csv:"val6"` // Additional Integer field.
	Val7 int `csv:"val7"` // Additional Integer field.
}
